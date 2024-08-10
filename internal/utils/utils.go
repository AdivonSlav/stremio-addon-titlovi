package utils

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-titlovi/internal/logger"
	"go-titlovi/internal/stremio"
	"go-titlovi/web"
	"io"
	"strings"

	"golang.org/x/net/html/charset"
)

// EncodeCreds encodes web.UserConfig received from the configuration page to a base64 JSON representation of a stremio.UserConfig.
func EncodeUserConfig(c web.UserConfig) (string, error) {
	config := &stremio.UserConfig{
		Username: c.Username,
		Password: c.Password,
	}

	json, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("EncodeUserConfig: could not marshal user config: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString([]byte(json)), nil
}

// DecodeUserConfig decodes a base64 JSON object into a stremio.UserConfig
func DecodeUserConfig(c string) (*stremio.UserConfig, error) {
	data, err := base64.RawURLEncoding.DecodeString(c)
	if err != nil {
		return nil, fmt.Errorf("DecodeUserConfig: failed to decode user config: %w", err)
	}

	var userConfig = &stremio.UserConfig{}
	err = json.Unmarshal(data, userConfig)
	if err != nil {
		return nil, fmt.Errorf("DecodeUserConfig: failed to unmarshal user config: %w", err)
	}

	return userConfig, nil
}

// ExtractSubtitleFromZIP extracts the first found subtitle found from ZIP file.
func ExtractSubtitleFromZIP(zipData []byte) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("ExtractSubtitleFromZIP: failed to read subtitle ZIP file: %w", err)
	}

	// Only look for SRT files
	desiredExtension := ".srt"
	var buffer []byte

	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, desiredExtension) {
			zipFile, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("ExtractSubtitleFromZIP: failed to open subtitle '%s' from zip file: %w", file.Name, err)
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

// ConvertSubtitleToUTF8 converts subtitle data to UTF-8.
func ConvertSubtitleToUTF8(subtitleData []byte) ([]byte, error) {
	e, name, _ := charset.DetermineEncoding(subtitleData, "")
	logger.LogInfo.Printf("ConvertSubtitleToUTF8: fallback to determing encoding, determined %s", name)

	utf8, err := e.NewDecoder().Bytes(subtitleData)
	if err != nil {
		return nil, fmt.Errorf("ConvertSubtitleToUTF8: %w", err)
	}

	return utf8, nil
}
