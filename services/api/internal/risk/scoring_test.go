package risk

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

func TestWeightedScoringAlgorithm_CalculateScore(t *testing.T) {
	algorithm := NewWeightedScoringAlgorithm()

	// Test case 1: Financial risk factors
	factors := []RiskFactor{
		{
			ID:         "financial-stability",
			Name:       "Financial Stability",
			Category:   RiskCategoryFinancial,
			Weight:     0.6,
			Thresholds: map[RiskLevel]float64{RiskLevelLow: 25, RiskLevelMedium: 50, RiskLevelHigh: 75, RiskLevelCritical: 90},
		},
		{
			ID:         "operational-efficiency",
			Name:       "Operational Efficiency",
			Category:   RiskCategoryOperational,
			Weight:     0.4,
			Thresholds: map[RiskLevel]float64{RiskLevelLow: 25, RiskLevelMedium: 50, RiskLevelHigh: 75, RiskLevelCritical: 90},
		},
	}

	data := map[string]interface{}{
		"financial-stability": map[string]interface{}{
			"revenue":       -50000.0,
			"debt_ratio":    0.7,
			"cash_flow":     -25000.0,
			"profit_margin": -0.1,
		},
		"operational-efficiency": map[string]interface{}{
			"employee_turnover":      0.25,
			"operational_efficiency": 0.6,
			"process_maturity":       2.5,
		},
	}

	score, confidence, err := algorithm.CalculateScore(factors, data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if score < 0 || score > 100 {
		t.Errorf("Score should be between 0 and 100, got %f", score)
	}

	if confidence < 0 || confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
	}

	t.Logf("Calculated score: %f, confidence: %f", score, confidence)
}

func TestWeightedScoringAlgorithm_CalculateLevel(t *testing.T) {
	algorithm := NewWeightedScoringAlgorithm()

	testCases := []struct {
		score      float64
		thresholds map[RiskLevel]float64
		expected   RiskLevel
	}{
		{
			score: 15.0,
			thresholds: map[RiskLevel]float64{
				RiskLevelLow: 25.0, RiskLevelMedium: 50.0, RiskLevelHigh: 75.0, RiskLevelCritical: 90.0,
			},
			expected: RiskLevelLow,
		},
		{
			score: 35.0,
			thresholds: map[RiskLevel]float64{
				RiskLevelLow: 25.0, RiskLevelMedium: 50.0, RiskLevelHigh: 75.0, RiskLevelCritical: 90.0,
			},
			expected: RiskLevelLow,
		},
		{
			score: 65.0,
			thresholds: map[RiskLevel]float64{
				RiskLevelLow: 25.0, RiskLevelMedium: 50.0, RiskLevelHigh: 75.0, RiskLevelCritical: 90.0,
			},
			expected: RiskLevelMedium,
		},
		{
			score: 95.0,
			thresholds: map[RiskLevel]float64{
				RiskLevelLow: 25.0, RiskLevelMedium: 50.0, RiskLevelHigh: 75.0, RiskLevelCritical: 90.0,
			},
			expected: RiskLevelCritical,
		},
	}

	for i, tc := range testCases {
		result := algorithm.CalculateLevel(tc.score, tc.thresholds)
		if result != tc.expected {
			t.Errorf("Test case %d: Expected %s, got %s for score %f with thresholds %v", i+1, tc.expected, result, tc.score, tc.thresholds)
		}
	}
}

