package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/compliance"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// ComplianceTrackingHandler handles compliance tracking API endpoints
type ComplianceTrackingHandler struct {
	logger  *observability.Logger
	service *compliance.ComplianceTrackingService
}

// NewComplianceTrackingHandler creates a new compliance tracking handler
func NewComplianceTrackingHandler(logger *observability.Logger, service *compliance.ComplianceTrackingService) *ComplianceTrackingHandler {
	return &ComplianceTrackingHandler{
		logger:  logger,
		service: service,
	}
}

// GetComplianceTrackingHandler handles GET /v1/compliance/tracking/{business_id}/{framework_id}
func (h *ComplianceTrackingHandler) GetComplianceTrackingHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id and framework_id from URL path
	businessID, frameworkID := h.extractTrackingParamsFromPath(r.URL.Path)
	if businessID == "" || frameworkID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "business_id and framework_id are required")
		return
	}

	h.logger.Info("Compliance tracking request received", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"business_id":  businessID,
		"framework_id": frameworkID,
		"user_agent":   r.UserAgent(),
		"remote_addr":  r.RemoteAddr,
	})

	// Get compliance tracking
	tracking, err := h.service.GetComplianceTracking(ctx, businessID, frameworkID)
	if err != nil {
		h.logger.Error("Failed to get compliance tracking", map[string]interface{}{
			"request_id":   ctx.Value("request_id"),
			"business_id":  businessID,
			"framework_id": frameworkID,
			"error":        err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "tracking_retrieval_failed", "Failed to retrieve compliance tracking")
		return
	}

	// Log successful request
	h.logger.Info("Compliance tracking retrieved successfully", map[string]interface{}{
		"request_id":       ctx.Value("request_id"),
		"business_id":      businessID,
		"framework_id":     frameworkID,
		"overall_progress": tracking.OverallProgress,
		"compliance_level": tracking.ComplianceLevel,
		"risk_level":       tracking.RiskLevel,
		"duration_ms":      time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tracking)
}

// UpdateComplianceTrackingHandler handles PUT /v1/compliance/tracking/{business_id}/{framework_id}
func (h *ComplianceTrackingHandler) UpdateComplianceTrackingHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id and framework_id from URL path
	businessID, frameworkID := h.extractTrackingParamsFromPath(r.URL.Path)
	if businessID == "" || frameworkID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "business_id and framework_id are required")
		return
	}

	// Parse request body
	var tracking compliance.ComplianceTracking
	if err := json.NewDecoder(r.Body).Decode(&tracking); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	// Set the IDs from URL
	tracking.BusinessID = businessID
	tracking.FrameworkID = frameworkID

	h.logger.Info("Update compliance tracking request received", map[string]interface{}{
		"request_id":       ctx.Value("request_id"),
		"business_id":      businessID,
		"framework_id":     frameworkID,
		"overall_progress": tracking.OverallProgress,
	})

	// Update compliance tracking
	err := h.service.UpdateComplianceTracking(ctx, &tracking)
	if err != nil {
		h.logger.Error("Failed to update compliance tracking", map[string]interface{}{
			"request_id":   ctx.Value("request_id"),
			"business_id":  businessID,
			"framework_id": frameworkID,
			"error":        err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "tracking_update_failed", "Failed to update compliance tracking")
		return
	}

	// Log successful update
	h.logger.Info("Compliance tracking updated successfully", map[string]interface{}{
		"request_id":       ctx.Value("request_id"),
		"business_id":      businessID,
		"framework_id":     frameworkID,
		"overall_progress": tracking.OverallProgress,
		"compliance_level": tracking.ComplianceLevel,
		"duration_ms":      time.Since(start).Milliseconds(),
	})

	// Return updated tracking
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tracking)
}

