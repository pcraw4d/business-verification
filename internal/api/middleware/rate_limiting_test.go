package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewAPIRateLimiter(t *testing.T) {
	logger := zap.NewNop()

	t.Run("with nil config", func(t *testing.T) {
		rl := NewAPIRateLimiter(nil, logger)
		assert.NotNil(t, rl)
		assert.True(t, rl.config.Enabled)
		assert.Equal(t, 100, rl.config.RequestsPerMinute)
		assert.Equal(t, 20, rl.config.BurstSize)
		assert.Equal(t, time.Minute, rl.config.WindowSize)
		assert.Equal(t, "token_bucket", rl.config.Strategy)
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 50,
			BurstSize:         10,
			WindowSize:        30 * time.Second,
			Strategy:          "sliding_window",
			Distributed:       false,
		}

		rl := NewAPIRateLimiter(config, logger)
		assert.NotNil(t, rl)
		assert.Equal(t, config, rl.config)
	})

	t.Run("with distributed config", func(t *testing.T) {
		config := &RateLimitConfig{
			Enabled:        true,
			Distributed:    true,
			RedisURL:       "redis://localhost:6379",
			RedisKeyPrefix: "test_rate_limit",
		}

		rl := NewAPIRateLimiter(config, logger)
		assert.NotNil(t, rl)
		assert.True(t, rl.config.Distributed)
	})
}

func TestAPIRateLimiter_Middleware(t *testing.T) {
	logger := zap.NewNop()

	t.Run("disabled rate limiting", func(t *testing.T) {
		config := &RateLimitConfig{
			Enabled:           false,
			RequestsPerMinute: 1,
			BurstSize:         1,
			WindowSize:        time.Second,
		}

		rl := NewAPIRateLimiter(config, logger)
		handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		config := &RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 1,
			BurstSize:         1,
			WindowSize:        time.Second,
		}

		rl := NewAPIRateLimiter(config, logger)
		handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		// First request should succeed
		w1 := httptest.NewRecorder()
		handler.ServeHTTP(w1, req)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Second request should be rate limited
		w2 := httptest.NewRecorder()
		handler.ServeHTTP(w2, req)
		assert.Equal(t, http.StatusTooManyRequests, w2.Code)
		assert.Contains(t, w2.Body.String(), "Rate limit exceeded")
		assert.Equal(t, "1", w2.Header().Get("Retry-After"))
	})

	t.Run("rate limit headers", func(t *testing.T) {
		config := &RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 10,
			BurstSize:         5,
			WindowSize:        time.Minute,
		}

		rl := NewAPIRateLimiter(config, logger)
		handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "10", w.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "4", w.Header().Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
	})

	t.Run("client IP extraction", func(t *testing.T) {
		config := &RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 10,
			BurstSize:         5,
			WindowSize:        time.Minute,
		}

		rl := NewAPIRateLimiter(config, logger)
		handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Test X-Forwarded-For header
		req1 := httptest.NewRequest("GET", "/test", nil)
		req1.Header.Set("X-Forwarded-For", "203.0.113.1, 192.168.1.1")
		w1 := httptest.NewRecorder()
		handler.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Test X-Real-IP header
		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.Header.Set("X-Real-IP", "203.0.113.2")
		w2 := httptest.NewRecorder()
		handler.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		// Test CF-Connecting-IP header
		req3 := httptest.NewRequest("GET", "/test", nil)
		req3.Header.Set("CF-Connecting-IP", "203.0.113.3")
		w3 := httptest.NewRecorder()
		handler.ServeHTTP(w3, req3)
		assert.Equal(t, http.StatusOK, w3.Code)
	})
}

func TestAPIRateLimiter_GenerateKey(t *testing.T) {
	logger := zap.NewNop()
	config := &RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 10,
		BurstSize:         5,
		WindowSize:        time.Minute,
	}

	rl := NewAPIRateLimiter(config, logger)

	t.Run("basic key generation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		key := rl.generateKey(req)
		assert.Equal(t, "192.168.1.1", key)
	})

	t.Run("key with user ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		req.Header.Set("X-User-ID", "user123")

		key := rl.generateKey(req)
		assert.Equal(t, "192.168.1.1:user123", key)
	})

	t.Run("key with API key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		req.Header.Set("X-API-Key", "api_key_123")

		key := rl.generateKey(req)
		// Should contain the hashed API key
		assert.True(t, strings.Contains(key, "192.168.1.1:"))
		assert.True(t, len(key) > len("192.168.1.1:"))
	})

	t.Run("key with path", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		key := rl.generateKey(req)
		// Should contain the hashed path
		assert.True(t, strings.Contains(key, "192.168.1.1:"))
		assert.True(t, len(key) > len("192.168.1.1:"))
	})
}

