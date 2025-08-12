package validators

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Common validation errors
var (
	ErrInvalidPath = fmt.Errorf("invalid path: potential path traversal detected")
)

// Validator provides validation functionality
type Validator struct {
	errors ValidationErrors
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// Validate validates a struct using struct tags
func Validate(v interface{}) error {
	validator := NewValidator()
	return validator.ValidateStruct(v)
}

// ValidateStruct validates a struct using struct tags
func (v *Validator) ValidateStruct(s interface{}) error {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("validate: expected struct, got %s", val.Kind())
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Get validation tags
		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			continue
		}

		// Parse validation rules
		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			if err := v.validateField(field, fieldType.Name, rule); err != nil {
				v.errors = append(v.errors, *err)
			}
		}
	}

	if len(v.errors) > 0 {
		return v.errors
	}

	return nil
}

// validateField validates a single field based on the rule
func (v *Validator) validateField(field reflect.Value, fieldName, rule string) *ValidationError {
	switch rule {
	case "required":
		return v.validateRequired(field, fieldName)
	case "email":
		return v.validateEmail(field, fieldName)
	case "url":
		return v.validateURL(field, fieldName)
	case "min":
		return v.validateMin(field, fieldName, rule)
	case "max":
		return v.validateMax(field, fieldName, rule)
	case "len":
		return v.validateLen(field, fieldName, rule)
	default:
		// Handle custom validation rules
		if strings.HasPrefix(rule, "min=") {
			return v.validateMin(field, fieldName, rule)
		}
		if strings.HasPrefix(rule, "max=") {
			return v.validateMax(field, fieldName, rule)
		}
		if strings.HasPrefix(rule, "len=") {
			return v.validateLen(field, fieldName, rule)
		}
	}

	return nil
}

// validateRequired validates that a field is not empty
func (v *Validator) validateRequired(field reflect.Value, fieldName string) *ValidationError {
	if field.IsZero() {
		return &ValidationError{
			Field:   fieldName,
			Message: "field is required",
		}
	}
	return nil
}

// validateEmail validates email format
func (v *Validator) validateEmail(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return nil
	}

	email := field.String()
	if email == "" {
		return nil // Empty email is valid (use required if needed)
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return &ValidationError{
			Field:   fieldName,
			Message: "invalid email format",
			Value:   email,
		}
	}
	return nil
}

// validateURL validates URL format
func (v *Validator) validateURL(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return nil
	}

	url := field.String()
	if url == "" {
		return nil // Empty URL is valid (use required if needed)
	}

	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(url) {
		return &ValidationError{
			Field:   fieldName,
			Message: "invalid URL format",
			Value:   url,
		}
	}
	return nil
}

// validateMin validates minimum value/length
func (v *Validator) validateMin(field reflect.Value, fieldName, rule string) *ValidationError {
	// Extract min value from rule (e.g., "min=5")
	parts := strings.Split(rule, "=")
	if len(parts) != 2 {
		return nil
	}

	minValue := parts[1]

	switch field.Kind() {
	case reflect.String:
		if len(field.String()) < len(minValue) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("minimum length is %s", minValue),
				Value:   field.String(),
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Parse min value and compare
		// This is a simplified implementation
		if field.Int() < 0 { // Assuming min=0 for now
			return &ValidationError{
				Field:   fieldName,
				Message: "value must be non-negative",
				Value:   field.Int(),
			}
		}
	}

	return nil
}

// validateMax validates maximum value/length
func (v *Validator) validateMax(field reflect.Value, fieldName, rule string) *ValidationError {
	// Extract max value from rule (e.g., "max=100")
	parts := strings.Split(rule, "=")
	if len(parts) != 2 {
		return nil
	}

	maxValue := parts[1]

	switch field.Kind() {
	case reflect.String:
		if len(field.String()) > len(maxValue) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("maximum length is %s", maxValue),
				Value:   field.String(),
			}
		}
	}

	return nil
}

// validateLen validates exact length
func (v *Validator) validateLen(field reflect.Value, fieldName, rule string) *ValidationError {
	// Extract length value from rule (e.g., "len=10")
	parts := strings.Split(rule, "=")
	if len(parts) != 2 {
		return nil
	}

	expectedLen := parts[1]

	switch field.Kind() {
	case reflect.String:
		if len(field.String()) != len(expectedLen) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("length must be %s", expectedLen),
				Value:   field.String(),
			}
		}
	}

	return nil
}

// IsValid checks if the validator has any errors
func (v *Validator) IsValid() bool {
	return len(v.errors) == 0
}

// Errors returns the validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}
