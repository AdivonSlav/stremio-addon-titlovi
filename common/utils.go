package common

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"go-titlovi/stremio"
	"io"
	"strings"

	"github.com/allegro/bigcache"
	"github.com/asticode/go-astisub"
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

func CacheSubtitles(imdbId string, cache *bigcache.BigCache, subtitles []*stremio.SubtitleItem) error {
	data, err := json.Marshal(subtitles)
	if err != nil {
		return fmt.Errorf("CacheSubtitles: %w", err)
	}

	cache.Set(imdbId, data)

	return nil
}

func GetSubtitlesFromCache(imdbId string, cache *bigcache.BigCache) ([]*stremio.SubtitleItem, error) {
	data, err := cache.Get(imdbId)
	if err != nil {
		return nil, fmt.Errorf("GetSubtitlesFromCache: %w", err)
	}

	var subtitles []*stremio.SubtitleItem
	err = json.Unmarshal(data, &subtitles)
	if err != nil {
		return nil, fmt.Errorf("GetSubtitlesFromCache: failed to unmarshal subtitles from cache: %w", err)
	}

	return subtitles, nil
}

func ExtractSubtitleFromZIP(zipData []byte) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("ExtractSubtitleFromZIP: failed to read subtitle ZIP file: %w", err)
	}

	desiredExtension := ".srt"
	var buffer []byte

	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, desiredExtension) {
			zipFile, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("ExtractSubtitleFromZIP: failed to opem subtitle '%s' from zip file: %w", file.Name, err)
			}
			defer zipFile.Close()

			buffer, err = io.ReadAll(zipFile)
			if err != nil {
				return nil, fmt.Errorf("ExtractSubtitleFromZIP: failed to read subtitle '%s' from zip file: %w", file.Name, err)
			}
			break
		}
	}

	return buffer, nil
}

func ConvertSubtitleToVTT(subtitleData []byte) (*bytes.Buffer, error) {
	subtitle, err := astisub.ReadFromSRT(bytes.NewReader(subtitleData))
	if err != nil {
		return nil, fmt.Errorf("ConvertSubtitleToVTT: failed to read subtitle: %w", err)
	}

	var buf = &bytes.Buffer{}
	err = subtitle.WriteToWebVTT(buf)
	if err != nil {
		return nil, fmt.Errorf("ConvertSubtitleToVTT: failed to write subtitle as VTT: %w", err)
	}

	return buf, nil
}
