package industry_codes

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestDataQualityScorer(t *testing.T) (*DataQualityScorer, *sql.DB) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	logger := zap.NewNop()
	industryDB := NewIndustryCodeDatabase(db, logger)
	scorer := NewDataQualityScorer(industryDB, logger)

	return scorer, db
}

func TestDataQualityScorer_AssessDataQuality(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	ctx := context.Background()
	config := &DataQualityConfig{
		EnableCompleteness: true,
		EnableAccuracy:     true,
		EnableConsistency:  true,
		EnableValidity:     true,
		EnableUniqueness:   true,
		Thresholds: map[string]float64{
			"completeness": 0.80,
			"accuracy":     0.85,
			"consistency":  0.80,
			"validity":     0.85,
			"uniqueness":   0.90,
		},
		Weights: map[string]float64{
			"completeness": 0.20,
			"accuracy":     0.25,
			"consistency":  0.20,
			"validity":     0.20,
			"uniqueness":   0.15,
		},
		QualityRules: []QualityRule{
			{
				ID:          "rule_001",
				Name:        "Required Business Name",
				Description: "Business name must be present",
				Type:        "completeness",
				Field:       "business_name",
				Condition:   "not_empty",
				Severity:    "high",
				Enabled:     true,
			},
		},
	}

	// Mock data
	data := map[string]interface{}{
		"business_name": "Test Company",
		"address":       "123 Test St",
		"phone":         "+1-555-123-4567",
		"email":         "test@company.com",
	}

	score, err := scorer.AssessDataQuality(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, score)

	// Test basic structure
	assert.NotEmpty(t, score.ID)
	assert.WithinDuration(t, time.Now(), score.GeneratedAt, 5*time.Second)
	assert.Greater(t, score.OverallScore, 0.0)
	assert.LessOrEqual(t, score.OverallScore, 1.0)
	assert.NotEmpty(t, score.QualityLevel)
	assert.Contains(t, []string{"excellent", "good", "fair", "poor", "critical"}, score.QualityLevel)

	// Test dimensions
	assert.NotNil(t, score.Dimensions.Completeness)
	assert.NotNil(t, score.Dimensions.Accuracy)
	assert.NotNil(t, score.Dimensions.Consistency)
	assert.NotNil(t, score.Dimensions.Validity)
	assert.NotNil(t, score.Dimensions.Uniqueness)

	// Test completeness metrics
	assert.Greater(t, score.Dimensions.Completeness.OverallCompleteness, 0.0)
	assert.LessOrEqual(t, score.Dimensions.Completeness.OverallCompleteness, 1.0)
	assert.NotEmpty(t, score.Dimensions.Completeness.FieldCompleteness)
	assert.Greater(t, score.Dimensions.Completeness.RecordCompleteness, 0.0)

	// Test accuracy metrics
	assert.Greater(t, score.Dimensions.Accuracy.OverallAccuracy, 0.0)
	assert.LessOrEqual(t, score.Dimensions.Accuracy.OverallAccuracy, 1.0)
	assert.NotEmpty(t, score.Dimensions.Accuracy.FieldAccuracy)
	assert.GreaterOrEqual(t, score.Dimensions.Accuracy.ErrorRate, 0.0)
	assert.LessOrEqual(t, score.Dimensions.Accuracy.ErrorRate, 1.0)

	// Test consistency metrics
	assert.Greater(t, score.Dimensions.Consistency.OverallConsistency, 0.0)
	assert.LessOrEqual(t, score.Dimensions.Consistency.OverallConsistency, 1.0)
	assert.NotEmpty(t, score.Dimensions.Consistency.FieldConsistency)

	// Test validity metrics
	assert.Greater(t, score.Dimensions.Validity.OverallValidity, 0.0)
	assert.LessOrEqual(t, score.Dimensions.Validity.OverallValidity, 1.0)
	assert.NotEmpty(t, score.Dimensions.Validity.FieldValidity)

	// Test uniqueness metrics
	assert.Greater(t, score.Dimensions.Uniqueness.OverallUniqueness, 0.0)
	assert.LessOrEqual(t, score.Dimensions.Uniqueness.OverallUniqueness, 1.0)
	assert.GreaterOrEqual(t, score.Dimensions.Uniqueness.DuplicateRate, 0.0)
	assert.LessOrEqual(t, score.Dimensions.Uniqueness.DuplicateRate, 1.0)

	// Test issues and recommendations
	assert.NotNil(t, score.Issues)
	assert.NotNil(t, score.Recommendations)

	// Test metadata
	assert.NotNil(t, score.Metadata)
	assert.Greater(t, score.Metadata.AssessmentDuration, time.Duration(0))
	assert.NotEmpty(t, score.Metadata.Thresholds)
	assert.NotEmpty(t, score.Metadata.QualityRules)

	// Test trends
	assert.NotNil(t, score.Trends)
	assert.NotEmpty(t, score.Trends.OverallTrend)
	assert.NotEmpty(t, score.Trends.DimensionTrends)

	// Test validation results
	assert.NotNil(t, score.ValidationResults)
	assert.Greater(t, score.ValidationResults.TotalValidations, 0)
	assert.GreaterOrEqual(t, score.ValidationResults.ValidationRate, 0.0)
	assert.LessOrEqual(t, score.ValidationResults.ValidationRate, 1.0)
}

