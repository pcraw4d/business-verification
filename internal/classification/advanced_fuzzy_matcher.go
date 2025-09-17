package classification

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

// KeywordIndex represents an index of keywords for fast lookup
type KeywordIndex struct {
	KeywordToIndustries map[string][]IndexKeywordMatch `json:"keyword_to_industries"`
	IndustryToKeywords  map[int][]string               `json:"industry_to_keywords"`
	TotalKeywords       int                            `json:"total_keywords"`
	LastUpdated         time.Time                      `json:"last_updated"`
}

// IndexKeywordMatch represents a keyword match in the index
type IndexKeywordMatch struct {
	Keyword    string  `json:"keyword"`
	IndustryID int     `json:"industry_id"`
	Weight     float64 `json:"weight"`
	MatchType  string  `json:"match_type"`
}

// AdvancedFuzzyMatcher provides sophisticated fuzzy string matching with multiple algorithms
type AdvancedFuzzyMatcher struct {
	logger *log.Logger
	config *AdvancedFuzzyConfig
	cache  map[string][]FuzzyMatch
	mutex  sync.RWMutex
}

// AdvancedFuzzyConfig holds configuration for advanced fuzzy matching
type AdvancedFuzzyConfig struct {
	// Algorithm selection
	EnableLevenshtein    bool `json:"enable_levenshtein"`     // Levenshtein distance
	EnableJaroWinkler    bool `json:"enable_jaro_winkler"`    // Jaro-Winkler similarity
	EnableJaccard        bool `json:"enable_jaccard"`         // Jaccard similarity
	EnableCosine         bool `json:"enable_cosine"`          // Cosine similarity
	EnableSoundex        bool `json:"enable_soundex"`         // Soundex phonetic matching
	EnableMetaphone      bool `json:"enable_metaphone"`       // Metaphone phonetic matching
	EnableSemanticExpand bool `json:"enable_semantic_expand"` // Semantic keyword expansion

	// Performance settings
	MaxCandidates       int     `json:"max_candidates"`       // Maximum candidates to consider
	SimilarityThreshold float64 `json:"similarity_threshold"` // Minimum similarity threshold
	CacheEnabled        bool    `json:"cache_enabled"`        // Enable result caching
	CacheTTL            int64   `json:"cache_ttl"`            // Cache TTL in seconds
	ParallelProcessing  bool    `json:"parallel_processing"`  // Enable parallel processing

	// Algorithm weights (must sum to 1.0)
	LevenshteinWeight float64 `json:"levenshtein_weight"`
	JaroWinklerWeight float64 `json:"jaro_winkler_weight"`
	JaccardWeight     float64 `json:"jaccard_weight"`
	CosineWeight      float64 `json:"cosine_weight"`
	SoundexWeight     float64 `json:"soundex_weight"`
	MetaphoneWeight   float64 `json:"metaphone_weight"`

	// Semantic expansion settings
	MaxSemanticExpansions int     `json:"max_semantic_expansions"` // Maximum semantic expansions
	SemanticThreshold     float64 `json:"semantic_threshold"`      // Semantic similarity threshold
}

// FuzzyMatch represents a fuzzy match result
type FuzzyMatch struct {
	Keyword       string  `json:"keyword"`
	Similarity    float64 `json:"similarity"`
	Algorithm     string  `json:"algorithm"`
	IndustryID    int     `json:"industry_id"`
	Weight        float64 `json:"weight"`
	Confidence    float64 `json:"confidence"`
	SemanticScore float64 `json:"semantic_score,omitempty"`
}

// SemanticExpansion represents semantic keyword expansion
type SemanticExpansion struct {
	OriginalKeyword string            `json:"original_keyword"`
	Expansions      []SemanticKeyword `json:"expansions"`
	Confidence      float64           `json:"confidence"`
}

// SemanticKeyword represents a semantically related keyword
type SemanticKeyword struct {
	Keyword    string  `json:"keyword"`
	Similarity float64 `json:"similarity"`
	Context    string  `json:"context"`
	IndustryID int     `json:"industry_id"`
}

