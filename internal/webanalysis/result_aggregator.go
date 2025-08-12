package webanalysis

import (
	"sort"
	"time"
)

// ResultAggregator handles result aggregation and scoring
type ResultAggregator struct {
	config ResultAggregatorConfig
}

// ResultAggregatorConfig holds configuration for result aggregation
type ResultAggregatorConfig struct {
	MaxIndustries    int
	MinConfidence    float64
	QualityThreshold float64
	EnableValidation bool
}

// NewResultAggregator creates a new result aggregator
func NewResultAggregator() *ResultAggregator {
	config := ResultAggregatorConfig{
		MaxIndustries:    3,
		MinConfidence:    0.5,
		QualityThreshold: 0.7,
		EnableValidation: true,
	}

	return &ResultAggregator{
		config: config,
	}
}

// AggregateResults aggregates results from multiple sources
func (ra *ResultAggregator) AggregateResults(results []*ClassificationResult) (*ClassificationResult, error) {
	if len(results) == 0 {
		return nil, nil
	}

	if len(results) == 1 {
		return results[0], nil
	}

	// Aggregate industries
	aggregatedIndustries := ra.aggregateIndustries(results)

	// Aggregate risk assessments
	aggregatedRiskAssessment := ra.aggregateRiskAssessments(results)

	// Aggregate connection validations
	aggregatedConnectionValidation := ra.aggregateConnectionValidations(results)

	// Calculate overall confidence
	overallConfidence := ra.calculateOverallConfidence(results)

	// Create aggregated result
	aggregatedResult := &ClassificationResult{
		RequestID:            results[0].RequestID,
		BusinessName:         results[0].BusinessName,
		FlowUsed:             results[0].FlowUsed,
		ProcessingTime:       ra.calculateAverageProcessingTime(results),
		Confidence:           overallConfidence,
		Industries:           aggregatedIndustries,
		RiskAssessment:       aggregatedRiskAssessment,
		ConnectionValidation: aggregatedConnectionValidation,
		Errors:               ra.aggregateErrors(results),
		Warnings:             ra.aggregateWarnings(results),
	}

	return aggregatedResult, nil
}

// aggregateIndustries aggregates industry classifications
func (ra *ResultAggregator) aggregateIndustries(results []*ClassificationResult) []IndustryClassification {
	industryMap := make(map[string]*IndustryClassification)

	// Collect all industries from all results
	for _, result := range results {
		for _, industry := range result.Industries {
			if existing, exists := industryMap[industry.Industry]; exists {
				// Merge with existing industry
				existing.Confidence = (existing.Confidence + industry.Confidence) / 2
				existing.Evidence = existing.Evidence + "; " + industry.Evidence
				existing.Keywords = ra.mergeKeywords(existing.Keywords, industry.Keywords)
			} else {
				// Add new industry
				industryCopy := industry
				industryMap[industry.Industry] = &industryCopy
			}
		}
	}

	// Convert map to slice and sort by confidence
	var industries []IndustryClassification
	for _, industry := range industryMap {
		industries = append(industries, *industry)
	}

	// Sort by confidence (descending)
	sort.Slice(industries, func(i, j int) bool {
		return industries[i].Confidence > industries[j].Confidence
	})

	// Limit to max industries
	if len(industries) > ra.config.MaxIndustries {
		industries = industries[:ra.config.MaxIndustries]
	}

	return industries
}

// aggregateRiskAssessments aggregates risk assessments
func (ra *ResultAggregator) aggregateRiskAssessments(results []*ClassificationResult) *RiskAssessment {
	var riskAssessments []*RiskAssessment

	// Collect all risk assessments
	for _, result := range results {
		if result.RiskAssessment != nil {
			riskAssessments = append(riskAssessments, result.RiskAssessment)
		}
	}

	if len(riskAssessments) == 0 {
		return nil
	}

	if len(riskAssessments) == 1 {
		return riskAssessments[0]
	}

	// Aggregate risk factors
	aggregatedRiskFactors := ra.aggregateRiskFactors(riskAssessments)

	// Aggregate risk indicators
	aggregatedRiskIndicators := ra.aggregateRiskIndicators(riskAssessments)

	// Calculate overall risk score
	overallRiskScore := ra.calculateOverallRiskScore(riskAssessments)

	// Determine overall risk level
	overallRisk := ra.determineOverallRisk(overallRiskScore)

	// Aggregate recommendations
	aggregatedRecommendations := ra.aggregateRecommendations(riskAssessments)

	return &RiskAssessment{
		OverallRisk:     overallRisk,
		RiskScore:       overallRiskScore,
		RiskFactors:     aggregatedRiskFactors,
		RiskIndicators:  aggregatedRiskIndicators,
		Recommendations: aggregatedRecommendations,
	}
}

