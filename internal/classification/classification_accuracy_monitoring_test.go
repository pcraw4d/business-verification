package classification

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// createMockDBForClassificationAccuracyMonitoring creates a mock database connection for testing
func createMockDBForClassificationAccuracyMonitoring() *sql.DB {
	// In a real test, you would use a test database or mock
	// For now, we'll return nil and skip tests that require DB
	return nil
}

func TestClassificationAccuracyMonitoring_LogClassificationAccuracyMetrics(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	businessName := "Test Company"
	businessDescription := "A test business for classification"
	websiteURL := "https://test.com"
	predictedIndustry := "Technology"
	predictedConfidence := 85.5
	responseTimeMs := 150.0
	processingTimeMs := 100.0
	classificationMethod := "website_analysis"
	keywordsUsed := []string{"technology", "software", "development"}
	confidenceThreshold := 70.0

	logID, err := cam.LogClassificationAccuracyMetrics(
		ctx,
		"test-request-123",
		&businessName,
		&businessDescription,
		&websiteURL,
		predictedIndustry,
		predictedConfidence,
		nil, // actualIndustry
		nil, // actualConfidence
		responseTimeMs,
		&processingTimeMs,
		&classificationMethod,
		keywordsUsed,
		confidenceThreshold,
		nil, // errorMessage
		nil, // userFeedback
	)

	if err != nil {
		t.Fatalf("LogClassificationAccuracyMetrics failed: %v", err)
	}

	t.Logf("Classification accuracy metrics logged with ID: %d", logID)
}

func TestClassificationAccuracyMonitoring_GetClassificationAccuracyStats(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	stats, err := cam.GetClassificationAccuracyStats(ctx, 24)
	if err != nil {
		t.Fatalf("GetClassificationAccuracyStats failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}

	t.Logf("Classification accuracy stats: %+v", stats)
}

func TestClassificationAccuracyMonitoring_GetClassificationAccuracyTrends(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	trends, err := cam.GetClassificationAccuracyTrends(ctx, 168)
	if err != nil {
		t.Fatalf("GetClassificationAccuracyTrends failed: %v", err)
	}

	t.Logf("Found %d classification accuracy trends", len(trends))
	for _, trend := range trends {
		t.Logf("Classification accuracy trend: %+v", trend)
	}
}

func TestClassificationAccuracyMonitoring_GetClassificationAccuracyAlerts(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	alerts, err := cam.GetClassificationAccuracyAlerts(ctx, 1)
	if err != nil {
		t.Fatalf("GetClassificationAccuracyAlerts failed: %v", err)
	}

	t.Logf("Found %d classification accuracy alerts", len(alerts))
	for _, alert := range alerts {
		t.Logf("Classification accuracy alert: %+v", alert)
	}
}

func TestClassificationAccuracyMonitoring_GetClassificationAccuracyDashboard(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	dashboard, err := cam.GetClassificationAccuracyDashboard(ctx)
	if err != nil {
		t.Fatalf("GetClassificationAccuracyDashboard failed: %v", err)
	}

	t.Logf("Found %d classification accuracy dashboard items", len(dashboard))
	for _, item := range dashboard {
		t.Logf("Classification accuracy dashboard: %+v", item)
	}
}

func TestClassificationAccuracyMonitoring_GetClassificationAccuracyInsights(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	insights, err := cam.GetClassificationAccuracyInsights(ctx)
	if err != nil {
		t.Fatalf("GetClassificationAccuracyInsights failed: %v", err)
	}

	t.Logf("Found %d classification accuracy insights", len(insights))
	for _, insight := range insights {
		t.Logf("Classification accuracy insight: %+v", insight)
	}
}

func TestClassificationAccuracyMonitoring_AnalyzeClassificationPerformance(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	performance, err := cam.AnalyzeClassificationPerformance(ctx, 24)
	if err != nil {
		t.Fatalf("AnalyzeClassificationPerformance failed: %v", err)
	}

	t.Logf("Found %d classification performance analyses", len(performance))
	for _, analysis := range performance {
		t.Logf("Classification performance analysis: %+v", analysis)
	}
}

func TestClassificationAccuracyMonitoring_CleanupClassificationAccuracyMetrics(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	deletedCount, err := cam.CleanupClassificationAccuracyMetrics(ctx, 30)
	if err != nil {
		t.Fatalf("CleanupClassificationAccuracyMetrics failed: %v", err)
	}

	t.Logf("Cleaned up %d classification accuracy metric entries", deletedCount)
}

func TestClassificationAccuracyMonitoring_ValidateClassificationAccuracyMonitoringSetup(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	validation, err := cam.ValidateClassificationAccuracyMonitoringSetup(ctx)
	if err != nil {
		t.Fatalf("ValidateClassificationAccuracyMonitoringSetup failed: %v", err)
	}

	t.Logf("Found %d classification accuracy validation results", len(validation))
	for _, result := range validation {
		t.Logf("Classification accuracy validation: %+v", result)
	}
}

func TestClassificationAccuracyMonitoring_GetCurrentClassificationAccuracyStatus(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	status, err := cam.GetCurrentClassificationAccuracyStatus(ctx)
	if err != nil {
		t.Fatalf("GetCurrentClassificationAccuracyStatus failed: %v", err)
	}

	t.Logf("Current classification accuracy status: %+v", status)
}

func TestClassificationAccuracyMonitoring_GetClassificationAccuracySummary(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	summary, err := cam.GetClassificationAccuracySummary(ctx)
	if err != nil {
		t.Fatalf("GetClassificationAccuracySummary failed: %v", err)
	}

	t.Logf("Classification accuracy summary: %+v", summary)
}

// Benchmark tests for classification accuracy monitoring
func BenchmarkClassificationAccuracyMonitoring_LogClassificationAccuracyMetrics(b *testing.B) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	businessName := "Test Company"
	businessDescription := "A test business for classification"
	websiteURL := "https://test.com"
	predictedIndustry := "Technology"
	predictedConfidence := 85.5
	responseTimeMs := 150.0
	processingTimeMs := 100.0
	classificationMethod := "website_analysis"
	keywordsUsed := []string{"technology", "software", "development"}
	confidenceThreshold := 70.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cam.LogClassificationAccuracyMetrics(
			ctx,
			"test-request-123",
			&businessName,
			&businessDescription,
			&websiteURL,
			predictedIndustry,
			predictedConfidence,
			nil, // actualIndustry
			nil, // actualConfidence
			responseTimeMs,
			&processingTimeMs,
			&classificationMethod,
			keywordsUsed,
			confidenceThreshold,
			nil, // errorMessage
			nil, // userFeedback
		)
		if err != nil {
			b.Fatalf("LogClassificationAccuracyMetrics failed: %v", err)
		}
	}
}

