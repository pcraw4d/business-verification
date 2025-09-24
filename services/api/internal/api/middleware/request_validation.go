package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ValidationRule represents a validation rule
type ValidationRule struct {
	Field       string
	Type        string
	Required    bool
	MinLength   int
	MaxLength   int
	MinValue    float64
	MaxValue    float64
	Pattern     string
	Enum        []string
	CustomFunc  func(interface{}) error
	Description string
}

// ValidationSchema defines validation rules for a request
type ValidationSchema struct {
	Fields     map[string]ValidationRule
	Required   []string
	MaxSize    int64
	AllowExtra bool
}

// RequestValidationConfig holds validation configuration
type RequestValidationConfig struct {
	// General Settings
	Enabled           bool  `json:"enabled" yaml:"enabled"`
	MaxBodySize       int64 `json:"max_body_size" yaml:"max_body_size"`
	MaxQueryParams    int   `json:"max_query_params" yaml:"max_query_params"`
	MaxQueryValueSize int   `json:"max_query_value_size" yaml:"max_query_value_size"`

	// Content Type Validation
	AllowedContentTypes []string `json:"allowed_content_types" yaml:"allowed_content_types"`
	RequireContentType  bool     `json:"require_content_type" yaml:"require_content_type"`

	// Path-based Validation
	PathRules map[string]*ValidationSchema `json:"path_rules" yaml:"path_rules"`

	// Security Settings
	SanitizeInputs        bool `json:"sanitize_inputs" yaml:"sanitize_inputs"`
	PreventInjection      bool `json:"prevent_injection" yaml:"prevent_injection"`
	LogValidationFailures bool `json:"log_validation_failures" yaml:"log_validation_failures"`

	// Performance Settings
	CacheSchemas      bool          `json:"cache_schemas" yaml:"cache_schemas"`
	ValidationTimeout time.Duration `json:"validation_timeout" yaml:"validation_timeout"`

	// Error Handling
	DetailedErrors   bool `json:"detailed_errors" yaml:"detailed_errors"`
	StopOnFirstError bool `json:"stop_on_first_error" yaml:"stop_on_first_error"`
}

// RequestValidationError represents a validation error
type RequestValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value,omitempty"`
	Rule    string      `json:"rule"`
	Message string      `json:"message"`
}

// RequestValidationResult holds validation results
type RequestValidationResult struct {
	Valid     bool                     `json:"valid"`
	Errors    []RequestValidationError `json:"errors,omitempty"`
	Data      interface{}              `json:"data,omitempty"`
	Sanitized map[string]interface{}   `json:"sanitized,omitempty"`
}

// RequestValidationMiddleware provides comprehensive request validation
type RequestValidationMiddleware struct {
	config      *RequestValidationConfig
	logger      *zap.Logger
	schemaCache map[string]*ValidationSchema
}

// NewRequestValidationMiddleware creates a new request validation middleware
func NewRequestValidationMiddleware(config *RequestValidationConfig, logger *zap.Logger) *RequestValidationMiddleware {
	if config == nil {
		config = GetDefaultRequestValidationConfig()
	}

	return &RequestValidationMiddleware{
		config:      config,
		logger:      logger,
		schemaCache: make(map[string]*ValidationSchema),
	}
}

// Middleware applies request validation to HTTP requests
func (m *RequestValidationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Create validation context with timeout
		ctx := r.Context()
		if m.config.ValidationTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, m.config.ValidationTimeout)
			defer cancel()
			r = r.WithContext(ctx)
		}

		// Validate request
		result, err := m.ValidateRequest(r)
		if err != nil {
			m.handleValidationError(w, r, err)
			return
		}

		if !result.Valid {
			m.handleValidationFailure(w, r, result)
			return
		}

		// Add validated data to context
		if result.Data != nil {
			ctx = context.WithValue(ctx, "validated_data", result.Data)
			r = r.WithContext(ctx)
		}

		// Add sanitized data to context
		if result.Sanitized != nil {
			ctx = context.WithValue(ctx, "sanitized_data", result.Sanitized)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// ValidateRequest validates an HTTP request
func (m *RequestValidationMiddleware) ValidateRequest(r *http.Request) (*RequestValidationResult, error) {
	result := &RequestValidationResult{
		Valid:  true,
		Errors: []RequestValidationError{},
	}

	// Validate content type
	if err := m.validateContentType(r); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, RequestValidationError{
			Field:   "Content-Type",
			Rule:    "content_type",
			Message: err.Error(),
		})
		if m.config.StopOnFirstError {
			return result, nil
		}
	}

	// Validate query parameters
	if queryErrors := m.validateQueryParameters(r); len(queryErrors) > 0 {
		result.Valid = false
		result.Errors = append(result.Errors, queryErrors...)
		if m.config.StopOnFirstError {
			return result, nil
		}
	}

	// Validate request body for POST/PUT/PATCH requests
	if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
		if bodyErrors := m.validateRequestBody(r); len(bodyErrors) > 0 {
			result.Valid = false
			result.Errors = append(result.Errors, bodyErrors...)
		} else {
			// Parse and validate JSON body
			if data, sanitized, err := m.parseAndValidateBody(r); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, RequestValidationError{
					Field:   "body",
					Rule:    "json_parse",
					Message: err.Error(),
				})
			} else {
				result.Data = data
				result.Sanitized = sanitized
			}
		}
	}

	return result, nil
}

