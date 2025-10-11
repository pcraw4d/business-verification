package validation

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

// MockModelValidator is a mock implementation of ModelValidator for testing
type MockModelValidator struct {
	name string
}

func (m *MockModelValidator) Train(ctx context.Context, features [][]float64, labels []float64) error {
	// Mock training - just sleep to simulate work
	time.Sleep(10 * time.Millisecond)
	return nil
}

func (m *MockModelValidator) Predict(ctx context.Context, features [][]float64) ([]float64, error) {
	// Mock prediction - return random values
	predictions := make([]float64, len(features))
	for i := range predictions {
		// Simulate some correlation with features
		if len(features[i]) > 0 {
			predictions[i] = features[i][0] + 0.1
		} else {
			predictions[i] = 0.5
		}
	}
	return predictions, nil
}

func (m *MockModelValidator) PredictProba(ctx context.Context, features [][]float64) ([][]float64, error) {
	// Mock probability prediction
	probabilities := make([][]float64, len(features))
	for i := range probabilities {
		probabilities[i] = []float64{0.3, 0.7} // Mock probabilities
	}
	return probabilities, nil
}

func (m *MockModelValidator) GetName() string {
	return m.name
}

func TestCrossValidator_CrossValidate(t *testing.T) {
	logger := zap.NewNop()
	cv := NewCrossValidator(logger)
	model := &MockModelValidator{name: "test-model"}

	// Generate test samples
	samples := generateTestSamples(100)

	// Test cross-validation
	ctx := context.Background()
	result, err := cv.CrossValidate(ctx, model, samples, 5, 0.95)

	if err != nil {
		t.Fatalf("Cross-validation failed: %v", err)
	}

	// Validate results
	if result.ModelName != "test-model" {
		t.Errorf("Expected model name 'test-model', got '%s'", result.ModelName)
	}

	if result.K != 5 {
		t.Errorf("Expected 5 folds, got %d", result.K)
	}

	if result.TotalSamples != 100 {
		t.Errorf("Expected 100 samples, got %d", result.TotalSamples)
	}

	if len(result.FoldResults) != 5 {
		t.Errorf("Expected 5 fold results, got %d", len(result.FoldResults))
	}

	// Check that metrics are reasonable
	if result.OverallMetrics.MeanAccuracy < 0 || result.OverallMetrics.MeanAccuracy > 1 {
		t.Errorf("Mean accuracy should be between 0 and 1, got %f", result.OverallMetrics.MeanAccuracy)
	}

	if result.OverallMetrics.MeanF1Score < 0 || result.OverallMetrics.MeanF1Score > 1 {
		t.Errorf("Mean F1 score should be between 0 and 1, got %f", result.OverallMetrics.MeanF1Score)
	}

	if result.OverallMetrics.MeanAUC < 0 || result.OverallMetrics.MeanAUC > 1 {
		t.Errorf("Mean AUC should be between 0 and 1, got %f", result.OverallMetrics.MeanAUC)
	}

	// Check confidence intervals
	if result.ConfidenceInterval.Confidence != 0.95 {
		t.Errorf("Expected confidence level 0.95, got %f", result.ConfidenceInterval.Confidence)
	}

	if result.ConfidenceInterval.Accuracy.Lower > result.ConfidenceInterval.Accuracy.Upper {
		t.Errorf("Confidence interval lower bound should be <= upper bound")
	}
}

func TestHistoricalDataGenerator_GenerateHistoricalData(t *testing.T) {
	logger := zap.NewNop()
	hdg := NewHistoricalDataGenerator(logger)

	config := DataGenerationConfig{
		TotalSamples:     50,
		TimeRange:        30 * 24 * time.Hour, // 30 days
		RiskCategories:   getTestRiskCategories(),
		IndustryWeights:  getTestIndustryWeights(),
		GeographicBias:   getTestGeographicBias(),
		SeasonalPatterns: true,
		TrendStrength:    0.01,
		NoiseLevel:       0.05,
	}

	ctx := context.Background()
	samples, historicalSamples, err := hdg.GenerateHistoricalData(ctx, config)

	if err != nil {
		t.Fatalf("Historical data generation failed: %v", err)
	}

	if len(samples) != 50 {
		t.Errorf("Expected 50 samples, got %d", len(samples))
	}

	if len(historicalSamples) != 50 {
		t.Errorf("Expected 50 historical samples, got %d", len(historicalSamples))
	}

	// Check that samples have features
	for i, sample := range samples {
		if len(sample.Features) == 0 {
			t.Errorf("Sample %d has no features", i)
		}
		if sample.Label < 0 || sample.Label > 1 {
			t.Errorf("Sample %d label should be between 0 and 1, got %f", i, sample.Label)
		}
	}

	// Check that historical samples have metadata
	for i, sample := range historicalSamples {
		if sample.BusinessID == "" {
			t.Errorf("Historical sample %d has no business ID", i)
		}
		if sample.Industry == "" {
			t.Errorf("Historical sample %d has no industry", i)
		}
		if sample.Country == "" {
			t.Errorf("Historical sample %d has no country", i)
		}
		if len(sample.RiskFactors) == 0 {
			t.Errorf("Historical sample %d has no risk factors", i)
		}
	}
}

