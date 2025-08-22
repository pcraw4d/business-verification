package classification

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// ConfidenceRankingEngine provides advanced confidence-based ranking for industry classifications
type ConfidenceRankingEngine struct {
	// Configuration
	baseConfidenceWeight     float64
	methodDiversityWeight    float64
	consistencyWeight        float64
	relevanceWeight          float64
	industryPopularityWeight float64

	// Industry popularity data (would be loaded from external source)
	industryPopularity map[string]float64
}

// NewConfidenceRankingEngine creates a new confidence ranking engine
func NewConfidenceRankingEngine() *ConfidenceRankingEngine {
	return &ConfidenceRankingEngine{
		baseConfidenceWeight:     0.4,
		methodDiversityWeight:    0.2,
		consistencyWeight:        0.15,
		relevanceWeight:          0.15,
		industryPopularityWeight: 0.1,
		industryPopularity:       make(map[string]float64),
	}
}

// RankClassifications performs advanced confidence-based ranking of classifications
func (c *ConfidenceRankingEngine) RankClassifications(classifications []IndustryClassification) []IndustryClassification {
	if len(classifications) == 0 {
		return classifications
	}

	// Calculate enhanced confidence scores
	enhancedClassifications := c.calculateEnhancedConfidenceScores(classifications)

	// Sort by enhanced confidence score (descending)
	sort.Slice(enhancedClassifications, func(i, j int) bool {
		return enhancedClassifications[i].ConfidenceScore > enhancedClassifications[j].ConfidenceScore
	})

	// Remove duplicates (same industry code)
	return c.removeDuplicates(enhancedClassifications)
}

// calculateEnhancedConfidenceScores calculates enhanced confidence scores using multiple factors
func (c *ConfidenceRankingEngine) calculateEnhancedConfidenceScores(classifications []IndustryClassification) []IndustryClassification {
	enhanced := make([]IndustryClassification, len(classifications))

	for i, classification := range classifications {
		enhanced[i] = classification

		// Calculate base confidence
		baseConfidence := classification.ConfidenceScore

		// Calculate method diversity bonus
		methodDiversityBonus := c.calculateMethodDiversityBonus(classifications, i)

		// Calculate consistency bonus
		consistencyBonus := c.calculateConsistencyBonus(classifications, i)

		// Calculate relevance bonus
		relevanceBonus := c.calculateRelevanceBonus(classification)

		// Calculate industry popularity bonus
		popularityBonus := c.calculateIndustryPopularityBonus(classification.IndustryCode)

		// Combine all factors
		enhancedConfidence := (baseConfidence * c.baseConfidenceWeight) +
			(methodDiversityBonus * c.methodDiversityWeight) +
			(consistencyBonus * c.consistencyWeight) +
			(relevanceBonus * c.relevanceWeight) +
			(popularityBonus * c.industryPopularityWeight)

		// Ensure confidence is within valid range [0, 1]
		enhancedConfidence = math.Max(0.0, math.Min(1.0, enhancedConfidence))

		enhanced[i].ConfidenceScore = enhancedConfidence
	}

	return enhanced
}

// calculateMethodDiversityBonus calculates bonus for using diverse classification methods
func (c *ConfidenceRankingEngine) calculateMethodDiversityBonus(classifications []IndustryClassification, currentIndex int) float64 {
	if len(classifications) == 0 {
		return 0.0
	}

	// Count unique methods used
	methods := make(map[string]bool)
	for _, classification := range classifications {
		methods[classification.ClassificationMethod] = true
	}

	// Calculate diversity score
	diversityScore := float64(len(methods)) / float64(len(classifications))

	// Bonus for current classification if it uses a unique method
	currentMethod := classifications[currentIndex].ClassificationMethod
	methodCount := 0
	for _, classification := range classifications {
		if classification.ClassificationMethod == currentMethod {
			methodCount++
		}
	}

	// Higher bonus for methods used less frequently
	uniquenessBonus := 1.0 / float64(methodCount)

	return diversityScore * uniquenessBonus
}

// calculateConsistencyBonus calculates bonus for consistency with other classifications
func (c *ConfidenceRankingEngine) calculateConsistencyBonus(classifications []IndustryClassification, currentIndex int) float64 {
	if len(classifications) < 2 {
		return 1.0 // Perfect consistency for single classification
	}

	current := classifications[currentIndex]
	relatedCount := 0
	totalComparisons := 0

	for i, other := range classifications {
		if i == currentIndex {
			continue
		}

		if c.areIndustriesRelated(current, other) {
			relatedCount++
		}
		totalComparisons++
	}

	if totalComparisons == 0 {
		return 1.0
	}

	return float64(relatedCount) / float64(totalComparisons)
}

