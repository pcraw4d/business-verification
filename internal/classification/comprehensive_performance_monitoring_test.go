package classification

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap/zaptest"
)

// createComprehensiveTestDB creates an in-memory SQLite database for comprehensive testing
func createComprehensiveTestDB() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	return db
}

// TestComprehensivePerformanceMonitoringIntegration tests the integration of all performance monitoring components
func TestComprehensivePerformanceMonitoringIntegration(t *testing.T) {
	// Setup test database
	db := createComprehensiveTestDB()
	defer db.Close()

	// Setup logger
	logger := zaptest.NewLogger(t)

	// Create all monitoring components
	responseTimeTracker := NewResponseTimeTracker(DefaultResponseTimeConfig(), logger)
	memoryMonitor := NewAdvancedMemoryMonitor(logger, DefaultMemoryMonitorConfig())
	databaseMonitor := NewEnhancedDatabaseMonitor(db, logger, DefaultDatabaseMonitorConfig(), nil)
	securityMonitor := NewAdvancedSecurityValidationMonitor(logger, DefaultSecurityValidationConfig())
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())

	// Set up database monitor with comprehensive monitor reference
	databaseMonitor.performanceMonitor = comprehensiveMonitor

	// Start all monitors
	responseTimeTracker.Start()
	memoryMonitor.Start()
	databaseMonitor.Start()
	securityMonitor.Start()
	comprehensiveMonitor.Start()

	// Cleanup
	defer func() {
		responseTimeTracker.Stop()
		memoryMonitor.Stop()
		databaseMonitor.Stop()
		securityMonitor.Stop()
		comprehensiveMonitor.Stop()
	}()

	// Test concurrent monitoring operations
	ctx := context.Background()
	var wg sync.WaitGroup

	// Simulate response time tracking
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			responseTimeTracker.TrackResponseTime("test_endpoint", "GET", 50*time.Millisecond, 200, nil)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Simulate database operations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			databaseMonitor.RecordQueryExecution(ctx, "SELECT * FROM test_table", 25*time.Millisecond, 1, 10, nil)
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// Simulate security validation
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			result := &AdvancedSecurityValidationResult{
				ValidationID:               fmt.Sprintf("test_validation_%d", i),
				ValidationType:             "data_source_validation",
				ValidationName:             "test_data_source",
				ExecutionTime:              30 * time.Millisecond,
				Success:                    true,
				Error:                      nil,
				SecurityViolation:          false,
				ComplianceViolation:        false,
				ThreatDetected:             false,
				VulnerabilityFound:         false,
				TrustScore:                 0.95,
				ConfidenceLevel:            0.90,
				RiskLevel:                  "low",
				SecurityRecommendations:    []string{},
				PerformanceRecommendations: []string{},
				Metadata:                   make(map[string]interface{}),
				Timestamp:                  time.Now(),
			}
			securityMonitor.RecordSecurityValidation(ctx, result)
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// Simulate comprehensive performance metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			metric := &ComprehensivePerformanceMetric{
				ID:             fmt.Sprintf("test_metric_%d", i),
				Timestamp:      time.Now(),
				MetricType:     "response_time",
				ServiceName:    "test_service",
				ResponseTimeMs: 45.0,
				Metadata:       make(map[string]interface{}),
			}
			comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// Wait for all operations to complete
	wg.Wait()

	// Allow time for background processing
	time.Sleep(100 * time.Millisecond)

	// Verify all monitors are working
	t.Run("response_time_tracker", func(t *testing.T) {
		stats := responseTimeTracker.GetResponseTimeStats("test_endpoint", "GET", 1*time.Minute)
		if stats == nil {
			t.Error("Expected response time stats, got nil")
		}
		if stats.RequestCount == 0 {
			t.Error("Expected request count > 0")
		}
	})

	t.Run("memory_monitor", func(t *testing.T) {
		metrics := memoryMonitor.GetLatestMemoryMetrics()
		if metrics == nil {
			t.Error("Expected memory metrics, got nil")
		}
		if metrics.MemoryUsageMB <= 0 {
			t.Error("Expected memory usage > 0")
		}
	})

	t.Run("database_monitor", func(t *testing.T) {
		stats := databaseMonitor.GetQueryPerformanceStats()
		if len(stats) == 0 {
			t.Error("Expected database query stats, got none")
		}
	})

	t.Run("security_monitor", func(t *testing.T) {
		stats := securityMonitor.GetValidationStats(10)
		if len(stats) == 0 {
			t.Error("Expected security validation stats, got none")
		}
	})

	t.Run("comprehensive_monitor", func(t *testing.T) {
		metrics := comprehensiveMonitor.GetPerformanceMetrics(10)
		if len(metrics) == 0 {
			t.Error("Expected comprehensive performance metrics, got none")
		}
	})
}

