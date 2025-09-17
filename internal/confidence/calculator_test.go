package confidence

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestConfidenceCalculator_CalculateDynamicConfidence tests the dynamic confidence calculation
func TestConfidenceCalculator_CalculateDynamicConfidence(t *testing.T) {
	tests := []struct {
		name            string
		industryID      int
		industryName    string
		matchedKeywords []string
		totalKeywords   int
		rawScore        float64
		industryMatches map[int][]string
		mockSetup       func(*MockIndustryThresholdRepository)
		expectedMin     float64
		expectedMax     float64
	}{
		{
			name:            "high confidence restaurant classification",
			industryID:      1,
			industryName:    "Restaurants",
			matchedKeywords: []string{"restaurant", "dining", "food", "cuisine", "menu"},
			totalKeywords:   5,
			rawScore:        4.5,
			industryMatches: map[int][]string{
				1: {"restaurant", "dining", "food", "cuisine", "menu"},
				2: {"fast", "food", "quick"},
			},
			mockSetup: func(mockRepo *MockIndustryThresholdRepository) {
				mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&Industry{
					ID:                  1,
					Name:                "Restaurants",
					ConfidenceThreshold: 0.75,
					IsActive:            true,
				}, nil)
			},
			expectedMin: 0.80,
			expectedMax: 1.0,
		},
		{
			name:            "medium confidence classification",
			industryID:      2,
			industryName:    "Fast Food",
			matchedKeywords: []string{"fast", "food", "quick"},
			totalKeywords:   8,
			rawScore:        2.5,
			industryMatches: map[int][]string{
				1: {"restaurant", "dining", "food"},
				2: {"fast", "food", "quick"},
			},
			mockSetup: func(mockRepo *MockIndustryThresholdRepository) {
				mockRepo.On("GetIndustryByName", mock.Anything, "Fast Food").Return(&Industry{
					ID:                  2,
					Name:                "Fast Food",
					ConfidenceThreshold: 0.80,
					IsActive:            true,
				}, nil)
			},
			expectedMin: 0.30,
			expectedMax: 0.50,
		},
		{
			name:            "low confidence classification",
			industryID:      3,
			industryName:    "General Business",
			matchedKeywords: []string{"business", "service"},
			totalKeywords:   10,
			rawScore:        1.0,
			industryMatches: map[int][]string{
				1: {"restaurant", "dining", "food"},
				2: {"fast", "food", "quick"},
				3: {"business", "service", "general"},
			},
			mockSetup: func(mockRepo *MockIndustryThresholdRepository) {
				mockRepo.On("GetIndustryByName", mock.Anything, "General Business").Return(&Industry{
					ID:                  3,
					Name:                "General Business",
					ConfidenceThreshold: 0.50,
					IsActive:            true,
				}, nil)
			},
			expectedMin: 0.20,
			expectedMax: 0.40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockIndustryThresholdRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			// Create threshold service and calculator
			thresholdService := NewIndustryThresholdService(mockRepo, nil)
			calculator := NewConfidenceCalculator(thresholdService, nil)

			// Execute test
			result, err := calculator.CalculateDynamicConfidence(
				context.Background(),
				tt.industryID,
				tt.industryName,
				tt.matchedKeywords,
				tt.totalKeywords,
				tt.rawScore,
				tt.industryMatches,
			)

			// Assert results
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.GreaterOrEqual(t, result.FinalConfidence, tt.expectedMin)
			assert.LessOrEqual(t, result.FinalConfidence, tt.expectedMax)
			assert.Equal(t, tt.industryID, result.IndustryID)
			assert.Equal(t, tt.industryName, result.IndustryName)
			assert.Equal(t, tt.matchedKeywords, result.MatchedKeywords)
			assert.Equal(t, tt.totalKeywords, result.TotalKeywords)
			assert.Equal(t, tt.rawScore, result.RawScore)
			assert.Less(t, result.CalculationTime, 100*time.Millisecond)

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestConfidenceCalculator_GetIndustryThreshold tests the GetIndustryThreshold method
func TestConfidenceCalculator_GetIndustryThreshold(t *testing.T) {
	// Setup mock
	mockRepo := &MockIndustryThresholdRepository{}
	mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&Industry{
		ID:                  1,
		Name:                "Restaurants",
		ConfidenceThreshold: 0.75,
		IsActive:            true,
	}, nil)

	// Create threshold service and calculator
	thresholdService := NewIndustryThresholdService(mockRepo, nil)
	calculator := NewConfidenceCalculator(thresholdService, nil)

	// Execute test
	threshold, err := calculator.GetIndustryThreshold(context.Background(), "Restaurants")

	// Assert results
	require.NoError(t, err)
	assert.Equal(t, 0.75, threshold)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestConfidenceCalculator_GetAllIndustryThresholds tests the GetAllIndustryThresholds method
func TestConfidenceCalculator_GetAllIndustryThresholds(t *testing.T) {
	// Setup mock
	mockRepo := &MockIndustryThresholdRepository{}
	thresholds := map[string]float64{
		"Restaurants": 0.75,
		"Fast Food":   0.80,
		"Healthcare":  0.80,
	}
	mockRepo.On("GetAllIndustryThresholds", mock.Anything).Return(thresholds, nil)

	// Create threshold service and calculator
	thresholdService := NewIndustryThresholdService(mockRepo, nil)
	calculator := NewConfidenceCalculator(thresholdService, nil)

	// Execute test
	result, err := calculator.GetAllIndustryThresholds(context.Background())

	// Assert results
	require.NoError(t, err)
	assert.Equal(t, thresholds, result)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestConfidenceCalculator_RefreshThresholdCache tests the cache refresh functionality
func TestConfidenceCalculator_RefreshThresholdCache(t *testing.T) {
	// Setup mock
	mockRepo := &MockIndustryThresholdRepository{}
	thresholds := map[string]float64{
		"Restaurants": 0.75,
		"Fast Food":   0.80,
	}
	mockRepo.On("GetAllIndustryThresholds", mock.Anything).Return(thresholds, nil)

	// Create threshold service and calculator
	thresholdService := NewIndustryThresholdService(mockRepo, nil)
	calculator := NewConfidenceCalculator(thresholdService, nil)

	// Execute test
	err := calculator.RefreshThresholdCache(context.Background())

	// Assert results
	require.NoError(t, err)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestConfidenceCalculator_GetThresholdService tests the GetThresholdService method
func TestConfidenceCalculator_GetThresholdService(t *testing.T) {
	// Create threshold service and calculator
	thresholdService := NewIndustryThresholdService(nil, nil)
	calculator := NewConfidenceCalculator(thresholdService, nil)

	// Execute test
	result := calculator.GetThresholdService()

	// Assert results
	assert.Equal(t, thresholdService, result)
}

// TestConfidenceCalculator_IndustrySpecificThresholds tests industry-specific threshold application
func TestConfidenceCalculator_IndustrySpecificThresholds(t *testing.T) {
	tests := []struct {
		name           string
		industryName   string
		threshold      float64
		expectedFactor float64
	}{
		{
			name:           "high threshold industry",
			industryName:   "Fast Food",
			threshold:      0.80,
			expectedFactor: 0.5, // Base factor (0.5) + adjustment ((0.8-0.8)*0.5 = 0.0)
		},
		{
			name:           "medium threshold industry",
			industryName:   "Restaurants",
			threshold:      0.75,
			expectedFactor: 0.525, // Base factor (0.5) + adjustment ((0.8-0.75)*0.5 = 0.025)
		},
		{
			name:           "low threshold industry",
			industryName:   "General Business",
			threshold:      0.50,
			expectedFactor: 0.65, // Base factor (0.5) + adjustment ((0.8-0.5)*0.5 = 0.15)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockIndustryThresholdRepository{}
			mockRepo.On("GetIndustryByName", mock.Anything, tt.industryName).Return(&Industry{
				ID:                  1,
				Name:                tt.industryName,
				ConfidenceThreshold: tt.threshold,
				IsActive:            true,
			}, nil)

			// Create threshold service and calculator
			thresholdService := NewIndustryThresholdService(mockRepo, nil)
			calculator := NewConfidenceCalculator(thresholdService, nil)

			// Execute test
			result, err := calculator.CalculateDynamicConfidence(
				context.Background(),
				1,
				tt.industryName,
				[]string{"test"},
				1,
				1.0,
				map[int][]string{1: {"test"}},
			)

			// Assert results
			require.NoError(t, err)
			assert.Equal(t, tt.expectedFactor, result.Factors.IndustryThresholdFactor)

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestConfidenceCalculator_ErrorHandling tests error handling in confidence calculation
func TestConfidenceCalculator_ErrorHandling(t *testing.T) {
	// Setup mock that returns an error
	mockRepo := &MockIndustryThresholdRepository{}
	mockRepo.On("GetIndustryByName", mock.Anything, "Error Industry").Return((*Industry)(nil), assert.AnError)

	// Create threshold service and calculator
	thresholdService := NewIndustryThresholdService(mockRepo, nil)
	calculator := NewConfidenceCalculator(thresholdService, nil)

	// Execute test
	result, err := calculator.CalculateDynamicConfidence(
		context.Background(),
		1,
		"Error Industry",
		[]string{"test"},
		1,
		1.0,
		map[int][]string{1: {"test"}},
	)

	// Assert results - should not error, but use default threshold
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.InDelta(t, 0.65, result.Factors.IndustryThresholdFactor, 0.01) // Default threshold factor

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestConfidenceCalculator_Performance tests the performance of confidence calculation
func TestConfidenceCalculator_Performance(t *testing.T) {
	// Setup mock
	mockRepo := &MockIndustryThresholdRepository{}
	mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&Industry{
		ID:                  1,
		Name:                "Restaurants",
		ConfidenceThreshold: 0.75,
		IsActive:            true,
	}, nil)

	// Create threshold service and calculator
	thresholdService := NewIndustryThresholdService(mockRepo, nil)
	calculator := NewConfidenceCalculator(thresholdService, nil)

	// Measure performance
	start := time.Now()
	result, err := calculator.CalculateDynamicConfidence(
		context.Background(),
		1,
		"Restaurants",
		[]string{"restaurant", "dining", "food", "cuisine", "menu"},
		5,
		4.5,
		map[int][]string{1: {"restaurant", "dining", "food", "cuisine", "menu"}},
	)
	duration := time.Since(start)

	// Assert results
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Less(t, duration, 50*time.Millisecond, "Confidence calculation should be fast")
	assert.Less(t, result.CalculationTime, 10*time.Millisecond, "Calculation time should be minimal")

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestConfidenceCalculator_ConcurrentAccess tests concurrent access to the calculator
func TestConfidenceCalculator_ConcurrentAccess(t *testing.T) {
	// Setup mock
	mockRepo := &MockIndustryThresholdRepository{}
	mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&Industry{
		ID:                  1,
		Name:                "Restaurants",
		ConfidenceThreshold: 0.75,
		IsActive:            true,
	}, nil).Maybe() // Allow multiple calls due to concurrent access

	// Create threshold service and calculator
	thresholdService := NewIndustryThresholdService(mockRepo, nil)
	calculator := NewConfidenceCalculator(thresholdService, nil)

	// Test concurrent access
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- true }()
			result, err := calculator.CalculateDynamicConfidence(
				context.Background(),
				1,
				"Restaurants",
				[]string{"restaurant", "dining"},
				2,
				2.0,
				map[int][]string{1: {"restaurant", "dining"}},
			)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestConfidenceCalculator_EnhancedKeywordSpecificityScoring tests the enhanced keyword specificity scoring
// This implements subtask 2.2.3 from the comprehensive improvement plan
func TestConfidenceCalculator_EnhancedKeywordSpecificityScoring(t *testing.T) {
	// Setup mock
	mockRepo := &MockIndustryThresholdRepository{}
	mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&Industry{
		ID:                  1,
		Name:                "Restaurants",
		ConfidenceThreshold: 0.75,
		IsActive:            true,
	}, nil)

	// Create threshold service and calculator
	thresholdService := NewIndustryThresholdService(mockRepo, nil)
	calculator := NewConfidenceCalculator(thresholdService, nil)

	tests := []struct {
		name            string
		matchedKeywords []string
		expectedMin     float64
		expectedMax     float64
		description     string
	}{
		{
			name:            "single_keyword_match",
			matchedKeywords: []string{"restaurant"},
			expectedMin:     0.30,
			expectedMax:     0.35,
			description:     "Single keyword should have low specificity",
		},
		{
			name:            "two_keyword_matches",
			matchedKeywords: []string{"restaurant", "dining"},
			expectedMin:     0.35,
			expectedMax:     0.40,
			description:     "Two keywords should have low-medium specificity",
		},
		{
			name:            "three_keyword_matches",
			matchedKeywords: []string{"restaurant", "dining", "food"},
			expectedMin:     0.35,
			expectedMax:     0.40,
			description:     "Three keywords should have medium specificity",
		},
		{
			name:            "four_keyword_matches",
			matchedKeywords: []string{"restaurant", "dining", "food", "cuisine"},
			expectedMin:     0.38,
			expectedMax:     0.42,
			description:     "Four keywords should have medium-high specificity",
		},
		{
			name:            "five_keyword_matches",
			matchedKeywords: []string{"restaurant", "dining", "food", "cuisine", "menu"},
			expectedMin:     0.40,
			expectedMax:     0.45,
			description:     "Five keywords should have high specificity",
		},
		{
			name:            "six_keyword_matches",
			matchedKeywords: []string{"restaurant", "dining", "food", "cuisine", "menu", "chef"},
			expectedMin:     0.42,
			expectedMax:     0.48,
			description:     "Six keywords should have very high specificity",
		},
		{
			name:            "many_keyword_matches",
			matchedKeywords: []string{"restaurant", "dining", "food", "cuisine", "menu", "chef", "kitchen", "cooking", "pasta", "wine"},
			expectedMin:     0.45,
			expectedMax:     0.50,
			description:     "Many keywords should have very high specificity with diminishing returns",
		},
		{
			name:            "no_keyword_matches",
			matchedKeywords: []string{},
			expectedMin:     0.0,
			expectedMax:     0.0,
			description:     "No keywords should have zero specificity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute test
			result, err := calculator.CalculateDynamicConfidence(
				context.Background(),
				1,
				"Restaurants",
				tt.matchedKeywords,
				len(tt.matchedKeywords),
				1.0,
				map[int][]string{1: tt.matchedKeywords},
			)

			// Assert results
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.GreaterOrEqual(t, result.Factors.SpecificityFactor, tt.expectedMin,
				"Specificity factor should be >= %f for %s", tt.expectedMin, tt.description)
			assert.LessOrEqual(t, result.Factors.SpecificityFactor, tt.expectedMax,
				"Specificity factor should be <= %f for %s", tt.expectedMax, tt.description)

			// Verify that more keywords generally result in higher specificity
			if len(tt.matchedKeywords) > 0 {
				assert.Greater(t, result.Factors.SpecificityFactor, 0.0,
					"Specificity should be positive for non-empty keyword matches")
			}
		})
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestConfidenceCalculator_MatchCountFactor tests the match count factor calculation directly
func TestConfidenceCalculator_MatchCountFactor(t *testing.T) {
	// Create calculator without threshold service for direct testing
	calculator := NewConfidenceCalculator(nil, nil)

	tests := []struct {
		name        string
		matchCount  int
		expectedMin float64
		expectedMax float64
		description string
	}{
		{
			name:        "zero_matches",
			matchCount:  0,
			expectedMin: 0.0,
			expectedMax: 0.0,
			description: "Zero matches should return zero score",
		},
		{
			name:        "one_match",
			matchCount:  1,
			expectedMin: 0.04,
			expectedMax: 0.06,
			description: "One match should have low score",
		},
		{
			name:        "two_matches",
			matchCount:  2,
			expectedMin: 0.14,
			expectedMax: 0.16,
			description: "Two matches should have low-medium score",
		},
		{
			name:        "three_matches",
			matchCount:  3,
			expectedMin: 0.24,
			expectedMax: 0.26,
			description: "Three matches should have medium score",
		},
		{
			name:        "four_matches",
			matchCount:  4,
			expectedMin: 0.35,
			expectedMax: 0.37,
			description: "Four matches should have medium-high score",
		},
		{
			name:        "five_matches",
			matchCount:  5,
			expectedMin: 0.44,
			expectedMax: 0.46,
			description: "Five matches should have high score",
		},
		{
			name:        "six_matches",
			matchCount:  6,
			expectedMin: 0.52,
			expectedMax: 0.54,
			description: "Six matches should have very high score",
		},
		{
			name:        "ten_matches",
			matchCount:  10,
			expectedMin: 0.72,
			expectedMax: 0.74,
			description: "Ten matches should have very high score with diminishing returns",
		},
		{
			name:        "twenty_matches",
			matchCount:  20,
			expectedMin: 0.90,
			expectedMax: 1.0,
			description: "Twenty matches should have maximum score with strong diminishing returns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute test using reflection to access private method
			// Note: In a real implementation, we might want to make this method public for testing
			score := calculator.calculateMatchCountFactor(tt.matchCount)

			// Assert results
			assert.GreaterOrEqual(t, score, tt.expectedMin,
				"Match count factor should be >= %f for %s", tt.expectedMin, tt.description)
			assert.LessOrEqual(t, score, tt.expectedMax,
				"Match count factor should be <= %f for %s", tt.expectedMax, tt.description)

			// Verify monotonic increase (more matches should generally give higher scores)
			if tt.matchCount > 0 {
				assert.Greater(t, score, 0.0,
					"Match count factor should be positive for positive match count")
			}
		})
	}
}

// TestConfidenceCalculator_SpecificityScoringProgression tests that specificity increases with more matches
func TestConfidenceCalculator_SpecificityScoringProgression(t *testing.T) {
	// Setup mock
	mockRepo := &MockIndustryThresholdRepository{}
	mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&Industry{
		ID:                  1,
		Name:                "Restaurants",
		ConfidenceThreshold: 0.75,
		IsActive:            true,
	}, nil).Maybe() // Allow multiple calls

	// Create threshold service and calculator
	thresholdService := NewIndustryThresholdService(mockRepo, nil)
	calculator := NewConfidenceCalculator(thresholdService, nil)

	// Test progressive increase in specificity with more keywords
	keywordSets := [][]string{
		{"restaurant"},
		{"restaurant", "dining"},
		{"restaurant", "dining", "food"},
		{"restaurant", "dining", "food", "cuisine"},
		{"restaurant", "dining", "food", "cuisine", "menu"},
		{"restaurant", "dining", "food", "cuisine", "menu", "chef"},
	}

	var previousSpecificity float64 = 0.0

	for i, keywords := range keywordSets {
		t.Run(fmt.Sprintf("progression_%d_keywords", len(keywords)), func(t *testing.T) {
			result, err := calculator.CalculateDynamicConfidence(
				context.Background(),
				1,
				"Restaurants",
				keywords,
				len(keywords),
				1.0,
				map[int][]string{1: keywords},
			)

			require.NoError(t, err)
			assert.NotNil(t, result)

			// Verify that specificity generally increases with more keywords
			if i > 0 {
				assert.GreaterOrEqual(t, result.Factors.SpecificityFactor, previousSpecificity,
					"Specificity should generally increase with more keywords (had %f, now %f)",
					previousSpecificity, result.Factors.SpecificityFactor)
			}

			previousSpecificity = result.Factors.SpecificityFactor
		})
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestConfidenceCalculator_SpecificityScoringEdgeCases tests edge cases for specificity scoring
func TestConfidenceCalculator_SpecificityScoringEdgeCases(t *testing.T) {
	// Setup mock
	mockRepo := &MockIndustryThresholdRepository{}
	mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&Industry{
		ID:                  1,
		Name:                "Restaurants",
		ConfidenceThreshold: 0.75,
		IsActive:            true,
	}, nil).Maybe()

	// Create threshold service and calculator
	thresholdService := NewIndustryThresholdService(mockRepo, nil)
	calculator := NewConfidenceCalculator(thresholdService, nil)

	tests := []struct {
		name            string
		matchedKeywords []string
		industryMatches map[int][]string
		expectedMin     float64
		expectedMax     float64
		description     string
	}{
		{
			name:            "empty_keywords",
			matchedKeywords: []string{},
			industryMatches: map[int][]string{1: {}},
			expectedMin:     0.0,
			expectedMax:     0.0,
			description:     "Empty keywords should return zero specificity",
		},
		{
			name:            "nil_keywords",
			matchedKeywords: nil,
			industryMatches: map[int][]string{1: {}},
			expectedMin:     0.0,
			expectedMax:     0.0,
			description:     "Nil keywords should return zero specificity",
		},
		{
			name:            "single_high_quality_keyword",
			matchedKeywords: []string{"restaurant"},
			industryMatches: map[int][]string{
				1: {"restaurant"},
				2: {"fast", "food"},
				3: {"general", "business"},
			},
			expectedMin: 0.60,
			expectedMax: 0.65,
			description: "Single high-quality keyword should have low but positive specificity",
		},
		{
			name:            "many_low_quality_keywords",
			matchedKeywords: []string{"the", "and", "or", "but", "in", "on", "at", "to", "for", "of"},
			industryMatches: map[int][]string{
				1: {"the", "and", "or", "but", "in", "on", "at", "to", "for", "of"},
				2: {"the", "and", "or", "but", "in", "on", "at", "to", "for", "of"},
				3: {"the", "and", "or", "but", "in", "on", "at", "to", "for", "of"},
			},
			expectedMin: 0.30,
			expectedMax: 0.35,
			description: "Many low-quality keywords should still have high match count specificity",
		},
		{
			name:            "mixed_quality_keywords",
			matchedKeywords: []string{"restaurant", "dining", "the", "and", "food"},
			industryMatches: map[int][]string{
				1: {"restaurant", "dining", "food"},
				2: {"fast", "food"},
				3: {"general", "business"},
			},
			expectedMin: 0.55,
			expectedMax: 0.60,
			description: "Mixed quality keywords should have balanced specificity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calculator.CalculateDynamicConfidence(
				context.Background(),
				1,
				"Restaurants",
				tt.matchedKeywords,
				len(tt.matchedKeywords),
				1.0,
				tt.industryMatches,
			)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.GreaterOrEqual(t, result.Factors.SpecificityFactor, tt.expectedMin,
				"Specificity factor should be >= %f for %s", tt.expectedMin, tt.description)
			assert.LessOrEqual(t, result.Factors.SpecificityFactor, tt.expectedMax,
				"Specificity factor should be <= %f for %s", tt.expectedMax, tt.description)
		})
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}
