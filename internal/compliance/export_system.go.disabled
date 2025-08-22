package compliance

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ExportSystem provides comprehensive compliance data export functionality
type ExportSystem struct {
	logger        *observability.Logger
	statusSystem  *ComplianceStatusSystem
	reportService *ReportGenerationService
	alertSystem   *AlertSystem
}

// ExportRequest represents a request to export compliance data
type ExportRequest struct {
	BusinessID     string                 `json:"business_id"`
	ExportType     ExportType             `json:"export_type"`
	Format         ExportFormat           `json:"format"`
	DateRange      *DateRange             `json:"date_range,omitempty"`
	Frameworks     []string               `json:"frameworks,omitempty"`
	IncludeDetails bool                   `json:"include_details,omitempty"`
	Filters        map[string]interface{} `json:"filters,omitempty"`
	GeneratedBy    string                 `json:"generated_by"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ExportType represents the type of data to export
type ExportType string

const (
	ExportTypeStatus        ExportType = "status"
	ExportTypeReports       ExportType = "reports"
	ExportTypeAlerts        ExportType = "alerts"
	ExportTypeAuditTrail    ExportType = "audit_trail"
	ExportTypeRequirements  ExportType = "requirements"
	ExportTypeControls      ExportType = "controls"
	ExportTypeExceptions    ExportType = "exceptions"
	ExportTypeRemediation   ExportType = "remediation"
	ExportTypeComprehensive ExportType = "comprehensive"
)

// ExportFormat represents the export format
type ExportFormat string

const (
	ExportFormatJSON  ExportFormat = "json"
	ExportFormatCSV   ExportFormat = "csv"
	ExportFormatExcel ExportFormat = "excel"
	ExportFormatPDF   ExportFormat = "pdf"
)

// ExportResult represents the result of an export operation
type ExportResult struct {
	ID          string                 `json:"id"`
	BusinessID  string                 `json:"business_id"`
	ExportType  ExportType             `json:"export_type"`
	Format      ExportFormat           `json:"format"`
	Data        interface{}            `json:"data"`
	RecordCount int                    `json:"record_count"`
	FileSize    int64                  `json:"file_size,omitempty"`
	DownloadURL string                 `json:"download_url,omitempty"`
	GeneratedAt time.Time              `json:"generated_at"`
	GeneratedBy string                 `json:"generated_by"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewExportSystem creates a new compliance export system
func NewExportSystem(logger *observability.Logger, statusSystem *ComplianceStatusSystem, reportService *ReportGenerationService, alertSystem *AlertSystem) *ExportSystem {
	return &ExportSystem{
		logger:        logger,
		statusSystem:  statusSystem,
		reportService: reportService,
		alertSystem:   alertSystem,
	}
}

// ExportData exports compliance data based on the request
func (s *ExportSystem) ExportData(ctx context.Context, request ExportRequest) (*ExportResult, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Starting compliance data export",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"export_type", request.ExportType,
		"format", request.Format,
	)

	// Validate request
	if request.BusinessID == "" {
		return nil, fmt.Errorf("business_id is required")
	}

	if request.ExportType == "" {
		return nil, fmt.Errorf("export_type is required")
	}

	if request.Format == "" {
		request.Format = ExportFormatJSON // Default to JSON
	}

	if request.GeneratedBy == "" {
		request.GeneratedBy = "system"
	}

	// Generate export ID
	exportID := fmt.Sprintf("export_%s_%s_%d", request.BusinessID, request.ExportType, time.Now().UnixNano())

	// Export data based on type
	var data interface{}
	var recordCount int
	var err error

	switch request.ExportType {
	case ExportTypeStatus:
		data, recordCount, err = s.exportStatusData(ctx, request)
	case ExportTypeReports:
		data, recordCount, err = s.exportReportData(ctx, request)
	case ExportTypeAlerts:
		data, recordCount, err = s.exportAlertData(ctx, request)
	case ExportTypeAuditTrail:
		data, recordCount, err = s.exportAuditTrailData(ctx, request)
	case ExportTypeRequirements:
		data, recordCount, err = s.exportRequirementsData(ctx, request)
	case ExportTypeControls:
		data, recordCount, err = s.exportControlsData(ctx, request)
	case ExportTypeExceptions:
		data, recordCount, err = s.exportExceptionsData(ctx, request)
	case ExportTypeRemediation:
		data, recordCount, err = s.exportRemediationData(ctx, request)
	case ExportTypeComprehensive:
		data, recordCount, err = s.exportComprehensiveData(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported export type: %s", request.ExportType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to export %s data: %w", request.ExportType, err)
	}

	// Format data based on requested format
	formattedData, fileSize, err := s.formatData(data, request.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to format data: %w", err)
	}

	// Set expiration (24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour)

	result := &ExportResult{
		ID:          exportID,
		BusinessID:  request.BusinessID,
		ExportType:  request.ExportType,
		Format:      request.Format,
		Data:        formattedData,
		RecordCount: recordCount,
		FileSize:    fileSize,
		GeneratedAt: time.Now(),
		GeneratedBy: request.GeneratedBy,
		ExpiresAt:   &expiresAt,
		Metadata:    request.Metadata,
	}

	s.logger.Info("Compliance data export completed successfully",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"export_type", request.ExportType,
		"format", request.Format,
		"record_count", recordCount,
		"file_size", fileSize,
	)

	return result, nil
}

