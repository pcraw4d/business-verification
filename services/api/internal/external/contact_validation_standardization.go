package external

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ContactValidationStandardizer handles validation and standardization of contact information
type ContactValidationStandardizer struct {
	config *ContactValidationConfig
	logger *zap.Logger
}

// ContactValidationConfig contains configuration for validation and standardization
type ContactValidationConfig struct {
	// Phone validation settings
	EnablePhoneValidation bool     `json:"enable_phone_validation"`
	EnableE164Format      bool     `json:"enable_e164_format"`
	AllowedCountryCodes   []string `json:"allowed_country_codes"`
	DefaultCountryCode    string   `json:"default_country_code"`

	// Email validation settings
	EnableEmailValidation  bool     `json:"enable_email_validation"`
	EnableDomainValidation bool     `json:"enable_domain_validation"`
	EnableMXValidation     bool     `json:"enable_mx_validation"`
	BlockedDomains         []string `json:"blocked_domains"`
	TrustedDomains         []string `json:"trusted_domains"`

	// Address validation settings
	EnableAddressValidation    bool     `json:"enable_address_validation"`
	EnableGeocoding            bool     `json:"enable_geocoding"`
	EnablePostalCodeValidation bool     `json:"enable_postal_code_validation"`
	SupportedCountries         []string `json:"supported_countries"`

	// Standardization settings
	EnablePhoneStandardization   bool `json:"enable_phone_standardization"`
	EnableEmailStandardization   bool `json:"enable_email_standardization"`
	EnableAddressStandardization bool `json:"enable_address_standardization"`

	// Quality settings
	MinValidationConfidence float64 `json:"min_validation_confidence"`
	EnableFuzzyMatching     bool    `json:"enable_fuzzy_matching"`
	EnableAutoCorrection    bool    `json:"enable_auto_correction"`

	// Performance settings
	ValidationTimeout time.Duration `json:"validation_timeout"`
	MaxBatchSize      int           `json:"max_batch_size"`
	EnableCaching     bool          `json:"enable_caching"`
}

// ValidationResult represents the result of contact validation
type ValidationResult struct {
	IsValid            bool                  `json:"is_valid"`
	ValidationScore    float64               `json:"validation_score"`
	StandardizedValue  string                `json:"standardized_value"`
	OriginalValue      string                `json:"original_value"`
	ValidationErrors   []ValidationError     `json:"validation_errors"`
	ValidationWarnings []ValidationWarning   `json:"validation_warnings"`
	QualityMetrics     ContactQualityMetrics `json:"quality_metrics"`
	GeographicInfo     GeographicInfo        `json:"geographic_info"`
	TechnicalInfo      TechnicalInfo         `json:"technical_info"`
	ValidatedAt        time.Time             `json:"validated_at"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Code       string    `json:"code"`
	Message    string    `json:"message"`
	Field      string    `json:"field"`
	Severity   string    `json:"severity"` // error, warning, info
	Suggestion string    `json:"suggestion"`
	DetectedAt time.Time `json:"detected_at"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Code       string    `json:"code"`
	Message    string    `json:"message"`
	Field      string    `json:"field"`
	Suggestion string    `json:"suggestion"`
	DetectedAt time.Time `json:"detected_at"`
}

// ContactQualityMetrics represents quality metrics for contact information
type ContactQualityMetrics struct {
	FormatCompliance float64 `json:"format_compliance"`
	DataCompleteness float64 `json:"data_completeness"`
	Accuracy         float64 `json:"accuracy"`
	Deliverability   float64 `json:"deliverability"`
	TrustScore       float64 `json:"trust_score"`
	OverallQuality   float64 `json:"overall_quality"`
}

