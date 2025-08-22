package external

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// EnhancedContactExtractorV2 provides advanced extraction capabilities for phone numbers, emails, and addresses
type EnhancedContactExtractorV2 struct {
	config *EnhancedExtractionConfig
	logger *zap.Logger
}

// EnhancedExtractionConfig extends the base configuration with advanced features
type EnhancedExtractionConfig struct {
	*ContactExtractionConfig
	// Advanced phone extraction
	EnableInternationalPhones bool     `json:"enable_international_phones"`
	EnableTollFreeNumbers     bool     `json:"enable_toll_free_numbers"`
	SupportedCountryCodes     []string `json:"supported_country_codes"`
	PhoneValidationStrict     bool     `json:"phone_validation_strict"`

	// Advanced email extraction
	EnableRoleBasedEmails bool     `json:"enable_role_based_emails"`
	EnablePersonalEmails  bool     `json:"enable_personal_emails"`
	EmailDomainWhitelist  []string `json:"email_domain_whitelist"`
	EmailDomainBlacklist  []string `json:"email_domain_blacklist"`

	// Advanced address extraction
	EnableGeocoding              bool     `json:"enable_geocoding"`
	EnableAddressStandardization bool     `json:"enable_address_standardization"`
	SupportedCountries           []string `json:"supported_countries"`
	RequirePostalCode            bool     `json:"require_postal_code"`

	// Quality and validation
	MinConfidenceThreshold     float64 `json:"min_confidence_threshold"`
	EnableDuplicateDetection   bool    `json:"enable_duplicate_detection"`
	EnableContextualValidation bool    `json:"enable_contextual_validation"`
}

// PhoneExtractionResult represents the result of phone number extraction
type PhoneExtractionResult struct {
	PhoneNumbers    []EnhancedPhoneNumber `json:"phone_numbers"`
	ExtractionStats PhoneExtractionStats  `json:"extraction_stats"`
}

// PhoneExtractionStats provides statistics about phone extraction
type PhoneExtractionStats struct {
	TotalMatches      int     `json:"total_matches"`
	ValidNumbers      int     `json:"valid_numbers"`
	InternationalNums int     `json:"international_numbers"`
	TollFreeNumbers   int     `json:"toll_free_numbers"`
	DuplicatesRemoved int     `json:"duplicates_removed"`
	AverageConfidence float64 `json:"average_confidence"`
}

// EmailExtractionResult represents the result of email extraction
type EmailExtractionResult struct {
	EmailAddresses  []EnhancedEmailAddress `json:"email_addresses"`
	ExtractionStats EmailExtractionStats   `json:"extraction_stats"`
}

// EmailExtractionStats provides statistics about email extraction
type EmailExtractionStats struct {
	TotalMatches      int     `json:"total_matches"`
	ValidEmails       int     `json:"valid_emails"`
	RoleBasedEmails   int     `json:"role_based_emails"`
	PersonalEmails    int     `json:"personal_emails"`
	DuplicatesRemoved int     `json:"duplicates_removed"`
	AverageConfidence float64 `json:"average_confidence"`
}

// AddressExtractionResult represents the result of address extraction
type AddressExtractionResult struct {
	Addresses       []EnhancedPhysicalAddress `json:"addresses"`
	ExtractionStats AddressExtractionStats    `json:"extraction_stats"`
}

// AddressExtractionStats provides statistics about address extraction
type AddressExtractionStats struct {
	TotalMatches      int     `json:"total_matches"`
	ValidAddresses    int     `json:"valid_addresses"`
	CompleteAddresses int     `json:"complete_addresses"`
	GeocodedAddresses int     `json:"geocoded_addresses"`
	DuplicatesRemoved int     `json:"duplicates_removed"`
	AverageConfidence float64 `json:"average_confidence"`
}

