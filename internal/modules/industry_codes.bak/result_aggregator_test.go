package industry_codes

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func setupTestResultAggregator(t *testing.T) (*ResultAggregator, func()) {
	db, cleanup := setupTestDatabase(t)

	database := NewIndustryCodeDatabase(db, zaptest.NewLogger(t))
	metadataManager := NewMetadataManager(db, zaptest.NewLogger(t))
	confidenceScorer := NewConfidenceScorer(database, metadataManager, zaptest.NewLogger(t))
	rankingEngine := NewRankingEngine(confidenceScorer, zaptest.NewLogger(t))
	aggregator := NewResultAggregator(confidenceScorer, rankingEngine, zaptest.NewLogger(t))

	return aggregator, cleanup
}

func createTestResultsForAggregation() []*ClassificationResult {
	return []*ClassificationResult{
		{
			Code: &IndustryCode{
				Code:        "5411",
				Type:        CodeTypeSIC,
				Description: "Legal Services",
				Category:    "Professional Services",
			},
			Confidence: 0.85,
			MatchType:  "exact",
			MatchedOn:  []string{"legal services"},
			Reasons:    []string{"exact match on business description"},
			Weight:     1.0,
		},
		{
			Code: &IndustryCode{
				Code:        "541110",
				Type:        CodeTypeNAICS,
				Description: "Offices of Lawyers",
				Category:    "Professional Services",
			},
			Confidence: 0.80,
			MatchType:  "keyword",
			MatchedOn:  []string{"legal", "offices"},
			Reasons:    []string{"keyword match on legal", "office type match"},
			Weight:     0.9,
		},
		{
			Code: &IndustryCode{
				Code:        "8111",
				Type:        CodeTypeMCC,
				Description: "Legal Services",
				Category:    "Professional Services",
			},
			Confidence: 0.75,
			MatchType:  "fuzzy",
			MatchedOn:  []string{"legal"},
			Reasons:    []string{"fuzzy match on legal services"},
			Weight:     0.8,
		},
		{
			Code: &IndustryCode{
				Code:        "5412",
				Type:        CodeTypeSIC,
				Description: "Accounting Services",
				Category:    "Professional Services",
			},
			Confidence: 0.45,
			MatchType:  "keyword",
			MatchedOn:  []string{"services"},
			Reasons:    []string{"weak keyword match"},
			Weight:     0.3,
		},
		{
			Code: &IndustryCode{
				Code:        "5413",
				Type:        CodeTypeSIC,
				Description: "Consulting Services",
				Category:    "Professional Services",
			},
			Confidence: 0.30,
			MatchType:  "fuzzy",
			MatchedOn:  []string{"services"},
			Reasons:    []string{"weak fuzzy match"},
			Weight:     0.2,
		},
	}
}

