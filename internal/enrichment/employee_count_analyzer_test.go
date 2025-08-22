package enrichment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewEmployeeCountAnalyzer(t *testing.T) {
	tests := []struct {
		name   string
		config *EmployeeCountConfig
		logger *zap.Logger
	}{
		{
			name:   "with nil config and logger",
			config: nil,
			logger: nil,
		},
		{
			name: "with custom config",
			config: &EmployeeCountConfig{
				EnableEmployeeExtraction: true,
				StartupThreshold:         25,
				SMEMinThreshold:          26,
				SMEMaxThreshold:          100,
				EnterpriseThreshold:      101,
			},
			logger: zap.NewNop(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewEmployeeCountAnalyzer(tt.config, tt.logger)
			assert.NotNil(t, analyzer)
			assert.NotNil(t, analyzer.config)
			assert.NotNil(t, analyzer.logger)
			assert.NotNil(t, analyzer.tracer)
		})
	}
}

func TestAnalyzeEmployeeCount_DirectMentions(t *testing.T) {
	analyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())
	ctx := context.Background()

	tests := []struct {
		name           string
		content        string
		expectedCount  int
		expectedMethod string
		expectedSize   string
	}{
		{
			name:           "direct employee count",
			content:        "We have 25 employees working on innovative solutions.",
			expectedCount:  25,
			expectedMethod: "direct_mention",
			expectedSize:   "startup",
		},
		{
			name:           "team of employees",
			content:        "Our team of 150 employees is dedicated to customer success.",
			expectedCount:  150,
			expectedMethod: "direct_mention",
			expectedSize:   "sme",
		},
		{
			name:           "staff count",
			content:        "We are a company with 500 staff members worldwide.",
			expectedCount:  500,
			expectedMethod: "direct_mention",
			expectedSize:   "enterprise",
		},
		{
			name:           "large enterprise",
			content:        "Join our team of 5000+ employees across the globe.",
			expectedCount:  5000,
			expectedMethod: "linkedin_style",
			expectedSize:   "enterprise",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeEmployeeCount(ctx, tt.content, "https://example.com")
			require.NoError(t, err)
			assert.NotNil(t, result)

			assert.Equal(t, tt.expectedCount, result.EmployeeCount)
			assert.Equal(t, tt.expectedMethod, result.ExtractionMethod)
			assert.Equal(t, tt.expectedSize, result.CompanySize)
			assert.True(t, result.ConfidenceScore > 0)
			assert.NotEmpty(t, result.Evidence)
			assert.NotEmpty(t, result.Reasoning)
		})
	}
}

func TestAnalyzeEmployeeCount_TeamIndicators(t *testing.T) {
	analyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())
	ctx := context.Background()

	tests := []struct {
		name           string
		content        string
		expectedCount  int
		expectedMethod string
	}{
		{
			name:           "small team",
			content:        "We are a small team passionate about innovation.",
			expectedCount:  5,
			expectedMethod: "team_indicator",
		},
		{
			name:           "startup team",
			content:        "Our startup team is growing rapidly.",
			expectedCount:  10,
			expectedMethod: "team_indicator",
		},
		{
			name:           "cross-functional team",
			content:        "Our cross-functional team delivers exceptional results.",
			expectedCount:  12,
			expectedMethod: "team_indicator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeEmployeeCount(ctx, tt.content, "https://example.com")
			require.NoError(t, err)
			assert.NotNil(t, result)

			assert.Equal(t, tt.expectedCount, result.EmployeeCount)
			assert.Equal(t, tt.expectedMethod, result.ExtractionMethod)
			assert.True(t, result.ConfidenceScore > 0)
		})
	}
}

