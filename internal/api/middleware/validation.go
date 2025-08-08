package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ValidationConfig holds validation configuration
type ValidationConfig struct {
	MaxBodySize   int64    // Maximum request body size in bytes
	RequiredPaths []string // Paths that require validation
	Enabled       bool
}

// Validator provides request validation middleware
type Validator struct {
	config *ValidationConfig
	logger *observability.Logger
}

// NewValidator creates a new validation middleware
func NewValidator(config *ValidationConfig, logger *observability.Logger) *Validator {
	return &Validator{
		config: config,
		logger: logger,
	}
}

// Middleware returns the validation middleware
func (v *Validator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !v.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Check if this path requires validation
		if !v.shouldValidate(r.URL.Path, r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		// Validate content type for POST/PUT requests
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			if err := v.validateContentType(r); err != nil {
				v.logger.WithComponent("validator").Warn("Invalid content type",
					"error", err,
					"path", r.URL.Path,
					"method", r.Method,
					"content_type", r.Header.Get("Content-Type"))

				http.Error(w, fmt.Sprintf("Invalid content type: %v", err), http.StatusBadRequest)
				return
			}

			// Validate request body size
			if err := v.validateBodySize(r); err != nil {
				v.logger.WithComponent("validator").Warn("Request body too large",
					"error", err,
					"path", r.URL.Path,
					"method", r.Method)

				http.Error(w, fmt.Sprintf("Request body too large: %v", err), http.StatusRequestEntityTooLarge)
				return
			}

			// Validate JSON structure
			if err := v.validateJSON(r); err != nil {
				v.logger.WithComponent("validator").Warn("Invalid JSON",
					"error", err,
					"path", r.URL.Path,
					"method", r.Method)

				http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
				return
			}
		}

		// Validate query parameters
		if err := v.validateQueryParams(r); err != nil {
			v.logger.WithComponent("validator").Warn("Invalid query parameters",
				"error", err,
				"path", r.URL.Path,
				"method", r.Method)

			http.Error(w, fmt.Sprintf("Invalid query parameters: %v", err), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// shouldValidate determines if a request should be validated
func (v *Validator) shouldValidate(path, method string) bool {
	// Always validate API endpoints
	if strings.HasPrefix(path, "/v1/") {
		return true
	}

	// Check specific paths
	for _, reqPath := range v.config.RequiredPaths {
		if strings.HasPrefix(path, reqPath) {
			return true
		}
	}

	return false
}

// validateContentType checks if the content type is valid for the request
func (v *Validator) validateContentType(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return fmt.Errorf("missing Content-Type header")
	}

	// Extract the media type (ignore charset and other parameters)
	mediaType := strings.Split(contentType, ";")[0]
	mediaType = strings.TrimSpace(mediaType)

	switch mediaType {
	case "application/json":
		return nil
	case "application/x-www-form-urlencoded":
		return nil
	case "multipart/form-data":
		return nil
	default:
		return fmt.Errorf("unsupported content type: %s", mediaType)
	}
}

// validateBodySize checks if the request body size is within limits
func (v *Validator) validateBodySize(r *http.Request) error {
	if r.ContentLength > v.config.MaxBodySize {
		return fmt.Errorf("request body size %d exceeds maximum %d bytes",
			r.ContentLength, v.config.MaxBodySize)
	}
	return nil
}

// validateJSON validates that the request body contains valid JSON
func (v *Validator) validateJSON(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return nil // Skip JSON validation for non-JSON requests
	}

	// Read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// Restore the body for subsequent handlers
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// Skip validation for empty bodies
	if len(body) == 0 {
		return nil
	}

	// Validate JSON structure
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	// Basic structure validation
	if err := v.validateJSONStructure(jsonData, r.URL.Path); err != nil {
		return err
	}

	return nil
}

// validateJSONStructure performs basic validation on JSON structure
func (v *Validator) validateJSONStructure(data interface{}, path string) error {
	switch value := data.(type) {
	case map[string]interface{}:
		return v.validateJSONObject(value, path)
	case []interface{}:
		return v.validateJSONArray(value, path)
	default:
		// Primitive types are generally acceptable
		return nil
	}
}

// validateJSONObject validates JSON objects
func (v *Validator) validateJSONObject(obj map[string]interface{}, path string) error {
	// Check for excessively nested objects
	if err := v.checkNestingDepth(obj, 0, 10); err != nil {
		return err
	}

	// Path-specific validations
	switch {
	case strings.Contains(path, "/classify"):
		return v.validateClassificationRequest(obj)
	case strings.Contains(path, "/auth/register"):
		return v.validateRegistrationRequest(obj)
	case strings.Contains(path, "/auth/login"):
		return v.validateLoginRequest(obj)
	default:
		return nil
	}
}

// validateJSONArray validates JSON arrays
func (v *Validator) validateJSONArray(arr []interface{}, path string) error {
	// Limit array size to prevent DoS attacks
	if len(arr) > 1000 {
		return fmt.Errorf("array too large: %d items (max 1000)", len(arr))
	}

	// Validate each item in the array
	for i, item := range arr {
		if err := v.validateJSONStructure(item, path); err != nil {
			return fmt.Errorf("invalid item at index %d: %w", i, err)
		}
	}

	return nil
}

// checkNestingDepth prevents deeply nested objects that could cause stack overflow
func (v *Validator) checkNestingDepth(obj map[string]interface{}, currentDepth, maxDepth int) error {
	if currentDepth > maxDepth {
		return fmt.Errorf("object nesting too deep: %d levels (max %d)", currentDepth, maxDepth)
	}

	for _, value := range obj {
		if nestedObj, ok := value.(map[string]interface{}); ok {
			if err := v.checkNestingDepth(nestedObj, currentDepth+1, maxDepth); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateClassificationRequest validates business classification requests
func (v *Validator) validateClassificationRequest(obj map[string]interface{}) error {
	// Check required fields
	businessName, hasName := obj["business_name"]
	if !hasName || businessName == nil {
		return fmt.Errorf("missing required field: business_name")
	}

	// Validate business_name is a non-empty string
	if name, ok := businessName.(string); !ok || strings.TrimSpace(name) == "" {
		return fmt.Errorf("business_name must be a non-empty string")
	}

	// Validate optional fields if present
	if businessType, exists := obj["business_type"]; exists && businessType != nil {
		if _, ok := businessType.(string); !ok {
			return fmt.Errorf("business_type must be a string")
		}
	}

	if industry, exists := obj["industry"]; exists && industry != nil {
		if _, ok := industry.(string); !ok {
			return fmt.Errorf("industry must be a string")
		}
	}

	return nil
}

// validateRegistrationRequest validates user registration requests
func (v *Validator) validateRegistrationRequest(obj map[string]interface{}) error {
	requiredFields := []string{"username", "email", "password"}

	for _, field := range requiredFields {
		value, exists := obj[field]
		if !exists || value == nil {
			return fmt.Errorf("missing required field: %s", field)
		}

		if str, ok := value.(string); !ok || strings.TrimSpace(str) == "" {
			return fmt.Errorf("%s must be a non-empty string", field)
		}
	}

	// Additional email validation
	if email, ok := obj["email"].(string); ok {
		if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
			return fmt.Errorf("invalid email format")
		}
	}

	return nil
}

// validateLoginRequest validates user login requests
func (v *Validator) validateLoginRequest(obj map[string]interface{}) error {
	requiredFields := []string{"email", "password"}

	for _, field := range requiredFields {
		value, exists := obj[field]
		if !exists || value == nil {
			return fmt.Errorf("missing required field: %s", field)
		}

		if str, ok := value.(string); !ok || strings.TrimSpace(str) == "" {
			return fmt.Errorf("%s must be a non-empty string", field)
		}
	}

	return nil
}

// validateQueryParams validates query parameters
func (v *Validator) validateQueryParams(r *http.Request) error {
	queryParams := r.URL.Query()

	// Check for suspicious patterns
	for key, values := range queryParams {
		for _, value := range values {
			// Check for excessively long parameters
			if len(value) > 1000 {
				return fmt.Errorf("query parameter '%s' too long: %d characters (max 1000)", key, len(value))
			}

			// Check for suspicious characters that might indicate injection attempts
			if strings.ContainsAny(value, "<>\"'&") {
				v.logger.WithComponent("validator").Info("Suspicious query parameter detected",
					"key", key,
					"value", value,
					"path", r.URL.Path)
			}
		}
	}

	return nil
}

// ValidateStruct validates a struct using reflection (utility function)
func (v *Validator) ValidateStruct(s interface{}) error {
	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)

	// Handle pointers
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", val.Kind())
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Check for required tag
		if tag := fieldType.Tag.Get("validate"); tag == "required" {
			if v.isZeroValue(field) {
				return fmt.Errorf("required field '%s' is missing or empty", fieldType.Name)
			}
		}
	}

	return nil
}

// isZeroValue checks if a reflect.Value is the zero value for its type
func (v *Validator) isZeroValue(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.String:
		return val.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return val.IsNil()
	default:
		return false
	}
}
