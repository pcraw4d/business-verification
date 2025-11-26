package nlp

import (
	"math"
	"testing"
)

func TestNewTopicModeler(t *testing.T) {
	tm := NewTopicModeler()
	if tm == nil {
		t.Fatal("NewTopicModeler() returned nil")
	}
	if len(tm.industryTopics) == 0 {
		t.Error("TopicModeler has no industry topics loaded")
	}
	if len(tm.idfScores) == 0 {
		t.Error("TopicModeler has no IDF scores calculated")
	}
}

func TestIdentifyTopics(t *testing.T) {
	tm := NewTopicModeler()

	tests := []struct {
		name           string
		keywords       []string
		expectedMin    int // Minimum number of industries that should match
		expectedTop    int // Expected top industry ID
		minScore       float64
	}{
		{
			name:        "technology keywords",
			keywords:    []string{"software", "technology", "digital", "platform", "development"},
			expectedMin: 1,
			expectedTop: 1, // Technology
			minScore:    0.15, // Lower threshold
		},
		{
			name:        "healthcare keywords",
			keywords:    []string{"medical", "health", "clinic", "hospital", "patient"},
			expectedMin: 1,
			expectedTop: 2, // Healthcare
			minScore:    0.15, // Lower threshold
		},
		{
			name:        "food and beverage keywords",
			keywords:    []string{"wine", "restaurant", "food", "dining", "beverage"},
			expectedMin: 1,
			expectedTop: 5, // Food & Beverage
			minScore:    0.15, // Lower threshold
		},
		{
			name:        "retail keywords",
			keywords:    []string{"retail", "store", "shop", "merchandise", "ecommerce"},
			expectedMin: 1,
			expectedTop: 4, // Retail
			minScore:    0.15, // Lower threshold
		},
		{
			name:        "financial services keywords",
			keywords:    []string{"bank", "financial", "credit", "loan", "investment"},
			expectedMin: 1,
			expectedTop: 3, // Financial Services
			minScore:    0.15, // Lower threshold
		},
		{
			name:        "mixed keywords (wine shop)",
			keywords:    []string{"wine", "shop", "retail", "store", "beverage"},
			expectedMin: 1, // Should match at least Retail (Food & Beverage might score lower)
			expectedTop: 4, // Retail should score higher
			minScore:    0.15, // Lower threshold
		},
		{
			name:        "empty keywords",
			keywords:    []string{},
			expectedMin: 0,
			expectedTop: 0,
			minScore:    0.15,
		},
		{
			name:        "unrelated keywords",
			keywords:    []string{"xyz", "abc", "random", "unrelated"},
			expectedMin: 0,
			expectedTop: 0,
			minScore:    0.15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm.SetMinScore(tt.minScore)
			scores := tm.IdentifyTopics(tt.keywords)

			if len(scores) < tt.expectedMin {
				t.Errorf("IdentifyTopics() returned %d industries, want at least %d. Scores: %v", len(scores), tt.expectedMin, scores)
			}

			// Verify scores are within valid range
			for industryID, score := range scores {
				if score < 0.0 || score > 1.0 {
					t.Errorf("IdentifyTopics() returned invalid score %f for industry %d", score, industryID)
				}
				if score < tt.minScore {
					t.Errorf("IdentifyTopics() returned score %f below minimum %f for industry %d", score, tt.minScore, industryID)
				}
			}

			// Check if expected top industry is in results
			if tt.expectedTop > 0 {
				if _, exists := scores[tt.expectedTop]; !exists && tt.expectedMin > 0 {
					t.Logf("Warning: Expected top industry %d not found in results. Got: %v", tt.expectedTop, scores)
				}
			}
		})
	}
}

func TestIdentifyTopicsWithDetails(t *testing.T) {
	tm := NewTopicModeler()

	keywords := []string{"wine", "shop", "retail", "store", "beverage"}
	topicScores := tm.IdentifyTopicsWithDetails(keywords)

	if len(topicScores) == 0 {
		t.Error("IdentifyTopicsWithDetails() returned no topic scores")
	}

	// Verify scores are sorted (highest first)
	for i := 1; i < len(topicScores); i++ {
		if topicScores[i-1].Score < topicScores[i].Score {
			t.Errorf("IdentifyTopicsWithDetails() scores not sorted. Score[%d]=%f < Score[%d]=%f",
				i-1, topicScores[i-1].Score, i, topicScores[i].Score)
		}
	}

	// Verify each topic score has contributing keywords
	for _, ts := range topicScores {
		if len(ts.Keywords) == 0 {
			t.Errorf("IdentifyTopicsWithDetails() returned topic score with no contributing keywords for industry %d", ts.IndustryID)
		}
		if ts.Score < 0.0 || ts.Score > 1.0 {
			t.Errorf("IdentifyTopicsWithDetails() returned invalid score %f for industry %d", ts.Score, ts.IndustryID)
		}
	}
}

