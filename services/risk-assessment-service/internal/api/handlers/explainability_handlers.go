package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/ml/explainability"
	"kyb-platform/services/risk-assessment-service/internal/ml/service"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// ExplainabilityHandlers handles explainability-related API requests
type ExplainabilityHandlers struct {
	mlService  *service.MLService
	explainer  *explainability.SHAPExplainer
	analyzer   *explainability.FeatureContributionAnalyzer
	visualizer *explainability.ExplainabilityVisualizer
	logger     *zap.Logger
}

// NewExplainabilityHandlers creates a new ExplainabilityHandlers
func NewExplainabilityHandlers(mlService *service.MLService, logger *zap.Logger) *ExplainabilityHandlers {
	// Initialize explainability components
	featureNames := []string{
		"industry_code", "country_code", "annual_revenue", "employee_count",
		"years_in_business", "has_website", "has_email", "has_phone", "name_length",
		"address_completeness", "phone_validity", "email_validity", "website_validity",
		"business_age_score", "revenue_stability", "employee_growth", "geographic_risk",
		"industry_risk", "compliance_score", "media_sentiment", "financial_health",
		"operational_efficiency", "market_position", "technology_adoption", "regulatory_compliance",
		"supply_chain_risk", "cybersecurity_score", "environmental_factors", "social_governance",
		"economic_indicators", "seasonal_patterns", "trend_analysis", "volatility_metrics",
		"correlation_factors", "anomaly_detection",
	}

	explainer := explainability.NewSHAPExplainer(featureNames, logger)
	analyzer := explainability.NewFeatureContributionAnalyzer(logger)
	visualizer := explainability.NewExplainabilityVisualizer(logger)

	return &ExplainabilityHandlers{
		mlService:  mlService,
		explainer:  explainer,
		analyzer:   analyzer,
		visualizer: visualizer,
		logger:     logger,
	}
}

// ExplainPredictionRequest represents the request to explain a prediction
type ExplainPredictionRequest struct {
	BusinessName    string    `json:"business_name" validate:"required"`
	BusinessAddress string    `json:"business_address"`
	Industry        string    `json:"industry"`
	Country         string    `json:"country"`
	Features        []float64 `json:"features"`
	Prediction      float64   `json:"prediction"`
}

// ExplainPredictionResponse represents the response for prediction explanation
type ExplainPredictionResponse struct {
	Explanation        *explainability.SHAPExplanation      `json:"explanation"`
	Analysis           *explainability.ContributionAnalysis `json:"analysis"`
	Visualization      *explainability.VisualizationData    `json:"visualization,omitempty"`
	TextualExplanation string                               `json:"textual_explanation,omitempty"`
	Timestamp          time.Time                            `json:"timestamp"`
}

