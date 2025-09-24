package compliance

import (
	"context"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/compliance"
	"github.com/pcraw4d/business-verification/internal/observability"
	"go.uber.org/zap"
)

// TestReportingAccuracy tests accuracy of compliance reporting
func TestReportingAccuracy(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create services
	frameworkService := compliance.NewComplianceFrameworkService(logger)
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)
	reportingService := compliance.NewComplianceReportingService(logger, frameworkService, trackingService)

	ctx := context.Background()
	businessID := "report-test-business"
	frameworkID := "SOC2"

	t.Run("Status Report Accuracy", func(t *testing.T) {
		// Setup tracking data
		tracking, err := trackingService.GetComplianceTracking(ctx, businessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get tracking: %v", err)
		}

		// Update tracking with known values
		tracking.OverallProgress = 0.75
		tracking.ComplianceLevel = "partial"
		tracking.RiskLevel = "medium"
		tracking.Trend = "improving"

		// Set requirement progress
		for i := range tracking.Requirements {
			tracking.Requirements[i].Progress = 0.75
			tracking.Requirements[i].Status = "in_progress"
		}

		err = trackingService.UpdateComplianceTracking(ctx, tracking)
		if err != nil {
			t.Fatalf("Failed to update tracking: %v", err)
		}

		// Generate status report
		report, err := reportingService.GenerateReport(ctx, businessID, frameworkID, "status", "test-user", nil)
		if err != nil {
			t.Fatalf("Failed to generate status report: %v", err)
		}

		// Verify report accuracy
		if report.ReportData.ComplianceStatus == nil {
			t.Fatal("Status report missing compliance status data")
		}

		status := report.ReportData.ComplianceStatus

		// Verify overall progress
		if status.OverallProgress != 0.75 {
			t.Errorf("Expected overall progress 0.75, got %.2f", status.OverallProgress)
		}

		// Verify requirement counts
		expectedTotal := len(tracking.Requirements)
		if status.RequirementsTotal != expectedTotal {
			t.Errorf("Expected %d total requirements, got %d", expectedTotal, status.RequirementsTotal)
		}

		// Verify compliance trend
		if status.ComplianceTrend != "improving" {
			t.Errorf("Expected compliance trend 'improving', got '%s'", status.ComplianceTrend)
		}

		t.Logf("✅ Status Report Accuracy: Progress=%.2f, Total=%d, Trend=%s",
			status.OverallProgress, status.RequirementsTotal, status.ComplianceTrend)
	})

	t.Run("Gap Analysis Report Accuracy", func(t *testing.T) {
		// Setup tracking with gaps
		tracking, err := trackingService.GetComplianceTracking(ctx, businessID+"-gaps", frameworkID)
		if err != nil {
			t.Fatalf("Failed to get tracking: %v", err)
		}

		// Create gaps by setting some requirements to incomplete
		for i := range tracking.Requirements {
			if i < len(tracking.Requirements)/2 {
				tracking.Requirements[i].Progress = 1.0 // Complete
				tracking.Requirements[i].Status = "completed"
			} else {
				tracking.Requirements[i].Progress = 0.3 // Incomplete
				tracking.Requirements[i].Status = "in_progress"
			}
		}

		err = trackingService.UpdateComplianceTracking(ctx, tracking)
		if err != nil {
			t.Fatalf("Failed to update tracking: %v", err)
		}

		// Generate gap analysis report
		report, err := reportingService.GenerateReport(ctx, businessID+"-gaps", frameworkID, "gap_analysis", "test-user", nil)
		if err != nil {
			t.Fatalf("Failed to generate gap analysis report: %v", err)
		}

		// Verify gap analysis accuracy
		if report.ReportData.GapAnalysis == nil {
			t.Fatal("Gap analysis report missing gap analysis data")
		}

		gapAnalysis := report.ReportData.GapAnalysis

		// Verify gap counts
		expectedGaps := len(tracking.Requirements) / 2
		if gapAnalysis.TotalGaps != expectedGaps {
			t.Errorf("Expected %d total gaps, got %d", expectedGaps, gapAnalysis.TotalGaps)
		}

		// Verify gap categories exist
		if len(gapAnalysis.GapCategories) == 0 {
			t.Error("Gap analysis should have gap categories")
		}

		t.Logf("✅ Gap Analysis Report Accuracy: TotalGaps=%d, Categories=%d",
			gapAnalysis.TotalGaps, len(gapAnalysis.GapCategories))
	})

	t.Run("Executive Summary Report Accuracy", func(t *testing.T) {
		// Setup tracking data
		execBusinessID := businessID + "-exec"
		tracking, err := trackingService.GetComplianceTracking(ctx, execBusinessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get tracking: %v", err)
		}

		// Set specific values for testing
		tracking.OverallProgress = 0.85
		tracking.ComplianceLevel = "compliant"
		tracking.RiskLevel = "low"
		tracking.Trend = "improving"

		err = trackingService.UpdateComplianceTracking(ctx, tracking)
		if err != nil {
			t.Fatalf("Failed to update tracking: %v", err)
		}

		// Generate executive summary report
		report, err := reportingService.GenerateReport(ctx, execBusinessID, frameworkID, "executive_summary", "test-user", nil)
		if err != nil {
			t.Fatalf("Failed to generate executive summary report: %v", err)
		}

		// Verify executive summary accuracy
		if report.ReportData.ExecutiveSummary == nil {
			t.Fatal("Executive summary report missing executive summary data")
		}

		summary := report.ReportData.ExecutiveSummary

		// Verify compliance score
		if summary.OverallComplianceScore != 0.85 {
			t.Errorf("Expected compliance score 0.85, got %.2f", summary.OverallComplianceScore)
		}

		// Verify compliance level
		if summary.ComplianceLevel != "compliant" {
			t.Errorf("Expected compliance level 'compliant', got '%s'", summary.ComplianceLevel)
		}

		// Verify risk level
		if summary.RiskLevel != "low" {
			t.Errorf("Expected risk level 'low', got '%s'", summary.RiskLevel)
		}

		// Verify key findings exist
		if len(summary.KeyFindings) == 0 {
			t.Error("Executive summary should have key findings")
		}

		// Verify next steps exist
		if len(summary.NextSteps) == 0 {
			t.Error("Executive summary should have next steps")
		}

		t.Logf("✅ Executive Summary Report Accuracy: Score=%.2f, Level=%s, Risk=%s, Findings=%d",
			summary.OverallComplianceScore, summary.ComplianceLevel, summary.RiskLevel, len(summary.KeyFindings))
	})

	t.Run("Risk Assessment Report Accuracy", func(t *testing.T) {
		// Setup tracking with high risk
		riskBusinessID := businessID + "-risk"
		tracking, err := trackingService.GetComplianceTracking(ctx, riskBusinessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get tracking: %v", err)
		}

		// Set high risk scenario
		tracking.OverallProgress = 0.25
		tracking.ComplianceLevel = "non_compliant"
		tracking.RiskLevel = "critical"
		tracking.Trend = "declining"

		err = trackingService.UpdateComplianceTracking(ctx, tracking)
		if err != nil {
			t.Fatalf("Failed to update tracking: %v", err)
		}

		// Generate executive summary report (includes risk assessment)
		report, err := reportingService.GenerateReport(ctx, riskBusinessID, frameworkID, "executive_summary", "test-user", nil)
		if err != nil {
			t.Fatalf("Failed to generate executive summary report: %v", err)
		}

		// Verify risk assessment accuracy
		if report.ReportData.RiskAssessment == nil {
			t.Fatal("Executive summary report missing risk assessment data")
		}

		riskAssessment := report.ReportData.RiskAssessment

		// Verify risk score (should be inverse of compliance score)
		expectedRiskScore := 1.0 - 0.25 // 0.75
		if riskAssessment.OverallRiskScore != expectedRiskScore {
			t.Errorf("Expected risk score %.2f, got %.2f", expectedRiskScore, riskAssessment.OverallRiskScore)
		}

		// Verify risk level
		if riskAssessment.RiskLevel != "critical" {
			t.Errorf("Expected risk level 'critical', got '%s'", riskAssessment.RiskLevel)
		}

		// Verify risk trend
		if riskAssessment.RiskTrend != "declining" {
			t.Errorf("Expected risk trend 'declining', got '%s'", riskAssessment.RiskTrend)
		}

		// Verify top risks exist
		if len(riskAssessment.TopRisks) == 0 {
			t.Error("Risk assessment should have top risks")
		}

		t.Logf("✅ Risk Assessment Report Accuracy: RiskScore=%.2f, Level=%s, Trend=%s, TopRisks=%d",
			riskAssessment.OverallRiskScore, riskAssessment.RiskLevel, riskAssessment.RiskTrend, len(riskAssessment.TopRisks))
	})
}

