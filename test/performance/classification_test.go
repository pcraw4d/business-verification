package performance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// BenchmarkClassificationAccuracy tests classification accuracy performance
func BenchmarkClassificationAccuracy(b *testing.B) {
	// Create test service
	service := createTestClassificationService(b)

	// Test data
	testCases := []struct {
		name         string
		businessName string
		description  string
		keywords     string
		expectedCode string
	}{
		{
			name:         "Software Development",
			businessName: "Tech Solutions Inc",
			description:  "Software development and consulting services",
			keywords:     "software, development, consulting, technology",
			expectedCode: "541511",
		},
		{
			name:         "Data Analytics",
			businessName: "Data Analytics Corp",
			description:  "Data analytics and business intelligence services",
			keywords:     "data, analytics, business intelligence, consulting",
			expectedCode: "541618",
		},
		{
			name:         "E-commerce",
			businessName: "Online Retail Store",
			description:  "Online retail store selling electronics",
			keywords:     "retail, e-commerce, electronics, online store",
			expectedCode: "454110",
		},
		{
			name:         "Healthcare",
			businessName: "Medical Practice LLC",
			description:  "Primary care medical practice",
			keywords:     "healthcare, medical, primary care, doctor",
			expectedCode: "621111",
		},
		{
			name:         "Manufacturing",
			businessName: "Industrial Manufacturing Co",
			description:  "Manufacturing of industrial equipment",
			keywords:     "manufacturing, industrial, equipment, factory",
			expectedCode: "333120",
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			request := &classification.ClassificationRequest{
				BusinessName: tc.businessName,
				Description:  tc.description,
				Keywords:     tc.keywords,
			}

			result, err := service.ClassifyBusiness(context.Background(), request)
			if err != nil {
				b.Fatalf("Classification failed: %v", err)
			}

			// Verify accuracy
			if result.PrimaryClassification.IndustryCode != tc.expectedCode {
				b.Errorf("Expected industry code %s, got %s for %s",
					tc.expectedCode, result.PrimaryClassification.IndustryCode, tc.name)
			}

			// Verify confidence score
			if result.PrimaryClassification.ConfidenceScore < 0.7 {
				b.Errorf("Low confidence score %f for %s",
					result.PrimaryClassification.ConfidenceScore, tc.name)
			}
		}
	}
}

// BenchmarkResponseTimeOptimization tests response time optimization
func BenchmarkResponseTimeOptimization(b *testing.B) {
	// Create test service with optimization
	service := createTestClassificationService(b)

	// Test request
	request := &classification.ClassificationRequest{
		BusinessName: "Performance Test Business",
		Description:  "Software development and consulting services",
		Keywords:     "software, development, consulting, technology",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		result, err := service.ClassifyBusiness(context.Background(), request)
		if err != nil {
			b.Fatalf("Classification failed: %v", err)
		}

		processingTime := time.Since(start)

		// Verify response time optimization
		if processingTime > time.Second*5 {
			b.Errorf("Response time too slow: %v", processingTime)
		}

		// Verify result quality
		if !result.Success {
			b.Error("Expected successful classification")
		}

		if result.PrimaryClassification.ConfidenceScore < 0.5 {
			b.Errorf("Low confidence score: %f", result.PrimaryClassification.ConfidenceScore)
		}
	}
}

// BenchmarkResourceUsageOptimization tests resource usage optimization
func BenchmarkResourceUsageOptimization(b *testing.B) {
	// Create test service
	service := createTestClassificationService(b)

	// Test concurrent requests
	concurrency := 10
	requestsPerGoroutine := 100

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create channel for results
		results := make(chan *classification.ClassificationResult, concurrency*requestsPerGoroutine)
		errors := make(chan error, concurrency*requestsPerGoroutine)

		// Start concurrent goroutines
		for j := 0; j < concurrency; j++ {
			go func(goroutineID int) {
				for k := 0; k < requestsPerGoroutine; k++ {
					request := &classification.ClassificationRequest{
						BusinessName: fmt.Sprintf("Concurrent Test Business %d-%d", goroutineID, k),
						Description:  "Software development services",
						Keywords:     "software, development, technology",
					}

					result, err := service.ClassifyBusiness(context.Background(), request)
					if err != nil {
						errors <- err
						return
					}
					results <- result
				}
			}(j)
		}

		// Collect results
		successCount := 0
		errorCount := 0

		for j := 0; j < concurrency*requestsPerGoroutine; j++ {
			select {
			case result := <-results:
				if result.Success {
					successCount++
				}
			case err := <-errors:
				errorCount++
				b.Logf("Error in concurrent test: %v", err)
			case <-time.After(time.Second * 30):
				b.Fatal("Timeout waiting for concurrent results")
			}
		}

		// Verify resource usage optimization
		successRate := float64(successCount) / float64(concurrency*requestsPerGoroutine)
		if successRate < 0.95 {
			b.Errorf("Low success rate: %f (expected >= 0.95)", successRate)
		}

		if errorCount > 0 {
			b.Errorf("Too many errors: %d", errorCount)
		}
	}
}

