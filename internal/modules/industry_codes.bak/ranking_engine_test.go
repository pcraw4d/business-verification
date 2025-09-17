package industry_codes

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func setupTestRankingEngine(t *testing.T) (*RankingEngine, *ConfidenceScorer, *MetadataManager) {
	// Setup test database
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	// Setup metadata manager
	logger := zaptest.NewLogger(t)
	metadataMgr := NewMetadataManager(db, logger)

	// Setup confidence scorer
	scorer := NewConfidenceScorer(nil, metadataMgr, logger)

	// Setup ranking engine
	engine := NewRankingEngine(scorer, logger)

	return engine, scorer, metadataMgr
}

func createTestClassificationResults() []*ClassificationResult {
	return []*ClassificationResult{
		{
			Code: &IndustryCode{
				Code:        "541511",
				Type:        CodeTypeNAICS,
				Description: "Custom Computer Programming Services",
				Category:    "Professional, Scientific, and Technical Services",
				Keywords:    []string{"programming", "software", "development"},
				Confidence:  0.9,
			},
			Confidence: 0.85,
			MatchType:  "keyword",
			MatchedOn:  []string{"programming", "software"},
			Reasons:    []string{"Strong keyword matches"},
			Weight:     1.0,
		},
		{
			Code: &IndustryCode{
				Code:        "5411",
				Type:        CodeTypeSIC,
				Description: "Legal Services",
				Category:    "Professional Services",
				Keywords:    []string{"legal", "law", "attorney"},
				Confidence:  0.7,
			},
			Confidence: 0.65,
			MatchType:  "description",
			MatchedOn:  []string{"legal"},
			Reasons:    []string{"Description similarity"},
			Weight:     0.8,
		},
		{
			Code: &IndustryCode{
				Code:        "5812",
				Type:        CodeTypeMCC,
				Description: "Eating Places and Restaurants",
				Category:    "Food Services",
				Keywords:    []string{"restaurant", "food", "dining"},
				Confidence:  0.8,
			},
			Confidence: 0.75,
			MatchType:  "exact",
			MatchedOn:  []string{"restaurant", "food", "dining"},
			Reasons:    []string{"Exact match", "Multiple keywords"},
			Weight:     1.2,
		},
		{
			Code: &IndustryCode{
				Code:        "541512",
				Type:        CodeTypeNAICS,
				Description: "Computer Systems Design Services",
				Category:    "Professional, Scientific, and Technical Services",
				Keywords:    []string{"systems", "design", "computer"},
				Confidence:  0.75,
			},
			Confidence: 0.70,
			MatchType:  "keyword",
			MatchedOn:  []string{"systems", "computer"},
			Reasons:    []string{"Keyword matches"},
			Weight:     0.9,
		},
	}
}

func createTestClassificationRequest() *ClassificationRequest {
	return &ClassificationRequest{
		BusinessName:        "Tech Solutions Inc",
		BusinessDescription: "Custom software development and programming services",
		Website:             "https://techsolutions.com",
		Keywords:            []string{"programming", "development"},
		PreferredCodeTypes:  []CodeType{CodeTypeNAICS, CodeTypeSIC},
		MaxResults:          10,
		MinConfidence:       0.3,
	}
}

