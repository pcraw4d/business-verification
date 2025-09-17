// Package services provides enhanced keyword classification with relationship mapping
package services

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/kyb-platform/internal/models"
	"github.com/kyb-platform/internal/repository"
)

// EnhancedKeywordClassifier provides advanced keyword classification with relationship mapping
type EnhancedKeywordClassifier struct {
	keywordRepo      repository.KeywordWeightRepository
	relationshipRepo repository.KeywordRelationshipRepository
	expansionService KeywordExpansionService
}

// NewEnhancedKeywordClassifier creates a new enhanced keyword classifier
func NewEnhancedKeywordClassifier(
	keywordRepo repository.KeywordWeightRepository,
	relationshipRepo repository.KeywordRelationshipRepository,
	expansionService KeywordExpansionService,
) *EnhancedKeywordClassifier {
	return &EnhancedKeywordClassifier{
		keywordRepo:      keywordRepo,
		relationshipRepo: relationshipRepo,
		expansionService: expansionService,
	}
}

// ClassifyWithExpansion classifies a business using enhanced keyword expansion
func (c *EnhancedKeywordClassifier) ClassifyWithExpansion(ctx context.Context, businessName, description string) ([]models.IndustryMatch, error) {
	// Extract keywords from business name and description
	allText := fmt.Sprintf("%s %s", businessName, description)
	extractedKeywords := c.extractKeywords(allText)

	if len(extractedKeywords) == 0 {
		return []models.IndustryMatch{}, nil
	}

	// Get all industries
	industries, err := c.keywordRepo.GetAllIndustries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get industries: %w", err)
	}

	industryMatches := make([]models.IndustryMatch, 0, len(industries))

	// Process each industry
	for _, industry := range industries {
		match, err := c.classifyForIndustry(ctx, extractedKeywords, industry)
		if err != nil {
			// Log error but continue with other industries
			continue
		}

		if match.ConfidenceScore > 0.1 { // Only include matches with reasonable confidence
			industryMatches = append(industryMatches, *match)
		}
	}

	// Sort by confidence score (highest first)
	for i := 0; i < len(industryMatches); i++ {
		for j := i + 1; j < len(industryMatches); j++ {
			if industryMatches[i].ConfidenceScore < industryMatches[j].ConfidenceScore {
				industryMatches[i], industryMatches[j] = industryMatches[j], industryMatches[i]
			}
		}
	}

	return industryMatches, nil
}

// classifyForIndustry classifies keywords for a specific industry with expansion
func (c *EnhancedKeywordClassifier) classifyForIndustry(ctx context.Context, keywords []string, industry models.Industry) (*models.IndustryMatch, error) {
	// Expand keywords using relationship mapping
	expandedKeywords, err := c.expansionService.ExpandKeywords(ctx, keywords, industry.ID)
	if err != nil {
		// If expansion fails, use original keywords
		expandedKeywords = keywords
	}

	// Get keyword weights for this industry
	keywordWeights, err := c.keywordRepo.GetKeywordWeightsByIndustry(ctx, industry.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyword weights for industry %d: %w", industry.ID, err)
	}

	// Create a map for quick lookup
	weightMap := make(map[string]float64)
	for _, kw := range keywordWeights {
		weightMap[strings.ToLower(kw.Keyword)] = kw.CalculatedWeight
	}

	// Calculate keyword relevance scores
	relevanceScores, err := c.expansionService.CalculateKeywordRelevance(ctx, expandedKeywords, industry.ID)
	if err != nil {
		// If relevance calculation fails, use default scores
		relevanceScores = make(map[string]float64)
		for _, keyword := range expandedKeywords {
			relevanceScores[strings.ToLower(keyword)] = 0.5
		}
	}

	// Calculate confidence score with multiple factors
	totalScore := 0.0
	matchedKeywords := 0
	keywordDetails := make([]models.KeywordMatch, 0)

	for _, keyword := range expandedKeywords {
		normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
		if normalizedKeyword == "" {
			continue
		}

		// Check if keyword exists in industry weights
		weight, exists := weightMap[normalizedKeyword]
		if !exists {
			// For keywords without explicit weights, use relevance score
			relevance, hasRelevance := relevanceScores[normalizedKeyword]
			if hasRelevance && relevance > 0.6 {
				weight = relevance * 0.5 // Lower weight for inferred keywords
				exists = true
			}
		}

		if exists {
			matchedKeywords++

			// Apply relevance multiplier
			relevance, hasRelevance := relevanceScores[normalizedKeyword]
			if hasRelevance {
				weight *= relevance
			}

			totalScore += weight

			keywordDetails = append(keywordDetails, models.KeywordMatch{
				Keyword:   normalizedKeyword,
				Weight:    weight,
				Relevance: relevance,
				MatchType: c.determineMatchType(ctx, normalizedKeyword, keywords),
			})
		}
	}

	// Calculate confidence score
	confidenceScore := c.calculateConfidenceScore(totalScore, matchedKeywords, len(expandedKeywords), industry.ConfidenceThreshold)

	match := &models.IndustryMatch{
		IndustryID:       industry.ID,
		IndustryName:     industry.Name,
		ConfidenceScore:  confidenceScore,
		MatchedKeywords:  matchedKeywords,
		TotalKeywords:    len(expandedKeywords),
		KeywordDetails:   keywordDetails,
		ExpandedKeywords: expandedKeywords,
		ReasoningDetails: c.generateReasoningDetails(keywordDetails, confidenceScore),
	}

	return match, nil
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
func (c *EnhancedKeywordClassifier) generateReasoningDetails(keywordDetails []models.KeywordMatch, confidence float64) string {
	if len(keywordDetails) == 0 {
		return "No relevant keywords found for this industry."
	}

	// Sort by weight (highest first)
	sortedKeywords := make([]models.KeywordMatch, len(keywordDetails))
	copy(sortedKeywords, keywordDetails)

	for i := 0; i < len(sortedKeywords); i++ {
		for j := i + 1; j < len(sortedKeywords); j++ {
			if sortedKeywords[i].Weight < sortedKeywords[j].Weight {
				sortedKeywords[i], sortedKeywords[j] = sortedKeywords[j], sortedKeywords[i]
			}
		}
	}

	// Take top keywords for reasoning
	topCount := 5
	if len(sortedKeywords) < topCount {
		topCount = len(sortedKeywords)
	}

	reasoning := fmt.Sprintf("Classification confidence: %.1f%%. ", confidence*100)
	reasoning += fmt.Sprintf("Key indicators: ")

	for i := 0; i < topCount; i++ {
		kw := sortedKeywords[i]
		if i > 0 {
			reasoning += ", "
		}
		reasoning += fmt.Sprintf("'%s' (%.2f)", kw.Keyword, kw.Weight)
	}

	if len(sortedKeywords) > topCount {
		remaining := len(sortedKeywords) - topCount
		reasoning += fmt.Sprintf(" and %d other indicators", remaining)
	}

	reasoning += "."

	return reasoning
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
