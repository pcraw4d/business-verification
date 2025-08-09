package handlers

import (
	"context"
	"encoding/json"
	"net/http"
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
	logger       *observability.Logger
	checkEngine  CheckEngineInterface
	statusSystem *compliance.ComplianceStatusSystem
}

func NewComplianceHandler(logger *observability.Logger, checkEngine CheckEngineInterface, statusSystem *compliance.ComplianceStatusSystem) *ComplianceHandler {
	return &ComplianceHandler{
		logger:       logger,
		checkEngine:  checkEngine,
		statusSystem: statusSystem,
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
	case "alert_id":
		if len(parts) >= 6 && parts[0] == "v1" && parts[1] == "compliance" && parts[2] == "status" && parts[4] == "alerts" {
			return parts[5]
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
