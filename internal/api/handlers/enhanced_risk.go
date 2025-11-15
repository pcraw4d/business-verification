package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"kyb-platform/internal/risk"
)

// EnhancedRiskHandler handles enhanced risk assessment API endpoints
type EnhancedRiskHandler struct {
	logger               *zap.Logger
	riskService          *risk.RiskDetectionService // TODO: RiskAssessmentService doesn't exist, using RiskDetectionService as stub
	enhancedCalculator   *risk.EnhancedRiskCalculator
	recommendationEngine *risk.RiskRecommendationEngine
	trendAnalysisService *risk.RiskTrendAnalysisService
	alertSystem          *risk.RiskAlertSystem
	thresholdManager     *risk.ThresholdManager
}

// NewEnhancedRiskHandler creates a new enhanced risk handler
func NewEnhancedRiskHandler(
	logger *zap.Logger,
	riskService *risk.RiskDetectionService, // TODO: RiskAssessmentService doesn't exist, using RiskDetectionService as stub
	enhancedCalculator *risk.EnhancedRiskCalculator,
	recommendationEngine *risk.RiskRecommendationEngine,
	trendAnalysisService *risk.RiskTrendAnalysisService,
	alertSystem *risk.RiskAlertSystem,
	thresholdManager *risk.ThresholdManager,
) *EnhancedRiskHandler {
	return &EnhancedRiskHandler{
		logger:               logger,
		riskService:          riskService,
		enhancedCalculator:   enhancedCalculator,
		recommendationEngine: recommendationEngine,
		trendAnalysisService: trendAnalysisService,
		alertSystem:          alertSystem,
		thresholdManager:     thresholdManager,
	}
}

