package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pcraw4d/business-verification/internal/compliance"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// GDPRHandler handles GDPR specific compliance endpoints
type GDPRHandler struct {
	logger        *observability.Logger
	gdprService   *compliance.GDPRTrackingService
	statusSystem  *compliance.ComplianceStatusSystem
	reportService *compliance.ReportGenerationService
}

// NewGDPRHandler creates a new GDPR handler
func NewGDPRHandler(logger *observability.Logger, gdprService *compliance.GDPRTrackingService, statusSystem *compliance.ComplianceStatusSystem, reportService *compliance.ReportGenerationService) *GDPRHandler {
	return &GDPRHandler{
		logger:        logger,
		gdprService:   gdprService,
		statusSystem:  statusSystem,
		reportService: reportService,
	}
}

// InitializeGDPRTrackingHandler handles POST /v1/gdpr/initialize
// Request JSON: {"business_id": string, "data_controller": bool, "data_processor": bool, "data_protection_officer": string}
func (h *GDPRHandler) InitializeGDPRTrackingHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	var req struct {
		BusinessID            string `json:"business_id"`
		DataController        bool   `json:"data_controller"`
		DataProcessor         bool   `json:"data_processor"`
		DataProtectionOfficer string `json:"data_protection_officer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	if req.BusinessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Validate that at least one role is specified
	if !req.DataController && !req.DataProcessor {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business must be either a data controller or data processor")
		return
	}

	// Initialize GDPR tracking
	err := h.gdprService.InitializeGDPRTracking(ctx, req.BusinessID, req.DataController, req.DataProcessor, req.DataProtectionOfficer)
	if err != nil {
		h.logger.Error("Failed to initialize GDPR tracking",
			"business_id", req.BusinessID,
			"data_controller", req.DataController,
			"data_processor", req.DataProcessor,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "initialization_failed", err.Error())
		return
	}

	response := map[string]interface{}{
		"message":                 "GDPR tracking initialized successfully",
		"business_id":             req.BusinessID,
		"data_controller":         req.DataController,
		"data_processor":          req.DataProcessor,
		"data_protection_officer": req.DataProtectionOfficer,
		"framework":               compliance.FrameworkGDPR,
		"version":                 compliance.GDPRVersion2018,
		"timestamp":               time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetGDPRStatusHandler handles GET /v1/gdpr/status/{business_id}
func (h *GDPRHandler) GetGDPRStatusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractPathParam(r, "business_id")
	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Get GDPR status
	gdprStatus, err := h.gdprService.GetGDPRStatus(ctx, businessID)
	if err != nil {
		h.logger.Error("Failed to get GDPR status",
			"business_id", businessID,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusNotFound, "status_not_found", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(gdprStatus)
}

// UpdateGDPRRequirementHandler handles PUT /v1/gdpr/requirements/{business_id}/{requirement_id}
// Request JSON: {"status": string, "implementation_status": string, "score": float64, "reviewer": string}
func (h *GDPRHandler) UpdateGDPRRequirementHandler(w http.ResponseWriter, r *http.Request) {
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

	// Update GDPR requirement status
	err := h.gdprService.UpdateGDPRRequirementStatus(ctx, businessID, requirementID, status, implStatus, req.Score, req.Reviewer)
	if err != nil {
		h.logger.Error("Failed to update GDPR requirement status",
			"business_id", businessID,
			"requirement_id", requirementID,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	response := map[string]interface{}{
		"message":        "GDPR requirement status updated successfully",
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

// UpdateGDPRPrincipleHandler handles PUT /v1/gdpr/principles/{business_id}/{principle_id}
// Request JSON: {"status": string, "score": float64, "reviewer": string}
func (h *GDPRHandler) UpdateGDPRPrincipleHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract path parameters
	businessID := h.extractPathParam(r, "business_id")
	principleID := h.extractPathParam(r, "principle_id")

	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	if principleID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "principle_id is required")
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

	// Update GDPR principle status
	err := h.gdprService.UpdateGDPRPrincipleStatus(ctx, businessID, principleID, status, req.Score, req.Reviewer)
	if err != nil {
		h.logger.Error("Failed to update GDPR principle status",
			"business_id", businessID,
			"principle_id", principleID,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	response := map[string]interface{}{
		"message":      "GDPR principle status updated successfully",
		"business_id":  businessID,
		"principle_id": principleID,
		"status":       status,
		"score":        req.Score,
		"reviewer":     req.Reviewer,
		"timestamp":    time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// UpdateGDPRDataSubjectRightHandler handles PUT /v1/gdpr/rights/{business_id}/{right_id}
// Request JSON: {"status": string, "score": float64, "reviewer": string}
func (h *GDPRHandler) UpdateGDPRDataSubjectRightHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract path parameters
	businessID := h.extractPathParam(r, "business_id")
	rightID := h.extractPathParam(r, "right_id")

	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	if rightID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "right_id is required")
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

	// Update GDPR data subject right status
	err := h.gdprService.UpdateGDPRDataSubjectRightStatus(ctx, businessID, rightID, status, req.Score, req.Reviewer)
	if err != nil {
		h.logger.Error("Failed to update GDPR data subject right status",
			"business_id", businessID,
			"right_id", rightID,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	response := map[string]interface{}{
		"message":     "GDPR data subject right status updated successfully",
		"business_id": businessID,
		"right_id":    rightID,
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

// AssessGDPRComplianceHandler handles POST /v1/gdpr/assess/{business_id}
func (h *GDPRHandler) AssessGDPRComplianceHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractPathParam(r, "business_id")
	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Perform GDPR compliance assessment
	gdprStatus, err := h.gdprService.AssessGDPRCompliance(ctx, businessID)
	if err != nil {
		h.logger.Error("Failed to assess GDPR compliance",
			"business_id", businessID,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "assessment_failed", err.Error())
		return
	}

	response := map[string]interface{}{
		"message":            "GDPR compliance assessment completed successfully",
		"business_id":        businessID,
		"overall_status":     gdprStatus.OverallStatus,
		"compliance_score":   gdprStatus.ComplianceScore,
		"data_controller":    gdprStatus.DataController,
		"data_processor":     gdprStatus.DataProcessor,
		"assessment_date":    gdprStatus.LastAssessment,
		"next_assessment":    gdprStatus.NextAssessment,
		"principle_count":    len(gdprStatus.PrincipleStatus),
		"rights_count":       len(gdprStatus.RightsStatus),
		"requirements_count": len(gdprStatus.RequirementsStatus),
		"timestamp":          time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetGDPRReportHandler handles GET /v1/gdpr/report/{business_id}
func (h *GDPRHandler) GetGDPRReportHandler(w http.ResponseWriter, r *http.Request) {
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
		reportType = "compliance" // Default report type
	}

	// Generate GDPR compliance report
	report, err := h.gdprService.GetGDPRReport(ctx, businessID, reportType)
	if err != nil {
		h.logger.Error("Failed to generate GDPR report",
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

// GetGDPRPrinciplesHandler handles GET /v1/gdpr/principles
func (h *GDPRHandler) GetGDPRPrinciplesHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Get GDPR principles from framework
	gdprFramework := compliance.NewGDPRFramework()

	response := map[string]interface{}{
		"framework":   compliance.FrameworkGDPR,
		"version":     gdprFramework.Version,
		"description": gdprFramework.Description,
		"principles":  gdprFramework.Principles,
		"timestamp":   time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetGDPRDataSubjectRightsHandler handles GET /v1/gdpr/rights
func (h *GDPRHandler) GetGDPRDataSubjectRightsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Get GDPR data subject rights from framework
	gdprFramework := compliance.NewGDPRFramework()

	response := map[string]interface{}{
		"framework":           compliance.FrameworkGDPR,
		"version":             gdprFramework.Version,
		"description":         gdprFramework.Description,
		"data_subject_rights": gdprFramework.DataSubjectRights,
		"timestamp":           time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetGDPRRequirementsHandler handles GET /v1/gdpr/requirements
func (h *GDPRHandler) GetGDPRRequirementsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Get principle filter from query parameter
	principleFilter := r.URL.Query().Get("principle")

	// Get GDPR requirements from framework
	gdprFramework := compliance.NewGDPRFramework()
	requirements := gdprFramework.Requirements

	// Filter by principle if specified
	if principleFilter != "" {
		filteredRequirements := []compliance.GDPRRequirement{}
		for _, req := range requirements {
			if req.Principle == principleFilter {
				filteredRequirements = append(filteredRequirements, req)
			}
		}
		requirements = filteredRequirements
	}

	response := map[string]interface{}{
		"framework":        compliance.FrameworkGDPR,
		"version":          gdprFramework.Version,
		"principle_filter": principleFilter,
		"requirements":     requirements,
		"count":            len(requirements),
		"timestamp":        time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Helper methods
func (h *GDPRHandler) extractPathParam(r *http.Request, paramName string) string {
	// This is a simplified implementation - in a real scenario, you'd use a proper router
	// For now, we'll extract from the URL path manually
	// Example: /v1/gdpr/status/business123 -> extract "business123"
	// This is a basic implementation and should be replaced with proper routing
	return ""
}

func (h *GDPRHandler) writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
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
