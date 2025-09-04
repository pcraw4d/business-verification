package classification

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// createMockDBForConnectionPoolMonitoring creates a mock database connection for testing
func createMockDBForConnectionPoolMonitoring() *sql.DB {
	// In a real test, you would use a test database or mock
	// For now, we'll return nil and skip tests that require DB
	return nil
}

func TestConnectionPoolMonitoring_GetConnectionPoolStats(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	stats, err := cpm.GetConnectionPoolStats(ctx)
	if err != nil {
		t.Fatalf("GetConnectionPoolStats failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}

	t.Logf("Connection pool stats: %+v", stats)
}

func TestConnectionPoolMonitoring_LogConnectionPoolMetrics(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	stats := &ConnectionPoolStats{
		ActiveConnections:            10,
		IdleConnections:              5,
		TotalConnections:             15,
		MaxConnections:               100,
		ConnectionUtilization:        15.0,
		AvgConnectionDurationSeconds: 30.5,
		ConnectionErrors:             0,
		ConnectionTimeouts:           0,
		PoolHitRatio:                 95.0,
		PoolMissRatio:                5.0,
		AvgWaitTimeMs:                10.0,
		MaxWaitTimeMs:                50.0,
		ConnectionCreationRate:       0.1,
		ConnectionDestructionRate:    0.1,
		PoolStatus:                   "HEALTHY",
	}

	logID, err := cpm.LogConnectionPoolMetrics(ctx, stats, nil)
	if err != nil {
		t.Fatalf("LogConnectionPoolMetrics failed: %v", err)
	}

	t.Logf("Connection pool metrics logged with ID: %d", logID)
}

func TestConnectionPoolMonitoring_GetConnectionPoolTrends(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	trends, err := cpm.GetConnectionPoolTrends(ctx, 24)
	if err != nil {
		t.Fatalf("GetConnectionPoolTrends failed: %v", err)
	}

	t.Logf("Found %d connection pool trends", len(trends))
	for _, trend := range trends {
		t.Logf("Connection pool trend: %+v", trend)
	}
}

func TestConnectionPoolMonitoring_GetConnectionPoolAlerts(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	alerts, err := cpm.GetConnectionPoolAlerts(ctx, 1)
	if err != nil {
		t.Fatalf("GetConnectionPoolAlerts failed: %v", err)
	}

	t.Logf("Found %d connection pool alerts", len(alerts))
	for _, alert := range alerts {
		t.Logf("Connection pool alert: %+v", alert)
	}
}

func TestConnectionPoolMonitoring_GetConnectionPoolDashboard(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	dashboard, err := cpm.GetConnectionPoolDashboard(ctx)
	if err != nil {
		t.Fatalf("GetConnectionPoolDashboard failed: %v", err)
	}

	t.Logf("Found %d connection pool dashboard items", len(dashboard))
	for _, item := range dashboard {
		t.Logf("Connection pool dashboard: %+v", item)
	}
}

func TestConnectionPoolMonitoring_GetConnectionPoolInsights(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	insights, err := cpm.GetConnectionPoolInsights(ctx)
	if err != nil {
		t.Fatalf("GetConnectionPoolInsights failed: %v", err)
	}

	t.Logf("Found %d connection pool insights", len(insights))
	for _, insight := range insights {
		t.Logf("Connection pool insight: %+v", insight)
	}
}

func TestConnectionPoolMonitoring_OptimizeConnectionPoolSettings(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	optimizations, err := cpm.OptimizeConnectionPoolSettings(ctx)
	if err != nil {
		t.Fatalf("OptimizeConnectionPoolSettings failed: %v", err)
	}

	t.Logf("Found %d connection pool optimizations", len(optimizations))
	for _, optimization := range optimizations {
		t.Logf("Connection pool optimization: %+v", optimization)
	}
}

func TestConnectionPoolMonitoring_CleanupConnectionPoolMetrics(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	deletedCount, err := cpm.CleanupConnectionPoolMetrics(ctx, 30)
	if err != nil {
		t.Fatalf("CleanupConnectionPoolMetrics failed: %v", err)
	}

	t.Logf("Cleaned up %d connection pool metric entries", deletedCount)
}

func TestConnectionPoolMonitoring_ValidateConnectionPoolMonitoringSetup(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	validation, err := cpm.ValidateConnectionPoolMonitoringSetup(ctx)
	if err != nil {
		t.Fatalf("ValidateConnectionPoolMonitoringSetup failed: %v", err)
	}

	t.Logf("Found %d connection pool validation results", len(validation))
	for _, result := range validation {
		t.Logf("Connection pool validation: %+v", result)
	}
}

func TestConnectionPoolMonitoring_GetCurrentConnectionPoolStatus(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	status, err := cpm.GetCurrentConnectionPoolStatus(ctx)
	if err != nil {
		t.Fatalf("GetCurrentConnectionPoolStatus failed: %v", err)
	}

	t.Logf("Current connection pool status: %+v", status)
}

func TestConnectionPoolMonitoring_GetConnectionPoolSummary(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	summary, err := cpm.GetConnectionPoolSummary(ctx)
	if err != nil {
		t.Fatalf("GetConnectionPoolSummary failed: %v", err)
	}

	t.Logf("Connection pool summary: %+v", summary)
}

func TestConnectionPoolMonitoring_AnalyzeConnectionPoolPerformance(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	analysis, err := cpm.AnalyzeConnectionPoolPerformance(ctx)
	if err != nil {
		t.Fatalf("AnalyzeConnectionPoolPerformance failed: %v", err)
	}

	t.Logf("Connection pool performance analysis: %+v", analysis)
}

// Benchmark tests for connection pool monitoring
func BenchmarkConnectionPoolMonitoring_GetConnectionPoolStats(b *testing.B) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cpm.GetConnectionPoolStats(ctx)
		if err != nil {
			b.Fatalf("GetConnectionPoolStats failed: %v", err)
		}
	}
}

