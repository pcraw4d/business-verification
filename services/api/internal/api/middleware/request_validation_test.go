package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewRequestValidationMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		config   *RequestValidationConfig
		logger   *zap.Logger
		expected *RequestValidationMiddleware
	}{
		{
			name:   "nil config uses default",
			config: nil,
			logger: zap.NewNop(),
			expected: &RequestValidationMiddleware{
				config:      GetDefaultRequestValidationConfig(),
				logger:      zap.NewNop(),
				schemaCache: make(map[string]*ValidationSchema),
			},
		},
		{
			name:   "custom config",
			config: &RequestValidationConfig{Enabled: false},
			logger: zap.NewNop(),
			expected: &RequestValidationMiddleware{
				config:      &RequestValidationConfig{Enabled: false},
				logger:      zap.NewNop(),
				schemaCache: make(map[string]*ValidationSchema),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRequestValidationMiddleware(tt.config, tt.logger)

			assert.NotNil(t, middleware)
			assert.Equal(t, tt.expected.logger, middleware.logger)
			assert.NotNil(t, middleware.schemaCache)

			if tt.config == nil {
				// Check default config
				assert.Equal(t, tt.expected.config.Enabled, middleware.config.Enabled)
				assert.Equal(t, tt.expected.config.MaxBodySize, middleware.config.MaxBodySize)
				assert.Equal(t, tt.expected.config.MaxQueryParams, middleware.config.MaxQueryParams)
			} else {
				assert.Equal(t, tt.config.Enabled, middleware.config.Enabled)
			}
		})
	}
}

func TestRequestValidationMiddleware_Middleware(t *testing.T) {
	tests := []struct {
		name           string
		config         *RequestValidationConfig
		request        *http.Request
		handler        http.HandlerFunc
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "disabled validation",
			config:  &RequestValidationConfig{Enabled: false},
			request: httptest.NewRequest("GET", "/test", nil),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:    "valid GET request",
			config:  GetDefaultRequestValidationConfig(),
			request: httptest.NewRequest("GET", "/test", nil),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "valid POST request with JSON",
			config: GetDefaultRequestValidationConfig(),
			request: func() *http.Request {
				body := `{"name": "test"}`
				req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				return req
			}(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "invalid content type",
			config: &RequestValidationConfig{
				Enabled:             true,
				RequireContentType:  true,
				AllowedContentTypes: []string{"application/json"},
			},
			request: httptest.NewRequest("POST", "/test", strings.NewReader("test")),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "body too large",
			config: &RequestValidationConfig{
				Enabled:     true,
				MaxBodySize: 10,
			},
			request: func() *http.Request {
				body := strings.Repeat("a", 100)
				req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				return req
			}(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "too many query parameters",
			config: &RequestValidationConfig{
				Enabled:        true,
				MaxQueryParams: 2,
			},
			request: httptest.NewRequest("GET", "/test?p1=1&p2=2&p3=3", nil),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, obs := observer.New(zapcore.InfoLevel)
			logger := zap.New(core)

			middleware := NewRequestValidationMiddleware(tt.config, logger)
			handler := middleware.Middleware(tt.handler)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, tt.request)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError {
				assert.Greater(t, obs.Len(), 0)
			}
		})
	}
}

func TestRequestValidationMiddleware_ValidateRequest(t *testing.T) {
	tests := []struct {
		name           string
		config         *RequestValidationConfig
		request        *http.Request
		expectedValid  bool
		expectedErrors int
	}{
		{
			name:   "valid request",
			config: GetDefaultRequestValidationConfig(),
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				return req
			}(),
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name: "invalid content type",
			config: &RequestValidationConfig{
				Enabled:             true,
				RequireContentType:  true,
				AllowedContentTypes: []string{"application/json"},
			},
			request:        httptest.NewRequest("POST", "/test", strings.NewReader("test")),
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name: "body too large",
			config: &RequestValidationConfig{
				Enabled:     true,
				MaxBodySize: 10,
			},
			request: func() *http.Request {
				body := strings.Repeat("a", 100)
				req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
				req.ContentLength = int64(len(body))
				return req
			}(),
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name: "too many query parameters",
			config: &RequestValidationConfig{
				Enabled:        true,
				MaxQueryParams: 2,
			},
			request:        httptest.NewRequest("GET", "/test?p1=1&p2=2&p3=3", nil),
			expectedValid:  false,
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			middleware := NewRequestValidationMiddleware(tt.config, logger)

			result, err := middleware.ValidateRequest(tt.request)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedValid, result.Valid)
			assert.Len(t, result.Errors, tt.expectedErrors)
		})
	}
}