func TestWeightedScoringAlgorithm_CalculateFinancialScore(t *testing.T) {
	algorithm := NewWeightedScoringAlgorithm()

	testCases := []struct {
		name     string
		data     interface{}
		expected float64
	}{
		{
			name: "High risk financial data",
			data: map[string]interface{}{
				"revenue":       -100000.0,
				"debt_ratio":    0.85,
				"cash_flow":     -50000.0,
				"profit_margin": -0.15,
			},
			expected: 80.0, // Should be high risk
		},
		{
			name: "Low risk financial data",
			data: map[string]interface{}{
				"revenue":       2000000.0,
				"debt_ratio":    0.3,
				"cash_flow":     100000.0,
				"profit_margin": 0.2,
			},
			expected: 20.0, // Should be low risk
		},
		{
			name: "Medium risk financial data",
			data: map[string]interface{}{
				"revenue":       500000.0,
				"debt_ratio":    0.55,
				"cash_flow":     25000.0,
				"profit_margin": 0.08,
			},
			expected: 50.0, // Should be medium risk
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			factor := RiskFactor{
				ID:       "test-factor",
				Category: RiskCategoryFinancial,
			}

			score, confidence, err := algorithm.calculateFinancialScore(factor, tc.data)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if score < 0 || score > 100 {
				t.Errorf("Score should be between 0 and 100, got %f", score)
			}

			if confidence < 0 || confidence > 1 {
				t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
			}

			t.Logf("Score: %f, Confidence: %f", score, confidence)
		})
	}
}

func TestWeightedScoringAlgorithm_CalculateOperationalScore(t *testing.T) {
	algorithm := NewWeightedScoringAlgorithm()

	testCases := []struct {
		name     string
		data     interface{}
		expected float64
	}{
		{
			name: "High risk operational data",
			data: map[string]interface{}{
				"employee_turnover":      0.35,
				"operational_efficiency": 0.4,
				"process_maturity":       1.5,
			},
			expected: 75.0, // Should be high risk
		},
		{
			name: "Low risk operational data",
			data: map[string]interface{}{
				"employee_turnover":      0.05,
				"operational_efficiency": 0.9,
				"process_maturity":       4.5,
			},
			expected: 15.0, // Should be low risk
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			factor := RiskFactor{
				ID:       "test-factor",
				Category: RiskCategoryOperational,
			}

			score, confidence, err := algorithm.calculateOperationalScore(factor, tc.data)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if score < 0 || score > 100 {
				t.Errorf("Score should be between 0 and 100, got %f", score)
			}

			if confidence < 0 || confidence > 1 {
				t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
			}

			t.Logf("Score: %f, Confidence: %f", score, confidence)
		})
	}
}

func TestWeightedScoringAlgorithm_CalculateRegulatoryScore(t *testing.T) {
	algorithm := NewWeightedScoringAlgorithm()

	testCases := []struct {
		name     string
		data     interface{}
		expected float64
	}{
		{
			name: "High risk regulatory data",
			data: map[string]interface{}{
				"compliance_violations": 8.0,
				"regulatory_fines":      150000.0,
				"license_status":        "suspended",
			},
			expected: 90.0, // Should be very high risk
		},
		{
			name: "Low risk regulatory data",
			data: map[string]interface{}{
				"compliance_violations": 0.0,
				"regulatory_fines":      0.0,
				"license_status":        "active",
			},
			expected: 10.0, // Should be low risk
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			factor := RiskFactor{
				ID:       "test-factor",
				Category: RiskCategoryRegulatory,
			}

			score, confidence, err := algorithm.calculateRegulatoryScore(factor, tc.data)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if score < 0 || score > 100 {
				t.Errorf("Score should be between 0 and 100, got %f", score)
			}

			if confidence < 0 || confidence > 1 {
				t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
			}

			t.Logf("Score: %f, Confidence: %f", score, confidence)
		})
	}
}

func TestWeightedScoringAlgorithm_CalculateReputationalScore(t *testing.T) {
	algorithm := NewWeightedScoringAlgorithm()

	testCases := []struct {
		name     string
		data     interface{}
		expected float64
	}{
		{
			name: "High risk reputational data",
			data: map[string]interface{}{
				"customer_satisfaction": 0.2,
				"negative_reviews":      0.6,
				"media_sentiment":       -0.6,
			},
			expected: 80.0, // Should be high risk
		},
		{
			name: "Low risk reputational data",
			data: map[string]interface{}{
				"customer_satisfaction": 0.85,
				"negative_reviews":      0.05,
				"media_sentiment":       0.3,
			},
			expected: 15.0, // Should be low risk
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			factor := RiskFactor{
				ID:       "test-factor",
				Category: RiskCategoryReputational,
			}

			score, confidence, err := algorithm.calculateReputationalScore(factor, tc.data)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if score < 0 || score > 100 {
				t.Errorf("Score should be between 0 and 100, got %f", score)
			}

			if confidence < 0 || confidence > 1 {
				t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
			}

			t.Logf("Score: %f, Confidence: %f", score, confidence)
		})
	}
}

