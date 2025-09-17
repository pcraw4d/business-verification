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

func setupTestConfidenceReporter(t *testing.T) (*ConfidenceReporter, *sql.DB) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	logger := zap.NewNop()
	industryDB := NewIndustryCodeDatabase(db, logger)
	reporter := NewConfidenceReporter(industryDB, logger)

	return reporter, db
}

func TestConfidenceReporter_GenerateReport(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	ctx := context.Background()
	config := &ReportConfig{
		TimeRange: TimeRange{
			StartDate: time.Now().AddDate(0, 0, -30),
			EndDate:   time.Now(),
			Duration:  "30 days",
		},
		IncludeTrends:            true,
		IncludeAnalytics:         true,
		IncludePerformance:       true,
		IncludeRecommendations:   true,
		IncludeDetailedBreakdown: true,
		IncludeExportData:        true,
		AnomalyThreshold:         1.5,
		TrendSensitivity:         0.1,
	}

	report, err := reporter.GenerateReport(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, report)

	// Test report structure
	assert.NotEmpty(t, report.ReportID)
	assert.WithinDuration(t, time.Now(), report.GeneratedAt, 5*time.Second)
	assert.Equal(t, config.TimeRange, report.TimeRange)

	// Test summary
	assert.NotNil(t, report.Summary)
	assert.Greater(t, report.Summary.TotalClassifications, 0)
	assert.Greater(t, report.Summary.AverageConfidence, 0.0)
	assert.LessOrEqual(t, report.Summary.AverageConfidence, 1.0)
	assert.NotEmpty(t, report.Summary.ConfidenceDistribution)
	assert.NotEmpty(t, report.Summary.KeyInsights)

	// Test trends
	assert.NotNil(t, report.Trends)
	// Note: Trends are only generated when IncludeTrends is true
	// The test config has IncludeTrends: true, so these should be populated
	if config.IncludeTrends {
		assert.NotEmpty(t, report.Trends.DailyTrends)
		assert.NotEmpty(t, report.Trends.WeeklyTrends)
		assert.NotEmpty(t, report.Trends.MonthlyTrends)
	}
	assert.NotNil(t, report.Trends.TrendAnalysis)
	assert.NotNil(t, report.Trends.Seasonality)

	// Test analytics
	assert.NotNil(t, report.Analytics)
	// Note: Analytics are only generated when IncludeAnalytics is true
	if config.IncludeAnalytics {
		assert.NotNil(t, report.Analytics.FactorAnalysis)
		assert.NotNil(t, report.Analytics.CodeTypeAnalysis)
		assert.NotNil(t, report.Analytics.IndustryAnalysis)
		assert.NotNil(t, report.Analytics.AnomalyDetection)
		assert.NotNil(t, report.Analytics.CorrelationAnalysis)
	}

	// Test performance
	assert.NotNil(t, report.Performance)
	// Note: Performance metrics are only generated when IncludePerformance is true
	if config.IncludePerformance {
		assert.NotNil(t, report.Performance.AccuracyMetrics)
		assert.NotNil(t, report.Performance.ReliabilityMetrics)
		assert.NotNil(t, report.Performance.EfficiencyMetrics)
		assert.NotNil(t, report.Performance.QualityMetrics)
	}

	// Test recommendations
	// Note: Recommendations are only generated when IncludeRecommendations is true
	if config.IncludeRecommendations {
		assert.NotEmpty(t, report.Recommendations)
	}

	// Test detailed breakdown
	assert.NotNil(t, report.DetailedBreakdown)
	// Note: Detailed breakdown is only generated when IncludeDetailedBreakdown is true
	if config.IncludeDetailedBreakdown {
		assert.NotEmpty(t, report.DetailedBreakdown.ByTimePeriod)
		assert.NotEmpty(t, report.DetailedBreakdown.ByCodeType)
		assert.NotEmpty(t, report.DetailedBreakdown.ByIndustry)
		assert.NotEmpty(t, report.DetailedBreakdown.ByConfidenceLevel)
	}

	// Test export data
	assert.NotNil(t, report.ExportData)
	// Note: Export data is only generated when IncludeExportData is true
	if config.IncludeExportData {
		assert.NotEmpty(t, report.ExportData.CSVData)
		assert.NotEmpty(t, report.ExportData.JSONData)
		assert.NotEmpty(t, report.ExportData.ChartData)
		assert.NotEmpty(t, report.ExportData.SummaryData)
	}
}

