package businessintelligence

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// ReportManager manages custom reports and data exports
type ReportManager struct {
	config    *BIConfig
	reports   map[string]*CustomReport
	templates map[string]*ReportTemplate
	mutex     sync.RWMutex
}

// CustomReport represents a custom report
type CustomReport struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Format      string                 `json:"format"`
	Parameters  map[string]interface{} `json:"parameters"`
	Filters     *ReportFilters         `json:"filters"`
	Schedule    *ReportSchedule        `json:"schedule"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Status      string                 `json:"status"`
	LastRun     *time.Time             `json:"last_run,omitempty"`
	NextRun     *time.Time             `json:"next_run,omitempty"`
}

// ReportTemplate represents a report template
type ReportTemplate struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Category    string             `json:"category"`
	Type        string             `json:"type"`
	Parameters  []*ReportParameter `json:"parameters"`
	Filters     []*ReportFilter    `json:"filters"`
	Columns     []*ReportColumn    `json:"columns"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// ReportFilters represents report filtering options
type ReportFilters struct {
	DateRange     *DateRange             `json:"date_range,omitempty"`
	TenantIDs     []string               `json:"tenant_ids,omitempty"`
	Categories    []string               `json:"categories,omitempty"`
	Statuses      []string               `json:"statuses,omitempty"`
	CustomFilters map[string]interface{} `json:"custom_filters,omitempty"`
}

// ReportSchedule represents report scheduling options
type ReportSchedule struct {
	Enabled    bool      `json:"enabled"`
	Frequency  string    `json:"frequency"` // daily, weekly, monthly, quarterly
	DayOfWeek  int       `json:"day_of_week,omitempty"`
	DayOfMonth int       `json:"day_of_month,omitempty"`
	Hour       int       `json:"hour"`
	Minute     int       `json:"minute"`
	Timezone   string    `json:"timezone"`
	NextRun    time.Time `json:"next_run"`
}

// ReportParameter represents a report parameter
type ReportParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Options     []string    `json:"options,omitempty"`
	Description string      `json:"description"`
}

// ReportFilter represents a report filter
type ReportFilter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Options     []string    `json:"options,omitempty"`
	Description string      `json:"description"`
}

// ReportColumn represents a report column
type ReportColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Format      string `json:"format,omitempty"`
	Description string `json:"description"`
	Width       int    `json:"width,omitempty"`
}

// DateRange represents a date range filter
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// NewReportManager creates a new report manager
func NewReportManager(config *BIConfig) *ReportManager {
	return &ReportManager{
		config:    config,
		reports:   make(map[string]*CustomReport),
		templates: make(map[string]*ReportTemplate),
	}
}

// CreateReport creates a new custom report
func (rm *ReportManager) CreateReport(ctx context.Context, req *CreateReportRequest) (*CustomReport, error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	report := &CustomReport{
		ID:          fmt.Sprintf("report_%d", time.Now().UnixNano()),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Format:      req.Format,
		Parameters:  req.Parameters,
		Filters:     req.Filters,
		Schedule:    req.Schedule,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Status:      "active",
	}

	// Calculate next run time if scheduled
	if report.Schedule != nil && report.Schedule.Enabled {
		report.NextRun = rm.calculateNextRun(report.Schedule)
	}

	rm.reports[report.ID] = report

	log.Printf("✅ Created custom report: %s (%s)", report.Name, report.ID)
	return report, nil
}

// GetReport retrieves a report by ID
func (rm *ReportManager) GetReport(ctx context.Context, reportID string) (*CustomReport, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	report, exists := rm.reports[reportID]
	if !exists {
		return nil, fmt.Errorf("report not found: %s", reportID)
	}

	return report, nil
}

// ListReports lists all reports with optional filtering
func (rm *ReportManager) ListReports(ctx context.Context, filter *ReportFilter) ([]*CustomReport, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	reports := make([]*CustomReport, 0, len(rm.reports))

	for _, report := range rm.reports {
		if rm.matchesFilter(report, filter) {
			reports = append(reports, report)
		}
	}

	return reports, nil
}

