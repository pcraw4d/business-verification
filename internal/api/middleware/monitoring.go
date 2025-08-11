package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"go.uber.org/zap"
)

// MonitoringMiddleware provides HTTP request monitoring
type MonitoringMiddleware struct {
	monitoring  *observability.MonitoringSystem
	logger      *zap.Logger
	environment string
}

// NewMonitoringMiddleware creates a new monitoring middleware
func NewMonitoringMiddleware(monitoring *observability.MonitoringSystem, logger *zap.Logger, environment string) *MonitoringMiddleware {
	return &MonitoringMiddleware{
		monitoring:  monitoring,
		logger:      logger,
		environment: environment,
	}
}

// MonitorHTTPRequests wraps HTTP handlers with monitoring
func (mm *MonitoringMiddleware) MonitorHTTPRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Record request start
		mm.monitoring.RecordHTTPRequestInFlight(r.Method, r.URL.Path, mm.environment, 1)

		// Process request
		next.ServeHTTP(wrappedWriter, r)

		// Calculate duration
		duration := time.Since(start)

		// Record request completion
		mm.monitoring.RecordHTTPRequestInFlight(r.Method, r.URL.Path, mm.environment, -1)

		// Record metrics
		statusCode := strconv.Itoa(wrappedWriter.statusCode)
		mm.monitoring.RecordHTTPRequest(r.Method, r.URL.Path, statusCode, mm.environment, duration)

		// Log request details
		mm.logger.Info("HTTP request completed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status_code", wrappedWriter.statusCode),
			zap.Duration("duration", duration),
			zap.String("user_agent", r.UserAgent()),
			zap.String("remote_addr", r.RemoteAddr),
		)

		// Record business metrics for specific endpoints
		mm.recordBusinessMetrics(r, wrappedWriter, duration)
	})
}

// recordBusinessMetrics records business-specific metrics
func (mm *MonitoringMiddleware) recordBusinessMetrics(r *http.Request, w *responseWriter, duration time.Duration) {
	path := r.URL.Path

	switch {
	case path == "/v1/classify":
		// Record classification metrics
		confidenceLevel := "unknown"
		if w.statusCode == http.StatusOK {
			confidenceLevel = "high" // This would be extracted from response
		}
		mm.monitoring.RecordClassificationRequest(r.Method, confidenceLevel, mm.environment, duration)

	case path == "/v1/risk/assess":
		// Record risk assessment metrics
		riskLevel := "unknown"
		if w.statusCode == http.StatusOK {
			riskLevel = "medium" // This would be extracted from response
		}
		mm.monitoring.RecordRiskAssessment(riskLevel, mm.environment, duration)

	case path == "/v1/compliance/check":
		// Record compliance check metrics
		framework := "unknown"
		status := "unknown"
		if w.statusCode == http.StatusOK {
			framework = "soc2"   // This would be extracted from request/response
			status = "compliant" // This would be extracted from response
		}
		mm.monitoring.RecordComplianceCheck(framework, status, mm.environment, duration)

	case path == "/v1/auth/login":
		// Record authentication metrics
		eventType := "login"
		status := "success"
		if w.statusCode != http.StatusOK {
			status = "failed"
		}
		mm.monitoring.RecordAuthenticationEvent(eventType, status, mm.environment)
	}
}

// MonitorAPIKeyUsage monitors API key usage
func (mm *MonitoringMiddleware) MonitorAPIKeyUsage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract API key from request
		apiKey := extractAPIKey(r)
		if apiKey != "" {
			mm.monitoring.RecordAPIKeyUsage(apiKey, r.URL.Path, mm.environment)
		}

		next.ServeHTTP(w, r)
	})
}

// MonitorRateLimiting monitors rate limit hits
func (mm *MonitoringMiddleware) MonitorRateLimiting(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This would be called when rate limit is hit
		// For now, we'll just pass through
		next.ServeHTTP(w, r)
	})
}

// RecordRateLimitHit records when rate limit is hit
func (mm *MonitoringMiddleware) RecordRateLimitHit(apiKey, endpoint string) {
	mm.monitoring.RecordRateLimitHit(apiKey, endpoint, mm.environment)
}

// extractAPIKey extracts API key from request
func extractAPIKey(r *http.Request) string {
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

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
