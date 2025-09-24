package classification

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap/zaptest"
)

// createSimpleTestDB creates an in-memory SQLite database for testing
func createSimpleTestDB() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	return db
}

// TestComprehensivePerformanceMonitor_BasicFunctionality tests basic functionality of the comprehensive performance monitor
func TestComprehensivePerformanceMonitor_BasicFunctionality(t *testing.T) {
	// Setup test database
	db := createSimpleTestDB()
	defer db.Close()

	// Setup logger
	logger := zaptest.NewLogger(t)

	// Create comprehensive performance monitor
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	// Test recording a performance metric
	metric := &ComprehensivePerformanceMetric{
		ID:             "test_metric_1",
		Timestamp:      time.Now(),
		MetricType:     "response_time",
		ServiceName:    "test_service",
		ResponseTimeMs: 50.0,
		Metadata:       make(map[string]interface{}),
	}

	err := comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
	if err != nil {
		t.Errorf("Expected no error recording metric, got: %v", err)
	}

	// Test retrieving performance metrics
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now().Add(1 * time.Hour)
	metrics, err := comprehensiveMonitor.GetPerformanceMetrics(ctx, startTime, endTime, "response_time")
	if err != nil {
		t.Errorf("Expected no error retrieving metrics, got: %v", err)
	}

	if len(metrics) == 0 {
		t.Error("Expected to find recorded metrics")
	}

	// Verify metric content
	found := false
	for _, m := range metrics {
		if m.ID == "test_metric_1" {
			found = true
			if m.ServiceName != "test_service" {
				t.Errorf("Expected service name 'test_service', got '%s'", m.ServiceName)
			}
			if m.ResponseTimeMs != 50.0 {
				t.Errorf("Expected response time 50.0, got %.2f", m.ResponseTimeMs)
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find the recorded metric")
	}
}

// TestComprehensivePerformanceMonitor_ErrorHandling tests error handling
func TestComprehensivePerformanceMonitor_ErrorHandling(t *testing.T) {
	db := createSimpleTestDB()
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

// TestComprehensivePerformanceMonitor_ConcurrentAccess tests concurrent access
func TestComprehensivePerformanceMonitor_ConcurrentAccess(t *testing.T) {
	db := createSimpleTestDB()
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
			startTime := time.Now().Add(-1 * time.Hour)
			endTime := time.Now().Add(1 * time.Hour)
			comprehensiveMonitor.GetPerformanceMetrics(ctx, startTime, endTime, "concurrency_test")
		}()
	}

	wg.Wait()

	// Verify no race conditions occurred (test should complete without panicking)
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now().Add(1 * time.Hour)
	metrics, err := comprehensiveMonitor.GetPerformanceMetrics(ctx, startTime, endTime, "concurrency_test")
	if err != nil {
		t.Errorf("Expected no error retrieving metrics, got: %v", err)
	}

	if len(metrics) == 0 {
		t.Error("Expected metrics to be recorded during concurrent operations")
	}
}

// TestAdvancedMemoryMonitor_BasicFunctionality tests basic functionality of the advanced memory monitor
func TestAdvancedMemoryMonitor_BasicFunctionality(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := DefaultMemoryMonitorConfig()
	monitor := NewAdvancedMemoryMonitor(logger, config)
	defer monitor.Stop()

	// Test memory metrics collection
	monitor.collectMemoryStats()
	metrics := monitor.GetCurrentStats()

	if metrics == nil {
		t.Error("Expected memory metrics to be collected")
	}

	if metrics.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	if metrics.AllocatedMB <= 0 {
		t.Error("Expected memory usage > 0")
	}

	if metrics.GoroutineCount <= 0 {
		t.Error("Expected goroutine count > 0")
	}
}

// TestAdvancedMemoryMonitor_StartStop tests start/stop functionality
func TestAdvancedMemoryMonitor_StartStop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := DefaultMemoryMonitorConfig()
	monitor := NewAdvancedMemoryMonitor(logger, config)

	// Test starting
	monitor.Start()
	time.Sleep(50 * time.Millisecond) // Give it a moment to start

	// Test stopping
	monitor.Stop()
	time.Sleep(50 * time.Millisecond) // Give it a moment to stop

	// Verify that the monitor stopped without panicking
	// Further checks could involve inspecting logs or mock channels if they were used
}