func TestWeightedScoringAlgorithm_CalculateCybersecurityScore(t *testing.T) {
	algorithm := NewWeightedScoringAlgorithm()

	testCases := []struct {
		name     string
		data     interface{}
		expected float64
	}{
		{
			name: "High risk cybersecurity data",
			data: map[string]interface{}{
				"security_incidents": 15.0,
				"data_breaches":      2.0,
				"security_maturity":  1.5,
			},
			expected: 90.0, // Should be very high risk
		},
		{
			name: "Low risk cybersecurity data",
			data: map[string]interface{}{
				"security_incidents": 0.0,
				"data_breaches":      0.0,
				"security_maturity":  4.5,
			},
			expected: 10.0, // Should be low risk
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			factor := RiskFactor{
				ID:       "test-factor",
				Category: RiskCategoryCybersecurity,
			}

			score, confidence, err := algorithm.calculateCybersecurityScore(factor, tc.data)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if score < 0 || score > 100 {
				t.Errorf("Score should be between 0 and 100, got %f", score)
			}

			if confidence < 0 || confidence > 1 {
				t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
			}

			t.Logf("Score: %f, Confidence: %f", score, confidence)
		})
	}
}

func TestWeightedScoringAlgorithm_CalculateConfidence(t *testing.T) {
	algorithm := NewWeightedScoringAlgorithm()

	factors := []RiskFactor{
		{ID: "factor-1", Category: RiskCategoryFinancial},
		{ID: "factor-2", Category: RiskCategoryOperational},
		{ID: "factor-3", Category: RiskCategoryRegulatory},
	}

	// Test with all data available
	data := map[string]interface{}{
		"factor-1": map[string]interface{}{"revenue": 1000000.0},
		"factor-2": map[string]interface{}{"efficiency": 0.8},
		"factor-3": map[string]interface{}{"violations": 0.0},
	}

	confidence := algorithm.CalculateConfidence(factors, data)
	if confidence < 0.7 {
		t.Errorf("Expected high confidence with all data available, got %f", confidence)
	}

	// Test with partial data
	data = map[string]interface{}{
		"factor-1": map[string]interface{}{"revenue": 1000000.0},
		// Missing factor-2 and factor-3
	}

	confidence = algorithm.CalculateConfidence(factors, data)
	if confidence > 0.5 {
		t.Errorf("Expected lower confidence with partial data, got %f", confidence)
	}
}

func TestTrendAnalysisAlgorithm_AnalyzeTrend(t *testing.T) {
	algorithm := NewTrendAnalysisAlgorithm()

	now := time.Now()
	trends := []RiskTrend{
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      30.0,
			Level:      RiskLevelLow,
			RecordedAt: now.Add(-30 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      45.0,
			Level:      RiskLevelMedium,
			RecordedAt: now.Add(-15 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      60.0,
			Level:      RiskLevelHigh,
			RecordedAt: now,
		},
	}

	predictedScore, confidence, err := algorithm.AnalyzeTrend(trends, 30*24*time.Hour)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if predictedScore < 0 || predictedScore > 100 {
		t.Errorf("Predicted score should be between 0 and 100, got %f", predictedScore)
	}

	if confidence < 0 || confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
	}

	t.Logf("Predicted score: %f, Confidence: %f", predictedScore, confidence)
}