func TestRequestValidationMiddleware_validateContentType(t *testing.T) {
	tests := []struct {
		name           string
		config         *RequestValidationConfig
		contentType    string
		requireContent bool
		expectedError  bool
	}{
		{
			name:           "no content type required",
			config:         &RequestValidationConfig{RequireContentType: false},
			contentType:    "",
			requireContent: false,
			expectedError:  false,
		},
		{
			name:           "content type required but missing",
			config:         &RequestValidationConfig{RequireContentType: true},
			contentType:    "",
			requireContent: true,
			expectedError:  true,
		},
		{
			name:           "valid JSON content type",
			config:         &RequestValidationConfig{AllowedContentTypes: []string{"application/json"}},
			contentType:    "application/json",
			requireContent: false,
			expectedError:  false,
		},
		{
			name:           "invalid content type",
			config:         &RequestValidationConfig{AllowedContentTypes: []string{"application/json"}},
			contentType:    "text/plain",
			requireContent: false,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			middleware := NewRequestValidationMiddleware(tt.config, logger)

			req := httptest.NewRequest("POST", "/test", nil)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			err := middleware.validateContentType(req)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequestValidationMiddleware_validateQueryParameters(t *testing.T) {
	tests := []struct {
		name           string
		config         *RequestValidationConfig
		queryString    string
		expectedErrors int
	}{
		{
			name:           "valid query parameters",
			config:         GetDefaultRequestValidationConfig(),
			queryString:    "p1=1&p2=2",
			expectedErrors: 0,
		},
		{
			name: "too many query parameters",
			config: &RequestValidationConfig{
				Enabled:        true,
				MaxQueryParams: 2,
			},
			queryString:    "p1=1&p2=2&p3=3",
			expectedErrors: 1,
		},
		{
			name: "query parameter value too long",
			config: &RequestValidationConfig{
				Enabled:           true,
				MaxQueryValueSize: 10,
				PreventInjection:  false,
			},
			queryString:    "p1=" + strings.Repeat("a", 20),
			expectedErrors: 1,
		},
		{
			name: "injection pattern detected",
			config: &RequestValidationConfig{
				Enabled:           true,
				MaxQueryValueSize: 1000,
				PreventInjection:  true,
			},
			queryString:    "p1=<script>alert('xss')</script>",
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			middleware := NewRequestValidationMiddleware(tt.config, logger)

			req := httptest.NewRequest("GET", "/test?"+tt.queryString, nil)

			errors := middleware.validateQueryParameters(req)

			assert.Len(t, errors, tt.expectedErrors)
		})
	}
}

func TestRequestValidationMiddleware_validateRequestBody(t *testing.T) {
	tests := []struct {
		name           string
		config         *RequestValidationConfig
		contentLength  int64
		expectedErrors int
	}{
		{
			name:           "valid body size",
			config:         GetDefaultRequestValidationConfig(),
			contentLength:  1024,
			expectedErrors: 0,
		},
		{
			name: "body too large",
			config: &RequestValidationConfig{
				Enabled:     true,
				MaxBodySize: 100,
			},
			contentLength:  200,
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			middleware := NewRequestValidationMiddleware(tt.config, logger)

			req := httptest.NewRequest("POST", "/test", nil)
			req.ContentLength = tt.contentLength

			errors := middleware.validateRequestBody(req)

			assert.Len(t, errors, tt.expectedErrors)
		})
	}
}

func TestRequestValidationMiddleware_parseAndValidateJSON(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		contentType   string
		expectedError bool
		expectedData  interface{}
	}{
		{
			name:          "valid JSON",
			body:          `{"name": "test", "value": 123}`,
			contentType:   "application/json",
			expectedError: false,
			expectedData: map[string]interface{}{
				"name":  "test",
				"value": float64(123),
			},
		},
		{
			name:          "invalid JSON",
			body:          `{"name": "test", "value": 123,}`,
			contentType:   "application/json",
			expectedError: true,
		},
		{
			name:          "empty body",
			body:          "",
			contentType:   "application/json",
			expectedError: false,
			expectedData:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			middleware := NewRequestValidationMiddleware(GetDefaultRequestValidationConfig(), logger)

			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			data, sanitized, err := middleware.parseAndValidateJSON(req)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedData != nil {
					assert.Equal(t, tt.expectedData, data)
				}
				if middleware.config.SanitizeInputs {
					assert.NotNil(t, sanitized)
				}
			}
		})
	}
}

