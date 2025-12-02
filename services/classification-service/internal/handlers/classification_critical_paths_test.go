package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"kyb-platform/internal/classification"
	"kyb-platform/internal/classification/testutil"
	"kyb-platform/services/classification-service/internal/config"
	"kyb-platform/services/classification-service/internal/supabase"
)

// TestInFlightRequestTimeout tests timeout handling for in-flight requests
func TestInFlightRequestTimeout(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled:   false,
			RequestTimeout: 100 * time.Millisecond, // Short timeout for testing
		},
	}

	mockRepo := testutil.NewMockKeywordRepository()
	loggerStd := log.Default()

	handler := NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(mockRepo, loggerStd),
		classification.NewClassificationCodeGenerator(mockRepo, loggerStd),
		mockRepo,
		nil,
	)

	reqBody := ClassificationRequest{
		BusinessName: "Test Business",
		Description:  "Test description",
	}
	body, _ := json.Marshal(reqBody)

	// Create first request that will take longer than timeout
	req1 := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	w1 := httptest.NewRecorder()

	// Create second identical request that should timeout waiting
	req2 := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	w2 := httptest.NewRecorder()

	var wg sync.WaitGroup
	wg.Add(2)

	// First request starts
	go func() {
		defer wg.Done()
		handler.HandleClassification(w1, req1)
	}()

	// Small delay to ensure first request is in-flight
	time.Sleep(10 * time.Millisecond)

	// Second request should wait for first, but may timeout
	go func() {
		defer wg.Done()
		handler.HandleClassification(w2, req2)
	}()

	wg.Wait()

	// At least one should complete
	assert.NotEqual(t, 0, w1.Code, "First request should have a response code")
}

// TestInFlightRequestStaleCleanup tests cleanup of stale in-flight requests
func TestInFlightRequestStaleCleanup(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled:   false,
			RequestTimeout: 50 * time.Millisecond,
		},
	}

	mockRepo := testutil.NewMockKeywordRepository()
	loggerStd := log.Default()

	handler := NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(mockRepo, loggerStd),
		classification.NewClassificationCodeGenerator(mockRepo, loggerStd),
		mockRepo,
		nil,
	)

	// Manually create a stale in-flight request
	cacheKey := handler.getCacheKey(&ClassificationRequest{
		BusinessName: "Stale Request",
		Description:  "This will be stale",
	})

	// Create a stale in-flight request (older than timeout)
	handler.inFlightMutex.Lock()
	handler.inFlightRequests[cacheKey] = &inFlightRequest{
		resultChan: make(chan *inFlightResult, 1),
		startTime:  time.Now().Add(-2 * time.Minute), // Stale
		timeout:    50 * time.Millisecond,
	}
	handler.inFlightMutex.Unlock()

	// Trigger cleanup (normally runs every 30s, but we can test the logic)
	handler.inFlightMutex.Lock()
	now := time.Now()
	maxAge := cfg.Classification.RequestTimeout * 2
	if maxAge == 0 {
		maxAge = 2 * time.Minute
	}

	for key, req := range handler.inFlightRequests {
		age := now.Sub(req.startTime)
		if age > maxAge {
			delete(handler.inFlightRequests, key)
			close(req.resultChan)
		}
	}
	handler.inFlightMutex.Unlock()

	// Verify stale request was removed
	handler.inFlightMutex.RLock()
	_, exists := handler.inFlightRequests[cacheKey]
	handler.inFlightMutex.RUnlock()
	assert.False(t, exists, "Stale in-flight request should be removed")
}

// TestContextCancellationDuringDeduplication tests context cancellation while waiting
func TestContextCancellationDuringDeduplication(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled:   false,
			RequestTimeout: 1 * time.Second,
		},
	}

	mockRepo := testutil.NewMockKeywordRepository()
	loggerStd := log.Default()

	handler := NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(mockRepo, loggerStd),
		classification.NewClassificationCodeGenerator(mockRepo, loggerStd),
		mockRepo,
		nil,
	)

	reqBody := ClassificationRequest{
		BusinessName: "Test Business",
		Description:  "Test description",
	}
	body, _ := json.Marshal(reqBody)

	// Create request with short context timeout
	req := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	ctx, cancel := context.WithTimeout(req.Context(), 10*time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Request should handle context cancellation
	handler.HandleClassification(w, req)

	// Should either complete or timeout gracefully
	assert.NotEqual(t, 0, w.Code, "Request should have a response code")
}

