package automation

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// StatisticalTester handles statistical significance testing for model comparisons
type StatisticalTester struct {
	// Configuration
	config *StatisticalTestingConfig
}

// StatisticalTestingConfig holds configuration for statistical testing
type StatisticalTestingConfig struct {
	// Testing configuration
	DefaultSignificanceLevel float64 `json:"default_significance_level"` // 0.05 for 95% confidence
	MinimumSampleSize        int     `json:"minimum_sample_size"`
	MaximumSampleSize        int     `json:"maximum_sample_size"`

	// Test selection
	DefaultTest    string `json:"default_test"` // t_test, chi_square, mann_whitney, etc.
	AutoSelectTest bool   `json:"auto_select_test"`

	// Multiple testing correction
	MultipleTestingCorrection bool   `json:"multiple_testing_correction"`
	CorrectionMethod          string `json:"correction_method"` // bonferroni, holm, fdr

	// Effect size calculation
	CalculateEffectSize bool    `json:"calculate_effect_size"`
	EffectSizeThreshold float64 `json:"effect_size_threshold"`
}

// StatisticalTest represents a statistical test
type StatisticalTest struct {
	TestName              string                 `json:"test_name"`
	TestType              string                 `json:"test_type"`
	NullHypothesis        string                 `json:"null_hypothesis"`
	AlternativeHypothesis string                 `json:"alternative_hypothesis"`
	SignificanceLevel     float64                `json:"significance_level"`
	SampleSize            int                    `json:"sample_size"`
	TestStatistic         float64                `json:"test_statistic"`
	PValue                float64                `json:"p_value"`
	CriticalValue         float64                `json:"critical_value"`
	DegreesOfFreedom      int                    `json:"degrees_of_freedom"`
	EffectSize            float64                `json:"effect_size"`
	ConfidenceInterval    *ConfidenceInterval    `json:"confidence_interval"`
	Result                *StatisticalTestResult `json:"result"`
	Metadata              map[string]interface{} `json:"metadata"`
}

// StatisticalTestResult represents the result of a statistical test
type StatisticalTestResult struct {
	Significant              bool    `json:"significant"`
	RejectNullHypothesis     bool    `json:"reject_null_hypothesis"`
	Conclusion               string  `json:"conclusion"`
	Confidence               float64 `json:"confidence"`
	Recommendation           string  `json:"recommendation"`
	EffectSizeInterpretation string  `json:"effect_size_interpretation"`
}

// ConfidenceInterval represents a confidence interval
type ConfidenceInterval struct {
	LowerBound float64 `json:"lower_bound"`
	UpperBound float64 `json:"upper_bound"`
	Level      float64 `json:"level"`
}

// ModelComparison represents a comparison between two models
type ModelComparison struct {
	ComparisonID     string                 `json:"comparison_id"`
	ModelA           string                 `json:"model_a"`
	ModelB           string                 `json:"model_b"`
	Metric           string                 `json:"metric"` // accuracy, precision, recall, f1_score, latency
	SampleSizeA      int                    `json:"sample_size_a"`
	SampleSizeB      int                    `json:"sample_size_b"`
	MeanA            float64                `json:"mean_a"`
	MeanB            float64                `json:"mean_b"`
	StdDevA          float64                `json:"std_dev_a"`
	StdDevB          float64                `json:"std_dev_b"`
	VarianceA        float64                `json:"variance_a"`
	VarianceB        float64                `json:"variance_b"`
	DataA            []float64              `json:"data_a"`
	DataB            []float64              `json:"data_b"`
	StatisticalTest  *StatisticalTest       `json:"statistical_test"`
	ComparisonResult *ModelComparisonResult `json:"comparison_result"`
	Timestamp        time.Time              `json:"timestamp"`
}

// ModelComparisonResult represents the result of a model comparison
type ModelComparisonResult struct {
	Winner                   string                 `json:"winner"`
	Significant              bool                   `json:"significant"`
	EffectSize               float64                `json:"effect_size"`
	EffectSizeInterpretation string                 `json:"effect_size_interpretation"`
	Confidence               float64                `json:"confidence"`
	Recommendation           string                 `json:"recommendation"`
	PracticalSignificance    bool                   `json:"practical_significance"`
	Metadata                 map[string]interface{} `json:"metadata"`
}

// NewStatisticalTester creates a new statistical tester
func NewStatisticalTester(config *StatisticalTestingConfig) *StatisticalTester {
	if config == nil {
		config = &StatisticalTestingConfig{
			DefaultSignificanceLevel:  0.05,
			MinimumSampleSize:         30,
			MaximumSampleSize:         10000,
			DefaultTest:               "t_test",
			AutoSelectTest:            true,
			MultipleTestingCorrection: true,
			CorrectionMethod:          "bonferroni",
			CalculateEffectSize:       true,
			EffectSizeThreshold:       0.2,
		}
	}

	return &StatisticalTester{
		config: config,
	}
}