// NewEnhancedContactExtractorV2 creates a new enhanced contact extractor
func NewEnhancedContactExtractorV2(config *EnhancedExtractionConfig, logger *zap.Logger) *EnhancedContactExtractorV2 {
	if config == nil {
		config = getDefaultEnhancedExtractionConfig()
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	return &EnhancedContactExtractorV2{
		config: config,
		logger: logger,
	}
}

// ExtractPhoneNumbersAdvanced performs advanced phone number extraction
func (ece *EnhancedContactExtractorV2) ExtractPhoneNumbersAdvanced(ctx context.Context, content string) (*PhoneExtractionResult, error) {
	startTime := time.Now()

	result := &PhoneExtractionResult{
		PhoneNumbers:    make([]EnhancedPhoneNumber, 0),
		ExtractionStats: PhoneExtractionStats{},
	}

	// Check context timeout
	if deadline, ok := ctx.Deadline(); ok && time.Now().After(deadline) {
		return nil, fmt.Errorf("context timeout before phone extraction started")
	}

	// Extract using multiple sophisticated patterns
	patterns := ece.getAdvancedPhonePatterns()
	allMatches := make(map[string]EnhancedPhoneNumber)

	for _, pattern := range patterns {
		matches, err := ece.extractPhonesByPattern(ctx, content, pattern)
		if err != nil {
			ece.logger.Error("phone pattern extraction failed",
				zap.String("pattern", pattern.Name),
				zap.Error(err))
			continue
		}

		for _, match := range matches {
			// Use normalized number as key to prevent duplicates
			normalizedKey := ece.normalizePhoneForDeduplication(match.Number)
			if existing, exists := allMatches[normalizedKey]; !exists || match.ConfidenceScore > existing.ConfidenceScore {
				allMatches[normalizedKey] = match
			}
		}
	}

	// Convert map to slice
	for _, phone := range allMatches {
		result.PhoneNumbers = append(result.PhoneNumbers, phone)
	}

	// Calculate statistics
	result.ExtractionStats = ece.calculatePhoneStats(result.PhoneNumbers)
	result.ExtractionStats.DuplicatesRemoved = len(allMatches) - len(result.PhoneNumbers)

	ece.logger.Info("advanced phone extraction completed",
		zap.Int("total_extracted", len(result.PhoneNumbers)),
		zap.Int("valid_numbers", result.ExtractionStats.ValidNumbers),
		zap.Duration("duration", time.Since(startTime)))

	return result, nil
}

// ExtractEmailAddressesAdvanced performs advanced email address extraction
func (ece *EnhancedContactExtractorV2) ExtractEmailAddressesAdvanced(ctx context.Context, content string) (*EmailExtractionResult, error) {
	startTime := time.Now()

	result := &EmailExtractionResult{
		EmailAddresses:  make([]EnhancedEmailAddress, 0),
		ExtractionStats: EmailExtractionStats{},
	}

	// Check context timeout
	if deadline, ok := ctx.Deadline(); ok && time.Now().After(deadline) {
		return nil, fmt.Errorf("context timeout before email extraction started")
	}

	// Extract using multiple sophisticated patterns
	patterns := ece.getAdvancedEmailPatterns()
	allMatches := make(map[string]EnhancedEmailAddress)

	for _, pattern := range patterns {
		matches, err := ece.extractEmailsByPattern(ctx, content, pattern)
		if err != nil {
			ece.logger.Error("email pattern extraction failed",
				zap.String("pattern", pattern.Name),
				zap.Error(err))
			continue
		}

		for _, match := range matches {
			// Use normalized email as key to prevent duplicates
			normalizedKey := strings.ToLower(strings.TrimSpace(match.Address))
			if existing, exists := allMatches[normalizedKey]; !exists || match.ConfidenceScore > existing.ConfidenceScore {
				allMatches[normalizedKey] = match
			}
		}
	}

	// Filter emails based on configuration
	filteredEmails := ece.filterEmailAddresses(allMatches)

	// Convert map to slice
	for _, email := range filteredEmails {
		result.EmailAddresses = append(result.EmailAddresses, email)
	}

	// Calculate statistics
	result.ExtractionStats = ece.calculateEmailStats(result.EmailAddresses)
	result.ExtractionStats.DuplicatesRemoved = len(allMatches) - len(result.EmailAddresses)

	ece.logger.Info("advanced email extraction completed",
		zap.Int("total_extracted", len(result.EmailAddresses)),
		zap.Int("valid_emails", result.ExtractionStats.ValidEmails),
		zap.Duration("duration", time.Since(startTime)))

	return result, nil
}

// ExtractPhysicalAddressesAdvanced performs advanced physical address extraction
func (ece *EnhancedContactExtractorV2) ExtractPhysicalAddressesAdvanced(ctx context.Context, content string) (*AddressExtractionResult, error) {
	startTime := time.Now()

	result := &AddressExtractionResult{
		Addresses:       make([]EnhancedPhysicalAddress, 0),
		ExtractionStats: AddressExtractionStats{},
	}

	// Check context timeout
	if deadline, ok := ctx.Deadline(); ok && time.Now().After(deadline) {
		return nil, fmt.Errorf("context timeout before address extraction started")
	}

	// Extract using multiple sophisticated patterns
	patterns := ece.getAdvancedAddressPatterns()
	allMatches := make(map[string]EnhancedPhysicalAddress)

	for _, pattern := range patterns {
		matches, err := ece.extractAddressesByPattern(ctx, content, pattern)
		if err != nil {
			ece.logger.Error("address pattern extraction failed",
				zap.String("pattern", pattern.Name),
				zap.Error(err))
			continue
		}

		for _, match := range matches {
			// Use normalized address as key to prevent duplicates
			normalizedKey := ece.normalizeAddressForDeduplication(match)
			if existing, exists := allMatches[normalizedKey]; !exists || match.ConfidenceScore > existing.ConfidenceScore {
				allMatches[normalizedKey] = match
			}
		}
	}

	// Convert map to slice
	for _, address := range allMatches {
		result.Addresses = append(result.Addresses, address)
	}

	// Calculate statistics
	result.ExtractionStats = ece.calculateAddressStats(result.Addresses)
	result.ExtractionStats.DuplicatesRemoved = len(allMatches) - len(result.Addresses)

	ece.logger.Info("advanced address extraction completed",
		zap.Int("total_extracted", len(result.Addresses)),
		zap.Int("valid_addresses", result.ExtractionStats.ValidAddresses),
		zap.Duration("duration", time.Since(startTime)))

	return result, nil
}

// Phone extraction helper types and methods
type PhonePattern struct {
	Name        string
	Regex       *regexp.Regexp
	CountryHint string
	Confidence  float64
}

func (ece *EnhancedContactExtractorV2) getAdvancedPhonePatterns() []PhonePattern {
	patterns := []PhonePattern{
		// US/Canada formats
		{
			Name:        "us_standard",
			Regex:       regexp.MustCompile(`\((\d{3})\)\s*(\d{3})-(\d{4})`),
			CountryHint: "US",
			Confidence:  0.9,
		},
		{
			Name:        "us_dash",
			Regex:       regexp.MustCompile(`(\d{3})-(\d{3})-(\d{4})`),
			CountryHint: "US",
			Confidence:  0.85,
		},
		{
			Name:        "us_dots",
			Regex:       regexp.MustCompile(`(\d{3})\.(\d{3})\.(\d{4})`),
			CountryHint: "US",
			Confidence:  0.8,
		},
		// International formats
		{
			Name:        "international_plus",
			Regex:       regexp.MustCompile(`\+(\d{1,3})\s*[\-\.\s]?\(?(\d{1,4})\)?[\-\.\s]?(\d{1,4})[\-\.\s]?(\d{1,4})`),
			CountryHint: "international",
			Confidence:  0.9,
		},
		// Toll-free numbers
		{
			Name:        "us_toll_free",
			Regex:       regexp.MustCompile(`\b(800|888|877|866|855|844|833|822)\-(\d{3})\-(\d{4})\b`),
			CountryHint: "US",
			Confidence:  0.95,
		},
		// UK formats
		{
			Name:        "uk_standard",
			Regex:       regexp.MustCompile(`\+44\s*(\d{2,4})\s*(\d{4,6})`),
			CountryHint: "UK",
			Confidence:  0.9,
		},
	}

	return patterns
}

func (ece *EnhancedContactExtractorV2) extractPhonesByPattern(ctx context.Context, content string, pattern PhonePattern) ([]EnhancedPhoneNumber, error) {
	var phones []EnhancedPhoneNumber

	matches := pattern.Regex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 0 {
			phone := EnhancedPhoneNumber{
				Number:           match[0],
				Type:             ece.determineAdvancedPhoneType(match[0], content),
				CountryCode:      ece.determineAdvancedCountryCode(match[0], pattern.CountryHint),
				ConfidenceScore:  ece.calculatePhoneConfidence(match[0], pattern),
				IsValidated:      false,
				ExtractionMethod: "advanced_pattern_" + pattern.Name,
			}

			// Additional validation
			if ece.config.PhoneValidationStrict && !ece.isValidPhoneNumberStrict(phone.Number) {
				continue
			}

			phones = append(phones, phone)
		}
	}

	return phones, nil
}

