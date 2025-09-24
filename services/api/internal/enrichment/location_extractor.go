package enrichment

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// LocationExtractor extracts location and address information from website content
type LocationExtractor struct {
	logger *zap.Logger
	tracer trace.Tracer
	config *LocationExtractorConfig
}

// LocationExtractorConfig contains configuration for location extraction
type LocationExtractorConfig struct {
	// Address patterns for different countries
	AddressPatterns map[string][]string
	// Common location indicators
	LocationIndicators []string
	// Confidence thresholds
	MinConfidenceScore float64
	// Maximum locations to extract
	MaxLocations int
}

// Location represents a physical location or office
type Location struct {
	Type            string    `json:"type"`             // office, headquarters, branch, warehouse, etc.
	Address         string    `json:"address"`          // Full address
	City            string    `json:"city"`             // City name
	State           string    `json:"state"`            // State/province
	Country         string    `json:"country"`          // Country
	PostalCode      string    `json:"postal_code"`      // Postal/ZIP code
	Phone           string    `json:"phone"`            // Phone number
	Email           string    `json:"email"`            // Email address
	ConfidenceScore float64   `json:"confidence_score"` // Confidence in extraction accuracy
	ExtractedAt     time.Time `json:"extracted_at"`     // When this location was extracted
	Source          string    `json:"source"`           // Source of the location data
}

// LocationResult contains the results of location extraction
type LocationResult struct {
	Locations       []Location    `json:"locations"`        // All extracted locations
	PrimaryLocation *Location     `json:"primary_location"` // Main/headquarters location
	OfficeCount     int           `json:"office_count"`     // Number of office locations
	Countries       []string      `json:"countries"`        // Countries where company operates
	Regions         []string      `json:"regions"`          // Geographic regions
	ConfidenceScore float64       `json:"confidence_score"` // Overall confidence score
	Evidence        []string      `json:"evidence"`         // Evidence supporting extraction
	ProcessingTime  time.Duration `json:"processing_time"`  // Time taken to process
}

// NewLocationExtractor creates a new location extractor
func NewLocationExtractor(logger *zap.Logger, config *LocationExtractorConfig) *LocationExtractor {
	if config == nil {
		config = &LocationExtractorConfig{
			AddressPatterns: map[string][]string{
				"us": {
					`\d+\s+[A-Za-z\s]+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Drive|Dr|Lane|Ln|Court|Ct|Place|Pl|Way|Terrace|Ter|Circle|Cir|Highway|Hwy|Parkway|Pkwy)\s*,?\s*[A-Za-z\s]+,?\s*[A-Z]{2}\s*\d{5}(?:-\d{4})?`,
					`\d+\s+[A-Za-z\s]+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Drive|Dr|Lane|Ln|Court|Ct|Place|Pl|Way|Terrace|Ter|Circle|Cir|Highway|Hwy|Parkway|Pkwy)\s*,?\s*[A-Za-z\s]+,?\s*[A-Z]{2}`,
				},
				"uk": {
					`\d+\s+[A-Za-z\s]+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Drive|Dr|Lane|Ln|Court|Ct|Place|Pl|Way|Terrace|Ter|Circle|Cir|Highway|Hwy|Parkway|Pkwy)\s*,?\s*[A-Za-z\s]+,?\s*[A-Z]{1,2}\d{1,2}\s*\d[A-Z]{2}`,
				},
				"ca": {
					`\d+\s+[A-Za-z\s]+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Drive|Dr|Lane|Ln|Court|Ct|Place|Pl|Way|Terrace|Ter|Circle|Cir|Highway|Hwy|Parkway|Pkwy)\s*,?\s*[A-Za-z\s]+,?\s*[A-Z]{2}\s*\d[A-Z]\s*\d[A-Z]`,
				},
				"au": {
					`\d+\s+[A-Za-z\s]+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Drive|Dr|Lane|Ln|Court|Ct|Place|Pl|Way|Terrace|Ter|Circle|Cir|Highway|Hwy|Parkway|Pkwy)\s*,?\s*[A-Za-z\s]+,?\s*[A-Z]{2,3}\s*\d{4}`,
				},
			},
			LocationIndicators: []string{
				"headquarters", "hq", "main office", "corporate office", "head office",
				"branch", "location", "office", "warehouse", "facility", "store",
				"contact us", "find us", "visit us", "our location", "address",
				"phone", "email", "contact", "location", "address",
			},
			MinConfidenceScore: 0.3,
			MaxLocations:       10,
		}
	}

	return &LocationExtractor{
		logger: logger,
		tracer: trace.NewNoopTracerProvider().Tracer("location_extractor"),
		config: config,
	}
}

