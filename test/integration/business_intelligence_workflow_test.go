//go:build !comprehensive_test
// +build !comprehensive_test

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/routes"
)

// TestBusinessIntelligenceEndToEndWorkflow tests the complete business intelligence workflow
func TestBusinessIntelligenceEndToEndWorkflow(t *testing.T) {
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

	t.Run("Complete Market Analysis Workflow", func(t *testing.T) {
		testCompleteMarketAnalysisWorkflow(t, mux)
	})

	t.Run("Complete Competitive Analysis Workflow", func(t *testing.T) {
		testCompleteCompetitiveAnalysisWorkflow(t, mux)
	})

	t.Run("Complete Growth Analytics Workflow", func(t *testing.T) {
		testCompleteGrowthAnalyticsWorkflow(t, mux)
	})

	t.Run("Complete Business Intelligence Aggregation Workflow", func(t *testing.T) {
		testCompleteBusinessIntelligenceAggregationWorkflow(t, mux)
	})

	t.Run("Cross-Service Integration Workflow", func(t *testing.T) {
		testCrossServiceIntegrationWorkflow(t, mux)
	})
}

func testCompleteMarketAnalysisWorkflow(t *testing.T, mux *http.ServeMux) {
	// Step 1: Create market analysis request
	requestBody := map[string]interface{}{
		"business_id":     "workflow-test-business",
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
		t.Fatalf("Market analysis creation failed: %d, response: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	analysisID := response["id"].(string)
	t.Logf("Created market analysis with ID: %s", analysisID)

	// Step 2: Retrieve the analysis
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/market-analysis?id=%s", analysisID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Market analysis retrieval failed: %d", w.Code)
	}

	var retrievedResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &retrievedResponse); err != nil {
		t.Fatalf("Failed to unmarshal retrieved response: %v", err)
	}

	// Validate the retrieved analysis
	if retrievedResponse["id"] != analysisID {
		t.Errorf("Expected analysis ID %s, got %s", analysisID, retrievedResponse["id"])
	}

	// Step 3: List all analyses
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

	// Step 4: Create background job
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
	t.Logf("Created market analysis job with ID: %s", jobID)

	// Step 5: Check job status
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/market-analysis/jobs?id=%s", jobID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Job status check failed: %d", w.Code)
	}

	var jobStatusResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &jobStatusResponse); err != nil {
		t.Fatalf("Failed to unmarshal job status response: %v", err)
	}

	// Validate job status
	if jobStatusResponse["job_id"] != jobID {
		t.Errorf("Expected job ID %s, got %s", jobID, jobStatusResponse["job_id"])
	}

	// Step 6: List jobs
	req = httptest.NewRequest("GET", "/v2/business-intelligence/market-analysis/jobs/list", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Jobs listing failed: %d", w.Code)
	}

	var jobsListResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &jobsListResponse); err != nil {
		t.Fatalf("Failed to unmarshal jobs list response: %v", err)
	}

	jobs, ok := jobsListResponse["jobs"].([]interface{})
	if !ok || len(jobs) == 0 {
		t.Fatal("Expected at least one job in the list")
	}

	t.Logf("Market analysis workflow completed successfully")
}

func testCompleteCompetitiveAnalysisWorkflow(t *testing.T, mux *http.ServeMux) {
	// Step 1: Create competitive analysis request
	requestBody := map[string]interface{}{
		"business_id":     "workflow-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"competitors":     []string{"Competitor A", "Competitor B", "Competitor C"},
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
		"parameters": map[string]interface{}{
			"analysis_depth": "comprehensive",
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
	req := httptest.NewRequest("POST", "/v2/business-intelligence/competitive-analysis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Competitive analysis creation failed: %d, response: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	analysisID := response["id"].(string)
	t.Logf("Created competitive analysis with ID: %s", analysisID)

	// Step 2: Retrieve the analysis
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/competitive-analysis?id=%s", analysisID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Competitive analysis retrieval failed: %d", w.Code)
	}

	// Step 3: List all analyses
	req = httptest.NewRequest("GET", "/v2/business-intelligence/competitive-analyses", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Competitive analyses listing failed: %d", w.Code)
	}

	// Step 4: Create background job
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
	t.Logf("Created competitive analysis job with ID: %s", jobID)

	// Step 5: Check job status
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/competitive-analysis/jobs?id=%s", jobID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Job status check failed: %d", w.Code)
	}

	// Step 6: List jobs
	req = httptest.NewRequest("GET", "/v2/business-intelligence/competitive-analysis/jobs/list", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Jobs listing failed: %d", w.Code)
	}

	t.Logf("Competitive analysis workflow completed successfully")
}

