package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool          `json:"enabled" yaml:"enabled"`
	RequestsPerMinute int           `json:"requests_per_minute" yaml:"requests_per_minute"`
	BurstSize         int           `json:"burst_size" yaml:"burst_size"`
	WindowSize        time.Duration `json:"window_size" yaml:"window_size"`
	Strategy          string        `json:"strategy" yaml:"strategy"` // "token_bucket", "sliding_window", "fixed_window"
	Distributed       bool          `json:"distributed" yaml:"distributed"`
	RedisURL          string        `json:"redis_url" yaml:"redis_url"`
	RedisKeyPrefix    string        `json:"redis_key_prefix" yaml:"redis_key_prefix"`
	CleanupInterval   time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
	MaxKeys           int           `json:"max_keys" yaml:"max_keys"`
}

// APIRateLimiter implements rate limiting middleware
type APIRateLimiter struct {
	config    *RateLimitConfig
	logger    *zap.Logger
	store     RateLimitStore
	mu        sync.RWMutex
	cleanupCh chan struct{}
	stats     *RateLimitStats
}

// RateLimitStats holds rate limiting statistics
type RateLimitStats struct {
	mu              sync.RWMutex
	TotalRequests   int64         `json:"total_requests"`
	BlockedRequests int64         `json:"blocked_requests"`
	ActiveKeys      int           `json:"active_keys"`
	LastReset       time.Time     `json:"last_reset"`
	WindowSize      time.Duration `json:"window_size"`
}

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed   bool      `json:"allowed"`
	Remaining int       `json:"remaining"`
	ResetTime time.Time `json:"reset_time"`
	Limit     int       `json:"limit"`
	Key       string    `json:"key"`
}

// NewAPIRateLimiter creates a new rate limiter with the given configuration
func NewAPIRateLimiter(config *RateLimitConfig, logger *zap.Logger) *APIRateLimiter {
	if config == nil {
		config = &RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 100,
			BurstSize:         20,
			WindowSize:        time.Minute,
			Strategy:          "token_bucket",
			Distributed:       false,
			CleanupInterval:   5 * time.Minute,
			MaxKeys:           10000,
		}
	}

	// Set default values
	if config.WindowSize == 0 {
		config.WindowSize = time.Minute
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 5 * time.Minute
	}
	if config.MaxKeys == 0 {
		config.MaxKeys = 10000
	}
	if config.RedisKeyPrefix == "" {
		config.RedisKeyPrefix = "rate_limit"
	}

	var store RateLimitStore
	if config.Distributed && config.RedisURL != "" {
		store = NewRedisRateLimitStore(config.RedisURL, config.RedisKeyPrefix, logger)
	} else {
		store = NewMemoryRateLimitStore(config.Strategy, config.MaxKeys, logger)
	}

	rl := &APIRateLimiter{
		config:    config,
		logger:    logger,
		store:     store,
		cleanupCh: make(chan struct{}),
		stats: &RateLimitStats{
			LastReset:  time.Now(),
			WindowSize: config.WindowSize,
		},
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Middleware returns the rate limiting middleware
func (rl *APIRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Generate rate limit key
		key := rl.generateKey(r)

		// Check rate limit
		allowed, remaining, err := rl.store.Allow(key, rl.config.RequestsPerMinute, rl.config.WindowSize)
		if err != nil {
			rl.logger.Error("Rate limit check failed", zap.Error(err), zap.String("key", key))
			// On error, allow the request but log the issue
			next.ServeHTTP(w, r)
			return
		}

		// Update statistics
		rl.updateStats(allowed)

		// Set rate limit headers
		rl.setRateLimitHeaders(w, remaining, rl.config.RequestsPerMinute)

		if !allowed {
			rl.logger.Warn("Rate limit exceeded",
				zap.String("key", key),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.String("client_ip", rl.getClientIP(r)))

			// Set retry-after header
			w.Header().Set("Retry-After", strconv.Itoa(int(rl.config.WindowSize.Seconds())))
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// generateKey generates a unique key for rate limiting based on the request
func (rl *APIRateLimiter) generateKey(r *http.Request) string {
	// Use client IP as primary identifier
	clientIP := rl.getClientIP(r)

	// For API endpoints, also consider the user ID if available
	userID := rl.getUserID(r)

	// For specific endpoints, consider the endpoint path
	path := r.URL.Path

	// Create a composite key
	keyParts := []string{clientIP}
	if userID != "" {
		keyParts = append(keyParts, userID)
	}
	if path != "" {
		// Hash the path to keep keys shorter
		pathHash := sha256.Sum256([]byte(path))
		keyParts = append(keyParts, hex.EncodeToString(pathHash[:8]))
	}

	return strings.Join(keyParts, ":")
}

// getClientIP extracts the client IP from the request
func (rl *APIRateLimiter) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (most common for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if commaIdx := strings.Index(xff, ","); commaIdx != -1 {
			return strings.TrimSpace(xff[:commaIdx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Check X-Client-IP header
	if xci := r.Header.Get("X-Client-IP"); xci != "" {
		return strings.TrimSpace(xci)
	}

	// Check CF-Connecting-IP header (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		return strings.TrimSpace(cfip)
	}

	// Fall back to RemoteAddr
	if r.RemoteAddr != "" {
		// Remove port if present
		if colonIdx := strings.LastIndex(r.RemoteAddr, ":"); colonIdx != -1 {
			return r.RemoteAddr[:colonIdx]
		}
		return r.RemoteAddr
	}

	// Fallback to a default value
	return "unknown"
}

// getUserID extracts the user ID from the request context or headers
func (rl *APIRateLimiter) getUserID(r *http.Request) string {
	// Check for user ID in context (set by auth middleware)
	if userID, ok := r.Context().Value("user_id").(string); ok && userID != "" {
		return userID
	}

	// Check for user ID in headers
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}

	// Check for API key in headers
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		// Hash the API key to keep it consistent
		hash := sha256.Sum256([]byte(apiKey))
		return hex.EncodeToString(hash[:16])
	}

	return ""
}

// setRateLimitHeaders sets the rate limit headers on the response
func (rl *APIRateLimiter) setRateLimitHeaders(w http.ResponseWriter, remaining, limit int) {
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rl.config.WindowSize).Unix(), 10))
}

