package classification

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"
)

// ContextualKeyword represents a keyword with its context information
type ContextualKeyword struct {
	Keyword    string  `json:"keyword"`
	Weight     float64 `json:"weight"`
	Context    string  `json:"context"`
	IndustryID int     `json:"industry_id"`
	MatchType  string  `json:"match_type"` // "direct", "phrase", "partial", "fuzzy"
	Confidence float64 `json:"confidence"`
}

// EnhancedScoringAlgorithm provides sophisticated scoring that combines direct matches, phrase matches, and context multipliers
type EnhancedScoringAlgorithm struct {
	logger               *log.Logger
	config               *EnhancedScoringConfig
	advancedFuzzyMatcher *AdvancedFuzzyMatcher
}

// EnhancedScoringConfig holds configuration for the enhanced scoring algorithm
type EnhancedScoringConfig struct {
	// Scoring weights
	DirectMatchWeight  float64 `json:"direct_match_weight"`  // Weight for exact keyword matches
	PhraseMatchWeight  float64 `json:"phrase_match_weight"`  // Weight for phrase matches
	PartialMatchWeight float64 `json:"partial_match_weight"` // Weight for partial matches
	ContextWeight      float64 `json:"context_weight"`       // Weight for context multipliers

	// Context-aware scoring weights
	BusinessNameWeight    float64 `json:"business_name_weight"`    // Weight for business name keywords
	DescriptionWeight     float64 `json:"description_weight"`      // Weight for description keywords
	WebsiteURLWeight      float64 `json:"website_url_weight"`      // Weight for website URL keywords
	IndustrySpecificBoost float64 `json:"industry_specific_boost"` // Boost for industry-specific keywords

	// Performance optimization
	MaxKeywordsToProcess int  `json:"max_keywords_to_process"` // Limit keywords for performance
	CacheResults         bool `json:"cache_results"`           // Enable result caching
	ParallelProcessing   bool `json:"parallel_processing"`     // Enable parallel processing

	// Quality thresholds
	MinMatchThreshold   float64 `json:"min_match_threshold"`  // Minimum score to consider
	ConfidenceThreshold float64 `json:"confidence_threshold"` // Minimum confidence threshold

	// Advanced features
	EnableFuzzyMatching       bool `json:"enable_fuzzy_matching"`        // Enable fuzzy string matching
	EnableSemanticBoost       bool `json:"enable_semantic_boost"`        // Enable semantic similarity boost
	EnableIndustryBoost       bool `json:"enable_industry_boost"`        // Enable industry-specific boosts
	EnableContextAwareScoring bool `json:"enable_context_aware_scoring"` // Enable context-aware scoring
	EnableDynamicWeightAdjust bool `json:"enable_dynamic_weight_adjust"` // Enable dynamic weight adjustment
}

