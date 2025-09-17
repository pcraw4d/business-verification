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

func setupTestConfidenceImprover(t *testing.T) (*ConfidenceImprover, *sql.DB) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	logger := zap.NewNop()
	industryDB := NewIndustryCodeDatabase(db, logger)
	reporter := NewConfidenceReporter(industryDB, logger)
	scorer := NewConfidenceScorer(industryDB, nil, logger)
	improver := NewConfidenceImprover(industryDB, reporter, scorer, logger)

	return improver, db
}

func TestConfidenceImprover_GenerateImprovementPlan(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	ctx := context.Background()
	config := &ImprovementConfig{
		TargetConfidence: 0.90,
		Timeline:         "12 months",
		Budget:           100000,
		Priorities:       []string{"data_quality", "algorithm_optimization"},
		Constraints:      []string{"limited_resources"},
		FocusAreas:       []string{"low_confidence_reduction"},
		RiskTolerance:    "medium",
		ResourceAvailability: map[string]float64{
			"engineering_time": 0.8,
			"data_team_time":   0.6,
		},
		SuccessMetrics: []string{"overall_confidence", "high_confidence_rate"},
	}

	plan, err := improver.GenerateImprovementPlan(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, plan)

	// Test plan structure
	assert.NotEmpty(t, plan.ID)
	assert.WithinDuration(t, time.Now(), plan.GeneratedAt, 5*time.Second)
	assert.Equal(t, config.TargetConfidence, plan.TargetMetrics.OverallConfidence)

	// Test current state analysis
	assert.NotNil(t, plan.CurrentState)
	assert.Greater(t, plan.CurrentState.OverallConfidence, 0.0)
	assert.LessOrEqual(t, plan.CurrentState.OverallConfidence, 1.0)
	assert.NotEmpty(t, plan.CurrentState.TrendDirection)

	// Test strategies
	assert.NotEmpty(t, plan.Strategies)
	for _, strategy := range plan.Strategies {
		assert.NotEmpty(t, strategy.ID)
		assert.NotEmpty(t, strategy.Name)
		assert.NotEmpty(t, strategy.Description)
		assert.NotEmpty(t, strategy.Category)
		assert.NotEmpty(t, strategy.Priority)
		assert.Greater(t, strategy.ImpactScore, 0.0)
		assert.LessOrEqual(t, strategy.ImpactScore, 1.0)
		assert.Greater(t, strategy.EffortScore, 0.0)
		assert.LessOrEqual(t, strategy.EffortScore, 1.0)
		assert.Greater(t, strategy.ROIScore, 0.0)
		assert.Greater(t, strategy.ExpectedImprovement, 0.0)
		assert.NotEmpty(t, strategy.Implementation)
	}

	// Test phases
	assert.NotEmpty(t, plan.Phases)
	for _, phase := range plan.Phases {
		assert.NotEmpty(t, phase.ID)
		assert.NotEmpty(t, phase.Name)
		assert.NotEmpty(t, phase.Description)
		assert.NotEmpty(t, phase.Duration)
		// Note: Strategies may be empty if no strategies match the phase priority
		assert.NotNil(t, phase.Strategies)
	}

	// Test timeline
	assert.NotNil(t, plan.Timeline)
	assert.NotEmpty(t, plan.Timeline.TotalDuration)
	assert.False(t, plan.Timeline.StartDate.IsZero())
	assert.False(t, plan.Timeline.EndDate.IsZero())
	assert.True(t, plan.Timeline.EndDate.After(plan.Timeline.StartDate))

	// Test budget estimate
	assert.NotNil(t, plan.BudgetEstimate)
	assert.Greater(t, plan.BudgetEstimate.TotalCost, float64(0))
	assert.NotEmpty(t, plan.BudgetEstimate.PhaseBreakdown)
	assert.NotEmpty(t, plan.BudgetEstimate.CategoryBreakdown)

	// Test risk assessment
	assert.NotNil(t, plan.RiskAssessment)
	assert.NotEmpty(t, plan.RiskAssessment.OverallRiskLevel)
	assert.NotEmpty(t, plan.RiskAssessment.IdentifiedRisks)

	// Test expected outcomes
	assert.NotNil(t, plan.ExpectedOutcomes)
	assert.Greater(t, plan.ExpectedOutcomes.ConfidenceImprovement, float64(0))
	assert.NotEmpty(t, plan.ExpectedOutcomes.TimeToValue)

	// Test monitoring plan
	assert.NotNil(t, plan.MonitoringPlan)
	assert.NotEmpty(t, plan.MonitoringPlan.Frequency)
	assert.NotEmpty(t, plan.MonitoringPlan.KeyMetrics)

	// Test recommendations
	assert.NotEmpty(t, plan.Recommendations)
	for _, rec := range plan.Recommendations {
		assert.NotEmpty(t, rec.ID)
		assert.NotEmpty(t, rec.Type)
		assert.NotEmpty(t, rec.Priority)
		assert.NotEmpty(t, rec.Title)
		assert.NotEmpty(t, rec.Description)
	}
}

