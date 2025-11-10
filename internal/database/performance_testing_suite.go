package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// DatabasePerformanceTestSuite provides comprehensive database performance testing
type DatabasePerformanceTestSuite struct {
	db     *sql.DB
	logger *log.Logger
	config *PerformanceTestConfig
}

// PerformanceTestConfig contains configuration for performance testing
type PerformanceTestConfig struct {
	// Test duration settings
	TestDuration     time.Duration
	WarmupDuration   time.Duration
	CooldownDuration time.Duration

	// Load testing settings
	ConcurrentUsers int
	RequestsPerUser int
	RequestInterval time.Duration

	// Performance thresholds
	MaxQueryTime      time.Duration
	MaxConnectionTime time.Duration
	MinThroughput     int // requests per second

	// Resource monitoring
	MonitorCPU         bool
	MonitorMemory      bool
	MonitorConnections bool
}

// PerformanceTestResult contains the results of a performance test
type PerformanceTestResult struct {
	TestName  string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration

	// Query performance metrics
	TotalQueries      int
	SuccessfulQueries int
	FailedQueries     int
	AverageQueryTime  time.Duration
	MinQueryTime      time.Duration
	MaxQueryTime      time.Duration
	P95QueryTime      time.Duration
	P99QueryTime      time.Duration

	// Throughput metrics
	QueriesPerSecond float64
	ConcurrentUsers  int

	// Resource usage
	CPUUsage        float64
	MemoryUsage     int64
	ConnectionCount int

	// Error analysis
	ErrorRate    float64
	CommonErrors map[string]int

	// Recommendations
	Recommendations []string
}

// NewDatabasePerformanceTestSuite creates a new performance testing suite
func NewDatabasePerformanceTestSuite(db *sql.DB, config *PerformanceTestConfig) *DatabasePerformanceTestSuite {
	if config == nil {
		config = &PerformanceTestConfig{
			TestDuration:       5 * time.Minute,
			WarmupDuration:     30 * time.Second,
			CooldownDuration:   30 * time.Second,
			ConcurrentUsers:    10,
			RequestsPerUser:    100,
			RequestInterval:    100 * time.Millisecond,
			MaxQueryTime:       1 * time.Second,
			MaxConnectionTime:  5 * time.Second,
			MinThroughput:      50,
			MonitorCPU:         true,
			MonitorMemory:      true,
			MonitorConnections: true,
		}
	}

	return &DatabasePerformanceTestSuite{
		db:     db,
		logger: log.New(log.Writer(), "[DB_PERF] ", log.LstdFlags),
		config: config,
	}
}

// RunComprehensivePerformanceTests runs all performance tests
func (suite *DatabasePerformanceTestSuite) RunComprehensivePerformanceTests(ctx context.Context) ([]*PerformanceTestResult, error) {
	suite.logger.Println("Starting comprehensive database performance testing...")

	var results []*PerformanceTestResult

	// Test 1: Basic Query Performance
	suite.logger.Println("Running basic query performance tests...")
	basicResult, err := suite.RunBasicQueryPerformanceTest(ctx)
	if err != nil {
		suite.logger.Printf("Basic query performance test failed: %v", err)
		return nil, fmt.Errorf("basic query performance test failed: %w", err)
	}
	results = append(results, basicResult)

	// Test 2: Index Performance
	suite.logger.Println("Running index performance tests...")
	indexResult, err := suite.RunIndexPerformanceTest(ctx)
	if err != nil {
		suite.logger.Printf("Index performance test failed: %v", err)
		return nil, fmt.Errorf("index performance test failed: %w", err)
	}
	results = append(results, indexResult)

	// Test 3: Concurrent Access Performance
	suite.logger.Println("Running concurrent access performance tests...")
	concurrentResult, err := suite.RunConcurrentAccessTest(ctx)
	if err != nil {
		suite.logger.Printf("Concurrent access test failed: %v", err)
		return nil, fmt.Errorf("concurrent access test failed: %w", err)
	}
	results = append(results, concurrentResult)

	// Test 4: Load Testing
	suite.logger.Println("Running load testing...")
	loadResult, err := suite.RunLoadTest(ctx)
	if err != nil {
		suite.logger.Printf("Load test failed: %v", err)
		return nil, fmt.Errorf("load test failed: %w", err)
	}
	results = append(results, loadResult)

	// Test 5: Resource Usage Monitoring
	suite.logger.Println("Running resource usage monitoring...")
	resourceResult, err := suite.RunResourceUsageTest(ctx)
	if err != nil {
		suite.logger.Printf("Resource usage test failed: %v", err)
		return nil, fmt.Errorf("resource usage test failed: %w", err)
	}
	results = append(results, resourceResult)

	suite.logger.Println("Comprehensive performance testing completed successfully")
	return results, nil
}

