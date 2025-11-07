package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kyb-platform/internal/observability"
	"kyb-platform/internal/risk"
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

	h.logger.Info("Risk assessment request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

	// Parse request body
	var request risk.RiskAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateRiskAssessmentRequest(request); err != nil {
		h.logger.Error("Invalid risk assessment request", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, fmt.Sprintf("Invalid request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Perform risk assessment
	response, err := h.riskService.AssessRisk(r.Context(), request.BusinessID)
	if err != nil {
		h.logger.Error("Risk assessment failed", map[string]interface{}{
			"request_id":  requestID,
			"business_id": request.BusinessID,
			"error":       err.Error(),
		})
		http.Error(w, fmt.Sprintf("Risk assessment failed: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Risk assessment request completed", map[string]interface{}{
		"request_id":    requestID,
		"business_id":   request.BusinessID,
		"overall_score": response.OverallScore,
		"overall_level": response.OverallLevel,
		"duration_ms":   duration.Milliseconds(),
		"status_code":   http.StatusOK,
	})
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

	h.logger.Info("Get risk categories request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
	})

	// Mock categories since GetCategoryRegistry doesn't exist
	categories := []map[string]interface{}{
		{
			"id":          "financial",
			"name":        "Financial Risk",
			"description": "Financial risk factors",
		},
		{
			"id":          "operational",
			"name":        "Operational Risk",
			"description": "Operational risk factors",
		},
	}

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
		h.logger.Error("Failed to encode categories response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Get risk categories request completed", map[string]interface{}{
		"request_id":       requestID,
		"total_categories": len(categories),
		"status_code":      http.StatusOK,
	})
}

// GetRiskFactorsHandler handles GET /v1/risk/factors requests
func (h *RiskHandler) GetRiskFactorsHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk factors request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
	})

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
		// Mock factors since GetCategoryRegistry doesn't exist
		factors = []*risk.RiskFactorDefinition{
			{
				ID:       "factor-1",
				Name:     "Financial Factor 1",
				Category: riskCategory,
			},
		}
	} else if subcategory != "" {
		// Get factors for specific subcategory (requires category parameter)
		// For now, we'll get all factors and filter by subcategory
		// Mock all factors since GetCategoryRegistry doesn't exist
		allFactors := []*risk.RiskFactorDefinition{
			{
				ID:       "factor-1",
				Name:     "Financial Factor 1",
				Category: risk.RiskCategoryFinancial,
			},
		}
		for _, factor := range allFactors {
			if factor.Subcategory == subcategory {
				factors = append(factors, factor)
			}
		}
	} else {
		// Get all factors
		// Mock all factors since GetCategoryRegistry doesn't exist
		factors = []*risk.RiskFactorDefinition{
			{
				ID:       "factor-1",
				Name:     "Financial Factor 1",
				Category: risk.RiskCategoryFinancial,
			},
		}
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
		h.logger.Error("Failed to encode factors response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Get risk factors request completed", map[string]interface{}{
		"request_id":    requestID,
		"total_factors": len(factors),
		"category":      category,
		"status_code":   http.StatusOK,
	})
}

// GetRiskThresholdsHandler handles GET /v1/risk/thresholds requests
func (h *RiskHandler) GetRiskThresholdsHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk thresholds request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
	})

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
		// Mock configs since GetThresholdManager doesn't exist
		configs = []*risk.ThresholdConfig{
			{
				ID:       "config-1",
				Name:     "Default Config",
				Category: riskCategory,
				RiskLevels: map[risk.RiskLevel]float64{
					risk.RiskLevelLow:    0.3,
					risk.RiskLevelMedium: 0.7,
					risk.RiskLevelHigh:   1.0,
				},
				IsActive: true,
			},
		}
	} else if industryCode != "" {
		// Get thresholds for specific industry
		// Mock configs since GetThresholdManager doesn't exist
		configs = []*risk.ThresholdConfig{
			{
				ID:           "config-2",
				Name:         "Industry Config",
				IndustryCode: industryCode,
				RiskLevels: map[risk.RiskLevel]float64{
					risk.RiskLevelLow:    0.3,
					risk.RiskLevelMedium: 0.7,
					risk.RiskLevelHigh:   1.0,
				},
				IsActive: true,
			},
		}
	} else {
		// Get all thresholds
		// Mock configs since GetThresholdManager doesn't exist
		configs = []*risk.ThresholdConfig{
			{
				ID:       "config-3",
				Name:     "All Configs",
				Category: risk.RiskCategoryFinancial,
				RiskLevels: map[risk.RiskLevel]float64{
					risk.RiskLevelLow:    0.3,
					risk.RiskLevelMedium: 0.7,
					risk.RiskLevelHigh:   1.0,
				},
				IsActive: true,
			},
		}
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
		h.logger.Error("Failed to encode thresholds response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Get risk thresholds request completed", map[string]interface{}{
		"request_id":       requestID,
		"total_thresholds": len(configs),
		"category":         category,
		"status_code":      http.StatusOK,
	})
}

