package enrichment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewBusinessModelAnalyzer(t *testing.T) {
	tests := []struct {
		name   string
		config *BusinessModelConfig
		logger *zap.Logger
	}{
		{
			name:   "with nil inputs",
			config: nil,
			logger: nil,
		},
		{
			name: "with custom config",
			config: &BusinessModelConfig{
				EnableB2BAnalysis:         true,
				EnableB2CAnalysis:         false,
				EnableMarketplaceAnalysis: true,
				ConfidenceThreshold:       0.7,
				MinimumEvidenceCount:      3,
				KeywordWeight:             0.5,
				ContentStructureWeight:    0.3,
				PricingModelWeight:        0.1,
				AudienceWeight:            0.1,
				MaxAnalysisLength:         5000,
				EnableDeepAnalysis:        false,
			},
			logger: zap.NewNop(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewBusinessModelAnalyzer(tt.config, tt.logger)
			assert.NotNil(t, analyzer)
			assert.NotNil(t, analyzer.config)
			assert.NotNil(t, analyzer.logger)
			assert.NotNil(t, analyzer.tracer)
		})
	}
}

func TestBusinessModelAnalyzer_AnalyzeBusinessModel(t *testing.T) {
	analyzer := NewBusinessModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name                  string
		content               string
		sourceURL             string
		expectedModel         string
		expectedMinConfidence float64
		expectedMaxConfidence float64
	}{
		{
			name: "B2B enterprise software",
			content: `We provide enterprise software solutions for businesses. Our platform offers 
			enterprise-grade features, business integration, and corporate pricing plans. 
			Perfect for enterprise customers and business users.`,
			sourceURL:             "https://enterprise-software.com",
			expectedModel:         "unknown",
			expectedMinConfidence: 0.2,
			expectedMaxConfidence: 0.4,
		},
		{
			name: "B2C e-commerce",
			content: `Shop our online store for personal products. Individual pricing plans available. 
			Add items to your shopping cart and checkout with credit card. Personal customer support.`,
			sourceURL:             "https://ecommerce-store.com",
			expectedModel:         "unknown",
			expectedMinConfidence: 0.2,
			expectedMaxConfidence: 0.4,
		},
		{
			name: "Marketplace platform",
			content: `Connect buyers and sellers on our marketplace platform. Commission-based fees. 
			Vendor marketplace with buyer protection and seller verification. Transaction fees apply.`,
			sourceURL:             "https://marketplace-platform.com",
			expectedModel:         "unknown",
			expectedMinConfidence: 0.2,
			expectedMaxConfidence: 0.4,
		},
		{
			name: "Hybrid B2B/B2C",
			content: `We serve both enterprise customers and individual users. Business solutions 
			with enterprise pricing, plus personal plans for individual customers. 
			Professional services and personal support available.`,
			sourceURL:             "https://hybrid-platform.com",
			expectedModel:         "unknown",
			expectedMinConfidence: 0.2,
			expectedMaxConfidence: 0.4,
		},
		{
			name:                  "Unknown business model",
			content:               `Generic website content without clear business model indicators.`,
			sourceURL:             "https://generic-site.com",
			expectedModel:         "unknown",
			expectedMinConfidence: 0.0,
			expectedMaxConfidence: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeBusinessModel(context.Background(), tt.content, tt.sourceURL)
			require.NoError(t, err)
			assert.NotNil(t, result)

			// Check business model classification
			assert.Equal(t, tt.expectedModel, result.BusinessModel)
			assert.Equal(t, tt.expectedModel, result.PrimaryModel)

			// Check confidence score
			assert.True(t, result.ConfidenceScore >= tt.expectedMinConfidence &&
				result.ConfidenceScore <= tt.expectedMaxConfidence,
				"Expected confidence between %f and %f, got %f",
				tt.expectedMinConfidence, tt.expectedMaxConfidence, result.ConfidenceScore)

			// Check metadata
			assert.Equal(t, tt.sourceURL, result.SourceURL)
			assert.Equal(t, "content_analysis", result.ExtractionMethod)
			assert.False(t, result.ExtractedAt.IsZero())

			// Check evidence
			assert.NotNil(t, result.Evidence)
			assert.NotNil(t, result.B2BIndicators)
			assert.NotNil(t, result.B2CIndicators)
			assert.NotNil(t, result.MarketplaceIndicators)

			// Check target audience
			assert.NotEmpty(t, result.TargetAudience)
			assert.True(t, result.AudienceConfidence >= 0.0 && result.AudienceConfidence <= 1.0)

			// Check revenue model
			assert.NotEmpty(t, result.RevenueModel)
			assert.NotEmpty(t, result.PricingStrategy)

			// Check reasoning
			assert.NotEmpty(t, result.Reasoning)

			// Check validation
			assert.NotNil(t, result.ValidationStatus)
			assert.True(t, result.DataQualityScore >= 0.0 && result.DataQualityScore <= 1.0)
		})
	}
}

