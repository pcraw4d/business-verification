package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/pkg/validators"
)

// EnhancedClassificationHandler handles enhanced classification requests with ML integration, crosswalk mappings, geographic awareness, and industry-specific algorithms
type EnhancedClassificationHandler struct {
	multiIndustryService *classification.MultiIndustryService
	mlClassifier         *classification.MLClassifier
	crosswalkMapper      *classification.CrosswalkMapper
	geographicManager    *classification.GeographicManager
	industryMapper       *classification.IndustryMapper
	feedbackCollector    *classification.FeedbackCollector
	accuracyValidator    *classification.AccuracyValidator
	logger               *observability.Logger
	metrics              *observability.Metrics
}

// NewEnhancedClassificationHandler creates a new enhanced classification handler
func NewEnhancedClassificationHandler(
	multiIndustryService *classification.MultiIndustryService,
	mlClassifier *classification.MLClassifier,
	crosswalkMapper *classification.CrosswalkMapper,
	geographicManager *classification.GeographicManager,
	industryMapper *classification.IndustryMapper,
	feedbackCollector *classification.FeedbackCollector,
	accuracyValidator *classification.AccuracyValidator,
	logger *observability.Logger,
	metrics *observability.Metrics,
) *EnhancedClassificationHandler {
	return &EnhancedClassificationHandler{
		multiIndustryService: multiIndustryService,
		mlClassifier:         mlClassifier,
		crosswalkMapper:      crosswalkMapper,
		geographicManager:    geographicManager,
		industryMapper:       industryMapper,
		feedbackCollector:    feedbackCollector,
		accuracyValidator:    accuracyValidator,
		logger:               logger,
		metrics:              metrics,
	}
}

// EnhancedClassificationRequest represents an enhanced classification request with all new features
type EnhancedClassificationRequest struct {
	// Basic fields (backward compatible)
	BusinessName       string `json:"business_name" validate:"required,min=1,max=200"`
	BusinessType       string `json:"business_type,omitempty" validate:"omitempty,max=100"`
	Industry           string `json:"industry,omitempty" validate:"omitempty,max=100"`
	Description        string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Keywords           string `json:"keywords,omitempty" validate:"omitempty,max=500"`
	RegistrationNumber string `json:"registration_number,omitempty" validate:"omitempty,max=50"`

	// Enhanced fields
	MLModelVersion    string                 `json:"ml_model_version,omitempty" validate:"omitempty,max=50"`
	GeographicRegion  string                 `json:"geographic_region,omitempty" validate:"omitempty,max=100"`
	IndustryType      string                 `json:"industry_type,omitempty" validate:"omitempty,max=50"`
	CrosswalkMappings map[string]interface{} `json:"crosswalk_mappings,omitempty"`
	EnhancedMetadata  map[string]interface{} `json:"enhanced_metadata,omitempty"`
	APIVersion        string                 `json:"api_version,omitempty" validate:"omitempty,max=20"`
}

// EnhancedClassificationResponse represents an enhanced classification response with all new features
type EnhancedClassificationResponse struct {
	// Basic fields (backward compatible)
	Success              bool                                    `json:"success"`
	BusinessID           string                                  `json:"business_id,omitempty"`
	Classifications      []classification.IndustryClassification `json:"classifications"`
	PrimaryIndustry      classification.IndustryClassification   `json:"primary_industry"`
	SecondaryIndustry    *classification.IndustryClassification  `json:"secondary_industry,omitempty"`
	TertiaryIndustry     *classification.IndustryClassification  `json:"tertiary_industry,omitempty"`
	OverallConfidence    float64                                 `json:"overall_confidence"`
	ValidationScore      float64                                 `json:"validation_score,omitempty"`
	ClassificationMethod string                                  `json:"classification_method"`
	ProcessingTime       time.Duration                           `json:"processing_time"`
	Timestamp            time.Time                               `json:"timestamp"`
	Error                string                                  `json:"error,omitempty"`

	// Enhanced fields
	MLModelVersion          string                 `json:"ml_model_version,omitempty"`
	MLConfidenceScore       *float64               `json:"ml_confidence_score,omitempty"`
	CrosswalkMappings       map[string]interface{} `json:"crosswalk_mappings,omitempty"`
	GeographicRegion        string                 `json:"geographic_region,omitempty"`
	RegionConfidenceScore   *float64               `json:"region_confidence_score,omitempty"`
	IndustrySpecificData    map[string]interface{} `json:"industry_specific_data,omitempty"`
	ClassificationAlgorithm string                 `json:"classification_algorithm,omitempty"`
	ValidationRulesApplied  []string               `json:"validation_rules_applied,omitempty"`
	EnhancedMetadata        map[string]interface{} `json:"enhanced_metadata,omitempty"`
	APIVersion              string                 `json:"api_version,omitempty"`
}

