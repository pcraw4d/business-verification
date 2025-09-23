// Package main provides a command-line tool to run query performance tests
// This tool validates the effectiveness of database query optimizations and caching
// to ensure the KYB Platform meets performance requirements.

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"kyb-platform/internal/cache"
	"kyb-platform/internal/testing"

	_ "github.com/lib/pq"
)

func main() {
	// Command line flags
	var (
		dbHost     = flag.String("db-host", "localhost", "Database host")
		dbPort     = flag.Int("db-port", 5432, "Database port")
		dbUser     = flag.String("db-user", "postgres", "Database user")
		dbPassword = flag.String("db-password", "", "Database password")
		dbName     = flag.String("db-name", "kyb_platform", "Database name")

		redisHost     = flag.String("redis-host", "localhost", "Redis host")
		redisPort     = flag.Int("redis-port", 6379, "Redis port")
		redisPassword = flag.String("redis-password", "", "Redis password")
		redisDB       = flag.Int("redis-db", 0, "Redis database number")

		concurrentUsers = flag.Int("concurrent-users", 10, "Number of concurrent users")
		testDuration    = flag.Duration("test-duration", 5*time.Minute, "Test duration")
		testDataSize    = flag.Int("test-data-size", 100, "Number of test iterations per query type")

		outputFile = flag.String("output", "performance_test_results.json", "Output file for test results")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	)

	flag.Parse()

	// Set up logging
	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	log.Println("Starting KYB Platform Query Performance Tests...")

	// Connect to database
	db, err := connectDatabase(*dbHost, *dbPort, *dbUser, *dbPassword, *dbName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Set up cache manager
	cacheConfig := &cache.CacheConfig{
		RedisHost:          *redisHost,
		RedisPort:          *redisPort,
		RedisPassword:      *redisPassword,
		RedisDB:            *redisDB,
		DefaultTTL:         15 * time.Minute,
		ClassificationTTL:  30 * time.Minute,
		RiskAssessmentTTL:  1 * time.Hour,
		UserDataTTL:        2 * time.Hour,
		BusinessDataTTL:    1 * time.Hour,
		LocalCacheSize:     1000,
		LocalCacheTTL:      5 * time.Minute,
		EnableInvalidation: true,
		InvalidationDelay:  1 * time.Second,
		EnableCompression:  false,
		EnableMetrics:      true,
	}

	cacheManager, err := cache.NewQueryCacheManager(cacheConfig)
	if err != nil {
		log.Fatalf("Failed to create cache manager: %v", err)
	}

	// Set up cached query executor
	executorConfig := &cache.ExecutorConfig{
		EnableCaching:      true,
		DefaultCacheTTL:    15 * time.Minute,
		CacheOnError:       false,
		MaxCacheSize:       10000,
		EnableQueryLogging: *verbose,
	}

	executor := cache.NewCachedQueryExecutor(cacheManager, db, executorConfig)

	// Set up test configuration
	testConfig := &testing.TestConfig{
		ConcurrentUsers:       *concurrentUsers,
		TestDuration:          *testDuration,
		RequestInterval:       100 * time.Millisecond,
		WarmupDuration:        30 * time.Second,
		MaxResponseTime:       200 * time.Millisecond,
		MinHitRate:            70.0,
		MaxErrorRate:          5.0,
		TestDataSize:          *testDataSize,
		EnableDataCleanup:     true,
		EnableDetailedLogging: *verbose,
		ReportInterval:        30 * time.Second,
	}

	// Create test suite
	testSuite := testing.NewQueryPerformanceTestSuite(db, cacheManager, executor, testConfig)

	// Run tests
	ctx, cancel := context.WithTimeout(context.Background(), *testDuration+10*time.Minute)
	defer cancel()

	results, err := testSuite.RunComprehensiveTests(ctx)
	if err != nil {
		log.Fatalf("Performance tests failed: %v", err)
	}

	// Output results
	if err := outputResults(results, *outputFile); err != nil {
		log.Fatalf("Failed to output results: %v", err)
	}

	// Print summary
	printSummary(results)

	// Exit with appropriate code
	if results.Passed {
		log.Println("✅ All performance tests PASSED!")
		os.Exit(0)
	} else {
		log.Println("❌ Some performance tests FAILED!")
		os.Exit(1)
	}
}

// connectDatabase connects to the PostgreSQL database
func connectDatabase(host string, port int, user, password, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	return db, nil
}

// outputResults outputs test results to a JSON file
func outputResults(results *testing.TestResults, filename string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}

	log.Printf("Test results written to %s", filename)
	return nil
}

// printSummary prints a summary of test results
func printSummary(results *testing.TestResults) {
	fmt.Println("\n" + "="*80)
	fmt.Println("KYB PLATFORM QUERY PERFORMANCE TEST SUMMARY")
	fmt.Println("=" * 80)

	fmt.Printf("Test Duration: %v\n", results.Duration)
	fmt.Printf("Total Requests: %d\n", results.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", results.SuccessfulRequests)
	fmt.Printf("Failed Requests: %d\n", results.FailedRequests)
	fmt.Printf("Error Rate: %.2f%%\n", results.ErrorRate)
	fmt.Printf("Throughput: %.2f requests/second\n", results.Throughput)

	fmt.Println("\nResponse Time Statistics:")
	fmt.Printf("  Average: %.2fms\n", float64(results.AverageResponseTime.Nanoseconds())/1e6)
	fmt.Printf("  Minimum: %.2fms\n", float64(results.MinResponseTime.Nanoseconds())/1e6)
	fmt.Printf("  Maximum: %.2fms\n", float64(results.MaxResponseTime.Nanoseconds())/1e6)
	fmt.Printf("  95th Percentile: %.2fms\n", float64(results.P95ResponseTime.Nanoseconds())/1e6)
	fmt.Printf("  99th Percentile: %.2fms\n", float64(results.P99ResponseTime.Nanoseconds())/1e6)

	fmt.Printf("\nCache Performance:")
	fmt.Printf("  Hit Rate: %.2f%%\n", results.CacheHitRate)

	fmt.Println("\nQuery Type Results:")
	for queryType, result := range results.QueryResults {
		fmt.Printf("  %s:\n", queryType)
		fmt.Printf("    Requests: %d\n", result.TotalRequests)
		fmt.Printf("    Average Response Time: %.2fms\n", float64(result.AverageResponseTime.Nanoseconds())/1e6)
		fmt.Printf("    Error Rate: %.2f%%\n", result.ErrorRate)
		fmt.Printf("    Cache Hit Rate: %.2f%%\n", result.CacheHitRate)
	}

	fmt.Printf("\nOverall Result: ")
	if results.Passed {
		fmt.Println("✅ PASSED")
	} else {
		fmt.Println("❌ FAILED")
		fmt.Println("\nFailures:")
		for _, failure := range results.Failures {
			fmt.Printf("  - %s\n", failure)
		}
	}

	fmt.Println("=" * 80)
}

// Example usage:
// go run scripts/run_query_performance_tests.go \
//   -db-host=localhost \
//   -db-port=5432 \
//   -db-user=postgres \
//   -db-password=password \
//   -db-name=kyb_platform \
//   -redis-host=localhost \
//   -redis-port=6379 \
//   -concurrent-users=20 \
//   -test-duration=10m \
//   -test-data-size=200 \
//   -output=performance_results.json \
//   -verbose
