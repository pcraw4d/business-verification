package risk

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestEnhancedRiskService_Integration(t *testing.T) {
	logger := zap.NewNop()
	factory := NewEnhancedRiskServiceFactory(logger)
	service := factory.CreateEnhancedRiskService()

	t.Run("PerformEnhancedRiskAssessment", func(t *testing.T) {
		request := &EnhancedRiskAssessmentRequest{
			AssessmentID: "integration-test-123",
			BusinessID:   "test-business-456",
			RiskFactorInputs: []RiskFactorInput{
				{
					FactorType: "financial",
					Data: map[string]interface{}{
						"revenue": 1000000.0,
						"debt":    500000.0,
						"assets":  2000000.0,
					},
					Weight: 0.3,
				},
				{
					FactorType: "operational",
					Data: map[string]interface{}{
						"employee_count":    50,
						"years_in_business": 5,
						"location_count":    2,
					},
					Weight: 0.2,
				},
			},
			IncludeTrendAnalysis:       true,
			IncludeCorrelationAnalysis: true,
			TimeRange: &TimeRange{
				StartTime: time.Now().AddDate(0, -6, 0),
				EndTime:   time.Now(),
				Duration:  "6 months",
			},
		}

		response, err := service.PerformEnhancedRiskAssessment(context.Background(), request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, request.AssessmentID, response.AssessmentID)
		assert.Equal(t, request.BusinessID, response.BusinessID)
		assert.NotZero(t, response.OverallRiskScore)
		assert.NotEmpty(t, response.OverallRiskLevel)
		assert.NotEmpty(t, response.RiskFactors)
		assert.NotZero(t, response.ConfidenceScore)
		assert.NotZero(t, response.ProcessingTimeMs)
		assert.NotNil(t, response.Metadata)
	})

	t.Run("GetActiveAlerts", func(t *testing.T) {
		alerts, err := service.GetActiveAlerts(context.Background(), "test-business-456")

		require.NoError(t, err)
		assert.NotNil(t, alerts)
		// For now, we expect empty alerts since we're using mock data
		assert.Empty(t, alerts)
	})

	t.Run("AcknowledgeAlert", func(t *testing.T) {
		err := service.AcknowledgeAlert(
			context.Background(),
			"test-alert-123",
			"test-user-456",
			"Alert acknowledged for testing",
		)

		require.NoError(t, err)
	})

	t.Run("ResolveAlert", func(t *testing.T) {
		err := service.ResolveAlert(
			context.Background(),
			"test-alert-123",
			"test-user-456",
			"Alert resolved - risk mitigated",
		)

		require.NoError(t, err)
	})
}

func TestEnhancedRiskFactorCalculator_Integration(t *testing.T) {
	logger := zap.NewNop()
	calculator := NewEnhancedRiskFactorCalculator(logger)

	t.Run("CalculateFinancialRiskFactor", func(t *testing.T) {
		input := RiskFactorInput{
			FactorType: "financial",
			Data: map[string]interface{}{
				"revenue":   1000000.0,
				"debt":      500000.0,
				"assets":    2000000.0,
				"cash_flow": 200000.0,
			},
			Weight: 0.3,
		}

		result, err := calculator.CalculateFactor(context.Background(), input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "financial", result.FactorType)
		assert.GreaterOrEqual(t, result.Score, 0.0)
		assert.LessOrEqual(t, result.Score, 1.0)
		assert.NotEmpty(t, result.RiskLevel)
		assert.GreaterOrEqual(t, result.Confidence, 0.0)
		assert.LessOrEqual(t, result.Confidence, 1.0)
		assert.NotEmpty(t, result.Description)
	})

	t.Run("CalculateOperationalRiskFactor", func(t *testing.T) {
		input := RiskFactorInput{
			FactorType: "operational",
			Data: map[string]interface{}{
				"employee_count":     50,
				"years_in_business":  5,
				"location_count":     2,
				"process_automation": 0.7,
			},
			Weight: 0.2,
		}

		result, err := calculator.CalculateFactor(context.Background(), input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "operational", result.FactorType)
		assert.GreaterOrEqual(t, result.Score, 0.0)
		assert.LessOrEqual(t, result.Score, 1.0)
		assert.NotEmpty(t, result.RiskLevel)
		assert.GreaterOrEqual(t, result.Confidence, 0.0)
		assert.LessOrEqual(t, result.Confidence, 1.0)
		assert.NotEmpty(t, result.Description)
	})

	t.Run("CalculateRegulatoryRiskFactor", func(t *testing.T) {
		input := RiskFactorInput{
			FactorType: "regulatory",
			Data: map[string]interface{}{
				"compliance_score":      0.8,
				"regulatory_violations": 0,
				"license_status":        "active",
				"audit_frequency":       "annual",
			},
			Weight: 0.25,
		}

		result, err := calculator.CalculateFactor(context.Background(), input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "regulatory", result.FactorType)
		assert.GreaterOrEqual(t, result.Score, 0.0)
		assert.LessOrEqual(t, result.Score, 1.0)
		assert.NotEmpty(t, result.RiskLevel)
		assert.GreaterOrEqual(t, result.Confidence, 0.0)
		assert.LessOrEqual(t, result.Confidence, 1.0)
		assert.NotEmpty(t, result.Description)
	})

	t.Run("CalculateReputationalRiskFactor", func(t *testing.T) {
		input := RiskFactorInput{
			FactorType: "reputational",
			Data: map[string]interface{}{
				"customer_satisfaction":  0.85,
				"online_reviews_score":   4.2,
				"social_media_sentiment": 0.7,
				"brand_recognition":      0.6,
			},
			Weight: 0.15,
		}

		result, err := calculator.CalculateFactor(context.Background(), input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "reputational", result.FactorType)
		assert.GreaterOrEqual(t, result.Score, 0.0)
		assert.LessOrEqual(t, result.Score, 1.0)
		assert.NotEmpty(t, result.RiskLevel)
		assert.GreaterOrEqual(t, result.Confidence, 0.0)
		assert.LessOrEqual(t, result.Confidence, 1.0)
		assert.NotEmpty(t, result.Description)
	})

	t.Run("CalculateCybersecurityRiskFactor", func(t *testing.T) {
		input := RiskFactorInput{
			FactorType: "cybersecurity",
			Data: map[string]interface{}{
				"security_score":    0.75,
				"incident_count":    2,
				"patch_frequency":   "monthly",
				"employee_training": 0.8,
			},
			Weight: 0.1,
		}

		result, err := calculator.CalculateFactor(context.Background(), input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "cybersecurity", result.FactorType)
		assert.GreaterOrEqual(t, result.Score, 0.0)
		assert.LessOrEqual(t, result.Score, 1.0)
		assert.NotEmpty(t, result.RiskLevel)
		assert.GreaterOrEqual(t, result.Confidence, 0.0)
		assert.LessOrEqual(t, result.Confidence, 1.0)
		assert.NotEmpty(t, result.Description)
	})
}

