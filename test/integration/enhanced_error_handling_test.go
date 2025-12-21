//go:build !comprehensive_test && !e2e_railway
// +build !comprehensive_test,!e2e_railway

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"kyb-platform/internal/shared"
	"kyb-platform/test/mocks"
)

// TestEnhancedErrorScenarios tests comprehensive error scenarios for the classification system
func TestEnhancedErrorScenarios(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	// Test 1: Database Connection Failures
	t.Run("DatabaseConnectionFailures", func(t *testing.T) {
		testDatabaseConnectionFailures(t)
	})

	// Test 2: Network Timeout Scenarios
	t.Run("NetworkTimeoutScenarios", func(t *testing.T) {
		testNetworkTimeoutScenarios(t)
	})

	// Test 3: Invalid Data Handling
	t.Run("InvalidDataHandling", func(t *testing.T) {
		testInvalidDataHandling(t)
	})

	// Test 4: Service Unavailability
	t.Run("ServiceUnavailability", func(t *testing.T) {
		testServiceUnavailability(t)
	})

	// Test 5: Memory Exhaustion Scenarios
	t.Run("MemoryExhaustionScenarios", func(t *testing.T) {
		testMemoryExhaustionScenarios(t)
	})

	// Test 6: Concurrent Access Issues
	t.Run("ConcurrentAccessIssues", func(t *testing.T) {
		testConcurrentAccessIssues(t)
	})

	// Test 7: External Service Failures
	t.Run("ExternalServiceFailures", func(t *testing.T) {
		testExternalServiceFailures(t)
	})

	// Test 8: Data Corruption Scenarios
	t.Run("DataCorruptionScenarios", func(t *testing.T) {
		testDataCorruptionScenarios(t)
	})
}

// testDatabaseConnectionFailures tests various database connection failure scenarios
func testDatabaseConnectionFailures(t *testing.T) {
	testCases := []struct {
		name           string
		errorType      string
		expectedStatus int
		description    string
	}{
		{
			name:           "Connection Pool Exhaustion",
			errorType:      "pool_exhaustion",
			expectedStatus: http.StatusServiceUnavailable,
			description:    "Database connection pool exhausted",
		},
		{
			name:           "Connection Timeout",
			errorType:      "connection_timeout",
			expectedStatus: http.StatusGatewayTimeout,
			description:    "Database connection timeout",
		},
		{
			name:           "Database Unavailable",
			errorType:      "database_unavailable",
			expectedStatus: http.StatusServiceUnavailable,
			description:    "Database service unavailable",
		},
		{
			name:           "Transaction Deadlock",
			errorType:      "deadlock",
			expectedStatus: http.StatusConflict,
			description:    "Database transaction deadlock",
		},
		{
			name:           "Query Timeout",
			errorType:      "query_timeout",
			expectedStatus: http.StatusGatewayTimeout,
			description:    "Database query timeout",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock database with specific error
			mockDB := mocks.NewMockDatabase()
			mockDB.SetConnectionError(fmt.Errorf("database error: %s", tc.errorType))

			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handleDatabaseErrorScenario(w, r, mockDB, tc.errorType)
			}))
			defer server.Close()

			// Test database error scenario
			resp, err := http.Post(server.URL+"/v1/classify", "application/json",
				createValidRequest())
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Validate error response
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d for %s, got %d", tc.expectedStatus, tc.name, resp.StatusCode)
			}

			// Validate error response structure
			var errorResponse map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			// Validate error response contains expected fields
			validateErrorResponse(t, errorResponse, tc.errorType)

			t.Logf("✅ %s test passed - Status: %d, Error: %v", tc.name, resp.StatusCode, errorResponse["error"])
		})
	}
}

