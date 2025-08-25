package data_extraction

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/shared"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// EnhancedContactExtractor extracts enhanced contact and business information
type EnhancedContactExtractor struct {
	// Configuration
	config *EnhancedContactConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Pattern matching
	emailPatterns         []*regexp.Regexp
	phonePatterns         []*regexp.Regexp
	addressPatterns       []*regexp.Regexp
	socialMediaPatterns   []*regexp.Regexp
	teamMemberPatterns    []*regexp.Regexp
	businessHoursPatterns []*regexp.Regexp
	locationPatterns      []*regexp.Regexp
}

// EnhancedContactConfig holds configuration for the enhanced contact extractor
type EnhancedContactConfig struct {
	// Pattern matching settings
	CaseSensitive bool
	MaxPatterns   int

	// Validation settings
	EnableValidation       bool
	StrictValidation       bool
	MinConfidenceThreshold float64

	// Processing settings
	Timeout time.Duration
}

// EnhancedContactInfo represents enhanced contact and business information
type EnhancedContactInfo struct {
	// Contact information
	Emails          []string           `json:"emails"`
	EmailConfidence map[string]float64 `json:"email_confidence"`
	Phones          []string           `json:"phones"`
	PhoneConfidence map[string]float64 `json:"phone_confidence"`

	// Address information
	Addresses          []string           `json:"addresses"`
	AddressConfidence  map[string]float64 `json:"address_confidence"`
	ValidatedAddresses []ValidatedAddress `json:"validated_addresses,omitempty"`

	// Social media presence
	SocialMediaAccounts   map[string]string  `json:"social_media_accounts"`
	SocialMediaConfidence map[string]float64 `json:"social_media_confidence"`

	// Team information
	TeamMembers          []TeamMember       `json:"team_members"`
	TeamMemberConfidence map[string]float64 `json:"team_member_confidence"`

	// Business hours and location
	BusinessHours           []BusinessHours    `json:"business_hours"`
	BusinessHoursConfidence map[string]float64 `json:"business_hours_confidence"`
	Locations               []BusinessLocation `json:"locations"`
	LocationConfidence      map[string]float64 `json:"location_confidence"`

	// Additional details
	ContactDetails     map[string]interface{} `json:"contact_details,omitempty"`
	SupportingEvidence []string               `json:"supporting_evidence,omitempty"`

	// Overall assessment
	OverallConfidence float64 `json:"overall_confidence"`

	// Metadata
	ExtractedAt time.Time `json:"extracted_at"`
	DataSources []string  `json:"data_sources"`
}

// ValidatedAddress represents a validated address
type ValidatedAddress struct {
	StreetAddress   string  `json:"street_address"`
	City            string  `json:"city"`
	State           string  `json:"state"`
	PostalCode      string  `json:"postal_code"`
	Country         string  `json:"country"`
	Latitude        float64 `json:"latitude,omitempty"`
	Longitude       float64 `json:"longitude,omitempty"`
	ValidationScore float64 `json:"validation_score"`
	IsValid         bool    `json:"is_valid"`
}

// TeamMember represents a team member
type TeamMember struct {
	Name       string  `json:"name"`
	Title      string  `json:"title"`
	Email      string  `json:"email,omitempty"`
	LinkedIn   string  `json:"linkedin,omitempty"`
	Twitter    string  `json:"twitter,omitempty"`
	Confidence float64 `json:"confidence"`
}

// BusinessHours represents business operating hours
type BusinessHours struct {
	DayOfWeek  string  `json:"day_of_week"`
	OpenTime   string  `json:"open_time"`
	CloseTime  string  `json:"close_time"`
	IsOpen     bool    `json:"is_open"`
	Confidence float64 `json:"confidence"`
}