func testCompleteGrowthAnalyticsWorkflow(t *testing.T, mux *http.ServeMux) {
	// Step 1: Create growth analytics request
	requestBody := map[string]interface{}{
		"business_id":     "workflow-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
		"parameters": map[string]interface{}{
			"growth_metrics": []string{"revenue", "market_share", "customer_base"},
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
	req := httptest.NewRequest("POST", "/v2/business-intelligence/growth-analytics", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Growth analytics creation failed: %d, response: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	analysisID := response["id"].(string)
	t.Logf("Created growth analytics with ID: %s", analysisID)

	// Step 2: Retrieve the analysis
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/growth-analytics?id=%s", analysisID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Growth analytics retrieval failed: %d", w.Code)
	}

	// Step 3: List all analyses
	req = httptest.NewRequest("GET", "/v2/business-intelligence/growth-analytics/list", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Growth analytics listing failed: %d", w.Code)
	}

	// Step 4: Create background job
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
	t.Logf("Created growth analytics job with ID: %s", jobID)

	// Step 5: Check job status
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/growth-analytics/jobs?id=%s", jobID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Job status check failed: %d", w.Code)
	}

	// Step 6: List jobs
	req = httptest.NewRequest("GET", "/v2/business-intelligence/growth-analytics/jobs/list", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Jobs listing failed: %d", w.Code)
	}

	t.Logf("Growth analytics workflow completed successfully")
}

func testCompleteBusinessIntelligenceAggregationWorkflow(t *testing.T, mux *http.ServeMux) {
	// Step 1: Create business intelligence aggregation request
	requestBody := map[string]interface{}{
		"business_id":     "workflow-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
		"analysis_types": []string{"market_analysis", "competitive_analysis", "growth_analytics"},
		"parameters": map[string]interface{}{
			"aggregation_level": "comprehensive",
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
	req := httptest.NewRequest("POST", "/v2/business-intelligence/aggregation", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Business intelligence aggregation creation failed: %d, response: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	aggregationID := response["id"].(string)
	t.Logf("Created business intelligence aggregation with ID: %s", aggregationID)

	// Step 2: Retrieve the aggregation
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/aggregation?id=%s", aggregationID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Business intelligence aggregation retrieval failed: %d", w.Code)
	}

	var retrievedResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &retrievedResponse); err != nil {
		t.Fatalf("Failed to unmarshal retrieved response: %v", err)
	}

	// Validate the retrieved aggregation
	if retrievedResponse["id"] != aggregationID {
		t.Errorf("Expected aggregation ID %s, got %s", aggregationID, retrievedResponse["id"])
	}

	// Step 3: List all aggregations
	req = httptest.NewRequest("GET", "/v2/business-intelligence/aggregations", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Business intelligence aggregations listing failed: %d", w.Code)
	}

	var listResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &listResponse); err != nil {
		t.Fatalf("Failed to unmarshal list response: %v", err)
	}

	aggregations, ok := listResponse["aggregations"].([]interface{})
	if !ok || len(aggregations) == 0 {
		t.Fatal("Expected at least one aggregation in the list")
	}

	// Step 4: Create background job
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
	t.Logf("Created business intelligence aggregation job with ID: %s", jobID)

	// Step 5: Check job status
	req = httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/aggregation/jobs?id=%s", jobID), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Job status check failed: %d", w.Code)
	}

	// Step 6: List jobs
	req = httptest.NewRequest("GET", "/v2/business-intelligence/aggregation/jobs/list", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Jobs listing failed: %d", w.Code)
	}

	t.Logf("Business intelligence aggregation workflow completed successfully")
}

