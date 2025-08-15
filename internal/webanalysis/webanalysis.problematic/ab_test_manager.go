package webanalysis

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ABTestManager manages A/B test results and analysis
type ABTestManager struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	results map[string][]*ABTestResult // testID -> results
	cache   map[string]*ABTestAnalysis // testID -> analysis
}

// ABTestAnalysis represents statistical analysis of A/B test results
type ABTestAnalysis struct {
	TestID                  string                   `json:"test_id"`
	TotalTests              int                      `json:"total_tests"`
	BasicTests              int                      `json:"basic_tests"`
	EnhancedTests           int                      `json:"enhanced_tests"`
	BasicSuccessRate        float64                  `json:"basic_success_rate"`
	EnhancedSuccessRate     float64                  `json:"enhanced_success_rate"`
	BasicAvgResponseTime    time.Duration            `json:"basic_avg_response_time"`
	EnhancedAvgResponseTime time.Duration            `json:"enhanced_avg_response_time"`
	BasicAvgAccuracy        float64                  `json:"basic_avg_accuracy"`
	EnhancedAvgAccuracy     float64                  `json:"enhanced_avg_accuracy"`
	BasicAvgDataQuality     float64                  `json:"basic_avg_data_quality"`
	EnhancedAvgDataQuality  float64                  `json:"enhanced_avg_data_quality"`
	ImprovementMetrics      *ImprovementMetrics      `json:"improvement_metrics"`
	StatisticalSignificance *StatisticalSignificance `json:"statistical_significance"`
	Generated               time.Time                `json:"generated"`
}

// ImprovementMetrics represents improvement calculations
type ImprovementMetrics struct {
	SuccessRateImprovement  float64       `json:"success_rate_improvement"`
	ResponseTimeImprovement time.Duration `json:"response_time_improvement"`
	AccuracyImprovement     float64       `json:"accuracy_improvement"`
	DataQualityImprovement  float64       `json:"data_quality_improvement"`
	OverallImprovement      float64       `json:"overall_improvement"`
}

// StatisticalSignificance represents statistical analysis
type StatisticalSignificance struct {
	SuccessRatePValue  float64 `json:"success_rate_p_value"`
	ResponseTimePValue float64 `json:"response_time_p_value"`
	AccuracyPValue     float64 `json:"accuracy_p_value"`
	DataQualityPValue  float64 `json:"data_quality_p_value"`
	IsSignificant      bool    `json:"is_significant"`
	ConfidenceLevel    float64 `json:"confidence_level"`
}

// NewABTestManager creates a new A/B test manager
func NewABTestManager(logger *zap.Logger) *ABTestManager {
	return &ABTestManager{
		logger:  logger,
		results: make(map[string][]*ABTestResult),
		cache:   make(map[string]*ABTestAnalysis),
	}
}

// StoreResult stores an A/B test result
func (atm *ABTestManager) StoreResult(result *ABTestResult) error {
	atm.mu.Lock()
	defer atm.mu.Unlock()

	if result == nil {
		return fmt.Errorf("cannot store nil result")
	}

	// Add to results
	atm.results[result.TestID] = append(atm.results[result.TestID], result)

	// Invalidate cache for this test
	delete(atm.cache, result.TestID)

	atm.logger.Debug("Stored A/B test result",
		zap.String("test_id", result.TestID),
		zap.String("url", result.URL),
		zap.String("method", result.Method),
		zap.Bool("success", result.Success),
		zap.Duration("response_time", result.ResponseTime),
	)

	return nil
}

// GetResults retrieves all results for a test ID
func (atm *ABTestManager) GetResults(testID string) ([]*ABTestResult, error) {
	atm.mu.RLock()
	defer atm.mu.RUnlock()

	results, exists := atm.results[testID]
	if !exists {
		return nil, fmt.Errorf("no results found for test ID: %s", testID)
	}

	return results, nil
}

// GetAnalysis retrieves or computes analysis for a test ID
func (atm *ABTestManager) GetAnalysis(ctx context.Context, testID string) (*ABTestAnalysis, error) {
	atm.mu.RLock()
	if analysis, exists := atm.cache[testID]; exists {
		atm.mu.RUnlock()
		return analysis, nil
	}
	atm.mu.RUnlock()

	// Compute analysis
	atm.mu.Lock()
	defer atm.mu.Unlock()

	// Double-check cache after acquiring write lock
	if analysis, exists := atm.cache[testID]; exists {
		return analysis, nil
	}

	results, exists := atm.results[testID]
	if !exists {
		return nil, fmt.Errorf("no results found for test ID: %s", testID)
	}

	analysis, err := atm.computeAnalysis(results)
	if err != nil {
		return nil, fmt.Errorf("failed to compute analysis: %w", err)
	}

	// Cache the analysis
	atm.cache[testID] = analysis

	return analysis, nil
}