// BusinessLocation represents a business location
type BusinessLocation struct {
	Name       string  `json:"name"`
	Address    string  `json:"address"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	Country    string  `json:"country"`
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
	Phone      string  `json:"phone,omitempty"`
	Confidence float64 `json:"confidence"`
}

// Social media platform constants
const (
	SocialLinkedIn  = "linkedin"
	SocialTwitter   = "twitter"
	SocialFacebook  = "facebook"
	SocialInstagram = "instagram"
	SocialYouTube   = "youtube"
	SocialTikTok    = "tiktok"
	SocialGitHub    = "github"
	SocialMedium    = "medium"
	SocialReddit    = "reddit"
	SocialDiscord   = "discord"
	SocialSlack     = "slack"
	SocialTelegram  = "telegram"
)

// NewEnhancedContactExtractor creates a new enhanced contact extractor
func NewEnhancedContactExtractor(
	config *EnhancedContactConfig,
	logger *observability.Logger,
	tracer trace.Tracer,
) *EnhancedContactExtractor {
	// Set default configuration
	if config == nil {
		config = &EnhancedContactConfig{
			CaseSensitive:          false,
			MaxPatterns:            100,
			EnableValidation:       true,
			StrictValidation:       false,
			MinConfidenceThreshold: 0.3,
			Timeout:                30 * time.Second,
		}
	}

	extractor := &EnhancedContactExtractor{
		config: config,
		logger: logger,
		tracer: tracer,
	}

	// Initialize pattern matching
	extractor.initializePatterns()

	return extractor
}

// initializePatterns initializes all pattern matching regexes
func (ece *EnhancedContactExtractor) initializePatterns() {
	// Email patterns
	ece.emailPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		regexp.MustCompile(`(?i)email[:\s]*([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`),
		regexp.MustCompile(`(?i)contact[:\s]*([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`),
	}

	// Phone patterns
	ece.phonePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(\+\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}`),
		regexp.MustCompile(`(?i)phone[:\s]*((\+\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4})`),
		regexp.MustCompile(`(?i)tel[:\s]*((\+\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4})`),
	}

	// Address patterns
	ece.addressPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\d+\s+[a-zA-Z\s]+(?:street|st|avenue|ave|road|rd|boulevard|blvd|lane|ln|drive|dr|way|place|pl|court|ct)\.?`),
		regexp.MustCompile(`(?i)address[:\s]*([^,\n]+)`),
		regexp.MustCompile(`(?i)location[:\s]*([^,\n]+)`),
	}

	// Social media patterns
	ece.socialMediaPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)linkedin\.com/(?:company/|in/)?([a-zA-Z0-9_-]+)`),
		regexp.MustCompile(`(?i)twitter\.com/([a-zA-Z0-9_]+)`),
		regexp.MustCompile(`(?i)facebook\.com/([a-zA-Z0-9.]+)`),
		regexp.MustCompile(`(?i)instagram\.com/([a-zA-Z0-9_.]+)`),
		regexp.MustCompile(`(?i)youtube\.com/(?:channel/|c/|user/)?([a-zA-Z0-9_-]+)`),
		regexp.MustCompile(`(?i)tiktok\.com/@([a-zA-Z0-9_.]+)`),
		regexp.MustCompile(`(?i)github\.com/([a-zA-Z0-9_-]+)`),
		regexp.MustCompile(`(?i)medium\.com/@([a-zA-Z0-9_-]+)`),
		regexp.MustCompile(`(?i)reddit\.com/r/([a-zA-Z0-9_]+)`),
		regexp.MustCompile(`(?i)discord\.gg/([a-zA-Z0-9]+)`),
		regexp.MustCompile(`(?i)slack\.com/archives/([a-zA-Z0-9]+)`),
		regexp.MustCompile(`(?i)t\.me/([a-zA-Z0-9_]+)`),
	}

	// Team member patterns
	ece.teamMemberPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:ceo|cto|cfo|coo|founder|co-founder|president|vice president|vp|director|manager|lead|senior|junior)\s+([a-zA-Z\s]+)`),
		regexp.MustCompile(`(?i)([a-zA-Z\s]+)\s+(?:ceo|cto|cfo|coo|founder|co-founder|president|vice president|vp|director|manager|lead|senior|junior)`),
		regexp.MustCompile(`(?i)team[:\s]*([^,\n]+)`),
		regexp.MustCompile(`(?i)leadership[:\s]*([^,\n]+)`),
	}

	// Business hours patterns
	ece.businessHoursPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:monday|mon)[:\s]*(\d{1,2}:\d{2}\s*(?:am|pm)?\s*-\s*\d{1,2}:\d{2}\s*(?:am|pm)?)`),
		regexp.MustCompile(`(?i)(?:tuesday|tue)[:\s]*(\d{1,2}:\d{2}\s*(?:am|pm)?\s*-\s*\d{1,2}:\d{2}\s*(?:am|pm)?)`),
		regexp.MustCompile(`(?i)(?:wednesday|wed)[:\s]*(\d{1,2}:\d{2}\s*(?:am|pm)?\s*-\s*\d{1,2}:\d{2}\s*(?:am|pm)?)`),
		regexp.MustCompile(`(?i)(?:thursday|thu)[:\s]*(\d{1,2}:\d{2}\s*(?:am|pm)?\s*-\s*\d{1,2}:\d{2}\s*(?:am|pm)?)`),
		regexp.MustCompile(`(?i)(?:friday|fri)[:\s]*(\d{1,2}:\d{2}\s*(?:am|pm)?\s*-\s*\d{1,2}:\d{2}\s*(?:am|pm)?)`),
		regexp.MustCompile(`(?i)(?:saturday|sat)[:\s]*(\d{1,2}:\d{2}\s*(?:am|pm)?\s*-\s*\d{1,2}:\d{2}\s*(?:am|pm)?)`),
		regexp.MustCompile(`(?i)(?:sunday|sun)[:\s]*(\d{1,2}:\d{2}\s*(?:am|pm)?\s*-\s*\d{1,2}:\d{2}\s*(?:am|pm)?)`),
		regexp.MustCompile(`(?i)hours[:\s]*([^,\n]+)`),
		regexp.MustCompile(`(?i)open[:\s]*([^,\n]+)`),
	}

	// Location patterns
	ece.locationPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)office[:\s]*([^,\n]+)`),
		regexp.MustCompile(`(?i)headquarters[:\s]*([^,\n]+)`),
		regexp.MustCompile(`(?i)location[:\s]*([^,\n]+)`),
		regexp.MustCompile(`(?i)address[:\s]*([^,\n]+)`),
	}
}

// ExtractEnhancedContactInfo extracts enhanced contact and business information
func (ece *EnhancedContactExtractor) ExtractEnhancedContactInfo(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
) (*EnhancedContactInfo, error) {
	ctx, span := ece.tracer.Start(ctx, "EnhancedContactExtractor.ExtractEnhancedContactInfo")
	defer span.End()

	span.SetAttributes(
		attribute.String("business_name", businessData.BusinessName),
		attribute.String("website", businessData.WebsiteURL),
	)

	// Create result structure
	result := &EnhancedContactInfo{
		ExtractedAt:             time.Now(),
		DataSources:             []string{"text_analysis", "pattern_matching"},
		ContactDetails:          make(map[string]interface{}),
		EmailConfidence:         make(map[string]float64),
		PhoneConfidence:         make(map[string]float64),
		AddressConfidence:       make(map[string]float64),
		SocialMediaAccounts:     make(map[string]string),
		SocialMediaConfidence:   make(map[string]float64),
		TeamMemberConfidence:    make(map[string]float64),
		BusinessHoursConfidence: make(map[string]float64),
		LocationConfidence:      make(map[string]float64),
	}

	// Extract contact information
	if err := ece.extractContactInformation(ctx, businessData, result); err != nil {
		ece.logger.Warn("failed to extract contact information", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Extract and validate addresses
	if err := ece.extractAndValidateAddresses(ctx, businessData, result); err != nil {
		ece.logger.Warn("failed to extract and validate addresses", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Extract social media presence
	if err := ece.extractSocialMediaPresence(ctx, businessData, result); err != nil {
		ece.logger.Warn("failed to extract social media presence", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Extract team members
	if err := ece.extractTeamMembers(ctx, businessData, result); err != nil {
		ece.logger.Warn("failed to extract team members", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Extract business hours
	if err := ece.extractBusinessHours(ctx, businessData, result); err != nil {
		ece.logger.Warn("failed to extract business hours", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Extract business locations
	if err := ece.extractBusinessLocations(ctx, businessData, result); err != nil {
		ece.logger.Warn("failed to extract business locations", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Calculate overall confidence
	ece.calculateOverallConfidence(result)

	// Validate results
	if err := ece.validateResults(result); err != nil {
		ece.logger.Warn("enhanced contact validation failed", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	ece.logger.Info("enhanced contact extraction completed", map[string]interface{}{
		"business_name":         businessData.BusinessName,
		"emails":                len(result.Emails),
		"phones":                len(result.Phones),
		"addresses":             len(result.Addresses),
		"social_media_accounts": len(result.SocialMediaAccounts),
		"team_members":          len(result.TeamMembers),
		"business_hours":        len(result.BusinessHours),
		"locations":             len(result.Locations),
		"overall_confidence":    result.OverallConfidence,
	})

	return result, nil
}

// extractContactInformation extracts emails and phone numbers
func (ece *EnhancedContactExtractor) extractContactInformation(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *EnhancedContactInfo,
) error {
	ctx, span := ece.tracer.Start(ctx, "EnhancedContactExtractor.extractContactInformation")
	defer span.End()

	text := ece.combineText(businessData)

	// Extract emails
	for _, pattern := range ece.emailPatterns {
		matches := pattern.FindAllString(text, -1)
		for _, match := range matches {
			email := ece.cleanEmail(match)
			if email != "" && !ece.containsEmail(result.Emails, email) {
				result.Emails = append(result.Emails, email)
				result.EmailConfidence[email] = 0.9
				result.SupportingEvidence = append(result.SupportingEvidence, match)
			}
		}
	}

	// Extract phone numbers
	for _, pattern := range ece.phonePatterns {
		matches := pattern.FindAllString(text, -1)
		for _, match := range matches {
			phone := ece.cleanPhone(match)
			if phone != "" && !ece.containsPhone(result.Phones, phone) {
				result.Phones = append(result.Phones, phone)
				result.PhoneConfidence[phone] = 0.8
				result.SupportingEvidence = append(result.SupportingEvidence, match)
			}
		}
	}

	span.SetAttributes(
		attribute.StringSlice("emails", result.Emails),
		attribute.StringSlice("phones", result.Phones),
		attribute.Int("email_count", len(result.Emails)),
		attribute.Int("phone_count", len(result.Phones)),
	)

	return nil
}

// extractAndValidateAddresses extracts and validates addresses
func (ece *EnhancedContactExtractor) extractAndValidateAddresses(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *EnhancedContactInfo,
) error {
	ctx, span := ece.tracer.Start(ctx, "EnhancedContactExtractor.extractAndValidateAddresses")
	defer span.End()

	text := ece.combineText(businessData)

	// Extract addresses
	for _, pattern := range ece.addressPatterns {
		matches := pattern.FindAllString(text, -1)
		for _, match := range matches {
			address := ece.cleanAddress(match)
			if address != "" && !ece.containsAddress(result.Addresses, address) {
				result.Addresses = append(result.Addresses, address)
				result.AddressConfidence[address] = 0.7
				result.SupportingEvidence = append(result.SupportingEvidence, match)

				// Validate address if enabled
				if ece.config.EnableValidation {
					validatedAddress := ece.validateAddress(address)
					result.ValidatedAddresses = append(result.ValidatedAddresses, validatedAddress)
				}
			}
		}
	}

	span.SetAttributes(
		attribute.StringSlice("addresses", result.Addresses),
		attribute.Int("address_count", len(result.Addresses)),
		attribute.Int("validated_address_count", len(result.ValidatedAddresses)),
	)

	return nil
}

// extractSocialMediaPresence extracts social media accounts
func (ece *EnhancedContactExtractor) extractSocialMediaPresence(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *EnhancedContactInfo,
) error {
	ctx, span := ece.tracer.Start(ctx, "EnhancedContactExtractor.extractSocialMediaPresence")
	defer span.End()

	text := ece.combineText(businessData)

	// Extract social media accounts
	for _, pattern := range ece.socialMediaPatterns {
		matches := pattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				platform := ece.mapPatternToSocialPlatform(pattern.String())
				account := match[1]
				if platform != "" && account != "" {
					result.SocialMediaAccounts[platform] = account
					result.SocialMediaConfidence[platform] = 0.9
					result.SupportingEvidence = append(result.SupportingEvidence, match[0])
				}
			}
		}
	}

	span.SetAttributes(
		attribute.Int("social_media_count", len(result.SocialMediaAccounts)),
	)

	return nil
}

// extractTeamMembers extracts team member information
func (ece *EnhancedContactExtractor) extractTeamMembers(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *EnhancedContactInfo,
) error {
	ctx, span := ece.tracer.Start(ctx, "EnhancedContactExtractor.extractTeamMembers")
	defer span.End()

	text := ece.combineText(businessData)

	// Extract team members
	for _, pattern := range ece.teamMemberPatterns {
		matches := pattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				name := strings.TrimSpace(match[1])
				if name != "" {
					teamMember := TeamMember{
						Name:       name,
						Title:      ece.extractTitle(match[0]),
						Confidence: 0.7,
					}
					result.TeamMembers = append(result.TeamMembers, teamMember)
					result.TeamMemberConfidence[name] = 0.7
					result.SupportingEvidence = append(result.SupportingEvidence, match[0])
				}
			}
		}
	}

	span.SetAttributes(
		attribute.Int("team_member_count", len(result.TeamMembers)),
	)

	return nil
}

// extractBusinessHours extracts business operating hours
func (ece *EnhancedContactExtractor) extractBusinessHours(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *EnhancedContactInfo,
) error {
	ctx, span := ece.tracer.Start(ctx, "EnhancedContactExtractor.extractBusinessHours")
	defer span.End()

	text := ece.combineText(businessData)

	// Extract business hours
	for _, pattern := range ece.businessHoursPatterns {
		matches := pattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				dayOfWeek := ece.extractDayOfWeek(match[0])
				hours := strings.TrimSpace(match[1])
				if dayOfWeek != "" && hours != "" {
					businessHours := BusinessHours{
						DayOfWeek:  dayOfWeek,
						OpenTime:   ece.extractOpenTime(hours),
						CloseTime:  ece.extractCloseTime(hours),
						IsOpen:     true,
						Confidence: 0.8,
					}
					result.BusinessHours = append(result.BusinessHours, businessHours)
					result.BusinessHoursConfidence[dayOfWeek] = 0.8
					result.SupportingEvidence = append(result.SupportingEvidence, match[0])
				}
			}
		}
	}

	span.SetAttributes(
		attribute.Int("business_hours_count", len(result.BusinessHours)),
	)

	return nil
}

// extractBusinessLocations extracts business locations
func (ece *EnhancedContactExtractor) extractBusinessLocations(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *EnhancedContactInfo,
) error {
	ctx, span := ece.tracer.Start(ctx, "EnhancedContactExtractor.extractBusinessLocations")
	defer span.End()

	text := ece.combineText(businessData)

	// Extract business locations
	for _, pattern := range ece.locationPatterns {
		matches := pattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				location := strings.TrimSpace(match[1])
				if location != "" {
					businessLocation := BusinessLocation{
						Name:       ece.extractLocationName(match[0]),
						Address:    location,
						City:       ece.extractCity(location),
						State:      ece.extractState(location),
						Country:    ece.extractCountry(location),
						Confidence: 0.7,
					}
					result.Locations = append(result.Locations, businessLocation)
					result.LocationConfidence[location] = 0.7
					result.SupportingEvidence = append(result.SupportingEvidence, match[0])
				}
			}
		}
	}

	span.SetAttributes(
		attribute.Int("location_count", len(result.Locations)),
	)

	return nil
}

// Helper functions for data cleaning and extraction
func (ece *EnhancedContactExtractor) combineText(businessData *shared.BusinessClassificationRequest) string {
	var parts []string

	// Add business name
	if businessData.BusinessName != "" {
		parts = append(parts, businessData.BusinessName)
	}

	// Add description
	if businessData.Description != "" {
		parts = append(parts, businessData.Description)
	}

	// Add keywords
	if len(businessData.Keywords) > 0 {
		parts = append(parts, strings.Join(businessData.Keywords, " "))
	}

	// Add address
	if businessData.Address != "" {
		parts = append(parts, businessData.Address)
	}

	// Combine all parts
	text := strings.Join(parts, " ")

	// Normalize text
	if !ece.config.CaseSensitive {
		text = strings.ToLower(text)
	}

	return text
}

func (ece *EnhancedContactExtractor) cleanEmail(email string) string {
	// Remove common prefixes
	email = strings.TrimPrefix(email, "email:")
	email = strings.TrimPrefix(email, "contact:")
	email = strings.TrimSpace(email)

	// Basic email validation
	if strings.Contains(email, "@") && strings.Contains(email, ".") {
		return email
	}
	return ""
}

func (ece *EnhancedContactExtractor) cleanPhone(phone string) string {
	// Remove common prefixes
	phone = strings.TrimPrefix(phone, "phone:")
	phone = strings.TrimPrefix(phone, "tel:")
	phone = strings.TrimSpace(phone)

	// Remove non-digit characters except +, (, ), -, .
	phone = regexp.MustCompile(`[^\d+\-\(\)\.]`).ReplaceAllString(phone, "")

	return phone
}

func (ece *EnhancedContactExtractor) cleanAddress(address string) string {
	// Remove common prefixes
	address = strings.TrimPrefix(address, "address:")
	address = strings.TrimPrefix(address, "location:")
	address = strings.TrimSpace(address)

	return address
}

func (ece *EnhancedContactExtractor) validateAddress(address string) ValidatedAddress {
	// Placeholder for address validation
	// In a real implementation, this would use a geocoding service
	validatedAddress := ValidatedAddress{
		StreetAddress:   address,
		ValidationScore: 0.5,   // Placeholder score
		IsValid:         false, // Placeholder validation
	}

	return validatedAddress
}

func (ece *EnhancedContactExtractor) mapPatternToSocialPlatform(pattern string) string {
	pattern = strings.ToLower(pattern)
	switch {
	case strings.Contains(pattern, "linkedin"):
		return SocialLinkedIn
	case strings.Contains(pattern, "twitter"):
		return SocialTwitter
	case strings.Contains(pattern, "facebook"):
		return SocialFacebook
	case strings.Contains(pattern, "instagram"):
		return SocialInstagram
	case strings.Contains(pattern, "youtube"):
		return SocialYouTube
	case strings.Contains(pattern, "tiktok"):
		return SocialTikTok
	case strings.Contains(pattern, "github"):
		return SocialGitHub
	case strings.Contains(pattern, "medium"):
		return SocialMedium
	case strings.Contains(pattern, "reddit"):
		return SocialReddit
	case strings.Contains(pattern, "discord"):
		return SocialDiscord
	case strings.Contains(pattern, "slack"):
		return SocialSlack
	case strings.Contains(pattern, "t\\.me"):
		return SocialTelegram
	default:
		return ""
	}
}

func (ece *EnhancedContactExtractor) extractTitle(text string) string {
	// Extract title from text
	titlePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(ceo|cto|cfo|coo|founder|co-founder|president|vice president|vp|director|manager|lead|senior|junior)`),
	}

	for _, pattern := range titlePatterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) >= 2 {
			return strings.Title(strings.ToLower(matches[1]))
		}
	}

	return ""
}

