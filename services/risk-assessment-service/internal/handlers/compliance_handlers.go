package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/audit"
	"kyb-platform/services/risk-assessment-service/internal/compliance"
)

// ComplianceHandler handles compliance-related API requests
type ComplianceHandler struct {
	reporter    *compliance.ComplianceReporter
	reportGen   *compliance.ReportGenerator
	auditLogger *audit.AuditLogger
	logger      *zap.Logger
}

// NewComplianceHandler creates a new compliance handler
func NewComplianceHandler(reporter *compliance.ComplianceReporter, reportGen *compliance.ReportGenerator, auditLogger *audit.AuditLogger, logger *zap.Logger) *ComplianceHandler {
	return &ComplianceHandler{
		reporter:    reporter,
		reportGen:   reportGen,
		auditLogger: auditLogger,
		logger:      logger,
	}
}

// GenerateComplianceReport generates a new compliance report
func (h *ComplianceHandler) GenerateComplianceReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req compliance.GenerateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode compliance report request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Generating compliance report",
		zap.String("tenant_id", req.TenantID),
		zap.String("template_id", req.TemplateID),
		zap.String("report_name", req.ReportName),
		zap.String("period", req.Period))

	// Generate compliance report
	report, err := h.reporter.GenerateComplianceReport(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to generate compliance report", zap.Error(err))
		http.Error(w, "Failed to generate compliance report", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Compliance report generated successfully",
		"data":    report,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetComplianceReport retrieves a compliance report
func (h *ComplianceHandler) GetComplianceReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	reportID := vars["report_id"]

	h.logger.Info("Retrieving compliance report",
		zap.String("report_id", reportID))

	// Get compliance report
	report, err := h.reporter.GetComplianceReport(ctx, reportID)
	if err != nil {
		h.logger.Error("Failed to get compliance report", zap.Error(err))
		http.Error(w, "Failed to retrieve compliance report", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    report,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetComplianceReports retrieves compliance reports for a tenant
func (h *ComplianceHandler) GetComplianceReports(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	// Parse query parameters
	reportType := r.URL.Query().Get("report_type")

	// Parse date range
	startDate := time.Now().AddDate(0, 0, -30) // Default to last 30 days
	endDate := time.Now()

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = parsed
		}
	}

	h.logger.Info("Retrieving compliance reports",
		zap.String("tenant_id", tenantID),
		zap.String("report_type", reportType),
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate))

	// Get compliance reports
	reports, err := h.reporter.GetComplianceReports(ctx, tenantID, reportType, startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get compliance reports", zap.Error(err))
		http.Error(w, "Failed to retrieve compliance reports", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    reports,
		"count":   len(reports),
		"filters": map[string]interface{}{
			"tenant_id":   tenantID,
			"report_type": reportType,
			"start_date":  startDate,
			"end_date":    endDate,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ExportComplianceReport exports a compliance report in the specified format
func (h *ComplianceHandler) ExportComplianceReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		compliance.GenerateReportRequest
		Format string `json:"format"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode export request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate format
	validFormats := map[string]bool{"json": true, "csv": true, "pdf": true}
	if !validFormats[req.Format] {
		http.Error(w, "Invalid export format. Supported formats: json, csv, pdf", http.StatusBadRequest)
		return
	}

	h.logger.Info("Exporting compliance report",
		zap.String("tenant_id", req.TenantID),
		zap.String("template_id", req.TemplateID),
		zap.String("format", req.Format))

	// Generate and export report
	export, err := h.reportGen.GenerateAndExportReport(ctx, &req.GenerateReportRequest, req.Format)
	if err != nil {
		h.logger.Error("Failed to export compliance report", zap.Error(err))
		http.Error(w, "Failed to export compliance report", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Compliance report exported successfully",
		"data":    export,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// CreateReportTemplate creates a new report template
func (h *ComplianceHandler) CreateReportTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req compliance.CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode template request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Creating report template",
		zap.String("name", req.Name),
		zap.String("type", req.Type),
		zap.String("created_by", req.CreatedBy))

	// Create report template
	template, err := h.reporter.CreateReportTemplate(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create report template", zap.Error(err))
		http.Error(w, "Failed to create report template", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Report template created successfully",
		"data":    template,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetReportTemplate retrieves a report template
func (h *ComplianceHandler) GetReportTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	templateID := vars["template_id"]

	h.logger.Info("Retrieving report template",
		zap.String("template_id", templateID))

	// Get report template
	template, err := h.reporter.GetReportTemplate(ctx, templateID)
	if err != nil {
		h.logger.Error("Failed to get report template", zap.Error(err))
		http.Error(w, "Failed to retrieve report template", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    template,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetReportTemplates retrieves report templates by type
func (h *ComplianceHandler) GetReportTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	reportType := r.URL.Query().Get("type")

	h.logger.Info("Retrieving report templates",
		zap.String("type", reportType))

	// Get report templates
	templates, err := h.reporter.GetReportTemplates(ctx, reportType)
	if err != nil {
		h.logger.Error("Failed to get report templates", zap.Error(err))
		http.Error(w, "Failed to retrieve report templates", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    templates,
		"count":   len(templates),
		"filters": map[string]interface{}{
			"type": reportType,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteComplianceReport deletes a compliance report
func (h *ComplianceHandler) DeleteComplianceReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	reportID := vars["report_id"]

	var req struct {
		DeletedBy string `json:"deleted_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode delete request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Deleting compliance report",
		zap.String("report_id", reportID),
		zap.String("deleted_by", req.DeletedBy))

	// Delete compliance report
	if err := h.reporter.DeleteComplianceReport(ctx, reportID, req.DeletedBy); err != nil {
		h.logger.Error("Failed to delete compliance report", zap.Error(err))
		http.Error(w, "Failed to delete compliance report", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Compliance report deleted successfully",
		"data": map[string]interface{}{
			"report_id":  reportID,
			"deleted_by": req.DeletedBy,
			"deleted_at": time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetComplianceStatus retrieves overall compliance status
func (h *ComplianceHandler) GetComplianceStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	h.logger.Info("Retrieving compliance status",
		zap.String("tenant_id", tenantID))

	// Mock compliance status (in production, this would come from compliance monitoring)
	status := map[string]interface{}{
		"overall_status": "compliant",
		"last_updated":   time.Now(),
		"frameworks": map[string]interface{}{
			"soc2": map[string]interface{}{
				"status":     "compliant",
				"last_audit": "2024-01-15",
				"next_audit": "2024-07-15",
				"score":      95,
			},
			"gdpr": map[string]interface{}{
				"status":      "compliant",
				"last_review": "2024-01-10",
				"next_review": "2024-07-10",
				"score":       98,
			},
			"pci_dss": map[string]interface{}{
				"status":          "compliant",
				"level":           "1",
				"last_assessment": "2024-01-20",
				"score":           92,
			},
		},
		"risk_score": 0.15,
		"risk_level": "low",
		"alerts": []map[string]interface{}{
			{
				"id":         "alert_1",
				"type":       "compliance",
				"severity":   "low",
				"message":    "Quarterly review due in 30 days",
				"created_at": time.Now().AddDate(0, 0, -60),
			},
		},
	}

	// Log compliance status access
	h.auditLogger.LogDataAccess(ctx, tenantID, "", "compliance_status", "", "read", map[string]interface{}{
		"overall_status": status["overall_status"],
		"risk_score":     status["risk_score"],
	})

	response := map[string]interface{}{
		"success": true,
		"data":    status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetComplianceMetrics retrieves compliance metrics
func (h *ComplianceHandler) GetComplianceMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	// Parse date range
	startDate := time.Now().AddDate(0, 0, -30) // Default to last 30 days
	endDate := time.Now()

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = parsed
		}
	}

	h.logger.Info("Retrieving compliance metrics",
		zap.String("tenant_id", tenantID),
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate))

	// Mock compliance metrics (in production, these would be calculated from real data)
	metrics := map[string]interface{}{
		"period": map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
		},
		"compliance_score": 94.5,
		"audit_events": map[string]interface{}{
			"total":           1250,
			"security_events": 45,
			"admin_actions":   89,
			"data_access":     1116,
		},
		"violations": map[string]interface{}{
			"total":    3,
			"resolved": 2,
			"pending":  1,
			"critical": 0,
			"high":     1,
			"medium":   2,
			"low":      0,
		},
		"reports_generated": map[string]interface{}{
			"total":      12,
			"audit":      5,
			"security":   3,
			"compliance": 4,
		},
		"trends": map[string]interface{}{
			"compliance_score_trend": "+2.5%",
			"violations_trend":       "-15%",
			"audit_events_trend":     "+8%",
		},
	}

	// Log compliance metrics access
	h.auditLogger.LogDataAccess(ctx, tenantID, "", "compliance_metrics", "", "read", map[string]interface{}{
		"compliance_score":   metrics["compliance_score"],
		"total_audit_events": metrics["audit_events"].(map[string]interface{})["total"],
	})

	response := map[string]interface{}{
		"success": true,
		"data":    metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ScheduleComplianceReport schedules automatic compliance report generation
func (h *ComplianceHandler) ScheduleComplianceReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req compliance.ReportSchedule
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode schedule request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set default values
	req.ID = generateScheduleID()
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	h.logger.Info("Scheduling compliance report",
		zap.String("tenant_id", req.TenantID),
		zap.String("template_id", req.TemplateID),
		zap.String("cron_expression", req.CronExpression),
		zap.String("created_by", req.CreatedBy))

	// Schedule report generation
	if err := h.reportGen.ScheduleReportGeneration(ctx, &req); err != nil {
		h.logger.Error("Failed to schedule compliance report", zap.Error(err))
		http.Error(w, "Failed to schedule compliance report", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Compliance report scheduled successfully",
		"data":    req,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Helper function
func generateScheduleID() string {
	return fmt.Sprintf("sched_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}
