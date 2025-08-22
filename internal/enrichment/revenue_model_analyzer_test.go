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

func TestNewRevenueModelAnalyzer(t *testing.T) {
	config := &RevenueModelConfig{
		MinConfidenceThreshold: 0.5,
		ModelWeight:            0.4,
	}
	logger := zap.NewNop()

	analyzer := NewRevenueModelAnalyzer(config, logger)

	assert.NotNil(t, analyzer)
	assert.Equal(t, config, analyzer.config)
	assert.Equal(t, logger, analyzer.logger)
	assert.NotNil(t, analyzer.tracer)
}

func TestGetDefaultRevenueModelConfig(t *testing.T) {
	config := GetDefaultRevenueModelConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 0.3, config.MinConfidenceThreshold)
	assert.Equal(t, 2, config.MinEvidenceCount)
	assert.Equal(t, 50, config.MinContentLength)
	assert.Equal(t, 0.30, config.ModelWeight)
	assert.Equal(t, 0.25, config.PricingWeight)
	assert.Equal(t, 0.20, config.StrategyWeight)
	assert.Equal(t, 0.15, config.MarketWeight)
	assert.Equal(t, 0.10, config.CompetitiveWeight)
	assert.True(t, config.RequireMultipleIndicators)
	assert.True(t, config.EnableFallbackAnalysis)
	assert.True(t, config.ValidateModels)
}

func TestRevenueModelAnalyzer_AnalyzeRevenueModel(t *testing.T) {
	analyzer := NewRevenueModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name                      string
		content                   string
		sourceURL                 string
		expectedPrimaryModel      string
		expectedMinConfidence     float64
		expectedMaxConfidence     float64
		expectedPricingStrategies int
		expectedRevenueStreams    int
	}{
		{
			name: "SaaS subscription model",
			content: `We offer a comprehensive SaaS platform with monthly and annual subscription plans. 
			Our pricing starts at $29/month for the basic plan, $99/month for professional, and $299/month for enterprise. 
			All plans include 24/7 support and regular updates.`,
			sourceURL:                 "https://saas-platform.com",
			expectedPrimaryModel:      "subscription",
			expectedMinConfidence:     0.6,
			expectedMaxConfidence:     0.8,
			expectedPricingStrategies: 1,
			expectedRevenueStreams:    1,
		},
		{
			name: "Marketplace with transaction fees",
			content: `Our marketplace connects buyers and sellers. We charge a 5% transaction fee on all sales. 
			Premium sellers pay a monthly subscription of $49 for enhanced features and priority support. 
			We also offer advertising space for featured listings.`,
			sourceURL:                 "https://marketplace.com",
			expectedPrimaryModel:      "marketplace",
			expectedMinConfidence:     0.6,
			expectedMaxConfidence:     0.8,
			expectedPricingStrategies: 1,
			expectedRevenueStreams:    2,
		},
		{
			name: "Freemium model",
			content: `Our app is free to download and use with basic features. Premium features are available 
			through in-app purchases starting at $4.99. Pro subscription costs $9.99/month and includes 
			all premium features plus cloud sync.`,
			sourceURL:                 "https://freemium-app.com",
			expectedPrimaryModel:      "freemium",
			expectedMinConfidence:     0.6,
			expectedMaxConfidence:     0.8,
			expectedPricingStrategies: 1,
			expectedRevenueStreams:    1,
		},
		{
			name: "Enterprise software licensing",
			content: `Our enterprise software is licensed on a per-seat basis. Annual licenses start at $500 per user. 
			We offer volume discounts for large deployments. Custom enterprise solutions are available 
			with dedicated support and implementation services.`,
			sourceURL:                 "https://enterprise-software.com",
			expectedPrimaryModel:      "enterprise",
			expectedMinConfidence:     0.6,
			expectedMaxConfidence:     0.8,
			expectedPricingStrategies: 1,
			expectedRevenueStreams:    2,
		},
		{
			name: "Advertising-based model",
			content: `Our platform is completely free to use. We generate revenue through targeted advertising 
			and sponsored content. Advertisers can reach our audience of over 1 million active users. 
			We also offer premium ad placements and sponsored posts.`,
			sourceURL:                 "https://ad-platform.com",
			expectedPrimaryModel:      "advertising",
			expectedMinConfidence:     0.6,
			expectedMaxConfidence:     0.8,
			expectedPricingStrategies: 1,
			expectedRevenueStreams:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeRevenueModel(context.Background(), tt.content, tt.sourceURL)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.sourceURL, result.SourceURL)
			assert.NotZero(t, result.AnalyzedAt)
			assert.NotZero(t, result.ProcessingTime)

			// Check primary revenue model
			if tt.expectedPrimaryModel != "" {
				assert.Contains(t, strings.ToLower(result.PrimaryRevenueModel), strings.ToLower(tt.expectedPrimaryModel))
			}

			// Check confidence score
			assert.GreaterOrEqual(t, result.ConfidenceScore, tt.expectedMinConfidence)
			assert.LessOrEqual(t, result.ConfidenceScore, tt.expectedMaxConfidence)

			// Check pricing strategies
			assert.Len(t, result.PricingStrategies, tt.expectedPricingStrategies)

			// Check revenue streams
			assert.Len(t, result.RevenueStreams, tt.expectedRevenueStreams)

			// Check validation
			assert.NotNil(t, result.ValidationStatus)
			assert.NotZero(t, result.DataQualityScore)
			assert.NotEmpty(t, result.Reasoning)
		})
	}
}