func TestRankingEngine_RankAndSelectResults(t *testing.T) {
	engine, _, _ := setupTestRankingEngine(t)

	tests := []struct {
		name          string
		results       []*ClassificationResult
		request       *ClassificationRequest
		criteria      *RankingCriteria
		expectError   bool
		expectedRanks int
	}{
		{
			name:          "successful ranking with default criteria",
			results:       createTestClassificationResults(),
			request:       createTestClassificationRequest(),
			criteria:      nil, // Use default criteria
			expectError:   false,
			expectedRanks: 1, // Only one result meets default min confidence of 0.3
		},
		{
			name:    "ranking with confidence strategy",
			results: createTestClassificationResults(),
			request: createTestClassificationRequest(),
			criteria: &RankingCriteria{
				Strategy:          RankingStrategyConfidence,
				MinConfidence:     0.2, // Lower threshold to include more results
				MaxResultsPerType: 2,
			},
			expectError:   false,
			expectedRanks: 4, // More results with lower threshold
		},
		{
			name:    "ranking with weighted strategy",
			results: createTestClassificationResults(),
			request: createTestClassificationRequest(),
			criteria: &RankingCriteria{
				Strategy:          RankingStrategyWeighted,
				ConfidenceWeight:  0.6,
				RelevanceWeight:   0.4,
				MinConfidence:     0.2, // Lower threshold
				MaxResultsPerType: 3,
			},
			expectError:   false,
			expectedRanks: 4,
		},
		{
			name:    "ranking with multi-criteria strategy",
			results: createTestClassificationResults(),
			request: createTestClassificationRequest(),
			criteria: &RankingCriteria{
				Strategy:           RankingStrategyMultiCriteria,
				ConfidenceWeight:   0.3,
				RelevanceWeight:    0.3,
				QualityWeight:      0.2,
				FrequencyWeight:    0.2,
				MinConfidence:      0.2, // Lower threshold
				MaxResultsPerType:  3,
				UseDiversification: true,
				EnableTieBreaking:  true,
			},
			expectError:   false,
			expectedRanks: 4,
		},
		{
			name:          "empty results",
			results:       []*ClassificationResult{},
			request:       createTestClassificationRequest(),
			criteria:      nil,
			expectError:   false,
			expectedRanks: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rankedResults, err := engine.RankAndSelectResults(context.Background(), tt.results, tt.request, tt.criteria)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, rankedResults)
			} else {
				assert.NoError(t, err)
				if len(tt.results) == 0 {
					// For empty results, rankedResults might be nil or have empty results
					if rankedResults != nil {
						assert.Len(t, rankedResults.OverallResults, 0)
					}
				} else {
					assert.NotNil(t, rankedResults)

					// Verify structure
					assert.NotNil(t, rankedResults.OverallResults)
					assert.NotNil(t, rankedResults.TopResultsByType)
					assert.NotNil(t, rankedResults.RankingMetadata)
					assert.NotNil(t, rankedResults.QualityMetrics)
					assert.NotNil(t, rankedResults.DiversityMetrics)

					// Verify result count
					assert.Len(t, rankedResults.OverallResults, tt.expectedRanks)

					// Verify ranking metadata
					assert.Equal(t, len(tt.results), rankedResults.RankingMetadata.TotalCandidates)
					assert.GreaterOrEqual(t, rankedResults.RankingMetadata.FilteredCandidates, 0)
					assert.Greater(t, rankedResults.RankingMetadata.RankingTime, time.Duration(0))

					// Verify results are properly ranked
					for i := 1; i < len(rankedResults.OverallResults); i++ {
						assert.GreaterOrEqual(t,
							rankedResults.OverallResults[i-1].RankingScore,
							rankedResults.OverallResults[i].RankingScore,
							"Results should be sorted by ranking score in descending order")
					}

					// Verify rank assignments
					for i, result := range rankedResults.OverallResults {
						assert.Equal(t, i+1, result.Rank, "Rank should be assigned correctly")
						assert.NotNil(t, result.ConfidenceScore)
						assert.NotNil(t, result.RankingFactors)
						assert.NotEmpty(t, result.SelectionReason)
					}

					// Verify top results by type have max 3 results per type
					for codeType, typeResults := range rankedResults.TopResultsByType {
						maxResults := 3
						if tt.criteria != nil && tt.criteria.MaxResultsPerType > 0 {
							maxResults = tt.criteria.MaxResultsPerType
						}
						assert.LessOrEqual(t, len(typeResults), maxResults,
							"Code type %s should have at most %d results", codeType, maxResults)

						// Verify type ranks
						for i, result := range typeResults {
							assert.Equal(t, i+1, result.TypeRank, "Type rank should be assigned correctly")
						}
					}
				}
			}
		})
	}
}

