package explainability

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

func TestSHAPExplainer_ExplainPrediction(t *testing.T) {
	logger := zap.NewNop()
	featureNames := []string{
		"industry_code", "country_code", "annual_revenue", "employee_count",
		"years_in_business", "has_website", "has_email", "has_phone",
	}

	explainer := NewSHAPExplainer(featureNames, logger)

	business := &models.RiskAssessmentRequest{
		BusinessName:    "Test Company",
		BusinessAddress: "123 Test St",
		Industry:        "technology",
		Country:         "US",
	}

	features := []float64{0.2, 0.1, 0.8, 0.6, 0.4, 1.0, 1.0, 1.0}
	prediction := 0.65

	featureImportance := map[string]float64{
		"industry_code":     0.25,
		"country_code":      0.20,
		"annual_revenue":    0.15,
		"employee_count":    0.10,
		"years_in_business": 0.12,
		"has_website":       0.08,
		"has_email":         0.05,
		"has_phone":         0.03,
	}

	explanation, err := explainer.ExplainPrediction(context.Background(), business, features, prediction, featureImportance)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if explanation == nil {
		t.Fatal("Expected explanation, got nil")
	}

	if explanation.PredictionValue != prediction {
		t.Errorf("Expected prediction value %f, got %f", prediction, explanation.PredictionValue)
	}

	if len(explanation.FeatureContributions) != len(features) {
		t.Errorf("Expected %d feature contributions, got %d", len(features), len(explanation.FeatureContributions))
	}

	if explanation.Confidence <= 0 || explanation.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", explanation.Confidence)
	}
}

func TestSHAPExplainer_ExplainRiskFactors(t *testing.T) {
	logger := zap.NewNop()
	featureNames := []string{"industry_code", "country_code"}
	explainer := NewSHAPExplainer(featureNames, logger)

	business := &models.RiskAssessmentRequest{
		BusinessName: "Test Company",
		Industry:     "technology",
		Country:      "US",
	}

	riskFactors := []models.RiskFactor{
		{
			Category:    models.RiskCategoryFinancial,
			Name:        "Revenue Risk",
			Score:       0.7,
			Weight:      0.3,
			Description: "High revenue volatility",
			Source:      "test",
			Confidence:  0.8,
		},
		{
			Category:    models.RiskCategoryOperational,
			Name:        "Operational Risk",
			Score:       0.4,
			Weight:      0.2,
			Description: "Stable operations",
			Source:      "test",
			Confidence:  0.9,
		},
	}

	explanations, err := explainer.ExplainRiskFactors(context.Background(), riskFactors, business)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(explanations) != len(riskFactors) {
		t.Errorf("Expected %d explanations, got %d", len(riskFactors), len(explanations))
	}

	// Check that explanations are sorted by importance
	for i := 1; i < len(explanations); i++ {
		prevImportance := explanations[i-1].RiskFactor.Weight * explanations[i-1].RiskFactor.Score
		currImportance := explanations[i].RiskFactor.Weight * explanations[i].RiskFactor.Score
		if prevImportance < currImportance {
			t.Error("Explanations should be sorted by importance (descending)")
		}
	}
}

func TestFeatureContributionAnalyzer_AnalyzeContributions(t *testing.T) {
	logger := zap.NewNop()
	analyzer := NewFeatureContributionAnalyzer(logger)

	business := &models.RiskAssessmentRequest{
		BusinessName: "Test Company",
		Industry:     "technology",
		Country:      "US",
	}

	features := []float64{0.2, 0.1, 0.8, 0.6, 0.4, 1.0, 1.0, 1.0, 0.3}
	featureNames := []string{
		"industry_code", "country_code", "annual_revenue", "employee_count",
		"years_in_business", "has_website", "has_email", "has_phone", "name_length",
	}
	prediction := 0.65

	analysis, err := analyzer.AnalyzeContributions(context.Background(), business, features, prediction, featureNames)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if analysis == nil {
		t.Fatal("Expected analysis, got nil")
	}

	if len(analysis.PositiveContributions)+len(analysis.NegativeContributions) != len(features) {
		t.Errorf("Expected total contributions to equal feature count")
	}

	if analysis.ContributionSummary.PositiveTotal < 0 {
		t.Error("Positive total should be non-negative")
	}

	if analysis.ContributionSummary.NegativeTotal < 0 {
		t.Error("Negative total should be non-negative")
	}

	// Check that top contributors are sorted by absolute contribution
	for i := 1; i < len(analysis.TopContributors); i++ {
		prevAbs := abs(analysis.TopContributors[i-1].Contribution)
		currAbs := abs(analysis.TopContributors[i].Contribution)
		if prevAbs < currAbs {
			t.Error("Top contributors should be sorted by absolute contribution (descending)")
		}
	}
}

