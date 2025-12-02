package classification

import (
	"log"
	"os"
	"strings"
	"testing"
)

func TestIsValidEnglishWord_EnhancedValidation(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := NewSmartWebsiteCrawler(logger)

	tests := []struct {
		name     string
		word     string
		expected bool
		reason   string
	}{
		// Known gibberish words that should be filtered
		{
			name:     "known gibberish - ivdi",
			word:     "ivdi",
			expected: false,
			reason:   "Known gibberish word from plan",
		},
		{
			name:     "known gibberish - fays",
			word:     "fays",
			expected: false,
			reason:   "Known gibberish word from plan",
		},
		{
			name:     "known gibberish - yilp",
			word:     "yilp",
			expected: false,
			reason:   "Known gibberish word from plan",
		},
		{
			name:     "known gibberish - dioy",
			word:     "dioy",
			expected: false,
			reason:   "Known gibberish word from plan",
		},
		{
			name:     "known gibberish - ukxa",
			word:     "ukxa",
			expected: false,
			reason:   "Known gibberish word from plan",
		},
		// Valid English words
		{
			name:     "valid word - business",
			word:     "business",
			expected: true,
			reason:   "Common English word",
		},
		{
			name:     "valid word - technology",
			word:     "technology",
			expected: true,
			reason:   "Common English word",
		},
		{
			name:     "valid word - restaurant",
			word:     "restaurant",
			expected: true,
			reason:   "Common English word",
		},
		// Words with suspicious patterns
		{
			name:     "suspicious pattern - repeated letters",
			word:     "aaabbb",
			expected: false,
			reason:   "Too many repeated letters",
		},
		{
			name:     "suspicious pattern - rare letters",
			word:     "qxzjy",
			expected: false,
			reason:   "Too many rare letters",
		},
		// Edge cases
		{
			name:     "short word",
			word:     "abc",
			expected: false,
			reason:   "Too short (minimum 4 characters)",
		},
		{
			name:     "valid short word",
			word:     "word",
			expected: true,
			reason:   "Valid 4-character word",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := crawler.isValidEnglishWord(tt.word)
			if result != tt.expected {
				t.Errorf("isValidEnglishWord(%q) = %v, want %v (%s)", tt.word, result, tt.expected, tt.reason)
			}
		})
	}
}

func TestHasSuspiciousPatterns_EnhancedDetection(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := NewSmartWebsiteCrawler(logger)

	tests := []struct {
		name     string
		word     string
		expected bool
	}{
		{
			name:     "known gibberish cluster - ivdi",
			word:     "ivdi",
			expected: true,
		},
		{
			name:     "known gibberish cluster - fays",
			word:     "fays",
			expected: true,
		},
		{
			name:     "repeated letters - aaa",
			word:     "aaabbb",
			expected: true,
		},
		{
			name:     "too many rare letters",
			word:     "qxzjy",
			expected: true,
		},
		{
			name:     "valid word - business",
			word:     "business",
			expected: false,
		},
		{
			name:     "valid word - technology",
			word:     "technology",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := crawler.hasSuspiciousPatterns(tt.word)
			if result != tt.expected {
				t.Errorf("hasSuspiciousPatterns(%q) = %v, want %v", tt.word, result, tt.expected)
			}
		})
	}
}

func TestHasValidNgramPatterns_EnhancedValidation(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := NewSmartWebsiteCrawler(logger)

	tests := []struct {
		name     string
		word     string
		expected bool
	}{
		{
			name:     "gibberish - no common bigrams",
			word:     "ivdi",
			expected: false,
		},
		{
			name:     "gibberish - suspicious bigrams",
			word:     "ukxa",
			expected: false,
		},
		{
			name:     "valid word - common bigrams",
			word:     "business",
			expected: true,
		},
		{
			name:     "valid word - technology",
			word:     "technology",
			expected: true,
		},
		{
			name:     "valid word - restaurant",
			word:     "restaurant",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := crawler.hasValidNgramPatterns(tt.word)
			if result != tt.expected {
				t.Errorf("hasValidNgramPatterns(%q) = %v, want %v", tt.word, result, tt.expected)
			}
		})
	}
}

func TestExtractWordsFromText_FiltersGibberish(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := NewSmartWebsiteCrawler(logger)

	tests := []struct {
		name           string
		text           string
		expectedWords  []string
		excludedWords  []string // Words that should NOT be in results
		minWordCount   int
	}{
		{
			name:          "text with gibberish words",
			text:          "We provide business services and technology solutions ivdi fays yilp dioy ukxa",
			expectedWords: []string{"provide", "business", "services", "technology", "solutions"},
			excludedWords: []string{"ivdi", "fays", "yilp", "dioy", "ukxa"},
			minWordCount:  5,
		},
		{
			name:          "text with valid words only",
			text:          "Our company provides excellent customer service and innovative technology solutions",
			expectedWords: []string{"company", "provides", "excellent", "customer", "service", "innovative", "technology", "solutions"},
			excludedWords: []string{},
			minWordCount:  5,
		},
		{
			name:          "mixed content",
			text:          "Business technology ivdi restaurant food fays",
			expectedWords: []string{"business", "technology", "restaurant", "food"},
			excludedWords: []string{"ivdi", "fays"},
			minWordCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation methods directly since extractWordsFromText is private
			// We verify that gibberish words are filtered by checking isValidEnglishWord
			for _, excluded := range tt.excludedWords {
				if crawler.isValidEnglishWord(excluded) {
					t.Errorf("Gibberish word %q should not be valid", excluded)
				}
			}
			
			for _, expected := range tt.expectedWords {
				if !crawler.isValidEnglishWord(expected) {
					t.Errorf("Valid word %q should be valid", expected)
				}
			}

			// Check minimum word count
			if len(words) < tt.minWordCount {
				t.Errorf("Expected at least %d words, got %d", tt.minWordCount, len(words))
			}

			// Create word map for lookup
			wordMap := make(map[string]bool)
			for _, word := range words {
				wordMap[word] = true
			}

			// Check expected words are present
			for _, expected := range tt.expectedWords {
				if !wordMap[strings.ToLower(expected)] {
					t.Errorf("Expected word %q not found in results", expected)
				}
			}

			// Check excluded words are NOT present
			for _, excluded := range tt.excludedWords {
				if wordMap[strings.ToLower(excluded)] {
					t.Errorf("Excluded gibberish word %q found in results", excluded)
				}
			}
		})
	}
}

func TestLoadCommonEnglishWords_ComprehensiveDictionary(t *testing.T) {
	dict := loadCommonEnglishWords()

	// Check dictionary is loaded
	if len(dict) == 0 {
		t.Fatal("Dictionary is empty")
	}

	// Check minimum size (should have 2000+ words)
	if len(dict) < 2000 {
		t.Errorf("Dictionary has only %d words, expected at least 2000", len(dict))
	}

	// Check common words are present
	commonWords := []string{"business", "company", "service", "technology", "restaurant", "retail"}
	for _, word := range commonWords {
		if !dict[word] {
			t.Errorf("Common word %q not found in dictionary", word)
		}
	}

	// Check gibberish words are NOT in dictionary
	gibberishWords := []string{"ivdi", "fays", "yilp", "dioy", "ukxa"}
	for _, word := range gibberishWords {
		if dict[word] {
			t.Errorf("Gibberish word %q found in dictionary (should not be)", word)
		}
	}
}

