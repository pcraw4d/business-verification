package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/middleware"
	"kyb-platform/services/risk-assessment-service/internal/ml/service"
	"kyb-platform/services/risk-assessment-service/internal/models"
	"kyb-platform/services/risk-assessment-service/internal/validation"
)

// AdvancedPredictionHandler handles advanced prediction requests
type AdvancedPredictionHandler struct {
	mlService    *service.MLService
	logger       *zap.Logger
	validator    *validation.Validator
	errorHandler *middleware.ErrorHandler
}

// NewAdvancedPredictionHandler creates a new advanced prediction handler
func NewAdvancedPredictionHandler(
	mlService *service.MLService,
	logger *zap.Logger,
) *AdvancedPredictionHandler {
	return &AdvancedPredictionHandler{
		mlService:    mlService,
		logger:       logger,
		validator:    validation.NewValidator(),
		errorHandler: middleware.NewErrorHandler(logger),
	}
}

// AdvancedPredictionRequest represents a request for advanced predictions
type AdvancedPredictionRequest struct {
	Business                *models.RiskAssessmentRequest `json:"business"`
	PredictionHorizons      []int                         `json:"prediction_horizons"`
	ModelPreference         string                        `json:"model_preference,omitempty"` // "auto", "xgboost", "lstm", "ensemble"
	IncludeTemporalAnalysis bool                          `json:"include_temporal_analysis,omitempty"`
	IncludeScenarioAnalysis bool                          `json:"include_scenario_analysis,omitempty"`
	IncludeModelComparison  bool                          `json:"include_model_comparison,omitempty"`
	ConfidenceThreshold     float64                       `json:"confidence_threshold,omitempty"`
	CustomScenarios         []string                      `json:"custom_scenarios,omitempty"`
	Metadata                map[string]interface{}        `json:"metadata,omitempty"`
}

// AdvancedPredictionResponse represents the response for advanced predictions
type AdvancedPredictionResponse struct {
	RequestID          string                    `json:"request_id"`
	BusinessID         string                    `json:"business_id"`
	Predictions        map[int]HorizonPrediction `json:"predictions"`
	ModelComparison    *ModelComparison          `json:"model_comparison,omitempty"`
	TemporalAnalysis   interface{}               `json:"temporal_analysis,omitempty"`
	ScenarioAnalysis   []ScenarioAnalysis        `json:"scenario_analysis,omitempty"`
	ConfidenceAnalysis *ConfidenceAnalysis       `json:"confidence_analysis,omitempty"`
	ProcessingTime     time.Duration             `json:"processing_time"`
	GeneratedAt        time.Time                 `json:"generated_at"`
	Metadata           map[string]interface{}    `json:"metadata,omitempty"`
}

// HorizonPrediction represents a prediction for a specific horizon
type HorizonPrediction struct {
	HorizonMonths    int                       `json:"horizon_months"`
	ModelType        string                    `json:"model_type"`
	PredictedScore   float64                   `json:"predicted_score"`
	PredictedLevel   models.RiskLevel          `json:"predicted_level"`
	ConfidenceScore  float64                   `json:"confidence_score"`
	RiskFactors      []models.RiskFactor       `json:"risk_factors"`
	ScenarioAnalysis []models.ScenarioAnalysis `json:"scenario_analysis"`
	PredictionDate   time.Time                 `json:"prediction_date"`
	ModelInfo        interface{}               `json:"model_info,omitempty"`
	Metadata         map[string]interface{}    `json:"metadata,omitempty"`
}

// ModelComparison compares predictions from different models
type ModelComparison struct {
	Horizons            map[int]ModelComparisonHorizon `json:"horizons"`
	BestModelPerHorizon map[int]string                 `json:"best_model_per_horizon"`
	AgreementAnalysis   *AgreementAnalysis             `json:"agreement_analysis,omitempty"`
	PerformanceMetrics  *PerformanceMetrics            `json:"performance_metrics,omitempty"`
}

