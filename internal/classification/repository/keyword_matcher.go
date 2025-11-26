package repository

import (
	"math"
	"strings"
	"unicode"
)

// MatchResult represents the result of a keyword matching operation
type MatchResult struct {
	Matched      bool
	MatchType    string // "exact", "synonym", "stem", "fuzzy"
	RelevancePenalty float64 // Multiplier for relevance score (0.0 to 1.0)
}

// KeywordMatcher provides enhanced keyword matching with synonyms, stemming, and fuzzy matching
type KeywordMatcher struct {
	synonymMap map[string][]string
	stemCache  map[string]string
}

// NewKeywordMatcher creates a new keyword matcher with default synonym dictionary
func NewKeywordMatcher() *KeywordMatcher {
	km := &KeywordMatcher{
		synonymMap: make(map[string][]string),
		stemCache:  make(map[string]string),
	}
	km.loadDefaultSynonyms()
	return km
}

// MatchKeyword performs enhanced keyword matching with multiple strategies
func (km *KeywordMatcher) MatchKeyword(searchKeyword, databaseKeyword string) MatchResult {
	searchLower := strings.ToLower(strings.TrimSpace(searchKeyword))
	dbLower := strings.ToLower(strings.TrimSpace(databaseKeyword))

	// Strategy 1: Exact match (highest priority)
	if searchLower == dbLower {
		return MatchResult{
			Matched:         true,
			MatchType:       "exact",
			RelevancePenalty: 1.0, // No penalty for exact match
		}
	}

	// Strategy 2: Synonym matching
	if km.matchSynonym(searchLower, dbLower) {
		return MatchResult{
			Matched:         true,
			MatchType:       "synonym",
			RelevancePenalty: 0.9, // 10% penalty for synonym match
		}
	}

	// Strategy 3: Stemming-based matching
	if km.matchStem(searchLower, dbLower) {
		return MatchResult{
			Matched:         true,
			MatchType:       "stem",
			RelevancePenalty: 0.85, // 15% penalty for stem match
		}
	}

	// Strategy 4: Fuzzy matching (for typos, low weight)
	fuzzyScore := km.fuzzyMatch(searchLower, dbLower)
	if fuzzyScore > 0.8 { // High threshold to prevent false positives
		return MatchResult{
			Matched:         true,
			MatchType:       "fuzzy",
			RelevancePenalty: 0.7 * fuzzyScore, // Penalty based on similarity
		}
	}

	return MatchResult{Matched: false}
}

// matchSynonym checks if two keywords are synonyms
func (km *KeywordMatcher) matchSynonym(searchKeyword, databaseKeyword string) bool {
	// Check if search keyword is a synonym of database keyword
	if synonyms, exists := km.synonymMap[searchKeyword]; exists {
		for _, synonym := range synonyms {
			if synonym == databaseKeyword {
				return true
			}
		}
	}

	// Check if database keyword is a synonym of search keyword
	if synonyms, exists := km.synonymMap[databaseKeyword]; exists {
		for _, synonym := range synonyms {
			if synonym == searchKeyword {
				return true
			}
		}
	}

	return false
}

// matchStem checks if two keywords match after stemming
func (km *KeywordMatcher) matchStem(searchKeyword, databaseKeyword string) bool {
	searchStem := km.stem(searchKeyword)
	dbStem := km.stem(databaseKeyword)
	return searchStem == dbStem && searchStem != "" && len(searchStem) >= 3
}

// stem performs simple Porter-like stemming
func (km *KeywordMatcher) stem(word string) string {
	// Check cache first
	if cached, exists := km.stemCache[word]; exists {
		return cached
	}

	stemmed := word

	// Only stem words longer than 3 characters
	if len(stemmed) <= 3 {
		km.stemCache[word] = stemmed
		return stemmed
	}

	// Remove common suffixes
	suffixes := []struct {
		old, new string
	}{
		{"ing", ""},
		{"ed", ""},
		{"er", ""},
		{"est", ""},
		{"ly", ""},
		{"tion", ""},
		{"sion", ""},
		{"ness", ""},
		{"ment", ""},
		{"able", ""},
		{"ible", ""},
		{"al", ""},
		{"ic", ""},
		{"ful", ""},
		{"less", ""},
		{"s", ""}, // Plural
	}

	for _, suffix := range suffixes {
		if strings.HasSuffix(stemmed, suffix.old) {
			// Check if removing suffix leaves at least 3 characters
			newStem := strings.TrimSuffix(stemmed, suffix.old)
			if len(newStem) >= 3 {
				stemmed = newStem + suffix.new
				break // Only apply one suffix removal
			}
		}
	}

	km.stemCache[word] = stemmed
	return stemmed
}

