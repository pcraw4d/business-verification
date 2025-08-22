package enrichment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewBusinessModelClassifier(t *testing.T) {
	tests := []struct {
		name   string
		config *BusinessModelClassifierConfig
		logger *zap.Logger
	}{
		{
			name:   "with nil inputs",
			config: nil,
			logger: nil,
		},
		{
			name: "with custom config",
			config: &BusinessModelClassifierConfig{
				MinConfidenceThreshold: 0.5,
				MinEvidenceCount:       5,
				ModelIndicatorWeight:   0.3,
				AudienceAnalysisWeight: 0.4,
				RevenueModelWeight:     0.3,
			},
			logger: zap.NewNop(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier := NewBusinessModelClassifier(tt.config, tt.logger)
			require.NotNil(t, classifier)
			assert.NotNil(t, classifier.config)
			assert.NotNil(t, classifier.logger)
			assert.NotNil(t, classifier.tracer)
		})
	}
}

func TestGetDefaultBusinessModelClassifierConfig(t *testing.T) {
	config := GetDefaultBusinessModelClassifierConfig()
	require.NotNil(t, config)

	assert.Equal(t, 0.3, config.MinConfidenceThreshold)
	assert.Equal(t, 3, config.MinEvidenceCount)
	assert.Equal(t, 100, config.MinContentLength)
	assert.Equal(t, 0.25, config.ModelIndicatorWeight)
	assert.Equal(t, 0.30, config.AudienceAnalysisWeight)
	assert.Equal(t, 0.25, config.RevenueModelWeight)
	assert.Equal(t, 0.15, config.ConsistencyWeight)
	assert.Equal(t, 0.05, config.EvidenceWeight)
	assert.True(t, config.RequireMultipleIndicators)
	assert.True(t, config.EnableFallbackAnalysis)
	assert.True(t, config.ValidateClassifications)
	assert.True(t, config.EnableDetailedBreakdown)
}

func TestBusinessModelClassifier_ClassifyBusinessModel(t *testing.T) {
	classifier := NewBusinessModelClassifier(nil, zap.NewNop())

	tests := []struct {
		name                  string
		content               string
		sourceURL             string
		expectedModel         string
		expectedType          string
		expectedMinConfidence float64
		expectedMaxConfidence float64
	}{
		{
			name: "B2B SaaS platform",
			content: `
				Enterprise software solution for large corporations.
				Our B2B platform helps businesses streamline operations.
				Subscription-based pricing with enterprise licensing.
				Targeting Fortune 500 companies and enterprise clients.
				Annual contracts with tiered pricing structure.
			`,
			sourceURL:             "https://b2b-saas.com",
			expectedModel:         "B2B",
			expectedType:          "SaaS",
			expectedMinConfidence: 0.4,
			expectedMaxConfidence: 0.9,
		},
		{
			name: "B2C E-commerce platform",
			content: `
				Online marketplace for consumers to buy products.
				Direct-to-consumer sales with one-time purchases.
				Targeting individual consumers and families.
				Retail pricing with promotional discounts.
				Customer-focused shopping experience.
			`,
			sourceURL:             "https://b2c-ecommerce.com",
			expectedModel:         "B2C",
			expectedType:          "E-commerce",
			expectedMinConfidence: 0.4,
			expectedMaxConfidence: 0.9,
		},
		{
			name: "Marketplace platform",
			content: `
				Multi-sided marketplace connecting buyers and sellers.
				Transaction fee revenue model with commission structure.
				Both B2B and B2C participants on our platform.
				Freemium model with premium features for power users.
				Network effects and marketplace dynamics.
			`,
			sourceURL:             "https://marketplace-platform.com",
			expectedModel:         "Marketplace",
			expectedType:          "B2B2C",
			expectedMinConfidence: 0.4,
			expectedMaxConfidence: 0.9,
		},
		{
			name: "Unknown business model",
			content: `
				Generic company description without clear business model indicators.
				No specific revenue model or target audience mentioned.
				Basic company information and general services.
			`,
			sourceURL:             "https://unknown-company.com",
			expectedModel:         "Unknown",
			expectedType:          "Unknown",
			expectedMinConfidence: 0.0,
			expectedMaxConfidence: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := classifier.ClassifyBusinessModel(context.Background(), tt.content, tt.sourceURL)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Check basic structure
			assert.NotEmpty(t, result.PrimaryBusinessModel)
			assert.NotEmpty(t, result.BusinessModelType)
			assert.GreaterOrEqual(t, result.ConfidenceScore, tt.expectedMinConfidence)
			assert.LessOrEqual(t, result.ConfidenceScore, tt.expectedMaxConfidence)
			assert.NotZero(t, result.ProcessingTime)
			assert.Equal(t, tt.sourceURL, result.SourceURL)
			assert.WithinDuration(t, time.Now(), result.AnalyzedAt, 2*time.Second)

			// Check component analyses
			assert.NotNil(t, result.ModelIndicators)
			assert.NotNil(t, result.AudienceAnalysis)
			assert.NotNil(t, result.RevenueModel)
			assert.NotNil(t, result.MarketPositioning)

			// Check component scores
			assert.GreaterOrEqual(t, result.ComponentScores.ModelIndicatorScore, 0.0)
			assert.LessOrEqual(t, result.ComponentScores.ModelIndicatorScore, 1.0)
			assert.GreaterOrEqual(t, result.ComponentScores.AudienceScore, 0.0)
			assert.LessOrEqual(t, result.ComponentScores.AudienceScore, 1.0)
			assert.GreaterOrEqual(t, result.ComponentScores.RevenueScore, 0.0)
			assert.LessOrEqual(t, result.ComponentScores.RevenueScore, 1.0)

			// Check validation
			assert.NotNil(t, result.ValidationStatus)
			assert.GreaterOrEqual(t, result.DataQualityScore, 0.0)
			assert.LessOrEqual(t, result.DataQualityScore, 1.0)
			assert.NotEmpty(t, result.Reasoning)
		})
	}
}

