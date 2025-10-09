package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// FeedbackCollector handles beta testing feedback collection
type FeedbackCollector struct {
	logger *zap.Logger
	store  FeedbackStore
}

// FeedbackStore interface for storing feedback
type FeedbackStore interface {
	StoreFeedback(feedback *BetaFeedback) error
	GetFeedback(betaTesterID string) ([]*BetaFeedback, error)
	GetAllFeedback() ([]*BetaFeedback, error)
	GetFeedbackStats() (*FeedbackStats, error)
}

// InMemoryFeedbackStore implements FeedbackStore using in-memory storage
type InMemoryFeedbackStore struct {
	feedback []*BetaFeedback
}

// BetaFeedback represents feedback from a beta tester
type BetaFeedback struct {
	ID                 string            `json:"id"`
	BetaTesterID       string            `json:"beta_tester_id"`
	BetaTesterName     string            `json:"beta_tester_name"`
	Company            string            `json:"company"`
	Email              string            `json:"email"`
	SubmittedAt        time.Time         `json:"submitted_at"`
	OverallRating      int               `json:"overall_rating"` // 1-5
	APIDesign          *CategoryFeedback `json:"api_design"`
	Performance        *CategoryFeedback `json:"performance"`
	DeveloperExp       *CategoryFeedback `json:"developer_experience"`
	Features           *CategoryFeedback `json:"features"`
	BugsFound          []BugReport       `json:"bugs_found"`
	FeatureRequests    []FeatureRequest  `json:"feature_requests"`
	AdditionalComments string            `json:"additional_comments"`
	TestScenarios      []TestScenario    `json:"test_scenarios"`
	SDKUsed            string            `json:"sdk_used"`
	IntegrationType    string            `json:"integration_type"`
}

// CategoryFeedback represents feedback for a specific category
type CategoryFeedback struct {
	Rating   int    `json:"rating"` // 1-5
	Comments string `json:"comments"`
}

// BugReport represents a bug found during testing
type BugReport struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"` // low, medium, high, critical
	Steps       []string  `json:"steps"`
	Expected    string    `json:"expected"`
	Actual      string    `json:"actual"`
	Environment string    `json:"environment"`
	ReportedAt  time.Time `json:"reported_at"`
}

// FeatureRequest represents a feature request from a beta tester
type FeatureRequest struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    string    `json:"priority"` // low, medium, high
	UseCase     string    `json:"use_case"`
	RequestedAt time.Time `json:"requested_at"`
}

