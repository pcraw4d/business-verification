package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ComplianceStatusHandler handles compliance status endpoints
type ComplianceStatusHandler struct {
	logger *observability.Logger
}

// NewComplianceStatusHandler creates a new compliance status handler
func NewComplianceStatusHandler(logger *observability.Logger) *ComplianceStatusHandler {
	return &ComplianceStatusHandler{
		logger: logger,
	}
}

// BusinessComplianceStatus represents the compliance status of a business
type BusinessComplianceStatus struct {
	BusinessID        string                      `json:"business_id"`
	OverallStatus     string                      `json:"overall_status"`   // "compliant", "non_compliant", "partial", "pending"
	ComplianceScore   float64                     `json:"compliance_score"` // 0.0 to 1.0
	LastUpdated       time.Time                   `json:"last_updated"`
	Frameworks        []FrameworkStatus           `json:"frameworks"`
	Requirements      []BusinessRequirementStatus `json:"requirements"`
	Alerts            []ComplianceAlert           `json:"alerts"`
	NextReviewDate    *time.Time                  `json:"next_review_date,omitempty"`
	ComplianceHistory []ComplianceHistoryItem     `json:"compliance_history,omitempty"`
}

// FrameworkStatus represents the status of a specific compliance framework
type FrameworkStatus struct {
	FrameworkID   string                      `json:"framework_id"` // "SOC2", "PCI_DSS", "GDPR", "HIPAA", etc.
	FrameworkName string                      `json:"framework_name"`
	Status        string                      `json:"status"` // "compliant", "non_compliant", "partial", "pending"
	Score         float64                     `json:"score"`  // 0.0 to 1.0
	LastAssessed  time.Time                   `json:"last_assessed"`
	Requirements  []BusinessRequirementStatus `json:"requirements"`
}

// BusinessRequirementStatus represents the status of a specific compliance requirement
type BusinessRequirementStatus struct {
	RequirementID   string     `json:"requirement_id"`
	RequirementName string     `json:"requirement_name"`
	Status          string     `json:"status"` // "compliant", "non_compliant", "partial", "pending", "not_applicable"
	Score           float64    `json:"score"`  // 0.0 to 1.0
	LastAssessed    time.Time  `json:"last_assessed"`
	Evidence        []string   `json:"evidence,omitempty"`
	Notes           string     `json:"notes,omitempty"`
	DueDate         *time.Time `json:"due_date,omitempty"`
}

