package data_discovery

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestQualityScorer_ScoreField(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	scorer := NewQualityScorer(config, logger)

	tests := []struct {
		name             string
		field            DiscoveredField
		patterns         []PatternMatch
		classification   *ClassificationResult
		businessContext  *BusinessContext
		expectedMinScore float64
		expectedCategory string
	}{
		{
			name: "high quality email field",
			field: DiscoveredField{
				FieldName:        "email",
				FieldType:        "email",
				DataType:         "string",
				ConfidenceScore:  0.9,
				ExtractionMethod: "regex",
				SampleValues:     []string{"test@example.com", "info@example.com"},
				ValidationRules:  []ValidationRule{{RuleType: "email_format"}},
				Priority:         1,
				BusinessValue:    0.9,
				Metadata:         map[string]interface{}{"source_url": "https://example.com"},
			},
			patterns: []PatternMatch{
				{
					FieldType:       "email",
					ConfidenceScore: 0.9,
					Context:         "Contact email",
				},
			},
			classification: &ClassificationResult{
				BusinessType:     "B2B",
				IndustryCategory: "technology",
				ConfidenceScore:  0.8,
			},
			businessContext: &BusinessContext{
				Industry:       "technology",
				BusinessType:   "B2B",
				UseCaseProfile: "verification",
				PriorityFields: []string{"email", "phone"},
				CustomWeights:  map[string]float64{"email": 1.0},
			},
			expectedMinScore: 0.8,
			expectedCategory: "excellent",
		},
		{
			name: "medium quality phone field",
			field: DiscoveredField{
				FieldName:        "phone",
				FieldType:        "phone",
				DataType:         "string",
				ConfidenceScore:  0.7,
				ExtractionMethod: "pattern_matching",
				SampleValues:     []string{"(555) 123-4567"},
				Priority:         2,
				BusinessValue:    0.7,
				Metadata:         map[string]interface{}{},
			},
			patterns: []PatternMatch{
				{
					FieldType:       "phone",
					ConfidenceScore: 0.7,
					Context:         "Phone number",
				},
			},
			classification: &ClassificationResult{
				BusinessType:     "B2C",
				IndustryCategory: "retail",
				ConfidenceScore:  0.7,
			},
			businessContext: &BusinessContext{
				Industry:       "retail",
				BusinessType:   "B2C",
				UseCaseProfile: "verification",
				PriorityFields: []string{"phone", "address"},
				CustomWeights:  map[string]float64{"phone": 0.9},
			},
			expectedMinScore: 0.5,
			expectedCategory: "good",
		},
		{
			name: "low quality field with missing data",
			field: DiscoveredField{
				FieldName:        "unknown_field",
				FieldType:        "unknown",
				DataType:         "",
				ConfidenceScore:  0.3,
				ExtractionMethod: "",
				SampleValues:     []string{},
				ValidationRules:  []ValidationRule{},
				Priority:         5,
				BusinessValue:    0.2,
				Metadata:         map[string]interface{}{},
			},
			patterns:         []PatternMatch{},
			classification:   nil,
			businessContext:  nil,
			expectedMinScore: 0.0,
			expectedCategory: "fair",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			assessment, err := scorer.ScoreField(ctx, tt.field, tt.patterns, tt.classification, tt.businessContext)

			require.NoError(t, err)
			assert.NotNil(t, assessment)
			assert.Equal(t, tt.field.FieldName, assessment.FieldName)
			assert.Equal(t, tt.field.FieldType, assessment.FieldType)
			assert.GreaterOrEqual(t, assessment.QualityScore.OverallScore, tt.expectedMinScore)
			assert.Equal(t, tt.expectedCategory, assessment.QualityCategory)

			// Verify scoring components
			assert.GreaterOrEqual(t, len(assessment.QualityScore.ScoringComponents), 6)

			// Verify all component scores are within valid range
			for _, component := range assessment.QualityScore.ScoringComponents {
				assert.GreaterOrEqual(t, component.Score, 0.0)
				assert.LessOrEqual(t, component.Score, 1.0)
				assert.Greater(t, component.Weight, 0.0)
				assert.NotEmpty(t, component.ComponentName)
				assert.NotEmpty(t, component.Description)
			}

			// Verify quality indicators
			assert.NotEmpty(t, assessment.QualityScore.QualityIndicators)

			// Verify value metrics
			assert.GreaterOrEqual(t, assessment.ValueMetrics.BusinessValue, 0.0)
			assert.LessOrEqual(t, assessment.ValueMetrics.BusinessValue, 1.0)

			t.Logf("Field: %s, Overall Score: %.3f, Category: %s",
				assessment.FieldName,
				assessment.QualityScore.OverallScore,
				assessment.QualityCategory)
		})
	}
}

