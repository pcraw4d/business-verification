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

// ComplianceFrameworkHandler handles compliance framework API endpoints
type ComplianceFrameworkHandler struct {
	logger  *observability.Logger
	service *compliance.ComplianceFrameworkService
}

// NewComplianceFrameworkHandler creates a new compliance framework handler
func NewComplianceFrameworkHandler(logger *observability.Logger, service *compliance.ComplianceFrameworkService) *ComplianceFrameworkHandler {
	return &ComplianceFrameworkHandler{
		logger:  logger,
		service: service,
	}
}

// GetFrameworksHandler handles GET /v1/compliance/frameworks
func (h *ComplianceFrameworkHandler) GetFrameworksHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse query parameters
	query := &compliance.FrameworkQuery{
		Category:        r.URL.Query().Get("category"),
		Status:          r.URL.Query().Get("status"),
		BusinessType:    r.URL.Query().Get("business_type"),
		IncludeInactive: r.URL.Query().Get("include_inactive") == "true",
	}

	// Parse jurisdiction filter
	if jurisdictionStr := r.URL.Query().Get("jurisdiction"); jurisdictionStr != "" {
		query.Jurisdiction = strings.Split(jurisdictionStr, ",")
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

	h.logger.Info("Compliance frameworks request received", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"query":       query,
		"user_agent":  r.UserAgent(),
		"remote_addr": r.RemoteAddr,
	})

	// Get frameworks
	frameworks, err := h.service.GetFrameworks(ctx, query)
	if err != nil {
		h.logger.Error("Failed to get compliance frameworks", map[string]interface{}{
			"request_id": ctx.Value("request_id"),
			"error":      err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "frameworks_retrieval_failed", "Failed to retrieve compliance frameworks")
		return
	}

	// Create response
	response := map[string]interface{}{
		"frameworks": frameworks,
		"pagination": map[string]interface{}{
			"limit":  query.Limit,
			"offset": query.Offset,
			"count":  len(frameworks),
		},
		"filters": map[string]interface{}{
			"category":         query.Category,
			"status":           query.Status,
			"business_type":    query.BusinessType,
			"jurisdiction":     query.Jurisdiction,
			"include_inactive": query.IncludeInactive,
		},
	}

	// Log successful request
	h.logger.Info("Compliance frameworks retrieved successfully", map[string]interface{}{
		"request_id":  ctx.Value("request_id"),
		"count":       len(frameworks),
		"duration_ms": time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetFrameworkHandler handles GET /v1/compliance/frameworks/{framework_id}
func (h *ComplianceFrameworkHandler) GetFrameworkHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract framework_id from URL path
	frameworkID := h.extractFrameworkIDFromPath(r.URL.Path)
	if frameworkID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "framework_id is required")
		return
	}

	h.logger.Info("Compliance framework request received", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"framework_id": frameworkID,
		"user_agent":   r.UserAgent(),
		"remote_addr":  r.RemoteAddr,
	})

	// Get framework
	framework, err := h.service.GetFramework(ctx, frameworkID)
	if err != nil {
		h.logger.Error("Failed to get compliance framework", map[string]interface{}{
			"request_id":   ctx.Value("request_id"),
			"framework_id": frameworkID,
			"error":        err.Error(),
		})
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, r, http.StatusNotFound, "framework_not_found", "Compliance framework not found")
		} else {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "framework_retrieval_failed", "Failed to retrieve compliance framework")
		}
		return
	}

	// Log successful request
	h.logger.Info("Compliance framework retrieved successfully", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"framework_id": frameworkID,
		"name":         framework.Name,
		"category":     framework.Category,
		"duration_ms":  time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(framework)
}