// RunBasicQueryPerformanceTest tests basic query performance
func (suite *DatabasePerformanceTestSuite) RunBasicQueryPerformanceTest(ctx context.Context) (*PerformanceTestResult, error) {
	startTime := time.Now()
	result := &PerformanceTestResult{
		TestName:     "Basic Query Performance",
		StartTime:    startTime,
		CommonErrors: make(map[string]int),
	}

	// Define test queries
	testQueries := []struct {
		name  string
		query string
	}{
		{
			name:  "User Lookup by Email",
			query: "SELECT id, email, name FROM users WHERE email = $1",
		},
		{
			name:  "Business Search by Name",
			query: "SELECT id, name, industry FROM businesses WHERE name ILIKE $1",
		},
		{
			name:  "Recent Classifications",
			query: "SELECT bc.*, b.name FROM business_classifications bc JOIN businesses b ON bc.business_id = b.id WHERE bc.created_at > $1 ORDER BY bc.created_at DESC LIMIT 100",
		},
		{
			name:  "Risk Assessment Summary",
			query: "SELECT risk_level, COUNT(*) as count FROM risk_assessments GROUP BY risk_level",
		},
		{
			name:  "Compliance Status Check",
			query: "SELECT status, COUNT(*) as count FROM compliance_checks WHERE created_at > $1 GROUP BY status",
		},
	}

	var totalQueries int
	var totalTime time.Duration
	var queryTimes []time.Duration

	for _, testQuery := range testQueries {
		suite.logger.Printf("Testing query: %s", testQuery.name)

		// Run query multiple times for accurate measurement
		for i := 0; i < 10; i++ {
			queryStart := time.Now()

			// Prepare test data based on query type
			var rows *sql.Rows
			var err error

			switch testQuery.name {
			case "User Lookup by Email":
				rows, err = suite.db.QueryContext(ctx, testQuery.query, "test@example.com")
			case "Business Search by Name":
				rows, err = suite.db.QueryContext(ctx, testQuery.query, "%test%")
			case "Recent Classifications":
				rows, err = suite.db.QueryContext(ctx, testQuery.query, time.Now().Add(-24*time.Hour))
			case "Risk Assessment Summary":
				rows, err = suite.db.QueryContext(ctx, testQuery.query)
			case "Compliance Status Check":
				rows, err = suite.db.QueryContext(ctx, testQuery.query, time.Now().Add(-24*time.Hour))
			default:
				rows, err = suite.db.QueryContext(ctx, testQuery.query)
			}

			queryTime := time.Since(queryStart)

			if err != nil {
				result.FailedQueries++
				result.CommonErrors[err.Error()]++
				suite.logger.Printf("Query failed: %v", err)
			} else {
				result.SuccessfulQueries++
				totalQueries++
				totalTime += queryTime
				queryTimes = append(queryTimes, queryTime)

				// Close rows to prevent connection leaks
				rows.Close()
			}
		}
	}

	// Calculate statistics
	result.TotalQueries = totalQueries
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if totalQueries > 0 {
		result.AverageQueryTime = totalTime / time.Duration(totalQueries)
		result.QueriesPerSecond = float64(totalQueries) / result.Duration.Seconds()

		// Calculate percentiles
		if len(queryTimes) > 0 {
			result.MinQueryTime = queryTimes[0]
			result.MaxQueryTime = queryTimes[0]

			for _, qt := range queryTimes {
				if qt < result.MinQueryTime {
					result.MinQueryTime = qt
				}
				if qt > result.MaxQueryTime {
					result.MaxQueryTime = qt
				}
			}

			// Simple percentile calculation (P95, P99)
			if len(queryTimes) >= 20 {
				p95Index := int(float64(len(queryTimes)) * 0.95)
				p99Index := int(float64(len(queryTimes)) * 0.99)
				result.P95QueryTime = queryTimes[p95Index]
				result.P99QueryTime = queryTimes[p99Index]
			}
		}
	}

	// Calculate error rate
	if result.TotalQueries > 0 {
		result.ErrorRate = float64(result.FailedQueries) / float64(result.TotalQueries) * 100
	}

	// Generate recommendations
	suite.generateBasicQueryRecommendations(result)

	return result, nil
}

