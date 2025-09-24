package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"kyb-platform/internal/disaster_recovery"
	"kyb-platform/internal/observability"
)

// DisasterRecoveryHandler handles disaster recovery HTTP requests
type DisasterRecoveryHandler struct {
	drService *disaster_recovery.DRService
	logger    *observability.Logger
}

// NewDisasterRecoveryHandler creates a new disaster recovery handler
func NewDisasterRecoveryHandler(drService *disaster_recovery.DRService, logger *observability.Logger) *DisasterRecoveryHandler {
	return &DisasterRecoveryHandler{
		drService: drService,
		logger:    logger,
	}
}

// DRStatusResponse represents the disaster recovery status response
type DRStatusResponse struct {
	CurrentRegion       string     `json:"current_region"`
	PrimaryHealth       bool       `json:"primary_health"`
	DRHealth            bool       `json:"dr_health"`
	LastFailover        *time.Time `json:"last_failover,omitempty"`
	LastFailback        *time.Time `json:"last_failback,omitempty"`
	FailoverCount       int        `json:"failover_count"`
	FailbackCount       int        `json:"failback_count"`
	AutoFailoverEnabled bool       `json:"auto_failover_enabled"`
	AutoFailbackEnabled bool       `json:"auto_failback_enabled"`
	Status              string     `json:"status"`
}

// HealthStatusResponse represents the health status response
type HealthStatusResponse struct {
	Primary disaster_recovery.HealthCheck `json:"primary"`
	DR      disaster_recovery.HealthCheck `json:"dr"`
}

// FailoverRequest represents a failover request
type FailoverRequest struct {
	Force bool `json:"force"`
}

// FailoverResponse represents a failover response
type FailoverResponse struct {
	Success    bool          `json:"success"`
	FromRegion string        `json:"from_region"`
	ToRegion   string        `json:"to_region"`
	Duration   time.Duration `json:"duration"`
	Error      string        `json:"error,omitempty"`
	Timestamp  time.Time     `json:"timestamp"`
}

