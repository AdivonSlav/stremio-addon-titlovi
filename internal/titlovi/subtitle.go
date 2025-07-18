package titlovi

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"
)

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
	e, _ := ianaindex.IANA.Encoding("windows-1250")

	r := transform.NewReader(bytes.NewBuffer(subtitleData), e.NewDecoder())
	utf8, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read buffer: %w", err)
	}

	return string(utf8), err
}