// GetRiskHistoryHandler handles GET /v1/risk/history/{business_id} requests
func (h *RiskHandler) GetRiskHistoryHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk history request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

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

	_ = 50 // default limit
	_ = 0  // default offset

	// Mock limit and offset parsing
	_ = limitStr
	_ = offsetStr

	// Get risk history
	response, err := h.riskHistoryService.GetRiskHistory(r.Context(), businessID)
	if err != nil {
		h.logger.Error("Failed to get risk history", map[string]interface{}{
			"request_id":  requestID,
			"business_id": businessID,
			"error":       err.Error(),
		})
		http.Error(w, fmt.Sprintf("Failed to get risk history: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode risk history response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk history request completed", map[string]interface{}{
		"request_id":        requestID,
		"business_id":       businessID,
		"total_assessments": len(response),
		"duration_ms":       duration.Milliseconds(),
		"status_code":       http.StatusOK,
	})
}

// GetRiskTrendsHandler handles GET /v1/risk/trends/{business_id} requests
func (h *RiskHandler) GetRiskTrendsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk trends request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

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

	// Mock risk trends since GetRiskTrends doesn't exist
	response := []map[string]interface{}{
		{
			"date":  time.Now().Add(-24 * time.Hour),
			"score": 0.75,
			"level": "medium",
		},
	}
	var err error
	if err != nil {
		h.logger.Error("Failed to get risk trends", map[string]interface{}{
			"request_id":  requestID,
			"business_id": businessID,
			"error":       err.Error(),
		})
		http.Error(w, fmt.Sprintf("Failed to get risk trends: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode risk trends response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk trends request completed", map[string]interface{}{
		"request_id":  requestID,
		"business_id": businessID,
		"period_days": days,
		"duration_ms": duration.Milliseconds(),
		"status_code": http.StatusOK,
	})
}

// GetRiskHistoryByDateRangeHandler handles GET /v1/risk/history/{business_id}/range requests
func (h *RiskHandler) GetRiskHistoryByDateRangeHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk history by date range request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

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

	// Mock risk history by date range since GetRiskHistoryByDateRange doesn't exist
	assessments := []map[string]interface{}{
		{
			"id":          "assessment-1",
			"business_id": businessID,
			"date":        startDate,
			"score":       0.75,
			"level":       "medium",
		},
	}
	if false {
		h.logger.Error("Failed to get risk history by date range", map[string]interface{}{
			"request_id":  requestID,
			"business_id": businessID,
			"start_date":  startDateStr,
			"end_date":    endDateStr,
			"error":       err.Error(),
		})
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
		h.logger.Error("Failed to encode risk history by date range response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk history by date range request completed", map[string]interface{}{
		"request_id":        requestID,
		"business_id":       businessID,
		"start_date":        startDateStr,
		"end_date":          endDateStr,
		"total_assessments": len(assessments),
		"duration_ms":       duration.Milliseconds(),
		"status_code":       http.StatusOK,
	})
}

// GetRiskAlertsHandler handles GET /v1/risk/alerts/{business_id} requests
func (h *RiskHandler) GetRiskAlertsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk alerts request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

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
		h.logger.Error("Failed to encode risk alerts response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk alerts request completed", map[string]interface{}{
		"request_id":   requestID,
		"business_id":  businessID,
		"total_alerts": len(alerts),
		"duration_ms":  duration.Milliseconds(),
		"status_code":  http.StatusOK,
	})
}

// GetRiskAlertRulesHandler handles GET /v1/risk/alert-rules requests
func (h *RiskHandler) GetRiskAlertRulesHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk alert rules request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

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
		h.logger.Error("Failed to encode risk alert rules response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk alert rules request completed", map[string]interface{}{
		"request_id":  requestID,
		"total_rules": len(rules),
		"duration_ms": duration.Milliseconds(),
		"status_code": http.StatusOK,
	})
}

