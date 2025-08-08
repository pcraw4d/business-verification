package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
	Enabled           bool
}

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	config    *RateLimitConfig
	logger    *observability.Logger
	buckets   map[string]*bucket
	mu        sync.RWMutex
	cleanupCh chan struct{}
}

// bucket represents a token bucket for a specific client
type bucket struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiting middleware
func NewRateLimiter(config *RateLimitConfig, logger *observability.Logger) *RateLimiter {
	rl := &RateLimiter{
		config:    config,
		logger:    logger,
		buckets:   make(map[string]*bucket),
		cleanupCh: make(chan struct{}),
	}

	// Start cleanup goroutine to remove stale buckets
	go rl.cleanup()

	return rl
}

// Middleware returns the rate limiting middleware
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := rl.getClientIP(r)

		if !rl.allow(clientIP) {
			rl.logger.WithComponent("rate_limiter").Warn("Rate limit exceeded",
				"client_ip", clientIP,
				"path", r.URL.Path,
				"method", r.Method)

			w.Header().Set("X-RateLimit-Limit", string(rune(rl.config.RequestsPerMinute)))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Add rate limit headers
		remaining := rl.getRemaining(clientIP)
		w.Header().Set("X-RateLimit-Limit", string(rune(rl.config.RequestsPerMinute)))
		w.Header().Set("X-RateLimit-Remaining", string(rune(remaining)))

		next.ServeHTTP(w, r)
	})
}

// allow checks if a request should be allowed for the given client
func (rl *RateLimiter) allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.buckets[clientIP]
	if !exists {
		b = &bucket{
			tokens:     rl.config.BurstSize,
			maxTokens:  rl.config.BurstSize,
			refillRate: time.Minute / time.Duration(rl.config.RequestsPerMinute),
			lastRefill: time.Now(),
		}
		rl.buckets[clientIP] = b
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill)

	// Refill tokens based on elapsed time
	tokensToAdd := int(elapsed / b.refillRate)
	if tokensToAdd > 0 {
		b.tokens = min(b.maxTokens, b.tokens+tokensToAdd)
		b.lastRefill = now
	}

	// Check if we have tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// getRemaining returns the number of remaining requests for a client
func (rl *RateLimiter) getRemaining(clientIP string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	b, exists := rl.buckets[clientIP]
	if !exists {
		return rl.config.BurstSize
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	return b.tokens
}

// getClientIP extracts the client IP from the request
func (rl *RateLimiter) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// cleanup removes stale buckets to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanupStale()
		case <-rl.cleanupCh:
			return
		}
	}
}

// cleanupStale removes buckets that haven't been used recently
func (rl *RateLimiter) cleanupStale() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	staleThreshold := 10 * time.Minute

	for ip, bucket := range rl.buckets {
		bucket.mu.Lock()
		if now.Sub(bucket.lastRefill) > staleThreshold {
			delete(rl.buckets, ip)
		}
		bucket.mu.Unlock()
	}
}

// Stop stops the rate limiter and cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.cleanupCh)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