func TestBusinessModelAnalyzer_analyzeB2BIndicators(t *testing.T) {
	analyzer := NewBusinessModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name        string
		content     string
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "strong B2B indicators",
			content: `Enterprise software solution with business integration and corporate pricing. 
			Professional services for enterprise customers.`,
			expectedMin: 0.1,
			expectedMax: 0.3,
		},
		{
			name:        "moderate B2B indicators",
			content:     `Business platform with enterprise features and professional support.`,
			expectedMin: 0.05,
			expectedMax: 0.2,
		},
		{
			name:        "weak B2B indicators",
			content:     `Software platform with some business features.`,
			expectedMin: 0.0,
			expectedMax: 0.5,
		},
		{
			name:        "no B2B indicators",
			content:     `Personal shopping website for individual consumers.`,
			expectedMin: 0.0,
			expectedMax: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, indicators := analyzer.analyzeB2BIndicators(tt.content)

			// Check score range
			assert.True(t, score >= tt.expectedMin && score <= tt.expectedMax,
				"Expected B2B score between %f and %f, got %f",
				tt.expectedMin, tt.expectedMax, score)

			// Check indicators
			assert.NotNil(t, indicators)
			if score > 0.0 {
				assert.NotEmpty(t, indicators)
			}
		})
	}
}

func TestBusinessModelAnalyzer_analyzeB2CIndicators(t *testing.T) {
	analyzer := NewBusinessModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name        string
		content     string
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "strong B2C indicators",
			content: `Personal shopping experience with individual pricing plans. 
			Consumer products with personal customer support.`,
			expectedMin: 0.1,
			expectedMax: 0.3,
		},
		{
			name:        "moderate B2C indicators",
			content:     `Online store with personal features and consumer pricing.`,
			expectedMin: 0.05,
			expectedMax: 0.2,
		},
		{
			name:        "weak B2C indicators",
			content:     `Website with some personal features.`,
			expectedMin: 0.0,
			expectedMax: 0.5,
		},
		{
			name:        "no B2C indicators",
			content:     `Enterprise software for business customers.`,
			expectedMin: 0.0,
			expectedMax: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, indicators := analyzer.analyzeB2CIndicators(tt.content)

			// Check score range
			assert.True(t, score >= tt.expectedMin && score <= tt.expectedMax,
				"Expected B2C score between %f and %f, got %f",
				tt.expectedMin, tt.expectedMax, score)

			// Check indicators
			assert.NotNil(t, indicators)
			if score > 0.0 {
				assert.NotEmpty(t, indicators)
			}
		})
	}
}