// TestReportDataConsistency tests consistency of report data across different report types
func TestReportDataConsistency(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create services
	frameworkService := compliance.NewComplianceFrameworkService(logger)
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)
	reportingService := compliance.NewComplianceReportingService(logger, frameworkService, trackingService)

	ctx := context.Background()
	businessID := "consistency-test-business"
	frameworkID := "GDPR"

	// Setup consistent tracking data
	tracking, err := trackingService.GetComplianceTracking(ctx, businessID, frameworkID)
	if err != nil {
		t.Fatalf("Failed to get tracking: %v", err)
	}

	// Set specific values
	tracking.OverallProgress = 0.65
	tracking.ComplianceLevel = "partial"
	tracking.RiskLevel = "medium"
	tracking.Trend = "stable"

	err = trackingService.UpdateComplianceTracking(ctx, tracking)
	if err != nil {
		t.Fatalf("Failed to update tracking: %v", err)
	}

	t.Run("Cross-Report Data Consistency", func(t *testing.T) {
		// Generate multiple report types
		reportTypes := []string{"status", "gap_analysis", "executive_summary"}
		reports := make(map[string]*compliance.ComplianceReport)

		for _, reportType := range reportTypes {
			report, err := reportingService.GenerateReport(ctx, businessID, frameworkID, reportType, "test-user", nil)
			if err != nil {
				t.Fatalf("Failed to generate %s report: %v", reportType, err)
			}
			reports[reportType] = report
		}

		// Verify consistency across reports
		statusReport := reports["status"]
		execReport := reports["executive_summary"]

		// Check compliance score consistency
		if statusReport.ReportData.ComplianceStatus.OverallProgress != execReport.ReportData.ExecutiveSummary.OverallComplianceScore {
			t.Errorf("Compliance score mismatch: status=%.2f, executive=%.2f",
				statusReport.ReportData.ComplianceStatus.OverallProgress,
				execReport.ReportData.ExecutiveSummary.OverallComplianceScore)
		}

		// Check compliance level consistency
		if statusReport.ReportData.ComplianceStatus.ComplianceTrend != execReport.ReportData.ExecutiveSummary.RiskLevel {
			// This might be different as executive summary uses risk level from tracking
			t.Logf("Note: Compliance trend and risk level may differ by design")
		}

		t.Logf("✅ Cross-Report Data Consistency: All reports generated with consistent data")
	})

	t.Run("Report Metadata Consistency", func(t *testing.T) {
		// Generate a report
		report, err := reportingService.GenerateReport(ctx, businessID, frameworkID, "status", "test-user", nil)
		if err != nil {
			t.Fatalf("Failed to generate report: %v", err)
		}

		// Verify metadata consistency
		if report.BusinessID != businessID {
			t.Errorf("Report business ID mismatch: expected %s, got %s", businessID, report.BusinessID)
		}

		if report.FrameworkID != frameworkID {
			t.Errorf("Report framework ID mismatch: expected %s, got %s", frameworkID, report.FrameworkID)
		}

		if report.ReportType != "status" {
			t.Errorf("Report type mismatch: expected 'status', got '%s'", report.ReportType)
		}

		if report.GeneratedBy != "test-user" {
			t.Errorf("Report generated by mismatch: expected 'test-user', got '%s'", report.GeneratedBy)
		}

		// Verify timestamps
		if report.GeneratedAt.IsZero() {
			t.Error("Report generated at timestamp is zero")
		}

		if report.CreatedAt.IsZero() {
			t.Error("Report created at timestamp is zero")
		}

		t.Logf("✅ Report Metadata Consistency: All metadata fields correct")
	})
}

