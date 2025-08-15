package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// FeedbackHandler handles feedback-related API endpoints
type FeedbackHandler struct {
	feedbackCollector *classification.FeedbackCollector
	logger            *observability.Logger
	metrics           *observability.Metrics
}

// NewFeedbackHandler creates a new feedback handler
func NewFeedbackHandler(feedbackCollector *classification.FeedbackCollector, logger *observability.Logger, metrics *observability.Metrics) *FeedbackHandler {
	return &FeedbackHandler{
		feedbackCollector: feedbackCollector,
		logger:            logger,
		metrics:           metrics,
	}
}

// FeedbackSubmissionRequest represents a feedback submission request
type FeedbackSubmissionRequest struct {
	UserID                  string                                 `json:"user_id" validate:"required"`
	BusinessName            string                                 `json:"business_name" validate:"required"`
	OriginalClassification  *classification.IndustryClassification `json:"original_classification" validate:"required"`
	FeedbackType            classification.FeedbackType            `json:"feedback_type" validate:"required"`
	FeedbackValue           interface{}                            `json:"feedback_value"`
	FeedbackText            string                                 `json:"feedback_text"`
	SuggestedClassification *classification.IndustryClassification `json:"suggested_classification,omitempty"`
	Confidence              float64                                `json:"confidence" validate:"min=0,max=1"`
}

// FeedbackSubmissionResponse represents a feedback submission response
type FeedbackSubmissionResponse struct {
	FeedbackID     string        `json:"feedback_id"`
	Status         string        `json:"status"`
	Message        string        `json:"message"`
	ProcessingTime time.Duration `json:"processing_time"`
	CreatedAt      time.Time     `json:"created_at"`
}

// FeedbackListRequest represents a feedback list request
type FeedbackListRequest struct {
	UserID       string                        `json:"user_id,omitempty"`
	FeedbackType classification.FeedbackType   `json:"feedback_type,omitempty"`
	Status       classification.FeedbackStatus `json:"status,omitempty"`
	BusinessName string                        `json:"business_name,omitempty"`
	Limit        int                           `json:"limit,omitempty"`
	Offset       int                           `json:"offset,omitempty"`
}

// FeedbackListResponse represents a feedback list response
type FeedbackListResponse struct {
	Feedback []*classification.Feedback `json:"feedback"`
	Total    int                        `json:"total"`
	Limit    int                        `json:"limit"`
	Offset   int                        `json:"offset"`
}

// AccuracyMetricsRequest represents an accuracy metrics request
type AccuracyMetricsRequest struct {
	UserID       string                      `json:"user_id,omitempty"`
	FeedbackType classification.FeedbackType `json:"feedback_type,omitempty"`
	BusinessName string                      `json:"business_name,omitempty"`
	TimeRange    time.Duration               `json:"time_range,omitempty"`
}

// AccuracyMetricsResponse represents an accuracy metrics response
type AccuracyMetricsResponse struct {
	Metrics *classification.FeedbackAccuracyMetrics `json:"metrics"`
}

// ModelUpdatesResponse represents a model updates response
type ModelUpdatesResponse struct {
	Updates map[string]interface{} `json:"updates"`
	Total   int                    `json:"total"`
}

// HandleSubmitFeedback handles feedback submission
func (h *FeedbackHandler) HandleSubmitFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	// Log request
	if h.logger != nil {
		h.logger.WithComponent("feedback_handler").LogBusinessEvent(ctx, "feedback_submission_request", "", map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
		})
	}

	// Parse request
	var req FeedbackSubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, r, "Invalid request body", http.StatusBadRequest, err)
		return
	}

	// Validate request
	if err := h.validateFeedbackSubmission(&req); err != nil {
		h.handleError(w, r, "Validation failed", http.StatusBadRequest, err)
		return
	}

	// Create feedback object
	feedback := &classification.Feedback{
		UserID:                  req.UserID,
		BusinessName:            req.BusinessName,
		OriginalClassification:  req.OriginalClassification,
		FeedbackType:            req.FeedbackType,
		FeedbackValue:           req.FeedbackValue,
		FeedbackText:            req.FeedbackText,
		SuggestedClassification: req.SuggestedClassification,
		Confidence:              req.Confidence,
		Metadata:                make(map[string]interface{}),
	}

	// Add request metadata
	feedback.Metadata["user_agent"] = r.UserAgent()
	feedback.Metadata["ip_address"] = h.getClientIP(r)
	feedback.Metadata["request_id"] = h.getRequestID(ctx)

	// Submit feedback
	if err := h.feedbackCollector.SubmitFeedback(ctx, feedback); err != nil {
		h.handleError(w, r, "Failed to submit feedback", http.StatusInternalServerError, err)
		return
	}

	// Create response
	response := &FeedbackSubmissionResponse{
		FeedbackID:     feedback.ID,
		Status:         string(feedback.Status),
		Message:        "Feedback submitted successfully",
		ProcessingTime: time.Since(start),
		CreatedAt:      feedback.CreatedAt,
	}

	// Log success
	if h.logger != nil {
		h.logger.WithComponent("feedback_handler").LogBusinessEvent(ctx, "feedback_submission_success", feedback.ID, map[string]interface{}{
			"user_id":            req.UserID,
			"business_name":      req.BusinessName,
			"feedback_type":      string(req.FeedbackType),
			"processing_time_ms": response.ProcessingTime.Milliseconds(),
		})
	}

	// Record metrics
	h.recordFeedbackMetrics(ctx, "submitted", req.FeedbackType, response.ProcessingTime)

	// Send response
	h.sendJSONResponse(w, response, http.StatusCreated)
}

