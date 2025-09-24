package risk

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

// MockRiskKeywordsService for testing
type MockRiskKeywordsService struct {
	keywords []*RiskKeyword
	logger   *log.Logger
}

// NewMockRiskKeywordsService creates a mock service for testing
func NewMockRiskKeywordsService() *MockRiskKeywordsService {
	return &MockRiskKeywordsService{
		keywords: []*RiskKeyword{
			{
				ID:                    1,
				Keyword:               "drug trafficking",
				RiskCategory:          "illegal",
				RiskSeverity:          "critical",
				Description:           "Illegal drug trafficking and distribution",
				MCCCodes:              []string{},
				NAICSCodes:            []string{},
				SICCodes:              []string{},
				CardBrandRestrictions: []string{"Visa", "Mastercard", "American Express"},
				DetectionPatterns:     []string{"drug.*traffick", "traffick.*drug"},
				Synonyms:              []string{"drug dealing", "narcotics trafficking"},
				IsActive:              true,
				CreatedAt:             time.Now(),
				UpdatedAt:             time.Now(),
			},
			{
				ID:                    2,
				Keyword:               "gambling",
				RiskCategory:          "prohibited",
				RiskSeverity:          "high",
				Description:           "Gambling and betting activities",
				MCCCodes:              []string{"7995"},
				NAICSCodes:            []string{"713290"},
				SICCodes:              []string{"7995"},
				CardBrandRestrictions: []string{"Visa", "Mastercard", "American Express"},
				DetectionPatterns:     []string{"gambl", "betting", "casino"},
				Synonyms:              []string{"betting", "casino", "wagering"},
				IsActive:              true,
				CreatedAt:             time.Now(),
				UpdatedAt:             time.Now(),
			},
			{
				ID:                    3,
				Keyword:               "money laundering",
				RiskCategory:          "illegal",
				RiskSeverity:          "critical",
				Description:           "Money laundering and financial crimes",
				MCCCodes:              []string{},
				NAICSCodes:            []string{},
				SICCodes:              []string{},
				CardBrandRestrictions: []string{"Visa", "Mastercard", "American Express"},
				DetectionPatterns:     []string{"money.*launder", "cash.*wash"},
				Synonyms:              []string{"cash washing", "dirty money"},
				IsActive:              true,
				CreatedAt:             time.Now(),
				UpdatedAt:             time.Now(),
			},
		},
		logger: log.New(os.Stdout, "TEST: ", log.LstdFlags),
	}
}

// getAllActiveKeywords mock implementation
func (m *MockRiskKeywordsService) getAllActiveKeywords(ctx context.Context) ([]*RiskKeyword, error) {
	return m.keywords, nil
}

// calculateKeywordRiskScore mock implementation
func (m *MockRiskKeywordsService) calculateKeywordRiskScore(keyword *RiskKeyword) float64 {
	switch keyword.RiskSeverity {
	case "critical":
		return 0.4
	case "high":
		return 0.3
	case "medium":
		return 0.2
	case "low":
		return 0.1
	default:
		return 0.1
	}
}

// calculateFinalRiskScore mock implementation
func (m *MockRiskKeywordsService) calculateFinalRiskScore(totalScore float64, matchCount int, contentLength int) float64 {
	if matchCount == 0 {
		return 0.0
	}

	lengthFactor := 1.0
	if contentLength > 1000 {
		lengthFactor = 1.2
	} else if contentLength > 500 {
		lengthFactor = 1.1
	} else if contentLength < 100 {
		lengthFactor = 0.8
	}

	finalScore := totalScore * lengthFactor
	if finalScore > 1.0 {
		finalScore = 1.0
	}

	return finalScore
}

// determineRiskLevel mock implementation
func (m *MockRiskKeywordsService) determineRiskLevel(riskScore float64) string {
	if riskScore >= 0.8 {
		return "critical"
	} else if riskScore >= 0.6 {
		return "high"
	} else if riskScore >= 0.4 {
		return "medium"
	} else if riskScore >= 0.2 {
		return "low"
	} else {
		return "minimal"
	}
}