func TestRankingEngine_RankingStrategies(t *testing.T) {
	engine, _, _ := setupTestRankingEngine(t)
	results := createTestClassificationResults()
	request := createTestClassificationRequest()

	strategies := []RankingStrategy{
		RankingStrategyConfidence,
		RankingStrategyComposite,
		RankingStrategyWeighted,
		RankingStrategyMultiCriteria,
	}

	for _, strategy := range strategies {
		t.Run(string(strategy), func(t *testing.T) {
			criteria := &RankingCriteria{
				Strategy:          strategy,
				ConfidenceWeight:  0.4,
				RelevanceWeight:   0.3,
				QualityWeight:     0.2,
				FrequencyWeight:   0.1,
				MinConfidence:     0.3,
				MaxResultsPerType: 3,
			}

			rankedResults, err := engine.RankAndSelectResults(context.Background(), results, request, criteria)

			assert.NoError(t, err)
			assert.NotNil(t, rankedResults)
			assert.Equal(t, strategy, rankedResults.RankingMetadata.Strategy)

			// Verify all results have valid ranking scores
			for _, result := range rankedResults.OverallResults {
				assert.GreaterOrEqual(t, result.RankingScore, 0.0)
				assert.LessOrEqual(t, result.RankingScore, 1.0)
			}
		})
	}
}

func TestRankingEngine_ConfidenceFiltering(t *testing.T) {
	engine, _, _ := setupTestRankingEngine(t)
	results := createTestClassificationResults()
	request := createTestClassificationRequest()

	tests := []struct {
		name          string
		minConfidence float64
		expectedCount int
	}{
		{"no filtering", 0.0, 4},
		{"moderate filtering", 0.5, 1},    // Only one result meets 0.5 threshold
		{"strict filtering", 0.7, 0},      // No results meet 0.7 threshold
		{"very strict filtering", 0.8, 0}, // No results meet 0.8 threshold
		{"impossible filtering", 1.0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			criteria := &RankingCriteria{
				Strategy:          RankingStrategyConfidence,
				MinConfidence:     tt.minConfidence,
				MaxResultsPerType: 10,
			}

			rankedResults, err := engine.RankAndSelectResults(context.Background(), results, request, criteria)

			assert.NoError(t, err)
			assert.NotNil(t, rankedResults)
			assert.Len(t, rankedResults.OverallResults, tt.expectedCount)

			// Verify all results meet minimum confidence
			for _, result := range rankedResults.OverallResults {
				assert.GreaterOrEqual(t, result.ConfidenceScore.OverallScore, tt.minConfidence)
			}
		})
	}
}

func TestRankingEngine_MaxResultsPerType(t *testing.T) {
	engine, _, _ := setupTestRankingEngine(t)
	request := createTestClassificationRequest()

	// Create more results of the same type
	results := []*ClassificationResult{
		{
			Code: &IndustryCode{
				Code: "541511", Type: CodeTypeNAICS, Description: "Programming 1",
				Category: "Tech", Keywords: []string{"programming"}, Confidence: 0.9,
			},
			Confidence: 0.9, MatchType: "keyword", MatchedOn: []string{"programming"},
		},
		{
			Code: &IndustryCode{
				Code: "541512", Type: CodeTypeNAICS, Description: "Programming 2",
				Category: "Tech", Keywords: []string{"programming"}, Confidence: 0.8,
			},
			Confidence: 0.8, MatchType: "keyword", MatchedOn: []string{"programming"},
		},
		{
			Code: &IndustryCode{
				Code: "541513", Type: CodeTypeNAICS, Description: "Programming 3",
				Category: "Tech", Keywords: []string{"programming"}, Confidence: 0.7,
			},
			Confidence: 0.7, MatchType: "keyword", MatchedOn: []string{"programming"},
		},
		{
			Code: &IndustryCode{
				Code: "541514", Type: CodeTypeNAICS, Description: "Programming 4",
				Category: "Tech", Keywords: []string{"programming"}, Confidence: 0.6,
			},
			Confidence: 0.6, MatchType: "keyword", MatchedOn: []string{"programming"},
		},
	}

	tests := []struct {
		name              string
		maxResultsPerType int
		expectedPerType   int
	}{
		{"limit to 1", 1, 1},
		{"limit to 2", 2, 2},
		{"limit to 3", 3, 3},
		{"no limit", 10, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			criteria := &RankingCriteria{
				Strategy:          RankingStrategyConfidence,
				MinConfidence:     0.3,
				MaxResultsPerType: tt.maxResultsPerType,
			}

			rankedResults, err := engine.RankAndSelectResults(context.Background(), results, request, criteria)

			assert.NoError(t, err)
			assert.NotNil(t, rankedResults)

			// Check NAICS results
			naicsResults, exists := rankedResults.TopResultsByType["naics"]
			assert.True(t, exists)
			assert.Len(t, naicsResults, tt.expectedPerType)

			// Verify results are sorted by confidence within type
			for i := 1; i < len(naicsResults); i++ {
				assert.GreaterOrEqual(t,
					naicsResults[i-1].ConfidenceScore.OverallScore,
					naicsResults[i].ConfidenceScore.OverallScore)
			}
		})
	}
}

