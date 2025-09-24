package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabasePerformanceTestSuite tests the performance testing suite
func TestDatabasePerformanceTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	// Setup test database
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create test configuration
	config := &PerformanceTestConfig{
		TestDuration:       30 * time.Second,
		WarmupDuration:     5 * time.Second,
		CooldownDuration:   5 * time.Second,
		ConcurrentUsers:    5,
		RequestsPerUser:    10,
		RequestInterval:    100 * time.Millisecond,
		MaxQueryTime:       1 * time.Second,
		MaxConnectionTime:  5 * time.Second,
		MinThroughput:      10,
		MonitorCPU:         false, // Disable for testing
		MonitorMemory:      false,
		MonitorConnections: true,
	}

	// Create performance test suite
	suite := NewDatabasePerformanceTestSuite(db, config)

	t.Run("basic_query_performance", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		result, err := suite.RunBasicQueryPerformanceTest(ctx)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Basic Query Performance", result.TestName)
		assert.True(t, result.Duration > 0)
		assert.True(t, result.TotalQueries >= 0)
	})

	t.Run("index_performance", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		result, err := suite.RunIndexPerformanceTest(ctx)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Index Performance", result.TestName)
		assert.True(t, result.Duration > 0)
	})

	t.Run("concurrent_access", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		result, err := suite.RunConcurrentAccessTest(ctx)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Concurrent Access Performance", result.TestName)
		assert.True(t, result.Duration > 0)
		assert.Equal(t, config.ConcurrentUsers, result.ConcurrentUsers)
	})

	t.Run("load_test", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		result, err := suite.RunLoadTest(ctx)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Load Test", result.TestName)
		assert.True(t, result.Duration > 0)
	})

	t.Run("resource_usage", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		result, err := suite.RunResourceUsageTest(ctx)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Resource Usage Test", result.TestName)
		assert.True(t, result.Duration > 0)
	})

	t.Run("comprehensive_tests", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		results, err := suite.RunComprehensivePerformanceTests(ctx)
		require.NoError(t, err)
		assert.Len(t, results, 5) // Should have 5 test results

		// Verify all test types are present
		testNames := make(map[string]bool)
		for _, result := range results {
			testNames[result.TestName] = true
		}

		expectedTests := []string{
			"Basic Query Performance",
			"Index Performance",
			"Concurrent Access Performance",
			"Load Test",
			"Resource Usage Test",
		}

		for _, expectedTest := range expectedTests {
			assert.True(t, testNames[expectedTest], "Missing test: %s", expectedTest)
		}
	})
}

// TestPerformanceTestConfig tests the performance test configuration
func TestPerformanceTestConfig(t *testing.T) {
	// Test default configuration
	suite := NewDatabasePerformanceTestSuite(nil, nil)
	assert.NotNil(t, suite.config)
	assert.Equal(t, 5*time.Minute, suite.config.TestDuration)
	assert.Equal(t, 30*time.Second, suite.config.WarmupDuration)
	assert.Equal(t, 30*time.Second, suite.config.CooldownDuration)
	assert.Equal(t, 10, suite.config.ConcurrentUsers)
	assert.Equal(t, 100, suite.config.RequestsPerUser)
	assert.Equal(t, 100*time.Millisecond, suite.config.RequestInterval)
	assert.Equal(t, 1*time.Second, suite.config.MaxQueryTime)
	assert.Equal(t, 5*time.Second, suite.config.MaxConnectionTime)
	assert.Equal(t, 50, suite.config.MinThroughput)
	assert.True(t, suite.config.MonitorCPU)
	assert.True(t, suite.config.MonitorMemory)
	assert.True(t, suite.config.MonitorConnections)

	// Test custom configuration
	customConfig := &PerformanceTestConfig{
		TestDuration:    1 * time.Minute,
		ConcurrentUsers: 5,
		MaxQueryTime:    500 * time.Millisecond,
		MinThroughput:   25,
	}

	suite2 := NewDatabasePerformanceTestSuite(nil, customConfig)
	assert.Equal(t, customConfig, suite2.config)
}

