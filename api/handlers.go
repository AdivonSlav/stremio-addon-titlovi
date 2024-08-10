package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-titlovi/internal/config"
	"go-titlovi/internal/logger"
	"go-titlovi/internal/stremio"
	"go-titlovi/internal/titlovi"
	"go-titlovi/internal/utils"
	"go-titlovi/web"
	"net/http"
	"strconv"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// BuildRouter builds a new router with handler functions to handle all necessary routes and
// also appends middleware.
func BuildRouter(client *titlovi.Client, cache *ristretto.Cache) http.Handler {
	r := mux.NewRouter()

	r.Handle("/", http.HandlerFunc(homeHandler()))

	r.Handle("/manifest.json", http.HandlerFunc(manifestHandler()))
	r.Handle("/{userConfig}/manifest.json", WithAuth(http.HandlerFunc(manifestHandler())))

	r.Handle("/{userConfig}/subtitles/{type}/{id}/{extraArgs}.json", WithAuth(http.HandlerFunc(subtitlesHandler(client, cache))))
	r.Handle("/serve-subtitle/{type}/{mediaid}", http.HandlerFunc(serveSubtitleHandler(client, cache)))

	r.Handle("/configure", http.HandlerFunc(configureHandler()))
	r.Handle("/{userConfig}/configure", WithAuth(http.HandlerFunc(configureHandler())))

	r.Use(WithLogging)

	return r
}

// Serve calls serve on a handler and listens to incoming requests.
//
// CORS is also configured to work with Stremio.
func Serve(r *http.Handler) error {
	// CORS configuration
	headersOk := handlers.AllowedHeaders([]string{
		"Content-Type",
		"X-Requested-With",
		"Accept",
		"Accept-Language",
		"Accept-Encoding",
		"Content-Language",
		"Origin",
	})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD"})

	// Listen
	logger.LogInfo.Printf("Serve: listening on port %s...\n", config.Port)
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", config.Port), handlers.CORS(originsOk, headersOk, methodsOk)(*r))
	if err != nil {
		return fmt.Errorf("Serve: %w", err)
	}

	return nil
}