func TestResultAggregator_AggregateAndPresent(t *testing.T) {
	aggregator, cleanup := setupTestResultAggregator(t)
	defer cleanup()

	tests := []struct {
		name          string
		request       *AggregationRequest
		expectedError bool
		checkResults  func(t *testing.T, results *AggregatedResults)
	}{
		{
			name: "basic aggregation with default settings",
			request: &AggregationRequest{
				Results:           createTestResultsForAggregation(),
				MaxResultsPerType: 3,
				MinConfidence:     0.3,
				IncludeMetadata:   true,
				IncludeAnalytics:  false,
				SortBy:            SortByConfidence,
				Presentation:      PresentationSummary,
			},
			expectedError: false,
			checkResults: func(t *testing.T, results *AggregatedResults) {
				assert.NotNil(t, results)
				assert.NotEmpty(t, results.AllResults)
				assert.NotNil(t, results.AggregationMetadata)
				assert.Equal(t, 5, results.AggregationMetadata.TotalInputResults)
				assert.Greater(t, results.AggregationMetadata.AggregatedCount, 0)

				// Check that results are sorted by confidence
				for i := 1; i < len(results.AllResults); i++ {
					assert.GreaterOrEqual(t, results.AllResults[i-1].Confidence, results.AllResults[i].Confidence)
				}
			},
		},
		{
			name: "aggregation with analytics enabled",
			request: &AggregationRequest{
				Results:           createTestResultsForAggregation(),
				MaxResultsPerType: 3,
				MinConfidence:     0.3,
				IncludeMetadata:   true,
				IncludeAnalytics:  true,
				SortBy:            SortByRelevance,
				Presentation:      PresentationDetailed,
			},
			expectedError: false,
			checkResults: func(t *testing.T, results *AggregatedResults) {
				assert.NotNil(t, results.Analytics)
				assert.NotNil(t, results.Analytics.ConfidenceStats)
				assert.NotNil(t, results.Analytics.QualityMetrics)
				assert.NotNil(t, results.Analytics.DiversityMetrics)
				assert.Greater(t, results.Analytics.ConfidenceStats.Mean, 0.0)
			},
		},
		{
			name: "high confidence filtering",
			request: &AggregationRequest{
				Results:           createTestResultsForAggregation(),
				MaxResultsPerType: 3,
				MinConfidence:     0.7,
				IncludeMetadata:   true,
				SortBy:            SortByConfidence,
				Presentation:      PresentationCompact,
			},
			expectedError: false,
			checkResults: func(t *testing.T, results *AggregatedResults) {
				// Should filter out low confidence results
				assert.LessOrEqual(t, len(results.AllResults), 3)
				for _, result := range results.AllResults {
					assert.GreaterOrEqual(t, result.Confidence, 0.7)
				}
			},
		},
		{
			name: "group by strategy",
			request: &AggregationRequest{
				Results:           createTestResultsForAggregation(),
				MaxResultsPerType: 3,
				MinConfidence:     0.3,
				GroupByStrategy:   true,
				SortBy:            SortByMatchStrength,
				Presentation:      PresentationAPI,
			},
			expectedError: false,
			checkResults: func(t *testing.T, results *AggregatedResults) {
				assert.NotNil(t, results.ResultsByStrategy)
				assert.NotEmpty(t, results.ResultsByStrategy)

				// Should have different strategy groups
				strategies := []string{"exact", "keyword", "fuzzy"}
				for _, strategy := range strategies {
					if strategyResults, exists := results.ResultsByStrategy[strategy]; exists {
						assert.NotEmpty(t, strategyResults)
					}
				}
			},
		},
		{
			name: "empty results",
			request: &AggregationRequest{
				Results:           []*ClassificationResult{},
				MaxResultsPerType: 3,
				MinConfidence:     0.3,
				SortBy:            SortByConfidence,
				Presentation:      PresentationSummary,
			},
			expectedError: false,
			checkResults: func(t *testing.T, results *AggregatedResults) {
				assert.Empty(t, results.AllResults)
				assert.Empty(t, results.TopThreeByType)
				assert.Equal(t, 0, results.AggregationMetadata.TotalInputResults)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := aggregator.AggregateAndPresent(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, results)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, results)
				if tt.checkResults != nil {
					tt.checkResults(t, results)
				}
			}
		})
	}
}

func TestResultAggregator_TopThreeByType(t *testing.T) {
	aggregator, cleanup := setupTestResultAggregator(t)
	defer cleanup()

	results := createTestResultsForAggregation()

	request := &AggregationRequest{
		Results:           results,
		MaxResultsPerType: 3,
		MinConfidence:     0.3,
		SortBy:            SortByConfidence,
		Presentation:      PresentationSummary,
	}

	aggregatedResults, err := aggregator.AggregateAndPresent(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, aggregatedResults)

	// Check that we have results grouped by type
	assert.NotEmpty(t, aggregatedResults.TopThreeByType)

	// Check SIC results (should have 3 results)
	sicResults, exists := aggregatedResults.TopThreeByType["sic"]
	assert.True(t, exists)
	assert.LessOrEqual(t, len(sicResults), 3)

	// Results within each type should be sorted by aggregation score
	for _, result := range sicResults {
		assert.Equal(t, CodeTypeSIC, result.Code.Type)
	}

	// Check NAICS results
	naicsResults, exists := aggregatedResults.TopThreeByType["naics"]
	assert.True(t, exists)
	assert.LessOrEqual(t, len(naicsResults), 3)

	// Check MCC results
	mccResults, exists := aggregatedResults.TopThreeByType["mcc"]
	assert.True(t, exists)
	assert.LessOrEqual(t, len(mccResults), 3)
}

