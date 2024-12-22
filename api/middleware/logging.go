package middleware

import (
	"fmt"
	"go-titlovi/internal/config"
	"go-titlovi/internal/logger"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Struct that holds response data.
type responseData struct {
	status int
	size   int
}

// Wrapper around  http.ResponseWriter which allows us to capture the response data after the handler function returns.
type loggingResponseWriter struct {
	http.ResponseWriter // compose original http.ResponseWriter
	responseData        *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode) // write status code using original http.ResponseWriter
	r.responseData.status = statusCode       // capture status code
}

func WithLogging(next http.Handler) http.Handler {
	loggingFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := loggingResponseWriter{
			ResponseWriter: w, // compose original http.ResponseWriter
			responseData:   &responseData{},
		}

		next.ServeHTTP(&lw, r)

		duration := time.Since(start)

		cacheStatus := w.Header().Get(config.CacheHeader)
		var msg string

		redactedURL, err := redactURL(r.URL.Path)
		if err != nil {
			logger.LogError.Printf("WithLogging: failed to redact URL: %v", err)
		}

		if cacheStatus != "" {
			msg = fmt.Sprintf("Request: method: %s | status: %d | cache: %s | duration: %s | url: %s",
				r.Method, lw.responseData.status, cacheStatus, duration, redactedURL)
		} else {
			msg = fmt.Sprintf("Request: method: %s | status: %d | duration: %s | url: %s",
				r.Method, lw.responseData.status, duration, redactedURL)
		}

		logger.LogInfo.Print(msg)
	}

	return http.HandlerFunc(loggingFn)
}

// redactURL is used to redact any sensitive parts of a raw URL.
func redactURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("redactURL: failed to parse raw URL: %w", err)
	}

	queryParams := parsedURL.Query()
	queryParams.Set("password", "REDACTED")
	queryParams.Set("token", "REDACTED")

	pathParts := strings.Split(parsedURL.Path, "/")

	// Redact the path var containing the base64-encoded user config.
	if len(pathParts) > 1 {
		pathParts[1] = "REDACTED"
	}

	parsedURL.Path = strings.Join(pathParts, "/")
	parsedURL.RawQuery = queryParams.Encode()

	return parsedURL.String(), nil
}
