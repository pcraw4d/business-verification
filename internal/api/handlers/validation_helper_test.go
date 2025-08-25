package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestValidationHelper_ValidateBusinessClassificationRequest(t *testing.T) {
	logger := zap.NewNop()
	errorHandler := NewErrorHandler(logger)
	validator := NewValidationHelper(errorHandler)

	tests := []struct {
		name       string
		request    *BusinessClassificationRequest
		expected   bool
		errorCount int
	}{
		{
			name: "valid request",
			request: &BusinessClassificationRequest{
				BusinessName: "Acme Corporation",
				WebsiteURL:   "https://www.acme.com",
				Email:        "contact@acme.com",
			},
			expected:   true,
			errorCount: 0,
		},
		{
			name: "missing business name",
			request: &BusinessClassificationRequest{
				WebsiteURL: "https://www.acme.com",
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "empty business name",
			request: &BusinessClassificationRequest{
				BusinessName: "",
				WebsiteURL:   "https://www.acme.com",
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "business name too long",
			request: &BusinessClassificationRequest{
				BusinessName: string(make([]byte, 256)), // 256 characters
				WebsiteURL:   "https://www.acme.com",
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "invalid business name format",
			request: &BusinessClassificationRequest{
				BusinessName: "Acme@#$%",
				WebsiteURL:   "https://www.acme.com",
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "invalid URL",
			request: &BusinessClassificationRequest{
				BusinessName: "Acme Corporation",
				WebsiteURL:   "not-a-url",
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "invalid email",
			request: &BusinessClassificationRequest{
				BusinessName: "Acme Corporation",
				Email:        "invalid-email",
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "invalid phone",
			request: &BusinessClassificationRequest{
				BusinessName: "Acme Corporation",
				Phone:        "123-456-7890", // Not E.164 format
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "description too long",
			request: &BusinessClassificationRequest{
				BusinessName: "Acme Corporation",
				Description:  string(make([]byte, 1001)), // 1001 characters
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "multiple validation errors",
			request: &BusinessClassificationRequest{
				BusinessName: "",              // Missing required field
				WebsiteURL:   "not-a-url",     // Invalid URL
				Email:        "invalid-email", // Invalid email
			},
			expected:   false,
			errorCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateBusinessClassificationRequest(tt.request)

			assert.Equal(t, tt.expected, result.IsValid)
			assert.Len(t, result.Errors, tt.errorCount)

			if !result.IsValid {
				// Verify that all errors are ValidationError types
				for _, err := range result.Errors {
					_, ok := err.(*ValidationError)
					assert.True(t, ok, "Error should be a ValidationError")
				}
			}
		})
	}
}

func TestValidationHelper_ValidateBatchClassificationRequest(t *testing.T) {
	logger := zap.NewNop()
	errorHandler := NewErrorHandler(logger)
	validator := NewValidationHelper(errorHandler)

	tests := []struct {
		name       string
		request    *BatchClassificationRequest
		expected   bool
		errorCount int
	}{
		{
			name: "valid batch request",
			request: &BatchClassificationRequest{
				Businesses: []BusinessClassificationRequest{
					{BusinessName: "Acme Corporation"},
					{BusinessName: "Tech Solutions Inc"},
				},
			},
			expected:   true,
			errorCount: 0,
		},
		{
			name: "empty businesses array",
			request: &BatchClassificationRequest{
				Businesses: []BusinessClassificationRequest{},
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "too many businesses",
			request: &BatchClassificationRequest{
				Businesses: make([]BusinessClassificationRequest, 101), // 101 businesses
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "invalid business in batch",
			request: &BatchClassificationRequest{
				Businesses: []BusinessClassificationRequest{
					{BusinessName: "Acme Corporation"}, // Valid
					{BusinessName: ""},                 // Invalid - missing name
				},
			},
			expected:   false,
			errorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateBatchClassificationRequest(tt.request)

			assert.Equal(t, tt.expected, result.IsValid)
			assert.Len(t, result.Errors, tt.errorCount)
		})
	}
}

func TestValidationHelper_ValidateWebsiteVerificationRequest(t *testing.T) {
	logger := zap.NewNop()
	errorHandler := NewErrorHandler(logger)
	validator := NewValidationHelper(errorHandler)

	tests := []struct {
		name       string
		request    *WebsiteVerificationRequest
		expected   bool
		errorCount int
	}{
		{
			name: "valid verification request",
			request: &WebsiteVerificationRequest{
				BusinessName: "Acme Corporation",
				WebsiteURL:   "https://www.acme.com",
				Email:        "contact@acme.com",
			},
			expected:   true,
			errorCount: 0,
		},
		{
			name: "missing business name",
			request: &WebsiteVerificationRequest{
				WebsiteURL: "https://www.acme.com",
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "missing website URL",
			request: &WebsiteVerificationRequest{
				BusinessName: "Acme Corporation",
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "invalid website URL",
			request: &WebsiteVerificationRequest{
				BusinessName: "Acme Corporation",
				WebsiteURL:   "not-a-url",
			},
			expected:   false,
			errorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateWebsiteVerificationRequest(tt.request)

			assert.Equal(t, tt.expected, result.IsValid)
			assert.Len(t, result.Errors, tt.errorCount)
		})
	}
}

func TestValidationHelper_ValidateRiskAssessmentRequest(t *testing.T) {
	logger := zap.NewNop()
	errorHandler := NewErrorHandler(logger)
	validator := NewValidationHelper(errorHandler)

	tests := []struct {
		name       string
		request    *RiskAssessmentRequest
		expected   bool
		errorCount int
	}{
		{
			name: "valid risk assessment request",
			request: &RiskAssessmentRequest{
				BusinessID:     "123e4567-e89b-12d3-a456-426614174000",
				AssessmentType: "comprehensive",
			},
			expected:   true,
			errorCount: 0,
		},
		{
			name: "missing business ID",
			request: &RiskAssessmentRequest{
				AssessmentType: "comprehensive",
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "invalid business ID format",
			request: &RiskAssessmentRequest{
				BusinessID:     "invalid-id",
				AssessmentType: "comprehensive",
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "invalid assessment type",
			request: &RiskAssessmentRequest{
				BusinessID:     "123e4567-e89b-12d3-a456-426614174000",
				AssessmentType: "invalid-type",
			},
			expected:   false,
			errorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateRiskAssessmentRequest(tt.request)

			assert.Equal(t, tt.expected, result.IsValid)
			assert.Len(t, result.Errors, tt.errorCount)
		})
	}
}

func TestValidationHelper_HelperMethods(t *testing.T) {
	logger := zap.NewNop()
	errorHandler := NewErrorHandler(logger)
	validator := NewValidationHelper(errorHandler)

	t.Run("isValidBusinessName", func(t *testing.T) {
		validNames := []string{
			"Acme Corporation",
			"Tech Solutions Inc.",
			"Smith & Sons",
			"ABC Company (LLC)",
		}

		invalidNames := []string{
			"",
			"Acme@#$%",
			"Company<script>",
			"   ", // Only whitespace
		}

		for _, name := range validNames {
			assert.True(t, validator.isValidBusinessName(name), "Business name should be valid: %s", name)
		}

		for _, name := range invalidNames {
			assert.False(t, validator.isValidBusinessName(name), "Business name should be invalid: %s", name)
		}
	})

	t.Run("isValidURL", func(t *testing.T) {
		validURLs := []string{
			"https://www.example.com",
			"http://example.com",
			"https://api.example.com/path",
		}

		invalidURLs := []string{
			"not-a-url",
			"ftp://example.com",
			"example.com",
			"",
		}

		for _, url := range validURLs {
			assert.True(t, validator.isValidURL(url), "URL should be valid: %s", url)
		}

		for _, url := range invalidURLs {
			assert.False(t, validator.isValidURL(url), "URL should be invalid: %s", url)
		}
	})

	t.Run("isValidEmail", func(t *testing.T) {
		validEmails := []string{
			"user@example.com",
			"user.name@example.com",
			"user+tag@example.com",
		}

		invalidEmails := []string{
			"invalid-email",
			"user@",
			"@example.com",
			"",
		}

		for _, email := range validEmails {
			assert.True(t, validator.isValidEmail(email), "Email should be valid: %s", email)
		}

		for _, email := range invalidEmails {
			assert.False(t, validator.isValidEmail(email), "Email should be invalid: %s", email)
		}
	})

	t.Run("isValidPhone", func(t *testing.T) {
		validPhones := []string{
			"+1-555-123-4567",
			"+44-20-7946-0958",
			"+81-3-1234-5678",
		}

		invalidPhones := []string{
			"123-456-7890",
			"+1-555-123-456", // Too short
			"",
		}

		for _, phone := range validPhones {
			assert.True(t, validator.isValidPhone(phone), "Phone should be valid: %s", phone)
		}

		for _, phone := range invalidPhones {
			assert.False(t, validator.isValidPhone(phone), "Phone should be invalid: %s", phone)
		}
	})

	t.Run("isValidBusinessID", func(t *testing.T) {
		validIDs := []string{
			"123e4567-e89b-12d3-a456-426614174000",
			"550e8400-e29b-41d4-a716-446655440000",
		}

		invalidIDs := []string{
			"invalid-id",
			"123e4567-e89b-12d3-a456-42661417400", // Too short
			"",
		}

		for _, id := range validIDs {
			assert.True(t, validator.isValidBusinessID(id), "Business ID should be valid: %s", id)
		}

		for _, id := range invalidIDs {
			assert.False(t, validator.isValidBusinessID(id), "Business ID should be invalid: %s", id)
		}
	})
}

func TestValidationHelper_GetValidationErrorMessage(t *testing.T) {
	logger := zap.NewNop()
	errorHandler := NewErrorHandler(logger)
	validator := NewValidationHelper(errorHandler)

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name: "missing required field",
			err: &ValidationError{
				Code:    ErrorCodeMissingRequiredField,
				Field:   "business_name",
				Message: "Business name is required",
			},
			expected: "The field 'business_name' is required. Please provide a value for this field.",
		},
		{
			name: "field too long",
			err: &ValidationError{
				Code:    ErrorCodeFieldTooLong,
				Field:   "description",
				Message: "Description is too long",
			},
			expected: "The field 'description' is too long. Please shorten the value to meet the length requirement.",
		},
		{
			name: "invalid URL",
			err: &ValidationError{
				Code:    ErrorCodeInvalidURL,
				Message: "Invalid URL format",
			},
			expected: "Please provide a valid URL starting with http:// or https://",
		},
		{
			name: "invalid email",
			err: &ValidationError{
				Code:    ErrorCodeInvalidEmail,
				Message: "Invalid email format",
			},
			expected: "Please provide a valid email address in the format user@example.com",
		},
		{
			name: "invalid phone",
			err: &ValidationError{
				Code:    ErrorCodeInvalidPhone,
				Message: "Invalid phone format",
			},
			expected: "Please provide a valid phone number in E.164 format (e.g., +1-555-123-4567)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := validator.GetValidationErrorMessage(tt.err)
			assert.Equal(t, tt.expected, message)
		})
	}
}

func TestValidationHelper_GetValidationHelpURL(t *testing.T) {
	logger := zap.NewNop()
	errorHandler := NewErrorHandler(logger)
	validator := NewValidationHelper(errorHandler)

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name: "missing required field",
			err: &ValidationError{
				Code: ErrorCodeMissingRequiredField,
			},
			expected: "https://docs.kyb-platform.com/api/validation/required-fields",
		},
		{
			name: "field too long",
			err: &ValidationError{
				Code: ErrorCodeFieldTooLong,
			},
			expected: "https://docs.kyb-platform.com/api/validation/field-lengths",
		},
		{
			name: "invalid URL",
			err: &ValidationError{
				Code: ErrorCodeInvalidURL,
			},
			expected: "https://docs.kyb-platform.com/api/validation/urls",
		},
		{
			name: "invalid email",
			err: &ValidationError{
				Code: ErrorCodeInvalidEmail,
			},
			expected: "https://docs.kyb-platform.com/api/validation/emails",
		},
		{
			name: "invalid phone",
			err: &ValidationError{
				Code: ErrorCodeInvalidPhone,
			},
			expected: "https://docs.kyb-platform.com/api/validation/phone-numbers",
		},
		{
			name: "unknown error",
			err: &ValidationError{
				Code: ErrorCodeInvalidFieldFormat,
			},
			expected: "https://docs.kyb-platform.com/api/validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpURL := validator.GetValidationHelpURL(tt.err)
			assert.Equal(t, tt.expected, helpURL)
		})
	}
}

func TestValidationResult(t *testing.T) {
	t.Run("new validation result", func(t *testing.T) {
		result := NewValidationResult()
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
	})

	t.Run("add error", func(t *testing.T) {
		result := NewValidationResult()
		err := &ValidationError{
			Code:    ErrorCodeMissingRequiredField,
			Message: "Test error",
		}

		result.AddError(err)

		assert.False(t, result.IsValid)
		assert.Len(t, result.Errors, 1)
		assert.Equal(t, err, result.Errors[0])
	})

	t.Run("add multiple errors", func(t *testing.T) {
		result := NewValidationResult()
		err1 := &ValidationError{
			Code:    ErrorCodeMissingRequiredField,
			Message: "Error 1",
		}
		err2 := &ValidationError{
			Code:    ErrorCodeInvalidURL,
			Message: "Error 2",
		}

		result.AddError(err1)
		result.AddError(err2)

		assert.False(t, result.IsValid)
		assert.Len(t, result.Errors, 2)
		assert.Equal(t, err1, result.Errors[0])
		assert.Equal(t, err2, result.Errors[1])
	})
}