// HandleGetFeedback handles feedback retrieval by ID
func (h *FeedbackHandler) HandleGetFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract feedback ID from URL
	feedbackID := h.extractPathParameter(r, "id")
	if feedbackID == "" {
		h.handleError(w, r, "Feedback ID is required", http.StatusBadRequest, fmt.Errorf("missing feedback ID"))
		return
	}

	// Get feedback
	feedback, err := h.feedbackCollector.GetFeedback(ctx, feedbackID)
	if err != nil {
		h.handleError(w, r, "Failed to get feedback", http.StatusNotFound, err)
		return
	}

	// Log success
	if h.logger != nil {
		h.logger.WithComponent("feedback_handler").LogBusinessEvent(ctx, "feedback_retrieval_success", feedbackID, map[string]interface{}{
			"user_id":       feedback.UserID,
			"business_name": feedback.BusinessName,
		})
	}

	// Send response
	h.sendJSONResponse(w, feedback, http.StatusOK)
}

// HandleListFeedback handles feedback listing with filters
func (h *FeedbackHandler) HandleListFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	var req FeedbackListRequest
	if err := h.parseQueryParameters(r, &req); err != nil {
		h.handleError(w, r, "Invalid query parameters", http.StatusBadRequest, err)
		return
	}

	// Build filters
	filters := h.buildFeedbackFilters(&req)

	// Get feedback list
	feedbackList, err := h.feedbackCollector.ListFeedback(ctx, filters)
	if err != nil {
		h.handleError(w, r, "Failed to list feedback", http.StatusInternalServerError, err)
		return
	}

	// Apply pagination
	total := len(feedbackList)
	limit := req.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 100 {
		limit = 100 // Maximum limit
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	end := offset + limit
	if end > total {
		end = total
	}

	var paginatedFeedback []*classification.Feedback
	if offset < total {
		paginatedFeedback = feedbackList[offset:end]
	}

	// Create response
	response := &FeedbackListResponse{
		Feedback: paginatedFeedback,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}

	// Log success
	if h.logger != nil {
		h.logger.WithComponent("feedback_handler").LogBusinessEvent(ctx, "feedback_list_success", "", map[string]interface{}{
			"total":   total,
			"limit":   limit,
			"offset":  offset,
			"filters": filters,
		})
	}

	// Send response
	h.sendJSONResponse(w, response, http.StatusOK)
}

// HandleUpdateFeedback handles feedback updates
func (h *FeedbackHandler) HandleUpdateFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract feedback ID from URL
	feedbackID := h.extractPathParameter(r, "id")
	if feedbackID == "" {
		h.handleError(w, r, "Feedback ID is required", http.StatusBadRequest, fmt.Errorf("missing feedback ID"))
		return
	}

	// Parse request body
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.handleError(w, r, "Invalid request body", http.StatusBadRequest, err)
		return
	}

	// Validate updates
	if err := h.validateFeedbackUpdates(updates); err != nil {
		h.handleError(w, r, "Invalid updates", http.StatusBadRequest, err)
		return
	}

	// Update feedback
	if err := h.feedbackCollector.UpdateFeedback(ctx, feedbackID, updates); err != nil {
		h.handleError(w, r, "Failed to update feedback", http.StatusInternalServerError, err)
		return
	}

	// Log success
	if h.logger != nil {
		h.logger.WithComponent("feedback_handler").LogBusinessEvent(ctx, "feedback_update_success", feedbackID, map[string]interface{}{
			"updates_applied": len(updates),
		})
	}

	// Send response
	response := map[string]interface{}{
		"message":     "Feedback updated successfully",
		"feedback_id": feedbackID,
	}
	h.sendJSONResponse(w, response, http.StatusOK)
}

