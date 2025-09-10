package classification

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/pcraw4d/business-verification/internal/shared"
)

// WeightedConfidenceScorer provides sophisticated weighted confidence scoring across multiple classification methods
type WeightedConfidenceScorer struct {
	logger *log.Logger
	config *WeightedConfidenceConfig
}

// WeightedConfidenceConfig holds configuration for weighted confidence scoring
type WeightedConfidenceConfig struct {
	// Method weights
	KeywordWeight     float64 `json:"keyword_weight"`     // Weight for keyword-based classification
	MLWeight          float64 `json:"ml_weight"`          // Weight for ML-based classification
	DescriptionWeight float64 `json:"description_weight"` // Weight for description-based classification

	// Confidence adjustment factors
	AgreementBonus      float64 `json:"agreement_bonus"`      // Bonus for method agreement
	DisagreementPenalty float64 `json:"disagreement_penalty"` // Penalty for method disagreement
	TimeDecayFactor     float64 `json:"time_decay_factor"`    // Factor for processing time impact

	// Quality factors
	EvidenceWeight         float64 `json:"evidence_weight"`          // Weight for evidence strength
	DataCompletenessWeight float64 `json:"data_completeness_weight"` // Weight for data completeness

	// Thresholds
	MinConfidence      float64 `json:"min_confidence"`      // Minimum confidence threshold
	MaxConfidence      float64 `json:"max_confidence"`      // Maximum confidence threshold
	AgreementThreshold float64 `json:"agreement_threshold"` // Threshold for method agreement
}

// DefaultWeightedConfidenceConfig returns the default configuration for weighted confidence scoring
func DefaultWeightedConfidenceConfig() *WeightedConfidenceConfig {
	return &WeightedConfidenceConfig{
		// Method weights (should sum to 1.0)
		KeywordWeight:     0.35, // Keyword matching is reliable
		MLWeight:          0.35, // ML is sophisticated
		DescriptionWeight: 0.30, // Description analysis is supplementary

		// Confidence adjustment factors
		AgreementBonus:      0.15, // 15% bonus for method agreement
		DisagreementPenalty: 0.10, // 10% penalty for method disagreement
		TimeDecayFactor:     0.05, // 5% impact for processing time

		// Quality factors
		EvidenceWeight:         0.20, // 20% weight for evidence strength
		DataCompletenessWeight: 0.15, // 15% weight for data completeness

		// Thresholds
		MinConfidence:      0.0, // No minimum confidence
		MaxConfidence:      1.0, // Maximum confidence is 1.0
		AgreementThreshold: 0.6, // 60% agreement threshold
	}
}

// NewWeightedConfidenceScorer creates a new weighted confidence scorer
func NewWeightedConfidenceScorer(logger *log.Logger) *WeightedConfidenceScorer {
	if logger == nil {
		logger = log.Default()
	}

	return &WeightedConfidenceScorer{
		logger: logger,
		config: DefaultWeightedConfidenceConfig(),
	}
}

// NewWeightedConfidenceScorerWithConfig creates a new weighted confidence scorer with custom configuration
func NewWeightedConfidenceScorerWithConfig(logger *log.Logger, config *WeightedConfidenceConfig) *WeightedConfidenceScorer {
	if logger == nil {
		logger = log.Default()
	}

	if config == nil {
		config = DefaultWeightedConfidenceConfig()
	}

	return &WeightedConfidenceScorer{
		logger: logger,
		config: config,
	}
}