func TestConfidenceReporter_GenerateReport_MinimalConfig(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	ctx := context.Background()
	config := &ReportConfig{
		TimeRange: TimeRange{
			StartDate: time.Now().AddDate(0, 0, -7),
			EndDate:   time.Now(),
			Duration:  "7 days",
		},
		IncludeTrends:            false,
		IncludeAnalytics:         false,
		IncludePerformance:       false,
		IncludeRecommendations:   false,
		IncludeDetailedBreakdown: false,
		IncludeExportData:        false,
	}

	report, err := reporter.GenerateReport(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, report)

	// Should only have summary
	assert.NotNil(t, report.Summary)
	assert.Empty(t, report.Trends.DailyTrends)
	assert.Empty(t, report.Analytics.FactorAnalysis.TopFactors)
	assert.Empty(t, report.Performance.AccuracyMetrics.OverallAccuracy)
	assert.Empty(t, report.Recommendations)
	assert.Empty(t, report.DetailedBreakdown.ByTimePeriod)
	assert.Empty(t, report.ExportData.CSVData)
}

func TestConfidenceReporter_GenerateReport_InvalidTimeRange(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	ctx := context.Background()
	config := &ReportConfig{
		TimeRange: TimeRange{
			StartDate: time.Now(),
			EndDate:   time.Now().AddDate(0, 0, -1), // End before start
			Duration:  "invalid",
		},
	}

	report, err := reporter.GenerateReport(ctx, config)
	require.NoError(t, err) // Should still work with mock data
	assert.NotNil(t, report)
}

func TestConfidenceReporter_TrendGeneration(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	config := &ReportConfig{
		TimeRange: TimeRange{
			StartDate: time.Now().AddDate(0, 0, -30),
			EndDate:   time.Now(),
		},
	}

	// Test daily trends
	dailyTrends := reporter.generateDailyTrends(config)
	assert.Len(t, dailyTrends, 30)
	for i, trend := range dailyTrends {
		assert.WithinDuration(t, config.TimeRange.StartDate.AddDate(0, 0, i), trend.Date, time.Second)
		assert.Greater(t, trend.AverageConfidence, 0.0)
		assert.LessOrEqual(t, trend.AverageConfidence, 1.0)
		assert.Greater(t, trend.TotalCount, 0)
		assert.GreaterOrEqual(t, trend.HighConfidenceCount, 0)
		assert.GreaterOrEqual(t, trend.LowConfidenceCount, 0)
	}

	// Test weekly trends
	weeklyTrends := reporter.generateWeeklyTrends(config)
	assert.Len(t, weeklyTrends, 12)
	for i, trend := range weeklyTrends {
		assert.WithinDuration(t, config.TimeRange.StartDate.AddDate(0, 0, i*7), trend.Date, time.Second)
		assert.Greater(t, trend.AverageConfidence, 0.0)
		assert.LessOrEqual(t, trend.AverageConfidence, 1.0)
		assert.Greater(t, trend.TotalCount, 0)
	}

	// Test monthly trends
	monthlyTrends := reporter.generateMonthlyTrends(config)
	assert.Len(t, monthlyTrends, 6)
	for i, trend := range monthlyTrends {
		assert.WithinDuration(t, config.TimeRange.StartDate.AddDate(0, i, 0), trend.Date, time.Second)
		assert.Greater(t, trend.AverageConfidence, 0.0)
		assert.LessOrEqual(t, trend.AverageConfidence, 1.0)
		assert.Greater(t, trend.TotalCount, 0)
	}
}