func TestDataQualityScorer_AssessDataQuality_DefaultConfig(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	ctx := context.Background()
	config := &DataQualityConfig{
		// Minimal config with defaults
		EnableCompleteness: true,
		EnableAccuracy:     true,
	}

	data := map[string]interface{}{
		"business_name": "Test Company",
	}

	score, err := scorer.AssessDataQuality(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, score)

	// Should still generate a valid score with defaults
	assert.Greater(t, score.OverallScore, 0.0)
	assert.NotEmpty(t, score.QualityLevel)
	assert.NotEmpty(t, score.Dimensions.Completeness.OverallCompleteness)
	assert.NotEmpty(t, score.Dimensions.Accuracy.OverallAccuracy)
}

func TestDataQualityScorer_AssessDataQuality_NoDimensions(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	ctx := context.Background()
	config := &DataQualityConfig{
		// No dimensions enabled
		EnableCompleteness: false,
		EnableAccuracy:     false,
		EnableConsistency:  false,
		EnableValidity:     false,
		EnableUniqueness:   false,
	}

	data := map[string]interface{}{
		"business_name": "Test Company",
	}

	score, err := scorer.AssessDataQuality(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, score)

	// Should return zero score when no dimensions are enabled
	assert.Equal(t, 0.0, score.OverallScore)
	assert.Equal(t, "critical", score.QualityLevel)
}

func TestDataQualityScorer_AssessCompleteness(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	ctx := context.Background()
	config := &DataQualityConfig{}

	data := map[string]interface{}{
		"business_name": "Test Company",
		"address":       "123 Test St",
	}

	completeness, err := scorer.assessCompleteness(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, completeness)

	// Test completeness metrics
	assert.Greater(t, completeness.OverallCompleteness, 0.0)
	assert.LessOrEqual(t, completeness.OverallCompleteness, 1.0)
	assert.NotEmpty(t, completeness.FieldCompleteness)
	assert.Greater(t, completeness.RecordCompleteness, 0.0)
	assert.Greater(t, completeness.RequiredFields, 0.0)
	assert.Greater(t, completeness.OptionalFields, 0.0)
	assert.NotNil(t, completeness.MissingPatterns)
	assert.GreaterOrEqual(t, completeness.CompletenessTrend, -1.0)
	assert.LessOrEqual(t, completeness.CompletenessTrend, 1.0)

	// Test field completeness
	for field, score := range completeness.FieldCompleteness {
		assert.NotEmpty(t, field)
		assert.GreaterOrEqual(t, score, 0.0)
		assert.LessOrEqual(t, score, 1.0)
	}

	// Test missing patterns
	for _, pattern := range completeness.MissingPatterns {
		assert.NotEmpty(t, pattern.Field)
		assert.GreaterOrEqual(t, pattern.MissingRate, 0.0)
		assert.LessOrEqual(t, pattern.MissingRate, 1.0)
		assert.NotEmpty(t, pattern.Pattern)
		assert.NotEmpty(t, pattern.Impact)
		assert.NotEmpty(t, pattern.Recommendation)
	}
}

