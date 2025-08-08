package risk

import (
	"context"
	"strings"
	"testing"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
)

func TestRiskService_GenerateRiskReport(t *testing.T) {
	// Create test logger
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})

	// Create test components
	categoryRegistry := CreateDefaultRiskCategories()
	thresholdManager := CreateDefaultThresholds()
	industryModelRegistry := CreateDefaultIndustryModels()
	calculator := NewRiskFactorCalculator(categoryRegistry)
	scoringAlgorithm := NewWeightedScoringAlgorithm()
	predictionAlgorithm := NewRiskPredictionAlgorithm()
	historyService := NewRiskHistoryService(logger, nil) // nil DB for testing
	alertService := NewAlertService(logger, thresholdManager)
	reportService := NewReportService(logger, historyService, alertService)

	// Create risk service
	service := NewRiskService(
		logger,
		calculator,
		scoringAlgorithm,
		predictionAlgorithm,
		thresholdManager,
		categoryRegistry,
		industryModelRegistry,
		historyService,
		alertService,
		reportService,
		nil, // No export service for this test
	)

	// Create test request
	request := ReportRequest{
		BusinessID: "test_business_123",
		ReportType: ReportTypeSummary,
		Format:     ReportFormatJSON,
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	// Create test context
	ctx := context.WithValue(context.Background(), "request_id", "test_request_123")

	// Generate report
	report, err := service.GenerateRiskReport(ctx, request)
	if err != nil {
		t.Fatalf("Failed to generate risk report: %v", err)
	}

	// Verify report structure
	if report == nil {
		t.Fatal("Report should not be nil")
	}

	if report.ID == "" {
		t.Error("Report ID should not be empty")
	}

	if report.BusinessID != request.BusinessID {
		t.Errorf("Expected business ID %s, got %s", request.BusinessID, report.BusinessID)
	}

	if report.ReportType != request.ReportType {
		t.Errorf("Expected report type %s, got %s", request.ReportType, report.ReportType)
	}

	if report.Format != request.Format {
		t.Errorf("Expected format %s, got %s", request.Format, report.Format)
	}

	if report.GeneratedAt.IsZero() {
		t.Error("GeneratedAt should not be zero")
	}

	if report.ValidUntil.IsZero() {
		t.Error("ValidUntil should not be zero")
	}

	// Verify summary exists
	if report.Summary == nil {
		t.Error("Report summary should not be nil")
	}

	// Verify summary has required fields
	if report.Summary.OverallScore == 0 {
		t.Error("Overall score should not be zero")
	}

	if report.Summary.OverallLevel == "" {
		t.Error("Overall level should not be empty")
	}

	// Verify recommendations exist
	if len(report.Recommendations) == 0 {
		t.Error("Report should have recommendations")
	}

	// Verify metadata is preserved
	if report.Metadata == nil {
		t.Error("Report metadata should not be nil")
	}

	if testValue, exists := report.Metadata["test"]; !exists || testValue != true {
		t.Error("Report metadata should preserve original metadata")
	}

	t.Logf("Generated report: ID=%s, BusinessID=%s, Type=%s, Format=%s",
		report.ID, report.BusinessID, report.ReportType, report.Format)
}

func TestRiskService_GenerateRiskReport_Detailed(t *testing.T) {
	// Create test logger
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})

	// Create test components
	categoryRegistry := CreateDefaultRiskCategories()
	thresholdManager := CreateDefaultThresholds()
	industryModelRegistry := CreateDefaultIndustryModels()
	calculator := NewRiskFactorCalculator(categoryRegistry)
	scoringAlgorithm := NewWeightedScoringAlgorithm()
	predictionAlgorithm := NewRiskPredictionAlgorithm()
	historyService := NewRiskHistoryService(logger, nil) // nil DB for testing
	alertService := NewAlertService(logger, thresholdManager)
	reportService := NewReportService(logger, historyService, alertService)

	// Create risk service
	service := NewRiskService(
		logger,
		calculator,
		scoringAlgorithm,
		predictionAlgorithm,
		thresholdManager,
		categoryRegistry,
		industryModelRegistry,
		historyService,
		alertService,
		reportService,
		nil, // No export service for this test
	)

	// Create test request for detailed report
	request := ReportRequest{
		BusinessID: "test_business_456",
		ReportType: ReportTypeDetailed,
		Format:     ReportFormatJSON,
	}

	// Create test context
	ctx := context.WithValue(context.Background(), "request_id", "test_request_456")

	// Generate detailed report
	report, err := service.GenerateRiskReport(ctx, request)
	if err != nil {
		t.Fatalf("Failed to generate detailed risk report: %v", err)
	}

	// Verify detailed report structure
	if report == nil {
		t.Fatal("Detailed report should not be nil")
	}

	if report.ReportType != ReportTypeDetailed {
		t.Errorf("Expected report type %s, got %s", ReportTypeDetailed, report.ReportType)
	}

	// Verify details exist for detailed report
	if report.Details == nil {
		t.Error("Detailed report should have details section")
	}

	// Verify factor breakdown exists
	if len(report.Details.FactorBreakdown) == 0 {
		t.Error("Detailed report should have factor breakdown")
	}

	// Verify category details exist
	if len(report.Details.CategoryDetails) == 0 {
		t.Error("Detailed report should have category details")
	}

	// Verify historical data exists
	if len(report.Details.HistoricalData) == 0 {
		t.Error("Detailed report should have historical data")
	}

	// Verify predictions exist
	if len(report.Details.Predictions) == 0 {
		t.Error("Detailed report should have predictions")
	}

	// Verify risk drivers exist
	if len(report.Details.RiskDrivers) == 0 {
		t.Error("Detailed report should have risk drivers")
	}

	t.Logf("Generated detailed report: ID=%s, BusinessID=%s, Type=%s",
		report.ID, report.BusinessID, report.ReportType)
}

func TestRiskService_GenerateRiskReport_NoReportService(t *testing.T) {
	// Create test logger
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})

	// Create test components
	categoryRegistry := CreateDefaultRiskCategories()
	thresholdManager := CreateDefaultThresholds()
	industryModelRegistry := CreateDefaultIndustryModels()
	calculator := NewRiskFactorCalculator(categoryRegistry)
	scoringAlgorithm := NewWeightedScoringAlgorithm()
	predictionAlgorithm := NewRiskPredictionAlgorithm()
	historyService := NewRiskHistoryService(logger, nil) // nil DB for testing
	alertService := NewAlertService(logger, thresholdManager)

	// Create risk service WITHOUT report service and export service
	service := NewRiskService(
		logger,
		calculator,
		scoringAlgorithm,
		predictionAlgorithm,
		thresholdManager,
		categoryRegistry,
		industryModelRegistry,
		historyService,
		alertService,
		nil, // No report service
		nil, // No export service
	)

	// Create test request
	request := ReportRequest{
		BusinessID: "test_business_789",
		ReportType: ReportTypeSummary,
		Format:     ReportFormatJSON,
	}

	// Create test context
	ctx := context.WithValue(context.Background(), "request_id", "test_request_789")

	// Attempt to generate report - should fail
	report, err := service.GenerateRiskReport(ctx, request)
	if err == nil {
		t.Fatal("Expected error when report service is not available")
	}

	if report != nil {
		t.Error("Report should be nil when report service is not available")
	}

	expectedError := "report service not available"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedError, err.Error())
	}

	t.Logf("Correctly handled missing report service: %v", err)
}