func TestAnalyzeEmployeeCount_CompanySizeKeywords(t *testing.T) {
	analyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())
	ctx := context.Background()

	tests := []struct {
		name           string
		content        string
		expectedCount  int
		expectedMethod string
	}{
		{
			name:           "startup keyword",
			content:        "We are a startup focused on AI technology.",
			expectedCount:  15,
			expectedMethod: "size_keyword",
		},
		{
			name:           "small business",
			content:        "As a small business, we provide personalized service.",
			expectedCount:  25,
			expectedMethod: "size_keyword",
		},
		{
			name:           "technology company",
			content:        "We are a technology company with global reach.",
			expectedCount:  100,
			expectedMethod: "size_keyword",
		},
		{
			name:           "multinational",
			content:        "We are a multinational corporation with global presence.",
			expectedCount:  2000,
			expectedMethod: "size_keyword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeEmployeeCount(ctx, tt.content, "https://example.com")
			require.NoError(t, err)
			assert.NotNil(t, result)

			assert.Equal(t, tt.expectedCount, result.EmployeeCount)
			assert.Equal(t, tt.expectedMethod, result.ExtractionMethod)
			assert.True(t, result.ConfidenceScore > 0)
		})
	}
}

func TestAnalyzeEmployeeCount_LinkedInStyle(t *testing.T) {
	analyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())
	ctx := context.Background()

	tests := []struct {
		name           string
		content        string
		expectedCount  int
		expectedMethod string
	}{
		{
			name:           "team of professionals",
			content:        "Join our team of 75 professionals worldwide.",
			expectedCount:  75,
			expectedMethod: "direct_mention",
		},
		{
			name:           "we are a team of",
			content:        "We are a team of 200+ people across multiple offices.",
			expectedCount:  200,
			expectedMethod: "linkedin_style",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeEmployeeCount(ctx, tt.content, "https://example.com")
			require.NoError(t, err)
			assert.NotNil(t, result)

			assert.Equal(t, tt.expectedCount, result.EmployeeCount)
			assert.Equal(t, tt.expectedMethod, result.ExtractionMethod)
			assert.True(t, result.ConfidenceScore > 0)
		})
	}
}

func TestAnalyzeEmployeeCount_NoInformation(t *testing.T) {
	analyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())
	ctx := context.Background()

	content := "This is a company that provides services to customers."
	result, err := analyzer.AnalyzeEmployeeCount(ctx, content, "https://example.com")
	require.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, 0, result.EmployeeCount)
	assert.Equal(t, "unknown", result.CompanySize)
	assert.Equal(t, "Unknown", result.EmployeeRange)
	assert.True(t, result.ConfidenceScore < 0.7) // Lower confidence for no information
}

func TestAnalyzeEmployeeCount_ContextCancellation(t *testing.T) {
	analyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	content := "We have 50 employees."
	_, err := analyzer.AnalyzeEmployeeCount(ctx, content, "https://example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
}

func TestClassifyCompanySize(t *testing.T) {
	analyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())
	ctx := context.Background()

	tests := []struct {
		name          string
		employeeCount int
		content       string
		expectedSize  string
		expectedMin   int
		expectedMax   int
	}{
		{
			name:          "startup",
			employeeCount: 25,
			content:       "We are a startup company.",
			expectedSize:  "startup",
			expectedMin:   1,
			expectedMax:   50,
		},
		{
			name:          "sme",
			employeeCount: 150,
			content:       "We are a medium-sized business.",
			expectedSize:  "sme",
			expectedMin:   51,
			expectedMax:   250,
		},
		{
			name:          "mid enterprise",
			employeeCount: 300,
			content:       "We are a large company.",
			expectedSize:  "enterprise",
			expectedMin:   252,
			expectedMax:   100000,
		},
		{
			name:          "enterprise",
			employeeCount: 5000,
			content:       "We are a Fortune 500 company.",
			expectedSize:  "enterprise",
			expectedMin:   252,
			expectedMax:   100000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category, err := analyzer.classifyCompanySize(ctx, tt.employeeCount, tt.content)
			require.NoError(t, err)
			assert.NotNil(t, category)

			assert.Equal(t, tt.expectedSize, category.Category)
			assert.Equal(t, tt.expectedMin, category.MinEmployees)
			assert.Equal(t, tt.expectedMax, category.MaxEmployees)
			assert.True(t, category.ConfidenceScore > 0)
			assert.NotEmpty(t, category.Evidence)
		})
	}
}

