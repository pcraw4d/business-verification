package risk

import (
	"testing"

	"go.uber.org/zap"
)

func TestRiskKeywordMatcher_MatchKeywords(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	config := DefaultRiskDetectionConfig()
	matcher := NewRiskKeywordMatcher(logger, config)

	// Test keywords
	riskKeywords := []RiskKeyword{
		{
			ID:           1,
			Keyword:      "drug trafficking",
			RiskCategory: "illegal",
			RiskSeverity: "critical",
			Synonyms:     []string{"drug dealing", "narcotics trafficking"},
			IsActive:     true,
		},
		{
			ID:           2,
			Keyword:      "adult entertainment",
			RiskCategory: "prohibited",
			RiskSeverity: "high",
			Synonyms:     []string{"pornography", "adult content"},
			IsActive:     true,
		},
	}

	tests := []struct {
		name            string
		content         string
		source          string
		expectedCount   int
		expectedKeyword string
	}{
		{
			name:            "Direct keyword match",
			content:         "We are involved in drug trafficking activities",
			source:          "business_description",
			expectedCount:   1,
			expectedKeyword: "drug trafficking",
		},
		{
			name:            "Synonym match",
			content:         "We deal in narcotics trafficking",
			source:          "business_description",
			expectedCount:   1,
			expectedKeyword: "drug trafficking",
		},
		{
			name:            "Multiple matches",
			content:         "We provide adult entertainment and deal in drug trafficking",
			source:          "business_description",
			expectedCount:   2,
			expectedKeyword: "",
		},
		{
			name:            "No matches",
			content:         "We sell coffee and pastries",
			source:          "business_description",
			expectedCount:   0,
			expectedKeyword: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := matcher.MatchKeywords(tt.content, riskKeywords, tt.source)

			if len(matches) != tt.expectedCount {
				t.Errorf("Expected %d matches, got %d", tt.expectedCount, len(matches))
			}

			if tt.expectedKeyword != "" && len(matches) > 0 {
				if matches[0].Keyword != tt.expectedKeyword {
					t.Errorf("Expected keyword %s, got %s", tt.expectedKeyword, matches[0].Keyword)
				}
			}

			// Check match properties
			for _, match := range matches {
				if match.Source != tt.source {
					t.Errorf("Expected source %s, got %s", tt.source, match.Source)
				}
				if match.Confidence <= 0 || match.Confidence > 1 {
					t.Errorf("Invalid confidence score: %f", match.Confidence)
				}
				if match.Context == "" {
					t.Error("Context should not be empty")
				}
			}
		})
	}
}

func TestRiskScorer_CalculateRiskScore(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultRiskDetectionConfig()
	scorer := NewRiskScorer(logger, config)

	tests := []struct {
		name          string
		keywords      []DetectedRiskKeyword
		expectedScore float64
		expectedLevel RiskLevel
	}{
		{
			name: "Critical risk keyword",
			keywords: []DetectedRiskKeyword{
				{
					Keyword:    "drug trafficking",
					Category:   "illegal",
					Severity:   "critical",
					Confidence: 0.9,
				},
			},
			expectedScore: 0.9,
			expectedLevel: RiskLevelCritical,
		},
		{
			name: "High risk keyword",
			keywords: []DetectedRiskKeyword{
				{
					Keyword:    "adult entertainment",
					Category:   "prohibited",
					Severity:   "high",
					Confidence: 0.8,
				},
			},
			expectedScore: 0.64, // 0.8 * 0.8 * 1.0
			expectedLevel: RiskLevelHigh,
		},
		{
			name: "Multiple keywords",
			keywords: []DetectedRiskKeyword{
				{
					Keyword:    "drug trafficking",
					Category:   "illegal",
					Severity:   "critical",
					Confidence: 0.9,
				},
				{
					Keyword:    "adult entertainment",
					Category:   "prohibited",
					Severity:   "high",
					Confidence: 0.8,
				},
			},
			expectedScore: 0.77, // Amplified due to multiple detections
			expectedLevel: RiskLevelHigh,
		},
		{
			name:          "No keywords",
			keywords:      []DetectedRiskKeyword{},
			expectedScore: 0.0,
			expectedLevel: RiskLevelMinimal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, level := scorer.CalculateRiskScore(tt.keywords)

			if score < tt.expectedScore-0.1 || score > tt.expectedScore+0.1 {
				t.Errorf("Expected score around %f, got %f", tt.expectedScore, score)
			}

			if level != tt.expectedLevel {
				t.Errorf("Expected level %s, got %s", tt.expectedLevel, level)
			}
		})
	}
}

func TestRiskPatternDetector_DetectPatterns(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultRiskDetectionConfig()
	detector := NewRiskPatternDetector(logger, config)

	tests := []struct {
		name          string
		request       *RiskDetectionRequest
		expectedCount int
		expectedType  string
	}{
		{
			name: "Money laundering patterns",
			request: &RiskDetectionRequest{
				BusinessName:        "Cash Only Business",
				BusinessDescription: "We only accept cash transactions and deal in high-value items",
			},
			expectedCount: 2, // Cash-intensive + High-value
			expectedType:  "money_laundering",
		},
		{
			name: "Fraud patterns",
			request: &RiskDetectionRequest{
				BusinessName:        "Identity Services",
				BusinessDescription: "We provide fake ID services and credit card fraud",
			},
			expectedCount: 2, // Identity fraud + Credit card fraud
			expectedType:  "fraud_pattern",
		},
		{
			name: "Shell company patterns",
			request: &RiskDetectionRequest{
				BusinessName:        "Offshore Holdings",
				BusinessDescription: "We are a shell company operating in tax havens",
			},
			expectedCount: 1,
			expectedType:  "shell_company",
		},
		{
			name: "No patterns",
			request: &RiskDetectionRequest{
				BusinessName:        "Coffee Shop",
				BusinessDescription: "We serve coffee and pastries to our customers",
			},
			expectedCount: 0,
			expectedType:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := detector.DetectPatterns(tt.request, []RiskKeyword{})

			if len(patterns) != tt.expectedCount {
				t.Errorf("Expected %d patterns, got %d", tt.expectedCount, len(patterns))
			}

			if tt.expectedType != "" && len(patterns) > 0 {
				found := false
				for _, pattern := range patterns {
					if pattern.PatternType == tt.expectedType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected pattern type %s not found", tt.expectedType)
				}
			}

			// Check pattern properties
			for _, pattern := range patterns {
				if pattern.Confidence <= 0 || pattern.Confidence > 1 {
					t.Errorf("Invalid confidence score: %f", pattern.Confidence)
				}
				if pattern.Context == "" {
					t.Error("Context should not be empty")
				}
				if pattern.PatternName == "" {
					t.Error("Pattern name should not be empty")
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkRiskKeywordMatcher_MatchKeywords(b *testing.B) {
	logger := zap.NewNop()
	config := DefaultRiskDetectionConfig()
	matcher := NewRiskKeywordMatcher(logger, config)

	riskKeywords := []RiskKeyword{
		{
			ID:           1,
			Keyword:      "drug trafficking",
			RiskCategory: "illegal",
			RiskSeverity: "critical",
			Synonyms:     []string{"drug dealing", "narcotics trafficking"},
			IsActive:     true,
		},
	}

	content := "We are involved in drug trafficking activities and provide adult entertainment services"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matcher.MatchKeywords(content, riskKeywords, "business_description")
	}
}
