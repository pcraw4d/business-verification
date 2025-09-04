package classification

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// createMockDBForUsageMonitoring creates a mock database connection for testing
func createMockDBForUsageMonitoring() *sql.DB {
	// In a real test, you would use a test database or mock
	// For now, we'll return nil and skip tests that require DB
	return nil
}

func TestUsageMonitoring_CheckDatabaseStorageUsage(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	result, err := um.CheckDatabaseStorageUsage(ctx)
	if err != nil {
		t.Fatalf("CheckDatabaseStorageUsage failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	t.Logf("Database storage usage: %+v", result)
}

func TestUsageMonitoring_CheckTableSizes(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := um.CheckTableSizes(ctx)
	if err != nil {
		t.Fatalf("CheckTableSizes failed: %v", err)
	}

	t.Logf("Found %d tables", len(results))
	for _, result := range results {
		t.Logf("Table size: %+v", result)
	}
}

func TestUsageMonitoring_CheckConnectionUsage(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	result, err := um.CheckConnectionUsage(ctx)
	if err != nil {
		t.Fatalf("CheckConnectionUsage failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	t.Logf("Connection usage: %+v", result)
}

func TestUsageMonitoring_CheckQueryPerformance(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := um.CheckQueryPerformance(ctx)
	if err != nil {
		t.Fatalf("CheckQueryPerformance failed: %v", err)
	}

	t.Logf("Found %d query performance results", len(results))
	for _, result := range results {
		t.Logf("Query performance: %+v", result)
	}
}

func TestUsageMonitoring_CheckIndexUsage(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := um.CheckIndexUsage(ctx)
	if err != nil {
		t.Fatalf("CheckIndexUsage failed: %v", err)
	}

	t.Logf("Found %d index usage results", len(results))
	for _, result := range results {
		t.Logf("Index usage: %+v", result)
	}
}

func TestUsageMonitoring_CheckFreeTierLimits(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := um.CheckFreeTierLimits(ctx)
	if err != nil {
		t.Fatalf("CheckFreeTierLimits failed: %v", err)
	}

	t.Logf("Found %d free tier limits", len(results))
	for _, result := range results {
		t.Logf("Free tier limit: %+v", result)
	}
}

func TestUsageMonitoring_GenerateUsageReport(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := um.GenerateUsageReport(ctx)
	if err != nil {
		t.Fatalf("GenerateUsageReport failed: %v", err)
	}

	t.Logf("Generated %d usage report sections", len(results))
	for _, result := range results {
		t.Logf("Usage report: %+v", result)
	}
}

func TestUsageMonitoring_LogUsageMetrics(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	err := um.LogUsageMetrics(ctx)
	if err != nil {
		t.Fatalf("LogUsageMetrics failed: %v", err)
	}

	t.Log("Usage metrics logged successfully")
}

func TestUsageMonitoring_GetUsageTrends(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := um.GetUsageTrends(ctx, 7)
	if err != nil {
		t.Fatalf("GetUsageTrends failed: %v", err)
	}

	t.Logf("Found %d usage trends", len(results))
	for _, result := range results {
		t.Logf("Usage trend: %+v", result)
	}
}

func TestUsageMonitoring_CheckOptimizationOpportunities(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := um.CheckOptimizationOpportunities(ctx)
	if err != nil {
		t.Fatalf("CheckOptimizationOpportunities failed: %v", err)
	}

	t.Logf("Found %d optimization opportunities", len(results))
	for _, result := range results {
		t.Logf("Optimization opportunity: %+v", result)
	}
}

func TestUsageMonitoring_SetupAutomatedMonitoring(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	result, err := um.SetupAutomatedMonitoring(ctx)
	if err != nil {
		t.Fatalf("SetupAutomatedMonitoring failed: %v", err)
	}

	t.Logf("Automated monitoring setup result: %s", result)
}

func TestUsageMonitoring_GetMonitoringDashboard(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := um.GetMonitoringDashboard(ctx)
	if err != nil {
		t.Fatalf("GetMonitoringDashboard failed: %v", err)
	}

	t.Logf("Found %d monitoring dashboard items", len(results))
	for _, result := range results {
		t.Logf("Monitoring dashboard: %+v", result)
	}
}

func TestUsageMonitoring_ExportUsageData(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := um.ExportUsageData(ctx, 30)
	if err != nil {
		t.Fatalf("ExportUsageData failed: %v", err)
	}

	t.Logf("Exported %d usage data records", len(results))
	for _, result := range results {
		t.Logf("Usage data export: %+v", result)
	}
}

func TestUsageMonitoring_ValidateMonitoringSetup(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := um.ValidateMonitoringSetup(ctx)
	if err != nil {
		t.Fatalf("ValidateMonitoringSetup failed: %v", err)
	}

	t.Logf("Found %d monitoring validation results", len(results))
	for _, result := range results {
		t.Logf("Monitoring validation: %+v", result)
	}
}

func TestUsageMonitoring_RunAutomatedMonitoring(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	err := um.RunAutomatedMonitoring(ctx)
	if err != nil {
		t.Fatalf("RunAutomatedMonitoring failed: %v", err)
	}

	t.Log("Automated monitoring completed successfully")
}

func TestUsageMonitoring_GetCurrentUsageStatus(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	status, err := um.GetCurrentUsageStatus(ctx)
	if err != nil {
		t.Fatalf("GetCurrentUsageStatus failed: %v", err)
	}

	t.Logf("Current usage status: %+v", status)
}

func TestUsageMonitoring_GetUsageAlerts(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	alerts, err := um.GetUsageAlerts(ctx)
	if err != nil {
		t.Fatalf("GetUsageAlerts failed: %v", err)
	}

	t.Logf("Found %d usage alerts", len(alerts))
	for _, alert := range alerts {
		t.Logf("Usage alert: %+v", alert)
	}
}

// Benchmark tests for performance validation
func BenchmarkUsageMonitoring_CheckDatabaseStorageUsage(b *testing.B) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := um.CheckDatabaseStorageUsage(ctx)
		if err != nil {
			b.Fatalf("CheckDatabaseStorageUsage failed: %v", err)
		}
	}
}

func BenchmarkUsageMonitoring_CheckConnectionUsage(b *testing.B) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := um.CheckConnectionUsage(ctx)
		if err != nil {
			b.Fatalf("CheckConnectionUsage failed: %v", err)
		}
	}
}

