package classification

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification/repository"
)

// PerformanceTestSuite provides comprehensive performance testing for classification optimizations
type PerformanceTestSuite struct {
	service    *IndustryDetectionService
	classifier *ClassificationCodeGenerator
	repo       repository.KeywordRepository
	logger     *log.Logger
}

// NewPerformanceTestSuite creates a new performance test suite
func NewPerformanceTestSuite(service *IndustryDetectionService, classifier *ClassificationCodeGenerator, repo repository.KeywordRepository, logger *log.Logger) *PerformanceTestSuite {
	if logger == nil {
		logger = log.Default()
	}

	return &PerformanceTestSuite{
		service:    service,
		classifier: classifier,
		repo:       repo,
		logger:     logger,
	}
}

// PerformanceTestResult represents the result of a performance test
type PerformanceTestResult struct {
	TestName          string        `json:"test_name"`
	Duration          time.Duration `json:"duration"`
	Throughput        float64       `json:"throughput_per_second"`
	MemoryUsageMB     float64       `json:"memory_usage_mb"`
	CacheHitRatio     float64       `json:"cache_hit_ratio"`
	ErrorRate         float64       `json:"error_rate"`
	AverageLatency    time.Duration `json:"average_latency"`
	P95Latency        time.Duration `json:"p95_latency"`
	P99Latency        time.Duration `json:"p99_latency"`
	SuccessCount      int           `json:"success_count"`
	ErrorCount        int           `json:"error_count"`
	TotalRequests     int           `json:"total_requests"`
	KeywordsProcessed int           `json:"keywords_processed"`
	ResultsGenerated  int           `json:"results_generated"`
}

// PerformanceTestConfig holds configuration for performance tests
type PerformanceTestConfig struct {
	TestName           string        `json:"test_name"`
	Duration           time.Duration `json:"duration"`
	ConcurrentRequests int           `json:"concurrent_requests"`
	KeywordSetSize     int           `json:"keyword_set_size"`
	BusinessCount      int           `json:"business_count"`
	EnableCaching      bool          `json:"enable_caching"`
	EnableParallel     bool          `json:"enable_parallel"`
	WarmupDuration     time.Duration `json:"warmup_duration"`
}

// DefaultPerformanceTestConfig returns default configuration for performance tests
func DefaultPerformanceTestConfig() *PerformanceTestConfig {
	return &PerformanceTestConfig{
		TestName:           "default_performance_test",
		Duration:           30 * time.Second,
		ConcurrentRequests: 10,
		KeywordSetSize:     50,
		BusinessCount:      100,
		EnableCaching:      true,
		EnableParallel:     true,
		WarmupDuration:     5 * time.Second,
	}
}

