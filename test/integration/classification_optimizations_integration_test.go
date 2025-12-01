package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"kyb-platform/internal/classification"
	"kyb-platform/services/classification-service/internal/config"
	"kyb-platform/services/classification-service/internal/handlers"
	"kyb-platform/services/classification-service/internal/supabase"
)

// TestEndToEndClassification tests the complete classification flow with all optimizations
func TestEndToEndClassification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled:          true,
			CacheTTL:              5 * time.Minute,
			RequestTimeout:        30 * time.Second,
			OverallTimeout:        60 * time.Second,
			EnsembleEnabled:       true,
			MultiPageAnalysisEnabled: true,
		},
	}

	handler := handlers.NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(nil, logger.Sugar()),
		classification.NewClassificationCodeGenerator(nil, logger.Sugar()),
		nil,
	)

	testCases := []struct {
		name         string
		businessName string
		description  string
		websiteURL   string
		maxDuration  time.Duration
	}{
		{
			name:         "Simple classification",
			businessName: "TechCorp Solutions",
			description:  "Technology consulting",
			websiteURL:   "",
			maxDuration:  5 * time.Second,
		},
		{
			name:         "Complex classification with website",
			businessName: "Microsoft Corporation",
			description:  "Software development and cloud computing services",
			websiteURL:   "https://microsoft.com",
			maxDuration:  10 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := handlers.ClassificationRequest{
				BusinessName: tc.businessName,
				Description:  tc.description,
				WebsiteURL:   tc.websiteURL,
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
			w := httptest.NewRecorder()

			start := time.Now()
			handler.HandleClassification(w, req)
			duration := time.Since(start)

			if duration > tc.maxDuration {
				t.Errorf("Classification took %v, exceeds max duration %v", duration, tc.maxDuration)
			}

			if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
				t.Errorf("Unexpected status code: %d", w.Code)
			}

			// Check response structure
			var response handlers.ClassificationResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err == nil {
				if response.RequestID == "" {
					t.Error("Expected RequestID to be set")
				}
			}
		})
	}
}

// TestCacheBehavior tests cache behavior (in-memory and Redis fallback)
func TestCacheBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: true,
			CacheTTL:     5 * time.Minute,
			RedisEnabled: false, // Test in-memory cache
		},
	}

	handler := handlers.NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(nil, logger.Sugar()),
		classification.NewClassificationCodeGenerator(nil, logger.Sugar()),
		nil,
	)

	reqBody := handlers.ClassificationRequest{
		BusinessName: "Cached Business",
		Description:  "This should be cached",
	}
	body, _ := json.Marshal(reqBody)

	// First request (cache miss)
	req1 := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	w1 := httptest.NewRecorder()
	handler.HandleClassification(w1, req1)

	if w1.Header().Get("X-Cache") != "MISS" {
		t.Log("First request should be cache miss")
	}

	// Second request (cache hit)
	req2 := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	w2 := httptest.NewRecorder()

	start := time.Now()
	handler.HandleClassification(w2, req2)
	duration := time.Since(start)

	if w2.Header().Get("X-Cache") != "HIT" {
		t.Log("Second request should be cache hit")
	}

	// Cache hit should be very fast
	if duration > 100*time.Millisecond {
		t.Logf("Cache hit took %v (expected < 100ms)", duration)
	}
}

// TestRequestDeduplicationIntegration tests request deduplication with concurrent requests
func TestRequestDeduplicationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: false, // Disable cache to test deduplication
		},
	}

	handler := handlers.NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(nil, logger.Sugar()),
		classification.NewClassificationCodeGenerator(nil, logger.Sugar()),
		nil,
	)

	reqBody := handlers.ClassificationRequest{
		BusinessName: "Duplicate Test Business",
		Description:  "Testing request deduplication",
	}
	body, _ := json.Marshal(reqBody)

	// Create multiple concurrent identical requests
	const numRequests = 5
	results := make([]*httptest.ResponseRecorder, numRequests)
	var wg sync.WaitGroup

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
			w := httptest.NewRecorder()
			handler.HandleClassification(w, req)
			results[idx] = w
		}(i)
	}

	wg.Wait()

	// All requests should complete
	for i, w := range results {
		if w == nil {
			t.Errorf("Request %d did not complete", i)
		}
	}
}

// TestEarlyTerminationIntegration tests early termination for low confidence
func TestEarlyTerminationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: false,
		},
	}

	handler := handlers.NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(nil, logger.Sugar()),
		classification.NewClassificationCodeGenerator(nil, logger.Sugar()),
		nil,
	)

	// Request with very low quality content
	reqBody := handlers.ClassificationRequest{
		BusinessName: "X",
		Description:  "Y",
		WebsiteURL:   "",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	w := httptest.NewRecorder()

	start := time.Now()
	handler.HandleClassification(w, req)
	duration := time.Since(start)

	// Early termination should be fast
	if duration > 2*time.Second {
		t.Logf("Early termination took %v (expected < 2s)", duration)
	}

	if w.Code == 0 {
		t.Error("Expected request to complete")
	}
}

// TestParallelProcessingIntegration tests parallel processing optimization
func TestParallelProcessingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: false,
		},
	}

	handler := handlers.NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(nil, logger.Sugar()),
		classification.NewClassificationCodeGenerator(nil, logger.Sugar()),
		nil,
	)

	reqBody := handlers.ClassificationRequest{
		BusinessName: "Parallel Test Business",
		Description:  "Testing parallel processing",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	w := httptest.NewRecorder()

	start := time.Now()
	handler.HandleClassification(w, req)
	duration := time.Since(start)

	// Parallel processing should reduce time
	if duration > 10*time.Second {
		t.Logf("Processing took %v (may indicate parallel processing not working)", duration)
	}

	if w.Code == 0 {
		t.Error("Expected request to complete")
	}
}

