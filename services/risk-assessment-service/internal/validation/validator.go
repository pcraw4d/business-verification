package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// Validator provides comprehensive input validation
type Validator struct {
	// Email validation regex
	emailRegex *regexp.Regexp

	// Phone validation regex (E.164 format)
	phoneRegex *regexp.Regexp

	// URL validation regex
	urlRegex *regexp.Regexp

	// SQL injection patterns
	sqlInjectionPatterns []*regexp.Regexp

	// XSS patterns
	xssPatterns []*regexp.Regexp
}

// NewValidator creates a new validator with compiled regex patterns
func NewValidator() *Validator {
	return &Validator{
		emailRegex: regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		phoneRegex: regexp.MustCompile(`^\+[1-9]\d{1,14}$`),
		urlRegex:   regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`),
		sqlInjectionPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`),
			regexp.MustCompile(`(?i)(script|javascript|vbscript|onload|onerror|onclick)`),
			regexp.MustCompile(`['\";]`),
			regexp.MustCompile(`(?i)(or|and)\s+\d+\s*=\s*\d+`),
			regexp.MustCompile(`(?i)(or|and)\s+'.*'\s*=\s*'.*'`),
		},
		xssPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
			regexp.MustCompile(`(?i)javascript:`),
			regexp.MustCompile(`(?i)on\w+\s*=`),
			regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`),
			regexp.MustCompile(`(?i)<object[^>]*>.*?</object>`),
		},
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", ve.Field, ve.Message)
}

// ValidationResult contains the result of validation
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}

// ValidateRiskAssessmentRequest validates a risk assessment request
func (v *Validator) ValidateRiskAssessmentRequest(req *models.RiskAssessmentRequest) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationError, 0),
	}

	// Validate business name
	if err := v.validateBusinessName(req.BusinessName); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "business_name",
			Message: err.Error(),
			Code:    "INVALID_BUSINESS_NAME",
		})
		result.Valid = false
	}

	// Validate business address
	if err := v.validateBusinessAddress(req.BusinessAddress); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "business_address",
			Message: err.Error(),
			Code:    "INVALID_BUSINESS_ADDRESS",
		})
		result.Valid = false
	}

	// Validate industry
	if err := v.validateIndustry(req.Industry); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "industry",
			Message: err.Error(),
			Code:    "INVALID_INDUSTRY",
		})
		result.Valid = false
	}

	// Validate country
	if err := v.validateCountry(req.Country); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "country",
			Message: err.Error(),
			Code:    "INVALID_COUNTRY",
		})
		result.Valid = false
	}

	// Validate optional fields
	if req.Phone != "" {
		if err := v.validatePhone(req.Phone); err != nil {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   "phone",
				Message: err.Error(),
				Code:    "INVALID_PHONE",
			})
		}
	}

	if req.Email != "" {
		if err := v.validateEmail(req.Email); err != nil {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   "email",
				Message: err.Error(),
				Code:    "INVALID_EMAIL",
			})
		}
	}

	if req.Website != "" {
		if err := v.validateURL(req.Website); err != nil {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   "website",
				Message: err.Error(),
				Code:    "INVALID_URL",
			})
		}
	}

	// Validate prediction horizon
	if req.PredictionHorizon < 0 || req.PredictionHorizon > 12 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "prediction_horizon",
			Message: "prediction horizon must be between 0 and 12 months",
			Code:    "INVALID_PREDICTION_HORIZON",
		})
		result.Valid = false
	}

	// Validate metadata
	if req.Metadata != nil {
		if warnings := v.validateMetadata(req.Metadata); len(warnings) > 0 {
			result.Warnings = append(result.Warnings, warnings...)
		}
	}

	return result
}

// validateBusinessName validates business name
func (v *Validator) validateBusinessName(name string) error {
	if name == "" {
		return fmt.Errorf("business name is required")
	}

	if len(name) < 1 {
		return fmt.Errorf("business name must be at least 1 character")
	}

	if len(name) > 255 {
		return fmt.Errorf("business name exceeds maximum length of 255 characters")
	}

	// Check for malicious content
	if v.containsMaliciousContent(name) {
		return fmt.Errorf("business name contains potentially harmful content")
	}

	// Check for valid characters (alphanumeric, spaces, common punctuation)
	if !v.isValidBusinessName(name) {
		return fmt.Errorf("business name contains invalid characters")
	}

	return nil
}

// validateBusinessAddress validates business address
func (v *Validator) validateBusinessAddress(address string) error {
	if address == "" {
		return fmt.Errorf("business address is required")
	}

	if len(address) < 10 {
		return fmt.Errorf("business address must be at least 10 characters")
	}

	if len(address) > 500 {
		return fmt.Errorf("business address exceeds maximum length of 500 characters")
	}

	// Check for malicious content
	if v.containsMaliciousContent(address) {
		return fmt.Errorf("business address contains potentially harmful content")
	}

	// Check for valid address format
	if !v.isValidAddress(address) {
		return fmt.Errorf("business address format appears invalid")
	}

	return nil
}

// validateIndustry validates industry
func (v *Validator) validateIndustry(industry string) error {
	if industry == "" {
		return fmt.Errorf("industry is required")
	}

	if len(industry) < 1 {
		return fmt.Errorf("industry must be at least 1 character")
	}

	if len(industry) > 100 {
		return fmt.Errorf("industry exceeds maximum length of 100 characters")
	}

	// Check for malicious content
	if v.containsMaliciousContent(industry) {
		return fmt.Errorf("industry contains potentially harmful content")
	}

	// Check for valid industry format
	if !v.isValidIndustry(industry) {
		return fmt.Errorf("industry contains invalid characters")
	}

	return nil
}

// validateCountry validates country code
func (v *Validator) validateCountry(country string) error {
	if country == "" {
		return fmt.Errorf("country is required")
	}

	if len(country) != 2 {
		return fmt.Errorf("country must be a 2-letter ISO code")
	}

	// Check if it's a valid ISO 3166-1 alpha-2 code
	if !v.isValidCountryCode(country) {
		return fmt.Errorf("invalid country code")
	}

	return nil
}

// validatePhone validates phone number
func (v *Validator) validatePhone(phone string) error {
	if phone == "" {
		return nil // Optional field
	}

	if len(phone) < 10 || len(phone) > 20 {
		return fmt.Errorf("phone number must be between 10 and 20 characters")
	}

	// Check E.164 format
	if !v.phoneRegex.MatchString(phone) {
		return fmt.Errorf("phone number must be in E.164 format (e.g., +1234567890)")
	}

	return nil
}

// validateEmail validates email address
func (v *Validator) validateEmail(email string) error {
	if email == "" {
		return nil // Optional field
	}

	if len(email) > 254 {
		return fmt.Errorf("email address exceeds maximum length of 254 characters")
	}

	// Check email format
	if !v.emailRegex.MatchString(email) {
		return fmt.Errorf("email address format is invalid")
	}

	// Check for malicious content
	if v.containsMaliciousContent(email) {
		return fmt.Errorf("email address contains potentially harmful content")
	}

	return nil
}

// validateURL validates URL
func (v *Validator) validateURL(urlStr string) error {
	if urlStr == "" {
		return nil // Optional field
	}

	if len(urlStr) > 2048 {
		return fmt.Errorf("URL exceeds maximum length of 2048 characters")
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format")
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	// Check for malicious content
	if v.containsMaliciousContent(urlStr) {
		return fmt.Errorf("URL contains potentially harmful content")
	}

	return nil
}

// validateMetadata validates metadata
func (v *Validator) validateMetadata(metadata map[string]interface{}) []ValidationError {
	warnings := make([]ValidationError, 0)

	// Check metadata size
	if len(metadata) > 50 {
		warnings = append(warnings, ValidationError{
			Field:   "metadata",
			Message: "metadata contains too many fields (max 50)",
			Code:    "METADATA_TOO_LARGE",
		})
	}

	// Validate specific metadata fields
	for key, value := range metadata {
		if len(key) > 100 {
			warnings = append(warnings, ValidationError{
				Field:   fmt.Sprintf("metadata.%s", key),
				Message: "metadata key exceeds maximum length of 100 characters",
				Code:    "INVALID_METADATA_KEY",
			})
		}

		// Check value type and size
		switch v := value.(type) {
		case string:
			if len(v) > 1000 {
				warnings = append(warnings, ValidationError{
					Field:   fmt.Sprintf("metadata.%s", key),
					Message: "metadata string value exceeds maximum length of 1000 characters",
					Code:    "INVALID_METADATA_VALUE",
				})
			}
		case float64:
			if v < -1e10 || v > 1e10 {
				warnings = append(warnings, ValidationError{
					Field:   fmt.Sprintf("metadata.%s", key),
					Message: "metadata numeric value is out of range",
					Code:    "INVALID_METADATA_VALUE",
				})
			}
		case bool:
			// Boolean values are always valid
		default:
			warnings = append(warnings, ValidationError{
				Field:   fmt.Sprintf("metadata.%s", key),
				Message: "metadata value must be string, number, or boolean",
				Code:    "INVALID_METADATA_VALUE",
			})
		}
	}

	return warnings
}

// containsMaliciousContent checks for SQL injection and XSS patterns
func (v *Validator) containsMaliciousContent(input string) bool {
	// Check SQL injection patterns
	for _, pattern := range v.sqlInjectionPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}

	// Check XSS patterns
	for _, pattern := range v.xssPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}

	return false
}

// isValidBusinessName checks if business name contains valid characters
func (v *Validator) isValidBusinessName(name string) bool {
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && !unicode.IsSpace(char) &&
			char != '.' && char != ',' && char != '-' && char != '&' && char != '(' && char != ')' {
			return false
		}
	}
	return true
}

// isValidAddress checks if address appears to be valid
func (v *Validator) isValidAddress(address string) bool {
	// Basic address validation - should contain numbers and letters
	hasNumber := false
	hasLetter := false

	for _, char := range address {
		if unicode.IsDigit(char) {
			hasNumber = true
		}
		if unicode.IsLetter(char) {
			hasLetter = true
		}
	}

	return hasNumber && hasLetter
}

// isValidIndustry checks if industry contains valid characters
func (v *Validator) isValidIndustry(industry string) bool {
	for _, char := range industry {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && !unicode.IsSpace(char) &&
			char != '-' && char != '&' {
			return false
		}
	}
	return true
}

// isValidCountryCode checks if country code is valid
func (v *Validator) isValidCountryCode(country string) bool {
	// List of valid ISO 3166-1 alpha-2 country codes
	validCountries := map[string]bool{
		"US": true, "CA": true, "GB": true, "DE": true, "FR": true, "IT": true,
		"ES": true, "NL": true, "AU": true, "JP": true, "CN": true, "IN": true,
		"BR": true, "MX": true, "AR": true, "CL": true, "CO": true, "PE": true,
		"ZA": true, "NG": true, "EG": true, "KE": true, "MA": true, "TN": true,
		"RU": true, "TR": true, "SA": true, "AE": true, "IL": true, "TH": true,
		"SG": true, "MY": true, "ID": true, "PH": true, "VN": true, "KR": true,
		"TW": true, "HK": true, "NZ": true, "NO": true, "SE": true, "DK": true,
		"FI": true, "CH": true, "AT": true, "BE": true, "IE": true, "PT": true,
		"GR": true, "PL": true, "CZ": true, "HU": true, "RO": true, "BG": true,
		"HR": true, "SI": true, "SK": true, "LT": true, "LV": true, "EE": true,
	}

	return validCountries[strings.ToUpper(country)]
}

// SanitizeInput sanitizes input to prevent injection attacks
func (v *Validator) SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove control characters
	var result strings.Builder
	for _, char := range input {
		if char >= 32 || char == '\t' || char == '\n' || char == '\r' {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// ValidatePredictionRequest validates a prediction request
func (v *Validator) ValidatePredictionRequest(horizonMonths int, scenarios []string) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationError, 0),
	}

	// Validate horizon
	if horizonMonths < 1 || horizonMonths > 12 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "horizon_months",
			Message: "horizon must be between 1 and 12 months",
			Code:    "INVALID_HORIZON",
		})
		result.Valid = false
	}

	// Validate scenarios
	if len(scenarios) > 10 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "scenarios",
			Message: "too many scenarios specified (max 10)",
			Code:    "TOO_MANY_SCENARIOS",
		})
	}

	for i, scenario := range scenarios {
		if len(scenario) > 50 {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   fmt.Sprintf("scenarios[%d]", i),
				Message: "scenario name exceeds maximum length of 50 characters",
				Code:    "INVALID_SCENARIO_NAME",
			})
		}

		if v.containsMaliciousContent(scenario) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("scenarios[%d]", i),
				Message: "scenario name contains potentially harmful content",
				Code:    "MALICIOUS_SCENARIO",
			})
			result.Valid = false
		}
	}

	return result
}
