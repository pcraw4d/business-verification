package industry_codes

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewVotingValidator(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Test with nil config
	validator := NewVotingValidator(nil, logger)
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.config)
	assert.Equal(t, 1, validator.config.MinResultCount)
	assert.Equal(t, 10, validator.config.MaxResultCount)
	assert.Equal(t, 0.3, validator.config.MinVotingScoreThreshold)
	assert.True(t, validator.config.EnableStatisticalValidation)
	assert.True(t, validator.config.EnableTemporalValidation)
	assert.Equal(t, 5*time.Minute, validator.config.TemporalWindow)

	// Test with custom config
	customConfig := &VotingValidationConfig{
		MinResultCount:              5,
		MaxResultCount:              15,
		MinVotingScoreThreshold:     0.5,
		EnableAnomalyDetection:      false,
		EnableCrossValidation:       false,
		EnableStatisticalValidation: false,
		EnableTemporalValidation:    false,
	}

	validator = NewVotingValidator(customConfig, logger)
	assert.NotNil(t, validator)
	assert.Equal(t, 5, validator.config.MinResultCount)
	assert.Equal(t, 15, validator.config.MaxResultCount)
	assert.Equal(t, 0.5, validator.config.MinVotingScoreThreshold)
	assert.False(t, validator.config.EnableAnomalyDetection)
	assert.False(t, validator.config.EnableCrossValidation)
	assert.False(t, validator.config.EnableStatisticalValidation)
	assert.False(t, validator.config.EnableTemporalValidation)
}

func TestVotingValidator_ValidateVotingResult(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewVotingValidator(nil, logger)

	// Create test voting result
	votingResult := &VotingResult{
		FinalResults: []*ClassificationResult{
			{
				Code: &IndustryCode{
					Type:        CodeTypeSIC,
					Code:        "1234",
					Description: "Test Industry",
				},
				Confidence: 0.85,
				Reasons:    []string{"Test reasoning"},
			},
		},
		VotingScore:    0.8,
		Agreement:      0.75,
		Consistency:    0.7,
		Diversity:      0.6,
		VotingStrategy: VotingStrategyWeightedAverage,
	}

	// Create test votes
	votes := []*StrategyVote{
		{
			StrategyName: "test_strategy_1",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.85,
				},
			},
			Weight:     0.5,
			Confidence: 0.85,
			VoteTime:   time.Now(),
		},
		{
			StrategyName: "test_strategy_2",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.75,
				},
			},
			Weight:     0.5,
			Confidence: 0.75,
			VoteTime:   time.Now(),
		},
	}

	// Test validation
	ctx := context.Background()
	validationResult, err := validator.ValidateVotingResult(ctx, votingResult, votes)

	require.NoError(t, err)
	assert.NotNil(t, validationResult)
	assert.True(t, validationResult.IsValid)
	assert.Greater(t, validationResult.ValidationScore, 0.0)
	assert.NotNil(t, validationResult.QualityMetrics)
	assert.NotNil(t, validationResult.ConsistencyChecks)
	// Recommendations may be empty if validation passes with no issues
	// This is expected behavior for good quality results
}

func TestVotingValidator_ValidateVotingResult_InvalidInputs(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewVotingValidator(nil, logger)

	ctx := context.Background()

	// Test with nil result
	validationResult, err := validator.ValidateVotingResult(ctx, nil, []*StrategyVote{})
	require.NoError(t, err)
	assert.False(t, validationResult.IsValid)
	assert.Equal(t, 0.0, validationResult.ValidationScore)
	assert.Len(t, validationResult.Issues, 1)
	assert.Equal(t, "input_validation", validationResult.Issues[0].Type)
	assert.Equal(t, "critical", validationResult.Issues[0].Severity)

	// Test with nil votes
	votingResult := &VotingResult{
		FinalResults: []*ClassificationResult{},
	}
	validationResult, err = validator.ValidateVotingResult(ctx, votingResult, nil)
	require.NoError(t, err)
	assert.False(t, validationResult.IsValid)
	assert.Equal(t, 0.0, validationResult.ValidationScore)
	assert.Len(t, validationResult.Issues, 1)
	assert.Equal(t, "input_validation", validationResult.Issues[0].Type)

	// Test with empty votes
	validationResult, err = validator.ValidateVotingResult(ctx, votingResult, []*StrategyVote{})
	require.NoError(t, err)
	assert.False(t, validationResult.IsValid)
	assert.Equal(t, 0.0, validationResult.ValidationScore)
	assert.Len(t, validationResult.Issues, 1)
	assert.Equal(t, "input_validation", validationResult.Issues[0].Type)
}

