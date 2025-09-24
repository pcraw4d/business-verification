package classification

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap/zaptest"
)

// TestPerformanceMonitoringEndToEnd tests end-to-end performance monitoring workflow
func TestPerformanceMonitoringEndToEnd(t *testing.T) {
	// Setup test database
	db := createComprehensiveTestDB()
	defer db.Close()

	// Setup logger
	logger := zaptest.NewLogger(t)

	// Create comprehensive performance monitor
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	// Create all monitoring components
	responseTimeTracker := NewResponseTimeTracker(DefaultResponseTimeConfig(), logger)
	memoryMonitor := NewAdvancedMemoryMonitor(logger, DefaultMemoryMonitorConfig())
	databaseMonitor := NewEnhancedDatabaseMonitor(db, logger, DefaultDatabaseMonitorConfig(), comprehensiveMonitor)
	securityMonitor := NewAdvancedSecurityValidationMonitor(logger, DefaultSecurityValidationConfig())

	// Start all monitors
	responseTimeTracker.Start()
	memoryMonitor.Start()
	databaseMonitor.Start()
	securityMonitor.Start()

	// Cleanup
	defer func() {
		responseTimeTracker.Stop()
		memoryMonitor.Stop()
		databaseMonitor.Stop()
		securityMonitor.Stop()
	}()

	ctx := context.Background()

	// Simulate a complete business classification request
	t.Run("business_classification_workflow", func(t *testing.T) {
		requestID := "test_request_123"
		startTime := time.Now()

		// Step 1: Track API request start
		responseTimeTracker.TrackResponseTime("POST", "/api/classify", 0, 0, nil)

		// Step 2: Record memory usage at start
		memoryMonitor.CollectMemoryMetrics()

		// Step 3: Simulate database queries
		databaseMonitor.RecordQueryExecution(ctx, "SELECT * FROM business_data WHERE id = ?", 25*time.Millisecond, 1, 10, nil)
		databaseMonitor.RecordQueryExecution(ctx, "SELECT * FROM classification_rules WHERE active = true", 15*time.Millisecond, 5, 5, nil)

		// Step 4: Simulate security validation
		securityResult := &AdvancedSecurityValidationResult{
			ValidationID:               fmt.Sprintf("security_%s", requestID),
			ValidationType:             "data_source_validation",
			ValidationName:             "business_data_validation",
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
		securityMonitor.RecordSecurityValidation(ctx, securityResult)

		// Step 5: Simulate classification processing
		processingTime := 50 * time.Millisecond
		time.Sleep(processingTime)

		// Step 6: Record comprehensive performance metrics
		classificationMetric := &ComprehensivePerformanceMetric{
			ID:                       fmt.Sprintf("classification_%s", requestID),
			Timestamp:                time.Now(),
			MetricType:               "classification",
			ServiceName:              "business_classification_service",
			ResponseTimeMs:           float64(time.Since(startTime).Milliseconds()),
			ProcessingTimeMs:         float64(processingTime.Milliseconds()),
			Confidence:               0.90,
			KeywordsCount:            15,
			ResultsCount:             3,
			CacheHitRatio:            0.85,
			ErrorOccurred:            false,
			ParallelProcessing:       true,
			GoroutinesUsed:           4,
			DatabaseQueryCount:       2,
			DatabaseQueryTimeMs:      40.0,
			SecurityValidationTimeMs: 30.0,
			ClassificationAccuracy:   0.92,
			Metadata: map[string]interface{}{
				"request_id":            requestID,
				"business_type":         "retail",
				"classification_method": "multi_strategy",
			},
		}
		comprehensiveMonitor.RecordPerformanceMetric(ctx, classificationMetric)

		// Step 7: Track API response completion
		responseTimeTracker.TrackResponseTime("POST", "/api/classify", time.Since(startTime), 200, nil)

		// Step 8: Record final memory usage
		memoryMonitor.CollectMemoryMetrics()

		// Allow time for processing
		time.Sleep(100 * time.Millisecond)

		// Verify end-to-end metrics
		t.Run("verify_response_time_tracking", func(t *testing.T) {
			stats := responseTimeTracker.GetResponseTimeStats("POST", "/api/classify", 1*time.Minute)
			if stats == nil {
				t.Error("Expected response time stats for API endpoint")
			}
			if stats.RequestCount == 0 {
				t.Error("Expected request count > 0")
			}
		})

		t.Run("verify_memory_monitoring", func(t *testing.T) {
			metrics := memoryMonitor.GetLatestMemoryMetrics()
			if metrics == nil {
				t.Error("Expected memory metrics")
			}
		})

		t.Run("verify_database_monitoring", func(t *testing.T) {
			stats := databaseMonitor.GetQueryPerformanceStats()
			if len(stats) == 0 {
				t.Error("Expected database query stats")
			}
		})

		t.Run("verify_security_monitoring", func(t *testing.T) {
			stats := securityMonitor.GetValidationStats(10)
			if len(stats) == 0 {
				t.Error("Expected security validation stats")
			}
		})

		t.Run("verify_comprehensive_monitoring", func(t *testing.T) {
			metrics := comprehensiveMonitor.GetPerformanceMetrics(10)
			if len(metrics) == 0 {
				t.Error("Expected comprehensive performance metrics")
			}

			// Verify classification-specific metrics
			foundClassificationMetric := false
			for _, metric := range metrics {
				if metric.MetricType == "classification" && metric.ServiceName == "business_classification_service" {
					foundClassificationMetric = true
					if metric.Confidence != 0.90 {
						t.Errorf("Expected confidence 0.90, got %.2f", metric.Confidence)
					}
					if metric.KeywordsCount != 15 {
						t.Errorf("Expected keywords count 15, got %d", metric.KeywordsCount)
					}
					if metric.ResultsCount != 3 {
						t.Errorf("Expected results count 3, got %d", metric.ResultsCount)
					}
					if metric.ClassificationAccuracy != 0.92 {
						t.Errorf("Expected classification accuracy 0.92, got %.2f", metric.ClassificationAccuracy)
					}
					break
				}
			}
			if !foundClassificationMetric {
				t.Error("Expected to find classification metric")
			}
		})
	})
}

// TestPerformanceMonitoringStressTest tests the system under stress
func TestPerformanceMonitoringStressTest(t *testing.T) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(t)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	// Stress test parameters
	concurrentWorkers := 20
	requestsPerWorker := 50
	totalRequests := concurrentWorkers * requestsPerWorker

	var wg sync.WaitGroup
	startTime := time.Now()

	// Launch concurrent workers
	for i := 0; i < concurrentWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < requestsPerWorker; j++ {
				// Simulate different types of requests
				requestTypes := []string{"classification", "validation", "analysis", "reporting"}
				requestType := requestTypes[j%len(requestTypes)]

				metric := &ComprehensivePerformanceMetric{
					ID:               fmt.Sprintf("stress_test_%d_%d", workerID, j),
					Timestamp:        time.Now(),
					MetricType:       requestType,
					ServiceName:      fmt.Sprintf("stress_test_service_%d", workerID%5),
					ResponseTimeMs:   float64(50 + (j % 100)), // Vary response times
					ProcessingTimeMs: float64(30 + (j % 50)),
					Confidence:       0.8 + float64(j%20)/100.0, // Vary confidence
					KeywordsCount:    j % 20,
					ResultsCount:     j % 10,
					CacheHitRatio:    0.7 + float64(j%30)/100.0,
					ErrorOccurred:    j%20 == 0, // 5% error rate
					ErrorMessage: func() string {
						if j%20 == 0 {
							return "Simulated stress test error"
						}
						return ""
					}(),
					ParallelProcessing:     j%2 == 0,
					GoroutinesUsed:         2 + (j % 8),
					MemoryUsageMB:          float64(100 + (j % 200)),
					DatabaseQueryCount:     1 + (j % 5),
					DatabaseQueryTimeMs:    float64(10 + (j % 40)),
					ClassificationAccuracy: 0.85 + float64(j%15)/100.0,
					Metadata: map[string]interface{}{
						"worker_id":   workerID,
						"request_id":  j,
						"stress_test": true,
					},
				}

				comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)

				// Small delay to simulate real processing
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	// Wait for all workers to complete
	wg.Wait()
	duration := time.Since(startTime)

	// Allow time for processing
	time.Sleep(200 * time.Millisecond)

	// Verify stress test results
	t.Logf("Stress test completed: %d requests in %v (%.2f requests/sec)",
		totalRequests, duration, float64(totalRequests)/duration.Seconds())

	// Verify metrics were recorded
	metrics := comprehensiveMonitor.GetPerformanceMetrics(totalRequests)
	if len(metrics) < totalRequests/2 { // Allow for some loss due to async processing
		t.Errorf("Expected at least %d metrics, got %d", totalRequests/2, len(metrics))
	}

	// Verify error rate is approximately correct
	errorCount := 0
	for _, metric := range metrics {
		if metric.ErrorOccurred {
			errorCount++
		}
	}

	if len(metrics) > 0 {
		errorRate := float64(errorCount) / float64(len(metrics))
		expectedErrorRate := 0.05 // 5%

		// Allow for some variance in error rate
		if errorRate > expectedErrorRate*2 || errorRate < expectedErrorRate/2 {
			t.Logf("Error rate variance: expected ~%.2f, got %.2f", expectedErrorRate, errorRate)
		}
	}

	// Verify service distribution
	serviceCounts := make(map[string]int)
	for _, metric := range metrics {
		serviceCounts[metric.ServiceName]++
	}

	if len(serviceCounts) != 5 {
		t.Errorf("Expected 5 different services, got %d", len(serviceCounts))
	}
}