func TestRecommendationEngine_Integration(t *testing.T) {
	logger := zap.NewNop()
	engine := NewRecommendationEngine(logger)

	t.Run("GenerateRecommendations", func(t *testing.T) {
		factors := []RiskFactorDetail{
			{
				FactorType:  "financial",
				Score:       0.75,
				RiskLevel:   RiskLevelHigh,
				Confidence:  0.85,
				Weight:      0.3,
				Description: "High financial risk due to debt ratio",
				LastUpdated: time.Now(),
			},
			{
				FactorType:  "operational",
				Score:       0.45,
				RiskLevel:   RiskLevelMedium,
				Confidence:  0.75,
				Weight:      0.2,
				Description: "Moderate operational risk",
				LastUpdated: time.Now(),
			},
		}

		recommendations, err := engine.GenerateRecommendations(context.Background(), factors)

		require.NoError(t, err)
		assert.NotNil(t, recommendations)
		assert.NotEmpty(t, recommendations)

		// Check that recommendations are properly structured
		for _, rec := range recommendations {
			assert.NotEmpty(t, rec.ID)
			assert.NotEmpty(t, rec.Title)
			assert.NotEmpty(t, rec.Description)
			assert.NotEmpty(t, rec.Priority)
			assert.NotEmpty(t, rec.Category)
			assert.NotEmpty(t, rec.Impact)
			assert.NotEmpty(t, rec.Effort)
			assert.NotEmpty(t, rec.Timeline)
			assert.NotZero(t, rec.CreatedAt)
			assert.NotZero(t, rec.UpdatedAt)
		}
	})
}

func TestTrendAnalysisService_Integration(t *testing.T) {
	logger := zap.NewNop()
	service := NewTrendAnalysisService(logger)

	t.Run("GetRiskTrends", func(t *testing.T) {
		historicalData := []RiskHistoryEntry{
			{
				ID:           "history-1",
				BusinessID:   "test-business-456",
				AssessmentID: "assessment-1",
				Timestamp:    time.Now().AddDate(0, -6, 0),
				RiskScore:    0.80,
				RiskLevel:    RiskLevelHigh,
				FactorScores: map[string]float64{
					"financial":   0.80,
					"operational": 0.60,
				},
				Confidence: 0.75,
			},
			{
				ID:           "history-2",
				BusinessID:   "test-business-456",
				AssessmentID: "assessment-2",
				Timestamp:    time.Now().AddDate(0, -3, 0),
				RiskScore:    0.70,
				RiskLevel:    RiskLevelHigh,
				FactorScores: map[string]float64{
					"financial":   0.70,
					"operational": 0.55,
				},
				Confidence: 0.80,
			},
			{
				ID:           "history-3",
				BusinessID:   "test-business-456",
				AssessmentID: "assessment-3",
				Timestamp:    time.Now(),
				RiskScore:    0.60,
				RiskLevel:    RiskLevelMedium,
				FactorScores: map[string]float64{
					"financial":   0.60,
					"operational": 0.50,
				},
				Confidence: 0.85,
			},
		}

		trends, err := service.GetRiskTrends(context.Background(), historicalData)

		require.NoError(t, err)
		assert.NotNil(t, trends)
		assert.NotEmpty(t, trends)

		// Check that trends are properly structured
		for _, trend := range trends {
			assert.NotEmpty(t, trend.FactorType)
			assert.NotEmpty(t, trend.Direction)
			assert.GreaterOrEqual(t, trend.Magnitude, 0.0)
			assert.GreaterOrEqual(t, trend.Confidence, 0.0)
			assert.LessOrEqual(t, trend.Confidence, 1.0)
			assert.NotEmpty(t, trend.Timeframe)
			assert.NotEmpty(t, trend.Description)
			assert.NotEmpty(t, trend.DataPoints)
		}
	})
}

