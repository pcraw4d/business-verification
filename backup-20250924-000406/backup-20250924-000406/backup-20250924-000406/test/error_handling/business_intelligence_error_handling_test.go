package error_handling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/api/routes"
)

func TestBusinessIntelligenceErrorHandling(t *testing.T) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	t.Run("Invalid JSON Handling", func(t *testing.T) {
		testInvalidJSONHandling(t, mux)
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		testMissingRequiredFields(t, mux)
	})

	t.Run("Invalid Data Types", func(t *testing.T) {
		testInvalidDataTypes(t, mux)
	})

	t.Run("Invalid Time Ranges", func(t *testing.T) {
		testInvalidTimeRanges(t, mux)
	})

	t.Run("Invalid Analysis Types", func(t *testing.T) {
		testInvalidAnalysisTypes(t, mux)
	})

	t.Run("Non-existent Resource Access", func(t *testing.T) {
		testNonExistentResourceAccess(t, mux)
	})

	t.Run("Concurrent Access Errors", func(t *testing.T) {
		testConcurrentAccessErrors(t, mux)
	})

	t.Run("Rate Limiting", func(t *testing.T) {
		testRateLimiting(t, mux)
	})

	t.Run("Memory Exhaustion", func(t *testing.T) {
		testMemoryExhaustion(t, mux)
	})

	t.Run("Network Timeout Simulation", func(t *testing.T) {
		testNetworkTimeoutSimulation(t, mux)
	})
}

func testInvalidJSONHandling(t *testing.T, mux *http.ServeMux) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "Invalid JSON in market analysis",
			method:         "POST",
			path:           "/v2/business-intelligence/market-analysis",
			body:           `{"business_id": "test", "industry": "Technology", "invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Malformed JSON in competitive analysis",
			method:         "POST",
			path:           "/v2/business-intelligence/competitive-analysis",
			body:           `{"business_id": "test", "industry": "Technology", "competitors": ["Competitor A", "Competitor B",]}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty JSON in growth analytics",
			method:         "POST",
			path:           "/v2/business-intelligence/growth-analytics",
			body:           ``,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON in aggregation",
			method:         "POST",
			path:           "/v2/business-intelligence/aggregation",
			body:           `{"business_id": "test", "industry": "Technology", "analysis_types": ["market_analysis", "competitive_analysis",]}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify error response structure
			var errorResponse map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err != nil {
				t.Errorf("Failed to unmarshal error response: %v", err)
			}

			if _, exists := errorResponse["error"]; !exists {
				t.Error("Expected 'error' field in response")
			}
		})
	}
}

func testMissingRequiredFields(t *testing.T, mux *http.ServeMux) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "Missing business_id in market analysis",
			method:         "POST",
			path:           "/v2/business-intelligence/market-analysis",
			body:           `{"industry": "Technology", "geographic_area": "North America"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing industry in competitive analysis",
			method:         "POST",
			path:           "/v2/business-intelligence/competitive-analysis",
			body:           `{"business_id": "test", "geographic_area": "North America", "competitors": ["Competitor A"]}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing geographic_area in growth analytics",
			method:         "POST",
			path:           "/v2/business-intelligence/growth-analytics",
			body:           `{"business_id": "test", "industry": "Technology"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing time_range in aggregation",
			method:         "POST",
			path:           "/v2/business-intelligence/aggregation",
			body:           `{"business_id": "test", "industry": "Technology", "geographic_area": "North America", "analysis_types": ["market_analysis"]}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing analysis_types in aggregation",
			method:         "POST",
			path:           "/v2/business-intelligence/aggregation",
			body:           `{"business_id": "test", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2023-01-01T00:00:00Z", "end_date": "2024-01-01T00:00:00Z"}}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify error response structure
			var errorResponse map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err != nil {
				t.Errorf("Failed to unmarshal error response: %v", err)
			}

			if _, exists := errorResponse["error"]; !exists {
				t.Error("Expected 'error' field in response")
			}
		})
	}
}