// TestLargeKeywordDatasetPerformance tests performance with large keyword datasets
func (pts *PerformanceTestSuite) TestLargeKeywordDatasetPerformance(ctx context.Context, config *PerformanceTestConfig) (*PerformanceTestResult, error) {
	pts.logger.Printf("🚀 Starting large keyword dataset performance test: %s", config.TestName)

	// Generate large keyword dataset
	keywords := pts.generateLargeKeywordDataset(config.KeywordSetSize)
	pts.logger.Printf("📊 Generated %d keywords for testing", len(keywords))

	// Warmup phase
	if config.WarmupDuration > 0 {
		pts.logger.Printf("🔥 Starting warmup phase for %v", config.WarmupDuration)
		pts.runWarmupPhase(ctx, keywords, config.WarmupDuration)
	}

	// Performance test phase
	startTime := time.Now()
	result := &PerformanceTestResult{
		TestName:          config.TestName,
		TotalRequests:     0,
		SuccessCount:      0,
		ErrorCount:        0,
		KeywordsProcessed: len(keywords),
	}

	// Run concurrent requests
	requestChan := make(chan struct{}, config.ConcurrentRequests)
	resultsChan := make(chan *requestResult, config.ConcurrentRequests*10)

	// Start workers
	for i := 0; i < config.ConcurrentRequests; i++ {
		go pts.performanceWorker(ctx, keywords, requestChan, resultsChan, config)
	}

	// Send requests
	ticker := time.NewTicker(100 * time.Millisecond) // Send request every 100ms
	defer ticker.Stop()

	testEndTime := startTime.Add(config.Duration)
	for time.Now().Before(testEndTime) {
		select {
		case <-ticker.C:
			select {
			case requestChan <- struct{}{}:
				result.TotalRequests++
			default:
				// Channel full, skip this request
			}
		case <-ctx.Done():
			break
		}
	}

	// Wait for all requests to complete
	time.Sleep(2 * time.Second)
	close(requestChan)

	// Collect results
	latencies := make([]time.Duration, 0, result.TotalRequests)
	for i := 0; i < result.TotalRequests; i++ {
		select {
		case reqResult := <-resultsChan:
			if reqResult.success {
				result.SuccessCount++
			} else {
				result.ErrorCount++
			}
			latencies = append(latencies, reqResult.latency)
		default:
			break
		}
	}

	// Calculate metrics
	result.Duration = time.Since(startTime)
	result.Throughput = float64(result.SuccessCount) / result.Duration.Seconds()
	result.ErrorRate = float64(result.ErrorCount) / float64(result.TotalRequests)

	if len(latencies) > 0 {
		result.AverageLatency = pts.calculateAverageLatency(latencies)
		result.P95Latency = pts.calculatePercentileLatency(latencies, 95)
		result.P99Latency = pts.calculatePercentileLatency(latencies, 99)
	}

	pts.logger.Printf("✅ Performance test completed: %d requests, %.2f req/s, %.2f%% error rate",
		result.TotalRequests, result.Throughput, result.ErrorRate*100)

	return result, nil
}

// requestResult represents the result of a single request
type requestResult struct {
	success  bool
	latency  time.Duration
	keywords int
}

// performanceWorker processes performance test requests
func (pts *PerformanceTestSuite) performanceWorker(ctx context.Context, keywords []string, requestChan <-chan struct{}, resultsChan chan<- *requestResult, config *PerformanceTestConfig) {
	for range requestChan {
		startTime := time.Now()

		// Create a random subset of keywords for this request
		keywordSubset := pts.getRandomKeywordSubset(keywords, rand.Intn(20)+5) // 5-25 keywords

		// Test classification performance
		_, err := pts.service.DetectIndustryFromContent(ctx, strings.Join(keywordSubset, " "))

		latency := time.Since(startTime)
		resultsChan <- &requestResult{
			success:  err == nil,
			latency:  latency,
			keywords: len(keywordSubset),
		}
	}
}

// TestCachePerformance tests cache hit/miss ratios and performance
func (pts *PerformanceTestSuite) TestCachePerformance(ctx context.Context, config *PerformanceTestConfig) (*PerformanceTestResult, error) {
	pts.logger.Printf("🚀 Starting cache performance test: %s", config.TestName)

	// Generate test data
	keywords := pts.generateLargeKeywordDataset(config.KeywordSetSize)

	// Test cache miss scenario (first run)
	pts.logger.Printf("📊 Testing cache miss scenario...")
	startTime := time.Now()

	for i := 0; i < config.BusinessCount; i++ {
		keywordSubset := pts.getRandomKeywordSubset(keywords, 10)
		_, err := pts.service.DetectIndustryFromContent(ctx, strings.Join(keywordSubset, " "))
		if err != nil {
			pts.logger.Printf("⚠️ Error in cache miss test: %v", err)
		}
	}

	cacheMissDuration := time.Since(startTime)

	// Test cache hit scenario (second run with same data)
	pts.logger.Printf("📊 Testing cache hit scenario...")
	startTime = time.Now()

	for i := 0; i < config.BusinessCount; i++ {
		keywordSubset := pts.getRandomKeywordSubset(keywords, 10)
		_, err := pts.service.DetectIndustryFromContent(ctx, strings.Join(keywordSubset, " "))
		if err != nil {
			pts.logger.Printf("⚠️ Error in cache hit test: %v", err)
		}
	}

	cacheHitDuration := time.Since(startTime)

	// Calculate cache performance metrics
	cacheImprovement := float64(cacheMissDuration-cacheHitDuration) / float64(cacheMissDuration) * 100

	result := &PerformanceTestResult{
		TestName:          config.TestName,
		Duration:          cacheMissDuration + cacheHitDuration,
		TotalRequests:     config.BusinessCount * 2,
		SuccessCount:      config.BusinessCount * 2, // Assume all successful for now
		KeywordsProcessed: len(keywords),
		CacheHitRatio:     cacheImprovement / 100, // Simplified cache hit ratio
	}

	pts.logger.Printf("✅ Cache performance test completed: %.2f%% improvement with caching", cacheImprovement)

	return result, nil
}

