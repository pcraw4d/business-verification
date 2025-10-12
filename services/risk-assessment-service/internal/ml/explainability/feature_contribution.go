package explainability

import (
	"context"
	"fmt"
	"math"
	"sort"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// FeatureContributionAnalyzer analyzes feature contributions to risk predictions
type FeatureContributionAnalyzer struct {
	featureWeights map[string]float64
	logger         *zap.Logger
}

// NewFeatureContributionAnalyzer creates a new feature contribution analyzer
func NewFeatureContributionAnalyzer(logger *zap.Logger) *FeatureContributionAnalyzer {
	return &FeatureContributionAnalyzer{
		featureWeights: map[string]float64{
			"industry_code":     0.25,
			"country_code":      0.20,
			"annual_revenue":    0.15,
			"years_in_business": 0.12,
			"employee_count":    0.10,
			"has_website":       0.08,
			"has_email":         0.05,
			"has_phone":         0.03,
			"name_length":       0.02,
		},
		logger: logger,
	}
}

// ContributionAnalysis represents the analysis of feature contributions
type ContributionAnalysis struct {
	TotalContribution     float64               `json:"total_contribution"`
	PositiveContributions []FeatureContribution `json:"positive_contributions"`
	NegativeContributions []FeatureContribution `json:"negative_contributions"`
	TopContributors       []FeatureContribution `json:"top_contributors"`
	ContributionSummary   ContributionSummary   `json:"contribution_summary"`
	RiskCategoryBreakdown map[string]float64    `json:"risk_category_breakdown"`
}

// ContributionSummary provides a summary of contributions
type ContributionSummary struct {
	PositiveTotal     float64 `json:"positive_total"`
	NegativeTotal     float64 `json:"negative_total"`
	NetContribution   float64 `json:"net_contribution"`
	ContributionRatio float64 `json:"contribution_ratio"`
	Uncertainty       float64 `json:"uncertainty"`
}

// AnalyzeContributions analyzes feature contributions to a prediction
func (fca *FeatureContributionAnalyzer) AnalyzeContributions(ctx context.Context, business *models.RiskAssessmentRequest, features []float64, prediction float64, featureNames []string) (*ContributionAnalysis, error) {
	fca.logger.Info("Analyzing feature contributions",
		zap.String("business_name", business.BusinessName),
		zap.Float64("prediction", prediction))

	// Calculate individual feature contributions
	contributions := fca.calculateFeatureContributions(features, prediction, featureNames)

	// Separate positive and negative contributions
	positiveContributions := make([]FeatureContribution, 0)
	negativeContributions := make([]FeatureContribution, 0)

	for _, contrib := range contributions {
		if contrib.Contribution > 0 {
			positiveContributions = append(positiveContributions, contrib)
		} else {
			negativeContributions = append(negativeContributions, contrib)
		}
	}

	// Sort by absolute contribution
	sort.Slice(positiveContributions, func(i, j int) bool {
		return positiveContributions[i].Contribution > positiveContributions[j].Contribution
	})
	sort.Slice(negativeContributions, func(i, j int) bool {
		return math.Abs(negativeContributions[i].Contribution) > math.Abs(negativeContributions[j].Contribution)
	})

	// Get top contributors (top 5 by absolute contribution)
	topContributors := fca.getTopContributors(contributions, 5)

	// Calculate summary statistics
	summary := fca.calculateContributionSummary(contributions)

	// Calculate risk category breakdown
	categoryBreakdown := fca.calculateRiskCategoryBreakdown(contributions, business)

	// Calculate total contribution
	totalContribution := 0.0
	for _, contrib := range contributions {
		totalContribution += contrib.Contribution
	}

	analysis := &ContributionAnalysis{
		TotalContribution:     totalContribution,
		PositiveContributions: positiveContributions,
		NegativeContributions: negativeContributions,
		TopContributors:       topContributors,
		ContributionSummary:   summary,
		RiskCategoryBreakdown: categoryBreakdown,
	}

	fca.logger.Info("Feature contribution analysis completed",
		zap.Float64("total_contribution", totalContribution),
		zap.Int("positive_count", len(positiveContributions)),
		zap.Int("negative_count", len(negativeContributions)))

	return analysis, nil
}

// calculateFeatureContributions calculates contributions for all features
func (fca *FeatureContributionAnalyzer) calculateFeatureContributions(features []float64, prediction float64, featureNames []string) []FeatureContribution {
	contributions := make([]FeatureContribution, 0, len(features))

	for i, featureValue := range features {
		if i >= len(featureNames) {
			break
		}

		featureName := featureNames[i]
		weight, exists := fca.featureWeights[featureName]
		if !exists {
			weight = 0.05 // Default weight for unknown features
		}

		// Calculate contribution using weighted marginal impact
		contribution := fca.calculateWeightedContribution(featureValue, weight, prediction)

		// Determine direction
		direction := "neutral"
		if contribution > 0.01 {
			direction = "positive"
		} else if contribution < -0.01 {
			direction = "negative"
		}

		// Generate description
		description := fca.generateContributionDescription(featureName, featureValue, contribution)

		// Calculate confidence
		confidence := fca.calculateContributionConfidence(featureValue, weight)

		contrib := FeatureContribution{
			FeatureName:  featureName,
			FeatureValue: featureValue,
			Contribution: contribution,
			Importance:   weight,
			Direction:    direction,
			Description:  description,
			Confidence:   confidence,
		}

		contributions = append(contributions, contrib)
	}

	return contributions
}

// calculateWeightedContribution calculates weighted contribution of a feature
func (fca *FeatureContributionAnalyzer) calculateWeightedContribution(featureValue, weight, prediction float64) float64 {
	// Base contribution from feature value deviation from neutral
	neutralValue := 0.5
	deviation := featureValue - neutralValue

	// Weight the contribution
	weightedContribution := deviation * weight

	// Apply prediction-based adjustment
	// Higher predictions amplify positive contributions
	if prediction > 0.5 {
		weightedContribution *= (1.0 + (prediction-0.5)*0.5)
	} else {
		weightedContribution *= (1.0 - (0.5-prediction)*0.5)
	}

	// Add non-linear effects for extreme values
	if math.Abs(featureValue-neutralValue) > 0.3 {
		nonLinearEffect := math.Sin((featureValue-neutralValue)*math.Pi) * 0.1
		weightedContribution += nonLinearEffect
	}

	return weightedContribution
}

// generateContributionDescription generates description for a feature contribution
func (fca *FeatureContributionAnalyzer) generateContributionDescription(featureName string, featureValue, contribution float64) string {
	// Generate contextual descriptions based on feature characteristics
	switch featureName {
	case "industry_code":
		if contribution > 0 {
			return fmt.Sprintf("Operating in a higher-risk industry sector (value: %.2f, contribution: +%.3f)", featureValue, contribution)
		}
		return fmt.Sprintf("Operating in a lower-risk industry sector (value: %.2f, contribution: %.3f)", featureValue, contribution)

	case "country_code":
		if contribution > 0 {
			return fmt.Sprintf("Operating in a higher-risk country (value: %.2f, contribution: +%.3f)", featureValue, contribution)
		}
		return fmt.Sprintf("Operating in a lower-risk country (value: %.2f, contribution: %.3f)", featureValue, contribution)

	case "annual_revenue":
		if contribution > 0 {
			return fmt.Sprintf("Revenue level suggests higher risk (value: %.2f, contribution: +%.3f)", featureValue, contribution)
		}
		return fmt.Sprintf("Revenue level suggests lower risk (value: %.2f, contribution: %.3f)", featureValue, contribution)

	case "employee_count":
		if contribution > 0 {
			return fmt.Sprintf("Company size indicates higher risk (value: %.2f, contribution: +%.3f)", featureValue, contribution)
		}
		return fmt.Sprintf("Company size indicates lower risk (value: %.2f, contribution: %.3f)", featureValue, contribution)

	case "years_in_business":
		if contribution > 0 {
			return fmt.Sprintf("Business age suggests higher risk (value: %.2f, contribution: +%.3f)", featureValue, contribution)
		}
		return fmt.Sprintf("Business age suggests lower risk (value: %.2f, contribution: %.3f)", featureValue, contribution)

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

	case "name_length":
		if contribution > 0 {
			return fmt.Sprintf("Business name length suggests higher risk (value: %.2f, contribution: +%.3f)", featureValue, contribution)
		}
		return fmt.Sprintf("Business name length suggests lower risk (value: %.2f, contribution: %.3f)", featureValue, contribution)

	default:
		if contribution > 0 {
			return fmt.Sprintf("Feature '%s' contributes positively to risk (value: %.2f, contribution: +%.3f)", featureName, featureValue, contribution)
		}
		return fmt.Sprintf("Feature '%s' contributes negatively to risk (value: %.2f, contribution: %.3f)", featureName, featureValue, contribution)
	}
}

// calculateContributionConfidence calculates confidence for a feature contribution
func (fca *FeatureContributionAnalyzer) calculateContributionConfidence(featureValue, weight float64) float64 {
	// Base confidence from feature weight
	confidence := weight * 2.0 // Scale weight to confidence

	// Adjust based on feature value completeness
	if featureValue > 0 {
		confidence += 0.1
	}

	// Adjust based on feature value range (extreme values are less confident)
	if featureValue < 0.1 || featureValue > 0.9 {
		confidence *= 0.8
	}

	return math.Max(0.1, math.Min(1.0, confidence))
}

// getTopContributors returns the top N contributors by absolute contribution
func (fca *FeatureContributionAnalyzer) getTopContributors(contributions []FeatureContribution, n int) []FeatureContribution {
	// Sort by absolute contribution
	sorted := make([]FeatureContribution, len(contributions))
	copy(sorted, contributions)

	sort.Slice(sorted, func(i, j int) bool {
		return math.Abs(sorted[i].Contribution) > math.Abs(sorted[j].Contribution)
	})

	// Return top N
	if n > len(sorted) {
		n = len(sorted)
	}

	return sorted[:n]
}

// calculateContributionSummary calculates summary statistics for contributions
func (fca *FeatureContributionAnalyzer) calculateContributionSummary(contributions []FeatureContribution) ContributionSummary {
	var positiveTotal, negativeTotal float64
	var uncertainty float64

	for _, contrib := range contributions {
		if contrib.Contribution > 0 {
			positiveTotal += contrib.Contribution
		} else {
			negativeTotal += math.Abs(contrib.Contribution)
		}

		// Uncertainty based on confidence
		uncertainty += (1.0 - contrib.Confidence) * math.Abs(contrib.Contribution)
	}

	netContribution := positiveTotal - negativeTotal
	contributionRatio := 0.0
	if positiveTotal+negativeTotal > 0 {
		contributionRatio = positiveTotal / (positiveTotal + negativeTotal)
	}

	return ContributionSummary{
		PositiveTotal:     positiveTotal,
		NegativeTotal:     negativeTotal,
		NetContribution:   netContribution,
		ContributionRatio: contributionRatio,
		Uncertainty:       uncertainty,
	}
}

// calculateRiskCategoryBreakdown calculates contribution breakdown by risk category
func (fca *FeatureContributionAnalyzer) calculateRiskCategoryBreakdown(contributions []FeatureContribution, business *models.RiskAssessmentRequest) map[string]float64 {
	breakdown := map[string]float64{
		string(models.RiskCategoryFinancial):     0.0,
		string(models.RiskCategoryOperational):   0.0,
		string(models.RiskCategoryCompliance):    0.0,
		string(models.RiskCategoryReputational):  0.0,
		string(models.RiskCategoryRegulatory):    0.0,
		string(models.RiskCategoryGeopolitical):  0.0,
		string(models.RiskCategoryTechnology):    0.0,
		string(models.RiskCategoryEnvironmental): 0.0,
	}

	// Map features to risk categories
	featureCategoryMap := map[string]models.RiskCategory{
		"industry_code":     models.RiskCategoryRegulatory,
		"country_code":      models.RiskCategoryGeopolitical,
		"annual_revenue":    models.RiskCategoryFinancial,
		"employee_count":    models.RiskCategoryFinancial,
		"years_in_business": models.RiskCategoryOperational,
		"has_website":       models.RiskCategoryCompliance,
		"has_email":         models.RiskCategoryCompliance,
		"has_phone":         models.RiskCategoryCompliance,
		"name_length":       models.RiskCategoryCompliance,
	}

	// Aggregate contributions by category
	for _, contrib := range contributions {
		category, exists := featureCategoryMap[contrib.FeatureName]
		if exists {
			breakdown[string(category)] += math.Abs(contrib.Contribution)
		}
	}

	return breakdown
}

// CompareContributions compares contributions between two predictions
func (fca *FeatureContributionAnalyzer) CompareContributions(ctx context.Context, business1, business2 *models.RiskAssessmentRequest, features1, features2 []float64, prediction1, prediction2 float64, featureNames []string) (*ContributionComparison, error) {
	fca.logger.Info("Comparing feature contributions between two predictions")

	// Analyze contributions for both predictions
	analysis1, err := fca.AnalyzeContributions(ctx, business1, features1, prediction1, featureNames)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze contributions for first prediction: %w", err)
	}

	analysis2, err := fca.AnalyzeContributions(ctx, business2, features2, prediction2, featureNames)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze contributions for second prediction: %w", err)
	}

	// Calculate differences for all features
	allContributions1 := append(analysis1.PositiveContributions, analysis1.NegativeContributions...)
	allContributions2 := append(analysis2.PositiveContributions, analysis2.NegativeContributions...)
	contributionDifferences := fca.calculateContributionDifferences(allContributions1, allContributions2)

	comparison := &ContributionComparison{
		Prediction1:             prediction1,
		Prediction2:             prediction2,
		PredictionDifference:    prediction2 - prediction1,
		Analysis1:               analysis1,
		Analysis2:               analysis2,
		ContributionDifferences: contributionDifferences,
		Summary:                 fca.generateComparisonSummary(analysis1, analysis2),
	}

	return comparison, nil
}

