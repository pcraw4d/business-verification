package industry_codes

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"
)

// RankingStrategy defines different ranking strategies
type RankingStrategy string

const (
	RankingStrategyConfidence    RankingStrategy = "confidence"
	RankingStrategyComposite     RankingStrategy = "composite"
	RankingStrategyWeighted      RankingStrategy = "weighted"
	RankingStrategyMultiCriteria RankingStrategy = "multi_criteria"
)

// RankingCriteria defines the criteria for ranking
type RankingCriteria struct {
	Strategy           RankingStrategy `json:"strategy"`
	ConfidenceWeight   float64         `json:"confidence_weight"`
	RelevanceWeight    float64         `json:"relevance_weight"`
	QualityWeight      float64         `json:"quality_weight"`
	FrequencyWeight    float64         `json:"frequency_weight"`
	MinConfidence      float64         `json:"min_confidence"`
	MaxResultsPerType  int             `json:"max_results_per_type"`
	UseDiversification bool            `json:"use_diversification"`
	EnableTieBreaking  bool            `json:"enable_tie_breaking"`
}

// RankingResult represents a ranked classification result
type RankingResult struct {
	*ClassificationResult
	ConfidenceScore   *ConfidenceScore `json:"confidence_score"`
	RankingScore      float64          `json:"ranking_score"`
	Rank              int              `json:"rank"`
	TypeRank          int              `json:"type_rank"`
	RankingFactors    *RankingFactors  `json:"ranking_factors"`
	TieBreaker        float64          `json:"tie_breaker"`
	SelectionReason   string           `json:"selection_reason"`
	QualityIndicators []string         `json:"quality_indicators"`
}

// RankingFactors represents the factors used in ranking
type RankingFactors struct {
	ConfidenceFactor float64 `json:"confidence_factor"`
	RelevanceFactor  float64 `json:"relevance_factor"`
	QualityFactor    float64 `json:"quality_factor"`
	FrequencyFactor  float64 `json:"frequency_factor"`
	DiversityBonus   float64 `json:"diversity_bonus"`
	PenaltyFactor    float64 `json:"penalty_factor"`
}

// RankedResults represents the complete ranked results
type RankedResults struct {
	OverallResults   []*RankingResult            `json:"overall_results"`
	TopResultsByType map[string][]*RankingResult `json:"top_results_by_type"`
	RankingMetadata  *RankingMetadata            `json:"ranking_metadata"`
	QualityMetrics   *QualityMetrics             `json:"quality_metrics"`
	DiversityMetrics *DiversityMetrics           `json:"diversity_metrics"`
}

// RankingMetadata provides metadata about the ranking process
type RankingMetadata struct {
	Strategy           RankingStrategy `json:"strategy"`
	TotalCandidates    int             `json:"total_candidates"`
	FilteredCandidates int             `json:"filtered_candidates"`
	RankingTime        time.Duration   `json:"ranking_time"`
	CriteriaUsed       []string        `json:"criteria_used"`
	TieBreaksUsed      int             `json:"tie_breaks_used"`
}

// QualityMetrics provides quality metrics for the ranking results
type QualityMetrics struct {
	AverageConfidence   float64        `json:"average_confidence"`
	ConfidenceRange     float64        `json:"confidence_range"`
	QualityDistribution map[string]int `json:"quality_distribution"`
	TypeCoverage        map[string]int `json:"type_coverage"`
	HighQualityCount    int            `json:"high_quality_count"`
	LowQualityCount     int            `json:"low_quality_count"`
}

// DiversityMetrics provides diversity metrics for the ranking results
type DiversityMetrics struct {
	TypeDiversity     float64        `json:"type_diversity"`
	CategoryDiversity float64        `json:"category_diversity"`
	ConfidenceSpread  float64        `json:"confidence_spread"`
	SourceDiversity   map[string]int `json:"source_diversity"`
	DiversityScore    float64        `json:"diversity_score"`
}

// RankingEngine provides advanced ranking and selection capabilities
type RankingEngine struct {
	confidenceScorer *ConfidenceScorer
	logger           *zap.Logger
	defaultCriteria  *RankingCriteria
}