// GetStatus handles GET /v1/disaster-recovery/status
func (h *DisasterRecoveryHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting disaster recovery status", map[string]interface{}{})

	status := h.drService.GetStatus()

	response := DRStatusResponse{
		CurrentRegion:       status.CurrentRegion,
		PrimaryHealth:       status.PrimaryHealth,
		DRHealth:            status.DRHealth,
		LastFailover:        status.LastFailover,
		LastFailback:        status.LastFailback,
		FailoverCount:       status.FailoverCount,
		FailbackCount:       status.FailbackCount,
		AutoFailoverEnabled: status.AutoFailoverEnabled,
		AutoFailbackEnabled: status.AutoFailbackEnabled,
		Status:              status.Status,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode DR status response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("DR status retrieved successfully", map[string]interface{}{})
}

// GetHealthStatus handles GET /v1/disaster-recovery/health
func (h *DisasterRecoveryHandler) GetHealthStatus(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting disaster recovery health status", map[string]interface{}{})

	healthStatus := h.drService.GetHealthStatus(r.Context())

	response := HealthStatusResponse{
		Primary: healthStatus["primary"],
		DR:      healthStatus["dr"],
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode DR health status response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("DR health status retrieved successfully", map[string]interface{}{})
}

// InitiateFailover handles POST /v1/disaster-recovery/failover
func (h *DisasterRecoveryHandler) InitiateFailover(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Parse request body
	var req FailoverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode failover request", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Initiating disaster recovery failover", map[string]interface{}{"force": req.Force})

	// Get current status
	status := h.drService.GetStatus()

	// Perform failover
	err := h.drService.InitiateFailover(r.Context())

	duration := time.Since(start)

	response := FailoverResponse{
		Success:    err == nil,
		FromRegion: status.CurrentRegion,
		ToRegion:   "", // Will be set based on success
		Duration:   duration,
		Timestamp:  time.Now().UTC(),
	}

	if err != nil {
		response.Error = err.Error()
		h.logger.Error("Failover failed", map[string]interface{}{"error": err.Error(), "duration": duration})
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		// Get updated status to determine target region
		updatedStatus := h.drService.GetStatus()
		response.ToRegion = updatedStatus.CurrentRegion
		h.logger.Info("Failover completed successfully", map[string]interface{}{"duration": duration})
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode failover response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// InitiateFailback handles POST /v1/disaster-recovery/failback
func (h *DisasterRecoveryHandler) InitiateFailback(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Parse request body
	var req FailoverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode failback request", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Initiating disaster recovery failback", map[string]interface{}{"force": req.Force})

	// Get current status
	status := h.drService.GetStatus()

	// Perform failback
	err := h.drService.InitiateFailback(r.Context())

	duration := time.Since(start)

	response := FailoverResponse{
		Success:    err == nil,
		FromRegion: status.CurrentRegion,
		ToRegion:   "", // Will be set based on success
		Duration:   duration,
		Timestamp:  time.Now().UTC(),
	}

	if err != nil {
		response.Error = err.Error()
		h.logger.Error("Failback failed", map[string]interface{}{"error": err.Error(), "duration": duration})
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		// Get updated status to determine target region
		updatedStatus := h.drService.GetStatus()
		response.ToRegion = updatedStatus.CurrentRegion
		h.logger.Info("Failback completed successfully", map[string]interface{}{"duration": duration})
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode failback response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// TestFailover handles POST /v1/disaster-recovery/test-failover
func (h *DisasterRecoveryHandler) TestFailover(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Testing disaster recovery failover", map[string]interface{}{})

	err := h.drService.TestFailover(r.Context())

	response := map[string]interface{}{
		"success":   err == nil,
		"timestamp": time.Now().UTC(),
	}

	if err != nil {
		response["error"] = err.Error()
		h.logger.Error("Test failover failed", map[string]interface{}{"error": err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		h.logger.Info("Test failover completed successfully", map[string]interface{}{})
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode test failover response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// TestFailback handles POST /v1/disaster-recovery/test-failback
func (h *DisasterRecoveryHandler) TestFailback(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Testing disaster recovery failback", map[string]interface{}{})

	err := h.drService.TestFailback(r.Context())

	response := map[string]interface{}{
		"success":   err == nil,
		"timestamp": time.Now().UTC(),
	}

	if err != nil {
		response["error"] = err.Error()
		h.logger.Error("Test failback failed", map[string]interface{}{"error": err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		h.logger.Info("Test failback completed successfully", map[string]interface{}{})
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode test failback response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// EnableAutoFailover handles POST /v1/disaster-recovery/auto-failover/enable
func (h *DisasterRecoveryHandler) EnableAutoFailover(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Enabling auto-failover", map[string]interface{}{})

	h.drService.EnableAutoFailover()

	response := map[string]interface{}{
		"message":   "Auto-failover enabled successfully",
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode auto-failover enable response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Auto-failover enabled successfully", map[string]interface{}{})
}

// DisableAutoFailover handles POST /v1/disaster-recovery/auto-failover/disable
func (h *DisasterRecoveryHandler) DisableAutoFailover(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Disabling auto-failover", map[string]interface{}{})

	h.drService.DisableAutoFailover()

	response := map[string]interface{}{
		"message":   "Auto-failover disabled successfully",
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode auto-failover disable response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Auto-failover disabled successfully", map[string]interface{}{})
}

// EnableAutoFailback handles POST /v1/disaster-recovery/auto-failback/enable
func (h *DisasterRecoveryHandler) EnableAutoFailback(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Enabling auto-failback", map[string]interface{}{})

	h.drService.EnableAutoFailback()

	response := map[string]interface{}{
		"message":   "Auto-failback enabled successfully",
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode auto-failback enable response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Auto-failback enabled successfully", map[string]interface{}{})
}

// DisableAutoFailback handles POST /v1/disaster-recovery/auto-failback/disable
func (h *DisasterRecoveryHandler) DisableAutoFailback(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Disabling auto-failback", map[string]interface{}{})

	h.drService.DisableAutoFailback()

	response := map[string]interface{}{
		"message":   "Auto-failback disabled successfully",
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode auto-failback disable response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Auto-failback disabled successfully", map[string]interface{}{})
}

// GetFailoverHistory handles GET /v1/disaster-recovery/history
func (h *DisasterRecoveryHandler) GetFailoverHistory(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting failover history", map[string]interface{}{})

	history := h.drService.GetFailoverHistory()

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(history); err != nil {
		h.logger.Error("Failed to encode failover history response", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Failover history retrieved successfully", map[string]interface{}{})
}