func TestDataQualityScorer_AssessAccuracy(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	ctx := context.Background()
	config := &DataQualityConfig{}

	data := map[string]interface{}{
		"business_name": "Test Company",
		"email":         "test@company.com",
	}

	accuracy, err := scorer.assessAccuracy(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, accuracy)

	// Test accuracy metrics
	assert.Greater(t, accuracy.OverallAccuracy, 0.0)
	assert.LessOrEqual(t, accuracy.OverallAccuracy, 1.0)
	assert.NotEmpty(t, accuracy.FieldAccuracy)
	assert.GreaterOrEqual(t, accuracy.ErrorRate, 0.0)
	assert.LessOrEqual(t, accuracy.ErrorRate, 1.0)
	assert.Greater(t, accuracy.Precision, 0.0)
	assert.LessOrEqual(t, accuracy.Precision, 1.0)
	assert.Greater(t, accuracy.Recall, 0.0)
	assert.LessOrEqual(t, accuracy.Recall, 1.0)
	assert.Greater(t, accuracy.F1Score, 0.0)
	assert.LessOrEqual(t, accuracy.F1Score, 1.0)
	assert.GreaterOrEqual(t, accuracy.AccuracyTrend, -1.0)
	assert.LessOrEqual(t, accuracy.AccuracyTrend, 1.0)

	// Test field accuracy
	for field, score := range accuracy.FieldAccuracy {
		assert.NotEmpty(t, field)
		assert.GreaterOrEqual(t, score, 0.0)
		assert.LessOrEqual(t, score, 1.0)
	}

	// Test validation errors
	for _, error := range accuracy.ValidationErrors {
		assert.NotEmpty(t, error.Field)
		assert.NotEmpty(t, error.ErrorType)
		assert.NotEmpty(t, error.Description)
		assert.NotEmpty(t, error.Severity)
		assert.GreaterOrEqual(t, error.Count, 0)
		assert.GreaterOrEqual(t, error.Percentage, 0.0)
		assert.LessOrEqual(t, error.Percentage, 1.0)
	}
}

func TestDataQualityScorer_AssessConsistency(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	ctx := context.Background()
	config := &DataQualityConfig{}

	data := map[string]interface{}{
		"business_name": "Test Company",
		"phone":         "+1-555-123-4567",
	}

	consistency, err := scorer.assessConsistency(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, consistency)

	// Test consistency metrics
	assert.Greater(t, consistency.OverallConsistency, 0.0)
	assert.LessOrEqual(t, consistency.OverallConsistency, 1.0)
	assert.NotEmpty(t, consistency.FieldConsistency)
	assert.Greater(t, consistency.CrossFieldConsistency, 0.0)
	assert.LessOrEqual(t, consistency.CrossFieldConsistency, 1.0)
	assert.Greater(t, consistency.FormatConsistency, 0.0)
	assert.LessOrEqual(t, consistency.FormatConsistency, 1.0)
	assert.Greater(t, consistency.ValueConsistency, 0.0)
	assert.LessOrEqual(t, consistency.ValueConsistency, 1.0)
	assert.GreaterOrEqual(t, consistency.ConsistencyTrend, -1.0)
	assert.LessOrEqual(t, consistency.ConsistencyTrend, 1.0)

	// Test field consistency
	for field, score := range consistency.FieldConsistency {
		assert.NotEmpty(t, field)
		assert.GreaterOrEqual(t, score, 0.0)
		assert.LessOrEqual(t, score, 1.0)
	}

	// Test inconsistencies
	for _, inconsistency := range consistency.Inconsistencies {
		assert.NotEmpty(t, inconsistency.Type)
		assert.NotEmpty(t, inconsistency.Fields)
		assert.NotEmpty(t, inconsistency.Description)
		assert.GreaterOrEqual(t, inconsistency.Count, 0)
		assert.NotEmpty(t, inconsistency.Impact)
	}
}

