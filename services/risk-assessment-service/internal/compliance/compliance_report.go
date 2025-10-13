package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/audit"
)

// ComplianceReporter handles compliance reporting operations
type ComplianceReporter struct {
	auditLogger *audit.AuditLogger
	logger      *zap.Logger
	repository  ComplianceRepository
}

// ComplianceRepository defines the interface for compliance data persistence
type ComplianceRepository interface {
	SaveComplianceReport(ctx context.Context, report *audit.ComplianceReport) error
	GetComplianceReport(ctx context.Context, reportID string) (*audit.ComplianceReport, error)
	GetComplianceReports(ctx context.Context, tenantID string, reportType string, startDate, endDate time.Time) ([]audit.ComplianceReport, error)
	SaveReportTemplate(ctx context.Context, template *audit.ReportTemplate) error
	GetReportTemplate(ctx context.Context, templateID string) (*audit.ReportTemplate, error)
	GetReportTemplates(ctx context.Context, reportType string) ([]audit.ReportTemplate, error)
	DeleteComplianceReport(ctx context.Context, reportID string) error
}

// NewComplianceReporter creates a new compliance reporter
func NewComplianceReporter(auditLogger *audit.AuditLogger, repository ComplianceRepository, logger *zap.Logger) *ComplianceReporter {
	return &ComplianceReporter{
		auditLogger: auditLogger,
		logger:      logger,
		repository:  repository,
	}
}