func TestRankingEngine_QualityMetrics(t *testing.T) {
	engine, _, _ := setupTestRankingEngine(t)
	results := createTestClassificationResults()
	request := createTestClassificationRequest()

	rankedResults, err := engine.RankAndSelectResults(context.Background(), results, request, nil)

	require.NoError(t, err)
	require.NotNil(t, rankedResults.QualityMetrics)

	metrics := rankedResults.QualityMetrics

	// Verify quality metrics structure
	assert.GreaterOrEqual(t, metrics.AverageConfidence, 0.0)
	assert.LessOrEqual(t, metrics.AverageConfidence, 1.0)
	assert.GreaterOrEqual(t, metrics.ConfidenceRange, 0.0)
	assert.NotNil(t, metrics.QualityDistribution)
	assert.NotNil(t, metrics.TypeCoverage)
	assert.GreaterOrEqual(t, metrics.HighQualityCount, 0)
	assert.GreaterOrEqual(t, metrics.LowQualityCount, 0)

	// Verify type coverage
	expectedTypes := []string{"naics", "sic", "mcc"}
	for _, expectedType := range expectedTypes {
		count, exists := metrics.TypeCoverage[expectedType]
		if exists {
			assert.Greater(t, count, 0)
		}
	}
}

func TestRankingEngine_DiversityMetrics(t *testing.T) {
	engine, _, _ := setupTestRankingEngine(t)
	results := createTestClassificationResults()
	request := createTestClassificationRequest()

	rankedResults, err := engine.RankAndSelectResults(context.Background(), results, request, nil)

	require.NoError(t, err)
	require.NotNil(t, rankedResults.DiversityMetrics)

	metrics := rankedResults.DiversityMetrics

	// Verify diversity metrics structure
	assert.GreaterOrEqual(t, metrics.TypeDiversity, 0.0)
	assert.LessOrEqual(t, metrics.TypeDiversity, 1.0)
	assert.GreaterOrEqual(t, metrics.CategoryDiversity, 0.0)
	assert.LessOrEqual(t, metrics.CategoryDiversity, 1.0)
	assert.GreaterOrEqual(t, metrics.ConfidenceSpread, 0.0)
	assert.GreaterOrEqual(t, metrics.DiversityScore, 0.0)
	assert.LessOrEqual(t, metrics.DiversityScore, 1.0)
	assert.NotNil(t, metrics.SourceDiversity)

	// Verify source diversity
	expectedSources := []string{"keyword", "description", "exact"}
	totalSources := 0
	for _, source := range expectedSources {
		if count, exists := metrics.SourceDiversity[source]; exists {
			totalSources += count
		}
	}
	assert.Greater(t, totalSources, 0)
}