func TestVotingValidator_ValidateVotingResult_LowScores(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &VotingValidationConfig{
		MinVotingScoreThreshold: 0.8,
		MinAgreementThreshold:   0.8,
		MinConsistencyThreshold: 0.8,
	}
	validator := NewVotingValidator(config, logger)

	votingResult := &VotingResult{
		FinalResults: []*ClassificationResult{
			{
				Code: &IndustryCode{
					Type:        CodeTypeSIC,
					Code:        "1234",
					Description: "Test Industry",
				},
				Confidence: 0.85,
			},
		},
		VotingScore: 0.3, // Below threshold
		Agreement:   0.3, // Below threshold
		Consistency: 0.3, // Below threshold
	}

	votes := []*StrategyVote{
		{
			StrategyName: "test_strategy_1",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.85,
				},
			},
			Weight:     0.5,
			Confidence: 0.85,
			VoteTime:   time.Now(),
		},
	}

	ctx := context.Background()
	validationResult, err := validator.ValidateVotingResult(ctx, votingResult, votes)

	require.NoError(t, err)
	assert.False(t, validationResult.IsValid)
	assert.Len(t, validationResult.Issues, 3) // voting_score, agreement, consistency
	assert.Less(t, validationResult.ValidationScore, 0.8)
}

func TestVotingValidator_QualityMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewVotingValidator(nil, logger)

	votingResult := &VotingResult{
		FinalResults: []*ClassificationResult{
			{
				Code: &IndustryCode{
					Type:        CodeTypeSIC,
					Code:        "1234",
					Description: "Test Industry",
				},
				Confidence: 0.85,
			},
			{
				Code: &IndustryCode{
					Type:        CodeTypeNAICS,
					Code:        "541511",
					Description: "Custom Computer Programming Services",
				},
				Confidence: 0.75,
			},
		},
		VotingScore: 0.8,
		Agreement:   0.75,
		Consistency: 0.7,
	}

	votes := []*StrategyVote{
		{
			StrategyName: "strategy_1",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.85,
				},
				{
					Code: &IndustryCode{
						Type:        CodeTypeNAICS,
						Code:        "541511",
						Description: "Custom Computer Programming Services",
					},
					Confidence: 0.75,
				},
			},
			Weight:     0.5,
			Confidence: 0.85,
			VoteTime:   time.Now(),
		},
		{
			StrategyName: "strategy_2",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.80,
				},
				{
					Code: &IndustryCode{
						Type:        CodeTypeNAICS,
						Code:        "541511",
						Description: "Custom Computer Programming Services",
					},
					Confidence: 0.70,
				},
			},
			Weight:     0.5,
			Confidence: 0.75,
			VoteTime:   time.Now(),
		},
	}

	ctx := context.Background()
	validationResult, err := validator.ValidateVotingResult(ctx, votingResult, votes)

	require.NoError(t, err)
	assert.NotNil(t, validationResult.QualityMetrics)

	metrics := validationResult.QualityMetrics
	assert.Greater(t, metrics.ResultCompleteness, 0.0)
	assert.Greater(t, metrics.ConfidenceReliability, 0.0)
	assert.Greater(t, metrics.StrategyConsistency, 0.0)
	assert.Greater(t, metrics.CodeFormatCompliance, 0.0)
	assert.Greater(t, metrics.OverallQuality, 0.0)
	assert.LessOrEqual(t, metrics.OverallQuality, 1.0)
}

func TestVotingValidator_ConsistencyChecks(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewVotingValidator(nil, logger)

	votingResult := &VotingResult{
		FinalResults: []*ClassificationResult{
			{
				Code: &IndustryCode{
					Type:        CodeTypeSIC,
					Code:        "1234",
					Description: "Test Industry",
				},
				Confidence: 0.85,
			},
		},
		VotingScore: 0.8,
		Agreement:   0.75,
		Consistency: 0.7,
	}

	votes := []*StrategyVote{
		{
			StrategyName: "strategy_1",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.85,
				},
			},
			Weight:     0.5,
			Confidence: 0.85,
			VoteTime:   time.Now(),
		},
		{
			StrategyName: "strategy_2",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.80,
				},
			},
			Weight:     0.5,
			Confidence: 0.80,
			VoteTime:   time.Now(),
		},
	}

	ctx := context.Background()
	validationResult, err := validator.ValidateVotingResult(ctx, votingResult, votes)

	require.NoError(t, err)
	assert.NotNil(t, validationResult.ConsistencyChecks)

	consistency := validationResult.ConsistencyChecks
	assert.Greater(t, consistency.CrossStrategyAgreement, 0.0)
	assert.Greater(t, consistency.ConfidenceConsistency, 0.0)
	assert.Greater(t, consistency.ResultStability, 0.0)
	assert.GreaterOrEqual(t, consistency.AnomalyScore, 0.0)
}