func TestConfidenceImprover_GenerateImprovementPlan_DefaultConfig(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	ctx := context.Background()
	config := &ImprovementConfig{
		// Minimal config with defaults
		Timeline: "6 months",
	}

	plan, err := improver.GenerateImprovementPlan(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, plan)

	// Should still generate a valid plan with defaults
	assert.NotEmpty(t, plan.Strategies)
	assert.NotEmpty(t, plan.Phases)
	assert.Greater(t, plan.TargetMetrics.OverallConfidence, plan.CurrentState.OverallConfidence)
}

func TestConfidenceImprover_GenerateImprovementPlan_BudgetConstraint(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	ctx := context.Background()
	config := &ImprovementConfig{
		TargetConfidence: 0.95,
		Timeline:         "12 months",
		Budget:           10000, // Low budget constraint
		RiskTolerance:    "low",
	}

	plan, err := improver.GenerateImprovementPlan(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, plan)

	// Should respect budget constraint
	assert.LessOrEqual(t, plan.BudgetEstimate.TotalCost, config.Budget*1.2) // Allow some buffer

	// Should still have strategies but potentially fewer due to budget
	assert.NotEmpty(t, plan.Strategies)
}

func TestConfidenceImprover_AnalyzeCurrentState(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	ctx := context.Background()

	state, err := improver.analyzeCurrentState(ctx)
	require.NoError(t, err)
	assert.NotNil(t, state)

	// Test state structure
	assert.Greater(t, state.OverallConfidence, 0.0)
	assert.LessOrEqual(t, state.OverallConfidence, 1.0)
	assert.GreaterOrEqual(t, state.HighConfidenceRate, 0.0)
	assert.LessOrEqual(t, state.HighConfidenceRate, 1.0)
	assert.GreaterOrEqual(t, state.LowConfidenceRate, 0.0)
	assert.LessOrEqual(t, state.LowConfidenceRate, 1.0)
	assert.NotEmpty(t, state.TrendDirection)
	assert.Contains(t, []string{"improving", "declining", "stable"}, state.TrendDirection)

	// Test that analysis includes various components
	assert.NotNil(t, state.FactorScores)
	assert.NotNil(t, state.CodeTypeScores)
	assert.NotNil(t, state.IndustryScores)
	assert.NotNil(t, state.KeyWeaknesses)
	assert.NotNil(t, state.KeyStrengths)
	assert.NotNil(t, state.CriticalIssues)
}