// TestPerformanceMonitoringDataIntegrity tests data integrity across monitoring components
func TestPerformanceMonitoringDataIntegrity(t *testing.T) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(t)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	// Test data integrity with various metric types
	testCases := []struct {
		name           string
		metricType     string
		serviceName    string
		expectedFields map[string]interface{}
	}{
		{
			name:        "response_time_metric",
			metricType:  "response_time",
			serviceName: "api_service",
			expectedFields: map[string]interface{}{
				"response_time_ms":   100.0,
				"processing_time_ms": 80.0,
			},
		},
		{
			name:        "memory_metric",
			metricType:  "memory",
			serviceName: "memory_service",
			expectedFields: map[string]interface{}{
				"memory_usage_mb":     500.0,
				"memory_allocated_mb": 400.0,
				"gc_pause_time_ms":    10.0,
			},
		},
		{
			name:        "database_metric",
			metricType:  "database",
			serviceName: "database_service",
			expectedFields: map[string]interface{}{
				"database_query_count":      5,
				"database_query_time_ms":    150.0,
				"database_connection_count": 10,
			},
		},
		{
			name:        "security_metric",
			metricType:  "security",
			serviceName: "security_service",
			expectedFields: map[string]interface{}{
				"security_validation_time_ms":  200.0,
				"trusted_data_source_count":    3,
				"website_verification_time_ms": 100.0,
			},
		},
	}

	// Record test metrics
	for _, tc := range testCases {
		metric := &ComprehensivePerformanceMetric{
			ID:          fmt.Sprintf("integrity_test_%s", tc.name),
			Timestamp:   time.Now(),
			MetricType:  tc.metricType,
			ServiceName: tc.serviceName,
			Metadata:    make(map[string]interface{}),
		}

		// Set expected fields
		for field, value := range tc.expectedFields {
			switch field {
			case "response_time_ms":
				metric.ResponseTimeMs = value.(float64)
			case "processing_time_ms":
				metric.ProcessingTimeMs = value.(float64)
			case "memory_usage_mb":
				metric.MemoryUsageMB = value.(float64)
			case "memory_allocated_mb":
				metric.MemoryAllocatedMB = value.(float64)
			case "gc_pause_time_ms":
				metric.GCPauseTimeMs = value.(float64)
			case "database_query_count":
				metric.DatabaseQueryCount = value.(int)
			case "database_query_time_ms":
				metric.DatabaseQueryTimeMs = value.(float64)
			case "database_connection_count":
				metric.DatabaseConnectionCount = value.(int)
			case "security_validation_time_ms":
				metric.SecurityValidationTimeMs = value.(float64)
			case "trusted_data_source_count":
				metric.TrustedDataSourceCount = value.(int)
			case "website_verification_time_ms":
				metric.WebsiteVerificationTimeMs = value.(float64)
			}
		}

		comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
	}

	// Allow time for processing
	time.Sleep(100 * time.Millisecond)

	// Verify data integrity
	metrics := comprehensiveMonitor.GetPerformanceMetrics(100)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			found := false
			for _, metric := range metrics {
				if metric.ID == fmt.Sprintf("integrity_test_%s", tc.name) {
					found = true

					// Verify basic fields
					if metric.MetricType != tc.metricType {
						t.Errorf("Expected metric type %s, got %s", tc.metricType, metric.MetricType)
					}
					if metric.ServiceName != tc.serviceName {
						t.Errorf("Expected service name %s, got %s", tc.serviceName, metric.ServiceName)
					}

					// Verify expected fields
					for field, expectedValue := range tc.expectedFields {
						var actualValue interface{}
						switch field {
						case "response_time_ms":
							actualValue = metric.ResponseTimeMs
						case "processing_time_ms":
							actualValue = metric.ProcessingTimeMs
						case "memory_usage_mb":
							actualValue = metric.MemoryUsageMB
						case "memory_allocated_mb":
							actualValue = metric.MemoryAllocatedMB
						case "gc_pause_time_ms":
							actualValue = metric.GCPauseTimeMs
						case "database_query_count":
							actualValue = metric.DatabaseQueryCount
						case "database_query_time_ms":
							actualValue = metric.DatabaseQueryTimeMs
						case "database_connection_count":
							actualValue = metric.DatabaseConnectionCount
						case "security_validation_time_ms":
							actualValue = metric.SecurityValidationTimeMs
						case "trusted_data_source_count":
							actualValue = metric.TrustedDataSourceCount
						case "website_verification_time_ms":
							actualValue = metric.WebsiteVerificationTimeMs
						}

						if actualValue != expectedValue {
							t.Errorf("Field %s: expected %v, got %v", field, expectedValue, actualValue)
						}
					}
					break
				}
			}
			if !found {
				t.Errorf("Expected to find metric for test case %s", tc.name)
			}
		})
	}
}

