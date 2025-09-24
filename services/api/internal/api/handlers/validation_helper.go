package handlers

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ValidationHelper provides comprehensive validation with detailed error messages
type ValidationHelper struct {
	errorHandler *ErrorHandler
}

// NewValidationHelper creates a new validation helper
func NewValidationHelper(errorHandler *ErrorHandler) *ValidationHelper {
	return &ValidationHelper{
		errorHandler: errorHandler,
	}
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}
}

// AddError adds an error to the validation result
func (r *ValidationResult) AddError(err ValidationError) {
	r.Valid = false
	r.Errors = append(r.Errors, err)
}

// addValidationError is a helper method to add validation errors from the error handler
func (h *ValidationHelper) addValidationError(result *ValidationResult, err error) {
	if validationErr, ok := err.(*ValidationError); ok {
		result.AddError(*validationErr)
	}
}

// BusinessClassificationRequest represents a business classification request
type BusinessClassificationRequest struct {
	BusinessName     string `json:"business_name"`
	BusinessType     string `json:"business_type,omitempty"`
	Industry         string `json:"industry,omitempty"`
	Description      string `json:"description,omitempty"`
	Keywords         string `json:"keywords,omitempty"`
	WebsiteURL       string `json:"website_url,omitempty"`
	GeographicRegion string `json:"geographic_region,omitempty"`
	Email            string `json:"email,omitempty"`
	Phone            string `json:"phone,omitempty"`
	Address          string `json:"address,omitempty"`
}

// ValidateBusinessClassificationRequest validates a business classification request
func (h *ValidationHelper) ValidateBusinessClassificationRequest(req *BusinessClassificationRequest) *ValidationResult {
	result := NewValidationResult()

	// Validate required fields
	if strings.TrimSpace(req.BusinessName) == "" {
		err := h.errorHandler.CreateValidationError(
			ErrorCodeMissingRequiredField,
			"Business name is required for classification",
			"business_name",
			"required",
			req.BusinessName,
		)
		h.addValidationError(result, err)
	} else {
		// Validate business name length first
		if utf8.RuneCountInString(req.BusinessName) > 255 {
			err := h.errorHandler.CreateValidationError(
				ErrorCodeFieldTooLong,
				"Business name must be 255 characters or less",
				"business_name",
				"max_length",
				req.BusinessName,
			)
			h.addValidationError(result, err)
		} else {
			// Only validate format if length is valid
			if !h.isValidBusinessName(req.BusinessName) {
				err := h.errorHandler.CreateValidationError(
					ErrorCodeInvalidFieldFormat,
					"Business name contains invalid characters. Use only letters, numbers, spaces, and common punctuation",
					"business_name",
					"format",
					req.BusinessName,
				)
				h.addValidationError(result, err)
			}
		}
	}

	// Validate optional fields if provided
	if req.WebsiteURL != "" {
		if !h.isValidURL(req.WebsiteURL) {
			err := h.errorHandler.CreateValidationError(
				ErrorCodeInvalidURL,
				"Website URL must be a valid URL starting with http:// or https://",
				"website_url",
				"url_format",
				req.WebsiteURL,
			)
			h.addValidationError(result, err)
		}
	}

	if req.Email != "" {
		if !h.isValidEmail(req.Email) {
			err := h.errorHandler.CreateValidationError(
				ErrorCodeInvalidEmail,
				"Email must be a valid email address (e.g., user@example.com)",
				"email",
				"email_format",
				req.Email,
			)
			h.addValidationError(result, err)
		}
	}

	if req.Phone != "" {
		if !h.isValidPhone(req.Phone) {
			err := h.errorHandler.CreateValidationError(
				ErrorCodeInvalidPhone,
				"Phone number must be in E.164 format (e.g., +1-555-123-4567)",
				"phone",
				"phone_format",
				req.Phone,
			)
			h.addValidationError(result, err)
		}
	}

	// Validate field lengths
	if req.Description != "" && utf8.RuneCountInString(req.Description) > 1000 {
		err := h.errorHandler.CreateValidationError(
			ErrorCodeFieldTooLong,
			"Description must be 1000 characters or less",
			"description",
			"max_length",
			req.Description,
		)
		h.addValidationError(result, err)
	}

	if req.Keywords != "" && utf8.RuneCountInString(req.Keywords) > 500 {
		err := h.errorHandler.CreateValidationError(
			ErrorCodeFieldTooLong,
			"Keywords must be 500 characters or less",
			"keywords",
			"max_length",
			req.Keywords,
		)
		h.addValidationError(result, err)
	}

	if req.Address != "" && utf8.RuneCountInString(req.Address) > 500 {
		err := h.errorHandler.CreateValidationError(
			ErrorCodeFieldTooLong,
			"Address must be 500 characters or less",
			"address",
			"max_length",
			req.Address,
		)
		h.addValidationError(result, err)
	}

	return result
}

