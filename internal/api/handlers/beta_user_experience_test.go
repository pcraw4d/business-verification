package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	// "github.com/pcraw4d/business-verification/internal/webanalysis" // Temporarily disabled for deployment
)

func TestNewBetaUserExperienceHandler(t *testing.T) {
	logger := zap.NewNop()
	betaFramework := &webanalysis.BetaTestingFramework{}

	handler := NewBetaUserExperienceHandler(logger, betaFramework)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.Equal(t, betaFramework, handler.betaFramework)
}

func TestSelectScrapingMethod(t *testing.T) {
	handler := createTestHandler(t)

	tests := []struct {
		name           string
		request        ScrapingMethodSelectionRequest
		expectedMethod string
		expectedReason string
		statusCode     int
	}{
		{
			name: "Valid request with enhanced method",
			request: ScrapingMethodSelectionRequest{
				UserID:          "user_1",
				URL:             "https://example.com",
				PreferredMethod: "enhanced",
			},
			expectedMethod: "enhanced",
			expectedReason: "User preference",
			statusCode:     200,
		},
		{
			name: "Valid request with auto method",
			request: ScrapingMethodSelectionRequest{
				UserID:          "user_1",
				URL:             "https://example.com",
				PreferredMethod: "auto",
			},
			expectedMethod: "enhanced",
			expectedReason: "Beta user with auto selection",
			statusCode:     200,
		},
		{
			name: "Valid request with forced method",
			request: ScrapingMethodSelectionRequest{
				UserID:          "user_1",
				URL:             "https://example.com",
				PreferredMethod: "basic",
				ForceMethod:     true,
			},
			expectedMethod: "basic",
			expectedReason: "User preference forced",
			statusCode:     200,
		},
		{
			name: "Missing user ID",
			request: ScrapingMethodSelectionRequest{
				URL:             "https://example.com",
				PreferredMethod: "enhanced",
			},
			statusCode: 400,
		},
		{
			name: "Missing URL",
			request: ScrapingMethodSelectionRequest{
				UserID:          "user_1",
				PreferredMethod: "enhanced",
			},
			statusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/v2/beta/scraping-method", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.SelectScrapingMethod(w, req)

			assert.Equal(t, tt.statusCode, w.Code)

			if tt.statusCode == 200 {
				var response ScrapingMethodSelectionResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedMethod, response.SelectedMethod)
				assert.Equal(t, tt.expectedMethod, response.MethodUsed)
				assert.True(t, response.IsBetaUser)
				assert.NotNil(t, response.Transparency)
				assert.Contains(t, response.Message, tt.expectedReason)
			}
		})
	}
}

func TestSubmitFeedback(t *testing.T) {
	handler := createTestHandler(t)

	tests := []struct {
		name       string
		request    UserFeedbackRequest
		statusCode int
	}{
		{
			name: "Valid feedback submission",
			request: UserFeedbackRequest{
				UserID:       "user_1",
				TestID:       "test_1",
				URL:          "https://example.com",
				Method:       "enhanced",
				Satisfaction: 4,
				Accuracy:     5,
				Speed:        4,
				Comments:     "Great results",
			},
			statusCode: 201,
		},
		{
			name: "Missing user ID",
			request: UserFeedbackRequest{
				TestID:       "test_1",
				URL:          "https://example.com",
				Method:       "enhanced",
				Satisfaction: 4,
				Accuracy:     5,
				Speed:        4,
			},
			statusCode: 400,
		},
		{
			name: "Missing test ID",
			request: UserFeedbackRequest{
				UserID:       "user_1",
				URL:          "https://example.com",
				Method:       "enhanced",
				Satisfaction: 4,
				Accuracy:     5,
				Speed:        4,
			},
			statusCode: 400,
		},
		{
			name: "Invalid satisfaction score",
			request: UserFeedbackRequest{
				UserID:       "user_1",
				TestID:       "test_1",
				URL:          "https://example.com",
				Method:       "enhanced",
				Satisfaction: 6, // Invalid: should be 1-5
				Accuracy:     5,
				Speed:        4,
			},
			statusCode: 400,
		},
		{
			name: "Invalid accuracy score",
			request: UserFeedbackRequest{
				UserID:       "user_1",
				TestID:       "test_1",
				URL:          "https://example.com",
				Method:       "enhanced",
				Satisfaction: 4,
				Accuracy:     0, // Invalid: should be 1-5
				Speed:        4,
			},
			statusCode: 400,
		},
		{
			name: "Invalid speed score",
			request: UserFeedbackRequest{
				UserID:       "user_1",
				TestID:       "test_1",
				URL:          "https://example.com",
				Method:       "enhanced",
				Satisfaction: 4,
				Accuracy:     5,
				Speed:        6, // Invalid: should be 1-5
			},
			statusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/v2/beta/feedback", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.SubmitFeedback(w, req)

			assert.Equal(t, tt.statusCode, w.Code)

			if tt.statusCode == 201 {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, "success", response["status"])
				assert.Equal(t, "Feedback submitted successfully", response["message"])
			}
		})
	}
}

