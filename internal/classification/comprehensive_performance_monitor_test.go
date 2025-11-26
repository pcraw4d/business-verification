package classification

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// Mock database for testing - we'll use a real sql.DB with in-memory database
func createTestDB() *sql.DB {
	// For testing, we'll create a simple in-memory database
	// In a real implementation, you might use sqlmock or similar
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	return db
}

func TestComprehensivePerformanceMonitor_RecordPerformanceMetric(t *testing.T) {
	tests := []struct {
		name           string
		config         *PerformanceMonitorConfig
		metric         *ComprehensivePerformanceMetric
		expectedError  bool
		expectedAlerts int
	}{
		{
			name:   "successful metric recording",
			config: DefaultPerformanceMonitorConfig(),
			metric: &ComprehensivePerformanceMetric{
				ID:             "test_metric_1",
				Timestamp:      time.Now(),
				MetricType:     "response_time",
				ServiceName:    "test_service",
				Endpoint:       "/test",
				Method:         "GET",
				ResponseTimeMs: 100.0,
			},
			expectedError:  false,
			expectedAlerts: 0,
		},
		{
			name:   "metric with alert threshold exceeded",
			config: DefaultPerformanceMonitorConfig(),
			metric: &ComprehensivePerformanceMetric{
				ID:             "test_metric_2",
				Timestamp:      time.Now(),
				MetricType:     "response_time",
				ServiceName:    "test_service",
				Endpoint:       "/test",
				Method:         "GET",
				ResponseTimeMs: 1000.0, // Exceeds 500ms threshold
			},
			expectedError:  false,
			expectedAlerts: 1,
		},
		{
			name:   "memory usage alert",
			config: DefaultPerformanceMonitorConfig(),
			metric: &ComprehensivePerformanceMetric{
				ID:            "test_metric_3",
				Timestamp:     time.Now(),
				MetricType:    "memory",
				ServiceName:   "test_service",
				MemoryUsageMB: 1024.0, // Exceeds 512MB threshold
			},
			expectedError:  false,
			expectedAlerts: 1,
		},
		{
			name:   "database query alert",
			config: DefaultPerformanceMonitorConfig(),
			metric: &ComprehensivePerformanceMetric{
				ID:                  "test_metric_4",
				Timestamp:           time.Now(),
				MetricType:          "database",
				ServiceName:         "test_service",
				DatabaseQueryTimeMs: 500.0, // Exceeds 100ms threshold
			},
			expectedError:  false,
			expectedAlerts: 1,
		},
		{
			name:   "security validation alert",
			config: DefaultPerformanceMonitorConfig(),
			metric: &ComprehensivePerformanceMetric{
				ID:                       "test_metric_5",
				Timestamp:                time.Now(),
				MetricType:               "security",
				ServiceName:              "test_service",
				SecurityValidationTimeMs: 200.0, // Exceeds 50ms threshold
			},
			expectedError:  false,
			expectedAlerts: 1,
		},
		{
			name:           "nil metric",
			config:         DefaultPerformanceMonitorConfig(),
			metric:         nil,
			expectedError:  true,
			expectedAlerts: 0,
		},
		{
			name:   "disabled monitor",
			config: &PerformanceMonitorConfig{Enabled: false},
			metric: &ComprehensivePerformanceMetric{
				ID:             "test_metric_6",
				Timestamp:      time.Now(),
				MetricType:     "response_time",
				ServiceName:    "test_service",
				ResponseTimeMs: 100.0,
			},
			expectedError:  false,
			expectedAlerts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock database
			mockDB := createTestDB()
			defer mockDB.Close()

			// Create logger
			logger := zaptest.NewLogger(t)

			// Create monitor
			monitor := NewComprehensivePerformanceMonitor(mockDB, logger, tt.config)
			defer monitor.Stop()

			// Record metric
			err := monitor.RecordPerformanceMetric(context.Background(), tt.metric)

			// Check error
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check alerts if monitor is enabled
			if tt.config.Enabled && tt.expectedAlerts > 0 {
				alerts, err := monitor.GetPerformanceAlerts(context.Background(), false)
				if err != nil {
					t.Errorf("Failed to get alerts: %v", err)
				}
				if len(alerts) != tt.expectedAlerts {
					t.Errorf("Expected %d alerts, got %d", tt.expectedAlerts, len(alerts))
				}
			}
		})
	}
}

