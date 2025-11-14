package risk

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ExportService provides functionality to export risk data in various formats
type ExportService struct {
	logger *zap.Logger
}

// NewExportService creates a new export service
func NewExportService(logger *zap.Logger) *ExportService {
	return &ExportService{
		logger: logger,
	}
}

// ExportRiskAssessment exports a risk assessment in the specified format
func (es *ExportService) ExportRiskAssessment(ctx context.Context, assessment *RiskAssessment, format ExportFormat) (*ExportResponse, error) {
	startTime := time.Now()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	es.logger.Info("Starting risk assessment export",
		zap.String("request_id", requestID.(string)),
		zap.String("business_id", assessment.BusinessID),
		zap.String("format", string(format)))

	var data interface{}
	var err error

	switch format {
	case ExportFormatJSON:
		data, err = es.exportToJSON(assessment)
	case ExportFormatCSV:
		data, err = es.exportToCSV(assessment)
	case ExportFormatXML:
		data, err = es.exportToXML(assessment)
	case ExportFormatPDF:
		data, err = es.exportToPDF(assessment)
	case ExportFormatXLSX:
		data, err = es.exportToXLSX(assessment)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	if err != nil {
		es.logger.Error("Failed to export risk assessment",
			zap.String("request_id", requestID.(string)),
			zap.String("business_id", assessment.BusinessID),
			zap.String("format", string(format)),
			zap.Error(err))
		return nil, fmt.Errorf("export failed: %w", err)
	}

	processingTime := time.Since(startTime)
	exportID := fmt.Sprintf("export_%s_%d", assessment.BusinessID, time.Now().Unix())

	response := &ExportResponse{
		ExportID:    exportID,
		BusinessID:  assessment.BusinessID,
		ExportType:  ExportTypeAssessments,
		Format:      format,
		Data:        data,
		RecordCount: 1,
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour), // Export expires in 24 hours
		Metadata: map[string]interface{}{
			"processing_time_ms": processingTime.Milliseconds(),
			"export_size_bytes":  es.calculateDataSize(data),
			"request_id":         requestID,
		},
	}

	es.logger.Info("Risk assessment export completed",
		zap.String("request_id", requestID.(string)),
		zap.String("business_id", assessment.BusinessID),
		zap.String("export_id", exportID),
		zap.String("format", string(format)),
		zap.Duration("processing_time", processingTime))

	return response, nil
}