// computeAnalysis computes statistical analysis for A/B test results
func (atm *ABTestManager) computeAnalysis(results []*ABTestResult) (*ABTestAnalysis, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to analyze")
	}

	analysis := &ABTestAnalysis{
		TestID:    results[0].TestID,
		Generated: time.Now(),
	}

	// Separate basic and enhanced results
	var basicResults, enhancedResults []*ABTestResult
	for _, result := range results {
		if result.Method == "basic" {
			basicResults = append(basicResults, result)
		} else if result.Method == "enhanced" {
			enhancedResults = append(enhancedResults, result)
		}
	}

	analysis.TotalTests = len(results)
	analysis.BasicTests = len(basicResults)
	analysis.EnhancedTests = len(enhancedResults)

	// Calculate basic metrics
	if len(basicResults) > 0 {
		analysis.BasicSuccessRate = atm.calculateSuccessRate(basicResults)
		analysis.BasicAvgResponseTime = atm.calculateAvgResponseTime(basicResults)
		analysis.BasicAvgAccuracy = atm.calculateAvgAccuracy(basicResults)
		analysis.BasicAvgDataQuality = atm.calculateAvgDataQuality(basicResults)
	}

	// Calculate enhanced metrics
	if len(enhancedResults) > 0 {
		analysis.EnhancedSuccessRate = atm.calculateSuccessRate(enhancedResults)
		analysis.EnhancedAvgResponseTime = atm.calculateAvgResponseTime(enhancedResults)
		analysis.EnhancedAvgAccuracy = atm.calculateAvgAccuracy(enhancedResults)
		analysis.EnhancedAvgDataQuality = atm.calculateAvgDataQuality(enhancedResults)
	}

	// Calculate improvement metrics
	analysis.ImprovementMetrics = &ImprovementMetrics{
		SuccessRateImprovement:  analysis.EnhancedSuccessRate - analysis.BasicSuccessRate,
		ResponseTimeImprovement: analysis.BasicAvgResponseTime - analysis.EnhancedAvgResponseTime,
		AccuracyImprovement:     analysis.EnhancedAvgAccuracy - analysis.BasicAvgAccuracy,
		DataQualityImprovement:  analysis.EnhancedAvgDataQuality - analysis.BasicAvgDataQuality,
	}

	// Calculate overall improvement score
	analysis.ImprovementMetrics.OverallImprovement = atm.calculateOverallImprovement(analysis.ImprovementMetrics)

	// Calculate statistical significance
	analysis.StatisticalSignificance = atm.calculateStatisticalSignificance(basicResults, enhancedResults)

	return analysis, nil
}

// calculateSuccessRate calculates success rate for a set of results
func (atm *ABTestManager) calculateSuccessRate(results []*ABTestResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	return float64(successCount) / float64(len(results))
}

// calculateAvgResponseTime calculates average response time
func (atm *ABTestManager) calculateAvgResponseTime(results []*ABTestResult) time.Duration {
	if len(results) == 0 {
		return 0
	}

	totalDuration := time.Duration(0)
	for _, result := range results {
		totalDuration += result.ResponseTime
	}

	return totalDuration / time.Duration(len(results))
}

// calculateAvgAccuracy calculates average accuracy
func (atm *ABTestManager) calculateAvgAccuracy(results []*ABTestResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	totalAccuracy := 0.0
	for _, result := range results {
		totalAccuracy += result.Accuracy
	}

	return totalAccuracy / float64(len(results))
}

// calculateAvgDataQuality calculates average data quality
func (atm *ABTestManager) calculateAvgDataQuality(results []*ABTestResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	totalQuality := 0.0
	for _, result := range results {
		totalQuality += result.DataQuality
	}

	return totalQuality / float64(len(results))
}

// calculateOverallImprovement calculates overall improvement score
func (atm *ABTestManager) calculateOverallImprovement(metrics *ImprovementMetrics) float64 {
	// Weighted combination of improvements
	weights := map[string]float64{
		"success_rate":  0.4, // Most important
		"accuracy":      0.3, // Second most important
		"data_quality":  0.2, // Third most important
		"response_time": 0.1, // Least important (normalized)
	}

	// Normalize response time improvement (convert to percentage)
	responseTimeImprovement := float64(metrics.ResponseTimeImprovement.Milliseconds())
	if responseTimeImprovement > 0 {
		responseTimeImprovement = responseTimeImprovement / 1000.0 // Convert to seconds
	}

	overallImprovement :=
		metrics.SuccessRateImprovement*weights["success_rate"] +
			metrics.AccuracyImprovement*weights["accuracy"] +
			metrics.DataQualityImprovement*weights["data_quality"] +
			responseTimeImprovement*weights["response_time"]

	return overallImprovement
}

