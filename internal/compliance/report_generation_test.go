package compliance

import (
	"context"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
)

func TestNewReportGenerationService(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	checkEngine := &CheckEngine{}
	tracking := &TrackingSystem{}
	gapAnalyzer := &GapAnalyzer{}
	recommendations := &RecommendationEngine{}

	service := NewReportGenerationService(logger, checkEngine, tracking, gapAnalyzer, recommendations)

	if service == nil {
		t.Fatal("Expected non-nil service")
	}

	if service.logger != logger {
		t.Error("Logger not set correctly")
	}

	if service.checkEngine != checkEngine {
		t.Error("CheckEngine not set correctly")
	}

	if service.tracking != tracking {
		t.Error("Tracking not set correctly")
	}

	if service.gapAnalyzer != gapAnalyzer {
		t.Error("GapAnalyzer not set correctly")
	}

	if service.recommendations != recommendations {
		t.Error("Recommendations not set correctly")
	}
}

func TestGenerateComplianceReport_Validation(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	service := NewReportGenerationService(logger, nil, nil, nil, nil)

	ctx := context.WithValue(context.Background(), "request_id", "test-request")

	// Test empty business ID
	request := ReportRequest{
		BusinessID: "",
		ReportType: ReportTypeStatus,
	}

	_, err := service.GenerateComplianceReport(ctx, request)
	if err == nil {
		t.Error("Expected error for empty business ID")
	}

	// Test valid request with unsupported report type
	request.BusinessID = "test-business"
	request.ReportType = "unsupported_type"

	_, err = service.GenerateComplianceReport(ctx, request)
	if err == nil {
		t.Error("Expected error for unsupported report type")
	}
}

func TestGenerateComplianceReport_DefaultValues(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	service := NewReportGenerationService(logger, nil, nil, nil, nil)

	ctx := context.WithValue(context.Background(), "request_id", "test-request")

	request := ReportRequest{
		BusinessID: "test-business",
		// ReportType and Format are empty, should use defaults
	}

	_, err := service.GenerateComplianceReport(ctx, request)
	if err == nil {
		t.Error("Expected error since dependencies are nil")
	}

	// The defaults are set inside the function, so we need to check the error message
	// or verify that the function handles defaults correctly
	if err != nil && err.Error() != "failed to determine frameworks: tracking system not initialized" {
		t.Errorf("Expected tracking system error, got: %v", err)
	}
}

func TestCalculateOverallStatus(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	service := NewReportGenerationService(logger, nil, nil, nil, nil)

	// Test empty results
	checkResp := &CheckResponse{
		BusinessID: "test-business",
		CheckedAt:  time.Now(),
		Results:    []FrameworkCheckResult{},
		Passed:     0,
		Failed:     0,
	}

	status, score := service.calculateOverallStatus(checkResp)
	if status != ComplianceStatusNotStarted {
		t.Errorf("Expected status %s, got %s", ComplianceStatusNotStarted, status)
	}
	if score != 0.0 {
		t.Errorf("Expected score 0.0, got %f", score)
	}

	// Test with results
	checkResp.Results = []FrameworkCheckResult{
		{
			FrameworkID: "SOC2",
			Summary: ComplianceCheckResult{
				BusinessID: "test-business",
				Framework:  "SOC2",
				Evaluated:  time.Now(),
				Passed:     8,
				Failed:     2,
				Outcomes:   []RuleOutcome{},
			},
		},
	}
	checkResp.Passed = 8
	checkResp.Failed = 2

	status, score = service.calculateOverallStatus(checkResp)
	if status != ComplianceStatusImplemented {
		t.Errorf("Expected status %s, got %s", ComplianceStatusImplemented, status)
	}
	if score != 80.0 {
		t.Errorf("Expected score 80.0, got %f", score)
	}
}

