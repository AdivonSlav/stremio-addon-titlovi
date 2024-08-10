package api

import (
	"context"
	"go-titlovi/internal/logger"
	"go-titlovi/internal/utils"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	cache "github.com/victorspringer/http-cache"
)

type contextKey string

const (
	UserConfigContextKey contextKey = "user-config"
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

		logger.LogInfo.Printf("Request: method: %s | status: %d | duration: %s | url: %s",
			r.Method, lw.responseData.status, duration, r.URL.Path)
	}

	return http.HandlerFunc(loggingFn)
}

func WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		userConfigEnc, ok := vars["userConfig"]
		if !ok {
			http.Error(w, "No user config passed", http.StatusUnauthorized)
			return
		}

		userConfig, err := utils.DecodeUserConfig(userConfigEnc)
		if err != nil {
			http.Error(w, "Cannot decode user config", http.StatusUnauthorized)
			return
		}

		if userConfig.Username == "" || userConfig.Password == "" {
			http.Error(w, "user config was invalid", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserConfigContextKey, userConfig)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithCaching(next http.Handler, cache *cache.Client) http.Handler {
	return cache.Middleware(next)
}