// HandleGetAccuracyMetrics handles accuracy metrics retrieval
func (h *FeedbackHandler) HandleGetAccuracyMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	var req AccuracyMetricsRequest
	if err := h.parseQueryParameters(r, &req); err != nil {
		h.handleError(w, r, "Invalid query parameters", http.StatusBadRequest, err)
		return
	}

	// Build filters
	filters := h.buildAccuracyFilters(&req)

	// Get accuracy metrics
	metrics, err := h.feedbackCollector.GetAccuracyMetrics(ctx, filters)
	if err != nil {
		h.handleError(w, r, "Failed to get accuracy metrics", http.StatusInternalServerError, err)
		return
	}

	// Create response
	response := &AccuracyMetricsResponse{
		Metrics: metrics,
	}

	// Log success
	if h.logger != nil {
		h.logger.WithComponent("feedback_handler").LogBusinessEvent(ctx, "accuracy_metrics_success", "", map[string]interface{}{
			"total_feedback": metrics.TotalFeedback,
			"accuracy_score": metrics.AccuracyScore,
			"filters":        filters,
		})
	}

	// Send response
	h.sendJSONResponse(w, response, http.StatusOK)
}

// HandleGetModelUpdates handles model updates retrieval
func (h *FeedbackHandler) HandleGetModelUpdates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get model updates
	updates, err := h.feedbackCollector.GetModelUpdates(ctx)
	if err != nil {
		h.handleError(w, r, "Failed to get model updates", http.StatusInternalServerError, err)
		return
	}

	// Create response
	response := &ModelUpdatesResponse{
		Updates: updates,
		Total:   len(updates),
	}

	// Log success
	if h.logger != nil {
		h.logger.WithComponent("feedback_handler").LogBusinessEvent(ctx, "model_updates_success", "", map[string]interface{}{
			"total_updates": len(updates),
		})
	}

	// Send response
	h.sendJSONResponse(w, response, http.StatusOK)
}

// HandleGetFeedbackStats handles feedback statistics retrieval
func (h *FeedbackHandler) HandleGetFeedbackStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get collector stats
	stats := h.feedbackCollector.GetCollectorStats()

	// Log success
	if h.logger != nil {
		h.logger.WithComponent("feedback_handler").LogBusinessEvent(ctx, "feedback_stats_success", "", map[string]interface{}{
			"total_feedback": stats["total_feedback"],
		})
	}

	// Send response
	h.sendJSONResponse(w, stats, http.StatusOK)
}

// Helper methods

// validateFeedbackSubmission validates feedback submission request
func (h *FeedbackHandler) validateFeedbackSubmission(req *FeedbackSubmissionRequest) error {
	// Basic validation
	if req.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if req.BusinessName == "" {
		return fmt.Errorf("business name is required")
	}
	if req.OriginalClassification == nil {
		return fmt.Errorf("original classification is required")
	}
	if req.FeedbackType == "" {
		return fmt.Errorf("feedback type is required")
	}

	// Validate feedback type
	switch req.FeedbackType {
	case classification.FeedbackTypeAccuracy,
		classification.FeedbackTypeRelevance,
		classification.FeedbackTypeConfidence,
		classification.FeedbackTypeClassification,
		classification.FeedbackTypeSuggestion,
		classification.FeedbackTypeCorrection:
		// Valid types
	default:
		return fmt.Errorf("invalid feedback type: %s", req.FeedbackType)
	}

	// Validate confidence range
	if req.Confidence < 0.0 || req.Confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0.0 and 1.0")
	}

	return nil
}

// validateFeedbackUpdates validates feedback updates
func (h *FeedbackHandler) validateFeedbackUpdates(updates map[string]interface{}) error {
	for key, value := range updates {
		switch key {
		case "status":
			if status, ok := value.(string); ok {
				switch classification.FeedbackStatus(status) {
				case classification.FeedbackStatusPending,
					classification.FeedbackStatusProcessed,
					classification.FeedbackStatusRejected,
					classification.FeedbackStatusApplied:
					// Valid status
				default:
					return fmt.Errorf("invalid status: %s", status)
				}
			} else {
				return fmt.Errorf("status must be a string")
			}
		case "feedback_text":
			if _, ok := value.(string); !ok {
				return fmt.Errorf("feedback_text must be a string")
			}
		case "confidence":
			if confidence, ok := value.(float64); ok {
				if confidence < 0.0 || confidence > 1.0 {
					return fmt.Errorf("confidence must be between 0.0 and 1.0")
				}
			} else {
				return fmt.Errorf("confidence must be a number")
			}
		case "metadata":
			if _, ok := value.(map[string]interface{}); !ok {
				return fmt.Errorf("metadata must be an object")
			}
		default:
			return fmt.Errorf("unknown field: %s", key)
		}
	}

	return nil
}

