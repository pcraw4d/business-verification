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

func setupTestCompletenessValidator(t *testing.T) (*CompletenessValidator, *sql.DB) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	logger := zap.NewNop()
	industryDB := NewIndustryCodeDatabase(db, logger)
	validator := NewCompletenessValidator(industryDB, logger)

	return validator, db
}

func TestCompletenessValidator_ValidateCompleteness(t *testing.T) {
	validator, db := setupTestCompletenessValidator(t)
	defer db.Close()

	ctx := context.Background()
	config := &CompletenessValidationConfig{
		ValidationMode:         "normal",
		EnableFieldAnalysis:    true,
		EnableRecordAnalysis:   true,
		EnablePatternAnalysis:  true,
		EnableTrendAnalysis:    true,
		EnableAnomalyDetection: true,
		RequiredFields:         []string{"business_name", "address"},
		OptionalFields:         []string{"website", "phone", "email"},
		CompletenessThresholds: CompletenessThresholds{
			OverallCompletenessMin: 0.80,
			FieldCompletenessMin:   0.70,
			RecordCompletenessMin:  0.75,
			RequiredFieldsMin:      0.95,
		},
		ValidationRules: []CompletenessValidationRule{
			{
				RuleID:       "rule_001",
				RuleName:     "Required Field Completeness",
				RuleType:     "threshold",
				Description:  "Required fields must meet minimum completeness",
				TargetFields: []string{"business_name", "address"},
				Threshold:    0.95,
				Operator:     ">=",
				Severity:     "critical",
				IsEnabled:    true,
				IsCritical:   true,
			},
		},
		QualityGates: []QualityGate{
			{
				GateID:          "gate_001",
				GateName:        "Overall Completeness Gate",
				Description:     "Overall completeness must exceed threshold",
				MetricName:      "overall_completeness",
				ThresholdValue:  0.80,
				Operator:        ">=",
				IsCritical:      true,
				BlocksExecution: false,
				IsEnabled:       true,
			},
		},
	}

	// Mock data
	data := map[string]interface{}{
		"business_name": "Test Company",
		"address":       "123 Test St",
		"phone":         "+1-555-123-4567",
		"email":         "test@company.com",
		"website":       "",
	}

	result, err := validator.ValidateCompleteness(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Test basic structure
	assert.NotEmpty(t, result.ID)
	assert.WithinDuration(t, time.Now(), result.GeneratedAt, 5*time.Second)
	assert.Greater(t, result.OverallCompleteness, 0.0)
	assert.LessOrEqual(t, result.OverallCompleteness, 1.0)
	assert.NotEmpty(t, result.CompletenessLevel)
	assert.Contains(t, []string{"excellent", "good", "fair", "poor", "critical"}, result.CompletenessLevel)

	// Test field analysis
	assert.NotNil(t, result.FieldAnalysis)
	assert.NotEmpty(t, result.FieldAnalysis)

	for fieldName, fieldAnalysis := range result.FieldAnalysis {
		assert.NotEmpty(t, fieldName)
		assert.Equal(t, fieldName, fieldAnalysis.FieldName)
		assert.GreaterOrEqual(t, fieldAnalysis.Completeness, 0.0)
		assert.LessOrEqual(t, fieldAnalysis.Completeness, 1.0)
		assert.Greater(t, fieldAnalysis.TotalRecords, 0)
		assert.GreaterOrEqual(t, fieldAnalysis.PresentRecords, 0)
		assert.GreaterOrEqual(t, fieldAnalysis.MissingRecords, 0)
		assert.Equal(t, fieldAnalysis.TotalRecords, fieldAnalysis.PresentRecords+fieldAnalysis.MissingRecords)
		assert.GreaterOrEqual(t, fieldAnalysis.QualityScore, 0.0)
		assert.LessOrEqual(t, fieldAnalysis.QualityScore, 1.0)
		assert.NotNil(t, fieldAnalysis.MissingPatterns)
		assert.NotNil(t, fieldAnalysis.ValidationStatus)
	}

	// Test record analysis
	assert.NotNil(t, result.RecordAnalysis)
	assert.Greater(t, result.RecordAnalysis.TotalRecords, 0)
	assert.GreaterOrEqual(t, result.RecordAnalysis.CompleteRecords, 0)
	assert.GreaterOrEqual(t, result.RecordAnalysis.PartialRecords, 0)
	assert.GreaterOrEqual(t, result.RecordAnalysis.IncompleteRecords, 0)
	assert.GreaterOrEqual(t, result.RecordAnalysis.RecordCompleteness, 0.0)
	assert.LessOrEqual(t, result.RecordAnalysis.RecordCompleteness, 1.0)
	assert.NotEmpty(t, result.RecordAnalysis.CompletenessDistribution)

	// Test pattern analysis
	assert.NotNil(t, result.PatternAnalysis)
	assert.NotNil(t, result.PatternAnalysis.OverallPatterns)
	assert.NotNil(t, result.PatternAnalysis.PatternSummary)
	assert.GreaterOrEqual(t, result.PatternAnalysis.PatternSummary.TotalPatterns, 0)
	assert.GreaterOrEqual(t, result.PatternAnalysis.PatternSummary.RandomMissingness, 0.0)
	assert.LessOrEqual(t, result.PatternAnalysis.PatternSummary.RandomMissingness, 1.0)
	assert.GreaterOrEqual(t, result.PatternAnalysis.PatternSummary.SystematicMissingness, 0.0)
	assert.LessOrEqual(t, result.PatternAnalysis.PatternSummary.SystematicMissingness, 1.0)

	// Test trend analysis
	assert.NotNil(t, result.TrendAnalysis)
	assert.NotEmpty(t, result.TrendAnalysis.TrendDirection)
	assert.Contains(t, []string{"improving", "declining", "stable"}, result.TrendAnalysis.TrendDirection)
	assert.GreaterOrEqual(t, result.TrendAnalysis.TrendStrength, 0.0)
	assert.LessOrEqual(t, result.TrendAnalysis.TrendStrength, 1.0)
	assert.GreaterOrEqual(t, result.TrendAnalysis.TrendConfidence, 0.0)
	assert.LessOrEqual(t, result.TrendAnalysis.TrendConfidence, 1.0)

	// Test validation report
	assert.NotNil(t, result.ValidationReport)
	assert.NotEmpty(t, result.ValidationReport.ValidationID)
	assert.NotEmpty(t, result.ValidationReport.ValidationStatus)
	assert.GreaterOrEqual(t, result.ValidationReport.OverallScore, 0.0)
	assert.LessOrEqual(t, result.ValidationReport.OverallScore, 1.0)
	assert.NotNil(t, result.ValidationReport.ValidationMetrics)
	assert.Greater(t, result.ValidationReport.ValidationMetrics.ValidationThroughput, 0.0)

	// Test recommendations
	assert.NotNil(t, result.Recommendations)

	// Test configuration
	assert.Equal(t, config.ValidationMode, result.Configuration.ValidationMode)
	assert.Equal(t, config.EnableFieldAnalysis, result.Configuration.EnableFieldAnalysis)

	// Test metadata
	assert.NotNil(t, result.Metadata)
	assert.Greater(t, result.Metadata.ProcessingTime, time.Duration(0))
	assert.NotEmpty(t, result.Metadata.ValidationVersion)
	assert.NotEmpty(t, result.Metadata.ValidationEngine)
}

func TestCompletenessValidator_ValidateCompleteness_FieldAnalysisDisabled(t *testing.T) {
	validator, db := setupTestCompletenessValidator(t)
	defer db.Close()

	ctx := context.Background()
	config := &CompletenessValidationConfig{
		ValidationMode:        "normal",
		EnableFieldAnalysis:   false,
		EnableRecordAnalysis:  true,
		EnablePatternAnalysis: true,
		EnableTrendAnalysis:   true,
	}

	data := map[string]interface{}{
		"business_name": "Test Company",
	}

	result, err := validator.ValidateCompleteness(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Field analysis should be empty when disabled
	assert.Empty(t, result.FieldAnalysis)

	// Other analyses should still be present
	assert.NotNil(t, result.RecordAnalysis)
	assert.NotNil(t, result.PatternAnalysis)
	assert.NotNil(t, result.TrendAnalysis)
}

func TestCompletenessValidator_AnalyzeFieldCompleteness(t *testing.T) {
	validator, db := setupTestCompletenessValidator(t)
	defer db.Close()

	ctx := context.Background()
	config := &CompletenessValidationConfig{
		RequiredFields: []string{"business_name"},
		OptionalFields: []string{"website"},
	}

	data := map[string]interface{}{
		"business_name": "Test Company",
		"website":       "",
	}

	analysis, err := validator.analyzeFieldCompleteness(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, analysis)

	// Check business_name analysis
	businessNameAnalysis, exists := analysis["business_name"]
	assert.True(t, exists)
	assert.Equal(t, "business_name", businessNameAnalysis.FieldName)
	assert.True(t, businessNameAnalysis.IsRequired)
	assert.False(t, businessNameAnalysis.IsOptional)
	assert.Greater(t, businessNameAnalysis.Completeness, 0.0)
	assert.LessOrEqual(t, businessNameAnalysis.Completeness, 1.0)
	assert.NotNil(t, businessNameAnalysis.MissingPatterns)
	assert.NotNil(t, businessNameAnalysis.ValidationStatus)

	// Validate missing patterns
	for _, pattern := range businessNameAnalysis.MissingPatterns {
		assert.NotEmpty(t, pattern.PatternType)
		assert.Contains(t, []string{"random", "systematic", "conditional", "temporal"}, pattern.PatternType)
		assert.GreaterOrEqual(t, pattern.MissingPercentage, 0.0)
		assert.LessOrEqual(t, pattern.MissingPercentage, 100.0)
		assert.NotEmpty(t, pattern.Description)
		assert.NotEmpty(t, pattern.Impact)
		assert.GreaterOrEqual(t, pattern.Predictability, 0.0)
		assert.LessOrEqual(t, pattern.Predictability, 1.0)
	}

	// Check website analysis
	websiteAnalysis, exists := analysis["website"]
	assert.True(t, exists)
	assert.Equal(t, "website", websiteAnalysis.FieldName)
	assert.False(t, websiteAnalysis.IsRequired)
	assert.True(t, websiteAnalysis.IsOptional)
	assert.GreaterOrEqual(t, websiteAnalysis.Completeness, 0.0)
	assert.LessOrEqual(t, websiteAnalysis.Completeness, 1.0)
}

func TestCompletenessValidator_AnalyzeRecordCompleteness(t *testing.T) {
	validator, db := setupTestCompletenessValidator(t)
	defer db.Close()

	ctx := context.Background()
	config := &CompletenessValidationConfig{}

	data := map[string]interface{}{
		"business_name": "Test Company",
		"address":       "123 Test St",
	}

	analysis, err := validator.analyzeRecordCompleteness(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, analysis)

	// Test record counts
	assert.Greater(t, analysis.TotalRecords, 0)
	assert.GreaterOrEqual(t, analysis.CompleteRecords, 0)
	assert.GreaterOrEqual(t, analysis.PartialRecords, 0)
	assert.GreaterOrEqual(t, analysis.IncompleteRecords, 0)
	assert.GreaterOrEqual(t, analysis.EmptyRecords, 0)
	assert.Equal(t, analysis.TotalRecords,
		analysis.CompleteRecords+analysis.PartialRecords+analysis.IncompleteRecords+analysis.EmptyRecords)

	// Test completeness score
	assert.GreaterOrEqual(t, analysis.RecordCompleteness, 0.0)
	assert.LessOrEqual(t, analysis.RecordCompleteness, 1.0)

	// Test completeness distribution
	assert.NotEmpty(t, analysis.CompletenessDistribution)
	totalDistributionRecords := 0
	for _, count := range analysis.CompletenessDistribution {
		totalDistributionRecords += count
	}
	assert.Equal(t, analysis.TotalRecords, totalDistributionRecords)

	// Test record patterns
	assert.NotNil(t, analysis.RecordPatterns)
	for _, pattern := range analysis.RecordPatterns {
		assert.NotEmpty(t, pattern.PatternType)
		assert.GreaterOrEqual(t, pattern.RecordCount, 0)
		assert.GreaterOrEqual(t, pattern.Percentage, 0.0)
		assert.LessOrEqual(t, pattern.Percentage, 100.0)
		assert.GreaterOrEqual(t, pattern.AverageFields, 0.0)
		assert.NotEmpty(t, pattern.Description)
	}

	// Test completeness profile
	profile := analysis.CompletenessProfile
	assert.GreaterOrEqual(t, profile.AverageCompleteness, 0.0)
	assert.LessOrEqual(t, profile.AverageCompleteness, 1.0)
	assert.GreaterOrEqual(t, profile.MedianCompleteness, 0.0)
	assert.LessOrEqual(t, profile.MedianCompleteness, 1.0)
	assert.GreaterOrEqual(t, profile.StandardDeviation, 0.0)
	assert.GreaterOrEqual(t, profile.MinCompleteness, 0.0)
	assert.LessOrEqual(t, profile.MaxCompleteness, 1.0)
	assert.LessOrEqual(t, profile.MinCompleteness, profile.MaxCompleteness)
	assert.NotEmpty(t, profile.PercentileDistribution)
}

func TestCompletenessValidator_AnalyzeCompletenessPatterns(t *testing.T) {
	validator, db := setupTestCompletenessValidator(t)
	defer db.Close()

	ctx := context.Background()
	config := &CompletenessValidationConfig{}

	data := map[string]interface{}{
		"business_name": "Test Company",
		"website":       "",
	}

	analysis, err := validator.analyzeCompletenessPatterns(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, analysis)

	// Test overall patterns
	assert.NotNil(t, analysis.OverallPatterns)
	for _, pattern := range analysis.OverallPatterns {
		assert.NotEmpty(t, pattern.PatternName)
		assert.NotEmpty(t, pattern.PatternType)
		assert.Contains(t, []string{"random", "systematic", "clustered", "structured"}, pattern.PatternType)
		assert.NotEmpty(t, pattern.Description)
		assert.NotEmpty(t, pattern.AffectedFields)
		assert.GreaterOrEqual(t, pattern.AffectedRecords, 0)
		assert.GreaterOrEqual(t, pattern.MissingPercentage, 0.0)
		assert.LessOrEqual(t, pattern.MissingPercentage, 100.0)
		assert.GreaterOrEqual(t, pattern.Confidence, 0.0)
		assert.LessOrEqual(t, pattern.Confidence, 1.0)
		assert.GreaterOrEqual(t, pattern.StatisticalSignificance, 0.0)
		assert.LessOrEqual(t, pattern.StatisticalSignificance, 1.0)
		assert.NotEmpty(t, pattern.Impact)
		assert.NotNil(t, pattern.RecommendedActions)
	}

	// Test pattern summary
	summary := analysis.PatternSummary
	assert.GreaterOrEqual(t, summary.TotalPatterns, 0)
	assert.GreaterOrEqual(t, summary.CriticalPatterns, 0)
	assert.GreaterOrEqual(t, summary.SystemicIssues, 0)
	assert.GreaterOrEqual(t, summary.PredictablePatterns, 0)
	assert.GreaterOrEqual(t, summary.RandomMissingness, 0.0)
	assert.LessOrEqual(t, summary.RandomMissingness, 1.0)
	assert.GreaterOrEqual(t, summary.SystematicMissingness, 0.0)
	assert.LessOrEqual(t, summary.SystematicMissingness, 1.0)
	assert.NotEmpty(t, summary.PatternComplexity)
	assert.NotEmpty(t, summary.DataQualityImpact)
	assert.NotEmpty(t, summary.RecoveryPotential)
	assert.NotNil(t, summary.PriorityActions)

	// Test priority actions
	for _, action := range summary.PriorityActions {
		assert.NotEmpty(t, action.ActionType)
		assert.NotEmpty(t, action.Description)
		assert.NotEmpty(t, action.Priority)
		assert.Contains(t, []string{"critical", "high", "medium", "low"}, action.Priority)
		assert.NotEmpty(t, action.EstimatedImpact)
		assert.NotEmpty(t, action.EstimatedEffort)
		assert.GreaterOrEqual(t, action.ROIScore, 0.0)
		assert.LessOrEqual(t, action.ROIScore, 1.0)
		assert.NotNil(t, action.Dependencies)
		assert.NotEmpty(t, action.ExpectedOutcome)
		assert.NotNil(t, action.SuccessCriteria)
	}
}

func TestCompletenessValidator_AnalyzeTrends(t *testing.T) {
	validator, db := setupTestCompletenessValidator(t)
	defer db.Close()

	ctx := context.Background()
	config := &CompletenessValidationConfig{}

	data := map[string]interface{}{
		"business_name": "Test Company",
	}

	analysis, err := validator.analyzeTrends(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, analysis)

	// Test trend direction and strength
	assert.NotEmpty(t, analysis.TrendDirection)
	assert.Contains(t, []string{"improving", "declining", "stable"}, analysis.TrendDirection)
	assert.GreaterOrEqual(t, analysis.TrendStrength, 0.0)
	assert.LessOrEqual(t, analysis.TrendStrength, 1.0)
	assert.GreaterOrEqual(t, analysis.TrendConfidence, 0.0)
	assert.LessOrEqual(t, analysis.TrendConfidence, 1.0)

	// Test seasonal patterns
	assert.NotNil(t, analysis.SeasonalPatterns)
	for _, pattern := range analysis.SeasonalPatterns {
		assert.NotEmpty(t, pattern.Season)
		assert.GreaterOrEqual(t, pattern.AverageCompleteness, 0.0)
		assert.LessOrEqual(t, pattern.AverageCompleteness, 1.0)
		assert.GreaterOrEqual(t, pattern.TypicalVariation, 0.0)
		assert.GreaterOrEqual(t, pattern.Confidence, 0.0)
		assert.LessOrEqual(t, pattern.Confidence, 1.0)
		assert.NotEmpty(t, pattern.BusinessRationale)
	}

	// Test trend analysis metrics
	metrics := analysis.TrendAnalysisMetrics
	assert.GreaterOrEqual(t, metrics.R2Score, 0.0)
	assert.LessOrEqual(t, metrics.R2Score, 1.0)
	assert.GreaterOrEqual(t, metrics.MeanAbsoluteError, 0.0)
	assert.GreaterOrEqual(t, metrics.RootMeanSquareError, 0.0)
	assert.GreaterOrEqual(t, metrics.TrendSignificance, 0.0)
	assert.LessOrEqual(t, metrics.TrendSignificance, 1.0)
	assert.NotEmpty(t, metrics.StationarityTest)
	assert.GreaterOrEqual(t, metrics.SeasonalityStrength, 0.0)
	assert.LessOrEqual(t, metrics.SeasonalityStrength, 1.0)
	assert.GreaterOrEqual(t, metrics.NoiseToSignalRatio, 0.0)

	// Test forecasting
	forecast := analysis.Forecasting
	assert.Greater(t, forecast.ForecastHorizon, time.Duration(0))
	assert.GreaterOrEqual(t, forecast.ForecastConfidence, 0.0)
	assert.LessOrEqual(t, forecast.ForecastConfidence, 1.0)
	assert.NotEmpty(t, forecast.ForecastModel)
	assert.GreaterOrEqual(t, forecast.ModelAccuracy, 0.0)
	assert.LessOrEqual(t, forecast.ModelAccuracy, 1.0)
}

func TestCompletenessValidator_ExecuteValidationRules(t *testing.T) {
	validator, db := setupTestCompletenessValidator(t)
	defer db.Close()

	ctx := context.Background()
	config := &CompletenessValidationConfig{
		ValidationRules: []CompletenessValidationRule{
			{
				RuleID:       "rule_001",
				RuleName:     "Required Field Test",
				RuleType:     "threshold",
				Description:  "Test rule for required fields",
				TargetFields: []string{"business_name"},
				Threshold:    0.90,
				Operator:     ">=",
				Severity:     "critical",
				IsEnabled:    true,
				IsCritical:   true,
				ErrorMessage: "Field completeness below threshold",
			},
		},
	}

	data := map[string]interface{}{
		"business_name": "Test Company",
	}

	result := &CompletenessValidationResult{
		ID: "test_validation",
	}

	report, err := validator.executeValidationRules(ctx, data, config, result)
	require.NoError(t, err)
	assert.NotNil(t, report)

	// Test validation report structure
	assert.Equal(t, result.ID, report.ValidationID)
	assert.WithinDuration(t, time.Now(), report.ValidationTimestamp, 5*time.Second)
	assert.NotEmpty(t, report.ValidationStatus)
	assert.GreaterOrEqual(t, report.OverallScore, 0.0)
	assert.LessOrEqual(t, report.OverallScore, 1.0)

	// Test score breakdown
	breakdown := report.ScoreBreakdown
	assert.GreaterOrEqual(t, breakdown.FieldCompletenessScore, 0.0)
	assert.LessOrEqual(t, breakdown.FieldCompletenessScore, 1.0)
	assert.GreaterOrEqual(t, breakdown.RecordCompletenessScore, 0.0)
	assert.LessOrEqual(t, breakdown.RecordCompletenessScore, 1.0)
	assert.GreaterOrEqual(t, breakdown.PatternAnalysisScore, 0.0)
	assert.LessOrEqual(t, breakdown.PatternAnalysisScore, 1.0)
	assert.GreaterOrEqual(t, breakdown.RuleComplianceScore, 0.0)
	assert.LessOrEqual(t, breakdown.RuleComplianceScore, 1.0)
	assert.GreaterOrEqual(t, breakdown.QualityGateScore, 0.0)
	assert.LessOrEqual(t, breakdown.QualityGateScore, 1.0)
	assert.GreaterOrEqual(t, breakdown.OverallWeightedScore, 0.0)
	assert.LessOrEqual(t, breakdown.OverallWeightedScore, 1.0)

	// Test rule results
	assert.NotEmpty(t, report.RuleResults)
	for _, ruleResult := range report.RuleResults {
		assert.NotEmpty(t, ruleResult.RuleID)
		assert.NotEmpty(t, ruleResult.RuleName)
		assert.NotEmpty(t, ruleResult.RuleType)
		assert.NotEmpty(t, ruleResult.Status)
		assert.Contains(t, []string{"passed", "failed", "skipped"}, ruleResult.Status)
		assert.GreaterOrEqual(t, ruleResult.Score, 0.0)
		assert.LessOrEqual(t, ruleResult.Score, 1.0)
		assert.GreaterOrEqual(t, ruleResult.Threshold, 0.0)
		assert.NotEmpty(t, ruleResult.Severity)
		assert.GreaterOrEqual(t, ruleResult.AffectedRecords, 0)
		assert.Greater(t, ruleResult.ExecutionTime, time.Duration(0))
	}

	// Test validation metrics
	metrics := report.ValidationMetrics
	assert.Greater(t, metrics.TotalValidationTime, time.Duration(0))
	assert.Greater(t, metrics.RulesExecuted, 0)
	assert.GreaterOrEqual(t, metrics.RulesPassed, 0)
	assert.GreaterOrEqual(t, metrics.RulesFailed, 0)
	assert.GreaterOrEqual(t, metrics.RulesSkipped, 0)
	assert.Equal(t, metrics.RulesExecuted, metrics.RulesPassed+metrics.RulesFailed+metrics.RulesSkipped)
	assert.GreaterOrEqual(t, metrics.CriticalIssuesFound, 0)
	assert.GreaterOrEqual(t, metrics.WarningsGenerated, 0)
	assert.Greater(t, metrics.DataPointsValidated, 0)
	assert.Greater(t, metrics.ValidationThroughput, 0.0)
	assert.GreaterOrEqual(t, metrics.ValidationEfficiency, 0.0)
	assert.LessOrEqual(t, metrics.ValidationEfficiency, 1.0)

	// Test compliance status
	compliance := report.ComplianceStatus
	assert.GreaterOrEqual(t, compliance.OverallCompliance, 0.0)
	assert.LessOrEqual(t, compliance.OverallCompliance, 1.0)
	assert.NotEmpty(t, compliance.StandardsCompliance)
	assert.NotEmpty(t, compliance.CertificationStatus)
}

func TestCompletenessValidator_GenerateRecommendations(t *testing.T) {
	validator, db := setupTestCompletenessValidator(t)
	defer db.Close()

	ctx := context.Background()
	config := &CompletenessValidationConfig{}

	result := &CompletenessValidationResult{
		ID:                  "test_validation",
		OverallCompleteness: 0.75,
		FieldAnalysis: map[string]FieldCompletenessAnalysis{
			"website": {
				FieldName:    "website",
				Completeness: 0.45,
				IsOptional:   true,
			},
		},
	}

	recommendations, err := validator.generateRecommendations(ctx, result, config)
	require.NoError(t, err)
	assert.NotNil(t, recommendations)
	assert.NotEmpty(t, recommendations)

	// Test recommendation structure
	for _, rec := range recommendations {
		assert.NotEmpty(t, rec.RecommendationID)
		assert.NotEmpty(t, rec.Type)
		assert.NotEmpty(t, rec.Priority)
		assert.Contains(t, []string{"critical", "high", "medium", "low"}, rec.Priority)
		assert.NotEmpty(t, rec.Title)
		assert.NotEmpty(t, rec.Description)
		assert.NotNil(t, rec.AffectedFields)

		// Test impact assessment
		impact := rec.ImpactAssessment
		assert.GreaterOrEqual(t, impact.DataQualityImprovement, 0.0)
		assert.LessOrEqual(t, impact.DataQualityImprovement, 1.0)
		assert.GreaterOrEqual(t, impact.CompletenessImprovement, 0.0)
		assert.LessOrEqual(t, impact.CompletenessImprovement, 1.0)
		assert.NotEmpty(t, impact.BusinessProcessImpact)
		assert.NotEmpty(t, impact.OverallBusinessValue)

		// Test implementation plan
		plan := rec.ImplementationPlan
		assert.Greater(t, plan.TotalEstimatedTime, time.Duration(0))
		assert.Greater(t, plan.TotalEstimatedCost, 0.0)
		assert.NotEmpty(t, plan.Phases)
		assert.NotNil(t, plan.RequiredResources)

		// Test phases
		for _, phase := range plan.Phases {
			assert.Greater(t, phase.PhaseNumber, 0)
			assert.NotEmpty(t, phase.PhaseName)
			assert.NotEmpty(t, phase.Description)
			assert.Greater(t, phase.EstimatedTime, time.Duration(0))
			assert.GreaterOrEqual(t, phase.EstimatedCost, 0.0)
			assert.NotNil(t, phase.Prerequisites)
			assert.NotNil(t, phase.Deliverables)
			assert.NotNil(t, phase.SuccessCriteria)
		}

		// Test ROI analysis
		roi := rec.ROIAnalysis
		assert.Greater(t, roi.InitialInvestment, 0.0)
		assert.GreaterOrEqual(t, roi.OngoingCosts, 0.0)
		assert.Greater(t, roi.ExpectedBenefits, 0.0)
		assert.Greater(t, roi.PaybackPeriod, time.Duration(0))
		assert.Greater(t, roi.ROIPercentage, 0.0)

		// Test timestamps
		assert.WithinDuration(t, time.Now(), rec.CreatedAt, 5*time.Second)
		assert.True(t, rec.EstimatedCompletion.After(rec.CreatedAt))
	}
}

func TestCompletenessValidator_CalculateOverallCompleteness(t *testing.T) {
	validator, db := setupTestCompletenessValidator(t)
	defer db.Close()

	// Test with mixed field types
	result := &CompletenessValidationResult{
		FieldAnalysis: map[string]FieldCompletenessAnalysis{
			"business_name": {
				FieldName:    "business_name",
				Completeness: 0.95,
				IsRequired:   true,
			},
			"address": {
				FieldName:    "address",
				Completeness: 0.88,
				IsRequired:   true,
			},
			"website": {
				FieldName:    "website",
				Completeness: 0.45,
				IsOptional:   true,
			},
		},
	}

	completeness := validator.calculateOverallCompleteness(result)
	assert.Greater(t, completeness, 0.0)
	assert.LessOrEqual(t, completeness, 1.0)

	// Should be higher than simple average due to required field weighting
	simpleAverage := (0.95 + 0.88 + 0.45) / 3.0
	assert.Greater(t, completeness, simpleAverage)

	// Test with empty field analysis
	emptyResult := &CompletenessValidationResult{
		FieldAnalysis: map[string]FieldCompletenessAnalysis{},
	}

	emptyCompleteness := validator.calculateOverallCompleteness(emptyResult)
	assert.Equal(t, 0.0, emptyCompleteness)
}

func TestCompletenessValidator_DetermineCompletenessLevel(t *testing.T) {
	validator, db := setupTestCompletenessValidator(t)
	defer db.Close()

	// Test different completeness levels
	assert.Equal(t, "excellent", validator.determineCompletenessLevel(0.98))
	assert.Equal(t, "excellent", validator.determineCompletenessLevel(0.95))
	assert.Equal(t, "good", validator.determineCompletenessLevel(0.92))
	assert.Equal(t, "good", validator.determineCompletenessLevel(0.85))
	assert.Equal(t, "fair", validator.determineCompletenessLevel(0.78))
	assert.Equal(t, "fair", validator.determineCompletenessLevel(0.70))
	assert.Equal(t, "poor", validator.determineCompletenessLevel(0.65))
	assert.Equal(t, "poor", validator.determineCompletenessLevel(0.50))
	assert.Equal(t, "critical", validator.determineCompletenessLevel(0.45))
	assert.Equal(t, "critical", validator.determineCompletenessLevel(0.0))
}

func TestCompletenessValidator_ExtractDatasetInfo(t *testing.T) {
	validator, db := setupTestCompletenessValidator(t)
	defer db.Close()

	data := map[string]interface{}{
		"business_name": "Test Company",
		"address":       "123 Test St",
		"phone":         "+1-555-123-4567",
	}

	info := validator.extractDatasetInfo(data)

	// Test dataset info structure
	assert.NotEmpty(t, info.DatasetID)
	assert.NotEmpty(t, info.DatasetName)
	assert.NotEmpty(t, info.DataSource)
	assert.Greater(t, info.RecordCount, 0)
	assert.Greater(t, info.FieldCount, 0)
	assert.Greater(t, info.DataSize, int64(0))
	assert.True(t, info.LastModified.Before(time.Now()))
	assert.NotEmpty(t, info.DataVersion)
	assert.NotEmpty(t, info.Schema)
	assert.NotNil(t, info.DataLineage)

	// Test schema content
	for fieldName, fieldType := range info.Schema {
		assert.NotEmpty(t, fieldName)
		assert.NotEmpty(t, fieldType)
		assert.Contains(t, []string{"string", "number", "date", "boolean"}, fieldType)
	}

	// Test data lineage
	for _, source := range info.DataLineage {
		assert.NotEmpty(t, source)
	}
}

func TestCompletenessValidator_Integration(t *testing.T) {
	validator, db := setupTestCompletenessValidator(t)
	defer db.Close()

	ctx := context.Background()

	// Test complete integration flow with comprehensive configuration
	config := &CompletenessValidationConfig{
		ValidationMode:         "strict",
		EnableFieldAnalysis:    true,
		EnableRecordAnalysis:   true,
		EnablePatternAnalysis:  true,
		EnableTrendAnalysis:    true,
		EnableAnomalyDetection: true,
		RequiredFields:         []string{"business_name", "address"},
		OptionalFields:         []string{"website", "phone", "email"},
		FieldDefinitions: map[string]FieldDefinition{
			"business_name": {
				FieldName:          "business_name",
				FieldType:          "string",
				IsRequired:         true,
				MinLength:          1,
				MaxLength:          255,
				CompletenessWeight: 2.0,
			},
			"website": {
				FieldName:          "website",
				FieldType:          "string",
				IsOptional:         true,
				CompletenessWeight: 0.5,
			},
		},
		CompletenessThresholds: CompletenessThresholds{
			OverallCompletenessMin: 0.80,
			FieldCompletenessMin:   0.70,
			RecordCompletenessMin:  0.75,
			RequiredFieldsMin:      0.95,
			CriticalFieldsMin:      0.98,
		},
		ValidationRules: []CompletenessValidationRule{
			{
				RuleID:            "rule_001",
				RuleName:          "Required Field Completeness",
				RuleType:          "threshold",
				Description:       "Required fields must meet minimum completeness",
				TargetFields:      []string{"business_name", "address"},
				Threshold:         0.95,
				Operator:          ">=",
				Severity:          "critical",
				IsEnabled:         true,
				IsCritical:        true,
				ErrorMessage:      "Required field completeness below threshold",
				RecommendedAction: "Review data collection process",
			},
			{
				RuleID:            "rule_002",
				RuleName:          "Optional Field Guidance",
				RuleType:          "threshold",
				Description:       "Optional fields should meet reasonable completeness",
				TargetFields:      []string{"website", "phone", "email"},
				Threshold:         0.60,
				Operator:          ">=",
				Severity:          "medium",
				IsEnabled:         true,
				IsCritical:        false,
				WarningMessage:    "Optional field completeness could be improved",
				RecommendedAction: "Consider improving data collection incentives",
			},
		},
		QualityGates: []QualityGate{
			{
				GateID:          "gate_001",
				GateName:        "Overall Completeness Gate",
				Description:     "Overall completeness must exceed threshold",
				MetricName:      "overall_completeness",
				ThresholdValue:  0.80,
				Operator:        ">=",
				IsCritical:      true,
				BlocksExecution: false,
				IsEnabled:       true,
				NotifyOnFailure: true,
			},
			{
				GateID:          "gate_002",
				GateName:        "Required Fields Gate",
				Description:     "Required fields must meet strict standards",
				MetricName:      "required_fields_completeness",
				ThresholdValue:  0.95,
				Operator:        ">=",
				IsCritical:      true,
				BlocksExecution: true,
				IsEnabled:       true,
				NotifyOnFailure: true,
			},
		},
		NotificationConfig: NotificationConfig{
			EnableNotifications:      true,
			NotificationChannels:     []string{"email", "slack"},
			CriticalIssueNotify:      true,
			QualityGateFailureNotify: true,
			TrendAnomalyNotify:       true,
			NotificationThreshold:    "medium",
			Recipients:               []string{"data-quality@company.com"},
		},
		ReportingConfig: ReportingConfig{
			EnableReporting:         true,
			ReportFormats:           []string{"json", "pdf", "html"},
			ReportFrequency:         "daily",
			IncludeDetailedAnalysis: true,
			IncludeTrendAnalysis:    true,
			IncludeRecommendations:  true,
		},
		PerformanceConfig: PerformanceConfig{
			MaxProcessingTime:    30 * time.Minute,
			MaxMemoryUsage:       1024 * 1024 * 1024, // 1GB
			BatchSize:            1000,
			ParallelProcessing:   true,
			MaxConcurrentWorkers: 4,
			OptimizeForAccuracy:  true,
		},
	}

	// Mock comprehensive dataset
	data := map[string]interface{}{
		"business_name": "Test Company Inc.",
		"address":       "123 Main Street, Suite 100, Anytown, ST 12345",
		"phone":         "+1-555-123-4567",
		"email":         "contact@testcompany.com",
		"website":       "https://www.testcompany.com",
		"industry":      "Technology",
		"employees":     100,
		"revenue":       1000000.0,
		"established":   "2010-01-01",
		"active":        true,
	}

	result, err := validator.ValidateCompleteness(ctx, data, config)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify comprehensive result structure
	assert.NotEmpty(t, result.ID)
	assert.WithinDuration(t, time.Now(), result.GeneratedAt, 5*time.Second)
	assert.GreaterOrEqual(t, result.OverallCompleteness, 0.0)
	assert.LessOrEqual(t, result.OverallCompleteness, 1.0)
	assert.NotEmpty(t, result.CompletenessLevel)

	// Verify all analysis components are present
	assert.NotEmpty(t, result.FieldAnalysis)
	assert.NotNil(t, result.RecordAnalysis)
	assert.NotNil(t, result.PatternAnalysis)
	assert.NotNil(t, result.TrendAnalysis)
	assert.NotNil(t, result.ValidationReport)
	assert.NotNil(t, result.Recommendations)

	// Verify field analysis completeness
	requiredFieldsAnalyzed := 0
	optionalFieldsAnalyzed := 0
	for fieldName, fieldAnalysis := range result.FieldAnalysis {
		assert.NotEmpty(t, fieldName)
		assert.Equal(t, fieldName, fieldAnalysis.FieldName)

		if fieldAnalysis.IsRequired {
			requiredFieldsAnalyzed++
		}
		if fieldAnalysis.IsOptional {
			optionalFieldsAnalyzed++
		}

		// All fields should have validation status
		assert.NotNil(t, fieldAnalysis.ValidationStatus)
		assert.GreaterOrEqual(t, fieldAnalysis.ValidationStatus.ValidationScore, 0.0)
		assert.LessOrEqual(t, fieldAnalysis.ValidationStatus.ValidationScore, 1.0)
	}

	// Verify validation report comprehensiveness
	assert.Equal(t, result.ID, result.ValidationReport.ValidationID)
	assert.NotEmpty(t, result.ValidationReport.ValidationStatus)
	assert.GreaterOrEqual(t, result.ValidationReport.OverallScore, 0.0)
	assert.LessOrEqual(t, result.ValidationReport.OverallScore, 1.0)
	assert.NotEmpty(t, result.ValidationReport.RuleResults)
	assert.Greater(t, result.ValidationReport.ValidationMetrics.RulesExecuted, 0)

	// Verify recommendations are actionable
	for _, recommendation := range result.Recommendations {
		assert.NotEmpty(t, recommendation.RecommendationID)
		assert.NotEmpty(t, recommendation.Title)
		assert.NotEmpty(t, recommendation.Description)
		assert.NotNil(t, recommendation.ImpactAssessment)
		assert.NotNil(t, recommendation.ImplementationPlan)
		assert.NotNil(t, recommendation.ROIAnalysis)

		// Implementation plan should be detailed
		assert.NotEmpty(t, recommendation.ImplementationPlan.Phases)
		assert.Greater(t, recommendation.ImplementationPlan.TotalEstimatedTime, time.Duration(0))
		assert.Greater(t, recommendation.ImplementationPlan.TotalEstimatedCost, 0.0)
	}

	// Verify configuration is preserved
	assert.Equal(t, config.ValidationMode, result.Configuration.ValidationMode)
	assert.Equal(t, config.EnableFieldAnalysis, result.Configuration.EnableFieldAnalysis)
	assert.Equal(t, config.EnableRecordAnalysis, result.Configuration.EnableRecordAnalysis)
	assert.Equal(t, config.EnablePatternAnalysis, result.Configuration.EnablePatternAnalysis)
	assert.Equal(t, config.EnableTrendAnalysis, result.Configuration.EnableTrendAnalysis)

	// Verify metadata completeness
	assert.Greater(t, result.Metadata.ProcessingTime, time.Duration(0))
	assert.NotEmpty(t, result.Metadata.ValidationVersion)
	assert.NotEmpty(t, result.Metadata.ValidationEngine)
	assert.Greater(t, result.Metadata.DatasetInfo.RecordCount, 0)
	assert.Greater(t, result.Metadata.DatasetInfo.FieldCount, 0)

	// Verify that the overall completeness makes sense given the input data
	// With comprehensive data provided, completeness should be reasonable
	assert.Greater(t, result.OverallCompleteness, 0.7, "With comprehensive data, completeness should be reasonable")

	// Verify that processing time is reasonable
	assert.Less(t, result.Metadata.ProcessingTime, 10*time.Second, "Processing should complete in reasonable time")
}