func TestRevenueModelAnalyzer_analyzeRevenueModels(t *testing.T) {
	analyzer := NewRevenueModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name           string
		content        string
		expectedModels []string
		minConfidence  float64
	}{
		{
			name:           "Subscription indicators",
			content:        `Monthly subscription, annual billing, recurring revenue, SaaS pricing`,
			expectedModels: []string{"subscription"},
			minConfidence:  0.2,
		},
		{
			name:           "Freemium indicators",
			content:        `Free tier, premium features, upgrade to pro, basic vs premium`,
			expectedModels: []string{"freemium"},
			minConfidence:  0.2,
		},
		{
			name:           "Marketplace indicators",
			content:        `Transaction fees, commission, marketplace, buyer seller platform`,
			expectedModels: []string{"marketplace"},
			minConfidence:  0.2,
		},
		{
			name:           "Enterprise indicators",
			content:        `Enterprise licensing, per-seat pricing, volume discounts, custom solutions`,
			expectedModels: []string{"enterprise"},
			minConfidence:  0.2,
		},
		{
			name:           "Advertising indicators",
			content:        `Ad revenue, sponsored content, advertising platform, free to use`,
			expectedModels: []string{"advertising"},
			minConfidence:  0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &RevenueModelResult{
				SourceURL:        "test.com",
				AnalyzedAt:       time.Now(),
				Evidence:         []string{},
				ExtractedPhrases: []string{},
				ComponentScores:  RevenueComponentScores{},
			}

			err := analyzer.analyzeRevenueModels(context.Background(), tt.content, result)

			assert.NoError(t, err)
			assert.NotEmpty(t, result.Evidence)
			assert.NotEmpty(t, result.ExtractedPhrases)

			// Check if expected models are found in evidence or extracted phrases
			for _, expectedModel := range tt.expectedModels {
				found := false
				if strings.Contains(strings.ToLower(result.PrimaryRevenueModel), strings.ToLower(expectedModel)) {
					found = true
				}
				if strings.Contains(strings.ToLower(result.SecondaryRevenueModel), strings.ToLower(expectedModel)) {
					found = true
				}
				// Also check evidence and extracted phrases
				for _, evidence := range result.Evidence {
					if strings.Contains(strings.ToLower(evidence), strings.ToLower(expectedModel)) {
						found = true
						break
					}
				}
				for _, phrase := range result.ExtractedPhrases {
					if strings.Contains(strings.ToLower(phrase), strings.ToLower(expectedModel)) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected model %s not found", expectedModel)
			}
		})
	}
}

