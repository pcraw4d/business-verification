// Package testing provides comprehensive performance testing for query optimizations
// This module validates the effectiveness of database query optimizations and caching
// to ensure the KYB Platform meets performance requirements.

package testing

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"kyb-platform/internal/cache"
)

// QueryPerformanceTestSuite provides comprehensive testing for query optimizations
type QueryPerformanceTestSuite struct {
	db           *sql.DB
	cacheManager *cache.QueryCacheManager
	executor     *cache.CachedQueryExecutor
	config       *TestConfig
	results      *TestResults
	mu           sync.RWMutex
}

// TestConfig defines configuration for performance testing
type TestConfig struct {
	// Test execution settings
	ConcurrentUsers int           `json:"concurrent_users"`
	TestDuration    time.Duration `json:"test_duration"`
	RequestInterval time.Duration `json:"request_interval"`
	WarmupDuration  time.Duration `json:"warmup_duration"`

	// Performance thresholds
	MaxResponseTime time.Duration `json:"max_response_time"`
	MinHitRate      float64       `json:"min_hit_rate"`
	MaxErrorRate    float64       `json:"max_error_rate"`

	// Test data settings
	TestDataSize      int  `json:"test_data_size"`
	EnableDataCleanup bool `json:"enable_data_cleanup"`

	// Reporting settings
	EnableDetailedLogging bool          `json:"enable_detailed_logging"`
	ReportInterval        time.Duration `json:"report_interval"`
}

// TestResults contains the results of performance testing
type TestResults struct {
	StartTime           time.Time                   `json:"start_time"`
	EndTime             time.Time                   `json:"end_time"`
	Duration            time.Duration               `json:"duration"`
	TotalRequests       int64                       `json:"total_requests"`
	SuccessfulRequests  int64                       `json:"successful_requests"`
	FailedRequests      int64                       `json:"failed_requests"`
	AverageResponseTime time.Duration               `json:"average_response_time"`
	MinResponseTime     time.Duration               `json:"min_response_time"`
	MaxResponseTime     time.Duration               `json:"max_response_time"`
	P95ResponseTime     time.Duration               `json:"p95_response_time"`
	P99ResponseTime     time.Duration               `json:"p99_response_time"`
	CacheHitRate        float64                     `json:"cache_hit_rate"`
	ErrorRate           float64                     `json:"error_rate"`
	Throughput          float64                     `json:"throughput"`
	QueryResults        map[string]*QueryTestResult `json:"query_results"`
	PerformanceMetrics  *QueryPerformanceMetrics    `json:"performance_metrics"`
	Passed              bool                        `json:"passed"`
	Failures            []string                    `json:"failures"`
}

// QueryTestResult contains results for a specific query type
type QueryTestResult struct {
	QueryType           string          `json:"query_type"`
	TotalRequests       int64           `json:"total_requests"`
	SuccessfulRequests  int64           `json:"successful_requests"`
	FailedRequests      int64           `json:"failed_requests"`
	AverageResponseTime time.Duration   `json:"average_response_time"`
	MinResponseTime     time.Duration   `json:"min_response_time"`
	MaxResponseTime     time.Duration   `json:"max_response_time"`
	P95ResponseTime     time.Duration   `json:"p95_response_time"`
	P99ResponseTime     time.Duration   `json:"p99_response_time"`
	CacheHits           int64           `json:"cache_hits"`
	CacheMisses         int64           `json:"cache_misses"`
	CacheHitRate        float64         `json:"cache_hit_rate"`
	ErrorRate           float64         `json:"error_rate"`
	ResponseTimes       []time.Duration `json:"response_times"`
}

// QueryPerformanceMetrics contains system performance metrics for query testing
type QueryPerformanceMetrics struct {
	CPUUsage            float64   `json:"cpu_usage"`
	MemoryUsage         float64   `json:"memory_usage"`
	DatabaseConnections int       `json:"database_connections"`
	CacheSize           int64     `json:"cache_size"`
	CacheMemoryUsage    int64     `json:"cache_memory_usage"`
	Timestamp           time.Time `json:"timestamp"`
}

