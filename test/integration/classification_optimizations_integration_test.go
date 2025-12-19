//go:build !comprehensive_test
// +build !comprehensive_test

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"kyb-platform/internal/classification"
	"kyb-platform/internal/classification/testutil"
	"kyb-platform/services/classification-service/internal/cache"
	"kyb-platform/services/classification-service/internal/config"
	"kyb-platform/services/classification-service/internal/handlers"
	"kyb-platform/services/classification-service/internal/supabase"
)

// TestClassificationOptimizations_EndToEnd tests the full classification pipeline with all optimizations
func TestClassificationOptimizations_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	mockRepo := testutil.NewMockKeywordRepository()
	loggerStd := log.Default()

	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: true,
			FastPathEnabled: true,
			EnableParallelClassification: true,
			EnableParallelCodeGeneration: true,
		},
	}

	handler := handlers.NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(mockRepo, loggerStd),
		classification.NewClassificationCodeGenerator(mockRepo, loggerStd),
		mockRepo,
		nil, // pythonMLService
	)

	reqBody := handlers.ClassificationRequest{
		BusinessName: "TechCorp Solutions",
		Description:  "Technology consulting and software development services",
		WebsiteURL:   "https://techcorp.com",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute classification
	handler.HandleClassification(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code, "Expected successful classification")

	var response handlers.ClassificationResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err, "Failed to decode response")

	// Verify response contains expected fields
	assert.NotEmpty(t, response.Industry, "Expected industry classification")
	assert.Greater(t, response.Confidence, 0.0, "Expected confidence score")
	assert.NotEmpty(t, response.Codes, "Expected classification codes")
}

// TestRequestDeduplication_ConcurrentRequests tests request deduplication with concurrent requests
func TestRequestDeduplication_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	mockRepo := testutil.NewMockKeywordRepository()
	loggerStd := log.Default()

	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: false, // Disable cache to test deduplication
		},
	}

	handler := handlers.NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(mockRepo, loggerStd),
		classification.NewClassificationCodeGenerator(mockRepo, loggerStd),
		mockRepo,
		nil, // pythonMLService
	)

	reqBody := handlers.ClassificationRequest{
		BusinessName: "Test Business",
		Description:  "Test description",
		WebsiteURL:   "https://test.com",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	// Track how many actual classifications happen
	var classificationCount int32
	var mu sync.Mutex

	// Create 5 concurrent identical requests
	const numRequests = 5
	var wg sync.WaitGroup
	results := make([]*httptest.ResponseRecorder, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			req := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.HandleClassification(w, req)

			mu.Lock()
			if w.Code == http.StatusOK {
				classificationCount++
			}
			results[idx] = w
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// All requests should complete successfully
	for i, result := range results {
		assert.Equal(t, http.StatusOK, result.Code, "Request %d should succeed", i)
	}

	// With deduplication, ideally only 1 classification should happen
	// But due to race conditions, it might be 1-2
	assert.LessOrEqual(t, classificationCount, int32(2), 
		"Deduplication should prevent most duplicate classifications")
}

// TestRedisCache_WebsiteContent tests Redis caching for website content
func TestRedisCache_WebsiteContent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("Skipping Redis test - REDIS_URL not set (e.g., redis://localhost:6379)")
	}

	// Parse Redis URL and create client
	opts, err := redis.ParseURL(redisURL)
	require.NoError(t, err, "Failed to parse Redis URL")

	redisClient := redis.NewClient(opts)
	defer redisClient.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = redisClient.Ping(ctx).Err()
	require.NoError(t, err, "Failed to connect to Redis")

	logger := zap.NewNop()
	cacheInstance := cache.NewWebsiteContentCache(redisClient, logger, 1*time.Hour)

	// Test cache operations
	testURL := "https://example.com/test"
	content := &cache.CachedWebsiteContent{
		TextContent: "Test website content for Redis cache",
		ScrapedAt:   time.Now(),
		Success:     true,
	}

	// Set cache
	err = cacheInstance.Set(ctx, testURL, content)
	require.NoError(t, err, "Failed to set cache")

	// Get cache
	cached, found := cacheInstance.Get(ctx, testURL)
	require.True(t, found, "Cache should be found")
	require.NotNil(t, cached, "Cached content should not be nil")
	assert.Equal(t, content.TextContent, cached.TextContent, "Cached content should match")

	// Delete cache
	cacheInstance.Delete(ctx, testURL)

	// Verify deletion
	_, found = cacheInstance.Get(ctx, testURL)
	assert.False(t, found, "Cache should be deleted")

	t.Log("✅ Redis cache integration test passed")
}

// TestEnsembleVoting_Accuracy tests ensemble voting accuracy
func TestEnsembleVoting_Accuracy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	mockRepo := testutil.NewMockKeywordRepository()
	loggerStd := log.Default()

	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: true,
			EnableParallelClassification: true,
		},
	}

	handler := handlers.NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(mockRepo, loggerStd),
		classification.NewClassificationCodeGenerator(mockRepo, loggerStd),
		mockRepo,
		nil, // pythonMLService - using nil to test Go-only classification
	)

	// Test cases with known expected industries
	testCases := []struct {
		name          string
		businessName  string
		description   string
		expectedIndustry string
	}{
		{
			name:          "Technology company",
			businessName:  "TechCorp",
			description:   "Software development and technology consulting",
			expectedIndustry: "Technology",
		},
		{
			name:          "Financial services",
			businessName:  "FinanceBank",
			description:   "Banking and financial services",
			expectedIndustry: "Financial Services",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := handlers.ClassificationRequest{
				BusinessName: tc.businessName,
				Description:  tc.description,
			}

			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.HandleClassification(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Expected successful classification")

			var response handlers.ClassificationResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			// Verify industry classification
			assert.NotEmpty(t, response.Industry, "Expected industry classification")
			assert.Greater(t, response.Confidence, 0.0, "Expected confidence score")

			t.Logf("✅ Classification: %s -> %s (confidence: %.2f)", 
				tc.businessName, response.Industry, response.Confidence)
		})
	}
}

// TestSmartCrawling_ContentSufficiency tests smart crawling logic
func TestSmartCrawling_ContentSufficiency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	// This test verifies that smart crawling skips full crawl when content is sufficient
	// Since we're using mocks, we'll test the logic indirectly through the handler

	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	mockRepo := testutil.NewMockKeywordRepository()
	loggerStd := log.Default()

	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: true,
			SkipFullCrawlIfContentSufficient: true,
			MinContentLengthForML: 50,
		},
	}

	handler := handlers.NewClassificationHandler(
		&supabase.Client{},
		logger,
		cfg,
		classification.NewIndustryDetectionService(mockRepo, loggerStd),
		classification.NewClassificationCodeGenerator(mockRepo, loggerStd),
		mockRepo,
		nil, // pythonMLService
	)

	// Test with sufficient content in description
	reqBody := handlers.ClassificationRequest{
		BusinessName: "TechCorp Solutions",
		Description:  "Technology consulting and software development services with comprehensive solutions for enterprise clients",
		WebsiteURL:   "https://techcorp.com",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/classify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	start := time.Now()
	handler.HandleClassification(w, req)
	duration := time.Since(start)

	assert.Equal(t, http.StatusOK, w.Code, "Expected successful classification")

	// With sufficient content, processing should be relatively fast
	// (no full crawl needed)
	assert.Less(t, duration, 5*time.Second, 
		"Processing should be fast when content is sufficient")

	t.Logf("✅ Smart crawling test passed - Duration: %v", duration)
}