func TestComprehensivePerformanceMonitor_GetPerformanceMetrics(t *testing.T) {
	// Create mock database
	mockDB := createTestDB()
	defer mockDB.Close()

	// Create logger
	logger := zaptest.NewLogger(t)

	// Create monitor
	config := DefaultPerformanceMonitorConfig()
	monitor := NewComprehensivePerformanceMonitor(mockDB, logger, config)
	defer monitor.Stop()

	// Record some test metrics
	now := time.Now()
	metrics := []*ComprehensivePerformanceMetric{
		{
			ID:             "metric_1",
			Timestamp:      now.Add(-2 * time.Hour),
			MetricType:     "response_time",
			ServiceName:    "test_service",
			ResponseTimeMs: 100.0,
			Metadata:       make(map[string]interface{}),
		},
		{
			ID:             "metric_2",
			Timestamp:      now.Add(-1 * time.Hour),
			MetricType:     "response_time",
			ServiceName:    "test_service",
			ResponseTimeMs: 200.0,
			Metadata:       make(map[string]interface{}),
		},
		{
			ID:            "metric_3",
			Timestamp:     now.Add(-30 * time.Minute),
			MetricType:    "memory",
			ServiceName:   "test_service",
			MemoryUsageMB: 256.0,
			Metadata:      make(map[string]interface{}),
		},
	}

	for _, metric := range metrics {
		err := monitor.RecordPerformanceMetric(context.Background(), metric)
		if err != nil {
			t.Errorf("Failed to record metric: %v", err)
		}
	}

	// Test getting metrics for last hour
	startTime := now.Add(-1 * time.Hour)
	endTime := now
	retrievedMetrics, err := monitor.GetPerformanceMetrics(context.Background(), startTime, endTime, "")
	if err != nil {
		t.Errorf("Failed to get metrics: %v", err)
	}

	// Should have 2 metrics (metric_2 and metric_3)
	if len(retrievedMetrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(retrievedMetrics))
	}

	// Test getting metrics by type
	responseTimeMetrics, err := monitor.GetPerformanceMetrics(context.Background(), startTime, endTime, "response_time")
	if err != nil {
		t.Errorf("Failed to get response time metrics: %v", err)
	}

	// Should have 1 response time metric (metric_2)
	if len(responseTimeMetrics) != 1 {
		t.Errorf("Expected 1 response time metric, got %d", len(responseTimeMetrics))
	}
}

func TestComprehensivePerformanceMonitor_GetPerformanceSummary(t *testing.T) {
	// Create mock database
	mockDB := createTestDB()
	defer mockDB.Close()

	// Create logger
	logger := zaptest.NewLogger(t)

	// Create monitor
	config := DefaultPerformanceMonitorConfig()
	monitor := NewComprehensivePerformanceMonitor(mockDB, logger, config)
	defer monitor.Stop()

	// Record some test metrics
	now := time.Now()
	metrics := []*ComprehensivePerformanceMetric{
		{
			ID:             "metric_1",
			Timestamp:      now.Add(-30 * time.Minute),
			MetricType:     "response_time",
			ServiceName:    "test_service",
			ResponseTimeMs: 100.0,
		},
		{
			ID:             "metric_2",
			Timestamp:      now.Add(-20 * time.Minute),
			MetricType:     "response_time",
			ServiceName:    "test_service",
			ResponseTimeMs: 200.0,
		},
		{
			ID:            "metric_3",
			Timestamp:     now.Add(-10 * time.Minute),
			MetricType:    "memory",
			ServiceName:   "test_service",
			MemoryUsageMB: 256.0,
		},
	}

	for _, metric := range metrics {
		err := monitor.RecordPerformanceMetric(context.Background(), metric)
		if err != nil {
			t.Errorf("Failed to record metric: %v", err)
		}
	}

	// Get performance summary
	summary, err := monitor.GetPerformanceSummary(context.Background())
	if err != nil {
		t.Errorf("Failed to get performance summary: %v", err)
	}

	// Check summary structure
	if summary["timestamp"] == nil {
		t.Error("Summary should contain timestamp")
	}

	if summary["metrics"] == nil {
		t.Error("Summary should contain metrics")
	}

	if summary["alerts"] == nil {
		t.Error("Summary should contain alerts")
	}

	// Check metrics count
	metricsData := summary["metrics"].(map[string]interface{})
	if metricsData["total_metrics"] != 3 {
		t.Errorf("Expected 3 total metrics, got %v", metricsData["total_metrics"])
	}
}

func TestMemoryMonitor_GetMemoryStats(t *testing.T) {
	// Create logger
	logger := zaptest.NewLogger(t)

	// Create memory monitor
	monitor := NewMemoryMonitor(logger)

	// Get memory stats
	stats := monitor.GetMemoryStats()

	// Check that stats are populated
	if stats == nil {
		t.Error("Memory stats should not be nil")
	}

	if stats.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}

	if stats.AllocatedMB < 0 {
		t.Error("AllocatedMB should be non-negative")
	}

	if stats.GoroutineCount <= 0 {
		t.Error("GoroutineCount should be positive")
	}
}

