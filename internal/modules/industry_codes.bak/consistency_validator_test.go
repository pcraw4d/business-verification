package industry_codes

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestConsistencyValidator(t *testing.T) *ConsistencyValidator {
	logger := zap.NewNop()
	db := &IndustryCodeDatabase{} // Mock database
	return NewConsistencyValidator(db, logger)
}

func TestConsistencyValidator_ValidateConsistency(t *testing.T) {
	validator := setupTestConsistencyValidator(t)

	config := ConsistencyValidationConfig{
		ValidationMode:                "comprehensive",
		EnableFieldConsistency:        true,
		EnableCrossFieldConsistency:   true,
		EnableFormatConsistency:       true,
		EnableValueConsistency:        true,
		EnableBusinessRuleConsistency: true,
		ConsistencyThresholds: ConsistencyThresholds{
			OverallConsistencyMin:      0.8,
			FieldConsistencyMin:        0.8,
			CrossFieldConsistencyMin:   0.8,
			FormatConsistencyMin:       0.8,
			ValueConsistencyMin:        0.8,
			BusinessRuleConsistencyMin: 0.8,
		},
		ValidationRules: []ConsistencyValidationRule{
			{
				RuleID:       "rule_001",
				RuleName:     "Test Rule",
				RuleType:     "format",
				Description:  "Test validation rule",
				TargetFields: []string{"company_name"},
				Threshold:    0.8,
				Operator:     ">=",
				Severity:     "medium",
				IsEnabled:    true,
				IsCritical:   false,
			},
		},
	}

	ctx := context.Background()
	result, err := validator.ValidateConsistency(ctx, config)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	assert.True(t, result.GeneratedAt.After(time.Now().Add(-time.Second)))
	assert.Greater(t, result.OverallConsistency, 0.0)
	assert.LessOrEqual(t, result.OverallConsistency, 1.0)
	assert.NotEmpty(t, result.ConsistencyLevel)
	assert.Equal(t, config, result.Configuration)

	// Verify field consistency analysis
	assert.NotEmpty(t, result.FieldConsistency)
	for fieldName, analysis := range result.FieldConsistency {
		assert.Equal(t, fieldName, analysis.FieldName)
		assert.Greater(t, analysis.ConsistencyScore, 0.0)
		assert.LessOrEqual(t, analysis.ConsistencyScore, 1.0)
		assert.Greater(t, analysis.TotalRecords, 0)
		assert.GreaterOrEqual(t, analysis.ConsistentRecords, 0)
		assert.GreaterOrEqual(t, analysis.InconsistentRecords, 0)
		assert.Equal(t, analysis.TotalRecords, analysis.ConsistentRecords+analysis.InconsistentRecords)
	}

	// Verify cross-field consistency analysis
	assert.Greater(t, result.CrossFieldConsistency.OverallCrossFieldConsistency, 0.0)
	assert.LessOrEqual(t, result.CrossFieldConsistency.OverallCrossFieldConsistency, 1.0)
	assert.NotEmpty(t, result.CrossFieldConsistency.FieldPairAnalysis)

	// Verify format consistency analysis
	assert.Greater(t, result.FormatConsistency.OverallFormatConsistency, 0.0)
	assert.LessOrEqual(t, result.FormatConsistency.OverallFormatConsistency, 1.0)
	assert.NotEmpty(t, result.FormatConsistency.FieldFormatConsistency)

	// Verify value consistency analysis
	assert.Greater(t, result.ValueConsistency.OverallValueConsistency, 0.0)
	assert.LessOrEqual(t, result.ValueConsistency.OverallValueConsistency, 1.0)
	assert.NotEmpty(t, result.ValueConsistency.ValueDistribution)

	// Verify business rule consistency analysis
	assert.Greater(t, result.BusinessRuleConsistency.OverallBusinessRuleConsistency, 0.0)
	assert.LessOrEqual(t, result.BusinessRuleConsistency.OverallBusinessRuleConsistency, 1.0)
	assert.NotEmpty(t, result.BusinessRuleConsistency.RuleCompliance)

	// Verify validation report
	assert.NotEmpty(t, result.ValidationReport.ValidationID)
	assert.Equal(t, result.ID, result.ValidationReport.ValidationID)
	assert.Greater(t, result.ValidationReport.OverallScore, 0.0)
	assert.LessOrEqual(t, result.ValidationReport.OverallScore, 1.0)
	assert.NotEmpty(t, result.ValidationReport.RuleResults)

	// Verify recommendations
	assert.NotNil(t, result.Recommendations)

	// Verify metadata
	assert.NotEmpty(t, result.Metadata.ValidationVersion)
	assert.NotEmpty(t, result.Metadata.ValidationEngine)
	assert.NotEmpty(t, result.Metadata.ConfigurationHash)
	assert.Greater(t, result.Metadata.ProcessingTime, time.Duration(0))
	assert.Greater(t, result.Metadata.MemoryUsage, int64(0))
	assert.Greater(t, result.Metadata.CPUUsage, 0.0)
}