func TestVotingValidator_AnomalyDetection(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &VotingValidationConfig{
		EnableAnomalyDetection: true,
		AnomalyThreshold:       1.0, // Lower threshold to detect the anomaly
	}
	validator := NewVotingValidator(config, logger)

	votingResult := &VotingResult{
		FinalResults: []*ClassificationResult{
			{
				Code: &IndustryCode{
					Type:        CodeTypeSIC,
					Code:        "1234",
					Description: "Test Industry",
				},
				Confidence: 0.85,
			},
		},
		VotingScore: 0.8,
		Agreement:   0.75,
		Consistency: 0.7,
	}

	// Create votes with one anomalous confidence
	votes := []*StrategyVote{
		{
			StrategyName: "strategy_1",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.85,
				},
			},
			Weight:     0.5,
			Confidence: 0.85,
			VoteTime:   time.Now(),
		},
		{
			StrategyName: "strategy_2",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.80,
				},
			},
			Weight:     0.5,
			Confidence: 0.80,
			VoteTime:   time.Now(),
		},
		{
			StrategyName: "anomalous_strategy",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.85,
				},
			},
			Weight:     0.5,
			Confidence: 0.99, // Anomalously high confidence
			VoteTime:   time.Now(),
		},
	}

	ctx := context.Background()
	validationResult, err := validator.ValidateVotingResult(ctx, votingResult, votes)

	require.NoError(t, err)
	assert.NotEmpty(t, validationResult.Warnings)

	// Debug: Print all warnings (for development only)
	// t.Logf("Generated warnings: %+v", validationResult.Warnings)
	// for i, warning := range validationResult.Warnings {
	// 	t.Logf("Warning %d: Type=%s, Message=%s", i, warning.Type, warning.Message)
	// }

	// Check for anomaly warnings
	hasAnomalyWarning := false
	for _, warning := range validationResult.Warnings {
		if warning.Type == "anomaly" {
			hasAnomalyWarning = true
			break
		}
	}
	assert.True(t, hasAnomalyWarning)
}

func TestVotingValidator_CrossValidation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &VotingValidationConfig{
		EnableCrossValidation:    true,
		CrossValidationThreshold: 0.5,
	}
	validator := NewVotingValidator(config, logger)

	votingResult := &VotingResult{
		FinalResults: []*ClassificationResult{
			{
				Code: &IndustryCode{
					Type:        CodeTypeSIC,
					Code:        "1234",
					Description: "Test Industry",
				},
				Confidence: 0.85,
			},
		},
		VotingScore: 0.8,
		Agreement:   0.75,
		Consistency: 0.7,
	}

	votes := []*StrategyVote{
		{
			StrategyName: "strategy_1",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.85,
				},
			},
			Weight:     0.5,
			Confidence: 0.85,
			VoteTime:   time.Now(),
		},
		{
			StrategyName: "strategy_2",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.80,
				},
			},
			Weight:     0.5,
			Confidence: 0.80,
			VoteTime:   time.Now(),
		},
	}

	ctx := context.Background()
	validationResult, err := validator.ValidateVotingResult(ctx, votingResult, votes)

	require.NoError(t, err)
	// Cross-validation should pass with consistent results
	assert.True(t, validationResult.IsValid)
}

func TestVotingValidator_StatisticalValidation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &VotingValidationConfig{
		EnableStatisticalValidation: true,
		StatisticalSignificance:     0.1,
	}
	validator := NewVotingValidator(config, logger)

	votingResult := &VotingResult{
		FinalResults: []*ClassificationResult{
			{
				Code: &IndustryCode{
					Type:        CodeTypeSIC,
					Code:        "1234",
					Description: "Test Industry",
				},
				Confidence: 0.85,
			},
		},
		VotingScore: 0.8,
		Agreement:   0.75,
		Consistency: 0.7,
	}

	votes := []*StrategyVote{
		{
			StrategyName: "strategy_1",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.85,
				},
			},
			Weight:     0.5,
			Confidence: 0.85,
			VoteTime:   time.Now(),
		},
		{
			StrategyName: "strategy_2",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.80,
				},
			},
			Weight:     0.5,
			Confidence: 0.80,
			VoteTime:   time.Now(),
		},
	}

	ctx := context.Background()
	validationResult, err := validator.ValidateVotingResult(ctx, votingResult, votes)

	require.NoError(t, err)
	// Statistical validation should not cause issues with normal data
	assert.True(t, validationResult.IsValid)
}