// NewRankingEngine creates a new ranking engine
func NewRankingEngine(confidenceScorer *ConfidenceScorer, logger *zap.Logger) *RankingEngine {
	return &RankingEngine{
		confidenceScorer: confidenceScorer,
		logger:           logger,
		defaultCriteria: &RankingCriteria{
			Strategy:           RankingStrategyComposite,
			ConfidenceWeight:   0.4,
			RelevanceWeight:    0.3,
			QualityWeight:      0.2,
			FrequencyWeight:    0.1,
			MinConfidence:      0.3,
			MaxResultsPerType:  3,
			UseDiversification: true,
			EnableTieBreaking:  true,
		},
	}
}

// RankAndSelectResults ranks and selects the top results for each code type
func (re *RankingEngine) RankAndSelectResults(ctx context.Context, results []*ClassificationResult, request *ClassificationRequest, criteria *RankingCriteria) (*RankedResults, error) {
	startTime := time.Now()

	if criteria == nil {
		criteria = re.defaultCriteria
	}

	re.logger.Info("Starting result ranking and selection",
		zap.String("strategy", string(criteria.Strategy)),
		zap.Int("total_candidates", len(results)),
		zap.Float64("min_confidence", criteria.MinConfidence))

	// Step 1: Calculate confidence scores for all results
	rankedResults, err := re.calculateConfidenceScores(ctx, results, request)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate confidence scores: %w", err)
	}

	// Step 2: Apply confidence filtering
	filteredResults := re.applyConfidenceFiltering(rankedResults, criteria.MinConfidence)

	// Step 3: Calculate ranking scores
	err = re.calculateRankingScores(filteredResults, criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate ranking scores: %w", err)
	}

	// Step 4: Apply ranking strategy
	err = re.applyRankingStrategy(filteredResults, criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to apply ranking strategy: %w", err)
	}

	// Step 5: Group by type and select top results
	topResultsByType := re.groupAndSelectTopByType(filteredResults, criteria)

	// Step 6: Apply diversification if enabled
	if criteria.UseDiversification {
		re.applyDiversification(topResultsByType, criteria)
	}

	// Step 7: Apply tie-breaking if enabled
	if criteria.EnableTieBreaking {
		re.applyTieBreaking(topResultsByType)
	}

	// Step 8: Calculate metrics
	qualityMetrics := re.calculateQualityMetrics(topResultsByType)
	diversityMetrics := re.calculateDiversityMetrics(topResultsByType)

	// Step 9: Create overall results list
	overallResults := re.createOverallResultsList(topResultsByType)

	rankingTime := time.Since(startTime)

	finalResults := &RankedResults{
		OverallResults:   overallResults,
		TopResultsByType: topResultsByType,
		RankingMetadata: &RankingMetadata{
			Strategy:           criteria.Strategy,
			TotalCandidates:    len(results),
			FilteredCandidates: len(filteredResults),
			RankingTime:        rankingTime,
			CriteriaUsed:       re.getCriteriaUsed(criteria),
			TieBreaksUsed:      re.countTieBreaks(topResultsByType),
		},
		QualityMetrics:   qualityMetrics,
		DiversityMetrics: diversityMetrics,
	}

	re.logger.Info("Ranking and selection completed",
		zap.String("strategy", string(criteria.Strategy)),
		zap.Int("filtered_candidates", len(filteredResults)),
		zap.Duration("ranking_time", rankingTime))

	return finalResults, nil
}

// calculateConfidenceScores calculates detailed confidence scores for all results
func (re *RankingEngine) calculateConfidenceScores(ctx context.Context, results []*ClassificationResult, request *ClassificationRequest) ([]*RankingResult, error) {
	rankedResults := make([]*RankingResult, 0, len(results))

	for _, result := range results {
		confidenceScore, err := re.confidenceScorer.CalculateConfidence(ctx, result, request)
		if err != nil {
			re.logger.Warn("Failed to calculate confidence score for result",
				zap.String("code", result.Code.Code),
				zap.Error(err))
			// Use the original confidence as fallback
			confidenceScore = &ConfidenceScore{
				OverallScore:     result.Confidence,
				ConfidenceLevel:  "unknown",
				ValidationStatus: "unchecked",
				Factors: &ConfidenceFactors{
					TextMatchScore:    result.Confidence,
					KeywordMatchScore: result.Confidence,
				},
				LastUpdated:  time.Now(),
				ScoreVersion: "fallback",
			}
		}

		rankedResult := &RankingResult{
			ClassificationResult: result,
			ConfidenceScore:      confidenceScore,
			RankingScore:         0.0, // Will be calculated later
			Rank:                 0,   // Will be assigned later
			TypeRank:             0,   // Will be assigned later
			RankingFactors: &RankingFactors{
				ConfidenceFactor: confidenceScore.OverallScore,
			},
			QualityIndicators: []string{},
		}

		rankedResults = append(rankedResults, rankedResult)
	}

	return rankedResults, nil
}