// validateContentType validates the request content type
func (m *RequestValidationMiddleware) validateContentType(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")

	if contentType == "" {
		if m.config.RequireContentType {
			return fmt.Errorf("missing Content-Type header")
		}
		return nil
	}

	// Extract media type
	mediaType := strings.Split(contentType, ";")[0]
	mediaType = strings.TrimSpace(mediaType)

	// Check against allowed content types
	if len(m.config.AllowedContentTypes) > 0 {
		allowed := false
		for _, allowedType := range m.config.AllowedContentTypes {
			if mediaType == allowedType {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("unsupported content type: %s", mediaType)
		}
	}

	return nil
}

// validateQueryParameters validates query parameters
func (m *RequestValidationMiddleware) validateQueryParameters(r *http.Request) []RequestValidationError {
	var errors []RequestValidationError
	queryParams := r.URL.Query()

	// Check number of query parameters
	if len(queryParams) > m.config.MaxQueryParams {
		errors = append(errors, RequestValidationError{
			Field:   "query_params",
			Rule:    "max_count",
			Message: fmt.Sprintf("too many query parameters: %d (max %d)", len(queryParams), m.config.MaxQueryParams),
		})
	}

	// Validate individual query parameters
	for key, values := range queryParams {
		for i, value := range values {
			// Check value size
			if len(value) > m.config.MaxQueryValueSize {
				errors = append(errors, RequestValidationError{
					Field:   fmt.Sprintf("query.%s[%d]", key, i),
					Value:   value,
					Rule:    "max_length",
					Message: fmt.Sprintf("query parameter value too long: %d characters (max %d)", len(value), m.config.MaxQueryValueSize),
				})
			}

			// Check for injection attempts
			if m.config.PreventInjection {
				if m.containsInjectionPattern(value) {
					errors = append(errors, RequestValidationError{
						Field:   fmt.Sprintf("query.%s[%d]", key, i),
						Value:   "[MASKED]",
						Rule:    "injection_prevention",
						Message: "potentially harmful content detected",
					})
				}
			}
		}
	}

	return errors
}

// validateRequestBody validates the request body
func (m *RequestValidationMiddleware) validateRequestBody(r *http.Request) []RequestValidationError {
	var errors []RequestValidationError

	// Check body size
	if r.ContentLength > m.config.MaxBodySize {
		errors = append(errors, RequestValidationError{
			Field:   "body",
			Rule:    "max_size",
			Message: fmt.Sprintf("request body too large: %d bytes (max %d)", r.ContentLength, m.config.MaxBodySize),
		})
	}

	return errors
}

// parseAndValidateBody parses and validates the request body
func (m *RequestValidationMiddleware) parseAndValidateBody(r *http.Request) (interface{}, map[string]interface{}, error) {
	contentType := r.Header.Get("Content-Type")

	// Handle JSON content
	if strings.Contains(contentType, "application/json") {
		return m.parseAndValidateJSON(r)
	}

	// Handle form data
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return m.parseAndValidateForm(r)
	}

	// Handle multipart form data
	if strings.Contains(contentType, "multipart/form-data") {
		return m.parseAndValidateMultipartForm(r)
	}

	return nil, nil, fmt.Errorf("unsupported content type for body parsing: %s", contentType)
}