// DefaultAdvancedFuzzyConfig returns the default configuration for advanced fuzzy matching
func DefaultAdvancedFuzzyConfig() *AdvancedFuzzyConfig {
	return &AdvancedFuzzyConfig{
		// Algorithm selection
		EnableLevenshtein:    true,
		EnableJaroWinkler:    true,
		EnableJaccard:        true,
		EnableCosine:         true,
		EnableSoundex:        true,
		EnableMetaphone:      true,
		EnableSemanticExpand: true,

		// Performance settings
		MaxCandidates:       1000,
		SimilarityThreshold: 0.8, // Increased from 0.6 to prevent cross-industry matches
		CacheEnabled:        true,
		CacheTTL:            3600, // 1 hour
		ParallelProcessing:  true,

		// Algorithm weights (must sum to 1.0)
		LevenshteinWeight: 0.25,
		JaroWinklerWeight: 0.20,
		JaccardWeight:     0.15,
		CosineWeight:      0.15,
		SoundexWeight:     0.15,
		MetaphoneWeight:   0.10,

		// Semantic expansion settings
		MaxSemanticExpansions: 10,
		SemanticThreshold:     0.7,
	}
}

// NewAdvancedFuzzyMatcher creates a new advanced fuzzy matcher
func NewAdvancedFuzzyMatcher(logger *log.Logger, config *AdvancedFuzzyConfig) *AdvancedFuzzyMatcher {
	if config == nil {
		config = DefaultAdvancedFuzzyConfig()
	}

	return &AdvancedFuzzyMatcher{
		logger: logger,
		config: config,
		cache:  make(map[string][]FuzzyMatch),
	}
}

// FindFuzzyMatches finds fuzzy matches for a given keyword using multiple algorithms
func (afm *AdvancedFuzzyMatcher) FindFuzzyMatches(
	ctx context.Context,
	inputKeyword string,
	keywordIndex *KeywordIndex,
) ([]FuzzyMatch, error) {
	startTime := time.Now()
	normalizedKeyword := strings.ToLower(strings.TrimSpace(inputKeyword))

	// Check cache first
	if afm.config.CacheEnabled {
		if cached, found := afm.getCachedMatches(normalizedKeyword); found {
			afm.logger.Printf("🎯 Cache hit for fuzzy matching: %s", normalizedKeyword)
			return cached, nil
		}
	}

	afm.logger.Printf("🔍 Starting advanced fuzzy matching for: %s", normalizedKeyword)

	var allMatches []FuzzyMatch
	var wg sync.WaitGroup
	var mutex sync.Mutex

	// Process each keyword in the index
	keywords := make([]string, 0, len(keywordIndex.KeywordToIndustries))
	for keyword := range keywordIndex.KeywordToIndustries {
		keywords = append(keywords, keyword)
	}

	// Limit candidates for performance
	if len(keywords) > afm.config.MaxCandidates {
		// Sort by length similarity for better candidates
		sort.Slice(keywords, func(i, j int) bool {
			lenDiffI := math.Abs(float64(len(keywords[i]) - len(normalizedKeyword)))
			lenDiffJ := math.Abs(float64(len(keywords[j]) - len(normalizedKeyword)))
			return lenDiffI < lenDiffJ
		})
		keywords = keywords[:afm.config.MaxCandidates]
	}

	// Process keywords in parallel if enabled
	if afm.config.ParallelProcessing {
		chunkSize := len(keywords) / 10 // Process in 10 chunks
		if chunkSize < 1 {
			chunkSize = 1
		}

		for i := 0; i < len(keywords); i += chunkSize {
			end := i + chunkSize
			if end > len(keywords) {
				end = len(keywords)
			}

			wg.Add(1)
			go func(keywordChunk []string) {
				defer wg.Done()
				chunkMatches := afm.processKeywordChunk(ctx, normalizedKeyword, keywordChunk, keywordIndex)

				mutex.Lock()
				allMatches = append(allMatches, chunkMatches...)
				mutex.Unlock()
			}(keywords[i:end])
		}
		wg.Wait()
	} else {
		// Sequential processing
		allMatches = afm.processKeywordChunk(ctx, normalizedKeyword, keywords, keywordIndex)
	}

	// Sort matches by combined similarity score
	sort.Slice(allMatches, func(i, j int) bool {
		return allMatches[i].Similarity > allMatches[j].Similarity
	})

	// Filter by threshold and limit results
	var filteredMatches []FuzzyMatch
	for _, match := range allMatches {
		if match.Similarity >= afm.config.SimilarityThreshold {
			filteredMatches = append(filteredMatches, match)
		}
	}

	// Cache results
	if afm.config.CacheEnabled {
		afm.setCachedMatches(normalizedKeyword, filteredMatches)
	}

	processingTime := time.Since(startTime)
	afm.logger.Printf("✅ Advanced fuzzy matching completed for %s: %d matches in %v",
		normalizedKeyword, len(filteredMatches), processingTime)

	return filteredMatches, nil
}

