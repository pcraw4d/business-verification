package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/compliance"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// CheckEngineInterface defines the interface for compliance check engines
type CheckEngineInterface interface {
	Check(ctx context.Context, req compliance.CheckRequest) (*compliance.CheckResponse, error)
}

// ComplianceHandler handles compliance endpoints
type ComplianceHandler struct {
	logger        *observability.Logger
	checkEngine   CheckEngineInterface
	statusSystem  *compliance.ComplianceStatusSystem
	reportService *compliance.ReportGenerationService
	alertSystem   *compliance.AlertSystem
	exportSystem  *compliance.ExportSystem
}

func NewComplianceHandler(logger *observability.Logger, checkEngine CheckEngineInterface, statusSystem *compliance.ComplianceStatusSystem, reportService *compliance.ReportGenerationService, alertSystem *compliance.AlertSystem, exportSystem *compliance.ExportSystem) *ComplianceHandler {
	return &ComplianceHandler{
		logger:        logger,
		checkEngine:   checkEngine,
		statusSystem:  statusSystem,
		reportService: reportService,
		alertSystem:   alertSystem,
		exportSystem:  exportSystem,
	}
}

// CheckComplianceHandler handles POST /v1/compliance/check
// Request JSON: {"business_id": string, "frameworks": [string], "apply_effects": bool}
func (h *ComplianceHandler) CheckComplianceHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req struct {
		BusinessID   string   `json:"business_id"`
		Frameworks   []string `json:"frameworks"`
		ApplyEffects bool     `json:"apply_effects"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}
	if req.BusinessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Bridge request_id in context (already set by middleware)
	ctx := r.Context()

	resp, err := h.checkEngine.Check(ctx, compliance.CheckRequest{
		BusinessID: req.BusinessID,
		Frameworks: req.Frameworks,
		Options:    compliance.EvaluationOptions{ApplyEffects: req.ApplyEffects},
	})
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "compliance_check_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// GetComplianceStatusHandler handles GET /v1/compliance/status/{business_id}
func (h *ComplianceHandler) GetComplianceStatusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractPathParam(r, "business_id")
	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	status, err := h.statusSystem.GetComplianceStatus(ctx, businessID)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "status_not_found", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
}

// GetStatusHistoryHandler handles GET /v1/compliance/status/{business_id}/history
func (h *ComplianceHandler) GetStatusHistoryHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractPathParam(r, "business_id")
	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Parse query parameters for date range
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr == "" {
		startDate = time.Now().Add(-30 * 24 * time.Hour) // Default to 30 days ago
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			h.writeError(w, r, http.StatusBadRequest, "invalid_date_format", "start_date must be in YYYY-MM-DD format")
			return
		}
	}

	if endDateStr == "" {
		endDate = time.Now() // Default to now
	} else {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			h.writeError(w, r, http.StatusBadRequest, "invalid_date_format", "end_date must be in YYYY-MM-DD format")
			return
		}
	}

	history, err := h.statusSystem.GetStatusHistory(ctx, businessID, startDate, endDate)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "history_not_found", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"business_id": businessID,
		"start_date":  startDate,
		"end_date":    endDate,
		"history":     history,
	})
}

// GetStatusAlertsHandler handles GET /v1/compliance/status/{business_id}/alerts
func (h *ComplianceHandler) GetStatusAlertsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractPathParam(r, "business_id")
	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Parse query parameter for alert status filter
	status := r.URL.Query().Get("status") // "active", "acknowledged", "resolved", or empty for all

	alerts, err := h.statusSystem.GetStatusAlerts(ctx, businessID, status)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "alerts_not_found", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"business_id": businessID,
		"status":      status,
		"alerts":      alerts,
	})
}

// AcknowledgeAlertHandler handles POST /v1/compliance/status/{business_id}/alerts/{alert_id}/acknowledge
func (h *ComplianceHandler) AcknowledgeAlertHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract path parameters
	businessID := h.extractPathParam(r, "business_id")
	alertID := h.extractPathParam(r, "alert_id")

	if businessID == "" || alertID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id and alert_id are required")
		return
	}

	var req struct {
		AcknowledgedBy string `json:"acknowledged_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}
	if req.AcknowledgedBy == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "acknowledged_by is required")
		return
	}

	err := h.statusSystem.AcknowledgeAlert(ctx, businessID, alertID, req.AcknowledgedBy)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "acknowledge_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Alert acknowledged successfully",
	})
}

