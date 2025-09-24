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

	"github.com/pcraw4d/business-verification/internal/shared"
	"github.com/pcraw4d/business-verification/test/mocks"
)

// TestAPIEndpointIntegration tests all classification API endpoints
func TestAPIEndpointIntegration(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	// Setup test environment

	// Create mock services
	mockClassificationService := mocks.NewMockClassificationService()

	// Create test server with all API endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/classify":
			handleClassificationEndpoint(w, r, mockClassificationService)
		case "/v1/classify/batch":
			handleBatchClassificationEndpoint(w, r, mockClassificationService)
		case "/v1/classify/status":
			handleClassificationStatusEndpoint(w, r, mockClassificationService)
		case "/v1/classify/history":
			handleClassificationHistoryEndpoint(w, r, mockClassificationService)
		case "/health":
			handleHealthEndpoint(w, r)
		case "/v1/status":
			handleStatusEndpoint(w, r)
		case "/v1/metrics":
			handleMetricsEndpoint(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Test 1: Single classification endpoint
	t.Run("SingleClassificationEndpoint", func(t *testing.T) {
		// Test valid request
		request := &shared.BusinessClassificationRequest{
			ID:           "api-test-001",
			BusinessName: "API Test Company",
			Description:  "Test company for API endpoint testing",
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

		// Validate response headers
		expectedHeaders := map[string]string{
			"Content-Type": "application/json",
			"X-Request-ID": request.ID,
		}

		for header, expectedValue := range expectedHeaders {
			actualValue := resp.Header.Get(header)
			if actualValue != expectedValue {
				t.Errorf("Expected header %s: %s, got %s", header, expectedValue, actualValue)
			}
		}

		t.Logf("✅ Single classification endpoint test passed - Response ID: %s, Classifications: %d",
			response.ID, len(response.Classifications))
	})

	// Test 2: Batch classification endpoint
	t.Run("BatchClassificationEndpoint", func(t *testing.T) {
		// Create batch request
		batchRequest := struct {
			Requests []shared.BusinessClassificationRequest `json:"requests"`
		}{
			Requests: []shared.BusinessClassificationRequest{
				{
					ID:           "batch-001",
					BusinessName: "Batch Company 1",
					Description:  "First batch company",
					RequestedAt:  time.Now(),
				},
				{
					ID:           "batch-002",
					BusinessName: "Batch Company 2",
					Description:  "Second batch company",
					RequestedAt:  time.Now(),
				},
				{
					ID:           "batch-003",
					BusinessName: "Batch Company 3",
					Description:  "Third batch company",
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
			Summary struct {
				TotalRequests      int `json:"total_requests"`
				SuccessfulRequests int `json:"successful_requests"`
				FailedRequests     int `json:"failed_requests"`
			} `json:"summary"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&batchResponse); err != nil {
			t.Fatalf("Failed to decode batch response: %v", err)
		}

		// Validate batch response
		if len(batchResponse.Results) != 3 {
			t.Errorf("Expected 3 batch results, got %d", len(batchResponse.Results))
		}

		if batchResponse.Summary.TotalRequests != 3 {
			t.Errorf("Expected total requests 3, got %d", batchResponse.Summary.TotalRequests)
		}

		if batchResponse.Summary.SuccessfulRequests != 3 {
			t.Errorf("Expected successful requests 3, got %d", batchResponse.Summary.SuccessfulRequests)
		}

		if batchResponse.Summary.FailedRequests != 0 {
			t.Errorf("Expected failed requests 0, got %d", batchResponse.Summary.FailedRequests)
		}

		// Validate individual results
		for i, result := range batchResponse.Results {
			expectedID := fmt.Sprintf("batch-%03d", i+1)
			if result.ID != expectedID {
				t.Errorf("Expected batch result ID %s, got %s", expectedID, result.ID)
			}
		}

		t.Logf("✅ Batch classification endpoint test passed - Results: %d, Success Rate: %.1f%%",
			len(batchResponse.Results), float64(batchResponse.Summary.SuccessfulRequests)/float64(batchResponse.Summary.TotalRequests)*100)
	})

	// Test 3: Classification status endpoint
	t.Run("ClassificationStatusEndpoint", func(t *testing.T) {
		// Test status request
		resp, err := http.Get(server.URL + "/v1/classify/status")
		if err != nil {
			t.Fatalf("Failed to make status request: %v", err)
		}
		defer resp.Body.Close()

		// Validate response status
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var statusResponse struct {
			Status    string `json:"status"`
			Timestamp string `json:"timestamp"`
			Version   string `json:"version"`
			Services  struct {
				Classification string `json:"classification"`
				Database       string `json:"database"`
				API            string `json:"api"`
			} `json:"services"`
			Metrics struct {
				TotalRequests      int     `json:"total_requests"`
				SuccessfulRequests int     `json:"successful_requests"`
				FailedRequests     int     `json:"failed_requests"`
				AverageLatency     float64 `json:"average_latency_ms"`
			} `json:"metrics"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&statusResponse); err != nil {
			t.Fatalf("Failed to decode status response: %v", err)
		}

		// Validate status response
		if statusResponse.Status != "operational" {
			t.Errorf("Expected status 'operational', got %s", statusResponse.Status)
		}

		if statusResponse.Services.Classification != "healthy" {
			t.Errorf("Expected classification service 'healthy', got %s", statusResponse.Services.Classification)
		}

		if statusResponse.Services.Database != "healthy" {
			t.Errorf("Expected database service 'healthy', got %s", statusResponse.Services.Database)
		}

		if statusResponse.Services.API != "healthy" {
			t.Errorf("Expected API service 'healthy', got %s", statusResponse.Services.API)
		}

		t.Logf("✅ Classification status endpoint test passed - Status: %s, Services: %+v",
			statusResponse.Status, statusResponse.Services)
	})

	// Test 4: Classification history endpoint
	t.Run("ClassificationHistoryEndpoint", func(t *testing.T) {
		// Test history request with query parameters
		resp, err := http.Get(server.URL + "/v1/classify/history?limit=10&offset=0")
		if err != nil {
			t.Fatalf("Failed to make history request: %v", err)
		}
		defer resp.Body.Close()

		// Validate response status
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var historyResponse struct {
			Classifications []shared.BusinessClassificationResponse `json:"classifications"`
			Pagination      struct {
				Limit  int `json:"limit"`
				Offset int `json:"offset"`
				Total  int `json:"total"`
			} `json:"pagination"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&historyResponse); err != nil {
			t.Fatalf("Failed to decode history response: %v", err)
		}

		// Validate history response
		if historyResponse.Pagination.Limit != 10 {
			t.Errorf("Expected limit 10, got %d", historyResponse.Pagination.Limit)
		}

		if historyResponse.Pagination.Offset != 0 {
			t.Errorf("Expected offset 0, got %d", historyResponse.Pagination.Offset)
		}

		// Validate classifications structure
		for _, classification := range historyResponse.Classifications {
			if classification.ID == "" {
				t.Error("Expected classification ID to be set")
			}
			if classification.BusinessName == "" {
				t.Error("Expected business name to be set")
			}
		}

		t.Logf("✅ Classification history endpoint test passed - Classifications: %d, Total: %d",
			len(historyResponse.Classifications), historyResponse.Pagination.Total)
	})

	// Test 5: Health endpoint
	t.Run("HealthEndpoint", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("Failed to make health check request: %v", err)
		}
		defer resp.Body.Close()

		// Validate response status
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var healthResponse struct {
			Status    string    `json:"status"`
			Timestamp time.Time `json:"timestamp"`
			Version   string    `json:"version"`
			Uptime    string    `json:"uptime"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&healthResponse); err != nil {
			t.Fatalf("Failed to decode health response: %v", err)
		}

		// Validate health response
		if healthResponse.Status != "healthy" {
			t.Errorf("Expected health status 'healthy', got %s", healthResponse.Status)
		}

		if healthResponse.Version == "" {
			t.Error("Expected version to be set")
		}

		t.Logf("✅ Health endpoint test passed - Status: %s, Version: %s",
			healthResponse.Status, healthResponse.Version)
	})

	// Test 6: Status endpoint
	t.Run("StatusEndpoint", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/v1/status")
		if err != nil {
			t.Fatalf("Failed to make status request: %v", err)
		}
		defer resp.Body.Close()

		// Validate response status
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var statusResponse struct {
			Status    string            `json:"status"`
			Timestamp time.Time         `json:"timestamp"`
			Version   string            `json:"version"`
			Uptime    string            `json:"uptime"`
			Services  map[string]string `json:"services"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&statusResponse); err != nil {
			t.Fatalf("Failed to decode status response: %v", err)
		}

		// Validate status response
		if statusResponse.Status != "operational" {
			t.Errorf("Expected status 'operational', got %s", statusResponse.Status)
		}

		// Validate services
		expectedServices := []string{"classification", "database", "api"}
		for _, service := range expectedServices {
			if status, exists := statusResponse.Services[service]; !exists {
				t.Errorf("Expected service %s to be present", service)
			} else if status != "healthy" {
				t.Errorf("Expected service %s to be 'healthy', got %s", service, status)
			}
		}

		t.Logf("✅ Status endpoint test passed - Status: %s, Services: %+v",
			statusResponse.Status, statusResponse.Services)
	})

	// Test 7: Metrics endpoint
	t.Run("MetricsEndpoint", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/v1/metrics")
		if err != nil {
			t.Fatalf("Failed to make metrics request: %v", err)
		}
		defer resp.Body.Close()

		// Validate response status
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var metricsResponse struct {
			Timestamp string `json:"timestamp"`
			Metrics   struct {
				Requests struct {
					Total      int     `json:"total"`
					Successful int     `json:"successful"`
					Failed     int     `json:"failed"`
					Rate       float64 `json:"rate_per_second"`
				} `json:"requests"`
				Latency struct {
					Average float64 `json:"average_ms"`
					P50     float64 `json:"p50_ms"`
					P95     float64 `json:"p95_ms"`
					P99     float64 `json:"p99_ms"`
				} `json:"latency"`
				Classification struct {
					TotalClassifications int     `json:"total_classifications"`
					AverageConfidence    float64 `json:"average_confidence"`
					SuccessRate          float64 `json:"success_rate"`
				} `json:"classification"`
			} `json:"metrics"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&metricsResponse); err != nil {
			t.Fatalf("Failed to decode metrics response: %v", err)
		}

		// Validate metrics response
		if metricsResponse.Metrics.Requests.Total < 0 {
			t.Error("Expected total requests to be non-negative")
		}

		if metricsResponse.Metrics.Latency.Average < 0 {
			t.Error("Expected average latency to be non-negative")
		}

		if metricsResponse.Metrics.Classification.SuccessRate < 0 || metricsResponse.Metrics.Classification.SuccessRate > 1 {
			t.Error("Expected success rate to be between 0 and 1")
		}

		t.Logf("✅ Metrics endpoint test passed - Total Requests: %d, Success Rate: %.2f%%, Avg Latency: %.2fms",
			metricsResponse.Metrics.Requests.Total,
			metricsResponse.Metrics.Classification.SuccessRate*100,
			metricsResponse.Metrics.Latency.Average)
	})

	// Test 8: API error handling
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

		// Test unsupported HTTP method
		resp, err = http.Get(server.URL + "/v1/classify")
		if err != nil {
			t.Fatalf("Failed to make GET request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405 for unsupported method, got %d", resp.StatusCode)
		}

		// Test non-existent endpoint
		resp, err = http.Get(server.URL + "/v1/nonexistent")
		if err != nil {
			t.Fatalf("Failed to make request to non-existent endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404 for non-existent endpoint, got %d", resp.StatusCode)
		}

		t.Logf("✅ API error handling test passed")
	})

	// Test 9: API performance under load
	t.Run("APIPerformanceLoad", func(t *testing.T) {
		// Test multiple concurrent requests
		concurrentRequests := 10
		results := make(chan error, concurrentRequests)

		for i := 0; i < concurrentRequests; i++ {
			go func(index int) {
				request := &shared.BusinessClassificationRequest{
					ID:           fmt.Sprintf("load-test-%d", index),
					BusinessName: fmt.Sprintf("Load Test Company %d", index),
					RequestedAt:  time.Now(),
				}

				requestBody, _ := json.Marshal(request)
				resp, err := http.Post(server.URL+"/v1/classify", "application/json",
					bytes.NewReader(requestBody))
				if err != nil {
					results <- err
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					results <- fmt.Errorf("expected status 200, got %d", resp.StatusCode)
					return
				}

				results <- nil
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
		if successRate < 0.9 { // Allow 10% failure rate
			t.Errorf("Expected success rate >= 90%%, got %.1f%%", successRate*100)
		}

		t.Logf("✅ API performance load test passed - Success: %d, Errors: %d, Success Rate: %.1f%%",
			successCount, errorCount, successRate*100)
	})

	// Test 10: API response time
	t.Run("APIResponseTime", func(t *testing.T) {
		request := &shared.BusinessClassificationRequest{
			ID:           "response-time-test",
			BusinessName: "Response Time Test Company",
			RequestedAt:  time.Now(),
		}

		requestBody, _ := json.Marshal(request)

		// Measure response time
		start := time.Now()
		resp, err := http.Post(server.URL+"/v1/classify", "application/json",
			bytes.NewReader(requestBody))
		duration := time.Since(start)
		defer resp.Body.Close()

		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Validate response time is reasonable
		if duration > 5*time.Second {
			t.Errorf("Expected response time < 5s, got %v", duration)
		}

		t.Logf("✅ API response time test passed - Duration: %v", duration)
	})
}

// handleClassificationEndpoint handles single classification API requests
func handleClassificationEndpoint(w http.ResponseWriter, r *http.Request, service *mocks.MockClassificationService) {
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

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", request.ID)
	json.NewEncoder(w).Encode(result)
}

// handleBatchClassificationEndpoint handles batch classification API requests
func handleBatchClassificationEndpoint(w http.ResponseWriter, r *http.Request, service *mocks.MockClassificationService) {
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
	successfulRequests := 0
	failedRequests := 0

	for _, request := range batchRequest.Requests {
		result, err := service.ClassifyBusiness(ctx, &request)
		if err != nil {
			failedRequests++
			continue
		}
		results = append(results, *result)
		successfulRequests++
	}

	response := struct {
		Results []shared.BusinessClassificationResponse `json:"results"`
		Summary struct {
			TotalRequests      int `json:"total_requests"`
			SuccessfulRequests int `json:"successful_requests"`
			FailedRequests     int `json:"failed_requests"`
		} `json:"summary"`
	}{
		Results: results,
		Summary: struct {
			TotalRequests      int `json:"total_requests"`
			SuccessfulRequests int `json:"successful_requests"`
			FailedRequests     int `json:"failed_requests"`
		}{
			TotalRequests:      len(batchRequest.Requests),
			SuccessfulRequests: successfulRequests,
			FailedRequests:     failedRequests,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleClassificationStatusEndpoint handles classification status API requests
func handleClassificationStatusEndpoint(w http.ResponseWriter, r *http.Request, service *mocks.MockClassificationService) {
	response := struct {
		Status    string `json:"status"`
		Timestamp string `json:"timestamp"`
		Version   string `json:"version"`
		Services  struct {
			Classification string `json:"classification"`
			Database       string `json:"database"`
			API            string `json:"api"`
		} `json:"services"`
		Metrics struct {
			TotalRequests      int     `json:"total_requests"`
			SuccessfulRequests int     `json:"successful_requests"`
			FailedRequests     int     `json:"failed_requests"`
			AverageLatency     float64 `json:"average_latency_ms"`
		} `json:"metrics"`
	}{
		Status:    "operational",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Services: struct {
			Classification string `json:"classification"`
			Database       string `json:"database"`
			API            string `json:"api"`
		}{
			Classification: "healthy",
			Database:       "healthy",
			API:            "healthy",
		},
		Metrics: struct {
			TotalRequests      int     `json:"total_requests"`
			SuccessfulRequests int     `json:"successful_requests"`
			FailedRequests     int     `json:"failed_requests"`
			AverageLatency     float64 `json:"average_latency_ms"`
		}{
			TotalRequests:      1000,
			SuccessfulRequests: 950,
			FailedRequests:     50,
			AverageLatency:     150.5,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleClassificationHistoryEndpoint handles classification history API requests
func handleClassificationHistoryEndpoint(w http.ResponseWriter, r *http.Request, service *mocks.MockClassificationService) {
	// Mock history data
	mockClassifications := []shared.BusinessClassificationResponse{
		{
			ID:           "hist-001",
			BusinessName: "Historical Company 1",
			Classifications: []shared.IndustryClassification{
				{
					IndustryCode:         "541511",
					IndustryName:         "Custom Computer Programming Services",
					ConfidenceScore:      0.92,
					ClassificationMethod: "keyword_matching",
				},
			},
			OverallConfidence: 0.92,
			CreatedAt:         time.Now().Add(-1 * time.Hour),
		},
		{
			ID:           "hist-002",
			BusinessName: "Historical Company 2",
			Classifications: []shared.IndustryClassification{
				{
					IndustryCode:         "541512",
					IndustryName:         "Computer Systems Design Services",
					ConfidenceScore:      0.88,
					ClassificationMethod: "keyword_matching",
				},
			},
			OverallConfidence: 0.88,
			CreatedAt:         time.Now().Add(-2 * time.Hour),
		},
	}

	response := struct {
		Classifications []shared.BusinessClassificationResponse `json:"classifications"`
		Pagination      struct {
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
			Total  int `json:"total"`
		} `json:"pagination"`
	}{
		Classifications: mockClassifications,
		Pagination: struct {
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
			Total  int `json:"total"`
		}{
			Limit:  10,
			Offset: 0,
			Total:  2,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealthEndpoint handles health check API requests
func handleHealthEndpoint(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Status    string    `json:"status"`
		Timestamp time.Time `json:"timestamp"`
		Version   string    `json:"version"`
		Uptime    string    `json:"uptime"`
	}{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0",
		Uptime:    "24h",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStatusEndpoint handles status API requests
func handleStatusEndpoint(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Status    string            `json:"status"`
		Timestamp time.Time         `json:"timestamp"`
		Version   string            `json:"version"`
		Uptime    string            `json:"uptime"`
		Services  map[string]string `json:"services"`
	}{
		Status:    "operational",
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0",
		Uptime:    "24h",
		Services: map[string]string{
			"classification": "healthy",
			"database":       "healthy",
			"api":            "healthy",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMetricsEndpoint handles metrics API requests
func handleMetricsEndpoint(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Timestamp string `json:"timestamp"`
		Metrics   struct {
			Requests struct {
				Total      int     `json:"total"`
				Successful int     `json:"successful"`
				Failed     int     `json:"failed"`
				Rate       float64 `json:"rate_per_second"`
			} `json:"requests"`
			Latency struct {
				Average float64 `json:"average_ms"`
				P50     float64 `json:"p50_ms"`
				P95     float64 `json:"p95_ms"`
				P99     float64 `json:"p99_ms"`
			} `json:"latency"`
			Classification struct {
				TotalClassifications int     `json:"total_classifications"`
				AverageConfidence    float64 `json:"average_confidence"`
				SuccessRate          float64 `json:"success_rate"`
			} `json:"classification"`
		} `json:"metrics"`
	}{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metrics: struct {
			Requests struct {
				Total      int     `json:"total"`
				Successful int     `json:"successful"`
				Failed     int     `json:"failed"`
				Rate       float64 `json:"rate_per_second"`
			} `json:"requests"`
			Latency struct {
				Average float64 `json:"average_ms"`
				P50     float64 `json:"p50_ms"`
				P95     float64 `json:"p95_ms"`
				P99     float64 `json:"p99_ms"`
			} `json:"latency"`
			Classification struct {
				TotalClassifications int     `json:"total_classifications"`
				AverageConfidence    float64 `json:"average_confidence"`
				SuccessRate          float64 `json:"success_rate"`
			} `json:"classification"`
		}{
			Requests: struct {
				Total      int     `json:"total"`
				Successful int     `json:"successful"`
				Failed     int     `json:"failed"`
				Rate       float64 `json:"rate_per_second"`
			}{
				Total:      1000,
				Successful: 950,
				Failed:     50,
				Rate:       10.5,
			},
			Latency: struct {
				Average float64 `json:"average_ms"`
				P50     float64 `json:"p50_ms"`
				P95     float64 `json:"p95_ms"`
				P99     float64 `json:"p99_ms"`
			}{
				Average: 150.5,
				P50:     120.0,
				P95:     300.0,
				P99:     500.0,
			},
			Classification: struct {
				TotalClassifications int     `json:"total_classifications"`
				AverageConfidence    float64 `json:"average_confidence"`
				SuccessRate          float64 `json:"success_rate"`
			}{
				TotalClassifications: 950,
				AverageConfidence:    0.85,
				SuccessRate:          0.95,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