// processKeywordChunk processes a chunk of keywords for fuzzy matching
func (afm *AdvancedFuzzyMatcher) processKeywordChunk(
	ctx context.Context,
	inputKeyword string,
	keywords []string,
	keywordIndex *KeywordIndex,
) []FuzzyMatch {
	var matches []FuzzyMatch

	for _, keyword := range keywords {
		if keyword == inputKeyword {
			continue // Skip exact matches
		}

		// Calculate combined similarity using multiple algorithms
		similarity := afm.calculateCombinedSimilarity(inputKeyword, keyword)

		if similarity >= afm.config.SimilarityThreshold {
			// Get industry matches for this keyword
			if industryMatches, exists := keywordIndex.KeywordToIndustries[keyword]; exists {
				for _, industryMatch := range industryMatches {
					// Additional validation: check if keywords are semantically related
					if afm.areKeywordsSemanticallyRelated(inputKeyword, keyword) {
						// Calculate confidence based on similarity and weight
						confidence := afm.calculateMatchConfidence(similarity, industryMatch.Weight)

						matches = append(matches, FuzzyMatch{
							Keyword:    keyword,
							Similarity: similarity,
							Algorithm:  "combined",
							IndustryID: industryMatch.IndustryID,
							Weight:     industryMatch.Weight,
							Confidence: confidence,
						})
					}
				}
			}
		}
	}

	return matches
}

// calculateCombinedSimilarity calculates similarity using multiple algorithms
func (afm *AdvancedFuzzyMatcher) calculateCombinedSimilarity(s1, s2 string) float64 {
	var totalSimilarity float64
	var totalWeight float64

	// Levenshtein distance
	if afm.config.EnableLevenshtein {
		levSimilarity := afm.calculateLevenshteinSimilarity(s1, s2)
		totalSimilarity += levSimilarity * afm.config.LevenshteinWeight
		totalWeight += afm.config.LevenshteinWeight
	}

	// Jaro-Winkler similarity
	if afm.config.EnableJaroWinkler {
		jaroSimilarity := afm.calculateJaroWinklerSimilarity(s1, s2)
		totalSimilarity += jaroSimilarity * afm.config.JaroWinklerWeight
		totalWeight += afm.config.JaroWinklerWeight
	}

	// Jaccard similarity
	if afm.config.EnableJaccard {
		jaccardSimilarity := afm.calculateJaccardSimilarity(s1, s2)
		totalSimilarity += jaccardSimilarity * afm.config.JaccardWeight
		totalWeight += afm.config.JaccardWeight
	}

	// Cosine similarity
	if afm.config.EnableCosine {
		cosineSimilarity := afm.calculateCosineSimilarity(s1, s2)
		totalSimilarity += cosineSimilarity * afm.config.CosineWeight
		totalWeight += afm.config.CosineWeight
	}

	// Soundex phonetic matching
	if afm.config.EnableSoundex {
		soundexSimilarity := afm.calculateSoundexSimilarity(s1, s2)
		totalSimilarity += soundexSimilarity * afm.config.SoundexWeight
		totalWeight += afm.config.SoundexWeight
	}

	// Metaphone phonetic matching
	if afm.config.EnableMetaphone {
		metaphoneSimilarity := afm.calculateMetaphoneSimilarity(s1, s2)
		totalSimilarity += metaphoneSimilarity * afm.config.MetaphoneWeight
		totalWeight += afm.config.MetaphoneWeight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalSimilarity / totalWeight
}

// calculateLevenshteinSimilarity calculates Levenshtein distance-based similarity
func (afm *AdvancedFuzzyMatcher) calculateLevenshteinSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	len1, len2 := len(s1), len(s2)
	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	distance := afm.levenshteinDistance(s1, s2)
	maxLen := math.Max(float64(len1), float64(len2))

	similarity := 1.0 - (float64(distance) / maxLen)
	return math.Max(0.0, similarity)
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func (afm *AdvancedFuzzyMatcher) levenshteinDistance(s1, s2 string) int {
	len1, len2 := len(s1), len(s2)

	// Create matrix
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	// Initialize first row and column
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = minInt(
				minInt(matrix[i-1][j]+1, matrix[i][j-1]+1), // deletion vs insertion
				matrix[i-1][j-1]+cost,                      // vs substitution
			)
		}
	}

	return matrix[len1][len2]
}