func TestDataQualityScorer_AssessValidity(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	ctx := context.Background()
	config := &DataQualityConfig{}

	data := map[string]interface{}{
		"business_name": "Test Company",
		"email":         "test@company.com",
	}

	validity, err := scorer.assessValidity(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, validity)

	// Test validity metrics
	assert.Greater(t, validity.OverallValidity, 0.0)
	assert.LessOrEqual(t, validity.OverallValidity, 1.0)
	assert.NotEmpty(t, validity.FieldValidity)
	assert.Greater(t, validity.FormatValidity, 0.0)
	assert.LessOrEqual(t, validity.FormatValidity, 1.0)
	assert.Greater(t, validity.RangeValidity, 0.0)
	assert.LessOrEqual(t, validity.RangeValidity, 1.0)
	assert.Greater(t, validity.DomainValidity, 0.0)
	assert.LessOrEqual(t, validity.DomainValidity, 1.0)
	assert.GreaterOrEqual(t, validity.ValidityTrend, -1.0)
	assert.LessOrEqual(t, validity.ValidityTrend, 1.0)

	// Test field validity
	for field, score := range validity.FieldValidity {
		assert.NotEmpty(t, field)
		assert.GreaterOrEqual(t, score, 0.0)
		assert.LessOrEqual(t, score, 1.0)
	}

	// Test invalid records
	for _, record := range validity.InvalidRecords {
		assert.NotEmpty(t, record.RecordID)
		assert.NotEmpty(t, record.Field)
		assert.NotEmpty(t, record.Value)
		assert.NotEmpty(t, record.Issue)
		assert.NotEmpty(t, record.Severity)
		assert.NotEmpty(t, record.Suggestions)
	}
}

func TestDataQualityScorer_AssessUniqueness(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	ctx := context.Background()
	config := &DataQualityConfig{}

	data := map[string]interface{}{
		"business_name": "Test Company",
		"address":       "123 Test St",
	}

	uniqueness, err := scorer.assessUniqueness(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, uniqueness)

	// Test uniqueness metrics
	assert.Greater(t, uniqueness.OverallUniqueness, 0.0)
	assert.LessOrEqual(t, uniqueness.OverallUniqueness, 1.0)
	assert.GreaterOrEqual(t, uniqueness.DuplicateRate, 0.0)
	assert.LessOrEqual(t, uniqueness.DuplicateRate, 1.0)
	assert.Greater(t, uniqueness.UniqueRecords, 0.0)
	assert.LessOrEqual(t, uniqueness.UniqueRecords, 1.0)
	assert.GreaterOrEqual(t, uniqueness.UniquenessTrend, -1.0)
	assert.LessOrEqual(t, uniqueness.UniquenessTrend, 1.0)

	// Test duplicate patterns
	for _, pattern := range uniqueness.DuplicatePatterns {
		assert.NotEmpty(t, pattern.Fields)
		assert.GreaterOrEqual(t, pattern.DuplicateCount, 0)
		assert.GreaterOrEqual(t, pattern.Percentage, 0.0)
		assert.LessOrEqual(t, pattern.Percentage, 1.0)
		assert.Greater(t, pattern.Confidence, 0.0)
		assert.LessOrEqual(t, pattern.Confidence, 1.0)
		assert.NotEmpty(t, pattern.Action)
	}
}

func TestDataQualityScorer_CalculateOverallScore(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	dimensions := QualityDimensions{
		Completeness: CompletenessMetrics{
			OverallCompleteness: 0.85,
		},
		Accuracy: DataAccuracyMetrics{
			OverallAccuracy: 0.88,
		},
		Consistency: ConsistencyMetrics{
			OverallConsistency: 0.83,
		},
		Validity: ValidityMetrics{
			OverallValidity: 0.86,
		},
		Uniqueness: UniquenessMetrics{
			OverallUniqueness: 0.91,
		},
	}

	config := &DataQualityConfig{
		EnableCompleteness: true,
		EnableAccuracy:     true,
		EnableConsistency:  true,
		EnableValidity:     true,
		EnableUniqueness:   true,
	}

	score := scorer.calculateOverallScore(dimensions, config)
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)

	// Test with custom weights
	config.Weights = map[string]float64{
		"completeness": 0.30,
		"accuracy":     0.40,
		"consistency":  0.15,
		"validity":     0.10,
		"uniqueness":   0.05,
	}

	scoreWithWeights := scorer.calculateOverallScore(dimensions, config)
	assert.Greater(t, scoreWithWeights, 0.0)
	assert.LessOrEqual(t, scoreWithWeights, 1.0)

	// Test with no dimensions enabled
	config.EnableCompleteness = false
	config.EnableAccuracy = false
	config.EnableConsistency = false
	config.EnableValidity = false
	config.EnableUniqueness = false

	scoreNoDimensions := scorer.calculateOverallScore(dimensions, config)
	assert.Equal(t, 0.0, scoreNoDimensions)
}