func TestCorrelationAnalyzer_Integration(t *testing.T) {
	logger := zap.NewNop()
	analyzer := NewCorrelationAnalyzer(logger)

	t.Run("AnalyzeCorrelation", func(t *testing.T) {
		factorData := [][]float64{
			{0.75, 0.70, 0.65}, // financial
			{0.60, 0.55, 0.50}, // operational
			{0.45, 0.40, 0.35}, // regulatory
		}
		factorNames := []string{"financial", "operational", "regulatory"}

		correlations, err := analyzer.AnalyzeCorrelation(context.Background(), factorData, factorNames)

		require.NoError(t, err)
		assert.NotNil(t, correlations)
		assert.NotEmpty(t, correlations)

		// Check that correlations are properly calculated
		for key, correlation := range correlations {
			assert.NotEmpty(t, key)
			assert.GreaterOrEqual(t, correlation, -1.0)
			assert.LessOrEqual(t, correlation, 1.0)
		}
	})
}

func TestConfidenceCalibrator_Integration(t *testing.T) {
	logger := zap.NewNop()
	calibrator := NewConfidenceCalibrator(logger)

	t.Run("CalibrateConfidence", func(t *testing.T) {
		historicalData := []HistoricalDataPoint{
			{
				Timestamp:     time.Now().AddDate(0, -3, 0),
				Value:         0.75,
				ActualOutcome: 0.80,
				Confidence:    0.85,
			},
			{
				Timestamp:     time.Now().AddDate(0, -2, 0),
				Value:         0.70,
				ActualOutcome: 0.75,
				Confidence:    0.80,
			},
			{
				Timestamp:     time.Now().AddDate(0, -1, 0),
				Value:         0.65,
				ActualOutcome: 0.70,
				Confidence:    0.75,
			},
		}

		calibration, err := calibrator.CalibrateConfidence("financial", 0.80, historicalData)

		require.NoError(t, err)
		assert.NotNil(t, calibration)
		assert.Equal(t, "financial", calibration.FactorID)
		assert.GreaterOrEqual(t, calibration.CalibratedConfidence, 0.0)
		assert.LessOrEqual(t, calibration.CalibratedConfidence, 1.0)
		assert.GreaterOrEqual(t, calibration.Confidence, 0.0)
		assert.LessOrEqual(t, calibration.Confidence, 1.0)
		assert.NotEmpty(t, calibration.CalibrationMethod)
		assert.NotZero(t, calibration.CalibratedAt)
	})
}

func TestRiskAlertSystem_Integration(t *testing.T) {
	logger := zap.NewNop()
	factory := NewEnhancedRiskServiceFactory(logger)
	alertSystem := factory.CreateAlertSystem()

	t.Run("CheckAndTriggerAlerts", func(t *testing.T) {
		factors := []RiskFactorDetail{
			{
				FactorType:  "financial",
				Score:       0.85,
				RiskLevel:   RiskLevelCritical,
				Confidence:  0.90,
				Weight:      0.3,
				Description: "Critical financial risk",
				LastUpdated: time.Now(),
			},
			{
				FactorType:  "operational",
				Score:       0.45,
				RiskLevel:   RiskLevelMedium,
				Confidence:  0.75,
				Weight:      0.2,
				Description: "Moderate operational risk",
				LastUpdated: time.Now(),
			},
		}

		alerts, err := alertSystem.CheckAndTriggerAlerts(context.Background(), factors)

		require.NoError(t, err)
		assert.NotNil(t, alerts)

		// Should have at least one alert for the high-risk financial factor
		assert.NotEmpty(t, alerts)

		// Check that alerts are properly structured
		for _, alert := range alerts {
			assert.NotEmpty(t, alert.ID)
			assert.NotEmpty(t, alert.AlertType)
			assert.NotEmpty(t, alert.Severity)
			assert.NotEmpty(t, alert.Title)
			assert.NotEmpty(t, alert.Description)
			assert.NotEmpty(t, alert.RiskFactor)
			assert.GreaterOrEqual(t, alert.Threshold, 0.0)
			assert.GreaterOrEqual(t, alert.CurrentValue, 0.0)
			assert.NotEmpty(t, alert.Status)
			assert.NotZero(t, alert.CreatedAt)
		}
	})
}
