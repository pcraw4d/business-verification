package industry_codes

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"
)

// VotingStrategy defines different voting strategies
type VotingStrategy string

const (
	VotingStrategyMajority        VotingStrategy = "majority"
	VotingStrategyWeightedAverage VotingStrategy = "weighted_average"
	VotingStrategyBordaCount      VotingStrategy = "borda_count"
	VotingStrategyConsensus       VotingStrategy = "consensus"
	VotingStrategyRankAggregation VotingStrategy = "rank_aggregation"
)

// VotingConfig defines the configuration for voting
type VotingConfig struct {
	Strategy               VotingStrategy `json:"strategy"`
	MinVoters              int            `json:"min_voters"`
	RequiredAgreement      float64        `json:"required_agreement"`
	ConfidenceWeight       float64        `json:"confidence_weight"`
	ConsistencyWeight      float64        `json:"consistency_weight"`
	DiversityWeight        float64        `json:"diversity_weight"`
	EnableTieBreaking      bool           `json:"enable_tie_breaking"`
	EnableOutlierFiltering bool           `json:"enable_outlier_filtering"`
	OutlierThreshold       float64        `json:"outlier_threshold"`
}

// StrategyVote represents a vote from a single classification strategy
type StrategyVote struct {
	StrategyName string                  `json:"strategy_name"`
	Results      []*ClassificationResult `json:"results"`
	Weight       float64                 `json:"weight"`
	Confidence   float64                 `json:"confidence"`
	VoteTime     time.Time               `json:"vote_time"`
	Metadata     map[string]interface{}  `json:"metadata"`
}

// VotingResult represents the result of a voting session
type VotingResult struct {
	FinalResults       []*ClassificationResult `json:"final_results"`
	VotingScore        float64                 `json:"voting_score"`
	Agreement          float64                 `json:"agreement"`
	Consistency        float64                 `json:"consistency"`
	Diversity          float64                 `json:"diversity"`
	ParticipatingVotes int                     `json:"participating_votes"`
	VotingStrategy     VotingStrategy          `json:"voting_strategy"`
	Metadata           *VotingMetadata         `json:"metadata"`
}

// VotingMetadata contains detailed information about the voting process
type VotingMetadata struct {
	VoteBreakdown      []*StrategySummary `json:"vote_breakdown"`
	AgreementMatrix    map[string]float64 `json:"agreement_matrix"`
	OutliersDetected   []string           `json:"outliers_detected"`
	TieBreakingApplied bool               `json:"tie_breaking_applied"`
	ProcessingTime     time.Duration      `json:"processing_time"`
	QualityIndicators  []string           `json:"quality_indicators"`
}

// StrategySummary summarizes a strategy's contribution to voting
type StrategySummary struct {
	StrategyName   string  `json:"strategy_name"`
	TopCodeVoted   string  `json:"top_code_voted"`
	VoteConfidence float64 `json:"vote_confidence"`
	ResultsCount   int     `json:"results_count"`
	AverageScore   float64 `json:"average_score"`
	Agreement      float64 `json:"agreement"`
}

// CodeVoteAggregation tracks votes for a specific industry code
type CodeVoteAggregation struct {
	Code               *IndustryCode   `json:"code"`
	Votes              []*StrategyVote `json:"votes"`
	TotalVotes         int             `json:"total_votes"`
	WeightedScore      float64         `json:"weighted_score"`
	AverageConfidence  float64         `json:"average_confidence"`
	ConfidenceVariance float64         `json:"confidence_variance"`
	AgreementScore     float64         `json:"agreement_score"`
	BordaPoints        int             `json:"borda_points"`
	RankSum            int             `json:"rank_sum"`
}

// VotingEngine provides majority voting and weighted averaging capabilities
type VotingEngine struct {
	config               *VotingConfig
	logger               *zap.Logger
	confidenceCalculator *ConfidenceCalculator
	votingValidator      *VotingValidator
}

