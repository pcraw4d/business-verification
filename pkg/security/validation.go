package security

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

// InputValidator validates and sanitizes user inputs
type InputValidator struct {
	// Compiled regex patterns
	emailRegex        *regexp.Regexp
	phoneRegex        *regexp.Regexp
	urlRegex          *regexp.Regexp
	sqlInjectionRegex *regexp.Regexp
	xssRegex          *regexp.Regexp
}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	return &InputValidator{
		emailRegex:        regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		phoneRegex:        regexp.MustCompile(`^\+[1-9]\d{1,14}$`),
		urlRegex:          regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`),
		sqlInjectionRegex: regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|vbscript|onload|onerror|onclick)`),
		xssRegex:          regexp.MustCompile(`(?i)(<script|</script|javascript:|vbscript:|onload=|onerror=|onclick=)`),
	}
}

// ValidationResult represents the result of input validation
type ValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// ValidateEmail validates an email address
func (iv *InputValidator) ValidateEmail(email string) ValidationResult {
	result := ValidationResult{IsValid: true}

	if email == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "email is required")
		return result
	}

	if len(email) > 254 {
		result.IsValid = false
		result.Errors = append(result.Errors, "email exceeds maximum length of 254 characters")
		return result
	}

	if !iv.emailRegex.MatchString(email) {
		result.IsValid = false
		result.Errors = append(result.Errors, "email format is invalid")
		return result
	}

	// Check for suspicious patterns
	if iv.sqlInjectionRegex.MatchString(email) {
		result.IsValid = false
		result.Errors = append(result.Errors, "email contains potentially harmful content")
		return result
	}

	return result
}

// ValidatePhone validates a phone number
func (iv *InputValidator) ValidatePhone(phone string) ValidationResult {
	result := ValidationResult{IsValid: true}

	if phone == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "phone is required")
		return result
	}

	if !iv.phoneRegex.MatchString(phone) {
		result.IsValid = false
		result.Errors = append(result.Errors, "phone format is invalid (must be in E.164 format)")
		return result
	}

	return result
}

// ValidateURL validates a URL
func (iv *InputValidator) ValidateURL(urlStr string) ValidationResult {
	result := ValidationResult{IsValid: true}

	if urlStr == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "URL is required")
		return result
	}

	// Parse URL to check format
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "URL format is invalid")
		return result
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		result.IsValid = false
		result.Errors = append(result.Errors, "URL must use http or https scheme")
		return result
	}

	// Check for suspicious patterns
	if iv.sqlInjectionRegex.MatchString(urlStr) || iv.xssRegex.MatchString(urlStr) {
		result.IsValid = false
		result.Errors = append(result.Errors, "URL contains potentially harmful content")
		return result
	}

	return result
}

// ValidateBusinessName validates a business name
func (iv *InputValidator) ValidateBusinessName(name string) ValidationResult {
	result := ValidationResult{IsValid: true}

	if name == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "business name is required")
		return result
	}

	if len(name) < 2 {
		result.IsValid = false
		result.Errors = append(result.Errors, "business name must be at least 2 characters")
		return result
	}

	if len(name) > 255 {
		result.IsValid = false
		result.Errors = append(result.Errors, "business name exceeds maximum length of 255 characters")
		return result
	}

	// Check for suspicious patterns
	if iv.sqlInjectionRegex.MatchString(name) || iv.xssRegex.MatchString(name) {
		result.IsValid = false
		result.Errors = append(result.Errors, "business name contains potentially harmful content")
		return result
	}

	// Check for valid characters (letters, numbers, spaces, common punctuation)
	validCharRegex := regexp.MustCompile(`^[a-zA-Z0-9\s\-.,&'()]+$`)
	if !validCharRegex.MatchString(name) {
		result.Warnings = append(result.Warnings, "business name contains unusual characters")
	}

	return result
}

// ValidateDescription validates a business description
func (iv *InputValidator) ValidateDescription(description string) ValidationResult {
	result := ValidationResult{IsValid: true}

	if description == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "description is required")
		return result
	}

	if len(description) < 10 {
		result.IsValid = false
		result.Errors = append(result.Errors, "description must be at least 10 characters")
		return result
	}

	if len(description) > 2000 {
		result.IsValid = false
		result.Errors = append(result.Errors, "description exceeds maximum length of 2000 characters")
		return result
	}

	// Check for suspicious patterns
	if iv.sqlInjectionRegex.MatchString(description) || iv.xssRegex.MatchString(description) {
		result.IsValid = false
		result.Errors = append(result.Errors, "description contains potentially harmful content")
		return result
	}

	return result
}

// SanitizeString sanitizes a string by removing potentially harmful content
func (iv *InputValidator) SanitizeString(input string) string {
	// Remove SQL injection patterns
	sanitized := iv.sqlInjectionRegex.ReplaceAllString(input, "")

	// Remove XSS patterns
	sanitized = iv.xssRegex.ReplaceAllString(sanitized, "")

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// ValidateJSON validates JSON input
func (iv *InputValidator) ValidateJSON(jsonStr string) ValidationResult {
	result := ValidationResult{IsValid: true}

	if jsonStr == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "JSON is required")
		return result
	}

	// Try to parse JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("invalid JSON format: %v", err))
		return result
	}

	// Check for suspicious patterns in JSON
	if iv.sqlInjectionRegex.MatchString(jsonStr) || iv.xssRegex.MatchString(jsonStr) {
		result.IsValid = false
		result.Errors = append(result.Errors, "JSON contains potentially harmful content")
		return result
	}

	return result
}

// ValidatePassword validates a password
func (iv *InputValidator) ValidatePassword(password string) ValidationResult {
	result := ValidationResult{IsValid: true}

	if password == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "password is required")
		return result
	}

	if len(password) < 8 {
		result.IsValid = false
		result.Errors = append(result.Errors, "password must be at least 8 characters")
		return result
	}

	if len(password) > 128 {
		result.IsValid = false
		result.Errors = append(result.Errors, "password exceeds maximum length of 128 characters")
		return result
	}

	// Check for at least one uppercase letter
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		result.Errors = append(result.Errors, "password must contain at least one uppercase letter")
		result.IsValid = false
	}

	if !hasLower {
		result.Errors = append(result.Errors, "password must contain at least one lowercase letter")
		result.IsValid = false
	}

	if !hasDigit {
		result.Errors = append(result.Errors, "password must contain at least one digit")
		result.IsValid = false
	}

	if !hasSpecial {
		result.Errors = append(result.Errors, "password must contain at least one special character")
		result.IsValid = false
	}

	return result
}