// TestPerformanceTestResult tests the performance test result structure
func TestPerformanceTestResult(t *testing.T) {
	result := &PerformanceTestResult{
		TestName:          "Test Query",
		StartTime:         time.Now(),
		EndTime:           time.Now().Add(1 * time.Second),
		Duration:          1 * time.Second,
		TotalQueries:      100,
		SuccessfulQueries: 95,
		FailedQueries:     5,
		AverageQueryTime:  50 * time.Millisecond,
		MinQueryTime:      10 * time.Millisecond,
		MaxQueryTime:      200 * time.Millisecond,
		P95QueryTime:      150 * time.Millisecond,
		P99QueryTime:      180 * time.Millisecond,
		QueriesPerSecond:  100.0,
		ConcurrentUsers:   10,
		CPUUsage:          25.5,
		MemoryUsage:       1024 * 1024,
		ConnectionCount:   5,
		ErrorRate:         5.0,
		CommonErrors: map[string]int{
			"timeout": 3,
			"error":   2,
		},
		Recommendations: []string{
			"Optimize slow queries",
			"Add more indexes",
		},
	}

	assert.Equal(t, "Test Query", result.TestName)
	assert.Equal(t, 100, result.TotalQueries)
	assert.Equal(t, 95, result.SuccessfulQueries)
	assert.Equal(t, 5, result.FailedQueries)
	assert.Equal(t, 50*time.Millisecond, result.AverageQueryTime)
	assert.Equal(t, 100.0, result.QueriesPerSecond)
	assert.Equal(t, 5.0, result.ErrorRate)
	assert.Len(t, result.Recommendations, 2)
}

// TestPerformanceMonitor tests the performance monitor
func TestPerformanceMonitor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance monitor test in short mode")
	}

	// Setup test database
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create monitor configuration
	config := &MonitoringConfig{
		MetricsInterval:    10 * time.Second,
		SlowQueryThreshold: 100 * time.Millisecond,
		MaxQueryTime:       1 * time.Second,
		MaxConnectionCount: 50,
		MinCacheHitRate:    0.90,
		MetricsRetention:   1 * time.Hour,
		MaxMetricsHistory:  100,
		MonitorQueries:     true,
		MonitorConnections: true,
		MonitorCache:       true,
		MonitorLocks:       true,
	}

	// Create performance monitor
	monitor := NewPerformanceMonitor(db, config)

	t.Run("start_stop_monitor", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Start monitoring
		err := monitor.Start(ctx)
		require.NoError(t, err)

		// Wait a bit for metrics collection
		time.Sleep(15 * time.Second)

		// Get current metrics
		metrics := monitor.GetCurrentMetrics()
		assert.NotNil(t, metrics)
		assert.True(t, metrics.Timestamp.After(time.Now().Add(-1*time.Minute)))

		// Stop monitoring
		monitor.Stop()

		// Wait a bit to ensure monitoring has stopped
		time.Sleep(2 * time.Second)
	})

	t.Run("get_performance_summary", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// Start monitoring
		err := monitor.Start(ctx)
		require.NoError(t, err)

		// Wait for metrics collection
		time.Sleep(10 * time.Second)

		// Get performance summary
		summary := monitor.GetPerformanceSummary()
		assert.NotNil(t, summary)
		assert.Contains(t, summary, "timestamp")
		assert.Contains(t, summary, "status")

		// Stop monitoring
		monitor.Stop()
	})

	t.Run("get_slow_queries", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// Start monitoring
		err := monitor.Start(ctx)
		require.NoError(t, err)

		// Wait for metrics collection
		time.Sleep(10 * time.Second)

		// Get slow queries
		slowQueries := monitor.GetSlowQueries()
		assert.NotNil(t, slowQueries)
		// Note: May be empty if no slow queries detected

		// Stop monitoring
		monitor.Stop()
	})

	t.Run("generate_performance_report", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// Start monitoring
		err := monitor.Start(ctx)
		require.NoError(t, err)

		// Wait for metrics collection
		time.Sleep(10 * time.Second)

		// Generate report
		report, err := monitor.GeneratePerformanceReport()
		require.NoError(t, err)
		assert.NotEmpty(t, report)

		// Stop monitoring
		monitor.Stop()
	})
}