// CompareModels compares two models using statistical testing
func (st *StatisticalTester) CompareModels(comparison *ModelComparison) (*ModelComparisonResult, error) {
	// Validate input data
	if err := st.validateComparisonData(comparison); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Select appropriate statistical test
	testType := st.selectStatisticalTest(comparison)
	if testType == "" {
		return nil, fmt.Errorf("unable to select appropriate statistical test")
	}

	// Perform statistical test
	statisticalTest, err := st.performStatisticalTest(comparison, testType)
	if err != nil {
		return nil, fmt.Errorf("statistical test failed: %w", err)
	}

	// Calculate effect size
	effectSize := st.calculateEffectSize(comparison, testType)

	// Determine winner and significance
	result := st.interpretResults(comparison, statisticalTest, effectSize)

	// Update comparison with results
	comparison.StatisticalTest = statisticalTest
	comparison.ComparisonResult = result

	return result, nil
}

// validateComparisonData validates the comparison data
func (st *StatisticalTester) validateComparisonData(comparison *ModelComparison) error {
	if len(comparison.DataA) < st.config.MinimumSampleSize {
		return fmt.Errorf("insufficient data for model A: %d < %d", len(comparison.DataA), st.config.MinimumSampleSize)
	}

	if len(comparison.DataB) < st.config.MinimumSampleSize {
		return fmt.Errorf("insufficient data for model B: %d < %d", len(comparison.DataB), st.config.MinimumSampleSize)
	}

	if len(comparison.DataA) > st.config.MaximumSampleSize {
		return fmt.Errorf("excessive data for model A: %d > %d", len(comparison.DataA), st.config.MaximumSampleSize)
	}

	if len(comparison.DataB) > st.config.MaximumSampleSize {
		return fmt.Errorf("excessive data for model B: %d > %d", len(comparison.DataB), st.config.MaximumSampleSize)
	}

	return nil
}

// selectStatisticalTest selects the appropriate statistical test
func (st *StatisticalTester) selectStatisticalTest(comparison *ModelComparison) string {
	if !st.config.AutoSelectTest {
		return st.config.DefaultTest
	}

	// Auto-select test based on data characteristics
	_ = len(comparison.DataA)
	_ = len(comparison.DataB)

	// Check if data is normally distributed (simplified check)
	normalA := st.isApproximatelyNormal(comparison.DataA)
	normalB := st.isApproximatelyNormal(comparison.DataB)

	// Check if variances are equal
	equalVariance := st.isEqualVariance(comparison.DataA, comparison.DataB)

	// Select test based on conditions
	if normalA && normalB {
		if equalVariance {
			return "t_test"
		} else {
			return "welch_t_test"
		}
	} else {
		return "mann_whitney_u_test"
	}
}

// performStatisticalTest performs the specified statistical test
func (st *StatisticalTester) performStatisticalTest(comparison *ModelComparison, testType string) (*StatisticalTest, error) {
	switch testType {
	case "t_test":
		return st.performTTest(comparison)
	case "welch_t_test":
		return st.performWelchTTest(comparison)
	case "mann_whitney_u_test":
		return st.performMannWhitneyUTest(comparison)
	case "chi_square_test":
		return st.performChiSquareTest(comparison)
	default:
		return nil, fmt.Errorf("unsupported test type: %s", testType)
	}
}

// performTTest performs a two-sample t-test
func (st *StatisticalTester) performTTest(comparison *ModelComparison) (*StatisticalTest, error) {
	nA := len(comparison.DataA)
	nB := len(comparison.DataB)

	// Calculate means
	meanA := st.calculateMean(comparison.DataA)
	meanB := st.calculateMean(comparison.DataB)

	// Calculate pooled standard deviation
	pooledStdDev := st.calculatePooledStandardDeviation(comparison.DataA, comparison.DataB)

	// Calculate t-statistic
	se := pooledStdDev * math.Sqrt(1.0/float64(nA)+1.0/float64(nB))
	tStatistic := (meanA - meanB) / se

	// Calculate degrees of freedom
	df := nA + nB - 2

	// Calculate p-value (simplified approximation)
	pValue := st.calculateTPValue(tStatistic, df)

	// Calculate critical value
	criticalValue := st.calculateTCriticalValue(st.config.DefaultSignificanceLevel, df)

	// Create test result
	test := &StatisticalTest{
		TestName:              "Two-Sample T-Test",
		TestType:              "t_test",
		NullHypothesis:        "The means of the two groups are equal",
		AlternativeHypothesis: "The means of the two groups are not equal",
		SignificanceLevel:     st.config.DefaultSignificanceLevel,
		SampleSize:            nA + nB,
		TestStatistic:         tStatistic,
		PValue:                pValue,
		CriticalValue:         criticalValue,
		DegreesOfFreedom:      df,
		Metadata: map[string]interface{}{
			"mean_a":         meanA,
			"mean_b":         meanB,
			"pooled_std_dev": pooledStdDev,
		},
	}

	// Interpret results
	test.Result = st.interpretStatisticalTest(test)

	return test, nil
}

