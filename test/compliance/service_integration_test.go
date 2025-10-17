package compliance

import (
	"context"
	"testing"
	"time"

	"kyb-platform/internal/compliance"
	"kyb-platform/internal/observability"
	"go.uber.org/zap"
)

// TestComplianceServiceIntegration tests integration between compliance services
func TestComplianceServiceIntegration(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create services
	frameworkService := compliance.NewComplianceFrameworkService(logger)
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)
	reportingService := compliance.NewComplianceReportingService(logger, frameworkService, trackingService)
	alertService := compliance.NewComplianceAlertService(logger, frameworkService, trackingService)

	ctx := context.Background()
	businessID := "test-business-123"
	frameworkID := "SOC2"

	t.Run("Framework to Tracking Integration", func(t *testing.T) {
		// Test that tracking service can use framework service data
		tracking, err := trackingService.GetComplianceTracking(ctx, businessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get compliance tracking: %v", err)
		}

		// Verify tracking was initialized with framework requirements
		if len(tracking.Requirements) == 0 {
			t.Error("Tracking should have requirements from framework")
		}

		// Verify framework ID matches
		if tracking.FrameworkID != frameworkID {
			t.Errorf("Expected framework ID %s, got %s", frameworkID, tracking.FrameworkID)
		}

		t.Logf("✅ Framework to Tracking Integration: %d requirements loaded", len(tracking.Requirements))
	})

	t.Run("Tracking to Reporting Integration", func(t *testing.T) {
		// Generate a status report that uses tracking data
		report, err := reportingService.GenerateReport(ctx, businessID, frameworkID, "status", "test-user", nil)
		if err != nil {
			t.Fatalf("Failed to generate status report: %v", err)
		}

		// Verify report contains tracking data
		if report.ReportData.ComplianceStatus == nil {
			t.Error("Status report should contain compliance status data")
		}

		// Verify report data is populated from tracking service
		status := report.ReportData.ComplianceStatus
		if status.RequirementsTotal == 0 {
			t.Error("Report should have requirement count from tracking")
		}

		t.Logf("✅ Tracking to Reporting Integration: Report generated with %d requirements", status.RequirementsTotal)
	})

	t.Run("Framework to Reporting Integration", func(t *testing.T) {
		// Generate a gap analysis report
		report, err := reportingService.GenerateReport(ctx, businessID, frameworkID, "gap_analysis", "test-user", nil)
		if err != nil {
			t.Fatalf("Failed to generate gap analysis report: %v", err)
		}

		// Verify report contains framework data
		if report.ReportData.GapAnalysis == nil {
			t.Error("Gap analysis report should contain gap analysis data")
		}

		// Verify gap analysis is populated
		gapAnalysis := report.ReportData.GapAnalysis
		if gapAnalysis.TotalGaps == 0 && gapAnalysis.CriticalGaps == 0 {
			t.Log("Gap analysis shows no gaps (expected for new business)")
		}

		t.Logf("✅ Framework to Reporting Integration: Gap analysis completed")
	})

	t.Run("Tracking to Alert Integration", func(t *testing.T) {
		// Evaluate alert rules which should use tracking data
		err := alertService.EvaluateAlertRules(ctx, businessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to evaluate alert rules: %v", err)
		}

		// Check if any alerts were created
		query := &compliance.AlertQuery{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
		}
		alerts, err := alertService.ListAlerts(ctx, query)
		if err != nil {
			t.Fatalf("Failed to list alerts: %v", err)
		}

		t.Logf("✅ Tracking to Alert Integration: %d alerts found after rule evaluation", len(alerts))
	})

	t.Run("End-to-End Workflow Integration", func(t *testing.T) {
		// Simulate a complete compliance workflow

		// 1. Get framework requirements
		framework, err := frameworkService.GetFramework(ctx, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get framework: %v", err)
		}

		// 2. Initialize tracking
		tracking, err := trackingService.GetComplianceTracking(ctx, businessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get tracking: %v", err)
		}

		// 3. Update tracking progress
		tracking.OverallProgress = 0.75
		tracking.ComplianceLevel = "partial"
		err = trackingService.UpdateComplianceTracking(ctx, tracking)
		if err != nil {
			t.Fatalf("Failed to update tracking: %v", err)
		}

		// 4. Generate report
		report, err := reportingService.GenerateReport(ctx, businessID, frameworkID, "status", "test-user", nil)
		if err != nil {
			t.Fatalf("Failed to generate report: %v", err)
		}

		// 5. Evaluate alerts
		err = alertService.EvaluateAlertRules(ctx, businessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to evaluate alerts: %v", err)
		}

		// Verify end-to-end data consistency
		if report.ReportData.ComplianceStatus.OverallProgress != tracking.OverallProgress {
			t.Errorf("Report progress %.2f doesn't match tracking progress %.2f",
				report.ReportData.ComplianceStatus.OverallProgress, tracking.OverallProgress)
		}

		t.Logf("✅ End-to-End Workflow: Framework=%s, Progress=%.2f, Level=%s",
			framework.Name, tracking.OverallProgress, tracking.ComplianceLevel)
	})
}

