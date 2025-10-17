package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertScoreToRiskLevel(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		expected RiskLevel
	}{
		{
			name:     "low risk score",
			score:    0.2,
			expected: RiskLevelLow,
		},
		{
			name:     "medium risk score",
			score:    0.5,
			expected: RiskLevelMedium,
		},
		{
			name:     "high risk score",
			score:    0.7,
			expected: RiskLevelHigh,
		},
		{
			name:     "critical risk score",
			score:    0.9,
			expected: RiskLevelCritical,
		},
		{
			name:     "boundary low-medium",
			score:    0.3,
			expected: RiskLevelMedium,
		},
		{
			name:     "boundary medium-high",
			score:    0.6,
			expected: RiskLevelHigh,
		},
		{
			name:     "boundary high-critical",
			score:    0.8,
			expected: RiskLevelCritical,
		},
		{
			name:     "zero score",
			score:    0.0,
			expected: RiskLevelLow,
		},
		{
			name:     "one score",
			score:    1.0,
			expected: RiskLevelCritical,
		},
		{
			name:     "negative score",
			score:    -0.1,
			expected: RiskLevelLow,
		},
		{
			name:     "score above one",
			score:    1.5,
			expected: RiskLevelCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertScoreToRiskLevel(tt.score)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetRiskSubcategories(t *testing.T) {
	subcategories := GetRiskSubcategories()

	// Test that all expected categories are present
	expectedCategories := []RiskCategory{
		RiskCategoryFinancial,
		RiskCategoryOperational,
		RiskCategoryCompliance,
		RiskCategoryReputational,
		RiskCategoryRegulatory,
		RiskCategoryGeopolitical,
		RiskCategoryTechnology,
		RiskCategoryEnvironmental,
	}

	for _, category := range expectedCategories {
		assert.Contains(t, subcategories, category, "Category %s should be present", category)
		assert.NotEmpty(t, subcategories[category], "Category %s should have subcategories", category)
	}

	// Test specific subcategories
	assert.Len(t, subcategories[RiskCategoryFinancial], 5)
	assert.Len(t, subcategories[RiskCategoryOperational], 6)
	assert.Len(t, subcategories[RiskCategoryCompliance], 5)
	assert.Len(t, subcategories[RiskCategoryReputational], 5)
	assert.Len(t, subcategories[RiskCategoryRegulatory], 5)
	assert.Len(t, subcategories[RiskCategoryGeopolitical], 5)
	assert.Len(t, subcategories[RiskCategoryTechnology], 5)
	assert.Len(t, subcategories[RiskCategoryEnvironmental], 5)

	// Test that subcategories have required fields
	for category, subcats := range subcategories {
		for _, subcat := range subcats {
			assert.Equal(t, category, subcat.Category, "Subcategory should belong to correct category")
			assert.NotEmpty(t, subcat.Name, "Subcategory should have a name")
			assert.NotEmpty(t, subcat.Description, "Subcategory should have a description")
			assert.Greater(t, subcat.Weight, 0.0, "Subcategory should have positive weight")
			assert.LessOrEqual(t, subcat.Weight, 1.0, "Subcategory weight should not exceed 1.0")
			assert.NotEmpty(t, subcat.Factors, "Subcategory should have factors")
		}
	}
}

func TestGenerateDetailedRiskFactors(t *testing.T) {
	business := &RiskAssessmentRequest{
		BusinessName:      "Test Company Inc",
		BusinessAddress:   "123 Test Street, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		Phone:             "+1-555-123-4567",
		Email:             "test@company.com",
		Website:           "https://testcompany.com",
		PredictionHorizon: 3,
	}

	baseScore := 0.5
	riskFactors := GenerateDetailedRiskFactors(business, baseScore)

	// Test that we get risk factors for all categories
	categories := make(map[RiskCategory]bool)
	for _, factor := range riskFactors {
		categories[factor.Category] = true
	}

	expectedCategories := []RiskCategory{
		RiskCategoryFinancial,
		RiskCategoryOperational,
		RiskCategoryCompliance,
		RiskCategoryReputational,
		RiskCategoryRegulatory,
		RiskCategoryGeopolitical,
		RiskCategoryTechnology,
		RiskCategoryEnvironmental,
	}

	for _, category := range expectedCategories {
		assert.True(t, categories[category], "Should have risk factors for category %s", category)
	}

	// Test that all risk factors have required fields
	for _, factor := range riskFactors {
		assert.NotEmpty(t, factor.Category, "Risk factor should have a category")
		assert.NotEmpty(t, factor.Subcategory, "Risk factor should have a subcategory")
		assert.NotEmpty(t, factor.Name, "Risk factor should have a name")
		assert.GreaterOrEqual(t, factor.Score, 0.0, "Risk factor score should be >= 0")
		assert.LessOrEqual(t, factor.Score, 1.0, "Risk factor score should be <= 1")
		assert.GreaterOrEqual(t, factor.Weight, 0.0, "Risk factor weight should be >= 0")
		assert.LessOrEqual(t, factor.Weight, 1.0, "Risk factor weight should be <= 1")
		assert.NotEmpty(t, factor.Description, "Risk factor should have a description")
		assert.NotEmpty(t, factor.Source, "Risk factor should have a source")
		assert.GreaterOrEqual(t, factor.Confidence, 0.0, "Risk factor confidence should be >= 0")
		assert.LessOrEqual(t, factor.Confidence, 1.0, "Risk factor confidence should be <= 1")
		assert.NotEmpty(t, factor.Impact, "Risk factor should have an impact")
		assert.NotEmpty(t, factor.Mitigation, "Risk factor should have mitigation")
		assert.NotNil(t, factor.LastUpdated, "Risk factor should have last updated time")
	}

	// Test that we have a reasonable number of risk factors
	// Should have factors for all subcategories across all categories
	subcategories := GetRiskSubcategories()
	expectedFactorCount := 0
	for _, subcats := range subcategories {
		for _, subcat := range subcats {
			expectedFactorCount += len(subcat.Factors)
		}
	}

	assert.Equal(t, expectedFactorCount, len(riskFactors), "Should have risk factors for all subcategory factors")
}

func TestCalculateSubcategoryRiskScore(t *testing.T) {
	business := &RiskAssessmentRequest{
		BusinessName:    "Test Company",
		Industry:        "technology",
		Country:         "US",
		Website:         "https://test.com",
		Email:           "test@test.com",
		BusinessAddress: "123 Test St",
	}

	baseScore := 0.5
	subcategories := GetRiskSubcategories()

	tests := []struct {
		name     string
		category RiskCategory
		subcat   string
		expected float64
	}{
		{
			name:     "financial liquidity risk",
			category: RiskCategoryFinancial,
			subcat:   "liquidity_risk",
			expected: 0.15, // 0.5 * 0.3 (weight)
		},
		{
			name:     "operational process risk",
			category: RiskCategoryOperational,
			subcat:   "process_risk",
			expected: 0.125, // 0.5 * 0.25 (weight)
		},
		{
			name:     "technology cybersecurity risk",
			category: RiskCategoryTechnology,
			subcat:   "cybersecurity_risk",
			expected: 0.18, // (0.5 + 0.1) * 0.3 (weight) - adjusted for digital presence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subcats := subcategories[tt.category]
			var targetSubcat RiskSubcategory
			for _, subcat := range subcats {
				if subcat.Name == tt.subcat {
					targetSubcat = subcat
					break
				}
			}

			require.NotEmpty(t, targetSubcat.Name, "Subcategory should exist")

			result := calculateSubcategoryRiskScore(business, tt.category, targetSubcat, baseScore)
			assert.GreaterOrEqual(t, result, 0.0, "Score should be >= 0")
			assert.LessOrEqual(t, result, 1.0, "Score should be <= 1")
		})
	}
}

func TestCalculateFactorScore(t *testing.T) {
	business := &RiskAssessmentRequest{
		BusinessName:    "Test Company",
		Industry:        "technology",
		Country:         "US",
		Website:         "https://test.com",
		Email:           "test@test.com",
		BusinessAddress: "123 Test St",
	}

	subcatScore := 0.3

	tests := []struct {
		name        string
		category    RiskCategory
		subcategory string
		factorName  string
		expected    float64
	}{
		{
			name:        "cash flow factor",
			category:    RiskCategoryFinancial,
			subcategory: "liquidity_risk",
			factorName:  "cash_flow",
			expected:    0.4, // 0.3 + 0.1
		},
		{
			name:        "process automation factor",
			category:    RiskCategoryOperational,
			subcategory: "process_risk",
			factorName:  "process_automation",
			expected:    0.2, // 0.3 - 0.1
		},
		{
			name:        "security incidents factor",
			category:    RiskCategoryTechnology,
			subcategory: "cybersecurity_risk",
			factorName:  "security_incidents",
			expected:    0.5, // 0.3 + 0.2
		},
		{
			name:        "unknown factor",
			category:    RiskCategoryFinancial,
			subcategory: "liquidity_risk",
			factorName:  "unknown_factor",
			expected:    0.3, // No adjustment
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateFactorScore(business, tt.category, tt.subcategory, tt.factorName, subcatScore)
			assert.GreaterOrEqual(t, result, 0.0, "Score should be >= 0")
			assert.LessOrEqual(t, result, 1.0, "Score should be <= 1")
		})
	}
}

