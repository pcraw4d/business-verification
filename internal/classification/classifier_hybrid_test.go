package classification

import (
	"context"
	"log"
	"testing"

	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/classification/testutil"
)

// TestGenerateCodesFromKeywords tests keyword-based code generation
func TestGenerateCodesFromKeywords(t *testing.T) {
	mockRepo := testutil.NewMockKeywordRepository()
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	tests := []struct {
		name                string
		keywords            []string
		codeType            string
		industryConfidence  float64
		expectedMatchCount  int
		expectError         bool
	}{
		{
			name:               "successful keyword match",
			keywords:           []string{"software", "technology"},
			codeType:           "MCC",
			industryConfidence: 0.85,
			expectedMatchCount: 2,
			expectError:        false,
		},
		{
			name:               "empty keywords",
			keywords:           []string{},
			codeType:           "MCC",
			industryConfidence: 0.85,
			expectedMatchCount: 0,
			expectError:        false,
		},
		{
			name:               "no keyword matches",
			keywords:           []string{"nonexistent"},
			codeType:           "SIC",
			industryConfidence: 0.85,
			expectedMatchCount: 0,
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			matches, err := generator.generateCodesFromKeywords(ctx, tt.keywords, tt.codeType, tt.industryConfidence)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(matches) != tt.expectedMatchCount {
				t.Errorf("Expected %d matches, got %d", tt.expectedMatchCount, len(matches))
			}

			// Verify match structure
			for _, match := range matches {
				if match.Source != "keyword" {
					t.Errorf("Expected source 'keyword', got %s", match.Source)
				}
				if match.Code.CodeType != tt.codeType {
					t.Errorf("Expected code type %s, got %s", tt.codeType, match.Code.CodeType)
				}
				if match.Confidence < 0 || match.Confidence > 1 {
					t.Errorf("Confidence out of range: %.2f", match.Confidence)
				}
			}
		})
	}
}

// TestMergeCodeResults tests the merging of industry and keyword-based codes
func TestMergeCodeResults(t *testing.T) {
	generator := NewClassificationCodeGenerator(nil, log.Default())

	tests := []struct {
		name                string
		industryCodes       []*repository.ClassificationCode
		keywordCodes        []CodeMatch
		industryConfidence  float64
		codeType            string
		expectedCount       int
		expectedBothSources int // Codes matched by both sources
	}{
		{
			name: "merge industry and keyword codes",
			industryCodes: []*repository.ClassificationCode{
				{ID: 1, Code: "1234", CodeType: "MCC", Description: "Test Code 1"},
				{ID: 2, Code: "5678", CodeType: "MCC", Description: "Test Code 2"},
			},
			keywordCodes: []CodeMatch{
				{
					Code:           &repository.ClassificationCode{ID: 1, Code: "1234", CodeType: "MCC", Description: "Test Code 1"},
					RelevanceScore: 0.8,
					MatchType:      "exact",
					Source:         "keyword",
					Confidence:     0.7,
				},
				{
					Code:           &repository.ClassificationCode{ID: 3, Code: "9012", CodeType: "MCC", Description: "Test Code 3"},
					RelevanceScore: 0.9,
					MatchType:      "exact",
					Source:         "keyword",
					Confidence:     0.75,
				},
			},
			industryConfidence:  0.85,
			codeType:           "MCC",
			expectedCount:      3, // 2 from industry + 1 unique from keyword
			expectedBothSources: 1, // Code 1234 matched by both
		},
		{
			name:                "only industry codes",
			industryCodes:       []*repository.ClassificationCode{{ID: 1, Code: "1234", CodeType: "MCC", Description: "Test"}},
			keywordCodes:        []CodeMatch{},
			industryConfidence:  0.85,
			codeType:           "MCC",
			expectedCount:      1,
			expectedBothSources: 0,
		},
		{
			name:                "only keyword codes",
			industryCodes:       []*repository.ClassificationCode{},
			keywordCodes:        []CodeMatch{{Code: &repository.ClassificationCode{ID: 1, Code: "1234", CodeType: "MCC", Description: "Test"}, Source: "keyword", Confidence: 0.7}},
			industryConfidence:  0.85,
			codeType:           "MCC",
			expectedCount:      1,
			expectedBothSources: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := generator.mergeCodeResults(tt.industryCodes, tt.keywordCodes, tt.industryConfidence, tt.codeType)

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
			}

			// Count codes matched by both sources
			bothSourcesCount := 0
			for _, result := range results {
				hasIndustry := false
				hasKeyword := false
				for _, source := range result.Sources {
					if source == "industry" {
						hasIndustry = true
					}
					if source == "keyword" {
						hasKeyword = true
					}
				}
				if hasIndustry && hasKeyword {
					bothSourcesCount++
					// Verify boost is applied (confidence should be higher)
					if result.CombinedConfidence < 0.7 {
						t.Errorf("Expected boosted confidence for both-source match, got %.2f", result.CombinedConfidence)
					}
				}
			}

			if bothSourcesCount != tt.expectedBothSources {
				t.Errorf("Expected %d codes matched by both sources, got %d", tt.expectedBothSources, bothSourcesCount)
			}

			// Verify results are sorted by confidence (descending)
			for i := 1; i < len(results); i++ {
				if results[i-1].CombinedConfidence < results[i].CombinedConfidence {
					t.Errorf("Results not sorted by confidence: %.2f < %.2f", results[i-1].CombinedConfidence, results[i].CombinedConfidence)
				}
			}
		})
	}
}