func TestRequestValidationMiddleware_validateAgainstSchema(t *testing.T) {
	tests := []struct {
		name           string
		data           interface{}
		schema         *ValidationSchema
		expectedErrors int
	}{
		{
			name: "valid data",
			data: map[string]interface{}{
				"name":  "test",
				"email": "test@example.com",
			},
			schema: &ValidationSchema{
				Required: []string{"name", "email"},
				Fields: map[string]ValidationRule{
					"name":  {Type: "string", Required: true},
					"email": {Type: "string", Required: true},
				},
			},
			expectedErrors: 0,
		},
		{
			name: "missing required field",
			data: map[string]interface{}{
				"name": "test",
			},
			schema: &ValidationSchema{
				Required: []string{"name", "email"},
				Fields: map[string]ValidationRule{
					"name":  {Type: "string", Required: true},
					"email": {Type: "string", Required: true},
				},
			},
			expectedErrors: 1,
		},
		{
			name: "extra field not allowed",
			data: map[string]interface{}{
				"name":  "test",
				"email": "test@example.com",
				"extra": "field",
			},
			schema: &ValidationSchema{
				Required:   []string{"name", "email"},
				AllowExtra: false,
				Fields: map[string]ValidationRule{
					"name":  {Type: "string", Required: true},
					"email": {Type: "string", Required: true},
				},
			},
			expectedErrors: 1,
		},
		{
			name: "invalid data type",
			data: map[string]interface{}{
				"name":  "test",
				"email": 123,
			},
			schema: &ValidationSchema{
				Required: []string{"name", "email"},
				Fields: map[string]ValidationRule{
					"name":  {Type: "string", Required: true},
					"email": {Type: "string", Required: true},
				},
			},
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			middleware := NewRequestValidationMiddleware(GetDefaultRequestValidationConfig(), logger)

			errors := middleware.validateAgainstSchema(tt.data, tt.schema)

			assert.Len(t, errors, tt.expectedErrors)
		})
	}
}

func TestRequestValidationMiddleware_validateField(t *testing.T) {
	tests := []struct {
		name           string
		fieldName      string
		value          interface{}
		rule           ValidationRule
		expectedErrors int
	}{
		{
			name:      "valid string",
			fieldName: "name",
			value:     "test",
			rule: ValidationRule{
				Type:     "string",
				Required: true,
			},
			expectedErrors: 0,
		},
		{
			name:      "required field missing",
			fieldName: "name",
			value:     "",
			rule: ValidationRule{
				Type:        "string",
				Required:    true,
				Description: "Name is required",
			},
			expectedErrors: 1,
		},
		{
			name:      "string too short",
			fieldName: "name",
			value:     "ab",
			rule: ValidationRule{
				Type:      "string",
				MinLength: 3,
			},
			expectedErrors: 1,
		},
		{
			name:      "string too long",
			fieldName: "name",
			value:     "very long name",
			rule: ValidationRule{
				Type:      "string",
				MaxLength: 5,
			},
			expectedErrors: 1,
		},
		{
			name:      "pattern mismatch",
			fieldName: "email",
			value:     "invalid-email",
			rule: ValidationRule{
				Type:    "string",
				Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
			expectedErrors: 1,
		},
		{
			name:      "enum value not allowed",
			fieldName: "status",
			value:     "invalid",
			rule: ValidationRule{
				Type: "string",
				Enum: []string{"active", "inactive"},
			},
			expectedErrors: 1,
		},
		{
			name:      "number too small",
			fieldName: "age",
			value:     15,
			rule: ValidationRule{
				Type:     "number",
				MinValue: 18,
			},
			expectedErrors: 1,
		},
		{
			name:      "number too large",
			fieldName: "score",
			value:     150,
			rule: ValidationRule{
				Type:     "number",
				MaxValue: 100,
			},
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			middleware := NewRequestValidationMiddleware(GetDefaultRequestValidationConfig(), logger)

			errors := middleware.validateField(tt.fieldName, tt.value, tt.rule)

			assert.Len(t, errors, tt.expectedErrors)
		})
	}
}

func TestRequestValidationMiddleware_sanitizeData(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		expected map[string]interface{}
	}{
		{
			name: "sanitize string values",
			data: map[string]interface{}{
				"name":  "  test<script>alert('xss')</script>  ",
				"email": "test@example.com",
			},
			expected: map[string]interface{}{
				"name":  "test&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
				"email": "test@example.com",
			},
		},
		{
			name: "sanitize nested data",
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "<script>alert('xss')</script>",
				},
				"tags": []interface{}{"tag1", "<script>alert('xss')</script>"},
			},
			expected: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
				},
				"tags": []interface{}{"tag1", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			middleware := NewRequestValidationMiddleware(GetDefaultRequestValidationConfig(), logger)

			sanitized := middleware.sanitizeData(tt.data)

			assert.Equal(t, tt.expected, sanitized)
		})
	}
}

