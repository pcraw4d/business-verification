package enrichment

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Helper functions for data standardization and suggestions
func (gv *GeographicValidator) standardizeAddress(address string) string {
	if address == "" {
		return ""
	}

	// Clean up address formatting
	address = strings.TrimSpace(address)

	// Use regex for more precise replacements to avoid conflicts
	patterns := []struct {
		regex   *regexp.Regexp
		replace string
	}{
		{regexp.MustCompile(`\b(St|St\.)\b`), "Street"},
		{regexp.MustCompile(`\b(Ave|Ave\.)\b`), "Avenue"},
		{regexp.MustCompile(`\b(Rd|Rd\.)\b`), "Road"},
		{regexp.MustCompile(`\b(Blvd|Blvd\.)\b`), "Boulevard"},
		{regexp.MustCompile(`\b(Dr|Dr\.)\b`), "Drive"},
		{regexp.MustCompile(`\b(Ln|Ln\.)\b`), "Lane"},
		{regexp.MustCompile(`\b(Ct|Ct\.)\b`), "Court"},
		{regexp.MustCompile(`\b(Pl|Pl\.)\b`), "Place"},
	}

	for _, pattern := range patterns {
		address = pattern.regex.ReplaceAllString(address, pattern.replace)
	}

	return address
}

func (gv *GeographicValidator) standardizeCity(city string) string {
	if city == "" {
		return ""
	}

	// Capitalize first letter of each word
	words := strings.Fields(strings.ToLower(city))
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}

	return strings.Join(words, " ")
}

func (gv *GeographicValidator) standardizeState(state, countryCode string) string {
	if state == "" {
		return ""
	}

	// For US states, standardize to full names
	if countryCode == "US" {
		stateMapping := map[string]string{
			"al": "Alabama", "ak": "Alaska", "az": "Arizona", "ar": "Arkansas",
			"ca": "California", "co": "Colorado", "ct": "Connecticut", "de": "Delaware",
			"fl": "Florida", "ga": "Georgia", "hi": "Hawaii", "id": "Idaho",
			"il": "Illinois", "in": "Indiana", "ia": "Iowa", "ks": "Kansas",
			"ky": "Kentucky", "la": "Louisiana", "me": "Maine", "md": "Maryland",
			"ma": "Massachusetts", "mi": "Michigan", "mn": "Minnesota", "ms": "Mississippi",
			"mo": "Missouri", "mt": "Montana", "ne": "Nebraska", "nv": "Nevada",
			"nh": "New Hampshire", "nj": "New Jersey", "nm": "New Mexico", "ny": "New York",
			"nc": "North Carolina", "nd": "North Dakota", "oh": "Ohio", "ok": "Oklahoma",
			"or": "Oregon", "pa": "Pennsylvania", "ri": "Rhode Island", "sc": "South Carolina",
			"sd": "South Dakota", "tn": "Tennessee", "tx": "Texas", "ut": "Utah",
			"vt": "Vermont", "va": "Virginia", "wa": "Washington", "wv": "West Virginia",
			"wi": "Wisconsin", "wy": "Wyoming",
		}

		lowerState := strings.ToLower(state)
		if fullName, exists := stateMapping[lowerState]; exists {
			return fullName
		}
	}

	return gv.standardizeCity(state) // Apply same capitalization rules
}

func (gv *GeographicValidator) standardizePostalCode(postalCode, countryCode string) string {
	if postalCode == "" {
		return ""
	}

	switch countryCode {
	case "US":
		// US ZIP codes: 12345 or 12345-6789
		if matched, _ := regexp.MatchString(`^\d{5}$`, postalCode); matched {
			return postalCode
		}
		if matched, _ := regexp.MatchString(`^\d{5}-\d{4}$`, postalCode); matched {
			return postalCode
		}
		// Try to format as ZIP+4
		digits := regexp.MustCompile(`\d`).FindAllString(postalCode, -1)
		if len(digits) == 9 {
			return strings.Join(digits[:5], "") + "-" + strings.Join(digits[5:], "")
		}
		if len(digits) == 5 {
			return strings.Join(digits, "")
		}

	case "CA":
		// Canadian postal codes: K1A 0A6
		// First convert to uppercase, then remove non-alphanumeric
		upper := strings.ToUpper(postalCode)
		cleaned := regexp.MustCompile(`[^A-Z0-9]`).ReplaceAllString(upper, "")
		if len(cleaned) == 6 {
			return cleaned[:3] + " " + cleaned[3:]
		}
		// If input already has space, preserve it
		if strings.Contains(postalCode, " ") {
			return strings.ToUpper(postalCode)
		}
		return strings.ToUpper(postalCode)

	case "UK", "GB":
		// UK postal codes: SW1A 1AA
		return strings.ToUpper(postalCode)
	}

	return strings.ToUpper(postalCode)
}

