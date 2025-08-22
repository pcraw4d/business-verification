package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// AdvancedMetricsHandler provides HTTP endpoints for advanced metrics access
type AdvancedMetricsHandler struct {
	advancedMetricsCollector *observability.AdvancedMetricsCollector
	logger                   *observability.Logger
}

// NewAdvancedMetricsHandler creates a new advanced metrics handler
func NewAdvancedMetricsHandler(advancedMetricsCollector *observability.AdvancedMetricsCollector, logger *observability.Logger) *AdvancedMetricsHandler {
	return &AdvancedMetricsHandler{
		advancedMetricsCollector: advancedMetricsCollector,
		logger:                   logger,
	}
}

// GetAdvancedMetricsSummary returns a summary of advanced metrics
func (h *AdvancedMetricsHandler) GetAdvancedMetricsSummary(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get advanced metrics summary
	summary := h.advancedMetricsCollector.GetMetricsSummary()

	// Add response metadata
	response := map[string]interface{}{
		"metrics":     summary,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"endpoint":    "/api/v3/advanced-metrics/summary",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode advanced metrics summary response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Advanced metrics summary requested",
		"method", r.Method,
		"path", r.URL.Path,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
}

// GetRealTimeMetrics returns real-time system metrics
func (h *AdvancedMetricsHandler) GetRealTimeMetrics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Collect current metrics
	metrics := h.advancedMetricsCollector.CollectMetrics()

	// Extract real-time metrics
	var realTimeMetrics interface{}
	if metrics != nil && metrics.AdvancedRealTimeMetrics != nil {
		realTimeMetrics = metrics.AdvancedRealTimeMetrics
	} else {
		realTimeMetrics = map[string]interface{}{
			"status": "no_real_time_metrics_available",
		}
	}

	// Add response metadata
	response := map[string]interface{}{
		"real_time_metrics": realTimeMetrics,
		"timestamp":         time.Now().UTC().Format(time.RFC3339),
		"endpoint":          "/api/v3/advanced-metrics/real-time",
		"duration_ms":       time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode real-time metrics response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Real-time metrics requested",
		"method", r.Method,
		"path", r.URL.Path,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
}

// GetBusinessIntelligenceMetrics returns business intelligence metrics
func (h *AdvancedMetricsHandler) GetBusinessIntelligenceMetrics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Collect current metrics
	metrics := h.advancedMetricsCollector.CollectMetrics()

	// Extract business intelligence metrics
	var businessMetrics interface{}
	if metrics != nil && metrics.AdvancedBusinessIntelligenceMetrics != nil {
		businessMetrics = metrics.AdvancedBusinessIntelligenceMetrics
	} else {
		businessMetrics = map[string]interface{}{
			"status": "no_business_intelligence_metrics_available",
		}
	}

	// Add response metadata
	response := map[string]interface{}{
		"business_intelligence_metrics": businessMetrics,
		"timestamp":                     time.Now().UTC().Format(time.RFC3339),
		"endpoint":                      "/api/v3/advanced-metrics/business-intelligence",
		"duration_ms":                   time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode business intelligence metrics response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Business intelligence metrics requested",
		"method", r.Method,
		"path", r.URL.Path,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
}

// GetPerformanceOptimizationMetrics returns performance optimization metrics
func (h *AdvancedMetricsHandler) GetPerformanceOptimizationMetrics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Collect current metrics
	metrics := h.advancedMetricsCollector.CollectMetrics()

	// Extract performance optimization metrics
	var performanceMetrics interface{}
	if metrics != nil && metrics.AdvancedPerformanceOptimizationMetrics != nil {
		performanceMetrics = metrics.AdvancedPerformanceOptimizationMetrics
	} else {
		performanceMetrics = map[string]interface{}{
			"status": "no_performance_optimization_metrics_available",
		}
	}

	// Add response metadata
	response := map[string]interface{}{
		"performance_optimization_metrics": performanceMetrics,
		"timestamp":                        time.Now().UTC().Format(time.RFC3339),
		"endpoint":                         "/api/v3/advanced-metrics/performance-optimization",
		"duration_ms":                      time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode performance optimization metrics response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Performance optimization metrics requested",
		"method", r.Method,
		"path", r.URL.Path,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
}

// GetQualityMetrics returns quality assurance metrics
func (h *AdvancedMetricsHandler) GetQualityMetrics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Collect current metrics
	metrics := h.advancedMetricsCollector.CollectMetrics()

	// Extract quality metrics
	var qualityMetrics interface{}
	if metrics != nil && metrics.AdvancedQualityMetrics != nil {
		qualityMetrics = metrics.AdvancedQualityMetrics
	} else {
		qualityMetrics = map[string]interface{}{
			"status": "no_quality_metrics_available",
		}
	}

	// Add response metadata
	response := map[string]interface{}{
		"quality_metrics": qualityMetrics,
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
		"endpoint":        "/api/v3/advanced-metrics/quality",
		"duration_ms":     time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode quality metrics response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Quality metrics requested",
		"method", r.Method,
		"path", r.URL.Path,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
}

// GetPredictiveMetrics returns predictive analytics metrics
func (h *AdvancedMetricsHandler) GetPredictiveMetrics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Collect current metrics
	metrics := h.advancedMetricsCollector.CollectMetrics()

	// Extract predictive metrics
	var predictiveMetrics interface{}
	if metrics != nil && metrics.AdvancedPredictiveMetrics != nil {
		predictiveMetrics = metrics.AdvancedPredictiveMetrics
	} else {
		predictiveMetrics = map[string]interface{}{
			"status": "no_predictive_metrics_available",
		}
	}

	// Add response metadata
	response := map[string]interface{}{
		"predictive_metrics": predictiveMetrics,
		"timestamp":          time.Now().UTC().Format(time.RFC3339),
		"endpoint":           "/api/v3/advanced-metrics/predictive",
		"duration_ms":        time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode predictive metrics response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Predictive metrics requested",
		"method", r.Method,
		"path", r.URL.Path,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
}

// GetTrendMetrics returns trend analysis metrics
func (h *AdvancedMetricsHandler) GetTrendMetrics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Collect current metrics
	metrics := h.advancedMetricsCollector.CollectMetrics()

	// Extract trend metrics
	var trendMetrics interface{}
	if metrics != nil && metrics.AdvancedTrendMetrics != nil {
		trendMetrics = metrics.AdvancedTrendMetrics
	} else {
		trendMetrics = map[string]interface{}{
			"status": "no_trend_metrics_available",
		}
	}

	// Add response metadata
	response := map[string]interface{}{
		"trend_metrics": trendMetrics,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"endpoint":      "/api/v3/advanced-metrics/trends",
		"duration_ms":   time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode trend metrics response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Trend metrics requested",
		"method", r.Method,
		"path", r.URL.Path,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
}

// GetMetricsHistory returns historical metrics data
func (h *AdvancedMetricsHandler) GetMetricsHistory(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default limit

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get historical data
	history := h.advancedMetricsCollector.GetMetricsHistory()

	// Limit the number of entries returned
	if len(history) > limit {
		history = history[len(history)-limit:]
	}

	// Add response metadata
	response := map[string]interface{}{
		"history":       history,
		"total_entries": len(history),
		"limit":         limit,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"endpoint":      "/api/v3/advanced-metrics/history",
		"duration_ms":   time.Since(startTime).Milliseconds(),
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
		"limit", limit,
		"total_entries", len(history),
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
}

// GetMetricsByCategory returns metrics filtered by category
func (h *AdvancedMetricsHandler) GetMetricsByCategory(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse query parameters
	category := r.URL.Query().Get("category")
	if category == "" {
		http.Error(w, "category parameter is required", http.StatusBadRequest)
		return
	}

	// Collect current metrics
	metrics := h.advancedMetricsCollector.CollectMetrics()

	// Filter metrics by category
	var categoryMetrics interface{}
	switch category {
	case "real-time":
		if metrics != nil && metrics.AdvancedRealTimeMetrics != nil {
			categoryMetrics = metrics.AdvancedRealTimeMetrics
		}
	case "business-intelligence":
		if metrics != nil && metrics.AdvancedBusinessIntelligenceMetrics != nil {
			categoryMetrics = metrics.AdvancedBusinessIntelligenceMetrics
		}
	case "performance-optimization":
		if metrics != nil && metrics.AdvancedPerformanceOptimizationMetrics != nil {
			categoryMetrics = metrics.AdvancedPerformanceOptimizationMetrics
		}
	case "quality":
		if metrics != nil && metrics.AdvancedQualityMetrics != nil {
			categoryMetrics = metrics.AdvancedQualityMetrics
		}
	case "predictive":
		if metrics != nil && metrics.AdvancedPredictiveMetrics != nil {
			categoryMetrics = metrics.AdvancedPredictiveMetrics
		}
	case "trends":
		if metrics != nil && metrics.AdvancedTrendMetrics != nil {
			categoryMetrics = metrics.AdvancedTrendMetrics
		}
	default:
		categoryMetrics = map[string]interface{}{
			"error": "invalid category",
			"valid_categories": []string{
				"real-time",
				"business-intelligence",
				"performance-optimization",
				"quality",
				"predictive",
				"trends",
			},
		}
	}

	if categoryMetrics == nil {
		categoryMetrics = map[string]interface{}{
			"status":   "no_metrics_available_for_category",
			"category": category,
		}
	}

	// Add response metadata
	response := map[string]interface{}{
		"category":    category,
		"metrics":     categoryMetrics,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"endpoint":    "/api/v3/advanced-metrics/category",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode category metrics response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Category metrics requested",
		"method", r.Method,
		"path", r.URL.Path,
		"category", category,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
}

// GetMetricsConfiguration returns the current metrics configuration
func (h *AdvancedMetricsHandler) GetMetricsConfiguration(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get configuration from the collector
	collector := h.advancedMetricsCollector
	config := map[string]interface{}{
		"collection_interval": collector.collectionInterval.String(),
		"aggregation_window":  collector.aggregationWindow.String(),
		"retention_period":    collector.retentionPeriod.String(),
		"max_history_size":    collector.maxHistorySize,
		"metrics_enabled":     true, // Assuming enabled if handler exists
	}

	// Add response metadata
	response := map[string]interface{}{
		"configuration": config,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"endpoint":      "/api/v3/advanced-metrics/configuration",
		"duration_ms":   time.Since(startTime).Milliseconds(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode metrics configuration response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("Metrics configuration requested",
		"method", r.Method,
		"path", r.URL.Path,
		"duration", time.Since(startTime),
		"status_code", http.StatusOK,
	)
}
