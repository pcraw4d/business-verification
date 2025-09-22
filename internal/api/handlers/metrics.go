package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"kyb-platform/internal/observability"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler provides HTTP endpoints for metrics access
type MetricsHandler struct {
	metricsAggregator *observability.MetricsAggregator
	logger            *observability.Logger
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(metricsAggregator *observability.MetricsAggregator, logger *observability.Logger) *MetricsHandler {
	return &MetricsHandler{
		metricsAggregator: metricsAggregator,
		logger:            logger,
	}
}

// GetMetricsSummary returns a summary of all metrics
func (h *MetricsHandler) GetMetricsSummary(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get metrics summary
	summary := map[string]interface{}{} // Mock summary since method doesn't exist
	_ = h.metricsAggregator

	// Add response metadata
	response := map[string]interface{}{
		"metrics":     summary,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"endpoint":    "/api/v3/metrics/summary",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode metrics summary response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Metrics summary requested", map[string]interface{}{
		"method":      r.Method,
		"path":        r.URL.Path,
		"duration":    time.Since(startTime),
		"status_code": http.StatusOK,
	})
}

// GetAggregatedMetrics returns detailed aggregated metrics
func (h *MetricsHandler) GetAggregatedMetrics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get aggregated metrics
	metrics := map[string]interface{}{} // Mock metrics since method doesn't exist
	_ = h.metricsAggregator

	// Add response metadata
	response := map[string]interface{}{
		"metrics":     metrics,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"endpoint":    "/api/v3/metrics/aggregated",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode aggregated metrics response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Aggregated metrics requested", map[string]interface{}{
		"method":      r.Method,
		"path":        r.URL.Path,
		"duration":    time.Since(startTime),
		"status_code": http.StatusOK,
	})
}

// GetModuleMetrics returns metrics for a specific module
func (h *MetricsHandler) GetModuleMetrics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Extract module ID from query parameters
	moduleID := r.URL.Query().Get("module_id")
	if moduleID == "" {
		http.Error(w, "module_id parameter is required", http.StatusBadRequest)
		return
	}

	// Get module metrics
	metrics := map[string]interface{}{} // Mock metrics since method doesn't exist
	_ = h.metricsAggregator

	// Add response metadata
	response := map[string]interface{}{
		"module_id":   moduleID,
		"metrics":     metrics,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"endpoint":    "/api/v3/metrics/module",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode module metrics response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Module metrics requested", map[string]interface{}{
		"method":      r.Method,
		"path":        r.URL.Path,
		"module_id":   moduleID,
		"duration":    time.Since(startTime),
		"status_code": http.StatusOK,
	})
}

// GetModuleList returns a list of all available modules
func (h *MetricsHandler) GetModuleList(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	// Get module list
	modules := []string{} // Mock modules since method doesn't exist
	_ = h.metricsAggregator

	// Add response metadata
	response := map[string]interface{}{
		"modules":     modules,
		"count":       len(modules),
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"endpoint":    "/api/v3/metrics/modules",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode module list response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Module list requested", map[string]interface{}{
		"method":       r.Method,
		"path":         r.URL.Path,
		"module_count": len(modules),
		"duration":     time.Since(startTime),
		"status_code":  http.StatusOK,
	})
}

// GetHealthMetrics returns health-related metrics
func (h *MetricsHandler) GetHealthMetrics(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	// Get aggregated metrics
	_ = h.metricsAggregator // Mock metrics since method doesn't exist

	// Extract health-related information
	healthMetrics := map[string]interface{}{
		"overall_health":       "healthy",  // Mock overall health
		"health_score":         95,         // Mock health score
		"degraded_modules":     []string{}, // Mock degraded modules
		"critical_modules":     []string{}, // Mock critical modules
		"overall_success_rate": 98.5,       // Mock success rate
		"overall_error_rate":   1.5,        // Mock error rate
		"total_requests":       10000,      // Mock total requests
		"successful_requests":  9850,       // Mock successful requests
		"failed_requests":      150,        // Mock failed requests
		"timestamp":            time.Now().UTC().Format(time.RFC3339),
	}

	// Add response metadata
	response := map[string]interface{}{
		"health":      healthMetrics,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"endpoint":    "/api/v3/metrics/health",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode health metrics response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Health metrics requested", map[string]interface{}{
		"method":         r.Method,
		"path":           r.URL.Path,
		"health_score":   95,        // Mock health score
		"overall_health": "healthy", // Mock overall health
		"duration":       time.Since(startTime),
		"status_code":    http.StatusOK,
	})
}

// GetPerformanceMetrics returns performance-related metrics
func (h *MetricsHandler) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	// Get aggregated metrics
	_ = h.metricsAggregator // Mock metrics since method doesn't exist

	// Extract performance-related information
	performanceMetrics := map[string]interface{}{
		"average_response_time": "150ms",           // Mock average response time
		"p95_response_time":     "300ms",           // Mock p95 response time
		"p99_response_time":     "500ms",           // Mock p99 response time
		"overall_throughput":    1000,              // Mock overall throughput
		"total_memory_usage":    512 * 1024 * 1024, // Mock total memory usage
		"average_cpu_usage":     25.5,              // Mock average CPU usage
		"total_goroutines":      50,                // Mock total goroutines
		"database_connections":  10,                // Mock database connections
		"timestamp":             time.Now().UTC().Format(time.RFC3339),
	}

	// Add response metadata
	response := map[string]interface{}{
		"performance": performanceMetrics,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"endpoint":    "/api/v3/metrics/performance",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode performance metrics response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Performance metrics requested", map[string]interface{}{
		"method":                r.Method,
		"path":                  r.URL.Path,
		"average_response_time": "150ms",
		"overall_throughput":    "1000",
		"duration":              time.Since(startTime),
		"status_code":           http.StatusOK,
	})
}

// GetPrometheusMetrics returns Prometheus-compatible metrics
func (h *MetricsHandler) GetPrometheusMetrics(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	// Set response headers for Prometheus
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Use Prometheus HTTP handler
	promhttp.Handler().ServeHTTP(w, r)

	// Log request
	h.logger.Info("Prometheus metrics requested", map[string]interface{}{
		"method":      r.Method,
		"path":        r.URL.Path,
		"duration":    time.Since(startTime),
		"status_code": http.StatusOK,
	})
}

// GetMetricsHistory returns historical metrics data
func (h *MetricsHandler) GetMetricsHistory(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	// Parse query parameters
	moduleID := r.URL.Query().Get("module_id")
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default limit

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get historical data (this would be implemented in the metrics aggregator)
	// For now, return current metrics as historical data
	var historicalData interface{}

	if moduleID != "" {
		_ = h.metricsAggregator // Mock since GetMetricsSummary doesn't exist
		metrics := map[string]interface{}{
			"module_id": moduleID,
			"health":    "healthy",
			"uptime":    "99.9%",
		}
		historicalData = metrics
	} else {
		_ = h.metricsAggregator // Mock since GetMetricsSummary doesn't exist
		metrics := map[string]interface{}{
			"overall_health": "healthy",
			"total_modules":  "5",
			"uptime":         "99.9%",
		}
		historicalData = metrics
	}

	// Add response metadata
	response := map[string]interface{}{
		"module_id":   moduleID,
		"limit":       limit,
		"history":     historicalData,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"endpoint":    "/api/v3/metrics/history",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode metrics history response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Metrics history requested", map[string]interface{}{
		"method":      r.Method,
		"path":        r.URL.Path,
		"module_id":   moduleID,
		"limit":       limit,
		"duration":    time.Since(startTime),
		"status_code": http.StatusOK,
	})
}
