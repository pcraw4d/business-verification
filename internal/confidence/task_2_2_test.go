package confidence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestTask2_2_DynamicConfidenceCalculation tests the dynamic confidence calculation as specified in the plan
// This implements the specific test requirements from Task 2.2
func TestTask2_2_DynamicConfidenceCalculation(t *testing.T) {
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
		totalKeywords   int
		rawScore        float64
		expectedMin     float64
		expectedMax     float64
		description     string
	}{
		{
			name:            "high_match_scenario",
			matchedKeywords: []string{"restaurant", "dining", "food", "cuisine", "menu", "chef", "kitchen", "cooking"},
			totalKeywords:   10,
			rawScore:        8.0,
			expectedMin:     0.80,
			expectedMax:     1.0,
			description:     "High match scenario should have high confidence",
		},
		{
			name:            "low_match_scenario",
			matchedKeywords: []string{"restaurant", "food"},
			totalKeywords:   10,
			rawScore:        2.0,
			expectedMin:     0.20,
			expectedMax:     0.50,
			description:     "Low match scenario should have low confidence",
		},
		{
			name:            "medium_match_scenario",
			matchedKeywords: []string{"restaurant", "dining", "food", "cuisine", "menu"},
			totalKeywords:   10,
			rawScore:        5.0,
			expectedMin:     0.50,
			expectedMax:     0.70,
			description:     "Medium match scenario should have medium confidence",
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
				tt.totalKeywords,
				tt.rawScore,
				map[int][]string{1: tt.matchedKeywords},
			)

			// Assert results
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.GreaterOrEqual(t, result.FinalConfidence, tt.expectedMin,
				"Confidence should be >= %f for %s", tt.expectedMin, tt.description)
			assert.LessOrEqual(t, result.FinalConfidence, tt.expectedMax,
				"Confidence should be <= %f for %s", tt.expectedMax, tt.description)

			// Verify confidence varies based on match quality (not fixed 0.45)
			assert.NotEqual(t, 0.45, result.FinalConfidence,
				"Confidence should not be fixed at 0.45, should vary based on match quality")

			// Verify confidence reflects keyword match strength
			assert.Greater(t, result.FinalConfidence, 0.0,
				"Confidence should be positive for any match")
			assert.LessOrEqual(t, result.FinalConfidence, 1.0,
				"Confidence should not exceed 1.0")

			// Verify calculation time < 10ms as specified in plan
			assert.Less(t, result.CalculationTime, 10*time.Millisecond,
				"Confidence calculation should be < 10ms as specified in plan")
		})
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestTask2_2_IndustrySpecificThresholds tests industry-specific threshold application as specified in the plan
func TestTask2_2_IndustrySpecificThresholds(t *testing.T) {
	tests := []struct {
		name         string
		industryName string
		threshold    float64
		expectedMin  float64
		expectedMax  float64
		description  string
	}{
		{
			name:         "restaurant_industry",
			industryName: "Restaurants",
			threshold:    0.75,
			expectedMin:  0.60,
			expectedMax:  0.70,
			description:  "Restaurant industry should have confidence in expected range",
		},
		{
			name:         "fast_food_industry",
			industryName: "Fast Food",
			threshold:    0.80,
			expectedMin:  0.60,
			expectedMax:  0.70,
			description:  "Fast Food industry should have confidence in expected range",
		},
		{
			name:         "general_business_industry",
			industryName: "General Business",
			threshold:    0.50,
			expectedMin:  0.40,
			expectedMax:  0.70,
			description:  "General Business industry should have confidence < 0.70",
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

			// Execute test with same input data
			result, err := calculator.CalculateDynamicConfidence(
				context.Background(),
				1,
				tt.industryName,
				[]string{"restaurant", "dining", "food", "cuisine", "menu"},
				10,
				6.0,
				map[int][]string{1: {"restaurant", "dining", "food", "cuisine", "menu"}},
			)

			// Assert results
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.GreaterOrEqual(t, result.FinalConfidence, tt.expectedMin,
				"Confidence should be >= %f for %s", tt.expectedMin, tt.description)
			assert.LessOrEqual(t, result.FinalConfidence, tt.expectedMax,
				"Confidence should be <= %f for %s", tt.expectedMax, tt.description)

			// Verify industry-specific threshold factor is applied (factor is calculated, not the raw threshold)
			assert.Greater(t, result.Factors.IndustryThresholdFactor, 0.0,
				"Industry threshold factor should be positive")
			assert.LessOrEqual(t, result.Factors.IndustryThresholdFactor, 1.0,
				"Industry threshold factor should not exceed 1.0")

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestTask2_2_KeywordSpecificity tests keyword specificity as specified in the plan
func TestTask2_2_KeywordSpecificity(t *testing.T) {
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
		expectedMin     float64
		expectedMax     float64
		description     string
	}{
		{
			name:            "high_specificity_many_matches",
			matchedKeywords: []string{"restaurant", "dining", "food", "cuisine", "menu", "chef", "kitchen", "cooking"},
			expectedMin:     0.40,
			expectedMax:     0.50,
			description:     "High specificity with many matches should have high specificity factor",
		},
		{
			name:            "low_specificity_few_matches",
			matchedKeywords: []string{"restaurant"},
			expectedMin:     0.30,
			expectedMax:     0.35,
			description:     "Low specificity with few matches should have low specificity factor",
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

			// Verify specificity factor is properly calculated
			assert.Greater(t, result.Factors.SpecificityFactor, 0.0,
				"Specificity factor should be positive for any matches")
			assert.LessOrEqual(t, result.Factors.SpecificityFactor, 1.0,
				"Specificity factor should not exceed 1.0")
		})
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestTask2_2_ConfidenceScoreVariation tests that confidence scores vary based on match quality
func TestTask2_2_ConfidenceScoreVariation(t *testing.T) {
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

	// Test different match scenarios to ensure confidence varies
	scenarios := []struct {
		name            string
		matchedKeywords []string
		totalKeywords   int
		rawScore        float64
		expectedRange   string
	}{
		{
			name:            "excellent_match",
			matchedKeywords: []string{"restaurant", "dining", "food", "cuisine", "menu", "chef", "kitchen", "cooking", "pasta", "wine"},
			totalKeywords:   10,
			rawScore:        9.5,
			expectedRange:   "0.80-0.90",
		},
		{
			name:            "good_match",
			matchedKeywords: []string{"restaurant", "dining", "food", "cuisine", "menu"},
			totalKeywords:   10,
			rawScore:        5.0,
			expectedRange:   "0.50-0.70",
		},
		{
			name:            "fair_match",
			matchedKeywords: []string{"restaurant", "food"},
			totalKeywords:   10,
			rawScore:        2.0,
			expectedRange:   "0.30-0.50",
		},
		{
			name:            "poor_match",
			matchedKeywords: []string{"restaurant"},
			totalKeywords:   10,
			rawScore:        1.0,
			expectedRange:   "0.20-0.40",
		},
	}

	var previousConfidence float64 = 0.0

	for i, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Execute test
			result, err := calculator.CalculateDynamicConfidence(
				context.Background(),
				1,
				"Restaurants",
				scenario.matchedKeywords,
				scenario.totalKeywords,
				scenario.rawScore,
				map[int][]string{1: scenario.matchedKeywords},
			)

			// Assert results
			require.NoError(t, err)
			assert.NotNil(t, result)

			// Verify confidence is in expected range
			switch scenario.expectedRange {
			case "0.80-0.90":
				assert.GreaterOrEqual(t, result.FinalConfidence, 0.80)
				assert.LessOrEqual(t, result.FinalConfidence, 0.90)
			case "0.50-0.70":
				assert.GreaterOrEqual(t, result.FinalConfidence, 0.50)
				assert.LessOrEqual(t, result.FinalConfidence, 0.70)
			case "0.30-0.50":
				assert.GreaterOrEqual(t, result.FinalConfidence, 0.30)
				assert.LessOrEqual(t, result.FinalConfidence, 0.50)
			case "0.20-0.40":
				assert.GreaterOrEqual(t, result.FinalConfidence, 0.20)
				assert.LessOrEqual(t, result.FinalConfidence, 0.40)
			}

			// Verify confidence generally decreases with worse matches
			if i > 0 {
				assert.LessOrEqual(t, result.FinalConfidence, previousConfidence,
					"Confidence should generally decrease with worse matches")
			}

			previousConfidence = result.FinalConfidence
		})
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestTask2_2_NoFixedConfidenceScores tests that we no longer have fixed 0.45 confidence scores
func TestTask2_2_NoFixedConfidenceScores(t *testing.T) {
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

	// Test multiple scenarios to ensure no fixed confidence scores
	scenarios := []struct {
		name            string
		matchedKeywords []string
		totalKeywords   int
		rawScore        float64
	}{
		{
			name:            "scenario_1",
			matchedKeywords: []string{"restaurant"},
			totalKeywords:   10,
			rawScore:        1.0,
		},
		{
			name:            "scenario_2",
			matchedKeywords: []string{"restaurant", "dining"},
			totalKeywords:   10,
			rawScore:        2.0,
		},
		{
			name:            "scenario_3",
			matchedKeywords: []string{"restaurant", "dining", "food"},
			totalKeywords:   10,
			rawScore:        3.0,
		},
		{
			name:            "scenario_4",
			matchedKeywords: []string{"restaurant", "dining", "food", "cuisine"},
			totalKeywords:   10,
			rawScore:        4.0,
		},
		{
			name:            "scenario_5",
			matchedKeywords: []string{"restaurant", "dining", "food", "cuisine", "menu"},
			totalKeywords:   10,
			rawScore:        5.0,
		},
	}

	confidences := make([]float64, len(scenarios))

	for i, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Execute test
			result, err := calculator.CalculateDynamicConfidence(
				context.Background(),
				1,
				"Restaurants",
				scenario.matchedKeywords,
				scenario.totalKeywords,
				scenario.rawScore,
				map[int][]string{1: scenario.matchedKeywords},
			)

			// Assert results
			require.NoError(t, err)
			assert.NotNil(t, result)

			// Verify confidence is not fixed at 0.45
			assert.NotEqual(t, 0.45, result.FinalConfidence,
				"Confidence should not be fixed at 0.45, should vary based on match quality")

			// Store confidence for later analysis
			confidences[i] = result.FinalConfidence
		})
	}

	// Verify that we have variation in confidence scores
	uniqueConfidences := make(map[float64]bool)
	for _, conf := range confidences {
		uniqueConfidences[conf] = true
	}

	assert.Greater(t, len(uniqueConfidences), 1,
		"Should have variation in confidence scores, not all the same value")

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestTask2_2_ConfidenceCalculationPerformance tests that confidence calculation time < 10ms
func TestTask2_2_ConfidenceCalculationPerformance(t *testing.T) {
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

	// Test performance with various scenarios
	scenarios := []struct {
		name            string
		matchedKeywords []string
		totalKeywords   int
		rawScore        float64
	}{
		{
			name:            "small_dataset",
			matchedKeywords: []string{"restaurant"},
			totalKeywords:   1,
			rawScore:        1.0,
		},
		{
			name:            "medium_dataset",
			matchedKeywords: []string{"restaurant", "dining", "food", "cuisine", "menu"},
			totalKeywords:   5,
			rawScore:        5.0,
		},
		{
			name:            "large_dataset",
			matchedKeywords: []string{"restaurant", "dining", "food", "cuisine", "menu", "chef", "kitchen", "cooking", "pasta", "wine", "service", "quality"},
			totalKeywords:   12,
			rawScore:        12.0,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Measure performance
			start := time.Now()
			result, err := calculator.CalculateDynamicConfidence(
				context.Background(),
				1,
				"Restaurants",
				scenario.matchedKeywords,
				scenario.totalKeywords,
				scenario.rawScore,
				map[int][]string{1: scenario.matchedKeywords},
			)
			duration := time.Since(start)

			// Assert results
			require.NoError(t, err)
			assert.NotNil(t, result)

			// Verify performance requirement: < 10ms as specified in plan
			assert.Less(t, duration, 10*time.Millisecond,
				"Confidence calculation should be < 10ms as specified in plan, got %v", duration)

			// Verify the reported calculation time is also < 10ms
			assert.Less(t, result.CalculationTime, 10*time.Millisecond,
				"Reported calculation time should be < 10ms, got %v", result.CalculationTime)
		})
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}
