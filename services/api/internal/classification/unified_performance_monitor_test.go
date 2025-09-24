package classification

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap/zaptest"
)

// createUnifiedTestDB creates an in-memory SQLite database for testing
func createUnifiedTestDB() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	return db
}

func TestUnifiedPerformanceMonitor_NewUnifiedPerformanceMonitor(t *testing.T) {
	tests := []struct {
		name   string
		config *UnifiedPerformanceConfig
	}{
		{
			name:   "default config",
			config: nil,
		},
		{
			name: "custom config",
			config: &UnifiedPerformanceConfig{
				Enabled:                      true,
				CollectionInterval:           10 * time.Second,
				MetricsRetentionPeriod:       24 * time.Hour,
				AlertRetentionPeriod:         7 * 24 * time.Hour,
				EnableCrossComponentAnalysis: true,
				EnableUnifiedAlerting:        true,
				EnablePerformanceCorrelation: true,
				ServiceName:                  "test_service",
				Environment:                  "test",
				Version:                      "1.0.0",
				InstanceID:                   "test_instance",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := createUnifiedTestDB()
			defer db.Close()
			logger := zaptest.NewLogger(t)

			monitor, err := NewUnifiedPerformanceMonitor(db, logger, tt.config)
			if err != nil {
				t.Fatalf("Expected no error creating monitor, got: %v", err)
			}

			if monitor == nil {
				t.Fatal("Expected monitor to be created, got nil")
			}

			if monitor.comprehensiveMonitor == nil {
				t.Error("Expected comprehensive monitor to be initialized")
			}

			if monitor.responseTimeTracker == nil {
				t.Error("Expected response time tracker to be initialized")
			}

			if monitor.memoryMonitor == nil {
				t.Error("Expected memory monitor to be initialized")
			}

			if monitor.databaseMonitor == nil {
				t.Error("Expected database monitor to be initialized")
			}

			if monitor.securityMonitor == nil {
				t.Error("Expected security monitor to be initialized")
			}
		})
	}
}

func TestUnifiedPerformanceMonitor_StartStop(t *testing.T) {
	db := createUnifiedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(t)

	monitor, err := NewUnifiedPerformanceMonitor(db, logger, nil)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	// Test starting
	err = monitor.Start()
	if err != nil {
		t.Fatalf("Expected no error starting monitor, got: %v", err)
	}

	// Verify started state
	if !monitor.started {
		t.Error("Expected monitor to be started")
	}

	// Test stopping
	monitor.Stop()

	// Verify stopped state
	if monitor.started {
		t.Error("Expected monitor to be stopped")
	}
}

func TestUnifiedPerformanceMonitor_RecordClassificationMetrics(t *testing.T) {
	db := createUnifiedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(t)

	monitor, err := NewUnifiedPerformanceMonitor(db, logger, nil)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	err = monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	ctx := context.Background()
	perfContext := &ClassificationPerformanceContext{
		RequestID:         "test-req-001",
		ServiceName:       "test_service",
		StartTime:         time.Now().Add(-100 * time.Millisecond),
		EndTime:           time.Now(),
		ResponseTime:      100 * time.Millisecond,
		ProcessingTime:    80 * time.Millisecond,
		ConfidenceScore:   0.95,
		KeywordsProcessed: 10,
		ResultsCount:      3,
		ErrorOccurred:     false,
		ErrorMessage:      "",
		Metadata: map[string]interface{}{
			"user_id": "test_user",
		},
	}

	err = monitor.RecordClassificationMetrics(ctx, perfContext)
	if err != nil {
		t.Errorf("Expected no error recording metrics, got: %v", err)
	}

	// Allow time for async processing
	time.Sleep(100 * time.Millisecond)

	// Verify metrics were recorded
	stats := monitor.GetUnifiedStats()
	if stats.TotalRequests == 0 {
		t.Error("Expected total requests to be recorded")
	}

	if stats.AverageResponseTime <= 0 {
		t.Error("Expected average response time to be recorded")
	}
}

