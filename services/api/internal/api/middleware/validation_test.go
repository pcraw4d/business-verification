package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestStruct represents a struct for testing validation
type TestStruct struct {
	ID          string  `json:"id" validate:"required,uuid"`
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Email       string  `json:"email" validate:"required,email"`
	Phone       string  `json:"phone" validate:"phone"`
	URL         string  `json:"url" validate:"url"`
	Age         int     `json:"age" validate:"min=18,max=120"`
	Score       float64 `json:"score" validate:"min=0.0,max=100.0"`
	Description string  `json:"description" validate:"max=1000"`
	Website     string  `json:"website" validate:"url"`
	Code        string  `json:"code" validate:"alphanumeric"`
	Number      string  `json:"number" validate:"numeric"`
	Date        string  `json:"date" validate:"date"`
	Path        string  `json:"path" validate:"path_traversal"`
	Script      string  `json:"script" validate:"xss"`
	SQL         string  `json:"sql" validate:"sql_injection"`
}

func TestNewValidator(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil config
	validator := NewValidator(nil, logger)
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.config)
	assert.Equal(t, int64(10*1024*1024), validator.config.MaxRequestSize)
	assert.True(t, validator.config.EnableSanitization)

	// Test with custom config
	config := &ValidationConfig{
		MaxRequestSize:      5 * 1024 * 1024,
		EnableSanitization:  false,
		StrictMode:          true,
		LogValidationErrors: false,
	}

	validator = NewValidator(config, logger)
	assert.NotNil(t, validator)
	assert.Equal(t, config.MaxRequestSize, validator.config.MaxRequestSize)
	assert.False(t, validator.config.EnableSanitization)
	assert.True(t, validator.config.StrictMode)
}

func TestValidateStruct_Required(t *testing.T) {
	logger := zap.NewNop()
	validator := NewValidator(nil, logger)

	tests := []struct {
		name     string
		input    interface{}
		expected bool
		errors   int
	}{
		{
			name: "valid struct",
			input: TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "John Doe",
				Email: "john@example.com",
			},
			expected: true,
			errors:   0,
		},
		{
			name: "missing required field",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			expected: false,
			errors:   1,
		},
		{
			name: "empty required string",
			input: TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "",
				Email: "john@example.com",
			},
			expected: false,
			errors:   1,
		},
		{
			name: "whitespace only string",
			input: TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "   ",
				Email: "john@example.com",
			},
			expected: false,
			errors:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateStruct(tt.input)
			assert.Equal(t, tt.expected, result.IsValid)
			assert.Len(t, result.Errors, tt.errors)
		})
	}
}

func TestValidateStruct_Email(t *testing.T) {
	logger := zap.NewNop()
	validator := NewValidator(nil, logger)

	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"valid email", "test@example.com", true},
		{"valid email with subdomain", "test@sub.example.com", true},
		{"valid email with plus", "test+tag@example.com", true},
		{"valid email with underscore", "test_user@example.com", true},
		{"invalid email - no @", "testexample.com", false},
		{"invalid email - no domain", "test@", false},
		{"invalid email - no local part", "@example.com", false},
		{"invalid email - multiple @", "test@@example.com", false},
		{"invalid email - spaces", "test @example.com", false},
		{"invalid email - too long", "a" + string(make([]byte, 250)) + "@example.com", false},
		{"empty email", "", true}, // Empty is allowed unless required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "Test",
				Email: tt.email,
			}

			result := validator.ValidateStruct(input)
			if tt.expected {
				assert.True(t, result.IsValid, "Expected valid for email: %s", tt.email)
			} else {
				assert.False(t, result.IsValid, "Expected invalid for email: %s", tt.email)
			}
		})
	}
}