func TestDataQualityScorer_DetermineQualityLevel(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	// Test different score ranges
	assert.Equal(t, "excellent", scorer.determineQualityLevel(0.95))
	assert.Equal(t, "excellent", scorer.determineQualityLevel(0.90))
	assert.Equal(t, "good", scorer.determineQualityLevel(0.85))
	assert.Equal(t, "good", scorer.determineQualityLevel(0.80))
	assert.Equal(t, "fair", scorer.determineQualityLevel(0.75))
	assert.Equal(t, "fair", scorer.determineQualityLevel(0.70))
	assert.Equal(t, "poor", scorer.determineQualityLevel(0.65))
	assert.Equal(t, "poor", scorer.determineQualityLevel(0.60))
	assert.Equal(t, "critical", scorer.determineQualityLevel(0.55))
	assert.Equal(t, "critical", scorer.determineQualityLevel(0.0))
}

func TestDataQualityScorer_GenerateQualityIssues(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	dimensions := QualityDimensions{
		Completeness: CompletenessMetrics{
			OverallCompleteness: 0.75, // Below threshold
		},
		Accuracy: DataAccuracyMetrics{
			OverallAccuracy: 0.82, // Below threshold
			ValidationErrors: []ValidationError{
				{
					Field:       "email",
					ErrorType:   "format",
					Description: "Invalid email format",
					Severity:    "high",
					Count:       100,
					Percentage:  0.15, // Above 5% threshold
				},
			},
		},
	}

	config := &DataQualityConfig{
		Thresholds: map[string]float64{
			"completeness": 0.80,
			"accuracy":     0.85,
		},
	}

	issues := scorer.generateQualityIssues(dimensions, config)
	assert.NotEmpty(t, issues)

	// Should have completeness issue
	hasCompletenessIssue := false
	for _, issue := range issues {
		if issue.Type == "completeness" {
			hasCompletenessIssue = true
			assert.Equal(t, "high", issue.Severity)
			assert.Equal(t, "high", issue.Priority)
			assert.Equal(t, "open", issue.Status)
			assert.NotEmpty(t, issue.Description)
			assert.NotEmpty(t, issue.Impact)
		}
	}
	assert.True(t, hasCompletenessIssue, "Should generate completeness issue")

	// Should have accuracy issue
	hasAccuracyIssue := false
	for _, issue := range issues {
		if issue.Type == "accuracy" {
			hasAccuracyIssue = true
			assert.Equal(t, "high", issue.Severity)
			assert.Equal(t, "high", issue.Priority)
			assert.Equal(t, "open", issue.Status)
			assert.NotEmpty(t, issue.Description)
			assert.NotEmpty(t, issue.Impact)
		}
	}
	assert.True(t, hasAccuracyIssue, "Should generate accuracy issue")

	// Should have validation issue
	hasValidationIssue := false
	for _, issue := range issues {
		if issue.Type == "validation" {
			hasValidationIssue = true
			assert.Equal(t, "high", issue.Severity)
			assert.Equal(t, "medium", issue.Priority)
			assert.Equal(t, "open", issue.Status)
			assert.NotEmpty(t, issue.Description)
			assert.NotEmpty(t, issue.Impact)
			assert.Contains(t, issue.Fields, "email")
		}
	}
	assert.True(t, hasValidationIssue, "Should generate validation issue")
}