// calculateJaroWinklerSimilarity calculates Jaro-Winkler similarity
func (afm *AdvancedFuzzyMatcher) calculateJaroWinklerSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	len1, len2 := len(s1), len(s2)
	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	// Calculate Jaro similarity
	jaro := afm.calculateJaroSimilarity(s1, s2)

	// Calculate Winkler prefix bonus
	prefixLen := 0
	maxPrefix := minInt(len1, len2)
	if maxPrefix > 4 {
		maxPrefix = 4
	}

	for i := 0; i < maxPrefix; i++ {
		if s1[i] == s2[i] {
			prefixLen++
		} else {
			break
		}
	}

	// Apply Winkler bonus (0.1 * prefix length)
	winklerBonus := 0.1 * float64(prefixLen)
	return jaro + winklerBonus
}

// calculateJaroSimilarity calculates Jaro similarity
func (afm *AdvancedFuzzyMatcher) calculateJaroSimilarity(s1, s2 string) float64 {
	len1, len2 := len(s1), len(s2)
	if len1 == 0 && len2 == 0 {
		return 1.0
	}
	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	// Calculate match window
	matchWindow := (maxInt(len1, len2) / 2) - 1
	if matchWindow < 0 {
		matchWindow = 0
	}

	// Find matches
	s1Matches := make([]bool, len1)
	s2Matches := make([]bool, len2)

	matches := 0
	transpositions := 0

	// Find matches in s1
	for i := 0; i < len1; i++ {
		start := maxInt(0, i-matchWindow)
		end := minInt(len2, i+matchWindow+1)

		for j := start; j < end; j++ {
			if s2Matches[j] || s1[i] != s2[j] {
				continue
			}
			s1Matches[i] = true
			s2Matches[j] = true
			matches++
			break
		}
	}

	if matches == 0 {
		return 0.0
	}

	// Count transpositions
	k := 0
	for i := 0; i < len1; i++ {
		if !s1Matches[i] {
			continue
		}
		for !s2Matches[k] {
			k++
		}
		if s1[i] != s2[k] {
			transpositions++
		}
		k++
	}

	// Calculate Jaro similarity
	jaro := (float64(matches)/float64(len1) +
		float64(matches)/float64(len2) +
		float64(matches-transpositions/2)/float64(matches)) / 3.0

	return jaro
}

// calculateJaccardSimilarity calculates Jaccard similarity using character n-grams
func (afm *AdvancedFuzzyMatcher) calculateJaccardSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	// Create character bigrams
	bigrams1 := afm.createNGrams(s1, 2)
	bigrams2 := afm.createNGrams(s2, 2)

	// Calculate intersection and union
	intersection := 0
	union := len(bigrams1) + len(bigrams2)

	for bigram := range bigrams1 {
		if bigrams2[bigram] {
			intersection++
			union-- // Remove from union count since it's in both
		}
	}

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// createNGrams creates n-grams from a string
func (afm *AdvancedFuzzyMatcher) createNGrams(s string, n int) map[string]bool {
	ngrams := make(map[string]bool)

	for i := 0; i <= len(s)-n; i++ {
		ngram := s[i : i+n]
		ngrams[ngram] = true
	}

	return ngrams
}