// NewVotingEngine creates a new voting engine
func NewVotingEngine(config *VotingConfig, logger *zap.Logger) *VotingEngine {
	if config == nil {
		config = &VotingConfig{
			Strategy:               VotingStrategyWeightedAverage,
			MinVoters:              2,
			RequiredAgreement:      0.5,
			ConfidenceWeight:       0.4,
			ConsistencyWeight:      0.3,
			DiversityWeight:        0.3,
			EnableTieBreaking:      true,
			EnableOutlierFiltering: true,
			OutlierThreshold:       2.0,
		}
	}

	// Initialize confidence calculator with adaptive weighting enabled
	confidenceConfig := &ConfidenceCalculatorConfig{
		EnableAdaptiveWeighting:    true,
		EnablePerformanceTracking:  true,
		EnableCrossValidation:      false,
		BaseWeightAdjustmentFactor: 0.1,
		PerformanceDecayFactor:     0.95,
		MinimumSampleSize:          10,
	}

	// Initialize voting validator
	votingValidator := NewVotingValidator(nil, logger)

	return &VotingEngine{
		config:               config,
		logger:               logger,
		confidenceCalculator: NewConfidenceCalculator(confidenceConfig, logger),
		votingValidator:      votingValidator,
	}
}

// ConductVoting performs voting on classification results from multiple strategies
func (ve *VotingEngine) ConductVoting(ctx context.Context, votes []*StrategyVote) (*VotingResult, error) {
	startTime := time.Now()

	ve.logger.Info("Starting voting process",
		zap.String("strategy", string(ve.config.Strategy)),
		zap.Int("vote_count", len(votes)))

	// Validate votes
	if err := ve.validateVotes(votes); err != nil {
		return nil, fmt.Errorf("vote validation failed: %w", err)
	}

	// Apply adaptive weighting if enabled
	enhancedVotes, err := ve.confidenceCalculator.CalculateAdaptiveWeights(ctx, votes)
	if err != nil {
		ve.logger.Warn("Failed to calculate adaptive weights, using original weights", zap.Error(err))
		enhancedVotes = votes
	}

	// Filter outliers if enabled
	if ve.config.EnableOutlierFiltering {
		enhancedVotes = ve.filterOutliers(enhancedVotes)
	}

	// Aggregate votes by industry code
	aggregations := ve.aggregateVotesByCode(enhancedVotes)

	// Apply voting strategy
	var finalResults []*ClassificationResult

	switch ve.config.Strategy {
	case VotingStrategyMajority:
		finalResults, err = ve.applyMajorityVoting(aggregations)
	case VotingStrategyWeightedAverage:
		finalResults, err = ve.confidenceCalculator.CalculateEnhancedWeightedAverage(aggregations)
	case VotingStrategyBordaCount:
		finalResults, err = ve.applyBordaCountVoting(aggregations)
	case VotingStrategyConsensus:
		finalResults, err = ve.applyConsensusVoting(aggregations)
	case VotingStrategyRankAggregation:
		finalResults, err = ve.applyRankAggregationVoting(aggregations)
	default:
		return nil, fmt.Errorf("unsupported voting strategy: %s", ve.config.Strategy)
	}

	if err != nil {
		return nil, fmt.Errorf("voting strategy failed: %w", err)
	}

	// Apply tie breaking if needed
	if ve.config.EnableTieBreaking {
		finalResults = ve.applyTieBreaking(finalResults)
	}

	// Calculate voting metrics
	agreement := ve.calculateAgreement(aggregations)
	consistency := ve.calculateConsistency(aggregations)
	diversity := ve.calculateDiversity(aggregations)
	votingScore := ve.calculateOverallVotingScore(agreement, consistency, diversity)

	// Create metadata
	metadata := ve.createVotingMetadata(votes, aggregations, time.Since(startTime))

	result := &VotingResult{
		FinalResults:       finalResults,
		VotingScore:        votingScore,
		Agreement:          agreement,
		Consistency:        consistency,
		Diversity:          diversity,
		ParticipatingVotes: len(votes),
		VotingStrategy:     ve.config.Strategy,
		Metadata:           metadata,
	}

	// Perform voting result validation
	validationResult, err := ve.votingValidator.ValidateVotingResult(ctx, result, votes)
	if err != nil {
		ve.logger.Warn("Voting validation failed, but continuing with result", zap.Error(err))
	} else {
		// Log validation results
		ve.logger.Info("Voting validation completed",
			zap.Bool("is_valid", validationResult.IsValid),
			zap.Float64("validation_score", validationResult.ValidationScore),
			zap.Int("issues", len(validationResult.Issues)),
			zap.Int("warnings", len(validationResult.Warnings)))

		// Add validation metadata to result
		if result.Metadata == nil {
			result.Metadata = &VotingMetadata{}
		}
		result.Metadata.QualityIndicators = append(result.Metadata.QualityIndicators,
			fmt.Sprintf("validation_score:%.3f", validationResult.ValidationScore))

		if !validationResult.IsValid {
			result.Metadata.QualityIndicators = append(result.Metadata.QualityIndicators, "validation_failed")
		}

		if len(validationResult.Issues) > 0 {
			result.Metadata.QualityIndicators = append(result.Metadata.QualityIndicators,
				fmt.Sprintf("validation_issues:%d", len(validationResult.Issues)))
		}
	}

	ve.logger.Info("Voting process completed",
		zap.Float64("voting_score", votingScore),
		zap.Float64("agreement", agreement),
		zap.Int("final_results", len(finalResults)),
		zap.Duration("processing_time", time.Since(startTime)))

	return result, nil
}

