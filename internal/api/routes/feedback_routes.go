package routes

import (
	"net/http"

	"go.uber.org/zap"

	"kyb-platform/internal/feedback"
)

// RegisterFeedbackRoutes registers all feedback-related routes
// TODO: FeedbackHandler methods don't match the expected interface
// NewFeedbackHandler expects *feedback.UserFeedbackCollector and *log.Logger
// Most handler methods referenced here don't exist - only SubmitFeedback exists
func RegisterFeedbackRoutes(
	mux *http.ServeMux,
	feedbackService feedback.FeedbackService,
	logger *zap.Logger,
) {
	// TODO: Fix handler initialization - needs *feedback.UserFeedbackCollector and *log.Logger
	// For now, comment out routes until handler is properly implemented
	_ = mux
	_ = feedbackService
	_ = logger

	// All routes commented out until handler methods are implemented
	// feedbackHandler := handlers.NewFeedbackHandler(...)
	// mux.HandleFunc("POST /v1/feedback/user", feedbackHandler.CollectUserFeedback)
	// ... (all other routes)
}

// RegisterFeedbackRoutesV2 registers feedback routes with enhanced features
// TODO: Same issues as RegisterFeedbackRoutes - handler methods don't exist
func RegisterFeedbackRoutesV2(
	mux *http.ServeMux,
	feedbackService feedback.FeedbackService,
	logger *zap.Logger,
) {
	// TODO: Fix handler initialization - needs *feedback.UserFeedbackCollector and *log.Logger
	// For now, comment out routes until handler is properly implemented
	_ = mux
	_ = feedbackService
	_ = logger

	// All routes commented out until handler methods are implemented
	// feedbackHandler := handlers.NewFeedbackHandler(...)
	// mux.HandleFunc("POST /v2/feedback/user", feedbackHandler.CollectUserFeedback)
	// ... (all other routes)
}