// TestAdvancedSecurityValidationMonitor_BasicFunctionality tests basic functionality of the security validation monitor
func TestAdvancedSecurityValidationMonitor_BasicFunctionality(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := DefaultSecurityValidationConfig()
	monitor := NewAdvancedSecurityValidationMonitor(logger, config)
	defer monitor.Stop()

	ctx := context.Background()

	// Test recording a security validation result
	result := &AdvancedSecurityValidationResult{
		ValidationID:               "test_validation_1",
		ValidationType:             "data_source_validation",
		ValidationName:             "test_data_source",
		ExecutionTime:              50 * time.Millisecond,
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

	monitor.RecordSecurityValidation(ctx, result)

	// Test retrieving validation stats
	stats := monitor.GetValidationStats(10)
	if len(stats) == 0 {
		t.Error("Expected validation stats to be recorded")
	}

	// Verify stats content
	found := false
	for _, stat := range stats {
		if stat.ValidationID == "test_validation_1" {
			found = true
			if stat.ValidationType != "data_source_validation" {
				t.Errorf("Expected validation type 'data_source_validation', got '%s'", stat.ValidationType)
			}
			if stat.ExecutionCount != 1 {
				t.Errorf("Expected execution count 1, got %d", stat.ExecutionCount)
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find the recorded validation stats")
	}
}

// TestAdvancedSecurityValidationMonitor_StartStop tests start/stop functionality
func TestAdvancedSecurityValidationMonitor_StartStop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := DefaultSecurityValidationConfig()
	monitor := NewAdvancedSecurityValidationMonitor(logger, config)

	// Test starting
	monitor.Start()
	time.Sleep(50 * time.Millisecond) // Give it a moment to start

	// Test stopping
	monitor.Stop()
	time.Sleep(50 * time.Millisecond) // Give it a moment to stop

	// Verify that the monitor stopped without panicking
}

// TestEnhancedDatabaseMonitor_BasicFunctionality tests basic functionality of the enhanced database monitor
func TestEnhancedDatabaseMonitor_BasicFunctionality(t *testing.T) {
	db := createSimpleTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(t)
	config := DefaultEnhancedDatabaseConfig()
	monitor := NewEnhancedDatabaseMonitor(db, logger, config)
	defer monitor.Stop()

	ctx := context.Background()

	// Test recording a query execution
	monitor.RecordQueryExecution(ctx, "SELECT * FROM test_table", 25*time.Millisecond, 1, 10, false, "test_query")

	// Test retrieving query performance stats
	stats := monitor.GetQueryStats(10)
	if len(stats) == 0 {
		t.Error("Expected query performance stats to be recorded")
	}

	// Verify stats content
	found := false
	for queryText, stat := range stats {
		if queryText == "SELECT * FROM test_table" {
			found = true
			if stat.ExecutionCount != 1 {
				t.Errorf("Expected execution count 1, got %d", stat.ExecutionCount)
			}
			if stat.AverageExecutionTime != 25.0 {
				t.Errorf("Expected average execution time 25.0, got %.2f", stat.AverageExecutionTime)
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find the recorded query stats")
	}
}

// TestEnhancedDatabaseMonitor_StartStop tests start/stop functionality
func TestEnhancedDatabaseMonitor_StartStop(t *testing.T) {
	db := createSimpleTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(t)
	config := DefaultEnhancedDatabaseConfig()
	monitor := NewEnhancedDatabaseMonitor(db, logger, config)

	// Test starting
	monitor.Start()
	time.Sleep(50 * time.Millisecond) // Give it a moment to start

	// Test stopping
	monitor.Stop()
	time.Sleep(50 * time.Millisecond) // Give it a moment to stop

	// Verify that the monitor stopped without panicking
}

// BenchmarkComprehensivePerformanceMonitor_RecordMetric benchmarks metric recording
func BenchmarkComprehensivePerformanceMonitor_RecordMetric(b *testing.B) {
	db := createSimpleTestDB()
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

// BenchmarkComprehensivePerformanceMonitor_GetMetrics benchmarks metric retrieval
func BenchmarkComprehensivePerformanceMonitor_GetMetrics(b *testing.B) {
	db := createSimpleTestDB()
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
		startTime := time.Now().Add(-1 * time.Hour)
		endTime := time.Now().Add(1 * time.Hour)
		comprehensiveMonitor.GetPerformanceMetrics(ctx, startTime, endTime, "benchmark_retrieval")
	}
}

// BenchmarkAdvancedMemoryMonitor_CollectMetrics benchmarks memory metrics collection
func BenchmarkAdvancedMemoryMonitor_CollectMetrics(b *testing.B) {
	logger := zaptest.NewLogger(b)
	monitor := NewAdvancedMemoryMonitor(logger, DefaultMemoryMonitorConfig())
	defer monitor.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.collectMemoryStats()
	}
}

// BenchmarkAdvancedSecurityValidationMonitor_RecordValidation benchmarks security validation recording
func BenchmarkAdvancedSecurityValidationMonitor_RecordValidation(b *testing.B) {
	logger := zaptest.NewLogger(b)
	monitor := NewAdvancedSecurityValidationMonitor(logger, DefaultSecurityValidationConfig())
	defer monitor.Stop()

	ctx := context.Background()
	result := &AdvancedSecurityValidationResult{
		ValidationID:               "benchmark_validation",
		ValidationType:             "data_source_validation",
		ValidationName:             "benchmark_data_source",
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

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			result.ValidationID = fmt.Sprintf("benchmark_validation_%d", i)
			monitor.RecordSecurityValidation(ctx, result)
			i++
		}
	})
}

// BenchmarkEnhancedDatabaseMonitor_RecordQuery benchmarks database query recording
func BenchmarkEnhancedDatabaseMonitor_RecordQuery(b *testing.B) {
	db := createSimpleTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(b)
	monitor := NewEnhancedDatabaseMonitor(db, logger, DefaultEnhancedDatabaseConfig())
	defer monitor.Stop()

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			monitor.RecordQueryExecution(ctx, "SELECT * FROM test_table", 25*time.Millisecond, 1, 10, false, "test_query")
			i++
		}
	})
}