func TestDataQualityScorer_GenerateQualityRecommendations(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	dimensions := QualityDimensions{
		Completeness: CompletenessMetrics{
			OverallCompleteness: 0.80, // Below threshold
		},
		Accuracy: DataAccuracyMetrics{
			OverallAccuracy: 0.85, // Below threshold
		},
	}

	issues := []QualityIssue{
		{
			ID:       "issue_validation_email",
			Type:     "validation",
			Severity: "high",
			Fields:   []string{"email"},
		},
	}

	config := &DataQualityConfig{}

	recommendations := scorer.generateQualityRecommendations(dimensions, issues, config)
	assert.NotEmpty(t, recommendations)

	// Should have completeness recommendation
	hasCompletenessRec := false
	for _, rec := range recommendations {
		if rec.Type == "completeness" {
			hasCompletenessRec = true
			assert.Equal(t, "high", rec.Priority)
			assert.NotEmpty(t, rec.Title)
			assert.NotEmpty(t, rec.Description)
			assert.NotEmpty(t, rec.Impact)
			assert.NotEmpty(t, rec.Effort)
			assert.Greater(t, rec.ROI, 0.0)
			assert.NotEmpty(t, rec.Actions)
			assert.Greater(t, rec.ExpectedImprovement, 0.0)
			assert.NotEmpty(t, rec.Timeline)
			assert.NotEmpty(t, rec.Resources)
			assert.NotEmpty(t, rec.SuccessMetrics)
		}
	}
	assert.True(t, hasCompletenessRec, "Should generate completeness recommendation")

	// Should have accuracy recommendation
	hasAccuracyRec := false
	for _, rec := range recommendations {
		if rec.Type == "accuracy" {
			hasAccuracyRec = true
			assert.Equal(t, "high", rec.Priority)
			assert.NotEmpty(t, rec.Title)
			assert.NotEmpty(t, rec.Description)
			assert.NotEmpty(t, rec.Impact)
			assert.NotEmpty(t, rec.Effort)
			assert.Greater(t, rec.ROI, 0.0)
			assert.NotEmpty(t, rec.Actions)
			assert.Greater(t, rec.ExpectedImprovement, 0.0)
			assert.NotEmpty(t, rec.Timeline)
			assert.NotEmpty(t, rec.Resources)
			assert.NotEmpty(t, rec.SuccessMetrics)
		}
	}
	assert.True(t, hasAccuracyRec, "Should generate accuracy recommendation")

	// Should have validation recommendation
	hasValidationRec := false
	for _, rec := range recommendations {
		if rec.Type == "validation" {
			hasValidationRec = true
			assert.Equal(t, "medium", rec.Priority)
			assert.NotEmpty(t, rec.Title)
			assert.NotEmpty(t, rec.Description)
			assert.NotEmpty(t, rec.Impact)
			assert.NotEmpty(t, rec.Effort)
			assert.Greater(t, rec.ROI, 0.0)
			assert.NotEmpty(t, rec.Actions)
			assert.Greater(t, rec.ExpectedImprovement, 0.0)
			assert.NotEmpty(t, rec.Timeline)
			assert.NotEmpty(t, rec.Resources)
			assert.NotEmpty(t, rec.SuccessMetrics)
		}
	}
	assert.True(t, hasValidationRec, "Should generate validation recommendation")
}

func TestDataQualityScorer_GenerateQualityTrends(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	score := &DataQualityScore{
		OverallScore: 0.85,
		Dimensions: QualityDimensions{
			Completeness: CompletenessMetrics{OverallCompleteness: 0.85},
			Accuracy:     DataAccuracyMetrics{OverallAccuracy: 0.88},
			Consistency:  ConsistencyMetrics{OverallConsistency: 0.83},
			Validity:     ValidityMetrics{OverallValidity: 0.86},
			Uniqueness:   UniquenessMetrics{OverallUniqueness: 0.91},
		},
	}

	trends := scorer.generateQualityTrends(score)
	assert.NotNil(t, trends)

	// Test trends structure
	assert.NotEmpty(t, trends.OverallTrend)
	assert.Contains(t, []string{"improving", "stable", "declining"}, trends.OverallTrend)
	assert.NotEmpty(t, trends.DimensionTrends)
	assert.NotNil(t, trends.HistoricalScores)
	assert.NotNil(t, trends.TrendAnalysis)

	// Test dimension trends
	for dimension, trend := range trends.DimensionTrends {
		assert.NotEmpty(t, dimension)
		assert.Contains(t, []string{"improving", "stable", "declining"}, trend)
	}

	// Test historical scores
	for _, historical := range trends.HistoricalScores {
		assert.False(t, historical.Date.IsZero())
		assert.Greater(t, historical.OverallScore, 0.0)
		assert.LessOrEqual(t, historical.OverallScore, 1.0)
		assert.NotEmpty(t, historical.Dimensions)
	}

	// Test trend analysis
	assert.NotEmpty(t, trends.TrendAnalysis.TrendDirection)
	assert.Contains(t, []string{"improving", "stable", "declining"}, trends.TrendAnalysis.TrendDirection)
	assert.Greater(t, trends.TrendAnalysis.TrendStrength, 0.0)
	assert.LessOrEqual(t, trends.TrendAnalysis.TrendStrength, 1.0)
	assert.GreaterOrEqual(t, trends.TrendAnalysis.Volatility, 0.0)
	assert.LessOrEqual(t, trends.TrendAnalysis.Volatility, 1.0)
	assert.NotNil(t, trends.TrendAnalysis.Outliers)
}