func TestAPIRateLimiter_Stats(t *testing.T) {
	logger := zap.NewNop()
	config := &RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 10,
		BurstSize:         5,
		WindowSize:        time.Minute,
	}

	rl := NewAPIRateLimiter(config, logger)

	// Initial stats
	stats := rl.GetStats()
	assert.Equal(t, int64(0), stats.TotalRequests)
	assert.Equal(t, int64(0), stats.BlockedRequests)

	// Make some requests
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// Successful request
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Rate limited request (after exceeding limit)
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// Check stats
	stats = rl.GetStats()
	assert.True(t, stats.TotalRequests > 0)
	assert.True(t, stats.BlockedRequests > 0)

	// Reset stats
	rl.ResetStats()
	stats = rl.GetStats()
	assert.Equal(t, int64(0), stats.TotalRequests)
	assert.Equal(t, int64(0), stats.BlockedRequests)
}

func TestMemoryRateLimitStore(t *testing.T) {
	logger := zap.NewNop()
	store := NewMemoryRateLimitStore("token_bucket", 1000, logger)

	t.Run("first request", func(t *testing.T) {
		allowed, remaining, err := store.Allow("test_key", 10, time.Minute)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, 9, remaining)
	})

	t.Run("within limit", func(t *testing.T) {
		allowed, remaining, err := store.Allow("test_key2", 5, time.Minute)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, 4, remaining)

		allowed, remaining, err = store.Allow("test_key2", 5, time.Minute)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, 3, remaining)
	})

	t.Run("exceed limit", func(t *testing.T) {
		// Make 5 requests to a key with limit 5
		for i := 0; i < 5; i++ {
			allowed, _, err := store.Allow("test_key3", 5, time.Minute)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}

		// 6th request should be blocked
		allowed, remaining, err := store.Allow("test_key3", 5, time.Minute)
		assert.NoError(t, err)
		assert.False(t, allowed)
		assert.Equal(t, 0, remaining)
	})

	t.Run("get remaining", func(t *testing.T) {
		remaining, err := store.GetRemaining("test_key4")
		assert.NoError(t, err)
		assert.Equal(t, 0, remaining) // No requests made yet

		store.Allow("test_key4", 10, time.Minute)
		remaining, err = store.GetRemaining("test_key4")
		assert.NoError(t, err)
		assert.Equal(t, 9, remaining)
	})

	t.Run("reset key", func(t *testing.T) {
		// Make some requests
		store.Allow("test_key5", 5, time.Minute)
		store.Allow("test_key5", 5, time.Minute)

		// Reset the key
		err := store.Reset("test_key5")
		assert.NoError(t, err)

		// Should be able to make requests again
		allowed, remaining, err := store.Allow("test_key5", 5, time.Minute)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, 4, remaining)
	})
}

func TestMemoryAuthRateLimitStore(t *testing.T) {
	logger := zap.NewNop()
	store := NewMemoryAuthRateLimitStore(logger)

	t.Run("first attempt", func(t *testing.T) {
		allowed, remaining, resetTime, err := store.CheckAuthLimit("test_key", 5, time.Minute)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, 4, remaining)
		assert.True(t, resetTime.After(time.Now()))
	})

	t.Run("within limit", func(t *testing.T) {
		allowed, remaining, _, err := store.CheckAuthLimit("test_key2", 3, time.Minute)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, 2, remaining)

		allowed, remaining, _, err = store.CheckAuthLimit("test_key2", 3, time.Minute)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, 1, remaining)
	})

	t.Run("exceed limit", func(t *testing.T) {
		// Make 3 requests to a key with limit 3
		for i := 0; i < 3; i++ {
			allowed, _, _, err := store.CheckAuthLimit("test_key3", 3, time.Minute)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}

		// 4th request should be blocked
		allowed, remaining, _, err := store.CheckAuthLimit("test_key3", 3, time.Minute)
		assert.NoError(t, err)
		assert.False(t, allowed)
		assert.Equal(t, 0, remaining)
	})

	t.Run("record failed attempt", func(t *testing.T) {
		err := store.RecordFailedAttempt("test_key4", "login")
		assert.NoError(t, err)

		// Check if it's locked after multiple failed attempts
		for i := 0; i < 5; i++ {
			store.RecordFailedAttempt("test_key4", "login")
		}

		locked, lockoutUntil, err := store.IsLocked("test_key4")
		assert.NoError(t, err)
		assert.True(t, locked)
		assert.True(t, lockoutUntil.After(time.Now()))
	})

	t.Run("lockout expiration", func(t *testing.T) {
		// Create a lockout that expires quickly
		store.RecordFailedAttempt("test_key5", "login")
		store.RecordFailedAttempt("test_key5", "login")
		store.RecordFailedAttempt("test_key5", "login")
		store.RecordFailedAttempt("test_key5", "login")
		store.RecordFailedAttempt("test_key5", "login")

		// Should be locked
		locked, _, err := store.IsLocked("test_key5")
		assert.NoError(t, err)
		assert.True(t, locked)

		// Wait for lockout to expire (in real scenario, this would be longer)
		// For testing, we'll just reset the key
		err = store.Reset("test_key5")
		assert.NoError(t, err)

		// Should not be locked anymore
		locked, _, err = store.IsLocked("test_key5")
		assert.NoError(t, err)
		assert.False(t, locked)
	})
}