// homeHandler handles requests to the root and provides a dummy response.
func homeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonResponse, err := json.Marshal(map[string]any{"path": "/"})
		if err != nil {
			logger.LogError.Printf("homeHandler: failed to marshal json: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}

// manifestHandler handles requests for the Stremio manifest.CacheManager
func manifestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		manifest := config.Manifest

		userConfig := r.Context().Value(UserConfigContextKey).(*stremio.UserConfig)
		if userConfig != nil {
			manifest.BehaviourHints.ConfigurationRequired = false
		} else {
			manifest.BehaviourHints.ConfigurationRequired = true
		}

		jsonResponse, err := json.Marshal(manifest)
		if err != nil {
			logger.LogError.Printf("manifestHandler: failed to marshal json: %v", err)
		}

		logger.LogInfo.Printf("Manifest was: %+v", manifest)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}

// subtitlesHandler handles requests for Titlovi.com search results.
func subtitlesHandler(client *titlovi.Client, cache *ristretto.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		path := r.URL.Path

		userConfig := r.Context().Value(UserConfigContextKey).(*stremio.UserConfig)
		if userConfig == nil {
			logger.LogError.Printf("subtitlesHandler: user config was nil")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, ok := params["type"]
		if !ok {
			logger.LogError.Printf("subtitlesHandler: failed to get 'type' from path, path was %s", path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id, ok := params["id"]
		if !ok {
			logger.LogError.Printf("subtitlesHandler: failed to get 'id' from path, path was %s", path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp := &stremio.SubtitlesResponse{
			Subtitles: []*stremio.SubtitleItem{},
		}

		// Serve the results from the cache if found.
		if val, found := cache.Get(id); found {
			w.Header().Set(config.CacheHeader, config.CacheHit)

			resp, ok = val.(*stremio.SubtitlesResponse)
			if !ok {
				logger.LogFatal.Printf("subtitlesHandler: value found in cache was of an unexpected type")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			w.Header().Set(config.CacheHeader, config.CacheMiss)

			subtitleData, err := client.Search(id, config.TitloviLanguages, userConfig.Username, userConfig.Password)
			if err != nil {
				logger.LogError.Printf("subtitlesHandler: failed to search for subtitles: %s: %s", err, path)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Pre-allocate according to what we got.
			resp.Subtitles = make([]*stremio.SubtitleItem, len(subtitleData))

			for i, data := range subtitleData {
				idStr := strconv.Itoa(int(data.Id))
				servePath := fmt.Sprintf("%s/serve-subtitle/%d/%s", config.ServerAddress, data.Type, idStr)
				resp.Subtitles[i] = &stremio.SubtitleItem{
					Id:   idStr,
					Url:  fmt.Sprintf("http://127.0.0.1:11470/subtitles.vtt?from=%s", servePath),
					Lang: fmt.Sprintf("%s|%s", data.Lang, config.SubtitleSuffix),
				}
				logger.LogInfo.Printf("subtitlesHandler: prepared %+v", *resp.Subtitles[i])
			}

			logger.LogInfo.Printf("subtitlesHandler: got %d subtitles for '%s'", len(resp.Subtitles), id)

			cache.SetWithTTL(id, resp, 0, config.CacheTTL)
		}

		jsonResponse, err := json.Marshal(resp)
		if err != nil {
			logger.LogError.Printf("subtitlesHandler: failed to marshal response: %s", err)
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}

// serveSubtitleHandler handles requests for downloading specific subtitles from Titlovi.com.
func serveSubtitleHandler(client *titlovi.Client, cache *ristretto.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		path := r.URL.Path

		mediaType, ok := params["type"]
		if !ok {
			logger.LogError.Printf("serveSubtitleHandler: failed to get 'type' from path, path was %s", path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mediaId, ok := params["mediaid"]
		if !ok {
			logger.LogError.Printf("serveSubtitleHandler: failed to get 'mediaid' from path, path was %s", path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var subData []byte

		if val, found := cache.Get(fmt.Sprintf("%s-%s", mediaType, mediaId)); found {
			w.Header().Set(config.CacheHeader, config.CacheHit)

			subData, ok = val.([]byte)
			if !ok {
				logger.LogFatal.Printf("subtitlesHandler: value found in cache was of an unexpected type")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			w.Header().Set(config.CacheHeader, config.CacheMiss)

			// We download the subtitle as a blob from Titlovi.com
			data, err := client.Download(mediaType, mediaId)
			if err != nil {
				logger.LogError.Printf("serveSubtitlesHandler: failed to download subtitle: %s: %s", err, path)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Titlovi.com responds with subtitles that are compressed in ZIP files.
			// We need to open this ZIP file and extract the first found subtitle as a byte blob.
			subData, err = utils.ExtractSubtitleFromZIP(data)
			if err != nil {
				logger.LogError.Printf("serveSubtitleHandler: failed to extract subtitle from ZIP: %s: %s", err, path)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			subData, err = utils.ConvertSubtitleToUTF8(subData)
			if err != nil {
				logger.LogError.Printf("serveSubtitleHandler: failed to convert subtitle: %s: %s", err, path)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			cache.SetWithTTL(fmt.Sprintf("%s-%s", mediaType, mediaId), subData, 0, config.CacheTTL)
		}

		logger.LogInfo.Printf("serveSubtitleHandler: serving %s", r.URL.Path)
		http.ServeContent(w, r, "file.srt", time.Now().UTC(), bytes.NewReader(subData))
	}
}

// configureHandler handles requests for addon configuration and redirects to Stremio when done.
func configureHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}

		if r.Method == http.MethodGet {
			if err := config.ConfigTemplate.Execute(w, nil); err != nil {
				logger.LogError.Printf("configureHandler: failed to execute template: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		creds := web.UserConfig{
			Username: r.FormValue("username"),
			Password: r.FormValue("password"),
		}

		if !creds.Validate() {
			if err := config.ConfigTemplate.Execute(w, creds); err != nil {
				logger.LogError.Printf("configureHandler: failed to execute template: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		enc, err := utils.EncodeUserConfig(creds)
		if err != nil {
			logger.LogError.Printf("configureHandler: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		redirectUrl := fmt.Sprintf("stremio://%s/%s/manifest.json", r.Host, enc)
		logger.LogInfo.Printf("configureHandler: redirecting to %s", redirectUrl)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		http.Redirect(w, r, redirectUrl, http.StatusPermanentRedirect)
	}
}