func TestQualityScorer_ScoreDiscoveredFields(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	scorer := NewQualityScorer(config, logger)

	fields := []DiscoveredField{
		{
			FieldName:        "email",
			FieldType:        "email",
			ConfidenceScore:  0.9,
			ExtractionMethod: "regex",
			SampleValues:     []string{"test@example.com"},
			Priority:         1,
			BusinessValue:    0.9,
		},
		{
			FieldName:        "phone",
			FieldType:        "phone",
			ConfidenceScore:  0.8,
			ExtractionMethod: "pattern_matching",
			SampleValues:     []string{"(555) 123-4567"},
			Priority:         1,
			BusinessValue:    0.8,
		},
		{
			FieldName:        "address",
			FieldType:        "address",
			ConfidenceScore:  0.7,
			ExtractionMethod: "regex",
			SampleValues:     []string{"123 Business Ave, Suite 100, Anytown, NY 12345"},
			Priority:         2,
			BusinessValue:    0.7,
		},
	}

	patterns := []PatternMatch{
		{FieldType: "email", ConfidenceScore: 0.9},
		{FieldType: "phone", ConfidenceScore: 0.8},
		{FieldType: "address", ConfidenceScore: 0.7},
	}

	classification := &ClassificationResult{
		BusinessType:     "B2B",
		IndustryCategory: "technology",
		ConfidenceScore:  0.8,
	}

	businessContext := &BusinessContext{
		Industry:       "technology",
		BusinessType:   "B2B",
		UseCaseProfile: "verification",
		PriorityFields: []string{"email", "phone", "address"},
	}

	ctx := context.Background()
	assessments, err := scorer.ScoreDiscoveredFields(ctx, fields, patterns, classification, businessContext)

	require.NoError(t, err)
	assert.Len(t, assessments, 3)

	// Verify all assessments have valid scores
	for _, assessment := range assessments {
		assert.GreaterOrEqual(t, assessment.QualityScore.OverallScore, 0.0)
		assert.LessOrEqual(t, assessment.QualityScore.OverallScore, 1.0)
		assert.NotEmpty(t, assessment.FieldName)
		assert.NotEmpty(t, assessment.FieldType)
		assert.NotEmpty(t, assessment.QualityCategory)
		assert.NotEmpty(t, assessment.BusinessImpact)

		t.Logf("Assessment: %s - Score: %.3f, Category: %s, Impact: %s",
			assessment.FieldName,
			assessment.QualityScore.OverallScore,
			assessment.QualityCategory,
			assessment.BusinessImpact)
	}
}

func TestQualityScorer_RelevanceScoring(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	scorer := NewQualityScorer(config, logger)

	field := DiscoveredField{
		FieldName:     "email",
		FieldType:     "email",
		Priority:      1,
		BusinessValue: 0.9,
	}

	tests := []struct {
		name            string
		classification  *ClassificationResult
		businessContext *BusinessContext
		expectedMin     float64
		expectedMax     float64
	}{
		{
			name: "technology industry high relevance",
			classification: &ClassificationResult{
				IndustryCategory: "technology",
				ConfidenceScore:  0.8,
			},
			businessContext: &BusinessContext{
				Industry:       "technology",
				BusinessType:   "B2B",
				PriorityFields: []string{"email"},
				CustomWeights:  map[string]float64{"email": 1.0},
			},
			expectedMin: 0.8,
			expectedMax: 1.0,
		},
		{
			name: "finance industry medium relevance",
			classification: &ClassificationResult{
				IndustryCategory: "finance",
				ConfidenceScore:  0.7,
			},
			businessContext: &BusinessContext{
				Industry:      "finance",
				BusinessType:  "B2B",
				CustomWeights: map[string]float64{"email": 0.9},
			},
			expectedMin: 0.6,
			expectedMax: 1.0,
		},
		{
			name:            "no context",
			classification:  nil,
			businessContext: nil,
			expectedMin:     0.4,
			expectedMax:     0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateRelevanceScore(field, tt.classification, tt.businessContext)

			assert.GreaterOrEqual(t, score, tt.expectedMin)
			assert.LessOrEqual(t, score, tt.expectedMax)
			assert.GreaterOrEqual(t, score, 0.0)
			assert.LessOrEqual(t, score, 1.0)

			t.Logf("Relevance score: %.3f", score)
		})
	}
}

