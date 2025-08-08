package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/risk"
)

// RiskHandler handles risk assessment API requests
type RiskHandler struct {
	logger      *observability.Logger
	riskService *risk.RiskService
}

// NewRiskHandler creates a new risk handler
func NewRiskHandler(logger *observability.Logger, riskService *risk.RiskService) *RiskHandler {
	return &RiskHandler{
		logger:      logger,
		riskService: riskService,
	}
}

// AssessRiskHandler handles POST /v1/risk/assess requests
func (h *RiskHandler) AssessRiskHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Risk assessment request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Parse request body
	var request risk.RiskAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateRiskAssessmentRequest(request); err != nil {
		h.logger.Error("Invalid risk assessment request",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, fmt.Sprintf("Invalid request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Perform risk assessment
	response, err := h.riskService.AssessRisk(r.Context(), request)
	if err != nil {
		h.logger.Error("Risk assessment failed",
			"request_id", requestID,
			"business_id", request.BusinessID,
			"error", err.Error(),
		)
		http.Error(w, fmt.Sprintf("Risk assessment failed: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Risk assessment request completed",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"overall_score", response.Assessment.OverallScore,
		"overall_level", response.Assessment.OverallLevel,
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}

// validateRiskAssessmentRequest validates the risk assessment request
func (h *RiskHandler) validateRiskAssessmentRequest(request risk.RiskAssessmentRequest) error {
	if request.BusinessID == "" {
		return fmt.Errorf("business ID is required")
	}
	if request.BusinessName == "" {
		return fmt.Errorf("business name is required")
	}

	// Validate categories if provided
	for _, category := range request.Categories {
		if !h.isValidRiskCategory(category) {
			return fmt.Errorf("invalid risk category: %s", category)
		}
	}

	// Validate factors if provided
	for _, factorID := range request.Factors {
		if factorID == "" {
			return fmt.Errorf("factor ID cannot be empty")
		}
	}

	return nil
}

// isValidRiskCategory checks if a risk category is valid
func (h *RiskHandler) isValidRiskCategory(category risk.RiskCategory) bool {
	validCategories := []risk.RiskCategory{
		risk.RiskCategoryFinancial,
		risk.RiskCategoryOperational,
		risk.RiskCategoryRegulatory,
		risk.RiskCategoryReputational,
		risk.RiskCategoryCybersecurity,
	}

	for _, validCategory := range validCategories {
		if category == validCategory {
			return true
		}
	}
	return false
}

// GetRiskCategoriesHandler handles GET /v1/risk/categories requests
func (h *RiskHandler) GetRiskCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk categories request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
	)

	// Get categories from registry
	categories := h.riskService.GetCategoryRegistry().ListCategories()

	// Create response
	response := map[string]interface{}{
		"categories": categories,
		"total":      len(categories),
		"timestamp":  time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode categories response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Get risk categories request completed",
		"request_id", requestID,
		"total_categories", len(categories),
		"status_code", http.StatusOK,
	)
}

// GetRiskFactorsHandler handles GET /v1/risk/factors requests
func (h *RiskHandler) GetRiskFactorsHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk factors request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
	)

	// Parse query parameters
	category := r.URL.Query().Get("category")
	subcategory := r.URL.Query().Get("subcategory")

	var factors []*risk.RiskFactorDefinition

	if category != "" {
		// Get factors for specific category
		riskCategory := risk.RiskCategory(category)
		if !h.isValidRiskCategory(riskCategory) {
			http.Error(w, fmt.Sprintf("Invalid risk category: %s", category), http.StatusBadRequest)
			return
		}
		factors = h.riskService.GetCategoryRegistry().GetFactorsByCategory(riskCategory)
	} else if subcategory != "" {
		// Get factors for specific subcategory (requires category parameter)
		// For now, we'll get all factors and filter by subcategory
		allFactors := h.riskService.GetCategoryRegistry().ListFactors()
		for _, factor := range allFactors {
			if factor.Subcategory == subcategory {
				factors = append(factors, factor)
			}
		}
	} else {
		// Get all factors
		factors = h.riskService.GetCategoryRegistry().ListFactors()
	}

	// Create response
	response := map[string]interface{}{
		"factors":   factors,
		"total":     len(factors),
		"category":  category,
		"timestamp": time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode factors response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Get risk factors request completed",
		"request_id", requestID,
		"total_factors", len(factors),
		"category", category,
		"status_code", http.StatusOK,
	)
}

// GetRiskThresholdsHandler handles GET /v1/risk/thresholds requests
func (h *RiskHandler) GetRiskThresholdsHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk thresholds request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
	)

	// Parse query parameters
	category := r.URL.Query().Get("category")
	industryCode := r.URL.Query().Get("industry_code")

	var configs []*risk.ThresholdConfig

	if category != "" {
		// Get thresholds for specific category
		riskCategory := risk.RiskCategory(category)
		if !h.isValidRiskCategory(riskCategory) {
			http.Error(w, fmt.Sprintf("Invalid risk category: %s", category), http.StatusBadRequest)
			return
		}
		configs = h.riskService.GetThresholdManager().GetConfigsByCategory(riskCategory)
	} else if industryCode != "" {
		// Get thresholds for specific industry
		configs = h.riskService.GetThresholdManager().GetConfigsByIndustry(industryCode)
	} else {
		// Get all thresholds
		configs = h.riskService.GetThresholdManager().ListConfigs()
	}

	// Create response
	response := map[string]interface{}{
		"thresholds": configs,
		"total":      len(configs),
		"category":   category,
		"timestamp":  time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode thresholds response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Get risk thresholds request completed",
		"request_id", requestID,
		"total_thresholds", len(configs),
		"category", category,
		"status_code", http.StatusOK,
	)
}
