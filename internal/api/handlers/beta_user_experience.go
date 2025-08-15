package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/pcraw4d/business-verification/internal/webanalysis.bak"
	"go.uber.org/zap"
	// "github.com/pcraw4d/business-verification/internal/webanalysis" // Temporarily disabled
)

// BetaUserExperienceHandler handles beta user experience features
type BetaUserExperienceHandler struct {
	logger        *zap.Logger
	betaFramework *webanalysis.BetaTestingFramework
}

// NewBetaUserExperienceHandler creates a new beta user experience handler
func NewBetaUserExperienceHandler(logger *zap.Logger, betaFramework *webanalysis.BetaTestingFramework) *BetaUserExperienceHandler {
	return &BetaUserExperienceHandler{
		logger:        logger,
		betaFramework: betaFramework,
	}
}

// ScrapingMethodSelectionRequest represents a request to select scraping method
type ScrapingMethodSelectionRequest struct {
	UserID          string `json:"user_id"`
	URL             string `json:"url"`
	PreferredMethod string `json:"preferred_method"` // "basic", "enhanced", "auto"
	ForceMethod     bool   `json:"force_method"`     // If true, use preferred method regardless of beta status
}

// ScrapingMethodSelectionResponse represents the response for method selection
type ScrapingMethodSelectionResponse struct {
	SelectedMethod string                `json:"selected_method"`
	MethodUsed     string                `json:"method_used"`
	IsBetaUser     bool                  `json:"is_beta_user"`
	Transparency   *ScrapingTransparency `json:"transparency"`
	Message        string                `json:"message"`
}

// ScrapingTransparency provides transparency about the scraping process
type ScrapingTransparency struct {
	MethodUsed         string              `json:"method_used"`
	ReasonForSelection string              `json:"reason_for_selection"`
	BetaTestID         string              `json:"beta_test_id,omitempty"`
	PerformanceMetrics *PerformanceMetrics `json:"performance_metrics,omitempty"`
	Timestamp          time.Time           `json:"timestamp"`
}

// PerformanceMetrics shows performance comparison
type PerformanceMetrics struct {
	EnhancedSuccessRate float64 `json:"enhanced_success_rate"`
	BasicSuccessRate    float64 `json:"basic_success_rate"`
	EnhancedAvgTime     float64 `json:"enhanced_avg_time"`
	BasicAvgTime        float64 `json:"basic_avg_time"`
	Improvement         float64 `json:"improvement"`
}

// UserFeedbackRequest represents user feedback submission
type UserFeedbackRequest struct {
	UserID       string `json:"user_id"`
	TestID       string `json:"test_id"`
	URL          string `json:"url"`
	Method       string `json:"method"`
	Satisfaction int    `json:"satisfaction"` // 1-5 scale
	Accuracy     int    `json:"accuracy"`     // 1-5 scale
	Speed        int    `json:"speed"`        // 1-5 scale
	Comments     string `json:"comments"`
}

// UserPreferenceRequest represents user preference update
type UserPreferenceRequest struct {
	UserID             string `json:"user_id"`
	DefaultMethod      string `json:"default_method"` // "basic", "enhanced", "auto"
	EnableBeta         bool   `json:"enable_beta"`
	EnableTransparency bool   `json:"enable_transparency"`
	EnableFeedback     bool   `json:"enable_feedback"`
}

// UserPreferenceResponse represents user preferences
type UserPreferenceResponse struct {
	UserID             string    `json:"user_id"`
	DefaultMethod      string    `json:"default_method"`
	EnableBeta         bool      `json:"enable_beta"`
	EnableTransparency bool      `json:"enable_transparency"`
	EnableFeedback     bool      `json:"enable_feedback"`
	LastUpdated        time.Time `json:"last_updated"`
}

// RegisterRoutes registers the beta user experience routes
func (h *BetaUserExperienceHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/v2/beta/scraping-method", h.SelectScrapingMethod).Methods("POST")
	router.HandleFunc("/v2/beta/feedback", h.SubmitFeedback).Methods("POST")
	router.HandleFunc("/v2/beta/preferences", h.UpdatePreferences).Methods("PUT")
	router.HandleFunc("/v2/beta/preferences/{user_id}", h.GetPreferences).Methods("GET")
	router.HandleFunc("/v2/beta/transparency/{test_id}", h.GetTransparency).Methods("GET")
	router.HandleFunc("/v2/beta/performance-comparison", h.GetPerformanceComparison).Methods("GET")
	router.HandleFunc("/v2/beta/user-stats/{user_id}", h.GetUserStats).Methods("GET")
}

