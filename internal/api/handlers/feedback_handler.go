package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/petercrawford/kyb-platform/internal/feedback"
)

// FeedbackHandler handles HTTP requests for user feedback collection
type FeedbackHandler struct {
	collector *feedback.UserFeedbackCollector
	logger    *log.Logger
}

// NewFeedbackHandler creates a new feedback handler
func NewFeedbackHandler(collector *feedback.UserFeedbackCollector, logger *log.Logger) *FeedbackHandler {
	return &FeedbackHandler{
		collector: collector,
		logger:    logger,
	}
}

// FeedbackRequest represents the request payload for feedback submission
type FeedbackRequest struct {
	UserID                 string                        `json:"user_id" validate:"required"`
	Category               feedback.FeedbackCategory     `json:"category" validate:"required"`
	Rating                 int                           `json:"rating" validate:"required,min=1,max=5"`
	Comments               string                        `json:"comments"`
	SpecificFeatures       []string                      `json:"specific_features"`
	ImprovementAreas       []string                      `json:"improvement_areas"`
	ClassificationAccuracy float64                       `json:"classification_accuracy" validate:"min=0,max=1"`
	PerformanceRating      int                           `json:"performance_rating" validate:"required,min=1,max=5"`
	UsabilityRating        int                           `json:"usability_rating" validate:"required,min=1,max=5"`
	BusinessImpact         feedback.BusinessImpactRating `json:"business_impact"`
}

