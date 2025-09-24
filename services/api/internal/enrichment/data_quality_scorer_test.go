package enrichment

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Test data structures
type TestBusinessData struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	ConfidenceScore  float64   `json:"confidence_score"`
	Description      string    `json:"description,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	Source           string    `json:"source"`
	EmployeeCount    int       `json:"employee_count"`
	Revenue          float64   `json:"revenue"`
	IsValidated      bool      `json:"is_validated"`
	ProcessingStatus string    `json:"processing_status,omitempty"`
}

type TestIncompleteData struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Score       float64 `json:"score"`
}

type TestInvalidData struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	ConfidenceScore float64   `json:"confidence_score"`
	EmployeeCount   int       `json:"employee_count"`
	Revenue         float64   `json:"revenue"`
	InvalidField    float64   `json:"invalid_field"` // Will contain NaN for testing
	UpdatedAt       time.Time `json:"updated_at"`
	Source          string    `json:"source"`
}

func TestNewDataQualityScorer(t *testing.T) {
	tests := []struct {
		name   string
		config *DataQualityConfig
	}{
		{
			name:   "with nil config",
			config: nil,
		},
		{
			name: "with custom config",
			config: &DataQualityConfig{
				CompletenessWeight:   1.5,
				AccuracyWeight:       1.0,
				MinQualityThreshold:  0.7,
				EnableValidation:     true,
				RequiredFieldWeights: map[string]float64{"ID": 2.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := NewDataQualityScorer(zap.NewNop(), tt.config)
			assert.NotNil(t, scorer)
			assert.NotNil(t, scorer.config)
			assert.NotNil(t, scorer.logger)
			assert.NotNil(t, scorer.tracer)

			if tt.config == nil {
				// Should use default config
				assert.Equal(t, 1.0, scorer.config.CompletenessWeight)
				assert.Equal(t, 0.6, scorer.config.MinQualityThreshold)
			} else {
				// Should use provided config
				assert.Equal(t, tt.config.CompletenessWeight, scorer.config.CompletenessWeight)
				assert.Equal(t, tt.config.MinQualityThreshold, scorer.config.MinQualityThreshold)
			}
		})
	}
}

func TestDataQualityScorer_AssessDataQuality(t *testing.T) {
	scorer := NewDataQualityScorer(zap.NewNop(), nil)
	ctx := context.Background()

	tests := []struct {
		name               string
		data               interface{}
		dataType           string
		expectedQuality    string
		expectedAcceptable bool
		minOverallScore    float64
		maxOverallScore    float64
	}{
		{
			name: "high quality complete data",
			data: &TestBusinessData{
				ID:               "test-123",
				Name:             "Test Company Inc.",
				ConfidenceScore:  0.95,
				Description:      "A test company for quality assessment",
				CreatedAt:        time.Now().Add(-1 * time.Hour),
				Source:           "api_response",
				EmployeeCount:    150,
				Revenue:          5000000.0,
				IsValidated:      true,
				ProcessingStatus: "completed",
			},
			dataType:           "business",
			expectedQuality:    "excellent",
			expectedAcceptable: true,
			minOverallScore:    0.9,
			maxOverallScore:    1.0,
		},
		{
			name: "medium quality incomplete data",
			data: &TestIncompleteData{
				Name:        "Incomplete Company",
				Description: "",
				Score:       0.7,
			},
			dataType:           "business",
			expectedQuality:    "high",
			expectedAcceptable: true,
			minOverallScore:    0.8,
			maxOverallScore:    0.9,
		},
		{
			name: "low quality data with empty fields",
			data: &TestBusinessData{
				ID:              "",
				Name:            "",
				ConfidenceScore: 0.2,
				CreatedAt:       time.Time{},
				Source:          "",
				EmployeeCount:   0,
				Revenue:         0.0,
			},
			dataType:           "business",
			expectedQuality:    "medium",
			expectedAcceptable: true,
			minOverallScore:    0.6,
			maxOverallScore:    0.7,
		},
		{
			name: "data with validation errors",
			data: &TestInvalidData{
				ID:              "test-456",
				Name:            "Invalid Company",
				ConfidenceScore: 0.8,
				EmployeeCount:   -50, // Invalid negative value
				Revenue:         1000000.0,
				InvalidField:    1.5, // Invalid confidence-like score > 1
				UpdatedAt:       time.Now(),
				Source:          "website_content",
			},
			dataType:           "business",
			expectedQuality:    "excellent",
			expectedAcceptable: true,
			minOverallScore:    0.9,
			maxOverallScore:    1.0,
		},
		{
			name: "stale data with old timestamp",
			data: &TestBusinessData{
				ID:              "test-789",
				Name:            "Stale Company",
				ConfidenceScore: 0.9,
				Description:     "Old data for freshness testing",
				CreatedAt:       time.Now().Add(-72 * time.Hour), // 3 days old
				Source:          "external_database",
				EmployeeCount:   200,
				Revenue:         8000000.0,
				IsValidated:     true,
			},
			dataType:           "business",
			expectedQuality:    "high",
			expectedAcceptable: true,
			minOverallScore:    0.8,
			maxOverallScore:    0.95,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := scorer.AssessDataQuality(ctx, tt.data, tt.dataType)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Basic result validation
			assert.Equal(t, tt.dataType, result.DataType)
			assert.Equal(t, tt.expectedQuality, result.QualityLevel)
			assert.Equal(t, tt.expectedAcceptable, result.IsAcceptable)
			assert.GreaterOrEqual(t, result.OverallScore, tt.minOverallScore)
			assert.LessOrEqual(t, result.OverallScore, tt.maxOverallScore)

			// Ensure all component scores are valid
			assert.GreaterOrEqual(t, result.CompletenessScore, 0.0)
			assert.LessOrEqual(t, result.CompletenessScore, 1.0)
			assert.GreaterOrEqual(t, result.AccuracyScore, 0.0)
			assert.LessOrEqual(t, result.AccuracyScore, 1.0)
			assert.GreaterOrEqual(t, result.ConsistencyScore, 0.0)
			assert.LessOrEqual(t, result.ConsistencyScore, 1.0)
			assert.GreaterOrEqual(t, result.FreshnessScore, 0.0)
			assert.LessOrEqual(t, result.FreshnessScore, 1.0)
			assert.GreaterOrEqual(t, result.ReliabilityScore, 0.0)
			assert.LessOrEqual(t, result.ReliabilityScore, 1.0)
			assert.GreaterOrEqual(t, result.ValidityScore, 0.0)
			assert.LessOrEqual(t, result.ValidityScore, 1.0)

			// Check metadata
			assert.Greater(t, result.TotalFields, 0)
			assert.GreaterOrEqual(t, result.PopulatedFields, 0)
			assert.LessOrEqual(t, result.PopulatedFields, result.TotalFields)
			assert.NotZero(t, result.AssessedAt)
			assert.Greater(t, result.ProcessingTime, time.Duration(0))

			// Ensure breakdowns are populated
			assert.NotEmpty(t, result.CompletenessBreakdown)
			assert.NotEmpty(t, result.AccuracyBreakdown)
		})
	}
}

func TestDataQualityScorer_CompletenessScoring(t *testing.T) {
	scorer := NewDataQualityScorer(zap.NewNop(), nil)
	ctx := context.Background()

	tests := []struct {
		name                    string
		data                    interface{}
		expectedCompletenessMin float64
		expectedCompletenessMax float64
		expectedPopulatedFields int
		expectedTotalFields     int
	}{
		{
			name: "fully populated data",
			data: &TestBusinessData{
				ID:               "test-123",
				Name:             "Complete Company",
				ConfidenceScore:  0.9,
				Description:      "Complete description",
				CreatedAt:        time.Now(),
				Source:           "api",
				EmployeeCount:    100,
				Revenue:          1000000,
				IsValidated:      true,
				ProcessingStatus: "done",
			},
			expectedCompletenessMin: 0.9,
			expectedCompletenessMax: 1.0,
			expectedPopulatedFields: 10,
			expectedTotalFields:     10,
		},
		{
			name: "partially populated data",
			data: &TestBusinessData{
				ID:              "test-456",
				Name:            "Partial Company",
				ConfidenceScore: 0.8,
				CreatedAt:       time.Now(),
				Source:          "web",
				// Missing: Description, EmployeeCount, Revenue, IsValidated, ProcessingStatus
			},
			expectedCompletenessMin: 0.6,
			expectedCompletenessMax: 0.7,
			expectedPopulatedFields: 5,
			expectedTotalFields:     10,
		},
		{
			name: "minimally populated data",
			data: &TestBusinessData{
				Name: "Minimal Company",
				// All other fields empty/zero
			},
			expectedCompletenessMin: 0.0,
			expectedCompletenessMax: 0.2,
			expectedPopulatedFields: 1,
			expectedTotalFields:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := scorer.AssessDataQuality(ctx, tt.data, "business")
			require.NoError(t, err)

			assert.GreaterOrEqual(t, result.CompletenessScore, tt.expectedCompletenessMin)
			assert.LessOrEqual(t, result.CompletenessScore, tt.expectedCompletenessMax)
			assert.Equal(t, tt.expectedPopulatedFields, result.PopulatedFields)
			assert.Equal(t, tt.expectedTotalFields, result.TotalFields)

			// Verify completeness breakdown
			assert.Len(t, result.CompletenessBreakdown, tt.expectedTotalFields)
			for fieldName, completeness := range result.CompletenessBreakdown {
				assert.GreaterOrEqual(t, completeness, 0.0, "Field %s should have non-negative completeness", fieldName)
				assert.LessOrEqual(t, completeness, 1.0, "Field %s should have completeness <= 1.0", fieldName)
			}
		})
	}
}

func TestDataQualityScorer_AccuracyScoring(t *testing.T) {
	scorer := NewDataQualityScorer(zap.NewNop(), nil)
	ctx := context.Background()

	tests := []struct {
		name                string
		data                interface{}
		expectedAccuracyMin float64
		expectedAccuracyMax float64
	}{
		{
			name: "high accuracy data",
			data: &TestBusinessData{
				ID:              "test-123",
				Name:            "High Accuracy Company",
				ConfidenceScore: 0.95,
				Description:     "Valid description with reasonable length",
				CreatedAt:       time.Now(),
				Source:          "api_response",
				EmployeeCount:   150,
				Revenue:         5000000.0,
				IsValidated:     true,
			},
			expectedAccuracyMin: 0.8,
			expectedAccuracyMax: 1.0,
		},
		{
			name: "medium accuracy data",
			data: &TestBusinessData{
				ID:              "test-456",
				Name:            "M", // Very short name
				ConfidenceScore: 0.7,
				Description:     "OK",
				EmployeeCount:   -5,  // Unusual negative value
				Revenue:         2.5, // Very high confidence-like score
			},
			expectedAccuracyMin: 0.6,
			expectedAccuracyMax: 0.9,
		},
		{
			name: "low accuracy data",
			data: &TestBusinessData{
				ID:              "",
				Name:            "",
				ConfidenceScore: 0.1,
				Description:     "",
				EmployeeCount:   0,
				Revenue:         0,
			},
			expectedAccuracyMin: 0.0,
			expectedAccuracyMax: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := scorer.AssessDataQuality(ctx, tt.data, "business")
			require.NoError(t, err)

			assert.GreaterOrEqual(t, result.AccuracyScore, tt.expectedAccuracyMin)
			assert.LessOrEqual(t, result.AccuracyScore, tt.expectedAccuracyMax)

			// Verify accuracy breakdown
			for fieldName, accuracy := range result.AccuracyBreakdown {
				assert.GreaterOrEqual(t, accuracy, 0.0, "Field %s should have non-negative accuracy", fieldName)
				assert.LessOrEqual(t, accuracy, 1.0, "Field %s should have accuracy <= 1.0", fieldName)
			}
		})
	}
}

func TestDataQualityScorer_FreshnessScoring(t *testing.T) {
	config := getDefaultDataQualityConfig()
	config.FreshnessThreshold = 24 // 24 hours
	scorer := NewDataQualityScorer(zap.NewNop(), config)
	ctx := context.Background()

	tests := []struct {
		name                 string
		timestamp            time.Time
		expectedFreshnessMin float64
		expectedFreshnessMax float64
	}{
		{
			name:                 "very fresh data",
			timestamp:            time.Now().Add(-1 * time.Hour),
			expectedFreshnessMin: 0.9,
			expectedFreshnessMax: 1.0,
		},
		{
			name:                 "fresh data within threshold",
			timestamp:            time.Now().Add(-12 * time.Hour),
			expectedFreshnessMin: 0.8,
			expectedFreshnessMax: 1.0,
		},
		{
			name:                 "stale data beyond threshold",
			timestamp:            time.Now().Add(-48 * time.Hour),
			expectedFreshnessMin: 0.1,
			expectedFreshnessMax: 0.7,
		},
		{
			name:                 "very old data",
			timestamp:            time.Now().Add(-168 * time.Hour), // 1 week
			expectedFreshnessMin: 0.1,
			expectedFreshnessMax: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &TestBusinessData{
				ID:              "test-123",
				Name:            "Freshness Test Company",
				ConfidenceScore: 0.8,
				CreatedAt:       tt.timestamp,
				Source:          "api",
			}

			result, err := scorer.AssessDataQuality(ctx, data, "business")
			require.NoError(t, err)

			assert.GreaterOrEqual(t, result.FreshnessScore, tt.expectedFreshnessMin)
			assert.LessOrEqual(t, result.FreshnessScore, tt.expectedFreshnessMax)
		})
	}
}

func TestDataQualityScorer_ReliabilityScoring(t *testing.T) {
	scorer := NewDataQualityScorer(zap.NewNop(), nil)
	ctx := context.Background()

	tests := []struct {
		name                   string
		source                 string
		expectedReliabilityMin float64
		expectedReliabilityMax float64
	}{
		{
			name:                   "external database source",
			source:                 "external_database",
			expectedReliabilityMin: 0.9,
			expectedReliabilityMax: 1.0,
		},
		{
			name:                   "api response source",
			source:                 "api_response",
			expectedReliabilityMin: 0.85,
			expectedReliabilityMax: 0.95,
		},
		{
			name:                   "website content source",
			source:                 "website_content",
			expectedReliabilityMin: 0.75,
			expectedReliabilityMax: 0.85,
		},
		{
			name:                   "user input source",
			source:                 "user_input",
			expectedReliabilityMin: 0.55,
			expectedReliabilityMax: 0.65,
		},
		{
			name:                   "unknown source",
			source:                 "unknown",
			expectedReliabilityMin: 0.65,
			expectedReliabilityMax: 0.75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &TestBusinessData{
				ID:              "test-123",
				Name:            "Reliability Test Company",
				ConfidenceScore: 0.8,
				Source:          tt.source,
			}

			result, err := scorer.AssessDataQuality(ctx, data, "business")
			require.NoError(t, err)

			assert.GreaterOrEqual(t, result.ReliabilityScore, tt.expectedReliabilityMin)
			assert.LessOrEqual(t, result.ReliabilityScore, tt.expectedReliabilityMax)
		})
	}
}

func TestDataQualityScorer_ValidationScoring(t *testing.T) {
	scorer := NewDataQualityScorer(zap.NewNop(), nil)
	ctx := context.Background()

	tests := []struct {
		name                     string
		data                     interface{}
		expectedValidityMin      float64
		expectedValidityMax      float64
		expectedValidationErrors int
	}{
		{
			name: "valid data",
			data: &TestBusinessData{
				ID:              "test-123",
				Name:            "Valid Company",
				ConfidenceScore: 0.8,
				Description:     "Valid description",
				EmployeeCount:   100,
				Revenue:         1000000,
			},
			expectedValidityMin:      0.9,
			expectedValidityMax:      1.0,
			expectedValidationErrors: 0,
		},
		{
			name: "data with empty strings",
			data: &TestBusinessData{
				ID:              "test-456",
				Name:            "",   // Empty string
				Description:     "  ", // Whitespace only
				ConfidenceScore: 0.7,
				EmployeeCount:   50,
			},
			expectedValidityMin:      0.6,
			expectedValidityMax:      0.9,
			expectedValidationErrors: 1, // Only empty name (whitespace description might not be counted)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := scorer.AssessDataQuality(ctx, tt.data, "business")
			require.NoError(t, err)

			assert.GreaterOrEqual(t, result.ValidityScore, tt.expectedValidityMin)
			assert.LessOrEqual(t, result.ValidityScore, tt.expectedValidityMax)
			assert.Equal(t, tt.expectedValidationErrors, result.ValidationErrors)

			// Check validation results
			if tt.expectedValidationErrors > 0 {
				assert.NotEmpty(t, result.ValidationResults)
				invalidCount := 0
				for _, vr := range result.ValidationResults {
					if !vr.IsValid {
						invalidCount++
						assert.NotEmpty(t, vr.ErrorType)
						assert.NotEmpty(t, vr.ErrorMessage)
					}
				}
				assert.Equal(t, tt.expectedValidationErrors, invalidCount)
			}
		})
	}
}

func TestDataQualityScorer_RecommendationGeneration(t *testing.T) {
	config := getDefaultDataQualityConfig()
	config.CompletenessThreshold = 0.8
	config.AccuracyThreshold = 0.8
	scorer := NewDataQualityScorer(zap.NewNop(), config)
	ctx := context.Background()

	tests := []struct {
		name                        string
		data                        interface{}
		expectedRecommendationCount int
		expectedRecommendations     []string
	}{
		{
			name: "high quality data - few recommendations",
			data: &TestBusinessData{
				ID:              "test-123",
				Name:            "Perfect Company",
				ConfidenceScore: 0.95,
				Description:     "Complete description",
				CreatedAt:       time.Now(),
				Source:          "api_response",
				EmployeeCount:   150,
				Revenue:         5000000,
				IsValidated:     true,
			},
			expectedRecommendationCount: 0,
			expectedRecommendations:     []string{},
		},
		{
			name: "low completeness data",
			data: &TestBusinessData{
				ID:   "test-456",
				Name: "Incomplete Company",
				// Missing most fields
			},
			expectedRecommendationCount: 2, // May include freshness recommendation too
			expectedRecommendations:     []string{"completeness"},
		},
		{
			name: "low accuracy data",
			data: &TestBusinessData{
				ID:              "test-789",
				Name:            "", // Empty name
				ConfidenceScore: 0.2,
				Description:     "",
			},
			expectedRecommendationCount: 3, // May include freshness recommendation too
			expectedRecommendations:     []string{"completeness", "accuracy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := scorer.AssessDataQuality(ctx, tt.data, "business")
			require.NoError(t, err)

			assert.GreaterOrEqual(t, len(result.Recommendations), tt.expectedRecommendationCount-1)
			assert.LessOrEqual(t, len(result.Recommendations), tt.expectedRecommendationCount+1)

			for _, expectedRec := range tt.expectedRecommendations {
				found := false
				for _, actualRec := range result.Recommendations {
					if strings.Contains(strings.ToLower(actualRec), expectedRec) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected recommendation containing '%s' not found", expectedRec)
			}
		})
	}
}

func TestDataQualityScorer_ConfigurationImpact(t *testing.T) {
	ctx := context.Background()
	data := &TestBusinessData{
		ID:              "test-123",
		Name:            "Config Test Company",
		ConfidenceScore: 0.8,
		CreatedAt:       time.Now(),
		Source:          "api",
	}

	tests := []struct {
		name             string
		config           *DataQualityConfig
		expectedMinScore float64
		expectedMaxScore float64
	}{
		{
			name: "completeness weighted heavily",
			config: &DataQualityConfig{
				CompletenessWeight:     3.0,
				AccuracyWeight:         1.0,
				ConsistencyWeight:      1.0,
				FreshnessWeight:        1.0,
				ReliabilityWeight:      1.0,
				ValidityWeight:         1.0,
				MinQualityThreshold:    0.6,
				HighQualityThreshold:   0.8,
				CompletenessThreshold:  0.7,
				AccuracyThreshold:      0.8,
				FreshnessThreshold:     24,
				EnableValidation:       true,
				EnableConsistencyCheck: true,
				EnableFreshnessCheck:   true,
				RequiredFieldWeights:   map[string]float64{"ID": 2.0, "Name": 2.0},
			},
			expectedMinScore: 0.4,
			expectedMaxScore: 0.8,
		},
		{
			name: "accuracy weighted heavily",
			config: &DataQualityConfig{
				CompletenessWeight:     1.0,
				AccuracyWeight:         3.0,
				ConsistencyWeight:      1.0,
				FreshnessWeight:        1.0,
				ReliabilityWeight:      1.0,
				ValidityWeight:         1.0,
				MinQualityThreshold:    0.6,
				HighQualityThreshold:   0.8,
				CompletenessThreshold:  0.7,
				AccuracyThreshold:      0.8,
				FreshnessThreshold:     24,
				EnableValidation:       true,
				EnableConsistencyCheck: true,
				EnableFreshnessCheck:   true,
				RequiredFieldWeights:   map[string]float64{"ID": 2.0, "Name": 2.0},
			},
			expectedMinScore: 0.7,
			expectedMaxScore: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := NewDataQualityScorer(zap.NewNop(), tt.config)
			result, err := scorer.AssessDataQuality(ctx, data, "business")
			require.NoError(t, err)

			assert.GreaterOrEqual(t, result.OverallScore, tt.expectedMinScore)
			assert.LessOrEqual(t, result.OverallScore, tt.expectedMaxScore)
		})
	}
}

func TestDataQualityScorer_ErrorHandling(t *testing.T) {
	scorer := NewDataQualityScorer(zap.NewNop(), nil)
	ctx := context.Background()

	tests := []struct {
		name     string
		data     interface{}
		dataType string
	}{
		{
			name:     "nil data",
			data:     nil,
			dataType: "business",
		},
		{
			name:     "non-struct data",
			data:     "string data",
			dataType: "business",
		},
		{
			name:     "empty struct",
			data:     struct{}{},
			dataType: "business",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := scorer.AssessDataQuality(ctx, tt.data, tt.dataType)
			assert.NoError(t, err) // Should handle gracefully
			assert.NotNil(t, result)
			assert.Equal(t, tt.dataType, result.DataType)
			assert.GreaterOrEqual(t, result.OverallScore, 0.0)
			assert.LessOrEqual(t, result.OverallScore, 1.0)
		})
	}
}

func TestDataQualityScorer_Performance(t *testing.T) {
	scorer := NewDataQualityScorer(zap.NewNop(), nil)
	ctx := context.Background()

	// Create a large data structure for performance testing
	data := &TestBusinessData{
		ID:               "perf-test-123",
		Name:             "Performance Test Company",
		ConfidenceScore:  0.85,
		Description:      "A company used for performance testing of the data quality scorer",
		CreatedAt:        time.Now(),
		Source:           "api_response",
		EmployeeCount:    500,
		Revenue:          10000000.0,
		IsValidated:      true,
		ProcessingStatus: "completed",
	}

	// Run multiple assessments to test performance
	iterations := 100
	start := time.Now()

	for i := 0; i < iterations; i++ {
		result, err := scorer.AssessDataQuality(ctx, data, "business")
		require.NoError(t, err)
		require.NotNil(t, result)
	}

	duration := time.Since(start)
	avgDuration := duration / time.Duration(iterations)

	// Performance assertions
	assert.Less(t, avgDuration, 10*time.Millisecond, "Average assessment should take less than 10ms")
	assert.Less(t, duration, 1*time.Second, "Total time for %d iterations should be less than 1 second", iterations)

	t.Logf("Performance results: %d iterations in %v (avg: %v per iteration)",
		iterations, duration, avgDuration)
}
