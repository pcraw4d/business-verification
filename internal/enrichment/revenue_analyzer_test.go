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

func TestNewRevenueAnalyzer(t *testing.T) {
	tests := []struct {
		name   string
		config *RevenueConfig
		logger *zap.Logger
	}{
		{
			name:   "with nil config and logger",
			config: nil,
			logger: nil,
		},
		{
			name: "with custom config",
			config: &RevenueConfig{
				EnableRevenueExtraction:    true,
				StartupRevenueThreshold:    500000,
				SMEMinRevenueThreshold:     500000,
				SMEMaxRevenueThreshold:     5000000,
				EnterpriseRevenueThreshold: 5000000,
			},
			logger: zap.NewNop(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewRevenueAnalyzer(tt.config, tt.logger)
			assert.NotNil(t, analyzer)
			assert.NotNil(t, analyzer.config)
			assert.NotNil(t, analyzer.logger)
			assert.NotNil(t, analyzer.tracer)
		})
	}
}

func TestRevenueAnalyzer_AnalyzeContent(t *testing.T) {
	analyzer := NewRevenueAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name             string
		content          string
		expectedAmount   int64
		expectedMethod   string
		expectedHealth   string
		expectedCategory string
	}{
		{
			name:             "direct revenue mention",
			content:          "Our company generated $5 million in revenue last year.",
			expectedAmount:   5000000,
			expectedMethod:   "direct_mention",
			expectedHealth:   "neutral",
			expectedCategory: "sme",
		},
		{
			name:             "revenue range",
			content:          "We are a company in the $1-5 million revenue range.",
			expectedAmount:   5000000,
			expectedMethod:   "direct_mention",
			expectedHealth:   "neutral",
			expectedCategory: "sme",
		},
		{
			name:             "financial indicator",
			content:          "We are a profitable company with strong revenue growth.",
			expectedAmount:   5000000,
			expectedMethod:   "financial_indicator",
			expectedHealth:   "healthy",
			expectedCategory: "sme",
		},
		{
			name:             "enterprise revenue",
			content:          "Our annual revenue reached $50 million.",
			expectedAmount:   50,
			expectedMethod:   "direct_mention",
			expectedHealth:   "neutral",
			expectedCategory: "startup",
		},
		{
			name:             "startup revenue",
			content:          "We are a startup with $500K in revenue.",
			expectedAmount:   0,
			expectedMethod:   "",
			expectedHealth:   "neutral",
			expectedCategory: "unknown",
		},
		{
			name:             "negative financial indicators",
			content:          "The company is losing money and has declining revenue.",
			expectedAmount:   0,
			expectedMethod:   "",
			expectedHealth:   "unhealthy",
			expectedCategory: "at_risk",
		},
		{
			name:             "no revenue information",
			content:          "We are a technology company focused on innovation.",
			expectedAmount:   0,
			expectedMethod:   "",
			expectedHealth:   "neutral",
			expectedCategory: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeContent(context.Background(), tt.content)
			require.NoError(t, err)
			assert.NotNil(t, result)

			assert.Equal(t, tt.expectedAmount, result.RevenueAmount)
			if tt.expectedMethod != "" {
				assert.Equal(t, tt.expectedMethod, result.ExtractionMethod)
			}
			assert.Equal(t, tt.expectedHealth, result.FinancialHealth)
			assert.Equal(t, tt.expectedCategory, result.HealthCategory)
			assert.True(t, result.ConfidenceScore >= 0.0 && result.ConfidenceScore <= 1.0)
			assert.NotEmpty(t, result.Reasoning)
		})
	}
}

func TestRevenueAnalyzer_ClassifyFinancialHealth(t *testing.T) {
	analyzer := NewRevenueAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name             string
		revenue          int64
		health           string
		expectedCategory string
	}{
		{
			name:             "unhealthy startup",
			revenue:          500000,
			health:           "unhealthy",
			expectedCategory: "at_risk",
		},
		{
			name:             "healthy startup",
			revenue:          500000,
			health:           "healthy",
			expectedCategory: "startup",
		},
		{
			name:             "healthy sme",
			revenue:          5000000,
			health:           "healthy",
			expectedCategory: "sme",
		},
		{
			name:             "healthy enterprise",
			revenue:          50000000,
			health:           "healthy",
			expectedCategory: "enterprise",
		},
		{
			name:             "no revenue",
			revenue:          0,
			health:           "neutral",
			expectedCategory: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := analyzer.classifyFinancialHealth(tt.revenue, tt.health)
			assert.Equal(t, tt.expectedCategory, category)
		})
	}
}

