package classification

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAccuracyValidationEngine(t *testing.T) {
	// Arrange & Act
	engine := NewAccuracyValidationEngine(nil, nil)

	// Assert
	assert.NotNil(t, engine)
	assert.Equal(t, 0.95, engine.accuracyThresholds["excellent"])
	assert.Equal(t, 0.85, engine.accuracyThresholds["good"])
	assert.Equal(t, 0.75, engine.accuracyThresholds["acceptable"])
	assert.Equal(t, 0.60, engine.accuracyThresholds["poor"])
	assert.NotNil(t, engine.validationRules)
	assert.NotNil(t, engine.feedbackRules)
	assert.NotNil(t, engine.knownClassifications)
	assert.NotNil(t, engine.industryBenchmarks)
	assert.NotNil(t, engine.accuracyHistory)
	assert.NotNil(t, engine.accuracyByIndustry)
	assert.NotNil(t, engine.accuracyByMethod)
}

func TestValidateClassification_WithKnownData(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	// Add known classification
	known := KnownClassification{
		BusinessName:       "Test Software Company",
		IndustryCode:       "511210",
		IndustryName:       "Software Publishers",
		ConfidenceLevel:    "high",
		Source:             "manual_verification",
		ValidationDate:     time.Now(),
		IsVerified:         true,
		VerificationMethod: "expert_review",
	}
	engine.AddKnownClassification(known)

	classification := IndustryClassification{
		IndustryCode:         "511210",
		IndustryName:         "Software Publishers",
		ConfidenceScore:      0.85,
		ClassificationMethod: "keyword_match",
		Keywords:             []string{"software", "publishing"},
	}

	// Act
	result := engine.ValidateClassification(context.Background(), classification, "Test Software Company")

	// Assert
	assert.NotNil(t, result)
	assert.True(t, result.IsAccurate)
	assert.Equal(t, 1.0, result.AccuracyScore)
	assert.Equal(t, "excellent", result.ConfidenceLevel)
	assert.Equal(t, "known_data_comparison", result.ValidationMethod)
	assert.NotNil(t, result.Feedback)
	assert.NotNil(t, result.Recommendations)
	assert.NotNil(t, result.BenchmarkComparison)
	assert.NotNil(t, result.HistoricalTrend)
}

func TestValidateClassification_WithBenchmarkData(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	// Add industry benchmark
	benchmark := IndustryBenchmark{
		IndustryCode:      "511210",
		IndustryName:      "Software Publishers",
		AverageAccuracy:   0.85,
		SampleSize:        100,
		CommonKeywords:    []string{"software", "publishing", "technology"},
		TypicalConfidence: 0.80,
		LastUpdated:       time.Now(),
	}
	engine.AddIndustryBenchmark(benchmark)

	classification := IndustryClassification{
		IndustryCode:         "511210",
		IndustryName:         "Software Publishers",
		ConfidenceScore:      0.85,
		ClassificationMethod: "keyword_match",
		Keywords:             []string{"software", "publishing"},
	}

	// Act
	result := engine.ValidateClassification(context.Background(), classification, "Unknown Company")

	// Assert
	assert.NotNil(t, result)
	assert.True(t, result.IsAccurate)
	assert.Greater(t, result.AccuracyScore, 0.0)
	assert.Equal(t, "benchmark_comparison", result.ValidationMethod)
	assert.NotNil(t, result.Feedback)
	assert.NotNil(t, result.Recommendations)
}

func TestValidateClassification_WithConfidenceBased(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	classification := IndustryClassification{
		IndustryCode:         "511210",
		IndustryName:         "Software Publishers",
		ConfidenceScore:      0.85,
		ClassificationMethod: "keyword_match",
		Keywords:             []string{"software", "publishing"},
	}

	// Act
	result := engine.ValidateClassification(context.Background(), classification, "Unknown Company")

	// Assert
	assert.NotNil(t, result)
	assert.True(t, result.IsAccurate)
	assert.Equal(t, 0.85, result.AccuracyScore)
	assert.Equal(t, "good", result.ConfidenceLevel)
	assert.Equal(t, "confidence_based", result.ValidationMethod)
	assert.NotNil(t, result.Feedback)
	assert.NotNil(t, result.Recommendations)
}

func TestCalculateAccuracyScore_ExactMatch(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	classification := IndustryClassification{
		IndustryCode:    "511210",
		ConfidenceScore: 0.85,
	}

	known := KnownClassification{
		IndustryCode: "511210",
	}

	// Act
	score := engine.calculateAccuracyScore(classification, known)

	// Assert
	assert.Equal(t, 1.0, score)
}

func TestCalculateAccuracyScore_MajorCategoryMatch(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	classification := IndustryClassification{
		IndustryCode:    "511210",
		ConfidenceScore: 0.85,
	}

	known := KnownClassification{
		IndustryCode: "511220",
	}

	// Act
	score := engine.calculateAccuracyScore(classification, known)

	// Assert
	assert.Equal(t, 0.8, score)
}

