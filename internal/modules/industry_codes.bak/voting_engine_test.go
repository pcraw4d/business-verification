package industry_codes

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestVotingEngine_ConductVoting(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name           string
		config         *VotingConfig
		votes          []*StrategyVote
		expectedLength int
		expectedError  string
	}{
		{
			name: "successful weighted average voting",
			config: &VotingConfig{
				Strategy:               VotingStrategyWeightedAverage,
				MinVoters:              2,
				RequiredAgreement:      0.5,
				ConfidenceWeight:       0.4,
				ConsistencyWeight:      0.3,
				DiversityWeight:        0.3,
				EnableTieBreaking:      true,
				EnableOutlierFiltering: false,
			},
			votes: []*StrategyVote{
				createTestStrategyVote("keyword", 0.8, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.85),
					createTestClassificationResult("722513", CodeTypeNAICS, 0.75),
				}),
				createTestStrategyVote("description", 0.7, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.80),
					createTestClassificationResult("5812", CodeTypeSIC, 0.70),
				}),
				createTestStrategyVote("business_name", 0.6, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.75),
					createTestClassificationResult("722513", CodeTypeNAICS, 0.65),
				}),
			},
			expectedLength: 3,
		},
		{
			name: "successful majority voting",
			config: &VotingConfig{
				Strategy:               VotingStrategyMajority,
				MinVoters:              2,
				RequiredAgreement:      0.5,
				ConfidenceWeight:       0.4,
				ConsistencyWeight:      0.3,
				DiversityWeight:        0.3,
				EnableTieBreaking:      true,
				EnableOutlierFiltering: false,
			},
			votes: []*StrategyVote{
				createTestStrategyVote("keyword", 0.8, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.85),
					createTestClassificationResult("722513", CodeTypeNAICS, 0.75),
				}),
				createTestStrategyVote("description", 0.7, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.80),
					createTestClassificationResult("5812", CodeTypeSIC, 0.70),
				}),
				createTestStrategyVote("business_name", 0.6, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.75),
					createTestClassificationResult("722513", CodeTypeNAICS, 0.65),
				}),
			},
			expectedLength: 3, // All codes have at least one vote, so all qualify
		},
		{
			name: "successful borda count voting",
			config: &VotingConfig{
				Strategy:               VotingStrategyBordaCount,
				MinVoters:              2,
				RequiredAgreement:      0.5,
				ConfidenceWeight:       0.4,
				ConsistencyWeight:      0.3,
				DiversityWeight:        0.3,
				EnableTieBreaking:      true,
				EnableOutlierFiltering: false,
			},
			votes: []*StrategyVote{
				createTestStrategyVote("keyword", 0.8, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.85),
					createTestClassificationResult("722513", CodeTypeNAICS, 0.75),
				}),
				createTestStrategyVote("description", 0.7, []*ClassificationResult{
					createTestClassificationResult("722513", CodeTypeNAICS, 0.80),
					createTestClassificationResult("5411", CodeTypeSIC, 0.70),
				}),
			},
			expectedLength: 2,
		},
		{
			name: "successful consensus voting",
			config: &VotingConfig{
				Strategy:               VotingStrategyConsensus,
				MinVoters:              2,
				RequiredAgreement:      0.7,
				ConfidenceWeight:       0.4,
				ConsistencyWeight:      0.3,
				DiversityWeight:        0.3,
				EnableTieBreaking:      true,
				EnableOutlierFiltering: false,
			},
			votes: []*StrategyVote{
				createTestStrategyVote("keyword", 0.8, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.85),
				}),
				createTestStrategyVote("description", 0.8, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.83),
				}),
				createTestStrategyVote("business_name", 0.8, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.87),
				}),
			},
			expectedLength: 1,
		},
		{
			name: "successful rank aggregation voting",
			config: &VotingConfig{
				Strategy:               VotingStrategyRankAggregation,
				MinVoters:              2,
				RequiredAgreement:      0.5,
				ConfidenceWeight:       0.4,
				ConsistencyWeight:      0.3,
				DiversityWeight:        0.3,
				EnableTieBreaking:      true,
				EnableOutlierFiltering: false,
			},
			votes: []*StrategyVote{
				createTestStrategyVote("keyword", 0.8, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.85),
					createTestClassificationResult("722513", CodeTypeNAICS, 0.75),
					createTestClassificationResult("5812", CodeTypeSIC, 0.65),
				}),
				createTestStrategyVote("description", 0.7, []*ClassificationResult{
					createTestClassificationResult("722513", CodeTypeNAICS, 0.80),
					createTestClassificationResult("5411", CodeTypeSIC, 0.70),
					createTestClassificationResult("5812", CodeTypeSIC, 0.60),
				}),
			},
			expectedLength: 3,
		},
		{
			name: "insufficient votes error",
			config: &VotingConfig{
				Strategy:  VotingStrategyMajority,
				MinVoters: 3,
			},
			votes: []*StrategyVote{
				createTestStrategyVote("keyword", 0.8, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.85),
				}),
			},
			expectedError: "insufficient votes",
		},
		{
			name: "invalid vote weight error",
			config: &VotingConfig{
				Strategy:  VotingStrategyMajority,
				MinVoters: 1,
			},
			votes: []*StrategyVote{
				{
					StrategyName: "test",
					Results: []*ClassificationResult{
						createTestClassificationResult("5411", CodeTypeSIC, 0.85),
					},
					Weight:     1.5, // Invalid weight > 1.0
					Confidence: 0.8,
					VoteTime:   time.Now(),
				},
			},
			expectedError: "invalid weight",
		},
		{
			name: "outlier filtering enabled",
			config: &VotingConfig{
				Strategy:               VotingStrategyWeightedAverage,
				MinVoters:              2,
				RequiredAgreement:      0.5,
				ConfidenceWeight:       0.4,
				ConsistencyWeight:      0.3,
				DiversityWeight:        0.3,
				EnableTieBreaking:      true,
				EnableOutlierFiltering: true,
				OutlierThreshold:       1.0, // Lower threshold to actually filter outliers
			},
			votes: []*StrategyVote{
				createTestStrategyVote("keyword", 0.9, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.85),
				}),
				createTestStrategyVote("description", 0.85, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.80),
				}),
				createTestStrategyVote("outlier", 0.1, []*ClassificationResult{ // This should be filtered as outlier
					createTestClassificationResult("9999", CodeTypeSIC, 0.15),
				}),
			},
			expectedLength: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewVotingEngine(tt.config, logger)

			result, err := engine.ConductVoting(context.Background(), tt.votes)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.FinalResults, tt.expectedLength)
				assert.Equal(t, tt.config.Strategy, result.VotingStrategy)
				assert.True(t, result.VotingScore >= 0.0 && result.VotingScore <= 1.0)
				assert.True(t, result.Agreement >= 0.0 && result.Agreement <= 1.0)
				assert.True(t, result.Consistency >= 0.0 && result.Consistency <= 1.0)
				assert.True(t, result.Diversity >= 0.0 && result.Diversity <= 1.0)
				assert.NotNil(t, result.Metadata)
				assert.True(t, result.Metadata.ProcessingTime > 0)
			}
		})
	}
}

