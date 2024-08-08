package main

import (
	"encoding/json"
	"fmt"
	"go-titlovi/logger"
	"go-titlovi/stremio"
	"go-titlovi/titlovi"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func buildRouter(client *titlovi.Client) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/manifest.json", manifestHandler)
	r.HandleFunc("/subtitles/{type}/{id}/{extraArgs}.json", func(w http.ResponseWriter, r *http.Request) {
		subtitlesHandler(w, r, client)
	})

	http.Handle("/", r)

	return r
}

func serve(r *mux.Router) error {
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
	logger.LogInfo.Printf("Listening on port %s...\n", Port)
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", Port), handlers.CORS(originsOk, headersOk, methodsOk)(r))
	if err != nil {
		return fmt.Errorf("serve: %w", err)
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

func subtitlesHandler(w http.ResponseWriter, r *http.Request, client *titlovi.Client) {
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

	err := client.Login(false)
	if err != nil {
		logger.LogError.Printf("subtitlesHandler: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	subtitleData, err := client.Search(imdbId, GetLanguagesToQuery())
	if err != nil {
		logger.LogError.Printf("subtitlesHandler: failed to search for subtitles: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	subtitles := make([]stremio.SubtitleItem, len(subtitleData))

	for i, data := range subtitleData {
		subtitles[i] = stremio.SubtitleItem{
			Id:   string(data.Id),
			Url:  data.Link,
			Lang: ConvertLangToISO(data.Lang),
		}

	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse, _ := json.Marshal(map[string]any{
		"subtitles": subtitles,
	})
	w.Write(jsonResponse)
}