// GetFrameworkRequirementsHandler handles GET /v1/compliance/frameworks/{framework_id}/requirements
func (h *ComplianceFrameworkHandler) GetFrameworkRequirementsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract framework_id from URL path
	frameworkID := h.extractFrameworkIDFromPath(r.URL.Path)
	if frameworkID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "framework_id is required")
		return
	}

	h.logger.Info("Framework requirements request received", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"framework_id": frameworkID,
		"user_agent":   r.UserAgent(),
		"remote_addr":  r.RemoteAddr,
	})

	// Get framework requirements
	requirements, err := h.service.GetFrameworkRequirements(ctx, frameworkID)
	if err != nil {
		h.logger.Error("Failed to get framework requirements", map[string]interface{}{
			"request_id":   ctx.Value("request_id"),
			"framework_id": frameworkID,
			"error":        err.Error(),
		})
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, r, http.StatusNotFound, "framework_not_found", "Compliance framework not found")
		} else {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "requirements_retrieval_failed", "Failed to retrieve framework requirements")
		}
		return
	}

	// Create response
	response := map[string]interface{}{
		"framework_id": frameworkID,
		"requirements": requirements,
		"count":        len(requirements),
	}

	// Log successful request
	h.logger.Info("Framework requirements retrieved successfully", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"framework_id": frameworkID,
		"count":        len(requirements),
		"duration_ms":  time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateAssessmentHandler handles POST /v1/compliance/assessments
func (h *ComplianceFrameworkHandler) CreateAssessmentHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse request body
	var assessment compliance.ComplianceAssessment
	if err := json.NewDecoder(r.Body).Decode(&assessment); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	// Validate required fields
	if assessment.BusinessID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}
	if assessment.FrameworkID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "framework_id is required")
		return
	}
	if assessment.AssessmentType == "" {
		assessment.AssessmentType = "periodic" // Default
	}
	if assessment.Status == "" {
		assessment.Status = "planned" // Default
	}
	if assessment.Assessor == "" {
		assessment.Assessor = "system" // Default
	}

	// Generate assessment ID if not provided
	if assessment.ID == "" {
		assessment.ID = h.generateAssessmentID()
	}

	h.logger.Info("Create assessment request received", map[string]interface{}{
		"request_id":      ctx.Value("request_id"),
		"assessment_id":   assessment.ID,
		"business_id":     assessment.BusinessID,
		"framework_id":    assessment.FrameworkID,
		"assessment_type": assessment.AssessmentType,
	})

	// Create assessment
	err := h.service.CreateAssessment(ctx, &assessment)
	if err != nil {
		h.logger.Error("Failed to create compliance assessment", map[string]interface{}{
			"request_id":    ctx.Value("request_id"),
			"assessment_id": assessment.ID,
			"business_id":   assessment.BusinessID,
			"framework_id":  assessment.FrameworkID,
			"error":         err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "assessment_creation_failed", "Failed to create compliance assessment")
		return
	}

	// Log successful creation
	h.logger.Info("Compliance assessment created successfully", map[string]interface{}{
		"request_id":    ctx.Value("request_id"),
		"assessment_id": assessment.ID,
		"business_id":   assessment.BusinessID,
		"framework_id":  assessment.FrameworkID,
		"duration_ms":   time.Since(start).Milliseconds(),
	})

	// Return created assessment
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(assessment)
}

// GetAssessmentHandler handles GET /v1/compliance/assessments/{assessment_id}
func (h *ComplianceFrameworkHandler) GetAssessmentHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract assessment_id from URL path
	assessmentID := h.extractAssessmentIDFromPath(r.URL.Path)
	if assessmentID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "assessment_id is required")
		return
	}

	h.logger.Info("Assessment request received", map[string]interface{}{
		"request_id":    ctx.Value("request_id"),
		"assessment_id": assessmentID,
		"user_agent":    r.UserAgent(),
		"remote_addr":   r.RemoteAddr,
	})

	// Get assessment
	assessment, err := h.service.GetAssessment(ctx, assessmentID)
	if err != nil {
		h.logger.Error("Failed to get compliance assessment", map[string]interface{}{
			"request_id":    ctx.Value("request_id"),
			"assessment_id": assessmentID,
			"error":         err.Error(),
		})
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, r, http.StatusNotFound, "assessment_not_found", "Compliance assessment not found")
		} else {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "assessment_retrieval_failed", "Failed to retrieve compliance assessment")
		}
		return
	}

	// Log successful request
	h.logger.Info("Compliance assessment retrieved successfully", map[string]interface{}{
		"request_id":    ctx.Value("request_id"),
		"assessment_id": assessmentID,
		"business_id":   assessment.BusinessID,
		"framework_id":  assessment.FrameworkID,
		"status":        assessment.Status,
		"duration_ms":   time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(assessment)
}

// UpdateAssessmentHandler handles PUT /v1/compliance/assessments/{assessment_id}
func (h *ComplianceFrameworkHandler) UpdateAssessmentHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract assessment_id from URL path
	assessmentID := h.extractAssessmentIDFromPath(r.URL.Path)
	if assessmentID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "assessment_id is required")
		return
	}

	// Parse request body
	var assessment compliance.ComplianceAssessment
	if err := json.NewDecoder(r.Body).Decode(&assessment); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	// Set the assessment ID from URL
	assessment.ID = assessmentID

	h.logger.Info("Update assessment request received", map[string]interface{}{
		"request_id":    ctx.Value("request_id"),
		"assessment_id": assessmentID,
		"business_id":   assessment.BusinessID,
		"framework_id":  assessment.FrameworkID,
	})

	// Update assessment
	err := h.service.UpdateAssessment(ctx, &assessment)
	if err != nil {
		h.logger.Error("Failed to update compliance assessment", map[string]interface{}{
			"request_id":    ctx.Value("request_id"),
			"assessment_id": assessmentID,
			"error":         err.Error(),
		})
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, r, http.StatusNotFound, "assessment_not_found", "Compliance assessment not found")
		} else {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "assessment_update_failed", "Failed to update compliance assessment")
		}
		return
	}

	// Log successful update
	h.logger.Info("Compliance assessment updated successfully", map[string]interface{}{
		"request_id":    ctx.Value("request_id"),
		"assessment_id": assessmentID,
		"business_id":   assessment.BusinessID,
		"framework_id":  assessment.FrameworkID,
		"status":        assessment.Status,
		"duration_ms":   time.Since(start).Milliseconds(),
	})

	// Return updated assessment
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(assessment)
}