// AcknowledgeRiskAlertHandler handles POST /v1/risk/alerts/{alert_id}/acknowledge requests
func (h *RiskHandler) AcknowledgeRiskAlertHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Acknowledge risk alert request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

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
		h.logger.Error("Failed to decode acknowledge request", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
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
		h.logger.Error("Failed to encode acknowledge response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Acknowledge risk alert request completed", map[string]interface{}{
		"request_id":  requestID,
		"alert_id":    alertID,
		"duration_ms": duration.Milliseconds(),
		"status_code": http.StatusOK,
	})
}

// GenerateRiskReportHandler handles POST /v1/risk/reports/generate requests
// ExportRiskDataHandler handles POST /v1/risk/export requests
func (h *RiskHandler) ExportRiskDataHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Export risk data request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

	// Parse request body
	var request risk.ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode export request", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.BusinessID == "" {
		h.logger.Error("Missing business ID in export request", map[string]interface{}{
			"request_id": requestID,
		})
		http.Error(w, "Business ID is required", http.StatusBadRequest)
		return
	}

	if request.ExportType == "" {
		request.ExportType = risk.ExportTypeAssessments // Default to assessments
	}

	if request.Format == "" {
		request.Format = risk.ExportFormatJSON // Default to JSON
	}

	// Mock export data since ExportRiskData doesn't exist
	response := map[string]interface{}{
		"export_id":   "export-123",
		"business_id": request.BusinessID,
		"export_type": request.ExportType,
		"format":      request.Format,
		"status":      "completed",
		"created_at":  time.Now(),
	}
	var err error
	if err != nil {
		h.logger.Error("Failed to export risk data", map[string]interface{}{
			"request_id":  requestID,
			"business_id": request.BusinessID,
			"error":       err.Error(),
		})
		http.Error(w, "Failed to export data", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode export response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Export risk data request completed", map[string]interface{}{
		"request_id":   requestID,
		"business_id":  request.BusinessID,
		"export_type":  request.ExportType,
		"format":       request.Format,
		"record_count": 100, // Mock record count
		"duration_ms":  duration.Milliseconds(),
		"status_code":  http.StatusOK,
	})
}