// exportStatusData exports compliance status data
func (s *ExportSystem) exportStatusData(ctx context.Context, request ExportRequest) (interface{}, int, error) {
	// Get compliance status
	status, err := s.statusSystem.GetComplianceStatus(ctx, request.BusinessID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get compliance status: %w", err)
	}

	// Get status history if date range is specified
	var history []StatusChange
	if request.DateRange != nil {
		history, err = s.statusSystem.GetStatusHistory(ctx, request.BusinessID, request.DateRange.StartDate, request.DateRange.EndDate)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get status history: %w", err)
		}
	}

	// Get status alerts
	alerts, err := s.statusSystem.GetStatusAlerts(ctx, request.BusinessID, "")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get status alerts: %w", err)
	}

	exportData := map[string]interface{}{
		"business_id":    request.BusinessID,
		"export_type":    "status",
		"exported_at":    time.Now(),
		"current_status": status,
		"history":        history,
		"alerts":         alerts,
		"summary": map[string]interface{}{
			"overall_score":     status.OverallScore,
			"overall_status":    status.OverallStatus,
			"risk_level":        status.RiskLevel,
			"trend":             status.Trend,
			"framework_count":   len(status.FrameworkStatuses),
			"requirement_count": len(status.RequirementStatuses),
			"control_count":     len(status.ControlStatuses),
			"history_count":     len(history),
			"alert_count":       len(alerts),
		},
	}

	recordCount := 1 + len(history) + len(alerts) // status + history + alerts
	return exportData, recordCount, nil
}

// exportReportData exports compliance report data
func (s *ExportSystem) exportReportData(ctx context.Context, request ExportRequest) (interface{}, int, error) {
	// Generate comprehensive report
	reportRequest := ReportRequest{
		BusinessID:     request.BusinessID,
		Framework:      "", // All frameworks
		ReportType:     ReportTypeExecutive,
		Format:         ReportFormatJSON,
		DateRange:      request.DateRange,
		IncludeDetails: request.IncludeDetails,
		GeneratedBy:    request.GeneratedBy,
		Metadata:       request.Metadata,
	}

	report, err := s.reportService.GenerateComplianceReport(ctx, reportRequest)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to generate compliance report: %w", err)
	}

	exportData := map[string]interface{}{
		"business_id": request.BusinessID,
		"export_type": "reports",
		"exported_at": time.Now(),
		"report":      report,
		"summary": map[string]interface{}{
			"report_id":         report.ID,
			"report_type":       report.ReportType,
			"overall_status":    report.OverallStatus,
			"compliance_score":  report.ComplianceScore,
			"requirement_count": len(report.Requirements),
			"control_count":     len(report.Controls),
			"exception_count":   len(report.Exceptions),
			"remediation_count": len(report.RemediationPlans),
		},
	}

	recordCount := 1 + len(report.Requirements) + len(report.Controls) + len(report.Exceptions) + len(report.RemediationPlans)
	return exportData, recordCount, nil
}

// exportAlertData exports alert data
func (s *ExportSystem) exportAlertData(ctx context.Context, request ExportRequest) (interface{}, int, error) {
	// Get alert analytics
	analytics, err := s.alertSystem.GetAlertAnalytics(ctx, request.BusinessID, "30d")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get alert analytics: %w", err)
	}

	// Get all alerts
	alerts, err := s.statusSystem.GetStatusAlerts(ctx, request.BusinessID, "")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get alerts: %w", err)
	}

	// Filter alerts by date range if specified
	var filteredAlerts []StatusAlert
	if request.DateRange != nil {
		for _, alert := range alerts {
			if alert.TriggeredAt.After(request.DateRange.StartDate) && alert.TriggeredAt.Before(request.DateRange.EndDate) {
				filteredAlerts = append(filteredAlerts, alert)
			}
		}
	} else {
		filteredAlerts = alerts
	}

	exportData := map[string]interface{}{
		"business_id": request.BusinessID,
		"export_type": "alerts",
		"exported_at": time.Now(),
		"analytics":   analytics,
		"alerts":      filteredAlerts,
		"summary": map[string]interface{}{
			"total_alerts":        analytics.TotalAlerts,
			"active_alerts":       analytics.ActiveAlerts,
			"resolved_alerts":     analytics.ResolvedAlerts,
			"critical_alerts":     analytics.AlertsBySeverity["critical"],
			"high_alerts":         analytics.AlertsBySeverity["high"],
			"medium_alerts":       analytics.AlertsBySeverity["medium"],
			"low_alerts":          analytics.AlertsBySeverity["low"],
			"avg_resolution_time": analytics.AverageResolutionTime.String(),
		},
	}

	recordCount := len(filteredAlerts)
	return exportData, recordCount, nil
}

