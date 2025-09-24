package repository

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"
)

// EnhancedScoringAlgorithm provides sophisticated scoring that combines direct matches, phrase matches, and context multipliers
type EnhancedScoringAlgorithm struct {
	logger *log.Logger
	config *EnhancedScoringConfig
}

// EnhancedScoringConfig holds configuration for the enhanced scoring algorithm
type EnhancedScoringConfig struct {
	// Scoring weights
	DirectMatchWeight  float64 `json:"direct_match_weight"`  // Weight for exact keyword matches
	PhraseMatchWeight  float64 `json:"phrase_match_weight"`  // Weight for phrase matches
	PartialMatchWeight float64 `json:"partial_match_weight"` // Weight for partial matches
	ContextWeight      float64 `json:"context_weight"`       // Weight for context multipliers

	// Performance optimization
	MaxKeywordsToProcess int  `json:"max_keywords_to_process"` // Limit keywords for performance
	CacheResults         bool `json:"cache_results"`           // Enable result caching
	ParallelProcessing   bool `json:"parallel_processing"`     // Enable parallel processing

	// Quality thresholds
	MinMatchThreshold   float64 `json:"min_match_threshold"`  // Minimum score to consider
	ConfidenceThreshold float64 `json:"confidence_threshold"` // Minimum confidence threshold

	// Advanced features
	EnableFuzzyMatching bool `json:"enable_fuzzy_matching"` // Enable fuzzy string matching
	EnableSemanticBoost bool `json:"enable_semantic_boost"` // Enable semantic similarity boost
	EnableIndustryBoost bool `json:"enable_industry_boost"` // Enable industry-specific boosts
}

// DefaultEnhancedScoringConfig returns the default configuration for enhanced scoring
func DefaultEnhancedScoringConfig() *EnhancedScoringConfig {
	return &EnhancedScoringConfig{
		// Scoring weights (must sum to 1.0)
		DirectMatchWeight:  0.40, // 40% weight for exact matches
		PhraseMatchWeight:  0.30, // 30% weight for phrase matches
		PartialMatchWeight: 0.20, // 20% weight for partial matches
		ContextWeight:      0.10, // 10% weight for context multipliers

		// Performance optimization
		MaxKeywordsToProcess: 1000, // Process up to 1000 keywords
		CacheResults:         true, // Enable caching for performance
		ParallelProcessing:   true, // Enable parallel processing

		// Quality thresholds
		MinMatchThreshold:   0.1, // Minimum score to consider
		ConfidenceThreshold: 0.5, // Minimum confidence threshold

		// Advanced features
		EnableFuzzyMatching: true, // Enable fuzzy matching
		EnableSemanticBoost: true, // Enable semantic boost
		EnableIndustryBoost: true, // Enable industry boost
	}
}

// EnhancedScoringResult represents the result of enhanced scoring
type EnhancedScoringResult struct {
	IndustryID         int                 `json:"industry_id"`
	IndustryName       string              `json:"industry_name"`
	TotalScore         float64             `json:"total_score"`
	Confidence         float64             `json:"confidence"`
	ScoreBreakdown     *ScoreBreakdown     `json:"score_breakdown"`
	MatchedKeywords    []MatchedKeyword    `json:"matched_keywords"`
	ProcessingTime     time.Duration       `json:"processing_time"`
	PerformanceMetrics *PerformanceMetrics `json:"performance_metrics"`
	QualityIndicators  *QualityIndicators  `json:"quality_indicators"`
	CreatedAt          time.Time           `json:"created_at"`
}

// ScoreBreakdown provides detailed breakdown of scoring components
type ScoreBreakdown struct {
	DirectMatchScore   float64 `json:"direct_match_score"`
	PhraseMatchScore   float64 `json:"phrase_match_score"`
	PartialMatchScore  float64 `json:"partial_match_score"`
	ContextScore       float64 `json:"context_score"`
	FuzzyMatchScore    float64 `json:"fuzzy_match_score"`
	SemanticBoostScore float64 `json:"semantic_boost_score"`
	IndustryBoostScore float64 `json:"industry_boost_score"`
	TotalWeightedScore float64 `json:"total_weighted_score"`
}

