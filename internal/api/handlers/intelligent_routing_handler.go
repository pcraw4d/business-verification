package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"kyb-platform/internal/classification"
	"kyb-platform/internal/observability"
	"kyb-platform/internal/routing"
	"kyb-platform/internal/shared"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// IntelligentRoutingHandler provides HTTP handlers for the intelligent routing system
type IntelligentRoutingHandler struct {
	router          *routing.IntelligentRouter
	detectionService *classification.IndustryDetectionService
	logger          *observability.Logger
	metrics         *observability.Metrics
	tracer          trace.Tracer
	requestIDGen    func() string
}

// NewIntelligentRoutingHandler creates a new intelligent routing handler
func NewIntelligentRoutingHandler(
	router *routing.IntelligentRouter,
	detectionService *classification.IndustryDetectionService,
	logger *observability.Logger,
	metrics *observability.Metrics,
	tracer trace.Tracer,
) *IntelligentRoutingHandler {
	return &IntelligentRoutingHandler{
		router:          router,
		detectionService: detectionService,
		logger:          logger,
		metrics:         metrics,
		tracer:          tracer,
		requestIDGen:    generateRequestID,
	}
}

// ClassifyBusiness handles single business classification requests using intelligent routing
func (h *IntelligentRoutingHandler) ClassifyBusiness(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := h.requestIDGen()

	ctx, span := h.tracer.Start(r.Context(), "IntelligentRoutingHandler.ClassifyBusiness")
	defer span.End()

	span.SetAttributes(
		attribute.String("request.id", requestID),
		attribute.String("http.method", r.Method),
		attribute.String("http.url", r.URL.String()),
	)

	// Add request ID to context
	ctx = context.WithValue(ctx, "request_id", requestID)

	// Parse and validate request
	req, err := h.parseClassificationRequest(r)
	if err != nil {
		h.handleError(w, err, http.StatusBadRequest, "invalid_request", requestID)
		return
	}

	// Set request ID
	req.ID = requestID

	// Log request start
	h.logger.WithComponent("intelligent_routing_handler").Info("classification_request_started", map[string]interface{}{
		"request_id":    requestID,
		"business_name": req.BusinessName,
		"website_url":   req.WebsiteURL,
		"user_agent":    r.UserAgent(),
	})

	// FIX: Actually perform classification using detection service
	if h.detectionService == nil {
		h.handleError(w, fmt.Errorf("classification service not available"), http.StatusInternalServerError, "service_unavailable", requestID)
		return
	}

	classificationResult, err := h.detectionService.DetectIndustry(
		ctx,
		req.BusinessName,
		req.Description,
		req.WebsiteURL,
	)
	if err != nil {
		h.handleError(w, err, http.StatusInternalServerError, "classification_failed", requestID)
		return
	}

	// Convert to API response format
	response := &shared.BusinessClassificationResponse{
		ID:                    requestID,
		BusinessName:          req.BusinessName,
		DetectedIndustry:      classificationResult.IndustryName,
		Confidence:            classificationResult.Confidence,
		ClassificationMethod:  classificationResult.Method,
		ProcessingTime:        classificationResult.ProcessingTime,
		CreatedAt:             classificationResult.CreatedAt,
		Timestamp:             time.Now(),
		Classifications: []shared.IndustryClassification{
			{
				IndustryName:         classificationResult.IndustryName,
				ConfidenceScore:      classificationResult.Confidence,
				ClassificationMethod: classificationResult.Method,
				Keywords:             classificationResult.Keywords,
			},
		},
		PrimaryClassification: &shared.IndustryClassification{
			IndustryName:         classificationResult.IndustryName,
			ConfidenceScore:      classificationResult.Confidence,
			ClassificationMethod: classificationResult.Method,
			Keywords:             classificationResult.Keywords,
		},
		OverallConfidence:     classificationResult.Confidence,
		ClassificationReasoning: classificationResult.Reasoning,
		Metadata: map[string]interface{}{
			"method":     classificationResult.Method,
			"request_id": requestID,
		},
	}

	// Record metrics
	h.recordMetrics(ctx, requestID, true, time.Since(startTime))

	// Log successful completion
	h.logger.WithComponent("intelligent_routing_handler").Info("classification_request_completed", map[string]interface{}{
		"request_id":         requestID,
		"business_name":      req.BusinessName,
		"detected_industry":   classificationResult.IndustryName,
		"confidence":         classificationResult.Confidence,
		"processing_time_ms": time.Since(startTime).Milliseconds(),
		"response_status":    "success",
	})

	// Return response
	h.writeResponse(w, response, http.StatusOK)
}

// ClassifyBusinessBatch is an alias for ClassifyBusinessesBatch for backward compatibility
func (h *IntelligentRoutingHandler) ClassifyBusinessBatch(w http.ResponseWriter, r *http.Request) {
	h.ClassifyBusinessesBatch(w, r)
}