func TestCompositeScoringAlgorithm(t *testing.T) {
	weightedAlgorithm := NewWeightedScoringAlgorithm()
	algorithms := []ScoringAlgorithm{weightedAlgorithm}
	weights := []float64{1.0}

	compositeAlgorithm := NewCompositeScoringAlgorithm(algorithms, weights)

	factors := []RiskFactor{
		{
			ID:       "financial-stability",
			Name:     "Financial Stability",
			Category: RiskCategoryFinancial,
			Weight:   1.0,
		},
	}

	data := map[string]interface{}{
		"financial-stability": map[string]interface{}{
			"revenue":       1000000.0,
			"debt_ratio":    0.4,
			"cash_flow":     100000.0,
			"profit_margin": 0.15,
		},
	}

	score, confidence, err := compositeAlgorithm.CalculateScore(factors, data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if score < 0 || score > 100 {
		t.Errorf("Score should be between 0 and 100, got %f", score)
	}

	if confidence < 0 || confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
	}

	level := compositeAlgorithm.CalculateLevel(score, nil)
	if level == "" {
		t.Error("Risk level should not be empty")
	}

	t.Logf("Composite score: %f, confidence: %f, level: %s", score, confidence, level)
}

func TestScoringAlgorithm_EdgeCases(t *testing.T) {
	algorithm := NewWeightedScoringAlgorithm()

	// Test with empty factors
	score, _, err := algorithm.CalculateScore([]RiskFactor{}, map[string]interface{}{})
	if err != nil {
		t.Errorf("Expected no error for empty factors, got %v", err)
	}
	if score != 0.0 {
		t.Errorf("Expected score 0.0 for empty factors, got %f", score)
	}

	// Test with empty data
	factors := []RiskFactor{
		{ID: "test-factor", Category: RiskCategoryFinancial, Weight: 1.0},
	}
	score, _, err = algorithm.CalculateScore(factors, map[string]interface{}{})
	if err != nil {
		t.Errorf("Expected no error for empty data, got %v", err)
	}
	if score != 0.0 {
		t.Errorf("Expected score 0.0 for empty data, got %f", score)
	}

	// Test with invalid data type
	score, _, err = algorithm.CalculateScore(factors, map[string]interface{}{
		"test-factor": "invalid-data-type",
	})
	if err != nil {
		t.Errorf("Expected no error for invalid data type, got %v", err)
	}
	if score != 0.0 {
		t.Errorf("Expected score 0.0 for invalid data type, got %f", score)
	}
}

func TestScoringAlgorithm_Performance(t *testing.T) {
	algorithm := NewWeightedScoringAlgorithm()

	// Create a large number of factors for performance testing
	factors := make([]RiskFactor, 100)
	data := make(map[string]interface{}, 100)

	for i := 0; i < 100; i++ {
		factorID := fmt.Sprintf("factor-%d", i)
		factors[i] = RiskFactor{
			ID:       factorID,
			Name:     fmt.Sprintf("Test Factor %d", i),
			Category: RiskCategoryFinancial,
			Weight:   0.01, // Equal weights
		}
		data[factorID] = map[string]interface{}{
			"revenue":       1000000.0,
			"debt_ratio":    0.4,
			"cash_flow":     100000.0,
			"profit_margin": 0.15,
		}
	}

	start := time.Now()
	score, confidence, err := algorithm.CalculateScore(factors, data)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if score < 0 || score > 100 {
		t.Errorf("Score should be between 0 and 100, got %f", score)
	}

	if confidence < 0 || confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
	}

	// Performance should be reasonable (under 100ms for 100 factors)
	if duration > 100*time.Millisecond {
		t.Errorf("Performance test took too long: %v", duration)
	}

	t.Logf("Performance test: %d factors processed in %v, score: %f, confidence: %f",
		len(factors), duration, score, confidence)
}

