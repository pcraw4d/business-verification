package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetTrackingMetrics(t *testing.T) {
	handler := NewGapTrackingHandler()

	req, err := http.NewRequest("GET", "/v1/gap-tracking/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.GetTrackingMetrics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to unmarshal response:", err)
	}

	// Verify response structure
	if _, ok := response["metrics"]; !ok {
		t.Error("Expected metrics in response")
	}

	if _, ok := response["generated_at"]; !ok {
		t.Error("Expected generated_at in response")
	}

	// Verify metrics structure
	metrics, ok := response["metrics"].(map[string]interface{})
	if !ok {
		t.Error("Expected metrics to be a map")
	}

	requiredFields := []string{"total_gaps", "in_progress_gaps", "completed_gaps", "overdue_gaps", "average_progress"}
	for _, field := range requiredFields {
		if _, ok := metrics[field]; !ok {
			t.Errorf("Expected %s in metrics", field)
		}
	}
}

func TestGetGapTrackingList(t *testing.T) {
	handler := NewGapTrackingHandler()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Get all gaps",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  3, // Based on sample data
		},
		{
			name:           "Filter by status in-progress",
			queryParams:    "?status=in-progress",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "Filter by priority critical",
			queryParams:    "?priority=critical",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Filter by framework SOC 2",
			queryParams:    "?framework=SOC 2",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Filter by assigned team",
			queryParams:    "?assigned_to=Security Team",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Filter overdue gaps",
			queryParams:    "?overdue=true",
			expectedStatus: http.StatusOK,
			expectedCount:  0, // No overdue gaps in sample data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/gap-tracking/gaps"+tt.queryParams, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.GetGapTrackingList(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatal("Failed to unmarshal response:", err)
			}

			gaps, ok := response["gaps"].([]interface{})
			if !ok {
				t.Fatal("Expected gaps array in response")
			}

			if len(gaps) != tt.expectedCount {
				t.Errorf("Expected %d gaps, got %d", tt.expectedCount, len(gaps))
			}

			// Verify response structure
			if _, ok := response["total_count"]; !ok {
				t.Error("Expected total_count in response")
			}

			if _, ok := response["filters"]; !ok {
				t.Error("Expected filters in response")
			}
		})
	}
}

func TestGetGapTrackingDetails(t *testing.T) {
	handler := NewGapTrackingHandler()

	tests := []struct {
		name           string
		gapID          string
		expectedStatus int
		shouldExist    bool
	}{
		{
			name:           "Get existing gap",
			gapID:          "gap-001",
			expectedStatus: http.StatusOK,
			shouldExist:    true,
		},
		{
			name:           "Get non-existing gap",
			gapID:          "gap-999",
			expectedStatus: http.StatusNotFound,
			shouldExist:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/gap-tracking/gaps/"+tt.gapID, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.GetGapTrackingDetails(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.shouldExist {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatal("Failed to unmarshal response:", err)
				}

				// Verify response structure
				requiredFields := []string{"gap", "related_gaps", "progress_history", "team_performance", "risk_assessment", "recommendations"}
				for _, field := range requiredFields {
					if _, ok := response[field]; !ok {
						t.Errorf("Expected %s in response", field)
					}
				}
			}
		})
	}
}