// applyConfidenceFiltering filters results by minimum confidence
func (re *RankingEngine) applyConfidenceFiltering(results []*RankingResult, minConfidence float64) []*RankingResult {
	filteredResults := make([]*RankingResult, 0, len(results))

	for _, result := range results {
		if result.ConfidenceScore.OverallScore >= minConfidence {
			filteredResults = append(filteredResults, result)
		}
	}

	re.logger.Debug("Applied confidence filtering",
		zap.Int("original_count", len(results)),
		zap.Int("filtered_count", len(filteredResults)),
		zap.Float64("min_confidence", minConfidence))

	return filteredResults
}

// calculateRankingScores calculates the ranking score for each result
func (re *RankingEngine) calculateRankingScores(results []*RankingResult, criteria *RankingCriteria) error {
	for _, result := range results {
		factors := result.RankingFactors

		// Calculate relevance factor based on match type and reasons
		factors.RelevanceFactor = re.calculateRelevanceFactor(result)

		// Calculate quality factor based on confidence score factors
		factors.QualityFactor = re.calculateQualityFactor(result)

		// Calculate frequency factor based on code usage
		factors.FrequencyFactor = re.calculateFrequencyFactor(result)

		// Calculate composite ranking score
		result.RankingScore = re.calculateCompositeScore(factors, criteria)

		// Add quality indicators
		result.QualityIndicators = re.generateQualityIndicators(result)

		// Generate selection reason
		result.SelectionReason = re.generateSelectionReason(result, criteria)
	}

	return nil
}

// calculateRelevanceFactor calculates the relevance factor for a result
func (re *RankingEngine) calculateRelevanceFactor(result *RankingResult) float64 {
	factor := 0.5 // Base factor

	// Boost based on match type
	switch result.MatchType {
	case "exact":
		factor += 0.4
	case "keyword":
		factor += 0.3
	case "description":
		factor += 0.2
	case "fuzzy":
		factor += 0.1
	}

	// Boost based on number of matched elements
	if len(result.MatchedOn) > 0 {
		// Logarithmic scaling for matched elements
		matchBoost := math.Log10(float64(len(result.MatchedOn))+1) * 0.1
		factor += matchBoost
	}

	// Boost based on number of reasons
	if len(result.Reasons) > 0 {
		reasonBoost := float64(len(result.Reasons)) * 0.05
		factor += reasonBoost
	}

	return math.Min(factor, 1.0)
}

// calculateQualityFactor calculates the quality factor for a result
func (re *RankingEngine) calculateQualityFactor(result *RankingResult) float64 {
	factors := result.ConfidenceScore.Factors

	// Weight different confidence factors
	qualityScore := factors.TextMatchScore*0.25 +
		factors.KeywordMatchScore*0.25 +
		factors.CodeQualityScore*0.20 +
		factors.UsageFrequencyScore*0.15 +
		factors.ValidationScore*0.15

	return qualityScore
}

// calculateFrequencyFactor calculates the frequency factor for a result
func (re *RankingEngine) calculateFrequencyFactor(result *RankingResult) float64 {
	factors := result.ConfidenceScore.Factors

	// Use usage frequency score as the base
	frequencyFactor := factors.UsageFrequencyScore

	// Boost for popular code types
	switch result.Code.Type {
	case CodeTypeNAICS:
		frequencyFactor += 0.1 // NAICS is widely used
	case CodeTypeSIC:
		frequencyFactor += 0.05 // SIC is less common but still used
	case CodeTypeMCC:
		frequencyFactor += 0.15 // MCC is very common for payments
	}

	return math.Min(frequencyFactor, 1.0)
}