func TestConfidenceImprover_DefineImprovementTargets(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	currentState := &ConfidenceState{
		OverallConfidence:  0.75,
		HighConfidenceRate: 0.60,
		LowConfidenceRate:  0.20,
		FactorScores: map[string]float64{
			"name_match":     0.85,
			"category_match": 0.70,
			"keyword_match":  0.60,
		},
		CodeTypeScores: map[string]float64{
			"NAICS": 0.80,
			"SIC":   0.70,
		},
		IndustryScores: map[string]float64{
			"Technology":    0.85,
			"Manufacturing": 0.70,
		},
	}

	config := &ImprovementConfig{
		TargetConfidence: 0.90,
		Timeline:         "12 months",
	}

	targets := improver.defineImprovementTargets(config, currentState)

	// Test target structure
	assert.Equal(t, config.TargetConfidence, targets.OverallConfidence)
	assert.Greater(t, targets.HighConfidenceRate, currentState.HighConfidenceRate)
	assert.Less(t, targets.LowConfidenceRate, currentState.LowConfidenceRate)

	// Test factor targets
	assert.NotEmpty(t, targets.FactorTargets)
	for factor, target := range targets.FactorTargets {
		current := currentState.FactorScores[factor]
		assert.Greater(t, target, current, "Target for %s should be higher than current", factor)
		assert.LessOrEqual(t, target, 0.95, "Target for %s should be realistic", factor)
	}

	// Test code type targets
	assert.NotEmpty(t, targets.CodeTypeTargets)
	for codeType, target := range targets.CodeTypeTargets {
		current := currentState.CodeTypeScores[codeType]
		assert.Greater(t, target, current, "Target for %s should be higher than current", codeType)
	}

	// Test industry targets
	assert.NotEmpty(t, targets.IndustryTargets)
	for industry, target := range targets.IndustryTargets {
		current := currentState.IndustryScores[industry]
		assert.Greater(t, target, current, "Target for %s should be higher than current", industry)
	}

	// Test timeline targets
	assert.NotEmpty(t, targets.TimelineTargets)
	assert.Contains(t, targets.TimelineTargets, "3_months")
	assert.Contains(t, targets.TimelineTargets, "6_months")
	assert.Contains(t, targets.TimelineTargets, "12_months")

	// Verify progressive improvement
	assert.Greater(t, targets.TimelineTargets["6_months"], targets.TimelineTargets["3_months"])
	assert.Greater(t, targets.TimelineTargets["12_months"], targets.TimelineTargets["6_months"])
	assert.Equal(t, targets.TimelineTargets["12_months"], targets.OverallConfidence)
}

func TestConfidenceImprover_IdentifyImprovementStrategies(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	ctx := context.Background()

	currentState := &ConfidenceState{
		OverallConfidence:  0.70, // Low confidence to trigger multiple strategies
		HighConfidenceRate: 0.50,
		LowConfidenceRate:  0.30,
		TrendDirection:     "declining",
		KeyWeaknesses:      []string{"training_data", "validation", "feature_quality"},
		KeyStrengths:       []string{"name_match"},
		CriticalIssues:     []string{"critically_low_confidence"},
		FactorScores: map[string]float64{
			"name_match":     0.85,
			"category_match": 0.65,
			"keyword_match":  0.55,
		},
	}

	targets := ImprovementTargets{
		OverallConfidence:  0.90,
		HighConfidenceRate: 0.80,
		LowConfidenceRate:  0.10,
	}

	config := &ImprovementConfig{
		Budget:        100000,
		RiskTolerance: "medium",
		Priorities:    []string{"data_quality", "algorithm_optimization"},
	}

	strategies, err := improver.identifyImprovementStrategies(ctx, currentState, targets, config)
	require.NoError(t, err)
	assert.NotEmpty(t, strategies)

	// Test that strategies are generated for identified needs
	categoryCount := make(map[string]int)
	for _, strategy := range strategies {
		categoryCount[strategy.Category]++

		// Test strategy structure
		assert.NotEmpty(t, strategy.ID)
		assert.NotEmpty(t, strategy.Name)
		assert.NotEmpty(t, strategy.Description)
		assert.NotEmpty(t, strategy.Category)
		assert.NotEmpty(t, strategy.Priority)
		assert.Greater(t, strategy.ImpactScore, 0.0)
		assert.LessOrEqual(t, strategy.ImpactScore, 1.0)
		assert.Greater(t, strategy.EffortScore, 0.0)
		assert.LessOrEqual(t, strategy.EffortScore, 1.0)
		assert.Greater(t, strategy.ROIScore, 0.0)
		assert.Greater(t, strategy.ExpectedImprovement, 0.0)
		assert.NotEmpty(t, strategy.TargetFactors)
		assert.NotEmpty(t, strategy.Implementation)
		assert.NotEmpty(t, strategy.Timeline)
		assert.NotEmpty(t, strategy.SuccessCriteria)
	}

	// Verify that strategies are sorted by ROI (highest first)
	for i := 1; i < len(strategies); i++ {
		assert.GreaterOrEqual(t, strategies[i-1].ROIScore, strategies[i].ROIScore,
			"Strategies should be sorted by ROI score in descending order")
	}

	// Test that multiple strategy categories are represented for low confidence
	assert.Greater(t, len(categoryCount), 1, "Should generate strategies from multiple categories")
}

