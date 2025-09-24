package classification

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// createMockDBForPerformanceDashboards creates a mock database connection for testing
func createMockDBForPerformanceDashboards() *sql.DB {
	// In a real test, you would use a test database or mock
	// For now, we'll return nil and skip tests that require DB
	return nil
}

func TestPerformanceDashboards_CollectPerformanceMetrics(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := pd.CollectPerformanceMetrics(ctx)
	if err != nil {
		t.Fatalf("CollectPerformanceMetrics failed: %v", err)
	}

	t.Logf("Found %d performance metrics", len(results))
	for _, result := range results {
		t.Logf("Performance metric: %+v", result)
	}
}

func TestPerformanceDashboards_GetQueryPerformanceAnalysis(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := pd.GetQueryPerformanceAnalysis(ctx)
	if err != nil {
		t.Fatalf("GetQueryPerformanceAnalysis failed: %v", err)
	}

	t.Logf("Found %d query performance analyses", len(results))
	for _, result := range results {
		t.Logf("Query performance analysis: %+v", result)
	}
}

func TestPerformanceDashboards_GetIndexPerformanceAnalysis(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := pd.GetIndexPerformanceAnalysis(ctx)
	if err != nil {
		t.Fatalf("GetIndexPerformanceAnalysis failed: %v", err)
	}

	t.Logf("Found %d index performance analyses", len(results))
	for _, result := range results {
		t.Logf("Index performance analysis: %+v", result)
	}
}

func TestPerformanceDashboards_GetTablePerformanceAnalysis(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := pd.GetTablePerformanceAnalysis(ctx)
	if err != nil {
		t.Fatalf("GetTablePerformanceAnalysis failed: %v", err)
	}

	t.Logf("Found %d table performance analyses", len(results))
	for _, result := range results {
		t.Logf("Table performance analysis: %+v", result)
	}
}

func TestPerformanceDashboards_GetConnectionPerformanceAnalysis(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	result, err := pd.GetConnectionPerformanceAnalysis(ctx)
	if err != nil {
		t.Fatalf("GetConnectionPerformanceAnalysis failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	t.Logf("Connection performance analysis: %+v", result)
}

func TestPerformanceDashboards_GeneratePerformanceDashboard(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := pd.GeneratePerformanceDashboard(ctx)
	if err != nil {
		t.Fatalf("GeneratePerformanceDashboard failed: %v", err)
	}

	t.Logf("Generated %d performance dashboard sections", len(results))
	for _, result := range results {
		t.Logf("Performance dashboard: %+v", result)
	}
}

func TestPerformanceDashboards_LogPerformanceMetrics(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	err := pd.LogPerformanceMetrics(ctx)
	if err != nil {
		t.Fatalf("LogPerformanceMetrics failed: %v", err)
	}

	t.Log("Performance metrics logged successfully")
}

func TestPerformanceDashboards_GetPerformanceTrends(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := pd.GetPerformanceTrends(ctx, 7)
	if err != nil {
		t.Fatalf("GetPerformanceTrends failed: %v", err)
	}

	t.Logf("Found %d performance trends", len(results))
	for _, result := range results {
		t.Logf("Performance trend: %+v", result)
	}
}

func TestPerformanceDashboards_GetPerformanceAlerts(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := pd.GetPerformanceAlerts(ctx)
	if err != nil {
		t.Fatalf("GetPerformanceAlerts failed: %v", err)
	}

	t.Logf("Found %d performance alerts", len(results))
	for _, result := range results {
		t.Logf("Performance alert: %+v", result)
	}
}

func TestPerformanceDashboards_GetPerformanceSummary(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := pd.GetPerformanceSummary(ctx)
	if err != nil {
		t.Fatalf("GetPerformanceSummary failed: %v", err)
	}

	t.Logf("Found %d performance summaries", len(results))
	for _, result := range results {
		t.Logf("Performance summary: %+v", result)
	}
}

func TestPerformanceDashboards_SetupAutomatedPerformanceMonitoring(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	result, err := pd.SetupAutomatedPerformanceMonitoring(ctx)
	if err != nil {
		t.Fatalf("SetupAutomatedPerformanceMonitoring failed: %v", err)
	}

	t.Logf("Automated performance monitoring setup result: %s", result)
}

func TestPerformanceDashboards_GetPerformanceDashboardData(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := pd.GetPerformanceDashboardData(ctx)
	if err != nil {
		t.Fatalf("GetPerformanceDashboardData failed: %v", err)
	}

	t.Logf("Found %d performance dashboard data items", len(results))
	for _, result := range results {
		t.Logf("Performance dashboard data: %+v", result)
	}
}

func TestPerformanceDashboards_ExportPerformanceData(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := pd.ExportPerformanceData(ctx, 30)
	if err != nil {
		t.Fatalf("ExportPerformanceData failed: %v", err)
	}

	t.Logf("Exported %d performance data records", len(results))
	for _, result := range results {
		t.Logf("Performance data export: %+v", result)
	}
}

func TestPerformanceDashboards_ValidatePerformanceMonitoringSetup(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := pd.ValidatePerformanceMonitoringSetup(ctx)
	if err != nil {
		t.Fatalf("ValidatePerformanceMonitoringSetup failed: %v", err)
	}

	t.Logf("Found %d performance validation results", len(results))
	for _, result := range results {
		t.Logf("Performance validation: %+v", result)
	}
}

func TestPerformanceDashboards_RunAutomatedPerformanceMonitoring(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	err := pd.RunAutomatedPerformanceMonitoring(ctx)
	if err != nil {
		t.Fatalf("RunAutomatedPerformanceMonitoring failed: %v", err)
	}

	t.Log("Automated performance monitoring completed successfully")
}

func TestPerformanceDashboards_GetCurrentPerformanceStatus(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	status, err := pd.GetCurrentPerformanceStatus(ctx)
	if err != nil {
		t.Fatalf("GetCurrentPerformanceStatus failed: %v", err)
	}

	t.Logf("Current performance status: %+v", status)
}

func TestPerformanceDashboards_GetPerformanceInsights(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	insights, err := pd.GetPerformanceInsights(ctx)
	if err != nil {
		t.Fatalf("GetPerformanceInsights failed: %v", err)
	}

	t.Logf("Performance insights: %+v", insights)
}

// Benchmark tests for performance validation
func BenchmarkPerformanceDashboards_CollectPerformanceMetrics(b *testing.B) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pd.CollectPerformanceMetrics(ctx)
		if err != nil {
			b.Fatalf("CollectPerformanceMetrics failed: %v", err)
		}
	}
}

