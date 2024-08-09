package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/asticode/go-astisub"
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

// ConvertSubtitleToVTT converts a subtitle to VTT and returns it as a byte buffer.
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