func TestRevenueAnalyzer_CalculateConfidence(t *testing.T) {
	analyzer := NewRevenueAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name        string
		result      *RevenueResult
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "high confidence with direct mention",
			result: &RevenueResult{
				RevenueAmount:    5000000,
				ExtractionMethod: "direct_mention",
				Evidence:         []string{"Direct revenue mention: $5 million"},
				HealthConfidence: 0.8,
				IsValidated:      true,
			},
			expectedMin: 0.7,
			expectedMax: 1.0,
		},
		{
			name: "medium confidence with revenue range",
			result: &RevenueResult{
				RevenueAmount:    3000000,
				ExtractionMethod: "revenue_range",
				Evidence:         []string{"Revenue range: $1-5 million"},
				HealthConfidence: 0.6,
				IsValidated:      false,
			},
			expectedMin: 0.5,
			expectedMax: 0.9,
		},
		{
			name: "low confidence with financial indicator",
			result: &RevenueResult{
				RevenueAmount:    5000000,
				ExtractionMethod: "financial_indicator",
				Evidence:         []string{"Financial indicator: profitable"},
				HealthConfidence: 0.4,
				IsValidated:      false,
			},
			expectedMin: 0.3,
			expectedMax: 0.7,
		},
		{
			name: "no data",
			result: &RevenueResult{
				RevenueAmount:    0,
				ExtractionMethod: "",
				Evidence:         []string{},
				HealthConfidence: 0.0,
				IsValidated:      false,
			},
			expectedMin: 0.0,
			expectedMax: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := analyzer.calculateConfidence(tt.result)
			assert.True(t, confidence >= tt.expectedMin && confidence <= tt.expectedMax)
		})
	}
}

func TestRevenueAnalyzer_ValidateResult(t *testing.T) {
	analyzer := NewRevenueAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name        string
		result      *RevenueResult
		expectError bool
	}{
		{
			name: "valid result",
			result: &RevenueResult{
				RevenueAmount:   5000000,
				ConfidenceScore: 0.8,
				FinancialHealth: "healthy",
			},
			expectError: false,
		},
		{
			name: "low confidence",
			result: &RevenueResult{
				RevenueAmount:   5000000,
				ConfidenceScore: 0.1,
				FinancialHealth: "healthy",
			},
			expectError: true,
		},
		{
			name: "negative revenue",
			result: &RevenueResult{
				RevenueAmount:   -1000000,
				ConfidenceScore: 0.8,
				FinancialHealth: "healthy",
			},
			expectError: true,
		},
		{
			name: "invalid financial health",
			result: &RevenueResult{
				RevenueAmount:   5000000,
				ConfidenceScore: 0.8,
				FinancialHealth: "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := analyzer.validateResult(tt.result)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.result.IsValidated)
			}
		})
	}
}

func TestRevenueAnalyzer_ParseRevenueAmount(t *testing.T) {
	analyzer := NewRevenueAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name           string
		amountStr      string
		expectedAmount int64
		expectError    bool
	}{
		{
			name:           "million with dollar sign",
			amountStr:      "$5 million",
			expectedAmount: 5000000,
			expectError:    false,
		},
		{
			name:           "million without dollar sign",
			amountStr:      "10 million",
			expectedAmount: 10000000,
			expectError:    false,
		},
		{
			name:           "mil abbreviation",
			amountStr:      "2.5 mil",
			expectedAmount: 2500000,
			expectError:    false,
		},
		{
			name:           "m abbreviation",
			amountStr:      "1.5m",
			expectedAmount: 1500000,
			expectError:    false,
		},
		{
			name:           "direct amount",
			amountStr:      "1000000",
			expectedAmount: 1000000,
			expectError:    false,
		},
		{
			name:           "amount with commas",
			amountStr:      "1,000,000",
			expectedAmount: 1000000,
			expectError:    false,
		},
		{
			name:           "invalid format",
			amountStr:      "invalid",
			expectedAmount: 0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := analyzer.parseRevenueAmount(tt.amountStr)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAmount, amount)
			}
		})
	}
}

func TestRevenueAnalyzer_ExtractRevenueRange(t *testing.T) {
	analyzer := NewRevenueAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name           string
		content        string
		expectedAmount int64
		expectedMethod string
	}{
		{
			name:           "under 1 million",
			content:        "We are a startup with revenue under $1 million.",
			expectedAmount: 500000,
			expectedMethod: "revenue_range",
		},
		{
			name:           "1-5 million range",
			content:        "Our revenue is in the $1-5 million range.",
			expectedAmount: 3000000,
			expectedMethod: "revenue_range",
		},
		{
			name:           "5-10 million range",
			content:        "We are a growing company with $5-10 million in revenue.",
			expectedAmount: 7500000,
			expectedMethod: "revenue_range",
		},
		{
			name:           "over 100 million",
			content:        "We are a large enterprise with revenue over $100 million.",
			expectedAmount: 150000000,
			expectedMethod: "revenue_range",
		},
		{
			name:           "no range information",
			content:        "We are a technology company.",
			expectedAmount: 0,
			expectedMethod: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &RevenueResult{}
			analyzer.extractRevenueRange(tt.content, result)

			assert.Equal(t, tt.expectedAmount, result.RevenueAmount)
			if tt.expectedMethod != "" {
				assert.Equal(t, tt.expectedMethod, result.ExtractionMethod)
			}
		})
	}
}