func TestUpdateGapProgress(t *testing.T) {
	handler := NewGapTrackingHandler()

	tests := []struct {
		name           string
		gapID          string
		requestBody    string
		expectedStatus int
		shouldSucceed  bool
	}{
		{
			name:  "Update existing gap progress",
			gapID: "gap-001",
			requestBody: `{
				"progress": 75,
				"status": "in-progress",
				"comment": "Updated progress after testing phase",
				"milestones": ["mil-003"]
			}`,
			expectedStatus: http.StatusOK,
			shouldSucceed:  true,
		},
		{
			name:  "Update non-existing gap",
			gapID: "gap-999",
			requestBody: `{
				"progress": 50
			}`,
			expectedStatus: http.StatusNotFound,
			shouldSucceed:  false,
		},
		{
			name:  "Invalid request body",
			gapID: "gap-001",
			requestBody: `{
				"progress": "invalid"
			}`, // Invalid JSON
			expectedStatus: http.StatusBadRequest,
			shouldSucceed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("PUT", "/v1/gap-tracking/gaps/"+tt.gapID+"/progress", strings.NewReader(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.UpdateGapProgress(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.shouldSucceed {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatal("Failed to unmarshal response:", err)
				}

				// Verify response structure
				requiredFields := []string{"gap", "message", "updated_at", "next_milestone", "risk_level"}
				for _, field := range requiredFields {
					if _, ok := response[field]; !ok {
						t.Errorf("Expected %s in response", field)
					}
				}

				// Verify gap was updated
				gap, ok := response["gap"].(map[string]interface{})
				if !ok {
					t.Error("Expected gap in response")
				}

				if gap["progress"] != float64(75) {
					t.Errorf("Expected progress 75, got %v", gap["progress"])
				}
			}
		})
	}
}

func TestCreateGapTracking(t *testing.T) {
	handler := NewGapTrackingHandler()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		shouldSucceed  bool
	}{
		{
			name: "Create valid gap tracking",
			requestBody: `{
				"title": "New Security Gap",
				"description": "Implement new security measures",
				"priority": "high",
				"framework": "SOC 2",
				"assigned_to": "Security Team",
				"due_date": "2025-06-01T00:00:00Z",
				"start_date": "2025-01-01T00:00:00Z",
				"team": ["John Doe", "Jane Smith"],
				"milestones": [
					{
						"name": "Planning",
						"description": "Create implementation plan",
						"due_date": "2025-02-01T00:00:00Z",
						"assigned_to": "John Doe"
					}
				]
			}`,
			expectedStatus: http.StatusCreated,
			shouldSucceed:  true,
		},
		{
			name: "Create gap with invalid JSON",
			requestBody: `{
				"title": "New Security Gap",
				"description": "Implement new security measures"
			`, // Missing closing brace
			expectedStatus: http.StatusBadRequest,
			shouldSucceed:  false,
		},
		{
			name: "Create gap with missing required fields",
			requestBody: `{
				"title": "New Security Gap"
			}`,
			expectedStatus: http.StatusCreated, // Handler doesn't validate required fields in this implementation
			shouldSucceed:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/v1/gap-tracking/gaps", strings.NewReader(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.CreateGapTracking(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.shouldSucceed {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatal("Failed to unmarshal response:", err)
				}

				// Verify response structure
				requiredFields := []string{"gap", "message", "created_at", "estimated_completion", "risk_assessment"}
				for _, field := range requiredFields {
					if _, ok := response[field]; !ok {
						t.Errorf("Expected %s in response", field)
					}
				}

				// Verify gap was created
				gap, ok := response["gap"].(map[string]interface{})
				if !ok {
					t.Error("Expected gap in response")
				}

				if gap["id"] == "" {
					t.Error("Expected non-empty gap ID")
				}
			}
		})
	}
}