// validateVotes validates the input votes
func (ve *VotingEngine) validateVotes(votes []*StrategyVote) error {
	if len(votes) < ve.config.MinVoters {
		return fmt.Errorf("insufficient votes: got %d, minimum required %d", len(votes), ve.config.MinVoters)
	}

	for i, vote := range votes {
		if vote == nil {
			return fmt.Errorf("vote %d is nil", i)
		}
		if vote.StrategyName == "" {
			return fmt.Errorf("vote %d missing strategy name", i)
		}
		if len(vote.Results) == 0 {
			return fmt.Errorf("vote %d has no results", i)
		}
		if vote.Weight < 0 || vote.Weight > 1 {
			return fmt.Errorf("vote %d has invalid weight: %f", i, vote.Weight)
		}
	}

	return nil
}

// filterOutliers removes outlier votes based on confidence and consistency
func (ve *VotingEngine) filterOutliers(votes []*StrategyVote) []*StrategyVote {
	if len(votes) <= 2 {
		return votes // Can't filter outliers with too few votes
	}

	// Calculate mean and standard deviation of vote confidences
	var confidences []float64
	for _, vote := range votes {
		confidences = append(confidences, vote.Confidence)
	}

	mean := ve.calculateMean(confidences)
	stdDev := ve.calculateStandardDeviation(confidences, mean)

	// Filter votes that are within threshold standard deviations
	var filteredVotes []*StrategyVote
	var outliers []string

	for _, vote := range votes {
		zScore := math.Abs(vote.Confidence-mean) / stdDev
		if zScore <= ve.config.OutlierThreshold {
			filteredVotes = append(filteredVotes, vote)
		} else {
			outliers = append(outliers, vote.StrategyName)
		}
	}

	if len(outliers) > 0 {
		ve.logger.Info("Filtered outlier votes",
			zap.Strings("outliers", outliers),
			zap.Float64("threshold", ve.config.OutlierThreshold))
	}

	return filteredVotes
}

// aggregateVotesByCode groups votes by industry code
func (ve *VotingEngine) aggregateVotesByCode(votes []*StrategyVote) map[string]*CodeVoteAggregation {
	aggregations := make(map[string]*CodeVoteAggregation)

	for _, vote := range votes {
		for rank, result := range vote.Results {
			codeKey := fmt.Sprintf("%s-%s", result.Code.Type, result.Code.Code)

			if aggregation, exists := aggregations[codeKey]; exists {
				aggregation.Votes = append(aggregation.Votes, vote)
				aggregation.TotalVotes++
				aggregation.WeightedScore += result.Confidence * vote.Weight
				aggregation.BordaPoints += len(vote.Results) - rank // Higher rank = more points
				aggregation.RankSum += rank + 1                     // 1-based ranking
			} else {
				aggregations[codeKey] = &CodeVoteAggregation{
					Code:              result.Code,
					Votes:             []*StrategyVote{vote},
					TotalVotes:        1,
					WeightedScore:     result.Confidence * vote.Weight,
					AverageConfidence: result.Confidence,
					BordaPoints:       len(vote.Results) - rank,
					RankSum:           rank + 1,
				}
			}
		}
	}

	// Calculate final metrics for each aggregation
	for _, aggregation := range aggregations {
		ve.calculateAggregationMetrics(aggregation)
	}

	return aggregations
}