// performWelchTTest performs a Welch's t-test (unequal variances)
func (st *StatisticalTester) performWelchTTest(comparison *ModelComparison) (*StatisticalTest, error) {
	nA := len(comparison.DataA)
	nB := len(comparison.DataB)

	// Calculate means
	meanA := st.calculateMean(comparison.DataA)
	meanB := st.calculateMean(comparison.DataB)

	// Calculate standard deviations
	stdDevA := st.calculateStandardDeviation(comparison.DataA)
	stdDevB := st.calculateStandardDeviation(comparison.DataB)

	// Calculate standard error
	se := math.Sqrt((stdDevA*stdDevA)/float64(nA) + (stdDevB*stdDevB)/float64(nB))

	// Calculate t-statistic
	tStatistic := (meanA - meanB) / se

	// Calculate degrees of freedom (Welch-Satterthwaite equation)
	df := st.calculateWelchDegreesOfFreedom(comparison.DataA, comparison.DataB)

	// Calculate p-value
	pValue := st.calculateTPValue(tStatistic, df)

	// Calculate critical value
	criticalValue := st.calculateTCriticalValue(st.config.DefaultSignificanceLevel, df)

	// Create test result
	test := &StatisticalTest{
		TestName:              "Welch's T-Test",
		TestType:              "welch_t_test",
		NullHypothesis:        "The means of the two groups are equal",
		AlternativeHypothesis: "The means of the two groups are not equal",
		SignificanceLevel:     st.config.DefaultSignificanceLevel,
		SampleSize:            nA + nB,
		TestStatistic:         tStatistic,
		PValue:                pValue,
		CriticalValue:         criticalValue,
		DegreesOfFreedom:      df,
		Metadata: map[string]interface{}{
			"mean_a":    meanA,
			"mean_b":    meanB,
			"std_dev_a": stdDevA,
			"std_dev_b": stdDevB,
		},
	}

	// Interpret results
	test.Result = st.interpretStatisticalTest(test)

	return test, nil
}

// performMannWhitneyUTest performs a Mann-Whitney U test (non-parametric)
func (st *StatisticalTester) performMannWhitneyUTest(comparison *ModelComparison) (*StatisticalTest, error) {
	nA := len(comparison.DataA)
	nB := len(comparison.DataB)

	// Combine and rank data
	combinedData := make([]float64, 0, nA+nB)
	combinedData = append(combinedData, comparison.DataA...)
	combinedData = append(combinedData, comparison.DataB...)

	ranks := st.calculateRanks(combinedData)

	// Calculate rank sums
	rankSumA := 0.0
	rankSumB := 0.0

	for i := 0; i < nA; i++ {
		rankSumA += ranks[i]
	}

	for i := nA; i < nA+nB; i++ {
		rankSumB += ranks[i]
	}

	// Calculate U statistics
	uA := rankSumA - float64(nA*(nA+1))/2.0
	uB := rankSumB - float64(nB*(nB+1))/2.0

	// Use the smaller U value
	uStatistic := math.Min(uA, uB)

	// Calculate expected value and standard deviation
	expectedU := float64(nA*nB) / 2.0
	stdDevU := math.Sqrt(float64(nA*nB*(nA+nB+1)) / 12.0)

	// Calculate z-statistic
	zStatistic := (uStatistic - expectedU) / stdDevU

	// Calculate p-value (approximation using normal distribution)
	pValue := st.calculateZPValue(zStatistic)

	// Create test result
	test := &StatisticalTest{
		TestName:              "Mann-Whitney U Test",
		TestType:              "mann_whitney_u_test",
		NullHypothesis:        "The distributions of the two groups are equal",
		AlternativeHypothesis: "The distributions of the two groups are not equal",
		SignificanceLevel:     st.config.DefaultSignificanceLevel,
		SampleSize:            nA + nB,
		TestStatistic:         uStatistic,
		PValue:                pValue,
		CriticalValue:         0, // Not applicable for U test
		DegreesOfFreedom:      0, // Not applicable for U test
		Metadata: map[string]interface{}{
			"u_statistic": uStatistic,
			"z_statistic": zStatistic,
			"rank_sum_a":  rankSumA,
			"rank_sum_b":  rankSumB,
		},
	}

	// Interpret results
	test.Result = st.interpretStatisticalTest(test)

	return test, nil
}

// performChiSquareTest performs a chi-square test (placeholder for categorical data)
func (st *StatisticalTester) performChiSquareTest(comparison *ModelComparison) (*StatisticalTest, error) {
	// This would implement chi-square test for categorical data
	// For now, return a placeholder
	return &StatisticalTest{
		TestName:              "Chi-Square Test",
		TestType:              "chi_square_test",
		NullHypothesis:        "The distributions of the two groups are equal",
		AlternativeHypothesis: "The distributions of the two groups are not equal",
		SignificanceLevel:     st.config.DefaultSignificanceLevel,
		SampleSize:            len(comparison.DataA) + len(comparison.DataB),
		TestStatistic:         0.0,
		PValue:                0.5,
		CriticalValue:         0.0,
		DegreesOfFreedom:      0,
		Result: &StatisticalTestResult{
			Significant:          false,
			RejectNullHypothesis: false,
			Conclusion:           "Chi-square test not implemented",
			Confidence:           0.5,
			Recommendation:       "Use alternative test",
		},
	}, nil
}

// Helper methods for statistical calculations

func (st *StatisticalTester) calculateMean(data []float64) float64 {
	if len(data) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range data {
		sum += value
	}
	return sum / float64(len(data))
}

