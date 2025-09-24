package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/internal/modules/classification_optimization"
)

// ClassificationOptimizationValidationHandler handles HTTP requests for classification optimization accuracy validation
type ClassificationOptimizationValidationHandler struct {
	validator *classification_optimization.AccuracyValidator
	logger    *zap.Logger
}

// NewClassificationOptimizationValidationHandler creates a new classification optimization validation handler
func NewClassificationOptimizationValidationHandler(validator *classification_optimization.AccuracyValidator, logger *zap.Logger) *ClassificationOptimizationValidationHandler {
	return &ClassificationOptimizationValidationHandler{
		validator: validator,
		logger:    logger,
	}
}

// ValidateAccuracyRequest represents a request to validate algorithm accuracy
type ValidateAccuracyRequest struct {
	AlgorithmID string                                        `json:"algorithm_id"`
	TestCases   []*classification_optimization.TestCase       `json:"test_cases"`
	Config      *classification_optimization.ValidationConfig `json:"config,omitempty"`
}

// ValidateAccuracyResponse represents a response for accuracy validation
type ValidateAccuracyResponse struct {
	Success bool                                          `json:"success"`
	Result  *classification_optimization.ValidationResult `json:"result,omitempty"`
	Error   string                                        `json:"error,omitempty"`
}

// CrossValidationRequest represents a request to perform cross validation
type CrossValidationRequest struct {
	AlgorithmID string                                        `json:"algorithm_id"`
	TestCases   []*classification_optimization.TestCase       `json:"test_cases"`
	Config      *classification_optimization.ValidationConfig `json:"config,omitempty"`
}

// CrossValidationResponse represents a response for cross validation
type CrossValidationResponse struct {
	Success bool                                          `json:"success"`
	Result  *classification_optimization.ValidationResult `json:"result,omitempty"`
	Error   string                                        `json:"error,omitempty"`
}

// ValidationHistoryResponse represents a response for validation history
type ValidationHistoryResponse struct {
	Success bool                                            `json:"success"`
	History []*classification_optimization.ValidationResult `json:"history,omitempty"`
	Count   int                                             `json:"count"`
	Error   string                                          `json:"error,omitempty"`
}

// ValidationSummaryResponse represents a response for validation summary
type ValidationSummaryResponse struct {
	Success bool                                           `json:"success"`
	Summary *classification_optimization.ValidationSummary `json:"summary,omitempty"`
	Error   string                                         `json:"error,omitempty"`
}

// ActiveValidationsResponse represents a response for active validations
type ActiveValidationsResponse struct {
	Success bool                                                     `json:"success"`
	Active  map[string]*classification_optimization.ValidationResult `json:"active,omitempty"`
	Count   int                                                      `json:"count"`
	Error   string                                                   `json:"error,omitempty"`
}