// NewQueryPerformanceTestSuite creates a new performance test suite
func NewQueryPerformanceTestSuite(db *sql.DB, cacheManager *cache.QueryCacheManager, executor *cache.CachedQueryExecutor, config *TestConfig) *QueryPerformanceTestSuite {
	if config == nil {
		config = getDefaultTestConfig()
	}

	return &QueryPerformanceTestSuite{
		db:           db,
		cacheManager: cacheManager,
		executor:     executor,
		config:       config,
		results: &TestResults{
			QueryResults: make(map[string]*QueryTestResult),
		},
	}
}

// RunComprehensiveTests runs all performance tests
func (qpts *QueryPerformanceTestSuite) RunComprehensiveTests(ctx context.Context) (*TestResults, error) {
	log.Println("Starting comprehensive query performance tests...")

	qpts.results.StartTime = time.Now()

	// Run individual test suites
	testSuites := []struct {
		name string
		test func(context.Context) error
	}{
		{"Time-based Classification Query Test", qpts.runTimeBasedClassificationTest},
		{"Industry-based Classification Query Test", qpts.runIndustryBasedClassificationTest},
		{"Business Classification Lookup Test", qpts.runBusinessClassificationLookupTest},
		{"Risk Assessment Query Test", qpts.runRiskAssessmentQueryTest},
		{"Industry Keyword Lookup Test", qpts.runIndustryKeywordLookupTest},
		{"Complex Join Query Test", qpts.runComplexJoinQueryTest},
		{"JSONB Query Test", qpts.runJSONBQueryTest},
		{"Array Query Test", qpts.runArrayQueryTest},
		{"Cache Performance Test", qpts.runCachePerformanceTest},
		{"Concurrent Load Test", qpts.runConcurrentLoadTest},
	}

	// Execute test suites
	for _, suite := range testSuites {
		log.Printf("Running %s...", suite.name)
		if err := suite.test(ctx); err != nil {
			log.Printf("Test suite %s failed: %v", suite.name, err)
			qpts.results.Failures = append(qpts.results.Failures, fmt.Sprintf("%s: %v", suite.name, err))
		}
	}

	qpts.results.EndTime = time.Now()
	qpts.results.Duration = qpts.results.EndTime.Sub(qpts.results.StartTime)

	// Calculate overall results
	qpts.calculateOverallResults()

	// Validate performance requirements
	qpts.validatePerformanceRequirements()

	log.Printf("Performance tests completed. Passed: %v", qpts.results.Passed)
	return qpts.results, nil
}

// runTimeBasedClassificationTest tests time-based classification queries
func (qpts *QueryPerformanceTestSuite) runTimeBasedClassificationTest(ctx context.Context) error {
	queryType := "time_based_classification"
	result := &QueryTestResult{
		QueryType:     queryType,
		ResponseTimes: make([]time.Duration, 0),
	}

	// Test parameters
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()
	limit := 100

	// Run test iterations
	for i := 0; i < qpts.config.TestDataSize; i++ {
		start := time.Now()

		// Execute query with cache
		cachedQuery := cache.NewTimeBasedClassificationQuery(startTime, endTime, limit)
		_, err := qpts.executor.ExecuteQuery(ctx, cachedQuery)

		duration := time.Since(start)
		result.ResponseTimes = append(result.ResponseTimes, duration)

		if err != nil {
			result.FailedRequests++
			if qpts.config.EnableDetailedLogging {
				log.Printf("Time-based classification query failed: %v", err)
			}
		} else {
			result.SuccessfulRequests++
		}

		result.TotalRequests++

		// Small delay between requests
		time.Sleep(qpts.config.RequestInterval)
	}

	// Calculate statistics
	qpts.calculateQueryStatistics(result)

	// Store result
	qpts.mu.Lock()
	qpts.results.QueryResults[queryType] = result
	qpts.mu.Unlock()

	return nil
}

// runIndustryBasedClassificationTest tests industry-based classification queries
func (qpts *QueryPerformanceTestSuite) runIndustryBasedClassificationTest(ctx context.Context) error {
	queryType := "industry_based_classification"
	result := &QueryTestResult{
		QueryType:     queryType,
		ResponseTimes: make([]time.Duration, 0),
	}

	// Test parameters
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()
	industries := []string{"Technology", "Finance", "Healthcare", "Manufacturing", "Retail"}
	limit := 100

	// Run test iterations
	for i := 0; i < qpts.config.TestDataSize; i++ {
		industry := industries[i%len(industries)]

		start := time.Now()

		// Execute query with cache
		cachedQuery := cache.NewIndustryBasedClassificationQuery(startTime, endTime, industry, limit)
		_, err := qpts.executor.ExecuteQuery(ctx, cachedQuery)

		duration := time.Since(start)
		result.ResponseTimes = append(result.ResponseTimes, duration)

		if err != nil {
			result.FailedRequests++
			if qpts.config.EnableDetailedLogging {
				log.Printf("Industry-based classification query failed: %v", err)
			}
		} else {
			result.SuccessfulRequests++
		}

		result.TotalRequests++

		// Small delay between requests
		time.Sleep(qpts.config.RequestInterval)
	}

	// Calculate statistics
	qpts.calculateQueryStatistics(result)

	// Store result
	qpts.mu.Lock()
	qpts.results.QueryResults[queryType] = result
	qpts.mu.Unlock()

	return nil
}