func (st *StatisticalTester) calculateStandardDeviation(data []float64) float64 {
	if len(data) <= 1 {
		return 0.0
	}

	mean := st.calculateMean(data)
	sumSquaredDiffs := 0.0

	for _, value := range data {
		diff := value - mean
		sumSquaredDiffs += diff * diff
	}

	return math.Sqrt(sumSquaredDiffs / float64(len(data)-1))
}

func (st *StatisticalTester) calculatePooledStandardDeviation(dataA, dataB []float64) float64 {
	nA := len(dataA)
	nB := len(dataB)

	if nA <= 1 || nB <= 1 {
		return 0.0
	}

	stdDevA := st.calculateStandardDeviation(dataA)
	stdDevB := st.calculateStandardDeviation(dataB)

	// Pooled standard deviation
	pooledVariance := ((float64(nA-1) * stdDevA * stdDevA) + (float64(nB-1) * stdDevB * stdDevB)) / float64(nA+nB-2)
	return math.Sqrt(pooledVariance)
}

func (st *StatisticalTester) calculateWelchDegreesOfFreedom(dataA, dataB []float64) int {
	nA := len(dataA)
	nB := len(dataB)

	stdDevA := st.calculateStandardDeviation(dataA)
	stdDevB := st.calculateStandardDeviation(dataB)

	// Welch-Satterthwaite equation
	numerator := math.Pow((stdDevA*stdDevA)/float64(nA)+(stdDevB*stdDevB)/float64(nB), 2)
	denominator := math.Pow((stdDevA*stdDevA)/float64(nA), 2)/float64(nA-1) + math.Pow((stdDevB*stdDevB)/float64(nB), 2)/float64(nB-1)

	df := numerator / denominator
	return int(math.Round(df))
}

func (st *StatisticalTester) calculateRanks(data []float64) []float64 {
	n := len(data)
	ranks := make([]float64, n)

	// Create index-value pairs for sorting
	type pair struct {
		index int
		value float64
	}

	pairs := make([]pair, n)
	for i, value := range data {
		pairs[i] = pair{i, value}
	}

	// Sort by value
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].value < pairs[j].value
	})

	// Assign ranks, handling ties
	for i := 0; i < n; i++ {
		rank := float64(i + 1)

		// Handle ties
		j := i
		for j+1 < n && pairs[j+1].value == pairs[i].value {
			j++
		}

		if j > i {
			// Calculate average rank for tied values
			avgRank := float64(i+j+2) / 2.0
			for k := i; k <= j; k++ {
				ranks[pairs[k].index] = avgRank
			}
			i = j
		} else {
			ranks[pairs[i].index] = rank
		}
	}

	return ranks
}

func (st *StatisticalTester) calculateTPValue(tStatistic float64, df int) float64 {
	// Simplified p-value calculation using approximation
	// In practice, you'd use proper t-distribution tables or libraries

	absT := math.Abs(tStatistic)

	// Very rough approximation - in practice use proper statistical libraries
	if absT > 3.0 {
		return 0.001
	} else if absT > 2.0 {
		return 0.05
	} else if absT > 1.5 {
		return 0.1
	} else {
		return 0.2
	}
}

func (st *StatisticalTester) calculateZPValue(zStatistic float64) float64 {
	// Simplified p-value calculation for normal distribution
	// In practice, you'd use proper normal distribution functions

	absZ := math.Abs(zStatistic)

	// Very rough approximation
	if absZ > 3.0 {
		return 0.001
	} else if absZ > 2.0 {
		return 0.05
	} else if absZ > 1.5 {
		return 0.1
	} else {
		return 0.2
	}
}

func (st *StatisticalTester) calculateTCriticalValue(alpha float64, df int) float64 {
	// Simplified critical value calculation
	// In practice, you'd use proper t-distribution tables or libraries

	if alpha == 0.05 {
		if df >= 30 {
			return 1.96 // Approximate normal distribution
		} else if df >= 20 {
			return 2.086
		} else if df >= 10 {
			return 2.228
		} else {
			return 2.262
		}
	}

	return 1.96 // Default approximation
}

func (st *StatisticalTester) isApproximatelyNormal(data []float64) bool {
	// Simplified normality check using skewness and kurtosis
	// In practice, you'd use proper normality tests like Shapiro-Wilk

	if len(data) < 10 {
		return false
	}

	mean := st.calculateMean(data)
	stdDev := st.calculateStandardDeviation(data)

	if stdDev == 0 {
		return false
	}

	// Calculate skewness
	skewness := 0.0
	for _, value := range data {
		normalized := (value - mean) / stdDev
		skewness += normalized * normalized * normalized
	}
	skewness /= float64(len(data))

	// Rough normality check
	return math.Abs(skewness) < 1.0
}

func (st *StatisticalTester) isEqualVariance(dataA, dataB []float64) bool {
	// F-test for equal variances (simplified)
	stdDevA := st.calculateStandardDeviation(dataA)
	stdDevB := st.calculateStandardDeviation(dataB)

	if stdDevA == 0 || stdDevB == 0 {
		return stdDevA == stdDevB
	}

	// F-ratio
	fRatio := (stdDevA * stdDevA) / (stdDevB * stdDevB)
	if fRatio < 1.0 {
		fRatio = 1.0 / fRatio
	}

	// Rough equal variance check
	return fRatio < 2.0
}