func TestQualityScorer_AccuracyScoring(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	scorer := NewQualityScorer(config, logger)

	tests := []struct {
		name        string
		field       DiscoveredField
		patterns    []PatternMatch
		expectedMin float64
	}{
		{
			name: "high accuracy with valid samples",
			field: DiscoveredField{
				FieldType:       "email",
				ConfidenceScore: 0.9,
				SampleValues:    []string{"test@example.com", "info@example.com"},
				ValidationRules: []ValidationRule{{RuleType: "email_format"}},
			},
			patterns: []PatternMatch{
				{FieldType: "email", ConfidenceScore: 0.9},
			},
			expectedMin: 0.8,
		},
		{
			name: "medium accuracy with some invalid samples",
			field: DiscoveredField{
				FieldType:       "email",
				ConfidenceScore: 0.7,
				SampleValues:    []string{"test@example.com", "invalid-email"},
			},
			patterns: []PatternMatch{
				{FieldType: "email", ConfidenceScore: 0.7},
			},
			expectedMin: 0.5,
		},
		{
			name: "low accuracy with no samples",
			field: DiscoveredField{
				FieldType:       "unknown",
				ConfidenceScore: 0.3,
				SampleValues:    []string{},
			},
			patterns:    []PatternMatch{},
			expectedMin: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateAccuracyScore(tt.field, tt.patterns)

			assert.GreaterOrEqual(t, score, tt.expectedMin)
			assert.GreaterOrEqual(t, score, 0.0)
			assert.LessOrEqual(t, score, 1.0)

			t.Logf("Accuracy score: %.3f", score)
		})
	}
}

func TestQualityScorer_CompletenessScoring(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	scorer := NewQualityScorer(config, logger)

	tests := []struct {
		name        string
		field       DiscoveredField
		expectedMin float64
	}{
		{
			name: "complete field with all data",
			field: DiscoveredField{
				FieldType:        "email",
				DataType:         "string",
				ExtractionMethod: "regex",
				SampleValues:     []string{"test1@example.com", "test2@example.com"},
				ValidationRules:  []ValidationRule{{RuleType: "email_format"}},
				Metadata:         map[string]interface{}{"source": "webpage"},
			},
			expectedMin: 0.9,
		},
		{
			name: "partial field with some data",
			field: DiscoveredField{
				FieldType:    "phone",
				DataType:     "string",
				SampleValues: []string{"(555) 123-4567"},
			},
			expectedMin: 0.6,
		},
		{
			name: "minimal field with basic data",
			field: DiscoveredField{
				FieldType: "unknown",
			},
			expectedMin: 0.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateCompletenessScore(tt.field)

			assert.GreaterOrEqual(t, score, tt.expectedMin)
			assert.GreaterOrEqual(t, score, 0.0)
			assert.LessOrEqual(t, score, 1.0)

			t.Logf("Completeness score: %.3f", score)
		})
	}
}

func TestQualityScorer_FreshnessScoring(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	scorer := NewQualityScorer(config, logger)

	tests := []struct {
		name        string
		field       DiscoveredField
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "fresh data",
			field: DiscoveredField{
				FieldType: "email",
				Metadata: map[string]interface{}{
					"extracted_at": time.Now().Add(-30 * time.Minute),
				},
			},
			expectedMin: 0.8,
			expectedMax: 1.0,
		},
		{
			name: "day old data",
			field: DiscoveredField{
				FieldType: "email",
				Metadata: map[string]interface{}{
					"extracted_at": time.Now().Add(-12 * time.Hour),
				},
			},
			expectedMin: 0.7,
			expectedMax: 0.9,
		},
		{
			name: "no timestamp (newly discovered)",
			field: DiscoveredField{
				FieldType: "email",
				Metadata:  map[string]interface{}{},
			},
			expectedMin: 0.8,
			expectedMax: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateFreshnessScore(tt.field)

			assert.GreaterOrEqual(t, score, tt.expectedMin)
			assert.LessOrEqual(t, score, tt.expectedMax)
			assert.GreaterOrEqual(t, score, 0.0)
			assert.LessOrEqual(t, score, 1.0)

			t.Logf("Freshness score: %.3f", score)
		})
	}
}