func TestRevenueModelAnalyzer_analyzePricingStrategies(t *testing.T) {
	analyzer := NewRevenueModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name               string
		content            string
		expectedStrategies []string
		minConfidence      float64
	}{
		{
			name:               "Tiered pricing",
			content:            `Basic plan $29, Professional $99, Enterprise $299`,
			expectedStrategies: []string{"tiered"},
			minConfidence:      0.2,
		},
		{
			name:               "Value-based pricing",
			content:            `Pricing based on value delivered, ROI-focused pricing`,
			expectedStrategies: []string{"value-based"},
			minConfidence:      0.2,
		},
		{
			name:               "Penetration pricing",
			content:            `Low introductory pricing, market entry strategy`,
			expectedStrategies: []string{"penetration"},
			minConfidence:      0.2,
		},
		{
			name:               "Premium pricing",
			content:            `Premium positioning, luxury pricing, high-end market`,
			expectedStrategies: []string{"premium"},
			minConfidence:      0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &RevenueModelResult{
				SourceURL:        "test.com",
				AnalyzedAt:       time.Now(),
				Evidence:         []string{},
				ExtractedPhrases: []string{},
				ComponentScores:  RevenueComponentScores{},
			}

			err := analyzer.analyzePricingStrategies(context.Background(), tt.content, result)

			assert.NoError(t, err)
			assert.NotEmpty(t, result.PricingStrategies)

			// Check if expected strategies are found
			for _, expectedStrategy := range tt.expectedStrategies {
				found := false
				for _, strategy := range result.PricingStrategies {
					if strings.Contains(strings.ToLower(strategy.Name), strings.ToLower(expectedStrategy)) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected strategy %s not found", expectedStrategy)
			}
		})
	}
}

func TestRevenueModelAnalyzer_analyzeRevenueStreams(t *testing.T) {
	analyzer := NewRevenueModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name            string
		content         string
		expectedStreams []string
		minConfidence   float64
	}{
		{
			name:            "Software licensing",
			content:         `Software licenses, per-user licensing, annual licenses`,
			expectedStreams: []string{"software licensing"},
			minConfidence:   0.2,
		},
		{
			name:            "Transaction fees",
			content:         `5% transaction fee, commission on sales, processing fees`,
			expectedStreams: []string{"transaction fees"},
			minConfidence:   0.2,
		},
		{
			name:            "Advertising revenue",
			content:         `Ad revenue, sponsored content, advertising space`,
			expectedStreams: []string{"advertising"},
			minConfidence:   0.2,
		},
		{
			name:            "Data monetization",
			content:         `Data insights, analytics services, market intelligence`,
			expectedStreams: []string{"data monetization"},
			minConfidence:   0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &RevenueModelResult{
				SourceURL:        "test.com",
				AnalyzedAt:       time.Now(),
				Evidence:         []string{},
				ExtractedPhrases: []string{},
				ComponentScores:  RevenueComponentScores{},
			}

			err := analyzer.analyzeRevenueStreams(context.Background(), tt.content, result)

			assert.NoError(t, err)
			assert.NotEmpty(t, result.RevenueStreams)

			// Check if expected streams are found
			for _, expectedStream := range tt.expectedStreams {
				found := false
				for _, stream := range result.RevenueStreams {
					if strings.Contains(strings.ToLower(stream.Name), strings.ToLower(expectedStream)) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected stream %s not found", expectedStream)
			}
		})
	}
}

func TestRevenueModelAnalyzer_analyzeMarketPositioning(t *testing.T) {
	analyzer := NewRevenueModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name                  string
		content               string
		expectedMarketSegment string
		expectedAdvantages    []string
		minConfidence         float64
	}{
		{
			name:                  "Enterprise segment",
			content:               `Enterprise solutions, Fortune 500 clients, large corporations`,
			expectedMarketSegment: "enterprise",
			expectedAdvantages:    []string{"enterprise-grade"},
			minConfidence:         0.2,
		},
		{
			name:                  "SMB segment",
			content:               `Small business solutions, affordable pricing, SMB market`,
			expectedMarketSegment: "smb",
			expectedAdvantages:    []string{"affordable"},
			minConfidence:         0.2,
		},
		{
			name:                  "Consumer segment",
			content:               `Consumer app, individual users, personal use`,
			expectedMarketSegment: "consumer",
			expectedAdvantages:    []string{"user-friendly"},
			minConfidence:         0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &RevenueModelResult{
				SourceURL:        "test.com",
				AnalyzedAt:       time.Now(),
				Evidence:         []string{},
				ExtractedPhrases: []string{},
				ComponentScores:  RevenueComponentScores{},
			}

			err := analyzer.analyzeMarketPositioning(context.Background(), tt.content, result)

			assert.NoError(t, err)
			assert.NotEmpty(t, result.MarketPositioning.MarketSegment)

			// Check market segment
			if tt.expectedMarketSegment != "" {
				assert.Contains(t, strings.ToLower(result.MarketPositioning.MarketSegment), strings.ToLower(tt.expectedMarketSegment))
			}

			// Check competitive advantages
			for _, expectedAdvantage := range tt.expectedAdvantages {
				found := false
				for _, advantage := range result.MarketPositioning.CompetitiveAdvantage {
					if strings.Contains(strings.ToLower(advantage), strings.ToLower(expectedAdvantage)) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected advantage %s not found", expectedAdvantage)
			}
		})
	}
}

