package classification

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// createMockDBForQueryPerformanceMonitoring creates a mock database connection for testing
func createMockDBForQueryPerformanceMonitoring() *sql.DB {
	// In a real test, you would use a test database or mock
	// For now, we'll return nil and skip tests that require DB
	return nil
}

func TestQueryPerformanceMonitoring_AnalyzeQueryPerformance(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	result, err := qpm.AnalyzeQueryPerformance(ctx, "SELECT * FROM users WHERE id = 1", 150.5, 1, 1)
	if err != nil {
		t.Fatalf("AnalyzeQueryPerformance failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	t.Logf("Query performance analysis: %+v", result)
}

func TestQueryPerformanceMonitoring_LogQueryPerformance(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	logID, err := qpm.LogQueryPerformance(ctx, "SELECT * FROM users WHERE id = 1", 150.5, 1, 1, nil, nil, nil)
	if err != nil {
		t.Fatalf("LogQueryPerformance failed: %v", err)
	}

	t.Logf("Query performance logged with ID: %d", logID)
}

func TestQueryPerformanceMonitoring_GetQueryPerformanceStats(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	stats, err := qpm.GetQueryPerformanceStats(ctx, 24)
	if err != nil {
		t.Fatalf("GetQueryPerformanceStats failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}

	t.Logf("Query performance stats: %+v", stats)
}

func TestQueryPerformanceMonitoring_GetQueryPerformanceTrends(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	trends, err := qpm.GetQueryPerformanceTrends(ctx, 168)
	if err != nil {
		t.Fatalf("GetQueryPerformanceTrends failed: %v", err)
	}

	t.Logf("Found %d query performance trends", len(trends))
	for _, trend := range trends {
		t.Logf("Query performance trend: %+v", trend)
	}
}

func TestQueryPerformanceMonitoring_GetQueryPerformanceAlerts(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	alerts, err := qpm.GetQueryPerformanceAlerts(ctx, 1)
	if err != nil {
		t.Fatalf("GetQueryPerformanceAlerts failed: %v", err)
	}

	t.Logf("Found %d query performance alerts", len(alerts))
	for _, alert := range alerts {
		t.Logf("Query performance alert: %+v", alert)
	}
}

func TestQueryPerformanceMonitoring_GetQueryPerformanceDashboard(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	dashboard, err := qpm.GetQueryPerformanceDashboard(ctx)
	if err != nil {
		t.Fatalf("GetQueryPerformanceDashboard failed: %v", err)
	}

	t.Logf("Found %d query performance dashboard items", len(dashboard))
	for _, item := range dashboard {
		t.Logf("Query performance dashboard: %+v", item)
	}
}

func TestQueryPerformanceMonitoring_CleanupQueryPerformanceLogs(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	deletedCount, err := qpm.CleanupQueryPerformanceLogs(ctx, 30)
	if err != nil {
		t.Fatalf("CleanupQueryPerformanceLogs failed: %v", err)
	}

	t.Logf("Cleaned up %d query performance log entries", deletedCount)
}

func TestQueryPerformanceMonitoring_GetQueryPerformanceInsights(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	insights, err := qpm.GetQueryPerformanceInsights(ctx)
	if err != nil {
		t.Fatalf("GetQueryPerformanceInsights failed: %v", err)
	}

	t.Logf("Found %d query performance insights", len(insights))
	for _, insight := range insights {
		t.Logf("Query performance insight: %+v", insight)
	}
}

func TestQueryPerformanceMonitoring_ValidateQueryPerformanceMonitoringSetup(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	validation, err := qpm.ValidateQueryPerformanceMonitoringSetup(ctx)
	if err != nil {
		t.Fatalf("ValidateQueryPerformanceMonitoringSetup failed: %v", err)
	}

	t.Logf("Found %d query performance validation results", len(validation))
	for _, result := range validation {
		t.Logf("Query performance validation: %+v", result)
	}
}

func TestQueryPerformanceMonitoring_GetCurrentQueryPerformanceStatus(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	status, err := qpm.GetCurrentQueryPerformanceStatus(ctx)
	if err != nil {
		t.Fatalf("GetCurrentQueryPerformanceStatus failed: %v", err)
	}

	t.Logf("Current query performance status: %+v", status)
}

func TestQueryPerformanceMonitoring_GetQueryPerformanceSummary(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	summary, err := qpm.GetQueryPerformanceSummary(ctx)
	if err != nil {
		t.Fatalf("GetQueryPerformanceSummary failed: %v", err)
	}

	t.Logf("Query performance summary: %+v", summary)
}

func TestQueryPerformanceMonitoring_AnalyzeSlowQueries(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	slowQueries, err := qpm.AnalyzeSlowQueries(ctx, 1000.0)
	if err != nil {
		t.Fatalf("AnalyzeSlowQueries failed: %v", err)
	}

	t.Logf("Found %d slow queries", len(slowQueries))
	for _, query := range slowQueries {
		t.Logf("Slow query: %+v", query)
	}
}

func TestQueryPerformanceMonitoring_GetQueryPerformanceMetrics(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	metrics, err := qpm.GetQueryPerformanceMetrics(ctx)
	if err != nil {
		t.Fatalf("GetQueryPerformanceMetrics failed: %v", err)
	}

	t.Logf("Query performance metrics: %+v", metrics)
}

// Benchmark tests for query performance monitoring
func BenchmarkQueryPerformanceMonitoring_AnalyzeQueryPerformance(b *testing.B) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := qpm.AnalyzeQueryPerformance(ctx, "SELECT * FROM users WHERE id = 1", 150.5, 1, 1)
		if err != nil {
			b.Fatalf("AnalyzeQueryPerformance failed: %v", err)
		}
	}
}

func BenchmarkQueryPerformanceMonitoring_LogQueryPerformance(b *testing.B) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := qpm.LogQueryPerformance(ctx, "SELECT * FROM users WHERE id = 1", 150.5, 1, 1, nil, nil, nil)
		if err != nil {
			b.Fatalf("LogQueryPerformance failed: %v", err)
		}
	}
}

