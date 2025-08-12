package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/pkg/validators"
)

// MultiIndustryClassificationHandler handles multi-industry classification requests
type MultiIndustryClassificationHandler struct {
	multiIndustryService *classification.MultiIndustryService
	logger               *observability.Logger
	metrics              *observability.Metrics
}

// NewMultiIndustryClassificationHandler creates a new multi-industry classification handler
func NewMultiIndustryClassificationHandler(
	multiIndustryService *classification.MultiIndustryService,
	logger *observability.Logger,
	metrics *observability.Metrics,
) *MultiIndustryClassificationHandler {
	return &MultiIndustryClassificationHandler{
		multiIndustryService: multiIndustryService,
		logger:               logger,
		metrics:              metrics,
	}
}

// MultiIndustryClassificationRequest represents a multi-industry classification request
type MultiIndustryClassificationRequest struct {
	BusinessName       string `json:"business_name" validate:"required,min=1,max=200"`
	BusinessType       string `json:"business_type,omitempty" validate:"omitempty,max=100"`
	Industry           string `json:"industry,omitempty" validate:"omitempty,max=100"`
	Description        string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Keywords           string `json:"keywords,omitempty" validate:"omitempty,max=500"`
	RegistrationNumber string `json:"registration_number,omitempty" validate:"omitempty,max=50"`
}

// MultiIndustryClassificationResponse represents a multi-industry classification response
type MultiIndustryClassificationResponse struct {
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
}

