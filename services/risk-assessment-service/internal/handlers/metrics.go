package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/engine"
	"kyb-platform/services/risk-assessment-service/internal/middleware"
)

// MetricsHandler handles metrics and performance monitoring
type MetricsHandler struct {
	riskEngine   *engine.RiskEngine
	logger       *zap.Logger
	errorHandler *middleware.ErrorHandler
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(riskEngine *engine.RiskEngine, logger *zap.Logger) *MetricsHandler {
	return &MetricsHandler{
		riskEngine:   riskEngine,
		logger:       logger,
		errorHandler: middleware.NewErrorHandler(logger),
	}
}

// HandleMetrics handles GET /api/v1/metrics
func (h *MetricsHandler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing metrics request")

	// Get engine metrics
	metrics := h.riskEngine.GetMetrics()
	if metrics == nil {
		h.errorHandler.HandleError(w, r, fmt.Errorf("metrics not available"))
		return
	}

	// Get cache stats
	cacheStats := h.riskEngine.GetCacheStats()

	// Get circuit breaker stats
	circuitBreakerStats := h.riskEngine.GetCircuitBreakerStats()

	// Get metrics stats
	metricsStats := metrics.GetStats()

	// Create comprehensive metrics response
	response := struct {
		Timestamp        time.Time                  `json:"timestamp"`
		EngineMetrics    engine.MetricsStats        `json:"engine_metrics"`
		CacheStats       engine.CacheStats          `json:"cache_stats"`
		CircuitBreaker   engine.CircuitBreakerStats `json:"circuit_breaker"`
		PerformanceCheck engine.PerformanceCheck    `json:"performance_check"`
	}{
		Timestamp:        time.Now(),
		EngineMetrics:    metricsStats,
		CacheStats:       cacheStats,
		CircuitBreaker:   circuitBreakerStats,
		PerformanceCheck: metrics.CheckPerformance(engine.DefaultPerformanceThresholds()),
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Metrics request completed")
}

// HandleHealth handles GET /api/v1/health
func (h *MetricsHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Processing health check request")

	// Get basic metrics
	metrics := h.riskEngine.GetMetrics()
	var healthStatus string
	var issues []string

	if metrics == nil {
		healthStatus = "unhealthy"
		issues = append(issues, "metrics not available")
	} else {
		// Check performance thresholds
		performanceCheck := metrics.CheckPerformance(engine.DefaultPerformanceThresholds())
		if performanceCheck.Passed {
			healthStatus = "healthy"
		} else {
			healthStatus = "degraded"
			issues = append(issues, performanceCheck.Issues...)
		}
	}

	// Create health response
	response := struct {
		Status    string    `json:"status"`
		Timestamp time.Time `json:"timestamp"`
		Issues    []string  `json:"issues,omitempty"`
	}{
		Status:    healthStatus,
		Timestamp: time.Now(),
		Issues:    issues,
	}

	// Set appropriate status code
	statusCode := http.StatusOK
	if healthStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if healthStatus == "degraded" {
		statusCode = http.StatusOK // Still operational but with issues
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)

	h.logger.Debug("Health check completed", zap.String("status", healthStatus))
}

// HandlePerformance handles GET /api/v1/performance
func (h *MetricsHandler) HandlePerformance(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing performance request")

	// Get engine metrics
	metrics := h.riskEngine.GetMetrics()
	if metrics == nil {
		h.errorHandler.HandleError(w, r, fmt.Errorf("metrics not available"))
		return
	}

	// Get performance check with thresholds
	performanceCheck := metrics.CheckPerformance(engine.DefaultPerformanceThresholds())
	stats := metrics.GetStats()

	// Create performance response
	response := struct {
		Timestamp        time.Time               `json:"timestamp"`
		PerformanceCheck engine.PerformanceCheck `json:"performance_check"`
		KeyMetrics       struct {
			AvgResponseTime   time.Duration `json:"avg_response_time"`
			SuccessRate       float64       `json:"success_rate"`
			CacheHitRate      float64       `json:"cache_hit_rate"`
			RequestsPerSecond float64       `json:"requests_per_second"`
			TotalRequests     int64         `json:"total_requests"`
		} `json:"key_metrics"`
		Thresholds struct {
			MaxResponseTime      time.Duration `json:"max_response_time"`
			MinSuccessRate       float64       `json:"min_success_rate"`
			MinCacheHitRate      float64       `json:"min_cache_hit_rate"`
			MinRequestsPerSecond float64       `json:"min_requests_per_second"`
		} `json:"thresholds"`
	}{
		Timestamp:        time.Now(),
		PerformanceCheck: performanceCheck,
		KeyMetrics: struct {
			AvgResponseTime   time.Duration `json:"avg_response_time"`
			SuccessRate       float64       `json:"success_rate"`
			CacheHitRate      float64       `json:"cache_hit_rate"`
			RequestsPerSecond float64       `json:"requests_per_second"`
			TotalRequests     int64         `json:"total_requests"`
		}{
			AvgResponseTime:   stats.AvgDuration,
			SuccessRate:       stats.SuccessRate,
			CacheHitRate:      stats.CacheHitRate,
			RequestsPerSecond: stats.RequestsPerSecond,
			TotalRequests:     stats.TotalRequests,
		},
		Thresholds: struct {
			MaxResponseTime      time.Duration `json:"max_response_time"`
			MinSuccessRate       float64       `json:"min_success_rate"`
			MinCacheHitRate      float64       `json:"min_cache_hit_rate"`
			MinRequestsPerSecond float64       `json:"min_requests_per_second"`
		}{
			MaxResponseTime:      1 * time.Second, // Sub-1-second target
			MinSuccessRate:       0.95,            // 95% success rate
			MinCacheHitRate:      0.8,             // 80% cache hit rate
			MinRequestsPerSecond: 100,             // 100 requests per second
		},
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Performance request completed")
}
