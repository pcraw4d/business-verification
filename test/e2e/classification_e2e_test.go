package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/shared"
	"github.com/pcraw4d/business-verification/test/mocks"
)

// TestClassificationE2E tests the complete classification flow end-to-end
func TestClassificationE2E(t *testing.T) {
	// Skip if not running E2E tests
	if os.Getenv("E2E_TESTS") != "true" {
		t.Skip("Skipping E2E tests - set E2E_TESTS=true to run")
	}

	// Setup test environment
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create mock services
	mockClassificationService := mocks.NewMockClassificationService()
	mockDB := mocks.NewMockDatabase()

	// Test database connection
	if err := mockDB.Connect(); err != nil {
		t.Fatalf("Failed to connect to mock database: %v", err)
	}
	defer mockDB.Disconnect()

	// Test 1: Basic classification request
	t.Run("BasicClassificationRequest", func(t *testing.T) {
		request := &shared.BusinessClassificationRequest{
			ID:                 "test-001",
			BusinessName:       "Tech Solutions Inc",
			BusinessType:       "Corporation",
			Industry:           "Technology",
			Description:        "Software development and consulting services",
			Keywords:           []string{"software", "development", "consulting", "technology"},
			RegistrationNumber: "REG123456",
			TaxID:              "TAX123456",
			RequestedAt:        time.Now(),
		}

		// Perform classification
		result, err := mockClassificationService.ClassifyBusiness(ctx, request)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Verify result structure
		if result == nil {
			t.Fatal("Expected classification result")
		}

		if result.ID != request.ID {
			t.Errorf("Expected result ID %s, got %s", request.ID, result.ID)
		}

		if result.BusinessName != request.BusinessName {
			t.Errorf("Expected business name %s, got %s", request.BusinessName, result.BusinessName)
		}

		if len(result.Classifications) == 0 {
			t.Fatal("Expected at least one classification result")
		}

		if result.OverallConfidence < 0.0 || result.OverallConfidence > 1.0 {
			t.Errorf("Expected confidence between 0.0 and 1.0, got %f", result.OverallConfidence)
		}

		// Verify primary classification
		if result.PrimaryClassification == nil {
			t.Fatal("Expected primary classification")
		}

		if result.PrimaryClassification.IndustryCode == "" {
			t.Fatal("Expected industry code in primary classification")
		}

		if result.PrimaryClassification.IndustryName == "" {
			t.Fatal("Expected industry name in primary classification")
		}

		t.Logf("✅ Basic classification test passed - Industry: %s (%s), Confidence: %.2f",
			result.PrimaryClassification.IndustryName,
			result.PrimaryClassification.IndustryCode,
			result.OverallConfidence)
	})

	// Test 2: Classification with error handling
	t.Run("ClassificationErrorHandling", func(t *testing.T) {
		// Configure mock to fail
		mockClassificationService.SetFailureMode(true, "mock classification error")

		request := &shared.BusinessClassificationRequest{
			ID:           "test-002",
			BusinessName: "Test Company",
			RequestedAt:  time.Now(),
		}

		// Perform classification (should fail)
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

		t.Logf("✅ Error handling test passed - Error: %v", err)
	})

	// Test 3: Performance and timeout handling
	t.Run("PerformanceAndTimeout", func(t *testing.T) {
		// Configure mock with delay
		mockClassificationService.SetDelay(200 * time.Millisecond)

		request := &shared.BusinessClassificationRequest{
			ID:           "test-003",
			BusinessName: "Performance Test Company",
			RequestedAt:  time.Now(),
		}

		start := time.Now()
		result, err := mockClassificationService.ClassifyBusiness(ctx, request)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result == nil {
			t.Fatal("Expected classification result")
		}

		// Verify processing time is reasonable
		if duration > 1*time.Second {
			t.Errorf("Expected processing time < 1s, got %v", duration)
		}

		// Verify processing time in result
		if result.ProcessingTime < 100*time.Millisecond {
			t.Errorf("Expected processing time >= 100ms, got %v", result.ProcessingTime)
		}

		t.Logf("✅ Performance test passed - Duration: %v, Processing Time: %v", duration, result.ProcessingTime)
	})

	// Test 4: Multiple classifications
	t.Run("MultipleClassifications", func(t *testing.T) {
		// Set custom mock results with multiple classifications
		customResults := []shared.IndustryClassification{
			{
				IndustryCode:         "541511",
				IndustryName:         "Custom Computer Programming Services",
				ConfidenceScore:      0.92,
				ClassificationMethod: "mock_keyword_matching",
				Keywords:             []string{"software", "development"},
				Description:          "Primary classification",
				Evidence:             "Mock evidence 1",
				ProcessingTime:       50 * time.Millisecond,
				Metadata:             map[string]interface{}{"source": "mock"},
			},
			{
				IndustryCode:         "541512",
				IndustryName:         "Computer Systems Design Services",
				ConfidenceScore:      0.78,
				ClassificationMethod: "mock_keyword_matching",
				Keywords:             []string{"systems", "design"},
				Description:          "Secondary classification",
				Evidence:             "Mock evidence 2",
				ProcessingTime:       30 * time.Millisecond,
				Metadata:             map[string]interface{}{"source": "mock"},
			},
		}

		mockClassificationService.SetMockResults(customResults)

		request := &shared.BusinessClassificationRequest{
			ID:           "test-004",
			BusinessName: "Multi-Industry Company",
			RequestedAt:  time.Now(),
		}

		result, err := mockClassificationService.ClassifyBusiness(ctx, request)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result == nil {
			t.Fatal("Expected classification result")
		}

		// Verify multiple classifications
		if len(result.Classifications) != 2 {
			t.Errorf("Expected 2 classifications, got %d", len(result.Classifications))
		}

		// Verify primary classification is the first one
		if result.PrimaryClassification == nil {
			t.Fatal("Expected primary classification")
		}

		if result.PrimaryClassification.IndustryCode != "541511" {
			t.Errorf("Expected primary classification code 541511, got %s", result.PrimaryClassification.IndustryCode)
		}

		// Verify module results
		if len(result.ModuleResults) == 0 {
			t.Fatal("Expected module results")
		}

		mockResult, exists := result.ModuleResults["mock_module"]
		if !exists {
			t.Fatal("Expected mock_module result")
		}

		if !mockResult.Success {
			t.Fatal("Expected successful module result")
		}

		if len(mockResult.Classifications) != 2 {
			t.Errorf("Expected 2 module classifications, got %d", len(mockResult.Classifications))
		}

		t.Logf("✅ Multiple classifications test passed - Count: %d, Primary: %s",
			len(result.Classifications), result.PrimaryClassification.IndustryName)
	})
}