// BenchmarkScalability tests system scalability
func BenchmarkScalability(b *testing.B) {
	// Create test service
	service := createTestClassificationService(b)

	// Test different load levels
	loadLevels := []int{10, 50, 100, 500, 1000}

	for _, load := range loadLevels {
		b.Run(fmt.Sprintf("Load_%d", load), func(b *testing.B) {
			// Create batch of requests
			var requests []*classification.ClassificationRequest
			for i := 0; i < load; i++ {
				requests = append(requests, &classification.ClassificationRequest{
					BusinessName: fmt.Sprintf("Scalability Test Business %d", i),
					Description:  "Software development services",
					Keywords:     "software, development, technology",
				})
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				start := time.Now()

				// Process batch
				results, err := service.ClassifyBusinessBatch(context.Background(), requests)
				if err != nil {
					b.Fatalf("Batch classification failed: %v", err)
				}

				processingTime := time.Since(start)

				// Verify scalability
				if len(results) != len(requests) {
					b.Errorf("Expected %d results, got %d", len(requests), len(results))
				}

				// Calculate throughput
				throughput := float64(len(requests)) / processingTime.Seconds()
				b.ReportMetric(throughput, "requests/sec")

				// Verify response time scales reasonably
				avgTimePerRequest := processingTime / time.Duration(len(requests))
				if avgTimePerRequest > time.Second*2 {
					b.Errorf("Average time per request too high: %v", avgTimePerRequest)
				}

				// Verify success rate
				successCount := 0
				for _, result := range results {
					if result.Success {
						successCount++
					}
				}

				successRate := float64(successCount) / float64(len(requests))
				if successRate < 0.95 {
					b.Errorf("Low success rate: %f (expected >= 0.95)", successRate)
				}
			}
		})
	}
}

// BenchmarkCachePerformance tests cache performance
func BenchmarkCachePerformance(b *testing.B) {
	// Create test cache manager
	logger := createTestLogger()
	metrics := createTestMetrics()
	cacheManager := classification.NewEnhancedCacheManager(logger, metrics, 10000, time.Hour)

	// Test cache key
	cacheKey := &classification.CacheKey{
		BusinessName: "Cache Performance Test Business",
		BusinessType: "LLC",
		Industry:     "Technology",
	}

	// Test result
	result := &classification.MultiIndustryClassificationResult{
		Success: true,
		Classifications: []classification.IndustryClassification{
			{
				IndustryCode:    "541511",
				IndustryName:    "Custom Computer Programming Services",
				ConfidenceScore: 0.95,
			},
		},
	}

	// Warm up cache
	err := cacheManager.Set(context.Background(), cacheKey, result, nil)
	if err != nil {
		b.Fatalf("Failed to set cache: %v", err)
	}

	b.ResetTimer()

	// Benchmark cache hits
	b.Run("CacheHits", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cachedResult, err := cacheManager.Get(context.Background(), cacheKey)
			if err != nil {
				b.Fatalf("Failed to get cached result: %v", err)
			}

			if cachedResult == nil {
				b.Fatal("Expected cached result")
			}
		}
	})

	// Benchmark cache misses
	b.Run("CacheMisses", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			missKey := &classification.CacheKey{
				BusinessName: fmt.Sprintf("Cache Miss Test Business %d", i),
				BusinessType: "LLC",
				Industry:     "Technology",
			}

			_, err := cacheManager.Get(context.Background(), missKey)
			if err == nil {
				b.Error("Expected cache miss error")
			}
		}
	})

	// Benchmark cache sets
	b.Run("CacheSets", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			setKey := &classification.CacheKey{
				BusinessName: fmt.Sprintf("Cache Set Test Business %d", i),
				BusinessType: "LLC",
				Industry:     "Technology",
			}

			err := cacheManager.Set(context.Background(), setKey, result, nil)
			if err != nil {
				b.Fatalf("Failed to set cache: %v", err)
			}
		}
	})
}