func TestConfidenceReporter_TrendAnalysis(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	// Create test trends with known pattern
	trends := []TrendPoint{
		{Date: time.Now().AddDate(0, 0, -2), AverageConfidence: 0.80, TotalCount: 100},
		{Date: time.Now().AddDate(0, 0, -1), AverageConfidence: 0.85, TotalCount: 110},
		{Date: time.Now(), AverageConfidence: 0.90, TotalCount: 120},
	}

	analysis := reporter.analyzeTrends(trends)
	assert.NotNil(t, analysis)
	assert.Equal(t, "positive", analysis.OverallTrend)
	assert.Greater(t, analysis.TrendSlope, 0.0)
	assert.Greater(t, analysis.TrendStrength, 0.0)
	assert.LessOrEqual(t, analysis.TrendStrength, 1.0)
	assert.GreaterOrEqual(t, analysis.Volatility, 0.0)
	assert.NotEmpty(t, analysis.SeasonalPatterns)

	// Test with declining trend
	decliningTrends := []TrendPoint{
		{Date: time.Now().AddDate(0, 0, -2), AverageConfidence: 0.90, TotalCount: 100},
		{Date: time.Now().AddDate(0, 0, -1), AverageConfidence: 0.85, TotalCount: 110},
		{Date: time.Now(), AverageConfidence: 0.80, TotalCount: 120},
	}

	decliningAnalysis := reporter.analyzeTrends(decliningTrends)
	assert.Equal(t, "negative", decliningAnalysis.OverallTrend)
	assert.Less(t, decliningAnalysis.TrendSlope, 0.0)

	// Test with stable trend
	stableTrends := []TrendPoint{
		{Date: time.Now().AddDate(0, 0, -2), AverageConfidence: 0.85, TotalCount: 100},
		{Date: time.Now().AddDate(0, 0, -1), AverageConfidence: 0.85, TotalCount: 110},
		{Date: time.Now(), AverageConfidence: 0.85, TotalCount: 120},
	}

	stableAnalysis := reporter.analyzeTrends(stableTrends)
	assert.Equal(t, "neutral", stableAnalysis.OverallTrend)
	assert.Equal(t, 0.0, stableAnalysis.TrendSlope)
}

func TestConfidenceReporter_SeasonalityAnalysis(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	trends := []TrendPoint{
		{Date: time.Now(), AverageConfidence: 0.85, TotalCount: 100},
	}

	seasonality := reporter.analyzeSeasonality(trends)
	assert.NotNil(t, seasonality)
	assert.True(t, seasonality.HasSeasonality)
	assert.NotEmpty(t, seasonality.SeasonalPeriods)
	assert.NotEmpty(t, seasonality.PeakSeasons)
	assert.NotEmpty(t, seasonality.LowSeasons)
	assert.Greater(t, seasonality.SeasonalStrength, 0.0)
	assert.LessOrEqual(t, seasonality.SeasonalStrength, 1.0)
}

func TestConfidenceReporter_FactorAnalysis(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	factorAnalysis := reporter.generateFactorAnalysis()
	assert.NotNil(t, factorAnalysis)

	// Test top factors
	assert.NotEmpty(t, factorAnalysis.TopFactors)
	for _, factor := range factorAnalysis.TopFactors {
		assert.NotEmpty(t, factor.FactorName)
		assert.Greater(t, factor.AverageScore, 0.0)
		assert.LessOrEqual(t, factor.AverageScore, 1.0)
		assert.Greater(t, factor.Impact, 0.0)
		assert.LessOrEqual(t, factor.Impact, 1.0)
		assert.NotEmpty(t, factor.Trend)
		assert.NotEmpty(t, factor.Recommendations)
	}

	// Test weakest factors
	assert.NotEmpty(t, factorAnalysis.WeakestFactors)
	for _, factor := range factorAnalysis.WeakestFactors {
		assert.NotEmpty(t, factor.FactorName)
		assert.Greater(t, factor.AverageScore, 0.0)
		assert.LessOrEqual(t, factor.AverageScore, 1.0)
		assert.Greater(t, factor.Impact, 0.0)
		assert.LessOrEqual(t, factor.Impact, 1.0)
		assert.NotEmpty(t, factor.Trend)
		assert.NotEmpty(t, factor.Recommendations)
	}

	// Test factor correlations
	assert.NotEmpty(t, factorAnalysis.FactorCorrelations)
	for correlation, value := range factorAnalysis.FactorCorrelations {
		assert.NotEmpty(t, correlation)
		assert.GreaterOrEqual(t, value, -1.0)
		assert.LessOrEqual(t, value, 1.0)
	}

	// Test factor trends
	assert.NotEmpty(t, factorAnalysis.FactorTrends)
	for factor, trend := range factorAnalysis.FactorTrends {
		assert.NotEmpty(t, factor)
		assert.NotEmpty(t, trend)
	}
}