// ExportRiskAssessments exports multiple risk assessments in the specified format
func (es *ExportService) ExportRiskAssessments(ctx context.Context, assessments []*RiskAssessment, format ExportFormat) (*ExportResponse, error) {
	startTime := time.Now()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	es.logger.Info("Starting multiple risk assessments export",
		zap.String("request_id", requestID.(string)),
		zap.Int("assessment_count", len(assessments)),
		zap.String("format", string(format)))

	if len(assessments) == 0 {
		return nil, fmt.Errorf("no assessments to export")
	}

	var data interface{}
	var err error

	switch format {
	case ExportFormatJSON:
		data, err = es.exportMultipleToJSON(assessments)
	case ExportFormatCSV:
		data, err = es.exportMultipleToCSV(assessments)
	case ExportFormatXML:
		data, err = es.exportMultipleToXML(assessments)
	case ExportFormatPDF:
		data, err = es.exportMultipleToPDF(assessments)
	case ExportFormatXLSX:
		data, err = es.exportMultipleToXLSX(assessments)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	if err != nil {
		es.logger.Error("Failed to export multiple risk assessments",
			zap.String("request_id", requestID.(string)),
			zap.Int("assessment_count", len(assessments)),
			zap.String("format", string(format)),
			zap.Error(err))
		return nil, fmt.Errorf("export failed: %w", err)
	}

	processingTime := time.Since(startTime)
	exportID := fmt.Sprintf("export_batch_%d", time.Now().Unix())

	response := &ExportResponse{
		ExportID:    exportID,
		BusinessID:  assessments[0].BusinessID, // Use first assessment's business ID
		ExportType:  ExportTypeAssessments,
		Format:      format,
		Data:        data,
		RecordCount: len(assessments),
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Metadata: map[string]interface{}{
			"processing_time_ms": processingTime.Milliseconds(),
			"export_size_bytes":  es.calculateDataSize(data),
			"request_id":         requestID,
			"assessment_count":   len(assessments),
		},
	}

	es.logger.Info("Multiple risk assessments export completed",
		zap.String("request_id", requestID.(string)),
		zap.String("export_id", exportID),
		zap.Int("assessment_count", len(assessments)),
		zap.String("format", string(format)),
		zap.Duration("processing_time", processingTime))

	return response, nil
}

// ExportRiskFactors exports risk factors in the specified format
func (es *ExportService) ExportRiskFactors(ctx context.Context, factors []RiskScore, format ExportFormat) (*ExportResponse, error) {
	startTime := time.Now()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	es.logger.Info("Starting risk factors export",
		zap.String("request_id", requestID.(string)),
		zap.Int("factor_count", len(factors)),
		zap.String("format", string(format)))

	if len(factors) == 0 {
		return nil, fmt.Errorf("no risk factors to export")
	}

	var data interface{}
	var err error

	switch format {
	case ExportFormatJSON:
		data, err = es.exportFactorsToJSON(factors)
	case ExportFormatCSV:
		data, err = es.exportFactorsToCSV(factors)
	case ExportFormatXML:
		data, err = es.exportFactorsToXML(factors)
	case ExportFormatPDF:
		data, err = es.exportFactorsToPDF(factors)
	case ExportFormatXLSX:
		data, err = es.exportFactorsToXLSX(factors)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	if err != nil {
		es.logger.Error("Failed to export risk factors",
			zap.String("request_id", requestID.(string)),
			zap.Int("factor_count", len(factors)),
			zap.String("format", string(format)),
			zap.Error(err))
		return nil, fmt.Errorf("export failed: %w", err)
	}

	processingTime := time.Since(startTime)
	exportID := fmt.Sprintf("export_factors_%d", time.Now().Unix())

	response := &ExportResponse{
		ExportID:    exportID,
		BusinessID:  factors[0].FactorID, // Use first factor's ID as business identifier
		ExportType:  ExportTypeFactors,
		Format:      format,
		Data:        data,
		RecordCount: len(factors),
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Metadata: map[string]interface{}{
			"processing_time_ms": processingTime.Milliseconds(),
			"export_size_bytes":  es.calculateDataSize(data),
			"request_id":         requestID,
			"factor_count":       len(factors),
		},
	}

	es.logger.Info("Risk factors export completed",
		zap.String("request_id", requestID.(string)),
		zap.String("export_id", exportID),
		zap.Int("factor_count", len(factors)),
		zap.String("format", string(format)),
		zap.Duration("processing_time", processingTime))

	return response, nil
}

// ExportRiskTrends exports risk trends in the specified format
func (es *ExportService) ExportRiskTrends(ctx context.Context, trends []RiskTrend, format ExportFormat) (*ExportResponse, error) {
	startTime := time.Now()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	es.logger.Info("Starting risk trends export",
		zap.String("request_id", requestID.(string)),
		zap.Int("trend_count", len(trends)),
		zap.String("format", string(format)))

	if len(trends) == 0 {
		return nil, fmt.Errorf("no risk trends to export")
	}

	var data interface{}
	var err error

	switch format {
	case ExportFormatJSON:
		data, err = es.exportTrendsToJSON(trends)
	case ExportFormatCSV:
		data, err = es.exportTrendsToCSV(trends)
	case ExportFormatXML:
		data, err = es.exportTrendsToXML(trends)
	case ExportFormatPDF:
		data, err = es.exportTrendsToPDF(trends)
	case ExportFormatXLSX:
		data, err = es.exportTrendsToXLSX(trends)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	if err != nil {
		es.logger.Error("Failed to export risk trends",
			zap.String("request_id", requestID.(string)),
			zap.Int("trend_count", len(trends)),
			zap.String("format", string(format)),
			zap.Error(err))
		return nil, fmt.Errorf("export failed: %w", err)
	}

	processingTime := time.Since(startTime)
	exportID := fmt.Sprintf("export_trends_%d", time.Now().Unix())

	response := &ExportResponse{
		ExportID:    exportID,
		BusinessID:  trends[0].BusinessID, // Use first trend's business ID
		ExportType:  ExportTypeTrends,
		Format:      format,
		Data:        data,
		RecordCount: len(trends),
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Metadata: map[string]interface{}{
			"processing_time_ms": processingTime.Milliseconds(),
			"export_size_bytes":  es.calculateDataSize(data),
			"request_id":         requestID,
			"trend_count":        len(trends),
		},
	}

	es.logger.Info("Risk trends export completed",
		zap.String("request_id", requestID.(string)),
		zap.String("export_id", exportID),
		zap.Int("trend_count", len(trends)),
		zap.String("format", string(format)),
		zap.Duration("processing_time", processingTime))

	return response, nil
}

// ExportRiskAlerts exports risk alerts in the specified format
func (es *ExportService) ExportRiskAlerts(ctx context.Context, alerts []RiskAlert, format ExportFormat) (*ExportResponse, error) {
	startTime := time.Now()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	es.logger.Info("Starting risk alerts export",
		zap.String("request_id", requestID.(string)),
		zap.Int("alert_count", len(alerts)),
		zap.String("format", string(format)))

	if len(alerts) == 0 {
		return nil, fmt.Errorf("no risk alerts to export")
	}

	var data interface{}
	var err error

	switch format {
	case ExportFormatJSON:
		data, err = es.exportAlertsToJSON(alerts)
	case ExportFormatCSV:
		data, err = es.exportAlertsToCSV(alerts)
	case ExportFormatXML:
		data, err = es.exportAlertsToXML(alerts)
	case ExportFormatPDF:
		data, err = es.exportAlertsToPDF(alerts)
	case ExportFormatXLSX:
		data, err = es.exportAlertsToXLSX(alerts)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	if err != nil {
		es.logger.Error("Failed to export risk alerts",
			zap.String("request_id", requestID.(string)),
			zap.Int("alert_count", len(alerts)),
			zap.String("format", string(format)),
			zap.Error(err))
		return nil, fmt.Errorf("export failed: %w", err)
	}

	processingTime := time.Since(startTime)
	exportID := fmt.Sprintf("export_alerts_%d", time.Now().Unix())

	response := &ExportResponse{
		ExportID:    exportID,
		BusinessID:  alerts[0].BusinessID, // Use first alert's business ID
		ExportType:  ExportTypeAlerts,
		Format:      format,
		Data:        data,
		RecordCount: len(alerts),
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Metadata: map[string]interface{}{
			"processing_time_ms": processingTime.Milliseconds(),
			"export_size_bytes":  es.calculateDataSize(data),
			"request_id":         requestID,
			"alert_count":        len(alerts),
		},
	}

	es.logger.Info("Risk alerts export completed",
		zap.String("request_id", requestID.(string)),
		zap.String("export_id", exportID),
		zap.Int("alert_count", len(alerts)),
		zap.String("format", string(format)),
		zap.Duration("processing_time", processingTime))

	return response, nil
}

// JSON Export Methods

func (es *ExportService) exportToJSON(assessment *RiskAssessment) (string, error) {
	data, err := json.MarshalIndent(assessment, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal assessment to JSON: %w", err)
	}
	return string(data), nil
}

func (es *ExportService) exportMultipleToJSON(assessments []*RiskAssessment) (string, error) {
	data, err := json.MarshalIndent(assessments, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal assessments to JSON: %w", err)
	}
	return string(data), nil
}

func (es *ExportService) exportFactorsToJSON(factors []RiskScore) (string, error) {
	data, err := json.MarshalIndent(factors, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal factors to JSON: %w", err)
	}
	return string(data), nil
}

func (es *ExportService) exportTrendsToJSON(trends []RiskTrend) (string, error) {
	data, err := json.MarshalIndent(trends, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal trends to JSON: %w", err)
	}
	return string(data), nil
}

func (es *ExportService) exportAlertsToJSON(alerts []RiskAlert) (string, error) {
	data, err := json.MarshalIndent(alerts, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal alerts to JSON: %w", err)
	}
	return string(data), nil
}

// CSV Export Methods

func (es *ExportService) exportToCSV(assessment *RiskAssessment) (string, error) {
	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)

	// Write header
	header := []string{
		"ID", "BusinessID", "BusinessName", "OverallScore", "OverallLevel",
		"AssessedAt", "ValidUntil", "AlertLevel",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write assessment data
	record := []string{
		assessment.ID,
		assessment.BusinessID,
		assessment.BusinessName,
		strconv.FormatFloat(assessment.OverallScore, 'f', 2, 64),
		string(assessment.OverallLevel),
		assessment.AssessedAt.Format(time.RFC3339),
		assessment.ValidUntil.Format(time.RFC3339),
		string(assessment.AlertLevel),
	}
	if err := writer.Write(record); err != nil {
		return "", fmt.Errorf("failed to write CSV record: %w", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV writer error: %w", err)
	}

	return buffer.String(), nil
}

func (es *ExportService) exportMultipleToCSV(assessments []*RiskAssessment) (string, error) {
	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)

	// Write header
	header := []string{
		"ID", "BusinessID", "BusinessName", "OverallScore", "OverallLevel",
		"AssessedAt", "ValidUntil", "AlertLevel",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write assessment data
	for _, assessment := range assessments {
		record := []string{
			assessment.ID,
			assessment.BusinessID,
			assessment.BusinessName,
			strconv.FormatFloat(assessment.OverallScore, 'f', 2, 64),
			string(assessment.OverallLevel),
			assessment.AssessedAt.Format(time.RFC3339),
			assessment.ValidUntil.Format(time.RFC3339),
			string(assessment.AlertLevel),
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV writer error: %w", err)
	}

	return buffer.String(), nil
}

func (es *ExportService) exportFactorsToCSV(factors []RiskScore) (string, error) {
	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)

	// Write header
	header := []string{
		"FactorID", "FactorName", "Category", "Score", "Level",
		"Confidence", "Explanation", "CalculatedAt",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write factor data
	for _, factor := range factors {
		record := []string{
			factor.FactorID,
			factor.FactorName,
			string(factor.Category),
			strconv.FormatFloat(factor.Score, 'f', 2, 64),
			string(factor.Level),
			strconv.FormatFloat(factor.Confidence, 'f', 2, 64),
			factor.Explanation,
			factor.CalculatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV writer error: %w", err)
	}

	return buffer.String(), nil
}

func (es *ExportService) exportTrendsToCSV(trends []RiskTrend) (string, error) {
	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)

	// Write header
	header := []string{
		"BusinessID", "Category", "Score", "Level", "RecordedAt",
		"ChangeFrom", "ChangePeriod",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write trend data
	for _, trend := range trends {
		record := []string{
			trend.BusinessID,
			string(trend.Category),
			strconv.FormatFloat(trend.Score, 'f', 2, 64),
			string(trend.Level),
			trend.RecordedAt.Format(time.RFC3339),
			strconv.FormatFloat(trend.ChangeFrom, 'f', 2, 64),
			trend.ChangePeriod,
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV writer error: %w", err)
	}

	return buffer.String(), nil
}

func (es *ExportService) exportAlertsToCSV(alerts []RiskAlert) (string, error) {
	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)

	// Write header
	header := []string{
		"ID", "BusinessID", "RiskFactor", "Level", "Message",
		"Score", "Threshold", "TriggeredAt", "Acknowledged",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write alert data
	for _, alert := range alerts {
		acknowledged := "false"
		if alert.Acknowledged {
			acknowledged = "true"
		}

		acknowledgedAt := ""
		if alert.AcknowledgedAt != nil {
			acknowledgedAt = alert.AcknowledgedAt.Format(time.RFC3339)
		}
		_ = acknowledgedAt // Suppress unused variable warning - may be used in future CSV export

		record := []string{
			alert.ID,
			alert.BusinessID,
			alert.RiskFactor,
			string(alert.Level),
			alert.Message,
			strconv.FormatFloat(alert.Score, 'f', 2, 64),
			strconv.FormatFloat(alert.Threshold, 'f', 2, 64),
			alert.TriggeredAt.Format(time.RFC3339),
			acknowledged,
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV writer error: %w", err)
	}

	return buffer.String(), nil
}

// XML Export Methods

func (es *ExportService) exportToXML(assessment *RiskAssessment) (string, error) {
	data, err := xml.MarshalIndent(assessment, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal assessment to XML: %w", err)
	}
	return string(data), nil
}

func (es *ExportService) exportMultipleToXML(assessments []*RiskAssessment) (string, error) {
	type AssessmentsWrapper struct {
		XMLName     xml.Name          `xml:"risk_assessments"`
		Assessments []*RiskAssessment `xml:"assessment"`
	}

	wrapper := AssessmentsWrapper{Assessments: assessments}
	data, err := xml.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal assessments to XML: %w", err)
	}
	return string(data), nil
}

func (es *ExportService) exportFactorsToXML(factors []RiskScore) (string, error) {
	type FactorsWrapper struct {
		XMLName xml.Name    `xml:"risk_factors"`
		Factors []RiskScore `xml:"factor"`
	}

	wrapper := FactorsWrapper{Factors: factors}
	data, err := xml.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal factors to XML: %w", err)
	}
	return string(data), nil
}

func (es *ExportService) exportTrendsToXML(trends []RiskTrend) (string, error) {
	type TrendsWrapper struct {
		XMLName xml.Name    `xml:"risk_trends"`
		Trends  []RiskTrend `xml:"trend"`
	}

	wrapper := TrendsWrapper{Trends: trends}
	data, err := xml.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal trends to XML: %w", err)
	}
	return string(data), nil
}

func (es *ExportService) exportAlertsToXML(alerts []RiskAlert) (string, error) {
	type AlertsWrapper struct {
		XMLName xml.Name    `xml:"risk_alerts"`
		Alerts  []RiskAlert `xml:"alert"`
	}

	wrapper := AlertsWrapper{Alerts: alerts}
	data, err := xml.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal alerts to XML: %w", err)
	}
	return string(data), nil
}

// PDF Export Methods (Simplified - would need actual PDF library in production)

func (es *ExportService) exportToPDF(assessment *RiskAssessment) (string, error) {
	// In a real implementation, this would use a PDF library like gofpdf
	// For now, we'll return a simple text representation
	content := fmt.Sprintf(`
Risk Assessment Report
=====================

Business ID: %s
Business Name: %s
Assessment ID: %s
Overall Score: %.2f
Overall Level: %s
Assessed At: %s
Valid Until: %s
Alert Level: %s

Category Scores:
`, assessment.BusinessID, assessment.BusinessName, assessment.ID,
		assessment.OverallScore, assessment.OverallLevel,
		assessment.AssessedAt.Format(time.RFC3339),
		assessment.ValidUntil.Format(time.RFC3339),
		assessment.AlertLevel)

	for category, score := range assessment.CategoryScores {
		content += fmt.Sprintf("- %s: %.2f (%s)\n", category, score.Score, score.Level)
	}

	content += "\nFactor Scores:\n"
	for _, factor := range assessment.FactorScores {
		content += fmt.Sprintf("- %s: %.2f (%s)\n", factor.FactorName, factor.Score, factor.Level)
	}

	return content, nil
}

func (es *ExportService) exportMultipleToPDF(assessments []*RiskAssessment) (string, error) {
	content := fmt.Sprintf("Risk Assessments Report\n======================\n\nTotal Assessments: %d\n\n", len(assessments))

	for i, assessment := range assessments {
		content += fmt.Sprintf("Assessment %d:\n", i+1)
		content += fmt.Sprintf("Business ID: %s\n", assessment.BusinessID)
		content += fmt.Sprintf("Business Name: %s\n", assessment.BusinessName)
		content += fmt.Sprintf("Overall Score: %.2f\n", assessment.OverallScore)
		content += fmt.Sprintf("Overall Level: %s\n\n", assessment.OverallLevel)
	}

	return content, nil
}

func (es *ExportService) exportFactorsToPDF(factors []RiskScore) (string, error) {
	content := fmt.Sprintf("Risk Factors Report\n==================\n\nTotal Factors: %d\n\n", len(factors))

	for _, factor := range factors {
		content += fmt.Sprintf("Factor: %s\n", factor.FactorName)
		content += fmt.Sprintf("Category: %s\n", factor.Category)
		content += fmt.Sprintf("Score: %.2f\n", factor.Score)
		content += fmt.Sprintf("Level: %s\n", factor.Level)
		content += fmt.Sprintf("Confidence: %.2f\n\n", factor.Confidence)
	}

	return content, nil
}

func (es *ExportService) exportTrendsToPDF(trends []RiskTrend) (string, error) {
	content := fmt.Sprintf("Risk Trends Report\n==================\n\nTotal Trends: %d\n\n", len(trends))

	for _, trend := range trends {
		content += fmt.Sprintf("Business ID: %s\n", trend.BusinessID)
		content += fmt.Sprintf("Category: %s\n", trend.Category)
		content += fmt.Sprintf("Score: %.2f\n", trend.Score)
		content += fmt.Sprintf("Level: %s\n", trend.Level)
		content += fmt.Sprintf("Change: %.2f (%s)\n\n", trend.ChangeFrom, trend.ChangePeriod)
	}

	return content, nil
}

func (es *ExportService) exportAlertsToPDF(alerts []RiskAlert) (string, error) {
	content := fmt.Sprintf("Risk Alerts Report\n==================\n\nTotal Alerts: %d\n\n", len(alerts))

	for _, alert := range alerts {
		content += fmt.Sprintf("Alert ID: %s\n", alert.ID)
		content += fmt.Sprintf("Business ID: %s\n", alert.BusinessID)
		content += fmt.Sprintf("Risk Factor: %s\n", alert.RiskFactor)
		content += fmt.Sprintf("Level: %s\n", alert.Level)
		content += fmt.Sprintf("Message: %s\n", alert.Message)
		content += fmt.Sprintf("Score: %.2f\n", alert.Score)
		content += fmt.Sprintf("Threshold: %.2f\n\n", alert.Threshold)
	}

	return content, nil
}

// XLSX Export Methods (Simplified - would need actual XLSX library in production)

func (es *ExportService) exportToXLSX(assessment *RiskAssessment) (string, error) {
	// In a real implementation, this would use an XLSX library like excelize
	// For now, we'll return a CSV-like representation
	return es.exportToCSV(assessment)
}

func (es *ExportService) exportMultipleToXLSX(assessments []*RiskAssessment) (string, error) {
	return es.exportMultipleToCSV(assessments)
}

func (es *ExportService) exportFactorsToXLSX(factors []RiskScore) (string, error) {
	return es.exportFactorsToCSV(factors)
}

func (es *ExportService) exportTrendsToXLSX(trends []RiskTrend) (string, error) {
	return es.exportTrendsToCSV(trends)
}

func (es *ExportService) exportAlertsToXLSX(alerts []RiskAlert) (string, error) {
	return es.exportAlertsToCSV(alerts)
}

// Helper Methods

func (es *ExportService) calculateDataSize(data interface{}) int64 {
	switch d := data.(type) {
	case string:
		return int64(len(d))
	default:
		// For complex types, estimate size
		return 1024 // Default estimate
	}
}

// ValidateExportRequest validates an export request
func (es *ExportService) ValidateExportRequest(request *ExportRequest) error {
	if request.BusinessID == "" {
		return fmt.Errorf("business_id is required")
	}

	if request.ExportType == "" {
		return fmt.Errorf("export_type is required")
	}

	if request.Format == "" {
		return fmt.Errorf("format is required")
	}

	// Validate export type
	switch request.ExportType {
	case ExportTypeAssessments, ExportTypeFactors, ExportTypeTrends, ExportTypeAlerts, ExportTypeReports, ExportTypeAll:
		// Valid export types
	default:
		return fmt.Errorf("invalid export_type: %s", request.ExportType)
	}

	// Validate format
	switch request.Format {
	case ExportFormatJSON, ExportFormatCSV, ExportFormatXML, ExportFormatPDF, ExportFormatXLSX:
		// Valid formats
	default:
		return fmt.Errorf("invalid format: %s", request.Format)
	}

	return nil
}
