package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/performance"
)

// PerformanceHandler handles performance-related HTTP requests
type PerformanceHandler struct {
	monitor *performance.PerformanceMonitor
	logger  *zap.Logger
}

// NewPerformanceHandler creates a new performance handler
func NewPerformanceHandler(monitor *performance.PerformanceMonitor, logger *zap.Logger) *PerformanceHandler {
	return &PerformanceHandler{
		monitor: monitor,
		logger:  logger,
	}
}

// GetMetrics returns current performance metrics
func (h *PerformanceHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.monitor.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		h.logger.Error("Failed to encode metrics", zap.Error(err))
		http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
		return
	}
}

// GetHealth returns system health status
func (h *PerformanceHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	health := h.monitor.GetHealthStatus()

	statusCode := http.StatusOK
	if health.Overall == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if health.Overall == "degraded" {
		statusCode = http.StatusTooManyRequests
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(health); err != nil {
		h.logger.Error("Failed to encode health status", zap.Error(err))
		http.Error(w, "Failed to encode health status", http.StatusInternalServerError)
		return
	}
}

// GetSystemInfo returns system information
func (h *PerformanceHandler) GetSystemInfo(w http.ResponseWriter, r *http.Request) {
	metrics := h.monitor.GetMetrics()

	systemInfo := map[string]interface{}{
		"timestamp":    time.Now(),
		"memory_usage": metrics.MemoryUsage,
		"database":     metrics.DatabaseMetrics,
		"cache":        metrics.CacheMetrics,
		"pool":         metrics.PoolMetrics,
		"queries":      metrics.QueryMetrics,
		"requests":     metrics.RequestMetrics,
		"errors":       metrics.ErrorMetrics,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(systemInfo); err != nil {
		h.logger.Error("Failed to encode system info", zap.Error(err))
		http.Error(w, "Failed to encode system info", http.StatusInternalServerError)
		return
	}
}

// ResetMetrics resets performance metrics
func (h *PerformanceHandler) ResetMetrics(w http.ResponseWriter, r *http.Request) {
	h.monitor.ResetMetrics()

	response := map[string]interface{}{
		"message":   "Metrics reset successfully",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode reset response", zap.Error(err))
		http.Error(w, "Failed to encode reset response", http.StatusInternalServerError)
		return
	}
}