// RunIndexPerformanceTest tests index performance
func (suite *DatabasePerformanceTestSuite) RunIndexPerformanceTest(ctx context.Context) (*PerformanceTestResult, error) {
	startTime := time.Now()
	result := &PerformanceTestResult{
		TestName:     "Index Performance",
		StartTime:    startTime,
		CommonErrors: make(map[string]int),
	}

	// Test index usage with EXPLAIN ANALYZE
	indexTests := []struct {
		name  string
		query string
		index string
	}{
		{
			name:  "Users Email Index",
			query: "SELECT id, email FROM users WHERE email = $1",
			index: "idx_users_email",
		},
		{
			name:  "Businesses User ID Index",
			query: "SELECT id, name FROM businesses WHERE user_id = $1",
			index: "idx_businesses_user_id",
		},
		{
			name:  "Classifications Business ID Index",
			query: "SELECT * FROM business_classifications WHERE business_id = $1",
			index: "idx_business_classifications_business_id",
		},
		{
			name:  "Risk Assessments Risk Level Index",
			query: "SELECT * FROM risk_assessments WHERE risk_level = $1",
			index: "idx_risk_assessments_risk_level",
		},
		{
			name:  "Audit Logs Timestamp Index",
			query: "SELECT * FROM audit_logs WHERE created_at > $1 ORDER BY created_at DESC",
			index: "idx_audit_logs_created_at",
		},
	}

	var totalQueries int
	var totalTime time.Duration
	var queryTimes []time.Duration

	for _, test := range indexTests {
		suite.logger.Printf("Testing index: %s", test.name)

		// Run query with EXPLAIN ANALYZE to check index usage
		for i := 0; i < 5; i++ {
			queryStart := time.Now()

			// Use EXPLAIN ANALYZE to check if index is being used
			explainQuery := fmt.Sprintf("EXPLAIN ANALYZE %s", test.query)
			var rows *sql.Rows
			var err error

			switch test.name {
			case "Users Email Index":
				rows, err = suite.db.QueryContext(ctx, explainQuery, "test@example.com")
			case "Businesses User ID Index":
				rows, err = suite.db.QueryContext(ctx, explainQuery, "00000000-0000-0000-0000-000000000000")
			case "Classifications Business ID Index":
				rows, err = suite.db.QueryContext(ctx, explainQuery, "00000000-0000-0000-0000-000000000000")
			case "Risk Assessments Risk Level Index":
				rows, err = suite.db.QueryContext(ctx, explainQuery, "low")
			case "Audit Logs Timestamp Index":
				rows, err = suite.db.QueryContext(ctx, explainQuery, time.Now().Add(-24*time.Hour))
			}

			queryTime := time.Since(queryStart)

			if err != nil {
				result.FailedQueries++
				result.CommonErrors[err.Error()]++
			} else {
				result.SuccessfulQueries++
				totalQueries++
				totalTime += queryTime
				queryTimes = append(queryTimes, queryTime)

				// Check if index is being used
				var explainText string
				for rows.Next() {
					var line string
					if err := rows.Scan(&line); err == nil {
						explainText += line + "\n"
					}
				}
				rows.Close()

				// Check if the expected index is mentioned in the explain plan
				if !performanceContains(explainText, test.index) && !performanceContains(explainText, "Index Scan") {
					suite.logger.Printf("Warning: Index %s may not be used for query %s", test.index, test.name)
				}
			}
		}
	}

	// Calculate statistics
	result.TotalQueries = totalQueries
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if totalQueries > 0 {
		result.AverageQueryTime = totalTime / time.Duration(totalQueries)
		result.QueriesPerSecond = float64(totalQueries) / result.Duration.Seconds()

		// Calculate percentiles
		if len(queryTimes) > 0 {
			result.MinQueryTime = queryTimes[0]
			result.MaxQueryTime = queryTimes[0]

			for _, qt := range queryTimes {
				if qt < result.MinQueryTime {
					result.MinQueryTime = qt
				}
				if qt > result.MaxQueryTime {
					result.MaxQueryTime = qt
				}
			}
		}
	}

	// Calculate error rate
	if result.TotalQueries > 0 {
		result.ErrorRate = float64(result.FailedQueries) / float64(result.TotalQueries) * 100
	}

	// Generate recommendations
	suite.generateIndexRecommendations(result)

	return result, nil
}