// CalculateWeightedConfidence calculates the weighted confidence score across multiple classification methods
func (wcs *WeightedConfidenceScorer) CalculateWeightedConfidence(
	ctx context.Context,
	methodResults []shared.ClassificationMethodResult,
) (*WeightedConfidenceResult, error) {
	if len(methodResults) == 0 {
		return nil, fmt.Errorf("no method results provided")
	}

	startTime := time.Now()
	requestID := wcs.generateRequestID()

	wcs.logger.Printf("🎯 Calculating weighted confidence for %d methods (request: %s)", len(methodResults), requestID)

	// Step 1: Calculate base weighted confidence
	baseConfidence := wcs.calculateBaseWeightedConfidence(methodResults)

	// Step 2: Calculate method agreement factor
	agreementFactor := wcs.calculateAgreementFactor(methodResults)

	// Step 3: Calculate evidence strength factor
	evidenceFactor := wcs.calculateEvidenceFactor(methodResults)

	// Step 4: Calculate data completeness factor
	completenessFactor := wcs.calculateDataCompletenessFactor(methodResults)

	// Step 5: Calculate processing time factor
	timeFactor := wcs.calculateProcessingTimeFactor(methodResults)

	// Step 6: Apply all factors to get final confidence
	finalConfidence := wcs.applyConfidenceFactors(
		baseConfidence,
		agreementFactor,
		evidenceFactor,
		completenessFactor,
		timeFactor,
	)

	// Step 7: Create detailed result
	result := &WeightedConfidenceResult{
		FinalConfidence:     finalConfidence,
		BaseConfidence:      baseConfidence,
		AgreementFactor:     agreementFactor,
		EvidenceFactor:      evidenceFactor,
		CompletenessFactor:  completenessFactor,
		TimeFactor:          timeFactor,
		MethodContributions: wcs.calculateMethodContributions(methodResults),
		ConfidenceBreakdown: wcs.generateConfidenceBreakdown(methodResults),
		ProcessingTime:      time.Since(startTime),
		RequestID:           requestID,
		CreatedAt:           time.Now(),
	}

	wcs.logger.Printf("✅ Weighted confidence calculated: %.3f (request: %s)", finalConfidence, requestID)

	return result, nil
}

// WeightedConfidenceResult represents the result of weighted confidence calculation
type WeightedConfidenceResult struct {
	FinalConfidence     float64                `json:"final_confidence"`
	BaseConfidence      float64                `json:"base_confidence"`
	AgreementFactor     float64                `json:"agreement_factor"`
	EvidenceFactor      float64                `json:"evidence_factor"`
	CompletenessFactor  float64                `json:"completeness_factor"`
	TimeFactor          float64                `json:"time_factor"`
	MethodContributions []MethodContribution   `json:"method_contributions"`
	ConfidenceBreakdown map[string]interface{} `json:"confidence_breakdown"`
	ProcessingTime      time.Duration          `json:"processing_time"`
	RequestID           string                 `json:"request_id"`
	CreatedAt           time.Time              `json:"created_at"`
}

// MethodContribution represents the contribution of a single method to the final confidence
type MethodContribution struct {
	MethodName     string  `json:"method_name"`
	MethodType     string  `json:"method_type"`
	BaseConfidence float64 `json:"base_confidence"`
	Weight         float64 `json:"weight"`
	Contribution   float64 `json:"contribution"`
	Success        bool    `json:"success"`
	Error          string  `json:"error,omitempty"`
}

// calculateBaseWeightedConfidence calculates the base weighted confidence from method results
func (wcs *WeightedConfidenceScorer) calculateBaseWeightedConfidence(
	methodResults []shared.ClassificationMethodResult,
) float64 {
	var totalWeightedConfidence float64
	var totalWeight float64

	for _, method := range methodResults {
		if !method.Success {
			continue // Skip failed methods
		}

		// Get method weight
		weight := wcs.getMethodWeight(method.MethodType)

		// Calculate weighted contribution
		weightedConfidence := method.Confidence * weight
		totalWeightedConfidence += weightedConfidence
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0 // No successful methods
	}

	return totalWeightedConfidence / totalWeight
}

