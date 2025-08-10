package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/compliance"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// AuditHandler handles compliance audit logging operations
type AuditHandler struct {
	auditSystem *compliance.ComplianceAuditSystem
	logger      *observability.Logger
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditSystem *compliance.ComplianceAuditSystem, logger *observability.Logger) *AuditHandler {
	return &AuditHandler{
		auditSystem: auditSystem,
		logger:      logger,
	}
}

// RecordAuditEventRequest represents a request to record an audit event
type RecordAuditEventRequest struct {
	BusinessID    string                 `json:"business_id" validate:"required"`
	EventType     string                 `json:"event_type" validate:"required"`
	EventCategory string                 `json:"event_category" validate:"required"`
	EntityType    string                 `json:"entity_type" validate:"required"`
	EntityID      string                 `json:"entity_id" validate:"required"`
	Action        compliance.AuditAction `json:"action" validate:"required"`
	Description   string                 `json:"description" validate:"required"`
	UserID        string                 `json:"user_id" validate:"required"`
	UserName      string                 `json:"user_name" validate:"required"`
	UserRole      string                 `json:"user_role" validate:"required"`
	UserEmail     string                 `json:"user_email" validate:"required,email"`
	IPAddress     string                 `json:"ip_address"`
	UserAgent     string                 `json:"user_agent"`
	SessionID     string                 `json:"session_id"`
	RequestID     string                 `json:"request_id"`
	OldValue      interface{}            `json:"old_value,omitempty"`
	NewValue      interface{}            `json:"new_value,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Severity      string                 `json:"severity" validate:"required,oneof=low medium high critical"`
	Impact        string                 `json:"impact" validate:"required,oneof=minimal low medium high critical"`
	Tags          []string               `json:"tags,omitempty"`
}

// RecordAuditEventResponse represents the response from recording an audit event
type RecordAuditEventResponse struct {
	EventID   string    `json:"event_id"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
}

// GetAuditEventsRequest represents a request to get audit events
type GetAuditEventsRequest struct {
	BusinessID      string    `json:"business_id" validate:"required"`
	EventTypes      []string  `json:"event_types,omitempty"`
	EventCategories []string  `json:"event_categories,omitempty"`
	EntityTypes     []string  `json:"entity_types,omitempty"`
	EntityIDs       []string  `json:"entity_ids,omitempty"`
	Actions         []string  `json:"actions,omitempty"`
	UserIDs         []string  `json:"user_ids,omitempty"`
	UserRoles       []string  `json:"user_roles,omitempty"`
	Severities      []string  `json:"severities,omitempty"`
	Impacts         []string  `json:"impacts,omitempty"`
	Success         *bool     `json:"success,omitempty"`
	StartDate       time.Time `json:"start_date,omitempty"`
	EndDate         time.Time `json:"end_date,omitempty"`
	Tags            []string  `json:"tags,omitempty"`
	Limit           int       `json:"limit,omitempty"`
	Offset          int       `json:"offset,omitempty"`
	SortBy          string    `json:"sort_by,omitempty"`
	SortOrder       string    `json:"sort_order,omitempty"`
}

// GetAuditEventsResponse represents the response from getting audit events
type GetAuditEventsResponse struct {
	Events []compliance.AuditEvent `json:"events"`
	Meta   struct {
		Total   int  `json:"total"`
		Limit   int  `json:"limit"`
		Offset  int  `json:"offset"`
		HasMore bool `json:"has_more"`
	} `json:"meta"`
}

// GetAuditTrailRequest represents a request to get audit trail
type GetAuditTrailRequest struct {
	BusinessID string    `json:"business_id" validate:"required"`
	StartDate  time.Time `json:"start_date" validate:"required"`
	EndDate    time.Time `json:"end_date" validate:"required"`
}