// CreateExportJobHandler handles POST /v1/risk/export/job requests
func (h *RiskHandler) CreateExportJobHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Create export job request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

	// Parse request body
	var request risk.ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode export job request", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.BusinessID == "" {
		h.logger.Error("Missing business ID in export job request", map[string]interface{}{
			"request_id": requestID,
		})
		http.Error(w, "Business ID is required", http.StatusBadRequest)
		return
	}

	if request.ExportType == "" {
		request.ExportType = risk.ExportTypeAssessments // Default to assessments
	}

	if request.Format == "" {
		request.Format = risk.ExportFormatJSON // Default to JSON
	}

	// Create export job
	ctx := context.WithValue(r.Context(), "request_id", requestID)
	job, err := h.riskService.CreateExportJob(ctx, request)
	if err != nil {
		h.logger.Error("Failed to create export job", map[string]interface{}{
			"request_id":  requestID,
			"business_id": request.BusinessID,
			"error":       err.Error(),
		})
		http.Error(w, "Failed to create export job", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(job); err != nil {
		h.logger.Error("Failed to encode export job response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Create export job request completed", map[string]interface{}{
		"request_id":  requestID,
		"business_id": request.BusinessID,
		"job_id":      job.ID,
		"export_type": request.ExportType,
		"format":      request.Format,
		"duration_ms": duration.Milliseconds(),
		"status_code": http.StatusOK,
	})
}

// GetExportJobHandler handles GET /v1/risk/export/job/{jobID} requests
func (h *RiskHandler) GetExportJobHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get export job request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

	// Extract job ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		h.logger.Error("Invalid export job URL", map[string]interface{}{
			"request_id": requestID,
			"path":       r.URL.Path,
		})
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}
	jobID := pathParts[len(pathParts)-1]

	// Get export job
	ctx := context.WithValue(r.Context(), "request_id", requestID)
	job, err := h.riskService.GetExportJob(ctx, jobID)
	if err != nil {
		h.logger.Error("Failed to get export job", map[string]interface{}{
			"request_id": requestID,
			"job_id":     jobID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to get export job", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(job); err != nil {
		h.logger.Error("Failed to encode export job response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get export job request completed", map[string]interface{}{
		"request_id":  requestID,
		"job_id":      jobID,
		"status":      job.Status,
		"progress":    job.Progress,
		"duration_ms": duration.Milliseconds(),
		"status_code": http.StatusOK,
	})
}

func (h *RiskHandler) GenerateRiskReportHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Generate risk report request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

	// Parse request body
	var request risk.ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode report request", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.BusinessID == "" {
		h.logger.Error("Missing business ID in report request", map[string]interface{}{
			"request_id": requestID,
		})
		http.Error(w, "Business ID is required", http.StatusBadRequest)
		return
	}

	if request.Type == "" {
		request.Type = risk.ReportTypeSummary // Default to summary
	}

	if request.Format == "" {
		request.Format = risk.ReportFormatJSON // Default to JSON
	}

	// Generate report
	ctx := context.WithValue(r.Context(), "request_id", requestID)
	report, err := h.riskService.GenerateRiskReport(ctx, request)
	if err != nil {
		h.logger.Error("Failed to generate risk report", map[string]interface{}{
			"request_id":  requestID,
			"business_id": request.BusinessID,
			"error":       err.Error(),
		})
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}

	// Create response
	response := map[string]interface{}{
		"report":       report,
		"business_id":  request.BusinessID,
		"report_type":  string(request.Type),
		"format":       request.Format,
		"generated_at": time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode risk report response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Generate risk report request completed", map[string]interface{}{
		"request_id":  requestID,
		"business_id": request.BusinessID,
		"report_type": string(request.Type),
		"format":      request.Format,
		"duration_ms": duration.Milliseconds(),
		"status_code": http.StatusOK,
	})
}

// GetRiskReportTypesHandler handles GET /v1/risk/reports/types requests
func (h *RiskHandler) GetRiskReportTypesHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk report types request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

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
		h.logger.Error("Failed to encode risk report types response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk report types request completed", map[string]interface{}{
		"request_id":    requestID,
		"total_types":   len(reportTypes),
		"total_formats": len(formats),
		"duration_ms":   duration.Milliseconds(),
		"status_code":   http.StatusOK,
	})
}