// runBusinessClassificationLookupTest tests business classification lookups
func (qpts *QueryPerformanceTestSuite) runBusinessClassificationLookupTest(ctx context.Context) error {
	queryType := "business_classification_lookup"
	result := &QueryTestResult{
		QueryType:     queryType,
		ResponseTimes: make([]time.Duration, 0),
	}

	// Get test business IDs
	businessIDs, err := qpts.getTestBusinessIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get test business IDs: %w", err)
	}

	// Run test iterations
	for i := 0; i < qpts.config.TestDataSize; i++ {
		businessID := businessIDs[i%len(businessIDs)]

		start := time.Now()

		// Execute query with cache
		cachedQuery := cache.NewClassificationQuery(businessID, "")
		_, err := qpts.executor.ExecuteQuery(ctx, cachedQuery)

		duration := time.Since(start)
		result.ResponseTimes = append(result.ResponseTimes, duration)

		if err != nil {
			result.FailedRequests++
			if qpts.config.EnableDetailedLogging {
				log.Printf("Business classification lookup failed: %v", err)
			}
		} else {
			result.SuccessfulRequests++
		}

		result.TotalRequests++

		// Small delay between requests
		time.Sleep(qpts.config.RequestInterval)
	}

	// Calculate statistics
	qpts.calculateQueryStatistics(result)

	// Store result
	qpts.mu.Lock()
	qpts.results.QueryResults[queryType] = result
	qpts.mu.Unlock()

	return nil
}

// runRiskAssessmentQueryTest tests risk assessment queries
func (qpts *QueryPerformanceTestSuite) runRiskAssessmentQueryTest(ctx context.Context) error {
	queryType := "risk_assessment_query"
	result := &QueryTestResult{
		QueryType:     queryType,
		ResponseTimes: make([]time.Duration, 0),
	}

	// Get test business IDs
	businessIDs, err := qpts.getTestBusinessIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get test business IDs: %w", err)
	}

	riskLevels := []string{"high", "critical", "medium", "low"}

	// Run test iterations
	for i := 0; i < qpts.config.TestDataSize; i++ {
		businessID := businessIDs[i%len(businessIDs)]
		riskLevel := riskLevels[i%len(riskLevels)]

		start := time.Now()

		// Execute query with cache
		cachedQuery := cache.NewRiskAssessmentQuery(businessID, []string{riskLevel})
		_, err := qpts.executor.ExecuteQuery(ctx, cachedQuery)

		duration := time.Since(start)
		result.ResponseTimes = append(result.ResponseTimes, duration)

		if err != nil {
			result.FailedRequests++
			if qpts.config.EnableDetailedLogging {
				log.Printf("Risk assessment query failed: %v", err)
			}
		} else {
			result.SuccessfulRequests++
		}

		result.TotalRequests++

		// Small delay between requests
		time.Sleep(qpts.config.RequestInterval)
	}

	// Calculate statistics
	qpts.calculateQueryStatistics(result)

	// Store result
	qpts.mu.Lock()
	qpts.results.QueryResults[queryType] = result
	qpts.mu.Unlock()

	return nil
}

