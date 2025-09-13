package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBusinessIntelligenceHandler_CreateMarketAnalysis(t *testing.T) {
	handler := NewBusinessIntelligenceHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedFields []string
	}{
		{
			name: "valid market analysis request",
			requestBody: map[string]interface{}{
				"business_id":     "test-business-123",
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
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"id", "business_id", "industry", "geographic_area", "market_size", "market_trends", "opportunities", "threats", "benchmarks", "insights", "recommendations", "statistics", "created_at", "status"},
		},
		{
			name: "missing business_id",
			requestBody: map[string]interface{}{
				"industry":        "Technology",
				"geographic_area": "North America",
				"time_range": map[string]interface{}{
					"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
					"end_date":   time.Now().Format(time.RFC3339),
					"time_zone":  "UTC",
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing industry",
			requestBody: map[string]interface{}{
				"business_id":     "test-business-123",
				"geographic_area": "North America",
				"time_range": map[string]interface{}{
					"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
					"end_date":   time.Now().Format(time.RFC3339),
					"time_zone":  "UTC",
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid time range",
			requestBody: map[string]interface{}{
				"business_id":     "test-business-123",
				"industry":        "Technology",
				"geographic_area": "North America",
				"time_range": map[string]interface{}{
					"start_date": time.Now().Format(time.RFC3339),
					"end_date":   time.Now().AddDate(0, -1, 0).Format(time.RFC3339), // End before start
					"time_zone":  "UTC",
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
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

			req := httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateMarketAnalysis(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				for _, field := range tt.expectedFields {
					if _, exists := response[field]; !exists {
						t.Errorf("Expected field '%s' not found in response", field)
					}
				}

				// Validate specific field types
				if id, ok := response["id"].(string); !ok || id == "" {
					t.Error("Expected non-empty string 'id' field")
				}
				if status, ok := response["status"].(string); !ok || status != "completed" {
					t.Error("Expected status to be 'completed'")
				}
			}
		})
	}
}

func TestBusinessIntelligenceHandler_GetMarketAnalysis(t *testing.T) {
	handler := NewBusinessIntelligenceHandler()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "valid analysis ID",
			queryParams:    "?id=test-analysis-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing analysis ID",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty analysis ID",
			queryParams:    "?id=",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/v2/business-intelligence/market-analysis"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.GetMarketAnalysis(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				// Validate response structure
				expectedFields := []string{"id", "business_id", "industry", "geographic_area", "market_size", "market_trends", "opportunities", "threats", "benchmarks", "insights", "recommendations", "statistics", "created_at", "status"}
				for _, field := range expectedFields {
					if _, exists := response[field]; !exists {
						t.Errorf("Expected field '%s' not found in response", field)
					}
				}
			}
		})
	}
}

func TestBusinessIntelligenceHandler_ListMarketAnalyses(t *testing.T) {
	handler := NewBusinessIntelligenceHandler()

	req := httptest.NewRequest("GET", "/v2/business-intelligence/market-analyses", nil)
	w := httptest.NewRecorder()

	handler.ListMarketAnalyses(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate response structure
	expectedFields := []string{"analyses", "total", "timestamp"}
	for _, field := range expectedFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Expected field '%s' not found in response", field)
		}
	}

	// Validate analyses array
	analyses, ok := response["analyses"].([]interface{})
	if !ok {
		t.Fatal("Expected 'analyses' to be an array")
	}

	if len(analyses) == 0 {
		t.Error("Expected at least one analysis in the list")
	}

	// Validate first analysis structure
	firstAnalysis, ok := analyses[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected first analysis to be an object")
	}

	analysisFields := []string{"id", "business_id", "industry", "geographic_area", "status"}
	for _, field := range analysisFields {
		if _, exists := firstAnalysis[field]; !exists {
			t.Errorf("Expected field '%s' not found in analysis", field)
		}
	}
}

func TestBusinessIntelligenceHandler_CreateMarketAnalysisJob(t *testing.T) {
	handler := NewBusinessIntelligenceHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "valid job request",
			requestBody: map[string]interface{}{
				"business_id":     "test-business-123",
				"industry":        "Technology",
				"geographic_area": "North America",
				"time_range": map[string]interface{}{
					"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
					"end_date":   time.Now().Format(time.RFC3339),
					"time_zone":  "UTC",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
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

			req := httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis/jobs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateMarketAnalysisJob(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				// Validate job creation response
				expectedFields := []string{"job_id", "status", "created_at"}
				for _, field := range expectedFields {
					if _, exists := response[field]; !exists {
						t.Errorf("Expected field '%s' not found in response", field)
					}
				}

				if status, ok := response["status"].(string); !ok || status != "created" {
					t.Error("Expected status to be 'created'")
				}
			}
		})
	}
}