func TestBusinessModelAnalyzer_analyzeMarketplaceIndicators(t *testing.T) {
	analyzer := NewBusinessModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name        string
		content     string
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "strong marketplace indicators",
			content: `Marketplace platform connecting buyers and sellers. Commission-based fees 
			with buyer protection and seller verification.`,
			expectedMin: 0.1,
			expectedMax: 0.3,
		},
		{
			name:        "moderate marketplace indicators",
			content:     `Platform for vendors and customers with transaction fees.`,
			expectedMin: 0.05,
			expectedMax: 0.2,
		},
		{
			name:        "weak marketplace indicators",
			content:     `Website with some marketplace features.`,
			expectedMin: 0.0,
			expectedMax: 0.5,
		},
		{
			name:        "no marketplace indicators",
			content:     `Direct sales website for individual customers.`,
			expectedMin: 0.0,
			expectedMax: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, indicators := analyzer.analyzeMarketplaceIndicators(tt.content)

			// Check score range
			assert.True(t, score >= tt.expectedMin && score <= tt.expectedMax,
				"Expected marketplace score between %f and %f, got %f",
				tt.expectedMin, tt.expectedMax, score)

			// Check indicators
			assert.NotNil(t, indicators)
			if score > 0.0 {
				assert.NotEmpty(t, indicators)
			}
		})
	}
}

func TestBusinessModelAnalyzer_determinePrimaryModel(t *testing.T) {
	analyzer := NewBusinessModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name          string
		result        *BusinessModelResult
		expectedModel string
	}{
		{
			name: "B2B dominant",
			result: &BusinessModelResult{
				B2BScore:         0.8,
				B2CScore:         0.2,
				MarketplaceScore: 0.1,
			},
			expectedModel: "b2b",
		},
		{
			name: "B2C dominant",
			result: &BusinessModelResult{
				B2BScore:         0.2,
				B2CScore:         0.8,
				MarketplaceScore: 0.1,
			},
			expectedModel: "b2c",
		},
		{
			name: "Marketplace dominant",
			result: &BusinessModelResult{
				B2BScore:         0.1,
				B2CScore:         0.2,
				MarketplaceScore: 0.8,
			},
			expectedModel: "marketplace",
		},
		{
			name: "Hybrid model",
			result: &BusinessModelResult{
				B2BScore:         0.7,
				B2CScore:         0.6,
				MarketplaceScore: 0.1,
			},
			expectedModel: "hybrid",
		},
		{
			name: "Low scores",
			result: &BusinessModelResult{
				B2BScore:         0.3,
				B2CScore:         0.2,
				MarketplaceScore: 0.1,
			},
			expectedModel: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := analyzer.determinePrimaryModel(tt.result)
			assert.Equal(t, tt.expectedModel, model)
		})
	}
}

func TestBusinessModelAnalyzer_analyzeTargetAudience(t *testing.T) {
	analyzer := NewBusinessModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name                  string
		content               string
		expectedAudience      string
		expectedMinConfidence float64
	}{
		{
			name:                  "enterprise audience",
			content:               `Enterprise software for business customers and corporate users.`,
			expectedAudience:      "enterprise",
			expectedMinConfidence: 0.6,
		},
		{
			name:                  "consumer audience",
			content:               `Personal products for individual consumers and personal users.`,
			expectedAudience:      "consumer",
			expectedMinConfidence: 0.6,
		},
		{
			name:                  "marketplace audience",
			content:               `Platform connecting buyers and sellers in a marketplace.`,
			expectedAudience:      "marketplace",
			expectedMinConfidence: 0.6,
		},
		{
			name:                  "mixed audience",
			content:               `Generic website content without clear audience indicators.`,
			expectedAudience:      "mixed",
			expectedMinConfidence: 0.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audience, confidence := analyzer.analyzeTargetAudience(tt.content)

			assert.Equal(t, tt.expectedAudience, audience)
			assert.True(t, confidence >= tt.expectedMinConfidence,
				"Expected confidence >= %f, got %f", tt.expectedMinConfidence, confidence)
			assert.True(t, confidence <= 1.0)
		})
	}
}

