package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-titlovi/internal/config"
	"go-titlovi/internal/logger"
	"go-titlovi/internal/stremio"
	"go-titlovi/web"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// client is a wrapper around rate.Limiter that also holds info on when the client was last seen.
type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// A map that holds a limiter for each client that made a request. The keys are IP addresses.
var clients = make(map[string]*client)
var mtx sync.RWMutex

// getLimiterFromClient returns a rate.Limiter for a specific IP address from the clients map.
func getLimiterFromClient(ip string) *rate.Limiter {
	mtx.RLock()
	c, ok := clients[ip]
	mtx.RUnlock()

	if !ok {
		limiter := rate.NewLimiter(rate.Limit(config.RateLimitingRate), config.RateLimitingBurst)

		mtx.Lock()
		clients[ip] = &client{limiter: limiter, lastSeen: time.Now()}
		mtx.Unlock()

		return limiter
	}

	mtx.Lock()
	c.lastSeen = time.Now()
	mtx.Unlock()

	return c.limiter
}

// cleanupLimiters is meant to be used in a goroutine to periodically clear unused limiters.
func cleanupLimiters(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// We could technically introduce read locks here for a more granular approach.
			var toDelete []string
			mtx.RLock()
			for ip, c := range clients {
				if time.Since(c.lastSeen) > config.RateLimitingCleanupTime {
					toDelete = append(toDelete, ip)
				}
			}
			mtx.RUnlock()

			if len(toDelete) < 0 {
				for _, ip := range toDelete {
					delete(clients, ip)
				}
				mtx.Unlock()
			}
		}
	}
}

// InitRateLimitCleanup initializes a cleanup goroutine to periodically clear limiters for rate limiting.
func InitRateLimitCleanup(ctx context.Context) {
	go cleanupLimiters(ctx)
}

// WithRateLimit applies token bucket rate limiting to a handler.
func WithRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, err := getIP(r)
		if err != nil {
			logger.LogError.Printf("WithRateLimit: could not retrieve IP: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		limiter := getLimiterFromClient(ip)
		if !limiter.Allow() {
			logger.LogInfo.Printf("WithRateLimit: rate-limited %s", ip)
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// EncodeUserConfig encodes web.UserConfig received from the configuration page to a base64 JSON representation of a stremio.UserConfig.
func EncodeUserConfig(c web.UserConfig) (string, error) {
	config := &stremio.UserConfig{
		Username: c.Username,
		Password: c.Password,
	}

	json, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("marshal user config struct: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString([]byte(json)), nil
}

// DecodeUserConfig decodes a base64 JSON object into a stremio.UserConfig
func DecodeUserConfig(c string) (*stremio.UserConfig, error) {
	data, err := base64.RawURLEncoding.DecodeString(c)
	if err != nil {
		return nil, fmt.Errorf("decode user config: %w", err)
	}

	var userConfig = &stremio.UserConfig{}
	err = json.Unmarshal(data, userConfig)
	if err != nil {
		return nil, fmt.Errorf("unmarshal user config struct: %w", err)
	}

	return userConfig, nil
}

// getIP attempts to retrieve the IP through multiple methods from an http.Request.
func getIP(r *http.Request) (string, error) {
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
