package explainability

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// ExplainabilityVisualizer provides visualization capabilities for explainability results
type ExplainabilityVisualizer struct {
	logger *zap.Logger
}

// NewExplainabilityVisualizer creates a new explainability visualizer
func NewExplainabilityVisualizer(logger *zap.Logger) *ExplainabilityVisualizer {
	return &ExplainabilityVisualizer{
		logger: logger,
	}
}

// VisualizationData represents data for visualization
type VisualizationData struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
	Config      map[string]interface{} `json:"config"`
}

// GenerateWaterfallChart generates waterfall chart data for SHAP explanations
func (ev *ExplainabilityVisualizer) GenerateWaterfallChart(ctx context.Context, explanation *SHAPExplanation) (*VisualizationData, error) {
	ev.logger.Info("Generating waterfall chart for SHAP explanation")

	// Sort contributions by absolute value for better visualization
	sortedContributions := make([]FeatureContribution, len(explanation.FeatureContributions))
	copy(sortedContributions, explanation.FeatureContributions)

	sort.Slice(sortedContributions, func(i, j int) bool {
		return math.Abs(sortedContributions[i].Contribution) > math.Abs(sortedContributions[j].Contribution)
	})

	// Build waterfall data
	waterfallData := make([]map[string]interface{}, 0, len(sortedContributions)+2)

	// Start with base value
	waterfallData = append(waterfallData, map[string]interface{}{
		"name":  "Base Value",
		"value": explanation.BaseValue,
		"type":  "base",
		"color": "#888888",
	})

	// Add feature contributions
	cumulativeValue := explanation.BaseValue
	for _, contrib := range sortedContributions {
		cumulativeValue += contrib.Contribution

		color := "#ff6b6b" // Red for positive
		if contrib.Contribution < 0 {
			color = "#4ecdc4" // Teal for negative
		}

		waterfallData = append(waterfallData, map[string]interface{}{
			"name":        contrib.FeatureName,
			"value":       contrib.Contribution,
			"cumulative":  cumulativeValue,
			"type":        "feature",
			"color":       color,
			"description": contrib.Description,
			"confidence":  contrib.Confidence,
			"importance":  contrib.Importance,
		})
	}

	// Add final prediction
	waterfallData = append(waterfallData, map[string]interface{}{
		"name":  "Final Prediction",
		"value": explanation.PredictionValue,
		"type":  "final",
		"color": "#2c3e50",
	})

	visualizationData := &VisualizationData{
		Type:        "waterfall_chart",
		Title:       "SHAP Waterfall Chart",
		Description: "Feature contributions to risk prediction",
		Data: map[string]interface{}{
			"waterfall": waterfallData,
			"summary": map[string]interface{}{
				"base_value":         explanation.BaseValue,
				"prediction_value":   explanation.PredictionValue,
				"total_contribution": explanation.TotalContribution,
				"confidence":         explanation.Confidence,
			},
		},
		Config: map[string]interface{}{
			"chart_type": "waterfall",
			"x_axis":     "Features",
			"y_axis":     "Contribution",
			"colors": map[string]string{
				"positive": "#ff6b6b",
				"negative": "#4ecdc4",
				"base":     "#888888",
				"final":    "#2c3e50",
			},
		},
	}

	return visualizationData, nil
}

// GenerateFeatureImportanceChart generates feature importance chart data
func (ev *ExplainabilityVisualizer) GenerateFeatureImportanceChart(ctx context.Context, analysis *ContributionAnalysis) (*VisualizationData, error) {
	ev.logger.Info("Generating feature importance chart")

	// Prepare data for horizontal bar chart
	importanceData := make([]map[string]interface{}, 0, len(analysis.TopContributors))

	for _, contrib := range analysis.TopContributors {
		importanceData = append(importanceData, map[string]interface{}{
			"feature_name": contrib.FeatureName,
			"importance":   contrib.Importance,
			"contribution": contrib.Contribution,
			"confidence":   contrib.Confidence,
			"direction":    contrib.Direction,
			"description":  contrib.Description,
		})
	}

	visualizationData := &VisualizationData{
		Type:        "feature_importance",
		Title:       "Feature Importance Chart",
		Description: "Most important features contributing to risk prediction",
		Data: map[string]interface{}{
			"features": importanceData,
			"summary": map[string]interface{}{
				"total_features":     len(analysis.TopContributors),
				"positive_features":  len(analysis.PositiveContributions),
				"negative_features":  len(analysis.NegativeContributions),
				"total_contribution": analysis.TotalContribution,
			},
		},
		Config: map[string]interface{}{
			"chart_type": "horizontal_bar",
			"x_axis":     "Importance",
			"y_axis":     "Features",
			"sort_by":    "importance",
		},
	}

	return visualizationData, nil
}