// ModelComparisonHorizon compares models for a specific horizon
type ModelComparisonHorizon struct {
	HorizonMonths        int                `json:"horizon_months"`
	XGBoostPrediction    *HorizonPrediction `json:"xgboost_prediction,omitempty"`
	LSTMPrediction       *HorizonPrediction `json:"lstm_prediction,omitempty"`
	EnsemblePrediction   *HorizonPrediction `json:"ensemble_prediction,omitempty"`
	BestModel            string             `json:"best_model"`
	ConfidenceDifference float64            `json:"confidence_difference"`
	ScoreDifference      float64            `json:"score_difference"`
}

// AgreementAnalysis analyzes agreement between models
type AgreementAnalysis struct {
	OverallAgreement         float64         `json:"overall_agreement"`
	AgreementByHorizon       map[int]float64 `json:"agreement_by_horizon"`
	DisagreementThreshold    float64         `json:"disagreement_threshold"`
	HighDisagreementHorizons []int           `json:"high_disagreement_horizons"`
}

// PerformanceMetrics provides performance information
type PerformanceMetrics struct {
	AverageLatency time.Duration            `json:"average_latency"`
	LatencyByModel map[string]time.Duration `json:"latency_by_model"`
	MemoryUsage    map[string]int64         `json:"memory_usage"`
	Throughput     float64                  `json:"throughput"`
}

// ScenarioAnalysis provides scenario-based analysis
type ScenarioAnalysis struct {
	ScenarioName         string           `json:"scenario_name"`
	Description          string           `json:"description"`
	Probability          float64          `json:"probability"`
	RiskScore            float64          `json:"risk_score"`
	RiskLevel            models.RiskLevel `json:"risk_level"`
	Impact               string           `json:"impact"`
	TimeHorizon          int              `json:"time_horizon"`
	KeyFactors           []string         `json:"key_factors"`
	MitigationStrategies []string         `json:"mitigation_strategies,omitempty"`
}

// ConfidenceAnalysis provides confidence analysis
type ConfidenceAnalysis struct {
	OverallConfidence         float64            `json:"overall_confidence"`
	ConfidenceByHorizon       map[int]float64    `json:"confidence_by_horizon"`
	ConfidenceByModel         map[string]float64 `json:"confidence_by_model"`
	LowConfidencePredictions  []int              `json:"low_confidence_predictions"`
	HighConfidencePredictions []int              `json:"high_confidence_predictions"`
	CalibrationScore          float64            `json:"calibration_score"`
}