// HandleMultiIndustryClassification handles multi-industry classification requests
func (h *MultiIndustryClassificationHandler) HandleMultiIndustryClassification(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Log request start
	h.logger.WithComponent("multi_industry_classification_handler").LogBusinessEvent(ctx, "multi_industry_classification_request_started", "", map[string]interface{}{
		"method":     r.Method,
		"user_agent": r.UserAgent(),
		"remote_ip":  r.RemoteAddr,
	})

	// Parse request
	var req MultiIndustryClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, "Invalid JSON request", http.StatusBadRequest, err)
		return
	}

	// Validate request
	if err := validators.Validate(req); err != nil {
		h.handleError(w, "Validation failed", http.StatusBadRequest, err)
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

	// Perform multi-industry classification
	result, err := h.multiIndustryService.ClassifyBusinessMultiIndustry(ctx, internalReq)
	if err != nil {
		h.handleError(w, "Classification failed", http.StatusInternalServerError, err)
		return
	}

	// Create enhanced result presentation
	presentationEngine := classification.NewResultPresentationEngine(h.logger, h.metrics)
	enhancedResult := presentationEngine.PresentMultiIndustryResult(ctx, result, internalReq)

	// Log successful completion
	h.logger.WithComponent("multi_industry_classification_handler").LogBusinessEvent(ctx, "multi_industry_classification_request_completed", "", map[string]interface{}{
		"business_name":         req.BusinessName,
		"primary_industry_code": result.PrimaryIndustry.IndustryCode,
		"primary_industry_name": result.PrimaryIndustry.IndustryName,
		"overall_confidence":    result.OverallConfidence,
		"validation_score":      result.ValidationScore,
		"total_classifications": len(result.Classifications),
		"processing_time_ms":    time.Since(start).Milliseconds(),
		"http_status":           http.StatusOK,
	})

	// Record metrics
	h.metrics.RecordBusinessClassification("multi_industry_api_success", "1")

	// Send enhanced response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(enhancedResult); err != nil {
		h.logger.WithComponent("multi_industry_classification_handler").LogBusinessEvent(ctx, "multi_industry_classification_response_encoding_error", "", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// HandleBatchMultiIndustryClassification handles batch multi-industry classification requests
func (h *MultiIndustryClassificationHandler) HandleBatchMultiIndustryClassification(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Log request start
	h.logger.WithComponent("multi_industry_classification_handler").LogBusinessEvent(ctx, "batch_multi_industry_classification_request_started", "", map[string]interface{}{
		"method":     r.Method,
		"user_agent": r.UserAgent(),
		"remote_ip":  r.RemoteAddr,
	})

	// Parse request
	var req struct {
		Businesses []MultiIndustryClassificationRequest `json:"businesses" validate:"required,min=1,max=100"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, "Invalid JSON request", http.StatusBadRequest, err)
		return
	}

	// Validate request
	if err := validators.Validate(req); err != nil {
		h.handleError(w, "Validation failed", http.StatusBadRequest, err)
		return
	}

	// Process each business
	var responses []MultiIndustryClassificationResponse
	successCount := 0
	errorCount := 0

	for _, businessReq := range req.Businesses {
		// Validate individual business request
		if err := validators.Validate(businessReq); err != nil {
			responses = append(responses, MultiIndustryClassificationResponse{
				Success:   false,
				Error:     "Validation failed: " + err.Error(),
				Timestamp: time.Now(),
			})
			errorCount++
			continue
		}

		// Convert to internal request format
		internalReq := &classification.ClassificationRequest{
			BusinessName:       businessReq.BusinessName,
			BusinessType:       businessReq.BusinessType,
			Industry:           businessReq.Industry,
			Description:        businessReq.Description,
			Keywords:           businessReq.Keywords,
			RegistrationNumber: businessReq.RegistrationNumber,
		}

		// Perform classification
		result, err := h.multiIndustryService.ClassifyBusinessMultiIndustry(ctx, internalReq)
		if err != nil {
			responses = append(responses, MultiIndustryClassificationResponse{
				Success:   false,
				Error:     "Classification failed: " + err.Error(),
				Timestamp: time.Now(),
			})
			errorCount++
			continue
		}

		// Create successful response
		response := MultiIndustryClassificationResponse{
			Success:              true,
			Classifications:      result.Classifications,
			PrimaryIndustry:      result.PrimaryIndustry,
			SecondaryIndustry:    result.SecondaryIndustry,
			TertiaryIndustry:     result.TertiaryIndustry,
			OverallConfidence:    result.OverallConfidence,
			ValidationScore:      result.ValidationScore,
			ClassificationMethod: result.ClassificationMethod,
			ProcessingTime:       result.ProcessingTime,
			Timestamp:            time.Now(),
		}

		responses = append(responses, response)
		successCount++
	}

	// Create batch response
	batchResponse := struct {
		Success        bool                                  `json:"success"`
		Results        []MultiIndustryClassificationResponse `json:"results"`
		TotalProcessed int                                   `json:"total_processed"`
		SuccessCount   int                                   `json:"success_count"`
		ErrorCount     int                                   `json:"error_count"`
		ProcessingTime time.Duration                         `json:"processing_time"`
		Timestamp      time.Time                             `json:"timestamp"`
	}{
		Success:        successCount > 0,
		Results:        responses,
		TotalProcessed: len(req.Businesses),
		SuccessCount:   successCount,
		ErrorCount:     errorCount,
		ProcessingTime: time.Since(start),
		Timestamp:      time.Now(),
	}

	// Log completion
	h.logger.WithComponent("multi_industry_classification_handler").LogBusinessEvent(ctx, "batch_multi_industry_classification_request_completed", "", map[string]interface{}{
		"total_processed":    len(req.Businesses),
		"success_count":      successCount,
		"error_count":        errorCount,
		"processing_time_ms": time.Since(start).Milliseconds(),
		"http_status":        http.StatusOK,
	})

	// Record metrics
	h.metrics.RecordBusinessClassification("batch_multi_industry_api_success", "1")

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(batchResponse); err != nil {
		h.logger.WithComponent("multi_industry_classification_handler").LogBusinessEvent(ctx, "batch_multi_industry_classification_response_encoding_error", "", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// handleError handles error responses
func (h *MultiIndustryClassificationHandler) handleError(w http.ResponseWriter, message string, statusCode int, err error) {
	ctx := context.Background()

	errorResponse := struct {
		Success   bool      `json:"success"`
		Error     string    `json:"error"`
		Timestamp time.Time `json:"timestamp"`
	}{
		Success:   false,
		Error:     message,
		Timestamp: time.Now(),
	}

	// Log error
	h.logger.WithComponent("multi_industry_classification_handler").LogBusinessEvent(ctx, "multi_industry_classification_request_error", "", map[string]interface{}{
		"error":       err.Error(),
		"status_code": statusCode,
		"message":     message,
	})

	// Record metrics
	h.metrics.RecordBusinessClassification("multi_industry_api_error", "1")

	// Send error response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		h.logger.WithComponent("multi_industry_classification_handler").LogBusinessEvent(ctx, "multi_industry_classification_error_response_encoding_error", "", map[string]interface{}{
			"error": err.Error(),
		})
	}
}