func TestResultAggregator_SortingStrategies(t *testing.T) {
	aggregator, cleanup := setupTestResultAggregator(t)
	defer cleanup()

	results := createTestResultsForAggregation()

	tests := []struct {
		name         string
		sortBy       SortCriteria
		checkSorting func(t *testing.T, results []*AggregatedResult)
	}{
		{
			name:   "sort by confidence",
			sortBy: SortByConfidence,
			checkSorting: func(t *testing.T, results []*AggregatedResult) {
				for i := 1; i < len(results); i++ {
					assert.GreaterOrEqual(t, results[i-1].Confidence, results[i].Confidence)
				}
			},
		},
		{
			name:   "sort by relevance (aggregation score)",
			sortBy: SortByRelevance,
			checkSorting: func(t *testing.T, results []*AggregatedResult) {
				for i := 1; i < len(results); i++ {
					assert.GreaterOrEqual(t, results[i-1].AggregationScore, results[i].AggregationScore)
				}
			},
		},
		{
			name:   "sort by code type",
			sortBy: SortByCodeType,
			checkSorting: func(t *testing.T, results []*AggregatedResult) {
				// Should be sorted by type first, then confidence within type
				for i := 1; i < len(results); i++ {
					if results[i-1].Code.Type == results[i].Code.Type {
						assert.GreaterOrEqual(t, results[i-1].Confidence, results[i].Confidence)
					}
				}
			},
		},
		{
			name:   "sort alphabetically",
			sortBy: SortByAlphabetical,
			checkSorting: func(t *testing.T, results []*AggregatedResult) {
				for i := 1; i < len(results); i++ {
					assert.LessOrEqual(t, results[i-1].Code.Description, results[i].Code.Description)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &AggregationRequest{
				Results:           results,
				MaxResultsPerType: 3,
				MinConfidence:     0.3,
				SortBy:            tt.sortBy,
				Presentation:      PresentationSummary,
			}

			aggregatedResults, err := aggregator.AggregateAndPresent(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, aggregatedResults)

			tt.checkSorting(t, aggregatedResults.AllResults)
		})
	}
}

func TestResultAggregator_PresentationFormats(t *testing.T) {
	aggregator, cleanup := setupTestResultAggregator(t)
	defer cleanup()

	results := createTestResultsForAggregation()

	tests := []struct {
		name         string
		presentation PresentationFormat
		checkFormat  func(t *testing.T, data *PresentationData)
	}{
		{
			name:         "detailed presentation",
			presentation: PresentationDetailed,
			checkFormat: func(t *testing.T, data *PresentationData) {
				assert.NotNil(t, data.DetailedView)
				assert.NotEmpty(t, data.DetailedView.MethodologyNotes)
				assert.NotEmpty(t, data.DetailedView.ConfidenceExplanation)
			},
		},
		{
			name:         "summary presentation",
			presentation: PresentationSummary,
			checkFormat: func(t *testing.T, data *PresentationData) {
				assert.NotNil(t, data.SummaryView)
				assert.NotEmpty(t, data.SummaryView.TopThree)
				assert.NotEmpty(t, data.SummaryView.KeyMetrics)
				assert.NotEmpty(t, data.SummaryView.QuickSummary)
			},
		},
		{
			name:         "compact presentation",
			presentation: PresentationCompact,
			checkFormat: func(t *testing.T, data *PresentationData) {
				assert.NotNil(t, data.CompactView)
				assert.NotNil(t, data.CompactView.BestMatch)
				assert.NotEmpty(t, data.CompactView.ConfidenceIndicator)
			},
		},
		{
			name:         "export presentation",
			presentation: PresentationExport,
			checkFormat: func(t *testing.T, data *PresentationData) {
				assert.NotNil(t, data.ExportData)
				assert.NotEmpty(t, data.ExportData.Headers)
				assert.NotEmpty(t, data.ExportData.CSVData)
				assert.Greater(t, len(data.ExportData.CSVData), 1) // Headers + data
			},
		},
		{
			name:         "dashboard presentation",
			presentation: PresentationDashboard,
			checkFormat: func(t *testing.T, data *PresentationData) {
				assert.NotNil(t, data.DashboardData)
				assert.NotEmpty(t, data.DashboardData.Widgets)
				assert.NotEmpty(t, data.DashboardData.KPIs)
			},
		},
		{
			name:         "API presentation",
			presentation: PresentationAPI,
			checkFormat: func(t *testing.T, data *PresentationData) {
				assert.NotNil(t, data.APIResponse)
				assert.Equal(t, "success", data.APIResponse.Status)
				assert.NotNil(t, data.APIResponse.Data)
				assert.NotEmpty(t, data.APIResponse.Metadata)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &AggregationRequest{
				Results:           results,
				MaxResultsPerType: 3,
				MinConfidence:     0.3,
				IncludeAnalytics:  true,
				SortBy:            SortByConfidence,
				Presentation:      tt.presentation,
			}

			aggregatedResults, err := aggregator.AggregateAndPresent(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, aggregatedResults)
			require.NotNil(t, aggregatedResults.PresentationData)

			assert.Equal(t, tt.presentation, aggregatedResults.PresentationData.Format)
			tt.checkFormat(t, aggregatedResults.PresentationData)
		})
	}
}

func TestResultAggregator_ConfidenceLevels(t *testing.T) {
	aggregator, cleanup := setupTestResultAggregator(t)
	defer cleanup()

	tests := []struct {
		name             string
		confidence       float64
		expectedLevel    ConfidenceLevel
		expectedStrength MatchStrength
	}{
		{
			name:             "very high confidence",
			confidence:       0.95,
			expectedLevel:    ConfidenceLevelVeryHigh,
			expectedStrength: MatchStrengthExact,
		},
		{
			name:             "high confidence",
			confidence:       0.80,
			expectedLevel:    ConfidenceLevelHigh,
			expectedStrength: MatchStrengthStrong,
		},
		{
			name:             "medium confidence",
			confidence:       0.60,
			expectedLevel:    ConfidenceLevelMedium,
			expectedStrength: MatchStrengthModerate,
		},
		{
			name:             "low confidence",
			confidence:       0.35,
			expectedLevel:    ConfidenceLevelLow,
			expectedStrength: MatchStrengthWeak,
		},
		{
			name:             "very low confidence",
			confidence:       0.15,
			expectedLevel:    ConfidenceLevelVeryLow,
			expectedStrength: MatchStrengthMinimal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := aggregator.determineConfidenceLevel(tt.confidence)
			assert.Equal(t, tt.expectedLevel, level)

			// Create a test result to check match strength
			result := &ClassificationResult{
				Code: &IndustryCode{
					Code:        "TEST",
					Type:        CodeTypeSIC,
					Description: "Test Code",
				},
				Confidence: tt.confidence,
				MatchType:  "test",
			}

			strength := aggregator.determineMatchStrength(result)
			assert.Equal(t, tt.expectedStrength, strength)
		})
	}
}

