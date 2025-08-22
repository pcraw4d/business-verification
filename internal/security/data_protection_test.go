package security

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewDataProtectionService(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil config
	service, err := NewDataProtectionService(nil, logger)
	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.config)
	assert.True(t, service.config.EnableAnonymization)
	assert.Equal(t, "hash", service.config.AnonymizationMethod)

	// Test with custom config
	customConfig := &DataProtectionConfig{
		EnableAnonymization:     false,
		AnonymizationMethod:     "mask",
		SaltLength:              16,
		HashAlgorithm:           "fnv",
		EnableEncryption:        false,
		EnablePrivacyValidation: false,
		StrictMode:              true,
		DefaultRetentionPeriod:  7 * 24 * time.Hour,
		MaxRetentionPeriod:      90 * 24 * time.Hour,
	}

	service, err = NewDataProtectionService(customConfig, logger)
	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, customConfig, service.config)
	assert.False(t, service.config.EnableAnonymization)
	assert.Equal(t, "mask", service.config.AnonymizationMethod)
}

func TestDataProtectionService_ProtectBusinessData(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewDataProtectionService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name     string
		data     map[string]interface{}
		expected []string // expected protected fields
	}{
		{
			name: "business data with sensitive information",
			data: map[string]interface{}{
				"business_name": "Acme Corp",
				"email":         "contact@acme.com",
				"phone":         "+1-555-123-4567",
				"address":       "123 Main St, Anytown, ST 12345",
				"industry":      "Technology",
				"revenue":       1000000,
			},
			expected: []string{"email", "phone", "address"},
		},
		{
			name: "business data without sensitive information",
			data: map[string]interface{}{
				"business_name":  "Acme Corp",
				"industry":       "Technology",
				"revenue":        1000000,
				"employee_count": 50,
			},
			expected: []string{},
		},
		{
			name: "data with SSN",
			data: map[string]interface{}{
				"business_name": "Test Corp",
				"tax_id":        "123-45-6789",
				"industry":      "Finance",
			},
			expected: []string{"tax_id"},
		},
		{
			name: "data with credit card",
			data: map[string]interface{}{
				"business_name": "E-commerce Corp",
				"payment_card":  "4111-1111-1111-1111",
				"industry":      "Retail",
			},
			expected: []string{"payment_card"},
		},
		{
			name:     "empty data",
			data:     map[string]interface{}{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ProtectBusinessData(ctx, tt.data)
			require.NoError(t, err)
			assert.NotNil(t, result)

			// Check that original data is preserved
			assert.Equal(t, tt.data, result.OriginalData)

			// Check that protected fields are detected
			assert.ElementsMatch(t, tt.expected, result.ProtectedFields)

			// Check that anonymized data has different values for protected fields
			for _, field := range tt.expected {
				originalValue := tt.data[field]
				anonymizedValue := result.AnonymizedData[field]

				assert.NotEqual(t, originalValue, anonymizedValue)
				assert.NotEmpty(t, anonymizedValue)
			}

			// Check that non-protected fields remain unchanged
			for key, value := range tt.data {
				if !contains(tt.expected, key) {
					assert.Equal(t, value, result.AnonymizedData[key])
				}
			}

			// Check metadata
			assert.NotNil(t, result.Metadata)
			assert.NotZero(t, result.Metadata["anonymization_timestamp"])
			assert.Equal(t, len(tt.expected), result.Metadata["protected_field_count"])

			// Check processing time
			assert.Greater(t, result.ProcessingTime, time.Duration(0))

			// Check confidence score
			assert.GreaterOrEqual(t, result.ConfidenceScore, 0.0)
			assert.LessOrEqual(t, result.ConfidenceScore, 1.0)
		})
	}
}

