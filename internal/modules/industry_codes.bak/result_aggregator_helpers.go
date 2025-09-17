package industry_codes

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// Helper methods for ResultAggregator

// createEmptyResults creates empty results structure
func (ra *ResultAggregator) createEmptyResults(request *AggregationRequest) *AggregatedResults {
	return &AggregatedResults{
		TopThreeByType:    make(map[string][]*AggregatedResult),
		OverallTopResults: []*AggregatedResult{},
		AllResults:        []*AggregatedResult{},
		AggregationMetadata: &AggregationMetadata{
			AggregationTime:     0,
			TotalInputResults:   0,
			AggregatedCount:     0,
			FilteredCount:       0,
			Strategy:            "enhanced_aggregation",
			Criteria:            request,
			QualityDistribution: make(map[ConfidenceLevel]int),
			TypeDistribution:    make(map[string]int),
			ProcessingSteps:     []ProcessingStep{},
		},
	}
}

// deduplicateResults removes duplicate results and merges similar ones
func (ra *ResultAggregator) deduplicateResults(results []*ClassificationResult) []*ClassificationResult {
	resultMap := make(map[string]*ClassificationResult)

	for _, result := range results {
		key := fmt.Sprintf("%s-%s", result.Code.Code, result.Code.Type)

		if existing, exists := resultMap[key]; exists {
			// Merge results for the same code
			existing.Confidence = math.Max(existing.Confidence, result.Confidence)
			existing.MatchedOn = ra.mergeStringSlices(existing.MatchedOn, result.MatchedOn)
			existing.Reasons = ra.mergeStringSlices(existing.Reasons, result.Reasons)
			existing.Weight = math.Max(existing.Weight, result.Weight)

			// Update match type to include multiple strategies
			if existing.MatchType != result.MatchType {
				existing.MatchType = "multi-strategy"
			}
		} else {
			resultMap[key] = result
		}
	}

	// Convert map back to slice
	var dedupedResults []*ClassificationResult
	for _, result := range resultMap {
		dedupedResults = append(dedupedResults, result)
	}

	return dedupedResults
}

// calculateEnhancedScores calculates enhanced scores for aggregation
func (ra *ResultAggregator) calculateEnhancedScores(ctx context.Context, results []*ClassificationResult) ([]*AggregatedResult, error) {
	var enhancedResults []*AggregatedResult

	for _, result := range results {
		aggregated := &AggregatedResult{
			ClassificationResult: result,
			AggregationScore:     ra.calculateAggregationScore(result),
			ConfidenceLevel:      ra.determineConfidenceLevel(result.Confidence),
			QualityIndicators:    ra.analyzeQualityIndicators(result),
			MatchStrength:        ra.determineMatchStrength(result),
			DisplayPriority:      ra.calculateDisplayPriority(result),
			UIHints:              ra.generateUIHints(result),
		}

		// Calculate related codes if available
		if ra.confidenceScorer != nil {
			relatedCodes, err := ra.findRelatedCodes(ctx, result)
			if err == nil {
				aggregated.RelatedCodes = relatedCodes
			}
		}

		enhancedResults = append(enhancedResults, aggregated)
	}

	return enhancedResults, nil
}

