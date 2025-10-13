package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/monitoring"
)

// PerformanceHandler handles performance-related HTTP requests
type PerformanceHandler struct {
	monitor *monitoring.PerformanceMonitor
	logger  *zap.Logger
}

// NewPerformanceHandler creates a new performance handler
func NewPerformanceHandler(monitor *monitoring.PerformanceMonitor, logger *zap.Logger) *PerformanceHandler {
	return &PerformanceHandler{
		monitor: monitor,
		logger:  logger,
	}
}

// GetMetrics returns current performance metrics
func (h *PerformanceHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.monitor.GetStats()

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
	isHealthy := h.monitor.IsHealthy()
	
	health := map[string]interface{}{
		"healthy": isHealthy,
		"timestamp": time.Now(),
	}

	statusCode := http.StatusOK
	if !isHealthy {
		statusCode = http.StatusServiceUnavailable
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
	stats := h.monitor.GetStats()

	systemInfo := map[string]interface{}{
		"timestamp":    time.Now(),
		"stats":        stats,
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
	h.monitor.Reset()

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

// HandlePerformanceStats handles GET /api/v1/performance/stats
func (h *PerformanceHandler) HandlePerformanceStats(w http.ResponseWriter, r *http.Request) {
	h.GetMetrics(w, r)
}

// HandlePerformanceAlerts handles GET /api/v1/performance/alerts
func (h *PerformanceHandler) HandlePerformanceAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := h.monitor.GetAlerts()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(alerts); err != nil {
		h.logger.Error("Failed to encode alerts", zap.Error(err))
		http.Error(w, "Failed to encode alerts", http.StatusInternalServerError)
		return
	}
}

// HandlePerformanceHealth handles GET /api/v1/performance/health
func (h *PerformanceHandler) HandlePerformanceHealth(w http.ResponseWriter, r *http.Request) {
	h.GetHealth(w, r)
}

// HandlePerformanceReset handles POST /api/v1/performance/reset
func (h *PerformanceHandler) HandlePerformanceReset(w http.ResponseWriter, r *http.Request) {
	h.ResetMetrics(w, r)
}

// HandlePerformanceTargets handles POST /api/v1/performance/targets
func (h *PerformanceHandler) HandlePerformanceTargets(w http.ResponseWriter, r *http.Request) {
	var targets struct {
		RPS       float64 `json:"rps"`
		Latency   string  `json:"latency"`
		ErrorRate float64 `json:"error_rate"`
		Throughput float64 `json:"throughput"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&targets); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	latency, err := time.ParseDuration(targets.Latency)
	if err != nil {
		http.Error(w, "Invalid latency format", http.StatusBadRequest)
		return
	}
	
	h.monitor.SetTargets(targets.RPS, latency, targets.ErrorRate, targets.Throughput)
	
	response := map[string]interface{}{
		"message":   "Performance targets set successfully",
		"timestamp": time.Now(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandlePerformanceClearAlerts handles POST /api/v1/performance/alerts/clear
func (h *PerformanceHandler) HandlePerformanceClearAlerts(w http.ResponseWriter, r *http.Request) {
	h.monitor.ClearAlerts()
	
	response := map[string]interface{}{
		"message":   "Alerts cleared successfully",
		"timestamp": time.Now(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