func TestRiskPredictionAlgorithm_PredictRiskScore(t *testing.T) {
	algorithm := NewRiskPredictionAlgorithm()

	now := time.Now()
	historicalTrends := []RiskTrend{
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      30.0,
			Level:      RiskLevelLow,
			RecordedAt: now.Add(-90 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      45.0,
			Level:      RiskLevelMedium,
			RecordedAt: now.Add(-60 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      60.0,
			Level:      RiskLevelHigh,
			RecordedAt: now.Add(-30 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      75.0,
			Level:      RiskLevelHigh,
			RecordedAt: now,
		},
	}

	prediction, err := algorithm.PredictRiskScore(historicalTrends, 30*24*time.Hour, 0.7)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if prediction == nil {
		t.Error("Prediction should not be nil")
	}

	if prediction.BusinessID != "business-1" {
		t.Errorf("Expected business ID 'business-1', got %s", prediction.BusinessID)
	}

	if prediction.PredictedScore < 0 || prediction.PredictedScore > 100 {
		t.Errorf("Predicted score should be between 0 and 100, got %f", prediction.PredictedScore)
	}

	if prediction.Confidence < 0 || prediction.Confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", prediction.Confidence)
	}

	if prediction.Horizon != "1month" {
		t.Errorf("Expected horizon '1month', got %s", prediction.Horizon)
	}

	if len(prediction.Factors) == 0 {
		t.Error("Prediction should have contributing factors")
	}

	t.Logf("Prediction: Score=%f, Level=%s, Confidence=%f, Factors=%v",
		prediction.PredictedScore, prediction.PredictedLevel, prediction.Confidence, prediction.Factors)
}

func TestRiskPredictionAlgorithm_PredictRiskScore_InsufficientData(t *testing.T) {
	algorithm := NewRiskPredictionAlgorithm()

	now := time.Now()
	historicalTrends := []RiskTrend{
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      30.0,
			Level:      RiskLevelLow,
			RecordedAt: now.Add(-30 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      45.0,
			Level:      RiskLevelMedium,
			RecordedAt: now,
		},
	}

	_, err := algorithm.PredictRiskScore(historicalTrends, 30*24*time.Hour, 0.7)
	if err == nil {
		t.Error("Expected error for insufficient data")
	}

	if !strings.Contains(err.Error(), "insufficient historical data") {
		t.Errorf("Expected error about insufficient data, got %v", err)
	}
}

func TestRiskPredictionAlgorithm_PredictMultipleHorizons(t *testing.T) {
	algorithm := NewRiskPredictionAlgorithm()

	now := time.Now()
	historicalTrends := []RiskTrend{
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      20.0,
			Level:      RiskLevelLow,
			RecordedAt: now.Add(-120 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      35.0,
			Level:      RiskLevelMedium,
			RecordedAt: now.Add(-90 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      50.0,
			Level:      RiskLevelMedium,
			RecordedAt: now.Add(-60 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      65.0,
			Level:      RiskLevelHigh,
			RecordedAt: now.Add(-30 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      80.0,
			Level:      RiskLevelHigh,
			RecordedAt: now,
		},
	}

	horizons := []time.Duration{
		30 * 24 * time.Hour,  // 1 month
		90 * 24 * time.Hour,  // 3 months
		180 * 24 * time.Hour, // 6 months
	}

	predictions, err := algorithm.PredictMultipleHorizons(historicalTrends, horizons)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(predictions) != 3 {
		t.Errorf("Expected 3 predictions, got %d", len(predictions))
	}

	// Check that predictions are ordered by horizon
	expectedHorizons := []string{"1month", "3months", "6months"}
	for i, prediction := range predictions {
		if prediction.Horizon != expectedHorizons[i] {
			t.Errorf("Expected horizon %s at position %d, got %s", expectedHorizons[i], i, prediction.Horizon)
		}

		// Later predictions should generally have higher scores for this trend
		if i > 0 && prediction.PredictedScore < predictions[i-1].PredictedScore {
			t.Logf("Warning: Later prediction (%s) has lower score than earlier prediction (%s)",
				prediction.Horizon, predictions[i-1].Horizon)
		}
	}

	t.Logf("Multiple horizon predictions: %d predictions generated", len(predictions))
}

func TestRiskPredictionAlgorithm_DeterminePredictedLevel(t *testing.T) {
	algorithm := NewRiskPredictionAlgorithm()

	testCases := []struct {
		score    float64
		expected RiskLevel
	}{
		{10.0, RiskLevelLow},
		{25.0, RiskLevelMedium},
		{50.0, RiskLevelHigh},
		{75.0, RiskLevelCritical},
		{90.0, RiskLevelCritical},
	}

	for _, tc := range testCases {
		level := algorithm.determinePredictedLevel(tc.score)
		if level != tc.expected {
			t.Errorf("For score %f, expected level %s, got %s", tc.score, tc.expected, level)
		}
	}
}

func TestRiskPredictionAlgorithm_IdentifyContributingFactors(t *testing.T) {
	algorithm := NewRiskPredictionAlgorithm()

	now := time.Now()
	trends := []RiskTrend{
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      30.0,
			Level:      RiskLevelLow,
			RecordedAt: now.Add(-60 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      50.0,
			Level:      RiskLevelMedium,
			RecordedAt: now,
		},
	}

	factors := algorithm.identifyContributingFactors(trends, 85.0)

	if len(factors) == 0 {
		t.Error("Should identify contributing factors")
	}

	// Check for expected factors
	foundFinancial := false
	foundHighRisk := false
	for _, factor := range factors {
		if strings.Contains(factor, "financial") {
			foundFinancial = true
		}
		if strings.Contains(factor, "high_risk") {
			foundHighRisk = true
		}
	}

	if !foundFinancial {
		t.Error("Should identify financial category as contributing factor")
	}

	if !foundHighRisk {
		t.Error("Should identify high risk prediction as contributing factor")
	}

	t.Logf("Identified contributing factors: %v", factors)
}

func TestRiskPredictionAlgorithm_FormatHorizon(t *testing.T) {
	algorithm := NewRiskPredictionAlgorithm()

	testCases := []struct {
		horizon  time.Duration
		expected string
	}{
		{15 * 24 * time.Hour, "1month"},
		{30 * 24 * time.Hour, "1month"},
		{60 * 24 * time.Hour, "3months"},
		{90 * 24 * time.Hour, "3months"},
		{150 * 24 * time.Hour, "6months"},
		{180 * 24 * time.Hour, "6months"},
		{300 * 24 * time.Hour, "1year"},
		{365 * 24 * time.Hour, "1year"},
	}

	for _, tc := range testCases {
		result := algorithm.formatHorizon(tc.horizon)
		if result != tc.expected {
			t.Errorf("For horizon %v, expected %s, got %s", tc.horizon, tc.expected, result)
		}
	}
}

func TestRiskPredictionAlgorithm_PredictRiskScoreWithConfidenceInterval(t *testing.T) {
	algorithm := NewRiskPredictionAlgorithm()

	now := time.Now()
	historicalTrends := []RiskTrend{
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      30.0,
			Level:      RiskLevelLow,
			RecordedAt: now.Add(-90 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      45.0,
			Level:      RiskLevelMedium,
			RecordedAt: now.Add(-60 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      60.0,
			Level:      RiskLevelHigh,
			RecordedAt: now.Add(-30 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      75.0,
			Level:      RiskLevelHigh,
			RecordedAt: now,
		},
	}

	prediction, interval, err := algorithm.PredictRiskScoreWithConfidenceInterval(historicalTrends, 30*24*time.Hour, 0.95)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if prediction == nil {
		t.Error("Prediction should not be nil")
	}

	if interval == nil {
		t.Error("Confidence interval should not be nil")
	}

	// Validate confidence interval bounds
	if interval.LowerBound < 0 || interval.LowerBound > 100 {
		t.Errorf("Lower bound should be between 0 and 100, got %f", interval.LowerBound)
	}

	if interval.UpperBound < 0 || interval.UpperBound > 100 {
		t.Errorf("Upper bound should be between 0 and 100, got %f", interval.UpperBound)
	}

	if interval.LowerBound >= interval.UpperBound {
		t.Errorf("Lower bound should be less than upper bound, got %f >= %f", interval.LowerBound, interval.UpperBound)
	}

	if interval.Confidence != 0.95 {
		t.Errorf("Expected confidence level 0.95, got %f", interval.Confidence)
	}

	// Validate that predicted score is within the confidence interval
	if prediction.PredictedScore < interval.LowerBound || prediction.PredictedScore > interval.UpperBound {
		t.Errorf("Predicted score %f should be within confidence interval [%f, %f]",
			prediction.PredictedScore, interval.LowerBound, interval.UpperBound)
	}

	t.Logf("Prediction: Score=%f, Level=%s, Confidence=%f",
		prediction.PredictedScore, prediction.PredictedLevel, prediction.Confidence)
	t.Logf("Confidence Interval: [%f, %f] at %f%% confidence",
		interval.LowerBound, interval.UpperBound, interval.Confidence*100)
}

func TestRiskPredictionAlgorithm_CalculateConfidenceInterval(t *testing.T) {
	algorithm := NewRiskPredictionAlgorithm()

	now := time.Now()
	trends := []RiskTrend{
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      20.0,
			Level:      RiskLevelLow,
			RecordedAt: now.Add(-60 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      40.0,
			Level:      RiskLevelMedium,
			RecordedAt: now.Add(-30 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      60.0,
			Level:      RiskLevelHigh,
			RecordedAt: now,
		},
	}

	interval, err := algorithm.calculateConfidenceInterval(trends, 80.0, 0.95)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if interval == nil {
		t.Error("Confidence interval should not be nil")
	}

	// Test different confidence levels
	testCases := []float64{0.90, 0.95, 0.99}
	for _, confidenceLevel := range testCases {
		interval, err := algorithm.calculateConfidenceInterval(trends, 80.0, confidenceLevel)
		if err != nil {
			t.Errorf("Failed to calculate confidence interval for level %f: %v", confidenceLevel, err)
		}

		if interval.Confidence != confidenceLevel {
			t.Errorf("Expected confidence level %f, got %f", confidenceLevel, interval.Confidence)
		}

		// Higher confidence levels should result in wider intervals
		if confidenceLevel == 0.95 && interval.UpperBound-interval.LowerBound < 5 {
			t.Logf("Warning: Confidence interval for 95%% level seems narrow: [%f, %f]",
				interval.LowerBound, interval.UpperBound)
		}
	}

	t.Logf("Confidence interval: [%f, %f] at %f%% confidence",
		interval.LowerBound, interval.UpperBound, interval.Confidence*100)
}

func TestRiskPredictionAlgorithm_CalculateMean(t *testing.T) {
	algorithm := NewRiskPredictionAlgorithm()

	testCases := []struct {
		values   []float64
		expected float64
	}{
		{[]float64{1, 2, 3, 4, 5}, 3.0},
		{[]float64{10, 20, 30}, 20.0},
		{[]float64{0, 0, 0}, 0.0},
		{[]float64{100}, 100.0},
		{[]float64{}, 0.0},
	}

	for _, tc := range testCases {
		result := algorithm.calculateMean(tc.values)
		if math.Abs(result-tc.expected) > 0.001 {
			t.Errorf("For values %v, expected mean %f, got %f", tc.values, tc.expected, result)
		}
	}
}

func TestRiskPredictionAlgorithm_CalculateStandardDeviation(t *testing.T) {
	algorithm := NewRiskPredictionAlgorithm()

	testCases := []struct {
		values   []float64
		expected float64
	}{
		{[]float64{1, 2, 3, 4, 5}, 1.5811388300841898},
		{[]float64{10, 20, 30}, 10.0},
		{[]float64{0, 0, 0}, 0.0},
		{[]float64{100, 100}, 0.0},
	}

	for _, tc := range testCases {
		mean := algorithm.calculateMean(tc.values)
		result := algorithm.calculateStandardDeviation(tc.values, mean)
		if math.Abs(result-tc.expected) > 0.001 {
			t.Errorf("For values %v, expected std dev %f, got %f", tc.values, tc.expected, result)
		}
	}

	// Test with single value (should return 0)
	result := algorithm.calculateStandardDeviation([]float64{100}, 100)
	if result != 0.0 {
		t.Errorf("For single value, expected std dev 0.0, got %f", result)
	}
}

func TestRiskPredictionAlgorithm_CalculateMarginOfError(t *testing.T) {
	algorithm := NewRiskPredictionAlgorithm()

	testCases := []struct {
		stdDev          float64
		sampleSize      int
		confidenceLevel float64
		expectedRange   [2]float64 // [min, max] expected range
	}{
		{10.0, 10, 0.95, [2]float64{5.0, 8.0}},  // Should be around 6.32
		{5.0, 20, 0.90, [2]float64{1.5, 3.0}},   // Should be around 1.84
		{15.0, 5, 0.99, [2]float64{15.0, 25.0}}, // Should be around 17.28
	}

	for _, tc := range testCases {
		result := algorithm.calculateMarginOfError(tc.stdDev, tc.sampleSize, tc.confidenceLevel)
		if result < tc.expectedRange[0] || result > tc.expectedRange[1] {
			t.Errorf("For stdDev=%f, sampleSize=%d, confidenceLevel=%f, expected margin between %f and %f, got %f",
				tc.stdDev, tc.sampleSize, tc.confidenceLevel, tc.expectedRange[0], tc.expectedRange[1], result)
		}
	}
}

func TestRiskPredictionAlgorithm_ConfidenceInterval_EdgeCases(t *testing.T) {
	algorithm := NewRiskPredictionAlgorithm()

	// Test with insufficient data
	trends := []RiskTrend{
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      30.0,
			Level:      RiskLevelLow,
			RecordedAt: time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			BusinessID: "business-1",
			Category:   RiskCategoryFinancial,
			Score:      45.0,
			Level:      RiskLevelMedium,
			RecordedAt: time.Now(),
		},
	}

	_, err := algorithm.calculateConfidenceInterval(trends, 80.0, 0.95)
	if err == nil {
		t.Error("Expected error for insufficient data")
	}

	if !strings.Contains(err.Error(), "insufficient data") {
		t.Errorf("Expected error about insufficient data, got %v", err)
	}

	// Test with very high predicted score (should cap at 100)
	trends = []RiskTrend{
		{BusinessID: "business-1", Category: RiskCategoryFinancial, Score: 80.0, Level: RiskLevelHigh, RecordedAt: time.Now().Add(-60 * 24 * time.Hour)},
		{BusinessID: "business-1", Category: RiskCategoryFinancial, Score: 85.0, Level: RiskLevelHigh, RecordedAt: time.Now().Add(-30 * 24 * time.Hour)},
		{BusinessID: "business-1", Category: RiskCategoryFinancial, Score: 90.0, Level: RiskLevelCritical, RecordedAt: time.Now()},
	}

	interval, err := algorithm.calculateConfidenceInterval(trends, 95.0, 0.95)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if interval.UpperBound > 100 {
		t.Errorf("Upper bound should be capped at 100, got %f", interval.UpperBound)
	}

	// Test with very low predicted score (should cap at 0)
	trends = []RiskTrend{
		{BusinessID: "business-1", Category: RiskCategoryFinancial, Score: 5.0, Level: RiskLevelLow, RecordedAt: time.Now().Add(-60 * 24 * time.Hour)},
		{BusinessID: "business-1", Category: RiskCategoryFinancial, Score: 10.0, Level: RiskLevelLow, RecordedAt: time.Now().Add(-30 * 24 * time.Hour)},
		{BusinessID: "business-1", Category: RiskCategoryFinancial, Score: 15.0, Level: RiskLevelLow, RecordedAt: time.Now()},
	}

	interval, err = algorithm.calculateConfidenceInterval(trends, 5.0, 0.95)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if interval.LowerBound < 0 {
		t.Errorf("Lower bound should be capped at 0, got %f", interval.LowerBound)
	}
}