// ValidateAccuracy handles accuracy validation requests
func (h *ClassificationOptimizationValidationHandler) ValidateAccuracy(w http.ResponseWriter, r *http.Request) {
	var req ValidateAccuracyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := ValidateAccuracyResponse{
			Success: false,
			Error:   "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate request
	if req.AlgorithmID == "" {
		response := ValidateAccuracyResponse{
			Success: false,
			Error:   "Algorithm ID is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if len(req.TestCases) == 0 {
		response := ValidateAccuracyResponse{
			Success: false,
			Error:   "Test cases are required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Perform validation
	result, err := h.validator.ValidateAccuracy(r.Context(), req.AlgorithmID, req.TestCases)
	if err != nil {
		h.logger.Error("Accuracy validation failed",
			zap.String("algorithm_id", req.AlgorithmID),
			zap.Error(err))

		response := ValidateAccuracyResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ValidateAccuracyResponse{
		Success: true,
		Result:  result,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// PerformCrossValidation handles cross validation requests
func (h *ClassificationOptimizationValidationHandler) PerformCrossValidation(w http.ResponseWriter, r *http.Request) {
	var req CrossValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := CrossValidationResponse{
			Success: false,
			Error:   "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate request
	if req.AlgorithmID == "" {
		response := CrossValidationResponse{
			Success: false,
			Error:   "Algorithm ID is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if len(req.TestCases) == 0 {
		response := CrossValidationResponse{
			Success: false,
			Error:   "Test cases are required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Perform cross validation
	result, err := h.validator.PerformCrossValidation(r.Context(), req.AlgorithmID, req.TestCases)
	if err != nil {
		h.logger.Error("Cross validation failed",
			zap.String("algorithm_id", req.AlgorithmID),
			zap.Error(err))

		response := CrossValidationResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CrossValidationResponse{
		Success: true,
		Result:  result,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetValidationHistory handles validation history requests
func (h *ClassificationOptimizationValidationHandler) GetValidationHistory(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	algorithmID := r.URL.Query().Get("algorithm_id")
	validationType := r.URL.Query().Get("validation_type")
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Parse limit and offset
	limit := 100 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get validation history
	history := h.validator.GetValidationHistory()

	// Apply filters
	var filteredHistory []*classification_optimization.ValidationResult
	for _, validation := range history {
		// Filter by algorithm ID
		if algorithmID != "" && validation.AlgorithmID != algorithmID {
			continue
		}

		// Filter by validation type
		if validationType != "" && string(validation.ValidationType) != validationType {
			continue
		}

		// Filter by status
		if status != "" && string(validation.Status) != status {
			continue
		}

		filteredHistory = append(filteredHistory, validation)
	}

	// Apply pagination
	totalCount := len(filteredHistory)
	end := offset + limit
	if end > totalCount {
		end = totalCount
	}
	if offset >= totalCount {
		filteredHistory = []*classification_optimization.ValidationResult{}
	} else {
		filteredHistory = filteredHistory[offset:end]
	}

	response := ValidationHistoryResponse{
		Success: true,
		History: filteredHistory,
		Count:   len(filteredHistory),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetValidationSummary handles validation summary requests
func (h *ClassificationOptimizationValidationHandler) GetValidationSummary(w http.ResponseWriter, r *http.Request) {
	summary := h.validator.GetValidationSummary()

	response := ValidationSummaryResponse{
		Success: true,
		Summary: summary,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetActiveValidations handles active validations requests
func (h *ClassificationOptimizationValidationHandler) GetActiveValidations(w http.ResponseWriter, r *http.Request) {
	active := h.validator.GetActiveValidations()

	response := ActiveValidationsResponse{
		Success: true,
		Active:  active,
		Count:   len(active),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetValidationByID handles requests for specific validation results
func (h *ClassificationOptimizationValidationHandler) GetValidationByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	validationID := vars["id"]

	if validationID == "" {
		response := ValidateAccuracyResponse{
			Success: false,
			Error:   "Validation ID is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get validation history and find the specific validation
	history := h.validator.GetValidationHistory()
	var targetValidation *classification_optimization.ValidationResult

	for _, validation := range history {
		if validation.ID == validationID {
			targetValidation = validation
			break
		}
	}

	if targetValidation == nil {
		response := ValidateAccuracyResponse{
			Success: false,
			Error:   "Validation not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ValidateAccuracyResponse{
		Success: true,
		Result:  targetValidation,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetValidationsByAlgorithm handles requests for validations by algorithm
func (h *ClassificationOptimizationValidationHandler) GetValidationsByAlgorithm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	algorithmID := vars["algorithm_id"]

	if algorithmID == "" {
		response := ValidationHistoryResponse{
			Success: false,
			Error:   "Algorithm ID is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Parse limit and offset
	limit := 100 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get validation history and filter by algorithm
	history := h.validator.GetValidationHistory()
	var algorithmValidations []*classification_optimization.ValidationResult

	for _, validation := range history {
		if validation.AlgorithmID == algorithmID {
			algorithmValidations = append(algorithmValidations, validation)
		}
	}

	// Apply pagination
	totalCount := len(algorithmValidations)
	end := offset + limit
	if end > totalCount {
		end = totalCount
	}
	if offset >= totalCount {
		algorithmValidations = []*classification_optimization.ValidationResult{}
	} else {
		algorithmValidations = algorithmValidations[offset:end]
	}

	response := ValidationHistoryResponse{
		Success: true,
		History: algorithmValidations,
		Count:   len(algorithmValidations),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetValidationsByType handles requests for validations by type
func (h *ClassificationOptimizationValidationHandler) GetValidationsByType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	validationType := vars["type"]

	if validationType == "" {
		response := ValidationHistoryResponse{
			Success: false,
			Error:   "Validation type is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Parse limit and offset
	limit := 100 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get validation history and filter by type
	history := h.validator.GetValidationHistory()
	var typeValidations []*classification_optimization.ValidationResult

	for _, validation := range history {
		if string(validation.ValidationType) == validationType {
			typeValidations = append(typeValidations, validation)
		}
	}

	// Apply pagination
	totalCount := len(typeValidations)
	end := offset + limit
	if end > totalCount {
		end = totalCount
	}
	if offset >= totalCount {
		typeValidations = []*classification_optimization.ValidationResult{}
	} else {
		typeValidations = typeValidations[offset:end]
	}

	response := ValidationHistoryResponse{
		Success: true,
		History: typeValidations,
		Count:   len(typeValidations),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CancelValidation handles validation cancellation requests
func (h *ClassificationOptimizationValidationHandler) CancelValidation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	validationID := vars["id"]

	if validationID == "" {
		response := ValidateAccuracyResponse{
			Success: false,
			Error:   "Validation ID is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check if validation is active
	active := h.validator.GetActiveValidations()
	validation, exists := active[validationID]

	if !exists {
		response := ValidateAccuracyResponse{
			Success: false,
			Error:   "Validation not found or not active",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// For now, we'll just return the validation status
	// In a real implementation, you would have a method to cancel active validations
	response := ValidateAccuracyResponse{
		Success: true,
		Result:  validation,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HealthCheck handles health check requests
func (h *ClassificationOptimizationValidationHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "classification-optimization-validation",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
