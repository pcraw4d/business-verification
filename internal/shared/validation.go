package shared

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// =============================================================================
// Validation Rules and Schemas
// =============================================================================

// ValidationRule represents a validation rule
type ValidationRule struct {
	Field       string                 `json:"field"`
	Rule        string                 `json:"rule"`
	Message     string                 `json:"message"`
	Required    bool                   `json:"required"`
	MinLength   int                    `json:"min_length,omitempty"`
	MaxLength   int                    `json:"max_length,omitempty"`
	Pattern     string                 `json:"pattern,omitempty"`
	MinValue    float64                `json:"min_value,omitempty"`
	MaxValue    float64                `json:"max_value,omitempty"`
	AllowedValues []string             `json:"allowed_values,omitempty"`
	Custom      func(interface{}) error `json:"-"`
}

// ValidationSchema represents a validation schema
type ValidationSchema struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Rules       []ValidationRule `json:"rules"`
}

// =============================================================================
// Predefined Validation Schemas
// =============================================================================

// BusinessClassificationRequestSchema defines validation rules for business classification requests
var BusinessClassificationRequestSchema = ValidationSchema{
	Name:        "BusinessClassificationRequest",
	Description: "Validation schema for business classification requests",
	Rules: []ValidationRule{
		{
			Field:    "business_name",
			Rule:     "required",
			Message:  "Business name is required",
			Required: true,
		},
		{
			Field:     "business_name",
			Rule:      "length",
			Message:   "Business name must be between 1 and 200 characters",
			MinLength: 1,
			MaxLength: 200,
		},
		{
			Field:     "business_name",
			Rule:      "pattern",
			Message:   "Business name contains invalid characters",
			Pattern:   `^[a-zA-Z0-9\s\-\.&'()]+$`,
		},
		{
			Field:     "business_type",
			Rule:      "length",
			Message:   "Business type must be between 0 and 50 characters",
			MaxLength: 50,
		},
		{
			Field:     "industry",
			Rule:      "length",
			Message:   "Industry must be between 0 and 100 characters",
			MaxLength: 100,
		},
		{
			Field:     "description",
			Rule:      "length",
			Message:   "Description must be between 0 and 1000 characters",
			MaxLength: 1000,
		},
		{
			Field:     "website_url",
			Rule:      "url",
			Message:   "Website URL must be a valid URL",
			Pattern:   `^https?://[^\s/$.?#].[^\s]*$`,
		},
		{
			Field:     "registration_number",
			Rule:      "length",
			Message:   "Registration number must be between 0 and 50 characters",
			MaxLength: 50,
		},
		{
			Field:     "tax_id",
			Rule:      "length",
			Message:   "Tax ID must be between 0 and 50 characters",
			MaxLength: 50,
		},
		{
			Field:     "address",
			Rule:      "length",
			Message:   "Address must be between 0 and 500 characters",
			MaxLength: 500,
		},
		{
			Field:     "geographic_region",
			Rule:      "length",
			Message:   "Geographic region must be between 0 and 100 characters",
			MaxLength: 100,
		},
		{
			Field:     "keywords",
			Rule:      "array_length",
			Message:   "Keywords array must have at most 50 items",
			MaxLength: 50,
		},
		{
			Field:     "keywords",
			Rule:      "keyword_length",
			Message:   "Each keyword must be between 1 and 50 characters",
			MinLength: 1,
			MaxLength: 50,
		},
	},
}

// MLClassificationRequestSchema defines validation rules for ML classification requests
var MLClassificationRequestSchema = ValidationSchema{
	Name:        "MLClassificationRequest",
	Description: "Validation schema for ML classification requests",
	Rules: []ValidationRule{
		{
			Field:    "business_name",
			Rule:     "required",
			Message:  "Business name is required",
			Required: true,
		},
		{
			Field:     "business_name",
			Rule:      "length",
			Message:   "Business name must be between 1 and 200 characters",
			MinLength: 1,
			MaxLength: 200,
		},
		{
			Field:     "business_description",
			Rule:      "length",
			Message:   "Business description must be between 0 and 2000 characters",
			MaxLength: 2000,
		},
		{
			Field:     "website_content",
			Rule:      "length",
			Message:   "Website content must be between 0 and 10000 characters",
			MaxLength: 10000,
		},
		{
			Field:     "industry_hints",
			Rule:      "array_length",
			Message:   "Industry hints array must have at most 20 items",
			MaxLength: 20,
		},
		{
			Field:     "geographic_region",
			Rule:      "length",
			Message:   "Geographic region must be between 0 and 100 characters",
			MaxLength: 100,
		},
		{
			Field:     "business_type",
			Rule:      "length",
			Message:   "Business type must be between 0 and 50 characters",
			MaxLength: 50,
		},
	},
}

