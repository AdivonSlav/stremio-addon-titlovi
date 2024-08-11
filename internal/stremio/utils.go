package stremio

import "strings"

// ParseVideoId returns the IMDB ID and (if applicable) the season and episode number from a provided Stremio video id.
func ParseVideoId(id string) (imdbId string, season string, episode string) {
	split := strings.Split(id, ":")

	if len(split) == 3 {
		return split[0], split[1], split[2]
	}
	return id, "", ""
}
