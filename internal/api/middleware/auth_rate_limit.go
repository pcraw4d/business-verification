package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// AuthRateLimitConfig holds authentication-specific rate limiting configuration
type AuthRateLimitConfig struct {
	Enabled                  bool
	LoginAttemptsPer         int           // Login attempts per window
	RegisterAttemptsPer      int           // Registration attempts per window
	PasswordResetAttemptsPer int           // Password reset attempts per window
	WindowSize               time.Duration // Time window for rate limiting
	LockoutDuration          time.Duration // Duration to lock out after exceeding limits
}

// AuthRateLimiter implements specialized rate limiting for authentication endpoints
type AuthRateLimiter struct {
	config    *AuthRateLimitConfig
	logger    *observability.Logger
	buckets   map[string]*authBucket
	mu        sync.RWMutex
	cleanupCh chan struct{}
}

// authBucket represents a rate limiting bucket for authentication endpoints
type authBucket struct {
	loginAttempts            int
	registerAttempts         int
	passwordResetAttempts    int
	lastLoginAttempt         time.Time
	lastRegisterAttempt      time.Time
	lastPasswordResetAttempt time.Time
	lockedUntil              *time.Time
	mu                       sync.Mutex
}

// NewAuthRateLimiter creates a new authentication rate limiting middleware
func NewAuthRateLimiter(config *AuthRateLimitConfig, logger *observability.Logger) *AuthRateLimiter {
	arl := &AuthRateLimiter{
		config:    config,
		logger:    logger,
		buckets:   make(map[string]*authBucket),
		cleanupCh: make(chan struct{}),
	}

	// Start cleanup goroutine to remove stale buckets
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

		clientIP := arl.getClientIP(r)
		endpointType := arl.getEndpointType(r.URL.Path)

		// Check if client is locked out
		if arl.isLockedOut(clientIP) {
			arl.logger.WithComponent("auth_rate_limiter").Warn("Authentication endpoint access blocked - client locked out",
				"client_ip", clientIP,
				"path", r.URL.Path,
				"method", r.Method)

			w.Header().Set("X-RateLimit-Limit", "0")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", "300") // 5 minutes
			http.Error(w, "Too many authentication attempts. Please try again later.", http.StatusTooManyRequests)
			return
		}

		// Check rate limit for specific endpoint type
		if !arl.allowRequest(clientIP, endpointType) {
			arl.logger.WithComponent("auth_rate_limiter").Warn("Authentication rate limit exceeded",
				"client_ip", clientIP,
				"path", r.URL.Path,
				"method", r.Method,
				"endpoint_type", endpointType)

			w.Header().Set("X-RateLimit-Limit", arl.getLimitForEndpoint(endpointType))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Too many authentication attempts. Please try again later.", http.StatusTooManyRequests)
			return
		}

		// Add rate limit headers
		remaining := arl.getRemaining(clientIP, endpointType)
		w.Header().Set("X-RateLimit-Limit", arl.getLimitForEndpoint(endpointType))
		w.Header().Set("X-RateLimit-Remaining", string(rune(remaining)))

		next.ServeHTTP(w, r)
	})
}

// isAuthEndpoint checks if the path is an authentication endpoint
func (arl *AuthRateLimiter) isAuthEndpoint(path string) bool {
	authEndpoints := []string{
		"/v1/auth/login",
		"/v1/auth/register",
		"/v1/auth/request-password-reset",
		"/v1/auth/reset-password",
		"/v1/auth/verify-email",
	}

	for _, endpoint := range authEndpoints {
		if strings.HasPrefix(path, endpoint) {
			return true
		}
	}
	return false
}

// getEndpointType determines the type of authentication endpoint
func (arl *AuthRateLimiter) getEndpointType(path string) string {
	switch {
	case strings.HasPrefix(path, "/v1/auth/login"):
		return "login"
	case strings.HasPrefix(path, "/v1/auth/register"):
		return "register"
	case strings.HasPrefix(path, "/v1/auth/request-password-reset"):
		return "password_reset"
	case strings.HasPrefix(path, "/v1/auth/reset-password"):
		return "password_reset"
	case strings.HasPrefix(path, "/v1/auth/verify-email"):
		return "email_verification"
	default:
		return "unknown"
	}
}

// getLimitForEndpoint returns the rate limit for a specific endpoint type
func (arl *AuthRateLimiter) getLimitForEndpoint(endpointType string) string {
	switch endpointType {
	case "login":
		return string(rune(arl.config.LoginAttemptsPer))
	case "register":
		return string(rune(arl.config.RegisterAttemptsPer))
	case "password_reset":
		return string(rune(arl.config.PasswordResetAttemptsPer))
	default:
		return "10"
	}
}

