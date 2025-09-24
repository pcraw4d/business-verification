package confidence

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// ConfidenceFactors represents the individual factors used in confidence calculation
type ConfidenceFactors struct {
	MatchRatioFactor        float64 `json:"match_ratio_factor"`        // 30% weight - ratio of matched keywords
	ScoreStrengthFactor     float64 `json:"score_strength_factor"`     // 40% weight - raw score strength
	IndustryThresholdFactor float64 `json:"industry_threshold_factor"` // 10% weight - industry-specific threshold
	SpecificityFactor       float64 `json:"specificity_factor"`        // 20% weight - keyword specificity
}

// ConfidenceCalculationResult represents the result of confidence calculation
type ConfidenceCalculationResult struct {
	FinalConfidence    float64           `json:"final_confidence"`
	Factors            ConfidenceFactors `json:"factors"`
	IndustryID         int               `json:"industry_id"`
	IndustryName       string            `json:"industry_name"`
	MatchedKeywords    []string          `json:"matched_keywords"`
	TotalKeywords      int               `json:"total_keywords"`
	RawScore           float64           `json:"raw_score"`
	CalculationTime    time.Duration     `json:"calculation_time"`
	CalculationVersion string            `json:"calculation_version"`
}

// ConfidenceCalculator provides multi-factor confidence calculation
type ConfidenceCalculator struct {
	thresholdService *IndustryThresholdService
	logger           Logger
}

// Logger interface for logging
type Logger interface {
	Printf(format string, v ...interface{})
}

// NewConfidenceCalculator creates a new confidence calculator
func NewConfidenceCalculator(thresholdService *IndustryThresholdService, logger Logger) *ConfidenceCalculator {
	if logger == nil {
		logger = &defaultLogger{}
	}

	return &ConfidenceCalculator{
		thresholdService: thresholdService,
		logger:           logger,
	}
}

// CalculateDynamicConfidence calculates confidence using multi-factor approach
func (cc *ConfidenceCalculator) CalculateDynamicConfidence(
	ctx context.Context,
	industryID int,
	industryName string,
	matchedKeywords []string,
	totalKeywords int,
	rawScore float64,
	industryMatches map[int][]string,
) (*ConfidenceCalculationResult, error) {
	startTime := time.Now()

	cc.logger.Printf("🎯 Calculating dynamic confidence for industry: %s (ID: %d)", industryName, industryID)

	// Calculate individual factors
	factors := cc.calculateConfidenceFactors(
		ctx,
		industryID,
		industryName,
		matchedKeywords,
		totalKeywords,
		rawScore,
		industryMatches,
	)

	// Calculate weighted final confidence
	finalConfidence := cc.calculateWeightedConfidence(factors)

	// Ensure confidence is within bounds
	finalConfidence = math.Max(0.1, math.Min(1.0, finalConfidence))

	calculationTime := time.Since(startTime)

	result := &ConfidenceCalculationResult{
		FinalConfidence:    finalConfidence,
		Factors:            factors,
		IndustryID:         industryID,
		IndustryName:       industryName,
		MatchedKeywords:    matchedKeywords,
		TotalKeywords:      totalKeywords,
		RawScore:           rawScore,
		CalculationTime:    calculationTime,
		CalculationVersion: "2.2.1",
	}

	cc.logger.Printf("✅ Dynamic confidence calculated: %.3f (factors: MR=%.3f, SS=%.3f, IT=%.3f, SP=%.3f) in %v",
		finalConfidence,
		factors.MatchRatioFactor,
		factors.ScoreStrengthFactor,
		factors.IndustryThresholdFactor,
		factors.SpecificityFactor,
		calculationTime)

	return result, nil
}