// RunConcurrentAccessTest tests concurrent access performance
func (suite *DatabasePerformanceTestSuite) RunConcurrentAccessTest(ctx context.Context) (*PerformanceTestResult, error) {
	startTime := time.Now()
	result := &PerformanceTestResult{
		TestName:     "Concurrent Access Performance",
		StartTime:    startTime,
		CommonErrors: make(map[string]int),
	}

	// Test concurrent access with multiple goroutines
	var wg sync.WaitGroup
	var mu sync.Mutex
	var totalQueries int
	var totalTime time.Duration
	var queryTimes []time.Duration

	concurrentUsers := suite.config.ConcurrentUsers
	if concurrentUsers == 0 {
		concurrentUsers = 10
	}

	suite.logger.Printf("Testing concurrent access with %d users", concurrentUsers)

	// Start concurrent users
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			// Each user performs multiple queries
			for j := 0; j < suite.config.RequestsPerUser; j++ {
				queryStart := time.Now()

				// Simulate different types of concurrent operations
				var err error
				switch j % 4 {
				case 0:
					// Read operation
					_, err = suite.db.QueryContext(ctx, "SELECT COUNT(*) FROM users")
				case 1:
					// Write operation (if safe)
					_, err = suite.db.ExecContext(ctx, "SELECT 1") // Safe query
				case 2:
					// Complex query
					_, err = suite.db.QueryContext(ctx, "SELECT u.email, COUNT(b.id) FROM users u LEFT JOIN businesses b ON u.id = b.user_id GROUP BY u.id, u.email LIMIT 10")
				case 3:
					// Index lookup
					_, err = suite.db.QueryContext(ctx, "SELECT * FROM users WHERE email = $1", fmt.Sprintf("user%d@example.com", userID))
				}

				queryTime := time.Since(queryStart)

				mu.Lock()
				if err != nil {
					result.FailedQueries++
					result.CommonErrors[err.Error()]++
				} else {
					result.SuccessfulQueries++
					totalQueries++
					totalTime += queryTime
					queryTimes = append(queryTimes, queryTime)
				}
				mu.Unlock()

				// Small delay between requests
				time.Sleep(suite.config.RequestInterval)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Calculate statistics
	result.TotalQueries = totalQueries
	result.ConcurrentUsers = concurrentUsers
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if totalQueries > 0 {
		result.AverageQueryTime = totalTime / time.Duration(totalQueries)
		result.QueriesPerSecond = float64(totalQueries) / result.Duration.Seconds()

		// Calculate percentiles
		if len(queryTimes) > 0 {
			result.MinQueryTime = queryTimes[0]
			result.MaxQueryTime = queryTimes[0]

			for _, qt := range queryTimes {
				if qt < result.MinQueryTime {
					result.MinQueryTime = qt
				}
				if qt > result.MaxQueryTime {
					result.MaxQueryTime = qt
				}
			}
		}
	}

	// Calculate error rate
	if result.TotalQueries > 0 {
		result.ErrorRate = float64(result.FailedQueries) / float64(result.TotalQueries) * 100
	}

	// Generate recommendations
	suite.generateConcurrentAccessRecommendations(result)

	return result, nil
}

// RunLoadTest performs load testing on the database
func (suite *DatabasePerformanceTestSuite) RunLoadTest(ctx context.Context) (*PerformanceTestResult, error) {
	startTime := time.Now()
	result := &PerformanceTestResult{
		TestName:     "Load Test",
		StartTime:    startTime,
		CommonErrors: make(map[string]int),
	}

	// Run load test for specified duration
	testDuration := suite.config.TestDuration
	if testDuration == 0 {
		testDuration = 2 * time.Minute
	}

	suite.logger.Printf("Running load test for %v", testDuration)

	// Create context with timeout
	loadCtx, cancel := context.WithTimeout(ctx, testDuration)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var totalQueries int
	var totalTime time.Duration
	var queryTimes []time.Duration

	// Start load test goroutines
	for i := 0; i < suite.config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			for {
				select {
				case <-loadCtx.Done():
					return
				default:
					queryStart := time.Now()

					// Perform various database operations
					var err error
					switch userID % 5 {
					case 0:
						// User operations
						_, err = suite.db.QueryContext(loadCtx, "SELECT id, email FROM users LIMIT 10")
					case 1:
						// Business operations
						_, err = suite.db.QueryContext(loadCtx, "SELECT id, name FROM businesses LIMIT 10")
					case 2:
						// Classification operations
						_, err = suite.db.QueryContext(loadCtx, "SELECT * FROM business_classifications LIMIT 10")
					case 3:
						// Risk assessment operations
						_, err = suite.db.QueryContext(loadCtx, "SELECT * FROM risk_assessments LIMIT 10")
					case 4:
						// Complex join operations
						_, err = suite.db.QueryContext(loadCtx, "SELECT u.email, b.name, bc.industry FROM users u JOIN businesses b ON u.id = b.user_id LEFT JOIN business_classifications bc ON b.id = bc.business_id LIMIT 10")
					}

					queryTime := time.Since(queryStart)

					mu.Lock()
					if err != nil {
						result.FailedQueries++
						result.CommonErrors[err.Error()]++
					} else {
						result.SuccessfulQueries++
						totalQueries++
						totalTime += queryTime
						queryTimes = append(queryTimes, queryTime)
					}
					mu.Unlock()

					// Small delay between requests
					time.Sleep(suite.config.RequestInterval)
				}
			}
		}(i)
	}

	// Wait for load test to complete
	wg.Wait()

	// Calculate statistics
	result.TotalQueries = totalQueries
	result.ConcurrentUsers = suite.config.ConcurrentUsers
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if totalQueries > 0 {
		result.AverageQueryTime = totalTime / time.Duration(totalQueries)
		result.QueriesPerSecond = float64(totalQueries) / result.Duration.Seconds()

		// Calculate percentiles
		if len(queryTimes) > 0 {
			result.MinQueryTime = queryTimes[0]
			result.MaxQueryTime = queryTimes[0]

			for _, qt := range queryTimes {
				if qt < result.MinQueryTime {
					result.MinQueryTime = qt
				}
				if qt > result.MaxQueryTime {
					result.MaxQueryTime = qt
				}
			}
		}
	}

	// Calculate error rate
	if result.TotalQueries > 0 {
		result.ErrorRate = float64(result.FailedQueries) / float64(result.TotalQueries) * 100
	}

	// Generate recommendations
	suite.generateLoadTestRecommendations(result)

	return result, nil
}