func TestConsistencyValidator_AnalyzeFieldConsistency(t *testing.T) {
	validator := setupTestConsistencyValidator(t)
	ctx := context.Background()

	fieldConsistency, err := validator.analyzeFieldConsistency(ctx)

	require.NoError(t, err)
	assert.NotEmpty(t, fieldConsistency)

	expectedFields := []string{"company_name", "industry_code", "business_type", "registration_number"}
	for _, expectedField := range expectedFields {
		analysis, exists := fieldConsistency[expectedField]
		assert.True(t, exists, "Field %s should exist in analysis", expectedField)
		assert.Equal(t, expectedField, analysis.FieldName)
		assert.Greater(t, analysis.ConsistencyScore, 0.0)
		assert.LessOrEqual(t, analysis.ConsistencyScore, 1.0)
		assert.Greater(t, analysis.TotalRecords, 0)
		assert.NotEmpty(t, analysis.Issues)
		assert.NotEmpty(t, analysis.Patterns)
		assert.Equal(t, "validated", analysis.ValidationStatus.Status)
	}
}

func TestConsistencyValidator_AnalyzeCrossFieldConsistency(t *testing.T) {
	validator := setupTestConsistencyValidator(t)
	ctx := context.Background()

	analysis, err := validator.analyzeCrossFieldConsistency(ctx)

	require.NoError(t, err)
	assert.Greater(t, analysis.OverallCrossFieldConsistency, 0.0)
	assert.LessOrEqual(t, analysis.OverallCrossFieldConsistency, 1.0)
	assert.NotEmpty(t, analysis.FieldPairAnalysis)
	assert.Greater(t, analysis.LogicalConsistency, 0.0)
	assert.Greater(t, analysis.ReferentialConsistency, 0.0)
	assert.Greater(t, analysis.BusinessLogicConsistency, 0.0)
	assert.NotEmpty(t, analysis.CrossFieldIssues)
	assert.NotEmpty(t, analysis.ConsistencyRules)

	// Verify field pair analysis
	for pairKey, pairAnalysis := range analysis.FieldPairAnalysis {
		assert.NotEmpty(t, pairKey)
		assert.NotEmpty(t, pairAnalysis.Field1)
		assert.NotEmpty(t, pairAnalysis.Field2)
		assert.Greater(t, pairAnalysis.ConsistencyScore, 0.0)
		assert.LessOrEqual(t, pairAnalysis.ConsistencyScore, 1.0)
		assert.GreaterOrEqual(t, pairAnalysis.Correlation, -1.0)
		assert.LessOrEqual(t, pairAnalysis.Correlation, 1.0)
	}
}

func TestConsistencyValidator_AnalyzeFormatConsistency(t *testing.T) {
	validator := setupTestConsistencyValidator(t)
	ctx := context.Background()

	analysis, err := validator.analyzeFormatConsistency(ctx)

	require.NoError(t, err)
	assert.Greater(t, analysis.OverallFormatConsistency, 0.0)
	assert.LessOrEqual(t, analysis.OverallFormatConsistency, 1.0)
	assert.NotEmpty(t, analysis.FieldFormatConsistency)
	assert.NotEmpty(t, analysis.FormatPatterns)
	assert.NotEmpty(t, analysis.FormatViolations)
	assert.NotEmpty(t, analysis.StandardCompliance)

	// Verify field format consistency
	for fieldName, score := range analysis.FieldFormatConsistency {
		assert.NotEmpty(t, fieldName)
		assert.Greater(t, score, 0.0)
		assert.LessOrEqual(t, score, 1.0)
	}

	// Verify format patterns
	for _, pattern := range analysis.FormatPatterns {
		assert.NotEmpty(t, pattern.Pattern)
		assert.NotEmpty(t, pattern.Description)
		assert.Greater(t, pattern.Frequency, 0.0)
		assert.LessOrEqual(t, pattern.Frequency, 1.0)
		assert.Greater(t, pattern.Compliance, 0.0)
		assert.LessOrEqual(t, pattern.Compliance, 1.0)
	}
}