// exportAuditTrailData exports audit trail data
func (s *ExportSystem) exportAuditTrailData(ctx context.Context, request ExportRequest) (interface{}, int, error) {
	// This would typically query a database for audit trail data
	// For now, we'll create a placeholder structure
	auditTrail := []ComplianceAuditTrail{
		{
			ID:          "audit_1",
			BusinessID:  request.BusinessID,
			Framework:   "SOC2",
			Action:      AuditActionUpdate,
			Description: "Compliance status updated",
			UserID:      request.GeneratedBy,
			UserName:    request.GeneratedBy,
			UserRole:    "compliance_officer",
			Timestamp:   time.Now().Add(-1 * time.Hour),
			IPAddress:   "192.168.1.100",
			UserAgent:   "Mozilla/5.0...",
			SessionID:   "session_123",
			RequestID:   ctx.Value("request_id").(string),
			OldValue:    "not_started",
			NewValue:    "in_progress",
		},
	}

	exportData := map[string]interface{}{
		"business_id": request.BusinessID,
		"export_type": "audit_trail",
		"exported_at": time.Now(),
		"audit_trail": auditTrail,
		"summary": map[string]interface{}{
			"total_entries": len(auditTrail),
			"date_range":    request.DateRange,
		},
	}

	recordCount := len(auditTrail)
	return exportData, recordCount, nil
}

// exportRequirementsData exports requirements data
func (s *ExportSystem) exportRequirementsData(ctx context.Context, request ExportRequest) (interface{}, int, error) {
	// Get compliance status to extract requirements
	status, err := s.statusSystem.GetComplianceStatus(ctx, request.BusinessID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get compliance status: %w", err)
	}

	// Convert requirement statuses to a more exportable format
	var requirements []map[string]interface{}
	for reqID, reqStatus := range status.RequirementStatuses {
		requirement := map[string]interface{}{
			"requirement_id":         reqID,
			"title":                  reqStatus.Title,
			"status":                 reqStatus.Status,
			"implementation_status":  reqStatus.ImplementationStatus,
			"compliance_score":       reqStatus.Score,
			"risk_level":             reqStatus.RiskLevel,
			"priority":               reqStatus.Priority,
			"last_reviewed":          reqStatus.LastReviewed,
			"next_review":            reqStatus.NextReview,
			"reviewer":               reqStatus.Reviewer,
			"evidence_count":         reqStatus.EvidenceCount,
			"exception_count":        reqStatus.ExceptionCount,
			"remediation_plan_count": reqStatus.RemediationPlanCount,
			"trend":                  reqStatus.Trend,
			"trend_strength":         reqStatus.TrendStrength,
		}
		requirements = append(requirements, requirement)
	}

	exportData := map[string]interface{}{
		"business_id":  request.BusinessID,
		"export_type":  "requirements",
		"exported_at":  time.Now(),
		"requirements": requirements,
		"summary": map[string]interface{}{
			"total_requirements":  len(requirements),
			"implemented_count":   countByStatus(requirements, "status", "implemented"),
			"in_progress_count":   countByStatus(requirements, "status", "in_progress"),
			"not_started_count":   countByStatus(requirements, "status", "not_started"),
			"non_compliant_count": countByStatus(requirements, "status", "non_compliant"),
		},
	}

	recordCount := len(requirements)
	return exportData, recordCount, nil
}

