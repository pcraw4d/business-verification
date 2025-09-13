package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// GapAnalysisReportsHandler handles gap analysis report generation
type GapAnalysisReportsHandler struct {
	reportData []ReportTemplate
}

// ReportTemplate represents a report template
type ReportTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Format      string                 `json:"format"`
	Template    map[string]interface{} `json:"template"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ReportRequest represents a report generation request
type ReportRequest struct {
	ReportType   string                 `json:"report_type"`
	Format       string                 `json:"format"`
	Filters      map[string]interface{} `json:"filters"`
	TemplateID   string                 `json:"template_id,omitempty"`
	Recipients   []string               `json:"recipients,omitempty"`
	Schedule     *ScheduleConfig        `json:"schedule,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// ScheduleConfig represents report scheduling configuration
type ScheduleConfig struct {
	Frequency string     `json:"frequency"` // daily, weekly, monthly
	Time      string     `json:"time"`      // HH:MM format
	Days      []string   `json:"days"`      // for weekly/monthly
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// ReportResponse represents a generated report
type ReportResponse struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`
	Format      string                 `json:"format"`
	Status      string                 `json:"status"`
	URL         string                 `json:"url,omitempty"`
	Size        int64                  `json:"size,omitempty"`
	GeneratedAt time.Time              `json:"generated_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ReportMetrics represents report generation metrics
type ReportMetrics struct {
	TotalReports    int       `json:"total_reports"`
	ReportsToday    int       `json:"reports_today"`
	AverageSize     float64   `json:"average_size"`
	MostPopularType string    `json:"most_popular_type"`
	SuccessRate     float64   `json:"success_rate"`
	LastGenerated   time.Time `json:"last_generated"`
}

// NewGapAnalysisReportsHandler creates a new gap analysis reports handler
func NewGapAnalysisReportsHandler() *GapAnalysisReportsHandler {
	return &GapAnalysisReportsHandler{
		reportData: getSampleReportTemplates(),
	}
}

// GenerateReport generates a new gap analysis report
func (h *GapAnalysisReportsHandler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	var request ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Validate request
	if request.ReportType == "" {
		http.Error(w, "Report type is required", http.StatusBadRequest)
		return
	}

	// Generate report based on type
	report := h.generateReportByType(request)

	response := map[string]interface{}{
		"report":       report,
		"message":      "Report generated successfully",
		"download_url": fmt.Sprintf("/v1/reports/download/%s", report.ID),
		"preview_url":  fmt.Sprintf("/v1/reports/preview/%s", report.ID),
		"generated_at": report.GeneratedAt,
		"expires_at":   report.ExpiresAt,
	}

	json.NewEncoder(w).Encode(response)
}

// GetReportTemplates returns available report templates
func (h *GapAnalysisReportsHandler) GetReportTemplates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Filter templates by type if specified
	reportType := r.URL.Query().Get("type")
	format := r.URL.Query().Get("format")

	var filteredTemplates []ReportTemplate
	for _, template := range h.reportData {
		if (reportType == "" || template.Type == reportType) &&
			(format == "" || template.Format == format) {
			filteredTemplates = append(filteredTemplates, template)
		}
	}

	response := map[string]interface{}{
		"templates":    filteredTemplates,
		"total_count":  len(filteredTemplates),
		"generated_at": time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// GetReportDetails returns details for a specific report
func (h *GapAnalysisReportsHandler) GetReportDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID := vars["id"]

	w.Header().Set("Content-Type", "application/json")

	// Find report by ID (in real implementation, this would query the database)
	report := h.getReportByID(reportID)
	if report == nil {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	// Get additional report details
	reportData := h.getReportData(reportID)
	analytics := h.getReportAnalytics(reportID)

	response := map[string]interface{}{
		"report":    report,
		"data":      reportData,
		"analytics": analytics,
		"metadata":  h.getReportMetadata(reportID),
	}

	json.NewEncoder(w).Encode(response)
}

// DownloadReport handles report download
func (h *GapAnalysisReportsHandler) DownloadReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID := vars["id"]

	// Find report by ID
	report := h.getReportByID(reportID)
	if report == nil {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	// Set appropriate headers based on format
	switch report.Format {
	case "pdf":
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.pdf\"", report.Title))
	case "excel":
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.xlsx\"", report.Title))
	case "html":
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.html\"", report.Title))
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", report.Title))
	}

	// In a real implementation, this would stream the actual file content
	// For now, we'll return a placeholder response
	w.Write([]byte(fmt.Sprintf("Report content for %s (ID: %s)", report.Title, report.ID)))
}

