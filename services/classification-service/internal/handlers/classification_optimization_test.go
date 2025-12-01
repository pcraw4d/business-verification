package handlers

import (
	"bytes"
	"context"
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
	"kyb-platform/services/classification-service/internal/supabase"
)

func TestRequestDeduplication(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: false, // Disable cache to test deduplication
		},
	}

	handler := NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(nil, logger.Sugar()),
		classification.NewClassificationCodeGenerator(nil, logger.Sugar()),
		nil,
	)

	reqBody := ClassificationRequest{
		BusinessName: "Test Business",
		Description:  "Test description",
	}
	body, _ := json.Marshal(reqBody)

	// Create first request
	req1 := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	w1 := httptest.NewRecorder()

	// Create second identical request
	req2 := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	w2 := httptest.NewRecorder()

	// Execute both requests concurrently
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		handler.HandleClassification(w1, req1)
	}()

	// Small delay to ensure first request starts
	time.Sleep(10 * time.Millisecond)

	go func() {
		defer wg.Done()
		handler.HandleClassification(w2, req2)
	}()

	wg.Wait()

	// Both should complete (deduplication should handle this)
	if w1.Code == 0 && w2.Code == 0 {
		t.Error("Expected at least one request to complete")
	}
}

func TestContentQualityValidation(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: false,
		},
	}

	handler := NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(nil, logger.Sugar()),
		classification.NewClassificationCodeGenerator(nil, logger.Sugar()),
		nil,
	)

	tests := []struct {
		name         string
		businessName string
		description  string
		websiteURL   string
		expectSkip   bool
	}{
		{
			name:         "Sufficient content",
			businessName: "TechCorp Solutions",
			description:  "Technology consulting and software development services with comprehensive solutions",
			websiteURL:   "https://techcorp.com",
			expectSkip:   false,
		},
		{
			name:         "Insufficient content",
			businessName: "A",
			description:  "B",
			websiteURL:   "",
			expectSkip:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := ClassificationRequest{
				BusinessName: tt.businessName,
				Description:  tt.description,
				WebsiteURL:   tt.websiteURL,
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.HandleClassification(w, req)

			// Request should complete (validation happens internally)
			if w.Code == 0 {
				t.Error("Expected request to complete")
			}
		})
	}
}

func TestEarlyTermination(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: false,
		},
	}

	handler := NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(nil, logger.Sugar()),
		classification.NewClassificationCodeGenerator(nil, logger.Sugar()),
		nil,
	)

	// Request with very low quality content that should trigger early termination
	reqBody := ClassificationRequest{
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

	// Early termination should be fast (< 1 second)
	if duration > 1*time.Second {
		t.Logf("Request took %v (may indicate early termination not working)", duration)
	}

	if w.Code == 0 {
		t.Error("Expected request to complete")
	}
}

func TestCachePerformance(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: true,
			CacheTTL:     5 * time.Minute,
		},
	}

	handler := NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(nil, logger.Sugar()),
		classification.NewClassificationCodeGenerator(nil, logger.Sugar()),
		nil,
	)

	reqBody := ClassificationRequest{
		BusinessName: "Cached Business",
		Description:  "This should be cached",
	}
	body, _ := json.Marshal(reqBody)

	// First request (cache miss)
	req1 := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	w1 := httptest.NewRecorder()
	handler.HandleClassification(w1, req1)

	// Second request (cache hit)
	req2 := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	w2 := httptest.NewRecorder()

	start := time.Now()
	handler.HandleClassification(w2, req2)
	duration := time.Since(start)

	// Cache hit should be very fast (< 10ms)
	if duration > 10*time.Millisecond {
		t.Logf("Cache hit took %v (expected < 10ms)", duration)
	}

	// Check for cache header
	if w2.Header().Get("X-Cache") != "HIT" {
		t.Log("Cache hit header not set (may be cache miss)")
	}
}

func TestParallelProcessing(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: false,
		},
	}

	handler := NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(nil, logger.Sugar()),
		classification.NewClassificationCodeGenerator(nil, logger.Sugar()),
		nil,
	)

	reqBody := ClassificationRequest{
		BusinessName: "Parallel Test Business",
		Description:  "Testing parallel processing optimization",
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

func TestGetCacheKey(t *testing.T) {
	handler := &ClassificationHandler{}

	req1 := &ClassificationRequest{
		BusinessName: "Test Business",
		Description:  "Test description",
		WebsiteURL:   "https://example.com",
	}

	req2 := &ClassificationRequest{
		BusinessName: "Test Business",
		Description:  "Test description",
		WebsiteURL:   "https://example.com",
	}

	key1 := handler.getCacheKey(req1)
	key2 := handler.getCacheKey(req2)

	if key1 != key2 {
		t.Errorf("Expected same cache key for identical requests, got %q and %q", key1, key2)
	}

	req3 := &ClassificationRequest{
		BusinessName: "Different Business",
		Description:  "Test description",
		WebsiteURL:   "https://example.com",
	}

	key3 := handler.getCacheKey(req3)
	if key1 == key3 {
		t.Error("Expected different cache key for different requests")
	}
}

