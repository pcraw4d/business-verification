package compliance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/api/routes"
	"github.com/pcraw4d/business-verification/internal/observability"
	"go.uber.org/zap"
)

// TestComplianceAPIEndpoints tests all compliance API endpoints
func TestComplianceAPIEndpoints(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create test server with all compliance routes
	mux := http.NewServeMux()

	// Register all compliance routes
	routes.RegisterComplianceStatusRoutes(mux, logger)
	routes.RegisterComplianceFrameworkRoutes(mux, logger)
	routes.RegisterComplianceTrackingRoutes(mux, logger)
	routes.RegisterComplianceReportingRoutes(mux, logger)
	routes.RegisterComplianceAlertRoutes(mux, logger)

	// Test cases for each endpoint
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Compliance Status Endpoints
		{
			name:           "GET compliance status",
			method:         "GET",
			path:           "/v1/compliance/status/test-business-123",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve compliance status for a business",
		},
		{
			name:   "PUT compliance status",
			method: "PUT",
			path:   "/v1/compliance/status/test-business-123",
			body: map[string]interface{}{
				"overall_status":   "partial",
				"compliance_score": 0.75,
			},
			expectedStatus: http.StatusOK,
			description:    "Should update compliance status for a business",
		},
		{
			name:           "GET compliance status history",
			method:         "GET",
			path:           "/v1/compliance/status/test-business-123/history",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve compliance status history",
		},

		// Compliance Framework Endpoints
		{
			name:           "GET compliance frameworks",
			method:         "GET",
			path:           "/v1/compliance/frameworks",
			expectedStatus: http.StatusOK,
			description:    "Should list all compliance frameworks",
		},
		{
			name:           "GET specific compliance framework",
			method:         "GET",
			path:           "/v1/compliance/frameworks/SOC2",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve specific compliance framework",
		},
		{
			name:           "GET framework requirements",
			method:         "GET",
			path:           "/v1/compliance/frameworks/SOC2/requirements",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve requirements for a framework",
		},
		{
			name:   "POST compliance assessment",
			method: "POST",
			path:   "/v1/compliance/assessments",
			body: map[string]interface{}{
				"business_id":  "test-business-123",
				"framework_id": "SOC2",
				"assessor_id":  "assessor-456",
			},
			expectedStatus: http.StatusCreated,
			description:    "Should create a new compliance assessment",
		},

		// Compliance Tracking Endpoints
		{
			name:           "GET compliance tracking",
			method:         "GET",
			path:           "/v1/compliance/tracking/test-business-123/SOC2",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve compliance tracking data",
		},
		{
			name:   "PUT compliance tracking",
			method: "PUT",
			path:   "/v1/compliance/tracking/test-business-123/SOC2",
			body: map[string]interface{}{
				"overall_progress": 0.85,
				"compliance_level": "partial",
			},
			expectedStatus: http.StatusOK,
			description:    "Should update compliance tracking data",
		},
		{
			name:           "GET compliance milestones",
			method:         "GET",
			path:           "/v1/compliance/milestones",
			expectedStatus: http.StatusOK,
			description:    "Should list compliance milestones",
		},
		{
			name:   "POST compliance milestone",
			method: "POST",
			path:   "/v1/compliance/milestones",
			body: map[string]interface{}{
				"business_id":  "test-business-123",
				"framework_id": "SOC2",
				"name":         "Initial Assessment",
				"type":         "assessment",
				"target_date":  time.Now().AddDate(0, 0, 30).Format(time.RFC3339),
			},
			expectedStatus: http.StatusCreated,
			description:    "Should create a new compliance milestone",
		},
		{
			name:           "GET progress metrics",
			method:         "GET",
			path:           "/v1/compliance/metrics/test-business-123/SOC2",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve progress metrics",
		},
		{
			name:           "GET compliance trends",
			method:         "GET",
			path:           "/v1/compliance/trends/test-business-123/SOC2",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve compliance trends",
		},

		// Compliance Reporting Endpoints
		{
			name:   "POST generate report",
			method: "POST",
			path:   "/v1/compliance/reports",
			body: map[string]interface{}{
				"business_id":  "test-business-123",
				"framework_id": "SOC2",
				"report_type":  "status",
				"generated_by": "test-user",
			},
			expectedStatus: http.StatusCreated,
			description:    "Should generate a compliance report",
		},
		{
			name:           "GET report templates",
			method:         "GET",
			path:           "/v1/compliance/report-templates",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve report templates",
		},

		// Compliance Alert Endpoints
		{
			name:   "POST create alert",
			method: "POST",
			path:   "/v1/compliance/alerts",
			body: map[string]interface{}{
				"business_id":  "test-business-123",
				"framework_id": "SOC2",
				"alert_type":   "compliance_change",
				"severity":     "high",
				"title":        "Test Alert",
				"description":  "Test alert description",
			},
			expectedStatus: http.StatusCreated,
			description:    "Should create a compliance alert",
		},
		{
			name:           "GET compliance alerts",
			method:         "GET",
			path:           "/v1/compliance/alerts",
			expectedStatus: http.StatusOK,
			description:    "Should list compliance alerts",
		},
		{
			name:   "POST create alert rule",
			method: "POST",
			path:   "/v1/compliance/alert-rules",
			body: map[string]interface{}{
				"name":        "Test Rule",
				"description": "Test alert rule",
				"alert_type":  "compliance_change",
				"severity":    "medium",
				"created_by":  "test-user",
				"conditions": []map[string]interface{}{
					{
						"field":     "compliance_score",
						"operator":  "lt",
						"threshold": 0.5,
					},
				},
				"actions": []map[string]interface{}{
					{
						"type":    "email",
						"config":  map[string]interface{}{"template": "low_compliance"},
						"enabled": true,
					},
				},
			},
			expectedStatus: http.StatusCreated,
			description:    "Should create an alert rule",
		},
		{
			name:   "POST evaluate alert rules",
			method: "POST",
			path:   "/v1/compliance/alerts/evaluate",
			body: map[string]interface{}{
				"business_id":  "test-business-123",
				"framework_id": "SOC2",
			},
			expectedStatus: http.StatusOK,
			description:    "Should evaluate alert rules",
		},
		{
			name:           "GET notifications",
			method:         "GET",
			path:           "/v1/compliance/notifications",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve notifications",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare request body if provided
			var body []byte
			var err error
			if tc.body != nil {
				body, err = json.Marshal(tc.body)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			// Create request
			req, err := http.NewRequest(tc.method, tc.path, bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Add request ID to context
			ctx := context.WithValue(req.Context(), "request_id", "test-request-123")
			req = req.WithContext(ctx)

			// Set content type for POST/PUT requests
			if tc.method == "POST" || tc.method == "PUT" {
				req.Header.Set("Content-Type", "application/json")
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			mux.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s",
					tc.expectedStatus, rr.Code, rr.Body.String())
			}

			// Validate response structure for successful requests
			if rr.Code >= 200 && rr.Code < 300 {
				// Check content type
				contentType := rr.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}

				// Validate JSON response
				var response interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("Invalid JSON response: %v", err)
				}
			}

			t.Logf("✅ %s: %s", tc.name, tc.description)
		})
	}
}

