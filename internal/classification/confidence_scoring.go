package classification

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ConfidenceScoringEngine provides sophisticated industry confidence scoring
type ConfidenceScoringEngine struct {
	logger  *observability.Logger
	metrics *observability.Metrics

	// Configuration
	baseConfidenceWeight   float64
	keywordMatchWeight     float64
	descriptionMatchWeight float64
	businessTypeWeight     float64
	industryHintWeight     float64
	fuzzyMatchWeight       float64
	consistencyWeight      float64
	diversityWeight        float64
	popularityWeight       float64
	recencyWeight          float64
	validationWeight       float64

	// Industry-specific scoring data
	industryPopularity      map[string]float64
	industryValidationRates map[string]float64
	keywordReliability      map[string]float64
}

// NewConfidenceScoringEngine creates a new confidence scoring engine
func NewConfidenceScoringEngine(logger *observability.Logger, metrics *observability.Metrics) *ConfidenceScoringEngine {
	return &ConfidenceScoringEngine{
		logger:  logger,
		metrics: metrics,

		// Configuration weights
		baseConfidenceWeight:   0.25,
		keywordMatchWeight:     0.20,
		descriptionMatchWeight: 0.15,
		businessTypeWeight:     0.10,
		industryHintWeight:     0.10,
		fuzzyMatchWeight:       0.05,
		consistencyWeight:      0.05,
		diversityWeight:        0.03,
		popularityWeight:       0.03,
		recencyWeight:          0.02,
		validationWeight:       0.02,

		// Initialize data maps
		industryPopularity:      make(map[string]float64),
		industryValidationRates: make(map[string]float64),
		keywordReliability:      make(map[string]float64),
	}
}

// ConfidenceScore represents a comprehensive confidence score with breakdown
type ConfidenceScore struct {
	OverallScore       float64   `json:"overall_score"`
	BaseConfidence     float64   `json:"base_confidence"`
	KeywordScore       float64   `json:"keyword_score"`
	DescriptionScore   float64   `json:"description_score"`
	BusinessTypeScore  float64   `json:"business_type_score"`
	IndustryHintScore  float64   `json:"industry_hint_score"`
	FuzzyMatchScore    float64   `json:"fuzzy_match_score"`
	ConsistencyScore   float64   `json:"consistency_score"`
	DiversityScore     float64   `json:"diversity_score"`
	PopularityScore    float64   `json:"popularity_score"`
	RecencyScore       float64   `json:"recency_score"`
	ValidationScore    float64   `json:"validation_score"`
	ConfidenceLevel    string    `json:"confidence_level"`
	ReliabilityFactors []string  `json:"reliability_factors"`
	UncertaintyFactors []string  `json:"uncertainty_factors"`
	ScoringTimestamp   time.Time `json:"scoring_timestamp"`
	ScoringMethod      string    `json:"scoring_method"`
}