// PreviewReport handles report preview
func (h *GapAnalysisReportsHandler) PreviewReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID := vars["id"]

	w.Header().Set("Content-Type", "application/json")

	// Find report by ID
	report := h.getReportByID(reportID)
	if report == nil {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	// Generate preview data
	previewData := h.generatePreviewData(report)

	response := map[string]interface{}{
		"report":       report,
		"preview_data": previewData,
		"preview_url":  fmt.Sprintf("/v1/reports/preview/%s", reportID),
	}

	json.NewEncoder(w).Encode(response)
}

// ScheduleReport schedules automated report generation
func (h *GapAnalysisReportsHandler) ScheduleReport(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ReportRequest
		Schedule ScheduleConfig `json:"schedule"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Validate schedule configuration
	if request.Schedule.Frequency == "" {
		http.Error(w, "Schedule frequency is required", http.StatusBadRequest)
		return
	}

	// Create scheduled report
	scheduledReport := map[string]interface{}{
		"id":          fmt.Sprintf("scheduled_%d", time.Now().Unix()),
		"report_type": request.ReportType,
		"format":      request.Format,
		"schedule":    request.Schedule,
		"recipients":  request.Recipients,
		"status":      "scheduled",
		"created_at":  time.Now(),
		"next_run":    h.calculateNextRun(request.Schedule),
		"filters":     request.Filters,
	}

	response := map[string]interface{}{
		"scheduled_report": scheduledReport,
		"message":          "Report scheduled successfully",
		"next_run":         h.calculateNextRun(request.Schedule),
	}

	json.NewEncoder(w).Encode(response)
}

// GetReportMetrics returns report generation metrics
func (h *GapAnalysisReportsHandler) GetReportMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := h.calculateReportMetrics()

	response := map[string]interface{}{
		"metrics":      metrics,
		"generated_at": time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// GetRecentReports returns recently generated reports
func (h *GapAnalysisReportsHandler) GetRecentReports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	reportType := r.URL.Query().Get("type")
	format := r.URL.Query().Get("format")

	// Get recent reports (in real implementation, this would query the database)
	recentReports := h.getRecentReports(limit, reportType, format)

	response := map[string]interface{}{
		"reports":      recentReports,
		"total_count":  len(recentReports),
		"generated_at": time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// ExportReportData exports report data in various formats
func (h *GapAnalysisReportsHandler) ExportReportData(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	reportType := r.URL.Query().Get("type")
	filters := r.URL.Query()

	// Generate export data
	exportData := h.generateExportData(reportType, filters)

	// Set appropriate headers
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=gap_analysis_export.json")
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=gap_analysis_export.csv")
	case "xml":
		w.Header().Set("Content-Type", "application/xml")
		w.Header().Set("Content-Disposition", "attachment; filename=gap_analysis_export.xml")
	default:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=gap_analysis_export.json")
	}

	json.NewEncoder(w).Encode(exportData)
}

// Helper methods

func (h *GapAnalysisReportsHandler) generateReportByType(request ReportRequest) ReportResponse {
	now := time.Now()
	expiresAt := now.AddDate(0, 0, 30) // Reports expire after 30 days

	report := ReportResponse{
		ID:          fmt.Sprintf("report_%d", now.Unix()),
		Title:       h.generateReportTitle(request.ReportType, request.Filters),
		Type:        request.ReportType,
		Format:      request.Format,
		Status:      "completed",
		URL:         fmt.Sprintf("/v1/reports/%d", now.Unix()),
		Size:        h.calculateReportSize(request.ReportType, request.Format),
		GeneratedAt: now,
		ExpiresAt:   &expiresAt,
		Metadata: map[string]interface{}{
			"filters":       request.Filters,
			"template_id":   request.TemplateID,
			"recipients":    request.Recipients,
			"custom_fields": request.CustomFields,
		},
	}

	return report
}

func (h *GapAnalysisReportsHandler) generateReportTitle(reportType string, filters map[string]interface{}) string {
	baseTitle := fmt.Sprintf("%s Report", reportType)

	if framework, ok := filters["framework"].(string); ok && framework != "" {
		baseTitle = fmt.Sprintf("%s - %s", baseTitle, framework)
	}

	if dateRange, ok := filters["date_range"].(string); ok && dateRange != "" {
		baseTitle = fmt.Sprintf("%s (%s)", baseTitle, dateRange)
	}

	return baseTitle
}

func (h *GapAnalysisReportsHandler) calculateReportSize(reportType, format string) int64 {
	// Simulate report size calculation
	baseSize := int64(1024 * 1024) // 1MB base

	switch reportType {
	case "executive":
		baseSize = baseSize / 2
	case "detailed":
		baseSize = baseSize * 3
	case "progress":
		baseSize = baseSize
	case "compliance":
		baseSize = baseSize * 2
	}

	switch format {
	case "pdf":
		baseSize = int64(float64(baseSize) * 0.8)
	case "excel":
		baseSize = int64(float64(baseSize) * 1.2)
	case "html":
		baseSize = int64(float64(baseSize) * 0.6)
	}

	return baseSize
}

func (h *GapAnalysisReportsHandler) getReportByID(id string) *ReportResponse {
	// In real implementation, this would query the database
	// For now, return a sample report
	now := time.Now()
	expiresAt := now.AddDate(0, 0, 30)

	return &ReportResponse{
		ID:          id,
		Title:       "Sample Gap Analysis Report",
		Type:        "executive",
		Format:      "pdf",
		Status:      "completed",
		URL:         fmt.Sprintf("/v1/reports/%s", id),
		Size:        1024 * 1024, // 1MB
		GeneratedAt: now,
		ExpiresAt:   &expiresAt,
		Metadata:    make(map[string]interface{}),
	}
}

func (h *GapAnalysisReportsHandler) getReportData(reportID string) map[string]interface{} {
	// Simulate report data
	return map[string]interface{}{
		"total_gaps":       24,
		"critical_gaps":    3,
		"completed_gaps":   12,
		"average_progress": 68.5,
		"frameworks": []map[string]interface{}{
			{"name": "SOC 2", "compliance": 75, "gaps": 8},
			{"name": "GDPR", "compliance": 83, "gaps": 6},
			{"name": "PCI DSS", "compliance": 60, "gaps": 5},
		},
	}
}

func (h *GapAnalysisReportsHandler) getReportAnalytics(reportID string) map[string]interface{} {
	return map[string]interface{}{
		"views":         15,
		"downloads":     8,
		"last_accessed": time.Now().AddDate(0, 0, -2),
		"access_count":  23,
	}
}

func (h *GapAnalysisReportsHandler) getReportMetadata(reportID string) map[string]interface{} {
	return map[string]interface{}{
		"generated_by":    "system",
		"version":         "1.0",
		"template":        "default",
		"filters_applied": []string{"framework", "priority", "status"},
	}
}

func (h *GapAnalysisReportsHandler) generatePreviewData(report *ReportResponse) map[string]interface{} {
	return map[string]interface{}{
		"summary": map[string]interface{}{
			"total_gaps":     24,
			"critical_gaps":  3,
			"completed_gaps": 12,
		},
		"charts": []map[string]interface{}{
			{
				"type":  "pie",
				"title": "Gap Distribution by Priority",
				"data": map[string]int{
					"Critical": 3,
					"High":     8,
					"Medium":   10,
					"Low":      3,
				},
			},
		},
		"tables": []map[string]interface{}{
			{
				"title":   "Top Priority Gaps",
				"headers": []string{"Title", "Priority", "Progress", "Due Date"},
				"rows": [][]string{
					{"Multi-Factor Authentication", "Critical", "65%", "2025-02-15"},
					{"Data Encryption Standards", "High", "40%", "2025-03-01"},
					{"Access Control Monitoring", "High", "85%", "2025-02-01"},
				},
			},
		},
	}
}

func (h *GapAnalysisReportsHandler) calculateNextRun(schedule ScheduleConfig) time.Time {
	now := time.Now()

	switch schedule.Frequency {
	case "daily":
		return now.AddDate(0, 0, 1)
	case "weekly":
		return now.AddDate(0, 0, 7)
	case "monthly":
		return now.AddDate(0, 1, 0)
	default:
		return now.AddDate(0, 0, 1)
	}
}

func (h *GapAnalysisReportsHandler) calculateReportMetrics() ReportMetrics {
	return ReportMetrics{
		TotalReports:    156,
		ReportsToday:    8,
		AverageSize:     2.3, // MB
		MostPopularType: "executive",
		SuccessRate:     98.5,
		LastGenerated:   time.Now().AddDate(0, 0, -1),
	}
}

func (h *GapAnalysisReportsHandler) getRecentReports(limit int, reportType, format string) []ReportResponse {
	// Simulate recent reports
	reports := []ReportResponse{
		{
			ID:          "report_001",
			Title:       "Executive Summary - Q4 2024",
			Type:        "executive",
			Format:      "pdf",
			Status:      "completed",
			Size:        2.3 * 1024 * 1024,
			GeneratedAt: time.Now().AddDate(0, 0, -5),
		},
		{
			ID:          "report_002",
			Title:       "Detailed Analysis - SOC 2",
			Type:        "detailed",
			Format:      "excel",
			Status:      "completed",
			Size:        1.8 * 1024 * 1024,
			GeneratedAt: time.Now().AddDate(0, 0, -10),
		},
		{
			ID:          "report_003",
			Title:       "Progress Report - December",
			Type:        "progress",
			Format:      "html",
			Status:      "completed",
			Size:        856 * 1024,
			GeneratedAt: time.Now().AddDate(0, 0, -15),
		},
	}

	// Filter by type and format if specified
	var filtered []ReportResponse
	for _, report := range reports {
		if (reportType == "" || report.Type == reportType) &&
			(format == "" || report.Format == format) {
			filtered = append(filtered, report)
		}
	}

	// Limit results
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered
}

func (h *GapAnalysisReportsHandler) generateExportData(reportType string, filters map[string][]string) map[string]interface{} {
	return map[string]interface{}{
		"export_type":  reportType,
		"filters":      filters,
		"generated_at": time.Now(),
		"data": map[string]interface{}{
			"gaps": []map[string]interface{}{
				{
					"id":        "gap-001",
					"title":     "Multi-Factor Authentication",
					"priority":  "critical",
					"status":    "in-progress",
					"progress":  65,
					"framework": "SOC 2",
					"due_date":  "2025-02-15",
				},
				{
					"id":        "gap-002",
					"title":     "Data Encryption Standards",
					"priority":  "high",
					"status":    "in-progress",
					"progress":  40,
					"framework": "GDPR",
					"due_date":  "2025-03-01",
				},
			},
			"metrics": map[string]interface{}{
				"total_gaps":       24,
				"critical_gaps":    3,
				"completed_gaps":   12,
				"average_progress": 68.5,
			},
		},
	}
}

func getSampleReportTemplates() []ReportTemplate {
	now := time.Now()
	return []ReportTemplate{
		{
			ID:          "template_001",
			Name:        "Executive Summary Template",
			Type:        "executive",
			Description: "High-level overview for executives and stakeholders",
			Format:      "pdf",
			Template: map[string]interface{}{
				"sections": []string{"summary", "metrics", "recommendations"},
				"charts":   []string{"pie", "bar"},
				"tables":   []string{"priority_breakdown", "framework_status"},
			},
			CreatedAt: now.AddDate(0, 0, -30),
			UpdatedAt: now.AddDate(0, 0, -5),
		},
		{
			ID:          "template_002",
			Name:        "Detailed Analysis Template",
			Type:        "detailed",
			Description: "Comprehensive gap analysis with recommendations",
			Format:      "excel",
			Template: map[string]interface{}{
				"sections": []string{"overview", "detailed_analysis", "recommendations", "timeline"},
				"charts":   []string{"pie", "bar", "line", "gantt"},
				"tables":   []string{"gap_details", "milestones", "team_assignments"},
			},
			CreatedAt: now.AddDate(0, 0, -25),
			UpdatedAt: now.AddDate(0, 0, -3),
		},
		{
			ID:          "template_003",
			Name:        "Progress Report Template",
			Type:        "progress",
			Description: "Current status and progress tracking",
			Format:      "html",
			Template: map[string]interface{}{
				"sections": []string{"progress_overview", "team_performance", "timeline"},
				"charts":   []string{"progress_bar", "timeline", "team_metrics"},
				"tables":   []string{"progress_details", "team_performance"},
			},
			CreatedAt: now.AddDate(0, 0, -20),
			UpdatedAt: now.AddDate(0, 0, -1),
		},
		{
			ID:          "template_004",
			Name:        "Compliance Status Template",
			Type:        "compliance",
			Description: "Framework-specific compliance assessment",
			Format:      "pdf",
			Template: map[string]interface{}{
				"sections": []string{"framework_status", "compliance_matrix", "recommendations"},
				"charts":   []string{"compliance_gauge", "framework_comparison"},
				"tables":   []string{"framework_details", "requirement_status"},
			},
			CreatedAt: now.AddDate(0, 0, -15),
			UpdatedAt: now,
		},
	}
}
