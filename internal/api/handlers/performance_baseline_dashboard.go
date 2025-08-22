package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// PerformanceBaselineDashboardHandler handles HTTP requests for performance baseline establishment
type PerformanceBaselineDashboardHandler struct {
	baselineSystem *observability.PerformanceBaselineEstablishmentSystem
	logger         *zap.Logger
}

// NewPerformanceBaselineDashboardHandler creates a new baseline dashboard handler
func NewPerformanceBaselineDashboardHandler(
	baselineSystem *observability.PerformanceBaselineEstablishmentSystem,
	logger *zap.Logger,
) *PerformanceBaselineDashboardHandler {
	return &PerformanceBaselineDashboardHandler{
		baselineSystem: baselineSystem,
		logger:         logger,
	}
}

// EstablishBaseline establishes a new performance baseline
func (h *PerformanceBaselineDashboardHandler) EstablishBaseline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request observability.BaselineEstablishmentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if request.MetricName == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	if request.Description == "" {
		request.Description = fmt.Sprintf("Baseline for %s", request.MetricName)
	}

	result, err := h.baselineSystem.EstablishBaseline(ctx, &request)
	if err != nil {
		h.logger.Error("Failed to establish baseline",
			zap.String("metric", request.MetricName),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to establish baseline: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetBaseline retrieves a performance baseline
func (h *PerformanceBaselineDashboardHandler) GetBaseline(w http.ResponseWriter, r *http.Request) {
	metricName := r.URL.Query().Get("metric")
	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	baseline, err := h.baselineSystem.GetBaseline(metricName)
	if err != nil {
		h.logger.Error("Failed to get baseline",
			zap.String("metric", metricName),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to get baseline: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(baseline)
}

// ListBaselines returns all performance baselines
func (h *PerformanceBaselineDashboardHandler) ListBaselines(w http.ResponseWriter, r *http.Request) {
	baselines := h.baselineSystem.ListBaselines()

	response := map[string]interface{}{
		"baselines": baselines,
		"count":     len(baselines),
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateBaseline updates an existing performance baseline
func (h *PerformanceBaselineDashboardHandler) UpdateBaseline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metricName := r.URL.Query().Get("metric")
	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	baseline, err := h.baselineSystem.UpdateBaseline(ctx, metricName, updates)
	if err != nil {
		h.logger.Error("Failed to update baseline",
			zap.String("metric", metricName),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update baseline: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(baseline)
}

// DeleteBaseline deletes a performance baseline
func (h *PerformanceBaselineDashboardHandler) DeleteBaseline(w http.ResponseWriter, r *http.Request) {
	metricName := r.URL.Query().Get("metric")
	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	err := h.baselineSystem.DeleteBaseline(metricName)
	if err != nil {
		h.logger.Error("Failed to delete baseline",
			zap.String("metric", metricName),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to delete baseline: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message":   "Baseline deleted successfully",
		"metric":    metricName,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateBaseline validates a performance baseline
func (h *PerformanceBaselineDashboardHandler) ValidateBaseline(w http.ResponseWriter, r *http.Request) {
	metricName := r.URL.Query().Get("metric")
	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	validation, err := h.baselineSystem.ValidateBaseline(metricName)
	if err != nil {
		h.logger.Error("Failed to validate baseline",
			zap.String("metric", metricName),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to validate baseline: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(validation)
}

// GetBaselineStatistics returns statistics about all baselines
func (h *PerformanceBaselineDashboardHandler) GetBaselineStatistics(w http.ResponseWriter, r *http.Request) {
	stats := h.baselineSystem.GetBaselineStatistics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// BulkEstablishBaselines establishes multiple baselines
func (h *PerformanceBaselineDashboardHandler) BulkEstablishBaselines(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		Metrics      []observability.BaselineEstablishmentRequest `json:"metrics"`
		ForceRefresh bool                                         `json:"force_refresh"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(request.Metrics) == 0 {
		http.Error(w, "No metrics provided", http.StatusBadRequest)
		return
	}

	results := make([]*observability.BaselineEstablishmentResult, 0, len(request.Metrics))
	successCount := 0
	errorCount := 0

	for _, metricRequest := range request.Metrics {
		if request.ForceRefresh {
			metricRequest.ForceRefresh = true
		}

		result, err := h.baselineSystem.EstablishBaseline(ctx, &metricRequest)
		if err != nil {
			h.logger.Error("Failed to establish baseline in bulk operation",
				zap.String("metric", metricRequest.MetricName),
				zap.Error(err))
			errorCount++
			continue
		}

		results = append(results, result)
		if result.Status == "success" {
			successCount++
		}
	}

	response := map[string]interface{}{
		"results":    results,
		"total":      len(request.Metrics),
		"successful": successCount,
		"failed":     errorCount,
		"timestamp":  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBaselineConfiguration returns the current baseline establishment configuration
func (h *PerformanceBaselineDashboardHandler) GetBaselineConfiguration(w http.ResponseWriter, r *http.Request) {
	// This would require exposing the config from the baseline system
	// For now, return a placeholder response
	response := map[string]interface{}{
		"message":   "Configuration endpoint not yet implemented",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateBaselineConfiguration updates the baseline establishment configuration
func (h *PerformanceBaselineDashboardHandler) UpdateBaselineConfiguration(w http.ResponseWriter, r *http.Request) {
	// This would require exposing the config update from the baseline system
	// For now, return a placeholder response
	response := map[string]interface{}{
		"message":   "Configuration update endpoint not yet implemented",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBaselineHealth returns health information about the baseline system
func (h *PerformanceBaselineDashboardHandler) GetBaselineHealth(w http.ResponseWriter, r *http.Request) {
	stats := h.baselineSystem.GetBaselineStatistics()

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"stats":     stats,
		"system": map[string]interface{}{
			"total_baselines":    stats.TotalBaselines,
			"active_baselines":   stats.ActiveBaselines,
			"valid_baselines":    stats.ValidBaselines,
			"average_confidence": stats.AverageConfidence,
		},
	}

	// Determine overall health status
	if stats.TotalBaselines == 0 {
		health["status"] = "warning"
		health["message"] = "No baselines established"
	} else if stats.ValidBaselines < stats.TotalBaselines/2 {
		health["status"] = "warning"
		health["message"] = "Many baselines have validation issues"
	} else if stats.AverageConfidence < 0.7 {
		health["status"] = "warning"
		health["message"] = "Low average baseline confidence"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// GetBaselineMetrics returns metrics about baseline establishment performance
func (h *PerformanceBaselineDashboardHandler) GetBaselineMetrics(w http.ResponseWriter, r *http.Request) {
	// This would require exposing metrics from the baseline system
	// For now, return a placeholder response
	response := map[string]interface{}{
		"message":   "Metrics endpoint not yet implemented",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExportBaselines exports baselines in various formats
func (h *PerformanceBaselineDashboardHandler) ExportBaselines(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	baselines := h.baselineSystem.ListBaselines()

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(baselines)
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=baselines.csv")
		// CSV export implementation would go here
		fmt.Fprintf(w, "metric,mean,median,std_dev,min,max,percentile_95,percentile_99,sample_size,confidence\n")
		for _, baseline := range baselines {
			fmt.Fprintf(w, "%s,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%d,%.3f\n",
				baseline.Metric,
				baseline.Mean,
				baseline.Median,
				baseline.StdDev,
				baseline.Min,
				baseline.Max,
				baseline.Percentile95,
				baseline.Percentile99,
				baseline.SampleSize,
				baseline.Confidence,
			)
		}
	default:
		http.Error(w, "Unsupported format", http.StatusBadRequest)
	}
}

// ImportBaselines imports baselines from external sources
func (h *PerformanceBaselineDashboardHandler) ImportBaselines(w http.ResponseWriter, r *http.Request) {
	// This would require implementing baseline import functionality
	// For now, return a placeholder response
	response := map[string]interface{}{
		"message":   "Import endpoint not yet implemented",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBaselineComparison compares baselines across different time periods or environments
func (h *PerformanceBaselineDashboardHandler) GetBaselineComparison(w http.ResponseWriter, r *http.Request) {
	metricName := r.URL.Query().Get("metric")
	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	// This would require implementing baseline comparison functionality
	// For now, return a placeholder response
	response := map[string]interface{}{
		"message":   "Comparison endpoint not yet implemented",
		"metric":    metricName,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBaselineTrends analyzes trends in baseline performance over time
func (h *PerformanceBaselineDashboardHandler) GetBaselineTrends(w http.ResponseWriter, r *http.Request) {
	metricName := r.URL.Query().Get("metric")
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}

	// This would require implementing trend analysis functionality
	// For now, return a placeholder response
	response := map[string]interface{}{
		"message":   "Trends endpoint not yet implemented",
		"metric":    metricName,
		"period":    period,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBaselineAlerts returns alerts related to baseline performance
func (h *PerformanceBaselineDashboardHandler) GetBaselineAlerts(w http.ResponseWriter, r *http.Request) {
	severity := r.URL.Query().Get("severity")
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	// This would require implementing alert functionality
	// For now, return a placeholder response
	response := map[string]interface{}{
		"message":   "Alerts endpoint not yet implemented",
		"severity":  severity,
		"limit":     limit,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBaselineReports generates reports about baseline performance
func (h *PerformanceBaselineDashboardHandler) GetBaselineReports(w http.ResponseWriter, r *http.Request) {
	reportType := r.URL.Query().Get("type")
	if reportType == "" {
		reportType = "summary"
	}

	// This would require implementing report generation functionality
	// For now, return a placeholder response
	response := map[string]interface{}{
		"message":   "Reports endpoint not yet implemented",
		"type":      reportType,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