// MatchedKeyword represents a keyword match with detailed information
type MatchedKeyword struct {
	Keyword           string  `json:"keyword"`
	MatchedKeyword    string  `json:"matched_keyword"`
	MatchType         string  `json:"match_type"` // "direct", "phrase", "partial", "fuzzy"
	BaseWeight        float64 `json:"base_weight"`
	ContextMultiplier float64 `json:"context_multiplier"`
	FinalWeight       float64 `json:"final_weight"`
	Confidence        float64 `json:"confidence"`
	Source            string  `json:"source"` // "business_name", "description", "website_url"
}

// PerformanceMetrics tracks performance-related metrics
type PerformanceMetrics struct {
	KeywordsProcessed int           `json:"keywords_processed"`
	MatchesFound      int           `json:"matches_found"`
	ProcessingTime    time.Duration `json:"processing_time"`
	MemoryUsage       int64         `json:"memory_usage"`
	CacheHits         int           `json:"cache_hits"`
	CacheMisses       int           `json:"cache_misses"`
}

// QualityIndicators provides quality assessment of the scoring result
type QualityIndicators struct {
	MatchDiversity      float64 `json:"match_diversity"`      // Diversity of match types
	KeywordRelevance    float64 `json:"keyword_relevance"`    // Relevance of matched keywords
	ContextConsistency  float64 `json:"context_consistency"`  // Consistency across contexts
	ConfidenceStability float64 `json:"confidence_stability"` // Stability of confidence score
	OverallQuality      float64 `json:"overall_quality"`      // Overall quality score
}

// NewEnhancedScoringAlgorithm creates a new enhanced scoring algorithm instance
func NewEnhancedScoringAlgorithm(logger *log.Logger, config *EnhancedScoringConfig) *EnhancedScoringAlgorithm {
	if config == nil {
		config = DefaultEnhancedScoringConfig()
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		logger.Printf("⚠️ Invalid configuration, using defaults: %v", err)
		config = DefaultEnhancedScoringConfig()
	}

	return &EnhancedScoringAlgorithm{
		logger: logger,
		config: config,
	}
}

// Validate validates the enhanced scoring configuration
func (config *EnhancedScoringConfig) Validate() error {
	// Check that weights sum to approximately 1.0
	totalWeight := config.DirectMatchWeight + config.PhraseMatchWeight +
		config.PartialMatchWeight + config.ContextWeight

	if math.Abs(totalWeight-1.0) > 0.01 {
		return fmt.Errorf("scoring weights must sum to 1.0, got %.3f", totalWeight)
	}

	// Check threshold values
	if config.MinMatchThreshold < 0 || config.MinMatchThreshold > 1 {
		return fmt.Errorf("min match threshold must be between 0 and 1, got %.3f", config.MinMatchThreshold)
	}

	if config.ConfidenceThreshold < 0 || config.ConfidenceThreshold > 1 {
		return fmt.Errorf("confidence threshold must be between 0 and 1, got %.3f", config.ConfidenceThreshold)
	}

	// Check performance limits
	if config.MaxKeywordsToProcess <= 0 {
		return fmt.Errorf("max keywords to process must be positive, got %d", config.MaxKeywordsToProcess)
	}

	return nil
}

