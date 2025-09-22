package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kyb-platform/internal/compliance"
	"kyb-platform/internal/observability"
)

// ComplianceAlertHandler handles compliance alert API endpoints
type ComplianceAlertHandler struct {
	logger  *observability.Logger
	service *compliance.ComplianceAlertService
}

// NewComplianceAlertHandler creates a new compliance alert handler
func NewComplianceAlertHandler(logger *observability.Logger, service *compliance.ComplianceAlertService) *ComplianceAlertHandler {
	return &ComplianceAlertHandler{
		logger:  logger,
		service: service,
	}
}

// CreateAlertHandler handles POST /v1/compliance/alerts
func (h *ComplianceAlertHandler) CreateAlertHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse request body
	var alert compliance.ComplianceAlert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	// Validate required fields
	if alert.BusinessID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}
	if alert.FrameworkID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "framework_id is required")
		return
	}
	if alert.AlertType == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "alert_type is required")
		return
	}
	if alert.Severity == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "severity is required")
		return
	}
	if alert.Title == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "title is required")
		return
	}

	// Validate alert type
	validAlertTypes := []string{"deadline", "risk_threshold", "compliance_change", "gap_detected", "milestone_overdue"}
	if !h.isValidAlertType(alert.AlertType, validAlertTypes) {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid alert_type. Must be one of: deadline, risk_threshold, compliance_change, gap_detected, milestone_overdue")
		return
	}

	// Validate severity
	validSeverities := []string{"low", "medium", "high", "critical"}
	if !h.isValidSeverity(alert.Severity, validSeverities) {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid severity. Must be one of: low, medium, high, critical")
		return
	}

	// Set default triggered_by if not provided
	if alert.TriggeredBy == "" {
		alert.TriggeredBy = "user"
	}

	h.logger.Info("Create alert request received", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"business_id":  alert.BusinessID,
		"framework_id": alert.FrameworkID,
		"alert_type":   alert.AlertType,
		"severity":     alert.Severity,
		"title":        alert.Title,
	})

	// Create alert
	err := h.service.CreateAlert(ctx, &alert)
	if err != nil {
		h.logger.Error("Failed to create compliance alert", map[string]interface{}{
			"request_id":   ctx.Value("request_id"),
			"business_id":  alert.BusinessID,
			"framework_id": alert.FrameworkID,
			"alert_type":   alert.AlertType,
			"error":        err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "alert_creation_failed", "Failed to create compliance alert")
		return
	}

	// Log successful creation
	h.logger.Info("Compliance alert created successfully", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"alert_id":     alert.ID,
		"business_id":  alert.BusinessID,
		"framework_id": alert.FrameworkID,
		"alert_type":   alert.AlertType,
		"severity":     alert.Severity,
		"status":       alert.Status,
		"duration_ms":  time.Since(start).Milliseconds(),
	})

	// Return created alert
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(alert)
}

// GetAlertHandler handles GET /v1/compliance/alerts/{alert_id}
func (h *ComplianceAlertHandler) GetAlertHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract alert_id from URL path
	alertID := h.extractAlertIDFromPath(r.URL.Path)
	if alertID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "alert_id is required")
		return
	}

	h.logger.Info("Get alert request received", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"alert_id":    alertID,
		"user_agent":  r.UserAgent(),
		"remote_addr": r.RemoteAddr,
	})

	// Get alert
	alert, err := h.service.GetAlert(ctx, alertID)
	if err != nil {
		h.logger.Error("Failed to get compliance alert", map[string]interface{}{
			"request_id": ctx.Value("request_id"),
			"alert_id":   alertID,
			"error":      err.Error(),
		})
		if err.Error() == "alert not found: "+alertID {
			h.writeErrorResponse(w, r, http.StatusNotFound, "alert_not_found", "Compliance alert not found")
		} else {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "alert_retrieval_failed", "Failed to retrieve compliance alert")
		}
		return
	}

	// Log successful request
	h.logger.Info("Compliance alert retrieved successfully", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"alert_id":     alertID,
		"business_id":  alert.BusinessID,
		"framework_id": alert.FrameworkID,
		"alert_type":   alert.AlertType,
		"severity":     alert.Severity,
		"status":       alert.Status,
		"duration_ms":  time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(alert)
}

