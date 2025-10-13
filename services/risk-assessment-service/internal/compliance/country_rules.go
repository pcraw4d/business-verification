package compliance

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// CountryRules implements country-specific compliance rules and risk factors
type CountryRules struct {
	logger *zap.Logger
	config *CountryRulesConfig
}

// CountryRulesConfig represents configuration for country-specific rules
type CountryRulesConfig struct {
	SupportedCountries []string                  `json:"supported_countries"`
	DefaultCountry     string                    `json:"default_country"`
	CountryConfigs     map[string]*CountryConfig `json:"country_configs"`
	EnableLocalization bool                      `json:"enable_localization"`
	EnableCompliance   bool                      `json:"enable_compliance"`
}

// CountryConfig represents configuration for a specific country
type CountryConfig struct {
	Code               string                 `json:"code"`
	Name               string                 `json:"name"`
	Region             string                 `json:"region"`
	Currency           string                 `json:"currency"`
	Language           string                 `json:"language"`
	Timezone           string                 `json:"timezone"`
	RiskFactors        []RiskFactor           `json:"risk_factors"`
	ComplianceRules    []ComplianceRule       `json:"compliance_rules"`
	ValidationRules    []ValidationRule       `json:"validation_rules"`
	SanctionsLists     []string               `json:"sanctions_lists"`
	RegulatoryBodies   []RegulatoryBody       `json:"regulatory_bodies"`
	DataResidencyRules DataResidencyRules     `json:"data_residency_rules"`
	BusinessTypes      []BusinessType         `json:"business_types"`
	DocumentTypes      []DocumentType         `json:"document_types"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// RiskFactor represents a country-specific risk factor
type RiskFactor struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Category      RiskCategory           `json:"category"`
	Severity      RiskSeverity           `json:"severity"`
	Weight        float64                `json:"weight"`
	IsActive      bool                   `json:"is_active"`
	LocalizedName map[string]string      `json:"localized_name"`
	LocalizedDesc map[string]string      `json:"localized_desc"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// RiskCategory represents the category of a risk factor
type RiskCategory string

const (
	RiskCategoryPolitical    RiskCategory = "political"
	RiskCategoryEconomic     RiskCategory = "economic"
	RiskCategoryRegulatory   RiskCategory = "regulatory"
	RiskCategoryOperational  RiskCategory = "operational"
	RiskCategoryReputational RiskCategory = "reputational"
	RiskCategoryCompliance   RiskCategory = "compliance"
	RiskCategoryGeographic   RiskCategory = "geographic"
	RiskCategoryIndustry     RiskCategory = "industry"
)

// RiskSeverity represents the severity of a risk factor
type RiskSeverity string

const (
	RiskSeverityLow      RiskSeverity = "low"
	RiskSeverityMedium   RiskSeverity = "medium"
	RiskSeverityHigh     RiskSeverity = "high"
	RiskSeverityCritical RiskSeverity = "critical"
)

// ComplianceRule represents a country-specific compliance rule
type ComplianceRule struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Type           ComplianceRuleType     `json:"type"`
	Category       ComplianceCategory     `json:"category"`
	IsMandatory    bool                   `json:"is_mandatory"`
	EffectiveDate  time.Time              `json:"effective_date"`
	ExpiryDate     *time.Time             `json:"expiry_date,omitempty"`
	RegulatoryBody string                 `json:"regulatory_body"`
	Penalty        string                 `json:"penalty,omitempty"`
	Requirements   []string               `json:"requirements"`
	LocalizedName  map[string]string      `json:"localized_name"`
	LocalizedDesc  map[string]string      `json:"localized_desc"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// ComplianceRuleType represents the type of compliance rule
type ComplianceRuleType string

const (
	ComplianceRuleTypeRegistration ComplianceRuleType = "registration"
	ComplianceRuleTypeReporting    ComplianceRuleType = "reporting"
	ComplianceRuleTypeDisclosure   ComplianceRuleType = "disclosure"
	ComplianceRuleTypeVerification ComplianceRuleType = "verification"
	ComplianceRuleTypeMonitoring   ComplianceRuleType = "monitoring"
	ComplianceRuleTypeAudit        ComplianceRuleType = "audit"
)

// ComplianceCategory represents the category of compliance
type ComplianceCategory string

const (
	ComplianceCategoryAML       ComplianceCategory = "aml"
	ComplianceCategoryKYC       ComplianceCategory = "kyc"
	ComplianceCategoryKYB       ComplianceCategory = "kyb"
	ComplianceCategorySanctions ComplianceCategory = "sanctions"
	ComplianceCategoryTax       ComplianceCategory = "tax"
	ComplianceCategoryPrivacy   ComplianceCategory = "privacy"
	ComplianceCategoryData      ComplianceCategory = "data"
)

// ValidationRule represents a country-specific validation rule
type ValidationRule struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Field          string                 `json:"field"`
	Type           ValidationType         `json:"type"`
	Pattern        string                 `json:"pattern,omitempty"`
	MinLength      int                    `json:"min_length,omitempty"`
	MaxLength      int                    `json:"max_length,omitempty"`
	Required       bool                   `json:"required"`
	ErrorMessage   string                 `json:"error_message"`
	LocalizedError map[string]string      `json:"localized_error"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// ValidationType represents the type of validation
type ValidationType string

const (
	ValidationTypeRegex    ValidationType = "regex"
	ValidationTypeLength   ValidationType = "length"
	ValidationTypeFormat   ValidationType = "format"
	ValidationTypeChecksum ValidationType = "checksum"
	ValidationTypeCustom   ValidationType = "custom"
)

// RegulatoryBody represents a regulatory body for a country
type RegulatoryBody struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Acronym          string                 `json:"acronym"`
	Type             RegulatoryBodyType     `json:"type"`
	Jurisdiction     string                 `json:"jurisdiction"`
	Website          string                 `json:"website,omitempty"`
	ContactInfo      ContactInfo            `json:"contact_info,omitempty"`
	Responsibilities []string               `json:"responsibilities"`
	LocalizedName    map[string]string      `json:"localized_name"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// RegulatoryBodyType represents the type of regulatory body
type RegulatoryBodyType string

const (
	RegulatoryBodyTypeCentralBank    RegulatoryBodyType = "central_bank"
	RegulatoryBodyTypeFinancial      RegulatoryBodyType = "financial"
	RegulatoryBodyTypeSecurities     RegulatoryBodyType = "securities"
	RegulatoryBodyTypeInsurance      RegulatoryBodyType = "insurance"
	RegulatoryBodyTypeTax            RegulatoryBodyType = "tax"
	RegulatoryBodyTypeDataProtection RegulatoryBodyType = "data_protection"
	RegulatoryBodyTypeCompetition    RegulatoryBodyType = "competition"
	RegulatoryBodyTypeOther          RegulatoryBodyType = "other"
)

// ContactInfo represents contact information
type ContactInfo struct {
	Address string `json:"address,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Email   string `json:"email,omitempty"`
	Website string `json:"website,omitempty"`
}

