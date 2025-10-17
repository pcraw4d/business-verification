package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/shared"
	"kyb-platform/test/mocks"
)

// TestEndToEndClassificationWorkflow tests the complete classification workflow from API request to response
func TestEndToEndClassificationWorkflow(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	// Setup test environment
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create mock services
	mockClassificationService := mocks.NewMockClassificationService()
	mockDB := mocks.NewMockDatabase()

	// Test database connection
	if err := mockDB.Connect(); err != nil {
		t.Fatalf("Failed to connect to mock database: %v", err)
	}
	defer mockDB.Disconnect()

	// Test 1: Complete workflow with valid business data
	t.Run("CompleteWorkflowValidData", func(t *testing.T) {
		// Create comprehensive test request
		request := &shared.BusinessClassificationRequest{
			ID:                 "e2e-test-001",
			BusinessName:       "Acme Technology Solutions",
			BusinessType:       "Corporation",
			Industry:           "Technology",
			Description:        "Leading provider of cloud-based software solutions and digital transformation services",
			Keywords:           []string{"software", "cloud", "digital", "transformation", "technology"},
			WebsiteURL:         "https://www.acmetech.com",
			Address:            "123 Tech Street, Silicon Valley, CA 94000",
			RegistrationNumber: "REG789012",
			TaxID:              "TAX789012",
			RequestedAt:        time.Now(),
		}

		// Step 1: Validate request structure
		if request.ID == "" {
			t.Fatal("Request ID is required")
		}
		if request.BusinessName == "" {
			t.Fatal("Business name is required")
		}

		// Step 2: Process classification through service
		startTime := time.Now()
		result, err := mockClassificationService.ClassifyBusiness(ctx, request)
		processingTime := time.Since(startTime)

		if err != nil {
			t.Fatalf("Classification failed: %v", err)
		}

		// Step 3: Validate response structure
		if result == nil {
			t.Fatal("Expected classification result")
		}

		// Validate response ID matches request ID
		if result.ID != request.ID {
			t.Errorf("Expected result ID %s, got %s", request.ID, result.ID)
		}

		// Validate business name
		if result.BusinessName != request.BusinessName {
			t.Errorf("Expected business name %s, got %s", request.BusinessName, result.BusinessName)
		}

		// Validate classifications exist
		if len(result.Classifications) == 0 {
			t.Fatal("Expected at least one classification result")
		}

		// Validate confidence score range
		if result.OverallConfidence < 0.0 || result.OverallConfidence > 1.0 {
			t.Errorf("Expected confidence between 0.0 and 1.0, got %f", result.OverallConfidence)
		}

		// Validate primary classification
		if result.PrimaryClassification == nil {
			t.Fatal("Expected primary classification")
		}

		if result.PrimaryClassification.IndustryCode == "" {
			t.Fatal("Expected industry code in primary classification")
		}

		if result.PrimaryClassification.IndustryName == "" {
			t.Fatal("Expected industry name in primary classification")
		}

		// Validate processing time
		if result.ProcessingTime <= 0 {
			t.Errorf("Expected positive processing time, got %v", result.ProcessingTime)
		}

		// Validate processing time is reasonable
		if processingTime > 5*time.Second {
			t.Errorf("Expected processing time < 5s, got %v", processingTime)
		}

		// Validate module results
		if len(result.ModuleResults) == 0 {
			t.Fatal("Expected module results")
		}

		// Validate at least one module succeeded
		successfulModules := 0
		for _, moduleResult := range result.ModuleResults {
			if moduleResult.Success {
				successfulModules++
			}
		}
		if successfulModules == 0 {
			t.Fatal("Expected at least one successful module result")
		}

		t.Logf("✅ Complete workflow test passed - ID: %s, Classifications: %d, Confidence: %.2f, Processing Time: %v",
			result.ID, len(result.Classifications), result.OverallConfidence, processingTime)
	})

	// Test 2: Workflow with minimal data
	t.Run("WorkflowMinimalData", func(t *testing.T) {
		// Create minimal test request
		request := &shared.BusinessClassificationRequest{
			ID:           "e2e-test-002",
			BusinessName: "Simple Corp",
			RequestedAt:  time.Now(),
		}

		// Process classification
		result, err := mockClassificationService.ClassifyBusiness(ctx, request)
		if err != nil {
			t.Fatalf("Classification failed with minimal data: %v", err)
		}

		// Validate basic response structure
		if result == nil {
			t.Fatal("Expected classification result")
		}

		if result.ID != request.ID {
			t.Errorf("Expected result ID %s, got %s", request.ID, result.ID)
		}

		if result.BusinessName != request.BusinessName {
			t.Errorf("Expected business name %s, got %s", request.BusinessName, result.BusinessName)
		}

		// Should still get classifications even with minimal data
		if len(result.Classifications) == 0 {
			t.Fatal("Expected at least one classification result even with minimal data")
		}

		t.Logf("✅ Minimal data workflow test passed - ID: %s, Classifications: %d",
			result.ID, len(result.Classifications))
	})

	// Test 3: Workflow with complex business data
	t.Run("WorkflowComplexData", func(t *testing.T) {
		// Create complex test request
		request := &shared.BusinessClassificationRequest{
			ID:                 "e2e-test-003",
			BusinessName:       "Global Financial Services & Investment Management Group",
			BusinessType:       "Limited Liability Company",
			Industry:           "Financial Services",
			Description:        "Comprehensive financial services including investment management, wealth planning, insurance, and retirement solutions for high-net-worth individuals and institutional clients",
			Keywords:           []string{"financial", "investment", "wealth", "management", "insurance", "retirement", "planning"},
			WebsiteURL:         "https://www.globalfinancial.com",
			Address:            "456 Wall Street, New York, NY 10005",
			RegistrationNumber: "REG345678",
			TaxID:              "TAX345678",
			RequestedAt:        time.Now(),
		}

		// Process classification
		result, err := mockClassificationService.ClassifyBusiness(ctx, request)
		if err != nil {
			t.Fatalf("Classification failed with complex data: %v", err)
		}

		// Validate response structure
		if result == nil {
			t.Fatal("Expected classification result")
		}

		// Validate multiple classifications for complex business
		if len(result.Classifications) < 2 {
			t.Logf("Warning: Expected multiple classifications for complex business, got %d", len(result.Classifications))
		}

		// Validate primary classification makes sense for financial services
		if result.PrimaryClassification != nil {
			if result.PrimaryClassification.IndustryCode == "" {
				t.Fatal("Expected industry code in primary classification")
			}
		}

		// Validate confidence score
		if result.OverallConfidence < 0.0 || result.OverallConfidence > 1.0 {
			t.Errorf("Expected confidence between 0.0 and 1.0, got %f", result.OverallConfidence)
		}

		t.Logf("✅ Complex data workflow test passed - ID: %s, Classifications: %d, Confidence: %.2f",
			result.ID, len(result.Classifications), result.OverallConfidence)
	})

	// Test 4: Workflow error handling
	t.Run("WorkflowErrorHandling", func(t *testing.T) {
		// Configure mock to fail
		mockClassificationService.SetFailureMode(true, "mock classification error")

		request := &shared.BusinessClassificationRequest{
			ID:           "e2e-test-004",
			BusinessName: "Error Test Company",
			RequestedAt:  time.Now(),
		}

		// Process classification (should fail)
		result, err := mockClassificationService.ClassifyBusiness(ctx, request)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if result != nil {
			t.Fatal("Expected nil result on error")
		}

		// Verify error message
		expectedError := "mock classification error"
		if err.Error() != expectedError {
			t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
		}

		// Reset mock to success mode
		mockClassificationService.SetFailureMode(false, "")

		t.Logf("✅ Error handling workflow test passed - Error: %v", err)
	})

	// Test 5: Workflow performance under load
	t.Run("WorkflowPerformanceLoad", func(t *testing.T) {
		// Test multiple concurrent requests
		concurrentRequests := 5
		results := make(chan error, concurrentRequests)

		for i := 0; i < concurrentRequests; i++ {
			go func(index int) {
				request := &shared.BusinessClassificationRequest{
					ID:           fmt.Sprintf("e2e-load-test-%d", index),
					BusinessName: fmt.Sprintf("Load Test Company %d", index),
					RequestedAt:  time.Now(),
				}

				_, err := mockClassificationService.ClassifyBusiness(ctx, request)
				results <- err
			}(i)
		}

		// Collect results
		successCount := 0
		errorCount := 0
		for i := 0; i < concurrentRequests; i++ {
			select {
			case err := <-results:
				if err != nil {
					errorCount++
					t.Logf("Load test request failed: %v", err)
				} else {
					successCount++
				}
			case <-time.After(10 * time.Second):
				t.Fatal("Load test timed out")
			}
		}

		// Validate success rate
		successRate := float64(successCount) / float64(concurrentRequests)
		if successRate < 0.8 { // Allow 20% failure rate
			t.Errorf("Expected success rate >= 80%%, got %.1f%%", successRate*100)
		}

		t.Logf("✅ Load test passed - Success: %d, Errors: %d, Success Rate: %.1f%%",
			successCount, errorCount, successRate*100)
	})
}