// HandleAdvancedPrediction handles POST /api/v1/risk/predict-advanced
func (h *AdvancedPredictionHandler) HandleAdvancedPrediction(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	h.logger.Info("Processing advanced prediction request")

	// Parse request
	var req AdvancedPredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorHandler.HandleError(w, r, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Validate request
	if err := h.validateAdvancedPredictionRequest(&req); err != nil {
		h.errorHandler.HandleError(w, r, fmt.Errorf("validation failed: %w", err))
		return
	}

	// Sanitize inputs
	h.sanitizeAdvancedPredictionRequest(&req)

	// Generate request ID
	requestID := h.generateRequestID()

	// Process predictions for each horizon
	predictions := make(map[int]HorizonPrediction)
	var modelComparison *ModelComparison
	var temporalAnalysis interface{}
	var scenarioAnalysis []ScenarioAnalysis
	var confidenceAnalysis *ConfidenceAnalysis

	// Get predictions for each horizon
	for _, horizon := range req.PredictionHorizons {
		prediction, err := h.getHorizonPrediction(r.Context(), req.Business, horizon, req.ModelPreference)
		if err != nil {
			h.logger.Error("Failed to get prediction for horizon",
				zap.Int("horizon", horizon),
				zap.Error(err))
			continue
		}

		predictions[horizon] = *prediction
	}

	// Generate model comparison if requested
	if req.IncludeModelComparison && len(req.PredictionHorizons) > 1 {
		modelComparison = h.generateModelComparison(r.Context(), req.Business, req.PredictionHorizons)
	}

	// Generate temporal analysis if requested
	if req.IncludeTemporalAnalysis {
		temporalAnalysis = h.generateTemporalAnalysis(r.Context(), req.Business, req.PredictionHorizons)
	}

	// Generate scenario analysis if requested
	if req.IncludeScenarioAnalysis {
		scenarioAnalysis = h.generateScenarioAnalysis(r.Context(), req.Business, req.PredictionHorizons, req.CustomScenarios)
	}

	// Generate confidence analysis
	confidenceAnalysis = h.generateConfidenceAnalysis(predictions)

	// Create response
	response := &AdvancedPredictionResponse{
		RequestID:          requestID,
		BusinessID:         req.Business.BusinessName, // TODO: Use actual business ID
		Predictions:        predictions,
		ModelComparison:    modelComparison,
		TemporalAnalysis:   temporalAnalysis,
		ScenarioAnalysis:   scenarioAnalysis,
		ConfidenceAnalysis: confidenceAnalysis,
		ProcessingTime:     time.Since(start),
		GeneratedAt:        time.Now(),
		Metadata:           req.Metadata,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Advanced prediction completed",
		zap.String("request_id", requestID),
		zap.Int("horizons", len(req.PredictionHorizons)),
		zap.Duration("processing_time", time.Since(start)))
}

// validateAdvancedPredictionRequest validates the advanced prediction request
func (h *AdvancedPredictionHandler) validateAdvancedPredictionRequest(req *AdvancedPredictionRequest) error {
	if req.Business == nil {
		return fmt.Errorf("business information is required")
	}

	if len(req.PredictionHorizons) == 0 {
		return fmt.Errorf("at least one prediction horizon is required")
	}

	if len(req.PredictionHorizons) > 5 {
		return fmt.Errorf("maximum of 5 prediction horizons allowed")
	}

	// Validate horizons
	for _, horizon := range req.PredictionHorizons {
		if horizon < 1 || horizon > 24 {
			return fmt.Errorf("prediction horizon must be between 1 and 24 months")
		}
	}

	// Validate model preference
	if req.ModelPreference != "" {
		validModels := []string{"auto", "xgboost", "lstm", "ensemble"}
		valid := false
		for _, model := range validModels {
			if req.ModelPreference == model {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid model preference: %s", req.ModelPreference)
		}
	}

	// Validate confidence threshold
	if req.ConfidenceThreshold < 0 || req.ConfidenceThreshold > 1 {
		return fmt.Errorf("confidence threshold must be between 0 and 1")
	}

	return nil
}

// sanitizeAdvancedPredictionRequest sanitizes the request inputs
func (h *AdvancedPredictionHandler) sanitizeAdvancedPredictionRequest(req *AdvancedPredictionRequest) {
	if req.Business != nil {
		req.Business.BusinessName = h.validator.SanitizeInput(req.Business.BusinessName)
		req.Business.BusinessAddress = h.validator.SanitizeInput(req.Business.BusinessAddress)
		req.Business.Industry = h.validator.SanitizeInput(req.Business.Industry)
		req.Business.Country = h.validator.SanitizeInput(req.Business.Country)
		if req.Business.Phone != "" {
			req.Business.Phone = h.validator.SanitizeInput(req.Business.Phone)
		}
		if req.Business.Email != "" {
			req.Business.Email = h.validator.SanitizeInput(req.Business.Email)
		}
		if req.Business.Website != "" {
			req.Business.Website = h.validator.SanitizeInput(req.Business.Website)
		}
	}

	// Sanitize custom scenarios
	for i, scenario := range req.CustomScenarios {
		req.CustomScenarios[i] = h.validator.SanitizeInput(scenario)
	}
}

// getHorizonPrediction gets a prediction for a specific horizon
func (h *AdvancedPredictionHandler) getHorizonPrediction(ctx context.Context, business *models.RiskAssessmentRequest, horizon int, modelPreference string) (*HorizonPrediction, error) {
	// Set the horizon in the business request
	business.PredictionHorizon = horizon

	// Get prediction based on model preference
	var prediction *models.RiskPrediction
	var err error

	if modelPreference == "" || modelPreference == "auto" {
		prediction, err = h.mlService.PredictFutureRisk(ctx, "auto", business, horizon)
	} else {
		prediction, err = h.mlService.PredictFutureRisk(ctx, modelPreference, business, horizon)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get prediction: %w", err)
	}

	// Convert to horizon prediction
	horizonPrediction := &HorizonPrediction{
		HorizonMonths:    horizon,
		ModelType:        modelPreference,
		PredictedScore:   prediction.PredictedScore,
		PredictedLevel:   prediction.PredictedLevel,
		ConfidenceScore:  prediction.ConfidenceScore,
		RiskFactors:      prediction.RiskFactors,
		ScenarioAnalysis: prediction.ScenarioAnalysis,
		PredictionDate:   prediction.PredictionDate,
		ModelInfo:        map[string]interface{}{"model_type": modelPreference},
		Metadata:         map[string]interface{}{"model_type": modelPreference},
	}

	return horizonPrediction, nil
}

// generateModelComparison generates model comparison analysis
func (h *AdvancedPredictionHandler) generateModelComparison(ctx context.Context, business *models.RiskAssessmentRequest, horizons []int) *ModelComparison {
	comparison := &ModelComparison{
		Horizons:            make(map[int]ModelComparisonHorizon),
		BestModelPerHorizon: make(map[int]string),
	}

	// Get predictions from all models for each horizon
	for _, horizon := range horizons {
		horizonComparison := ModelComparisonHorizon{
			HorizonMonths: horizon,
		}

		// Get XGBoost prediction
		if xgbPrediction, err := h.mlService.PredictFutureRisk(ctx, "xgboost", business, horizon); err == nil {
			horizonComparison.XGBoostPrediction = &HorizonPrediction{
				HorizonMonths:    horizon,
				ModelType:        "xgboost",
				PredictedScore:   xgbPrediction.PredictedScore,
				PredictedLevel:   xgbPrediction.PredictedLevel,
				ConfidenceScore:  xgbPrediction.ConfidenceScore,
				RiskFactors:      xgbPrediction.RiskFactors,
				ScenarioAnalysis: xgbPrediction.ScenarioAnalysis,
				PredictionDate:   xgbPrediction.PredictionDate,
				Metadata:         map[string]interface{}{"model_type": "xgboost"},
			}
		}

		// Get LSTM prediction
		if lstmPrediction, err := h.mlService.PredictFutureRisk(ctx, "lstm", business, horizon); err == nil {
			horizonComparison.LSTMPrediction = &HorizonPrediction{
				HorizonMonths:    horizon,
				ModelType:        "lstm",
				PredictedScore:   lstmPrediction.PredictedScore,
				PredictedLevel:   lstmPrediction.PredictedLevel,
				ConfidenceScore:  lstmPrediction.ConfidenceScore,
				RiskFactors:      lstmPrediction.RiskFactors,
				ScenarioAnalysis: lstmPrediction.ScenarioAnalysis,
				PredictionDate:   lstmPrediction.PredictionDate,
				Metadata:         map[string]interface{}{"model_type": "lstm"},
			}
		}

		// Get Ensemble prediction
		if ensemblePrediction, err := h.mlService.PredictFutureRisk(ctx, "ensemble", business, horizon); err == nil {
			horizonComparison.EnsemblePrediction = &HorizonPrediction{
				HorizonMonths:    horizon,
				ModelType:        "ensemble",
				PredictedScore:   ensemblePrediction.PredictedScore,
				PredictedLevel:   ensemblePrediction.PredictedLevel,
				ConfidenceScore:  ensemblePrediction.ConfidenceScore,
				RiskFactors:      ensemblePrediction.RiskFactors,
				ScenarioAnalysis: ensemblePrediction.ScenarioAnalysis,
				PredictionDate:   ensemblePrediction.PredictionDate,
				Metadata:         map[string]interface{}{"model_type": "ensemble"},
			}
		}

		// Determine best model for this horizon
		horizonComparison.BestModel = h.determineBestModel(horizonComparison)

		// Calculate differences
		horizonComparison.ConfidenceDifference = h.calculateConfidenceDifference(horizonComparison)
		horizonComparison.ScoreDifference = h.calculateScoreDifference(horizonComparison)

		comparison.Horizons[horizon] = horizonComparison
		comparison.BestModelPerHorizon[horizon] = horizonComparison.BestModel
	}

	// Generate agreement analysis
	comparison.AgreementAnalysis = h.generateAgreementAnalysis(comparison)

	return comparison
}

// generateTemporalAnalysis generates temporal analysis
func (h *AdvancedPredictionHandler) generateTemporalAnalysis(ctx context.Context, business *models.RiskAssessmentRequest, horizons []int) interface{} {
	// This would integrate with the temporal feature builder
	// For now, return a mock analysis
	return map[string]interface{}{
		"trend_analysis": map[string]interface{}{
			"overall_trend":     "stable",
			"trend_strength":    0.3,
			"seasonal_patterns": []string{"q1_increase", "q4_decrease"},
		},
		"volatility_analysis": map[string]interface{}{
			"historical_volatility": 0.15,
			"volatility_trend":      "decreasing",
			"volatility_forecast":   0.12,
		},
		"time_series_features": map[string]interface{}{
			"sequence_length": 12,
			"feature_count":   25,
			"data_quality":    "high",
		},
	}
}

// generateScenarioAnalysis generates scenario analysis
func (h *AdvancedPredictionHandler) generateScenarioAnalysis(ctx context.Context, business *models.RiskAssessmentRequest, horizons []int, customScenarios []string) []ScenarioAnalysis {
	scenarios := []ScenarioAnalysis{
		{
			ScenarioName:         "Optimistic",
			Description:          "Best case scenario with favorable market conditions",
			Probability:          0.2,
			RiskScore:            0.2,
			RiskLevel:            models.RiskLevelLow,
			Impact:               "low",
			TimeHorizon:          horizons[0],
			KeyFactors:           []string{"market_growth", "regulatory_support", "strong_financials"},
			MitigationStrategies: []string{"maintain_current_strategy", "monitor_market_conditions"},
		},
		{
			ScenarioName:         "Base Case",
			Description:          "Most likely scenario based on current trends",
			Probability:          0.6,
			RiskScore:            0.5,
			RiskLevel:            models.RiskLevelMedium,
			Impact:               "medium",
			TimeHorizon:          horizons[0],
			KeyFactors:           []string{"stable_market", "normal_regulatory", "average_financials"},
			MitigationStrategies: []string{"regular_monitoring", "diversification"},
		},
		{
			ScenarioName:         "Pessimistic",
			Description:          "Worst case scenario with challenging conditions",
			Probability:          0.2,
			RiskScore:            0.8,
			RiskLevel:            models.RiskLevelHigh,
			Impact:               "high",
			TimeHorizon:          horizons[0],
			KeyFactors:           []string{"market_downturn", "regulatory_changes", "financial_stress"},
			MitigationStrategies: []string{"cost_reduction", "risk_hedging", "contingency_planning"},
		},
	}

	// Add custom scenarios if provided
	for _, customScenario := range customScenarios {
		scenarios = append(scenarios, ScenarioAnalysis{
			ScenarioName:         customScenario,
			Description:          fmt.Sprintf("Custom scenario: %s", customScenario),
			Probability:          0.1,
			RiskScore:            0.4,
			RiskLevel:            models.RiskLevelMedium,
			Impact:               "medium",
			TimeHorizon:          horizons[0],
			KeyFactors:           []string{"custom_factor"},
			MitigationStrategies: []string{"custom_mitigation"},
		})
	}

	return scenarios
}

// generateConfidenceAnalysis generates confidence analysis
func (h *AdvancedPredictionHandler) generateConfidenceAnalysis(predictions map[int]HorizonPrediction) *ConfidenceAnalysis {
	analysis := &ConfidenceAnalysis{
		ConfidenceByHorizon:       make(map[int]float64),
		ConfidenceByModel:         make(map[string]float64),
		LowConfidencePredictions:  []int{},
		HighConfidencePredictions: []int{},
	}

	var totalConfidence float64
	var count int

	// Analyze confidence by horizon and model
	for horizon, prediction := range predictions {
		analysis.ConfidenceByHorizon[horizon] = prediction.ConfidenceScore
		analysis.ConfidenceByModel[prediction.ModelType] = prediction.ConfidenceScore

		totalConfidence += prediction.ConfidenceScore
		count++

		// Categorize predictions by confidence
		if prediction.ConfidenceScore < 0.6 {
			analysis.LowConfidencePredictions = append(analysis.LowConfidencePredictions, horizon)
		} else if prediction.ConfidenceScore > 0.8 {
			analysis.HighConfidencePredictions = append(analysis.HighConfidencePredictions, horizon)
		}
	}

	// Calculate overall confidence
	if count > 0 {
		analysis.OverallConfidence = totalConfidence / float64(count)
	}

	// Calculate calibration score (simplified)
	analysis.CalibrationScore = 0.85 // Mock value

	return analysis
}

// Helper methods
func (h *AdvancedPredictionHandler) getModelTypeFromMetadata(metadata map[string]interface{}) string {
	if modelType, exists := metadata["model_type"]; exists {
		return modelType.(string)
	}
	return "unknown"
}

func (h *AdvancedPredictionHandler) determineBestModel(comparison ModelComparisonHorizon) string {
	// Simple logic to determine best model based on confidence
	bestModel := "ensemble"
	bestConfidence := 0.0

	if comparison.XGBoostPrediction != nil && comparison.XGBoostPrediction.ConfidenceScore > bestConfidence {
		bestConfidence = comparison.XGBoostPrediction.ConfidenceScore
		bestModel = "xgboost"
	}

	if comparison.LSTMPrediction != nil && comparison.LSTMPrediction.ConfidenceScore > bestConfidence {
		bestConfidence = comparison.LSTMPrediction.ConfidenceScore
		bestModel = "lstm"
	}

	if comparison.EnsemblePrediction != nil && comparison.EnsemblePrediction.ConfidenceScore > bestConfidence {
		bestConfidence = comparison.EnsemblePrediction.ConfidenceScore
		bestModel = "ensemble"
	}

	return bestModel
}

func (h *AdvancedPredictionHandler) calculateConfidenceDifference(comparison ModelComparisonHorizon) float64 {
	var confidences []float64

	if comparison.XGBoostPrediction != nil {
		confidences = append(confidences, comparison.XGBoostPrediction.ConfidenceScore)
	}
	if comparison.LSTMPrediction != nil {
		confidences = append(confidences, comparison.LSTMPrediction.ConfidenceScore)
	}
	if comparison.EnsemblePrediction != nil {
		confidences = append(confidences, comparison.EnsemblePrediction.ConfidenceScore)
	}

	if len(confidences) < 2 {
		return 0
	}

	// Calculate max difference
	maxDiff := 0.0
	for i := 0; i < len(confidences); i++ {
		for j := i + 1; j < len(confidences); j++ {
			diff := confidences[i] - confidences[j]
			if diff < 0 {
				diff = -diff
			}
			if diff > maxDiff {
				maxDiff = diff
			}
		}
	}

	return maxDiff
}

func (h *AdvancedPredictionHandler) calculateScoreDifference(comparison ModelComparisonHorizon) float64 {
	var scores []float64

	if comparison.XGBoostPrediction != nil {
		scores = append(scores, comparison.XGBoostPrediction.PredictedScore)
	}
	if comparison.LSTMPrediction != nil {
		scores = append(scores, comparison.LSTMPrediction.PredictedScore)
	}
	if comparison.EnsemblePrediction != nil {
		scores = append(scores, comparison.EnsemblePrediction.PredictedScore)
	}

	if len(scores) < 2 {
		return 0
	}

	// Calculate max difference
	maxDiff := 0.0
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			diff := scores[i] - scores[j]
			if diff < 0 {
				diff = -diff
			}
			if diff > maxDiff {
				maxDiff = diff
			}
		}
	}

	return maxDiff
}