// GenerateReport generates a report with the specified parameters
func (rm *ReportManager) GenerateReport(ctx context.Context, reportID string, params map[string]interface{}) (*ReportResult, error) {
	rm.mutex.RLock()
	report, exists := rm.reports[reportID]
	rm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("report not found: %s", reportID)
	}

	// Generate report data based on type
	var data interface{}
	var err error

	switch report.Type {
	case "classification_summary":
		data, err = rm.generateClassificationSummary(ctx, report, params)
	case "revenue_analysis":
		data, err = rm.generateRevenueAnalysis(ctx, report, params)
	case "performance_metrics":
		data, err = rm.generatePerformanceMetrics(ctx, report, params)
	case "tenant_usage":
		data, err = rm.generateTenantUsage(ctx, report, params)
	case "custom":
		data, err = rm.generateCustomReport(ctx, report, params)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", report.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate report data: %w", err)
	}

	// Format data based on requested format
	var formattedData []byte
	switch report.Format {
	case "json":
		formattedData, err = json.MarshalIndent(data, "", "  ")
	case "csv":
		formattedData, err = rm.formatAsCSV(data)
	case "xlsx":
		formattedData, err = rm.formatAsXLSX(data)
	case "pdf":
		formattedData, err = rm.formatAsPDF(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", report.Format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to format report data: %w", err)
	}

	// Update report last run time
	rm.mutex.Lock()
	now := time.Now()
	report.LastRun = &now
	rm.mutex.Unlock()

	result := &ReportResult{
		ReportID:    reportID,
		ReportName:  report.Name,
		Format:      report.Format,
		Data:        formattedData,
		Size:        int64(len(formattedData)),
		GeneratedAt: time.Now(),
		URL:         fmt.Sprintf("/reports/%s/download", reportID),
	}

	return result, nil
}

// generateClassificationSummary generates classification summary report
func (rm *ReportManager) generateClassificationSummary(ctx context.Context, report *CustomReport, params map[string]interface{}) (interface{}, error) {
	// Simulate classification summary data
	summary := map[string]interface{}{
		"report_info": map[string]interface{}{
			"name":         report.Name,
			"generated_at": time.Now().Format(time.RFC3339),
			"date_range":   report.Filters.DateRange,
		},
		"summary": map[string]interface{}{
			"total_classifications": 45000,
			"successful":            44500,
			"failed":                500,
			"success_rate":          98.9,
			"avg_processing_time":   "45ms",
		},
		"by_industry": []map[string]interface{}{
			{"industry": "Retail", "count": 8500, "percentage": 18.9},
			{"industry": "Technology", "count": 7200, "percentage": 16.0},
			{"industry": "Finance", "count": 6800, "percentage": 15.1},
			{"industry": "Healthcare", "count": 5200, "percentage": 11.6},
			{"industry": "Manufacturing", "count": 4800, "percentage": 10.7},
			{"industry": "Services", "count": 4200, "percentage": 9.3},
			{"industry": "Other", "count": 3300, "percentage": 7.3},
		},
		"by_tenant": []map[string]interface{}{
			{"tenant": "Acme Corporation", "count": 8500, "percentage": 18.9},
			{"tenant": "TechStart Inc", "count": 7200, "percentage": 16.0},
			{"tenant": "Global Finance Ltd", "count": 6800, "percentage": 15.1},
			{"tenant": "HealthCorp", "count": 5200, "percentage": 11.6},
			{"tenant": "Manufacturing Co", "count": 4800, "percentage": 10.7},
		},
		"trends": map[string]interface{}{
			"daily_average":  1500,
			"weekly_growth":  12.5,
			"monthly_growth": 18.3,
			"peak_hours":     []int{14, 15, 16},
		},
	}

	return summary, nil
}