// RunResourceUsageTest monitors resource usage during database operations
func (suite *DatabasePerformanceTestSuite) RunResourceUsageTest(ctx context.Context) (*PerformanceTestResult, error) {
	startTime := time.Now()
	result := &PerformanceTestResult{
		TestName:     "Resource Usage Test",
		StartTime:    startTime,
		CommonErrors: make(map[string]int),
	}

	// Monitor database connection count
	var connectionCount int
	err := suite.db.QueryRowContext(ctx, "SELECT count(*) FROM pg_stat_activity WHERE state = 'active'").Scan(&connectionCount)
	if err != nil {
		suite.logger.Printf("Failed to get connection count: %v", err)
	} else {
		result.ConnectionCount = connectionCount
	}

	// Monitor database size
	var dbSize int64
	err = suite.db.QueryRowContext(ctx, "SELECT pg_database_size(current_database())").Scan(&dbSize)
	if err != nil {
		suite.logger.Printf("Failed to get database size: %v", err)
	} else {
		result.MemoryUsage = dbSize
	}

	// Run some queries to measure resource usage
	testQueries := []string{
		"SELECT COUNT(*) FROM users",
		"SELECT COUNT(*) FROM businesses",
		"SELECT COUNT(*) FROM business_classifications",
		"SELECT COUNT(*) FROM risk_assessments",
		"SELECT COUNT(*) FROM compliance_checks",
	}

	var totalQueries int
	var totalTime time.Duration

	for _, query := range testQueries {
		queryStart := time.Now()
		_, err := suite.db.QueryContext(ctx, query)
		queryTime := time.Since(queryStart)

		if err != nil {
			result.FailedQueries++
			result.CommonErrors[err.Error()]++
		} else {
			result.SuccessfulQueries++
			totalQueries++
			totalTime += queryTime
		}
	}

	// Calculate statistics
	result.TotalQueries = totalQueries
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if totalQueries > 0 {
		result.AverageQueryTime = totalTime / time.Duration(totalQueries)
		result.QueriesPerSecond = float64(totalQueries) / result.Duration.Seconds()
	}

	// Calculate error rate
	if result.TotalQueries > 0 {
		result.ErrorRate = float64(result.FailedQueries) / float64(result.TotalQueries) * 100
	}

	// Generate recommendations
	suite.generateResourceUsageRecommendations(result)

	return result, nil
}

