package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// RequestIDKey is the context key for request ID
type RequestIDKey string

const RequestIDContextKey RequestIDKey = "request_id"

// Middleware provides common middleware functionality
type Middleware struct {
	logger *zap.Logger
}

// NewMiddleware creates a new middleware instance
func NewMiddleware(logger *zap.Logger) *Middleware {
	return &Middleware{
		logger: logger,
	}
}

// LoggingMiddleware logs HTTP requests with enhanced structured logging
func (m *Middleware) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get correlation and request IDs
		correlationID := r.Header.Get("X-Correlation-ID")
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		if correlationID == "" {
			correlationID = requestID
		}

		// Extract user and tenant information from headers or context
		userID := r.Header.Get("X-User-ID")
		tenantID := r.Header.Get("X-Tenant-ID")

		// Add IDs to context
		ctx := context.WithValue(r.Context(), RequestIDContextKey, requestID)
		ctx = context.WithValue(ctx, "correlation_id", correlationID)
		ctx = context.WithValue(ctx, "user_id", userID)
		ctx = context.WithValue(ctx, "tenant_id", tenantID)
		r = r.WithContext(ctx)

		// Set headers
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Correlation-ID", correlationID)

		// Read request body for logging (if not too large)
		var requestBody []byte
		if r.Body != nil && r.ContentLength < 1024*1024 { // 1MB limit
			requestBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create response writer wrapper to capture status code and body
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK, body: &bytes.Buffer{}}

		// Log request with enhanced structured data
		requestFields := []zap.Field{
			zap.String("correlation_id", correlationID),
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.String("referer", r.Referer()),
			zap.String("content_type", r.Header.Get("Content-Type")),
			zap.Int64("content_length", r.ContentLength),
		}

		// Add user context if available
		if userID != "" {
			requestFields = append(requestFields, zap.String("user_id", userID))
		}
		if tenantID != "" {
			requestFields = append(requestFields, zap.String("tenant_id", tenantID))
		}

		// Add request body if available and not too large
		if len(requestBody) > 0 && len(requestBody) < 1024 { // Log body if < 1KB
			requestFields = append(requestFields, zap.String("request_body", string(requestBody)))
		}

		m.logger.Info("Request started", requestFields...)

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Log response with enhanced structured data
		duration := time.Since(start)
		responseFields := []zap.Field{
			zap.String("correlation_id", correlationID),
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status_code", wrapped.statusCode),
			zap.Duration("duration", duration),
			zap.Int64("response_size", wrapped.size),
			zap.String("response_content_type", w.Header().Get("Content-Type")),
		}

		// Add user context if available
		if userID != "" {
			responseFields = append(responseFields, zap.String("user_id", userID))
		}
		if tenantID != "" {
			responseFields = append(responseFields, zap.String("tenant_id", tenantID))
		}

		// Add response body if available and not too large
		if wrapped.body.Len() > 0 && wrapped.body.Len() < 1024 { // Log body if < 1KB
			responseFields = append(responseFields, zap.String("response_body", wrapped.body.String()))
		}

		// Log at appropriate level based on status code
		if wrapped.statusCode >= 500 {
			m.logger.Error("Request completed with server error", responseFields...)
		} else if wrapped.statusCode >= 400 {
			m.logger.Warn("Request completed with client error", responseFields...)
		} else {
			m.logger.Info("Request completed successfully", responseFields...)
		}
	})
}

// CORSMiddleware handles CORS headers
func (m *Middleware) CORSMiddleware(allowedOrigins []string, allowedMethods []string, allowedHeaders []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if isOriginAllowed(origin, allowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware implements rate limiting
func (m *Middleware) RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	// Simple in-memory rate limiter (in production, use Redis)
	rateLimiter := NewRateLimiter(requestsPerMinute)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			if !rateLimiter.Allow(clientIP) {
				m.logger.Warn("Rate limit exceeded",
					zap.String("client_ip", clientIP),
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
				)

				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", time.Now().Add(time.Minute).Format(time.RFC3339))

				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// Set rate limit headers
			remaining := rateLimiter.GetRemaining(clientIP)
			resetTime := rateLimiter.GetResetTime(clientIP)

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", resetTime.Format(time.RFC3339))

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityMiddleware adds security headers
func (m *Middleware) SecurityMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

			next.ServeHTTP(w, r)
		})
	}
}

// RequestSizeMiddleware limits request body size
func (m *Middleware) RequestSizeMiddleware(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxSize {
				m.logger.Warn("Request body too large",
					zap.Int64("content_length", r.ContentLength),
					zap.Int64("max_size", maxSize),
					zap.String("path", r.URL.Path),
				)

				http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
				return
			}

			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)

			next.ServeHTTP(w, r)
		})
	}
}

// TimeoutMiddleware adds request timeout
func (m *Middleware) TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a channel to signal completion
			done := make(chan struct{})

			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// Request completed normally
			case <-ctx.Done():
				// Request timed out
				m.logger.Warn("Request timeout",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.Duration("timeout", timeout),
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusRequestTimeout)
				fmt.Fprintf(w, `{"error":{"code":"REQUEST_TIMEOUT","message":"Request timeout"}}`)
			}
		})
	}
}

// RecoveryMiddleware recovers from panics
func (m *Middleware) RecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					m.logger.Error("Panic recovered",
						zap.Any("error", err),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
						zap.String("remote_addr", r.RemoteAddr),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, `{"error":{"code":"INTERNAL_ERROR","message":"Internal server error"}}`)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// HealthCheckMiddleware provides health check endpoint
func (m *Middleware) HealthCheckMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// MetricsMiddleware collects basic metrics
func (m *Middleware) MetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response writer wrapper
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK, body: &bytes.Buffer{}}

			next.ServeHTTP(wrapped, r)

			// Log metrics
			duration := time.Since(start)
			m.logger.Info("Request metrics",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status_code", wrapped.statusCode),
				zap.Duration("duration", duration),
				zap.Int64("response_size", wrapped.size),
			)
		})
	}
}

// Helper functions

// responseWriter wraps http.ResponseWriter to capture status code, size, and body
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
	body       *bytes.Buffer
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	// Write to the original response writer
	size, err := rw.ResponseWriter.Write(b)
	rw.size += int64(size)

	// Also write to our buffer for logging (if not too large)
	if rw.body != nil && rw.body.Len() < 1024 { // 1KB limit
		rw.body.Write(b)
	}

	return size, err
}

// isOriginAllowed checks if origin is in allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}

	return false
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	return ip
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDContextKey).(string); ok {
		return requestID
	}
	return ""
}