// calculateRelevanceBonus calculates bonus based on classification relevance
func (c *ConfidenceRankingEngine) calculateRelevanceBonus(classification IndustryClassification) float64 {
	relevanceScore := 0.0

	// Bonus for having keywords
	if len(classification.Keywords) > 0 {
		relevanceScore += 0.3
	}

	// Bonus for having description
	if classification.Description != "" {
		relevanceScore += 0.2
	}

	// Bonus for high base confidence
	if classification.ConfidenceScore > 0.8 {
		relevanceScore += 0.3
	} else if classification.ConfidenceScore > 0.6 {
		relevanceScore += 0.2
	} else if classification.ConfidenceScore > 0.4 {
		relevanceScore += 0.1
	}

	// Bonus for specific classification methods
	switch classification.ClassificationMethod {
	case "keyword_match":
		relevanceScore += 0.1
	case "description_match":
		relevanceScore += 0.1
	case "fuzzy_match":
		relevanceScore += 0.05
	}

	return math.Min(1.0, relevanceScore)
}

// calculateIndustryPopularityBonus calculates bonus based on industry popularity
func (c *ConfidenceRankingEngine) calculateIndustryPopularityBonus(industryCode string) float64 {
	if popularity, exists := c.industryPopularity[industryCode]; exists {
		return popularity
	}

	// Default to moderate popularity if not found
	return 0.5
}

// areIndustriesRelated checks if two industries are related
func (c *ConfidenceRankingEngine) areIndustriesRelated(industry1, industry2 IndustryClassification) bool {
	// Check if industries are in the same major category
	if c.isSameMajorCategory(industry1.IndustryCode, industry2.IndustryCode) {
		return true
	}

	// Check for keyword overlap
	overlap := c.calculateKeywordOverlap(industry1, industry2)
	return overlap > 0.3 // 30% keyword overlap threshold

	// Additional checks could be added here:
	// - Supply chain relationships
	// - Geographic proximity
	// - Market segment overlap
}

// isSameMajorCategory checks if two industry codes belong to the same major category
func (c *ConfidenceRankingEngine) isSameMajorCategory(code1, code2 string) bool {
	// For NAICS codes, check first 2 digits
	if len(code1) >= 2 && len(code2) >= 2 {
		return code1[:2] == code2[:2]
	}

	// For SIC codes, check first 2 digits
	if len(code1) >= 2 && len(code2) >= 2 {
		return code1[:2] == code2[:2]
	}

	return false
}

// calculateKeywordOverlap calculates the overlap between keywords of two classifications
func (c *ConfidenceRankingEngine) calculateKeywordOverlap(industry1, industry2 IndustryClassification) float64 {
	if len(industry1.Keywords) == 0 || len(industry2.Keywords) == 0 {
		return 0.0
	}

	// Create sets of keywords
	keywords1 := make(map[string]bool)
	for _, keyword := range industry1.Keywords {
		keywords1[strings.ToLower(keyword)] = true
	}

	keywords2 := make(map[string]bool)
	for _, keyword := range industry2.Keywords {
		keywords2[strings.ToLower(keyword)] = true
	}

	// Calculate intersection
	intersection := 0
	for keyword := range keywords1 {
		if keywords2[keyword] {
			intersection++
		}
	}

	// Calculate union
	union := len(keywords1) + len(keywords2) - intersection

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// removeDuplicates removes duplicate classifications based on industry code
func (c *ConfidenceRankingEngine) removeDuplicates(classifications []IndustryClassification) []IndustryClassification {
	seen := make(map[string]bool)
	var unique []IndustryClassification

	for _, classification := range classifications {
		if !seen[classification.IndustryCode] {
			seen[classification.IndustryCode] = true
			unique = append(unique, classification)
		}
	}

	return unique
}

// SetIndustryPopularity sets the popularity data for industries
func (c *ConfidenceRankingEngine) SetIndustryPopularity(popularity map[string]float64) {
	c.industryPopularity = popularity
}

// SetWeights allows customization of ranking weights
func (c *ConfidenceRankingEngine) SetWeights(baseConfidence, methodDiversity, consistency, relevance, popularity float64) {
	c.baseConfidenceWeight = baseConfidence
	c.methodDiversityWeight = methodDiversity
	c.consistencyWeight = consistency
	c.relevanceWeight = relevance
	c.industryPopularityWeight = popularity
}

// GetRankingMetrics returns metrics about the ranking process
func (c *ConfidenceRankingEngine) GetRankingMetrics(classifications []IndustryClassification) map[string]interface{} {
	if len(classifications) == 0 {
		return map[string]interface{}{
			"total_classifications": 0,
			"unique_methods":        0,
			"average_confidence":    0.0,
			"confidence_range":      "0.0-0.0",
		}
	}

	// Calculate metrics
	methods := make(map[string]bool)
	totalConfidence := 0.0
	minConfidence := 1.0
	maxConfidence := 0.0

	for _, classification := range classifications {
		methods[classification.ClassificationMethod] = true
		totalConfidence += classification.ConfidenceScore
		minConfidence = math.Min(minConfidence, classification.ConfidenceScore)
		maxConfidence = math.Max(maxConfidence, classification.ConfidenceScore)
	}

	avgConfidence := totalConfidence / float64(len(classifications))

	return map[string]interface{}{
		"total_classifications": len(classifications),
		"unique_methods":        len(methods),
		"average_confidence":    avgConfidence,
		"confidence_range":      fmt.Sprintf("%.3f-%.3f", minConfidence, maxConfidence),
		"min_confidence":        minConfidence,
		"max_confidence":        maxConfidence,
	}
}