// calculateAggregationMetrics calculates derived metrics for code aggregations
func (ve *VotingEngine) calculateAggregationMetrics(aggregation *CodeVoteAggregation) {
	if aggregation.TotalVotes == 0 {
		return
	}

	// Calculate average confidence
	var confidenceSum float64
	var confidences []float64
	for _, vote := range aggregation.Votes {
		for _, result := range vote.Results {
			if result.Code.Code == aggregation.Code.Code {
				confidenceSum += result.Confidence
				confidences = append(confidences, result.Confidence)
				break
			}
		}
	}

	aggregation.AverageConfidence = confidenceSum / float64(len(confidences))

	// Calculate confidence variance
	if len(confidences) > 1 {
		aggregation.ConfidenceVariance = ve.calculateVariance(confidences, aggregation.AverageConfidence)
	}

	// Calculate agreement score (how much strategies agree on this code)
	agreementSum := 0.0
	for i := 0; i < len(confidences); i++ {
		for j := i + 1; j < len(confidences); j++ {
			// Agreement is inversely related to difference in confidence
			diff := math.Abs(confidences[i] - confidences[j])
			agreement := 1.0 - diff // Max difference is 1.0
			agreementSum += agreement
		}
	}

	if len(confidences) > 1 {
		pairs := float64(len(confidences) * (len(confidences) - 1) / 2)
		aggregation.AgreementScore = agreementSum / pairs
	} else {
		aggregation.AgreementScore = 1.0 // Single vote = perfect agreement
	}
}

// applyMajorityVoting applies majority voting strategy
func (ve *VotingEngine) applyMajorityVoting(aggregations map[string]*CodeVoteAggregation) ([]*ClassificationResult, error) {
	// Sort codes by number of votes (majority)
	var sortedCodes []*CodeVoteAggregation
	for _, aggregation := range aggregations {
		sortedCodes = append(sortedCodes, aggregation)
	}

	sort.Slice(sortedCodes, func(i, j int) bool {
		if sortedCodes[i].TotalVotes == sortedCodes[j].TotalVotes {
			// Tie-break by average confidence
			return sortedCodes[i].AverageConfidence > sortedCodes[j].AverageConfidence
		}
		return sortedCodes[i].TotalVotes > sortedCodes[j].TotalVotes
	})

	var results []*ClassificationResult
	for i, aggregation := range sortedCodes {
		if aggregation.TotalVotes < int(math.Ceil(float64(len(aggregation.Votes))/2.0)) {
			break // No majority
		}

		result := &ClassificationResult{
			Code:       aggregation.Code,
			Confidence: aggregation.AverageConfidence,
			MatchType:  "majority_vote",
			MatchedOn:  []string{"majority_consensus"},
			Reasons:    []string{fmt.Sprintf("Majority vote: %d/%d strategies", aggregation.TotalVotes, len(aggregation.Votes))},
			Weight:     1.0,
		}
		results = append(results, result)

		if i >= 10 { // Limit results
			break
		}
	}

	return results, nil
}

// applyWeightedAverageVoting applies weighted average voting strategy
func (ve *VotingEngine) applyWeightedAverageVoting(aggregations map[string]*CodeVoteAggregation) ([]*ClassificationResult, error) {
	// Sort codes by weighted score
	var sortedCodes []*CodeVoteAggregation
	for _, aggregation := range aggregations {
		sortedCodes = append(sortedCodes, aggregation)
	}

	sort.Slice(sortedCodes, func(i, j int) bool {
		return sortedCodes[i].WeightedScore > sortedCodes[j].WeightedScore
	})

	var results []*ClassificationResult
	for i, aggregation := range sortedCodes {
		// Calculate final confidence using weighted average
		finalConfidence := aggregation.WeightedScore / float64(aggregation.TotalVotes)

		// Boost confidence based on agreement
		agreementBoost := aggregation.AgreementScore * 0.2
		finalConfidence = math.Min(1.0, finalConfidence+agreementBoost)

		result := &ClassificationResult{
			Code:       aggregation.Code,
			Confidence: finalConfidence,
			MatchType:  "weighted_average",
			MatchedOn:  []string{"weighted_consensus"},
			Reasons:    []string{fmt.Sprintf("Weighted average: %.3f (agreement: %.3f)", finalConfidence, aggregation.AgreementScore)},
			Weight:     1.0,
		}
		results = append(results, result)

		if i >= 10 { // Limit results
			break
		}
	}

	return results, nil
}