// aggregateConnectionValidations aggregates connection validations
func (ra *ResultAggregator) aggregateConnectionValidations(results []*ClassificationResult) *ConnectionValidation {
	var connectionValidations []*ConnectionValidation

	// Collect all connection validations
	for _, result := range results {
		if result.ConnectionValidation != nil {
			connectionValidations = append(connectionValidations, result.ConnectionValidation)
		}
	}

	if len(connectionValidations) == 0 {
		return nil
	}

	if len(connectionValidations) == 1 {
		return connectionValidations[0]
	}

	// Calculate overall confidence
	overallConfidence := ra.calculateConnectionConfidence(connectionValidations)

	// Determine if connected
	isConnected := overallConfidence >= ra.config.MinConfidence

	// Aggregate validation factors
	aggregatedValidationFactors := ra.aggregateValidationFactors(connectionValidations)

	// Aggregate recommendations
	aggregatedRecommendations := ra.aggregateConnectionRecommendations(connectionValidations)

	// Combine evidence
	combinedEvidence := ra.combineEvidence(connectionValidations)

	return &ConnectionValidation{
		IsConnected:       isConnected,
		Confidence:        overallConfidence,
		Evidence:          combinedEvidence,
		ValidationFactors: aggregatedValidationFactors,
		Recommendations:   aggregatedRecommendations,
	}
}

// aggregateRiskFactors aggregates risk factors
func (ra *ResultAggregator) aggregateRiskFactors(riskAssessments []*RiskAssessment) []RiskFactor {
	riskFactorMap := make(map[string]*RiskFactor)

	for _, assessment := range riskAssessments {
		for _, factor := range assessment.RiskFactors {
			key := factor.Category + ":" + factor.Description
			if existing, exists := riskFactorMap[key]; exists {
				// Merge with existing factor
				existing.Confidence = (existing.Confidence + factor.Confidence) / 2
				existing.Evidence = existing.Evidence + "; " + factor.Evidence
			} else {
				// Add new factor
				factorCopy := factor
				riskFactorMap[key] = &factorCopy
			}
		}
	}

	var riskFactors []RiskFactor
	for _, factor := range riskFactorMap {
		riskFactors = append(riskFactors, *factor)
	}

	return riskFactors
}

// aggregateRiskIndicators aggregates risk indicators
func (ra *ResultAggregator) aggregateRiskIndicators(riskAssessments []*RiskAssessment) []RiskIndicator {
	indicatorMap := make(map[string]*RiskIndicator)

	for _, assessment := range riskAssessments {
		for _, indicator := range assessment.RiskIndicators {
			key := indicator.Type + ":" + indicator.Description
			if existing, exists := indicatorMap[key]; exists {
				// Merge with existing indicator
				existing.Confidence = (existing.Confidence + indicator.Confidence) / 2
			} else {
				// Add new indicator
				indicatorCopy := indicator
				indicatorMap[key] = &indicatorCopy
			}
		}
	}

	var riskIndicators []RiskIndicator
	for _, indicator := range indicatorMap {
		riskIndicators = append(riskIndicators, *indicator)
	}

	return riskIndicators
}

// aggregateValidationFactors aggregates validation factors
func (ra *ResultAggregator) aggregateValidationFactors(connectionValidations []*ConnectionValidation) []ValidationFactor {
	validationFactorMap := make(map[string]*ValidationFactor)

	for _, validation := range connectionValidations {
		for _, factor := range validation.ValidationFactors {
			if existing, exists := validationFactorMap[factor.Factor]; exists {
				// Merge with existing factor
				existing.Confidence = (existing.Confidence + factor.Confidence) / 2
				existing.Match = existing.Match && factor.Match
				existing.Details = existing.Details + "; " + factor.Details
			} else {
				// Add new factor
				factorCopy := factor
				validationFactorMap[factor.Factor] = &factorCopy
			}
		}
	}

	var validationFactors []ValidationFactor
	for _, factor := range validationFactorMap {
		validationFactors = append(validationFactors, *factor)
	}

	return validationFactors
}

// calculateOverallConfidence calculates overall confidence from multiple results
func (ra *ResultAggregator) calculateOverallConfidence(results []*ClassificationResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, result := range results {
		totalConfidence += result.Confidence
	}

	return totalConfidence / float64(len(results))
}

// calculateOverallRiskScore calculates overall risk score from multiple assessments
func (ra *ResultAggregator) calculateOverallRiskScore(riskAssessments []*RiskAssessment) float64 {
	if len(riskAssessments) == 0 {
		return 0.0
	}

	totalRiskScore := 0.0
	for _, assessment := range riskAssessments {
		totalRiskScore += assessment.RiskScore
	}

	return totalRiskScore / float64(len(riskAssessments))
}

