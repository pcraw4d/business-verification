package models

import (
	"context"
	"fmt"
	"math"
	"time"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// XGBoostModel implements the RiskModel interface using XGBoost algorithm
type XGBoostModel struct {
	name             string
	version          string
	trained          bool
	featureExtractor *FeatureExtractor
	riskLevelEncoder *RiskLevelEncoder

	// Model parameters (simplified XGBoost implementation)
	nEstimators     int
	maxDepth        int
	learningRate    float64
	subsample       float64
	colsampleByTree float64

	// Model state (in a real implementation, this would be the actual XGBoost model)
	featureImportance map[string]float64
	treeStructures    []TreeStructure
	baseScore         float64
}

// TreeStructure represents a decision tree in the XGBoost ensemble
type TreeStructure struct {
	Nodes []TreeNode
}

// TreeNode represents a node in a decision tree
type TreeNode struct {
	FeatureIndex int
	Threshold    float64
	LeftChild    int
	RightChild   int
	LeafValue    float64
	IsLeaf       bool
}

// NewXGBoostModel creates a new XGBoost model
func NewXGBoostModel(name, version string) *XGBoostModel {
	return &XGBoostModel{
		name:             name,
		version:          version,
		trained:          false,
		featureExtractor: NewFeatureExtractor(),
		riskLevelEncoder: NewRiskLevelEncoder(),

		// Default XGBoost parameters
		nEstimators:     100,
		maxDepth:        6,
		learningRate:    0.1,
		subsample:       0.8,
		colsampleByTree: 0.8,

		featureImportance: make(map[string]float64),
		treeStructures:    make([]TreeStructure, 0),
		baseScore:         0.5,
	}
}

// Predict performs risk prediction for a given business
func (xgb *XGBoostModel) Predict(ctx context.Context, business *models.RiskAssessmentRequest) (*models.RiskAssessment, error) {
	if !xgb.trained {
		return nil, fmt.Errorf("model is not trained")
	}

	// Extract features
	features, err := xgb.featureExtractor.ExtractFeatures(business)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Make prediction
	prediction, err := xgb.predict(features)
	if err != nil {
		return nil, fmt.Errorf("prediction failed: %w", err)
	}

	// Convert prediction to risk assessment
	riskScore := prediction
	riskLevel := xgb.riskLevelEncoder.DecodeRiskLevel(riskScore)

	// Generate risk factors based on feature importance
	riskFactors := xgb.generateRiskFactors(features, riskScore)

	// Create risk assessment
	assessment := &models.RiskAssessment{
		ID:                fmt.Sprintf("risk_%d", time.Now().UnixNano()),
		BusinessName:      business.BusinessName,
		BusinessAddress:   business.BusinessAddress,
		Industry:          business.Industry,
		Country:           business.Country,
		RiskScore:         riskScore,
		RiskLevel:         riskLevel,
		RiskFactors:       riskFactors,
		PredictionHorizon: business.PredictionHorizon,
		ConfidenceScore:   xgb.calculateConfidence(features),
		Status:            models.StatusCompleted,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Metadata:          business.Metadata,
	}

	return assessment, nil
}

// PredictFuture performs future risk prediction for a given horizon
func (xgb *XGBoostModel) PredictFuture(ctx context.Context, business *models.RiskAssessmentRequest, horizonMonths int) (*models.RiskPrediction, error) {
	if !xgb.trained {
		return nil, fmt.Errorf("model is not trained")
	}

	// Extract features
	features, err := xgb.featureExtractor.ExtractFeatures(business)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Adjust features for future prediction (simplified time series adjustment)
	futureFeatures := xgb.adjustFeaturesForHorizon(features, horizonMonths)

	// Make prediction
	prediction, err := xgb.predict(futureFeatures)
	if err != nil {
		return nil, fmt.Errorf("prediction failed: %w", err)
	}

	// Convert prediction to risk prediction
	riskScore := prediction
	riskLevel := xgb.riskLevelEncoder.DecodeRiskLevel(riskScore)

	// Generate risk factors
	riskFactors := xgb.generateRiskFactors(futureFeatures, riskScore)

	// Generate scenario analysis
	scenarios := xgb.generateScenarioAnalysis(features, horizonMonths)

	// Create risk prediction
	riskPrediction := &models.RiskPrediction{
		BusinessID:       fmt.Sprintf("biz_%d", time.Now().UnixNano()),
		PredictionDate:   time.Now(),
		HorizonMonths:    horizonMonths,
		PredictedScore:   riskScore,
		PredictedLevel:   riskLevel,
		ConfidenceScore:  xgb.calculateConfidence(futureFeatures),
		RiskFactors:      riskFactors,
		ScenarioAnalysis: scenarios,
		CreatedAt:        time.Now(),
	}

	return riskPrediction, nil
}

// GetModelInfo returns information about the model
func (xgb *XGBoostModel) GetModelInfo() *ModelInfo {
	return &ModelInfo{
		Name:         xgb.name,
		Version:      xgb.version,
		Type:         "XGBoost",
		TrainingDate: time.Now(),
		Accuracy:     0.85, // Mock accuracy
		Precision:    0.82,
		Recall:       0.88,
		F1Score:      0.85,
		Features:     xgb.featureExtractor.GetFeatureNames(),
		Hyperparameters: map[string]interface{}{
			"n_estimators":     xgb.nEstimators,
			"max_depth":        xgb.maxDepth,
			"learning_rate":    xgb.learningRate,
			"subsample":        xgb.subsample,
			"colsample_bytree": xgb.colsampleByTree,
		},
		Metadata: map[string]interface{}{
			"trained":    xgb.trained,
			"base_score": xgb.baseScore,
		},
	}
}

// LoadModel loads the model from storage
func (xgb *XGBoostModel) LoadModel(ctx context.Context, modelPath string) error {
	// In a real implementation, this would load the actual XGBoost model
	// For now, we'll simulate loading by setting trained to true
	xgb.trained = true

	// Mock feature importance
	xgb.featureImportance = map[string]float64{
		"industry_code":     0.25,
		"country_code":      0.20,
		"annual_revenue":    0.15,
		"years_in_business": 0.12,
		"employee_count":    0.10,
		"has_website":       0.08,
		"has_email":         0.05,
		"has_phone":         0.03,
		"name_length":       0.02,
	}

	// Mock tree structures (simplified)
	xgb.treeStructures = []TreeStructure{
		{
			Nodes: []TreeNode{
				{FeatureIndex: 2, Threshold: 0.5, LeftChild: 1, RightChild: 2, IsLeaf: false},
				{FeatureIndex: -1, Threshold: 0, LeftChild: -1, RightChild: -1, LeafValue: 0.3, IsLeaf: true},
				{FeatureIndex: -1, Threshold: 0, LeftChild: -1, RightChild: -1, LeafValue: 0.7, IsLeaf: true},
			},
		},
	}

	return nil
}

// SaveModel saves the model to storage
func (xgb *XGBoostModel) SaveModel(ctx context.Context, modelPath string) error {
	// In a real implementation, this would save the actual XGBoost model
	// For now, we'll just return success
	return nil
}

// ValidateModel validates the model performance
func (xgb *XGBoostModel) ValidateModel(ctx context.Context, testData []*models.RiskAssessment) (*ValidationResult, error) {
	if !xgb.trained {
		return nil, fmt.Errorf("model is not trained")
	}

	// Mock validation results
	return &ValidationResult{
		Accuracy:  0.85,
		Precision: 0.82,
		Recall:    0.88,
		F1Score:   0.85,
		ConfusionMatrix: map[string]map[string]int{
			"low":      {"low": 45, "medium": 5, "high": 0, "critical": 0},
			"medium":   {"low": 8, "medium": 52, "high": 10, "critical": 0},
			"high":     {"low": 2, "medium": 12, "high": 38, "critical": 8},
			"critical": {"low": 0, "medium": 1, "high": 6, "critical": 13},
		},
		FeatureImportance: xgb.featureImportance,
		ValidationDate:    time.Now(),
		TestDataSize:      len(testData),
	}, nil
}

// predict makes a prediction using the XGBoost model
func (xgb *XGBoostModel) predict(features []float64) (float64, error) {
	if len(features) == 0 {
		return 0, fmt.Errorf("no features provided")
	}

	// Simplified XGBoost prediction (ensemble of decision trees)
	prediction := xgb.baseScore

	for _, tree := range xgb.treeStructures {
		treePrediction := xgb.predictTree(features, tree)
		prediction += xgb.learningRate * treePrediction
	}

	// Apply sigmoid activation to get probability-like output
	prediction = 1.0 / (1.0 + math.Exp(-prediction))

	// Ensure prediction is in valid range
	if prediction < 0 {
		prediction = 0
	} else if prediction > 1 {
		prediction = 1
	}

	return prediction, nil
}

// predictTree makes a prediction using a single decision tree
func (xgb *XGBoostModel) predictTree(features []float64, tree TreeStructure) float64 {
	if len(tree.Nodes) == 0 {
		return 0
	}

	nodeIndex := 0
	for {
		node := tree.Nodes[nodeIndex]
		if node.IsLeaf {
			return node.LeafValue
		}

		if node.FeatureIndex >= 0 && node.FeatureIndex < len(features) {
			if features[node.FeatureIndex] <= node.Threshold {
				nodeIndex = node.LeftChild
			} else {
				nodeIndex = node.RightChild
			}
		} else {
			// Invalid feature index, return default value
			return 0
		}

		if nodeIndex < 0 || nodeIndex >= len(tree.Nodes) {
			// Invalid node index, return default value
			return 0
		}
	}
}

// adjustFeaturesForHorizon adjusts features for future prediction
func (xgb *XGBoostModel) adjustFeaturesForHorizon(features []float64, horizonMonths int) []float64 {
	adjusted := make([]float64, len(features))
	copy(adjusted, features)

	// Adjust prediction horizon feature
	if len(adjusted) > 7 {
		adjusted[7] = float64(horizonMonths) / 12.0
	}

	// Apply time-based adjustments (simplified)
	timeAdjustment := float64(horizonMonths) * 0.01 // 1% risk increase per month

	// Adjust risk-sensitive features
	if len(adjusted) > 8 && adjusted[8] > 0 { // Annual revenue
		adjusted[8] *= (1.0 + timeAdjustment)
	}

	return adjusted
}

// generateRiskFactors generates detailed risk factors with subcategories
func (xgb *XGBoostModel) generateRiskFactors(features []float64, riskScore float64) []models.RiskFactor {
	// Use the comprehensive detailed risk factors from the models package
	// Create a mock business request for the detailed risk factor generation
	business := &models.RiskAssessmentRequest{
		BusinessName:    "Assessment Target",
		BusinessAddress: "Unknown",
		Industry:        "general",
		Country:         "US",
	}

	// Generate detailed risk factors with subcategories
	detailedFactors := models.GenerateDetailedRiskFactors(business, riskScore)

	// Enhance with XGBoost-specific feature importance
	enhancedFactors := xgb.enhanceRiskFactorsWithFeatureImportance(detailedFactors, features)

	return enhancedFactors
}

// enhanceRiskFactorsWithFeatureImportance enhances detailed risk factors with XGBoost feature importance
func (xgb *XGBoostModel) enhanceRiskFactorsWithFeatureImportance(detailedFactors []models.RiskFactor, features []float64) []models.RiskFactor {
	enhancedFactors := make([]models.RiskFactor, 0, len(detailedFactors))

	for _, factor := range detailedFactors {
		// Enhance with XGBoost feature importance if available
		if importance, exists := xgb.featureImportance[factor.Name]; exists {
			factor.Weight = (factor.Weight + importance) / 2 // Average with XGBoost importance
			factor.Confidence = math.Max(factor.Confidence, importance)
			factor.Source = "xgboost_enhanced_model"
		}

		// Add XGBoost-specific insights
		factor.Description = xgb.enhanceFactorDescription(factor)

		enhancedFactors = append(enhancedFactors, factor)
	}

	return enhancedFactors
}

// enhanceFactorDescription enhances factor description with XGBoost insights
func (xgb *XGBoostModel) enhanceFactorDescription(factor models.RiskFactor) string {
	baseDesc := factor.Description

	// Add XGBoost-specific insights
	if importance, exists := xgb.featureImportance[factor.Name]; exists {
		if importance > 0.1 {
			baseDesc += " (High importance in XGBoost model)"
		} else if importance > 0.05 {
			baseDesc += " (Moderate importance in XGBoost model)"
		}
	}

	return baseDesc
}

// generateScenarioAnalysis generates scenario analysis for future predictions
func (xgb *XGBoostModel) generateScenarioAnalysis(features []float64, horizonMonths int) []models.ScenarioAnalysis {
	scenarios := []models.ScenarioAnalysis{
		{
			ScenarioName: "optimistic",
			Description:  "Best case scenario with favorable market conditions",
			RiskScore:    math.Max(0, features[0]*0.8), // Reduce risk by 20%
			RiskLevel:    xgb.riskLevelEncoder.DecodeRiskLevel(features[0] * 0.8),
			Probability:  0.3,
			Impact:       "Low impact on business operations",
		},
		{
			ScenarioName: "realistic",
			Description:  "Most likely scenario based on current trends",
			RiskScore:    features[0],
			RiskLevel:    xgb.riskLevelEncoder.DecodeRiskLevel(features[0]),
			Probability:  0.5,
			Impact:       "Moderate impact on business operations",
		},
		{
			ScenarioName: "pessimistic",
			Description:  "Worst case scenario with adverse market conditions",
			RiskScore:    math.Min(1, features[0]*1.3), // Increase risk by 30%
			RiskLevel:    xgb.riskLevelEncoder.DecodeRiskLevel(math.Min(1, features[0]*1.3)),
			Probability:  0.2,
			Impact:       "High impact on business operations",
		},
	}

	return scenarios
}

// calculateConfidence calculates confidence score for the prediction
func (xgb *XGBoostModel) calculateConfidence(features []float64) float64 {
	// Simplified confidence calculation based on feature completeness
	confidence := 0.5 // Base confidence

	// Increase confidence for complete features
	if len(features) > 4 && features[4] > 0 { // Has phone
		confidence += 0.1
	}
	if len(features) > 5 && features[5] > 0 { // Has email
		confidence += 0.1
	}
	if len(features) > 6 && features[6] > 0 { // Has website
		confidence += 0.1
	}
	if len(features) > 8 && features[8] > 0 { // Has revenue data
		confidence += 0.1
	}
	if len(features) > 9 && features[9] > 0 { // Has employee data
		confidence += 0.1
	}

	// Ensure confidence is in valid range
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// getRiskCategory maps feature names to risk categories
func (xgb *XGBoostModel) getRiskCategory(featureName string) models.RiskCategory {
	switch featureName {
	case "industry_code", "country_code":
		return models.RiskCategoryRegulatory
	case "annual_revenue", "employee_count":
		return models.RiskCategoryFinancial
	case "years_in_business":
		return models.RiskCategoryOperational
	case "has_website", "has_email", "has_phone":
		return models.RiskCategoryCompliance
	default:
		return models.RiskCategoryOperational
	}
}

// calculateFeatureScore calculates the risk score for a specific feature
func (xgb *XGBoostModel) calculateFeatureScore(features []float64, featureName string, featureNames []string) float64 {
	// Find feature index
	featureIndex := -1
	for i, name := range featureNames {
		if name == featureName {
			featureIndex = i
			break
		}
	}

	if featureIndex == -1 || featureIndex >= len(features) {
		return 0.5 // Default score
	}

	// Normalize feature value to risk score
	value := features[featureIndex]

	// Apply feature-specific normalization
	switch featureName {
	case "industry_code":
		// Higher industry codes (more regulated industries) = higher risk
		return value / 10.0
	case "country_code":
		// Higher country codes (less stable countries) = higher risk
		return value / 10.0
	case "annual_revenue":
		// Higher revenue = lower risk (more established)
		return 1.0 - math.Min(1.0, value)
	case "employee_count":
		// More employees = lower risk (more established)
		return 1.0 - math.Min(1.0, value)
	case "years_in_business":
		// More years = lower risk (more established)
		return 1.0 - math.Min(1.0, value)
	default:
		// For binary features, return the value directly
		return value
	}
}

// getFeatureDescription returns a description for a feature
func (xgb *XGBoostModel) getFeatureDescription(featureName string) string {
	descriptions := map[string]string{
		"industry_code":     "Industry classification and regulatory environment",
		"country_code":      "Country of operation and political stability",
		"annual_revenue":    "Annual revenue as indicator of business stability",
		"employee_count":    "Number of employees as indicator of business size",
		"years_in_business": "Years in business as indicator of experience",
		"has_website":       "Presence of professional website",
		"has_email":         "Presence of business email address",
		"has_phone":         "Presence of business phone number",
		"name_length":       "Business name length as indicator of professionalism",
		"address_length":    "Address completeness as indicator of legitimacy",
	}

	if desc, exists := descriptions[featureName]; exists {
		return desc
	}

	return fmt.Sprintf("Risk factor: %s", featureName)
}