// testNetworkTimeoutScenarios tests various network timeout scenarios
func testNetworkTimeoutScenarios(t *testing.T) {
	testCases := []struct {
		name           string
		timeout        time.Duration
		expectedStatus int
		description    string
	}{
		{
			name:           "Short Timeout",
			timeout:        100 * time.Millisecond,
			expectedStatus: http.StatusRequestTimeout,
			description:    "Request timeout with short duration",
		},
		{
			name:           "Medium Timeout",
			timeout:        500 * time.Millisecond,
			expectedStatus: http.StatusRequestTimeout,
			description:    "Request timeout with medium duration",
		},
		{
			name:           "Long Timeout",
			timeout:        2 * time.Second,
			expectedStatus: http.StatusRequestTimeout,
			description:    "Request timeout with long duration",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock service with delay
			mockService := mocks.NewMockClassificationService()
			mockService.SetDelay(tc.timeout + 100*time.Millisecond) // Exceed timeout

			// Create test server with timeout
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handleTimeoutScenario(w, r, mockService, tc.timeout)
			}))
			defer server.Close()

			// Test timeout scenario
			resp, err := http.Post(server.URL+"/v1/classify", "application/json",
				createValidRequest())
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Validate timeout response
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d for %s, got %d", tc.expectedStatus, tc.name, resp.StatusCode)
			}

			// Validate timeout response structure
			var errorResponse map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			// Validate timeout response contains expected fields
			validateTimeoutResponse(t, errorResponse)

			t.Logf("✅ %s test passed - Status: %d, Timeout: %v", tc.name, resp.StatusCode, tc.timeout)
		})
	}
}