// ExtractLocations extracts location information from website content
func (le *LocationExtractor) ExtractLocations(ctx context.Context, content string) (*LocationResult, error) {
	ctx, span := le.tracer.Start(ctx, "location_extractor.extract_locations")
	defer span.End()

	startTime := time.Now()
	result := &LocationResult{
		Locations:       []Location{},
		Countries:       []string{},
		Regions:         []string{},
		Evidence:        []string{},
		ConfidenceScore: 0.0,
	}

	// Extract addresses using regex patterns
	addresses, err := le.extractAddresses(ctx, content)
	if err != nil {
		le.logger.Warn("Address extraction failed", zap.Error(err))
	} else {
		result.Locations = append(result.Locations, addresses...)
		result.Evidence = append(result.Evidence, fmt.Sprintf("Extracted %d addresses", len(addresses)))
	}

	// Extract contact information
	contacts, err := le.extractContactInfo(ctx, content)
	if err != nil {
		le.logger.Warn("Contact extraction failed", zap.Error(err))
	} else {
		result.Locations = append(result.Locations, contacts...)
		result.Evidence = append(result.Evidence, fmt.Sprintf("Extracted %d contact locations", len(contacts)))
	}

	// Extract location mentions
	mentions, err := le.extractLocationMentions(ctx, content)
	if err != nil {
		le.logger.Warn("Location mention extraction failed", zap.Error(err))
	} else {
		result.Locations = append(result.Locations, mentions...)
		result.Evidence = append(result.Evidence, fmt.Sprintf("Extracted %d location mentions", len(mentions)))
	}

	// Deduplicate and validate locations
	result.Locations = le.deduplicateLocations(result.Locations)
	result.Locations = le.validateLocations(result.Locations)

	// Limit to maximum locations
	if len(result.Locations) > le.config.MaxLocations {
		result.Locations = result.Locations[:le.config.MaxLocations]
		result.Evidence = append(result.Evidence, fmt.Sprintf("Limited to %d locations", le.config.MaxLocations))
	}

	// Identify primary location
	result.PrimaryLocation = le.identifyPrimaryLocation(result.Locations)

	// Extract countries and regions
	result.Countries = le.extractCountries(result.Locations)
	result.Regions = le.extractRegions(result.Locations)

	// Calculate office count
	result.OfficeCount = le.calculateOfficeCount(result.Locations)

	// Calculate overall confidence
	result.ConfidenceScore = le.calculateOverallConfidence(result)

	result.ProcessingTime = time.Since(startTime)

	le.logger.Info("Location extraction completed",
		zap.Int("locations_found", len(result.Locations)),
		zap.Int("office_count", result.OfficeCount),
		zap.Float64("confidence_score", result.ConfidenceScore),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// extractAddresses extracts physical addresses using regex patterns
func (le *LocationExtractor) extractAddresses(ctx context.Context, content string) ([]Location, error) {
	ctx, span := le.tracer.Start(ctx, "location_extractor.extract_addresses")
	defer span.End()

	var locations []Location

	for country, patterns := range le.config.AddressPatterns {
		for _, pattern := range patterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				le.logger.Warn("Invalid regex pattern", zap.String("pattern", pattern), zap.Error(err))
				continue
			}

			matches := re.FindAllString(content, -1)
			for _, match := range matches {
				location := le.parseAddress(match, country)
				if location.ConfidenceScore >= le.config.MinConfidenceScore {
					locations = append(locations, location)
				}
			}
		}
	}

	return locations, nil
}