// TestScenario represents a test scenario that was executed
type TestScenario struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`   // passed, failed, partial
	Duration    int64     `json:"duration"` // milliseconds
	Notes       string    `json:"notes"`
	ExecutedAt  time.Time `json:"executed_at"`
}

// FeedbackStats represents aggregated feedback statistics
type FeedbackStats struct {
	TotalFeedback       int                `json:"total_feedback"`
	AverageRating       float64            `json:"average_rating"`
	CategoryRatings     map[string]float64 `json:"category_ratings"`
	BugCount            int                `json:"bug_count"`
	FeatureRequestCount int                `json:"feature_request_count"`
	SDKUsage            map[string]int     `json:"sdk_usage"`
	IntegrationTypes    map[string]int     `json:"integration_types"`
	TestScenarioStats   map[string]int     `json:"test_scenario_stats"`
	RecentFeedback      []*BetaFeedback    `json:"recent_feedback"`
}

// NewFeedbackCollector creates a new feedback collector
func NewFeedbackCollector(logger *zap.Logger) *FeedbackCollector {
	return &FeedbackCollector{
		logger: logger,
		store:  NewInMemoryFeedbackStore(),
	}
}

// NewInMemoryFeedbackStore creates a new in-memory feedback store
func NewInMemoryFeedbackStore() *InMemoryFeedbackStore {
	return &InMemoryFeedbackStore{
		feedback: make([]*BetaFeedback, 0),
	}
}

// StoreFeedback stores feedback from a beta tester
func (s *InMemoryFeedbackStore) StoreFeedback(feedback *BetaFeedback) error {
	s.feedback = append(s.feedback, feedback)
	return nil
}

// GetFeedback retrieves feedback for a specific beta tester
func (s *InMemoryFeedbackStore) GetFeedback(betaTesterID string) ([]*BetaFeedback, error) {
	var result []*BetaFeedback
	for _, f := range s.feedback {
		if f.BetaTesterID == betaTesterID {
			result = append(result, f)
		}
	}
	return result, nil
}

// GetAllFeedback retrieves all feedback
func (s *InMemoryFeedbackStore) GetAllFeedback() ([]*BetaFeedback, error) {
	return s.feedback, nil
}

// GetFeedbackStats calculates feedback statistics
func (s *InMemoryFeedbackStore) GetFeedbackStats() (*FeedbackStats, error) {
	if len(s.feedback) == 0 {
		return &FeedbackStats{}, nil
	}

	stats := &FeedbackStats{
		TotalFeedback:     len(s.feedback),
		CategoryRatings:   make(map[string]float64),
		SDKUsage:          make(map[string]int),
		IntegrationTypes:  make(map[string]int),
		TestScenarioStats: make(map[string]int),
		RecentFeedback:    make([]*BetaFeedback, 0),
	}

	var totalRating float64
	var apiDesignTotal, performanceTotal, devExpTotal, featuresTotal float64
	var apiDesignCount, performanceCount, devExpCount, featuresCount int

	// Get recent feedback (last 10)
	recentCount := 10
	if len(s.feedback) < recentCount {
		recentCount = len(s.feedback)
	}
	stats.RecentFeedback = s.feedback[len(s.feedback)-recentCount:]

	for _, f := range s.feedback {
		totalRating += float64(f.OverallRating)

		// Category ratings
		if f.APIDesign != nil {
			apiDesignTotal += float64(f.APIDesign.Rating)
			apiDesignCount++
		}
		if f.Performance != nil {
			performanceTotal += float64(f.Performance.Rating)
			performanceCount++
		}
		if f.DeveloperExp != nil {
			devExpTotal += float64(f.DeveloperExp.Rating)
			devExpCount++
		}
		if f.Features != nil {
			featuresTotal += float64(f.Features.Rating)
			featuresCount++
		}

		// Bug and feature request counts
		stats.BugCount += len(f.BugsFound)
		stats.FeatureRequestCount += len(f.FeatureRequests)

		// SDK usage
		if f.SDKUsed != "" {
			stats.SDKUsage[f.SDKUsed]++
		}

		// Integration types
		if f.IntegrationType != "" {
			stats.IntegrationTypes[f.IntegrationType]++
		}

		// Test scenario stats
		for _, scenario := range f.TestScenarios {
			stats.TestScenarioStats[scenario.Status]++
		}
	}

	// Calculate averages
	stats.AverageRating = totalRating / float64(len(s.feedback))

	if apiDesignCount > 0 {
		stats.CategoryRatings["api_design"] = apiDesignTotal / float64(apiDesignCount)
	}
	if performanceCount > 0 {
		stats.CategoryRatings["performance"] = performanceTotal / float64(performanceCount)
	}
	if devExpCount > 0 {
		stats.CategoryRatings["developer_experience"] = devExpTotal / float64(devExpCount)
	}
	if featuresCount > 0 {
		stats.CategoryRatings["features"] = featuresTotal / float64(featuresCount)
	}

	return stats, nil
}

// HandleSubmitFeedback handles feedback submission
func (fc *FeedbackCollector) HandleSubmitFeedback(w http.ResponseWriter, r *http.Request) {
	var feedback BetaFeedback
	if err := json.NewDecoder(r.Body).Decode(&feedback); err != nil {
		fc.logger.Error("Failed to decode feedback", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if feedback.BetaTesterID == "" || feedback.BetaTesterName == "" || feedback.Email == "" {
		http.Error(w, "Missing required fields: beta_tester_id, beta_tester_name, email", http.StatusBadRequest)
		return
	}

	// Set submission time and generate ID
	feedback.SubmittedAt = time.Now()
	feedback.ID = fmt.Sprintf("feedback_%d", time.Now().UnixNano())

	// Store feedback
	if err := fc.store.StoreFeedback(&feedback); err != nil {
		fc.logger.Error("Failed to store feedback", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fc.logger.Info("Feedback submitted",
		zap.String("beta_tester_id", feedback.BetaTesterID),
		zap.String("beta_tester_name", feedback.BetaTesterName),
		zap.Int("overall_rating", feedback.OverallRating))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Feedback submitted successfully",
		"id":           feedback.ID,
		"submitted_at": feedback.SubmittedAt,
	})
}

// HandleGetFeedback handles retrieving feedback for a beta tester
func (fc *FeedbackCollector) HandleGetFeedback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	betaTesterID := vars["betaTesterID"]

	feedback, err := fc.store.GetFeedback(betaTesterID)
	if err != nil {
		fc.logger.Error("Failed to get feedback", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"beta_tester_id": betaTesterID,
		"feedback":       feedback,
		"count":          len(feedback),
	})
}

// HandleGetAllFeedback handles retrieving all feedback
func (fc *FeedbackCollector) HandleGetAllFeedback(w http.ResponseWriter, r *http.Request) {
	feedback, err := fc.store.GetAllFeedback()
	if err != nil {
		fc.logger.Error("Failed to get all feedback", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"feedback": feedback,
		"count":    len(feedback),
	})
}

// HandleGetFeedbackStats handles retrieving feedback statistics
func (fc *FeedbackCollector) HandleGetFeedbackStats(w http.ResponseWriter, r *http.Request) {
	stats, err := fc.store.GetFeedbackStats()
	if err != nil {
		fc.logger.Error("Failed to get feedback stats", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleSubmitBugReport handles bug report submission
func (fc *FeedbackCollector) HandleSubmitBugReport(w http.ResponseWriter, r *http.Request) {
	var bugReport BugReport
	if err := json.NewDecoder(r.Body).Decode(&bugReport); err != nil {
		fc.logger.Error("Failed to decode bug report", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if bugReport.Title == "" || bugReport.Description == "" {
		http.Error(w, "Missing required fields: title, description", http.StatusBadRequest)
		return
	}

	// Set report time and generate ID
	bugReport.ReportedAt = time.Now()
	bugReport.ID = fmt.Sprintf("bug_%d", time.Now().UnixNano())

	fc.logger.Info("Bug report submitted",
		zap.String("id", bugReport.ID),
		zap.String("title", bugReport.Title),
		zap.String("severity", bugReport.Severity))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Bug report submitted successfully",
		"id":          bugReport.ID,
		"reported_at": bugReport.ReportedAt,
	})
}

// HandleSubmitFeatureRequest handles feature request submission
func (fc *FeedbackCollector) HandleSubmitFeatureRequest(w http.ResponseWriter, r *http.Request) {
	var featureRequest FeatureRequest
	if err := json.NewDecoder(r.Body).Decode(&featureRequest); err != nil {
		fc.logger.Error("Failed to decode feature request", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if featureRequest.Title == "" || featureRequest.Description == "" {
		http.Error(w, "Missing required fields: title, description", http.StatusBadRequest)
		return
	}

	// Set request time and generate ID
	featureRequest.RequestedAt = time.Now()
	featureRequest.ID = fmt.Sprintf("feature_%d", time.Now().UnixNano())

	fc.logger.Info("Feature request submitted",
		zap.String("id", featureRequest.ID),
		zap.String("title", featureRequest.Title),
		zap.String("priority", featureRequest.Priority))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Feature request submitted successfully",
		"id":           featureRequest.ID,
		"requested_at": featureRequest.RequestedAt,
	})
}

// SetupRoutes sets up the feedback collection routes
func (fc *FeedbackCollector) SetupRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1/beta").Subrouter()

	// Feedback routes
	api.HandleFunc("/feedback", fc.HandleSubmitFeedback).Methods("POST")
	api.HandleFunc("/feedback/{betaTesterID}", fc.HandleGetFeedback).Methods("GET")
	api.HandleFunc("/feedback", fc.HandleGetAllFeedback).Methods("GET")
	api.HandleFunc("/feedback/stats", fc.HandleGetFeedbackStats).Methods("GET")

	// Bug report routes
	api.HandleFunc("/bugs", fc.HandleSubmitBugReport).Methods("POST")

	// Feature request routes
	api.HandleFunc("/features", fc.HandleSubmitFeatureRequest).Methods("POST")
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Create feedback collector
	collector := NewFeedbackCollector(logger)

	// Setup router
	router := mux.NewRouter()
	collector.SetupRoutes(router)

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "beta-feedback-collector",
			"timestamp": time.Now(),
		})
	}).Methods("GET")

	// Start server
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	logger.Info("Starting beta feedback collector server", zap.String("port", port))
	log.Fatal(http.ListenAndServe(":"+port, router))
}
