package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ValidationConfig holds configuration for input validation
type ValidationConfig struct {
	MaxRequestSize      int64    `json:"max_request_size"`
	AllowedOrigins      []string `json:"allowed_origins"`
	EnableSanitization  bool     `json:"enable_sanitization"`
	StrictMode          bool     `json:"strict_mode"`
	LogValidationErrors bool     `json:"log_validation_errors"`
}

// Validator provides comprehensive input validation and sanitization
type Validator struct {
	config *ValidationConfig
	logger *zap.Logger

	// Compiled regex patterns for performance
	emailPattern         *regexp.Regexp
	urlPattern           *regexp.Regexp
	phonePattern         *regexp.Regexp
	sqlInjectionPattern  *regexp.Regexp
	xssPattern           *regexp.Regexp
	pathTraversalPattern *regexp.Regexp
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Field    string      `json:"field"`
	Message  string      `json:"message"`
	Value    interface{} `json:"value,omitempty"`
	Rule     string      `json:"rule,omitempty"`
	Severity string      `json:"severity"`
	Code     string      `json:"code"`
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
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// InputValidationResult represents the result of input validation
type InputValidationResult struct {
	IsValid   bool             `json:"is_valid"`
	Errors    ValidationErrors `json:"errors,omitempty"`
	Warnings  ValidationErrors `json:"warnings,omitempty"`
	Sanitized interface{}      `json:"sanitized,omitempty"`
	Duration  time.Duration    `json:"duration"`
}

// NewValidator creates a new validator with the given configuration
func NewValidator(config *ValidationConfig, logger *zap.Logger) *Validator {
	if config == nil {
		config = &ValidationConfig{
			MaxRequestSize:      10 * 1024 * 1024, // 10MB
			EnableSanitization:  true,
			StrictMode:          false,
			LogValidationErrors: true,
		}
	}

	// Compile regex patterns for performance
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	urlPattern := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	phonePattern := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	sqlInjectionPattern := regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|vbscript|onload|onerror|onclick)`)
	xssPattern := regexp.MustCompile(`(?i)(<script|javascript:|vbscript:|onload=|onerror=|onclick=)`)
	pathTraversalPattern := regexp.MustCompile(`\.\./|\.\.\\|/etc/|/proc/|/sys/|/dev/`)

	return &Validator{
		config:               config,
		logger:               logger,
		emailPattern:         emailPattern,
		urlPattern:           urlPattern,
		phonePattern:         phonePattern,
		sqlInjectionPattern:  sqlInjectionPattern,
		xssPattern:           xssPattern,
		pathTraversalPattern: pathTraversalPattern,
	}
}

// ValidationMiddleware creates middleware for request validation
func (v *Validator) ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create validation context
		ctx := context.WithValue(r.Context(), "validator", v)
		r = r.WithContext(ctx)

		// Validate request size
		if r.ContentLength > v.config.MaxRequestSize {
			v.logValidationError("request_size_exceeded", "Request size exceeds maximum allowed", nil)
			http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
			return
		}

		// Validate content type for POST/PUT requests
		if r.Method == "POST" || r.Method == "PUT" {
			contentType := r.Header.Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				v.logValidationError("invalid_content_type", "Content-Type must be application/json", nil)
				http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
				return
			}
		}

		// Validate origin if CORS is enabled
		if len(v.config.AllowedOrigins) > 0 {
			origin := r.Header.Get("Origin")
			if origin != "" {
				allowed := false
				for _, allowedOrigin := range v.config.AllowedOrigins {
					if origin == allowedOrigin || allowedOrigin == "*" {
						allowed = true
						break
					}
				}
				if !allowed {
					v.logValidationError("invalid_origin", "Origin not allowed", map[string]interface{}{"origin": origin})
					http.Error(w, "Origin not allowed", http.StatusForbidden)
					return
				}
			}
		}

		// Continue to next handler
		next.ServeHTTP(w, r)

		// Log validation duration
		duration := time.Since(start)
		if duration > 100*time.Millisecond {
			v.logger.Warn("validation took longer than expected",
				zap.Duration("duration", duration),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
		}
	})
}

// ValidateStruct validates a struct using struct tags and returns validation result
func (v *Validator) ValidateStruct(s interface{}) *InputValidationResult {
	start := time.Now()
	result := &InputValidationResult{
		IsValid:  true,
		Errors:   make(ValidationErrors, 0),
		Warnings: make(ValidationErrors, 0),
	}

	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:    "struct",
			Message:  "Expected struct, got " + val.Kind().String(),
			Severity: "error",
			Code:     "invalid_type",
		})
		return result
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
				if err.Severity == "error" {
					result.Errors = append(result.Errors, *err)
					result.IsValid = false
				} else {
					result.Warnings = append(result.Warnings, *err)
				}
			}
		}

		// Sanitize field if enabled
		if v.config.EnableSanitization {
			v.sanitizeField(field, fieldType.Name)
		}
	}

	result.Duration = time.Since(start)
	return result
}

// validateField validates a single field based on the given rule
func (v *Validator) validateField(field reflect.Value, fieldName, rule string) *ValidationError {
	// Parse rule (e.g., "required,email,max=255")
	ruleParts := strings.Split(rule, "=")
	ruleName := ruleParts[0]
	ruleValue := ""
	if len(ruleParts) > 1 {
		ruleValue = ruleParts[1]
	}

	switch ruleName {
	case "required":
		return v.validateRequired(field, fieldName)
	case "email":
		return v.validateEmail(field, fieldName)
	case "url":
		return v.validateURL(field, fieldName)
	case "phone":
		return v.validatePhone(field, fieldName)
	case "min":
		return v.validateMin(field, fieldName, ruleValue)
	case "max":
		return v.validateMax(field, fieldName, ruleValue)
	case "len":
		return v.validateLen(field, fieldName, ruleValue)
	case "regex":
		return v.validateRegex(field, fieldName, ruleValue)
	case "sql_injection":
		return v.validateSQLInjection(field, fieldName)
	case "xss":
		return v.validateXSS(field, fieldName)
	case "path_traversal":
		return v.validatePathTraversal(field, fieldName)
	case "alphanumeric":
		return v.validateAlphanumeric(field, fieldName)
	case "numeric":
		return v.validateNumeric(field, fieldName)
	case "date":
		return v.validateDate(field, fieldName)
	case "uuid":
		return v.validateUUID(field, fieldName)
	default:
		return &ValidationError{
			Field:    fieldName,
			Message:  "Unknown validation rule: " + ruleName,
			Rule:     rule,
			Severity: "error",
			Code:     "unknown_rule",
		}
	}
}

// validateRequired checks if a field is required and not empty
func (v *Validator) validateRequired(field reflect.Value, fieldName string) *ValidationError {
	if field.IsZero() {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Field is required",
			Rule:     "required",
			Severity: "error",
			Code:     "required_field",
		}
	}

	// Check for empty strings
	if field.Kind() == reflect.String && strings.TrimSpace(field.String()) == "" {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Field cannot be empty",
			Value:    field.String(),
			Rule:     "required",
			Severity: "error",
			Code:     "empty_string",
		}
	}

	return nil
}

// validateEmail validates email format
func (v *Validator) validateEmail(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Email validation requires string type",
			Value:    field.Interface(),
			Rule:     "email",
			Severity: "error",
			Code:     "invalid_type",
		}
	}

	email := field.String()
	if email == "" {
		return nil // Empty email is allowed unless required
	}

	if !v.emailPattern.MatchString(email) {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Invalid email format",
			Value:    email,
			Rule:     "email",
			Severity: "error",
			Code:     "invalid_email",
		}
	}

	// Additional checks for email length
	if len(email) > 254 {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Email address too long",
			Value:    email,
			Rule:     "email",
			Severity: "error",
			Code:     "email_too_long",
		}
	}

	return nil
}

// validateURL validates URL format
func (v *Validator) validateURL(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return &ValidationError{
			Field:    fieldName,
			Message:  "URL validation requires string type",
			Value:    field.Interface(),
			Rule:     "url",
			Severity: "error",
			Code:     "invalid_type",
		}
	}

	url := field.String()
	if url == "" {
		return nil // Empty URL is allowed unless required
	}

	if !v.urlPattern.MatchString(url) {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Invalid URL format",
			Value:    url,
			Rule:     "url",
			Severity: "error",
			Code:     "invalid_url",
		}
	}

	return nil
}

// validatePhone validates phone number format (E.164)
func (v *Validator) validatePhone(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Phone validation requires string type",
			Value:    field.Interface(),
			Rule:     "phone",
			Severity: "error",
			Code:     "invalid_type",
		}
	}

	phone := field.String()
	if phone == "" {
		return nil // Empty phone is allowed unless required
	}

	if !v.phonePattern.MatchString(phone) {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Invalid phone number format (E.164 required)",
			Value:    phone,
			Rule:     "phone",
			Severity: "error",
			Code:     "invalid_phone",
		}
	}

	return nil
}

// validateMin validates minimum value/length
func (v *Validator) validateMin(field reflect.Value, fieldName, minStr string) *ValidationError {
	min, err := parseNumber(minStr)
	if err != nil {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Invalid min rule value: " + minStr,
			Rule:     "min=" + minStr,
			Severity: "error",
			Code:     "invalid_rule",
		}
	}

	switch field.Kind() {
	case reflect.String:
		if len(field.String()) < int(min) {
			return &ValidationError{
				Field:    fieldName,
				Message:  fmt.Sprintf("String length must be at least %d", int(min)),
				Value:    field.String(),
				Rule:     "min=" + minStr,
				Severity: "error",
				Code:     "string_too_short",
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() < int64(min) {
			return &ValidationError{
				Field:    fieldName,
				Message:  fmt.Sprintf("Value must be at least %d", int(min)),
				Value:    field.Int(),
				Rule:     "min=" + minStr,
				Severity: "error",
				Code:     "value_too_small",
			}
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() < min {
			return &ValidationError{
				Field:    fieldName,
				Message:  fmt.Sprintf("Value must be at least %f", min),
				Value:    field.Float(),
				Rule:     "min=" + minStr,
				Severity: "error",
				Code:     "value_too_small",
			}
		}
	default:
		return &ValidationError{
			Field:    fieldName,
			Message:  "Min validation not supported for this type",
			Value:    field.Interface(),
			Rule:     "min=" + minStr,
			Severity: "error",
			Code:     "unsupported_type",
		}
	}

	return nil
}

// validateMax validates maximum value/length
func (v *Validator) validateMax(field reflect.Value, fieldName, maxStr string) *ValidationError {
	max, err := parseNumber(maxStr)
	if err != nil {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Invalid max rule value: " + maxStr,
			Rule:     "max=" + maxStr,
			Severity: "error",
			Code:     "invalid_rule",
		}
	}

	switch field.Kind() {
	case reflect.String:
		if len(field.String()) > int(max) {
			return &ValidationError{
				Field:    fieldName,
				Message:  fmt.Sprintf("String length must be at most %d", int(max)),
				Value:    field.String(),
				Rule:     "max=" + maxStr,
				Severity: "error",
				Code:     "string_too_long",
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() > int64(max) {
			return &ValidationError{
				Field:    fieldName,
				Message:  fmt.Sprintf("Value must be at most %d", int(max)),
				Value:    field.Int(),
				Rule:     "max=" + maxStr,
				Severity: "error",
				Code:     "value_too_large",
			}
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() > max {
			return &ValidationError{
				Field:    fieldName,
				Message:  fmt.Sprintf("Value must be at most %f", max),
				Value:    field.Float(),
				Rule:     "max=" + maxStr,
				Severity: "error",
				Code:     "value_too_large",
			}
		}
	default:
		return &ValidationError{
			Field:    fieldName,
			Message:  "Max validation not supported for this type",
			Value:    field.Interface(),
			Rule:     "max=" + maxStr,
			Severity: "error",
			Code:     "unsupported_type",
		}
	}

	return nil
}

// validateLen validates exact length
func (v *Validator) validateLen(field reflect.Value, fieldName, lenStr string) *ValidationError {
	expectedLen, err := parseNumber(lenStr)
	if err != nil {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Invalid len rule value: " + lenStr,
			Rule:     "len=" + lenStr,
			Severity: "error",
			Code:     "invalid_rule",
		}
	}

	switch field.Kind() {
	case reflect.String:
		if len(field.String()) != int(expectedLen) {
			return &ValidationError{
				Field:    fieldName,
				Message:  fmt.Sprintf("String length must be exactly %d", int(expectedLen)),
				Value:    field.String(),
				Rule:     "len=" + lenStr,
				Severity: "error",
				Code:     "invalid_length",
			}
		}
	case reflect.Slice, reflect.Array:
		if field.Len() != int(expectedLen) {
			return &ValidationError{
				Field:    fieldName,
				Message:  fmt.Sprintf("Array length must be exactly %d", int(expectedLen)),
				Value:    field.Len(),
				Rule:     "len=" + lenStr,
				Severity: "error",
				Code:     "invalid_length",
			}
		}
	default:
		return &ValidationError{
			Field:    fieldName,
			Message:  "Len validation not supported for this type",
			Value:    field.Interface(),
			Rule:     "len=" + lenStr,
			Severity: "error",
			Code:     "unsupported_type",
		}
	}

	return nil
}

// validateRegex validates against a regular expression
func (v *Validator) validateRegex(field reflect.Value, fieldName, pattern string) *ValidationError {
	if field.Kind() != reflect.String {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Regex validation requires string type",
			Value:    field.Interface(),
			Rule:     "regex=" + pattern,
			Severity: "error",
			Code:     "invalid_type",
		}
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Invalid regex pattern: " + err.Error(),
			Rule:     "regex=" + pattern,
			Severity: "error",
			Code:     "invalid_regex",
		}
	}

	if !re.MatchString(field.String()) {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Value does not match required pattern",
			Value:    field.String(),
			Rule:     "regex=" + pattern,
			Severity: "error",
			Code:     "pattern_mismatch",
		}
	}

	return nil
}

// validateSQLInjection checks for SQL injection patterns
func (v *Validator) validateSQLInjection(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return &ValidationError{
			Field:    fieldName,
			Message:  "SQL injection validation requires string type",
			Value:    field.Interface(),
			Rule:     "sql_injection",
			Severity: "error",
			Code:     "invalid_type",
		}
	}

	value := field.String()
	if v.sqlInjectionPattern.MatchString(value) {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Potential SQL injection detected",
			Value:    value,
			Rule:     "sql_injection",
			Severity: "error",
			Code:     "sql_injection_detected",
		}
	}

	return nil
}

// validateXSS checks for XSS patterns
func (v *Validator) validateXSS(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return &ValidationError{
			Field:    fieldName,
			Message:  "XSS validation requires string type",
			Value:    field.Interface(),
			Rule:     "xss",
			Severity: "error",
			Code:     "invalid_type",
		}
	}

	value := field.String()
	if v.xssPattern.MatchString(value) {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Potential XSS attack detected",
			Value:    value,
			Rule:     "xss",
			Severity: "error",
			Code:     "xss_detected",
		}
	}

	return nil
}

// validatePathTraversal checks for path traversal patterns
func (v *Validator) validatePathTraversal(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Path traversal validation requires string type",
			Value:    field.Interface(),
			Rule:     "path_traversal",
			Severity: "error",
			Code:     "invalid_type",
		}
	}

	value := field.String()
	if v.pathTraversalPattern.MatchString(value) {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Potential path traversal detected",
			Value:    value,
			Rule:     "path_traversal",
			Severity: "error",
			Code:     "path_traversal_detected",
		}
	}

	return nil
}

// validateAlphanumeric checks if string contains only alphanumeric characters
func (v *Validator) validateAlphanumeric(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Alphanumeric validation requires string type",
			Value:    field.Interface(),
			Rule:     "alphanumeric",
			Severity: "error",
			Code:     "invalid_type",
		}
	}

	value := field.String()
	alphanumericPattern := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphanumericPattern.MatchString(value) {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Value must contain only alphanumeric characters",
			Value:    value,
			Rule:     "alphanumeric",
			Severity: "error",
			Code:     "invalid_alphanumeric",
		}
	}

	return nil
}

// validateNumeric checks if string contains only numeric characters
func (v *Validator) validateNumeric(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Numeric validation requires string type",
			Value:    field.Interface(),
			Rule:     "numeric",
			Severity: "error",
			Code:     "invalid_type",
		}
	}

	value := field.String()
	numericPattern := regexp.MustCompile(`^[0-9]+$`)
	if !numericPattern.MatchString(value) {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Value must contain only numeric characters",
			Value:    value,
			Rule:     "numeric",
			Severity: "error",
			Code:     "invalid_numeric",
		}
	}

	return nil
}

// validateDate validates date format
func (v *Validator) validateDate(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Date validation requires string type",
			Value:    field.Interface(),
			Rule:     "date",
			Severity: "error",
			Code:     "invalid_type",
		}
	}

	value := field.String()
	if value == "" {
		return nil // Empty date is allowed unless required
	}

	// Try common date formats
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
		"01/02/2006",
		"02/01/2006",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, value); err == nil {
			return nil
		}
	}

	return &ValidationError{
		Field:    fieldName,
		Message:  "Invalid date format",
		Value:    value,
		Rule:     "date",
		Severity: "error",
		Code:     "invalid_date",
	}
}

// validateUUID validates UUID format
func (v *Validator) validateUUID(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return &ValidationError{
			Field:    fieldName,
			Message:  "UUID validation requires string type",
			Value:    field.Interface(),
			Rule:     "uuid",
			Severity: "error",
			Code:     "invalid_type",
		}
	}

	value := field.String()
	if value == "" {
		return nil // Empty UUID is allowed unless required
	}

	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidPattern.MatchString(strings.ToLower(value)) {
		return &ValidationError{
			Field:    fieldName,
			Message:  "Invalid UUID format",
			Value:    value,
			Rule:     "uuid",
			Severity: "error",
			Code:     "invalid_uuid",
		}
	}

	return nil
}

// sanitizeField sanitizes a field value
func (v *Validator) sanitizeField(field reflect.Value, fieldName string) {
	if field.Kind() != reflect.String {
		return
	}

	if !field.CanSet() {
		return
	}

	value := field.String()

	// Trim whitespace
	value = strings.TrimSpace(value)

	// Remove null bytes
	value = strings.ReplaceAll(value, "\x00", "")

	// Normalize line endings
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")

	// Set sanitized value
	field.SetString(value)
}

// logValidationError logs validation errors if enabled
func (v *Validator) logValidationError(code, message string, context map[string]interface{}) {
	if !v.config.LogValidationErrors {
		return
	}

	fields := []zap.Field{
		zap.String("code", code),
		zap.String("message", message),
	}

	if context != nil {
		for key, value := range context {
			fields = append(fields, zap.Any(key, value))
		}
	}

	v.logger.Warn("validation error", fields...)
}

// parseNumber parses a string to float64
func parseNumber(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// ValidateRequest validates an HTTP request
func (v *Validator) ValidateRequest(r *http.Request, target interface{}) *InputValidationResult {
	start := time.Now()
	result := &InputValidationResult{
		IsValid:  true,
		Errors:   make(ValidationErrors, 0),
		Warnings: make(ValidationErrors, 0),
	}

	// Parse JSON body
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(target); err != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:    "body",
				Message:  "Invalid JSON format",
				Severity: "error",
				Code:     "invalid_json",
			})
			return result
		}
	}

	// Validate the struct
	structResult := v.ValidateStruct(target)
	result.Errors = append(result.Errors, structResult.Errors...)
	result.Warnings = append(result.Warnings, structResult.Warnings...)
	result.IsValid = result.IsValid && structResult.IsValid
	result.Sanitized = structResult.Sanitized

	result.Duration = time.Since(start)
	return result
}