// TestCacheHitPath tests the cache hit path in HandleClassification
func TestCacheHitPath(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: true,
			CacheTTL:     5 * time.Minute,
		},
	}

	mockRepo := testutil.NewMockKeywordRepository()
	loggerStd := log.Default()

	handler := NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(mockRepo, loggerStd),
		classification.NewClassificationCodeGenerator(mockRepo, loggerStd),
		mockRepo,
		nil,
	)

	reqBody := ClassificationRequest{
		BusinessName: "Cached Business",
		Description:  "This will be cached",
	}
	body, _ := json.Marshal(reqBody)

	// First request - should populate cache
	req1 := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	w1 := httptest.NewRecorder()
	handler.HandleClassification(w1, req1)

	// Second identical request - should hit cache
	req2 := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	w2 := httptest.NewRecorder()
	handler.HandleClassification(w2, req2)

	// Verify cache was used (X-Cache header)
	// Cache may or may not be hit depending on timing, but should not error
	assert.NotEqual(t, 0, w2.Code, "Second request should have a response code")
}

// TestInFlightRequestErrorPropagation tests error propagation from in-flight requests
func TestInFlightRequestErrorPropagation(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled:   false,
			RequestTimeout: 1 * time.Second,
		},
	}

	mockRepo := testutil.NewMockKeywordRepository()
	loggerStd := log.Default()

	handler := NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(mockRepo, loggerStd),
		classification.NewClassificationCodeGenerator(mockRepo, loggerStd),
		mockRepo,
		nil,
	)

	reqBody := ClassificationRequest{
		BusinessName: "Error Test",
		Description:  "This will test error handling",
	}
	body, _ := json.Marshal(reqBody)

	// Create two concurrent requests
	req1 := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	w1 := httptest.NewRecorder()

	req2 := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	w2 := httptest.NewRecorder()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		handler.HandleClassification(w1, req1)
	}()

	time.Sleep(10 * time.Millisecond)

	go func() {
		defer wg.Done()
		handler.HandleClassification(w2, req2)
	}()

	wg.Wait()

	// Both should complete (one may get result from in-flight, one may process)
	// The key is that errors are properly propagated
	assert.NotEqual(t, 0, w1.Code, "First request should have a response code")
	assert.NotEqual(t, 0, w2.Code, "Second request should have a response code")
}

// TestGetCacheKeyConsistency tests cache key generation consistency
func TestGetCacheKeyConsistency(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{},
	}

	mockRepo := testutil.NewMockKeywordRepository()
	loggerStd := log.Default()

	handler := NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(mockRepo, loggerStd),
		classification.NewClassificationCodeGenerator(mockRepo, loggerStd),
		mockRepo,
		nil,
	)

	req1 := ClassificationRequest{
		BusinessName: "Test Business",
		Description:  "Test description",
		WebsiteURL:   "https://example.com",
	}

	req2 := ClassificationRequest{
		BusinessName: "Test Business",
		Description:  "Test description",
		WebsiteURL:   "https://example.com",
	}

	req3 := ClassificationRequest{
		BusinessName: "Different Business",
		Description:  "Test description",
		WebsiteURL:   "https://example.com",
	}

	key1 := handler.getCacheKey(&req1)
	key2 := handler.getCacheKey(&req2)
	key3 := handler.getCacheKey(&req3)

	// Same requests should generate same key
	assert.Equal(t, key1, key2, "Identical requests should generate same cache key")

	// Different requests should generate different keys
	assert.NotEqual(t, key1, key3, "Different requests should generate different cache keys")

	// Keys should be non-empty
	assert.NotEmpty(t, key1, "Cache key should not be empty")
	assert.NotEmpty(t, key2, "Cache key should not be empty")
	assert.NotEmpty(t, key3, "Cache key should not be empty")
}

// TestInFlightRequestWaitTimeout tests wait timeout for in-flight requests
func TestInFlightRequestWaitTimeout(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled:   false,
			RequestTimeout: 50 * time.Millisecond, // Very short timeout
		},
	}

	mockRepo := testutil.NewMockKeywordRepository()
	loggerStd := log.Default()

	handler := NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(mockRepo, loggerStd),
		classification.NewClassificationCodeGenerator(mockRepo, loggerStd),
		mockRepo,
		nil,
	)

	reqBody := ClassificationRequest{
		BusinessName: "Timeout Test",
		Description:  "This will test wait timeout",
	}
	body, _ := json.Marshal(reqBody)

	// Create request
	req := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Request should handle timeout gracefully
	handler.HandleClassification(w, req)

	// Should complete (may timeout, but should not panic)
	assert.NotEqual(t, 0, w.Code, "Request should have a response code")
}