// GeographicInfo represents geographic information
type GeographicInfo struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	PostalCode  string  `json:"postal_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
}

// TechnicalInfo represents technical information
type TechnicalInfo struct {
	Format   string            `json:"format"`
	Protocol string            `json:"protocol"`
	Provider string            `json:"provider"`
	Carrier  string            `json:"carrier"`
	LineType string            `json:"line_type"`
	Metadata map[string]string `json:"metadata"`
}

// BatchValidationResult represents results for batch validation
type BatchValidationResult struct {
	Results         []ValidationResult `json:"results"`
	TotalProcessed  int                `json:"total_processed"`
	TotalValid      int                `json:"total_valid"`
	TotalInvalid    int                `json:"total_invalid"`
	ValidationTime  time.Duration      `json:"validation_time"`
	ErrorRate       float64            `json:"error_rate"`
	ProcessingStats ProcessingStats    `json:"processing_stats"`
}

// ProcessingStats represents processing statistics
type ProcessingStats struct {
	PhoneValidated    int `json:"phone_validated"`
	EmailValidated    int `json:"email_validated"`
	AddressValidated  int `json:"address_validated"`
	StandardizedCount int `json:"standardized_count"`
	CorrectedCount    int `json:"corrected_count"`
}

// NewContactValidationStandardizer creates a new contact validation standardizer
func NewContactValidationStandardizer(config *ContactValidationConfig, logger *zap.Logger) *ContactValidationStandardizer {
	if config == nil {
		config = getDefaultContactValidationConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &ContactValidationStandardizer{
		config: config,
		logger: logger,
	}
}

// ValidatePhoneNumber validates and standardizes a phone number
func (cvs *ContactValidationStandardizer) ValidatePhoneNumber(ctx context.Context, phoneNumber string) (*ValidationResult, error) {
	startTime := time.Now()

	result := &ValidationResult{
		OriginalValue:  phoneNumber,
		ValidatedAt:    time.Now(),
		QualityMetrics: ContactQualityMetrics{},
		GeographicInfo: GeographicInfo{},
		TechnicalInfo:  TechnicalInfo{},
	}

	// Check context timeout
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("validation cancelled: %w", ctx.Err())
	default:
	}

	// Basic format validation
	if !cvs.config.EnablePhoneValidation {
		result.IsValid = true
		result.StandardizedValue = phoneNumber
		result.ValidationScore = 1.0
		return result, nil
	}

	// Clean and normalize phone number
	cleanedPhone := cvs.cleanPhoneNumber(phoneNumber)
	if cleanedPhone == "" {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors, ValidationError{
			Code:       "EMPTY_PHONE",
			Message:    "Phone number is empty after cleaning",
			Field:      "phone",
			Severity:   "error",
			DetectedAt: time.Now(),
		})
		return result, nil
	}

	// Validate phone format
	formatValid, formatScore := cvs.validatePhoneFormat(cleanedPhone)
	result.QualityMetrics.FormatCompliance = formatScore

	if !formatValid {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors, ValidationError{
			Code:       "INVALID_FORMAT",
			Message:    "Phone number format is invalid",
			Field:      "phone",
			Severity:   "error",
			Suggestion: "Please provide a valid phone number with country code",
			DetectedAt: time.Now(),
		})
	}

	// Standardize to E.164 format if enabled
	if cvs.config.EnableE164Format {
		standardized := cvs.standardizePhoneToE164(cleanedPhone)
		result.StandardizedValue = standardized
	} else {
		result.StandardizedValue = cleanedPhone
	}

	// Extract geographic information
	countryCode := cvs.extractCountryCode(cleanedPhone)
	if countryCode != "" {
		result.GeographicInfo.CountryCode = countryCode
		result.GeographicInfo.Country = cvs.getCountryFromCode(countryCode)
	}

	// Validate against allowed country codes
	if len(cvs.config.AllowedCountryCodes) > 0 {
		allowed := false
		for _, code := range cvs.config.AllowedCountryCodes {
			if code == countryCode {
				allowed = true
				break
			}
		}
		if !allowed {
			result.ValidationWarnings = append(result.ValidationWarnings, ValidationWarning{
				Code:       "COUNTRY_NOT_ALLOWED",
				Message:    fmt.Sprintf("Country code %s is not in allowed list", countryCode),
				Field:      "phone",
				Suggestion: fmt.Sprintf("Use phone numbers from allowed countries: %v", cvs.config.AllowedCountryCodes),
				DetectedAt: time.Now(),
			})
		}
	}

	// Calculate technical info
	result.TechnicalInfo.Format = cvs.detectPhoneFormat(cleanedPhone)
	result.TechnicalInfo.LineType = cvs.detectLineType(cleanedPhone)
	result.TechnicalInfo.Provider = cvs.detectProvider(cleanedPhone)

	// Calculate quality metrics
	result.QualityMetrics.DataCompleteness = cvs.calculatePhoneCompleteness(cleanedPhone)
	result.QualityMetrics.Accuracy = cvs.calculatePhoneAccuracy(cleanedPhone)
	result.QualityMetrics.TrustScore = cvs.calculatePhoneTrustScore(cleanedPhone)
	result.QualityMetrics.OverallQuality = cvs.calculateOverallPhoneQuality(result.QualityMetrics)

	// Final validation score
	result.ValidationScore = result.QualityMetrics.OverallQuality
	result.IsValid = result.ValidationScore >= cvs.config.MinValidationConfidence && formatValid

	cvs.logger.Debug("phone validation completed",
		zap.String("original", phoneNumber),
		zap.String("standardized", result.StandardizedValue),
		zap.Bool("valid", result.IsValid),
		zap.Float64("score", result.ValidationScore),
		zap.Duration("duration", time.Since(startTime)))

	return result, nil
}

// ValidateEmailAddress validates and standardizes an email address
func (cvs *ContactValidationStandardizer) ValidateEmailAddress(ctx context.Context, emailAddress string) (*ValidationResult, error) {
	startTime := time.Now()

	result := &ValidationResult{
		OriginalValue:  emailAddress,
		ValidatedAt:    time.Now(),
		QualityMetrics: ContactQualityMetrics{},
		GeographicInfo: GeographicInfo{},
		TechnicalInfo:  TechnicalInfo{},
	}

	// Check context timeout
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("validation cancelled: %w", ctx.Err())
	default:
	}

	// Basic format validation
	if !cvs.config.EnableEmailValidation {
		result.IsValid = true
		result.StandardizedValue = emailAddress
		result.ValidationScore = 1.0
		return result, nil
	}

	// Clean and normalize email
	cleanedEmail := cvs.cleanEmailAddress(emailAddress)
	if cleanedEmail == "" {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors, ValidationError{
			Code:       "EMPTY_EMAIL",
			Message:    "Email address is empty after cleaning",
			Field:      "email",
			Severity:   "error",
			DetectedAt: time.Now(),
		})
		return result, nil
	}

	// Validate email format using regex
	formatValid, formatScore := cvs.validateEmailFormat(cleanedEmail)
	result.QualityMetrics.FormatCompliance = formatScore

	if !formatValid {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors, ValidationError{
			Code:       "INVALID_FORMAT",
			Message:    "Email address format is invalid",
			Field:      "email",
			Severity:   "error",
			Suggestion: "Please provide a valid email address (e.g., user@domain.com)",
			DetectedAt: time.Now(),
		})
		return result, nil
	}

	// Standardize email address
	if cvs.config.EnableEmailStandardization {
		result.StandardizedValue = cvs.standardizeEmail(cleanedEmail)
	} else {
		result.StandardizedValue = cleanedEmail
	}

	// Extract domain and validate
	domain := cvs.extractDomain(cleanedEmail)
	if domain == "" {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors, ValidationError{
			Code:       "MISSING_DOMAIN",
			Message:    "Email address is missing domain part",
			Field:      "email",
			Severity:   "error",
			DetectedAt: time.Now(),
		})
		return result, nil
	}

	// Check blocked domains
	if cvs.isDomainBlocked(domain) {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors, ValidationError{
			Code:       "BLOCKED_DOMAIN",
			Message:    fmt.Sprintf("Domain %s is blocked", domain),
			Field:      "email",
			Severity:   "error",
			DetectedAt: time.Now(),
		})
		return result, nil
	}

	// Domain validation
	if cvs.config.EnableDomainValidation {
		domainValid := cvs.validateDomain(ctx, domain)
		if !domainValid {
			result.ValidationWarnings = append(result.ValidationWarnings, ValidationWarning{
				Code:       "DOMAIN_INVALID",
				Message:    fmt.Sprintf("Domain %s appears to be invalid", domain),
				Field:      "email",
				Suggestion: "Please verify the domain name is correct",
				DetectedAt: time.Now(),
			})
		}
	}

	// MX record validation
	if cvs.config.EnableMXValidation {
		mxValid := cvs.validateMXRecord(ctx, domain)
		result.QualityMetrics.Deliverability = cvs.calculateDeliverabilityScore(mxValid, domain)

		if !mxValid {
			result.ValidationWarnings = append(result.ValidationWarnings, ValidationWarning{
				Code:       "NO_MX_RECORD",
				Message:    fmt.Sprintf("Domain %s has no MX record", domain),
				Field:      "email",
				Suggestion: "Email may not be deliverable - verify domain can receive emails",
				DetectedAt: time.Now(),
			})
		}
	}

	// Calculate technical info
	result.TechnicalInfo.Format = cvs.detectEmailFormat(cleanedEmail)
	result.TechnicalInfo.Provider = cvs.detectEmailProvider(domain)
	result.TechnicalInfo.Protocol = "SMTP"

	// Check if domain is trusted
	isTrusted := cvs.isDomainTrusted(domain)
	result.QualityMetrics.TrustScore = cvs.calculateEmailTrustScore(domain, isTrusted)

	// Calculate quality metrics
	result.QualityMetrics.DataCompleteness = cvs.calculateEmailCompleteness(cleanedEmail)
	result.QualityMetrics.Accuracy = cvs.calculateEmailAccuracy(cleanedEmail, formatValid)
	result.QualityMetrics.OverallQuality = cvs.calculateOverallEmailQuality(result.QualityMetrics)

	// Final validation score
	result.ValidationScore = result.QualityMetrics.OverallQuality
	result.IsValid = result.ValidationScore >= cvs.config.MinValidationConfidence && formatValid && !cvs.isDomainBlocked(domain)

	cvs.logger.Debug("email validation completed",
		zap.String("original", emailAddress),
		zap.String("standardized", result.StandardizedValue),
		zap.String("domain", domain),
		zap.Bool("valid", result.IsValid),
		zap.Float64("score", result.ValidationScore),
		zap.Duration("duration", time.Since(startTime)))

	return result, nil
}

// ValidatePhysicalAddress validates and standardizes a physical address
func (cvs *ContactValidationStandardizer) ValidatePhysicalAddress(ctx context.Context, address string) (*ValidationResult, error) {
	startTime := time.Now()

	result := &ValidationResult{
		OriginalValue:  address,
		ValidatedAt:    time.Now(),
		QualityMetrics: ContactQualityMetrics{},
		GeographicInfo: GeographicInfo{},
		TechnicalInfo:  TechnicalInfo{},
	}

	// Check context timeout
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("validation cancelled: %w", ctx.Err())
	default:
	}

	// Basic format validation
	if !cvs.config.EnableAddressValidation {
		result.IsValid = true
		result.StandardizedValue = address
		result.ValidationScore = 1.0
		return result, nil
	}

	// Clean and normalize address
	cleanedAddress := cvs.cleanAddress(address)
	if cleanedAddress == "" {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors, ValidationError{
			Code:       "EMPTY_ADDRESS",
			Message:    "Address is empty after cleaning",
			Field:      "address",
			Severity:   "error",
			DetectedAt: time.Now(),
		})
		return result, nil
	}

	// Parse address components
	addressComponents := cvs.parseAddressComponents(cleanedAddress)

	// Validate address format
	formatValid, formatScore := cvs.validateAddressFormat(cleanedAddress, addressComponents)
	result.QualityMetrics.FormatCompliance = formatScore

	// Standardize address if enabled
	if cvs.config.EnableAddressStandardization {
		result.StandardizedValue = cvs.standardizeAddress(cleanedAddress, addressComponents)
	} else {
		result.StandardizedValue = cleanedAddress
	}

	// Validate postal code if present
	if cvs.config.EnablePostalCodeValidation && addressComponents["postal_code"] != "" {
		postalValid := cvs.validatePostalCode(addressComponents["postal_code"], addressComponents["country"])
		if !postalValid {
			result.ValidationWarnings = append(result.ValidationWarnings, ValidationWarning{
				Code:       "INVALID_POSTAL_CODE",
				Message:    fmt.Sprintf("Postal code %s appears invalid", addressComponents["postal_code"]),
				Field:      "address",
				Suggestion: "Please verify the postal code is correct",
				DetectedAt: time.Now(),
			})
		}
	}

	// Extract geographic information
	result.GeographicInfo.Country = addressComponents["country"]
	result.GeographicInfo.Region = addressComponents["region"]
	result.GeographicInfo.City = addressComponents["city"]
	result.GeographicInfo.PostalCode = addressComponents["postal_code"]

	// Geocoding if enabled
	if cvs.config.EnableGeocoding {
		lat, lng := cvs.geocodeAddress(ctx, result.StandardizedValue)
		result.GeographicInfo.Latitude = lat
		result.GeographicInfo.Longitude = lng
	}

	// Calculate technical info
	result.TechnicalInfo.Format = cvs.detectAddressFormat(cleanedAddress)

	// Calculate quality metrics
	result.QualityMetrics.DataCompleteness = cvs.calculateAddressCompleteness(addressComponents)
	result.QualityMetrics.Accuracy = cvs.calculateAddressAccuracy(cleanedAddress, addressComponents)
	result.QualityMetrics.TrustScore = cvs.calculateAddressTrustScore(addressComponents)
	result.QualityMetrics.OverallQuality = cvs.calculateOverallAddressQuality(result.QualityMetrics)

	// Final validation score
	result.ValidationScore = result.QualityMetrics.OverallQuality
	result.IsValid = result.ValidationScore >= cvs.config.MinValidationConfidence && formatValid

	cvs.logger.Debug("address validation completed",
		zap.String("original", address),
		zap.String("standardized", result.StandardizedValue),
		zap.Bool("valid", result.IsValid),
		zap.Float64("score", result.ValidationScore),
		zap.Duration("duration", time.Since(startTime)))

	return result, nil
}

// ValidateBatch validates multiple contact items in batch
func (cvs *ContactValidationStandardizer) ValidateBatch(ctx context.Context, contacts []string, contactType string) (*BatchValidationResult, error) {
	startTime := time.Now()

	batchResult := &BatchValidationResult{
		Results:         make([]ValidationResult, 0, len(contacts)),
		TotalProcessed:  len(contacts),
		ProcessingStats: ProcessingStats{},
	}

	// Check batch size limit
	if len(contacts) > cvs.config.MaxBatchSize {
		return nil, fmt.Errorf("batch size %d exceeds maximum allowed %d", len(contacts), cvs.config.MaxBatchSize)
	}

	validCount := 0
	invalidCount := 0

	for _, contact := range contacts {
		// Check context timeout
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("batch validation cancelled: %w", ctx.Err())
		default:
		}

		var result *ValidationResult
		var err error

		switch strings.ToLower(contactType) {
		case "phone":
			result, err = cvs.ValidatePhoneNumber(ctx, contact)
			batchResult.ProcessingStats.PhoneValidated++
		case "email":
			result, err = cvs.ValidateEmailAddress(ctx, contact)
			batchResult.ProcessingStats.EmailValidated++
		case "address":
			result, err = cvs.ValidatePhysicalAddress(ctx, contact)
			batchResult.ProcessingStats.AddressValidated++
		default:
			return nil, fmt.Errorf("unsupported contact type: %s", contactType)
		}

		if err != nil {
			// Create error result
			result = &ValidationResult{
				OriginalValue: contact,
				IsValid:       false,
				ValidationErrors: []ValidationError{{
					Code:       "VALIDATION_ERROR",
					Message:    err.Error(),
					Field:      contactType,
					Severity:   "error",
					DetectedAt: time.Now(),
				}},
				ValidatedAt: time.Now(),
			}
		}

		batchResult.Results = append(batchResult.Results, *result)

		if result.IsValid {
			validCount++
		} else {
			invalidCount++
		}

		if result.StandardizedValue != result.OriginalValue {
			batchResult.ProcessingStats.StandardizedCount++
		}
	}

	batchResult.TotalValid = validCount
	batchResult.TotalInvalid = invalidCount
	batchResult.ValidationTime = time.Since(startTime)
	batchResult.ErrorRate = float64(invalidCount) / float64(len(contacts))

	cvs.logger.Info("batch validation completed",
		zap.Int("total", len(contacts)),
		zap.Int("valid", validCount),
		zap.Int("invalid", invalidCount),
		zap.Float64("error_rate", batchResult.ErrorRate),
		zap.Duration("duration", batchResult.ValidationTime))

	return batchResult, nil
}

// Helper methods for phone validation

func (cvs *ContactValidationStandardizer) cleanPhoneNumber(phone string) string {
	// Remove all non-digit characters except + and ()
	cleaned := regexp.MustCompile(`[^0-9+()\-\s]`).ReplaceAllString(phone, "")
	// Remove extra spaces and normalize
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
	return strings.TrimSpace(cleaned)
}

func (cvs *ContactValidationStandardizer) validatePhoneFormat(phone string) (bool, float64) {
	// Basic phone validation patterns
	patterns := []struct {
		regex *regexp.Regexp
		score float64
	}{
		{regexp.MustCompile(`^\+[1-9]\d{1,14}$`), 1.0},        // E.164 format
		{regexp.MustCompile(`^\+\d{1,3}\s\d{3,14}$`), 0.9},    // International with space
		{regexp.MustCompile(`^\(\d{3}\)\s\d{3}-\d{4}$`), 0.8}, // US format (xxx) xxx-xxxx
		{regexp.MustCompile(`^\d{3}-\d{3}-\d{4}$`), 0.7},      // US format xxx-xxx-xxxx
		{regexp.MustCompile(`^\d{10,15}$`), 0.6},              // Just digits
	}

	for _, pattern := range patterns {
		if pattern.regex.MatchString(phone) {
			return true, pattern.score
		}
	}

	return false, 0.0
}

func (cvs *ContactValidationStandardizer) standardizePhoneToE164(phone string) string {
	// Remove all non-digit characters except +
	digits := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")

	// If already starts with +, return as is if valid
	if strings.HasPrefix(digits, "+") {
		return digits
	}

	// Add default country code if configured
	if cvs.config.DefaultCountryCode != "" && !strings.HasPrefix(digits, "+") {
		return "+" + cvs.config.DefaultCountryCode + digits
	}

	// If starts with country code but no +, add it
	if len(digits) > 10 {
		return "+" + digits
	}

	return digits
}

func (cvs *ContactValidationStandardizer) extractCountryCode(phone string) string {
	// Extract country code from E.164 format
	if strings.HasPrefix(phone, "+") {
		// Common country code patterns
		codes := []string{"1", "44", "49", "33", "39", "34", "81", "86", "91"}
		for _, code := range codes {
			if strings.HasPrefix(phone[1:], code) {
				return code
			}
		}
	}
	return ""
}

func (cvs *ContactValidationStandardizer) getCountryFromCode(code string) string {
	countryMap := map[string]string{
		"1":  "US",
		"44": "GB",
		"49": "DE",
		"33": "FR",
		"39": "IT",
		"34": "ES",
		"81": "JP",
		"86": "CN",
		"91": "IN",
	}
	return countryMap[code]
}

func (cvs *ContactValidationStandardizer) detectPhoneFormat(phone string) string {
	if regexp.MustCompile(`^\+[1-9]\d{1,14}$`).MatchString(phone) {
		return "E.164"
	}
	if regexp.MustCompile(`^\(\d{3}\)\s\d{3}-\d{4}$`).MatchString(phone) {
		return "US_STANDARD"
	}
	if regexp.MustCompile(`^\d{3}-\d{3}-\d{4}$`).MatchString(phone) {
		return "US_HYPHENATED"
	}
	return "UNKNOWN"
}

func (cvs *ContactValidationStandardizer) detectLineType(phone string) string {
	// Simple heuristics for line type detection
	if strings.Contains(phone, "800") || strings.Contains(phone, "888") || strings.Contains(phone, "877") {
		return "TOLL_FREE"
	}
	return "STANDARD"
}

func (cvs *ContactValidationStandardizer) detectProvider(phone string) string {
	// This would typically involve lookup against carrier databases
	// For now, return unknown
	return "UNKNOWN"
}

func (cvs *ContactValidationStandardizer) calculatePhoneCompleteness(phone string) float64 {
	score := 0.0
	if phone != "" {
		score += 0.5
	}
	if len(phone) >= 10 {
		score += 0.3
	}
	if strings.HasPrefix(phone, "+") {
		score += 0.2
	}
	return score
}

func (cvs *ContactValidationStandardizer) calculatePhoneAccuracy(phone string) float64 {
	// Based on format validation
	_, score := cvs.validatePhoneFormat(phone)
	return score
}

func (cvs *ContactValidationStandardizer) calculatePhoneTrustScore(phone string) float64 {
	// Higher trust for numbers with country codes and proper formatting
	if strings.HasPrefix(phone, "+") {
		return 0.9
	}
	if len(phone) >= 10 {
		return 0.7
	}
	return 0.5
}

func (cvs *ContactValidationStandardizer) calculateOverallPhoneQuality(metrics ContactQualityMetrics) float64 {
	return (metrics.FormatCompliance*0.3 + metrics.DataCompleteness*0.2 +
		metrics.Accuracy*0.3 + metrics.TrustScore*0.2)
}

// Helper methods for email validation

func (cvs *ContactValidationStandardizer) cleanEmailAddress(email string) string {
	// Convert to lowercase and trim spaces
	cleaned := strings.ToLower(strings.TrimSpace(email))
	return cleaned
}

func (cvs *ContactValidationStandardizer) validateEmailFormat(email string) (bool, float64) {
	// RFC 5322 compliant email regex (simplified)
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(pattern)

	if regex.MatchString(email) {
		// Additional checks for better score
		score := 0.8

		// Bonus for common TLDs
		if strings.HasSuffix(email, ".com") || strings.HasSuffix(email, ".org") || strings.HasSuffix(email, ".net") {
			score += 0.1
		}

		// Penalty for suspicious patterns
		if strings.Contains(email, "..") || strings.Contains(email, "--") {
			score -= 0.2
		}

		return true, score
	}

	return false, 0.0
}

func (cvs *ContactValidationStandardizer) standardizeEmail(email string) string {
	// Convert to lowercase
	email = strings.ToLower(email)

	// Handle Gmail alias removal (optional)
	parts := strings.Split(email, "@")
	if len(parts) == 2 && parts[1] == "gmail.com" {
		localPart := strings.Split(parts[0], "+")[0]       // Remove + aliases
		localPart = strings.ReplaceAll(localPart, ".", "") // Remove dots
		email = localPart + "@gmail.com"
	}

	return email
}

func (cvs *ContactValidationStandardizer) extractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

func (cvs *ContactValidationStandardizer) isDomainBlocked(domain string) bool {
	for _, blocked := range cvs.config.BlockedDomains {
		if domain == blocked {
			return true
		}
	}
	return false
}

func (cvs *ContactValidationStandardizer) isDomainTrusted(domain string) bool {
	for _, trusted := range cvs.config.TrustedDomains {
		if domain == trusted {
			return true
		}
	}
	return false
}

func (cvs *ContactValidationStandardizer) validateDomain(ctx context.Context, domain string) bool {
	// Basic domain validation - check if it resolves
	_, err := net.LookupHost(domain)
	return err == nil
}

func (cvs *ContactValidationStandardizer) validateMXRecord(ctx context.Context, domain string) bool {
	// Check if domain has MX records
	_, err := net.LookupMX(domain)
	return err == nil
}

func (cvs *ContactValidationStandardizer) calculateDeliverabilityScore(mxValid bool, domain string) float64 {
	score := 0.0
	if mxValid {
		score += 0.6
	}
	if cvs.isDomainTrusted(domain) {
		score += 0.4
	} else {
		score += 0.2
	}
	return score
}

func (cvs *ContactValidationStandardizer) detectEmailFormat(email string) string {
	if regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(email) {
		return "RFC5322"
	}
	return "UNKNOWN"
}

func (cvs *ContactValidationStandardizer) detectEmailProvider(domain string) string {
	providers := map[string]string{
		"gmail.com":   "Google",
		"yahoo.com":   "Yahoo",
		"outlook.com": "Microsoft",
		"hotmail.com": "Microsoft",
		"icloud.com":  "Apple",
		"aol.com":     "AOL",
	}

	if provider, exists := providers[domain]; exists {
		return provider
	}
	return "UNKNOWN"
}

func (cvs *ContactValidationStandardizer) calculateEmailCompleteness(email string) float64 {
	score := 0.0
	if email != "" {
		score += 0.5
	}
	if strings.Contains(email, "@") {
		score += 0.3
	}
	if len(strings.Split(email, "@")) == 2 {
		score += 0.2
	}
	return score
}

func (cvs *ContactValidationStandardizer) calculateEmailAccuracy(email string, formatValid bool) float64 {
	if formatValid {
		return 0.9
	}
	return 0.1
}

func (cvs *ContactValidationStandardizer) calculateEmailTrustScore(domain string, isTrusted bool) float64 {
	if isTrusted {
		return 1.0
	}

	// Common email providers get higher trust
	commonProviders := []string{"gmail.com", "yahoo.com", "outlook.com", "hotmail.com"}
	for _, provider := range commonProviders {
		if domain == provider {
			return 0.8
		}
	}

	// Business domains get medium trust
	if !strings.Contains(domain, "temp") && !strings.Contains(domain, "disposable") {
		return 0.6
	}

	return 0.3
}

func (cvs *ContactValidationStandardizer) calculateOverallEmailQuality(metrics ContactQualityMetrics) float64 {
	return (metrics.FormatCompliance*0.3 + metrics.DataCompleteness*0.2 +
		metrics.Accuracy*0.2 + metrics.Deliverability*0.15 + metrics.TrustScore*0.15)
}

// Helper methods for address validation

func (cvs *ContactValidationStandardizer) cleanAddress(address string) string {
	// Basic address cleaning
	cleaned := strings.TrimSpace(address)
	// Remove extra spaces
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
	return cleaned
}

func (cvs *ContactValidationStandardizer) parseAddressComponents(address string) map[string]string {
	components := make(map[string]string)

	// Simple address parsing (would be more sophisticated in production)
	lines := strings.Split(address, ",")

	if len(lines) >= 1 {
		components["street"] = strings.TrimSpace(lines[0])
	}
	if len(lines) >= 2 {
		components["city"] = strings.TrimSpace(lines[1])
	}
	if len(lines) >= 3 {
		// Try to extract state and postal code
		lastPart := strings.TrimSpace(lines[len(lines)-1])
		parts := strings.Fields(lastPart)
		if len(parts) >= 2 {
			components["region"] = parts[0]
			components["postal_code"] = parts[1]
		}
	}

	// Extract country (if present)
	if len(lines) >= 4 {
		components["country"] = strings.TrimSpace(lines[len(lines)-1])
	} else {
		components["country"] = "US" // Default
	}

	return components
}

func (cvs *ContactValidationStandardizer) validateAddressFormat(address string, components map[string]string) (bool, float64) {
	score := 0.0

	// Check for required components
	if components["street"] != "" {
		score += 0.4
	}
	if components["city"] != "" {
		score += 0.3
	}
	if components["region"] != "" {
		score += 0.2
	}
	if components["postal_code"] != "" {
		score += 0.1
	}

	return score >= 0.5, score
}

func (cvs *ContactValidationStandardizer) standardizeAddress(address string, components map[string]string) string {
	// Basic address standardization
	var parts []string

	if components["street"] != "" {
		parts = append(parts, components["street"])
	}
	if components["city"] != "" {
		parts = append(parts, components["city"])
	}
	if components["region"] != "" && components["postal_code"] != "" {
		parts = append(parts, components["region"]+" "+components["postal_code"])
	}
	if components["country"] != "" && components["country"] != "US" {
		parts = append(parts, components["country"])
	}

	return strings.Join(parts, ", ")
}

func (cvs *ContactValidationStandardizer) validatePostalCode(postalCode, country string) bool {
	patterns := map[string]*regexp.Regexp{
		"US": regexp.MustCompile(`^\d{5}(-\d{4})?$`),
		"GB": regexp.MustCompile(`^[A-Z]{1,2}\d[A-Z\d]?\s?\d[A-Z]{2}$`),
		"CA": regexp.MustCompile(`^[A-Z]\d[A-Z]\s?\d[A-Z]\d$`),
		"DE": regexp.MustCompile(`^\d{5}$`),
		"FR": regexp.MustCompile(`^\d{5}$`),
	}

	if pattern, exists := patterns[country]; exists {
		return pattern.MatchString(postalCode)
	}

	// Default validation - just check if it contains digits
	return regexp.MustCompile(`\d`).MatchString(postalCode)
}

func (cvs *ContactValidationStandardizer) geocodeAddress(ctx context.Context, address string) (float64, float64) {
	// Placeholder for geocoding service integration
	// Would integrate with Google Maps, MapBox, or similar service
	return 0.0, 0.0
}

func (cvs *ContactValidationStandardizer) detectAddressFormat(address string) string {
	if strings.Contains(address, ",") {
		return "COMMA_SEPARATED"
	}
	if strings.Contains(address, "\n") {
		return "MULTI_LINE"
	}
	return "SINGLE_LINE"
}

func (cvs *ContactValidationStandardizer) calculateAddressCompleteness(components map[string]string) float64 {
	score := 0.0
	totalComponents := 4.0

	if components["street"] != "" {
		score += 1.0
	}
	if components["city"] != "" {
		score += 1.0
	}
	if components["region"] != "" {
		score += 1.0
	}
	if components["postal_code"] != "" {
		score += 1.0
	}

	return score / totalComponents
}

func (cvs *ContactValidationStandardizer) calculateAddressAccuracy(address string, components map[string]string) float64 {
	// Simple accuracy based on component parsing success
	if len(components) >= 3 {
		return 0.8
	}
	if len(components) >= 2 {
		return 0.6
	}
	return 0.4
}

func (cvs *ContactValidationStandardizer) calculateAddressTrustScore(components map[string]string) float64 {
	// Higher trust for complete addresses
	completeness := cvs.calculateAddressCompleteness(components)
	return completeness*0.8 + 0.2 // Base trust of 0.2
}

func (cvs *ContactValidationStandardizer) calculateOverallAddressQuality(metrics ContactQualityMetrics) float64 {
	return (metrics.FormatCompliance*0.3 + metrics.DataCompleteness*0.4 +
		metrics.Accuracy*0.2 + metrics.TrustScore*0.1)
}

// UpdateConfig updates the validation configuration
func (cvs *ContactValidationStandardizer) UpdateConfig(config *ContactValidationConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	cvs.config = config
	cvs.logger.Info("contact validation config updated")
	return nil
}

// GetConfig returns the current configuration
func (cvs *ContactValidationStandardizer) GetConfig() *ContactValidationConfig {
	return cvs.config
}

// getDefaultContactValidationConfig returns default configuration
func getDefaultContactValidationConfig() *ContactValidationConfig {
	return &ContactValidationConfig{
		EnablePhoneValidation:        true,
		EnableE164Format:             true,
		AllowedCountryCodes:          []string{},
		DefaultCountryCode:           "1",
		EnableEmailValidation:        true,
		EnableDomainValidation:       true,
		EnableMXValidation:           false, // Disabled by default to avoid DNS lookups
		BlockedDomains:               []string{"tempmail.com", "10minutemail.com", "guerrillamail.com"},
		TrustedDomains:               []string{"gmail.com", "yahoo.com", "outlook.com", "company.com"},
		EnableAddressValidation:      true,
		EnableGeocoding:              false, // Disabled by default
		EnablePostalCodeValidation:   true,
		SupportedCountries:           []string{"US", "CA", "GB", "DE", "FR"},
		EnablePhoneStandardization:   true,
		EnableEmailStandardization:   true,
		EnableAddressStandardization: true,
		MinValidationConfidence:      0.7,
		EnableFuzzyMatching:          false,
		EnableAutoCorrection:         false,
		ValidationTimeout:            30 * time.Second,
		MaxBatchSize:                 1000,
		EnableCaching:                true,
	}
}
