package external

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ContactExtractor manages contact information extraction with enhanced features
type ContactExtractor struct {
	config *ContactExtractionConfig
	logger *zap.Logger
}

// ContactExtractionConfig holds configuration for contact extraction
type ContactExtractionConfig struct {
	EnablePhoneExtraction   bool          `json:"enable_phone_extraction"`
	EnableEmailExtraction   bool          `json:"enable_email_extraction"`
	EnableAddressExtraction bool          `json:"enable_address_extraction"`
	EnableTeamExtraction    bool          `json:"enable_team_extraction"`
	PhonePatterns           []string      `json:"phone_patterns"`
	EmailPatterns           []string      `json:"email_patterns"`
	AddressPatterns         []string      `json:"address_patterns"`
	TeamPatterns            []string      `json:"team_patterns"`
	MaxExtractionTime       time.Duration `json:"max_extraction_time"`
	ConfidenceThreshold     float64       `json:"confidence_threshold"`
	EnableValidation        bool          `json:"enable_validation"`
	EnableStandardization   bool          `json:"enable_standardization"`
	EnablePrivacyCompliance bool          `json:"enable_privacy_compliance"`
	DataRetentionPeriod     time.Duration `json:"data_retention_period"`
	EnableAnonymization     bool          `json:"enable_anonymization"`
}

// EnhancedContactInfo represents extracted contact information with enhanced features
type EnhancedContactInfo struct {
	ID                string                    `json:"id"`
	BusinessID        string                    `json:"business_id"`
	PhoneNumbers      []EnhancedPhoneNumber     `json:"phone_numbers"`
	EmailAddresses    []EnhancedEmailAddress    `json:"email_addresses"`
	PhysicalAddresses []EnhancedPhysicalAddress `json:"physical_addresses"`
	TeamMembers       []EnhancedTeamMember      `json:"team_members"`
	ExtractedAt       time.Time                 `json:"extracted_at"`
	ConfidenceScore   float64                   `json:"confidence_score"`
	DataQuality       DataQualityMetrics        `json:"data_quality"`
	ValidationStatus  ValidationStatus          `json:"validation_status"`
	PrivacyCompliance PrivacyComplianceInfo     `json:"privacy_compliance"`
	Metadata          map[string]interface{}    `json:"metadata"`
}

// EnhancedPhoneNumber represents an extracted phone number with metadata
type EnhancedPhoneNumber struct {
	Number           string  `json:"number"`
	Type             string  `json:"type"` // main, support, sales, etc.
	CountryCode      string  `json:"country_code"`
	ConfidenceScore  float64 `json:"confidence_score"`
	IsValidated      bool    `json:"is_validated"`
	ExtractionMethod string  `json:"extraction_method"`
}

// EnhancedEmailAddress represents an extracted email address with metadata
type EnhancedEmailAddress struct {
	Address          string  `json:"address"`
	Type             string  `json:"type"` // general, support, sales, etc.
	ConfidenceScore  float64 `json:"confidence_score"`
	IsValidated      bool    `json:"is_validated"`
	ExtractionMethod string  `json:"extraction_method"`
}

// EnhancedPhysicalAddress represents an extracted physical address with metadata
type EnhancedPhysicalAddress struct {
	StreetAddress    string  `json:"street_address"`
	City             string  `json:"city"`
	State            string  `json:"state"`
	PostalCode       string  `json:"postal_code"`
	Country          string  `json:"country"`
	ConfidenceScore  float64 `json:"confidence_score"`
	IsValidated      bool    `json:"is_validated"`
	ExtractionMethod string  `json:"extraction_method"`
}

// EnhancedTeamMember represents an extracted team member with metadata
type EnhancedTeamMember struct {
	Name             string  `json:"name"`
	Title            string  `json:"title"`
	Email            string  `json:"email"`
	Department       string  `json:"department"`
	ConfidenceScore  float64 `json:"confidence_score"`
	IsValidated      bool    `json:"is_validated"`
	ExtractionMethod string  `json:"extraction_method"`
}

// DataQualityMetrics represents data quality assessment
type DataQualityMetrics struct {
	Completeness  float64  `json:"completeness"`
	Accuracy      float64  `json:"accuracy"`
	Consistency   float64  `json:"consistency"`
	Timeliness    float64  `json:"timeliness"`
	OverallScore  float64  `json:"overall_score"`
	MissingFields []string `json:"missing_fields"`
	InvalidFields []string `json:"invalid_fields"`
}