// DefaultEnhancedScoringConfig returns the default configuration for enhanced scoring
func DefaultEnhancedScoringConfig() *EnhancedScoringConfig {
	return &EnhancedScoringConfig{
		// Scoring weights (must sum to 1.0)
		DirectMatchWeight:  0.40, // 40% weight for exact matches
		PhraseMatchWeight:  0.30, // 30% weight for phrase matches
		PartialMatchWeight: 0.20, // 20% weight for partial matches
		ContextWeight:      0.10, // 10% weight for context multipliers

		// Context-aware scoring weights
		BusinessNameWeight:    1.5, // 50% boost for business name keywords (highest priority)
		DescriptionWeight:     1.0, // No boost for description keywords (baseline)
		WebsiteURLWeight:      0.8, // 20% reduction for website URL keywords (lowest priority)
		IndustrySpecificBoost: 1.3, // 30% boost for industry-specific keywords

		// Performance optimization
		MaxKeywordsToProcess: 1000, // Process up to 1000 keywords
		CacheResults:         true, // Enable caching for performance
		ParallelProcessing:   true, // Enable parallel processing

		// Quality thresholds
		MinMatchThreshold:   0.1, // Minimum score to consider
		ConfidenceThreshold: 0.5, // Minimum confidence threshold

		// Advanced features
		EnableFuzzyMatching:       true, // Enable fuzzy matching
		EnableSemanticBoost:       true, // Enable semantic boost
		EnableIndustryBoost:       true, // Enable industry boost
		EnableContextAwareScoring: true, // Enable context-aware scoring
		EnableDynamicWeightAdjust: true, // Enable dynamic weight adjustment
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

// ContextAwareScore represents a context-aware scoring result
type ContextAwareScore struct {
	Source            string  `json:"source"`             // "business_name", "description", "website_url"
	BaseWeight        float64 `json:"base_weight"`        // Base weight from keyword match
	ContextMultiplier float64 `json:"context_multiplier"` // Context-specific multiplier
	IndustryBoost     float64 `json:"industry_boost"`     // Industry-specific boost
	FinalWeight       float64 `json:"final_weight"`       // Final calculated weight
	Confidence        float64 `json:"confidence"`         // Confidence in this score
}

// IndustrySpecificKeyword represents a keyword with industry-specific importance
type IndustrySpecificKeyword struct {
	Keyword     string  `json:"keyword"`
	IndustryID  int     `json:"industry_id"`
	Importance  float64 `json:"importance"`  // 0.0-1.0 importance score
	Specificity float64 `json:"specificity"` // 0.0-1.0 specificity score
	Frequency   int     `json:"frequency"`   // Frequency in industry
}

// DynamicWeightAdjustment represents dynamic weight adjustment factors
type DynamicWeightAdjustment struct {
	KeywordDensity     float64 `json:"keyword_density"`     // Density of keywords in context
	IndustryRelevance  float64 `json:"industry_relevance"`  // Relevance to target industry
	ContextConsistency float64 `json:"context_consistency"` // Consistency across contexts
	MatchQuality       float64 `json:"match_quality"`       // Quality of matches
	AdjustmentFactor   float64 `json:"adjustment_factor"`   // Final adjustment factor
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

	// Create advanced fuzzy matcher with configuration
	fuzzyConfig := DefaultAdvancedFuzzyConfig()
	fuzzyConfig.SimilarityThreshold = config.MinMatchThreshold
	fuzzyConfig.EnableSemanticExpand = config.EnableSemanticBoost
	advancedFuzzyMatcher := NewAdvancedFuzzyMatcher(logger, fuzzyConfig)

	return &EnhancedScoringAlgorithm{
		logger:               logger,
		config:               config,
		advancedFuzzyMatcher: advancedFuzzyMatcher,
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

	// Check context deadline at start
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		if timeRemaining < 1*time.Second {
			return nil, fmt.Errorf("context deadline too short for enhanced scoring: %v remaining", timeRemaining)
		}
		esa.logger.Printf("⏱️ [PROFILING] CalculateEnhancedScore start - time remaining: %v", timeRemaining)
	}

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

	// Determine batch size and early termination threshold based on context deadline
	batchSize := 10 // Default batch size
	earlyTerminationThreshold := 0.9 // Early termination if confidence > 0.9
	maxKeywordsToProcess := len(contextualKeywords)
	
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		// Adjust batch size and max keywords based on remaining time
		if timeRemaining < 5*time.Second {
			batchSize = 5 // Smaller batches for short deadlines
			if maxKeywordsToProcess > 20 {
				maxKeywordsToProcess = 20 // Limit to 20 keywords
			}
			earlyTerminationThreshold = 0.85 // Lower threshold for early termination
		} else if timeRemaining < 10*time.Second {
			if maxKeywordsToProcess > 50 {
				maxKeywordsToProcess = 50 // Limit to 50 keywords
			}
		}
		esa.logger.Printf("📊 [BATCHING] Processing %d keywords in batches of %d (max: %d, early termination: %.2f)", 
			len(contextualKeywords), batchSize, maxKeywordsToProcess, earlyTerminationThreshold)
	}

	// Process keywords in batches with early termination
	keywordsProcessed := 0
	keywordProcessingStart := time.Now()
	
	for i := 0; i < len(contextualKeywords) && i < maxKeywordsToProcess; i += batchSize {
		// Check context deadline before each batch
		if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
			timeRemaining := time.Until(deadline)
			if timeRemaining < 2*time.Second {
				esa.logger.Printf("⚠️ Context deadline approaching, stopping keyword processing after %d keywords", keywordsProcessed)
				break
			}
		}
		
		// Process batch
		batchEnd := i + batchSize
		if batchEnd > len(contextualKeywords) {
			batchEnd = len(contextualKeywords)
		}
		if batchEnd > maxKeywordsToProcess {
			batchEnd = maxKeywordsToProcess
		}
		for j := i; j < batchEnd; j++ {
			contextualKeyword := contextualKeywords[j]
			matches := esa.findEnhancedMatches(ctx, contextualKeyword, keywordIndex)
			esa.updateIndustryScores(industryScores, matches, contextualKeyword)
			performanceMetrics.MatchesFound += len(matches)
			keywordsProcessed++
		}
		
		// Check for early termination after each batch
		if len(industryScores) > 0 {
			// Calculate current best score to check for early termination
			tempBestID, tempBestScore := esa.findBestIndustry(industryScores)
			if tempBestID != 0 {
				// Calculate temporary confidence for early termination check
				tempBreakdown := esa.calculateScoreBreakdown(industryScores[tempBestID])
				tempConfidence := esa.calculateEnhancedConfidence(tempBestScore, tempBreakdown, keywordsProcessed)
				
				if tempConfidence >= earlyTerminationThreshold {
					esa.logger.Printf("✅ [EARLY TERMINATION] Achieved confidence %.3f (threshold: %.2f) after processing %d keywords, stopping early", 
						tempConfidence, earlyTerminationThreshold, keywordsProcessed)
					break
				}
			}
		}
	}
	
	keywordProcessingDuration := time.Since(keywordProcessingStart)
	performanceMetrics.KeywordsProcessed = keywordsProcessed
	esa.logger.Printf("⏱️ [PROFILING] Keyword processing completed: %d keywords processed in %v", keywordsProcessed, keywordProcessingDuration)

	// Apply dynamic weight adjustment if enabled (check deadline first)
	if esa.config.EnableDynamicWeightAdjust {
		if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
			if time.Until(deadline) < 2*time.Second {
				esa.logger.Printf("⚠️ Context deadline too short for dynamic weight adjustment, skipping")
			} else {
				esa.applyDynamicWeightAdjustment(industryScores, contextualKeywords)
			}
		} else {
			esa.applyDynamicWeightAdjustment(industryScores, contextualKeywords)
		}
	}

	// Check context deadline before finding best industry
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		if time.Until(deadline) < 1*time.Second {
			esa.logger.Printf("⚠️ Context deadline approaching, using partial results")
			// Return early with partial results if deadline is too short
			if len(industryScores) == 0 {
				return nil, fmt.Errorf("no industry scores calculated before deadline")
			}
		}
	}

	// Find best industry
	bestIndustryID, bestScore := esa.findBestIndustry(industryScores)

	// Calculate detailed score breakdown
	scoreBreakdown := esa.calculateScoreBreakdown(industryScores[bestIndustryID])

	// Calculate confidence
	confidence := esa.calculateEnhancedConfidence(bestScore, scoreBreakdown, len(contextualKeywords))

	// Get industry name
	industryName := esa.getIndustryName(bestIndustryID, keywordIndex)

	// Calculate quality indicators
	qualityIndicators := esa.calculateQualityIndicators(industryScores[bestIndustryID], contextualKeywords)

	// Update performance metrics
	performanceMetrics.ProcessingTime = time.Since(startTime)

	// Create result
	result := &EnhancedScoringResult{
		IndustryID:         bestIndustryID,
		IndustryName:       industryName,
		TotalScore:         bestScore,
		Confidence:         confidence,
		ScoreBreakdown:     scoreBreakdown,
		MatchedKeywords:    esa.extractMatchedKeywords(industryScores[bestIndustryID]),
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
	ctx context.Context,
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

	// 3. Partial matches (substring matches) - optimized with context checks
	partialMatches := esa.findPartialMatches(ctx, normalizedKeyword, contextualKeyword, keywordIndex)
	matches = append(matches, partialMatches...)

	// 4. Fuzzy matches (if enabled) - optimized with context checks
	if esa.config.EnableFuzzyMatching {
		// Check context deadline before expensive fuzzy matching
		if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
			if time.Until(deadline) < 5*time.Second {
				esa.logger.Printf("⚠️ Context deadline too short for fuzzy matching, skipping")
			} else {
				fuzzyMatches := esa.findFuzzyMatches(ctx, normalizedKeyword, contextualKeyword, keywordIndex)
				matches = append(matches, fuzzyMatches...)
			}
		} else {
			fuzzyMatches := esa.findFuzzyMatches(ctx, normalizedKeyword, contextualKeyword, keywordIndex)
			matches = append(matches, fuzzyMatches...)
		}
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

// findPartialMatches finds partial (substring) matches with optimizations
func (esa *EnhancedScoringAlgorithm) findPartialMatches(
	ctx context.Context,
	normalizedKeyword string,
	contextualKeyword ContextualKeyword,
	keywordIndex *KeywordIndex,
) []KeywordMatch {
	const maxPartialMatches = 50 // Maximum number of partial matches to return
	const lengthSimilarityThreshold = 0.5 // Only consider keywords within 50% length difference
	
	var matches []KeywordMatch
	normalizedKeywordLen := len(normalizedKeyword)
	
	// Pre-filter candidates by length similarity to reduce search space
	type candidateMatch struct {
		keyword        string
		industryMatches []IndexKeywordMatch
		lengthDiff     float64
	}
	
	var candidates []candidateMatch
	for keyword, industryMatches := range keywordIndex.KeywordToIndustries {
		// Skip exact matches (already handled by direct matches)
		if keyword == normalizedKeyword {
			continue
		}
		
		// Length-based filtering: only consider keywords with similar length
		keywordLen := len(keyword)
		lengthDiff := float64(absInt(normalizedKeywordLen - keywordLen)) / float64(maxInt(normalizedKeywordLen, keywordLen))
		if lengthDiff <= lengthSimilarityThreshold {
			candidates = append(candidates, candidateMatch{
				keyword:          keyword,
				industryMatches:  industryMatches,
				lengthDiff:       lengthDiff,
			})
		}
	}
	
	// Sort candidates by length similarity (closer length = better match)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].lengthDiff < candidates[j].lengthDiff
	})
	
	// Process candidates with early termination
	for i, candidate := range candidates {
		// Check context deadline periodically (every 10 candidates)
		if i%10 == 0 {
			if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
				if time.Until(deadline) < 1*time.Second {
					esa.logger.Printf("⚠️ Context deadline approaching, stopping partial match search")
					break
				}
			}
		}
		
		// Early termination if we have enough matches
		if len(matches) >= maxPartialMatches {
			break
		}
		
		// Check for substring matches (optimized: check shorter string first)
		hasMatch := false
		if len(candidate.keyword) <= normalizedKeywordLen {
			hasMatch = strings.Contains(normalizedKeyword, candidate.keyword)
		} else {
			hasMatch = strings.Contains(candidate.keyword, normalizedKeyword)
		}
		
		if hasMatch {
			for _, match := range candidate.industryMatches {
				// Early termination check
				if len(matches) >= maxPartialMatches {
					break
				}
				
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

// Helper functions for optimization
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Helper functions absInt is defined here, maxInt and minInt are in advanced_fuzzy_matcher.go

// findFuzzyMatches finds fuzzy string matches using advanced fuzzy matching algorithms with optimizations
func (esa *EnhancedScoringAlgorithm) findFuzzyMatches(
	ctx context.Context,
	normalizedKeyword string,
	contextualKeyword ContextualKeyword,
	keywordIndex *KeywordIndex,
) []KeywordMatch {
	const maxFuzzyCandidates = 100 // Limit fuzzy matching to top 100 candidates
	const lengthSimilarityThreshold = 0.6 // Only consider keywords within 60% length difference
	
	var matches []KeywordMatch
	normalizedKeywordLen := len(normalizedKeyword)
	
	// Pre-filter candidates by length similarity before expensive fuzzy matching
	type candidateInfo struct {
		keyword string
		lengthDiff float64
	}
	
	var candidates []candidateInfo
	for keyword := range keywordIndex.KeywordToIndustries {
		// Skip exact matches
		if keyword == normalizedKeyword {
			continue
		}
		
		// Length-based filtering: only consider keywords with similar length
		keywordLen := len(keyword)
		lengthDiff := float64(absInt(normalizedKeywordLen - keywordLen)) / float64(maxInt(normalizedKeywordLen, keywordLen))
		if lengthDiff <= lengthSimilarityThreshold {
			candidates = append(candidates, candidateInfo{
				keyword:    keyword,
				lengthDiff: lengthDiff,
			})
		}
	}
	
	// Sort candidates by length similarity (closer length = better match)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].lengthDiff < candidates[j].lengthDiff
	})
	
	// Limit to top N candidates for performance
	if len(candidates) > maxFuzzyCandidates {
		candidates = candidates[:maxFuzzyCandidates]
	}
	
	// Check context deadline before expensive fuzzy matching
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		if time.Until(deadline) < 3*time.Second {
			esa.logger.Printf("⚠️ Context deadline too short for fuzzy matching, using simple fallback")
			// Convert candidates to string slice for simple matching
			candidateKeywords := make([]string, len(candidates))
			for i, c := range candidates {
				candidateKeywords[i] = c.keyword
			}
			return esa.findSimpleFuzzyMatches(ctx, normalizedKeyword, contextualKeyword, keywordIndex, candidateKeywords)
		}
	}

	// Use advanced fuzzy matcher for sophisticated matching (with filtered candidates)
	// Note: The advanced fuzzy matcher will process the filtered candidates
	fuzzyMatches, err := esa.advancedFuzzyMatcher.FindFuzzyMatches(ctx, normalizedKeyword, keywordIndex)
	if err != nil {
		esa.logger.Printf("⚠️ Advanced fuzzy matching failed, falling back to simple matching: %v", err)
		// Convert candidates to string slice for simple matching
		candidateKeywords := make([]string, len(candidates))
		for i, c := range candidates {
			candidateKeywords[i] = c.keyword
		}
		return esa.findSimpleFuzzyMatches(ctx, normalizedKeyword, contextualKeyword, keywordIndex, candidateKeywords)
	}

	// Convert fuzzy matches to keyword matches
	for _, fuzzyMatch := range fuzzyMatches {
		contextMultiplier := esa.getContextMultiplier(contextualKeyword.Context)
		fuzzyMultiplier := fuzzyMatch.Similarity * 0.4 // Higher weight for advanced fuzzy matches
		finalWeight := fuzzyMatch.Weight * fuzzyMultiplier * contextMultiplier

		matches = append(matches, KeywordMatch{
			InputKeyword:      normalizedKeyword,
			MatchedKeyword:    fuzzyMatch.Keyword,
			MatchType:         "advanced_fuzzy",
			BaseWeight:        fuzzyMatch.Weight,
			ContextMultiplier: contextMultiplier,
			FinalWeight:       finalWeight,
			Confidence:        fuzzyMatch.Confidence,
			Source:            contextualKeyword.Context,
			IndustryID:        fuzzyMatch.IndustryID,
		})
	}

	// Perform semantic expansion if enabled (only if we have time)
	if esa.config.EnableSemanticBoost {
		if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
			if time.Until(deadline) < 2*time.Second {
				esa.logger.Printf("⚠️ Context deadline too short for semantic expansion, skipping")
			} else {
				semanticExpansion, err := esa.advancedFuzzyMatcher.ExpandSemanticKeywords(ctx, normalizedKeyword, keywordIndex)
				if err == nil && semanticExpansion != nil {
					for _, semanticKeyword := range semanticExpansion.Expansions {
						contextMultiplier := esa.getContextMultiplier(contextualKeyword.Context)
						semanticMultiplier := semanticKeyword.Similarity * 0.2      // Lower weight for semantic matches
						finalWeight := 1.0 * semanticMultiplier * contextMultiplier // Default weight for semantic keywords

						matches = append(matches, KeywordMatch{
							InputKeyword:      normalizedKeyword,
							MatchedKeyword:    semanticKeyword.Keyword,
							MatchType:         "semantic",
							BaseWeight:        1.0,
							ContextMultiplier: contextMultiplier,
							FinalWeight:       finalWeight,
							Confidence:        semanticKeyword.Similarity,
							Source:            contextualKeyword.Context,
							IndustryID:        semanticKeyword.IndustryID,
						})
					}
				}
			}
		} else {
			semanticExpansion, err := esa.advancedFuzzyMatcher.ExpandSemanticKeywords(ctx, normalizedKeyword, keywordIndex)
			if err == nil && semanticExpansion != nil {
				for _, semanticKeyword := range semanticExpansion.Expansions {
					contextMultiplier := esa.getContextMultiplier(contextualKeyword.Context)
					semanticMultiplier := semanticKeyword.Similarity * 0.2      // Lower weight for semantic matches
					finalWeight := 1.0 * semanticMultiplier * contextMultiplier // Default weight for semantic keywords

					matches = append(matches, KeywordMatch{
						InputKeyword:      normalizedKeyword,
						MatchedKeyword:    semanticKeyword.Keyword,
						MatchType:         "semantic",
						BaseWeight:        1.0,
						ContextMultiplier: contextMultiplier,
						FinalWeight:       finalWeight,
						Confidence:        semanticKeyword.Similarity,
						Source:            contextualKeyword.Context,
						IndustryID:        semanticKeyword.IndustryID,
					})
				}
			}
		}
	}

	return matches
}