// calculateAggregationScore calculates the aggregation score
func (ra *ResultAggregator) calculateAggregationScore(result *ClassificationResult) float64 {
	score := result.Confidence * 0.6 // Base confidence weight

	// Add bonus for match type
	switch result.MatchType {
	case "exact":
		score += 0.2
	case "keyword":
		score += 0.15
	case "fuzzy":
		score += 0.1
	case "multi-strategy":
		score += 0.25
	}

	// Add bonus for multiple reasons
	if len(result.Reasons) > 1 {
		score += 0.05 * float64(len(result.Reasons)-1)
	}

	// Add bonus for multiple matched terms
	if len(result.MatchedOn) > 1 {
		score += 0.05 * float64(len(result.MatchedOn)-1)
	}

	// Apply weight factor
	if result.Weight > 0 {
		score *= (1.0 + result.Weight*0.1)
	}

	// Ensure score doesn't exceed 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// determineConfidenceLevel categorizes confidence level
func (ra *ResultAggregator) determineConfidenceLevel(confidence float64) ConfidenceLevel {
	switch {
	case confidence >= 0.9:
		return ConfidenceLevelVeryHigh
	case confidence >= 0.75:
		return ConfidenceLevelHigh
	case confidence >= 0.5:
		return ConfidenceLevelMedium
	case confidence >= 0.25:
		return ConfidenceLevelLow
	default:
		return ConfidenceLevelVeryLow
	}
}

// analyzeQualityIndicators analyzes quality indicators for a result
func (ra *ResultAggregator) analyzeQualityIndicators(result *ClassificationResult) []string {
	var indicators []string

	// Confidence-based indicators
	if result.Confidence >= 0.9 {
		indicators = append(indicators, "very_high_confidence")
	} else if result.Confidence >= 0.75 {
		indicators = append(indicators, "high_confidence")
	}

	// Match type indicators
	if result.MatchType == "exact" {
		indicators = append(indicators, "exact_match")
	} else if result.MatchType == "multi-strategy" {
		indicators = append(indicators, "multi_strategy_validated")
	}

	// Multiple evidence indicators
	if len(result.Reasons) > 2 {
		indicators = append(indicators, "multiple_evidence_points")
	}

	if len(result.MatchedOn) > 2 {
		indicators = append(indicators, "multiple_match_terms")
	}

	// Code quality indicators
	if result.Code != nil {
		if result.Code.Description != "" && len(result.Code.Description) > 20 {
			indicators = append(indicators, "detailed_description")
		}

		if result.Code.Category != "" {
			indicators = append(indicators, "categorized")
		}
	}

	return indicators
}

// determineMatchStrength determines the strength of the match
func (ra *ResultAggregator) determineMatchStrength(result *ClassificationResult) MatchStrength {
	score := result.Confidence

	// Adjust based on match type
	switch result.MatchType {
	case "exact":
		score += 0.1
	case "multi-strategy":
		score += 0.15
	}

	// Adjust based on evidence
	if len(result.Reasons) > 2 {
		score += 0.05
	}

	if len(result.MatchedOn) > 2 {
		score += 0.05
	}

	// Categorize strength
	switch {
	case score >= 0.85:
		return MatchStrengthExact
	case score >= 0.7:
		return MatchStrengthStrong
	case score >= 0.5:
		return MatchStrengthModerate
	case score >= 0.3:
		return MatchStrengthWeak
	default:
		return MatchStrengthMinimal
	}
}

// calculateDisplayPriority calculates display priority
func (ra *ResultAggregator) calculateDisplayPriority(result *ClassificationResult) int {
	priority := int(result.Confidence * 100)

	// Boost priority for exact matches
	if result.MatchType == "exact" {
		priority += 10
	}

	// Boost priority for multi-strategy matches
	if result.MatchType == "multi-strategy" {
		priority += 15
	}

	// Boost priority for multiple evidence
	priority += len(result.Reasons) * 2
	priority += len(result.MatchedOn) * 2

	return priority
}

// generateUIHints generates UI hints for presentation
func (ra *ResultAggregator) generateUIHints(result *ClassificationResult) map[string]interface{} {
	hints := make(map[string]interface{})

	// Confidence visualization hints
	if result.Confidence >= 0.9 {
		hints["confidence_color"] = "green"
		hints["confidence_icon"] = "check-circle"
	} else if result.Confidence >= 0.75 {
		hints["confidence_color"] = "blue"
		hints["confidence_icon"] = "info-circle"
	} else if result.Confidence >= 0.5 {
		hints["confidence_color"] = "orange"
		hints["confidence_icon"] = "warning"
	} else {
		hints["confidence_color"] = "red"
		hints["confidence_icon"] = "alert-triangle"
	}

	// Match type hints
	hints["match_type_badge"] = result.MatchType

	// Priority hints
	if result.Confidence >= 0.8 {
		hints["priority"] = "high"
		hints["featured"] = true
	} else if result.Confidence >= 0.6 {
		hints["priority"] = "medium"
	} else {
		hints["priority"] = "low"
	}

	// Evidence hints
	hints["evidence_count"] = len(result.Reasons)
	hints["match_terms_count"] = len(result.MatchedOn)

	return hints
}

// findRelatedCodes finds related codes for a result
func (ra *ResultAggregator) findRelatedCodes(ctx context.Context, result *ClassificationResult) ([]*RelatedCode, error) {
	// This would integrate with the metadata manager to find related codes
	// For now, return empty slice
	return []*RelatedCode{}, nil
}

// applyFiltering applies filtering based on request criteria
func (ra *ResultAggregator) applyFiltering(results []*AggregatedResult, request *AggregationRequest) []*AggregatedResult {
	var filtered []*AggregatedResult

	for _, result := range results {
		// Apply minimum confidence filter
		if result.Confidence < request.MinConfidence {
			continue
		}

		// Apply quality filters if needed
		if ra.passesQualityFilters(result) {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// passesQualityFilters checks if result passes quality filters
func (ra *ResultAggregator) passesQualityFilters(result *AggregatedResult) bool {
	// Basic quality checks
	if result.Code == nil {
		return false
	}

	if result.Code.Code == "" {
		return false
	}

	// Require some level of evidence
	if len(result.Reasons) == 0 && len(result.MatchedOn) == 0 {
		return false
	}

	return true
}

// sortResults sorts results based on specified criteria
func (ra *ResultAggregator) sortResults(results []*AggregatedResult, sortBy SortCriteria) []*AggregatedResult {
	sorted := make([]*AggregatedResult, len(results))
	copy(sorted, results)

	switch sortBy {
	case SortByConfidence:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Confidence > sorted[j].Confidence
		})
	case SortByRelevance:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].AggregationScore > sorted[j].AggregationScore
		})
	case SortByQuality:
		sort.Slice(sorted, func(i, j int) bool {
			return len(sorted[i].QualityIndicators) > len(sorted[j].QualityIndicators)
		})
	case SortByAlphabetical:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Code.Description < sorted[j].Code.Description
		})
	case SortByCodeType:
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].Code.Type != sorted[j].Code.Type {
				return sorted[i].Code.Type < sorted[j].Code.Type
			}
			return sorted[i].Confidence > sorted[j].Confidence
		})
	case SortByMatchStrength:
		sort.Slice(sorted, func(i, j int) bool {
			strengthOrder := map[MatchStrength]int{
				MatchStrengthExact:    5,
				MatchStrengthStrong:   4,
				MatchStrengthModerate: 3,
				MatchStrengthWeak:     2,
				MatchStrengthMinimal:  1,
			}
			return strengthOrder[sorted[i].MatchStrength] > strengthOrder[sorted[j].MatchStrength]
		})
	default:
		// Default to confidence sorting
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Confidence > sorted[j].Confidence
		})
	}

	// Update ranks
	for i, result := range sorted {
		result.OverallRank = i + 1
	}

	return sorted
}

