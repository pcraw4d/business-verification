package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/compliance"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAuditHandler_RecordAuditEvent(t *testing.T) {
	// Create test logger
	obsConfig := &config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	}
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)

	// Create real audit system for testing
	auditSystem := compliance.NewComplianceAuditSystem()
	handler := NewAuditHandler(auditSystem, logger)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Valid audit event",
			requestBody: RecordAuditEventRequest{
				BusinessID:    "business-123",
				EventType:     "compliance_check",
				EventCategory: "compliance",
				EntityType:    "compliance",
				EntityID:      "SOC2",
				Action:        compliance.AuditActionCreate,
				Description:   "SOC 2 compliance check initiated",
				UserID:        "user-123",
				UserName:      "John Doe",
				UserRole:      "compliance_officer",
				UserEmail:     "john.doe@example.com",
				Severity:      "medium",
				Impact:        "high",
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "Missing required fields",
			requestBody: RecordAuditEventRequest{
				BusinessID: "business-123",
				// Missing other required fields
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/v1/audit/events", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.RecordAuditEvent(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectedError {
				var response RecordAuditEventResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.EventID == "" {
					t.Error("Expected event ID to be set")
				}

				if response.Timestamp.IsZero() {
					t.Error("Expected timestamp to be set")
				}

				if !response.Success {
					t.Error("Expected success to be true")
				}
			}
		})
	}
}

func TestAuditHandler_GetAuditEvents(t *testing.T) {
	// Create test logger
	obsConfig := &config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	}
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)

	// Create real audit system for testing
	auditSystem := compliance.NewComplianceAuditSystem()
	handler := NewAuditHandler(auditSystem, logger)

	// Add a test event first
	testEvent := &compliance.AuditEvent{
		ID:            "test-event-1",
		BusinessID:    "business-123",
		EventType:     "compliance_check",
		EventCategory: "compliance",
		EntityType:    "compliance",
		EntityID:      "SOC2",
		Action:        compliance.AuditActionCreate,
		Description:   "SOC 2 compliance check initiated",
		UserID:        "user-123",
		UserName:      "John Doe",
		UserRole:      "compliance_officer",
		UserEmail:     "john.doe@example.com",
		Timestamp:     time.Now(),
		Success:       true,
		Severity:      "medium",
		Impact:        "high",
	}

	// Record the test event
	err := auditSystem.RecordAuditEvent(context.Background(), testEvent)
	if err != nil {
		t.Fatalf("Failed to record test event: %v", err)
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedEvents int
	}{
		{
			name:           "Get audit events with business ID",
			queryParams:    "business_id=business-123",
			expectedStatus: http.StatusOK,
			expectedEvents: 1,
		},
		{
			name:           "Missing business ID",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectedEvents: 0,
		},
		{
			name:           "With filters",
			queryParams:    "business_id=business-123&event_types=compliance_check&severities=medium",
			expectedStatus: http.StatusOK,
			expectedEvents: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/v1/audit/events"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetAuditEvents(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response GetAuditEventsResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if len(response.Events) != tt.expectedEvents {
					t.Errorf("Expected %d events, got %d", tt.expectedEvents, len(response.Events))
				}

				if response.Meta.Total != tt.expectedEvents {
					t.Errorf("Expected total %d, got %d", tt.expectedEvents, response.Meta.Total)
				}
			}
		})
	}
}

func TestAuditHandler_GetAuditTrail(t *testing.T) {
	// Create test audit handler
	handler := &AuditHandler{
		auditService: &MockAuditService{},
		logger:       &MockLogger{},
	}

	// Test successful audit trail retrieval
	req := httptest.NewRequest("GET", "/audit/trail?limit=10&offset=0", nil)
	w := httptest.NewRecorder()

	handler.GetAuditTrail(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Contains(t, response, "audit_entries")
	assert.Contains(t, response, "total_count")
}

func TestAuditHandler_GetAuditTrailWithFilters(t *testing.T) {
	// Create test audit handler
	handler := &AuditHandler{
		auditService: &MockAuditService{},
		logger:       &MockLogger{},
	}

	// Test audit trail with filters
	req := httptest.NewRequest("GET", "/audit/trail?action=create&resource_type=business&user_id=test-user", nil)
	w := httptest.NewRecorder()

	handler.GetAuditTrail(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Contains(t, response, "audit_entries")
}

func TestAuditHandler_GenerateAuditReport(t *testing.T) {
	// Create test audit handler
	handler := &AuditHandler{
		auditService: &MockAuditService{},
		logger:       &MockLogger{},
	}

	// Test successful audit report generation
	req := httptest.NewRequest("POST", "/audit/report", strings.NewReader(`{
		"start_date": "2024-01-01",
		"end_date": "2024-12-31",
		"format": "json",
		"include_details": true
	}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.GenerateAuditReport(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Contains(t, response, "report_id")
	assert.Contains(t, response, "status")
}

func TestAuditHandler_GenerateAuditReportInvalidRequest(t *testing.T) {
	// Create test audit handler
	handler := &AuditHandler{
		auditService: &MockAuditService{},
		logger:       &MockLogger{},
	}

	// Test invalid request
	req := httptest.NewRequest("POST", "/audit/report", strings.NewReader(`{
		"invalid": "request"
	}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.GenerateAuditReport(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuditHandler_GetAuditMetrics(t *testing.T) {
	// Create test logger
	obsConfig := &config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	}
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)

	// Create real audit system for testing
	auditSystem := compliance.NewComplianceAuditSystem()
	handler := NewAuditHandler(auditSystem, logger)

	tests := []struct {
		name           string
		businessID     string
		expectedStatus int
	}{
		{
			name:           "Get audit metrics for valid business",
			businessID:     "business-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get audit metrics for non-existent business",
			businessID:     "non-existent",
			expectedStatus: http.StatusOK, // Should return empty metrics
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/v1/audit/metrics/" + tt.businessID
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetAuditMetrics(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response GetAuditMetricsResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Metrics == nil {
					t.Error("Expected metrics to be present")
				}

				if response.Metrics.BusinessID != tt.businessID {
					t.Errorf("Expected business ID %s, got %s", tt.businessID, response.Metrics.BusinessID)
				}
			}
		})
	}
}

// MockAuditSystem implements a mock audit system for testing
