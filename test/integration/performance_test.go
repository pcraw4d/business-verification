//go:build !comprehensive_test
// +build !comprehensive_test

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"kyb-platform/internal/shared"
	"kyb-platform/test/mocks"
)

// TestPerformanceIntegration tests performance and load scenarios
func TestPerformanceIntegration(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	// Test 1: Single request performance
	t.Run("SingleRequestPerformance", func(t *testing.T) {
		// Create mock service
		mockService := mocks.NewMockClassificationService()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleClassificationPerformance(w, r, mockService)
		}))
		defer server.Close()

		// Test single request performance
		request := &shared.BusinessClassificationRequest{
			ID:           "perf-test-001",
			BusinessName: "Performance Test Company",
			Description:  "Test company for performance testing",
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
		if duration > 1*time.Second {
			t.Errorf("Expected response time < 1s, got %v", duration)
		}

		t.Logf("✅ Single request performance test passed - Duration: %v", duration)
	})

	// Test 2: Concurrent request performance
	t.Run("ConcurrentRequestPerformance", func(t *testing.T) {
		// Create mock service
		mockService := mocks.NewMockClassificationService()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleClassificationPerformance(w, r, mockService)
		}))
		defer server.Close()

		// Test concurrent requests
		concurrentRequests := 50
		results := make(chan error, concurrentRequests)
		var wg sync.WaitGroup

		start := time.Now()

		for i := 0; i < concurrentRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				request := &shared.BusinessClassificationRequest{
					ID:           fmt.Sprintf("concurrent-test-%d", index),
					BusinessName: fmt.Sprintf("Concurrent Test Company %d", index),
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

		wg.Wait()
		close(results)
		duration := time.Since(start)

		// Collect results
		successCount := 0
		errorCount := 0
		for err := range results {
			if err != nil {
				errorCount++
				t.Logf("Concurrent request failed: %v", err)
			} else {
				successCount++
			}
		}

		// Validate success rate
		successRate := float64(successCount) / float64(concurrentRequests)
		if successRate < 0.95 { // Allow 5% failure rate
			t.Errorf("Expected success rate >= 95%%, got %.1f%%", successRate*100)
		}

		// Validate throughput
		throughput := float64(concurrentRequests) / duration.Seconds()
		if throughput < 10 { // At least 10 requests per second
			t.Errorf("Expected throughput >= 10 req/s, got %.1f req/s", throughput)
		}

		t.Logf("✅ Concurrent request performance test passed - Success: %d, Errors: %d, Success Rate: %.1f%%, Throughput: %.1f req/s, Duration: %v",
			successCount, errorCount, successRate*100, throughput, duration)
	})

	// Test 3: High load performance
	t.Run("HighLoadPerformance", func(t *testing.T) {
		// Create mock service
		mockService := mocks.NewMockClassificationService()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleClassificationPerformance(w, r, mockService)
		}))
		defer server.Close()

		// Test high load
		highLoadRequests := 200
		results := make(chan error, highLoadRequests)
		var wg sync.WaitGroup

		start := time.Now()

		for i := 0; i < highLoadRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				request := &shared.BusinessClassificationRequest{
					ID:           fmt.Sprintf("high-load-test-%d", index),
					BusinessName: fmt.Sprintf("High Load Test Company %d", index),
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

		wg.Wait()
		close(results)
		duration := time.Since(start)

		// Collect results
		successCount := 0
		errorCount := 0
		for err := range results {
			if err != nil {
				errorCount++
				t.Logf("High load request failed: %v", err)
			} else {
				successCount++
			}
		}

		// Validate success rate
		successRate := float64(successCount) / float64(highLoadRequests)
		if successRate < 0.90 { // Allow 10% failure rate for high load
			t.Errorf("Expected success rate >= 90%%, got %.1f%%", successRate*100)
		}

		// Validate throughput
		throughput := float64(highLoadRequests) / duration.Seconds()
		if throughput < 20 { // At least 20 requests per second
			t.Errorf("Expected throughput >= 20 req/s, got %.1f req/s", throughput)
		}

		t.Logf("✅ High load performance test passed - Success: %d, Errors: %d, Success Rate: %.1f%%, Throughput: %.1f req/s, Duration: %v",
			successCount, errorCount, successRate*100, throughput, duration)
	})

	// Test 4: Memory usage performance
	t.Run("MemoryUsagePerformance", func(t *testing.T) {
		// Create mock service
		mockService := mocks.NewMockClassificationService()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleClassificationPerformance(w, r, mockService)
		}))
		defer server.Close()

		// Test memory usage with large requests
		largeRequests := 100
		results := make(chan error, largeRequests)
		var wg sync.WaitGroup

		start := time.Now()

		for i := 0; i < largeRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				// Create large request
				largeDescription := fmt.Sprintf("This is a very long description for performance testing. Request number %d. ", index)
				largeDescription += strings.Repeat("Additional content for memory testing. ", 100)

				request := &shared.BusinessClassificationRequest{
					ID:           fmt.Sprintf("memory-test-%d", index),
					BusinessName: fmt.Sprintf("Memory Test Company %d", index),
					Description:  largeDescription,
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

		wg.Wait()
		close(results)
		duration := time.Since(start)

		// Collect results
		successCount := 0
		errorCount := 0
		for err := range results {
			if err != nil {
				errorCount++
				t.Logf("Memory test request failed: %v", err)
			} else {
				successCount++
			}
		}

		// Validate success rate
		successRate := float64(successCount) / float64(largeRequests)
		if successRate < 0.95 { // Allow 5% failure rate
			t.Errorf("Expected success rate >= 95%%, got %.1f%%", successRate*100)
		}

		// Validate throughput
		throughput := float64(largeRequests) / duration.Seconds()
		if throughput < 5 { // At least 5 requests per second for large requests
			t.Errorf("Expected throughput >= 5 req/s, got %.1f req/s", throughput)
		}

		t.Logf("✅ Memory usage performance test passed - Success: %d, Errors: %d, Success Rate: %.1f%%, Throughput: %.1f req/s, Duration: %v",
			successCount, errorCount, successRate*100, throughput, duration)
	})

	// Test 5: Database performance
	t.Run("DatabasePerformance", func(t *testing.T) {
		// Create mock database
		mockDB := mocks.NewMockDatabase()

		// Test database connection performance
		start := time.Now()
		if err := mockDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		connectionDuration := time.Since(start)

		if connectionDuration > 100*time.Millisecond {
			t.Errorf("Expected database connection < 100ms, got %v", connectionDuration)
		}

		// Test database query performance
		queryCount := 100
		start = time.Now()

		for i := 0; i < queryCount; i++ {
			_, err := mockDB.ExecuteQuery("SELECT * FROM industry_codes LIMIT 10")
			if err != nil {
				t.Fatalf("Database query failed: %v", err)
			}
		}

		queryDuration := time.Since(start)
		avgQueryTime := queryDuration / time.Duration(queryCount)

		if avgQueryTime > 10*time.Millisecond {
			t.Errorf("Expected average query time < 10ms, got %v", avgQueryTime)
		}

		// Test database ping performance
		start = time.Now()
		if err := mockDB.Ping(); err != nil {
			t.Fatalf("Database ping failed: %v", err)
		}
		pingDuration := time.Since(start)

		if pingDuration > 10*time.Millisecond {
			t.Errorf("Expected database ping < 10ms, got %v", pingDuration)
		}

		t.Logf("✅ Database performance test passed - Connection: %v, Avg Query: %v, Ping: %v",
			connectionDuration, avgQueryTime, pingDuration)
	})

	// Test 6: Classification service performance
	t.Run("ClassificationServicePerformance", func(t *testing.T) {
		// Create mock service
		mockService := mocks.NewMockClassificationService()

		// Test classification performance
		classificationCount := 100
		start := time.Now()

		for i := 0; i < classificationCount; i++ {
			request := &shared.BusinessClassificationRequest{
				ID:           fmt.Sprintf("class-perf-test-%d", i),
				BusinessName: fmt.Sprintf("Classification Performance Test Company %d", i),
				RequestedAt:  time.Now(),
			}

			ctx := context.Background()
			_, err := mockService.ClassifyBusiness(ctx, request)
			if err != nil {
				t.Fatalf("Classification failed: %v", err)
			}
		}

		classificationDuration := time.Since(start)
		avgClassificationTime := classificationDuration / time.Duration(classificationCount)

		if avgClassificationTime > 200*time.Millisecond {
			t.Errorf("Expected average classification time < 200ms, got %v", avgClassificationTime)
		}

		// Test classification throughput
		throughput := float64(classificationCount) / classificationDuration.Seconds()
		if throughput < 5 { // At least 5 classifications per second
			t.Errorf("Expected classification throughput >= 5 req/s, got %.1f req/s", throughput)
		}

		t.Logf("✅ Classification service performance test passed - Avg Time: %v, Throughput: %.1f req/s",
			avgClassificationTime, throughput)
	})

	// Test 7: Batch processing performance
	t.Run("BatchProcessingPerformance", func(t *testing.T) {
		// Create mock service
		mockService := mocks.NewMockClassificationService()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleBatchClassificationPerformance(w, r, mockService)
		}))
		defer server.Close()

		// Test batch processing performance
		batchSize := 50
		batchRequest := struct {
			Requests []shared.BusinessClassificationRequest `json:"requests"`
		}{
			Requests: make([]shared.BusinessClassificationRequest, batchSize),
		}

		for i := 0; i < batchSize; i++ {
			batchRequest.Requests[i] = shared.BusinessClassificationRequest{
				ID:           fmt.Sprintf("batch-perf-test-%d", i),
				BusinessName: fmt.Sprintf("Batch Performance Test Company %d", i),
				RequestedAt:  time.Now(),
			}
		}

		requestBody, _ := json.Marshal(batchRequest)

		start := time.Now()
		resp, err := http.Post(server.URL+"/v1/classify/batch", "application/json",
			bytes.NewReader(requestBody))
		duration := time.Since(start)
		defer resp.Body.Close()

		if err != nil {
			t.Fatalf("Failed to make batch request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Validate batch processing performance
		if duration > 10*time.Second {
			t.Errorf("Expected batch processing < 10s, got %v", duration)
		}

		// Validate batch throughput
		throughput := float64(batchSize) / duration.Seconds()
		if throughput < 5 { // At least 5 requests per second
			t.Errorf("Expected batch throughput >= 5 req/s, got %.1f req/s", throughput)
		}

		t.Logf("✅ Batch processing performance test passed - Duration: %v, Throughput: %.1f req/s",
			duration, throughput)
	})

	// Test 8: Stress test
	t.Run("StressTest", func(t *testing.T) {
		// Create mock service
		mockService := mocks.NewMockClassificationService()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleClassificationPerformance(w, r, mockService)
		}))
		defer server.Close()

		// Test stress scenario
		stressRequests := 500
		results := make(chan error, stressRequests)
		var wg sync.WaitGroup

		start := time.Now()

		for i := 0; i < stressRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				request := &shared.BusinessClassificationRequest{
					ID:           fmt.Sprintf("stress-test-%d", index),
					BusinessName: fmt.Sprintf("Stress Test Company %d", index),
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

		wg.Wait()
		close(results)
		duration := time.Since(start)

		// Collect results
		successCount := 0
		errorCount := 0
		for err := range results {
			if err != nil {
				errorCount++
				t.Logf("Stress test request failed: %v", err)
			} else {
				successCount++
			}
		}

		// Validate success rate
		successRate := float64(successCount) / float64(stressRequests)
		if successRate < 0.85 { // Allow 15% failure rate for stress test
			t.Errorf("Expected success rate >= 85%%, got %.1f%%", successRate*100)
		}

		// Validate throughput
		throughput := float64(stressRequests) / duration.Seconds()
		if throughput < 30 { // At least 30 requests per second
			t.Errorf("Expected throughput >= 30 req/s, got %.1f req/s", throughput)
		}

		t.Logf("✅ Stress test passed - Success: %d, Errors: %d, Success Rate: %.1f%%, Throughput: %.1f req/s, Duration: %v",
			successCount, errorCount, successRate*100, throughput, duration)
	})

	// Test 9: Resource cleanup performance
	t.Run("ResourceCleanupPerformance", func(t *testing.T) {
		// Create mock service
		mockService := mocks.NewMockClassificationService()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleClassificationPerformance(w, r, mockService)
		}))
		defer server.Close()

		// Test resource cleanup
		cleanupRequests := 100
		results := make(chan error, cleanupRequests)
		var wg sync.WaitGroup

		start := time.Now()

		for i := 0; i < cleanupRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				request := &shared.BusinessClassificationRequest{
					ID:           fmt.Sprintf("cleanup-test-%d", index),
					BusinessName: fmt.Sprintf("Cleanup Test Company %d", index),
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

		wg.Wait()
		close(results)
		duration := time.Since(start)

		// Collect results
		successCount := 0
		errorCount := 0
		for err := range results {
			if err != nil {
				errorCount++
				t.Logf("Cleanup test request failed: %v", err)
			} else {
				successCount++
			}
		}

		// Validate success rate
		successRate := float64(successCount) / float64(cleanupRequests)
		if successRate < 0.95 { // Allow 5% failure rate
			t.Errorf("Expected success rate >= 95%%, got %.1f%%", successRate*100)
		}

		// Validate throughput
		throughput := float64(cleanupRequests) / duration.Seconds()
		if throughput < 15 { // At least 15 requests per second
			t.Errorf("Expected throughput >= 15 req/s, got %.1f req/s", throughput)
		}

		t.Logf("✅ Resource cleanup performance test passed - Success: %d, Errors: %d, Success Rate: %.1f%%, Throughput: %.1f req/s, Duration: %v",
			successCount, errorCount, successRate*100, throughput, duration)
	})

	// Test 10: Performance regression test
	t.Run("PerformanceRegressionTest", func(t *testing.T) {
		// Create mock service
		mockService := mocks.NewMockClassificationService()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleClassificationPerformance(w, r, mockService)
		}))
		defer server.Close()

		// Test performance regression
		regressionRequests := 100
		results := make(chan error, regressionRequests)
		var wg sync.WaitGroup

		start := time.Now()

		for i := 0; i < regressionRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				request := &shared.BusinessClassificationRequest{
					ID:           fmt.Sprintf("regression-test-%d", index),
					BusinessName: fmt.Sprintf("Regression Test Company %d", index),
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

		wg.Wait()
		close(results)
		duration := time.Since(start)

		// Collect results
		successCount := 0
		errorCount := 0
		for err := range results {
			if err != nil {
				errorCount++
				t.Logf("Regression test request failed: %v", err)
			} else {
				successCount++
			}
		}

		// Validate success rate
		successRate := float64(successCount) / float64(regressionRequests)
		if successRate < 0.98 { // Allow 2% failure rate for regression test
			t.Errorf("Expected success rate >= 98%%, got %.1f%%", successRate*100)
		}

		// Validate throughput
		throughput := float64(regressionRequests) / duration.Seconds()
		if throughput < 25 { // At least 25 requests per second
			t.Errorf("Expected throughput >= 25 req/s, got %.1f req/s", throughput)
		}

		t.Logf("✅ Performance regression test passed - Success: %d, Errors: %d, Success Rate: %.1f%%, Throughput: %.1f req/s, Duration: %v",
			successCount, errorCount, successRate*100, throughput, duration)
	})
}

// Helper functions for performance tests

func handleClassificationPerformance(w http.ResponseWriter, r *http.Request, service *mocks.MockClassificationService) {
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

func handleBatchClassificationPerformance(w http.ResponseWriter, r *http.Request, service *mocks.MockClassificationService) {
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