// generateRevenueAnalysis generates revenue analysis report
func (rm *ReportManager) generateRevenueAnalysis(ctx context.Context, report *CustomReport, params map[string]interface{}) (interface{}, error) {
	// Simulate revenue analysis data
	analysis := map[string]interface{}{
		"report_info": map[string]interface{}{
			"name":         report.Name,
			"generated_at": time.Now().Format(time.RFC3339),
			"date_range":   report.Filters.DateRange,
		},
		"summary": map[string]interface{}{
			"total_revenue":          1250000.0,
			"monthly_revenue":        125000.0,
			"growth_rate":            15.2,
			"avg_revenue_per_tenant": 27777.78,
		},
		"revenue_by_tenant": []map[string]interface{}{
			{"tenant": "Acme Corporation", "revenue": 250000.0, "percentage": 20.0},
			{"tenant": "TechStart Inc", "revenue": 180000.0, "percentage": 14.4},
			{"tenant": "Global Finance Ltd", "revenue": 220000.0, "percentage": 17.6},
			{"tenant": "HealthCorp", "revenue": 150000.0, "percentage": 12.0},
			{"tenant": "Manufacturing Co", "revenue": 120000.0, "percentage": 9.6},
		},
		"revenue_trends": []map[string]interface{}{
			{"month": "January", "revenue": 85000.0},
			{"month": "February", "revenue": 92000.0},
			{"month": "March", "revenue": 88000.0},
			{"month": "April", "revenue": 105000.0},
			{"month": "May", "revenue": 98000.0},
			{"month": "June", "revenue": 112000.0},
			{"month": "July", "revenue": 108000.0},
			{"month": "August", "revenue": 125000.0},
			{"month": "September", "revenue": 118000.0},
			{"month": "October", "revenue": 135000.0},
			{"month": "November", "revenue": 128000.0},
			{"month": "December", "revenue": 125000.0},
		},
		"forecasting": map[string]interface{}{
			"next_month_prediction": 130000.0,
			"confidence":            0.85,
			"growth_trend":          "positive",
		},
	}

	return analysis, nil
}

// generatePerformanceMetrics generates performance metrics report
func (rm *ReportManager) generatePerformanceMetrics(ctx context.Context, report *CustomReport, params map[string]interface{}) (interface{}, error) {
	// Simulate performance metrics data
	metrics := map[string]interface{}{
		"report_info": map[string]interface{}{
			"name":         report.Name,
			"generated_at": time.Now().Format(time.RFC3339),
			"date_range":   report.Filters.DateRange,
		},
		"response_times": map[string]interface{}{
			"average": 45.0,
			"median":  42.0,
			"p95":     120.0,
			"p99":     250.0,
			"min":     12.0,
			"max":     2000.0,
		},
		"throughput": map[string]interface{}{
			"requests_per_second":  1250.0,
			"peak_rps":             2100.0,
			"avg_concurrent_users": 450.0,
			"max_concurrent_users": 850.0,
		},
		"error_rates": map[string]interface{}{
			"overall":        0.8,
			"4xx_errors":     0.5,
			"5xx_errors":     0.3,
			"timeout_errors": 0.1,
		},
		"availability": map[string]interface{}{
			"uptime_percentage": 99.9,
			"downtime_minutes":  43.2,
			"incidents":         2,
			"avg_recovery_time": "5.2 minutes",
		},
		"resource_usage": map[string]interface{}{
			"cpu_usage":     65.0,
			"memory_usage":  78.0,
			"disk_usage":    45.0,
			"network_usage": 32.0,
		},
	}

	return metrics, nil
}

// generateTenantUsage generates tenant usage report
func (rm *ReportManager) generateTenantUsage(ctx context.Context, report *CustomReport, params map[string]interface{}) (interface{}, error) {
	// Simulate tenant usage data
	usage := map[string]interface{}{
		"report_info": map[string]interface{}{
			"name":         report.Name,
			"generated_at": time.Now().Format(time.RFC3339),
			"date_range":   report.Filters.DateRange,
		},
		"summary": map[string]interface{}{
			"total_tenants":     45,
			"active_tenants":    42,
			"suspended_tenants": 3,
			"total_requests":    1250000,
			"total_storage_gb":  1250.5,
		},
		"tenant_details": []map[string]interface{}{
			{
				"tenant_id":         "tenant_001",
				"name":              "Acme Corporation",
				"status":            "active",
				"requests_per_hour": 8500,
				"storage_used_gb":   125.5,
				"concurrent_users":  45,
				"api_calls_today":   75000,
				"quota_utilization": 75.0,
			},
			{
				"tenant_id":         "tenant_002",
				"name":              "TechStart Inc",
				"status":            "active",
				"requests_per_hour": 3200,
				"storage_used_gb":   45.2,
				"concurrent_users":  18,
				"api_calls_today":   28000,
				"quota_utilization": 64.0,
			},
			{
				"tenant_id":         "tenant_003",
				"name":              "Global Finance Ltd",
				"status":            "suspended",
				"requests_per_hour": 0,
				"storage_used_gb":   15.8,
				"concurrent_users":  0,
				"api_calls_today":   0,
				"quota_utilization": 0.0,
			},
		},
		"usage_trends": map[string]interface{}{
			"avg_requests_per_tenant": 27777.78,
			"avg_storage_per_tenant":  27.79,
			"growth_rate":             12.5,
			"peak_usage_hours":        []int{14, 15, 16},
		},
	}

	return usage, nil
}

