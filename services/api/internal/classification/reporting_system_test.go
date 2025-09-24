package classification

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// MockDB is a mock database for testing
type MockDB struct{}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// Return a mock row for testing
	return &sql.Row{}
}

func (m *MockDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// Return mock rows for testing
	return &sql.Rows{}, nil
}

// Implement sql.DB interface methods needed for the services
func (m *MockDB) Close() error                                               { return nil }
func (m *MockDB) Ping() error                                                { return nil }
func (m *MockDB) PingContext(ctx context.Context) error                      { return nil }
func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) { return nil, nil }
func (m *MockDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error)        { return &sql.Rows{}, nil }
func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row               { return &sql.Row{} }
func (m *MockDB) Begin() (*sql.Tx, error)                                           { return nil, nil }
func (m *MockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) { return nil, nil }
func (m *MockDB) Driver() driver.Driver                                             { return nil }
func (m *MockDB) SetConnMaxLifetime(d time.Duration)                                {}
func (m *MockDB) SetMaxIdleConns(n int)                                             {}
func (m *MockDB) SetMaxOpenConns(n int)                                             {}
func (m *MockDB) Stats() sql.DBStats                                                { return sql.DBStats{} }

// Helper function to create a mock database connection
func createMockDB() *sql.DB {
	// In a real test, you would use a test database or mock
	// For now, we'll skip the database-dependent tests
	return nil
}

func TestAccuracyReportingService_GenerateAccuracyReport(t *testing.T) {
	t.Skip("Skipping database-dependent test - requires real database connection")

	logger := zaptest.NewLogger(t)
	mockDB := createMockDB()

	if mockDB == nil {
		t.Skip("Mock database not available")
	}

	service := NewAccuracyReportingService(mockDB, logger)

	ctx := context.Background()
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	// Test report generation
	report, err := service.GenerateAccuracyReport(ctx, startTime, endTime)
	if err != nil {
		t.Fatalf("Failed to generate accuracy report: %v", err)
	}

	// Validate report structure
	if report.ID == "" {
		t.Error("Report ID should not be empty")
	}

	if report.Title == "" {
		t.Error("Report title should not be empty")
	}

	if report.GeneratedAt.IsZero() {
		t.Error("Report generated time should not be zero")
	}

	// Validate period
	if report.Period.StartTime.IsZero() || report.Period.EndTime.IsZero() {
		t.Error("Report period should have valid start and end times")
	}

	// Validate overall metrics
	if report.OverallMetrics.TotalClassifications < 0 {
		t.Error("Total classifications should not be negative")
	}

	// Validate industry metrics
	if len(report.IndustryMetrics) == 0 {
		t.Error("Industry metrics should not be empty")
	}

	// Validate confidence metrics
	if report.ConfidenceMetrics.CalibrationScore < 0 || report.ConfidenceMetrics.CalibrationScore > 1 {
		t.Error("Calibration score should be between 0 and 1")
	}

	// Validate performance metrics
	if report.PerformanceMetrics.AverageResponseTime < 0 {
		t.Error("Average response time should not be negative")
	}

	// Validate security metrics
	if report.SecurityMetrics.TrustedDataSourceRate < 0 || report.SecurityMetrics.TrustedDataSourceRate > 1 {
		t.Error("Trusted data source rate should be between 0 and 1")
	}

	// Validate trends
	if len(report.Trends) == 0 {
		t.Error("Trends should not be empty")
	}

	// Validate recommendations
	if len(report.Recommendations) == 0 {
		t.Error("Recommendations should not be empty")
	}

	// Validate metadata
	if report.Metadata == nil {
		t.Error("Metadata should not be nil")
	}

	t.Logf("Generated report with ID: %s", report.ID)
	t.Logf("Overall accuracy: %.2f%%", report.OverallMetrics.OverallAccuracy*100)
	t.Logf("Total classifications: %d", report.OverallMetrics.TotalClassifications)
	t.Logf("Industries analyzed: %d", len(report.IndustryMetrics))
}

