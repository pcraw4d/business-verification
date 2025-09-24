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

// AuthRateLimitConfig holds authentication rate limiting configuration
type AuthRateLimitConfig struct {
	Enabled                  bool          `json:"enabled" yaml:"enabled"`
	LoginAttemptsPer         int           `json:"login_attempts_per" yaml:"login_attempts_per"`
	RegisterAttemptsPer      int           `json:"register_attempts_per" yaml:"register_attempts_per"`
	PasswordResetAttemptsPer int           `json:"password_reset_attempts_per" yaml:"password_reset_attempts_per"`
	WindowSize               time.Duration `json:"window_size" yaml:"window_size"`
	LockoutDuration          time.Duration `json:"lockout_duration" yaml:"lockout_duration"`
	MaxLockouts              int           `json:"max_lockouts" yaml:"max_lockouts"`
	PermanentLockoutDuration time.Duration `json:"permanent_lockout_duration" yaml:"permanent_lockout_duration"`
	Distributed              bool          `json:"distributed" yaml:"distributed"`
	RedisURL                 string        `json:"redis_url" yaml:"redis_url"`
	RedisKeyPrefix           string        `json:"redis_key_prefix" yaml:"redis_key_prefix"`
}

// AuthRateLimiter implements authentication-specific rate limiting
type AuthRateLimiter struct {
	config    *AuthRateLimitConfig
	logger    *zap.Logger
	store     AuthRateLimitStore
	mu        sync.RWMutex
	cleanupCh chan struct{}
	stats     *AuthRateLimitStats
}

// AuthRateLimitStore defines the interface for auth rate limit storage
type AuthRateLimitStore interface {
	CheckAuthLimit(key string, limit int, window time.Duration) (bool, int, time.Time, error)
	RecordFailedAttempt(key string, attemptType string) error
	IsLocked(key string) (bool, time.Time, error)
	Reset(key string) error
	Cleanup() error
}

// AuthRateLimitStats holds authentication rate limiting statistics
type AuthRateLimitStats struct {
	mu             sync.RWMutex
	TotalAttempts  int64         `json:"total_attempts"`
	FailedAttempts int64         `json:"failed_attempts"`
	Lockouts       int64         `json:"lockouts"`
	ActiveLockouts int           `json:"active_lockouts"`
	LastReset      time.Time     `json:"last_reset"`
	WindowSize     time.Duration `json:"window_size"`
}

// AuthAttemptType represents the type of authentication attempt
type AuthAttemptType string

const (
	LoginAttempt         AuthAttemptType = "login"
	RegisterAttempt      AuthAttemptType = "register"
	PasswordResetAttempt AuthAttemptType = "password_reset"
)

// NewAuthRateLimiter creates a new authentication rate limiter
func NewAuthRateLimiter(config *AuthRateLimitConfig, logger *zap.Logger) *AuthRateLimiter {
	if config == nil {
		config = &AuthRateLimitConfig{
			Enabled:                  true,
			LoginAttemptsPer:         5,
			RegisterAttemptsPer:      3,
			PasswordResetAttemptsPer: 3,
			WindowSize:               60 * time.Second,
			LockoutDuration:          15 * time.Minute,
			MaxLockouts:              3,
			PermanentLockoutDuration: 24 * time.Hour,
			Distributed:              false,
			RedisKeyPrefix:           "auth_rate_limit",
		}
	}

	// Set default values
	if config.WindowSize == 0 {
		config.WindowSize = 60 * time.Second
	}
	if config.LockoutDuration == 0 {
		config.LockoutDuration = 15 * time.Minute
	}
	if config.MaxLockouts == 0 {
		config.MaxLockouts = 3
	}
	if config.PermanentLockoutDuration == 0 {
		config.PermanentLockoutDuration = 24 * time.Hour
	}
	if config.RedisKeyPrefix == "" {
		config.RedisKeyPrefix = "auth_rate_limit"
	}

	var store AuthRateLimitStore
	if config.Distributed && config.RedisURL != "" {
		store = NewRedisAuthRateLimitStore(config.RedisURL, config.RedisKeyPrefix, logger)
	} else {
		store = NewMemoryAuthRateLimitStore(logger)
	}

	arl := &AuthRateLimiter{
		config:    config,
		logger:    logger,
		store:     store,
		cleanupCh: make(chan struct{}),
		stats: &AuthRateLimitStats{
			LastReset:  time.Now(),
			WindowSize: config.WindowSize,
		},
	}

	// Start cleanup goroutine
	go arl.cleanup()

	return arl
}