// CalculateConfidenceScore calculates comprehensive confidence score for an industry classification
func (c *ConfidenceScoringEngine) CalculateConfidenceScore(ctx context.Context, classification IndustryClassification, request *ClassificationRequest, allClassifications []IndustryClassification) *ConfidenceScore {
	start := time.Now()

	// Log scoring start
	if c.logger != nil {
		c.logger.WithComponent("confidence_scoring").LogBusinessEvent(ctx, "confidence_scoring_started", "", map[string]interface{}{
			"industry_code": classification.IndustryCode,
			"method":        classification.ClassificationMethod,
		})
	}

	// Calculate individual component scores
	baseConfidence := c.calculateBaseConfidence(classification)
	keywordScore := c.calculateKeywordScore(classification, request)
	descriptionScore := c.calculateDescriptionScore(classification, request)
	businessTypeScore := c.calculateBusinessTypeScore(classification, request)
	industryHintScore := c.calculateIndustryHintScore(classification, request)
	fuzzyMatchScore := c.calculateFuzzyMatchScore(classification, request)
	consistencyScore := c.calculateConsistencyScore(classification, allClassifications)
	diversityScore := c.calculateDiversityScore(classification, allClassifications)
	popularityScore := c.calculatePopularityScore(classification)
	recencyScore := c.calculateRecencyScore(classification)
	validationScore := c.calculateValidationScore(classification)

	// Calculate weighted overall score
	overallScore := (baseConfidence * c.baseConfidenceWeight) +
		(keywordScore * c.keywordMatchWeight) +
		(descriptionScore * c.descriptionMatchWeight) +
		(businessTypeScore * c.businessTypeWeight) +
		(industryHintScore * c.industryHintWeight) +
		(fuzzyMatchScore * c.fuzzyMatchWeight) +
		(consistencyScore * c.consistencyWeight) +
		(diversityScore * c.diversityWeight) +
		(popularityScore * c.popularityWeight) +
		(recencyScore * c.recencyWeight) +
		(validationScore * c.validationWeight)

	// Ensure score is within valid range
	overallScore = math.Max(0.0, math.Min(1.0, overallScore))

	// Determine confidence level
	confidenceLevel := c.determineConfidenceLevel(overallScore)

	// Identify reliability and uncertainty factors
	reliabilityFactors := c.identifyReliabilityFactors(classification, request)
	uncertaintyFactors := c.identifyUncertaintyFactors(classification, request)

	// Create confidence score result
	result := &ConfidenceScore{
		OverallScore:       overallScore,
		BaseConfidence:     baseConfidence,
		KeywordScore:       keywordScore,
		DescriptionScore:   descriptionScore,
		BusinessTypeScore:  businessTypeScore,
		IndustryHintScore:  industryHintScore,
		FuzzyMatchScore:    fuzzyMatchScore,
		ConsistencyScore:   consistencyScore,
		DiversityScore:     diversityScore,
		PopularityScore:    popularityScore,
		RecencyScore:       recencyScore,
		ValidationScore:    validationScore,
		ConfidenceLevel:    confidenceLevel,
		ReliabilityFactors: reliabilityFactors,
		UncertaintyFactors: uncertaintyFactors,
		ScoringTimestamp:   time.Now(),
		ScoringMethod:      "enhanced_confidence_scoring",
	}

	// Log scoring completion
	if c.logger != nil {
		c.logger.WithComponent("confidence_scoring").LogBusinessEvent(ctx, "confidence_scoring_completed", "", map[string]interface{}{
			"industry_code":      classification.IndustryCode,
			"overall_score":      overallScore,
			"confidence_level":   confidenceLevel,
			"processing_time_ms": time.Since(start).Milliseconds(),
		})
	}

	return result
}

// calculateBaseConfidence calculates the base confidence score
func (c *ConfidenceScoringEngine) calculateBaseConfidence(classification IndustryClassification) float64 {
	baseScore := classification.ConfidenceScore

	// Apply method-specific adjustments
	switch classification.ClassificationMethod {
	case "keyword_match":
		baseScore *= 1.1 // Boost for keyword matches
	case "description_match":
		baseScore *= 1.05 // Slight boost for description matches
	case "fuzzy_match":
		baseScore *= 0.9 // Reduce for fuzzy matches
	case "business_type":
		baseScore *= 0.95 // Slight reduction for business type
	case "industry_hint":
		baseScore *= 0.9 // Reduce for industry hints
	}

	return math.Max(0.0, math.Min(1.0, baseScore))
}

// calculateKeywordScore calculates keyword-based confidence score
func (c *ConfidenceScoringEngine) calculateKeywordScore(classification IndustryClassification, request *ClassificationRequest) float64 {
	if request.Keywords == "" || len(classification.Keywords) == 0 {
		return 0.0
	}

	// Extract keywords from request
	requestKeywords := strings.Split(strings.ToLower(request.Keywords), ",")
	for i, keyword := range requestKeywords {
		requestKeywords[i] = strings.TrimSpace(keyword)
	}

	// Calculate keyword overlap
	matches := 0
	totalRequestKeywords := len(requestKeywords)

	for _, requestKeyword := range requestKeywords {
		for _, classificationKeyword := range classification.Keywords {
			if strings.Contains(strings.ToLower(classificationKeyword), requestKeyword) ||
				strings.Contains(requestKeyword, strings.ToLower(classificationKeyword)) {
				matches++
				break
			}
		}
	}

	if totalRequestKeywords == 0 {
		return 0.0
	}

	// Calculate base keyword score
	keywordScore := float64(matches) / float64(totalRequestKeywords)

	// Apply keyword reliability factor
	reliabilityFactor := c.getKeywordReliabilityFactor(classification.IndustryCode)
	keywordScore *= reliabilityFactor

	return keywordScore
}