func TestFeatureContributionAnalyzer_CompareContributions(t *testing.T) {
	logger := zap.NewNop()
	analyzer := NewFeatureContributionAnalyzer(logger)

	business1 := &models.RiskAssessmentRequest{
		BusinessName: "Test Company 1",
		Industry:     "technology",
		Country:      "US",
	}

	business2 := &models.RiskAssessmentRequest{
		BusinessName: "Test Company 2",
		Industry:     "finance",
		Country:      "US",
	}

	features1 := []float64{0.2, 0.1, 0.8, 0.6, 0.4, 1.0, 1.0, 1.0, 0.3}
	features2 := []float64{0.4, 0.1, 0.6, 0.8, 0.2, 1.0, 1.0, 1.0, 0.5}
	featureNames := []string{
		"industry_code", "country_code", "annual_revenue", "employee_count",
		"years_in_business", "has_website", "has_email", "has_phone", "name_length",
	}
	prediction1 := 0.45
	prediction2 := 0.75

	comparison, err := analyzer.CompareContributions(context.Background(), business1, business2, features1, features2, prediction1, prediction2, featureNames)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if comparison == nil {
		t.Fatal("Expected comparison, got nil")
	}

	if comparison.Prediction1 != prediction1 {
		t.Errorf("Expected prediction1 %f, got %f", prediction1, comparison.Prediction1)
	}

	if comparison.Prediction2 != prediction2 {
		t.Errorf("Expected prediction2 %f, got %f", prediction2, comparison.Prediction2)
	}

	if comparison.PredictionDifference != prediction2-prediction1 {
		t.Errorf("Expected prediction difference %f, got %f", prediction2-prediction1, comparison.PredictionDifference)
	}

	if len(comparison.ContributionDifferences) != len(features1) {
		t.Errorf("Expected %d contribution differences, got %d", len(features1), len(comparison.ContributionDifferences))
	}
}

func TestExplainabilityVisualizer_GenerateWaterfallChart(t *testing.T) {
	logger := zap.NewNop()
	visualizer := NewExplainabilityVisualizer(logger)

	explanation := &SHAPExplanation{
		PredictionValue: 0.65,
		BaseValue:       0.5,
		FeatureContributions: []FeatureContribution{
			{
				FeatureName:  "industry_code",
				FeatureValue: 0.2,
				Contribution: 0.1,
				Importance:   0.25,
				Direction:    "positive",
				Description:  "Technology industry",
				Confidence:   0.8,
			},
			{
				FeatureName:  "country_code",
				FeatureValue: 0.1,
				Contribution: -0.05,
				Importance:   0.20,
				Direction:    "negative",
				Description:  "US country",
				Confidence:   0.9,
			},
		},
		TotalContribution: 0.15,
		Confidence:        0.85,
		ExplanationType:   "shap_like",
	}

	visualization, err := visualizer.GenerateWaterfallChart(context.Background(), explanation)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if visualization == nil {
		t.Fatal("Expected visualization, got nil")
	}

	if visualization.Type != "waterfall_chart" {
		t.Errorf("Expected type 'waterfall_chart', got '%s'", visualization.Type)
	}

	waterfallData, ok := visualization.Data["waterfall"]
	if !ok {
		t.Fatal("Expected waterfall data in visualization")
	}

	waterfall, ok := waterfallData.([]map[string]interface{})
	if !ok {
		t.Fatal("Expected waterfall data to be slice of maps")
	}

	// Should have base value + features + final prediction
	expectedLength := 1 + len(explanation.FeatureContributions) + 1
	if len(waterfall) != expectedLength {
		t.Errorf("Expected %d waterfall items, got %d", expectedLength, len(waterfall))
	}
}