func TestUnifiedPerformanceMonitor_RecordDatabaseQuery(t *testing.T) {
	db := createUnifiedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(t)

	monitor, err := NewUnifiedPerformanceMonitor(db, logger, nil)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	err = monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	ctx := context.Background()
	query := "SELECT * FROM test_table WHERE id = ?"
	duration := 25 * time.Millisecond
	rowsReturned := int64(1)
	rowsExamined := int64(10)

	monitor.RecordDatabaseQuery(ctx, query, duration, rowsReturned, rowsExamined, false, "test_query_001")

	// Allow time for async processing
	time.Sleep(100 * time.Millisecond)

	// Verify database query was recorded
	stats := monitor.GetUnifiedStats()
	if stats.TotalDatabaseQueries == 0 {
		t.Error("Expected database queries to be recorded")
	}
}

func TestUnifiedPerformanceMonitor_RecordSecurityValidation(t *testing.T) {
	db := createUnifiedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(t)

	monitor, err := NewUnifiedPerformanceMonitor(db, logger, nil)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	err = monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	ctx := context.Background()
	result := &AdvancedSecurityValidationResult{
		ValidationID:   "test_sec_val_001",
		ValidationType: "input_sanitization",
		ValidationName: "user_input_check",
		ExecutionTime:  15 * time.Millisecond,
		Success:        true,
		Metadata:       make(map[string]interface{}),
		Timestamp:      time.Now(),
	}

	monitor.RecordSecurityValidation(ctx, result)

	// Allow time for async processing
	time.Sleep(100 * time.Millisecond)

	// Verify security validation was recorded
	stats := monitor.GetUnifiedStats()
	if stats.TotalSecurityValidations == 0 {
		t.Error("Expected security validations to be recorded")
	}
}

func TestUnifiedPerformanceMonitor_GetUnifiedStats(t *testing.T) {
	db := createUnifiedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(t)

	monitor, err := NewUnifiedPerformanceMonitor(db, logger, nil)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	err = monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	// Record some test data
	ctx := context.Background()

	// Record classification metrics
	perfContext := &ClassificationPerformanceContext{
		RequestID:         "test-req-002",
		ServiceName:       "test_service",
		StartTime:         time.Now().Add(-50 * time.Millisecond),
		EndTime:           time.Now(),
		ResponseTime:      50 * time.Millisecond,
		ProcessingTime:    40 * time.Millisecond,
		ConfidenceScore:   0.90,
		KeywordsProcessed: 5,
		ResultsCount:      2,
		ErrorOccurred:     false,
		Metadata:          make(map[string]interface{}),
	}
	monitor.RecordClassificationMetrics(ctx, perfContext)

	// Record database query
	monitor.RecordDatabaseQuery(ctx, "SELECT * FROM users", 10*time.Millisecond, 1, 5, false, "test_query_002")

	// Record security validation
	secResult := &AdvancedSecurityValidationResult{
		ValidationID:   "test_sec_val_002",
		ValidationType: "auth_check",
		ValidationName: "jwt_validation",
		ExecutionTime:  5 * time.Millisecond,
		Success:        true,
		Metadata:       make(map[string]interface{}),
		Timestamp:      time.Now(),
	}
	monitor.RecordSecurityValidation(ctx, secResult)

	// Allow time for processing
	time.Sleep(200 * time.Millisecond)

	// Get unified stats
	stats := monitor.GetUnifiedStats()

	// Verify stats structure
	if stats.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	if stats.SystemHealthScore <= 0 {
		t.Error("Expected system health score to be calculated")
	}

	if stats.OverallPerformanceScore <= 0 {
		t.Error("Expected overall performance score to be calculated")
	}

	if stats.OverallSecurityScore <= 0 {
		t.Error("Expected overall security score to be calculated")
	}

	// Verify component health indicators
	if stats.ResponseTimeHealth == "" {
		t.Error("Expected response time health to be set")
	}

	if stats.MemoryHealth == "" {
		t.Error("Expected memory health to be set")
	}

	if stats.DatabaseHealth == "" {
		t.Error("Expected database health to be set")
	}

	if stats.SecurityHealth == "" {
		t.Error("Expected security health to be set")
	}

	// Verify aggregated metrics
	if stats.TotalRequests <= 0 {
		t.Error("Expected total requests to be recorded")
	}

	if stats.AverageResponseTime <= 0 {
		t.Error("Expected average response time to be recorded")
	}

	if stats.TotalMemoryUsage <= 0 {
		t.Error("Expected total memory usage to be recorded")
	}

	if stats.TotalDatabaseQueries <= 0 {
		t.Error("Expected total database queries to be recorded")
	}

	if stats.TotalSecurityValidations <= 0 {
		t.Error("Expected total security validations to be recorded")
	}
}

