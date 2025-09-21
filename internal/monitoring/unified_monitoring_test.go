package monitoring

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// TestUnifiedMonitoringService tests the unified monitoring service
func TestUnifiedMonitoringService(t *testing.T) {
	// Skip if no database connection available
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database tests")
	}

	// Create a mock database connection for testing
	// In a real test, you would use a test database
	db, err := sql.Open("postgres", "postgres://test:test@localhost/test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available for testing")
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		t.Skip("Database not available for testing")
	}

	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := NewUnifiedMonitoringService(db, logger)

	ctx := context.Background()

	t.Run("RecordMetric", func(t *testing.T) {
		metric := &UnifiedMetric{
			ID:                uuid.New(),
			Timestamp:         time.Now(),
			Component:         "test",
			ComponentInstance: "test_instance",
			ServiceName:       "test_service",
			MetricType:        MetricTypePerformance,
			MetricCategory:    MetricCategoryLatency,
			MetricName:        "test_metric",
			MetricValue:       100.5,
			MetricUnit:        "ms",
			Tags: map[string]interface{}{
				"test_tag": "test_value",
			},
			Metadata: map[string]interface{}{
				"test_metadata": "test_value",
			},
			ConfidenceScore: 0.95,
			DataSource:      "test",
			CreatedAt:       time.Now(),
		}

		err := service.RecordMetric(ctx, metric)
		if err != nil {
			t.Errorf("Failed to record metric: %v", err)
		}
	})

	t.Run("RecordAlert", func(t *testing.T) {
		alert := &UnifiedAlert{
			ID:                uuid.New(),
			CreatedAt:         time.Now(),
			AlertType:         AlertTypeThreshold,
			AlertCategory:     AlertCategoryPerformance,
			Severity:          AlertSeverityWarning,
			Component:         "test",
			ComponentInstance: "test_instance",
			ServiceName:       "test_service",
			AlertName:         "test_alert",
			Description:       "Test alert description",
			Condition: map[string]interface{}{
				"metric_name": "test_metric",
				"operator":    ">",
				"threshold":   100.0,
			},
			CurrentValue:   &[]float64{150.0}[0],
			ThresholdValue: &[]float64{100.0}[0],
			Status:         AlertStatusActive,
			Tags: map[string]interface{}{
				"test_tag": "test_value",
			},
			Metadata: map[string]interface{}{
				"test_metadata": "test_value",
			},
		}

		err := service.RecordAlert(ctx, alert)
		if err != nil {
			t.Errorf("Failed to record alert: %v", err)
		}
	})

	t.Run("GetMetrics", func(t *testing.T) {
		filters := &MetricFilters{
			Component:   "test",
			ServiceName: "test_service",
			Limit:       10,
		}

		metrics, err := service.GetMetrics(ctx, filters)
		if err != nil {
			t.Errorf("Failed to get metrics: %v", err)
		}

		if len(metrics) == 0 {
			t.Log("No metrics found (this is expected for a clean test database)")
		} else {
			t.Logf("Found %d metrics", len(metrics))
		}
	})

	t.Run("GetAlerts", func(t *testing.T) {
		filters := &AlertFilters{
			Component: "test",
			Status:    AlertStatusActive,
			Limit:     10,
		}

		alerts, err := service.GetAlerts(ctx, filters)
		if err != nil {
			t.Errorf("Failed to get alerts: %v", err)
		}

		if len(alerts) == 0 {
			t.Log("No alerts found (this is expected for a clean test database)")
		} else {
			t.Logf("Found %d alerts", len(alerts))
		}
	})

	t.Run("GetMetricsSummary", func(t *testing.T) {
		endTime := time.Now()
		startTime := endTime.Add(-1 * time.Hour)

		summary, err := service.GetMetricsSummary(ctx, "test", "test_service", startTime, endTime)
		if err != nil {
			t.Errorf("Failed to get metrics summary: %v", err)
		}

		if summary == nil {
			t.Error("Metrics summary should not be nil")
		} else {
			t.Logf("Metrics summary: %+v", summary)
		}
	})

	t.Run("GetActiveAlertsCount", func(t *testing.T) {
		counts, err := service.GetActiveAlertsCount(ctx)
		if err != nil {
			t.Errorf("Failed to get active alerts count: %v", err)
		}

		if counts == nil {
			t.Error("Active alerts count should not be nil")
		} else {
			t.Logf("Active alerts count: %+v", counts)
		}
	})
}