// WebsiteAnalysisRequestSchema defines validation rules for website analysis requests
var WebsiteAnalysisRequestSchema = ValidationSchema{
	Name:        "WebsiteAnalysisRequest",
	Description: "Validation schema for website analysis requests",
	Rules: []ValidationRule{
		{
			Field:    "business_name",
			Rule:     "required",
			Message:  "Business name is required",
			Required: true,
		},
		{
			Field:     "business_name",
			Rule:      "length",
			Message:   "Business name must be between 1 and 200 characters",
			MinLength: 1,
			MaxLength: 200,
		},
		{
			Field:    "website_url",
			Rule:     "required",
			Message:  "Website URL is required",
			Required: true,
		},
		{
			Field:     "website_url",
			Rule:      "url",
			Message:   "Website URL must be a valid URL",
			Pattern:   `^https?://[^\s/$.?#].[^\s]*$`,
		},
		{
			Field:     "max_pages",
			Rule:      "range",
			Message:   "Max pages must be between 1 and 20",
			MinValue:  1,
			MaxValue:  20,
		},
	},
}

// WebSearchAnalysisRequestSchema defines validation rules for web search analysis requests
var WebSearchAnalysisRequestSchema = ValidationSchema{
	Name:        "WebSearchAnalysisRequest",
	Description: "Validation schema for web search analysis requests",
	Rules: []ValidationRule{
		{
			Field:    "business_name",
			Rule:     "required",
			Message:  "Business name is required",
			Required: true,
		},
		{
			Field:     "business_name",
			Rule:      "length",
			Message:   "Business name must be between 1 and 200 characters",
			MinLength: 1,
			MaxLength: 200,
		},
		{
			Field:     "search_query",
			Rule:      "length",
			Message:   "Search query must be between 0 and 500 characters",
			MaxLength: 500,
		},
		{
			Field:     "business_type",
			Rule:      "length",
			Message:   "Business type must be between 0 and 50 characters",
			MaxLength: 50,
		},
		{
			Field:     "industry",
			Rule:      "length",
			Message:   "Industry must be between 0 and 100 characters",
			MaxLength: 100,
		},
		{
			Field:     "address",
			Rule:      "length",
			Message:   "Address must be between 0 and 500 characters",
			MaxLength: 500,
		},
		{
			Field:     "max_results",
			Rule:      "range",
			Message:   "Max results must be between 1 and 50",
			MinValue:  1,
			MaxValue:  50,
		},
		{
			Field:     "search_engines",
			Rule:      "array_length",
			Message:   "Search engines array must have at most 10 items",
			MaxLength: 10,
		},
		{
			Field:        "search_engines",
			Rule:         "allowed_values",
			Message:      "Search engine must be one of: google, bing, duckduckgo",
			AllowedValues: []string{"google", "bing", "duckduckgo"},
		},
	},
}

// =============================================================================
// Validation Functions
// =============================================================================

// ValidateBusinessClassificationRequest validates a business classification request
func ValidateBusinessClassificationRequest(req *BusinessClassificationRequest) (*ValidationResult, error) {
	return ValidateWithSchema(req, BusinessClassificationRequestSchema)
}

// ValidateMLClassificationRequest validates an ML classification request
func ValidateMLClassificationRequest(req *MLClassificationRequest) (*ValidationResult, error) {
	return ValidateWithSchema(req, MLClassificationRequestSchema)
}

// ValidateWebsiteAnalysisRequest validates a website analysis request
func ValidateWebsiteAnalysisRequest(req *WebsiteAnalysisRequest) (*ValidationResult, error) {
	return ValidateWithSchema(req, WebsiteAnalysisRequestSchema)
}

// ValidateWebSearchAnalysisRequest validates a web search analysis request
func ValidateWebSearchAnalysisRequest(req *WebSearchAnalysisRequest) (*ValidationResult, error) {
	return ValidateWithSchema(req, WebSearchAnalysisRequestSchema)
}

