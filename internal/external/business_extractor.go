package external

import (
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// BusinessExtractor extracts business information from website content
type BusinessExtractor struct {
	logger *zap.Logger
}

// NewBusinessExtractor creates a new business information extractor
func NewBusinessExtractor(logger *zap.Logger) *BusinessExtractor {
	return &BusinessExtractor{
		logger: logger,
	}
}

// BusinessInfo represents extracted business information
type BusinessInfo struct {
	Name            string            `json:"name"`
	LegalName       string            `json:"legal_name,omitempty"`
	Address         Address           `json:"address"`
	Phone           []string          `json:"phone"`
	Email           []string          `json:"email"`
	Website         string            `json:"website"`
	SocialMedia     map[string]string `json:"social_media"`
	BusinessHours   []BusinessHours   `json:"business_hours,omitempty"`
	Services        []string          `json:"services,omitempty"`
	Industry        string            `json:"industry,omitempty"`
	Founded         string            `json:"founded,omitempty"`
	TeamMembers     []TeamMember      `json:"team_members,omitempty"`
	ContactInfo     []ContactInfo     `json:"contact_info,omitempty"`
	Confidence      float64           `json:"confidence"`
	ExtractionDate  string            `json:"extraction_date"`
}

// Address represents a business address
type Address struct {
	Street     string `json:"street,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country,omitempty"`
	Full       string `json:"full,omitempty"`
}

// BusinessHours represents business operating hours
type BusinessHours struct {
	Day     string `json:"day"`
	Open    string `json:"open"`
	Close   string `json:"close"`
	Closed  bool   `json:"closed"`
}

// TeamMember represents a team member or employee
type TeamMember struct {
	Name        string `json:"name"`
	Title       string `json:"title,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	LinkedIn    string `json:"linkedin,omitempty"`
	Bio         string `json:"bio,omitempty"`
}

// ContactInfo represents contact information
type ContactInfo struct {
	Type    string `json:"type"` // phone, email, address, etc.
	Value   string `json:"value"`
	Label   string `json:"label,omitempty"`
	Primary bool   `json:"primary"`
}

// ExtractBusinessInfo extracts business information from parsed HTML content
func (e *BusinessExtractor) ExtractBusinessInfo(parsedContent *ParsedContent) (*BusinessInfo, error) {
	info := &BusinessInfo{
		SocialMedia: make(map[string]string),
		Phone:       []string{},
		Email:       []string{},
		Services:    []string{},
		TeamMembers: []TeamMember{},
		ContactInfo: []ContactInfo{},
	}

	// Extract business name
	info.Name = e.extractBusinessName(parsedContent)
	
	// Extract address
	info.Address = e.extractAddress(parsedContent)
	
	// Extract phone numbers
	info.Phone = e.extractPhoneNumbers(parsedContent)
	
	// Extract email addresses
	info.Email = e.extractEmailAddresses(parsedContent)
	
	// Extract website
	info.Website = e.extractWebsite(parsedContent)
	
	// Extract social media
	info.SocialMedia = e.extractSocialMedia(parsedContent)
	
	// Extract business hours
	info.BusinessHours = e.extractBusinessHours(parsedContent)
	
	// Extract services
	info.Services = e.extractServices(parsedContent)
	
	// Extract industry
	info.Industry = e.extractIndustry(parsedContent)
	
	// Extract founded year
	info.Founded = e.extractFoundedYear(parsedContent)
	
	// Extract team members
	info.TeamMembers = e.extractTeamMembers(parsedContent)
	
	// Extract contact information
	info.ContactInfo = e.extractContactInfo(parsedContent)
	
	// Calculate confidence score
	info.Confidence = e.calculateConfidence(info)
	
	// Set extraction date
	info.ExtractionDate = time.Now().Format("2006-01-02T15:04:05Z")

	return info, nil
}

// extractBusinessName extracts the business name from content
func (e *BusinessExtractor) extractBusinessName(content *ParsedContent) string {
	// Try structured data first
	if content.Structured != nil && content.Structured.BusinessName != "" {
		return content.Structured.BusinessName
	}

	// Try title
	if content.Title != "" {
		// Clean title and extract business name
		title := strings.TrimSpace(content.Title)
		// Remove common suffixes
		suffixes := []string{" - Home", " | Home", " - Welcome", " | Welcome", " - Official Site"}
		for _, suffix := range suffixes {
			title = strings.TrimSuffix(title, suffix)
		}
		return title
	}

	// Try to find business name in text using patterns
	patterns := []string{
		`(?i)(?:about|welcome to|contact)\s+([A-Z][a-zA-Z\s&]+(?:Inc|LLC|Corp|Company|Ltd|Limited))`,
		`(?i)([A-Z][a-zA-Z\s&]+(?:Inc|LLC|Corp|Company|Ltd|Limited))`,
		`(?i)(?:we are|we're)\s+([A-Z][a-zA-Z\s&]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content.Text)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

// extractAddress extracts address information
func (e *BusinessExtractor) extractAddress(content *ParsedContent) Address {
	address := Address{}

	// Try structured data first
	if content.Structured != nil && content.Structured.Address != "" {
		address.Full = content.Structured.Address
		return address
	}

	// Extract address using patterns
	addressPatterns := []string{
		`(?i)(?:address|location|visit us|find us)[:\s]+([0-9]+[^,\n]+(?:street|st|avenue|ave|road|rd|boulevard|blvd|lane|ln)[^,\n]*(?:,|\n)[^,\n]*(?:,|\n)[^,\n]*)`,
		`(?i)([0-9]+[^,\n]+(?:street|st|avenue|ave|road|rd|boulevard|blvd|lane|ln)[^,\n]*(?:,|\n)[^,\n]*(?:,|\n)[^,\n]*)`,
	}

	for _, pattern := range addressPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content.Text)
		if len(matches) > 1 {
			address.Full = strings.TrimSpace(matches[1])
			break
		}
	}

	// Parse address components if full address found
	if address.Full != "" {
		address = e.parseAddressComponents(address.Full)
	}

	return address
}

// parseAddressComponents parses address string into components
func (e *BusinessExtractor) parseAddressComponents(fullAddress string) Address {
	address := Address{Full: fullAddress}

	// Split by commas
	parts := strings.Split(fullAddress, ",")
	if len(parts) >= 3 {
		address.Street = strings.TrimSpace(parts[0])
		address.City = strings.TrimSpace(parts[1])
		
		// Parse state and postal code
		stateZip := strings.TrimSpace(parts[2])
		stateZipPattern := regexp.MustCompile(`([A-Z]{2})\s*(\d{5}(?:-\d{4})?)`)
		matches := stateZipPattern.FindStringSubmatch(stateZip)
		if len(matches) > 2 {
			address.State = matches[1]
			address.PostalCode = matches[2]
		}
	}

	return address
}

// extractPhoneNumbers extracts phone numbers from content
func (e *BusinessExtractor) extractPhoneNumbers(content *ParsedContent) []string {
	var phones []string

	// Try structured data first
	if content.Structured != nil && content.Structured.Phone != "" {
		phones = append(phones, content.Structured.Phone)
	}

	// Extract phone numbers using patterns
	phonePatterns := []string{
		`(?:\+?1[-.\s]?)?\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})`,
		`(?:\+?1[-.\s]?)?([0-9]{3})[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})`,
		`(?i)(?:phone|tel|call)[:\s]+([0-9\-\(\)\s\+]+)`,
	}

	for _, pattern := range phonePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content.Text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				phone := strings.TrimSpace(match[0])
				if !contains(phones, phone) {
					phones = append(phones, phone)
				}
			}
		}
	}

	return phones
}

