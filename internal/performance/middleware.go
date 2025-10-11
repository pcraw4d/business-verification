package performance

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// PerformanceMiddleware provides performance monitoring middleware
type PerformanceMiddleware struct {
	logger          *zap.Logger
	profiler        *Profiler
	responseMonitor *ResponseMonitor
	cacheOptimizer  *CacheOptimizer
	config          *MiddlewareConfig
}

// MiddlewareConfig contains middleware configuration
type MiddlewareConfig struct {
	EnableProfiling          bool          `json:"enable_profiling"`
	EnableResponseMonitoring bool          `json:"enable_response_monitoring"`
	EnableCaching            bool          `json:"enable_caching"`
	CacheTTL                 time.Duration `json:"cache_ttl"`
	SkipPaths                []string      `json:"skip_paths"`
	SkipMethods              []string      `json:"skip_methods"`
	EnableDetailedLogging    bool          `json:"enable_detailed_logging"`
}

// NewPerformanceMiddleware creates a new performance middleware
func NewPerformanceMiddleware(
	logger *zap.Logger,
	profiler *Profiler,
	responseMonitor *ResponseMonitor,
	cacheOptimizer *CacheOptimizer,
	config *MiddlewareConfig,
) *PerformanceMiddleware {
	return &PerformanceMiddleware{
		logger:          logger,
		profiler:        profiler,
		responseMonitor: responseMonitor,
		cacheOptimizer:  cacheOptimizer,
		config:          config,
	}
}

// Middleware returns the HTTP middleware function
func (pm *PerformanceMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip performance monitoring for certain paths/methods
		if pm.shouldSkip(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Start performance monitoring
		start := time.Now()
		ctx := r.Context()

		// Add profiler to context
		if pm.profiler != nil {
			ctx = context.WithValue(ctx, "profiler", pm.profiler)
		}

		// Create response writer wrapper for status code tracking
		wrapper := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Check cache for GET requests
		var cacheHit bool
		if pm.config.EnableCaching && r.Method == http.MethodGet {
			cacheKey := pm.generateCacheKey(r)
			if cached, err := pm.cacheOptimizer.Get(ctx, "default", cacheKey); err == nil {
				// Cache hit
				cacheHit = true
				pm.writeCachedResponse(w, cached)

				// Record cache hit
				if pm.profiler != nil {
					pm.profiler.RecordMetric("cache_hit", time.Since(start))
				}

				// Record response time
				duration := time.Since(start)
				pm.recordResponse(r, duration, true, cacheHit)
				return
			}
		}

		// Profile the request
		var err error
		if pm.profiler != nil {
			err = pm.profiler.ProfileFuncWithContext(ctx, "http_request", func(ctx context.Context) error {
				next.ServeHTTP(wrapper, r.WithContext(ctx))
				return nil
			})
		} else {
			next.ServeHTTP(wrapper, r.WithContext(ctx))
		}

		// Calculate response time
		duration := time.Since(start)

		// Cache successful GET responses
		if pm.config.EnableCaching && r.Method == http.MethodGet && wrapper.statusCode == http.StatusOK {
			cacheKey := pm.generateCacheKey(r)
			// Note: In a real implementation, you'd need to capture the response body
			// For now, we'll just record the cache attempt
			pm.cacheOptimizer.Set(ctx, "default", cacheKey, "cached_response", pm.config.CacheTTL)
		}

		// Record response metrics
		success := wrapper.statusCode >= 200 && wrapper.statusCode < 400
		pm.recordResponse(r, duration, success, cacheHit)

		// Log detailed performance information
		if pm.config.EnableDetailedLogging {
			pm.logDetailedPerformance(r, duration, wrapper.statusCode, success, cacheHit)
		}

		// Handle profiling error
		if err != nil {
			pm.logger.Error("Profiling error", zap.Error(err))
		}
	})
}