func TestResultAggregator_QualityIndicators(t *testing.T) {
	aggregator, cleanup := setupTestResultAggregator(t)
	defer cleanup()

	tests := []struct {
		name         string
		result       *ClassificationResult
		checkQuality func(t *testing.T, indicators []string)
	}{
		{
			name: "high quality exact match",
			result: &ClassificationResult{
				Code: &IndustryCode{
					Code:        "5411",
					Type:        CodeTypeSIC,
					Description: "Legal Services - detailed description of legal services provided",
					Category:    "Professional Services",
				},
				Confidence: 0.95,
				MatchType:  "exact",
				MatchedOn:  []string{"legal", "services", "professional"},
				Reasons:    []string{"exact match", "category match", "keyword match"},
			},
			checkQuality: func(t *testing.T, indicators []string) {
				assert.Contains(t, indicators, "very_high_confidence")
				assert.Contains(t, indicators, "exact_match")
				assert.Contains(t, indicators, "multiple_evidence_points")
				assert.Contains(t, indicators, "multiple_match_terms")
				assert.Contains(t, indicators, "detailed_description")
				assert.Contains(t, indicators, "categorized")
			},
		},
		{
			name: "medium quality fuzzy match",
			result: &ClassificationResult{
				Code: &IndustryCode{
					Code:        "5412",
					Type:        CodeTypeSIC,
					Description: "Accounting",
					Category:    "",
				},
				Confidence: 0.60,
				MatchType:  "fuzzy",
				MatchedOn:  []string{"accounting"},
				Reasons:    []string{"fuzzy match"},
			},
			checkQuality: func(t *testing.T, indicators []string) {
				assert.NotContains(t, indicators, "very_high_confidence")
				assert.NotContains(t, indicators, "exact_match")
				assert.NotContains(t, indicators, "multiple_evidence_points")
				assert.NotContains(t, indicators, "detailed_description")
				assert.NotContains(t, indicators, "categorized")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indicators := aggregator.analyzeQualityIndicators(tt.result)
			tt.checkQuality(t, indicators)
		})
	}
}

