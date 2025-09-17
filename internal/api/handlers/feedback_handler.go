package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/feedback"
)

// FeedbackHandler handles feedback-related HTTP requests
type FeedbackHandler struct {
	feedbackService feedback.FeedbackService
	logger          *zap.Logger
}

// NewFeedbackHandler creates a new feedback handler
func NewFeedbackHandler(feedbackService feedback.FeedbackService, logger *zap.Logger) *FeedbackHandler {
	return &FeedbackHandler{
		feedbackService: feedbackService,
		logger:          logger,
	}
}

// CollectUserFeedback handles user feedback collection requests
func (h *FeedbackHandler) CollectUserFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	h.logger.Info("handling user feedback collection request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr))

	// Set CORS headers
	h.setCORSHeaders(w)

	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var request feedback.FeedbackCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("failed to decode request body",
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateFeedbackRequest(&request); err != nil {
		h.logger.Error("request validation failed",
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Collect feedback
	response, err := h.feedbackService.CollectUserFeedback(ctx, request)
	if err != nil {
		h.logger.Error("failed to collect user feedback",
			zap.String("user_id", request.UserID),
			zap.Error(err))
		http.Error(w, "Failed to collect feedback", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response",
			zap.Error(err))
		return
	}

	processingTime := time.Since(startTime)
	h.logger.Info("user feedback collection completed",
		zap.String("feedback_id", response.ID),
		zap.Duration("processing_time", processingTime))
}

// CollectMLModelFeedback handles ML model feedback collection requests
func (h *FeedbackHandler) CollectMLModelFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	h.logger.Info("handling ML model feedback collection request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	// Set CORS headers
	h.setCORSHeaders(w)

	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var mlFeedback feedback.MLModelFeedback
	if err := json.NewDecoder(r.Body).Decode(&mlFeedback); err != nil {
		h.logger.Error("failed to decode ML model feedback request body",
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Collect ML model feedback
	if err := h.feedbackService.CollectMLModelFeedback(ctx, mlFeedback); err != nil {
		h.logger.Error("failed to collect ML model feedback",
			zap.String("model_version_id", mlFeedback.ModelVersionID),
			zap.Error(err))
		http.Error(w, "Failed to collect ML model feedback", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Return success response
	response := map[string]interface{}{
		"status":      "success",
		"message":     "ML model feedback collected successfully",
		"feedback_id": mlFeedback.ID,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode ML model feedback response",
			zap.Error(err))
		return
	}

	processingTime := time.Since(startTime)
	h.logger.Info("ML model feedback collection completed",
		zap.String("feedback_id", mlFeedback.ID),
		zap.Duration("processing_time", processingTime))
}

// CollectSecurityValidationFeedback handles security validation feedback collection requests
func (h *FeedbackHandler) CollectSecurityValidationFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	h.logger.Info("handling security validation feedback collection request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	// Set CORS headers
	h.setCORSHeaders(w)

	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var securityFeedback feedback.SecurityValidationFeedback
	if err := json.NewDecoder(r.Body).Decode(&securityFeedback); err != nil {
		h.logger.Error("failed to decode security validation feedback request body",
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Collect security validation feedback
	if err := h.feedbackService.CollectSecurityValidationFeedback(ctx, securityFeedback); err != nil {
		h.logger.Error("failed to collect security validation feedback",
			zap.String("validation_type", securityFeedback.ValidationType),
			zap.Error(err))
		http.Error(w, "Failed to collect security validation feedback", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Return success response
	response := map[string]interface{}{
		"status":      "success",
		"message":     "Security validation feedback collected successfully",
		"feedback_id": securityFeedback.ID,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode security validation feedback response",
			zap.Error(err))
		return
	}

	processingTime := time.Since(startTime)
	h.logger.Info("security validation feedback collection completed",
		zap.String("feedback_id", securityFeedback.ID),
		zap.Duration("processing_time", processingTime))
}

// CollectBatchFeedback handles batch feedback collection requests
func (h *FeedbackHandler) CollectBatchFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	h.logger.Info("handling batch feedback collection request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	// Set CORS headers
	h.setCORSHeaders(w)

	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var requests []feedback.FeedbackCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		h.logger.Error("failed to decode batch feedback request body",
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate batch size
	if len(requests) > 100 {
		http.Error(w, "Batch size cannot exceed 100 requests", http.StatusBadRequest)
		return
	}

	// Collect batch feedback
	responses, err := h.feedbackService.CollectBatchFeedback(ctx, requests)
	if err != nil {
		h.logger.Error("failed to collect batch feedback",
			zap.Int("request_count", len(requests)),
			zap.Error(err))
		http.Error(w, "Failed to collect batch feedback", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Return batch response
	response := map[string]interface{}{
		"status":      "success",
		"message":     "Batch feedback collection completed",
		"responses":   responses,
		"total_count": len(responses),
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode batch feedback response",
			zap.Error(err))
		return
	}

	processingTime := time.Since(startTime)
	h.logger.Info("batch feedback collection completed",
		zap.Int("request_count", len(requests)),
		zap.Int("response_count", len(responses)),
		zap.Duration("processing_time", processingTime))
}

// AnalyzeFeedbackTrends handles feedback trends analysis requests
func (h *FeedbackHandler) AnalyzeFeedbackTrends(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	h.logger.Info("handling feedback trends analysis request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	// Set CORS headers
	h.setCORSHeaders(w)

	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	request := h.parseFeedbackAnalysisRequest(r)

	// Analyze feedback trends
	response, err := h.feedbackService.AnalyzeFeedbackTrends(ctx, request)
	if err != nil {
		h.logger.Error("failed to analyze feedback trends",
			zap.Error(err))
		http.Error(w, "Failed to analyze feedback trends", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode feedback trends response",
			zap.Error(err))
		return
	}

	processingTime := time.Since(startTime)
	h.logger.Info("feedback trends analysis completed",
		zap.Int("trend_count", len(response.Trends)),
		zap.Duration("processing_time", processingTime))
}

// GetFeedbackInsights handles feedback insights requests
func (h *FeedbackHandler) GetFeedbackInsights(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	h.logger.Info("handling feedback insights request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	// Set CORS headers
	h.setCORSHeaders(w)

	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse method from query parameters
	methodStr := r.URL.Query().Get("method")
	if methodStr == "" {
		http.Error(w, "Method parameter is required", http.StatusBadRequest)
		return
	}

	method := feedback.ClassificationMethod(methodStr)

	// Generate feedback insights
	insights, err := h.feedbackService.GenerateFeedbackInsights(ctx, method)
	if err != nil {
		h.logger.Error("failed to generate feedback insights",
			zap.String("method", string(method)),
			zap.Error(err))
		http.Error(w, "Failed to generate feedback insights", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode response
	if err := json.NewEncoder(w).Encode(insights); err != nil {
		h.logger.Error("failed to encode feedback insights response",
			zap.Error(err))
		return
	}

	processingTime := time.Since(startTime)
	h.logger.Info("feedback insights generation completed",
		zap.String("method", string(method)),
		zap.Duration("processing_time", processingTime))
}

// GetServiceHealth handles service health check requests
func (h *FeedbackHandler) GetServiceHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.Info("handling service health check request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	// Set CORS headers
	h.setCORSHeaders(w)

	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get service health
	health, err := h.feedbackService.GetServiceHealth(ctx)
	if err != nil {
		h.logger.Error("failed to get service health",
			zap.Error(err))
		http.Error(w, "Failed to get service health", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode response
	if err := json.NewEncoder(w).Encode(health); err != nil {
		h.logger.Error("failed to encode service health response",
			zap.Error(err))
		return
	}
}

// GetServiceMetrics handles service metrics requests
func (h *FeedbackHandler) GetServiceMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.Info("handling service metrics request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	// Set CORS headers
	h.setCORSHeaders(w)

	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get service metrics
	metrics, err := h.feedbackService.GetServiceMetrics(ctx)
	if err != nil {
		h.logger.Error("failed to get service metrics",
			zap.Error(err))
		http.Error(w, "Failed to get service metrics", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode response
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		h.logger.Error("failed to encode service metrics response",
			zap.Error(err))
		return
	}
}

// Helper methods

// setCORSHeaders sets CORS headers for cross-origin requests
func (h *FeedbackHandler) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// validateFeedbackRequest validates a feedback collection request
func (h *FeedbackHandler) validateFeedbackRequest(request *feedback.FeedbackCollectionRequest) error {
	if request.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if request.BusinessName == "" {
		return fmt.Errorf("business name is required")
	}

	if request.OriginalClassificationID == "" {
		return fmt.Errorf("original classification ID is required")
	}

	if request.FeedbackType == "" {
		return fmt.Errorf("feedback type is required")
	}

	return nil
}

// parseFeedbackAnalysisRequest parses query parameters into a feedback analysis request
func (h *FeedbackHandler) parseFeedbackAnalysisRequest(r *http.Request) feedback.FeedbackAnalysisRequest {
	request := feedback.FeedbackAnalysisRequest{}

	if method := r.URL.Query().Get("method"); method != "" {
		request.Method = feedback.ClassificationMethod(method)
	}

	if timeWindow := r.URL.Query().Get("time_window"); timeWindow != "" {
		request.TimeWindow = timeWindow
	}

	if feedbackType := r.URL.Query().Get("feedback_type"); feedbackType != "" {
		request.FeedbackType = feedback.FeedbackType(feedbackType)
	}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			request.StartDate = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			request.EndDate = &endDate
		}
	}

	return request
}

// HandleOptions handles CORS preflight requests
func (h *FeedbackHandler) HandleOptions(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w)
	w.WriteHeader(http.StatusOK)
}
