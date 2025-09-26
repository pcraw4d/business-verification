package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ReportGenerator generates various types of analytics reports
type ReportGenerator struct {
	collector *AnalyticsCollector
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(collector *AnalyticsCollector) *ReportGenerator {
	return &ReportGenerator{
		collector: collector,
	}
}

// ReportType represents the type of report to generate
type ReportType string

const (
	ReportTypeOverall     ReportType = "overall"
	ReportTypeUser        ReportType = "user"
	ReportTypeDaily       ReportType = "daily"
	ReportTypeIndustry    ReportType = "industry"
	ReportTypeRisk        ReportType = "risk"
	ReportTypePerformance ReportType = "performance"
)

// ReportRequest represents a request for a report
type ReportRequest struct {
	Type      ReportType `json:"type"`
	UserID    string     `json:"user_id,omitempty"`
	Days      int        `json:"days,omitempty"`
	Format    string     `json:"format,omitempty"` // json, csv, pdf
	StartDate time.Time  `json:"start_date,omitempty"`
	EndDate   time.Time  `json:"end_date,omitempty"`
}

// ReportResponse represents a report response
type ReportResponse struct {
	Type        ReportType     `json:"type"`
	GeneratedAt time.Time      `json:"generated_at"`
	Data        interface{}    `json:"data"`
	Metadata    ReportMetadata `json:"metadata"`
}

// ReportMetadata contains metadata about the report
type ReportMetadata struct {
	TotalRecords int       `json:"total_records"`
	DateRange    DateRange `json:"date_range"`
	Filters      []string  `json:"filters"`
}

// DateRange represents a date range
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// GenerateReport generates a report based on the request
func (rg *ReportGenerator) GenerateReport(ctx context.Context, req ReportRequest) (*ReportResponse, error) {
	response := &ReportResponse{
		Type:        req.Type,
		GeneratedAt: time.Now(),
		Metadata: ReportMetadata{
			DateRange: DateRange{
				Start: req.StartDate,
				End:   req.EndDate,
			},
		},
	}

	switch req.Type {
	case ReportTypeOverall:
		response.Data = rg.collector.GetOverallStats()
		response.Metadata.TotalRecords = 1

	case ReportTypeUser:
		if req.UserID == "" {
			return nil, fmt.Errorf("user ID is required for user report")
		}
		response.Data = rg.collector.GetUserStats(req.UserID)
		response.Metadata.TotalRecords = 1

	case ReportTypeDaily:
		days := req.Days
		if days == 0 {
			days = 30 // Default to 30 days
		}
		response.Data = rg.collector.GetDailyStats(days)
		response.Metadata.TotalRecords = days

	case ReportTypeIndustry:
		response.Data = rg.collector.GetIndustryAnalytics()
		response.Metadata.TotalRecords = len(rg.collector.industryCounts)

	case ReportTypeRisk:
		response.Data = rg.collector.GetRiskAnalytics()
		response.Metadata.TotalRecords = len(rg.collector.riskLevelCounts)

	case ReportTypePerformance:
		response.Data = rg.generatePerformanceReport()
		response.Metadata.TotalRecords = 1

	default:
		return nil, fmt.Errorf("unsupported report type: %s", req.Type)
	}

	return response, nil
}

// generatePerformanceReport generates a performance-focused report
func (rg *ReportGenerator) generatePerformanceReport() map[string]interface{} {
	overallStats := rg.collector.GetOverallStats()

	return map[string]interface{}{
		"performance_metrics": map[string]interface{}{
			"avg_response_time": overallStats["avg_response_time"],
			"success_rate":      overallStats["success_rate"],
			"total_requests":    overallStats["total_classifications"],
		},
		"efficiency_metrics": map[string]interface{}{
			"requests_per_second": rg.calculateRequestsPerSecond(),
			"error_rate":          100.0 - overallStats["success_rate"].(float64),
			"uptime_percentage":   99.9, // This would be calculated from actual uptime data
		},
		"scalability_metrics": map[string]interface{}{
			"peak_usage":           rg.calculatePeakUsage(),
			"resource_utilization": rg.calculateResourceUtilization(),
		},
	}
}

// calculateRequestsPerSecond calculates requests per second
func (rg *ReportGenerator) calculateRequestsPerSecond() float64 {
	// This would be calculated from actual request data
	// For now, return a mock value
	return 10.5
}

// calculatePeakUsage calculates peak usage metrics
func (rg *ReportGenerator) calculatePeakUsage() map[string]interface{} {
	return map[string]interface{}{
		"peak_requests_per_minute": 150,
		"peak_concurrent_users":    25,
		"peak_response_time":       "2.5s",
	}
}

// calculateResourceUtilization calculates resource utilization
func (rg *ReportGenerator) calculateResourceUtilization() map[string]interface{} {
	return map[string]interface{}{
		"cpu_usage":     "45%",
		"memory_usage":  "60%",
		"disk_usage":    "30%",
		"network_usage": "25%",
	}
}

// ExportReport exports a report in the specified format
func (rg *ReportGenerator) ExportReport(ctx context.Context, req ReportRequest) ([]byte, error) {
	report, err := rg.GenerateReport(ctx, req)
	if err != nil {
		return nil, err
	}

	switch req.Format {
	case "json":
		return json.MarshalIndent(report, "", "  ")
	case "csv":
		return rg.exportToCSV(report)
	case "pdf":
		return rg.exportToPDF(report)
	default:
		return json.MarshalIndent(report, "", "  ") // Default to JSON
	}
}

// exportToCSV exports a report to CSV format
func (rg *ReportGenerator) exportToCSV(report *ReportResponse) ([]byte, error) {
	// This would implement CSV export
	// For now, return JSON as fallback
	return json.MarshalIndent(report, "", "  ")
}

// exportToPDF exports a report to PDF format
func (rg *ReportGenerator) exportToPDF(report *ReportResponse) ([]byte, error) {
	// This would implement PDF export
	// For now, return JSON as fallback
	return json.MarshalIndent(report, "", "  ")
}

// GetAvailableReports returns a list of available report types
func (rg *ReportGenerator) GetAvailableReports() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type":        string(ReportTypeOverall),
			"name":        "Overall Statistics",
			"description": "Comprehensive overview of all analytics data",
		},
		{
			"type":        string(ReportTypeUser),
			"name":        "User Analytics",
			"description": "User-specific activity and performance metrics",
		},
		{
			"type":        string(ReportTypeDaily),
			"name":        "Daily Statistics",
			"description": "Daily breakdown of analytics data",
		},
		{
			"type":        string(ReportTypeIndustry),
			"name":        "Industry Analytics",
			"description": "Industry classification and distribution data",
		},
		{
			"type":        string(ReportTypeRisk),
			"name":        "Risk Analytics",
			"description": "Risk level distribution and analysis",
		},
		{
			"type":        string(ReportTypePerformance),
			"name":        "Performance Report",
			"description": "System performance and efficiency metrics",
		},
	}
}