func TestMetricsExportService_ExportMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewMetricsExportService(logger)

	ctx := context.Background()

	// Create a mock accuracy report
	report := &AccuracyReport{
		ID:          "test_report_123",
		Title:       "Test Accuracy Report",
		GeneratedAt: time.Now(),
		Period: ReportPeriod{
			StartTime: time.Now().Add(-24 * time.Hour),
			EndTime:   time.Now(),
			Duration:  "24h",
		},
		OverallMetrics: OverallAccuracyMetrics{
			TotalClassifications:   1000,
			CorrectClassifications: 850,
			OverallAccuracy:        0.85,
			AverageConfidence:      0.78,
		},
		IndustryMetrics: []IndustryMetrics{
			{
				IndustryName:         "Technology",
				TotalClassifications: 250,
				Accuracy:             0.88,
				AverageConfidence:    0.82,
			},
		},
		ConfidenceMetrics: ConfidenceMetrics{
			CalibrationScore: 0.75,
		},
		PerformanceMetrics: AccuracyPerformanceMetrics{
			AverageResponseTime: 1200,
			CacheHitRate:        0.85,
		},
		SecurityMetrics: SecurityMetrics{
			TrustedDataSourceRate: 1.0,
		},
		Trends: []TrendAnalysis{
			{
				MetricName:    "Overall Accuracy",
				CurrentValue:  0.85,
				PreviousValue: 0.82,
				Trend:         "improving",
			},
		},
		Recommendations: []string{
			"Continue monitoring system performance",
		},
		Metadata: map[string]interface{}{
			"report_version": "1.0.0",
		},
	}

	// Test JSON export
	exportRequest := &ExportRequest{
		Format:          ExportFormatJSON,
		ReportTypes:     []string{"accuracy", "performance", "security"},
		StartTime:       time.Now().Add(-24 * time.Hour),
		EndTime:         time.Now(),
		IncludeMetadata: true,
	}

	exportResponse, err := service.ExportMetrics(ctx, exportRequest, report)
	if err != nil {
		t.Fatalf("Failed to export metrics: %v", err)
	}

	// Validate export response
	if exportResponse.ExportID == "" {
		t.Error("Export ID should not be empty")
	}

	if exportResponse.Format != ExportFormatJSON {
		t.Error("Export format should be JSON")
	}

	if exportResponse.FileName == "" {
		t.Error("File name should not be empty")
	}

	if exportResponse.FileSize <= 0 {
		t.Error("File size should be positive")
	}

	if exportResponse.RecordCount <= 0 {
		t.Error("Record count should be positive")
	}

	if exportResponse.GeneratedAt.IsZero() {
		t.Error("Generated time should not be zero")
	}

	if exportResponse.ExpiresAt.IsZero() {
		t.Error("Expires time should not be zero")
	}

	// Test CSV export
	exportRequest.Format = ExportFormatCSV
	exportResponse, err = service.ExportMetrics(ctx, exportRequest, report)
	if err != nil {
		t.Fatalf("Failed to export CSV metrics: %v", err)
	}

	if exportResponse.Format != ExportFormatCSV {
		t.Error("Export format should be CSV")
	}

	// Test XML export
	exportRequest.Format = ExportFormatXML
	exportResponse, err = service.ExportMetrics(ctx, exportRequest, report)
	if err != nil {
		t.Fatalf("Failed to export XML metrics: %v", err)
	}

	if exportResponse.Format != ExportFormatXML {
		t.Error("Export format should be XML")
	}

	t.Logf("Export completed successfully: %s", exportResponse.ExportID)
}