// DataResidencyRules represents data residency rules for a country
type DataResidencyRules struct {
	RequiresLocalStorage bool     `json:"requires_local_storage"`
	AllowedRegions       []string `json:"allowed_regions"`
	RestrictedRegions    []string `json:"restricted_regions"`
	CrossBorderTransfer  bool     `json:"cross_border_transfer"`
	TransferRequirements []string `json:"transfer_requirements"`
	RetentionPeriod      int      `json:"retention_period_days"`
	DeletionRequirements []string `json:"deletion_requirements"`
}

// BusinessType represents a business type for a country
type BusinessType struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Code          string                 `json:"code"`
	Category      string                 `json:"category"`
	RiskLevel     RiskSeverity           `json:"risk_level"`
	Requirements  []string               `json:"requirements"`
	LocalizedName map[string]string      `json:"localized_name"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// DocumentType represents a document type for a country
type DocumentType struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Code           string                 `json:"code"`
	Category       string                 `json:"category"`
	Required       bool                   `json:"required"`
	ValidityPeriod int                    `json:"validity_period_days"`
	Format         string                 `json:"format"`
	LocalizedName  map[string]string      `json:"localized_name"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// NewCountryRules creates a new country rules instance
func NewCountryRules(config *CountryRulesConfig, logger *zap.Logger) *CountryRules {
	return &CountryRules{
		logger: logger,
		config: config,
	}
}