// BenchmarkModelOptimization tests model optimization performance
func BenchmarkModelOptimization(b *testing.B) {
	// Create test model optimizer
	config := &classification.ModelOptimizationConfig{
		QuantizationEnabled:   true,
		QuantizationLevel:     16,
		CacheEnabled:          true,
		PreloadEnabled:        true,
		PerformanceMonitoring: true,
		OptimizationLevel:     "medium",
		MaxCacheSize:          1000,
		CacheTTL:              time.Hour,
	}

	logger := createTestLogger()
	metrics := createTestMetrics()
	optimizer := classification.NewModelOptimizer(config, logger, metrics)

	// Test model data
	modelData := make([]byte, 1024*1024) // 1MB test data

	b.ResetTimer()

	// Benchmark model optimization
	b.Run("ModelOptimization", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			modelName := fmt.Sprintf("benchmark_model_%d", i)
			modelVersion := "v1.0.0"

			result, err := optimizer.OptimizeModel(context.Background(), modelName, modelVersion, modelData)
			if err != nil {
				b.Fatalf("Failed to optimize model: %v", err)
			}

			if result == nil {
				b.Fatal("Expected optimization result")
			}

			// Verify optimization quality
			if result.OptimizationStatus == "failed" {
				b.Error("Model optimization failed")
			}

			if result.SpeedupRatio < 1.0 {
				b.Errorf("No speedup achieved: %f", result.SpeedupRatio)
			}
		}
	})

	// Benchmark optimized model retrieval
	b.Run("OptimizedModelRetrieval", func(b *testing.B) {
		// Pre-optimize a model
		modelName := "retrieval_test_model"
		modelVersion := "v1.0.0"

		_, err := optimizer.OptimizeModel(context.Background(), modelName, modelVersion, modelData)
		if err != nil {
			b.Fatalf("Failed to pre-optimize model: %v", err)
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			cachedModel, err := optimizer.GetOptimizedModel(context.Background(), modelName, modelVersion)
			if err != nil {
				b.Fatalf("Failed to get optimized model: %v", err)
			}

			if cachedModel == nil {
				b.Fatal("Expected cached model")
			}
		}
	})
}

// BenchmarkMLClassification tests ML classification performance
func BenchmarkMLClassification(b *testing.B) {
	// Create test ML classifier
	logger := createTestLogger()
	metrics := createTestMetrics()
	modelManager := createTestModelManager(b)
	modelOptimizer := createTestModelOptimizer(b)
	mlClassifier := classification.NewMLClassifier(logger, metrics, modelManager, modelOptimizer)

	// Test request
	request := &classification.MLClassificationRequest{
		BusinessName:        "ML Performance Test Business",
		BusinessDescription: "Software development and consulting services",
		Keywords:            []string{"software", "development", "consulting"},
		WebsiteContent:      "We provide software development services",
		IndustryHints:       []string{"technology", "software"},
		GeographicRegion:    "California",
		BusinessType:        "Corporation",
	}

	b.ResetTimer()

	// Benchmark ML classification
	b.Run("MLClassification", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := mlClassifier.Classify(context.Background(), request)
			if err != nil {
				b.Fatalf("ML classification failed: %v", err)
			}

			if result == nil {
				b.Fatal("Expected ML classification result")
			}

			// Verify result quality
			if result.IndustryCode == "" {
				b.Error("Expected industry code")
			}

			if result.ConfidenceScore < 0.5 {
				b.Errorf("Low confidence score: %f", result.ConfidenceScore)
			}
		}
	})

	// Benchmark optimized ML classification
	b.Run("OptimizedMLClassification", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := mlClassifier.ClassifyWithOptimization(context.Background(), request)
			if err != nil {
				b.Fatalf("Optimized ML classification failed: %v", err)
			}

			if result == nil {
				b.Fatal("Expected optimized ML classification result")
			}

			// Verify optimization benefits
			if result.InferenceTime > time.Second*2 {
				b.Errorf("Inference time too slow: %v", result.InferenceTime)
			}
		}
	})
}

// Helper functions

func createTestClassificationService(b *testing.B) *classification.ClassificationService {
	cfg := &config.ExternalServicesConfig{
		BusinessDataAPI: config.BusinessDataAPIConfig{
			Enabled: true,
			BaseURL: "https://api.example.com",
			APIKey:  "test-key",
			Timeout: 30 * time.Second,
		},
	}

	logger := createTestLogger()
	metrics := createTestMetrics()

	return classification.NewClassificationService(cfg, nil, logger, metrics)
}

func createTestModelManager(b *testing.B) *classification.ModelManager {
	logger := createTestLogger()
	metrics := createTestMetrics()
	return classification.NewModelManager(logger, metrics)
}

func createTestModelOptimizer(b *testing.B) *classification.ModelOptimizer {
	config := &classification.ModelOptimizationConfig{
		QuantizationEnabled:   true,
		QuantizationLevel:     16,
		CacheEnabled:          true,
		PreloadEnabled:        true,
		PerformanceMonitoring: true,
		OptimizationLevel:     "medium",
		MaxCacheSize:          1000,
		CacheTTL:              time.Hour,
	}

	logger := createTestLogger()
	metrics := createTestMetrics()
	return classification.NewModelOptimizer(config, logger, metrics)
}

func createTestLogger() *observability.Logger {
	return observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})
}

func createTestMetrics() *observability.Metrics {
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{
		MetricsEnabled: true,
	})
	return metrics
}