func TestQualityScorer_ValidationHelpers(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	scorer := NewQualityScorer(config, logger)

	t.Run("validateSampleValue", func(t *testing.T) {
		tests := []struct {
			value     string
			fieldType string
			expected  bool
		}{
			{"test@example.com", "email", true},
			{"invalid-email", "email", false},
			{"(555) 123-4567", "phone", true},
			{"+1-555-123-4567", "phone", true},
			{"invalid-phone", "phone", false},
			{"https://example.com", "url", true},
			{"http://example.com", "url", true},
			{"invalid-url", "url", false},
			{"123 Main St, City, State", "address", true},
			{"Short", "address", false},
		}

		for _, tt := range tests {
			result := scorer.validateSampleValue(tt.value, tt.fieldType)
			assert.Equal(t, tt.expected, result,
				"validateSampleValue(%s, %s) should be %v", tt.value, tt.fieldType, tt.expected)
		}
	})

	t.Run("evaluateURLCredibility", func(t *testing.T) {
		tests := []struct {
			url         string
			expectedMin float64
		}{
			{"https://example.gov", 0.8},
			{"https://example.edu", 0.8},
			{"https://example.org", 0.6},
			{"https://example.com", 0.5},
			{"http://example.com", 0.5},
		}

		for _, tt := range tests {
			score := scorer.evaluateURLCredibility(tt.url)
			assert.GreaterOrEqual(t, score, tt.expectedMin,
				"evaluateURLCredibility(%s) should be >= %f", tt.url, tt.expectedMin)
			assert.LessOrEqual(t, score, 1.0)
		}
	})

	t.Run("evaluateSampleConsistency", func(t *testing.T) {
		tests := []struct {
			name      string
			samples   []string
			fieldType string
			expected  float64
		}{
			{
				name:      "consistent emails",
				samples:   []string{"test@example.com", "info@example.com"},
				fieldType: "email",
				expected:  1.0,
			},
			{
				name:      "inconsistent emails",
				samples:   []string{"test@example.com", "info@different.com"},
				fieldType: "email",
				expected:  0.0,
			},
			{
				name:      "single sample",
				samples:   []string{"test@example.com"},
				fieldType: "email",
				expected:  1.0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				score := scorer.evaluateSampleConsistency(tt.samples, tt.fieldType)
				assert.Equal(t, tt.expected, score)
			})
		}
	})
}

func TestQualityScorer_Integration(t *testing.T) {
	// Test the quality scorer integration with the data discovery service
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	service := NewDataDiscoveryService(config, logger)

	content := &ContentInput{
		RawContent: `
			Acme Corporation
			Contact: info@acme.com | Phone: (555) 123-4567
			Address: 123 Business Ave, Suite 100, Anytown, NY 12345
			Website: https://www.acme.com
			Founded: 2010
		`,
		ContentType: "html",
		URL:         "https://www.acme.com",
		MetaData: map[string]string{
			"industry":      "technology",
			"business_type": "B2B",
		},
	}

	ctx := context.Background()
	result, err := service.DiscoverDataPoints(ctx, content)

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify quality assessments were generated
	assert.Greater(t, len(result.QualityAssessments), 0)

	t.Logf("Generated %d quality assessments", len(result.QualityAssessments))

	// Check each quality assessment
	for _, assessment := range result.QualityAssessments {
		assert.NotEmpty(t, assessment.FieldName)
		assert.NotEmpty(t, assessment.FieldType)
		assert.GreaterOrEqual(t, assessment.QualityScore.OverallScore, 0.0)
		assert.LessOrEqual(t, assessment.QualityScore.OverallScore, 1.0)
		assert.NotEmpty(t, assessment.QualityCategory)
		assert.NotEmpty(t, assessment.BusinessImpact)

		// Verify scoring components
		assert.GreaterOrEqual(t, len(assessment.QualityScore.ScoringComponents), 6)

		t.Logf("Field: %s, Score: %.3f, Category: %s, Impact: %s",
			assessment.FieldName,
			assessment.QualityScore.OverallScore,
			assessment.QualityCategory,
			assessment.BusinessImpact)
	}

	// Test helper methods
	highQualityFields := service.GetHighQualityFields(result, 0.7)
	assert.GreaterOrEqual(t, len(highQualityFields), 0)

	sortedAssessments := service.GetQualityAssessmentsByScore(result)
	assert.Len(t, sortedAssessments, len(result.QualityAssessments))

	// Verify sorting order
	for i := 0; i < len(sortedAssessments)-1; i++ {
		assert.GreaterOrEqual(t,
			sortedAssessments[i].QualityScore.OverallScore,
			sortedAssessments[i+1].QualityScore.OverallScore)
	}

	criticalFields := service.GetCriticalBusinessImpactFields(result)
	assert.GreaterOrEqual(t, len(criticalFields), 0)

	t.Logf("High quality fields: %d, Critical impact fields: %d",
		len(highQualityFields), len(criticalFields))
}