func BenchmarkConnectionPoolMonitoring_LogConnectionPoolMetrics(b *testing.B) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	stats := &ConnectionPoolStats{
		ActiveConnections:            10,
		IdleConnections:              5,
		TotalConnections:             15,
		MaxConnections:               100,
		ConnectionUtilization:        15.0,
		AvgConnectionDurationSeconds: 30.5,
		ConnectionErrors:             0,
		ConnectionTimeouts:           0,
		PoolHitRatio:                 95.0,
		PoolMissRatio:                5.0,
		AvgWaitTimeMs:                10.0,
		MaxWaitTimeMs:                50.0,
		ConnectionCreationRate:       0.1,
		ConnectionDestructionRate:    0.1,
		PoolStatus:                   "HEALTHY",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cpm.LogConnectionPoolMetrics(ctx, stats, nil)
		if err != nil {
			b.Fatalf("LogConnectionPoolMetrics failed: %v", err)
		}
	}
}

func BenchmarkConnectionPoolMonitoring_GetConnectionPoolDashboard(b *testing.B) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cpm.GetConnectionPoolDashboard(ctx)
		if err != nil {
			b.Fatalf("GetConnectionPoolDashboard failed: %v", err)
		}
	}
}

// Integration test that would require a real database
func TestConnectionPoolMonitoring_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This would require a real database connection
	// db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
	// if err != nil {
	//     t.Fatalf("Failed to connect to database: %v", err)
	// }
	// defer db.Close()

	// cpm := NewConnectionPoolMonitoring(db)
	// ctx := context.Background()

	// Test all functions with real data
	t.Log("Integration test would run here with real database")
}