// GetBusinessAssessmentsHandler handles GET /v1/compliance/businesses/{business_id}/assessments
func (h *ComplianceFrameworkHandler) GetBusinessAssessmentsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractBusinessIDFromPath(r.URL.Path)
	if businessID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Parse framework filter
	frameworkID := r.URL.Query().Get("framework_id")

	h.logger.Info("Business assessments request received", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"business_id":  businessID,
		"framework_id": frameworkID,
		"user_agent":   r.UserAgent(),
		"remote_addr":  r.RemoteAddr,
	})

	// Get business assessments
	assessments, err := h.service.GetBusinessAssessments(ctx, businessID, frameworkID)
	if err != nil {
		h.logger.Error("Failed to get business assessments", map[string]interface{}{
			"request_id":   ctx.Value("request_id"),
			"business_id":  businessID,
			"framework_id": frameworkID,
			"error":        err.Error(),
		})
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "assessments_retrieval_failed", "Failed to retrieve business assessments")
		return
	}

	// Create response
	response := map[string]interface{}{
		"business_id":  businessID,
		"framework_id": frameworkID,
		"assessments":  assessments,
		"count":        len(assessments),
	}

	// Log successful request
	h.logger.Info("Business assessments retrieved successfully", map[string]interface{}{
		"request_id":   ctx.Value("request_id"),
		"business_id":  businessID,
		"framework_id": frameworkID,
		"count":        len(assessments),
		"duration_ms":  time.Since(start).Milliseconds(),
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper methods

// extractFrameworkIDFromPath extracts framework_id from URL path
func (h *ComplianceFrameworkHandler) extractFrameworkIDFromPath(path string) string {
	// Expected path format: /v1/compliance/frameworks/{framework_id} or /v1/compliance/frameworks/{framework_id}/requirements
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "v1" && parts[1] == "compliance" && parts[2] == "frameworks" {
		return parts[3]
	}
	return ""
}

// extractAssessmentIDFromPath extracts assessment_id from URL path
func (h *ComplianceFrameworkHandler) extractAssessmentIDFromPath(path string) string {
	// Expected path format: /v1/compliance/assessments/{assessment_id}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "v1" && parts[1] == "compliance" && parts[2] == "assessments" {
		return parts[3]
	}
	return ""
}

// extractBusinessIDFromPath extracts business_id from URL path
func (h *ComplianceFrameworkHandler) extractBusinessIDFromPath(path string) string {
	// Expected path format: /v1/compliance/businesses/{business_id}/assessments
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 5 && parts[0] == "v1" && parts[1] == "compliance" && parts[2] == "businesses" && parts[4] == "assessments" {
		return parts[3]
	}
	return ""
}

// generateAssessmentID generates a unique assessment ID
func (h *ComplianceFrameworkHandler) generateAssessmentID() string {
	return "assessment_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// writeErrorResponse writes an error response
func (h *ComplianceFrameworkHandler) writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, errorCode, message string) {
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