func TestUpdatePreferences(t *testing.T) {
	handler := createTestHandler(t)

	tests := []struct {
		name       string
		request    UserPreferenceRequest
		statusCode int
	}{
		{
			name: "Valid preferences update",
			request: UserPreferenceRequest{
				UserID:             "user_1",
				DefaultMethod:      "enhanced",
				EnableBeta:         true,
				EnableTransparency: true,
				EnableFeedback:     true,
			},
			statusCode: 200,
		},
		{
			name: "Missing user ID",
			request: UserPreferenceRequest{
				DefaultMethod:      "enhanced",
				EnableBeta:         true,
				EnableTransparency: true,
				EnableFeedback:     true,
			},
			statusCode: 400,
		},
		{
			name: "Invalid default method",
			request: UserPreferenceRequest{
				UserID:             "user_1",
				DefaultMethod:      "invalid",
				EnableBeta:         true,
				EnableTransparency: true,
				EnableFeedback:     true,
			},
			statusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("PUT", "/v2/beta/preferences", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.UpdatePreferences(w, req)

			assert.Equal(t, tt.statusCode, w.Code)

			if tt.statusCode == 200 {
				var response UserPreferenceResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, tt.request.UserID, response.UserID)
				assert.Equal(t, tt.request.DefaultMethod, response.DefaultMethod)
				assert.Equal(t, tt.request.EnableBeta, response.EnableBeta)
				assert.Equal(t, tt.request.EnableTransparency, response.EnableTransparency)
				assert.Equal(t, tt.request.EnableFeedback, response.EnableFeedback)
				assert.NotZero(t, response.LastUpdated)
			}
		})
	}
}

func TestGetPreferences(t *testing.T) {
	handler := createTestHandler(t)

	req := httptest.NewRequest("GET", "/v2/beta/preferences/user_1", nil)
	w := httptest.NewRecorder()

	// Set up router to handle path parameters
	router := mux.NewRouter()
	router.HandleFunc("/v2/beta/preferences/{user_id}", handler.GetPreferences)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response UserPreferenceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "user_1", response.UserID)
	assert.Equal(t, "auto", response.DefaultMethod)
	assert.True(t, response.EnableBeta)
	assert.True(t, response.EnableTransparency)
	assert.True(t, response.EnableFeedback)
	assert.NotZero(t, response.LastUpdated)
}