// ValidationStatus represents validation status
type ValidationStatus struct {
	IsValid          bool      `json:"is_valid"`
	ValidationErrors []string  `json:"validation_errors"`
	LastValidated    time.Time `json:"last_validated"`
}

// PrivacyComplianceInfo represents privacy compliance status
type PrivacyComplianceInfo struct {
	IsGDPRCompliant bool          `json:"is_gdpr_compliant"`
	IsAnonymized    bool          `json:"is_anonymized"`
	RetentionPeriod time.Duration `json:"retention_period"`
	LastAudit       time.Time     `json:"last_audit"`
	ComplianceScore float64       `json:"compliance_score"`
}

// NewContactExtractor creates a new contact extractor with default configuration
func NewContactExtractor(logger *zap.Logger) *ContactExtractor {
	return &ContactExtractor{
		config: getDefaultContactExtractionConfig(),
		logger: logger,
	}
}

// NewContactExtractorWithConfig creates a new contact extractor with custom configuration
func NewContactExtractorWithConfig(config *ContactExtractionConfig, logger *zap.Logger) *ContactExtractor {
	return &ContactExtractor{
		config: config,
		logger: logger,
	}
}

// ExtractContactInfo extracts contact information from website content
func (ce *ContactExtractor) ExtractContactInfo(ctx context.Context, businessID string, content string) (*EnhancedContactInfo, error) {
	startTime := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, ce.config.MaxExtractionTime)
	defer cancel()

	contactInfo := &EnhancedContactInfo{
		ID:          generateID(),
		BusinessID:  businessID,
		ExtractedAt: time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Extract phone numbers
	if ce.config.EnablePhoneExtraction {
		phoneNumbers, err := ce.extractPhoneNumbers(ctx, content)
		if err != nil {
			ce.logger.Error("failed to extract phone numbers", zap.Error(err))
		} else {
			contactInfo.PhoneNumbers = phoneNumbers
		}
	}

	// Extract email addresses
	if ce.config.EnableEmailExtraction {
		emailAddresses, err := ce.extractEmailAddresses(ctx, content)
		if err != nil {
			ce.logger.Error("failed to extract email addresses", zap.Error(err))
		} else {
			contactInfo.EmailAddresses = emailAddresses
		}
	}

	// Extract physical addresses
	if ce.config.EnableAddressExtraction {
		addresses, err := ce.extractPhysicalAddresses(ctx, content)
		if err != nil {
			ce.logger.Error("failed to extract physical addresses", zap.Error(err))
		} else {
			contactInfo.PhysicalAddresses = addresses
		}
	}

	// Extract team members
	if ce.config.EnableTeamExtraction {
		teamMembers, err := ce.extractTeamMembers(ctx, content)
		if err != nil {
			ce.logger.Error("failed to extract team members", zap.Error(err))
		} else {
			contactInfo.TeamMembers = teamMembers
		}
	}

	// Calculate confidence score
	contactInfo.ConfidenceScore = ce.calculateConfidenceScore(contactInfo)

	// Validate data if enabled
	if ce.config.EnableValidation {
		contactInfo.ValidationStatus = ce.validateContactInfo(contactInfo)
	}

	// Standardize data if enabled
	if ce.config.EnableStandardization {
		ce.standardizeContactInfo(contactInfo)
	}

	// Apply privacy compliance if enabled
	if ce.config.EnablePrivacyCompliance {
		contactInfo.PrivacyCompliance = ce.applyPrivacyCompliance(contactInfo)
	}

	// Calculate data quality metrics
	contactInfo.DataQuality = ce.calculateDataQuality(contactInfo)

	// Add extraction metadata
	contactInfo.Metadata["extraction_duration"] = time.Since(startTime)
	contactInfo.Metadata["content_length"] = len(content)
	contactInfo.Metadata["extraction_methods"] = ce.getExtractionMethods(contactInfo)

	ce.logger.Info("contact information extraction completed",
		zap.String("business_id", businessID),
		zap.Int("phone_count", len(contactInfo.PhoneNumbers)),
		zap.Int("email_count", len(contactInfo.EmailAddresses)),
		zap.Int("address_count", len(contactInfo.PhysicalAddresses)),
		zap.Int("team_count", len(contactInfo.TeamMembers)),
		zap.Float64("confidence_score", contactInfo.ConfidenceScore),
		zap.Duration("duration", time.Since(startTime)))

	return contactInfo, nil
}

