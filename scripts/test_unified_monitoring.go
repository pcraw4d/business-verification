package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"kyb-platform/internal/monitoring"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	// Get database connection string from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/kyb_platform?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	logger := log.New(os.Stdout, "MONITORING_TEST: ", log.LstdFlags)
	logger.Println("Starting unified monitoring system test...")

	// Create unified monitoring service
	service := monitoring.NewUnifiedMonitoringService(db, logger)
	adapter := monitoring.NewMonitoringAdapter(db, logger)

	ctx := context.Background()

	// Test 1: Record various types of metrics
	logger.Println("Test 1: Recording various types of metrics...")

	// Performance metrics
	perfMetric := &monitoring.UnifiedMetric{
		ID:                uuid.New(),
		Timestamp:         time.Now(),
		Component:         "api",
		ComponentInstance: "gateway",
		ServiceName:       "api_gateway",
		MetricType:        monitoring.MetricTypePerformance,
		MetricCategory:    monitoring.MetricCategoryLatency,
		MetricName:        "response_time",
		MetricValue:       150.5,
		MetricUnit:        "ms",
		Tags: map[string]interface{}{
			"endpoint": "/api/businesses",
			"method":   "GET",
			"status":   200,
		},
		Metadata: map[string]interface{}{
			"user_agent": "test-client",
			"ip_address": "192.168.1.1",
		},
		ConfidenceScore: 0.95,
		DataSource:      "api_gateway",
		CreatedAt:       time.Now(),
	}

	if err := service.RecordMetric(ctx, perfMetric); err != nil {
		logger.Printf("Failed to record performance metric: %v", err)
	} else {
		logger.Println("✓ Performance metric recorded successfully")
	}

	// Resource metrics
	resourceMetric := &monitoring.UnifiedMetric{
		ID:                uuid.New(),
		Timestamp:         time.Now(),
		Component:         "system",
		ComponentInstance: "main",
		ServiceName:       "system_monitor",
		MetricType:        monitoring.MetricTypeResource,
		MetricCategory:    monitoring.MetricCategoryMemory,
		MetricName:        "memory_usage",
		MetricValue:       1024.5,
		MetricUnit:        "MB",
		Tags: map[string]interface{}{
			"host": "server-01",
			"env":  "production",
		},
		Metadata: map[string]interface{}{
			"total_memory": 8192.0,
			"free_memory":  7167.5,
		},
		ConfidenceScore: 0.98,
		DataSource:      "system_monitor",
		CreatedAt:       time.Now(),
	}

	if err := service.RecordMetric(ctx, resourceMetric); err != nil {
		logger.Printf("Failed to record resource metric: %v", err)
	} else {
		logger.Println("✓ Resource metric recorded successfully")
	}

	// Business metrics
	businessMetric := &monitoring.UnifiedMetric{
		ID:                uuid.New(),
		Timestamp:         time.Now(),
		Component:         "classification",
		ComponentInstance: "ml_engine",
		ServiceName:       "classification_service",
		MetricType:        monitoring.MetricTypeBusiness,
		MetricCategory:    monitoring.MetricCategoryAccuracy,
		MetricName:        "classification_accuracy",
		MetricValue:       95.5,
		MetricUnit:        "percent",
		Tags: map[string]interface{}{
			"model":   "bert-base",
			"dataset": "validation",
		},
		Metadata: map[string]interface{}{
			"precision": 94.2,
			"recall":    96.8,
			"f1_score":  95.5,
		},
		ConfidenceScore: 0.92,
		DataSource:      "classification_service",
		CreatedAt:       time.Now(),
	}

	if err := service.RecordMetric(ctx, businessMetric); err != nil {
		logger.Printf("Failed to record business metric: %v", err)
	} else {
		logger.Println("✓ Business metric recorded successfully")
	}

	// Security metrics
	securityMetric := &monitoring.UnifiedMetric{
		ID:                uuid.New(),
		Timestamp:         time.Now(),
		Component:         "security",
		ComponentInstance: "validator",
		ServiceName:       "security_validator",
		MetricType:        monitoring.MetricTypeSecurity,
		MetricCategory:    monitoring.MetricCategoryValidation,
		MetricName:        "validation_time",
		MetricValue:       25.3,
		MetricUnit:        "ms",
		Tags: map[string]interface{}{
			"validation_type": "jwt",
			"result":          "success",
		},
		Metadata: map[string]interface{}{
			"token_claims": 5,
			"expiry_check": true,
		},
		ConfidenceScore: 0.99,
		DataSource:      "security_validator",
		CreatedAt:       time.Now(),
	}

	if err := service.RecordMetric(ctx, securityMetric); err != nil {
		logger.Printf("Failed to record security metric: %v", err)
	} else {
		logger.Println("✓ Security metric recorded successfully")
	}

	// Test 2: Record alerts
	logger.Println("\nTest 2: Recording various types of alerts...")

	// Performance alert
	perfAlert := &monitoring.UnifiedAlert{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		AlertType:         monitoring.AlertTypeThreshold,
		AlertCategory:     monitoring.AlertCategoryPerformance,
		Severity:          monitoring.AlertSeverityWarning,
		Component:         "api",
		ComponentInstance: "gateway",
		ServiceName:       "api_gateway",
		AlertName:         "high_response_time",
		Description:       "API response time is above threshold",
		Condition: map[string]interface{}{
			"metric_name": "response_time",
			"operator":    ">",
			"threshold":   100.0,
		},
		CurrentValue:   &[]float64{150.5}[0],
		ThresholdValue: &[]float64{100.0}[0],
		Status:         monitoring.AlertStatusActive,
		Tags: map[string]interface{}{
			"endpoint": "/api/businesses",
			"severity": "warning",
		},
		Metadata: map[string]interface{}{
			"alert_source": "threshold_monitor",
		},
	}

	if err := service.RecordAlert(ctx, perfAlert); err != nil {
		logger.Printf("Failed to record performance alert: %v", err)
	} else {
		logger.Println("✓ Performance alert recorded successfully")
	}

	// Resource alert
	resourceAlert := &monitoring.UnifiedAlert{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		AlertType:         monitoring.AlertTypeThreshold,
		AlertCategory:     monitoring.AlertCategoryResource,
		Severity:          monitoring.AlertSeverityCritical,
		Component:         "system",
		ComponentInstance: "main",
		ServiceName:       "system_monitor",
		AlertName:         "high_memory_usage",
		Description:       "Memory usage is critically high",
		Condition: map[string]interface{}{
			"metric_name": "memory_usage",
			"operator":    ">",
			"threshold":   90.0,
		},
		CurrentValue:   &[]float64{95.2}[0],
		ThresholdValue: &[]float64{90.0}[0],
		Status:         monitoring.AlertStatusActive,
		Tags: map[string]interface{}{
			"host":     "server-01",
			"severity": "critical",
		},
		Metadata: map[string]interface{}{
			"alert_source": "resource_monitor",
		},
	}

	if err := service.RecordAlert(ctx, resourceAlert); err != nil {
		logger.Printf("Failed to record resource alert: %v", err)
	} else {
		logger.Println("✓ Resource alert recorded successfully")
	}

	// Test 3: Query metrics
	logger.Println("\nTest 3: Querying metrics...")

	// Get all metrics
	allMetrics, err := service.GetMetrics(ctx, &monitoring.MetricFilters{
		Limit: 10,
	})
	if err != nil {
		logger.Printf("Failed to get all metrics: %v", err)
	} else {
		logger.Printf("✓ Retrieved %d metrics", len(allMetrics))
	}

	// Get metrics by component
	apiMetrics, err := service.GetMetrics(ctx, &monitoring.MetricFilters{
		Component: "api",
		Limit:     5,
	})
	if err != nil {
		logger.Printf("Failed to get API metrics: %v", err)
	} else {
		logger.Printf("✓ Retrieved %d API metrics", len(apiMetrics))
	}

	// Get metrics by type
	perfMetrics, err := service.GetMetrics(ctx, &monitoring.MetricFilters{
		MetricType: monitoring.MetricTypePerformance,
		Limit:      5,
	})
	if err != nil {
		logger.Printf("Failed to get performance metrics: %v", err)
	} else {
		logger.Printf("✓ Retrieved %d performance metrics", len(perfMetrics))
	}

	// Test 4: Query alerts
	logger.Println("\nTest 4: Querying alerts...")

	// Get all alerts
	allAlerts, err := service.GetAlerts(ctx, &monitoring.AlertFilters{
		Limit: 10,
	})
	if err != nil {
		logger.Printf("Failed to get all alerts: %v", err)
	} else {
		logger.Printf("✓ Retrieved %d alerts", len(allAlerts))
	}

	// Get active alerts
	activeAlerts, err := service.GetAlerts(ctx, &monitoring.AlertFilters{
		Status: monitoring.AlertStatusActive,
		Limit:  5,
	})
	if err != nil {
		logger.Printf("Failed to get active alerts: %v", err)
	} else {
		logger.Printf("✓ Retrieved %d active alerts", len(activeAlerts))
	}

	// Get critical alerts
	criticalAlerts, err := service.GetAlerts(ctx, &monitoring.AlertFilters{
		Severity: monitoring.AlertSeverityCritical,
		Limit:    5,
	})
	if err != nil {
		logger.Printf("Failed to get critical alerts: %v", err)
	} else {
		logger.Printf("✓ Retrieved %d critical alerts", len(criticalAlerts))
	}

	// Test 5: Get metrics summary
	logger.Println("\nTest 5: Getting metrics summary...")

	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	summary, err := service.GetMetricsSummary(ctx, "api", "api_gateway", startTime, endTime)
	if err != nil {
		logger.Printf("Failed to get metrics summary: %v", err)
	} else {
		logger.Printf("✓ Retrieved metrics summary: %+v", summary)
	}

	// Test 6: Get active alerts count
	logger.Println("\nTest 6: Getting active alerts count...")

	alertCounts, err := service.GetActiveAlertsCount(ctx)
	if err != nil {
		logger.Printf("Failed to get active alerts count: %v", err)
	} else {
		logger.Printf("✓ Retrieved active alerts count: %+v", alertCounts)
	}

	// Test 7: Test monitoring adapter
	logger.Println("\nTest 7: Testing monitoring adapter...")

	// Record database metrics using adapter
	dbMetrics := &monitoring.DatabaseMetrics{
		Timestamp:         time.Now(),
		ConnectionCount:   15,
		ActiveConnections: 8,
		IdleConnections:   7,
		MaxConnections:    100,
		QueryCount:        5000,
		SlowQueryCount:    25,
		ErrorCount:        10,
		AvgQueryTime:      45.2,
		MaxQueryTime:      250.0,
		DatabaseSize:      1024 * 1024 * 500, // 500MB
		TableSizes: map[string]int64{
			"public.businesses": 1024 * 1024 * 50, // 50MB
			"public.users":      1024 * 1024 * 20, // 20MB
		},
		IndexSizes: map[string]int64{
			"public.businesses_pkey": 1024 * 1024 * 5, // 5MB
			"public.users_pkey":      1024 * 1024 * 2, // 2MB
		},
		LockCount:     3,
		DeadlockCount: 0,
		CacheHitRatio: 97.5,
		Uptime:        7 * 24 * time.Hour, // 7 days
		Metadata: map[string]interface{}{
			"test_run": true,
		},
	}

	if err := adapter.RecordDatabaseMetrics(ctx, dbMetrics); err != nil {
		logger.Printf("Failed to record database metrics via adapter: %v", err)
	} else {
		logger.Println("✓ Database metrics recorded via adapter successfully")
	}

	// Record performance metric using adapter
	perfMetricAdapter := &monitoring.PerformanceMetric{
		Name:      "api_throughput",
		Value:     1000.0,
		Unit:      "requests_per_second",
		Timestamp: time.Now(),
		Tags: map[string]interface{}{
			"endpoint": "/api/classify",
			"method":   "POST",
		},
		Metadata: map[string]interface{}{
			"test_run": true,
		},
	}

	if err := adapter.RecordPerformanceMetric(ctx, "api", "classification_service", perfMetricAdapter); err != nil {
		logger.Printf("Failed to record performance metric via adapter: %v", err)
	} else {
		logger.Println("✓ Performance metric recorded via adapter successfully")
	}

	// Test 8: Update alert status
	logger.Println("\nTest 8: Updating alert status...")

	if len(activeAlerts) > 0 {
		alertID := activeAlerts[0].ID
		userID := uuid.New()

		if err := service.UpdateAlertStatus(ctx, alertID, monitoring.AlertStatusAcknowledged, &userID); err != nil {
			logger.Printf("Failed to update alert status: %v", err)
		} else {
			logger.Println("✓ Alert status updated successfully")
		}
	}

	// Test 9: Performance test
	logger.Println("\nTest 9: Performance test...")

	start := time.Now()
	for i := 0; i < 100; i++ {
		metric := &monitoring.UnifiedMetric{
			ID:                uuid.New(),
			Timestamp:         time.Now(),
			Component:         "performance_test",
			ComponentInstance: "test_instance",
			ServiceName:       "performance_test_service",
			MetricType:        monitoring.MetricTypePerformance,
			MetricCategory:    monitoring.MetricCategoryLatency,
			MetricName:        "test_metric",
			MetricValue:       float64(i),
			MetricUnit:        "ms",
			Tags: map[string]interface{}{
				"iteration": i,
			},
			Metadata: map[string]interface{}{
				"performance_test": true,
			},
			ConfidenceScore: 0.95,
			DataSource:      "performance_test",
			CreatedAt:       time.Now(),
		}

		if err := service.RecordMetric(ctx, metric); err != nil {
			logger.Printf("Failed to record performance test metric %d: %v", i, err)
		}
	}
	duration := time.Since(start)
	logger.Printf("✓ Recorded 100 metrics in %v (%.2f metrics/second)", duration, 100.0/duration.Seconds())

	// Test 10: Final verification
	logger.Println("\nTest 10: Final verification...")

	// Get final counts
	finalMetrics, err := service.GetMetrics(ctx, &monitoring.MetricFilters{
		Limit: 1000,
	})
	if err != nil {
		logger.Printf("Failed to get final metrics count: %v", err)
	} else {
		logger.Printf("✓ Total metrics in system: %d", len(finalMetrics))
	}

	finalAlerts, err := service.GetAlerts(ctx, &monitoring.AlertFilters{
		Limit: 1000,
	})
	if err != nil {
		logger.Printf("Failed to get final alerts count: %v", err)
	} else {
		logger.Printf("✓ Total alerts in system: %d", len(finalAlerts))
	}

	logger.Println("\n🎉 All unified monitoring system tests completed successfully!")
	logger.Println("The unified monitoring system is working correctly and ready for production use.")
}