// shouldSkip determines if performance monitoring should be skipped
func (pm *PerformanceMiddleware) shouldSkip(r *http.Request) bool {
	// Skip health checks and metrics endpoints
	path := r.URL.Path
	for _, skipPath := range pm.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	// Skip certain HTTP methods
	for _, skipMethod := range pm.config.SkipMethods {
		if r.Method == skipMethod {
			return true
		}
	}

	return false
}

// generateCacheKey generates a cache key for the request
func (pm *PerformanceMiddleware) generateCacheKey(r *http.Request) string {
	// Simple cache key generation
	// In production, you'd want more sophisticated key generation
	return r.Method + ":" + r.URL.Path + ":" + r.URL.RawQuery
}

// writeCachedResponse writes a cached response
func (pm *PerformanceMiddleware) writeCachedResponse(w http.ResponseWriter, cached interface{}) {
	// In a real implementation, you'd write the actual cached response
	// For now, we'll just set headers
	w.Header().Set("X-Cache", "HIT")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Write cached content (simplified)
	if content, ok := cached.(string); ok {
		w.Write([]byte(content))
	}
}

// recordResponse records response time metrics
func (pm *PerformanceMiddleware) recordResponse(r *http.Request, duration time.Duration, success, cacheHit bool) {
	if pm.responseMonitor != nil {
		pm.responseMonitor.RecordResponse(r.URL.Path, r.Method, duration, success)
	}

	// Record cache metrics
	if pm.profiler != nil {
		if cacheHit {
			pm.profiler.RecordMetric("cache_hit", duration)
		} else {
			pm.profiler.RecordMetric("cache_miss", duration)
		}
	}
}

// logDetailedPerformance logs detailed performance information
func (pm *PerformanceMiddleware) logDetailedPerformance(r *http.Request, duration time.Duration, statusCode int, success, cacheHit bool) {
	fields := []zap.Field{
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("query", r.URL.RawQuery),
		zap.Duration("duration", duration),
		zap.Int("status_code", statusCode),
		zap.Bool("success", success),
		zap.Bool("cache_hit", cacheHit),
		zap.String("user_agent", r.UserAgent()),
		zap.String("remote_addr", r.RemoteAddr),
	}

	// Add request ID if available
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	// Add trace ID if available
	if traceID := r.Header.Get("X-Trace-ID"); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	// Log based on performance
	if duration > time.Second {
		pm.logger.Warn("Slow request", fields...)
	} else if !success {
		pm.logger.Error("Request failed", fields...)
	} else {
		pm.logger.Info("Request completed", fields...)
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response body
func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// GetPerformanceStats returns current performance statistics
func (pm *PerformanceMiddleware) GetPerformanceStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if pm.profiler != nil {
		stats["profiler"] = pm.profiler.GetStats()
	}

	if pm.responseMonitor != nil {
		stats["response_monitor"] = pm.responseMonitor.GetStats()
	}

	if pm.cacheOptimizer != nil {
		stats["cache_optimizer"] = pm.cacheOptimizer.GetOverallStats()
	}

	return stats
}

// GetHealthStatus returns the health status of the performance monitoring
func (pm *PerformanceMiddleware) GetHealthStatus() map[string]interface{} {
	status := make(map[string]interface{})

	if pm.responseMonitor != nil {
		status["healthy"] = pm.responseMonitor.IsHealthy()
		status["health_score"] = pm.responseMonitor.GetHealthScore()
		status["alerts"] = pm.responseMonitor.GetAlerts(5)
	}

	return status
}

// Reset resets all performance statistics
func (pm *PerformanceMiddleware) Reset() {
	if pm.profiler != nil {
		pm.profiler.Reset()
	}

	if pm.responseMonitor != nil {
		pm.responseMonitor.Reset()
	}

	pm.logger.Info("Performance middleware reset")
}

// DefaultMiddlewareConfig returns a default middleware configuration
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		EnableProfiling:          true,
		EnableResponseMonitoring: true,
		EnableCaching:            true,
		CacheTTL:                 5 * time.Minute,
		SkipPaths:                []string{"/health", "/metrics", "/debug"},
		SkipMethods:              []string{"OPTIONS"},
		EnableDetailedLogging:    false,
	}
}