// SelectScrapingMethod handles scraping method selection
func (h *BetaUserExperienceHandler) SelectScrapingMethod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ScrapingMethodSelectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.UserID == "" || req.URL == "" {
		http.Error(w, "User ID and URL are required", http.StatusBadRequest)
		return
	}

	// Determine if user is a beta user
	isBetaUser := h.isBetaUser(req.UserID)

	// Select method based on preferences and beta status
	selectedMethod, reason := h.selectMethod(req, isBetaUser)

	// Create transparency information
	transparency := &ScrapingTransparency{
		MethodUsed:         selectedMethod,
		ReasonForSelection: reason,
		Timestamp:          time.Now(),
	}

	// If this is a beta test, generate test ID and add performance metrics
	if isBetaUser && selectedMethod == "enhanced" {
		testID := h.generateTestID(req.UserID, req.URL)
		transparency.BetaTestID = testID

		// Get performance comparison metrics
		comparison, err := h.betaFramework.GetPerformanceComparison(ctx, 24*time.Hour)
		if err == nil && comparison != nil {
			transparency.PerformanceMetrics = &PerformanceMetrics{
				EnhancedSuccessRate: comparison.EnhancedMetrics.SuccessRate,
				BasicSuccessRate:    comparison.BasicMetrics.SuccessRate,
				EnhancedAvgTime:     float64(comparison.EnhancedMetrics.AverageResponseTime.Milliseconds()) / 1000.0,
				BasicAvgTime:        float64(comparison.BasicMetrics.AverageResponseTime.Milliseconds()) / 1000.0,
				Improvement:         comparison.SuccessRateImprovement,
			}
		}
	}

	response := ScrapingMethodSelectionResponse{
		SelectedMethod: selectedMethod,
		MethodUsed:     selectedMethod,
		IsBetaUser:     isBetaUser,
		Transparency:   transparency,
		Message:        fmt.Sprintf("Using %s method: %s", selectedMethod, reason),
	}

	h.logger.Info("Scraping method selected",
		zap.String("user_id", req.UserID),
		zap.String("url", req.URL),
		zap.String("selected_method", selectedMethod),
		zap.String("reason", reason),
		zap.Bool("is_beta_user", isBetaUser),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SubmitFeedback handles user feedback submission
func (h *BetaUserExperienceHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req UserFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode feedback request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.UserID == "" || req.TestID == "" || req.URL == "" || req.Method == "" {
		http.Error(w, "User ID, Test ID, URL, and Method are required", http.StatusBadRequest)
		return
	}

	if req.Satisfaction < 1 || req.Satisfaction > 5 {
		http.Error(w, "Satisfaction must be between 1 and 5", http.StatusBadRequest)
		return
	}

	if req.Accuracy < 1 || req.Accuracy > 5 {
		http.Error(w, "Accuracy must be between 1 and 5", http.StatusBadRequest)
		return
	}

	if req.Speed < 1 || req.Speed > 5 {
		http.Error(w, "Speed must be between 1 and 5", http.StatusBadRequest)
		return
	}

	// Create feedback object
	feedback := &webanalysis.BetaFeedback{
		UserID:       req.UserID,
		TestID:       req.TestID,
		URL:          req.URL,
		Method:       req.Method,
		Satisfaction: req.Satisfaction,
		Accuracy:     req.Accuracy,
		Speed:        req.Speed,
		Comments:     req.Comments,
	}

	// Store feedback
	err := h.betaFramework.CollectUserFeedback(ctx, feedback)
	if err != nil {
		h.logger.Error("Failed to store feedback", zap.Error(err))
		http.Error(w, "Failed to store feedback", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User feedback submitted",
		zap.String("user_id", req.UserID),
		zap.String("test_id", req.TestID),
		zap.String("method", req.Method),
		zap.Int("satisfaction", req.Satisfaction),
		zap.Int("accuracy", req.Accuracy),
		zap.Int("speed", req.Speed),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Feedback submitted successfully",
		"status":  "success",
	})
}

// UpdatePreferences handles user preference updates
func (h *BetaUserExperienceHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	var req UserPreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode preferences request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.UserID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	if req.DefaultMethod != "" && req.DefaultMethod != "basic" && req.DefaultMethod != "enhanced" && req.DefaultMethod != "auto" {
		http.Error(w, "Default method must be 'basic', 'enhanced', or 'auto'", http.StatusBadRequest)
		return
	}

	// Store user preferences (in a real implementation, this would be stored in a database)
	preferences := &UserPreferenceResponse{
		UserID:             req.UserID,
		DefaultMethod:      req.DefaultMethod,
		EnableBeta:         req.EnableBeta,
		EnableTransparency: req.EnableTransparency,
		EnableFeedback:     req.EnableFeedback,
		LastUpdated:        time.Now(),
	}

	h.logger.Info("User preferences updated",
		zap.String("user_id", req.UserID),
		zap.String("default_method", req.DefaultMethod),
		zap.Bool("enable_beta", req.EnableBeta),
		zap.Bool("enable_transparency", req.EnableTransparency),
		zap.Bool("enable_feedback", req.EnableFeedback),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preferences)
}

// GetPreferences retrieves user preferences
func (h *BetaUserExperienceHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// In a real implementation, this would retrieve from a database
	// For now, return default preferences
	preferences := &UserPreferenceResponse{
		UserID:             userID,
		DefaultMethod:      "auto",
		EnableBeta:         true,
		EnableTransparency: true,
		EnableFeedback:     true,
		LastUpdated:        time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preferences)
}

// GetTransparency retrieves transparency information for a test
func (h *BetaUserExperienceHandler) GetTransparency(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	testID := vars["test_id"]

	if testID == "" {
		http.Error(w, "Test ID is required", http.StatusBadRequest)
		return
	}

	// Get A/B test analysis
	analysis, err := h.betaFramework.abTestManager.GetAnalysis(ctx, testID)
	if err != nil {
		h.logger.Error("Failed to get test analysis", zap.Error(err))
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}

	// Create transparency response
	transparency := &ScrapingTransparency{
		MethodUsed:         "enhanced",
		ReasonForSelection: "Beta test participant",
		BetaTestID:         testID,
		PerformanceMetrics: &PerformanceMetrics{
			EnhancedSuccessRate: analysis.EnhancedSuccessRate,
			BasicSuccessRate:    analysis.BasicSuccessRate,
			EnhancedAvgTime:     float64(analysis.EnhancedAvgResponseTime.Milliseconds()) / 1000.0,
			BasicAvgTime:        float64(analysis.BasicAvgResponseTime.Milliseconds()) / 1000.0,
			Improvement:         analysis.ImprovementMetrics.SuccessRateImprovement,
		},
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transparency)
}

// GetPerformanceComparison retrieves performance comparison data
func (h *BetaUserExperienceHandler) GetPerformanceComparison(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse time range from query parameters
	timeRangeStr := r.URL.Query().Get("time_range")
	timeRange := 24 * time.Hour // Default to 24 hours

	if timeRangeStr != "" {
		if hours, err := strconv.Atoi(timeRangeStr); err == nil {
			timeRange = time.Duration(hours) * time.Hour
		}
	}

	comparison, err := h.betaFramework.GetPerformanceComparison(ctx, timeRange)
	if err != nil {
		h.logger.Error("Failed to get performance comparison", zap.Error(err))
		http.Error(w, "Failed to get performance comparison", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comparison)
}

// GetUserStats retrieves user statistics
func (h *BetaUserExperienceHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	userID := vars["user_id"]

	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get user feedback
	feedback, err := h.betaFramework.feedbackCollector.GetUserFeedback(userID)
	if err != nil {
		h.logger.Error("Failed to get user feedback", zap.Error(err))
		http.Error(w, "Failed to get user stats", http.StatusInternalServerError)
		return
	}

	// Calculate user statistics
	stats := h.calculateUserStats(feedback)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Helper methods

func (h *BetaUserExperienceHandler) isBetaUser(userID string) bool {
	// In a real implementation, this would check against a beta user database
	// For now, consider all users as potential beta users
	return true
}

func (h *BetaUserExperienceHandler) selectMethod(req ScrapingMethodSelectionRequest, isBetaUser bool) (string, string) {
	// If user forces a specific method, use it
	if req.ForceMethod {
		return req.PreferredMethod, "User preference forced"
	}

	// If user is not a beta user, use basic method
	if !isBetaUser {
		return "basic", "User not in beta program"
	}

	// If user has a preferred method, use it
	if req.PreferredMethod != "" && req.PreferredMethod != "auto" {
		return req.PreferredMethod, "User preference"
	}

	// For beta users with auto selection, use enhanced method
	return "enhanced", "Beta user with auto selection"
}

func (h *BetaUserExperienceHandler) generateTestID(userID, url string) string {
	return fmt.Sprintf("beta_test_%s_%d", userID, time.Now().Unix())
}

func (h *BetaUserExperienceHandler) calculateUserStats(feedback []*webanalysis.BetaFeedback) map[string]interface{} {
	if len(feedback) == 0 {
		return map[string]interface{}{
			"total_feedback":       0,
			"average_satisfaction": 0.0,
			"average_accuracy":     0.0,
			"average_speed":        0.0,
		}
	}

	var totalSatisfaction, totalAccuracy, totalSpeed int
	var satisfactionCount, accuracyCount, speedCount int

	for _, f := range feedback {
		if f.Satisfaction > 0 {
			totalSatisfaction += f.Satisfaction
			satisfactionCount++
		}
		if f.Accuracy > 0 {
			totalAccuracy += f.Accuracy
			accuracyCount++
		}
		if f.Speed > 0 {
			totalSpeed += f.Speed
			speedCount++
		}
	}

	stats := map[string]interface{}{
		"total_feedback": len(feedback),
	}

	if satisfactionCount > 0 {
		stats["average_satisfaction"] = float64(totalSatisfaction) / float64(satisfactionCount)
	} else {
		stats["average_satisfaction"] = 0.0
	}

	if accuracyCount > 0 {
		stats["average_accuracy"] = float64(totalAccuracy) / float64(accuracyCount)
	} else {
		stats["average_accuracy"] = 0.0
	}

	if speedCount > 0 {
		stats["average_speed"] = float64(totalSpeed) / float64(speedCount)
	} else {
		stats["average_speed"] = 0.0
	}

	return stats
}