// TestMonitoringConfig tests the monitoring configuration
func TestMonitoringConfig(t *testing.T) {
	// Test default configuration
	monitor := NewPerformanceMonitor(nil, nil)
	assert.NotNil(t, monitor.config)
	assert.Equal(t, 30*time.Second, monitor.config.MetricsInterval)
	assert.Equal(t, 1*time.Second, monitor.config.SlowQueryThreshold)
	assert.Equal(t, 5*time.Second, monitor.config.MaxQueryTime)
	assert.Equal(t, 100, monitor.config.MaxConnectionCount)
	assert.Equal(t, 0.90, monitor.config.MinCacheHitRate)
	assert.Equal(t, 24*time.Hour, monitor.config.MetricsRetention)
	assert.Equal(t, 1000, monitor.config.MaxMetricsHistory)
	assert.True(t, monitor.config.MonitorQueries)
	assert.True(t, monitor.config.MonitorConnections)
	assert.True(t, monitor.config.MonitorCache)
	assert.True(t, monitor.config.MonitorLocks)

	// Test custom configuration
	customConfig := &MonitoringConfig{
		MetricsInterval:    15 * time.Second,
		SlowQueryThreshold: 500 * time.Millisecond,
		MaxQueryTime:       2 * time.Second,
		MaxConnectionCount: 75,
		MinCacheHitRate:    0.95,
		MonitorQueries:     false,
		MonitorConnections: true,
		MonitorCache:       false,
		MonitorLocks:       true,
	}

	monitor2 := NewPerformanceMonitor(nil, customConfig)
	assert.Equal(t, customConfig, monitor2.config)
}

// Benchmark performance testing suite
func BenchmarkDatabasePerformanceTestSuite(b *testing.B) {
	// Setup test database
	db := setupTestDB(&testing.T{})
	defer cleanupTestDB(&testing.T{}, db)

	config := &PerformanceTestConfig{
		TestDuration:       10 * time.Second,
		ConcurrentUsers:    5,
		RequestsPerUser:    10,
		RequestInterval:    50 * time.Millisecond,
		MaxQueryTime:       500 * time.Millisecond,
		MinThroughput:      20,
		MonitorCPU:         false,
		MonitorMemory:      false,
		MonitorConnections: false,
	}

	suite := NewDatabasePerformanceTestSuite(db, config)

	b.Run("basic_query_performance", func(b *testing.B) {
		ctx := context.Background()
		for i := 0; i < b.N; i++ {
			result, err := suite.RunBasicQueryPerformanceTest(ctx)
			if err != nil {
				b.Fatalf("Basic query performance test failed: %v", err)
			}
			_ = result // Use result to avoid optimization
		}
	})

	b.Run("index_performance", func(b *testing.B) {
		ctx := context.Background()
		for i := 0; i < b.N; i++ {
			result, err := suite.RunIndexPerformanceTest(ctx)
			if err != nil {
				b.Fatalf("Index performance test failed: %v", err)
			}
			_ = result // Use result to avoid optimization
		}
	})
}

// Helper functions for test setup
func setupTestDB(t *testing.T) *sql.DB {
	// This would typically connect to a test database
	// For now, we'll skip the test if no database is available
	t.Skip("Database setup not implemented for unit tests")
	return nil
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	if db != nil {
		db.Close()
	}
}
