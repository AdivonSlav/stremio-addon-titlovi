package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-titlovi/logger"
	"go-titlovi/stremio"
	"go-titlovi/titlovi"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/allegro/bigcache"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func BuildRouter(client *titlovi.Client, cache *bigcache.BigCache) *http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/manifest.json", manifestHandler)
	r.HandleFunc("/serve-subtitle/{type}/{mediaid}", serveSubtitleHandler)
	r.HandleFunc("/subtitles/{type}/{id}/{extraArgs}.json", func(w http.ResponseWriter, r *http.Request) {
		subtitlesHandler(w, r, client, cache)
	})

	http.Handle("/", r)

	handler := http.TimeoutHandler(r, 30*time.Second, "")

	return &handler
}

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
	logger.LogInfo.Printf("Serve: Listening on port %s...\n", Port)
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", Port), handlers.CORS(originsOk, headersOk, methodsOk)(*r))
	if err != nil {
		return fmt.Errorf("Serve: %w", err)
	}

	return nil
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse, err := json.Marshal(map[string]any{"Path": "/"})
	if err != nil {
		logger.LogError.Printf("Failed to marshal json: %v", err)
	}

	logger.LogInfo.Printf("Received request to %s\n", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func manifestHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse, err := json.Marshal(Manifest)
	if err != nil {
		logger.LogError.Printf("Failed to marshal json: %v", err)
	}

	logger.LogInfo.Printf("Received request to %s\n", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func subtitlesHandler(w http.ResponseWriter, r *http.Request, client *titlovi.Client, cache *bigcache.BigCache) {
	path := r.URL.Path
	params := mux.Vars(r)

	logger.LogInfo.Printf("Received request to %s\n", r.URL.Path)

	_, ok := params["type"]
	if !ok {
		logger.LogError.Printf("subtitlesHandler: failed to get 'type' from path, path was %s\n", path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	imdbId, ok := params["id"]
	if !ok {
		logger.LogError.Printf("subtitlesHandler: failed to get 'type' from path, path was %s\n", path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	subtitleData, err := client.Search(imdbId, GetLanguagesToQuery())
	if err != nil {
		logger.LogError.Printf("subtitlesHandler: failed to search for subtitles: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	subtitles := make([]*stremio.SubtitleItem, len(subtitleData))

	for i, data := range subtitleData {
		idStr := strconv.Itoa(int(data.Id))
		servePath := fmt.Sprintf("%s:%s/serve-subtitle/%d/%s", ServerAddress, Port, data.Type, idStr)
		subtitles[i] = &stremio.SubtitleItem{
			Id:   idStr,
			Url:  fmt.Sprintf("http://127.0.0.1:11470/subtitles.vtt?from=%s", servePath),
			Lang: fmt.Sprintf("%s|%s", data.Lang, SubtitleSuffix),
		}
		logger.LogInfo.Printf("subtitlesHandler: prepared %+v", subtitles[i])
	}

	logger.LogInfo.Printf("subtitlesHandler: got %d subtitles for '%s'", len(subtitles), imdbId)

	// CacheSubtitles(imdbId, cache, subtitles)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	jsonResponse, _ := json.Marshal(map[string]any{
		"subtitles": subtitles,
	})
	w.Write(jsonResponse)
}

func serveSubtitleHandler(w http.ResponseWriter, r *http.Request) {
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

	url := fmt.Sprintf("%s/?type=%s&mediaid=%s", TitloviDownload, mediaType, mediaId)
	resp, err := http.Get(url)
	if err != nil {
		logger.LogError.Printf("serveSubtitleHandler: failed to download subtitle at %s: %s", url, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		logger.LogError.Printf("serveSubtitleHandler: status %d, %s: %s", resp.StatusCode, url, resp.Status)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.LogError.Printf("serveSubtitleHandler: failed to read response body from %s: %s", url, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	subData, err := ExtractSubtitleFromZIP(data)
	if err != nil {
		logger.LogError.Printf("serveSubtitleHandler: failed to extract subtitle from ZIP from %s: %s", url, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	convertedSubData, err := ConvertSubtitleToVTT(subData)
	if err != nil {
		logger.LogError.Printf("serveSubtitleHandler: failed to convert subtitle from %s: %s", url, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.LogInfo.Printf("serveSubtitleHandler: serving %s", url)
	http.ServeContent(w, r, "file.vtt", time.Now().UTC(), bytes.NewReader(convertedSubData.Bytes()))
}