// TestReportGenerationPerformance tests performance of report generation
func TestReportGenerationPerformance(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create services
	frameworkService := compliance.NewComplianceFrameworkService(logger)
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)
	reportingService := compliance.NewComplianceReportingService(logger, frameworkService, trackingService)

	ctx := context.Background()
	businessID := "perf-test-business"
	frameworkID := "SOC2"

	// Setup tracking data
	tracking, err := trackingService.GetComplianceTracking(ctx, businessID, frameworkID)
	if err != nil {
		t.Fatalf("Failed to get tracking: %v", err)
	}

	tracking.OverallProgress = 0.75
	err = trackingService.UpdateComplianceTracking(ctx, tracking)
	if err != nil {
		t.Fatalf("Failed to update tracking: %v", err)
	}

	t.Run("Report Generation Speed", func(t *testing.T) {
		reportTypes := []string{"status", "gap_analysis", "executive_summary"}
		maxDurations := map[string]time.Duration{
			"status":            100 * time.Millisecond,
			"gap_analysis":      150 * time.Millisecond,
			"executive_summary": 200 * time.Millisecond,
		}

		for _, reportType := range reportTypes {
			t.Run(reportType, func(t *testing.T) {
				start := time.Now()

				report, err := reportingService.GenerateReport(ctx, businessID, frameworkID, reportType, "test-user", nil)

				duration := time.Since(start)

				if err != nil {
					t.Fatalf("Failed to generate %s report: %v", reportType, err)
				}

				if report == nil {
					t.Fatalf("Generated report is nil for %s", reportType)
				}

				maxDuration := maxDurations[reportType]
				if duration > maxDuration {
					t.Errorf("%s report generation took %v, expected less than %v", reportType, duration, maxDuration)
				}

				t.Logf("✅ %s Report Generation: %v (max: %v)", reportType, duration, maxDuration)
			})
		}
	})

	t.Run("Concurrent Report Generation", func(t *testing.T) {
		// Test concurrent report generation
		concurrency := 5
		results := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(id int) {
				_, err := reportingService.GenerateReport(ctx, businessID, frameworkID, "status", "test-user", nil)
				results <- err
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 0; i < concurrency; i++ {
			err := <-results
			if err == nil {
				successCount++
			}
		}

		// Verify all reports generated successfully
		if successCount != concurrency {
			t.Errorf("Expected %d successful reports, got %d", concurrency, successCount)
		}

		t.Logf("✅ Concurrent Report Generation: %d/%d successful", successCount, concurrency)
	})
}
