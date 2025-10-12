package explainability

import (
	"context"
	"fmt"
	"math"
	"sort"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// SHAPExplainer provides SHAP-like explainability for risk assessment models
type SHAPExplainer struct {
	featureNames []string
	logger       *zap.Logger
}

// NewSHAPExplainer creates a new SHAP explainer
func NewSHAPExplainer(featureNames []string, logger *zap.Logger) *SHAPExplainer {
	return &SHAPExplainer{
		featureNames: featureNames,
		logger:       logger,
	}
}

// SHAPExplanation represents the explanation for a prediction
type SHAPExplanation struct {
	PredictionValue      float64               `json:"prediction_value"`
	BaseValue            float64               `json:"base_value"`
	FeatureContributions []FeatureContribution `json:"feature_contributions"`
	TotalContribution    float64               `json:"total_contribution"`
	Confidence           float64               `json:"confidence"`
	ExplanationType      string                `json:"explanation_type"`
}

// FeatureContribution represents the contribution of a single feature
type FeatureContribution struct {
	FeatureName  string  `json:"feature_name"`
	FeatureValue float64 `json:"feature_value"`
	Contribution float64 `json:"contribution"`
	Importance   float64 `json:"importance"`
	Direction    string  `json:"direction"` // "positive", "negative", "neutral"
	Description  string  `json:"description"`
	Confidence   float64 `json:"confidence"`
}

// ExplainPrediction generates a SHAP-like explanation for a risk prediction
func (se *SHAPExplainer) ExplainPrediction(ctx context.Context, business *models.RiskAssessmentRequest, features []float64, prediction float64, featureImportance map[string]float64) (*SHAPExplanation, error) {
	se.logger.Info("Generating SHAP-like explanation for prediction",
		zap.String("business_name", business.BusinessName),
		zap.Float64("prediction", prediction))

	// Calculate base value (average prediction across all possible inputs)
	baseValue := se.calculateBaseValue(features)

	// Calculate feature contributions using marginal impact
	contributions := se.calculateFeatureContributions(features, prediction, featureImportance)

	// Calculate total contribution
	totalContribution := 0.0
	for _, contrib := range contributions {
		totalContribution += contrib.Contribution
	}

	// Calculate confidence based on feature completeness and importance distribution
	confidence := se.calculateExplanationConfidence(features, contributions)

	explanation := &SHAPExplanation{
		PredictionValue:      prediction,
		BaseValue:            baseValue,
		FeatureContributions: contributions,
		TotalContribution:    totalContribution,
		Confidence:           confidence,
		ExplanationType:      "shap_like",
	}

	se.logger.Info("SHAP-like explanation generated",
		zap.Float64("base_value", baseValue),
		zap.Float64("total_contribution", totalContribution),
		zap.Float64("confidence", confidence))

	return explanation, nil
}

// calculateBaseValue calculates the base value for SHAP explanation
func (se *SHAPExplainer) calculateBaseValue(features []float64) float64 {
	// In a full SHAP implementation, this would be the expected value across all possible inputs
	// For our simplified version, we use the mean of typical business risk scores
	return 0.5 // Neutral base value
}

// calculateFeatureContributions calculates the contribution of each feature
func (se *SHAPExplainer) calculateFeatureContributions(features []float64, prediction float64, featureImportance map[string]float64) []FeatureContribution {
	contributions := make([]FeatureContribution, 0, len(features))

	for i, featureValue := range features {
		if i >= len(se.featureNames) {
			break
		}

		featureName := se.featureNames[i]
		importance, exists := featureImportance[featureName]
		if !exists {
			importance = 0.1 // Default importance
		}

		// Calculate marginal contribution
		contribution := se.calculateMarginalContribution(featureValue, importance, prediction)

		// Determine direction
		direction := "neutral"
		if contribution > 0.01 {
			direction = "positive"
		} else if contribution < -0.01 {
			direction = "negative"
		}

		// Generate description
		description := se.generateFeatureDescription(featureName, featureValue, contribution)

		// Calculate confidence for this feature
		featureConfidence := se.calculateFeatureConfidence(featureValue, importance)

		contrib := FeatureContribution{
			FeatureName:  featureName,
			FeatureValue: featureValue,
			Contribution: contribution,
			Importance:   importance,
			Direction:    direction,
			Description:  description,
			Confidence:   featureConfidence,
		}

		contributions = append(contributions, contrib)
	}

	// Sort by absolute contribution (most important first)
	sort.Slice(contributions, func(i, j int) bool {
		return math.Abs(contributions[i].Contribution) > math.Abs(contributions[j].Contribution)
	})

	return contributions
}

// calculateMarginalContribution calculates the marginal contribution of a feature
func (se *SHAPExplainer) calculateMarginalContribution(featureValue, importance, prediction float64) float64 {
	// Simplified marginal contribution calculation
	// In full SHAP, this would involve calculating the difference between predictions
	// with and without the feature across all possible feature combinations

	// Base contribution based on feature importance
	baseContribution := importance * (featureValue - 0.5) // 0.5 is neutral value

	// Adjust based on prediction value
	// Higher predictions get more positive contributions from high feature values
	if prediction > 0.5 {
		baseContribution *= 1.2
	} else {
		baseContribution *= 0.8
	}

	// Add some non-linearity
	nonLinearAdjustment := math.Sin(featureValue*math.Pi) * 0.1
	baseContribution += nonLinearAdjustment

	return baseContribution
}

// generateFeatureDescription generates a human-readable description for a feature contribution
func (se *SHAPExplainer) generateFeatureDescription(featureName string, featureValue, contribution float64) string {
	// Generate contextual descriptions based on feature name and contribution
	switch featureName {
	case "industry_code":
		if contribution > 0 {
			return fmt.Sprintf("Industry classification indicates higher risk sector (contribution: +%.3f)", contribution)
		}
		return fmt.Sprintf("Industry classification indicates lower risk sector (contribution: %.3f)", contribution)

	case "country_code":
		if contribution > 0 {
			return fmt.Sprintf("Country of operation has higher risk profile (contribution: +%.3f)", contribution)
		}
		return fmt.Sprintf("Country of operation has lower risk profile (contribution: %.3f)", contribution)

	case "annual_revenue":
		if contribution > 0 {
			return fmt.Sprintf("Revenue level suggests higher risk (contribution: +%.3f)", contribution)
		}
		return fmt.Sprintf("Revenue level suggests lower risk (contribution: %.3f)", contribution)

	case "employee_count":
		if contribution > 0 {
			return fmt.Sprintf("Company size indicates higher risk (contribution: +%.3f)", contribution)
		}
		return fmt.Sprintf("Company size indicates lower risk (contribution: %.3f)", contribution)

	case "years_in_business":
		if contribution > 0 {
			return fmt.Sprintf("Business age suggests higher risk (contribution: +%.3f)", contribution)
		}
		return fmt.Sprintf("Business age suggests lower risk (contribution: %.3f)", contribution)

	case "has_website":
		if contribution > 0 {
			return "Website presence increases risk assessment"
		}
		return "Website presence decreases risk assessment"

	case "has_email":
		if contribution > 0 {
			return "Email presence increases risk assessment"
		}
		return "Email presence decreases risk assessment"

	case "has_phone":
		if contribution > 0 {
			return "Phone presence increases risk assessment"
		}
		return "Phone presence decreases risk assessment"

	default:
		if contribution > 0 {
			return fmt.Sprintf("Feature '%s' contributes positively to risk (contribution: +%.3f)", featureName, contribution)
		}
		return fmt.Sprintf("Feature '%s' contributes negatively to risk (contribution: %.3f)", featureName, contribution)
	}
}

// calculateFeatureConfidence calculates confidence for a specific feature contribution
func (se *SHAPExplainer) calculateFeatureConfidence(featureValue, importance float64) float64 {
	// Confidence based on feature importance and value completeness
	baseConfidence := importance

	// Adjust based on feature value (extreme values are less confident)
	if featureValue < 0.1 || featureValue > 0.9 {
		baseConfidence *= 0.8
	}

	// Adjust based on importance (more important features have higher confidence)
	if importance > 0.2 {
		baseConfidence *= 1.1
	}

	return math.Max(0.1, math.Min(1.0, baseConfidence))
}

// calculateExplanationConfidence calculates overall confidence in the explanation
func (se *SHAPExplainer) calculateExplanationConfidence(features []float64, contributions []FeatureContribution) float64 {
	// Base confidence
	confidence := 0.8

	// Adjust based on feature completeness
	completeFeatures := 0
	for _, feature := range features {
		if feature > 0 {
			completeFeatures++
		}
	}
	completenessRatio := float64(completeFeatures) / float64(len(features))
	confidence += completenessRatio * 0.1

	// Adjust based on contribution consistency
	if len(contributions) > 0 {
		// Check if contributions are well-distributed
		totalAbsContribution := 0.0
		for _, contrib := range contributions {
			totalAbsContribution += math.Abs(contrib.Contribution)
		}

		if totalAbsContribution > 0.5 {
			confidence += 0.05 // Good contribution distribution
		}
	}

	return math.Max(0.3, math.Min(1.0, confidence))
}

// ExplainRiskFactors generates explanations for risk factors
func (se *SHAPExplainer) ExplainRiskFactors(ctx context.Context, riskFactors []models.RiskFactor, business *models.RiskAssessmentRequest) ([]RiskFactorExplanation, error) {
	explanations := make([]RiskFactorExplanation, 0, len(riskFactors))

	for _, factor := range riskFactors {
		explanation := se.explainRiskFactor(factor, business)
		explanations = append(explanations, explanation)
	}

	// Sort by importance (weight * score)
	sort.Slice(explanations, func(i, j int) bool {
		importanceI := explanations[i].RiskFactor.Weight * explanations[i].RiskFactor.Score
		importanceJ := explanations[j].RiskFactor.Weight * explanations[j].RiskFactor.Score
		return importanceI > importanceJ
	})

	return explanations, nil
}

// RiskFactorExplanation provides detailed explanation for a risk factor
type RiskFactorExplanation struct {
	RiskFactor     models.RiskFactor `json:"risk_factor"`
	Explanation    string            `json:"explanation"`
	Impact         string            `json:"impact"`
	Recommendation string            `json:"recommendation"`
	Confidence     float64           `json:"confidence"`
}

// explainRiskFactor generates explanation for a single risk factor
func (se *SHAPExplainer) explainRiskFactor(factor models.RiskFactor, business *models.RiskAssessmentRequest) RiskFactorExplanation {
	explanation := se.generateRiskFactorExplanation(factor, business)
	impact := se.generateRiskFactorImpact(factor)
	recommendation := se.generateRiskFactorRecommendation(factor, business)
	confidence := factor.Confidence

	return RiskFactorExplanation{
		RiskFactor:     factor,
		Explanation:    explanation,
		Impact:         impact,
		Recommendation: recommendation,
		Confidence:     confidence,
	}
}

// generateRiskFactorExplanation generates explanation for a risk factor
func (se *SHAPExplainer) generateRiskFactorExplanation(factor models.RiskFactor, business *models.RiskAssessmentRequest) string {
	switch factor.Category {
	case models.RiskCategoryFinancial:
		return fmt.Sprintf("Financial risk factor '%s' indicates %s financial health with a score of %.2f. %s",
			factor.Name, se.getRiskLevelDescription(factor.Score), factor.Score, factor.Description)

	case models.RiskCategoryOperational:
		return fmt.Sprintf("Operational risk factor '%s' shows %s operational stability with a score of %.2f. %s",
			factor.Name, se.getRiskLevelDescription(factor.Score), factor.Score, factor.Description)

	case models.RiskCategoryCompliance:
		return fmt.Sprintf("Compliance risk factor '%s' indicates %s compliance posture with a score of %.2f. %s",
			factor.Name, se.getRiskLevelDescription(factor.Score), factor.Score, factor.Description)

	case models.RiskCategoryReputational:
		return fmt.Sprintf("Reputational risk factor '%s' shows %s reputation status with a score of %.2f. %s",
			factor.Name, se.getRiskLevelDescription(factor.Score), factor.Score, factor.Description)

	case models.RiskCategoryRegulatory:
		return fmt.Sprintf("Regulatory risk factor '%s' indicates %s regulatory compliance with a score of %.2f. %s",
			factor.Name, se.getRiskLevelDescription(factor.Score), factor.Score, factor.Description)

	default:
		return fmt.Sprintf("Risk factor '%s' in category '%s' has a score of %.2f. %s",
			factor.Name, factor.Category, factor.Score, factor.Description)
	}
}

// generateRiskFactorImpact generates impact description for a risk factor
func (se *SHAPExplainer) generateRiskFactorImpact(factor models.RiskFactor) string {
	impact := factor.Weight * factor.Score

	if impact > 0.3 {
		return "High impact on overall risk assessment"
	} else if impact > 0.15 {
		return "Moderate impact on overall risk assessment"
	} else {
		return "Low impact on overall risk assessment"
	}
}

// generateRiskFactorRecommendation generates recommendation for a risk factor
func (se *SHAPExplainer) generateRiskFactorRecommendation(factor models.RiskFactor, business *models.RiskAssessmentRequest) string {
	if factor.Score > 0.7 {
		return fmt.Sprintf("Consider implementing additional controls for %s risk management", factor.Category)
	} else if factor.Score > 0.4 {
		return fmt.Sprintf("Monitor %s risk factors closely and maintain current controls", factor.Category)
	} else {
		return fmt.Sprintf("Continue current %s risk management practices", factor.Category)
	}
}

// getRiskLevelDescription returns a description for a risk score
func (se *SHAPExplainer) getRiskLevelDescription(score float64) string {
	if score > 0.8 {
		return "critical"
	} else if score > 0.6 {
		return "high"
	} else if score > 0.4 {
		return "moderate"
	} else {
		return "low"
	}
}
