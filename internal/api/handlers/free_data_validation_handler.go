package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"kyb-platform/internal/integrations"
)

// FreeDataValidationHandler handles free data validation requests
type FreeDataValidationHandler struct {
	validationService integrations.FreeDataValidationServiceInterface
	logger            *zap.Logger
}

// NewFreeDataValidationHandler creates a new free data validation handler
func NewFreeDataValidationHandler(validationService integrations.FreeDataValidationServiceInterface, logger *zap.Logger) *FreeDataValidationHandler {
	return &FreeDataValidationHandler{
		validationService: validationService,
		logger:            logger,
	}
}

// ValidateBusinessDataRequest represents a request to validate business data
type ValidateBusinessDataRequest struct {
	BusinessID         string `json:"business_id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	Address            string `json:"address"`
	Phone              string `json:"phone"`
	Email              string `json:"email"`
	Website            string `json:"website"`
	Industry           string `json:"industry"`
	Country            string `json:"country"`
	RegistrationNumber string `json:"registration_number"`
	TaxID              string `json:"tax_id"`
}

// ValidateBusinessDataResponse represents the response from business data validation
type ValidateBusinessDataResponse struct {
	Success               bool                       `json:"success"`
	BusinessID            string                     `json:"business_id"`
	IsValid               bool                       `json:"is_valid"`
	QualityScore          float64                    `json:"quality_score"`
	ConsistencyScore      float64                    `json:"consistency_score"`
	CompletenessScore     float64                    `json:"completeness_score"`
	AccuracyScore         float64                    `json:"accuracy_score"`
	FreshnessScore        float64                    `json:"freshness_score"`
	ValidationErrors      []HandlerValidationError   `json:"validation_errors"`
	ValidationWarnings    []HandlerValidationWarning `json:"validation_warnings"`
	DataSources           []DataSourceInfo           `json:"data_sources"`
	CrossReferenceResults map[string]interface{}     `json:"cross_reference_results"`
	ValidatedAt           time.Time                  `json:"validated_at"`
	ValidationTime        time.Duration              `json:"validation_time"`
	Cost                  float64                    `json:"cost"`
	Message               string                     `json:"message,omitempty"`
}

// HandlerValidationError represents a validation error for the handler
type HandlerValidationError struct {
	Field    string `json:"field"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	Source   string `json:"source"`
	Code     string `json:"code"`
}