// Test error handling
func TestConnectionPoolMonitoring_ErrorHandling(t *testing.T) {
	// Test with nil database
	cpm := NewConnectionPoolMonitoring(nil)
	ctx := context.Background()

	// These should return errors due to nil database
	_, err := cpm.GetConnectionPoolStats(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = cpm.GetConnectionPoolDashboard(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = cpm.GetCurrentConnectionPoolStatus(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}
}

// Test continuous monitoring
func TestConnectionPoolMonitoring_ContinuousMonitoring(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start continuous monitoring with short interval
	go cpm.MonitorConnectionPoolContinuously(ctx, 1*time.Second)

	// Wait for context to timeout
	<-ctx.Done()

	t.Log("Continuous connection pool monitoring test completed")
}

// Test connection pool performance analysis
func TestConnectionPoolMonitoring_PerformanceAnalysis(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test performance analysis
	analysis, err := cpm.AnalyzeConnectionPoolPerformance(ctx)
	if err != nil {
		t.Fatalf("AnalyzeConnectionPoolPerformance failed: %v", err)
	}

	// Test optimization settings
	optimizations, err := cpm.OptimizeConnectionPoolSettings(ctx)
	if err != nil {
		t.Fatalf("OptimizeConnectionPoolSettings failed: %v", err)
	}

	t.Logf("Performance analysis: %+v", analysis)
	t.Logf("Optimizations: %d", len(optimizations))
}

// Test connection pool dashboard
func TestConnectionPoolMonitoring_Dashboard(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test dashboard data
	dashboard, err := cpm.GetConnectionPoolDashboard(ctx)
	if err != nil {
		t.Fatalf("GetConnectionPoolDashboard failed: %v", err)
	}

	// Test current status
	status, err := cpm.GetCurrentConnectionPoolStatus(ctx)
	if err != nil {
		t.Fatalf("GetCurrentConnectionPoolStatus failed: %v", err)
	}

	// Test summary
	summary, err := cpm.GetConnectionPoolSummary(ctx)
	if err != nil {
		t.Fatalf("GetConnectionPoolSummary failed: %v", err)
	}

	t.Logf("Dashboard items: %d", len(dashboard))
	t.Logf("Status keys: %d", len(status))
	t.Logf("Summary keys: %d", len(summary))
}

// Test connection pool insights
func TestConnectionPoolMonitoring_Insights(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test insights
	insights, err := cpm.GetConnectionPoolInsights(ctx)
	if err != nil {
		t.Fatalf("GetConnectionPoolInsights failed: %v", err)
	}

	// Test trends
	trends, err := cpm.GetConnectionPoolTrends(ctx, 168)
	if err != nil {
		t.Fatalf("GetConnectionPoolTrends failed: %v", err)
	}

	// Test alerts
	alerts, err := cpm.GetConnectionPoolAlerts(ctx, 1)
	if err != nil {
		t.Fatalf("GetConnectionPoolAlerts failed: %v", err)
	}

	t.Logf("Insights: %d", len(insights))
	t.Logf("Trends: %d", len(trends))
	t.Logf("Alerts: %d", len(alerts))
}

// Test connection pool validation
func TestConnectionPoolMonitoring_Validation(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test validation
	validation, err := cpm.ValidateConnectionPoolMonitoringSetup(ctx)
	if err != nil {
		t.Fatalf("ValidateConnectionPoolMonitoringSetup failed: %v", err)
	}

	t.Logf("Validation results: %d", len(validation))
	for _, result := range validation {
		t.Logf("Validation: %+v", result)
	}
}

// Test connection pool cleanup
func TestConnectionPoolMonitoring_Cleanup(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(createMockDBForConnectionPoolMonitoring())
	if cpm.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test cleanup
	deletedCount, err := cpm.CleanupConnectionPoolMetrics(ctx, 30)
	if err != nil {
		t.Fatalf("CleanupConnectionPoolMetrics failed: %v", err)
	}

	t.Logf("Cleaned up %d metric entries", deletedCount)
}

// Test helper methods
func TestConnectionPoolMonitoring_HelperMethods(t *testing.T) {
	cpm := NewConnectionPoolMonitoring(nil)

	// Test utilization recommendation
	recommendation := cpm.getUtilizationRecommendation(95.0)
	if recommendation == "" {
		t.Error("Expected utilization recommendation, got empty string")
	}

	// Test hit ratio status
	status := cpm.getHitRatioStatus(85.0)
	if status == "" {
		t.Error("Expected hit ratio status, got empty string")
	}

	// Test wait time status
	waitStatus := cpm.getWaitTimeStatus(150.0)
	if waitStatus == "" {
		t.Error("Expected wait time status, got empty string")
	}

	// Test error status
	errorStatus := cpm.getErrorStatus(3)
	if errorStatus == "" {
		t.Error("Expected error status, got empty string")
	}

	// Test performance score calculation
	stats := &ConnectionPoolStats{
		ConnectionUtilization: 80.0,
		PoolHitRatio:          90.0,
		AvgWaitTimeMs:         50.0,
		ConnectionErrors:      1,
	}

	score := cpm.calculatePerformanceScore(stats)
	if score < 0 || score > 100 {
		t.Errorf("Expected performance score between 0-100, got %f", score)
	}

	// Test overall recommendation
	recommendation = cpm.getOverallRecommendation(score)
	if recommendation == "" {
		t.Error("Expected overall recommendation, got empty string")
	}

	t.Logf("Utilization recommendation: %s", cpm.getUtilizationRecommendation(95.0))
	t.Logf("Hit ratio status: %s", cpm.getHitRatioStatus(85.0))
	t.Logf("Wait time status: %s", cpm.getWaitTimeStatus(150.0))
	t.Logf("Error status: %s", cpm.getErrorStatus(3))
	t.Logf("Performance score: %f", score)
	t.Logf("Overall recommendation: %s", recommendation)
}