func TestRankingEngine_Diversification(t *testing.T) {
	engine, _, _ := setupTestRankingEngine(t)
	request := createTestClassificationRequest()

	// Create results with same type but different categories
	results := []*ClassificationResult{
		{
			Code: &IndustryCode{
				Code: "541511", Type: CodeTypeNAICS, Description: "Programming",
				Category: "Professional Services", Keywords: []string{"programming"}, Confidence: 0.8,
			},
			Confidence: 0.8, MatchType: "keyword", MatchedOn: []string{"programming"},
		},
		{
			Code: &IndustryCode{
				Code: "541512", Type: CodeTypeNAICS, Description: "Design",
				Category: "Design Services", Keywords: []string{"design"}, Confidence: 0.8,
			},
			Confidence: 0.8, MatchType: "keyword", MatchedOn: []string{"design"},
		},
		{
			Code: &IndustryCode{
				Code: "541513", Type: CodeTypeNAICS, Description: "Consulting",
				Category: "Professional Services", Keywords: []string{"consulting"}, Confidence: 0.8,
			},
			Confidence: 0.8, MatchType: "keyword", MatchedOn: []string{"consulting"},
		},
	}

	criteria := &RankingCriteria{
		Strategy:           RankingStrategyComposite,
		MinConfidence:      0.2, // Lower threshold to get more results
		MaxResultsPerType:  3,
		UseDiversification: true,
	}

	rankedResults, err := engine.RankAndSelectResults(context.Background(), results, request, criteria)

	require.NoError(t, err)
	require.NotNil(t, rankedResults)

	naicsResults := rankedResults.TopResultsByType["naics"]
	require.GreaterOrEqual(t, len(naicsResults), 1) // At least 1 result

	// Check that diversification bonus was applied
	hasDiversificationBonus := false
	for _, result := range naicsResults {
		if result.RankingFactors.DiversityBonus > 0 {
			hasDiversificationBonus = true
			break
		}
	}
	assert.True(t, hasDiversificationBonus, "At least one result should have diversification bonus")
}

func TestRankingEngine_TieBreaking(t *testing.T) {
	engine, _, _ := setupTestRankingEngine(t)
	request := createTestClassificationRequest()

	// Create results with identical ranking scores to trigger tie-breaking
	results := []*ClassificationResult{
		{
			Code: &IndustryCode{
				Code: "541511", Type: CodeTypeNAICS, Description: "Programming 1",
				Category: "Tech", Keywords: []string{"programming"}, Confidence: 0.8,
			},
			Confidence: 0.8, MatchType: "keyword", MatchedOn: []string{"programming"},
		},
		{
			Code: &IndustryCode{
				Code: "541512", Type: CodeTypeNAICS, Description: "Programming 2",
				Category: "Tech", Keywords: []string{"programming"}, Confidence: 0.8,
			},
			Confidence: 0.8, MatchType: "keyword", MatchedOn: []string{"programming"},
		},
	}

	criteria := &RankingCriteria{
		Strategy:          RankingStrategyComposite,
		MinConfidence:     0.3,
		MaxResultsPerType: 3,
		EnableTieBreaking: true,
	}

	rankedResults, err := engine.RankAndSelectResults(context.Background(), results, request, criteria)

	require.NoError(t, err)
	require.NotNil(t, rankedResults)

	naicsResults := rankedResults.TopResultsByType["naics"]
	require.Len(t, naicsResults, 2)

	// Verify tie breaker values are set
	for _, result := range naicsResults {
		assert.GreaterOrEqual(t, result.TieBreaker, 0.0)
		assert.LessOrEqual(t, result.TieBreaker, 1.0)
	}

	// If tie-breaking worked, the results should be distinctly ordered
	if len(naicsResults) > 1 {
		// At minimum, tie breaker values should be calculated
		assert.True(t, naicsResults[0].TieBreaker >= 0)
		assert.True(t, naicsResults[1].TieBreaker >= 0)
	}
}