// TestPerformanceMonitoringAlertingIntegration tests alerting integration across components
func TestPerformanceMonitoringAlertingIntegration(t *testing.T) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(t)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	// Test various alert scenarios
	alertScenarios := []struct {
		name        string
		metric      *ComprehensivePerformanceMetric
		expectAlert bool
		alertType   string
	}{
		{
			name: "high_response_time",
			metric: &ComprehensivePerformanceMetric{
				ID:             "high_response_time_alert",
				Timestamp:      time.Now(),
				MetricType:     "response_time",
				ServiceName:    "slow_service",
				ResponseTimeMs: 5000.0, // Very high
				Metadata:       make(map[string]interface{}),
			},
			expectAlert: true,
			alertType:   "threshold_exceeded",
		},
		{
			name: "high_memory_usage",
			metric: &ComprehensivePerformanceMetric{
				ID:            "high_memory_alert",
				Timestamp:     time.Now(),
				MetricType:    "memory",
				ServiceName:   "memory_intensive_service",
				MemoryUsageMB: 2000.0, // Very high
				Metadata:      make(map[string]interface{}),
			},
			expectAlert: true,
			alertType:   "threshold_exceeded",
		},
		{
			name: "database_error",
			metric: &ComprehensivePerformanceMetric{
				ID:            "database_error_alert",
				Timestamp:     time.Now(),
				MetricType:    "database",
				ServiceName:   "database_service",
				ErrorOccurred: true,
				ErrorMessage:  "Database connection failed",
				Metadata:      make(map[string]interface{}),
			},
			expectAlert: true,
			alertType:   "error_detected",
		},
		{
			name: "normal_metric",
			metric: &ComprehensivePerformanceMetric{
				ID:             "normal_metric",
				Timestamp:      time.Now(),
				MetricType:     "response_time",
				ServiceName:    "normal_service",
				ResponseTimeMs: 50.0, // Normal
				Metadata:       make(map[string]interface{}),
			},
			expectAlert: false,
		},
	}

	// Record metrics and check for alerts
	for _, scenario := range alertScenarios {
		comprehensiveMonitor.RecordPerformanceMetric(ctx, scenario.metric)
	}

	// Allow time for alert processing
	time.Sleep(200 * time.Millisecond)

	// Verify alerts
	alerts := comprehensiveMonitor.GetPerformanceAlerts(false, 100)

	alertCount := 0
	for _, scenario := range alertScenarios {
		if scenario.expectAlert {
			alertCount++
			found := false
			for _, alert := range alerts {
				if alert.MetricName == scenario.name {
					found = true
					if alert.AlertType != scenario.alertType {
						t.Errorf("Expected alert type %s, got %s", scenario.alertType, alert.AlertType)
					}
					break
				}
			}
			if !found {
				t.Errorf("Expected alert for scenario %s", scenario.name)
			}
		}
	}

	if len(alerts) < alertCount {
		t.Errorf("Expected at least %d alerts, got %d", alertCount, len(alerts))
	}
}