// ValidateWithSchema validates an object using a validation schema
func ValidateWithSchema(obj interface{}, schema ValidationSchema) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		Score:    1.0,
	}

	// Use reflection to get field values and apply validation rules
	// This is a simplified implementation - in a real system, you might use a library like go-playground/validator

	for _, rule := range schema.Rules {
		if err := validateRule(obj, rule); err != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   rule.Field,
				Message: rule.Message,
				Code:    rule.Rule,
			})
			result.Score -= 0.1 // Reduce score for each error
		}
	}

	// Ensure score doesn't go below 0
	if result.Score < 0 {
		result.Score = 0
	}

	return result, nil
}

// validateRule validates a single rule against an object
func validateRule(obj interface{}, rule ValidationRule) error {
	// This is a simplified implementation
	// In a real system, you would use reflection to get field values
	
	switch rule.Rule {
	case "required":
		// Check if required field is present and not empty
		return validateRequired(obj, rule)
	case "length":
		// Check string length
		return validateLength(obj, rule)
	case "array_length":
		// Check array length
		return validateArrayLength(obj, rule)
	case "pattern":
		// Check regex pattern
		return validatePattern(obj, rule)
	case "range":
		// Check numeric range
		return validateRange(obj, rule)
	case "url":
		// Check URL format
		return validateURL(obj, rule)
	case "allowed_values":
		// Check if value is in allowed list
		return validateAllowedValues(obj, rule)
	case "keyword_length":
		// Check keyword array item lengths
		return validateKeywordLength(obj, rule)
	default:
		if rule.Custom != nil {
			return rule.Custom(obj)
		}
		return fmt.Errorf("unknown validation rule: %s", rule.Rule)
	}
}

// validateRequired checks if a required field is present and not empty
func validateRequired(obj interface{}, rule ValidationRule) error {
	// Simplified implementation - in reality, you'd use reflection
	// For now, we'll assume the field is present if the object is not nil
	if obj == nil {
		return fmt.Errorf("field %s is required", rule.Field)
	}
	return nil
}

// validateLength checks if a string field meets length requirements
func validateLength(obj interface{}, rule ValidationRule) error {
	// Simplified implementation
	// In reality, you'd use reflection to get the field value
	return nil
}

// validateArrayLength checks if an array field meets length requirements
func validateArrayLength(obj interface{}, rule ValidationRule) error {
	// Simplified implementation
	return nil
}

// validatePattern checks if a string field matches a regex pattern
func validatePattern(obj interface{}, rule ValidationRule) error {
	if rule.Pattern == "" {
		return nil
	}
	
	// Simplified implementation
	// In reality, you'd use reflection to get the field value and apply the regex
	return nil
}

// validateRange checks if a numeric field is within the specified range
func validateRange(obj interface{}, rule ValidationRule) error {
	// Simplified implementation
	return nil
}

// validateURL checks if a string field is a valid URL
func validateURL(obj interface{}, rule ValidationRule) error {
	// Simplified implementation
	return nil
}

// validateAllowedValues checks if a field value is in the allowed list
func validateAllowedValues(obj interface{}, rule ValidationRule) error {
	// Simplified implementation
	return nil
}

// validateKeywordLength checks if keyword array items meet length requirements
func validateKeywordLength(obj interface{}, rule ValidationRule) error {
	// Simplified implementation
	return nil
}

// =============================================================================
// Utility Validation Functions
// =============================================================================

// IsValidBusinessName checks if a business name is valid
func IsValidBusinessName(name string) bool {
	if name == "" || len(name) > 200 {
		return false
	}
	
	// Check for invalid characters
	pattern := regexp.MustCompile(`^[a-zA-Z0-9\s\-\.&'()]+$`)
	return pattern.MatchString(name)
}

// IsValidURL checks if a URL is valid
func IsValidURL(url string) bool {
	if url == "" {
		return false
	}
	
	pattern := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return pattern.MatchString(url)
}

// IsValidEmail checks if an email address is valid
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	
	pattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return pattern.MatchString(email)
}

// IsValidPhoneNumber checks if a phone number is valid
func IsValidPhoneNumber(phone string) bool {
	if phone == "" {
		return false
	}
	
	// Basic US phone number pattern
	pattern := regexp.MustCompile(`^\d{3}[-.]?\d{3}[-.]?\d{4}$`)
	return pattern.MatchString(phone)
}

