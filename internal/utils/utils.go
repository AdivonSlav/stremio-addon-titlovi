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
	"net"
	"net/http"
	"strings"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"
)

// GetIP attempts to retrieve the IP through multiple methods from an http.Request.
func GetIP(r *http.Request) (string, error) {
	var err error

	ip := r.Header.Get("X-Forwarded-For")

	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}

	if ip == "" {
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return "", fmt.Errorf("GetIP: %w", err)
		}
	}

	if ip == "" {
		return "", fmt.Errorf("GetIP: no IP found")
	}

	return ip, nil
}

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

// ConvertSubtitleToUTF8 takes subtitle data, determines the charset and converts it to UTF-8.
func ConvertSubtitleToUTF8(subtitleData []byte) (string, error) {
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(subtitleData)
	if err != nil {
		return "", fmt.Errorf("ConvertSubtitleToUTF8: can't detect subtitle encoding: %s", err)
	}

	logger.LogInfo.Printf("Detected %s with confidence of %d", result.Charset, result.Confidence)

	e, err := ianaindex.IANA.Encoding(result.Charset)
	if err != nil {
		return "", fmt.Errorf("ConvertSubtitleToUTF8: failed to retrieve charset from IANA name: %s", err)
	}

	r := transform.NewReader(bytes.NewBuffer(subtitleData), e.NewDecoder())
	utf8, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("ConvertSubtitleToUTF8: failed to read buffer: %s", err)
	}

	return string(utf8), err
}