func TestGenerateEmployeeRange(t *testing.T) {
	analyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name          string
		employeeCount int
		expectedRange string
	}{
		{name: "zero employees", employeeCount: 0, expectedRange: "Unknown"},
		{name: "small startup", employeeCount: 5, expectedRange: "1-10 employees"},
		{name: "medium startup", employeeCount: 25, expectedRange: "11-50 employees"},
		{name: "sme", employeeCount: 75, expectedRange: "51-100 employees"},
		{name: "large sme", employeeCount: 200, expectedRange: "101-250 employees"},
		{name: "mid enterprise", employeeCount: 400, expectedRange: "251-500 employees"},
		{name: "enterprise", employeeCount: 800, expectedRange: "501-1,000 employees"},
		{name: "large enterprise", employeeCount: 3000, expectedRange: "1,001-5,000 employees"},
		{name: "very large enterprise", employeeCount: 8000, expectedRange: "5,001-10,000 employees"},
		{name: "mega enterprise", employeeCount: 50000, expectedRange: "10,000+ employees"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rangeStr := analyzer.generateEmployeeRange(tt.employeeCount)
			assert.Equal(t, tt.expectedRange, rangeStr)
		})
	}
}

func TestCalculateConfidenceScore(t *testing.T) {
	analyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name        string
		result      *EmployeeCountResult
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "high confidence with all data",
			result: &EmployeeCountResult{
				EmployeeCount:  100,
				Evidence:       []string{"Direct mention: 100 employees"},
				SizeConfidence: 0.9,
				IsValidated:    true,
			},
			expectedMin: 0.8,
			expectedMax: 1.0,
		},
		{
			name: "medium confidence with partial data",
			result: &EmployeeCountResult{
				EmployeeCount:  50,
				Evidence:       []string{"Team indicator: startup team"},
				SizeConfidence: 0.7,
				IsValidated:    false,
			},
			expectedMin: 0.6,
			expectedMax: 1.0,
		},
		{
			name: "low confidence with minimal data",
			result: &EmployeeCountResult{
				EmployeeCount:  0,
				Evidence:       []string{},
				SizeConfidence: 0.0,
				IsValidated:    false,
			},
			expectedMin: 0.4,
			expectedMax: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := analyzer.calculateConfidenceScore(tt.result)
			assert.GreaterOrEqual(t, confidence, tt.expectedMin)
			assert.LessOrEqual(t, confidence, tt.expectedMax)
		})
	}
}

func TestValidateResults(t *testing.T) {
	analyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name           string
		result         *EmployeeCountResult
		expectedValid  bool
		expectedErrors int
	}{
		{
			name: "valid result",
			result: &EmployeeCountResult{
				EmployeeCount:   100,
				ConfidenceScore: 0.8,
				CompanySize:     "sme",
			},
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name: "negative employee count",
			result: &EmployeeCountResult{
				EmployeeCount:   -10,
				ConfidenceScore: 0.8,
				CompanySize:     "sme",
			},
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name: "unreasonably high employee count",
			result: &EmployeeCountResult{
				EmployeeCount:   2000000,
				ConfidenceScore: 0.8,
				CompanySize:     "enterprise",
			},
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name: "invalid confidence score",
			result: &EmployeeCountResult{
				EmployeeCount:   100,
				ConfidenceScore: 1.5,
				CompanySize:     "sme",
			},
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name: "missing company size",
			result: &EmployeeCountResult{
				EmployeeCount:   100,
				ConfidenceScore: 0.8,
				CompanySize:     "",
			},
			expectedValid:  false,
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := analyzer.validateResults(tt.result)
			assert.Equal(t, tt.expectedValid, status.IsValid)
			assert.Len(t, status.ValidationErrors, tt.expectedErrors)
			assert.True(t, status.LastValidated.After(time.Now().Add(-time.Second)))
		})
	}
}

