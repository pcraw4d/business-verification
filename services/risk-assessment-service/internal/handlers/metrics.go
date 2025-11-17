package handlers

import (
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

// HandleGetMetrics returns overall system metrics
func (h *MetricsHandler) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get overall metrics
	overallMetrics := h.metricsCollector.GetOverallMetrics()

	// Create response
	response := map[string]interface{}{
		"overall_metrics": overallMetrics,
		"timestamp":       time.Now(),
		"status":          "success",
	}

	// Add health status
	snapshot := h.metricsCollector.GetSnapshot()
	response["health_status"] = snapshot.HealthStatus
	response["alerts"] = snapshot.Alerts

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode metrics response", zap.Error(err))
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.logger.Info("Metrics requested",
		zap.String("request_id", h.getRequestID(ctx)),
		zap.String("health_status", snapshot.HealthStatus),
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