// calculateConfidenceFactors calculates individual confidence factors
func (cc *ConfidenceCalculator) calculateConfidenceFactors(
	ctx context.Context,
	industryID int,
	industryName string,
	matchedKeywords []string,
	totalKeywords int,
	rawScore float64,
	industryMatches map[int][]string,
) ConfidenceFactors {
	// Factor 1: Match Ratio (30% weight) - How many keywords matched vs total
	matchRatioFactor := cc.calculateMatchRatioFactor(matchedKeywords, totalKeywords)

	// Factor 2: Score Strength (40% weight) - Raw score strength normalized
	scoreStrengthFactor := cc.calculateScoreStrengthFactor(rawScore, totalKeywords)

	// Factor 3: Industry Threshold (10% weight) - Industry-specific threshold adjustment
	industryThresholdFactor := cc.calculateIndustryThresholdFactor(ctx, industryName)

	// Factor 4: Specificity (20% weight) - How specific the keyword matches are
	specificityFactor := cc.calculateSpecificityFactor(matchedKeywords, industryMatches, industryID)

	return ConfidenceFactors{
		MatchRatioFactor:        matchRatioFactor,
		ScoreStrengthFactor:     scoreStrengthFactor,
		IndustryThresholdFactor: industryThresholdFactor,
		SpecificityFactor:       specificityFactor,
	}
}

// calculateMatchRatioFactor calculates the match ratio factor (30% weight)
func (cc *ConfidenceCalculator) calculateMatchRatioFactor(matchedKeywords []string, totalKeywords int) float64 {
	if totalKeywords == 0 {
		return 0.0
	}

	// Calculate ratio of matched keywords to total keywords
	matchRatio := float64(len(matchedKeywords)) / float64(totalKeywords)

	// Apply logarithmic scaling to prevent over-weighting high ratios
	// This ensures that going from 0.8 to 1.0 has less impact than 0.2 to 0.4
	scaledRatio := math.Log(1+matchRatio*9) / math.Log(10) // Maps [0,1] to [0,1] with log scaling

	// Apply bonus for high match ratios
	if matchRatio >= 0.8 {
		scaledRatio *= 1.1 // 10% bonus for high match ratios
	} else if matchRatio >= 0.6 {
		scaledRatio *= 1.05 // 5% bonus for medium-high match ratios
	}

	return math.Min(1.0, scaledRatio)
}

// calculateScoreStrengthFactor calculates the score strength factor (40% weight)
func (cc *ConfidenceCalculator) calculateScoreStrengthFactor(rawScore float64, totalKeywords int) float64 {
	if totalKeywords == 0 {
		return 0.0
	}

	// Normalize score by number of keywords to get average score per keyword
	normalizedScore := rawScore / float64(totalKeywords)

	// Apply sigmoid function to map to [0,1] range with smooth transitions
	// This prevents extreme scores from dominating the calculation
	sigmoidScore := 1.0 / (1.0 + math.Exp(-6*(normalizedScore-0.5)))

	// Apply bonus for very high scores
	if normalizedScore >= 0.8 {
		sigmoidScore *= 1.15 // 15% bonus for very high scores
	} else if normalizedScore >= 0.6 {
		sigmoidScore *= 1.08 // 8% bonus for high scores
	}

	return math.Min(1.0, sigmoidScore)
}

// calculateIndustryThresholdFactor calculates the industry threshold factor (10% weight)
func (cc *ConfidenceCalculator) calculateIndustryThresholdFactor(ctx context.Context, industryName string) float64 {
	// Get industry-specific threshold from the dynamic service
	threshold, err := cc.thresholdService.GetIndustryThreshold(ctx, industryName)
	if err != nil {
		cc.logger.Printf("⚠️ Failed to get threshold for %s, using default: %v", industryName, err)
		threshold = 0.50 // Default threshold
	}

	// Convert threshold to a factor (higher threshold = higher confidence requirement)
	// Industries with higher thresholds get a slight penalty, those with lower get a bonus
	baseFactor := 0.5                              // Base factor
	thresholdAdjustment := (0.8 - threshold) * 0.5 // Adjust based on threshold difference from 0.8

	return math.Max(0.0, math.Min(1.0, baseFactor+thresholdAdjustment))
}