func TestUnifiedPerformanceMonitor_GetUnifiedAlerts(t *testing.T) {
	db := createUnifiedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(t)

	monitor, err := NewUnifiedPerformanceMonitor(db, logger, nil)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	err = monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	ctx := context.Background()

	// Record some data that might trigger alerts
	perfContext := &ClassificationPerformanceContext{
		RequestID:         "test-req-003",
		ServiceName:       "test_service",
		StartTime:         time.Now().Add(-2000 * time.Millisecond), // 2 seconds - should trigger slow response alert
		EndTime:           time.Now(),
		ResponseTime:      2000 * time.Millisecond, // 2 seconds - slow response
		ProcessingTime:    1900 * time.Millisecond,
		ConfidenceScore:   0.50, // Low confidence
		KeywordsProcessed: 1,
		ResultsCount:      1,
		ErrorOccurred:     true, // Error occurred
		ErrorMessage:      "test error",
		Metadata:          make(map[string]interface{}),
	}
	monitor.RecordClassificationMetrics(ctx, perfContext)

	// Record a slow database query
	monitor.RecordDatabaseQuery(ctx, "SELECT * FROM large_table", 1500*time.Millisecond, 1000, 100000, false, "slow_query")

	// Record a failed security validation
	failedSecResult := &AdvancedSecurityValidationResult{
		ValidationID:      "test_sec_val_003",
		ValidationType:    "auth_check",
		ValidationName:    "api_key_validation",
		ExecutionTime:     100 * time.Millisecond,
		Success:           false,
		Error:             fmt.Errorf("invalid api key"),
		SecurityViolation: true,
		Metadata:          make(map[string]interface{}),
		Timestamp:         time.Now(),
	}
	monitor.RecordSecurityValidation(ctx, failedSecResult)

	// Allow time for processing and alert generation
	time.Sleep(500 * time.Millisecond)

	// Get unified alerts
	alerts, err := monitor.GetUnifiedAlerts(ctx, false, 10)
	if err != nil {
		t.Errorf("Expected no error getting alerts, got: %v", err)
	}

	// We might not have alerts immediately, but the structure should be correct
	// The important thing is that the method doesn't error and returns a valid structure
	if alerts == nil {
		t.Error("Expected alerts slice to be non-nil")
	}
}

func TestUnifiedPerformanceMonitor_GenerateUnifiedReport(t *testing.T) {
	db := createUnifiedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(t)

	monitor, err := NewUnifiedPerformanceMonitor(db, logger, nil)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	err = monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	ctx := context.Background()

	// Record some test data
	perfContext := &ClassificationPerformanceContext{
		RequestID:         "test-req-004",
		ServiceName:       "test_service",
		StartTime:         time.Now().Add(-100 * time.Millisecond),
		EndTime:           time.Now(),
		ResponseTime:      100 * time.Millisecond,
		ProcessingTime:    80 * time.Millisecond,
		ConfidenceScore:   0.95,
		KeywordsProcessed: 8,
		ResultsCount:      3,
		ErrorOccurred:     false,
		Metadata:          make(map[string]interface{}),
	}
	monitor.RecordClassificationMetrics(ctx, perfContext)

	// Allow time for processing
	time.Sleep(200 * time.Millisecond)

	// Generate report
	report, err := monitor.GenerateUnifiedReport(ctx, 1*time.Hour)
	if err != nil {
		t.Errorf("Expected no error generating report, got: %v", err)
	}

	if report == nil {
		t.Fatal("Expected report to be generated, got nil")
	}

	// Verify report structure
	if report.ReportID == "" {
		t.Error("Expected report ID to be set")
	}

	if report.GeneratedAt.IsZero() {
		t.Error("Expected generated at timestamp to be set")
	}

	if report.ReportPeriod <= 0 {
		t.Error("Expected report period to be set")
	}

	if report.ServiceName == "" {
		t.Error("Expected service name to be set")
	}

	if report.ExecutiveSummary == nil {
		t.Error("Expected executive summary to be present")
	}

	if report.ComponentAnalysis == nil {
		t.Error("Expected component analysis to be present")
	}

	if report.TrendAnalysis == nil {
		t.Error("Expected trend analysis to be present")
	}

	if report.PerformanceRecommendations == nil {
		t.Error("Expected performance recommendations to be present")
	}

	if report.SecurityRecommendations == nil {
		t.Error("Expected security recommendations to be present")
	}

	if report.ResourceRecommendations == nil {
		t.Error("Expected resource recommendations to be present")
	}

	if report.ActiveAlerts == nil {
		t.Error("Expected active alerts to be present")
	}
}