func TestDataProtectionService_ValidatePrivacyCompliance(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewDataProtectionService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name             string
		data             map[string]interface{}
		expectCompliant  bool
		expectViolations int
		expectWarnings   int
	}{
		{
			name: "compliant data",
			data: map[string]interface{}{
				"business_name": "Acme Corp",
				"industry":      "Technology",
				"consent_given": true,
			},
			expectCompliant:  true,
			expectViolations: 0,
			expectWarnings:   1, // business_name warning
		},
		{
			name: "data with PII exposure",
			data: map[string]interface{}{
				"business_name": "Acme Corp",
				"email":         "contact@acme.com",
				"phone":         "+1-555-123-4567",
			},
			expectCompliant:  false,
			expectViolations: 3, // email, phone, and consent missing
			expectWarnings:   3, // business_name, email, phone warnings
		},
		{
			name: "data with PII but consent given",
			data: map[string]interface{}{
				"business_name": "Acme Corp",
				"email":         "contact@acme.com",
				"consent_given": true,
			},
			expectCompliant:  false, // Still has PII exposure
			expectViolations: 1,     // email
			expectWarnings:   2,     // business_name and email warnings
		},
		{
			name: "data with retention violation",
			data: map[string]interface{}{
				"business_name":    "Acme Corp",
				"retention_period": 2 * 365 * 24 * time.Hour, // 2 years
			},
			expectCompliant:  false,
			expectViolations: 1, // retention violation
			expectWarnings:   1, // business_name warning
		},
		{
			name: "data with potential PII patterns",
			data: map[string]interface{}{
				"business_name": "Acme Corp",
				"contact_info":  "John Doe - john@example.com",
			},
			expectCompliant:  false, // Has PII exposure
			expectViolations: 2,     // contact_info PII and consent missing
			expectWarnings:   2,     // business_name and contact_info warnings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ValidatePrivacyCompliance(ctx, tt.data)
			require.NoError(t, err)
			assert.NotNil(t, result)

			assert.Equal(t, tt.expectCompliant, result.IsCompliant)
			assert.Len(t, result.Violations, tt.expectViolations)
			assert.Len(t, result.Warnings, tt.expectWarnings)

			// Check compliance score
			assert.GreaterOrEqual(t, result.ComplianceScore, 0.0)
			assert.LessOrEqual(t, result.ComplianceScore, 100.0)

			// Check processing time
			assert.Greater(t, result.ValidationTime, time.Duration(0))

			// Check recommendations
			if len(result.Violations) > 0 || len(result.Warnings) > 0 {
				assert.NotEmpty(t, result.Recommendations)
			}
		})
	}
}

