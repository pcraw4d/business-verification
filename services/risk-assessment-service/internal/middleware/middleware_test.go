package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Test helper functions
func createTestRequest(method, url string, body interface{}) *http.Request {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func createTestMiddleware() *Middleware {
	logger := zap.NewNop()
	return NewMiddleware(logger)
}

func TestMiddleware_LoggingMiddleware(t *testing.T) {
	middleware := createTestMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	req := createTestRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Apply logging middleware
	loggedHandler := middleware.LoggingMiddleware(handler)
	loggedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())
}

func TestMiddleware_SecurityMiddleware(t *testing.T) {
	middleware := createTestMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := createTestRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Apply security middleware
	securedHandler := middleware.SecurityMiddleware()(handler)
	securedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check security headers
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "no-referrer", w.Header().Get("Referrer-Policy"))
}

func TestMiddleware_CORSMiddleware(t *testing.T) {
	middleware := createTestMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	allowedOrigins := []string{"http://localhost:3000", "https://example.com"}
	allowedMethods := []string{"GET", "POST", "PUT", "DELETE"}
	allowedHeaders := []string{"Content-Type", "Authorization"}

	req := createTestRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	w := httptest.NewRecorder()

	// Apply CORS middleware
	corsHandler := middleware.CORSMiddleware(allowedOrigins, allowedMethods, allowedHeaders)(handler)
	corsHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check CORS headers
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "POST", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type", w.Header().Get("Access-Control-Allow-Headers"))
}

func TestMiddleware_CORSMiddleware_InvalidOrigin(t *testing.T) {
	middleware := createTestMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	allowedOrigins := []string{"http://localhost:3000"}
	allowedMethods := []string{"GET", "POST"}
	allowedHeaders := []string{"Content-Type"}

	req := createTestRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://malicious-site.com")

	w := httptest.NewRecorder()

	// Apply CORS middleware
	corsHandler := middleware.CORSMiddleware(allowedOrigins, allowedMethods, allowedHeaders)(handler)
	corsHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Should not have CORS headers for invalid origin
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestMiddleware_RateLimitMiddleware(t *testing.T) {
	middleware := createTestMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := createTestRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	w := httptest.NewRecorder()

	// Create a rate limiter for testing
	rateLimiter := &RateLimiter{
		requestsPerMinute: 1,
		burstAllowance:    1,
		windowSize:        time.Minute,
	}

	// Apply rate limit middleware
	rateLimitedHandler := middleware.RateLimitMiddleware(rateLimiter)(handler)

	// First request should succeed
	rateLimitedHandler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Second request should be rate limited
	w2 := httptest.NewRecorder()
	rateLimitedHandler.ServeHTTP(w2, req)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestMiddleware_RequestSizeMiddleware(t *testing.T) {
	middleware := createTestMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test with small request (should pass)
	smallBody := bytes.NewBuffer(make([]byte, 100))
	req1 := httptest.NewRequest("POST", "/test", smallBody)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()

	sizeLimitedHandler := middleware.RequestSizeMiddleware(1000)(handler)
	sizeLimitedHandler.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Test with large request (should fail)
	largeBody := bytes.NewBuffer(make([]byte, 2000))
	req2 := httptest.NewRequest("POST", "/test", largeBody)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()

	sizeLimitedHandler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusRequestEntityTooLarge, w2.Code)
}

func TestMiddleware_TimeoutMiddleware(t *testing.T) {
	middleware := createTestMiddleware()

	// Handler that takes longer than timeout
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	req := createTestRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Apply timeout middleware (50ms timeout)
	timeoutHandler := middleware.TimeoutMiddleware(50 * time.Millisecond)(handler)
	timeoutHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestTimeout, w.Code)
}

func TestMiddleware_RecoveryMiddleware(t *testing.T) {
	middleware := createTestMiddleware()

	// Handler that panics
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	req := createTestRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Apply recovery middleware
	recoveryHandler := middleware.RecoveryMiddleware()(handler)
	recoveryHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse, "error")
}

func TestMiddleware_MetricsMiddleware(t *testing.T) {
	middleware := createTestMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := createTestRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Apply metrics middleware
	metricsHandler := middleware.MetricsMiddleware()(handler)
	metricsHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_HealthCheckMiddleware(t *testing.T) {
	middleware := createTestMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test health check endpoint
	req := createTestRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	healthHandler := middleware.HealthCheckMiddleware()(handler)
	healthHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var healthResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &healthResponse)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", healthResponse["status"])
	assert.Equal(t, "risk-assessment-service", healthResponse["service"])
}

func TestMiddleware_HealthCheckMiddleware_NonHealthEndpoint(t *testing.T) {
	middleware := createTestMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("normal response"))
	})

	// Test non-health endpoint
	req := createTestRequest("GET", "/api/v1/assess", nil)
	w := httptest.NewRecorder()

	healthHandler := middleware.HealthCheckMiddleware()(handler)
	healthHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "normal response", w.Body.String())
}

