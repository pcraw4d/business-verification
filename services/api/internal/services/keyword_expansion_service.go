// Package services provides business logic for the KYB platform
package services

import (
	"context"
	"fmt"
	"strings"

	"kyb-platform/internal/models"
	"kyb-platform/internal/repository"
)

// KeywordExpansionService provides keyword expansion and relationship mapping functionality
type KeywordExpansionService interface {
	// ExpandKeywords expands a list of keywords using relationships and context
	ExpandKeywords(ctx context.Context, keywords []string, industryID int) ([]string, error)

	// ExpandKeywordWithDetails expands a keyword with detailed relationship information
	ExpandKeywordWithDetails(ctx context.Context, keyword string, industryID int) (*models.KeywordExpansionResult, error)

	// GetSynonyms gets synonyms for a keyword
	GetSynonyms(ctx context.Context, keyword string) ([]string, error)

	// GetAbbreviations gets abbreviations for a keyword
	GetAbbreviations(ctx context.Context, keyword string) ([]string, error)

	// GetRelatedTerms gets related terms for a keyword
	GetRelatedTerms(ctx context.Context, keyword string) ([]string, error)

	// ExpandBusinessDescription expands keywords from a business description
	ExpandBusinessDescription(ctx context.Context, description string, industryID int) ([]string, error)

	// CalculateKeywordRelevance calculates relevance score for keywords in an industry context
	CalculateKeywordRelevance(ctx context.Context, keywords []string, industryID int) (map[string]float64, error)
}

// keywordExpansionService implements KeywordExpansionService
type keywordExpansionService struct {
	keywordRepo repository.KeywordRelationshipRepository
}

// NewKeywordExpansionService creates a new keyword expansion service
func NewKeywordExpansionService(keywordRepo repository.KeywordRelationshipRepository) KeywordExpansionService {
	return &keywordExpansionService{
		keywordRepo: keywordRepo,
	}
}