func (ece *EnhancedContactExtractorV2) calculatePhoneConfidence(number string, pattern PhonePattern) float64 {
	baseConfidence := pattern.Confidence

	// Adjust confidence based on number characteristics
	if strings.HasPrefix(number, "+") {
		baseConfidence += 0.05 // International prefix increases confidence
	}

	if ece.isTollFreeNumber(number) {
		baseConfidence += 0.05 // Toll-free numbers are usually business numbers
	}

	// Check if number length is appropriate
	cleanNumber := ece.cleanPhoneNumber(number)
	if len(cleanNumber) >= 10 && len(cleanNumber) <= 15 {
		baseConfidence += 0.02
	}

	// Ensure confidence doesn't exceed 1.0
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}

	return baseConfidence
}

// Email extraction helper types and methods
type EmailPattern struct {
	Name       string
	Regex      *regexp.Regexp
	Type       string
	Confidence float64
}

func (ece *EnhancedContactExtractorV2) getAdvancedEmailPatterns() []EmailPattern {
	patterns := []EmailPattern{
		// Standard email patterns
		{
			Name:       "standard_email",
			Regex:      regexp.MustCompile(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`),
			Type:       "general",
			Confidence: 0.9,
		},
		// Role-based emails
		{
			Name:       "contact_emails",
			Regex:      regexp.MustCompile(`\b(contact|info|hello|inquiries)@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`),
			Type:       "contact",
			Confidence: 0.95,
		},
		{
			Name:       "sales_emails",
			Regex:      regexp.MustCompile(`\b(sales|business|commercial)@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`),
			Type:       "sales",
			Confidence: 0.95,
		},
		{
			Name:       "support_emails",
			Regex:      regexp.MustCompile(`\b(support|help|service|customer)@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`),
			Type:       "support",
			Confidence: 0.95,
		},
	}

	return patterns
}

func (ece *EnhancedContactExtractorV2) extractEmailsByPattern(ctx context.Context, content string, pattern EmailPattern) ([]EnhancedEmailAddress, error) {
	var emails []EnhancedEmailAddress

	matches := pattern.Regex.FindAllString(content, -1)
	for _, match := range matches {
		email := EnhancedEmailAddress{
			Address:          strings.ToLower(strings.TrimSpace(match)),
			Type:             ece.determineAdvancedEmailType(match, pattern.Type),
			ConfidenceScore:  ece.calculateEmailConfidence(match, pattern),
			IsValidated:      false,
			ExtractionMethod: "advanced_pattern_" + pattern.Name,
		}

		// Additional validation
		if !ece.isValidEmailAddressAdvanced(email.Address) {
			continue
		}

		emails = append(emails, email)
	}

	return emails, nil
}

// Address extraction helper types and methods
type AddressPattern struct {
	Name       string
	Regex      *regexp.Regexp
	Country    string
	Confidence float64
}

func (ece *EnhancedContactExtractorV2) getAdvancedAddressPatterns() []AddressPattern {
	patterns := []AddressPattern{
		// US address patterns
		{
			Name:       "us_full_address",
			Regex:      regexp.MustCompile(`\d+\s+[A-Za-z0-9\s]+,\s*[A-Za-z\s]+,\s*[A-Z]{2}\s*\d{5}(?:-\d{4})?`),
			Country:    "US",
			Confidence: 0.9,
		},
		{
			Name:       "us_city_state_zip",
			Regex:      regexp.MustCompile(`[A-Za-z\s]+,\s*[A-Z]{2}\s*\d{5}(?:-\d{4})?`),
			Country:    "US",
			Confidence: 0.8,
		},
		// International patterns
		{
			Name:       "international_postal",
			Regex:      regexp.MustCompile(`[A-Za-z0-9\s]+,\s*[A-Za-z\s]+,\s*[A-Za-z\s]+\s*[A-Z0-9\s]{3,10}`),
			Country:    "international",
			Confidence: 0.7,
		},
	}

	return patterns
}

func (ece *EnhancedContactExtractorV2) extractAddressesByPattern(ctx context.Context, content string, pattern AddressPattern) ([]EnhancedPhysicalAddress, error) {
	var addresses []EnhancedPhysicalAddress

	matches := pattern.Regex.FindAllString(content, -1)
	for _, match := range matches {
		address := ece.parseAdvancedAddress(match)
		address.Country = pattern.Country
		address.ConfidenceScore = ece.calculateAddressConfidence(match, pattern)
		address.ExtractionMethod = "advanced_pattern_" + pattern.Name

		// Additional validation
		if ece.config.RequirePostalCode && address.PostalCode == "" {
			continue
		}

		addresses = append(addresses, address)
	}

	return addresses, nil
}

// Helper methods for validation and processing
func (ece *EnhancedContactExtractorV2) isValidPhoneNumberStrict(number string) bool {
	cleanNumber := ece.cleanPhoneNumber(number)

	// Must be between 10-15 digits
	if len(cleanNumber) < 10 || len(cleanNumber) > 15 {
		return false
	}

	// Should not start with 0 or 1 for US numbers
	if strings.HasPrefix(cleanNumber, "0") || strings.HasPrefix(cleanNumber, "1") {
		if len(cleanNumber) == 10 {
			return false
		}
	}

	return true
}

func (ece *EnhancedContactExtractorV2) isValidEmailAddressAdvanced(email string) bool {
	// More sophisticated email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return false
	}

	// Check domain blacklist
	domain := strings.Split(email, "@")[1]
	for _, blacklisted := range ece.config.EmailDomainBlacklist {
		if domain == blacklisted {
			return false
		}
	}

	// Check domain whitelist (if configured)
	if len(ece.config.EmailDomainWhitelist) > 0 {
		whitelisted := false
		for _, allowed := range ece.config.EmailDomainWhitelist {
			if domain == allowed {
				whitelisted = true
				break
			}
		}
		if !whitelisted {
			return false
		}
	}

	return true
}

// Statistical calculation methods
func (ece *EnhancedContactExtractorV2) calculatePhoneStats(phones []EnhancedPhoneNumber) PhoneExtractionStats {
	stats := PhoneExtractionStats{
		TotalMatches: len(phones),
	}

	if len(phones) == 0 {
		return stats
	}

	totalConfidence := 0.0
	for _, phone := range phones {
		totalConfidence += phone.ConfidenceScore

		if ece.isValidPhoneNumberStrict(phone.Number) {
			stats.ValidNumbers++
		}

		if strings.HasPrefix(phone.Number, "+") {
			stats.InternationalNums++
		}

		if ece.isTollFreeNumber(phone.Number) {
			stats.TollFreeNumbers++
		}
	}

	stats.AverageConfidence = totalConfidence / float64(len(phones))
	return stats
}

func (ece *EnhancedContactExtractorV2) calculateEmailStats(emails []EnhancedEmailAddress) EmailExtractionStats {
	stats := EmailExtractionStats{
		TotalMatches: len(emails),
	}

	if len(emails) == 0 {
		return stats
	}

	totalConfidence := 0.0
	for _, email := range emails {
		totalConfidence += email.ConfidenceScore

		if ece.isValidEmailAddressAdvanced(email.Address) {
			stats.ValidEmails++
		}

		if ece.isRoleBasedEmail(email.Address) {
			stats.RoleBasedEmails++
		} else {
			stats.PersonalEmails++
		}
	}

	stats.AverageConfidence = totalConfidence / float64(len(emails))
	return stats
}

func (ece *EnhancedContactExtractorV2) calculateAddressStats(addresses []EnhancedPhysicalAddress) AddressExtractionStats {
	stats := AddressExtractionStats{
		TotalMatches: len(addresses),
	}

	if len(addresses) == 0 {
		return stats
	}

	totalConfidence := 0.0
	for _, address := range addresses {
		totalConfidence += address.ConfidenceScore

		if ece.isValidAddress(address) {
			stats.ValidAddresses++
		}

		if ece.isCompleteAddress(address) {
			stats.CompleteAddresses++
		}
	}

	stats.AverageConfidence = totalConfidence / float64(len(addresses))
	return stats
}

// Utility methods
func (ece *EnhancedContactExtractorV2) cleanPhoneNumber(number string) string {
	// Remove all non-digit characters except +
	cleaned := regexp.MustCompile(`[^\d+]`).ReplaceAllString(number, "")
	return cleaned
}

func (ece *EnhancedContactExtractorV2) isTollFreeNumber(number string) bool {
	tollFreePatterns := []string{"800", "888", "877", "866", "855", "844", "833", "822"}
	for _, pattern := range tollFreePatterns {
		if strings.Contains(number, pattern) {
			return true
		}
	}
	return false
}

func (ece *EnhancedContactExtractorV2) isRoleBasedEmail(email string) bool {
	roleKeywords := []string{"info", "contact", "sales", "support", "admin", "service", "help"}
	emailLower := strings.ToLower(email)
	for _, keyword := range roleKeywords {
		if strings.Contains(emailLower, keyword) {
			return true
		}
	}
	return false
}

func (ece *EnhancedContactExtractorV2) isValidAddress(address EnhancedPhysicalAddress) bool {
	return address.StreetAddress != "" && address.City != ""
}

func (ece *EnhancedContactExtractorV2) isCompleteAddress(address EnhancedPhysicalAddress) bool {
	return address.StreetAddress != "" && address.City != "" && address.State != "" && address.PostalCode != ""
}

// Configuration and setup methods
func getDefaultEnhancedExtractionConfig() *EnhancedExtractionConfig {
	return &EnhancedExtractionConfig{
		ContactExtractionConfig:      getDefaultContactExtractionConfig(),
		EnableInternationalPhones:    true,
		EnableTollFreeNumbers:        true,
		SupportedCountryCodes:        []string{"US", "CA", "UK", "AU"},
		PhoneValidationStrict:        true,
		EnableRoleBasedEmails:        true,
		EnablePersonalEmails:         true,
		EmailDomainBlacklist:         []string{"example.com", "test.com", "localhost"},
		EnableGeocoding:              false,
		EnableAddressStandardization: true,
		SupportedCountries:           []string{"US", "CA", "UK", "AU"},
		RequirePostalCode:            false,
		MinConfidenceThreshold:       0.7,
		EnableDuplicateDetection:     true,
		EnableContextualValidation:   true,
	}
}

// Additional helper methods for enhanced functionality
func (ece *EnhancedContactExtractorV2) determineAdvancedPhoneType(number string, content string) string {
	if ece.isTollFreeNumber(number) {
		return "toll_free"
	}

	if strings.HasPrefix(number, "+") {
		return "international"
	}

	// Context-based determination
	context := strings.ToLower(content)
	if strings.Contains(context, "mobile") || strings.Contains(context, "cell") {
		return "mobile"
	}

	if strings.Contains(context, "office") || strings.Contains(context, "main") {
		return "office"
	}

	if strings.Contains(context, "fax") {
		return "fax"
	}

	return "general"
}

func (ece *EnhancedContactExtractorV2) determineAdvancedCountryCode(number string, hint string) string {
	if hint != "international" {
		return hint
	}

	// Extract country code from international number
	if strings.HasPrefix(number, "+1") {
		return "US"
	}
	if strings.HasPrefix(number, "+44") {
		return "UK"
	}
	if strings.HasPrefix(number, "+61") {
		return "AU"
	}
	if strings.HasPrefix(number, "+49") {
		return "DE"
	}
	if strings.HasPrefix(number, "+33") {
		return "FR"
	}

	return "unknown"
}

func (ece *EnhancedContactExtractorV2) determineAdvancedEmailType(email string, patternType string) string {
	if patternType != "general" {
		return patternType
	}

	emailLower := strings.ToLower(email)

	if strings.Contains(emailLower, "ceo") || strings.Contains(emailLower, "founder") {
		return "executive"
	}
	if strings.Contains(emailLower, "admin") || strings.Contains(emailLower, "administrator") {
		return "admin"
	}
	if strings.Contains(emailLower, "marketing") {
		return "marketing"
	}
	if strings.Contains(emailLower, "hr") || strings.Contains(emailLower, "human") {
		return "hr"
	}

	return "general"
}

func (ece *EnhancedContactExtractorV2) parseAdvancedAddress(addressText string) EnhancedPhysicalAddress {
	address := EnhancedPhysicalAddress{}

	// More sophisticated address parsing
	parts := strings.Split(addressText, ",")

	if len(parts) >= 1 {
		address.StreetAddress = strings.TrimSpace(parts[0])
	}

	if len(parts) >= 2 {
		address.City = strings.TrimSpace(parts[1])
	}

	if len(parts) >= 3 {
		// Try to parse state and postal code from last part
		lastPart := strings.TrimSpace(parts[len(parts)-1])

		// US format: State ZIP
		stateZipRegex := regexp.MustCompile(`([A-Z]{2})\s*(\d{5}(?:-\d{4})?)`)
		if match := stateZipRegex.FindStringSubmatch(lastPart); len(match) == 3 {
			address.State = match[1]
			address.PostalCode = match[2]
		} else {
			// International format - just use as state/region
			address.State = lastPart
		}
	}

	return address
}

func (ece *EnhancedContactExtractorV2) calculateEmailConfidence(email string, pattern EmailPattern) float64 {
	baseConfidence := pattern.Confidence

	// Adjust based on email characteristics
	if ece.isRoleBasedEmail(email) {
		baseConfidence += 0.05 // Role-based emails are typically business emails
	}

	// Check domain reputation (simplified)
	domain := strings.Split(email, "@")[1]
	if ece.isBusinessDomain(domain) {
		baseConfidence += 0.03
	}

	// Ensure confidence doesn't exceed 1.0
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}

	return baseConfidence
}

func (ece *EnhancedContactExtractorV2) calculateAddressConfidence(address string, pattern AddressPattern) float64 {
	baseConfidence := pattern.Confidence

	// Adjust based on address completeness
	if strings.Contains(address, ",") {
		baseConfidence += 0.02 // Comma-separated parts indicate structure
	}

	// Check for postal code
	if regexp.MustCompile(`\d{5}`).MatchString(address) {
		baseConfidence += 0.05 // Postal code increases confidence
	}

	// Ensure confidence doesn't exceed 1.0
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}

	return baseConfidence
}

func (ece *EnhancedContactExtractorV2) isBusinessDomain(domain string) bool {
	// List of common business domains (simplified)
	businessDomains := []string{
		"gmail.com", "outlook.com", "yahoo.com", // Actually personal, but commonly used by small businesses
	}

	for _, businessDomain := range businessDomains {
		if domain == businessDomain {
			return false // These are actually personal domains
		}
	}

	// If it's not a common personal domain, assume it's business
	return !strings.Contains(domain, "gmail") && !strings.Contains(domain, "yahoo") && !strings.Contains(domain, "hotmail")
}

// Filtering and deduplication methods
func (ece *EnhancedContactExtractorV2) filterEmailAddresses(emails map[string]EnhancedEmailAddress) map[string]EnhancedEmailAddress {
	filtered := make(map[string]EnhancedEmailAddress)

	for key, email := range emails {
		// Apply confidence threshold
		if email.ConfidenceScore < ece.config.MinConfidenceThreshold {
			continue
		}

		// Apply role-based email filter
		if !ece.config.EnableRoleBasedEmails && ece.isRoleBasedEmail(email.Address) {
			continue
		}

		// Apply personal email filter
		if !ece.config.EnablePersonalEmails && !ece.isRoleBasedEmail(email.Address) {
			continue
		}

		filtered[key] = email
	}

	return filtered
}

func (ece *EnhancedContactExtractorV2) normalizePhoneForDeduplication(number string) string {
	// Remove all formatting and normalize for deduplication
	normalized := ece.cleanPhoneNumber(number)

	// Remove leading country codes for US numbers
	if strings.HasPrefix(normalized, "+1") && len(normalized) == 12 {
		normalized = normalized[2:]
	}

	return normalized
}

func (ece *EnhancedContactExtractorV2) normalizeAddressForDeduplication(address EnhancedPhysicalAddress) string {
	// Create a normalized string for deduplication
	normalized := strings.ToLower(
		address.StreetAddress + "|" +
			address.City + "|" +
			address.State + "|" +
			address.PostalCode,
	)

	// Remove extra spaces
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")

	return strings.TrimSpace(normalized)
}