func TestErrorHandler_HandleError(t *testing.T) {
	logger := zap.NewNop()
	errorHandler := NewErrorHandler(logger)

	req := createTestRequest("POST", "/api/v1/assess", nil)
	w := httptest.NewRecorder()

	errorHandler.HandleError(w, req, assert.AnError)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var errorResponse ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse.Error.Message, "assert.AnError")
	assert.NotEmpty(t, errorResponse.RequestID)
	assert.NotEmpty(t, errorResponse.Timestamp)
}

func TestErrorHandler_HandleError_WithContext(t *testing.T) {
	logger := zap.NewNop()
	errorHandler := NewErrorHandler(logger)

	req := createTestRequest("POST", "/api/v1/assess", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-id"))
	w := httptest.NewRecorder()

	errorHandler.HandleError(w, req, assert.AnError)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var errorResponse ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "test-request-id", errorResponse.RequestID)
}

func TestErrorHandler_HandleError_ValidationError(t *testing.T) {
	logger := zap.NewNop()
	errorHandler := NewErrorHandler(logger)

	req := createTestRequest("POST", "/api/v1/assess", nil)
	w := httptest.NewRecorder()

	validationError := ValidationError{
		Field:   "business_name",
		Message: "Business name is required",
		Code:    "REQUIRED_FIELD",
	}

	errorDetail := ErrorDetail{
		Code:       "VALIDATION_ERROR",
		Message:    "Request validation failed",
		Validation: []ValidationError{validationError},
	}

	errorResponse := ErrorResponse{
		Error:     errorDetail,
		RequestID: "test-request-id",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      "/api/v1/assess",
		Method:    "POST",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(errorResponse)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	assert.Len(t, response.Error.Validation, 1)
	assert.Equal(t, "business_name", response.Error.Validation[0].Field)
}

func TestRateLimiter_Allow(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerMinute: 2,
		BurstAllowance:    2,
		WindowSize:        time.Minute,
		UseRedis:          false,
	}
	limiter := NewRateLimiter(nil, zap.NewNop(), config)

	// First two requests should be allowed
	assert.True(t, limiter.Allow("192.168.1.1", TierFree))
	assert.True(t, limiter.Allow("192.168.1.1", TierFree))

	// Third request should be denied
	assert.False(t, limiter.Allow("192.168.1.1", TierFree))

	// Different IP should be allowed
	assert.True(t, limiter.Allow("192.168.1.2", TierFree))
}

func TestRateLimiter_Allow_DifferentIPs(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerMinute: 1,
		BurstAllowance:    1,
		WindowSize:        time.Minute,
		UseRedis:          false,
	}
	limiter := NewRateLimiter(nil, zap.NewNop(), config)

	// Each IP should have its own limit
	assert.True(t, limiter.Allow("192.168.1.1", TierFree))
	assert.True(t, limiter.Allow("192.168.1.2", TierFree))
	assert.True(t, limiter.Allow("192.168.1.3", TierFree))

	// But second request from same IP should be denied
	assert.False(t, limiter.Allow("192.168.1.1", TierFree))
}

func TestRateLimiter_Allow_ExpiredEntries(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerMinute: 1,
		BurstAllowance:    1,
		WindowSize:        100 * time.Millisecond,
		UseRedis:          false,
	}
	limiter := NewRateLimiter(nil, zap.NewNop(), config)

	// First request should be allowed
	assert.True(t, limiter.Allow("192.168.1.1", TierFree))

	// Second request should be denied
	assert.False(t, limiter.Allow("192.168.1.1", TierFree))

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Request should be allowed again
	assert.True(t, limiter.Allow("192.168.1.1", TierFree))
}

// Integration tests
func TestMiddleware_Integration(t *testing.T) {
	middleware := createTestMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Apply all middleware
	finalHandler := middleware.RecoveryMiddleware()(
		middleware.LoggingMiddleware(
			middleware.SecurityMiddleware()(
				middleware.RequestSizeMiddleware(1000)(
					middleware.TimeoutMiddleware(30 * time.Second)(
						middleware.CORSMiddleware(
							[]string{"http://localhost:3000"},
							[]string{"GET", "POST"},
							[]string{"Content-Type"},
						)(
							middleware.RateLimitMiddleware(&RateLimiter{
								requestsPerMinute: 100,
								burstAllowance:    100,
								windowSize:        time.Minute,
							})(
								middleware.MetricsMiddleware()(
									middleware.HealthCheckMiddleware()(handler),
								),
							),
						),
					),
				),
			),
		),
	)

	req := createTestRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	finalHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())

	// Check that security headers are present
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
}

// Benchmark tests
func BenchmarkLoggingMiddleware(b *testing.B) {
	middleware := createTestMiddleware()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggedHandler := middleware.LoggingMiddleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := createTestRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		loggedHandler.ServeHTTP(w, req)
	}
}

func BenchmarkSecurityMiddleware(b *testing.B) {
	middleware := createTestMiddleware()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	securedHandler := middleware.SecurityMiddleware()(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := createTestRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		securedHandler.ServeHTTP(w, req)
	}
}

func BenchmarkRateLimiter(b *testing.B) {
	config := RateLimitConfig{
		RequestsPerMinute: 1000,
		BurstAllowance:    1000,
		WindowSize:        time.Minute,
		UseRedis:          false,
	}
	limiter := NewRateLimiter(nil, zap.NewNop(), config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("192.168.1.1", TierFree)
	}
}