// HandleExplainPrediction handles requests to explain a prediction
func (eh *ExplainabilityHandlers) HandleExplainPrediction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	eh.logger.Info("Handling explain prediction request")

	var req ExplainPredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		eh.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BusinessName == "" {
		http.Error(w, "Business name is required", http.StatusBadRequest)
		return
	}

	if len(req.Features) == 0 {
		http.Error(w, "Features are required", http.StatusBadRequest)
		return
	}

	// Create business request
	business := &models.RiskAssessmentRequest{
		BusinessName:    req.BusinessName,
		BusinessAddress: req.BusinessAddress,
		Industry:        req.Industry,
		Country:         req.Country,
	}

	// Generate feature importance map (mock implementation)
	featureImportance := make(map[string]float64)
	featureNames := []string{
		"industry_code", "country_code", "annual_revenue", "employee_count",
		"years_in_business", "has_website", "has_email", "has_phone", "name_length",
		"address_completeness", "phone_validity", "email_validity", "website_validity",
		"business_age_score", "revenue_stability", "employee_growth", "geographic_risk",
		"industry_risk", "compliance_score", "media_sentiment", "financial_health",
		"operational_efficiency", "market_position", "technology_adoption", "regulatory_compliance",
		"supply_chain_risk", "cybersecurity_score", "environmental_factors", "social_governance",
		"economic_indicators", "seasonal_patterns", "trend_analysis", "volatility_metrics",
		"correlation_factors", "anomaly_detection",
	}

	// Generate mock feature importance
	for i, name := range featureNames {
		if i < len(req.Features) {
			featureImportance[name] = req.Features[i] * 0.1 // Mock importance
		}
	}

	// Explain prediction
	explanation, err := eh.explainer.ExplainPrediction(ctx, business, req.Features, req.Prediction, featureImportance)
	if err != nil {
		eh.logger.Error("Failed to explain prediction", zap.Error(err))
		http.Error(w, "Failed to explain prediction", http.StatusInternalServerError)
		return
	}

	// Analyze contributions
	analysis, err := eh.analyzer.AnalyzeContributions(ctx, business, req.Features, req.Prediction, featureNames)
	if err != nil {
		eh.logger.Error("Failed to analyze contributions", zap.Error(err))
		http.Error(w, "Failed to analyze contributions", http.StatusInternalServerError)
		return
	}

	// Generate visualization
	visualization, err := eh.visualizer.GenerateWaterfallChart(ctx, explanation)
	if err != nil {
		eh.logger.Error("Failed to generate visualization", zap.Error(err))
		// Continue without visualization
	}

	// Generate textual explanation
	textualExplanation, err := eh.visualizer.GenerateTextualExplanation(ctx, explanation, analysis)
	if err != nil {
		eh.logger.Error("Failed to generate textual explanation", zap.Error(err))
		// Continue without textual explanation
	}

	response := ExplainPredictionResponse{
		Explanation:        explanation,
		Analysis:           analysis,
		Visualization:      visualization,
		TextualExplanation: textualExplanation,
		Timestamp:          time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		eh.logger.Error("Failed to encode response", zap.Error(err))
	}

	eh.logger.Info("Explain prediction request completed",
		zap.Duration("duration", time.Since(startTime)),
		zap.String("business_name", req.BusinessName))
}

// ComparePredictionsRequest represents the request to compare predictions
type ComparePredictionsRequest struct {
	Business1 ExplainPredictionRequest `json:"business_1" validate:"required"`
	Business2 ExplainPredictionRequest `json:"business_2" validate:"required"`
}

// ComparePredictionsResponse represents the response for prediction comparison
type ComparePredictionsResponse struct {
	Comparison    *explainability.ContributionComparison `json:"comparison"`
	Visualization *explainability.VisualizationData      `json:"visualization,omitempty"`
	Timestamp     time.Time                              `json:"timestamp"`
}