func TestRequestValidationMiddleware_containsInjectionPattern(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult bool
	}{
		{
			name:           "no injection pattern",
			input:          "normal text",
			expectedResult: false,
		},
		{
			name:           "SQL injection pattern",
			input:          "SELECT * FROM users",
			expectedResult: true,
		},
		{
			name:           "XSS pattern",
			input:          "<script>alert('xss')</script>",
			expectedResult: true,
		},
		{
			name:           "JavaScript pattern",
			input:          "javascript:alert('xss')",
			expectedResult: true,
		},
		{
			name:           "VBScript pattern",
			input:          "vbscript:alert('xss')",
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			middleware := NewRequestValidationMiddleware(GetDefaultRequestValidationConfig(), logger)

			result := middleware.containsInjectionPattern(tt.input)

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestGetValidatedData(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected interface{}
	}{
		{
			name:     "no validated data",
			ctx:      context.Background(),
			expected: nil,
		},
		{
			name: "with validated data",
			ctx: context.WithValue(context.Background(), "validated_data", map[string]interface{}{
				"name": "test",
			}),
			expected: map[string]interface{}{
				"name": "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetValidatedData(tt.ctx)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSanitizedData(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected map[string]interface{}
	}{
		{
			name:     "no sanitized data",
			ctx:      context.Background(),
			expected: nil,
		},
		{
			name: "with sanitized data",
			ctx: context.WithValue(context.Background(), "sanitized_data", map[string]interface{}{
				"name": "sanitized_test",
			}),
			expected: map[string]interface{}{
				"name": "sanitized_test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSanitizedData(tt.ctx)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigurationFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func() *RequestValidationConfig
		expected *RequestValidationConfig
	}{
		{
			name:     "default config",
			function: GetDefaultRequestValidationConfig,
			expected: &RequestValidationConfig{
				Enabled:               true,
				MaxBodySize:           10 * 1024 * 1024,
				MaxQueryParams:        50,
				MaxQueryValueSize:     1000,
				AllowedContentTypes:   []string{"application/json", "application/x-www-form-urlencoded", "multipart/form-data"},
				RequireContentType:    false,
				PathRules:             make(map[string]*ValidationSchema),
				SanitizeInputs:        true,
				PreventInjection:      true,
				LogValidationFailures: true,
				CacheSchemas:          true,
				ValidationTimeout:     5 * time.Second,
				DetailedErrors:        true,
				StopOnFirstError:      false,
			},
		},
		{
			name:     "strict config",
			function: GetStrictRequestValidationConfig,
			expected: &RequestValidationConfig{
				Enabled:               true,
				MaxBodySize:           1 * 1024 * 1024,
				MaxQueryParams:        20,
				MaxQueryValueSize:     500,
				AllowedContentTypes:   []string{"application/json"},
				RequireContentType:    true,
				PathRules:             make(map[string]*ValidationSchema),
				SanitizeInputs:        true,
				PreventInjection:      true,
				LogValidationFailures: true,
				CacheSchemas:          true,
				ValidationTimeout:     3 * time.Second,
				DetailedErrors:        false,
				StopOnFirstError:      true,
			},
		},
		{
			name:     "permissive config",
			function: GetPermissiveRequestValidationConfig,
			expected: &RequestValidationConfig{
				Enabled:               true,
				MaxBodySize:           50 * 1024 * 1024,
				MaxQueryParams:        100,
				MaxQueryValueSize:     2000,
				AllowedContentTypes:   []string{"application/json", "application/x-www-form-urlencoded", "multipart/form-data", "text/plain"},
				RequireContentType:    false,
				PathRules:             make(map[string]*ValidationSchema),
				SanitizeInputs:        false,
				PreventInjection:      false,
				LogValidationFailures: false,
				CacheSchemas:          false,
				ValidationTimeout:     10 * time.Second,
				DetailedErrors:        true,
				StopOnFirstError:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.function()

			assert.Equal(t, tt.expected.Enabled, config.Enabled)
			assert.Equal(t, tt.expected.MaxBodySize, config.MaxBodySize)
			assert.Equal(t, tt.expected.MaxQueryParams, config.MaxQueryParams)
			assert.Equal(t, tt.expected.MaxQueryValueSize, config.MaxQueryValueSize)
			assert.Equal(t, tt.expected.RequireContentType, config.RequireContentType)
			assert.Equal(t, tt.expected.SanitizeInputs, config.SanitizeInputs)
			assert.Equal(t, tt.expected.PreventInjection, config.PreventInjection)
			assert.Equal(t, tt.expected.LogValidationFailures, config.LogValidationFailures)
			assert.Equal(t, tt.expected.CacheSchemas, config.CacheSchemas)
			assert.Equal(t, tt.expected.ValidationTimeout, config.ValidationTimeout)
			assert.Equal(t, tt.expected.DetailedErrors, config.DetailedErrors)
			assert.Equal(t, tt.expected.StopOnFirstError, config.StopOnFirstError)
		})
	}
}