func TestVotingValidator_TemporalValidation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &VotingValidationConfig{
		EnableTemporalValidation: true,
		TemporalWindow:           1 * time.Minute,
	}
	validator := NewVotingValidator(config, logger)

	votingResult := &VotingResult{
		FinalResults: []*ClassificationResult{
			{
				Code: &IndustryCode{
					Type:        CodeTypeSIC,
					Code:        "1234",
					Description: "Test Industry",
				},
				Confidence: 0.85,
			},
		},
		VotingScore: 0.8,
		Agreement:   0.75,
		Consistency: 0.7,
	}

	now := time.Now()
	votes := []*StrategyVote{
		{
			StrategyName: "strategy_1",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.85,
				},
			},
			Weight:     0.5,
			Confidence: 0.85,
			VoteTime:   now,
		},
		{
			StrategyName: "strategy_2",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.80,
				},
			},
			Weight:     0.5,
			Confidence: 0.80,
			VoteTime:   now.Add(30 * time.Second), // Within temporal window
		},
	}

	ctx := context.Background()
	validationResult, err := validator.ValidateVotingResult(ctx, votingResult, votes)

	require.NoError(t, err)
	// Temporal validation should pass with votes within window
	assert.True(t, validationResult.IsValid)

	// Test with votes outside temporal window
	votes[1].VoteTime = now.Add(2 * time.Minute) // Outside temporal window

	validationResult, err = validator.ValidateVotingResult(ctx, votingResult, votes)
	require.NoError(t, err)

	// Should have temporal warning
	hasTemporalWarning := false
	for _, warning := range validationResult.Warnings {
		if warning.Type == "temporal" {
			hasTemporalWarning = true
			break
		}
	}
	assert.True(t, hasTemporalWarning)
}

func TestVotingValidator_Recommendations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewVotingValidator(nil, logger)

	votingResult := &VotingResult{
		FinalResults: []*ClassificationResult{
			{
				Code: &IndustryCode{
					Type:        CodeTypeSIC,
					Code:        "1234",
					Description: "Test Industry",
				},
				Confidence: 0.85,
			},
		},
		VotingScore: 0.8,
		Agreement:   0.75,
		Consistency: 0.7,
	}

	votes := []*StrategyVote{
		{
			StrategyName: "strategy_1",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.85,
				},
			},
			Weight:     0.5,
			Confidence: 0.85,
			VoteTime:   time.Now(),
		},
		{
			StrategyName: "strategy_2",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.80,
				},
			},
			Weight:     0.5,
			Confidence: 0.80,
			VoteTime:   time.Now(),
		},
	}

	ctx := context.Background()
	validationResult, err := validator.ValidateVotingResult(ctx, votingResult, votes)

	require.NoError(t, err)
	// Recommendations may be empty if validation passes with no issues
	// This is expected behavior for good quality results

	// Check that recommendations are unique if any exist
	if len(validationResult.Recommendations) > 0 {
		seen := make(map[string]bool)
		for _, rec := range validationResult.Recommendations {
			assert.False(t, seen[rec], "Duplicate recommendation: %s", rec)
			seen[rec] = true
		}
	}
}

func TestVotingValidator_ValidationScore(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewVotingValidator(nil, logger)

	votingResult := &VotingResult{
		FinalResults: []*ClassificationResult{
			{
				Code: &IndustryCode{
					Type:        CodeTypeSIC,
					Code:        "1234",
					Description: "Test Industry",
				},
				Confidence: 0.85,
			},
		},
		VotingScore: 0.8,
		Agreement:   0.75,
		Consistency: 0.7,
	}

	votes := []*StrategyVote{
		{
			StrategyName: "strategy_1",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.85,
				},
			},
			Weight:     0.5,
			Confidence: 0.85,
			VoteTime:   time.Now(),
		},
		{
			StrategyName: "strategy_2",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.80,
				},
			},
			Weight:     0.5,
			Confidence: 0.80,
			VoteTime:   time.Now(),
		},
	}

	ctx := context.Background()
	validationResult, err := validator.ValidateVotingResult(ctx, votingResult, votes)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, validationResult.ValidationScore, 0.0)
	assert.LessOrEqual(t, validationResult.ValidationScore, 1.0)
}

