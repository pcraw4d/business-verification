package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
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
	summary := h.metricsAggregator.GetMetricsSummary()

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
		h.logger.Error("Failed to encode metrics summary response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Metrics summary requested",
		"method", r.Method,
		"path", r.URL.Path,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
}

// GetAggregatedMetrics returns detailed aggregated metrics
func (h *MetricsHandler) GetAggregatedMetrics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get aggregated metrics
	metrics := h.metricsAggregator.GetAggregatedMetrics()

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
		h.logger.Error("Failed to encode aggregated metrics response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Aggregated metrics requested",
		"method", r.Method,
		"path", r.URL.Path,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
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
	metrics := h.metricsAggregator.GetModuleMetrics(moduleID)
	if metrics == nil {
		http.Error(w, "Module not found", http.StatusNotFound)
		return
	}

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
		h.logger.Error("Failed to encode module metrics response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Module metrics requested",
		"method", r.Method,
		"path", r.URL.Path,
		"module_id", moduleID,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
}

// GetModuleList returns a list of all available modules
func (h *MetricsHandler) GetModuleList(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	// Get module list
	modules := h.metricsAggregator.GetModuleList()

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
		h.logger.Error("Failed to encode module list response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Module list requested",
		"method", r.Method,
		"path", r.URL.Path,
		"module_count", len(modules),
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
}

// GetHealthMetrics returns health-related metrics
func (h *MetricsHandler) GetHealthMetrics(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	// Get aggregated metrics
	metrics := h.metricsAggregator.GetAggregatedMetrics()

	// Extract health-related information
	healthMetrics := map[string]interface{}{
		"overall_health":       metrics.OverallHealth,
		"health_score":         metrics.HealthScore,
		"degraded_modules":     metrics.DegradedModules,
		"critical_modules":     metrics.CriticalModules,
		"overall_success_rate": metrics.OverallSuccessRate,
		"overall_error_rate":   metrics.OverallErrorRate,
		"total_requests":       metrics.TotalRequests,
		"successful_requests":  metrics.SuccessfulRequests,
		"failed_requests":      metrics.FailedRequests,
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
		h.logger.Error("Failed to encode health metrics response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Health metrics requested",
		"method", r.Method,
		"path", r.URL.Path,
		"health_score", metrics.HealthScore,
		"overall_health", metrics.OverallHealth,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
}

// GetPerformanceMetrics returns performance-related metrics
func (h *MetricsHandler) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	// Get aggregated metrics
	metrics := h.metricsAggregator.GetAggregatedMetrics()

	// Extract performance-related information
	performanceMetrics := map[string]interface{}{
		"average_response_time": metrics.AverageResponseTime.String(),
		"p95_response_time":     metrics.P95ResponseTime.String(),
		"p99_response_time":     metrics.P99ResponseTime.String(),
		"overall_throughput":    metrics.OverallThroughput,
		"total_memory_usage":    metrics.TotalMemoryUsage,
		"average_cpu_usage":     metrics.AverageCPUUsage,
		"total_goroutines":      metrics.TotalGoroutines,
		"database_connections":  metrics.DatabaseConnections,
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
		h.logger.Error("Failed to encode performance metrics response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Performance metrics requested",
		"method", r.Method,
		"path", r.URL.Path,
		"average_response_time", metrics.AverageResponseTime,
		"overall_throughput", metrics.OverallThroughput,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
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
	h.logger.Info("Prometheus metrics requested",
		"method", r.Method,
		"path", r.URL.Path,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
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
		metrics := h.metricsAggregator.GetModuleMetrics(moduleID)
		if metrics == nil {
			http.Error(w, "Module not found", http.StatusNotFound)
			return
		}
		historicalData = metrics
	} else {
		metrics := h.metricsAggregator.GetAggregatedMetrics()
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
		h.logger.Error("Failed to encode metrics history response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Metrics history requested",
		"method", r.Method,
		"path", r.URL.Path,
		"module_id", moduleID,
		"limit", limit,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
}