func TestConfidenceReporter_CodeTypeAnalysis(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	codeTypeAnalysis := reporter.generateCodeTypeAnalysis()
	assert.NotNil(t, codeTypeAnalysis)

	// Test by code type
	assert.NotEmpty(t, codeTypeAnalysis.ByCodeType)
	for codeType, metrics := range codeTypeAnalysis.ByCodeType {
		assert.NotEmpty(t, codeType)
		assert.Greater(t, metrics.AverageConfidence, 0.0)
		assert.LessOrEqual(t, metrics.AverageConfidence, 1.0)
		assert.Greater(t, metrics.TotalCount, 0)
		assert.Greater(t, metrics.HighConfidenceRate, 0.0)
		assert.LessOrEqual(t, metrics.HighConfidenceRate, 1.0)
		assert.NotEmpty(t, metrics.Trend)
		assert.NotEmpty(t, metrics.KeyFactors)
	}

	// Test best and worst performing
	assert.NotEmpty(t, codeTypeAnalysis.BestPerforming)
	assert.NotEmpty(t, codeTypeAnalysis.WorstPerforming)
	assert.NotEqual(t, codeTypeAnalysis.BestPerforming, codeTypeAnalysis.WorstPerforming)

	// Test improvement opportunities
	assert.NotEmpty(t, codeTypeAnalysis.ImprovementOpportunities)
}

func TestConfidenceReporter_IndustryAnalysis(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	industryAnalysis := reporter.generateIndustryAnalysis()
	assert.NotNil(t, industryAnalysis)

	// Test by industry
	assert.NotEmpty(t, industryAnalysis.ByIndustry)
	for industry, metrics := range industryAnalysis.ByIndustry {
		assert.NotEmpty(t, industry)
		assert.Greater(t, metrics.AverageConfidence, 0.0)
		assert.LessOrEqual(t, metrics.AverageConfidence, 1.0)
		assert.Greater(t, metrics.TotalCount, 0)
		assert.Greater(t, metrics.HighConfidenceRate, 0.0)
		assert.LessOrEqual(t, metrics.HighConfidenceRate, 1.0)
		assert.NotEmpty(t, metrics.CommonFactors)
		assert.NotEmpty(t, metrics.Challenges)
	}

	// Test top and bottom industries
	assert.NotEmpty(t, industryAnalysis.TopIndustries)
	assert.NotEmpty(t, industryAnalysis.BottomIndustries)

	for _, ranking := range industryAnalysis.TopIndustries {
		assert.NotEmpty(t, ranking.IndustryName)
		assert.Greater(t, ranking.AverageConfidence, 0.0)
		assert.LessOrEqual(t, ranking.AverageConfidence, 1.0)
		assert.Greater(t, ranking.Rank, 0)
		assert.Greater(t, ranking.TotalCount, 0)
	}

	for _, ranking := range industryAnalysis.BottomIndustries {
		assert.NotEmpty(t, ranking.IndustryName)
		assert.Greater(t, ranking.AverageConfidence, 0.0)
		assert.LessOrEqual(t, ranking.AverageConfidence, 1.0)
		assert.Greater(t, ranking.Rank, 0)
		assert.Greater(t, ranking.TotalCount, 0)
	}

	// Test industry trends
	assert.NotEmpty(t, industryAnalysis.IndustryTrends)
	for industry, trend := range industryAnalysis.IndustryTrends {
		assert.NotEmpty(t, industry)
		assert.NotEmpty(t, trend)
	}
}