// calculateStatisticalSignificance calculates statistical significance using t-test
func (atm *ABTestManager) calculateStatisticalSignificance(basicResults, enhancedResults []*ABTestResult) *StatisticalSignificance {
	significance := &StatisticalSignificance{
		ConfidenceLevel: 0.95, // 95% confidence level
	}

	if len(basicResults) < 2 || len(enhancedResults) < 2 {
		// Not enough data for statistical significance
		significance.IsSignificant = false
		return significance
	}

	// Calculate p-values for different metrics
	significance.SuccessRatePValue = atm.calculateSuccessRatePValue(basicResults, enhancedResults)
	significance.ResponseTimePValue = atm.calculateResponseTimePValue(basicResults, enhancedResults)
	significance.AccuracyPValue = atm.calculateAccuracyPValue(basicResults, enhancedResults)
	significance.DataQualityPValue = atm.calculateDataQualityPValue(basicResults, enhancedResults)

	// Determine if results are statistically significant
	alpha := 1.0 - significance.ConfidenceLevel
	significance.IsSignificant =
		significance.SuccessRatePValue < alpha ||
			significance.ResponseTimePValue < alpha ||
			significance.AccuracyPValue < alpha ||
			significance.DataQualityPValue < alpha

	return significance
}

// calculateSuccessRatePValue calculates p-value for success rate comparison
func (atm *ABTestManager) calculateSuccessRatePValue(basicResults, enhancedResults []*ABTestResult) float64 {
	// Simplified chi-square test for proportions
	basicSuccess := 0
	enhancedSuccess := 0

	for _, result := range basicResults {
		if result.Success {
			basicSuccess++
		}
	}

	for _, result := range enhancedResults {
		if result.Success {
			enhancedSuccess++
		}
	}

	basicTotal := len(basicResults)
	enhancedTotal := len(enhancedResults)

	if basicTotal == 0 || enhancedTotal == 0 {
		return 1.0
	}

	basicRate := float64(basicSuccess) / float64(basicTotal)
	enhancedRate := float64(enhancedSuccess) / float64(enhancedTotal)

	// Simple z-test for proportions
	pooledRate := float64(basicSuccess+enhancedSuccess) / float64(basicTotal+enhancedTotal)
	standardError := math.Sqrt(pooledRate * (1 - pooledRate) * (1.0/float64(basicTotal) + 1.0/float64(enhancedTotal)))

	if standardError == 0 {
		return 1.0
	}

	zScore := (enhancedRate - basicRate) / standardError

	// Convert z-score to p-value (two-tailed test)
	pValue := 2 * (1 - atm.normalCDF(math.Abs(zScore)))

	return pValue
}

// calculateResponseTimePValue calculates p-value for response time comparison
func (atm *ABTestManager) calculateResponseTimePValue(basicResults, enhancedResults []*ABTestResult) float64 {
	// Extract response times
	var basicTimes, enhancedTimes []float64

	for _, result := range basicResults {
		basicTimes = append(basicTimes, float64(result.ResponseTime.Milliseconds()))
	}

	for _, result := range enhancedResults {
		enhancedTimes = append(enhancedTimes, float64(result.ResponseTime.Milliseconds()))
	}

	return atm.tTest(basicTimes, enhancedTimes)
}

// calculateAccuracyPValue calculates p-value for accuracy comparison
func (atm *ABTestManager) calculateAccuracyPValue(basicResults, enhancedResults []*ABTestResult) float64 {
	var basicAccuracies, enhancedAccuracies []float64

	for _, result := range basicResults {
		basicAccuracies = append(basicAccuracies, result.Accuracy)
	}

	for _, result := range enhancedResults {
		enhancedAccuracies = append(enhancedAccuracies, result.Accuracy)
	}

	return atm.tTest(basicAccuracies, enhancedAccuracies)
}

// calculateDataQualityPValue calculates p-value for data quality comparison
func (atm *ABTestManager) calculateDataQualityPValue(basicResults, enhancedResults []*ABTestResult) float64 {
	var basicQualities, enhancedQualities []float64

	for _, result := range basicResults {
		basicQualities = append(basicQualities, result.DataQuality)
	}

	for _, result := range enhancedResults {
		enhancedQualities = append(enhancedQualities, result.DataQuality)
	}

	return atm.tTest(basicQualities, enhancedQualities)
}