// ClassifyBusinessesBatch handles batch business classification requests using intelligent routing
func (h *IntelligentRoutingHandler) ClassifyBusinessesBatch(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := h.requestIDGen()

	ctx, span := h.tracer.Start(r.Context(), "IntelligentRoutingHandler.ClassifyBusinessesBatch")
	defer span.End()

	span.SetAttributes(
		attribute.String("request.id", requestID),
		attribute.String("http.method", r.Method),
		attribute.String("http.url", r.URL.String()),
	)

	// Add request ID to context
	ctx = context.WithValue(ctx, "request_id", requestID)

	// Parse and validate batch request
	batchReq, err := h.parseBatchClassificationRequest(r)
	if err != nil {
		h.handleError(w, err, http.StatusBadRequest, "invalid_batch_request", requestID)
		return
	}

	// Validate batch size
	if len(batchReq.Requests) == 0 {
		h.handleError(w, fmt.Errorf("batch must contain at least one business"), http.StatusBadRequest, "empty_batch", requestID)
		return
	}

	if len(batchReq.Requests) > 100 {
		h.handleError(w, fmt.Errorf("batch size exceeds maximum of 100 businesses"), http.StatusBadRequest, "batch_size_exceeded", requestID)
		return
	}

	// Log batch request start
	h.logger.WithComponent("intelligent_routing_handler").Info("batch_classification_request_started", map[string]interface{}{
		"request_id":     requestID,
		"business_count": len(batchReq.Requests),
		"user_agent":     r.UserAgent(),
	})

	// Process each business in the batch
	responses := make([]shared.BusinessClassificationResponse, 0, len(batchReq.Requests))
	errors := make([]shared.BatchError, 0)

	if h.detectionService == nil {
		h.handleError(w, fmt.Errorf("classification service not available"), http.StatusInternalServerError, "service_unavailable", requestID)
		return
	}

	for i, businessReq := range batchReq.Requests {
		// Set request ID for each business
		businessReq.ID = fmt.Sprintf("%s_%d", requestID, i)

		// FIX: Actually perform classification using detection service
		classificationResult, err := h.detectionService.DetectIndustry(
			ctx,
			businessReq.BusinessName,
			businessReq.Description,
			businessReq.WebsiteURL,
		)
		if err != nil {
			batchError := shared.BatchError{
				Index:        i,
				BusinessName: businessReq.BusinessName,
				Error:        err.Error(),
			}
			errors = append(errors, batchError)
			h.logger.WithComponent("intelligent_routing_handler").Warn("batch_business_failed", map[string]interface{}{
				"request_id":     requestID,
				"business_name":  businessReq.BusinessName,
				"business_index": i,
				"error":          err.Error(),
			})
			continue
		}

		// Convert to API response format
		response := shared.BusinessClassificationResponse{
			ID:                    businessReq.ID,
			BusinessName:          businessReq.BusinessName,
			DetectedIndustry:      classificationResult.IndustryName,
			Confidence:            classificationResult.Confidence,
			ClassificationMethod:  classificationResult.Method,
			ProcessingTime:        classificationResult.ProcessingTime,
			CreatedAt:             classificationResult.CreatedAt,
			Timestamp:             time.Now(),
			Classifications: []shared.IndustryClassification{
				{
					IndustryName:         classificationResult.IndustryName,
					ConfidenceScore:      classificationResult.Confidence,
					ClassificationMethod: classificationResult.Method,
					Keywords:             classificationResult.Keywords,
				},
			},
			PrimaryClassification: &shared.IndustryClassification{
				IndustryName:         classificationResult.IndustryName,
				ConfidenceScore:      classificationResult.Confidence,
				ClassificationMethod: classificationResult.Method,
				Keywords:             classificationResult.Keywords,
			},
			OverallConfidence:     classificationResult.Confidence,
			ClassificationReasoning: classificationResult.Reasoning,
			Metadata: map[string]interface{}{
				"method":     classificationResult.Method,
				"request_id": businessReq.ID,
			},
		}

		responses = append(responses, response)
	}

	// Create batch response
	batchResponse := &shared.BatchClassificationResponse{
		ID:             requestID,
		Responses:      responses,
		TotalCount:     len(batchReq.Requests),
		SuccessCount:   len(responses),
		ErrorCount:     len(errors),
		Errors:         errors,
		ProcessingTime: time.Since(startTime),
		CompletedAt:    time.Now(),
	}

	// Record metrics
	h.recordBatchMetrics(ctx, requestID, len(batchReq.Requests), len(responses), time.Since(startTime))

	// Log batch completion
	h.logger.WithComponent("intelligent_routing_handler").Info("batch_classification_request_completed", map[string]interface{}{
		"request_id":         requestID,
		"total_businesses":   len(batchReq.Requests),
		"successful_count":   len(responses),
		"error_count":        len(errors),
		"processing_time_ms": time.Since(startTime).Milliseconds(),
	})

	// Return batch response
	h.writeResponse(w, batchResponse, http.StatusOK)
}

// GetRoutingHealth returns the health status of the intelligent routing system
func (h *IntelligentRoutingHandler) GetRoutingHealth(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.Start(r.Context(), "IntelligentRoutingHandler.GetRoutingHealth")
	defer span.End()

	// Get router health status
	health := map[string]interface{}{
		"status":          "healthy",
		"timestamp":       time.Now(),
		"router_id":       "intelligent_routing_system",
		"version":         "1.0.0",
		"uptime":          time.Since(time.Now().Add(-24 * time.Hour)), // Mock uptime
		"active_requests": 0,                                           // Would be implemented with actual tracking
	}

	h.writeResponse(w, health, http.StatusOK)
}