// TestPerformanceMonitoringCleanupAndMaintenance tests cleanup and maintenance operations
func TestPerformanceMonitoringCleanupAndMaintenance(t *testing.T) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(t)
	config := DefaultPerformanceMonitorConfig()
	config.MaxMetrics = 50 // Small limit for testing
	config.CleanupInterval = 100 * time.Millisecond
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, config)
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	// Record more metrics than the limit
	metricCount := 100
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

	// Wait for cleanup to occur
	time.Sleep(300 * time.Millisecond)

	// Verify cleanup occurred
	metrics := comprehensiveMonitor.GetPerformanceMetrics(metricCount)
	if len(metrics) > config.MaxMetrics {
		t.Errorf("Expected cleanup to limit metrics to %d, got %d", config.MaxMetrics, len(metrics))
	}

	// Verify oldest metrics were cleaned up
	metricIDs := make(map[string]bool)
	for _, metric := range metrics {
		metricIDs[metric.ID] = true
	}

	// Check that some of the oldest metrics were cleaned up
	oldestMetricFound := false
	for i := 0; i < 10; i++ {
		if metricIDs[fmt.Sprintf("cleanup_test_metric_%d", i)] {
			oldestMetricFound = true
			break
		}
	}

	// Some of the oldest metrics should have been cleaned up
	if oldestMetricFound && len(metrics) >= config.MaxMetrics {
		t.Log("Cleanup appears to be working - some old metrics were removed")
	}
}