// TestMergeCodeResults_ConfidenceFiltering tests confidence threshold filtering
func TestMergeCodeResults_ConfidenceFiltering(t *testing.T) {
	generator := NewClassificationCodeGenerator(nil, log.Default())

	// Create codes with varying confidence levels
	industryCodes := []*repository.ClassificationCode{
		{ID: 1, Code: "1234", CodeType: "MCC", Description: "High Confidence"},
		{ID: 2, Code: "5678", CodeType: "MCC", Description: "Low Confidence"},
	}

	keywordCodes := []CodeMatch{
		{
			Code:       &repository.ClassificationCode{ID: 3, Code: "9012", CodeType: "MCC", Description: "Medium Confidence"},
			Source:     "keyword",
			Confidence: 0.65, // Just above threshold
		},
		{
			Code:       &repository.ClassificationCode{ID: 4, Code: "3456", CodeType: "MCC", Description: "Very Low Confidence"},
			Source:     "keyword",
			Confidence: 0.4, // Below threshold
		},
	}

	// Use low industry confidence to create some codes below threshold
	results := generator.mergeCodeResults(industryCodes, keywordCodes, 0.5, "MCC")

	// All results should be above threshold (0.6)
	for _, result := range results {
		if result.CombinedConfidence < 0.6 {
			t.Errorf("Result with confidence %.2f should have been filtered out", result.CombinedConfidence)
		}
	}
}

// TestMergeCodeResults_TopNLimiting tests top-N limiting
func TestMergeCodeResults_TopNLimiting(t *testing.T) {
	generator := NewClassificationCodeGenerator(nil, log.Default())

	// Create many codes
	industryCodes := make([]*repository.ClassificationCode, 15)
	for i := 0; i < 15; i++ {
		industryCodes[i] = &repository.ClassificationCode{
			ID:          i + 1,
			Code:        string(rune('0' + i)),
			CodeType:    "MCC",
			Description: "Test Code",
		}
	}

	results := generator.mergeCodeResults(industryCodes, []CodeMatch{}, 0.9, "MCC")

	// Should be limited to top 10
	if len(results) > 10 {
		t.Errorf("Expected max 10 results, got %d", len(results))
	}
}

// TestGenerateCodesForMultipleIndustries tests multi-industry code generation
func TestGenerateCodesForMultipleIndustries(t *testing.T) {
	mockRepo := testutil.NewMockKeywordRepository()
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	ctx := context.Background()
	industries := []IndustryResult{
		{IndustryName: "Technology", Confidence: 0.9},
		{IndustryName: "Software", Confidence: 0.7},
	}

	codes := generator.generateCodesForMultipleIndustries(ctx, industries, "MCC")

	// Should have codes from both industries
	if len(codes) == 0 {
		t.Error("Expected codes from multiple industries")
	}

	// Verify deduplication (same code from multiple industries should appear once)
	codeMap := make(map[int]bool)
	for _, code := range codes {
		if codeMap[code.ID] {
			t.Errorf("Duplicate code found: %d", code.ID)
		}
		codeMap[code.ID] = true
	}
}

// TestGenerateClassificationCodes_MultiIndustry tests the full multi-industry flow
func TestGenerateClassificationCodes_MultiIndustry(t *testing.T) {
	mockRepo := testutil.NewMockKeywordRepository()
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	ctx := context.Background()
	keywords := []string{"software", "technology"}
	detectedIndustry := "Technology"
	confidence := 0.85
	additionalIndustries := []IndustryResult{
		{IndustryName: "Software", Confidence: 0.75},
	}

	codes, err := generator.GenerateClassificationCodes(ctx, keywords, detectedIndustry, confidence, additionalIndustries...)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if codes == nil {
		t.Error("Expected codes to be generated")
		return
	}

	// Should have codes from both primary and additional industries
	totalCodes := len(codes.MCC) + len(codes.SIC) + len(codes.NAICS)
	if totalCodes == 0 {
		t.Error("Expected codes to be generated from multiple industries")
	}
}

// Note: MockKeywordRepositoryWithKeywords has been replaced by testutil.NewMockKeywordRepository()