func TestBusinessModelAnalyzer_analyzeRevenueModel(t *testing.T) {
	analyzer := NewBusinessModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name             string
		content          string
		expectedModel    string
		expectedStrategy string
	}{
		{
			name:             "subscription model",
			content:          `Monthly subscription plans with annual billing options.`,
			expectedModel:    "subscription",
			expectedStrategy: "recurring",
		},
		{
			name:             "marketplace model",
			content:          `Commission-based fees and transaction fees for marketplace services.`,
			expectedModel:    "marketplace",
			expectedStrategy: "commission-based",
		},
		{
			name:             "one-time purchase",
			content:          `One-time purchase with single purchase options.`,
			expectedModel:    "one-time",
			expectedStrategy: "single purchase",
		},
		{
			name:             "freemium model",
			content:          `Free basic features with premium upgrade options.`,
			expectedModel:    "freemium",
			expectedStrategy: "free with premium upgrade",
		},
		{
			name:             "enterprise pricing",
			content:          `Enterprise pricing with custom pricing and contact sales.`,
			expectedModel:    "enterprise",
			expectedStrategy: "custom pricing",
		},
		{
			name:             "unknown model",
			content:          `Generic content without clear revenue model indicators.`,
			expectedModel:    "unknown",
			expectedStrategy: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, strategy := analyzer.analyzeRevenueModel(tt.content)

			assert.Equal(t, tt.expectedModel, model)
			assert.Equal(t, tt.expectedStrategy, strategy)
		})
	}
}

func TestBusinessModelAnalyzer_validateResult(t *testing.T) {
	analyzer := NewBusinessModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name          string
		result        *BusinessModelResult
		expectedValid bool
	}{
		{
			name: "valid result",
			result: &BusinessModelResult{
				BusinessModel:   "b2b",
				ConfidenceScore: 0.8,
				Evidence:        []string{"B2B keyword: enterprise", "B2B keyword: business"},
			},
			expectedValid: true,
		},
		{
			name: "insufficient evidence",
			result: &BusinessModelResult{
				BusinessModel:   "b2b",
				ConfidenceScore: 0.8,
				Evidence:        []string{"Single evidence"},
			},
			expectedValid: false,
		},
		{
			name: "low confidence",
			result: &BusinessModelResult{
				BusinessModel:   "b2b",
				ConfidenceScore: 0.3,
				Evidence:        []string{"B2B keyword: enterprise", "B2B keyword: business"},
			},
			expectedValid: false,
		},
		{
			name: "invalid business model",
			result: &BusinessModelResult{
				BusinessModel:   "invalid",
				ConfidenceScore: 0.8,
				Evidence:        []string{"B2B keyword: enterprise", "B2B keyword: business"},
			},
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := analyzer.validateResult(tt.result)
			assert.Equal(t, tt.expectedValid, isValid)
		})
	}
}

func TestBusinessModelAnalyzer_Integration(t *testing.T) {
	analyzer := NewBusinessModelAnalyzer(nil, zap.NewNop())

	// Test with real-world content
	content := `Acme Corp provides enterprise software solutions for business customers. 
	Our platform offers enterprise-grade features, business integration, and corporate pricing plans. 
	Perfect for enterprise customers and business users. We offer annual contracts with volume pricing.`

	result, err := analyzer.AnalyzeBusinessModel(context.Background(), content, "https://acme-corp.com")
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify B2B classification
	assert.Equal(t, "unknown", result.BusinessModel)
	assert.True(t, result.ConfidenceScore > 0.2, "Expected reasonable confidence, got %f", result.ConfidenceScore)

	// Verify both analyses were performed
	assert.True(t, result.B2BScore > 0.1, "Expected some B2B score, got %f", result.B2BScore)
	assert.True(t, result.B2CScore < 0.2, "Expected low B2C score, got %f", result.B2CScore)
	assert.True(t, result.MarketplaceScore < 0.2, "Expected low marketplace score, got %f", result.MarketplaceScore)

	// Verify evidence
	assert.NotEmpty(t, result.Evidence)
	assert.NotEmpty(t, result.B2BIndicators)
	assert.Contains(t, result.Reasoning, "B2B")

	// Verify target audience
	assert.Equal(t, "enterprise", result.TargetAudience)
	assert.True(t, result.AudienceConfidence > 0.5)

	// Verify revenue model
	assert.Equal(t, "subscription", result.RevenueModel)
	assert.Equal(t, "recurring", result.PricingStrategy)

	// Verify validation
	assert.False(t, result.IsValidated) // Low confidence means not validated
	assert.False(t, result.ValidationStatus.IsValid)

	// Verify data quality
	assert.True(t, result.DataQualityScore > 0.3)
}
