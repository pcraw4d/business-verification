package classification

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// KeywordExtractor extracts keywords from text descriptions
type KeywordExtractor struct {
	stopWords map[string]bool
	synonyms  map[string][]string
}

// NewKeywordExtractor creates a new keyword extractor
func NewKeywordExtractor() *KeywordExtractor {
	return &KeywordExtractor{
		stopWords: getStopWords(),
		synonyms:  getSynonyms(),
	}
}

// ExtractKeywords extracts keywords from a text description
func (ke *KeywordExtractor) ExtractKeywords(text string) []string {
	if text == "" {
		return []string{}
	}

	// Normalize text
	normalized := ke.normalizeText(text)

	// Tokenize
	tokens := ke.tokenize(normalized)

	// Filter stop words and short words
	keywords := ke.filterKeywords(tokens)

	// Expand with synonyms
	expanded := ke.expandWithSynonyms(keywords)

	// Remove duplicates while preserving order
	unique := ke.removeDuplicates(expanded)

	return unique
}

// ExtractKeywordsWithRelevance extracts keywords with relevance scores
func (ke *KeywordExtractor) ExtractKeywordsWithRelevance(text string) map[string]float64 {
	keywords := ke.ExtractKeywords(text)
	result := make(map[string]float64)

	for i, keyword := range keywords {
		// Higher relevance for earlier keywords (assuming they're more important)
		relevance := 1.0 - (float64(i) * 0.02)
		if relevance < 0.5 {
			relevance = 0.5
		}
		result[keyword] = relevance
	}

	return result
}

// normalizeText normalizes text by removing accents and converting to lowercase
func (ke *KeywordExtractor) normalizeText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove accents
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, _ := transform.String(t, text)

	return normalized
}

// tokenize splits text into tokens
func (ke *KeywordExtractor) tokenize(text string) []string {
	// Remove punctuation and split by whitespace
	re := regexp.MustCompile(`[^\w\s]`)
	cleaned := re.ReplaceAllString(text, " ")

	// Split by whitespace
	tokens := strings.Fields(cleaned)

	return tokens
}

// filterKeywords filters out stop words and short words
func (ke *KeywordExtractor) filterKeywords(tokens []string) []string {
	var keywords []string

	for _, token := range tokens {
		// Skip short words (less than 3 characters)
		if len(token) < 3 {
			continue
		}

		// Skip stop words
		if ke.stopWords[token] {
			continue
		}

		// Skip pure numbers
		if isNumeric(token) {
			continue
		}

		keywords = append(keywords, token)
	}

	return keywords
}

// expandWithSynonyms expands keywords with their synonyms
func (ke *KeywordExtractor) expandWithSynonyms(keywords []string) []string {
	expanded := make([]string, 0, len(keywords)*2)
	seen := make(map[string]bool)

	for _, keyword := range keywords {
		// Add original keyword
		if !seen[keyword] {
			expanded = append(expanded, keyword)
			seen[keyword] = true
		}

		// Add synonyms
		if synonyms, ok := ke.synonyms[keyword]; ok {
			for _, synonym := range synonyms {
				if !seen[synonym] {
					expanded = append(expanded, synonym)
					seen[synonym] = true
				}
			}
		}
	}

	return expanded
}

// removeDuplicates removes duplicate keywords while preserving order
func (ke *KeywordExtractor) removeDuplicates(keywords []string) []string {
	seen := make(map[string]bool)
	var unique []string

	for _, keyword := range keywords {
		if !seen[keyword] {
			seen[keyword] = true
			unique = append(unique, keyword)
		}
	}

	return unique
}

// isNumeric checks if a string is purely numeric
func isNumeric(s string) bool {
	re := regexp.MustCompile(`^\d+$`)
	return re.MatchString(s)
}

// getStopWords returns a map of common stop words
func getStopWords() map[string]bool {
	stopWords := []string{
		"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for",
		"of", "with", "by", "from", "as", "is", "was", "are", "were", "be",
		"been", "being", "have", "has", "had", "do", "does", "did", "will",
		"would", "should", "could", "may", "might", "must", "can", "this",
		"that", "these", "those", "it", "its", "they", "them", "their", "our",
		"your", "my", "his", "her", "he", "she", "we", "you", "i", "me", "us",
		"services", "service", "products", "product", "business", "company",
		"companies", "other", "etc", "including", "excluding", "related", "not",
		"primarily", "engaged", "establishments", "establishment", "comprises",
		"industry", "industries", "u.s.", "us", "providing", "provides",
	}

	result := make(map[string]bool)
	for _, word := range stopWords {
		result[word] = true
	}

	return result
}