// GetComplianceMilestonesHandler handles GET /v1/compliance/milestones
func (h *ComplianceTrackingHandler) GetComplianceMilestonesHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse query parameters
	query := &compliance.MilestoneQuery{
		BusinessID:  r.URL.Query().Get("business_id"),
		FrameworkID: r.URL.Query().Get("framework_id"),
		Status:      r.URL.Query().Get("status"),
		Type:        r.URL.Query().Get("type"),
		Priority:    r.URL.Query().Get("priority"),
		Overdue:     r.URL.Query().Get("overdue") == "true",
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

	h.logger.Info("Compliance milestones request received", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"query":       query,
		"user_agent":  r.UserAgent(),
		"remote_addr": r.RemoteAddr,
	})

	// Get compliance milestones
	milestones, err := h.service.GetComplianceMilestones(ctx, query)
	if err != nil {
		h.logger.Error("Failed to get compliance milestones", map[string]interface{}{
			"request_id": ctx.Value("request_id"),
			"query":      query,
			"error":      err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "milestones_retrieval_failed", "Failed to retrieve compliance milestones")
		return
	}

	// Create response
	response := map[string]interface{}{
		"milestones": milestones,
		"pagination": map[string]interface{}{
			"limit":  query.Limit,
			"offset": query.Offset,
			"count":  len(milestones),
		},
		"filters": map[string]interface{}{
			"business_id":  query.BusinessID,
			"framework_id": query.FrameworkID,
			"status":       query.Status,
			"type":         query.Type,
			"priority":     query.Priority,
			"overdue":      query.Overdue,
		},
	}

	// Log successful request
	h.logger.Info("Compliance milestones retrieved successfully", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"count":       len(milestones),
		"duration_ms": time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateMilestoneHandler handles POST /v1/compliance/milestones
func (h *ComplianceTrackingHandler) CreateMilestoneHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse request body
	var milestone compliance.ComplianceMilestone
	if err := json.NewDecoder(r.Body).Decode(&milestone); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	// Validate required fields
	if milestone.BusinessID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}
	if milestone.FrameworkID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "framework_id is required")
		return
	}
	if milestone.Name == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "name is required")
		return
	}
	if milestone.Type == "" {
		milestone.Type = "assessment" // Default
	}
	if milestone.Priority == "" {
		milestone.Priority = "medium" // Default
	}

	// Generate milestone ID if not provided
	if milestone.ID == "" {
		milestone.ID = h.generateMilestoneID()
	}

	h.logger.Info("Create milestone request received", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"milestone_id": milestone.ID,
		"business_id":  milestone.BusinessID,
		"framework_id": milestone.FrameworkID,
		"name":         milestone.Name,
		"type":         milestone.Type,
	})

	// Create milestone
	err := h.service.CreateMilestone(ctx, &milestone)
	if err != nil {
		h.logger.Error("Failed to create compliance milestone", map[string]interface{}{
			"request_id":   ctx.Value("request_id"),
			"milestone_id": milestone.ID,
			"business_id":  milestone.BusinessID,
			"framework_id": milestone.FrameworkID,
			"error":        err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "milestone_creation_failed", "Failed to create compliance milestone")
		return
	}

	// Log successful creation
	h.logger.Info("Compliance milestone created successfully", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"milestone_id": milestone.ID,
		"business_id":  milestone.BusinessID,
		"framework_id": milestone.FrameworkID,
		"name":         milestone.Name,
		"duration_ms":  time.Since(start).Milliseconds(),
	})

	// Return created milestone
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(milestone)
}

// UpdateMilestoneHandler handles PUT /v1/compliance/milestones/{milestone_id}
func (h *ComplianceTrackingHandler) UpdateMilestoneHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract milestone_id from URL path
	milestoneID := h.extractMilestoneIDFromPath(r.URL.Path)
	if milestoneID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "milestone_id is required")
		return
	}

	// Parse request body
	var milestone compliance.ComplianceMilestone
	if err := json.NewDecoder(r.Body).Decode(&milestone); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	// Set the milestone ID from URL
	milestone.ID = milestoneID

	h.logger.Info("Update milestone request received", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"milestone_id": milestoneID,
		"business_id":  milestone.BusinessID,
		"framework_id": milestone.FrameworkID,
		"status":       milestone.Status,
	})

	// Update milestone
	err := h.service.UpdateMilestone(ctx, &milestone)
	if err != nil {
		h.logger.Error("Failed to update compliance milestone", map[string]interface{}{
			"request_id":   ctx.Value("request_id"),
			"milestone_id": milestoneID,
			"error":        err.Error(),
		})
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, r, http.StatusNotFound, "milestone_not_found", "Compliance milestone not found")
		} else {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "milestone_update_failed", "Failed to update compliance milestone")
		}
		return
	}

	// Log successful update
	h.logger.Info("Compliance milestone updated successfully", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"milestone_id": milestoneID,
		"business_id":  milestone.BusinessID,
		"framework_id": milestone.FrameworkID,
		"status":       milestone.Status,
		"progress":     milestone.Progress,
		"duration_ms":  time.Since(start).Milliseconds(),
	})

	// Return updated milestone
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(milestone)
}