// CalculateEnhancedScore calculates the enhanced score for a business classification
func (esa *EnhancedScoringAlgorithm) CalculateEnhancedScore(
	ctx context.Context,
	contextualKeywords []ContextualKeyword,
	keywordIndex *KeywordIndex,
) (*EnhancedScoringResult, error) {
	startTime := time.Now()
	requestID := esa.generateRequestID()

	esa.logger.Printf("🚀 Starting enhanced scoring calculation (request: %s)", requestID)

	// Validate inputs
	if len(contextualKeywords) == 0 {
		return nil, fmt.Errorf("no contextual keywords provided")
	}

	if keywordIndex == nil || len(keywordIndex.KeywordToIndustries) == 0 {
		return nil, fmt.Errorf("keyword index is empty or nil")
	}

	// Limit keywords for performance if configured
	if len(contextualKeywords) > esa.config.MaxKeywordsToProcess {
		contextualKeywords = contextualKeywords[:esa.config.MaxKeywordsToProcess]
		esa.logger.Printf("⚠️ Limited keywords to %d for performance", esa.config.MaxKeywordsToProcess)
	}

	// Initialize performance metrics
	performanceMetrics := &PerformanceMetrics{
		KeywordsProcessed: len(contextualKeywords),
		ProcessingTime:    0,
		MemoryUsage:       0,
		CacheHits:         0,
		CacheMisses:       0,
	}

	// Calculate scores for all industries
	industryScores := make(map[int]*IndustryScore)

	// Process keywords with enhanced algorithm
	for _, contextualKeyword := range contextualKeywords {
		matches := esa.findEnhancedMatches(contextualKeyword, keywordIndex)
		esa.updateIndustryScores(industryScores, matches, contextualKeyword)
		performanceMetrics.MatchesFound += len(matches)
	}

	// Find best industry
	bestIndustryID, bestScore := esa.findBestIndustry(industryScores)

	// Get the best industry score (may be nil if no matches)
	bestIndustryScore := industryScores[bestIndustryID]

	// Calculate detailed score breakdown
	scoreBreakdown := esa.calculateScoreBreakdown(bestIndustryScore)

	// Calculate confidence
	confidence := esa.calculateEnhancedConfidence(bestScore, scoreBreakdown, len(contextualKeywords))

	// Get industry name
	industryName := esa.getIndustryName(bestIndustryID, keywordIndex)

	// Calculate quality indicators
	qualityIndicators := esa.calculateQualityIndicators(bestIndustryScore, contextualKeywords)

	// Update performance metrics
	performanceMetrics.ProcessingTime = time.Since(startTime)

	// Create result
	result := &EnhancedScoringResult{
		IndustryID:         bestIndustryID,
		IndustryName:       industryName,
		TotalScore:         bestScore,
		Confidence:         confidence,
		ScoreBreakdown:     scoreBreakdown,
		MatchedKeywords:    esa.extractMatchedKeywords(bestIndustryScore),
		ProcessingTime:     performanceMetrics.ProcessingTime,
		PerformanceMetrics: performanceMetrics,
		QualityIndicators:  qualityIndicators,
		CreatedAt:          time.Now(),
	}

	esa.logger.Printf("✅ Enhanced scoring completed: %s (score: %.3f, confidence: %.3f) (request: %s)",
		industryName, bestScore, confidence, requestID)

	return result, nil
}

// IndustryScore holds detailed scoring information for an industry
type IndustryScore struct {
	IndustryID         int
	TotalScore         float64
	DirectMatches      []KeywordMatch
	PhraseMatches      []KeywordMatch
	PartialMatches     []KeywordMatch
	FuzzyMatches       []KeywordMatch
	ContextMultipliers map[string]float64
	MatchCount         int
	UniqueKeywords     map[string]bool
}

// KeywordMatch represents a single keyword match
type KeywordMatch struct {
	InputKeyword      string
	MatchedKeyword    string
	MatchType         string
	BaseWeight        float64
	ContextMultiplier float64
	FinalWeight       float64
	Confidence        float64
	Source            string
	IndustryID        int
}

// findEnhancedMatches finds all types of matches for a contextual keyword
func (esa *EnhancedScoringAlgorithm) findEnhancedMatches(
	contextualKeyword ContextualKeyword,
	keywordIndex *KeywordIndex,
) []KeywordMatch {
	var matches []KeywordMatch
	normalizedKeyword := strings.ToLower(strings.TrimSpace(contextualKeyword.Keyword))

	// 1. Direct matches (exact keyword matches)
	directMatches := esa.findDirectMatches(normalizedKeyword, contextualKeyword, keywordIndex)
	matches = append(matches, directMatches...)

	// 2. Phrase matches (multi-word phrase matches)
	phraseMatches := esa.findPhraseMatches(normalizedKeyword, contextualKeyword, keywordIndex)
	matches = append(matches, phraseMatches...)

	// 3. Partial matches (substring matches)
	partialMatches := esa.findPartialMatches(normalizedKeyword, contextualKeyword, keywordIndex)
	matches = append(matches, partialMatches...)

	// 4. Fuzzy matches (if enabled)
	if esa.config.EnableFuzzyMatching {
		fuzzyMatches := esa.findFuzzyMatches(normalizedKeyword, contextualKeyword, keywordIndex)
		matches = append(matches, fuzzyMatches...)
	}

	return matches
}