func TestConfidenceReporter_AnomalyDetection(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	// Test with different thresholds
	anomalyDetection1 := reporter.generateAnomalyDetection(1.0)
	assert.NotNil(t, anomalyDetection1)
	assert.GreaterOrEqual(t, anomalyDetection1.AnomalyRate, 0.0)
	assert.LessOrEqual(t, anomalyDetection1.AnomalyRate, 1.0)
	assert.NotEmpty(t, anomalyDetection1.AnomalyPatterns)
	assert.NotEmpty(t, anomalyDetection1.RiskAssessment)

	// Test detected anomalies
	if len(anomalyDetection1.DetectedAnomalies) > 0 {
		for _, anomaly := range anomalyDetection1.DetectedAnomalies {
			assert.NotEmpty(t, anomaly.ID)
			assert.NotEmpty(t, anomaly.Type)
			assert.NotEmpty(t, anomaly.Severity)
			assert.NotEmpty(t, anomaly.Description)
			// Allow for 3 days since the mock data sets DetectedAt to 2 days ago
			assert.WithinDuration(t, time.Now(), anomaly.DetectedAt, 3*24*time.Hour)
			assert.NotEmpty(t, anomaly.Impact)
			assert.NotEmpty(t, anomaly.Recommendations)
		}
	}

	// Test with higher threshold
	anomalyDetection2 := reporter.generateAnomalyDetection(2.0)
	assert.NotNil(t, anomalyDetection2)
}

func TestConfidenceReporter_CorrelationAnalysis(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	correlationAnalysis := reporter.generateCorrelationAnalysis()
	assert.NotNil(t, correlationAnalysis)

	// Test factor correlations
	assert.NotEmpty(t, correlationAnalysis.FactorCorrelations)
	for factor, correlations := range correlationAnalysis.FactorCorrelations {
		assert.NotEmpty(t, factor)
		for variable, correlation := range correlations {
			assert.NotEmpty(t, variable)
			assert.GreaterOrEqual(t, correlation, -1.0)
			assert.LessOrEqual(t, correlation, 1.0)
		}
	}

	// Test time correlations
	assert.NotEmpty(t, correlationAnalysis.TimeCorrelations)
	for timeVar, correlation := range correlationAnalysis.TimeCorrelations {
		assert.NotEmpty(t, timeVar)
		assert.GreaterOrEqual(t, correlation, -1.0)
		assert.LessOrEqual(t, correlation, 1.0)
	}

	// Test business correlations
	assert.NotEmpty(t, correlationAnalysis.BusinessCorrelations)
	for businessVar, correlations := range correlationAnalysis.BusinessCorrelations {
		assert.NotEmpty(t, businessVar)
		for variable, correlation := range correlations {
			assert.NotEmpty(t, variable)
			assert.GreaterOrEqual(t, correlation, -1.0)
			assert.LessOrEqual(t, correlation, 1.0)
		}
	}

	// Test significant correlations
	assert.NotEmpty(t, correlationAnalysis.SignificantCorrelations)
	for _, insight := range correlationAnalysis.SignificantCorrelations {
		assert.NotEmpty(t, insight.Variable1)
		assert.NotEmpty(t, insight.Variable2)
		assert.GreaterOrEqual(t, insight.Correlation, -1.0)
		assert.LessOrEqual(t, insight.Correlation, 1.0)
		assert.NotEmpty(t, insight.Strength)
		assert.NotEmpty(t, insight.Direction)
		assert.Greater(t, insight.Significance, 0.0)
		assert.LessOrEqual(t, insight.Significance, 1.0)
		assert.NotEmpty(t, insight.Interpretation)
	}
}