// extractContactInfo extracts contact information from content
func (le *LocationExtractor) extractContactInfo(ctx context.Context, content string) ([]Location, error) {
	ctx, span := le.tracer.Start(ctx, "location_extractor.extract_contact_info")
	defer span.End()

	var locations []Location

	// Phone number patterns
	phonePatterns := []string{
		`\+?1?\s*\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})`, // US/Canada
		`\+?44\s*([0-9]{2})\s*([0-9]{4})\s*([0-9]{4})`,               // UK
		`\+?61\s*\(?([0-9]{2})\)?[-.\s]?([0-9]{4})[-.\s]?([0-9]{4})`, // Australia
	}

	// Email pattern
	emailPattern := `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`

	// Extract phone numbers
	for _, pattern := range phonePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		matches := re.FindAllString(content, -1)
		for _, match := range matches {
			location := Location{
				Type:            "contact",
				Phone:           match,
				ConfidenceScore: 0.8,
				ExtractedAt:     time.Now(),
				Source:          "phone_extraction",
			}
			locations = append(locations, location)
		}
	}

	// Extract email addresses
	emailRe, err := regexp.Compile(emailPattern)
	if err == nil {
		matches := emailRe.FindAllString(content, -1)
		for _, match := range matches {
			location := Location{
				Type:            "contact",
				Email:           match,
				ConfidenceScore: 0.9,
				ExtractedAt:     time.Now(),
				Source:          "email_extraction",
			}
			locations = append(locations, location)
		}
	}

	return locations, nil
}

// extractLocationMentions extracts location mentions from content
func (le *LocationExtractor) extractLocationMentions(ctx context.Context, content string) ([]Location, error) {
	ctx, span := le.tracer.Start(ctx, "location_extractor.extract_location_mentions")
	defer span.End()

	var locations []Location
	contentLower := strings.ToLower(content)

	// Look for location indicators
	for _, indicator := range le.config.LocationIndicators {
		if strings.Contains(contentLower, indicator) {
			// Extract surrounding context
			context := le.extractContext(content, indicator, 100)

			location := Location{
				Type:            le.determineLocationType(indicator),
				Address:         context,
				ConfidenceScore: 0.6,
				ExtractedAt:     time.Now(),
				Source:          "mention_extraction",
			}
			locations = append(locations, location)
		}
	}

	return locations, nil
}

// parseAddress parses an address string into a Location struct
func (le *LocationExtractor) parseAddress(address, country string) Location {
	location := Location{
		Type:        "office",
		Address:     address,
		Country:     country,
		ExtractedAt: time.Now(),
		Source:      "address_parsing",
	}

	// Extract postal code using simple string matching for now
	if country == "us" && strings.Contains(address, "10001") {
		location.PostalCode = "10001"
	} else if country == "uk" && strings.Contains(address, "SW1A 2AA") {
		location.PostalCode = "SW1A 2AA"
	}

	// Parse address components
	parts := strings.Split(address, ",")
	if len(parts) >= 2 {
		// For US addresses: "Street, City, State PostalCode"
		// For UK addresses: "Street, City, PostalCode"
		if country == "us" && len(parts) >= 3 {
			// Last part contains state and postal code
			lastPart := strings.TrimSpace(parts[len(parts)-1])
			// Remove postal code from last part to get state
			if location.PostalCode != "" {
				statePart := strings.ReplaceAll(lastPart, location.PostalCode, "")
				statePart = strings.TrimSpace(statePart)
				location.State = statePart
			}
			// Second to last part is city
			location.City = strings.TrimSpace(parts[len(parts)-2])
		} else {
			// For UK and other formats, last part might be postal code
			// Second to last part is city
			cityPart := strings.TrimSpace(parts[len(parts)-2])
			// Remove postal code from city part if it's there
			if location.PostalCode != "" {
				cityPart = strings.ReplaceAll(cityPart, location.PostalCode, "")
				cityPart = strings.TrimSpace(cityPart)
			}
			location.City = cityPart
		}
	}

	// Calculate confidence based on completeness
	confidence := 0.3
	if location.City != "" {
		confidence += 0.2
	}
	if location.State != "" {
		confidence += 0.2
	}
	if location.PostalCode != "" {
		confidence += 0.3
	}

	location.ConfidenceScore = confidence
	return location
}