func TestCalculateAccuracyScore_IndustryGroupMatch(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	classification := IndustryClassification{
		IndustryCode:    "511210",
		ConfidenceScore: 0.85,
	}

	known := KnownClassification{
		IndustryCode: "511230",
	}

	// Act
	score := engine.calculateAccuracyScore(classification, known)

	// Assert
	assert.Equal(t, 0.8, score) // Major category match (51)
}

func TestCalculateAccuracyScore_NoMatch(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	classification := IndustryClassification{
		IndustryCode:    "511210",
		ConfidenceScore: 0.85,
	}

	known := KnownClassification{
		IndustryCode: "541511",
	}

	// Act
	score := engine.calculateAccuracyScore(classification, known)

	// Assert
	assert.Less(t, score, 0.5)
	assert.GreaterOrEqual(t, score, 0.0)
}

func TestCalculateBenchmarkAccuracy(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	classification := IndustryClassification{
		IndustryCode:    "511210",
		ConfidenceScore: 0.85,
	}

	benchmark := IndustryBenchmark{
		IndustryCode:      "511210",
		AverageAccuracy:   0.85,
		TypicalConfidence: 0.80,
	}

	// Act
	score := engine.calculateBenchmarkAccuracy(classification, benchmark)

	// Assert
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)
}

func TestDetermineAccuracyLevel(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	// Act & Assert
	assert.Equal(t, "excellent", engine.determineAccuracyLevel(0.95))
	assert.Equal(t, "excellent", engine.determineAccuracyLevel(1.0))
	assert.Equal(t, "good", engine.determineAccuracyLevel(0.85))
	assert.Equal(t, "good", engine.determineAccuracyLevel(0.94))
	assert.Equal(t, "acceptable", engine.determineAccuracyLevel(0.75))
	assert.Equal(t, "acceptable", engine.determineAccuracyLevel(0.84))
	assert.Equal(t, "poor", engine.determineAccuracyLevel(0.60))
	assert.Equal(t, "poor", engine.determineAccuracyLevel(0.74))
	assert.Equal(t, "poor", engine.determineAccuracyLevel(0.50))
}

func TestRecordAccuracyMetrics(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	classification := IndustryClassification{
		IndustryCode:         "511210",
		ClassificationMethod: "keyword_match",
	}

	result := &ValidationResult{
		IsAccurate:    true,
		AccuracyScore: 0.85,
	}

	// Act
	engine.recordAccuracyMetrics(classification, result)

	// Assert
	assert.Equal(t, int64(1), engine.totalValidations)
	assert.Equal(t, int64(1), engine.successfulValidations)
	assert.Equal(t, 1.0, engine.averageAccuracy)
	assert.Equal(t, 0.85, engine.accuracyByIndustry["511210"])
	assert.Equal(t, 0.85, engine.accuracyByMethod["keyword_match"])
}

func TestGenerateFeedback(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	classification := IndustryClassification{
		IndustryCode:         "511210",
		ConfidenceScore:      0.95,
		ClassificationMethod: "keyword_match",
		Keywords:             []string{"software", "publishing"},
	}

	known := KnownClassification{
		IndustryCode: "511210",
	}

	// Act
	feedback := engine.generateFeedback(classification, known)

	// Assert
	assert.NotNil(t, feedback)
	assert.Greater(t, len(feedback), 0)
}

func TestGenerateRecommendations(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	classification := IndustryClassification{
		IndustryCode:         "511210",
		ConfidenceScore:      0.65,
		ClassificationMethod: "fuzzy_match",
		Keywords:             []string{},
	}

	result := &ValidationResult{
		AccuracyScore: 0.65,
	}

	// Act
	recommendations := engine.generateRecommendations(classification, result)

	// Assert
	assert.NotNil(t, recommendations)
	assert.Greater(t, len(recommendations), 0)
}

func TestCalculateBenchmarkComparison(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	benchmark := IndustryBenchmark{
		IndustryCode:    "511210",
		AverageAccuracy: 0.85,
	}
	engine.AddIndustryBenchmark(benchmark)

	// Act
	comparison := engine.calculateBenchmarkComparison("511210", 0.90)

	// Assert
	assert.NotNil(t, comparison)
	assert.Equal(t, 0.85, comparison.IndustryBenchmark)
	assert.Equal(t, 0.90, comparison.CurrentAccuracy)
	assert.InDelta(t, 0.05, comparison.Difference, 0.001)
	assert.True(t, comparison.IsAboveBenchmark)
	assert.Equal(t, "industry_benchmark", comparison.BenchmarkSource)
}