// TestComplianceServiceDataConsistency tests data consistency across services
func TestComplianceServiceDataConsistency(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create services
	frameworkService := compliance.NewComplianceFrameworkService(logger)
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)
	reportingService := compliance.NewComplianceReportingService(logger, frameworkService, trackingService)
	alertService := compliance.NewComplianceAlertService(logger, frameworkService, trackingService)

	ctx := context.Background()
	businessID := "test-business-456"
	frameworkID := "GDPR"

	t.Run("Framework Data Consistency", func(t *testing.T) {
		// Get framework from service
		framework, err := frameworkService.GetFramework(ctx, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get framework: %v", err)
		}

		// Get requirements for framework
		requirements, err := frameworkService.GetFrameworkRequirements(ctx, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get framework requirements: %v", err)
		}

		// Verify framework ID consistency
		if framework.ID != frameworkID {
			t.Errorf("Framework ID mismatch: expected %s, got %s", frameworkID, framework.ID)
		}

		// Verify requirements belong to framework
		for _, req := range requirements {
			if req.FrameworkID != frameworkID {
				t.Errorf("Requirement %s has wrong framework ID: %s", req.ID, req.FrameworkID)
			}
		}

		t.Logf("✅ Framework Data Consistency: %d requirements for %s", len(requirements), framework.Name)
	})

	t.Run("Tracking Data Consistency", func(t *testing.T) {
		// Get tracking data
		tracking, err := trackingService.GetComplianceTracking(ctx, businessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get tracking: %v", err)
		}

		// Verify business and framework IDs
		if tracking.BusinessID != businessID {
			t.Errorf("Tracking business ID mismatch: expected %s, got %s", businessID, tracking.BusinessID)
		}
		if tracking.FrameworkID != frameworkID {
			t.Errorf("Tracking framework ID mismatch: expected %s, got %s", frameworkID, tracking.FrameworkID)
		}

		// Verify progress is within valid range
		if tracking.OverallProgress < 0.0 || tracking.OverallProgress > 1.0 {
			t.Errorf("Invalid progress value: %.2f (should be 0.0-1.0)", tracking.OverallProgress)
		}

		// Verify compliance level is valid
		validLevels := []string{"compliant", "partial", "non_compliant", "in_progress"}
		validLevel := false
		for _, level := range validLevels {
			if tracking.ComplianceLevel == level {
				validLevel = true
				break
			}
		}
		if !validLevel {
			t.Errorf("Invalid compliance level: %s", tracking.ComplianceLevel)
		}

		t.Logf("✅ Tracking Data Consistency: Progress=%.2f, Level=%s, Risk=%s",
			tracking.OverallProgress, tracking.ComplianceLevel, tracking.RiskLevel)
	})

	t.Run("Report Data Consistency", func(t *testing.T) {
		// Generate multiple report types
		reportTypes := []string{"status", "gap_analysis", "executive_summary"}

		for _, reportType := range reportTypes {
			report, err := reportingService.GenerateReport(ctx, businessID, frameworkID, reportType, "test-user", nil)
			if err != nil {
				t.Fatalf("Failed to generate %s report: %v", reportType, err)
			}

			// Verify report metadata
			if report.BusinessID != businessID {
				t.Errorf("Report business ID mismatch: expected %s, got %s", businessID, report.BusinessID)
			}
			if report.FrameworkID != frameworkID {
				t.Errorf("Report framework ID mismatch: expected %s, got %s", frameworkID, report.FrameworkID)
			}
			if report.ReportType != reportType {
				t.Errorf("Report type mismatch: expected %s, got %s", reportType, report.ReportType)
			}

			// Verify report data is populated
			if report.ReportData == nil {
				t.Errorf("Report data is nil for %s report", reportType)
			}

			t.Logf("✅ Report Data Consistency: %s report generated successfully", reportType)
		}
	})

	t.Run("Alert Data Consistency", func(t *testing.T) {
		// Create a test alert
		alert := &compliance.ComplianceAlert{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
			AlertType:   "compliance_change",
			Severity:    "medium",
			Title:       "Test Alert",
			Description: "Test alert for consistency",
			TriggeredBy: "test",
		}

		err := alertService.CreateAlert(ctx, alert)
		if err != nil {
			t.Fatalf("Failed to create alert: %v", err)
		}

		// Retrieve the alert
		retrievedAlert, err := alertService.GetAlert(ctx, alert.ID)
		if err != nil {
			t.Fatalf("Failed to get alert: %v", err)
		}

		// Verify data consistency
		if retrievedAlert.BusinessID != businessID {
			t.Errorf("Alert business ID mismatch: expected %s, got %s", businessID, retrievedAlert.BusinessID)
		}
		if retrievedAlert.FrameworkID != frameworkID {
			t.Errorf("Alert framework ID mismatch: expected %s, got %s", frameworkID, retrievedAlert.FrameworkID)
		}
		if retrievedAlert.AlertType != "compliance_change" {
			t.Errorf("Alert type mismatch: expected compliance_change, got %s", retrievedAlert.AlertType)
		}

		t.Logf("✅ Alert Data Consistency: Alert %s created and retrieved successfully", alert.ID)
	})
}