// extractEmailAddresses extracts email addresses from content
func (e *BusinessExtractor) extractEmailAddresses(content *ParsedContent) []string {
	var emails []string

	// Try structured data first
	if content.Structured != nil && content.Structured.Email != "" {
		emails = append(emails, content.Structured.Email)
	}

	// Extract email addresses using patterns
	emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	matches := emailPattern.FindAllString(content.Text, -1)
	
	for _, match := range matches {
		email := strings.TrimSpace(match)
		if !contains(emails, email) {
			emails = append(emails, email)
		}
	}

	return emails
}

// extractWebsite extracts website URL
func (e *BusinessExtractor) extractWebsite(content *ParsedContent) string {
	// Try structured data first
	if content.Structured != nil && content.Structured.Website != "" {
		return content.Structured.Website
	}

	// Extract from links
	for _, link := range content.Links {
		if strings.HasPrefix(link, "http") && !strings.Contains(link, "facebook") && 
		   !strings.Contains(link, "twitter") && !strings.Contains(link, "linkedin") {
			return link
		}
	}

	return ""
}

// extractSocialMedia extracts social media links
func (e *BusinessExtractor) extractSocialMedia(content *ParsedContent) map[string]string {
	socialMedia := make(map[string]string)

	// Extract from structured data
	if content.Structured != nil {
		for _, social := range content.Structured.SocialMedia {
			if strings.Contains(social, "facebook") {
				socialMedia["facebook"] = social
			} else if strings.Contains(social, "twitter") {
				socialMedia["twitter"] = social
			} else if strings.Contains(social, "linkedin") {
				socialMedia["linkedin"] = social
			} else if strings.Contains(social, "instagram") {
				socialMedia["instagram"] = social
			}
		}
	}

	// Extract from links
	for _, link := range content.Links {
		if strings.Contains(link, "facebook.com") {
			socialMedia["facebook"] = link
		} else if strings.Contains(link, "twitter.com") {
			socialMedia["twitter"] = link
		} else if strings.Contains(link, "linkedin.com") {
			socialMedia["linkedin"] = link
		} else if strings.Contains(link, "instagram.com") {
			socialMedia["instagram"] = link
		}
	}

	return socialMedia
}

