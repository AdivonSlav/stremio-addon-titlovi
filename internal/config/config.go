package config

import (
	"fmt"
	"go-titlovi/internal/logger"
	"go-titlovi/internal/stremio"
	"html/template"
	"os"
	"strconv"
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
	BehaviourHints: stremio.BehaviourHints{
		Configurable:          true,
		ConfigurationRequired: true,
	},
}

var (
	Development bool = false

	ServerAddress            string = ""
	ConfigureRedirectAddress string = ""

	Port string = ""

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

	ConfigTemplate *template.Template = template.Must(template.ParseFiles("web/templates/configuration-form.html"))
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
	var err error

	Port = os.Getenv("PORT")
	if Port == "" {
		logger.LogFatal.Fatalf("InitConfig: The environment variable PORT must be supplied")
	}

	isDev := os.Getenv("DEVELOPMENT")
	if isDev == "" {
		Development = false
	} else {
		Development, err = strconv.ParseBool(isDev)
		if err != nil {
			logger.LogFatal.Fatalf("InitConfig: cannot set DEVELOPMENT: %s", err)
		}
	}

	ServerAddress = os.Getenv("SERVER_ADDRESS")
	if ServerAddress == "" {
		ServerAddress = fmt.Sprintf("http://127.0.0.1:%s", Port)
	}

	ConfigureRedirectAddress = os.Getenv("REDIRECT_ADDRESS")
	if ConfigureRedirectAddress == "" {
		ConfigureRedirectAddress = fmt.Sprintf("stremio://127.0.0.1:%s", Port)
	}
}