func TestBusinessIntelligenceHandler_CreateCompetitiveAnalysis(t *testing.T) {
	handler := NewBusinessIntelligenceHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "valid competitive analysis request",
			requestBody: map[string]interface{}{
				"business_id":     "test-business-123",
				"industry":        "Technology",
				"geographic_area": "North America",
				"competitors":     []string{"Competitor A", "Competitor B", "Competitor C"},
				"time_range": map[string]interface{}{
					"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
					"end_date":   time.Now().Format(time.RFC3339),
					"time_zone":  "UTC",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing competitors",
			requestBody: map[string]interface{}{
				"business_id":     "test-business-123",
				"industry":        "Technology",
				"geographic_area": "North America",
				"time_range": map[string]interface{}{
					"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
					"end_date":   time.Now().Format(time.RFC3339),
					"time_zone":  "UTC",
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty competitors list",
			requestBody: map[string]interface{}{
				"business_id":     "test-business-123",
				"industry":        "Technology",
				"geographic_area": "North America",
				"competitors":     []string{},
				"time_range": map[string]interface{}{
					"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
					"end_date":   time.Now().Format(time.RFC3339),
					"time_zone":  "UTC",
				},
			},
			expectedStatus: http.StatusBadRequest,
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

			req := httptest.NewRequest("POST", "/v2/business-intelligence/competitive-analysis", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateCompetitiveAnalysis(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				// Validate response structure
				expectedFields := []string{"id", "business_id", "industry", "geographic_area", "competitors", "market_position", "competitive_gaps", "advantages", "threats", "insights", "recommendations", "statistics", "created_at", "status"}
				for _, field := range expectedFields {
					if _, exists := response[field]; !exists {
						t.Errorf("Expected field '%s' not found in response", field)
					}
				}
			}
		})
	}
}

func TestBusinessIntelligenceHandler_CreateGrowthAnalytics(t *testing.T) {
	handler := NewBusinessIntelligenceHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "valid growth analytics request",
			requestBody: map[string]interface{}{
				"business_id":     "test-business-123",
				"industry":        "Technology",
				"geographic_area": "North America",
				"time_range": map[string]interface{}{
					"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
					"end_date":   time.Now().Format(time.RFC3339),
					"time_zone":  "UTC",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing geographic area",
			requestBody: map[string]interface{}{
				"business_id": "test-business-123",
				"industry":    "Technology",
				"time_range": map[string]interface{}{
					"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
					"end_date":   time.Now().Format(time.RFC3339),
					"time_zone":  "UTC",
				},
			},
			expectedStatus: http.StatusBadRequest,
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

			req := httptest.NewRequest("POST", "/v2/business-intelligence/growth-analytics", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateGrowthAnalytics(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				// Validate response structure
				expectedFields := []string{"id", "business_id", "industry", "geographic_area", "growth_trends", "growth_projections", "growth_drivers", "growth_barriers", "growth_opportunities", "insights", "recommendations", "statistics", "created_at", "status"}
				for _, field := range expectedFields {
					if _, exists := response[field]; !exists {
						t.Errorf("Expected field '%s' not found in response", field)
					}
				}
			}
		})
	}
}

func TestBusinessIntelligenceHandler_CreateBusinessIntelligenceAggregation(t *testing.T) {
	handler := NewBusinessIntelligenceHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "valid aggregation request with all analysis types",
			requestBody: map[string]interface{}{
				"business_id":     "test-business-123",
				"industry":        "Technology",
				"geographic_area": "North America",
				"time_range": map[string]interface{}{
					"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
					"end_date":   time.Now().Format(time.RFC3339),
					"time_zone":  "UTC",
				},
				"analysis_types": []string{"market_analysis", "competitive_analysis", "growth_analytics"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "valid aggregation request with single analysis type",
			requestBody: map[string]interface{}{
				"business_id":     "test-business-123",
				"industry":        "Technology",
				"geographic_area": "North America",
				"time_range": map[string]interface{}{
					"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
					"end_date":   time.Now().Format(time.RFC3339),
					"time_zone":  "UTC",
				},
				"analysis_types": []string{"market_analysis"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing analysis types",
			requestBody: map[string]interface{}{
				"business_id":     "test-business-123",
				"industry":        "Technology",
				"geographic_area": "North America",
				"time_range": map[string]interface{}{
					"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
					"end_date":   time.Now().Format(time.RFC3339),
					"time_zone":  "UTC",
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty analysis types",
			requestBody: map[string]interface{}{
				"business_id":     "test-business-123",
				"industry":        "Technology",
				"geographic_area": "North America",
				"time_range": map[string]interface{}{
					"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
					"end_date":   time.Now().Format(time.RFC3339),
					"time_zone":  "UTC",
				},
				"analysis_types": []string{},
			},
			expectedStatus: http.StatusBadRequest,
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

			req := httptest.NewRequest("POST", "/v2/business-intelligence/aggregation", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateBusinessIntelligenceAggregation(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				// Validate response structure
				expectedFields := []string{"id", "business_id", "industry", "geographic_area", "analysis_types", "analyses", "insights", "recommendations", "summary", "created_at", "status", "processing_time", "confidence_score", "completeness_score"}
				for _, field := range expectedFields {
					if _, exists := response[field]; !exists {
						t.Errorf("Expected field '%s' not found in response", field)
					}
				}

				// Validate analyses object
				analyses, ok := response["analyses"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected 'analyses' to be an object")
				}

				// Check that requested analysis types are present
				analysisTypes, ok := response["analysis_types"].([]interface{})
				if !ok {
					t.Fatal("Expected 'analysis_types' to be an array")
				}

				for _, analysisType := range analysisTypes {
					if typeStr, ok := analysisType.(string); ok {
						if _, exists := analyses[typeStr]; !exists {
							t.Errorf("Expected analysis type '%s' to be present in analyses", typeStr)
						}
					}
				}
			}
		})
	}
}

func TestBusinessIntelligenceHandler_JobOperations(t *testing.T) {
	handler := NewBusinessIntelligenceHandler()

	// First create a job
	jobRequest := map[string]interface{}{
		"business_id":     "test-business-123",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
	}

	body, _ := json.Marshal(jobRequest)
	req := httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis/jobs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateMarketAnalysisJob(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to create job: %d", w.Code)
	}

	var jobResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &jobResponse); err != nil {
		t.Fatalf("Failed to unmarshal job response: %v", err)
	}

	jobID, ok := jobResponse["job_id"].(string)
	if !ok {
		t.Fatal("Expected job_id in response")
	}

	// Test getting the job
	req = httptest.NewRequest("GET", "/v2/business-intelligence/market-analysis/jobs?id="+jobID, nil)
	w = httptest.NewRecorder()

	handler.GetMarketAnalysisJob(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var job map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &job); err != nil {
		t.Fatalf("Failed to unmarshal job: %v", err)
	}

	// Validate job structure
	expectedFields := []string{"id", "type", "status", "progress", "created_at"}
	for _, field := range expectedFields {
		if _, exists := job[field]; !exists {
			t.Errorf("Expected field '%s' not found in job", field)
		}
	}

	// Test listing jobs
	req = httptest.NewRequest("GET", "/v2/business-intelligence/market-analysis/jobs/list", nil)
	w = httptest.NewRecorder()

	handler.ListMarketAnalysisJobs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var jobsResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &jobsResponse); err != nil {
		t.Fatalf("Failed to unmarshal jobs response: %v", err)
	}

	// Validate jobs list structure
	expectedFields = []string{"jobs", "total", "timestamp"}
	for _, field := range expectedFields {
		if _, exists := jobsResponse[field]; !exists {
			t.Errorf("Expected field '%s' not found in jobs response", field)
		}
	}

	jobs, ok := jobsResponse["jobs"].([]interface{})
	if !ok {
		t.Fatal("Expected 'jobs' to be an array")
	}

	if len(jobs) == 0 {
		t.Error("Expected at least one job in the list")
	}
}

func TestBusinessIntelligenceHandler_ErrorHandling(t *testing.T) {
	handler := NewBusinessIntelligenceHandler()

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "GET job with missing ID",
			method:         "GET",
			path:           "/v2/business-intelligence/market-analysis/jobs",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "GET job with non-existent ID",
			method:         "GET",
			path:           "/v2/business-intelligence/market-analysis/jobs?id=non-existent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "POST with malformed JSON",
			method:         "POST",
			path:           "/v2/business-intelligence/market-analysis",
			body:           `{"business_id": "test", "industry": "Technology", "invalid": json}`,
			expectedStatus: http.StatusBadRequest,
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

			switch {
			case strings.HasPrefix(tt.path, "/v2/business-intelligence/market-analysis") && tt.method == "POST":
				handler.CreateMarketAnalysis(w, req)
			case strings.HasPrefix(tt.path, "/v2/business-intelligence/market-analysis/jobs") && tt.method == "GET":
				handler.GetMarketAnalysisJob(w, req)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// Benchmark tests for performance testing
func BenchmarkBusinessIntelligenceHandler_CreateMarketAnalysis(b *testing.B) {
	handler := NewBusinessIntelligenceHandler()
	requestBody := map[string]interface{}{
		"business_id":     "test-business-123",
		"industry":        "Technology",
		"geographic_area": "North America",
		"time_range": map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			"end_date":   time.Now().Format(time.RFC3339),
			"time_zone":  "UTC",
		},
	}

	body, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateMarketAnalysis(w, req)
	}
}

func BenchmarkBusinessIntelligenceHandler_CreateCompetitiveAnalysis(b *testing.B) {
	handler := NewBusinessIntelligenceHandler()
	requestBody := map[string]interface{}{
		"business_id":     "test-business-123",
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/v2/business-intelligence/competitive-analysis", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateCompetitiveAnalysis(w, req)
	}
}

func BenchmarkBusinessIntelligenceHandler_CreateBusinessIntelligenceAggregation(b *testing.B) {
	handler := NewBusinessIntelligenceHandler()
	requestBody := map[string]interface{}{
		"business_id":     "test-business-123",
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/v2/business-intelligence/aggregation", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateBusinessIntelligenceAggregation(w, req)
	}
}
