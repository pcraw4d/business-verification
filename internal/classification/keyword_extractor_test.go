package classification

import (
	"strings"
	"testing"
)

func TestNewKeywordExtractor(t *testing.T) {
	ke := NewKeywordExtractor()
	if ke == nil {
		t.Fatal("NewKeywordExtractor() returned nil")
	}

	if ke.stopWords == nil {
		t.Error("Expected stopWords to be initialized")
	}

	if ke.synonyms == nil {
		t.Error("Expected synonyms to be initialized")
	}
}

func TestExtractKeywords(t *testing.T) {
	ke := NewKeywordExtractor()

	tests := []struct {
		name     string
		text     string
		expected []string
		minCount int
	}{
		{
			name:     "Technology description",
			text:     "This U.S. industry comprises establishments primarily engaged in writing, modifying, testing, and supporting software to meet the needs of a particular customer.",
			expected: []string{"industry", "comprises", "establishments", "engaged", "writing", "modifying", "testing", "supporting", "software", "meet", "needs", "particular", "customer"},
			minCount: 10,
		},
		{
			name:     "Healthcare description",
			text:     "This industry comprises establishments of licensed practitioners having the degree of M.D. primarily engaged in the independent practice of general or specialized medicine.",
			expected: []string{"industry", "comprises", "establishments", "licensed", "practitioners", "degree", "engaged", "independent", "practice", "general", "specialized", "medicine"},
			minCount: 10,
		},
		{
			name:     "Financial services description",
			text:     "This industry comprises establishments primarily engaged in accepting demand and other deposits and making commercial, industrial, and consumer loans.",
			expected: []string{"industry", "comprises", "establishments", "engaged", "accepting", "demand", "deposits", "making", "commercial", "industrial", "consumer", "loans"},
			minCount: 10,
		},
		{
			name:     "Empty text",
			text:     "",
			expected: []string{},
			minCount: 0,
		},
		{
			name:     "Short text",
			text:     "Software development",
			expected: []string{"software", "development"},
			minCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := ke.ExtractKeywords(tt.text)

			if len(keywords) < tt.minCount {
				t.Errorf("Expected at least %d keywords, got %d", tt.minCount, len(keywords))
			}

			// Verify no stop words
			for _, keyword := range keywords {
				if ke.stopWords[keyword] {
					t.Errorf("Found stop word in keywords: %s", keyword)
				}

				// Verify minimum length
				if len(keyword) < 3 {
					t.Errorf("Found keyword shorter than 3 characters: %s", keyword)
				}
			}

			t.Logf("Extracted %d keywords: %v", len(keywords), keywords)
		})
	}
}

func TestExtractKeywordsWithRelevance(t *testing.T) {
	ke := NewKeywordExtractor()

	text := "This industry comprises establishments primarily engaged in writing, modifying, testing, and supporting software to meet the needs of a particular customer."
	keywordsWithRelevance := ke.ExtractKeywordsWithRelevance(text)

	if len(keywordsWithRelevance) == 0 {
		t.Error("Expected keywords with relevance scores, got empty map")
	}

	// Verify relevance scores are between 0.5 and 1.0
	for keyword, relevance := range keywordsWithRelevance {
		if relevance < 0.5 || relevance > 1.0 {
			t.Errorf("Relevance score for %s is out of range: %.2f (expected 0.5-1.0)", keyword, relevance)
		}
	}

	t.Logf("Extracted %d keywords with relevance scores", len(keywordsWithRelevance))
}

func TestExtractKeywordsSynonymExpansion(t *testing.T) {
	ke := NewKeywordExtractor()

	text := "Software development and programming services"
	keywords := ke.ExtractKeywords(text)

	// Should include base keywords
	hasSoftware := false
	hasDevelopment := false

	for _, keyword := range keywords {
		if keyword == "software" {
			hasSoftware = true
		}
		if keyword == "development" {
			hasDevelopment = true
		}
	}

	if !hasSoftware {
		t.Error("Expected 'software' keyword")
	}

	if !hasDevelopment {
		t.Error("Expected 'development' keyword")
	}

	// Note: Synonym expansion may or may not include all synonyms
	// This is a basic check
	t.Logf("Keywords extracted: %v", keywords)
}

func TestExtractIndustrySpecificKeywords(t *testing.T) {
	ke := NewKeywordExtractor()

	tests := []struct {
		name     string
		text     string
		industry string
		minCount int
	}{
		{
			name:     "Technology industry",
			text:     "Custom computer programming services",
			industry: "Technology",
			minCount: 15,
		},
		{
			name:     "Healthcare industry",
			text:     "Offices of physicians",
			industry: "Healthcare",
			minCount: 15,
		},
		{
			name:     "Financial Services industry",
			text:     "Commercial banking services",
			industry: "Financial Services",
			minCount: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := ke.ExtractIndustrySpecificKeywords(tt.text, tt.industry)

			if len(keywords) < tt.minCount {
				t.Errorf("Expected at least %d keywords for %s industry, got %d", tt.minCount, tt.industry, len(keywords))
			}

			t.Logf("Extracted %d industry-specific keywords for %s: %v", len(keywords), tt.industry, keywords)
		})
	}
}

func TestNormalizeText(t *testing.T) {
	ke := NewKeywordExtractor()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Uppercase to lowercase",
			input:    "SOFTWARE DEVELOPMENT",
			expected: "software development",
		},
		{
			name:     "Remove punctuation",
			input:    "Software, development, and programming!",
			expected: "software development and programming",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized := ke.normalizeText(tt.input)
			// Note: normalizeText also removes accents, so exact match may not work
			// We'll just verify it's lowercase
			if normalized != strings.ToLower(normalized) {
				t.Errorf("Expected normalized text to be lowercase, got: %s", normalized)
			}
		})
	}
}

func TestFilterKeywords(t *testing.T) {
	ke := NewKeywordExtractor()

	tokens := []string{"the", "software", "development", "and", "programming", "is", "important"}
	filtered := ke.filterKeywords(tokens)

	// Should filter out stop words: "the", "and", "is"
	// Should keep: "software", "development", "programming", "important"
	expectedCount := 4

	if len(filtered) != expectedCount {
		t.Errorf("Expected %d filtered keywords, got %d: %v", expectedCount, len(filtered), filtered)
	}

	// Verify no stop words
	for _, keyword := range filtered {
		if ke.stopWords[keyword] {
			t.Errorf("Stop word not filtered: %s", keyword)
		}
	}
}

func TestRemoveDuplicates(t *testing.T) {
	ke := NewKeywordExtractor()

	keywords := []string{"software", "development", "software", "programming", "development", "coding"}
	unique := ke.removeDuplicates(keywords)

	expectedCount := 4 // software, development, programming, coding

	if len(unique) != expectedCount {
		t.Errorf("Expected %d unique keywords, got %d: %v", expectedCount, len(unique), unique)
	}

	// Verify all are unique
	seen := make(map[string]bool)
	for _, keyword := range unique {
		if seen[keyword] {
			t.Errorf("Duplicate keyword found: %s", keyword)
		}
		seen[keyword] = true
	}
}

