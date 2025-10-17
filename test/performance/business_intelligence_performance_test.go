package performance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/routes"
)

// BenchmarkMarketAnalysisCreation benchmarks the creation of market analysis
func BenchmarkMarketAnalysisCreation(b *testing.B) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	requestBody := map[string]interface{}{
		"business_id":     "benchmark-test-business",
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

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Market analysis creation failed: %d", w.Code)
			}
		}
	})
}

// BenchmarkCompetitiveAnalysisCreation benchmarks the creation of competitive analysis
func BenchmarkCompetitiveAnalysisCreation(b *testing.B) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	requestBody := map[string]interface{}{
		"business_id":     "benchmark-test-business",
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
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/v2/business-intelligence/competitive-analysis", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Competitive analysis creation failed: %d", w.Code)
			}
		}
	})
}

// BenchmarkGrowthAnalyticsCreation benchmarks the creation of growth analytics
func BenchmarkGrowthAnalyticsCreation(b *testing.B) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	requestBody := map[string]interface{}{
		"business_id":     "benchmark-test-business",
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
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/v2/business-intelligence/growth-analytics", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Growth analytics creation failed: %d", w.Code)
			}
		}
	})
}

// BenchmarkBusinessIntelligenceAggregationCreation benchmarks the creation of business intelligence aggregation
func BenchmarkBusinessIntelligenceAggregationCreation(b *testing.B) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	requestBody := map[string]interface{}{
		"business_id":     "benchmark-test-business",
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
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/v2/business-intelligence/aggregation", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Business intelligence aggregation creation failed: %d", w.Code)
			}
		}
	})
}

// BenchmarkMarketAnalysisRetrieval benchmarks the retrieval of market analysis
func BenchmarkMarketAnalysisRetrieval(b *testing.B) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	// Create a market analysis first
	requestBody := map[string]interface{}{
		"business_id":     "benchmark-test-business",
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
		b.Fatalf("Market analysis creation failed: %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		b.Fatalf("Failed to unmarshal response: %v", err)
	}

	analysisID := response["id"].(string)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/market-analysis?id=%s", analysisID), nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Market analysis retrieval failed: %d", w.Code)
			}
		}
	})
}

// BenchmarkMarketAnalysisListing benchmarks the listing of market analyses
func BenchmarkMarketAnalysisListing(b *testing.B) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	// Create multiple market analyses first
	for i := 0; i < 100; i++ {
		requestBody := map[string]interface{}{
			"business_id":     fmt.Sprintf("benchmark-test-business-%d", i),
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
			b.Fatalf("Market analysis creation failed: %d", w.Code)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/v2/business-intelligence/market-analyses", nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Market analysis listing failed: %d", w.Code)
			}
		}
	})
}

// BenchmarkJobCreation benchmarks the creation of background jobs
func BenchmarkJobCreation(b *testing.B) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	requestBody := map[string]interface{}{
		"business_id":     "benchmark-test-business",
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
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/v2/business-intelligence/market-analysis/jobs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Job creation failed: %d", w.Code)
			}
		}
	})
}

// BenchmarkJobStatusCheck benchmarks the checking of job status
func BenchmarkJobStatusCheck(b *testing.B) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	// Create a job first
	requestBody := map[string]interface{}{
		"business_id":     "benchmark-test-business",
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
		b.Fatalf("Job creation failed: %d", w.Code)
	}

	var jobResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &jobResponse); err != nil {
		b.Fatalf("Failed to unmarshal job response: %v", err)
	}

	jobID := jobResponse["job_id"].(string)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", fmt.Sprintf("/v2/business-intelligence/market-analysis/jobs?id=%s", jobID), nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Job status check failed: %d", w.Code)
			}
		}
	})
}

// TestConcurrentAccess tests concurrent access to business intelligence services
func TestConcurrentAccess(t *testing.T) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	concurrency := 100
	done := make(chan bool, concurrency)
	var wg sync.WaitGroup

	// Test concurrent market analysis creation
	t.Run("Concurrent Market Analysis Creation", func(t *testing.T) {
		start := time.Now()

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
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

		wg.Wait()
		elapsed := time.Since(start)

		t.Logf("Created %d market analyses in %v (%.2f per second)", concurrency, elapsed, float64(concurrency)/elapsed.Seconds())

		// Wait for all goroutines to complete
		for i := 0; i < concurrency; i++ {
			<-done
		}
	})

	// Test concurrent job creation
	t.Run("Concurrent Job Creation", func(t *testing.T) {
		start := time.Now()

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				defer func() { done <- true }()

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
				}
			}(i)
		}

		wg.Wait()
		elapsed := time.Since(start)

		t.Logf("Created %d jobs in %v (%.2f per second)", concurrency, elapsed, float64(concurrency)/elapsed.Seconds())

		// Wait for all goroutines to complete
		for i := 0; i < concurrency; i++ {
			<-done
		}
	})
}