// calculateCosineSimilarity calculates cosine similarity using character frequency vectors
func (afm *AdvancedFuzzyMatcher) calculateCosineSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	// Create character frequency vectors
	vec1 := afm.createCharacterVector(s1)
	vec2 := afm.createCharacterVector(s2)

	// Calculate dot product
	dotProduct := 0.0
	for char := range vec1 {
		dotProduct += vec1[char] * vec2[char]
	}

	// Calculate magnitudes
	mag1 := 0.0
	for _, freq := range vec1 {
		mag1 += freq * freq
	}
	mag1 = math.Sqrt(mag1)

	mag2 := 0.0
	for _, freq := range vec2 {
		mag2 += freq * freq
	}
	mag2 = math.Sqrt(mag2)

	if mag1 == 0 || mag2 == 0 {
		return 0.0
	}

	return dotProduct / (mag1 * mag2)
}

// createCharacterVector creates a character frequency vector
func (afm *AdvancedFuzzyMatcher) createCharacterVector(s string) map[rune]float64 {
	vector := make(map[rune]float64)

	for _, char := range s {
		vector[char]++
	}

	// Normalize by string length
	length := float64(len(s))
	for char := range vector {
		vector[char] /= length
	}

	return vector
}

// calculateSoundexSimilarity calculates Soundex-based phonetic similarity
func (afm *AdvancedFuzzyMatcher) calculateSoundexSimilarity(s1, s2 string) float64 {
	soundex1 := afm.calculateSoundex(s1)
	soundex2 := afm.calculateSoundex(s2)

	if soundex1 == soundex2 {
		return 1.0
	}

	// Calculate similarity based on Soundex code differences
	return afm.calculateLevenshteinSimilarity(soundex1, soundex2)
}

// calculateSoundex calculates Soundex code for a string
func (afm *AdvancedFuzzyMatcher) calculateSoundex(s string) string {
	if len(s) == 0 {
		return ""
	}

	// Convert to uppercase
	s = strings.ToUpper(s)

	// Keep first letter
	result := string(s[0])

	// Soundex mapping
	soundexMap := map[rune]string{
		'B': "1", 'F': "1", 'P': "1", 'V': "1",
		'C': "2", 'G': "2", 'J': "2", 'K': "2", 'Q': "2", 'S': "2", 'X': "2", 'Z': "2",
		'D': "3", 'T': "3",
		'L': "4",
		'M': "5", 'N': "5",
		'R': "6",
	}

	lastCode := ""
	for _, char := range s[1:] {
		if code, exists := soundexMap[char]; exists {
			if code != lastCode {
				result += code
				lastCode = code
			}
		} else {
			lastCode = ""
		}
	}

	// Pad with zeros and truncate to 4 characters
	for len(result) < 4 {
		result += "0"
	}

	return result[:4]
}

// calculateMetaphoneSimilarity calculates Metaphone-based phonetic similarity
func (afm *AdvancedFuzzyMatcher) calculateMetaphoneSimilarity(s1, s2 string) float64 {
	metaphone1 := afm.calculateMetaphone(s1)
	metaphone2 := afm.calculateMetaphone(s2)

	if metaphone1 == metaphone2 {
		return 1.0
	}

	// Calculate similarity based on Metaphone code differences
	return afm.calculateLevenshteinSimilarity(metaphone1, metaphone2)
}