func TestValidateStruct_URL(t *testing.T) {
	logger := zap.NewNop()
	validator := NewValidator(nil, logger)

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"valid http url", "http://example.com", true},
		{"valid https url", "https://example.com", true},
		{"valid url with path", "https://example.com/path", true},
		{"valid url with query", "https://example.com?param=value", true},
		{"valid url with fragment", "https://example.com#section", true},
		{"invalid url - no protocol", "example.com", false},
		{"invalid url - wrong protocol", "ftp://example.com", false},
		{"invalid url - spaces", "https://example.com/path with spaces", false},
		{"empty url", "", true}, // Empty is allowed unless required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "Test",
				Email: "test@example.com",
				URL:   tt.url,
			}

			result := validator.ValidateStruct(input)
			if tt.expected {
				assert.True(t, result.IsValid, "Expected valid for URL: %s", tt.url)
			} else {
				assert.False(t, result.IsValid, "Expected invalid for URL: %s", tt.url)
			}
		})
	}
}

func TestValidateStruct_Phone(t *testing.T) {
	logger := zap.NewNop()
	validator := NewValidator(nil, logger)

	tests := []struct {
		name     string
		phone    string
		expected bool
	}{
		{"valid phone", "+1234567890", true},
		{"valid phone with country code", "+44123456789", true},
		{"valid phone with longer number", "+123456789012345", true},
		{"invalid phone - no plus", "1234567890", false},
		{"invalid phone - starts with 0", "+01234567890", false},
		{"invalid phone - too short", "+123", false},
		{"invalid phone - too long", "+1234567890123456", false},
		{"invalid phone - letters", "+1234567890a", false},
		{"invalid phone - spaces", "+1 234 567 890", false},
		{"empty phone", "", true}, // Empty is allowed unless required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "Test",
				Email: "test@example.com",
				Phone: tt.phone,
			}

			result := validator.ValidateStruct(input)
			if tt.expected {
				assert.True(t, result.IsValid, "Expected valid for phone: %s", tt.phone)
			} else {
				assert.False(t, result.IsValid, "Expected invalid for phone: %s", tt.phone)
			}
		})
	}
}

func TestValidateStruct_UUID(t *testing.T) {
	logger := zap.NewNop()
	validator := NewValidator(nil, logger)

	tests := []struct {
		name     string
		uuid     string
		expected bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid uuid uppercase", "550E8400-E29B-41D4-A716-446655440000", true},
		{"invalid uuid - wrong format", "550e8400-e29b-41d4-a716-44665544000", false},
		{"invalid uuid - missing hyphens", "550e8400e29b41d4a716446655440000", false},
		{"invalid uuid - wrong characters", "550e8400-e29b-41d4-a716-44665544000g", false},
		{"invalid uuid - too short", "550e8400-e29b-41d4-a716", false},
		{"invalid uuid - too long", "550e8400-e29b-41d4-a716-446655440000-extra", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := TestStruct{
				ID:    tt.uuid,
				Name:  "Test",
				Email: "test@example.com",
			}

			result := validator.ValidateStruct(input)
			if tt.expected {
				assert.True(t, result.IsValid, "Expected valid for UUID: %s", tt.uuid)
			} else {
				assert.False(t, result.IsValid, "Expected invalid for UUID: %s", tt.uuid)
			}
		})
	}
}

func TestValidateStruct_MinMax(t *testing.T) {
	logger := zap.NewNop()
	validator := NewValidator(nil, logger)

	tests := []struct {
		name     string
		input    TestStruct
		expected bool
	}{
		{
			name: "valid string length",
			input: TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "John",
				Email: "test@example.com",
			},
			expected: true,
		},
		{
			name: "string too short",
			input: TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "J",
				Email: "test@example.com",
			},
			expected: false,
		},
		{
			name: "string too long",
			input: TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  string(make([]byte, 101)),
				Email: "test@example.com",
			},
			expected: false,
		},
		{
			name: "valid age",
			input: TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "John",
				Email: "test@example.com",
				Age:   25,
			},
			expected: true,
		},
		{
			name: "age too young",
			input: TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "John",
				Email: "test@example.com",
				Age:   16,
			},
			expected: false,
		},
		{
			name: "age too old",
			input: TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "John",
				Email: "test@example.com",
				Age:   150,
			},
			expected: false,
		},
		{
			name: "valid score",
			input: TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "John",
				Email: "test@example.com",
				Score: 85.5,
			},
			expected: true,
		},
		{
			name: "score too low",
			input: TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "John",
				Email: "test@example.com",
				Score: -5.0,
			},
			expected: false,
		},
		{
			name: "score too high",
			input: TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "John",
				Email: "test@example.com",
				Score: 150.0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateStruct(tt.input)
			assert.Equal(t, tt.expected, result.IsValid)
		})
	}
}

