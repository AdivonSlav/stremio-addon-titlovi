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

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	cache "github.com/victorspringer/http-cache"
)

// BuildRouter builds a new router with handler functions to handle all necessary routes and
// also appends middleware.
func BuildRouter(client *titlovi.Client, cache *cache.Client) http.Handler {
	r := mux.NewRouter()

	// Route handlers
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/configure", configureHandler)
	r.HandleFunc("/{creds}/manifest.json", manifestHandler)
	r.HandleFunc("/serve-subtitle/{type}/{mediaid}", func(w http.ResponseWriter, r *http.Request) {
		serveSubtitleHandler(w, r, client)
	})
	r.HandleFunc("/{creds}/subtitles/{type}/{id}/{extraArgs}.json", func(w http.ResponseWriter, r *http.Request) {
		subtitlesHandler(w, r, client)
	})

	// Create a caching router that wraps the main router
	cachedRouter := mux.NewRouter()
	cachedRouter.PathPrefix("/").Handler(r)

	// Apply caching middleware and timeout handler to the caching router
	cachedHandler := cache.Middleware(cachedRouter)
	timeoutHandler := http.TimeoutHandler(cachedHandler, 30*time.Second, "")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/configure" {
			configureHandler(w, r)
		} else {
			// Apply caching and timeout middleware to other routes
			timeoutHandler.ServeHTTP(w, r)
		}
	})
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
	methodsOk := handlers.AllowedMethods([]string{"GET"})

	// Listen
	logger.LogInfo.Printf("Serve: listening on port %s...\n", config.Port)
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", config.Port), handlers.CORS(originsOk, headersOk, methodsOk)(*r))
	if err != nil {
		return fmt.Errorf("Serve: %w", err)
	}

	return nil
}

// homeHandler handles requests to the root and provides a dummy response.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse, err := json.Marshal(map[string]any{"path": "/"})
	if err != nil {
		logger.LogError.Printf("homeHandler: failed to marshal json: %v", err)
	}

	logger.LogInfo.Printf("homeHandler: received request to %s", r.URL.Path)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

// manifestHandler handles requests for the Stremio manifest.
func manifestHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	logger.LogInfo.Printf("manifestHandler: received request to %s", r.URL.Path)

	manifest := config.Manifest

	if _, ok := params["creds"]; !ok {
		manifest.BehaviourHints.ConfigurationRequired = true
	} else {
		manifest.BehaviourHints.ConfigurationRequired = false
	}

	jsonResponse, err := json.Marshal(manifest)
	if err != nil {
		logger.LogError.Printf("manifestHandler: failed to marshal json: %v", err)
	}

	logger.LogInfo.Printf("Manifest was: %+v", manifest)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

// subtitlesHandler handles requests for Titlovi.com search results.
func subtitlesHandler(w http.ResponseWriter, r *http.Request, client *titlovi.Client) {
	path := r.URL.Path
	params := mux.Vars(r)

	logger.LogInfo.Printf("subtitlesHandler: received request to %s", r.URL.Path)

	credsEnc, ok := params["creds"]
	if !ok {
		logger.LogError.Printf("subtitlesHandler: failed to get 'creds' from path, path was %s", path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	creds, err := utils.DecodeCreds(credsEnc)
	if err != nil {
		logger.LogError.Printf("subtitlesHandler: invalid creds: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, ok = params["type"]
	if !ok {
		logger.LogError.Printf("subtitlesHandler: failed to get 'type' from path, path was %s", path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	imdbId, ok := params["id"]
	if !ok {
		logger.LogError.Printf("subtitlesHandler: failed to get 'id' from path, path was %s", path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	subtitleData, err := client.Search(imdbId, config.TitloviLanguages, creds.Username, creds.Password)
	if err != nil {
		logger.LogError.Printf("subtitlesHandler: failed to search for subtitles: %s: %s", err, path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	subtitles := make([]*stremio.SubtitleItem, len(subtitleData))

	for i, data := range subtitleData {
		idStr := strconv.Itoa(int(data.Id))
		servePath := fmt.Sprintf("%s/serve-subtitle/%d/%s", config.ServerAddress, data.Type, idStr)
		subtitles[i] = &stremio.SubtitleItem{
			Id:   idStr,
			Url:  fmt.Sprintf("http://127.0.0.1:11470/subtitles.vtt?from=%s", servePath),
			Lang: fmt.Sprintf("%s|%s", data.Lang, config.SubtitleSuffix),
		}
		logger.LogInfo.Printf("subtitlesHandler: prepared %+v", subtitles[i])
	}

	logger.LogInfo.Printf("subtitlesHandler: got %d subtitles for '%s'", len(subtitles), imdbId)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	jsonResponse, _ := json.Marshal(map[string]any{
		"subtitles": subtitles,
	})
	w.Write(jsonResponse)
}

// serveSubtitleHandler handles requests for downloading specific subtitles from Titlovi.com.
func serveSubtitleHandler(w http.ResponseWriter, r *http.Request, client *titlovi.Client) {
	params := mux.Vars(r)
	path := r.URL.Path

	logger.LogInfo.Printf("subtitlesHandler: received request to %s", path)

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

	// We download the subtitle as a blob from Titlovi.com
	data, err := client.Download(mediaType, mediaId)
	if err != nil {
		logger.LogError.Printf("serveSubtitlesHandler: failed to download subtitle: %s: %s", err, path)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Titlovi.com responds with subtitles that are compressed in ZIP files.
	// We need to open this ZIP file and extract the first found subtitle as a byte blob.
	subData, err := utils.ExtractSubtitleFromZIP(data)
	if err != nil {
		logger.LogError.Printf("serveSubtitleHandler: failed to extract subtitle from ZIP: %s: %s", err, path)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Then, to make sure Stremio has no issues, we take the subtitle and convert it to VTT.
	// The conversion also ensures UTF-8(?)
	convertedSubData, err := utils.ConvertSubtitleToVTT(subData)
	if err != nil {
		logger.LogError.Printf("serveSubtitleHandler: failed to convert subtitle: %s: %s", err, path)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.LogInfo.Printf("serveSubtitleHandler: serving %s", r.URL.Path)
	http.ServeContent(w, r, "file.vtt", time.Now().UTC(), bytes.NewReader(convertedSubData.Bytes()))
}

// configureHandler handles requests for addon configuration and redirects to Stremio when done.
func configureHandler(w http.ResponseWriter, r *http.Request) {
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

	creds := web.Credentials{
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

	redirectUrl := fmt.Sprintf("%s/%s/manifest.json", config.ConfigureRedirectAddress, utils.EncodeCreds(creds))
	logger.LogInfo.Printf("configureHandler: redirecting to %s", redirectUrl)

	http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
}
