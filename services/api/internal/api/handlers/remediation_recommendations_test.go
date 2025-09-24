package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetRemediationRecommendations(t *testing.T) {
	handler := NewComplianceGapHandler()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Get all recommendations",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  5, // Based on our test data
		},
		{
			name:           "Filter by priority critical",
			queryParams:    "?priority=critical",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Filter by category technical",
			queryParams:    "?category=technical",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "Filter by timeframe immediate",
			queryParams:    "?timeframe=immediate",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "Multiple filters",
			queryParams:    "?priority=high&category=technical",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/compliance/recommendations"+tt.queryParams, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.GetRemediationRecommendations(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatal("Failed to unmarshal response:", err)
			}

			recommendations, ok := response["recommendations"].([]interface{})
			if !ok {
				t.Fatal("Expected recommendations array in response")
			}

			if len(recommendations) != tt.expectedCount {
				t.Errorf("Expected %d recommendations, got %d", tt.expectedCount, len(recommendations))
			}

			// Verify response structure
			if _, ok := response["total_count"]; !ok {
				t.Error("Expected total_count in response")
			}

			if _, ok := response["filters"]; !ok {
				t.Error("Expected filters in response")
			}

			if _, ok := response["generated_at"]; !ok {
				t.Error("Expected generated_at in response")
			}
		})
	}
}

func TestGetRecommendationDetails(t *testing.T) {
	handler := NewComplianceGapHandler()

	tests := []struct {
		name             string
		recommendationID string
		expectedStatus   int
		shouldExist      bool
	}{
		{
			name:             "Get existing recommendation",
			recommendationID: "rec-001",
			expectedStatus:   http.StatusOK,
			shouldExist:      true,
		},
		{
			name:             "Get non-existing recommendation",
			recommendationID: "rec-999",
			expectedStatus:   http.StatusNotFound,
			shouldExist:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/compliance/recommendations/"+tt.recommendationID, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.GetRecommendationDetails(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.shouldExist {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatal("Failed to unmarshal response:", err)
				}

				// Verify response structure
				if _, ok := response["recommendation"]; !ok {
					t.Error("Expected recommendation in response")
				}

				if _, ok := response["related_gaps"]; !ok {
					t.Error("Expected related_gaps in response")
				}

				if _, ok := response["implementation"]; !ok {
					t.Error("Expected implementation in response")
				}

				if _, ok := response["resources"]; !ok {
					t.Error("Expected resources in response")
				}

				if _, ok := response["timeline"]; !ok {
					t.Error("Expected timeline in response")
				}
			}
		})
	}
}

func TestCreateRemediationPlan(t *testing.T) {
	handler := NewComplianceGapHandler()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		shouldSucceed  bool
	}{
		{
			name: "Create valid remediation plan",
			requestBody: `{
				"recommendation_ids": ["rec-001", "rec-002"],
				"plan_name": "Security Enhancement Plan",
				"owner": "security_team",
				"target_date": "2025-06-01",
				"budget": 50000
			}`,
			expectedStatus: http.StatusOK,
			shouldSucceed:  true,
		},
		{
			name: "Create plan with invalid JSON",
			requestBody: `{
				"recommendation_ids": ["rec-001", "rec-002"],
				"plan_name": "Security Enhancement Plan",
				"owner": "security_team",
				"target_date": "2025-06-01",
				"budget": 50000
			`, // Missing closing brace
			expectedStatus: http.StatusBadRequest,
			shouldSucceed:  false,
		},
		{
			name: "Create plan with empty recommendation IDs",
			requestBody: `{
				"recommendation_ids": [],
				"plan_name": "Empty Plan",
				"owner": "security_team",
				"target_date": "2025-06-01",
				"budget": 0
			}`,
			expectedStatus: http.StatusOK,
			shouldSucceed:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/v1/compliance/plans", strings.NewReader(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.CreateRemediationPlan(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.shouldSucceed {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatal("Failed to unmarshal response:", err)
				}

				// Verify response structure
				if _, ok := response["plan"]; !ok {
					t.Error("Expected plan in response")
				}

				if _, ok := response["message"]; !ok {
					t.Error("Expected message in response")
				}

				if _, ok := response["next_steps"]; !ok {
					t.Error("Expected next_steps in response")
				}

				if _, ok := response["estimated_cost"]; !ok {
					t.Error("Expected estimated_cost in response")
				}

				if _, ok := response["timeline"]; !ok {
					t.Error("Expected timeline in response")
				}
			}
		})
	}
}