func TestGetTransparency(t *testing.T) {
	handler := createTestHandler(t)

	// Mock the beta framework to return test analysis
	handler.betaFramework.abTestManager = &MockABTestManager{
		analysis: &webanalysis.ABTestAnalysis{
			TestID:                  "test_1",
			EnhancedSuccessRate:     0.95,
			BasicSuccessRate:        0.80,
			EnhancedAvgResponseTime: 1500 * time.Millisecond,
			BasicAvgResponseTime:    2000 * time.Millisecond,
			ImprovementMetrics: &webanalysis.ImprovementMetrics{
				SuccessRateImprovement: 0.15,
			},
		},
	}

	req := httptest.NewRequest("GET", "/v2/beta/transparency/test_1", nil)
	w := httptest.NewRecorder()

	// Set up router to handle path parameters
	router := mux.NewRouter()
	router.HandleFunc("/v2/beta/transparency/{test_id}", handler.GetTransparency)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response ScrapingTransparency
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "enhanced", response.MethodUsed)
	assert.Equal(t, "Beta test participant", response.ReasonForSelection)
	assert.Equal(t, "test_1", response.BetaTestID)
	assert.NotNil(t, response.PerformanceMetrics)
	assert.Equal(t, 0.95, response.PerformanceMetrics.EnhancedSuccessRate)
	assert.Equal(t, 0.80, response.PerformanceMetrics.BasicSuccessRate)
	assert.Equal(t, 0.15, response.PerformanceMetrics.Improvement)
}

func TestGetPerformanceComparison(t *testing.T) {
	handler := createTestHandler(t)

	// Mock the beta framework to return test comparison
	handler.betaFramework.performanceTracker = &MockSuccessRateTracker{
		basicMetrics: &webanalysis.ScrapingMetrics{
			SuccessRate:         0.80,
			AverageResponseTime: 2 * time.Second,
			DataQuality:         0.70,
		},
		enhancedMetrics: &webanalysis.ScrapingMetrics{
			SuccessRate:         0.95,
			AverageResponseTime: 1.5 * time.Second,
			DataQuality:         0.90,
		},
	}

	req := httptest.NewRequest("GET", "/v2/beta/performance-comparison?time_range=24", nil)
	w := httptest.NewRecorder()

	handler.GetPerformanceComparison(w, req)

	assert.Equal(t, 200, w.Code)

	var response webanalysis.PerformanceComparison
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 24*time.Hour, response.TimeRange)
	assert.NotNil(t, response.BasicMetrics)
	assert.NotNil(t, response.EnhancedMetrics)
	assert.Greater(t, response.SuccessRateImprovement, 0.0)
	assert.Greater(t, response.ResponseTimeImprovement, time.Duration(0))
	assert.Greater(t, response.DataQualityImprovement, 0.0)
}

func TestGetUserStats(t *testing.T) {
	handler := createTestHandler(t)

	// Mock the beta framework to return test feedback
	handler.betaFramework.feedbackCollector = &MockBetaFeedbackCollector{
		feedback: []*webanalysis.BetaFeedback{
			{
				UserID:       "user_1",
				TestID:       "test_1",
				Method:       "enhanced",
				Satisfaction: 4,
				Accuracy:     5,
				Speed:        4,
				Timestamp:    time.Now(),
			},
			{
				UserID:       "user_1",
				TestID:       "test_2",
				Method:       "enhanced",
				Satisfaction: 5,
				Accuracy:     4,
				Speed:        5,
				Timestamp:    time.Now(),
			},
		},
	}

	req := httptest.NewRequest("GET", "/v2/beta/user-stats/user_1", nil)
	w := httptest.NewRecorder()

	// Set up router to handle path parameters
	router := mux.NewRouter()
	router.HandleFunc("/v2/beta/user-stats/{user_id}", handler.GetUserStats)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["total_feedback"])
	assert.Equal(t, 4.5, response["average_satisfaction"])
	assert.Equal(t, 4.5, response["average_accuracy"])
	assert.Equal(t, 4.5, response["average_speed"])
}

func TestIsBetaUser(t *testing.T) {
	handler := createTestHandler(t)

	// Test that all users are considered beta users in the current implementation
	assert.True(t, handler.isBetaUser("user_1"))
	assert.True(t, handler.isBetaUser("user_2"))
	assert.True(t, handler.isBetaUser(""))
}