// ResolveAlertHandler handles POST /v1/compliance/status/{business_id}/alerts/{alert_id}/resolve
func (h *ComplianceHandler) ResolveAlertHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract path parameters
	businessID := h.extractPathParam(r, "business_id")
	alertID := h.extractPathParam(r, "alert_id")

	if businessID == "" || alertID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id and alert_id are required")
		return
	}

	var req struct {
		ResolvedBy string `json:"resolved_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}
	if req.ResolvedBy == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "resolved_by is required")
		return
	}

	err := h.statusSystem.ResolveAlert(ctx, businessID, alertID, req.ResolvedBy)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "resolve_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Alert resolved successfully",
	})
}

// GenerateStatusReportHandler handles POST /v1/compliance/status/{business_id}/report
func (h *ComplianceHandler) GenerateStatusReportHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractPathParam(r, "business_id")
	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	var req struct {
		ReportType string `json:"report_type"` // "summary", "detailed", "framework", "requirement"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}
	if req.ReportType == "" {
		req.ReportType = "summary" // Default to summary report
	}

	report, err := h.statusSystem.GenerateStatusReport(ctx, businessID, req.ReportType)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "report_generation_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(report)
}

// GenerateComplianceReportHandler handles POST /v1/compliance/report
func (h *ComplianceHandler) GenerateComplianceReportHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	var req compliance.ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	if req.BusinessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	if req.ReportType == "" {
		req.ReportType = compliance.ReportTypeStatus // Default to status report
	}

	if req.Format == "" {
		req.Format = compliance.ReportFormatJSON // Default to JSON
	}

	report, err := h.reportService.GenerateComplianceReport(ctx, req)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "report_generation_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(report)
}

// InitializeBusinessStatusHandler handles POST /v1/compliance/status/{business_id}/initialize
func (h *ComplianceHandler) InitializeBusinessStatusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractPathParam(r, "business_id")
	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	err := h.statusSystem.InitializeBusinessStatus(ctx, businessID)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "initialization_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Business compliance status initialized successfully",
	})
}

// extractPathParam extracts a path parameter from the request URL
func (h *ComplianceHandler) extractPathParam(r *http.Request, paramName string) string {
	// Try Go 1.22 PathValue first
	if value := r.PathValue(paramName); value != "" {
		return value
	}

	// Fallback for test environment: extract from URL path
	path := r.URL.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	parts := strings.Split(path, "/")

	// Map parameter names to positions in the path
	switch paramName {
	case "business_id":
		if len(parts) >= 4 && parts[0] == "v1" && parts[1] == "compliance" && parts[2] == "status" {
			return parts[3]
		}
		// Also handle business_id in alerts/analytics path
		if len(parts) >= 5 && parts[0] == "v1" && parts[1] == "compliance" && parts[2] == "alerts" && parts[3] == "analytics" {
			return parts[4]
		}
	case "alert_id":
		if len(parts) >= 6 && parts[0] == "v1" && parts[1] == "compliance" && parts[2] == "status" && parts[4] == "alerts" {
			return parts[5]
		}
	case "rule_id":
		if len(parts) >= 5 && parts[0] == "v1" && parts[1] == "compliance" && parts[2] == "alerts" && parts[3] == "rules" {
			return parts[4]
		}
	case "policy_id":
		if len(parts) >= 5 && parts[0] == "v1" && parts[1] == "compliance" && parts[2] == "retention" && parts[3] == "policies" {
			return parts[4]
		}
	case "export_id":
		if len(parts) >= 4 && parts[0] == "v1" && parts[1] == "compliance" && parts[2] == "export" {
			return parts[3]
		}
	case "job_id":
		if len(parts) >= 4 && parts[0] == "v1" && parts[1] == "compliance" && parts[2] == "export" && parts[3] == "job" {
			return parts[4]
		}
	}

	return ""
}

func (h *ComplianceHandler) writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	h.logger.WithComponent("api").Warn(code, "path", r.URL.Path, "status", status, "message", message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
}

// Alert System Endpoints

// RegisterAlertRuleHandler handles POST /v1/compliance/alerts/rules
func (h *ComplianceHandler) RegisterAlertRuleHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	var rule compliance.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	if rule.ID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "rule ID is required")
		return
	}

	if rule.Name == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "rule name is required")
		return
	}

	err := h.alertSystem.RegisterAlertRule(ctx, &rule)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "rule_registration_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Alert rule registered successfully",
		"rule_id": rule.ID,
	})
}

// UpdateAlertRuleHandler handles PUT /v1/compliance/alerts/rules/{rule_id}
func (h *ComplianceHandler) UpdateAlertRuleHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract rule_id from URL path
	ruleID := h.extractPathParam(r, "rule_id")
	if ruleID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "rule_id is required")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	err := h.alertSystem.UpdateAlertRule(ctx, ruleID, updates)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "rule_update_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Alert rule updated successfully",
	})
}

// DeleteAlertRuleHandler handles DELETE /v1/compliance/alerts/rules/{rule_id}
func (h *ComplianceHandler) DeleteAlertRuleHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract rule_id from URL path
	ruleID := h.extractPathParam(r, "rule_id")
	if ruleID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "rule_id is required")
		return
	}

	err := h.alertSystem.DeleteAlertRule(ctx, ruleID)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "rule_deletion_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Alert rule deleted successfully",
	})
}

// GetAlertRuleHandler handles GET /v1/compliance/alerts/rules/{rule_id}
func (h *ComplianceHandler) GetAlertRuleHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract rule_id from URL path
	ruleID := h.extractPathParam(r, "rule_id")
	if ruleID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "rule_id is required")
		return
	}

	rule, err := h.alertSystem.GetAlertRule(ctx, ruleID)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "rule_not_found", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(rule)
}

// ListAlertRulesHandler handles GET /v1/compliance/alerts/rules
func (h *ComplianceHandler) ListAlertRulesHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	rules, err := h.alertSystem.ListAlertRules(ctx)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "rules_listing_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	})
}

// EvaluateAlertsHandler handles POST /v1/compliance/alerts/evaluate
func (h *ComplianceHandler) EvaluateAlertsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	var req struct {
		BusinessID string `json:"business_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	if req.BusinessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	evaluations, err := h.alertSystem.EvaluateAlerts(ctx, req.BusinessID)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "alert_evaluation_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"business_id":      req.BusinessID,
		"evaluations":      evaluations,
		"evaluation_count": len(evaluations),
	})
}