func (ece *EnhancedContactExtractor) extractDayOfWeek(text string) string {
	text = strings.ToLower(text)
	switch {
	case strings.Contains(text, "monday") || strings.Contains(text, "mon"):
		return "Monday"
	case strings.Contains(text, "tuesday") || strings.Contains(text, "tue"):
		return "Tuesday"
	case strings.Contains(text, "wednesday") || strings.Contains(text, "wed"):
		return "Wednesday"
	case strings.Contains(text, "thursday") || strings.Contains(text, "thu"):
		return "Thursday"
	case strings.Contains(text, "friday") || strings.Contains(text, "fri"):
		return "Friday"
	case strings.Contains(text, "saturday") || strings.Contains(text, "sat"):
		return "Saturday"
	case strings.Contains(text, "sunday") || strings.Contains(text, "sun"):
		return "Sunday"
	default:
		return ""
	}
}

func (ece *EnhancedContactExtractor) extractOpenTime(hours string) string {
	// Extract open time from hours string
	timePattern := regexp.MustCompile(`(\d{1,2}:\d{2}\s*(?:am|pm)?)`)
	matches := timePattern.FindAllString(hours, -1)
	if len(matches) >= 1 {
		return matches[0]
	}
	return ""
}

func (ece *EnhancedContactExtractor) extractCloseTime(hours string) string {
	// Extract close time from hours string
	timePattern := regexp.MustCompile(`(\d{1,2}:\d{2}\s*(?:am|pm)?)`)
	matches := timePattern.FindAllString(hours, -1)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func (ece *EnhancedContactExtractor) extractLocationName(text string) string {
	// Extract location name from text
	locationPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)office[:\s]*([^,\n]+)`),
		regexp.MustCompile(`(?i)headquarters[:\s]*([^,\n]+)`),
		regexp.MustCompile(`(?i)location[:\s]*([^,\n]+)`),
	}

	for _, pattern := range locationPatterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) >= 2 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

func (ece *EnhancedContactExtractor) extractCity(address string) string {
	// Placeholder for city extraction
	// In a real implementation, this would use address parsing
	return ""
}

func (ece *EnhancedContactExtractor) extractState(address string) string {
	// Placeholder for state extraction
	// In a real implementation, this would use address parsing
	return ""
}

func (ece *EnhancedContactExtractor) extractCountry(address string) string {
	// Placeholder for country extraction
	// In a real implementation, this would use address parsing
	return ""
}

// Utility functions for checking duplicates
func (ece *EnhancedContactExtractor) containsEmail(emails []string, email string) bool {
	for _, e := range emails {
		if strings.EqualFold(e, email) {
			return true
		}
	}
	return false
}

func (ece *EnhancedContactExtractor) containsPhone(phones []string, phone string) bool {
	for _, p := range phones {
		if strings.EqualFold(p, phone) {
			return true
		}
	}
	return false
}

func (ece *EnhancedContactExtractor) containsAddress(addresses []string, address string) bool {
	for _, a := range addresses {
		if strings.EqualFold(a, address) {
			return true
		}
	}
	return false
}

// calculateOverallConfidence calculates the overall confidence score
func (ece *EnhancedContactExtractor) calculateOverallConfidence(result *EnhancedContactInfo) {
	var scores []float64
	var weights []float64

	// Calculate average confidence for each category
	if len(result.EmailConfidence) > 0 {
		avgConfidence := ece.calculateAverageConfidence(result.EmailConfidence)
		scores = append(scores, avgConfidence)
		weights = append(weights, 0.2) // 20% weight
	}

	if len(result.PhoneConfidence) > 0 {
		avgConfidence := ece.calculateAverageConfidence(result.PhoneConfidence)
		scores = append(scores, avgConfidence)
		weights = append(weights, 0.2) // 20% weight
	}

	if len(result.AddressConfidence) > 0 {
		avgConfidence := ece.calculateAverageConfidence(result.AddressConfidence)
		scores = append(scores, avgConfidence)
		weights = append(weights, 0.15) // 15% weight
	}

	if len(result.SocialMediaConfidence) > 0 {
		avgConfidence := ece.calculateAverageConfidence(result.SocialMediaConfidence)
		scores = append(scores, avgConfidence)
		weights = append(weights, 0.15) // 15% weight
	}

	if len(result.TeamMemberConfidence) > 0 {
		avgConfidence := ece.calculateAverageConfidence(result.TeamMemberConfidence)
		scores = append(scores, avgConfidence)
		weights = append(weights, 0.15) // 15% weight
	}

	if len(result.BusinessHoursConfidence) > 0 {
		avgConfidence := ece.calculateAverageConfidence(result.BusinessHoursConfidence)
		scores = append(scores, avgConfidence)
		weights = append(weights, 0.1) // 10% weight
	}

	if len(result.LocationConfidence) > 0 {
		avgConfidence := ece.calculateAverageConfidence(result.LocationConfidence)
		scores = append(scores, avgConfidence)
		weights = append(weights, 0.05) // 5% weight
	}

	// Calculate weighted average
	if len(scores) > 0 {
		totalWeight := 0.0
		weightedSum := 0.0

		for i, score := range scores {
			weight := weights[i]
			weightedSum += score * weight
			totalWeight += weight
		}

		if totalWeight > 0 {
			result.OverallConfidence = weightedSum / totalWeight
		}
	} else {
		result.OverallConfidence = 0.0
	}
}

// calculateAverageConfidence calculates the average confidence for a map of confidences
func (ece *EnhancedContactExtractor) calculateAverageConfidence(confidences map[string]float64) float64 {
	if len(confidences) == 0 {
		return 0.0
	}

	total := 0.0
	for _, confidence := range confidences {
		total += confidence
	}

	return total / float64(len(confidences))
}

// validateResults validates the extracted results
func (ece *EnhancedContactExtractor) validateResults(result *EnhancedContactInfo) error {
	// Validate confidence scores
	if result.OverallConfidence < 0 || result.OverallConfidence > 1 {
		return fmt.Errorf("overall confidence score %f is out of range [0,1]", result.OverallConfidence)
	}

	// Validate that we have at least some contact information
	totalContactInfo := len(result.Emails) + len(result.Phones) + len(result.Addresses) +
		len(result.SocialMediaAccounts) + len(result.TeamMembers) + len(result.BusinessHours) + len(result.Locations)

	if totalContactInfo == 0 {
		ece.logger.Warn("no contact information detected", map[string]interface{}{
			"total_contact_info": totalContactInfo,
		})
	}

	return nil
}