func TestConfidenceImprover_StrategyCreation(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	currentState := &ConfidenceState{
		OverallConfidence: 0.75,
	}
	targets := ImprovementTargets{
		OverallConfidence: 0.90,
	}

	// Test data quality strategies
	dataStrategies := improver.createDataQualityStrategies(currentState, targets)
	assert.NotEmpty(t, dataStrategies)
	for _, strategy := range dataStrategies {
		assert.Equal(t, "data_quality", strategy.Category)
		assert.NotEmpty(t, strategy.Implementation)
		assert.Greater(t, strategy.ExpectedImprovement, 0.0)
	}

	// Test algorithm optimization strategies
	algoStrategies := improver.createAlgorithmOptimizationStrategies(currentState, targets)
	assert.NotEmpty(t, algoStrategies)
	for _, strategy := range algoStrategies {
		assert.Equal(t, "algorithm_optimization", strategy.Category)
		assert.Greater(t, strategy.ImpactScore, 0.0)
		assert.Greater(t, strategy.EffortScore, 0.0)
	}

	// Test training data strategies
	trainingStrategies := improver.createTrainingDataStrategies(currentState, targets)
	assert.NotEmpty(t, trainingStrategies)
	for _, strategy := range trainingStrategies {
		assert.Equal(t, "training_data", strategy.Category)
		assert.Contains(t, strategy.TargetFactors, "training_coverage")
	}

	// Test validation strategies
	validationStrategies := improver.createValidationStrategies(currentState, targets)
	assert.NotEmpty(t, validationStrategies)
	for _, strategy := range validationStrategies {
		assert.Equal(t, "validation", strategy.Category)
		assert.Contains(t, strategy.TargetFactors, "validation_accuracy")
	}

	// Test feature engineering strategies
	featureStrategies := improver.createFeatureEngineeringStrategies(currentState, targets)
	assert.NotEmpty(t, featureStrategies)
	for _, strategy := range featureStrategies {
		assert.Equal(t, "feature_engineering", strategy.Category)
		assert.Contains(t, strategy.TargetFactors, "feature_quality")
	}

	// Test ensemble strategies
	ensembleStrategies := improver.createEnsembleStrategies(currentState, targets)
	assert.NotEmpty(t, ensembleStrategies)
	for _, strategy := range ensembleStrategies {
		assert.Equal(t, "ensemble_methods", strategy.Category)
		assert.Contains(t, strategy.TargetFactors, "ensemble_accuracy")
	}

	// Test calibration strategies
	calibrationStrategies := improver.createCalibrationStrategies(currentState, targets)
	assert.NotEmpty(t, calibrationStrategies)
	for _, strategy := range calibrationStrategies {
		assert.Equal(t, "calibration", strategy.Category)
		assert.Contains(t, strategy.TargetFactors, "calibration_accuracy")
	}
}

func TestConfidenceImprover_NeedsAssessment(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	// Test data quality improvement need
	lowConfidenceState := &ConfidenceState{
		OverallConfidence: 0.70,
		LowConfidenceRate: 0.20,
	}
	assert.True(t, improver.needsDataQualityImprovement(lowConfidenceState))

	goodConfidenceState := &ConfidenceState{
		OverallConfidence: 0.90,
		LowConfidenceRate: 0.05,
	}
	assert.False(t, improver.needsDataQualityImprovement(goodConfidenceState))

	// Test algorithm optimization need
	decliningState := &ConfidenceState{
		TrendDirection: "declining",
	}
	assert.True(t, improver.needsAlgorithmOptimization(decliningState))

	stableState := &ConfidenceState{
		TrendDirection:    "stable",
		OverallConfidence: 0.85,
	}
	assert.False(t, improver.needsAlgorithmOptimization(stableState))

	// Test training data enhancement need
	trainingWeaknessState := &ConfidenceState{
		KeyWeaknesses: []string{"training_data", "other_issue"},
	}
	assert.True(t, improver.needsTrainingDataEnhancement(trainingWeaknessState))

	noTrainingWeaknessState := &ConfidenceState{
		OverallConfidence: 0.85, // High enough to not trigger the < 0.8 condition
		KeyWeaknesses:     []string{"other_issue"},
	}
	assert.False(t, improver.needsTrainingDataEnhancement(noTrainingWeaknessState))
}

