package classification

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// createMockDBForPerformanceAlerting creates a mock database connection for testing
func createMockDBForPerformanceAlerting() *sql.DB {
	// In a real test, you would use a test database or mock
	// For now, we'll return nil and skip tests that require DB
	return nil
}

func TestPerformanceAlerting_GeneratePerformanceAlert(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	alertType := "TEST_ALERT"
	alertLevel := "HIGH"
	alertCategory := "TESTING"
	alertTitle := "Test Alert"
	alertMessage := "This is a test alert"
	metricName := "test_metric"
	metricValue := 85.5
	thresholdValue := 80.0
	thresholdType := "greater_than"
	affectedSystems := []string{"test_system"}
	recommendations := []string{"test_recommendation"}

	alertID, err := pa.GeneratePerformanceAlert(
		ctx,
		alertType,
		alertLevel,
		alertCategory,
		alertTitle,
		alertMessage,
		metricName,
		&metricValue,
		&thresholdValue,
		thresholdType,
		affectedSystems,
		recommendations,
	)

	if err != nil {
		t.Fatalf("GeneratePerformanceAlert failed: %v", err)
	}

	t.Logf("Performance alert generated with ID: %s", alertID)
}

func TestPerformanceAlerting_CheckDatabasePerformanceAlerts(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	alerts, err := pa.CheckDatabasePerformanceAlerts(ctx)
	if err != nil {
		t.Fatalf("CheckDatabasePerformanceAlerts failed: %v", err)
	}

	t.Logf("Found %d database performance alerts", len(alerts))
	for _, alert := range alerts {
		t.Logf("Database performance alert: %+v", alert)
	}
}

func TestPerformanceAlerting_CheckClassificationAccuracyAlerts(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	alerts, err := pa.CheckClassificationAccuracyAlerts(ctx)
	if err != nil {
		t.Fatalf("CheckClassificationAccuracyAlerts failed: %v", err)
	}

	t.Logf("Found %d classification accuracy alerts", len(alerts))
	for _, alert := range alerts {
		t.Logf("Classification accuracy alert: %+v", alert)
	}
}

func TestPerformanceAlerting_CheckSystemResourceAlerts(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	alerts, err := pa.CheckSystemResourceAlerts(ctx)
	if err != nil {
		t.Fatalf("CheckSystemResourceAlerts failed: %v", err)
	}

	t.Logf("Found %d system resource alerts", len(alerts))
	for _, alert := range alerts {
		t.Logf("System resource alert: %+v", alert)
	}
}

func TestPerformanceAlerting_GetActivePerformanceAlerts(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	alerts, err := pa.GetActivePerformanceAlerts(ctx)
	if err != nil {
		t.Fatalf("GetActivePerformanceAlerts failed: %v", err)
	}

	t.Logf("Found %d active performance alerts", len(alerts))
	for _, alert := range alerts {
		t.Logf("Active performance alert: %+v", alert)
	}
}

func TestPerformanceAlerting_AcknowledgePerformanceAlert(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	alertID := "test-alert-123"
	acknowledgedBy := "test-user"

	acknowledged, err := pa.AcknowledgePerformanceAlert(ctx, alertID, acknowledgedBy)
	if err != nil {
		t.Fatalf("AcknowledgePerformanceAlert failed: %v", err)
	}

	t.Logf("Alert acknowledged: %t", acknowledged)
}

func TestPerformanceAlerting_ResolvePerformanceAlert(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	alertID := "test-alert-123"
	resolutionNotes := "Test resolution"

	resolved, err := pa.ResolvePerformanceAlert(ctx, alertID, &resolutionNotes)
	if err != nil {
		t.Fatalf("ResolvePerformanceAlert failed: %v", err)
	}

	t.Logf("Alert resolved: %t", resolved)
}

func TestPerformanceAlerting_RunAllPerformanceChecks(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := pa.RunAllPerformanceChecks(ctx)
	if err != nil {
		t.Fatalf("RunAllPerformanceChecks failed: %v", err)
	}

	t.Logf("Found %d performance check results", len(results))
	for _, result := range results {
		t.Logf("Performance check result: %+v", result)
	}
}