func TestRevenueModelAnalyzer_analyzeCompetitiveLandscape(t *testing.T) {
	analyzer := NewRevenueModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name              string
		content           string
		expectedLandscape string
		expectedGaps      []string
		minConfidence     float64
	}{
		{
			name:              "Competitive market",
			content:           `Competitive landscape, market competition, industry rivals`,
			expectedLandscape: "competitive",
			expectedGaps:      []string{"market gap"},
			minConfidence:     0.2,
		},
		{
			name:              "Emerging market",
			content:           `Emerging market, new technology, innovative solution`,
			expectedLandscape: "emerging",
			expectedGaps:      []string{"innovation"},
			minConfidence:     0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &RevenueModelResult{
				SourceURL:        "test.com",
				AnalyzedAt:       time.Now(),
				Evidence:         []string{},
				ExtractedPhrases: []string{},
				ComponentScores:  RevenueComponentScores{},
			}

			err := analyzer.analyzeCompetitiveLandscape(context.Background(), tt.content, result)

			assert.NoError(t, err)
			assert.NotEmpty(t, result.CompetitiveAnalysis.CompetitiveLandscape)

			// Check competitive landscape
			if tt.expectedLandscape != "" {
				assert.Contains(t, strings.ToLower(result.CompetitiveAnalysis.CompetitiveLandscape), strings.ToLower(tt.expectedLandscape))
			}

			// Check market gaps
			for _, expectedGap := range tt.expectedGaps {
				found := false
				for _, gap := range result.CompetitiveAnalysis.MarketGaps {
					if strings.Contains(strings.ToLower(gap), strings.ToLower(expectedGap)) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected gap %s not found", expectedGap)
			}
		})
	}
}

func TestRevenueModelAnalyzer_determinePrimaryRevenueModel(t *testing.T) {
	analyzer := NewRevenueModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name                   string
		revenueModels          []string
		expectedPrimaryModel   string
		expectedSecondaryModel string
	}{
		{
			name:                   "Single model",
			revenueModels:          []string{"subscription"},
			expectedPrimaryModel:   "subscription",
			expectedSecondaryModel: "",
		},
		{
			name:                   "Multiple models",
			revenueModels:          []string{"subscription", "freemium", "enterprise"},
			expectedPrimaryModel:   "subscription",
			expectedSecondaryModel: "freemium",
		},
		{
			name:                   "No models",
			revenueModels:          []string{},
			expectedPrimaryModel:   "",
			expectedSecondaryModel: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &RevenueModelResult{
				SourceURL:        "test.com",
				AnalyzedAt:       time.Now(),
				Evidence:         []string{},
				ExtractedPhrases: []string{},
				ComponentScores:  RevenueComponentScores{},
			}

			// Simulate revenue models found
			for _, model := range tt.revenueModels {
				result.Evidence = append(result.Evidence, "Found "+model+" model")
			}

			analyzer.determinePrimaryRevenueModel(result)

			assert.Equal(t, tt.expectedPrimaryModel, result.PrimaryRevenueModel)
			assert.Equal(t, tt.expectedSecondaryModel, result.SecondaryRevenueModel)
		})
	}
}