// HandleComparePredictions handles requests to compare predictions
func (eh *ExplainabilityHandlers) HandleComparePredictions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	eh.logger.Info("Handling compare predictions request")

	var req ComparePredictionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		eh.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Business1.BusinessName == "" || req.Business2.BusinessName == "" {
		http.Error(w, "Both business names are required", http.StatusBadRequest)
		return
	}

	if len(req.Business1.Features) == 0 || len(req.Business2.Features) == 0 {
		http.Error(w, "Features for both businesses are required", http.StatusBadRequest)
		return
	}

	// Create business requests
	business1 := &models.RiskAssessmentRequest{
		BusinessName:    req.Business1.BusinessName,
		BusinessAddress: req.Business1.BusinessAddress,
		Industry:        req.Business1.Industry,
		Country:         req.Business1.Country,
	}

	business2 := &models.RiskAssessmentRequest{
		BusinessName:    req.Business2.BusinessName,
		BusinessAddress: req.Business2.BusinessAddress,
		Industry:        req.Business2.Industry,
		Country:         req.Business2.Country,
	}

	featureNames := []string{
		"industry_code", "country_code", "annual_revenue", "employee_count",
		"years_in_business", "has_website", "has_email", "has_phone", "name_length",
		"address_completeness", "phone_validity", "email_validity", "website_validity",
		"business_age_score", "revenue_stability", "employee_growth", "geographic_risk",
		"industry_risk", "compliance_score", "media_sentiment", "financial_health",
		"operational_efficiency", "market_position", "technology_adoption", "regulatory_compliance",
		"supply_chain_risk", "cybersecurity_score", "environmental_factors", "social_governance",
		"economic_indicators", "seasonal_patterns", "trend_analysis", "volatility_metrics",
		"correlation_factors", "anomaly_detection",
	}

	// Compare contributions
	comparison, err := eh.analyzer.CompareContributions(
		ctx,
		business1, business2,
		req.Business1.Features, req.Business2.Features,
		req.Business1.Prediction, req.Business2.Prediction,
		featureNames,
	)
	if err != nil {
		eh.logger.Error("Failed to compare contributions", zap.Error(err))
		http.Error(w, "Failed to compare contributions", http.StatusInternalServerError)
		return
	}

	// Generate visualization
	visualization, err := eh.visualizer.GenerateContributionComparison(ctx, comparison)
	if err != nil {
		eh.logger.Error("Failed to generate comparison visualization", zap.Error(err))
		// Continue without visualization
	}

	response := ComparePredictionsResponse{
		Comparison:    comparison,
		Visualization: visualization,
		Timestamp:     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		eh.logger.Error("Failed to encode response", zap.Error(err))
	}

	eh.logger.Info("Compare predictions request completed",
		zap.Duration("duration", time.Since(startTime)),
		zap.String("business_1", req.Business1.BusinessName),
		zap.String("business_2", req.Business2.BusinessName))
}

// ExplainRiskFactorsRequest represents the request to explain risk factors
type ExplainRiskFactorsRequest struct {
	BusinessName    string `json:"business_name" validate:"required"`
	BusinessAddress string `json:"business_address"`
	Industry        string `json:"industry"`
	Country         string `json:"country"`
}

// ExplainRiskFactorsResponse represents the response for risk factor explanations
type ExplainRiskFactorsResponse struct {
	Explanations  []explainability.RiskFactorExplanation `json:"explanations"`
	Visualization *explainability.VisualizationData      `json:"visualization,omitempty"`
	Timestamp     time.Time                              `json:"timestamp"`
}

// HandleExplainRiskFactors handles requests to explain risk factors
func (eh *ExplainabilityHandlers) HandleExplainRiskFactors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	eh.logger.Info("Handling explain risk factors request")

	var req ExplainRiskFactorsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		eh.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BusinessName == "" {
		http.Error(w, "Business name is required", http.StatusBadRequest)
		return
	}

	// Create business request
	business := &models.RiskAssessmentRequest{
		BusinessName:    req.BusinessName,
		BusinessAddress: req.BusinessAddress,
		Industry:        req.Industry,
		Country:         req.Country,
	}

	// Generate mock risk factors
	riskFactors := []models.RiskFactor{
		{
			Category:    models.RiskCategoryFinancial,
			Name:        "Revenue Risk",
			Score:       0.7,
			Weight:      0.3,
			Description: "High revenue volatility",
			Source:      "financial_analysis",
			Confidence:  0.8,
		},
		{
			Category:    models.RiskCategoryOperational,
			Name:        "Operational Risk",
			Score:       0.4,
			Weight:      0.2,
			Description: "Stable operations",
			Source:      "operational_analysis",
			Confidence:  0.9,
		},
		{
			Category:    models.RiskCategoryCompliance,
			Name:        "Compliance Risk",
			Score:       0.6,
			Weight:      0.25,
			Description: "Moderate compliance risk",
			Source:      "compliance_check",
			Confidence:  0.7,
		},
		{
			Category:    models.RiskCategoryReputational,
			Name:        "Reputational Risk",
			Score:       0.3,
			Weight:      0.15,
			Description: "Low reputational risk",
			Source:      "media_monitoring",
			Confidence:  0.85,
		},
		{
			Category:    models.RiskCategoryRegulatory,
			Name:        "Regulatory Risk",
			Score:       0.5,
			Weight:      0.1,
			Description: "Moderate regulatory risk",
			Source:      "regulatory_analysis",
			Confidence:  0.75,
		},
	}

	// Explain risk factors
	explanations, err := eh.explainer.ExplainRiskFactors(ctx, riskFactors, business)
	if err != nil {
		eh.logger.Error("Failed to explain risk factors", zap.Error(err))
		http.Error(w, "Failed to explain risk factors", http.StatusInternalServerError)
		return
	}

	// Generate visualization
	visualization, err := eh.visualizer.GenerateRiskFactorExplanation(ctx, explanations)
	if err != nil {
		eh.logger.Error("Failed to generate risk factor visualization", zap.Error(err))
		// Continue without visualization
	}

	response := ExplainRiskFactorsResponse{
		Explanations:  explanations,
		Visualization: visualization,
		Timestamp:     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		eh.logger.Error("Failed to encode response", zap.Error(err))
	}

	eh.logger.Info("Explain risk factors request completed",
		zap.Duration("duration", time.Since(startTime)),
		zap.String("business_name", req.BusinessName))
}