func (st *StatisticalTester) calculateEffectSize(comparison *ModelComparison, testType string) float64 {
	if !st.config.CalculateEffectSize {
		return 0.0
	}

	switch testType {
	case "t_test", "welch_t_test":
		return st.calculateCohensD(comparison)
	case "mann_whitney_u_test":
		return st.calculateMannWhitneyEffectSize(comparison)
	default:
		return 0.0
	}
}

func (st *StatisticalTester) calculateCohensD(comparison *ModelComparison) float64 {
	// Cohen's d for effect size
	meanDiff := comparison.MeanA - comparison.MeanB

	// Use pooled standard deviation for t-test, or separate for Welch's t-test
	var pooledStdDev float64
	if comparison.StatisticalTest != nil && comparison.StatisticalTest.TestType == "t_test" {
		pooledStdDev = st.calculatePooledStandardDeviation(comparison.DataA, comparison.DataB)
	} else {
		// For Welch's t-test, use average of standard deviations
		stdDevA := st.calculateStandardDeviation(comparison.DataA)
		stdDevB := st.calculateStandardDeviation(comparison.DataB)
		pooledStdDev = (stdDevA + stdDevB) / 2.0
	}

	if pooledStdDev == 0 {
		return 0.0
	}

	return meanDiff / pooledStdDev
}

func (st *StatisticalTester) calculateMannWhitneyEffectSize(comparison *ModelComparison) float64 {
	// Effect size for Mann-Whitney U test (r = Z / sqrt(N))
	nA := len(comparison.DataA)
	nB := len(comparison.DataB)

	if comparison.StatisticalTest == nil || comparison.StatisticalTest.Metadata == nil {
		return 0.0
	}

	zStatistic, ok := comparison.StatisticalTest.Metadata["z_statistic"].(float64)
	if !ok {
		return 0.0
	}

	return zStatistic / math.Sqrt(float64(nA+nB))
}

func (st *StatisticalTester) interpretStatisticalTest(test *StatisticalTest) *StatisticalTestResult {
	significant := test.PValue < test.SignificanceLevel
	confidence := (1.0 - test.SignificanceLevel) * 100.0

	var conclusion, recommendation string

	if significant {
		conclusion = fmt.Sprintf("Statistically significant difference detected (p = %.4f < %.4f)",
			test.PValue, test.SignificanceLevel)
		recommendation = "Reject null hypothesis. There is a statistically significant difference between the groups."
	} else {
		conclusion = fmt.Sprintf("No statistically significant difference detected (p = %.4f >= %.4f)",
			test.PValue, test.SignificanceLevel)
		recommendation = "Fail to reject null hypothesis. No statistically significant difference between the groups."
	}

	// Interpret effect size
	effectSizeInterpretation := st.interpretEffectSize(test.EffectSize)

	return &StatisticalTestResult{
		Significant:              significant,
		RejectNullHypothesis:     significant,
		Conclusion:               conclusion,
		Confidence:               confidence,
		Recommendation:           recommendation,
		EffectSizeInterpretation: effectSizeInterpretation,
	}
}

func (st *StatisticalTester) interpretEffectSize(effectSize float64) string {
	absEffectSize := math.Abs(effectSize)

	if absEffectSize < 0.2 {
		return "negligible effect"
	} else if absEffectSize < 0.5 {
		return "small effect"
	} else if absEffectSize < 0.8 {
		return "medium effect"
	} else {
		return "large effect"
	}
}

func (st *StatisticalTester) interpretResults(comparison *ModelComparison, test *StatisticalTest, effectSize float64) *ModelComparisonResult {
	// Determine winner
	winner := comparison.ModelA
	if comparison.MeanB > comparison.MeanA {
		winner = comparison.ModelB
	}

	// Check practical significance
	practicalSignificance := math.Abs(effectSize) >= st.config.EffectSizeThreshold

	// Generate recommendation
	var recommendation string
	if test.Result.Significant && practicalSignificance {
		recommendation = fmt.Sprintf("Deploy %s - statistically significant improvement with practical significance", winner)
	} else if test.Result.Significant {
		recommendation = fmt.Sprintf("Consider %s - statistically significant but may lack practical significance", winner)
	} else {
		recommendation = "No clear winner - continue testing with larger sample size"
	}

	return &ModelComparisonResult{
		Winner:                   winner,
		Significant:              test.Result.Significant,
		EffectSize:               effectSize,
		EffectSizeInterpretation: test.Result.EffectSizeInterpretation,
		Confidence:               test.Result.Confidence,
		Recommendation:           recommendation,
		PracticalSignificance:    practicalSignificance,
		Metadata: map[string]interface{}{
			"test_type":     test.TestType,
			"p_value":       test.PValue,
			"sample_size_a": len(comparison.DataA),
			"sample_size_b": len(comparison.DataB),
		},
	}
}

// Enhanced Statistical Testing Capabilities