// runIndustryKeywordLookupTest tests industry keyword lookups
func (qpts *QueryPerformanceTestSuite) runIndustryKeywordLookupTest(ctx context.Context) error {
	queryType := "industry_keyword_lookup"
	result := &QueryTestResult{
		QueryType:     queryType,
		ResponseTimes: make([]time.Duration, 0),
	}

	// Get test industry IDs
	industryIDs, err := qpts.getTestIndustryIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get test industry IDs: %w", err)
	}

	// Run test iterations
	for i := 0; i < qpts.config.TestDataSize; i++ {
		industryID := industryIDs[i%len(industryIDs)]

		start := time.Now()

		// Execute query with cache
		query := `
			SELECT 
				ik.id,
				ik.industry_id,
				ik.keyword,
				ik.weight,
				ik.category,
				ik.synonyms,
				ik.is_primary
			FROM industry_keywords ik
			WHERE ik.industry_id = $1 AND ik.is_primary = true
			ORDER BY ik.weight DESC
		`
		_, err := qpts.executor.ExecuteQueryWithCache(ctx, query, []interface{}{industryID}, queryType, 30*time.Minute)

		duration := time.Since(start)
		result.ResponseTimes = append(result.ResponseTimes, duration)

		if err != nil {
			result.FailedRequests++
			if qpts.config.EnableDetailedLogging {
				log.Printf("Industry keyword lookup failed: %v", err)
			}
		} else {
			result.SuccessfulRequests++
		}

		result.TotalRequests++

		// Small delay between requests
		time.Sleep(qpts.config.RequestInterval)
	}

	// Calculate statistics
	qpts.calculateQueryStatistics(result)

	// Store result
	qpts.mu.Lock()
	qpts.results.QueryResults[queryType] = result
	qpts.mu.Unlock()

	return nil
}

// runComplexJoinQueryTest tests complex join queries
func (qpts *QueryPerformanceTestSuite) runComplexJoinQueryTest(ctx context.Context) error {
	queryType := "complex_join_query"
	result := &QueryTestResult{
		QueryType:     queryType,
		ResponseTimes: make([]time.Duration, 0),
	}

	// Run test iterations
	for i := 0; i < qpts.config.TestDataSize; i++ {
		start := time.Now()

		// Execute complex join query with cache
		query := `
			SELECT 
				u.email,
				b.name,
				COALESCE(bc.primary_industry->>'name', 'Unknown') as primary_industry,
				COALESCE(ra.risk_level, 'low') as risk_level,
				u.created_at
			FROM users u
			INNER JOIN businesses b ON u.id = b.user_id
			LEFT JOIN business_classifications bc ON b.id = bc.business_id
			LEFT JOIN business_risk_assessments ra ON b.id = ra.business_id
			WHERE u.created_at >= $1
			ORDER BY u.created_at DESC
			LIMIT $2
		`
		startTime := time.Now().Add(-30 * 24 * time.Hour)
		limit := 50

		_, err := qpts.executor.ExecuteQueryWithCache(ctx, query, []interface{}{startTime, limit}, queryType, 10*time.Minute)

		duration := time.Since(start)
		result.ResponseTimes = append(result.ResponseTimes, duration)

		if err != nil {
			result.FailedRequests++
			if qpts.config.EnableDetailedLogging {
				log.Printf("Complex join query failed: %v", err)
			}
		} else {
			result.SuccessfulRequests++
		}

		result.TotalRequests++

		// Small delay between requests
		time.Sleep(qpts.config.RequestInterval)
	}

	// Calculate statistics
	qpts.calculateQueryStatistics(result)

	// Store result
	qpts.mu.Lock()
	qpts.results.QueryResults[queryType] = result
	qpts.mu.Unlock()

	return nil
}

// runJSONBQueryTest tests JSONB queries
func (qpts *QueryPerformanceTestSuite) runJSONBQueryTest(ctx context.Context) error {
	queryType := "jsonb_query"
	result := &QueryTestResult{
		QueryType:     queryType,
		ResponseTimes: make([]time.Duration, 0),
	}

	// Run test iterations
	for i := 0; i < qpts.config.TestDataSize; i++ {
		start := time.Now()

		// Execute JSONB query with cache
		query := `
			SELECT 
				u.id,
				u.email,
				u.name,
				u.metadata->>'role' as role,
				u.metadata->>'status' as status,
				u.created_at
			FROM users u
			WHERE u.metadata->>'role' = $1
			ORDER BY u.created_at DESC
			LIMIT $2
		`
		roles := []string{"admin", "user", "compliance_officer", "risk_manager"}
		role := roles[i%len(roles)]
		limit := 100

		_, err := qpts.executor.ExecuteQueryWithCache(ctx, query, []interface{}{role, limit}, queryType, 30*time.Minute)

		duration := time.Since(start)
		result.ResponseTimes = append(result.ResponseTimes, duration)

		if err != nil {
			result.FailedRequests++
			if qpts.config.EnableDetailedLogging {
				log.Printf("JSONB query failed: %v", err)
			}
		} else {
			result.SuccessfulRequests++
		}

		result.TotalRequests++

		// Small delay between requests
		time.Sleep(qpts.config.RequestInterval)
	}

	// Calculate statistics
	qpts.calculateQueryStatistics(result)

	// Store result
	qpts.mu.Lock()
	qpts.results.QueryResults[queryType] = result
	qpts.mu.Unlock()

	return nil
}

