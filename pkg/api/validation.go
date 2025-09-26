package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// RequestValidator validates API requests
type RequestValidator struct {
	rules map[string]ValidationRule
}

// ValidationRule defines a validation rule
type ValidationRule struct {
	Required bool
	Type     string
	Min      interface{}
	Max      interface{}
	Pattern  string
	Custom   func(interface{}) error
}

// NewRequestValidator creates a new request validator
func NewRequestValidator() *RequestValidator {
	return &RequestValidator{
		rules: make(map[string]ValidationRule),
	}
}

// AddRule adds a validation rule for a field
func (rv *RequestValidator) AddRule(field string, rule ValidationRule) {
	rv.rules[field] = rule
}

// ValidateRequest validates a request against the defined rules
func (rv *RequestValidator) ValidateRequest(data map[string]interface{}) ValidationResult {
	result := ValidationResult{IsValid: true}

	for field, rule := range rv.rules {
		value, exists := data[field]

		// Check if required field is present
		if rule.Required && !exists {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("field '%s' is required", field),
			})
			continue
		}

		// Skip validation if field is not present and not required
		if !exists {
			continue
		}

		// Validate field value
		if err := rv.validateField(field, value, rule); err != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   field,
				Message: err.Error(),
				Value:   value,
			})
		}
	}

	return result
}

// validateField validates a single field
func (rv *RequestValidator) validateField(field string, value interface{}, rule ValidationRule) error {
	// Check type
	if rule.Type != "" {
		if err := rv.validateType(value, rule.Type); err != nil {
			return fmt.Errorf("invalid type for field '%s': %v", field, err)
		}
	}

	// Check minimum value
	if rule.Min != nil {
		if err := rv.validateMin(value, rule.Min); err != nil {
			return fmt.Errorf("value below minimum for field '%s': %v", field, err)
		}
	}

	// Check maximum value
	if rule.Max != nil {
		if err := rv.validateMax(value, rule.Max); err != nil {
			return fmt.Errorf("value above maximum for field '%s': %v", field, err)
		}
	}

	// Check pattern
	if rule.Pattern != "" {
		if err := rv.validatePattern(value, rule.Pattern); err != nil {
			return fmt.Errorf("pattern validation failed for field '%s': %v", field, err)
		}
	}

	// Custom validation
	if rule.Custom != nil {
		if err := rule.Custom(value); err != nil {
			return fmt.Errorf("custom validation failed for field '%s': %v", field, err)
		}
	}

	return nil
}

// validateType validates the type of a value
func (rv *RequestValidator) validateType(value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "int":
		if _, ok := value.(int); !ok {
			return fmt.Errorf("expected int, got %T", value)
		}
	case "float64":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("expected float64, got %T", value)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
	case "array":
		if reflect.TypeOf(value).Kind() != reflect.Slice {
			return fmt.Errorf("expected array, got %T", value)
		}
	case "object":
		if reflect.TypeOf(value).Kind() != reflect.Map {
			return fmt.Errorf("expected object, got %T", value)
		}
	}
	return nil
}

// validateMin validates minimum value
func (rv *RequestValidator) validateMin(value interface{}, min interface{}) error {
	switch v := value.(type) {
	case int:
		if minInt, ok := min.(int); ok && v < minInt {
			return fmt.Errorf("value %d is below minimum %d", v, minInt)
		}
	case float64:
		if minFloat, ok := min.(float64); ok && v < minFloat {
			return fmt.Errorf("value %f is below minimum %f", v, minFloat)
		}
	case string:
		if minLen, ok := min.(int); ok && len(v) < minLen {
			return fmt.Errorf("string length %d is below minimum %d", len(v), minLen)
		}
	}
	return nil
}

// validateMax validates maximum value
func (rv *RequestValidator) validateMax(value interface{}, max interface{}) error {
	switch v := value.(type) {
	case int:
		if maxInt, ok := max.(int); ok && v > maxInt {
			return fmt.Errorf("value %d is above maximum %d", v, maxInt)
		}
	case float64:
		if maxFloat, ok := max.(float64); ok && v > maxFloat {
			return fmt.Errorf("value %f is above maximum %f", v, maxFloat)
		}
	case string:
		if maxLen, ok := max.(int); ok && len(v) > maxLen {
			return fmt.Errorf("string length %d is above maximum %d", len(v), maxLen)
		}
	}
	return nil
}

// validatePattern validates pattern matching
func (rv *RequestValidator) validatePattern(value interface{}, pattern string) error {
	if str, ok := value.(string); ok {
		// Simple pattern matching - in production, use regex
		if !strings.Contains(str, pattern) {
			return fmt.Errorf("string does not match pattern")
		}
	}
	return nil
}

// ValidationMiddleware provides request validation middleware
func (rv *RequestValidator) ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only validate POST, PUT, PATCH requests
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			// Parse request body
			var data map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			// Validate request
			result := rv.ValidateRequest(data)
			if !result.IsValid {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(result)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// ResponseValidator validates API responses
type ResponseValidator struct {
	schema map[string]interface{}
}

// NewResponseValidator creates a new response validator
func NewResponseValidator() *ResponseValidator {
	return &ResponseValidator{
		schema: make(map[string]interface{}),
	}
}

// SetSchema sets the response schema
func (rv *ResponseValidator) SetSchema(schema map[string]interface{}) {
	rv.schema = schema
}

// ValidateResponse validates a response against the schema
func (rv *ResponseValidator) ValidateResponse(data interface{}) ValidationResult {
	// This would implement response validation against a schema
	// For now, return a valid result
	return ValidationResult{IsValid: true}
}