// GetSupportedCountries returns the list of supported countries
func (cr *CountryRules) GetSupportedCountries() []string {
	return cr.config.SupportedCountries
}

// GetCountryConfig returns the configuration for a specific country
func (cr *CountryRules) GetCountryConfig(countryCode string) (*CountryConfig, error) {
	config, exists := cr.config.CountryConfigs[strings.ToUpper(countryCode)]
	if !exists {
		return nil, fmt.Errorf("country configuration not found for code: %s", countryCode)
	}
	return config, nil
}

// GetRiskFactors returns risk factors for a specific country
func (cr *CountryRules) GetRiskFactors(ctx context.Context, countryCode string, language string) ([]*RiskFactor, error) {
	config, err := cr.GetCountryConfig(countryCode)
	if err != nil {
		return nil, err
	}

	riskFactors := make([]*RiskFactor, 0)
	for _, rf := range config.RiskFactors {
		if rf.IsActive {
			// Localize risk factor if language is specified
			if language != "" && cr.config.EnableLocalization {
				rf = cr.localizeRiskFactor(rf, language)
			}
			riskFactors = append(riskFactors, &rf)
		}
	}

	cr.logger.Info("Retrieved risk factors for country",
		zap.String("country_code", countryCode),
		zap.String("language", language),
		zap.Int("count", len(riskFactors)))

	return riskFactors, nil
}

// GetComplianceRules returns compliance rules for a specific country
func (cr *CountryRules) GetComplianceRules(ctx context.Context, countryCode string, category ComplianceCategory, language string) ([]*ComplianceRule, error) {
	config, err := cr.GetCountryConfig(countryCode)
	if err != nil {
		return nil, err
	}

	complianceRules := make([]*ComplianceRule, 0)
	for _, rule := range config.ComplianceRules {
		// Filter by category if specified
		if category != "" && rule.Category != category {
			continue
		}

		// Check if rule is currently effective
		now := time.Now()
		if rule.EffectiveDate.After(now) {
			continue
		}
		if rule.ExpiryDate != nil && rule.ExpiryDate.Before(now) {
			continue
		}

		// Localize compliance rule if language is specified
		if language != "" && cr.config.EnableLocalization {
			rule = cr.localizeComplianceRule(rule, language)
		}

		complianceRules = append(complianceRules, &rule)
	}

	cr.logger.Info("Retrieved compliance rules for country",
		zap.String("country_code", countryCode),
		zap.String("category", string(category)),
		zap.String("language", language),
		zap.Int("count", len(complianceRules)))

	return complianceRules, nil
}

// ValidateBusinessData validates business data against country-specific rules
func (cr *CountryRules) ValidateBusinessData(ctx context.Context, countryCode string, data map[string]interface{}) error {
	config, err := cr.GetCountryConfig(countryCode)
	if err != nil {
		return err
	}

	var validationErrors []string

	for _, rule := range config.ValidationRules {
		value, exists := data[rule.Field]
		if !exists {
			if rule.Required {
				validationErrors = append(validationErrors, fmt.Sprintf("required field '%s' is missing", rule.Field))
			}
			continue
		}

		if err := cr.validateField(value, rule); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("validation failed for field '%s': %s", rule.Field, err.Error()))
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
	}

	cr.logger.Info("Business data validation completed",
		zap.String("country_code", countryCode),
		zap.Int("fields_validated", len(data)))

	return nil
}

// GetRegulatoryBodies returns regulatory bodies for a specific country
func (cr *CountryRules) GetRegulatoryBodies(ctx context.Context, countryCode string, bodyType RegulatoryBodyType, language string) ([]*RegulatoryBody, error) {
	config, err := cr.GetCountryConfig(countryCode)
	if err != nil {
		return nil, err
	}

	regulatoryBodies := make([]*RegulatoryBody, 0)
	for _, body := range config.RegulatoryBodies {
		// Filter by type if specified
		if bodyType != "" && body.Type != bodyType {
			continue
		}

		// Localize regulatory body if language is specified
		if language != "" && cr.config.EnableLocalization {
			body = cr.localizeRegulatoryBody(body, language)
		}

		regulatoryBodies = append(regulatoryBodies, &body)
	}

	cr.logger.Info("Retrieved regulatory bodies for country",
		zap.String("country_code", countryCode),
		zap.String("body_type", string(bodyType)),
		zap.String("language", language),
		zap.Int("count", len(regulatoryBodies)))

	return regulatoryBodies, nil
}