// findDirectMatches finds exact keyword matches
func (esa *EnhancedScoringAlgorithm) findDirectMatches(
	normalizedKeyword string,
	contextualKeyword ContextualKeyword,
	keywordIndex *KeywordIndex,
) []KeywordMatch {
	var matches []KeywordMatch

	if industryMatches, exists := keywordIndex.KeywordToIndustries[normalizedKeyword]; exists {
		for _, match := range industryMatches {
			contextMultiplier := esa.getContextMultiplier(contextualKeyword.Context)
			finalWeight := match.Weight * contextMultiplier

			matches = append(matches, KeywordMatch{
				InputKeyword:      normalizedKeyword,
				MatchedKeyword:    match.Keyword,
				MatchType:         "direct",
				BaseWeight:        match.Weight,
				ContextMultiplier: contextMultiplier,
				FinalWeight:       finalWeight,
				Confidence:        esa.calculateMatchConfidence("direct", match.Weight, contextMultiplier),
				Source:            contextualKeyword.Context,
				IndustryID:        match.IndustryID,
			})
		}
	}

	return matches
}

// findPhraseMatches finds phrase matches with enhanced phrase detection
func (esa *EnhancedScoringAlgorithm) findPhraseMatches(
	normalizedKeyword string,
	contextualKeyword ContextualKeyword,
	keywordIndex *KeywordIndex,
) []KeywordMatch {
	var matches []KeywordMatch

	// Check if input is a phrase (contains spaces)
	if !strings.Contains(normalizedKeyword, " ") {
		return matches // Not a phrase, skip phrase matching
	}

	// Find exact phrase matches
	if industryMatches, exists := keywordIndex.KeywordToIndustries[normalizedKeyword]; exists {
		for _, match := range industryMatches {
			contextMultiplier := esa.getContextMultiplier(contextualKeyword.Context)
			phraseMultiplier := 1.5 // 50% boost for phrase matches
			finalWeight := match.Weight * phraseMultiplier * contextMultiplier

			matches = append(matches, KeywordMatch{
				InputKeyword:      normalizedKeyword,
				MatchedKeyword:    match.Keyword,
				MatchType:         "phrase",
				BaseWeight:        match.Weight,
				ContextMultiplier: contextMultiplier,
				FinalWeight:       finalWeight,
				Confidence:        esa.calculateMatchConfidence("phrase", match.Weight, contextMultiplier),
				Source:            contextualKeyword.Context,
				IndustryID:        match.IndustryID,
			})
		}
	}

	// Find phrase-to-phrase partial matches
	for keyword, industryMatches := range keywordIndex.KeywordToIndustries {
		if strings.Contains(keyword, " ") && keyword != normalizedKeyword {
			// Both are phrases - check for phrase overlap
			if esa.hasPhraseOverlap(normalizedKeyword, keyword) {
				for _, match := range industryMatches {
					contextMultiplier := esa.getContextMultiplier(contextualKeyword.Context)
					phraseMultiplier := 0.8 // Reduced weight for partial phrase matches
					finalWeight := match.Weight * phraseMultiplier * contextMultiplier

					matches = append(matches, KeywordMatch{
						InputKeyword:      normalizedKeyword,
						MatchedKeyword:    match.Keyword,
						MatchType:         "phrase_partial",
						BaseWeight:        match.Weight,
						ContextMultiplier: contextMultiplier,
						FinalWeight:       finalWeight,
						Confidence:        esa.calculateMatchConfidence("phrase_partial", match.Weight, contextMultiplier),
						Source:            contextualKeyword.Context,
						IndustryID:        match.IndustryID,
					})
				}
			}
		}
	}

	return matches
}

// findPartialMatches finds partial (substring) matches
func (esa *EnhancedScoringAlgorithm) findPartialMatches(
	normalizedKeyword string,
	contextualKeyword ContextualKeyword,
	keywordIndex *KeywordIndex,
) []KeywordMatch {
	var matches []KeywordMatch

	for keyword, industryMatches := range keywordIndex.KeywordToIndustries {
		// Skip exact matches (already handled by direct matches)
		if keyword == normalizedKeyword {
			continue
		}

		// Check for substring matches
		if strings.Contains(normalizedKeyword, keyword) || strings.Contains(keyword, normalizedKeyword) {
			for _, match := range industryMatches {
				contextMultiplier := esa.getContextMultiplier(contextualKeyword.Context)
				partialMultiplier := 0.5 // Reduced weight for partial matches
				finalWeight := match.Weight * partialMultiplier * contextMultiplier

				matches = append(matches, KeywordMatch{
					InputKeyword:      normalizedKeyword,
					MatchedKeyword:    match.Keyword,
					MatchType:         "partial",
					BaseWeight:        match.Weight,
					ContextMultiplier: contextMultiplier,
					FinalWeight:       finalWeight,
					Confidence:        esa.calculateMatchConfidence("partial", match.Weight, contextMultiplier),
					Source:            contextualKeyword.Context,
					IndustryID:        match.IndustryID,
				})
			}
		}
	}

	return matches
}

