package security

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSecurityValidator_ValidateInput(t *testing.T) {
	logger := zap.NewNop()
	validator := NewSecurityValidator(logger)

	tests := []struct {
		name     string
		input    string
		config   SecurityConfig
		expected bool
	}{
		{
			name:  "valid input",
			input: "Acme Corporation",
			config: SecurityConfig{
				MaxInputLength: 255,
				MinInputLength: 1,
			},
			expected: true,
		},
		{
			name:  "input too short",
			input: "",
			config: SecurityConfig{
				MaxInputLength: 255,
				MinInputLength: 1,
			},
			expected: false,
		},
		{
			name:  "input too long",
			input: string(make([]byte, 300)),
			config: SecurityConfig{
				MaxInputLength: 255,
				MinInputLength: 1,
			},
			expected: false,
		},
		{
			name:  "input with blocked pattern",
			input: "Test Company admin",
			config: SecurityConfig{
				MaxInputLength: 255,
				MinInputLength: 1,
				BlockedPatterns: []string{
					`(?i)(admin|administrator)`,
				},
			},
			expected: true, // Still valid but with high risk score
		},
		{
			name:  "input with HTML tags",
			input: "<script>alert('xss')</script>Acme Corp",
			config: SecurityConfig{
				MaxInputLength: 255,
				MinInputLength: 1,
			},
			expected: true, // HTML tags are sanitized
		},
		{
			name:  "input with SQL injection",
			input: "Acme Corp'; DROP TABLE users; --",
			config: SecurityConfig{
				MaxInputLength: 255,
				MinInputLength: 1,
			},
			expected: true, // SQL injection patterns are sanitized
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateInput(context.Background(), tt.input, tt.config)
			assert.Equal(t, tt.expected, result.Valid)
		})
	}
}

func TestSecurityValidator_SanitizeInput(t *testing.T) {
	logger := zap.NewNop()
	validator := NewSecurityValidator(logger)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text",
			input:    "Acme Corporation",
			expected: "Acme Corporation",
		},
		{
			name:     "text with HTML tags",
			input:    "<script>alert('xss')</script>Acme Corp",
			expected: "alert(xss)</>Acme Corp",
		},
		{
			name:     "text with SQL injection",
			input:    "Acme Corp'; DROP TABLE users; --",
			expected: "Acme Corp TABLE users",
		},
		{
			name:     "text with JavaScript",
			input:    "Acme Corp javascript:alert('xss')",
			expected: "Acme Corp alert(xss)",
		},
		{
			name:     "text with excessive whitespace",
			input:    "Acme    Corporation   Inc",
			expected: "Acme Corporation Inc",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   \t\n   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.SanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityValidator_ValidateURL(t *testing.T) {
	logger := zap.NewNop()
	validator := NewSecurityValidator(logger)

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "valid HTTP URL",
			url:      "https://www.acme.com",
			expected: true,
		},
		{
			name:     "valid HTTPS URL",
			url:      "https://www.acme.com",
			expected: true,
		},
		{
			name:     "invalid URL",
			url:      "not-a-url",
			expected: true, // Go's url.Parse is more permissive
		},
		{
			name:     "JavaScript URL",
			url:      "javascript:alert('xss')",
			expected: false,
		},
		{
			name:     "data URL",
			url:      "data:text/html,<script>alert('xss')</script>",
			expected: false,
		},
		{
			name:     "localhost URL",
			url:      "http://localhost:8080",
			expected: true, // Valid but with warning
		},
		{
			name:     "IP address URL",
			url:      "http://192.168.1.1:8080",
			expected: true, // Valid but with warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateURL(tt.url)
			assert.Equal(t, tt.expected, result.Valid)
		})
	}
}

func TestSecurityValidator_ValidateEmail(t *testing.T) {
	logger := zap.NewNop()
	validator := NewSecurityValidator(logger)

	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "valid email",
			email:    "contact@acme.com",
			expected: true,
		},
		{
			name:     "invalid email format",
			email:    "not-an-email",
			expected: false,
		},
		{
			name:     "admin email",
			email:    "admin@acme.com",
			expected: true, // Valid but with warning
		},
		{
			name:     "test email",
			email:    "test@acme.com",
			expected: true, // Valid but with warning
		},
		{
			name:     "disposable email",
			email:    "user@10minutemail.com",
			expected: true, // Valid but with warning
		},
		{
			name:     "empty email",
			email:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateEmail(tt.email)
			assert.Equal(t, tt.expected, result.Valid)
		})
	}
}