func TestCalculateDataQuality(t *testing.T) {
	analyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name           string
		result         *EmployeeCountResult
		expectedFields []string
	}{
		{
			name: "complete data",
			result: &EmployeeCountResult{
				EmployeeCount:   100,
				CompanySize:     "sme",
				Evidence:        []string{"Direct mention"},
				IsValidated:     true,
				ConfidenceScore: 0.8,
			},
			expectedFields: []string{},
		},
		{
			name: "missing employee count",
			result: &EmployeeCountResult{
				EmployeeCount:   0,
				CompanySize:     "unknown",
				Evidence:        []string{},
				IsValidated:     false,
				ConfidenceScore: 0.3,
			},
			expectedFields: []string{"employee_count"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quality := analyzer.calculateDataQuality(tt.result)
			assert.GreaterOrEqual(t, quality.OverallScore, 0.0)
			assert.LessOrEqual(t, quality.OverallScore, 1.0)
			assert.Equal(t, tt.expectedFields, quality.MissingFields)
		})
	}
}

func TestGenerateReasoning(t *testing.T) {
	analyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name             string
		result           *EmployeeCountResult
		expectedContains []string
	}{
		{
			name: "with employee count",
			result: &EmployeeCountResult{
				EmployeeCount:   100,
				CompanySize:     "sme",
				ConfidenceScore: 0.8,
				SizeConfidence:  0.85,
				Evidence:        []string{"Direct mention: 100 employees"},
			},
			expectedContains: []string{"100 employees", "Sme", "80% confidence"},
		},
		{
			name: "no employee count",
			result: &EmployeeCountResult{
				EmployeeCount:   0,
				CompanySize:     "unknown",
				ConfidenceScore: 0.3,
				SizeConfidence:  0.0,
				Evidence:        []string{},
			},
			expectedContains: []string{"No employee count information"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reasoning := analyzer.generateReasoning(tt.result)
			assert.NotEmpty(t, reasoning)
			for _, expected := range tt.expectedContains {
				assert.Contains(t, reasoning, expected)
			}
		})
	}
}

func TestGetDefaultEmployeeCountConfig(t *testing.T) {
	config := getDefaultEmployeeCountConfig()
	assert.NotNil(t, config)
	assert.True(t, config.EnableEmployeeExtraction)
	assert.True(t, config.EnableSizeClassification)
	assert.True(t, config.EnableConfidenceScoring)
	assert.True(t, config.EnableValidation)
	assert.Equal(t, 50, config.StartupThreshold)
	assert.Equal(t, 51, config.SMEMinThreshold)
	assert.Equal(t, 250, config.SMEMaxThreshold)
	assert.Equal(t, 251, config.EnterpriseThreshold)
	assert.Equal(t, 0.3, config.MinConfidenceThreshold)
	assert.Equal(t, 5, config.MaxExtractionAttempts)
	assert.True(t, config.EnableDuplicateDetection)
	assert.True(t, config.EnableContextValidation)
}

func TestGetExtractionMethods(t *testing.T) {
	analyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name            string
		result          *EmployeeCountResult
		expectedMethods []string
	}{
		{
			name: "complete extraction",
			result: &EmployeeCountResult{
				ExtractionMethod: "direct_mention",
				CompanySize:      "sme",
				ConfidenceScore:  0.8,
				IsValidated:      true,
			},
			expectedMethods: []string{"direct_mention", "size_classification", "confidence_scoring", "validation"},
		},
		{
			name: "minimal extraction",
			result: &EmployeeCountResult{
				ExtractionMethod: "",
				CompanySize:      "",
				ConfidenceScore:  0.0,
				IsValidated:      false,
			},
			expectedMethods: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			methods := analyzer.getExtractionMethods(tt.result)
			assert.Equal(t, tt.expectedMethods, methods)
		})
	}
}