func TestGenerateRecommendations(t *testing.T) {
	handler := NewComplianceGapHandler()

	tests := []struct {
		name          string
		priority      string
		category      string
		timeframe     string
		expectedCount int
	}{
		{
			name:          "All recommendations",
			priority:      "",
			category:      "",
			timeframe:     "",
			expectedCount: 5,
		},
		{
			name:          "Critical priority only",
			priority:      "critical",
			category:      "",
			timeframe:     "",
			expectedCount: 1,
		},
		{
			name:          "Technical category only",
			priority:      "",
			category:      "technical",
			timeframe:     "",
			expectedCount: 3,
		},
		{
			name:          "Immediate timeframe only",
			priority:      "",
			category:      "",
			timeframe:     "immediate",
			expectedCount: 2,
		},
		{
			name:          "High priority and technical",
			priority:      "high",
			category:      "technical",
			timeframe:     "",
			expectedCount: 2,
		},
		{
			name:          "Non-existent filter",
			priority:      "nonexistent",
			category:      "",
			timeframe:     "",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := handler.generateRecommendations(tt.priority, tt.category, tt.timeframe)

			if len(recommendations) != tt.expectedCount {
				t.Errorf("Expected %d recommendations, got %d", tt.expectedCount, len(recommendations))
			}

			// Verify that all returned recommendations match the filters
			for _, rec := range recommendations {
				if tt.priority != "" && tt.priority != "all" && rec.Priority != tt.priority {
					t.Errorf("Expected priority %s, got %s", tt.priority, rec.Priority)
				}
				if tt.category != "" && tt.category != "all" && rec.Category != tt.category {
					t.Errorf("Expected category %s, got %s", tt.category, rec.Category)
				}
				if tt.timeframe != "" && tt.timeframe != "all" && rec.Timeframe != tt.timeframe {
					t.Errorf("Expected timeframe %s, got %s", tt.timeframe, rec.Timeframe)
				}
			}
		})
	}
}

func TestGetRecommendationByID(t *testing.T) {
	handler := NewComplianceGapHandler()

	tests := []struct {
		name          string
		id            string
		shouldExist   bool
		expectedTitle string
	}{
		{
			name:          "Existing recommendation",
			id:            "rec-001",
			shouldExist:   true,
			expectedTitle: "Implement Multi-Factor Authentication",
		},
		{
			name:        "Non-existing recommendation",
			id:          "rec-999",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendation := handler.getRecommendationByID(tt.id)

			if tt.shouldExist {
				if recommendation == nil {
					t.Error("Expected recommendation to exist")
					return
				}
				if recommendation.Title != tt.expectedTitle {
					t.Errorf("Expected title %s, got %s", tt.expectedTitle, recommendation.Title)
				}
			} else {
				if recommendation != nil {
					t.Error("Expected recommendation to not exist")
				}
			}
		})
	}
}

