package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/reporting"
)

// ReportHandler handles report API requests
type ReportHandler struct {
	reportService ReportService
	logger        *zap.Logger
}

// ReportService interface for report management
type ReportService interface {
	GenerateReport(ctx context.Context, request *reporting.ReportRequest) (*reporting.ReportResponse, error)
	GetReport(ctx context.Context, tenantID, reportID string) (*reporting.Report, error)
	ListReports(ctx context.Context, filter *reporting.ReportFilter) (*reporting.ReportListResponse, error)
	DeleteReport(ctx context.Context, tenantID, reportID string) error
	GetReportMetrics(ctx context.Context, tenantID string) (*reporting.ReportMetrics, error)
	CreateTemplate(ctx context.Context, request *reporting.ReportTemplateRequest) (*reporting.ReportTemplateResponse, error)
	GetTemplate(ctx context.Context, tenantID, templateID string) (*reporting.ReportTemplate, error)
	ListTemplates(ctx context.Context, filter *reporting.ReportTemplateFilter) (*reporting.ReportTemplateListResponse, error)
	UpdateTemplate(ctx context.Context, tenantID, templateID string, request *reporting.ReportTemplateRequest) (*reporting.ReportTemplateResponse, error)
	DeleteTemplate(ctx context.Context, tenantID, templateID string) error
	CreateScheduledReport(ctx context.Context, request *reporting.ScheduledReportRequest) (*reporting.ScheduledReportResponse, error)
	GetScheduledReport(ctx context.Context, tenantID, scheduledReportID string) (*reporting.ScheduledReport, error)
	ListScheduledReports(ctx context.Context, filter *reporting.ScheduledReportFilter) (*reporting.ScheduledReportListResponse, error)
	UpdateScheduledReport(ctx context.Context, tenantID, scheduledReportID string, request *reporting.ScheduledReportRequest) (*reporting.ScheduledReportResponse, error)
	DeleteScheduledReport(ctx context.Context, tenantID, scheduledReportID string) error
	RunScheduledReport(ctx context.Context, tenantID, scheduledReportID string) (*reporting.ReportResponse, error)
}

// NewReportHandler creates a new report handler
func NewReportHandler(reportService ReportService, logger *zap.Logger) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		logger:        logger,
	}
}