// findFuzzyMatches finds fuzzy string matches (if enabled)
func (esa *EnhancedScoringAlgorithm) findFuzzyMatches(
	normalizedKeyword string,
	contextualKeyword ContextualKeyword,
	keywordIndex *KeywordIndex,
) []KeywordMatch {
	var matches []KeywordMatch

	// Simple fuzzy matching based on edit distance
	for keyword, industryMatches := range keywordIndex.KeywordToIndustries {
		if keyword == normalizedKeyword {
			continue // Skip exact matches
		}

		// Calculate similarity score
		similarity := esa.calculateStringSimilarity(normalizedKeyword, keyword)

		// Only consider matches with similarity > 0.7
		if similarity > 0.7 {
			for _, match := range industryMatches {
				contextMultiplier := esa.getContextMultiplier(contextualKeyword.Context)
				fuzzyMultiplier := similarity * 0.3 // Reduced weight for fuzzy matches
				finalWeight := match.Weight * fuzzyMultiplier * contextMultiplier

				matches = append(matches, KeywordMatch{
					InputKeyword:      normalizedKeyword,
					MatchedKeyword:    match.Keyword,
					MatchType:         "fuzzy",
					BaseWeight:        match.Weight,
					ContextMultiplier: contextMultiplier,
					FinalWeight:       finalWeight,
					Confidence:        esa.calculateMatchConfidence("fuzzy", match.Weight, contextMultiplier),
					Source:            contextualKeyword.Context,
					IndustryID:        match.IndustryID,
				})
			}
		}
	}

	return matches
}

// updateIndustryScores updates the industry scores with new matches
func (esa *EnhancedScoringAlgorithm) updateIndustryScores(
	industryScores map[int]*IndustryScore,
	matches []KeywordMatch,
	contextualKeyword ContextualKeyword,
) {
	for _, match := range matches {
		// Get industry ID from the match - this should be stored in the match structure
		industryID := match.IndustryID

		// Initialize industry score if not exists
		if industryScores[industryID] == nil {
			industryScores[industryID] = &IndustryScore{
				IndustryID:         industryID,
				TotalScore:         0.0,
				DirectMatches:      []KeywordMatch{},
				PhraseMatches:      []KeywordMatch{},
				PartialMatches:     []KeywordMatch{},
				FuzzyMatches:       []KeywordMatch{},
				ContextMultipliers: make(map[string]float64),
				MatchCount:         0,
				UniqueKeywords:     make(map[string]bool),
			}
		}

		// Add match to appropriate category
		switch match.MatchType {
		case "direct":
			industryScores[industryID].DirectMatches = append(industryScores[industryID].DirectMatches, match)
		case "phrase", "phrase_partial":
			industryScores[industryID].PhraseMatches = append(industryScores[industryID].PhraseMatches, match)
		case "partial":
			industryScores[industryID].PartialMatches = append(industryScores[industryID].PartialMatches, match)
		case "fuzzy":
			industryScores[industryID].FuzzyMatches = append(industryScores[industryID].FuzzyMatches, match)
		}

		// Update totals
		industryScores[industryID].TotalScore += match.FinalWeight
		industryScores[industryID].MatchCount++
		industryScores[industryID].UniqueKeywords[match.MatchedKeyword] = true
		industryScores[industryID].ContextMultipliers[match.Source] = match.ContextMultiplier
	}
}

// findBestIndustry finds the industry with the highest score
func (esa *EnhancedScoringAlgorithm) findBestIndustry(industryScores map[int]*IndustryScore) (int, float64) {
	bestIndustryID := 26 // Default industry
	bestScore := 0.0

	// If no matches found, return default industry
	if len(industryScores) == 0 {
		return bestIndustryID, bestScore
	}

	for industryID, score := range industryScores {
		if score.TotalScore > bestScore {
			bestScore = score.TotalScore
			bestIndustryID = industryID
		}
	}

	return bestIndustryID, bestScore
}

