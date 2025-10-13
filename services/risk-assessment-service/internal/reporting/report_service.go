package reporting

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// ReportService provides report generation and management
type ReportService interface {
	// GenerateReport generates a new report
	GenerateReport(ctx context.Context, request *ReportRequest) (*ReportResponse, error)

	// GetReport retrieves a report by ID
	GetReport(ctx context.Context, tenantID, reportID string) (*Report, error)

	// ListReports lists reports with filters
	ListReports(ctx context.Context, filter *ReportFilter) (*ReportListResponse, error)

	// DeleteReport deletes a report
	DeleteReport(ctx context.Context, tenantID, reportID string) error

	// GetReportMetrics gets report usage metrics
	GetReportMetrics(ctx context.Context, tenantID string) (*ReportMetrics, error)

	// Template Management
	CreateTemplate(ctx context.Context, request *ReportTemplateRequest) (*ReportTemplateResponse, error)
	GetTemplate(ctx context.Context, tenantID, templateID string) (*ReportTemplate, error)
	ListTemplates(ctx context.Context, filter *ReportTemplateFilter) (*ReportTemplateListResponse, error)
	UpdateTemplate(ctx context.Context, tenantID, templateID string, request *ReportTemplateRequest) (*ReportTemplateResponse, error)
	DeleteTemplate(ctx context.Context, tenantID, templateID string) error

	// Scheduled Reports
	CreateScheduledReport(ctx context.Context, request *ScheduledReportRequest) (*ScheduledReportResponse, error)
	GetScheduledReport(ctx context.Context, tenantID, scheduledReportID string) (*ScheduledReport, error)
	ListScheduledReports(ctx context.Context, filter *ScheduledReportFilter) (*ScheduledReportListResponse, error)
	UpdateScheduledReport(ctx context.Context, tenantID, scheduledReportID string, request *ScheduledReportRequest) (*ScheduledReportResponse, error)
	DeleteScheduledReport(ctx context.Context, tenantID, scheduledReportID string) error
	RunScheduledReport(ctx context.Context, tenantID, scheduledReportID string) (*ReportResponse, error)
}

// DefaultReportService implements ReportService
type DefaultReportService struct {
	repository    ReportRepository
	templateRepo  ReportTemplateRepository
	schedulerRepo ScheduledReportRepository
	dataProvider  ReportDataProvider
	generator     ReportGenerator
	scheduler     ReportScheduler
	logger        *zap.Logger
}

// ReportRepository defines the interface for report data access
type ReportRepository interface {
	SaveReport(ctx context.Context, report *Report) error
	GetReport(ctx context.Context, tenantID, reportID string) (*Report, error)
	ListReports(ctx context.Context, filter *ReportFilter) ([]*Report, error)
	DeleteReport(ctx context.Context, tenantID, reportID string) error
	GetReportMetrics(ctx context.Context, tenantID string) (*ReportMetrics, error)
	UpdateReportStatus(ctx context.Context, tenantID, reportID string, status ReportStatus, errorMsg string) error
	UpdateReportFile(ctx context.Context, tenantID, reportID string, fileSize int64, downloadURL string) error
}

// ReportTemplateRepository defines the interface for template data access
type ReportTemplateRepository interface {
	SaveTemplate(ctx context.Context, template *ReportTemplate) error
	GetTemplate(ctx context.Context, tenantID, templateID string) (*ReportTemplate, error)
	ListTemplates(ctx context.Context, filter *ReportTemplateFilter) ([]*ReportTemplate, error)
	UpdateTemplate(ctx context.Context, template *ReportTemplate) error
	DeleteTemplate(ctx context.Context, tenantID, templateID string) error
}

// ScheduledReportRepository defines the interface for scheduled report data access
type ScheduledReportRepository interface {
	SaveScheduledReport(ctx context.Context, scheduledReport *ScheduledReport) error
	GetScheduledReport(ctx context.Context, tenantID, scheduledReportID string) (*ScheduledReport, error)
	ListScheduledReports(ctx context.Context, filter *ScheduledReportFilter) ([]*ScheduledReport, error)
	UpdateScheduledReport(ctx context.Context, scheduledReport *ScheduledReport) error
	DeleteScheduledReport(ctx context.Context, tenantID, scheduledReportID string) error
	GetScheduledReportsToRun(ctx context.Context) ([]*ScheduledReport, error)
	UpdateScheduledReportLastRun(ctx context.Context, tenantID, scheduledReportID string, lastRunAt time.Time, nextRunAt *time.Time) error
}

// ReportDataProvider defines the interface for providing report data
type ReportDataProvider interface {
	GetRiskAssessments(ctx context.Context, tenantID string, filters *ReportFilters) ([]*models.RiskAssessment, error)
	GetRiskPredictions(ctx context.Context, tenantID string, filters *ReportFilters) ([]*models.RiskPrediction, error)
	GetBatchJobs(ctx context.Context, tenantID string, filters *ReportFilters) ([]*BatchJobData, error)
	GetComplianceData(ctx context.Context, tenantID string, filters *ReportFilters) (*ComplianceData, error)
	GetPerformanceData(ctx context.Context, tenantID string, filters *ReportFilters) (*PerformanceData, error)
	GetDashboardData(ctx context.Context, tenantID string, filters *ReportFilters) (*RiskDashboard, error)
}