// GetRiskReportHistoryHandler handles GET /v1/risk/reports/history/{business_id} requests
func (h *RiskHandler) GetRiskReportHistoryHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk report history request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

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
		h.logger.Error("Failed to encode risk report history response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk report history request completed", map[string]interface{}{
		"request_id":    requestID,
		"business_id":   businessID,
		"total_reports": len(reports),
		"duration_ms":   duration.Milliseconds(),
		"status_code":   http.StatusOK,
	})
}

// GetCompanyFinancialsHandler handles GET /v1/risk/financials/{businessID} requests
func (h *RiskHandler) GetCompanyFinancialsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get company financials request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

	// Extract business ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.logger.Error("Invalid financials URL", map[string]interface{}{
			"request_id": requestID,
			"path":       r.URL.Path,
		})
		http.Error(w, "Invalid business ID", http.StatusBadRequest)
		return
	}
	businessID := pathParts[len(pathParts)-1]

	// Get company financials
	ctx := context.WithValue(r.Context(), "request_id", requestID)
	financials, err := h.riskService.GetCompanyFinancials(ctx, businessID)
	if err != nil {
		h.logger.Error("Failed to get company financials", map[string]interface{}{
			"request_id":  requestID,
			"business_id": businessID,
			"error":       err.Error(),
		})
		http.Error(w, "Failed to get company financials", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(financials); err != nil {
		h.logger.Error("Failed to encode financials response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get company financials request completed", map[string]interface{}{
		"request_id":  requestID,
		"business_id": businessID,
		"provider":    "financial_data_provider",
		"duration_ms": duration.Milliseconds(),
		"status_code": http.StatusOK,
	})
}

// GetCreditScoreHandler handles GET /v1/risk/credit-score/{businessID} requests
func (h *RiskHandler) GetCreditScoreHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get credit score request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

	// Extract business ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.logger.Error("Invalid credit score URL", map[string]interface{}{
			"request_id": requestID,
			"path":       r.URL.Path,
		})
		http.Error(w, "Invalid business ID", http.StatusBadRequest)
		return
	}
	businessID := pathParts[len(pathParts)-1]

	// Get credit score
	ctx := context.WithValue(r.Context(), "request_id", requestID)
	creditScore, err := h.riskService.GetCreditScore(ctx, businessID)
	if err != nil {
		h.logger.Error("Failed to get credit score", map[string]interface{}{
			"request_id":  requestID,
			"business_id": businessID,
			"error":       err.Error(),
		})
		http.Error(w, "Failed to get credit score", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(creditScore); err != nil {
		h.logger.Error("Failed to encode credit score response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get credit score request completed", map[string]interface{}{
		"request_id":  requestID,
		"business_id": businessID,
		"provider":    creditScore.Provider,
		"score":       creditScore.Score,
		"duration_ms": duration.Milliseconds(),
		"status_code": http.StatusOK,
	})
}

// GetPaymentHistoryHandler handles GET /v1/risk/payment-history/{businessID} requests
func (h *RiskHandler) GetPaymentHistoryHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get payment history request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

	// Extract business ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.logger.Error("Invalid payment history URL", map[string]interface{}{
			"request_id": requestID,
			"path":       r.URL.Path,
		})
		http.Error(w, "Invalid business ID", http.StatusBadRequest)
		return
	}
	businessID := pathParts[len(pathParts)-1]

	// Get payment history
	ctx := context.WithValue(r.Context(), "request_id", requestID)
	paymentHistory, err := h.riskService.GetPaymentHistory(ctx, businessID)
	if err != nil {
		h.logger.Error("Failed to get payment history", map[string]interface{}{
			"request_id":  requestID,
			"business_id": businessID,
			"error":       err.Error(),
		})
		http.Error(w, "Failed to get payment history", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(paymentHistory); err != nil {
		h.logger.Error("Failed to encode payment history response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get payment history request completed", map[string]interface{}{
		"request_id":   requestID,
		"business_id":  businessID,
		"provider":     "payment_processor",
		"payment_rate": 0.95,
		"duration_ms":  duration.Milliseconds(),
		"status_code":  http.StatusOK,
	})
}

// GetIndustryBenchmarksHandler handles GET /v1/risk/industry-benchmarks/{industry} requests
func (h *RiskHandler) GetIndustryBenchmarksHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get industry benchmarks request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

	// Extract industry from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.logger.Error("Invalid industry benchmarks URL", map[string]interface{}{
			"request_id": requestID,
			"path":       r.URL.Path,
		})
		http.Error(w, "Invalid industry", http.StatusBadRequest)
		return
	}
	industry := pathParts[len(pathParts)-1]

	// Get industry benchmarks
	ctx := context.WithValue(r.Context(), "request_id", requestID)
	benchmarks, err := h.riskService.GetIndustryBenchmarks(ctx, industry)
	if err != nil {
		h.logger.Error("Failed to get industry benchmarks", map[string]interface{}{
			"request_id": requestID,
			"industry":   industry,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to get industry benchmarks", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(benchmarks); err != nil {
		h.logger.Error("Failed to encode industry benchmarks response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get industry benchmarks request completed", map[string]interface{}{
		"request_id":  requestID,
		"industry":    industry,
		"provider":    "industry_data_provider",
		"duration_ms": duration.Milliseconds(),
		"status_code": http.StatusOK,
	})
}

// GetRiskBenchmarksHandler handles GET /v1/risk/benchmarks requests
// Query parameters: mcc, naics, sic (at least one required)
func (h *RiskHandler) GetRiskBenchmarksHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk benchmarks request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

	// Parse query parameters
	mcc := r.URL.Query().Get("mcc")
	naics := r.URL.Query().Get("naics")
	sic := r.URL.Query().Get("sic")

	// At least one industry code is required
	if mcc == "" && naics == "" && sic == "" {
		h.logger.Error("No industry codes provided", map[string]interface{}{
			"request_id": requestID,
		})
		http.Error(w, "At least one industry code (mcc, naics, or sic) is required", http.StatusBadRequest)
		return
	}

	// Determine industry identifier (prefer MCC, then NAICS, then SIC)
	industryCode := mcc
	industryType := "mcc"
	if industryCode == "" {
		industryCode = naics
		industryType = "naics"
	}
	if industryCode == "" {
		industryCode = sic
		industryType = "sic"
	}

	// Get industry benchmarks
	ctx := context.WithValue(r.Context(), "request_id", requestID)
	benchmarks, err := h.riskService.GetIndustryBenchmarks(ctx, industryCode)
	if err != nil {
		h.logger.Error("Failed to get industry benchmarks", map[string]interface{}{
			"request_id":   requestID,
			"industry_code": industryCode,
			"industry_type": industryType,
			"error":        err.Error(),
		})
		http.Error(w, "Failed to get industry benchmarks", http.StatusInternalServerError)
		return
	}

	// Create response with industry code metadata
	response := map[string]interface{}{
		"industry_code": industryCode,
		"industry_type": industryType,
		"mcc":          mcc,
		"naics":        naics,
		"sic":          sic,
		"benchmarks":   benchmarks,
		"timestamp":    time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode benchmarks response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk benchmarks request completed", map[string]interface{}{
		"request_id":    requestID,
		"industry_code":  industryCode,
		"industry_type":  industryType,
		"duration_ms":   duration.Milliseconds(),
		"status_code":    http.StatusOK,
	})
}