// runArrayQueryTest tests array queries
func (qpts *QueryPerformanceTestSuite) runArrayQueryTest(ctx context.Context) error {
	queryType := "array_query"
	result := &QueryTestResult{
		QueryType:     queryType,
		ResponseTimes: make([]time.Duration, 0),
	}

	// Run test iterations
	for i := 0; i < qpts.config.TestDataSize; i++ {
		start := time.Now()

		// Execute array query with cache
		query := `
			SELECT 
				rk.id,
				rk.keyword,
				rk.risk_category,
				rk.risk_severity,
				rk.mcc_codes,
				rk.description
			FROM risk_keywords rk
			WHERE rk.mcc_codes @> $1
				AND rk.risk_severity = $2
				AND rk.is_active = true
			ORDER BY rk.risk_severity DESC, rk.keyword
		`
		mccCodes := []string{"7995", "5812", "5813", "5814", "5999"}
		mccCode := mccCodes[i%len(mccCodes)]
		riskSeverities := []string{"high", "critical", "medium", "low"}
		riskSeverity := riskSeverities[i%len(riskSeverities)]

		_, err := qpts.executor.ExecuteQueryWithCache(ctx, query, []interface{}{[]string{mccCode}, riskSeverity}, queryType, 1*time.Hour)

		duration := time.Since(start)
		result.ResponseTimes = append(result.ResponseTimes, duration)

		if err != nil {
			result.FailedRequests++
			if qpts.config.EnableDetailedLogging {
				log.Printf("Array query failed: %v", err)
			}
		} else {
			result.SuccessfulRequests++
		}

		result.TotalRequests++

		// Small delay between requests
		time.Sleep(qpts.config.RequestInterval)
	}

	// Calculate statistics
	qpts.calculateQueryStatistics(result)

	// Store result
	qpts.mu.Lock()
	qpts.results.QueryResults[queryType] = result
	qpts.mu.Unlock()

	return nil
}

// runCachePerformanceTest tests cache performance
func (qpts *QueryPerformanceTestSuite) runCachePerformanceTest(ctx context.Context) error {
	log.Println("Running cache performance test...")

	// Test cache hit rate
	initialMetrics := qpts.cacheManager.GetMetrics()

	// Run queries to populate cache
	for i := 0; i < 100; i++ {
		businessID := fmt.Sprintf("test-business-%d", i)
		cachedQuery := cache.NewClassificationQuery(businessID, "")
		qpts.executor.ExecuteQuery(ctx, cachedQuery)
	}

	// Run same queries again to test cache hits
	for i := 0; i < 100; i++ {
		businessID := fmt.Sprintf("test-business-%d", i)
		cachedQuery := cache.NewClassificationQuery(businessID, "")
		qpts.executor.ExecuteQuery(ctx, cachedQuery)
	}

	// Get final metrics
	finalMetrics := qpts.cacheManager.GetMetrics()

	// Calculate cache hit rate improvement
	hitRateImprovement := finalMetrics.HitRate - initialMetrics.HitRate

	log.Printf("Cache hit rate improvement: %.2f%%", hitRateImprovement)

	if hitRateImprovement < 50.0 {
		return fmt.Errorf("cache hit rate improvement too low: %.2f%%", hitRateImprovement)
	}

	return nil
}

