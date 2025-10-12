package scenario

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

func TestMonteCarloSimulator_Enhanced(t *testing.T) {
	logger := zap.NewNop()
	simulator := NewMonteCarloSimulator(1000, logger)

	t.Run("GetIterations", func(t *testing.T) {
		assert.Equal(t, 1000, simulator.GetIterations())
	})

	t.Run("RunSimulation with enhanced parameters", func(t *testing.T) {
		params := &ScenarioParameters{
			BaseRiskScore:    0.5,
			Volatility:       0.15,
			Trend:            0.05,
			ShockProbability: 0.1,
			ShockMagnitude:   0.2,
			TimeHorizon:      12,
			BusinessFactors: &models.RiskAssessmentRequest{
				BusinessName: "Test Company",
				Industry:     "technology",
				Country:      "US",
			},
			CorrelationFactors: map[string]float64{
				"market_conditions":      0.6,
				"regulatory_environment": 0.4,
				"economic_indicators":    0.7,
			},
		}

		result, err := simulator.RunSimulation(context.Background(), "enhanced_test", params)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "enhanced_test", result.ScenarioName)
		assert.Equal(t, 1000, result.Iterations)
		assert.GreaterOrEqual(t, result.MeanRiskScore, 0.0)
		assert.LessOrEqual(t, result.MeanRiskScore, 1.0)
		assert.GreaterOrEqual(t, result.StandardDeviation, 0.0)
		assert.NotEmpty(t, result.Percentiles)
		assert.NotEmpty(t, result.RiskDistribution)
		assert.NotEmpty(t, result.ConfidenceIntervals)
	})

	t.Run("RunMultipleScenarios with comparison", func(t *testing.T) {
		scenarios := map[string]*ScenarioParameters{
			"optimistic": {
				BaseRiskScore:    0.3,
				Volatility:       0.05,
				Trend:            -0.1,
				ShockProbability: 0.05,
				ShockMagnitude:   0.1,
				TimeHorizon:      6,
				BusinessFactors: &models.RiskAssessmentRequest{
					BusinessName: "Test Company",
					Industry:     "technology",
					Country:      "US",
				},
			},
			"pessimistic": {
				BaseRiskScore:    0.7,
				Volatility:       0.2,
				Trend:            0.1,
				ShockProbability: 0.2,
				ShockMagnitude:   0.3,
				TimeHorizon:      6,
				BusinessFactors: &models.RiskAssessmentRequest{
					BusinessName: "Test Company",
					Industry:     "technology",
					Country:      "US",
				},
			},
		}

		results, err := simulator.RunMultipleScenarios(context.Background(), scenarios)
		require.NoError(t, err)
		require.Len(t, results, 2)

		// Test comparison
		comparison, err := simulator.CompareScenarios(results)
		require.NoError(t, err)
		require.NotNil(t, comparison)

		assert.Len(t, comparison.Scenarios, 2)
		assert.NotEmpty(t, comparison.ComparisonMetrics)
		assert.NotEmpty(t, comparison.RiskRanking)
		assert.NotEmpty(t, comparison.StatisticalTests)

		// Verify risk ranking
		assert.Len(t, comparison.RiskRanking, 2)
		assert.Equal(t, 1, comparison.RiskRanking[0].Rank)
		assert.Equal(t, 2, comparison.RiskRanking[1].Rank)
	})

	t.Run("Statistical analysis accuracy", func(t *testing.T) {
		params := &ScenarioParameters{
			BaseRiskScore:    0.5,
			Volatility:       0.1,
			Trend:            0.0,
			ShockProbability: 0.0, // No shocks for consistent results
			ShockMagnitude:   0.0,
			TimeHorizon:      6,
			BusinessFactors: &models.RiskAssessmentRequest{
				BusinessName: "Test Company",
				Industry:     "technology",
				Country:      "US",
			},
		}

		result, err := simulator.RunSimulation(context.Background(), "statistical_test", params)
		require.NoError(t, err)

		// Verify statistical properties
		assert.InDelta(t, 0.5, result.MeanRiskScore, 0.1) // Should be close to base score
		assert.Greater(t, result.StandardDeviation, 0.0)  // Should have some variance
		assert.GreaterOrEqual(t, result.MinRiskScore, 0.0)
		assert.LessOrEqual(t, result.MaxRiskScore, 1.0)

		// Verify percentiles are ordered correctly
		assert.LessOrEqual(t, result.Percentiles["5th"], result.Percentiles["25th"])
		assert.LessOrEqual(t, result.Percentiles["25th"], result.Percentiles["75th"])
		assert.LessOrEqual(t, result.Percentiles["75th"], result.Percentiles["95th"])

		// Verify confidence intervals
		assert.Len(t, result.ConfidenceIntervals, 3)
		for _, ci := range result.ConfidenceIntervals {
			assert.LessOrEqual(t, ci.LowerBound, ci.UpperBound)
			assert.GreaterOrEqual(t, ci.Level, 0.0)
			assert.LessOrEqual(t, ci.Level, 1.0)
		}
	})
}