// ComplianceAlert represents a compliance alert or notification
type ComplianceAlert struct {
	AlertID       string     `json:"alert_id"`
	Type          string     `json:"type"`     // "requirement_due", "non_compliance", "review_required", "evidence_expired"
	Severity      string     `json:"severity"` // "low", "medium", "high", "critical"
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	FrameworkID   string     `json:"framework_id,omitempty"`
	RequirementID string     `json:"requirement_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	DueDate       *time.Time `json:"due_date,omitempty"`
	Status        string     `json:"status"` // "active", "acknowledged", "resolved"
}

// ComplianceHistoryItem represents a historical compliance status change
type ComplianceHistoryItem struct {
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
	Score       float64   `json:"score"`
	ChangeType  string    `json:"change_type"` // "assessment", "requirement_update", "framework_change"
	Description string    `json:"description"`
	UpdatedBy   string    `json:"updated_by"`
}

// GetComplianceStatusHandler handles GET /v1/compliance/status/{business_id}
func (h *ComplianceStatusHandler) GetComplianceStatusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractBusinessIDFromPath(r.URL.Path)
	if businessID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Parse query parameters
	includeHistory := r.URL.Query().Get("include_history") == "true"
	includeAlerts := r.URL.Query().Get("include_alerts") == "true"
	frameworkFilter := r.URL.Query().Get("framework")

	h.logger.Info("Compliance status request received", map[string]interface{}{
		"request_id":      ctx.Value("request_id"),
		"business_id":     businessID,
		"include_history": includeHistory,
		"include_alerts":  includeAlerts,
		"framework":       frameworkFilter,
		"user_agent":      r.UserAgent(),
		"remote_addr":     r.RemoteAddr,
	})

	// Get compliance status (mock implementation for now)
	status, err := h.getComplianceStatus(ctx, businessID, includeHistory, includeAlerts, frameworkFilter)
	if err != nil {
		h.logger.Error("Failed to get compliance status", map[string]interface{}{
			"request_id":  ctx.Value("request_id"),
			"business_id": businessID,
			"error":       err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "compliance_status_failed", "Failed to retrieve compliance status")
		return
	}

	// Log successful request
	h.logger.Info("Compliance status retrieved successfully", map[string]interface{}{
		"request_id":       ctx.Value("request_id"),
		"business_id":      businessID,
		"overall_status":   status.OverallStatus,
		"compliance_score": status.ComplianceScore,
		"frameworks_count": len(status.Frameworks),
		"duration_ms":      time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

// UpdateComplianceStatusHandler handles PUT /v1/compliance/status/{business_id}
func (h *ComplianceStatusHandler) UpdateComplianceStatusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractBusinessIDFromPath(r.URL.Path)
	if businessID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Parse request body
	var updateReq struct {
		FrameworkID   string                 `json:"framework_id,omitempty"`
		RequirementID string                 `json:"requirement_id,omitempty"`
		Status        string                 `json:"status,omitempty"`
		Score         *float64               `json:"score,omitempty"`
		Evidence      []string               `json:"evidence,omitempty"`
		Notes         string                 `json:"notes,omitempty"`
		UpdatedBy     string                 `json:"updated_by,omitempty"`
		Metadata      map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	// Validate request
	if updateReq.FrameworkID == "" && updateReq.RequirementID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "framework_id or requirement_id is required")
		return
	}

	if updateReq.Status != "" {
		validStatuses := []string{"compliant", "non_compliant", "partial", "pending", "not_applicable"}
		if !h.isValidStatus(updateReq.Status, validStatuses) {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid status value")
			return
		}
	}

	if updateReq.Score != nil && (*updateReq.Score < 0.0 || *updateReq.Score > 1.0) {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Score must be between 0.0 and 1.0")
		return
	}

	h.logger.Info("Compliance status update request received", map[string]interface{}{
		"request_id":     ctx.Value("request_id"),
		"business_id":    businessID,
		"framework_id":   updateReq.FrameworkID,
		"requirement_id": updateReq.RequirementID,
		"status":         updateReq.Status,
		"score":          updateReq.Score,
		"updated_by":     updateReq.UpdatedBy,
	})

	// Update compliance status (mock implementation for now)
	updatedStatus, err := h.updateComplianceStatus(ctx, businessID, updateReq)
	if err != nil {
		h.logger.Error("Failed to update compliance status", map[string]interface{}{
			"request_id":  ctx.Value("request_id"),
			"business_id": businessID,
			"error":       err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "compliance_update_failed", "Failed to update compliance status")
		return
	}

	// Log successful update
	h.logger.Info("Compliance status updated successfully", map[string]interface{}{
		"request_id":       ctx.Value("request_id"),
		"business_id":      businessID,
		"overall_status":   updatedStatus.OverallStatus,
		"compliance_score": updatedStatus.ComplianceScore,
		"duration_ms":      time.Since(start).Milliseconds(),
	})

	// Return updated status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedStatus)
}

// GetComplianceStatusHistoryHandler handles GET /v1/compliance/status/{business_id}/history
func (h *ComplianceStatusHandler) GetComplianceStatusHistoryHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractBusinessIDFromPath(r.URL.Path)
	if businessID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	frameworkFilter := r.URL.Query().Get("framework")

	// Parse pagination parameters
	limit := 50 // default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	offset := 0 // default offset
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	h.logger.Info("Compliance status history request received", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"business_id": businessID,
		"limit":       limit,
		"offset":      offset,
		"start_date":  startDate,
		"end_date":    endDate,
		"framework":   frameworkFilter,
	})

	// Get compliance status history (mock implementation for now)
	history, totalCount, err := h.getComplianceStatusHistory(ctx, businessID, limit, offset, startDate, endDate, frameworkFilter)
	if err != nil {
		h.logger.Error("Failed to get compliance status history", map[string]interface{}{
			"request_id":  ctx.Value("request_id"),
			"business_id": businessID,
			"error":       err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "compliance_history_failed", "Failed to retrieve compliance status history")
		return
	}

	// Create response
	response := map[string]interface{}{
		"business_id": businessID,
		"history":     history,
		"pagination": map[string]interface{}{
			"limit":       limit,
			"offset":      offset,
			"total_count": totalCount,
			"has_more":    offset+limit < totalCount,
		},
	}

	// Log successful request
	h.logger.Info("Compliance status history retrieved successfully", map[string]interface{}{
		"request_id":    ctx.Value("request_id"),
		"business_id":   businessID,
		"history_count": len(history),
		"total_count":   totalCount,
		"duration_ms":   time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper methods

// extractBusinessIDFromPath extracts business_id from URL path
func (h *ComplianceStatusHandler) extractBusinessIDFromPath(path string) string {
	// Expected path format: /v1/compliance/status/{business_id} or /v1/compliance/status/{business_id}/history
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "v1" && parts[1] == "compliance" && parts[2] == "status" {
		return parts[3]
	}
	return ""
}

// isValidStatus checks if a status is valid
func (h *ComplianceStatusHandler) isValidStatus(status string, validStatuses []string) bool {
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// writeErrorResponse writes an error response
func (h *ComplianceStatusHandler) writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, errorCode, message string) {
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

// Mock implementation methods (to be replaced with actual database/service calls)

// getComplianceStatus retrieves compliance status for a business
func (h *ComplianceStatusHandler) getComplianceStatus(ctx context.Context, businessID string, includeHistory, includeAlerts bool, frameworkFilter string) (*BusinessComplianceStatus, error) {
	// Mock implementation - replace with actual service call
	now := time.Now()

	status := &BusinessComplianceStatus{
		BusinessID:      businessID,
		OverallStatus:   "compliant",
		ComplianceScore: 0.85,
		LastUpdated:     now,
		Frameworks: []FrameworkStatus{
			{
				FrameworkID:   "SOC2",
				FrameworkName: "SOC 2 Type II",
				Status:        "compliant",
				Score:         0.90,
				LastAssessed:  now.AddDate(0, -1, 0),
				Requirements: []BusinessRequirementStatus{
					{
						RequirementID:   "CC6.1",
						RequirementName: "Logical and Physical Access Controls",
						Status:          "compliant",
						Score:           0.95,
						LastAssessed:    now.AddDate(0, -1, 0),
						Evidence:        []string{"Access control policy", "User access review"},
					},
				},
			},
			{
				FrameworkID:   "GDPR",
				FrameworkName: "General Data Protection Regulation",
				Status:        "partial",
				Score:         0.75,
				LastAssessed:  now.AddDate(0, -2, 0),
				Requirements: []BusinessRequirementStatus{
					{
						RequirementID:   "GDPR_32",
						RequirementName: "Security of Processing",
						Status:          "compliant",
						Score:           0.90,
						LastAssessed:    now.AddDate(0, -2, 0),
					},
					{
						RequirementID:   "GDPR_25",
						RequirementName: "Data Protection by Design and by Default",
						Status:          "partial",
						Score:           0.60,
						LastAssessed:    now.AddDate(0, -2, 0),
						Notes:           "Implementation in progress",
					},
				},
			},
		},
		Requirements: []BusinessRequirementStatus{
			{
				RequirementID:   "CC6.1",
				RequirementName: "Logical and Physical Access Controls",
				Status:          "compliant",
				Score:           0.95,
				LastAssessed:    now.AddDate(0, -1, 0),
			},
		},
	}

	if includeAlerts {
		status.Alerts = []ComplianceAlert{
			{
				AlertID:     "alert_001",
				Type:        "review_required",
				Severity:    "medium",
				Title:       "GDPR Review Due",
				Description: "Annual GDPR compliance review is due in 30 days",
				FrameworkID: "GDPR",
				CreatedAt:   now.AddDate(0, 0, -5),
				DueDate:     &[]time.Time{now.AddDate(0, 0, 30)}[0],
				Status:      "active",
			},
		}
	}

	if includeHistory {
		status.ComplianceHistory = []ComplianceHistoryItem{
			{
				Timestamp:   now.AddDate(0, -1, 0),
				Status:      "compliant",
				Score:       0.85,
				ChangeType:  "assessment",
				Description: "Quarterly compliance assessment completed",
				UpdatedBy:   "compliance_team",
			},
			{
				Timestamp:   now.AddDate(0, -2, 0),
				Status:      "partial",
				Score:       0.75,
				ChangeType:  "requirement_update",
				Description: "GDPR requirement updated",
				UpdatedBy:   "compliance_officer",
			},
		}
	}

	// Apply framework filter if specified
	if frameworkFilter != "" {
		filteredFrameworks := []FrameworkStatus{}
		for _, framework := range status.Frameworks {
			if framework.FrameworkID == frameworkFilter {
				filteredFrameworks = append(filteredFrameworks, framework)
			}
		}
		status.Frameworks = filteredFrameworks
	}

	return status, nil
}

// updateComplianceStatus updates compliance status for a business
func (h *ComplianceStatusHandler) updateComplianceStatus(ctx context.Context, businessID string, updateReq interface{}) (*BusinessComplianceStatus, error) {
	// Mock implementation - replace with actual service call
	// For now, return the current status
	return h.getComplianceStatus(ctx, businessID, false, true, "")
}

// getComplianceStatusHistory retrieves compliance status history for a business
func (h *ComplianceStatusHandler) getComplianceStatusHistory(ctx context.Context, businessID string, limit, offset int, startDate, endDate, frameworkFilter string) ([]ComplianceHistoryItem, int, error) {
	// Mock implementation - replace with actual service call
	now := time.Now()

	history := []ComplianceHistoryItem{
		{
			Timestamp:   now.AddDate(0, -1, 0),
			Status:      "compliant",
			Score:       0.85,
			ChangeType:  "assessment",
			Description: "Quarterly compliance assessment completed",
			UpdatedBy:   "compliance_team",
		},
		{
			Timestamp:   now.AddDate(0, -2, 0),
			Status:      "partial",
			Score:       0.75,
			ChangeType:  "requirement_update",
			Description: "GDPR requirement updated",
			UpdatedBy:   "compliance_officer",
		},
		{
			Timestamp:   now.AddDate(0, -3, 0),
			Status:      "non_compliant",
			Score:       0.60,
			ChangeType:  "assessment",
			Description: "Initial compliance assessment",
			UpdatedBy:   "compliance_team",
		},
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start >= len(history) {
		return []ComplianceHistoryItem{}, len(history), nil
	}
	if end > len(history) {
		end = len(history)
	}

	return history[start:end], len(history), nil
}