func TestConfidenceReporter_PerformanceMetrics(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	// Test accuracy metrics
	accuracyMetrics := reporter.generateAccuracyMetrics()
	assert.NotNil(t, accuracyMetrics)
	assert.Greater(t, accuracyMetrics.OverallAccuracy, 0.0)
	assert.LessOrEqual(t, accuracyMetrics.OverallAccuracy, 1.0)
	assert.Greater(t, accuracyMetrics.Precision, 0.0)
	assert.LessOrEqual(t, accuracyMetrics.Precision, 1.0)
	assert.Greater(t, accuracyMetrics.Recall, 0.0)
	assert.LessOrEqual(t, accuracyMetrics.Recall, 1.0)
	assert.Greater(t, accuracyMetrics.F1Score, 0.0)
	assert.LessOrEqual(t, accuracyMetrics.F1Score, 1.0)
	assert.NotEmpty(t, accuracyMetrics.AccuracyByConfidence)

	// Test reliability metrics
	reliabilityMetrics := reporter.generateReliabilityMetrics()
	assert.NotNil(t, reliabilityMetrics)
	assert.Greater(t, reliabilityMetrics.ConsistencyScore, 0.0)
	assert.LessOrEqual(t, reliabilityMetrics.ConsistencyScore, 1.0)
	assert.Greater(t, reliabilityMetrics.StabilityScore, 0.0)
	assert.LessOrEqual(t, reliabilityMetrics.StabilityScore, 1.0)
	assert.Greater(t, reliabilityMetrics.ReliabilityIndex, 0.0)
	assert.LessOrEqual(t, reliabilityMetrics.ReliabilityIndex, 1.0)
	assert.Greater(t, reliabilityMetrics.ConfidenceCalibration, 0.0)
	assert.LessOrEqual(t, reliabilityMetrics.ConfidenceCalibration, 1.0)

	// Test efficiency metrics
	efficiencyMetrics := reporter.generateEfficiencyMetrics()
	assert.NotNil(t, efficiencyMetrics)
	assert.Greater(t, efficiencyMetrics.ProcessingTime, 0.0)
	assert.Greater(t, efficiencyMetrics.Throughput, 0.0)
	assert.Greater(t, efficiencyMetrics.ResourceUsage, 0.0)
	assert.LessOrEqual(t, efficiencyMetrics.ResourceUsage, 1.0)
	assert.Greater(t, efficiencyMetrics.OptimizationScore, 0.0)
	assert.LessOrEqual(t, efficiencyMetrics.OptimizationScore, 1.0)

	// Test quality metrics
	qualityMetrics := reporter.generateQualityMetrics()
	assert.NotNil(t, qualityMetrics)
	assert.Greater(t, qualityMetrics.DataQualityScore, 0.0)
	assert.LessOrEqual(t, qualityMetrics.DataQualityScore, 1.0)
	assert.Greater(t, qualityMetrics.ValidationScore, 0.0)
	assert.LessOrEqual(t, qualityMetrics.ValidationScore, 1.0)
	assert.Greater(t, qualityMetrics.CompletenessScore, 0.0)
	assert.LessOrEqual(t, qualityMetrics.CompletenessScore, 1.0)
	assert.Greater(t, qualityMetrics.ConsistencyScore, 0.0)
	assert.LessOrEqual(t, qualityMetrics.ConsistencyScore, 1.0)
}