// applyBordaCountVoting applies Borda count voting strategy
func (ve *VotingEngine) applyBordaCountVoting(aggregations map[string]*CodeVoteAggregation) ([]*ClassificationResult, error) {
	// Sort codes by Borda points
	var sortedCodes []*CodeVoteAggregation
	for _, aggregation := range aggregations {
		sortedCodes = append(sortedCodes, aggregation)
	}

	sort.Slice(sortedCodes, func(i, j int) bool {
		if sortedCodes[i].BordaPoints == sortedCodes[j].BordaPoints {
			// Tie-break by average confidence
			return sortedCodes[i].AverageConfidence > sortedCodes[j].AverageConfidence
		}
		return sortedCodes[i].BordaPoints > sortedCodes[j].BordaPoints
	})

	var results []*ClassificationResult
	for i, aggregation := range sortedCodes {
		// Normalize Borda points to confidence score
		maxPossiblePoints := len(aggregation.Votes) * 10 // Assuming max 10 results per vote
		normalizedScore := float64(aggregation.BordaPoints) / float64(maxPossiblePoints)
		finalConfidence := math.Min(1.0, normalizedScore)

		result := &ClassificationResult{
			Code:       aggregation.Code,
			Confidence: finalConfidence,
			MatchType:  "borda_count",
			MatchedOn:  []string{"borda_ranking"},
			Reasons:    []string{fmt.Sprintf("Borda count: %d points", aggregation.BordaPoints)},
			Weight:     1.0,
		}
		results = append(results, result)

		if i >= 10 { // Limit results
			break
		}
	}

	return results, nil
}

// applyConsensusVoting applies consensus voting strategy
func (ve *VotingEngine) applyConsensusVoting(aggregations map[string]*CodeVoteAggregation) ([]*ClassificationResult, error) {
	var results []*ClassificationResult

	for _, aggregation := range aggregations {
		if aggregation.AgreementScore >= ve.config.RequiredAgreement {
			result := &ClassificationResult{
				Code:       aggregation.Code,
				Confidence: aggregation.AverageConfidence * aggregation.AgreementScore,
				MatchType:  "consensus",
				MatchedOn:  []string{"consensus_agreement"},
				Reasons:    []string{fmt.Sprintf("Consensus achieved: %.3f agreement", aggregation.AgreementScore)},
				Weight:     1.0,
			}
			results = append(results, result)
		}
	}

	// Sort by confidence
	sort.Slice(results, func(i, j int) bool {
		return results[i].Confidence > results[j].Confidence
	})

	if len(results) > 10 {
		results = results[:10] // Limit results
	}

	return results, nil
}

// applyRankAggregationVoting applies rank aggregation voting strategy
func (ve *VotingEngine) applyRankAggregationVoting(aggregations map[string]*CodeVoteAggregation) ([]*ClassificationResult, error) {
	// Sort codes by average rank (lower rank sum = better)
	var sortedCodes []*CodeVoteAggregation
	for _, aggregation := range aggregations {
		sortedCodes = append(sortedCodes, aggregation)
	}

	sort.Slice(sortedCodes, func(i, j int) bool {
		avgRankI := float64(sortedCodes[i].RankSum) / float64(sortedCodes[i].TotalVotes)
		avgRankJ := float64(sortedCodes[j].RankSum) / float64(sortedCodes[j].TotalVotes)
		return avgRankI < avgRankJ // Lower average rank is better
	})

	var results []*ClassificationResult
	for i, aggregation := range sortedCodes {
		avgRank := float64(aggregation.RankSum) / float64(aggregation.TotalVotes)
		// Convert rank to confidence (lower rank = higher confidence)
		finalConfidence := math.Max(0.1, 1.0-(avgRank-1.0)/10.0) // Normalize rank to confidence

		result := &ClassificationResult{
			Code:       aggregation.Code,
			Confidence: finalConfidence,
			MatchType:  "rank_aggregation",
			MatchedOn:  []string{"aggregated_ranking"},
			Reasons:    []string{fmt.Sprintf("Average rank: %.2f", avgRank)},
			Weight:     1.0,
		}
		results = append(results, result)

		if i >= 10 { // Limit results
			break
		}
	}

	return results, nil
}