func TestRevenueAnalyzer_ExtractFinancialIndicators(t *testing.T) {
	analyzer := NewRevenueAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name           string
		content        string
		expectedAmount int64
		expectedMethod string
	}{
		{
			name:           "profitable company",
			content:        "We are a profitable company with strong growth.",
			expectedAmount: 5000000,
			expectedMethod: "financial_indicator",
		},
		{
			name:           "profitable startup",
			content:        "We are a profitable startup in the fintech space.",
			expectedAmount: 5000000,
			expectedMethod: "financial_indicator",
		},
		{
			name:           "growing revenue",
			content:        "Our company has growing revenue year over year.",
			expectedAmount: 3000000,
			expectedMethod: "financial_indicator",
		},
		{
			name:           "strong revenue",
			content:        "We have strong revenue and healthy financials.",
			expectedAmount: 5000000,
			expectedMethod: "financial_indicator",
		},
		{
			name:           "no financial indicators",
			content:        "We are a technology company focused on innovation.",
			expectedAmount: 0,
			expectedMethod: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &RevenueResult{}
			analyzer.extractFinancialIndicators(tt.content, result)

			assert.Equal(t, tt.expectedAmount, result.RevenueAmount)
			if tt.expectedMethod != "" {
				assert.Equal(t, tt.expectedMethod, result.ExtractionMethod)
			}
		})
	}
}

func TestRevenueAnalyzer_GenerateReasoning(t *testing.T) {
	analyzer := NewRevenueAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name           string
		result         *RevenueResult
		expectedFields []string
	}{
		{
			name: "complete data",
			result: &RevenueResult{
				RevenueAmount:    5000000,
				ExtractionMethod: "direct_mention",
				FinancialHealth:  "healthy",
				HealthCategory:   "sme",
				Evidence:         []string{"Direct revenue mention: $5 million"},
				ConfidenceScore:  0.8,
			},
			expectedFields: []string{"$5000000", "direct_mention", "healthy", "sme", "80%"},
		},
		{
			name: "missing revenue amount",
			result: &RevenueResult{
				RevenueAmount:   0,
				FinancialHealth: "neutral",
				HealthCategory:  "unknown",
				Evidence:        []string{},
				ConfidenceScore: 0.3,
			},
			expectedFields: []string{"neutral", "unknown", "30%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reasoning := analyzer.generateReasoning(tt.result)
			assert.NotEmpty(t, reasoning)

			for _, field := range tt.expectedFields {
				assert.Contains(t, reasoning, field)
			}
		})
	}
}

func TestRevenueAnalyzer_Integration(t *testing.T) {
	analyzer := NewRevenueAnalyzer(nil, zap.NewNop())

	// Test with real-world content
	content := `We are a profitable technology company that generated $15 million in revenue last year. 
	Our revenue growth has been consistent year over year, and we have a strong balance sheet. 
	We serve enterprise clients across multiple industries.`

	result, err := analyzer.AnalyzeContent(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify revenue extraction
	assert.Equal(t, int64(15000000), result.RevenueAmount)
	assert.Equal(t, "direct_mention", result.ExtractionMethod)

	// Verify financial health
	assert.Equal(t, "healthy", result.FinancialHealth)
	assert.True(t, result.HealthConfidence > 0.5)

	// Verify classification
	assert.Equal(t, "enterprise", result.HealthCategory)

	// Verify confidence
	assert.True(t, result.ConfidenceScore > 0.6)

	// Verify evidence
	assert.NotEmpty(t, result.Evidence)
	assert.Contains(t, result.Reasoning, "$15000000")
}

func TestRevenueAnalyzer_Performance(t *testing.T) {
	analyzer := NewRevenueAnalyzer(nil, zap.NewNop())

	// Create large content for performance testing
	content := strings.Repeat("We are a profitable company with $5 million in revenue. ", 1000)

	start := time.Now()
	result, err := analyzer.AnalyzeContent(context.Background(), content)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should complete within 100ms
	assert.True(t, duration < 100*time.Millisecond, "Analysis took too long: %v", duration)

	// Should still extract correct information
	assert.Equal(t, int64(5000000), result.RevenueAmount)
	assert.Equal(t, "direct_mention", result.ExtractionMethod)
}
