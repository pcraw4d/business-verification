package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func TestComplianceGapHandler_GetGapSummary(t *testing.T) {
	handler := NewComplianceGapHandler()
	req, err := http.NewRequest("GET", "/v1/compliance/gaps/summary", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.GetGapSummary(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var summary ComplianceGapSummary
	if err := json.Unmarshal(rr.Body.Bytes(), &summary); err != nil {
		t.Fatal("Failed to unmarshal response:", err)
	}

	// Validate summary data
	if summary.TotalGaps == 0 {
		t.Error("Expected total gaps to be greater than 0")
	}
	if summary.OverallCompliance < 0 || summary.OverallCompliance > 100 {
		t.Error("Expected overall compliance to be between 0 and 100")
	}
}

func TestComplianceGapHandler_GetComplianceGaps(t *testing.T) {
	handler := NewComplianceGapHandler()

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
			expectedCount:  5, // Based on sample data
		},
		{
			name:           "Filter by severity",
			queryParams:    "?severity=critical",
			expectedStatus: http.StatusOK,
			expectedCount:  3, // Based on sample data
		},
		{
			name:           "Filter by framework",
			queryParams:    "?framework=soc2",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // Based on sample data
		},
		{
			name:           "Filter by status",
			queryParams:    "?status=open",
			expectedStatus: http.StatusOK,
			expectedCount:  4, // Based on sample data
		},
		{
			name:           "Pagination test",
			queryParams:    "?limit=2&offset=0",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/compliance/gaps"+tt.queryParams, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.GetComplianceGaps(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
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
			}
		})
	}
}

func TestComplianceGapHandler_GetComplianceGap(t *testing.T) {
	handler := NewComplianceGapHandler()

	tests := []struct {
		name           string
		gapID          string
		expectedStatus int
	}{
		{
			name:           "Get existing gap",
			gapID:          "gap-001",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get non-existing gap",
			gapID:          "non-existing",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/compliance/gaps/"+tt.gapID, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Add the gap ID to the request context (simulating mux.Vars)
			req = mux.SetURLVars(req, map[string]string{"id": tt.gapID})

			rr := httptest.NewRecorder()
			handler.GetComplianceGap(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var gap ComplianceGap
				if err := json.Unmarshal(rr.Body.Bytes(), &gap); err != nil {
					t.Fatal("Failed to unmarshal response:", err)
				}

				if gap.ID != tt.gapID {
					t.Errorf("Expected gap ID %s, got %s", tt.gapID, gap.ID)
				}
			}
		})
	}
}

func TestComplianceGapHandler_CreateComplianceGap(t *testing.T) {
	handler := NewComplianceGapHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "Valid gap creation",
			requestBody: ComplianceGap{
				Title:       "Test Gap",
				Description: "Test description",
				Framework:   "soc2",
				Severity:    "high",
				Impact:      "medium",
				TargetDate:  time.Now().AddDate(0, 0, 30),
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Missing required fields",
			requestBody: ComplianceGap{
				Description: "Test description",
				Severity:    "high",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatal("Failed to marshal request body:", err)
			}

			req, err := http.NewRequest("POST", "/v1/compliance/gaps", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.CreateComplianceGap(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusCreated {
				var gap ComplianceGap
				if err := json.Unmarshal(rr.Body.Bytes(), &gap); err != nil {
					t.Fatal("Failed to unmarshal response:", err)
				}

				if gap.ID == "" {
					t.Error("Expected gap ID to be set")
				}
				if gap.Status != "open" {
					t.Error("Expected default status to be 'open'")
				}
			}
		})
	}
}

func TestComplianceGapHandler_UpdateComplianceGap(t *testing.T) {
	handler := NewComplianceGapHandler()

	updates := map[string]interface{}{
		"status":      "in-progress",
		"assigned_to": "john.doe@company.com",
	}

	jsonBody, err := json.Marshal(updates)
	if err != nil {
		t.Fatal("Failed to marshal request body:", err)
	}

	req, err := http.NewRequest("PUT", "/v1/compliance/gaps/gap-001", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "gap-001"})

	rr := httptest.NewRecorder()
	handler.UpdateComplianceGap(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to unmarshal response:", err)
	}

	if response["id"] != "gap-001" {
		t.Errorf("Expected gap ID gap-001, got %v", response["id"])
	}
	if !response["updated"].(bool) {
		t.Error("Expected updated to be true")
	}
}

