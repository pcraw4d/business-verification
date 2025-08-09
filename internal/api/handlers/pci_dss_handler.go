package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pcraw4d/business-verification/internal/compliance"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// PCIDSSHandler handles PCI DSS specific compliance endpoints
type PCIDSSHandler struct {
	logger        *observability.Logger
	pciService    *compliance.PCIDSSTrackingService
	statusSystem  *compliance.ComplianceStatusSystem
	reportService *compliance.ReportGenerationService
}

// NewPCIDSSHandler creates a new PCI DSS handler
func NewPCIDSSHandler(logger *observability.Logger, pciService *compliance.PCIDSSTrackingService, statusSystem *compliance.ComplianceStatusSystem, reportService *compliance.ReportGenerationService) *PCIDSSHandler {
	return &PCIDSSHandler{
		logger:        logger,
		pciService:    pciService,
		statusSystem:  statusSystem,
		reportService: reportService,
	}
}

// InitializePCIDSSTrackingHandler handles POST /v1/pci-dss/initialize
// Request JSON: {"business_id": string, "merchant_level": string, "service_provider": bool}
func (h *PCIDSSHandler) InitializePCIDSSTrackingHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	var req struct {
		BusinessID      string `json:"business_id"`
		MerchantLevel   string `json:"merchant_level"`
		ServiceProvider bool   `json:"service_provider"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON in request body")
		return
	}

	if req.BusinessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	if req.MerchantLevel == "" {
		req.MerchantLevel = "Level 4" // Default to Level 4
	}

	// Initialize PCI DSS tracking
	err := h.pciService.InitializePCIDSSTracking(ctx, req.BusinessID, req.MerchantLevel, req.ServiceProvider)
	if err != nil {
		h.logger.Error("Failed to initialize PCI DSS tracking",
			"business_id", req.BusinessID,
			"merchant_level", req.MerchantLevel,
			"service_provider", req.ServiceProvider,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "initialization_failed", err.Error())
		return
	}

	response := map[string]interface{}{
		"message":         "PCI DSS tracking initialized successfully",
		"business_id":     req.BusinessID,
		"merchant_level":  req.MerchantLevel,
		"service_provider": req.ServiceProvider,
		"framework":       compliance.FrameworkPCIDSS,
		"version":         compliance.PCIDSSVersion4,
		"timestamp":       time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetPCIDSSStatusHandler handles GET /v1/pci-dss/status/{business_id}
func (h *PCIDSSHandler) GetPCIDSSStatusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractPathParam(r, "business_id")
	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Get PCI DSS status
	pciStatus, err := h.pciService.GetPCIDSSStatus(ctx, businessID)
	if err != nil {
		h.logger.Error("Failed to get PCI DSS status",
			"business_id", businessID,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusNotFound, "status_not_found", err.Error())
		return
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(pciStatus)
}

// UpdatePCIDSSRequirementHandler handles PUT /v1/pci-dss/requirements/{business_id}/{requirement_id}
// Request JSON: {"status": string, "implementation_status": string, "score": float64, "reviewer": string}
func (h *PCIDSSHandler) UpdatePCIDSSRequirementHandler(w http.ResponseWriter, r *http.Request) {
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

	// Update PCI DSS requirement status
	err := h.pciService.UpdatePCIDSSRequirementStatus(ctx, businessID, requirementID, status, implStatus, req.Score, req.Reviewer)
	if err != nil {
		h.logger.Error("Failed to update PCI DSS requirement status",
			"business_id", businessID,
			"requirement_id", requirementID,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	response := map[string]interface{}{
		"message":        "PCI DSS requirement status updated successfully",
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

// UpdatePCIDSSCategoryHandler handles PUT /v1/pci-dss/categories/{business_id}/{category_id}
// Request JSON: {"status": string, "score": float64, "reviewer": string}
func (h *PCIDSSHandler) UpdatePCIDSSCategoryHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract path parameters
	businessID := h.extractPathParam(r, "business_id")
	categoryID := h.extractPathParam(r, "category_id")

	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	if categoryID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "category_id is required")
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

	// Update PCI DSS category status
	err := h.pciService.UpdatePCIDSSCategoryStatus(ctx, businessID, categoryID, status, req.Score, req.Reviewer)
	if err != nil {
		h.logger.Error("Failed to update PCI DSS category status",
			"business_id", businessID,
			"category_id", categoryID,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	response := map[string]interface{}{
		"message":     "PCI DSS category status updated successfully",
		"business_id": businessID,
		"category_id": categoryID,
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

// AssessPCIDSSComplianceHandler handles POST /v1/pci-dss/assess/{business_id}
func (h *PCIDSSHandler) AssessPCIDSSComplianceHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Extract business_id from URL path
	businessID := h.extractPathParam(r, "business_id")
	if businessID == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "business_id is required")
		return
	}

	// Perform PCI DSS compliance assessment
	pciStatus, err := h.pciService.AssessPCIDSSCompliance(ctx, businessID)
	if err != nil {
		h.logger.Error("Failed to assess PCI DSS compliance",
			"business_id", businessID,
			"error", err.Error(),
		)
		h.writeError(w, r, http.StatusInternalServerError, "assessment_failed", err.Error())
		return
	}

	response := map[string]interface{}{
		"message":            "PCI DSS compliance assessment completed successfully",
		"business_id":        businessID,
		"overall_status":     pciStatus.OverallStatus,
		"compliance_score":   pciStatus.ComplianceScore,
		"merchant_level":     pciStatus.MerchantLevel,
		"service_provider":   pciStatus.ServiceProvider,
		"assessment_date":    pciStatus.LastAssessment,
		"next_assessment":    pciStatus.NextAssessment,
		"category_count":     len(pciStatus.CategoryStatus),
		"requirements_count": len(pciStatus.RequirementsStatus),
		"timestamp":          time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetPCIDSSReportHandler handles GET /v1/pci-dss/report/{business_id}
func (h *PCIDSSHandler) GetPCIDSSReportHandler(w http.ResponseWriter, r *http.Request) {
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

	// Generate PCI DSS compliance report
	report, err := h.pciService.GetPCIDSSReport(ctx, businessID, reportType)
	if err != nil {
		h.logger.Error("Failed to generate PCI DSS report",
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

// GetPCIDSSCategoriesHandler handles GET /v1/pci-dss/categories
func (h *PCIDSSHandler) GetPCIDSSCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Get PCI DSS categories from framework
	pciFramework := compliance.NewPCIDSSFramework()

	response := map[string]interface{}{
		"framework":   compliance.FrameworkPCIDSS,
		"version":     pciFramework.Version,
		"description": pciFramework.Description,
		"categories":  pciFramework.Categories,
		"timestamp":   time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetPCIDSSRequirementsHandler handles GET /v1/pci-dss/requirements
func (h *PCIDSSHandler) GetPCIDSSRequirementsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Get category filter from query parameter
	categoryFilter := r.URL.Query().Get("category")

	// Get PCI DSS requirements from framework
	pciFramework := compliance.NewPCIDSSFramework()
	requirements := pciFramework.Requirements

	// Filter by category if specified
	if categoryFilter != "" {
		filteredRequirements := []compliance.PCIDSSRequirement{}
		for _, req := range requirements {
			if req.Category == categoryFilter {
				filteredRequirements = append(filteredRequirements, req)
			}
		}
		requirements = filteredRequirements
	}

	response := map[string]interface{}{
		"framework":       compliance.FrameworkPCIDSS,
		"version":         pciFramework.Version,
		"category_filter": categoryFilter,
		"requirements":    requirements,
		"count":           len(requirements),
		"timestamp":       time.Now(),
	}

	h.logger.WithComponent("api").LogAPIRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Helper methods
func (h *PCIDSSHandler) extractPathParam(r *http.Request, paramName string) string {
	// This is a simplified implementation - in a real scenario, you'd use a proper router
	// For now, we'll extract from the URL path manually
	// Example: /v1/pci-dss/status/business123 -> extract "business123"
	// This is a basic implementation and should be replaced with proper routing
	return ""
}

func (h *PCIDSSHandler) writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
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