// calculateMetaphone calculates Metaphone code for a string (simplified version)
func (afm *AdvancedFuzzyMatcher) calculateMetaphone(s string) string {
	if len(s) == 0 {
		return ""
	}

	// Convert to uppercase and remove non-alphabetic characters
	s = strings.ToUpper(s)
	var clean strings.Builder
	for _, char := range s {
		if unicode.IsLetter(char) {
			clean.WriteRune(char)
		}
	}
	s = clean.String()

	if len(s) == 0 {
		return ""
	}

	// Simplified Metaphone rules
	result := strings.Builder{}

	for i, char := range s {
		switch char {
		case 'A', 'E', 'I', 'O', 'U':
			if i == 0 {
				result.WriteRune(char)
			}
		case 'B':
			result.WriteRune('B')
		case 'C':
			if i+1 < len(s) && s[i+1] == 'H' {
				result.WriteRune('X')
				i++ // Skip next character
			} else if i+1 < len(s) && (s[i+1] == 'I' || s[i+1] == 'E' || s[i+1] == 'Y') {
				result.WriteRune('S')
			} else {
				result.WriteRune('K')
			}
		case 'D':
			if i+1 < len(s) && s[i+1] == 'G' && i+2 < len(s) && (s[i+2] == 'E' || s[i+2] == 'I' || s[i+2] == 'Y') {
				result.WriteRune('J')
				i += 2 // Skip next two characters
			} else {
				result.WriteRune('T')
			}
		case 'F':
			result.WriteRune('F')
		case 'G':
			if i+1 < len(s) && s[i+1] == 'H' && i+2 < len(s) && !isVowel(s[i+2]) {
				// Silent GH
			} else if i+1 < len(s) && (s[i+1] == 'I' || s[i+1] == 'E' || s[i+1] == 'Y') {
				result.WriteRune('J')
			} else {
				result.WriteRune('K')
			}
		case 'H':
			if i == 0 || isVowel(s[i-1]) {
				result.WriteRune('H')
			}
		case 'J':
			result.WriteRune('J')
		case 'K':
			if i == 0 || s[i-1] != 'C' {
				result.WriteRune('K')
			}
		case 'L':
			result.WriteRune('L')
		case 'M':
			result.WriteRune('M')
		case 'N':
			result.WriteRune('N')
		case 'P':
			if i+1 < len(s) && s[i+1] == 'H' {
				result.WriteRune('F')
				i++ // Skip next character
			} else {
				result.WriteRune('P')
			}
		case 'Q':
			result.WriteRune('K')
		case 'R':
			result.WriteRune('R')
		case 'S':
			if i+1 < len(s) && s[i+1] == 'H' {
				result.WriteRune('X')
				i++ // Skip next character
			} else {
				result.WriteRune('S')
			}
		case 'T':
			if i+1 < len(s) && s[i+1] == 'H' {
				result.WriteRune('0') // TH sound
				i++                   // Skip next character
			} else {
				result.WriteRune('T')
			}
		case 'V':
			result.WriteRune('F')
		case 'W':
			if i == 0 || isVowel(s[i-1]) {
				result.WriteRune('W')
			}
		case 'X':
			result.WriteRune('K')
			result.WriteRune('S')
		case 'Y':
			if i == 0 || isVowel(s[i-1]) {
				result.WriteRune('Y')
			}
		case 'Z':
			result.WriteRune('S')
		}
	}

	return result.String()
}

// isVowel checks if a character is a vowel
func isVowel(char byte) bool {
	return char == 'A' || char == 'E' || char == 'I' || char == 'O' || char == 'U'
}

// calculateMatchConfidence calculates confidence for a fuzzy match
func (afm *AdvancedFuzzyMatcher) calculateMatchConfidence(similarity, weight float64) float64 {
	// Base confidence from similarity
	baseConfidence := similarity

	// Apply weight multiplier
	weightMultiplier := 0.5 + (weight * 0.5) // Weight between 0.5 and 1.0

	// Calculate final confidence
	confidence := baseConfidence * weightMultiplier

	return math.Max(0.0, math.Min(1.0, confidence))
}

// ExpandSemanticKeywords performs semantic keyword expansion using Supabase
func (afm *AdvancedFuzzyMatcher) ExpandSemanticKeywords(
	ctx context.Context,
	keyword string,
	keywordIndex *KeywordIndex,
) (*SemanticExpansion, error) {
	if !afm.config.EnableSemanticExpand {
		return nil, fmt.Errorf("semantic expansion is disabled")
	}

	afm.logger.Printf("🧠 Starting semantic expansion for: %s", keyword)

	// Find semantically related keywords using fuzzy matching
	fuzzyMatches, err := afm.FindFuzzyMatches(ctx, keyword, keywordIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to find fuzzy matches for semantic expansion: %w", err)
	}

	// Filter and rank semantic keywords
	var semanticKeywords []SemanticKeyword
	for _, match := range fuzzyMatches {
		if match.Similarity >= afm.config.SemanticThreshold {
			semanticKeywords = append(semanticKeywords, SemanticKeyword{
				Keyword:    match.Keyword,
				Similarity: match.Similarity,
				Context:    "semantic_expansion",
				IndustryID: match.IndustryID,
			})
		}
	}

	// Sort by similarity and limit results
	sort.Slice(semanticKeywords, func(i, j int) bool {
		return semanticKeywords[i].Similarity > semanticKeywords[j].Similarity
	})

	if len(semanticKeywords) > afm.config.MaxSemanticExpansions {
		semanticKeywords = semanticKeywords[:afm.config.MaxSemanticExpansions]
	}

	// Calculate overall confidence
	confidence := 0.0
	if len(semanticKeywords) > 0 {
		for _, sk := range semanticKeywords {
			confidence += sk.Similarity
		}
		confidence /= float64(len(semanticKeywords))
	}

	expansion := &SemanticExpansion{
		OriginalKeyword: keyword,
		Expansions:      semanticKeywords,
		Confidence:      confidence,
	}

	afm.logger.Printf("✅ Semantic expansion completed for %s: %d expansions with confidence %.2f",
		keyword, len(semanticKeywords), confidence)

	return expansion, nil
}

