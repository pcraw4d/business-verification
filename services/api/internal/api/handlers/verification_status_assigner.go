package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"kyb-platform/internal/external"

	"go.uber.org/zap"
)

// VerificationStatusHandler handles verification status assignment API requests
type VerificationStatusHandler struct {
	statusAssigner *external.StatusAssigner
	logger         *zap.Logger
}

// NewVerificationStatusHandler creates a new verification status handler
func NewVerificationStatusHandler(statusAssigner *external.StatusAssigner, logger *zap.Logger) *VerificationStatusHandler {
	return &VerificationStatusHandler{
		statusAssigner: statusAssigner,
		logger:         logger,
	}
}

// AssignStatusRequest represents the request for status assignment
type AssignStatusRequest struct {
	ComparisonResult *external.ComparisonResult   `json:"comparison_result"`
	CustomCriteria   *VerificationCriteriaRequest `json:"custom_criteria,omitempty"`
}

// VerificationCriteriaRequest represents custom verification criteria
type VerificationCriteriaRequest struct {
	PassedThreshold    *float64                           `json:"passed_threshold,omitempty"`
	PartialThreshold   *float64                           `json:"partial_threshold,omitempty"`
	CriticalFields     []string                           `json:"critical_fields,omitempty"`
	MaxDistanceKm      *float64                           `json:"max_distance_km,omitempty"`
	MinConfidenceLevel *string                            `json:"min_confidence_level,omitempty"`
	FieldRequirements  map[string]FieldRequirementRequest `json:"field_requirements,omitempty"`
}

// FieldRequirementRequest represents field requirement configuration
type FieldRequirementRequest struct {
	Required      *bool    `json:"required,omitempty"`
	MinScore      *float64 `json:"min_score,omitempty"`
	MinConfidence *float64 `json:"min_confidence,omitempty"`
	Weight        *float64 `json:"weight,omitempty"`
}

// AssignStatusResponse represents the response from status assignment
type AssignStatusResponse struct {
	Success bool                         `json:"success"`
	Result  *external.VerificationResult `json:"result,omitempty"`
	Error   string                       `json:"error,omitempty"`
}

// BatchAssignStatusRequest represents batch status assignment request
type BatchAssignStatusRequest struct {
	Comparisons []AssignStatusRequest `json:"comparisons"`
}

// BatchAssignStatusResponse represents batch status assignment response
type BatchAssignStatusResponse struct {
	Success bool                           `json:"success"`
	Results []*external.VerificationResult `json:"results,omitempty"`
	Errors  []string                       `json:"errors,omitempty"`
	Summary map[string]interface{}         `json:"summary,omitempty"`
}

// AssignVerificationStatus handles verification status assignment requests
func (h *VerificationStatusHandler) AssignVerificationStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req AssignStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.ComparisonResult == nil {
		h.logger.Error("Missing comparison result in request")
		http.Error(w, "Missing comparison result", http.StatusBadRequest)
		return
	}

	// Create custom status assigner if custom criteria provided
	statusAssigner := h.statusAssigner
	if req.CustomCriteria != nil {
		customCriteria := h.createCustomCriteria(req.CustomCriteria)
		statusAssigner = external.NewStatusAssigner(customCriteria, h.logger)
	}

	// Assign verification status
	result, err := statusAssigner.AssignVerificationStatus(r.Context(), req.ComparisonResult)
	if err != nil {
		h.logger.Error("Status assignment failed", zap.Error(err))
		http.Error(w, "Status assignment failed", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := AssignStatusResponse{
		Success: true,
		Result:  result,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Verification status assignment completed successfully",
		zap.String("verification_id", result.ID),
		zap.String("status", string(result.Status)),
		zap.Float64("overall_score", result.OverallScore))
}

// BatchAssignVerificationStatus handles batch verification status assignment
func (h *VerificationStatusHandler) BatchAssignVerificationStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req BatchAssignStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode batch request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.Comparisons) == 0 {
		h.logger.Error("Empty batch request")
		http.Error(w, "No comparisons provided", http.StatusBadRequest)
		return
	}

	if len(req.Comparisons) > 100 {
		h.logger.Error("Batch size too large", zap.Int("size", len(req.Comparisons)))
		http.Error(w, "Batch size exceeds maximum of 100", http.StatusBadRequest)
		return
	}

	// Process batch
	var results []*external.VerificationResult
	var errors []string

	for i, comparison := range req.Comparisons {
		// Create custom status assigner if custom criteria provided
		statusAssigner := h.statusAssigner
		if comparison.CustomCriteria != nil {
			customCriteria := h.createCustomCriteria(comparison.CustomCriteria)
			statusAssigner = external.NewStatusAssigner(customCriteria, h.logger)
		}

		result, err := statusAssigner.AssignVerificationStatus(r.Context(), comparison.ComparisonResult)
		if err != nil {
			h.logger.Error("Batch status assignment failed",
				zap.Int("index", i),
				zap.Error(err))
			errors = append(errors, "Comparison "+strconv.Itoa(i+1)+": "+err.Error())
			continue
		}

		results = append(results, result)
	}

	// Prepare response
	response := BatchAssignStatusResponse{
		Success: len(errors) == 0,
		Results: results,
		Errors:  errors,
		Summary: map[string]interface{}{
			"total_comparisons": len(req.Comparisons),
			"successful":        len(results),
			"failed":            len(errors),
		},
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode batch response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Batch verification status assignment completed",
		zap.Int("total", len(req.Comparisons)),
		zap.Int("successful", len(results)),
		zap.Int("failed", len(errors)))
}

