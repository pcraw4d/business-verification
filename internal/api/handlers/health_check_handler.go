package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/services"
)

// HealthCheckHandler handles health check API endpoints
type HealthCheckHandler struct {
	logger             *zap.Logger
	healthCheckService *services.HealthCheckService
}

// NewHealthCheckHandler creates a new health check handler
func NewHealthCheckHandler(logger *zap.Logger, healthCheckService *services.HealthCheckService) *HealthCheckHandler {
	return &HealthCheckHandler{
		logger:             logger,
		healthCheckService: healthCheckService,
	}
}

// HealthCheckResponse represents the health check response
type HealthCheckResponse struct {
	Status      string                       `json:"status"`
	Timestamp   time.Time                    `json:"timestamp"`
	Version     string                       `json:"version"`
	Environment string                       `json:"environment"`
	Checks      []services.HealthCheckResult `json:"checks"`
	Summary     HealthCheckSummary           `json:"summary"`
}

// HealthCheckSummary provides a summary of health check results
type HealthCheckSummary struct {
	Total    int `json:"total"`
	Healthy  int `json:"healthy"`
	Warning  int `json:"warning"`
	Critical int `json:"critical"`
}

// GetHealth returns the overall system health
func (h *HealthCheckHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Health check requested")

	start := time.Now()
	checks := h.healthCheckService.GetAllHealthChecks()
	duration := time.Since(start)

	// Calculate summary
	summary := HealthCheckSummary{
		Total: len(checks),
	}

	for _, check := range checks {
		switch check.Status {
		case services.HealthStatusHealthy:
			summary.Healthy++
		case services.HealthStatusWarning:
			summary.Warning++
		case services.HealthStatusCritical:
			summary.Critical++
		}
	}

	// Determine overall status
	overallStatus := "healthy"
	if summary.Critical > 0 {
		overallStatus = "critical"
	} else if summary.Warning > 0 {
		overallStatus = "warning"
	}

	response := HealthCheckResponse{
		Status:      overallStatus,
		Timestamp:   time.Now(),
		Version:     "1.0.0",       // In real implementation, get from build info
		Environment: "development", // In real implementation, get from config
		Checks:      checks,
		Summary:     summary,
	}

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if summary.Critical > 0 {
		statusCode = http.StatusServiceUnavailable
	} else if summary.Warning > 0 {
		statusCode = http.StatusOK // Still OK, but with warnings
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode health check response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Health check completed",
		zap.String("status", overallStatus),
		zap.Duration("duration", duration),
		zap.Int("total_checks", summary.Total),
		zap.Int("healthy", summary.Healthy),
		zap.Int("warning", summary.Warning),
		zap.Int("critical", summary.Critical))
}

// GetHealthDetailed returns detailed health information
func (h *HealthCheckHandler) GetHealthDetailed(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Detailed health check requested")

	checks := h.healthCheckService.GetAllHealthChecks()

	// Add additional system information
	detailedResponse := map[string]interface{}{
		"timestamp":   time.Now(),
		"version":     "1.0.0",
		"environment": "development",
		"checks":      checks,
		"system": map[string]interface{}{
			"uptime":     time.Since(time.Now()).String(), // In real implementation, get actual uptime
			"goroutines": 0,                               // In real implementation, get actual goroutine count
			"memory": map[string]interface{}{
				"alloc_mb": 0, // In real implementation, get actual memory stats
				"sys_mb":   0,
			},
		},
		"dependencies": map[string]interface{}{
			"database":      h.healthCheckService.CheckDatabaseHealth(),
			"cache":         h.healthCheckService.CheckCacheHealth(),
			"external_apis": h.healthCheckService.CheckExternalAPIsHealth(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(detailedResponse); err != nil {
		h.logger.Error("Failed to encode detailed health check response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetHealthLiveness returns a simple liveness check
func (h *HealthCheckHandler) GetHealthLiveness(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check - just return OK if the service is running
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// GetHealthReadiness returns a readiness check
func (h *HealthCheckHandler) GetHealthReadiness(w http.ResponseWriter, r *http.Request) {
	// Readiness check - verify that the service is ready to accept traffic
	checks := h.healthCheckService.GetAllHealthChecks()

	// Check if critical services are healthy
	ready := true
	for _, check := range checks {
		if check.Status == services.HealthStatusCritical {
			ready = false
			break
		}
	}

	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now(),
		"checks":    len(checks),
	}

	if !ready {
		response["status"] = "not_ready"
	}

	json.NewEncoder(w).Encode(response)
}

// GetHealthStartup returns a startup check
func (h *HealthCheckHandler) GetHealthStartup(w http.ResponseWriter, r *http.Request) {
	// Startup check - verify that the service has finished starting up
	// In a real implementation, you might check if initialization is complete

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "started",
		"timestamp": time.Now(),
		"message":   "Service has completed startup",
	}

	json.NewEncoder(w).Encode(response)
}