func TestValidationService_ValidateModel(t *testing.T) {
	logger := zap.NewNop()
	vs := NewValidationService(logger)
	model := &MockModelValidator{name: "test-model"}

	config := ValidationConfig{
		CrossValidation: CrossValidationConfig{
			KFolds:          3,
			ConfidenceLevel: 0.95,
			RandomSeed:      12345,
			ParallelFolds:   false,
			MaxConcurrency:  2,
		},
		DataGeneration: DataGenerationConfig{
			TotalSamples:     30,
			TimeRange:        7 * 24 * time.Hour, // 7 days
			RiskCategories:   getTestRiskCategories(),
			IndustryWeights:  getTestIndustryWeights(),
			GeographicBias:   getTestGeographicBias(),
			SeasonalPatterns: false,
			TrendStrength:    0.01,
			NoiseLevel:       0.05,
		},
		OutputFormat: "json",
		SaveResults:  false,
		ResultsPath:  "",
	}

	ctx := context.Background()
	report, err := vs.ValidateModel(ctx, model, config)

	if err != nil {
		t.Fatalf("Model validation failed: %v", err)
	}

	// Validate report structure
	if report.Summary.OverallScore < 0 || report.Summary.OverallScore > 1 {
		t.Errorf("Overall score should be between 0 and 1, got %f", report.Summary.OverallScore)
	}

	if report.Summary.AccuracyScore < 0 || report.Summary.AccuracyScore > 1 {
		t.Errorf("Accuracy score should be between 0 and 1, got %f", report.Summary.AccuracyScore)
	}

	if report.Summary.ReliabilityScore < 0 || report.Summary.ReliabilityScore > 1 {
		t.Errorf("Reliability score should be between 0 and 1, got %f", report.Summary.ReliabilityScore)
	}

	if report.Summary.PerformanceScore < 0 || report.Summary.PerformanceScore > 1 {
		t.Errorf("Performance score should be between 0 and 1, got %f", report.Summary.PerformanceScore)
	}

	if report.Summary.Recommendation == "" {
		t.Error("Recommendation should not be empty")
	}

	if report.Summary.RiskLevel == "" {
		t.Error("Risk level should not be empty")
	}

	// Check cross-validation results
	if report.CrossValidation == nil {
		t.Error("Cross-validation results should not be nil")
	} else {
		if report.CrossValidation.ModelName != "test-model" {
			t.Errorf("Expected model name 'test-model', got '%s'", report.CrossValidation.ModelName)
		}
	}

	// Check historical data summary
	if report.HistoricalData.TotalSamples != 30 {
		t.Errorf("Expected 30 historical samples, got %d", report.HistoricalData.TotalSamples)
	}

	if report.HistoricalData.DataQuality.OverallQuality < 0 || report.HistoricalData.DataQuality.OverallQuality > 1 {
		t.Errorf("Data quality should be between 0 and 1, got %f", report.HistoricalData.DataQuality.OverallQuality)
	}

	// Check model comparison
	if len(report.ModelComparison) == 0 {
		t.Error("Model comparison should not be empty")
	}

	// Check recommendations
	if len(report.Recommendations) == 0 {
		t.Error("Recommendations should not be empty")
	}

	// Check timing
	if report.ValidationTime <= 0 {
		t.Error("Validation time should be positive")
	}

	if report.GeneratedAt.IsZero() {
		t.Error("Generated at time should not be zero")
	}
}