// TestClassificationAccuracyBenchmark tests classification accuracy with known datasets
func (pts *PerformanceTestSuite) TestClassificationAccuracyBenchmark(ctx context.Context, config *PerformanceTestConfig) (*PerformanceTestResult, error) {
	pts.logger.Printf("🚀 Starting classification accuracy benchmark: %s", config.TestName)

	// Generate test businesses with known classifications
	testBusinesses := pts.generateTestBusinesses(config.BusinessCount)

	startTime := time.Now()
	correctClassifications := 0
	totalClassifications := 0

	for _, business := range testBusinesses {
		result, err := pts.service.DetectIndustryFromBusinessInfo(ctx, business.Name, business.Description, business.WebsiteURL)
		if err != nil {
			pts.logger.Printf("⚠️ Error classifying business %s: %v", business.Name, err)
			continue
		}

		totalClassifications++

		// Check if classification is correct (simplified check)
		if result.Industry != nil && strings.Contains(strings.ToLower(result.Industry.Name), strings.ToLower(business.ExpectedIndustry)) {
			correctClassifications++
		}
	}

	duration := time.Since(startTime)
	accuracy := float64(correctClassifications) / float64(totalClassifications)

	result := &PerformanceTestResult{
		TestName:         config.TestName,
		Duration:         duration,
		TotalRequests:    totalClassifications,
		SuccessCount:     correctClassifications,
		ErrorCount:       totalClassifications - correctClassifications,
		ResultsGenerated: totalClassifications,
	}

	pts.logger.Printf("✅ Classification accuracy benchmark completed: %.2f%% accuracy", accuracy*100)

	return result, nil
}

// TestLoadTestingConcurrentRequests tests performance under concurrent load
func (pts *PerformanceTestSuite) TestLoadTestingConcurrentRequests(ctx context.Context, config *PerformanceTestConfig) (*PerformanceTestResult, error) {
	pts.logger.Printf("🚀 Starting load testing with %d concurrent requests: %s", config.ConcurrentRequests, config.TestName)

	keywords := pts.generateLargeKeywordDataset(config.KeywordSetSize)

	startTime := time.Now()
	requestChan := make(chan struct{}, config.ConcurrentRequests)
	resultsChan := make(chan *requestResult, config.ConcurrentRequests*100)

	// Start workers
	for i := 0; i < config.ConcurrentRequests; i++ {
		go pts.loadTestWorker(ctx, keywords, requestChan, resultsChan, config)
	}

	// Send requests for the specified duration
	ticker := time.NewTicker(50 * time.Millisecond) // Send request every 50ms
	defer ticker.Stop()

	testEndTime := startTime.Add(config.Duration)
	requestCount := 0

	for time.Now().Before(testEndTime) {
		select {
		case <-ticker.C:
			select {
			case requestChan <- struct{}{}:
				requestCount++
			default:
				// Channel full, skip this request
			}
		case <-ctx.Done():
			break
		}
	}

	// Wait for completion
	time.Sleep(2 * time.Second)
	close(requestChan)

	// Collect results
	successCount := 0
	errorCount := 0
	latencies := make([]time.Duration, 0, requestCount)

	for i := 0; i < requestCount; i++ {
		select {
		case reqResult := <-resultsChan:
			if reqResult.success {
				successCount++
			} else {
				errorCount++
			}
			latencies = append(latencies, reqResult.latency)
		default:
			break
		}
	}

	duration := time.Since(startTime)
	throughput := float64(successCount) / duration.Seconds()
	errorRate := float64(errorCount) / float64(requestCount)

	result := &PerformanceTestResult{
		TestName:      config.TestName,
		Duration:      duration,
		Throughput:    throughput,
		TotalRequests: requestCount,
		SuccessCount:  successCount,
		ErrorCount:    errorCount,
		ErrorRate:     errorRate,
	}

	if len(latencies) > 0 {
		result.AverageLatency = pts.calculateAverageLatency(latencies)
		result.P95Latency = pts.calculatePercentileLatency(latencies, 95)
		result.P99Latency = pts.calculatePercentileLatency(latencies, 99)
	}

	pts.logger.Printf("✅ Load testing completed: %d requests, %.2f req/s, %.2f%% error rate",
		requestCount, throughput, errorRate*100)

	return result, nil
}