// generateCustomReport generates a custom report
func (rm *ReportManager) generateCustomReport(ctx context.Context, report *CustomReport, params map[string]interface{}) (interface{}, error) {
	// Simulate custom report data based on parameters
	custom := map[string]interface{}{
		"report_info": map[string]interface{}{
			"name":         report.Name,
			"generated_at": time.Now().Format(time.RFC3339),
			"parameters":   params,
		},
		"data": map[string]interface{}{
			"custom_metric_1": 1250.0,
			"custom_metric_2": 850.0,
			"custom_metric_3": 420.0,
		},
		"details": []map[string]interface{}{
			{"category": "A", "value": 500.0},
			{"category": "B", "value": 300.0},
			{"category": "C", "value": 200.0},
		},
	}

	return custom, nil
}

// formatAsCSV formats data as CSV
func (rm *ReportManager) formatAsCSV(data interface{}) ([]byte, error) {
	// Convert data to CSV format
	var csvData strings.Builder
	writer := csv.NewWriter(&csvData)

	// Add headers
	headers := []string{"Metric", "Value", "Unit"}
	writer.Write(headers)

	// Add data rows (simplified for demo)
	rows := [][]string{
		{"Total Classifications", "45000", "count"},
		{"Success Rate", "98.9", "%"},
		{"Avg Response Time", "45", "ms"},
		{"Total Revenue", "1250000", "USD"},
	}

	for _, row := range rows {
		writer.Write(row)
	}

	writer.Flush()
	return []byte(csvData.String()), nil
}

// formatAsXLSX formats data as XLSX
func (rm *ReportManager) formatAsXLSX(data interface{}) ([]byte, error) {
	// In a real implementation, this would use a library like excelize
	// For demo purposes, return JSON data
	return json.MarshalIndent(data, "", "  ")
}

// formatAsPDF formats data as PDF
func (rm *ReportManager) formatAsPDF(data interface{}) ([]byte, error) {
	// In a real implementation, this would use a library like gofpdf
	// For demo purposes, return JSON data
	return json.MarshalIndent(data, "", "  ")
}

// calculateNextRun calculates the next run time for a scheduled report
func (rm *ReportManager) calculateNextRun(schedule *ReportSchedule) *time.Time {
	now := time.Now()

	switch schedule.Frequency {
	case "daily":
		next := time.Date(now.Year(), now.Month(), now.Day(), schedule.Hour, schedule.Minute, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		return &next
	case "weekly":
		daysUntilTarget := (schedule.DayOfWeek - int(now.Weekday()) + 7) % 7
		next := time.Date(now.Year(), now.Month(), now.Day()+daysUntilTarget, schedule.Hour, schedule.Minute, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(7 * 24 * time.Hour)
		}
		return &next
	case "monthly":
		next := time.Date(now.Year(), now.Month(), schedule.DayOfMonth, schedule.Hour, schedule.Minute, 0, 0, now.Location())
		if next.Before(now) {
			next = next.AddDate(0, 1, 0)
		}
		return &next
	}

	return nil
}

// matchesFilter checks if report matches the filter criteria
func (rm *ReportManager) matchesFilter(report *CustomReport, filter *ReportFilter) bool {
	if filter == nil {
		return true
	}

	if filter.Type != "" && report.Type != filter.Type {
		return false
	}

	if filter.Status != "" && report.Status != filter.Status {
		return false
	}

	return true
}

// CreateReportRequest represents a report creation request
type CreateReportRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Format      string                 `json:"format"`
	Parameters  map[string]interface{} `json:"parameters"`
	Filters     *ReportFilters         `json:"filters"`
	Schedule    *ReportSchedule        `json:"schedule"`
}

// ReportFilter represents report filtering criteria
type ReportFilter struct {
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
}

// ReportResult represents a generated report result
type ReportResult struct {
	ReportID    string    `json:"report_id"`
	ReportName  string    `json:"report_name"`
	Format      string    `json:"format"`
	Data        []byte    `json:"data"`
	Size        int64     `json:"size"`
	GeneratedAt time.Time `json:"generated_at"`
	URL         string    `json:"url"`
}