// calculateAgreementFactor calculates the agreement factor based on method consensus
func (wcs *WeightedConfidenceScorer) calculateAgreementFactor(
	methodResults []shared.ClassificationMethodResult,
) float64 {
	// Group methods by their industry classification
	industryGroups := make(map[string][]shared.ClassificationMethodResult)

	for _, method := range methodResults {
		if !method.Success || method.Result == nil {
			continue
		}

		industryName := method.Result.IndustryName
		industryGroups[industryName] = append(industryGroups[industryName], method)
	}

	if len(industryGroups) == 0 {
		return 0.0 // No successful methods
	}

	// Find the largest group (most agreement)
	maxGroupSize := 0
	for _, group := range industryGroups {
		if len(group) > maxGroupSize {
			maxGroupSize = len(group)
		}
	}

	totalMethods := len(methodResults)
	agreementRatio := float64(maxGroupSize) / float64(totalMethods)

	// Apply agreement bonus/penalty
	if agreementRatio >= wcs.config.AgreementThreshold {
		return wcs.config.AgreementBonus * agreementRatio
	} else {
		return -wcs.config.DisagreementPenalty * (1.0 - agreementRatio)
	}
}

// calculateEvidenceFactor calculates the evidence strength factor
func (wcs *WeightedConfidenceScorer) calculateEvidenceFactor(
	methodResults []shared.ClassificationMethodResult,
) float64 {
	var totalEvidenceStrength float64
	var methodCount int

	for _, method := range methodResults {
		if !method.Success {
			continue
		}

		// Calculate evidence strength for this method
		evidenceStrength := float64(len(method.Evidence)) * method.Confidence
		totalEvidenceStrength += evidenceStrength
		methodCount++
	}

	if methodCount == 0 {
		return 0.0
	}

	averageEvidenceStrength := totalEvidenceStrength / float64(methodCount)
	return wcs.config.EvidenceWeight * averageEvidenceStrength
}

// calculateDataCompletenessFactor calculates the data completeness factor
func (wcs *WeightedConfidenceScorer) calculateDataCompletenessFactor(
	methodResults []shared.ClassificationMethodResult,
) float64 {
	successfulMethods := 0
	for _, method := range methodResults {
		if method.Success {
			successfulMethods++
		}
	}

	completenessRatio := float64(successfulMethods) / float64(len(methodResults))
	return wcs.config.DataCompletenessWeight * completenessRatio
}

// calculateProcessingTimeFactor calculates the processing time factor
func (wcs *WeightedConfidenceScorer) calculateProcessingTimeFactor(
	methodResults []shared.ClassificationMethodResult,
) float64 {
	var totalProcessingTime time.Duration
	var methodCount int

	for _, method := range methodResults {
		if !method.Success {
			continue
		}

		totalProcessingTime += method.ProcessingTime
		methodCount++
	}

	if methodCount == 0 {
		return 0.0
	}

	averageProcessingTime := totalProcessingTime / time.Duration(methodCount)

	// Convert to seconds and apply time decay factor
	timeInSeconds := averageProcessingTime.Seconds()

	// Faster processing gets a bonus, slower processing gets a penalty
	// Optimal time is around 1 second
	optimalTime := 1.0
	timeDifference := math.Abs(timeInSeconds - optimalTime)

	// Apply time decay factor
	timeFactor := wcs.config.TimeDecayFactor * (1.0 - timeDifference/optimalTime)

	// Ensure factor is within reasonable bounds
	if timeFactor < -0.1 {
		timeFactor = -0.1
	}
	if timeFactor > 0.1 {
		timeFactor = 0.1
	}

	return timeFactor
}

// applyConfidenceFactors applies all factors to get the final confidence score
func (wcs *WeightedConfidenceScorer) applyConfidenceFactors(
	baseConfidence, agreementFactor, evidenceFactor, completenessFactor, timeFactor float64,
) float64 {
	// Start with base confidence
	finalConfidence := baseConfidence

	// Apply all factors
	finalConfidence += agreementFactor
	finalConfidence += evidenceFactor
	finalConfidence += completenessFactor
	finalConfidence += timeFactor

	// Ensure confidence is within bounds
	if finalConfidence < wcs.config.MinConfidence {
		finalConfidence = wcs.config.MinConfidence
	}
	if finalConfidence > wcs.config.MaxConfidence {
		finalConfidence = wcs.config.MaxConfidence
	}

	return finalConfidence
}