// calculateSpecificityFactor calculates the specificity factor (20% weight)
// Enhanced to include match count factor as specified in the improvement plan
func (cc *ConfidenceCalculator) calculateSpecificityFactor(
	matchedKeywords []string,
	industryMatches map[int][]string,
	industryID int,
) float64 {
	if len(matchedKeywords) == 0 {
		return 0.0
	}

	// Calculate specificity based on multiple factors including match count
	specificityScore := 0.0

	// Factor 1: Match count factor (40% of specificity score) - Higher specificity for more matches
	matchCountScore := cc.calculateMatchCountFactor(len(matchedKeywords))

	// Factor 2: Keyword uniqueness (30% of specificity score)
	uniquenessScore := cc.calculateKeywordUniqueness(matchedKeywords, industryMatches, industryID)

	// Factor 3: Industry focus (20% of specificity score)
	focusScore := cc.calculateIndustryFocus(matchedKeywords, industryID)

	// Factor 4: Keyword quality (10% of specificity score)
	qualityScore := cc.calculateKeywordQuality(matchedKeywords)

	// Weighted combination with enhanced match count emphasis
	specificityScore = (matchCountScore * 0.4) + (uniquenessScore * 0.3) + (focusScore * 0.2) + (qualityScore * 0.1)

	return math.Min(1.0, specificityScore)
}

// calculateKeywordUniqueness calculates how unique the matched keywords are
func (cc *ConfidenceCalculator) calculateKeywordUniqueness(
	matchedKeywords []string,
	industryMatches map[int][]string,
	industryID int,
) float64 {
	if len(matchedKeywords) == 0 {
		return 0.0
	}

	// Count how many industries each keyword appears in
	totalIndustries := len(industryMatches)
	uniqueKeywords := 0

	for _, keyword := range matchedKeywords {
		keywordIndustries := 0
		for _, matches := range industryMatches {
			for _, match := range matches {
				if strings.EqualFold(match, keyword) {
					keywordIndustries++
					break
				}
			}
		}

		// Keywords that appear in fewer industries are more unique
		if keywordIndustries <= totalIndustries/3 { // Appears in less than 1/3 of industries
			uniqueKeywords++
		}
	}

	return float64(uniqueKeywords) / float64(len(matchedKeywords))
}

// calculateIndustryFocus calculates how focused the keywords are on the specific industry
func (cc *ConfidenceCalculator) calculateIndustryFocus(matchedKeywords []string, industryID int) float64 {
	if len(matchedKeywords) == 0 {
		return 0.0
	}

	// Define industry-specific keyword patterns
	industryPatterns := map[int][]string{
		// Restaurant industry patterns
		1: {"restaurant", "dining", "food", "cuisine", "menu", "chef", "kitchen", "cooking"},
		2: {"fast food", "quick service", "drive through", "takeout", "delivery"},
		3: {"beverage", "drink", "bar", "pub", "cocktail", "wine", "beer"},

		// Legal industry patterns
		4: {"law", "legal", "attorney", "lawyer", "court", "litigation", "legal advice"},

		// Healthcare industry patterns
		5: {"medical", "healthcare", "doctor", "clinic", "hospital", "patient", "treatment"},

		// Technology industry patterns
		6: {"software", "technology", "digital", "tech", "programming", "development", "IT"},

		// Add more industry patterns as needed
	}

	patterns, exists := industryPatterns[industryID]
	if !exists {
		return 0.5 // Default focus score for unknown industries
	}

	// Count how many keywords match industry-specific patterns
	focusedKeywords := 0
	for _, keyword := range matchedKeywords {
		normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
		for _, pattern := range patterns {
			if strings.Contains(normalizedKeyword, pattern) || strings.Contains(pattern, normalizedKeyword) {
				focusedKeywords++
				break
			}
		}
	}

	return float64(focusedKeywords) / float64(len(matchedKeywords))
}