func TestValidateStruct_Security(t *testing.T) {
	logger := zap.NewNop()
	validator := NewValidator(nil, logger)

	tests := []struct {
		name     string
		input    TestStruct
		expected bool
	}{
		{
			name: "valid input",
			input: TestStruct{
				ID:     "550e8400-e29b-41d4-a716-446655440000",
				Name:   "John Doe",
				Email:  "test@example.com",
				Path:   "/safe/path",
				Script: "normal text",
				SQL:    "normal query",
			},
			expected: true,
		},
		{
			name: "path traversal attempt",
			input: TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "John Doe",
				Email: "test@example.com",
				Path:  "../../../etc/passwd",
			},
			expected: false,
		},
		{
			name: "XSS attempt",
			input: TestStruct{
				ID:     "550e8400-e29b-41d4-a716-446655440000",
				Name:   "John Doe",
				Email:  "test@example.com",
				Script: "<script>alert('xss')</script>",
			},
			expected: false,
		},
		{
			name: "SQL injection attempt",
			input: TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "John Doe",
				Email: "test@example.com",
				SQL:   "'; DROP TABLE users; --",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateStruct(tt.input)
			assert.Equal(t, tt.expected, result.IsValid)
		})
	}
}

func TestValidateStruct_Alphanumeric(t *testing.T) {
	logger := zap.NewNop()
	validator := NewValidator(nil, logger)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"valid alphanumeric", "ABC123", true},
		{"valid alphanumeric lowercase", "abc123", true},
		{"valid alphanumeric mixed", "AbC123", true},
		{"invalid - spaces", "ABC 123", false},
		{"invalid - special chars", "ABC-123", false},
		{"invalid - underscore", "ABC_123", false},
		{"empty string", "", true}, // Empty is allowed unless required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "Test",
				Email: "test@example.com",
				Code:  tt.code,
			}

			result := validator.ValidateStruct(input)
			if tt.expected {
				assert.True(t, result.IsValid, "Expected valid for code: %s", tt.code)
			} else {
				assert.False(t, result.IsValid, "Expected invalid for code: %s", tt.code)
			}
		})
	}
}

func TestValidateStruct_Numeric(t *testing.T) {
	logger := zap.NewNop()
	validator := NewValidator(nil, logger)

	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{"valid numeric", "123456", true},
		{"valid numeric zero", "0", true},
		{"invalid - letters", "123abc", false},
		{"invalid - special chars", "123-456", false},
		{"invalid - spaces", "123 456", false},
		{"empty string", "", true}, // Empty is allowed unless required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := TestStruct{
				ID:     "550e8400-e29b-41d4-a716-446655440000",
				Name:   "Test",
				Email:  "test@example.com",
				Number: tt.number,
			}

			result := validator.ValidateStruct(input)
			if tt.expected {
				assert.True(t, result.IsValid, "Expected valid for number: %s", tt.number)
			} else {
				assert.False(t, result.IsValid, "Expected invalid for number: %s", tt.number)
			}
		})
	}
}