// GetAuditTrailResponse represents the response from getting audit trail
type GetAuditTrailResponse struct {
	AuditTrail []compliance.ComplianceAuditTrail `json:"audit_trail"`
	Meta       struct {
		TotalEntries int           `json:"total_entries"`
		StartDate    time.Time     `json:"start_date"`
		EndDate      time.Time     `json:"end_date"`
		Duration     time.Duration `json:"duration"`
	} `json:"meta"`
}

// GenerateAuditReportRequest represents a request to generate an audit report
type GenerateAuditReportRequest struct {
	BusinessID string    `json:"business_id" validate:"required"`
	ReportType string    `json:"report_type" validate:"required,oneof=summary detailed activity changes exceptions remediations"`
	StartDate  time.Time `json:"start_date" validate:"required"`
	EndDate    time.Time `json:"end_date" validate:"required"`
}

// GenerateAuditReportResponse represents the response from generating an audit report
type GenerateAuditReportResponse struct {
	ReportID    string                    `json:"report_id"`
	ReportType  string                    `json:"report_type"`
	GeneratedAt time.Time                 `json:"generated_at"`
	GeneratedBy string                    `json:"generated_by"`
	Period      string                    `json:"period"`
	StartDate   time.Time                 `json:"start_date"`
	EndDate     time.Time                 `json:"end_date"`
	TotalEvents int                       `json:"total_events"`
	Summary     compliance.AuditSummary   `json:"summary"`
	Trends      compliance.AuditTrends    `json:"trends"`
	Anomalies   []compliance.AuditAnomaly `json:"anomalies"`
}

// GetAuditMetricsRequest represents a request to get audit metrics
type GetAuditMetricsRequest struct {
	BusinessID string `json:"business_id" validate:"required"`
}

// GetAuditMetricsResponse represents the response from getting audit metrics
type GetAuditMetricsResponse struct {
	Metrics *compliance.AuditMetrics `json:"metrics"`
}

// RecordAuditEvent handles recording a new audit event
func (h *AuditHandler) RecordAuditEvent(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse request body
	var req RecordAuditEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, time.Since(start))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.BusinessID == "" || req.EventType == "" || req.EventCategory == "" || req.EntityType == "" || req.EntityID == "" || req.Description == "" || req.UserID == "" || req.UserName == "" || req.UserRole == "" || req.UserEmail == "" || req.Severity == "" || req.Impact == "" {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, time.Since(start))
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create audit event
	event := &compliance.AuditEvent{
		ID:            generateID(),
		BusinessID:    req.BusinessID,
		EventType:     req.EventType,
		EventCategory: req.EventCategory,
		EntityType:    req.EntityType,
		EntityID:      req.EntityID,
		Action:        req.Action,
		Description:   req.Description,
		UserID:        req.UserID,
		UserName:      req.UserName,
		UserRole:      req.UserRole,
		UserEmail:     req.UserEmail,
		IPAddress:     req.IPAddress,
		UserAgent:     req.UserAgent,
		SessionID:     req.SessionID,
		RequestID:     req.RequestID,
		Timestamp:     time.Now(),
		Success:       true,
		OldValue:      req.OldValue,
		NewValue:      req.NewValue,
		Metadata:      req.Metadata,
		Severity:      req.Severity,
		Impact:        req.Impact,
		Tags:          req.Tags,
	}

	// Record the audit event
	if err := h.auditSystem.RecordAuditEvent(ctx, event); err != nil {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusInternalServerError, time.Since(start))
		http.Error(w, "Failed to record audit event", http.StatusInternalServerError)
		return
	}

	// Return response
	response := RecordAuditEventResponse{
		EventID:   event.ID,
		Timestamp: event.Timestamp,
		Success:   true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusCreated, time.Since(start))
}

