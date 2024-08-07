package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	log.Println("Initializing...")

	r := mux.NewRouter()

	r.HandleFunc("/", HomeHandler)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse, err := json.Marshal(map[string]any{"Path": "/"})
	if err != nil {
		log.Printf("Failed to marshal json: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