// calculateDescriptionScore calculates description-based confidence score
func (c *ConfidenceScoringEngine) calculateDescriptionScore(classification IndustryClassification, request *ClassificationRequest) float64 {
	if request.Description == "" {
		return 0.0
	}

	// Calculate text similarity between description and industry keywords
	description := strings.ToLower(request.Description)
	keywordMatches := 0
	totalKeywords := len(classification.Keywords)

	if totalKeywords == 0 {
		return 0.0
	}

	for _, keyword := range classification.Keywords {
		if strings.Contains(description, strings.ToLower(keyword)) {
			keywordMatches++
		}
	}

	// Calculate base description score
	descriptionScore := float64(keywordMatches) / float64(totalKeywords)

	// Apply length factor (longer descriptions get higher scores)
	lengthFactor := math.Min(1.0, float64(len(description))/100.0)
	descriptionScore *= (0.7 + 0.3*lengthFactor)

	return descriptionScore
}

// calculateBusinessTypeScore calculates business type-based confidence score
func (c *ConfidenceScoringEngine) calculateBusinessTypeScore(classification IndustryClassification, request *ClassificationRequest) float64 {
	if request.BusinessType == "" {
		return 0.0
	}

	// Map business types to confidence scores
	businessTypeScores := map[string]float64{
		"corporation":         0.8,
		"llc":                 0.7,
		"partnership":         0.6,
		"sole_proprietorship": 0.5,
		"nonprofit":           0.9,
		"government":          0.9,
		"cooperative":         0.6,
	}

	businessType := strings.ToLower(request.BusinessType)
	if score, exists := businessTypeScores[businessType]; exists {
		return score
	}

	// Default score for unknown business types
	return 0.5
}

// calculateIndustryHintScore calculates industry hint-based confidence score
func (c *ConfidenceScoringEngine) calculateIndustryHintScore(classification IndustryClassification, request *ClassificationRequest) float64 {
	if request.Industry == "" {
		return 0.0
	}

	// Check if the provided industry hint matches the classification
	industryHint := strings.ToLower(request.Industry)
	industryName := strings.ToLower(classification.IndustryName)

	// Direct match
	if strings.Contains(industryName, industryHint) || strings.Contains(industryHint, industryName) {
		return 0.9
	}

	// Check keyword matches
	for _, keyword := range classification.Keywords {
		if strings.Contains(industryHint, strings.ToLower(keyword)) {
			return 0.7
		}
	}

	// No match
	return 0.0
}

// calculateFuzzyMatchScore calculates fuzzy match-based confidence score
func (c *ConfidenceScoringEngine) calculateFuzzyMatchScore(classification IndustryClassification, request *ClassificationRequest) float64 {
	if request.BusinessName == "" {
		return 0.0
	}

	// Calculate similarity between business name and industry name
	businessName := strings.ToLower(request.BusinessName)
	industryName := strings.ToLower(classification.IndustryName)

	// Simple similarity calculation (can be enhanced with more sophisticated algorithms)
	similarity := c.calculateStringSimilarity(businessName, industryName)

	// Apply fuzzy match penalty
	fuzzyScore := similarity * 0.8

	return fuzzyScore
}

// calculateConsistencyScore calculates consistency with other classifications
func (c *ConfidenceScoringEngine) calculateConsistencyScore(classification IndustryClassification, allClassifications []IndustryClassification) float64 {
	if len(allClassifications) < 2 {
		return 1.0
	}

	// Count related classifications
	relatedCount := 0
	totalComparisons := 0

	for _, other := range allClassifications {
		if other.IndustryCode == classification.IndustryCode {
			continue
		}

		if c.areIndustriesRelated(classification, other) {
			relatedCount++
		}
		totalComparisons++
	}

	if totalComparisons == 0 {
		return 1.0
	}

	return float64(relatedCount) / float64(totalComparisons)
}