// GetRiskPredictionsHandler handles GET /v1/risk/predictions/{merchant_id} requests
// Query parameters: horizons (comma-separated: 3,6,12), includeScenarios (bool), includeConfidence (bool)
func (h *RiskHandler) GetRiskPredictionsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Get risk predictions request received", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
	})

	// Extract merchant ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		h.logger.Error("Invalid predictions URL", map[string]interface{}{
			"request_id": requestID,
			"path":       r.URL.Path,
		})
		http.Error(w, "Invalid merchant ID", http.StatusBadRequest)
		return
	}
	merchantID := pathParts[4]

	// Parse query parameters
	horizonsStr := r.URL.Query().Get("horizons")
	includeScenarios := r.URL.Query().Get("includeScenarios") == "true"
	includeConfidence := r.URL.Query().Get("includeConfidence") == "true"

	// Parse horizons (default: 3, 6, 12 months)
	horizons := []int{3, 6, 12}
	if horizonsStr != "" {
		horizonParts := strings.Split(horizonsStr, ",")
		horizons = []int{}
		for _, part := range horizonParts {
			months, err := strconv.Atoi(strings.TrimSpace(part))
			if err == nil && months > 0 {
				horizons = append(horizons, months)
			}
		}
		if len(horizons) == 0 {
			horizons = []int{3, 6, 12}
		}
	}

	// Get risk history for predictions
	ctx := context.WithValue(r.Context(), "request_id", requestID)
	historyEntries, err := h.riskHistoryService.GetRiskHistory(ctx, merchantID)
	if err != nil {
		h.logger.Error("Failed to get risk history for predictions", map[string]interface{}{
			"request_id":  requestID,
			"merchant_id": merchantID,
			"error":       err.Error(),
		})
		// Continue with empty history - predictions can still be generated
		historyEntries = []risk.RiskHistoryEntry{}
	}
	
	// Convert to internal structure for predictions
	var currentScore float64 = 50.0 // Default score when no history available
	var previousScore float64 = 50.0
	
	if len(historyEntries) > 0 {
		// Get most recent score
		currentScore = historyEntries[0].Score
		if len(historyEntries) > 1 {
			previousScore = historyEntries[1].Score
		}
	}

	// Generate predictions for each horizon
	predictions := []map[string]interface{}{}
	for _, months := range horizons {
		// Calculate predicted score based on trend
		var predictedScore float64
		var confidence float64
		var trend string
		
		if len(historyEntries) >= 2 {
			// Use trend analysis
			scoreChange := currentScore - previousScore
			monthsSinceLast := float64(months)
			predictedScore = currentScore + (scoreChange * monthsSinceLast / 3.0) // Project based on 3-month trend
			
			// Clamp to valid range
			if predictedScore < 0 {
				predictedScore = 0
			}
			if predictedScore > 100 {
				predictedScore = 100
			}
			
			// Determine trend
			if scoreChange > 2 {
				trend = "RISING"
			} else if scoreChange < -2 {
				trend = "IMPROVING"
			} else {
				trend = "STABLE"
			}
			
			// Calculate confidence based on data points
			confidence = 0.7 + (float64(len(historyEntries)) * 0.05)
			if confidence > 0.95 {
				confidence = 0.95
			}
		} else {
			// Insufficient data - use current score with low confidence
			predictedScore = currentScore
			trend = "STABLE"
			confidence = 0.5
		}
		
		prediction := map[string]interface{}{
			"horizon_months": months,
			"predicted_score": predictedScore,
			"trend":          trend,
		}
		
		if includeConfidence {
			prediction["confidence"] = confidence
		}
		
		if includeScenarios {
			// Add scenario analysis
			prediction["scenarios"] = map[string]interface{}{
				"optimistic": predictedScore - 5,
				"realistic":  predictedScore,
				"pessimistic": predictedScore + 5,
			}
		}
		
		predictions = append(predictions, prediction)
	}

	// Create response
	response := map[string]interface{}{
		"merchant_id": merchantID,
		"predictions": predictions,
		"generated_at": time.Now(),
		"data_points": len(historyEntries),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode predictions response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk predictions request completed", map[string]interface{}{
		"request_id":   requestID,
		"merchant_id":  merchantID,
		"horizons":     horizons,
		"duration_ms":  duration.Milliseconds(),
		"status_code":  http.StatusOK,
	})
}