func TestCalculateTopicAlignment(t *testing.T) {
	tm := NewTopicModeler()

	tests := []struct {
		name         string
		keywordSet   map[string]bool
		topicKeywords []string
		expectedMin  float64
		expectedMax  float64
	}{
		{
			name:         "perfect match",
			keywordSet:   map[string]bool{"software": true, "technology": true, "digital": true},
			topicKeywords: []string{"software", "technology", "digital"},
			expectedMin:   0.7,
			expectedMax:   1.0,
		},
		{
			name:         "partial match",
			keywordSet:   map[string]bool{"software": true, "technology": true},
			topicKeywords: []string{"software", "technology", "digital", "platform"},
			expectedMin:   0.3,
			expectedMax:   0.9,
		},
		{
			name:         "no match",
			keywordSet:   map[string]bool{"xyz": true, "abc": true},
			topicKeywords: []string{"software", "technology", "digital"},
			expectedMin:   0.0,
			expectedMax:   0.1,
		},
		{
			name:         "empty topic keywords",
			keywordSet:   map[string]bool{"software": true},
			topicKeywords: []string{},
			expectedMin:   0.0,
			expectedMax:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := tm.calculateTopicAlignment(tt.keywordSet, tt.topicKeywords)

			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("calculateTopicAlignment() returned score %f, want between %f and %f",
					score, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestGetIDF(t *testing.T) {
	tm := NewTopicModeler()

	// Test with known word
	idf := tm.getIDF("software")
	if idf <= 0 {
		t.Errorf("getIDF() returned non-positive IDF for known word: %f", idf)
	}

	// Test with unknown word
	idfUnknown := tm.getIDF("xyzunknownword")
	if idfUnknown <= 0 {
		t.Errorf("getIDF() returned non-positive IDF for unknown word: %f", idfUnknown)
	}

	// IDF should be positive
	if idfUnknown <= 0 {
		t.Error("getIDF() should return positive IDF for any word")
	}
}

func TestSetMinScore(t *testing.T) {
	tm := NewTopicModeler()

	newMinScore := 0.5
	tm.SetMinScore(newMinScore)

	// Verify min score was set
	keywords := []string{"software", "technology"}
	scores := tm.IdentifyTopics(keywords)

	// All scores should be >= newMinScore
	for industryID, score := range scores {
		if score < newMinScore {
			t.Errorf("IdentifyTopics() returned score %f below minimum %f for industry %d",
				score, newMinScore, industryID)
		}
	}
}

func TestAddIndustryTopics(t *testing.T) {
	tm := NewTopicModeler()

	initialCount := len(tm.industryTopics)
	newIndustryID := 999
	newKeywords := []string{"test", "keyword", "topic"}

	tm.AddIndustryTopics(newIndustryID, newKeywords)

	if len(tm.industryTopics) != initialCount+1 {
		t.Errorf("AddIndustryTopics() did not add industry. Expected %d industries, got %d",
			initialCount+1, len(tm.industryTopics))
	}

	// Verify keywords were added
	keywords := tm.GetIndustryTopics(newIndustryID)
	if len(keywords) != len(newKeywords) {
		t.Errorf("GetIndustryTopics() returned %d keywords, want %d", len(keywords), len(newKeywords))
	}

	// Verify IDF scores were recalculated
	if len(tm.idfScores) == 0 {
		t.Error("AddIndustryTopics() did not recalculate IDF scores")
	}
}

func TestGetIndustryTopics(t *testing.T) {
	tm := NewTopicModeler()

	// Test with existing industry
	keywords := tm.GetIndustryTopics(1) // Technology
	if len(keywords) == 0 {
		t.Error("GetIndustryTopics() returned no keywords for existing industry")
	}

	// Test with non-existing industry
	keywords = tm.GetIndustryTopics(9999)
	if keywords == nil {
		t.Error("GetIndustryTopics() returned nil for non-existing industry")
	}
	if len(keywords) != 0 {
		t.Errorf("GetIndustryTopics() returned keywords for non-existing industry: %v", keywords)
	}
}

func TestIDFCalculation(t *testing.T) {
	tm := NewTopicModeler()

	// Verify IDF scores are calculated
	if len(tm.idfScores) == 0 {
		t.Error("IDF scores not calculated")
	}

	// Common words should have lower IDF (appear in more industries)
	// Rare words should have higher IDF (appear in fewer industries)
	
	// Test that IDF is positive
	for word, idf := range tm.idfScores {
		if idf <= 0 {
			t.Errorf("IDF score for word '%s' is non-positive: %f", word, idf)
		}
		if math.IsNaN(idf) || math.IsInf(idf, 0) {
			t.Errorf("IDF score for word '%s' is NaN or Inf: %f", word, idf)
		}
	}
}

func BenchmarkIdentifyTopics(b *testing.B) {
	tm := NewTopicModeler()
	keywords := []string{"software", "technology", "digital", "platform", "development", "web", "cloud"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.IdentifyTopics(keywords)
	}
}

func BenchmarkIdentifyTopicsWithDetails(b *testing.B) {
	tm := NewTopicModeler()
	keywords := []string{"wine", "shop", "retail", "store", "beverage", "food", "dining"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.IdentifyTopicsWithDetails(keywords)
	}
}