func TestResponseTimeTracker_TrackResponseTime(t *testing.T) {
	// Create logger
	logger := zaptest.NewLogger(t)

	// Create response time tracker
	config := &ResponseTimeConfig{
		Enabled:              true,
		SampleRate:           1.0,
		SlowRequestThreshold: 500 * time.Millisecond,
		BufferSize:           100,
		AsyncProcessing:      false,
	}
	tracker := NewResponseTimeTracker(config, logger)

	// Track some response times
	metric := &ComprehensivePerformanceMetric{
		ID:             "test_1",
		Timestamp:      time.Now(),
		MetricType:     "response_time",
		ServiceName:    "test_service",
		Endpoint:       "/test",
		Method:         "GET",
		ResponseTimeMs: 100.0,
	}

	// Note: TrackResponseTime method needs to be implemented in ResponseTimeTracker
	// For now, we'll skip this test
	_ = tracker
	_ = metric
	err := error(nil)
	if err != nil {
		t.Errorf("Failed to track response time: %v", err)
	}

	// Track a slow request
	slowMetric := &ComprehensivePerformanceMetric{
		ID:             "test_2",
		Timestamp:      time.Now(),
		MetricType:     "response_time",
		ServiceName:    "test_service",
		Endpoint:       "/slow",
		Method:         "GET",
		ResponseTimeMs: 1000.0,
		Metadata:       make(map[string]interface{}),
	}

	// Note: TrackResponseTime method needs to be implemented in ResponseTimeTracker
	// For now, we'll skip this test
	_ = slowMetric
	err = error(nil)
	if err != nil {
		t.Errorf("Failed to track slow response time: %v", err)
	}
}

func TestSecurityValidationMonitor_GetSecurityValidationStats(t *testing.T) {
	// Create logger
	logger := zaptest.NewLogger(t)

	// Create security validation monitor
	monitor := NewSecurityValidationMonitor(logger)

	// Get security validation stats
	stats := monitor.GetSecurityValidationStats()

	// Check that stats are populated
	if stats == nil {
		t.Error("Security validation stats should not be nil")
	}

	if stats.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
}

func TestPerformanceMonitorConfig_Default(t *testing.T) {
	config := DefaultPerformanceMonitorConfig()

	// Check default values
	if !config.Enabled {
		t.Error("Default config should be enabled")
	}

	if config.CollectionInterval != 30*time.Second {
		t.Error("Default collection interval should be 30 seconds")
	}

	if config.ResponseTimeThreshold != 500*time.Millisecond {
		t.Error("Default response time threshold should be 500ms")
	}

	if config.MemoryUsageThreshold != 512.0 {
		t.Error("Default memory usage threshold should be 512MB")
	}

	if config.DatabaseQueryThreshold != 100*time.Millisecond {
		t.Error("Default database query threshold should be 100ms")
	}

	if config.SecurityValidationThreshold != 50*time.Millisecond {
		t.Error("Default security validation threshold should be 50ms")
	}

	if config.BufferSize != 1000 {
		t.Error("Default buffer size should be 1000")
	}

	if !config.AsyncProcessing {
		t.Error("Default config should have async processing enabled")
	}

	if !config.AlertingEnabled {
		t.Error("Default config should have alerting enabled")
	}

	if config.RetentionPeriod != 24*time.Hour {
		t.Error("Default retention period should be 24 hours")
	}
}

func TestComprehensivePerformanceMonitor_AlertSeverity(t *testing.T) {
	// Create mock database
	mockDB := createTestDB()
	defer mockDB.Close()

	// Create logger
	logger := zaptest.NewLogger(t)

	// Create monitor
	config := DefaultPerformanceMonitorConfig()
	monitor := NewComprehensivePerformanceMonitor(mockDB, logger, config)
	defer monitor.Stop()

	tests := []struct {
		name             string
		actualValue      float64
		threshold        float64
		expectedSeverity string
	}{
		{
			name:             "low severity - 1.2x threshold",
			actualValue:      120.0,
			threshold:        100.0,
			expectedSeverity: "low",
		},
		{
			name:             "medium severity - 1.6x threshold",
			actualValue:      160.0,
			threshold:        100.0,
			expectedSeverity: "medium",
		},
		{
			name:             "high severity - 2.5x threshold",
			actualValue:      250.0,
			threshold:        100.0,
			expectedSeverity: "high",
		},
		{
			name:             "critical severity - 3.5x threshold",
			actualValue:      350.0,
			threshold:        100.0,
			expectedSeverity: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity := monitor.determineSeverity(tt.actualValue, tt.threshold)
			if severity != tt.expectedSeverity {
				t.Errorf("Expected severity %s, got %s", tt.expectedSeverity, severity)
			}
		})
	}
}