// ContributionComparison represents a comparison between two contribution analyses
type ContributionComparison struct {
	Prediction1             float64                  `json:"prediction_1"`
	Prediction2             float64                  `json:"prediction_2"`
	PredictionDifference    float64                  `json:"prediction_difference"`
	Analysis1               *ContributionAnalysis    `json:"analysis_1"`
	Analysis2               *ContributionAnalysis    `json:"analysis_2"`
	ContributionDifferences []ContributionDifference `json:"contribution_differences"`
	Summary                 ComparisonSummary        `json:"summary"`
}

// ContributionDifference represents the difference in contribution for a feature
type ContributionDifference struct {
	FeatureName    string  `json:"feature_name"`
	Contribution1  float64 `json:"contribution_1"`
	Contribution2  float64 `json:"contribution_2"`
	Difference     float64 `json:"difference"`
	RelativeChange float64 `json:"relative_change"`
	Impact         string  `json:"impact"`
}

// ComparisonSummary provides a summary of the comparison
type ComparisonSummary struct {
	KeyDifferences   []string `json:"key_differences"`
	RiskShift        string   `json:"risk_shift"`
	ConfidenceChange float64  `json:"confidence_change"`
	Recommendation   string   `json:"recommendation"`
}

// calculateContributionDifferences calculates differences between contributions
func (fca *FeatureContributionAnalyzer) calculateContributionDifferences(contributions1, contributions2 []FeatureContribution) []ContributionDifference {
	differences := make([]ContributionDifference, 0)

	// Create maps for easy lookup
	contribMap1 := make(map[string]FeatureContribution)
	contribMap2 := make(map[string]FeatureContribution)

	for _, contrib := range contributions1 {
		contribMap1[contrib.FeatureName] = contrib
	}

	for _, contrib := range contributions2 {
		contribMap2[contrib.FeatureName] = contrib
	}

	// Calculate differences for all features
	allFeatures := make(map[string]bool)
	for feature := range contribMap1 {
		allFeatures[feature] = true
	}
	for feature := range contribMap2 {
		allFeatures[feature] = true
	}

	for feature := range allFeatures {
		contrib1, exists1 := contribMap1[feature]
		contrib2, exists2 := contribMap2[feature]

		contribution1 := 0.0
		contribution2 := 0.0

		if exists1 {
			contribution1 = contrib1.Contribution
		}
		if exists2 {
			contribution2 = contrib2.Contribution
		}

		difference := contribution2 - contribution1
		relativeChange := 0.0
		if contribution1 != 0 {
			relativeChange = difference / math.Abs(contribution1)
		}

		impact := "neutral"
		if math.Abs(difference) > 0.05 {
			if difference > 0 {
				impact = "increased_risk"
			} else {
				impact = "decreased_risk"
			}
		}

		diff := ContributionDifference{
			FeatureName:    feature,
			Contribution1:  contribution1,
			Contribution2:  contribution2,
			Difference:     difference,
			RelativeChange: relativeChange,
			Impact:         impact,
		}

		differences = append(differences, diff)
	}

	// Sort by absolute difference
	sort.Slice(differences, func(i, j int) bool {
		return math.Abs(differences[i].Difference) > math.Abs(differences[j].Difference)
	})

	return differences
}

