package ensemble

import (
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// EnsembleCombiner combines predictions from multiple models
type EnsembleCombiner struct {
	logger *zap.Logger
}

// NewEnsembleCombiner creates a new ensemble combiner
func NewEnsembleCombiner() *EnsembleCombiner {
	return &EnsembleCombiner{
		logger: zap.NewNop(),
	}
}

// SetLogger sets the logger for the ensemble combiner
func (ec *EnsembleCombiner) SetLogger(logger *zap.Logger) {
	ec.logger = logger
}

// CombinePredictions combines risk assessments from multiple models
func (ec *EnsembleCombiner) CombinePredictions(xgbPrediction, lstmPrediction *models.RiskAssessment, horizonMonths int) (*models.RiskAssessment, error) {
	if xgbPrediction == nil && lstmPrediction == nil {
		return nil, fmt.Errorf("no predictions to combine")
	}

	// If only one prediction, return it
	if xgbPrediction == nil {
		return lstmPrediction, nil
	}
	if lstmPrediction == nil {
		return xgbPrediction, nil
	}

	// Get weights based on horizon
	xgbWeight, lstmWeight := ec.getHorizonWeights(horizonMonths)

	// Combine risk scores
	combinedRiskScore := xgbPrediction.RiskScore*xgbWeight + lstmPrediction.RiskScore*lstmWeight

	// Combine confidence scores (weighted average)
	combinedConfidence := xgbPrediction.ConfidenceScore*xgbWeight + lstmPrediction.ConfidenceScore*lstmWeight

	// Determine combined risk level
	combinedRiskLevel := ec.determineRiskLevel(combinedRiskScore)

	// Combine risk factors
	combinedRiskFactors := ec.combineRiskFactors(xgbPrediction.RiskFactors, lstmPrediction.RiskFactors, xgbWeight, lstmWeight)

	// Create combined assessment
	combinedAssessment := &models.RiskAssessment{
		ID:                fmt.Sprintf("ensemble_%d", time.Now().UnixNano()),
		BusinessID:        xgbPrediction.BusinessID,
		BusinessName:      xgbPrediction.BusinessName,
		BusinessAddress:   xgbPrediction.BusinessAddress,
		Industry:          xgbPrediction.Industry,
		Country:           xgbPrediction.Country,
		RiskScore:         combinedRiskScore,
		RiskLevel:         combinedRiskLevel,
		RiskFactors:       combinedRiskFactors,
		PredictionHorizon: horizonMonths,
		ConfidenceScore:   combinedConfidence,
		Status:            models.StatusCompleted,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Metadata: map[string]interface{}{
			"model_type":         "ensemble",
			"xgb_weight":         xgbWeight,
			"lstm_weight":        lstmWeight,
			"xgb_risk_score":     xgbPrediction.RiskScore,
			"lstm_risk_score":    lstmPrediction.RiskScore,
			"xgb_confidence":     xgbPrediction.ConfidenceScore,
			"lstm_confidence":    lstmPrediction.ConfidenceScore,
			"prediction_horizon": horizonMonths,
		},
	}

	ec.logger.Debug("Combined predictions",
		zap.Float64("xgb_weight", xgbWeight),
		zap.Float64("lstm_weight", lstmWeight),
		zap.Float64("combined_risk_score", combinedRiskScore),
		zap.Float64("combined_confidence", combinedConfidence))

	return combinedAssessment, nil
}

// CombineFuturePredictions combines future risk predictions from multiple models
func (ec *EnsembleCombiner) CombineFuturePredictions(xgbPrediction, lstmPrediction *models.RiskPrediction, horizonMonths int) (*models.RiskPrediction, error) {
	if xgbPrediction == nil && lstmPrediction == nil {
		return nil, fmt.Errorf("no future predictions to combine")
	}

	// If only one prediction, return it
	if xgbPrediction == nil {
		return lstmPrediction, nil
	}
	if lstmPrediction == nil {
		return xgbPrediction, nil
	}

	// Get weights based on horizon
	xgbWeight, lstmWeight := ec.getHorizonWeights(horizonMonths)

	// Combine predicted scores
	combinedPredictedScore := xgbPrediction.PredictedScore*xgbWeight + lstmPrediction.PredictedScore*lstmWeight

	// Combine confidence scores
	combinedConfidence := xgbPrediction.ConfidenceScore*xgbWeight + lstmPrediction.ConfidenceScore*lstmWeight

	// Determine combined predicted level
	combinedPredictedLevel := ec.determineRiskLevel(combinedPredictedScore)

	// Combine risk factors
	combinedRiskFactors := ec.combineRiskFactors(xgbPrediction.RiskFactors, lstmPrediction.RiskFactors, xgbWeight, lstmWeight)

	// Combine scenario analysis
	combinedScenarios := ec.combineScenarioAnalysis(xgbPrediction.ScenarioAnalysis, lstmPrediction.ScenarioAnalysis, xgbWeight, lstmWeight)

	// Create combined prediction
	combinedPrediction := &models.RiskPrediction{
		BusinessID:       fmt.Sprintf("ensemble_biz_%d", time.Now().UnixNano()),
		PredictionDate:   time.Now(),
		HorizonMonths:    horizonMonths,
		PredictedScore:   combinedPredictedScore,
		PredictedLevel:   combinedPredictedLevel,
		ConfidenceScore:  combinedConfidence,
		RiskFactors:      combinedRiskFactors,
		ScenarioAnalysis: combinedScenarios,
		CreatedAt:        time.Now(),
	}

	ec.logger.Debug("Combined future predictions",
		zap.Float64("xgb_weight", xgbWeight),
		zap.Float64("lstm_weight", lstmWeight),
		zap.Float64("combined_predicted_score", combinedPredictedScore),
		zap.Float64("combined_confidence", combinedConfidence))

	return combinedPrediction, nil
}

// getHorizonWeights returns the weights for XGBoost and LSTM based on prediction horizon
func (ec *EnsembleCombiner) getHorizonWeights(horizonMonths int) (float64, float64) {
	switch {
	case horizonMonths <= 3:
		// Short-term: heavily favor XGBoost
		return 0.8, 0.2
	case horizonMonths >= 6:
		// Long-term: heavily favor LSTM
		return 0.2, 0.8
	default:
		// Medium-term: balanced ensemble
		return 0.5, 0.5
	}
}

// determineRiskLevel determines the risk level based on risk score
func (ec *EnsembleCombiner) determineRiskLevel(riskScore float64) models.RiskLevel {
	switch {
	case riskScore <= 0.25:
		return models.RiskLevelLow
	case riskScore <= 0.5:
		return models.RiskLevelMedium
	case riskScore <= 0.75:
		return models.RiskLevelHigh
	default:
		return models.RiskLevelCritical
	}
}

// combineRiskFactors combines risk factors from multiple models
func (ec *EnsembleCombiner) combineRiskFactors(xgbFactors, lstmFactors []models.RiskFactor, xgbWeight, lstmWeight float64) []models.RiskFactor {
	// Create a map to store combined factors
	factorMap := make(map[string]*models.RiskFactor)

	// Process XGBoost factors
	for _, factor := range xgbFactors {
		combinedFactor := &models.RiskFactor{
			Category:    factor.Category,
			Name:        factor.Name,
			Score:       factor.Score * xgbWeight,
			Weight:      factor.Weight * xgbWeight,
			Description: factor.Description,
			Source:      "ensemble_xgb",
			Confidence:  factor.Confidence * xgbWeight,
		}
		factorMap[factor.Name] = combinedFactor
	}

	// Process LSTM factors and combine
	for _, factor := range lstmFactors {
		if existingFactor, exists := factorMap[factor.Name]; exists {
			// Combine with existing factor
			existingFactor.Score += factor.Score * lstmWeight
			existingFactor.Weight += factor.Weight * lstmWeight
			existingFactor.Confidence += factor.Confidence * lstmWeight
			existingFactor.Source = "ensemble_combined"
		} else {
			// Add new factor
			combinedFactor := &models.RiskFactor{
				Category:    factor.Category,
				Name:        factor.Name,
				Score:       factor.Score * lstmWeight,
				Weight:      factor.Weight * lstmWeight,
				Description: factor.Description,
				Source:      "ensemble_lstm",
				Confidence:  factor.Confidence * lstmWeight,
			}
			factorMap[factor.Name] = combinedFactor
		}
	}

	// Convert map back to slice
	combinedFactors := make([]models.RiskFactor, 0, len(factorMap))
	for _, factor := range factorMap {
		combinedFactors = append(combinedFactors, *factor)
	}

	return combinedFactors
}

// combineScenarioAnalysis combines scenario analysis from multiple models
func (ec *EnsembleCombiner) combineScenarioAnalysis(xgbScenarios, lstmScenarios []models.ScenarioAnalysis, xgbWeight, lstmWeight float64) []models.ScenarioAnalysis {
	// Create a map to store combined scenarios
	scenarioMap := make(map[string]*models.ScenarioAnalysis)

	// Process XGBoost scenarios
	for _, scenario := range xgbScenarios {
		combinedScenario := &models.ScenarioAnalysis{
			ScenarioName: scenario.ScenarioName,
			Description:  scenario.Description,
			RiskScore:    scenario.RiskScore * xgbWeight,
			RiskLevel:    ec.determineRiskLevel(scenario.RiskScore * xgbWeight),
			Probability:  scenario.Probability,
			Impact:       scenario.Impact,
		}
		scenarioMap[scenario.ScenarioName] = combinedScenario
	}

	// Process LSTM scenarios and combine
	for _, scenario := range lstmScenarios {
		if existingScenario, exists := scenarioMap[scenario.ScenarioName]; exists {
			// Combine with existing scenario
			existingScenario.RiskScore += scenario.RiskScore * lstmWeight
			existingScenario.RiskLevel = ec.determineRiskLevel(existingScenario.RiskScore)
			// Keep the average probability
			existingScenario.Probability = (existingScenario.Probability + scenario.Probability) / 2
		} else {
			// Add new scenario
			combinedScenario := &models.ScenarioAnalysis{
				ScenarioName: scenario.ScenarioName,
				Description:  scenario.Description,
				RiskScore:    scenario.RiskScore * lstmWeight,
				RiskLevel:    ec.determineRiskLevel(scenario.RiskScore * lstmWeight),
				Probability:  scenario.Probability,
				Impact:       scenario.Impact,
			}
			scenarioMap[scenario.ScenarioName] = combinedScenario
		}
	}

	// Convert map back to slice
	combinedScenarios := make([]models.ScenarioAnalysis, 0, len(scenarioMap))
	for _, scenario := range scenarioMap {
		combinedScenarios = append(combinedScenarios, *scenario)
	}

	return combinedScenarios
}

// CalculateModelAgreement calculates the agreement between two model predictions
func (ec *EnsembleCombiner) CalculateModelAgreement(xgbPrediction, lstmPrediction *models.RiskAssessment) float64 {
	if xgbPrediction == nil || lstmPrediction == nil {
		return 0.0
	}

	// Calculate agreement based on risk score difference
	scoreDiff := math.Abs(xgbPrediction.RiskScore - lstmPrediction.RiskScore)
	agreement := 1.0 - scoreDiff

	// Adjust for risk level agreement
	if xgbPrediction.RiskLevel == lstmPrediction.RiskLevel {
		agreement += 0.2 // Bonus for exact risk level match
	}

	// Ensure agreement is in valid range
	if agreement < 0 {
		agreement = 0
	} else if agreement > 1 {
		agreement = 1
	}

	return agreement
}

// CalculateEnsembleConfidence calculates the confidence of the ensemble prediction
func (ec *EnsembleCombiner) CalculateEnsembleConfidence(xgbPrediction, lstmPrediction *models.RiskAssessment, agreement float64) float64 {
	if xgbPrediction == nil && lstmPrediction == nil {
		return 0.0
	}

	var baseConfidence float64
	if xgbPrediction != nil && lstmPrediction != nil {
		// Average confidence of both models
		baseConfidence = (xgbPrediction.ConfidenceScore + lstmPrediction.ConfidenceScore) / 2
	} else if xgbPrediction != nil {
		baseConfidence = xgbPrediction.ConfidenceScore
	} else {
		baseConfidence = lstmPrediction.ConfidenceScore
	}

	// Adjust confidence based on model agreement
	// Higher agreement = higher confidence
	adjustedConfidence := baseConfidence * (0.5 + 0.5*agreement)

	// Ensure confidence is in valid range
	if adjustedConfidence < 0.1 {
		adjustedConfidence = 0.1
	} else if adjustedConfidence > 1.0 {
		adjustedConfidence = 1.0
	}

	return adjustedConfidence
}

// GetEnsembleMetrics returns metrics about the ensemble combination
func (ec *EnsembleCombiner) GetEnsembleMetrics(xgbPrediction, lstmPrediction *models.RiskAssessment) map[string]interface{} {
	metrics := make(map[string]interface{})

	if xgbPrediction != nil && lstmPrediction != nil {
		agreement := ec.CalculateModelAgreement(xgbPrediction, lstmPrediction)
		ensembleConfidence := ec.CalculateEnsembleConfidence(xgbPrediction, lstmPrediction, agreement)

		metrics["model_agreement"] = agreement
		metrics["ensemble_confidence"] = ensembleConfidence
		metrics["xgb_risk_score"] = xgbPrediction.RiskScore
		metrics["lstm_risk_score"] = lstmPrediction.RiskScore
		metrics["risk_score_difference"] = math.Abs(xgbPrediction.RiskScore - lstmPrediction.RiskScore)
		metrics["xgb_confidence"] = xgbPrediction.ConfidenceScore
		metrics["lstm_confidence"] = lstmPrediction.ConfidenceScore
		metrics["risk_level_agreement"] = xgbPrediction.RiskLevel == lstmPrediction.RiskLevel
	} else if xgbPrediction != nil {
		metrics["single_model"] = "xgboost"
		metrics["risk_score"] = xgbPrediction.RiskScore
		metrics["confidence"] = xgbPrediction.ConfidenceScore
	} else if lstmPrediction != nil {
		metrics["single_model"] = "lstm"
		metrics["risk_score"] = lstmPrediction.RiskScore
		metrics["confidence"] = lstmPrediction.ConfidenceScore
	}

	return metrics
}