// GenerateVisualizationRequest represents the request to generate visualization
type GenerateVisualizationRequest struct {
	Type   string                 `json:"type" validate:"required"`
	Data   map[string]interface{} `json:"data" validate:"required"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// HandleGenerateVisualization handles requests to generate visualizations
func (eh *ExplainabilityHandlers) HandleGenerateVisualization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	eh.logger.Info("Handling generate visualization request")

	var req GenerateVisualizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		eh.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Type == "" {
		http.Error(w, "Visualization type is required", http.StatusBadRequest)
		return
	}

	// Generate visualization based on type
	var visualization *explainability.VisualizationData
	var err error

	switch req.Type {
	case "waterfall_chart":
		// This would typically come from a previous explanation request
		http.Error(w, "Waterfall chart requires explanation data", http.StatusBadRequest)
		return

	case "feature_importance":
		// This would typically come from a previous analysis request
		http.Error(w, "Feature importance chart requires analysis data", http.StatusBadRequest)
		return

	case "risk_category_breakdown":
		if breakdown, ok := req.Data["breakdown"].(map[string]float64); ok {
			visualization, err = eh.visualizer.GenerateRiskCategoryBreakdown(ctx, breakdown)
		} else {
			http.Error(w, "Invalid breakdown data", http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, "Unsupported visualization type", http.StatusBadRequest)
		return
	}

	if err != nil {
		eh.logger.Error("Failed to generate visualization", zap.Error(err))
		http.Error(w, "Failed to generate visualization", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(visualization); err != nil {
		eh.logger.Error("Failed to encode response", zap.Error(err))
	}

	eh.logger.Info("Generate visualization request completed",
		zap.Duration("duration", time.Since(startTime)),
		zap.String("type", req.Type))
}

// GetExplainabilityInfo handles requests to get explainability information
func (eh *ExplainabilityHandlers) HandleGetExplainabilityInfo(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	eh.logger.Info("Handling get explainability info request")

	info := map[string]interface{}{
		"service": "explainability",
		"version": "1.0.0",
		"capabilities": []string{
			"prediction_explanation",
			"feature_contribution_analysis",
			"prediction_comparison",
			"risk_factor_explanation",
			"visualization_generation",
		},
		"supported_visualizations": []string{
			"waterfall_chart",
			"feature_importance",
			"risk_category_breakdown",
			"contribution_comparison",
			"risk_factor_explanation",
		},
		"feature_count": 35,
		"timestamp":     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(info); err != nil {
		eh.logger.Error("Failed to encode response", zap.Error(err))
	}

	eh.logger.Info("Get explainability info request completed",
		zap.Duration("duration", time.Since(startTime)))
}