// isLockedOut checks if a client is currently locked out
func (arl *AuthRateLimiter) isLockedOut(clientIP string) bool {
	arl.mu.RLock()
	defer arl.mu.RUnlock()

	bucket, exists := arl.buckets[clientIP]
	if !exists {
		return false
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	if bucket.lockedUntil != nil && time.Now().Before(*bucket.lockedUntil) {
		return true
	}

	// Clear lockout if expired
	if bucket.lockedUntil != nil && time.Now().After(*bucket.lockedUntil) {
		bucket.lockedUntil = nil
	}

	return false
}

// allowRequest checks if a request should be allowed for the given client and endpoint type
func (arl *AuthRateLimiter) allowRequest(clientIP, endpointType string) bool {
	arl.mu.Lock()
	defer arl.mu.Unlock()

	bucket, exists := arl.buckets[clientIP]
	if !exists {
		bucket = &authBucket{}
		arl.buckets[clientIP] = bucket
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()

	// Check if we need to reset counters based on window size
	arl.resetCountersIfNeeded(bucket, now)

	// Check limits based on endpoint type
	switch endpointType {
	case "login":
		if bucket.loginAttempts >= arl.config.LoginAttemptsPer {
			arl.lockoutClient(bucket, now)
			return false
		}
		bucket.loginAttempts++
		bucket.lastLoginAttempt = now

	case "register":
		if bucket.registerAttempts >= arl.config.RegisterAttemptsPer {
			arl.lockoutClient(bucket, now)
			return false
		}
		bucket.registerAttempts++
		bucket.lastRegisterAttempt = now

	case "password_reset":
		if bucket.passwordResetAttempts >= arl.config.PasswordResetAttemptsPer {
			arl.lockoutClient(bucket, now)
			return false
		}
		bucket.passwordResetAttempts++
		bucket.lastPasswordResetAttempt = now

	default:
		// Allow unknown endpoint types
		return true
	}

	return true
}

// resetCountersIfNeeded resets counters if the window has passed
func (arl *AuthRateLimiter) resetCountersIfNeeded(bucket *authBucket, now time.Time) {
	windowAgo := now.Add(-arl.config.WindowSize)

	if bucket.lastLoginAttempt.Before(windowAgo) {
		bucket.loginAttempts = 0
	}
	if bucket.lastRegisterAttempt.Before(windowAgo) {
		bucket.registerAttempts = 0
	}
	if bucket.lastPasswordResetAttempt.Before(windowAgo) {
		bucket.passwordResetAttempts = 0
	}
}

// lockoutClient locks out a client for the configured duration
func (arl *AuthRateLimiter) lockoutClient(bucket *authBucket, now time.Time) {
	lockoutUntil := now.Add(arl.config.LockoutDuration)
	bucket.lockedUntil = &lockoutUntil

	arl.logger.WithComponent("auth_rate_limiter").Warn("Client locked out due to excessive authentication attempts",
		"lockout_until", lockoutUntil.Format(time.RFC3339),
		"duration", arl.config.LockoutDuration)
}

// getRemaining returns the number of remaining requests for a client and endpoint type
func (arl *AuthRateLimiter) getRemaining(clientIP, endpointType string) int {
	arl.mu.RLock()
	defer arl.mu.RUnlock()

	bucket, exists := arl.buckets[clientIP]
	if !exists {
		return arl.getLimitForEndpointType(endpointType)
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Reset counters if needed
	arl.resetCountersIfNeeded(bucket, time.Now())

	switch endpointType {
	case "login":
		return max(0, arl.config.LoginAttemptsPer-bucket.loginAttempts)
	case "register":
		return max(0, arl.config.RegisterAttemptsPer-bucket.registerAttempts)
	case "password_reset":
		return max(0, arl.config.PasswordResetAttemptsPer-bucket.passwordResetAttempts)
	default:
		return 10
	}
}

// getLimitForEndpointType returns the limit for a specific endpoint type
func (arl *AuthRateLimiter) getLimitForEndpointType(endpointType string) int {
	switch endpointType {
	case "login":
		return arl.config.LoginAttemptsPer
	case "register":
		return arl.config.RegisterAttemptsPer
	case "password_reset":
		return arl.config.PasswordResetAttemptsPer
	default:
		return 10
	}
}

// getClientIP extracts the client IP from the request
func (arl *AuthRateLimiter) getClientIP(r *http.Request) string {
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
func (arl *AuthRateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			arl.cleanupStale()
		case <-arl.cleanupCh:
			return
		}
	}
}

// cleanupStale removes buckets that haven't been used recently
func (arl *AuthRateLimiter) cleanupStale() {
	arl.mu.Lock()
	defer arl.mu.Unlock()

	now := time.Now()
	staleThreshold := 30 * time.Minute

	for ip, bucket := range arl.buckets {
		bucket.mu.Lock()

		// Check if bucket is stale (no recent activity and not locked out)
		isStale := true
		if bucket.lockedUntil != nil && now.Before(*bucket.lockedUntil) {
			isStale = false
		} else if now.Sub(bucket.lastLoginAttempt) < staleThreshold ||
			now.Sub(bucket.lastRegisterAttempt) < staleThreshold ||
			now.Sub(bucket.lastPasswordResetAttempt) < staleThreshold {
			isStale = false
		}

		if isStale {
			delete(arl.buckets, ip)
		}

		bucket.mu.Unlock()
	}
}

// Stop stops the rate limiter and cleanup goroutine
func (arl *AuthRateLimiter) Stop() {
	close(arl.cleanupCh)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