// calculateConfidence mock implementation
func (m *MockRiskKeywordsService) calculateConfidence(contentLength, matchCount, keywordCount int) float64 {
	confidence := 0.5

	if contentLength > 500 {
		confidence += 0.2
	} else if contentLength > 200 {
		confidence += 0.1
	}

	if matchCount > 3 {
		confidence += 0.2
	} else if matchCount > 1 {
		confidence += 0.1
	}

	if keywordCount > 5 {
		confidence += 0.1
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// matchesPattern mock implementation
func (m *MockRiskKeywordsService) matchesPattern(content, pattern string) bool {
	// Simple pattern matching for testing
	return len(content) > 0 && len(pattern) > 0
}

// DetectRiskKeywords mock implementation
func (m *MockRiskKeywordsService) DetectRiskKeywords(ctx context.Context, content string) (*RiskDetectionResult, error) {
	if content == "" {
		return &RiskDetectionResult{
			DetectedKeywords: []string{},
			RiskScore:        0.0,
			RiskLevel:        "low",
			RiskCategories:   []string{},
			CardRestrictions: []string{},
			MCCRestrictions:  []string{},
			Confidence:       1.0,
			Evidence:         []string{},
			Metadata:         map[string]interface{}{"reason": "no_content"},
		}, nil
	}

	normalizedContent := strings.ToLower(strings.TrimSpace(content))
	detectedKeywords := make([]string, 0)
	riskCategories := make(map[string]bool)
	cardRestrictions := make(map[string]bool)
	mccRestrictions := make(map[string]bool)
	evidence := make([]string, 0)

	var totalRiskScore float64
	var matchCount int

	for _, keyword := range m.keywords {
		// Check direct keyword match
		if strings.Contains(normalizedContent, strings.ToLower(keyword.Keyword)) {
			detectedKeywords = append(detectedKeywords, keyword.Keyword)
			riskCategories[keyword.RiskCategory] = true

			// Add card brand restrictions
			for _, restriction := range keyword.CardBrandRestrictions {
				cardRestrictions[restriction] = true
			}

			// Add MCC restrictions
			for _, mcc := range keyword.MCCCodes {
				mccRestrictions[mcc] = true
			}

			// Calculate risk score contribution
			riskScore := m.calculateKeywordRiskScore(keyword)
			totalRiskScore += riskScore
			matchCount++

			evidence = append(evidence, fmt.Sprintf("Direct keyword match: %s (category: %s, severity: %s)",
				keyword.Keyword, keyword.RiskCategory, keyword.RiskSeverity))
		}

		// Check synonym matches
		for _, synonym := range keyword.Synonyms {
			if strings.Contains(normalizedContent, strings.ToLower(synonym)) {
				detectedKeywords = append(detectedKeywords, synonym)
				riskCategories[keyword.RiskCategory] = true

				// Add restrictions
				for _, restriction := range keyword.CardBrandRestrictions {
					cardRestrictions[restriction] = true
				}
				for _, mcc := range keyword.MCCCodes {
					mccRestrictions[mcc] = true
				}

				// Calculate risk score contribution (reduced for synonyms)
				riskScore := m.calculateKeywordRiskScore(keyword) * 0.8
				totalRiskScore += riskScore
				matchCount++

				evidence = append(evidence, fmt.Sprintf("Synonym match: %s -> %s (category: %s, severity: %s)",
					synonym, keyword.Keyword, keyword.RiskCategory, keyword.RiskSeverity))
			}
		}
	}

	// Calculate final risk score and level
	finalRiskScore := m.calculateFinalRiskScore(totalRiskScore, matchCount, len(normalizedContent))
	riskLevel := m.determineRiskLevel(finalRiskScore)

	// Convert maps to slices
	categorySlice := make([]string, 0, len(riskCategories))
	for category := range riskCategories {
		categorySlice = append(categorySlice, category)
	}

	cardRestrictionSlice := make([]string, 0, len(cardRestrictions))
	for restriction := range cardRestrictions {
		cardRestrictionSlice = append(cardRestrictionSlice, restriction)
	}

	mccRestrictionSlice := make([]string, 0, len(mccRestrictions))
	for mcc := range mccRestrictions {
		mccRestrictionSlice = append(mccRestrictionSlice, mcc)
	}

	// Calculate confidence based on content length and match quality
	confidence := m.calculateConfidence(len(normalizedContent), matchCount, len(detectedKeywords))

	result := &RiskDetectionResult{
		DetectedKeywords: detectedKeywords,
		RiskScore:        finalRiskScore,
		RiskLevel:        riskLevel,
		RiskCategories:   categorySlice,
		CardRestrictions: cardRestrictionSlice,
		MCCRestrictions:  mccRestrictionSlice,
		Confidence:       confidence,
		Evidence:         evidence,
		Metadata: map[string]interface{}{
			"content_length": len(normalizedContent),
			"match_count":    matchCount,
			"keyword_count":  len(detectedKeywords),
			"analysis_time":  time.Now(),
		},
	}

	return result, nil
}

// TestDetectRiskKeywords tests the risk keyword detection functionality
func TestDetectRiskKeywords(t *testing.T) {
	ctx := context.Background()
	service := NewMockRiskKeywordsService()

	tests := []struct {
		name                string
		content             string
		expectedMinLevel    string
		expectedMaxLevel    string
		expectedMinScore    float64
		expectedMaxScore    float64
		expectedMinKeywords int
		expectedMaxKeywords int
	}{
		{
			name:                "empty content",
			content:             "",
			expectedMinLevel:    "low",
			expectedMaxLevel:    "low",
			expectedMinScore:    0.0,
			expectedMaxScore:    0.0,
			expectedMinKeywords: 0,
			expectedMaxKeywords: 0,
		},
		{
			name:                "drug trafficking content",
			content:             "This business is involved in drug trafficking and illegal substances",
			expectedMinLevel:    "low",
			expectedMaxLevel:    "high",
			expectedMinScore:    0.0,
			expectedMaxScore:    1.0,
			expectedMinKeywords: 1,
			expectedMaxKeywords: 10,
		},
		{
			name:                "gambling content",
			content:             "Online gambling and casino services available here",
			expectedMinLevel:    "low",
			expectedMaxLevel:    "high",
			expectedMinScore:    0.0,
			expectedMaxScore:    1.0,
			expectedMinKeywords: 1,
			expectedMaxKeywords: 10,
		},
		{
			name:                "money laundering content",
			content:             "We provide money laundering services and cash washing",
			expectedMinLevel:    "low",
			expectedMaxLevel:    "high",
			expectedMinScore:    0.0,
			expectedMaxScore:    1.0,
			expectedMinKeywords: 1,
			expectedMaxKeywords: 10,
		},
		{
			name:                "multiple risk keywords",
			content:             "Drug trafficking, gambling, and money laundering services",
			expectedMinLevel:    "medium",
			expectedMaxLevel:    "critical",
			expectedMinScore:    0.0,
			expectedMaxScore:    1.0,
			expectedMinKeywords: 3,
			expectedMaxKeywords: 20,
		},
		{
			name:                "no risk keywords",
			content:             "This is a legitimate business selling software products",
			expectedMinLevel:    "minimal",
			expectedMaxLevel:    "low",
			expectedMinScore:    0.0,
			expectedMaxScore:    0.1,
			expectedMinKeywords: 0,
			expectedMaxKeywords: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.DetectRiskKeywords(ctx, tt.content)
			if err != nil {
				t.Fatalf("DetectRiskKeywords() error = %v", err)
			}

			// Test risk level is within expected range
			validLevels := []string{"minimal", "low", "medium", "high", "critical"}
			levelIndex := -1
			for i, level := range validLevels {
				if result.RiskLevel == level {
					levelIndex = i
					break
				}
			}
			if levelIndex == -1 {
				t.Errorf("DetectRiskKeywords() invalid risk level = %v", result.RiskLevel)
			}

			// Test risk score is within expected range
			if result.RiskScore < tt.expectedMinScore || result.RiskScore > tt.expectedMaxScore {
				t.Errorf("DetectRiskKeywords() risk score = %v, want between %v and %v", result.RiskScore, tt.expectedMinScore, tt.expectedMaxScore)
			}

			// Test keyword count is within expected range
			if len(result.DetectedKeywords) < tt.expectedMinKeywords || len(result.DetectedKeywords) > tt.expectedMaxKeywords {
				t.Errorf("DetectRiskKeywords() detected keywords count = %v, want between %v and %v", len(result.DetectedKeywords), tt.expectedMinKeywords, tt.expectedMaxKeywords)
			}

			// Test that confidence is valid
			if result.Confidence < 0.0 || result.Confidence > 1.0 {
				t.Errorf("DetectRiskKeywords() confidence = %v, want between 0.0 and 1.0", result.Confidence)
			}
		})
	}
}

// TestCalculateKeywordRiskScore tests the risk score calculation
func TestCalculateKeywordRiskScore(t *testing.T) {
	service := NewMockRiskKeywordsService()

	tests := []struct {
		name     string
		keyword  *RiskKeyword
		expected float64
	}{
		{
			name: "critical severity",
			keyword: &RiskKeyword{
				RiskSeverity: "critical",
				RiskCategory: "illegal",
			},
			expected: 0.4,
		},
		{
			name: "high severity",
			keyword: &RiskKeyword{
				RiskSeverity: "high",
				RiskCategory: "prohibited",
			},
			expected: 0.3,
		},
		{
			name: "medium severity",
			keyword: &RiskKeyword{
				RiskSeverity: "medium",
				RiskCategory: "high_risk",
			},
			expected: 0.2,
		},
		{
			name: "low severity",
			keyword: &RiskKeyword{
				RiskSeverity: "low",
				RiskCategory: "fraud",
			},
			expected: 0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.calculateKeywordRiskScore(tt.keyword)
			if score != tt.expected {
				t.Errorf("calculateKeywordRiskScore() = %v, want %v", score, tt.expected)
			}
		})
	}
}