// EnhancedRiskAssessmentHandler handles POST /v1/risk/enhanced/assess requests
func (h *EnhancedRiskHandler) EnhancedRiskAssessmentHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	h.logger.Info("Enhanced risk assessment request received",
		zap.String("request_id", requestID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	// Parse request body
	var request risk.EnhancedRiskAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateEnhancedRiskRequest(request); err != nil {
		h.logger.Error("Invalid enhanced risk assessment request",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Perform enhanced risk assessment
	response, err := h.performEnhancedRiskAssessment(r.Context(), request)
	if err != nil {
		h.logger.Error("Enhanced risk assessment failed",
			zap.String("request_id", requestID),
			zap.String("business_id", request.BusinessID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Risk assessment failed: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Enhanced risk assessment request completed",
		zap.String("request_id", requestID),
		zap.Duration("duration", duration))
}

// RiskFactorCalculationHandler handles POST /v1/risk/factors/calculate requests
func (h *EnhancedRiskHandler) RiskFactorCalculationHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	h.logger.Info("Risk factor calculation request received",
		zap.String("request_id", requestID))

	// Parse request body
	var request risk.EnhancedRiskFactorInput
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateRiskFactorRequest(request); err != nil {
		h.logger.Error("Invalid risk factor request",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Perform enhanced risk factor calculation
	result, err := h.enhancedCalculator.CalculateEnhancedFactor(r.Context(), request)
	if err != nil {
		h.logger.Error("Risk factor calculation failed",
			zap.String("request_id", requestID),
			zap.String("factor_id", request.FactorID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Calculation failed: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Risk factor calculation completed",
		zap.String("request_id", requestID),
		zap.String("factor_id", request.FactorID),
		zap.Duration("duration", duration))
}

// RiskRecommendationsHandler handles POST /v1/risk/recommendations requests
func (h *EnhancedRiskHandler) RiskRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	h.logger.Info("Risk recommendations request received",
		zap.String("request_id", requestID))

	// Parse request body
	var request risk.RecommendationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateRecommendationRequest(request); err != nil {
		h.logger.Error("Invalid recommendation request",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Generate recommendations
	response, err := h.recommendationEngine.GenerateRecommendations(r.Context(), request)
	if err != nil {
		h.logger.Error("Recommendation generation failed",
			zap.String("request_id", requestID),
			zap.String("business_id", request.BusinessID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Recommendation generation failed: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Risk recommendations completed",
		zap.String("request_id", requestID),
		zap.Int("recommendations", len(response.Recommendations)),
		zap.Duration("duration", duration))
}

// RiskTrendAnalysisHandler handles POST /v1/risk/trends/analyze requests
func (h *EnhancedRiskHandler) RiskTrendAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	h.logger.Info("Risk trend analysis request received",
		zap.String("request_id", requestID))

	// Parse request body
	var request risk.RiskTrendAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateTrendAnalysisRequest(request); err != nil {
		h.logger.Error("Invalid trend analysis request",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Perform trend analysis
	response, err := h.trendAnalysisService.AnalyzeTrends(r.Context(), request)
	if err != nil {
		h.logger.Error("Trend analysis failed",
			zap.String("request_id", requestID),
			zap.String("business_id", request.BusinessID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Trend analysis failed: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Risk trend analysis completed",
		zap.String("request_id", requestID),
		zap.Int("trends", len(response.Trends)),
		zap.Duration("duration", duration))
}

// RiskAlertsHandler handles GET /v1/risk/alerts requests
func (h *EnhancedRiskHandler) RiskAlertsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	h.logger.Info("Risk alerts request received",
		zap.String("request_id", requestID))

	// Parse query parameters
	businessID := r.URL.Query().Get("business_id")
	if businessID == "" {
		http.Error(w, "business_id parameter is required", http.StatusBadRequest)
		return
	}

	// Get active alerts
	alerts, err := h.alertSystem.GetActiveAlerts(r.Context(), businessID)
	if err != nil {
		h.logger.Error("Failed to get active alerts",
			zap.String("request_id", requestID),
			zap.String("business_id", businessID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to get alerts: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	response := map[string]interface{}{
		"business_id": businessID,
		"alerts":      alerts,
		"count":       len(alerts),
		"timestamp":   time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Risk alerts request completed",
		zap.String("request_id", requestID),
		zap.String("business_id", businessID),
		zap.Int("alert_count", len(alerts)),
		zap.Duration("duration", duration))
}

// AcknowledgeAlertHandler handles POST /v1/risk/alerts/{alert_id}/acknowledge requests
func (h *EnhancedRiskHandler) AcknowledgeAlertHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	// Extract alert ID from URL path
	alertID := strings.TrimPrefix(r.URL.Path, "/v1/risk/alerts/")
	alertID = strings.TrimSuffix(alertID, "/acknowledge")

	if alertID == "" {
		http.Error(w, "Alert ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Acknowledge alert request received",
		zap.String("request_id", requestID),
		zap.String("alert_id", alertID))

	// Parse request body for user ID
	var request struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	// Acknowledge alert
	if err := h.alertSystem.AcknowledgeAlert(r.Context(), alertID, request.UserID); err != nil {
		h.logger.Error("Failed to acknowledge alert",
			zap.String("request_id", requestID),
			zap.String("alert_id", alertID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to acknowledge alert: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	response := map[string]interface{}{
		"alert_id":  alertID,
		"user_id":   request.UserID,
		"status":    "acknowledged",
		"timestamp": time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Alert acknowledged",
		zap.String("request_id", requestID),
		zap.String("alert_id", alertID),
		zap.String("user_id", request.UserID),
		zap.Duration("duration", duration))
}

// ResolveAlertHandler handles POST /v1/risk/alerts/{alert_id}/resolve requests
func (h *EnhancedRiskHandler) ResolveAlertHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	// Extract alert ID from URL path
	alertID := strings.TrimPrefix(r.URL.Path, "/v1/risk/alerts/")
	alertID = strings.TrimSuffix(alertID, "/resolve")

	if alertID == "" {
		http.Error(w, "Alert ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Resolve alert request received",
		zap.String("request_id", requestID),
		zap.String("alert_id", alertID))

	// Parse request body
	var request struct {
		UserID     string `json:"user_id"`
		Resolution string `json:"resolution"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	// Resolve alert
	if err := h.alertSystem.ResolveAlert(r.Context(), alertID, request.UserID, request.Resolution); err != nil {
		h.logger.Error("Failed to resolve alert",
			zap.String("request_id", requestID),
			zap.String("alert_id", alertID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to resolve alert: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	response := map[string]interface{}{
		"alert_id":   alertID,
		"user_id":    request.UserID,
		"resolution": request.Resolution,
		"status":     "resolved",
		"timestamp":  time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Alert resolved",
		zap.String("request_id", requestID),
		zap.String("alert_id", alertID),
		zap.String("user_id", request.UserID),
		zap.Duration("duration", duration))
}

// RiskFactorHistoryHandler handles GET /v1/risk/factors/{factor_id}/history requests
func (h *EnhancedRiskHandler) RiskFactorHistoryHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	// Extract factor ID from URL path
	factorID := strings.TrimPrefix(r.URL.Path, "/v1/risk/factors/")
	factorID = strings.TrimSuffix(factorID, "/history")

	if factorID == "" {
		http.Error(w, "Factor ID is required", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	businessID := r.URL.Query().Get("business_id")
	if businessID == "" {
		http.Error(w, "business_id parameter is required", http.StatusBadRequest)
		return
	}

	// Parse date range
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30) // Default to 30 days ago
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	} else {
		endDate = time.Now() // Default to now
	}

	h.logger.Info("Risk factor history request received",
		zap.String("request_id", requestID),
		zap.String("factor_id", factorID),
		zap.String("business_id", businessID))

	// Get factor history
	history, err := h.trendAnalysisService.GetLatestRiskData(r.Context(), businessID, factorID)
	if err != nil {
		h.logger.Error("Failed to get factor history",
			zap.String("request_id", requestID),
			zap.String("factor_id", factorID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to get factor history: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	response := map[string]interface{}{
		"factor_id":   factorID,
		"business_id": businessID,
		"start_date":  startDate,
		"end_date":    endDate,
		"history":     history,
		"timestamp":   time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Risk factor history request completed",
		zap.String("request_id", requestID),
		zap.String("factor_id", factorID),
		zap.Duration("duration", duration))
}

// Validation methods
func (h *EnhancedRiskHandler) validateEnhancedRiskRequest(request risk.EnhancedRiskAssessmentRequest) error {
	if request.BusinessID == "" {
		return fmt.Errorf("business_id is required")
	}
	// TODO: EnhancedRiskAssessmentRequest doesn't have BusinessName field
	// It has BusinessID, AssessmentID, RiskFactorInputs, etc.
	// Stub: skip BusinessName validation
	return nil
}

func (h *EnhancedRiskHandler) validateRiskFactorRequest(request risk.EnhancedRiskFactorInput) error {
	if request.FactorID == "" {
		return fmt.Errorf("factor_id is required")
	}
	if request.Source == "" {
		return fmt.Errorf("source is required")
	}
	if request.Reliability < 0 || request.Reliability > 1 {
		return fmt.Errorf("reliability must be between 0 and 1")
	}
	return nil
}

func (h *EnhancedRiskHandler) validateRecommendationRequest(request risk.RecommendationRequest) error {
	if request.BusinessID == "" {
		return fmt.Errorf("business_id is required")
	}
	if request.RiskAssessment == nil && len(request.RiskFactors) == 0 {
		return fmt.Errorf("either risk_assessment or risk_factors is required")
	}
	return nil
}

func (h *EnhancedRiskHandler) validateTrendAnalysisRequest(request risk.RiskTrendAnalysisRequest) error {
	if request.BusinessID == "" {
		return fmt.Errorf("business_id is required")
	}
	return nil
}

// performEnhancedRiskAssessment performs the enhanced risk assessment
func (h *EnhancedRiskHandler) performEnhancedRiskAssessment(ctx context.Context, request risk.EnhancedRiskAssessmentRequest) (*risk.EnhancedRiskAssessmentResponse, error) {
	// Start timing at the beginning of the function
	startTime := time.Now()

	// This would integrate all the enhanced services
	// For now, we'll create a basic response structure

	response := &risk.EnhancedRiskAssessmentResponse{
		AssessmentID:     request.AssessmentID,
		BusinessID:       request.BusinessID,
		Timestamp:        time.Now(),
		OverallRiskScore: 75.0, // Placeholder
		OverallRiskLevel: risk.RiskLevelHigh,
		RiskFactors:      []risk.RiskFactorDetail{},
		Recommendations:  []risk.RecommendationDetail{},
		Alerts:           []risk.AlertDetail{},
		ConfidenceScore:  0.85,
		ProcessingTimeMs: int64(time.Since(startTime).Milliseconds()),
	}

	return response, nil
}

// GetRiskFactorsHandler handles GET /v1/risk/factors requests
func (h *EnhancedRiskHandler) GetRiskFactorsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	h.logger.Info("Get risk factors request received",
		zap.String("request_id", requestID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	// Parse query parameters
	category := r.URL.Query().Get("category")

	// Get risk factors (mock data for now - TODO: integrate with actual service)
	factors := []risk.RiskFactor{
		{
			ID:          "financial_stability",
			Name:        "Financial Stability",
			Description: "Measures the financial health and stability of the business",
			Category:    risk.RiskCategoryFinancial,
			Weight:      0.3,
			Thresholds: map[risk.RiskLevel]float64{
				risk.RiskLevelLow:      25.0,
				risk.RiskLevelMedium:   50.0,
				risk.RiskLevelHigh:     75.0,
				risk.RiskLevelCritical: 90.0,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "operational_efficiency",
			Name:        "Operational Efficiency",
			Description: "Assesses the operational efficiency and process quality",
			Category:    risk.RiskCategoryOperational,
			Weight:      0.25,
			Thresholds: map[risk.RiskLevel]float64{
				risk.RiskLevelLow:      20.0,
				risk.RiskLevelMedium:   45.0,
				risk.RiskLevelHigh:     70.0,
				risk.RiskLevelCritical: 85.0,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "regulatory_compliance",
			Name:        "Regulatory Compliance",
			Description: "Evaluates compliance with regulatory requirements",
			Category:    risk.RiskCategoryRegulatory,
			Weight:      0.25,
			Thresholds: map[risk.RiskLevel]float64{
				risk.RiskLevelLow:      30.0,
				risk.RiskLevelMedium:   55.0,
				risk.RiskLevelHigh:     80.0,
				risk.RiskLevelCritical: 95.0,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "cybersecurity_posture",
			Name:        "Cybersecurity Posture",
			Description: "Measures the cybersecurity readiness and protection level",
			Category:    risk.RiskCategoryCybersecurity,
			Weight:      0.2,
			Thresholds: map[risk.RiskLevel]float64{
				risk.RiskLevelLow:      15.0,
				risk.RiskLevelMedium:   40.0,
				risk.RiskLevelHigh:     75.0,
				risk.RiskLevelCritical: 90.0,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Filter by category if provided
	if category != "" {
		filtered := []risk.RiskFactor{}
		for _, factor := range factors {
			if string(factor.Category) == category {
				filtered = append(filtered, factor)
			}
		}
		factors = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"factors":   factors,
		"count":     len(factors),
		"timestamp": time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk factors request completed",
		zap.String("request_id", requestID),
		zap.Duration("duration", duration))
}

// GetRiskCategoriesHandler handles GET /v1/risk/categories requests
func (h *EnhancedRiskHandler) GetRiskCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	h.logger.Info("Get risk categories request received",
		zap.String("request_id", requestID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	// Return all risk categories
	categories := []map[string]interface{}{
		{
			"category":    string(risk.RiskCategoryFinancial),
			"name":        "Financial Risk",
			"description": "Risks related to financial stability, liquidity, and creditworthiness",
		},
		{
			"category":    string(risk.RiskCategoryOperational),
			"name":        "Operational Risk",
			"description": "Risks related to business operations, processes, and internal controls",
		},
		{
			"category":    string(risk.RiskCategoryRegulatory),
			"name":        "Regulatory Risk",
			"description": "Risks related to compliance with laws, regulations, and industry standards",
		},
		{
			"category":    string(risk.RiskCategoryReputational),
			"name":        "Reputational Risk",
			"description": "Risks related to brand reputation and public perception",
		},
		{
			"category":    string(risk.RiskCategoryCybersecurity),
			"name":        "Cybersecurity Risk",
			"description": "Risks related to information security and data protection",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"categories": categories,
		"count":      len(categories),
		"timestamp":  time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk categories request completed",
		zap.String("request_id", requestID),
		zap.Duration("duration", duration))
}

// GetRiskThresholdsHandler handles GET /v1/risk/thresholds requests
func (h *EnhancedRiskHandler) GetRiskThresholdsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	h.logger.Info("Get risk thresholds request received",
		zap.String("request_id", requestID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	// Parse query parameters
	category := r.URL.Query().Get("category")
	industryCode := r.URL.Query().Get("industry_code")

	// Get thresholds from ThresholdManager
	var configs []*risk.ThresholdConfig
	if h.thresholdManager == nil {
		// Fallback to empty list if threshold manager is not initialized
		h.logger.Warn("ThresholdManager is not initialized, returning empty thresholds",
			zap.String("request_id", requestID))
		configs = []*risk.ThresholdConfig{}
	} else {
		if category != "" {
			// Get configs by category
			configs = h.thresholdManager.GetConfigsByCategory(risk.RiskCategory(category))
		} else if industryCode != "" {
			// Get configs by industry
			configs = h.thresholdManager.GetConfigsByIndustry(industryCode)
		} else {
			// Get all configs
			configs = h.thresholdManager.ListConfigs()
		}
	}

	// Convert ThresholdConfig to RiskThreshold for response
	thresholds := make([]risk.RiskThreshold, 0, len(configs))
	for _, config := range configs {
		if !config.IsActive {
			continue // Skip inactive configs
		}

		// Extract risk level values from config
		lowMax := config.RiskLevels[risk.RiskLevelLow]
		mediumMax := config.RiskLevels[risk.RiskLevelMedium]
		highMax := config.RiskLevels[risk.RiskLevelHigh]
		criticalMin := config.RiskLevels[risk.RiskLevelCritical]

		thresholds = append(thresholds, risk.RiskThreshold{
			Category:    config.Category,
			LowMax:      lowMax,
			MediumMax:   mediumMax,
			HighMax:     highMax,
			CriticalMin: criticalMin,
			UpdatedAt:   config.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"thresholds": thresholds,
		"count":      len(thresholds),
		"category":   category,
		"industry":   industryCode,
		"timestamp":  time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get risk thresholds request completed",
		zap.String("request_id", requestID),
		zap.Duration("duration", duration))
	_ = industryCode // Suppress unused variable warning
}

// CreateRiskThresholdHandler handles POST /v1/admin/risk/thresholds requests
func (h *EnhancedRiskHandler) CreateRiskThresholdHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	h.logger.Info("Create risk threshold request received",
		zap.String("request_id", requestID))

	var req risk.ThresholdConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" || req.Category == "" || len(req.RiskLevels) == 0 {
		http.Error(w, "Name, category, and risk_levels are required", http.StatusBadRequest)
		return
	}

	// Check if ThresholdManager is available
	if h.thresholdManager == nil {
		h.logger.Error("ThresholdManager is not initialized",
			zap.String("request_id", requestID))
		http.Error(w, "Threshold management service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Create threshold configuration
	// Default is_active to true if not explicitly set
	isActive := true
	if req.IsActive != nil {
		// Use the explicitly provided value
		isActive = *req.IsActive
	}

	config := &risk.ThresholdConfig{
		ID:             req.ID,
		Name:           req.Name,
		Description:    req.Description,
		Category:       req.Category,
		IndustryCode:   req.IndustryCode,
		BusinessType:   req.BusinessType,
		RiskLevels:     req.RiskLevels,
		IsDefault:      req.IsDefault,
		IsActive:       isActive,
		Priority:       req.Priority,
		Metadata:       req.Metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      req.CreatedBy,
		LastModifiedBy: req.CreatedBy,
	}

	if config.ID == "" {
		// Generate a UUID for the threshold ID
		config.ID = uuid.New().String()
	}

	// Register the configuration with ThresholdManager
	if err := h.thresholdManager.RegisterConfig(config); err != nil {
		h.logger.Error("Failed to register threshold configuration",
			zap.String("request_id", requestID),
			zap.String("threshold_id", config.ID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to create threshold: %s", err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusCreated)

	response := map[string]interface{}{
		"id":        config.ID,
		"config":    config,
		"timestamp": time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Create risk threshold request completed",
		zap.String("request_id", requestID),
		zap.String("threshold_id", config.ID),
		zap.Duration("duration", duration))
}

// UpdateRiskThresholdHandler handles PUT /v1/admin/risk/thresholds/{threshold_id} requests
func (h *EnhancedRiskHandler) UpdateRiskThresholdHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	// Extract threshold ID from path
	thresholdID := extractIDFromPath(r.URL.Path, "/v1/admin/risk/thresholds/")
	if thresholdID == "" {
		http.Error(w, "Threshold ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Update risk threshold request received",
		zap.String("request_id", requestID),
		zap.String("threshold_id", thresholdID))

	// Check if ThresholdManager is available
	if h.thresholdManager == nil {
		h.logger.Error("ThresholdManager is not initialized",
			zap.String("request_id", requestID))
		http.Error(w, "Threshold management service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Check if threshold exists
	_, exists := h.thresholdManager.GetConfig(thresholdID)
	if !exists {
		http.Error(w, fmt.Sprintf("Threshold with ID %s not found", thresholdID), http.StatusNotFound)
		return
	}

	var req risk.ThresholdConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build updates map for ThresholdManager.UpdateConfig
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if len(req.RiskLevels) > 0 {
		updates["risk_levels"] = req.RiskLevels
	}
	// Only update is_active if explicitly provided
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	updates["is_default"] = req.IsDefault
	updates["priority"] = req.Priority
	if req.Metadata != nil {
		updates["metadata"] = req.Metadata
	}
	if req.CreatedBy != "" {
		updates["last_modified_by"] = req.CreatedBy
	}

	// Update the configuration
	if err := h.thresholdManager.UpdateConfig(thresholdID, updates); err != nil {
		h.logger.Error("Failed to update threshold configuration",
			zap.String("request_id", requestID),
			zap.String("threshold_id", thresholdID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update threshold: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Get updated config for response
	config, _ := h.thresholdManager.GetConfig(thresholdID)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"id":        config.ID,
		"config":    config,
		"timestamp": time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Update risk threshold request completed",
		zap.String("request_id", requestID),
		zap.String("threshold_id", thresholdID),
		zap.Duration("duration", duration))
}

// DeleteRiskThresholdHandler handles DELETE /v1/admin/risk/thresholds/{threshold_id} requests
func (h *EnhancedRiskHandler) DeleteRiskThresholdHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	// Extract threshold ID from path
	thresholdID := extractIDFromPath(r.URL.Path, "/v1/admin/risk/thresholds/")
	if thresholdID == "" {
		http.Error(w, "Threshold ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Delete risk threshold request received",
		zap.String("request_id", requestID),
		zap.String("threshold_id", thresholdID))

	// Check if ThresholdManager is available
	if h.thresholdManager == nil {
		h.logger.Error("ThresholdManager is not initialized",
			zap.String("request_id", requestID))
		http.Error(w, "Threshold management service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Check if threshold exists
	_, exists := h.thresholdManager.GetConfig(thresholdID)
	if !exists {
		http.Error(w, fmt.Sprintf("Threshold with ID %s not found", thresholdID), http.StatusNotFound)
		return
	}

	// Delete the configuration
	if err := h.thresholdManager.DeleteConfig(thresholdID); err != nil {
		h.logger.Error("Failed to delete threshold configuration",
			zap.String("request_id", requestID),
			zap.String("threshold_id", thresholdID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to delete threshold: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"id":        thresholdID,
		"deleted":   true,
		"timestamp": time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Delete risk threshold request completed",
		zap.String("request_id", requestID),
		zap.String("threshold_id", thresholdID),
		zap.Duration("duration", duration))
}

// CreateRecommendationRuleHandler handles POST /v1/admin/risk/recommendation-rules requests
func (h *EnhancedRiskHandler) CreateRecommendationRuleHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	h.logger.Info("Create recommendation rule request received",
		zap.String("request_id", requestID))

	// Check if recommendation engine is available
	if h.recommendationEngine == nil {
		h.logger.Error("Recommendation engine not available",
			zap.String("request_id", requestID))
		http.Error(w, "Recommendation engine not available", http.StatusServiceUnavailable)
		return
	}

	// Parse request body
	var rule risk.RecommendationRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		h.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}

	// Get rule engine and add the rule
	ruleEngine := h.recommendationEngine.GetRuleEngine()
	if err := ruleEngine.AddCustomRule(rule); err != nil {
		h.logger.Error("Failed to create recommendation rule",
			zap.String("request_id", requestID),
			zap.String("rule_id", rule.ID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to create rule: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Return created rule
	response := map[string]interface{}{
		"id":        rule.ID,
		"name":      rule.Name,
		"category":  rule.Category,
		"enabled":   rule.Enabled,
		"priority":  rule.Priority,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Create recommendation rule request completed",
		zap.String("request_id", requestID),
		zap.String("rule_id", rule.ID),
		zap.Duration("duration", duration))
}

// UpdateRecommendationRuleHandler handles PUT /v1/admin/risk/recommendation-rules/{rule_id} requests
func (h *EnhancedRiskHandler) UpdateRecommendationRuleHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	ruleID := extractIDFromPath(r.URL.Path, "/v1/admin/risk/recommendation-rules/")
	if ruleID == "" {
		http.Error(w, "Rule ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Update recommendation rule request received",
		zap.String("request_id", requestID),
		zap.String("rule_id", ruleID))

	// Check if recommendation engine is available
	if h.recommendationEngine == nil {
		h.logger.Error("Recommendation engine not available",
			zap.String("request_id", requestID))
		http.Error(w, "Recommendation engine not available", http.StatusServiceUnavailable)
		return
	}

	// Parse request body
	var updatedRule risk.RecommendationRule
	if err := json.NewDecoder(r.Body).Decode(&updatedRule); err != nil {
		h.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure the rule ID matches the path parameter
	updatedRule.ID = ruleID

	// Get rule engine and update the rule
	ruleEngine := h.recommendationEngine.GetRuleEngine()
	if err := ruleEngine.UpdateRule(ruleID, updatedRule); err != nil {
		h.logger.Error("Failed to update recommendation rule",
			zap.String("request_id", requestID),
			zap.String("rule_id", ruleID),
			zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Rule not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to update rule: %s", err.Error()), http.StatusBadRequest)
		}
		return
	}

	// Return updated rule
	response := map[string]interface{}{
		"id":        ruleID,
		"name":      updatedRule.Name,
		"category":  updatedRule.Category,
		"enabled":   updatedRule.Enabled,
		"priority":  updatedRule.Priority,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Update recommendation rule request completed",
		zap.String("request_id", requestID),
		zap.String("rule_id", ruleID),
		zap.Duration("duration", duration))
}

// DeleteRecommendationRuleHandler handles DELETE /v1/admin/risk/recommendation-rules/{rule_id} requests
func (h *EnhancedRiskHandler) DeleteRecommendationRuleHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	ruleID := extractIDFromPath(r.URL.Path, "/v1/admin/risk/recommendation-rules/")
	if ruleID == "" {
		http.Error(w, "Rule ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Delete recommendation rule request received",
		zap.String("request_id", requestID),
		zap.String("rule_id", ruleID))

	// Check if recommendation engine is available
	if h.recommendationEngine == nil {
		h.logger.Error("Recommendation engine not available",
			zap.String("request_id", requestID))
		http.Error(w, "Recommendation engine not available", http.StatusServiceUnavailable)
		return
	}

	// Get rule engine and delete the rule
	ruleEngine := h.recommendationEngine.GetRuleEngine()
	if err := ruleEngine.DeleteRule(ruleID); err != nil {
		h.logger.Error("Failed to delete recommendation rule",
			zap.String("request_id", requestID),
			zap.String("rule_id", ruleID),
			zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Rule not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete rule: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	response := map[string]interface{}{
		"id":        ruleID,
		"deleted":   true,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Delete recommendation rule request completed",
		zap.String("request_id", requestID),
		zap.String("rule_id", ruleID),
		zap.Duration("duration", duration))
}

// CreateNotificationChannelHandler handles POST /v1/admin/risk/notification-channels requests
func (h *EnhancedRiskHandler) CreateNotificationChannelHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	h.logger.Info("Create notification channel request received",
		zap.String("request_id", requestID))

	// Check if alert system is available
	if h.alertSystem == nil {
		h.logger.Error("Alert system not available",
			zap.String("request_id", requestID))
		http.Error(w, "Alert system not available", http.StatusServiceUnavailable)
		return
	}

	// Parse request body
	var req struct {
		Name    string                 `json:"name"`
		Type    string                 `json:"type"` // email, sms, slack, webhook, teams, discord, pagerduty
		Enabled bool                   `json:"enabled"`
		Config  map[string]interface{} `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "Channel name is required", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		http.Error(w, "Channel type is required", http.StatusBadRequest)
		return
	}

	// Create channel based on type
	channel := h.createNotificationChannelFromRequest(req.Type, req.Name, req.Enabled, req.Config)
	if channel == nil {
		http.Error(w, fmt.Sprintf("Unsupported channel type: %s", req.Type), http.StatusBadRequest)
		return
	}

	// Get notification service and add channel
	notificationService := h.alertSystem.GetNotificationService()
	if notificationService == nil {
		h.logger.Error("Notification service not available",
			zap.String("request_id", requestID))
		http.Error(w, "Notification service not available", http.StatusServiceUnavailable)
		return
	}

	notificationService.AddChannel(req.Name, channel)

	// Return created channel info
	response := map[string]interface{}{
		"id":        req.Name,
		"type":      req.Type,
		"enabled":   req.Enabled,
		"config":    req.Config,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Create notification channel request completed",
		zap.String("request_id", requestID),
		zap.String("channel_name", req.Name),
		zap.String("channel_type", req.Type),
		zap.Duration("duration", duration))
}

// UpdateNotificationChannelHandler handles PUT /v1/admin/risk/notification-channels/{channel_id} requests
func (h *EnhancedRiskHandler) UpdateNotificationChannelHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	channelID := extractIDFromPath(r.URL.Path, "/v1/admin/risk/notification-channels/")
	if channelID == "" {
		http.Error(w, "Channel ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Update notification channel request received",
		zap.String("request_id", requestID),
		zap.String("channel_id", channelID))

	// Check if alert system is available
	if h.alertSystem == nil {
		h.logger.Error("Alert system not available",
			zap.String("request_id", requestID))
		http.Error(w, "Alert system not available", http.StatusServiceUnavailable)
		return
	}

	// Parse request body
	var req struct {
		Enabled *bool                  `json:"enabled,omitempty"`
		Config  map[string]interface{} `json:"config,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get notification service
	notificationService := h.alertSystem.GetNotificationService()
	if notificationService == nil {
		h.logger.Error("Notification service not available",
			zap.String("request_id", requestID))
		http.Error(w, "Notification service not available", http.StatusServiceUnavailable)
		return
	}

	// Get existing channel
	existingChannel, err := notificationService.GetChannel(channelID)
	if err != nil {
		h.logger.Error("Channel not found",
			zap.String("request_id", requestID),
			zap.String("channel_id", channelID),
			zap.Error(err))
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	// Update channel: Remove old and add new with updated config
	// Since channels are structs with private fields, we need to recreate them
	channelType := h.getChannelType(existingChannel)
	enabled := existingChannel.IsEnabled()
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	config := existingChannel.GetConfig()
	if req.Config != nil {
		// Merge configs
		for k, v := range req.Config {
			config[k] = v
		}
	}

	// Recreate channel with updated values
	updatedChannel := h.createNotificationChannelFromRequest(channelType, channelID, enabled, config)
	if updatedChannel == nil {
		http.Error(w, fmt.Sprintf("Failed to recreate channel type: %s", channelType), http.StatusInternalServerError)
		return
	}

	// Replace the channel
	notificationService.RemoveChannel(channelID)
	notificationService.AddChannel(channelID, updatedChannel)

	// Return updated channel info
	response := map[string]interface{}{
		"id":        channelID,
		"type":      channelType,
		"enabled":   enabled,
		"config":    config,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Update notification channel request completed",
		zap.String("request_id", requestID),
		zap.String("channel_id", channelID),
		zap.Duration("duration", duration))
}

// DeleteNotificationChannelHandler handles DELETE /v1/admin/risk/notification-channels/{channel_id} requests
func (h *EnhancedRiskHandler) DeleteNotificationChannelHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	channelID := extractIDFromPath(r.URL.Path, "/v1/admin/risk/notification-channels/")
	if channelID == "" {
		http.Error(w, "Channel ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Delete notification channel request received",
		zap.String("request_id", requestID),
		zap.String("channel_id", channelID))

	// Check if alert system is available
	if h.alertSystem == nil {
		h.logger.Error("Alert system not available",
			zap.String("request_id", requestID))
		http.Error(w, "Alert system not available", http.StatusServiceUnavailable)
		return
	}

	// Get notification service
	notificationService := h.alertSystem.GetNotificationService()
	if notificationService == nil {
		h.logger.Error("Notification service not available",
			zap.String("request_id", requestID))
		http.Error(w, "Notification service not available", http.StatusServiceUnavailable)
		return
	}

	// Check if channel exists
	_, err := notificationService.GetChannel(channelID)
	if err != nil {
		h.logger.Error("Channel not found",
			zap.String("request_id", requestID),
			zap.String("channel_id", channelID),
			zap.Error(err))
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	// Remove the channel
	notificationService.RemoveChannel(channelID)

	// Return success response
	response := map[string]interface{}{
		"id":        channelID,
		"deleted":   true,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Delete notification channel request completed",
		zap.String("request_id", requestID),
		zap.String("channel_id", channelID),
		zap.Duration("duration", duration))
}

// GetSystemHealthHandler handles GET /v1/admin/risk/system/health requests
func (h *EnhancedRiskHandler) GetSystemHealthHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	h.logger.Info("Get system health request received",
		zap.String("request_id", requestID))

	// TODO: Implement actual health checks
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"services": map[string]interface{}{
			"risk_detection":        "operational",
			"recommendation_engine": "operational",
			"trend_analysis":        "operational",
			"alert_system":          "operational",
		},
		"version": "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(health); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get system health request completed",
		zap.String("request_id", requestID),
		zap.Duration("duration", duration))
}

// GetSystemMetricsHandler handles GET /v1/admin/risk/system/metrics requests
func (h *EnhancedRiskHandler) GetSystemMetricsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	h.logger.Info("Get system metrics request received",
		zap.String("request_id", requestID))

	// TODO: Implement actual metrics collection
	metrics := map[string]interface{}{
		"timestamp": time.Now(),
		"assessments": map[string]interface{}{
			"total":     0,
			"completed": 0,
			"pending":   0,
			"failed":    0,
		},
		"alerts": map[string]interface{}{
			"active":       0,
			"acknowledged": 0,
			"resolved":     0,
		},
		"performance": map[string]interface{}{
			"avg_processing_time_ms": 0,
			"p95_processing_time_ms": 0,
			"p99_processing_time_ms": 0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Get system metrics request completed",
		zap.String("request_id", requestID),
		zap.Duration("duration", duration))
}

// CleanupSystemDataHandler handles POST /v1/admin/risk/system/cleanup requests
func (h *EnhancedRiskHandler) CleanupSystemDataHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	h.logger.Info("Cleanup system data request received",
		zap.String("request_id", requestID))

	// Parse request body for cleanup parameters
	var req struct {
		OlderThanDays int      `json:"older_than_days"`
		DataTypes     []string `json:"data_types"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Use defaults if body is empty
		req.OlderThanDays = 90
		req.DataTypes = []string{"assessments", "alerts", "trends"}
	}

	// Validate older_than_days
	if req.OlderThanDays <= 0 {
		req.OlderThanDays = 90
	}

	// If no data types specified, clean all
	if len(req.DataTypes) == 0 {
		req.DataTypes = []string{"assessments", "alerts", "trends"}
	}

	ctx := r.Context()
	cleaned := make(map[string]int)
	var cleanupErrors []string

	// Cleanup alerts if requested
	if contains(req.DataTypes, "alerts") {
		if h.alertSystem != nil {
			if err := h.alertSystem.CleanupOldAlerts(ctx); err != nil {
				h.logger.Error("Failed to cleanup alerts",
					zap.String("request_id", requestID),
					zap.Error(err))
				cleanupErrors = append(cleanupErrors, fmt.Sprintf("alerts: %v", err))
			} else {
				// Note: CleanupOldAlerts doesn't return count, so we'll use 0 for now
				cleaned["alerts"] = 0
				h.logger.Info("Cleaned up old alerts",
					zap.String("request_id", requestID),
					zap.Int("older_than_days", req.OlderThanDays))
			}
		} else {
			cleaned["alerts"] = 0
		}
	}

	// Cleanup trend data if requested
	if contains(req.DataTypes, "trends") {
		if h.trendAnalysisService != nil {
			if err := h.trendAnalysisService.CleanupOldData(ctx); err != nil {
				h.logger.Error("Failed to cleanup trend data",
					zap.String("request_id", requestID),
					zap.Error(err))
				cleanupErrors = append(cleanupErrors, fmt.Sprintf("trends: %v", err))
			} else {
				// Note: CleanupOldData doesn't return count, so we'll use 0 for now
				cleaned["trends"] = 0
				h.logger.Info("Cleaned up old trend data",
					zap.String("request_id", requestID),
					zap.Int("older_than_days", req.OlderThanDays))
			}
		} else {
			cleaned["trends"] = 0
		}
	}

	// Assessments cleanup would require a repository/service that we don't have direct access to
	if contains(req.DataTypes, "assessments") {
		cleaned["assessments"] = 0
		h.logger.Info("Assessment cleanup not yet implemented",
			zap.String("request_id", requestID))
	}

	// Build response
	result := map[string]interface{}{
		"cleaned":        cleaned,
		"older_than_days": req.OlderThanDays,
		"data_types":      req.DataTypes,
		"timestamp":       time.Now(),
	}

	if len(cleanupErrors) > 0 {
		result["errors"] = cleanupErrors
		result["message"] = "Cleanup completed with some errors"
	} else {
		result["message"] = "Cleanup completed successfully"
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Cleanup system data request completed",
		zap.String("request_id", requestID),
		zap.Int("older_than_days", req.OlderThanDays),
		zap.Strings("data_types", req.DataTypes),
		zap.Duration("duration", duration))
}

// ExportThresholdsHandler handles GET /v1/admin/risk/threshold-export requests
func (h *EnhancedRiskHandler) ExportThresholdsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	// Debug: Log that handler was called
	h.logger.Info("ExportThresholdsHandler CALLED",
		zap.String("request_id", requestID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("raw_path", r.URL.RawPath))

	h.logger.Info("Export thresholds request received",
		zap.String("request_id", requestID))

	// Check if ThresholdManager is available
	if h.thresholdManager == nil {
		h.logger.Error("ThresholdManager is not initialized",
			zap.String("request_id", requestID))
		http.Error(w, "Threshold management service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Create ThresholdConfigService
	thresholdService := risk.NewThresholdConfigService(h.thresholdManager)

	// Export thresholds
	exportData, err := thresholdService.ExportThresholds()
	if err != nil {
		h.logger.Error("Failed to export thresholds",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to export thresholds: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set response headers for file download
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=thresholds_export.json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)

	// Write JSON data directly
	if _, err := w.Write(exportData); err != nil {
		h.logger.Error("Failed to write export data",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Export thresholds request completed",
		zap.String("request_id", requestID),
		zap.Int("bytes_exported", len(exportData)),
		zap.Duration("duration", duration))
}

// ImportThresholdsHandler handles POST /v1/admin/risk/threshold-import requests
func (h *EnhancedRiskHandler) ImportThresholdsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := getRequestID(r)

	h.logger.Info("Import thresholds request received",
		zap.String("request_id", requestID))

	// Check if ThresholdManager is available
	if h.thresholdManager == nil {
		h.logger.Error("ThresholdManager is not initialized",
			zap.String("request_id", requestID))
		http.Error(w, "Threshold management service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Read request body
	importData, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Failed to read import data",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, "Failed to read import data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(importData) == 0 {
		http.Error(w, "Import data is required", http.StatusBadRequest)
		return
	}

	// Create ThresholdConfigService
	thresholdService := risk.NewThresholdConfigService(h.thresholdManager)

	// Import thresholds
	if err := thresholdService.ImportThresholds(importData); err != nil {
		h.logger.Error("Failed to import thresholds",
			zap.String("request_id", requestID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to import thresholds: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Get count of imported thresholds
	configs := h.thresholdManager.ListConfigs()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"message":        "Thresholds imported successfully",
		"imported_count": len(configs),
		"timestamp":      time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Import thresholds request completed",
		zap.String("request_id", requestID),
		zap.Int("imported_count", len(configs)),
		zap.Duration("duration", duration))
}

// Helper functions

// getRequestID safely extracts request ID from context
func getRequestID(r *http.Request) string {
	if id := r.Context().Value("request_id"); id != nil {
		if str, ok := id.(string); ok {
			return str
		}
	}
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// extractIDFromPath extracts an ID from a URL path after a given prefix
// createNotificationChannelFromRequest creates a notification channel based on type
func (h *EnhancedRiskHandler) createNotificationChannelFromRequest(channelType, name string, enabled bool, config map[string]interface{}) risk.NotificationChannel {
	if config == nil {
		config = make(map[string]interface{})
	}

	switch strings.ToLower(channelType) {
	case "email":
		return risk.NewEmailNotificationChannel(name, enabled, config)
	case "sms":
		return risk.NewSMSNotificationChannel(name, enabled, config)
	case "slack":
		return risk.NewSlackNotificationChannel(name, enabled, config)
	case "webhook":
		return risk.NewWebhookNotificationChannel(name, enabled, config)
	case "teams":
		return risk.NewTeamsNotificationChannel(name, enabled, config)
	case "discord":
		return risk.NewDiscordNotificationChannel(name, enabled, config)
	case "pagerduty":
		return risk.NewPagerDutyNotificationChannel(name, enabled, config)
	case "dashboard":
		return risk.NewDashboardNotificationChannel(name, enabled, config)
	default:
		return nil
	}
}

// getChannelType identifies the type of a notification channel
func (h *EnhancedRiskHandler) getChannelType(channel risk.NotificationChannel) string {
	switch channel.(type) {
	case *risk.EmailNotificationChannel:
		return "email"
	case *risk.SMSNotificationChannel:
		return "sms"
	case *risk.SlackNotificationChannel:
		return "slack"
	case *risk.WebhookNotificationChannel:
		return "webhook"
	case *risk.TeamsNotificationChannel:
		return "teams"
	case *risk.DiscordNotificationChannel:
		return "discord"
	case *risk.PagerDutyNotificationChannel:
		return "pagerduty"
	case *risk.DashboardNotificationChannel:
		return "dashboard"
	default:
		return "unknown"
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func extractIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	id := strings.TrimPrefix(path, prefix)
	// Remove any trailing path segments
	if idx := strings.Index(id, "/"); idx != -1 {
		id = id[:idx]
	}
	return id
}