// GetAuditEvents handles retrieving audit events with filtering
func (h *AuditHandler) GetAuditEvents(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse query parameters
	businessID := r.URL.Query().Get("business_id")
	if businessID == "" {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, time.Since(start))
		http.Error(w, "business_id is required", http.StatusBadRequest)
		return
	}

	// Parse filter parameters
	filter := &compliance.AuditFilter{
		BusinessID: businessID,
	}

	// Parse event types
	if eventTypes := r.URL.Query().Get("event_types"); eventTypes != "" {
		filter.EventTypes = parseCommaSeparated(eventTypes)
	}

	// Parse event categories
	if eventCategories := r.URL.Query().Get("event_categories"); eventCategories != "" {
		filter.EventCategories = parseCommaSeparated(eventCategories)
	}

	// Parse entity types
	if entityTypes := r.URL.Query().Get("entity_types"); entityTypes != "" {
		filter.EntityTypes = parseCommaSeparated(entityTypes)
	}

	// Parse entity IDs
	if entityIDs := r.URL.Query().Get("entity_ids"); entityIDs != "" {
		filter.EntityIDs = parseCommaSeparated(entityIDs)
	}

	// Parse actions
	if actions := r.URL.Query().Get("actions"); actions != "" {
		actionStrings := parseCommaSeparated(actions)
		for _, actionStr := range actionStrings {
			filter.Actions = append(filter.Actions, compliance.AuditAction(actionStr))
		}
	}

	// Parse user IDs
	if userIDs := r.URL.Query().Get("user_ids"); userIDs != "" {
		filter.UserIDs = parseCommaSeparated(userIDs)
	}

	// Parse user roles
	if userRoles := r.URL.Query().Get("user_roles"); userRoles != "" {
		filter.UserRoles = parseCommaSeparated(userRoles)
	}

	// Parse severities
	if severities := r.URL.Query().Get("severities"); severities != "" {
		filter.Severities = parseCommaSeparated(severities)
	}

	// Parse impacts
	if impacts := r.URL.Query().Get("impacts"); impacts != "" {
		filter.Impacts = parseCommaSeparated(impacts)
	}

	// Parse success filter
	if successStr := r.URL.Query().Get("success"); successStr != "" {
		success, err := strconv.ParseBool(successStr)
		if err == nil {
			filter.Success = &success
		}
	}

	// Parse date range
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	// Parse tags
	if tags := r.URL.Query().Get("tags"); tags != "" {
		filter.Tags = parseCommaSeparated(tags)
	}

	// Parse pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	} else {
		filter.Limit = 50 // Default limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// Parse sorting
	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	} else {
		filter.SortBy = "timestamp"
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	} else {
		filter.SortOrder = "desc"
	}

	// Get audit events
	events, err := h.auditSystem.GetAuditEvents(ctx, businessID, filter)
	if err != nil {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusInternalServerError, time.Since(start))
		http.Error(w, "Failed to get audit events", http.StatusInternalServerError)
		return
	}

	// Calculate metadata
	total := len(events)
	hasMore := total == filter.Limit

	// Return response
	response := GetAuditEventsResponse{
		Events: events,
		Meta: struct {
			Total   int  `json:"total"`
			Limit   int  `json:"limit"`
			Offset  int  `json:"offset"`
			HasMore bool `json:"has_more"`
		}{
			Total:   total,
			Limit:   filter.Limit,
			Offset:  filter.Offset,
			HasMore: hasMore,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
}

// GetAuditTrail handles retrieving audit trail for a business
func (h *AuditHandler) GetAuditTrail(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse query parameters
	businessID := r.URL.Query().Get("business_id")
	if businessID == "" {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, time.Since(start))
		http.Error(w, "business_id is required", http.StatusBadRequest)
		return
	}

	startDateStr := r.URL.Query().Get("start_date")
	if startDateStr == "" {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, time.Since(start))
		http.Error(w, "start_date is required", http.StatusBadRequest)
		return
	}

	endDateStr := r.URL.Query().Get("end_date")
	if endDateStr == "" {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, time.Since(start))
		http.Error(w, "end_date is required", http.StatusBadRequest)
		return
	}

	// Parse dates
	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, time.Since(start))
		http.Error(w, "Invalid start_date format", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, time.Since(start))
		http.Error(w, "Invalid end_date format", http.StatusBadRequest)
		return
	}

	// Get audit trail
	auditTrail, err := h.auditSystem.GetAuditTrail(ctx, businessID, startDate, endDate)
	if err != nil {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusInternalServerError, time.Since(start))
		http.Error(w, "Failed to get audit trail", http.StatusInternalServerError)
		return
	}

	// Calculate metadata
	totalEntries := len(auditTrail)
	duration := endDate.Sub(startDate)

	// Return response
	response := GetAuditTrailResponse{
		AuditTrail: auditTrail,
		Meta: struct {
			TotalEntries int           `json:"total_entries"`
			StartDate    time.Time     `json:"start_date"`
			EndDate      time.Time     `json:"end_date"`
			Duration     time.Duration `json:"duration"`
		}{
			TotalEntries: totalEntries,
			StartDate:    startDate,
			EndDate:      endDate,
			Duration:     duration,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
}

// GenerateAuditReport handles generating an audit report
func (h *AuditHandler) GenerateAuditReport(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse request body
	var req GenerateAuditReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, time.Since(start))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.BusinessID == "" || req.ReportType == "" {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, time.Since(start))
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Generate audit report
	report, err := h.auditSystem.GenerateAuditReport(ctx, req.BusinessID, req.ReportType, req.StartDate, req.EndDate)
	if err != nil {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusInternalServerError, time.Since(start))
		http.Error(w, "Failed to generate audit report", http.StatusInternalServerError)
		return
	}

	// Return response
	response := GenerateAuditReportResponse{
		ReportID:    report.ID,
		ReportType:  report.ReportType,
		GeneratedAt: report.GeneratedAt,
		GeneratedBy: report.GeneratedBy,
		Period:      report.Period,
		StartDate:   report.StartDate,
		EndDate:     report.EndDate,
		TotalEvents: report.TotalEvents,
		Summary:     report.Summary,
		Trends:      report.Trends,
		Anomalies:   report.Anomalies,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusCreated, time.Since(start))
}