// getCachedMatches retrieves cached fuzzy matches
func (afm *AdvancedFuzzyMatcher) getCachedMatches(keyword string) ([]FuzzyMatch, bool) {
	afm.mutex.RLock()
	defer afm.mutex.RUnlock()

	matches, exists := afm.cache[keyword]
	if !exists {
		return nil, false
	}

	// Check if cache entry is still valid
	// For simplicity, we'll assume cache entries are valid
	// In a real implementation, you'd check timestamps
	return matches, true
}

// setCachedMatches stores fuzzy matches in cache
func (afm *AdvancedFuzzyMatcher) setCachedMatches(keyword string, matches []FuzzyMatch) {
	afm.mutex.Lock()
	defer afm.mutex.Unlock()

	afm.cache[keyword] = matches
}

// Utility functions
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// areKeywordsSemanticallyRelated checks if two keywords are semantically related
// to prevent cross-industry matches
func (afm *AdvancedFuzzyMatcher) areKeywordsSemanticallyRelated(keyword1, keyword2 string) bool {
	// Define semantic domains to prevent cross-industry matches
	semanticDomains := map[string][]string{
		"food_restaurant": {
			"restaurant", "food", "dining", "menu", "chef", "cook", "kitchen",
			"meal", "cuisine", "catering", "bakery", "cafe", "bar", "pub",
			"diner", "bistro", "grill", "pizza", "burger", "sandwich",
		},
		"technology_software": {
			"software", "technology", "development", "programming", "api",
			"code", "application", "system", "platform", "database", "server",
			"cloud", "mobile", "web", "digital", "tech", "computer", "data",
		},
		"health_medical": {
			"health", "medical", "doctor", "hospital", "clinic", "pharmacy",
			"medicine", "therapy", "treatment", "care", "wellness", "dental",
		},
		"retail_shopping": {
			"retail", "shop", "store", "market", "sale", "buy", "sell",
			"merchandise", "product", "inventory", "customer", "commerce",
		},
		"finance_banking": {
			"finance", "bank", "banking", "loan", "credit", "investment",
			"money", "financial", "accounting", "tax", "insurance",
		},
	}

	// Check if both keywords belong to the same semantic domain
	for _, domain := range semanticDomains {
		keyword1InDomain := false
		keyword2InDomain := false

		for _, domainKeyword := range domain {
			if strings.Contains(strings.ToLower(keyword1), domainKeyword) ||
				strings.Contains(strings.ToLower(domainKeyword), keyword1) {
				keyword1InDomain = true
			}
			if strings.Contains(strings.ToLower(keyword2), domainKeyword) ||
				strings.Contains(strings.ToLower(domainKeyword), keyword2) {
				keyword2InDomain = true
			}
		}

		// If both keywords are in the same domain, they are semantically related
		if keyword1InDomain && keyword2InDomain {
			return true
		}
	}

	// If no semantic domain match found, check for basic similarity
	// Allow matches only if they share significant character overlap
	commonChars := 0
	minLen := minInt(len(keyword1), len(keyword2))

	for i := 0; i < minLen; i++ {
		if keyword1[i] == keyword2[i] {
			commonChars++
		}
	}

	// Require at least 70% character overlap for non-domain matches
	overlapRatio := float64(commonChars) / float64(minLen)
	return overlapRatio >= 0.7
}
