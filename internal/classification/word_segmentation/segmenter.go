package word_segmentation

import (
	"strings"
	"sync"
)

// Segmenter performs word segmentation on compound domain names
// Uses a hybrid approach: dictionary-based + heuristics
type Segmenter struct {
	dictionary map[string]bool
	cache      map[string][]string
	mutex      sync.RWMutex
}

// NewSegmenter creates a new word segmenter with business dictionary
func NewSegmenter() *Segmenter {
	return &Segmenter{
		dictionary: loadBusinessDictionary(),
		cache:      make(map[string][]string),
	}
}

// Segment segments a domain name into meaningful words
// Example: "thegreenegrape" → ["the", "green", "grape"]
func (s *Segmenter) Segment(domain string) []string {
	if domain == "" {
		return []string{}
	}

	// Normalize: lowercase and remove TLD
	normalized := s.normalizeDomain(domain)
	if normalized == "" {
		return []string{}
	}

	// Check cache
	s.mutex.RLock()
	if cached, exists := s.cache[normalized]; exists {
		s.mutex.RUnlock()
		return cached
	}
	s.mutex.RUnlock()

	// Perform segmentation
	var segments []string

	// First, try dictionary-based segmentation (most accurate)
	segments = s.segmentWithDictionary(normalized)

	// If dictionary fails, fall back to heuristics
	if len(segments) == 0 {
		segments = s.segmentWithHeuristics(normalized)
	}

	// If still no segments, return the whole domain as a single word
	if len(segments) == 0 {
		segments = []string{normalized}
	}

	// Cache result
	s.mutex.Lock()
	s.cache[normalized] = segments
	s.mutex.Unlock()

	return segments
}

// normalizeDomain normalizes a domain name by removing TLD and converting to lowercase
func (s *Segmenter) normalizeDomain(domain string) string {
	// Remove protocol if present
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "www.")

	// Remove port if present
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove path if present
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	// Convert to lowercase
	domain = strings.ToLower(domain)

	// Extract domain name (remove TLD)
	parts := strings.Split(domain, ".")
	if len(parts) == 0 {
		return ""
	}

	// Return the first part (domain name without TLD)
	domainName := parts[0]

	// Remove common separators (they'll be handled in segmentation)
	domainName = strings.ReplaceAll(domainName, "-", "")
	domainName = strings.ReplaceAll(domainName, "_", "")

	return domainName
}

// segmentWithDictionary segments using dictionary lookup (dynamic programming approach)
func (s *Segmenter) segmentWithDictionary(text string) []string {
	if text == "" {
		return []string{}
	}

	// Use dynamic programming to find valid word combinations
	n := len(text)
	if n == 0 {
		return []string{}
	}

	// dp[i] = best segmentation ending at position i (nil if no valid segmentation)
	dp := make([][]string, n+1)
	dp[0] = []string{}

	for i := 1; i <= n; i++ {
		for j := 0; j < i; j++ {
			word := text[j:i]
			// Check if word is in dictionary and we have a valid segmentation up to position j
			if s.dictionary[word] && dp[j] != nil {
				// Found a valid word ending at position i
				candidate := make([]string, len(dp[j]))
				copy(candidate, dp[j])
				candidate = append(candidate, word)

				// Update if this is the first valid segmentation or if it's better (fewer words preferred)
				if dp[i] == nil || len(candidate) < len(dp[i]) {
					dp[i] = candidate
				}
			}
		}
	}

	// If we found a complete segmentation, return it
	if len(dp[n]) > 0 {
		return dp[n]
	}

	return []string{}
}

// segmentWithHeuristics segments using heuristic rules when dictionary fails
func (s *Segmenter) segmentWithHeuristics(text string) []string {
	if text == "" {
		return []string{}
	}

	var segments []string

	// Strategy 1: Split on common word boundaries (vowel-consonant patterns)
	segments = s.splitOnVowelConsonantBoundaries(text)
	if len(segments) > 1 {
		return segments
	}

	// Strategy 2: Split on repeated consonants (e.g., "bookkeeper" → "book", "keeper")
	segments = s.splitOnRepeatedConsonants(text)
	if len(segments) > 1 {
		return segments
	}

	// Strategy 3: Split on common prefixes/suffixes
	segments = s.splitOnPrefixesSuffixes(text)
	if len(segments) > 1 {
		return segments
	}

	// Strategy 4: Minimum word length heuristic (prefer words of 3+ characters)
	segments = s.splitByMinimumLength(text)
	if len(segments) > 1 {
		return segments
	}

	// If all heuristics fail, return the whole text as a single segment
	return []string{text}
}

