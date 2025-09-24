package integration

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

func TestBusinessIntelligenceIntegration(t *testing.T) {
	// Create handlers
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()

	// Create route config
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	// Create mux and register routes
	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	t.Run("Market Analysis Workflow", func(t *testing.T) {
		// Test complete market analysis workflow
		testMarketAnalysisWorkflow(t, mux)
	})

	t.Run("Competitive Analysis Workflow", func(t *testing.T) {
		// Test complete competitive analysis workflow
		testCompetitiveAnalysisWorkflow(t, mux)
	})

	t.Run("Growth Analytics Workflow", func(t *testing.T) {
		// Test complete growth analytics workflow
		testGrowthAnalyticsWorkflow(t, mux)
	})

	t.Run("Business Intelligence Aggregation Workflow", func(t *testing.T) {
		// Test complete aggregation workflow
		testBusinessIntelligenceAggregationWorkflow(t, mux)
	})

	t.Run("Cross-Service Integration", func(t *testing.T) {
		// Test integration between different services
		testCrossServiceIntegration(t, mux)
	})
}

func testMarketAnalysisWorkflow(t *testing.T, mux *http.ServeMux) {
	// 1. Create market analysis
	requestBody := map[string]interface{}{
		"business_id":     "integration-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
		"parameters": map[string]interface{}{
			"market_size_focus": "total",
		},
		"options": map[string]interface{}{
			"real_time":     true,
			"batch_mode":    false,
			"parallel":      true,
			"notifications": true,
			"audit_trail":   true,
			"monitoring":    true,
			"validation":    true,
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

	// 2. Retrieve the analysis
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/market-analysis?id=%s", analysisID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Market analysis retrieval failed: %d", w.Code)
	}

	// 3. List all analyses
	req = httptest.NewRequest("GET", "/v2/business-intelligence/market-analyses", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Market analyses listing failed: %d", w.Code)
	}

	var listResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &listResponse); err != nil {
		t.Fatalf("Failed to unmarshal list response: %v", err)
	}

	analyses, ok := listResponse["analyses"].([]interface{})
	if !ok || len(analyses) == 0 {
		t.Fatal("Expected at least one analysis in the list")
	}

	// 4. Create background job
	req = httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis/jobs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Market analysis job creation failed: %d", w.Code)
	}

	var jobResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &jobResponse); err != nil {
		t.Fatalf("Failed to unmarshal job response: %v", err)
	}

	jobID := jobResponse["job_id"].(string)

	// 5. Check job status
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/market-analysis/jobs?id=%s", jobID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Job status check failed: %d", w.Code)
	}

	// 6. List jobs
	req = httptest.NewRequest("GET", "/v2/business-intelligence/market-analysis/jobs/list", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Jobs listing failed: %d", w.Code)
	}
}

func testCompetitiveAnalysisWorkflow(t *testing.T, mux *http.ServeMux) {
	// 1. Create competitive analysis
	requestBody := map[string]interface{}{
		"business_id":     "integration-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"competitors":     []string{"Competitor A", "Competitor B", "Competitor C"},
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/v2/business-intelligence/competitive-analysis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Competitive analysis creation failed: %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	analysisID := response["id"].(string)

	// 2. Retrieve the analysis
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/competitive-analysis?id=%s", analysisID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Competitive analysis retrieval failed: %d", w.Code)
	}

	// 3. List all analyses
	req = httptest.NewRequest("GET", "/v2/business-intelligence/competitive-analyses", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Competitive analyses listing failed: %d", w.Code)
	}

	// 4. Create background job
	req = httptest.NewRequest("POST", "/v2/business-intelligence/competitive-analysis/jobs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Competitive analysis job creation failed: %d", w.Code)
	}

	var jobResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &jobResponse); err != nil {
		t.Fatalf("Failed to unmarshal job response: %v", err)
	}

	jobID := jobResponse["job_id"].(string)

	// 5. Check job status
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/competitive-analysis/jobs?id=%s", jobID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Job status check failed: %d", w.Code)
	}

	// 6. List jobs
	req = httptest.NewRequest("GET", "/v2/business-intelligence/competitive-analysis/jobs/list", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Jobs listing failed: %d", w.Code)
	}
}

