package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

func TestValidator_SanitizeInput(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal text",
			input:    "Test Company",
			expected: "Test Company",
		},
		{
			name:     "Text with HTML tags",
			input:    "<script>alert('xss')</script>Test Company",
			expected: "Test Company",
		},
		{
			name:     "Text with SQL injection",
			input:    "Test'; DROP TABLE companies; --",
			expected: "Test",
		},
		{
			name:     "Text with special characters",
			input:    "Test & Company < > \" '",
			expected: "Test & Company < > \" '",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Whitespace only",
			input:    "   ",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.SanitizeInput(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidator_ValidateRiskAssessmentRequest(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		name       string
		request    *models.RiskAssessmentRequest
		expected   bool
		errorCount int
	}{
		{
			name: "Valid request",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Acme Corporation",
				BusinessAddress:   "123 Main St, Anytown, ST 12345",
				Industry:          "Technology",
				Country:           "US",
				Email:             "contact@acme.com",
				Phone:             "+15551234567",
				Website:           "https://acme.com",
				PredictionHorizon: 3,
			},
			expected:   true,
			errorCount: 0,
		},
		{
			name: "Invalid - missing business name",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "",
				BusinessAddress:   "123 Main St, Anytown, ST 12345",
				Industry:          "Technology",
				Country:           "US",
				PredictionHorizon: 3,
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "Invalid - missing business address",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Acme Corporation",
				BusinessAddress:   "",
				Industry:          "Technology",
				Country:           "US",
				PredictionHorizon: 3,
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "Invalid - missing industry",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Acme Corporation",
				BusinessAddress:   "123 Main St, Anytown, ST 12345",
				Industry:          "",
				Country:           "US",
				PredictionHorizon: 3,
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "Invalid - missing country",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Acme Corporation",
				BusinessAddress:   "123 Main St, Anytown, ST 12345",
				Industry:          "Technology",
				Country:           "",
				PredictionHorizon: 3,
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "Invalid - bad email format",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Acme Corporation",
				BusinessAddress:   "123 Main St, Anytown, ST 12345",
				Industry:          "Technology",
				Country:           "US",
				Email:             "invalid-email",
				PredictionHorizon: 3,
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "Invalid - bad phone format",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Acme Corporation",
				BusinessAddress:   "123 Main St, Anytown, ST 12345",
				Industry:          "Technology",
				Country:           "US",
				Phone:             "invalid-phone",
				PredictionHorizon: 3,
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "Invalid - bad website format",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Acme Corporation",
				BusinessAddress:   "123 Main St, Anytown, ST 12345",
				Industry:          "Technology",
				Country:           "US",
				Website:           "invalid-url",
				PredictionHorizon: 3,
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "Invalid - negative prediction horizon",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Acme Corporation",
				BusinessAddress:   "123 Main St, Anytown, ST 12345",
				Industry:          "Technology",
				Country:           "US",
				PredictionHorizon: -1,
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "Invalid - prediction horizon too high",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "Acme Corporation",
				BusinessAddress:   "123 Main St, Anytown, ST 12345",
				Industry:          "Technology",
				Country:           "US",
				PredictionHorizon: 25,
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "Multiple validation errors",
			request: &models.RiskAssessmentRequest{
				BusinessName:      "",
				BusinessAddress:   "",
				Industry:          "",
				Country:           "",
				Email:             "invalid-email",
				Phone:             "invalid-phone",
				Website:           "invalid-url",
				PredictionHorizon: -1,
			},
			expected:   false,
			errorCount: 8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidateRiskAssessmentRequest(tc.request)
			assert.Equal(t, tc.expected, result.Valid)
			assert.Len(t, result.Errors, tc.errorCount)
		})
	}
}

func TestValidator_ValidatePredictionRequest(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		name          string
		horizonMonths int
		scenarios     []string
		expected      bool
		errorCount    int
	}{
		{
			name:          "Valid prediction request",
			horizonMonths: 6,
			scenarios:     []string{"base_case", "stress_test"},
			expected:      true,
			errorCount:    0,
		},
		{
			name:          "Invalid - negative horizon",
			horizonMonths: -1,
			scenarios:     []string{"base_case"},
			expected:      false,
			errorCount:    1,
		},
		{
			name:          "Invalid - horizon too high",
			horizonMonths: 25,
			scenarios:     []string{"base_case"},
			expected:      false,
			errorCount:    1,
		},
		{
			name:          "Valid - no scenarios",
			horizonMonths: 6,
			scenarios:     []string{},
			expected:      true,
			errorCount:    0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidatePredictionRequest(tc.horizonMonths, tc.scenarios)
			assert.Equal(t, tc.expected, result.Valid)
			assert.Len(t, result.Errors, tc.errorCount)
		})
	}
}

// Test validation through the main validation function
func TestValidator_Integration(t *testing.T) {
	validator := NewValidator()

	// Test valid request with all fields
	validRequest := &models.RiskAssessmentRequest{
		BusinessName:      "Acme Corporation",
		BusinessAddress:   "123 Main St, Anytown, ST 12345",
		Industry:          "Technology",
		Country:           "US",
		Email:             "contact@acme.com",
		Phone:             "+15551234567",
		Website:           "https://acme.com",
		PredictionHorizon: 3,
	}

	result := validator.ValidateRiskAssessmentRequest(validRequest)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
	assert.Empty(t, result.Warnings)

	// Test invalid request with multiple issues
	invalidRequest := &models.RiskAssessmentRequest{
		BusinessName:      "",
		BusinessAddress:   "",
		Industry:          "",
		Country:           "",
		Email:             "invalid-email",
		Phone:             "invalid-phone",
		Website:           "invalid-url",
		PredictionHorizon: -1,
	}

	result = validator.ValidateRiskAssessmentRequest(invalidRequest)
	assert.False(t, result.Valid)
	assert.Greater(t, len(result.Errors), 5) // Should have multiple errors
}

// Benchmark tests
func BenchmarkValidateRiskAssessmentRequest(b *testing.B) {
	validator := NewValidator()
	request := &models.RiskAssessmentRequest{
		BusinessName:      "Acme Corporation",
		BusinessAddress:   "123 Main St, Anytown, ST 12345",
		Industry:          "Technology",
		Country:           "US",
		Email:             "contact@acme.com",
		Phone:             "+1-555-123-4567",
		Website:           "https://acme.com",
		PredictionHorizon: 3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateRiskAssessmentRequest(request)
	}
}

func BenchmarkSanitizeInput(b *testing.B) {
	validator := NewValidator()
	input := "Test Company with <script>alert('xss')</script> and SQL injection'; DROP TABLE companies; --"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.SanitizeInput(input)
	}
}