func testCrossServiceIntegrationWorkflow(t *testing.T, mux *http.ServeMux) {
	// This test creates analyses from different services and then aggregates them
	// to verify cross-service integration

	// Step 1: Create market analysis
	marketRequest := map[string]interface{}{
		"business_id":     "cross-service-test-business",
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

	body, _ := json.Marshal(marketRequest)
	req := httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Market analysis creation failed: %d", w.Code)
	}

	var marketResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &marketResponse); err != nil {
		t.Fatalf("Failed to unmarshal market response: %v", err)
	}

	marketAnalysisID := marketResponse["id"].(string)
	t.Logf("Created market analysis with ID: %s", marketAnalysisID)

	// Step 2: Create competitive analysis
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
		"parameters": map[string]interface{}{
			"analysis_depth": "comprehensive",
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

	body, _ = json.Marshal(competitiveRequest)
	req = httptest.NewRequest("POST", "/v2/business-intelligence/competitive-analysis", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Competitive analysis creation failed: %d", w.Code)
	}

	var competitiveResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &competitiveResponse); err != nil {
		t.Fatalf("Failed to unmarshal competitive response: %v", err)
	}

	competitiveAnalysisID := competitiveResponse["id"].(string)
	t.Logf("Created competitive analysis with ID: %s", competitiveAnalysisID)

	// Step 3: Create growth analytics
	growthRequest := map[string]interface{}{
		"business_id":     "cross-service-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
		"parameters": map[string]interface{}{
			"growth_metrics": []string{"revenue", "market_share", "customer_base"},
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

	body, _ = json.Marshal(growthRequest)
	req = httptest.NewRequest("POST", "/v2/business-intelligence/growth-analytics", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Growth analytics creation failed: %d", w.Code)
	}

	var growthResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &growthResponse); err != nil {
		t.Fatalf("Failed to unmarshal growth response: %v", err)
	}

	growthAnalysisID := growthResponse["id"].(string)
	t.Logf("Created growth analytics with ID: %s", growthAnalysisID)

	// Step 4: Create aggregation that combines all three
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
		"parameters": map[string]interface{}{
			"aggregation_level": "comprehensive",
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

	aggregationID := aggregationResponse["id"].(string)
	t.Logf("Created business intelligence aggregation with ID: %s", aggregationID)

	// Step 5: Validate that aggregation contains all three analysis types
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

	// Step 6: Validate aggregation insights and recommendations
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

	t.Logf("Cross-service integration workflow completed successfully")
	t.Logf("Created analyses: Market=%s, Competitive=%s, Growth=%s", marketAnalysisID, competitiveAnalysisID, growthAnalysisID)
	t.Logf("Created aggregation: %s", aggregationID)
}

// TestBusinessIntelligenceDataAccuracy tests data accuracy across all business intelligence services
func TestBusinessIntelligenceDataAccuracy(t *testing.T) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	t.Run("Market Analysis Data Accuracy", func(t *testing.T) {
		testMarketAnalysisDataAccuracy(t, mux)
	})

	t.Run("Competitive Analysis Data Accuracy", func(t *testing.T) {
		testCompetitiveAnalysisDataAccuracy(t, mux)
	})

	t.Run("Growth Analytics Data Accuracy", func(t *testing.T) {
		testGrowthAnalyticsDataAccuracy(t, mux)
	})

	t.Run("Aggregation Data Accuracy", func(t *testing.T) {
		testAggregationDataAccuracy(t, mux)
	})
}

func testMarketAnalysisDataAccuracy(t *testing.T, mux *http.ServeMux) {
	// Test with known data to validate accuracy
	requestBody := map[string]interface{}{
		"business_id":     "accuracy-test-business",
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

	// Validate response structure
	requiredFields := []string{"id", "business_id", "industry", "geographic_area", "status", "created_at"}
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Required field '%s' missing from response", field)
		}
	}

	// Validate data types
	if _, ok := response["id"].(string); !ok {
		t.Error("Expected 'id' to be a string")
	}

	if _, ok := response["business_id"].(string); !ok {
		t.Error("Expected 'business_id' to be a string")
	}

	if _, ok := response["industry"].(string); !ok {
		t.Error("Expected 'industry' to be a string")
	}

	// Validate business_id matches input
	if response["business_id"] != "accuracy-test-business" {
		t.Errorf("Expected business_id 'accuracy-test-business', got '%s'", response["business_id"])
	}

	// Validate industry matches input
	if response["industry"] != "Technology" {
		t.Errorf("Expected industry 'Technology', got '%s'", response["industry"])
	}

	t.Logf("Market analysis data accuracy validation passed")
}