func TestExplainabilityVisualizer_GenerateFeatureImportanceChart(t *testing.T) {
	logger := zap.NewNop()
	visualizer := NewExplainabilityVisualizer(logger)

	analysis := &ContributionAnalysis{
		TopContributors: []FeatureContribution{
			{
				FeatureName:  "industry_code",
				FeatureValue: 0.2,
				Contribution: 0.1,
				Importance:   0.25,
				Direction:    "positive",
				Description:  "Technology industry",
				Confidence:   0.8,
			},
			{
				FeatureName:  "country_code",
				FeatureValue: 0.1,
				Contribution: -0.05,
				Importance:   0.20,
				Direction:    "negative",
				Description:  "US country",
				Confidence:   0.9,
			},
		},
		PositiveContributions: []FeatureContribution{
			{
				FeatureName:  "industry_code",
				FeatureValue: 0.2,
				Contribution: 0.1,
				Importance:   0.25,
				Direction:    "positive",
				Description:  "Technology industry",
				Confidence:   0.8,
			},
		},
		NegativeContributions: []FeatureContribution{
			{
				FeatureName:  "country_code",
				FeatureValue: 0.1,
				Contribution: -0.05,
				Importance:   0.20,
				Direction:    "negative",
				Description:  "US country",
				Confidence:   0.9,
			},
		},
		TotalContribution: 0.05,
		ContributionSummary: ContributionSummary{
			PositiveTotal:     0.1,
			NegativeTotal:     0.05,
			NetContribution:   0.05,
			ContributionRatio: 0.67,
			Uncertainty:       0.1,
		},
		RiskCategoryBreakdown: map[string]float64{
			"financial":   0.1,
			"operational": 0.05,
		},
	}

	visualization, err := visualizer.GenerateFeatureImportanceChart(context.Background(), analysis)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if visualization == nil {
		t.Fatal("Expected visualization, got nil")
	}

	if visualization.Type != "feature_importance" {
		t.Errorf("Expected type 'feature_importance', got '%s'", visualization.Type)
	}
}

func TestExplainabilityVisualizer_GenerateRiskCategoryBreakdown(t *testing.T) {
	logger := zap.NewNop()
	visualizer := NewExplainabilityVisualizer(logger)

	breakdown := map[string]float64{
		"financial":   0.3,
		"operational": 0.2,
		"compliance":  0.1,
		"regulatory":  0.15,
	}

	visualization, err := visualizer.GenerateRiskCategoryBreakdown(context.Background(), breakdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if visualization == nil {
		t.Fatal("Expected visualization, got nil")
	}

	if visualization.Type != "risk_category_breakdown" {
		t.Errorf("Expected type 'risk_category_breakdown', got '%s'", visualization.Type)
	}

	categoriesData, ok := visualization.Data["categories"]
	if !ok {
		t.Fatal("Expected categories data in visualization")
	}

	categories, ok := categoriesData.([]map[string]interface{})
	if !ok {
		t.Fatal("Expected categories data to be slice of maps")
	}

	if len(categories) != len(breakdown) {
		t.Errorf("Expected %d categories, got %d", len(breakdown), len(categories))
	}

	// Check that categories are sorted by contribution
	for i := 1; i < len(categories); i++ {
		prevContrib := categories[i-1]["contribution"].(float64)
		currContrib := categories[i]["contribution"].(float64)
		if prevContrib < currContrib {
			t.Error("Categories should be sorted by contribution (descending)")
		}
	}
}

func TestExplainabilityVisualizer_GenerateTextualExplanation(t *testing.T) {
	logger := zap.NewNop()
	visualizer := NewExplainabilityVisualizer(logger)

	explanation := &SHAPExplanation{
		PredictionValue: 0.65,
		BaseValue:       0.5,
		FeatureContributions: []FeatureContribution{
			{
				FeatureName:  "industry_code",
				FeatureValue: 0.2,
				Contribution: 0.1,
				Importance:   0.25,
				Direction:    "positive",
				Description:  "Technology industry",
				Confidence:   0.8,
			},
		},
		TotalContribution: 0.15,
		Confidence:        0.85,
		ExplanationType:   "shap_like",
	}

	analysis := &ContributionAnalysis{
		TopContributors: explanation.FeatureContributions,
		ContributionSummary: ContributionSummary{
			PositiveTotal:     0.1,
			NegativeTotal:     0.0,
			NetContribution:   0.1,
			ContributionRatio: 1.0,
			Uncertainty:       0.1,
		},
	}

	text, err := visualizer.GenerateTextualExplanation(context.Background(), explanation, analysis)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if text == "" {
		t.Fatal("Expected non-empty textual explanation")
	}

	// Check that text contains key information
	if !contains(text, "Risk Assessment Explanation") {
		t.Error("Expected text to contain header")
	}

	if !contains(text, "0.650") {
		t.Error("Expected text to contain prediction value")
	}

	if !contains(text, "industry_code") {
		t.Error("Expected text to contain feature name")
	}
}

// Helper functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
