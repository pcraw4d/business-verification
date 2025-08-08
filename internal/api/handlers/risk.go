package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/risk"
)

// RiskHandler handles risk assessment API requests
type RiskHandler struct {
	logger             *observability.Logger
	riskService        *risk.RiskService
	riskHistoryService *risk.RiskHistoryService
}

// NewRiskHandler creates a new risk handler
func NewRiskHandler(logger *observability.Logger, riskService *risk.RiskService, riskHistoryService *risk.RiskHistoryService) *RiskHandler {
	return &RiskHandler{
		logger:             logger,
		riskService:        riskService,
		riskHistoryService: riskHistoryService,
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

// GetRiskHistoryHandler handles GET /v1/risk/history/{business_id} requests
func (h *RiskHandler) GetRiskHistoryHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk history request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Extract business ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid business ID", http.StatusBadRequest)
		return
	}
	businessID := pathParts[4]

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default limit
	offset := 0 // default offset

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get risk history
	response, err := h.riskHistoryService.GetRiskHistory(r.Context(), businessID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get risk history",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		http.Error(w, fmt.Sprintf("Failed to get risk history: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode risk history response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk history request completed",
		"request_id", requestID,
		"business_id", businessID,
		"total_assessments", response.TotalAssessments,
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}

// GetRiskTrendsHandler handles GET /v1/risk/trends/{business_id} requests
func (h *RiskHandler) GetRiskTrendsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk trends request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Extract business ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid business ID", http.StatusBadRequest)
		return
	}
	businessID := pathParts[4]

	// Parse query parameters
	daysStr := r.URL.Query().Get("days")
	days := 30 // default to 30 days

	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	// Get risk trends
	response, err := h.riskHistoryService.GetRiskTrends(r.Context(), businessID, days)
	if err != nil {
		h.logger.Error("Failed to get risk trends",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		http.Error(w, fmt.Sprintf("Failed to get risk trends: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode risk trends response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk trends request completed",
		"request_id", requestID,
		"business_id", businessID,
		"period_days", days,
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}

// GetRiskHistoryByDateRangeHandler handles GET /v1/risk/history/{business_id}/range requests
func (h *RiskHandler) GetRiskHistoryByDateRangeHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk history by date range request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Extract business ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid business ID", http.StatusBadRequest)
		return
	}
	businessID := pathParts[4]

	// Parse query parameters
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		http.Error(w, "start_date and end_date are required", http.StatusBadRequest)
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	if startDate.After(endDate) {
		http.Error(w, "start_date cannot be after end_date", http.StatusBadRequest)
		return
	}

	// Get risk history by date range
	assessments, err := h.riskHistoryService.GetRiskHistoryByDateRange(r.Context(), businessID, startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get risk history by date range",
			"request_id", requestID,
			"business_id", businessID,
			"start_date", startDateStr,
			"end_date", endDateStr,
			"error", err.Error(),
		)
		http.Error(w, fmt.Sprintf("Failed to get risk history: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Create response
	response := map[string]interface{}{
		"business_id":  businessID,
		"start_date":   startDateStr,
		"end_date":     endDateStr,
		"assessments":  assessments,
		"total":        len(assessments),
		"generated_at": time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode risk history by date range response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk history by date range request completed",
		"request_id", requestID,
		"business_id", businessID,
		"start_date", startDateStr,
		"end_date", endDateStr,
		"total_assessments", len(assessments),
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}

// GetRiskAlertsHandler handles GET /v1/risk/alerts/{business_id} requests
func (h *RiskHandler) GetRiskAlertsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk alerts request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Extract business ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid business ID", http.StatusBadRequest)
		return
	}
	businessID := pathParts[4]

	// Parse query parameters
	level := r.URL.Query().Get("level")
	acknowledged := r.URL.Query().Get("acknowledged")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default limit
	offset := 0 // default offset

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Create mock alerts for demonstration
	// In a real implementation, this would query the database
	alerts := []risk.RiskAlert{
		{
			ID:           fmt.Sprintf("alert_%s_1", businessID),
			BusinessID:   businessID,
			RiskFactor:   "financial_risk",
			Level:        risk.RiskLevelHigh,
			Message:      "High financial risk detected: 75.5 (Level: high)",
			Score:        75.5,
			Threshold:    70.0,
			TriggeredAt:  time.Now().Add(-2 * time.Hour),
			Acknowledged: false,
		},
		{
			ID:           fmt.Sprintf("alert_%s_2", businessID),
			BusinessID:   businessID,
			RiskFactor:   "operational_risk",
			Level:        risk.RiskLevelMedium,
			Message:      "Operational risk exceeds threshold: 65.2",
			Score:        65.2,
			Threshold:    60.0,
			TriggeredAt:  time.Now().Add(-1 * time.Hour),
			Acknowledged: false,
		},
	}

	// Filter by level if specified
	if level != "" {
		filteredAlerts := []risk.RiskAlert{}
		for _, alert := range alerts {
			if string(alert.Level) == level {
				filteredAlerts = append(filteredAlerts, alert)
			}
		}
		alerts = filteredAlerts
	}

	// Filter by acknowledged status if specified
	if acknowledged != "" {
		acknowledgedBool := acknowledged == "true"
		filteredAlerts := []risk.RiskAlert{}
		for _, alert := range alerts {
			if alert.Acknowledged == acknowledgedBool {
				filteredAlerts = append(filteredAlerts, alert)
			}
		}
		alerts = filteredAlerts
	}

	// Apply pagination
	if offset < len(alerts) {
		end := offset + limit
		if end > len(alerts) {
			end = len(alerts)
		}
		alerts = alerts[offset:end]
	} else {
		alerts = []risk.RiskAlert{}
	}

	// Create response
	response := map[string]interface{}{
		"business_id":  businessID,
		"alerts":       alerts,
		"total":        len(alerts),
		"limit":        limit,
		"offset":       offset,
		"level":        level,
		"acknowledged": acknowledged,
		"generated_at": time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode risk alerts response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk alerts request completed",
		"request_id", requestID,
		"business_id", businessID,
		"total_alerts", len(alerts),
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}

// GetRiskAlertRulesHandler handles GET /v1/risk/alert-rules requests
func (h *RiskHandler) GetRiskAlertRulesHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk alert rules request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Parse query parameters
	category := r.URL.Query().Get("category")
	enabled := r.URL.Query().Get("enabled")

	// Create mock alert rules for demonstration
	// In a real implementation, this would query the alert service
	rules := []risk.AlertRule{
		{
			ID:          "rule_overall_critical",
			Name:        "Overall Critical Risk",
			Description: "Alert when overall risk score exceeds 80",
			Category:    risk.RiskCategoryOperational,
			Condition:   risk.AlertConditionGreaterThan,
			Threshold:   80.0,
			Level:       risk.RiskLevelCritical,
			Message:     "Critical overall risk detected",
			Enabled:     true,
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rule_financial_high",
			Name:        "High Financial Risk",
			Description: "Alert when financial risk score exceeds 70",
			Category:    risk.RiskCategoryFinancial,
			Condition:   risk.AlertConditionGreaterThan,
			Threshold:   70.0,
			Level:       risk.RiskLevelHigh,
			Message:     "High financial risk detected",
			Enabled:     true,
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rule_regulatory_critical",
			Name:        "Critical Regulatory Risk",
			Description: "Alert when regulatory risk score exceeds 85",
			Category:    risk.RiskCategoryRegulatory,
			Condition:   risk.AlertConditionGreaterThan,
			Threshold:   85.0,
			Level:       risk.RiskLevelCritical,
			Message:     "Critical regulatory risk detected",
			Enabled:     true,
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
	}

	// Filter by category if specified
	if category != "" {
		filteredRules := []risk.AlertRule{}
		for _, rule := range rules {
			if string(rule.Category) == category {
				filteredRules = append(filteredRules, rule)
			}
		}
		rules = filteredRules
	}

	// Filter by enabled status if specified
	if enabled != "" {
		enabledBool := enabled == "true"
		filteredRules := []risk.AlertRule{}
		for _, rule := range rules {
			if rule.Enabled == enabledBool {
				filteredRules = append(filteredRules, rule)
			}
		}
		rules = filteredRules
	}

	// Create response
	response := map[string]interface{}{
		"rules":        rules,
		"total":        len(rules),
		"category":     category,
		"enabled":      enabled,
		"generated_at": time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode risk alert rules response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk alert rules request completed",
		"request_id", requestID,
		"total_rules", len(rules),
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}

// AcknowledgeRiskAlertHandler handles POST /v1/risk/alerts/{alert_id}/acknowledge requests
func (h *RiskHandler) AcknowledgeRiskAlertHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Acknowledge risk alert request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Extract alert ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}
	alertID := pathParts[4]

	// Parse request body for acknowledgment details
	var request struct {
		UserID         string     `json:"user_id"`
		Comment        string     `json:"comment,omitempty"`
		AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode acknowledge request",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// In a real implementation, this would update the alert in the database
	// For now, we'll simulate the acknowledgment
	acknowledgedAt := time.Now()
	if request.AcknowledgedAt != nil {
		acknowledgedAt = *request.AcknowledgedAt
	}

	// Create response
	response := map[string]interface{}{
		"alert_id":        alertID,
		"acknowledged":    true,
		"acknowledged_at": acknowledgedAt,
		"user_id":         request.UserID,
		"comment":         request.Comment,
		"updated_at":      time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode acknowledge response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Acknowledge risk alert request completed",
		"request_id", requestID,
		"alert_id", alertID,
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}

// GenerateRiskReportHandler handles POST /v1/risk/reports/generate requests
func (h *RiskHandler) GenerateRiskReportHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Generate risk report request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Parse request body
	var request risk.ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode report request",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.BusinessID == "" {
		h.logger.Error("Missing business ID in report request",
			"request_id", requestID,
		)
		http.Error(w, "Business ID is required", http.StatusBadRequest)
		return
	}

	if request.ReportType == "" {
		request.ReportType = risk.ReportTypeSummary // Default to summary
	}

	if request.Format == "" {
		request.Format = risk.ReportFormatJSON // Default to JSON
	}

	// Generate report
	ctx := context.WithValue(r.Context(), "request_id", requestID)
	report, err := h.riskService.GenerateRiskReport(ctx, request)
	if err != nil {
		h.logger.Error("Failed to generate risk report",
			"request_id", requestID,
			"business_id", request.BusinessID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}

	// Create response
	response := map[string]interface{}{
		"report":       report,
		"business_id":  request.BusinessID,
		"report_type":  request.ReportType,
		"format":       request.Format,
		"generated_at": time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode risk report response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Generate risk report request completed",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"report_type", request.ReportType,
		"format", request.Format,
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}

// GetRiskReportTypesHandler handles GET /v1/risk/reports/types requests
func (h *RiskHandler) GetRiskReportTypesHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk report types request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Define available report types
	reportTypes := []map[string]interface{}{
		{
			"type":        "summary",
			"name":        "Summary Report",
			"description": "High-level overview of risk assessment with key metrics and alerts",
			"format":      []string{"json", "pdf", "html"},
		},
		{
			"type":        "detailed",
			"name":        "Detailed Report",
			"description": "Comprehensive analysis with factor breakdowns and historical data",
			"format":      []string{"json", "pdf", "html"},
		},
		{
			"type":        "trend",
			"name":        "Trend Analysis Report",
			"description": "Risk trend analysis with forecasting and pattern identification",
			"format":      []string{"json", "pdf", "html"},
		},
		{
			"type":        "executive",
			"name":        "Executive Report",
			"description": "Executive summary with key insights and recommendations",
			"format":      []string{"json", "pdf", "html"},
		},
		{
			"type":        "compliance",
			"name":        "Compliance Report",
			"description": "Compliance-focused report with regulatory requirements",
			"format":      []string{"json", "pdf", "html"},
		},
		{
			"type":        "alert",
			"name":        "Alert Report",
			"description": "Report focused on active alerts and their resolution",
			"format":      []string{"json", "pdf", "html"},
		},
	}

	// Define available formats
	formats := []map[string]interface{}{
		{
			"format":       "json",
			"name":         "JSON",
			"description":  "Machine-readable JSON format",
			"content_type": "application/json",
		},
		{
			"format":       "pdf",
			"name":         "PDF",
			"description":  "Portable Document Format for printing and sharing",
			"content_type": "application/pdf",
		},
		{
			"format":       "html",
			"name":         "HTML",
			"description":  "Web-friendly HTML format",
			"content_type": "text/html",
		},
		{
			"format":       "csv",
			"name":         "CSV",
			"description":  "Comma-separated values for data analysis",
			"content_type": "text/csv",
		},
	}

	// Create response
	response := map[string]interface{}{
		"report_types": reportTypes,
		"formats":      formats,
		"generated_at": time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode risk report types response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk report types request completed",
		"request_id", requestID,
		"total_types", len(reportTypes),
		"total_formats", len(formats),
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}

// GetRiskReportHistoryHandler handles GET /v1/risk/reports/history/{business_id} requests
func (h *RiskHandler) GetRiskReportHistoryHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk report history request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Extract business ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid business ID", http.StatusBadRequest)
		return
	}
	businessID := pathParts[5]

	// Parse query parameters
	reportType := r.URL.Query().Get("report_type")
	format := r.URL.Query().Get("format")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // default limit
	offset := 0 // default offset

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Create mock report history for demonstration
	// In a real implementation, this would query the database
	reports := []map[string]interface{}{
		{
			"id":            fmt.Sprintf("report_%s_1", businessID),
			"business_id":   businessID,
			"business_name": "Test Business",
			"report_type":   "summary",
			"format":        "json",
			"generated_at":  time.Now().Add(-24 * time.Hour),
			"valid_until":   time.Now().Add(24 * time.Hour),
			"file_size":     1024,
			"status":        "completed",
		},
		{
			"id":            fmt.Sprintf("report_%s_2", businessID),
			"business_id":   businessID,
			"business_name": "Test Business",
			"report_type":   "detailed",
			"format":        "pdf",
			"generated_at":  time.Now().Add(-48 * time.Hour),
			"valid_until":   time.Now().Add(24 * time.Hour),
			"file_size":     2048,
			"status":        "completed",
		},
		{
			"id":            fmt.Sprintf("report_%s_3", businessID),
			"business_id":   businessID,
			"business_name": "Test Business",
			"report_type":   "trend",
			"format":        "html",
			"generated_at":  time.Now().Add(-72 * time.Hour),
			"valid_until":   time.Now().Add(24 * time.Hour),
			"file_size":     1536,
			"status":        "completed",
		},
	}

	// Filter by report type if specified
	if reportType != "" {
		filteredReports := []map[string]interface{}{}
		for _, report := range reports {
			if report["report_type"] == reportType {
				filteredReports = append(filteredReports, report)
			}
		}
		reports = filteredReports
	}

	// Filter by format if specified
	if format != "" {
		filteredReports := []map[string]interface{}{}
		for _, report := range reports {
			if report["format"] == format {
				filteredReports = append(filteredReports, report)
			}
		}
		reports = filteredReports
	}

	// Apply pagination
	if offset < len(reports) {
		end := offset + limit
		if end > len(reports) {
			end = len(reports)
		}
		reports = reports[offset:end]
	} else {
		reports = []map[string]interface{}{}
	}

	// Create response
	response := map[string]interface{}{
		"business_id":  businessID,
		"reports":      reports,
		"total":        len(reports),
		"limit":        limit,
		"offset":       offset,
		"report_type":  reportType,
		"format":       format,
		"generated_at": time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode risk report history response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk report history request completed",
		"request_id", requestID,
		"business_id", businessID,
		"total_reports", len(reports),
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}