// HandlerValidationWarning represents a validation warning for the handler
type HandlerValidationWarning struct {
	Field    string `json:"field"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	Source   string `json:"source"`
	Code     string `json:"code"`
}

// DataSourceInfo represents information about a data source
type DataSourceInfo struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	IsFree      bool      `json:"is_free"`
	Cost        float64   `json:"cost"`
	LastUpdated time.Time `json:"last_updated"`
	Reliability float64   `json:"reliability"`
}

// ValidationStatsResponse represents validation statistics
type ValidationStatsResponse struct {
	Success             bool    `json:"success"`
	TotalValidations    int     `json:"total_validations"`
	ValidCount          int     `json:"valid_count"`
	InvalidCount        int     `json:"invalid_count"`
	AverageQualityScore float64 `json:"average_quality_score"`
	CacheSize           int     `json:"cache_size"`
	CostPerValidation   float64 `json:"cost_per_validation"`
	Message             string  `json:"message,omitempty"`
}

// ValidateBusinessData validates business data using free government APIs
func (h *FreeDataValidationHandler) ValidateBusinessData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidateBusinessDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.BusinessID == "" {
		http.Error(w, "business_id is required", http.StatusBadRequest)
		return
	}

	// Convert request to service data structure
	data := integrations.BusinessDataForValidation{
		BusinessID:         req.BusinessID,
		Name:               req.Name,
		Description:        req.Description,
		Address:            req.Address,
		Phone:              req.Phone,
		Email:              req.Email,
		Website:            req.Website,
		Industry:           req.Industry,
		Country:            req.Country,
		RegistrationNumber: req.RegistrationNumber,
		TaxID:              req.TaxID,
	}

	// Perform validation
	result, err := h.validationService.ValidateBusinessData(r.Context(), data)
	if err != nil {
		h.logger.Error("Validation failed",
			zap.String("business_id", req.BusinessID),
			zap.Error(err))

		response := ValidateBusinessDataResponse{
			Success:    false,
			BusinessID: req.BusinessID,
			Message:    fmt.Sprintf("Validation failed: %v", err),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Convert result to response format
	response := ValidateBusinessDataResponse{
		Success:               true,
		BusinessID:            result.BusinessID,
		IsValid:               result.IsValid,
		QualityScore:          result.QualityScore,
		ConsistencyScore:      result.ConsistencyScore,
		CompletenessScore:     result.CompletenessScore,
		AccuracyScore:         result.AccuracyScore,
		FreshnessScore:        result.FreshnessScore,
		ValidationErrors:      convertValidationErrors(result.ValidationErrors),
		ValidationWarnings:    convertValidationWarnings(result.ValidationWarnings),
		DataSources:           convertDataSources(result.DataSources),
		CrossReferenceResults: result.CrossReferenceResults,
		ValidatedAt:           result.ValidatedAt,
		ValidationTime:        result.ValidationTime,
		Cost:                  result.Cost,
		Message:               "Validation completed successfully",
	}

	h.logger.Info("Business data validation completed",
		zap.String("business_id", req.BusinessID),
		zap.Bool("is_valid", result.IsValid),
		zap.Float64("quality_score", result.QualityScore),
		zap.Duration("validation_time", result.ValidationTime),
		zap.Float64("cost", result.Cost))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetValidationStats returns validation statistics
func (h *FreeDataValidationHandler) GetValidationStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.validationService.GetValidationStats()

	response := ValidationStatsResponse{
		Success:             true,
		TotalValidations:    stats["total_validations"].(int),
		ValidCount:          stats["valid_count"].(int),
		InvalidCount:        stats["invalid_count"].(int),
		AverageQualityScore: stats["average_quality_score"].(float64),
		CacheSize:           stats["cache_size"].(int),
		CostPerValidation:   stats["cost_per_validation"].(float64),
		Message:             "Statistics retrieved successfully",
	}

	h.logger.Info("Validation statistics retrieved",
		zap.Int("total_validations", response.TotalValidations),
		zap.Int("valid_count", response.ValidCount),
		zap.Float64("average_quality_score", response.AverageQualityScore),
		zap.Float64("cost_per_validation", response.CostPerValidation))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HealthCheck performs a health check on the validation service
func (h *FreeDataValidationHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Perform a simple validation to test the service
	testData := integrations.BusinessDataForValidation{
		BusinessID:  "health-check",
		Name:        "Health Check Company",
		Description: "A company for health checking",
		Address:     "123 Health St, Health City, HC 12345",
		Country:     "US",
	}

	startTime := time.Now()
	_, err := h.validationService.ValidateBusinessData(r.Context(), testData)
	validationTime := time.Since(startTime)

	healthStatus := map[string]interface{}{
		"status":          "healthy",
		"service":         "free_data_validation",
		"timestamp":       time.Now(),
		"validation_time": validationTime,
		"error":           nil,
	}

	if err != nil {
		healthStatus["status"] = "unhealthy"
		healthStatus["error"] = err.Error()
		h.logger.Error("Health check failed", zap.Error(err))
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		h.logger.Debug("Health check passed", zap.Duration("validation_time", validationTime))
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(healthStatus)
}

// GetValidationConfig returns the current validation configuration
func (h *FreeDataValidationHandler) GetValidationConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config := integrations.GetDefaultFreeDataValidationConfig()

	response := map[string]interface{}{
		"success": true,
		"config": map[string]interface{}{
			"min_quality_score":            config.MinQualityScore,
			"consistency_weight":           config.ConsistencyWeight,
			"completeness_weight":          config.CompletenessWeight,
			"accuracy_weight":              config.AccuracyWeight,
			"freshness_weight":             config.FreshnessWeight,
			"enable_cross_reference":       config.EnableCrossReference,
			"max_validation_time":          config.MaxValidationTime,
			"cache_validation_results":     config.CacheValidationResults,
			"max_api_calls_per_validation": config.MaxAPICallsPerValidation,
			"rate_limit_delay":             config.RateLimitDelay,
		},
		"message": "Configuration retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// BatchValidateBusinessData validates multiple business data entries
func (h *FreeDataValidationHandler) BatchValidateBusinessData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Businesses    []ValidateBusinessDataRequest `json:"businesses"`
		MaxConcurrent int                           `json:"max_concurrent,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode batch request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Businesses) == 0 {
		http.Error(w, "No businesses provided", http.StatusBadRequest)
		return
	}

	// Set default max concurrent if not provided
	maxConcurrent := req.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 5 // Default to 5 concurrent validations
	}

	// Limit max concurrent to prevent overwhelming the system
	if maxConcurrent > 10 {
		maxConcurrent = 10
	}

	// Process validations
	results := make([]ValidateBusinessDataResponse, len(req.Businesses))
	semaphore := make(chan struct{}, maxConcurrent)

	for i, business := range req.Businesses {
		semaphore <- struct{}{} // Acquire semaphore

		go func(index int, biz ValidateBusinessDataRequest) {
			defer func() { <-semaphore }() // Release semaphore

			data := integrations.BusinessDataForValidation{
				BusinessID:         biz.BusinessID,
				Name:               biz.Name,
				Description:        biz.Description,
				Address:            biz.Address,
				Phone:              biz.Phone,
				Email:              biz.Email,
				Website:            biz.Website,
				Industry:           biz.Industry,
				Country:            biz.Country,
				RegistrationNumber: biz.RegistrationNumber,
				TaxID:              biz.TaxID,
			}

			result, err := h.validationService.ValidateBusinessData(r.Context(), data)
			if err != nil {
				results[index] = ValidateBusinessDataResponse{
					Success:    false,
					BusinessID: biz.BusinessID,
					Message:    fmt.Sprintf("Validation failed: %v", err),
				}
				return
			}

			results[index] = ValidateBusinessDataResponse{
				Success:               true,
				BusinessID:            result.BusinessID,
				IsValid:               result.IsValid,
				QualityScore:          result.QualityScore,
				ConsistencyScore:      result.ConsistencyScore,
				CompletenessScore:     result.CompletenessScore,
				AccuracyScore:         result.AccuracyScore,
				FreshnessScore:        result.FreshnessScore,
				ValidationErrors:      convertValidationErrors(result.ValidationErrors),
				ValidationWarnings:    convertValidationWarnings(result.ValidationWarnings),
				DataSources:           convertDataSources(result.DataSources),
				CrossReferenceResults: result.CrossReferenceResults,
				ValidatedAt:           result.ValidatedAt,
				ValidationTime:        result.ValidationTime,
				Cost:                  result.Cost,
				Message:               "Validation completed successfully",
			}
		}(i, business)
	}

	// Wait for all validations to complete
	for i := 0; i < maxConcurrent; i++ {
		semaphore <- struct{}{}
	}

	// Calculate batch statistics
	totalValidations := len(results)
	validCount := 0
	totalQualityScore := 0.0
	totalCost := 0.0

	for _, result := range results {
		if result.Success && result.IsValid {
			validCount++
		}
		if result.Success {
			totalQualityScore += result.QualityScore
			totalCost += result.Cost
		}
	}

	avgQualityScore := 0.0
	if totalValidations > 0 {
		avgQualityScore = totalQualityScore / float64(totalValidations)
	}

	response := map[string]interface{}{
		"success":               true,
		"total_validations":     totalValidations,
		"valid_count":           validCount,
		"invalid_count":         totalValidations - validCount,
		"average_quality_score": avgQualityScore,
		"total_cost":            totalCost,
		"results":               results,
		"message":               "Batch validation completed successfully",
	}

	h.logger.Info("Batch validation completed",
		zap.Int("total_validations", totalValidations),
		zap.Int("valid_count", validCount),
		zap.Float64("average_quality_score", avgQualityScore),
		zap.Float64("total_cost", totalCost))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper functions to convert between service and handler types

