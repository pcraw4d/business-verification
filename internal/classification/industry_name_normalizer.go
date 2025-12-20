package classification

import (
	"strings"
)

// IndustryNameNormalizer provides industry name normalization and mapping
type IndustryNameNormalizer struct {
	// Industry name mappings: expected name -> actual name
	industryMappings map[string]string
}

// NewIndustryNameNormalizer creates a new industry name normalizer
func NewIndustryNameNormalizer() *IndustryNameNormalizer {
	return &IndustryNameNormalizer{
		industryMappings: buildIndustryMappings(),
	}
}

// Normalize is a convenience method that calls NormalizeIndustryName
// Returns just the normalized name
func (n *IndustryNameNormalizer) Normalize(industryName string) string {
	normalized, _ := n.NormalizeIndustryName(industryName)
	return normalized
}

// NormalizeIndustryName normalizes an industry name to match database values
// Returns the normalized name and whether a mapping was found
func (n *IndustryNameNormalizer) NormalizeIndustryName(industryName string) (string, bool) {
	if industryName == "" {
		return "General Business", false
	}

	// First, try exact match (case-insensitive)
	normalized := strings.TrimSpace(industryName)
	if mapped, found := n.industryMappings[strings.ToLower(normalized)]; found {
		return mapped, true
	}

	// Try fuzzy matching for common variations
	normalizedLower := strings.ToLower(normalized)
	
	// Handle common variations
	variations := map[string]string{
		"tech":           "Technology",
		"it":             "Technology",
		"software":       "Technology",
		"health":         "Healthcare",
		"medical":        "Healthcare",
		"finance":        "Financial Services",
		"banking":        "Financial Services",
		"retail":         "Retail",
		"store":          "Retail",
		"manufacturing":  "Manufacturing",
		"construction":   "Construction",
		"transport":      "Transportation",
		"transportation": "Transportation",
		"professional":   "Professional Services",
		"consulting":     "Professional Services",
		"food":           "Food & Beverage",
		"beverage":       "Food & Beverage",
		"catering":       "Catering",
		"casual dining":  "Casual Dining",
		"fast food":      "Fast Food",
		"winery":         "Wineries",
		"wineries":       "Wineries",
		"brewery":        "Breweries",
		"breweries":      "Breweries",
		"food truck":     "Food Trucks",
		"food trucks":    "Food Trucks",
		"bar":            "Bars & Pubs",
		"pub":            "Bars & Pubs",
		"bars":           "Bars & Pubs",
		"pubs":           "Bars & Pubs",
		"restaurants":    "Restaurants",
		"quick service": "Quick Service",
		"general":        "General Business",
		"business":       "General Business",
	}

	// Check if any variation matches
	for key, value := range variations {
		if strings.Contains(normalizedLower, key) {
			return value, true
		}
	}

	// If no mapping found, return original (capitalized properly)
	return capitalizeIndustryName(normalized), false
}