func testGrowthAnalyticsWorkflow(t *testing.T, mux *http.ServeMux) {
	// 1. Create growth analytics
	requestBody := map[string]interface{}{
		"business_id":     "integration-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/v2/business-intelligence/growth-analytics", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Growth analytics creation failed: %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	analysisID := response["id"].(string)

	// 2. Retrieve the analysis
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/growth-analytics?id=%s", analysisID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Growth analytics retrieval failed: %d", w.Code)
	}

	// 3. List all analyses
	req = httptest.NewRequest("GET", "/v2/business-intelligence/growth-analytics/list", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Growth analytics listing failed: %d", w.Code)
	}

	// 4. Create background job
	req = httptest.NewRequest("POST", "/v2/business-intelligence/growth-analytics/jobs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Growth analytics job creation failed: %d", w.Code)
	}

	var jobResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &jobResponse); err != nil {
		t.Fatalf("Failed to unmarshal job response: %v", err)
	}

	jobID := jobResponse["job_id"].(string)

	// 5. Check job status
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/growth-analytics/jobs?id=%s", jobID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Job status check failed: %d", w.Code)
	}

	// 6. List jobs
	req = httptest.NewRequest("GET", "/v2/business-intelligence/growth-analytics/jobs/list", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Jobs listing failed: %d", w.Code)
	}
}

func testBusinessIntelligenceAggregationWorkflow(t *testing.T, mux *http.ServeMux) {
	// 1. Create business intelligence aggregation
	requestBody := map[string]interface{}{
		"business_id":     "integration-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
		"analysis_types": []string{"market_analysis", "competitive_analysis", "growth_analytics"},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/v2/business-intelligence/aggregation", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Business intelligence aggregation creation failed: %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	aggregationID := response["id"].(string)

	// 2. Retrieve the aggregation
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/aggregation?id=%s", aggregationID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Business intelligence aggregation retrieval failed: %d", w.Code)
	}

	// 3. List all aggregations
	req = httptest.NewRequest("GET", "/v2/business-intelligence/aggregations", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Business intelligence aggregations listing failed: %d", w.Code)
	}

	// 4. Create background job
	req = httptest.NewRequest("POST", "/v2/business-intelligence/aggregation/jobs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Business intelligence aggregation job creation failed: %d", w.Code)
	}

	var jobResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &jobResponse); err != nil {
		t.Fatalf("Failed to unmarshal job response: %v", err)
	}

	jobID := jobResponse["job_id"].(string)

	// 5. Check job status
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/aggregation/jobs?id=%s", jobID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Job status check failed: %d", w.Code)
	}

	// 6. List jobs
	req = httptest.NewRequest("GET", "/v2/business-intelligence/aggregation/jobs/list", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Jobs listing failed: %d", w.Code)
	}
}

