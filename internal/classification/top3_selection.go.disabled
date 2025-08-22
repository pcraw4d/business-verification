package classification

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// Top3SelectionEngine provides sophisticated top-3 industry selection logic
type Top3SelectionEngine struct {
	logger  *observability.Logger
	metrics *observability.Metrics

	// Configuration
	minConfidenceThreshold float64
	maxConfidenceGap       float64
	diversityPenalty       float64
	consistencyBonus       float64
	industryCoverageWeight float64
	confidenceDecayFactor  float64
}

// NewTop3SelectionEngine creates a new top-3 selection engine
func NewTop3SelectionEngine(logger *observability.Logger, metrics *observability.Metrics) *Top3SelectionEngine {
	return &Top3SelectionEngine{
		logger:  logger,
		metrics: metrics,

		// Configuration
		minConfidenceThreshold: 0.15, // Minimum confidence for inclusion
		maxConfidenceGap:       0.3,  // Maximum gap between consecutive classifications
		diversityPenalty:       0.1,  // Penalty for too similar industries
		consistencyBonus:       0.15, // Bonus for consistent classifications
		industryCoverageWeight: 0.2,  // Weight for industry coverage
		confidenceDecayFactor:  0.8,  // Decay factor for confidence scores
	}
}

// SelectTop3Classifications performs sophisticated top-3 industry selection
func (t *Top3SelectionEngine) SelectTop3Classifications(ctx context.Context, classifications []IndustryClassification) *Top3SelectionResult {
	start := time.Now()

	// Log selection start
	if t.logger != nil {
		t.logger.WithComponent("top3_selection").LogBusinessEvent(ctx, "top3_selection_started", "", map[string]interface{}{
			"total_classifications": len(classifications),
		})
	}

	// Step 1: Apply confidence threshold filtering
	filteredClassifications := t.applyConfidenceThreshold(classifications)

	// Step 2: Calculate enhanced scores
	enhancedClassifications := t.calculateEnhancedScores(filteredClassifications)

	// Step 3: Apply diversity and consistency rules
	diversifiedClassifications := t.applyDiversityRules(enhancedClassifications)

	// Step 4: Select final top-3 with validation
	finalSelection := t.selectFinalTop3(diversifiedClassifications)

	// Step 5: Calculate selection metrics
	metrics := t.calculateSelectionMetrics(finalSelection, classifications)

	// Create result
	result := &Top3SelectionResult{
		AllClassifications: finalSelection,
		SelectionMetrics:   metrics,
		ProcessingTime:     time.Since(start),
		SelectionMethod:    "enhanced_top3_selection",
	}

	// Set primary, secondary, and tertiary industries if available
	if len(finalSelection) > 0 {
		result.PrimaryIndustry = finalSelection[0]
		result.SecondaryIndustry = t.getSecondaryIndustry(finalSelection)
		result.TertiaryIndustry = t.getTertiaryIndustry(finalSelection)
	}

	// Log completion
	if t.logger != nil {
		logData := map[string]interface{}{
			"processing_time_ms":   time.Since(start).Milliseconds(),
			"selection_confidence": metrics.OverallConfidence,
		}

		if len(finalSelection) > 0 {
			logData["primary_industry"] = result.PrimaryIndustry.IndustryCode
			logData["secondary_industry"] = t.getIndustryCode(result.SecondaryIndustry)
			logData["tertiary_industry"] = t.getIndustryCode(result.TertiaryIndustry)
		}

		t.logger.WithComponent("top3_selection").LogBusinessEvent(ctx, "top3_selection_completed", "", logData)
	}

	return result
}

// Top3SelectionResult represents the result of top-3 industry selection
type Top3SelectionResult struct {
	PrimaryIndustry    IndustryClassification   `json:"primary_industry"`
	SecondaryIndustry  *IndustryClassification  `json:"secondary_industry,omitempty"`
	TertiaryIndustry   *IndustryClassification  `json:"tertiary_industry,omitempty"`
	AllClassifications []IndustryClassification `json:"all_classifications"`
	SelectionMetrics   *SelectionMetrics        `json:"selection_metrics"`
	ProcessingTime     time.Duration            `json:"processing_time"`
	SelectionMethod    string                   `json:"selection_method"`
}

// SelectionMetrics provides detailed metrics about the selection process
type SelectionMetrics struct {
	OverallConfidence  float64        `json:"overall_confidence"`
	ConfidenceSpread   float64        `json:"confidence_spread"`
	DiversityScore     float64        `json:"diversity_score"`
	ConsistencyScore   float64        `json:"consistency_score"`
	CoverageScore      float64        `json:"coverage_score"`
	SelectionQuality   float64        `json:"selection_quality"`
	ConfidenceGaps     []float64      `json:"confidence_gaps"`
	MethodDistribution map[string]int `json:"method_distribution"`
	IndustryCategories []string       `json:"industry_categories"`
}