func TestResultAggregator_Analytics(t *testing.T) {
	aggregator, cleanup := setupTestResultAggregator(t)
	defer cleanup()

	results := createTestResultsForAggregation()

	request := &AggregationRequest{
		Results:           results,
		MaxResultsPerType: 3,
		MinConfidence:     0.3,
		IncludeAnalytics:  true,
		SortBy:            SortByConfidence,
		Presentation:      PresentationDetailed,
	}

	aggregatedResults, err := aggregator.AggregateAndPresent(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, aggregatedResults)
	require.NotNil(t, aggregatedResults.Analytics)

	analytics := aggregatedResults.Analytics

	// Test confidence statistics
	assert.NotNil(t, analytics.ConfidenceStats)
	assert.Greater(t, analytics.ConfidenceStats.Mean, 0.0)
	assert.Greater(t, analytics.ConfidenceStats.Max, analytics.ConfidenceStats.Min)
	assert.GreaterOrEqual(t, analytics.ConfidenceStats.StdDev, 0.0)
	assert.Len(t, analytics.ConfidenceStats.Quartiles, 3)

	// Test quality metrics
	assert.NotNil(t, analytics.QualityMetrics)
	assert.Greater(t, analytics.QualityMetrics.OverallQuality, 0.0)
	assert.NotEmpty(t, analytics.QualityMetrics.QualityByType)

	// Test diversity metrics
	assert.NotNil(t, analytics.DiversityMetrics)
	assert.GreaterOrEqual(t, analytics.DiversityMetrics.DiversityScore, 0.0)
	assert.LessOrEqual(t, analytics.DiversityMetrics.DiversityScore, 1.0)

	// Test recommendation and certainty scores
	assert.GreaterOrEqual(t, analytics.RecommendationScore, 0.0)
	assert.LessOrEqual(t, analytics.RecommendationScore, 1.0)
	assert.GreaterOrEqual(t, analytics.Certainty, 0.0)
	assert.LessOrEqual(t, analytics.Certainty, 1.0)
}

func TestResultAggregator_Deduplication(t *testing.T) {
	aggregator, cleanup := setupTestResultAggregator(t)
	defer cleanup()

	// Create duplicate results
	results := []*ClassificationResult{
		{
			Code: &IndustryCode{
				Code:        "5411",
				Type:        CodeTypeSIC,
				Description: "Legal Services",
			},
			Confidence: 0.85,
			MatchType:  "exact",
			MatchedOn:  []string{"legal"},
			Reasons:    []string{"exact match"},
		},
		{
			Code: &IndustryCode{
				Code:        "5411",
				Type:        CodeTypeSIC,
				Description: "Legal Services",
			},
			Confidence: 0.75, // Lower confidence, should be merged
			MatchType:  "keyword",
			MatchedOn:  []string{"services"},
			Reasons:    []string{"keyword match"},
		},
		{
			Code: &IndustryCode{
				Code:        "5412",
				Type:        CodeTypeSIC,
				Description: "Accounting Services",
			},
			Confidence: 0.70,
			MatchType:  "fuzzy",
			MatchedOn:  []string{"accounting"},
			Reasons:    []string{"fuzzy match"},
		},
	}

	deduplicated := aggregator.deduplicateResults(results)

	// Should have 2 unique results (one merged, one separate)
	assert.Len(t, deduplicated, 2)

	// Find the merged result for 5411
	var mergedResult *ClassificationResult
	for _, result := range deduplicated {
		if result.Code.Code == "5411" {
			mergedResult = result
			break
		}
	}

	require.NotNil(t, mergedResult)

	// Should have highest confidence from merged results
	assert.Equal(t, 0.85, mergedResult.Confidence)

	// Should have combined match terms and reasons
	assert.Contains(t, mergedResult.MatchedOn, "legal")
	assert.Contains(t, mergedResult.MatchedOn, "services")
	assert.Contains(t, mergedResult.Reasons, "exact match")
	assert.Contains(t, mergedResult.Reasons, "keyword match")

	// Match type should be updated
	assert.Equal(t, "multi-strategy", mergedResult.MatchType)
}