func TestConsistencyValidator_AnalyzeValueConsistency(t *testing.T) {
	validator := setupTestConsistencyValidator(t)
	ctx := context.Background()

	analysis, err := validator.analyzeValueConsistency(ctx)

	require.NoError(t, err)
	assert.Greater(t, analysis.OverallValueConsistency, 0.0)
	assert.LessOrEqual(t, analysis.OverallValueConsistency, 1.0)
	assert.Greater(t, analysis.ValueRangeConsistency, 0.0)
	assert.LessOrEqual(t, analysis.ValueRangeConsistency, 1.0)
	assert.NotEmpty(t, analysis.ValueDistribution)
	assert.NotEmpty(t, analysis.OutlierAnalysis)

	// Verify value distribution
	var totalDistribution float64
	for category, percentage := range analysis.ValueDistribution {
		assert.NotEmpty(t, category)
		assert.GreaterOrEqual(t, percentage, 0.0)
		assert.LessOrEqual(t, percentage, 1.0)
		totalDistribution += percentage
	}
	assert.LessOrEqual(t, totalDistribution, 1.0) // Should not exceed 100%

	// Verify outlier analysis
	for _, outlier := range analysis.OutlierAnalysis {
		assert.NotEmpty(t, outlier.FieldName)
		assert.NotEmpty(t, outlier.Value)
		assert.NotEmpty(t, outlier.OutlierType)
		assert.NotEmpty(t, outlier.Severity)
		assert.GreaterOrEqual(t, outlier.StatisticalScore, 0.0)
		assert.LessOrEqual(t, outlier.StatisticalScore, 1.0)
	}

	// Verify statistical consistency
	stats := analysis.StatisticalConsistency
	assert.GreaterOrEqual(t, stats.Mean, 0.0)
	assert.GreaterOrEqual(t, stats.Median, 0.0)
	assert.GreaterOrEqual(t, stats.StandardDeviation, 0.0)
}

func TestConsistencyValidator_AnalyzeBusinessRuleConsistency(t *testing.T) {
	validator := setupTestConsistencyValidator(t)
	ctx := context.Background()

	analysis, err := validator.analyzeBusinessRuleConsistency(ctx)

	require.NoError(t, err)
	assert.Greater(t, analysis.OverallBusinessRuleConsistency, 0.0)
	assert.LessOrEqual(t, analysis.OverallBusinessRuleConsistency, 1.0)
	assert.NotEmpty(t, analysis.RuleCompliance)
	assert.NotEmpty(t, analysis.ViolatedRules)
	assert.NotEmpty(t, analysis.BusinessLogicIssues)
	assert.Greater(t, analysis.ComplianceScore, 0.0)
	assert.LessOrEqual(t, analysis.ComplianceScore, 1.0)

	// Verify rule compliance
	for ruleName, compliance := range analysis.RuleCompliance {
		assert.NotEmpty(t, ruleName)
		assert.GreaterOrEqual(t, compliance, 0.0)
		assert.LessOrEqual(t, compliance, 1.0)
	}

	// Verify violated rules
	for _, violation := range analysis.ViolatedRules {
		assert.NotEmpty(t, violation.RuleID)
		assert.NotEmpty(t, violation.RuleName)
		assert.NotEmpty(t, violation.Description)
		assert.GreaterOrEqual(t, violation.AffectedRecords, 0)
		assert.NotEmpty(t, violation.Severity)
	}

	// Verify business logic issues
	for _, issue := range analysis.BusinessLogicIssues {
		assert.NotEmpty(t, issue.IssueID)
		assert.NotEmpty(t, issue.IssueType)
		assert.NotEmpty(t, issue.Description)
		assert.NotEmpty(t, issue.AffectedFields)
		assert.NotEmpty(t, issue.Severity)
	}
}

func TestConsistencyValidator_CalculateOverallConsistency(t *testing.T) {
	validator := setupTestConsistencyValidator(t)

	// Test with all components
	result := &ConsistencyValidationResult{
		FieldConsistency: map[string]FieldConsistencyAnalysis{
			"field1": {ConsistencyScore: 0.9},
			"field2": {ConsistencyScore: 0.8},
		},
		CrossFieldConsistency: CrossFieldConsistencyAnalysis{
			OverallCrossFieldConsistency: 0.85,
		},
		FormatConsistency: FormatConsistencyAnalysis{
			OverallFormatConsistency: 0.92,
		},
		ValueConsistency: ValueConsistencyAnalysis{
			OverallValueConsistency: 0.88,
		},
		BusinessRuleConsistency: BusinessRuleConsistencyAnalysis{
			OverallBusinessRuleConsistency: 0.90,
		},
	}

	score := validator.calculateOverallConsistency(result)
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)

	// Test with empty result
	emptyResult := &ConsistencyValidationResult{}
	emptyScore := validator.calculateOverallConsistency(emptyResult)
	assert.Equal(t, 0.0, emptyScore)
}

