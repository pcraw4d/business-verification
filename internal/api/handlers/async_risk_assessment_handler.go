package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
	"kyb-platform/internal/services"
)

// AsyncRiskAssessmentHandler handles async risk assessment API endpoints
type AsyncRiskAssessmentHandler struct {
	riskService services.RiskAssessmentService
	logger      *log.Logger
}

// NewAsyncRiskAssessmentHandler creates a new async risk assessment handler
func NewAsyncRiskAssessmentHandler(
	riskService services.RiskAssessmentService,
	logger *log.Logger,
) *AsyncRiskAssessmentHandler {
	if logger == nil {
		logger = log.Default()
	}

	return &AsyncRiskAssessmentHandler{
		riskService: riskService,
		logger:      logger,
	}
}

// AssessRisk handles POST /api/v1/risk/assess
func (h *AsyncRiskAssessmentHandler) AssessRisk(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req models.RiskAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Error decoding request: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.MerchantID == "" {
		h.logger.Printf("Missing merchantId in request")
		http.Error(w, "merchant ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Printf("Starting risk assessment for merchant: %s", req.MerchantID)

	// Start assessment (async)
	assessmentID, err := h.riskService.StartAssessment(ctx, req.MerchantID, req.Options)
	if err != nil {
		h.logger.Printf("Error starting assessment: %v", err)
		http.Error(w, "failed to start assessment", http.StatusInternalServerError)
		return
	}

	// Get status to include estimated completion
	status, err := h.riskService.GetAssessmentStatus(ctx, assessmentID)
	if err != nil {
		h.logger.Printf("Error getting assessment status: %v", err)
		// Still return 202 with assessment ID even if status lookup fails
	}

	// Return 202 Accepted with assessment ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	response := models.RiskAssessmentResponse{
		AssessmentID: assessmentID,
		Status:       "pending",
	}

	if status != nil && status.EstimatedCompletion != nil {
		response.EstimatedCompletion = status.EstimatedCompletion
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
	}
}

// extractAssessmentID extracts assessment ID from request path
func (h *AsyncRiskAssessmentHandler) extractAssessmentID(r *http.Request) string {
	// Try Go 1.22+ PathValue first
	if value := r.PathValue("assessmentId"); value != "" {
		return value
	}

	// Fallback: extract from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "assess" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	return ""
}

// extractMerchantID extracts merchant ID from request path
func (h *AsyncRiskAssessmentHandler) extractMerchantID(r *http.Request) string {
	// Try Go 1.22+ PathValue first
	if value := r.PathValue("merchantId"); value != "" {
		return value
	}

	// Fallback: extract from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if (part == "history" || part == "predictions" || part == "merchants") && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	return ""
}

// GetRiskHistory handles GET /api/v1/risk/history/{merchantId}
func (h *AsyncRiskAssessmentHandler) GetRiskHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	merchantID := h.extractMerchantID(r)
	if merchantID == "" {
		h.logger.Printf("Missing merchantId parameter")
		http.Error(w, "merchant ID is required", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	limit := 10
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	h.logger.Printf("Getting risk history for merchant: %s (limit: %d, offset: %d)", merchantID, limit, offset)

	// Call service method
	if h.riskService == nil {
		h.logger.Printf("Risk service not available")
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	history, err := h.riskService.GetRiskHistory(ctx, merchantID, limit, offset)
	if err != nil {
		h.logger.Printf("Error getting risk history: %v", err)
		http.Error(w, "failed to get risk history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"merchantId": merchantID,
		"history":    history,
		"limit":      limit,
		"offset":     offset,
		"total":      len(history),
	}); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
	}
}

// GetRiskPredictions handles GET /api/v1/risk/predictions/{merchantId}
func (h *AsyncRiskAssessmentHandler) GetRiskPredictions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	merchantID := h.extractMerchantID(r)
	if merchantID == "" {
		h.logger.Printf("Missing merchantId parameter")
		http.Error(w, "merchant ID is required", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	horizonsStr := r.URL.Query().Get("horizons")
	includeScenarios := r.URL.Query().Get("includeScenarios") == "true"
	includeConfidence := r.URL.Query().Get("includeConfidence") == "true"

	// Parse horizons (default: 3, 6, 12 months)
	horizons := []int{3, 6, 12}
	if horizonsStr != "" {
		parts := strings.Split(horizonsStr, ",")
		horizons = []int{}
		for _, part := range parts {
			if months, err := strconv.Atoi(strings.TrimSpace(part)); err == nil && months > 0 {
				horizons = append(horizons, months)
			}
		}
		if len(horizons) == 0 {
			horizons = []int{3, 6, 12}
		}
	}

	h.logger.Printf("Getting risk predictions for merchant: %s", merchantID)

	// Call service method
	if h.riskService == nil {
		h.logger.Printf("Risk service not available")
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	predictions, err := h.riskService.GetPredictions(ctx, merchantID, horizons, includeScenarios, includeConfidence)
	if err != nil {
		h.logger.Printf("Error getting risk predictions: %v", err)
		http.Error(w, "failed to get risk predictions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(predictions); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
	}
}

// ExplainRiskAssessment handles GET /api/v1/risk/explain/{assessmentId}
func (h *AsyncRiskAssessmentHandler) ExplainRiskAssessment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	assessmentID := h.extractAssessmentID(r)
	if assessmentID == "" {
		h.logger.Printf("Missing assessmentId parameter")
		http.Error(w, "assessment ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Printf("Getting explanation for assessment: %s", assessmentID)

	// Call service method
	if h.riskService == nil {
		h.logger.Printf("Risk service not available")
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	explanation, err := h.riskService.ExplainAssessment(ctx, assessmentID)
	if err != nil {
		h.logger.Printf("Error getting explanation: %v", err)
		if errors.Is(err, database.ErrAssessmentNotFound) || strings.Contains(err.Error(), "assessment not found") {
			http.Error(w, "assessment not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get explanation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(explanation); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
	}
}

// GetRiskRecommendations handles GET /api/v1/merchants/{merchantId}/risk-recommendations
func (h *AsyncRiskAssessmentHandler) GetRiskRecommendations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	merchantID := h.extractMerchantID(r)
	if merchantID == "" {
		h.logger.Printf("Missing merchantId parameter")
		http.Error(w, "merchant ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Printf("Getting risk recommendations for merchant: %s", merchantID)

	// Call service method
	if h.riskService == nil {
		h.logger.Printf("Risk service not available")
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	recommendations, err := h.riskService.GetRecommendations(ctx, merchantID)
	if err != nil {
		h.logger.Printf("Error getting recommendations: %v", err)
		http.Error(w, "failed to get recommendations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"merchantId":      merchantID,
		"recommendations": recommendations,
		"timestamp":       time.Now(),
	}); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
	}
}

// GetAssessmentStatus handles GET /api/v1/risk/assess/{assessmentId}
func (h *AsyncRiskAssessmentHandler) GetAssessmentStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract assessmentId from path
	assessmentID := h.extractAssessmentID(r)
	if assessmentID == "" {
		h.logger.Printf("Missing assessmentId parameter")
		http.Error(w, "assessment ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Printf("Getting assessment status: %s", assessmentID)

	// Get assessment status
	status, err := h.riskService.GetAssessmentStatus(ctx, assessmentID)
	if err != nil {
		h.logger.Printf("Error getting assessment status: %v", err)
		// Check if error is wrapped ErrAssessmentNotFound
		if errors.Is(err, database.ErrAssessmentNotFound) || strings.Contains(err.Error(), "assessment not found") {
			http.Error(w, "assessment not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get assessment status", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
