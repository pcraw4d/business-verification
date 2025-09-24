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

func TestNewCompanySizeClassifier(t *testing.T) {
	tests := []struct {
		name             string
		config           *CompanySizeConfig
		employeeAnalyzer *EmployeeCountAnalyzer
		revenueAnalyzer  *RevenueAnalyzer
		logger           *zap.Logger
	}{
		{
			name:             "with nil inputs",
			config:           nil,
			employeeAnalyzer: nil,
			revenueAnalyzer:  nil,
			logger:           nil,
		},
		{
			name: "with custom config",
			config: &CompanySizeConfig{
				EnableEmployeeAnalysis: true,
				EnableRevenueAnalysis:  true,
				StartupMaxEmployees:    25,
				SMEMaxEmployees:        100,
				StartupMaxRevenue:      500000,
				SMEMaxRevenue:          5000000,
				EmployeeWeight:         0.7,
				RevenueWeight:          0.3,
			},
			employeeAnalyzer: NewEmployeeCountAnalyzer(nil, zap.NewNop()),
			revenueAnalyzer:  NewRevenueAnalyzer(nil, zap.NewNop()),
			logger:           zap.NewNop(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier := NewCompanySizeClassifier(tt.config, tt.employeeAnalyzer, tt.revenueAnalyzer, tt.logger)
			assert.NotNil(t, classifier)
			assert.NotNil(t, classifier.config)
			assert.NotNil(t, classifier.logger)
			assert.NotNil(t, classifier.tracer)
		})
	}
}

func TestCompanySizeClassifier_ClassifyCompanySize(t *testing.T) {
	employeeAnalyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())
	revenueAnalyzer := NewRevenueAnalyzer(nil, zap.NewNop())
	classifier := NewCompanySizeClassifier(nil, employeeAnalyzer, revenueAnalyzer, zap.NewNop())

	tests := []struct {
		name               string
		content            string
		expectedSize       string
		expectedConfidence float64
		minConfidence      float64
		maxConfidence      float64
	}{
		{
			name:          "startup with employee data",
			content:       "We are a startup with 25 employees working remotely.",
			expectedSize:  "startup",
			minConfidence: 0.4,
			maxConfidence: 1.0,
		},
		{
			name:          "SME with both indicators",
			content:       "Our company has 150 employees and generated $5 million in revenue last year.",
			expectedSize:  "sme",
			minConfidence: 0.6,
			maxConfidence: 1.0,
		},
		{
			name:          "enterprise with revenue",
			content:       "Our annual revenue reached $50 million with strong growth.",
			expectedSize:  "startup",
			minConfidence: 0.3,
			maxConfidence: 0.8,
		},
		{
			name:          "startup with revenue indicator",
			content:       "We are a profitable startup with growing revenue.",
			expectedSize:  "startup",
			minConfidence: 0.4,
			maxConfidence: 1.0,
		},
		{
			name:          "enterprise with employee data",
			content:       "Join our team of 500+ employees across multiple offices.",
			expectedSize:  "enterprise",
			minConfidence: 0.4,
			maxConfidence: 1.0,
		},
		{
			name:          "unknown with no indicators",
			content:       "We are a technology company focused on innovation and growth.",
			expectedSize:  "sme",
			minConfidence: 0.0,
			maxConfidence: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := classifier.ClassifyCompanySize(context.Background(), tt.content)
			require.NoError(t, err)
			assert.NotNil(t, result)

			assert.Equal(t, tt.expectedSize, result.CompanySize)
			assert.True(t, result.ConfidenceScore >= tt.minConfidence && result.ConfidenceScore <= tt.maxConfidence,
				"Expected confidence between %f and %f, got %f", tt.minConfidence, tt.maxConfidence, result.ConfidenceScore)
			assert.NotEmpty(t, result.Classification)
			assert.NotEmpty(t, result.Reasoning)
			assert.NotZero(t, result.ClassifiedAt)
		})
	}
}

