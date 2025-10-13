package validation

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// CountryDataValidator provides comprehensive multi-country data validation
type CountryDataValidator struct {
	logger *zap.Logger
	config *CountryDataValidatorConfig
}

// CountryDataValidatorConfig represents configuration for country data validation
type CountryDataValidatorConfig struct {
	SupportedCountries       []string                `json:"supported_countries"`
	ValidationRules          map[string]CountryRules `json:"validation_rules"`
	AccuracyThreshold        float64                 `json:"accuracy_threshold"`
	ComplianceThreshold      float64                 `json:"compliance_threshold"`
	EnableRealTimeValidation bool                    `json:"enable_real_time_validation"`
	EnableBatchValidation    bool                    `json:"enable_batch_validation"`
	ValidationTimeout        time.Duration           `json:"validation_timeout"`
	RetryAttempts            int                     `json:"retry_attempts"`
	Metadata                 map[string]interface{}  `json:"metadata"`
}

// CountryRules represents validation rules for a specific country
type CountryRules struct {
	CountryCode        string                 `json:"country_code"`
	CountryName        string                 `json:"country_name"`
	BusinessIDRules    []BusinessIDRule       `json:"business_id_rules"`
	TaxIDRules         []TaxIDRule            `json:"tax_id_rules"`
	AddressRules       []AddressRule          `json:"address_rules"`
	PhoneRules         []PhoneRule            `json:"phone_rules"`
	EmailRules         []EmailRule            `json:"email_rules"`
	WebsiteRules       []WebsiteRule          `json:"website_rules"`
	ComplianceRules    []CountryComplianceRule `json:"compliance_rules"`
	DataResidencyRules DataResidencyRule      `json:"data_residency_rules"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// BusinessIDRule represents business ID validation rules
type BusinessIDRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Pattern     string                 `json:"pattern"`
	Required    bool                   `json:"required"`
	MinLength   int                    `json:"min_length"`
	MaxLength   int                    `json:"max_length"`
	Checksum    bool                   `json:"checksum"`
	Examples    []string               `json:"examples"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TaxIDRule represents tax ID validation rules
type TaxIDRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Pattern     string                 `json:"pattern"`
	Required    bool                   `json:"required"`
	MinLength   int                    `json:"min_length"`
	MaxLength   int                    `json:"max_length"`
	Checksum    bool                   `json:"checksum"`
	Examples    []string               `json:"examples"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AddressRule represents address validation rules
type AddressRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	MinLength   int                    `json:"min_length"`
	MaxLength   int                    `json:"max_length"`
	Components  []AddressComponent     `json:"components"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AddressComponent represents an address component
type AddressComponent struct {
	Name      string `json:"name"`
	Required  bool   `json:"required"`
	MinLength int    `json:"min_length"`
	MaxLength int    `json:"max_length"`
	Pattern   string `json:"pattern"`
}

// PhoneRule represents phone number validation rules
type PhoneRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Pattern     string                 `json:"pattern"`
	Required    bool                   `json:"required"`
	MinLength   int                    `json:"min_length"`
	MaxLength   int                    `json:"max_length"`
	CountryCode string                 `json:"country_code"`
	Examples    []string               `json:"examples"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// EmailRule represents email validation rules
type EmailRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Pattern     string                 `json:"pattern"`
	Required    bool                   `json:"required"`
	MaxLength   int                    `json:"max_length"`
	Examples    []string               `json:"examples"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// WebsiteRule represents website validation rules
type WebsiteRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Pattern     string                 `json:"pattern"`
	Required    bool                   `json:"required"`
	MaxLength   int                    `json:"max_length"`
	Examples    []string               `json:"examples"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CountryComplianceRule represents compliance validation rules for country-specific validation
type CountryComplianceRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Regulation  string                 `json:"regulation"`
	Required    bool                   `json:"required"`
	Validation  string                 `json:"validation"`
	Examples    []string               `json:"examples"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DataResidencyRule represents data residency rules
type DataResidencyRule struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	AllowedRegions     []string               `json:"allowed_regions"`
	RestrictedRegions  []string               `json:"restricted_regions"`
	EncryptionRequired bool                   `json:"encryption_required"`
	RetentionPeriod    time.Duration          `json:"retention_period"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// ValidationResult represents the result of data validation
type ValidationResult struct {
	ID              string                     `json:"id"`
	CountryCode     string                     `json:"country_code"`
	ValidationType  string                     `json:"validation_type"`
	Status          ValidationStatus           `json:"status"`
	Accuracy        float64                    `json:"accuracy"`
	Compliance      float64                    `json:"compliance"`
	ValidatedFields []ValidatedField           `json:"validated_fields"`
	Errors          []ValidationError          `json:"errors"`
	Warnings        []ValidationWarning        `json:"warnings"`
	Recommendations []ValidationRecommendation `json:"recommendations"`
	Timestamp       time.Time                  `json:"timestamp"`
	Metadata        map[string]interface{}     `json:"metadata"`
}

// ValidationStatus represents the status of validation
type ValidationStatus string

const (
	ValidationStatusValid      ValidationStatus = "valid"
	ValidationStatusInvalid    ValidationStatus = "invalid"
	ValidationStatusWarning    ValidationStatus = "warning"
	ValidationStatusError      ValidationStatus = "error"
	ValidationStatusIncomplete ValidationStatus = "incomplete"
)

// ValidatedField represents a validated field
type ValidatedField struct {
	FieldName   string                 `json:"field_name"`
	FieldValue  string                 `json:"field_value"`
	IsValid     bool                   `json:"is_valid"`
	Accuracy    float64                `json:"accuracy"`
	Confidence  float64                `json:"confidence"`
	Issues      []string               `json:"issues"`
	Suggestions []string               `json:"suggestions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ValidationError represents a validation error
type ValidationError struct {
	ID        string                 `json:"id"`
	FieldName string                 `json:"field_name"`
	ErrorCode string                 `json:"error_code"`
	Message   string                 `json:"message"`
	Severity  ErrorSeverity          `json:"severity"`
	RuleID    string                 `json:"rule_id"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	ID          string                 `json:"id"`
	FieldName   string                 `json:"field_name"`
	WarningCode string                 `json:"warning_code"`
	Message     string                 `json:"message"`
	Severity    WarningSeverity        `json:"severity"`
	RuleID      string                 `json:"rule_id"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ValidationRecommendation represents a validation recommendation
type ValidationRecommendation struct {
	ID        string                 `json:"id"`
	FieldName string                 `json:"field_name"`
	Type      RecommendationType     `json:"type"`
	Message   string                 `json:"message"`
	Priority  RecommendationPriority `json:"priority"`
	Action    string                 `json:"action"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ErrorSeverity represents the severity of an error
type ErrorSeverity string

const (
	ErrorSeverityCritical ErrorSeverity = "critical"
	ErrorSeverityHigh     ErrorSeverity = "high"
	ErrorSeverityMedium   ErrorSeverity = "medium"
	ErrorSeverityLow      ErrorSeverity = "low"
)

// WarningSeverity represents the severity of a warning
type WarningSeverity string

const (
	WarningSeverityHigh   WarningSeverity = "high"
	WarningSeverityMedium WarningSeverity = "medium"
	WarningSeverityLow    WarningSeverity = "low"
)

// RecommendationType represents the type of recommendation
type RecommendationType string

const (
	RecommendationTypeDataQuality  RecommendationType = "data_quality"
	RecommendationTypeCompliance   RecommendationType = "compliance"
	RecommendationTypeAccuracy     RecommendationType = "accuracy"
	RecommendationTypeCompleteness RecommendationType = "completeness"
)

// RecommendationPriority represents the priority of a recommendation
type RecommendationPriority string

const (
	RecommendationPriorityCritical RecommendationPriority = "critical"
	RecommendationPriorityHigh     RecommendationPriority = "high"
	RecommendationPriorityMedium   RecommendationPriority = "medium"
	RecommendationPriorityLow      RecommendationPriority = "low"
)

// NewCountryDataValidator creates a new country data validator
func NewCountryDataValidator(logger *zap.Logger, config *CountryDataValidatorConfig) *CountryDataValidator {
	return &CountryDataValidator{
		logger: logger,
		config: config,
	}
}

// ValidateBusinessData validates business data for a specific country
func (cdv *CountryDataValidator) ValidateBusinessData(ctx context.Context, countryCode string, data map[string]interface{}) (*ValidationResult, error) {
	result := &ValidationResult{
		ID:             fmt.Sprintf("validation_%d", time.Now().UnixNano()),
		CountryCode:    countryCode,
		ValidationType: "business_data",
		Status:         ValidationStatusValid,
		Timestamp:      time.Now(),
		Metadata:       make(map[string]interface{}),
	}

	cdv.logger.Info("Starting business data validation",
		zap.String("country_code", countryCode),
		zap.String("validation_id", result.ID))

	// Get country rules
	rules, exists := cdv.config.ValidationRules[countryCode]
	if !exists {
		return nil, fmt.Errorf("unsupported country code: %s", countryCode)
	}

	// Validate business ID
	if businessID, exists := data["business_id"]; exists {
		fieldResult := cdv.validateBusinessID(businessID.(string), rules.BusinessIDRules)
		result.ValidatedFields = append(result.ValidatedFields, *fieldResult)
	}

	// Validate tax ID
	if taxID, exists := data["tax_id"]; exists {
		fieldResult := cdv.validateTaxID(taxID.(string), rules.TaxIDRules)
		result.ValidatedFields = append(result.ValidatedFields, *fieldResult)
	}

	// Validate address
	if address, exists := data["address"]; exists {
		fieldResult := cdv.validateAddress(address.(string), rules.AddressRules)
		result.ValidatedFields = append(result.ValidatedFields, *fieldResult)
	}

	// Validate phone
	if phone, exists := data["phone"]; exists {
		fieldResult := cdv.validatePhone(phone.(string), rules.PhoneRules)
		result.ValidatedFields = append(result.ValidatedFields, *fieldResult)
	}

	// Validate email
	if email, exists := data["email"]; exists {
		fieldResult := cdv.validateEmail(email.(string), rules.EmailRules)
		result.ValidatedFields = append(result.ValidatedFields, *fieldResult)
	}

	// Validate website
	if website, exists := data["website"]; exists {
		fieldResult := cdv.validateWebsite(website.(string), rules.WebsiteRules)
		result.ValidatedFields = append(result.ValidatedFields, *fieldResult)
	}

	// Calculate overall accuracy and compliance
	cdv.calculateOverallMetrics(result)

	// Determine final status
	cdv.determineValidationStatus(result)

	cdv.logger.Info("Business data validation completed",
		zap.String("validation_id", result.ID),
		zap.Float64("accuracy", result.Accuracy),
		zap.Float64("compliance", result.Compliance),
		zap.String("status", string(result.Status)))

	return result, nil
}

// validateBusinessID validates business ID against country rules
func (cdv *CountryDataValidator) validateBusinessID(businessID string, rules []BusinessIDRule) *ValidatedField {
	field := &ValidatedField{
		FieldName:   "business_id",
		FieldValue:  businessID,
		IsValid:     true,
		Accuracy:    1.0,
		Confidence:  1.0,
		Issues:      make([]string, 0),
		Suggestions: make([]string, 0),
		Metadata:    make(map[string]interface{}),
	}

	for _, rule := range rules {
		// Check required
		if rule.Required && businessID == "" {
			field.IsValid = false
			field.Issues = append(field.Issues, fmt.Sprintf("Business ID is required for %s", rule.Name))
			continue
		}

		if businessID == "" {
			continue
		}

		// Check length
		if len(businessID) < rule.MinLength || len(businessID) > rule.MaxLength {
			field.IsValid = false
			field.Issues = append(field.Issues, fmt.Sprintf("Business ID length must be between %d and %d characters", rule.MinLength, rule.MaxLength))
		}

		// Check pattern
		if rule.Pattern != "" {
			matched, err := regexp.MatchString(rule.Pattern, businessID)
			if err != nil {
				cdv.logger.Error("Invalid regex pattern", zap.String("pattern", rule.Pattern), zap.Error(err))
				continue
			}

			if !matched {
				field.IsValid = false
				field.Issues = append(field.Issues, fmt.Sprintf("Business ID does not match required pattern for %s", rule.Name))
			}
		}

		// Check checksum if required
		if rule.Checksum {
			if !cdv.validateChecksum(businessID, rule.Name) {
				field.IsValid = false
				field.Issues = append(field.Issues, fmt.Sprintf("Business ID checksum validation failed for %s", rule.Name))
			}
		}
	}

	// Calculate accuracy based on validation results
	if len(field.Issues) > 0 {
		field.Accuracy = 1.0 - (float64(len(field.Issues)) / float64(len(rules)))
		field.Confidence = field.Accuracy
	}

	return field
}

// validateTaxID validates tax ID against country rules
func (cdv *CountryDataValidator) validateTaxID(taxID string, rules []TaxIDRule) *ValidatedField {
	field := &ValidatedField{
		FieldName:   "tax_id",
		FieldValue:  taxID,
		IsValid:     true,
		Accuracy:    1.0,
		Confidence:  1.0,
		Issues:      make([]string, 0),
		Suggestions: make([]string, 0),
		Metadata:    make(map[string]interface{}),
	}

	for _, rule := range rules {
		// Check required
		if rule.Required && taxID == "" {
			field.IsValid = false
			field.Issues = append(field.Issues, fmt.Sprintf("Tax ID is required for %s", rule.Name))
			continue
		}

		if taxID == "" {
			continue
		}

		// Check length
		if len(taxID) < rule.MinLength || len(taxID) > rule.MaxLength {
			field.IsValid = false
			field.Issues = append(field.Issues, fmt.Sprintf("Tax ID length must be between %d and %d characters", rule.MinLength, rule.MaxLength))
		}

		// Check pattern
		if rule.Pattern != "" {
			matched, err := regexp.MatchString(rule.Pattern, taxID)
			if err != nil {
				cdv.logger.Error("Invalid regex pattern", zap.String("pattern", rule.Pattern), zap.Error(err))
				continue
			}

			if !matched {
				field.IsValid = false
				field.Issues = append(field.Issues, fmt.Sprintf("Tax ID does not match required pattern for %s", rule.Name))
			}
		}

		// Check checksum if required
		if rule.Checksum {
			if !cdv.validateChecksum(taxID, rule.Name) {
				field.IsValid = false
				field.Issues = append(field.Issues, fmt.Sprintf("Tax ID checksum validation failed for %s", rule.Name))
			}
		}
	}

	// Calculate accuracy based on validation results
	if len(field.Issues) > 0 {
		field.Accuracy = 1.0 - (float64(len(field.Issues)) / float64(len(rules)))
		field.Confidence = field.Accuracy
	}

	return field
}

// validateAddress validates address against country rules
func (cdv *CountryDataValidator) validateAddress(address string, rules []AddressRule) *ValidatedField {
	field := &ValidatedField{
		FieldName:   "address",
		FieldValue:  address,
		IsValid:     true,
		Accuracy:    1.0,
		Confidence:  1.0,
		Issues:      make([]string, 0),
		Suggestions: make([]string, 0),
		Metadata:    make(map[string]interface{}),
	}

	for _, rule := range rules {
		// Check required
		if rule.Required && address == "" {
			field.IsValid = false
			field.Issues = append(field.Issues, fmt.Sprintf("Address is required for %s", rule.Name))
			continue
		}

		if address == "" {
			continue
		}

		// Check length
		if len(address) < rule.MinLength || len(address) > rule.MaxLength {
			field.IsValid = false
			field.Issues = append(field.Issues, fmt.Sprintf("Address length must be between %d and %d characters", rule.MinLength, rule.MaxLength))
		}

		// Validate address components
		for _, component := range rule.Components {
			if component.Required {
				// This is a simplified validation - in a real implementation,
				// you would parse the address and validate each component
				if !cdv.validateAddressComponent(address, component) {
					field.IsValid = false
					field.Issues = append(field.Issues, fmt.Sprintf("Address component '%s' is required", component.Name))
				}
			}
		}
	}

	// Calculate accuracy based on validation results
	if len(field.Issues) > 0 {
		field.Accuracy = 1.0 - (float64(len(field.Issues)) / float64(len(rules)))
		field.Confidence = field.Accuracy
	}

	return field
}

// validatePhone validates phone number against country rules
func (cdv *CountryDataValidator) validatePhone(phone string, rules []PhoneRule) *ValidatedField {
	field := &ValidatedField{
		FieldName:   "phone",
		FieldValue:  phone,
		IsValid:     true,
		Accuracy:    1.0,
		Confidence:  1.0,
		Issues:      make([]string, 0),
		Suggestions: make([]string, 0),
		Metadata:    make(map[string]interface{}),
	}

	for _, rule := range rules {
		// Check required
		if rule.Required && phone == "" {
			field.IsValid = false
			field.Issues = append(field.Issues, fmt.Sprintf("Phone number is required for %s", rule.Name))
			continue
		}

		if phone == "" {
			continue
		}

		// Check length
		if len(phone) < rule.MinLength || len(phone) > rule.MaxLength {
			field.IsValid = false
			field.Issues = append(field.Issues, fmt.Sprintf("Phone number length must be between %d and %d characters", rule.MinLength, rule.MaxLength))
		}

		// Check pattern
		if rule.Pattern != "" {
			matched, err := regexp.MatchString(rule.Pattern, phone)
			if err != nil {
				cdv.logger.Error("Invalid regex pattern", zap.String("pattern", rule.Pattern), zap.Error(err))
				continue
			}

			if !matched {
				field.IsValid = false
				field.Issues = append(field.Issues, fmt.Sprintf("Phone number does not match required pattern for %s", rule.Name))
			}
		}
	}

	// Calculate accuracy based on validation results
	if len(field.Issues) > 0 {
		field.Accuracy = 1.0 - (float64(len(field.Issues)) / float64(len(rules)))
		field.Confidence = field.Accuracy
	}

	return field
}

// validateEmail validates email against country rules
func (cdv *CountryDataValidator) validateEmail(email string, rules []EmailRule) *ValidatedField {
	field := &ValidatedField{
		FieldName:   "email",
		FieldValue:  email,
		IsValid:     true,
		Accuracy:    1.0,
		Confidence:  1.0,
		Issues:      make([]string, 0),
		Suggestions: make([]string, 0),
		Metadata:    make(map[string]interface{}),
	}

	for _, rule := range rules {
		// Check required
		if rule.Required && email == "" {
			field.IsValid = false
			field.Issues = append(field.Issues, fmt.Sprintf("Email is required for %s", rule.Name))
			continue
		}

		if email == "" {
			continue
		}

		// Check length
		if len(email) > rule.MaxLength {
			field.IsValid = false
			field.Issues = append(field.Issues, fmt.Sprintf("Email length must not exceed %d characters", rule.MaxLength))
		}

		// Check pattern
		if rule.Pattern != "" {
			matched, err := regexp.MatchString(rule.Pattern, email)
			if err != nil {
				cdv.logger.Error("Invalid regex pattern", zap.String("pattern", rule.Pattern), zap.Error(err))
				continue
			}

			if !matched {
				field.IsValid = false
				field.Issues = append(field.Issues, fmt.Sprintf("Email does not match required pattern for %s", rule.Name))
			}
		}
	}

	// Calculate accuracy based on validation results
	if len(field.Issues) > 0 {
		field.Accuracy = 1.0 - (float64(len(field.Issues)) / float64(len(rules)))
		field.Confidence = field.Accuracy
	}

	return field
}

// validateWebsite validates website against country rules
func (cdv *CountryDataValidator) validateWebsite(website string, rules []WebsiteRule) *ValidatedField {
	field := &ValidatedField{
		FieldName:   "website",
		FieldValue:  website,
		IsValid:     true,
		Accuracy:    1.0,
		Confidence:  1.0,
		Issues:      make([]string, 0),
		Suggestions: make([]string, 0),
		Metadata:    make(map[string]interface{}),
	}

	for _, rule := range rules {
		// Check required
		if rule.Required && website == "" {
			field.IsValid = false
			field.Issues = append(field.Issues, fmt.Sprintf("Website is required for %s", rule.Name))
			continue
		}

		if website == "" {
			continue
		}

		// Check length
		if len(website) > rule.MaxLength {
			field.IsValid = false
			field.Issues = append(field.Issues, fmt.Sprintf("Website length must not exceed %d characters", rule.MaxLength))
		}

		// Check pattern
		if rule.Pattern != "" {
			matched, err := regexp.MatchString(rule.Pattern, website)
			if err != nil {
				cdv.logger.Error("Invalid regex pattern", zap.String("pattern", rule.Pattern), zap.Error(err))
				continue
			}

			if !matched {
				field.IsValid = false
				field.Issues = append(field.Issues, fmt.Sprintf("Website does not match required pattern for %s", rule.Name))
			}
		}
	}

	// Calculate accuracy based on validation results
	if len(field.Issues) > 0 {
		field.Accuracy = 1.0 - (float64(len(field.Issues)) / float64(len(rules)))
		field.Confidence = field.Accuracy
	}

	return field
}

// validateAddressComponent validates an address component
func (cdv *CountryDataValidator) validateAddressComponent(address string, component AddressComponent) bool {
	// This is a simplified validation - in a real implementation,
	// you would parse the address and validate each component
	// For now, we'll just check if the component name appears in the address
	return strings.Contains(strings.ToLower(address), strings.ToLower(component.Name))
}

// validateChecksum validates checksum for business ID or tax ID
func (cdv *CountryDataValidator) validateChecksum(id string, ruleName string) bool {
	// This is a simplified checksum validation - in a real implementation,
	// you would implement the specific checksum algorithm for each country
	// For now, we'll just return true for demonstration
	return true
}

// calculateOverallMetrics calculates overall accuracy and compliance metrics
func (cdv *CountryDataValidator) calculateOverallMetrics(result *ValidationResult) {
	if len(result.ValidatedFields) == 0 {
		result.Accuracy = 0.0
		result.Compliance = 0.0
		return
	}

	totalAccuracy := 0.0
	totalCompliance := 0.0

	for _, field := range result.ValidatedFields {
		totalAccuracy += field.Accuracy
		if field.IsValid {
			totalCompliance += 1.0
		}
	}

	result.Accuracy = totalAccuracy / float64(len(result.ValidatedFields))
	result.Compliance = totalCompliance / float64(len(result.ValidatedFields))
}

// determineValidationStatus determines the final validation status
func (cdv *CountryDataValidator) determineValidationStatus(result *ValidationResult) {
	if result.Accuracy >= cdv.config.AccuracyThreshold && result.Compliance >= cdv.config.ComplianceThreshold {
		result.Status = ValidationStatusValid
	} else if result.Accuracy >= cdv.config.AccuracyThreshold*0.8 && result.Compliance >= cdv.config.ComplianceThreshold*0.8 {
		result.Status = ValidationStatusWarning
	} else if result.Accuracy >= cdv.config.AccuracyThreshold*0.5 && result.Compliance >= cdv.config.ComplianceThreshold*0.5 {
		result.Status = ValidationStatusIncomplete
	} else {
		result.Status = ValidationStatusInvalid
	}
}

// GetSupportedCountries returns the list of supported countries
func (cdv *CountryDataValidator) GetSupportedCountries() []string {
	return cdv.config.SupportedCountries
}

// GetCountryRules returns validation rules for a specific country
func (cdv *CountryDataValidator) GetCountryRules(countryCode string) (*CountryRules, error) {
	rules, exists := cdv.config.ValidationRules[countryCode]
	if !exists {
		return nil, fmt.Errorf("unsupported country code: %s", countryCode)
	}

	return &rules, nil
}

// ValidateCompliance validates compliance for a specific country
func (cdv *CountryDataValidator) ValidateCompliance(ctx context.Context, countryCode string, data map[string]interface{}) (*ValidationResult, error) {
	result := &ValidationResult{
		ID:             fmt.Sprintf("compliance_validation_%d", time.Now().UnixNano()),
		CountryCode:    countryCode,
		ValidationType: "compliance",
		Status:         ValidationStatusValid,
		Timestamp:      time.Now(),
		Metadata:       make(map[string]interface{}),
	}

	cdv.logger.Info("Starting compliance validation",
		zap.String("country_code", countryCode),
		zap.String("validation_id", result.ID))

	// Get country rules
	rules, exists := cdv.config.ValidationRules[countryCode]
	if !exists {
		return nil, fmt.Errorf("unsupported country code: %s", countryCode)
	}

	// Validate compliance rules
	for _, rule := range rules.ComplianceRules {
		if rule.Required {
			field := &ValidatedField{
				FieldName:   rule.Name,
				FieldValue:  fmt.Sprintf("%v", data[rule.Name]),
				IsValid:     true,
				Accuracy:    1.0,
				Confidence:  1.0,
				Issues:      make([]string, 0),
				Suggestions: make([]string, 0),
				Metadata:    make(map[string]interface{}),
			}

			// Check if required field exists
			if _, exists := data[rule.Name]; !exists {
				field.IsValid = false
				field.Issues = append(field.Issues, fmt.Sprintf("Required compliance field '%s' is missing", rule.Name))
			}

			result.ValidatedFields = append(result.ValidatedFields, *field)
		}
	}

	// Calculate overall metrics
	cdv.calculateOverallMetrics(result)

	// Determine final status
	cdv.determineValidationStatus(result)

	cdv.logger.Info("Compliance validation completed",
		zap.String("validation_id", result.ID),
		zap.Float64("accuracy", result.Accuracy),
		zap.Float64("compliance", result.Compliance),
		zap.String("status", string(result.Status)))

	return result, nil
}
