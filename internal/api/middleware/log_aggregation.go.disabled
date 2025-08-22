package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"go.uber.org/zap"
)

// LogAggregationMiddleware provides HTTP request log aggregation
type LogAggregationMiddleware struct {
	logAggregation *observability.LogAggregationSystem
	logger         *zap.Logger
	environment    string
}

// NewLogAggregationMiddleware creates a new log aggregation middleware
func NewLogAggregationMiddleware(logAggregation *observability.LogAggregationSystem, logger *zap.Logger, environment string) *LogAggregationMiddleware {
	return &LogAggregationMiddleware{
		logAggregation: logAggregation,
		logger:         logger,
		environment:    environment,
	}
}

// LogHTTPRequests wraps HTTP handlers with log aggregation
func (lam *LogAggregationMiddleware) LogHTTPRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create context with correlation IDs
		ctx := r.Context()
		ctx = lam.addCorrelationIDs(ctx, r)

		// Create a response writer wrapper to capture status code
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(wrappedWriter, r.WithContext(ctx))

		// Calculate duration
		duration := time.Since(start)

		// Log HTTP request details
		lam.logAggregation.LogHTTPRequest(ctx, r, wrappedWriter.statusCode, duration)

		// Create and ship log entry
		lam.shipLogEntry(ctx, r, wrappedWriter, duration)
	})
}

// addCorrelationIDs adds correlation IDs to the context
func (lam *LogAggregationMiddleware) addCorrelationIDs(ctx context.Context, r *http.Request) context.Context {
	// Add request ID if not present
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		ctx = context.WithValue(ctx, "request_id", requestID)
	} else {
		// Generate request ID if not provided
		requestID := generateRequestID()
		ctx = context.WithValue(ctx, "request_id", requestID)
	}

	// Add trace ID if present
	if traceID := r.Header.Get("X-Trace-ID"); traceID != "" {
		ctx = context.WithValue(ctx, "trace_id", traceID)
	}

	// Add span ID if present
	if spanID := r.Header.Get("X-Span-ID"); spanID != "" {
		ctx = context.WithValue(ctx, "span_id", spanID)
	}

	// Add user ID if present (from JWT token or header)
	if userID := lam.extractUserID(r); userID != "" {
		ctx = context.WithValue(ctx, "user_id", userID)
	}

	return ctx
}

// extractUserID extracts user ID from request
func (lam *LogAggregationMiddleware) extractUserID(r *http.Request) string {
	// Check X-User-ID header
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}

	// Check Authorization header for JWT token
	if auth := r.Header.Get("Authorization"); auth != "" {
		// This would extract user ID from JWT token
		// For now, return empty string
		return ""
	}

	return ""
}

// shipLogEntry creates and ships a log entry
func (lam *LogAggregationMiddleware) shipLogEntry(ctx context.Context, r *http.Request, w *responseWriter, duration time.Duration) {
	// Create log entry
	entry := lam.logAggregation.CreateLogEntry(ctx, "INFO", "HTTP request completed", map[string]interface{}{
		"method":       r.Method,
		"path":         r.URL.Path,
		"query":        r.URL.RawQuery,
		"status_code":  w.statusCode,
		"duration_ms":  duration.Milliseconds(),
		"user_agent":   r.UserAgent(),
		"remote_addr":  r.RemoteAddr,
		"referer":      r.Referer(),
		"content_type": r.Header.Get("Content-Type"),
		"accept":       r.Header.Get("Accept"),
		"endpoint":     r.URL.Path,
		"duration":     duration.Seconds(),
		"ip_address":   r.RemoteAddr,
	})

	// Ship log entry if log shipper is available
	if logShipper := lam.logAggregation.CreateLogShipper(); logShipper != nil {
		logShipper.ShipLog(entry)
	}
}

// LogBusinessEvents logs business-specific events
func (lam *LogAggregationMiddleware) LogBusinessEvents(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Log business events for specific endpoints
		lam.logBusinessEvent(ctx, r)

		next.ServeHTTP(w, r)
	})
}

