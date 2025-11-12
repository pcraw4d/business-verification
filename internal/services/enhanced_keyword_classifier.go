// Package services provides enhanced keyword classification with relationship mapping
package services

import (
	"context"
	"fmt"
	"math"
	"strings"

	"kyb-platform/internal/repository"
)

// EnhancedKeywordClassifier provides advanced keyword classification with relationship mapping
// Stub: KeywordWeightRepository doesn't exist - needs implementation
type EnhancedKeywordClassifier struct {
	// keywordRepo      repository.KeywordWeightRepository // Stub - interface doesn't exist
	relationshipRepo repository.KeywordRelationshipRepository
	expansionService KeywordExpansionService
}

// NewEnhancedKeywordClassifier creates a new enhanced keyword classifier
// Stub: KeywordWeightRepository parameter removed - interface doesn't exist
func NewEnhancedKeywordClassifier(
	// keywordRepo repository.KeywordWeightRepository, // Stub - interface doesn't exist
	relationshipRepo repository.KeywordRelationshipRepository,
	expansionService KeywordExpansionService,
) *EnhancedKeywordClassifier {
	return &EnhancedKeywordClassifier{
		// keywordRepo:      keywordRepo, // Stub
		relationshipRepo: relationshipRepo,
		expansionService: expansionService,
	}
}

// ClassifyWithExpansion classifies a business using enhanced keyword expansion
// Stub: models.IndustryMatch doesn't exist - needs implementation
func (c *EnhancedKeywordClassifier) ClassifyWithExpansion(ctx context.Context, businessName, description string) ([]interface{}, error) {
	// Extract keywords from business name and description
	allText := fmt.Sprintf("%s %s", businessName, description)
	extractedKeywords := c.extractKeywords(allText)

	// Stub: models.IndustryMatch doesn't exist - return empty slice
	if len(extractedKeywords) == 0 {
		return []interface{}{}, nil
	}

	// TODO: Implement industry matching when models.IndustryMatch is available
	return []interface{}{}, nil
}

// classifyForIndustry classifies keywords for a specific industry with expansion
// Stub: models.Industry and models.IndustryMatch don't exist - needs implementation
// TODO: Implement when models.Industry, models.IndustryMatch, and models.KeywordMatch are available
func (c *EnhancedKeywordClassifier) classifyForIndustry(ctx context.Context, keywords []string, industry interface{}) (interface{}, error) {
	// Stub: This function requires models.Industry, models.IndustryMatch, and models.KeywordMatch
	// which don't exist. Also requires keywordRepo which was commented out.
	// Return nil for now - this function is not used since ClassifyWithExpansion returns empty slice
	return nil, fmt.Errorf("classifyForIndustry not implemented: models.Industry and models.IndustryMatch types are not available")
}

// determineMatchType determines how a keyword was matched (direct, synonym, abbreviation, etc.)
func (c *EnhancedKeywordClassifier) determineMatchType(ctx context.Context, matchedKeyword string, originalKeywords []string) string {
	// Check if it's a direct match
	for _, original := range originalKeywords {
		if strings.ToLower(strings.TrimSpace(original)) == matchedKeyword {
			return "direct"
		}
	}

	// If not direct, it must be from expansion
	// We could enhance this by checking the relationship type from the expansion result
	return "expanded"
}

// calculateConfidenceScore calculates the final confidence score with multiple factors
func (c *EnhancedKeywordClassifier) calculateConfidenceScore(totalScore float64, matchedKeywords, totalKeywords int, threshold float64) float64 {
	if totalKeywords == 0 {
		return 0.0
	}

	// Base score from keyword weights
	baseScore := totalScore / float64(totalKeywords)

	// Coverage factor (percentage of keywords matched)
	coverageFactor := float64(matchedKeywords) / float64(totalKeywords)

	// Density bonus (more matches = higher confidence)
	densityBonus := math.Min(float64(matchedKeywords)*0.1, 0.3)

	// Threshold adjustment
	thresholdAdjustment := 1.0
	if threshold > 0 {
		thresholdAdjustment = math.Min(baseScore/threshold, 1.2)
	}

	// Combine factors
	confidenceScore := (baseScore * coverageFactor * thresholdAdjustment) + densityBonus

	// Normalize to 0-1 range
	if confidenceScore > 1.0 {
		confidenceScore = 1.0
	}

	return confidenceScore
}

// generateReasoningDetails generates human-readable reasoning for the classification
// Stub: models.KeywordMatch doesn't exist - needs implementation
func (c *EnhancedKeywordClassifier) generateReasoningDetails(keywordDetails []interface{}, confidence float64) string {
	// Stub: keywordDetails is []interface{} since models.KeywordMatch doesn't exist
	if len(keywordDetails) == 0 {
		return "No relevant keywords found for this industry."
	}

	// Stub: Can't process keywordDetails without models.KeywordMatch type
	return fmt.Sprintf("Classification confidence: %.2f%% based on %d keyword matches.", confidence*100, len(keywordDetails))
}

// extractKeywords extracts keywords from text
func (c *EnhancedKeywordClassifier) extractKeywords(text string) []string {
	if text == "" {
		return []string{}
	}

	// Simple word extraction (can be enhanced with NLP)
	words := strings.Fields(strings.ToLower(text))

	// Filter out common stop words and short words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "will": true,
		"would": true, "could": true, "should": true, "may": true, "might": true,
		"can": true, "must": true, "shall": true, "this": true, "that": true,
		"these": true, "those": true, "i": true, "we": true, "you": true, "he": true,
		"she": true, "it": true, "they": true, "me": true, "us": true, "him": true,
		"her": true, "them": true, "my": true, "our": true, "your": true, "his": true,
		"its": true, "their": true, "as": true, "if": true, "then": true, "than": true,
		"so": true, "very": true, "just": true, "now": true, "only": true, "also": true,
	}

	filteredWords := make([]string, 0)
	for _, word := range words {
		// Remove punctuation
		cleaned := strings.Trim(word, ".,!?;:\"'()[]{}/*&^%$#@+=<>|\\`~")

		// Skip stop words and short words
		if len(cleaned) > 2 && !stopWords[cleaned] {
			filteredWords = append(filteredWords, cleaned)
		}
	}

	return filteredWords
}
