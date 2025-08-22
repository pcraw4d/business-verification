package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	RequestsPerHour   int
	BurstSize         int
	EnableRateLimit   bool
	ExemptPaths       []string // Paths that don't require rate limiting
}

// RateLimiter represents a rate limiter for a specific client
type RateLimiter struct {
	requests     []time.Time
	lastReset    time.Time
	mu           sync.RWMutex
	config       RateLimitConfig
}

// RateLimitStore holds rate limiters for different clients
type RateLimitStore struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
	config   RateLimitConfig
}

// NewRateLimitStore creates a new rate limit store
func NewRateLimitStore(config RateLimitConfig) *RateLimitStore {
	return &RateLimitStore{
		limiters: make(map[string]*RateLimiter),
		config:   config,
	}
}

// getClientIdentifier returns a unique identifier for the client
func getClientIdentifier(r *http.Request) string {
	// Try to get from X-Forwarded-For header first (for proxy scenarios)
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		return forwardedFor
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// RateLimitMiddleware provides rate limiting for API endpoints
func RateLimitMiddleware(store *RateLimitStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path is exempt from rate limiting
			if isExemptPath(r.URL.Path, store.config.ExemptPaths) {
				next.ServeHTTP(w, r)
				return
			}

			if !store.config.EnableRateLimit {
				next.ServeHTTP(w, r)
				return
			}

			clientID := getClientIdentifier(r)
			limiter := store.getLimiter(clientID)

			if !limiter.allow() {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", store.config.RequestsPerMinute))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"success": false, "error": "Rate limit exceeded", "meta": {"retry_after": 60}}`))
				return
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", store.config.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", limiter.remaining()))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))

			next.ServeHTTP(w, r)
		})
	}
}

// getLimiter gets or creates a rate limiter for a client
func (store *RateLimitStore) getLimiter(clientID string) *RateLimiter {
	store.mu.RLock()
	limiter, exists := store.limiters[clientID]
	store.mu.RUnlock()

	if exists {
		return limiter
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists = store.limiters[clientID]; exists {
		return limiter
	}

	limiter = &RateLimiter{
		requests:  make([]time.Time, 0),
		lastReset: time.Now(),
		config:    store.config,
	}

	store.limiters[clientID] = limiter
	return limiter
}

// allow checks if a request is allowed
func (rl *RateLimiter) allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Clean up old requests (older than 1 minute)
	cutoff := now.Add(-time.Minute)
	var validRequests []time.Time
	for _, reqTime := range rl.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	rl.requests = validRequests

	// Check if we're within the rate limit
	if len(rl.requests) >= rl.config.RequestsPerMinute {
		return false
	}

	// Add current request
	rl.requests = append(rl.requests, now)
	return true
}

// remaining returns the number of remaining requests
func (rl *RateLimiter) remaining() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-time.Minute)
	
	var validRequests int
	for _, reqTime := range rl.requests {
		if reqTime.After(cutoff) {
			validRequests++
		}
	}

	return rl.config.RequestsPerMinute - validRequests
}

// cleanup removes old rate limiters to prevent memory leaks
func (store *RateLimitStore) cleanup() {
	store.mu.Lock()
	defer store.mu.Unlock()

	cutoff := time.Now().Add(-time.Hour)
	for clientID, limiter := range store.limiters {
		limiter.mu.RLock()
		hasRecentRequests := false
		for _, reqTime := range limiter.requests {
			if reqTime.After(cutoff) {
				hasRecentRequests = true
				break
			}
		}
		limiter.mu.RUnlock()

		if !hasRecentRequests {
			delete(store.limiters, clientID)
		}
	}
}

// StartCleanup starts periodic cleanup of old rate limiters
func (store *RateLimitStore) StartCleanup() {
	ticker := time.NewTicker(time.Hour)
	go func() {
		for range ticker.C {
			store.cleanup()
		}
	}()
}

// GetRateLimitStats returns rate limiting statistics
func (store *RateLimitStore) GetRateLimitStats() map[string]interface{} {
	store.mu.RLock()
	defer store.mu.RUnlock()

	stats := map[string]interface{}{
		"total_clients": len(store.limiters),
		"config": map[string]interface{}{
			"requests_per_minute": store.config.RequestsPerMinute,
			"requests_per_hour":   store.config.RequestsPerHour,
			"burst_size":          store.config.BurstSize,
			"enabled":             store.config.EnableRateLimit,
		},
	}

	return stats
}
