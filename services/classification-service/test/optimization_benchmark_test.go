package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"kyb-platform/internal/classification"
	"kyb-platform/services/classification-service/internal/config"
	"kyb-platform/services/classification-service/internal/handlers"
	"kyb-platform/services/classification-service/internal/supabase"
)

// BenchmarkOptimizations benchmarks all Phase 1 and Phase 2 optimizations
func BenchmarkOptimizations(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping optimization benchmarks in short mode")
	}

	// Setup logger
	logger := zaptest.NewLogger(b, zaptest.Level(zap.InfoLevel))

	// Setup config
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled:          true,
			CacheTTL:              5 * time.Minute,
			RequestTimeout:        30 * time.Second,
			OverallTimeout:        60 * time.Second,
			RedisEnabled:          false, // Disable Redis for benchmarks
			EnsembleEnabled:       true,
			MultiPageAnalysisEnabled: true,
		},
	}

	// Create mock Supabase client (or use real one if available)
	supabaseClient := &supabase.Client{} // Mock client

	// Create industry detector and code generator
	industryDetector := classification.NewIndustryDetectionService(nil, logger.Sugar())
	codeGenerator := classification.NewClassificationCodeGenerator(nil, logger.Sugar())

	// Create handler
	handler := handlers.NewClassificationHandler(
		supabaseClient,
		logger,
		cfg,
		industryDetector,
		codeGenerator,
		nil, // No Python ML service for benchmarks
	)

	// Test cases
	testCases := []struct {
		name         string
		businessName string
		description  string
		websiteURL   string
	}{
		{
			name:         "Simple_Business_Name_Only",
			businessName: "TechCorp Solutions",
			description:  "",
			websiteURL:   "",
		},
		{
			name:         "Medium_Business_With_Description",
			businessName: "TechCorp Solutions",
			description:  "Technology consulting and software development services",
			websiteURL:   "",
		},
		{
			name:         "Complex_Business_With_Website",
			businessName: "Microsoft Corporation",
			description:  "Software development and cloud computing services",
			websiteURL:   "https://microsoft.com",
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Create request
			reqBody := map[string]string{
				"business_name": tc.businessName,
				"description":    tc.description,
				"website_url":    tc.websiteURL,
			}
			body, _ := json.Marshal(reqBody)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Create HTTP request
				req := httptest.NewRequest("POST", "/classify", nil)
				req.Body = httptest.NewRequest("POST", "/classify", nil).Body
				req.Body = &mockBody{data: body}
				w := httptest.NewRecorder()

				// Execute handler
				start := time.Now()
				handler.HandleClassification(w, req)
				duration := time.Since(start)

				// Log if duration exceeds target
				if duration > 5*time.Second {
					b.Logf("Request took %v (exceeds 5s target)", duration)
				}

				// Check response
				if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
					b.Logf("Unexpected status code: %d", w.Code)
				}
			}
		})
	}
}

// mockBody implements io.ReadCloser for testing
type mockBody struct {
	data []byte
	pos  int
}

func (m *mockBody) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.data) {
		return 0, nil
	}
	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

func (m *mockBody) Close() error {
	return nil
}

// BenchmarkRequestDeduplication benchmarks request deduplication optimization
func BenchmarkRequestDeduplication(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping deduplication benchmark in short mode")
	}

	logger := zaptest.NewLogger(b, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: true,
			CacheTTL:     5 * time.Minute,
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

	reqBody := map[string]string{
		"business_name": "Test Business",
		"description":    "Test description",
	}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	b.ReportAllocs()

	// Simulate concurrent duplicate requests
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/classify", nil)
			req.Body = &mockBody{data: body}
			w := httptest.NewRecorder()

			handler.HandleClassification(w, req)
		}
	})
}

// BenchmarkCachePerformance benchmarks cache performance (in-memory and Redis)
func BenchmarkCachePerformance(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping cache benchmark in short mode")
	}

	logger := zaptest.NewLogger(b, zaptest.Level(zap.InfoLevel))
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

	reqBody := map[string]string{
		"business_name": "Cached Business",
		"description":    "This should be cached",
	}
	body, _ := json.Marshal(reqBody)

	// First request (cache miss)
	req1 := httptest.NewRequest("POST", "/classify", nil)
	req1.Body = &mockBody{data: body}
	w1 := httptest.NewRecorder()
	handler.HandleClassification(w1, req1)

	b.ResetTimer()
	b.ReportAllocs()

	// Subsequent requests (cache hits)
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/classify", nil)
		req.Body = &mockBody{data: body}
		w := httptest.NewRecorder()

		start := time.Now()
		handler.HandleClassification(w, req)
		duration := time.Since(start)

		// Cache hits should be very fast (< 10ms)
		if duration > 10*time.Millisecond {
			b.Logf("Cache hit took %v (expected < 10ms)", duration)
		}
	}
}

// BenchmarkParallelProcessing benchmarks parallel processing optimization
func BenchmarkParallelProcessing(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping parallel processing benchmark in short mode")
	}

	// This benchmark tests that parallel processing reduces overall time
	// by running risk assessment and verification status in parallel

	logger := zaptest.NewLogger(b, zaptest.Level(zap.InfoLevel))
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			CacheEnabled: false, // Disable cache to test processing
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

	reqBody := map[string]string{
		"business_name": "Parallel Test Business",
		"description":    "Testing parallel processing optimization",
	}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/classify", nil)
		req.Body = &mockBody{data: body}
		w := httptest.NewRecorder()

		start := time.Now()
		handler.HandleClassification(w, req)
		duration := time.Since(start)

		// Parallel processing should reduce time
		if duration > 10*time.Second {
			b.Logf("Processing took %v (may indicate parallel processing not working)", duration)
		}
	}
}