func TestConsistencyValidator_DetermineConsistencyLevel(t *testing.T) {
	validator := setupTestConsistencyValidator(t)

	tests := []struct {
		score float64
		level string
	}{
		{0.98, "excellent"},
		{0.90, "good"},
		{0.85, "good"},
		{0.78, "fair"},
		{0.75, "fair"},
		{0.65, "poor"},
		{0.60, "poor"},
		{0.30, "critical"},
	}

	for _, tt := range tests {
		level := validator.determineConsistencyLevel(tt.score)
		assert.Equal(t, tt.level, level, "Score %.2f should be level %s", tt.score, tt.level)
	}
}

func TestConsistencyValidator_ExecuteValidationRules(t *testing.T) {
	validator := setupTestConsistencyValidator(t)
	ctx := context.Background()

	rules := []ConsistencyValidationRule{
		{
			RuleID:       "rule_001",
			RuleName:     "Test Rule 1",
			RuleType:     "format",
			Description:  "Test validation rule 1",
			TargetFields: []string{"company_name"},
			Threshold:    0.8,
			Operator:     ">=",
			Severity:     "medium",
			IsEnabled:    true,
			IsCritical:   false,
		},
		{
			RuleID:       "rule_002",
			RuleName:     "Test Rule 2",
			RuleType:     "business",
			Description:  "Test validation rule 2",
			TargetFields: []string{"industry_code"},
			Threshold:    0.9,
			Operator:     ">=",
			Severity:     "high",
			IsEnabled:    true,
			IsCritical:   true,
		},
	}

	result := &ConsistencyValidationResult{
		OverallConsistency: 0.85,
	}

	ruleResults, err := validator.executeValidationRules(ctx, rules, result)

	require.NoError(t, err)
	assert.Len(t, ruleResults, len(rules))

	for i, ruleResult := range ruleResults {
		assert.Equal(t, rules[i].RuleID, ruleResult.RuleID)
		assert.Equal(t, rules[i].RuleName, ruleResult.RuleName)
		assert.NotEmpty(t, ruleResult.Status)
		assert.GreaterOrEqual(t, ruleResult.Score, 0.0)
		assert.LessOrEqual(t, ruleResult.Score, 1.0)
		assert.NotEmpty(t, ruleResult.Message)
		assert.True(t, ruleResult.ExecutedAt.After(time.Now().Add(-time.Second)))
	}
}

func TestConsistencyValidator_GenerateValidationReport(t *testing.T) {
	validator := setupTestConsistencyValidator(t)
	ctx := context.Background()

	result := &ConsistencyValidationResult{
		ID:                 "test_validation_123",
		OverallConsistency: 0.87,
	}

	ruleResults := []ConsistencyRuleValidationResult{
		{
			RuleID:     "rule_001",
			RuleName:   "Test Rule",
			Status:     "passed",
			Score:      0.85,
			Message:    "Rule passed",
			ExecutedAt: time.Now(),
		},
	}

	report, err := validator.generateValidationReport(ctx, result, ruleResults)

	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, result.ID, report.ValidationID)
	assert.True(t, report.ValidationTimestamp.After(time.Now().Add(-time.Second)))
	assert.NotEmpty(t, report.ValidationStatus)
	assert.Equal(t, result.OverallConsistency, report.OverallScore)
	assert.NotEmpty(t, report.ScoreBreakdown)
	assert.Len(t, report.RuleResults, len(ruleResults))
	assert.NotEmpty(t, report.CriticalIssues)
	assert.NotEmpty(t, report.Warnings)
	assert.NotEmpty(t, report.QualityGates)
	assert.NotEmpty(t, report.ValidationMetrics)

	// Verify score breakdown
	breakdown := report.ScoreBreakdown
	assert.Greater(t, breakdown.FieldConsistencyScore, 0.0)
	assert.LessOrEqual(t, breakdown.FieldConsistencyScore, 1.0)
	assert.Greater(t, breakdown.CrossFieldConsistencyScore, 0.0)
	assert.LessOrEqual(t, breakdown.CrossFieldConsistencyScore, 1.0)
	assert.Greater(t, breakdown.FormatConsistencyScore, 0.0)
	assert.LessOrEqual(t, breakdown.FormatConsistencyScore, 1.0)
	assert.Greater(t, breakdown.ValueConsistencyScore, 0.0)
	assert.LessOrEqual(t, breakdown.ValueConsistencyScore, 1.0)
	assert.Greater(t, breakdown.BusinessRuleConsistencyScore, 0.0)
	assert.LessOrEqual(t, breakdown.BusinessRuleConsistencyScore, 1.0)

	// Verify validation metrics
	metrics := report.ValidationMetrics
	assert.Greater(t, metrics.TotalRecordsProcessed, 0)
	assert.GreaterOrEqual(t, metrics.RecordsWithIssues, 0)
	assert.GreaterOrEqual(t, metrics.CriticalIssues, 0)
	assert.GreaterOrEqual(t, metrics.Warnings, 0)
	assert.Greater(t, metrics.AverageProcessingTime, 0.0)
}