// TestMonitoringAdapter tests the monitoring adapter
func TestMonitoringAdapter(t *testing.T) {
	// Skip if no database connection available
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database tests")
	}

	// Create a mock database connection for testing
	db, err := sql.Open("postgres", "postgres://test:test@localhost/test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available for testing")
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		t.Skip("Database not available for testing")
	}

	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	adapter := NewMonitoringAdapter(db, logger)

	ctx := context.Background()

	t.Run("RecordDatabaseMetrics", func(t *testing.T) {
		metrics := &DatabaseMetrics{
			Timestamp:         time.Now(),
			ConnectionCount:   10,
			ActiveConnections: 5,
			IdleConnections:   5,
			MaxConnections:    100,
			QueryCount:        1000,
			SlowQueryCount:    10,
			ErrorCount:        5,
			AvgQueryTime:      50.5,
			MaxQueryTime:      500.0,
			DatabaseSize:      1024 * 1024 * 100, // 100MB
			TableSizes: map[string]int64{
				"public.users": 1024 * 1024 * 10, // 10MB
			},
			IndexSizes: map[string]int64{
				"public.users_pkey": 1024 * 1024 * 2, // 2MB
			},
			LockCount:     2,
			DeadlockCount: 0,
			CacheHitRatio: 95.5,
			Uptime:        24 * time.Hour,
			Metadata: map[string]interface{}{
				"test_metadata": "test_value",
			},
		}

		err := adapter.RecordDatabaseMetrics(ctx, metrics)
		if err != nil {
			t.Errorf("Failed to record database metrics: %v", err)
		}
	})

	t.Run("RecordPerformanceMetric", func(t *testing.T) {
		metric := &PerformanceMetric{
			Name:      "response_time",
			Value:     150.5,
			Unit:      "ms",
			Timestamp: time.Now(),
			Tags: map[string]interface{}{
				"endpoint": "/api/test",
				"method":   "GET",
			},
			Metadata: map[string]interface{}{
				"test_metadata": "test_value",
			},
		}

		err := adapter.RecordPerformanceMetric(ctx, "api", "test_service", metric)
		if err != nil {
			t.Errorf("Failed to record performance metric: %v", err)
		}
	})

	t.Run("RecordAlert", func(t *testing.T) {
		condition := map[string]interface{}{
			"metric_name": "response_time",
			"operator":    ">",
			"threshold":   100.0,
		}

		err := adapter.RecordAlert(ctx, "api", "test_service", "high_response_time", "Response time is too high", AlertSeverityWarning, condition)
		if err != nil {
			t.Errorf("Failed to record alert: %v", err)
		}
	})

	t.Run("GetDatabaseMetricsSummary", func(t *testing.T) {
		endTime := time.Now()
		startTime := endTime.Add(-1 * time.Hour)

		summary, err := adapter.GetDatabaseMetricsSummary(ctx, startTime, endTime)
		if err != nil {
			t.Errorf("Failed to get database metrics summary: %v", err)
		}

		if summary == nil {
			t.Error("Database metrics summary should not be nil")
		} else {
			t.Logf("Database metrics summary: %+v", summary)
		}
	})

	t.Run("GetPerformanceMetrics", func(t *testing.T) {
		endTime := time.Now()
		startTime := endTime.Add(-1 * time.Hour)

		metrics, err := adapter.GetPerformanceMetrics(ctx, "api", "test_service", startTime, endTime, 10)
		if err != nil {
			t.Errorf("Failed to get performance metrics: %v", err)
		}

		if metrics == nil {
			t.Error("Performance metrics should not be nil")
		} else {
			t.Logf("Found %d performance metrics", len(metrics))
		}
	})

	t.Run("GetActiveAlerts", func(t *testing.T) {
		alerts, err := adapter.GetActiveAlerts(ctx, "api")
		if err != nil {
			t.Errorf("Failed to get active alerts: %v", err)
		}

		if alerts == nil {
			t.Error("Active alerts should not be nil")
		} else {
			t.Logf("Found %d active alerts", len(alerts))
		}
	})
}

// BenchmarkUnifiedMonitoringService benchmarks the unified monitoring service
func BenchmarkUnifiedMonitoringService(b *testing.B) {
	// Skip if no database connection available
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		b.Skip("Skipping database tests")
	}

	// Create a mock database connection for testing
	db, err := sql.Open("postgres", "postgres://test:test@localhost/test?sslmode=disable")
	if err != nil {
		b.Skip("Database not available for testing")
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		b.Skip("Database not available for testing")
	}

	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)
	service := NewUnifiedMonitoringService(db, logger)

	ctx := context.Background()

	b.Run("RecordMetric", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			metric := &UnifiedMetric{
				ID:                uuid.New(),
				Timestamp:         time.Now(),
				Component:         "benchmark",
				ComponentInstance: "benchmark_instance",
				ServiceName:       "benchmark_service",
				MetricType:        MetricTypePerformance,
				MetricCategory:    MetricCategoryLatency,
				MetricName:        "benchmark_metric",
				MetricValue:       float64(i),
				MetricUnit:        "ms",
				Tags: map[string]interface{}{
					"iteration": i,
				},
				Metadata: map[string]interface{}{
					"benchmark": true,
				},
				ConfidenceScore: 0.95,
				DataSource:      "benchmark",
				CreatedAt:       time.Now(),
			}

			err := service.RecordMetric(ctx, metric)
			if err != nil {
				b.Errorf("Failed to record metric: %v", err)
			}
		}
	})

	b.Run("GetMetrics", func(b *testing.B) {
		filters := &MetricFilters{
			Component:   "benchmark",
			ServiceName: "benchmark_service",
			Limit:       100,
		}

		for i := 0; i < b.N; i++ {
			_, err := service.GetMetrics(ctx, filters)
			if err != nil {
				b.Errorf("Failed to get metrics: %v", err)
			}
		}
	})
}