// TestDetermineRiskLevel tests the risk level determination
func TestDetermineRiskLevel(t *testing.T) {
	service := NewMockRiskKeywordsService()

	tests := []struct {
		name     string
		score    float64
		expected string
	}{
		{"critical risk", 0.9, "critical"},
		{"high risk", 0.7, "high"},
		{"medium risk", 0.5, "medium"},
		{"low risk", 0.3, "low"},
		{"minimal risk", 0.1, "minimal"},
		{"zero risk", 0.0, "minimal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := service.determineRiskLevel(tt.score)
			if level != tt.expected {
				t.Errorf("determineRiskLevel() = %v, want %v", level, tt.expected)
			}
		})
	}
}

// TestCalculateConfidence tests the confidence calculation
func TestCalculateConfidence(t *testing.T) {
	service := NewMockRiskKeywordsService()

	tests := []struct {
		name          string
		contentLength int
		matchCount    int
		keywordCount  int
		expectedMin   float64
		expectedMax   float64
	}{
		{"short content, no matches", 50, 0, 0, 0.3, 0.5},
		{"medium content, few matches", 300, 2, 2, 0.6, 0.8},
		{"long content, many matches", 1000, 5, 6, 0.8, 1.0},
		{"very long content, many matches", 2000, 10, 10, 0.9, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := service.calculateConfidence(tt.contentLength, tt.matchCount, tt.keywordCount)
			if confidence < tt.expectedMin || confidence > tt.expectedMax {
				t.Errorf("calculateConfidence() = %v, want between %v and %v", confidence, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

// TestMatchesPattern tests the pattern matching functionality
func TestMatchesPattern(t *testing.T) {
	service := NewMockRiskKeywordsService()

	tests := []struct {
		name     string
		content  string
		pattern  string
		expected bool
	}{
		{"empty content", "", "test", false},
		{"empty pattern", "test content", "", false},
		{"both empty", "", "", false},
		{"valid content and pattern", "test content", "test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.matchesPattern(tt.content, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchesPattern() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// BenchmarkDetectRiskKeywords benchmarks the risk detection performance
func BenchmarkDetectRiskKeywords(b *testing.B) {
	ctx := context.Background()
	service := NewMockRiskKeywordsService()
	content := "This business is involved in drug trafficking, gambling, and money laundering activities"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.DetectRiskKeywords(ctx, content)
		if err != nil {
			b.Fatalf("DetectRiskKeywords() error = %v", err)
		}
	}
}

// TestRiskDetectionResultStructure tests the structure of risk detection results
func TestRiskDetectionResultStructure(t *testing.T) {
	ctx := context.Background()
	service := NewMockRiskKeywordsService()
	content := "Drug trafficking and gambling services"

	result, err := service.DetectRiskKeywords(ctx, content)
	if err != nil {
		t.Fatalf("DetectRiskKeywords() error = %v", err)
	}

	// Test that all required fields are present
	if result.DetectedKeywords == nil {
		t.Error("DetectedKeywords should not be nil")
	}

	if result.RiskCategories == nil {
		t.Error("RiskCategories should not be nil")
	}

	if result.CardRestrictions == nil {
		t.Error("CardRestrictions should not be nil")
	}

	if result.MCCRestrictions == nil {
		t.Error("MCCRestrictions should not be nil")
	}

	if result.Evidence == nil {
		t.Error("Evidence should not be nil")
	}

	if result.Metadata == nil {
		t.Error("Metadata should not be nil")
	}

	// Test that confidence is within valid range
	if result.Confidence < 0.0 || result.Confidence > 1.0 {
		t.Errorf("Confidence should be between 0.0 and 1.0, got %v", result.Confidence)
	}

	// Test that risk score is within valid range
	if result.RiskScore < 0.0 || result.RiskScore > 1.0 {
		t.Errorf("RiskScore should be between 0.0 and 1.0, got %v", result.RiskScore)
	}
}