func TestPerformanceAlerting_GetAlertStatistics(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	stats, err := pa.GetAlertStatistics(ctx, 24)
	if err != nil {
		t.Fatalf("GetAlertStatistics failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}

	t.Logf("Alert statistics: %+v", stats)
}

func TestPerformanceAlerting_CleanupOldPerformanceAlerts(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	deletedCount, err := pa.CleanupOldPerformanceAlerts(ctx, 30)
	if err != nil {
		t.Fatalf("CleanupOldPerformanceAlerts failed: %v", err)
	}

	t.Logf("Cleaned up %d old performance alert entries", deletedCount)
}

func TestPerformanceAlerting_ValidateAlertingSetup(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	validation, err := pa.ValidateAlertingSetup(ctx)
	if err != nil {
		t.Fatalf("ValidateAlertingSetup failed: %v", err)
	}

	t.Logf("Found %d alerting validation results", len(validation))
	for _, result := range validation {
		t.Logf("Alerting validation: %+v", result)
	}
}

func TestPerformanceAlerting_GetCurrentAlertStatus(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	status, err := pa.GetCurrentAlertStatus(ctx)
	if err != nil {
		t.Fatalf("GetCurrentAlertStatus failed: %v", err)
	}

	t.Logf("Current alert status: %+v", status)
}

func TestPerformanceAlerting_GetPerformanceAlertSummary(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	summary, err := pa.GetPerformanceAlertSummary(ctx)
	if err != nil {
		t.Fatalf("GetPerformanceAlertSummary failed: %v", err)
	}

	t.Logf("Performance alert summary: %+v", summary)
}

// Benchmark tests for performance alerting
func BenchmarkPerformanceAlerting_GeneratePerformanceAlert(b *testing.B) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	alertType := "TEST_ALERT"
	alertLevel := "HIGH"
	alertCategory := "TESTING"
	alertTitle := "Test Alert"
	alertMessage := "This is a test alert"
	metricName := "test_metric"
	metricValue := 85.5
	thresholdValue := 80.0
	thresholdType := "greater_than"
	affectedSystems := []string{"test_system"}
	recommendations := []string{"test_recommendation"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pa.GeneratePerformanceAlert(
			ctx,
			alertType,
			alertLevel,
			alertCategory,
			alertTitle,
			alertMessage,
			metricName,
			&metricValue,
			&thresholdValue,
			thresholdType,
			affectedSystems,
			recommendations,
		)
		if err != nil {
			b.Fatalf("GeneratePerformanceAlert failed: %v", err)
		}
	}
}

func BenchmarkPerformanceAlerting_GetActivePerformanceAlerts(b *testing.B) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pa.GetActivePerformanceAlerts(ctx)
		if err != nil {
			b.Fatalf("GetActivePerformanceAlerts failed: %v", err)
		}
	}
}

func BenchmarkPerformanceAlerting_RunAllPerformanceChecks(b *testing.B) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pa.RunAllPerformanceChecks(ctx)
		if err != nil {
			b.Fatalf("RunAllPerformanceChecks failed: %v", err)
		}
	}
}

// Integration test that would require a real database
func TestPerformanceAlerting_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This would require a real database connection
	// db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
	// if err != nil {
	//     t.Fatalf("Failed to connect to database: %v", err)
	// }
	// defer db.Close()

	// pa := NewPerformanceAlerting(db)
	// ctx := context.Background()

	// Test all functions with real data
	t.Log("Integration test would run here with real database")
}