func TestConfidenceImprover_ApplyConstraints(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	strategies := []ImprovementStrategy{
		{
			ID:          "strategy_1",
			Priority:    "high",
			ImpactScore: 0.9,
			EffortScore: 0.8, // Cost: 8000
			ROIScore:    1.125,
		},
		{
			ID:          "strategy_2",
			Priority:    "medium",
			ImpactScore: 0.7,
			EffortScore: 0.5, // Cost: 5000
			ROIScore:    1.4,
		},
		{
			ID:          "strategy_3",
			Priority:    "low",
			ImpactScore: 0.6,
			EffortScore: 0.3, // Cost: 3000
			ROIScore:    2.0,
		},
	}

	// Test with budget constraint
	config := &ImprovementConfig{
		Budget: 12000, // Should allow first two strategies (8000 + 5000 = 13000, close to budget)
	}

	constrainedStrategies := improver.applyConstraints(strategies, config)

	// Should include strategies within budget
	assert.LessOrEqual(t, len(constrainedStrategies), len(strategies))
	assert.NotEmpty(t, constrainedStrategies)

	// Test with no budget constraint
	configNoBudget := &ImprovementConfig{
		Budget: 0, // No constraint
	}

	unconstrainedStrategies := improver.applyConstraints(strategies, configNoBudget)
	assert.Equal(t, len(strategies), len(unconstrainedStrategies))
}

func TestConfidenceImprover_CreateImplementationPhases(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	strategies := []ImprovementStrategy{
		{ID: "strategy_1", Priority: "high"},
		{ID: "strategy_2", Priority: "high"},
		{ID: "strategy_3", Priority: "medium"},
		{ID: "strategy_4", Priority: "low"},
	}

	config := &ImprovementConfig{
		Timeline: "12 months",
	}

	phases := improver.createImplementationPhases(strategies, config)

	assert.Len(t, phases, 3) // Foundation, Optimization, Refinement

	// Test phase structure
	for _, phase := range phases {
		assert.NotEmpty(t, phase.ID)
		assert.NotEmpty(t, phase.Name)
		assert.NotEmpty(t, phase.Description)
		assert.NotEmpty(t, phase.Duration)
		assert.NotEmpty(t, phase.Strategies)
		assert.False(t, phase.StartDate.IsZero())
		assert.False(t, phase.EndDate.IsZero())
		assert.True(t, phase.EndDate.After(phase.StartDate))
	}

	// Test that phases are in chronological order
	for i := 1; i < len(phases); i++ {
		assert.True(t, phases[i].StartDate.After(phases[i-1].StartDate) ||
			phases[i].StartDate.Equal(phases[i-1].EndDate),
			"Phases should be in chronological order")
	}
}

func TestConfidenceImprover_CreatePlanTimeline(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	phases := []ImprovementPhase{
		{
			ID:        "phase_1",
			StartDate: time.Now(),
			EndDate:   time.Now().AddDate(0, 3, 0),
		},
		{
			ID:        "phase_2",
			StartDate: time.Now().AddDate(0, 3, 0),
			EndDate:   time.Now().AddDate(0, 6, 0),
		},
	}

	config := &ImprovementConfig{
		Timeline: "12 months",
	}

	timeline := improver.createPlanTimeline(phases, config)

	assert.NotEmpty(t, timeline.TotalDuration)
	assert.False(t, timeline.StartDate.IsZero())
	assert.False(t, timeline.EndDate.IsZero())
	assert.True(t, timeline.EndDate.After(timeline.StartDate))
	assert.NotEmpty(t, timeline.KeyMilestones)
	assert.NotEmpty(t, timeline.CriticalPath)
	assert.NotEmpty(t, timeline.PhaseSchedule)
	assert.NotEmpty(t, timeline.BufferTime)

	// Test milestones
	for _, milestone := range timeline.KeyMilestones {
		assert.NotEmpty(t, milestone.ID)
		assert.NotEmpty(t, milestone.Name)
		assert.NotEmpty(t, milestone.Description)
		assert.False(t, milestone.TargetDate.IsZero())
	}
}