// calculateConnectionConfidence calculates overall connection confidence
func (ra *ResultAggregator) calculateConnectionConfidence(connectionValidations []*ConnectionValidation) float64 {
	if len(connectionValidations) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, validation := range connectionValidations {
		totalConfidence += validation.Confidence
	}

	return totalConfidence / float64(len(connectionValidations))
}

// determineOverallRisk determines overall risk level from score
func (ra *ResultAggregator) determineOverallRisk(riskScore float64) string {
	if riskScore >= 0.8 {
		return "Critical"
	} else if riskScore >= 0.6 {
		return "High"
	} else if riskScore >= 0.4 {
		return "Medium"
	} else if riskScore >= 0.2 {
		return "Low"
	} else {
		return "Minimal"
	}
}

// aggregateRecommendations aggregates recommendations from risk assessments
func (ra *ResultAggregator) aggregateRecommendations(riskAssessments []*RiskAssessment) []string {
	recommendationMap := make(map[string]bool)

	for _, assessment := range riskAssessments {
		for _, recommendation := range assessment.Recommendations {
			recommendationMap[recommendation] = true
		}
	}

	var recommendations []string
	for recommendation := range recommendationMap {
		recommendations = append(recommendations, recommendation)
	}

	return recommendations
}

// aggregateConnectionRecommendations aggregates recommendations from connection validations
func (ra *ResultAggregator) aggregateConnectionRecommendations(connectionValidations []*ConnectionValidation) []string {
	recommendationMap := make(map[string]bool)

	for _, validation := range connectionValidations {
		for _, recommendation := range validation.Recommendations {
			recommendationMap[recommendation] = true
		}
	}

	var recommendations []string
	for recommendation := range recommendationMap {
		recommendations = append(recommendations, recommendation)
	}

	return recommendations
}

// combineEvidence combines evidence from multiple sources
func (ra *ResultAggregator) combineEvidence(connectionValidations []*ConnectionValidation) string {
	var evidence []string

	for _, validation := range connectionValidations {
		if validation.Evidence != "" {
			evidence = append(evidence, validation.Evidence)
		}
	}

	if len(evidence) == 0 {
		return "No evidence available"
	}

	// Simple combination - in production, you might want more sophisticated merging
	return evidence[0]
}

// aggregateErrors aggregates errors from multiple results
func (ra *ResultAggregator) aggregateErrors(results []*ClassificationResult) []string {
	errorMap := make(map[string]bool)

	for _, result := range results {
		for _, err := range result.Errors {
			errorMap[err] = true
		}
	}

	var errors []string
	for err := range errorMap {
		errors = append(errors, err)
	}

	return errors
}

// aggregateWarnings aggregates warnings from multiple results
func (ra *ResultAggregator) aggregateWarnings(results []*ClassificationResult) []string {
	warningMap := make(map[string]bool)

	for _, result := range results {
		for _, warning := range result.Warnings {
			warningMap[warning] = true
		}
	}

	var warnings []string
	for warning := range warningMap {
		warnings = append(warnings, warning)
	}

	return warnings
}

// calculateAverageProcessingTime calculates average processing time
func (ra *ResultAggregator) calculateAverageProcessingTime(results []*ClassificationResult) time.Duration {
	if len(results) == 0 {
		return 0
	}

	totalTime := time.Duration(0)
	for _, result := range results {
		totalTime += result.ProcessingTime
	}

	return totalTime / time.Duration(len(results))
}

// mergeKeywords merges keyword lists
func (ra *ResultAggregator) mergeKeywords(keywords1, keywords2 []string) []string {
	keywordMap := make(map[string]bool)

	for _, keyword := range keywords1 {
		keywordMap[keyword] = true
	}
	for _, keyword := range keywords2 {
		keywordMap[keyword] = true
	}

	var mergedKeywords []string
	for keyword := range keywordMap {
		mergedKeywords = append(mergedKeywords, keyword)
	}

	return mergedKeywords
}

// ValidateResult validates a classification result
func (ra *ResultAggregator) ValidateResult(result *ClassificationResult) (bool, []string) {
	var issues []string

	// Check if result is nil
	if result == nil {
		return false, []string{"Result is nil"}
	}

	// Check business name
	if result.BusinessName == "" {
		issues = append(issues, "Business name is missing")
	}

	// Check confidence
	if result.Confidence < ra.config.MinConfidence {
		issues = append(issues, "Confidence below minimum threshold")
	}

	// Check industries
	if len(result.Industries) == 0 {
		issues = append(issues, "No industries classified")
	}

	// Check processing time
	if result.ProcessingTime > time.Minute*5 {
		issues = append(issues, "Processing time exceeds reasonable limit")
	}

	return len(issues) == 0, issues
}
