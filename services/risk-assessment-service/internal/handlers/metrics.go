package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/middleware"
	"kyb-platform/services/risk-assessment-service/internal/ml/monitoring"
)

// MetricsHandler handles metrics-related HTTP requests
type MetricsHandler struct {
	metricsCollector *monitoring.MetricsCollector
	logger           *zap.Logger
	errorHandler     *middleware.ErrorHandler
}

// getRequestID safely extracts request ID from context
func (h *MetricsHandler) getRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value("request_id").(string); ok && reqID != "" {
		return reqID
	}
	return "unknown"
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(
	metricsCollector *monitoring.MetricsCollector,
	logger *zap.Logger,
) *MetricsHandler {
	return &MetricsHandler{
		metricsCollector: metricsCollector,
		logger:           logger,
		errorHandler:     middleware.NewErrorHandler(logger),
	}
}

// HandleGetMetrics returns risk metrics matching frontend RiskMetricsSchema
// Frontend expects: overallRiskScore, highRiskMerchants, riskAssessments, riskTrend, riskDistribution, timestamp
func (h *MetricsHandler) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check if this is a risk metrics request (routed from /api/v1/risk/metrics)
	// The gateway transforms /api/v1/risk/metrics to /api/v1/metrics
	// We can detect this by checking the Referer header or accept risk metrics format
	// For now, we'll return risk metrics format to match frontend expectations
	
	// TODO: Query actual risk_assessments table from Supabase to get real metrics
	// For now, return properly formatted mock data matching RiskMetricsSchema
	response := map[string]interface{}{
		"overallRiskScore": 0.65,  // Average risk score (0-1)
		"highRiskMerchants": 250,  // Count of merchants with high risk
		"riskAssessments": 4500,    // Total number of risk assessments
		"riskTrend": -0.05,         // Risk trend (negative = improving, positive = worsening)
		"riskDistribution": map[string]interface{}{
			"low":     0.2,  // 20%
			"medium":  0.6,  // 60%
			"high":    0.2,  // 20%
			"critical": 0.0, // 0% (optional)
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode risk metrics response", zap.Error(err))
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.logger.Info("Risk metrics requested",
		zap.String("request_id", h.getRequestID(ctx)),
		zap.Float64("overall_risk_score", 0.65),
		zap.Int("high_risk_merchants", 250),
	)
}

// HandleGetModelMetrics returns metrics for a specific model
func (h *MetricsHandler) HandleGetModelMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract model type from URL path
	modelType := r.URL.Query().Get("model_type")
	if modelType == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("model_type parameter is required"))
		return
	}

	// Get model metrics
	modelMetrics, exists := h.metricsCollector.GetModelMetrics(modelType)
	if !exists {
		h.errorHandler.HandleError(w, r, &middleware.NotFoundError{
			Resource: "Model: " + modelType,
		})
		return
	}

	// Create response
	response := map[string]interface{}{
		"model_metrics": modelMetrics,
		"timestamp":     time.Now(),
		"status":        "success",
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode model metrics response", zap.Error(err))
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.logger.Info("Model metrics requested",
		zap.String("request_id", h.getRequestID(ctx)),
		zap.String("model_type", modelType),
	)
}

// HandleGetPerformanceSnapshot returns a complete performance snapshot
func (h *MetricsHandler) HandleGetPerformanceSnapshot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get performance snapshot
	snapshot := h.metricsCollector.GetSnapshot()

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(snapshot); err != nil {
		h.logger.Error("Failed to encode performance snapshot response", zap.Error(err))
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.logger.Info("Performance snapshot requested",
		zap.String("request_id", h.getRequestID(ctx)),
		zap.String("health_status", snapshot.HealthStatus),
		zap.Int("alerts_count", len(snapshot.Alerts)),
	)
}

// HandleResetMetrics resets all metrics (admin only)
func (h *MetricsHandler) HandleResetMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check if this is a POST request
	if r.Method != http.MethodPost {
		h.errorHandler.HandleError(w, r, fmt.Errorf("only POST method is allowed"))
		return
	}

	// Reset metrics
	h.metricsCollector.Reset()

	// Create response
	response := map[string]interface{}{
		"message":   "Metrics reset successfully",
		"timestamp": time.Now(),
		"status":    "success",
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode reset metrics response", zap.Error(err))
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.logger.Info("Metrics reset requested",
		zap.String("request_id", h.getRequestID(ctx)),
	)
}

// HandleGetHealth returns health status
func (h *MetricsHandler) HandleGetHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get health status
	snapshot := h.metricsCollector.GetSnapshot()

	// Determine HTTP status code based on health
	statusCode := http.StatusOK
	if snapshot.HealthStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if snapshot.HealthStatus == "degraded" {
		statusCode = http.StatusOK // Still OK but with warnings
	}

	// Create response
	response := map[string]interface{}{
		"status":         snapshot.HealthStatus,
		"timestamp":      time.Now(),
		"uptime":         snapshot.OverallMetrics.Uptime,
		"total_requests": snapshot.OverallMetrics.TotalRequests,
		"error_rate":     snapshot.OverallMetrics.ErrorRate,
		"alerts_count":   len(snapshot.Alerts),
	}

	// Add alerts if any
	if len(snapshot.Alerts) > 0 {
		response["alerts"] = snapshot.Alerts
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode health response", zap.Error(err))
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.logger.Info("Health check requested",
		zap.String("request_id", h.getRequestID(ctx)),
		zap.String("health_status", snapshot.HealthStatus),
		zap.Int("status_code", statusCode),
	)
}

// HandleGetModelPerformance returns model performance metrics
func (h *MetricsHandler) HandleGetModelPerformance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get all model metrics
	snapshot := h.metricsCollector.GetSnapshot()

	// Create performance response
	performanceResponse := map[string]interface{}{
		"models":          snapshot.ModelMetrics,
		"overall_metrics": snapshot.OverallMetrics,
		"last_updated":    time.Now(),
		"health_status":   snapshot.HealthStatus,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(performanceResponse); err != nil {
		h.logger.Error("Failed to encode model performance response", zap.Error(err))
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.logger.Info("Model performance requested",
		zap.String("request_id", h.getRequestID(ctx)),
		zap.Int("models_count", len(snapshot.ModelMetrics)),
	)
}

// HandleGetMetricsHistory returns historical metrics (if available)
func (h *MetricsHandler) HandleGetMetricsHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	hoursStr := r.URL.Query().Get("hours")
	hours := 24 // default to 24 hours
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 && h <= 168 { // max 1 week
			hours = h
		}
	}

	// For now, return current snapshot as we don't have historical storage
	// In a production system, this would query a time-series database
	snapshot := h.metricsCollector.GetSnapshot()

	// Create response
	response := map[string]interface{}{
		"message":          "Historical metrics not yet implemented",
		"current_snapshot": snapshot,
		"requested_hours":  hours,
		"timestamp":        time.Now(),
		"status":           "success",
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode metrics history response", zap.Error(err))
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.logger.Info("Metrics history requested",
		zap.String("request_id", h.getRequestID(ctx)),
		zap.Int("requested_hours", hours),
	)
}