func TestCalculateStatusFromGaps(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	service := NewReportGenerationService(logger, nil, nil, nil, nil)

	// Test no gaps
	requirements := []RequirementReport{}
	controls := []ControlReport{}

	status, score := service.calculateStatusFromGaps(requirements, controls)
	if status != ComplianceStatusVerified {
		t.Errorf("Expected status %s, got %s", ComplianceStatusVerified, status)
	}
	if score != 100.0 {
		t.Errorf("Expected score 100.0, got %f", score)
	}

	// Test with gaps
	requirements = []RequirementReport{
		{
			RequirementID: "REQ1",
			Title:         "Test Requirement",
			Status:        ComplianceStatusNonCompliant,
		},
	}
	controls = []ControlReport{
		{
			ControlID: "CTRL1",
			Title:     "Test Control",
			Status:    ComplianceStatusNonCompliant,
		},
	}

	status, score = service.calculateStatusFromGaps(requirements, controls)
	if status != ComplianceStatusImplemented {
		t.Errorf("Expected status %s, got %s", ComplianceStatusImplemented, status)
	}
	if score != 80.0 {
		t.Errorf("Expected score 80.0, got %f", score)
	}
}

func TestCalculateStatusFromRemediation(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	service := NewReportGenerationService(logger, nil, nil, nil, nil)

	// Test no requirements
	requirements := []RequirementReport{}
	remediationPlans := []RemediationReport{}

	status, score := service.calculateStatusFromRemediation(requirements, remediationPlans)
	if status != ComplianceStatusVerified {
		t.Errorf("Expected status %s, got %s", ComplianceStatusVerified, status)
	}
	if score != 100.0 {
		t.Errorf("Expected score 100.0, got %f", score)
	}

	// Test with requirements
	requirements = []RequirementReport{
		{
			RequirementID: "REQ1",
			Title:         "Test Requirement",
			Status:        ComplianceStatusImplemented,
		},
		{
			RequirementID: "REQ2",
			Title:         "Test Requirement 2",
			Status:        ComplianceStatusNonCompliant,
		},
	}

	status, score = service.calculateStatusFromRemediation(requirements, remediationPlans)
	if status != ComplianceStatusInProgress {
		t.Errorf("Expected status %s, got %s", ComplianceStatusInProgress, status)
	}
	if score != 50.0 {
		t.Errorf("Expected score 50.0, got %f", score)
	}
}

func TestCalculateStatusFromAudit(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	service := NewReportGenerationService(logger, nil, nil, nil, nil)

	// Test no requirements
	requirements := []RequirementReport{}
	controls := []ControlReport{}

	status, score := service.calculateStatusFromAudit(requirements, controls)
	if status != ComplianceStatusVerified {
		t.Errorf("Expected status %s, got %s", ComplianceStatusVerified, status)
	}
	if score != 100.0 {
		t.Errorf("Expected score 100.0, got %f", score)
	}

	// Test with requirements
	requirements = []RequirementReport{
		{
			RequirementID: "REQ1",
			Title:         "Test Requirement",
			Status:        ComplianceStatusVerified,
		},
		{
			RequirementID: "REQ2",
			Title:         "Test Requirement 2",
			Status:        ComplianceStatusNonCompliant,
		},
	}

	status, score = service.calculateStatusFromAudit(requirements, controls)
	if status != ComplianceStatusInProgress {
		t.Errorf("Expected status %s, got %s", ComplianceStatusInProgress, status)
	}
	if score != 50.0 {
		t.Errorf("Expected score 50.0, got %f", score)
	}
}

func TestFilterHighPriorityRecommendations(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	service := NewReportGenerationService(logger, nil, nil, nil, nil)

	recommendations := []ComplianceRecommendation{
		{
			ID:       "REC1",
			Title:    "Critical Recommendation",
			Priority: CompliancePriorityCritical,
		},
		{
			ID:       "REC2",
			Title:    "High Priority Recommendation",
			Priority: CompliancePriorityHigh,
		},
		{
			ID:       "REC3",
			Title:    "Medium Priority Recommendation",
			Priority: CompliancePriorityMedium,
		},
		{
			ID:       "REC4",
			Title:    "Low Priority Recommendation",
			Priority: CompliancePriorityLow,
		},
	}

	filtered := service.filterHighPriorityRecommendations(recommendations)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 high priority recommendations, got %d", len(filtered))
	}

	// Check that only critical and high priority recommendations are included
	for _, rec := range filtered {
		if rec.Priority != CompliancePriorityCritical && rec.Priority != CompliancePriorityHigh {
			t.Errorf("Unexpected priority %s in filtered recommendations", rec.Priority)
		}
	}
}