// parseAndValidateJSON parses and validates JSON body
func (m *RequestValidationMiddleware) parseAndValidateJSON(r *http.Request) (interface{}, map[string]interface{}, error) {
	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read request body: %w", err)
	}

	// Restore body for subsequent handlers
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// Skip validation for empty bodies
	if len(body) == 0 {
		return nil, nil, nil
	}

	// Parse JSON
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Get schema for this path
	schema := m.getSchemaForPath(r.URL.Path, r.Method)
	if schema != nil {
		// Validate against schema
		if errors := m.validateAgainstSchema(data, schema); len(errors) > 0 {
			return nil, nil, fmt.Errorf("validation failed: %v", errors)
		}
	}

	// Sanitize data if enabled
	var sanitized map[string]interface{}
	if m.config.SanitizeInputs {
		sanitized = m.sanitizeData(data)
	}

	return data, sanitized, nil
}

// parseAndValidateForm parses and validates form data
func (m *RequestValidationMiddleware) parseAndValidateForm(r *http.Request) (interface{}, map[string]interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, nil, fmt.Errorf("failed to parse form: %w", err)
	}

	// Convert form data to map
	data := make(map[string]interface{})
	for key, values := range r.Form {
		if len(values) == 1 {
			data[key] = values[0]
		} else {
			data[key] = values
		}
	}

	// Get schema for this path
	schema := m.getSchemaForPath(r.URL.Path, r.Method)
	if schema != nil {
		// Validate against schema
		if errors := m.validateAgainstSchema(data, schema); len(errors) > 0 {
			return nil, nil, fmt.Errorf("validation failed: %v", errors)
		}
	}

	// Sanitize data if enabled
	var sanitized map[string]interface{}
	if m.config.SanitizeInputs {
		sanitized = m.sanitizeData(data)
	}

	return data, sanitized, nil
}

// parseAndValidateMultipartForm parses and validates multipart form data
func (m *RequestValidationMiddleware) parseAndValidateMultipartForm(r *http.Request) (interface{}, map[string]interface{}, error) {
	if err := r.ParseMultipartForm(m.config.MaxBodySize); err != nil {
		return nil, nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	// Convert form data to map
	data := make(map[string]interface{})
	for key, values := range r.MultipartForm.Value {
		if len(values) == 1 {
			data[key] = values[0]
		} else {
			data[key] = values
		}
	}

	// Add file information
	if len(r.MultipartForm.File) > 0 {
		fileInfo := make(map[string]interface{})
		for key, files := range r.MultipartForm.File {
			fileInfo[key] = len(files)
		}
		data["_files"] = fileInfo
	}

	// Get schema for this path
	schema := m.getSchemaForPath(r.URL.Path, r.Method)
	if schema != nil {
		// Validate against schema
		if errors := m.validateAgainstSchema(data, schema); len(errors) > 0 {
			return nil, nil, fmt.Errorf("validation failed: %v", errors)
		}
	}

	// Sanitize data if enabled
	var sanitized map[string]interface{}
	if m.config.SanitizeInputs {
		sanitized = m.sanitizeData(data)
	}

	return data, sanitized, nil
}

// validateAgainstSchema validates data against a schema
func (m *RequestValidationMiddleware) validateAgainstSchema(data interface{}, schema *ValidationSchema) []ValidationError {
	var errors []ValidationError

	// Convert data to map if possible
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return []ValidationError{{
			Field:   "data",
			Rule:    "type",
			Message: "expected object",
		}}
	}

	// Check required fields
	for _, requiredField := range schema.Required {
		if value, exists := dataMap[requiredField]; !exists || value == nil {
			errors = append(errors, ValidationError{
				Field:   requiredField,
				Rule:    "required",
				Message: "field is required",
			})
		}
	}

	// Validate each field
	for fieldName, value := range dataMap {
		if rule, exists := schema.Fields[fieldName]; exists {
			if fieldErrors := m.validateField(fieldName, value, rule); len(fieldErrors) > 0 {
				errors = append(errors, fieldErrors...)
			}
		} else if !schema.AllowExtra {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Rule:    "extra_field",
				Message: "extra field not allowed",
			})
		}
	}

	return errors
}