// getSynonyms returns a map of keyword synonyms
func getSynonyms() map[string][]string {
	return map[string][]string{
		"software":     {"app", "application", "program", "system"},
		"technology":   {"tech", "it", "information", "digital"},
		"computer":    {"pc", "desktop", "laptop", "machine"},
		"programming": {"coding", "development", "software", "engineering"},
		"development": {"programming", "coding", "engineering", "building"},
		"healthcare":  {"medical", "health", "medicine", "clinical"},
		"medical":     {"healthcare", "health", "clinical", "hospital"},
		"doctor":      {"physician", "practitioner", "clinician", "md"},
		"hospital":    {"medical", "clinic", "healthcare", "facility"},
		"retail":      {"store", "shop", "commerce", "merchant"},
		"store":       {"shop", "retail", "merchant", "outlet"},
		"restaurant":  {"cafe", "dining", "eatery", "food"},
		"cafe":        {"restaurant", "coffee", "dining", "eatery"},
		"banking":     {"bank", "financial", "finance", "institution"},
		"bank":        {"banking", "financial", "institution", "lender"},
		"education":  {"school", "learning", "academic", "training"},
		"school":      {"education", "academic", "learning", "institution"},
		"construction": {"building", "contractor", "development", "building"},
		"manufacturing": {"production", "factory", "industrial", "making"},
		"transportation": {"transport", "shipping", "logistics", "delivery"},
		"real estate":   {"property", "realty", "land", "housing"},
		"professional":  {"business", "corporate", "commercial", "enterprise"},
	}
}

// ExtractIndustrySpecificKeywords extracts industry-specific keywords from descriptions
func (ke *KeywordExtractor) ExtractIndustrySpecificKeywords(text string, industry string) []string {
	baseKeywords := ke.ExtractKeywords(text)

	// Add industry-specific terms
	industryTerms := getIndustryTerms(industry)
	allKeywords := append(baseKeywords, industryTerms...)

	// Remove duplicates
	return ke.removeDuplicates(allKeywords)
}

// getIndustryTerms returns industry-specific terminology
func getIndustryTerms(industry string) []string {
	terms := map[string][]string{
		"Technology": {
			"software", "technology", "tech", "digital", "computer", "programming",
			"development", "coding", "app", "application", "platform", "system",
			"saas", "cloud", "api", "web", "mobile", "ios", "android",
			"devops", "infrastructure", "server", "database", "backend", "frontend",
		},
		"Healthcare": {
			"healthcare", "medical", "health", "hospital", "clinic", "doctor",
			"physician", "patient", "medicine", "pharmaceutical", "drug", "therapy",
			"treatment", "diagnosis", "surgery", "nursing", "wellness", "fitness",
		},
		"Financial Services": {
			"banking", "bank", "financial", "finance", "investment", "credit",
			"loan", "mortgage", "insurance", "securities", "trading", "brokerage",
		},
		"Retail & Commerce": {
			"retail", "store", "shop", "commerce", "sales", "merchandise",
			"products", "ecommerce", "online", "marketplace", "shopping",
		},
		"Food & Beverage": {
			"restaurant", "cafe", "dining", "food", "beverage", "catering",
			"bakery", "coffee", "bar", "pub", "kitchen", "cuisine",
		},
		"Manufacturing": {
			"manufacturing", "production", "factory", "industrial", "making",
			"assembly", "fabrication", "processing", "machinery",
		},
		"Construction": {
			"construction", "building", "contractor", "development", "building",
			"renovation", "remodeling", "architecture", "engineering",
		},
		"Transportation": {
			"transportation", "transport", "shipping", "logistics", "delivery",
			"freight", "trucking", "airline", "railway",
		},
		"Education": {
			"education", "school", "learning", "academic", "training",
			"university", "college", "teaching", "instruction",
		},
		"Hospitality": {
			"hospitality", "hotel", "lodging", "accommodation", "resort",
			"tourism", "travel", "vacation", "hospitality",
		},
		"Professional Services": {
			"professional", "business", "corporate", "commercial", "enterprise",
			"consulting", "advisory", "services", "management",
		},
		"Real Estate": {
			"real estate", "property", "realty", "land", "housing",
			"rental", "leasing", "brokerage", "development",
		},
		"Arts and Entertainment": {
			"arts", "entertainment", "theater", "music", "performance",
			"cultural", "creative", "artistic", "amusement",
		},
	}

	if industryTerms, ok := terms[industry]; ok {
		return industryTerms
	}

	return []string{}
}

