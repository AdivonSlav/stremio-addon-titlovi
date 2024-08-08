package config

import (
	"go-titlovi/internal/stremio"
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
	IdPrefixes:  []string{"tt"},
}

var (
	ServerAddress   string = "http://127.0.0.1"
	Port            string = ""
	TitloviUsername string = ""
	TitloviPassword string = ""
	TitloviApi      string = "https://kodi.titlovi.com/api/subtitles"
	TitloviDownload string = "https://titlovi.com/download"
	SubtitleSuffix  string = "Titlovi.com"
)

func InitConfig() {
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