// GetAuditMetrics handles retrieving audit metrics
func (h *AuditHandler) GetAuditMetrics(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, time.Since(start))
		http.Error(w, "business_id is required", http.StatusBadRequest)
		return
	}
	businessID := pathParts[len(pathParts)-1]
	if businessID == "" {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, time.Since(start))
		http.Error(w, "business_id is required", http.StatusBadRequest)
		return
	}

	// Get audit metrics
	metrics, err := h.auditSystem.GetAuditMetrics(ctx, businessID)
	if err != nil {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusInternalServerError, time.Since(start))
		http.Error(w, "Failed to get audit metrics", http.StatusInternalServerError)
		return
	}

	// Return response
	response := GetAuditMetricsResponse{
		Metrics: metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
}

// UpdateAuditMetrics handles updating audit metrics
func (h *AuditHandler) UpdateAuditMetrics(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, time.Since(start))
		http.Error(w, "business_id is required", http.StatusBadRequest)
		return
	}
	businessID := pathParts[len(pathParts)-1]
	if businessID == "" {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusBadRequest, time.Since(start))
		http.Error(w, "business_id is required", http.StatusBadRequest)
		return
	}

	// Update audit metrics
	if err := h.auditSystem.UpdateAuditMetrics(ctx, businessID); err != nil {
		h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusInternalServerError, time.Since(start))
		http.Error(w, "Failed to update audit metrics", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Audit metrics updated successfully",
	})

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
}

// Helper function to parse comma-separated values
func parseCommaSeparated(value string) []string {
	if value == "" {
		return nil
	}
	// Split by comma and trim whitespace
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Helper function to generate unique IDs
func generateID() string {
	return fmt.Sprintf("audit-%d", time.Now().UnixNano())
}