// validateField validates a single field
func (m *RequestValidationMiddleware) validateField(fieldName string, value interface{}, rule ValidationRule) []ValidationError {
	var errors []ValidationError

	// Check if field is required
	if rule.Required {
		if value == nil || (reflect.ValueOf(value).Kind() == reflect.String && value.(string) == "") {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Rule:    "required",
				Message: rule.Description,
			})
			return errors
		}
	}

	// Skip validation for nil values if not required
	if value == nil {
		return errors
	}

	// Type validation
	if rule.Type != "" {
		if typeError := m.validateType(fieldName, value, rule.Type); typeError != nil {
			errors = append(errors, *typeError)
		}
	}

	// String validations
	if strValue, ok := value.(string); ok {
		// Length validation
		if rule.MinLength > 0 && len(strValue) < rule.MinLength {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Value:   strValue,
				Rule:    "min_length",
				Message: fmt.Sprintf("minimum length is %d characters", rule.MinLength),
			})
		}

		if rule.MaxLength > 0 && len(strValue) > rule.MaxLength {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Value:   strValue,
				Rule:    "max_length",
				Message: fmt.Sprintf("maximum length is %d characters", rule.MaxLength),
			})
		}

		// Pattern validation
		if rule.Pattern != "" {
			if matched, _ := regexp.MatchString(rule.Pattern, strValue); !matched {
				errors = append(errors, ValidationError{
					Field:   fieldName,
					Value:   strValue,
					Rule:    "pattern",
					Message: fmt.Sprintf("must match pattern: %s", rule.Pattern),
				})
			}
		}

		// Enum validation
		if len(rule.Enum) > 0 {
			found := false
			for _, enumValue := range rule.Enum {
				if strValue == enumValue {
					found = true
					break
				}
			}
			if !found {
				errors = append(errors, ValidationError{
					Field:   fieldName,
					Value:   strValue,
					Rule:    "enum",
					Message: fmt.Sprintf("must be one of: %v", rule.Enum),
				})
			}
		}
	}

	// Numeric validations
	if numValue, ok := m.getNumericValue(value); ok {
		if rule.MinValue != 0 && numValue < rule.MinValue {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Value:   value,
				Rule:    "min_value",
				Message: fmt.Sprintf("minimum value is %f", rule.MinValue),
			})
		}

		if rule.MaxValue != 0 && numValue > rule.MaxValue {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Value:   value,
				Rule:    "max_value",
				Message: fmt.Sprintf("maximum value is %f", rule.MaxValue),
			})
		}
	}

	// Custom validation
	if rule.CustomFunc != nil {
		if err := rule.CustomFunc(value); err != nil {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Value:   value,
				Rule:    "custom",
				Message: err.Error(),
			})
		}
	}

	return errors
}

// validateType validates the type of a value
func (m *RequestValidationMiddleware) validateType(fieldName string, value interface{}, expectedType string) *ValidationError {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return &ValidationError{
				Field:   fieldName,
				Value:   value,
				Rule:    "type",
				Message: "expected string",
			}
		}
	case "number":
		if _, ok := m.getNumericValue(value); !ok {
			return &ValidationError{
				Field:   fieldName,
				Value:   value,
				Rule:    "type",
				Message: "expected number",
			}
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return &ValidationError{
				Field:   fieldName,
				Value:   value,
				Rule:    "type",
				Message: "expected boolean",
			}
		}
	case "array":
		if reflect.TypeOf(value).Kind() != reflect.Slice && reflect.TypeOf(value).Kind() != reflect.Array {
			return &ValidationError{
				Field:   fieldName,
				Value:   value,
				Rule:    "type",
				Message: "expected array",
			}
		}
	case "object":
		if reflect.TypeOf(value).Kind() != reflect.Map {
			return &ValidationError{
				Field:   fieldName,
				Value:   value,
				Rule:    "type",
				Message: "expected object",
			}
		}
	}
	return nil
}

// getNumericValue extracts numeric value from interface{}
func (m *RequestValidationMiddleware) getNumericValue(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// getSchemaForPath gets the validation schema for a specific path and method
func (m *RequestValidationMiddleware) getSchemaForPath(path, method string) *ValidationSchema {
	// Create cache key
	cacheKey := fmt.Sprintf("%s:%s", method, path)

	// Check cache first
	if m.config.CacheSchemas {
		if schema, exists := m.schemaCache[cacheKey]; exists {
			return schema
		}
	}

	// Find matching schema
	for pathPattern, schema := range m.config.PathRules {
		if m.pathMatches(path, pathPattern) {
			// Cache the result
			if m.config.CacheSchemas {
				m.schemaCache[cacheKey] = schema
			}
			return schema
		}
	}

	return nil
}

// pathMatches checks if a path matches a pattern
func (m *RequestValidationMiddleware) pathMatches(path, pattern string) bool {
	// Simple prefix matching for now
	// Could be enhanced with regex patterns
	return strings.HasPrefix(path, pattern)
}

// sanitizeData sanitizes input data
func (m *RequestValidationMiddleware) sanitizeData(data interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	if dataMap, ok := data.(map[string]interface{}); ok {
		for key, value := range dataMap {
			sanitized[key] = m.sanitizeValue(value)
		}
	}

	return sanitized
}

// sanitizeValue sanitizes a single value
func (m *RequestValidationMiddleware) sanitizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return m.sanitizeString(v)
	case []interface{}:
		sanitized := make([]interface{}, len(v))
		for i, item := range v {
			sanitized[i] = m.sanitizeValue(item)
		}
		return sanitized
	case map[string]interface{}:
		sanitized := make(map[string]interface{})
		for key, item := range v {
			sanitized[key] = m.sanitizeValue(item)
		}
		return sanitized
	default:
		return value
	}
}