// Helper functions for generating recommendations
func (suite *DatabasePerformanceTestSuite) generateBasicQueryRecommendations(result *PerformanceTestResult) {
	if result.AverageQueryTime > suite.config.MaxQueryTime {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Average query time (%.2fms) exceeds threshold (%.2fms). Consider query optimization or additional indexing.",
				float64(result.AverageQueryTime.Nanoseconds())/1e6,
				float64(suite.config.MaxQueryTime.Nanoseconds())/1e6))
	}

	if result.ErrorRate > 5.0 {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("High error rate (%.2f%%). Review query logic and database constraints.", result.ErrorRate))
	}

	if result.QueriesPerSecond < float64(suite.config.MinThroughput) {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Low throughput (%.2f qps) below threshold (%d qps). Consider performance optimization.",
				result.QueriesPerSecond, suite.config.MinThroughput))
	}
}

func (suite *DatabasePerformanceTestSuite) generateIndexRecommendations(result *PerformanceTestResult) {
	if result.AverageQueryTime > 100*time.Millisecond {
		result.Recommendations = append(result.Recommendations,
			"Index queries are slower than expected. Review index usage and consider additional indexes.")
	}

	if result.ErrorRate > 1.0 {
		result.Recommendations = append(result.Recommendations,
			"Index-related errors detected. Verify index existence and query syntax.")
	}
}

func (suite *DatabasePerformanceTestSuite) generateConcurrentAccessRecommendations(result *PerformanceTestResult) {
	if result.ErrorRate > 10.0 {
		result.Recommendations = append(result.Recommendations,
			"High error rate under concurrent access. Review connection pooling and transaction handling.")
	}

	if result.AverageQueryTime > 2*time.Second {
		result.Recommendations = append(result.Recommendations,
			"Slow queries under concurrent load. Consider query optimization and connection tuning.")
	}

	if result.ConnectionCount > 80 {
		result.Recommendations = append(result.Recommendations,
			"High connection count detected. Review connection pooling configuration.")
	}
}

func (suite *DatabasePerformanceTestSuite) generateLoadTestRecommendations(result *PerformanceTestResult) {
	if result.QueriesPerSecond < float64(suite.config.MinThroughput) {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Load test throughput (%.2f qps) below target (%d qps). Consider scaling or optimization.",
				result.QueriesPerSecond, suite.config.MinThroughput))
	}

	if result.ErrorRate > 5.0 {
		result.Recommendations = append(result.Recommendations,
			"High error rate during load test. Review system capacity and error handling.")
	}

	if result.MaxQueryTime > 5*time.Second {
		result.Recommendations = append(result.Recommendations,
			"Some queries are very slow under load. Identify and optimize slow queries.")
	}
}

func (suite *DatabasePerformanceTestSuite) generateResourceUsageRecommendations(result *PerformanceTestResult) {
	if result.ConnectionCount > 50 {
		result.Recommendations = append(result.Recommendations,
			"High connection count. Consider optimizing connection pooling or reducing connection lifetime.")
	}

	if result.MemoryUsage > 1024*1024*1024 { // 1GB
		result.Recommendations = append(result.Recommendations,
			"Large database size detected. Consider data archiving or partitioning strategies.")
	}

	if result.AverageQueryTime > 500*time.Millisecond {
		result.Recommendations = append(result.Recommendations,
			"Slow queries detected. Review query performance and consider additional optimization.")
	}
}

// Helper function to check if string contains substring
// (renamed to avoid conflict with load_testing_suite.go)
func performanceContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			performanceContainsSubstring(s, substr))))
}

func performanceContainsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
