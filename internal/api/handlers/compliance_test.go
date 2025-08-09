package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/compliance"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// MockCheckEngine is a mock implementation of the CheckEngine for testing
type MockCheckEngine struct {
	logger *observability.Logger
}

func (m *MockCheckEngine) Check(ctx context.Context, req compliance.CheckRequest) (*compliance.CheckResponse, error) {
	// Return a mock response
	return &compliance.CheckResponse{
		BusinessID: req.BusinessID,
		CheckedAt:  time.Now(),
		Results: []compliance.FrameworkCheckResult{
			{
				FrameworkID: "SOC2",
				Summary: compliance.ComplianceCheckResult{
					BusinessID: req.BusinessID,
					Framework:  "SOC2",
					Evaluated:  time.Now(),
					Passed:     8,
					Failed:     2,
					Outcomes:   []compliance.RuleOutcome{},
				},
			},
		},
		Passed: 8,
		Failed: 2,
	}, nil
}

func setupComplianceTest(t *testing.T) (*ComplianceHandler, *compliance.ComplianceStatusSystem) {
	// Create test observability config
	obsConfig := &config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	}
	logger := observability.NewLogger(obsConfig)
	statusSystem := compliance.NewComplianceStatusSystem(logger)

	// Create a mock check engine that returns a valid response
	checkEngine := &MockCheckEngine{
		logger: logger,
	}

	handler := NewComplianceHandler(logger, checkEngine, statusSystem)

	// Initialize test business status
	ctx := context.WithValue(context.Background(), "request_id", "test-request-id")
	err := statusSystem.InitializeBusinessStatus(ctx, "test-business-123")
	if err != nil {
		t.Fatalf("Failed to initialize test business status: %v", err)
	}

	return handler, statusSystem
}

func TestGetComplianceStatusHandler(t *testing.T) {
	handler, _ := setupComplianceTest(t)

	tests := []struct {
		name           string
		businessID     string
		expectedStatus int
	}{
		{
			name:           "Valid business ID",
			businessID:     "test-business-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid business ID",
			businessID:     "non-existent-business",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty business ID",
			businessID:     "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/v1/compliance/status/"+tt.businessID, nil)
			// Set up the URL pattern for path value extraction
			req.URL.Path = "/v1/compliance/status/" + tt.businessID
			req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-id"))

			w := httptest.NewRecorder()
			handler.GetComplianceStatusHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if response["business_id"] != tt.businessID {
					t.Errorf("Expected business_id %s, got %s", tt.businessID, response["business_id"])
				}
			}
		})
	}
}

func TestGetStatusHistoryHandler(t *testing.T) {
	handler, _ := setupComplianceTest(t)

	tests := []struct {
		name           string
		businessID     string
		startDate      string
		endDate        string
		expectedStatus int
	}{
		{
			name:           "Valid request with default dates",
			businessID:     "test-business-123",
			startDate:      "",
			endDate:        "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid request with custom dates",
			businessID:     "test-business-123",
			startDate:      "2024-01-01",
			endDate:        "2024-12-31",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid date format",
			businessID:     "test-business-123",
			startDate:      "invalid-date",
			endDate:        "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent business",
			businessID:     "non-existent-business",
			startDate:      "",
			endDate:        "",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/v1/compliance/status/" + tt.businessID + "/history"
			if tt.startDate != "" {
				url += "?start_date=" + tt.startDate
			}
			if tt.endDate != "" {
				if tt.startDate != "" {
					url += "&"
				} else {
					url += "?"
				}
				url += "end_date=" + tt.endDate
			}

			req := httptest.NewRequest("GET", url, nil)
			req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-id"))

			w := httptest.NewRecorder()
			handler.GetStatusHistoryHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if response["business_id"] != tt.businessID {
					t.Errorf("Expected business_id %s, got %s", tt.businessID, response["business_id"])
				}
			}
		})
	}
}

func TestGetStatusAlertsHandler(t *testing.T) {
	handler, _ := setupComplianceTest(t)

	tests := []struct {
		name           string
		businessID     string
		status         string
		expectedStatus int
	}{
		{
			name:           "Valid request - all alerts",
			businessID:     "test-business-123",
			status:         "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid request - active alerts only",
			businessID:     "test-business-123",
			status:         "active",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Non-existent business",
			businessID:     "non-existent-business",
			status:         "",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/v1/compliance/status/" + tt.businessID + "/alerts"
			if tt.status != "" {
				url += "?status=" + tt.status
			}

			req := httptest.NewRequest("GET", url, nil)
			req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-id"))

			w := httptest.NewRecorder()
			handler.GetStatusAlertsHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if response["business_id"] != tt.businessID {
					t.Errorf("Expected business_id %s, got %s", tt.businessID, response["business_id"])
				}
			}
		})
	}
}