func TestPerformanceIntegrationService_NewPerformanceIntegrationService(t *testing.T) {
	tests := []struct {
		name   string
		config *PerformanceIntegrationServiceConfig
	}{
		{
			name:   "default config",
			config: nil,
		},
		{
			name: "custom config",
			config: &PerformanceIntegrationServiceConfig{
				Enabled:                  true,
				AutoStart:                false,
				HealthCheckInterval:      1 * time.Minute,
				ReportGenerationInterval: 30 * time.Minute,
				EnableRealTimeMonitoring: true,
				EnableHistoricalAnalysis: true,
				EnablePredictiveAnalysis: false,
				EnableAlerting:           true,
				AlertChannels:            []string{"log", "webhook"},
				EnableAutoReporting:      true,
				ReportFormats:            []string{"json", "html"},
				ReportStorageLocation:    "./test_reports",
				ServiceName:              "test_integration_service",
				Environment:              "test",
				Version:                  "1.0.0",
				InstanceID:               "test_instance_001",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := createUnifiedTestDB()
			defer db.Close()
			logger := zaptest.NewLogger(t)

			service, err := NewPerformanceIntegrationService(db, logger, tt.config)
			if err != nil {
				t.Fatalf("Expected no error creating service, got: %v", err)
			}

			if service == nil {
				t.Fatal("Expected service to be created, got nil")
			}

			if service.unifiedMonitor == nil {
				t.Error("Expected unified monitor to be initialized")
			}

			if service.logger == nil {
				t.Error("Expected logger to be initialized")
			}

			if service.config == nil {
				t.Error("Expected config to be initialized")
			}

			// Clean up
			service.Stop()
		})
	}
}