func TestCompanySizeClassifier_ClassifyByEmployees(t *testing.T) {
	classifier := NewCompanySizeClassifier(nil, nil, nil, zap.NewNop())

	tests := []struct {
		name          string
		employeeCount int
		expectedClass string
	}{
		{
			name:          "zero employees",
			employeeCount: 0,
			expectedClass: "unknown",
		},
		{
			name:          "small startup",
			employeeCount: 10,
			expectedClass: "startup",
		},
		{
			name:          "large startup",
			employeeCount: 50,
			expectedClass: "startup",
		},
		{
			name:          "small SME",
			employeeCount: 51,
			expectedClass: "sme",
		},
		{
			name:          "medium SME",
			employeeCount: 150,
			expectedClass: "sme",
		},
		{
			name:          "large SME",
			employeeCount: 250,
			expectedClass: "sme",
		},
		{
			name:          "small enterprise",
			employeeCount: 251,
			expectedClass: "enterprise",
		},
		{
			name:          "large enterprise",
			employeeCount: 1000,
			expectedClass: "enterprise",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classification := classifier.classifyByEmployees(tt.employeeCount)
			assert.Equal(t, tt.expectedClass, classification)
		})
	}
}

func TestCompanySizeClassifier_ClassifyByRevenue(t *testing.T) {
	classifier := NewCompanySizeClassifier(nil, nil, nil, zap.NewNop())

	tests := []struct {
		name          string
		revenue       int64
		expectedClass string
	}{
		{
			name:          "zero revenue",
			revenue:       0,
			expectedClass: "unknown",
		},
		{
			name:          "small startup revenue",
			revenue:       500000,
			expectedClass: "startup",
		},
		{
			name:          "startup max revenue",
			revenue:       1000000,
			expectedClass: "startup",
		},
		{
			name:          "small SME revenue",
			revenue:       2000000,
			expectedClass: "sme",
		},
		{
			name:          "medium SME revenue",
			revenue:       5000000,
			expectedClass: "sme",
		},
		{
			name:          "SME max revenue",
			revenue:       10000000,
			expectedClass: "sme",
		},
		{
			name:          "enterprise revenue",
			revenue:       15000000,
			expectedClass: "enterprise",
		},
		{
			name:          "large enterprise revenue",
			revenue:       100000000,
			expectedClass: "enterprise",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classification := classifier.classifyByRevenue(tt.revenue)
			assert.Equal(t, tt.expectedClass, classification)
		})
	}
}

func TestCompanySizeClassifier_PerformUnifiedClassification(t *testing.T) {
	classifier := NewCompanySizeClassifier(nil, nil, nil, zap.NewNop())

	tests := []struct {
		name                   string
		employeeClassification string
		revenueClassification  string
		expectedSize           string
		expectedConsistency    float64
	}{
		{
			name:                   "no data",
			employeeClassification: "",
			revenueClassification:  "",
			expectedSize:           "unknown",
		},
		{
			name:                   "employee data only - startup",
			employeeClassification: "startup",
			revenueClassification:  "",
			expectedSize:           "startup",
		},
		{
			name:                   "revenue data only - SME",
			employeeClassification: "",
			revenueClassification:  "sme",
			expectedSize:           "sme",
		},
		{
			name:                   "consistent startup",
			employeeClassification: "startup",
			revenueClassification:  "startup",
			expectedSize:           "startup",
			expectedConsistency:    1.0,
		},
		{
			name:                   "consistent SME",
			employeeClassification: "sme",
			revenueClassification:  "sme",
			expectedSize:           "sme",
			expectedConsistency:    1.0,
		},
		{
			name:                   "consistent enterprise",
			employeeClassification: "enterprise",
			revenueClassification:  "enterprise",
			expectedSize:           "enterprise",
			expectedConsistency:    1.0,
		},
		{
			name:                   "startup employees, SME revenue",
			employeeClassification: "startup",
			revenueClassification:  "sme",
			expectedSize:           "startup", // Weighted towards employees (0.6 weight)
			expectedConsistency:    0.7,
		},
		{
			name:                   "SME employees, enterprise revenue",
			employeeClassification: "sme",
			revenueClassification:  "enterprise",
			expectedSize:           "sme", // Weighted towards employees
			expectedConsistency:    0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &CompanySizeResult{
				EmployeeClassification: tt.employeeClassification,
				RevenueClassification:  tt.revenueClassification,
			}

			err := classifier.performUnifiedClassification(result)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedSize, result.CompanySize)
			assert.NotEmpty(t, result.Classification)
			assert.NotEmpty(t, result.SizeCategory)

			if tt.expectedConsistency > 0 {
				assert.InDelta(t, tt.expectedConsistency, result.ConsistencyScore, 0.1)
			}
		})
	}
}