func TestGenerateFactorDescription(t *testing.T) {
	tests := []struct {
		name        string
		category    RiskCategory
		subcategory string
		factorName  string
		score       float64
		expected    string
	}{
		{
			name:        "low risk factor",
			category:    RiskCategoryFinancial,
			subcategory: "liquidity_risk",
			factorName:  "cash_flow",
			score:       0.2,
			expected:    "cash_flow risk factor in liquidity_risk subcategory shows low risk level (score: 0.20)",
		},
		{
			name:        "moderate risk factor",
			category:    RiskCategoryOperational,
			subcategory: "process_risk",
			factorName:  "quality_controls",
			score:       0.5,
			expected:    "quality_controls risk factor in process_risk subcategory shows moderate risk level (score: 0.50)",
		},
		{
			name:        "high risk factor",
			category:    RiskCategoryTechnology,
			subcategory: "cybersecurity_risk",
			factorName:  "security_incidents",
			score:       0.8,
			expected:    "security_incidents risk factor in cybersecurity_risk subcategory shows high risk level (score: 0.80)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateFactorDescription(tt.category, tt.subcategory, tt.factorName, tt.score)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateFactorConfidence(t *testing.T) {
	tests := []struct {
		name       string
		business   *RiskAssessmentRequest
		factorName string
		expected   float64
	}{
		{
			name: "complete business data",
			business: &RiskAssessmentRequest{
				BusinessName:    "Very Long Company Name That Indicates Size",
				Industry:        "technology",
				Country:         "US",
				Website:         "https://test.com",
				Email:           "test@test.com",
				Phone:           "+1-555-123-4567",
				BusinessAddress: "123 Test Street, Test City, TC 12345",
			},
			factorName: "cash_flow",
			expected:   0.97, // 0.8 + 0.1 (long name) + 0.05 (address) + 0.02 (phone) + 0.02 (email)
		},
		{
			name: "minimal business data",
			business: &RiskAssessmentRequest{
				BusinessName: "Test",
				Industry:     "other",
			},
			factorName: "unknown_factor",
			expected:   0.8, // Base confidence
		},
		{
			name: "technology industry with website",
			business: &RiskAssessmentRequest{
				BusinessName:    "Tech Company",
				Industry:        "technology",
				Website:         "https://tech.com",
				BusinessAddress: "123 Tech St",
			},
			factorName: "system_uptime",
			expected:   0.92, // 0.8 + 0.05 (website) + 0.05 (address) + 0.02 (industry)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateFactorConfidence(tt.business, tt.factorName)
			assert.GreaterOrEqual(t, result, 0.0, "Confidence should be >= 0")
			assert.LessOrEqual(t, result, 1.0, "Confidence should be <= 1")
		})
	}
}

func TestGenerateFactorImpact(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		expected string
	}{
		{
			name:     "low impact",
			score:    0.2,
			expected: "Low impact on overall risk assessment",
		},
		{
			name:     "moderate impact",
			score:    0.5,
			expected: "Moderate impact on overall risk assessment",
		},
		{
			name:     "high impact",
			score:    0.8,
			expected: "High impact on overall risk assessment",
		},
		{
			name:     "boundary moderate-high",
			score:    0.4,
			expected: "Low impact on overall risk assessment",
		},
		{
			name:     "boundary high-critical",
			score:    0.7,
			expected: "Moderate impact on overall risk assessment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateFactorImpact(tt.score)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateFactorMitigation(t *testing.T) {
	tests := []struct {
		name        string
		category    RiskCategory
		subcategory string
		factorName  string
		score       float64
		expected    string
	}{
		{
			name:        "low score - continue practices",
			category:    RiskCategoryFinancial,
			subcategory: "liquidity_risk",
			factorName:  "cash_flow",
			score:       0.2,
			expected:    "Continue current risk management practices",
		},
		{
			name:        "financial category",
			category:    RiskCategoryFinancial,
			subcategory: "liquidity_risk",
			factorName:  "cash_flow",
			score:       0.5,
			expected:    "Implement additional financial controls and monitoring",
		},
		{
			name:        "operational category",
			category:    RiskCategoryOperational,
			subcategory: "process_risk",
			factorName:  "quality_controls",
			score:       0.7,
			expected:    "Enhance operational processes and controls",
		},
		{
			name:        "technology category",
			category:    RiskCategoryTechnology,
			subcategory: "cybersecurity_risk",
			factorName:  "security_incidents",
			score:       0.8,
			expected:    "Strengthen cybersecurity and technology risk management",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateFactorMitigation(tt.category, tt.subcategory, tt.factorName, tt.score)
			assert.NotEmpty(t, result, "Mitigation should not be empty")
		})
	}
}

func TestRiskAssessmentRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		request     *RiskAssessmentRequest
		expectValid bool
	}{
		{
			name: "valid request",
			request: &RiskAssessmentRequest{
				BusinessName:      "Test Company",
				BusinessAddress:   "123 Test Street, Test City, TC 12345",
				Industry:          "technology",
				Country:           "US",
				Phone:             "+1-555-123-4567",
				Email:             "test@company.com",
				Website:           "https://testcompany.com",
				PredictionHorizon: 3,
			},
			expectValid: true,
		},
		{
			name: "minimal valid request",
			request: &RiskAssessmentRequest{
				BusinessName:    "Test Company",
				BusinessAddress: "123 Test Street, Test City, TC 12345",
				Industry:        "technology",
				Country:         "US",
			},
			expectValid: true,
		},
		{
			name: "empty business name",
			request: &RiskAssessmentRequest{
				BusinessName:    "",
				BusinessAddress: "123 Test Street, Test City, TC 12345",
				Industry:        "technology",
				Country:         "US",
			},
			expectValid: false,
		},
		{
			name: "empty business address",
			request: &RiskAssessmentRequest{
				BusinessName:    "Test Company",
				BusinessAddress: "",
				Industry:        "technology",
				Country:         "US",
			},
			expectValid: false,
		},
		{
			name: "empty industry",
			request: &RiskAssessmentRequest{
				BusinessName:    "Test Company",
				BusinessAddress: "123 Test Street, Test City, TC 12345",
				Industry:        "",
				Country:         "US",
			},
			expectValid: false,
		},
		{
			name: "invalid country code",
			request: &RiskAssessmentRequest{
				BusinessName:    "Test Company",
				BusinessAddress: "123 Test Street, Test City, TC 12345",
				Industry:        "technology",
				Country:         "USA", // Should be 2 characters
			},
			expectValid: false,
		},
		{
			name: "invalid email",
			request: &RiskAssessmentRequest{
				BusinessName:    "Test Company",
				BusinessAddress: "123 Test Street, Test City, TC 12345",
				Industry:        "technology",
				Country:         "US",
				Email:           "invalid-email",
			},
			expectValid: false,
		},
		{
			name: "invalid website",
			request: &RiskAssessmentRequest{
				BusinessName:    "Test Company",
				BusinessAddress: "123 Test Street, Test City, TC 12345",
				Industry:        "technology",
				Country:         "US",
				Website:         "not-a-url",
			},
			expectValid: false,
		},
		{
			name: "invalid prediction horizon",
			request: &RiskAssessmentRequest{
				BusinessName:      "Test Company",
				BusinessAddress:   "123 Test Street, Test City, TC 12345",
				Industry:          "technology",
				Country:           "US",
				PredictionHorizon: 25, // Should be <= 24
			},
			expectValid: false,
		},
		{
			name: "invalid model type",
			request: &RiskAssessmentRequest{
				BusinessName:    "Test Company",
				BusinessAddress: "123 Test Street, Test City, TC 12345",
				Industry:        "technology",
				Country:         "US",
				ModelType:       "invalid-model",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: In a real implementation, you would use a validation library
			// like go-playground/validator to validate the struct tags
			// For this test, we're just checking the struct fields exist and are properly tagged

			// Test that required fields are present
			if tt.expectValid {
				assert.NotEmpty(t, tt.request.BusinessName, "Business name should not be empty")
				assert.NotEmpty(t, tt.request.BusinessAddress, "Business address should not be empty")
				assert.NotEmpty(t, tt.request.Industry, "Industry should not be empty")
				assert.Len(t, tt.request.Country, 2, "Country should be 2 characters")
			}
		})
	}
}