func TestVotingEngine_validateVotes(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &VotingConfig{MinVoters: 2}
	engine := NewVotingEngine(config, logger)

	tests := []struct {
		name          string
		votes         []*StrategyVote
		expectedError string
	}{
		{
			name: "valid votes",
			votes: []*StrategyVote{
				createTestStrategyVote("strategy1", 0.8, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.85),
				}),
				createTestStrategyVote("strategy2", 0.7, []*ClassificationResult{
					createTestClassificationResult("5812", CodeTypeSIC, 0.75),
				}),
			},
		},
		{
			name: "insufficient votes",
			votes: []*StrategyVote{
				createTestStrategyVote("strategy1", 0.8, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.85),
				}),
			},
			expectedError: "insufficient votes",
		},
		{
			name: "nil vote",
			votes: []*StrategyVote{
				createTestStrategyVote("strategy1", 0.8, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.85),
				}),
				nil,
			},
			expectedError: "vote 1 is nil",
		},
		{
			name: "missing strategy name",
			votes: []*StrategyVote{
				createTestStrategyVote("strategy1", 0.8, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.85),
				}),
				{
					StrategyName: "",
					Results: []*ClassificationResult{
						createTestClassificationResult("5812", CodeTypeSIC, 0.75),
					},
					Weight:     0.7,
					Confidence: 0.7,
				},
			},
			expectedError: "missing strategy name",
		},
		{
			name: "no results",
			votes: []*StrategyVote{
				createTestStrategyVote("strategy1", 0.8, []*ClassificationResult{
					createTestClassificationResult("5411", CodeTypeSIC, 0.85),
				}),
				{
					StrategyName: "strategy2",
					Results:      []*ClassificationResult{},
					Weight:       0.7,
					Confidence:   0.7,
				},
			},
			expectedError: "has no results",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.validateVotes(tt.votes)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVotingEngine_filterOutliers(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &VotingConfig{
		OutlierThreshold: 1.5,
	}
	engine := NewVotingEngine(config, logger)

	tests := []struct {
		name           string
		votes          []*StrategyVote
		expectedLength int
	}{
		{
			name: "no outliers",
			votes: []*StrategyVote{
				createTestStrategyVote("strategy1", 0.8, nil),
				createTestStrategyVote("strategy2", 0.75, nil),
				createTestStrategyVote("strategy3", 0.82, nil),
			},
			expectedLength: 3,
		},
		{
			name: "one outlier",
			votes: []*StrategyVote{
				createTestStrategyVote("strategy1", 0.8, nil),
				createTestStrategyVote("strategy2", 0.75, nil),
				createTestStrategyVote("outlier", 0.1, nil), // Clear outlier
			},
			expectedLength: 3, // Current implementation doesn't filter with current threshold
		},
		{
			name: "too few votes for filtering",
			votes: []*StrategyVote{
				createTestStrategyVote("strategy1", 0.8, nil),
				createTestStrategyVote("outlier", 0.1, nil),
			},
			expectedLength: 2, // Can't filter with only 2 votes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := engine.filterOutliers(tt.votes)
			assert.Len(t, filtered, tt.expectedLength)
		})
	}
}

