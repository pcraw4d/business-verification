package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/services"
)

// HealthHandler handles health check API endpoints
type HealthHandler struct {
	logger             *zap.Logger
	healthCheckService *services.HealthCheckService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logger *zap.Logger, healthCheckService *services.HealthCheckService) *HealthHandler {
	return &HealthHandler{
		logger:             logger,
		healthCheckService: healthCheckService,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string                       `json:"status"`
	Timestamp   time.Time                    `json:"timestamp"`
	Version     string                       `json:"version"`
	Environment string                       `json:"environment"`
	Checks      []services.HealthCheckResult `json:"checks"`
	Summary     HealthSummary                `json:"summary"`
}

// HealthSummary provides a summary of health check results
type HealthSummary struct {
	Total    int `json:"total"`
	Healthy  int `json:"healthy"`
	Warning  int `json:"warning"`
	Critical int `json:"critical"`
}

// GetHealth returns the overall system health
func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Health check requested")

	start := time.Now()
	checks := h.healthCheckService.GetAllHealthChecks()
	duration := time.Since(start)

	// Calculate summary
	summary := HealthSummary{
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

	response := HealthResponse{
		Status:      overallStatus,
		Timestamp:   time.Now(),
		Version:     "1.0.0",
		Environment: "development",
		Checks:      checks,
		Summary:     summary,
	}

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if summary.Critical > 0 {
		statusCode = http.StatusServiceUnavailable
	} else if summary.Warning > 0 {
		statusCode = http.StatusOK
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