func TestCalculatePlanCost(t *testing.T) {
	handler := NewComplianceGapHandler()

	// Create test recommendations
	recommendations := []RemediationRecommendation{
		{
			ID:    "rec-001",
			Title: "Test Recommendation 1",
			Cost:  "$5,000 - $15,000",
		},
		{
			ID:    "rec-002",
			Title: "Test Recommendation 2",
			Cost:  "$10,000 - $25,000",
		},
	}

	cost := handler.calculatePlanCost(recommendations)

	// Verify cost calculation
	if cost["estimated_min"] != 15000.0 {
		t.Errorf("Expected estimated_min 15000.0, got %v", cost["estimated_min"])
	}

	if cost["estimated_max"] != 40000.0 {
		t.Errorf("Expected estimated_max 40000.0, got %v", cost["estimated_max"])
	}

	if cost["currency"] != "USD" {
		t.Errorf("Expected currency USD, got %v", cost["currency"])
	}
}

func TestCalculatePlanTimeline(t *testing.T) {
	handler := NewComplianceGapHandler()

	tests := []struct {
		name             string
		recommendations  []RemediationRecommendation
		expectedDuration string
		expectedParallel bool
	}{
		{
			name: "Immediate timeframe",
			recommendations: []RemediationRecommendation{
				{Timeframe: "immediate"},
				{Timeframe: "immediate"},
			},
			expectedDuration: "immediate",
			expectedParallel: true,
		},
		{
			name: "Mixed timeframes",
			recommendations: []RemediationRecommendation{
				{Timeframe: "immediate"},
				{Timeframe: "short"},
			},
			expectedDuration: "short",
			expectedParallel: true,
		},
		{
			name: "Long timeframe",
			recommendations: []RemediationRecommendation{
				{Timeframe: "immediate"},
				{Timeframe: "long"},
			},
			expectedDuration: "long",
			expectedParallel: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeline := handler.calculatePlanTimeline(tt.recommendations)

			if timeline["estimated_duration"] != tt.expectedDuration {
				t.Errorf("Expected duration %s, got %v", tt.expectedDuration, timeline["estimated_duration"])
			}

			if timeline["parallel_execution"] != tt.expectedParallel {
				t.Errorf("Expected parallel_execution %v, got %v", tt.expectedParallel, timeline["parallel_execution"])
			}

			if _, ok := timeline["critical_path"]; !ok {
				t.Error("Expected critical_path in timeline")
			}
		})
	}
}

func TestRemediationRecommendationStructure(t *testing.T) {
	handler := NewComplianceGapHandler()
	recommendations := handler.generateRecommendations("", "", "")

	if len(recommendations) == 0 {
		t.Fatal("Expected at least one recommendation")
	}

	// Test the first recommendation structure
	rec := recommendations[0]

	// Verify required fields
	if rec.ID == "" {
		t.Error("Expected non-empty ID")
	}

	if rec.Title == "" {
		t.Error("Expected non-empty Title")
	}

	if rec.Description == "" {
		t.Error("Expected non-empty Description")
	}

	if rec.Priority == "" {
		t.Error("Expected non-empty Priority")
	}

	if rec.Category == "" {
		t.Error("Expected non-empty Category")
	}

	if rec.Timeframe == "" {
		t.Error("Expected non-empty Timeframe")
	}

	if rec.Effort == "" {
		t.Error("Expected non-empty Effort")
	}

	if rec.Cost == "" {
		t.Error("Expected non-empty Cost")
	}

	if rec.Resources == "" {
		t.Error("Expected non-empty Resources")
	}

	if rec.Impact == "" {
		t.Error("Expected non-empty Impact")
	}

	if len(rec.Steps) == 0 {
		t.Error("Expected non-empty Steps")
	}

	// Verify priority values
	validPriorities := []string{"critical", "high", "medium", "low"}
	validPriority := false
	for _, p := range validPriorities {
		if rec.Priority == p {
			validPriority = true
			break
		}
	}
	if !validPriority {
		t.Errorf("Expected valid priority, got %s", rec.Priority)
	}

	// Verify category values
	validCategories := []string{"technical", "process", "training", "documentation", "governance"}
	validCategory := false
	for _, c := range validCategories {
		if rec.Category == c {
			validCategory = true
			break
		}
	}
	if !validCategory {
		t.Errorf("Expected valid category, got %s", rec.Category)
	}

	// Verify timeframe values
	validTimeframes := []string{"immediate", "short", "medium", "long"}
	validTimeframe := false
	for _, tf := range validTimeframes {
		if rec.Timeframe == tf {
			validTimeframe = true
			break
		}
	}
	if !validTimeframe {
		t.Errorf("Expected valid timeframe, got %s", rec.Timeframe)
	}
}