// runConcurrentLoadTest tests concurrent query execution
func (qpts *QueryPerformanceTestSuite) runConcurrentLoadTest(ctx context.Context) error {
	log.Printf("Running concurrent load test with %d users...", qpts.config.ConcurrentUsers)

	var wg sync.WaitGroup
	startTime := time.Now()

	// Start concurrent users
	for i := 0; i < qpts.config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			// Run queries for the duration of the test
			for time.Since(startTime) < qpts.config.TestDuration {
				// Execute different types of queries
				switch userID % 8 {
				case 0:
					cachedQuery := cache.NewTimeBasedClassificationQuery(
						time.Now().Add(-24*time.Hour), time.Now(), 100)
					qpts.executor.ExecuteQuery(ctx, cachedQuery)
				case 1:
					cachedQuery := cache.NewIndustryBasedClassificationQuery(
						time.Now().Add(-24*time.Hour), time.Now(), "Technology", 100)
					qpts.executor.ExecuteQuery(ctx, cachedQuery)
				case 2:
					cachedQuery := cache.NewClassificationQuery("test-business-1", "")
					qpts.executor.ExecuteQuery(ctx, cachedQuery)
				case 3:
					cachedQuery := cache.NewRiskAssessmentQuery("test-business-1", []string{"high"})
					qpts.executor.ExecuteQuery(ctx, cachedQuery)
				case 4:
					cachedQuery := cache.NewUserDataQuery("test-user-1")
					qpts.executor.ExecuteQuery(ctx, cachedQuery)
				case 5:
					cachedQuery := cache.NewBusinessDataQuery("test-business-1")
					qpts.executor.ExecuteQuery(ctx, cachedQuery)
				case 6:
					// Complex join query
					query := `
						SELECT u.email, b.name, bc.primary_industry
						FROM users u
						JOIN businesses b ON u.id = b.user_id
						LEFT JOIN business_classifications bc ON b.id = bc.business_id
						WHERE u.created_at >= $1
						LIMIT 50
					`
					qpts.executor.ExecuteQueryWithCache(ctx, query, []interface{}{time.Now().Add(-30 * 24 * time.Hour)}, "complex_join", 10*time.Minute)
				case 7:
					// JSONB query
					query := `
						SELECT u.id, u.email, u.metadata->>'role' as role
						FROM users u
						WHERE u.metadata->>'role' = $1
						LIMIT 100
					`
					qpts.executor.ExecuteQueryWithCache(ctx, query, []interface{}{"admin"}, "jsonb_query", 30*time.Minute)
				}

				// Small delay between requests
				time.Sleep(qpts.config.RequestInterval)
			}
		}(i)
	}

	// Wait for all users to complete
	wg.Wait()

	log.Println("Concurrent load test completed")
	return nil
}

// calculateQueryStatistics calculates statistics for a query test result
func (qpts *QueryPerformanceTestSuite) calculateQueryStatistics(result *QueryTestResult) {
	if len(result.ResponseTimes) == 0 {
		return
	}

	// Calculate basic statistics
	var totalTime time.Duration
	for _, rt := range result.ResponseTimes {
		totalTime += rt
	}

	result.AverageResponseTime = totalTime / time.Duration(len(result.ResponseTimes))

	// Find min and max
	result.MinResponseTime = result.ResponseTimes[0]
	result.MaxResponseTime = result.ResponseTimes[0]
	for _, rt := range result.ResponseTimes {
		if rt < result.MinResponseTime {
			result.MinResponseTime = rt
		}
		if rt > result.MaxResponseTime {
			result.MaxResponseTime = rt
		}
	}

	// Calculate percentiles
	result.P95ResponseTime = qpts.calculatePercentile(result.ResponseTimes, 95)
	result.P99ResponseTime = qpts.calculatePercentile(result.ResponseTimes, 99)

	// Calculate error rate
	if result.TotalRequests > 0 {
		result.ErrorRate = float64(result.FailedRequests) / float64(result.TotalRequests) * 100
	}

	// Calculate cache hit rate
	if result.CacheHits+result.CacheMisses > 0 {
		result.CacheHitRate = float64(result.CacheHits) / float64(result.CacheHits+result.CacheMisses) * 100
	}
}