func TestMetricsExportService_ValidateExportRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewMetricsExportService(logger)

	// Test valid request
	validRequest := &ExportRequest{
		Format:      ExportFormatJSON,
		ReportTypes: []string{"accuracy", "performance"},
		StartTime:   time.Now().Add(-24 * time.Hour),
		EndTime:     time.Now(),
	}

	err := service.ValidateExportRequest(validRequest)
	if err != nil {
		t.Errorf("Valid request should not return error: %v", err)
	}

	// Test invalid format
	invalidRequest := &ExportRequest{
		Format:    "invalid_format",
		StartTime: time.Now().Add(-24 * time.Hour),
		EndTime:   time.Now(),
	}

	err = service.ValidateExportRequest(invalidRequest)
	if err == nil {
		t.Error("Invalid format should return error")
	}

	// Test invalid time range
	invalidTimeRequest := &ExportRequest{
		Format:    ExportFormatJSON,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(-24 * time.Hour), // End before start
	}

	err = service.ValidateExportRequest(invalidTimeRequest)
	if err == nil {
		t.Error("Invalid time range should return error")
	}

	// Test invalid report type
	invalidReportTypeRequest := &ExportRequest{
		Format:      ExportFormatJSON,
		ReportTypes: []string{"invalid_type"},
		StartTime:   time.Now().Add(-24 * time.Hour),
		EndTime:     time.Now(),
	}

	err = service.ValidateExportRequest(invalidReportTypeRequest)
	if err == nil {
		t.Error("Invalid report type should return error")
	}
}

func TestPerformanceDashboardService_GetDashboardData(t *testing.T) {
	t.Skip("Skipping database-dependent test - requires real database connection")

	logger := zaptest.NewLogger(t)
	mockDB := createMockDB()

	if mockDB == nil {
		t.Skip("Mock database not available")
	}

	service := NewPerformanceDashboardService(mockDB, logger)

	ctx := context.Background()

	// Test dashboard data generation
	dashboard, err := service.GetDashboardData(ctx)
	if err != nil {
		t.Fatalf("Failed to get dashboard data: %v", err)
	}

	// Validate dashboard structure
	if dashboard.ID == "" {
		t.Error("Dashboard ID should not be empty")
	}

	if dashboard.Title == "" {
		t.Error("Dashboard title should not be empty")
	}

	if dashboard.LastUpdated.IsZero() {
		t.Error("Dashboard last updated should not be zero")
	}

	if dashboard.RefreshInterval <= 0 {
		t.Error("Refresh interval should be positive")
	}

	// Validate overall status
	if dashboard.OverallStatus.Status == "" {
		t.Error("Overall status should not be empty")
	}

	if dashboard.OverallStatus.HealthScore < 0 || dashboard.OverallStatus.HealthScore > 100 {
		t.Error("Health score should be between 0 and 100")
	}

	// Validate real-time metrics
	if dashboard.RealTimeMetrics.LastUpdated.IsZero() {
		t.Error("Real-time metrics last updated should not be zero")
	}

	// Validate accuracy overview
	if dashboard.AccuracyOverview.AccuracyTarget <= 0 {
		t.Error("Accuracy target should be positive")
	}

	// Validate performance overview
	if dashboard.PerformanceOverview.PerformanceTarget <= 0 {
		t.Error("Performance target should be positive")
	}

	// Validate security overview
	if dashboard.SecurityOverview.SecurityTarget <= 0 {
		t.Error("Security target should be positive")
	}

	// Validate industry breakdown
	if len(dashboard.IndustryBreakdown) == 0 {
		t.Error("Industry breakdown should not be empty")
	}

	// Validate trends
	if len(dashboard.Trends) == 0 {
		t.Error("Trends should not be empty")
	}

	// Validate recommendations
	if len(dashboard.Recommendations) == 0 {
		t.Error("Recommendations should not be empty")
	}

	// Validate metadata
	if dashboard.Metadata == nil {
		t.Error("Metadata should not be nil")
	}

	t.Logf("Generated dashboard with ID: %s", dashboard.ID)
	t.Logf("Overall status: %s", dashboard.OverallStatus.Status)
	t.Logf("Health score: %.2f", dashboard.OverallStatus.HealthScore)
	t.Logf("Active alerts: %d", dashboard.OverallStatus.ActiveAlerts)
}