func TestValidationService_ValidateMultipleModels(t *testing.T) {
	logger := zap.NewNop()
	vs := NewValidationService(logger)

	models := []ModelValidator{
		&MockModelValidator{name: "model-1"},
		&MockModelValidator{name: "model-2"},
		&MockModelValidator{name: "model-3"},
	}

	config := ValidationConfig{
		CrossValidation: CrossValidationConfig{
			KFolds:          3,
			ConfidenceLevel: 0.95,
			RandomSeed:      12345,
			ParallelFolds:   false,
			MaxConcurrency:  2,
		},
		DataGeneration: DataGenerationConfig{
			TotalSamples:     20,
			TimeRange:        3 * 24 * time.Hour, // 3 days
			RiskCategories:   getTestRiskCategories(),
			IndustryWeights:  getTestIndustryWeights(),
			GeographicBias:   getTestGeographicBias(),
			SeasonalPatterns: false,
			TrendStrength:    0.01,
			NoiseLevel:       0.05,
		},
		OutputFormat: "json",
		SaveResults:  false,
		ResultsPath:  "",
	}

	ctx := context.Background()
	report, err := vs.ValidateMultipleModels(ctx, models, config)

	if err != nil {
		t.Fatalf("Multi-model validation failed: %v", err)
	}

	// Check that we have model comparison results
	if len(report.ModelComparison) != 3 {
		t.Errorf("Expected 3 model comparisons, got %d", len(report.ModelComparison))
	}

	// Check that models are ranked
	for i, model := range report.ModelComparison {
		if model.Rank != i+1 {
			t.Errorf("Expected rank %d, got %d", i+1, model.Rank)
		}
	}
}

func TestCrossValidator_EdgeCases(t *testing.T) {
	logger := zap.NewNop()
	cv := NewCrossValidator(logger)
	model := &MockModelValidator{name: "test-model"}

	// Test with insufficient samples
	samples := generateTestSamples(3) // Less than k=5
	ctx := context.Background()

	_, err := cv.CrossValidate(ctx, model, samples, 5, 0.95)
	if err == nil {
		t.Error("Expected error for insufficient samples")
	}

	// Test with empty samples
	_, err = cv.CrossValidate(ctx, model, []RiskSample{}, 3, 0.95)
	if err == nil {
		t.Error("Expected error for empty samples")
	}

	// Test with single sample
	samples = generateTestSamples(1)
	_, err = cv.CrossValidate(ctx, model, samples, 1, 0.95)
	if err == nil {
		t.Error("Expected error for single sample")
	}
}

func TestHistoricalDataGenerator_EdgeCases(t *testing.T) {
	logger := zap.NewNop()
	hdg := NewHistoricalDataGenerator(logger)

	// Test with zero samples
	config := DataGenerationConfig{
		TotalSamples:     0,
		TimeRange:        24 * time.Hour,
		RiskCategories:   getTestRiskCategories(),
		IndustryWeights:  getTestIndustryWeights(),
		GeographicBias:   getTestGeographicBias(),
		SeasonalPatterns: false,
		TrendStrength:    0.01,
		NoiseLevel:       0.05,
	}

	ctx := context.Background()
	samples, historicalSamples, err := hdg.GenerateHistoricalData(ctx, config)

	if err != nil {
		t.Fatalf("Historical data generation failed: %v", err)
	}

	if len(samples) != 0 {
		t.Errorf("Expected 0 samples, got %d", len(samples))
	}

	if len(historicalSamples) != 0 {
		t.Errorf("Expected 0 historical samples, got %d", len(historicalSamples))
	}
}

// Helper functions for testing

func generateTestSamples(count int) []RiskSample {
	samples := make([]RiskSample, count)
	for i := 0; i < count; i++ {
		samples[i] = RiskSample{
			Features: []float64{float64(i) / float64(count), float64(i%10) / 10.0},
			Label:    float64(i % 2), // Binary labels
			Metadata: map[string]interface{}{
				"index": i,
			},
		}
	}
	return samples
}

func getTestRiskCategories() []RiskCategory {
	return []RiskCategory{
		{
			Name:           "Test Risk",
			BaseRisk:       0.3,
			Volatility:     0.1,
			IndustryBias:   map[string]float64{"Technology": 1.0},
			GeographicBias: map[string]float64{"United States": 1.0},
			SizeBias:       map[string]float64{"Medium": 1.0},
			AgeBias:        map[string]float64{"3-5": 1.0},
		},
	}
}

func getTestIndustryWeights() map[string]float64 {
	return map[string]float64{
		"Technology": 0.5,
		"Finance":    0.3,
		"Healthcare": 0.2,
	}
}

func getTestGeographicBias() map[string]float64 {
	return map[string]float64{
		"United States":  0.6,
		"Canada":         0.2,
		"United Kingdom": 0.2,
	}
}
