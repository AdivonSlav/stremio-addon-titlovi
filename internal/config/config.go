package config

import (
	"go-titlovi/internal/stremio"
	"log"
	"os"
	"time"

	"github.com/victorspringer/http-cache/adapter/memory"
)

// This is the manifest of the addon that will be shown to Stremio
// in order to describe some general information about the addon.
var Manifest = stremio.Manifest{
	Id:          "com.github.titlovi-unofficial.stremio",
	Version:     "0.0.1",
	Name:        "Titlovi.com Unofficial",
	Description: "Unofficial addon for fetching subtitles from Titlovi.com.",
	Types:       []string{"movie", "series"},
	Resources:   []string{"subtitles"},
	IdPrefixes:  []string{"tt"},
}

var (
	ServerAddress string = "http://127.0.0.1"
	Port          string = ""

	TitloviUsername string = ""                                       // Username for the Titlovi.com account.
	TitloviPassword string = ""                                       // Password for the Titlovi.com account.
	TitloviApi      string = "https://kodi.titlovi.com/api/subtitles" // Titlovi.com API where we can search for subtitles.
	TitloviDownload string = "https://titlovi.com/download"           // URL where subtitles can be downloaded from Titlovi.com.

	TitloviLanguages = []string{ // The languages to query for on Titlovi.com.
		"Bosanski",
		"Hrvatski",
		"Srpski",
		"Cirilica",
		"English",
		"Makedonski",
		"Slovenski",
	}

	SubtitleSuffix string = "Titlovi.com" // This will be appended as a suffix to subtitle languages when returned to Stremio.
)

const (
	MemoryCacheTTL          time.Duration    = 10 * time.Minute // How long to cache a single response.
	MemoryCacheAlgorithm    memory.Algorithm = memory.LRU       // Caching algorithm to use.
	MemoryCacheMaxResponses int              = 10000000         // Max number of responses to cache.
	MemoryCacheRefreshKey   string           = "opn"            // If this key is provided, cache is circumvented.

	TitloviClientRetryAttempts uint          = 3                      // How many times to retry a failed request to Titlovi.com.
	TitloviClientRetryDelay    time.Duration = 500 * time.Millisecond // The delay in-between retries for requests to Titlovi.com.
)

// InitConfig initializes some global variables from the environment.
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