func TestSelectMethod(t *testing.T) {
	handler := createTestHandler(t)

	tests := []struct {
		name           string
		request        ScrapingMethodSelectionRequest
		isBetaUser     bool
		expectedMethod string
		expectedReason string
	}{
		{
			name: "Force method",
			request: ScrapingMethodSelectionRequest{
				PreferredMethod: "basic",
				ForceMethod:     true,
			},
			isBetaUser:     true,
			expectedMethod: "basic",
			expectedReason: "User preference forced",
		},
		{
			name: "Non-beta user",
			request: ScrapingMethodSelectionRequest{
				PreferredMethod: "enhanced",
			},
			isBetaUser:     false,
			expectedMethod: "basic",
			expectedReason: "User not in beta program",
		},
		{
			name: "Beta user with preferred method",
			request: ScrapingMethodSelectionRequest{
				PreferredMethod: "enhanced",
			},
			isBetaUser:     true,
			expectedMethod: "enhanced",
			expectedReason: "User preference",
		},
		{
			name: "Beta user with auto selection",
			request: ScrapingMethodSelectionRequest{
				PreferredMethod: "auto",
			},
			isBetaUser:     true,
			expectedMethod: "enhanced",
			expectedReason: "Beta user with auto selection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, reason := handler.selectMethod(tt.request, tt.isBetaUser)
			assert.Equal(t, tt.expectedMethod, method)
			assert.Equal(t, tt.expectedReason, reason)
		})
	}
}

func TestGenerateTestID(t *testing.T) {
	handler := createTestHandler(t)

	testID1 := handler.generateTestID("user_1", "https://example.com")
	testID2 := handler.generateTestID("user_1", "https://example.com")

	assert.Contains(t, testID1, "beta_test_user_1")
	assert.Contains(t, testID2, "beta_test_user_1")
	assert.NotEqual(t, testID1, testID2) // Should be different due to timestamp
}

func TestCalculateUserStats(t *testing.T) {
	handler := createTestHandler(t)

	tests := []struct {
		name     string
		feedback []*webanalysis.BetaFeedback
		expected map[string]interface{}
	}{
		{
			name:     "Empty feedback",
			feedback: []*webanalysis.BetaFeedback{},
			expected: map[string]interface{}{
				"total_feedback":       0,
				"average_satisfaction": 0.0,
				"average_accuracy":     0.0,
				"average_speed":        0.0,
			},
		},
		{
			name: "Single feedback",
			feedback: []*webanalysis.BetaFeedback{
				{
					Satisfaction: 4,
					Accuracy:     5,
					Speed:        3,
				},
			},
			expected: map[string]interface{}{
				"total_feedback":       1,
				"average_satisfaction": 4.0,
				"average_accuracy":     5.0,
				"average_speed":        3.0,
			},
		},
		{
			name: "Multiple feedback",
			feedback: []*webanalysis.BetaFeedback{
				{
					Satisfaction: 4,
					Accuracy:     5,
					Speed:        3,
				},
				{
					Satisfaction: 5,
					Accuracy:     4,
					Speed:        5,
				},
			},
			expected: map[string]interface{}{
				"total_feedback":       2,
				"average_satisfaction": 4.5,
				"average_accuracy":     4.5,
				"average_speed":        4.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := handler.calculateUserStats(tt.feedback)
			assert.Equal(t, tt.expected["total_feedback"], stats["total_feedback"])
			assert.Equal(t, tt.expected["average_satisfaction"], stats["average_satisfaction"])
			assert.Equal(t, tt.expected["average_accuracy"], stats["average_accuracy"])
			assert.Equal(t, tt.expected["average_speed"], stats["average_speed"])
		})
	}
}

// Helper functions

func createTestHandler(t *testing.T) *BetaUserExperienceHandler {
	logger := zap.NewNop()
	betaFramework := &webanalysis.BetaTestingFramework{}

	return NewBetaUserExperienceHandler(logger, betaFramework)
}

// Mock implementations for testing

type MockABTestManager struct {
	analysis *webanalysis.ABTestAnalysis
	err      error
}

func (m *MockABTestManager) GetAnalysis(ctx context.Context, testID string) (*webanalysis.ABTestAnalysis, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.analysis, nil
}

type MockBetaFeedbackCollector struct {
	feedback []*webanalysis.BetaFeedback
	err      error
}

func (m *MockBetaFeedbackCollector) GetUserFeedback(userID string) ([]*webanalysis.BetaFeedback, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.feedback, nil
}