// extractPhoneNumbers extracts phone numbers from content
func (ce *ContactExtractor) extractPhoneNumbers(ctx context.Context, content string) ([]EnhancedPhoneNumber, error) {
	var phoneNumbers []EnhancedPhoneNumber

	// Use default patterns if none provided
	patterns := ce.config.PhonePatterns
	if len(patterns) == 0 {
		patterns = getDefaultPhonePatterns()
	}

	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			ce.logger.Error("invalid phone pattern", zap.String("pattern", pattern), zap.Error(err))
			continue
		}

		matches := re.FindAllString(content, -1)
		for _, match := range matches {
			phoneNumber := EnhancedPhoneNumber{
				Number:           strings.TrimSpace(match),
				Type:             ce.determinePhoneType(match, content),
				CountryCode:      ce.determineCountryCode(match),
				ConfidenceScore:  0.8, // Base confidence
				IsValidated:      false,
				ExtractionMethod: "regex_pattern",
			}

			phoneNumbers = append(phoneNumbers, phoneNumber)
		}
	}

	return phoneNumbers, nil
}

// extractEmailAddresses extracts email addresses from content
func (ce *ContactExtractor) extractEmailAddresses(ctx context.Context, content string) ([]EnhancedEmailAddress, error) {
	var emailAddresses []EnhancedEmailAddress

	// Use default patterns if none provided
	patterns := ce.config.EmailPatterns
	if len(patterns) == 0 {
		patterns = getDefaultEmailPatterns()
	}

	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			ce.logger.Error("invalid email pattern", zap.String("pattern", pattern), zap.Error(err))
			continue
		}

		matches := re.FindAllString(content, -1)
		for _, match := range matches {
			emailAddress := EnhancedEmailAddress{
				Address:          strings.TrimSpace(match),
				Type:             ce.determineEmailType(match, content),
				ConfidenceScore:  0.9, // High confidence for email patterns
				IsValidated:      false,
				ExtractionMethod: "regex_pattern",
			}

			emailAddresses = append(emailAddresses, emailAddress)
		}
	}

	return emailAddresses, nil
}

// extractPhysicalAddresses extracts physical addresses from content
func (ce *ContactExtractor) extractPhysicalAddresses(ctx context.Context, content string) ([]EnhancedPhysicalAddress, error) {
	var addresses []EnhancedPhysicalAddress

	// Use default patterns if none provided
	patterns := ce.config.AddressPatterns
	if len(patterns) == 0 {
		patterns = getDefaultAddressPatterns()
	}

	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			ce.logger.Error("invalid address pattern", zap.String("pattern", pattern), zap.Error(err))
			continue
		}

		matches := re.FindAllString(content, -1)
		for _, match := range matches {
			address := ce.parseAddress(match)
			address.ConfidenceScore = 0.7 // Medium confidence for address patterns
			address.IsValidated = false
			address.ExtractionMethod = "regex_pattern"

			addresses = append(addresses, address)
		}
	}

	return addresses, nil
}

// extractTeamMembers extracts team member information from content
func (ce *ContactExtractor) extractTeamMembers(ctx context.Context, content string) ([]EnhancedTeamMember, error) {
	var teamMembers []EnhancedTeamMember

	// Use default patterns if none provided
	patterns := ce.config.TeamPatterns
	if len(patterns) == 0 {
		patterns = getDefaultTeamPatterns()
	}

	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			ce.logger.Error("invalid team pattern", zap.String("pattern", pattern), zap.Error(err))
			continue
		}

		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				teamMember := EnhancedTeamMember{
					Name:             strings.TrimSpace(match[1]),
					Title:            strings.TrimSpace(match[2]),
					Department:       ce.extractDepartment(match[2]),
					ConfidenceScore:  0.6, // Lower confidence for team extraction
					IsValidated:      false,
					ExtractionMethod: "regex_pattern",
				}

				// Extract email if available
				if len(match) > 3 && match[3] != "" {
					teamMember.Email = strings.TrimSpace(match[3])
				}

				teamMembers = append(teamMembers, teamMember)
			}
		}
	}

	return teamMembers, nil
}