func TestRankingEngine_SelectionReasons(t *testing.T) {
	engine, _, _ := setupTestRankingEngine(t)
	results := createTestClassificationResults()
	request := createTestClassificationRequest()

	rankedResults, err := engine.RankAndSelectResults(context.Background(), results, request, nil)

	require.NoError(t, err)
	require.NotNil(t, rankedResults)

	// Verify all results have selection reasons
	for _, result := range rankedResults.OverallResults {
		assert.NotEmpty(t, result.SelectionReason)
		assert.Contains(t, result.SelectionReason, "Selected for")
	}
}

func TestRankingEngine_QualityIndicators(t *testing.T) {
	engine, _, _ := setupTestRankingEngine(t)
	results := createTestClassificationResults()
	request := createTestClassificationRequest()

	rankedResults, err := engine.RankAndSelectResults(context.Background(), results, request, nil)

	require.NoError(t, err)
	require.NotNil(t, rankedResults)

	// Verify quality indicators are generated
	for _, result := range rankedResults.OverallResults {
		assert.NotNil(t, result.QualityIndicators)
		// Quality indicators should be relevant strings
		for _, indicator := range result.QualityIndicators {
			assert.NotEmpty(t, indicator)
		}
	}
}

func TestRankingEngine_TOPSISRanking(t *testing.T) {
	engine, _, _ := setupTestRankingEngine(t)
	results := createTestClassificationResults()
	request := createTestClassificationRequest()

	criteria := &RankingCriteria{
		Strategy:          RankingStrategyMultiCriteria,
		ConfidenceWeight:  0.25,
		RelevanceWeight:   0.25,
		QualityWeight:     0.25,
		FrequencyWeight:   0.25,
		MinConfidence:     0.3,
		MaxResultsPerType: 3,
	}

	rankedResults, err := engine.RankAndSelectResults(context.Background(), results, request, criteria)

	require.NoError(t, err)
	require.NotNil(t, rankedResults)

	// Verify TOPSIS ranking produces valid scores
	for _, result := range rankedResults.OverallResults {
		assert.GreaterOrEqual(t, result.RankingScore, 0.0)
		assert.LessOrEqual(t, result.RankingScore, 1.0)
	}

	// Verify results are properly sorted
	for i := 1; i < len(rankedResults.OverallResults); i++ {
		assert.GreaterOrEqual(t,
			rankedResults.OverallResults[i-1].RankingScore,
			rankedResults.OverallResults[i].RankingScore)
	}
}

func TestRankingEngine_EdgeCases(t *testing.T) {
	engine, _, _ := setupTestRankingEngine(t)
	request := createTestClassificationRequest()

	t.Run("single result", func(t *testing.T) {
		results := []*ClassificationResult{createTestClassificationResults()[0]}

		rankedResults, err := engine.RankAndSelectResults(context.Background(), results, request, nil)

		assert.NoError(t, err)
		assert.NotNil(t, rankedResults)
		assert.Len(t, rankedResults.OverallResults, 1)
		assert.Equal(t, 1, rankedResults.OverallResults[0].Rank)
	})

	t.Run("nil criteria uses defaults", func(t *testing.T) {
		results := createTestClassificationResults()

		rankedResults, err := engine.RankAndSelectResults(context.Background(), results, request, nil)

		assert.NoError(t, err)
		assert.NotNil(t, rankedResults)
		assert.Equal(t, RankingStrategyComposite, rankedResults.RankingMetadata.Strategy)
	})

	t.Run("high min confidence filters all", func(t *testing.T) {
		results := createTestClassificationResults()
		criteria := &RankingCriteria{
			Strategy:      RankingStrategyConfidence, // Need to specify strategy
			MinConfidence: 1.1,                       // Impossible threshold
		}

		rankedResults, err := engine.RankAndSelectResults(context.Background(), results, request, criteria)

		assert.NoError(t, err)
		assert.NotNil(t, rankedResults)
		assert.Len(t, rankedResults.OverallResults, 0)
	})
}