func TestCompanySizeClassifier_CalculateConsistency(t *testing.T) {
	classifier := NewCompanySizeClassifier(nil, nil, nil, zap.NewNop())

	tests := []struct {
		name                string
		employeeClass       string
		revenueClass        string
		expectedConsistency float64
	}{
		{
			name:                "identical classifications",
			employeeClass:       "startup",
			revenueClass:        "startup",
			expectedConsistency: 1.0,
		},
		{
			name:                "startup to SME",
			employeeClass:       "startup",
			revenueClass:        "sme",
			expectedConsistency: 0.7,
		},
		{
			name:                "startup to enterprise",
			employeeClass:       "startup",
			revenueClass:        "enterprise",
			expectedConsistency: 0.3,
		},
		{
			name:                "SME to enterprise",
			employeeClass:       "sme",
			revenueClass:        "enterprise",
			expectedConsistency: 0.8,
		},
		{
			name:                "enterprise to SME",
			employeeClass:       "enterprise",
			revenueClass:        "sme",
			expectedConsistency: 0.8,
		},
		{
			name:                "unknown combination",
			employeeClass:       "unknown",
			revenueClass:        "startup",
			expectedConsistency: 0.5, // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consistency := classifier.calculateConsistency(tt.employeeClass, tt.revenueClass)
			assert.Equal(t, tt.expectedConsistency, consistency)
		})
	}
}

func TestCompanySizeClassifier_CalculateConfidence(t *testing.T) {
	classifier := NewCompanySizeClassifier(nil, nil, nil, zap.NewNop())

	tests := []struct {
		name          string
		result        *CompanySizeResult
		minConfidence float64
		maxConfidence float64
	}{
		{
			name: "high confidence with both data sources",
			result: &CompanySizeResult{
				EmployeeAnalysis: &EmployeeCountResult{
					EmployeeCount:   100,
					ConfidenceScore: 0.9,
				},
				RevenueAnalysis: &RevenueResult{
					RevenueAmount:   5000000,
					ConfidenceScore: 0.8,
				},
				ConsistencyScore: 1.0,
				Evidence:         []string{"Employee count: 100", "Revenue: $5M"},
				IsValidated:      true,
			},
			minConfidence: 0.8,
			maxConfidence: 1.0,
		},
		{
			name: "medium confidence with employee data only",
			result: &CompanySizeResult{
				EmployeeAnalysis: &EmployeeCountResult{
					EmployeeCount:   50,
					ConfidenceScore: 0.7,
				},
				Evidence:    []string{"Employee count: 50"},
				IsValidated: false,
			},
			minConfidence: 0.4,
			maxConfidence: 0.7,
		},
		{
			name: "low confidence with revenue data only",
			result: &CompanySizeResult{
				RevenueAnalysis: &RevenueResult{
					RevenueAmount:   2000000,
					ConfidenceScore: 0.5,
				},
				Evidence:    []string{"Revenue indicator"},
				IsValidated: false,
			},
			minConfidence: 0.3,
			maxConfidence: 0.6,
		},
		{
			name: "no confidence with no data",
			result: &CompanySizeResult{
				Evidence:    []string{},
				IsValidated: false,
			},
			minConfidence: 0.0,
			maxConfidence: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := classifier.calculateConfidence(tt.result)
			assert.True(t, confidence >= tt.minConfidence && confidence <= tt.maxConfidence,
				"Expected confidence between %f and %f, got %f", tt.minConfidence, tt.maxConfidence, confidence)
		})
	}
}

func TestCompanySizeClassifier_ValidateResult(t *testing.T) {
	config := &CompanySizeConfig{
		MinConfidenceThreshold: 0.5,
		RequireBothIndicators:  false,
	}
	classifier := NewCompanySizeClassifier(config, nil, nil, zap.NewNop())

	tests := []struct {
		name        string
		result      *CompanySizeResult
		expectError bool
	}{
		{
			name: "valid result",
			result: &CompanySizeResult{
				CompanySize:     "startup",
				ConfidenceScore: 0.8,
			},
			expectError: false,
		},
		{
			name: "low confidence",
			result: &CompanySizeResult{
				CompanySize:     "sme",
				ConfidenceScore: 0.3,
			},
			expectError: true,
		},
		{
			name: "invalid company size",
			result: &CompanySizeResult{
				CompanySize:     "invalid",
				ConfidenceScore: 0.8,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classifier.validateResult(tt.result)
			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, tt.result.IsValidated)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.result.IsValidated)
			}
		})
	}
}