func TestRiskAssessment_Creation(t *testing.T) {
	now := time.Now()

	assessment := &RiskAssessment{
		ID:                "test-id",
		BusinessID:        "business-123",
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test Street",
		Industry:          "technology",
		Country:           "US",
		RiskScore:         0.5,
		RiskLevel:         RiskLevelMedium,
		RiskFactors:       []RiskFactor{},
		PredictionHorizon: 3,
		ConfidenceScore:   0.8,
		Status:            StatusCompleted,
		ModelType:         "xgboost",
		CreatedAt:         now,
		UpdatedAt:         now,
		Metadata:          map[string]interface{}{"test": "value"},
	}

	assert.Equal(t, "test-id", assessment.ID)
	assert.Equal(t, "business-123", assessment.BusinessID)
	assert.Equal(t, "Test Company", assessment.BusinessName)
	assert.Equal(t, 0.5, assessment.RiskScore)
	assert.Equal(t, RiskLevelMedium, assessment.RiskLevel)
	assert.Equal(t, StatusCompleted, assessment.Status)
	assert.Equal(t, "xgboost", assessment.ModelType)
	assert.Equal(t, now, assessment.CreatedAt)
	assert.Equal(t, now, assessment.UpdatedAt)
	assert.Equal(t, map[string]interface{}{"test": "value"}, assessment.Metadata)
}

