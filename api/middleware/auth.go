package middleware

import (
	"context"
	"go-titlovi/internal/utils"
	"net/http"

	"github.com/gorilla/mux"
)

type contextKey string

const (
	UserConfigContextKey contextKey = "user-config"
)

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
