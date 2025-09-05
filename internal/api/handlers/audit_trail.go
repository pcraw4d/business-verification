package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/external"
)

// AuditTrailHandler handles audit trail API requests
type AuditTrailHandler struct {
	manager *external.AuditTrailManager
	logger  *zap.Logger
}

// NewAuditTrailHandler creates a new audit trail handler
func NewAuditTrailHandler(manager *external.AuditTrailManager, logger *zap.Logger) *AuditTrailHandler {
	return &AuditTrailHandler{
		manager: manager,
		logger:  logger,
	}
}

// CreateHistoryRequest represents a request to create verification history
type CreateHistoryRequest struct {
	VerificationID string                 `json:"verification_id"`
	BusinessName   string                 `json:"business_name"`
	WebsiteURL     string                 `json:"website_url"`
	Events         []external.AuditEvent  `json:"events"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CreateHistoryResponse represents the response from history creation
type CreateHistoryResponse struct {
	Success   bool                          `json:"success"`
	History   *external.VerificationHistory `json:"history,omitempty"`
	Error     string                        `json:"error,omitempty"`
	Timestamp time.Time                     `json:"timestamp"`
}

// QueryAuditTrailRequest represents a request to query audit trail
type QueryAuditTrailRequest struct {
	VerificationID string                `json:"verification_id,omitempty"`
	BusinessName   string                `json:"business_name,omitempty"`
	EventType      string                `json:"event_type,omitempty"`
	Severity       string                `json:"severity,omitempty"`
	UserID         string                `json:"user_id,omitempty"`
	StartTime      *time.Time            `json:"start_time,omitempty"`
	EndTime        *time.Time            `json:"end_time,omitempty"`
	Limit          int                   `json:"limit,omitempty"`
	Offset         int                   `json:"offset,omitempty"`
	Events         []external.AuditEvent `json:"events"`
}

// QueryAuditTrailResponse represents the response from audit trail query
type QueryAuditTrailResponse struct {
	Success   bool                  `json:"success"`
	Events    []external.AuditEvent `json:"events,omitempty"`
	Total     int                   `json:"total"`
	Limit     int                   `json:"limit"`
	Offset    int                   `json:"offset"`
	Error     string                `json:"error,omitempty"`
	Timestamp time.Time             `json:"timestamp"`
}

// GenerateSummaryRequest represents a request to generate history summary
type GenerateSummaryRequest struct {
	History *external.VerificationHistory `json:"history"`
}

// GenerateSummaryResponse represents the response from summary generation
type GenerateSummaryResponse struct {
	Success   bool                     `json:"success"`
	Summary   *external.HistorySummary `json:"summary,omitempty"`
	Error     string                   `json:"error,omitempty"`
	Timestamp time.Time                `json:"timestamp"`
}

// UpdateConfigResponse represents the response from config update
type UpdateConfigResponse struct {
	Success   bool                       `json:"success"`
	Config    *external.AuditTrailConfig `json:"config,omitempty"`
	Error     string                     `json:"error,omitempty"`
	Timestamp time.Time                  `json:"timestamp"`
}

// CreateHistory handles POST /create-history
func (h *AuditTrailHandler) CreateHistory(w http.ResponseWriter, r *http.Request) {
	var req CreateHistoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		response := CreateHistoryResponse{
			Success:   false,
			Error:     "Invalid request body",
			Timestamp: time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CreateHistoryResponse{
		Timestamp: time.Now(),
	}

	// Validate request
	if req.VerificationID == "" {
		response.Success = false
		response.Error = "verification_id is required"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.BusinessName == "" {
		response.Success = false
		response.Error = "business_name is required"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if len(req.Events) == 0 {
		response.Success = false
		response.Error = "events are required"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create verification history
	history, err := h.manager.CreateVerificationHistory(
		r.Context(),
		req.VerificationID,
		req.BusinessName,
		req.WebsiteURL,
		req.Events,
		req.Metadata,
	)

	if err != nil {
		h.logger.Error("failed to create verification history",
			zap.String("verification_id", req.VerificationID),
			zap.Error(err))
		response.Success = false
		response.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		h.logger.Info("verification history created successfully",
			zap.String("verification_id", req.VerificationID),
			zap.String("business_name", req.BusinessName),
			zap.Int("event_count", len(req.Events)))

		response.Success = true
		response.History = history
		w.WriteHeader(http.StatusOK)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// QueryAuditTrail handles POST /query
func (h *AuditTrailHandler) QueryAuditTrail(w http.ResponseWriter, r *http.Request) {
	var req QueryAuditTrailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		response := QueryAuditTrailResponse{
			Success:   false,
			Error:     "Invalid request body",
			Timestamp: time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := QueryAuditTrailResponse{
		Timestamp: time.Now(),
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	// Create query object
	query := external.AuditQuery{
		VerificationID: req.VerificationID,
		BusinessName:   req.BusinessName,
		EventType:      req.EventType,
		Severity:       req.Severity,
		UserID:         req.UserID,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Limit:          req.Limit,
		Offset:         req.Offset,
	}

	// Query audit trail
	events, err := h.manager.QueryAuditTrail(r.Context(), query, req.Events)
	if err != nil {
		h.logger.Error("failed to query audit trail", zap.Error(err))
		response.Success = false
		response.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Success = true
		response.Events = events
		response.Total = len(events)
		w.WriteHeader(http.StatusOK)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// GenerateSummary handles POST /generate-summary
func (h *AuditTrailHandler) GenerateSummary(w http.ResponseWriter, r *http.Request) {
	var req GenerateSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		response := GenerateSummaryResponse{
			Success:   false,
			Error:     "Invalid request body",
			Timestamp: time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := GenerateSummaryResponse{
		Timestamp: time.Now(),
	}

	// Validate request
	if req.History == nil {
		response.Success = false
		response.Error = "history is required"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generate summary
	summary := h.manager.GenerateHistorySummary(req.History)

	h.logger.Info("history summary generated successfully",
		zap.String("verification_id", req.History.VerificationID),
		zap.Int("event_count", summary.EventCount),
		zap.Float64("completion_rate", summary.CompletionRate))

	response.Success = true
	response.Summary = summary
	w.WriteHeader(http.StatusOK)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// GetConfig handles GET /config
func (h *AuditTrailHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	config := h.manager.GetConfig()

	response := UpdateConfigResponse{
		Success:   true,
		Config:    config,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// UpdateConfig handles PUT /config
func (h *AuditTrailHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		response := UpdateConfigResponse{
			Success:   false,
			Error:     "Invalid request body",
			Timestamp: time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := UpdateConfigResponse{
		Timestamp: time.Now(),
	}

	// Convert config map to AuditTrailConfig
	config := &external.AuditTrailConfig{
		// In a real implementation, this would properly map the config values
	}

	// Validate and update configuration
	if err := h.manager.UpdateConfig(config); err != nil {
		h.logger.Error("failed to update config", zap.Error(err))
		response.Success = false
		response.Error = err.Error()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	h.logger.Info("audit trail config updated successfully")

	response.Success = true
	response.Config = h.manager.GetConfig()
	w.WriteHeader(http.StatusOK)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// GetHealth handles GET /health
func (h *AuditTrailHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "audit_trail",
		"timestamp": time.Now(),
		"config":    h.manager.GetConfig(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(health); err != nil {
		h.logger.Error("failed to encode health response", zap.Error(err))
	}
}

// ParseTimeFromQuery parses time from query parameter
func (h *AuditTrailHandler) ParseTimeFromQuery(r *http.Request, param string) (*time.Time, error) {
	timeStr := r.URL.Query().Get(param)
	if timeStr == "" {
		return nil, nil
	}

	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, err
	}

	return &parsedTime, nil
}

// ParseIntFromQuery parses integer from query parameter
func (h *AuditTrailHandler) ParseIntFromQuery(r *http.Request, param string, defaultValue int) int {
	valueStr := r.URL.Query().Get(param)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// RegisterRoutes registers all audit trail routes
func (h *AuditTrailHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/create-history", h.CreateHistory).Methods("POST")
	router.HandleFunc("/query", h.QueryAuditTrail).Methods("POST")
	router.HandleFunc("/generate-summary", h.GenerateSummary).Methods("POST")
	router.HandleFunc("/config", h.GetConfig).Methods("GET")
	router.HandleFunc("/config", h.UpdateConfig).Methods("PUT")
	router.HandleFunc("/health", h.GetHealth).Methods("GET")
}