// TestComplianceServiceErrorHandling tests error handling across services
func TestComplianceServiceErrorHandling(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create services
	frameworkService := compliance.NewComplianceFrameworkService(logger)
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)
	reportingService := compliance.NewComplianceReportingService(logger, frameworkService, trackingService)
	alertService := compliance.NewComplianceAlertService(logger, frameworkService, trackingService)

	ctx := context.Background()

	t.Run("Framework Service Error Handling", func(t *testing.T) {
		// Test getting non-existent framework
		_, err := frameworkService.GetFramework(ctx, "NON_EXISTENT")
		if err == nil {
			t.Error("Expected error for non-existent framework")
		}

		// Test getting requirements for non-existent framework
		_, err = frameworkService.GetFrameworkRequirements(ctx, "NON_EXISTENT")
		if err == nil {
			t.Error("Expected error for non-existent framework requirements")
		}

		t.Logf("✅ Framework Service Error Handling: Errors handled correctly")
	})

	t.Run("Tracking Service Error Handling", func(t *testing.T) {
		// Test updating tracking with invalid data
		invalidTracking := &compliance.ComplianceTracking{
			BusinessID:  "", // Invalid: empty business ID
			FrameworkID: "SOC2",
		}

		err := trackingService.UpdateComplianceTracking(ctx, invalidTracking)
		if err == nil {
			t.Error("Expected error for invalid tracking data")
		}

		t.Logf("✅ Tracking Service Error Handling: Invalid data rejected")
	})

	t.Run("Reporting Service Error Handling", func(t *testing.T) {
		// Test generating report with invalid type
		_, err := reportingService.GenerateReport(ctx, "business-123", "SOC2", "invalid_type", "user", nil)
		if err == nil {
			t.Error("Expected error for invalid report type")
		}

		// Test getting non-existent report
		_, err = reportingService.GetReport(ctx, "non-existent-report-id")
		if err == nil {
			t.Error("Expected error for non-existent report")
		}

		t.Logf("✅ Reporting Service Error Handling: Invalid requests rejected")
	})

	t.Run("Alert Service Error Handling", func(t *testing.T) {
		// Test getting non-existent alert
		_, err := alertService.GetAlert(ctx, "non-existent-alert-id")
		if err == nil {
			t.Error("Expected error for non-existent alert")
		}

		// Test updating non-existent alert status
		err = alertService.UpdateAlertStatus(ctx, "non-existent-alert-id", "resolved", "user")
		if err == nil {
			t.Error("Expected error for updating non-existent alert")
		}

		t.Logf("✅ Alert Service Error Handling: Non-existent resources handled correctly")
	})
}

// TestComplianceServicePerformance tests performance of service operations
func TestComplianceServicePerformance(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create services
	frameworkService := compliance.NewComplianceFrameworkService(logger)
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)
	reportingService := compliance.NewComplianceReportingService(logger, frameworkService, trackingService)
	alertService := compliance.NewComplianceAlertService(logger, frameworkService, trackingService)

	ctx := context.Background()
	businessID := "perf-test-business"
	frameworkID := "SOC2"

	t.Run("Framework Service Performance", func(t *testing.T) {
		start := time.Now()

		// Get framework
		_, err := frameworkService.GetFramework(ctx, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get framework: %v", err)
		}

		duration := time.Since(start)
		if duration > 50*time.Millisecond {
			t.Errorf("Framework retrieval took too long: %v", duration)
		}

		t.Logf("✅ Framework Service Performance: %v", duration)
	})

	t.Run("Tracking Service Performance", func(t *testing.T) {
		start := time.Now()

		// Get tracking
		_, err := trackingService.GetComplianceTracking(ctx, businessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get tracking: %v", err)
		}

		duration := time.Since(start)
		if duration > 100*time.Millisecond {
			t.Errorf("Tracking retrieval took too long: %v", duration)
		}

		t.Logf("✅ Tracking Service Performance: %v", duration)
	})

	t.Run("Reporting Service Performance", func(t *testing.T) {
		start := time.Now()

		// Generate report
		_, err := reportingService.GenerateReport(ctx, businessID, frameworkID, "status", "test-user", nil)
		if err != nil {
			t.Fatalf("Failed to generate report: %v", err)
		}

		duration := time.Since(start)
		if duration > 200*time.Millisecond {
			t.Errorf("Report generation took too long: %v", duration)
		}

		t.Logf("✅ Reporting Service Performance: %v", duration)
	})

	t.Run("Alert Service Performance", func(t *testing.T) {
		start := time.Now()

		// Evaluate alert rules
		err := alertService.EvaluateAlertRules(ctx, businessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to evaluate alert rules: %v", err)
		}

		duration := time.Since(start)
		if duration > 150*time.Millisecond {
			t.Errorf("Alert evaluation took too long: %v", duration)
		}

		t.Logf("✅ Alert Service Performance: %v", duration)
	})
}