// logBusinessEvent logs business events for specific endpoints
func (lam *LogAggregationMiddleware) logBusinessEvent(ctx context.Context, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/v1/classify":
		lam.logAggregation.LogBusinessEvent(ctx, "classification", "business_classification_request", map[string]interface{}{
			"endpoint": path,
			"method":   r.Method,
		})

	case path == "/v1/risk/assess":
		lam.logAggregation.LogBusinessEvent(ctx, "risk_assessment", "risk_assessment_request", map[string]interface{}{
			"endpoint": path,
			"method":   r.Method,
		})

	case path == "/v1/compliance/check":
		lam.logAggregation.LogBusinessEvent(ctx, "compliance", "compliance_check_request", map[string]interface{}{
			"endpoint": path,
			"method":   r.Method,
		})

	case path == "/v1/auth/login":
		lam.logAggregation.LogBusinessEvent(ctx, "authentication", "user_login_attempt", map[string]interface{}{
			"endpoint": path,
			"method":   r.Method,
		})

	case path == "/v1/auth/register":
		lam.logAggregation.LogBusinessEvent(ctx, "authentication", "user_registration_attempt", map[string]interface{}{
			"endpoint": path,
			"method":   r.Method,
		})
	}
}

// LogSecurityEvents logs security-related events
func (lam *LogAggregationMiddleware) LogSecurityEvents(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Log security events
		lam.logSecurityEvent(ctx, r)

		next.ServeHTTP(w, r)
	})
}

// logSecurityEvent logs security events
func (lam *LogAggregationMiddleware) logSecurityEvent(ctx context.Context, r *http.Request) {
	// Log failed authentication attempts
	if r.URL.Path == "/v1/auth/login" {
		// This would be logged after the request is processed
		// For now, we'll log the attempt
		lam.logAggregation.LogSecurityEvent(ctx, "authentication", "login_attempt", "info", map[string]interface{}{
			"endpoint":   r.URL.Path,
			"method":     r.Method,
			"ip_address": r.RemoteAddr,
			"user_agent": r.UserAgent(),
			"timestamp":  time.Now().UTC(),
		})
	}

	// Log API key usage
	if apiKey := lam.extractAPIKey(r); apiKey != "" {
		lam.logAggregation.LogSecurityEvent(ctx, "api_usage", "api_key_used", "info", map[string]interface{}{
			"endpoint":   r.URL.Path,
			"method":     r.Method,
			"api_key_id": apiKey,
			"ip_address": r.RemoteAddr,
			"timestamp":  time.Now().UTC(),
		})
	}
}

// LogPerformanceEvents logs performance-related events
func (lam *LogAggregationMiddleware) LogPerformanceEvents(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Process request
		next.ServeHTTP(w, r)

		// Calculate duration
		duration := time.Since(start)

		// Log performance event for slow requests
		if duration > 1*time.Second {
			ctx := r.Context()
			lam.logAggregation.LogPerformanceEvent(ctx, "http_request", duration, map[string]interface{}{
				"endpoint":     r.URL.Path,
				"method":       r.Method,
				"duration_ms":  duration.Milliseconds(),
				"threshold_ms": 1000,
			})
		}
	})
}

// LogDatabaseEvents logs database-related events
func (lam *LogAggregationMiddleware) LogDatabaseEvents(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This middleware would be used to log database events
		// For now, we'll just pass through
		next.ServeHTTP(w, r)
	})
}

// LogExternalAPIEvents logs external API calls
func (lam *LogAggregationMiddleware) LogExternalAPIEvents(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This middleware would be used to log external API calls
		// For now, we'll just pass through
		next.ServeHTTP(w, r)
	})
}

// extractAPIKey extracts API key from request
func (lam *LogAggregationMiddleware) extractAPIKey(r *http.Request) string {
	// Check Authorization header
	if auth := r.Header.Get("Authorization"); auth != "" {
		if len(auth) > 7 && auth[:7] == "Bearer " {
			return auth[7:]
		}
	}

	// Check X-API-Key header
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// Check query parameter
	if apiKey := r.URL.Query().Get("api_key"); apiKey != "" {
		return apiKey
	}

	return ""
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// This would generate a unique request ID
	// For now, return a simple timestamp-based ID
	return time.Now().Format("20060102150405") + "-" + fmt.Sprintf("%d", time.Now().UnixNano()%1000)
}