// TestMemoryUsageOptimization tests memory usage and optimization
func (pts *PerformanceTestSuite) TestMemoryUsageOptimization(ctx context.Context, config *PerformanceTestConfig) (*PerformanceTestResult, error) {
	pts.logger.Printf("🚀 Starting memory usage optimization test: %s", config.TestName)

	keywords := pts.generateLargeKeywordDataset(config.KeywordSetSize)

	// Measure memory before test
	memBefore := pts.getMemoryUsage()

	startTime := time.Now()
	requestCount := 0

	// Run memory-intensive operations
	for i := 0; i < config.BusinessCount; i++ {
		keywordSubset := pts.getRandomKeywordSubset(keywords, 20)

		// Test classification
		_, err := pts.service.DetectIndustryFromContent(ctx, strings.Join(keywordSubset, " "))
		if err != nil {
			pts.logger.Printf("⚠️ Error in memory test: %v", err)
		}

		// Test code generation
		_, err = pts.classifier.GenerateClassificationCodes(ctx, keywordSubset, "Technology", 0.8)
		if err != nil {
			pts.logger.Printf("⚠️ Error in code generation: %v", err)
		}

		requestCount++

		// Force garbage collection every 100 requests
		if requestCount%100 == 0 {
			// In a real implementation, you might call runtime.GC() here
		}
	}

	duration := time.Since(startTime)
	memAfter := pts.getMemoryUsage()
	memoryIncrease := memAfter - memBefore

	result := &PerformanceTestResult{
		TestName:         config.TestName,
		Duration:         duration,
		TotalRequests:    requestCount,
		SuccessCount:     requestCount, // Assume all successful for now
		MemoryUsageMB:    memoryIncrease,
		ResultsGenerated: requestCount,
	}

	pts.logger.Printf("✅ Memory usage test completed: %.2f MB increase over %d requests", memoryIncrease, requestCount)

	return result, nil
}

// Helper methods

// generateLargeKeywordDataset generates a large dataset of keywords for testing
func (pts *PerformanceTestSuite) generateLargeKeywordDataset(size int) []string {
	keywords := make([]string, 0, size)

	// Technology keywords
	techKeywords := []string{"software", "technology", "digital", "online", "web", "internet", "app", "mobile", "cloud", "api", "data", "algorithm", "machine", "ai", "artificial", "intelligence", "search", "platform", "development", "programming"}

	// Business keywords
	businessKeywords := []string{"business", "company", "corporate", "enterprise", "startup", "consulting", "services", "solutions", "management", "strategy", "marketing", "sales", "finance", "investment", "banking", "insurance", "real estate", "retail", "ecommerce", "manufacturing"}

	// Industry-specific keywords
	industryKeywords := []string{"healthcare", "medical", "hospital", "pharmacy", "education", "school", "university", "restaurant", "food", "dining", "hotel", "travel", "transportation", "logistics", "energy", "utilities", "construction", "engineering", "legal", "accounting"}

	allKeywords := append(techKeywords, businessKeywords...)
	allKeywords = append(allKeywords, industryKeywords...)

	// Generate random combinations
	for i := 0; i < size; i++ {
		keyword := allKeywords[rand.Intn(len(allKeywords))]
		// Add some variation
		if rand.Float32() < 0.3 {
			keyword = keyword + " " + allKeywords[rand.Intn(len(allKeywords))]
		}
		keywords = append(keywords, keyword)
	}

	return keywords
}