// BatchClassificationRequest represents a batch classification request
type BatchClassificationRequest struct {
	Businesses       []BusinessClassificationRequest `json:"businesses"`
	GeographicRegion string                          `json:"geographic_region,omitempty"`
}

// ValidateBatchClassificationRequest validates a batch classification request
func (h *ValidationHelper) ValidateBatchClassificationRequest(req *BatchClassificationRequest) *ValidationResult {
	result := NewValidationResult()

	// Validate businesses array
	if len(req.Businesses) == 0 {
		err := h.errorHandler.CreateValidationError(
			ErrorCodeMissingRequiredField,
			"At least one business must be provided for batch classification",
			"businesses",
			"required",
			req.Businesses,
		)
		h.addValidationError(result, err)
	} else if len(req.Businesses) > 100 {
		err := h.errorHandler.CreateValidationError(
			ErrorCodeFieldTooLong,
			"Batch classification supports a maximum of 100 businesses per request",
			"businesses",
			"max_count",
			len(req.Businesses),
		)
		h.addValidationError(result, err)
	} else {
		// Validate each business in the batch
		for i, business := range req.Businesses {
			businessResult := h.ValidateBusinessClassificationRequest(&business)
			if !businessResult.Valid {
				for _, err := range businessResult.Errors {
					// Add context about which business in the batch has the error
					err.Field = fmt.Sprintf("businesses[%d].%s", i, err.Field)
					h.addValidationError(result, err)
				}
			}
		}
	}

	return result
}

// WebsiteVerificationRequest represents a website verification request
type WebsiteVerificationRequest struct {
	BusinessName string `json:"business_name"`
	WebsiteURL   string `json:"website_url"`
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Address      string `json:"address,omitempty"`
}

// ValidateWebsiteVerificationRequest validates a website verification request
func (h *ValidationHelper) ValidateWebsiteVerificationRequest(req *WebsiteVerificationRequest) *ValidationResult {
	result := NewValidationResult()

	// Validate required fields
	if strings.TrimSpace(req.BusinessName) == "" {
		err := h.errorHandler.CreateValidationError(
			ErrorCodeMissingRequiredField,
			"Business name is required for website verification",
			"business_name",
			"required",
			req.BusinessName,
		)
		h.addValidationError(result, err)
	}

	if strings.TrimSpace(req.WebsiteURL) == "" {
		err := h.errorHandler.CreateValidationError(
			ErrorCodeMissingRequiredField,
			"Website URL is required for website verification",
			"website_url",
			"required",
			req.WebsiteURL,
		)
		h.addValidationError(result, err)
	} else if !h.isValidURL(req.WebsiteURL) {
		err := h.errorHandler.CreateValidationError(
			ErrorCodeInvalidURL,
			"Website URL must be a valid URL starting with http:// or https://",
			"website_url",
			"url_format",
			req.WebsiteURL,
		)
		h.addValidationError(result, err)
	}

	// Validate optional fields
	if req.Email != "" && !h.isValidEmail(req.Email) {
		err := h.errorHandler.CreateValidationError(
			ErrorCodeInvalidEmail,
			"Email must be a valid email address (e.g., user@example.com)",
			"email",
			"email_format",
			req.Email,
		)
		h.addValidationError(result, err)
	}

	if req.Phone != "" && !h.isValidPhone(req.Phone) {
		err := h.errorHandler.CreateValidationError(
			ErrorCodeInvalidPhone,
			"Phone number must be in E.164 format (e.g., +1-555-123-4567)",
			"phone",
			"phone_format",
			req.Phone,
		)
		h.addValidationError(result, err)
	}

	return result
}

// RiskAssessmentRequest represents a risk assessment request
type RiskAssessmentRequest struct {
	BusinessID        string `json:"business_id"`
	AssessmentType    string `json:"assessment_type,omitempty"`
	IncludeFinancial  bool   `json:"include_financial,omitempty"`
	IncludeCompliance bool   `json:"include_compliance,omitempty"`
}