// extractBusinessHours extracts business hours
func (e *BusinessExtractor) extractBusinessHours(content *ParsedContent) []BusinessHours {
	var hours []BusinessHours

	// Look for business hours patterns
	hourPatterns := []string{
		`(?i)(monday|tuesday|wednesday|thursday|friday|saturday|sunday)[:\s]+([0-9:]+\s*(?:am|pm)?\s*[-–]\s*[0-9:]+\s*(?:am|pm)?)`,
		`(?i)(mon|tue|wed|thu|fri|sat|sun)[:\s]+([0-9:]+\s*(?:am|pm)?\s*[-–]\s*[0-9:]+\s*(?:am|pm)?)`,
	}

	for _, pattern := range hourPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content.Text, -1)
		for _, match := range matches {
			if len(match) > 2 {
				day := strings.Title(strings.ToLower(match[1]))
				timeRange := strings.TrimSpace(match[2])
				
				hours = append(hours, BusinessHours{
					Day:   day,
					Open:  timeRange,
					Close: "",
				})
			}
		}
	}

	return hours
}

// extractServices extracts services offered
func (e *BusinessExtractor) extractServices(content *ParsedContent) []string {
	var services []string

	// Look for services patterns
	servicePatterns := []string{
		`(?i)(?:services|what we do|our services|we offer)[:\s]+([^.\n]+)`,
		`(?i)(?:specializing in|specialize in|expert in)[:\s]+([^.\n]+)`,
	}

	for _, pattern := range servicePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content.Text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				serviceText := strings.TrimSpace(match[1])
				// Split by common delimiters
				serviceList := strings.Split(serviceText, ",")
				for _, service := range serviceList {
					service = strings.TrimSpace(service)
					if service != "" && !contains(services, service) {
						services = append(services, service)
					}
				}
			}
		}
	}

	return services
}