// GenerateRiskCategoryBreakdown generates risk category breakdown visualization
func (ev *ExplainabilityVisualizer) GenerateRiskCategoryBreakdown(ctx context.Context, breakdown map[string]float64) (*VisualizationData, error) {
	ev.logger.Info("Generating risk category breakdown")

	// Convert to slice for visualization
	categoryData := make([]map[string]interface{}, 0, len(breakdown))
	totalContribution := 0.0

	for category, contribution := range breakdown {
		totalContribution += contribution
		categoryData = append(categoryData, map[string]interface{}{
			"category":     category,
			"contribution": contribution,
			"percentage":   0.0, // Will be calculated below
		})
	}

	// Calculate percentages
	for i := range categoryData {
		if totalContribution > 0 {
			categoryData[i]["percentage"] = (categoryData[i]["contribution"].(float64) / totalContribution) * 100
		}
	}

	// Sort by contribution
	sort.Slice(categoryData, func(i, j int) bool {
		return categoryData[i]["contribution"].(float64) > categoryData[j]["contribution"].(float64)
	})

	visualizationData := &VisualizationData{
		Type:        "risk_category_breakdown",
		Title:       "Risk Category Breakdown",
		Description: "Contribution breakdown by risk category",
		Data: map[string]interface{}{
			"categories": categoryData,
			"summary": map[string]interface{}{
				"total_contribution": totalContribution,
				"category_count":     len(breakdown),
			},
		},
		Config: map[string]interface{}{
			"chart_type":       "pie_chart",
			"show_percentages": true,
			"colors": []string{
				"#ff6b6b", "#4ecdc4", "#45b7d1", "#96ceb4",
				"#feca57", "#ff9ff3", "#54a0ff", "#5f27cd",
			},
		},
	}

	return visualizationData, nil
}

// GenerateContributionComparison generates comparison visualization
func (ev *ExplainabilityVisualizer) GenerateContributionComparison(ctx context.Context, comparison *ContributionComparison) (*VisualizationData, error) {
	ev.logger.Info("Generating contribution comparison")

	// Prepare comparison data
	comparisonData := make([]map[string]interface{}, 0, len(comparison.ContributionDifferences))

	for _, diff := range comparison.ContributionDifferences {
		comparisonData = append(comparisonData, map[string]interface{}{
			"feature_name":    diff.FeatureName,
			"contribution_1":  diff.Contribution1,
			"contribution_2":  diff.Contribution2,
			"difference":      diff.Difference,
			"relative_change": diff.RelativeChange,
			"impact":          diff.Impact,
		})
	}

	visualizationData := &VisualizationData{
		Type:        "contribution_comparison",
		Title:       "Feature Contribution Comparison",
		Description: "Comparison of feature contributions between two predictions",
		Data: map[string]interface{}{
			"comparison": comparisonData,
			"summary": map[string]interface{}{
				"prediction_1":          comparison.Prediction1,
				"prediction_2":          comparison.Prediction2,
				"prediction_difference": comparison.PredictionDifference,
				"risk_shift":            comparison.Summary.RiskShift,
				"confidence_change":     comparison.Summary.ConfidenceChange,
			},
		},
		Config: map[string]interface{}{
			"chart_type": "grouped_bar",
			"x_axis":     "Features",
			"y_axis":     "Contribution",
			"series":     []string{"Prediction 1", "Prediction 2", "Difference"},
		},
	}

	return visualizationData, nil
}

// GenerateRiskFactorExplanation generates visualization for risk factor explanations
func (ev *ExplainabilityVisualizer) GenerateRiskFactorExplanation(ctx context.Context, explanations []RiskFactorExplanation) (*VisualizationData, error) {
	ev.logger.Info("Generating risk factor explanation visualization")

	// Prepare risk factor data
	riskFactorData := make([]map[string]interface{}, 0, len(explanations))

	for _, explanation := range explanations {
		riskFactorData = append(riskFactorData, map[string]interface{}{
			"category":       explanation.RiskFactor.Category,
			"name":           explanation.RiskFactor.Name,
			"score":          explanation.RiskFactor.Score,
			"weight":         explanation.RiskFactor.Weight,
			"impact":         explanation.Impact,
			"confidence":     explanation.Confidence,
			"explanation":    explanation.Explanation,
			"recommendation": explanation.Recommendation,
		})
	}

	// Group by category
	categoryGroups := make(map[string][]map[string]interface{})
	for _, factor := range riskFactorData {
		category := factor["category"].(models.RiskCategory)
		categoryGroups[string(category)] = append(categoryGroups[string(category)], factor)
	}

	visualizationData := &VisualizationData{
		Type:        "risk_factor_explanation",
		Title:       "Risk Factor Explanations",
		Description: "Detailed explanations for each risk factor",
		Data: map[string]interface{}{
			"risk_factors": riskFactorData,
			"categories":   categoryGroups,
			"summary": map[string]interface{}{
				"total_factors": len(explanations),
				"categories":    len(categoryGroups),
			},
		},
		Config: map[string]interface{}{
			"chart_type":   "accordion",
			"group_by":     "category",
			"show_details": true,
		},
	}

	return visualizationData, nil
}