func TestConsistencyValidator_GenerateRecommendations(t *testing.T) {
	validator := setupTestConsistencyValidator(t)
	ctx := context.Background()

	// Test with low consistency score
	lowResult := &ConsistencyValidationResult{
		OverallConsistency: 0.75, // Below 0.9 threshold
	}

	recommendations, err := validator.generateRecommendations(ctx, lowResult)

	require.NoError(t, err)
	assert.NotEmpty(t, recommendations)

	for _, rec := range recommendations {
		assert.NotEmpty(t, rec.RecommendationID)
		assert.NotEmpty(t, rec.Type)
		assert.NotEmpty(t, rec.Priority)
		assert.NotEmpty(t, rec.Title)
		assert.NotEmpty(t, rec.Description)
		assert.NotEmpty(t, rec.AffectedFields)
		assert.NotEmpty(t, rec.ImpactAssessment)
		assert.NotEmpty(t, rec.ImplementationPlan)
		assert.NotEmpty(t, rec.ROIAnalysis)
		assert.NotEmpty(t, rec.RiskAssessment)
		assert.NotEmpty(t, rec.SuccessMetrics)
		assert.True(t, rec.CreatedAt.After(time.Now().Add(-time.Second)))
		assert.True(t, rec.EstimatedCompletion.After(rec.CreatedAt))
	}

	// Test with high consistency score
	highResult := &ConsistencyValidationResult{
		OverallConsistency: 0.95, // Above 0.9 threshold
	}

	highRecommendations, err := validator.generateRecommendations(ctx, highResult)

	require.NoError(t, err)
	assert.Empty(t, highRecommendations) // Should not generate recommendations for high scores
}

func TestConsistencyValidator_Integration(t *testing.T) {
	validator := setupTestConsistencyValidator(t)

	config := ConsistencyValidationConfig{
		ValidationMode:                "comprehensive",
		EnableFieldConsistency:        true,
		EnableCrossFieldConsistency:   true,
		EnableFormatConsistency:       true,
		EnableValueConsistency:        true,
		EnableBusinessRuleConsistency: true,
		ConsistencyThresholds: ConsistencyThresholds{
			OverallConsistencyMin:      0.7,
			FieldConsistencyMin:        0.7,
			CrossFieldConsistencyMin:   0.7,
			FormatConsistencyMin:       0.7,
			ValueConsistencyMin:        0.7,
			BusinessRuleConsistencyMin: 0.7,
		},
	}

	ctx := context.Background()
	result, err := validator.ValidateConsistency(ctx, config)

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify overall consistency is reasonable
	assert.Greater(t, result.OverallConsistency, 0.7)
	assert.LessOrEqual(t, result.OverallConsistency, 1.0)

	// Verify all components are present
	assert.NotEmpty(t, result.FieldConsistency)
	assert.Greater(t, result.CrossFieldConsistency.OverallCrossFieldConsistency, 0.0)
	assert.Greater(t, result.FormatConsistency.OverallFormatConsistency, 0.0)
	assert.Greater(t, result.ValueConsistency.OverallValueConsistency, 0.0)
	assert.Greater(t, result.BusinessRuleConsistency.OverallBusinessRuleConsistency, 0.0)

	// Verify validation report
	assert.Equal(t, result.ID, result.ValidationReport.ValidationID)
	assert.Equal(t, result.OverallConsistency, result.ValidationReport.OverallScore)

	// Verify metadata
	assert.Greater(t, result.Metadata.ProcessingTime, time.Duration(0))
	assert.NotEmpty(t, result.Metadata.ValidationEngine)
	assert.NotEmpty(t, result.Metadata.ConfigurationHash)
}
