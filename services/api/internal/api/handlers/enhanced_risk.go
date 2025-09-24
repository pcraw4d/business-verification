package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"kyb-platform/internal/risk"
)

// EnhancedRiskHandler handles enhanced risk assessment API endpoints
type EnhancedRiskHandler struct {
	logger               *zap.Logger
	riskService          *risk.RiskAssessmentService
	enhancedCalculator   *risk.EnhancedRiskCalculator
	recommendationEngine *risk.RiskRecommendationEngine
	trendAnalysisService *risk.RiskTrendAnalysisService
	alertSystem          *risk.RiskAlertSystem
}

// NewEnhancedRiskHandler creates a new enhanced risk handler
func NewEnhancedRiskHandler(
	logger *zap.Logger,
	riskService *risk.RiskAssessmentService,
	enhancedCalculator *risk.EnhancedRiskCalculator,
	recommendationEngine *risk.RiskRecommendationEngine,
	trendAnalysisService *risk.RiskTrendAnalysisService,
	alertSystem *risk.RiskAlertSystem,
) *EnhancedRiskHandler {
	return &EnhancedRiskHandler{
		logger:               logger,
		riskService:          riskService,
		enhancedCalculator:   enhancedCalculator,
		recommendationEngine: recommendationEngine,
		trendAnalysisService: trendAnalysisService,
		alertSystem:          alertSystem,
	}
}

// EnhancedRiskAssessmentHandler handles POST /v1/risk/enhanced/assess requests
func (h *EnhancedRiskHandler) EnhancedRiskAssessmentHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

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
	requestID := r.Context().Value("request_id").(string)

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
	requestID := r.Context().Value("request_id").(string)

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
	requestID := r.Context().Value("request_id").(string)

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
	requestID := r.Context().Value("request_id").(string)

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
	requestID := r.Context().Value("request_id").(string)

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
	requestID := r.Context().Value("request_id").(string)

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
	requestID := r.Context().Value("request_id").(string)

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
	if request.BusinessName == "" {
		return fmt.Errorf("business_name is required")
	}
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
	// This would integrate all the enhanced services
	// For now, we'll create a basic response structure

	response := &risk.EnhancedRiskAssessmentResponse{
		BusinessID:          request.BusinessID,
		AssessmentTimestamp: time.Now(),
		OverallScore:        75.0, // Placeholder
		OverallLevel:        risk.RiskLevelHigh,
		Confidence:          0.85,
		ProcessingTime:      time.Since(time.Now()),
	}

	return response, nil
}