func TestConfidenceImprover_EstimateBudget(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	strategies := []ImprovementStrategy{
		{
			ID:          "strategy_1",
			Category:    "data_quality",
			EffortScore: 0.8,
		},
		{
			ID:          "strategy_2",
			Category:    "algorithm_optimization",
			EffortScore: 0.6,
		},
	}

	phases := []ImprovementPhase{
		{ID: "phase_1"},
		{ID: "phase_2"},
	}

	config := &ImprovementConfig{}

	budget := improver.estimateBudget(strategies, phases, config)

	assert.Greater(t, budget.TotalCost, 0.0)
	assert.NotEmpty(t, budget.PhaseBreakdown)
	assert.NotEmpty(t, budget.CategoryBreakdown)
	assert.NotEmpty(t, budget.ResourceCosts)
	assert.Greater(t, budget.ContingencyFund, 0.0)
	assert.Greater(t, budget.ROIProjection, 0.0)
	assert.NotEmpty(t, budget.PaybackPeriod)

	// Test that breakdown sums match total
	var phaseSum float64
	for _, cost := range budget.PhaseBreakdown {
		phaseSum += cost
	}
	assert.InDelta(t, budget.TotalCost, phaseSum, 0.01)

	var categorySum float64
	for _, cost := range budget.CategoryBreakdown {
		categorySum += cost
	}
	assert.InDelta(t, budget.TotalCost, categorySum, 0.01)
}

func TestConfidenceImprover_AssessRisks(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	strategies := []ImprovementStrategy{
		{ID: "strategy_1", Category: "algorithm_optimization"},
	}
	phases := []ImprovementPhase{
		{ID: "phase_1"},
	}

	riskAssessment := improver.assessRisks(strategies, phases)

	assert.NotEmpty(t, riskAssessment.OverallRiskLevel)
	assert.NotEmpty(t, riskAssessment.IdentifiedRisks)
	assert.NotEmpty(t, riskAssessment.MitigationPlan)
	assert.NotEmpty(t, riskAssessment.ContingencyPlan)
	assert.NotEmpty(t, riskAssessment.RiskMonitoring)

	// Test risk structure
	for _, risk := range riskAssessment.IdentifiedRisks {
		assert.NotEmpty(t, risk.ID)
		assert.NotEmpty(t, risk.Name)
		assert.NotEmpty(t, risk.Description)
		assert.NotEmpty(t, risk.Category)
		assert.GreaterOrEqual(t, risk.Probability, 0.0)
		assert.LessOrEqual(t, risk.Probability, 1.0)
		assert.GreaterOrEqual(t, risk.Impact, 0.0)
		assert.LessOrEqual(t, risk.Impact, 1.0)
		assert.GreaterOrEqual(t, risk.RiskScore, 0.0)
		assert.LessOrEqual(t, risk.RiskScore, 1.0)
		assert.InDelta(t, risk.RiskScore, risk.Probability*risk.Impact, 0.01)
	}

	// Test mitigation structure
	for _, mitigation := range riskAssessment.MitigationPlan {
		assert.NotEmpty(t, mitigation.RiskID)
		assert.NotEmpty(t, mitigation.Strategy)
		assert.NotEmpty(t, mitigation.Actions)
		assert.NotEmpty(t, mitigation.Owner)
		assert.NotEmpty(t, mitigation.Timeline)
		assert.GreaterOrEqual(t, mitigation.Cost, 0.0)
		assert.GreaterOrEqual(t, mitigation.Effectiveness, 0.0)
		assert.LessOrEqual(t, mitigation.Effectiveness, 1.0)
	}
}