// splitOnVowelConsonantBoundaries splits on vowel-consonant transitions
// Example: "techstartup" → ["tech", "startup"]
func (s *Segmenter) splitOnVowelConsonantBoundaries(text string) []string {
	if len(text) < 4 {
		return []string{text}
	}

	var segments []string
	var current strings.Builder
	vowels := map[byte]bool{'a': true, 'e': true, 'i': true, 'o': true, 'u': true, 'y': true}

	for i := 0; i < len(text); i++ {
		char := text[i]
		current.WriteByte(char)

		// Check if we should split after this character
		if i < len(text)-1 {
			nextChar := text[i+1]
			isVowel := vowels[char]
			nextIsConsonant := !vowels[nextChar]

			// Split after vowel if next is consonant and we have a minimum word length
			if isVowel && nextIsConsonant && current.Len() >= 3 {
				word := current.String()
				// Check if it's a valid-looking word (has vowels)
				if s.hasVowels(word) {
					segments = append(segments, word)
					current.Reset()
				}
			}
		}
	}

	// Add remaining characters
	if current.Len() > 0 {
		segments = append(segments, current.String())
	}

	if len(segments) > 1 {
		return segments
	}

	return []string{text}
}

// splitOnRepeatedConsonants splits on repeated consonant patterns
// Example: "bookkeeper" → ["book", "keeper"]
func (s *Segmenter) splitOnRepeatedConsonants(text string) []string {
	if len(text) < 6 {
		return []string{text}
	}

	var segments []string
	var current strings.Builder

	for i := 0; i < len(text); i++ {
		char := text[i]
		current.WriteByte(char)

		// Check for repeated consonants (potential word boundary)
		if i < len(text)-2 {
			next1 := text[i+1]
			next2 := text[i+2]

			// Pattern: consonant-consonant-consonant (likely word boundary)
			if !s.isVowel(char) && !s.isVowel(next1) && !s.isVowel(next2) && current.Len() >= 3 {
				word := current.String()
				if s.hasVowels(word) {
					segments = append(segments, word)
					current.Reset()
					i++ // Skip next character as it's part of the next word
					continue
				}
			}
		}
	}

	// Add remaining characters
	if current.Len() > 0 {
		segments = append(segments, current.String())
	}

	if len(segments) > 1 {
		return segments
	}

	return []string{text}
}

// splitOnPrefixesSuffixes splits on common business prefixes and suffixes
func (s *Segmenter) splitOnPrefixesSuffixes(text string) []string {
	prefixes := []string{"tech", "digital", "online", "web", "net", "cyber", "smart", "fast", "quick", "easy", "best", "top", "pro", "super", "mega", "ultra", "max", "mini", "micro", "mega"}
	suffixes := []string{"shop", "store", "market", "mart", "mall", "center", "hub", "zone", "place", "spot", "corp", "inc", "llc", "ltd", "group", "co", "com", "net", "org", "io"}

	var segments []string
	remaining := text

	// Check prefixes
	for _, prefix := range prefixes {
		if strings.HasPrefix(remaining, prefix) && len(remaining) > len(prefix)+2 {
			segments = append(segments, prefix)
			remaining = remaining[len(prefix):]
			break
		}
	}

	// Check suffixes
	for _, suffix := range suffixes {
		if strings.HasSuffix(remaining, suffix) && len(remaining) > len(suffix)+2 {
			segments = append(segments, remaining[:len(remaining)-len(suffix)])
			segments = append(segments, suffix)
			remaining = ""
			break
		}
	}

	// If we found prefix/suffix, add remaining as segment
	if len(segments) > 0 && remaining != "" {
		segments = append(segments, remaining)
	}

	if len(segments) > 1 {
		return segments
	}

	return []string{text}
}

// splitByMinimumLength splits text into words of minimum length (greedy approach)
func (s *Segmenter) splitByMinimumLength(text string) []string {
	if len(text) < 6 {
		return []string{text}
	}

	var segments []string
	minLength := 3

	for i := 0; i < len(text); {
		// Try to find the longest valid word starting at position i
		found := false
		for length := len(text) - i; length >= minLength; length-- {
			if i+length <= len(text) {
				word := text[i : i+length]
				// Check if it looks like a valid word (has vowels, reasonable length)
				if s.hasVowels(word) && length >= minLength {
					segments = append(segments, word)
					i += length
					found = true
					break
				}
			}
		}

		if !found {
			// If no valid word found, take minimum length and continue
			if i+minLength <= len(text) {
				segments = append(segments, text[i:i+minLength])
				i += minLength
			} else {
				segments = append(segments, text[i:])
				break
			}
		}
	}

	if len(segments) > 1 {
		return segments
	}

	return []string{text}
}

// Helper functions

func (s *Segmenter) isVowel(char byte) bool {
	vowels := map[byte]bool{'a': true, 'e': true, 'i': true, 'o': true, 'u': true, 'y': true}
	return vowels[char]
}

func (s *Segmenter) hasVowels(word string) bool {
	for i := 0; i < len(word); i++ {
		if s.isVowel(word[i]) {
			return true
		}
	}
	return false
}