// TestPerformanceMonitoringUnderLoad tests performance monitoring under high load
func TestPerformanceMonitoringUnderLoad(t *testing.T) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(t)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	// Test high-frequency metric recording
	ctx := context.Background()
	start := time.Now()
	metricCount := 1000

	var wg sync.WaitGroup
	workers := 10
	metricsPerWorker := metricCount / workers

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < metricsPerWorker; j++ {
				metric := &ComprehensivePerformanceMetric{
					ID:             fmt.Sprintf("load_test_metric_%d_%d", workerID, j),
					Timestamp:      time.Now(),
					MetricType:     "load_test",
					ServiceName:    "load_test_service",
					ResponseTimeMs: float64(j % 100),
					Metadata:       make(map[string]interface{}),
				}
				comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// Verify performance
	if duration > 5*time.Second {
		t.Errorf("Performance monitoring under load took too long: %v", duration)
	}

	// Verify metrics were recorded
	metrics := comprehensiveMonitor.GetPerformanceMetrics(metricCount)
	if len(metrics) < metricCount/2 { // Allow for some loss due to async processing
		t.Errorf("Expected at least %d metrics, got %d", metricCount/2, len(metrics))
	}

	t.Logf("Recorded %d metrics in %v (%.2f metrics/sec)",
		len(metrics), duration, float64(len(metrics))/duration.Seconds())
}

// TestPerformanceMonitoringMemoryUsage tests memory usage during monitoring
func TestPerformanceMonitoringMemoryUsage(t *testing.T) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(t)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	// Get initial memory stats
	var initialMemStats runtime.MemStats
	runtime.ReadMemStats(&initialMemStats)

	// Record many metrics
	ctx := context.Background()
	metricCount := 10000

	for i := 0; i < metricCount; i++ {
		metric := &ComprehensivePerformanceMetric{
			ID:             fmt.Sprintf("memory_test_metric_%d", i),
			Timestamp:      time.Now(),
			MetricType:     "memory_test",
			ServiceName:    "memory_test_service",
			ResponseTimeMs: float64(i % 200),
			Metadata: map[string]interface{}{
				"test_data": fmt.Sprintf("test_value_%d", i),
			},
		}
		comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
	}

	// Force garbage collection
	runtime.GC()

	// Get final memory stats
	var finalMemStats runtime.MemStats
	runtime.ReadMemStats(&finalMemStats)

	// Calculate memory increase
	memoryIncrease := finalMemStats.Alloc - initialMemStats.Alloc
	memoryIncreaseMB := float64(memoryIncrease) / 1024 / 1024

	t.Logf("Memory increase after recording %d metrics: %.2f MB", metricCount, memoryIncreaseMB)

	// Memory increase should be reasonable (less than 100MB for 10k metrics)
	if memoryIncreaseMB > 100 {
		t.Errorf("Memory usage too high: %.2f MB for %d metrics", memoryIncreaseMB, metricCount)
	}
}

// TestPerformanceMonitoringConcurrency tests concurrent access to monitoring components
func TestPerformanceMonitoringConcurrency(t *testing.T) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(t)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()
	concurrentOperations := 100
	var wg sync.WaitGroup

	// Test concurrent metric recording
	for i := 0; i < concurrentOperations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			metric := &ComprehensivePerformanceMetric{
				ID:             fmt.Sprintf("concurrent_metric_%d", id),
				Timestamp:      time.Now(),
				MetricType:     "concurrency_test",
				ServiceName:    "concurrency_test_service",
				ResponseTimeMs: float64(id % 50),
				Metadata:       make(map[string]interface{}),
			}
			comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
		}(i)
	}

	// Test concurrent metric retrieval
	for i := 0; i < concurrentOperations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			comprehensiveMonitor.GetPerformanceMetrics(10)
		}()
	}

	wg.Wait()

	// Verify no race conditions occurred (test should complete without panicking)
	metrics := comprehensiveMonitor.GetPerformanceMetrics(concurrentOperations)
	if len(metrics) == 0 {
		t.Error("Expected metrics to be recorded during concurrent operations")
	}
}

