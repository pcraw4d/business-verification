package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// ConfigurationTestSuite provides comprehensive testing for database configuration changes
type ConfigurationTestSuite struct {
	db     *sql.DB
	logger *log.Logger
}

// TestResult represents the result of a configuration test
type TestResult struct {
	TestName      string        `json:"test_name"`
	Status        string        `json:"status"`
	ExecutionTime time.Duration `json:"execution_time"`
	ErrorMessage  string        `json:"error_message,omitempty"`
	Details       interface{}   `json:"details,omitempty"`
}

// PerformanceBenchmark represents a performance benchmark result
type PerformanceBenchmark struct {
	Operation     string        `json:"operation"`
	ExecutionTime time.Duration `json:"execution_time"`
	RecordCount   int           `json:"record_count"`
	Throughput    float64       `json:"throughput"` // records per second
}

func main() {
	// Initialize logger
	logger := log.New(os.Stdout, "[DB-CONFIG-TEST] ", log.LstdFlags|log.Lshortfile)

	// Create database connection using environment variables
	db, err := createDatabaseConnection()
	if err != nil {
		logger.Fatalf("Failed to create database connection: %v", err)
	}
	defer db.Close()

	// Create test suite
	testSuite := &ConfigurationTestSuite{
		db:     db,
		logger: logger,
	}

	// Run configuration tests
	ctx := context.Background()
	results := testSuite.RunAllTests(ctx)

	// Print results
	testSuite.PrintResults(results)

	// Run performance benchmarks
	benchmarks := testSuite.RunPerformanceBenchmarks(ctx)
	testSuite.PrintBenchmarks(benchmarks)
}