// calculateDiversityScore calculates diversity score
func (c *ConfidenceScoringEngine) calculateDiversityScore(classification IndustryClassification, allClassifications []IndustryClassification) float64 {
	if len(allClassifications) < 2 {
		return 1.0
	}

	// Count unique major categories
	categories := make(map[string]bool)
	for _, other := range allClassifications {
		if len(other.IndustryCode) >= 2 {
			categories[other.IndustryCode[:2]] = true
		}
	}

	// Higher diversity means this classification is more unique
	diversityRatio := float64(len(categories)) / float64(len(allClassifications))
	return diversityRatio
}

// calculatePopularityScore calculates industry popularity score
func (c *ConfidenceScoringEngine) calculatePopularityScore(classification IndustryClassification) float64 {
	if popularity, exists := c.industryPopularity[classification.IndustryCode]; exists {
		return popularity
	}

	// Default to moderate popularity
	return 0.5
}

// calculateRecencyScore calculates recency score
func (c *ConfidenceScoringEngine) calculateRecencyScore(classification IndustryClassification) float64 {
	// This could be based on when the industry classification was last updated
	// For now, return a default score
	return 0.8
}

// calculateValidationScore calculates validation score
func (c *ConfidenceScoringEngine) calculateValidationScore(classification IndustryClassification) float64 {
	if validationRate, exists := c.industryValidationRates[classification.IndustryCode]; exists {
		return validationRate
	}

	// Default validation rate
	return 0.7
}

// determineConfidenceLevel determines the confidence level based on overall score
func (c *ConfidenceScoringEngine) determineConfidenceLevel(overallScore float64) string {
	switch {
	case overallScore >= 0.9:
		return "very_high"
	case overallScore >= 0.8:
		return "high"
	case overallScore >= 0.7:
		return "medium_high"
	case overallScore >= 0.6:
		return "medium"
	case overallScore >= 0.5:
		return "medium_low"
	case overallScore >= 0.3:
		return "low"
	default:
		return "very_low"
	}
}

// identifyReliabilityFactors identifies factors that increase confidence
func (c *ConfidenceScoringEngine) identifyReliabilityFactors(classification IndustryClassification, request *ClassificationRequest) []string {
	var factors []string

	// Check for strong keyword matches
	if request.Keywords != "" && len(classification.Keywords) > 0 {
		factors = append(factors, "strong_keyword_matches")
	}

	// Check for detailed description
	if request.Description != "" && len(request.Description) > 50 {
		factors = append(factors, "detailed_description")
	}

	// Check for specific business type
	if request.BusinessType != "" {
		factors = append(factors, "specific_business_type")
	}

	// Check for industry hint
	if request.Industry != "" {
		factors = append(factors, "industry_hint_provided")
	}

	// Check for high base confidence
	if classification.ConfidenceScore > 0.8 {
		factors = append(factors, "high_base_confidence")
	}

	// Check for reliable classification method
	if classification.ClassificationMethod == "keyword_match" || classification.ClassificationMethod == "description_match" {
		factors = append(factors, "reliable_classification_method")
	}

	return factors
}

// identifyUncertaintyFactors identifies factors that decrease confidence
func (c *ConfidenceScoringEngine) identifyUncertaintyFactors(classification IndustryClassification, request *ClassificationRequest) []string {
	var factors []string

	// Check for missing keywords
	if request.Keywords == "" {
		factors = append(factors, "no_keywords_provided")
	}

	// Check for missing description
	if request.Description == "" {
		factors = append(factors, "no_description_provided")
	}

	// Check for low base confidence
	if classification.ConfidenceScore < 0.5 {
		factors = append(factors, "low_base_confidence")
	}

	// Check for unreliable classification method
	if classification.ClassificationMethod == "fuzzy_match" {
		factors = append(factors, "fuzzy_match_method")
	}

	// Check for generic business type
	if request.BusinessType == "" {
		factors = append(factors, "no_business_type_specified")
	}

	return factors
}