// IsValidIndustryCode checks if an industry code is valid
func IsValidIndustryCode(code string) bool {
	if code == "" {
		return false
	}
	
	// Check if it's a valid NAICS code (6 digits)
	naicsPattern := regexp.MustCompile(`^\d{6}$`)
	if naicsPattern.MatchString(code) {
		return true
	}
	
	// Check if it's a valid SIC code (4 digits)
	sicPattern := regexp.MustCompile(`^\d{4}$`)
	if sicPattern.MatchString(code) {
		return true
	}
	
	// Check if it's a valid MCC code (4 digits)
	mccPattern := regexp.MustCompile(`^\d{4}$`)
	if mccPattern.MatchString(code) {
		return true
	}
	
	return false
}

// IsValidConfidenceScore checks if a confidence score is valid
func IsValidConfidenceScore(score float64) bool {
	return score >= 0.0 && score <= 1.0
}

// IsValidProcessingTime checks if a processing time is valid
func IsValidProcessingTime(duration time.Duration) bool {
	return duration >= 0 && duration <= 24*time.Hour // Max 24 hours
}

// IsValidClassificationMethodString checks if a classification method string is valid
func IsValidClassificationMethodString(method string) bool {
	validMethods := []string{
		"keyword",
		"ml",
		"website",
		"web_search",
		"ensemble",
		"hybrid",
	}
	
	for _, valid := range validMethods {
		if method == valid {
			return true
		}
	}
	return false
}

// IsValidModelType checks if a model type is valid
func IsValidModelType(modelType ModelType) bool {
	validTypes := []ModelType{
		ModelTypeBERT,
		ModelTypeEnsemble,
		ModelTypeTransformer,
		ModelTypeCustom,
	}
	
	for _, valid := range validTypes {
		if modelType == valid {
			return true
		}
	}
	return false
}

// IsValidModelStatus checks if a model status is valid
func IsValidModelStatus(status ModelStatus) bool {
	validStatuses := []ModelStatus{
		ModelStatusLoading,
		ModelStatusReady,
		ModelStatusError,
		ModelStatusUpdating,
		ModelStatusDeprecated,
	}
	
	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

// =============================================================================
// Sanitization Functions
// =============================================================================

// SanitizeBusinessName sanitizes a business name
func SanitizeBusinessName(name string) string {
	// Remove extra whitespace
	name = strings.TrimSpace(name)
	
	// Replace multiple spaces with single space
	spacePattern := regexp.MustCompile(`\s+`)
	name = spacePattern.ReplaceAllString(name, " ")
	
	// Remove invalid characters
	invalidPattern := regexp.MustCompile(`[^a-zA-Z0-9\s\-\.&'()]`)
	name = invalidPattern.ReplaceAllString(name, "")
	
	// Limit length
	if len(name) > 200 {
		name = name[:200]
	}
	
	return name
}

// SanitizeURL sanitizes a URL
func SanitizeURL(url string) string {
	// Remove extra whitespace
	url = strings.TrimSpace(url)
	
	// Ensure URL has protocol
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	
	// Remove invalid characters
	invalidPattern := regexp.MustCompile(`[^\w\-\.~:/?#[\]@!$&'()*+,;=%]`)
	url = invalidPattern.ReplaceAllString(url, "")
	
	return url
}

// SanitizeEmail sanitizes an email address
func SanitizeEmail(email string) string {
	// Remove extra whitespace and convert to lowercase
	email = strings.ToLower(strings.TrimSpace(email))
	
	// Remove invalid characters
	invalidPattern := regexp.MustCompile(`[^a-zA-Z0-9._%+-@]`)
	email = invalidPattern.ReplaceAllString(email, "")
	
	return email
}

// SanitizePhoneNumber sanitizes a phone number
func SanitizePhoneNumber(phone string) string {
	// Remove all non-digit characters
	digitPattern := regexp.MustCompile(`\D`)
	phone = digitPattern.ReplaceAllString(phone, "")
	
	// Format as XXX-XXX-XXXX if it's 10 digits
	if len(phone) == 10 {
		return fmt.Sprintf("%s-%s-%s", phone[:3], phone[3:6], phone[6:])
	}
	
	return phone
}

// SanitizeIndustryCode sanitizes an industry code
func SanitizeIndustryCode(code string) string {
	// Remove all non-digit characters
	digitPattern := regexp.MustCompile(`\D`)
	code = digitPattern.ReplaceAllString(code, "")
	
	// Pad with zeros if needed for NAICS (6 digits)
	if len(code) < 6 {
		code = strings.Repeat("0", 6-len(code)) + code
	}
	
	// Limit to 6 digits
	if len(code) > 6 {
		code = code[:6]
	}
	
	return code
}
