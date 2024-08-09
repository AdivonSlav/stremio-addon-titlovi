package utils

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"go-titlovi/web"
	"io"
	"strings"

	"github.com/asticode/go-astisub"
)

// EncodeCreds encodes credentials into a concatenated base64 string.
func EncodeCreds(c web.Credentials) string {
	creds := fmt.Sprintf("%s:%s", c.Username, c.Password)
	return base64.RawURLEncoding.EncodeToString([]byte(creds))
}

func DecodeCreds(c string) (*web.Credentials, error) {
	data, err := base64.RawURLEncoding.DecodeString(c)
	if err != nil {
		return nil, fmt.Errorf("DecodeCreds: %w", err)
	}

	split := strings.Split(string(data), ":")
	if len(split) != 2 {
		return nil, errors.New("DecodeCreds: the decoded credentials were not formatted correctly")
	}

	return &web.Credentials{
		Username: split[0],
		Password: split[1],
	}, nil
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