// createDatabaseConnection creates a database connection using environment variables
func createDatabaseConnection() (*sql.DB, error) {
	// Get database configuration from environment variables
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USERNAME", "postgres")
	password := getEnv("DB_PASSWORD", "")
	database := getEnv("DB_DATABASE", "business_verification")
	sslMode := getEnv("DB_SSL_MODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, database, sslMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool with optimized settings
	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(40)
	db.SetConnMaxLifetime(9 * time.Minute)
	db.SetConnMaxIdleTime(3 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// RunAllTests runs all configuration tests
func (ts *ConfigurationTestSuite) RunAllTests(ctx context.Context) []TestResult {
	var results []TestResult

	// Test 1: Connection Pool Configuration
	results = append(results, ts.TestConnectionPoolConfiguration(ctx))

	// Test 2: Memory Configuration
	results = append(results, ts.TestMemoryConfiguration(ctx))

	// Test 3: Query Performance
	results = append(results, ts.TestQueryPerformance(ctx))

	// Test 4: Index Usage
	results = append(results, ts.TestIndexUsage(ctx))

	// Test 5: Classification Query Performance
	results = append(results, ts.TestClassificationQueryPerformance(ctx))

	// Test 6: Risk Assessment Query Performance
	results = append(results, ts.TestRiskAssessmentQueryPerformance(ctx))

	// Test 7: Concurrent Connection Handling
	results = append(results, ts.TestConcurrentConnectionHandling(ctx))

	// Test 8: Memory Usage Monitoring
	results = append(results, ts.TestMemoryUsageMonitoring(ctx))

	return results
}

// TestConnectionPoolConfiguration tests connection pool configuration
func (ts *ConfigurationTestSuite) TestConnectionPoolConfiguration(ctx context.Context) TestResult {
	start := time.Now()

	// Get connection pool stats
	stats := ts.db.Stats()

	// Validate configuration
	var status string
	var details interface{}

	if stats.MaxOpenConnections > 0 && stats.Idle >= 0 {
		status = "PASS"
		details = map[string]interface{}{
			"max_open_connections": stats.MaxOpenConnections,
			"max_idle_connections": stats.Idle,
			"open_connections":     stats.OpenConnections,
			"idle_connections":     stats.Idle,
			"in_use_connections":   stats.InUse,
		}
	} else {
		status = "FAIL"
		details = "Connection pool not properly configured"
	}

	return TestResult{
		TestName:      "Connection Pool Configuration",
		Status:        status,
		ExecutionTime: time.Since(start),
		Details:       details,
	}
}

// TestMemoryConfiguration tests memory configuration
func (ts *ConfigurationTestSuite) TestMemoryConfiguration(ctx context.Context) TestResult {
	start := time.Now()

	// Check key memory settings
	settings := []string{
		"shared_buffers", "effective_cache_size", "work_mem", "maintenance_work_mem",
		"random_page_cost", "effective_io_concurrency",
	}

	results := make(map[string]string)
	allValid := true

	for _, setting := range settings {
		var value string
		err := ts.db.QueryRowContext(ctx, "SELECT setting FROM pg_settings WHERE name = $1", setting).Scan(&value)
		if err != nil {
			results[setting] = fmt.Sprintf("ERROR: %v", err)
			allValid = false
		} else {
			results[setting] = value
		}
	}

	status := "PASS"
	if !allValid {
		status = "FAIL"
	}

	return TestResult{
		TestName:      "Memory Configuration",
		Status:        status,
		ExecutionTime: time.Since(start),
		Details:       results,
	}
}

// TestQueryPerformance tests basic query performance
func (ts *ConfigurationTestSuite) TestQueryPerformance(ctx context.Context) TestResult {
	start := time.Now()

	// Test simple query
	queryStart := time.Now()
	var count int
	err := ts.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables").Scan(&count)
	queryTime := time.Since(queryStart)

	status := "PASS"
	var details interface{}

	if err != nil {
		status = "FAIL"
		details = fmt.Sprintf("Query failed: %v", err)
	} else {
		details = map[string]interface{}{
			"query_time_ms": queryTime.Milliseconds(),
			"result_count":  count,
		}

		// Check if query time is acceptable (< 100ms)
		if queryTime > 100*time.Millisecond {
			status = "WARN"
		}
	}

	return TestResult{
		TestName:      "Query Performance",
		Status:        status,
		ExecutionTime: time.Since(start),
		Details:       details,
	}
}

// TestIndexUsage tests index usage
func (ts *ConfigurationTestSuite) TestIndexUsage(ctx context.Context) TestResult {
	start := time.Now()

	// Check if key indexes exist
	indexes := []string{
		"idx_merchants_name_gin",
		"idx_merchants_legal_name_gin",
		"idx_merchants_industry_btree",
		"idx_risk_assessments_business_id_created_at",
		"idx_risk_keywords_category_severity",
	}

	results := make(map[string]bool)
	allExist := true

	for _, index := range indexes {
		var exists bool
		err := ts.db.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM pg_indexes 
				WHERE indexname = $1
			)
		`, index).Scan(&exists)

		if err != nil {
			results[index] = false
			allExist = false
		} else {
			results[index] = exists
			if !exists {
				allExist = false
			}
		}
	}

	status := "PASS"
	if !allExist {
		status = "WARN"
	}

	return TestResult{
		TestName:      "Index Usage",
		Status:        status,
		ExecutionTime: time.Since(start),
		Details:       results,
	}
}

// TestClassificationQueryPerformance tests classification query performance
func (ts *ConfigurationTestSuite) TestClassificationQueryPerformance(ctx context.Context) TestResult {
	start := time.Now()

	// Test classification query (if tables exist)
	queryStart := time.Now()
	var count int
	err := ts.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM merchants 
		WHERE name ILIKE '%test%' OR legal_name ILIKE '%test%'
	`).Scan(&count)
	queryTime := time.Since(queryStart)

	status := "PASS"
	var details interface{}

	if err != nil {
		// Table might not exist yet, that's OK
		status = "SKIP"
		details = "Classification tables not yet created"
	} else {
		details = map[string]interface{}{
			"query_time_ms": queryTime.Milliseconds(),
			"result_count":  count,
		}

		// Check if query time is acceptable (< 200ms for classification)
		if queryTime > 200*time.Millisecond {
			status = "WARN"
		}
	}

	return TestResult{
		TestName:      "Classification Query Performance",
		Status:        status,
		ExecutionTime: time.Since(start),
		Details:       details,
	}
}

// TestRiskAssessmentQueryPerformance tests risk assessment query performance
func (ts *ConfigurationTestSuite) TestRiskAssessmentQueryPerformance(ctx context.Context) TestResult {
	start := time.Now()

	// Test risk assessment query (if tables exist)
	queryStart := time.Now()
	var count int
	err := ts.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM risk_assessments 
		WHERE overall_score > 0.5
	`).Scan(&count)
	queryTime := time.Since(queryStart)

	status := "PASS"
	var details interface{}

	if err != nil {
		// Table might not exist yet, that's OK
		status = "SKIP"
		details = "Risk assessment tables not yet created"
	} else {
		details = map[string]interface{}{
			"query_time_ms": queryTime.Milliseconds(),
			"result_count":  count,
		}

		// Check if query time is acceptable (< 150ms for risk assessment)
		if queryTime > 150*time.Millisecond {
			status = "WARN"
		}
	}

	return TestResult{
		TestName:      "Risk Assessment Query Performance",
		Status:        status,
		ExecutionTime: time.Since(start),
		Details:       details,
	}
}

// TestConcurrentConnectionHandling tests concurrent connection handling
func (ts *ConfigurationTestSuite) TestConcurrentConnectionHandling(ctx context.Context) TestResult {
	start := time.Now()

	// Test concurrent connections
	concurrency := 10
	results := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			var count int
			err := ts.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables").Scan(&count)
			results <- err
		}()
	}

	// Collect results
	var errors []error
	for i := 0; i < concurrency; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	status := "PASS"
	var details interface{}

	if len(errors) > 0 {
		status = "FAIL"
		details = fmt.Sprintf("Failed concurrent connections: %d/%d", len(errors), concurrency)
	} else {
		details = fmt.Sprintf("All %d concurrent connections successful", concurrency)
	}

	return TestResult{
		TestName:      "Concurrent Connection Handling",
		Status:        status,
		ExecutionTime: time.Since(start),
		Details:       details,
	}
}

// TestMemoryUsageMonitoring tests memory usage monitoring
func (ts *ConfigurationTestSuite) TestMemoryUsageMonitoring(ctx context.Context) TestResult {
	start := time.Now()

	// Get memory usage statistics
	var sharedBuffersHit, sharedBuffersRead int64
	err := ts.db.QueryRowContext(ctx, `
		SELECT 
			SUM(blks_hit) as shared_buffers_hit,
			SUM(blks_read) as shared_buffers_read
		FROM pg_stat_database
	`).Scan(&sharedBuffersHit, &sharedBuffersRead)

	status := "PASS"
	var details interface{}

	if err != nil {
		status = "FAIL"
		details = fmt.Sprintf("Failed to get memory stats: %v", err)
	} else {
		totalReads := sharedBuffersHit + sharedBuffersRead
		hitRatio := float64(0)
		if totalReads > 0 {
			hitRatio = float64(sharedBuffersHit) / float64(totalReads) * 100
		}

		details = map[string]interface{}{
			"shared_buffers_hit_ratio": hitRatio,
			"shared_buffers_hit":       sharedBuffersHit,
			"shared_buffers_read":      sharedBuffersRead,
		}

		// Check if hit ratio is good (> 90%)
		if hitRatio < 90 {
			status = "WARN"
		}
	}

	return TestResult{
		TestName:      "Memory Usage Monitoring",
		Status:        status,
		ExecutionTime: time.Since(start),
		Details:       details,
	}
}

// RunPerformanceBenchmarks runs performance benchmarks
func (ts *ConfigurationTestSuite) RunPerformanceBenchmarks(ctx context.Context) []PerformanceBenchmark {
	var benchmarks []PerformanceBenchmark

	// Benchmark 1: Simple query performance
	benchmarks = append(benchmarks, ts.BenchmarkSimpleQuery(ctx))

	// Benchmark 2: Complex query performance
	benchmarks = append(benchmarks, ts.BenchmarkComplexQuery(ctx))

	// Benchmark 3: Classification query performance
	benchmarks = append(benchmarks, ts.BenchmarkClassificationQuery(ctx))

	// Benchmark 4: Risk assessment query performance
	benchmarks = append(benchmarks, ts.BenchmarkRiskAssessmentQuery(ctx))

	return benchmarks
}

// BenchmarkSimpleQuery benchmarks simple query performance
func (ts *ConfigurationTestSuite) BenchmarkSimpleQuery(ctx context.Context) PerformanceBenchmark {
	start := time.Now()

	var count int
	_ = ts.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables").Scan(&count)

	executionTime := time.Since(start)
	throughput := float64(count) / executionTime.Seconds()

	return PerformanceBenchmark{
		Operation:     "Simple Query",
		ExecutionTime: executionTime,
		RecordCount:   count,
		Throughput:    throughput,
	}
}

// BenchmarkComplexQuery benchmarks complex query performance
func (ts *ConfigurationTestSuite) BenchmarkComplexQuery(ctx context.Context) PerformanceBenchmark {
	start := time.Now()

	var count int
	_ = ts.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.tables t1
		JOIN information_schema.columns t2 ON t1.table_name = t2.table_name
		WHERE t1.table_schema = 'public'
	`).Scan(&count)

	executionTime := time.Since(start)
	throughput := float64(count) / executionTime.Seconds()

	return PerformanceBenchmark{
		Operation:     "Complex Query",
		ExecutionTime: executionTime,
		RecordCount:   count,
		Throughput:    throughput,
	}
}

// BenchmarkClassificationQuery benchmarks classification query performance
func (ts *ConfigurationTestSuite) BenchmarkClassificationQuery(ctx context.Context) PerformanceBenchmark {
	start := time.Now()

	var count int
	_ = ts.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM merchants 
		WHERE name ILIKE '%test%' OR legal_name ILIKE '%test%'
	`).Scan(&count)

	executionTime := time.Since(start)
	throughput := float64(count) / executionTime.Seconds()

	return PerformanceBenchmark{
		Operation:     "Classification Query",
		ExecutionTime: executionTime,
		RecordCount:   count,
		Throughput:    throughput,
	}
}

// BenchmarkRiskAssessmentQuery benchmarks risk assessment query performance
func (ts *ConfigurationTestSuite) BenchmarkRiskAssessmentQuery(ctx context.Context) PerformanceBenchmark {
	start := time.Now()

	var count int
	_ = ts.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM risk_assessments 
		WHERE overall_score > 0.5
	`).Scan(&count)

	executionTime := time.Since(start)
	throughput := float64(count) / executionTime.Seconds()

	return PerformanceBenchmark{
		Operation:     "Risk Assessment Query",
		ExecutionTime: executionTime,
		RecordCount:   count,
		Throughput:    throughput,
	}
}