// TestMemoryUsage tests memory usage during high load
func TestMemoryUsage(t *testing.T) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	// Create many analyses to test memory usage
	numAnalyses := 1000
	start := time.Now()

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
			t.Errorf("Market analysis creation failed: %d", w.Code)
		}
	}

	elapsed := time.Since(start)
	t.Logf("Created %d market analyses in %v (%.2f per second)", numAnalyses, elapsed, float64(numAnalyses)/elapsed.Seconds())

	// Test listing all analyses
	start = time.Now()
	req := httptest.NewRequest("GET", "/v2/business-intelligence/market-analyses", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Market analysis listing failed: %d", w.Code)
	}

	elapsed = time.Since(start)
	t.Logf("Listed all market analyses in %v", elapsed)

	// Test job listing
	start = time.Now()
	req = httptest.NewRequest("GET", "/v2/business-intelligence/market-analysis/jobs/list", nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Job listing failed: %d", w.Code)
	}

	elapsed = time.Since(start)
	t.Logf("Listed all jobs in %v", elapsed)
}

// TestResponseTime tests response times for different operations
func TestResponseTime(t *testing.T) {
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandler()
	config := &routes.RouteConfig{
		BusinessIntelligenceHandler: businessIntelligenceHandler,
		EnableEnhancedFeatures:      true,
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, config)

	tests := []struct {
		name    string
		method  string
		path    string
		body    string
		maxTime time.Duration
	}{
		{
			name:    "Market Analysis Creation",
			method:  "POST",
			path:    "/v2/business-intelligence/market-analysis",
			body:    `{"business_id": "response-time-test", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2023-01-01T00:00:00Z", "end_date": "2024-01-01T00:00:00Z"}}`,
			maxTime: 100 * time.Millisecond,
		},
		{
			name:    "Competitive Analysis Creation",
			method:  "POST",
			path:    "/v2/business-intelligence/competitive-analysis",
			body:    `{"business_id": "response-time-test", "industry": "Technology", "geographic_area": "North America", "competitors": ["Competitor A", "Competitor B"], "time_range": {"start_date": "2023-01-01T00:00:00Z", "end_date": "2024-01-01T00:00:00Z"}}`,
			maxTime: 100 * time.Millisecond,
		},
		{
			name:    "Growth Analytics Creation",
			method:  "POST",
			path:    "/v2/business-intelligence/growth-analytics",
			body:    `{"business_id": "response-time-test", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2023-01-01T00:00:00Z", "end_date": "2024-01-01T00:00:00Z"}}`,
			maxTime: 100 * time.Millisecond,
		},
		{
			name:    "Business Intelligence Aggregation Creation",
			method:  "POST",
			path:    "/v2/business-intelligence/aggregation",
			body:    `{"business_id": "response-time-test", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2023-01-01T00:00:00Z", "end_date": "2024-01-01T00:00:00Z"}, "analysis_types": ["market_analysis", "competitive_analysis"]}`,
			maxTime: 150 * time.Millisecond,
		},
		{
			name:    "Market Analysis Listing",
			method:  "GET",
			path:    "/v2/business-intelligence/market-analyses",
			body:    "",
			maxTime: 50 * time.Millisecond,
		},
		{
			name:    "Job Listing",
			method:  "GET",
			path:    "/v2/business-intelligence/market-analysis/jobs/list",
			body:    "",
			maxTime: 50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()

			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			elapsed := time.Since(start)

			if w.Code != http.StatusOK {
				t.Errorf("Request failed: %d", w.Code)
			}

			if elapsed > tt.maxTime {
				t.Errorf("Response time %v exceeded maximum allowed time %v", elapsed, tt.maxTime)
			}

			t.Logf("Response time: %v", elapsed)
		})
	}
}