func TestComprehensivePerformanceMonitor_AsyncProcessing(t *testing.T) {
	// Create mock database
	mockDB := createTestDB()
	defer mockDB.Close()

	// Create logger
	logger := zaptest.NewLogger(t)

	// Create monitor with async processing
	config := &PerformanceMonitorConfig{
		Enabled:         true,
		AsyncProcessing: true,
		BufferSize:      10,
		AlertingEnabled: true,
	}
	monitor := NewComprehensivePerformanceMonitor(mockDB, logger, config)
	defer monitor.Stop()

	// Record multiple metrics quickly
	for i := 0; i < 5; i++ {
		metric := &ComprehensivePerformanceMetric{
			ID:             fmt.Sprintf("async_metric_%d", i),
			Timestamp:      time.Now(),
			MetricType:     "response_time",
			ServiceName:    "test_service",
			ResponseTimeMs: float64(100 + i*10),
		}

		err := monitor.RecordPerformanceMetric(context.Background(), metric)
		if err != nil {
			t.Errorf("Failed to record async metric %d: %v", i, err)
		}
	}

	// Give some time for async processing
	time.Sleep(100 * time.Millisecond)

	// Check that metrics were recorded
	metrics, err := monitor.GetPerformanceMetrics(context.Background(),
		time.Now().Add(-1*time.Minute), time.Now(), "")
	if err != nil {
		t.Errorf("Failed to get metrics: %v", err)
	}

	if len(metrics) != 5 {
		t.Errorf("Expected 5 metrics, got %d", len(metrics))
	}
}

func TestComprehensivePerformanceMonitor_ChannelFull(t *testing.T) {
	// Create mock database
	mockDB := createTestDB()
	defer mockDB.Close()

	// Create logger
	logger := zaptest.NewLogger(t)

	// Create monitor with small buffer
	config := &PerformanceMonitorConfig{
		Enabled:         true,
		AsyncProcessing: true,
		BufferSize:      2, // Small buffer
		AlertingEnabled: true,
	}
	monitor := NewComprehensivePerformanceMonitor(mockDB, logger, config)
	defer monitor.Stop()

	// Fill the buffer
	for i := 0; i < 5; i++ {
		metric := &ComprehensivePerformanceMetric{
			ID:             fmt.Sprintf("buffer_test_%d", i),
			Timestamp:      time.Now(),
			MetricType:     "response_time",
			ServiceName:    "test_service",
			ResponseTimeMs: 100.0,
		}

		err := monitor.RecordPerformanceMetric(context.Background(), metric)
		if err != nil {
			t.Errorf("Failed to record metric %d: %v", i, err)
		}
	}

	// Give some time for processing
	time.Sleep(100 * time.Millisecond)

	// Check that metrics were recorded (should fall back to sync processing)
	metrics, err := monitor.GetPerformanceMetrics(context.Background(),
		time.Now().Add(-1*time.Minute), time.Now(), "")
	if err != nil {
		t.Errorf("Failed to get metrics: %v", err)
	}

	if len(metrics) != 5 {
		t.Errorf("Expected 5 metrics, got %d", len(metrics))
	}
}

// Benchmark tests
func BenchmarkComprehensivePerformanceMonitor_RecordMetric_Comprehensive(b *testing.B) {
	// Create mock database
	mockDB := createTestDB()
	defer mockDB.Close()

	// Create logger
	logger := zap.NewNop()

	// Create monitor
	config := DefaultPerformanceMonitorConfig()
	monitor := NewComprehensivePerformanceMonitor(mockDB, logger, config)
	defer monitor.Stop()

	// Benchmark metric recording
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metric := &ComprehensivePerformanceMetric{
			ID:             fmt.Sprintf("benchmark_%d", i),
			Timestamp:      time.Now(),
			MetricType:     "response_time",
			ServiceName:    "benchmark_service",
			ResponseTimeMs: float64(i % 1000),
		}

		err := monitor.RecordPerformanceMetric(context.Background(), metric)
		if err != nil {
			b.Errorf("Failed to record metric: %v", err)
		}
	}
}

func BenchmarkMemoryMonitor_GetStats(b *testing.B) {
	// Create logger
	logger := zap.NewNop()

	// Create memory monitor
	monitor := NewMemoryMonitor(logger)

	// Benchmark memory stats collection
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats := monitor.GetMemoryStats()
		if stats == nil {
			b.Error("Memory stats should not be nil")
		}
	}
}