func BenchmarkQueryPerformanceMonitoring_GetQueryPerformanceStats(b *testing.B) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := qpm.GetQueryPerformanceStats(ctx, 24)
		if err != nil {
			b.Fatalf("GetQueryPerformanceStats failed: %v", err)
		}
	}
}

// Integration test that would require a real database
func TestQueryPerformanceMonitoring_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This would require a real database connection
	// db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
	// if err != nil {
	//     t.Fatalf("Failed to connect to database: %v", err)
	// }
	// defer db.Close()

	// qpm := NewQueryPerformanceMonitoring(db)
	// ctx := context.Background()

	// Test all functions with real data
	t.Log("Integration test would run here with real database")
}

// Test error handling
func TestQueryPerformanceMonitoring_ErrorHandling(t *testing.T) {
	// Test with nil database
	qpm := NewQueryPerformanceMonitoring(nil)
	ctx := context.Background()

	// These should return errors due to nil database
	_, err := qpm.AnalyzeQueryPerformance(ctx, "SELECT * FROM users", 100, 1, 1)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = qpm.GetQueryPerformanceStats(ctx, 24)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = qpm.GetCurrentQueryPerformanceStatus(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}
}

// Test continuous monitoring
func TestQueryPerformanceMonitoring_ContinuousMonitoring(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start continuous monitoring with short interval
	go qpm.MonitorQueryPerformanceContinuously(ctx, 1*time.Second)

	// Wait for context to timeout
	<-ctx.Done()

	t.Log("Continuous query performance monitoring test completed")
}

// Test query performance analysis
func TestQueryPerformanceMonitoring_PerformanceAnalysis(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test performance analysis
	analysis, err := qpm.AnalyzeQueryPerformance(ctx, "SELECT * FROM users WHERE id = 1", 150.5, 1, 1)
	if err != nil {
		t.Fatalf("AnalyzeQueryPerformance failed: %v", err)
	}

	// Test slow query analysis
	slowQueries, err := qpm.AnalyzeSlowQueries(ctx, 1000.0)
	if err != nil {
		t.Fatalf("AnalyzeSlowQueries failed: %v", err)
	}

	// Test performance metrics
	metrics, err := qpm.GetQueryPerformanceMetrics(ctx)
	if err != nil {
		t.Fatalf("GetQueryPerformanceMetrics failed: %v", err)
	}

	t.Logf("Performance analysis: %+v", analysis)
	t.Logf("Slow queries: %d", len(slowQueries))
	t.Logf("Performance metrics: %+v", metrics)
}

// Test query performance dashboard
func TestQueryPerformanceMonitoring_Dashboard(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test dashboard data
	dashboard, err := qpm.GetQueryPerformanceDashboard(ctx)
	if err != nil {
		t.Fatalf("GetQueryPerformanceDashboard failed: %v", err)
	}

	// Test current status
	status, err := qpm.GetCurrentQueryPerformanceStatus(ctx)
	if err != nil {
		t.Fatalf("GetCurrentQueryPerformanceStatus failed: %v", err)
	}

	// Test summary
	summary, err := qpm.GetQueryPerformanceSummary(ctx)
	if err != nil {
		t.Fatalf("GetQueryPerformanceSummary failed: %v", err)
	}

	t.Logf("Dashboard items: %d", len(dashboard))
	t.Logf("Status keys: %d", len(status))
	t.Logf("Summary keys: %d", len(summary))
}

// Test query performance insights
func TestQueryPerformanceMonitoring_Insights(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test insights
	insights, err := qpm.GetQueryPerformanceInsights(ctx)
	if err != nil {
		t.Fatalf("GetQueryPerformanceInsights failed: %v", err)
	}

	// Test trends
	trends, err := qpm.GetQueryPerformanceTrends(ctx, 168)
	if err != nil {
		t.Fatalf("GetQueryPerformanceTrends failed: %v", err)
	}

	// Test alerts
	alerts, err := qpm.GetQueryPerformanceAlerts(ctx, 1)
	if err != nil {
		t.Fatalf("GetQueryPerformanceAlerts failed: %v", err)
	}

	t.Logf("Insights: %d", len(insights))
	t.Logf("Trends: %d", len(trends))
	t.Logf("Alerts: %d", len(alerts))
}

// Test query performance validation
func TestQueryPerformanceMonitoring_Validation(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test validation
	validation, err := qpm.ValidateQueryPerformanceMonitoringSetup(ctx)
	if err != nil {
		t.Fatalf("ValidateQueryPerformanceMonitoringSetup failed: %v", err)
	}

	t.Logf("Validation results: %d", len(validation))
	for _, result := range validation {
		t.Logf("Validation: %+v", result)
	}
}

// Test query performance cleanup
func TestQueryPerformanceMonitoring_Cleanup(t *testing.T) {
	qpm := NewQueryPerformanceMonitoring(createMockDBForQueryPerformanceMonitoring())
	if qpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test cleanup
	deletedCount, err := qpm.CleanupQueryPerformanceLogs(ctx, 30)
	if err != nil {
		t.Fatalf("CleanupQueryPerformanceLogs failed: %v", err)
	}

	t.Logf("Cleaned up %d log entries", deletedCount)
}
