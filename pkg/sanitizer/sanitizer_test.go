package sanitizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSanitizer(t *testing.T) {
	sanitizer := NewSanitizer()
	assert.NotNil(t, sanitizer)
	assert.NotNil(t, sanitizer.htmlTagPattern)
	assert.NotNil(t, sanitizer.sqlInjectionPattern)
	assert.NotNil(t, sanitizer.xssPattern)
	assert.NotNil(t, sanitizer.pathTraversalPattern)
}

func TestNewSanitizerWithConfig(t *testing.T) {
	config := &SanitizationConfig{
		RemoveHTMLTags:       false,
		RemoveScripts:        true,
		NormalizeUnicode:     true,
		TrimWhitespace:       false,
		RemoveNullBytes:      true,
		NormalizeLineEndings: false,
		MaxLength:            5000,
		AllowHTML:            true,
		StrictMode:           true,
	}

	sanitizer := NewSanitizerWithConfig(config)
	assert.NotNil(t, sanitizer)
}

func TestSanitizeString(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		config   *SanitizationConfig
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			config:   nil,
			expected: "",
		},
		{
			name:     "normal string",
			input:    "Hello, World!",
			config:   nil,
			expected: "Hello, World!",
		},
		{
			name:     "string with null bytes",
			input:    "Hello\x00World",
			config:   nil,
			expected: "HelloWorld",
		},
		{
			name:     "string with HTML tags",
			input:    "<p>Hello <b>World</b></p>",
			config:   nil,
			expected: "Hello World",
		},
		{
			name:     "string with scripts",
			input:    "Hello <script>alert('xss')</script> World",
			config:   nil,
			expected: "Hello  World",
		},
		{
			name:     "string with mixed line endings",
			input:    "Hello\r\nWorld\rTest",
			config:   nil,
			expected: "Hello\nWorld\nTest",
		},
		{
			name:     "string with extra whitespace",
			input:    "  Hello   World  ",
			config:   nil,
			expected: "Hello World",
		},
		{
			name:     "string exceeding max length",
			input:    string(make([]byte, 15000)),
			config:   nil,
			expected: string(make([]byte, 10000)),
		},
		{
			name:     "strict mode sanitization",
			input:    "Hello\x01\x02\x03World",
			config:   &SanitizationConfig{StrictMode: true},
			expected: "HelloWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input, tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeEmail(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name        string
		email       string
		expected    string
		expectError bool
	}{
		{
			name:        "valid email",
			email:       "test@example.com",
			expected:    "test@example.com",
			expectError: false,
		},
		{
			name:        "valid email with plus",
			email:       "test+tag@example.com",
			expected:    "test+tag@example.com",
			expectError: false,
		},
		{
			name:        "valid email with underscore",
			email:       "test_user@example.com",
			expected:    "test_user@example.com",
			expectError: false,
		},
		{
			name:        "email with HTML tags",
			email:       "<script>test@example.com</script>",
			expected:    "test@example.com",
			expectError: false,
		},
		{
			name:        "email with null bytes",
			email:       "test\x00@example.com",
			expected:    "test@example.com",
			expectError: false,
		},
		{
			name:        "invalid email - no @",
			email:       "testexample.com",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid email - no domain",
			email:       "test@",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid email - consecutive dots",
			email:       "test..user@example.com",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid email - starts with dot",
			email:       ".test@example.com",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid email - ends with dot",
			email:       "test@example.com.",
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty email",
			email:       "",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizer.SanitizeEmail(tt.email)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSanitizeURL(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name        string
		url         string
		expected    string
		expectError bool
	}{
		{
			name:        "valid http URL",
			url:         "http://example.com",
			expected:    "http://example.com",
			expectError: false,
		},
		{
			name:        "valid https URL",
			url:         "https://example.com",
			expected:    "https://example.com",
			expectError: false,
		},
		{
			name:        "valid URL with path",
			url:         "https://example.com/path",
			expected:    "https://example.com/path",
			expectError: false,
		},
		{
			name:        "valid URL with query",
			url:         "https://example.com?param=value",
			expected:    "https://example.com?param=value",
			expectError: false,
		},
		{
			name:        "URL with HTML tags",
			url:         "<script>https://example.com</script>",
			expected:    "https://example.com",
			expectError: false,
		},
		{
			name:        "URL with null bytes",
			url:         "https://example\x00.com",
			expected:    "https://example.com",
			expectError: false,
		},
		{
			name:        "invalid URL - no protocol",
			url:         "example.com",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid URL - wrong protocol",
			url:         "ftp://example.com",
			expected:    "",
			expectError: true,
		},
		{
			name:        "dangerous URL - javascript",
			url:         "javascript:alert('xss')",
			expected:    "",
			expectError: true,
		},
		{
			name:        "dangerous URL - vbscript",
			url:         "vbscript:msgbox('xss')",
			expected:    "",
			expectError: true,
		},
		{
			name:        "dangerous URL - data",
			url:         "data:text/html,<script>alert('xss')</script>",
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty URL",
			url:         "",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizer.SanitizeURL(tt.url)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSanitizePhone(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name        string
		phone       string
		expected    string
		expectError bool
	}{
		{
			name:        "valid phone",
			phone:       "+1234567890",
			expected:    "+1234567890",
			expectError: false,
		},
		{
			name:        "valid phone with country code",
			phone:       "+44123456789",
			expected:    "+44123456789",
			expectError: false,
		},
		{
			name:        "phone with HTML tags",
			phone:       "<script>+1234567890</script>",
			expected:    "+1234567890",
			expectError: false,
		},
		{
			name:        "phone with null bytes",
			phone:       "+123\x004567890",
			expected:    "+1234567890",
			expectError: false,
		},
		{
			name:        "invalid phone - no plus",
			phone:       "1234567890",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid phone - starts with 0",
			phone:       "+01234567890",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid phone - too short",
			phone:       "+123",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid phone - too long",
			phone:       "+1234567890123456",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid phone - letters",
			phone:       "+1234567890a",
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty phone",
			phone:       "",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizer.SanitizePhone(tt.phone)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSanitizeUUID(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name        string
		uuid        string
		expected    string
		expectError bool
	}{
		{
			name:        "valid UUID",
			uuid:        "550e8400-e29b-41d4-a716-446655440000",
			expected:    "550e8400-e29b-41d4-a716-446655440000",
			expectError: false,
		},
		{
			name:        "valid UUID uppercase",
			uuid:        "550E8400-E29B-41D4-A716-446655440000",
			expected:    "550e8400-e29b-41d4-a716-446655440000",
			expectError: false,
		},
		{
			name:        "UUID with HTML tags",
			uuid:        "<script>550e8400-e29b-41d4-a716-446655440000</script>",
			expected:    "550e8400-e29b-41d4-a716-446655440000",
			expectError: false,
		},
		{
			name:        "UUID with null bytes",
			uuid:        "550e8400\x00-e29b-41d4-a716-446655440000",
			expected:    "550e8400-e29b-41d4-a716-446655440000",
			expectError: false,
		},
		{
			name:        "invalid UUID - wrong format",
			uuid:        "550e8400-e29b-41d4-a716-44665544000",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid UUID - missing hyphens",
			uuid:        "550e8400e29b41d4a716446655440000",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid UUID - wrong characters",
			uuid:        "550e8400-e29b-41d4-a716-44665544000g",
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty UUID",
			uuid:        "",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizer.SanitizeUUID(tt.uuid)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSanitizeHTML(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "normal text",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "HTML with scripts",
			input:    "<p>Hello <script>alert('xss')</script> World</p>",
			expected: "<p>Hello  World</p>",
		},
		{
			name:     "HTML with dangerous attributes",
			input:    "<p onclick=\"alert('xss')\">Hello World</p>",
			expected: "<p>Hello World</p>",
		},
		{
			name:     "HTML with dangerous tags",
			input:    "<p>Hello <iframe src=\"evil.com\"></iframe> World</p>",
			expected: "<p>Hello  World</p>",
		},
		{
			name:     "HTML with null bytes",
			input:    "<p>Hello\x00World</p>",
			expected: "<p>HelloWorld</p>",
		},
		{
			name:     "safe HTML",
			input:    "<p>Hello <strong>World</strong></p>",
			expected: "<p>Hello <strong>World</strong></p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeHTML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid filename",
			input:    "document.pdf",
			expected: "document.pdf",
		},
		{
			name:     "filename with path traversal",
			input:    "../../../etc/passwd",
			expected: "etcpasswd",
		},
		{
			name:     "filename with dangerous characters",
			input:    "file<name>.txt",
			expected: "filename.txt",
		},
		{
			name:     "filename with null bytes",
			input:    "file\x00name.txt",
			expected: "filename.txt",
		},
		{
			name:     "filename with spaces and dots",
			input:    "  file.name.txt  ",
			expected: "file.name.txt",
		},
		{
			name:     "filename with control characters",
			input:    "file\x01\x02\x03name.txt",
			expected: "filename.txt",
		},
		{
			name:     "very long filename",
			input:    string(make([]byte, 300)),
			expected: string(make([]byte, 255)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeFilename(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeSQL(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "SQL injection attempt",
			input:    "'; DROP TABLE users; --",
			expected: "DROP TABLE users;",
		},
		{
			name:     "SQL with comments",
			input:    "SELECT * FROM users -- comment",
			expected: "SELECT * FROM users",
		},
		{
			name:     "SQL with multiple spaces",
			input:    "SELECT    *    FROM    users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "SQL with mixed case",
			input:    "SeLeCt * FrOm UsErS",
			expected: "* FrOm UsErS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeSQL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSecureToken(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name        string
		length      int
		expectError bool
	}{
		{
			name:        "valid length",
			length:      32,
			expectError: false,
		},
		{
			name:        "zero length",
			length:      0,
			expectError: true,
		},
		{
			name:        "negative length",
			length:      -1,
			expectError: true,
		},
		{
			name:        "very long token",
			length:      1000,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := sanitizer.GenerateSecureToken(tt.length)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, token, tt.length)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestValidateInput(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name        string
		input       string
		inputType   string
		expectError bool
	}{
		{
			name:        "valid email",
			input:       "test@example.com",
			inputType:   "email",
			expectError: false,
		},
		{
			name:        "invalid email",
			input:       "invalid-email",
			inputType:   "email",
			expectError: true,
		},
		{
			name:        "valid URL",
			input:       "https://example.com",
			inputType:   "url",
			expectError: false,
		},
		{
			name:        "invalid URL",
			input:       "not-a-url",
			inputType:   "url",
			expectError: true,
		},
		{
			name:        "valid phone",
			input:       "+1234567890",
			inputType:   "phone",
			expectError: false,
		},
		{
			name:        "invalid phone",
			input:       "1234567890",
			inputType:   "phone",
			expectError: true,
		},
		{
			name:        "valid UUID",
			input:       "550e8400-e29b-41d4-a716-446655440000",
			inputType:   "uuid",
			expectError: false,
		},
		{
			name:        "invalid UUID",
			input:       "invalid-uuid",
			inputType:   "uuid",
			expectError: true,
		},
		{
			name:        "unknown input type",
			input:       "test",
			inputType:   "unknown",
			expectError: true,
		},
		{
			name:        "empty input",
			input:       "",
			inputType:   "email",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizer.ValidateInput(tt.input, tt.inputType)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUtilityFunctions(t *testing.T) {
	sanitizer := NewSanitizer()

	t.Run("IsValidUTF8", func(t *testing.T) {
		assert.True(t, sanitizer.IsValidUTF8("Hello, World!"))
		assert.False(t, sanitizer.IsValidUTF8("\xff\xfe"))
	})

	t.Run("NormalizeUnicode", func(t *testing.T) {
		result := sanitizer.NormalizeUnicode("café")
		assert.Equal(t, "café", result)
	})

	t.Run("EscapeHTML", func(t *testing.T) {
		result := sanitizer.EscapeHTML("<p>Hello & World</p>")
		assert.Equal(t, "&lt;p&gt;Hello &amp; World&lt;/p&gt;", result)
	})

	t.Run("UnescapeHTML", func(t *testing.T) {
		result := sanitizer.UnescapeHTML("&lt;p&gt;Hello &amp; World&lt;/p&gt;")
		assert.Equal(t, "<p>Hello & World</p>", result)
	})

	t.Run("RemoveControlCharacters", func(t *testing.T) {
		result := sanitizer.RemoveControlCharacters("Hello\x01\x02\x03World")
		assert.Equal(t, "HelloWorld", result)
	})

	t.Run("SanitizeJSON", func(t *testing.T) {
		result := sanitizer.SanitizeJSON(`{"name": "John\x00Doe"}`)
		assert.Equal(t, `{"name": "JohnDoe"}`, result)
	})

	t.Run("SanitizeXML", func(t *testing.T) {
		result := sanitizer.SanitizeXML(`<name>John\x00Doe</name>`)
		assert.Equal(t, `<name>JohnDoe</name>`, result)
	})
}

func BenchmarkSanitizeString(b *testing.B) {
	sanitizer := NewSanitizer()
	input := "Hello <script>alert('xss')</script> World with\x00null bytes and\r\nmixed\rline endings"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sanitizer.SanitizeString(input, nil)
	}
}

func BenchmarkSanitizeEmail(b *testing.B) {
	sanitizer := NewSanitizer()
	email := "test+tag@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sanitizer.SanitizeEmail(email)
	}
}

func BenchmarkSanitizeURL(b *testing.B) {
	sanitizer := NewSanitizer()
	url := "https://example.com/path?param=value#fragment"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sanitizer.SanitizeURL(url)
	}
}

func BenchmarkGenerateSecureToken(b *testing.B) {
	sanitizer := NewSanitizer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sanitizer.GenerateSecureToken(32)
	}
}