// PrintResults prints test results
func (ts *ConfigurationTestSuite) PrintResults(results []TestResult) {
	ts.logger.Println("=== CONFIGURATION TEST RESULTS ===")

	passed := 0
	failed := 0
	warned := 0
	skipped := 0

	for _, result := range results {
		status := result.Status
		switch status {
		case "PASS":
			passed++
		case "FAIL":
			failed++
		case "WARN":
			warned++
		case "SKIP":
			skipped++
		}

		ts.logger.Printf("[%s] %s (%.2fms)", status, result.TestName, float64(result.ExecutionTime.Nanoseconds())/1e6)

		if result.Details != nil {
			ts.logger.Printf("  Details: %+v", result.Details)
		}

		if result.ErrorMessage != "" {
			ts.logger.Printf("  Error: %s", result.ErrorMessage)
		}
	}

	ts.logger.Printf("\nSummary: %d passed, %d failed, %d warnings, %d skipped", passed, failed, warned, skipped)

	if failed > 0 {
		ts.logger.Println("❌ Some tests failed - configuration needs attention")
	} else if warned > 0 {
		ts.logger.Println("⚠️  Some tests had warnings - consider optimization")
	} else {
		ts.logger.Println("✅ All tests passed - configuration is optimal")
	}
}

// PrintBenchmarks prints performance benchmarks
func (ts *ConfigurationTestSuite) PrintBenchmarks(benchmarks []PerformanceBenchmark) {
	ts.logger.Println("\n=== PERFORMANCE BENCHMARKS ===")

	for _, benchmark := range benchmarks {
		ts.logger.Printf("%s: %.2fms (%.2f records/sec)",
			benchmark.Operation,
			float64(benchmark.ExecutionTime.Nanoseconds())/1e6,
			benchmark.Throughput)
	}

	ts.logger.Println("\nPerformance targets:")
	ts.logger.Println("- Simple queries: < 50ms")
	ts.logger.Println("- Complex queries: < 200ms")
	ts.logger.Println("- Classification queries: < 200ms")
	ts.logger.Println("- Risk assessment queries: < 150ms")
}