// sanitizeString sanitizes a string value
func (m *RequestValidationMiddleware) sanitizeString(s string) string {
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")

	// Trim whitespace
	s = strings.TrimSpace(s)

	// Basic HTML entity encoding
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")

	return s
}

// containsInjectionPattern checks if a string contains injection patterns
func (m *RequestValidationMiddleware) containsInjectionPattern(s string) bool {
	patterns := []string{
		`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`,
		`(?i)(script|javascript|vbscript|onload|onerror|onclick)`,
		`['\";]`,
		`<script`,
		`javascript:`,
		`vbscript:`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, s); matched {
			return true
		}
	}

	return false
}

// handleValidationError handles validation errors
func (m *RequestValidationMiddleware) handleValidationError(w http.ResponseWriter, r *http.Request, err error) {
	// Log the error
	if m.config.LogValidationFailures {
		m.logger.Warn("Request validation error",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.Error(err),
		)
	}

	// Create validation error
	validationErr := CreateValidationError("Request validation failed", err.Error())

	// Use error handling middleware if available
	if errorHandler := r.Context().Value("error_handler"); errorHandler != nil {
		if handler, ok := errorHandler.(*ErrorHandlingMiddleware); ok {
			handler.handleErrorResponse(w, r, http.StatusBadRequest, validationErr)
			return
		}
	}

	// Fallback error response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "Request validation failed",
		"details": err.Error(),
		"success": false,
	})
}

// handleValidationFailure handles validation failures
func (m *RequestValidationMiddleware) handleValidationFailure(w http.ResponseWriter, r *http.Request, result *RequestValidationResult) {
	// Log the validation failure
	if m.config.LogValidationFailures {
		m.logger.Warn("Request validation failed",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.Int("error_count", len(result.Errors)),
			zap.Any("errors", result.Errors),
		)
	}

	// Create validation error
	validationErr := CreateValidationError("Request validation failed", fmt.Sprintf("%d validation errors", len(result.Errors)))

	// Use error handling middleware if available
	if errorHandler := r.Context().Value("error_handler"); errorHandler != nil {
		if handler, ok := errorHandler.(*ErrorHandlingMiddleware); ok {
			handler.handleErrorResponse(w, r, http.StatusBadRequest, validationErr)
			return
		}
	}

	// Fallback error response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "Request validation failed",
		"errors":  result.Errors,
		"success": false,
	})
}

// GetValidatedData gets validated data from context
func GetValidatedData(ctx context.Context) interface{} {
	if data := ctx.Value("validated_data"); data != nil {
		return data
	}
	return nil
}

// GetSanitizedData gets sanitized data from context
func GetSanitizedData(ctx context.Context) map[string]interface{} {
	if data := ctx.Value("sanitized_data"); data != nil {
		if sanitized, ok := data.(map[string]interface{}); ok {
			return sanitized
		}
	}
	return nil
}

// GetDefaultRequestValidationConfig returns a default validation configuration
func GetDefaultRequestValidationConfig() *RequestValidationConfig {
	return &RequestValidationConfig{
		Enabled:               true,
		MaxBodySize:           10 * 1024 * 1024, // 10MB
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
	}
}

// GetStrictRequestValidationConfig returns a strict validation configuration
func GetStrictRequestValidationConfig() *RequestValidationConfig {
	return &RequestValidationConfig{
		Enabled:               true,
		MaxBodySize:           1 * 1024 * 1024, // 1MB
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
	}
}

// GetPermissiveRequestValidationConfig returns a permissive validation configuration
func GetPermissiveRequestValidationConfig() *RequestValidationConfig {
	return &RequestValidationConfig{
		Enabled:               true,
		MaxBodySize:           50 * 1024 * 1024, // 50MB
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
	}
}