// calculateStringSimilarity calculates similarity between two strings
func (c *ConfidenceScoringEngine) calculateStringSimilarity(str1, str2 string) float64 {
	// Simple Jaccard similarity implementation
	words1 := strings.Fields(str1)
	words2 := strings.Fields(str2)

	if len(words1) == 0 && len(words2) == 0 {
		return 1.0
	}

	// Create word sets
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, word := range words1 {
		set1[word] = true
	}
	for _, word := range words2 {
		set2[word] = true
	}

	// Calculate intersection
	intersection := 0
	for word := range set1 {
		if set2[word] {
			intersection++
		}
	}

	// Calculate union
	union := len(set1) + len(set2) - intersection

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// areIndustriesRelated checks if two industries are related
func (c *ConfidenceScoringEngine) areIndustriesRelated(industry1, industry2 IndustryClassification) bool {
	// Check if industries are in the same major category
	if len(industry1.IndustryCode) >= 2 && len(industry2.IndustryCode) >= 2 {
		if industry1.IndustryCode[:2] == industry2.IndustryCode[:2] {
			return true
		}
	}

	// Check for keyword overlap
	if len(industry1.Keywords) > 0 && len(industry2.Keywords) > 0 {
		overlap := c.calculateKeywordOverlap(industry1.Keywords, industry2.Keywords)
		return overlap > 0.3
	}

	return false
}

// calculateKeywordOverlap calculates overlap between keyword sets
func (c *ConfidenceScoringEngine) calculateKeywordOverlap(keywords1, keywords2 []string) float64 {
	if len(keywords1) == 0 || len(keywords2) == 0 {
		return 0.0
	}

	// Create keyword sets
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, keyword := range keywords1 {
		set1[strings.ToLower(keyword)] = true
	}
	for _, keyword := range keywords2 {
		set2[strings.ToLower(keyword)] = true
	}

	// Calculate intersection
	intersection := 0
	for keyword := range set1 {
		if set2[keyword] {
			intersection++
		}
	}

	// Calculate union
	union := len(set1) + len(set2) - intersection

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// getKeywordReliabilityFactor gets the reliability factor for keywords in an industry
func (c *ConfidenceScoringEngine) getKeywordReliabilityFactor(industryCode string) float64 {
	if reliability, exists := c.keywordReliability[industryCode]; exists {
		return reliability
	}

	// Default reliability factor
	return 0.8
}

// SetIndustryData sets industry-specific scoring data
func (c *ConfidenceScoringEngine) SetIndustryData(popularity, validationRates, keywordReliability map[string]float64) {
	c.industryPopularity = popularity
	c.industryValidationRates = validationRates
	c.keywordReliability = keywordReliability
}

// SetWeights allows customization of scoring weights
func (c *ConfidenceScoringEngine) SetWeights(weights map[string]float64) {
	if baseConfidence, exists := weights["base_confidence"]; exists {
		c.baseConfidenceWeight = baseConfidence
	}
	if keywordMatch, exists := weights["keyword_match"]; exists {
		c.keywordMatchWeight = keywordMatch
	}
	if descriptionMatch, exists := weights["description_match"]; exists {
		c.descriptionMatchWeight = descriptionMatch
	}
	if businessType, exists := weights["business_type"]; exists {
		c.businessTypeWeight = businessType
	}
	if industryHint, exists := weights["industry_hint"]; exists {
		c.industryHintWeight = industryHint
	}
	if fuzzyMatch, exists := weights["fuzzy_match"]; exists {
		c.fuzzyMatchWeight = fuzzyMatch
	}
	if consistency, exists := weights["consistency"]; exists {
		c.consistencyWeight = consistency
	}
	if diversity, exists := weights["diversity"]; exists {
		c.diversityWeight = diversity
	}
	if popularity, exists := weights["popularity"]; exists {
		c.popularityWeight = popularity
	}
	if recency, exists := weights["recency"]; exists {
		c.recencyWeight = recency
	}
	if validation, exists := weights["validation"]; exists {
		c.validationWeight = validation
	}
}

// GetWeights returns current scoring weights
func (c *ConfidenceScoringEngine) GetWeights() map[string]float64 {
	return map[string]float64{
		"base_confidence":   c.baseConfidenceWeight,
		"keyword_match":     c.keywordMatchWeight,
		"description_match": c.descriptionMatchWeight,
		"business_type":     c.businessTypeWeight,
		"industry_hint":     c.industryHintWeight,
		"fuzzy_match":       c.fuzzyMatchWeight,
		"consistency":       c.consistencyWeight,
		"diversity":         c.diversityWeight,
		"popularity":        c.popularityWeight,
		"recency":           c.recencyWeight,
		"validation":        c.validationWeight,
	}
}