func TestStressTester_Enhanced(t *testing.T) {
	logger := zap.NewNop()
	stressTester := NewStressTester(logger)

	t.Run("RunStressTest with enhanced scenario", func(t *testing.T) {
		scenario := &StressTestScenario{
			Name:        "Enhanced Market Crisis",
			Description: "Comprehensive market crisis scenario with multiple stress factors",
			Severity:    "high",
			Duration:    6, // months
			Business: &models.RiskAssessmentRequest{
				BusinessName: "Test Company",
				Industry:     "finance",
				Country:      "US",
			},
			Factors: map[string]StressFactor{
				"market_volatility": {
					FactorName:    "market_volatility",
					FactorType:    "market_volatility",
					BaseValue:     0.1,
					StressedValue: 0.4,
					Description:   "Market volatility increases significantly",
					Severity:      "high",
				},
				"regulatory_change": {
					FactorName:    "regulatory_change",
					FactorType:    "regulatory_change",
					BaseValue:     0.05,
					StressedValue: 0.2,
					Description:   "New regulatory requirements",
					Severity:      "medium",
				},
				"economic_shock": {
					FactorName:    "economic_shock",
					FactorType:    "economic_shock",
					BaseValue:     0.0,
					StressedValue: 0.3,
					Description:   "Economic recession impact",
					Severity:      "high",
				},
			},
		}

		baseRiskScore := 0.4
		result, err := stressTester.RunStressTest(context.Background(), scenario, baseRiskScore)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Enhanced Market Crisis", result.TestName)
		assert.Equal(t, baseRiskScore, result.BaseRiskScore)
		assert.Greater(t, result.StressedRiskScore, result.BaseRiskScore)
		assert.Greater(t, result.RiskIncrease, 0.0)
		assert.Greater(t, result.RiskIncreasePercent, 0.0)
		assert.Len(t, result.StressFactors, 3)
		assert.NotEmpty(t, result.ImpactAnalysis)
		assert.NotEmpty(t, result.MitigationOptions)
	})

	t.Run("RunMultipleStressTests with standard scenarios", func(t *testing.T) {
		business := &models.RiskAssessmentRequest{
			BusinessName: "Test Company",
			Industry:     "technology",
			Country:      "US",
		}

		scenarios := stressTester.CreateStandardScenarios(business)
		require.NotEmpty(t, scenarios)

		baseRiskScore := 0.5
		results, err := stressTester.RunMultipleStressTests(context.Background(), scenarios, baseRiskScore)
		require.NoError(t, err)
		require.Len(t, results, len(scenarios))

		// Verify each result
		for i, result := range results {
			assert.NotEmpty(t, result.TestName)
			assert.Equal(t, baseRiskScore, result.BaseRiskScore)
			assert.GreaterOrEqual(t, result.StressedRiskScore, 0.0)
			assert.LessOrEqual(t, result.StressedRiskScore, 1.0)
			assert.NotEmpty(t, result.StressFactors)
			assert.NotEmpty(t, result.ImpactAnalysis)
			assert.NotEmpty(t, result.MitigationOptions)

			// Verify impact analysis
			assert.NotEmpty(t, result.ImpactAnalysis.OverallImpact)
			assert.NotEmpty(t, result.ImpactAnalysis.RiskLevelChange)
			assert.NotEmpty(t, result.ImpactAnalysis.ImpactSummary)

			// Verify mitigation options
			for _, mitigation := range result.MitigationOptions {
				assert.NotEmpty(t, mitigation.OptionName)
				assert.NotEmpty(t, mitigation.Description)
				assert.GreaterOrEqual(t, mitigation.Effectiveness, 0.0)
				assert.LessOrEqual(t, mitigation.Effectiveness, 1.0)
				assert.NotEmpty(t, mitigation.ImplementationCost)
			}

			t.Logf("Scenario %d: %s - Risk increase: %.2f%%", i+1, result.TestName, result.RiskIncreasePercent)
		}
	})

	t.Run("Business-specific stress adjustments", func(t *testing.T) {
		// Test different industries
		industries := []string{"technology", "finance", "healthcare", "retail", "manufacturing"}

		for _, industry := range industries {
			business := &models.RiskAssessmentRequest{
				BusinessName: "Test Company",
				Industry:     industry,
				Country:      "US",
			}

			scenario := &StressTestScenario{
				Name:     "Industry Test",
				Severity: "medium",
				Duration: 3, // months
				Business: business,
				Factors: map[string]StressFactor{
					"market_volatility": {
						FactorName:    "market_volatility",
						FactorType:    "market_volatility",
						BaseValue:     0.1,
						StressedValue: 0.3,
						Description:   "Market volatility stress test",
						Severity:      "medium",
					},
				},
			}

			baseRiskScore := 0.4
			result, err := stressTester.RunStressTest(context.Background(), scenario, baseRiskScore)
			require.NoError(t, err)

			// Each industry should have different stress impacts
			assert.Greater(t, result.StressedRiskScore, result.BaseRiskScore)
			t.Logf("Industry %s: Risk increase %.2f%%", industry, result.RiskIncreasePercent)
		}
	})

	t.Run("Country-specific stress adjustments", func(t *testing.T) {
		// Test different countries
		countries := []string{"US", "GB", "DE", "FR", "JP", "CN", "IN", "BR"}

		for _, country := range countries {
			business := &models.RiskAssessmentRequest{
				BusinessName: "Test Company",
				Industry:     "technology",
				Country:      country,
			}

			scenario := &StressTestScenario{
				Name:     "Country Test",
				Severity: "medium",
				Duration: 3, // months
				Business: business,
				Factors: map[string]StressFactor{
					"regulatory_change": {
						FactorName:    "regulatory_change",
						FactorType:    "regulatory_change",
						BaseValue:     0.05,
						StressedValue: 0.2,
						Description:   "Regulatory change stress test",
						Severity:      "medium",
					},
				},
			}

			baseRiskScore := 0.4
			result, err := stressTester.RunStressTest(context.Background(), scenario, baseRiskScore)
			require.NoError(t, err)

			assert.Greater(t, result.StressedRiskScore, result.BaseRiskScore)
			t.Logf("Country %s: Risk increase %.2f%%", country, result.RiskIncreasePercent)
		}
	})
}