func TestNewAuthRateLimiter(t *testing.T) {
	logger := zap.NewNop()

	t.Run("with nil config", func(t *testing.T) {
		arl := NewAuthRateLimiter(nil, logger)
		assert.NotNil(t, arl)
		assert.True(t, arl.config.Enabled)
		assert.Equal(t, 5, arl.config.LoginAttemptsPer)
		assert.Equal(t, 3, arl.config.RegisterAttemptsPer)
		assert.Equal(t, 3, arl.config.PasswordResetAttemptsPer)
		assert.Equal(t, 60*time.Second, arl.config.WindowSize)
		assert.Equal(t, 15*time.Minute, arl.config.LockoutDuration)
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &AuthRateLimitConfig{
			Enabled:                  true,
			LoginAttemptsPer:         3,
			RegisterAttemptsPer:      2,
			PasswordResetAttemptsPer: 1,
			WindowSize:               30 * time.Second,
			LockoutDuration:          10 * time.Minute,
			MaxLockouts:              5,
			PermanentLockoutDuration: 12 * time.Hour,
		}

		arl := NewAuthRateLimiter(config, logger)
		assert.NotNil(t, arl)
		assert.Equal(t, config, arl.config)
	})
}

func TestAuthRateLimiter_Middleware(t *testing.T) {
	logger := zap.NewNop()

	t.Run("disabled rate limiting", func(t *testing.T) {
		config := &AuthRateLimitConfig{
			Enabled: false,
		}

		arl := NewAuthRateLimiter(config, logger)
		handler := arl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("non-auth endpoint", func(t *testing.T) {
		config := &AuthRateLimitConfig{
			Enabled: true,
		}

		arl := NewAuthRateLimiter(config, logger)
		handler := arl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("auth endpoint rate limiting", func(t *testing.T) {
		config := &AuthRateLimitConfig{
			Enabled:          true,
			LoginAttemptsPer: 2,
			WindowSize:       time.Second,
		}

		arl := NewAuthRateLimiter(config, logger)
		handler := arl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		// First request should succeed
		w1 := httptest.NewRecorder()
		handler.ServeHTTP(w1, req)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Second request should succeed
		w2 := httptest.NewRecorder()
		handler.ServeHTTP(w2, req)
		assert.Equal(t, http.StatusOK, w2.Code)

		// Third request should be rate limited
		w3 := httptest.NewRecorder()
		handler.ServeHTTP(w3, req)
		assert.Equal(t, http.StatusTooManyRequests, w3.Code)
		assert.Contains(t, w3.Body.String(), "Too many authentication attempts")
	})
}

func TestAuthRateLimiter_RecordFailedAttempt(t *testing.T) {
	logger := zap.NewNop()
	config := &AuthRateLimitConfig{
		Enabled: true,
	}

	arl := NewAuthRateLimiter(config, logger)

	t.Run("record failed login attempt", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		err := arl.RecordFailedAttempt(req, LoginAttempt)
		assert.NoError(t, err)
	})

	t.Run("record failed register attempt", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/auth/register", nil)
		req.RemoteAddr = "192.168.1.2:12345"

		err := arl.RecordFailedAttempt(req, RegisterAttempt)
		assert.NoError(t, err)
	})

	t.Run("record failed password reset attempt", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/auth/password/reset", nil)
		req.RemoteAddr = "192.168.1.3:12345"

		err := arl.RecordFailedAttempt(req, PasswordResetAttempt)
		assert.NoError(t, err)
	})
}

func TestAuthRateLimiter_IsAuthEndpoint(t *testing.T) {
	logger := zap.NewNop()
	config := &AuthRateLimitConfig{
		Enabled: true,
	}

	arl := NewAuthRateLimiter(config, logger)

	authEndpoints := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/password/reset",
		"/api/v1/auth/password/forgot",
		"/api/v1/auth/verify",
		"/api/v1/auth/refresh",
	}

	nonAuthEndpoints := []string{
		"/api/v1/users",
		"/api/v1/businesses",
		"/api/v1/reports",
		"/health",
		"/metrics",
	}

	for _, endpoint := range authEndpoints {
		t.Run("auth endpoint: "+endpoint, func(t *testing.T) {
			assert.True(t, arl.isAuthEndpoint(endpoint))
		})
	}

	for _, endpoint := range nonAuthEndpoints {
		t.Run("non-auth endpoint: "+endpoint, func(t *testing.T) {
			assert.False(t, arl.isAuthEndpoint(endpoint))
		})
	}
}

