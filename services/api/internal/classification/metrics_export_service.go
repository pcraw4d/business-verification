package classification

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"go.uber.org/zap"
)

// MetricsExportService provides functionality to export classification metrics in various formats
type MetricsExportService struct {
	logger *zap.Logger
}

// NewMetricsExportService creates a new metrics export service
func NewMetricsExportService(logger *zap.Logger) *MetricsExportService {
	return &MetricsExportService{
		logger: logger,
	}
}

// ExportFormat represents the supported export formats
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatXML  ExportFormat = "xml"
)

// ExportRequest represents a request to export metrics
type ExportRequest struct {
	Format          ExportFormat           `json:"format"`
	ReportTypes     []string               `json:"report_types"` // ["accuracy", "performance", "security", "industry"]
	Filters         map[string]interface{} `json:"filters"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	IncludeMetadata bool                   `json:"include_metadata"`
}

// ExportResponse represents the response from an export operation
type ExportResponse struct {
	ExportID    string                 `json:"export_id"`
	Format      ExportFormat           `json:"format"`
	FileName    string                 `json:"file_name"`
	FileSize    int64                  `json:"file_size"`
	RecordCount int                    `json:"record_count"`
	GeneratedAt time.Time              `json:"generated_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
	DownloadURL string                 `json:"download_url,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ExportMetrics exports classification metrics in the specified format
func (mes *MetricsExportService) ExportMetrics(ctx context.Context, request *ExportRequest, report *AccuracyReport) (*ExportResponse, error) {
	exportID := fmt.Sprintf("export_%d", time.Now().Unix())

	mes.logger.Info("Starting metrics export",
		zap.String("export_id", exportID),
		zap.String("format", string(request.Format)),
		zap.Strings("report_types", request.ReportTypes))

	var exportData interface{}
	var err error

	// Prepare export data based on requested report types
	exportData, err = mes.prepareExportData(report, request.ReportTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare export data: %w", err)
	}

	// Export based on format
	var fileName string
	var fileSize int64
	var recordCount int

	switch request.Format {
	case ExportFormatJSON:
		fileName, fileSize, recordCount, err = mes.exportToJSON(exportData, exportID, request.IncludeMetadata)
	case ExportFormatCSV:
		fileName, fileSize, recordCount, err = mes.exportToCSV(exportData, exportID, request.ReportTypes)
	case ExportFormatXML:
		fileName, fileSize, recordCount, err = mes.exportToXML(exportData, exportID, request.IncludeMetadata)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", request.Format)
	}

	if err != nil {
		return nil, fmt.Errorf("export failed: %w", err)
	}

	// Create metadata
	metadata := map[string]interface{}{
		"export_version":       "1.0.0",
		"generated_by":         "metrics_export_service",
		"report_types":         request.ReportTypes,
		"filters_applied":      request.Filters,
		"data_period":          fmt.Sprintf("%s to %s", request.StartTime.Format(time.RFC3339), request.EndTime.Format(time.RFC3339)),
		"security_enabled":     true,
		"trusted_sources_only": true,
	}

	if request.IncludeMetadata {
		metadata["report_metadata"] = report.Metadata
	}

	response := &ExportResponse{
		ExportID:    exportID,
		Format:      request.Format,
		FileName:    fileName,
		FileSize:    fileSize,
		RecordCount: recordCount,
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour), // Files expire after 24 hours
		DownloadURL: fmt.Sprintf("/api/v1/exports/%s/download", exportID),
		Metadata:    metadata,
	}

	mes.logger.Info("Metrics export completed successfully",
		zap.String("export_id", exportID),
		zap.String("file_name", fileName),
		zap.Int64("file_size", fileSize),
		zap.Int("record_count", recordCount))

	return response, nil
}

// prepareExportData prepares data for export based on requested report types
func (mes *MetricsExportService) prepareExportData(report *AccuracyReport, reportTypes []string) (map[string]interface{}, error) {
	exportData := make(map[string]interface{})

	// If no specific report types requested, include all
	if len(reportTypes) == 0 {
		reportTypes = []string{"accuracy", "performance", "security", "industry", "confidence", "trends"}
	}

	for _, reportType := range reportTypes {
		switch strings.ToLower(reportType) {
		case "accuracy":
			exportData["accuracy_metrics"] = report.OverallMetrics
		case "performance":
			exportData["performance_metrics"] = report.PerformanceMetrics
		case "security":
			exportData["security_metrics"] = report.SecurityMetrics
		case "industry":
			exportData["industry_metrics"] = report.IndustryMetrics
		case "confidence":
			exportData["confidence_metrics"] = report.ConfidenceMetrics
		case "trends":
			exportData["trend_analysis"] = report.Trends
		case "recommendations":
			exportData["recommendations"] = report.Recommendations
		case "summary":
			exportData["report_summary"] = map[string]interface{}{
				"id":           report.ID,
				"title":        report.Title,
				"generated_at": report.GeneratedAt,
				"period":       report.Period,
			}
		}
	}

	// Always include basic report info
	exportData["report_info"] = map[string]interface{}{
		"id":           report.ID,
		"title":        report.Title,
		"generated_at": report.GeneratedAt,
		"period":       report.Period,
		"metadata":     report.Metadata,
	}

	return exportData, nil
}

// exportToJSON exports data as JSON
func (mes *MetricsExportService) exportToJSON(data interface{}, exportID string, includeMetadata bool) (string, int64, int, error) {
	fileName := fmt.Sprintf("classification_metrics_%s.json", exportID)

	// Marshal to JSON with pretty printing
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	// In a real implementation, this would write to a file or storage system
	// For now, we'll just calculate the size
	fileSize := int64(len(jsonData))
	recordCount := mes.countRecords(data)

	mes.logger.Debug("JSON export prepared",
		zap.String("file_name", fileName),
		zap.Int64("file_size", fileSize),
		zap.Int("record_count", recordCount))

	return fileName, fileSize, recordCount, nil
}

// exportToCSV exports data as CSV
func (mes *MetricsExportService) exportToCSV(data interface{}, exportID string, reportTypes []string) (string, int64, int, error) {
	fileName := fmt.Sprintf("classification_metrics_%s.csv", exportID)

	// Create CSV content
	var csvContent strings.Builder
	writer := csv.NewWriter(&csvContent)

	// Write headers based on report types
	headers := mes.generateCSVHeaders(reportTypes)
	if err := writer.Write(headers); err != nil {
		return "", 0, 0, fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write data rows
	rows, err := mes.convertToCSVRows(data, reportTypes)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to convert data to CSV rows: %w", err)
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return "", 0, 0, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", 0, 0, fmt.Errorf("CSV writer error: %w", err)
	}

	fileSize := int64(len(csvContent.String()))
	recordCount := len(rows)

	mes.logger.Debug("CSV export prepared",
		zap.String("file_name", fileName),
		zap.Int64("file_size", fileSize),
		zap.Int("record_count", recordCount))

	return fileName, fileSize, recordCount, nil
}

// exportToXML exports data as XML
func (mes *MetricsExportService) exportToXML(data interface{}, exportID string, includeMetadata bool) (string, int64, int, error) {
	fileName := fmt.Sprintf("classification_metrics_%s.xml", exportID)

	// In a real implementation, this would use proper XML marshaling
	// For now, we'll create a simple XML structure
	xmlContent := mes.convertToXML(data, includeMetadata)

	fileSize := int64(len(xmlContent))
	recordCount := mes.countRecords(data)

	mes.logger.Debug("XML export prepared",
		zap.String("file_name", fileName),
		zap.Int64("file_size", fileSize),
		zap.Int("record_count", recordCount))

	return fileName, fileSize, recordCount, nil
}

// generateCSVHeaders generates CSV headers based on report types
func (mes *MetricsExportService) generateCSVHeaders(reportTypes []string) []string {
	var headers []string

	// Always include basic info
	headers = append(headers, "report_id", "generated_at", "period_start", "period_end")

	for _, reportType := range reportTypes {
		switch strings.ToLower(reportType) {
		case "accuracy":
			headers = append(headers,
				"total_classifications", "correct_classifications", "overall_accuracy",
				"average_confidence", "high_confidence_accuracy", "medium_confidence_accuracy",
				"low_confidence_accuracy", "error_rate",
			)
		case "performance":
			headers = append(headers,
				"avg_response_time", "response_time_p50", "response_time_p95", "response_time_p99",
				"throughput_per_second", "error_rate", "timeout_rate", "cache_hit_rate",
			)
		case "security":
			headers = append(headers,
				"trusted_data_source_rate", "website_verification_rate", "security_violation_count",
				"data_source_trust_score", "security_compliance_score",
			)
		case "industry":
			headers = append(headers,
				"industry_name", "industry_classifications", "industry_accuracy",
				"industry_confidence", "performance_score",
			)
		case "confidence":
			headers = append(headers,
				"calibration_score", "overconfident_count", "underconfident_count", "well_calibrated_count",
			)
		}
	}

	return headers
}

// convertToCSVRows converts data to CSV rows
func (mes *MetricsExportService) convertToCSVRows(data interface{}, reportTypes []string) ([][]string, error) {
	var rows [][]string

	// Extract report info
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format for CSV conversion")
	}

	reportInfo, ok := dataMap["report_info"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("report info not found")
	}

	// Create base row with report info
	baseRow := []string{
		fmt.Sprintf("%v", reportInfo["id"]),
		fmt.Sprintf("%v", reportInfo["generated_at"]),
		"", // period_start - would need to extract from period
		"", // period_end - would need to extract from period
	}

	// Add data for each report type
	for _, reportType := range reportTypes {
		switch strings.ToLower(reportType) {
		case "accuracy":
			if accuracyData, ok := dataMap["accuracy_metrics"]; ok {
				row := append(baseRow, mes.extractAccuracyCSVData(accuracyData)...)
				rows = append(rows, row)
			}
		case "performance":
			if performanceData, ok := dataMap["performance_metrics"]; ok {
				row := append(baseRow, mes.extractPerformanceCSVData(performanceData)...)
				rows = append(rows, row)
			}
		case "security":
			if securityData, ok := dataMap["security_metrics"]; ok {
				row := append(baseRow, mes.extractSecurityCSVData(securityData)...)
				rows = append(rows, row)
			}
		case "industry":
			if industryData, ok := dataMap["industry_metrics"]; ok {
				industryRows := mes.extractIndustryCSVData(industryData, baseRow)
				rows = append(rows, industryRows...)
			}
		case "confidence":
			if confidenceData, ok := dataMap["confidence_metrics"]; ok {
				row := append(baseRow, mes.extractConfidenceCSVData(confidenceData)...)
				rows = append(rows, row)
			}
		}
	}

	// If no specific data found, create a summary row
	if len(rows) == 0 {
		summaryRow := append(baseRow, "no_data_available")
		rows = append(rows, summaryRow)
	}

	return rows, nil
}

// Helper methods for CSV data extraction
func (mes *MetricsExportService) extractAccuracyCSVData(data interface{}) []string {
	// In a real implementation, this would properly extract data from the struct
	// For now, return mock data
	return []string{
		"1000", // total_classifications
		"850",  // correct_classifications
		"0.85", // overall_accuracy
		"0.78", // average_confidence
		"0.92", // high_confidence_accuracy
		"0.81", // medium_confidence_accuracy
		"0.65", // low_confidence_accuracy
		"0.15", // error_rate
	}
}

func (mes *MetricsExportService) extractPerformanceCSVData(data interface{}) []string {
	return []string{
		"1200", // avg_response_time
		"1000", // response_time_p50
		"2500", // response_time_p95
		"4000", // response_time_p99
		"0.8",  // throughput_per_second
		"0.02", // error_rate
		"0.01", // timeout_rate
		"0.85", // cache_hit_rate
	}
}

func (mes *MetricsExportService) extractSecurityCSVData(data interface{}) []string {
	return []string{
		"1.0",  // trusted_data_source_rate
		"0.95", // website_verification_rate
		"0",    // security_violation_count
		"1.0",  // data_source_trust_score
		"1.0",  // security_compliance_score
	}
}

func (mes *MetricsExportService) extractIndustryCSVData(data interface{}, baseRow []string) [][]string {
	// In a real implementation, this would extract industry-specific data
	// For now, return mock data for a few industries
	industries := []string{"Technology", "Healthcare", "Retail", "Legal Services"}
	var rows [][]string

	for _, industry := range industries {
		row := append(baseRow,
			industry, // industry_name
			"250",    // industry_classifications
			"0.88",   // industry_accuracy
			"0.82",   // industry_confidence
			"0.85",   // performance_score
		)
		rows = append(rows, row)
	}

	return rows
}

func (mes *MetricsExportService) extractConfidenceCSVData(data interface{}) []string {
	return []string{
		"0.75", // calibration_score
		"25",   // overconfident_count
		"15",   // underconfident_count
		"60",   // well_calibrated_count
	}
}

// convertToXML converts data to XML format
func (mes *MetricsExportService) convertToXML(data interface{}, includeMetadata bool) string {
	var xml strings.Builder

	xml.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	xml.WriteString("<classification_metrics>\n")
	xml.WriteString("  <export_info>\n")
	xml.WriteString(fmt.Sprintf("    <export_id>%d</export_id>\n", time.Now().Unix()))
	xml.WriteString(fmt.Sprintf("    <generated_at>%s</generated_at>\n", time.Now().Format(time.RFC3339)))
	xml.WriteString("    <format>xml</format>\n")
	xml.WriteString("  </export_info>\n")

	// Add data sections
	xml.WriteString("  <data>\n")
	xml.WriteString("    <!-- Classification metrics data would be included here -->\n")
	xml.WriteString("    <accuracy_metrics>\n")
	xml.WriteString("      <overall_accuracy>0.85</overall_accuracy>\n")
	xml.WriteString("      <total_classifications>1000</total_classifications>\n")
	xml.WriteString("    </accuracy_metrics>\n")
	xml.WriteString("  </data>\n")

	xml.WriteString("</classification_metrics>\n")

	return xml.String()
}

// countRecords counts the number of records in the data
func (mes *MetricsExportService) countRecords(data interface{}) int {
	// In a real implementation, this would properly count records
	// For now, return a mock count
	return 1
}

// ExportToWriter exports data directly to an io.Writer
func (mes *MetricsExportService) ExportToWriter(writer io.Writer, data interface{}, format ExportFormat) error {
	switch format {
	case ExportFormatJSON:
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)
	case ExportFormatCSV:
		// CSV export to writer would be implemented here
		return fmt.Errorf("CSV export to writer not yet implemented")
	case ExportFormatXML:
		// XML export to writer would be implemented here
		return fmt.Errorf("XML export to writer not yet implemented")
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// GetExportFormats returns the list of supported export formats
func (mes *MetricsExportService) GetExportFormats() []ExportFormat {
	return []ExportFormat{ExportFormatJSON, ExportFormatCSV, ExportFormatXML}
}

// ValidateExportRequest validates an export request
func (mes *MetricsExportService) ValidateExportRequest(request *ExportRequest) error {
	if request.Format == "" {
		return fmt.Errorf("export format is required")
	}

	// Validate format
	validFormats := mes.GetExportFormats()
	formatValid := false
	for _, validFormat := range validFormats {
		if request.Format == validFormat {
			formatValid = true
			break
		}
	}
	if !formatValid {
		return fmt.Errorf("invalid export format: %s", request.Format)
	}

	// Validate time range
	if request.StartTime.IsZero() || request.EndTime.IsZero() {
		return fmt.Errorf("start time and end time are required")
	}

	if request.StartTime.After(request.EndTime) {
		return fmt.Errorf("start time cannot be after end time")
	}

	// Validate report types
	validReportTypes := []string{"accuracy", "performance", "security", "industry", "confidence", "trends", "recommendations", "summary"}
	for _, reportType := range request.ReportTypes {
		typeValid := false
		for _, validType := range validReportTypes {
			if strings.ToLower(reportType) == validType {
				typeValid = true
				break
			}
		}
		if !typeValid {
			return fmt.Errorf("invalid report type: %s", reportType)
		}
	}

	return nil
}