// applyConfidenceThreshold filters classifications by minimum confidence
func (t *Top3SelectionEngine) applyConfidenceThreshold(classifications []IndustryClassification) []IndustryClassification {
	var filtered []IndustryClassification

	for _, classification := range classifications {
		if classification.ConfidenceScore >= t.minConfidenceThreshold {
			filtered = append(filtered, classification)
		}
	}

	return filtered
}

// calculateEnhancedScores calculates enhanced scores for selection
func (t *Top3SelectionEngine) calculateEnhancedScores(classifications []IndustryClassification) []IndustryClassification {
	enhanced := make([]IndustryClassification, len(classifications))

	for i, classification := range classifications {
		enhanced[i] = classification

		// Apply confidence decay based on position
		decayFactor := math.Pow(t.confidenceDecayFactor, float64(i))
		enhanced[i].ConfidenceScore *= decayFactor

		// Add method-specific bonuses
		methodBonus := t.calculateMethodBonus(classification)
		enhanced[i].ConfidenceScore += methodBonus

		// Ensure score stays within bounds
		enhanced[i].ConfidenceScore = math.Max(0.0, math.Min(1.0, enhanced[i].ConfidenceScore))
	}

	// Sort by enhanced confidence score
	sort.Slice(enhanced, func(i, j int) bool {
		return enhanced[i].ConfidenceScore > enhanced[j].ConfidenceScore
	})

	return enhanced
}

// calculateMethodBonus calculates bonus based on classification method
func (t *Top3SelectionEngine) calculateMethodBonus(classification IndustryClassification) float64 {
	switch classification.ClassificationMethod {
	case "keyword_match":
		return 0.05
	case "description_match":
		return 0.03
	case "business_type":
		return 0.02
	case "industry_hint":
		return 0.01
	case "fuzzy_match":
		return 0.0
	default:
		return 0.0
	}
}

// applyDiversityRules applies diversity and consistency rules
func (t *Top3SelectionEngine) applyDiversityRules(classifications []IndustryClassification) []IndustryClassification {
	if len(classifications) <= 3 {
		return classifications
	}

	var diversified []IndustryClassification
	selectedIndustries := make(map[string]bool)

	for _, classification := range classifications {
		// Check if this industry is too similar to already selected ones
		isTooSimilar := false
		for selected := range selectedIndustries {
			if t.areIndustriesTooSimilar(classification.IndustryCode, selected) {
				isTooSimilar = true
				break
			}
		}

		if !isTooSimilar {
			diversified = append(diversified, classification)
			selectedIndustries[classification.IndustryCode] = true

			// Stop when we have enough diverse classifications
			if len(diversified) >= 5 {
				break
			}
		}
	}

	return diversified
}

// areIndustriesTooSimilar checks if two industries are too similar
func (t *Top3SelectionEngine) areIndustriesTooSimilar(code1, code2 string) bool {
	// Check if they're in the same major category
	if len(code1) >= 2 && len(code2) >= 2 {
		if code1[:2] == code2[:2] {
			return true
		}
	}

	// Additional similarity checks could be added here
	// - Sub-category similarity
	// - Keyword overlap
	// - Market segment overlap

	return false
}

// selectFinalTop3 selects the final top-3 classifications
func (t *Top3SelectionEngine) selectFinalTop3(classifications []IndustryClassification) []IndustryClassification {
	if len(classifications) <= 3 {
		return classifications
	}

	// Apply confidence gap validation
	validated := t.validateConfidenceGaps(classifications)

	// Select top-3
	if len(validated) > 3 {
		return validated[:3]
	}

	return validated
}

// validateConfidenceGaps validates confidence gaps between classifications
func (t *Top3SelectionEngine) validateConfidenceGaps(classifications []IndustryClassification) []IndustryClassification {
	if len(classifications) < 2 {
		return classifications
	}

	var validated []IndustryClassification
	validated = append(validated, classifications[0]) // Always include the best

	for i := 1; i < len(classifications); i++ {
		prevConfidence := classifications[i-1].ConfidenceScore
		currentConfidence := classifications[i].ConfidenceScore

		// Check if the gap is acceptable
		gap := prevConfidence - currentConfidence
		if gap <= t.maxConfidenceGap {
			validated = append(validated, classifications[i])
		} else {
			// Gap is too large, stop adding more
			break
		}
	}

	return validated
}

// getSecondaryIndustry gets the secondary industry if available
func (t *Top3SelectionEngine) getSecondaryIndustry(classifications []IndustryClassification) *IndustryClassification {
	if len(classifications) > 1 {
		return &classifications[1]
	}
	return nil
}

// getTertiaryIndustry gets the tertiary industry if available
func (t *Top3SelectionEngine) getTertiaryIndustry(classifications []IndustryClassification) *IndustryClassification {
	if len(classifications) > 2 {
		return &classifications[2]
	}
	return nil
}

// getIndustryCode safely gets industry code from pointer
func (t *Top3SelectionEngine) getIndustryCode(industry *IndustryClassification) string {
	if industry == nil {
		return ""
	}
	return industry.IndustryCode
}