// BatchEnhancedClassificationRequest represents a batch enhanced classification request
type BatchEnhancedClassificationRequest struct {
	Businesses []EnhancedClassificationRequest `json:"businesses" validate:"required,min=1,max=100"`
	APIVersion string                          `json:"api_version,omitempty" validate:"omitempty,max=20"`
}

// BatchEnhancedClassificationResponse represents a batch enhanced classification response
type BatchEnhancedClassificationResponse struct {
	Success        bool                             `json:"success"`
	Results        []EnhancedClassificationResponse `json:"results"`
	TotalProcessed int                              `json:"total_processed"`
	SuccessCount   int                              `json:"success_count"`
	ErrorCount     int                              `json:"error_count"`
	ProcessingTime time.Duration                    `json:"processing_time"`
	Timestamp      time.Time                        `json:"timestamp"`
	APIVersion     string                           `json:"api_version,omitempty"`
}

// HandleEnhancedClassification handles enhanced classification requests with all new features
func (h *EnhancedClassificationHandler) HandleEnhancedClassification(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Determine API version from header or query parameter
	apiVersion := h.getAPIVersion(r)

	// Log request start
	h.logger.WithComponent("enhanced_classification_handler").LogBusinessEvent(ctx, "enhanced_classification_request_started", "", map[string]interface{}{
		"method":      r.Method,
		"user_agent":  r.UserAgent(),
		"remote_ip":   r.RemoteAddr,
		"api_version": apiVersion,
	})

	// Parse request
	var req EnhancedClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, "Invalid JSON request", http.StatusBadRequest, err, apiVersion)
		return
	}

	// Set API version if not provided
	if req.APIVersion == "" {
		req.APIVersion = apiVersion
	}

	// Validate request
	if err := validators.Validate(req); err != nil {
		h.handleError(w, "Validation failed", http.StatusBadRequest, err, apiVersion)
		return
	}

	// Perform enhanced classification
	result, err := h.performEnhancedClassification(ctx, &req)
	if err != nil {
		h.handleError(w, "Enhanced classification failed", http.StatusInternalServerError, err, apiVersion)
		return
	}

	// Log successful completion
	h.logger.WithComponent("enhanced_classification_handler").LogBusinessEvent(ctx, "enhanced_classification_request_completed", "", map[string]interface{}{
		"business_name":         req.BusinessName,
		"primary_industry_code": result.PrimaryIndustry.IndustryCode,
		"primary_industry_name": result.PrimaryIndustry.IndustryName,
		"overall_confidence":    result.OverallConfidence,
		"ml_confidence_score":   result.MLConfidenceScore,
		"geographic_region":     result.GeographicRegion,
		"api_version":           apiVersion,
		"processing_time_ms":    time.Since(start).Milliseconds(),
		"http_status":           http.StatusOK,
	})

	// Record metrics
	h.metrics.RecordBusinessClassification("enhanced_classification_api_success", "1")

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-API-Version", apiVersion)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.WithComponent("enhanced_classification_handler").LogBusinessEvent(ctx, "enhanced_classification_response_encoding_error", "", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// HandleBatchEnhancedClassification handles batch enhanced classification requests
func (h *EnhancedClassificationHandler) HandleBatchEnhancedClassification(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Determine API version
	apiVersion := h.getAPIVersion(r)

	// Log request start
	h.logger.WithComponent("enhanced_classification_handler").LogBusinessEvent(ctx, "batch_enhanced_classification_request_started", "", map[string]interface{}{
		"method":      r.Method,
		"user_agent":  r.UserAgent(),
		"remote_ip":   r.RemoteAddr,
		"api_version": apiVersion,
	})

	// Parse request
	var req BatchEnhancedClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, "Invalid JSON request", http.StatusBadRequest, err, apiVersion)
		return
	}

	// Set API version if not provided
	if req.APIVersion == "" {
		req.APIVersion = apiVersion
	}

	// Validate request
	if err := validators.Validate(req); err != nil {
		h.handleError(w, "Validation failed", http.StatusBadRequest, err, apiVersion)
		return
	}

	// Process each business
	var responses []EnhancedClassificationResponse
	successCount := 0
	errorCount := 0

	for _, businessReq := range req.Businesses {
		// Set API version
		businessReq.APIVersion = req.APIVersion

		// Validate individual business request
		if err := validators.Validate(businessReq); err != nil {
			responses = append(responses, EnhancedClassificationResponse{
				Success:    false,
				Error:      "Validation failed: " + err.Error(),
				Timestamp:  time.Now(),
				APIVersion: req.APIVersion,
			})
			errorCount++
			continue
		}

		// Perform enhanced classification
		result, err := h.performEnhancedClassification(ctx, &businessReq)
		if err != nil {
			responses = append(responses, EnhancedClassificationResponse{
				Success:    false,
				Error:      "Enhanced classification failed: " + err.Error(),
				Timestamp:  time.Now(),
				APIVersion: req.APIVersion,
			})
			errorCount++
			continue
		}

		responses = append(responses, *result)
		successCount++
	}

	// Create batch response
	batchResponse := BatchEnhancedClassificationResponse{
		Success:        successCount > 0,
		Results:        responses,
		TotalProcessed: len(req.Businesses),
		SuccessCount:   successCount,
		ErrorCount:     errorCount,
		ProcessingTime: time.Since(start),
		Timestamp:      time.Now(),
		APIVersion:     req.APIVersion,
	}

	// Log completion
	h.logger.WithComponent("enhanced_classification_handler").LogBusinessEvent(ctx, "batch_enhanced_classification_request_completed", "", map[string]interface{}{
		"total_processed":    len(req.Businesses),
		"success_count":      successCount,
		"error_count":        errorCount,
		"api_version":        apiVersion,
		"processing_time_ms": time.Since(start).Milliseconds(),
		"http_status":        http.StatusOK,
	})

	// Record metrics
	h.metrics.RecordBusinessClassification("batch_enhanced_classification_api_success", "1")

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-API-Version", apiVersion)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(batchResponse); err != nil {
		h.logger.WithComponent("enhanced_classification_handler").LogBusinessEvent(ctx, "batch_enhanced_classification_response_encoding_error", "", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// HandleBackwardCompatibleClassification handles backward compatible classification requests
func (h *EnhancedClassificationHandler) HandleBackwardCompatibleClassification(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Log request start
	h.logger.WithComponent("enhanced_classification_handler").LogBusinessEvent(ctx, "backward_compatible_classification_request_started", "", map[string]interface{}{
		"method":     r.Method,
		"user_agent": r.UserAgent(),
		"remote_ip":  r.RemoteAddr,
	})

	// Parse legacy request format
	var legacyReq struct {
		BusinessName       string `json:"business_name" validate:"required,min=1,max=200"`
		BusinessType       string `json:"business_type,omitempty" validate:"omitempty,max=100"`
		Industry           string `json:"industry,omitempty" validate:"omitempty,max=100"`
		Description        string `json:"description,omitempty" validate:"omitempty,max=1000"`
		Keywords           string `json:"keywords,omitempty" validate:"omitempty,max=500"`
		RegistrationNumber string `json:"registration_number,omitempty" validate:"omitempty,max=50"`
	}

	if err := json.NewDecoder(r.Body).Decode(&legacyReq); err != nil {
		h.handleError(w, "Invalid JSON request", http.StatusBadRequest, err, "v1")
		return
	}

	// Validate request
	if err := validators.Validate(legacyReq); err != nil {
		h.handleError(w, "Validation failed", http.StatusBadRequest, err, "v1")
		return
	}

	// Convert to enhanced request format
	enhancedReq := &EnhancedClassificationRequest{
		BusinessName:       legacyReq.BusinessName,
		BusinessType:       legacyReq.BusinessType,
		Industry:           legacyReq.Industry,
		Description:        legacyReq.Description,
		Keywords:           legacyReq.Keywords,
		RegistrationNumber: legacyReq.RegistrationNumber,
		APIVersion:         "v1",
	}

	// Perform enhanced classification
	result, err := h.performEnhancedClassification(ctx, enhancedReq)
	if err != nil {
		h.handleError(w, "Classification failed", http.StatusInternalServerError, err, "v1")
		return
	}

	// Convert to legacy response format for backward compatibility
	legacyResponse := struct {
		Success              bool                                    `json:"success"`
		BusinessID           string                                  `json:"business_id,omitempty"`
		Classifications      []classification.IndustryClassification `json:"classifications"`
		PrimaryIndustry      classification.IndustryClassification   `json:"primary_industry"`
		SecondaryIndustry    *classification.IndustryClassification  `json:"secondary_industry,omitempty"`
		TertiaryIndustry     *classification.IndustryClassification  `json:"tertiary_industry,omitempty"`
		OverallConfidence    float64                                 `json:"overall_confidence"`
		ValidationScore      float64                                 `json:"validation_score,omitempty"`
		ClassificationMethod string                                  `json:"classification_method"`
		ProcessingTime       time.Duration                           `json:"processing_time"`
		Timestamp            time.Time                               `json:"timestamp"`
		Error                string                                  `json:"error,omitempty"`
	}{
		Success:              result.Success,
		BusinessID:           result.BusinessID,
		Classifications:      result.Classifications,
		PrimaryIndustry:      result.PrimaryIndustry,
		SecondaryIndustry:    result.SecondaryIndustry,
		TertiaryIndustry:     result.TertiaryIndustry,
		OverallConfidence:    result.OverallConfidence,
		ValidationScore:      result.ValidationScore,
		ClassificationMethod: result.ClassificationMethod,
		ProcessingTime:       result.ProcessingTime,
		Timestamp:            result.Timestamp,
		Error:                result.Error,
	}

	// Log successful completion
	h.logger.WithComponent("enhanced_classification_handler").LogBusinessEvent(ctx, "backward_compatible_classification_request_completed", "", map[string]interface{}{
		"business_name":         legacyReq.BusinessName,
		"primary_industry_code": result.PrimaryIndustry.IndustryCode,
		"primary_industry_name": result.PrimaryIndustry.IndustryName,
		"overall_confidence":    result.OverallConfidence,
		"processing_time_ms":    time.Since(start).Milliseconds(),
		"http_status":           http.StatusOK,
	})

	// Record metrics
	h.metrics.RecordBusinessClassification("backward_compatible_classification_api_success", "1")

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-API-Version", "v1")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(legacyResponse); err != nil {
		h.logger.WithComponent("enhanced_classification_handler").LogBusinessEvent(ctx, "backward_compatible_classification_response_encoding_error", "", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// performEnhancedClassification performs enhanced classification with all new features
func (h *EnhancedClassificationHandler) performEnhancedClassification(ctx context.Context, req *EnhancedClassificationRequest) (*EnhancedClassificationResponse, error) {
	start := time.Now()

	// Convert to internal request format
	internalReq := &classification.ClassificationRequest{
		BusinessName:       req.BusinessName,
		BusinessType:       req.BusinessType,
		Industry:           req.Industry,
		Description:        req.Description,
		Keywords:           req.Keywords,
		RegistrationNumber: req.RegistrationNumber,
	}

	// Perform base multi-industry classification
	result, err := h.multiIndustryService.ClassifyBusinessMultiIndustry(ctx, internalReq)
	if err != nil {
		return nil, fmt.Errorf("base classification failed: %w", err)
	}

	// Initialize enhanced response
	enhancedResult := &EnhancedClassificationResponse{
		Success:              result.Success,
		BusinessID:           result.BusinessID,
		Classifications:      result.Classifications,
		PrimaryIndustry:      result.PrimaryIndustry,
		SecondaryIndustry:    result.SecondaryIndustry,
		TertiaryIndustry:     result.TertiaryIndustry,
		OverallConfidence:    result.OverallConfidence,
		ValidationScore:      result.ValidationScore,
		ClassificationMethod: result.ClassificationMethod,
		ProcessingTime:       result.ProcessingTime,
		Timestamp:            result.Timestamp,
		APIVersion:           req.APIVersion,
	}

	// Apply ML classification if requested
	if req.MLModelVersion != "" && h.mlClassifier != nil {
		mlResult, err := h.mlClassifier.ClassifyWithML(ctx, &classification.MLClassificationRequest{
			BusinessName: req.BusinessName,
			Description:  req.Description,
			Keywords:     req.Keywords,
			ModelVersion: req.MLModelVersion,
			Metadata:     req.EnhancedMetadata,
		})
		if err == nil && mlResult != nil {
			enhancedResult.MLModelVersion = mlResult.ModelVersion
			enhancedResult.MLConfidenceScore = &mlResult.ConfidenceScore
			enhancedResult.ClassificationAlgorithm = "ml_enhanced"
		}
	}

	// Apply crosswalk mappings if available
	if h.crosswalkMapper != nil {
		crosswalkResult, err := h.crosswalkMapper.MapCodes(ctx, result.PrimaryIndustry.IndustryCode, "naics")
		if err == nil && crosswalkResult != nil {
			enhancedResult.CrosswalkMappings = map[string]interface{}{
				"naics_to_sic": crosswalkResult.NAICSToSIC,
				"naics_to_mcc": crosswalkResult.NAICSToMCC,
				"confidence":   crosswalkResult.OverallConfidence,
			}
		}
	}

	// Apply geographic region detection if requested
	if req.GeographicRegion != "" && h.geographicManager != nil {
		regionResult, err := h.geographicManager.DetectRegion(ctx, req.GeographicRegion)
		if err == nil && regionResult != nil {
			enhancedResult.GeographicRegion = regionResult.RegionName
			enhancedResult.RegionConfidenceScore = &regionResult.ConfidenceScore
		}
	}

	// Apply industry-specific mapping if requested
	if req.IndustryType != "" && h.industryMapper != nil {
		industryResult, err := h.industryMapper.ClassifyIndustry(ctx, req.IndustryType, req.BusinessName, req.Description, req.Keywords)
		if err == nil && industryResult != nil {
			enhancedResult.IndustrySpecificData = map[string]interface{}{
				"industry_type":            industryResult.IndustryType,
				"classification_algorithm": industryResult.ClassificationAlgorithm,
				"confidence_score":         industryResult.ConfidenceScore,
				"validation_rules_passed":  industryResult.ValidationRulesPassed,
			}
			enhancedResult.ClassificationAlgorithm = industryResult.ClassificationAlgorithm
		}
	}

	// Add enhanced metadata
	if req.EnhancedMetadata != nil {
		enhancedResult.EnhancedMetadata = req.EnhancedMetadata
	}

	// Update processing time
	enhancedResult.ProcessingTime = time.Since(start)

	return enhancedResult, nil
}

// getAPIVersion determines the API version from request headers or query parameters
func (h *EnhancedClassificationHandler) getAPIVersion(r *http.Request) string {
	// Check header first
	if version := r.Header.Get("X-API-Version"); version != "" {
		return version
	}

	// Check Accept header
	if accept := r.Header.Get("Accept"); accept != "" {
		if strings.Contains(accept, "application/vnd.api+json") {
			// Parse version from Accept header
			if strings.Contains(accept, "version=") {
				parts := strings.Split(accept, "version=")
				if len(parts) > 1 {
					version := strings.Split(parts[1], ";")[0]
					return strings.Trim(version, `"`)
				}
			}
		}
	}

	// Check query parameter
	if version := r.URL.Query().Get("version"); version != "" {
		return version
	}

	// Default to latest version
	return "v2"
}

// handleError handles error responses with API versioning
func (h *EnhancedClassificationHandler) handleError(w http.ResponseWriter, message string, statusCode int, err error, apiVersion string) {
	ctx := context.Background()

	errorResponse := struct {
		Success    bool      `json:"success"`
		Error      string    `json:"error"`
		Timestamp  time.Time `json:"timestamp"`
		APIVersion string    `json:"api_version,omitempty"`
	}{
		Success:    false,
		Error:      message,
		Timestamp:  time.Now(),
		APIVersion: apiVersion,
	}

	// Log error
	h.logger.WithComponent("enhanced_classification_handler").LogBusinessEvent(ctx, "enhanced_classification_request_error", "", map[string]interface{}{
		"error":       err.Error(),
		"status_code": statusCode,
		"message":     message,
		"api_version": apiVersion,
	})

	// Record metrics
	h.metrics.RecordBusinessClassification("enhanced_classification_api_error", "1")

	// Send error response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-API-Version", apiVersion)
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		h.logger.WithComponent("enhanced_classification_handler").LogBusinessEvent(ctx, "enhanced_classification_error_response_encoding_error", "", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// HandleAPIVersionInfo returns information about available API versions
func (h *EnhancedClassificationHandler) HandleAPIVersionInfo(w http.ResponseWriter, r *http.Request) {
	versionInfo := struct {
		CurrentVersion    string            `json:"current_version"`
		SupportedVersions []string          `json:"supported_versions"`
		VersionDetails    map[string]string `json:"version_details"`
		DeprecationInfo   map[string]string `json:"deprecation_info,omitempty"`
	}{
		CurrentVersion:    "v2",
		SupportedVersions: []string{"v1", "v2"},
		VersionDetails: map[string]string{
			"v1": "Legacy API with basic classification features",
			"v2": "Enhanced API with ML integration, crosswalk mappings, geographic awareness, and industry-specific algorithms",
		},
		DeprecationInfo: map[string]string{
			"v1": "v1 will be deprecated on 2024-12-31. Please migrate to v2 for enhanced features.",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(versionInfo); err != nil {
		h.logger.WithComponent("enhanced_classification_handler").LogBusinessEvent(r.Context(), "api_version_info_encoding_error", "", map[string]interface{}{
			"error": err.Error(),
		})
	}
}