func TestConfidenceReporter_Recommendations(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	config := &ReportConfig{}

	// Test with high low confidence rate
	summary := &ConfidenceSummary{
		LowConfidenceRate: 0.20, // Above 15% threshold
		AverageConfidence: 0.85,
	}

	recommendations, err := reporter.generateRecommendations(context.Background(), config, summary)
	require.NoError(t, err)
	assert.NotEmpty(t, recommendations)

	// Check for low confidence recommendation
	foundLowConfidenceRec := false
	for _, rec := range recommendations {
		if rec.Category == "confidence_improvement" && rec.Priority == "high" {
			foundLowConfidenceRec = true
			assert.NotEmpty(t, rec.ID)
			assert.NotEmpty(t, rec.Title)
			assert.NotEmpty(t, rec.Description)
			assert.NotEmpty(t, rec.Impact)
			assert.NotEmpty(t, rec.Effort)
			assert.NotEmpty(t, rec.Implementation)
			assert.NotEmpty(t, rec.ExpectedBenefit)
		}
	}
	assert.True(t, foundLowConfidenceRec)

	// Test with low average confidence
	summary2 := &ConfidenceSummary{
		LowConfidenceRate: 0.10,
		AverageConfidence: 0.75, // Below 0.80 threshold
	}

	recommendations2, err := reporter.generateRecommendations(context.Background(), config, summary2)
	require.NoError(t, err)
	assert.NotEmpty(t, recommendations2)

	// Check for average confidence recommendation
	foundAvgConfidenceRec := false
	for _, rec := range recommendations2 {
		if rec.Category == "confidence_improvement" && rec.Priority == "medium" {
			foundAvgConfidenceRec = true
			assert.NotEmpty(t, rec.ID)
			assert.NotEmpty(t, rec.Title)
			assert.NotEmpty(t, rec.Description)
			assert.NotEmpty(t, rec.Impact)
			assert.NotEmpty(t, rec.Effort)
			assert.NotEmpty(t, rec.Implementation)
			assert.NotEmpty(t, rec.ExpectedBenefit)
		}
	}
	assert.True(t, foundAvgConfidenceRec)
}

func TestConfidenceReporter_DetailedBreakdown(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	config := &ReportConfig{}

	breakdown, err := reporter.generateDetailedBreakdown(context.Background(), config)
	require.NoError(t, err)
	assert.NotNil(t, breakdown)

	// Test time period breakdown
	assert.NotEmpty(t, breakdown.ByTimePeriod)
	for period, periodBreakdown := range breakdown.ByTimePeriod {
		assert.NotEmpty(t, period)
		assert.NotEmpty(t, periodBreakdown.Period)
		assert.Greater(t, periodBreakdown.TotalCount, 0)
		assert.Greater(t, periodBreakdown.AverageConfidence, 0.0)
		assert.LessOrEqual(t, periodBreakdown.AverageConfidence, 1.0)
		assert.NotEmpty(t, periodBreakdown.Distribution)
		assert.NotEmpty(t, periodBreakdown.TopFactors)
	}

	// Test code type breakdown
	assert.NotEmpty(t, breakdown.ByCodeType)
	for codeType, codeTypeBreakdown := range breakdown.ByCodeType {
		assert.NotEmpty(t, codeType)
		assert.NotEmpty(t, codeTypeBreakdown.CodeType)
		assert.Greater(t, codeTypeBreakdown.TotalCount, 0)
		assert.Greater(t, codeTypeBreakdown.AverageConfidence, 0.0)
		assert.LessOrEqual(t, codeTypeBreakdown.AverageConfidence, 1.0)
		assert.NotEmpty(t, codeTypeBreakdown.FactorScores)
		assert.NotEmpty(t, codeTypeBreakdown.Trends)
	}

	// Test industry breakdown
	assert.NotEmpty(t, breakdown.ByIndustry)
	for industry, industryBreakdown := range breakdown.ByIndustry {
		assert.NotEmpty(t, industry)
		assert.NotEmpty(t, industryBreakdown.Industry)
		assert.Greater(t, industryBreakdown.TotalCount, 0)
		assert.Greater(t, industryBreakdown.AverageConfidence, 0.0)
		assert.LessOrEqual(t, industryBreakdown.AverageConfidence, 1.0)
		assert.NotEmpty(t, industryBreakdown.CodeTypeDistribution)
		assert.NotEmpty(t, industryBreakdown.CommonFactors)
	}

	// Test confidence level breakdown
	assert.NotEmpty(t, breakdown.ByConfidenceLevel)
	for level, levelBreakdown := range breakdown.ByConfidenceLevel {
		assert.NotEmpty(t, level)
		assert.NotEmpty(t, levelBreakdown.Level)
		assert.Greater(t, levelBreakdown.TotalCount, 0)
		assert.Greater(t, levelBreakdown.Percentage, 0.0)
		assert.LessOrEqual(t, levelBreakdown.Percentage, 100.0)
		assert.NotEmpty(t, levelBreakdown.CommonFactors)
		assert.NotEmpty(t, levelBreakdown.Trends)
	}
}