// FeedbackResponse represents the response for feedback submission
type FeedbackResponse struct {
	Success    bool      `json:"success"`
	Message    string    `json:"message"`
	FeedbackID string    `json:"feedback_id,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// FeedbackAnalysisResponse represents the response for feedback analysis
type FeedbackAnalysisResponse struct {
	Success   bool                       `json:"success"`
	Analysis  *feedback.FeedbackAnalysis `json:"analysis,omitempty"`
	Message   string                     `json:"message,omitempty"`
	Timestamp time.Time                  `json:"timestamp"`
}

// FeedbackStatsResponse represents the response for feedback statistics
type FeedbackStatsResponse struct {
	Success   bool                    `json:"success"`
	Stats     *feedback.FeedbackStats `json:"stats,omitempty"`
	Message   string                  `json:"message,omitempty"`
	Timestamp time.Time               `json:"timestamp"`
}

// SubmitFeedback handles POST /api/feedback/submit
func (fh *FeedbackHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var req FeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fh.logger.Printf("Failed to decode feedback request: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := fh.validateFeedbackRequest(&req); err != nil {
		fh.logger.Printf("Feedback request validation failed: %v", err)
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	// Convert to UserFeedback
	userFeedback := &feedback.UserFeedback{
		UserID:                 req.UserID,
		Category:               req.Category,
		Rating:                 req.Rating,
		Comments:               req.Comments,
		SpecificFeatures:       req.SpecificFeatures,
		ImprovementAreas:       req.ImprovementAreas,
		ClassificationAccuracy: req.ClassificationAccuracy,
		PerformanceRating:      req.PerformanceRating,
		UsabilityRating:        req.UsabilityRating,
		BusinessImpact:         req.BusinessImpact,
	}

	// Collect feedback
	if err := fh.collector.CollectFeedback(ctx, userFeedback); err != nil {
		fh.logger.Printf("Failed to collect feedback: %v", err)
		http.Error(w, "Failed to submit feedback", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := FeedbackResponse{
		Success:    true,
		Message:    "Feedback submitted successfully",
		FeedbackID: userFeedback.ID.String(),
		Timestamp:  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	fh.logger.Printf("Feedback submitted successfully: ID=%s, User=%s, Category=%s",
		userFeedback.ID, req.UserID, req.Category)
}

// GetFeedbackAnalysis handles GET /api/feedback/analysis/{category}
func (fh *FeedbackHandler) GetFeedbackAnalysis(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get category from URL
	vars := mux.Vars(r)
	category := feedback.FeedbackCategory(vars["category"])

	// Validate category
	if !fh.isValidCategory(category) {
		http.Error(w, "Invalid feedback category", http.StatusBadRequest)
		return
	}

	// Get analysis
	analysis, err := fh.collector.GetFeedbackAnalysis(ctx, category)
	if err != nil {
		fh.logger.Printf("Failed to get feedback analysis: %v", err)
		http.Error(w, "Failed to retrieve feedback analysis", http.StatusInternalServerError)
		return
	}

	// Return analysis response
	response := FeedbackAnalysisResponse{
		Success:   true,
		Analysis:  analysis,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetFeedbackStats handles GET /api/feedback/stats
func (fh *FeedbackHandler) GetFeedbackStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get statistics
	stats, err := fh.collector.GetFeedbackStats(ctx)
	if err != nil {
		fh.logger.Printf("Failed to get feedback stats: %v", err)
		http.Error(w, "Failed to retrieve feedback statistics", http.StatusInternalServerError)
		return
	}

	// Return stats response
	response := FeedbackStatsResponse{
		Success:   true,
		Stats:     stats,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExportFeedback handles GET /api/feedback/export/{category}?format={json|csv}
func (fh *FeedbackHandler) ExportFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get category from URL
	vars := mux.Vars(r)
	category := feedback.FeedbackCategory(vars["category"])

	// Validate category
	if !fh.isValidCategory(category) {
		http.Error(w, "Invalid feedback category", http.StatusBadRequest)
		return
	}

	// Get format from query parameter
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json" // Default format
	}

	// Validate format
	if format != "json" && format != "csv" {
		http.Error(w, "Invalid export format. Supported formats: json, csv", http.StatusBadRequest)
		return
	}

	// Export feedback
	data, err := fh.collector.ExportFeedback(ctx, format, category)
	if err != nil {
		fh.logger.Printf("Failed to export feedback: %v", err)
		http.Error(w, "Failed to export feedback", http.StatusInternalServerError)
		return
	}

	// Set appropriate headers
	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=feedback_%s.csv", category))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=feedback_%s.json", category))
	}

	w.Write(data)
}

// GetFeedbackByTimeRange handles GET /api/feedback/range?start={timestamp}&end={timestamp}
func (fh *FeedbackHandler) GetFeedbackByTimeRange(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse time range parameters
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr == "" || endStr == "" {
		http.Error(w, "Both start and end timestamps are required", http.StatusBadRequest)
		return
	}

	// Parse timestamps
	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		http.Error(w, "Invalid start timestamp format. Use RFC3339 format", http.StatusBadRequest)
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		http.Error(w, "Invalid end timestamp format. Use RFC3339 format", http.StatusBadRequest)
		return
	}

	// Validate time range
	if start.After(end) {
		http.Error(w, "Start time must be before end time", http.StatusBadRequest)
		return
	}

	// Get feedback by time range
	feedbackData, err := fh.collector.GetFeedbackByTimeRange(ctx, start, end)
	if err != nil {
		fh.logger.Printf("Failed to get feedback by time range: %v", err)
		http.Error(w, "Failed to retrieve feedback", http.StatusInternalServerError)
		return
	}

	// Return feedback data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"feedback":  feedbackData,
		"count":     len(feedbackData),
		"timestamp": time.Now(),
	})
}

// validateFeedbackRequest validates the feedback request
func (fh *FeedbackHandler) validateFeedbackRequest(req *FeedbackRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if req.Category == "" {
		return fmt.Errorf("category is required")
	}

	if req.Rating < 1 || req.Rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}

	if req.PerformanceRating < 1 || req.PerformanceRating > 5 {
		return fmt.Errorf("performance rating must be between 1 and 5")
	}

	if req.UsabilityRating < 1 || req.UsabilityRating > 5 {
		return fmt.Errorf("usability rating must be between 1 and 5")
	}

	if req.ClassificationAccuracy < 0 || req.ClassificationAccuracy > 1 {
		return fmt.Errorf("classification accuracy must be between 0 and 1")
	}

	return nil
}

// isValidCategory checks if the category is valid
func (fh *FeedbackHandler) isValidCategory(category feedback.FeedbackCategory) bool {
	validCategories := []feedback.FeedbackCategory{
		feedback.CategoryDatabasePerformance,
		feedback.CategoryClassificationAccuracy,
		feedback.CategoryUserExperience,
		feedback.CategoryRiskDetection,
		feedback.CategoryOverallSatisfaction,
		feedback.CategoryFeatureRequest,
		feedback.CategoryBugReport,
	}

	for _, valid := range validCategories {
		if category == valid {
			return true
		}
	}
	return false
}

// RegisterFeedbackRoutes registers all feedback-related routes
func RegisterFeedbackRoutes(router *mux.Router, handler *FeedbackHandler) {
	// Feedback submission
	router.HandleFunc("/api/feedback/submit", handler.SubmitFeedback).Methods("POST")

	// Feedback analysis
	router.HandleFunc("/api/feedback/analysis/{category}", handler.GetFeedbackAnalysis).Methods("GET")

	// Feedback statistics
	router.HandleFunc("/api/feedback/stats", handler.GetFeedbackStats).Methods("GET")

	// Feedback export
	router.HandleFunc("/api/feedback/export/{category}", handler.ExportFeedback).Methods("GET")

	// Feedback by time range
	router.HandleFunc("/api/feedback/range", handler.GetFeedbackByTimeRange).Methods("GET")
}
