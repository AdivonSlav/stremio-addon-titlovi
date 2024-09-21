package stremio

import "strings"

var langCodes = map[string]string{
	"Bosanski":   "bos",
	"Hrvatski":   "hrv",
	"Srpski":     "srp",
	"Cirilica":   "cir",
	"English":    "eng",
	"Makedonski": "mkd",
	"Slovenski":  "slv",
}

// ParseVideoId returns the IMDB ID and (if applicable) the season and episode number from a provided Stremio video id.
func ParseVideoId(id string) (imdbId string, season string, episode string) {
	split := strings.Split(id, ":")

	if len(split) == 3 {
		return split[0], split[1], split[2]
	}
	return id, "", ""
}

// GetLangCode returns the ISO 639-1 code for a given language.
func GetLangCode(lang string) string {
	return langCodes[lang]
}