// calculateScoreBreakdown calculates detailed score breakdown
func (esa *EnhancedScoringAlgorithm) calculateScoreBreakdown(industryScore *IndustryScore) *ScoreBreakdown {
	if industryScore == nil {
		return &ScoreBreakdown{}
	}

	// Calculate scores for each match type
	directScore := esa.calculateMatchTypeScore(industryScore.DirectMatches)
	phraseScore := esa.calculateMatchTypeScore(industryScore.PhraseMatches)
	partialScore := esa.calculateMatchTypeScore(industryScore.PartialMatches)
	fuzzyScore := esa.calculateMatchTypeScore(industryScore.FuzzyMatches)

	// Calculate context score
	contextScore := esa.calculateContextScore(industryScore.ContextMultipliers)

	// Calculate semantic and industry boost scores
	semanticBoostScore := 0.0
	industryBoostScore := 0.0

	if esa.config.EnableSemanticBoost {
		semanticBoostScore = esa.calculateSemanticBoostScore(industryScore)
	}

	if esa.config.EnableIndustryBoost {
		industryBoostScore = esa.calculateIndustryBoostScore(industryScore)
	}

	// Calculate total weighted score
	totalWeightedScore := (directScore * esa.config.DirectMatchWeight) +
		(phraseScore * esa.config.PhraseMatchWeight) +
		(partialScore * esa.config.PartialMatchWeight) +
		(contextScore * esa.config.ContextWeight) +
		semanticBoostScore +
		industryBoostScore

	return &ScoreBreakdown{
		DirectMatchScore:   directScore,
		PhraseMatchScore:   phraseScore,
		PartialMatchScore:  partialScore,
		ContextScore:       contextScore,
		FuzzyMatchScore:    fuzzyScore,
		SemanticBoostScore: semanticBoostScore,
		IndustryBoostScore: industryBoostScore,
		TotalWeightedScore: totalWeightedScore,
	}
}

// calculateMatchTypeScore calculates the total score for a specific match type
func (esa *EnhancedScoringAlgorithm) calculateMatchTypeScore(matches []KeywordMatch) float64 {
	totalScore := 0.0
	for _, match := range matches {
		totalScore += match.FinalWeight
	}
	return totalScore
}

// calculateContextScore calculates the context score based on context multipliers
func (esa *EnhancedScoringAlgorithm) calculateContextScore(contextMultipliers map[string]float64) float64 {
	if len(contextMultipliers) == 0 {
		return 0.0
	}

	totalMultiplier := 0.0
	count := 0

	for _, multiplier := range contextMultipliers {
		totalMultiplier += multiplier
		count++
	}

	return totalMultiplier / float64(count)
}

// calculateSemanticBoostScore calculates semantic boost score
func (esa *EnhancedScoringAlgorithm) calculateSemanticBoostScore(industryScore *IndustryScore) float64 {
	// Simple semantic boost based on keyword diversity and relevance
	diversityScore := float64(len(industryScore.UniqueKeywords)) / 10.0 // Normalize to 0-1
	if diversityScore > 1.0 {
		diversityScore = 1.0
	}

	return diversityScore * 0.1 // 10% boost maximum
}

// calculateIndustryBoostScore calculates industry-specific boost score
func (esa *EnhancedScoringAlgorithm) calculateIndustryBoostScore(industryScore *IndustryScore) float64 {
	// Industry boost based on match count and keyword relevance
	matchCountScore := float64(industryScore.MatchCount) / 20.0 // Normalize to 0-1
	if matchCountScore > 1.0 {
		matchCountScore = 1.0
	}

	return matchCountScore * 0.05 // 5% boost maximum
}

// calculateEnhancedConfidence calculates enhanced confidence score
func (esa *EnhancedScoringAlgorithm) calculateEnhancedConfidence(
	totalScore float64,
	scoreBreakdown *ScoreBreakdown,
	totalKeywords int,
) float64 {
	// Base confidence from total score
	baseConfidence := totalScore

	// Apply score breakdown factors
	directFactor := scoreBreakdown.DirectMatchScore * 0.4
	phraseFactor := scoreBreakdown.PhraseMatchScore * 0.3
	partialFactor := scoreBreakdown.PartialMatchScore * 0.2
	contextFactor := scoreBreakdown.ContextScore * 0.1

	// Calculate enhanced confidence
	confidence := baseConfidence + directFactor + phraseFactor + partialFactor + contextFactor

	// Apply semantic and industry boosts
	confidence += scoreBreakdown.SemanticBoostScore
	confidence += scoreBreakdown.IndustryBoostScore

	// Normalize confidence to 0-1 range
	confidence = math.Max(0.1, math.Min(1.0, confidence))

	return confidence
}

