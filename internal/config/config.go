package config

import (
	"fmt"
	"go-titlovi/internal/logger"
	"go-titlovi/internal/stremio"
	"html/template"
	"os"
	"strconv"
	"time"
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
	Catalogs:    []stremio.CatalogItem{},
	BehaviourHints: stremio.BehaviourHints{
		Configurable:          true,
		ConfigurationRequired: true,
	},
}

var (
	Development   bool   = false
	ServerAddress string = ""
	Port          string = ""

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
	CacheTTL         time.Duration = 60 * time.Minute // How long does a value stay in the cache before being evicted.
	CacheNumCounters int64         = 1e7              // How many counters will the cache instantiate. See https://pkg.go.dev/github.com/dgraph-io/ristretto#readme-Config
	CacheMaxCost     int64         = 1 << 28          // Max size of the cache in bytes. Roughly 256MB. See https://pkg.go.dev/github.com/dgraph-io/ristretto#readme-Config
	CacheBufferItems int64         = 64               // Max size of the get buffer for the cache. See https://pkg.go.dev/github.com/dgraph-io/ristretto#readme-Config

	CacheHeader string = "Cache-Status" // Header to set to indicate cache status.
	CacheHit    string = "HIT"          // Set if the cache was hit.
	CacheMiss   string = "MISS"         // Set if not hit.

	TitloviClientRetryAttempts uint          = 3                      // How many times to retry a failed request to Titlovi.com.
	TitloviClientRetryDelay    time.Duration = 500 * time.Millisecond // The delay in-between retries for requests to Titlovi.com.

	RateLimitingRate        int           = 2               // How many requests to allow within a second.
	RateLimitingBurst       int           = 3               // How many burst requests do we allow.
	RateLimitingCleanupTime time.Duration = 3 * time.Minute // The duration to hold a single rate limiter for a client for. After this, it is deleted.
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
}