func TestSecurityComplianceReportingService_GenerateSecurityComplianceReport(t *testing.T) {
	t.Skip("Skipping database-dependent test - requires real database connection")

	logger := zaptest.NewLogger(t)
	mockDB := createMockDB()

	if mockDB == nil {
		t.Skip("Mock database not available")
	}

	service := NewSecurityComplianceReportingService(mockDB, logger)

	ctx := context.Background()
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	// Test security compliance report generation
	report, err := service.GenerateSecurityComplianceReport(ctx, startTime, endTime)
	if err != nil {
		t.Fatalf("Failed to generate security compliance report: %v", err)
	}

	// Validate report structure
	if report.ID == "" {
		t.Error("Report ID should not be empty")
	}

	if report.Title == "" {
		t.Error("Report title should not be empty")
	}

	if report.GeneratedAt.IsZero() {
		t.Error("Report generated time should not be zero")
	}

	// Validate compliance status
	if report.ComplianceStatus.OverallStatus == "" {
		t.Error("Compliance status should not be empty")
	}

	if report.ComplianceStatus.ComplianceScore < 0 || report.ComplianceStatus.ComplianceScore > 1 {
		t.Error("Compliance score should be between 0 and 1")
	}

	// Validate data source trust metrics
	if report.DataSourceTrust.TrustedDataSourceRate < 0 || report.DataSourceTrust.TrustedDataSourceRate > 1 {
		t.Error("Trusted data source rate should be between 0 and 1")
	}

	// Validate website verification metrics
	if report.WebsiteVerification.WebsiteVerificationRate < 0 || report.WebsiteVerification.WebsiteVerificationRate > 1 {
		t.Error("Website verification rate should be between 0 and 1")
	}

	// Validate security violation metrics
	if report.SecurityViolations.TotalViolations < 0 {
		t.Error("Total violations should not be negative")
	}

	// Validate trusted data usage metrics
	if report.TrustedDataUsage.TrustedDataPercentage < 0 || report.TrustedDataUsage.TrustedDataPercentage > 1 {
		t.Error("Trusted data percentage should be between 0 and 1")
	}

	// Validate compliance scores
	if report.ComplianceScores.OverallComplianceScore < 0 || report.ComplianceScores.OverallComplianceScore > 1 {
		t.Error("Overall compliance score should be between 0 and 1")
	}

	// Validate recommendations
	if len(report.Recommendations) == 0 {
		t.Error("Recommendations should not be empty")
	}

	// Validate audit trail
	if len(report.AuditTrail) == 0 {
		t.Error("Audit trail should not be empty")
	}

	// Validate metadata
	if report.Metadata == nil {
		t.Error("Metadata should not be nil")
	}

	t.Logf("Generated security compliance report with ID: %s", report.ID)
	t.Logf("Compliance status: %s", report.ComplianceStatus.OverallStatus)
	t.Logf("Compliance score: %.2f%%", report.ComplianceStatus.ComplianceScore*100)
	t.Logf("Trusted data source rate: %.2f%%", report.DataSourceTrust.TrustedDataSourceRate*100)
}

