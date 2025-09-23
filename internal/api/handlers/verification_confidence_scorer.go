package handlers

import (
	"encoding/json"
	"net/http"

	"kyb-platform/internal/external"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// ConfidenceScorerHandler handles HTTP requests for confidence scoring
type ConfidenceScorerHandler struct {
	scorer *external.ConfidenceScorer
	logger *zap.Logger
}

// NewConfidenceScorerHandler creates a new confidence scorer handler
func NewConfidenceScorerHandler(scorer *external.ConfidenceScorer, logger *zap.Logger) *ConfidenceScorerHandler {
	return &ConfidenceScorerHandler{
		scorer: scorer,
		logger: logger,
	}
}

// Request/Response types
type CalculateConfidenceRequest struct {
	VerificationResult *external.VerificationResult `json:"verification_result"`
}

type CalculateConfidenceResponse struct {
	ConfidenceScore *external.ConfidenceScore `json:"confidence_score"`
	Success         bool                      `json:"success"`
	Error           string                    `json:"error,omitempty"`
}

type BatchCalculateConfidenceRequest struct {
	VerificationResults []*external.VerificationResult `json:"verification_results"`
}

type BatchCalculateConfidenceResponse struct {
	ConfidenceScores []*external.ConfidenceScore `json:"confidence_scores"`
	Success          bool                        `json:"success"`
	Error            string                      `json:"error,omitempty"`
}

type UpdateCalibrationRequest struct {
	Status string    `json:"status"`
	Scores []float64 `json:"scores"`
}

type UpdateCalibrationResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ConfidenceScorerConfigRequest struct {
	Config external.ConfidenceScorerConfig `json:"config"`
}

type ConfidenceScorerConfigResponse struct {
	Config  external.ConfidenceScorerConfig `json:"config"`
	Success bool                            `json:"success"`
	Error   string                          `json:"error,omitempty"`
}

type ConfidenceScorerStatsResponse struct {
	Statistics map[string]interface{} `json:"statistics"`
	Success    bool                   `json:"success"`
	Error      string                 `json:"error,omitempty"`
}

// CalculateConfidence handles single confidence score calculation
func (h *ConfidenceScorerHandler) CalculateConfidence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CalculateConfidenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.VerificationResult == nil {
		http.Error(w, "Verification result is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	confidenceScore, err := h.scorer.CalculateConfidenceScore(ctx, req.VerificationResult)
	if err != nil {
		h.logger.Error("Failed to calculate confidence score", zap.Error(err))
		response := CalculateConfidenceResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CalculateConfidenceResponse{
		ConfidenceScore: confidenceScore,
		Success:         true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// BatchCalculateConfidence handles batch confidence score calculation
func (h *ConfidenceScorerHandler) BatchCalculateConfidence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BatchCalculateConfidenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.VerificationResults) == 0 {
		http.Error(w, "At least one verification result is required", http.StatusBadRequest)
		return
	}

	if len(req.VerificationResults) > 100 {
		http.Error(w, "Maximum 100 verification results allowed per batch", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	var confidenceScores []*external.ConfidenceScore
	var errors []string

	for i, result := range req.VerificationResults {
		confidenceScore, err := h.scorer.CalculateConfidenceScore(ctx, result)
		if err != nil {
			h.logger.Error("Failed to calculate confidence score for result",
				zap.Int("index", i), zap.Error(err))
			errors = append(errors, err.Error())
			continue
		}
		confidenceScores = append(confidenceScores, confidenceScore)
	}

	response := BatchCalculateConfidenceResponse{
		ConfidenceScores: confidenceScores,
		Success:          len(errors) == 0,
	}

	if len(errors) > 0 {
		response.Error = "Some calculations failed: " + errors[0] // Return first error
	}

	w.Header().Set("Content-Type", "application/json")
	if len(errors) > 0 {
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(response)
}

// GetConfidenceScorerConfig returns the current configuration
func (h *ConfidenceScorerHandler) GetConfidenceScorerConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config := h.scorer.GetConfig()
	response := ConfidenceScorerConfigResponse{
		Config:  config,
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateConfidenceScorerConfig updates the configuration
func (h *ConfidenceScorerHandler) UpdateConfidenceScorerConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConfidenceScorerConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.scorer.UpdateConfig(req.Config)
	if err != nil {
		h.logger.Error("Failed to update configuration", zap.Error(err))
		response := ConfidenceScorerConfigResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ConfidenceScorerConfigResponse{
		Config:  req.Config,
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateCalibrationData updates calibration data for a specific status
func (h *ConfidenceScorerHandler) UpdateCalibrationData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateCalibrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	if len(req.Scores) == 0 {
		http.Error(w, "At least one score is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err := h.scorer.UpdateCalibrationData(ctx, req.Status, req.Scores)
	if err != nil {
		h.logger.Error("Failed to update calibration data", zap.Error(err))
		response := UpdateCalibrationResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := UpdateCalibrationResponse{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCalibrationData returns calibration data for all statuses
func (h *ConfidenceScorerHandler) GetCalibrationData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	calibrationData := h.scorer.GetCalibrationData()

	response := map[string]interface{}{
		"calibration_data": calibrationData,
		"success":          true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetConfidenceScorerStats returns statistics about the confidence scorer
func (h *ConfidenceScorerHandler) GetConfidenceScorerStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	statistics := h.scorer.GetStatistics()
	response := ConfidenceScorerStatsResponse{
		Statistics: statistics,
		Success:    true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateConfidenceScore validates a confidence score
func (h *ConfidenceScorerHandler) ValidateConfidenceScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var confidenceScore external.ConfidenceScore
	if err := json.NewDecoder(r.Body).Decode(&confidenceScore); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.scorer.ValidateConfidenceScore(&confidenceScore)
	response := map[string]interface{}{
		"valid":   err == nil,
		"success": true,
	}

	if err != nil {
		response["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes registers the confidence scorer routes
func (h *ConfidenceScorerHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/calculate-confidence", h.CalculateConfidence).Methods("POST")
	router.HandleFunc("/calculate-confidence/batch", h.BatchCalculateConfidence).Methods("POST")
	router.HandleFunc("/config", h.GetConfidenceScorerConfig).Methods("GET")
	router.HandleFunc("/config", h.UpdateConfidenceScorerConfig).Methods("PUT")
	router.HandleFunc("/calibration", h.UpdateCalibrationData).Methods("POST")
	router.HandleFunc("/calibration", h.GetCalibrationData).Methods("GET")
	router.HandleFunc("/stats", h.GetConfidenceScorerStats).Methods("GET")
	router.HandleFunc("/validate", h.ValidateConfidenceScore).Methods("POST")
}