// GetProgressMetricsHandler handles GET /v1/compliance/metrics/{business_id}/{framework_id}
func (h *ComplianceTrackingHandler) GetProgressMetricsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id and framework_id from URL path
	businessID, frameworkID := h.extractTrackingParamsFromPath(r.URL.Path)
	if businessID == "" || frameworkID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "business_id and framework_id are required")
		return
	}

	// Parse period parameter
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly" // Default
	}

	h.logger.Info("Progress metrics request received", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"business_id":  businessID,
		"framework_id": frameworkID,
		"period":       period,
		"user_agent":   r.UserAgent(),
		"remote_addr":  r.RemoteAddr,
	})

	// Get progress metrics
	metrics, err := h.service.GetProgressMetrics(ctx, businessID, frameworkID, period)
	if err != nil {
		h.logger.Error("Failed to get progress metrics", map[string]interface{}{
			"request_id":   ctx.Value("request_id"),
			"business_id":  businessID,
			"framework_id": frameworkID,
			"period":       period,
			"error":        err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "metrics_retrieval_failed", "Failed to retrieve progress metrics")
		return
	}

	// Log successful request
	h.logger.Info("Progress metrics retrieved successfully", map[string]interface{}{
		"request_id":       ctx.Value("request_id"),
		"business_id":      businessID,
		"framework_id":     frameworkID,
		"period":           period,
		"overall_progress": metrics.OverallProgress,
		"velocity":         metrics.AverageVelocity,
		"duration_ms":      time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}

// GetComplianceTrendsHandler handles GET /v1/compliance/trends/{business_id}/{framework_id}
func (h *ComplianceTrackingHandler) GetComplianceTrendsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id and framework_id from URL path
	businessID, frameworkID := h.extractTrackingParamsFromPath(r.URL.Path)
	if businessID == "" || frameworkID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "business_id and framework_id are required")
		return
	}

	// Parse days parameter
	daysStr := r.URL.Query().Get("days")
	days := 30 // Default
	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 && parsedDays <= 365 {
			days = parsedDays
		}
	}

	h.logger.Info("Compliance trends request received", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"business_id":  businessID,
		"framework_id": frameworkID,
		"days":         days,
		"user_agent":   r.UserAgent(),
		"remote_addr":  r.RemoteAddr,
	})

	// Get compliance trends
	trends, err := h.service.GetComplianceTrends(ctx, businessID, frameworkID, days)
	if err != nil {
		h.logger.Error("Failed to get compliance trends", map[string]interface{}{
			"request_id":   ctx.Value("request_id"),
			"business_id":  businessID,
			"framework_id": frameworkID,
			"days":         days,
			"error":        err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "trends_retrieval_failed", "Failed to retrieve compliance trends")
		return
	}

	// Create response
	response := map[string]interface{}{
		"business_id":  businessID,
		"framework_id": frameworkID,
		"days":         days,
		"trends":       trends,
		"count":        len(trends),
	}

	// Log successful request
	h.logger.Info("Compliance trends retrieved successfully", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"business_id":  businessID,
		"framework_id": frameworkID,
		"days":         days,
		"data_points":  len(trends),
		"duration_ms":  time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper methods

// extractTrackingParamsFromPath extracts business_id and framework_id from URL path
func (h *ComplianceTrackingHandler) extractTrackingParamsFromPath(path string) (string, string) {
	// Expected path format: /v1/compliance/tracking/{business_id}/{framework_id} or /v1/compliance/metrics/{business_id}/{framework_id} or /v1/compliance/trends/{business_id}/{framework_id}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 5 && parts[0] == "v1" && parts[1] == "compliance" {
		if (parts[2] == "tracking" || parts[2] == "metrics" || parts[2] == "trends") && len(parts) >= 5 {
			return parts[3], parts[4]
		}
	}
	return "", ""
}

// extractMilestoneIDFromPath extracts milestone_id from URL path
func (h *ComplianceTrackingHandler) extractMilestoneIDFromPath(path string) string {
	// Expected path format: /v1/compliance/milestones/{milestone_id}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "v1" && parts[1] == "compliance" && parts[2] == "milestones" {
		return parts[3]
	}
	return ""
}

// generateMilestoneID generates a unique milestone ID
func (h *ComplianceTrackingHandler) generateMilestoneID() string {
	return "milestone_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// writeErrorResponse writes an error response
func (h *ComplianceTrackingHandler) writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, errorCode, message string) {
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