// ListAlertsHandler handles GET /v1/compliance/alerts
func (h *ComplianceAlertHandler) ListAlertsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse query parameters
	query := &compliance.AlertQuery{
		BusinessID:      r.URL.Query().Get("business_id"),
		FrameworkID:     r.URL.Query().Get("framework_id"),
		AlertType:       r.URL.Query().Get("alert_type"),
		Severity:        r.URL.Query().Get("severity"),
		Status:          r.URL.Query().Get("status"),
		TriggeredBy:     r.URL.Query().Get("triggered_by"),
		IncludeResolved: r.URL.Query().Get("include_resolved") == "true",
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

	h.logger.Info("List alerts request received", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"query":       query,
		"user_agent":  r.UserAgent(),
		"remote_addr": r.RemoteAddr,
	})

	// List alerts
	alerts, err := h.service.ListAlerts(ctx, query)
	if err != nil {
		h.logger.Error("Failed to list compliance alerts", map[string]interface{}{
			"request_id": ctx.Value("request_id"),
			"query":      query,
			"error":      err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "alerts_listing_failed", "Failed to list compliance alerts")
		return
	}

	// Create response
	response := map[string]interface{}{
		"alerts": alerts,
		"pagination": map[string]interface{}{
			"limit":  query.Limit,
			"offset": query.Offset,
			"count":  len(alerts),
		},
		"filters": map[string]interface{}{
			"business_id":      query.BusinessID,
			"framework_id":     query.FrameworkID,
			"alert_type":       query.AlertType,
			"severity":         query.Severity,
			"status":           query.Status,
			"triggered_by":     query.TriggeredBy,
			"include_resolved": query.IncludeResolved,
		},
	}

	// Log successful request
	h.logger.Info("Compliance alerts listed successfully", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"count":       len(alerts),
		"duration_ms": time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateAlertStatusHandler handles PUT /v1/compliance/alerts/{alert_id}/status
func (h *ComplianceAlertHandler) UpdateAlertStatusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract alert_id from URL path
	alertID := h.extractAlertIDFromPath(r.URL.Path)
	if alertID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "alert_id is required")
		return
	}

	// Parse request body
	var request struct {
		Status    string `json:"status"`
		UpdatedBy string `json:"updated_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	// Validate required fields
	if request.Status == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "status is required")
		return
	}
	if request.UpdatedBy == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "updated_by is required")
		return
	}

	// Validate status
	validStatuses := []string{"active", "acknowledged", "resolved", "dismissed"}
	if !h.isValidStatus(request.Status, validStatuses) {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid status. Must be one of: active, acknowledged, resolved, dismissed")
		return
	}

	h.logger.Info("Update alert status request received", map[string]interface{}{
		"request_id": ctx.Value("request_id"),
		"alert_id":   alertID,
		"new_status": request.Status,
		"updated_by": request.UpdatedBy,
	})

	// Update alert status
	err := h.service.UpdateAlertStatus(ctx, alertID, request.Status, request.UpdatedBy)
	if err != nil {
		h.logger.Error("Failed to update alert status", map[string]interface{}{
			"request_id": ctx.Value("request_id"),
			"alert_id":   alertID,
			"new_status": request.Status,
			"updated_by": request.UpdatedBy,
			"error":      err.Error(),
		})
		if err.Error() == "alert not found: "+alertID {
			h.writeErrorResponse(w, r, http.StatusNotFound, "alert_not_found", "Compliance alert not found")
		} else {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "alert_update_failed", "Failed to update alert status")
		}
		return
	}

	// Get updated alert
	alert, err := h.service.GetAlert(ctx, alertID)
	if err != nil {
		h.logger.Error("Failed to get updated alert", map[string]interface{}{
			"request_id": ctx.Value("request_id"),
			"alert_id":   alertID,
			"error":      err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "alert_retrieval_failed", "Failed to retrieve updated alert")
		return
	}

	// Log successful update
	h.logger.Info("Alert status updated successfully", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"alert_id":    alertID,
		"new_status":  request.Status,
		"updated_by":  request.UpdatedBy,
		"duration_ms": time.Since(start).Milliseconds(),
	})

	// Return updated alert
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(alert)
}

// CreateAlertRuleHandler handles POST /v1/compliance/alert-rules
func (h *ComplianceAlertHandler) CreateAlertRuleHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse request body
	var rule compliance.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	// Validate required fields
	if rule.Name == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "name is required")
		return
	}
	if rule.AlertType == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "alert_type is required")
		return
	}
	if rule.Severity == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "severity is required")
		return
	}
	if rule.CreatedBy == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "created_by is required")
		return
	}
	if len(rule.Conditions) == 0 {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "at least one condition is required")
		return
	}

	// Validate alert type
	validAlertTypes := []string{"deadline", "risk_threshold", "compliance_change", "gap_detected", "milestone_overdue"}
	if !h.isValidAlertType(rule.AlertType, validAlertTypes) {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid alert_type. Must be one of: deadline, risk_threshold, compliance_change, gap_detected, milestone_overdue")
		return
	}

	// Validate severity
	validSeverities := []string{"low", "medium", "high", "critical"}
	if !h.isValidSeverity(rule.Severity, validSeverities) {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid severity. Must be one of: low, medium, high, critical")
		return
	}

	// Set default enabled if not provided
	if !rule.Enabled {
		rule.Enabled = true
	}

	h.logger.Info("Create alert rule request received", map[string]interface{}{
		"request_id": ctx.Value("request_id"),
		"name":       rule.Name,
		"alert_type": rule.AlertType,
		"severity":   rule.Severity,
		"created_by": rule.CreatedBy,
		"enabled":    rule.Enabled,
	})

	// Create alert rule
	err := h.service.CreateAlertRule(ctx, &rule)
	if err != nil {
		h.logger.Error("Failed to create alert rule", map[string]interface{}{
			"request_id": ctx.Value("request_id"),
			"name":       rule.Name,
			"alert_type": rule.AlertType,
			"error":      err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "rule_creation_failed", "Failed to create alert rule")
		return
	}

	// Log successful creation
	h.logger.Info("Alert rule created successfully", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"rule_id":     rule.ID,
		"name":        rule.Name,
		"alert_type":  rule.AlertType,
		"severity":    rule.Severity,
		"enabled":     rule.Enabled,
		"duration_ms": time.Since(start).Milliseconds(),
	})

	// Return created rule
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

// EvaluateAlertRulesHandler handles POST /v1/compliance/alerts/evaluate
func (h *ComplianceAlertHandler) EvaluateAlertRulesHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse request body
	var request struct {
		BusinessID  string `json:"business_id"`
		FrameworkID string `json:"framework_id"`
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

	h.logger.Info("Evaluate alert rules request received", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"business_id":  request.BusinessID,
		"framework_id": request.FrameworkID,
	})

	// Evaluate alert rules
	err := h.service.EvaluateAlertRules(ctx, request.BusinessID, request.FrameworkID)
	if err != nil {
		h.logger.Error("Failed to evaluate alert rules", map[string]interface{}{
			"request_id":   ctx.Value("request_id"),
			"business_id":  request.BusinessID,
			"framework_id": request.FrameworkID,
			"error":        err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "rules_evaluation_failed", "Failed to evaluate alert rules")
		return
	}

	// Log successful evaluation
	h.logger.Info("Alert rules evaluated successfully", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"business_id":  request.BusinessID,
		"framework_id": request.FrameworkID,
		"duration_ms":  time.Since(start).Milliseconds(),
	})

	// Return success response
	response := map[string]interface{}{
		"message":      "Alert rules evaluated successfully",
		"business_id":  request.BusinessID,
		"framework_id": request.FrameworkID,
		"evaluated_at": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetNotificationsHandler handles GET /v1/compliance/notifications
func (h *ComplianceAlertHandler) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse query parameters
	query := &compliance.NotificationQuery{
		AlertID:   r.URL.Query().Get("alert_id"),
		Type:      r.URL.Query().Get("type"),
		Recipient: r.URL.Query().Get("recipient"),
		Status:    r.URL.Query().Get("status"),
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

	h.logger.Info("Get notifications request received", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"query":       query,
		"user_agent":  r.UserAgent(),
		"remote_addr": r.RemoteAddr,
	})

	// Get notifications
	notifications, err := h.service.GetNotifications(ctx, query)
	if err != nil {
		h.logger.Error("Failed to get notifications", map[string]interface{}{
			"request_id": ctx.Value("request_id"),
			"query":      query,
			"error":      err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "notifications_retrieval_failed", "Failed to retrieve notifications")
		return
	}

	// Create response
	response := map[string]interface{}{
		"notifications": notifications,
		"pagination": map[string]interface{}{
			"limit":  query.Limit,
			"offset": query.Offset,
			"count":  len(notifications),
		},
		"filters": map[string]interface{}{
			"alert_id":  query.AlertID,
			"type":      query.Type,
			"recipient": query.Recipient,
			"status":    query.Status,
		},
	}

	// Log successful request
	h.logger.Info("Notifications retrieved successfully", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"count":       len(notifications),
		"duration_ms": time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper methods

// extractAlertIDFromPath extracts alert_id from URL path
func (h *ComplianceAlertHandler) extractAlertIDFromPath(path string) string {
	// Expected path format: /v1/compliance/alerts/{alert_id} or /v1/compliance/alerts/{alert_id}/status
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "v1" && parts[1] == "compliance" && parts[2] == "alerts" {
		return parts[3]
	}
	return ""
}

// isValidAlertType checks if alert type is valid
func (h *ComplianceAlertHandler) isValidAlertType(alertType string, validTypes []string) bool {
	for _, validType := range validTypes {
		if alertType == validType {
			return true
		}
	}
	return false
}

// isValidSeverity checks if severity is valid
func (h *ComplianceAlertHandler) isValidSeverity(severity string, validSeverities []string) bool {
	for _, validSeverity := range validSeverities {
		if severity == validSeverity {
			return true
		}
	}
	return false
}

// isValidStatus checks if status is valid
func (h *ComplianceAlertHandler) isValidStatus(status string, validStatuses []string) bool {
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// writeErrorResponse writes an error response
func (h *ComplianceAlertHandler) writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, errorCode, message string) {
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