func (h *AdvancedPredictionHandler) generateAgreementAnalysis(comparison *ModelComparison) *AgreementAnalysis {
	analysis := &AgreementAnalysis{
		AgreementByHorizon:       make(map[int]float64),
		DisagreementThreshold:    0.2,
		HighDisagreementHorizons: []int{},
	}

	var totalAgreement float64
	var count int

	for horizon, horizonComparison := range comparison.Horizons {
		agreement := 1.0 - horizonComparison.ScoreDifference
		analysis.AgreementByHorizon[horizon] = agreement

		totalAgreement += agreement
		count++

		if agreement < (1.0 - analysis.DisagreementThreshold) {
			analysis.HighDisagreementHorizons = append(analysis.HighDisagreementHorizons, horizon)
		}
	}

	if count > 0 {
		analysis.OverallAgreement = totalAgreement / float64(count)
	}

	return analysis
}

func (h *AdvancedPredictionHandler) generateRequestID() string {
	return fmt.Sprintf("adv_pred_%d", time.Now().UnixNano())
}

// HandleGetModelInfo handles GET /api/v1/models/info
func (h *AdvancedPredictionHandler) HandleGetModelInfo(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing model info request")

	// Get model type from query parameter
	modelType := r.URL.Query().Get("model")
	if modelType == "" {
		modelType = "all"
	}

	// Get model information
	response := make(map[string]interface{})

	if modelType == "all" || modelType == "xgboost" {
		if xgbInfo, err := h.mlService.GetModelInfo("xgboost"); err == nil {
			response["xgboost"] = xgbInfo
		}
	}

	if modelType == "all" || modelType == "lstm" {
		if lstmInfo, err := h.mlService.GetModelInfo("lstm"); err == nil {
			response["lstm"] = lstmInfo
		}
	}

	if modelType == "all" {
		response["ensemble"] = h.mlService.GetEnsembleInfo()
		response["available_models"] = h.mlService.ListModels()
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Model info request completed", zap.String("model_type", modelType))
}

// HandleGetModelPerformance handles GET /api/v1/models/performance
func (h *AdvancedPredictionHandler) HandleGetModelPerformance(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing model performance request")

	// Get performance metrics from the metrics collector
	_ = h.mlService.GetMetricsCollector() // TODO: Use actual metrics in future

	// Create performance response
	response := map[string]interface{}{
		"timestamp": time.Now(),
		"models": map[string]interface{}{
			"xgboost": map[string]interface{}{
				"status":          "active",
				"inference_count": 1000, // Mock data
				"average_latency": "50ms",
				"accuracy":        0.92,
				"last_updated":    time.Now().Add(-1 * time.Hour),
			},
			"lstm": map[string]interface{}{
				"status":          "active",
				"inference_count": 500, // Mock data
				"average_latency": "80ms",
				"accuracy":        0.88,
				"last_updated":    time.Now().Add(-30 * time.Minute),
			},
			"ensemble": map[string]interface{}{
				"status":          "active",
				"inference_count": 750, // Mock data
				"average_latency": "120ms",
				"accuracy":        0.90,
				"last_updated":    time.Now().Add(-15 * time.Minute),
			},
		},
		"system_metrics": map[string]interface{}{
			"total_requests":        2250,
			"success_rate":          0.99,
			"average_response_time": "85ms",
			"memory_usage":          "1.2GB",
			"cpu_usage":             "45%",
		},
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Model performance request completed")
}