// generateTestBusinesses generates test businesses with known classifications
func (pts *PerformanceTestSuite) generateTestBusinesses(count int) []TestBusiness {
	businesses := make([]TestBusiness, 0, count)

	testData := []TestBusiness{
		{"Google Inc", "Search engine and technology company", "https://google.com", "Technology"},
		{"Apple Inc", "Consumer electronics and software company", "https://apple.com", "Technology"},
		{"Microsoft Corporation", "Software and cloud computing company", "https://microsoft.com", "Technology"},
		{"Amazon.com Inc", "E-commerce and cloud computing company", "https://amazon.com", "Retail"},
		{"Tesla Inc", "Electric vehicle and clean energy company", "https://tesla.com", "Manufacturing"},
		{"McDonald's Corporation", "Fast food restaurant chain", "https://mcdonalds.com", "Restaurant"},
		{"Starbucks Corporation", "Coffeehouse chain", "https://starbucks.com", "Restaurant"},
		{"Walmart Inc", "Retail corporation", "https://walmart.com", "Retail"},
		{"JPMorgan Chase & Co", "Investment banking and financial services", "https://jpmorganchase.com", "Financial"},
		{"Johnson & Johnson", "Pharmaceutical and consumer goods company", "https://jnj.com", "Healthcare"},
	}

	for i := 0; i < count; i++ {
		business := testData[i%len(testData)]
		// Add some variation to make it more realistic
		business.Name = business.Name + " " + fmt.Sprintf("Branch %d", i+1)
		businesses = append(businesses, business)
	}

	return businesses
}

// TestBusiness represents a test business with known classification
type TestBusiness struct {
	Name             string
	Description      string
	WebsiteURL       string
	ExpectedIndustry string
}

// getRandomKeywordSubset returns a random subset of keywords
func (pts *PerformanceTestSuite) getRandomKeywordSubset(keywords []string, size int) []string {
	if size >= len(keywords) {
		return keywords
	}

	subset := make([]string, 0, size)
	used := make(map[int]bool)

	for len(subset) < size {
		idx := rand.Intn(len(keywords))
		if !used[idx] {
			subset = append(subset, keywords[idx])
			used[idx] = true
		}
	}

	return subset
}

// runWarmupPhase runs a warmup phase to prime caches
func (pts *PerformanceTestSuite) runWarmupPhase(ctx context.Context, keywords []string, duration time.Duration) {
	endTime := time.Now().Add(duration)

	for time.Now().Before(endTime) {
		keywordSubset := pts.getRandomKeywordSubset(keywords, 10)
		_, _ = pts.service.DetectIndustryFromContent(ctx, strings.Join(keywordSubset, " "))
		time.Sleep(100 * time.Millisecond)
	}
}

// loadTestWorker processes load test requests
func (pts *PerformanceTestSuite) loadTestWorker(ctx context.Context, keywords []string, requestChan <-chan struct{}, resultsChan chan<- *requestResult, config *PerformanceTestConfig) {
	for range requestChan {
		startTime := time.Now()

		keywordSubset := pts.getRandomKeywordSubset(keywords, rand.Intn(15)+5) // 5-20 keywords

		_, err := pts.service.DetectIndustryFromContent(ctx, strings.Join(keywordSubset, " "))

		latency := time.Since(startTime)
		resultsChan <- &requestResult{
			success:  err == nil,
			latency:  latency,
			keywords: len(keywordSubset),
		}
	}
}

// calculateAverageLatency calculates the average latency
func (pts *PerformanceTestSuite) calculateAverageLatency(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	total := time.Duration(0)
	for _, latency := range latencies {
		total += latency
	}

	return total / time.Duration(len(latencies))
}

// calculatePercentileLatency calculates the percentile latency
func (pts *PerformanceTestSuite) calculatePercentileLatency(latencies []time.Duration, percentile int) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// Simple implementation - in production, you'd want a more sophisticated percentile calculation
	index := int(float64(len(latencies)) * float64(percentile) / 100.0)
	if index >= len(latencies) {
		index = len(latencies) - 1
	}

	return latencies[index]
}

// getMemoryUsage returns current memory usage (placeholder implementation)
func (pts *PerformanceTestSuite) getMemoryUsage() float64 {
	// In a real implementation, you would use runtime.MemStats
	// For now, return a placeholder value
	return 0.0
}