// GenerateSummaryReport generates a comprehensive summary report
func (ev *ExplainabilityVisualizer) GenerateSummaryReport(ctx context.Context, explanation *SHAPExplanation, analysis *ContributionAnalysis, riskFactors []RiskFactorExplanation) (*VisualizationData, error) {
	ev.logger.Info("Generating comprehensive summary report")

	// Calculate summary statistics
	totalFeatures := len(explanation.FeatureContributions)
	positiveFeatures := 0
	negativeFeatures := 0
	highConfidenceFeatures := 0

	for _, contrib := range explanation.FeatureContributions {
		if contrib.Contribution > 0 {
			positiveFeatures++
		} else {
			negativeFeatures++
		}
		if contrib.Confidence > 0.8 {
			highConfidenceFeatures++
		}
	}

	// Risk level interpretation
	riskLevel := "Low"
	if explanation.PredictionValue > 0.8 {
		riskLevel = "Critical"
	} else if explanation.PredictionValue > 0.6 {
		riskLevel = "High"
	} else if explanation.PredictionValue > 0.4 {
		riskLevel = "Medium"
	}

	// Key insights
	keyInsights := ev.generateKeyInsights(explanation, analysis, riskFactors)

	// Recommendations
	recommendations := ev.generateRecommendations(explanation, analysis, riskFactors)

	summaryData := &VisualizationData{
		Type:        "summary_report",
		Title:       "Risk Assessment Summary Report",
		Description: "Comprehensive summary of risk assessment and explanations",
		Data: map[string]interface{}{
			"prediction": map[string]interface{}{
				"value":      explanation.PredictionValue,
				"level":      riskLevel,
				"confidence": explanation.Confidence,
				"base_value": explanation.BaseValue,
			},
			"features": map[string]interface{}{
				"total":            totalFeatures,
				"positive":         positiveFeatures,
				"negative":         negativeFeatures,
				"high_confidence":  highConfidenceFeatures,
				"confidence_ratio": float64(highConfidenceFeatures) / float64(totalFeatures),
			},
			"contributions": map[string]interface{}{
				"total":              analysis.TotalContribution,
				"positive_total":     analysis.ContributionSummary.PositiveTotal,
				"negative_total":     analysis.ContributionSummary.NegativeTotal,
				"net_contribution":   analysis.ContributionSummary.NetContribution,
				"contribution_ratio": analysis.ContributionSummary.ContributionRatio,
			},
			"risk_factors": map[string]interface{}{
				"total":      len(riskFactors),
				"categories": ev.countRiskCategories(riskFactors),
			},
			"key_insights":    keyInsights,
			"recommendations": recommendations,
		},
		Config: map[string]interface{}{
			"report_type": "comprehensive",
			"sections": []string{
				"prediction_summary",
				"feature_analysis",
				"contribution_breakdown",
				"risk_factors",
				"insights",
				"recommendations",
			},
		},
	}

	return summaryData, nil
}

// generateKeyInsights generates key insights from the analysis
func (ev *ExplainabilityVisualizer) generateKeyInsights(explanation *SHAPExplanation, analysis *ContributionAnalysis, riskFactors []RiskFactorExplanation) []string {
	insights := make([]string, 0)

	// Prediction insights
	if explanation.PredictionValue > 0.7 {
		insights = append(insights, "High risk prediction indicates need for immediate attention")
	} else if explanation.PredictionValue < 0.3 {
		insights = append(insights, "Low risk prediction suggests stable business profile")
	}

	// Feature insights
	if len(analysis.PositiveContributions) > len(analysis.NegativeContributions) {
		insights = append(insights, "More features contribute positively to risk than negatively")
	}

	// Confidence insights
	if explanation.Confidence > 0.8 {
		insights = append(insights, "High confidence in prediction due to complete feature data")
	} else if explanation.Confidence < 0.6 {
		insights = append(insights, "Lower confidence suggests need for additional data")
	}

	// Top contributor insights
	if len(analysis.TopContributors) > 0 {
		topContrib := analysis.TopContributors[0]
		insights = append(insights, fmt.Sprintf("Primary risk driver: %s (contribution: %.3f)", topContrib.FeatureName, topContrib.Contribution))
	}

	// Risk factor insights
	highRiskFactors := 0
	for _, factor := range riskFactors {
		if factor.RiskFactor.Score > 0.7 {
			highRiskFactors++
		}
	}

	if highRiskFactors > 0 {
		insights = append(insights, fmt.Sprintf("%d risk factors show high risk levels", highRiskFactors))
	}

	return insights
}