func testCrossServiceIntegration(t *testing.T, mux *http.ServeMux) {
	// Test that different services can work together
	// This test creates analyses from different services and then aggregates them

	// 1. Create market analysis
	marketRequest := map[string]interface{}{
		"business_id":     "cross-service-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
	}

	body, _ := json.Marshal(marketRequest)
	req := httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Market analysis creation failed: %d", w.Code)
	}

	// 2. Create competitive analysis
	competitiveRequest := map[string]interface{}{
		"business_id":     "cross-service-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"competitors":     []string{"Competitor A", "Competitor B"},
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
	}

	body, _ = json.Marshal(competitiveRequest)
	req = httptest.NewRequest("POST", "/v2/business-intelligence/competitive-analysis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Competitive analysis creation failed: %d", w.Code)
	}

	// 3. Create growth analytics
	growthRequest := map[string]interface{}{
		"business_id":     "cross-service-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
	}

	body, _ = json.Marshal(growthRequest)
	req = httptest.NewRequest("POST", "/v2/business-intelligence/growth-analytics", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Growth analytics creation failed: %d", w.Code)
	}

	// 4. Create aggregation that combines all three
	aggregationRequest := map[string]interface{}{
		"business_id":     "cross-service-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
		"analysis_types": []string{"market_analysis", "competitive_analysis", "growth_analytics"},
	}

	body, _ = json.Marshal(aggregationRequest)
	req = httptest.NewRequest("POST", "/v2/business-intelligence/aggregation", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Business intelligence aggregation creation failed: %d", w.Code)
	}

	var aggregationResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &aggregationResponse); err != nil {
		t.Fatalf("Failed to unmarshal aggregation response: %v", err)
	}

	// 5. Validate that aggregation contains all three analysis types
	analyses, ok := aggregationResponse["analyses"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'analyses' to be an object")
	}

	expectedAnalysisTypes := []string{"market_analysis", "competitive_analysis", "growth_analytics"}
	for _, analysisType := range expectedAnalysisTypes {
		if _, exists := analyses[analysisType]; !exists {
			t.Errorf("Expected analysis type '%s' to be present in aggregation", analysisType)
		}
	}

	// 6. Validate aggregation insights and recommendations
	insights, ok := aggregationResponse["insights"].([]interface{})
	if !ok || len(insights) == 0 {
		t.Fatal("Expected non-empty insights array")
	}

	recommendations, ok := aggregationResponse["recommendations"].([]interface{})
	if !ok || len(recommendations) == 0 {
		t.Fatal("Expected non-empty recommendations array")
	}

	summary, ok := aggregationResponse["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'summary' to be an object")
	}

	// Validate summary structure
	expectedSummaryFields := []string{"executive_summary", "key_findings", "strategic_priorities", "success_metrics", "next_steps"}
	for _, field := range expectedSummaryFields {
		if _, exists := summary[field]; !exists {
			t.Errorf("Expected field '%s' not found in summary", field)
		}
	}
}

// Test error handling across all services
func TestBusinessIntelligenceErrorHandling(t *testing.T) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

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
			name:           "Missing required fields in competitive analysis",
			method:         "POST",
			path:           "/v2/business-intelligence/competitive-analysis",
			body:           `{"business_id": "test", "industry": "Technology"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid time range in growth analytics",
			method:         "POST",
			path:           "/v2/business-intelligence/growth-analytics",
			body:           `{"business_id": "test", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2024-01-01T00:00:00Z", "end_date": "2023-01-01T00:00:00Z"}}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty analysis types in aggregation",
			method:         "POST",
			path:           "/v2/business-intelligence/aggregation",
			body:           `{"business_id": "test", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2023-01-01T00:00:00Z", "end_date": "2024-01-01T00:00:00Z"}, "analysis_types": []}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent job ID",
			method:         "GET",
			path:           "/v2/business-intelligence/market-analysis/jobs?id=non-existent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// Test concurrent access to business intelligence services
func TestBusinessIntelligenceConcurrency(t *testing.T) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	// Test concurrent market analysis creation
	t.Run("Concurrent Market Analysis Creation", func(t *testing.T) {
		concurrency := 10
		done := make(chan bool, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(index int) {
				defer func() { done <- true }()

				requestBody := map[string]interface{}{
					"business_id":     fmt.Sprintf("concurrent-test-business-%d", index),
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
					t.Errorf("Concurrent market analysis creation failed: %d", w.Code)
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < concurrency; i++ {
			<-done
		}
	})

	// Test concurrent job creation and status checking
	t.Run("Concurrent Job Operations", func(t *testing.T) {
		concurrency := 5
		done := make(chan bool, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(index int) {
				defer func() { done <- true }()

				// Create job
				requestBody := map[string]interface{}{
					"business_id":     fmt.Sprintf("concurrent-job-test-business-%d", index),
					"industry":        "Technology",
					"geographic_area": "North America",
					"time_range": map[string]interface{}{
						"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
						"end_date":   time.Now().Format(time.RFC3339),
						"time_zone":  "UTC",
					},
				}

				body, _ := json.Marshal(requestBody)
				req := httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis/jobs", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				mux.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Concurrent job creation failed: %d", w.Code)
					return
				}

				var jobResponse map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &jobResponse); err != nil {
					t.Errorf("Failed to unmarshal job response: %v", err)
					return
				}

				jobID := jobResponse["job_id"].(string)

				// Check job status
				req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/market-analysis/jobs?id=%s", jobID), nil)
				w = httptest.NewRecorder()

				mux.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Concurrent job status check failed: %d", w.Code)
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < concurrency; i++ {
			<-done
		}
	})
}
