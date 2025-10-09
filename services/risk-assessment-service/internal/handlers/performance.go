package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/monitoring"
)

// PerformanceHandler handles performance monitoring endpoints
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

// HandlePerformanceStats handles GET /api/v1/performance/stats
func (h *PerformanceHandler) HandlePerformanceStats(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling performance stats request")

	stats := h.monitor.GetStats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.logger.Error("Failed to encode performance stats", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Performance stats returned successfully",
		zap.Int64("request_count", stats.RequestCount),
		zap.Float64("requests_per_second", stats.RequestsPerSecond),
		zap.Float64("error_rate", stats.ErrorRate))
}

// HandlePerformanceAlerts handles GET /api/v1/performance/alerts
func (h *PerformanceHandler) HandlePerformanceAlerts(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling performance alerts request")

	alerts := h.monitor.GetAlerts()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts":    alerts,
		"count":     len(alerts),
		"timestamp": time.Now(),
	}); err != nil {
		h.logger.Error("Failed to encode performance alerts", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Performance alerts returned successfully",
		zap.Int("alert_count", len(alerts)))
}

// HandlePerformanceHealth handles GET /api/v1/performance/health
func (h *PerformanceHandler) HandlePerformanceHealth(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling performance health request")

	isHealthy := h.monitor.IsHealthy()
	stats := h.monitor.GetStats()

	healthStatus := "healthy"
	if !isHealthy {
		healthStatus = "unhealthy"
	}

	response := map[string]interface{}{
		"status":              healthStatus,
		"is_healthy":          isHealthy,
		"requests_per_second": stats.RequestsPerSecond,
		"requests_per_minute": stats.RequestsPerMinute,
		"error_rate":          stats.ErrorRate,
		"max_response_time":   stats.MaxResponseTime.String(),
		"memory_usage_mb":     stats.MemoryUsage / (1024 * 1024),
		"cpu_usage_percent":   stats.CPUUsage,
		"goroutine_count":     stats.GoroutineCount,
		"alert_count":         len(stats.Alerts),
		"timestamp":           time.Now(),
	}

	statusCode := http.StatusOK
	if !isHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode performance health", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Performance health returned successfully",
		zap.String("status", healthStatus),
		zap.Bool("is_healthy", isHealthy))
}

// HandlePerformanceReset handles POST /api/v1/performance/reset
func (h *PerformanceHandler) HandlePerformanceReset(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling performance reset request")

	h.monitor.Reset()

	response := map[string]interface{}{
		"message":   "Performance metrics reset successfully",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode performance reset response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Performance metrics reset successfully")
}

// HandlePerformanceTargets handles POST /api/v1/performance/targets
func (h *PerformanceHandler) HandlePerformanceTargets(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling performance targets request")

	var targets struct {
		RPS        float64 `json:"rps"`
		Latency    string  `json:"latency"`
		ErrorRate  float64 `json:"error_rate"`
		Throughput float64 `json:"throughput"`
	}

	if err := json.NewDecoder(r.Body).Decode(&targets); err != nil {
		h.logger.Error("Failed to decode performance targets", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse latency duration
	latency, err := time.ParseDuration(targets.Latency)
	if err != nil {
		h.logger.Error("Invalid latency format", zap.Error(err))
		http.Error(w, "Invalid latency format", http.StatusBadRequest)
		return
	}

	// Set targets
	h.monitor.SetTargets(targets.RPS, latency, targets.ErrorRate, targets.Throughput)

	response := map[string]interface{}{
		"message":   "Performance targets updated successfully",
		"targets":   targets,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode performance targets response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Performance targets updated successfully",
		zap.Float64("rps", targets.RPS),
		zap.Duration("latency", latency),
		zap.Float64("error_rate", targets.ErrorRate),
		zap.Float64("throughput", targets.Throughput))
}

// HandlePerformanceClearAlerts handles POST /api/v1/performance/alerts/clear
func (h *PerformanceHandler) HandlePerformanceClearAlerts(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling performance clear alerts request")

	h.monitor.ClearAlerts()

	response := map[string]interface{}{
		"message":   "Performance alerts cleared successfully",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode performance clear alerts response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Performance alerts cleared successfully")
}