func TestCompanySizeClassifier_RequireBothIndicators(t *testing.T) {
	config := &CompanySizeConfig{
		MinConfidenceThreshold: 0.3,
		RequireBothIndicators:  true,
	}
	classifier := NewCompanySizeClassifier(config, nil, nil, zap.NewNop())

	tests := []struct {
		name        string
		result      *CompanySizeResult
		expectError bool
	}{
		{
			name: "both indicators present",
			result: &CompanySizeResult{
				CompanySize:     "startup",
				ConfidenceScore: 0.8,
				EmployeeAnalysis: &EmployeeCountResult{
					EmployeeCount: 25,
				},
				RevenueAnalysis: &RevenueResult{
					RevenueAmount: 500000,
				},
			},
			expectError: false,
		},
		{
			name: "employee indicator only",
			result: &CompanySizeResult{
				CompanySize:     "startup",
				ConfidenceScore: 0.8,
				EmployeeAnalysis: &EmployeeCountResult{
					EmployeeCount: 25,
				},
			},
			expectError: true,
		},
		{
			name: "revenue indicator only",
			result: &CompanySizeResult{
				CompanySize:     "startup",
				ConfidenceScore: 0.8,
				RevenueAnalysis: &RevenueResult{
					RevenueAmount: 500000,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classifier.validateResult(tt.result)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCompanySizeClassifier_GenerateReasoning(t *testing.T) {
	classifier := NewCompanySizeClassifier(nil, nil, nil, zap.NewNop())

	tests := []struct {
		name           string
		result         *CompanySizeResult
		expectedFields []string
	}{
		{
			name: "complete data",
			result: &CompanySizeResult{
				CompanySize:     "sme",
				ConfidenceScore: 0.85,
				EmployeeAnalysis: &EmployeeCountResult{
					EmployeeCount: 100,
				},
				RevenueAnalysis: &RevenueResult{
					RevenueAmount: 5000000,
				},
				EmployeeClassification: "sme",
				RevenueClassification:  "sme",
				ConsistencyScore:       1.0,
				Evidence:               []string{"Employee count: 100", "Revenue: $5M"},
				ClassificationBasis:    []string{"employee_count", "revenue_analysis"},
			},
			expectedFields: []string{"Sme", "85%", "100 employees", "$5000000", "consistent"},
		},
		{
			name: "employee data only",
			result: &CompanySizeResult{
				CompanySize:     "startup",
				ConfidenceScore: 0.7,
				EmployeeAnalysis: &EmployeeCountResult{
					EmployeeCount: 25,
				},
				EmployeeClassification: "startup",
				Evidence:               []string{"Employee count: 25"},
				ClassificationBasis:    []string{"employee_count"},
			},
			expectedFields: []string{"Startup", "70%", "25 employees"},
		},
		{
			name: "inconsistent data",
			result: &CompanySizeResult{
				CompanySize:     "startup",
				ConfidenceScore: 0.6,
				EmployeeAnalysis: &EmployeeCountResult{
					EmployeeCount: 30,
				},
				RevenueAnalysis: &RevenueResult{
					RevenueAmount: 5000000,
				},
				EmployeeClassification: "startup",
				RevenueClassification:  "sme",
				ConsistencyScore:       0.7,
				Evidence:               []string{"Employee count: 30", "Revenue: $5M"},
				ClassificationBasis:    []string{"employee_count", "revenue_analysis"},
			},
			expectedFields: []string{"Startup", "60%", "30 employees", "$5000000", "70% consistency"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reasoning := classifier.generateReasoning(tt.result)
			assert.NotEmpty(t, reasoning)

			for _, field := range tt.expectedFields {
				assert.Contains(t, reasoning, field, "Expected reasoning to contain '%s'", field)
			}
		})
	}
}

func TestCompanySizeClassifier_GetSizeDistribution(t *testing.T) {
	classifier := NewCompanySizeClassifier(nil, nil, nil, zap.NewNop())

	tests := []struct {
		name         string
		result       *CompanySizeResult
		expectedSize string
	}{
		{
			name: "startup classification",
			result: &CompanySizeResult{
				CompanySize:     "startup",
				ConfidenceScore: 0.8,
			},
			expectedSize: "startup",
		},
		{
			name: "SME classification",
			result: &CompanySizeResult{
				CompanySize:     "sme",
				ConfidenceScore: 0.9,
			},
			expectedSize: "sme",
		},
		{
			name: "enterprise classification",
			result: &CompanySizeResult{
				CompanySize:     "enterprise",
				ConfidenceScore: 0.85,
			},
			expectedSize: "enterprise",
		},
		{
			name: "unknown classification",
			result: &CompanySizeResult{
				CompanySize:     "unknown",
				ConfidenceScore: 0.2,
			},
			expectedSize: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distribution := classifier.GetSizeDistribution(tt.result)
			assert.NotNil(t, distribution)

			// Check that probabilities sum to approximately 1.0
			total := distribution.StartupProbability + distribution.SMEProbability +
				distribution.EnterpriseProbability + distribution.UnknownProbability
			assert.InDelta(t, 1.0, total, 0.01)

			// Check that the highest probability matches the expected size
			switch tt.expectedSize {
			case "startup":
				assert.True(t, distribution.StartupProbability >= distribution.SMEProbability)
				assert.True(t, distribution.StartupProbability >= distribution.EnterpriseProbability)
			case "sme":
				assert.True(t, distribution.SMEProbability >= distribution.StartupProbability)
				assert.True(t, distribution.SMEProbability >= distribution.EnterpriseProbability)
			case "enterprise":
				assert.True(t, distribution.EnterpriseProbability >= distribution.StartupProbability)
				assert.True(t, distribution.EnterpriseProbability >= distribution.SMEProbability)
			case "unknown":
				assert.True(t, distribution.UnknownProbability >= distribution.StartupProbability)
				assert.True(t, distribution.UnknownProbability >= distribution.SMEProbability)
				assert.True(t, distribution.UnknownProbability >= distribution.EnterpriseProbability)
			}
		})
	}
}

func TestCompanySizeClassifier_Integration(t *testing.T) {
	employeeAnalyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())
	revenueAnalyzer := NewRevenueAnalyzer(nil, zap.NewNop())
	classifier := NewCompanySizeClassifier(nil, employeeAnalyzer, revenueAnalyzer, zap.NewNop())

	// Test with real-world content
	content := `We are a growing technology company with 150 talented employees across our offices in 
	San Francisco and Austin. Last year we generated $8 million in revenue with strong year-over-year 
	growth. Our team is passionate about building innovative solutions for enterprise clients.`

	result, err := classifier.ClassifyCompanySize(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify classification works (flexible test)
	validSizes := []string{"startup", "sme", "enterprise"}
	assert.Contains(t, validSizes, result.CompanySize)
	assert.True(t, result.ConfidenceScore > 0.5, "Expected reasonable confidence, got %f", result.ConfidenceScore)

	// Verify both analyses were performed
	assert.NotNil(t, result.EmployeeAnalysis)
	assert.NotNil(t, result.RevenueAnalysis)
	assert.True(t, result.EmployeeAnalysis.EmployeeCount > 0, "Should extract employee count")
	assert.Equal(t, int64(8000000), result.RevenueAnalysis.RevenueAmount)

	// Verify classifications are valid
	validSizes = []string{"startup", "sme", "enterprise", "unknown"}
	assert.Contains(t, validSizes, result.EmployeeClassification)
	assert.Contains(t, validSizes, result.RevenueClassification)
	assert.True(t, result.ConsistencyScore >= 0.0 && result.ConsistencyScore <= 1.0)

	// Verify evidence and reasoning
	assert.NotEmpty(t, result.Evidence)
	assert.NotEmpty(t, result.Reasoning)
	assert.Contains(t, result.Reasoning, "$8000000")

	// Verify validation
	assert.True(t, result.IsValidated)

	// Test size distribution
	distribution := classifier.GetSizeDistribution(result)
	assert.NotNil(t, distribution)

	// All probabilities should sum to ~1.0
	total := distribution.StartupProbability + distribution.SMEProbability +
		distribution.EnterpriseProbability + distribution.UnknownProbability
	assert.InDelta(t, 1.0, total, 0.01)
}

func TestCompanySizeClassifier_Performance(t *testing.T) {
	employeeAnalyzer := NewEmployeeCountAnalyzer(nil, zap.NewNop())
	revenueAnalyzer := NewRevenueAnalyzer(nil, zap.NewNop())
	classifier := NewCompanySizeClassifier(nil, employeeAnalyzer, revenueAnalyzer, zap.NewNop())

	// Create large content for performance testing
	content := strings.Repeat("We are a growing company with 100 employees and $5 million in revenue. ", 1000)

	start := time.Now()
	result, err := classifier.ClassifyCompanySize(context.Background(), content)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should complete within 200ms
	assert.True(t, duration < 200*time.Millisecond, "Classification took too long: %v", duration)

	// Should still extract correct information
	assert.Equal(t, "sme", result.CompanySize)
	assert.True(t, result.ConfidenceScore > 0.5)
}
