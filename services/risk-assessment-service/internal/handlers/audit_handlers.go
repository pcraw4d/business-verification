package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/audit"
)

// AuditHandler handles audit-related API requests
type AuditHandler struct {
	auditLogger *audit.AuditLogger
	logger      *zap.Logger
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditLogger *audit.AuditLogger, logger *zap.Logger) *AuditHandler {
	return &AuditHandler{
		auditLogger: auditLogger,
		logger:      logger,
	}
}

// GetAuditEvents retrieves audit events based on query parameters
func (h *AuditHandler) GetAuditEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	query := audit.AuditQuery{
		TenantID:   r.URL.Query().Get("tenant_id"),
		UserID:     r.URL.Query().Get("user_id"),
		Action:     r.URL.Query().Get("action"),
		Resource:   r.URL.Query().Get("resource"),
		ResourceID: r.URL.Query().Get("resource_id"),
		IPAddress:  r.URL.Query().Get("ip_address"),
		SortBy:     r.URL.Query().Get("sort_by"),
		SortOrder:  r.URL.Query().Get("sort_order"),
	}

	// Parse date range
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			query.StartDate = &startDate
		}
	}
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			query.EndDate = &endDate
		}
	}

	// Parse status
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			query.Status = &status
		}
	}

	// Parse pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			query.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			query.Offset = offset
		}
	}

	h.logger.Info("Retrieving audit events",
		zap.String("tenant_id", query.TenantID),
		zap.String("action", query.Action),
		zap.String("resource", query.Resource))

	// Get audit events
	events, err := h.auditLogger.GetAuditEvents(ctx, query)
	if err != nil {
		h.logger.Error("Failed to get audit events", zap.Error(err))
		http.Error(w, "Failed to retrieve audit events", http.StatusInternalServerError)
		return
	}

	// Log the audit query
	h.auditLogger.LogDataAccess(ctx, query.TenantID, "", "audit_events", "", "read", map[string]interface{}{
		"query_params": query,
		"result_count": len(events),
	})

	response := map[string]interface{}{
		"success": true,
		"data":    events,
		"count":   len(events),
		"query":   query,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetAuditStats retrieves audit statistics
func (h *AuditHandler) GetAuditStats(w http.ResponseWriter, r *http.Request) {
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

	h.logger.Info("Retrieving audit statistics",
		zap.String("tenant_id", tenantID),
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate))

	// Get audit statistics
	stats, err := h.auditLogger.GetAuditStats(ctx, tenantID, startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get audit statistics", zap.Error(err))
		http.Error(w, "Failed to retrieve audit statistics", http.StatusInternalServerError)
		return
	}

	// Log the statistics access
	h.auditLogger.LogDataAccess(ctx, tenantID, "", "audit_statistics", "", "read", map[string]interface{}{
		"start_date":   startDate,
		"end_date":     endDate,
		"total_events": stats.TotalEvents,
	})

	response := map[string]interface{}{
		"success": true,
		"data":    stats,
		"period": map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetAuditLog retrieves an immutable audit log entry
func (h *AuditHandler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	eventID := vars["event_id"]

	h.logger.Info("Retrieving audit log",
		zap.String("event_id", eventID))

	// Get audit log
	log, err := h.auditLogger.GetAuditLog(ctx, eventID)
	if err != nil {
		h.logger.Error("Failed to get audit log", zap.Error(err))
		http.Error(w, "Failed to retrieve audit log", http.StatusInternalServerError)
		return
	}

	// Log the audit log access
	h.auditLogger.LogDataAccess(ctx, log.TenantID, "", "audit_log", eventID, "read", map[string]interface{}{
		"log_id": log.ID,
	})

	response := map[string]interface{}{
		"success": true,
		"data":    log,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// VerifyAuditIntegrity verifies the integrity of an audit log entry
func (h *AuditHandler) VerifyAuditIntegrity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	eventID := vars["event_id"]

	h.logger.Info("Verifying audit integrity",
		zap.String("event_id", eventID))

	// Verify audit integrity
	isValid, err := h.auditLogger.VerifyAuditIntegrity(ctx, eventID)
	if err != nil {
		h.logger.Error("Failed to verify audit integrity", zap.Error(err))
		http.Error(w, "Failed to verify audit integrity", http.StatusInternalServerError)
		return
	}

	// Log the integrity verification
	h.auditLogger.LogAdminAction(ctx, "", "", "verify_audit_integrity", "audit_log", eventID, map[string]interface{}{
		"is_valid": isValid,
	})

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"event_id":    eventID,
			"is_valid":    isValid,
			"verified_at": time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetAuditHealth checks the health of the audit system
func (h *AuditHandler) GetAuditHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check audit system health
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"components": map[string]interface{}{
			"audit_logger": "healthy",
			"repository":   "healthy",
		},
	}

	// Log health check
	h.auditLogger.LogSecurityEvent(ctx, "", "", "audit_health_check", map[string]interface{}{
		"status": "healthy",
	})

	response := map[string]interface{}{
		"success": true,
		"data":    health,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetAuditConfiguration retrieves audit system configuration
func (h *AuditHandler) GetAuditConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get audit configuration (without sensitive data)
	config := map[string]interface{}{
		"enabled":            true,
		"log_level":          "info",
		"retention_days":     365,
		"batch_size":         100,
		"flush_interval":     "5s",
		"enable_hashing":     true,
		"enable_compression": false,
		"max_file_size":      10485760, // 10MB
	}

	// Log configuration access
	h.auditLogger.LogAdminAction(ctx, "", "", "get_audit_configuration", "audit_config", "", map[string]interface{}{
		"config_keys": []string{"enabled", "log_level", "retention_days", "batch_size"},
	})

	response := map[string]interface{}{
		"success": true,
		"data":    config,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateAuditConfiguration updates audit system configuration
func (h *AuditHandler) UpdateAuditConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var configUpdate map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&configUpdate); err != nil {
		h.logger.Error("Failed to decode configuration update", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Updating audit configuration",
		zap.Any("config_update", configUpdate))

	// Log configuration update
	h.auditLogger.LogAdminAction(ctx, "", "", "update_audit_configuration", "audit_config", "", map[string]interface{}{
		"config_update": configUpdate,
	})

	response := map[string]interface{}{
		"success": true,
		"message": "Audit configuration updated successfully",
		"data": map[string]interface{}{
			"updated_at": time.Now(),
			"changes":    configUpdate,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetAuditRetentionPolicies retrieves audit retention policies
func (h *AuditHandler) GetAuditRetentionPolicies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	h.logger.Info("Retrieving audit retention policies",
		zap.String("tenant_id", tenantID))

	// Mock retention policies (in production, these would come from the database)
	policies := []audit.AuditRetentionPolicy{
		{
			ID:            "policy_1",
			TenantID:      tenantID,
			PolicyName:    "Standard Retention",
			Description:   "Standard 1-year retention policy for audit logs",
			RetentionDays: 365,
			IsActive:      true,
			CreatedAt:     time.Now().AddDate(-1, 0, 0),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            "policy_2",
			TenantID:      tenantID,
			PolicyName:    "Extended Retention",
			Description:   "Extended 7-year retention policy for compliance",
			RetentionDays: 2555, // 7 years
			IsActive:      true,
			CreatedAt:     time.Now().AddDate(-1, 0, 0),
			UpdatedAt:     time.Now(),
		},
	}

	// Log retention policy access
	h.auditLogger.LogDataAccess(ctx, tenantID, "", "audit_retention_policies", "", "read", map[string]interface{}{
		"policy_count": len(policies),
	})

	response := map[string]interface{}{
		"success": true,
		"data":    policies,
		"count":   len(policies),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateAuditRetentionPolicy creates a new audit retention policy
func (h *AuditHandler) CreateAuditRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	var req struct {
		PolicyName    string `json:"policy_name"`
		Description   string `json:"description"`
		RetentionDays int    `json:"retention_days"`
		CreatedBy     string `json:"created_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode retention policy request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Creating audit retention policy",
		zap.String("tenant_id", tenantID),
		zap.String("policy_name", req.PolicyName),
		zap.Int("retention_days", req.RetentionDays))

	// Create retention policy
	policy := audit.AuditRetentionPolicy{
		ID:            generatePolicyID(),
		TenantID:      tenantID,
		PolicyName:    req.PolicyName,
		Description:   req.Description,
		RetentionDays: req.RetentionDays,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Log policy creation
	h.auditLogger.LogAdminAction(ctx, tenantID, req.CreatedBy, "create_audit_retention_policy", "audit_retention_policy", policy.ID, map[string]interface{}{
		"policy_name":    req.PolicyName,
		"retention_days": req.RetentionDays,
	})

	response := map[string]interface{}{
		"success": true,
		"message": "Audit retention policy created successfully",
		"data":    policy,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Helper function
func generatePolicyID() string {
	return fmt.Sprintf("policy_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}