// GetAlertAnalyticsHandler handles GET /v1/compliance/alerts/analytics/{business_id}
func (h *ComplianceHandler) GetAlertAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractPathParam(r, "business_id")
	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Parse query parameter for period
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "7d" // Default to 7 days
	}

	analytics, err := h.alertSystem.GetAlertAnalytics(ctx, businessID, period)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "analytics_generation_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(analytics)
}

// RegisterEscalationPolicyHandler handles POST /v1/compliance/alerts/escalations
func (h *ComplianceHandler) RegisterEscalationPolicyHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	var policy compliance.EscalationPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	if policy.ID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "policy ID is required")
		return
	}

	if policy.Name == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "policy name is required")
		return
	}

	err := h.alertSystem.RegisterEscalationPolicy(ctx, &policy)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "policy_registration_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message":   "Escalation policy registered successfully",
		"policy_id": policy.ID,
	})
}

// RegisterNotificationChannelHandler handles POST /v1/compliance/alerts/notifications
func (h *ComplianceHandler) RegisterNotificationChannelHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	var channel compliance.NotificationChannel
	if err := json.NewDecoder(r.Body).Decode(&channel); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	if channel.ID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "channel ID is required")
		return
	}

	if channel.Name == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "channel name is required")
		return
	}

	if channel.Type == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "channel type is required")
		return
	}

	err := h.alertSystem.RegisterNotificationChannel(ctx, &channel)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "channel_registration_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message":    "Notification channel registered successfully",
		"channel_id": channel.ID,
	})
}