func TestValidateStruct_Date(t *testing.T) {
	logger := zap.NewNop()
	validator := NewValidator(nil, logger)

	tests := []struct {
		name     string
		date     string
		expected bool
	}{
		{"valid ISO date", "2023-12-25", true},
		{"valid ISO datetime", "2023-12-25T15:04:05Z", true},
		{"valid ISO datetime with milliseconds", "2023-12-25T15:04:05.000Z", true},
		{"valid datetime with space", "2023-12-25 15:04:05", true},
		{"valid US date", "12/25/2023", true},
		{"valid EU date", "25/12/2023", true},
		{"invalid date", "2023-13-45", false},
		{"invalid format", "25-12-2023", false},
		{"empty string", "", true}, // Empty is allowed unless required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := TestStruct{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Name:  "Test",
				Email: "test@example.com",
				Date:  tt.date,
			}

			result := validator.ValidateStruct(input)
			if tt.expected {
				assert.True(t, result.IsValid, "Expected valid for date: %s", tt.date)
			} else {
				assert.False(t, result.IsValid, "Expected invalid for date: %s", tt.date)
			}
		})
	}
}

func TestValidationMiddleware(t *testing.T) {
	logger := zap.NewNop()
	validator := NewValidator(nil, logger)

	tests := []struct {
		name           string
		method         string
		contentType    string
		body           string
		expectedStatus int
	}{
		{
			name:           "valid POST request",
			method:         "POST",
			contentType:    "application/json",
			body:           `{"name":"test"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid content type",
			method:         "POST",
			contentType:    "text/plain",
			body:           `{"name":"test"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "GET request (no content type check)",
			method:         "GET",
			contentType:    "",
			body:           "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()

			handler := validator.ValidationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestValidateRequest(t *testing.T) {
	logger := zap.NewNop()
	validator := NewValidator(nil, logger)

	tests := []struct {
		name           string
		body           string
		expectedValid  bool
		expectedErrors int
	}{
		{
			name:           "valid JSON",
			body:           `{"id":"550e8400-e29b-41d4-a716-446655440000","name":"John","email":"test@example.com"}`,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "invalid JSON",
			body:           `{"id":"550e8400-e29b-41d4-a716-446655440000","name":"John","email":"test@example.com"`,
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "valid JSON with validation errors",
			body:           `{"id":"invalid-uuid","name":"J","email":"invalid-email"}`,
			expectedValid:  false,
			expectedErrors: 3, // invalid uuid, name too short, invalid email
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			var target TestStruct
			result := validator.ValidateRequest(req, &target)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			assert.Len(t, result.Errors, tt.expectedErrors)
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	validationError := ValidationError{
		Field:    "email",
		Message:  "Invalid email format",
		Value:    "invalid-email",
		Rule:     "email",
		Severity: "error",
		Code:     "invalid_email",
	}

	expected := "validation failed for field 'email': Invalid email format"
	assert.Equal(t, expected, validationError.Error())
}

func TestValidationErrors_Error(t *testing.T) {
	errors := ValidationErrors{
		{
			Field:    "email",
			Message:  "Invalid email format",
			Severity: "error",
			Code:     "invalid_email",
		},
		{
			Field:    "name",
			Message:  "Field is required",
			Severity: "error",
			Code:     "required_field",
		},
	}

	expected := "validation failed for field 'email': Invalid email format; validation failed for field 'name': Field is required"
	assert.Equal(t, expected, errors.Error())
}

func TestValidationErrors_Empty(t *testing.T) {
	errors := ValidationErrors{}
	assert.Equal(t, "no validation errors", errors.Error())
}

func BenchmarkValidateStruct(b *testing.B) {
	logger := zap.NewNop()
	validator := NewValidator(nil, logger)

	input := TestStruct{
		ID:    "550e8400-e29b-41d4-a716-446655440000",
		Name:  "John Doe",
		Email: "john@example.com",
		Phone: "+1234567890",
		URL:   "https://example.com",
		Age:   25,
		Score: 85.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateStruct(input)
	}
}

func BenchmarkValidationMiddleware(b *testing.B) {
	logger := zap.NewNop()
	validator := NewValidator(nil, logger)

	handler := validator.ValidationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}