// GetBusinessTypes returns business types for a specific country
func (cr *CountryRules) GetBusinessTypes(ctx context.Context, countryCode string, language string) ([]*BusinessType, error) {
	config, err := cr.GetCountryConfig(countryCode)
	if err != nil {
		return nil, err
	}

	businessTypes := make([]*BusinessType, 0)
	for _, bt := range config.BusinessTypes {
		// Localize business type if language is specified
		if language != "" && cr.config.EnableLocalization {
			bt = cr.localizeBusinessType(bt, language)
		}
		businessTypes = append(businessTypes, &bt)
	}

	cr.logger.Info("Retrieved business types for country",
		zap.String("country_code", countryCode),
		zap.String("language", language),
		zap.Int("count", len(businessTypes)))

	return businessTypes, nil
}

// GetDocumentTypes returns document types for a specific country
func (cr *CountryRules) GetDocumentTypes(ctx context.Context, countryCode string, language string) ([]*DocumentType, error) {
	config, err := cr.GetCountryConfig(countryCode)
	if err != nil {
		return nil, err
	}

	documentTypes := make([]*DocumentType, 0)
	for _, dt := range config.DocumentTypes {
		// Localize document type if language is specified
		if language != "" && cr.config.EnableLocalization {
			dt = cr.localizeDocumentType(dt, language)
		}
		documentTypes = append(documentTypes, &dt)
	}

	cr.logger.Info("Retrieved document types for country",
		zap.String("country_code", countryCode),
		zap.String("language", language),
		zap.Int("count", len(documentTypes)))

	return documentTypes, nil
}

// CheckDataResidencyCompliance checks if data residency rules are met
func (cr *CountryRules) CheckDataResidencyCompliance(ctx context.Context, countryCode string, dataLocation string) (bool, error) {
	config, err := cr.GetCountryConfig(countryCode)
	if err != nil {
		return false, err
	}

	rules := config.DataResidencyRules

	// Check if local storage is required
	if rules.RequiresLocalStorage {
		// In a real implementation, this would check if data is stored locally
		// For now, we'll assume compliance if data location is provided
		if dataLocation == "" {
			return false, fmt.Errorf("local storage required but data location not specified")
		}
	}

	// Check allowed regions
	if len(rules.AllowedRegions) > 0 {
		allowed := false
		for _, region := range rules.AllowedRegions {
			if strings.Contains(strings.ToLower(dataLocation), strings.ToLower(region)) {
				allowed = true
				break
			}
		}
		if !allowed {
			return false, fmt.Errorf("data location '%s' not in allowed regions", dataLocation)
		}
	}

	// Check restricted regions
	if len(rules.RestrictedRegions) > 0 {
		for _, region := range rules.RestrictedRegions {
			if strings.Contains(strings.ToLower(dataLocation), strings.ToLower(region)) {
				return false, fmt.Errorf("data location '%s' is in restricted region: %s", dataLocation, region)
			}
		}
	}

	cr.logger.Info("Data residency compliance checked",
		zap.String("country_code", countryCode),
		zap.String("data_location", dataLocation),
		zap.Bool("compliant", true))

	return true, nil
}

// CalculateCountryRiskScore calculates a risk score based on country-specific factors
func (cr *CountryRules) CalculateCountryRiskScore(ctx context.Context, countryCode string, businessData map[string]interface{}) (float64, error) {
	riskFactors, err := cr.GetRiskFactors(ctx, countryCode, "")
	if err != nil {
		return 0, err
	}

	totalScore := 0.0
	totalWeight := 0.0

	for _, rf := range riskFactors {
		// Calculate risk score based on severity and weight
		severityScore := cr.getSeverityScore(rf.Severity)
		weightedScore := severityScore * rf.Weight

		totalScore += weightedScore
		totalWeight += rf.Weight
	}

	if totalWeight == 0 {
		return 0.0, nil
	}

	// Normalize score to 0-100 range
	normalizedScore := (totalScore / totalWeight) * 100

	cr.logger.Info("Country risk score calculated",
		zap.String("country_code", countryCode),
		zap.Float64("score", normalizedScore),
		zap.Int("risk_factors", len(riskFactors)))

	return normalizedScore, nil
}