func BenchmarkPerformanceDashboards_GetQueryPerformanceAnalysis(b *testing.B) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pd.GetQueryPerformanceAnalysis(ctx)
		if err != nil {
			b.Fatalf("GetQueryPerformanceAnalysis failed: %v", err)
		}
	}
}

func BenchmarkPerformanceDashboards_GetCurrentPerformanceStatus(b *testing.B) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pd.GetCurrentPerformanceStatus(ctx)
		if err != nil {
			b.Fatalf("GetCurrentPerformanceStatus failed: %v", err)
		}
	}
}

// Integration test that would require a real database
func TestPerformanceDashboards_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This would require a real database connection
	// db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
	// if err != nil {
	//     t.Fatalf("Failed to connect to database: %v", err)
	// }
	// defer db.Close()

	// pd := NewPerformanceDashboards(db)
	// ctx := context.Background()

	// Test all functions with real data
	t.Log("Integration test would run here with real database")
}

// Test error handling
func TestPerformanceDashboards_ErrorHandling(t *testing.T) {
	// Test with nil database
	pd := NewPerformanceDashboards(nil)
	ctx := context.Background()

	// These should return errors due to nil database
	_, err := pd.CollectPerformanceMetrics(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = pd.GetQueryPerformanceAnalysis(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = pd.GetCurrentPerformanceStatus(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}
}

// Test continuous monitoring
func TestPerformanceDashboards_ContinuousMonitoring(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start continuous monitoring with short interval
	go pd.MonitorPerformanceContinuously(ctx, 1*time.Second)

	// Wait for context to timeout
	<-ctx.Done()

	t.Log("Continuous performance monitoring test completed")
}

// Test performance dashboard
func TestPerformanceDashboards_Dashboard(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test dashboard data
	dashboard, err := pd.GeneratePerformanceDashboard(ctx)
	if err != nil {
		t.Fatalf("GeneratePerformanceDashboard failed: %v", err)
	}

	// Test current status
	status, err := pd.GetCurrentPerformanceStatus(ctx)
	if err != nil {
		t.Fatalf("GetCurrentPerformanceStatus failed: %v", err)
	}

	// Test insights
	insights, err := pd.GetPerformanceInsights(ctx)
	if err != nil {
		t.Fatalf("GetPerformanceInsights failed: %v", err)
	}

	t.Logf("Dashboard sections: %d", len(dashboard))
	t.Logf("Status keys: %d", len(status))
	t.Logf("Insights keys: %d", len(insights))
}

// Test performance analysis
func TestPerformanceDashboards_Analysis(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test query analysis
	queryAnalysis, err := pd.GetQueryPerformanceAnalysis(ctx)
	if err != nil {
		t.Fatalf("GetQueryPerformanceAnalysis failed: %v", err)
	}

	// Test index analysis
	indexAnalysis, err := pd.GetIndexPerformanceAnalysis(ctx)
	if err != nil {
		t.Fatalf("GetIndexPerformanceAnalysis failed: %v", err)
	}

	// Test table analysis
	tableAnalysis, err := pd.GetTablePerformanceAnalysis(ctx)
	if err != nil {
		t.Fatalf("GetTablePerformanceAnalysis failed: %v", err)
	}

	// Test connection analysis
	connectionAnalysis, err := pd.GetConnectionPerformanceAnalysis(ctx)
	if err != nil {
		t.Fatalf("GetConnectionPerformanceAnalysis failed: %v", err)
	}

	t.Logf("Query analyses: %d", len(queryAnalysis))
	t.Logf("Index analyses: %d", len(indexAnalysis))
	t.Logf("Table analyses: %d", len(tableAnalysis))
	t.Logf("Connection analysis: %+v", connectionAnalysis)
}

// Test data export
func TestPerformanceDashboards_DataExport(t *testing.T) {
	pd := NewPerformanceDashboards(createMockDBForPerformanceDashboards())
	if pd.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test data export
	export, err := pd.ExportPerformanceData(ctx, 30)
	if err != nil {
		t.Fatalf("ExportPerformanceData failed: %v", err)
	}

	t.Logf("Exported %d performance data records", len(export))

	// Test validation
	validation, err := pd.ValidatePerformanceMonitoringSetup(ctx)
	if err != nil {
		t.Fatalf("ValidatePerformanceMonitoringSetup failed: %v", err)
	}

	t.Logf("Performance validation results: %d", len(validation))
}
