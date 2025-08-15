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

// ClassificationHandler handles classification requests with enhanced features for beta testing
type ClassificationHandler struct {
	classificationSvc *classification.ClassificationService
	enhancedHandler   *EnhancedClassificationHandler
	logger            *observability.Logger
	metrics           *observability.Metrics
	validator         *validators.Validator
}

// NewClassificationHandler creates a new classification handler
func NewClassificationHandler(
	classificationSvc *classification.ClassificationService,
	enhancedHandler *EnhancedClassificationHandler,
	logger *observability.Logger,
	metrics *observability.Metrics,
	validator *validators.Validator,
) *ClassificationHandler {
	return &ClassificationHandler{
		classificationSvc: classificationSvc,
		enhancedHandler:   enhancedHandler,
		logger:            logger,
		metrics:           metrics,
		validator:         validator,
	}
}

// EnhancedClassificationRequest represents an enhanced classification request with geographic region support
type EnhancedClassificationRequest struct {
	// Basic fields (backward compatible)
	BusinessName       string `json:"business_name" validate:"required,min=1,max=200"`
	BusinessType       string `json:"business_type,omitempty" validate:"omitempty,max=100"`
	Industry           string `json:"industry,omitempty" validate:"omitempty,max=100"`
	Description        string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Keywords           string `json:"keywords,omitempty" validate:"omitempty,max=500"`
	RegistrationNumber string `json:"registration_number,omitempty" validate:"omitempty,max=50"`

	// Enhanced fields for beta testing
	GeographicRegion string                 `json:"geographic_region,omitempty" validate:"omitempty,max=100"`
	IndustryType     string                 `json:"industry_type,omitempty" validate:"omitempty,max=50"`
	EnhancedMetadata map[string]interface{} `json:"enhanced_metadata,omitempty"`
	APIVersion       string                 `json:"api_version,omitempty" validate:"omitempty,max=20"`
}

// EnhancedClassificationResponse represents an enhanced classification response with geographic region information
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

	// Enhanced fields for beta testing
	GeographicRegion      string                 `json:"geographic_region,omitempty"`
	RegionConfidenceScore *float64               `json:"region_confidence_score,omitempty"`
	IndustrySpecificData  map[string]interface{} `json:"industry_specific_data,omitempty"`
	EnhancedMetadata      map[string]interface{} `json:"enhanced_metadata,omitempty"`
	APIVersion            string                 `json:"api_version,omitempty"`
}

// ClassifyBusiness handles single business classification with enhanced features
func (h *ClassificationHandler) ClassifyBusiness(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse request body
	var req EnhancedClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithComponent("api").WithError(err).Error("Failed to parse classification request")
		h.handleError(w, "Invalid JSON in request body", http.StatusBadRequest, err)
		return
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		h.logger.WithComponent("api").WithError(err).Error("Classification request validation failed")
		h.handleError(w, "Invalid request data", http.StatusBadRequest, err)
		return
	}

	// Convert to internal request format
	internalReq := &classification.ClassificationRequest{
		BusinessName:       req.BusinessName,
		BusinessType:       req.BusinessType,
		Industry:           req.Industry,
		Description:        req.Description,
		Keywords:           req.Keywords,
		RegistrationNumber: req.RegistrationNumber,
	}

	// Add geographic region if provided
	if req.GeographicRegion != "" {
		internalReq.GeographicRegion = req.GeographicRegion
	}

	// Add enhanced metadata if provided
	if req.EnhancedMetadata != nil {
		internalReq.EnhancedMetadata = req.EnhancedMetadata
	}

	// Perform classification
	response, err := h.classificationSvc.ClassifyBusiness(ctx, internalReq)
	if err != nil {
		h.logger.WithComponent("api").WithError(err).Error("Classification failed")
		if strings.Contains(strings.ToLower(err.Error()), "invalid request") {
			h.handleError(w, "Invalid classification request", http.StatusBadRequest, err)
		} else {
			h.handleError(w, "Failed to classify business", http.StatusInternalServerError, err)
		}
		return
	}

	// Convert to enhanced response format
	enhancedResponse := &EnhancedClassificationResponse{
		Success:              response.Success,
		BusinessID:           response.BusinessID,
		Classifications:      response.Classifications,
		PrimaryIndustry:      response.PrimaryIndustry,
		SecondaryIndustry:    response.SecondaryIndustry,
		TertiaryIndustry:     response.TertiaryIndustry,
		OverallConfidence:    response.OverallConfidence,
		ValidationScore:      response.ValidationScore,
		ClassificationMethod: response.ClassificationMethod,
		ProcessingTime:       response.ProcessingTime,
		Timestamp:            response.Timestamp,
		Error:                response.Error,
		GeographicRegion:     req.GeographicRegion,
		EnhancedMetadata:     req.EnhancedMetadata,
		APIVersion:           h.getAPIVersion(r),
	}

	// Add region confidence score if geographic region was provided
	if req.GeographicRegion != "" {
		regionConfidence := h.calculateRegionConfidence(req.GeographicRegion, response)
		enhancedResponse.RegionConfidenceScore = &regionConfidence
	}

	// Add industry-specific data if available
	if response.IndustrySpecificData != nil {
		enhancedResponse.IndustrySpecificData = response.IndustrySpecificData
	}

	// Log successful classification
	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))

	// Record metrics
	h.recordClassificationMetrics(ctx, req, response, time.Since(start))

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(enhancedResponse)
}

