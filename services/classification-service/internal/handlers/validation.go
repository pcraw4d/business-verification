package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"kyb-platform/internal/classification/repository"
	"kyb-platform/services/classification-service/internal/errors"
)

// ValidationRequest represents a request to validate a classification
// OPTIMIZATION #5.2: Confidence Calibration - Validation API
type ValidationRequest struct {
	RequestID      string `json:"request_id"`       // Required: The classification request ID
	ActualIndustry string `json:"actual_industry"`  // Required: The actual/correct industry
	ValidatedBy    string `json:"validated_by"`     // Optional: Who validated (e.g., "manual", "api", "automated")
}

// ValidationResponse represents the response to a validation request
type ValidationResponse struct {
	RequestID      string `json:"request_id"`
	Success        bool   `json:"success"`
	IsCorrect      bool   `json:"is_correct"`      // Whether the prediction was correct
	Message        string `json:"message"`
	PredictedIndustry string `json:"predicted_industry,omitempty"`
	ActualIndustry   string `json:"actual_industry,omitempty"`
}

// HandleValidation handles classification validation requests
// OPTIMIZATION #5.2: Confidence Calibration - Validation API
// This endpoint allows updating the actual industry for a classification,
// which is used for accuracy tracking and confidence calibration.
func (h *ClassificationHandler) HandleValidation(w http.ResponseWriter, r *http.Request) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Parse request
	var req ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode validation request", zap.Error(err))
		errors.WriteBadRequest(w, r, "Invalid request body: Please provide valid JSON")
		return
	}

	// Validate request
	if req.RequestID == "" {
		errors.WriteBadRequest(w, r, "request_id is required")
		return
	}

	if req.ActualIndustry == "" {
		errors.WriteBadRequest(w, r, "actual_industry is required")
		return
	}

	// Set default validated_by if not provided
	if req.ValidatedBy == "" {
		req.ValidatedBy = "manual"
	}

	// Sanitize input
	req.RequestID = sanitizeInput(req.RequestID)
	req.ActualIndustry = sanitizeInput(req.ActualIndustry)
	req.ValidatedBy = sanitizeInput(req.ValidatedBy)

	// Update classification accuracy in database
	if h.keywordRepo == nil {
		h.logger.Error("Keyword repository is nil - cannot update validation")
		errors.WriteInternalError(w, r, "Validation service not available")
		return
	}

	// Update the classification accuracy record
	err := h.keywordRepo.UpdateClassificationAccuracy(
		r.Context(),
		req.RequestID,
		req.ActualIndustry,
		req.ValidatedBy,
	)

	if err != nil {
		h.logger.Error("Failed to update classification accuracy",
			zap.String("request_id", req.RequestID),
			zap.String("actual_industry", req.ActualIndustry),
			zap.Error(err))

		// Check if it's a "not found" error
		if err.Error() == "failed to find classification record" {
			errors.WriteNotFound(w, r, "Classification record not found for request_id: "+req.RequestID)
			return
		}

		errors.WriteInternalError(w, r, "Failed to update classification accuracy: "+err.Error())
		return
	}

	// Get the updated record to return is_correct status
	// Note: We could enhance this to return the full record, but for now we'll
	// just return success. The is_correct is calculated in UpdateClassificationAccuracy.
	
	// For now, we'll return a simple success response
	// In the future, we could query the database to get the is_correct value
	response := &ValidationResponse{
		RequestID:      req.RequestID,
		Success:        true,
		Message:        "Classification accuracy updated successfully",
		ActualIndustry: req.ActualIndustry,
	}

	h.logger.Info("Classification accuracy updated",
		zap.String("request_id", req.RequestID),
		zap.String("actual_industry", req.ActualIndustry),
		zap.String("validated_by", req.ValidatedBy))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

