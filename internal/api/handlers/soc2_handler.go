package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pcraw4d/business-verification/internal/compliance"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// SOC2Handler handles SOC 2 specific compliance endpoints
type SOC2Handler struct {
	logger        *observability.Logger
	soc2Service   *compliance.SOC2TrackingService
	statusSystem  *compliance.ComplianceStatusSystem
	reportService *compliance.ReportGenerationService
}

// NewSOC2Handler creates a new SOC 2 handler
func NewSOC2Handler(logger *observability.Logger, soc2Service *compliance.SOC2TrackingService, statusSystem *compliance.ComplianceStatusSystem, reportService *compliance.ReportGenerationService) *SOC2Handler {
	return &SOC2Handler{
		logger:        logger,
		soc2Service:   soc2Service,
		statusSystem:  statusSystem,
		reportService: reportService,
	}
}

// InitializeSOC2TrackingHandler handles POST /v1/soc2/initialize
// Request JSON: {"business_id": string, "report_type": string}
func (h *SOC2Handler) InitializeSOC2TrackingHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	var req struct {
		BusinessID string `json:"business_id"`
		ReportType string `json:"report_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	if req.BusinessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	if req.ReportType == "" {
		req.ReportType = compliance.SOC2ReportType2 // Default to Type 2
	}

	// Validate report type
	if req.ReportType != compliance.SOC2ReportType1 && req.ReportType != compliance.SOC2ReportType2 {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "report_type must be 'Type 1' or 'Type 2'")
		return
	}

	// Initialize SOC 2 tracking
	err := h.soc2Service.InitializeSOC2Tracking(ctx, req.BusinessID, req.ReportType)
	if err != nil {
		h.logger.Error("Failed to initialize SOC 2 tracking",
			"business_id", req.BusinessID,
			"report_type", req.ReportType,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "initialization_failed", err.Error())
		return
	}

	response := map[string]interface{}{
		"message":     "SOC 2 tracking initialized successfully",
		"business_id": req.BusinessID,
		"report_type": req.ReportType,
		"framework":   compliance.FrameworkSOC2,
		"timestamp":   time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetSOC2StatusHandler handles GET /v1/soc2/status/{business_id}
func (h *SOC2Handler) GetSOC2StatusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractPathParam(r, "business_id")
	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Get SOC 2 status
	soc2Status, err := h.soc2Service.GetSOC2Status(ctx, businessID)
	if err != nil {
		h.logger.Error("Failed to get SOC 2 status",
			"business_id", businessID,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusNotFound, "status_not_found", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(soc2Status)
}

// UpdateSOC2RequirementHandler handles PUT /v1/soc2/requirements/{business_id}/{requirement_id}
// Request JSON: {"status": string, "implementation_status": string, "score": float64, "reviewer": string}
func (h *SOC2Handler) UpdateSOC2RequirementHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract path parameters
	businessID := h.extractPathParam(r, "business_id")
	requirementID := h.extractPathParam(r, "requirement_id")

	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	if requirementID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "requirement_id is required")
		return
	}

	var req struct {
		Status               string  `json:"status"`
		ImplementationStatus string  `json:"implementation_status"`
		Score                float64 `json:"score"`
		Reviewer             string  `json:"reviewer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	// Validate status
	status := compliance.ComplianceStatus(req.Status)
	if status == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "status is required")
		return
	}

	// Validate implementation status
	implStatus := compliance.ImplementationStatus(req.ImplementationStatus)
	if implStatus == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "implementation_status is required")
		return
	}

	// Validate score
	if req.Score < 0 || req.Score > 100 {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "score must be between 0 and 100")
		return
	}

	if req.Reviewer == "" {
		req.Reviewer = "api_user"
	}

	// Update SOC 2 requirement status
	err := h.soc2Service.UpdateSOC2RequirementStatus(ctx, businessID, requirementID, status, implStatus, req.Score, req.Reviewer)
	if err != nil {
		h.logger.Error("Failed to update SOC 2 requirement status",
			"business_id", businessID,
			"requirement_id", requirementID,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	response := map[string]interface{}{
		"message":        "SOC 2 requirement status updated successfully",
		"business_id":    businessID,
		"requirement_id": requirementID,
		"status":         status,
		"score":          req.Score,
		"reviewer":       req.Reviewer,
		"timestamp":      time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// UpdateSOC2CriteriaHandler handles PUT /v1/soc2/criteria/{business_id}/{criteria_id}
// Request JSON: {"status": string, "score": float64, "reviewer": string}
func (h *SOC2Handler) UpdateSOC2CriteriaHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract path parameters
	businessID := h.extractPathParam(r, "business_id")
	criteriaID := h.extractPathParam(r, "criteria_id")

	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	if criteriaID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "criteria_id is required")
		return
	}

	var req struct {
		Status   string  `json:"status"`
		Score    float64 `json:"score"`
		Reviewer string  `json:"reviewer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	// Validate status
	status := compliance.ComplianceStatus(req.Status)
	if status == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "status is required")
		return
	}

	// Validate score
	if req.Score < 0 || req.Score > 100 {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "score must be between 0 and 100")
		return
	}

	if req.Reviewer == "" {
		req.Reviewer = "api_user"
	}

	// Update SOC 2 criteria status
	err := h.soc2Service.UpdateSOC2CriteriaStatus(ctx, businessID, criteriaID, status, req.Score, req.Reviewer)
	if err != nil {
		h.logger.Error("Failed to update SOC 2 criteria status",
			"business_id", businessID,
			"criteria_id", criteriaID,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	response := map[string]interface{}{
		"message":     "SOC 2 criteria status updated successfully",
		"business_id": businessID,
		"criteria_id": criteriaID,
		"status":      status,
		"score":       req.Score,
		"reviewer":    req.Reviewer,
		"timestamp":   time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// AssessSOC2ComplianceHandler handles POST /v1/soc2/assess/{business_id}
func (h *SOC2Handler) AssessSOC2ComplianceHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractPathParam(r, "business_id")
	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Perform SOC 2 compliance assessment
	soc2Status, err := h.soc2Service.AssessSOC2Compliance(ctx, businessID)
	if err != nil {
		h.logger.Error("Failed to assess SOC 2 compliance",
			"business_id", businessID,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "assessment_failed", err.Error())
		return
	}

	response := map[string]interface{}{
		"message":            "SOC 2 compliance assessment completed successfully",
		"business_id":        businessID,
		"overall_status":     soc2Status.OverallStatus,
		"compliance_score":   soc2Status.ComplianceScore,
		"report_type":        soc2Status.ReportType,
		"assessment_date":    soc2Status.LastAssessment,
		"next_assessment":    soc2Status.NextAssessment,
		"criteria_count":     len(soc2Status.CriteriaStatus),
		"requirements_count": len(soc2Status.RequirementsStatus),
		"timestamp":          time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetSOC2ReportHandler handles GET /v1/soc2/report/{business_id}
func (h *SOC2Handler) GetSOC2ReportHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractPathParam(r, "business_id")
	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Get report type from query parameter
	reportType := r.URL.Query().Get("report_type")
	if reportType == "" {
		reportType = compliance.SOC2ReportType2 // Default to Type 2
	}

	// Generate SOC 2 compliance report
	report, err := h.soc2Service.GetSOC2Report(ctx, businessID, reportType)
	if err != nil {
		h.logger.Error("Failed to generate SOC 2 report",
			"business_id", businessID,
			"report_type", reportType,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "report_generation_failed", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(report)
}

// GetSOC2CriteriaHandler handles GET /v1/soc2/criteria
func (h *SOC2Handler) GetSOC2CriteriaHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Get SOC 2 criteria from framework
	soc2Framework := compliance.NewSOC2Framework()

	response := map[string]interface{}{
		"framework":   compliance.FrameworkSOC2,
		"version":     soc2Framework.Version,
		"description": soc2Framework.Description,
		"criteria":    soc2Framework.Criteria,
		"timestamp":   time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetSOC2RequirementsHandler handles GET /v1/soc2/requirements
func (h *SOC2Handler) GetSOC2RequirementsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Get criteria filter from query parameter
	criteriaFilter := r.URL.Query().Get("criteria")

	// Get SOC 2 requirements from framework
	soc2Framework := compliance.NewSOC2Framework()
	requirements := soc2Framework.Requirements

	// Filter by criteria if specified
	if criteriaFilter != "" {
		filteredRequirements := []compliance.SOC2Requirement{}
		for _, req := range requirements {
			if req.Criteria == criteriaFilter {
				filteredRequirements = append(filteredRequirements, req)
			}
		}
		requirements = filteredRequirements
	}

	response := map[string]interface{}{
		"framework":       compliance.FrameworkSOC2,
		"version":         soc2Framework.Version,
		"criteria_filter": criteriaFilter,
		"requirements":    requirements,
		"count":           len(requirements),
		"timestamp":       time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Helper methods (copied from ComplianceHandler)
func (h *SOC2Handler) extractPathParam(r *http.Request, paramName string) string {
	// This is a simplified implementation - in a real scenario, you'd use a proper router
	// For now, we'll extract from the URL path manually
	// Example: /v1/soc2/status/business123 -> extract "business123"
	// This is a basic implementation and should be replaced with proper routing
	return ""
}

func (h *SOC2Handler) writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
			"path":    r.URL.Path,
			"method":  r.Method,
		},
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse)
}