// calculateOverallResults calculates overall test results
func (qpts *QueryPerformanceTestSuite) calculateOverallResults() {
	var totalRequests, successfulRequests, failedRequests int64
	var totalResponseTime time.Duration
	var responseTimes []time.Duration

	// Aggregate results from all query types
	for _, result := range qpts.results.QueryResults {
		totalRequests += result.TotalRequests
		successfulRequests += result.SuccessfulRequests
		failedRequests += result.FailedRequests
		totalResponseTime += result.AverageResponseTime * time.Duration(result.TotalRequests)
		responseTimes = append(responseTimes, result.ResponseTimes...)
	}

	qpts.results.TotalRequests = totalRequests
	qpts.results.SuccessfulRequests = successfulRequests
	qpts.results.FailedRequests = failedRequests

	if totalRequests > 0 {
		qpts.results.AverageResponseTime = totalResponseTime / time.Duration(totalRequests)
		qpts.results.ErrorRate = float64(failedRequests) / float64(totalRequests) * 100
		qpts.results.Throughput = float64(successfulRequests) / qpts.results.Duration.Seconds()
	}

	// Calculate percentiles
	if len(responseTimes) > 0 {
		qpts.results.P95ResponseTime = qpts.calculatePercentile(responseTimes, 95)
		qpts.results.P99ResponseTime = qpts.calculatePercentile(responseTimes, 99)

		// Find min and max
		qpts.results.MinResponseTime = responseTimes[0]
		qpts.results.MaxResponseTime = responseTimes[0]
		for _, rt := range responseTimes {
			if rt < qpts.results.MinResponseTime {
				qpts.results.MinResponseTime = rt
			}
			if rt > qpts.results.MaxResponseTime {
				qpts.results.MaxResponseTime = rt
			}
		}
	}

	// Calculate overall cache hit rate
	metrics := qpts.cacheManager.GetMetrics()
	qpts.results.CacheHitRate = metrics.HitRate
}

// validatePerformanceRequirements validates that performance requirements are met
func (qpts *QueryPerformanceTestSuite) validatePerformanceRequirements() {
	qpts.results.Passed = true

	// Check response time requirements
	if qpts.results.AverageResponseTime > qpts.config.MaxResponseTime {
		qpts.results.Passed = false
		qpts.results.Failures = append(qpts.results.Failures,
			fmt.Sprintf("Average response time %.2fms exceeds maximum %.2fms",
				float64(qpts.results.AverageResponseTime.Nanoseconds())/1e6,
				float64(qpts.config.MaxResponseTime.Nanoseconds())/1e6))
	}

	// Check cache hit rate requirements
	if qpts.results.CacheHitRate < qpts.config.MinHitRate {
		qpts.results.Passed = false
		qpts.results.Failures = append(qpts.results.Failures,
			fmt.Sprintf("Cache hit rate %.2f%% below minimum %.2f%%",
				qpts.results.CacheHitRate, qpts.config.MinHitRate))
	}

	// Check error rate requirements
	if qpts.results.ErrorRate > qpts.config.MaxErrorRate {
		qpts.results.Passed = false
		qpts.results.Failures = append(qpts.results.Failures,
			fmt.Sprintf("Error rate %.2f%% exceeds maximum %.2f%%",
				qpts.results.ErrorRate, qpts.config.MaxErrorRate))
	}
}

// Helper methods

func (qpts *QueryPerformanceTestSuite) calculatePercentile(times []time.Duration, percentile int) time.Duration {
	if len(times) == 0 {
		return 0
	}

	// Sort times
	for i := 0; i < len(times); i++ {
		for j := i + 1; j < len(times); j++ {
			if times[i] > times[j] {
				times[i], times[j] = times[j], times[i]
			}
		}
	}

	index := int(float64(len(times)) * float64(percentile) / 100.0)
	if index >= len(times) {
		index = len(times) - 1
	}

	return times[index]
}

func (qpts *QueryPerformanceTestSuite) getTestBusinessIDs(ctx context.Context) ([]string, error) {
	query := "SELECT id FROM businesses LIMIT 100"
	rows, err := qpts.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var businessIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		businessIDs = append(businessIDs, id)
	}

	return businessIDs, nil
}

func (qpts *QueryPerformanceTestSuite) getTestIndustryIDs(ctx context.Context) ([]int, error) {
	query := "SELECT id FROM industries LIMIT 50"
	rows, err := qpts.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var industryIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		industryIDs = append(industryIDs, id)
	}

	return industryIDs, nil
}

// getDefaultTestConfig returns default test configuration
func getDefaultTestConfig() *TestConfig {
	return &TestConfig{
		ConcurrentUsers:       10,
		TestDuration:          5 * time.Minute,
		RequestInterval:       100 * time.Millisecond,
		WarmupDuration:        30 * time.Second,
		MaxResponseTime:       200 * time.Millisecond,
		MinHitRate:            70.0,
		MaxErrorRate:          5.0,
		TestDataSize:          100,
		EnableDataCleanup:     true,
		EnableDetailedLogging: false,
		ReportInterval:        30 * time.Second,
	}
}