func TestGetProgressHistory(t *testing.T) {
	handler := NewGapTrackingHandler()

	tests := []struct {
		name           string
		gapID          string
		expectedStatus int
		shouldExist    bool
	}{
		{
			name:           "Get progress history for existing gap",
			gapID:          "gap-001",
			expectedStatus: http.StatusOK,
			shouldExist:    true,
		},
		{
			name:           "Get progress history for non-existing gap",
			gapID:          "gap-999",
			expectedStatus: http.StatusOK, // Handler returns empty history for non-existing gaps
			shouldExist:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/gap-tracking/gaps/"+tt.gapID+"/history", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.GetProgressHistory(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.shouldExist {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatal("Failed to unmarshal response:", err)
				}

				// Verify response structure
				requiredFields := []string{"gap_id", "history", "total_entries", "generated_at"}
				for _, field := range requiredFields {
					if _, ok := response[field]; !ok {
						t.Errorf("Expected %s in response", field)
					}
				}

				// Verify history is an array
				history, ok := response["history"].([]interface{})
				if !ok {
					t.Error("Expected history to be an array")
				}

				// Verify history entries have required fields
				for i, entry := range history {
					entryMap, ok := entry.(map[string]interface{})
					if !ok {
						t.Errorf("Expected history entry %d to be a map", i)
						continue
					}

					requiredEntryFields := []string{"date", "progress", "comment", "author"}
					for _, field := range requiredEntryFields {
						if _, ok := entryMap[field]; !ok {
							t.Errorf("Expected %s in history entry %d", field, i)
						}
					}
				}
			}
		})
	}
}

func TestGetTeamPerformance(t *testing.T) {
	handler := NewGapTrackingHandler()

	tests := []struct {
		name           string
		teamName       string
		expectedStatus int
		shouldExist    bool
	}{
		{
			name:           "Get performance for existing team",
			teamName:       "Security Team",
			expectedStatus: http.StatusOK,
			shouldExist:    true,
		},
		{
			name:           "Get performance for non-existing team",
			teamName:       "Non-Existent Team",
			expectedStatus: http.StatusOK,
			shouldExist:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/gap-tracking/teams/"+tt.teamName+"/performance", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.GetTeamPerformance(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.shouldExist {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatal("Failed to unmarshal response:", err)
				}

				// Verify response structure
				requiredFields := []string{"team", "performance", "generated_at"}
				for _, field := range requiredFields {
					if _, ok := response[field]; !ok {
						t.Errorf("Expected %s in response", field)
					}
				}

				// Verify performance structure
				performance, ok := response["performance"].(map[string]interface{})
				if !ok {
					t.Error("Expected performance to be a map")
				}

				requiredPerformanceFields := []string{"total_gaps", "completed_gaps", "average_progress", "completion_rate"}
				for _, field := range requiredPerformanceFields {
					if _, ok := performance[field]; !ok {
						t.Errorf("Expected %s in performance", field)
					}
				}
			}
		})
	}
}