// calculateCompositeScore calculates the final composite ranking score
func (re *RankingEngine) calculateCompositeScore(factors *RankingFactors, criteria *RankingCriteria) float64 {
	score := factors.ConfidenceFactor*criteria.ConfidenceWeight +
		factors.RelevanceFactor*criteria.RelevanceWeight +
		factors.QualityFactor*criteria.QualityWeight +
		factors.FrequencyFactor*criteria.FrequencyWeight

	// Apply diversity bonus if applicable
	score += factors.DiversityBonus

	// Apply penalty factor if applicable
	score -= factors.PenaltyFactor

	return math.Max(0.0, math.Min(score, 1.0))
}

// applyRankingStrategy applies the specified ranking strategy
func (re *RankingEngine) applyRankingStrategy(results []*RankingResult, criteria *RankingCriteria) error {
	switch criteria.Strategy {
	case RankingStrategyConfidence:
		re.applyConfidenceRanking(results)
	case RankingStrategyComposite:
		re.applyCompositeRanking(results)
	case RankingStrategyWeighted:
		re.applyWeightedRanking(results, criteria)
	case RankingStrategyMultiCriteria:
		re.applyMultiCriteriaRanking(results, criteria)
	default:
		return fmt.Errorf("unsupported ranking strategy: %s", criteria.Strategy)
	}

	return nil
}

// applyConfidenceRanking sorts results by confidence score only
func (re *RankingEngine) applyConfidenceRanking(results []*RankingResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].ConfidenceScore.OverallScore > results[j].ConfidenceScore.OverallScore
	})

	// Assign ranks
	for i, result := range results {
		result.Rank = i + 1
	}
}

// applyCompositeRanking sorts results by composite ranking score
func (re *RankingEngine) applyCompositeRanking(results []*RankingResult) {
	sort.Slice(results, func(i, j int) bool {
		if math.Abs(results[i].RankingScore-results[j].RankingScore) < 0.001 {
			// Tie-breaker: use confidence score
			return results[i].ConfidenceScore.OverallScore > results[j].ConfidenceScore.OverallScore
		}
		return results[i].RankingScore > results[j].RankingScore
	})

	// Assign ranks
	for i, result := range results {
		result.Rank = i + 1
	}
}

// applyWeightedRanking applies weighted ranking based on custom weights
func (re *RankingEngine) applyWeightedRanking(results []*RankingResult, criteria *RankingCriteria) {
	// Recalculate scores with custom weights
	for _, result := range results {
		result.RankingScore = re.calculateCompositeScore(result.RankingFactors, criteria)
	}

	// Sort by recalculated scores
	re.applyCompositeRanking(results)
}

// applyMultiCriteriaRanking applies multi-criteria decision analysis
func (re *RankingEngine) applyMultiCriteriaRanking(results []*RankingResult, criteria *RankingCriteria) {
	// Implement TOPSIS (Technique for Order Preference by Similarity to Ideal Solution)
	re.applyTOPSISRanking(results, criteria)
}

// applyTOPSISRanking implements TOPSIS multi-criteria ranking
func (re *RankingEngine) applyTOPSISRanking(results []*RankingResult, criteria *RankingCriteria) {
	if len(results) == 0 {
		return
	}

	// Define criteria weights
	weights := []float64{
		criteria.ConfidenceWeight,
		criteria.RelevanceWeight,
		criteria.QualityWeight,
		criteria.FrequencyWeight,
	}

	// Build decision matrix
	matrix := make([][]float64, len(results))
	for i, result := range results {
		matrix[i] = []float64{
			result.RankingFactors.ConfidenceFactor,
			result.RankingFactors.RelevanceFactor,
			result.RankingFactors.QualityFactor,
			result.RankingFactors.FrequencyFactor,
		}
	}

	// Calculate TOPSIS scores
	topsisScores := re.calculateTOPSISScores(matrix, weights)

	// Assign TOPSIS scores as ranking scores
	for i, result := range results {
		result.RankingScore = topsisScores[i]
	}

	// Sort by TOPSIS scores
	re.applyCompositeRanking(results)
}