func TestVotingValidator_EmptyResults(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewVotingValidator(nil, logger)

	votingResult := &VotingResult{
		FinalResults: []*ClassificationResult{}, // Empty results
		VotingScore:  0.8,
		Agreement:    0.75,
		Consistency:  0.7,
	}

	votes := []*StrategyVote{
		{
			StrategyName: "strategy_1",
			Results:      []*ClassificationResult{},
			Weight:       0.5,
			Confidence:   0.85,
			VoteTime:     time.Now(),
		},
	}

	ctx := context.Background()
	validationResult, err := validator.ValidateVotingResult(ctx, votingResult, votes)

	require.NoError(t, err)
	assert.False(t, validationResult.IsValid)
	assert.Len(t, validationResult.Issues, 1)
	assert.Equal(t, "result_count", validationResult.Issues[0].Type)
}

func TestVotingValidator_NilVotes(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewVotingValidator(nil, logger)

	votingResult := &VotingResult{
		FinalResults: []*ClassificationResult{
			{
				Code: &IndustryCode{
					Type:        CodeTypeSIC,
					Code:        "1234",
					Description: "Test Industry",
				},
				Confidence: 0.85,
			},
		},
		VotingScore: 0.8,
		Agreement:   0.75,
		Consistency: 0.7,
	}

	ctx := context.Background()
	validationResult, err := validator.ValidateVotingResult(ctx, votingResult, nil)

	require.NoError(t, err)
	assert.False(t, validationResult.IsValid)
	assert.Equal(t, 0.0, validationResult.ValidationScore)
	assert.Len(t, validationResult.Issues, 1)
	assert.Equal(t, "input_validation", validationResult.Issues[0].Type)
}

func TestVotingValidator_MultipleCodeTypes(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewVotingValidator(nil, logger)

	votingResult := &VotingResult{
		FinalResults: []*ClassificationResult{
			{
				Code: &IndustryCode{
					Type:        CodeTypeSIC,
					Code:        "1234",
					Description: "Test Industry",
				},
				Confidence: 0.85,
			},
			{
				Code: &IndustryCode{
					Type:        CodeTypeNAICS,
					Code:        "541511",
					Description: "Custom Computer Programming Services",
				},
				Confidence: 0.75,
			},
			{
				Code: &IndustryCode{
					Type:        CodeTypeMCC,
					Code:        "5734",
					Description: "Computer Software Stores",
				},
				Confidence: 0.65,
			},
		},
		VotingScore: 0.8,
		Agreement:   0.75,
		Consistency: 0.7,
	}

	votes := []*StrategyVote{
		{
			StrategyName: "strategy_1",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.85,
				},
				{
					Code: &IndustryCode{
						Type:        CodeTypeNAICS,
						Code:        "541511",
						Description: "Custom Computer Programming Services",
					},
					Confidence: 0.75,
				},
				{
					Code: &IndustryCode{
						Type:        CodeTypeMCC,
						Code:        "5734",
						Description: "Computer Software Stores",
					},
					Confidence: 0.65,
				},
			},
			Weight:     0.5,
			Confidence: 0.85,
			VoteTime:   time.Now(),
		},
		{
			StrategyName: "strategy_2",
			Results: []*ClassificationResult{
				{
					Code: &IndustryCode{
						Type:        CodeTypeSIC,
						Code:        "1234",
						Description: "Test Industry",
					},
					Confidence: 0.80,
				},
				{
					Code: &IndustryCode{
						Type:        CodeTypeNAICS,
						Code:        "541511",
						Description: "Custom Computer Programming Services",
					},
					Confidence: 0.70,
				},
				{
					Code: &IndustryCode{
						Type:        CodeTypeMCC,
						Code:        "5734",
						Description: "Computer Software Stores",
					},
					Confidence: 0.60,
				},
			},
			Weight:     0.5,
			Confidence: 0.80,
			VoteTime:   time.Now(),
		},
	}

	ctx := context.Background()
	validationResult, err := validator.ValidateVotingResult(ctx, votingResult, votes)

	require.NoError(t, err)
	assert.True(t, validationResult.IsValid)
	assert.NotNil(t, validationResult.QualityMetrics)
	assert.NotNil(t, validationResult.ConsistencyChecks)
	assert.Greater(t, validationResult.ValidationScore, 0.0)
}
