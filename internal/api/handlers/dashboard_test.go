package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kyb-platform/internal/config"
	"kyb-platform/internal/observability"
	"kyb-platform/internal/risk"
)

func setupDashboardTest(t *testing.T) *DashboardHandler {
	// Create test observability config
	obsConfig := &config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	}
	logger := observability.NewLogger(obsConfig)

	// Create mock risk service
	riskService := &risk.RiskService{}

	handler := NewDashboardHandler(logger, riskService)
	return handler
}

func TestGetDashboardComplianceOverviewHandler(t *testing.T) {
	handler := setupDashboardTest(t)

	req := httptest.NewRequest("GET", "/v1/dashboard/compliance/overview", nil)
	ctx := context.WithValue(req.Context(), "request_id", "test-request-id")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetDashboardComplianceOverviewHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify expected fields are present
	expectedFields := []string{
		"total_businesses",
		"compliant_businesses",
		"non_compliant_businesses",
		"in_progress_businesses",
		"active_alerts",
		"critical_alerts",
		"average_compliance_score",
		"framework_distribution",
		"recent_compliance_events",
		"upcoming_reviews",
		"last_updated",
	}

	for _, field := range expectedFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Expected field '%s' in response, but it was not found", field)
		}
	}

	// Verify specific data types and values
	if totalBusinesses, ok := response["total_businesses"].(float64); !ok || totalBusinesses <= 0 {
		t.Errorf("Expected positive total_businesses, got %v", totalBusinesses)
	}

	if avgScore, ok := response["average_compliance_score"].(float64); !ok || avgScore < 0 || avgScore > 100 {
		t.Errorf("Expected average_compliance_score between 0 and 100, got %v", avgScore)
	}

	if frameworkDist, ok := response["framework_distribution"].(map[string]interface{}); !ok || len(frameworkDist) == 0 {
		t.Errorf("Expected non-empty framework_distribution, got %v", frameworkDist)
	}
}

func TestGetDashboardComplianceBusinessHandler(t *testing.T) {
	handler := setupDashboardTest(t)

	tests := []struct {
		name           string
		businessID     string
		expectedStatus int
	}{
		{
			name:           "Valid business ID",
			businessID:     "business-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty business ID",
			businessID:     "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/v1/dashboard/compliance/business/" + tt.businessID
			req := httptest.NewRequest("GET", url, nil)
			// Set up the URL pattern for path value extraction
			req.URL.Path = "/v1/dashboard/compliance/business/" + tt.businessID
			ctx := context.WithValue(req.Context(), "request_id", "test-request-id")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.GetDashboardComplianceBusinessHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				// Verify expected fields are present
				expectedFields := []string{
					"business_id",
					"business_name",
					"overall_compliance_score",
					"overall_status",
					"framework_scores",
					"framework_statuses",
					"recent_alerts",
					"recent_assessments",
					"upcoming_reviews",
					"compliance_trends",
					"last_updated",
				}

				for _, field := range expectedFields {
					if _, exists := response[field]; !exists {
						t.Errorf("Expected field '%s' in response, but it was not found", field)
					}
				}

				// Verify business_id matches
				if businessID, ok := response["business_id"].(string); !ok || businessID != tt.businessID {
					t.Errorf("Expected business_id '%s', got '%s'", tt.businessID, businessID)
				}

				// Verify compliance score is valid
				if score, ok := response["overall_compliance_score"].(float64); !ok || score < 0 || score > 100 {
					t.Errorf("Expected overall_compliance_score between 0 and 100, got %v", score)
				}
			}
		})
	}
}

func TestGetDashboardComplianceAnalyticsHandler(t *testing.T) {
	handler := setupDashboardTest(t)

	req := httptest.NewRequest("GET", "/v1/dashboard/compliance/analytics", nil)
	ctx := context.WithValue(req.Context(), "request_id", "test-request-id")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetDashboardComplianceAnalyticsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify expected fields are present
	expectedFields := []string{
		"compliance_score_distribution",
		"framework_compliance_averages",
		"alert_trends",
		"assessment_trends",
		"top_compliance_issues",
		"geographic_compliance_data",
		"industry_compliance_data",
		"time_range",
		"last_updated",
	}

	for _, field := range expectedFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Expected field '%s' in response, but it was not found", field)
		}
	}

	// Verify score distribution is valid
	if scoreDist, ok := response["compliance_score_distribution"].(map[string]interface{}); !ok || len(scoreDist) == 0 {
		t.Errorf("Expected non-empty compliance_score_distribution, got %v", scoreDist)
	}

	// Verify framework averages are valid
	if frameworkAvgs, ok := response["framework_compliance_averages"].(map[string]interface{}); !ok || len(frameworkAvgs) == 0 {
		t.Errorf("Expected non-empty framework_compliance_averages, got %v", frameworkAvgs)
	}

	// Verify time range is set
	if timeRange, ok := response["time_range"].(string); !ok || timeRange == "" {
		t.Errorf("Expected non-empty time_range, got %v", timeRange)
	}
}

func TestGetDashboardComplianceOverviewHandler_WithTimeRange(t *testing.T) {
	handler := setupDashboardTest(t)

	req := httptest.NewRequest("GET", "/v1/dashboard/compliance/overview?time_range=7d", nil)
	ctx := context.WithValue(req.Context(), "request_id", "test-request-id")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetDashboardComplianceOverviewHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify the response contains the expected data structure
	if totalBusinesses, ok := response["total_businesses"].(float64); !ok || totalBusinesses <= 0 {
		t.Errorf("Expected positive total_businesses, got %v", totalBusinesses)
	}

	if recentEvents, ok := response["recent_compliance_events"].([]interface{}); !ok {
		t.Errorf("Expected recent_compliance_events array, got %v", recentEvents)
	}

	if upcomingReviews, ok := response["upcoming_reviews"].([]interface{}); !ok {
		t.Errorf("Expected upcoming_reviews array, got %v", upcomingReviews)
	}
}