// ClassifyBusinesses handles batch business classification with enhanced features
func (h *ClassificationHandler) ClassifyBusinesses(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse request body
	var req struct {
		Businesses       []EnhancedClassificationRequest `json:"businesses" validate:"required,min=1,max=100"`
		GeographicRegion string                          `json:"geographic_region,omitempty" validate:"omitempty,max=100"`
		APIVersion       string                          `json:"api_version,omitempty" validate:"omitempty,max=20"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithComponent("api").WithError(err).Error("Failed to parse batch classification request")
		h.handleError(w, "Invalid JSON in request body", http.StatusBadRequest, err)
		return
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		h.logger.WithComponent("api").WithError(err).Error("Batch classification request validation failed")
		h.handleError(w, "Invalid request data", http.StatusBadRequest, err)
		return
	}

	// Convert to internal batch request format
	internalBusinesses := make([]*classification.ClassificationRequest, len(req.Businesses))
	for i, business := range req.Businesses {
		internalReq := &classification.ClassificationRequest{
			BusinessName:       business.BusinessName,
			BusinessType:       business.BusinessType,
			Industry:           business.Industry,
			Description:        business.Description,
			Keywords:           business.Keywords,
			RegistrationNumber: business.RegistrationNumber,
		}

		// Add geographic region (from batch or individual request)
		if business.GeographicRegion != "" {
			internalReq.GeographicRegion = business.GeographicRegion
		} else if req.GeographicRegion != "" {
			internalReq.GeographicRegion = req.GeographicRegion
		}

		// Add enhanced metadata
		if business.EnhancedMetadata != nil {
			internalReq.EnhancedMetadata = business.EnhancedMetadata
		}

		internalBusinesses[i] = internalReq
	}

	internalBatchReq := &classification.BatchClassificationRequest{
		Businesses: internalBusinesses,
	}

	// Perform batch classification
	response, err := h.classificationSvc.ClassifyBusinessesBatch(ctx, internalBatchReq)
	if err != nil {
		h.logger.WithComponent("api").WithError(err).Error("Batch classification failed")
		if strings.Contains(strings.ToLower(err.Error()), "batch size") {
			h.handleError(w, "Batch size exceeds limit", http.StatusBadRequest, err)
		} else {
			h.handleError(w, "Failed to classify businesses", http.StatusInternalServerError, err)
		}
		return
	}

	// Convert to enhanced batch response format
	enhancedResponses := make([]*EnhancedClassificationResponse, len(response.Classifications))
	for i, resp := range response.Classifications {
		enhancedResp := &EnhancedClassificationResponse{
			Success:              resp.Success,
			BusinessID:           resp.BusinessID,
			Classifications:      resp.Classifications,
			PrimaryIndustry:      resp.PrimaryIndustry,
			SecondaryIndustry:    resp.SecondaryIndustry,
			TertiaryIndustry:     resp.TertiaryIndustry,
			OverallConfidence:    resp.OverallConfidence,
			ValidationScore:      resp.ValidationScore,
			ClassificationMethod: resp.ClassificationMethod,
			ProcessingTime:       resp.ProcessingTime,
			Timestamp:            resp.Timestamp,
			Error:                resp.Error,
			GeographicRegion:     req.GeographicRegion,
			APIVersion:           h.getAPIVersion(r),
		}

		// Add region confidence score if geographic region was provided
		if req.GeographicRegion != "" {
			regionConfidence := h.calculateRegionConfidence(req.GeographicRegion, resp)
			enhancedResp.RegionConfidenceScore = &regionConfidence
		}

		// Add industry-specific data if available
		if resp.IndustrySpecificData != nil {
			enhancedResp.IndustrySpecificData = resp.IndustrySpecificData
		}

		enhancedResponses[i] = enhancedResp
	}

	enhancedBatchResponse := struct {
		Success         bool                              `json:"success"`
		Classifications []*EnhancedClassificationResponse `json:"classifications"`
		ProcessingTime  time.Duration                     `json:"processing_time"`
		Timestamp       time.Time                         `json:"timestamp"`
		Error           string                            `json:"error,omitempty"`
		APIVersion      string                            `json:"api_version,omitempty"`
	}{
		Success:         response.Success,
		Classifications: enhancedResponses,
		ProcessingTime:  response.ProcessingTime,
		Timestamp:       response.Timestamp,
		Error:           response.Error,
		APIVersion:      h.getAPIVersion(r),
	}

	// Log successful batch classification
	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))

	// Record batch metrics
	h.recordBatchClassificationMetrics(ctx, req, response, time.Since(start))

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(enhancedBatchResponse)
}

// GetClassification retrieves a specific classification by business ID
func (h *ClassificationHandler) GetClassification(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	businessID := r.PathValue("business_id")
	if businessID == "" {
		h.handleError(w, "business_id is required", http.StatusBadRequest, fmt.Errorf("missing business_id"))
		return
	}

	// Get classification from database
	classification, err := h.classificationSvc.GetClassificationByID(ctx, businessID)
	if err != nil {
		h.logger.WithComponent("api").WithError(err).Error("Failed to retrieve classification")
		h.handleError(w, "Failed to retrieve classification", http.StatusInternalServerError, err)
		return
	}

	if classification == nil {
		h.handleError(w, "Classification not found", http.StatusNotFound, fmt.Errorf("classification not found"))
		return
	}

	// Convert to enhanced response format
	enhancedResponse := &EnhancedClassificationResponse{
		Success:              classification.Success,
		BusinessID:           classification.BusinessID,
		Classifications:      classification.Classifications,
		PrimaryIndustry:      classification.PrimaryIndustry,
		SecondaryIndustry:    classification.SecondaryIndustry,
		TertiaryIndustry:     classification.TertiaryIndustry,
		OverallConfidence:    classification.OverallConfidence,
		ValidationScore:      classification.ValidationScore,
		ClassificationMethod: classification.ClassificationMethod,
		ProcessingTime:       classification.ProcessingTime,
		Timestamp:            classification.Timestamp,
		Error:                classification.Error,
		GeographicRegion:     classification.GeographicRegion,
		APIVersion:           h.getAPIVersion(r),
	}

	// Add region confidence score if geographic region is available
	if classification.GeographicRegion != "" {
		regionConfidence := h.calculateRegionConfidence(classification.GeographicRegion, classification)
		enhancedResponse.RegionConfidenceScore = &regionConfidence
	}

	// Add industry-specific data if available
	if classification.IndustrySpecificData != nil {
		enhancedResponse.IndustrySpecificData = classification.IndustrySpecificData
	}

	// Log successful retrieval
	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(enhancedResponse)
}

// calculateRegionConfidence calculates confidence score based on geographic region
func (h *ClassificationHandler) calculateRegionConfidence(region string, response *classification.ClassificationResponse) float64 {
	// Base confidence on the overall confidence
	baseConfidence := response.OverallConfidence

	// Apply region-specific modifiers
	regionModifiers := map[string]float64{
		"us": 1.0,  // United States - no modifier
		"ca": 0.95, // Canada - slight reduction
		"uk": 0.95, // United Kingdom - slight reduction
		"au": 0.9,  // Australia - moderate reduction
		"de": 0.9,  // Germany - moderate reduction
		"fr": 0.9,  // France - moderate reduction
		"jp": 0.85, // Japan - higher reduction
		"cn": 0.8,  // China - significant reduction
		"in": 0.8,  // India - significant reduction
		"br": 0.85, // Brazil - higher reduction
	}

	// Normalize region to lowercase
	normalizedRegion := strings.ToLower(region)

	// Apply modifier if available
	if modifier, exists := regionModifiers[normalizedRegion]; exists {
		return baseConfidence * modifier
	}

	// Default modifier for unknown regions
	return baseConfidence * 0.85
}

// getAPIVersion extracts API version from request
func (h *ClassificationHandler) getAPIVersion(r *http.Request) string {
	// Check for API version in header
	if version := r.Header.Get("X-API-Version"); version != "" {
		return version
	}

	// Check for API version in query parameter
	if version := r.URL.Query().Get("api_version"); version != "" {
		return version
	}

	// Default to v1
	return "v1"
}

// recordClassificationMetrics records metrics for single classification
func (h *ClassificationHandler) recordClassificationMetrics(ctx context.Context, req EnhancedClassificationRequest, response *classification.ClassificationResponse, duration time.Duration) {
	if h.metrics != nil {
		// Record classification metrics
		h.metrics.RecordClassificationRequest(ctx, "single", response.ClassificationMethod, response.OverallConfidence, duration)

		// Record geographic region metrics if provided
		if req.GeographicRegion != "" {
			h.metrics.RecordGeographicClassification(ctx, req.GeographicRegion, response.OverallConfidence, duration)
		}

		// Record industry-specific metrics
		if response.PrimaryIndustry.Industry != "" {
			h.metrics.RecordIndustryClassification(ctx, response.PrimaryIndustry.Industry, response.OverallConfidence, duration)
		}
	}
}

// recordBatchClassificationMetrics records metrics for batch classification
func (h *ClassificationHandler) recordBatchClassificationMetrics(ctx context.Context, req struct {
	Businesses       []EnhancedClassificationRequest `json:"businesses"`
	GeographicRegion string                          `json:"geographic_region,omitempty"`
	APIVersion       string                          `json:"api_version,omitempty"`
}, response *classification.BatchClassificationResponse, duration time.Duration) {
	if h.metrics != nil {
		// Record batch classification metrics
		h.metrics.RecordClassificationRequest(ctx, "batch", "batch", 0.0, duration)

		// Record geographic region metrics if provided
		if req.GeographicRegion != "" {
			h.metrics.RecordGeographicClassification(ctx, req.GeographicRegion, 0.0, duration)
		}

		// Record individual classification metrics
		for _, resp := range response.Classifications {
			if resp.PrimaryIndustry.Industry != "" {
				h.metrics.RecordIndustryClassification(ctx, resp.PrimaryIndustry.Industry, resp.OverallConfidence, duration)
			}
		}
	}
}

// handleError handles error responses consistently
func (h *ClassificationHandler) handleError(w http.ResponseWriter, message string, statusCode int, err error) {
	h.logger.WithComponent("api").WithError(err).Error(message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"error":   "classification_error",
		"message": message,
		"status":  statusCode,
	}

	if err != nil {
		errorResponse["details"] = err.Error()
	}

	json.NewEncoder(w).Encode(errorResponse)
}