func TestDataQualityScorer_PerformValidation(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	data := map[string]interface{}{
		"business_name": "Test Company",
		"email":         "test@company.com",
	}

	config := &DataQualityConfig{}

	results := scorer.performValidation(data, config)
	assert.NotNil(t, results)

	// Test validation results structure
	assert.Greater(t, results.TotalValidations, 0)
	assert.GreaterOrEqual(t, results.PassedValidations, 0)
	assert.GreaterOrEqual(t, results.FailedValidations, 0)
	assert.Equal(t, results.TotalValidations, results.PassedValidations+results.FailedValidations)
	assert.GreaterOrEqual(t, results.ValidationRate, 0.0)
	assert.LessOrEqual(t, results.ValidationRate, 1.0)
	assert.NotEmpty(t, results.FieldResults)
	assert.NotEmpty(t, results.RuleResults)

	// Test field results
	for field, fieldResult := range results.FieldResults {
		assert.NotEmpty(t, field)
		assert.Equal(t, field, fieldResult.FieldName)
		assert.Greater(t, fieldResult.TotalRecords, 0)
		assert.GreaterOrEqual(t, fieldResult.ValidRecords, 0)
		assert.GreaterOrEqual(t, fieldResult.InvalidRecords, 0)
		assert.Equal(t, fieldResult.TotalRecords, fieldResult.ValidRecords+fieldResult.InvalidRecords)
		assert.GreaterOrEqual(t, fieldResult.ValidationRate, 0.0)
		assert.LessOrEqual(t, fieldResult.ValidationRate, 1.0)
		assert.NotNil(t, fieldResult.CommonErrors)
		assert.NotEmpty(t, fieldResult.ErrorDistribution)
	}

	// Test rule results
	for rule, ruleResult := range results.RuleResults {
		assert.NotEmpty(t, rule)
		assert.Equal(t, rule, ruleResult.RuleName)
		assert.NotEmpty(t, ruleResult.RuleType)
		assert.Greater(t, ruleResult.TotalChecks, 0)
		assert.GreaterOrEqual(t, ruleResult.PassedChecks, 0)
		assert.GreaterOrEqual(t, ruleResult.FailedChecks, 0)
		assert.Equal(t, ruleResult.TotalChecks, ruleResult.PassedChecks+ruleResult.FailedChecks)
		assert.GreaterOrEqual(t, ruleResult.SuccessRate, 0.0)
		assert.LessOrEqual(t, ruleResult.SuccessRate, 1.0)
		assert.NotEmpty(t, ruleResult.Severity)
		assert.NotEmpty(t, ruleResult.Impact)
	}
}

func TestDataQualityScorer_ExtractRuleNames(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	rules := []QualityRule{
		{
			ID:      "rule_001",
			Name:    "Required Business Name",
			Enabled: true,
		},
		{
			ID:      "rule_002",
			Name:    "Email Format",
			Enabled: false,
		},
		{
			ID:      "rule_003",
			Name:    "Phone Format",
			Enabled: true,
		},
	}

	names := scorer.extractRuleNames(rules)
	assert.Len(t, names, 2) // Only enabled rules
	assert.Contains(t, names, "Required Business Name")
	assert.Contains(t, names, "Phone Format")
	assert.NotContains(t, names, "Email Format")
}

