package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func buildRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/manifest.json", manifestHandler)

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
	log.Printf("Listening on port %s...\n", Port)
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", Port), handlers.CORS(originsOk, headersOk, methodsOk)(r))
	if err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse, err := json.Marshal(map[string]any{"Path": "/"})
	if err != nil {
		log.Printf("Failed to marshal json: %v", err)
	}

	log.Printf("Received request to %s\n", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func manifestHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse, err := json.Marshal(Manifest)
	if err != nil {
		log.Printf("Failed to marshal json: %v", err)
	}

	log.Printf("Received request to %s\n", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func subtitlesHandler(w http.ResponseWriter, r *http.Request) {

}