func TestRevenueModelAnalyzer_calculateConfidenceScores(t *testing.T) {
	analyzer := NewRevenueModelAnalyzer(nil, zap.NewNop())

	result := &RevenueModelResult{
		SourceURL:        "test.com",
		AnalyzedAt:       time.Now(),
		Evidence:         []string{"Evidence 1", "Evidence 2"},
		ExtractedPhrases: []string{"Phrase 1", "Phrase 2"},
		ComponentScores:  RevenueComponentScores{},
		PricingStrategies: []PricingStrategy{
			{Name: "Tiered Pricing", ConfidenceScore: 0.8},
		},
		RevenueStreams: []RevenueStream{
			{Name: "Subscription", ConfidenceScore: 0.9},
		},
	}

	analyzer.calculateConfidenceScores(result)

	// Check component scores
	assert.Greater(t, result.ComponentScores.ModelScore, 0.0)
	assert.Greater(t, result.ComponentScores.PricingScore, 0.0)
	assert.Greater(t, result.ComponentScores.StrategyScore, 0.0)
	assert.Greater(t, result.ComponentScores.MarketScore, 0.0)
	assert.Greater(t, result.ComponentScores.CompetitiveScore, 0.0)

	// Check overall confidence
	assert.Greater(t, result.ConfidenceScore, 0.0)
	assert.LessOrEqual(t, result.ConfidenceScore, 1.0)
}

func TestRevenueModelAnalyzer_validateResult(t *testing.T) {
	analyzer := NewRevenueModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name          string
		result        *RevenueModelResult
		expectedValid bool
	}{
		{
			name: "Valid result",
			result: &RevenueModelResult{
				PrimaryRevenueModel: "subscription",
				ConfidenceScore:     0.8,
				Evidence:            []string{"Evidence 1", "Evidence 2"},
				PricingStrategies:   []PricingStrategy{{Name: "Tiered"}},
			},
			expectedValid: true,
		},
		{
			name: "Invalid - no primary model",
			result: &RevenueModelResult{
				PrimaryRevenueModel: "",
				ConfidenceScore:     0.8,
				Evidence:            []string{"Evidence 1"},
			},
			expectedValid: false,
		},
		{
			name: "Invalid - low confidence",
			result: &RevenueModelResult{
				PrimaryRevenueModel: "subscription",
				ConfidenceScore:     0.1,
				Evidence:            []string{"Evidence 1"},
			},
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer.validateResult(tt.result)

			assert.Equal(t, tt.expectedValid, tt.result.ValidationStatus.IsValid)
			assert.NotNil(t, tt.result.ValidationStatus.LastValidated)
		})
	}
}

func TestRevenueModelAnalyzer_calculateDataQuality(t *testing.T) {
	analyzer := NewRevenueModelAnalyzer(nil, zap.NewNop())

	result := &RevenueModelResult{
		SourceURL:        "test.com",
		AnalyzedAt:       time.Now(),
		Evidence:         []string{"Evidence 1", "Evidence 2", "Evidence 3"},
		ExtractedPhrases: []string{"Phrase 1", "Phrase 2"},
		PricingStrategies: []PricingStrategy{
			{Name: "Strategy 1", ConfidenceScore: 0.8},
			{Name: "Strategy 2", ConfidenceScore: 0.9},
		},
		RevenueStreams: []RevenueStream{
			{Name: "Stream 1", ConfidenceScore: 0.7},
		},
		ModelDetails: RevenueModelDetails{
			ModelType:        "subscription",
			RevenueSources:   []string{"Source 1", "Source 2"},
			ValueProposition: "Great value",
		},
	}

	qualityScore := analyzer.calculateDataQuality(result)

	assert.Greater(t, qualityScore, 0.0)
	assert.LessOrEqual(t, qualityScore, 1.0)
}