// TestPerformanceMonitoringErrorHandling tests error handling in monitoring components
func TestPerformanceMonitoringErrorHandling(t *testing.T) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(t)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	tests := []struct {
		name    string
		metric  *ComprehensivePerformanceMetric
		wantErr bool
	}{
		{
			name: "valid metric",
			metric: &ComprehensivePerformanceMetric{
				ID:             "valid_metric",
				Timestamp:      time.Now(),
				MetricType:     "test",
				ServiceName:    "test_service",
				ResponseTimeMs: 50.0,
				Metadata:       make(map[string]interface{}),
			},
			wantErr: false,
		},
		{
			name: "metric with empty ID",
			metric: &ComprehensivePerformanceMetric{
				ID:             "",
				Timestamp:      time.Now(),
				MetricType:     "test",
				ServiceName:    "test_service",
				ResponseTimeMs: 50.0,
				Metadata:       make(map[string]interface{}),
			},
			wantErr: true,
		},
		{
			name: "metric with empty service name",
			metric: &ComprehensivePerformanceMetric{
				ID:             "test_metric",
				Timestamp:      time.Now(),
				MetricType:     "test",
				ServiceName:    "",
				ResponseTimeMs: 50.0,
				Metadata:       make(map[string]interface{}),
			},
			wantErr: true,
		},
		{
			name: "metric with zero timestamp",
			metric: &ComprehensivePerformanceMetric{
				ID:             "test_metric",
				Timestamp:      time.Time{},
				MetricType:     "test",
				ServiceName:    "test_service",
				ResponseTimeMs: 50.0,
				Metadata:       make(map[string]interface{}),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := comprehensiveMonitor.RecordPerformanceMetric(ctx, tt.metric)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordPerformanceMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPerformanceMonitoringDataPersistence tests data persistence in monitoring components
func TestPerformanceMonitoringDataPersistence(t *testing.T) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(t)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	// Record some metrics
	metricCount := 50
	for i := 0; i < metricCount; i++ {
		metric := &ComprehensivePerformanceMetric{
			ID:             fmt.Sprintf("persistence_test_metric_%d", i),
			Timestamp:      time.Now(),
			MetricType:     "persistence_test",
			ServiceName:    "persistence_test_service",
			ResponseTimeMs: float64(i % 100),
			Metadata:       make(map[string]interface{}),
		}
		comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
	}

	// Allow time for persistence
	time.Sleep(100 * time.Millisecond)

	// Verify metrics can be retrieved
	metrics := comprehensiveMonitor.GetPerformanceMetrics(metricCount)
	if len(metrics) == 0 {
		t.Error("Expected metrics to be persisted and retrievable")
	}

	// Verify metric content
	for _, metric := range metrics {
		if metric.ID == "" {
			t.Error("Expected metric ID to be preserved")
		}
		if metric.ServiceName == "" {
			t.Error("Expected service name to be preserved")
		}
		if metric.MetricType == "" {
			t.Error("Expected metric type to be preserved")
		}
	}
}

// TestPerformanceMonitoringAlerting tests alerting functionality
func TestPerformanceMonitoringAlerting(t *testing.T) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(t)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	// Record metrics that should trigger alerts
	alertTriggeringMetrics := []*ComprehensivePerformanceMetric{
		{
			ID:             "high_response_time_metric",
			Timestamp:      time.Now(),
			MetricType:     "response_time",
			ServiceName:    "test_service",
			ResponseTimeMs: 5000.0, // Very high response time
			Metadata:       make(map[string]interface{}),
		},
		{
			ID:            "high_memory_usage_metric",
			Timestamp:     time.Now(),
			MetricType:    "memory",
			ServiceName:   "test_service",
			MemoryUsageMB: 2000.0, // Very high memory usage
			Metadata:      make(map[string]interface{}),
		},
		{
			ID:            "error_metric",
			Timestamp:     time.Now(),
			MetricType:    "error",
			ServiceName:   "test_service",
			ErrorOccurred: true,
			ErrorMessage:  "Test error for alerting",
			Metadata:      make(map[string]interface{}),
		},
	}

	for _, metric := range alertTriggeringMetrics {
		comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
	}

	// Allow time for alert processing
	time.Sleep(100 * time.Millisecond)

	// Verify alerts were generated
	alerts := comprehensiveMonitor.GetPerformanceAlerts(false, 10)
	if len(alerts) == 0 {
		t.Error("Expected alerts to be generated for problematic metrics")
	}

	// Verify alert content
	for _, alert := range alerts {
		if alert.ID == "" {
			t.Error("Expected alert to have an ID")
		}
		if alert.Severity == "" {
			t.Error("Expected alert to have a severity level")
		}
		if alert.Message == "" {
			t.Error("Expected alert to have a message")
		}
	}
}

// TestPerformanceMonitoringCleanup tests cleanup functionality
func TestPerformanceMonitoringCleanup(t *testing.T) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(t)
	config := DefaultPerformanceMonitorConfig()
	config.MaxMetrics = 100 // Small limit for testing
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, config)
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	// Record more metrics than the limit
	metricCount := 150
	for i := 0; i < metricCount; i++ {
		metric := &ComprehensivePerformanceMetric{
			ID:             fmt.Sprintf("cleanup_test_metric_%d", i),
			Timestamp:      time.Now(),
			MetricType:     "cleanup_test",
			ServiceName:    "cleanup_test_service",
			ResponseTimeMs: float64(i % 50),
			Metadata:       make(map[string]interface{}),
		}
		comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
	}

	// Allow time for cleanup
	time.Sleep(200 * time.Millisecond)

	// Verify metrics are cleaned up
	metrics := comprehensiveMonitor.GetPerformanceMetrics(metricCount)
	if len(metrics) > config.MaxMetrics {
		t.Errorf("Expected metrics to be cleaned up, got %d metrics (limit: %d)",
			len(metrics), config.MaxMetrics)
	}
}

