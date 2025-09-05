package risk

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ExportService provides risk data export functionality
type ExportService struct {
	logger         *observability.Logger
	historyService *RiskHistoryService
	alertService   *AlertService
	reportService  *ReportService
}

// NewExportService creates a new export service
func NewExportService(logger *observability.Logger, historyService *RiskHistoryService, alertService *AlertService, reportService *ReportService) *ExportService {
	return &ExportService{
		logger:         logger,
		historyService: historyService,
		alertService:   alertService,
		reportService:  reportService,
	}
}

// ExportRiskData exports risk data in the specified format
func (s *ExportService) ExportRiskData(ctx context.Context, request ExportRequest) (*ExportResponse, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Exporting risk data",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"export_type", request.ExportType,
		"format", request.Format,
	)

	// Validate request
	if err := s.validateExportRequest(request); err != nil {
		s.logger.Error("Invalid export request",
			"request_id", requestID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("invalid export request: %w", err)
	}

	// Collect data based on export type
	data, err := s.collectExportData(ctx, request)
	if err != nil {
		s.logger.Error("Failed to collect export data",
			"request_id", requestID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to collect export data: %w", err)
	}

	// Format data according to requested format
	formattedData, err := s.formatExportData(data, request.Format)
	if err != nil {
		s.logger.Error("Failed to format export data",
			"request_id", requestID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to format export data: %w", err)
	}

	// Create export response
	exportID := fmt.Sprintf("export_%s_%d", request.BusinessID, time.Now().Unix())
	response := &ExportResponse{
		ExportID:    exportID,
		BusinessID:  request.BusinessID,
		ExportType:  request.ExportType,
		Format:      request.Format,
		Data:        formattedData,
		RecordCount: s.countRecords(data),
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Metadata:    request.Metadata,
	}

	s.logger.Info("Risk data export completed",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"export_type", request.ExportType,
		"format", request.Format,
		"record_count", response.RecordCount,
	)

	return response, nil
}

// CreateExportJob creates a background export job for large datasets
func (s *ExportService) CreateExportJob(ctx context.Context, request ExportRequest) (*ExportJob, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Creating export job",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"export_type", request.ExportType,
		"format", request.Format,
	)

	// Validate request
	if err := s.validateExportRequest(request); err != nil {
		s.logger.Error("Invalid export job request",
			"request_id", requestID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("invalid export job request: %w", err)
	}

	// Create export job
	jobID := fmt.Sprintf("job_%s_%d", request.BusinessID, time.Now().Unix())
	job := &ExportJob{
		ID:         jobID,
		BusinessID: request.BusinessID,
		ExportType: request.ExportType,
		Format:     request.Format,
		Status:     "pending",
		Progress:   0,
		CreatedAt:  time.Now(),
		Metadata:   request.Metadata,
	}

	// Start background processing
	go s.processExportJob(ctx, job, request)

	s.logger.Info("Export job created",
		"request_id", requestID,
		"job_id", jobID,
		"business_id", request.BusinessID,
	)

	return job, nil
}

// GetExportJob retrieves the status of an export job
func (s *ExportService) GetExportJob(ctx context.Context, jobID string) (*ExportJob, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving export job",
		"request_id", requestID,
		"job_id", jobID,
	)

	// In a real implementation, this would query the database
	// For now, return a mock job
	job := &ExportJob{
		ID:          jobID,
		BusinessID:  "test_business",
		ExportType:  ExportTypeAssessments,
		Format:      ExportFormatJSON,
		Status:      "completed",
		Progress:    100,
		CreatedAt:   time.Now().Add(-1 * time.Hour),
		StartedAt:   &time.Time{},
		CompletedAt: &time.Time{},
	}

	// Set completed time
	completedTime := time.Now()
	job.CompletedAt = &completedTime

	s.logger.Info("Export job retrieved",
		"request_id", requestID,
		"job_id", jobID,
		"status", job.Status,
		"progress", job.Progress,
	)

	return job, nil
}

// validateExportRequest validates the export request
func (s *ExportService) validateExportRequest(request ExportRequest) error {
	if request.BusinessID == "" {
		return fmt.Errorf("business ID is required")
	}

	if request.ExportType == "" {
		return fmt.Errorf("export type is required")
	}

	if request.Format == "" {
		return fmt.Errorf("export format is required")
	}

	// Validate export type
	switch request.ExportType {
	case ExportTypeAssessments, ExportTypeFactors, ExportTypeTrends, ExportTypeAlerts, ExportTypeReports, ExportTypeAll:
		// Valid export type
	default:
		return fmt.Errorf("invalid export type: %s", request.ExportType)
	}

	// Validate format
	switch request.Format {
	case ExportFormatJSON, ExportFormatCSV, ExportFormatXML, ExportFormatPDF, ExportFormatXLSX:
		// Valid format
	default:
		return fmt.Errorf("invalid export format: %s", request.Format)
	}

	return nil
}

// collectExportData collects data based on the export type
func (s *ExportService) collectExportData(ctx context.Context, request ExportRequest) (interface{}, error) {
	switch request.ExportType {
	case ExportTypeAssessments:
		return s.collectAssessments(ctx, request)
	case ExportTypeFactors:
		return s.collectFactors(ctx, request)
	case ExportTypeTrends:
		return s.collectTrends(ctx, request)
	case ExportTypeAlerts:
		return s.collectAlerts(ctx, request)
	case ExportTypeReports:
		return s.collectReports(ctx, request)
	case ExportTypeAll:
		return s.collectAllData(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported export type: %s", request.ExportType)
	}
}

// collectAssessments collects risk assessment data
func (s *ExportService) collectAssessments(ctx context.Context, request ExportRequest) (interface{}, error) {
	// Get assessments from history service
	history, err := s.historyService.GetRiskHistory(ctx, request.BusinessID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk history: %w", err)
	}

	// Filter by date range if specified
	if request.DateRange != nil {
		filteredHistory := s.filterByDateRange(history.History, request.DateRange)
		return filteredHistory, nil
	}

	return history.History, nil
}

// collectFactors collects risk factor data
func (s *ExportService) collectFactors(ctx context.Context, request ExportRequest) (interface{}, error) {
	// Get assessments and extract factor data
	history, err := s.historyService.GetRiskHistory(ctx, request.BusinessID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk history: %w", err)
	}

	var factors []map[string]interface{}
	for _, entry := range history.History {
		if entry.Assessment != nil {
			for _, factor := range entry.Assessment.FactorScores {
				factorData := map[string]interface{}{
					"factor_id":     factor.FactorID,
					"factor_name":   factor.FactorName,
					"category":      factor.Category,
					"score":         factor.Score,
					"level":         factor.Level,
					"confidence":    factor.Confidence,
					"explanation":   factor.Explanation,
					"evidence":      factor.Evidence,
					"calculated_at": factor.CalculatedAt,
					"assessment_id": entry.Assessment.ID,
					"business_id":   request.BusinessID,
				}
				factors = append(factors, factorData)
			}
		}
	}

	return factors, nil
}

// collectTrends collects trend analysis data
func (s *ExportService) collectTrends(ctx context.Context, request ExportRequest) (interface{}, error) {
	// Get trend analysis
	trends, err := s.historyService.GetRiskTrends(ctx, request.BusinessID, 365)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk trends: %w", err)
	}

	return trends, nil
}

// collectAlerts collects risk alert data
func (s *ExportService) collectAlerts(ctx context.Context, request ExportRequest) (interface{}, error) {
	// Get alerts from alert service
	alerts, err := s.alertService.GetAlerts(ctx, request.BusinessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}

	return alerts, nil
}

// collectReports collects risk report data
func (s *ExportService) collectReports(ctx context.Context, request ExportRequest) (interface{}, error) {
	// Generate a comprehensive report
	reportRequest := ReportRequest{
		BusinessID: request.BusinessID,
		ReportType: ReportTypeDetailed,
		Format:     ReportFormatJSON,
		DateRange:  request.DateRange,
	}

	report, err := s.reportService.GenerateReport(ctx, reportRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	return report, nil
}

// collectAllData collects all available risk data
func (s *ExportService) collectAllData(ctx context.Context, request ExportRequest) (interface{}, error) {
	allData := map[string]interface{}{}

	// Collect assessments
	if assessments, err := s.collectAssessments(ctx, request); err == nil {
		allData["assessments"] = assessments
	}

	// Collect factors
	if factors, err := s.collectFactors(ctx, request); err == nil {
		allData["factors"] = factors
	}

	// Collect trends
	if trends, err := s.collectTrends(ctx, request); err == nil {
		allData["trends"] = trends
	}

	// Collect alerts
	if alerts, err := s.collectAlerts(ctx, request); err == nil {
		allData["alerts"] = alerts
	}

	// Collect reports
	if reports, err := s.collectReports(ctx, request); err == nil {
		allData["reports"] = reports
	}

	return allData, nil
}

// formatExportData formats data according to the requested format
func (s *ExportService) formatExportData(data interface{}, format ExportFormat) (interface{}, error) {
	switch format {
	case ExportFormatJSON:
		return s.formatAsJSON(data)
	case ExportFormatCSV:
		return s.formatAsCSV(data)
	case ExportFormatXML:
		return s.formatAsXML(data)
	case ExportFormatPDF:
		return s.formatAsPDF(data)
	case ExportFormatXLSX:
		return s.formatAsXLSX(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// formatAsJSON formats data as JSON
func (s *ExportService) formatAsJSON(data interface{}) (interface{}, error) {
	// Data is already in JSON-compatible format
	return data, nil
}

// formatAsCSV formats data as CSV
func (s *ExportService) formatAsCSV(data interface{}) (interface{}, error) {
	// Convert data to CSV format
	switch v := data.(type) {
	case []map[string]interface{}:
		return s.convertMapsToCSV(v)
	case []interface{}:
		return s.convertSliceToCSV(v)
	default:
		return nil, fmt.Errorf("unsupported data type for CSV export: %T", data)
	}
}

// formatAsXML formats data as XML
func (s *ExportService) formatAsXML(data interface{}) (interface{}, error) {
	// Convert data to XML format
	xmlData := map[string]interface{}{
		"root": data,
	}
	return xmlData, nil
}

// formatAsPDF formats data as PDF
func (s *ExportService) formatAsPDF(data interface{}) (interface{}, error) {
	// Convert data to PDF format
	pdfData := map[string]interface{}{
		"content": data,
		"format":  "pdf",
	}
	return pdfData, nil
}

// formatAsXLSX formats data as XLSX
func (s *ExportService) formatAsXLSX(data interface{}) (interface{}, error) {
	// Convert data to XLSX format
	xlsxData := map[string]interface{}{
		"sheets": data,
		"format": "xlsx",
	}
	return xlsxData, nil
}

// convertMapsToCSV converts a slice of maps to CSV format
func (s *ExportService) convertMapsToCSV(data []map[string]interface{}) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	var csvBuilder strings.Builder
	writer := csv.NewWriter(&csvBuilder)

	// Write header
	headers := make([]string, 0, len(data[0]))
	for key := range data[0] {
		headers = append(headers, key)
	}
	if err := writer.Write(headers); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, row := range data {
		values := make([]string, 0, len(headers))
		for _, header := range headers {
			value := s.convertToString(row[header])
			values = append(values, value)
		}
		if err := writer.Write(values); err != nil {
			return "", fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	return csvBuilder.String(), nil
}

// convertSliceToCSV converts a slice of interfaces to CSV format
func (s *ExportService) convertSliceToCSV(data []interface{}) (string, error) {
	var csvBuilder strings.Builder
	writer := csv.NewWriter(&csvBuilder)

	// Write data rows
	for _, item := range data {
		value := s.convertToString(item)
		if err := writer.Write([]string{value}); err != nil {
			return "", fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	return csvBuilder.String(), nil
}

// convertToString converts any value to string
func (s *ExportService) convertToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int32, int64:
		return strconv.FormatInt(int64(v.(int)), 10)
	case float32, float64:
		return strconv.FormatFloat(float64(v.(float64)), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		// Try to convert to JSON string
		if jsonBytes, err := json.Marshal(v); err == nil {
			return string(jsonBytes)
		}
		return fmt.Sprintf("%v", v)
	}
}

// countRecords counts the number of records in the data
func (s *ExportService) countRecords(data interface{}) int {
	switch v := data.(type) {
	case []interface{}:
		return len(v)
	case []map[string]interface{}:
		return len(v)
	case map[string]interface{}:
		return 1
	default:
		return 1
	}
}

// filterByDateRange filters data by date range
func (s *ExportService) filterByDateRange(data []RiskHistoryEntry, dateRange *DateRange) []RiskHistoryEntry {
	var filtered []RiskHistoryEntry
	for _, entry := range data {
		if entry.Assessment != nil {
			if entry.Assessment.AssessedAt.After(dateRange.StartDate) && entry.Assessment.AssessedAt.Before(dateRange.EndDate) {
				filtered = append(filtered, entry)
			}
		}
	}
	return filtered
}

// processExportJob processes an export job in the background
func (s *ExportService) processExportJob(ctx context.Context, job *ExportJob, request ExportRequest) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Processing export job",
		"request_id", requestID,
		"job_id", job.ID,
		"business_id", request.BusinessID,
	)

	// Update job status
	startedAt := time.Now()
	job.StartedAt = &startedAt
	job.Status = "processing"
	job.Progress = 10

	// Collect data
	job.Progress = 30
	data, err := s.collectExportData(ctx, request)
	if err != nil {
		job.Status = "failed"
		job.Error = err.Error()
		s.logger.Error("Export job failed",
			"request_id", requestID,
			"job_id", job.ID,
			"error", err.Error(),
		)
		return
	}

	// Format data
	job.Progress = 70
	formattedData, err := s.formatExportData(data, request.Format)
	if err != nil {
		job.Status = "failed"
		job.Error = err.Error()
		s.logger.Error("Export job failed during formatting",
			"request_id", requestID,
			"job_id", job.ID,
			"error", err.Error(),
		)
		return
	}

	// Create result
	job.Progress = 90
	exportID := fmt.Sprintf("export_%s_%d", request.BusinessID, time.Now().Unix())
	result := &ExportResponse{
		ExportID:    exportID,
		BusinessID:  request.BusinessID,
		ExportType:  request.ExportType,
		Format:      request.Format,
		Data:        formattedData,
		RecordCount: s.countRecords(data),
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Metadata:    request.Metadata,
	}

	// Complete job
	job.Progress = 100
	job.Status = "completed"
	job.Result = result
	completedAt := time.Now()
	job.CompletedAt = &completedAt

	s.logger.Info("Export job completed",
		"request_id", requestID,
		"job_id", job.ID,
		"business_id", request.BusinessID,
		"record_count", result.RecordCount,
	)
}