// buildIndustryMappings builds the industry name mapping table
func buildIndustryMappings() map[string]string {
	return map[string]string{
		// Technology variations
		"technology":     "Technology",
		"tech":           "Technology",
		"it":             "Technology",
		"software":       "Technology",
		"information technology": "Technology",
		"tech services":  "Technology",
		"tech service":   "Technology",

		// Healthcare variations
		"healthcare":     "Healthcare",
		"health":         "Healthcare",
		"medical":        "Healthcare",
		"health care":    "Healthcare",

		// Financial Services variations
		"financial services": "Financial Services",
		"finance":           "Financial Services",
		"financial":         "Financial Services",
		"banking":           "Financial Services",
		"financial service": "Financial Services",
		"bank":              "Financial Services",
		"banks":             "Financial Services",
		"investment banking": "Financial Services",
		"investment bank":   "Financial Services",

		// Retail variations
		"retail":         "Retail",
		"retailer":       "Retail",
		"retail shop":    "Retail",
		"retail & commerce": "Retail",
		"retail and commerce": "Retail",
		"commerce":       "Retail",
		"e-commerce":     "Retail",
		"ecommerce":      "Retail",

		// Manufacturing variations
		"manufacturing":  "Manufacturing",
		"manufacturer":   "Manufacturing",
		"production":     "Manufacturing",
		"industrial manufacturing": "Manufacturing",
		"industrial":     "Manufacturing",

		// Construction variations
		"construction":   "Construction",
		"contractor":     "Construction",

		// Transportation variations
		"transportation": "Transportation",
		"transport":      "Transportation",
		"logistics":      "Transportation",
		"shipping":       "Transportation",

		// Professional Services variations
		"professional services": "Professional Services",
		"professional":          "Professional Services",
		"consulting":            "Professional Services",

		// Food & Beverage variations
		"food & beverage": "Food & Beverage",
		"food and beverage": "Food & Beverage",
		"food":            "Food & Beverage",
		"beverage":        "Food & Beverage",
		"dining":          "Food & Beverage",
		"restaurants":     "Food & Beverage",
		"restaurant":      "Food & Beverage",
		"cafes & coffee shops": "Food & Beverage",
		"cafes":           "Food & Beverage",
		"cafe":            "Food & Beverage",
		"coffee shops":    "Food & Beverage",
		"coffee shop":     "Food & Beverage",
		"fast food":       "Food & Beverage",
		"food service":    "Food & Beverage",

		// Catering
		"catering":        "Catering",
		"caterer":         "Catering",

		// Casual Dining
		"casual dining":   "Casual Dining",
		"casual restaurant": "Casual Dining",

		// Fast Food
		"fast-food":       "Fast Food",

		// Wineries
		"winery":          "Wineries",
		"wineries":        "Wineries",
		"wine":            "Wineries",

		// Breweries
		"brewery":         "Breweries",
		"breweries":       "Breweries",
		"beer":            "Breweries",

		// Food Trucks
		"food truck":      "Food Trucks",
		"food trucks":     "Food Trucks",

		// Bars & Pubs
		"bars & pubs":     "Bars & Pubs",
		"bars and pubs":   "Bars & Pubs",
		"bar":             "Bars & Pubs",
		"pub":             "Bars & Pubs",
		"bars":            "Bars & Pubs",
		"pubs":            "Bars & Pubs",

		// Restaurants (mapped to Food & Beverage above)

		// Quick Service
		"quick service":   "Quick Service",
		"quick-service":   "Quick Service",

		// General Business
		"general business": "General Business",
		"general":          "General Business",
		"business":         "General Business",
		"other":            "General Business",
		"unknown":          "General Business",

		// Gambling
		"gambling":        "Gambling",
		"casino":          "Gambling",
		"gaming":          "Gambling",
		"online casino":   "Gambling",
		"casino platform": "Gambling",

		// Entertainment variations
		"entertainment":   "Entertainment",
		"media":           "Entertainment",
		"streaming":       "Entertainment",
		"streaming services": "Entertainment",
		"entertainment services": "Entertainment",
		"media services": "Entertainment",
		"content creation": "Entertainment",
		"video streaming": "Entertainment",
		"music streaming": "Entertainment",

		// Additional common variations
		"it services":     "Technology",
		"software development": "Technology",
		"software company": "Technology",
		"tech company":    "Technology",
		"digital services": "Technology",
		"cloud services":  "Technology",
		"saas":            "Technology",
		"platform":        "Technology",
		"app development": "Technology",
		"web development": "Technology",

		"healthcare services": "Healthcare",
		"hospital services": "Healthcare",
		"clinic services": "Healthcare",
		"medical care":    "Healthcare",
		"patient care":    "Healthcare",

		"financial institution": "Financial Services",
		"banking services": "Financial Services",
		"investment services": "Financial Services",
		"credit services": "Financial Services",
		"insurance services": "Financial Services",
		"financial company": "Financial Services",

		"retail business": "Retail",
		"retail company":  "Retail",
		"online retail":   "Retail",

		"construction company": "Construction",
		"construction contractor": "Construction",
		"general contractor": "Construction",
		"building contractor": "Construction",
		"construction firm": "Construction",

		"consulting services": "Professional Services",
		"consulting firm": "Professional Services",
		"advisory services": "Professional Services",
		"management consulting": "Professional Services",
		"business consulting": "Professional Services",

		"manufacturing company": "Manufacturing",
		"manufacturing services": "Manufacturing",
		"production company": "Manufacturing",

		"transportation services": "Transportation",
		"logistics services": "Transportation",
		"shipping services": "Transportation",
		"delivery services": "Transportation",
		"freight services": "Transportation",
	}
}

// capitalizeIndustryName capitalizes industry name properly
func capitalizeIndustryName(name string) string {
	if name == "" {
		return "General Business"
	}

	// Handle special cases with "&"
	if strings.Contains(name, "&") {
		parts := strings.Split(name, "&")
		result := ""
		for i, part := range parts {
			if i > 0 {
				result += " & "
			}
			result += strings.Title(strings.TrimSpace(part))
		}
		return result
	}

	// Handle multi-word names
	words := strings.Fields(name)
	result := ""
	for i, word := range words {
		if i > 0 {
			result += " "
		}
		result += strings.Title(strings.ToLower(word))
	}

	return result
}

// GetIndustryAliases returns all known aliases for an industry name
func (n *IndustryNameNormalizer) GetIndustryAliases(industryName string) []string {
	aliases := []string{industryName}
	normalizedLower := strings.ToLower(industryName)

	// Find all mappings that point to this industry
	for key, value := range n.industryMappings {
		if strings.EqualFold(value, industryName) && key != normalizedLower {
			aliases = append(aliases, capitalizeIndustryName(key))
		}
	}

	return aliases
}

// AreIndustriesEquivalent checks if two industry names refer to the same industry
// Priority 5.3: Enhanced industry matching for accuracy improvement
func (n *IndustryNameNormalizer) AreIndustriesEquivalent(industry1, industry2 string) bool {
	if industry1 == "" || industry2 == "" {
		return false
	}

	// Normalize both industry names
	normalized1, _ := n.NormalizeIndustryName(industry1)
	normalized2, _ := n.NormalizeIndustryName(industry2)

	// Check if normalized names match (case-insensitive)
	if strings.EqualFold(normalized1, normalized2) {
		return true
	}

	// Check if one is an alias of the other
	aliases1 := n.GetIndustryAliases(normalized1)
	aliases2 := n.GetIndustryAliases(normalized2)

	for _, alias1 := range aliases1 {
		for _, alias2 := range aliases2 {
			if strings.EqualFold(alias1, alias2) {
				return true
			}
		}
	}

	return false
}