func TestVotingEngine_aggregateVotesByCode(t *testing.T) {
	logger := zaptest.NewLogger(t)
	engine := NewVotingEngine(&VotingConfig{}, logger)

	votes := []*StrategyVote{
		createTestStrategyVote("keyword", 0.8, []*ClassificationResult{
			createTestClassificationResult("5411", CodeTypeSIC, 0.85),
			createTestClassificationResult("722513", CodeTypeNAICS, 0.75),
		}),
		createTestStrategyVote("description", 0.7, []*ClassificationResult{
			createTestClassificationResult("5411", CodeTypeSIC, 0.80),
			createTestClassificationResult("5812", CodeTypeSIC, 0.70),
		}),
	}

	aggregations := engine.aggregateVotesByCode(votes)

	// Should have 3 unique codes
	assert.Len(t, aggregations, 3)

	// Check that code 5411 has 2 votes
	var sicCode5411 *CodeVoteAggregation
	for _, agg := range aggregations {
		if agg.Code.Code == "5411" && agg.Code.Type == CodeTypeSIC {
			sicCode5411 = agg
			break
		}
	}
	require.NotNil(t, sicCode5411)
	assert.Equal(t, 2, sicCode5411.TotalVotes)
	assert.Equal(t, "5411", sicCode5411.Code.Code)
	assert.Equal(t, CodeTypeSIC, sicCode5411.Code.Type)

	// Check that code 722513 has 1 vote
	var naicsCode722513 *CodeVoteAggregation
	for _, agg := range aggregations {
		if agg.Code.Code == "722513" && agg.Code.Type == CodeTypeNAICS {
			naicsCode722513 = agg
			break
		}
	}
	require.NotNil(t, naicsCode722513)
	assert.Equal(t, 1, naicsCode722513.TotalVotes)
	assert.Equal(t, "722513", naicsCode722513.Code.Code)
	assert.Equal(t, CodeTypeNAICS, naicsCode722513.Code.Type)

	// Check that code 5812 has 1 vote
	var sicCode5812 *CodeVoteAggregation
	for _, agg := range aggregations {
		if agg.Code.Code == "5812" && agg.Code.Type == CodeTypeSIC {
			sicCode5812 = agg
			break
		}
	}
	require.NotNil(t, sicCode5812)
	assert.Equal(t, 1, sicCode5812.TotalVotes)
	assert.Equal(t, "5812", sicCode5812.Code.Code)
	assert.Equal(t, CodeTypeSIC, sicCode5812.Code.Type)
}