func TestAcknowledgeAlertHandler(t *testing.T) {
	handler, statusSystem := setupComplianceTest(t)

	// Create a test alert first
	ctx := context.WithValue(context.Background(), "request_id", "test-request-id")
	err := statusSystem.CreateStatusAlert(ctx, "test-business-123", "status_change", "medium", "framework", "test-framework", "Test Alert", "Test alert description", 75.0, 80.0)
	if err != nil {
		t.Fatalf("Failed to create test alert: %v", err)
	}

	// Get the alert ID
	alerts, err := statusSystem.GetStatusAlerts(ctx, "test-business-123", "active")
	if err != nil || len(alerts) == 0 {
		t.Fatalf("Failed to get test alert: %v", err)
	}
	alertID := alerts[0].ID

	tests := []struct {
		name           string
		businessID     string
		alertID        string
		acknowledgedBy string
		expectedStatus int
	}{
		{
			name:           "Valid acknowledgment",
			businessID:     "test-business-123",
			alertID:        alertID,
			acknowledgedBy: "test-user",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing acknowledged_by",
			businessID:     "test-business-123",
			alertID:        alertID,
			acknowledgedBy: "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid business ID",
			businessID:     "non-existent-business",
			alertID:        alertID,
			acknowledgedBy: "test-user",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody := map[string]string{
				"acknowledged_by": tt.acknowledgedBy,
			}
			bodyBytes, _ := json.Marshal(requestBody)

			url := "/v1/compliance/status/" + tt.businessID + "/alerts/" + tt.alertID + "/acknowledge"
			req := httptest.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-id"))

			w := httptest.NewRecorder()
			handler.AcknowledgeAlertHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if response["message"] != "Alert acknowledged successfully" {
					t.Errorf("Expected success message, got %s", response["message"])
				}
			}
		})
	}
}

func TestResolveAlertHandler(t *testing.T) {
	handler, statusSystem := setupComplianceTest(t)

	// Create a test alert first
	ctx := context.WithValue(context.Background(), "request_id", "test-request-id")
	err := statusSystem.CreateStatusAlert(ctx, "test-business-123", "status_change", "medium", "framework", "test-framework", "Test Alert", "Test alert description", 75.0, 80.0)
	if err != nil {
		t.Fatalf("Failed to create test alert: %v", err)
	}

	// Get the alert ID
	alerts, err := statusSystem.GetStatusAlerts(ctx, "test-business-123", "active")
	if err != nil || len(alerts) == 0 {
		t.Fatalf("Failed to get test alert: %v", err)
	}
	alertID := alerts[0].ID

	tests := []struct {
		name           string
		businessID     string
		alertID        string
		resolvedBy     string
		expectedStatus int
	}{
		{
			name:           "Valid resolution",
			businessID:     "test-business-123",
			alertID:        alertID,
			resolvedBy:     "test-user",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing resolved_by",
			businessID:     "test-business-123",
			alertID:        alertID,
			resolvedBy:     "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid business ID",
			businessID:     "non-existent-business",
			alertID:        alertID,
			resolvedBy:     "test-user",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody := map[string]string{
				"resolved_by": tt.resolvedBy,
			}
			bodyBytes, _ := json.Marshal(requestBody)

			url := "/v1/compliance/status/" + tt.businessID + "/alerts/" + tt.alertID + "/resolve"
			req := httptest.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-id"))

			w := httptest.NewRecorder()
			handler.ResolveAlertHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if response["message"] != "Alert resolved successfully" {
					t.Errorf("Expected success message, got %s", response["message"])
				}
			}
		})
	}
}

func TestGenerateStatusReportHandler(t *testing.T) {
	handler, _ := setupComplianceTest(t)

	tests := []struct {
		name           string
		businessID     string
		reportType     string
		expectedStatus int
	}{
		{
			name:           "Valid request - summary report",
			businessID:     "test-business-123",
			reportType:     "summary",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid request - detailed report",
			businessID:     "test-business-123",
			reportType:     "detailed",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid request - default report type",
			businessID:     "test-business-123",
			reportType:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Non-existent business",
			businessID:     "non-existent-business",
			reportType:     "summary",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody := map[string]string{
				"report_type": tt.reportType,
			}
			bodyBytes, _ := json.Marshal(requestBody)

			url := "/v1/compliance/status/" + tt.businessID + "/report"
			req := httptest.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-id"))

			w := httptest.NewRecorder()
			handler.GenerateStatusReportHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if response["business_id"] != tt.businessID {
					t.Errorf("Expected business_id %s, got %s", tt.businessID, response["business_id"])
				}

				expectedReportType := tt.reportType
				if expectedReportType == "" {
					expectedReportType = "summary"
				}
				if response["report_type"] != expectedReportType {
					t.Errorf("Expected report_type %s, got %s", expectedReportType, response["report_type"])
				}
			}
		})
	}
}

func TestInitializeBusinessStatusHandler(t *testing.T) {
	handler, _ := setupComplianceTest(t)

	tests := []struct {
		name           string
		businessID     string
		expectedStatus int
	}{
		{
			name:           "Valid business ID",
			businessID:     "new-business-456",
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
			url := "/v1/compliance/status/" + tt.businessID + "/initialize"
			req := httptest.NewRequest("POST", url, nil)
			req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-id"))

			w := httptest.NewRecorder()
			handler.InitializeBusinessStatusHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if response["message"] != "Business compliance status initialized successfully" {
					t.Errorf("Expected success message, got %s", response["message"])
				}
			}
		})
	}
}

func TestCheckComplianceHandler(t *testing.T) {
	handler, _ := setupComplianceTest(t)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Valid request",
			requestBody: map[string]interface{}{
				"business_id":   "test-business-123",
				"frameworks":    []string{"SOC2", "PCI-DSS"},
				"apply_effects": true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Missing business_id",
			requestBody: map[string]interface{}{
				"frameworks":    []string{"SOC2"},
				"apply_effects": false,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid JSON",
			requestBody: map[string]interface{}{
				"business_id": "test-business-123",
				"frameworks":  "invalid", // Should be array
			},
			expectedStatus: http.StatusBadRequest, // Invalid JSON should return 400
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.requestBody)

			req := httptest.NewRequest("POST", "/v1/compliance/check", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-id"))

			w := httptest.NewRecorder()
			handler.CheckComplianceHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
