package risk

import (
	"fmt"
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