func TestCalculateAccuracyTrend(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	// Add some accuracy history
	engine.accuracyHistory = []AccuracyRecord{
		{
			PredictedCode: "511210",
			Accuracy:      0.80,
			Timestamp:     time.Now().Add(-24 * time.Hour),
		},
		{
			PredictedCode: "511210",
			Accuracy:      0.85,
			Timestamp:     time.Now(),
		},
	}

	// Act
	trend := engine.calculateAccuracyTrend("511210")

	// Assert
	assert.NotNil(t, trend)
	assert.Equal(t, 30, trend.PeriodDays)
	assert.Equal(t, 0.825, trend.AverageAccuracy)
	assert.Equal(t, "stable", trend.TrendDirection)
	assert.Equal(t, 2, trend.DataPoints)
}

func TestCalculatePercentile(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	// Add accuracy history
	engine.accuracyHistory = []AccuracyRecord{
		{PredictedCode: "511210", Accuracy: 0.70},
		{PredictedCode: "511210", Accuracy: 0.80},
		{PredictedCode: "511210", Accuracy: 0.90},
	}

	// Act
	percentile := engine.calculatePercentile(0.85, "511210")

	// Assert
	assert.Equal(t, 66.66666666666666, percentile)
}

func TestAddKnownClassification(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	known := KnownClassification{
		BusinessName: "Test Company",
		IndustryCode: "511210",
	}

	// Act
	engine.AddKnownClassification(known)

	// Assert
	assert.Equal(t, 1, len(engine.knownClassifications))
	assert.Equal(t, "511210", engine.knownClassifications["Test Company"].IndustryCode)
}

func TestAddIndustryBenchmark(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	benchmark := IndustryBenchmark{
		IndustryCode:    "511210",
		AverageAccuracy: 0.85,
	}

	// Act
	engine.AddIndustryBenchmark(benchmark)

	// Assert
	assert.Equal(t, 1, len(engine.industryBenchmarks))
	assert.Equal(t, 0.85, engine.industryBenchmarks["511210"].AverageAccuracy)
}

func TestGetAccuracyMetrics(t *testing.T) {
	// Arrange
	engine := NewAccuracyValidationEngine(nil, nil)

	// Add some data
	engine.totalValidations = 10
	engine.successfulValidations = 8
	engine.averageAccuracy = 0.8
	engine.accuracyByIndustry["511210"] = 0.85
	engine.accuracyByMethod["keyword_match"] = 0.9

	// Act
	metrics := engine.GetAccuracyMetrics()

	// Assert
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(10), metrics["total_validations"])
	assert.Equal(t, int64(8), metrics["successful_validations"])
	assert.Equal(t, 0.8, metrics["average_accuracy"])
	assert.Equal(t, 0.85, metrics["accuracy_by_industry"].(map[string]float64)["511210"])
	assert.Equal(t, 0.9, metrics["accuracy_by_method"].(map[string]float64)["keyword_match"])
}

func TestValidationResult_Complete(t *testing.T) {
	// Arrange
	result := &ValidationResult{
		IsAccurate:       true,
		AccuracyScore:    0.85,
		ConfidenceLevel:  "good",
		ValidationMethod: "known_data_comparison",
		Feedback:         []string{"High-confidence classification"},
		Recommendations:  []string{"Classification accuracy is within acceptable range"},
		Timestamp:        time.Now(),
	}

	// Act & Assert
	assert.True(t, result.IsAccurate)
	assert.Equal(t, 0.85, result.AccuracyScore)
	assert.Equal(t, "good", result.ConfidenceLevel)
	assert.Equal(t, "known_data_comparison", result.ValidationMethod)
	assert.NotNil(t, result.Feedback)
	assert.NotNil(t, result.Recommendations)
	assert.NotNil(t, result.Timestamp)
}

func TestBenchmarkComparison_Complete(t *testing.T) {
	// Arrange
	comparison := &BenchmarkComparison{
		IndustryBenchmark: 0.85,
		CurrentAccuracy:   0.90,
		Difference:        0.05,
		Percentile:        75.0,
		IsAboveBenchmark:  true,
		BenchmarkSource:   "industry_benchmark",
	}

	// Act & Assert
	assert.Equal(t, 0.85, comparison.IndustryBenchmark)
	assert.Equal(t, 0.90, comparison.CurrentAccuracy)
	assert.Equal(t, 0.05, comparison.Difference)
	assert.Equal(t, 75.0, comparison.Percentile)
	assert.True(t, comparison.IsAboveBenchmark)
	assert.Equal(t, "industry_benchmark", comparison.BenchmarkSource)
}

func TestAccuracyTrend_Complete(t *testing.T) {
	// Arrange
	trend := &AccuracyTrend{
		PeriodDays:      30,
		AverageAccuracy: 0.85,
		TrendDirection:  "improving",
		TrendStrength:   0.05,
		DataPoints:      10,
	}

	// Act & Assert
	assert.Equal(t, 30, trend.PeriodDays)
	assert.Equal(t, 0.85, trend.AverageAccuracy)
	assert.Equal(t, "improving", trend.TrendDirection)
	assert.Equal(t, 0.05, trend.TrendStrength)
	assert.Equal(t, 10, trend.DataPoints)
}