// calculateQualityIndicators calculates quality indicators for the result
func (esa *EnhancedScoringAlgorithm) calculateQualityIndicators(
	industryScore *IndustryScore,
	contextualKeywords []ContextualKeyword,
) *QualityIndicators {
	if industryScore == nil {
		return &QualityIndicators{}
	}

	// Calculate match diversity
	matchDiversity := esa.calculateMatchDiversity(industryScore)

	// Calculate keyword relevance
	keywordRelevance := esa.calculateKeywordRelevance(industryScore)

	// Calculate context consistency
	contextConsistency := esa.calculateContextConsistency(industryScore, contextualKeywords)

	// Calculate confidence stability
	confidenceStability := esa.calculateConfidenceStability(industryScore)

	// Calculate overall quality
	overallQuality := (matchDiversity + keywordRelevance + contextConsistency + confidenceStability) / 4.0

	return &QualityIndicators{
		MatchDiversity:      matchDiversity,
		KeywordRelevance:    keywordRelevance,
		ContextConsistency:  contextConsistency,
		ConfidenceStability: confidenceStability,
		OverallQuality:      overallQuality,
	}
}

// Helper methods for quality indicators
func (esa *EnhancedScoringAlgorithm) calculateMatchDiversity(industryScore *IndustryScore) float64 {
	matchTypes := 0
	if len(industryScore.DirectMatches) > 0 {
		matchTypes++
	}
	if len(industryScore.PhraseMatches) > 0 {
		matchTypes++
	}
	if len(industryScore.PartialMatches) > 0 {
		matchTypes++
	}
	if len(industryScore.FuzzyMatches) > 0 {
		matchTypes++
	}

	return float64(matchTypes) / 4.0 // Normalize to 0-1
}

func (esa *EnhancedScoringAlgorithm) calculateKeywordRelevance(industryScore *IndustryScore) float64 {
	if industryScore.MatchCount == 0 {
		return 0.0
	}

	// Calculate average keyword weight
	totalWeight := 0.0
	allMatches := append(append(append(industryScore.DirectMatches, industryScore.PhraseMatches...), industryScore.PartialMatches...), industryScore.FuzzyMatches...)

	for _, match := range allMatches {
		totalWeight += match.BaseWeight
	}

	avgWeight := totalWeight / float64(len(allMatches))
	return math.Min(1.0, avgWeight) // Normalize to 0-1
}

func (esa *EnhancedScoringAlgorithm) calculateContextConsistency(industryScore *IndustryScore, contextualKeywords []ContextualKeyword) float64 {
	if len(contextualKeywords) == 0 {
		return 0.0
	}

	// Calculate consistency across different contexts
	contextCounts := make(map[string]int)
	for _, keyword := range contextualKeywords {
		contextCounts[keyword.Context]++
	}

	// Simple consistency measure based on context distribution
	expectedPerContext := len(contextualKeywords) / len(contextCounts)
	consistency := 1.0

	for _, count := range contextCounts {
		deviation := math.Abs(float64(count-expectedPerContext)) / float64(expectedPerContext)
		consistency -= deviation * 0.1 // Reduce consistency for deviations
	}

	return math.Max(0.0, consistency)
}

func (esa *EnhancedScoringAlgorithm) calculateConfidenceStability(industryScore *IndustryScore) float64 {
	if industryScore.MatchCount == 0 {
		return 0.0
	}

	// Calculate confidence variance across matches
	allMatches := append(append(append(industryScore.DirectMatches, industryScore.PhraseMatches...), industryScore.PartialMatches...), industryScore.FuzzyMatches...)

	if len(allMatches) == 0 {
		return 0.0
	}

	// Calculate average confidence
	totalConfidence := 0.0
	for _, match := range allMatches {
		totalConfidence += match.Confidence
	}
	avgConfidence := totalConfidence / float64(len(allMatches))

	// Calculate variance
	variance := 0.0
	for _, match := range allMatches {
		diff := match.Confidence - avgConfidence
		variance += diff * diff
	}
	variance /= float64(len(allMatches))

	// Convert variance to stability (lower variance = higher stability)
	stability := 1.0 - math.Min(1.0, variance)
	return stability
}

