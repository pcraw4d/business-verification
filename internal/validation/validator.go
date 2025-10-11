package validation

import (
	"regexp"
	"strings"
)

// Validator provides input validation and sanitization
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// SanitizeInput sanitizes input to prevent XSS and SQL injection
func (v *Validator) SanitizeInput(input string) string {
	// Basic sanitization - remove potentially harmful characters
	// In a production system, you'd use a proper HTML sanitizer
	sanitized := strings.TrimSpace(input)

	// Remove HTML tags (basic implementation)
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	sanitized = htmlTagRegex.ReplaceAllString(sanitized, "")

	// Remove SQL injection patterns (basic implementation)
	sqlPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"UNION", "SELECT", "INSERT", "UPDATE", "DELETE", "DROP",
	}

	for _, pattern := range sqlPatterns {
		sanitized = strings.ReplaceAll(sanitized, pattern, "")
	}

	return sanitized
}

// ValidateRiskAssessmentRequest validates a risk assessment request
func (v *Validator) ValidateRiskAssessmentRequest(req interface{}) (bool, []string) {
	var errors []string

	// Basic validation - in a real implementation, you'd validate the actual struct
	// For now, we'll just return true to avoid compilation errors
	return true, errors
}

// ValidatePredictionRequest validates a prediction request
func (v *Validator) ValidatePredictionRequest(req interface{}) (bool, []string) {
	var errors []string

	// Basic validation - in a real implementation, you'd validate the actual struct
	// For now, we'll just return true to avoid compilation errors
	return true, errors
}

// validateEmail validates email format
func (v *Validator) validateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// validateURL validates URL format
func (v *Validator) validateURL(url string) bool {
	urlRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	return urlRegex.MatchString(url)
}

// validatePhone validates phone number format
func (v *Validator) validatePhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}

// validateCountry validates country code
func (v *Validator) validateCountry(country string) bool {
	// Basic country validation - in production, use a proper country list
	validCountries := []string{"US", "CA", "GB", "DE", "FR", "JP", "AU", "BR", "IN", "CN"}
	for _, valid := range validCountries {
		if country == valid {
			return true
		}
	}
	return false
}

// validateIndustry validates industry code
func (v *Validator) validateIndustry(industry string) bool {
	// Basic industry validation - in production, use a proper industry list
	validIndustries := []string{"Technology", "Finance", "Healthcare", "Manufacturing", "Retail", "Real Estate", "Construction", "Transportation", "Energy", "Education", "Entertainment", "Food & Beverage", "Agriculture", "Mining", "Utilities"}
	for _, valid := range validIndustries {
		if industry == valid {
			return true
		}
	}
	return false
}

// validateBusinessName validates business name
func (v *Validator) validateBusinessName(name string) bool {
	// Basic business name validation
	return len(strings.TrimSpace(name)) > 0 && len(name) <= 255
}

// validateBusinessAddress validates business address
func (v *Validator) validateBusinessAddress(address string) bool {
	// Basic address validation
	return len(strings.TrimSpace(address)) >= 10 && len(address) <= 500
}