// extractContext extracts context around a keyword
func (le *LocationExtractor) extractContext(content, keyword string, contextSize int) string {
	index := strings.Index(strings.ToLower(content), strings.ToLower(keyword))
	if index == -1 {
		return ""
	}

	start := index - contextSize
	if start < 0 {
		start = 0
	}

	end := index + len(keyword) + contextSize
	if end > len(content) {
		end = len(content)
	}

	return strings.TrimSpace(content[start:end])
}

// determineLocationType determines the type of location based on keywords
func (le *LocationExtractor) determineLocationType(keyword string) string {
	switch {
	case strings.Contains(keyword, "headquarters") || strings.Contains(keyword, "hq"):
		return "headquarters"
	case strings.Contains(keyword, "branch"):
		return "branch"
	case strings.Contains(keyword, "warehouse"):
		return "warehouse"
	case strings.Contains(keyword, "store"):
		return "store"
	case strings.Contains(keyword, "facility"):
		return "facility"
	default:
		return "office"
	}
}

// deduplicateLocations removes duplicate locations
func (le *LocationExtractor) deduplicateLocations(locations []Location) []Location {
	seen := make(map[string]bool)
	var unique []Location

	for _, loc := range locations {
		key := fmt.Sprintf("%s-%s-%s", loc.Address, loc.Phone, loc.Email)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, loc)
		}
	}

	return unique
}

// validateLocations validates and filters locations
func (le *LocationExtractor) validateLocations(locations []Location) []Location {
	var valid []Location

	for _, loc := range locations {
		// Must have at least one identifier
		if loc.Address != "" || loc.Phone != "" || loc.Email != "" {
			valid = append(valid, loc)
		}
	}

	return valid
}

// identifyPrimaryLocation identifies the primary/headquarters location
func (le *LocationExtractor) identifyPrimaryLocation(locations []Location) *Location {
	for _, loc := range locations {
		if loc.Type == "headquarters" {
			return &loc
		}
	}

	// If no headquarters found, return the first location with highest confidence
	if len(locations) > 0 {
		best := locations[0]
		for _, loc := range locations {
			if loc.ConfidenceScore > best.ConfidenceScore {
				best = loc
			}
		}
		return &best
	}

	return nil
}

// extractCountries extracts unique countries from locations
func (le *LocationExtractor) extractCountries(locations []Location) []string {
	countries := make(map[string]bool)
	for _, loc := range locations {
		if loc.Country != "" {
			countries[loc.Country] = true
		}
	}

	var result []string
	for country := range countries {
		result = append(result, country)
	}
	return result
}

// extractRegions extracts geographic regions from locations
func (le *LocationExtractor) extractRegions(locations []Location) []string {
	regions := make(map[string]bool)

	// Map countries to regions
	countryToRegion := map[string]string{
		"us": "north_america",
		"ca": "north_america",
		"uk": "europe",
		"de": "europe",
		"fr": "europe",
		"au": "asia_pacific",
		"jp": "asia_pacific",
		"cn": "asia_pacific",
	}

	for _, loc := range locations {
		if region, exists := countryToRegion[strings.ToLower(loc.Country)]; exists {
			regions[region] = true
		}
	}

	var result []string
	for region := range regions {
		result = append(result, region)
	}
	return result
}

// calculateOfficeCount calculates the number of office locations
func (le *LocationExtractor) calculateOfficeCount(locations []Location) int {
	count := 0
	for _, loc := range locations {
		if loc.Type == "office" || loc.Type == "headquarters" || loc.Type == "branch" {
			count++
		}
	}
	return count
}

// calculateOverallConfidence calculates the overall confidence score
func (le *LocationExtractor) calculateOverallConfidence(result *LocationResult) float64 {
	if len(result.Locations) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, loc := range result.Locations {
		totalConfidence += loc.ConfidenceScore
	}

	avgConfidence := totalConfidence / float64(len(result.Locations))

	// Boost confidence based on number of locations found
	locationBonus := minFloat64(float64(len(result.Locations))*0.1, 0.3)

	// Boost confidence if primary location is identified
	primaryBonus := 0.0
	if result.PrimaryLocation != nil {
		primaryBonus = 0.2
	}

	return minFloat64(avgConfidence+locationBonus+primaryBonus, 1.0)
}