// tTest performs a two-sample t-test
func (atm *ABTestManager) tTest(sample1, sample2 []float64) float64 {
	if len(sample1) < 2 || len(sample2) < 2 {
		return 1.0
	}

	// Calculate means
	mean1 := atm.mean(sample1)
	mean2 := atm.mean(sample2)

	// Calculate variances
	var1 := atm.variance(sample1, mean1)
	var2 := atm.variance(sample2, mean2)

	// Calculate pooled standard error
	n1 := float64(len(sample1))
	n2 := float64(len(sample2))
	pooledVar := ((n1-1)*var1 + (n2-1)*var2) / (n1 + n2 - 2)
	standardError := math.Sqrt(pooledVar * (1/n1 + 1/n2))

	if standardError == 0 {
		return 1.0
	}

	// Calculate t-statistic
	tStat := (mean2 - mean1) / standardError

	// Calculate degrees of freedom
	df := n1 + n2 - 2

	// Convert to p-value (two-tailed test)
	pValue := 2 * (1 - atm.tCDF(math.Abs(tStat), df))

	return pValue
}

// mean calculates the mean of a slice of float64
func (atm *ABTestManager) mean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}

	return sum / float64(len(values))
}

// variance calculates the variance of a slice of float64
func (atm *ABTestManager) variance(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	sumSquaredDiff := 0.0
	for _, value := range values {
		diff := value - mean
		sumSquaredDiff += diff * diff
	}

	return sumSquaredDiff / float64(len(values)-1)
}

// normalCDF calculates the cumulative distribution function of the standard normal distribution
func (atm *ABTestManager) normalCDF(x float64) float64 {
	// Approximation using error function
	return 0.5 * (1 + atm.erf(x/math.Sqrt(2)))
}

// erf calculates the error function
func (atm *ABTestManager) erf(x float64) float64 {
	// Approximation using Taylor series
	if x < 0 {
		return -atm.erf(-x)
	}

	a1 := 0.254829592
	a2 := -0.284496736
	a3 := 1.421413741
	a4 := -1.453152027
	a5 := 1.061405429
	p := 0.3275911

	t := 1.0 / (1.0 + p*x)
	y := 1.0 - (((((a5*t+a4)*t)+a3)*t+a2)*t+a1)*t*math.Exp(-x*x)

	return y
}

// tCDF calculates the cumulative distribution function of the t-distribution
func (atm *ABTestManager) tCDF(t, df float64) float64 {
	// Simplified approximation for large degrees of freedom
	if df > 30 {
		return atm.normalCDF(t)
	}

	// For smaller degrees of freedom, use a simplified approximation
	// This is a basic implementation - in production, use a proper statistical library
	x := t * t / (df + t*t)
	sign := 1.0
	if t < 0 {
		sign = -1.0
	}
	return 0.5 * (1 + sign*atm.betaIncomplete(0.5, df/2, x))
}

// betaIncomplete calculates the incomplete beta function
func (atm *ABTestManager) betaIncomplete(a, b, x float64) float64 {
	// Simplified approximation - in production, use a proper statistical library
	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}

	// Use a simple approximation for small x
	if x < 0.5 {
		return math.Pow(x, a) / (a * atm.beta(a, b))
	}

	// Use symmetry for x > 0.5
	return 1 - atm.betaIncomplete(b, a, 1-x)
}

// beta calculates the beta function
func (atm *ABTestManager) beta(a, b float64) float64 {
	// Simplified approximation using gamma function
	// In production, use a proper statistical library
	return math.Exp(atm.logGamma(a) + atm.logGamma(b) - atm.logGamma(a+b))
}

// logGamma calculates the natural logarithm of the gamma function
func (atm *ABTestManager) logGamma(x float64) float64 {
	// Simplified approximation - in production, use a proper statistical library
	if x <= 0 {
		return math.Inf(-1)
	}

	// Use Stirling's approximation for large x
	if x > 10 {
		return (x-0.5)*math.Log(x) - x + 0.5*math.Log(2*math.Pi) + 1/(12*x)
	}

	// For smaller x, use a simple approximation
	return math.Log(math.Gamma(x))
}

// GetAllTestIDs returns all test IDs that have results
func (atm *ABTestManager) GetAllTestIDs() []string {
	atm.mu.RLock()
	defer atm.mu.RUnlock()

	var testIDs []string
	for testID := range atm.results {
		testIDs = append(testIDs, testID)
	}

	return testIDs
}

// ClearResults clears all results for a test ID
func (atm *ABTestManager) ClearResults(testID string) error {
	atm.mu.Lock()
	defer atm.mu.Unlock()

	delete(atm.results, testID)
	delete(atm.cache, testID)

	atm.logger.Info("Cleared A/B test results",
		zap.String("test_id", testID),
	)

	return nil
}

// ExportResults exports results to JSON format
func (atm *ABTestManager) ExportResults(testID string) ([]byte, error) {
	results, err := atm.GetResults(testID)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(results, "", "  ")
}

// ExportAnalysis exports analysis to JSON format
func (atm *ABTestManager) ExportAnalysis(ctx context.Context, testID string) ([]byte, error) {
	analysis, err := atm.GetAnalysis(ctx, testID)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(analysis, "", "  ")
}