// findSimpleFuzzyMatches provides fallback simple fuzzy matching with optimizations
func (esa *EnhancedScoringAlgorithm) findSimpleFuzzyMatches(
	ctx context.Context,
	normalizedKeyword string,
	contextualKeyword ContextualKeyword,
	keywordIndex *KeywordIndex,
	prefilteredCandidates []string,
) []KeywordMatch {
	const maxSimpleFuzzyMatches = 30 // Maximum number of simple fuzzy matches
	const similarityThreshold = 0.7 // Only consider matches with similarity > 0.7
	
	var matches []KeywordMatch
	normalizedKeywordLen := len(normalizedKeyword)
	
	// Use prefiltered candidates if provided, otherwise filter by length
	var candidates []string
	if len(prefilteredCandidates) > 0 {
		candidates = prefilteredCandidates
	} else {
		// Fallback: filter by length similarity
		type candidateInfo struct {
			keyword    string
			lengthDiff float64
		}
		var candidateInfos []candidateInfo
		for keyword := range keywordIndex.KeywordToIndustries {
			if keyword == normalizedKeyword {
				continue
			}
			keywordLen := len(keyword)
			lengthDiff := float64(absInt(normalizedKeywordLen - keywordLen)) / float64(maxInt(normalizedKeywordLen, keywordLen))
			if lengthDiff <= 0.6 {
				candidateInfos = append(candidateInfos, candidateInfo{
					keyword:    keyword,
					lengthDiff: lengthDiff,
				})
			}
		}
		// Sort by length similarity
		sort.Slice(candidateInfos, func(i, j int) bool {
			return candidateInfos[i].lengthDiff < candidateInfos[j].lengthDiff
		})
		// Limit to top candidates and extract keywords
		if len(candidateInfos) > 100 {
			candidateInfos = candidateInfos[:100]
		}
		candidates = make([]string, len(candidateInfos))
		for i, c := range candidateInfos {
			candidates[i] = c.keyword
		}
	}

	// Process candidates with early termination
	for i, candidateKeyword := range candidates {
		// Check context deadline periodically
		if i%10 == 0 {
			if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
				if time.Until(deadline) < 1*time.Second {
					esa.logger.Printf("⚠️ Context deadline approaching, stopping simple fuzzy match search")
					break
				}
			}
		}
		
		// Early termination if we have enough matches
		if len(matches) >= maxSimpleFuzzyMatches {
			break
		}
		
		// Get industry matches for this keyword
		industryMatches, exists := keywordIndex.KeywordToIndustries[candidateKeyword]
		if !exists {
			continue
		}

		// Calculate similarity score
		similarity := esa.calculateStringSimilarity(normalizedKeyword, candidateKeyword)

		// Only consider matches with similarity > threshold
		if similarity > similarityThreshold {
			for _, match := range industryMatches {
				// Early termination check
				if len(matches) >= maxSimpleFuzzyMatches {
					break
				}
				
				contextMultiplier := esa.getContextMultiplier(contextualKeyword.Context)
				fuzzyMultiplier := similarity * 0.3 // Reduced weight for fuzzy matches
				finalWeight := match.Weight * fuzzyMultiplier * contextMultiplier

				matches = append(matches, KeywordMatch{
					InputKeyword:      normalizedKeyword,
					MatchedKeyword:    match.Keyword,
					MatchType:         "simple_fuzzy",
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

// updateIndustryScores updates the industry scores with new matches using context-aware scoring
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

		// Calculate context-aware score if enabled
		var finalWeight float64
		var finalConfidence float64

		if esa.config.EnableContextAwareScoring {
			contextAwareScore := esa.calculateContextAwareScore(contextualKeyword, match, industryID)
			finalWeight = contextAwareScore.FinalWeight
			finalConfidence = contextAwareScore.Confidence

			// Update the match with context-aware values
			match.FinalWeight = finalWeight
			match.Confidence = finalConfidence
			match.ContextMultiplier = contextAwareScore.ContextMultiplier
		} else {
			// Use original scoring
			finalWeight = match.FinalWeight
			finalConfidence = match.Confidence
		}

		// Add match to appropriate category
		switch match.MatchType {
		case "direct":
			industryScores[industryID].DirectMatches = append(industryScores[industryID].DirectMatches, match)
		case "phrase", "phrase_partial":
			industryScores[industryID].PhraseMatches = append(industryScores[industryID].PhraseMatches, match)
		case "partial":
			industryScores[industryID].PartialMatches = append(industryScores[industryID].PartialMatches, match)
		case "fuzzy", "simple_fuzzy", "advanced_fuzzy", "semantic":
			industryScores[industryID].FuzzyMatches = append(industryScores[industryID].FuzzyMatches, match)
		}

		// Update totals with context-aware scoring
		industryScores[industryID].TotalScore += finalWeight
		industryScores[industryID].MatchCount++
		industryScores[industryID].UniqueKeywords[match.MatchedKeyword] = true
		industryScores[industryID].ContextMultipliers[match.Source] = match.ContextMultiplier
	}
}

// findBestIndustry finds the industry with the highest score
func (esa *EnhancedScoringAlgorithm) findBestIndustry(industryScores map[int]*IndustryScore) (int, float64) {
	bestIndustryID := 26 // Default industry
	bestScore := 0.0

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
	contextConsistency := esa.calculateContextConsistency(contextualKeywords)

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

			matrix[i][j] = minInt(
				minInt(matrix[i-1][j]+1, matrix[i][j-1]+1), // deletion vs insertion
				matrix[i-1][j-1]+cost,                      // substitution
			)
		}
	}

	return matrix[len1][len2]
}

func (esa *EnhancedScoringAlgorithm) getIndustryName(industryID int, keywordIndex *KeywordIndex) string {
	// Map industry IDs to names
	industryNames := map[int]string{
		1:  "Restaurant",
		2:  "Technology",
		3:  "Healthcare",
		4:  "Legal Services",
		5:  "Retail",
		26: "General Business",
	}

	if name, exists := industryNames[industryID]; exists {
		return name
	}

	return "General Business"
}

func (esa *EnhancedScoringAlgorithm) extractMatchedKeywords(industryScore *IndustryScore) []MatchedKeyword {
	var matchedKeywords []MatchedKeyword

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

// Context-Aware Scoring Methods

// calculateContextAwareScore calculates context-aware score for a keyword match
func (esa *EnhancedScoringAlgorithm) calculateContextAwareScore(
	contextualKeyword ContextualKeyword,
	match KeywordMatch,
	industryID int,
) *ContextAwareScore {
	// Get base weight from the match
	baseWeight := match.BaseWeight

	// Calculate context multiplier based on source
	contextMultiplier := esa.getContextMultiplier(contextualKeyword.Context)

	// Calculate industry-specific boost
	industryBoost := esa.calculateIndustrySpecificBoost(contextualKeyword.Keyword, industryID)

	// Calculate final weight with context-aware adjustments
	finalWeight := baseWeight * contextMultiplier * industryBoost

	// Calculate confidence based on context and industry relevance
	confidence := esa.calculateContextAwareConfidence(contextualKeyword, match, industryID)

	return &ContextAwareScore{
		Source:            contextualKeyword.Context,
		BaseWeight:        baseWeight,
		ContextMultiplier: contextMultiplier,
		IndustryBoost:     industryBoost,
		FinalWeight:       finalWeight,
		Confidence:        confidence,
	}
}

// getContextMultiplier returns the appropriate multiplier based on keyword context
func (esa *EnhancedScoringAlgorithm) getContextMultiplier(context string) float64 {
	if !esa.config.EnableContextAwareScoring {
		return 1.0 // No context adjustment if disabled
	}

	switch context {
	case "business_name":
		return esa.config.BusinessNameWeight
	case "description":
		return esa.config.DescriptionWeight
	case "website_url":
		return esa.config.WebsiteURLWeight
	default:
		return 1.0 // Default to no boost for unknown contexts
	}
}

// calculateIndustrySpecificBoost calculates industry-specific boost for a keyword
func (esa *EnhancedScoringAlgorithm) calculateIndustrySpecificBoost(keyword string, industryID int) float64 {
	if !esa.config.EnableIndustryBoost {
		return 1.0 // No industry boost if disabled
	}

	// Define industry-specific keyword importance
	industryKeywords := esa.getIndustrySpecificKeywords(industryID)

	// Check if keyword is industry-specific
	for _, industryKeyword := range industryKeywords {
		if strings.EqualFold(industryKeyword.Keyword, keyword) {
			// Apply boost based on importance and specificity
			boost := 1.0 + (industryKeyword.Importance * industryKeyword.Specificity * (esa.config.IndustrySpecificBoost - 1.0))
			return math.Min(boost, 2.0) // Cap at 2x boost
		}
	}

	return 1.0 // No boost for non-industry-specific keywords
}

// getIndustrySpecificKeywords returns industry-specific keywords with importance scores
func (esa *EnhancedScoringAlgorithm) getIndustrySpecificKeywords(industryID int) []IndustrySpecificKeyword {
	// Define industry-specific keywords with importance and specificity scores
	industryKeywords := map[int][]IndustrySpecificKeyword{
		// Restaurant industry (ID: 1)
		1: {
			{Keyword: "restaurant", IndustryID: 1, Importance: 0.9, Specificity: 0.8, Frequency: 100},
			{Keyword: "dining", IndustryID: 1, Importance: 0.8, Specificity: 0.7, Frequency: 80},
			{Keyword: "food", IndustryID: 1, Importance: 0.7, Specificity: 0.5, Frequency: 90},
			{Keyword: "cuisine", IndustryID: 1, Importance: 0.8, Specificity: 0.6, Frequency: 70},
			{Keyword: "menu", IndustryID: 1, Importance: 0.9, Specificity: 0.9, Frequency: 95},
			{Keyword: "chef", IndustryID: 1, Importance: 0.8, Specificity: 0.8, Frequency: 60},
			{Keyword: "kitchen", IndustryID: 1, Importance: 0.7, Specificity: 0.6, Frequency: 75},
			{Keyword: "catering", IndustryID: 1, Importance: 0.8, Specificity: 0.7, Frequency: 50},
		},
		// Technology industry (ID: 2)
		2: {
			{Keyword: "software", IndustryID: 2, Importance: 0.9, Specificity: 0.8, Frequency: 100},
			{Keyword: "technology", IndustryID: 2, Importance: 0.8, Specificity: 0.6, Frequency: 90},
			{Keyword: "development", IndustryID: 2, Importance: 0.8, Specificity: 0.7, Frequency: 85},
			{Keyword: "programming", IndustryID: 2, Importance: 0.9, Specificity: 0.9, Frequency: 80},
			{Keyword: "application", IndustryID: 2, Importance: 0.7, Specificity: 0.6, Frequency: 75},
			{Keyword: "platform", IndustryID: 2, Importance: 0.8, Specificity: 0.7, Frequency: 70},
			{Keyword: "api", IndustryID: 2, Importance: 0.9, Specificity: 0.9, Frequency: 65},
			{Keyword: "database", IndustryID: 2, Importance: 0.8, Specificity: 0.8, Frequency: 60},
		},
		// Healthcare industry (ID: 3)
		3: {
			{Keyword: "medical", IndustryID: 3, Importance: 0.9, Specificity: 0.8, Frequency: 100},
			{Keyword: "healthcare", IndustryID: 3, Importance: 0.9, Specificity: 0.7, Frequency: 95},
			{Keyword: "clinic", IndustryID: 3, Importance: 0.8, Specificity: 0.8, Frequency: 80},
			{Keyword: "hospital", IndustryID: 3, Importance: 0.8, Specificity: 0.9, Frequency: 75},
			{Keyword: "doctor", IndustryID: 3, Importance: 0.8, Specificity: 0.8, Frequency: 70},
			{Keyword: "patient", IndustryID: 3, Importance: 0.7, Specificity: 0.7, Frequency: 85},
			{Keyword: "treatment", IndustryID: 3, Importance: 0.8, Specificity: 0.7, Frequency: 80},
			{Keyword: "therapy", IndustryID: 3, Importance: 0.8, Specificity: 0.8, Frequency: 65},
		},
		// Legal services industry (ID: 4)
		4: {
			{Keyword: "legal", IndustryID: 4, Importance: 0.9, Specificity: 0.8, Frequency: 100},
			{Keyword: "law", IndustryID: 4, Importance: 0.9, Specificity: 0.7, Frequency: 95},
			{Keyword: "attorney", IndustryID: 4, Importance: 0.8, Specificity: 0.8, Frequency: 80},
			{Keyword: "lawyer", IndustryID: 4, Importance: 0.8, Specificity: 0.8, Frequency: 75},
			{Keyword: "litigation", IndustryID: 4, Importance: 0.8, Specificity: 0.9, Frequency: 60},
			{Keyword: "court", IndustryID: 4, Importance: 0.7, Specificity: 0.7, Frequency: 70},
			{Keyword: "legal advice", IndustryID: 4, Importance: 0.9, Specificity: 0.9, Frequency: 65},
			{Keyword: "compliance", IndustryID: 4, Importance: 0.7, Specificity: 0.6, Frequency: 55},
		},
		// Retail industry (ID: 5)
		5: {
			{Keyword: "retail", IndustryID: 5, Importance: 0.9, Specificity: 0.8, Frequency: 100},
			{Keyword: "store", IndustryID: 5, Importance: 0.8, Specificity: 0.6, Frequency: 90},
			{Keyword: "shop", IndustryID: 5, Importance: 0.8, Specificity: 0.6, Frequency: 85},
			{Keyword: "sales", IndustryID: 5, Importance: 0.7, Specificity: 0.5, Frequency: 80},
			{Keyword: "products", IndustryID: 5, Importance: 0.7, Specificity: 0.5, Frequency: 75},
			{Keyword: "merchandise", IndustryID: 5, Importance: 0.8, Specificity: 0.7, Frequency: 60},
			{Keyword: "inventory", IndustryID: 5, Importance: 0.7, Specificity: 0.7, Frequency: 55},
			{Keyword: "customer", IndustryID: 5, Importance: 0.6, Specificity: 0.4, Frequency: 70},
		},
	}

	if keywords, exists := industryKeywords[industryID]; exists {
		return keywords
	}

	return []IndustrySpecificKeyword{} // Return empty slice for unknown industries
}

// calculateContextAwareConfidence calculates confidence based on context and industry relevance
func (esa *EnhancedScoringAlgorithm) calculateContextAwareConfidence(
	contextualKeyword ContextualKeyword,
	match KeywordMatch,
	industryID int,
) float64 {
	baseConfidence := match.Confidence

	// Boost confidence for business name keywords
	contextBoost := 1.0
	switch contextualKeyword.Context {
	case "business_name":
		contextBoost = 1.2 // 20% confidence boost
	case "description":
		contextBoost = 1.0 // No boost
	case "website_url":
		contextBoost = 0.9 // 10% confidence reduction
	}

	// Boost confidence for industry-specific keywords
	industryBoost := 1.0
	industryKeywords := esa.getIndustrySpecificKeywords(industryID)
	for _, industryKeyword := range industryKeywords {
		if strings.EqualFold(industryKeyword.Keyword, contextualKeyword.Keyword) {
			industryBoost = 1.0 + (industryKeyword.Importance * 0.2) // Up to 20% boost
			break
		}
	}

	// Calculate final confidence
	finalConfidence := baseConfidence * contextBoost * industryBoost

	// Cap confidence at 1.0
	return math.Min(finalConfidence, 1.0)
}

// calculateDynamicWeightAdjustment calculates dynamic weight adjustment based on context
func (esa *EnhancedScoringAlgorithm) calculateDynamicWeightAdjustment(
	contextualKeywords []ContextualKeyword,
	industryID int,
) *DynamicWeightAdjustment {
	if !esa.config.EnableDynamicWeightAdjust {
		return &DynamicWeightAdjustment{
			AdjustmentFactor: 1.0,
		}
	}

	// Calculate keyword density in each context
	keywordDensity := esa.calculateKeywordDensity(contextualKeywords)

	// Calculate industry relevance
	industryRelevance := esa.calculateIndustryRelevance(contextualKeywords, industryID)

	// Calculate context consistency
	contextConsistency := esa.calculateContextConsistency(contextualKeywords)

	// Calculate match quality
	matchQuality := esa.calculateMatchQuality(contextualKeywords)

	// Calculate final adjustment factor
	adjustmentFactor := (keywordDensity * 0.2) + (industryRelevance * 0.3) +
		(contextConsistency * 0.2) + (matchQuality * 0.3)

	// Normalize to reasonable range (0.8 to 1.3) - less aggressive adjustment
	adjustmentFactor = math.Max(0.8, math.Min(1.3, adjustmentFactor))

	return &DynamicWeightAdjustment{
		KeywordDensity:     keywordDensity,
		IndustryRelevance:  industryRelevance,
		ContextConsistency: contextConsistency,
		MatchQuality:       matchQuality,
		AdjustmentFactor:   adjustmentFactor,
	}
}

// calculateKeywordDensity calculates the density of keywords in different contexts
func (esa *EnhancedScoringAlgorithm) calculateKeywordDensity(contextualKeywords []ContextualKeyword) float64 {
	if len(contextualKeywords) == 0 {
		return 0.0
	}

	contextCounts := make(map[string]int)
	for _, keyword := range contextualKeywords {
		contextCounts[keyword.Context]++
	}

	// Calculate diversity of contexts
	contextDiversity := float64(len(contextCounts)) / 3.0 // Max 3 contexts

	// Calculate average keywords per context
	totalKeywords := len(contextualKeywords)
	avgKeywordsPerContext := float64(totalKeywords) / float64(len(contextCounts))

	// Normalize density score (0.0 to 1.0)
	density := (contextDiversity * 0.5) + (math.Min(avgKeywordsPerContext/10.0, 1.0) * 0.5)

	return math.Min(density, 1.0)
}

// calculateIndustryRelevance calculates relevance to target industry
func (esa *EnhancedScoringAlgorithm) calculateIndustryRelevance(contextualKeywords []ContextualKeyword, industryID int) float64 {
	if len(contextualKeywords) == 0 {
		return 0.0
	}

	industryKeywords := esa.getIndustrySpecificKeywords(industryID)
	industryKeywordMap := make(map[string]IndustrySpecificKeyword)
	for _, keyword := range industryKeywords {
		industryKeywordMap[strings.ToLower(keyword.Keyword)] = keyword
	}

	relevantKeywords := 0
	totalImportance := 0.0

	for _, contextualKeyword := range contextualKeywords {
		if industryKeyword, exists := industryKeywordMap[strings.ToLower(contextualKeyword.Keyword)]; exists {
			relevantKeywords++
			totalImportance += industryKeyword.Importance
		}
	}

	if relevantKeywords == 0 {
		return 0.0
	}

	// Calculate relevance score
	relevanceRatio := float64(relevantKeywords) / float64(len(contextualKeywords))
	avgImportance := totalImportance / float64(relevantKeywords)

	return (relevanceRatio * 0.6) + (avgImportance * 0.4)
}

// calculateContextConsistency calculates consistency across different contexts
func (esa *EnhancedScoringAlgorithm) calculateContextConsistency(contextualKeywords []ContextualKeyword) float64 {
	if len(contextualKeywords) == 0 {
		return 0.0
	}

	// Group keywords by context
	contextGroups := make(map[string][]string)
	for _, keyword := range contextualKeywords {
		contextGroups[keyword.Context] = append(contextGroups[keyword.Context], keyword.Keyword)
	}

	// Calculate consistency between contexts
	if len(contextGroups) < 2 {
		return 1.0 // Perfect consistency if only one context
	}

	// Find common keywords across contexts
	allKeywords := make(map[string]int)
	for _, keywords := range contextGroups {
		for _, keyword := range keywords {
			allKeywords[keyword]++
		}
	}

	commonKeywords := 0
	for _, count := range allKeywords {
		if count > 1 {
			commonKeywords++
		}
	}

	// Calculate consistency score
	totalUniqueKeywords := len(allKeywords)
	if totalUniqueKeywords == 0 {
		return 0.0
	}

	consistency := float64(commonKeywords) / float64(totalUniqueKeywords)
	return math.Min(consistency, 1.0)
}

// calculateMatchQuality calculates the overall quality of matches
func (esa *EnhancedScoringAlgorithm) calculateMatchQuality(contextualKeywords []ContextualKeyword) float64 {
	if len(contextualKeywords) == 0 {
		return 0.0
	}

	totalQuality := 0.0
	for _, keyword := range contextualKeywords {
		// Quality based on keyword length and context
		lengthQuality := math.Min(float64(len(keyword.Keyword))/10.0, 1.0)

		// Context quality
		contextQuality := 1.0
		switch keyword.Context {
		case "business_name":
			contextQuality = 1.0
		case "description":
			contextQuality = 0.8
		case "website_url":
			contextQuality = 0.6
		}

		keywordQuality := (lengthQuality * 0.4) + (contextQuality * 0.6)
		totalQuality += keywordQuality
	}

	return totalQuality / float64(len(contextualKeywords))
}

// applyDynamicWeightAdjustment applies dynamic weight adjustment to industry scores
func (esa *EnhancedScoringAlgorithm) applyDynamicWeightAdjustment(
	industryScores map[int]*IndustryScore,
	contextualKeywords []ContextualKeyword,
) {
	for industryID, industryScore := range industryScores {
		// Calculate dynamic weight adjustment for this industry
		dynamicAdjustment := esa.calculateDynamicWeightAdjustment(contextualKeywords, industryID)

		// Apply adjustment factor to total score
		industryScore.TotalScore *= dynamicAdjustment.AdjustmentFactor

		// Log the adjustment for debugging
		esa.logger.Printf("🔧 Applied dynamic weight adjustment to industry %d: factor=%.3f, density=%.3f, relevance=%.3f, consistency=%.3f, quality=%.3f",
			industryID,
			dynamicAdjustment.AdjustmentFactor,
			dynamicAdjustment.KeywordDensity,
			dynamicAdjustment.IndustryRelevance,
			dynamicAdjustment.ContextConsistency,
			dynamicAdjustment.MatchQuality,
		)
	}
}
