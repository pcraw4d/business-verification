package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kyb-platform/internal/compliance"
	"kyb-platform/internal/observability"
)

// ComplianceReportingHandler handles compliance reporting API endpoints
type ComplianceReportingHandler struct {
	logger  *observability.Logger
	service *compliance.ComplianceReportingService
}

// NewComplianceReportingHandler creates a new compliance reporting handler
func NewComplianceReportingHandler(logger *observability.Logger, service *compliance.ComplianceReportingService) *ComplianceReportingHandler {
	return &ComplianceReportingHandler{
		logger:  logger,
		service: service,
	}
}

// GenerateReportHandler handles POST /v1/compliance/reports
func (h *ComplianceReportingHandler) GenerateReportHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse request body
	var request struct {
		BusinessID  string                 `json:"business_id"`
		FrameworkID string                 `json:"framework_id"`
		ReportType  string                 `json:"report_type"`
		GeneratedBy string                 `json:"generated_by"`
		Options     map[string]interface{} `json:"options,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	// Validate required fields
	if request.BusinessID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}
	if request.FrameworkID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "framework_id is required")
		return
	}
	if request.ReportType == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "report_type is required")
		return
	}
	if request.GeneratedBy == "" {
		request.GeneratedBy = "system" // Default
	}

	// Validate report type
	validReportTypes := []string{"status", "gap_analysis", "audit", "executive_summary"}
	if !h.isValidReportType(request.ReportType, validReportTypes) {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid report_type. Must be one of: status, gap_analysis, audit, executive_summary")
		return
	}

	h.logger.Info("Generate report request received", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"business_id":  request.BusinessID,
		"framework_id": request.FrameworkID,
		"report_type":  request.ReportType,
		"generated_by": request.GeneratedBy,
	})

	// Generate report
	report, err := h.service.GenerateReport(ctx, request.BusinessID, request.FrameworkID, request.ReportType, request.GeneratedBy, request.Options)
	if err != nil {
		h.logger.Error("Failed to generate compliance report", map[string]interface{}{
			"request_id":   ctx.Value("request_id"),
			"business_id":  request.BusinessID,
			"framework_id": request.FrameworkID,
			"report_type":  request.ReportType,
			"error":        err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "report_generation_failed", "Failed to generate compliance report")
		return
	}

	// Log successful generation
	h.logger.Info("Compliance report generated successfully", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"report_id":    report.ID,
		"business_id":  request.BusinessID,
		"framework_id": request.FrameworkID,
		"report_type":  request.ReportType,
		"title":        report.Title,
		"duration_ms":  time.Since(start).Milliseconds(),
	})

	// Return generated report
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(report)
}

// GetReportHandler handles GET /v1/compliance/reports/{report_id}
func (h *ComplianceReportingHandler) GetReportHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract report_id from URL path
	reportID := h.extractReportIDFromPath(r.URL.Path)
	if reportID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "report_id is required")
		return
	}

	h.logger.Info("Get report request received", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"report_id":   reportID,
		"user_agent":  r.UserAgent(),
		"remote_addr": r.RemoteAddr,
	})

	// Get report
	report, err := h.service.GetReport(ctx, reportID)
	if err != nil {
		h.logger.Error("Failed to get compliance report", map[string]interface{}{
			"request_id": ctx.Value("request_id"),
			"report_id":  reportID,
			"error":      err.Error(),
		})
		if err.Error() == "report not found: "+reportID {
			h.writeErrorResponse(w, r, http.StatusNotFound, "report_not_found", "Compliance report not found")
		} else {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "report_retrieval_failed", "Failed to retrieve compliance report")
		}
		return
	}

	// Log successful request
	h.logger.Info("Compliance report retrieved successfully", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"report_id":    reportID,
		"business_id":  report.BusinessID,
		"framework_id": report.FrameworkID,
		"report_type":  report.ReportType,
		"status":       report.Status,
		"duration_ms":  time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}

// ListReportsHandler handles GET /v1/compliance/reports
func (h *ComplianceReportingHandler) ListReportsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse query parameters
	query := &compliance.ReportQuery{
		BusinessID:    r.URL.Query().Get("business_id"),
		FrameworkID:   r.URL.Query().Get("framework_id"),
		ReportType:    r.URL.Query().Get("report_type"),
		Status:        r.URL.Query().Get("status"),
		GeneratedBy:   r.URL.Query().Get("generated_by"),
		IncludeDrafts: r.URL.Query().Get("include_drafts") == "true",
	}

	// Parse date filters
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			query.StartDate = &startDate
		}
	}
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			query.EndDate = &endDate
		}
	}

	// Parse pagination parameters
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			query.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			query.Offset = offset
		}
	}

	h.logger.Info("List reports request received", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"query":       query,
		"user_agent":  r.UserAgent(),
		"remote_addr": r.RemoteAddr,
	})

	// List reports
	reports, err := h.service.ListReports(ctx, query)
	if err != nil {
		h.logger.Error("Failed to list compliance reports", map[string]interface{}{
			"request_id": ctx.Value("request_id"),
			"query":      query,
			"error":      err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "reports_listing_failed", "Failed to list compliance reports")
		return
	}

	// Create response
	response := map[string]interface{}{
		"reports": reports,
		"pagination": map[string]interface{}{
			"limit":  query.Limit,
			"offset": query.Offset,
			"count":  len(reports),
		},
		"filters": map[string]interface{}{
			"business_id":    query.BusinessID,
			"framework_id":   query.FrameworkID,
			"report_type":    query.ReportType,
			"status":         query.Status,
			"generated_by":   query.GeneratedBy,
			"include_drafts": query.IncludeDrafts,
		},
	}

	// Log successful request
	h.logger.Info("Compliance reports listed successfully", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"count":       len(reports),
		"duration_ms": time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetReportTemplatesHandler handles GET /v1/compliance/report-templates
func (h *ComplianceReportingHandler) GetReportTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse query parameters
	reportType := r.URL.Query().Get("report_type")

	h.logger.Info("Get report templates request received", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"report_type": reportType,
		"user_agent":  r.UserAgent(),
		"remote_addr": r.RemoteAddr,
	})

	// Get report templates (mock implementation)
	templates := h.getReportTemplates(reportType)

	// Create response
	response := map[string]interface{}{
		"templates":   templates,
		"count":       len(templates),
		"report_type": reportType,
	}

	// Log successful request
	h.logger.Info("Report templates retrieved successfully", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"count":       len(templates),
		"report_type": reportType,
		"duration_ms": time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ExportReportHandler handles GET /v1/compliance/reports/{report_id}/export
func (h *ComplianceReportingHandler) ExportReportHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract report_id from URL path
	reportID := h.extractReportIDFromPath(r.URL.Path)
	if reportID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "report_id is required")
		return
	}

	// Parse export format
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json" // Default format
	}

	// Validate format
	validFormats := []string{"json", "pdf", "html", "excel"}
	if !h.isValidFormat(format, validFormats) {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid format. Must be one of: json, pdf, html, excel")
		return
	}

	h.logger.Info("Export report request received", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"report_id":   reportID,
		"format":      format,
		"user_agent":  r.UserAgent(),
		"remote_addr": r.RemoteAddr,
	})

	// Get report
	report, err := h.service.GetReport(ctx, reportID)
	if err != nil {
		h.logger.Error("Failed to get report for export", map[string]interface{}{
			"request_id": ctx.Value("request_id"),
			"report_id":  reportID,
			"error":      err.Error(),
		})
		if err.Error() == "report not found: "+reportID {
			h.writeErrorResponse(w, r, http.StatusNotFound, "report_not_found", "Compliance report not found")
		} else {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "report_retrieval_failed", "Failed to retrieve compliance report")
		}
		return
	}

	// Export report (mock implementation)
	exportData, err := h.exportReport(report, format)
	if err != nil {
		h.logger.Error("Failed to export report", map[string]interface{}{
			"request_id": ctx.Value("request_id"),
			"report_id":  reportID,
			"format":     format,
			"error":      err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "export_failed", "Failed to export report")
		return
	}

	// Set appropriate headers based on format
	h.setExportHeaders(w, format, report.Title)

	// Log successful export
	h.logger.Info("Report exported successfully", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"report_id":   reportID,
		"format":      format,
		"title":       report.Title,
		"duration_ms": time.Since(start).Milliseconds(),
	})

	// Return export data
	w.WriteHeader(http.StatusOK)
	w.Write(exportData)
}

// Helper methods

// extractReportIDFromPath extracts report_id from URL path
func (h *ComplianceReportingHandler) extractReportIDFromPath(path string) string {
	// Expected path format: /v1/compliance/reports/{report_id} or /v1/compliance/reports/{report_id}/export
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "v1" && parts[1] == "compliance" && parts[2] == "reports" {
		return parts[3]
	}
	return ""
}

// isValidReportType checks if report type is valid
func (h *ComplianceReportingHandler) isValidReportType(reportType string, validTypes []string) bool {
	for _, validType := range validTypes {
		if reportType == validType {
			return true
		}
	}
	return false
}

// isValidFormat checks if export format is valid
func (h *ComplianceReportingHandler) isValidFormat(format string, validFormats []string) bool {
	for _, validFormat := range validFormats {
		if format == validFormat {
			return true
		}
	}
	return false
}

// getReportTemplates returns available report templates
func (h *ComplianceReportingHandler) getReportTemplates(reportType string) []map[string]interface{} {
	templates := []map[string]interface{}{
		{
			"id":             "status_template",
			"name":           "Compliance Status Report",
			"description":    "Standard compliance status report",
			"report_type":    "status",
			"default_format": "json",
			"customizable":   true,
		},
		{
			"id":             "gap_analysis_template",
			"name":           "Gap Analysis Report",
			"description":    "Compliance gap analysis report",
			"report_type":    "gap_analysis",
			"default_format": "json",
			"customizable":   true,
		},
		{
			"id":             "executive_summary_template",
			"name":           "Executive Summary Report",
			"description":    "Executive compliance summary report",
			"report_type":    "executive_summary",
			"default_format": "json",
			"customizable":   true,
		},
		{
			"id":             "audit_template",
			"name":           "Audit Report",
			"description":    "Compliance audit report",
			"report_type":    "audit",
			"default_format": "json",
			"customizable":   true,
		},
	}

	// Filter by report type if specified
	if reportType != "" {
		var filtered []map[string]interface{}
		for _, template := range templates {
			if template["report_type"] == reportType {
				filtered = append(filtered, template)
			}
		}
		return filtered
	}

	return templates
}

// exportReport exports a report in the specified format
func (h *ComplianceReportingHandler) exportReport(report *compliance.ComplianceReport, format string) ([]byte, error) {
	switch format {
	case "json":
		return json.MarshalIndent(report, "", "  ")
	case "pdf":
		// Mock PDF export - would integrate with PDF generation library
		return []byte("PDF export not implemented"), nil
	case "html":
		// Mock HTML export - would generate HTML report
		return []byte("HTML export not implemented"), nil
	case "excel":
		// Mock Excel export - would integrate with Excel generation library
		return []byte("Excel export not implemented"), nil
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// setExportHeaders sets appropriate headers for export
func (h *ComplianceReportingHandler) setExportHeaders(w http.ResponseWriter, format, title string) {
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.json\"", title))
	case "pdf":
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.pdf\"", title))
	case "html":
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.html\"", title))
	case "excel":
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.xlsx\"", title))
	}
}

// writeErrorResponse writes an error response
func (h *ComplianceReportingHandler) writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, errorCode, message string) {
	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    errorCode,
			"message": message,
		},
		"timestamp": time.Now().UTC(),
		"path":      r.URL.Path,
		"method":    r.Method,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}