// Helper functions for type determination and parsing
func (ce *ContactExtractor) determinePhoneType(number string, content string) string {
	number = strings.ToLower(number)
	content = strings.ToLower(content)

	if strings.Contains(content, "support") || strings.Contains(content, "help") {
		return "support"
	}
	if strings.Contains(content, "sales") {
		return "sales"
	}
	if strings.Contains(content, "main") || strings.Contains(content, "office") {
		return "main"
	}

	return "general"
}

func (ce *ContactExtractor) determineCountryCode(number string) string {
	if strings.HasPrefix(number, "+1") {
		return "US"
	}
	if strings.HasPrefix(number, "+44") {
		return "UK"
	}
	if strings.HasPrefix(number, "+61") {
		return "AU"
	}

	return "unknown"
}

func (ce *ContactExtractor) determineEmailType(email string, content string) string {
	email = strings.ToLower(email)
	content = strings.ToLower(content)

	if strings.Contains(email, "support") || strings.Contains(email, "help") {
		return "support"
	}
	if strings.Contains(email, "sales") {
		return "sales"
	}
	if strings.Contains(email, "info") {
		return "general"
	}

	return "general"
}

func (ce *ContactExtractor) parseAddress(address string) EnhancedPhysicalAddress {
	parts := strings.Split(address, ",")
	addr := EnhancedPhysicalAddress{}

	if len(parts) >= 1 {
		addr.StreetAddress = strings.TrimSpace(parts[0])
	}
	if len(parts) >= 2 {
		addr.City = strings.TrimSpace(parts[1])
	}
	if len(parts) >= 3 {
		addr.State = strings.TrimSpace(parts[2])
	}
	if len(parts) >= 4 {
		addr.PostalCode = strings.TrimSpace(parts[3])
	}
	if len(parts) >= 5 {
		addr.Country = strings.TrimSpace(parts[4])
	}

	return addr
}

func (ce *ContactExtractor) extractDepartment(title string) string {
	title = strings.ToLower(title)

	if strings.Contains(title, "ceo") || strings.Contains(title, "chief") {
		return "executive"
	}
	if strings.Contains(title, "marketing") {
		return "marketing"
	}
	if strings.Contains(title, "developer") || strings.Contains(title, "engineer") {
		return "engineering"
	}

	return "general"
}