// TestAPIIntegrationWorkflow tests the complete API integration workflow
func TestAPIIntegrationWorkflow(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	// Create mock services
	mockClassificationService := mocks.NewMockClassificationService()

	// Create test server with realistic API endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/classify":
			handleClassificationAPI(w, r, mockClassificationService)
		case "/v1/classify/batch":
			handleBatchClassificationAPI(w, r, mockClassificationService)
		case "/health":
			handleHealthCheckAPI(w, r)
		case "/v1/status":
			handleStatusAPI(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Test 1: Single classification API workflow
	t.Run("SingleClassificationAPI", func(t *testing.T) {
		request := &shared.BusinessClassificationRequest{
			ID:           "api-e2e-001",
			BusinessName: "API Test Company",
			Description:  "Test company for API integration workflow",
			RequestedAt:  time.Now(),
		}

		// Marshal request
		requestBody, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		// Make API request
		resp, err := http.Post(server.URL+"/v1/classify", "application/json",
			bytes.NewReader(requestBody))
		if err != nil {
			t.Fatalf("Failed to make classification request: %v", err)
		}
		defer resp.Body.Close()

		// Validate response status
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Validate content type
		contentType := resp.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected content type application/json, got %s", contentType)
		}

		// Parse response
		var response shared.BusinessClassificationResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Validate response structure
		if response.ID != request.ID {
			t.Errorf("Expected response ID %s, got %s", request.ID, response.ID)
		}

		if response.BusinessName != request.BusinessName {
			t.Errorf("Expected business name %s, got %s", request.BusinessName, response.BusinessName)
		}

		if len(response.Classifications) == 0 {
			t.Fatal("Expected at least one classification result")
		}

		t.Logf("✅ Single classification API test passed - Response ID: %s, Classifications: %d",
			response.ID, len(response.Classifications))
	})

	// Test 2: Batch classification API workflow
	t.Run("BatchClassificationAPI", func(t *testing.T) {
		// Create batch request
		batchRequest := struct {
			Requests []shared.BusinessClassificationRequest `json:"requests"`
		}{
			Requests: []shared.BusinessClassificationRequest{
				{
					ID:           "batch-001",
					BusinessName: "Batch Company 1",
					RequestedAt:  time.Now(),
				},
				{
					ID:           "batch-002",
					BusinessName: "Batch Company 2",
					RequestedAt:  time.Now(),
				},
			},
		}

		// Marshal request
		requestBody, err := json.Marshal(batchRequest)
		if err != nil {
			t.Fatalf("Failed to marshal batch request: %v", err)
		}

		// Make API request
		resp, err := http.Post(server.URL+"/v1/classify/batch", "application/json",
			bytes.NewReader(requestBody))
		if err != nil {
			t.Fatalf("Failed to make batch classification request: %v", err)
		}
		defer resp.Body.Close()

		// Validate response status
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var batchResponse struct {
			Results []shared.BusinessClassificationResponse `json:"results"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&batchResponse); err != nil {
			t.Fatalf("Failed to decode batch response: %v", err)
		}

		// Validate batch response
		if len(batchResponse.Results) != 2 {
			t.Errorf("Expected 2 batch results, got %d", len(batchResponse.Results))
		}

		for i, result := range batchResponse.Results {
			if result.ID != fmt.Sprintf("batch-%03d", i+1) {
				t.Errorf("Expected batch result ID batch-%03d, got %s", i+1, result.ID)
			}
		}

		t.Logf("✅ Batch classification API test passed - Results: %d", len(batchResponse.Results))
	})

	// Test 3: Health check API workflow
	t.Run("HealthCheckAPI", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("Failed to make health check request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse health response
		var healthResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&healthResponse); err != nil {
			t.Fatalf("Failed to decode health response: %v", err)
		}

		// Validate health response structure
		if status, exists := healthResponse["status"]; !exists || status != "healthy" {
			t.Errorf("Expected health status 'healthy', got %v", status)
		}

		t.Logf("✅ Health check API test passed - Status: %v", healthResponse["status"])
	})

	// Test 4: Status API workflow
	t.Run("StatusAPI", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/v1/status")
		if err != nil {
			t.Fatalf("Failed to make status request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse status response
		var statusResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&statusResponse); err != nil {
			t.Fatalf("Failed to decode status response: %v", err)
		}

		// Validate status response has required fields
		requiredFields := []string{"status", "timestamp", "version"}
		for _, field := range requiredFields {
			if _, exists := statusResponse[field]; !exists {
				t.Errorf("Expected status response to have field '%s'", field)
			}
		}

		t.Logf("✅ Status API test passed - Status: %v", statusResponse["status"])
	})

	// Test 5: API error handling workflow
	t.Run("APIErrorHandling", func(t *testing.T) {
		// Test invalid JSON
		resp, err := http.Post(server.URL+"/v1/classify", "application/json",
			bytes.NewReader([]byte("invalid json")))
		if err != nil {
			t.Fatalf("Failed to make invalid request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid JSON, got %d", resp.StatusCode)
		}

		// Test missing required fields
		invalidRequest := map[string]interface{}{
			"id": "test-invalid",
			// Missing business_name
		}
		requestBody, _ := json.Marshal(invalidRequest)

		resp, err = http.Post(server.URL+"/v1/classify", "application/json",
			bytes.NewReader(requestBody))
		if err != nil {
			t.Fatalf("Failed to make invalid request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for missing required fields, got %d", resp.StatusCode)
		}

		t.Logf("✅ API error handling test passed")
	})
}

// handleClassificationAPI handles single classification API requests
func handleClassificationAPI(w http.ResponseWriter, r *http.Request, service *mocks.MockClassificationService) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request shared.BusinessClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if request.ID == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}
	if request.BusinessName == "" {
		http.Error(w, "Business name is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	result, err := service.ClassifyBusiness(ctx, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleBatchClassificationAPI handles batch classification API requests
func handleBatchClassificationAPI(w http.ResponseWriter, r *http.Request, service *mocks.MockClassificationService) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var batchRequest struct {
		Requests []shared.BusinessClassificationRequest `json:"requests"`
	}
	if err := json.NewDecoder(r.Body).Decode(&batchRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(batchRequest.Requests) == 0 {
		http.Error(w, "At least one request is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	var results []shared.BusinessClassificationResponse

	for _, request := range batchRequest.Requests {
		result, err := service.ClassifyBusiness(ctx, &request)
		if err != nil {
			http.Error(w, fmt.Sprintf("Classification failed for request %s: %v", request.ID, err), http.StatusInternalServerError)
			return
		}
		results = append(results, *result)
	}

	response := struct {
		Results []shared.BusinessClassificationResponse `json:"results"`
	}{
		Results: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealthCheckAPI handles health check API requests
func handleHealthCheckAPI(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"test_mode": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStatusAPI handles status API requests
func handleStatusAPI(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "operational",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"uptime":    "24h",
		"services": map[string]interface{}{
			"classification": "healthy",
			"database":       "healthy",
			"api":            "healthy",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
