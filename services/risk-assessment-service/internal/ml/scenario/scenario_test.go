package scenario

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

func TestMonteCarloSimulator_RunSimulation(t *testing.T) {
	logger := zap.NewNop()
	simulator := NewMonteCarloSimulator(1000, logger)

	params := &ScenarioParameters{
		BaseRiskScore:    0.5,
		Volatility:       0.1,
		Trend:            0.05,
		ShockProbability: 0.1,
		ShockMagnitude:   0.2,
		CorrelationFactors: map[string]float64{
			"market_conditions":      0.3,
			"regulatory_environment": 0.2,
		},
		TimeHorizon: 6,
		BusinessFactors: &models.RiskAssessmentRequest{
			BusinessName: "Test Company",
			Industry:     "technology",
			Country:      "US",
		},
	}

	result, err := simulator.RunSimulation(context.Background(), "test_scenario", params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.ScenarioName != "test_scenario" {
		t.Errorf("Expected scenario name 'test_scenario', got '%s'", result.ScenarioName)
	}

	if result.Iterations != 1000 {
		t.Errorf("Expected 1000 iterations, got %d", result.Iterations)
	}

	if result.MeanRiskScore < 0 || result.MeanRiskScore > 1 {
		t.Errorf("Expected mean risk score between 0 and 1, got %f", result.MeanRiskScore)
	}

	if result.StandardDeviation < 0 {
		t.Errorf("Expected non-negative standard deviation, got %f", result.StandardDeviation)
	}

	if len(result.Percentiles) == 0 {
		t.Error("Expected percentiles to be calculated")
	}

	if len(result.RiskDistribution) == 0 {
		t.Error("Expected risk distribution to be calculated")
	}

	if len(result.ConfidenceIntervals) == 0 {
		t.Error("Expected confidence intervals to be calculated")
	}
}

func TestMonteCarloSimulator_RunMultipleScenarios(t *testing.T) {
	logger := zap.NewNop()
	simulator := NewMonteCarloSimulator(500, logger)

	scenarios := map[string]*ScenarioParameters{
		"optimistic": {
			BaseRiskScore: 0.3,
			Volatility:    0.05,
			Trend:         -0.02,
		},
		"realistic": {
			BaseRiskScore: 0.5,
			Volatility:    0.1,
			Trend:         0.0,
		},
		"pessimistic": {
			BaseRiskScore: 0.7,
			Volatility:    0.15,
			Trend:         0.05,
		},
	}

	results, err := simulator.RunMultipleScenarios(context.Background(), scenarios)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Check that optimistic scenario has lower mean risk than pessimistic
	if results["optimistic"].MeanRiskScore >= results["pessimistic"].MeanRiskScore {
		t.Error("Expected optimistic scenario to have lower mean risk than pessimistic")
	}
}

func TestMonteCarloSimulator_CompareScenarios(t *testing.T) {
	logger := zap.NewNop()
	simulator := NewMonteCarloSimulator(500, logger)

	// Create mock results
	results := map[string]*SimulationResult{
		"scenario1": {
			ScenarioName:      "scenario1",
			MeanRiskScore:     0.4,
			StandardDeviation: 0.1,
			Iterations:        500,
		},
		"scenario2": {
			ScenarioName:      "scenario2",
			MeanRiskScore:     0.6,
			StandardDeviation: 0.12,
			Iterations:        500,
		},
	}

	comparison, err := simulator.CompareScenarios(results)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if comparison == nil {
		t.Fatal("Expected comparison, got nil")
	}

	if len(comparison.Scenarios) != 2 {
		t.Errorf("Expected 2 scenarios in comparison, got %d", len(comparison.Scenarios))
	}

	if len(comparison.ComparisonMetrics) != 2 {
		t.Errorf("Expected 2 comparison metrics, got %d", len(comparison.ComparisonMetrics))
	}

	if len(comparison.RiskRanking) != 2 {
		t.Errorf("Expected 2 risk rankings, got %d", len(comparison.RiskRanking))
	}

	// Check that scenario2 is ranked higher (more risky)
	if comparison.RiskRanking[0].ScenarioName != "scenario2" {
		t.Error("Expected scenario2 to be ranked first (most risky)")
	}
}

func TestStressTester_RunStressTest(t *testing.T) {
	logger := zap.NewNop()
	tester := NewStressTester(logger)

	scenario := &StressTestScenario{
		Name:        "Test Stress Scenario",
		Description: "A test stress scenario",
		Severity:    "moderate",
		Duration:    6,
		Business: &models.RiskAssessmentRequest{
			BusinessName: "Test Company",
			Industry:     "technology",
			Country:      "US",
		},
		Factors: map[string]StressFactor{
			"market_volatility": {
				FactorName:    "Market Volatility",
				FactorType:    "market_volatility",
				BaseValue:     0.3,
				StressedValue: 0.8,
				Description:   "Increased market volatility",
				Severity:      "high",
			},
		},
	}

	baseRiskScore := 0.4
	result, err := tester.RunStressTest(context.Background(), scenario, baseRiskScore)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.TestName != "Test Stress Scenario" {
		t.Errorf("Expected test name 'Test Stress Scenario', got '%s'", result.TestName)
	}

	if result.BaseRiskScore != baseRiskScore {
		t.Errorf("Expected base risk score %f, got %f", baseRiskScore, result.BaseRiskScore)
	}

	if result.StressedRiskScore <= baseRiskScore {
		t.Error("Expected stressed risk score to be higher than base risk score")
	}

	if result.RiskIncrease <= 0 {
		t.Error("Expected positive risk increase")
	}

	if len(result.StressFactors) == 0 {
		t.Error("Expected stress factors to be present")
	}

	if result.ImpactAnalysis.OverallImpact == "" {
		t.Error("Expected impact analysis to be present")
	}

	if len(result.MitigationOptions) == 0 {
		t.Error("Expected mitigation options to be present")
	}
}

func TestStressTester_RunMultipleStressTests(t *testing.T) {
	logger := zap.NewNop()
	tester := NewStressTester(logger)

	business := &models.RiskAssessmentRequest{
		BusinessName: "Test Company",
		Industry:     "technology",
		Country:      "US",
	}

	scenarios := []*StressTestScenario{
		{
			Name:     "Scenario 1",
			Severity: "moderate",
			Duration: 3,
			Business: business,
			Factors: map[string]StressFactor{
				"factor1": {
					FactorName:    "Factor 1",
					FactorType:    "market_volatility",
					BaseValue:     0.3,
					StressedValue: 0.6,
					Description:   "Test factor 1",
					Severity:      "moderate",
				},
			},
		},
		{
			Name:     "Scenario 2",
			Severity: "severe",
			Duration: 6,
			Business: business,
			Factors: map[string]StressFactor{
				"factor2": {
					FactorName:    "Factor 2",
					FactorType:    "economic_shock",
					BaseValue:     0.4,
					StressedValue: 0.8,
					Description:   "Test factor 2",
					Severity:      "high",
				},
			},
		},
	}

	baseRiskScore := 0.4
	results, err := tester.RunMultipleStressTests(context.Background(), scenarios, baseRiskScore)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Check that severe scenario has higher stressed risk score
	if results[1].StressedRiskScore <= results[0].StressedRiskScore {
		t.Error("Expected severe scenario to have higher stressed risk score")
	}
}

func TestStressTester_CreateStandardScenarios(t *testing.T) {
	logger := zap.NewNop()
	tester := NewStressTester(logger)

	business := &models.RiskAssessmentRequest{
		BusinessName: "Test Company",
		Industry:     "technology",
		Country:      "US",
	}

	scenarios := tester.CreateStandardScenarios(business)
	if len(scenarios) == 0 {
		t.Error("Expected standard scenarios to be created")
	}

	// Check for expected scenario types
	scenarioNames := make(map[string]bool)
	for _, scenario := range scenarios {
		scenarioNames[scenario.Name] = true
	}

	expectedScenarios := []string{
		"Market Crisis",
		"Regulatory Change",
		"Operational Disruption",
		"Cybersecurity Incident",
		"Economic Recession",
		"Natural Disaster",
	}

	for _, expected := range expectedScenarios {
		if !scenarioNames[expected] {
			t.Errorf("Expected scenario '%s' to be present", expected)
		}
	}
}

func TestStressTester_ImpactAnalysis(t *testing.T) {
	logger := zap.NewNop()
	tester := NewStressTester(logger)

	// Test different impact levels
	testCases := []struct {
		baseScore      float64
		stressedScore  float64
		expectedImpact string
	}{
		{0.3, 0.35, "low"},     // 0.05 increase
		{0.3, 0.4, "moderate"}, // 0.1 increase
		{0.3, 0.5, "moderate"}, // 0.2 increase
		{0.3, 0.6, "high"},     // 0.3 increase
		{0.3, 0.8, "severe"},   // 0.5 increase
	}

	for _, tc := range testCases {
		scenario := &StressTestScenario{
			Name: "Test Scenario",
			Business: &models.RiskAssessmentRequest{
				BusinessName: "Test Company",
			},
			Factors: map[string]StressFactor{},
		}

		impact := tester.analyzeImpact(tc.baseScore, tc.stressedScore, scenario)
		if impact.OverallImpact != tc.expectedImpact {
			t.Errorf("Expected impact '%s' for scores %f->%f, got '%s'",
				tc.expectedImpact, tc.baseScore, tc.stressedScore, impact.OverallImpact)
		}
	}
}

func TestMonteCarloSimulator_Statistics(t *testing.T) {
	logger := zap.NewNop()
	simulator := NewMonteCarloSimulator(1000, logger)

	// Test with a simple simulation to verify statistics calculation
	params := &ScenarioParameters{
		BaseRiskScore: 0.5,
		Volatility:    0.1,
		Trend:         0.0,
		BusinessFactors: &models.RiskAssessmentRequest{
			BusinessName: "Test Company",
		},
	}

	result, err := simulator.RunSimulation(context.Background(), "test_stats", params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify that statistics are calculated correctly
	if result.MeanRiskScore < 0 || result.MeanRiskScore > 1 {
		t.Errorf("Expected mean risk score between 0 and 1, got %f", result.MeanRiskScore)
	}

	if result.StandardDeviation < 0 {
		t.Errorf("Expected non-negative standard deviation, got %f", result.StandardDeviation)
	}

	if len(result.Percentiles) == 0 {
		t.Error("Expected percentiles to be calculated")
	}

	// Check that percentiles are in ascending order
	if result.Percentiles["5th"] > result.Percentiles["95th"] {
		t.Error("Expected 5th percentile to be less than 95th percentile")
	}
}

func TestStressTester_SeverityMultipliers(t *testing.T) {
	logger := zap.NewNop()
	tester := NewStressTester(logger)

	testCases := []struct {
		severity string
		expected float64
	}{
		{"mild", 1.1},
		{"moderate", 1.3},
		{"severe", 1.6},
		{"extreme", 2.0},
		{"unknown", 1.2},
	}

	for _, tc := range testCases {
		multiplier := tester.getSeverityMultiplier(tc.severity)
		if multiplier != tc.expected {
			t.Errorf("Expected multiplier %f for severity '%s', got %f",
				tc.expected, tc.severity, multiplier)
		}
	}
}

func TestStressTester_DurationAdjustments(t *testing.T) {
	logger := zap.NewNop()
	tester := NewStressTester(logger)

	testCases := []struct {
		duration int
		expected float64
	}{
		{1, 0.0},
		{2, 0.05},
		{4, 0.10},
		{8, 0.15},
		{15, 0.20},
	}

	for _, tc := range testCases {
		adjustment := tester.getDurationAdjustment(tc.duration)
		if adjustment != tc.expected {
			t.Errorf("Expected adjustment %f for duration %d, got %f",
				tc.expected, tc.duration, adjustment)
		}
	}
}