// applyTieBreaking applies tie-breaking logic to results with similar confidence
func (ve *VotingEngine) applyTieBreaking(results []*ClassificationResult) []*ClassificationResult {
	if len(results) <= 1 {
		return results
	}

	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			// Check if confidences are very close (within 0.01)
			if math.Abs(results[i].Confidence-results[j].Confidence) < 0.01 {
				// Apply tie-breaking: prefer more specific industry codes
				if len(results[i].Code.Description) > len(results[j].Code.Description) {
					results[i].Confidence += 0.005 // Small boost for more detailed description
				} else {
					results[j].Confidence += 0.005
				}
			}
		}
	}

	// Re-sort after tie-breaking
	sort.Slice(results, func(i, j int) bool {
		return results[i].Confidence > results[j].Confidence
	})

	return results
}

// Helper functions for statistical calculations

func (ve *VotingEngine) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (ve *VotingEngine) calculateStandardDeviation(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}
	variance := ve.calculateVariance(values, mean)
	return math.Sqrt(variance)
}

func (ve *VotingEngine) calculateVariance(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	return sumSquares / float64(len(values)-1)
}

// calculateAgreement calculates overall agreement between voting strategies
func (ve *VotingEngine) calculateAgreement(aggregations map[string]*CodeVoteAggregation) float64 {
	if len(aggregations) == 0 {
		return 0
	}

	totalAgreement := 0.0
	for _, aggregation := range aggregations {
		totalAgreement += aggregation.AgreementScore
	}

	return totalAgreement / float64(len(aggregations))
}

// calculateConsistency calculates consistency in voting results
func (ve *VotingEngine) calculateConsistency(aggregations map[string]*CodeVoteAggregation) float64 {
	if len(aggregations) == 0 {
		return 0
	}

	totalConsistency := 0.0
	for _, aggregation := range aggregations {
		// Consistency is inversely related to confidence variance
		consistency := 1.0 - math.Min(1.0, aggregation.ConfidenceVariance)
		totalConsistency += consistency
	}

	return totalConsistency / float64(len(aggregations))
}

// calculateDiversity calculates diversity in voting results
func (ve *VotingEngine) calculateDiversity(aggregations map[string]*CodeVoteAggregation) float64 {
	if len(aggregations) == 0 {
		return 0
	}

	// Count unique code types
	codeTypes := make(map[CodeType]bool)
	for _, aggregation := range aggregations {
		codeTypes[aggregation.Code.Type] = true
	}

	// Diversity is the ratio of unique types to total possible types (3: MCC, SIC, NAICS)
	return float64(len(codeTypes)) / 3.0
}

// calculateOverallVotingScore calculates overall voting quality score
func (ve *VotingEngine) calculateOverallVotingScore(agreement, consistency, diversity float64) float64 {
	score := agreement*ve.config.ConfidenceWeight +
		consistency*ve.config.ConsistencyWeight +
		diversity*ve.config.DiversityWeight

	return math.Min(1.0, score)
}

// createVotingMetadata creates detailed metadata about the voting process
func (ve *VotingEngine) createVotingMetadata(votes []*StrategyVote, aggregations map[string]*CodeVoteAggregation, processingTime time.Duration) *VotingMetadata {
	metadata := &VotingMetadata{
		VoteBreakdown:      make([]*StrategySummary, 0, len(votes)),
		AgreementMatrix:    make(map[string]float64),
		OutliersDetected:   make([]string, 0),
		TieBreakingApplied: ve.config.EnableTieBreaking,
		ProcessingTime:     processingTime,
		QualityIndicators:  make([]string, 0),
	}

	// Create vote breakdown
	for _, vote := range votes {
		summary := &StrategySummary{
			StrategyName:   vote.StrategyName,
			VoteConfidence: vote.Confidence,
			ResultsCount:   len(vote.Results),
		}

		if len(vote.Results) > 0 {
			summary.TopCodeVoted = vote.Results[0].Code.Code
			totalScore := 0.0
			for _, result := range vote.Results {
				totalScore += result.Confidence
			}
			summary.AverageScore = totalScore / float64(len(vote.Results))
		}

		metadata.VoteBreakdown = append(metadata.VoteBreakdown, summary)
	}

	// Add quality indicators
	if len(aggregations) > 5 {
		metadata.QualityIndicators = append(metadata.QualityIndicators, "diverse_results")
	}
	if len(votes) >= 3 {
		metadata.QualityIndicators = append(metadata.QualityIndicators, "multi_strategy_consensus")
	}

	return metadata
}