// GetRoutingMetrics returns performance metrics for the intelligent routing system
func (h *IntelligentRoutingHandler) GetRoutingMetrics(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.Start(r.Context(), "IntelligentRoutingHandler.GetRoutingMetrics")
	defer span.End()

	// Get router metrics
	metrics := map[string]interface{}{
		"total_requests":        0, // Would be implemented with actual metrics
		"successful_requests":   0,
		"failed_requests":       0,
		"average_response_time": 0,
		"requests_per_minute":   0,
		"timestamp":             time.Now(),
	}

	h.writeResponse(w, metrics, http.StatusOK)
}

// parseClassificationRequest parses and validates a classification request
func (h *IntelligentRoutingHandler) parseClassificationRequest(r *http.Request) (*shared.BusinessClassificationRequest, error) {
	var req shared.BusinessClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("failed to decode request body: %w", err)
	}

	// Validate required fields
	if req.BusinessName == "" {
		return nil, fmt.Errorf("business_name is required")
	}

	// Set defaults
	if req.ID == "" {
		req.ID = h.requestIDGen()
	}

	return &req, nil
}

// parseBatchClassificationRequest parses and validates a batch classification request
func (h *IntelligentRoutingHandler) parseBatchClassificationRequest(r *http.Request) (*shared.BatchClassificationRequest, error) {
	var batchReq shared.BatchClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&batchReq); err != nil {
		return nil, fmt.Errorf("failed to decode batch request body: %w", err)
	}

	// Validate batch request
	if len(batchReq.Requests) == 0 {
		return nil, fmt.Errorf("batch must contain at least one business")
	}

	// Validate each business in the batch
	for i, business := range batchReq.Requests {
		if business.BusinessName == "" {
			return nil, fmt.Errorf("business_name is required for business at index %d", i)
		}
	}

	return &batchReq, nil
}

// handleError handles error responses with proper logging and metrics
func (h *IntelligentRoutingHandler) handleError(w http.ResponseWriter, err error, statusCode int, errorCode, requestID string) {
	// Log error
	h.logger.WithComponent("intelligent_routing_handler").Error("classification_request_failed", map[string]interface{}{
		"request_id":  requestID,
		"error_code":  errorCode,
		"status_code": statusCode,
		"error":       err.Error(),
	})

	// Record error metrics
	if h.metrics != nil {
		h.metrics.IncCounter("classification_errors_total", map[string]string{
			"error_code":  errorCode,
			"status_code": fmt.Sprintf("%d", statusCode),
		})
	}

	// Write error response
	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    errorCode,
			"message": err.Error(),
		},
		"request_id": requestID,
		"timestamp":  time.Now(),
	}

	h.writeResponse(w, errorResponse, statusCode)
}

// writeResponse writes a JSON response with proper headers
func (h *IntelligentRoutingHandler) writeResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract request ID from response if available
	if resp, ok := data.(*shared.BusinessClassificationResponse); ok {
		w.Header().Set("X-Request-ID", resp.ID)
	} else if resp, ok := data.(*shared.BatchClassificationResponse); ok {
		w.Header().Set("X-Request-ID", resp.ID)
	} else if respMap, ok := data.(map[string]interface{}); ok {
		if reqID, exists := respMap["request_id"]; exists {
			w.Header().Set("X-Request-ID", fmt.Sprintf("%v", reqID))
		}
	}
	
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithComponent("intelligent_routing_handler").Error("failed_to_encode_response", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// recordMetrics records metrics for successful requests
func (h *IntelligentRoutingHandler) recordMetrics(ctx context.Context, requestID string, success bool, duration time.Duration) {
	if h.metrics == nil {
		return
	}

	// Record request count
	h.metrics.IncCounter("classification_requests_total", map[string]string{
		"status": "success",
	})

	// Record request duration
	h.metrics.RecordHistogram("classification_request_duration_seconds", duration.Seconds(), map[string]string{
		"status": "success",
	})
}

// recordBatchMetrics records metrics for batch requests
func (h *IntelligentRoutingHandler) recordBatchMetrics(ctx context.Context, requestID string, totalCount, successCount int, duration time.Duration) {
	if h.metrics == nil {
		return
	}

	// Record batch request count
	h.metrics.IncCounter("batch_classification_requests_total", map[string]string{
		"status": "success",
	})

	// Record batch size
	h.metrics.RecordHistogram("batch_classification_size", float64(totalCount), map[string]string{
		"status": "success",
	})

	// Record batch duration
	h.metrics.RecordHistogram("batch_classification_duration_seconds", duration.Seconds(), map[string]string{
		"status": "success",
	})

	// Record success rate
	successRate := float64(successCount) / float64(totalCount)
	h.metrics.RecordHistogram("batch_classification_success_rate", successRate, map[string]string{
		"status": "success",
	})
}