// calculateTOPSISScores calculates TOPSIS scores for the decision matrix
func (re *RankingEngine) calculateTOPSISScores(matrix [][]float64, weights []float64) []float64 {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return []float64{}
	}

	rows := len(matrix)
	cols := len(matrix[0])

	// Step 1: Normalize the decision matrix
	normalizedMatrix := make([][]float64, rows)
	for i := range normalizedMatrix {
		normalizedMatrix[i] = make([]float64, cols)
	}

	for j := 0; j < cols; j++ {
		// Calculate column sum of squares
		sumSquares := 0.0
		for i := 0; i < rows; i++ {
			sumSquares += matrix[i][j] * matrix[i][j]
		}
		norm := math.Sqrt(sumSquares)

		// Normalize column
		for i := 0; i < rows; i++ {
			if norm > 0 {
				normalizedMatrix[i][j] = matrix[i][j] / norm
			} else {
				normalizedMatrix[i][j] = 0
			}
		}
	}

	// Step 2: Calculate weighted normalized matrix
	weightedMatrix := make([][]float64, rows)
	for i := range weightedMatrix {
		weightedMatrix[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			weightedMatrix[i][j] = normalizedMatrix[i][j] * weights[j]
		}
	}

	// Step 3: Determine ideal and negative-ideal solutions
	idealSolution := make([]float64, cols)
	negativeIdealSolution := make([]float64, cols)

	for j := 0; j < cols; j++ {
		max := weightedMatrix[0][j]
		min := weightedMatrix[0][j]

		for i := 1; i < rows; i++ {
			if weightedMatrix[i][j] > max {
				max = weightedMatrix[i][j]
			}
			if weightedMatrix[i][j] < min {
				min = weightedMatrix[i][j]
			}
		}

		idealSolution[j] = max
		negativeIdealSolution[j] = min
	}

	// Step 4: Calculate distances to ideal and negative-ideal solutions
	scores := make([]float64, rows)
	for i := 0; i < rows; i++ {
		distanceToIdeal := 0.0
		distanceToNegativeIdeal := 0.0

		for j := 0; j < cols; j++ {
			diffIdeal := weightedMatrix[i][j] - idealSolution[j]
			diffNegativeIdeal := weightedMatrix[i][j] - negativeIdealSolution[j]

			distanceToIdeal += diffIdeal * diffIdeal
			distanceToNegativeIdeal += diffNegativeIdeal * diffNegativeIdeal
		}

		distanceToIdeal = math.Sqrt(distanceToIdeal)
		distanceToNegativeIdeal = math.Sqrt(distanceToNegativeIdeal)

		// Step 5: Calculate relative closeness to ideal solution
		if distanceToIdeal+distanceToNegativeIdeal > 0 {
			scores[i] = distanceToNegativeIdeal / (distanceToIdeal + distanceToNegativeIdeal)
		} else {
			scores[i] = 0.5
		}
	}

	return scores
}

// groupAndSelectTopByType groups results by code type and selects top results
func (re *RankingEngine) groupAndSelectTopByType(results []*RankingResult, criteria *RankingCriteria) map[string][]*RankingResult {
	groupedResults := make(map[string][]*RankingResult)

	// Group by code type
	for _, result := range results {
		codeType := string(result.Code.Type)
		groupedResults[codeType] = append(groupedResults[codeType], result)
	}

	// Sort each group and select top results
	for codeType, typeResults := range groupedResults {
		// Sort by ranking score
		sort.Slice(typeResults, func(i, j int) bool {
			if math.Abs(typeResults[i].RankingScore-typeResults[j].RankingScore) < 0.001 {
				// Tie-breaker: use confidence score
				return typeResults[i].ConfidenceScore.OverallScore > typeResults[j].ConfidenceScore.OverallScore
			}
			return typeResults[i].RankingScore > typeResults[j].RankingScore
		})

		// Assign type ranks
		for i, result := range typeResults {
			result.TypeRank = i + 1
		}

		// Select top results
		if len(typeResults) > criteria.MaxResultsPerType {
			groupedResults[codeType] = typeResults[:criteria.MaxResultsPerType]
		}
	}

	return groupedResults
}

// applyDiversification applies diversification to improve result variety
func (re *RankingEngine) applyDiversification(groupedResults map[string][]*RankingResult, criteria *RankingCriteria) {
	for codeType, results := range groupedResults {
		if len(results) <= 1 {
			continue
		}

		// Calculate diversity bonuses
		categories := make(map[string]bool)
		for _, result := range results {
			category := result.Code.Category

			// Apply diversity bonus if this is a new category
			if !categories[category] {
				result.RankingFactors.DiversityBonus = 0.05
				categories[category] = true
			}

			// Recalculate ranking score with diversity bonus
			result.RankingScore = re.calculateCompositeScore(result.RankingFactors, criteria)
		}

		// Re-sort after applying diversity bonuses
		sort.Slice(results, func(i, j int) bool {
			return results[i].RankingScore > results[j].RankingScore
		})

		groupedResults[codeType] = results
	}
}