// validateField validates a field against a validation rule
func (cr *CountryRules) validateField(value interface{}, rule ValidationRule) error {
	switch rule.Type {
	case ValidationTypeRegex:
		return cr.validateRegex(value, rule)
	case ValidationTypeLength:
		return cr.validateLength(value, rule)
	case ValidationTypeFormat:
		return cr.validateFormat(value, rule)
	case ValidationTypeChecksum:
		return cr.validateChecksum(value, rule)
	case ValidationTypeCustom:
		return cr.validateCustom(value, rule)
	default:
		return fmt.Errorf("unknown validation type: %s", rule.Type)
	}
}

// validateRegex validates a field using regex pattern
func (cr *CountryRules) validateRegex(value interface{}, rule ValidationRule) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string for regex validation")
	}

	matched, err := regexp.MatchString(rule.Pattern, str)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %s", err.Error())
	}

	if !matched {
		return fmt.Errorf("value does not match pattern: %s", rule.Pattern)
	}

	return nil
}

// validateLength validates a field's length
func (cr *CountryRules) validateLength(value interface{}, rule ValidationRule) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string for length validation")
	}

	length := len(str)
	if rule.MinLength > 0 && length < rule.MinLength {
		return fmt.Errorf("length %d is less than minimum %d", length, rule.MinLength)
	}

	if rule.MaxLength > 0 && length > rule.MaxLength {
		return fmt.Errorf("length %d exceeds maximum %d", length, rule.MaxLength)
	}

	return nil
}

// validateFormat validates a field's format
func (cr *CountryRules) validateFormat(value interface{}, rule ValidationRule) error {
	// In a real implementation, this would validate specific formats
	// For now, we'll do basic string validation
	_, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string for format validation")
	}

	return nil
}

// validateChecksum validates a field's checksum
func (cr *CountryRules) validateChecksum(value interface{}, rule ValidationRule) error {
	// In a real implementation, this would validate checksums
	// For now, we'll return success
	return nil
}

// validateCustom validates a field using custom logic
func (cr *CountryRules) validateCustom(value interface{}, rule ValidationRule) error {
	// In a real implementation, this would execute custom validation logic
	// For now, we'll return success
	return nil
}

// getSeverityScore returns a numeric score for risk severity
func (cr *CountryRules) getSeverityScore(severity RiskSeverity) float64 {
	switch severity {
	case RiskSeverityLow:
		return 1.0
	case RiskSeverityMedium:
		return 2.0
	case RiskSeverityHigh:
		return 3.0
	case RiskSeverityCritical:
		return 4.0
	default:
		return 0.0
	}
}

// Localization helper functions
func (cr *CountryRules) localizeRiskFactor(rf RiskFactor, language string) RiskFactor {
	if localizedName, exists := rf.LocalizedName[language]; exists {
		rf.Name = localizedName
	}
	if localizedDesc, exists := rf.LocalizedDesc[language]; exists {
		rf.Description = localizedDesc
	}
	return rf
}

func (cr *CountryRules) localizeComplianceRule(rule ComplianceRule, language string) ComplianceRule {
	if localizedName, exists := rule.LocalizedName[language]; exists {
		rule.Name = localizedName
	}
	if localizedDesc, exists := rule.LocalizedDesc[language]; exists {
		rule.Description = localizedDesc
	}
	return rule
}

func (cr *CountryRules) localizeRegulatoryBody(body RegulatoryBody, language string) RegulatoryBody {
	if localizedName, exists := body.LocalizedName[language]; exists {
		body.Name = localizedName
	}
	return body
}

func (cr *CountryRules) localizeBusinessType(bt BusinessType, language string) BusinessType {
	if localizedName, exists := bt.LocalizedName[language]; exists {
		bt.Name = localizedName
	}
	return bt
}

func (cr *CountryRules) localizeDocumentType(dt DocumentType, language string) DocumentType {
	if localizedName, exists := dt.LocalizedName[language]; exists {
		dt.Name = localizedName
	}
	return dt
}