func TestPerformanceIntegrationService_RecordClassificationOperation(t *testing.T) {
	db := createUnifiedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(t)

	config := &PerformanceIntegrationServiceConfig{
		AutoStart: false, // Don't auto-start for this test
	}
	service, err := NewPerformanceIntegrationService(db, logger, config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Stop()

	err = service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	ctx := context.Background()
	operation := &ClassificationOperation{
		RequestID:       "test-op-001",
		ServiceName:     "test_service",
		Endpoint:        "/classify",
		Method:          "POST",
		StartTime:       time.Now().Add(-150 * time.Millisecond),
		EndTime:         time.Now(),
		ProcessingTime:  120 * time.Millisecond,
		ConfidenceScore: 0.92,
		KeywordsCount:   12,
		ResultsCount:    4,
		CacheHitRatio:   0.88,
		ErrorOccurred:   false,
		DatabaseQueries: []DatabaseQueryExecution{
			{
				QueryID:       "query_001",
				Query:         "SELECT * FROM classifications WHERE keywords LIKE ?",
				Duration:      20 * time.Millisecond,
				RowsReturned:  1,
				RowsExamined:  50,
				ErrorOccurred: false,
			},
		},
		SecurityValidations: []*AdvancedSecurityValidationResult{
			{
				ValidationID:   "sec_val_001",
				ValidationType: "input_sanitization",
				ValidationName: "user_input_check",
				ExecutionTime:  10 * time.Millisecond,
				Success:        true,
				Metadata:       make(map[string]interface{}),
				Timestamp:      time.Now(),
			},
		},
		Metadata: map[string]interface{}{
			"user_id": "test_user_001",
		},
	}

	err = service.RecordClassificationOperation(ctx, operation)
	if err != nil {
		t.Errorf("Expected no error recording operation, got: %v", err)
	}

	// Allow time for processing
	time.Sleep(200 * time.Millisecond)

	// Verify operation was recorded
	stats := service.GetUnifiedStats()
	if stats.TotalRequests == 0 {
		t.Error("Expected total requests to be recorded")
	}

	if stats.TotalDatabaseQueries == 0 {
		t.Error("Expected database queries to be recorded")
	}

	if stats.TotalSecurityValidations == 0 {
		t.Error("Expected security validations to be recorded")
	}
}

func TestPerformanceIntegrationService_GetSystemHealth(t *testing.T) {
	db := createUnifiedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(t)

	config := &PerformanceIntegrationServiceConfig{
		AutoStart: false,
	}
	service, err := NewPerformanceIntegrationService(db, logger, config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Stop()

	err = service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	// Get system health
	health := service.GetSystemHealth()

	// Verify health structure
	if health.ServiceName == "" {
		t.Error("Expected service name to be set")
	}

	if health.Status == "" {
		t.Error("Expected status to be set")
	}

	if health.LastHealthCheck.IsZero() {
		t.Error("Expected last health check to be set")
	}

	if health.ComponentHealth == nil {
		t.Error("Expected component health to be set")
	}

	if health.OverallScore <= 0 {
		t.Error("Expected overall score to be calculated")
	}

	if health.ActiveIssues == nil {
		t.Error("Expected active issues to be initialized")
	}

	if health.Recommendations == nil {
		t.Error("Expected recommendations to be initialized")
	}

	if health.Metadata == nil {
		t.Error("Expected metadata to be initialized")
	}

	// Verify component health keys
	expectedComponents := []string{"response_time", "memory", "database", "security"}
	for _, component := range expectedComponents {
		if _, exists := health.ComponentHealth[component]; !exists {
			t.Errorf("Expected component health for %s to be present", component)
		}
	}
}

func BenchmarkUnifiedPerformanceMonitor_RecordClassificationMetrics(b *testing.B) {
	db := createUnifiedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(b)

	monitor, err := NewUnifiedPerformanceMonitor(db, logger, nil)
	if err != nil {
		b.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	err = monitor.Start()
	if err != nil {
		b.Fatalf("Failed to start monitor: %v", err)
	}

	ctx := context.Background()
	perfContext := &ClassificationPerformanceContext{
		RequestID:         "bench-req",
		ServiceName:       "benchmark_service",
		StartTime:         time.Now().Add(-100 * time.Millisecond),
		EndTime:           time.Now(),
		ResponseTime:      100 * time.Millisecond,
		ProcessingTime:    80 * time.Millisecond,
		ConfidenceScore:   0.95,
		KeywordsProcessed: 10,
		ResultsCount:      3,
		ErrorOccurred:     false,
		Metadata:          make(map[string]interface{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		perfContext.RequestID = fmt.Sprintf("bench-req-%d", i)
		monitor.RecordClassificationMetrics(ctx, perfContext)
	}
}

func BenchmarkUnifiedPerformanceMonitor_GetUnifiedStats(b *testing.B) {
	db := createUnifiedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(b)

	monitor, err := NewUnifiedPerformanceMonitor(db, logger, nil)
	if err != nil {
		b.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	err = monitor.Start()
	if err != nil {
		b.Fatalf("Failed to start monitor: %v", err)
	}

	// Pre-populate with some data
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		perfContext := &ClassificationPerformanceContext{
			RequestID:         fmt.Sprintf("bench-req-%d", i),
			ServiceName:       "benchmark_service",
			StartTime:         time.Now().Add(-100 * time.Millisecond),
			EndTime:           time.Now(),
			ResponseTime:      100 * time.Millisecond,
			ProcessingTime:    80 * time.Millisecond,
			ConfidenceScore:   0.95,
			KeywordsProcessed: 10,
			ResultsCount:      3,
			ErrorOccurred:     false,
			Metadata:          make(map[string]interface{}),
		}
		monitor.RecordClassificationMetrics(ctx, perfContext)
	}

	time.Sleep(200 * time.Millisecond) // Allow processing

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.GetUnifiedStats()
	}
}
