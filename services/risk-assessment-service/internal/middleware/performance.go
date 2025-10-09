package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/monitoring"
)

// PerformanceMiddleware provides performance monitoring for HTTP requests
type PerformanceMiddleware struct {
	monitor *monitoring.PerformanceMonitor
	logger  *zap.Logger
}

// NewPerformanceMiddleware creates a new performance middleware
func NewPerformanceMiddleware(monitor *monitoring.PerformanceMonitor, logger *zap.Logger) *PerformanceMiddleware {
	return &PerformanceMiddleware{
		monitor: monitor,
		logger:  logger,
	}
}

// Middleware returns the performance monitoring middleware function
func (pm *PerformanceMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapper := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process request
			next.ServeHTTP(wrapper, r)

			// Calculate duration
			duration := time.Since(start)

			// Determine if request was successful
			isError := wrapper.statusCode >= 400

			// Record metrics
			pm.monitor.RecordRequest(duration, isError)

			// Log request details
			pm.logger.Info("Request processed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status_code", wrapper.statusCode),
				zap.Duration("duration", duration),
				zap.Bool("is_error", isError),
				zap.String("user_agent", r.UserAgent()),
				zap.String("remote_addr", r.RemoteAddr))

			// Add performance headers
			w.Header().Set("X-Response-Time", duration.String())
			w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))
		})
	}
}

// responseWriterWrapper wraps http.ResponseWriter to capture status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Write implements the http.ResponseWriter interface
func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}