// TestAPIIntegrationE2E tests API integration end-to-end
func TestAPIIntegrationE2E(t *testing.T) {
	// Skip if not running E2E tests
	if os.Getenv("E2E_TESTS") != "true" {
		t.Skip("Skipping E2E tests - set E2E_TESTS=true to run")
	}

	// Create mock services
	mockClassificationService := mocks.NewMockClassificationService()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/classify":
			handleClassificationRequest(w, r, mockClassificationService)
		case "/health":
			handleHealthCheck(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Test 1: Health check endpoint
	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("Failed to make health check request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		t.Logf("✅ Health check test passed - Status: %d", resp.StatusCode)
	})

	// Test 2: Classification API endpoint
	t.Run("ClassificationAPI", func(t *testing.T) {
		request := &shared.BusinessClassificationRequest{
			ID:           "api-test-001",
			BusinessName: "API Test Company",
			Description:  "Test company for API integration",
			RequestedAt:  time.Now(),
		}

		// Marshal request
		requestBody, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		// Make API request
		resp, err := http.Post(server.URL+"/api/v1/classify", "application/json", 
			bytes.NewReader(requestBody))
		if err != nil {
			t.Fatalf("Failed to make classification request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var response shared.BusinessClassificationResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Verify response
		if response.ID != request.ID {
			t.Errorf("Expected response ID %s, got %s", request.ID, response.ID)
		}

		if response.BusinessName != request.BusinessName {
			t.Errorf("Expected business name %s, got %s", request.BusinessName, response.BusinessName)
		}

		if len(response.Classifications) == 0 {
			t.Fatal("Expected at least one classification result")
		}

		t.Logf("✅ API integration test passed - Response ID: %s, Classifications: %d",
			response.ID, len(response.Classifications))
	})
}

// handleClassificationRequest handles classification API requests
func handleClassificationRequest(w http.ResponseWriter, r *http.Request, service *mocks.MockClassificationService) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request shared.BusinessClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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

// handleHealthCheck handles health check requests
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"test_mode": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