// Helper methods
func (esa *EnhancedScoringAlgorithm) getContextMultiplier(context string) float64 {
	switch context {
	case "business_name":
		return 1.2 // 20% boost for business name keywords
	case "description":
		return 1.0 // No boost for description keywords
	case "website_url":
		return 1.0 // No boost for website URL keywords
	default:
		return 1.0
	}
}

func (esa *EnhancedScoringAlgorithm) calculateMatchConfidence(matchType string, baseWeight, contextMultiplier float64) float64 {
	// Base confidence from weight
	confidence := baseWeight * contextMultiplier

	// Apply match type multiplier
	switch matchType {
	case "direct":
		confidence *= 1.0 // No additional multiplier
	case "phrase":
		confidence *= 1.1 // 10% boost for phrase matches
	case "phrase_partial":
		confidence *= 0.9 // 10% reduction for partial phrase matches
	case "partial":
		confidence *= 0.7 // 30% reduction for partial matches
	case "fuzzy":
		confidence *= 0.5 // 50% reduction for fuzzy matches
	}

	return math.Max(0.1, math.Min(1.0, confidence))
}

func (esa *EnhancedScoringAlgorithm) hasPhraseOverlap(phrase1, phrase2 string) bool {
	words1 := strings.Fields(phrase1)
	words2 := strings.Fields(phrase2)

	// Check if any word from phrase1 exists in phrase2
	for _, word1 := range words1 {
		for _, word2 := range words2 {
			if word1 == word2 {
				return true
			}
		}
	}

	return false
}

func (esa *EnhancedScoringAlgorithm) calculateStringSimilarity(s1, s2 string) float64 {
	// Simple Levenshtein distance-based similarity
	if s1 == s2 {
		return 1.0
	}

	len1, len2 := len(s1), len(s2)
	if len1 == 0 {
		return 0.0
	}
	if len2 == 0 {
		return 0.0
	}

	// Calculate edit distance
	distance := esa.calculateEditDistance(s1, s2)
	maxLen := math.Max(float64(len1), float64(len2))

	// Convert distance to similarity (0-1)
	similarity := 1.0 - (float64(distance) / maxLen)
	return math.Max(0.0, similarity)
}

func (esa *EnhancedScoringAlgorithm) calculateEditDistance(s1, s2 string) int {
	len1, len2 := len(s1), len(s2)

	// Create matrix
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	// Initialize first row and column
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				min(matrix[i-1][j]+1, matrix[i][j-1]+1), // deletion vs insertion
				matrix[i-1][j-1]+cost,                   // vs substitution
			)
		}
	}

	return matrix[len1][len2]
}

func (esa *EnhancedScoringAlgorithm) getIndustryName(industryID int, keywordIndex *KeywordIndex) string {
	// Extract industry name from the keyword index or industry mapping
	// For now, return a default value
	return "General Business"
}

func (esa *EnhancedScoringAlgorithm) extractMatchedKeywords(industryScore *IndustryScore) []MatchedKeyword {
	var matchedKeywords []MatchedKeyword

	// Check if industryScore is nil
	if industryScore == nil {
		return matchedKeywords
	}

	// Combine all matches
	allMatches := append(append(append(industryScore.DirectMatches, industryScore.PhraseMatches...), industryScore.PartialMatches...), industryScore.FuzzyMatches...)

	for _, match := range allMatches {
		matchedKeywords = append(matchedKeywords, MatchedKeyword{
			Keyword:           match.InputKeyword,
			MatchedKeyword:    match.MatchedKeyword,
			MatchType:         match.MatchType,
			BaseWeight:        match.BaseWeight,
			ContextMultiplier: match.ContextMultiplier,
			FinalWeight:       match.FinalWeight,
			Confidence:        match.Confidence,
			Source:            match.Source,
		})
	}

	return matchedKeywords
}

func (esa *EnhancedScoringAlgorithm) generateRequestID() string {
	return fmt.Sprintf("esa_%d", time.Now().UnixNano())
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