func TestConfidenceReporter_ExportData(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	report := &ConfidenceReport{
		ReportID:    "test_report",
		GeneratedAt: time.Now(),
		Summary: ConfidenceSummary{
			TotalClassifications: 1000,
			AverageConfidence:    0.85,
		},
	}

	exportData, err := reporter.generateExportData(context.Background(), report)
	require.NoError(t, err)
	assert.NotNil(t, exportData)

	// Test CSV data
	assert.NotEmpty(t, exportData.CSVData)
	assert.Contains(t, exportData.CSVData, "Date,AverageConfidence,TotalCount")

	// Test JSON data
	assert.NotEmpty(t, exportData.JSONData)
	assert.Contains(t, exportData.JSONData, "summary")

	// Test chart data
	assert.NotEmpty(t, exportData.ChartData)
	assert.Contains(t, exportData.ChartData, "type")

	// Test summary data
	assert.NotEmpty(t, exportData.SummaryData)
	assert.Contains(t, exportData.SummaryData, "Confidence Score Report Summary")
}

func TestConfidenceReporter_Integration(t *testing.T) {
	reporter, db := setupTestConfidenceReporter(t)
	defer db.Close()

	ctx := context.Background()
	config := &ReportConfig{
		TimeRange: TimeRange{
			StartDate: time.Now().AddDate(0, 0, -7),
			EndDate:   time.Now(),
			Duration:  "7 days",
		},
		IncludeTrends:            true,
		IncludeAnalytics:         true,
		IncludePerformance:       true,
		IncludeRecommendations:   true,
		IncludeDetailedBreakdown: true,
		IncludeExportData:        true,
		AnomalyThreshold:         1.5,
		TrendSensitivity:         0.1,
	}

	// Generate full report
	report, err := reporter.GenerateReport(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, report)

	// Verify data consistency
	assert.Equal(t, report.Summary.TotalClassifications, 1000)
	assert.Equal(t, report.Summary.AverageConfidence, 0.85)
	assert.Equal(t, report.Summary.HighConfidenceRate, 0.60)
	assert.Equal(t, report.Summary.LowConfidenceRate, 0.10)

	// Verify trend data consistency
	assert.Len(t, report.Trends.DailyTrends, 30)
	assert.Len(t, report.Trends.WeeklyTrends, 12)
	assert.Len(t, report.Trends.MonthlyTrends, 6)

	// Verify analytics consistency
	if config.IncludeAnalytics {
		assert.NotEmpty(t, report.Analytics.FactorAnalysis.TopFactors)
		assert.NotEmpty(t, report.Analytics.CodeTypeAnalysis.ByCodeType)
		assert.NotEmpty(t, report.Analytics.IndustryAnalysis.ByIndustry)
	}

	// Verify performance consistency
	if config.IncludePerformance {
		assert.Greater(t, report.Performance.AccuracyMetrics.OverallAccuracy, 0.9)
		assert.Greater(t, report.Performance.ReliabilityMetrics.ReliabilityIndex, 0.8)
	}

	// Verify recommendations are generated
	if config.IncludeRecommendations {
		assert.NotEmpty(t, report.Recommendations)
	}

	// Verify export data is generated
	if config.IncludeExportData {
		assert.NotEmpty(t, report.ExportData.CSVData)
		assert.NotEmpty(t, report.ExportData.JSONData)
		assert.NotEmpty(t, report.ExportData.ChartData)
		assert.NotEmpty(t, report.ExportData.SummaryData)
	}
}