// generateRecommendations generates recommendations based on the analysis
func (ev *ExplainabilityVisualizer) generateRecommendations(explanation *SHAPExplanation, analysis *ContributionAnalysis, riskFactors []RiskFactorExplanation) []string {
	recommendations := make([]string, 0)

	// General recommendations based on risk level
	if explanation.PredictionValue > 0.7 {
		recommendations = append(recommendations, "Implement immediate risk mitigation measures")
		recommendations = append(recommendations, "Consider additional monitoring and controls")
	} else if explanation.PredictionValue > 0.4 {
		recommendations = append(recommendations, "Monitor risk factors closely")
		recommendations = append(recommendations, "Maintain current risk management practices")
	} else {
		recommendations = append(recommendations, "Continue current risk management approach")
		recommendations = append(recommendations, "Regular risk assessment reviews recommended")
	}

	// Feature-specific recommendations
	for _, contrib := range analysis.TopContributors {
		if contrib.Contribution > 0.1 {
			recommendations = append(recommendations, fmt.Sprintf("Address high-risk feature: %s", contrib.FeatureName))
		}
	}

	// Risk factor recommendations
	for _, factor := range riskFactors {
		if factor.RiskFactor.Score > 0.7 {
			recommendations = append(recommendations, fmt.Sprintf("Focus on %s risk management", factor.RiskFactor.Category))
		}
	}

	// Data quality recommendations
	if explanation.Confidence < 0.7 {
		recommendations = append(recommendations, "Improve data completeness for better risk assessment")
	}

	return recommendations
}

// countRiskCategories counts the number of unique risk categories
func (ev *ExplainabilityVisualizer) countRiskCategories(riskFactors []RiskFactorExplanation) map[string]int {
	categoryCount := make(map[string]int)

	for _, factor := range riskFactors {
		category := string(factor.RiskFactor.Category)
		categoryCount[category]++
	}

	return categoryCount
}

// GenerateTextualExplanation generates a human-readable textual explanation
func (ev *ExplainabilityVisualizer) GenerateTextualExplanation(ctx context.Context, explanation *SHAPExplanation, analysis *ContributionAnalysis) (string, error) {
	ev.logger.Info("Generating textual explanation")

	var text strings.Builder

	// Header
	text.WriteString("Risk Assessment Explanation\n")
	text.WriteString("==========================\n\n")

	// Prediction summary
	text.WriteString(fmt.Sprintf("Overall Risk Score: %.3f\n", explanation.PredictionValue))
	text.WriteString(fmt.Sprintf("Confidence Level: %.1f%%\n", explanation.Confidence*100))
	text.WriteString(fmt.Sprintf("Base Risk Level: %.3f\n\n", explanation.BaseValue))

	// Key contributors
	text.WriteString("Key Risk Contributors:\n")
	text.WriteString("---------------------\n")

	for i, contrib := range analysis.TopContributors {
		if i >= 5 { // Limit to top 5
			break
		}

		direction := "increases"
		if contrib.Contribution < 0 {
			direction = "decreases"
		}

		text.WriteString(fmt.Sprintf("%d. %s: %.3f (%.1f%% confidence) - %s risk\n",
			i+1, contrib.FeatureName, contrib.Contribution, contrib.Confidence*100, direction))
		text.WriteString(fmt.Sprintf("   %s\n\n", contrib.Description))
	}

	// Summary
	text.WriteString("Summary:\n")
	text.WriteString("--------\n")
	text.WriteString(fmt.Sprintf("Total contribution: %.3f\n", analysis.TotalContribution))
	text.WriteString(fmt.Sprintf("Positive contributions: %.3f\n", analysis.ContributionSummary.PositiveTotal))
	text.WriteString(fmt.Sprintf("Negative contributions: %.3f\n", analysis.ContributionSummary.NegativeTotal))
	text.WriteString(fmt.Sprintf("Net contribution: %.3f\n", analysis.ContributionSummary.NetContribution))

	return text.String(), nil
}