func testInvalidDataTypes(t *testing.T, mux *http.ServeMux) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "Invalid business_id type in market analysis",
			method:         "POST",
			path:           "/v2/business-intelligence/market-analysis",
			body:           `{"business_id": 123, "industry": "Technology", "geographic_area": "North America"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid industry type in competitive analysis",
			method:         "POST",
			path:           "/v2/business-intelligence/competitive-analysis",
			body:           `{"business_id": "test", "industry": 123, "geographic_area": "North America", "competitors": ["Competitor A"]}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid competitors type in competitive analysis",
			method:         "POST",
			path:           "/v2/business-intelligence/competitive-analysis",
			body:           `{"business_id": "test", "industry": "Technology", "geographic_area": "North America", "competitors": "Competitor A"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid analysis_types type in aggregation",
			method:         "POST",
			path:           "/v2/business-intelligence/aggregation",
			body:           `{"business_id": "test", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2023-01-01T00:00:00Z", "end_date": "2024-01-01T00:00:00Z"}, "analysis_types": "market_analysis"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify error response structure
			var errorResponse map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err != nil {
				t.Errorf("Failed to unmarshal error response: %v", err)
			}

			if _, exists := errorResponse["error"]; !exists {
				t.Error("Expected 'error' field in response")
			}
		})
	}
}

func testInvalidTimeRanges(t *testing.T, mux *http.ServeMux) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "End date before start date in market analysis",
			method:         "POST",
			path:           "/v2/business-intelligence/market-analysis",
			body:           `{"business_id": "test", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2024-01-01T00:00:00Z", "end_date": "2023-01-01T00:00:00Z"}}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid date format in competitive analysis",
			method:         "POST",
			path:           "/v2/business-intelligence/competitive-analysis",
			body:           `{"business_id": "test", "industry": "Technology", "geographic_area": "North America", "competitors": ["Competitor A"], "time_range": {"start_date": "invalid-date", "end_date": "2024-01-01T00:00:00Z"}}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing time_zone in growth analytics",
			method:         "POST",
			path:           "/v2/business-intelligence/growth-analytics",
			body:           `{"business_id": "test", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2023-01-01T00:00:00Z", "end_date": "2024-01-01T00:00:00Z"}}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Future start date in aggregation",
			method:         "POST",
			path:           "/v2/business-intelligence/aggregation",
			body:           `{"business_id": "test", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2030-01-01T00:00:00Z", "end_date": "2031-01-01T00:00:00Z"}, "analysis_types": ["market_analysis"]}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify error response structure
			var errorResponse map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err != nil {
				t.Errorf("Failed to unmarshal error response: %v", err)
			}

			if _, exists := errorResponse["error"]; !exists {
				t.Error("Expected 'error' field in response")
			}
		})
	}
}

