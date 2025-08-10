package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ReportGenerationService provides comprehensive compliance report generation functionality
type ReportGenerationService struct {
	logger          *observability.Logger
	checkEngine     *CheckEngine
	tracking        *TrackingSystem
	gapAnalyzer     *GapAnalyzer
	recommendations *RecommendationEngine
}

// NewReportGenerationService creates a new compliance report generation service
func NewReportGenerationService(logger *observability.Logger, checkEngine *CheckEngine, tracking *TrackingSystem, gapAnalyzer *GapAnalyzer, recommendations *RecommendationEngine) *ReportGenerationService {
	return &ReportGenerationService{
		logger:          logger,
		checkEngine:     checkEngine,
		tracking:        tracking,
		gapAnalyzer:     gapAnalyzer,
		recommendations: recommendations,
	}
}

// ReportRequest represents a request to generate a compliance report
type ReportRequest struct {
	BusinessID     string                 `json:"business_id"`
	Framework      string                 `json:"framework,omitempty"` // if empty, generate for all frameworks
	ReportType     ReportType             `json:"report_type"`
	Format         ReportFormat           `json:"format,omitempty"`
	DateRange      *DateRange             `json:"date_range,omitempty"`
	IncludeDetails bool                   `json:"include_details,omitempty"`
	GeneratedBy    string                 `json:"generated_by"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// DateRange represents a date range for report generation
type DateRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// ReportFormat represents the format of the report
type ReportFormat string

const (
	ReportFormatJSON ReportFormat = "json"
	ReportFormatPDF  ReportFormat = "pdf"
	ReportFormatHTML ReportFormat = "html"
	ReportFormatCSV  ReportFormat = "csv"
)

// GenerateComplianceReport generates a comprehensive compliance report
func (s *ReportGenerationService) GenerateComplianceReport(ctx context.Context, request ReportRequest) (*ComplianceReport, error) {
	requestID := ""
	if ctx.Value("request_id") != nil {
		requestID = ctx.Value("request_id").(string)
	}

	s.logger.Info("Generating compliance report",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"framework", request.Framework,
		"report_type", request.ReportType,
		"format", request.Format,
	)

	// Validate request
	if request.BusinessID == "" {
		return nil, fmt.Errorf("business_id is required")
	}

	if request.ReportType == "" {
		request.ReportType = ReportTypeStatus // Default to status report
	}

	if request.Format == "" {
		request.Format = ReportFormatJSON // Default to JSON
	}

	// Determine frameworks to include
	frameworks, err := s.getFrameworksToInclude(ctx, request.BusinessID, request.Framework)
	if err != nil {
		return nil, fmt.Errorf("failed to determine frameworks: %w", err)
	}

	// Generate report based on type
	var report *ComplianceReport
	switch request.ReportType {
	case ReportTypeStatus:
		report, err = s.generateStatusReport(ctx, request, frameworks)
	case ReportTypeGap:
		report, err = s.generateGapReport(ctx, request, frameworks)
	case ReportTypeRemediation:
		report, err = s.generateRemediationReport(ctx, request, frameworks)
	case ReportTypeAudit:
		report, err = s.generateAuditReport(ctx, request, frameworks)
	case ReportTypeExecutive:
		report, err = s.generateExecutiveReport(ctx, request, frameworks)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", request.ReportType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate %s report: %w", request.ReportType, err)
	}

	// Set common report fields
	report.GeneratedBy = request.GeneratedBy
	report.Metadata = request.Metadata

	s.logger.Info("Compliance report generated successfully",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"report_type", request.ReportType,
		"framework_count", len(frameworks),
		"requirement_count", len(report.Requirements),
		"control_count", len(report.Controls),
	)

	return report, nil
}

// generateStatusReport generates a status report showing current compliance status
func (s *ReportGenerationService) generateStatusReport(ctx context.Context, request ReportRequest, frameworks []string) (*ComplianceReport, error) {
	report := &ComplianceReport{
		ID:          fmt.Sprintf("status_report_%s_%d", request.BusinessID, time.Now().Unix()),
		BusinessID:  request.BusinessID,
		ReportType:  ReportTypeStatus,
		Title:       "Compliance Status Report",
		Description: "Current compliance status across all frameworks",
		GeneratedAt: time.Now(),
		Period:      "current",
	}

	// Get compliance check results
	checkReq := CheckRequest{
		BusinessID: request.BusinessID,
		Frameworks: frameworks,
		Options:    EvaluationOptions{},
	}

	checkResp, err := s.checkEngine.Check(ctx, checkReq)
	if err != nil {
		return nil, fmt.Errorf("failed to run compliance check: %w", err)
	}

	// Calculate overall status and score
	report.OverallStatus, report.ComplianceScore = s.calculateOverallStatus(checkResp)

	// Build requirements and controls from check results
	report.Requirements, report.Controls = s.buildRequirementsAndControls(ctx, checkResp, request.IncludeDetails)

	// Get recommendations
	if s.recommendations != nil {
		recommendations, err := s.recommendations.GenerateRecommendations(ctx, request.BusinessID, frameworks)
		if err != nil {
			s.logger.Warn("Failed to generate recommendations", "error", err.Error())
		} else {
			report.Recommendations = recommendations
		}
	}

	return report, nil
}

// generateGapReport generates a gap analysis report
func (s *ReportGenerationService) generateGapReport(ctx context.Context, request ReportRequest, frameworks []string) (*ComplianceReport, error) {
	report := &ComplianceReport{
		ID:          fmt.Sprintf("gap_report_%s_%d", request.BusinessID, time.Now().Unix()),
		BusinessID:  request.BusinessID,
		ReportType:  ReportTypeGap,
		Title:       "Compliance Gap Analysis Report",
		Description: "Detailed analysis of compliance gaps and remediation needs",
		GeneratedAt: time.Now(),
		Period:      "current",
	}

	// Run gap analysis for each framework
	for _, framework := range frameworks {
		gapReport, err := s.gapAnalyzer.AnalyzeGaps(ctx, request.BusinessID, framework)
		if err != nil {
			s.logger.Warn("Failed to analyze gaps for framework", "framework", framework, "error", err.Error())
			continue
		}

		// Convert gap analysis to report format
		requirements, controls := s.convertGapAnalysisToReport(ctx, gapReport)
		report.Requirements = append(report.Requirements, requirements...)
		report.Controls = append(report.Controls, controls...)
	}

	// Calculate overall status based on gaps
	report.OverallStatus, report.ComplianceScore = s.calculateStatusFromGaps(report.Requirements, report.Controls)

	// Get remediation recommendations
	if s.recommendations != nil {
		recommendations, err := s.recommendations.GenerateRecommendations(ctx, request.BusinessID, frameworks)
		if err != nil {
			s.logger.Warn("Failed to generate recommendations", "error", err.Error())
		} else {
			report.Recommendations = recommendations
		}
	}

	return report, nil
}

// generateRemediationReport generates a remediation-focused report
func (s *ReportGenerationService) generateRemediationReport(ctx context.Context, request ReportRequest, frameworks []string) (*ComplianceReport, error) {
	report := &ComplianceReport{
		ID:          fmt.Sprintf("remediation_report_%s_%d", request.BusinessID, time.Now().Unix()),
		BusinessID:  request.BusinessID,
		ReportType:  ReportTypeRemediation,
		Title:       "Compliance Remediation Report",
		Description: "Comprehensive remediation plans and progress tracking",
		GeneratedAt: time.Now(),
		Period:      "current",
	}

	// Get tracking data for remediation information
	for _, framework := range frameworks {
		tracking, err := s.tracking.GetComplianceTracking(ctx, request.BusinessID, framework)
		if err != nil {
			s.logger.Warn("Failed to get tracking for framework", "framework", framework, "error", err.Error())
			continue
		}

		// Extract remediation plans and exceptions
		requirements, remediationPlans, exceptions := s.extractRemediationData(ctx, tracking)
		report.Requirements = append(report.Requirements, requirements...)
		report.RemediationPlans = append(report.RemediationPlans, remediationPlans...)
		report.Exceptions = append(report.Exceptions, exceptions...)
	}

	// Calculate overall status
	report.OverallStatus, report.ComplianceScore = s.calculateStatusFromRemediation(report.Requirements, report.RemediationPlans)

	return report, nil
}

// generateAuditReport generates an audit-focused report
func (s *ReportGenerationService) generateAuditReport(ctx context.Context, request ReportRequest, frameworks []string) (*ComplianceReport, error) {
	report := &ComplianceReport{
		ID:          fmt.Sprintf("audit_report_%s_%d", request.BusinessID, time.Now().Unix()),
		BusinessID:  request.BusinessID,
		ReportType:  ReportTypeAudit,
		Title:       "Compliance Audit Report",
		Description: "Audit-focused compliance report with evidence and testing results",
		GeneratedAt: time.Now(),
		Period:      "current",
	}

	// Get tracking data for audit information
	for _, framework := range frameworks {
		tracking, err := s.tracking.GetComplianceTracking(ctx, request.BusinessID, framework)
		if err != nil {
			s.logger.Warn("Failed to get tracking for framework", "framework", framework, "error", err.Error())
			continue
		}

		// Extract audit-related data
		requirements, controls := s.extractAuditData(ctx, tracking)
		report.Requirements = append(report.Requirements, requirements...)
		report.Controls = append(report.Controls, controls...)
	}

	// Calculate overall status
	report.OverallStatus, report.ComplianceScore = s.calculateStatusFromAudit(report.Requirements, report.Controls)

	return report, nil
}

// generateExecutiveReport generates an executive summary report
func (s *ReportGenerationService) generateExecutiveReport(ctx context.Context, request ReportRequest, frameworks []string) (*ComplianceReport, error) {
	report := &ComplianceReport{
		ID:          fmt.Sprintf("executive_report_%s_%d", request.BusinessID, time.Now().Unix()),
		BusinessID:  request.BusinessID,
		ReportType:  ReportTypeExecutive,
		Title:       "Executive Compliance Summary",
		Description: "High-level compliance summary for executive review",
		GeneratedAt: time.Now(),
		Period:      "current",
	}

	// Get compliance check results
	checkReq := CheckRequest{
		BusinessID: request.BusinessID,
		Frameworks: frameworks,
		Options:    EvaluationOptions{},
	}

	checkResp, err := s.checkEngine.Check(ctx, checkReq)
	if err != nil {
		return nil, fmt.Errorf("failed to run compliance check: %w", err)
	}

	// Calculate overall status and score
	report.OverallStatus, report.ComplianceScore = s.calculateOverallStatus(checkResp)

	// Build high-level requirements summary
	report.Requirements = s.buildExecutiveRequirements(ctx, checkResp)

	// Get high-level recommendations
	if s.recommendations != nil {
		recommendations, err := s.recommendations.GenerateRecommendations(ctx, request.BusinessID, frameworks)
		if err != nil {
			s.logger.Warn("Failed to generate recommendations", "error", err.Error())
		} else {
			// Filter to high-priority recommendations only
			report.Recommendations = s.filterHighPriorityRecommendations(recommendations)
		}
	}

	return report, nil
}

// Helper methods

func (s *ReportGenerationService) getFrameworksToInclude(ctx context.Context, businessID, specificFramework string) ([]string, error) {
	if specificFramework != "" {
		return []string{specificFramework}, nil
	}

	// Check if tracking system is available
	if s.tracking == nil {
		return nil, fmt.Errorf("tracking system not initialized")
	}

	// Get all frameworks for the business
	summary, err := s.tracking.GetBusinessComplianceSummary(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get business compliance summary: %w", err)
	}

	frameworks := make([]string, 0, len(summary))
	for framework := range summary {
		frameworks = append(frameworks, framework)
	}

	return frameworks, nil
}

func (s *ReportGenerationService) calculateOverallStatus(checkResp *CheckResponse) (ComplianceStatus, float64) {
	if len(checkResp.Results) == 0 {
		return ComplianceStatusNotStarted, 0.0
	}

	totalPassed := checkResp.Passed
	totalFailed := checkResp.Failed
	total := totalPassed + totalFailed

	if total == 0 {
		return ComplianceStatusNotStarted, 0.0
	}

	score := float64(totalPassed) / float64(total) * 100.0

	// Determine status based on score
	if score >= 90.0 {
		return ComplianceStatusVerified, score
	} else if score >= 70.0 {
		return ComplianceStatusImplemented, score
	} else if score >= 30.0 {
		return ComplianceStatusInProgress, score
	} else {
		return ComplianceStatusNotStarted, score
	}
}

func (s *ReportGenerationService) buildRequirementsAndControls(ctx context.Context, checkResp *CheckResponse, includeDetails bool) ([]RequirementReport, []ControlReport) {
	var requirements []RequirementReport
	var controls []ControlReport

	for _, result := range checkResp.Results {
		// Convert framework check result to requirement reports
		reqReports := s.convertCheckResultToRequirements(ctx, result, includeDetails)
		requirements = append(requirements, reqReports...)

		// Extract controls from requirements
		for _, req := range reqReports {
			controls = append(controls, req.Controls...)
		}
	}

	return requirements, controls
}

func (s *ReportGenerationService) convertCheckResultToRequirements(ctx context.Context, result FrameworkCheckResult, includeDetails bool) []RequirementReport {
	// This is a simplified conversion - in a real implementation,
	// you would map the check results to actual requirement data
	var requirements []RequirementReport

	// Create a summary requirement for the framework
	requirement := RequirementReport{
		RequirementID:        result.FrameworkID,
		Title:                fmt.Sprintf("%s Framework Compliance", result.FrameworkID),
		Status:               ComplianceStatusImplemented, // Default status since ComplianceCheckResult doesn't have OverallStatus
		ImplementationStatus: ImplementationStatusImplemented,
		ComplianceScore:      float64(result.Summary.Passed) / float64(result.Summary.Passed+result.Summary.Failed) * 100.0,
		RiskLevel:            ComplianceRiskLevelMedium,
		Priority:             CompliancePriorityHigh,
		LastReviewed:         time.Now(),
		NextReview:           time.Now().Add(30 * 24 * time.Hour),
	}

	if includeDetails {
		// Add detailed control information
		requirement.Controls = s.buildDetailedControls(ctx, result)
	}

	requirements = append(requirements, requirement)
	return requirements
}

func (s *ReportGenerationService) buildDetailedControls(ctx context.Context, result FrameworkCheckResult) []ControlReport {
	// This would be populated with actual control data from the check results
	var controls []ControlReport

	// Create a summary control
	control := ControlReport{
		ControlID:            fmt.Sprintf("%s_summary", result.FrameworkID),
		Title:                fmt.Sprintf("%s Framework Summary Control", result.FrameworkID),
		Status:               ComplianceStatusImplemented, // Default status since ComplianceCheckResult doesn't have OverallStatus
		ImplementationStatus: ImplementationStatusImplemented,
		Effectiveness:        ControlEffectivenessEffective,
		LastTested:           &time.Time{},
		NextTestDate:         &time.Time{},
	}

	controls = append(controls, control)
	return controls
}

func (s *ReportGenerationService) convertGapAnalysisToReport(ctx context.Context, gapReport *GapAnalysisReport) ([]RequirementReport, []ControlReport) {
	var requirements []RequirementReport
	var controls []ControlReport

	// Convert requirement gaps to requirement reports
	for _, gap := range gapReport.RequirementGaps {
		requirement := RequirementReport{
			RequirementID:        gap.RequirementID,
			Title:                gap.Title,
			Status:               ComplianceStatusNonCompliant,
			ImplementationStatus: ImplementationStatusNotImplemented,
			ComplianceScore:      0.0,
			RiskLevel:            ComplianceRiskLevelMedium, // Default since RequirementGap doesn't have RiskLevel
			Priority:             CompliancePriorityMedium,  // Default since RequirementGap doesn't have Priority
			LastReviewed:         time.Now(),
			NextReview:           time.Now().Add(7 * 24 * time.Hour),
		}
		requirements = append(requirements, requirement)
	}

	// Convert control gaps to control reports
	for _, gap := range gapReport.ControlGaps {
		control := ControlReport{
			ControlID:            gap.ControlID,
			Title:                gap.Title,
			Status:               ComplianceStatusNonCompliant,
			ImplementationStatus: ImplementationStatusNotImplemented,
			Effectiveness:        ControlEffectivenessIneffective,
			LastTested:           &time.Time{},
			NextTestDate:         &time.Time{},
		}
		controls = append(controls, control)
	}

	return requirements, controls
}

func (s *ReportGenerationService) calculateStatusFromGaps(requirements []RequirementReport, controls []ControlReport) (ComplianceStatus, float64) {
	if len(requirements) == 0 && len(controls) == 0 {
		return ComplianceStatusVerified, 100.0
	}

	totalGaps := len(requirements) + len(controls)
	if totalGaps == 0 {
		return ComplianceStatusVerified, 100.0
	}

	// Calculate score based on gap severity
	score := 100.0 - float64(totalGaps)*10.0 // Simple scoring - each gap reduces score by 10%
	if score < 0 {
		score = 0.0
	}

	if score >= 90.0 {
		return ComplianceStatusVerified, score
	} else if score >= 70.0 {
		return ComplianceStatusImplemented, score
	} else if score >= 30.0 {
		return ComplianceStatusInProgress, score
	} else {
		return ComplianceStatusNotStarted, score
	}
}

func (s *ReportGenerationService) extractRemediationData(ctx context.Context, tracking *ComplianceTracking) ([]RequirementReport, []RemediationReport, []ExceptionReport) {
	var requirements []RequirementReport
	var remediationPlans []RemediationReport
	var exceptions []ExceptionReport

	for _, req := range tracking.Requirements {
		requirement := RequirementReport{
			RequirementID:        req.RequirementID,
			Status:               req.Status,
			ImplementationStatus: req.ImplementationStatus,
			ComplianceScore:      req.ComplianceScore,
			LastReviewed:         req.LastReviewed,
			NextReview:           req.NextReview,
		}

		// Add remediation plan if exists
		if req.RemediationPlan != nil {
			remediationPlan := RemediationReport{
				PlanID:     req.RemediationPlan.ID,
				Title:      req.RemediationPlan.Title,
				Status:     req.RemediationPlan.Status,
				Priority:   req.RemediationPlan.Priority,
				TargetDate: req.RemediationPlan.TargetDate,
				Progress:   req.RemediationPlan.Progress,
				AssignedTo: req.RemediationPlan.AssignedTo,
			}
			remediationPlans = append(remediationPlans, remediationPlan)
		}

		// Add exceptions
		for _, exception := range req.Exceptions {
			exceptionReport := ExceptionReport{
				ExceptionID:   exception.ID,
				RequirementID: exception.RequirementID,
				Type:          exception.Type,
				Reason:        exception.Reason,
				Status:        exception.Status,
				ApprovedBy:    exception.ApprovedBy,
				ApprovedAt:    exception.ApprovedAt,
				ExpiresAt:     exception.ExpiresAt,
			}
			exceptions = append(exceptions, exceptionReport)
		}

		requirements = append(requirements, requirement)
	}

	return requirements, remediationPlans, exceptions
}

func (s *ReportGenerationService) calculateStatusFromRemediation(requirements []RequirementReport, remediationPlans []RemediationReport) (ComplianceStatus, float64) {
	if len(requirements) == 0 {
		return ComplianceStatusVerified, 100.0
	}

	totalRequirements := len(requirements)
	compliantRequirements := 0
	activeRemediationPlans := 0

	for _, req := range requirements {
		if req.Status == ComplianceStatusVerified || req.Status == ComplianceStatusImplemented {
			compliantRequirements++
		}
	}

	for _, plan := range remediationPlans {
		if plan.Status == RemediationStatusInProgress || plan.Status == RemediationStatusNotStarted {
			activeRemediationPlans++
		}
	}

	score := float64(compliantRequirements) / float64(totalRequirements) * 100.0

	if score >= 90.0 {
		return ComplianceStatusVerified, score
	} else if score >= 70.0 {
		return ComplianceStatusImplemented, score
	} else if score >= 30.0 {
		return ComplianceStatusInProgress, score
	} else {
		return ComplianceStatusNotStarted, score
	}
}

func (s *ReportGenerationService) extractAuditData(ctx context.Context, tracking *ComplianceTracking) ([]RequirementReport, []ControlReport) {
	var requirements []RequirementReport
	var controls []ControlReport

	for _, req := range tracking.Requirements {
		requirement := RequirementReport{
			RequirementID:        req.RequirementID,
			Status:               req.Status,
			ImplementationStatus: req.ImplementationStatus,
			ComplianceScore:      req.ComplianceScore,
			LastReviewed:         req.LastReviewed,
			NextReview:           req.NextReview,
		}

		// Add controls with audit information
		for _, ctrl := range req.Controls {
			control := ControlReport{
				ControlID:            ctrl.ControlID,
				Status:               ctrl.Status,
				ImplementationStatus: ctrl.ImplementationStatus,
				Effectiveness:        ctrl.Effectiveness,
				LastTested:           ctrl.LastTested,
				NextTestDate:         ctrl.NextTestDate,
				TestResults:          ctrl.TestResults,
				Evidence:             ctrl.Evidence,
			}
			controls = append(controls, control)
		}

		requirements = append(requirements, requirement)
	}

	return requirements, controls
}

func (s *ReportGenerationService) calculateStatusFromAudit(requirements []RequirementReport, controls []ControlReport) (ComplianceStatus, float64) {
	if len(requirements) == 0 {
		return ComplianceStatusVerified, 100.0
	}

	totalRequirements := len(requirements)
	compliantRequirements := 0

	for _, req := range requirements {
		if req.Status == ComplianceStatusVerified || req.Status == ComplianceStatusImplemented {
			compliantRequirements++
		}
	}

	score := float64(compliantRequirements) / float64(totalRequirements) * 100.0

	if score >= 90.0 {
		return ComplianceStatusVerified, score
	} else if score >= 70.0 {
		return ComplianceStatusImplemented, score
	} else if score >= 30.0 {
		return ComplianceStatusInProgress, score
	} else {
		return ComplianceStatusNotStarted, score
	}
}

func (s *ReportGenerationService) buildExecutiveRequirements(ctx context.Context, checkResp *CheckResponse) []RequirementReport {
	var requirements []RequirementReport

	// Create high-level summary requirements for each framework
	for _, result := range checkResp.Results {
		requirement := RequirementReport{
			RequirementID:        result.FrameworkID,
			Title:                fmt.Sprintf("%s Framework", result.FrameworkID),
			Status:               ComplianceStatusImplemented, // Default status since ComplianceCheckResult doesn't have OverallStatus
			ImplementationStatus: ImplementationStatusImplemented,
			ComplianceScore:      float64(result.Summary.Passed) / float64(result.Summary.Passed+result.Summary.Failed) * 100.0,
			RiskLevel:            ComplianceRiskLevelMedium,
			Priority:             CompliancePriorityHigh,
			LastReviewed:         time.Now(),
			NextReview:           time.Now().Add(30 * 24 * time.Hour),
		}
		requirements = append(requirements, requirement)
	}

	return requirements
}

func (s *ReportGenerationService) filterHighPriorityRecommendations(recommendations []ComplianceRecommendation) []ComplianceRecommendation {
	var highPriority []ComplianceRecommendation

	for _, rec := range recommendations {
		if rec.Priority == CompliancePriorityCritical || rec.Priority == CompliancePriorityHigh {
			highPriority = append(highPriority, rec)
		}
	}

	return highPriority
}
