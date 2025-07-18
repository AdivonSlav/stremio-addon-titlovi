package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-titlovi/internal/stremio"
	"go-titlovi/web"
	"net"
	"net/http"
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