func TestGetImplementationPlan(t *testing.T) {
	handler := NewComplianceGapHandler()
	plan := handler.getImplementationPlan("rec-001")

	// Verify plan structure
	if _, ok := plan["phases"]; !ok {
		t.Error("Expected phases in implementation plan")
	}

	if _, ok := plan["estimated_duration"]; !ok {
		t.Error("Expected estimated_duration in implementation plan")
	}

	if _, ok := plan["success_criteria"]; !ok {
		t.Error("Expected success_criteria in implementation plan")
	}

	// Verify phases is an array
	phases, ok := plan["phases"].([]string)
	if !ok {
		t.Error("Expected phases to be a string array")
	}

	if len(phases) == 0 {
		t.Error("Expected non-empty phases array")
	}

	// Verify success criteria is an array
	criteria, ok := plan["success_criteria"].([]string)
	if !ok {
		t.Error("Expected success_criteria to be a string array")
	}

	if len(criteria) == 0 {
		t.Error("Expected non-empty success_criteria array")
	}
}

func TestGetResourceRequirements(t *testing.T) {
	handler := NewComplianceGapHandler()
	resources := handler.getResourceRequirements("rec-001")

	// Verify resources structure
	if _, ok := resources["team_members"]; !ok {
		t.Error("Expected team_members in resource requirements")
	}

	if _, ok := resources["external_resources"]; !ok {
		t.Error("Expected external_resources in resource requirements")
	}

	if _, ok := resources["budget_breakdown"]; !ok {
		t.Error("Expected budget_breakdown in resource requirements")
	}

	// Verify team_members is an array
	teamMembers, ok := resources["team_members"].([]string)
	if !ok {
		t.Error("Expected team_members to be a string array")
	}

	if len(teamMembers) == 0 {
		t.Error("Expected non-empty team_members array")
	}

	// Verify external_resources is an array
	externalResources, ok := resources["external_resources"].([]string)
	if !ok {
		t.Error("Expected external_resources to be a string array")
	}

	if len(externalResources) == 0 {
		t.Error("Expected non-empty external_resources array")
	}

	// Verify budget_breakdown is a map
	budgetBreakdown, ok := resources["budget_breakdown"].(map[string]string)
	if !ok {
		t.Error("Expected budget_breakdown to be a map[string]string")
	}

	if len(budgetBreakdown) == 0 {
		t.Error("Expected non-empty budget_breakdown map")
	}
}

func TestGetImplementationTimeline(t *testing.T) {
	handler := NewComplianceGapHandler()
	timeline := handler.getImplementationTimeline("rec-001")

	// Verify timeline structure
	if _, ok := timeline["milestones"]; !ok {
		t.Error("Expected milestones in implementation timeline")
	}

	if _, ok := timeline["total_duration"]; !ok {
		t.Error("Expected total_duration in implementation timeline")
	}

	// Verify milestones is an array
	milestones, ok := timeline["milestones"].([]map[string]interface{})
	if !ok {
		t.Error("Expected milestones to be an array of maps")
	}

	if len(milestones) == 0 {
		t.Error("Expected non-empty milestones array")
	}

	// Verify milestone structure
	for i, milestone := range milestones {
		if _, ok := milestone["name"]; !ok {
			t.Errorf("Expected name in milestone %d", i)
		}
		if _, ok := milestone["date"]; !ok {
			t.Errorf("Expected date in milestone %d", i)
		}
		if _, ok := milestone["description"]; !ok {
			t.Errorf("Expected description in milestone %d", i)
		}
	}
}