// applyTieBreaking applies tie-breaking mechanisms
func (re *RankingEngine) applyTieBreaking(groupedResults map[string][]*RankingResult) {
	for _, results := range groupedResults {
		for i := 0; i < len(results); i++ {
			result := results[i]

			// Calculate tie-breaker score based on multiple factors
			tieBreaker := result.ConfidenceScore.OverallScore*0.4 +
				result.RankingFactors.QualityFactor*0.3 +
				result.RankingFactors.FrequencyFactor*0.2 +
				result.RankingFactors.RelevanceFactor*0.1

			result.TieBreaker = tieBreaker
		}

		// Final sort with tie-breakers
		sort.Slice(results, func(i, j int) bool {
			if math.Abs(results[i].RankingScore-results[j].RankingScore) < 0.001 {
				return results[i].TieBreaker > results[j].TieBreaker
			}
			return results[i].RankingScore > results[j].RankingScore
		})
	}
}

// generateQualityIndicators generates quality indicators for a result
func (re *RankingEngine) generateQualityIndicators(result *RankingResult) []string {
	indicators := []string{}

	// High confidence indicator
	if result.ConfidenceScore.OverallScore >= 0.8 {
		indicators = append(indicators, "high_confidence")
	}

	// Strong match indicator
	if len(result.MatchedOn) >= 3 {
		indicators = append(indicators, "strong_match")
	}

	// Quality data indicator
	if result.ConfidenceScore.Factors.CodeQualityScore >= 0.7 {
		indicators = append(indicators, "quality_data")
	}

	// Frequent usage indicator
	if result.ConfidenceScore.Factors.UsageFrequencyScore >= 0.6 {
		indicators = append(indicators, "frequent_usage")
	}

	// Validation passed indicator
	if result.ConfidenceScore.ValidationStatus == "valid" {
		indicators = append(indicators, "validation_passed")
	}

	return indicators
}

// generateSelectionReason generates a human-readable selection reason
func (re *RankingEngine) generateSelectionReason(result *RankingResult, criteria *RankingCriteria) string {
	reasons := []string{}

	// Primary reason based on strongest factor
	factors := result.RankingFactors
	maxFactor := math.Max(
		math.Max(factors.ConfidenceFactor, factors.RelevanceFactor),
		math.Max(factors.QualityFactor, factors.FrequencyFactor),
	)

	if maxFactor == factors.ConfidenceFactor {
		reasons = append(reasons, "high confidence score")
	} else if maxFactor == factors.RelevanceFactor {
		reasons = append(reasons, "strong relevance match")
	} else if maxFactor == factors.QualityFactor {
		reasons = append(reasons, "excellent data quality")
	} else if maxFactor == factors.FrequencyFactor {
		reasons = append(reasons, "frequent usage pattern")
	}

	// Additional reasons
	if result.ConfidenceScore.OverallScore >= 0.8 {
		reasons = append(reasons, "exceeds confidence threshold")
	}

	if len(result.MatchedOn) >= 2 {
		reasons = append(reasons, "multiple match criteria")
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "meets selection criteria")
	}

	// Combine reasons
	if len(reasons) == 1 {
		return "Selected for " + reasons[0]
	} else if len(reasons) == 2 {
		return "Selected for " + reasons[0] + " and " + reasons[1]
	} else {
		return "Selected for " + reasons[0] + " and other factors"
	}
}

