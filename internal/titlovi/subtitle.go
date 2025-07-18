package titlovi

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"go-titlovi/internal/logger"
	"io"
	"strings"

	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"
)

// ExtractSubtitleFromZIP extracts the first found subtitle found from ZIP file.
//
// Returns the subtitle as byte data or an error if extraction fails.
func ExtractSubtitleFromZIP(zipData []byte) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("read zip: %w", err)
	}

	// Only look for SRT files
	desiredExtension := ".srt"
	var buffer []byte

	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, desiredExtension) {
			zipFile, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("open file: %w", err)
			}

			buffer, err = io.ReadAll(zipFile)
			if err != nil {
				return nil, fmt.Errorf("read file: %w", err)
			}
			err = zipFile.Close()
			if err != nil {
				logger.LogError.Printf("ExtractSubtitleFromZIP: failed to close file from ZIP")
			}

			return buffer, nil
		}
	}

	return nil, errors.New("no subtitle found in zip")
}

// ConvertSubtitleToUTF8 takes subtitle data, determines the charset and converts it to UTF-8.
//
// Returns the converted subtitle data or an error if conversion fails.
func ConvertSubtitleToUTF8(subtitleData []byte) ([]byte, error) {
	e, err := ianaindex.IANA.Encoding("windows-1250")
	if err != nil || e == nil {
		return nil, fmt.Errorf("encoding not found: %w", err)
	}

	r := transform.NewReader(bytes.NewBuffer(subtitleData), e.NewDecoder())
	utf8, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read buffer: %w", err)
	}

	return utf8, err
}