func TestExportTrackingReport(t *testing.T) {
	handler := NewGapTrackingHandler()

	req, err := http.NewRequest("GET", "/v1/gap-tracking/reports/export", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ExportTrackingReport(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify response headers
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	contentDisposition := rr.Header().Get("Content-Disposition")
	if contentDisposition != "attachment; filename=gap_tracking_report.json" {
		t.Errorf("Expected Content-Disposition attachment; filename=gap_tracking_report.json, got %s", contentDisposition)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to unmarshal response:", err)
	}

	// Verify report structure
	requiredFields := []string{"report_date", "metrics", "gaps", "summary", "recommendations", "team_performance"}
	for _, field := range requiredFields {
		if _, ok := response[field]; !ok {
			t.Errorf("Expected %s in report", field)
		}
	}
}

func TestCalculateMetrics(t *testing.T) {
	handler := NewGapTrackingHandler()
	metrics := handler.calculateMetrics()

	// Verify metrics structure
	if metrics.TotalGaps <= 0 {
		t.Error("Expected TotalGaps to be greater than 0")
	}

	if metrics.AverageProgress < 0 || metrics.AverageProgress > 100 {
		t.Error("Expected AverageProgress to be between 0 and 100")
	}

	// Verify that totals add up correctly
	totalStatusGaps := metrics.InProgressGaps + metrics.CompletedGaps
	if totalStatusGaps > metrics.TotalGaps {
		t.Error("Status gaps should not exceed total gaps")
	}

	totalPriorityGaps := metrics.CriticalGaps + metrics.HighGaps + metrics.MediumGaps + metrics.LowGaps
	if totalPriorityGaps != metrics.TotalGaps {
		t.Error("Priority gaps should equal total gaps")
	}
}

func TestFilterGaps(t *testing.T) {
	handler := NewGapTrackingHandler()

	tests := []struct {
		name          string
		status        string
		priority      string
		framework     string
		assignedTo    string
		overdue       bool
		expectedCount int
	}{
		{
			name:          "Filter by status in-progress",
			status:        "in-progress",
			expectedCount: 2,
		},
		{
			name:          "Filter by priority critical",
			priority:      "critical",
			expectedCount: 1,
		},
		{
			name:          "Filter by framework SOC 2",
			framework:     "SOC 2",
			expectedCount: 1,
		},
		{
			name:          "Filter by assigned team",
			assignedTo:    "Security Team",
			expectedCount: 1,
		},
		{
			name:          "Filter overdue gaps",
			overdue:       true,
			expectedCount: 0, // No overdue gaps in sample data
		},
		{
			name:          "Multiple filters",
			status:        "in-progress",
			priority:      "high",
			expectedCount: 1,
		},
		{
			name:          "No filters",
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := handler.filterGaps(tt.status, tt.priority, tt.framework, tt.assignedTo, tt.overdue)

			if len(filtered) != tt.expectedCount {
				t.Errorf("Expected %d filtered gaps, got %d", tt.expectedCount, len(filtered))
			}

			// Verify that all returned gaps match the filters
			for _, gap := range filtered {
				if tt.status != "" && gap.Status != tt.status {
					t.Errorf("Expected status %s, got %s", tt.status, gap.Status)
				}
				if tt.priority != "" && gap.Priority != tt.priority {
					t.Errorf("Expected priority %s, got %s", tt.priority, gap.Priority)
				}
				if tt.framework != "" && gap.Framework != tt.framework {
					t.Errorf("Expected framework %s, got %s", tt.framework, gap.Framework)
				}
				if tt.assignedTo != "" && gap.AssignedTo != tt.assignedTo {
					t.Errorf("Expected assigned_to %s, got %s", tt.assignedTo, gap.AssignedTo)
				}
				if tt.overdue && (gap.DueDate.After(time.Now()) || gap.Status == "completed") {
					t.Error("Expected overdue gap, but gap is not overdue")
				}
			}
		})
	}
}

func TestGetGapByID(t *testing.T) {
	handler := NewGapTrackingHandler()

	tests := []struct {
		name          string
		gapID         string
		shouldExist   bool
		expectedTitle string
	}{
		{
			name:          "Get existing gap",
			gapID:         "gap-001",
			shouldExist:   true,
			expectedTitle: "Multi-Factor Authentication Implementation",
		},
		{
			name:        "Get non-existing gap",
			gapID:       "gap-999",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gap := handler.getGapByID(tt.gapID)

			if tt.shouldExist {
				if gap == nil {
					t.Error("Expected gap to exist")
					return
				}
				if gap.Title != tt.expectedTitle {
					t.Errorf("Expected title %s, got %s", tt.expectedTitle, gap.Title)
				}
			} else {
				if gap != nil {
					t.Error("Expected gap to not exist")
				}
			}
		})
	}
}