// ReportGenerator defines the interface for generating report files
type ReportGenerator interface {
	GeneratePDF(ctx context.Context, report *Report, template *ReportTemplate) ([]byte, error)
	GenerateExcel(ctx context.Context, report *Report, template *ReportTemplate) ([]byte, error)
	GenerateCSV(ctx context.Context, report *Report, template *ReportTemplate) ([]byte, error)
	GenerateJSON(ctx context.Context, report *Report, template *ReportTemplate) ([]byte, error)
	GenerateHTML(ctx context.Context, report *Report, template *ReportTemplate) ([]byte, error)
}

// ReportScheduler defines the interface for scheduling reports
type ReportScheduler interface {
	ScheduleReport(ctx context.Context, scheduledReport *ScheduledReport) error
	UnscheduleReport(ctx context.Context, scheduledReportID string) error
	GetScheduledReports(ctx context.Context) ([]*ScheduledReport, error)
	RunScheduledReports(ctx context.Context) error
}

// ReportTemplateRequest represents a request to create/update a template
type ReportTemplateRequest struct {
	Name        string                 `json:"name" validate:"required,min=1,max=255"`
	Type        ReportType             `json:"type" validate:"required"`
	Description string                 `json:"description,omitempty"`
	Template    ReportTemplateConfig   `json:"template" validate:"required"`
	IsPublic    bool                   `json:"is_public,omitempty"`
	CreatedBy   string                 `json:"created_by" validate:"required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ReportTemplateResponse represents a template response
type ReportTemplateResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Type        ReportType `json:"type"`
	Description string     `json:"description"`
	IsPublic    bool       `json:"is_public"`
	IsDefault   bool       `json:"is_default"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ReportTemplateListResponse represents a list of templates
type ReportTemplateListResponse struct {
	Templates []ReportTemplateResponse `json:"templates"`
	Total     int                      `json:"total"`
	Page      int                      `json:"page"`
	PageSize  int                      `json:"page_size"`
}

// ReportTemplateFilter represents filters for querying templates
type ReportTemplateFilter struct {
	TenantID  string     `json:"tenant_id,omitempty"`
	Type      ReportType `json:"type,omitempty"`
	IsPublic  *bool      `json:"is_public,omitempty"`
	IsDefault *bool      `json:"is_default,omitempty"`
	CreatedBy string     `json:"created_by,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// ScheduledReportRequest represents a request to create/update a scheduled report
type ScheduledReportRequest struct {
	Name       string                 `json:"name" validate:"required,min=1,max=255"`
	TemplateID string                 `json:"template_id" validate:"required"`
	Schedule   ReportSchedule         `json:"schedule" validate:"required"`
	Filters    ReportFilters          `json:"filters,omitempty"`
	Recipients []ReportRecipient      `json:"recipients,omitempty"`
	IsActive   bool                   `json:"is_active,omitempty"`
	CreatedBy  string                 `json:"created_by" validate:"required"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ScheduledReportResponse represents a scheduled report response
type ScheduledReportResponse struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	TemplateID string         `json:"template_id"`
	Schedule   ReportSchedule `json:"schedule"`
	IsActive   bool           `json:"is_active"`
	LastRunAt  *time.Time     `json:"last_run_at"`
	NextRunAt  *time.Time     `json:"next_run_at"`
	CreatedBy  string         `json:"created_by"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// ScheduledReportListResponse represents a list of scheduled reports
type ScheduledReportListResponse struct {
	ScheduledReports []ScheduledReportResponse `json:"scheduled_reports"`
	Total            int                       `json:"total"`
	Page             int                       `json:"page"`
	PageSize         int                       `json:"page_size"`
}

// ScheduledReportFilter represents filters for querying scheduled reports
type ScheduledReportFilter struct {
	TenantID  string `json:"tenant_id,omitempty"`
	IsActive  *bool  `json:"is_active,omitempty"`
	CreatedBy string `json:"created_by,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// NewDefaultReportService creates a new default report service
func NewDefaultReportService(
	repository ReportRepository,
	templateRepo ReportTemplateRepository,
	schedulerRepo ScheduledReportRepository,
	dataProvider ReportDataProvider,
	generator ReportGenerator,
	scheduler ReportScheduler,
	logger *zap.Logger,
) *DefaultReportService {
	return &DefaultReportService{
		repository:    repository,
		templateRepo:  templateRepo,
		schedulerRepo: schedulerRepo,
		dataProvider:  dataProvider,
		generator:     generator,
		scheduler:     scheduler,
		logger:        logger,
	}
}

// GenerateReport generates a new report
func (rs *DefaultReportService) GenerateReport(ctx context.Context, request *ReportRequest) (*ReportResponse, error) {
	rs.logger.Info("Generating report",
		zap.String("name", request.Name),
		zap.String("type", string(request.Type)),
		zap.String("format", string(request.Format)),
		zap.String("created_by", request.CreatedBy))

	// Generate report ID
	reportID := generateReportID()

	// Create report
	report := &Report{
		ID:         reportID,
		TenantID:   getTenantIDFromContext(ctx),
		Name:       request.Name,
		Type:       request.Type,
		Status:     ReportStatusPending,
		Format:     request.Format,
		TemplateID: request.TemplateID,
		Filters:    request.Filters,
		CreatedBy:  request.CreatedBy,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Metadata:   request.Metadata,
	}

	// Set expiration time
	if request.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(request.ExpiresIn) * time.Hour)
		report.ExpiresAt = &expiresAt
	}

	// Save report
	if err := rs.repository.SaveReport(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	// Start report generation in background
	go rs.generateReportAsync(ctx, report)

	response := &ReportResponse{
		ID:        report.ID,
		Name:      report.Name,
		Type:      report.Type,
		Status:    report.Status,
		Format:    report.Format,
		CreatedBy: report.CreatedBy,
	}

	rs.logger.Info("Report generation started",
		zap.String("report_id", reportID),
		zap.String("name", request.Name))

	return response, nil
}

// generateReportAsync generates the report asynchronously
func (rs *DefaultReportService) generateReportAsync(ctx context.Context, report *Report) {
	rs.logger.Info("Starting async report generation",
		zap.String("report_id", report.ID),
		zap.String("type", string(report.Type)))

	// Update status to generating
	if err := rs.repository.UpdateReportStatus(ctx, report.TenantID, report.ID, ReportStatusGenerating, ""); err != nil {
		rs.logger.Error("Failed to update report status", zap.Error(err))
		return
	}

	// Get template if specified
	var template *ReportTemplate
	if report.TemplateID != "" {
		var err error
		template, err = rs.templateRepo.GetTemplate(ctx, report.TenantID, report.TemplateID)
		if err != nil {
			rs.logger.Error("Failed to get template", zap.Error(err))
			rs.repository.UpdateReportStatus(ctx, report.TenantID, report.ID, ReportStatusFailed, err.Error())
			return
		}
	} else {
		// Use default template for the report type
		template = rs.getDefaultTemplate(report.Type)
	}

	// Generate report data
	reportData, err := rs.generateReportData(ctx, report, template)
	if err != nil {
		rs.logger.Error("Failed to generate report data", zap.Error(err))
		rs.repository.UpdateReportStatus(ctx, report.TenantID, report.ID, ReportStatusFailed, err.Error())
		return
	}

	report.Data = *reportData

	// Generate report file
	var fileData []byte
	switch report.Format {
	case ReportFormatPDF:
		fileData, err = rs.generator.GeneratePDF(ctx, report, template)
	case ReportFormatExcel:
		fileData, err = rs.generator.GenerateExcel(ctx, report, template)
	case ReportFormatCSV:
		fileData, err = rs.generator.GenerateCSV(ctx, report, template)
	case ReportFormatJSON:
		fileData, err = rs.generator.GenerateJSON(ctx, report, template)
	case ReportFormatHTML:
		fileData, err = rs.generator.GenerateHTML(ctx, report, template)
	default:
		err = fmt.Errorf("unsupported report format: %s", report.Format)
	}

	if err != nil {
		rs.logger.Error("Failed to generate report file", zap.Error(err))
		rs.repository.UpdateReportStatus(ctx, report.TenantID, report.ID, ReportStatusFailed, err.Error())
		return
	}

	// Save report file and update status
	fileSize := int64(len(fileData))
	downloadURL := rs.generateDownloadURL(report.ID)

	if err := rs.repository.UpdateReportFile(ctx, report.TenantID, report.ID, fileSize, downloadURL); err != nil {
		rs.logger.Error("Failed to update report file", zap.Error(err))
		rs.repository.UpdateReportStatus(ctx, report.TenantID, report.ID, ReportStatusFailed, err.Error())
		return
	}

	// Update status to completed
	now := time.Now()
	report.Status = ReportStatusCompleted
	report.GeneratedAt = &now
	report.FileSize = fileSize
	report.DownloadURL = downloadURL

	if err := rs.repository.SaveReport(ctx, report); err != nil {
		rs.logger.Error("Failed to save completed report", zap.Error(err))
		return
	}

	rs.logger.Info("Report generation completed",
		zap.String("report_id", report.ID),
		zap.Int64("file_size", fileSize))
}

// generateReportData generates the data content for a report
func (rs *DefaultReportService) generateReportData(ctx context.Context, report *Report, template *ReportTemplate) (*ReportData, error) {
	rs.logger.Debug("Generating report data",
		zap.String("report_id", report.ID),
		zap.String("type", string(report.Type)))

	// Get data based on report type
	var data ReportData

	switch report.Type {
	case ReportTypeExecutiveSummary:
		data = rs.generateExecutiveSummaryData(ctx, report)
	case ReportTypeCompliance:
		data = rs.generateComplianceData(ctx, report)
	case ReportTypeRiskAudit:
		data = rs.generateRiskAuditData(ctx, report)
	case ReportTypeTrendAnalysis:
		data = rs.generateTrendAnalysisData(ctx, report)
	case ReportTypeBatchResults:
		data = rs.generateBatchResultsData(ctx, report)
	case ReportTypePerformance:
		data = rs.generatePerformanceData(ctx, report)
	case ReportTypeCustom:
		data = rs.generateCustomData(ctx, report, template)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", report.Type)
	}

	return &data, nil
}

// generateExecutiveSummaryData generates executive summary report data
func (rs *DefaultReportService) generateExecutiveSummaryData(ctx context.Context, report *Report) ReportData {
	// Get risk assessments
	assessments, _ := rs.dataProvider.GetRiskAssessments(ctx, report.TenantID, &report.Filters)

	// Calculate summary metrics
	totalAssessments := len(assessments)
	var totalRiskScore float64
	var highRiskCount, mediumRiskCount, lowRiskCount int

	for _, assessment := range assessments {
		totalRiskScore += assessment.RiskScore
		switch assessment.RiskLevel {
		case models.RiskLevelHigh, models.RiskLevelCritical:
			highRiskCount++
		case models.RiskLevelMedium:
			mediumRiskCount++
		case models.RiskLevelLow:
			lowRiskCount++
		}
	}

	avgRiskScore := float64(0)
	if totalAssessments > 0 {
		avgRiskScore = totalRiskScore / float64(totalAssessments)
	}

	summary := ReportSummary{
		Title:        "Executive Risk Assessment Summary",
		Description:  "High-level overview of risk assessment activities and key metrics",
		Period:       rs.getPeriodString(report.Filters.DateRange),
		TotalRecords: totalAssessments,
		KeyMetrics: map[string]interface{}{
			"average_risk_score": avgRiskScore,
			"high_risk_count":    highRiskCount,
			"medium_risk_count":  mediumRiskCount,
			"low_risk_count":     lowRiskCount,
		},
		ExecutiveSummary: fmt.Sprintf("During the reporting period, %d risk assessments were conducted with an average risk score of %.2f. %d assessments were classified as high risk, requiring immediate attention.", totalAssessments, avgRiskScore, highRiskCount),
		GeneratedAt:      time.Now(),
	}

	// Generate insights
	insights := []ReportInsight{
		{
			ID:    "insight_1",
			Type:  InsightTypeRiskTrend,
			Title: "Risk Distribution Analysis",
			Description: fmt.Sprintf("Risk distribution shows %d%% high risk, %d%% medium risk, and %d%% low risk assessments.",
				int(float64(highRiskCount)/float64(totalAssessments)*100),
				int(float64(mediumRiskCount)/float64(totalAssessments)*100),
				int(float64(lowRiskCount)/float64(totalAssessments)*100)),
			Impact:     InsightImpactMedium,
			Confidence: 0.85,
		},
	}

	// Generate recommendations
	recommendations := []ReportRecommendation{
		{
			ID:          "rec_1",
			Type:        RecommendationTypeRiskMitigation,
			Title:       "High Risk Assessment Review",
			Description: fmt.Sprintf("Review %d high-risk assessments to identify common risk factors and develop mitigation strategies.", highRiskCount),
			Priority:    RecommendationPriorityHigh,
			Action:      "Conduct detailed analysis of high-risk assessments",
			Timeline:    "Within 2 weeks",
			Resources:   []string{"Risk Analyst", "Compliance Team"},
		},
	}

	return ReportData{
		Summary:         summary,
		Charts:          []ReportChart{}, // Would be populated with actual chart data
		Tables:          []ReportTable{}, // Would be populated with actual table data
		Insights:        insights,
		Recommendations: recommendations,
		RawData:         assessments,
	}
}

// generateComplianceData generates compliance report data
func (rs *DefaultReportService) generateComplianceData(ctx context.Context, report *Report) ReportData {
	// Get compliance data
	complianceData, _ := rs.dataProvider.GetComplianceData(ctx, report.TenantID, &report.Filters)

	summary := ReportSummary{
		Title:        "Compliance Assessment Report",
		Description:  "Comprehensive compliance status and violation analysis",
		Period:       rs.getPeriodString(report.Filters.DateRange),
		TotalRecords: complianceData.TotalChecks,
		KeyMetrics: map[string]interface{}{
			"compliance_rate": float64(complianceData.Compliant) / float64(complianceData.TotalChecks) * 100,
			"violation_count": complianceData.NonCompliant,
			"pending_count":   complianceData.Pending,
		},
		ExecutiveSummary: fmt.Sprintf("Compliance assessment shows %d%% compliance rate with %d violations identified.",
			int(float64(complianceData.Compliant)/float64(complianceData.TotalChecks)*100),
			complianceData.NonCompliant),
		GeneratedAt: time.Now(),
	}

	return ReportData{
		Summary:         summary,
		Charts:          []ReportChart{},
		Tables:          []ReportTable{},
		Insights:        []ReportInsight{},
		Recommendations: []ReportRecommendation{},
		RawData:         complianceData,
	}
}

// generateRiskAuditData generates risk audit report data
func (rs *DefaultReportService) generateRiskAuditData(ctx context.Context, report *Report) ReportData {
	// Implementation for risk audit data generation
	summary := ReportSummary{
		Title:            "Risk Audit Report",
		Description:      "Detailed risk audit findings and recommendations",
		Period:           rs.getPeriodString(report.Filters.DateRange),
		TotalRecords:     0,
		KeyMetrics:       map[string]interface{}{},
		ExecutiveSummary: "Risk audit completed with detailed findings and recommendations.",
		GeneratedAt:      time.Now(),
	}

	return ReportData{
		Summary:         summary,
		Charts:          []ReportChart{},
		Tables:          []ReportTable{},
		Insights:        []ReportInsight{},
		Recommendations: []ReportRecommendation{},
		RawData:         nil,
	}
}

// generateTrendAnalysisData generates trend analysis report data
func (rs *DefaultReportService) generateTrendAnalysisData(ctx context.Context, report *Report) ReportData {
	// Implementation for trend analysis data generation
	summary := ReportSummary{
		Title:            "Risk Trend Analysis Report",
		Description:      "Historical trend analysis and future predictions",
		Period:           rs.getPeriodString(report.Filters.DateRange),
		TotalRecords:     0,
		KeyMetrics:       map[string]interface{}{},
		ExecutiveSummary: "Trend analysis shows risk patterns and future projections.",
		GeneratedAt:      time.Now(),
	}

	return ReportData{
		Summary:         summary,
		Charts:          []ReportChart{},
		Tables:          []ReportTable{},
		Insights:        []ReportInsight{},
		Recommendations: []ReportRecommendation{},
		RawData:         nil,
	}
}

// generateBatchResultsData generates batch results report data
func (rs *DefaultReportService) generateBatchResultsData(ctx context.Context, report *Report) ReportData {
	// Get batch job data
	batchJobs, _ := rs.dataProvider.GetBatchJobs(ctx, report.TenantID, &report.Filters)

	summary := ReportSummary{
		Title:            "Batch Processing Results Report",
		Description:      "Summary of batch job processing results and performance",
		Period:           rs.getPeriodString(report.Filters.DateRange),
		TotalRecords:     len(batchJobs),
		KeyMetrics:       map[string]interface{}{},
		ExecutiveSummary: fmt.Sprintf("Processed %d batch jobs during the reporting period.", len(batchJobs)),
		GeneratedAt:      time.Now(),
	}

	return ReportData{
		Summary:         summary,
		Charts:          []ReportChart{},
		Tables:          []ReportTable{},
		Insights:        []ReportInsight{},
		Recommendations: []ReportRecommendation{},
		RawData:         batchJobs,
	}
}

// generatePerformanceData generates performance report data
func (rs *DefaultReportService) generatePerformanceData(ctx context.Context, report *Report) ReportData {
	// Get performance data
	performanceData, _ := rs.dataProvider.GetPerformanceData(ctx, report.TenantID, &report.Filters)

	summary := ReportSummary{
		Title:        "Performance Metrics Report",
		Description:  "System performance metrics and optimization recommendations",
		Period:       rs.getPeriodString(report.Filters.DateRange),
		TotalRecords: 0,
		KeyMetrics: map[string]interface{}{
			"average_response_time": performanceData.ResponseTime.Average,
			"error_rate":            performanceData.ErrorRate.Average,
			"availability":          performanceData.Availability.Average,
		},
		ExecutiveSummary: fmt.Sprintf("System performance shows %.2fms average response time with %.2f%% availability.",
			performanceData.ResponseTime.Average, performanceData.Availability.Average),
		GeneratedAt: time.Now(),
	}

	return ReportData{
		Summary:         summary,
		Charts:          []ReportChart{},
		Tables:          []ReportTable{},
		Insights:        []ReportInsight{},
		Recommendations: []ReportRecommendation{},
		RawData:         performanceData,
	}
}

// generateCustomData generates custom report data
func (rs *DefaultReportService) generateCustomData(ctx context.Context, report *Report, template *ReportTemplate) ReportData {
	// Implementation for custom data generation based on template
	summary := ReportSummary{
		Title:            "Custom Report",
		Description:      "Custom report generated based on template configuration",
		Period:           rs.getPeriodString(report.Filters.DateRange),
		TotalRecords:     0,
		KeyMetrics:       map[string]interface{}{},
		ExecutiveSummary: "Custom report generated successfully.",
		GeneratedAt:      time.Now(),
	}

	return ReportData{
		Summary:         summary,
		Charts:          []ReportChart{},
		Tables:          []ReportTable{},
		Insights:        []ReportInsight{},
		Recommendations: []ReportRecommendation{},
		RawData:         nil,
	}
}

// Helper methods

func (rs *DefaultReportService) getDefaultTemplate(reportType ReportType) *ReportTemplate {
	// Return default template for the report type
	return &ReportTemplate{
		ID:   "default_" + string(reportType),
		Type: reportType,
		Template: ReportTemplateConfig{
			Layout: ReportLayout{
				PageSize:    "A4",
				Orientation: "portrait",
				Margins: ReportMargins{
					Top:    1.0,
					Bottom: 1.0,
					Left:   1.0,
					Right:  1.0,
				},
			},
			Sections: []ReportSection{},
			Charts:   []ReportChartTemplate{},
			Tables:   []ReportTableTemplate{},
		},
	}
}

func (rs *DefaultReportService) getPeriodString(dateRange DateRangeFilter) string {
	if dateRange.StartDate != nil && dateRange.EndDate != nil {
		return fmt.Sprintf("%s to %s",
			dateRange.StartDate.Format("2006-01-02"),
			dateRange.EndDate.Format("2006-01-02"))
	}
	return "All time"
}

func (rs *DefaultReportService) generateDownloadURL(reportID string) string {
	return fmt.Sprintf("/api/v1/reports/%s/download", reportID)
}

// Template Management Methods

func (rs *DefaultReportService) CreateTemplate(ctx context.Context, request *ReportTemplateRequest) (*ReportTemplateResponse, error) {
	rs.logger.Info("Creating report template",
		zap.String("name", request.Name),
		zap.String("type", string(request.Type)),
		zap.String("created_by", request.CreatedBy))

	// Generate template ID
	templateID := generateTemplateID()

	// Create template
	template := &ReportTemplate{
		ID:          templateID,
		TenantID:    getTenantIDFromContext(ctx),
		Name:        request.Name,
		Type:        request.Type,
		Description: request.Description,
		Template:    request.Template,
		IsPublic:    request.IsPublic,
		IsDefault:   false,
		CreatedBy:   request.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    request.Metadata,
	}

	// Save template
	if err := rs.templateRepo.SaveTemplate(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to save template: %w", err)
	}

	response := &ReportTemplateResponse{
		ID:          template.ID,
		Name:        template.Name,
		Type:        template.Type,
		Description: template.Description,
		IsPublic:    template.IsPublic,
		IsDefault:   template.IsDefault,
		CreatedBy:   template.CreatedBy,
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}

	rs.logger.Info("Template created successfully",
		zap.String("template_id", templateID),
		zap.String("name", request.Name))

	return response, nil
}

func (rs *DefaultReportService) GetTemplate(ctx context.Context, tenantID, templateID string) (*ReportTemplate, error) {
	rs.logger.Debug("Getting template",
		zap.String("template_id", templateID),
		zap.String("tenant_id", tenantID))

	template, err := rs.templateRepo.GetTemplate(ctx, tenantID, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if template == nil {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	rs.logger.Debug("Template retrieved successfully",
		zap.String("template_id", templateID))

	return template, nil
}

func (rs *DefaultReportService) ListTemplates(ctx context.Context, filter *ReportTemplateFilter) (*ReportTemplateListResponse, error) {
	rs.logger.Debug("Listing templates",
		zap.String("tenant_id", filter.TenantID),
		zap.String("type", string(filter.Type)))

	templates, err := rs.templateRepo.ListTemplates(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	// Convert to response format
	responses := make([]ReportTemplateResponse, len(templates))
	for i, template := range templates {
		responses[i] = ReportTemplateResponse{
			ID:          template.ID,
			Name:        template.Name,
			Type:        template.Type,
			Description: template.Description,
			IsPublic:    template.IsPublic,
			IsDefault:   template.IsDefault,
			CreatedBy:   template.CreatedBy,
			CreatedAt:   template.CreatedAt,
			UpdatedAt:   template.UpdatedAt,
		}
	}

	response := &ReportTemplateListResponse{
		Templates: responses,
		Total:     len(responses),
		Page:      1, // This would be calculated based on offset/limit
		PageSize:  len(responses),
	}

	rs.logger.Debug("Templates listed successfully",
		zap.String("tenant_id", filter.TenantID),
		zap.Int("count", len(responses)))

	return response, nil
}

func (rs *DefaultReportService) UpdateTemplate(ctx context.Context, tenantID, templateID string, request *ReportTemplateRequest) (*ReportTemplateResponse, error) {
	rs.logger.Info("Updating template",
		zap.String("template_id", templateID),
		zap.String("tenant_id", tenantID))

	// Get existing template
	template, err := rs.templateRepo.GetTemplate(ctx, tenantID, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if template == nil {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	// Update template fields
	template.Name = request.Name
	template.Type = request.Type
	template.Description = request.Description
	template.Template = request.Template
	template.IsPublic = request.IsPublic
	template.Metadata = request.Metadata
	template.UpdatedAt = time.Now()

	// Save updated template
	if err := rs.templateRepo.UpdateTemplate(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	response := &ReportTemplateResponse{
		ID:          template.ID,
		Name:        template.Name,
		Type:        template.Type,
		Description: template.Description,
		IsPublic:    template.IsPublic,
		IsDefault:   template.IsDefault,
		CreatedBy:   template.CreatedBy,
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}

	rs.logger.Info("Template updated successfully",
		zap.String("template_id", templateID))

	return response, nil
}

func (rs *DefaultReportService) DeleteTemplate(ctx context.Context, tenantID, templateID string) error {
	rs.logger.Info("Deleting template",
		zap.String("template_id", templateID),
		zap.String("tenant_id", tenantID))

	if err := rs.templateRepo.DeleteTemplate(ctx, tenantID, templateID); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rs.logger.Info("Template deleted successfully",
		zap.String("template_id", templateID))

	return nil
}

// Scheduled Report Methods

func (rs *DefaultReportService) CreateScheduledReport(ctx context.Context, request *ScheduledReportRequest) (*ScheduledReportResponse, error) {
	rs.logger.Info("Creating scheduled report",
		zap.String("name", request.Name),
		zap.String("template_id", request.TemplateID),
		zap.String("created_by", request.CreatedBy))

	// Generate scheduled report ID
	scheduledReportID := generateScheduledReportID()

	// Create scheduled report
	scheduledReport := &ScheduledReport{
		ID:         scheduledReportID,
		TenantID:   getTenantIDFromContext(ctx),
		Name:       request.Name,
		TemplateID: request.TemplateID,
		Schedule:   request.Schedule,
		Filters:    request.Filters,
		Recipients: request.Recipients,
		IsActive:   request.IsActive,
		CreatedBy:  request.CreatedBy,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Metadata:   request.Metadata,
	}

	// Calculate next run time
	nextRunAt := rs.calculateNextRunTime(request.Schedule)
	scheduledReport.NextRunAt = &nextRunAt

	// Save scheduled report
	if err := rs.schedulerRepo.SaveScheduledReport(ctx, scheduledReport); err != nil {
		return nil, fmt.Errorf("failed to save scheduled report: %w", err)
	}

	// Schedule the report
	if request.IsActive {
		if err := rs.scheduler.ScheduleReport(ctx, scheduledReport); err != nil {
			rs.logger.Error("Failed to schedule report", zap.Error(err))
			// Don't fail the creation, just log the error
		}
	}

	response := &ScheduledReportResponse{
		ID:         scheduledReport.ID,
		Name:       scheduledReport.Name,
		TemplateID: scheduledReport.TemplateID,
		Schedule:   scheduledReport.Schedule,
		IsActive:   scheduledReport.IsActive,
		LastRunAt:  scheduledReport.LastRunAt,
		NextRunAt:  scheduledReport.NextRunAt,
		CreatedBy:  scheduledReport.CreatedBy,
		CreatedAt:  scheduledReport.CreatedAt,
		UpdatedAt:  scheduledReport.UpdatedAt,
	}

	rs.logger.Info("Scheduled report created successfully",
		zap.String("scheduled_report_id", scheduledReportID),
		zap.String("name", request.Name))

	return response, nil
}

func (rs *DefaultReportService) GetScheduledReport(ctx context.Context, tenantID, scheduledReportID string) (*ScheduledReport, error) {
	rs.logger.Debug("Getting scheduled report",
		zap.String("scheduled_report_id", scheduledReportID),
		zap.String("tenant_id", tenantID))

	scheduledReport, err := rs.schedulerRepo.GetScheduledReport(ctx, tenantID, scheduledReportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled report: %w", err)
	}

	if scheduledReport == nil {
		return nil, fmt.Errorf("scheduled report not found: %s", scheduledReportID)
	}

	rs.logger.Debug("Scheduled report retrieved successfully",
		zap.String("scheduled_report_id", scheduledReportID))

	return scheduledReport, nil
}

func (rs *DefaultReportService) ListScheduledReports(ctx context.Context, filter *ScheduledReportFilter) (*ScheduledReportListResponse, error) {
	rs.logger.Debug("Listing scheduled reports",
		zap.String("tenant_id", filter.TenantID))

	scheduledReports, err := rs.schedulerRepo.ListScheduledReports(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list scheduled reports: %w", err)
	}

	// Convert to response format
	responses := make([]ScheduledReportResponse, len(scheduledReports))
	for i, scheduledReport := range scheduledReports {
		responses[i] = ScheduledReportResponse{
			ID:         scheduledReport.ID,
			Name:       scheduledReport.Name,
			TemplateID: scheduledReport.TemplateID,
			Schedule:   scheduledReport.Schedule,
			IsActive:   scheduledReport.IsActive,
			LastRunAt:  scheduledReport.LastRunAt,
			NextRunAt:  scheduledReport.NextRunAt,
			CreatedBy:  scheduledReport.CreatedBy,
			CreatedAt:  scheduledReport.CreatedAt,
			UpdatedAt:  scheduledReport.UpdatedAt,
		}
	}

	response := &ScheduledReportListResponse{
		ScheduledReports: responses,
		Total:            len(responses),
		Page:             1, // This would be calculated based on offset/limit
		PageSize:         len(responses),
	}

	rs.logger.Debug("Scheduled reports listed successfully",
		zap.String("tenant_id", filter.TenantID),
		zap.Int("count", len(responses)))

	return response, nil
}

func (rs *DefaultReportService) UpdateScheduledReport(ctx context.Context, tenantID, scheduledReportID string, request *ScheduledReportRequest) (*ScheduledReportResponse, error) {
	rs.logger.Info("Updating scheduled report",
		zap.String("scheduled_report_id", scheduledReportID),
		zap.String("tenant_id", tenantID))

	// Get existing scheduled report
	scheduledReport, err := rs.schedulerRepo.GetScheduledReport(ctx, tenantID, scheduledReportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled report: %w", err)
	}

	if scheduledReport == nil {
		return nil, fmt.Errorf("scheduled report not found: %s", scheduledReportID)
	}

	// Update scheduled report fields
	scheduledReport.Name = request.Name
	scheduledReport.TemplateID = request.TemplateID
	scheduledReport.Schedule = request.Schedule
	scheduledReport.Filters = request.Filters
	scheduledReport.Recipients = request.Recipients
	scheduledReport.IsActive = request.IsActive
	scheduledReport.Metadata = request.Metadata
	scheduledReport.UpdatedAt = time.Now()

	// Calculate next run time
	nextRunAt := rs.calculateNextRunTime(request.Schedule)
	scheduledReport.NextRunAt = &nextRunAt

	// Save updated scheduled report
	if err := rs.schedulerRepo.UpdateScheduledReport(ctx, scheduledReport); err != nil {
		return nil, fmt.Errorf("failed to update scheduled report: %w", err)
	}

	// Update scheduler
	if request.IsActive {
		if err := rs.scheduler.ScheduleReport(ctx, scheduledReport); err != nil {
			rs.logger.Error("Failed to reschedule report", zap.Error(err))
		}
	} else {
		if err := rs.scheduler.UnscheduleReport(ctx, scheduledReportID); err != nil {
			rs.logger.Error("Failed to unschedule report", zap.Error(err))
		}
	}

	response := &ScheduledReportResponse{
		ID:         scheduledReport.ID,
		Name:       scheduledReport.Name,
		TemplateID: scheduledReport.TemplateID,
		Schedule:   scheduledReport.Schedule,
		IsActive:   scheduledReport.IsActive,
		LastRunAt:  scheduledReport.LastRunAt,
		NextRunAt:  scheduledReport.NextRunAt,
		CreatedBy:  scheduledReport.CreatedBy,
		CreatedAt:  scheduledReport.CreatedAt,
		UpdatedAt:  scheduledReport.UpdatedAt,
	}

	rs.logger.Info("Scheduled report updated successfully",
		zap.String("scheduled_report_id", scheduledReportID))

	return response, nil
}

func (rs *DefaultReportService) DeleteScheduledReport(ctx context.Context, tenantID, scheduledReportID string) error {
	rs.logger.Info("Deleting scheduled report",
		zap.String("scheduled_report_id", scheduledReportID),
		zap.String("tenant_id", tenantID))

	// Unschedule the report
	if err := rs.scheduler.UnscheduleReport(ctx, scheduledReportID); err != nil {
		rs.logger.Error("Failed to unschedule report", zap.Error(err))
	}

	// Delete the scheduled report
	if err := rs.schedulerRepo.DeleteScheduledReport(ctx, tenantID, scheduledReportID); err != nil {
		return fmt.Errorf("failed to delete scheduled report: %w", err)
	}

	rs.logger.Info("Scheduled report deleted successfully",
		zap.String("scheduled_report_id", scheduledReportID))

	return nil
}

func (rs *DefaultReportService) RunScheduledReport(ctx context.Context, tenantID, scheduledReportID string) (*ReportResponse, error) {
	rs.logger.Info("Running scheduled report",
		zap.String("scheduled_report_id", scheduledReportID),
		zap.String("tenant_id", tenantID))

	// Get scheduled report
	scheduledReport, err := rs.schedulerRepo.GetScheduledReport(ctx, tenantID, scheduledReportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled report: %w", err)
	}

	if scheduledReport == nil {
		return nil, fmt.Errorf("scheduled report not found: %s", scheduledReportID)
	}

	// Create report request
	reportRequest := &ReportRequest{
		Name:       scheduledReport.Name,
		Type:       ReportTypeCustom, // This would be determined from template
		TemplateID: scheduledReport.TemplateID,
		Format:     ReportFormatPDF, // Default format
		Filters:    scheduledReport.Filters,
		Recipients: scheduledReport.Recipients,
		CreatedBy:  scheduledReport.CreatedBy,
		Metadata:   scheduledReport.Metadata,
	}

	// Generate report
	reportResponse, err := rs.GenerateReport(ctx, reportRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	// Update last run time
	now := time.Now()
	nextRunAt := rs.calculateNextRunTime(scheduledReport.Schedule)
	if err := rs.schedulerRepo.UpdateScheduledReportLastRun(ctx, tenantID, scheduledReportID, now, &nextRunAt); err != nil {
		rs.logger.Error("Failed to update scheduled report last run", zap.Error(err))
	}

	rs.logger.Info("Scheduled report executed successfully",
		zap.String("scheduled_report_id", scheduledReportID),
		zap.String("report_id", reportResponse.ID))

	return reportResponse, nil
}

// Helper methods for scheduled reports

func (rs *DefaultReportService) calculateNextRunTime(schedule ReportSchedule) time.Time {
	now := time.Now()

	switch schedule.Frequency {
	case ScheduleFrequencyOnce:
		if schedule.StartDate != nil {
			return *schedule.StartDate
		}
		return now.Add(24 * time.Hour) // Default to tomorrow

	case ScheduleFrequencyDaily:
		nextRun := now.Add(24 * time.Hour)
		if schedule.TimeOfDay != "" {
			// Parse time and set for next day
			// This is a simplified implementation
			nextRun = nextRun.Truncate(24 * time.Hour).Add(9 * time.Hour) // Default to 9 AM
		}
		return nextRun

	case ScheduleFrequencyWeekly:
		nextRun := now.Add(7 * 24 * time.Hour)
		// This would be more complex to handle specific days of week
		return nextRun

	case ScheduleFrequencyMonthly:
		nextRun := now.AddDate(0, 1, 0)
		// This would be more complex to handle specific days of month
		return nextRun

	default:
		return now.Add(24 * time.Hour)
	}
}

// Other methods (GetReport, ListReports, DeleteReport, GetReportMetrics)
// These would follow similar patterns to the template methods

func (rs *DefaultReportService) GetReport(ctx context.Context, tenantID, reportID string) (*Report, error) {
	rs.logger.Debug("Getting report",
		zap.String("report_id", reportID),
		zap.String("tenant_id", tenantID))

	report, err := rs.repository.GetReport(ctx, tenantID, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	if report == nil {
		return nil, fmt.Errorf("report not found: %s", reportID)
	}

	rs.logger.Debug("Report retrieved successfully",
		zap.String("report_id", reportID))

	return report, nil
}

func (rs *DefaultReportService) ListReports(ctx context.Context, filter *ReportFilter) (*ReportListResponse, error) {
	rs.logger.Debug("Listing reports",
		zap.String("tenant_id", filter.TenantID),
		zap.String("type", string(filter.Type)))

	reports, err := rs.repository.ListReports(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}

	// Convert to response format
	responses := make([]ReportResponse, len(reports))
	for i, report := range reports {
		responses[i] = ReportResponse{
			ID:          report.ID,
			Name:        report.Name,
			Type:        report.Type,
			Status:      report.Status,
			Format:      report.Format,
			GeneratedAt: report.GeneratedAt,
			ExpiresAt:   report.ExpiresAt,
			FileSize:    report.FileSize,
			DownloadURL: report.DownloadURL,
			CreatedBy:   report.CreatedBy,
		}
	}

	response := &ReportListResponse{
		Reports:  responses,
		Total:    len(responses),
		Page:     1, // This would be calculated based on offset/limit
		PageSize: len(responses),
	}

	rs.logger.Debug("Reports listed successfully",
		zap.String("tenant_id", filter.TenantID),
		zap.Int("count", len(responses)))

	return response, nil
}

func (rs *DefaultReportService) DeleteReport(ctx context.Context, tenantID, reportID string) error {
	rs.logger.Info("Deleting report",
		zap.String("report_id", reportID),
		zap.String("tenant_id", tenantID))

	if err := rs.repository.DeleteReport(ctx, tenantID, reportID); err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	rs.logger.Info("Report deleted successfully",
		zap.String("report_id", reportID))

	return nil
}

func (rs *DefaultReportService) GetReportMetrics(ctx context.Context, tenantID string) (*ReportMetrics, error) {
	rs.logger.Debug("Getting report metrics",
		zap.String("tenant_id", tenantID))

	metrics, err := rs.repository.GetReportMetrics(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report metrics: %w", err)
	}

	rs.logger.Debug("Report metrics retrieved",
		zap.String("tenant_id", tenantID),
		zap.Int("total_reports", metrics.TotalReports))

	return metrics, nil
}

// Helper functions

func generateReportID() string {
	return fmt.Sprintf("report_%d", time.Now().UnixNano())
}

func generateTemplateID() string {
	return fmt.Sprintf("template_%d", time.Now().UnixNano())
}

func generateScheduledReportID() string {
	return fmt.Sprintf("scheduled_report_%d", time.Now().UnixNano())
}