// generateComparisonSummary generates a summary of the comparison
func (fca *FeatureContributionAnalyzer) generateComparisonSummary(analysis1, analysis2 *ContributionAnalysis) ComparisonSummary {
	keyDifferences := make([]string, 0)
	riskShift := "stable"
	confidenceChange := analysis2.ContributionSummary.Uncertainty - analysis1.ContributionSummary.Uncertainty

	// Determine risk shift
	if analysis2.TotalContribution > analysis1.TotalContribution+0.1 {
		riskShift = "increased"
	} else if analysis2.TotalContribution < analysis1.TotalContribution-0.1 {
		riskShift = "decreased"
	}

	// Identify key differences
	if math.Abs(analysis2.ContributionSummary.NetContribution-analysis1.ContributionSummary.NetContribution) > 0.1 {
		keyDifferences = append(keyDifferences, "Significant change in net contribution")
	}

	if math.Abs(analysis2.ContributionSummary.ContributionRatio-analysis1.ContributionSummary.ContributionRatio) > 0.1 {
		keyDifferences = append(keyDifferences, "Change in positive/negative contribution ratio")
	}

	// Generate recommendation
	recommendation := "Continue monitoring current risk factors"
	if riskShift == "increased" {
		recommendation = "Consider implementing additional risk controls"
	} else if riskShift == "decreased" {
		recommendation = "Maintain current risk management practices"
	}

	return ComparisonSummary{
		KeyDifferences:   keyDifferences,
		RiskShift:        riskShift,
		ConfidenceChange: confidenceChange,
		Recommendation:   recommendation,
	}
}