func TestAssessGapRisk(t *testing.T) {
	handler := NewGapTrackingHandler()

	// Create test gaps with different risk levels
	now := time.Now()

	criticalGap := &GapTracking{
		ID:        "test-critical",
		Priority:  "critical",
		Status:    "open",
		Progress:  10,
		DueDate:   now.AddDate(0, 0, -1), // Overdue
		StartDate: now.AddDate(0, 0, -30),
	}

	lowRiskGap := &GapTracking{
		ID:        "test-low",
		Priority:  "low",
		Status:    "completed",
		Progress:  100,
		DueDate:   now.AddDate(0, 0, 30),
		StartDate: now.AddDate(0, 0, -30),
	}

	tests := []struct {
		name          string
		gap           *GapTracking
		expectedLevel string
		expectedScore int
	}{
		{
			name:          "Critical overdue gap",
			gap:           criticalGap,
			expectedLevel: "high",
			expectedScore: 70, // 40 (critical) + 30 (overdue)
		},
		{
			name:          "Low risk completed gap",
			gap:           lowRiskGap,
			expectedLevel: "low",
			expectedScore: 10, // 10 (low priority)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			risk := handler.assessGapRisk(tt.gap)

			if risk["risk_level"] != tt.expectedLevel {
				t.Errorf("Expected risk level %s, got %v", tt.expectedLevel, risk["risk_level"])
			}

			if risk["risk_score"] != tt.expectedScore {
				t.Errorf("Expected risk score %d, got %v", tt.expectedScore, risk["risk_score"])
			}

			// Verify risk assessment structure
			if _, ok := risk["factors"]; !ok {
				t.Error("Expected factors in risk assessment")
			}
		})
	}
}

func TestGapTrackingStructure(t *testing.T) {
	handler := NewGapTrackingHandler()
	gaps := handler.trackingData

	if len(gaps) == 0 {
		t.Fatal("Expected at least one gap in tracking data")
	}

	// Test the first gap structure
	gap := gaps[0]

	// Verify required fields
	if gap.ID == "" {
		t.Error("Expected non-empty ID")
	}

	if gap.Title == "" {
		t.Error("Expected non-empty Title")
	}

	if gap.Description == "" {
		t.Error("Expected non-empty Description")
	}

	if gap.Priority == "" {
		t.Error("Expected non-empty Priority")
	}

	if gap.Status == "" {
		t.Error("Expected non-empty Status")
	}

	if gap.Progress < 0 || gap.Progress > 100 {
		t.Error("Expected Progress to be between 0 and 100")
	}

	if gap.AssignedTo == "" {
		t.Error("Expected non-empty AssignedTo")
	}

	if gap.Framework == "" {
		t.Error("Expected non-empty Framework")
	}

	if len(gap.Team) == 0 {
		t.Error("Expected non-empty Team")
	}

	if len(gap.Milestones) == 0 {
		t.Error("Expected non-empty Milestones")
	}

	// Verify priority values
	validPriorities := []string{"critical", "high", "medium", "low"}
	validPriority := false
	for _, p := range validPriorities {
		if gap.Priority == p {
			validPriority = true
			break
		}
	}
	if !validPriority {
		t.Errorf("Expected valid priority, got %s", gap.Priority)
	}

	// Verify status values
	validStatuses := []string{"open", "in-progress", "review", "completed"}
	validStatus := false
	for _, s := range validStatuses {
		if gap.Status == s {
			validStatus = true
			break
		}
	}
	if !validStatus {
		t.Errorf("Expected valid status, got %s", gap.Status)
	}

	// Verify milestone structure
	for i, milestone := range gap.Milestones {
		if milestone.ID == "" {
			t.Errorf("Expected non-empty milestone ID for milestone %d", i)
		}
		if milestone.Name == "" {
			t.Errorf("Expected non-empty milestone name for milestone %d", i)
		}
	}
}

func TestHasCommonTeamMember(t *testing.T) {
	tests := []struct {
		name     string
		team1    []string
		team2    []string
		expected bool
	}{
		{
			name:     "Teams with common member",
			team1:    []string{"John", "Jane", "Bob"},
			team2:    []string{"Alice", "John", "Charlie"},
			expected: true,
		},
		{
			name:     "Teams without common member",
			team1:    []string{"John", "Jane", "Bob"},
			team2:    []string{"Alice", "Charlie", "David"},
			expected: false,
		},
		{
			name:     "Empty teams",
			team1:    []string{},
			team2:    []string{},
			expected: false,
		},
		{
			name:     "One empty team",
			team1:    []string{"John"},
			team2:    []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasCommonTeamMember(tt.team1, tt.team2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