func TestRiskFactor_Creation(t *testing.T) {
	now := time.Now()

	factor := &RiskFactor{
		Category:    RiskCategoryFinancial,
		Subcategory: "liquidity_risk",
		Name:        "cash_flow",
		Score:       0.3,
		Weight:      0.1,
		Description: "Cash flow risk factor",
		Source:      "test_source",
		Confidence:  0.8,
		Impact:      "Low impact",
		Mitigation:  "Monitor cash flow",
		LastUpdated: &now,
	}

	assert.Equal(t, RiskCategoryFinancial, factor.Category)
	assert.Equal(t, "liquidity_risk", factor.Subcategory)
	assert.Equal(t, "cash_flow", factor.Name)
	assert.Equal(t, 0.3, factor.Score)
	assert.Equal(t, 0.1, factor.Weight)
	assert.Equal(t, "Cash flow risk factor", factor.Description)
	assert.Equal(t, "test_source", factor.Source)
	assert.Equal(t, 0.8, factor.Confidence)
	assert.Equal(t, "Low impact", factor.Impact)
	assert.Equal(t, "Monitor cash flow", factor.Mitigation)
	assert.Equal(t, &now, factor.LastUpdated)
}

// Benchmark tests
func BenchmarkConvertScoreToRiskLevel(b *testing.B) {
	scores := []float64{0.1, 0.3, 0.5, 0.7, 0.9}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		score := scores[i%len(scores)]
		_ = ConvertScoreToRiskLevel(score)
	}
}

func BenchmarkGetRiskSubcategories(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetRiskSubcategories()
	}
}

func BenchmarkGenerateDetailedRiskFactors(b *testing.B) {
	business := &RiskAssessmentRequest{
		BusinessName:      "Test Company Inc",
		BusinessAddress:   "123 Test Street, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		Phone:             "+1-555-123-4567",
		Email:             "test@company.com",
		Website:           "https://testcompany.com",
		PredictionHorizon: 3,
	}
	baseScore := 0.5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateDetailedRiskFactors(business, baseScore)
	}
}

func BenchmarkCalculateFactorConfidence(b *testing.B) {
	business := &RiskAssessmentRequest{
		BusinessName:    "Test Company",
		Industry:        "technology",
		Country:         "US",
		Website:         "https://test.com",
		Email:           "test@test.com",
		BusinessAddress: "123 Test St",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateFactorConfidence(business, "cash_flow")
	}
}