// updateStats updates the rate limiting statistics
func (rl *APIRateLimiter) updateStats(allowed bool) {
	rl.stats.mu.Lock()
	defer rl.stats.mu.Unlock()

	rl.stats.TotalRequests++
	if !allowed {
		rl.stats.BlockedRequests++
	}
}

// GetStats returns the current rate limiting statistics
func (rl *APIRateLimiter) GetStats() *RateLimitStats {
	rl.stats.mu.RLock()
	defer rl.stats.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := &RateLimitStats{
		TotalRequests:   rl.stats.TotalRequests,
		BlockedRequests: rl.stats.BlockedRequests,
		ActiveKeys:      rl.stats.ActiveKeys,
		LastReset:       rl.stats.LastReset,
		WindowSize:      rl.stats.WindowSize,
	}

	return stats
}

// ResetStats resets the rate limiting statistics
func (rl *APIRateLimiter) ResetStats() {
	rl.stats.mu.Lock()
	defer rl.stats.mu.Unlock()

	rl.stats.TotalRequests = 0
	rl.stats.BlockedRequests = 0
	rl.stats.LastReset = time.Now()
}

// cleanup periodically cleans up expired rate limit entries
func (rl *APIRateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := rl.store.Cleanup(); err != nil {
				rl.logger.Error("Rate limit cleanup failed", zap.Error(err))
			}
		case <-rl.cleanupCh:
			return
		}
	}
}

// Stop stops the rate limiter and cleanup goroutine
func (rl *APIRateLimiter) Stop() {
	close(rl.cleanupCh)
}

// CheckRateLimit checks if a request would be allowed without actually consuming a token
func (rl *APIRateLimiter) CheckRateLimit(key string) (*RateLimitResult, error) {
	remaining, err := rl.store.GetRemaining(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get remaining requests: %w", err)
	}

	allowed := remaining > 0
	resetTime := time.Now().Add(rl.config.WindowSize)

	return &RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: resetTime,
		Limit:     rl.config.RequestsPerMinute,
		Key:       key,
	}, nil
}

// ResetKey resets the rate limit for a specific key
func (rl *APIRateLimiter) ResetKey(key string) error {
	return rl.store.Reset(key)
}

// GetConfig returns the current rate limiting configuration
func (rl *APIRateLimiter) GetConfig() *RateLimitConfig {
	return rl.config
}

// UpdateConfig updates the rate limiting configuration
func (rl *APIRateLimiter) UpdateConfig(config *RateLimitConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.config = config
	rl.logger.Info("Rate limiting configuration updated",
		zap.Bool("enabled", config.Enabled),
		zap.Int("requests_per_minute", config.RequestsPerMinute),
		zap.Int("burst_size", config.BurstSize),
		zap.String("strategy", config.Strategy))
}