// TestComplianceAPIErrorHandling tests error handling for compliance API endpoints
func TestComplianceAPIErrorHandling(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create test server
	mux := http.NewServeMux()
	routes.RegisterComplianceStatusRoutes(mux, logger)
	routes.RegisterComplianceFrameworkRoutes(mux, logger)
	routes.RegisterComplianceTrackingRoutes(mux, logger)
	routes.RegisterComplianceReportingRoutes(mux, logger)
	routes.RegisterComplianceAlertRoutes(mux, logger)

	// Error test cases
	errorTestCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		{
			name:           "Invalid business ID format",
			method:         "GET",
			path:           "/v1/compliance/status/",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 for missing business ID",
		},
		{
			name:           "Invalid JSON in request body",
			method:         "POST",
			path:           "/v1/compliance/alerts",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 for invalid JSON",
		},
		{
			name:   "Missing required fields",
			method: "POST",
			path:   "/v1/compliance/alerts",
			body: map[string]interface{}{
				"title": "Test Alert",
				// Missing required fields: business_id, framework_id, alert_type, severity
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 for missing required fields",
		},
		{
			name:   "Invalid alert type",
			method: "POST",
			path:   "/v1/compliance/alerts",
			body: map[string]interface{}{
				"business_id":  "test-business-123",
				"framework_id": "SOC2",
				"alert_type":   "invalid_type",
				"severity":     "high",
				"title":        "Test Alert",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 for invalid alert type",
		},
		{
			name:   "Invalid severity level",
			method: "POST",
			path:   "/v1/compliance/alerts",
			body: map[string]interface{}{
				"business_id":  "test-business-123",
				"framework_id": "SOC2",
				"alert_type":   "compliance_change",
				"severity":     "invalid_severity",
				"title":        "Test Alert",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 for invalid severity",
		},
		{
			name:   "Invalid report type",
			method: "POST",
			path:   "/v1/compliance/reports",
			body: map[string]interface{}{
				"business_id":  "test-business-123",
				"framework_id": "SOC2",
				"report_type":  "invalid_report_type",
				"generated_by": "test-user",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 for invalid report type",
		},
		{
			name:           "Non-existent resource",
			method:         "GET",
			path:           "/v1/compliance/alerts/non-existent-alert-id",
			expectedStatus: http.StatusNotFound,
			description:    "Should return 404 for non-existent resource",
		},
	}

	// Run error test cases
	for _, tc := range errorTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare request body if provided
			var body []byte
			var err error
			if tc.body != nil {
				if bodyStr, ok := tc.body.(string); ok {
					body = []byte(bodyStr)
				} else {
					body, err = json.Marshal(tc.body)
					if err != nil {
						t.Fatalf("Failed to marshal request body: %v", err)
					}
				}
			}

			// Create request
			req, err := http.NewRequest(tc.method, tc.path, bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Add request ID to context
			ctx := context.WithValue(req.Context(), "request_id", "test-request-123")
			req = req.WithContext(ctx)

			// Set content type for POST/PUT requests
			if tc.method == "POST" || tc.method == "PUT" {
				req.Header.Set("Content-Type", "application/json")
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			mux.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s",
					tc.expectedStatus, rr.Code, rr.Body.String())
			}

			// Validate error response structure
			if rr.Code >= 400 {
				var errorResponse map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &errorResponse); err != nil {
					t.Errorf("Invalid error response JSON: %v", err)
				}

				// Check error response structure
				if _, ok := errorResponse["error"]; !ok {
					t.Errorf("Error response missing 'error' field")
				}
			}

			t.Logf("✅ %s: %s", tc.name, tc.description)
		})
	}
}