// Middleware returns the authentication rate limiting middleware
func (arl *AuthRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !arl.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Only apply to authentication endpoints
		if !arl.isAuthEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Generate rate limit key
		key := arl.generateKey(r)

		// Check if the client is locked out
		locked, lockoutUntil, err := arl.store.IsLocked(key)
		if err != nil {
			arl.logger.Error("Failed to check lockout status", zap.Error(err), zap.String("key", key))
			// On error, allow the request but log the issue
			next.ServeHTTP(w, r)
			return
		}

		if locked {
			arl.logger.Warn("Authentication attempt blocked due to lockout",
				zap.String("key", key),
				zap.String("path", r.URL.Path),
				zap.String("client_ip", arl.getClientIP(r)),
				zap.Time("lockout_until", lockoutUntil))

			// Set retry-after header
			retryAfter := int(time.Until(lockoutUntil).Seconds())
			if retryAfter > 0 {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			}
			http.Error(w, "Account temporarily locked due to too many failed attempts", http.StatusTooManyRequests)
			return
		}

		// Get the appropriate limit for this endpoint
		limit := arl.getLimitForEndpoint(r.URL.Path)

		// Check rate limit
		allowed, remaining, resetTime, err := arl.store.CheckAuthLimit(key, limit, arl.config.WindowSize)
		if err != nil {
			arl.logger.Error("Auth rate limit check failed", zap.Error(err), zap.String("key", key))
			// On error, allow the request but log the issue
			next.ServeHTTP(w, r)
			return
		}

		// Update statistics
		arl.updateStats(allowed)

		// Set rate limit headers
		arl.setAuthRateLimitHeaders(w, remaining, limit, resetTime)

		if !allowed {
			arl.logger.Warn("Authentication rate limit exceeded",
				zap.String("key", key),
				zap.String("path", r.URL.Path),
				zap.String("client_ip", arl.getClientIP(r)))

			// Set retry-after header
			retryAfter := int(time.Until(resetTime).Seconds())
			if retryAfter > 0 {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			}
			http.Error(w, "Too many authentication attempts", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RecordFailedAttempt records a failed authentication attempt
func (arl *AuthRateLimiter) RecordFailedAttempt(r *http.Request, attemptType AuthAttemptType) error {
	if !arl.config.Enabled {
		return nil
	}

	key := arl.generateKey(r)
	err := arl.store.RecordFailedAttempt(key, string(attemptType))
	if err != nil {
		arl.logger.Error("Failed to record authentication attempt", zap.Error(err), zap.String("key", key))
		return err
	}

	arl.logger.Info("Recorded failed authentication attempt",
		zap.String("key", key),
		zap.String("attempt_type", string(attemptType)),
		zap.String("client_ip", arl.getClientIP(r)))

	return nil
}

// isAuthEndpoint checks if the given path is an authentication endpoint
func (arl *AuthRateLimiter) isAuthEndpoint(path string) bool {
	authEndpoints := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/password/reset",
		"/api/v1/auth/password/forgot",
		"/api/v1/auth/verify",
		"/api/v1/auth/refresh",
	}

	for _, endpoint := range authEndpoints {
		if strings.HasPrefix(path, endpoint) {
			return true
		}
	}
	return false
}

// generateKey generates a unique key for authentication rate limiting
func (arl *AuthRateLimiter) generateKey(r *http.Request) string {
	clientIP := arl.getClientIP(r)

	// For authentication, we might want to include additional identifiers
	userAgent := r.Header.Get("User-Agent")
	if userAgent != "" {
		// Hash the user agent to keep keys shorter
		uaHash := sha256.Sum256([]byte(userAgent))
		return fmt.Sprintf("%s:%s", clientIP, hex.EncodeToString(uaHash[:8]))
	}

	return clientIP
}

// getClientIP extracts the client IP from the request
func (arl *AuthRateLimiter) getClientIP(r *http.Request) string {
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

// getLimitForEndpoint returns the appropriate rate limit for the given endpoint
func (arl *AuthRateLimiter) getLimitForEndpoint(path string) int {
	switch {
	case strings.Contains(path, "/login"):
		return arl.config.LoginAttemptsPer
	case strings.Contains(path, "/register"):
		return arl.config.RegisterAttemptsPer
	case strings.Contains(path, "/password/reset") || strings.Contains(path, "/password/forgot"):
		return arl.config.PasswordResetAttemptsPer
	default:
		return arl.config.LoginAttemptsPer // Default to login limit
	}
}

// setAuthRateLimitHeaders sets the authentication rate limit headers on the response
func (arl *AuthRateLimiter) setAuthRateLimitHeaders(w http.ResponseWriter, remaining, limit int, resetTime time.Time) {
	w.Header().Set("X-AuthRateLimit-Limit", strconv.Itoa(limit))
	w.Header().Set("X-AuthRateLimit-Remaining", strconv.Itoa(remaining))
	w.Header().Set("X-AuthRateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
}

// updateStats updates the authentication rate limiting statistics
func (arl *AuthRateLimiter) updateStats(allowed bool) {
	arl.stats.mu.Lock()
	defer arl.stats.mu.Unlock()

	arl.stats.TotalAttempts++
	if !allowed {
		arl.stats.FailedAttempts++
	}
}

// GetStats returns the current authentication rate limiting statistics
func (arl *AuthRateLimiter) GetStats() *AuthRateLimitStats {
	arl.stats.mu.RLock()
	defer arl.stats.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := &AuthRateLimitStats{
		TotalAttempts:  arl.stats.TotalAttempts,
		FailedAttempts: arl.stats.FailedAttempts,
		Lockouts:       arl.stats.Lockouts,
		ActiveLockouts: arl.stats.ActiveLockouts,
		LastReset:      arl.stats.LastReset,
		WindowSize:     arl.stats.WindowSize,
	}

	return stats
}

// ResetStats resets the authentication rate limiting statistics
func (arl *AuthRateLimiter) ResetStats() {
	arl.stats.mu.Lock()
	defer arl.stats.mu.Unlock()

	arl.stats.TotalAttempts = 0
	arl.stats.FailedAttempts = 0
	arl.stats.Lockouts = 0
	arl.stats.LastReset = time.Now()
}

// cleanup periodically cleans up expired rate limit entries
func (arl *AuthRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := arl.store.Cleanup(); err != nil {
				arl.logger.Error("Auth rate limit cleanup failed", zap.Error(err))
			}
		case <-arl.cleanupCh:
			return
		}
	}
}

// Stop stops the auth rate limiter and cleanup goroutine
func (arl *AuthRateLimiter) Stop() {
	close(arl.cleanupCh)
}

// ResetKey resets the rate limit for a specific key
func (arl *AuthRateLimiter) ResetKey(key string) error {
	return arl.store.Reset(key)
}

// GetConfig returns the current authentication rate limiting configuration
func (arl *AuthRateLimiter) GetConfig() *AuthRateLimitConfig {
	return arl.config
}

// UpdateConfig updates the authentication rate limiting configuration
func (arl *AuthRateLimiter) UpdateConfig(config *AuthRateLimitConfig) {
	arl.mu.Lock()
	defer arl.mu.Unlock()

	arl.config = config
	arl.logger.Info("Authentication rate limiting configuration updated",
		zap.Bool("enabled", config.Enabled),
		zap.Int("login_attempts_per", config.LoginAttemptsPer),
		zap.Int("register_attempts_per", config.RegisterAttemptsPer),
		zap.Int("password_reset_attempts_per", config.PasswordResetAttemptsPer))
}