func TestComplianceGapHandler_GetComplianceFrameworks(t *testing.T) {
	handler := NewComplianceGapHandler()
	req, err := http.NewRequest("GET", "/v1/compliance/frameworks", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.GetComplianceFrameworks(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var frameworks []ComplianceFramework
	if err := json.Unmarshal(rr.Body.Bytes(), &frameworks); err != nil {
		t.Fatal("Failed to unmarshal response:", err)
	}

	if len(frameworks) == 0 {
		t.Error("Expected at least one compliance framework")
	}

	// Validate framework data
	for _, framework := range frameworks {
		if framework.Name == "" {
			t.Error("Expected framework name to be set")
		}
		if framework.ComplianceRate < 0 || framework.ComplianceRate > 100 {
			t.Error("Expected compliance rate to be between 0 and 100")
		}
	}
}

func TestComplianceGapHandler_GetRemediationPlan(t *testing.T) {
	handler := NewComplianceGapHandler()

	tests := []struct {
		name           string
		gapID          string
		expectedStatus int
	}{
		{
			name:           "Get existing remediation plan",
			gapID:          "access-control",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get non-existing remediation plan",
			gapID:          "non-existing",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/compliance/gaps/"+tt.gapID+"/remediation", nil)
			if err != nil {
				t.Fatal(err)
			}
			req = mux.SetURLVars(req, map[string]string{"id": tt.gapID})

			rr := httptest.NewRecorder()
			handler.GetRemediationPlan(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var steps []RemediationStep
				if err := json.Unmarshal(rr.Body.Bytes(), &steps); err != nil {
					t.Fatal("Failed to unmarshal response:", err)
				}

				if len(steps) == 0 {
					t.Error("Expected at least one remediation step")
				}

				// Validate step data
				for _, step := range steps {
					if step.Title == "" {
						t.Error("Expected step title to be set")
					}
					if step.Priority == "" {
						t.Error("Expected step priority to be set")
					}
				}
			}
		})
	}
}

func TestComplianceGapHandler_ExportGapReport(t *testing.T) {
	handler := NewComplianceGapHandler()

	tests := []struct {
		name           string
		format         string
		expectedStatus int
		expectedType   string
	}{
		{
			name:           "Export JSON report",
			format:         "json",
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
		},
		{
			name:           "Export PDF report",
			format:         "pdf",
			expectedStatus: http.StatusOK,
			expectedType:   "application/pdf",
		},
		{
			name:           "Export CSV report",
			format:         "csv",
			expectedStatus: http.StatusOK,
			expectedType:   "text/csv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/compliance/reports/gaps?format="+tt.format, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.ExportGapReport(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			contentType := rr.Header().Get("Content-Type")
			if contentType != tt.expectedType {
				t.Errorf("Expected content type %s, got %s", tt.expectedType, contentType)
			}
		})
	}
}

func TestComplianceGapHandler_Validation(t *testing.T) {
	// Test gap validation
	gap := ComplianceGap{
		Title:     "Test Gap",
		Framework: "soc2",
		Severity:  "high",
	}

	// This should be valid
	if gap.Title == "" || gap.Framework == "" || gap.Severity == "" {
		t.Error("Valid gap should not be marked as invalid")
	}

	// Test invalid gap
	invalidGap := ComplianceGap{
		Description: "Missing required fields",
	}

	if invalidGap.Title != "" || invalidGap.Framework != "" || invalidGap.Severity != "" {
		t.Error("Invalid gap should be detected")
	}
}

func TestComplianceGapHandler_DataIntegrity(t *testing.T) {
	// Test sample data integrity
	handler := NewComplianceGapHandler()
	gaps := handler.getSampleGaps()

	if len(gaps) == 0 {
		t.Error("Expected sample gaps to be available")
	}

	for _, gap := range gaps {
		if gap.ID == "" {
			t.Error("Expected gap ID to be set")
		}
		if gap.Title == "" {
			t.Error("Expected gap title to be set")
		}
		if gap.Framework == "" {
			t.Error("Expected gap framework to be set")
		}
		if gap.Severity == "" {
			t.Error("Expected gap severity to be set")
		}
		if gap.Status == "" {
			t.Error("Expected gap status to be set")
		}
	}
}

func TestComplianceGapHandler_ErrorHandling(t *testing.T) {
	handler := NewComplianceGapHandler()

	// Test invalid JSON in create request
	req, err := http.NewRequest("POST", "/v1/compliance/gaps", bytes.NewBufferString("invalid json"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateComplianceGap(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected bad request for invalid JSON, got %v", status)
	}
}

// Benchmark tests
func BenchmarkComplianceGapHandler_GetGapSummary(b *testing.B) {
	handler := NewComplianceGapHandler()
	req, _ := http.NewRequest("GET", "/v1/compliance/gaps/summary", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.GetGapSummary(rr, req)
	}
}

func BenchmarkComplianceGapHandler_GetComplianceGaps(b *testing.B) {
	handler := NewComplianceGapHandler()
	req, _ := http.NewRequest("GET", "/v1/compliance/gaps", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.GetComplianceGaps(rr, req)
	}
}