func (gv *GeographicValidator) getRegionForCountry(country string) string {
	for region, countries := range gv.config.ValidRegions {
		for _, c := range countries {
			if strings.EqualFold(c, country) {
				return region
			}
		}
	}
	return ""
}

func (gv *GeographicValidator) isValidRegion(region string) bool {
	_, exists := gv.config.ValidRegions[region]
	return exists
}

func (gv *GeographicValidator) suggestCountry(country string) string {
	if country == "" {
		return ""
	}

	normalized := strings.ToLower(strings.TrimSpace(country))

	// Check for partial matches in a predictable order
	longestMatch := ""
	longestMatchResult := ""
	for key, standardized := range gv.config.CountryMappings {
		if strings.Contains(normalized, key) && len(key) > len(longestMatch) {
			longestMatch = key
			longestMatchResult = standardized
		}
	}

	if longestMatchResult != "" {
		return longestMatchResult
	}

	// Check for common typos or variations
	suggestions := map[string]string{
		"usa":     "United States",
		"america": "United States",
		"britain": "United Kingdom",
		"england": "United Kingdom",
	}

	if suggestion, exists := suggestions[normalized]; exists {
		return suggestion
	}

	return ""
}

func (gv *GeographicValidator) suggestRegion(region string) string {
	if region == "" {
		return ""
	}

	normalized := strings.ToLower(strings.TrimSpace(region))

	// Common region mappings
	suggestions := map[string]string{
		"north america": "North America",
		"na":            "North America",
		"europe":        "Europe",
		"eu":            "Europe",
		"asia":          "Asia",
		"apac":          "Asia Pacific",
		"asia pacific":  "Asia Pacific",
		"latam":         "Latin America",
		"south america": "Latin America",
		"middle east":   "Middle East",
		"mena":          "Middle East",
		"africa":        "Africa",
		"oceania":       "Oceania",
	}

	if suggestion, exists := suggestions[normalized]; exists {
		return suggestion
	}

	return ""
}

func (gv *GeographicValidator) suggestPostalCodeCorrection(postalCode, countryCode string) string {
	if postalCode == "" {
		return ""
	}

	digits := regexp.MustCompile(`\d`).FindAllString(postalCode, -1)

	switch countryCode {
	case "US":
		if len(digits) >= 5 {
			if len(digits) >= 9 {
				return strings.Join(digits[:5], "") + "-" + strings.Join(digits[5:9], "")
			}
			return strings.Join(digits[:5], "")
		}

	case "CA":
		letters := regexp.MustCompile(`[A-Za-z]`).FindAllString(strings.ToUpper(postalCode), -1)
		if len(letters) >= 3 && len(digits) >= 3 {
			return letters[0] + digits[0] + letters[1] + " " + digits[1] + letters[2] + digits[2]
		}
	}

	return postalCode
}

func (gv *GeographicValidator) calculateLocationConfidence(location *StandardizedLocation, errors []ValidationError) float64 {
	// Start with lower confidence if no data provided
	baseConfidence := 0.5
	if location.OriginalAddress != "" || location.OriginalCountry != "" {
		baseConfidence = 1.0
	}

	// Reduce confidence for each error
	for _, err := range errors {
		switch err.Severity {
		case "error":
			baseConfidence -= 0.3
		case "warning":
			baseConfidence -= 0.1
		case "info":
			baseConfidence -= 0.05
		}
	}

	// Bonus for having standardized data
	if location.Country != "" && location.CountryCode != "" {
		baseConfidence += 0.1
	}
	if location.Region != "" {
		baseConfidence += 0.05
	}
	if location.PostalCode != "" {
		baseConfidence += 0.05
	}

	return math.Max(0.0, math.Min(1.0, baseConfidence))
}