func TestBusinessModelClassifier_validateInput(t *testing.T) {
	classifier := NewBusinessModelClassifier(nil, zap.NewNop())

	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:        "valid content",
			content:     "This is a valid content with sufficient length for analysis. It contains enough characters to meet the minimum requirement for business model classification analysis.",
			expectError: false,
		},
		{
			name:        "empty content",
			content:     "",
			expectError: true,
		},
		{
			name:        "short content",
			content:     "Short",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classifier.validateInput(tt.content)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBusinessModelClassifier_analyzeModelIndicators(t *testing.T) {
	classifier := NewBusinessModelClassifier(nil, zap.NewNop())

	tests := []struct {
		name          string
		content       string
		expectedType  string
		expectedScore float64
	}{
		{
			name: "B2B indicators",
			content: `
				Enterprise software solution for businesses.
				B2B platform with corporate clients.
				Business-to-business services.
			`,
			expectedType:  "B2B",
			expectedScore: 0.6,
		},
		{
			name: "B2C indicators",
			content: `
				Consumer-focused product for individuals.
				Direct-to-consumer sales.
				Personal use and retail customers.
			`,
			expectedType:  "B2C",
			expectedScore: 0.6,
		},
		{
			name: "Marketplace indicators",
			content: `
				Multi-sided platform connecting buyers and sellers.
				Marketplace with transaction fees.
				Network effects and marketplace dynamics.
			`,
			expectedType:  "Marketplace",
			expectedScore: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &BusinessModelClassification{
				Evidence:         []string{},
				ExtractedPhrases: []string{},
			}

			err := classifier.analyzeModelIndicators(context.Background(), tt.content, result)
			assert.NoError(t, err)

			assert.NotEmpty(t, result.ModelIndicators.Indicators)
			assert.GreaterOrEqual(t, result.ModelIndicators.ConfidenceScore, 0.0)
			assert.LessOrEqual(t, result.ModelIndicators.ConfidenceScore, 1.0)
			assert.NotEmpty(t, result.ModelIndicators.Evidence)
		})
	}
}