// calculateQualityMetrics calculates quality metrics for the ranked results
func (re *RankingEngine) calculateQualityMetrics(groupedResults map[string][]*RankingResult) *QualityMetrics {
	var allResults []*RankingResult
	qualityDistribution := make(map[string]int)
	typeCoverage := make(map[string]int)

	totalConfidence := 0.0
	highQualityCount := 0
	lowQualityCount := 0
	minConfidence := 1.0
	maxConfidence := 0.0

	// Flatten results and calculate metrics
	for codeType, results := range groupedResults {
		typeCoverage[codeType] = len(results)
		allResults = append(allResults, results...)

		for _, result := range results {
			confidence := result.ConfidenceScore.OverallScore
			totalConfidence += confidence

			if confidence > maxConfidence {
				maxConfidence = confidence
			}
			if confidence < minConfidence {
				minConfidence = confidence
			}

			// Quality distribution
			level := result.ConfidenceScore.ConfidenceLevel
			qualityDistribution[level]++

			// Quality counts
			if confidence >= 0.8 {
				highQualityCount++
			} else if confidence < 0.5 {
				lowQualityCount++
			}
		}
	}

	averageConfidence := 0.0
	if len(allResults) > 0 {
		averageConfidence = totalConfidence / float64(len(allResults))
	}

	confidenceRange := maxConfidence - minConfidence

	return &QualityMetrics{
		AverageConfidence:   averageConfidence,
		ConfidenceRange:     confidenceRange,
		QualityDistribution: qualityDistribution,
		TypeCoverage:        typeCoverage,
		HighQualityCount:    highQualityCount,
		LowQualityCount:     lowQualityCount,
	}
}

// calculateDiversityMetrics calculates diversity metrics for the ranked results
func (re *RankingEngine) calculateDiversityMetrics(groupedResults map[string][]*RankingResult) *DiversityMetrics {
	typeCount := len(groupedResults)
	categories := make(map[string]bool)
	sourceDiversity := make(map[string]int)
	confidences := []float64{}

	for _, results := range groupedResults {
		for _, result := range results {
			// Category diversity
			categories[result.Code.Category] = true

			// Source diversity (based on match type)
			sourceDiversity[result.MatchType]++

			// Confidence spread
			confidences = append(confidences, result.ConfidenceScore.OverallScore)
		}
	}

	// Calculate diversity scores
	typeDiversity := float64(typeCount) / 3.0            // Assuming max 3 types (NAICS, SIC, MCC)
	categoryDiversity := float64(len(categories)) / 10.0 // Assuming max 10 categories

	// Calculate confidence spread (standard deviation)
	confidenceSpread := 0.0
	if len(confidences) > 1 {
		mean := 0.0
		for _, conf := range confidences {
			mean += conf
		}
		mean /= float64(len(confidences))

		variance := 0.0
		for _, conf := range confidences {
			variance += (conf - mean) * (conf - mean)
		}
		variance /= float64(len(confidences))
		confidenceSpread = math.Sqrt(variance)
	}

	// Overall diversity score
	diversityScore := (typeDiversity + categoryDiversity) / 2.0

	return &DiversityMetrics{
		TypeDiversity:     typeDiversity,
		CategoryDiversity: categoryDiversity,
		ConfidenceSpread:  confidenceSpread,
		SourceDiversity:   sourceDiversity,
		DiversityScore:    diversityScore,
	}
}

// createOverallResultsList creates a unified list of all top results
func (re *RankingEngine) createOverallResultsList(groupedResults map[string][]*RankingResult) []*RankingResult {
	var allResults []*RankingResult

	for _, results := range groupedResults {
		allResults = append(allResults, results...)
	}

	// Sort by ranking score
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].RankingScore > allResults[j].RankingScore
	})

	// Assign overall ranks
	for i, result := range allResults {
		result.Rank = i + 1
	}

	return allResults
}

// getCriteriaUsed returns a list of criteria used in ranking
func (re *RankingEngine) getCriteriaUsed(criteria *RankingCriteria) []string {
	used := []string{}

	if criteria.ConfidenceWeight > 0 {
		used = append(used, "confidence")
	}
	if criteria.RelevanceWeight > 0 {
		used = append(used, "relevance")
	}
	if criteria.QualityWeight > 0 {
		used = append(used, "quality")
	}
	if criteria.FrequencyWeight > 0 {
		used = append(used, "frequency")
	}
	if criteria.UseDiversification {
		used = append(used, "diversification")
	}
	if criteria.EnableTieBreaking {
		used = append(used, "tie_breaking")
	}

	return used
}

// countTieBreaks counts the number of tie-breaks used
func (re *RankingEngine) countTieBreaks(groupedResults map[string][]*RankingResult) int {
	tieBreaks := 0

	for _, results := range groupedResults {
		for i := 1; i < len(results); i++ {
			if math.Abs(results[i-1].RankingScore-results[i].RankingScore) < 0.001 {
				tieBreaks++
			}
		}
	}

	return tieBreaks
}