// HandleGenerateReport handles POST /api/v1/reports/generate
func (h *ReportHandler) HandleGenerateReport(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling generate report request")

	// Parse request body
	var request reporting.ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateReportRequest(&request); err != nil {
		h.logger.Error("Invalid report request", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract tenant ID from context
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Set tenant ID in request context
	ctx := r.Context()
	ctx = context.WithValue(ctx, "tenant_id", tenantID)

	// Generate report
	response, err := h.reportService.GenerateReport(ctx, &request)
	if err != nil {
		h.logger.Error("Failed to generate report", zap.Error(err))
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // 202 Accepted for async processing
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetReport handles GET /api/v1/reports/{id}
func (h *ReportHandler) HandleGetReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID := vars["id"]

	h.logger.Info("Handling get report request",
		zap.String("report_id", reportID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get report
	report, err := h.reportService.GetReport(r.Context(), tenantID, reportID)
	if err != nil {
		h.logger.Error("Failed to get report", zap.Error(err))
		http.Error(w, "Failed to get report", http.StatusInternalServerError)
		return
	}

	if report == nil {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleListReports handles GET /api/v1/reports
func (h *ReportHandler) HandleListReports(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling list reports request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := h.parseReportFilter(r, tenantID)

	// List reports
	response, err := h.reportService.ListReports(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list reports", zap.Error(err))
		http.Error(w, "Failed to list reports", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleDeleteReport handles DELETE /api/v1/reports/{id}
func (h *ReportHandler) HandleDeleteReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID := vars["id"]

	h.logger.Info("Handling delete report request",
		zap.String("report_id", reportID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Delete report
	if err := h.reportService.DeleteReport(r.Context(), tenantID, reportID); err != nil {
		h.logger.Error("Failed to delete report", zap.Error(err))
		http.Error(w, "Failed to delete report", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// HandleGetReportMetrics handles GET /api/v1/reports/metrics
func (h *ReportHandler) HandleGetReportMetrics(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling get report metrics request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get report metrics
	metrics, err := h.reportService.GetReportMetrics(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get report metrics", zap.Error(err))
		http.Error(w, "Failed to get report metrics", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleDownloadReport handles GET /api/v1/reports/{id}/download
func (h *ReportHandler) HandleDownloadReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID := vars["id"]

	h.logger.Info("Handling download report request",
		zap.String("report_id", reportID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get report
	report, err := h.reportService.GetReport(r.Context(), tenantID, reportID)
	if err != nil {
		h.logger.Error("Failed to get report", zap.Error(err))
		http.Error(w, "Failed to get report", http.StatusInternalServerError)
		return
	}

	if report == nil {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	if report.Status != reporting.ReportStatusCompleted {
		http.Error(w, "Report not ready for download", http.StatusBadRequest)
		return
	}

	// Check if report has expired
	if report.ExpiresAt != nil && time.Now().After(*report.ExpiresAt) {
		http.Error(w, "Report has expired", http.StatusGone)
		return
	}

	// In a real implementation, you would serve the actual file
	// For now, we'll return a redirect to the download URL
	http.Redirect(w, r, report.DownloadURL, http.StatusTemporaryRedirect)
}

// Template Management Handlers

// HandleCreateTemplate handles POST /api/v1/reports/templates
func (h *ReportHandler) HandleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling create template request")

	// Parse request body
	var request reporting.ReportTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateTemplateRequest(&request); err != nil {
		h.logger.Error("Invalid template request", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract tenant ID from context
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Set tenant ID in request context
	ctx := r.Context()
	ctx = context.WithValue(ctx, "tenant_id", tenantID)

	// Create template
	response, err := h.reportService.CreateTemplate(ctx, &request)
	if err != nil {
		h.logger.Error("Failed to create template", zap.Error(err))
		http.Error(w, "Failed to create template", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetTemplate handles GET /api/v1/reports/templates/{id}
func (h *ReportHandler) HandleGetTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["id"]

	h.logger.Info("Handling get template request",
		zap.String("template_id", templateID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get template
	template, err := h.reportService.GetTemplate(r.Context(), tenantID, templateID)
	if err != nil {
		h.logger.Error("Failed to get template", zap.Error(err))
		http.Error(w, "Failed to get template", http.StatusInternalServerError)
		return
	}

	if template == nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(template); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleListTemplates handles GET /api/v1/reports/templates
func (h *ReportHandler) HandleListTemplates(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling list templates request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := h.parseTemplateFilter(r, tenantID)

	// List templates
	response, err := h.reportService.ListTemplates(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list templates", zap.Error(err))
		http.Error(w, "Failed to list templates", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleUpdateTemplate handles PUT /api/v1/reports/templates/{id}
func (h *ReportHandler) HandleUpdateTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["id"]

	h.logger.Info("Handling update template request",
		zap.String("template_id", templateID))

	// Parse request body
	var request reporting.ReportTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateTemplateRequest(&request); err != nil {
		h.logger.Error("Invalid template request", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Update template
	response, err := h.reportService.UpdateTemplate(r.Context(), tenantID, templateID, &request)
	if err != nil {
		h.logger.Error("Failed to update template", zap.Error(err))
		http.Error(w, "Failed to update template", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleDeleteTemplate handles DELETE /api/v1/reports/templates/{id}
func (h *ReportHandler) HandleDeleteTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["id"]

	h.logger.Info("Handling delete template request",
		zap.String("template_id", templateID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Delete template
	if err := h.reportService.DeleteTemplate(r.Context(), tenantID, templateID); err != nil {
		h.logger.Error("Failed to delete template", zap.Error(err))
		http.Error(w, "Failed to delete template", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// Scheduled Report Handlers

// HandleCreateScheduledReport handles POST /api/v1/reports/scheduled
func (h *ReportHandler) HandleCreateScheduledReport(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling create scheduled report request")

	// Parse request body
	var request reporting.ScheduledReportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateScheduledReportRequest(&request); err != nil {
		h.logger.Error("Invalid scheduled report request", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract tenant ID from context
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Set tenant ID in request context
	ctx := r.Context()
	ctx = context.WithValue(ctx, "tenant_id", tenantID)

	// Create scheduled report
	response, err := h.reportService.CreateScheduledReport(ctx, &request)
	if err != nil {
		h.logger.Error("Failed to create scheduled report", zap.Error(err))
		http.Error(w, "Failed to create scheduled report", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetScheduledReport handles GET /api/v1/reports/scheduled/{id}
func (h *ReportHandler) HandleGetScheduledReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scheduledReportID := vars["id"]

	h.logger.Info("Handling get scheduled report request",
		zap.String("scheduled_report_id", scheduledReportID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get scheduled report
	scheduledReport, err := h.reportService.GetScheduledReport(r.Context(), tenantID, scheduledReportID)
	if err != nil {
		h.logger.Error("Failed to get scheduled report", zap.Error(err))
		http.Error(w, "Failed to get scheduled report", http.StatusInternalServerError)
		return
	}

	if scheduledReport == nil {
		http.Error(w, "Scheduled report not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(scheduledReport); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleListScheduledReports handles GET /api/v1/reports/scheduled
func (h *ReportHandler) HandleListScheduledReports(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling list scheduled reports request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := h.parseScheduledReportFilter(r, tenantID)

	// List scheduled reports
	response, err := h.reportService.ListScheduledReports(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list scheduled reports", zap.Error(err))
		http.Error(w, "Failed to list scheduled reports", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleUpdateScheduledReport handles PUT /api/v1/reports/scheduled/{id}
func (h *ReportHandler) HandleUpdateScheduledReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scheduledReportID := vars["id"]

	h.logger.Info("Handling update scheduled report request",
		zap.String("scheduled_report_id", scheduledReportID))

	// Parse request body
	var request reporting.ScheduledReportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateScheduledReportRequest(&request); err != nil {
		h.logger.Error("Invalid scheduled report request", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Update scheduled report
	response, err := h.reportService.UpdateScheduledReport(r.Context(), tenantID, scheduledReportID, &request)
	if err != nil {
		h.logger.Error("Failed to update scheduled report", zap.Error(err))
		http.Error(w, "Failed to update scheduled report", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleDeleteScheduledReport handles DELETE /api/v1/reports/scheduled/{id}
func (h *ReportHandler) HandleDeleteScheduledReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scheduledReportID := vars["id"]

	h.logger.Info("Handling delete scheduled report request",
		zap.String("scheduled_report_id", scheduledReportID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Delete scheduled report
	if err := h.reportService.DeleteScheduledReport(r.Context(), tenantID, scheduledReportID); err != nil {
		h.logger.Error("Failed to delete scheduled report", zap.Error(err))
		http.Error(w, "Failed to delete scheduled report", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// HandleRunScheduledReport handles POST /api/v1/reports/scheduled/{id}/run
func (h *ReportHandler) HandleRunScheduledReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scheduledReportID := vars["id"]

	h.logger.Info("Handling run scheduled report request",
		zap.String("scheduled_report_id", scheduledReportID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Run scheduled report
	response, err := h.reportService.RunScheduledReport(r.Context(), tenantID, scheduledReportID)
	if err != nil {
		h.logger.Error("Failed to run scheduled report", zap.Error(err))
		http.Error(w, "Failed to run scheduled report", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// Helper functions

// validateReportRequest validates a report request
func validateReportRequest(request *reporting.ReportRequest) error {
	if request.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(request.Name) > 255 {
		return fmt.Errorf("name exceeds maximum length of 255 characters")
	}

	if request.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}

	// Validate report type
	validTypes := map[reporting.ReportType]bool{
		reporting.ReportTypeExecutiveSummary: true,
		reporting.ReportTypeCompliance:       true,
		reporting.ReportTypeRiskAudit:        true,
		reporting.ReportTypeTrendAnalysis:    true,
		reporting.ReportTypeCustom:           true,
		reporting.ReportTypeBatchResults:     true,
		reporting.ReportTypePerformance:      true,
	}

	if !validTypes[request.Type] {
		return fmt.Errorf("invalid report type: %s", request.Type)
	}

	// Validate report format
	validFormats := map[reporting.ReportFormat]bool{
		reporting.ReportFormatPDF:   true,
		reporting.ReportFormatExcel: true,
		reporting.ReportFormatCSV:   true,
		reporting.ReportFormatJSON:  true,
		reporting.ReportFormatHTML:  true,
	}

	if !validFormats[request.Format] {
		return fmt.Errorf("invalid report format: %s", request.Format)
	}

	return nil
}

// validateTemplateRequest validates a template request
func validateTemplateRequest(request *reporting.ReportTemplateRequest) error {
	if request.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(request.Name) > 255 {
		return fmt.Errorf("name exceeds maximum length of 255 characters")
	}

	if request.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}

	// Validate report type
	validTypes := map[reporting.ReportType]bool{
		reporting.ReportTypeExecutiveSummary: true,
		reporting.ReportTypeCompliance:       true,
		reporting.ReportTypeRiskAudit:        true,
		reporting.ReportTypeTrendAnalysis:    true,
		reporting.ReportTypeCustom:           true,
		reporting.ReportTypeBatchResults:     true,
		reporting.ReportTypePerformance:      true,
	}

	if !validTypes[request.Type] {
		return fmt.Errorf("invalid report type: %s", request.Type)
	}

	return nil
}

// validateScheduledReportRequest validates a scheduled report request
func validateScheduledReportRequest(request *reporting.ScheduledReportRequest) error {
	if request.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(request.Name) > 255 {
		return fmt.Errorf("name exceeds maximum length of 255 characters")
	}

	if request.TemplateID == "" {
		return fmt.Errorf("template_id is required")
	}

	if request.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}

	// Validate schedule frequency
	validFrequencies := map[reporting.ScheduleFrequency]bool{
		reporting.ScheduleFrequencyOnce:      true,
		reporting.ScheduleFrequencyDaily:     true,
		reporting.ScheduleFrequencyWeekly:    true,
		reporting.ScheduleFrequencyMonthly:   true,
		reporting.ScheduleFrequencyQuarterly: true,
		reporting.ScheduleFrequencyYearly:    true,
	}

	if !validFrequencies[request.Schedule.Frequency] {
		return fmt.Errorf("invalid schedule frequency: %s", request.Schedule.Frequency)
	}

	return nil
}

// extractTenantID extracts tenant ID from request context
func (h *ReportHandler) extractTenantID(r *http.Request) string {
	// This would be implemented based on your authentication/authorization system
	// For now, return a default tenant ID
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}
	return "default"
}

// parseReportFilter parses query parameters into a report filter
func (h *ReportHandler) parseReportFilter(r *http.Request, tenantID string) *reporting.ReportFilter {
	filter := &reporting.ReportFilter{
		TenantID: tenantID,
	}

	// Parse report type filter
	if reportType := r.URL.Query().Get("type"); reportType != "" {
		filter.Type = reporting.ReportType(reportType)
	}

	// Parse status filter
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = reporting.ReportStatus(status)
	}

	// Parse format filter
	if format := r.URL.Query().Get("format"); format != "" {
		filter.Format = reporting.ReportFormat(format)
	}

	// Parse created by filter
	if createdBy := r.URL.Query().Get("created_by"); createdBy != "" {
		filter.CreatedBy = createdBy
	}

	// Parse date range
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	// Parse pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	return filter
}

// parseTemplateFilter parses query parameters into a template filter
func (h *ReportHandler) parseTemplateFilter(r *http.Request, tenantID string) *reporting.ReportTemplateFilter {
	filter := &reporting.ReportTemplateFilter{
		TenantID: tenantID,
	}

	// Parse template type filter
	if templateType := r.URL.Query().Get("type"); templateType != "" {
		filter.Type = reporting.ReportType(templateType)
	}

	// Parse public filter
	if isPublicStr := r.URL.Query().Get("is_public"); isPublicStr != "" {
		if isPublic, err := strconv.ParseBool(isPublicStr); err == nil {
			filter.IsPublic = &isPublic
		}
	}

	// Parse default filter
	if isDefaultStr := r.URL.Query().Get("is_default"); isDefaultStr != "" {
		if isDefault, err := strconv.ParseBool(isDefaultStr); err == nil {
			filter.IsDefault = &isDefault
		}
	}

	// Parse created by filter
	if createdBy := r.URL.Query().Get("created_by"); createdBy != "" {
		filter.CreatedBy = createdBy
	}

	// Parse pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	return filter
}

// parseScheduledReportFilter parses query parameters into a scheduled report filter
func (h *ReportHandler) parseScheduledReportFilter(r *http.Request, tenantID string) *reporting.ScheduledReportFilter {
	filter := &reporting.ScheduledReportFilter{
		TenantID: tenantID,
	}

	// Parse active filter
	if isActiveStr := r.URL.Query().Get("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filter.IsActive = &isActive
		}
	}

	// Parse created by filter
	if createdBy := r.URL.Query().Get("created_by"); createdBy != "" {
		filter.CreatedBy = createdBy
	}

	// Parse pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	return filter
}