func TestResultAggregator_EmptyResults(t *testing.T) {
	aggregator, cleanup := setupTestResultAggregator(t)
	defer cleanup()

	request := &AggregationRequest{
		Results:           []*ClassificationResult{},
		MaxResultsPerType: 3,
		MinConfidence:     0.3,
		SortBy:            SortByConfidence,
		Presentation:      PresentationSummary,
	}

	results, err := aggregator.AggregateAndPresent(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, results)

	// Should return empty but valid results structure
	assert.Empty(t, results.AllResults)
	assert.Empty(t, results.TopThreeByType)
	assert.Empty(t, results.OverallTopResults)
	assert.NotNil(t, results.AggregationMetadata)
	assert.Equal(t, 0, results.AggregationMetadata.TotalInputResults)
}

func TestResultAggregator_ProcessingSteps(t *testing.T) {
	aggregator, cleanup := setupTestResultAggregator(t)
	defer cleanup()

	results := createTestResultsForAggregation()

	request := &AggregationRequest{
		Results:           results,
		MaxResultsPerType: 3,
		MinConfidence:     0.3,
		IncludeMetadata:   true,
		SortBy:            SortByConfidence,
		Presentation:      PresentationSummary,
	}

	aggregatedResults, err := aggregator.AggregateAndPresent(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, aggregatedResults)

	metadata := aggregatedResults.AggregationMetadata
	require.NotNil(t, metadata)
	require.NotEmpty(t, metadata.ProcessingSteps)

	// Check expected processing steps
	expectedSteps := []string{"deduplication", "score_calculation", "filtering", "sorting", "type_grouping"}

	assert.GreaterOrEqual(t, len(metadata.ProcessingSteps), len(expectedSteps))

	for i, expectedStep := range expectedSteps {
		if i < len(metadata.ProcessingSteps) {
			assert.Equal(t, expectedStep, metadata.ProcessingSteps[i].Step)
			assert.True(t, metadata.ProcessingSteps[i].Success)
			assert.Greater(t, metadata.ProcessingSteps[i].Duration, time.Duration(0))
		}
	}

	// Check total aggregation time
	assert.Greater(t, metadata.AggregationTime, time.Duration(0))
}

func TestResultAggregator_UIHints(t *testing.T) {
	aggregator, cleanup := setupTestResultAggregator(t)
	defer cleanup()

	tests := []struct {
		name       string
		confidence float64
		matchType  string
		checkHints func(t *testing.T, hints map[string]interface{})
	}{
		{
			name:       "high confidence exact match",
			confidence: 0.95,
			matchType:  "exact",
			checkHints: func(t *testing.T, hints map[string]interface{}) {
				assert.Equal(t, "green", hints["confidence_color"])
				assert.Equal(t, "check-circle", hints["confidence_icon"])
				assert.Equal(t, "high", hints["priority"])
				assert.True(t, hints["featured"].(bool))
			},
		},
		{
			name:       "medium confidence keyword match",
			confidence: 0.65,
			matchType:  "keyword",
			checkHints: func(t *testing.T, hints map[string]interface{}) {
				assert.Equal(t, "orange", hints["confidence_color"])
				assert.Equal(t, "warning", hints["confidence_icon"])
				assert.Equal(t, "medium", hints["priority"])
			},
		},
		{
			name:       "low confidence fuzzy match",
			confidence: 0.25,
			matchType:  "fuzzy",
			checkHints: func(t *testing.T, hints map[string]interface{}) {
				assert.Equal(t, "red", hints["confidence_color"])
				assert.Equal(t, "alert-triangle", hints["confidence_icon"])
				assert.Equal(t, "low", hints["priority"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ClassificationResult{
				Code: &IndustryCode{
					Code:        "TEST",
					Type:        CodeTypeSIC,
					Description: "Test Code",
				},
				Confidence: tt.confidence,
				MatchType:  tt.matchType,
				Reasons:    []string{"test reason"},
				MatchedOn:  []string{"test"},
			}

			hints := aggregator.generateUIHints(result)
			require.NotNil(t, hints)

			tt.checkHints(t, hints)
		})
	}
}