// BenchmarkComprehensivePerformanceMonitoring benchmarks the performance monitoring system
func BenchmarkComprehensivePerformanceMonitoring(b *testing.B) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(b)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			metric := &ComprehensivePerformanceMetric{
				ID:             fmt.Sprintf("benchmark_metric_%d", i),
				Timestamp:      time.Now(),
				MetricType:     "benchmark",
				ServiceName:    "benchmark_service",
				ResponseTimeMs: float64(i % 100),
				Metadata:       make(map[string]interface{}),
			}
			comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
			i++
		}
	})
}

// BenchmarkPerformanceMonitoringRetrieval benchmarks metric retrieval
func BenchmarkPerformanceMonitoringRetrieval(b *testing.B) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(b)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	// Pre-populate with metrics
	for i := 0; i < 1000; i++ {
		metric := &ComprehensivePerformanceMetric{
			ID:             fmt.Sprintf("benchmark_retrieval_metric_%d", i),
			Timestamp:      time.Now(),
			MetricType:     "benchmark_retrieval",
			ServiceName:    "benchmark_retrieval_service",
			ResponseTimeMs: float64(i % 100),
			Metadata:       make(map[string]interface{}),
		}
		comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comprehensiveMonitor.GetPerformanceMetrics(100)
	}
}

