package middleware

import (
	"go-titlovi/internal/config"
	"go-titlovi/internal/logger"
	"go-titlovi/internal/utils"
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
func cleanupLimiters() {
	for {
		time.Sleep(time.Minute)

		// We could technically introduce read locks here for a more granular approach.
		mtx.Lock()
		for ip, c := range clients {
			if time.Since(c.lastSeen) > config.RateLimitingCleanupTime {
				delete(clients, ip)
			}
		}
		mtx.Unlock()
	}
}

// WithRateLimit applies token bucket rate limiting to a handler.
func WithRateLimit(next http.Handler) http.Handler {
	go cleanupLimiters()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, err := utils.GetIP(r)
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