func TestReportingSystemIntegration(t *testing.T) {
	t.Skip("Skipping database-dependent test - requires real database connection")

	logger := zaptest.NewLogger(t)
	mockDB := createMockDB()

	if mockDB == nil {
		t.Skip("Mock database not available")
	}

	// Test integration between all reporting services
	accuracyService := NewAccuracyReportingService(mockDB, logger)
	exportService := NewMetricsExportService(logger)
	dashboardService := NewPerformanceDashboardService(mockDB, logger)
	securityService := NewSecurityComplianceReportingService(mockDB, logger)

	ctx := context.Background()
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	// Generate accuracy report
	accuracyReport, err := accuracyService.GenerateAccuracyReport(ctx, startTime, endTime)
	if err != nil {
		t.Fatalf("Failed to generate accuracy report: %v", err)
	}

	// Export the report
	exportRequest := &ExportRequest{
		Format:      ExportFormatJSON,
		ReportTypes: []string{"accuracy", "performance", "security"},
		StartTime:   startTime,
		EndTime:     endTime,
	}

	exportResponse, err := exportService.ExportMetrics(ctx, exportRequest, accuracyReport)
	if err != nil {
		t.Fatalf("Failed to export metrics: %v", err)
	}

	// Get dashboard data
	dashboard, err := dashboardService.GetDashboardData(ctx)
	if err != nil {
		t.Fatalf("Failed to get dashboard data: %v", err)
	}

	// Generate security compliance report
	securityReport, err := securityService.GenerateSecurityComplianceReport(ctx, startTime, endTime)
	if err != nil {
		t.Fatalf("Failed to generate security compliance report: %v", err)
	}

	// Validate integration
	if accuracyReport.ID == "" {
		t.Error("Accuracy report ID should not be empty")
	}

	if exportResponse.ExportID == "" {
		t.Error("Export response ID should not be empty")
	}

	if dashboard.ID == "" {
		t.Error("Dashboard ID should not be empty")
	}

	if securityReport.ID == "" {
		t.Error("Security report ID should not be empty")
	}

	// Validate data consistency
	if accuracyReport.OverallMetrics.OverallAccuracy < 0 || accuracyReport.OverallMetrics.OverallAccuracy > 1 {
		t.Error("Accuracy should be between 0 and 1")
	}

	if dashboard.OverallStatus.HealthScore < 0 || dashboard.OverallStatus.HealthScore > 100 {
		t.Error("Health score should be between 0 and 100")
	}

	if securityReport.ComplianceStatus.ComplianceScore < 0 || securityReport.ComplianceStatus.ComplianceScore > 1 {
		t.Error("Compliance score should be between 0 and 1")
	}

	t.Logf("Integration test completed successfully")
	t.Logf("Accuracy report ID: %s", accuracyReport.ID)
	t.Logf("Export response ID: %s", exportResponse.ExportID)
	t.Logf("Dashboard ID: %s", dashboard.ID)
	t.Logf("Security report ID: %s", securityReport.ID)
}

// Benchmark tests
func BenchmarkAccuracyReportingService_GenerateAccuracyReport(b *testing.B) {
	b.Skip("Skipping database-dependent benchmark - requires real database connection")

	logger := zap.NewNop()
	mockDB := createMockDB()

	if mockDB == nil {
		b.Skip("Mock database not available")
	}

	service := NewAccuracyReportingService(mockDB, logger)

	ctx := context.Background()
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GenerateAccuracyReport(ctx, startTime, endTime)
		if err != nil {
			b.Fatalf("Failed to generate accuracy report: %v", err)
		}
	}
}

func BenchmarkMetricsExportService_ExportMetrics(b *testing.B) {
	logger := zap.NewNop()
	service := NewMetricsExportService(logger)

	ctx := context.Background()

	// Create a mock report
	report := &AccuracyReport{
		ID:          "benchmark_report",
		Title:       "Benchmark Report",
		GeneratedAt: time.Now(),
		OverallMetrics: OverallAccuracyMetrics{
			TotalClassifications:   1000,
			CorrectClassifications: 850,
			OverallAccuracy:        0.85,
		},
		IndustryMetrics: []IndustryMetrics{
			{
				IndustryName:         "Technology",
				TotalClassifications: 250,
				Accuracy:             0.88,
			},
		},
	}

	exportRequest := &ExportRequest{
		Format:      ExportFormatJSON,
		ReportTypes: []string{"accuracy"},
		StartTime:   time.Now().Add(-24 * time.Hour),
		EndTime:     time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ExportMetrics(ctx, exportRequest, report)
		if err != nil {
			b.Fatalf("Failed to export metrics: %v", err)
		}
	}
}

func BenchmarkPerformanceDashboardService_GetDashboardData(b *testing.B) {
	b.Skip("Skipping database-dependent benchmark - requires real database connection")

	logger := zap.NewNop()
	mockDB := createMockDB()

	if mockDB == nil {
		b.Skip("Mock database not available")
	}

	service := NewPerformanceDashboardService(mockDB, logger)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetDashboardData(ctx)
		if err != nil {
			b.Fatalf("Failed to get dashboard data: %v", err)
		}
	}
}