// calculateConfidenceScore calculates overall confidence score
func (ce *ContactExtractor) calculateConfidenceScore(contactInfo *EnhancedContactInfo) float64 {
	totalScore := 0.0
	totalWeight := 0.0

	// Weight phone numbers
	if len(contactInfo.PhoneNumbers) > 0 {
		weight := 0.25
		totalWeight += weight
		for _, phone := range contactInfo.PhoneNumbers {
			totalScore += phone.ConfidenceScore * weight
		}
	}

	// Weight email addresses
	if len(contactInfo.EmailAddresses) > 0 {
		weight := 0.25
		totalWeight += weight
		for _, email := range contactInfo.EmailAddresses {
			totalScore += email.ConfidenceScore * weight
		}
	}

	// Weight addresses
	if len(contactInfo.PhysicalAddresses) > 0 {
		weight := 0.25
		totalWeight += weight
		for _, address := range contactInfo.PhysicalAddresses {
			totalScore += address.ConfidenceScore * weight
		}
	}

	// Weight team members
	if len(contactInfo.TeamMembers) > 0 {
		weight := 0.25
		totalWeight += weight
		for _, member := range contactInfo.TeamMembers {
			totalScore += member.ConfidenceScore * weight
		}
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalScore / totalWeight
}

// validateContactInfo validates extracted contact information
func (ce *ContactExtractor) validateContactInfo(contactInfo *EnhancedContactInfo) ValidationStatus {
	status := ValidationStatus{
		IsValid:       true,
		LastValidated: time.Now(),
	}

	// Validate phone numbers
	for _, phone := range contactInfo.PhoneNumbers {
		if !ce.isValidPhoneNumber(phone.Number) {
			status.IsValid = false
			status.ValidationErrors = append(status.ValidationErrors,
				fmt.Sprintf("invalid phone number: %s", phone.Number))
		}
	}

	// Validate email addresses
	for _, email := range contactInfo.EmailAddresses {
		if !ce.isValidEmailAddress(email.Address) {
			status.IsValid = false
			status.ValidationErrors = append(status.ValidationErrors,
				fmt.Sprintf("invalid email address: %s", email.Address))
		}
	}

	return status
}

// standardizeContactInfo standardizes extracted contact information
func (ce *ContactExtractor) standardizeContactInfo(contactInfo *EnhancedContactInfo) {
	// Standardize phone numbers
	for i := range contactInfo.PhoneNumbers {
		contactInfo.PhoneNumbers[i].Number = ce.standardizePhoneNumber(contactInfo.PhoneNumbers[i].Number)
	}

	// Standardize email addresses
	for i := range contactInfo.EmailAddresses {
		contactInfo.EmailAddresses[i].Address = strings.ToLower(strings.TrimSpace(contactInfo.EmailAddresses[i].Address))
	}

	// Standardize addresses
	for i := range contactInfo.PhysicalAddresses {
		contactInfo.PhysicalAddresses[i] = ce.standardizeAddress(contactInfo.PhysicalAddresses[i])
	}
}

// applyPrivacyCompliance applies privacy compliance measures
func (ce *ContactExtractor) applyPrivacyCompliance(contactInfo *EnhancedContactInfo) PrivacyComplianceInfo {
	compliance := PrivacyComplianceInfo{
		IsGDPRCompliant: true,
		IsAnonymized:    ce.config.EnableAnonymization,
		RetentionPeriod: ce.config.DataRetentionPeriod,
		LastAudit:       time.Now(),
		ComplianceScore: 0.9,
	}

	if ce.config.EnableAnonymization {
		ce.anonymizeContactInfo(contactInfo)
	}

	return compliance
}

// calculateDataQuality calculates data quality metrics
func (ce *ContactExtractor) calculateDataQuality(contactInfo *EnhancedContactInfo) DataQualityMetrics {
	metrics := DataQualityMetrics{
		Completeness: 0.0,
		Accuracy:     0.0,
		Consistency:  0.8, // Assume good consistency for extracted data
		Timeliness:   1.0, // Always fresh for new extractions
	}

	// Calculate completeness
	totalFields := 0
	filledFields := 0

	if len(contactInfo.PhoneNumbers) > 0 {
		filledFields++
	}
	totalFields++

	if len(contactInfo.EmailAddresses) > 0 {
		filledFields++
	}
	totalFields++

	if len(contactInfo.PhysicalAddresses) > 0 {
		filledFields++
	}
	totalFields++

	if len(contactInfo.TeamMembers) > 0 {
		filledFields++
	}
	totalFields++

	metrics.Completeness = float64(filledFields) / float64(totalFields)

	// Calculate accuracy based on confidence scores
	totalConfidence := 0.0
	confidenceCount := 0

	for _, phone := range contactInfo.PhoneNumbers {
		totalConfidence += phone.ConfidenceScore
		confidenceCount++
	}

	for _, email := range contactInfo.EmailAddresses {
		totalConfidence += email.ConfidenceScore
		confidenceCount++
	}

	for _, address := range contactInfo.PhysicalAddresses {
		totalConfidence += address.ConfidenceScore
		confidenceCount++
	}

	for _, member := range contactInfo.TeamMembers {
		totalConfidence += member.ConfidenceScore
		confidenceCount++
	}

	if confidenceCount > 0 {
		metrics.Accuracy = totalConfidence / float64(confidenceCount)
	}

	// Calculate overall score
	metrics.OverallScore = (metrics.Completeness + metrics.Accuracy + metrics.Consistency + metrics.Timeliness) / 4.0

	return metrics
}

// Validation helper functions
func (ce *ContactExtractor) isValidPhoneNumber(number string) bool {
	// Remove common formatting
	cleanNumber := strings.ReplaceAll(number, "-", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, " ", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "(", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, ")", "")

	return len(cleanNumber) >= 10
}

func (ce *ContactExtractor) isValidEmailAddress(email string) bool {
	// Basic email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// Standardization helper functions
func (ce *ContactExtractor) standardizePhoneNumber(number string) string {
	// Remove all non-digit characters except +
	cleanNumber := strings.ReplaceAll(number, "-", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, " ", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "(", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, ")", "")

	return cleanNumber
}

func (ce *ContactExtractor) standardizeAddress(address EnhancedPhysicalAddress) EnhancedPhysicalAddress {
	address.StreetAddress = strings.TrimSpace(address.StreetAddress)
	address.City = strings.TrimSpace(address.City)
	address.State = strings.TrimSpace(address.State)
	address.PostalCode = strings.TrimSpace(address.PostalCode)
	address.Country = strings.TrimSpace(address.Country)

	return address
}

// Anonymization helper functions
func (ce *ContactExtractor) anonymizeContactInfo(contactInfo *EnhancedContactInfo) {
	// Anonymize phone numbers
	for i := range contactInfo.PhoneNumbers {
		if len(contactInfo.PhoneNumbers[i].Number) >= 4 {
			contactInfo.PhoneNumbers[i].Number = "***-***-" + contactInfo.PhoneNumbers[i].Number[len(contactInfo.PhoneNumbers[i].Number)-4:]
		}
	}

	for i := range contactInfo.EmailAddresses {
		parts := strings.Split(contactInfo.EmailAddresses[i].Address, "@")
		if len(parts) == 2 {
			if len(parts[0]) > 2 {
				contactInfo.EmailAddresses[i].Address = parts[0][:2] + "***@" + parts[1]
			}
		}
	}
}

// Utility functions
func (ce *ContactExtractor) getExtractionMethods(contactInfo *EnhancedContactInfo) []string {
	methods := make(map[string]bool)

	for _, phone := range contactInfo.PhoneNumbers {
		methods[phone.ExtractionMethod] = true
	}

	for _, email := range contactInfo.EmailAddresses {
		methods[email.ExtractionMethod] = true
	}

	for _, address := range contactInfo.PhysicalAddresses {
		methods[address.ExtractionMethod] = true
	}

	for _, member := range contactInfo.TeamMembers {
		methods[member.ExtractionMethod] = true
	}

	result := make([]string, 0, len(methods))
	for method := range methods {
		result = append(result, method)
	}

	return result
}

// Configuration management
func (ce *ContactExtractor) GetConfig() *ContactExtractionConfig {
	return ce.config
}

func (ce *ContactExtractor) UpdateConfig(config *ContactExtractionConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	ce.config = config
	return nil
}

// Default pattern functions
func getDefaultContactExtractionConfig() *ContactExtractionConfig {
	return &ContactExtractionConfig{
		EnablePhoneExtraction:   true,
		EnableEmailExtraction:   true,
		EnableAddressExtraction: true,
		EnableTeamExtraction:    true,
		MaxExtractionTime:       30 * time.Second,
		ConfidenceThreshold:     0.7,
		EnableValidation:        true,
		EnableStandardization:   true,
		EnablePrivacyCompliance: true,
		DataRetentionPeriod:     90 * 24 * time.Hour, // 90 days
		EnableAnonymization:     false,
	}
}

func getDefaultPhonePatterns() []string {
	return []string{
		`\(\d{3}\)\s*\d{3}-\d{4}`,
		`\d{3}-\d{3}-\d{4}`,
		`\+\d{1,3}\s*\d{1,4}\s*\d{1,4}\s*\d{1,4}`,
		`\d{10,15}`,
	}
}

func getDefaultEmailPatterns() []string {
	return []string{
		`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
	}
}

func getDefaultAddressPatterns() []string {
	return []string{
		`\d+\s+[A-Za-z\s]+,\s*[A-Za-z\s]+,\s*[A-Z]{2}\s*\d{5}`,
		`\d+\s+[A-Za-z\s]+,\s*[A-Za-z\s]+,\s*[A-Za-z\s]+,\s*\d{5}`,
	}
}

func getDefaultTeamPatterns() []string {
	return []string{
		`([A-Za-z\s]+),\s*([A-Za-z\s]+)(?:,\s*([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}))?`,
	}
}

// generateID generates a unique ID for contact information
func generateID() string {
	return uuid.New().String()
}