// testInvalidDataHandling tests various invalid data scenarios
func testInvalidDataHandling(t *testing.T) {
	testCases := []struct {
		name           string
		requestBody    string
		expectedStatus int
		description    string
	}{
		{
			name:           "Malformed JSON",
			requestBody:    `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
			description:    "Malformed JSON should return 400",
		},
		{
			name:           "Invalid Business Name Format",
			requestBody:    `{"id": "test-001", "business_name": "!@#$%^&*()"}`,
			expectedStatus: http.StatusBadRequest,
			description:    "Invalid business name format should return 400",
		},
		{
			name:           "SQL Injection Attempt",
			requestBody:    `{"id": "test-001", "business_name": "'; DROP TABLE users; --"}`,
			expectedStatus: http.StatusBadRequest,
			description:    "SQL injection attempt should return 400",
		},
		{
			name:           "XSS Attempt",
			requestBody:    `{"id": "test-001", "business_name": "<script>alert('xss')</script>"}`,
			expectedStatus: http.StatusBadRequest,
			description:    "XSS attempt should return 400",
		},
		{
			name:           "Oversized Request",
			requestBody:    createOversizedRequest(),
			expectedStatus: http.StatusRequestEntityTooLarge,
			description:    "Oversized request should return 413",
		},
		{
			name:           "Invalid URL Format",
			requestBody:    `{"id": "test-001", "business_name": "Test Company", "website_url": "not-a-valid-url"}`,
			expectedStatus: http.StatusBadRequest,
			description:    "Invalid URL format should return 400",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handleInvalidDataScenario(w, r)
			}))
			defer server.Close()

			// Test invalid data scenario
			resp, err := http.Post(server.URL+"/v1/classify", "application/json",
				strings.NewReader(tc.requestBody))
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Validate error response
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d for %s, got %d", tc.expectedStatus, tc.name, resp.StatusCode)
			}

			// Validate error response structure
			var errorResponse map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			// Validate error response contains expected fields
			validateErrorResponse(t, errorResponse, "validation_error")

			t.Logf("✅ %s test passed - Status: %d, Error: %v", tc.name, resp.StatusCode, errorResponse["error"])
		})
	}
}

// testServiceUnavailability tests various service unavailability scenarios
func testServiceUnavailability(t *testing.T) {
	testCases := []struct {
		name           string
		serviceType    string
		expectedStatus int
		description    string
	}{
		{
			name:           "Classification Service Down",
			serviceType:    "classification",
			expectedStatus: http.StatusServiceUnavailable,
			description:    "Classification service unavailable",
		},
		{
			name:           "Risk Assessment Service Down",
			serviceType:    "risk_assessment",
			expectedStatus: http.StatusServiceUnavailable,
			description:    "Risk assessment service unavailable",
		},
		{
			name:           "ML Service Down",
			serviceType:    "ml_service",
			expectedStatus: http.StatusServiceUnavailable,
			description:    "ML service unavailable",
		},
		{
			name:           "Database Service Down",
			serviceType:    "database",
			expectedStatus: http.StatusServiceUnavailable,
			description:    "Database service unavailable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock service configured to fail
			mockService := mocks.NewMockClassificationService()
			mockService.SetFailureMode(true, fmt.Sprintf("%s service unavailable", tc.serviceType))

			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handleServiceUnavailabilityScenario(w, r, mockService, tc.serviceType)
			}))
			defer server.Close()

			// Test service unavailability scenario
			resp, err := http.Post(server.URL+"/v1/classify", "application/json",
				createValidRequest())
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Validate error response
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d for %s, got %d", tc.expectedStatus, tc.name, resp.StatusCode)
			}

			// Validate error response structure
			var errorResponse map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			// Validate error response contains expected fields
			validateErrorResponse(t, errorResponse, "service_unavailable")

			t.Logf("✅ %s test passed - Status: %d, Error: %v", tc.name, resp.StatusCode, errorResponse["error"])
		})
	}
}

// testMemoryExhaustionScenarios tests memory exhaustion scenarios
func testMemoryExhaustionScenarios(t *testing.T) {
	testCases := []struct {
		name           string
		requestSize    int
		expectedStatus int
		description    string
	}{
		{
			name:           "Large Request",
			requestSize:    1024 * 1024, // 1MB
			expectedStatus: http.StatusRequestEntityTooLarge,
			description:    "Large request should be rejected",
		},
		{
			name:           "Very Large Request",
			requestSize:    10 * 1024 * 1024, // 10MB
			expectedStatus: http.StatusRequestEntityTooLarge,
			description:    "Very large request should be rejected",
		},
		{
			name:           "Extremely Large Request",
			requestSize:    100 * 1024 * 1024, // 100MB
			expectedStatus: http.StatusRequestEntityTooLarge,
			description:    "Extremely large request should be rejected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handleMemoryExhaustionScenario(w, r, tc.requestSize)
			}))
			defer server.Close()

			// Create large request
			largeRequest := createLargeRequestWithSize(tc.requestSize)

			// Test memory exhaustion scenario
			resp, err := http.Post(server.URL+"/v1/classify", "application/json",
				strings.NewReader(largeRequest))
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Validate error response
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d for %s, got %d", tc.expectedStatus, tc.name, resp.StatusCode)
			}

			// Validate error response structure
			var errorResponse map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			// Validate error response contains expected fields
			validateErrorResponse(t, errorResponse, "request_too_large")

			t.Logf("✅ %s test passed - Status: %d, Size: %d bytes", tc.name, resp.StatusCode, tc.requestSize)
		})
	}
}

// testConcurrentAccessIssues tests concurrent access scenarios
func testConcurrentAccessIssues(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleConcurrentAccessScenario(w, r)
	}))
	defer server.Close()

	// Test concurrent access
	concurrency := 50
	successCount := 0
	errorCount := 0

	// Create channel to collect results
	results := make(chan int, concurrency)

	// Launch concurrent requests
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			resp, err := http.Post(server.URL+"/v1/classify", "application/json",
				createValidRequest())
			if err != nil {
				results <- -1 // Error
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				results <- 1 // Success
			} else {
				results <- 0 // Error
			}
		}(i)
	}

	// Collect results
	for i := 0; i < concurrency; i++ {
		result := <-results
		if result == 1 {
			successCount++
		} else if result == 0 {
			errorCount++
		}
	}

	// Validate concurrent access handling
	if successCount == 0 {
		t.Error("Expected some requests to succeed under concurrent access")
	}

	if errorCount > concurrency/2 {
		t.Errorf("Too many errors under concurrent access: %d/%d", errorCount, concurrency)
	}

	t.Logf("✅ Concurrent access test passed - Success: %d, Errors: %d, Total: %d",
		successCount, errorCount, concurrency)
}

// testExternalServiceFailures tests external service failure scenarios
func testExternalServiceFailures(t *testing.T) {
	testCases := []struct {
		name           string
		serviceName    string
		expectedStatus int
		description    string
	}{
		{
			name:           "Website Scraping Service Failure",
			serviceName:    "website_scraper",
			expectedStatus: http.StatusServiceUnavailable,
			description:    "Website scraping service failure",
		},
		{
			name:           "External API Failure",
			serviceName:    "external_api",
			expectedStatus: http.StatusServiceUnavailable,
			description:    "External API service failure",
		},
		{
			name:           "Notification Service Failure",
			serviceName:    "notification",
			expectedStatus: http.StatusServiceUnavailable,
			description:    "Notification service failure",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handleExternalServiceFailureScenario(w, r, tc.serviceName)
			}))
			defer server.Close()

			// Test external service failure scenario
			resp, err := http.Post(server.URL+"/v1/classify", "application/json",
				createValidRequest())
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Validate error response
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d for %s, got %d", tc.expectedStatus, tc.name, resp.StatusCode)
			}

			// Validate error response structure
			var errorResponse map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			// Validate error response contains expected fields
			validateErrorResponse(t, errorResponse, "external_service_error")

			t.Logf("✅ %s test passed - Status: %d, Error: %v", tc.name, resp.StatusCode, errorResponse["error"])
		})
	}
}

// testDataCorruptionScenarios tests data corruption scenarios
func testDataCorruptionScenarios(t *testing.T) {
	testCases := []struct {
		name           string
		corruptionType string
		expectedStatus int
		description    string
	}{
		{
			name:           "Invalid JSON Structure",
			corruptionType: "invalid_json",
			expectedStatus: http.StatusBadRequest,
			description:    "Invalid JSON structure should return 400",
		},
		{
			name:           "Corrupted Data Fields",
			corruptionType: "corrupted_fields",
			expectedStatus: http.StatusBadRequest,
			description:    "Corrupted data fields should return 400",
		},
		{
			name:           "Invalid Encoding",
			corruptionType: "invalid_encoding",
			expectedStatus: http.StatusBadRequest,
			description:    "Invalid encoding should return 400",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handleDataCorruptionScenario(w, r, tc.corruptionType)
			}))
			defer server.Close()

			// Create corrupted request
			corruptedRequest := createCorruptedRequest(tc.corruptionType)

			// Test data corruption scenario
			resp, err := http.Post(server.URL+"/v1/classify", "application/json",
				strings.NewReader(corruptedRequest))
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Validate error response
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d for %s, got %d", tc.expectedStatus, tc.name, resp.StatusCode)
			}

			// Validate error response structure
			var errorResponse map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			// Validate error response contains expected fields
			validateErrorResponse(t, errorResponse, "data_corruption")

			t.Logf("✅ %s test passed - Status: %d, Error: %v", tc.name, resp.StatusCode, errorResponse["error"])
		})
	}
}

// Helper functions for enhanced error handling tests

func createOversizedRequest() string {
	// Create a request that exceeds the maximum allowed size
	largeDescription := strings.Repeat("This is a very long description. ", 10000)
	return fmt.Sprintf(`{"id": "test-oversized", "business_name": "Oversized Test Company", "description": "%s"}`, largeDescription)
}

func createLargeRequestWithSize(size int) string {
	// Create a request of specified size
	description := strings.Repeat("A", size-100) // Leave room for JSON structure
	return fmt.Sprintf(`{"id": "test-large", "business_name": "Large Test Company", "description": "%s"}`, description)
}

func createCorruptedRequest(corruptionType string) string {
	switch corruptionType {
	case "invalid_json":
		return `{"id": "test-corrupted", "business_name": "Corrupted Test Company", "description": "Test description"` // Missing closing brace
	case "corrupted_fields":
		return `{"id": "test-corrupted", "business_name": null, "description": "Test description"}`
	case "invalid_encoding":
		return `{"id": "test-corrupted", "business_name": "Test Company\x00", "description": "Test description"}`
	default:
		return `{"id": "test-corrupted", "business_name": "Test Company", "description": "Test description"}`
	}
}

func validateErrorResponse(t *testing.T, errorResponse map[string]interface{}, expectedErrorType string) {
	// Validate required error response fields
	requiredFields := []string{"error", "code", "message", "error_id", "timestamp"}
	for _, field := range requiredFields {
		if errorResponse[field] == nil {
			t.Errorf("Expected %s field in error response", field)
		}
	}

	// Validate error type
	if errorResponse["code"] != nil {
		errorCode := errorResponse["code"].(string)
		if !strings.Contains(strings.ToLower(errorCode), strings.ToLower(expectedErrorType)) {
			t.Errorf("Expected error code to contain %s, got %s", expectedErrorType, errorCode)
		}
	}
}

func validateTimeoutResponse(t *testing.T, errorResponse map[string]interface{}) {
	// Validate timeout-specific fields
	requiredFields := []string{"error", "code", "message", "timeout", "error_id", "timestamp"}
	for _, field := range requiredFields {
		if errorResponse[field] == nil {
			t.Errorf("Expected %s field in timeout response", field)
		}
	}

	// Validate timeout field is a number
	if timeout, ok := errorResponse["timeout"].(float64); !ok {
		t.Error("Expected timeout field to be a number")
	} else if timeout <= 0 {
		t.Error("Expected timeout to be positive")
	}
}

// Handler functions for enhanced error scenarios

func handleDatabaseErrorScenario(w http.ResponseWriter, r *http.Request, mockDB *mocks.MockDatabase, errorType string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simulate database error based on type
	var statusCode int
	var errorCode string
	var message string

	switch errorType {
	case "pool_exhaustion":
		statusCode = http.StatusServiceUnavailable
		errorCode = "DATABASE_POOL_EXHAUSTED"
		message = "Database connection pool exhausted"
	case "connection_timeout":
		statusCode = http.StatusGatewayTimeout
		errorCode = "DATABASE_CONNECTION_TIMEOUT"
		message = "Database connection timeout"
	case "database_unavailable":
		statusCode = http.StatusServiceUnavailable
		errorCode = "DATABASE_UNAVAILABLE"
		message = "Database service unavailable"
	case "deadlock":
		statusCode = http.StatusConflict
		errorCode = "DATABASE_DEADLOCK"
		message = "Database transaction deadlock"
	case "query_timeout":
		statusCode = http.StatusGatewayTimeout
		errorCode = "DATABASE_QUERY_TIMEOUT"
		message = "Database query timeout"
	default:
		statusCode = http.StatusInternalServerError
		errorCode = "DATABASE_ERROR"
		message = "Database error occurred"
	}

	errorResponse := map[string]interface{}{
		"error":     message,
		"code":      errorCode,
		"message":   message,
		"error_id":  fmt.Sprintf("ERR_%d", time.Now().UnixNano()),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"details": map[string]interface{}{
			"error_type": errorType,
			"service":    "database",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

func handleTimeoutScenario(w http.ResponseWriter, r *http.Request, mockService *mocks.MockClassificationService, timeout time.Duration) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	// Check for timeout
	select {
	case <-ctx.Done():
		errorResponse := map[string]interface{}{
			"error":     "Request timeout",
			"code":      "REQUEST_TIMEOUT",
			"message":   "Request processing timed out",
			"timeout":   timeout.Milliseconds(),
			"error_id":  fmt.Sprintf("ERR_%d", time.Now().UnixNano()),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"details": map[string]interface{}{
				"timeout_duration": timeout.String(),
				"service":          "classification",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusRequestTimeout)
		json.NewEncoder(w).Encode(errorResponse)
		return
	default:
		// Continue processing
	}

	// Simulate processing
	_, err := mockService.ClassifyBusiness(ctx, &shared.BusinessClassificationRequest{
		ID:           "timeout-test",
		BusinessName: "Timeout Test Company",
	})

	if err != nil {
		errorResponse := map[string]interface{}{
			"error":     err.Error(),
			"code":      "SERVICE_ERROR",
			"message":   "Classification service error",
			"error_id":  fmt.Sprintf("ERR_%d", time.Now().UnixNano()),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Success response
	response := map[string]interface{}{
		"id":     "timeout-test",
		"status": "success",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleInvalidDataScenario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check request size
	if r.ContentLength > 1024*1024 { // 1MB limit
		errorResponse := map[string]interface{}{
			"error":     "Request too large",
			"code":      "REQUEST_TOO_LARGE",
			"message":   "Request size exceeds maximum allowed",
			"error_id":  fmt.Sprintf("ERR_%d", time.Now().UnixNano()),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"details": map[string]interface{}{
				"max_size":    "1MB",
				"actual_size": r.ContentLength,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	var request struct {
		ID           string `json:"id"`
		BusinessName string `json:"business_name"`
		WebsiteURL   string `json:"website_url,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errorResponse := map[string]interface{}{
			"error":     "Invalid JSON format",
			"code":      "VALIDATION_ERROR",
			"message":   "Request body must be valid JSON",
			"error_id":  fmt.Sprintf("ERR_%d", time.Now().UnixNano()),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"details": map[string]interface{}{
				"error_type": "json_parsing",
				"service":    "validation",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Validate required fields
	if request.ID == "" {
		errorResponse := map[string]interface{}{
			"error":     "ID is required",
			"code":      "VALIDATION_ERROR",
			"message":   "Request ID is required",
			"error_id":  fmt.Sprintf("ERR_%d", time.Now().UnixNano()),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"details": map[string]interface{}{
				"error_type": "missing_field",
				"field":      "id",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	if request.BusinessName == "" || strings.TrimSpace(request.BusinessName) == "" {
		errorResponse := map[string]interface{}{
			"error":     "Business name is required",
			"code":      "VALIDATION_ERROR",
			"message":   "Business name is required and cannot be empty",
			"error_id":  fmt.Sprintf("ERR_%d", time.Now().UnixNano()),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"details": map[string]interface{}{
				"error_type": "missing_field",
				"field":      "business_name",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Validate business name format
	if strings.ContainsAny(request.BusinessName, "!@#$%^&*()") {
		errorResponse := map[string]interface{}{
			"error":     "Invalid business name format",
			"code":      "VALIDATION_ERROR",
			"message":   "Business name contains invalid characters",
			"error_id":  fmt.Sprintf("ERR_%d", time.Now().UnixNano()),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"details": map[string]interface{}{
				"error_type": "invalid_format",
				"field":      "business_name",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Validate URL format if provided
	if request.WebsiteURL != "" {
		if !strings.HasPrefix(request.WebsiteURL, "http://") && !strings.HasPrefix(request.WebsiteURL, "https://") {
			errorResponse := map[string]interface{}{
				"error":     "Invalid URL format",
				"code":      "VALIDATION_ERROR",
				"message":   "Website URL must start with http:// or https://",
				"error_id":  fmt.Sprintf("ERR_%d", time.Now().UnixNano()),
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"details": map[string]interface{}{
					"error_type": "invalid_format",
					"field":      "website_url",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
			return
		}
	}

	// Success response
	response := map[string]interface{}{
		"id":            request.ID,
		"business_name": request.BusinessName,
		"status":        "success",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleServiceUnavailabilityScenario(w http.ResponseWriter, r *http.Request, mockService *mocks.MockClassificationService, serviceType string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simulate service unavailability
	errorResponse := map[string]interface{}{
		"error":     fmt.Sprintf("%s service unavailable", serviceType),
		"code":      "SERVICE_UNAVAILABLE",
		"message":   fmt.Sprintf("The %s service is temporarily unavailable", serviceType),
		"error_id":  fmt.Sprintf("ERR_%d", time.Now().UnixNano()),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"details": map[string]interface{}{
			"service_type": serviceType,
			"status":       "unavailable",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	json.NewEncoder(w).Encode(errorResponse)
}

func handleMemoryExhaustionScenario(w http.ResponseWriter, r *http.Request, maxSize int) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check request size
	if r.ContentLength > int64(maxSize) {
		errorResponse := map[string]interface{}{
			"error":     "Request too large",
			"code":      "REQUEST_TOO_LARGE",
			"message":   "Request size exceeds maximum allowed",
			"error_id":  fmt.Sprintf("ERR_%d", time.Now().UnixNano()),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"details": map[string]interface{}{
				"max_size":    maxSize,
				"actual_size": r.ContentLength,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Success response
	response := map[string]interface{}{
		"id":     "memory-test",
		"status": "success",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleConcurrentAccessScenario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simulate processing delay
	time.Sleep(10 * time.Millisecond)

	// Success response
	response := map[string]interface{}{
		"id":     "concurrent-test",
		"status": "success",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleExternalServiceFailureScenario(w http.ResponseWriter, r *http.Request, serviceName string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simulate external service failure
	errorResponse := map[string]interface{}{
		"error":     fmt.Sprintf("External %s service failure", serviceName),
		"code":      "EXTERNAL_SERVICE_ERROR",
		"message":   fmt.Sprintf("Unable to connect to %s service", serviceName),
		"error_id":  fmt.Sprintf("ERR_%d", time.Now().UnixNano()),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"details": map[string]interface{}{
			"service_name": serviceName,
			"service_type": "external",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	json.NewEncoder(w).Encode(errorResponse)
}

func handleDataCorruptionScenario(w http.ResponseWriter, r *http.Request, corruptionType string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simulate data corruption based on type
	var statusCode int
	var errorCode string
	var message string

	switch corruptionType {
	case "invalid_json":
		statusCode = http.StatusBadRequest
		errorCode = "INVALID_JSON"
		message = "Invalid JSON structure"
	case "corrupted_fields":
		statusCode = http.StatusBadRequest
		errorCode = "CORRUPTED_DATA"
		message = "Data fields are corrupted"
	case "invalid_encoding":
		statusCode = http.StatusBadRequest
		errorCode = "INVALID_ENCODING"
		message = "Invalid data encoding"
	default:
		statusCode = http.StatusBadRequest
		errorCode = "DATA_CORRUPTION"
		message = "Data corruption detected"
	}

	errorResponse := map[string]interface{}{
		"error":     message,
		"code":      errorCode,
		"message":   message,
		"error_id":  fmt.Sprintf("ERR_%d", time.Now().UnixNano()),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"details": map[string]interface{}{
			"corruption_type": corruptionType,
			"service":         "validation",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}