// exportControlsData exports controls data
func (s *ExportSystem) exportControlsData(ctx context.Context, request ExportRequest) (interface{}, int, error) {
	// Get compliance status to extract controls
	status, err := s.statusSystem.GetComplianceStatus(ctx, request.BusinessID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get compliance status: %w", err)
	}

	// Convert control statuses to a more exportable format
	var controls []map[string]interface{}
	for controlID, controlStatus := range status.ControlStatuses {
		control := map[string]interface{}{
			"control_id":            controlID,
			"title":                 controlStatus.Title,
			"status":                controlStatus.Status,
			"implementation_status": controlStatus.ImplementationStatus,
			"effectiveness":         controlStatus.Effectiveness,
			"compliance_score":      controlStatus.Score,
			"last_tested":           controlStatus.LastTested,
			"next_test_date":        controlStatus.NextTestDate,
			"test_result_count":     controlStatus.TestResultCount,
			"pass_count":            controlStatus.PassCount,
			"fail_count":            controlStatus.FailCount,
			"evidence_count":        controlStatus.EvidenceCount,
			"trend":                 controlStatus.Trend,
			"trend_strength":        controlStatus.TrendStrength,
		}
		controls = append(controls, control)
	}

	exportData := map[string]interface{}{
		"business_id": request.BusinessID,
		"export_type": "controls",
		"exported_at": time.Now(),
		"controls":    controls,
		"summary": map[string]interface{}{
			"total_controls":    len(controls),
			"implemented_count": countByStatus(controls, "status", "implemented"),
			"in_progress_count": countByStatus(controls, "status", "in_progress"),
			"not_started_count": countByStatus(controls, "status", "not_started"),
			"effective_count":   countByStatus(controls, "effectiveness", "effective"),
		},
	}

	recordCount := len(controls)
	return exportData, recordCount, nil
}

// exportExceptionsData exports exceptions data
func (s *ExportSystem) exportExceptionsData(ctx context.Context, request ExportRequest) (interface{}, int, error) {
	// This would typically query a database for exceptions data
	// For now, we'll create a placeholder structure
	exceptions := []ComplianceException{
		{
			ID:             "exception_1",
			RequirementID:  "req_001",
			Type:           ExceptionTypeTemporary,
			Reason:         "System maintenance",
			Justification:  "Planned maintenance window",
			RiskAssessment: "Low risk - temporary exception",
			MitigationPlan: "Resume compliance after maintenance",
			ApprovedBy:     "compliance_officer",
			ApprovedAt:     time.Now().Add(-24 * time.Hour),
			ExpiresAt:      &[]time.Time{time.Now().Add(7 * 24 * time.Hour)}[0],
			Status:         ExceptionStatusApproved,
			Notes:          "Standard maintenance exception",
		},
	}

	exportData := map[string]interface{}{
		"business_id": request.BusinessID,
		"export_type": "exceptions",
		"exported_at": time.Now(),
		"exceptions":  exceptions,
		"summary": map[string]interface{}{
			"total_exceptions": len(exceptions),
			"active_count":     countByStatus(exceptions, "status", "active"),
			"expired_count":    countByStatus(exceptions, "status", "expired"),
		},
	}

	recordCount := len(exceptions)
	return exportData, recordCount, nil
}

// exportRemediationData exports remediation data
func (s *ExportSystem) exportRemediationData(ctx context.Context, request ExportRequest) (interface{}, int, error) {
	// This would typically query a database for remediation data
	// For now, we'll create a placeholder structure
	remediationPlans := []RemediationPlan{
		{
			ID:            "remediation_1",
			RequirementID: "req_002",
			Title:         "Implement Access Controls",
			Description:   "Implement proper access controls for sensitive data",
			Priority:      CompliancePriorityHigh,
			Status:        RemediationStatusInProgress,
			TargetDate:    time.Now().Add(30 * 24 * time.Hour),
			AssignedTo:    "security_team",
			Budget:        50000.0,
			Progress:      75.0,
			Notes:         "On track for completion",
		},
	}

	exportData := map[string]interface{}{
		"business_id":       request.BusinessID,
		"export_type":       "remediation",
		"exported_at":       time.Now(),
		"remediation_plans": remediationPlans,
		"summary": map[string]interface{}{
			"total_plans":       len(remediationPlans),
			"completed_count":   countByStatus(remediationPlans, "status", "completed"),
			"in_progress_count": countByStatus(remediationPlans, "status", "in_progress"),
			"planned_count":     countByStatus(remediationPlans, "status", "planned"),
		},
	}

	recordCount := len(remediationPlans)
	return exportData, recordCount, nil
}