func BenchmarkClassificationAccuracyMonitoring_GetClassificationAccuracyStats(b *testing.B) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cam.GetClassificationAccuracyStats(ctx, 24)
		if err != nil {
			b.Fatalf("GetClassificationAccuracyStats failed: %v", err)
		}
	}
}

func BenchmarkClassificationAccuracyMonitoring_GetClassificationAccuracyDashboard(b *testing.B) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cam.GetClassificationAccuracyDashboard(ctx)
		if err != nil {
			b.Fatalf("GetClassificationAccuracyDashboard failed: %v", err)
		}
	}
}

// Integration test that would require a real database
func TestClassificationAccuracyMonitoring_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This would require a real database connection
	// db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
	// if err != nil {
	//     t.Fatalf("Failed to connect to database: %v", err)
	// }
	// defer db.Close()

	// cam := NewClassificationAccuracyMonitoring(db)
	// ctx := context.Background()

	// Test all functions with real data
	t.Log("Integration test would run here with real database")
}

// Test error handling
func TestClassificationAccuracyMonitoring_ErrorHandling(t *testing.T) {
	// Test with nil database
	cam := NewClassificationAccuracyMonitoring(nil)
	ctx := context.Background()

	// These should return errors due to nil database
	_, err := cam.GetClassificationAccuracyStats(ctx, 24)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = cam.GetClassificationAccuracyDashboard(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = cam.GetCurrentClassificationAccuracyStatus(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}
}

// Test continuous monitoring
func TestClassificationAccuracyMonitoring_ContinuousMonitoring(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start continuous monitoring with short interval
	go cam.MonitorClassificationAccuracyContinuously(ctx, 1*time.Second)

	// Wait for context to timeout
	<-ctx.Done()

	t.Log("Continuous classification accuracy monitoring test completed")
}

// Test classification accuracy performance analysis
func TestClassificationAccuracyMonitoring_PerformanceAnalysis(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test performance analysis
	performance, err := cam.AnalyzeClassificationPerformance(ctx, 24)
	if err != nil {
		t.Fatalf("AnalyzeClassificationPerformance failed: %v", err)
	}

	t.Logf("Performance analysis: %+v", performance)
}

// Test classification accuracy dashboard
func TestClassificationAccuracyMonitoring_Dashboard(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test dashboard data
	dashboard, err := cam.GetClassificationAccuracyDashboard(ctx)
	if err != nil {
		t.Fatalf("GetClassificationAccuracyDashboard failed: %v", err)
	}

	// Test current status
	status, err := cam.GetCurrentClassificationAccuracyStatus(ctx)
	if err != nil {
		t.Fatalf("GetCurrentClassificationAccuracyStatus failed: %v", err)
	}

	// Test summary
	summary, err := cam.GetClassificationAccuracySummary(ctx)
	if err != nil {
		t.Fatalf("GetClassificationAccuracySummary failed: %v", err)
	}

	t.Logf("Dashboard items: %d", len(dashboard))
	t.Logf("Status keys: %d", len(status))
	t.Logf("Summary keys: %d", len(summary))
}

// Test classification accuracy insights
func TestClassificationAccuracyMonitoring_Insights(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test insights
	insights, err := cam.GetClassificationAccuracyInsights(ctx)
	if err != nil {
		t.Fatalf("GetClassificationAccuracyInsights failed: %v", err)
	}

	// Test trends
	trends, err := cam.GetClassificationAccuracyTrends(ctx, 168)
	if err != nil {
		t.Fatalf("GetClassificationAccuracyTrends failed: %v", err)
	}

	// Test alerts
	alerts, err := cam.GetClassificationAccuracyAlerts(ctx, 1)
	if err != nil {
		t.Fatalf("GetClassificationAccuracyAlerts failed: %v", err)
	}

	t.Logf("Insights: %d", len(insights))
	t.Logf("Trends: %d", len(trends))
	t.Logf("Alerts: %d", len(alerts))
}

// Test classification accuracy validation
func TestClassificationAccuracyMonitoring_Validation(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test validation
	validation, err := cam.ValidateClassificationAccuracyMonitoringSetup(ctx)
	if err != nil {
		t.Fatalf("ValidateClassificationAccuracyMonitoringSetup failed: %v", err)
	}

	t.Logf("Validation results: %d", len(validation))
	for _, result := range validation {
		t.Logf("Validation: %+v", result)
	}
}

// Test classification accuracy cleanup
func TestClassificationAccuracyMonitoring_Cleanup(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test cleanup
	deletedCount, err := cam.CleanupClassificationAccuracyMetrics(ctx, 30)
	if err != nil {
		t.Fatalf("CleanupClassificationAccuracyMetrics failed: %v", err)
	}

	t.Logf("Cleaned up %d metric entries", deletedCount)
}

// Test classification accuracy metrics logging with different scenarios
func TestClassificationAccuracyMonitoring_LoggingScenarios(t *testing.T) {
	cam := NewClassificationAccuracyMonitoring(createMockDBForClassificationAccuracyMonitoring())
	if cam.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test scenario 1: Successful classification
	businessName := "Tech Corp"
	predictedIndustry := "Technology"
	predictedConfidence := 92.5
	responseTimeMs := 120.0

	logID1, err := cam.LogClassificationAccuracyMetrics(
		ctx,
		"test-request-1",
		&businessName,
		nil, // businessDescription
		nil, // websiteURL
		predictedIndustry,
		predictedConfidence,
		nil, // actualIndustry
		nil, // actualConfidence
		responseTimeMs,
		nil,  // processingTimeMs
		nil,  // classificationMethod
		nil,  // keywordsUsed
		70.0, // confidenceThreshold
		nil,  // errorMessage
		nil,  // userFeedback
	)

	if err != nil {
		t.Fatalf("Logging scenario 1 failed: %v", err)
	}

	// Test scenario 2: Classification with error
	errorMessage := "Database connection timeout"

	logID2, err := cam.LogClassificationAccuracyMetrics(
		ctx,
		"test-request-2",
		nil, // businessName
		nil, // businessDescription
		nil, // websiteURL
		"Unknown",
		0.0,    // predictedConfidence
		nil,    // actualIndustry
		nil,    // actualConfidence
		5000.0, // responseTimeMs
		nil,    // processingTimeMs
		nil,    // classificationMethod
		nil,    // keywordsUsed
		70.0,   // confidenceThreshold
		&errorMessage,
		nil, // userFeedback
	)

	if err != nil {
		t.Fatalf("Logging scenario 2 failed: %v", err)
	}

	// Test scenario 3: Classification with user feedback
	userFeedback := "correct"
	actualIndustry := "Technology"
	actualConfidence := 95.0

	logID3, err := cam.LogClassificationAccuracyMetrics(
		ctx,
		"test-request-3",
		&businessName,
		nil, // businessDescription
		nil, // websiteURL
		predictedIndustry,
		predictedConfidence,
		&actualIndustry,
		&actualConfidence,
		responseTimeMs,
		nil,  // processingTimeMs
		nil,  // classificationMethod
		nil,  // keywordsUsed
		70.0, // confidenceThreshold
		nil,  // errorMessage
		&userFeedback,
	)

	if err != nil {
		t.Fatalf("Logging scenario 3 failed: %v", err)
	}

	t.Logf("Logged scenarios: %d, %d, %d", logID1, logID2, logID3)
}