func convertValidationErrors(errors []integrations.ValidationError) []HandlerValidationError {
	result := make([]HandlerValidationError, len(errors))
	for i, err := range errors {
		result[i] = HandlerValidationError{
			Field:    err.Field,
			Message:  err.Message,
			Severity: err.Severity,
			Source:   err.Source,
			Code:     err.Code,
		}
	}
	return result
}

func convertValidationWarnings(warnings []integrations.ValidationWarning) []HandlerValidationWarning {
	result := make([]HandlerValidationWarning, len(warnings))
	for i, warning := range warnings {
		result[i] = HandlerValidationWarning{
			Field:    warning.Field,
			Message:  warning.Message,
			Severity: warning.Severity,
			Source:   warning.Source,
			Code:     warning.Code,
		}
	}
	return result
}

func convertDataSources(sources []integrations.DataSourceInfo) []DataSourceInfo {
	result := make([]DataSourceInfo, len(sources))
	for i, source := range sources {
		result[i] = DataSourceInfo{
			Name:        source.Name,
			Type:        source.Type,
			IsFree:      source.IsFree,
			Cost:        source.Cost,
			LastUpdated: source.LastUpdated,
			Reliability: source.Reliability,
		}
	}
	return result
}

// RegisterRoutes registers the free data validation routes
func (h *FreeDataValidationHandler) RegisterRoutes(mux *http.ServeMux) {
	// Individual validation
	mux.HandleFunc("/api/v3/validate/business-data", h.ValidateBusinessData)

	// Batch validation
	mux.HandleFunc("/api/v3/validate/business-data/batch", h.BatchValidateBusinessData)

	// Statistics and monitoring
	mux.HandleFunc("/api/v3/validate/stats", h.GetValidationStats)
	mux.HandleFunc("/api/v3/validate/config", h.GetValidationConfig)
	mux.HandleFunc("/api/v3/validate/health", h.HealthCheck)
}