// calculateSelectionMetrics calculates comprehensive selection metrics
func (t *Top3SelectionEngine) calculateSelectionMetrics(selected []IndustryClassification, all []IndustryClassification) *SelectionMetrics {
	metrics := &SelectionMetrics{
		MethodDistribution: make(map[string]int),
		IndustryCategories: make([]string, 0),
	}

	if len(selected) == 0 {
		return metrics
	}

	// Calculate overall confidence
	totalConfidence := 0.0
	for _, classification := range selected {
		totalConfidence += classification.ConfidenceScore
	}
	metrics.OverallConfidence = totalConfidence / float64(len(selected))

	// Calculate confidence spread
	if len(selected) > 1 {
		minConfidence := selected[len(selected)-1].ConfidenceScore
		maxConfidence := selected[0].ConfidenceScore
		metrics.ConfidenceSpread = maxConfidence - minConfidence
	}

	// Calculate confidence gaps
	metrics.ConfidenceGaps = make([]float64, 0)
	for i := 1; i < len(selected); i++ {
		gap := selected[i-1].ConfidenceScore - selected[i].ConfidenceScore
		metrics.ConfidenceGaps = append(metrics.ConfidenceGaps, gap)
	}

	// Calculate diversity score
	metrics.DiversityScore = t.calculateDiversityScore(selected)

	// Calculate consistency score
	metrics.ConsistencyScore = t.calculateConsistencyScore(selected)

	// Calculate coverage score
	metrics.CoverageScore = t.calculateCoverageScore(selected, all)

	// Calculate method distribution
	for _, classification := range selected {
		metrics.MethodDistribution[classification.ClassificationMethod]++
	}

	// Extract industry categories (unique)
	categories := make(map[string]bool)
	for _, classification := range selected {
		if len(classification.IndustryCode) >= 2 {
			category := classification.IndustryCode[:2]
			categories[category] = true
		}
	}

	// Convert map keys to slice
	for category := range categories {
		metrics.IndustryCategories = append(metrics.IndustryCategories, category)
	}

	// Calculate overall selection quality
	metrics.SelectionQuality = (metrics.OverallConfidence * 0.4) +
		(metrics.DiversityScore * 0.3) +
		(metrics.ConsistencyScore * 0.2) +
		(metrics.CoverageScore * 0.1)

	return metrics
}

// calculateDiversityScore calculates diversity score
func (t *Top3SelectionEngine) calculateDiversityScore(classifications []IndustryClassification) float64 {
	if len(classifications) < 2 {
		return 1.0
	}

	// Count unique major categories
	categories := make(map[string]bool)
	for _, classification := range classifications {
		if len(classification.IndustryCode) >= 2 {
			categories[classification.IndustryCode[:2]] = true
		}
	}

	return float64(len(categories)) / float64(len(classifications))
}

// calculateConsistencyScore calculates consistency score
func (t *Top3SelectionEngine) calculateConsistencyScore(classifications []IndustryClassification) float64 {
	if len(classifications) < 2 {
		return 1.0
	}

	// Calculate average confidence consistency
	totalConsistency := 0.0
	pairs := 0

	for i := 0; i < len(classifications); i++ {
		for j := i + 1; j < len(classifications); j++ {
			// Calculate confidence similarity
			diff := math.Abs(classifications[i].ConfidenceScore - classifications[j].ConfidenceScore)
			consistency := 1.0 - diff
			totalConsistency += consistency
			pairs++
		}
	}

	if pairs == 0 {
		return 1.0
	}

	return totalConsistency / float64(pairs)
}

// calculateCoverageScore calculates coverage score
func (t *Top3SelectionEngine) calculateCoverageScore(selected []IndustryClassification, all []IndustryClassification) float64 {
	if len(all) == 0 {
		return 0.0
	}

	// Calculate what percentage of the total confidence is covered by selected classifications
	totalConfidence := 0.0
	selectedConfidence := 0.0

	for _, classification := range all {
		totalConfidence += classification.ConfidenceScore
	}

	for _, classification := range selected {
		selectedConfidence += classification.ConfidenceScore
	}

	if totalConfidence == 0 {
		return 0.0
	}

	return selectedConfidence / totalConfidence
}

// SetConfiguration allows customization of selection parameters
func (t *Top3SelectionEngine) SetConfiguration(minConfidence, maxGap, diversityPenalty, consistencyBonus, coverageWeight, decayFactor float64) {
	t.minConfidenceThreshold = minConfidence
	t.maxConfidenceGap = maxGap
	t.diversityPenalty = diversityPenalty
	t.consistencyBonus = consistencyBonus
	t.industryCoverageWeight = coverageWeight
	t.confidenceDecayFactor = decayFactor
}

// GetConfiguration returns current configuration
func (t *Top3SelectionEngine) GetConfiguration() map[string]float64 {
	return map[string]float64{
		"min_confidence_threshold": t.minConfidenceThreshold,
		"max_confidence_gap":       t.maxConfidenceGap,
		"diversity_penalty":        t.diversityPenalty,
		"consistency_bonus":        t.consistencyBonus,
		"coverage_weight":          t.industryCoverageWeight,
		"decay_factor":             t.confidenceDecayFactor,
	}
}