// GenerateComplianceReport generates a compliance report
func (cr *ComplianceReporter) GenerateComplianceReport(ctx context.Context, req *GenerateReportRequest) (*audit.ComplianceReport, error) {
	// Get report template
	template, err := cr.repository.GetReportTemplate(ctx, req.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report template: %w", err)
	}

	// Generate report data based on template type
	reportData, err := cr.generateReportData(ctx, template, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report data: %w", err)
	}

	// Create compliance report
	report := &audit.ComplianceReport{
		ID:          generateReportID(),
		TenantID:    req.TenantID,
		ReportType:  template.Type,
		ReportName:  req.ReportName,
		Period:      req.Period,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Status:      "completed",
		Data:        reportData,
		GeneratedBy: req.GeneratedBy,
		GeneratedAt: time.Now(),
		Hash:        generateReportHash(reportData),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set expiration date if specified
	if req.ExpiresInDays > 0 {
		expiresAt := time.Now().AddDate(0, 0, req.ExpiresInDays)
		report.ExpiresAt = &expiresAt
	}

	// Save report
	if err := cr.repository.SaveComplianceReport(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to save compliance report: %w", err)
	}

	// Log report generation
	cr.auditLogger.LogAdminAction(ctx, req.TenantID, req.GeneratedBy, "generate_compliance_report", "compliance_report", report.ID, map[string]interface{}{
		"report_type": template.Type,
		"report_name": req.ReportName,
		"period":      req.Period,
	})

	return report, nil
}

// GetComplianceReport retrieves a compliance report
func (cr *ComplianceReporter) GetComplianceReport(ctx context.Context, reportID string) (*audit.ComplianceReport, error) {
	report, err := cr.repository.GetComplianceReport(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance report: %w", err)
	}

	// Log report access
	cr.auditLogger.LogDataAccess(ctx, "", "", "compliance_report", reportID, "read", map[string]interface{}{
		"report_type": report.ReportType,
		"report_name": report.ReportName,
	})

	return report, nil
}

// GetComplianceReports retrieves compliance reports for a tenant
func (cr *ComplianceReporter) GetComplianceReports(ctx context.Context, tenantID, reportType string, startDate, endDate time.Time) ([]audit.ComplianceReport, error) {
	reports, err := cr.repository.GetComplianceReports(ctx, tenantID, reportType, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance reports: %w", err)
	}

	// Log report listing
	cr.auditLogger.LogDataAccess(ctx, tenantID, "", "compliance_reports", "", "list", map[string]interface{}{
		"report_type": reportType,
		"count":       len(reports),
	})

	return reports, nil
}

// CreateReportTemplate creates a new report template
func (cr *ComplianceReporter) CreateReportTemplate(ctx context.Context, req *CreateTemplateRequest) (*audit.ReportTemplate, error) {
	template := &audit.ReportTemplate{
		ID:          generateTemplateID(),
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Template:    req.Template,
		Parameters:  req.Parameters,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := cr.repository.SaveReportTemplate(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to save report template: %w", err)
	}

	// Log template creation
	cr.auditLogger.LogAdminAction(ctx, "", req.CreatedBy, "create_report_template", "report_template", template.ID, map[string]interface{}{
		"template_name": req.Name,
		"template_type": req.Type,
	})

	return template, nil
}

// GetReportTemplate retrieves a report template
func (cr *ComplianceReporter) GetReportTemplate(ctx context.Context, templateID string) (*audit.ReportTemplate, error) {
	return cr.repository.GetReportTemplate(ctx, templateID)
}

// GetReportTemplates retrieves report templates by type
func (cr *ComplianceReporter) GetReportTemplates(ctx context.Context, reportType string) ([]audit.ReportTemplate, error) {
	return cr.repository.GetReportTemplates(ctx, reportType)
}

// DeleteComplianceReport deletes a compliance report
func (cr *ComplianceReporter) DeleteComplianceReport(ctx context.Context, reportID, deletedBy string) error {
	// Get report first to log deletion
	report, err := cr.repository.GetComplianceReport(ctx, reportID)
	if err != nil {
		return fmt.Errorf("failed to get compliance report: %w", err)
	}

	if err := cr.repository.DeleteComplianceReport(ctx, reportID); err != nil {
		return fmt.Errorf("failed to delete compliance report: %w", err)
	}

	// Log report deletion
	cr.auditLogger.LogAdminAction(ctx, report.TenantID, deletedBy, "delete_compliance_report", "compliance_report", reportID, map[string]interface{}{
		"report_type": report.ReportType,
		"report_name": report.ReportName,
	})

	return nil
}

// generateReportData generates report data based on template and parameters
func (cr *ComplianceReporter) generateReportData(ctx context.Context, template *audit.ReportTemplate, req *GenerateReportRequest) (map[string]interface{}, error) {
	switch template.Type {
	case "audit_summary":
		return cr.generateAuditSummaryReport(ctx, req)
	case "security_events":
		return cr.generateSecurityEventsReport(ctx, req)
	case "data_access":
		return cr.generateDataAccessReport(ctx, req)
	case "compliance_status":
		return cr.generateComplianceStatusReport(ctx, req)
	case "risk_assessment":
		return cr.generateRiskAssessmentReport(ctx, req)
	case "custom":
		return cr.generateCustomReport(ctx, template, req)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", template.Type)
	}
}

// generateAuditSummaryReport generates an audit summary report
func (cr *ComplianceReporter) generateAuditSummaryReport(ctx context.Context, req *GenerateReportRequest) (map[string]interface{}, error) {
	// Get audit statistics
	stats, err := cr.auditLogger.GetAuditStats(ctx, req.TenantID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit stats: %w", err)
	}

	// Get audit events for detailed analysis
	query := audit.AuditQuery{
		TenantID:  req.TenantID,
		StartDate: &req.StartDate,
		EndDate:   &req.EndDate,
		Limit:     1000,
	}

	events, err := cr.auditLogger.GetAuditEvents(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit events: %w", err)
	}

	// Generate summary data
	summary := map[string]interface{}{
		"period": map[string]interface{}{
			"start_date": req.StartDate,
			"end_date":   req.EndDate,
		},
		"statistics": stats,
		"events":     events,
		"summary": map[string]interface{}{
			"total_events":     stats.TotalEvents,
			"error_rate":       stats.ErrorRate,
			"average_duration": stats.AverageDuration,
			"top_actions":      getTopItems(stats.EventsByAction, 5),
			"top_users":        stats.TopUsers,
			"top_endpoints":    stats.TopEndpoints,
		},
	}

	return summary, nil
}

// generateSecurityEventsReport generates a security events report
func (cr *ComplianceReporter) generateSecurityEventsReport(ctx context.Context, req *GenerateReportRequest) (map[string]interface{}, error) {
	// Query for security-related events
	query := audit.AuditQuery{
		TenantID:  req.TenantID,
		Action:    "security_event",
		StartDate: &req.StartDate,
		EndDate:   &req.EndDate,
		Limit:     1000,
	}

	events, err := cr.auditLogger.GetAuditEvents(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get security events: %w", err)
	}

	// Analyze security events
	securityAnalysis := analyzeSecurityEvents(events)

	report := map[string]interface{}{
		"period": map[string]interface{}{
			"start_date": req.StartDate,
			"end_date":   req.EndDate,
		},
		"events":   events,
		"analysis": securityAnalysis,
		"summary": map[string]interface{}{
			"total_security_events": len(events),
			"failed_logins":         securityAnalysis["failed_logins"],
			"suspicious_activities": securityAnalysis["suspicious_activities"],
			"admin_actions":         securityAnalysis["admin_actions"],
		},
	}

	return report, nil
}

// generateDataAccessReport generates a data access report
func (cr *ComplianceReporter) generateDataAccessReport(ctx context.Context, req *GenerateReportRequest) (map[string]interface{}, error) {
	// Query for data access events
	query := audit.AuditQuery{
		TenantID:  req.TenantID,
		Action:    "data_access",
		StartDate: &req.StartDate,
		EndDate:   &req.EndDate,
		Limit:     1000,
	}

	events, err := cr.auditLogger.GetAuditEvents(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get data access events: %w", err)
	}

	// Analyze data access patterns
	accessAnalysis := analyzeDataAccess(events)

	report := map[string]interface{}{
		"period": map[string]interface{}{
			"start_date": req.StartDate,
			"end_date":   req.EndDate,
		},
		"events":   events,
		"analysis": accessAnalysis,
		"summary": map[string]interface{}{
			"total_access_events": len(events),
			"unique_users":        accessAnalysis["unique_users"],
			"unique_resources":    accessAnalysis["unique_resources"],
			"access_patterns":     accessAnalysis["access_patterns"],
		},
	}

	return report, nil
}

// generateComplianceStatusReport generates a compliance status report
func (cr *ComplianceReporter) generateComplianceStatusReport(ctx context.Context, req *GenerateReportRequest) (map[string]interface{}, error) {
	// This would integrate with compliance monitoring systems
	// For now, we'll generate a mock compliance status
	complianceStatus := map[string]interface{}{
		"soc2": map[string]interface{}{
			"status":     "compliant",
			"last_audit": "2024-01-15",
			"next_audit": "2024-07-15",
			"controls": map[string]interface{}{
				"security":             "compliant",
				"availability":         "compliant",
				"confidentiality":      "compliant",
				"processing_integrity": "compliant",
				"privacy":              "compliant",
			},
		},
		"gdpr": map[string]interface{}{
			"status":      "compliant",
			"last_review": "2024-01-10",
			"next_review": "2024-07-10",
		},
		"pci_dss": map[string]interface{}{
			"status":          "compliant",
			"level":           "1",
			"last_assessment": "2024-01-20",
		},
	}

	report := map[string]interface{}{
		"period": map[string]interface{}{
			"start_date": req.StartDate,
			"end_date":   req.EndDate,
		},
		"compliance_status": complianceStatus,
		"summary": map[string]interface{}{
			"overall_status": "compliant",
			"frameworks":     []string{"SOC2", "GDPR", "PCI-DSS"},
			"last_updated":   time.Now(),
		},
	}

	return report, nil
}

// generateRiskAssessmentReport generates a risk assessment report
func (cr *ComplianceReporter) generateRiskAssessmentReport(ctx context.Context, req *GenerateReportRequest) (map[string]interface{}, error) {
	// This would integrate with the risk assessment service
	// For now, we'll generate a mock risk assessment report
	riskAssessment := map[string]interface{}{
		"overall_risk_score": 0.15,
		"risk_level":         "low",
		"risk_factors": []map[string]interface{}{
			{
				"factor": "data_access_patterns",
				"score":  0.1,
				"status": "normal",
			},
			{
				"factor": "security_events",
				"score":  0.2,
				"status": "monitoring",
			},
			{
				"factor": "compliance_violations",
				"score":  0.0,
				"status": "compliant",
			},
		},
		"recommendations": []string{
			"Continue monitoring security events",
			"Review data access patterns monthly",
			"Maintain current compliance controls",
		},
	}

	report := map[string]interface{}{
		"period": map[string]interface{}{
			"start_date": req.StartDate,
			"end_date":   req.EndDate,
		},
		"risk_assessment": riskAssessment,
		"summary": map[string]interface{}{
			"assessment_date": time.Now(),
			"next_assessment": time.Now().AddDate(0, 3, 0), // 3 months
		},
	}

	return report, nil
}

// generateCustomReport generates a custom report based on template
func (cr *ComplianceReporter) generateCustomReport(ctx context.Context, template *audit.ReportTemplate, req *GenerateReportRequest) (map[string]interface{}, error) {
	// For custom reports, we would execute the template logic
	// This is a simplified implementation
	customData := map[string]interface{}{
		"template_id":   template.ID,
		"template_name": template.Name,
		"parameters":    req.Parameters,
		"generated_at":  time.Now(),
		"period": map[string]interface{}{
			"start_date": req.StartDate,
			"end_date":   req.EndDate,
		},
	}

	return customData, nil
}

// Request structures
type GenerateReportRequest struct {
	TenantID      string                 `json:"tenant_id"`
	TemplateID    string                 `json:"template_id"`
	ReportName    string                 `json:"report_name"`
	Period        string                 `json:"period"`
	StartDate     time.Time              `json:"start_date"`
	EndDate       time.Time              `json:"end_date"`
	Parameters    map[string]interface{} `json:"parameters"`
	GeneratedBy   string                 `json:"generated_by"`
	ExpiresInDays int                    `json:"expires_in_days"`
}

type CreateTemplateRequest struct {
	Name        string                  `json:"name"`
	Type        string                  `json:"type"`
	Description string                  `json:"description"`
	Template    map[string]interface{}  `json:"template"`
	Parameters  []audit.ReportParameter `json:"parameters"`
	CreatedBy   string                  `json:"created_by"`
}

// Helper functions
func generateReportID() string {
	return fmt.Sprintf("rpt_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

func generateTemplateID() string {
	return fmt.Sprintf("tpl_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

func generateReportHash(data map[string]interface{}) string {
	jsonData, _ := json.Marshal(data)
	return fmt.Sprintf("%x", jsonData) // Simplified hash
}

func getTopItems(items map[string]int64, limit int) []map[string]interface{} {
	var result []map[string]interface{}

	// Sort by count (simplified - in production, use proper sorting)
	count := 0
	for key, value := range items {
		if count >= limit {
			break
		}
		result = append(result, map[string]interface{}{
			"key":   key,
			"count": value,
		})
		count++
	}

	return result
}

func analyzeSecurityEvents(events []audit.AuditEvent) map[string]interface{} {
	analysis := map[string]interface{}{
		"failed_logins":         0,
		"suspicious_activities": 0,
		"admin_actions":         0,
	}

	for _, event := range events {
		if event.Action == "login_failed" {
			analysis["failed_logins"] = analysis["failed_logins"].(int) + 1
		}
		if event.Status >= 400 {
			analysis["suspicious_activities"] = analysis["suspicious_activities"].(int) + 1
		}
		if event.Resource == "admin" {
			analysis["admin_actions"] = analysis["admin_actions"].(int) + 1
		}
	}

	return analysis
}

func analyzeDataAccess(events []audit.AuditEvent) map[string]interface{} {
	uniqueUsers := make(map[string]bool)
	uniqueResources := make(map[string]bool)
	accessPatterns := make(map[string]int)

	for _, event := range events {
		if event.UserID != "" {
			uniqueUsers[event.UserID] = true
		}
		if event.Resource != "" {
			uniqueResources[event.Resource] = true
		}
		pattern := fmt.Sprintf("%s_%s", event.Action, event.Resource)
		accessPatterns[pattern]++
	}

	return map[string]interface{}{
		"unique_users":     len(uniqueUsers),
		"unique_resources": len(uniqueResources),
		"access_patterns":  accessPatterns,
	}
}