func TestBusinessModelClassifier_analyzeAudience(t *testing.T) {
	classifier := NewBusinessModelClassifier(nil, zap.NewNop())

	tests := []struct {
		name             string
		content          string
		expectedAudience string
	}{
		{
			name: "Enterprise audience",
			content: `
				Targeting Fortune 500 companies.
				Enterprise clients and large corporations.
				Business solutions for enterprise customers.
			`,
			expectedAudience: "Enterprise",
		},
		{
			name: "Consumer audience",
			content: `
				Individual consumers and families.
				Personal use and retail customers.
				Direct-to-consumer sales.
			`,
			expectedAudience: "Consumer",
		},
		{
			name: "Mixed audience",
			content: `
				Serving both businesses and consumers.
				B2B and B2C customers on our platform.
				Multi-sided marketplace.
			`,
			expectedAudience: "Mixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &BusinessModelClassification{
				Evidence:         []string{},
				ExtractedPhrases: []string{},
			}

			err := classifier.analyzeAudience(context.Background(), tt.content, result)
			assert.NoError(t, err)

			assert.NotEmpty(t, result.AudienceAnalysis.PrimaryAudience)
			assert.GreaterOrEqual(t, result.AudienceAnalysis.ConfidenceScore, 0.0)
			assert.LessOrEqual(t, result.AudienceAnalysis.ConfidenceScore, 1.0)
			assert.NotEmpty(t, result.AudienceAnalysis.Evidence)
		})
	}
}

func TestBusinessModelClassifier_analyzeRevenueModel(t *testing.T) {
	classifier := NewBusinessModelClassifier(nil, zap.NewNop())

	tests := []struct {
		name          string
		content       string
		expectedModel string
	}{
		{
			name: "Subscription model",
			content: `
				Monthly and annual subscription plans.
				Recurring revenue from subscriptions.
				Subscription-based pricing model.
			`,
			expectedModel: "subscription",
		},
		{
			name: "Marketplace model",
			content: `
				Transaction fees and commission structure.
				Marketplace revenue from fees.
				Percentage-based transaction fees.
			`,
			expectedModel: "marketplace",
		},
		{
			name: "Freemium model",
			content: `
				Free tier with premium features.
				Freemium business model.
				Basic free, premium paid features.
			`,
			expectedModel: "freemium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &BusinessModelClassification{
				Evidence:         []string{},
				ExtractedPhrases: []string{},
			}

			err := classifier.analyzeRevenueModel(context.Background(), tt.content, result)
			assert.NoError(t, err)

			assert.NotEmpty(t, result.RevenueModel.PrimaryRevenueModel)
			assert.GreaterOrEqual(t, result.RevenueModel.ConfidenceScore, 0.0)
			assert.LessOrEqual(t, result.RevenueModel.ConfidenceScore, 1.0)
			assert.NotEmpty(t, result.RevenueModel.Evidence)
		})
	}
}

func TestBusinessModelClassifier_analyzeMarketPositioning(t *testing.T) {
	classifier := NewBusinessModelClassifier(nil, zap.NewNop())

	tests := []struct {
		name            string
		content         string
		expectedSegment string
	}{
		{
			name: "Enterprise segment",
			content: `
				Enterprise-grade solutions for large corporations.
				Fortune 500 clients and enterprise customers.
				High-end business solutions.
			`,
			expectedSegment: "enterprise",
		},
		{
			name: "SMB segment",
			content: `
				Small and medium business solutions.
				SMB customers and mid-market companies.
				Affordable business tools.
			`,
			expectedSegment: "smb",
		},
		{
			name: "Consumer segment",
			content: `
				Consumer-focused products and services.
				Individual users and personal customers.
				Mass market consumer products.
			`,
			expectedSegment: "consumer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &BusinessModelClassification{
				Evidence:         []string{},
				ExtractedPhrases: []string{},
			}

			err := classifier.analyzeMarketPositioning(context.Background(), tt.content, result)
			assert.NoError(t, err)

			assert.NotEmpty(t, result.MarketPositioning.MarketSegment)
			assert.GreaterOrEqual(t, result.MarketPositioning.ConfidenceScore, 0.0)
			assert.LessOrEqual(t, result.MarketPositioning.ConfidenceScore, 1.0)
			assert.NotEmpty(t, result.MarketPositioning.Evidence)
		})
	}
}