// calculateKeywordQuality calculates the quality of matched keywords
func (cc *ConfidenceCalculator) calculateKeywordQuality(matchedKeywords []string) float64 {
	if len(matchedKeywords) == 0 {
		return 0.0
	}

	// Define quality indicators
	highQualityIndicators := []string{
		"restaurant", "dining", "medical", "legal", "software", "technology",
		"healthcare", "construction", "manufacturing", "retail", "education",
	}

	lowQualityIndicators := []string{
		"the", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with",
		"by", "from", "up", "about", "into", "through", "during", "before",
		"after", "above", "below", "between", "among", "under", "over",
	}

	qualityScore := 0.0
	for _, keyword := range matchedKeywords {
		normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))

		// Check for high quality indicators
		for _, indicator := range highQualityIndicators {
			if strings.Contains(normalizedKeyword, indicator) {
				qualityScore += 1.0
				break
			}
		}

		// Check for low quality indicators (common words)
		for _, indicator := range lowQualityIndicators {
			if normalizedKeyword == indicator {
				qualityScore -= 0.5
				break
			}
		}
	}

	// Normalize to [0,1] range
	normalizedQuality := qualityScore / float64(len(matchedKeywords))
	return math.Max(0.0, math.Min(1.0, normalizedQuality))
}

// calculateMatchCountFactor calculates specificity based on number of matched keywords
// Higher specificity for more matches as specified in the improvement plan
func (cc *ConfidenceCalculator) calculateMatchCountFactor(matchCount int) float64 {
	if matchCount <= 0 {
		return 0.0
	}

	// Use logarithmic scaling to provide diminishing returns for very high match counts
	// This ensures that 1-2 matches get low scores, 3-5 matches get medium scores,
	// and 6+ matches get high scores, with diminishing returns beyond that
	logScore := math.Log(float64(matchCount)) / math.Log(10.0) // Log base 10

	// Normalize to [0,1] range with sigmoid-like curve for better distribution
	// This gives us: 1 match ≈ 0.1, 2 matches ≈ 0.3, 3 matches ≈ 0.5,
	// 4 matches ≈ 0.7, 5 matches ≈ 0.8, 6+ matches ≈ 0.9+
	normalizedScore := math.Min(1.0, logScore/1.5) // Scale to reasonable range

	// Apply sigmoid transformation for smooth curve
	sigmoidScore := 1.0 / (1.0 + math.Exp(-6.0*(normalizedScore-0.5)))

	return math.Max(0.0, math.Min(1.0, sigmoidScore))
}

// calculateWeightedConfidence calculates the final weighted confidence score
func (cc *ConfidenceCalculator) calculateWeightedConfidence(factors ConfidenceFactors) float64 {
	// Apply weights as specified in the plan:
	// Match ratio: 30%, Score strength: 40%, Industry threshold: 10%, Specificity: 20%
	weightedConfidence := (factors.MatchRatioFactor * 0.30) +
		(factors.ScoreStrengthFactor * 0.40) +
		(factors.IndustryThresholdFactor * 0.10) +
		(factors.SpecificityFactor * 0.20)

	return weightedConfidence
}

// GetIndustryThreshold returns the confidence threshold for a specific industry
func (cc *ConfidenceCalculator) GetIndustryThreshold(ctx context.Context, industryName string) (float64, error) {
	return cc.thresholdService.GetIndustryThreshold(ctx, industryName)
}

// GetAllIndustryThresholds returns all industry thresholds
func (cc *ConfidenceCalculator) GetAllIndustryThresholds(ctx context.Context) (map[string]float64, error) {
	return cc.thresholdService.GetAllIndustryThresholds(ctx)
}

// RefreshThresholdCache refreshes the threshold cache
func (cc *ConfidenceCalculator) RefreshThresholdCache(ctx context.Context) error {
	return cc.thresholdService.RefreshCache(ctx)
}

// GetThresholdService returns the threshold service for advanced operations
func (cc *ConfidenceCalculator) GetThresholdService() *IndustryThresholdService {
	return cc.thresholdService
}

// defaultLogger provides a simple logger implementation
type defaultLogger struct{}

func (dl *defaultLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}
