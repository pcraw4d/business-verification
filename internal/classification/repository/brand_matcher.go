package repository

import (
	"log"
	"strings"
)

// BrandMatcher checks if business names match known hotel brands for MCC codes 3000-3831
type BrandMatcher struct {
	logger      *log.Logger
	hotelBrands map[string]bool
}

// NewBrandMatcher creates a new brand matcher with hardcoded hotel brands
func NewBrandMatcher(logger *log.Logger) *BrandMatcher {
	// Hardcoded hotel brands for MCC 3000-3831 (Hotels, Motels, and Resorts)
	hotelBrands := map[string]bool{
		// Hilton brands
		"hilton":           true,
		"hilton hotels":    true,
		"hilton garden inn": true,
		"doubletree":       true,
		"double tree":      true,
		"hampton":          true,
		"hampton inn":      true,
		"embassy suites":   true,
		"homewood suites":  true,
		"home2 suites":     true,
		"tru":              true,
		"waldorf astoria":  true,
		"conrad":           true,
		"canopy":           true,
		"curio":            true,
		"tapestry":         true,
		"motto":            true,
		"signia":           true,
		"lxr":              true,

		// Marriott brands
		"marriott":          true,
		"marriott hotels":   true,
		"courtyard":         true,
		"residence inn":     true,
		"fairfield":         true,
		"fairfield inn":     true,
		"springhill":       true,
		"springhill suites": true,
		"towneplace":       true,
		"towneplace suites": true,
		"aloft":            true,
		"element":           true,
		"moxy":              true,
		"ac hotels":         true,
		"autograph":         true,
		"delta":             true,
		"le meridien":       true,
		"westin":            true,
		"w hotels":          true,
		"st regis":          true,
		"luxury collection": true,
		"ritz carlton":      true,
		"ritz-carlton":      true,
		"edition":           true,
		"four points":       true,
		"protea":            true,
		"design hotels":     true,
		"gaylord":           true,
		"bulgari":           true,

		// Hyatt brands
		"hyatt":           true,
		"hyatt hotels":    true,
		"hyatt regency":    true,
		"grand hyatt":      true,
		"park hyatt":       true,
		"hyatt place":      true,
		"hyatt house":       true,
		"andaz":            true,
		"alila":            true,
		"miraval":          true,
		"destination":      true,
		"joie de vivre":    true,
		"tomorrow":         true,
		"toh":              true,
		"caption":          true,
		"urcove":           true,
		"secrets":          true,
		"dreams":           true,
		"breathless":       true,
		"zoetry":           true,
		"vivid":            true,
		"impressions":      true,
		"sunscape":         true,
		"world of hyatt":   true,

		// IHG brands
		"ihg":                    true,
		"intercontinental":       true,
		"inter continental":      true,
		"holiday inn":            true,
		"holiday inn express":    true,
		"crowne plaza":           true,
		"kimpton":                true,
		"hotel indigo":           true,
		"even":                   true,
		"avid":                  true,
		"staybridge":            true,
		"staybridge suites":     true,
		"candlewood":            true,
		"candlewood suites":     true,
		"voco":                  true,
		"regent":                true,
		"six senses":            true,
		"vignette":              true,

		// Accor brands
		"accor":        true,
		"novotel":      true,
		"ibis":         true,
		"sofitel":      true,
		"mercure":      true,
		"pullman":      true,
		"swissotel":    true,
		"mgallery":     true,
		"grand mercure": true,
		"the sebel":    true,
		"adagio":       true,
		"mantra":       true,
		"peppers":      true,
		"breakfree":    true,
		"quest":        true,
		"apartments":   true,
		"fairmont":     true,
		"raffles":      true,
		"orient express": true,
		"banyan tree":   true,
		"ennismore":     true,
		"21c museum":   true,
		"25hours":      true,
		"delano":       true,
		"mondrian":     true,
		"sbe":          true,
		"sls":          true,
		"hyde":         true,
		"jo&joe":       true,
		"tribe":        true,
		"greet":        true,
		"handwritten":  true,

		// Wyndham brands
		"wyndham":         true,
		"wyndham hotels": true,
		"ramada":          true,
		"days inn":        true,
		"super 8":         true,
		"super8":          true,
		"travelodge":      true,
		"travel lodge":    true,
		"hawthorn":        true,
		"hawthorn suites": true,
		"microtel":        true,
		"wingate":         true,
		"baymont":         true,
		"la quinta":       true,
		"dolce":           true,
		"esplendor":       true,
		"dazzler":         true,
		"trademark":       true,
		"wyndham garden":  true,
		"wyndham grand":   true,
		"wyndham alltra":  true,
		"registry":        true,
		"tryp":            true,
		"vib":             true,
		"americinn":       true,
		"wyndham rewards": true,
	}

	return &BrandMatcher{
		logger:      logger,
		hotelBrands: hotelBrands,
	}
}

// IsHighConfidenceBrandMatch checks if business name matches a known hotel brand
// Only returns true for brands that would be classified in MCC range 3000-3831
func (bm *BrandMatcher) IsHighConfidenceBrandMatch(businessName string) (bool, string, float64) {
	if businessName == "" {
		return false, "", 0.0
	}

	// Normalize business name
	normalized := bm.normalizeBusinessName(businessName)
	if normalized == "" {
		return false, "", 0.0
	}

	// Check for exact match
	if bm.hotelBrands[normalized] {
		bm.logger.Printf("✅ Brand match found: %s (normalized: %s)", businessName, normalized)
		return true, normalized, 0.95
	}

	// Check for partial match (brand name contains normalized or vice versa)
	for brand := range bm.hotelBrands {
		if strings.Contains(normalized, brand) || strings.Contains(brand, normalized) {
			// Additional validation: ensure it's a meaningful match
			if len(brand) >= 4 && len(normalized) >= 4 {
				bm.logger.Printf("✅ Partial brand match found: %s (matched: %s)", businessName, brand)
				return true, brand, 0.85
			}
		}
	}

	return false, "", 0.0
}

// normalizeBusinessName normalizes business name by removing common suffixes and converting to lowercase
func (bm *BrandMatcher) normalizeBusinessName(name string) string {
	// Convert to lowercase
	normalized := strings.ToLower(strings.TrimSpace(name))

	// Remove common suffixes
	suffixes := []string{
		" inc", " inc.", " incorporated",
		" llc", " llc.", " limited liability company",
		" corp", " corp.", " corporation",
		" ltd", " ltd.", " limited",
		" hotels", " hotel", " resorts", " resort",
		" international", " intl", " intl.",
		" group", " companies", " company", " co", " co.",
	}

	for _, suffix := range suffixes {
		if strings.HasSuffix(normalized, suffix) {
			normalized = strings.TrimSuffix(normalized, suffix)
			normalized = strings.TrimSpace(normalized)
		}
	}

	return normalized
}

// GetMCCRangeForBrandMatch returns the MCC range for brand matches (3000-3831)
func (bm *BrandMatcher) GetMCCRangeForBrandMatch() string {
	return "3000-3831"
}