func TestGetFrameworksToInclude(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	service := NewReportGenerationService(logger, nil, nil, nil, nil)

	ctx := context.WithValue(context.Background(), "request_id", "test-request")

	// Test with specific framework
	frameworks, err := service.getFrameworksToInclude(ctx, "test-business", "SOC2")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(frameworks) != 1 {
		t.Errorf("Expected 1 framework, got %d", len(frameworks))
	}
	if frameworks[0] != "SOC2" {
		t.Errorf("Expected framework SOC2, got %s", frameworks[0])
	}

	// Test with empty framework (should fail since tracking is nil)
	_, err = service.getFrameworksToInclude(ctx, "test-business", "")
	if err == nil {
		t.Error("Expected error when tracking is nil")
	}
}

func TestBuildRequirementsAndControls(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	service := NewReportGenerationService(logger, nil, nil, nil, nil)

	checkResp := &CheckResponse{
		BusinessID: "test-business",
		CheckedAt:  time.Now(),
		Results: []FrameworkCheckResult{
			{
				FrameworkID: "SOC2",
				Summary: ComplianceCheckResult{
					BusinessID: "test-business",
					Framework:  "SOC2",
					Evaluated:  time.Now(),
					Passed:     8,
					Failed:     2,
					Outcomes:   []RuleOutcome{},
				},
			},
		},
		Passed: 8,
		Failed: 2,
	}

	requirements, controls := service.buildRequirementsAndControls(context.Background(), checkResp, false)
	if len(requirements) != 1 {
		t.Errorf("Expected 1 requirement, got %d", len(requirements))
	}
	if len(controls) != 0 {
		t.Errorf("Expected 0 controls (no details), got %d", len(controls))
	}

	// Test with details
	requirements, controls = service.buildRequirementsAndControls(context.Background(), checkResp, true)
	if len(requirements) != 1 {
		t.Errorf("Expected 1 requirement, got %d", len(requirements))
	}
	if len(controls) != 1 {
		t.Errorf("Expected 1 control (with details), got %d", len(controls))
	}
}

func TestConvertGapAnalysisToReport(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	service := NewReportGenerationService(logger, nil, nil, nil, nil)

	gapReport := &GapAnalysisReport{
		BusinessID:  "test-business",
		FrameworkID: "SOC2",
		GeneratedAt: time.Now(),
		RequirementGaps: []RequirementGap{
			{
				RequirementID: "REQ1",
				Title:         "Test Requirement Gap",
				GapType:       GapMissingRequirement,
				Severity:      GapSeverityHigh,
			},
		},
		ControlGaps: []ControlGap{
			{
				ControlID:     "CTRL1",
				RequirementID: "REQ1",
				Title:         "Test Control Gap",
				GapType:       GapMissingControl,
				Severity:      GapSeverityMedium,
			},
		},
	}

	requirements, controls := service.convertGapAnalysisToReport(context.Background(), gapReport)
	if len(requirements) != 1 {
		t.Errorf("Expected 1 requirement, got %d", len(requirements))
	}
	if len(controls) != 1 {
		t.Errorf("Expected 1 control, got %d", len(controls))
	}

	// Check requirement details
	req := requirements[0]
	if req.RequirementID != "REQ1" {
		t.Errorf("Expected requirement ID REQ1, got %s", req.RequirementID)
	}
	if req.Status != ComplianceStatusNonCompliant {
		t.Errorf("Expected status %s, got %s", ComplianceStatusNonCompliant, req.Status)
	}

	// Check control details
	ctrl := controls[0]
	if ctrl.ControlID != "CTRL1" {
		t.Errorf("Expected control ID CTRL1, got %s", ctrl.ControlID)
	}
	if ctrl.Status != ComplianceStatusNonCompliant {
		t.Errorf("Expected status %s, got %s", ComplianceStatusNonCompliant, ctrl.Status)
	}
}