// groupAndSelectTopByType groups results by type and selects top results
func (ra *ResultAggregator) groupAndSelectTopByType(results []*AggregatedResult, maxPerType int) map[string][]*AggregatedResult {
	grouped := make(map[string][]*AggregatedResult)

	// Group by type
	for _, result := range results {
		codeType := string(result.Code.Type)
		grouped[codeType] = append(grouped[codeType], result)
	}

	// Sort each group and limit to top results
	for codeType, typeResults := range grouped {
		// Sort by aggregation score within type
		sort.Slice(typeResults, func(i, j int) bool {
			return typeResults[i].AggregationScore > typeResults[j].AggregationScore
		})

		// Update type ranks
		for i, result := range typeResults {
			result.TypeRank = i + 1
		}

		// Limit to max results per type
		if len(typeResults) > maxPerType {
			grouped[codeType] = typeResults[:maxPerType]
		}
	}

	return grouped
}

// getOverallTopResults gets the overall top results
func (ra *ResultAggregator) getOverallTopResults(results []*AggregatedResult, maxResults int) []*AggregatedResult {
	if len(results) <= maxResults {
		return results
	}
	return results[:maxResults]
}

// groupByStrategy groups results by match strategy
func (ra *ResultAggregator) groupByStrategy(results []*AggregatedResult) map[string][]*AggregatedResult {
	grouped := make(map[string][]*AggregatedResult)

	for _, result := range results {
		strategy := result.MatchType
		grouped[strategy] = append(grouped[strategy], result)
	}

	return grouped
}

// countResultsInMap counts total results in a map
func (ra *ResultAggregator) countResultsInMap(resultMap map[string][]*AggregatedResult) int {
	count := 0
	for _, results := range resultMap {
		count += len(results)
	}
	return count
}

// calculateQualityDistribution calculates quality distribution
func (ra *ResultAggregator) calculateQualityDistribution(results []*AggregatedResult) map[ConfidenceLevel]int {
	distribution := make(map[ConfidenceLevel]int)

	for _, result := range results {
		distribution[result.ConfidenceLevel]++
	}

	return distribution
}

// calculateTypeDistribution calculates type distribution
func (ra *ResultAggregator) calculateTypeDistribution(results []*AggregatedResult) map[string]int {
	distribution := make(map[string]int)

	for _, result := range results {
		codeType := string(result.Code.Type)
		distribution[codeType]++
	}

	return distribution
}

