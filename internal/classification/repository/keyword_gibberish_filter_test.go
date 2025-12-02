package repository

import (
	"log"
	"os"
	"testing"
)

func TestFilterGibberishKeywords_RemovesKnownGibberish(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	tests := []struct {
		name           string
		keywords       []string
		expectedOutput []string
		excludedWords  []string
	}{
		{
			name:           "filters known gibberish words",
			keywords:       []string{"business", "ivdi", "technology", "fays", "service", "yilp"},
			expectedOutput: []string{"business", "technology", "service"},
			excludedWords:  []string{"ivdi", "fays", "yilp"},
		},
		{
			name:           "keeps valid multi-word phrases",
			keywords:       []string{"business service", "ivdi", "technology solutions"},
			expectedOutput: []string{"business service", "technology solutions"},
			excludedWords:  []string{"ivdi"},
		},
		{
			name:           "filters suspicious patterns",
			keywords:       []string{"business", "aaabbb", "technology", "qxzjy"},
			expectedOutput: []string{"business", "technology"},
			excludedWords:  []string{"aaabbb", "qxzjy"},
		},
		{
			name:           "all valid words",
			keywords:       []string{"business", "technology", "service", "restaurant"},
			expectedOutput: []string{"business", "technology", "service", "restaurant"},
			excludedWords:  []string{},
		},
		{
			name:           "all gibberish",
			keywords:       []string{"ivdi", "fays", "yilp", "dioy", "ukxa"},
			expectedOutput: []string{},
			excludedWords:  []string{"ivdi", "fays", "yilp", "dioy", "ukxa"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := repo.filterGibberishKeywords(tt.keywords)

			// Create map for lookup
			filteredMap := make(map[string]bool)
			for _, word := range filtered {
				filteredMap[word] = true
			}

			// Check expected words are present
			for _, expected := range tt.expectedOutput {
				if !filteredMap[expected] {
					t.Errorf("Expected word %q not found in filtered results", expected)
				}
			}

			// Check excluded words are NOT present
			for _, excluded := range tt.excludedWords {
				if filteredMap[excluded] {
					t.Errorf("Excluded gibberish word %q found in filtered results", excluded)
				}
			}
		})
	}
}

func TestHasSuspiciousPattern_Repository(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	tests := []struct {
		name     string
		word     string
		expected bool
	}{
		{
			name:     "known gibberish - ivdi",
			word:     "ivdi",
			expected: true,
		},
		{
			name:     "known gibberish - fays",
			word:     "fays",
			expected: true,
		},
		{
			name:     "repeated letters",
			word:     "aaabbb",
			expected: true,
		},
		{
			name:     "too many rare letters",
			word:     "qxzjy",
			expected: true,
		},
		{
			name:     "valid word",
			word:     "business",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repo.hasSuspiciousPattern(tt.word)
			if result != tt.expected {
				t.Errorf("hasSuspiciousPattern(%q) = %v, want %v", tt.word, result, tt.expected)
			}
		})
	}
}

func TestHasValidNgramPattern_Repository(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	commonBigrams := map[string]bool{
		"th": true, "in": true, "er": true, "ed": true, "an": true, "re": true,
		"he": true, "on": true, "en": true, "at": true, "it": true, "is": true,
		"or": true, "ti": true, "as": true, "to": true, "of": true, "al": true,
		"ar": true, "st": true, "ng": true, "le": true, "ou": true, "nt": true,
		"ea": true, "nd": true, "te": true, "es": true, "hi": true,
		"ri": true, "ve": true, "co": true, "de": true, "ra": true, "li": true,
		"se": true, "ne": true, "me": true, "be": true, "we": true, "wa": true,
		"ma": true, "ha": true, "ca": true, "la": true, "pa": true, "ta": true,
		"sa": true, "na": true, "ga": true, "fa": true, "da": true, "ba": true,
	}

	suspiciousBigrams := map[string]bool{
		"iv": true, "vd": true, "di": true, "xa": true, "uk": true, "kx": true,
		"fa": true, "ay": true, "ys": true, "yi": true, "il": true, "lp": true,
		"gu": true, "oi": true, "je": true, "yl": true, "lb": true, "io": true,
		"fv": true, "yz": true, "zx": true, "qw": true, "xc": true, "vb": true,
	}

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
			name:     "valid word - common bigrams",
			word:     "business",
			expected: true,
		},
		{
			name:     "valid word - technology",
			word:     "technology",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repo.hasValidNgramPattern(tt.word, commonBigrams, suspiciousBigrams)
			if result != tt.expected {
				t.Errorf("hasValidNgramPattern(%q) = %v, want %v", tt.word, result, tt.expected)
			}
		})
	}
}