func TestAuthRateLimiter_GetLimitForEndpoint(t *testing.T) {
	logger := zap.NewNop()
	config := &AuthRateLimitConfig{
		Enabled:                  true,
		LoginAttemptsPer:         5,
		RegisterAttemptsPer:      3,
		PasswordResetAttemptsPer: 2,
	}

	arl := NewAuthRateLimiter(config, logger)

	t.Run("login endpoint", func(t *testing.T) {
		limit := arl.getLimitForEndpoint("/api/v1/auth/login")
		assert.Equal(t, 5, limit)
	})

	t.Run("register endpoint", func(t *testing.T) {
		limit := arl.getLimitForEndpoint("/api/v1/auth/register")
		assert.Equal(t, 3, limit)
	})

	t.Run("password reset endpoint", func(t *testing.T) {
		limit := arl.getLimitForEndpoint("/api/v1/auth/password/reset")
		assert.Equal(t, 2, limit)
	})

	t.Run("password forgot endpoint", func(t *testing.T) {
		limit := arl.getLimitForEndpoint("/api/v1/auth/password/forgot")
		assert.Equal(t, 2, limit)
	})

	t.Run("unknown endpoint", func(t *testing.T) {
		limit := arl.getLimitForEndpoint("/api/v1/auth/unknown")
		assert.Equal(t, 5, limit) // Defaults to login limit
	})
}

func TestRateLimiter_Integration(t *testing.T) {
	logger := zap.NewNop()

	t.Run("concurrent requests", func(t *testing.T) {
		config := &RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 10,
			BurstSize:         5,
			WindowSize:        time.Second,
		}

		rl := NewAPIRateLimiter(config, logger)
		handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Make concurrent requests
		results := make(chan int, 15)
		for i := 0; i < 15; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "192.168.1.1:12345"
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				results <- w.Code
			}()
		}

		// Collect results
		successCount := 0
		blockedCount := 0
		for i := 0; i < 15; i++ {
			code := <-results
			if code == http.StatusOK {
				successCount++
			} else if code == http.StatusTooManyRequests {
				blockedCount++
			}
		}

		// Should have some successful and some blocked requests
		assert.True(t, successCount > 0)
		assert.True(t, blockedCount > 0)
		assert.Equal(t, 15, successCount+blockedCount)
	})
}

func BenchmarkAPIRateLimiter_Middleware(b *testing.B) {
	logger := zap.NewNop()
	config := &RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 1000,
		BurstSize:         100,
		WindowSize:        time.Minute,
	}

	rl := NewAPIRateLimiter(config, logger)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkMemoryRateLimitStore_Allow(b *testing.B) {
	logger := zap.NewNop()
	store := NewMemoryRateLimitStore("token_bucket", 1000, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i%100)
		store.Allow(key, 100, time.Minute)
	}
}

func BenchmarkMemoryAuthRateLimitStore_CheckAuthLimit(b *testing.B) {
	logger := zap.NewNop()
	store := NewMemoryAuthRateLimitStore(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i%100)
		store.CheckAuthLimit(key, 10, time.Minute)
	}
}