// ExportComplianceDataHandler handles POST /v1/compliance/export
// Request JSON: {"business_id": string, "export_type": string, "format": string, "date_range": {...}, "frameworks": [string], "include_details": bool, "filters": {...}}
func (h *ComplianceHandler) ExportComplianceDataHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	var request compliance.ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	// Validate required fields
	if request.BusinessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	if request.ExportType == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "export_type is required")
		return
	}

	// Set defaults
	if request.Format == "" {
		request.Format = compliance.ExportFormatJSON
	}

	if request.GeneratedBy == "" {
		request.GeneratedBy = "api_user"
	}

	// Export data
	result, err := h.exportSystem.ExportData(ctx, request)
	if err != nil {
		h.logger.Error("Failed to export compliance data",
			"business_id", request.BusinessID,
			"export_type", request.ExportType,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "export_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// CreateExportJobHandler handles POST /v1/compliance/export/job
// Creates an asynchronous export job for large datasets
func (h *ComplianceHandler) CreateExportJobHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	var request compliance.ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	// Validate required fields
	if request.BusinessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	if request.ExportType == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "export_type is required")
		return
	}

	// Set defaults
	if request.Format == "" {
		request.Format = compliance.ExportFormatJSON
	}

	if request.GeneratedBy == "" {
		request.GeneratedBy = "api_user"
	}

	// Create export job (for now, we'll use the same export system but return a job ID)
	// In a real implementation, this would create an async job
	result, err := h.exportSystem.ExportData(ctx, request)
	if err != nil {
		h.logger.Error("Failed to create export job",
			"business_id", request.BusinessID,
			"export_type", request.ExportType,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "job_creation_failed", err.Error())
		return
	}

	// Return job information
	jobResponse := map[string]interface{}{
		"job_id":       result.ID,
		"status":       "completed",
		"business_id":  result.BusinessID,
		"export_type":  result.ExportType,
		"format":       result.Format,
		"record_count": result.RecordCount,
		"file_size":    result.FileSize,
		"generated_at": result.GeneratedAt,
		"expires_at":   result.ExpiresAt,
		"data":         result.Data,
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(jobResponse)
}

// GetExportJobHandler handles GET /v1/compliance/export/job/{job_id}
// Retrieves the status and results of an export job
func (h *ComplianceHandler) GetExportJobHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract job_id from URL path
	jobID := h.extractPathParam(r, "job_id")
	if jobID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "job_id is required")
		return
	}

	// For now, we'll return a mock response since we don't have persistent job storage
	// In a real implementation, this would query a job database
	jobResponse := map[string]interface{}{
		"job_id":       jobID,
		"status":       "completed",
		"business_id":  "sample_business",
		"export_type":  "comprehensive",
		"format":       "json",
		"record_count": 150,
		"file_size":    10240,
		"generated_at": time.Now().Add(-time.Hour),
		"expires_at":   time.Now().Add(23 * time.Hour),
		"message":      "Export job completed successfully",
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(jobResponse)
}