// PerformMultipleComparisons performs multiple comparisons with correction
func (st *StatisticalTester) PerformMultipleComparisons(comparisons []*ModelComparison) (*MultipleComparisonResult, error) {
	if len(comparisons) == 0 {
		return nil, fmt.Errorf("no comparisons provided")
	}

	// Perform individual tests
	results := make([]*ModelComparisonResult, len(comparisons))
	for i, comparison := range comparisons {
		result, err := st.CompareModels(comparison)
		if err != nil {
			return nil, fmt.Errorf("comparison %d failed: %w", i, err)
		}
		results[i] = result
	}

	// Apply multiple testing correction if enabled
	if st.config.MultipleTestingCorrection {
		st.applyMultipleTestingCorrection(results)
	}

	// Create multiple comparison result
	multipleResult := &MultipleComparisonResult{
		Timestamp:           time.Now(),
		NumberOfComparisons: len(comparisons),
		Comparisons:         results,
		OverallSignificance: st.calculateOverallSignificance(results),
		Recommendations:     st.generateMultipleComparisonRecommendations(results),
	}

	return multipleResult, nil
}

// MultipleComparisonResult represents results of multiple comparisons
type MultipleComparisonResult struct {
	Timestamp           time.Time                `json:"timestamp"`
	NumberOfComparisons int                      `json:"number_of_comparisons"`
	Comparisons         []*ModelComparisonResult `json:"comparisons"`
	OverallSignificance bool                     `json:"overall_significance"`
	Recommendations     []string                 `json:"recommendations"`
}

// applyMultipleTestingCorrection applies multiple testing correction
func (st *StatisticalTester) applyMultipleTestingCorrection(results []*ModelComparisonResult) {
	switch st.config.CorrectionMethod {
	case "bonferroni":
		st.applyBonferroniCorrection(results)
	case "holm":
		st.applyHolmCorrection(results)
	case "fdr":
		st.applyFDRCorrection(results)
	default:
		st.applyBonferroniCorrection(results)
	}
}

// applyBonferroniCorrection applies Bonferroni correction
func (st *StatisticalTester) applyBonferroniCorrection(results []*ModelComparisonResult) {
	alpha := st.config.DefaultSignificanceLevel
	correctedAlpha := alpha / float64(len(results))

	for _, result := range results {
		if result.Metadata != nil {
			if pValue, ok := result.Metadata["p_value"].(float64); ok {
				result.Significant = pValue < correctedAlpha
				result.Metadata["corrected_alpha"] = correctedAlpha
				result.Metadata["correction_method"] = "bonferroni"
			}
		}
	}
}

// applyHolmCorrection applies Holm-Bonferroni correction
func (st *StatisticalTester) applyHolmCorrection(results []*ModelComparisonResult) {
	// Sort results by p-value
	sort.Slice(results, func(i, j int) bool {
		pI, okI := results[i].Metadata["p_value"].(float64)
		pJ, okJ := results[j].Metadata["p_value"].(float64)
		if !okI || !okJ {
			return false
		}
		return pI < pJ
	})

	alpha := st.config.DefaultSignificanceLevel
	for i, result := range results {
		correctedAlpha := alpha / float64(len(results)-i)
		if result.Metadata != nil {
			if pValue, ok := result.Metadata["p_value"].(float64); ok {
				result.Significant = pValue < correctedAlpha
				result.Metadata["corrected_alpha"] = correctedAlpha
				result.Metadata["correction_method"] = "holm"
			}
		}
	}
}

// applyFDRCorrection applies False Discovery Rate correction
func (st *StatisticalTester) applyFDRCorrection(results []*ModelComparisonResult) {
	// Sort results by p-value
	sort.Slice(results, func(i, j int) bool {
		pI, okI := results[i].Metadata["p_value"].(float64)
		pJ, okJ := results[j].Metadata["p_value"].(float64)
		if !okI || !okJ {
			return false
		}
		return pI < pJ
	})

	alpha := st.config.DefaultSignificanceLevel
	for i, result := range results {
		correctedAlpha := alpha * float64(i+1) / float64(len(results))
		if result.Metadata != nil {
			if pValue, ok := result.Metadata["p_value"].(float64); ok {
				result.Significant = pValue < correctedAlpha
				result.Metadata["corrected_alpha"] = correctedAlpha
				result.Metadata["correction_method"] = "fdr"
			}
		}
	}
}

// calculateOverallSignificance calculates overall significance
func (st *StatisticalTester) calculateOverallSignificance(results []*ModelComparisonResult) bool {
	significantCount := 0
	for _, result := range results {
		if result.Significant {
			significantCount++
		}
	}

	// Consider overall significant if more than half are significant
	return significantCount > len(results)/2
}