func testInvalidAnalysisTypes(t *testing.T, mux *http.ServeMux) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "Empty analysis_types in aggregation",
			method:         "POST",
			path:           "/v2/business-intelligence/aggregation",
			body:           `{"business_id": "test", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2023-01-01T00:00:00Z", "end_date": "2024-01-01T00:00:00Z"}, "analysis_types": []}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid analysis type in aggregation",
			method:         "POST",
			path:           "/v2/business-intelligence/aggregation",
			body:           `{"business_id": "test", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2023-01-01T00:00:00Z", "end_date": "2024-01-01T00:00:00Z"}, "analysis_types": ["invalid_type"]}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Duplicate analysis types in aggregation",
			method:         "POST",
			path:           "/v2/business-intelligence/aggregation",
			body:           `{"business_id": "test", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2023-01-01T00:00:00Z", "end_date": "2024-01-01T00:00:00Z"}, "analysis_types": ["market_analysis", "market_analysis"]}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify error response structure
			var errorResponse map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err != nil {
				t.Errorf("Failed to unmarshal error response: %v", err)
			}

			if _, exists := errorResponse["error"]; !exists {
				t.Error("Expected 'error' field in response")
			}
		})
	}
}

func testNonExistentResourceAccess(t *testing.T, mux *http.ServeMux) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "Non-existent market analysis",
			method:         "GET",
			path:           "/v2/business-intelligence/market-analysis?id=non-existent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Non-existent competitive analysis",
			method:         "GET",
			path:           "/v2/business-intelligence/competitive-analysis?id=non-existent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Non-existent growth analytics",
			method:         "GET",
			path:           "/v2/business-intelligence/growth-analytics?id=non-existent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Non-existent aggregation",
			method:         "GET",
			path:           "/v2/business-intelligence/aggregation?id=non-existent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Non-existent job",
			method:         "GET",
			path:           "/v2/business-intelligence/market-analysis/jobs?id=non-existent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify error response structure
			var errorResponse map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err != nil {
				t.Errorf("Failed to unmarshal error response: %v", err)
			}

			if _, exists := errorResponse["error"]; !exists {
				t.Error("Expected 'error' field in response")
			}
		})
	}
}

func testConcurrentAccessErrors(t *testing.T, mux *http.ServeMux) {
	// Test concurrent access to the same resource
	concurrency := 10
	done := make(chan bool, concurrency)

	// Create a market analysis first
	requestBody := map[string]interface{}{
		"business_id":     "concurrent-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Market analysis creation failed: %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	analysisID := response["id"].(string)

	// Test concurrent access to the same analysis
	for i := 0; i < concurrency; i++ {
		go func() {
			defer func() { done <- true }()

			req := httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/market-analysis?id=%s", analysisID), nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Concurrent access failed: %d", w.Code)
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}
}

func testRateLimiting(t *testing.T, mux *http.ServeMux) {
	// Test rapid successive requests
	requestBody := map[string]interface{}{
		"business_id":     "rate-limit-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
	}

	body, _ := json.Marshal(requestBody)

	// Make rapid successive requests
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		// All requests should succeed since we're not implementing actual rate limiting in the test
		if w.Code != http.StatusOK {
			t.Errorf("Rate limiting test failed: %d", w.Code)
		}
	}
}

func testMemoryExhaustion(t *testing.T, mux *http.ServeMux) {
	// Test creating many analyses to simulate memory pressure
	numAnalyses := 1000

	for i := 0; i < numAnalyses; i++ {
		requestBody := map[string]interface{}{
			"business_id":     fmt.Sprintf("memory-test-business-%d", i),
			"industry":        "Technology",
			"geographic_area": "North America",
			"time_range": map[string]interface{}{
				"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
				"end_date":   time.Now().Format(time.RFC3339),
				"time_zone":  "UTC",
			},
		}

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Memory exhaustion test failed: %d", w.Code)
		}
	}

	// Test listing all analyses
	req := httptest.NewRequest("GET", "/v2/business-intelligence/market-analyses", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Memory exhaustion test - listing failed: %d", w.Code)
	}
}

func testNetworkTimeoutSimulation(t *testing.T, mux *http.ServeMux) {
	// Test with very large request body to simulate network issues
	largeRequestBody := map[string]interface{}{
		"business_id":     "timeout-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
		"large_data": make([]string, 10000), // Large data to simulate network issues
	}

	body, _ := json.Marshal(largeRequestBody)
	req := httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	// Should still succeed since we're not implementing actual network timeout simulation
	if w.Code != http.StatusOK {
		t.Errorf("Network timeout simulation test failed: %d", w.Code)
	}
}

// Test error recovery and graceful degradation
func TestErrorRecovery(t *testing.T) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	// Test that the system can recover from errors
	t.Run("Recovery from Invalid Request", func(t *testing.T) {
		// Make an invalid request
		req := httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis", bytes.NewBufferString(`{"invalid": json}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		// Make a valid request after the error
		validRequestBody := map[string]interface{}{
			"business_id":     "recovery-test-business",
			"industry":        "Technology",
			"geographic_area": "North America",
			"time_range": map[string]interface{}{
				"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
				"end_date":   time.Now().Format(time.RFC3339),
				"time_zone":  "UTC",
			},
		}

		body, _ := json.Marshal(validRequestBody)
		req = httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Recovery test failed: %d", w.Code)
		}
	})

	// Test that the system can handle mixed valid and invalid requests
	t.Run("Mixed Valid and Invalid Requests", func(t *testing.T) {
		requests := []struct {
			body           string
			expectedStatus int
		}{
			{`{"business_id": "test1", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2023-01-01T00:00:00Z", "end_date": "2024-01-01T00:00:00Z"}}`, http.StatusOK},
			{`{"invalid": json}`, http.StatusBadRequest},
			{`{"business_id": "test2", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2023-01-01T00:00:00Z", "end_date": "2024-01-01T00:00:00Z"}}`, http.StatusOK},
			{`{"business_id": "test3", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2023-01-01T00:00:00Z", "end_date": "2024-01-01T00:00:00Z"}}`, http.StatusOK},
		}

		for i, req := range requests {
			httpReq := httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis", bytes.NewBufferString(req.body))
			httpReq.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, httpReq)

			if w.Code != req.expectedStatus {
				t.Errorf("Request %d: Expected status %d, got %d", i, req.expectedStatus, w.Code)
			}
		}
	})
}