// GetVerificationCriteria returns the current verification criteria
func (h *VerificationStatusHandler) GetVerificationCriteria(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	criteria := h.statusAssigner.GetCriteria()

	response := map[string]interface{}{
		"success":  true,
		"criteria": criteria,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode criteria response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateVerificationCriteria updates the verification criteria
func (h *VerificationStatusHandler) UpdateVerificationCriteria(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req VerificationCriteriaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode criteria update request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create new criteria
	newCriteria := h.createCustomCriteria(&req)

	// Update criteria
	h.statusAssigner.UpdateCriteria(newCriteria)

	response := map[string]interface{}{
		"success":  true,
		"message":  "Verification criteria updated successfully",
		"criteria": newCriteria,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode criteria update response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Verification criteria updated successfully")
}

// GetVerificationStats returns verification statistics
func (h *VerificationStatusHandler) GetVerificationStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For now, return basic stats
	// In a real implementation, this would query a database or metrics store
	stats := map[string]interface{}{
		"total_verifications": 0,
		"passed_count":        0,
		"partial_count":       0,
		"failed_count":        0,
		"skipped_count":       0,
		"average_score":       0.0,
		"success_rate":        0.0,
	}

	response := map[string]interface{}{
		"success": true,
		"stats":   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode stats response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// createCustomCriteria creates custom verification criteria from request
func (h *VerificationStatusHandler) createCustomCriteria(req *VerificationCriteriaRequest) *external.VerificationCriteria {
	// Start with current criteria
	currentCriteria := h.statusAssigner.GetCriteria()
	criteria := &external.VerificationCriteria{
		PassedThreshold:    currentCriteria.PassedThreshold,
		PartialThreshold:   currentCriteria.PartialThreshold,
		CriticalFields:     currentCriteria.CriticalFields,
		MaxDistanceKm:      currentCriteria.MaxDistanceKm,
		MinConfidenceLevel: currentCriteria.MinConfidenceLevel,
		FieldRequirements:  make(map[string]external.FieldRequirement),
	}

	// Copy current field requirements
	for field, requirement := range currentCriteria.FieldRequirements {
		criteria.FieldRequirements[field] = requirement
	}

	// Update with custom values
	if req.PassedThreshold != nil {
		criteria.PassedThreshold = *req.PassedThreshold
	}

	if req.PartialThreshold != nil {
		criteria.PartialThreshold = *req.PartialThreshold
	}

	if req.CriticalFields != nil {
		criteria.CriticalFields = req.CriticalFields
	}

	if req.MaxDistanceKm != nil {
		criteria.MaxDistanceKm = *req.MaxDistanceKm
	}

	if req.MinConfidenceLevel != nil {
		criteria.MinConfidenceLevel = *req.MinConfidenceLevel
	}

	// Update field requirements
	for field, fieldReq := range req.FieldRequirements {
		currentReq := criteria.FieldRequirements[field]

		if fieldReq.Required != nil {
			currentReq.Required = *fieldReq.Required
		}
		if fieldReq.MinScore != nil {
			currentReq.MinScore = *fieldReq.MinScore
		}
		if fieldReq.MinConfidence != nil {
			currentReq.MinConfidence = *fieldReq.MinConfidence
		}
		if fieldReq.Weight != nil {
			currentReq.Weight = *fieldReq.Weight
		}

		criteria.FieldRequirements[field] = currentReq
	}

	return criteria
}