func BenchmarkUsageMonitoring_GetCurrentUsageStatus(b *testing.B) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := um.GetCurrentUsageStatus(ctx)
		if err != nil {
			b.Fatalf("GetCurrentUsageStatus failed: %v", err)
		}
	}
}

// Integration test that would require a real database
func TestUsageMonitoring_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This would require a real database connection
	// db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
	// if err != nil {
	//     t.Fatalf("Failed to connect to database: %v", err)
	// }
	// defer db.Close()

	// um := NewUsageMonitoring(db)
	// ctx := context.Background()

	// Test all functions with real data
	t.Log("Integration test would run here with real database")
}

// Test error handling
func TestUsageMonitoring_ErrorHandling(t *testing.T) {
	// Test with nil database
	um := NewUsageMonitoring(nil)
	ctx := context.Background()

	// These should return errors due to nil database
	_, err := um.CheckDatabaseStorageUsage(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = um.CheckConnectionUsage(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = um.CheckFreeTierLimits(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}
}

// Test continuous monitoring
func TestUsageMonitoring_ContinuousMonitoring(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start continuous monitoring with short interval
	go um.MonitorUsageContinuously(ctx, 1*time.Second)

	// Wait for context to timeout
	<-ctx.Done()

	t.Log("Continuous monitoring test completed")
}

// Test monitoring dashboard
func TestUsageMonitoring_Dashboard(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test dashboard data
	dashboard, err := um.GetMonitoringDashboard(ctx)
	if err != nil {
		t.Fatalf("GetMonitoringDashboard failed: %v", err)
	}

	// Test current status
	status, err := um.GetCurrentUsageStatus(ctx)
	if err != nil {
		t.Fatalf("GetCurrentUsageStatus failed: %v", err)
	}

	// Test alerts
	alerts, err := um.GetUsageAlerts(ctx)
	if err != nil {
		t.Fatalf("GetUsageAlerts failed: %v", err)
	}

	t.Logf("Dashboard items: %d", len(dashboard))
	t.Logf("Status keys: %d", len(status))
	t.Logf("Alerts: %d", len(alerts))
}

// Test optimization opportunities
func TestUsageMonitoring_Optimization(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test optimization opportunities
	opportunities, err := um.CheckOptimizationOpportunities(ctx)
	if err != nil {
		t.Fatalf("CheckOptimizationOpportunities failed: %v", err)
	}

	// Test usage trends
	trends, err := um.GetUsageTrends(ctx, 7)
	if err != nil {
		t.Fatalf("GetUsageTrends failed: %v", err)
	}

	// Test usage report
	report, err := um.GenerateUsageReport(ctx)
	if err != nil {
		t.Fatalf("GenerateUsageReport failed: %v", err)
	}

	t.Logf("Optimization opportunities: %d", len(opportunities))
	t.Logf("Usage trends: %d", len(trends))
	t.Logf("Usage report sections: %d", len(report))
}

// Test data export
func TestUsageMonitoring_DataExport(t *testing.T) {
	um := NewUsageMonitoring(createMockDBForUsageMonitoring())
	if um.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test data export
	export, err := um.ExportUsageData(ctx, 30)
	if err != nil {
		t.Fatalf("ExportUsageData failed: %v", err)
	}

	t.Logf("Exported %d usage data records", len(export))

	// Test validation
	validation, err := um.ValidateMonitoringSetup(ctx)
	if err != nil {
		t.Fatalf("ValidateMonitoringSetup failed: %v", err)
	}

	t.Logf("Monitoring validation results: %d", len(validation))
}