// TestComplianceAPIPerformance tests performance of compliance API endpoints
func TestComplianceAPIPerformance(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create test server
	mux := http.NewServeMux()
	routes.RegisterComplianceStatusRoutes(mux, logger)
	routes.RegisterComplianceFrameworkRoutes(mux, logger)
	routes.RegisterComplianceTrackingRoutes(mux, logger)
	routes.RegisterComplianceReportingRoutes(mux, logger)
	routes.RegisterComplianceAlertRoutes(mux, logger)

	// Performance test cases
	performanceTests := []struct {
		name        string
		method      string
		path        string
		maxDuration time.Duration
		description string
	}{
		{
			name:        "GET compliance status performance",
			method:      "GET",
			path:        "/v1/compliance/status/test-business-123",
			maxDuration: 100 * time.Millisecond,
			description: "Should retrieve compliance status within 100ms",
		},
		{
			name:        "GET compliance frameworks performance",
			method:      "GET",
			path:        "/v1/compliance/frameworks",
			maxDuration: 200 * time.Millisecond,
			description: "Should list frameworks within 200ms",
		},
		{
			name:        "GET compliance tracking performance",
			method:      "GET",
			path:        "/v1/compliance/tracking/test-business-123/SOC2",
			maxDuration: 150 * time.Millisecond,
			description: "Should retrieve tracking data within 150ms",
		},
		{
			name:        "GET compliance alerts performance",
			method:      "GET",
			path:        "/v1/compliance/alerts",
			maxDuration: 200 * time.Millisecond,
			description: "Should list alerts within 200ms",
		},
	}

	// Run performance tests
	for _, pt := range performanceTests {
		t.Run(pt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest(pt.method, pt.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Add request ID to context
			ctx := context.WithValue(req.Context(), "request_id", "test-request-123")
			req = req.WithContext(ctx)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Measure execution time
			start := time.Now()
			mux.ServeHTTP(rr, req)
			duration := time.Since(start)

			// Check performance
			if duration > pt.maxDuration {
				t.Errorf("Request took %v, expected less than %v", duration, pt.maxDuration)
			}

			// Check that request was successful
			if rr.Code < 200 || rr.Code >= 300 {
				t.Errorf("Request failed with status %d: %s", rr.Code, rr.Body.String())
			}

			t.Logf("✅ %s: %s (took %v)", pt.name, pt.description, duration)
		})
	}
}

// TestComplianceAPIConcurrency tests concurrent access to compliance API endpoints
func TestComplianceAPIConcurrency(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create test server
	mux := http.NewServeMux()
	routes.RegisterComplianceStatusRoutes(mux, logger)
	routes.RegisterComplianceFrameworkRoutes(mux, logger)
	routes.RegisterComplianceTrackingRoutes(mux, logger)
	routes.RegisterComplianceReportingRoutes(mux, logger)
	routes.RegisterComplianceAlertRoutes(mux, logger)

	// Test concurrent requests
	requests := 100

	t.Run("Concurrent compliance status requests", func(t *testing.T) {
		// Channel to collect results
		results := make(chan int, requests)

		// Launch concurrent requests
		for i := 0; i < requests; i++ {
			go func(id int) {
				req, _ := http.NewRequest("GET", "/v1/compliance/status/test-business-123", nil)
				ctx := context.WithValue(req.Context(), "request_id", fmt.Sprintf("test-request-%d", id))
				req = req.WithContext(ctx)

				rr := httptest.NewRecorder()
				mux.ServeHTTP(rr, req)
				results <- rr.Code
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 0; i < requests; i++ {
			status := <-results
			if status >= 200 && status < 300 {
				successCount++
			}
		}

		// Validate results
		successRate := float64(successCount) / float64(requests)
		if successRate < 0.95 { // 95% success rate
			t.Errorf("Success rate too low: %.2f%% (%d/%d)", successRate*100, successCount, requests)
		}

		t.Logf("✅ Concurrent requests: %d/%d successful (%.2f%%)", successCount, requests, successRate*100)
	})
}