func TestRevenueModelAnalyzer_generateReasoning(t *testing.T) {
	analyzer := NewRevenueModelAnalyzer(nil, zap.NewNop())

	result := &RevenueModelResult{
		PrimaryRevenueModel: "subscription",
		ConfidenceScore:     0.85,
		Evidence:            []string{"Evidence 1", "Evidence 2"},
		PricingStrategies:   []PricingStrategy{{Name: "Tiered Pricing"}},
		RevenueStreams:      []RevenueStream{{Name: "Monthly Subscriptions"}},
		MarketPositioning: MarketPositioning{
			MarketSegment:        "SMB",
			CompetitiveAdvantage: []string{"Affordable pricing"},
		},
	}

	reasoning := analyzer.generateReasoning(result)

	assert.NotEmpty(t, reasoning)
	assert.Contains(t, reasoning, "subscription")
	assert.Contains(t, reasoning, "85.0%")
	assert.Contains(t, reasoning, "Tiered Pricing")
	assert.Contains(t, reasoning, "Monthly Subscriptions")
	assert.Contains(t, reasoning, "SMB")
	assert.Contains(t, reasoning, "Affordable pricing")
	assert.Contains(t, reasoning, "2 pieces of evidence")
}

func TestRevenueModelAnalyzer_Integration(t *testing.T) {
	analyzer := NewRevenueModelAnalyzer(nil, zap.NewNop())

	content := `Our SaaS platform offers comprehensive business solutions with a tiered pricing model. 
	Basic plan costs $29/month, Professional plan is $99/month, and Enterprise plan is $299/month. 
	We also offer annual billing with 20% discount. Our platform serves small to medium businesses 
	with enterprise-grade features. We compete with traditional software vendors by offering 
	cloud-based solutions with automatic updates and 24/7 support.`

	result, err := analyzer.AnalyzeRevenueModel(context.Background(), content, "https://test-saas.com")

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Check primary model
	assert.Contains(t, strings.ToLower(result.PrimaryRevenueModel), "subscription")

	// Check pricing strategies
	assert.NotEmpty(t, result.PricingStrategies)
	foundTiered := false
	for _, strategy := range result.PricingStrategies {
		if strings.Contains(strings.ToLower(strategy.Name), "tiered") {
			foundTiered = true
			break
		}
	}
	assert.True(t, foundTiered, "Tiered pricing strategy not found")

	// Check revenue streams
	assert.NotEmpty(t, result.RevenueStreams)

	// Check market positioning
	assert.NotEmpty(t, result.MarketPositioning.MarketSegment)
	assert.Contains(t, strings.ToLower(result.MarketPositioning.MarketSegment), "smb")

	// Check competitive analysis
	assert.NotEmpty(t, result.CompetitiveAnalysis.CompetitiveLandscape)

	// Check validation
	assert.True(t, result.IsValidated)
	assert.Greater(t, result.DataQualityScore, 0.0)
	assert.NotEmpty(t, result.Reasoning)
}

func TestRevenueModelAnalyzer_Performance(t *testing.T) {
	analyzer := NewRevenueModelAnalyzer(nil, zap.NewNop())

	// Create large content for performance testing
	content := strings.Repeat(`Our SaaS platform offers comprehensive business solutions with subscription pricing. 
	We serve enterprise customers with advanced features and dedicated support. `, 100)

	start := time.Now()
	result, err := analyzer.AnalyzeRevenueModel(context.Background(), content, "https://performance-test.com")
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Performance assertion - should complete within reasonable time
	assert.Less(t, duration, 5*time.Second, "Analysis took too long: %v", duration)

	// Check that processing time is recorded
	assert.Greater(t, result.ProcessingTime, time.Duration(0))
	assert.LessOrEqual(t, result.ProcessingTime, duration)
}

func TestRevenueModelAnalyzer_ErrorHandling(t *testing.T) {
	analyzer := NewRevenueModelAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name    string
		content string
		url     string
	}{
		{
			name:    "Very short content",
			content: "Short content that is too brief for meaningful analysis",
			url:     "https://test.com",
		},
		{
			name:    "Very short content",
			content: "Short",
			url:     "https://test.com",
		},
		{
			name:    "No revenue indicators",
			content: "This is just some random content without any revenue model indicators.",
			url:     "https://test.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeRevenueModel(context.Background(), tt.content, tt.url)

			// Should not error, but may return low confidence results
			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Should have low confidence for poor inputs
			assert.Less(t, result.ConfidenceScore, 0.5)
			assert.False(t, result.ValidationStatus.IsValid)
		})
	}
}