// exportComprehensiveData exports all compliance data in a comprehensive format
func (s *ExportSystem) exportComprehensiveData(ctx context.Context, request ExportRequest) (interface{}, int, error) {
	// Export all data types
	statusData, _, err := s.exportStatusData(ctx, request)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to export status data: %w", err)
	}

	reportData, _, err := s.exportReportData(ctx, request)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to export report data: %w", err)
	}

	alertData, _, err := s.exportAlertData(ctx, request)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to export alert data: %w", err)
	}

	requirementsData, _, err := s.exportRequirementsData(ctx, request)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to export requirements data: %w", err)
	}

	controlsData, _, err := s.exportControlsData(ctx, request)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to export controls data: %w", err)
	}

	comprehensiveData := map[string]interface{}{
		"business_id":  request.BusinessID,
		"export_type":  "comprehensive",
		"exported_at":  time.Now(),
		"status":       statusData,
		"reports":      reportData,
		"alerts":       alertData,
		"requirements": requirementsData,
		"controls":     controlsData,
		"export_metadata": map[string]interface{}{
			"generated_by": request.GeneratedBy,
			"date_range":   request.DateRange,
			"frameworks":   request.Frameworks,
			"filters":      request.Filters,
		},
	}

	// Calculate total record count
	totalRecords := 0
	if statusMap, ok := statusData.(map[string]interface{}); ok {
		if summary, ok := statusMap["summary"].(map[string]interface{}); ok {
			if count, ok := summary["history_count"].(int); ok {
				totalRecords += count
			}
			if count, ok := summary["alert_count"].(int); ok {
				totalRecords += count
			}
		}
	}

	if requirementsMap, ok := requirementsData.(map[string]interface{}); ok {
		if requirements, ok := requirementsMap["requirements"].([]map[string]interface{}); ok {
			totalRecords += len(requirements)
		}
	}

	if controlsMap, ok := controlsData.(map[string]interface{}); ok {
		if controls, ok := controlsMap["controls"].([]map[string]interface{}); ok {
			totalRecords += len(controls)
		}
	}

	return comprehensiveData, totalRecords, nil
}

// formatData formats the data according to the requested format
func (s *ExportSystem) formatData(data interface{}, format ExportFormat) (interface{}, int64, error) {
	switch format {
	case ExportFormatJSON:
		return s.formatAsJSON(data)
	case ExportFormatCSV:
		return s.formatAsCSV(data)
	case ExportFormatExcel:
		return s.formatAsExcel(data)
	case ExportFormatPDF:
		return s.formatAsPDF(data)
	default:
		return nil, 0, fmt.Errorf("unsupported format: %s", format)
	}
}

// formatAsJSON formats data as JSON
func (s *ExportSystem) formatAsJSON(data interface{}) (interface{}, int64, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonData), int64(len(jsonData)), nil
}

// formatAsCSV formats data as CSV
func (s *ExportSystem) formatAsCSV(data interface{}) (interface{}, int64, error) {
	// Convert data to CSV format
	// This is a simplified implementation - in practice, you'd want more sophisticated CSV generation
	var csvData strings.Builder
	writer := csv.NewWriter(&csvData)

	// For now, we'll create a simple CSV with key-value pairs
	if dataMap, ok := data.(map[string]interface{}); ok {
		// Write header
		writer.Write([]string{"Key", "Value"})

		// Write data
		for key, value := range dataMap {
			writer.Write([]string{key, fmt.Sprintf("%v", value)})
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, 0, fmt.Errorf("failed to write CSV: %w", err)
	}

	csvString := csvData.String()
	return csvString, int64(len(csvString)), nil
}

// formatAsExcel formats data as Excel
func (s *ExportSystem) formatAsExcel(data interface{}) (interface{}, int64, error) {
	// This would use a library like excelize to create Excel files
	// For now, we'll return a placeholder
	_ = "Excel format not yet implemented - returning JSON instead"
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal data for Excel: %w", err)
	}

	return string(jsonData), int64(len(jsonData)), nil
}

// formatAsPDF formats data as PDF
func (s *ExportSystem) formatAsPDF(data interface{}) (interface{}, int64, error) {
	// This would use a library like wkhtmltopdf or similar to create PDF files
	// For now, we'll return a placeholder
	_ = "PDF format not yet implemented - returning JSON instead"
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal data for PDF: %w", err)
	}

	return string(jsonData), int64(len(jsonData)), nil
}

// countByStatus counts items by a specific status field
func countByStatus(items interface{}, fieldName, statusValue string) int {
	count := 0

	switch v := items.(type) {
	case []map[string]interface{}:
		for _, item := range v {
			if status, ok := item[fieldName].(string); ok && status == statusValue {
				count++
			}
		}
	case []ComplianceException:
		for _, item := range v {
			if fieldName == "status" && string(item.Status) == statusValue {
				count++
			}
		}
	case []RemediationPlan:
		for _, item := range v {
			if fieldName == "status" && string(item.Status) == statusValue {
				count++
			}
		}
	}

	return count
}