// mergeStringSlices merges two string slices removing duplicates
func (ra *ResultAggregator) mergeStringSlices(slice1, slice2 []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range slice1 {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	for _, item := range slice2 {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// calculateAnalytics calculates comprehensive analytics
func (ra *ResultAggregator) calculateAnalytics(results []*AggregatedResult, topByType map[string][]*AggregatedResult) *AggregationAnalytics {
	if len(results) == 0 {
		return &AggregationAnalytics{}
	}

	return &AggregationAnalytics{
		ConfidenceStats:     ra.calculateConfidenceStatistics(results),
		TypeCoverage:        ra.calculateTypeCoverage(topByType),
		IndustryCoverage:    ra.calculateIndustryCoverage(results),
		QualityMetrics:      ra.calculateQualityAnalytics(results),
		DiversityMetrics:    ra.calculateDiversityAnalytics(results),
		RecommendationScore: ra.calculateRecommendationScore(results),
		Certainty:           ra.calculateCertainty(results),
		CrossTypeAnalysis:   ra.calculateCrossTypeAnalysis(topByType),
	}
}

// calculateConfidenceStatistics calculates confidence statistics
func (ra *ResultAggregator) calculateConfidenceStatistics(results []*AggregatedResult) *ConfidenceStatistics {
	if len(results) == 0 {
		return &ConfidenceStatistics{}
	}

	confidences := make([]float64, len(results))
	sum := 0.0
	min := 1.0
	max := 0.0

	for i, result := range results {
		conf := result.Confidence
		confidences[i] = conf
		sum += conf
		if conf < min {
			min = conf
		}
		if conf > max {
			max = conf
		}
	}

	mean := sum / float64(len(results))

	// Calculate median
	sort.Float64s(confidences)
	var median float64
	n := len(confidences)
	if n%2 == 0 {
		median = (confidences[n/2-1] + confidences[n/2]) / 2
	} else {
		median = confidences[n/2]
	}

	// Calculate standard deviation
	sumSquares := 0.0
	for _, conf := range confidences {
		diff := conf - mean
		sumSquares += diff * diff
	}
	stdDev := math.Sqrt(sumSquares / float64(len(confidences)))

	// Calculate quartiles
	q1 := confidences[n/4]
	q3 := confidences[3*n/4]

	return &ConfidenceStatistics{
		Mean:      mean,
		Median:    median,
		Mode:      ra.calculateMode(confidences),
		StdDev:    stdDev,
		Min:       min,
		Max:       max,
		Range:     max - min,
		Quartiles: []float64{q1, median, q3},
	}
}

// calculateMode calculates the mode of confidence values
func (ra *ResultAggregator) calculateMode(values []float64) float64 {
	// Round to 2 decimal places for mode calculation
	rounded := make(map[int]int)
	for _, val := range values {
		key := int(val * 100)
		rounded[key]++
	}

	maxCount := 0
	mode := 0
	for key, count := range rounded {
		if count > maxCount {
			maxCount = count
			mode = key
		}
	}

	return float64(mode) / 100.0
}

// calculateTypeCoverage calculates type coverage
func (ra *ResultAggregator) calculateTypeCoverage(topByType map[string][]*AggregatedResult) map[string]float64 {
	coverage := make(map[string]float64)

	expectedTypes := []string{"mcc", "sic", "naics"}

	for _, codeType := range expectedTypes {
		if results, exists := topByType[codeType]; exists && len(results) > 0 {
			coverage[codeType] = 1.0
		} else {
			coverage[codeType] = 0.0
		}
	}

	return coverage
}

// calculateIndustryCoverage calculates industry coverage
func (ra *ResultAggregator) calculateIndustryCoverage(results []*AggregatedResult) map[string]float64 {
	coverage := make(map[string]float64)

	categories := make(map[string]int)
	total := 0

	for _, result := range results {
		if result.Code.Category != "" {
			categories[result.Code.Category]++
			total++
		}
	}

	for category, count := range categories {
		coverage[category] = float64(count) / float64(total)
	}

	return coverage
}

// calculateQualityAnalytics calculates quality analytics
func (ra *ResultAggregator) calculateQualityAnalytics(results []*AggregatedResult) *QualityAnalytics {
	totalQuality := 0.0
	qualityByType := make(map[string]float64)
	typeCount := make(map[string]int)

	var qualityIndicators, qualityIssues, recommendations []string

	for _, result := range results {
		quality := result.Confidence
		totalQuality += quality

		codeType := string(result.Code.Type)
		qualityByType[codeType] += quality
		typeCount[codeType]++

		// Collect quality indicators and issues
		if len(result.QualityIndicators) > 0 {
			qualityIndicators = append(qualityIndicators, result.QualityIndicators...)
		}

		if quality < 0.5 {
			qualityIssues = append(qualityIssues, fmt.Sprintf("Low confidence for %s (%s)", result.Code.Code, result.Code.Description))
		}
	}

	// Calculate averages by type
	for codeType, total := range qualityByType {
		qualityByType[codeType] = total / float64(typeCount[codeType])
	}

	// Generate recommendations
	if len(qualityIssues) > 0 {
		recommendations = append(recommendations, "Consider reviewing low-confidence classifications")
	}

	overallQuality := totalQuality / float64(len(results))
	if overallQuality < 0.7 {
		recommendations = append(recommendations, "Overall classification confidence is low - consider refining input data")
	}

	return &QualityAnalytics{
		OverallQuality:    overallQuality,
		QualityByType:     qualityByType,
		QualityIndicators: ra.deduplicateStrings(qualityIndicators),
		QualityIssues:     qualityIssues,
		Recommendations:   recommendations,
	}
}

// calculateDiversityAnalytics calculates diversity analytics
func (ra *ResultAggregator) calculateDiversityAnalytics(results []*AggregatedResult) *DiversityAnalytics {
	if len(results) == 0 {
		return &DiversityAnalytics{}
	}

	// Count unique types and categories
	types := make(map[string]bool)
	categories := make(map[string]bool)

	for _, result := range results {
		types[string(result.Code.Type)] = true
		categories[result.Code.Category] = true
	}

	typeDiversity := float64(len(types)) / 3.0 // Normalize by max expected types (3)
	categoryDiversity := float64(len(categories)) / float64(len(results))

	// Calculate industry spread (how evenly distributed across industries)
	industrySpread := ra.calculateIndustrySpread(results)

	// Calculate concentration index (higher = more concentrated)
	concentrationIndex := ra.calculateConcentrationIndex(results)

	diversityScore := (typeDiversity + categoryDiversity + industrySpread) / 3.0

	return &DiversityAnalytics{
		TypeDiversity:      typeDiversity,
		CategoryDiversity:  categoryDiversity,
		IndustrySpread:     industrySpread,
		ConcentrationIndex: concentrationIndex,
		DiversityScore:     diversityScore,
	}
}

// calculateIndustrySpread calculates how evenly results are spread across industries
func (ra *ResultAggregator) calculateIndustrySpread(results []*AggregatedResult) float64 {
	categories := make(map[string]int)

	for _, result := range results {
		if result.Code.Category != "" {
			categories[result.Code.Category]++
		}
	}

	if len(categories) <= 1 {
		return 0.0
	}

	// Calculate entropy-based spread
	total := float64(len(results))
	entropy := 0.0

	for _, count := range categories {
		p := float64(count) / total
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	// Normalize by max possible entropy
	maxEntropy := math.Log2(float64(len(categories)))
	if maxEntropy == 0 {
		return 0.0
	}

	return entropy / maxEntropy
}

// calculateConcentrationIndex calculates how concentrated results are
func (ra *ResultAggregator) calculateConcentrationIndex(results []*AggregatedResult) float64 {
	// Herfindahl-Hirschman Index for concentration
	categories := make(map[string]int)

	for _, result := range results {
		if result.Code.Category != "" {
			categories[result.Code.Category]++
		}
	}

	if len(categories) == 0 {
		return 1.0
	}

	total := float64(len(results))
	hhi := 0.0

	for _, count := range categories {
		share := float64(count) / total
		hhi += share * share
	}

	return hhi
}

// calculateRecommendationScore calculates overall recommendation score
func (ra *ResultAggregator) calculateRecommendationScore(results []*AggregatedResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	// Weight top results more heavily
	weights := []float64{0.5, 0.3, 0.2}
	score := 0.0
	totalWeight := 0.0

	for i, result := range results {
		weight := 1.0
		if i < len(weights) {
			weight = weights[i]
		} else {
			weight = 0.1 / float64(i-len(weights)+1)
		}

		score += result.AggregationScore * weight
		totalWeight += weight

		if i >= 10 { // Don't consider too many results
			break
		}
	}

	if totalWeight == 0 {
		return 0.0
	}

	return score / totalWeight
}

// calculateCertainty calculates overall certainty of classification
func (ra *ResultAggregator) calculateCertainty(results []*AggregatedResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	// High certainty if top result is significantly better than others
	if len(results) == 1 {
		return results[0].Confidence
	}

	topScore := results[0].AggregationScore
	secondScore := results[1].AggregationScore

	gap := topScore - secondScore
	confidence := results[0].Confidence

	// Certainty is based on both absolute confidence and relative gap
	certainty := confidence*0.7 + gap*0.3

	if certainty > 1.0 {
		certainty = 1.0
	}

	return certainty
}

// calculateCrossTypeAnalysis calculates cross-type analysis
func (ra *ResultAggregator) calculateCrossTypeAnalysis(topByType map[string][]*AggregatedResult) *CrossTypeAnalysis {
	// Simple consistency check for now
	consistencyScore := 1.0

	// Check if all types have results
	expectedTypes := []string{"mcc", "sic", "naics"}
	foundTypes := 0

	for _, codeType := range expectedTypes {
		if results, exists := topByType[codeType]; exists && len(results) > 0 {
			foundTypes++
		}
	}

	consistencyScore = float64(foundTypes) / float64(len(expectedTypes))

	return &CrossTypeAnalysis{
		TypeCorrelations:   make(map[string]map[string]float64),
		ConsistencyScore:   consistencyScore,
		ConflictingCodes:   []CodeConflict{},
		RecommendedPrimary: ra.findRecommendedPrimaryType(topByType),
	}
}

// findRecommendedPrimaryType finds the recommended primary code type
func (ra *ResultAggregator) findRecommendedPrimaryType(topByType map[string][]*AggregatedResult) string {
	bestType := ""
	bestScore := 0.0

	for codeType, results := range topByType {
		if len(results) > 0 {
			score := results[0].AggregationScore
			if score > bestScore {
				bestScore = score
				bestType = codeType
			}
		}
	}

	return bestType
}

// deduplicateStrings removes duplicates from string slice
func (ra *ResultAggregator) deduplicateStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// createPresentationData creates presentation data based on format
func (ra *ResultAggregator) createPresentationData(results []*AggregatedResult, topByType map[string][]*AggregatedResult, format PresentationFormat, analytics *AggregationAnalytics) *PresentationData {
	data := &PresentationData{
		Format: format,
		Title:  "Industry Code Classification Results",
	}

	// Generate summary and key findings
	data.Summary = ra.generateSummary(results, topByType)
	data.KeyFindings = ra.generateKeyFindings(results, analytics)
	data.Recommendations = ra.generateRecommendations(results, analytics)

	// Create format-specific data
	switch format {
	case PresentationDetailed:
		data.DetailedView = ra.createDetailedPresentation(results, analytics)
	case PresentationSummary:
		data.SummaryView = ra.createSummaryPresentation(results, topByType)
	case PresentationCompact:
		data.CompactView = ra.createCompactPresentation(results)
	case PresentationExport:
		data.ExportData = ra.createExportPresentation(results)
	case PresentationDashboard:
		data.DashboardData = ra.createDashboardPresentation(results, topByType, analytics)
	case PresentationAPI:
		data.APIResponse = ra.createAPIPresentation(results, topByType)
	}

	return data
}

// generateSummary generates a summary of results
func (ra *ResultAggregator) generateSummary(results []*AggregatedResult, topByType map[string][]*AggregatedResult) string {
	if len(results) == 0 {
		return "No classification results found."
	}

	topResult := results[0]
	typeCount := len(topByType)

	summary := fmt.Sprintf("Found %d potential industry classifications across %d code types. ",
		len(results), typeCount)

	summary += fmt.Sprintf("Top match: %s (%s) with %.1f%% confidence.",
		topResult.Code.Description, topResult.Code.Code, topResult.Confidence*100)

	return summary
}

// generateKeyFindings generates key findings
func (ra *ResultAggregator) generateKeyFindings(results []*AggregatedResult, analytics *AggregationAnalytics) []string {
	var findings []string

	if len(results) == 0 {
		return findings
	}

	// Confidence findings
	topResult := results[0]
	if topResult.Confidence >= 0.9 {
		findings = append(findings, fmt.Sprintf("High confidence match found: %s", topResult.Code.Description))
	} else if topResult.Confidence < 0.5 {
		findings = append(findings, "All matches have relatively low confidence - may need manual review")
	}

	// Consistency findings
	if analytics != nil && analytics.CrossTypeAnalysis != nil {
		if analytics.CrossTypeAnalysis.ConsistencyScore >= 0.8 {
			findings = append(findings, "Classifications are consistent across different code types")
		} else {
			findings = append(findings, "Some inconsistency found between different code types")
		}
	}

	// Quality findings
	exactMatches := 0
	for _, result := range results {
		if result.MatchStrength == MatchStrengthExact {
			exactMatches++
		}
	}

	if exactMatches > 0 {
		findings = append(findings, fmt.Sprintf("Found %d exact match(es)", exactMatches))
	}

	return findings
}

// generateRecommendations generates recommendations
func (ra *ResultAggregator) generateRecommendations(results []*AggregatedResult, analytics *AggregationAnalytics) []string {
	var recommendations []string

	if len(results) == 0 {
		recommendations = append(recommendations, "Consider refining business description for better classification")
		return recommendations
	}

	topResult := results[0]

	// Confidence-based recommendations
	if topResult.Confidence >= 0.9 {
		recommendations = append(recommendations, "Top classification has high confidence - recommended for use")
	} else if topResult.Confidence >= 0.7 {
		recommendations = append(recommendations, "Top classification has good confidence - consider for use")
	} else {
		recommendations = append(recommendations, "Consider manual review due to lower confidence scores")
	}

	// Multi-code recommendations
	if len(results) > 1 && results[0].Confidence-results[1].Confidence < 0.1 {
		recommendations = append(recommendations, "Multiple similar matches found - consider reviewing top options")
	}

	// Quality recommendations
	if analytics != nil && analytics.QualityMetrics != nil {
		recommendations = append(recommendations, analytics.QualityMetrics.Recommendations...)
	}

	return recommendations
}

// Helper functions for different presentation formats...
func (ra *ResultAggregator) createDetailedPresentation(results []*AggregatedResult, analytics *AggregationAnalytics) *DetailedPresentation {
	return &DetailedPresentation{
		FullResults:           results,
		DetailedAnalytics:     analytics,
		MethodologyNotes:      ra.getMethodologyNotes(),
		ConfidenceExplanation: ra.getConfidenceExplanation(),
	}
}

func (ra *ResultAggregator) createSummaryPresentation(results []*AggregatedResult, topByType map[string][]*AggregatedResult) *SummaryPresentation {
	var topThree []*AggregatedResult
	count := 0
	for _, result := range results {
		if count < 3 {
			topThree = append(topThree, result)
			count++
		}
	}

	return &SummaryPresentation{
		TopThree:          topThree,
		KeyMetrics:        ra.calculateKeyMetrics(results),
		QuickSummary:      ra.generateQuickSummary(results),
		RecommendedAction: ra.getRecommendedAction(results),
	}
}

func (ra *ResultAggregator) createCompactPresentation(results []*AggregatedResult) *CompactPresentation {
	var bestMatch *AggregatedResult
	var alternatives []*AggregatedResult

	if len(results) > 0 {
		bestMatch = results[0]
		if len(results) > 1 {
			alternatives = results[1:]
			if len(alternatives) > 2 {
				alternatives = alternatives[:2]
			}
		}
	}

	return &CompactPresentation{
		BestMatch:           bestMatch,
		AlternativeMatches:  alternatives,
		ConfidenceIndicator: ra.getConfidenceIndicator(bestMatch),
	}
}

func (ra *ResultAggregator) createExportPresentation(results []*AggregatedResult) *ExportPresentation {
	headers := []string{"Code", "Type", "Description", "Confidence", "Match Type", "Reasons"}
	var csvData [][]string
	csvData = append(csvData, headers)

	for _, result := range results {
		row := []string{
			result.Code.Code,
			string(result.Code.Type),
			result.Code.Description,
			fmt.Sprintf("%.3f", result.Confidence),
			result.MatchType,
			strings.Join(result.Reasons, "; "),
		}
		csvData = append(csvData, row)
	}

	return &ExportPresentation{
		CSVData:        csvData,
		Headers:        headers,
		StructuredData: ra.createStructuredExportData(results),
		ExportMetadata: ra.createExportMetadata(results),
	}
}

func (ra *ResultAggregator) createDashboardPresentation(results []*AggregatedResult, topByType map[string][]*AggregatedResult, analytics *AggregationAnalytics) *DashboardPresentation {
	return &DashboardPresentation{
		Widgets:        ra.createDashboardWidgets(results, topByType),
		Charts:         ra.createChartData(results, analytics),
		KPIs:           ra.calculateKPIs(results, analytics),
		AlertsWarnings: ra.generateAlertsWarnings(results, analytics),
	}
}

func (ra *ResultAggregator) createAPIPresentation(results []*AggregatedResult, topByType map[string][]*AggregatedResult) *APIPresentation {
	return &APIPresentation{
		Status:   "success",
		Data:     topByType,
		Metadata: ra.createAPIMetadata(results),
		Links:    ra.createAPILinks(),
	}
}

// Additional helper methods for presentation...
func (ra *ResultAggregator) getMethodologyNotes() []string {
	return []string{
		"Classification based on business name and description analysis",
		"Confidence scores calculated using multiple factors including exact matches, keyword matches, and contextual analysis",
		"Results ranked by aggregation score combining confidence, relevance, and quality indicators",
		"Top 3 results selected per code type (MCC, SIC, NAICS) to provide comprehensive coverage",
	}
}

func (ra *ResultAggregator) getConfidenceExplanation() string {
	return "Confidence scores range from 0.0 to 1.0, where 1.0 indicates perfect confidence. " +
		"Scores above 0.9 are considered very high confidence, 0.75-0.9 high confidence, " +
		"0.5-0.75 medium confidence, and below 0.5 low confidence."
}

func (ra *ResultAggregator) calculateKeyMetrics(results []*AggregatedResult) map[string]float64 {
	metrics := make(map[string]float64)

	if len(results) == 0 {
		return metrics
	}

	// Calculate average confidence
	totalConfidence := 0.0
	for _, result := range results {
		totalConfidence += result.Confidence
	}
	metrics["average_confidence"] = totalConfidence / float64(len(results))

	// Best confidence
	metrics["best_confidence"] = results[0].Confidence

	// Total results
	metrics["total_results"] = float64(len(results))

	// High confidence count (>= 0.75)
	highConfCount := 0.0
	for _, result := range results {
		if result.Confidence >= 0.75 {
			highConfCount++
		}
	}
	metrics["high_confidence_count"] = highConfCount

	return metrics
}

func (ra *ResultAggregator) generateQuickSummary(results []*AggregatedResult) string {
	if len(results) == 0 {
		return "No results found"
	}

	topResult := results[0]
	return fmt.Sprintf("Best match: %s (%.1f%% confidence)",
		topResult.Code.Description, topResult.Confidence*100)
}

func (ra *ResultAggregator) getRecommendedAction(results []*AggregatedResult) string {
	if len(results) == 0 {
		return "Refine search criteria"
	}

	topResult := results[0]
	if topResult.Confidence >= 0.8 {
		return "Use top classification result"
	} else if topResult.Confidence >= 0.6 {
		return "Review top results before selecting"
	} else {
		return "Manual review recommended"
	}
}

func (ra *ResultAggregator) getConfidenceIndicator(result *AggregatedResult) string {
	if result == nil {
		return "none"
	}

	switch result.ConfidenceLevel {
	case ConfidenceLevelVeryHigh:
		return "🟢 Very High"
	case ConfidenceLevelHigh:
		return "🔵 High"
	case ConfidenceLevelMedium:
		return "🟡 Medium"
	case ConfidenceLevelLow:
		return "🟠 Low"
	default:
		return "🔴 Very Low"
	}
}

func (ra *ResultAggregator) createStructuredExportData(results []*AggregatedResult) map[string]interface{} {
	data := make(map[string]interface{})

	data["results"] = results
	data["export_timestamp"] = time.Now().Format(time.RFC3339)
	data["total_count"] = len(results)

	return data
}

func (ra *ResultAggregator) createExportMetadata(results []*AggregatedResult) map[string]interface{} {
	metadata := make(map[string]interface{})

	metadata["generated_at"] = time.Now().Format(time.RFC3339)
	metadata["version"] = "1.0"
	metadata["format"] = "industry_classification_export"
	metadata["total_results"] = len(results)

	return metadata
}

func (ra *ResultAggregator) createDashboardWidgets(results []*AggregatedResult, topByType map[string][]*AggregatedResult) []DashboardWidget {
	var widgets []DashboardWidget

	// Confidence distribution widget
	widgets = append(widgets, DashboardWidget{
		Type:  "chart",
		Title: "Confidence Distribution",
		Data:  ra.createConfidenceDistributionData(results),
		Size:  "medium",
	})

	// Top results widget
	widgets = append(widgets, DashboardWidget{
		Type:  "table",
		Title: "Top Classifications",
		Data:  ra.createTopResultsData(results),
		Size:  "large",
	})

	// Type coverage widget
	widgets = append(widgets, DashboardWidget{
		Type:  "pie",
		Title: "Code Type Coverage",
		Data:  ra.createTypeCoverageData(topByType),
		Size:  "small",
	})

	return widgets
}

func (ra *ResultAggregator) createChartData(results []*AggregatedResult, analytics *AggregationAnalytics) []ChartData {
	var charts []ChartData

	// Confidence chart
	if analytics != nil && analytics.ConfidenceStats != nil {
		charts = append(charts, ChartData{
			Type:   "bar",
			Title:  "Confidence Statistics",
			Labels: []string{"Min", "Q1", "Median", "Q3", "Max"},
			Datasets: []ChartDataset{
				{
					Label: "Confidence",
					Data:  []float64{analytics.ConfidenceStats.Min, analytics.ConfidenceStats.Quartiles[0], analytics.ConfidenceStats.Median, analytics.ConfidenceStats.Quartiles[2], analytics.ConfidenceStats.Max},
					Color: "#3498db",
				},
			},
		})
	}

	return charts
}

func (ra *ResultAggregator) calculateKPIs(results []*AggregatedResult, analytics *AggregationAnalytics) map[string]float64 {
	kpis := make(map[string]float64)

	if len(results) > 0 {
		kpis["best_confidence"] = results[0].Confidence
		kpis["total_results"] = float64(len(results))
	}

	if analytics != nil {
		if analytics.ConfidenceStats != nil {
			kpis["average_confidence"] = analytics.ConfidenceStats.Mean
		}
		if analytics.DiversityMetrics != nil {
			kpis["diversity_score"] = analytics.DiversityMetrics.DiversityScore
		}
		kpis["recommendation_score"] = analytics.RecommendationScore
		kpis["certainty"] = analytics.Certainty
	}

	return kpis
}

func (ra *ResultAggregator) generateAlertsWarnings(results []*AggregatedResult, analytics *AggregationAnalytics) []string {
	var alerts []string

	if len(results) == 0 {
		alerts = append(alerts, "No classification results found")
		return alerts
	}

	if results[0].Confidence < 0.5 {
		alerts = append(alerts, "Low confidence in top result - manual review recommended")
	}

	if analytics != nil && analytics.Certainty < 0.6 {
		alerts = append(alerts, "Low certainty in classification - consider additional data")
	}

	return alerts
}

func (ra *ResultAggregator) createAPIMetadata(results []*AggregatedResult) map[string]interface{} {
	metadata := make(map[string]interface{})

	metadata["timestamp"] = time.Now().Format(time.RFC3339)
	metadata["total_results"] = len(results)
	metadata["version"] = "v1"

	return metadata
}

func (ra *ResultAggregator) createAPILinks() map[string]string {
	return map[string]string{
		"self": "/api/v1/classify",
		"docs": "/api/v1/docs",
	}
}

// Helper methods for dashboard data creation
func (ra *ResultAggregator) createConfidenceDistributionData(results []*AggregatedResult) interface{} {
	distribution := make(map[string]int)

	for _, result := range results {
		level := string(result.ConfidenceLevel)
		distribution[level]++
	}

	return distribution
}

func (ra *ResultAggregator) createTopResultsData(results []*AggregatedResult) interface{} {
	var topResults []map[string]interface{}

	limit := 5
	if len(results) < limit {
		limit = len(results)
	}

	for i := 0; i < limit; i++ {
		result := results[i]
		topResults = append(topResults, map[string]interface{}{
			"rank":        i + 1,
			"code":        result.Code.Code,
			"type":        result.Code.Type,
			"description": result.Code.Description,
			"confidence":  result.Confidence,
			"match_type":  result.MatchType,
		})
	}

	return topResults
}

func (ra *ResultAggregator) createTypeCoverageData(topByType map[string][]*AggregatedResult) interface{} {
	coverage := make(map[string]int)

	for codeType, results := range topByType {
		coverage[codeType] = len(results)
	}

	return coverage
}