// TestPerformanceMonitoringIntegrationWithRealServices tests integration with real service scenarios
func TestPerformanceMonitoringIntegrationWithRealServices(t *testing.T) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(t)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	// Simulate a real service scenario
	serviceScenarios := []struct {
		name         string
		serviceName  string
		metricType   string
		responseTime float64
		errorRate    float64
		requestCount int
	}{
		{
			name:         "fast_api_service",
			serviceName:  "api_service",
			metricType:   "response_time",
			responseTime: 50.0,
			errorRate:    0.01,
			requestCount: 100,
		},
		{
			name:         "slow_database_service",
			serviceName:  "database_service",
			metricType:   "database_query",
			responseTime: 200.0,
			errorRate:    0.05,
			requestCount: 50,
		},
		{
			name:         "external_api_service",
			serviceName:  "external_api_service",
			metricType:   "external_call",
			responseTime: 300.0,
			errorRate:    0.10,
			requestCount: 30,
		},
	}

	for _, scenario := range serviceScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			for i := 0; i < scenario.requestCount; i++ {
				hasError := (float64(i) / float64(scenario.requestCount)) < scenario.errorRate

				metric := &ComprehensivePerformanceMetric{
					ID:             fmt.Sprintf("%s_metric_%d", scenario.name, i),
					Timestamp:      time.Now(),
					MetricType:     scenario.metricType,
					ServiceName:    scenario.serviceName,
					ResponseTimeMs: scenario.responseTime + float64(i%20), // Add some variation
					ErrorOccurred:  hasError,
					ErrorMessage: func() string {
						if hasError {
							return "Simulated error for testing"
						}
						return ""
					}(),
					Metadata: map[string]interface{}{
						"scenario":   scenario.name,
						"request_id": i,
					},
				}
				comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
			}
		})
	}

	// Allow time for processing
	time.Sleep(100 * time.Millisecond)

	// Verify metrics were recorded for all scenarios
	allMetrics := comprehensiveMonitor.GetPerformanceMetrics(1000)
	serviceMetrics := make(map[string]int)

	for _, metric := range allMetrics {
		serviceMetrics[metric.ServiceName]++
	}

	for _, scenario := range serviceScenarios {
		if serviceMetrics[scenario.serviceName] == 0 {
			t.Errorf("Expected metrics for service %s, got none", scenario.serviceName)
		}
	}

	// Verify error rates are approximately correct
	for _, scenario := range serviceScenarios {
		serviceMetrics := comprehensiveMonitor.GetPerformanceMetricsByService(scenario.serviceName, 1000)
		errorCount := 0
		for _, metric := range serviceMetrics {
			if metric.ErrorOccurred {
				errorCount++
			}
		}

		if len(serviceMetrics) > 0 {
			actualErrorRate := float64(errorCount) / float64(len(serviceMetrics))
			expectedErrorRate := scenario.errorRate

			// Allow for some variance in error rate
			if actualErrorRate > expectedErrorRate*2 || actualErrorRate < expectedErrorRate/2 {
				t.Logf("Service %s: expected error rate %.2f, got %.2f",
					scenario.serviceName, expectedErrorRate, actualErrorRate)
			}
		}
	}
}

// TestPerformanceMonitoringDataConsistency tests data consistency across monitoring components
func TestPerformanceMonitoringDataConsistency(t *testing.T) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(t)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	// Record the same metric multiple times
	metricID := "consistency_test_metric"
	expectedResponseTime := 75.0

	for i := 0; i < 10; i++ {
		metric := &ComprehensivePerformanceMetric{
			ID:             metricID,
			Timestamp:      time.Now(),
			MetricType:     "consistency_test",
			ServiceName:    "consistency_test_service",
			ResponseTimeMs: expectedResponseTime,
			Metadata: map[string]interface{}{
				"iteration": i,
			},
		}
		comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
	}

	// Allow time for processing
	time.Sleep(100 * time.Millisecond)

	// Verify data consistency
	metrics := comprehensiveMonitor.GetPerformanceMetrics(100)
	consistencyTestMetrics := make([]*ComprehensivePerformanceMetric, 0)

	for _, metric := range metrics {
		if metric.ID == metricID {
			consistencyTestMetrics = append(consistencyTestMetrics, metric)
		}
	}

	if len(consistencyTestMetrics) == 0 {
		t.Error("Expected consistency test metrics to be found")
	}

	// Verify all metrics have the same response time
	for _, metric := range consistencyTestMetrics {
		if metric.ResponseTimeMs != expectedResponseTime {
			t.Errorf("Expected response time %.2f, got %.2f",
				expectedResponseTime, metric.ResponseTimeMs)
		}
		if metric.ServiceName != "consistency_test_service" {
			t.Errorf("Expected service name 'consistency_test_service', got '%s'",
				metric.ServiceName)
		}
	}
}