func TestScenarioAnalysis_Integration(t *testing.T) {
	logger := zap.NewNop()
	monteCarloSimulator := NewMonteCarloSimulator(1000, logger)
	stressTester := NewStressTester(logger)

	t.Run("Comprehensive scenario analysis", func(t *testing.T) {
		business := &models.RiskAssessmentRequest{
			BusinessName: "Integration Test Company",
			Industry:     "technology",
			Country:      "US",
		}

		baseRiskScore := 0.5

		// Run Monte Carlo simulations with more distinct parameters
		monteCarloScenarios := map[string]*ScenarioParameters{
			"optimistic": {
				BaseRiskScore:    0.3,  // Lower base risk
				Volatility:       0.02, // Very low volatility
				Trend:            -0.2, // Strong decreasing trend
				ShockProbability: 0.01, // Very low shock probability
				ShockMagnitude:   0.05, // Small shocks
				TimeHorizon:      6,
				BusinessFactors:  business,
			},
			"pessimistic": {
				BaseRiskScore:    0.7, // Higher base risk
				Volatility:       0.3, // High volatility
				Trend:            0.2, // Strong increasing trend
				ShockProbability: 0.3, // High shock probability
				ShockMagnitude:   0.4, // Large shocks
				TimeHorizon:      6,
				BusinessFactors:  business,
			},
		}

		monteCarloResults, err := monteCarloSimulator.RunMultipleScenarios(context.Background(), monteCarloScenarios)
		require.NoError(t, err)

		// Run stress tests
		stressScenarios := stressTester.CreateStandardScenarios(business)
		stressResults, err := stressTester.RunMultipleStressTests(context.Background(), stressScenarios, baseRiskScore)
		require.NoError(t, err)

		// Verify integration results
		assert.Len(t, monteCarloResults, 2)
		assert.Len(t, stressResults, len(stressScenarios))

		// Verify Monte Carlo results are consistent
		for name, result := range monteCarloResults {
			assert.Equal(t, name, result.ScenarioName)
			assert.GreaterOrEqual(t, result.MeanRiskScore, 0.0)
			assert.LessOrEqual(t, result.MeanRiskScore, 1.0)
			t.Logf("Monte Carlo %s: Mean risk %.3f, StdDev %.3f", name, result.MeanRiskScore, result.StandardDeviation)
		}

		// Verify stress test results are consistent
		for _, result := range stressResults {
			assert.Greater(t, result.StressedRiskScore, result.BaseRiskScore)
			assert.Greater(t, result.RiskIncrease, 0.0)
			t.Logf("Stress Test %s: Risk increase %.2f%%", result.TestName, result.RiskIncreasePercent)
		}

		// Verify that optimistic Monte Carlo has lower risk than pessimistic
		optimisticRisk := monteCarloResults["optimistic"].MeanRiskScore
		pessimisticRisk := monteCarloResults["pessimistic"].MeanRiskScore
		assert.Less(t, optimisticRisk, pessimisticRisk)

		// Verify that stress tests show significant risk increases
		maxStressIncrease := 0.0
		for _, result := range stressResults {
			if result.RiskIncreasePercent > maxStressIncrease {
				maxStressIncrease = result.RiskIncreasePercent
			}
		}
		assert.Greater(t, maxStressIncrease, 10.0) // At least 10% increase in stress scenarios
	})

	t.Run("Performance test", func(t *testing.T) {
		business := &models.RiskAssessmentRequest{
			BusinessName: "Performance Test Company",
			Industry:     "technology",
			Country:      "US",
		}

		baseRiskScore := 0.5

		// Test Monte Carlo performance
		start := time.Now()
		params := &ScenarioParameters{
			BaseRiskScore:    baseRiskScore,
			Volatility:       0.1,
			Trend:            0.0,
			ShockProbability: 0.1,
			ShockMagnitude:   0.15,
			TimeHorizon:      6,
			BusinessFactors:  business,
		}

		_, err := monteCarloSimulator.RunSimulation(context.Background(), "performance_test", params)
		require.NoError(t, err)
		monteCarloDuration := time.Since(start)

		// Test stress testing performance
		start = time.Now()
		scenarios := stressTester.CreateStandardScenarios(business)
		_, err = stressTester.RunMultipleStressTests(context.Background(), scenarios, baseRiskScore)
		require.NoError(t, err)
		stressTestDuration := time.Since(start)

		// Performance should be reasonable
		assert.Less(t, monteCarloDuration, 5*time.Second, "Monte Carlo simulation should complete within 5 seconds")
		assert.Less(t, stressTestDuration, 2*time.Second, "Stress testing should complete within 2 seconds")

		t.Logf("Monte Carlo duration: %v", monteCarloDuration)
		t.Logf("Stress test duration: %v", stressTestDuration)
	})
}