// ValidateRiskAssessmentRequest validates a risk assessment request
func (h *ValidationHelper) ValidateRiskAssessmentRequest(req *RiskAssessmentRequest) *ValidationResult {
	result := NewValidationResult()

	// Validate required fields
	if strings.TrimSpace(req.BusinessID) == "" {
		err := h.errorHandler.CreateValidationError(
			ErrorCodeMissingRequiredField,
			"Business ID is required for risk assessment",
			"business_id",
			"required",
			req.BusinessID,
		)
		h.addValidationError(result, err)
	} else if !h.isValidBusinessID(req.BusinessID) {
		err := h.errorHandler.CreateValidationError(
			ErrorCodeInvalidFieldFormat,
			"Business ID must be a valid UUID format",
			"business_id",
			"uuid_format",
			req.BusinessID,
		)
		h.addValidationError(result, err)
	}

	// Validate assessment type if provided
	if req.AssessmentType != "" {
		validTypes := []string{"basic", "comprehensive", "financial", "compliance", "security"}
		isValid := false
		for _, validType := range validTypes {
			if req.AssessmentType == validType {
				isValid = true
				break
			}
		}
		if !isValid {
			err := h.errorHandler.CreateValidationError(
				ErrorCodeInvalidFieldFormat,
				fmt.Sprintf("Assessment type must be one of: %s", strings.Join(validTypes, ", ")),
				"assessment_type",
				"enum",
				req.AssessmentType,
			)
			h.addValidationError(result, err)
		}
	}

	return result
}

// Helper methods for validation

// isValidBusinessName validates business name format
func (h *ValidationHelper) isValidBusinessName(name string) bool {
	// Allow letters, numbers, spaces, and common punctuation
	pattern := `^[a-zA-Z0-9\s\-&.,'()]+$`
	matched, _ := regexp.MatchString(pattern, name)
	return matched && len(strings.TrimSpace(name)) > 0
}

// isValidURL validates URL format
func (h *ValidationHelper) isValidURL(urlStr string) bool {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

// isValidEmail validates email format
func (h *ValidationHelper) isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched && len(email) <= 254
}

// isValidPhone validates phone number format (E.164)
func (h *ValidationHelper) isValidPhone(phone string) bool {
	// More flexible E.164 format validation - allow common separators
	// Remove common separators first, then validate E.164 format
	cleanPhone := strings.ReplaceAll(phone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "(", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ")", "")

	// E.164 format: +[country code][national number] (7-15 digits total)
	// But we'll be more strict for this API - require at least 11 digits total
	pattern := `^\+[1-9]\d{10,14}$`
	matched, _ := regexp.MatchString(pattern, cleanPhone)
	return matched
}

// isValidBusinessID validates business ID format (UUID)
func (h *ValidationHelper) isValidBusinessID(id string) bool {
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(pattern, strings.ToLower(id))
	return matched
}

// GetValidationErrorMessage provides detailed error messages with actionable guidance
func (h *ValidationHelper) GetValidationErrorMessage(err error) string {
	if validationErr, ok := err.(*ValidationError); ok {
		switch validationErr.Code {
		case string(ErrorCodeMissingRequiredField):
			return fmt.Sprintf("The field '%s' is required. Please provide a value for this field.", validationErr.Field)
		case string(ErrorCodeFieldTooLong):
			return fmt.Sprintf("The field '%s' is too long. Please shorten the value to meet the length requirement.", validationErr.Field)
		case string(ErrorCodeFieldTooShort):
			return fmt.Sprintf("The field '%s' is too short. Please provide a longer value.", validationErr.Field)
		case string(ErrorCodeInvalidURL):
			return "Please provide a valid URL starting with http:// or https://"
		case string(ErrorCodeInvalidEmail):
			return "Please provide a valid email address in the format user@example.com"
		case string(ErrorCodeInvalidPhone):
			return "Please provide a valid phone number in E.164 format (e.g., +1-555-123-4567)"
		case string(ErrorCodeInvalidFieldFormat):
			return fmt.Sprintf("The field '%s' has an invalid format. %s", validationErr.Field, validationErr.Message)
		default:
			return validationErr.Message
		}
	}
	return err.Error()
}

// GetValidationHelpURL provides help URLs for different validation errors
func (h *ValidationHelper) GetValidationHelpURL(err error) string {
	if validationErr, ok := err.(*ValidationError); ok {
		switch validationErr.Code {
		case string(ErrorCodeMissingRequiredField):
			return "https://docs.kyb-platform.com/api/validation/required-fields"
		case string(ErrorCodeFieldTooLong):
			return "https://docs.kyb-platform.com/api/validation/field-lengths"
		case string(ErrorCodeInvalidURL):
			return "https://docs.kyb-platform.com/api/validation/urls"
		case string(ErrorCodeInvalidEmail):
			return "https://docs.kyb-platform.com/api/validation/emails"
		case string(ErrorCodeInvalidPhone):
			return "https://docs.kyb-platform.com/api/validation/phone-numbers"
		default:
			return "https://docs.kyb-platform.com/api/validation"
		}
	}
	return "https://docs.kyb-platform.com/api/errors"
}