// calculateMethodContributions calculates the contribution of each method to the final confidence
func (wcs *WeightedConfidenceScorer) calculateMethodContributions(
	methodResults []shared.ClassificationMethodResult,
) []MethodContribution {
	var contributions []MethodContribution

	for _, method := range methodResults {
		weight := wcs.getMethodWeight(method.MethodType)
		contribution := method.Confidence * weight

		contrib := MethodContribution{
			MethodName:     method.MethodName,
			MethodType:     method.MethodType,
			BaseConfidence: method.Confidence,
			Weight:         weight,
			Contribution:   contribution,
			Success:        method.Success,
		}

		if !method.Success {
			contrib.Error = method.Error
		}

		contributions = append(contributions, contrib)
	}

	return contributions
}

// generateConfidenceBreakdown generates a detailed breakdown of confidence calculation
func (wcs *WeightedConfidenceScorer) generateConfidenceBreakdown(
	methodResults []shared.ClassificationMethodResult,
) map[string]interface{} {
	breakdown := map[string]interface{}{
		"total_methods":      len(methodResults),
		"successful_methods": wcs.countSuccessfulMethods(methodResults),
		"failed_methods":     len(methodResults) - wcs.countSuccessfulMethods(methodResults),
		"method_weights": map[string]float64{
			"keyword":     wcs.config.KeywordWeight,
			"ml":          wcs.config.MLWeight,
			"description": wcs.config.DescriptionWeight,
		},
		"adjustment_factors": map[string]float64{
			"agreement_bonus":      wcs.config.AgreementBonus,
			"disagreement_penalty": wcs.config.DisagreementPenalty,
			"evidence_weight":      wcs.config.EvidenceWeight,
			"completeness_weight":  wcs.config.DataCompletenessWeight,
			"time_decay_factor":    wcs.config.TimeDecayFactor,
		},
		"thresholds": map[string]float64{
			"min_confidence":      wcs.config.MinConfidence,
			"max_confidence":      wcs.config.MaxConfidence,
			"agreement_threshold": wcs.config.AgreementThreshold,
		},
	}

	return breakdown
}

// getMethodWeight returns the weight for a specific method type
func (wcs *WeightedConfidenceScorer) getMethodWeight(methodType string) float64 {
	switch methodType {
	case "keyword":
		return wcs.config.KeywordWeight
	case "ml":
		return wcs.config.MLWeight
	case "description":
		return wcs.config.DescriptionWeight
	default:
		return 0.1 // Default weight for unknown methods
	}
}

// countSuccessfulMethods counts the number of successful methods
func (wcs *WeightedConfidenceScorer) countSuccessfulMethods(
	methodResults []shared.ClassificationMethodResult,
) int {
	count := 0
	for _, method := range methodResults {
		if method.Success {
			count++
		}
	}
	return count
}

// generateRequestID generates a unique request ID for tracking
func (wcs *WeightedConfidenceScorer) generateRequestID() string {
	return fmt.Sprintf("weighted_conf_%d", time.Now().UnixNano())
}

// UpdateConfig updates the configuration for the weighted confidence scorer
func (wcs *WeightedConfidenceScorer) UpdateConfig(config *WeightedConfidenceConfig) {
	if config != nil {
		wcs.config = config
		wcs.logger.Printf("📊 Updated weighted confidence scorer configuration")
	}
}

// GetConfig returns the current configuration
func (wcs *WeightedConfidenceScorer) GetConfig() *WeightedConfidenceConfig {
	return wcs.config
}
