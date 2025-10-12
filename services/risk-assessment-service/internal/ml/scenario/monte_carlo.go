package scenario

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// MonteCarloSimulator performs Monte Carlo simulations for risk scenarios
type MonteCarloSimulator struct {
	iterations int
	logger     *zap.Logger
}

// NewMonteCarloSimulator creates a new Monte Carlo simulator
func NewMonteCarloSimulator(iterations int, logger *zap.Logger) *MonteCarloSimulator {
	if iterations <= 0 {
		iterations = 10000 // Default iterations
	}

	return &MonteCarloSimulator{
		iterations: iterations,
		logger:     logger,
	}
}

// GetIterations returns the current number of iterations
func (mcs *MonteCarloSimulator) GetIterations() int {
	return mcs.iterations
}

// SimulationResult represents the result of a Monte Carlo simulation
type SimulationResult struct {
	ScenarioName        string                 `json:"scenario_name"`
	Iterations          int                    `json:"iterations"`
	MeanRiskScore       float64                `json:"mean_risk_score"`
	MedianRiskScore     float64                `json:"median_risk_score"`
	StandardDeviation   float64                `json:"standard_deviation"`
	MinRiskScore        float64                `json:"min_risk_score"`
	MaxRiskScore        float64                `json:"max_risk_score"`
	Percentiles         map[string]float64     `json:"percentiles"`
	RiskDistribution    []RiskDistribution     `json:"risk_distribution"`
	ConfidenceIntervals []ConfidenceInterval   `json:"confidence_intervals"`
	SimulationData      []float64              `json:"simulation_data,omitempty"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// RiskDistribution represents the distribution of risk scores
type RiskDistribution struct {
	RiskLevel  string  `json:"risk_level"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// ConfidenceInterval represents a confidence interval
type ConfidenceInterval struct {
	Level      float64 `json:"level"`
	LowerBound float64 `json:"lower_bound"`
	UpperBound float64 `json:"upper_bound"`
}

// ScenarioParameters defines parameters for scenario simulation
type ScenarioParameters struct {
	BaseRiskScore      float64                       `json:"base_risk_score"`
	Volatility         float64                       `json:"volatility"`
	Trend              float64                       `json:"trend"`
	ShockProbability   float64                       `json:"shock_probability"`
	ShockMagnitude     float64                       `json:"shock_magnitude"`
	CorrelationFactors map[string]float64            `json:"correlation_factors"`
	TimeHorizon        int                           `json:"time_horizon"`
	BusinessFactors    *models.RiskAssessmentRequest `json:"business_factors"`
}

// RunSimulation runs a Monte Carlo simulation for a given scenario
func (mcs *MonteCarloSimulator) RunSimulation(ctx context.Context, scenarioName string, params *ScenarioParameters) (*SimulationResult, error) {
	mcs.logger.Info("Running Monte Carlo simulation",
		zap.String("scenario", scenarioName),
		zap.Int("iterations", mcs.iterations))

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Run simulations
	simulationData := make([]float64, mcs.iterations)

	for i := 0; i < mcs.iterations; i++ {
		riskScore := mcs.simulateSingleIteration(params)
		simulationData[i] = riskScore
	}

	// Calculate statistics
	result := mcs.calculateStatistics(scenarioName, simulationData, params)

	mcs.logger.Info("Monte Carlo simulation completed",
		zap.String("scenario", scenarioName),
		zap.Float64("mean_risk_score", result.MeanRiskScore),
		zap.Float64("standard_deviation", result.StandardDeviation))

	return result, nil
}

// simulateSingleIteration simulates a single iteration of the Monte Carlo simulation
func (mcs *MonteCarloSimulator) simulateSingleIteration(params *ScenarioParameters) float64 {
	// Start with base risk score
	riskScore := params.BaseRiskScore

	// Apply trend
	riskScore += params.Trend * (rand.Float64() - 0.5) * 0.1

	// Apply volatility (random walk)
	volatilityShock := rand.NormFloat64() * params.Volatility
	riskScore += volatilityShock

	// Apply correlation factors
	for factor, correlation := range params.CorrelationFactors {
		factorShock := mcs.generateCorrelatedShock(factor, correlation, params)
		riskScore += factorShock
	}

	// Apply potential shock
	if rand.Float64() < params.ShockProbability {
		shock := (rand.Float64() - 0.5) * params.ShockMagnitude
		riskScore += shock
	}

	// Apply business-specific adjustments
	riskScore = mcs.applyBusinessAdjustments(riskScore, params.BusinessFactors)

	// Ensure score is between 0 and 1
	if riskScore > 1.0 {
		riskScore = 1.0
	} else if riskScore < 0.0 {
		riskScore = 0.0
	}

	return riskScore
}

// generateCorrelatedShock generates a shock correlated with other factors
func (mcs *MonteCarloSimulator) generateCorrelatedShock(factor string, correlation float64, params *ScenarioParameters) float64 {
	// Generate base shock
	baseShock := rand.NormFloat64() * 0.1

	// Apply correlation
	correlatedShock := baseShock * correlation

	// Factor-specific adjustments
	switch factor {
	case "market_conditions":
		correlatedShock *= 0.15
	case "regulatory_environment":
		correlatedShock *= 0.1
	case "economic_indicators":
		correlatedShock *= 0.12
	case "industry_trends":
		correlatedShock *= 0.08
	default:
		correlatedShock *= 0.05
	}

	return correlatedShock
}

// applyBusinessAdjustments applies business-specific risk adjustments
func (mcs *MonteCarloSimulator) applyBusinessAdjustments(riskScore float64, business *models.RiskAssessmentRequest) float64 {
	if business == nil {
		return riskScore
	}

	// Industry-specific adjustments
	switch business.Industry {
	case "technology":
		riskScore += (rand.Float64() - 0.5) * 0.1 // Higher volatility
	case "finance":
		riskScore += (rand.Float64() - 0.5) * 0.08 // Regulatory sensitivity
	case "healthcare":
		riskScore += (rand.Float64() - 0.5) * 0.06 // Compliance sensitivity
	case "retail":
		riskScore += (rand.Float64() - 0.5) * 0.12 // Market sensitivity
	default:
		riskScore += (rand.Float64() - 0.5) * 0.05 // Default volatility
	}

	// Country-specific adjustments
	switch business.Country {
	case "US", "CA", "GB":
		riskScore -= (rand.Float64() - 0.5) * 0.03 // Lower risk in stable countries
	default:
		riskScore += (rand.Float64() - 0.5) * 0.05 // Higher risk in other countries
	}

	// Digital presence adjustments
	if business.Website != "" {
		riskScore += (rand.Float64() - 0.5) * 0.02 // Slight increase due to cyber risk
	}

	return riskScore
}

// calculateStatistics calculates statistical measures from simulation data
func (mcs *MonteCarloSimulator) calculateStatistics(scenarioName string, data []float64, params *ScenarioParameters) *SimulationResult {
	// Sort data for percentile calculations
	sortedData := make([]float64, len(data))
	copy(sortedData, data)
	sort.Float64s(sortedData)

	// Calculate basic statistics
	mean := mcs.calculateMean(data)
	median := mcs.calculateMedian(sortedData)
	stdDev := mcs.calculateStandardDeviation(data, mean)
	min := sortedData[0]
	max := sortedData[len(sortedData)-1]

	// Calculate percentiles
	percentiles := mcs.calculatePercentiles(sortedData)

	// Calculate risk distribution
	riskDistribution := mcs.calculateRiskDistribution(data)

	// Calculate confidence intervals
	confidenceIntervals := mcs.calculateConfidenceIntervals(sortedData)

	// Create metadata
	metadata := map[string]interface{}{
		"simulation_timestamp": time.Now(),
		"base_risk_score":      params.BaseRiskScore,
		"volatility":           params.Volatility,
		"trend":                params.Trend,
		"shock_probability":    params.ShockProbability,
		"shock_magnitude":      params.ShockMagnitude,
		"time_horizon":         params.TimeHorizon,
	}

	return &SimulationResult{
		ScenarioName:        scenarioName,
		Iterations:          mcs.iterations,
		MeanRiskScore:       mean,
		MedianRiskScore:     median,
		StandardDeviation:   stdDev,
		MinRiskScore:        min,
		MaxRiskScore:        max,
		Percentiles:         percentiles,
		RiskDistribution:    riskDistribution,
		ConfidenceIntervals: confidenceIntervals,
		SimulationData:      data,
		Metadata:            metadata,
	}
}

// calculateMean calculates the mean of the data
func (mcs *MonteCarloSimulator) calculateMean(data []float64) float64 {
	sum := 0.0
	for _, value := range data {
		sum += value
	}
	return sum / float64(len(data))
}

// calculateMedian calculates the median of the sorted data
func (mcs *MonteCarloSimulator) calculateMedian(sortedData []float64) float64 {
	n := len(sortedData)
	if n%2 == 0 {
		return (sortedData[n/2-1] + sortedData[n/2]) / 2.0
	}
	return sortedData[n/2]
}

// calculateStandardDeviation calculates the standard deviation
func (mcs *MonteCarloSimulator) calculateStandardDeviation(data []float64, mean float64) float64 {
	sumSquaredDiffs := 0.0
	for _, value := range data {
		diff := value - mean
		sumSquaredDiffs += diff * diff
	}
	variance := sumSquaredDiffs / float64(len(data))
	return math.Sqrt(variance)
}

// calculatePercentiles calculates various percentiles
func (mcs *MonteCarloSimulator) calculatePercentiles(sortedData []float64) map[string]float64 {
	percentiles := map[string]float64{
		"5th":  mcs.getPercentile(sortedData, 5),
		"10th": mcs.getPercentile(sortedData, 10),
		"25th": mcs.getPercentile(sortedData, 25),
		"75th": mcs.getPercentile(sortedData, 75),
		"90th": mcs.getPercentile(sortedData, 90),
		"95th": mcs.getPercentile(sortedData, 95),
		"99th": mcs.getPercentile(sortedData, 99),
	}
	return percentiles
}

// getPercentile calculates a specific percentile
func (mcs *MonteCarloSimulator) getPercentile(sortedData []float64, percentile float64) float64 {
	index := (percentile / 100.0) * float64(len(sortedData)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedData[lower]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sortedData[lower]*(1-weight) + sortedData[upper]*weight
}

// calculateRiskDistribution calculates the distribution of risk levels
func (mcs *MonteCarloSimulator) calculateRiskDistribution(data []float64) []RiskDistribution {
	distribution := []RiskDistribution{
		{RiskLevel: "low", Count: 0},
		{RiskLevel: "medium", Count: 0},
		{RiskLevel: "high", Count: 0},
		{RiskLevel: "critical", Count: 0},
	}

	total := len(data)
	for _, score := range data {
		switch {
		case score < 0.3:
			distribution[0].Count++
		case score < 0.6:
			distribution[1].Count++
		case score < 0.8:
			distribution[2].Count++
		default:
			distribution[3].Count++
		}
	}

	// Calculate percentages
	for i := range distribution {
		distribution[i].Percentage = float64(distribution[i].Count) / float64(total) * 100
	}

	return distribution
}

// calculateConfidenceIntervals calculates confidence intervals
func (mcs *MonteCarloSimulator) calculateConfidenceIntervals(sortedData []float64) []ConfidenceInterval {
	intervals := []ConfidenceInterval{
		{Level: 0.90},
		{Level: 0.95},
		{Level: 0.99},
	}

	for i := range intervals {
		alpha := 1.0 - intervals[i].Level
		lowerPercentile := (alpha / 2.0) * 100
		upperPercentile := (1.0 - alpha/2.0) * 100

		intervals[i].LowerBound = mcs.getPercentile(sortedData, lowerPercentile)
		intervals[i].UpperBound = mcs.getPercentile(sortedData, upperPercentile)
	}

	return intervals
}

// RunMultipleScenarios runs Monte Carlo simulations for multiple scenarios
func (mcs *MonteCarloSimulator) RunMultipleScenarios(ctx context.Context, scenarios map[string]*ScenarioParameters) (map[string]*SimulationResult, error) {
	mcs.logger.Info("Running multiple scenario simulations",
		zap.Int("scenario_count", len(scenarios)))

	results := make(map[string]*SimulationResult)

	for scenarioName, params := range scenarios {
		result, err := mcs.RunSimulation(ctx, scenarioName, params)
		if err != nil {
			return nil, fmt.Errorf("failed to run simulation for scenario %s: %w", scenarioName, err)
		}
		results[scenarioName] = result
	}

	return results, nil
}

// CompareScenarios compares multiple scenario results
func (mcs *MonteCarloSimulator) CompareScenarios(results map[string]*SimulationResult) (*ScenarioComparison, error) {
	if len(results) < 2 {
		return nil, fmt.Errorf("need at least 2 scenarios for comparison")
	}

	comparison := &ScenarioComparison{
		Scenarios:         results,
		ComparisonMetrics: make(map[string]map[string]float64),
		RiskRanking:       make([]ScenarioRanking, 0),
		StatisticalTests:  make(map[string]StatisticalTest),
	}

	// Calculate comparison metrics
	for scenarioName, result := range results {
		comparison.ComparisonMetrics[scenarioName] = map[string]float64{
			"mean_risk_score":    result.MeanRiskScore,
			"median_risk_score":  result.MedianRiskScore,
			"standard_deviation": result.StandardDeviation,
			"min_risk_score":     result.MinRiskScore,
			"max_risk_score":     result.MaxRiskScore,
			"95th_percentile":    result.Percentiles["95th"],
			"99th_percentile":    result.Percentiles["99th"],
		}
	}

	// Create risk ranking
	comparison.RiskRanking = mcs.createRiskRanking(results)

	// Perform statistical tests
	comparison.StatisticalTests = mcs.performStatisticalTests(results)

	return comparison, nil
}

// ScenarioComparison represents a comparison between multiple scenarios
type ScenarioComparison struct {
	Scenarios         map[string]*SimulationResult  `json:"scenarios"`
	ComparisonMetrics map[string]map[string]float64 `json:"comparison_metrics"`
	RiskRanking       []ScenarioRanking             `json:"risk_ranking"`
	StatisticalTests  map[string]StatisticalTest    `json:"statistical_tests"`
}

// ScenarioRanking represents the ranking of a scenario
type ScenarioRanking struct {
	ScenarioName string  `json:"scenario_name"`
	MeanRisk     float64 `json:"mean_risk"`
	Rank         int     `json:"rank"`
	RiskLevel    string  `json:"risk_level"`
}

// StatisticalTest represents a statistical test result
type StatisticalTest struct {
	TestName    string  `json:"test_name"`
	PValue      float64 `json:"p_value"`
	Significant bool    `json:"significant"`
	Description string  `json:"description"`
}

// createRiskRanking creates a ranking of scenarios by risk
func (mcs *MonteCarloSimulator) createRiskRanking(results map[string]*SimulationResult) []ScenarioRanking {
	rankings := make([]ScenarioRanking, 0, len(results))

	for scenarioName, result := range results {
		riskLevel := "low"
		if result.MeanRiskScore > 0.8 {
			riskLevel = "critical"
		} else if result.MeanRiskScore > 0.6 {
			riskLevel = "high"
		} else if result.MeanRiskScore > 0.4 {
			riskLevel = "medium"
		}

		rankings = append(rankings, ScenarioRanking{
			ScenarioName: scenarioName,
			MeanRisk:     result.MeanRiskScore,
			RiskLevel:    riskLevel,
		})
	}

	// Sort by mean risk (descending)
	sort.Slice(rankings, func(i, j int) bool {
		return rankings[i].MeanRisk > rankings[j].MeanRisk
	})

	// Assign ranks
	for i := range rankings {
		rankings[i].Rank = i + 1
	}

	return rankings
}

// performStatisticalTests performs statistical tests between scenarios
func (mcs *MonteCarloSimulator) performStatisticalTests(results map[string]*SimulationResult) map[string]StatisticalTest {
	tests := make(map[string]StatisticalTest)

	// Get scenario names
	scenarioNames := make([]string, 0, len(results))
	for name := range results {
		scenarioNames = append(scenarioNames, name)
	}

	// Perform pairwise comparisons
	for i := 0; i < len(scenarioNames); i++ {
		for j := i + 1; j < len(scenarioNames); j++ {
			scenario1 := scenarioNames[i]
			scenario2 := scenarioNames[j]

			testName := fmt.Sprintf("%s_vs_%s", scenario1, scenario2)

			// Simple t-test approximation
			result1 := results[scenario1]
			result2 := results[scenario2]

			// Calculate t-statistic (simplified)
			meanDiff := result1.MeanRiskScore - result2.MeanRiskScore
			seDiff := math.Sqrt((result1.StandardDeviation*result1.StandardDeviation)/float64(result1.Iterations) +
				(result2.StandardDeviation*result2.StandardDeviation)/float64(result2.Iterations))

			tStat := meanDiff / seDiff
			pValue := mcs.calculatePValue(tStat)

			tests[testName] = StatisticalTest{
				TestName:    "t_test",
				PValue:      pValue,
				Significant: pValue < 0.05,
				Description: fmt.Sprintf("Comparison between %s and %s scenarios", scenario1, scenario2),
			}
		}
	}

	return tests
}

// calculatePValue calculates an approximate p-value for t-statistic
func (mcs *MonteCarloSimulator) calculatePValue(tStat float64) float64 {
	// Simplified p-value calculation
	// In practice, you would use proper statistical libraries
	absT := math.Abs(tStat)

	if absT > 2.576 {
		return 0.01 // 99% confidence
	} else if absT > 1.96 {
		return 0.05 // 95% confidence
	} else if absT > 1.645 {
		return 0.10 // 90% confidence
	}

	return 0.20 // Not significant
}
