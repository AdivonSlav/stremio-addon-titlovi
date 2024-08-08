package common

import (
	"encoding/json"
	"fmt"
	"go-titlovi/stremio"

	"github.com/allegro/bigcache"
)

var langMap = map[string]string{
	"Bosanski": "bos",
	"Hrvatski": "hrv",
	"Srpski":   "srb",
	"Cirilica": "cpb",
	"English":  "eng",
}

func ConvertLangToISO(language string) string {
	code, ok := langMap[language]
	if !ok {
		return ""
	}

	return code
}

func GetLanguagesToQuery() []string {
	keys := make([]string, len(langMap))

	i := 0
	for k := range langMap {
		keys[i] = k
		i++
	}

	return keys
}

func CacheSubtitles(imdbId string, cache *bigcache.BigCache, subtitles []stremio.SubtitleItem) error {
	data, err := json.Marshal(subtitles)
	if err != nil {
		return fmt.Errorf("CacheSubtitles: %w", err)
	}

	cache.Set(imdbId, data)

	return nil
}