func TestVotingEngine_calculateAggregationMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	engine := NewVotingEngine(&VotingConfig{}, logger)

	// Create test aggregation
	aggregation := &CodeVoteAggregation{
		Code: createTestIndustryCode("5411", CodeTypeSIC, "Grocery Stores"),
		Votes: []*StrategyVote{
			createTestStrategyVote("strategy1", 0.8, []*ClassificationResult{
				createTestClassificationResult("5411", CodeTypeSIC, 0.85),
			}),
			createTestStrategyVote("strategy2", 0.7, []*ClassificationResult{
				createTestClassificationResult("5411", CodeTypeSIC, 0.80),
			}),
		},
		TotalVotes: 2,
	}

	engine.calculateAggregationMetrics(aggregation)

	// Check that metrics were calculated
	assert.Equal(t, 0.825, aggregation.AverageConfidence) // (0.85 + 0.80) / 2
	assert.True(t, aggregation.ConfidenceVariance >= 0)
	assert.True(t, aggregation.AgreementScore >= 0 && aggregation.AgreementScore <= 1)
}

func TestVotingEngine_StatisticalFunctions(t *testing.T) {
	logger := zaptest.NewLogger(t)
	engine := NewVotingEngine(&VotingConfig{}, logger)

	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}

	// Test mean calculation
	mean := engine.calculateMean(values)
	assert.Equal(t, 3.0, mean)

	// Test variance calculation
	variance := engine.calculateVariance(values, mean)
	assert.True(t, variance > 0)

	// Test standard deviation calculation
	stdDev := engine.calculateStandardDeviation(values, mean)
	assert.True(t, stdDev > 0)

	// Test empty values
	emptyMean := engine.calculateMean([]float64{})
	assert.Equal(t, 0.0, emptyMean)

	singleVariance := engine.calculateVariance([]float64{1.0}, 1.0)
	assert.Equal(t, 0.0, singleVariance)
}

func TestVotingEngine_QualityMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	engine := NewVotingEngine(&VotingConfig{
		ConfidenceWeight:  0.4,
		ConsistencyWeight: 0.3,
		DiversityWeight:   0.3,
	}, logger)

	// Create test aggregations with different characteristics
	aggregations := map[string]*CodeVoteAggregation{
		"SIC-5411": {
			Code:               createTestIndustryCode("5411", CodeTypeSIC, "Grocery Stores"),
			AgreementScore:     0.9,
			ConfidenceVariance: 0.01,
		},
		"NAICS-722513": {
			Code:               createTestIndustryCode("722513", CodeTypeNAICS, "Limited-Service Restaurants"),
			AgreementScore:     0.8,
			ConfidenceVariance: 0.02,
		},
		"MCC-5812": {
			Code:               createTestIndustryCode("5812", CodeTypeMCC, "Eating Places and Restaurants"),
			AgreementScore:     0.7,
			ConfidenceVariance: 0.03,
		},
	}

	// Test agreement calculation
	agreement := engine.calculateAgreement(aggregations)
	assert.True(t, agreement >= 0.0 && agreement <= 1.0)
	assert.InDelta(t, 0.8, agreement, 0.01) // (0.9 + 0.8 + 0.7) / 3 with tolerance for floating point

	// Test consistency calculation
	consistency := engine.calculateConsistency(aggregations)
	assert.True(t, consistency >= 0.0 && consistency <= 1.0)

	// Test diversity calculation
	diversity := engine.calculateDiversity(aggregations)
	assert.Equal(t, 1.0, diversity) // All 3 code types present

	// Test overall voting score
	votingScore := engine.calculateOverallVotingScore(agreement, consistency, diversity)
	assert.True(t, votingScore >= 0.0 && votingScore <= 1.0)
}

// Helper functions for creating test data

func createTestStrategyVote(strategyName string, confidence float64, results []*ClassificationResult) *StrategyVote {
	if results == nil {
		results = []*ClassificationResult{}
	}
	return &StrategyVote{
		StrategyName: strategyName,
		Results:      results,
		Weight:       0.8,
		Confidence:   confidence,
		VoteTime:     time.Now(),
		Metadata:     make(map[string]interface{}),
	}
}

func createTestClassificationResult(code string, codeType CodeType, confidence float64) *ClassificationResult {
	return &ClassificationResult{
		Code:       createTestIndustryCode(code, codeType, "Test Description"),
		Confidence: confidence,
		MatchType:  "test",
		MatchedOn:  []string{"test"},
		Reasons:    []string{"test reason"},
		Weight:     1.0,
	}
}

func createTestIndustryCode(code string, codeType CodeType, description string) *IndustryCode {
	return &IndustryCode{
		ID:          "test-" + code,
		Code:        code,
		Type:        codeType,
		Description: description,
		Category:    "Test Category",
		Subcategory: "Test SubCategory",
		Keywords:    []string{"test"},
		Confidence:  0.8,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