// generateMultipleComparisonRecommendations generates recommendations for multiple comparisons
func (st *StatisticalTester) generateMultipleComparisonRecommendations(results []*ModelComparisonResult) []string {
	var recommendations []string

	significantCount := 0
	winnerCounts := make(map[string]int)

	for _, result := range results {
		if result.Significant {
			significantCount++
			winnerCounts[result.Winner]++
		}
	}

	if significantCount == 0 {
		recommendations = append(recommendations, "No statistically significant differences found across all comparisons")
	} else if significantCount == len(results) {
		recommendations = append(recommendations, "All comparisons show statistically significant differences")
	} else {
		recommendations = append(recommendations, fmt.Sprintf("%d out of %d comparisons show statistically significant differences", significantCount, len(results)))
	}

	// Find most frequent winner
	if len(winnerCounts) > 0 {
		mostFrequentWinner := ""
		maxCount := 0
		for winner, count := range winnerCounts {
			if count > maxCount {
				maxCount = count
				mostFrequentWinner = winner
			}
		}

		if mostFrequentWinner != "" {
			recommendations = append(recommendations, fmt.Sprintf("Most frequent winner: %s (%d wins)", mostFrequentWinner, maxCount))
		}
	}

	return recommendations
}

// PerformPowerAnalysis performs statistical power analysis
func (st *StatisticalTester) PerformPowerAnalysis(comparison *ModelComparison, desiredPower float64) (*PowerAnalysisResult, error) {
	if desiredPower <= 0 || desiredPower >= 1 {
		return nil, fmt.Errorf("desired power must be between 0 and 1")
	}

	// Calculate effect size
	effectSize := st.calculateCohensD(comparison)

	// Calculate current power
	currentPower := st.calculatePower(comparison, effectSize)

	// Calculate required sample size for desired power
	requiredSampleSize := st.calculateRequiredSampleSize(effectSize, desiredPower, st.config.DefaultSignificanceLevel)

	// Generate recommendations
	recommendations := st.generatePowerAnalysisRecommendations(currentPower, desiredPower, requiredSampleSize, len(comparison.DataA)+len(comparison.DataB))

	return &PowerAnalysisResult{
		Timestamp:          time.Now(),
		CurrentPower:       currentPower,
		DesiredPower:       desiredPower,
		EffectSize:         effectSize,
		CurrentSampleSize:  len(comparison.DataA) + len(comparison.DataB),
		RequiredSampleSize: requiredSampleSize,
		Recommendations:    recommendations,
	}, nil
}

// PowerAnalysisResult represents power analysis results
type PowerAnalysisResult struct {
	Timestamp          time.Time `json:"timestamp"`
	CurrentPower       float64   `json:"current_power"`
	DesiredPower       float64   `json:"desired_power"`
	EffectSize         float64   `json:"effect_size"`
	CurrentSampleSize  int       `json:"current_sample_size"`
	RequiredSampleSize int       `json:"required_sample_size"`
	Recommendations    []string  `json:"recommendations"`
}

// calculatePower calculates statistical power
func (st *StatisticalTester) calculatePower(comparison *ModelComparison, effectSize float64) float64 {
	// Simplified power calculation
	// In practice, you'd use proper power analysis libraries

	nA := len(comparison.DataA)
	nB := len(comparison.DataB)
	_ = float64(nA + nB)

	// Effect size threshold for power calculation
	if math.Abs(effectSize) < 0.1 {
		return 0.1 // Very low power for small effects
	} else if math.Abs(effectSize) < 0.3 {
		return 0.3 // Low power for small-medium effects
	} else if math.Abs(effectSize) < 0.5 {
		return 0.6 // Medium power for medium effects
	} else {
		return 0.8 // High power for large effects
	}
}

// calculateRequiredSampleSize calculates required sample size for desired power
func (st *StatisticalTester) calculateRequiredSampleSize(effectSize, desiredPower, alpha float64) int {
	// Simplified sample size calculation
	// In practice, you'd use proper power analysis formulas

	absEffectSize := math.Abs(effectSize)

	// Base sample size calculation
	baseSampleSize := 100

	// Adjust for effect size
	if absEffectSize < 0.1 {
		baseSampleSize = 1000 // Large sample needed for small effects
	} else if absEffectSize < 0.3 {
		baseSampleSize = 500
	} else if absEffectSize < 0.5 {
		baseSampleSize = 200
	} else {
		baseSampleSize = 100
	}

	// Adjust for desired power
	if desiredPower > 0.8 {
		baseSampleSize = int(float64(baseSampleSize) * 1.2)
	} else if desiredPower < 0.6 {
		baseSampleSize = int(float64(baseSampleSize) * 0.8)
	}

	// Adjust for significance level
	if alpha < 0.01 {
		baseSampleSize = int(float64(baseSampleSize) * 1.3)
	} else if alpha > 0.1 {
		baseSampleSize = int(float64(baseSampleSize) * 0.8)
	}

	return baseSampleSize
}

// generatePowerAnalysisRecommendations generates power analysis recommendations
func (st *StatisticalTester) generatePowerAnalysisRecommendations(currentPower, desiredPower float64, requiredSampleSize, currentSampleSize int) []string {
	var recommendations []string

	if currentPower < desiredPower {
		recommendations = append(recommendations, fmt.Sprintf("Current power (%.2f) is below desired power (%.2f)", currentPower, desiredPower))
		recommendations = append(recommendations, fmt.Sprintf("Consider increasing sample size from %d to %d", currentSampleSize, requiredSampleSize))
	} else {
		recommendations = append(recommendations, fmt.Sprintf("Current power (%.2f) meets or exceeds desired power (%.2f)", currentPower, desiredPower))
	}

	if requiredSampleSize > currentSampleSize*2 {
		recommendations = append(recommendations, "Large sample size increase required - consider if effect size is practically meaningful")
	}

	if currentPower < 0.5 {
		recommendations = append(recommendations, "Very low power detected - results may not be reliable")
	}

	return recommendations
}