// fuzzyMatch calculates fuzzy similarity using Levenshtein distance
func (km *KeywordMatcher) fuzzyMatch(s1, s2 string) float64 {
	// Skip fuzzy matching for very short words
	if len(s1) < 3 || len(s2) < 3 {
		return 0.0
	}

	// Skip if length difference is too large
	lenDiff := math.Abs(float64(len(s1) - len(s2)))
	if lenDiff > math.Max(float64(len(s1)), float64(len(s2)))*0.3 {
		return 0.0
	}

	// Calculate Levenshtein distance
	distance := km.levenshteinDistance(s1, s2)
	maxLen := math.Max(float64(len(s1)), float64(len(s2)))
	
	// Convert to similarity score (0.0 to 1.0)
	similarity := 1.0 - (float64(distance) / maxLen)
	return similarity
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func (km *KeywordMatcher) levenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	column := make([]int, len(r1)+1)

	for y := 1; y <= len(r1); y++ {
		column[y] = y
	}

	for x := 1; x <= len(r2); x++ {
		column[0] = x
		lastDiag := x - 1
		for y := 1; y <= len(r1); y++ {
			oldDiag := column[y]
			cost := 0
			if r1[y-1] != r2[x-1] {
				cost = 1
			}
			column[y] = min3(column[y]+1, column[y-1]+1, lastDiag+cost)
			lastDiag = oldDiag
		}
	}

	return column[len(r1)]
}

// min3 returns the minimum of three integers
func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// loadDefaultSynonyms loads a default synonym dictionary
func (km *KeywordMatcher) loadDefaultSynonyms() {
	// Business and commerce synonyms
	km.synonymMap["shop"] = []string{"store", "retail", "retailer", "merchant", "vendor"}
	km.synonymMap["store"] = []string{"shop", "retail", "retailer", "merchant", "vendor"}
	km.synonymMap["retail"] = []string{"shop", "store", "retailer", "merchant", "vendor"}
	km.synonymMap["restaurant"] = []string{"eatery", "diner", "cafe", "bistro", "tavern"}
	km.synonymMap["cafe"] = []string{"restaurant", "coffee shop", "coffeehouse", "bistro"}
	km.synonymMap["food"] = []string{"cuisine", "meal", "dining", "catering"}
	km.synonymMap["beverage"] = []string{"drink", "liquid", "refreshment"}
	km.synonymMap["wine"] = []string{"vino", "grape wine"}
	km.synonymMap["beer"] = []string{"ale", "lager", "brew"}
	
	// Technology synonyms
	km.synonymMap["software"] = []string{"application", "app", "program", "system"}
	km.synonymMap["app"] = []string{"application", "software", "program"}
	km.synonymMap["technology"] = []string{"tech", "IT", "information technology"}
	km.synonymMap["tech"] = []string{"technology", "IT", "information technology"}
	km.synonymMap["digital"] = []string{"electronic", "online", "web", "internet"}
	km.synonymMap["online"] = []string{"digital", "web", "internet", "ecommerce"}
	km.synonymMap["ecommerce"] = []string{"e-commerce", "online commerce", "digital commerce"}
	
	// Healthcare synonyms
	km.synonymMap["medical"] = []string{"health", "healthcare", "clinical", "therapeutic"}
	km.synonymMap["health"] = []string{"medical", "healthcare", "wellness", "wellbeing"}
	km.synonymMap["healthcare"] = []string{"medical", "health", "health services"}
	km.synonymMap["doctor"] = []string{"physician", "MD", "medical doctor", "practitioner"}
	km.synonymMap["clinic"] = []string{"medical center", "health center", "practice"}
	km.synonymMap["hospital"] = []string{"medical center", "health facility", "medical facility"}
	
	// Financial synonyms
	km.synonymMap["bank"] = []string{"banking", "financial institution", "lender"}
	km.synonymMap["financial"] = []string{"finance", "monetary", "economic"}
	km.synonymMap["finance"] = []string{"financial", "banking", "investment"}
	km.synonymMap["credit"] = []string{"loan", "lending", "financing"}
	km.synonymMap["loan"] = []string{"credit", "lending", "financing", "mortgage"}
	km.synonymMap["insurance"] = []string{"coverage", "policy", "protection"}
	
	// Professional services synonyms
	km.synonymMap["consulting"] = []string{"advisory", "consultancy", "consultation"}
	km.synonymMap["legal"] = []string{"law", "attorney", "lawyer", "litigation"}
	km.synonymMap["law"] = []string{"legal", "attorney", "lawyer", "litigation"}
	km.synonymMap["accounting"] = []string{"accountancy", "bookkeeping", "audit"}
	
	// Manufacturing and industrial synonyms
	km.synonymMap["manufacturing"] = []string{"production", "fabrication", "assembly"}
	km.synonymMap["production"] = []string{"manufacturing", "fabrication", "assembly"}
	km.synonymMap["factory"] = []string{"plant", "manufacturing facility", "production facility"}
	
	// Transportation synonyms
	km.synonymMap["transportation"] = []string{"transport", "shipping", "logistics", "delivery"}
	km.synonymMap["shipping"] = []string{"transportation", "delivery", "freight", "logistics"}
	km.synonymMap["delivery"] = []string{"shipping", "transportation", "logistics"}
	
	// Education synonyms
	km.synonymMap["education"] = []string{"learning", "teaching", "training", "instruction"}
	km.synonymMap["school"] = []string{"academy", "institute", "educational institution"}
	km.synonymMap["university"] = []string{"college", "institution", "academic institution"}
	
	// Real estate synonyms
	km.synonymMap["real estate"] = []string{"property", "realty", "real property"}
	km.synonymMap["property"] = []string{"real estate", "realty", "real property"}
	
	// Add reverse mappings for bidirectional lookup
	km.buildReverseSynonymMap()
}