// ExpandKeywords expands a list of keywords using relationships and context
func (s *keywordExpansionService) ExpandKeywords(ctx context.Context, keywords []string, industryID int) ([]string, error) {
	if len(keywords) == 0 {
		return []string{}, nil
	}

	// Use a map to avoid duplicates
	expandedSet := make(map[string]bool)

	// Add original keywords
	for _, keyword := range keywords {
		normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
		if normalizedKeyword != "" {
			expandedSet[normalizedKeyword] = true
		}
	}

	// Expand each keyword
	for _, keyword := range keywords {
		normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
		if normalizedKeyword == "" {
			continue
		}

		// Get expanded keywords for this keyword
		expansionResult, err := s.keywordRepo.ExpandKeyword(ctx, normalizedKeyword, industryID)
		if err != nil {
			// Log error but continue with other keywords
			continue
		}

		if expansionResult != nil {
			for _, expandedKeyword := range expansionResult.ExpandedKeywords {
				// Only include high-confidence expansions
				if expandedKeyword.Confidence >= 0.7 {
					normalizedExpanded := strings.ToLower(strings.TrimSpace(expandedKeyword.Keyword))
					if normalizedExpanded != "" {
						expandedSet[normalizedExpanded] = true
					}
				}
			}
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(expandedSet))
	for keyword := range expandedSet {
		result = append(result, keyword)
	}

	return result, nil
}

// ExpandKeywordWithDetails expands a keyword with detailed relationship information
func (s *keywordExpansionService) ExpandKeywordWithDetails(ctx context.Context, keyword string, industryID int) (*models.KeywordExpansionResult, error) {
	normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
	if normalizedKeyword == "" {
		return nil, fmt.Errorf("keyword cannot be empty")
	}

	return s.keywordRepo.ExpandKeyword(ctx, normalizedKeyword, industryID)
}

// GetSynonyms gets synonyms for a keyword
func (s *keywordExpansionService) GetSynonyms(ctx context.Context, keyword string) ([]string, error) {
	normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
	if normalizedKeyword == "" {
		return []string{}, nil
	}

	relationshipTypes := []string{models.RelationshipTypeSynonym}
	expandedKeywords, err := s.keywordRepo.GetRelatedKeywords(ctx, normalizedKeyword, relationshipTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to get synonyms: %w", err)
	}

	synonyms := make([]string, 0, len(expandedKeywords))
	for _, ek := range expandedKeywords {
		if ek.Confidence >= 0.7 { // Only high-confidence synonyms
			synonyms = append(synonyms, ek.Keyword)
		}
	}

	return synonyms, nil
}

// GetAbbreviations gets abbreviations for a keyword
func (s *keywordExpansionService) GetAbbreviations(ctx context.Context, keyword string) ([]string, error) {
	normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
	if normalizedKeyword == "" {
		return []string{}, nil
	}

	relationshipTypes := []string{models.RelationshipTypeAbbreviation}
	expandedKeywords, err := s.keywordRepo.GetRelatedKeywords(ctx, normalizedKeyword, relationshipTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to get abbreviations: %w", err)
	}

	abbreviations := make([]string, 0, len(expandedKeywords))
	for _, ek := range expandedKeywords {
		if ek.Confidence >= 0.7 { // Only high-confidence abbreviations
			abbreviations = append(abbreviations, ek.Keyword)
		}
	}

	return abbreviations, nil
}

// GetRelatedTerms gets related terms for a keyword
func (s *keywordExpansionService) GetRelatedTerms(ctx context.Context, keyword string) ([]string, error) {
	normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
	if normalizedKeyword == "" {
		return []string{}, nil
	}

	relationshipTypes := []string{models.RelationshipTypeRelated, models.RelationshipTypeVariant}
	expandedKeywords, err := s.keywordRepo.GetRelatedKeywords(ctx, normalizedKeyword, relationshipTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to get related terms: %w", err)
	}

	relatedTerms := make([]string, 0, len(expandedKeywords))
	for _, ek := range expandedKeywords {
		if ek.Confidence >= 0.6 { // Slightly lower threshold for related terms
			relatedTerms = append(relatedTerms, ek.Keyword)
		}
	}

	return relatedTerms, nil
}

// ExpandBusinessDescription expands keywords from a business description
func (s *keywordExpansionService) ExpandBusinessDescription(ctx context.Context, description string, industryID int) ([]string, error) {
	if description == "" {
		return []string{}, nil
	}

	// Extract keywords from description (simple word splitting for now)
	words := strings.Fields(strings.ToLower(description))

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
		"here": true, "there": true, "when": true, "where": true, "how": true, "what": true,
		"who": true, "which": true, "why": true, "do": true, "does": true, "did": true,
		"don't": true, "doesn't": true, "didn't": true, "won't": true, "wouldn't": true,
		"can't": true, "couldn't": true, "shouldn't": true, "mustn't": true,
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

	// Expand the filtered keywords
	return s.ExpandKeywords(ctx, filteredWords, industryID)
}

// CalculateKeywordRelevance calculates relevance score for keywords in an industry context
func (s *keywordExpansionService) CalculateKeywordRelevance(ctx context.Context, keywords []string, industryID int) (map[string]float64, error) {
	relevanceScores := make(map[string]float64)

	for _, keyword := range keywords {
		normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
		if normalizedKeyword == "" {
			continue
		}

		// Get contexts for this keyword in the industry
		contexts, err := s.keywordRepo.GetKeywordContexts(ctx, normalizedKeyword, industryID)
		if err != nil {
			// Log error but continue with other keywords
			relevanceScores[normalizedKeyword] = 0.5 // Default score
			continue
		}

		// Calculate relevance based on context weights
		totalWeight := 0.0
		maxWeight := 0.0

		for _, context := range contexts {
			totalWeight += context.ContextWeight
			if context.ContextWeight > maxWeight {
				maxWeight = context.ContextWeight
			}
		}

		// Use maximum weight as relevance score, normalized to 0-1 range
		relevance := maxWeight
		if relevance > 2.0 {
			relevance = 1.0 // Cap at 1.0
		} else if relevance == 0.0 {
			relevance = 0.5 // Default for keywords without context
		}

		relevanceScores[normalizedKeyword] = relevance
	}

	return relevanceScores, nil
}
