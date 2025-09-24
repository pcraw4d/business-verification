package routes

import (
	"net/http"

	"go.uber.org/zap"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/feedback"
)

// RegisterFeedbackRoutes registers all feedback-related routes
func RegisterFeedbackRoutes(
	mux *http.ServeMux,
	feedbackService feedback.FeedbackService,
	logger *zap.Logger,
) {
	// Create the feedback handler
	feedbackHandler := handlers.NewFeedbackHandler(feedbackService, logger)

	// User feedback collection endpoints
	mux.HandleFunc("POST /v1/feedback/user", feedbackHandler.CollectUserFeedback)
	mux.HandleFunc("OPTIONS /v1/feedback/user", feedbackHandler.HandleOptions)

	// ML model feedback collection endpoints
	mux.HandleFunc("POST /v1/feedback/ml-model", feedbackHandler.CollectMLModelFeedback)
	mux.HandleFunc("OPTIONS /v1/feedback/ml-model", feedbackHandler.HandleOptions)

	// Security validation feedback collection endpoints
	mux.HandleFunc("POST /v1/feedback/security", feedbackHandler.CollectSecurityValidationFeedback)
	mux.HandleFunc("OPTIONS /v1/feedback/security", feedbackHandler.HandleOptions)

	// Batch feedback collection endpoints
	mux.HandleFunc("POST /v1/feedback/batch", feedbackHandler.CollectBatchFeedback)
	mux.HandleFunc("OPTIONS /v1/feedback/batch", feedbackHandler.HandleOptions)

	// Feedback analysis endpoints
	mux.HandleFunc("GET /v1/feedback/trends", feedbackHandler.AnalyzeFeedbackTrends)
	mux.HandleFunc("GET /v1/feedback/insights", feedbackHandler.GetFeedbackInsights)

	// Service management endpoints
	mux.HandleFunc("GET /v1/feedback/health", feedbackHandler.GetServiceHealth)
	mux.HandleFunc("GET /v1/feedback/metrics", feedbackHandler.GetServiceMetrics)

	logger.Info("Feedback routes registered", map[string]interface{}{
		"version": "v1",
		"endpoints": []string{
			"POST /v1/feedback/user",
			"POST /v1/feedback/ml-model",
			"POST /v1/feedback/security",
			"POST /v1/feedback/batch",
			"GET /v1/feedback/trends",
			"GET /v1/feedback/insights",
			"GET /v1/feedback/health",
			"GET /v1/feedback/metrics",
		},
	})
}

// RegisterFeedbackRoutesV2 registers feedback routes with enhanced features
func RegisterFeedbackRoutesV2(
	mux *http.ServeMux,
	feedbackService feedback.FeedbackService,
	logger *zap.Logger,
) {
	// Create the feedback handler
	feedbackHandler := handlers.NewFeedbackHandler(feedbackService, logger)

	// Enhanced user feedback collection endpoints with additional features
	mux.HandleFunc("POST /v2/feedback/user", feedbackHandler.CollectUserFeedback)
	mux.HandleFunc("OPTIONS /v2/feedback/user", feedbackHandler.HandleOptions)

	// Enhanced ML model feedback collection endpoints
	mux.HandleFunc("POST /v2/feedback/ml-model", feedbackHandler.CollectMLModelFeedback)
	mux.HandleFunc("OPTIONS /v2/feedback/ml-model", feedbackHandler.HandleOptions)

	// Enhanced security validation feedback collection endpoints
	mux.HandleFunc("POST /v2/feedback/security", feedbackHandler.CollectSecurityValidationFeedback)
	mux.HandleFunc("OPTIONS /v2/feedback/security", feedbackHandler.HandleOptions)

	// Enhanced batch feedback collection endpoints
	mux.HandleFunc("POST /v2/feedback/batch", feedbackHandler.CollectBatchFeedback)
	mux.HandleFunc("OPTIONS /v2/feedback/batch", feedbackHandler.HandleOptions)

	// Enhanced feedback analysis endpoints
	mux.HandleFunc("GET /v2/feedback/trends", feedbackHandler.AnalyzeFeedbackTrends)
	mux.HandleFunc("GET /v2/feedback/insights", feedbackHandler.GetFeedbackInsights)

	// Enhanced service management endpoints
	mux.HandleFunc("GET /v2/feedback/health", feedbackHandler.GetServiceHealth)
	mux.HandleFunc("GET /v2/feedback/metrics", feedbackHandler.GetServiceMetrics)

	logger.Info("Enhanced feedback routes registered", map[string]interface{}{
		"version": "v2",
		"endpoints": []string{
			"POST /v2/feedback/user",
			"POST /v2/feedback/ml-model",
			"POST /v2/feedback/security",
			"POST /v2/feedback/batch",
			"GET /v2/feedback/trends",
			"GET /v2/feedback/insights",
			"GET /v2/feedback/health",
			"GET /v2/feedback/metrics",
		},
		"features": []string{
			"enhanced_security_validation",
			"model_versioning",
			"trend_analysis",
			"batch_processing",
			"real_time_insights",
		},
	})
}