func TestConfidenceImprover_CalculateExpectedOutcomes(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	currentState := &ConfidenceState{
		OverallConfidence: 0.75,
	}
	targets := ImprovementTargets{
		OverallConfidence: 0.90,
	}
	strategies := []ImprovementStrategy{
		{
			ImpactScore:         0.8,
			ExpectedImprovement: 15.0,
		},
		{
			ImpactScore:         0.6,
			ExpectedImprovement: 10.0,
		},
	}

	outcomes := improver.calculateExpectedOutcomes(currentState, targets, strategies)

	assert.Greater(t, outcomes.ConfidenceImprovement, 0.0)
	assert.Greater(t, outcomes.QualityImprovement, 0.0)
	assert.Greater(t, outcomes.EfficiencyGains, 0.0)
	assert.Greater(t, outcomes.CostSavings, 0.0)
	assert.NotEmpty(t, outcomes.TimeToValue)
	assert.NotEmpty(t, outcomes.LongTermBenefits)
	assert.NotEmpty(t, outcomes.KPIImprovements)

	// Test that quality and efficiency are fractions of confidence improvement
	assert.Less(t, outcomes.QualityImprovement, outcomes.ConfidenceImprovement)
	assert.Less(t, outcomes.EfficiencyGains, outcomes.ConfidenceImprovement)

	// Test KPI improvements
	for kpi, improvement := range outcomes.KPIImprovements {
		assert.NotEmpty(t, kpi)
		assert.Greater(t, improvement, 0.0)
	}
}

func TestConfidenceImprover_CreateMonitoringPlan(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	targets := ImprovementTargets{
		OverallConfidence: 0.90,
	}
	strategies := []ImprovementStrategy{
		{ID: "strategy_1"},
	}

	monitoringPlan := improver.createMonitoringPlan(targets, strategies)

	assert.NotEmpty(t, monitoringPlan.Frequency)
	assert.NotEmpty(t, monitoringPlan.KeyMetrics)
	assert.NotEmpty(t, monitoringPlan.ReportingSchedule)
	assert.NotEmpty(t, monitoringPlan.AlertThresholds)
	assert.NotEmpty(t, monitoringPlan.ReviewPoints)
	assert.NotEmpty(t, monitoringPlan.Stakeholders)
	assert.NotEmpty(t, monitoringPlan.Tools)
	assert.NotEmpty(t, monitoringPlan.Dashboards)

	// Test alert thresholds
	for metric, threshold := range monitoringPlan.AlertThresholds {
		assert.NotEmpty(t, metric)
		assert.Greater(t, threshold, 0.0)
		assert.LessOrEqual(t, threshold, 1.0)
	}

	// Test review points are in order
	for i := 1; i < len(monitoringPlan.ReviewPoints); i++ {
		assert.True(t, monitoringPlan.ReviewPoints[i].After(monitoringPlan.ReviewPoints[i-1]),
			"Review points should be in chronological order")
	}
}

func TestConfidenceImprover_GeneratePlanRecommendations(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	currentState := &ConfidenceState{
		OverallConfidence: 0.70, // Low confidence to trigger recommendations
	}
	strategies := []ImprovementStrategy{
		{ID: "strategy_1", ROIScore: 1.5},
	}
	phases := []ImprovementPhase{
		{ID: "phase_1"},
	}

	recommendations := improver.generatePlanRecommendations(currentState, strategies, phases)

	assert.NotEmpty(t, recommendations)

	// Test recommendation structure
	for _, rec := range recommendations {
		assert.NotEmpty(t, rec.ID)
		assert.NotEmpty(t, rec.Type)
		assert.NotEmpty(t, rec.Priority)
		assert.NotEmpty(t, rec.Title)
		assert.NotEmpty(t, rec.Description)
		assert.NotEmpty(t, rec.Rationale)
		assert.NotEmpty(t, rec.Actions)
		assert.NotEmpty(t, rec.ExpectedBenefit)
		assert.NotEmpty(t, rec.Timeline)
		assert.Contains(t, []string{"strategic", "tactical", "operational"}, rec.Type)
		assert.Contains(t, []string{"critical", "high", "medium", "low"}, rec.Priority)
	}

	// Should include strategic recommendation for low confidence
	hasStrategicRec := false
	for _, rec := range recommendations {
		if rec.Type == "strategic" {
			hasStrategicRec = true
			break
		}
	}
	assert.True(t, hasStrategicRec, "Should include strategic recommendation for low confidence")
}