// Test error handling
func TestPerformanceAlerting_ErrorHandling(t *testing.T) {
	// Test with nil database
	pa := NewPerformanceAlerting(nil)
	ctx := context.Background()

	// These should return errors due to nil database
	_, err := pa.GetActivePerformanceAlerts(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = pa.GetCurrentAlertStatus(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = pa.RunAllPerformanceChecks(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}
}

// Test continuous monitoring
func TestPerformanceAlerting_ContinuousMonitoring(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start continuous monitoring with short interval
	go pa.MonitorPerformanceContinuously(ctx, 1*time.Second)

	// Wait for context to timeout
	<-ctx.Done()

	t.Log("Continuous performance monitoring test completed")
}

// Test performance alert management
func TestPerformanceAlerting_AlertManagement(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test alert generation
	alertID, err := pa.GeneratePerformanceAlert(
		ctx,
		"TEST_ALERT",
		"HIGH",
		"TESTING",
		"Test Alert",
		"This is a test alert",
		"test_metric",
		nil, // metricValue
		nil, // thresholdValue
		"greater_than",
		[]string{"test_system"},
		[]string{"test_recommendation"},
	)
	if err != nil {
		t.Fatalf("GeneratePerformanceAlert failed: %v", err)
	}

	// Test alert acknowledgment
	acknowledged, err := pa.AcknowledgePerformanceAlert(ctx, alertID, "test-user")
	if err != nil {
		t.Fatalf("AcknowledgePerformanceAlert failed: %v", err)
	}

	// Test alert resolution
	resolved, err := pa.ResolvePerformanceAlert(ctx, alertID, nil)
	if err != nil {
		t.Fatalf("ResolvePerformanceAlert failed: %v", err)
	}

	t.Logf("Alert management test: generated=%s, acknowledged=%t, resolved=%t",
		alertID, acknowledged, resolved)
}

// Test performance checks
func TestPerformanceAlerting_PerformanceChecks(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test database performance checks
	dbAlerts, err := pa.CheckDatabasePerformanceAlerts(ctx)
	if err != nil {
		t.Fatalf("CheckDatabasePerformanceAlerts failed: %v", err)
	}

	// Test classification accuracy checks
	classificationAlerts, err := pa.CheckClassificationAccuracyAlerts(ctx)
	if err != nil {
		t.Fatalf("CheckClassificationAccuracyAlerts failed: %v", err)
	}

	// Test system resource checks
	systemAlerts, err := pa.CheckSystemResourceAlerts(ctx)
	if err != nil {
		t.Fatalf("CheckSystemResourceAlerts failed: %v", err)
	}

	// Test all performance checks
	checkResults, err := pa.RunAllPerformanceChecks(ctx)
	if err != nil {
		t.Fatalf("RunAllPerformanceChecks failed: %v", err)
	}

	t.Logf("Performance checks: DB=%d, Classification=%d, System=%d, Total=%d",
		len(dbAlerts), len(classificationAlerts), len(systemAlerts), len(checkResults))
}

// Test alert statistics
func TestPerformanceAlerting_AlertStatistics(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test alert statistics
	stats, err := pa.GetAlertStatistics(ctx, 24)
	if err != nil {
		t.Fatalf("GetAlertStatistics failed: %v", err)
	}

	// Test current alert status
	status, err := pa.GetCurrentAlertStatus(ctx)
	if err != nil {
		t.Fatalf("GetCurrentAlertStatus failed: %v", err)
	}

	// Test performance alert summary
	summary, err := pa.GetPerformanceAlertSummary(ctx)
	if err != nil {
		t.Fatalf("GetPerformanceAlertSummary failed: %v", err)
	}

	t.Logf("Alert statistics: %+v", stats)
	t.Logf("Alert status: %+v", status)
	t.Logf("Alert summary keys: %d", len(summary))
}

// Test alert validation
func TestPerformanceAlerting_AlertValidation(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test alerting setup validation
	validation, err := pa.ValidateAlertingSetup(ctx)
	if err != nil {
		t.Fatalf("ValidateAlertingSetup failed: %v", err)
	}

	t.Logf("Alerting validation results: %d", len(validation))
	for _, result := range validation {
		t.Logf("Validation: %+v", result)
	}
}

// Test alert cleanup
func TestPerformanceAlerting_AlertCleanup(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test alert cleanup
	deletedCount, err := pa.CleanupOldPerformanceAlerts(ctx, 30)
	if err != nil {
		t.Fatalf("CleanupOldPerformanceAlerts failed: %v", err)
	}

	t.Logf("Cleaned up %d alert entries", deletedCount)
}

// Test performance alerting with different scenarios
func TestPerformanceAlerting_AlertScenarios(t *testing.T) {
	pa := NewPerformanceAlerting(createMockDBForPerformanceAlerting())
	if pa.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test scenario 1: Critical alert
	alertID1, err := pa.GeneratePerformanceAlert(
		ctx,
		"CRITICAL_ALERT",
		"CRITICAL",
		"PERFORMANCE",
		"Critical Performance Issue",
		"System performance is critically degraded",
		"response_time_ms",
		nil, // metricValue
		nil, // thresholdValue
		"greater_than",
		[]string{"api", "database"},
		[]string{"Scale up resources", "Check system logs", "Contact support"},
	)
	if err != nil {
		t.Fatalf("Critical alert generation failed: %v", err)
	}

	// Test scenario 2: High alert
	alertID2, err := pa.GeneratePerformanceAlert(
		ctx,
		"HIGH_ALERT",
		"HIGH",
		"ACCURACY",
		"High Accuracy Issue",
		"Classification accuracy is below threshold",
		"accuracy_percentage",
		nil, // metricValue
		nil, // thresholdValue
		"less_than",
		[]string{"classification"},
		[]string{"Review algorithms", "Check training data"},
	)
	if err != nil {
		t.Fatalf("High alert generation failed: %v", err)
	}

	// Test scenario 3: Medium alert
	alertID3, err := pa.GeneratePerformanceAlert(
		ctx,
		"MEDIUM_ALERT",
		"MEDIUM",
		"RESOURCES",
		"Medium Resource Issue",
		"Resource usage is elevated",
		"cpu_usage_percentage",
		nil, // metricValue
		nil, // thresholdValue
		"greater_than",
		[]string{"system"},
		[]string{"Monitor usage", "Consider optimization"},
	)
	if err != nil {
		t.Fatalf("Medium alert generation failed: %v", err)
	}

	t.Logf("Alert scenarios: Critical=%s, High=%s, Medium=%s",
		alertID1, alertID2, alertID3)
}