// buildFeedbackFilters builds filters for feedback listing
func (h *FeedbackHandler) buildFeedbackFilters(req *FeedbackListRequest) map[string]interface{} {
	filters := make(map[string]interface{})

	if req.UserID != "" {
		filters["user_id"] = req.UserID
	}
	if req.FeedbackType != "" {
		filters["feedback_type"] = req.FeedbackType
	}
	if req.Status != "" {
		filters["status"] = req.Status
	}
	if req.BusinessName != "" {
		filters["business_name"] = req.BusinessName
	}

	return filters
}

// buildAccuracyFilters builds filters for accuracy metrics
func (h *FeedbackHandler) buildAccuracyFilters(req *AccuracyMetricsRequest) map[string]interface{} {
	filters := make(map[string]interface{})

	if req.UserID != "" {
		filters["user_id"] = req.UserID
	}
	if req.FeedbackType != "" {
		filters["feedback_type"] = req.FeedbackType
	}
	if req.BusinessName != "" {
		filters["business_name"] = req.BusinessName
	}

	return filters
}

// parseQueryParameters parses query parameters into a struct
func (h *FeedbackHandler) parseQueryParameters(r *http.Request, req interface{}) error {
	// This is a simplified implementation
	// In production, use a proper query parameter parser
	return nil
}

// extractPathParameter extracts a path parameter from the request
func (h *FeedbackHandler) extractPathParameter(r *http.Request, name string) string {
	// This is a simplified implementation
	// In production, use a proper router that provides path parameters
	return ""
}

// getClientIP gets the client IP address from the request
func (h *FeedbackHandler) getClientIP(r *http.Request) string {
	// Check for forwarded headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

// getRequestID gets the request ID from the context
func (h *FeedbackHandler) getRequestID(ctx context.Context) string {
	// This is a simplified implementation
	// In production, use a proper request ID middleware
	return ""
}

// handleError handles errors and sends error responses
func (h *FeedbackHandler) handleError(w http.ResponseWriter, r *http.Request, message string, statusCode int, err error) {
	// Log error
	if h.logger != nil {
		h.logger.WithComponent("feedback_handler").LogBusinessEvent(r.Context(), "feedback_handler_error", "", map[string]interface{}{
			"error":       err.Error(),
			"status_code": statusCode,
			"message":     message,
			"method":      r.Method,
			"path":        r.URL.Path,
		})
	}

	// Record error metrics
	h.recordErrorMetrics(r.Context(), statusCode)

	// Send error response
	errorResponse := map[string]interface{}{
		"error":   message,
		"details": err.Error(),
	}
	h.sendJSONResponse(w, errorResponse, statusCode)
}

// sendJSONResponse sends a JSON response
func (h *FeedbackHandler) sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log encoding error
		if h.logger != nil {
			h.logger.WithComponent("feedback_handler").LogBusinessEvent(context.Background(), "json_encoding_error", "", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}
}

// recordFeedbackMetrics records feedback-related metrics
func (h *FeedbackHandler) recordFeedbackMetrics(ctx context.Context, operation string, feedbackType classification.FeedbackType, processingTime time.Duration) {
	if h.metrics == nil {
		return
	}

	h.metrics.RecordHistogram(ctx, "feedback_processing_time", float64(processingTime.Milliseconds()), map[string]string{
		"operation":     operation,
		"feedback_type": string(feedbackType),
	})

	h.metrics.RecordHistogram(ctx, "feedback_operations", 1.0, map[string]string{
		"operation":     operation,
		"feedback_type": string(feedbackType),
	})
}

// recordErrorMetrics records error metrics
func (h *FeedbackHandler) recordErrorMetrics(ctx context.Context, statusCode int) {
	if h.metrics == nil {
		return
	}

	h.metrics.RecordHistogram(ctx, "feedback_errors", 1.0, map[string]string{
		"status_code": fmt.Sprintf("%d", statusCode),
	})
}