func TestConfidenceImprover_Integration(t *testing.T) {
	improver, db := setupTestConfidenceImprover(t)
	defer db.Close()

	ctx := context.Background()

	// Test complete flow with realistic configuration
	config := &ImprovementConfig{
		TargetConfidence: 0.88,
		Timeline:         "9 months",
		Budget:           75000,
		Priorities:       []string{"data_quality", "algorithm_optimization", "validation"},
		Constraints:      []string{"limited_engineering_time"},
		FocusAreas:       []string{"reduce_low_confidence", "improve_accuracy"},
		RiskTolerance:    "medium",
		ResourceAvailability: map[string]float64{
			"data_team":        0.8,
			"engineering_team": 0.6,
			"ml_team":          0.4,
		},
		SuccessMetrics: []string{
			"overall_confidence",
			"high_confidence_rate",
			"low_confidence_rate",
			"classification_accuracy",
		},
	}

	plan, err := improver.GenerateImprovementPlan(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, plan)

	// Verify plan consistency
	assert.Equal(t, config.TargetConfidence, plan.TargetMetrics.OverallConfidence)
	assert.Greater(t, plan.TargetMetrics.OverallConfidence, plan.CurrentState.OverallConfidence)
	assert.Greater(t, plan.TargetMetrics.HighConfidenceRate, plan.CurrentState.HighConfidenceRate)
	assert.Less(t, plan.TargetMetrics.LowConfidenceRate, plan.CurrentState.LowConfidenceRate)

	// Verify budget constraint is respected (with some tolerance)
	assert.LessOrEqual(t, plan.BudgetEstimate.TotalCost, config.Budget*1.3)

	// Verify strategies address identified priorities
	priorityStrategies := make(map[string]bool)
	for _, strategy := range plan.Strategies {
		priorityStrategies[strategy.Category] = true
	}

	// Should have strategies for at least some priorities
	foundPriorities := 0
	for _, priority := range config.Priorities {
		if priorityStrategies[priority] {
			foundPriorities++
		}
	}
	assert.Greater(t, foundPriorities, 0, "Should have strategies addressing configured priorities")

	// Verify timeline consistency
	assert.NotEmpty(t, plan.Timeline.TotalDuration)
	assert.True(t, plan.Timeline.EndDate.After(plan.Timeline.StartDate))

	// Verify phases have realistic timelines
	totalDuration := plan.Timeline.EndDate.Sub(plan.Timeline.StartDate)
	assert.Greater(t, totalDuration.Hours(), float64(24*30), "Plan should span at least 30 days") // At least 30 days
	assert.Less(t, totalDuration.Hours(), float64(24*365*2), "Plan should not exceed 2 years")    // Less than 2 years

	// Verify monitoring plan includes success metrics
	for _, metric := range config.SuccessMetrics {
		assert.Contains(t, plan.MonitoringPlan.KeyMetrics, metric,
			"Monitoring plan should include configured success metric: %s", metric)
	}

	// Verify expected outcomes are positive
	assert.Greater(t, plan.ExpectedOutcomes.ConfidenceImprovement, float64(0))
	assert.Greater(t, plan.ExpectedOutcomes.QualityImprovement, float64(0))
	assert.NotEmpty(t, plan.ExpectedOutcomes.LongTermBenefits)

	// Verify risk assessment includes realistic risks
	assert.NotEmpty(t, plan.RiskAssessment.IdentifiedRisks)
	assert.NotEmpty(t, plan.RiskAssessment.MitigationPlan)
	assert.Contains(t, []string{"low", "medium", "high"}, plan.RiskAssessment.OverallRiskLevel)
}