func TestDataQualityScorer_Integration(t *testing.T) {
	scorer, db := setupTestDataQualityScorer(t)
	defer db.Close()

	ctx := context.Background()

	// Test complete flow with realistic configuration
	config := &DataQualityConfig{
		EnableCompleteness: true,
		EnableAccuracy:     true,
		EnableConsistency:  true,
		EnableValidity:     true,
		EnableUniqueness:   true,
		Thresholds: map[string]float64{
			"completeness": 0.80,
			"accuracy":     0.85,
			"consistency":  0.80,
			"validity":     0.85,
			"uniqueness":   0.90,
		},
		Weights: map[string]float64{
			"completeness": 0.20,
			"accuracy":     0.25,
			"consistency":  0.20,
			"validity":     0.20,
			"uniqueness":   0.15,
		},
		QualityRules: []QualityRule{
			{
				ID:          "rule_001",
				Name:        "Required Business Name",
				Description: "Business name must be present",
				Type:        "completeness",
				Field:       "business_name",
				Condition:   "not_empty",
				Severity:    "high",
				Enabled:     true,
			},
			{
				ID:          "rule_002",
				Name:        "Email Format",
				Description: "Email must be in valid format",
				Type:        "validity",
				Field:       "email",
				Condition:   "email_format",
				Severity:    "medium",
				Enabled:     true,
			},
		},
	}

	data := map[string]interface{}{
		"business_name": "Test Company",
		"address":       "123 Test St",
		"phone":         "+1-555-123-4567",
		"email":         "test@company.com",
		"website":       "https://testcompany.com",
	}

	score, err := scorer.AssessDataQuality(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, score)

	// Verify score consistency
	assert.Greater(t, score.OverallScore, 0.0)
	assert.LessOrEqual(t, score.OverallScore, 1.0)
	assert.NotEmpty(t, score.QualityLevel)
	assert.Contains(t, []string{"excellent", "good", "fair", "poor", "critical"}, score.QualityLevel)

	// Verify dimensions are populated
	assert.Greater(t, score.Dimensions.Completeness.OverallCompleteness, 0.0)
	assert.Greater(t, score.Dimensions.Accuracy.OverallAccuracy, 0.0)
	assert.Greater(t, score.Dimensions.Consistency.OverallConsistency, 0.0)
	assert.Greater(t, score.Dimensions.Validity.OverallValidity, 0.0)
	assert.Greater(t, score.Dimensions.Uniqueness.OverallUniqueness, 0.0)

	// Verify issues and recommendations are generated
	assert.NotNil(t, score.Issues)
	assert.NotNil(t, score.Recommendations)

	// Verify metadata is populated
	assert.Greater(t, score.Metadata.AssessmentDuration, time.Duration(0))
	assert.NotEmpty(t, score.Metadata.Thresholds)
	assert.Len(t, score.Metadata.QualityRules, 2) // Both rules are enabled

	// Verify trends are generated
	assert.NotEmpty(t, score.Trends.OverallTrend)
	assert.NotEmpty(t, score.Trends.DimensionTrends)

	// Verify validation results
	assert.Greater(t, score.ValidationResults.TotalValidations, 0)
	assert.GreaterOrEqual(t, score.ValidationResults.ValidationRate, 0.0)
	assert.LessOrEqual(t, score.ValidationResults.ValidationRate, 1.0)
	assert.NotEmpty(t, score.ValidationResults.FieldResults)
	assert.NotEmpty(t, score.ValidationResults.RuleResults)

	// Verify that the overall score makes sense given the dimensions
	expectedScore := (score.Dimensions.Completeness.OverallCompleteness*0.20 +
		score.Dimensions.Accuracy.OverallAccuracy*0.25 +
		score.Dimensions.Consistency.OverallConsistency*0.20 +
		score.Dimensions.Validity.OverallValidity*0.20 +
		score.Dimensions.Uniqueness.OverallUniqueness*0.15)
	assert.InDelta(t, expectedScore, score.OverallScore, 0.01)
}