// extractIndustry extracts industry information
func (e *BusinessExtractor) extractIndustry(content *ParsedContent) string {
	// Look for industry patterns
	industryPatterns := []string{
		`(?i)(?:industry|sector|field)[:\s]+([^.\n]+)`,
		`(?i)(?:we are a|we're a)\s+([^.\n]+)`,
	}

	for _, pattern := range industryPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content.Text)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

// extractFoundedYear extracts founded year
func (e *BusinessExtractor) extractFoundedYear(content *ParsedContent) string {
	// Look for founded year patterns
	foundedPatterns := []string{
		`(?i)(?:founded|established|since|started)[:\s]+(\d{4})`,
		`(?i)(?:in business since|operating since)[:\s]+(\d{4})`,
	}

	for _, pattern := range foundedPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content.Text)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// extractTeamMembers extracts team member information
func (e *BusinessExtractor) extractTeamMembers(content *ParsedContent) []TeamMember {
	var members []TeamMember

	// Look for team member patterns
	teamPatterns := []string{
		`(?i)([A-Z][a-z]+ [A-Z][a-z]+)[\s\n]+([^,\n]+(?:CEO|CTO|CFO|President|Director|Manager|Lead))`,
		`(?i)([A-Z][a-z]+ [A-Z][a-z]+)[\s\n]+([^,\n]+(?:Founder|Co-founder|Owner))`,
	}

	for _, pattern := range teamPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content.Text, -1)
		for _, match := range matches {
			if len(match) > 2 {
				members = append(members, TeamMember{
					Name:  strings.TrimSpace(match[1]),
					Title: strings.TrimSpace(match[2]),
				})
			}
		}
	}

	return members
}

// extractContactInfo extracts contact information
func (e *BusinessExtractor) extractContactInfo(content *ParsedContent) []ContactInfo {
	var contacts []ContactInfo

	// Extract phone contacts
	for _, phone := range e.extractPhoneNumbers(content) {
		contacts = append(contacts, ContactInfo{
			Type:    "phone",
			Value:   phone,
			Primary: len(contacts) == 0, // First phone is primary
		})
	}

	// Extract email contacts
	for _, email := range e.extractEmailAddresses(content) {
		contacts = append(contacts, ContactInfo{
			Type:    "email",
			Value:   email,
			Primary: len(contacts) == 0, // First email is primary
		})
	}

	// Extract address contact
	if address := e.extractAddress(content); address.Full != "" {
		contacts = append(contacts, ContactInfo{
			Type:    "address",
			Value:   address.Full,
			Primary: true,
		})
	}

	return contacts
}

// calculateConfidence calculates confidence score for extracted information
func (e *BusinessExtractor) calculateConfidence(info *BusinessInfo) float64 {
	score := 0.0
	total := 0.0

	// Business name (30% weight)
	if info.Name != "" {
		score += 30.0
	}
	total += 30.0

	// Address (25% weight)
	if info.Address.Full != "" {
		score += 25.0
	}
	total += 25.0

	// Phone (15% weight)
	if len(info.Phone) > 0 {
		score += 15.0
	}
	total += 15.0

	// Email (10% weight)
	if len(info.Email) > 0 {
		score += 10.0
	}
	total += 10.0

	// Website (10% weight)
	if info.Website != "" {
		score += 10.0
	}
	total += 10.0

	// Social media (5% weight)
	if len(info.SocialMedia) > 0 {
		score += 5.0
	}
	total += 5.0

	// Additional info (5% weight)
	if len(info.Services) > 0 || info.Industry != "" || len(info.TeamMembers) > 0 {
		score += 5.0
	}
	total += 5.0

	if total == 0 {
		return 0.0
	}

	return (score / total) * 100.0
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
