package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

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