func TestBusinessModelClassifier_determineBusinessModels(t *testing.T) {
	classifier := NewBusinessModelClassifier(nil, zap.NewNop())

	tests := []struct {
		name            string
		result          *BusinessModelClassification
		expectedPrimary string
		expectedType    string
	}{
		{
			name: "B2B dominant",
			result: &BusinessModelClassification{
				ModelIndicators: ModelIndicatorAnalysis{
					BusinessType:    "B2B",
					ConfidenceScore: 0.8,
				},
				AudienceAnalysis: AudienceAnalysis{
					PrimaryAudience: "Enterprise",
					ConfidenceScore: 0.7,
				},
				RevenueModel: RevenueModelAnalysis{
					PrimaryRevenueModel: "subscription",
					ConfidenceScore:     0.6,
				},
			},
			expectedPrimary: "B2B",
			expectedType:    "SaaS",
		},
		{
			name: "B2C dominant",
			result: &BusinessModelClassification{
				ModelIndicators: ModelIndicatorAnalysis{
					BusinessType:    "B2C",
					ConfidenceScore: 0.8,
				},
				AudienceAnalysis: AudienceAnalysis{
					PrimaryAudience: "Consumer",
					ConfidenceScore: 0.7,
				},
				RevenueModel: RevenueModelAnalysis{
					PrimaryRevenueModel: "one-time",
					ConfidenceScore:     0.6,
				},
			},
			expectedPrimary: "B2C",
			expectedType:    "E-commerce",
		},
		{
			name: "Marketplace dominant",
			result: &BusinessModelClassification{
				ModelIndicators: ModelIndicatorAnalysis{
					BusinessType:    "Marketplace",
					ConfidenceScore: 0.8,
				},
				AudienceAnalysis: AudienceAnalysis{
					PrimaryAudience: "Mixed",
					ConfidenceScore: 0.7,
				},
				RevenueModel: RevenueModelAnalysis{
					PrimaryRevenueModel: "marketplace",
					ConfidenceScore:     0.6,
				},
			},
			expectedPrimary: "Marketplace",
			expectedType:    "B2B2C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier.determineBusinessModels(tt.result)

			assert.NotEmpty(t, tt.result.PrimaryBusinessModel)
			assert.NotEmpty(t, tt.result.BusinessModelType)
		})
	}
}

func TestBusinessModelClassifier_calculateConfidenceScores(t *testing.T) {
	classifier := NewBusinessModelClassifier(nil, zap.NewNop())

	result := &BusinessModelClassification{
		ModelIndicators: ModelIndicatorAnalysis{
			ConfidenceScore: 0.8,
		},
		AudienceAnalysis: AudienceAnalysis{
			ConfidenceScore: 0.7,
		},
		RevenueModel: RevenueModelAnalysis{
			ConfidenceScore: 0.6,
		},
		MarketPositioning: MarketPositioningAnalysis{
			ConfidenceScore: 0.5,
		},
	}

	classifier.calculateConfidenceScores(result)

	assert.GreaterOrEqual(t, result.ConfidenceScore, 0.0)
	assert.LessOrEqual(t, result.ConfidenceScore, 1.0)
	assert.GreaterOrEqual(t, result.ComponentScores.ModelIndicatorScore, 0.0)
	assert.GreaterOrEqual(t, result.ComponentScores.AudienceScore, 0.0)
	assert.GreaterOrEqual(t, result.ComponentScores.RevenueScore, 0.0)
	assert.GreaterOrEqual(t, result.ComponentScores.ConsistencyScore, 0.0)
	assert.GreaterOrEqual(t, result.ComponentScores.EvidenceScore, 0.0)
}

func TestBusinessModelClassifier_validateResult(t *testing.T) {
	classifier := NewBusinessModelClassifier(nil, zap.NewNop())

	tests := []struct {
		name          string
		result        *BusinessModelClassification
		expectedValid bool
	}{
		{
			name: "valid result",
			result: &BusinessModelClassification{
				PrimaryBusinessModel: "B2B",
				ConfidenceScore:      0.7,
				Evidence:             []string{"evidence1", "evidence2", "evidence3"},
			},
			expectedValid: true,
		},
		{
			name: "low confidence",
			result: &BusinessModelClassification{
				PrimaryBusinessModel: "B2B",
				ConfidenceScore:      0.2,
				Evidence:             []string{"evidence1"},
			},
			expectedValid: false,
		},
		{
			name: "insufficient evidence",
			result: &BusinessModelClassification{
				PrimaryBusinessModel: "B2B",
				ConfidenceScore:      0.7,
				Evidence:             []string{},
			},
			expectedValid: false,
		},
		{
			name: "no primary model",
			result: &BusinessModelClassification{
				PrimaryBusinessModel: "",
				ConfidenceScore:      0.7,
				Evidence:             []string{"evidence1", "evidence2", "evidence3"},
			},
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier.validateResult(tt.result)

			if tt.expectedValid {
				assert.True(t, tt.result.IsValidated)
				assert.True(t, tt.result.ValidationStatus.IsValid)
			} else {
				assert.False(t, tt.result.ValidationStatus.IsValid)
			}
		})
	}
}