func (gv *GeographicValidator) calculateOverallConfidence(geography *StandardizedGeography, errors []ValidationError) float64 {
	if len(geography.Locations) == 0 && len(geography.Countries) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	count := 0

	// Average location confidence
	for _, location := range geography.Locations {
		totalConfidence += location.ConfidenceScore
		count++
	}

	// Factor in error count
	errorPenalty := float64(len(errors)) * 0.02

	if count > 0 {
		avgConfidence := totalConfidence / float64(count)
		return math.Max(0.0, math.Min(1.0, avgConfidence-errorPenalty))
	}

	// Base confidence for country/region data only
	baseConfidence := 0.8
	if len(geography.Countries) > 0 {
		baseConfidence += 0.1
	}
	if len(geography.Regions) > 0 {
		baseConfidence += 0.1
	}

	return math.Max(0.0, math.Min(1.0, baseConfidence-errorPenalty))
}

func (gv *GeographicValidator) hasOnlyWarnings(errors []ValidationError) bool {
	for _, err := range errors {
		if err.Severity == "error" {
			return false
		}
	}
	return true
}

func (gv *GeographicValidator) removeDuplicateStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] && item != "" {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func (gv *GeographicValidator) createValidationSummary(geography *StandardizedGeography, errors []ValidationError) *ValidationSummary {
	totalFields := len(geography.Countries) + len(geography.Locations) + len(geography.ServiceAreas)

	errorsByType := make(map[string]int)
	warningCount := 0
	errorCount := 0

	for _, err := range errors {
		errorsByType[err.Field]++
		switch err.Severity {
		case "error":
			errorCount++
		case "warning":
			warningCount++
		}
	}

	validFields := totalFields - errorCount

	// Calculate overall score
	overallScore := geography.ConfidenceScore

	// Generate recommendations
	recommendations := []string{}
	if errorCount > 0 {
		recommendations = append(recommendations, "Fix validation errors to improve data quality")
	}
	if warningCount > 0 {
		recommendations = append(recommendations, "Review warnings and consider standardizing data")
	}
	if len(geography.Countries) == 0 {
		recommendations = append(recommendations, "Add country information for better geographic context")
	}

	return &ValidationSummary{
		TotalFields:     totalFields,
		ValidFields:     validFields,
		InvalidFields:   errorCount,
		WarningFields:   warningCount,
		ErrorsByType:    errorsByType,
		OverallScore:    overallScore,
		Recommendations: recommendations,
	}
}

// parseCoordinates parses latitude and longitude from a string
func (gv *GeographicValidator) parseCoordinates(coordStr string) (*float64, *float64, error) {
	if coordStr == "" {
		return nil, nil, nil
	}

	// Try to parse coordinates in various formats
	// Format: "lat,lng" or "lat, lng"
	parts := strings.Split(coordStr, ",")
	if len(parts) == 2 {
		latStr := strings.TrimSpace(parts[0])
		lngStr := strings.TrimSpace(parts[1])

		lat, err1 := strconv.ParseFloat(latStr, 64)
		lng, err2 := strconv.ParseFloat(lngStr, 64)

		if err1 == nil && err2 == nil {
			if err := gv.validateCoordinates(lat, lng); err == nil {
				return &lat, &lng, nil
			}
		}
	}

	return nil, nil, fmt.Errorf("invalid coordinate format: %s", coordStr)
}

// getDefaultGeographicValidatorConfig returns default configuration
func getDefaultGeographicValidatorConfig() *GeographicValidatorConfig {
	return &GeographicValidatorConfig{
		CountryMappings: map[string]string{
			"usa":                  "United States",
			"us":                   "United States",
			"united states":        "United States",
			"america":              "United States",
			"uk":                   "United Kingdom",
			"united kingdom":       "United Kingdom",
			"england":              "United Kingdom",
			"britain":              "United Kingdom",
			"canada":               "Canada",
			"ca":                   "Canada",
			"australia":            "Australia",
			"au":                   "Australia",
			"germany":              "Germany",
			"de":                   "Germany",
			"deutschland":          "Germany",
			"france":               "France",
			"fr":                   "France",
			"spain":                "Spain",
			"es":                   "Spain",
			"españa":               "Spain",
			"italy":                "Italy",
			"it":                   "Italy",
			"italia":               "Italy",
			"japan":                "Japan",
			"jp":                   "Japan",
			"china":                "China",
			"cn":                   "China",
			"india":                "India",
			"in":                   "India",
			"brazil":               "Brazil",
			"br":                   "Brazil",
			"brasil":               "Brazil",
			"mexico":               "Mexico",
			"mx":                   "Mexico",
			"méxico":               "Mexico",
			"singapore":            "Singapore",
			"sg":                   "Singapore",
			"hong kong":            "Hong Kong",
			"hk":                   "Hong Kong",
			"dubai":                "UAE",
			"uae":                  "UAE",
			"united arab emirates": "UAE",
		},

		CountryCodeMapping: map[string]string{
			"United States":  "US",
			"United Kingdom": "GB",
			"Canada":         "CA",
			"Australia":      "AU",
			"Germany":        "DE",
			"France":         "FR",
			"Spain":          "ES",
			"Italy":          "IT",
			"Japan":          "JP",
			"China":          "CN",
			"India":          "IN",
			"Brazil":         "BR",
			"Mexico":         "MX",
			"Singapore":      "SG",
			"Hong Kong":      "HK",
			"UAE":            "AE",
		},

		ValidCountries: map[string]bool{
			"United States": true, "United Kingdom": true, "Canada": true,
			"Australia": true, "Germany": true, "France": true, "Spain": true,
			"Italy": true, "Japan": true, "China": true, "India": true,
			"Brazil": true, "Mexico": true, "Singapore": true, "Hong Kong": true,
			"UAE": true,
		},

		ValidRegions: map[string][]string{
			"North America": {"United States", "Canada", "Mexico"},
			"Europe":        {"United Kingdom", "Germany", "France", "Spain", "Italy"},
			"Asia Pacific":  {"Japan", "China", "India", "Singapore", "Hong Kong", "Australia"},
			"Asia":          {"Japan", "China", "India", "Singapore", "Hong Kong"},
			"Latin America": {"Brazil", "Mexico"},
			"Middle East":   {"UAE"},
			"Oceania":       {"Australia"},
		},

		PostalCodePatterns: map[string]string{
			"US": `^\d{5}(-\d{4})?$`,
			"CA": `^[A-Z]\d[A-Z] \d[A-Z]\d$`,
			"GB": `^[A-Z]{1,2}\d[A-Z\d]? \d[A-Z]{2}$`,
			"DE": `^\d{5}$`,
			"FR": `^\d{5}$`,
			"AU": `^\d{4}$`,
			"JP": `^\d{3}-\d{4}$`,
		},

		PhonePatterns: map[string]string{
			"US": `^\+?1?[2-9]\d{9}$`,
			"CA": `^\+?1?[2-9]\d{9}$`,
			"GB": `^\+?44[1-9]\d{8,9}$`,
			"DE": `^\+?49[1-9]\d{6,12}$`,
			"FR": `^\+?33[1-9]\d{8}$`,
			"AU": `^\+?61[2-478]\d{8}$`,
		},

		AddressPatterns: map[string][]string{
			"US": {
				`\d+\s+[\w\s]+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Drive|Dr|Lane|Ln|Court|Ct|Place|Pl|Way)`,
			},
			"CA": {
				`\d+\s+[\w\s]+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Drive|Dr|Lane|Ln|Court|Ct|Place|Pl|Way)`,
			},
		},

		MinLatitude:         -90.0,
		MaxLatitude:         90.0,
		MinLongitude:        -180.0,
		MaxLongitude:        180.0,
		MinConfidenceScore:  0.3,
		MaxValidationErrors: 10,
	}
}