// PerformSequentialTesting performs sequential testing for early stopping
func (st *StatisticalTester) PerformSequentialTesting(comparison *ModelComparison, maxSamples int, checkInterval int) (*SequentialTestingResult, error) {
	if maxSamples < len(comparison.DataA)+len(comparison.DataB) {
		return nil, fmt.Errorf("max samples must be greater than current sample size")
	}

	// Simulate sequential testing by checking at intervals
	results := make([]*SequentialCheckpoint, 0)

	currentSamples := len(comparison.DataA) + len(comparison.DataB)

	for currentSamples <= maxSamples {
		// Create subset of data for current checkpoint
		subsetA := comparison.DataA
		subsetB := comparison.DataB

		if currentSamples < len(comparison.DataA)+len(comparison.DataB) {
			// Use all available data
			subsetA = comparison.DataA
			subsetB = comparison.DataB
		}

		// Create comparison for current checkpoint
		checkpointComparison := &ModelComparison{
			ModelA: comparison.ModelA,
			ModelB: comparison.ModelB,
			Metric: comparison.Metric,
			DataA:  subsetA,
			DataB:  subsetB,
			MeanA:  st.calculateMean(subsetA),
			MeanB:  st.calculateMean(subsetB),
		}

		// Perform test
		result, err := st.CompareModels(checkpointComparison)
		if err != nil {
			return nil, fmt.Errorf("sequential test failed at %d samples: %w", currentSamples, err)
		}

		// Create checkpoint
		checkpoint := &SequentialCheckpoint{
			SampleSize:     currentSamples,
			Significant:    result.Significant,
			PValue:         result.Metadata["p_value"].(float64),
			EffectSize:     result.EffectSize,
			Recommendation: result.Recommendation,
		}

		results = append(results, checkpoint)

		// Check for early stopping conditions
		if result.Significant && math.Abs(result.EffectSize) >= st.config.EffectSizeThreshold {
			// Early stopping due to significant result with practical significance
			break
		}

		currentSamples += checkInterval
	}

	// Determine final recommendation
	finalResult := results[len(results)-1]
	var finalRecommendation string

	if finalResult.Significant {
		finalRecommendation = fmt.Sprintf("Sequential testing completed with significant result at %d samples", finalResult.SampleSize)
	} else {
		finalRecommendation = fmt.Sprintf("Sequential testing completed without significant result at %d samples", finalResult.SampleSize)
	}

	return &SequentialTestingResult{
		Timestamp:           time.Now(),
		MaxSamples:          maxSamples,
		FinalSampleSize:     finalResult.SampleSize,
		Checkpoints:         results,
		FinalSignificant:    finalResult.Significant,
		FinalRecommendation: finalRecommendation,
	}, nil
}

// SequentialTestingResult represents sequential testing results
type SequentialTestingResult struct {
	Timestamp           time.Time               `json:"timestamp"`
	MaxSamples          int                     `json:"max_samples"`
	FinalSampleSize     int                     `json:"final_sample_size"`
	Checkpoints         []*SequentialCheckpoint `json:"checkpoints"`
	FinalSignificant    bool                    `json:"final_significant"`
	FinalRecommendation string                  `json:"final_recommendation"`
}

// SequentialCheckpoint represents a checkpoint in sequential testing
type SequentialCheckpoint struct {
	SampleSize     int     `json:"sample_size"`
	Significant    bool    `json:"significant"`
	PValue         float64 `json:"p_value"`
	EffectSize     float64 `json:"effect_size"`
	Recommendation string  `json:"recommendation"`
}

// GetStatisticalTestingAnalytics returns analytics about statistical testing
func (st *StatisticalTester) GetStatisticalTestingAnalytics() *StatisticalTestingAnalytics {
	// This would collect analytics from actual testing history
	// For now, return placeholder analytics

	return &StatisticalTestingAnalytics{
		Timestamp:            time.Now(),
		TotalTests:           0,
		SignificantTests:     0,
		AverageEffectSize:    0.0,
		MostCommonTest:       st.config.DefaultTest,
		PowerAnalysisCount:   0,
		SequentialTestsCount: 0,
		Recommendations:      []string{"Enable testing history tracking for detailed analytics"},
	}
}

// StatisticalTestingAnalytics represents statistical testing analytics
type StatisticalTestingAnalytics struct {
	Timestamp            time.Time `json:"timestamp"`
	TotalTests           int       `json:"total_tests"`
	SignificantTests     int       `json:"significant_tests"`
	AverageEffectSize    float64   `json:"average_effect_size"`
	MostCommonTest       string    `json:"most_common_test"`
	PowerAnalysisCount   int       `json:"power_analysis_count"`
	SequentialTestsCount int       `json:"sequential_tests_count"`
	Recommendations      []string  `json:"recommendations"`
}