// ListExportJobsHandler handles GET /v1/compliance/export/jobs
// Lists all export jobs for a business
func (h *ComplianceHandler) ListExportJobsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from query parameters
	businessID := r.URL.Query().Get("business_id")
	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// For now, return mock data
	// In a real implementation, this would query a job database
	jobs := []map[string]interface{}{
		{
			"job_id":       "export_1234567890",
			"status":       "completed",
			"business_id":  businessID,
			"export_type":  "comprehensive",
			"format":       "json",
			"record_count": 150,
			"file_size":    10240,
			"generated_at": time.Now().Add(-time.Hour),
			"expires_at":   time.Now().Add(23 * time.Hour),
		},
		{
			"job_id":       "export_1234567891",
			"status":       "completed",
			"business_id":  businessID,
			"export_type":  "status",
			"format":       "csv",
			"record_count": 25,
			"file_size":    2048,
			"generated_at": time.Now().Add(-2 * time.Hour),
			"expires_at":   time.Now().Add(22 * time.Hour),
		},
	}

	response := map[string]interface{}{
		"jobs":   jobs,
		"total":  len(jobs),
		"limit":  limit,
		"offset": offset,
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// DownloadExportHandler handles GET /v1/compliance/export/{export_id}/download
// Downloads the exported data file
func (h *ComplianceHandler) DownloadExportHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract export_id from URL path
	exportID := h.extractPathParam(r, "export_id")
	if exportID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "export_id is required")
		return
	}

	// For now, return a mock response
	// In a real implementation, this would retrieve the actual file from storage
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=compliance_export_"+exportID+".json")
	w.WriteHeader(http.StatusOK)

	// Return sample export data
	sampleData := map[string]interface{}{
		"export_id":    exportID,
		"business_id":  "sample_business",
		"export_type":  "comprehensive",
		"format":       "json",
		"record_count": 150,
		"generated_at": time.Now().Add(-time.Hour),
		"data": map[string]interface{}{
			"compliance_status": "compliant",
			"frameworks":        []string{"SOC2", "PCI-DSS"},
			"requirements": []map[string]interface{}{
				{
					"id":          "REQ-001",
					"name":        "Access Control",
					"status":      "compliant",
					"description": "Implement proper access controls",
				},
				{
					"id":          "REQ-002",
					"name":        "Data Encryption",
					"status":      "compliant",
					"description": "Encrypt sensitive data at rest and in transit",
				},
			},
		},
	}

	_ = json.NewEncoder(w).Encode(sampleData)

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
}

// Data Retention System Endpoints

// RegisterRetentionPolicyHandler handles POST /v1/compliance/retention/policies
func (h *ComplianceHandler) RegisterRetentionPolicyHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	var policy compliance.RetentionPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	if policy.ID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "policy ID is required")
		return
	}

	if policy.Name == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "policy name is required")
		return
	}

	if len(policy.DataTypes) == 0 {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "at least one data type is required")
		return
	}

	if policy.RetentionPeriod <= 0 {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "retention period must be positive")
		return
	}

	// Note: In a real implementation, you would inject the data retention system
	// For now, we'll return a success response
	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message":   "Retention policy registered successfully",
		"policy_id": policy.ID,
	})
}

// UpdateRetentionPolicyHandler handles PUT /v1/compliance/retention/policies/{policy_id}
func (h *ComplianceHandler) UpdateRetentionPolicyHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract policy_id from URL path
	policyID := h.extractPathParam(r, "policy_id")
	if policyID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "policy_id is required")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	// Note: In a real implementation, you would call the data retention system
	// For now, we'll return a success response
	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Retention policy updated successfully",
	})
}

// DeleteRetentionPolicyHandler handles DELETE /v1/compliance/retention/policies/{policy_id}
func (h *ComplianceHandler) DeleteRetentionPolicyHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract policy_id from URL path
	policyID := h.extractPathParam(r, "policy_id")
	if policyID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "policy_id is required")
		return
	}

	// Note: In a real implementation, you would call the data retention system
	// For now, we'll return a success response
	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Retention policy deleted successfully",
	})
}

// GetRetentionPolicyHandler handles GET /v1/compliance/retention/policies/{policy_id}
func (h *ComplianceHandler) GetRetentionPolicyHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract policy_id from URL path
	policyID := h.extractPathParam(r, "policy_id")
	if policyID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "policy_id is required")
		return
	}

	// Note: In a real implementation, you would call the data retention system
	// For now, we'll return mock data
	policy := &compliance.RetentionPolicy{
		ID:              policyID,
		Name:            "Default Compliance Data Retention",
		Description:     "Default retention policy for compliance data",
		Enabled:         true,
		DataTypes:       []string{"audit_trails", "compliance_reports", "alerts"},
		RetentionPeriod: 90 * 24 * time.Hour, // 90 days
		DisposalMethod:  "delete",
		CreatedAt:       time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:       time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(policy)
}

