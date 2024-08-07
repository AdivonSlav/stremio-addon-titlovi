package main

import (
	"go-titlovi/stremio"
	"log"
	"os"
)

var Manifest = stremio.Manifest{
	Id:          "com.adivonslav.go-titlovi.test",
	Version:     "0.0.1",
	Name:        "Go Titlovi",
	Description: "Test addon for fetching Titlovi.com subtitles",
	Types:       []string{"movie", "series"},
	Resources:   []string{"subtitles"},
}

var (
	Port            string = ""
	TitloviUsername string = ""
	TitloviPassword string = ""
)

func initConfig() {
	Port = os.Getenv("PORT")
	if Port == "" {
		log.Fatalf("The environment variable PORT must be supplied\n")
	}
	TitloviUsername = os.Getenv("TITLOVI_USERNAME")

	if TitloviUsername == "" {
		log.Fatalf("The environment variable TITLOVI_USERNAME must be supplied\n")
	}

	TitloviPassword = os.Getenv("TITLOVI_PASSWORD")
	if TitloviPassword == "" {
		log.Fatalf("The environment variable TITLOVI_PASSWORD must be supplied\n")
	}
}