func TestBusinessModelClassifier_generateReasoning(t *testing.T) {
	classifier := NewBusinessModelClassifier(nil, zap.NewNop())

	result := &BusinessModelClassification{
		PrimaryBusinessModel: "B2B",
		BusinessModelType:    "SaaS",
		ModelIndicators: ModelIndicatorAnalysis{
			BusinessType:    "B2B",
			ConfidenceScore: 0.8,
		},
		AudienceAnalysis: AudienceAnalysis{
			PrimaryAudience: "Enterprise",
			ConfidenceScore: 0.7,
		},
		RevenueModel: RevenueModelAnalysis{
			PrimaryRevenueModel: "subscription",
			ConfidenceScore:     0.6,
		},
		ConfidenceScore: 0.7,
	}

	reasoning := classifier.generateReasoning(result)

	assert.NotEmpty(t, reasoning)
	assert.Contains(t, reasoning, "B2B")
	assert.Contains(t, reasoning, "SaaS")
	assert.Contains(t, reasoning, "Enterprise")
	assert.Contains(t, reasoning, "subscription")
}

func TestBusinessModelClassifier_Integration(t *testing.T) {
	classifier := NewBusinessModelClassifier(nil, zap.NewNop())

	content := `
		Enterprise SaaS platform for large corporations.
		Our B2B solution helps Fortune 500 companies streamline operations.
		Subscription-based pricing with annual contracts.
		Targeting enterprise clients with enterprise-grade features.
		Monthly and annual subscription plans available.
		Business-to-business software for corporate customers.
	`

	result, err := classifier.ClassifyBusinessModel(context.Background(), content, "https://test-company.com")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify comprehensive analysis
	assert.Equal(t, "B2B", result.PrimaryBusinessModel)
	assert.Contains(t, []string{"SaaS", "B2B"}, result.BusinessModelType)
	assert.GreaterOrEqual(t, result.ConfidenceScore, 0.6)
	assert.True(t, result.IsValidated)
	assert.True(t, result.ValidationStatus.IsValid)
	assert.Greater(t, result.DataQualityScore, 0.5)
	assert.NotEmpty(t, result.Reasoning)
	assert.NotEmpty(t, result.Evidence)
	assert.NotEmpty(t, result.ExtractedPhrases)

	// Verify component analyses
	assert.NotEmpty(t, result.ModelIndicators.Indicators)
	assert.Equal(t, "B2B", result.ModelIndicators.BusinessType)
	assert.Greater(t, result.ModelIndicators.ConfidenceScore, 0.5)

	assert.Equal(t, "Enterprise", result.AudienceAnalysis.PrimaryAudience)
	assert.Greater(t, result.AudienceAnalysis.ConfidenceScore, 0.5)

	assert.Equal(t, "subscription", result.RevenueModel.PrimaryRevenueModel)
	assert.Greater(t, result.RevenueModel.ConfidenceScore, 0.5)

	assert.Contains(t, []string{"enterprise", "Enterprise"}, result.MarketPositioning.MarketSegment)
	assert.Greater(t, result.MarketPositioning.ConfidenceScore, 0.5)
}

func TestBusinessModelClassifier_Performance(t *testing.T) {
	classifier := NewBusinessModelClassifier(nil, zap.NewNop())

	content := `
		Enterprise SaaS platform for large corporations.
		Our B2B solution helps Fortune 500 companies streamline operations.
		Subscription-based pricing with annual contracts.
		Targeting enterprise clients with enterprise-grade features.
		Monthly and annual subscription plans available.
		Business-to-business software for corporate customers.
	`

	start := time.Now()
	result, err := classifier.ClassifyBusinessModel(context.Background(), content, "https://test-company.com")
	duration := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Performance assertions
	assert.Less(t, duration, 100*time.Millisecond, "Classification should complete within 100ms")
	assert.Greater(t, result.ProcessingTime, time.Duration(0))
	assert.Less(t, result.ProcessingTime, 100*time.Millisecond)
}