func testCompetitiveAnalysisDataAccuracy(t *testing.T, mux *http.ServeMux) {
	// Test with known data to validate accuracy
	requestBody := map[string]interface{}{
		"business_id":     "accuracy-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"competitors":     []string{"Competitor A", "Competitor B", "Competitor C"},
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
		"parameters": map[string]interface{}{
			"analysis_depth": "comprehensive",
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

	// Validate response structure
	requiredFields := []string{"id", "business_id", "industry", "geographic_area", "competitors", "status", "created_at"}
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Required field '%s' missing from response", field)
		}
	}

	// Validate competitors array
	competitors, ok := response["competitors"].([]interface{})
	if !ok {
		t.Error("Expected 'competitors' to be an array")
	}

	if len(competitors) != 3 {
		t.Errorf("Expected 3 competitors, got %d", len(competitors))
	}

	// Validate competitor names
	expectedCompetitors := []string{"Competitor A", "Competitor B", "Competitor C"}
	for i, expected := range expectedCompetitors {
		if i < len(competitors) && competitors[i] != expected {
			t.Errorf("Expected competitor '%s', got '%s'", expected, competitors[i])
		}
	}

	t.Logf("Competitive analysis data accuracy validation passed")
}

func testGrowthAnalyticsDataAccuracy(t *testing.T, mux *http.ServeMux) {
	// Test with known data to validate accuracy
	requestBody := map[string]interface{}{
		"business_id":     "accuracy-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
		"parameters": map[string]interface{}{
			"growth_metrics": []string{"revenue", "market_share", "customer_base"},
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

	// Validate response structure
	requiredFields := []string{"id", "business_id", "industry", "geographic_area", "status", "created_at"}
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Required field '%s' missing from response", field)
		}
	}

	// Validate growth metrics
	parameters, ok := response["parameters"].(map[string]interface{})
	if !ok {
		t.Error("Expected 'parameters' to be an object")
	}

	growthMetrics, ok := parameters["growth_metrics"].([]interface{})
	if !ok {
		t.Error("Expected 'growth_metrics' to be an array")
	}

	if len(growthMetrics) != 3 {
		t.Errorf("Expected 3 growth metrics, got %d", len(growthMetrics))
	}

	// Validate growth metric names
	expectedMetrics := []string{"revenue", "market_share", "customer_base"}
	for i, expected := range expectedMetrics {
		if i < len(growthMetrics) && growthMetrics[i] != expected {
			t.Errorf("Expected growth metric '%s', got '%s'", expected, growthMetrics[i])
		}
	}

	t.Logf("Growth analytics data accuracy validation passed")
}

func testAggregationDataAccuracy(t *testing.T, mux *http.ServeMux) {
	// Test with known data to validate accuracy
	requestBody := map[string]interface{}{
		"business_id":     "accuracy-test-business",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
		"analysis_types": []string{"market_analysis", "competitive_analysis", "growth_analytics"},
		"parameters": map[string]interface{}{
			"aggregation_level": "comprehensive",
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

	// Validate response structure
	requiredFields := []string{"id", "business_id", "industry", "geographic_area", "analysis_types", "status", "created_at"}
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Required field '%s' missing from response", field)
		}
	}

	// Validate analysis types
	analysisTypes, ok := response["analysis_types"].([]interface{})
	if !ok {
		t.Error("Expected 'analysis_types' to be an array")
	}

	if len(analysisTypes) != 3 {
		t.Errorf("Expected 3 analysis types, got %d", len(analysisTypes))
	}

	// Validate analysis type names
	expectedTypes := []string{"market_analysis", "competitive_analysis", "growth_analytics"}
	for i, expected := range expectedTypes {
		if i < len(analysisTypes) && analysisTypes[i] != expected {
			t.Errorf("Expected analysis type '%s', got '%s'", expected, analysisTypes[i])
		}
	}

	t.Logf("Aggregation data accuracy validation passed")
}