func TestDataProtectionService_EncryptDecryptSensitiveData(t *testing.T) {
	logger := zap.NewNop()
	service, err := NewDataProtectionService(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()

	originalData := map[string]interface{}{
		"company_name": "Acme Corp",
		"email":        "contact@acme.com",
		"phone":        "+1-555-123-4567",
		"industry":     "Technology",
		"revenue":      1000000,
	}

	// Encrypt sensitive data
	encryptedData, err := service.EncryptSensitiveData(ctx, originalData)
	require.NoError(t, err)
	assert.NotNil(t, encryptedData)

	// Check that sensitive fields are encrypted
	assert.NotEqual(t, originalData["email"], encryptedData["email"])
	assert.NotEqual(t, originalData["phone"], encryptedData["phone"])

	// Check that non-sensitive fields remain unchanged
	assert.Equal(t, originalData["company_name"], encryptedData["company_name"])
	assert.Equal(t, originalData["industry"], encryptedData["industry"])
	assert.Equal(t, originalData["revenue"], encryptedData["revenue"])

	// Decrypt sensitive data
	decryptedData, err := service.DecryptSensitiveData(ctx, encryptedData)
	require.NoError(t, err)
	assert.NotNil(t, decryptedData)

	// Check that all data is restored
	assert.Equal(t, originalData, decryptedData)
}

func TestDataAnonymizer_AnonymizeValue(t *testing.T) {
	logger := zap.NewNop()
	config := &DataProtectionConfig{
		EnableAnonymization: true,
		SaltLength:          16,
		HashAlgorithm:       "sha256",
	}

	anonymizer := &DataAnonymizer{
		config: config,
		logger: logger,
	}

	tests := []struct {
		name      string
		method    string
		value     interface{}
		fieldName string
	}{
		{
			name:      "hash method",
			method:    "hash",
			value:     "test@example.com",
			fieldName: "email",
		},
		{
			name:      "mask method",
			method:    "mask",
			value:     "John Doe",
			fieldName: "name",
		},
		{
			name:      "pseudonymize method",
			method:    "pseudonymize",
			value:     "+1-555-123-4567",
			fieldName: "phone",
		},
		{
			name:      "nil value",
			method:    "hash",
			value:     nil,
			fieldName: "test",
		},
		{
			name:      "empty string",
			method:    "hash",
			value:     "",
			fieldName: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anonymizer.config.AnonymizationMethod = tt.method

			result, err := anonymizer.AnonymizeValue(tt.value, tt.fieldName)
			require.NoError(t, err)

			if tt.value == nil {
				assert.Nil(t, result)
			} else if tt.value == "" {
				assert.Equal(t, "", result)
			} else {
				assert.NotEqual(t, tt.value, result)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestDataAnonymizer_HashValue(t *testing.T) {
	logger := zap.NewNop()
	config := &DataProtectionConfig{
		SaltLength:    16,
		HashAlgorithm: "sha256",
	}

	anonymizer := &DataAnonymizer{
		config: config,
		logger: logger,
	}

	value := "test@example.com"
	fieldName := "email"

	// Test SHA256
	config.HashAlgorithm = "sha256"
	hash1, err := anonymizer.hashValue(value, fieldName)
	require.NoError(t, err)
	assert.NotEmpty(t, hash1)

	// Test FNV
	config.HashAlgorithm = "fnv"
	hash2, err := anonymizer.hashValue(value, fieldName)
	require.NoError(t, err)
	assert.NotEmpty(t, hash2)

	// Hashes should be different due to different algorithms
	assert.NotEqual(t, hash1, hash2)

	// Same input should produce same hash with same algorithm
	config.HashAlgorithm = "sha256"
	hash3, err := anonymizer.hashValue(value, fieldName)
	require.NoError(t, err)
	assert.NotEqual(t, hash1, hash3) // Different due to random salt
}

func TestDataAnonymizer_MaskValue(t *testing.T) {
	logger := zap.NewNop()
	config := &DataProtectionConfig{}

	anonymizer := &DataAnonymizer{
		config: config,
		logger: logger,
	}

	tests := []struct {
		name      string
		value     string
		fieldName string
		expected  string
	}{
		{
			name:      "long string",
			value:     "John Doe",
			fieldName: "name",
			expected:  "J******e",
		},
		{
			name:      "short string",
			value:     "Hi",
			fieldName: "greeting",
			expected:  "**",
		},
		{
			name:      "very short string",
			value:     "A",
			fieldName: "letter",
			expected:  "*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := anonymizer.maskValue(tt.value, tt.fieldName)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDataAnonymizer_PseudonymizeValue(t *testing.T) {
	logger := zap.NewNop()
	config := &DataProtectionConfig{}

	anonymizer := &DataAnonymizer{
		config: config,
		logger: logger,
	}

	value := "test@example.com"
	fieldName := "email"

	result, err := anonymizer.pseudonymizeValue(value, fieldName)
	require.NoError(t, err)

	// Should be a pseudonym format
	assert.Contains(t, result, "pseudo_")
	assert.Contains(t, result, fieldName)

	// Same input should produce same pseudonym
	result2, err := anonymizer.pseudonymizeValue(value, fieldName)
	require.NoError(t, err)
	assert.Equal(t, result, result2)

	// Different input should produce different pseudonym
	result3, err := anonymizer.pseudonymizeValue("different@example.com", fieldName)
	require.NoError(t, err)
	assert.NotEqual(t, result, result3)
}

func TestDataEncryptor_EncryptDecryptValue(t *testing.T) {
	logger := zap.NewNop()
	config := &DataProtectionConfig{}

	encryptor := &DataEncryptor{
		config: config,
		logger: logger,
		key:    make([]byte, 32),
	}

	tests := []struct {
		name  string
		value interface{}
	}{
		{
			name:  "string value",
			value: "test@example.com",
		},
		{
			name:  "number value",
			value: float64(12345), // Use float64 to match JSON unmarshaling
		},
		{
			name:  "map value",
			value: map[string]interface{}{"key": "value"},
		},
		{
			name:  "nil value",
			value: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := encryptor.EncryptValue(tt.value)
			require.NoError(t, err)

			if tt.value == nil {
				assert.Equal(t, "", encrypted)
			} else {
				assert.NotEmpty(t, encrypted)
				assert.NotEqual(t, tt.value, encrypted)
			}

			// Decrypt
			decrypted, err := encryptor.DecryptValue(encrypted)
			require.NoError(t, err)

			if tt.value == nil {
				assert.Nil(t, decrypted)
			} else {
				assert.Equal(t, tt.value, decrypted)
			}
		})
	}
}

func TestPrivacyValidator_ContainsPII(t *testing.T) {
	logger := zap.NewNop()
	config := &DataProtectionConfig{
		PIIPatterns: map[string][]string{
			"email": {
				`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
			},
			"phone": {
				`(\+?1[-.\s]?)?\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})`,
			},
		},
	}

	validator := &PrivacyValidator{
		config: config,
		logger: logger,
	}

	tests := []struct {
		name      string
		value     interface{}
		expectPII bool
	}{
		{
			name:      "email address",
			value:     "test@example.com",
			expectPII: true,
		},
		{
			name:      "phone number",
			value:     "+1-555-123-4567",
			expectPII: true,
		},
		{
			name:      "regular text",
			value:     "Acme Corporation",
			expectPII: false,
		},
		{
			name:      "nil value",
			value:     nil,
			expectPII: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.containsPII(tt.value)
			assert.Equal(t, tt.expectPII, result)
		})
	}
}

func TestPrivacyValidator_MatchesPIIPattern(t *testing.T) {
	logger := zap.NewNop()
	config := &DataProtectionConfig{}

	validator := &PrivacyValidator{
		config: config,
		logger: logger,
	}

	tests := []struct {
		name        string
		value       interface{}
		expectMatch bool
	}{
		{
			name:        "email with @",
			value:       "test@example.com",
			expectMatch: true,
		},
		{
			name:        "phone with parentheses",
			value:       "(555) 123-4567",
			expectMatch: true,
		},
		{
			name:        "name with space",
			value:       "John Doe",
			expectMatch: true,
		},
		{
			name:        "simple text",
			value:       "AcmeCorp",
			expectMatch: false,
		},
		{
			name:        "nil value",
			value:       nil,
			expectMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.matchesPIIPattern(tt.value)
			assert.Equal(t, tt.expectMatch, result)
		})
	}
}

func TestPrivacyValidator_CalculateComplianceScore(t *testing.T) {
	logger := zap.NewNop()
	config := &DataProtectionConfig{}

	validator := &PrivacyValidator{
		config: config,
		logger: logger,
	}

	// Test with no violations or warnings
	violations := []PrivacyViolation{}
	warnings := []PrivacyWarning{}

	score := validator.calculateComplianceScore(violations, warnings)
	assert.Equal(t, 100.0, score)

	// Test with violations
	violations = []PrivacyViolation{
		{Severity: "high"},
		{Severity: "medium"},
		{Severity: "low"},
	}
	warnings = []PrivacyWarning{{}, {}}

	score = validator.calculateComplianceScore(violations, warnings)
	expectedScore := 100.0 - 15.0 - 10.0 - 5.0 - 2.0 - 2.0 // 66.0
	assert.Equal(t, expectedScore, score)

	// Test with critical violation
	violations = []PrivacyViolation{
		{Severity: "critical"},
	}
	warnings = []PrivacyWarning{}

	score = validator.calculateComplianceScore(violations, warnings)
	assert.Equal(t, 75.0, score)

	// Test minimum score
	violations = []PrivacyViolation{
		{Severity: "critical"},
		{Severity: "critical"},
		{Severity: "critical"},
		{Severity: "critical"},
		{Severity: "critical"},
	}
	warnings = []PrivacyWarning{}

	score = validator.calculateComplianceScore(violations, warnings)
	assert.Equal(t, 0.0, score) // Should not go below 0
}

func TestPrivacyValidator_GenerateRecommendations(t *testing.T) {
	logger := zap.NewNop()
	config := &DataProtectionConfig{}

	validator := &PrivacyValidator{
		config: config,
		logger: logger,
	}

	// Test with violations
	violations := []PrivacyViolation{
		{Recommendation: "Fix PII exposure"},
		{Recommendation: "Reduce retention period"},
	}
	warnings := []PrivacyWarning{
		{Recommendation: "Review potential PII"},
	}

	recommendations := validator.generateRecommendations(violations, warnings)

	// Should include all violation and warning recommendations
	assert.Contains(t, recommendations, "Fix PII exposure")
	assert.Contains(t, recommendations, "Reduce retention period")
	assert.Contains(t, recommendations, "Review potential PII")

	// Should include general recommendations when violations exist
	assert.Contains(t, recommendations, "Implement comprehensive data protection measures")
	assert.Contains(t, recommendations, "Conduct regular privacy impact assessments")

	// Test with no violations
	violations = []PrivacyViolation{}
	warnings = []PrivacyWarning{{Recommendation: "Review potential PII"}}

	recommendations = validator.generateRecommendations(violations, warnings)

	// Should not include general recommendations when no violations
	assert.NotContains(t, recommendations, "Implement comprehensive data protection measures")
	assert.NotContains(t, recommendations, "Conduct regular privacy impact assessments")
}
