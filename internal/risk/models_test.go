package risk

import (
	"testing"
	"time"
)

func TestRiskCategoryConstants(t *testing.T) {
	expectedCategories := []RiskCategory{
		RiskCategoryOperational,
		RiskCategoryFinancial,
		RiskCategoryRegulatory,
		RiskCategoryReputational,
		RiskCategoryCybersecurity,
	}

	for _, category := range expectedCategories {
		if category == "" {
			t.Errorf("Risk category should not be empty")
		}
	}
}

func TestRiskLevelConstants(t *testing.T) {
	expectedLevels := []RiskLevel{
		RiskLevelLow,
		RiskLevelMedium,
		RiskLevelHigh,
		RiskLevelCritical,
	}

	for _, level := range expectedLevels {
		if level == "" {
			t.Errorf("Risk level should not be empty")
		}
	}
}

func TestRiskFactorCreation(t *testing.T) {
	now := time.Now()
	thresholds := map[RiskLevel]float64{
		RiskLevelLow:      25.0,
		RiskLevelMedium:   50.0,
		RiskLevelHigh:     75.0,
		RiskLevelCritical: 90.0,
	}

	factor := RiskFactor{
		ID:          "test-factor-1",
		Name:        "Financial Stability",
		Description: "Assessment of business financial stability",
		Category:    RiskCategoryFinancial,
		Weight:      0.3,
		Thresholds:  thresholds,
		Metadata: map[string]interface{}{
			"source": "financial_data_provider",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if factor.ID != "test-factor-1" {
		t.Errorf("Expected ID 'test-factor-1', got '%s'", factor.ID)
	}

	if factor.Category != RiskCategoryFinancial {
		t.Errorf("Expected category 'financial', got '%s'", factor.Category)
	}

	if factor.Weight < 0.0 || factor.Weight > 1.0 {
		t.Errorf("Weight should be between 0.0 and 1.0, got %f", factor.Weight)
	}

	if len(factor.Thresholds) != 4 {
		t.Errorf("Expected 4 thresholds, got %d", len(factor.Thresholds))
	}
}

func TestRiskScoreCalculation(t *testing.T) {
	now := time.Now()
	score := RiskScore{
		FactorID:     "financial-stability",
		FactorName:   "Financial Stability",
		Category:     RiskCategoryFinancial,
		Score:        65.5,
		Level:        RiskLevelHigh,
		Confidence:   0.85,
		Explanation:  "Business shows signs of financial stress",
		Evidence:     []string{"Declining revenue", "High debt ratio"},
		CalculatedAt: now,
	}

	if score.Score < 0.0 || score.Score > 100.0 {
		t.Errorf("Score should be between 0.0 and 100.0, got %f", score.Score)
	}

	if score.Confidence < 0.0 || score.Confidence > 1.0 {
		t.Errorf("Confidence should be between 0.0 and 1.0, got %f", score.Confidence)
	}

	if len(score.Evidence) == 0 {
		t.Error("Evidence should not be empty")
	}
}

func TestRiskAssessmentCreation(t *testing.T) {
	now := time.Now()
	validUntil := now.Add(24 * time.Hour)

	assessment := RiskAssessment{
		ID:           "assessment-1",
		BusinessID:   "business-123",
		BusinessName: "Test Company",
		OverallScore: 72.5,
		OverallLevel: RiskLevelHigh,
		CategoryScores: map[RiskCategory]RiskScore{
			RiskCategoryFinancial: {
				FactorID:   "financial-overall",
				FactorName: "Financial Risk",
				Category:   RiskCategoryFinancial,
				Score:      75.0,
				Level:      RiskLevelHigh,
				Confidence: 0.9,
			},
		},
		FactorScores: []RiskScore{
			{
				FactorID:   "cash-flow",
				FactorName: "Cash Flow",
				Category:   RiskCategoryFinancial,
				Score:      80.0,
				Level:      RiskLevelHigh,
				Confidence: 0.85,
			},
		},
		Recommendations: []RiskRecommendation{
			{
				ID:          "rec-1",
				RiskFactor:  "cash-flow",
				Title:       "Improve Cash Flow",
				Description: "Implement better cash flow management",
				Priority:    RiskLevelHigh,
				Action:      "Review payment terms",
				Impact:      "Reduce financial risk",
				Timeline:    "30 days",
			},
		},
		AlertLevel: RiskLevelHigh,
		AssessedAt: now,
		ValidUntil: validUntil,
	}

	if assessment.OverallScore < 0.0 || assessment.OverallScore > 100.0 {
		t.Errorf("Overall score should be between 0.0 and 100.0, got %f", assessment.OverallScore)
	}

	if len(assessment.CategoryScores) == 0 {
		t.Error("Category scores should not be empty")
	}

	if len(assessment.FactorScores) == 0 {
		t.Error("Factor scores should not be empty")
	}

	if len(assessment.Recommendations) == 0 {
		t.Error("Recommendations should not be empty")
	}
}

func TestRiskThresholdValidation(t *testing.T) {
	now := time.Now()
	threshold := RiskThreshold{
		Category:    RiskCategoryFinancial,
		LowMax:      25.0,
		MediumMax:   50.0,
		HighMax:     75.0,
		CriticalMin: 90.0,
		UpdatedAt:   now,
	}

	// Validate threshold progression
	if threshold.LowMax >= threshold.MediumMax {
		t.Error("LowMax should be less than MediumMax")
	}

	if threshold.MediumMax >= threshold.HighMax {
		t.Error("MediumMax should be less than HighMax")
	}

	if threshold.HighMax >= threshold.CriticalMin {
		t.Error("HighMax should be less than CriticalMin")
	}
}

func TestRiskAlertCreation(t *testing.T) {
	now := time.Now()
	alert := RiskAlert{
		ID:           "alert-1",
		BusinessID:   "business-123",
		RiskFactor:   "financial-stability",
		Level:        RiskLevelHigh,
		Message:      "Financial stability score exceeded threshold",
		Score:        85.0,
		Threshold:    75.0,
		TriggeredAt:  now,
		Acknowledged: false,
	}

	if alert.Score <= alert.Threshold {
		t.Error("Alert score should be greater than threshold")
	}

	if alert.Acknowledged {
		t.Error("New alert should not be acknowledged")
	}
}

func TestRiskDataValidation(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	data := RiskData{
		ID:         "data-1",
		BusinessID: "business-123",
		Source:     "financial_api",
		DataType:   "financial_metrics",
		Data: map[string]interface{}{
			"revenue":    1000000.0,
			"debt_ratio": 0.6,
			"cash_flow":  -50000.0,
		},
		Reliability: 0.9,
		CollectedAt: now,
		ExpiresAt:   &expiresAt,
	}

	if data.Reliability < 0.0 || data.Reliability > 1.0 {
		t.Errorf("Reliability should be between 0.0 and 1.0, got %f", data.Reliability)
	}

	if len(data.Data) == 0 {
		t.Error("Data should not be empty")
	}
}

func TestRiskAssessmentRequestValidation(t *testing.T) {
	request := RiskAssessmentRequest{
		BusinessID:         "business-123",
		BusinessName:       "Test Company",
		Categories:         []RiskCategory{RiskCategoryFinancial, RiskCategoryOperational},
		Factors:            []string{"cash-flow", "operational-efficiency"},
		IncludeHistory:     true,
		IncludePredictions: true,
		Metadata: map[string]interface{}{
			"assessment_type": "comprehensive",
		},
	}

	if request.BusinessID == "" {
		t.Error("BusinessID should not be empty")
	}

	if request.BusinessName == "" {
		t.Error("BusinessName should not be empty")
	}

	if len(request.Categories) == 0 {
		t.Error("Categories should not be empty")
	}
}

func TestRiskAssessmentResponseCreation(t *testing.T) {
	now := time.Now()
	assessment := &RiskAssessment{
		ID:           "assessment-1",
		BusinessID:   "business-123",
		BusinessName: "Test Company",
		OverallScore: 72.5,
		OverallLevel: RiskLevelHigh,
		AssessedAt:   now,
	}

	response := RiskAssessmentResponse{
		Assessment: assessment,
		Trends: []RiskTrend{
			{
				BusinessID:   "business-123",
				Category:     RiskCategoryFinancial,
				Score:        72.5,
				Level:        RiskLevelHigh,
				RecordedAt:   now,
				ChangeFrom:   65.0,
				ChangePeriod: "1month",
			},
		},
		Predictions: []RiskPrediction{
			{
				ID:             "pred-1",
				BusinessID:     "business-123",
				FactorID:       "financial-stability",
				PredictedScore: 78.0,
				PredictedLevel: RiskLevelHigh,
				Confidence:     0.8,
				Horizon:        "3months",
				PredictedAt:    now,
				Factors:        []string{"market_conditions", "debt_increase"},
			},
		},
		Alerts: []RiskAlert{
			{
				ID:          "alert-1",
				BusinessID:  "business-123",
				RiskFactor:  "financial-stability",
				Level:       RiskLevelHigh,
				Message:     "Financial risk threshold exceeded",
				Score:       85.0,
				Threshold:   75.0,
				TriggeredAt: now,
			},
		},
		GeneratedAt: now,
	}

	if response.Assessment == nil {
		t.Error("Assessment should not be nil")
	}

	if len(response.Trends) == 0 {
		t.Error("Trends should not be empty")
	}

	if len(response.Predictions) == 0 {
		t.Error("Predictions should not be empty")
	}

	if len(response.Alerts) == 0 {
		t.Error("Alerts should not be empty")
	}
}
