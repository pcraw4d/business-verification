package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// RequestIDKey is the context key for request ID
const RequestIDKey = "request_id"

// RequestIDHeader is the HTTP header for request ID
const RequestIDHeader = "X-Request-ID"

// RequestIDMiddleware adds request ID to the context
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get request ID from header or generate new one
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = GenerateRequestID()
		}

		// Add request ID to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		// Add request ID to response headers
		w.Header().Set(RequestIDHeader, requestID)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GenerateRequestID generates a new request ID
func GenerateRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}

	return ""
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// ExtractRequestIDFromHeaders extracts request ID from HTTP headers
func ExtractRequestIDFromHeaders(headers map[string]string) string {
	for key, value := range headers {
		if strings.EqualFold(key, RequestIDHeader) {
			return value
		}
	}
	return ""
}

// InjectRequestIDIntoHeaders injects request ID into HTTP headers
func InjectRequestIDIntoHeaders(headers map[string]string, requestID string) {
	headers[RequestIDHeader] = requestID
}

// PropagateRequestID propagates request ID through context
func PropagateRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		requestID = GenerateRequestID()
	}
	return WithRequestID(ctx, requestID)
}

// RequestIDFromContext extracts request ID from context with fallback
func RequestIDFromContext(ctx context.Context) string {
	requestID := GetRequestID(ctx)
	if requestID == "" {
		requestID = GenerateRequestID()
	}
	return requestID
}

// RequestIDMiddlewareWithLogger adds request ID to the context and logs it
func RequestIDMiddlewareWithLogger(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get request ID from header or generate new one
			requestID := r.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = GenerateRequestID()
			}

			// Add request ID to context
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

			// Add request ID to response headers
			w.Header().Set(RequestIDHeader, requestID)

			// Log request with request ID
			logger.WithContext(ctx).WithFields(map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"user_agent":  r.UserAgent(),
				"remote_addr": r.RemoteAddr,
			}).Info("Request started")

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestIDMiddlewareWithTracer adds request ID to the context and traces it
// TODO: Re-enable when tracing is fully implemented
func RequestIDMiddlewareWithTracer(tracer interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get request ID from header or generate new one
			requestID := r.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = GenerateRequestID()
			}

			// Add request ID to context
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

			// TODO: Re-enable tracing when tracer is fully implemented
			// Start span with request information
			// spanCtx, span := tracer.StartSpanWithRequest(ctx, r.Method, r.URL.Path, r.UserAgent())
			// defer span.End()

			// Add request ID to span attributes
			// tracer.SetAttributes(spanCtx, map[string]interface{}{
			// 	"request_id": requestID,
			// })

			// Add request ID to response headers
			w.Header().Set(RequestIDHeader, requestID)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestIDMiddlewareWithMetrics adds request ID to the context and records metrics
func RequestIDMiddlewareWithMetrics(metrics *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get request ID from header or generate new one
			requestID := r.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = GenerateRequestID()
			}

			// Add request ID to context
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

			// Record request start
			metrics.RecordHTTPRequestStart(r.Method, r.URL.Path)

			// Create response writer wrapper to capture status code
			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Call next handler with updated context
			next.ServeHTTP(wrappedWriter, r.WithContext(ctx))

			// Record request end
			metrics.RecordHTTPRequestEnd(r.Method, r.URL.Path)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestIDMiddlewareWithAll adds request ID to the context with logging, tracing, and metrics
func RequestIDMiddlewareWithAll(logger *Logger, tracer interface{}, metrics *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request ID from header or generate new one
			requestID := r.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = GenerateRequestID()
			}

			// Add request ID to context
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

			// TODO: Re-enable tracing when tracer is fully implemented
			// Start span with request information
			// spanCtx, span := tracer.StartSpanWithRequest(ctx, r.Method, r.URL.Path, r.UserAgent())
			// defer span.End()

			// Add request ID to span attributes
			// tracer.SetAttributes(spanCtx, map[string]interface{}{
			// 	"request_id": requestID,
			// })

			// Log request start with request ID
			logger.WithContext(ctx).WithFields(map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"user_agent":  r.UserAgent(),
				"remote_addr": r.RemoteAddr,
			}).Info("Request started")

			// Record request start
			metrics.RecordHTTPRequestStart(r.Method, r.URL.Path)

			// Create response writer wrapper to capture status code
			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Call next handler with updated context
			next.ServeHTTP(wrappedWriter, r.WithContext(ctx))

			// Calculate duration
			duration := time.Since(start)

			// Record request end
			metrics.RecordHTTPRequestEnd(r.Method, r.URL.Path)

			// Record HTTP request metrics
			metrics.RecordHTTPRequest(r.Method, r.URL.Path, wrappedWriter.statusCode, duration)

			// Log request completion
			logger.WithContext(ctx).WithRequest(r.Method, r.URL.Path, r.UserAgent(), wrappedWriter.statusCode).
				WithDuration(duration).
				Info("Request completed")

			// Add request ID to response headers
			w.Header().Set(RequestIDHeader, requestID)
		})
	}
}