// buildReverseSynonymMap builds reverse mappings for synonyms
func (km *KeywordMatcher) buildReverseSynonymMap() {
	// Create a copy of current map to iterate over
	currentMap := make(map[string][]string)
	for k, v := range km.synonymMap {
		currentMap[k] = v
	}

	// Add reverse mappings
	for word, synonyms := range currentMap {
		for _, synonym := range synonyms {
			// Add word to synonym's list if not already present
			if existing, exists := km.synonymMap[synonym]; exists {
				// Check if word is already in the list
				found := false
				for _, w := range existing {
					if w == word {
						found = true
						break
					}
				}
				if !found {
					km.synonymMap[synonym] = append(existing, word)
				}
			} else {
				km.synonymMap[synonym] = []string{word}
			}
		}
	}
}

// AddSynonym adds a custom synonym mapping
func (km *KeywordMatcher) AddSynonym(word, synonym string) {
	if km.synonymMap[word] == nil {
		km.synonymMap[word] = []string{}
	}
	
	// Check if synonym already exists
	for _, existing := range km.synonymMap[word] {
		if existing == synonym {
			return // Already exists
		}
	}
	
	km.synonymMap[word] = append(km.synonymMap[word], synonym)
	
	// Add reverse mapping
	if km.synonymMap[synonym] == nil {
		km.synonymMap[synonym] = []string{}
	}
	
	// Check if word already exists in synonym's list
	for _, existing := range km.synonymMap[synonym] {
		if existing == word {
			return // Already exists
		}
	}
	
	km.synonymMap[synonym] = append(km.synonymMap[synonym], word)
}

// isStopWord checks if a word is a common stop word
func isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "as": true, "is": true, "was": true,
		"are": true, "were": true, "been": true, "be": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"should": true, "could": true, "may": true, "might": true, "must": true,
		"can": true, "this": true, "that": true, "these": true, "those": true,
	}
	return stopWords[strings.ToLower(word)]
}

// isPunctuation checks if a rune is punctuation
func isPunctuation(r rune) bool {
	return unicode.IsPunct(r) || unicode.IsSymbol(r)
}