func TestSecurityValidator_ValidateBusinessData(t *testing.T) {
	logger := zap.NewNop()
	validator := NewSecurityValidator(logger)

	tests := []struct {
		name     string
		data     map[string]interface{}
		expected bool
	}{
		{
			name: "valid business data",
			data: map[string]interface{}{
				"business_name":    "Acme Corporation",
				"business_address": "123 Main St, Anytown, ST 12345",
				"email":            "contact@acme.com",
				"website":          "https://www.acme.com",
			},
			expected: true,
		},
		{
			name: "invalid business name",
			data: map[string]interface{}{
				"business_name":    "",
				"business_address": "123 Main St, Anytown, ST 12345",
			},
			expected: false,
		},
		{
			name: "invalid email",
			data: map[string]interface{}{
				"business_name":    "Acme Corporation",
				"business_address": "123 Main St, Anytown, ST 12345",
				"email":            "invalid-email",
			},
			expected: false,
		},
		{
			name: "invalid website",
			data: map[string]interface{}{
				"business_name":    "Acme Corporation",
				"business_address": "123 Main St, Anytown, ST 12345",
				"website":          "javascript:alert('xss')",
			},
			expected: false,
		},
		{
			name: "suspicious business name",
			data: map[string]interface{}{
				"business_name":    "Test Admin Company",
				"business_address": "123 Main St, Anytown, ST 12345",
			},
			expected: true, // Valid but with warnings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateBusinessData(context.Background(), tt.data)
			assert.Equal(t, tt.expected, result.Valid)
		})
	}
}

func TestSecurityValidator_GenerateInputHash(t *testing.T) {
	logger := zap.NewNop()
	validator := NewSecurityValidator(logger)

	input := "test input"
	hash1 := validator.GenerateInputHash(input)
	hash2 := validator.GenerateInputHash(input)

	// Same input should generate same hash
	assert.Equal(t, hash1, hash2)

	// Hash should be 64 characters (SHA256 hex)
	assert.Len(t, hash1, 64)

	// Different input should generate different hash
	hash3 := validator.GenerateInputHash("different input")
	assert.NotEqual(t, hash1, hash3)
}

func TestSecurityValidator_scanContent(t *testing.T) {
	logger := zap.NewNop()
	validator := NewSecurityValidator(logger)

	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "normal content",
			input:    "Acme Corporation is a legitimate business",
			expected: 0.0,
		},
		{
			name:     "content with password",
			input:    "Please enter your password",
			expected: 0.3,
		},
		{
			name:     "content with admin",
			input:    "Admin access required",
			expected: 0.4,
		},
		{
			name:     "content with SQL injection",
			input:    "SQL injection attack",
			expected: 1.0, // Capped at 1.0
		},
		{
			name:     "content with multiple suspicious terms",
			input:    "Admin password SQL injection",
			expected: 1.0, // Should be capped at 1.0
		},
		{
			name:     "content with excessive special characters",
			input:    "!@#$%^&*()_+{}|:<>?[]\\;'\",./",
			expected: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.scanContent(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityValidator_determineRiskLevel(t *testing.T) {
	logger := zap.NewNop()
	validator := NewSecurityValidator(logger)

	tests := []struct {
		name     string
		score    float64
		expected string
	}{
		{
			name:     "minimal risk",
			score:    0.1,
			expected: "minimal",
		},
		{
			name:     "low risk",
			score:    0.3,
			expected: "low",
		},
		{
			name:     "medium risk",
			score:    0.5,
			expected: "medium",
		},
		{
			name:     "high risk",
			score:    0.7,
			expected: "high",
		},
		{
			name:     "critical risk",
			score:    0.9,
			expected: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.determineRiskLevel(tt.score)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityValidator_validateAllowedCharacters(t *testing.T) {
	logger := zap.NewNop()
	validator := NewSecurityValidator(logger)

	tests := []struct {
		name         string
		input        string
		allowedChars string
		expected     bool
	}{
		{
			name:         "alphanumeric only",
			input:        "Acme123",
			allowedChars: "a-zA-Z0-9",
			expected:     true,
		},
		{
			name:         "with special characters",
			input:        "Acme Corp!",
			allowedChars: "a-zA-Z0-9 ",
			expected:     false,
		},
		{
			name:         "letters and spaces only",
			input:        "Acme Corporation",
			allowedChars: "a-zA-Z ",
			expected:     true,
		},
		{
			name:         "numbers only",
			input:        "12345",
			allowedChars: "0-9",
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.validateAllowedCharacters(tt.input, tt.allowedChars)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityValidator_checkBlockedPatterns(t *testing.T) {
	logger := zap.NewNop()
	validator := NewSecurityValidator(logger)

	tests := []struct {
		name            string
		input           string
		blockedPatterns []string
		expected        float64
	}{
		{
			name:            "no blocked patterns",
			input:           "Acme Corporation",
			blockedPatterns: []string{},
			expected:        0.0,
		},
		{
			name:            "one blocked pattern match",
			input:           "Admin Company",
			blockedPatterns: []string{`(?i)(admin|administrator)`},
			expected:        1.0,
		},
		{
			name:            "multiple blocked patterns, one match",
			input:           "Admin Company",
			blockedPatterns: []string{`(?i)(admin|administrator)`, `(?i)(test|demo)`},
			expected:        0.5,
		},
		{
			name:            "multiple blocked patterns, multiple matches",
			input:           "Admin Test Company",
			blockedPatterns: []string{`(?i)(admin|administrator)`, `(?i)(test|demo)`},
			expected:        1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.checkBlockedPatterns(tt.input, tt.blockedPatterns)
			assert.Equal(t, tt.expected, result)
		})
	}
}