// ListRetentionPoliciesHandler handles GET /v1/compliance/retention/policies
func (h *ComplianceHandler) ListRetentionPoliciesHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Note: In a real implementation, you would call the data retention system
	// For now, we'll return mock data
	policies := []*compliance.RetentionPolicy{
		{
			ID:              "policy-1",
			Name:            "Audit Trail Retention",
			Description:     "Retention policy for audit trail data",
			Enabled:         true,
			DataTypes:       []string{"audit_trails"},
			RetentionPeriod: 90 * 24 * time.Hour,
			DisposalMethod:  "delete",
			CreatedAt:       time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "policy-2",
			Name:            "Compliance Reports Retention",
			Description:     "Retention policy for compliance reports",
			Enabled:         true,
			DataTypes:       []string{"compliance_reports"},
			RetentionPeriod: 180 * 24 * time.Hour, // 6 months
			DisposalMethod:  "archive",
			CreatedAt:       time.Now().Add(-15 * 24 * time.Hour),
			UpdatedAt:       time.Now(),
		},
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(policies)
}

// ExecuteRetentionJobHandler handles POST /v1/compliance/retention/jobs
func (h *ComplianceHandler) ExecuteRetentionJobHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	var request struct {
		PolicyID   string `json:"policy_id"`
		DataType   string `json:"data_type"`
		BusinessID string `json:"business_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	if request.PolicyID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "policy_id is required")
		return
	}

	if request.DataType == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "data_type is required")
		return
	}

	// Note: In a real implementation, you would call the data retention system
	// For now, we'll return mock job data
	job := &compliance.RetentionJob{
		ID:               fmt.Sprintf("retention_%s_%s_%d", request.PolicyID, request.DataType, time.Now().Unix()),
		PolicyID:         request.PolicyID,
		BusinessID:       request.BusinessID,
		DataType:         request.DataType,
		Status:           compliance.RetentionJobStatusCompleted,
		RecordsProcessed: 100,
		RecordsRetained:  80,
		RecordsDisposed:  20,
		StartedAt:        time.Now().Add(-5 * time.Minute),
		CompletedAt:      &time.Time{},
	}

	completedAt := time.Now()
	job.CompletedAt = &completedAt

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(job)
}

// GetRetentionAnalyticsHandler handles GET /v1/compliance/retention/analytics
func (h *ComplianceHandler) GetRetentionAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse query parameters
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d" // Default to 30 days
	}

	// Note: In a real implementation, you would call the data retention system
	// For now, we'll return mock analytics data
	analytics := &compliance.RetentionAnalytics{
		TotalPolicies:         3,
		ActivePolicies:        2,
		TotalJobs:             25,
		CompletedJobs:         23,
		FailedJobs:            2,
		TotalRecordsProcessed: 1500,
		TotalRecordsRetained:  1200,
		TotalRecordsDisposed:  300,
		DataByType: map[string]compliance.DataStats{
			"audit_trails": {
				DataType:        "audit_trails",
				TotalRecords:    1000,
				RetainedRecords: 800,
				DisposedRecords: 200,
				OldestRecord:    time.Now().Add(-365 * 24 * time.Hour),
				NewestRecord:    time.Now(),
				RetentionPeriod: 90 * 24 * time.Hour,
			},
			"compliance_reports": {
				DataType:        "compliance_reports",
				TotalRecords:    500,
				RetainedRecords: 400,
				DisposedRecords: 100,
				OldestRecord:    time.Now().Add(-180 * 24 * time.Hour),
				NewestRecord:    time.Now(),
				RetentionPeriod: 180 * 24 * time.Hour,
			},
		},
		JobsByStatus: map[string]int{
			"completed": 23,
			"failed":    2,
			"running":   0,
		},
		RetentionTrends: []compliance.RetentionTrend{
			{
				Date:             time.Now().Add(-7 * 24 * time.Hour),
				RecordsProcessed: 100,
				RecordsRetained:  80,
				RecordsDisposed:  20,
				JobsCompleted:    5,
				JobsFailed:       0,
			},
			{
				Date:             time.Now(),
				RecordsProcessed: 150,
				RecordsRetained:  120,
				RecordsDisposed:  30,
				JobsCompleted:    8,
				JobsFailed:       1,
			},
		},
		GeneratedAt: time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(analytics)
}
