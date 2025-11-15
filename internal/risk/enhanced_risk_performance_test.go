package risk

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

func BenchmarkEnhancedRiskService_PerformEnhancedRiskAssessment(b *testing.B) {
	logger := zap.NewNop()
	factory := NewEnhancedRiskServiceFactory(logger)
	service := factory.CreateEnhancedRiskService()

	request := &EnhancedRiskAssessmentRequest{
		AssessmentID: "benchmark-test",
		BusinessID:   "benchmark-business",
		RiskFactorInputs: []RiskFactorInput{
			{
				FactorID: "financial",
				Data: map[string]interface{}{
					"revenue":   1000000.0,
					"debt":      500000.0,
					"assets":    2000000.0,
					"cash_flow": 200000.0,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			{
				FactorID: "operational",
				Data: map[string]interface{}{
					"employee_count":     50,
					"years_in_business":  5,
					"location_count":     2,
					"process_automation": 0.7,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			{
				FactorID: "regulatory",
				Data: map[string]interface{}{
					"compliance_score":      0.8,
					"regulatory_violations": 0,
					"license_status":        "active",
					"audit_frequency":       "annual",
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			{
				FactorID: "reputational",
				Data: map[string]interface{}{
					"customer_satisfaction":  0.85,
					"online_reviews_score":   4.2,
					"social_media_sentiment": 0.7,
					"brand_recognition":      0.6,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			{
				FactorID: "cybersecurity",
				Data: map[string]interface{}{
					"security_score":    0.75,
					"incident_count":    2,
					"patch_frequency":   "monthly",
					"employee_training": 0.8,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.PerformEnhancedRiskAssessment(context.Background(), request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEnhancedRiskFactorCalculator_CalculateFactor(b *testing.B) {
	logger := zap.NewNop()
	calculator := NewEnhancedRiskFactorCalculator(logger)

	inputs := []RiskFactorInput{
		{
			FactorID: "financial",
			Data: map[string]interface{}{
				"revenue":   1000000.0,
				"debt":      500000.0,
				"assets":    2000000.0,
				"cash_flow": 200000.0,
			},
			Timestamp:   time.Now(),
			Source:      "test",
			Reliability: 0.9,
		},
		{
			FactorID: "operational",
			Data: map[string]interface{}{
				"employee_count":     50,
				"years_in_business":  5,
				"location_count":     2,
				"process_automation": 0.7,
			},
			Timestamp:   time.Now(),
			Source:      "test",
			Reliability: 0.9,
		},
		{
			FactorID: "regulatory",
			Data: map[string]interface{}{
				"compliance_score":      0.8,
				"regulatory_violations": 0,
				"license_status":        "active",
				"audit_frequency":       "annual",
			},
			Timestamp:   time.Now(),
			Source:      "test",
			Reliability: 0.9,
		},
		{
			FactorID: "reputational",
			Data: map[string]interface{}{
				"customer_satisfaction":  0.85,
				"online_reviews_score":   4.2,
				"social_media_sentiment": 0.7,
				"brand_recognition":      0.6,
			},
			Timestamp:   time.Now(),
			Source:      "test",
			Reliability: 0.9,
		},
		{
			FactorID: "cybersecurity",
			Data: map[string]interface{}{
				"security_score":    0.75,
				"incident_count":    2,
				"patch_frequency":   "monthly",
				"employee_training": 0.8,
			},
			Timestamp:   time.Now(),
			Source:      "test",
			Reliability: 0.9,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		_, err := calculator.CalculateFactor(context.Background(), input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRecommendationEngine_GenerateRecommendations(b *testing.B) {
	logger := zap.NewNop()
	engine := NewRecommendationEngine(logger)

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
		{
			FactorType:  "regulatory",
			Score:       0.30,
			RiskLevel:   RiskLevelLow,
			Confidence:  0.90,
			Weight:      0.25,
			Description: "Low regulatory risk",
			LastUpdated: time.Now(),
		},
		{
			FactorType:  "reputational",
			Score:       0.60,
			RiskLevel:   RiskLevelMedium,
			Confidence:  0.80,
			Weight:      0.15,
			Description: "Moderate reputational risk",
			LastUpdated: time.Now(),
		},
		{
			FactorType:  "cybersecurity",
			Score:       0.40,
			RiskLevel:   RiskLevelLow,
			Confidence:  0.85,
			Weight:      0.1,
			Description: "Low cybersecurity risk",
			LastUpdated: time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.GenerateRecommendations(context.Background(), factors)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTrendAnalysisService_GetRiskTrends(b *testing.B) {
	logger := zap.NewNop()
	service := NewTrendAnalysisService(logger)

	historicalData := []RiskHistoryEntry{
		{
			ID:           "history-1",
			BusinessID:   "test-business-456",
			AssessmentID: "assessment-1",
			Timestamp:    time.Now().AddDate(0, -12, 0),
			RiskScore:    0.80,
			RiskLevel:    RiskLevelHigh,
			FactorScores: map[string]float64{
				"financial":     0.80,
				"operational":   0.60,
				"regulatory":    0.30,
				"reputational":  0.60,
				"cybersecurity": 0.40,
			},
			Confidence: 0.75,
		},
		{
			ID:           "history-2",
			BusinessID:   "test-business-456",
			AssessmentID: "assessment-2",
			Timestamp:    time.Now().AddDate(0, -9, 0),
			RiskScore:    0.75,
			RiskLevel:    RiskLevelHigh,
			FactorScores: map[string]float64{
				"financial":     0.75,
				"operational":   0.55,
				"regulatory":    0.25,
				"reputational":  0.55,
				"cybersecurity": 0.35,
			},
			Confidence: 0.80,
		},
		{
			ID:           "history-3",
			BusinessID:   "test-business-456",
			AssessmentID: "assessment-3",
			Timestamp:    time.Now().AddDate(0, -6, 0),
			RiskScore:    0.70,
			RiskLevel:    RiskLevelHigh,
			FactorScores: map[string]float64{
				"financial":     0.70,
				"operational":   0.50,
				"regulatory":    0.20,
				"reputational":  0.50,
				"cybersecurity": 0.30,
			},
			Confidence: 0.85,
		},
		{
			ID:           "history-4",
			BusinessID:   "test-business-456",
			AssessmentID: "assessment-4",
			Timestamp:    time.Now().AddDate(0, -3, 0),
			RiskScore:    0.65,
			RiskLevel:    RiskLevelMedium,
			FactorScores: map[string]float64{
				"financial":     0.65,
				"operational":   0.45,
				"regulatory":    0.15,
				"reputational":  0.45,
				"cybersecurity": 0.25,
			},
			Confidence: 0.90,
		},
		{
			ID:           "history-5",
			BusinessID:   "test-business-456",
			AssessmentID: "assessment-5",
			Timestamp:    time.Now(),
			RiskScore:    0.60,
			RiskLevel:    RiskLevelMedium,
			FactorScores: map[string]float64{
				"financial":     0.60,
				"operational":   0.40,
				"regulatory":    0.10,
				"reputational":  0.40,
				"cybersecurity": 0.20,
			},
			Confidence: 0.95,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetRiskTrends(context.Background(), historicalData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCorrelationAnalyzer_AnalyzeCorrelation(b *testing.B) {
	logger := zap.NewNop()
	analyzer := NewCorrelationAnalyzer(logger)

	factorData := [][]float64{
		{0.75, 0.70, 0.65, 0.60, 0.55}, // financial
		{0.60, 0.55, 0.50, 0.45, 0.40}, // operational
		{0.45, 0.40, 0.35, 0.30, 0.25}, // regulatory
		{0.60, 0.55, 0.50, 0.45, 0.40}, // reputational
		{0.40, 0.35, 0.30, 0.25, 0.20}, // cybersecurity
	}
	factorNames := []string{"financial", "operational", "regulatory", "reputational", "cybersecurity"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeCorrelation(context.Background(), factorData, factorNames)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConfidenceCalibrator_CalibrateConfidence(b *testing.B) {
	logger := zap.NewNop()
	calibrator := NewConfidenceCalibrator(logger)

	historicalData := []HistoricalDataPoint{
		{
			Timestamp:     time.Now().AddDate(0, -12, 0),
			Value:         0.80,
			ActualOutcome: 0.85,
			Confidence:    0.75,
		},
		{
			Timestamp:     time.Now().AddDate(0, -9, 0),
			Value:         0.75,
			ActualOutcome: 0.80,
			Confidence:    0.80,
		},
		{
			Timestamp:     time.Now().AddDate(0, -6, 0),
			Value:         0.70,
			ActualOutcome: 0.75,
			Confidence:    0.85,
		},
		{
			Timestamp:     time.Now().AddDate(0, -3, 0),
			Value:         0.65,
			ActualOutcome: 0.70,
			Confidence:    0.90,
		},
		{
			Timestamp:     time.Now(),
			Value:         0.60,
			ActualOutcome: 0.65,
			Confidence:    0.95,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := calibrator.CalibrateConfidence("financial", 0.80, historicalData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRiskAlertSystem_CheckAndTriggerAlerts(b *testing.B) {
	logger := zap.NewNop()
	factory := NewEnhancedRiskServiceFactory(logger)
	alertSystem := factory.CreateAlertSystem()

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
		{
			FactorType:  "regulatory",
			Score:       0.30,
			RiskLevel:   RiskLevelLow,
			Confidence:  0.90,
			Weight:      0.25,
			Description: "Low regulatory risk",
			LastUpdated: time.Now(),
		},
		{
			FactorType:  "reputational",
			Score:       0.60,
			RiskLevel:   RiskLevelMedium,
			Confidence:  0.80,
			Weight:      0.15,
			Description: "Moderate reputational risk",
			LastUpdated: time.Now(),
		},
		{
			FactorType:  "cybersecurity",
			Score:       0.40,
			RiskLevel:   RiskLevelLow,
			Confidence:  0.85,
			Weight:      0.1,
			Description: "Low cybersecurity risk",
			LastUpdated: time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := alertSystem.CheckAndTriggerAlerts(context.Background(), factors)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test concurrent access to ensure thread safety
func TestEnhancedRiskService_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	factory := NewEnhancedRiskServiceFactory(logger)
	service := factory.CreateEnhancedRiskService()

	request := &EnhancedRiskAssessmentRequest{
		AssessmentID: "concurrent-test",
		BusinessID:   "concurrent-business",
		RiskFactorInputs: []RiskFactorInput{
			{
				FactorID: "financial",
				Data: map[string]interface{}{
					"revenue": 1000000.0,
					"debt":    500000.0,
					"assets":  2000000.0,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
		},
		IncludeTrendAnalysis:       true,
		IncludeCorrelationAnalysis: true,
	}

	// Run multiple goroutines concurrently
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Create a unique request for each goroutine
			req := *request
			req.AssessmentID = fmt.Sprintf("concurrent-test-%d", id)
			req.BusinessID = fmt.Sprintf("concurrent-business-%d", id)

			_, err := service.PerformEnhancedRiskAssessment(context.Background(), &req)
			if err != nil {
				t.Errorf("Goroutine %d failed: %v", id, err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// Test memory usage and garbage collection
func TestEnhancedRiskService_MemoryUsage(t *testing.T) {
	logger := zap.NewNop()
	factory := NewEnhancedRiskServiceFactory(logger)
	service := factory.CreateEnhancedRiskService()

	request := &EnhancedRiskAssessmentRequest{
		AssessmentID: "memory-test",
		BusinessID:   "memory-business",
		RiskFactorInputs: []RiskFactorInput{
			{
				FactorID: "financial",
				Data: map[string]interface{}{
					"revenue":   1000000.0,
					"debt":      500000.0,
					"assets":    2000000.0,
					"cash_flow": 200000.0,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
			{
				FactorID: "operational",
				Data: map[string]interface{}{
					"employee_count":     50,
					"years_in_business":  5,
					"location_count":     2,
					"process_automation": 0.7,
				},
				Timestamp:   time.Now(),
				Source:      "test",
				Reliability: 0.9,
			},
		},
		IncludeTrendAnalysis:       true,
		IncludeCorrelationAnalysis: true,
	}

	// Run multiple assessments to test memory usage
	numAssessments := 100
	for i := 0; i < numAssessments; i++ {
		req := *request
		req.AssessmentID = fmt.Sprintf("memory-test-%d", i)
		req.BusinessID = fmt.Sprintf("memory-business-%d", i)

		_, err := service.PerformEnhancedRiskAssessment(context.Background(), &req)
		if err != nil {
			t.Errorf("Assessment %d failed: %v", i, err)
		}
	}
}
