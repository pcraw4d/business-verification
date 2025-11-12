package feedback

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"go.uber.org/zap"
)

// MisclassificationDetector detects and analyzes model-specific misclassifications
type MisclassificationDetector struct {
	config *MLAnalysisConfig
	logger *zap.Logger
}

// NewMisclassificationDetector creates a new misclassification detector
func NewMisclassificationDetector(config *MLAnalysisConfig, logger *zap.Logger) *MisclassificationDetector {
	return &MisclassificationDetector{
		config: config,
		logger: logger,
	}
}

// DetectModelMisclassifications detects misclassifications specific to ML models
// TODO: This function expects ClassificationClassificationUserFeedback but receives UserFeedback
// For now, return empty results as UserFeedback doesn't have the required fields
func (md *MisclassificationDetector) DetectModelMisclassifications(ctx context.Context, feedback []*UserFeedback) ([]*ModelMisclassification, error) {
	// Stub implementation - UserFeedback doesn't have ClassificationMethod, FeedbackText, etc.
	// This needs to be refactored to use ClassificationClassificationUserFeedback
	_ = ctx
	_ = feedback
	return []*ModelMisclassification{}, nil
}

// analyzeModelMisclassification analyzes misclassifications for a specific model
func (md *MisclassificationDetector) analyzeModelMisclassification(modelType string, feedback []*UserFeedback) *ModelMisclassification {
	// Calculate frequency and confidence
	frequency := len(feedback)
	confidence := md.calculateMisclassificationConfidence(feedback)

	if confidence < md.config.ConfidenceThreshold {
		return nil
	}

	// Analyze affected industries
	affectedIndustries := md.analyzeAffectedIndustries(feedback)

	// Analyze common inputs
	commonInputs := md.analyzeCommonInputs(feedback)

	// Analyze root causes
	rootCauses := md.analyzeRootCauses(feedback)

	// Generate recommendations
	recommendations := md.generateMisclassificationRecommendations(modelType, feedback, rootCauses)

	return &ModelMisclassification{
		ModelID:               fmt.Sprintf("model_%s", modelType),
		ModelType:             modelType,
		MisclassificationType: "model_specific",
		Frequency:             frequency,
		Confidence:            confidence,
		AffectedIndustries:    affectedIndustries,
		CommonInputs:          commonInputs,
		RootCauses:            rootCauses,
		Recommendations:       recommendations,
	}
}

// analyzeMisclassificationType analyzes a specific type of misclassification
func (md *MisclassificationDetector) analyzeMisclassificationType(misclassificationType string, feedback []*UserFeedback) *ModelMisclassification {
	// Calculate frequency and confidence
	frequency := len(feedback)
	confidence := md.calculateMisclassificationConfidence(feedback)

	if confidence < md.config.ConfidenceThreshold {
		return nil
	}

	// Determine primary model type from feedback
	primaryModelType := md.determinePrimaryModelType(feedback)

	// Analyze affected industries
	affectedIndustries := md.analyzeAffectedIndustries(feedback)

	// Analyze common inputs
	commonInputs := md.analyzeCommonInputs(feedback)

	// Analyze root causes
	rootCauses := md.analyzeRootCauses(feedback)

	// Generate recommendations
	recommendations := md.generateMisclassificationRecommendations(primaryModelType, feedback, rootCauses)

	return &ModelMisclassification{
		ModelID:               fmt.Sprintf("misclassification_%s", misclassificationType),
		ModelType:             primaryModelType,
		MisclassificationType: misclassificationType,
		Frequency:             frequency,
		Confidence:            confidence,
		AffectedIndustries:    affectedIndustries,
		CommonInputs:          commonInputs,
		RootCauses:            rootCauses,
		Recommendations:       recommendations,
	}
}

// determineModelType determines the model type from feedback
// TODO: This function expects ClassificationClassificationUserFeedback but receives UserFeedback
func (md *MisclassificationDetector) determineModelType(feedback *UserFeedback) string {
	// Stub - UserFeedback doesn't have ClassificationMethod or FeedbackText
	return "ensemble_model"
}

// determineMisclassificationType determines the type of misclassification
// TODO: This function expects ClassificationClassificationUserFeedback but receives UserFeedback
func (md *MisclassificationDetector) determineMisclassificationType(feedback *UserFeedback) string {
	// Stub - UserFeedback doesn't have FeedbackText or FeedbackType
	return "general_misclassification"
}

// calculateMisclassificationConfidence calculates confidence in the misclassification detection
func (md *MisclassificationDetector) calculateMisclassificationConfidence(feedback []*UserFeedback) float64 {
	if len(feedback) == 0 {
		return 0.0
	}

	var totalConfidence float64
	var confidenceCount int

	for _, fb := range feedback {
		// Stub - UserFeedback doesn't have ConfidenceScore
		// Calculate confidence based on feedback type and content
		confidence := md.calculateFeedbackConfidence(fb)
		totalConfidence += confidence
		confidenceCount++
	}

	if confidenceCount == 0 {
		return 0.5 // Default confidence
	}

	return totalConfidence / float64(confidenceCount)
}

// calculateFeedbackConfidence calculates confidence based on feedback content
// TODO: This function expects ClassificationClassificationUserFeedback but receives UserFeedback
func (md *MisclassificationDetector) calculateFeedbackConfidence(feedback *UserFeedback) float64 {
	// Stub - UserFeedback doesn't have FeedbackType or FeedbackText
	return 0.5 // Default confidence
}

// analyzeAffectedIndustries analyzes which industries are most affected
func (md *MisclassificationDetector) analyzeAffectedIndustries(feedback []*UserFeedback) []string {
	industryCounts := make(map[string]int)

	// Stub - UserFeedback doesn't have BusinessName or FeedbackText
	// All industry extraction functions need ClassificationClassificationUserFeedback
	// For now, return empty list
	_ = feedback

	// Sort industries by frequency
	var industries []string
	for industry := range industryCounts {
		industries = append(industries, industry)
	}

	sort.Slice(industries, func(i, j int) bool {
		return industryCounts[industries[i]] > industryCounts[industries[j]]
	})

	// Return top 5 industries
	if len(industries) > 5 {
		industries = industries[:5]
	}

	return industries
}

// extractIndustriesFromFeedback extracts industry information from feedback
// TODO: This function expects ClassificationClassificationUserFeedback but receives UserFeedback
func (md *MisclassificationDetector) extractIndustriesFromFeedback(feedback *UserFeedback) []string {
	// Stub - UserFeedback doesn't have FeedbackText or BusinessName
	return []string{}
	// Original implementation:
	// var industries []string
	// text := strings.ToLower(feedback.FeedbackText)

	// Common industry keywords
	industryKeywords := map[string]string{
		"restaurant":    "restaurant",
		"food":          "restaurant",
		"dining":        "restaurant",
		"tech":          "technology",
		"software":      "technology",
		"it":            "technology",
		"healthcare":    "healthcare",
		"medical":       "healthcare",
		"hospital":      "healthcare",
		"legal":         "legal_services",
		"law":           "legal_services",
		"attorney":      "legal_services",
		"retail":        "retail",
		"store":         "retail",
		"shop":          "retail",
		"finance":       "financial_services",
		"bank":          "financial_services",
		"insurance":     "financial_services",
		"education":     "education",
		"school":        "education",
		"university":    "education",
		"manufacturing": "manufacturing",
		"production":    "manufacturing",
		"construction":  "construction",
		"building":      "construction",
		"real estate":   "real_estate",
		"property":      "real_estate",
	}

	// Check for industry keywords
	// Stub - text and industries are undefined because UserFeedback doesn't have FeedbackText
	// All code below is unreachable but kept for reference
	_ = industryKeywords
	return []string{}
}

// analyzeCommonInputs analyzes common input patterns in misclassifications
func (md *MisclassificationDetector) analyzeCommonInputs(feedback []*UserFeedback) []string {
	var commonInputs []string

	// Analyze business name patterns
	namePatterns := md.analyzeBusinessNamePatterns(feedback)
	commonInputs = append(commonInputs, namePatterns...)

	// Analyze description patterns
	descriptionPatterns := md.analyzeDescriptionPatterns(feedback)
	commonInputs = append(commonInputs, descriptionPatterns...)

	// Analyze website patterns
	websitePatterns := md.analyzeWebsitePatterns(feedback)
	commonInputs = append(commonInputs, websitePatterns...)

	return commonInputs
}

// analyzeBusinessNamePatterns analyzes patterns in business names
// Stub: UserFeedback doesn't have BusinessName field - needs refactoring
func (md *MisclassificationDetector) analyzeBusinessNamePatterns(feedback []*UserFeedback) []string {
	// TODO: Refactor to use ClassificationClassificationUserFeedback which has BusinessName
	return []string{}
}

// analyzeDescriptionPatterns analyzes patterns in descriptions
func (md *MisclassificationDetector) analyzeDescriptionPatterns(feedback []*UserFeedback) []string {
	var patterns []string
	descriptionLengths := make(map[string]int)
	hasWebsite := 0
	hasEmail := 0

	for _, fb := range feedback {
		// Stub - UserFeedback doesn't have FeedbackText
		// Check if feedback mentions description issues
		text := "" // strings.ToLower(fb.FeedbackText)
		if strings.Contains(text, "description") {
			// Analyze description characteristics from metadata if available
			if metadata, ok := fb.Metadata["description_length"].(int); ok {
				if metadata < 50 {
					descriptionLengths["short_descriptions"]++
				} else if metadata > 500 {
					descriptionLengths["long_descriptions"]++
				} else {
					descriptionLengths["medium_descriptions"]++
				}
			}

			// Check for website/email mentions
			if strings.Contains(text, "website") {
				hasWebsite++
			}
			if strings.Contains(text, "email") {
				hasEmail++
			}
		}
	}

	// Add patterns that occur frequently
	for pattern, count := range descriptionLengths {
		if count > len(feedback)/4 {
			patterns = append(patterns, fmt.Sprintf("description_%s", pattern))
		}
	}

	if hasWebsite > len(feedback)/3 {
		patterns = append(patterns, "description_website_issues")
	}
	if hasEmail > len(feedback)/3 {
		patterns = append(patterns, "description_email_issues")
	}

	return patterns
}

// analyzeWebsitePatterns analyzes patterns in website-related issues
func (md *MisclassificationDetector) analyzeWebsitePatterns(feedback []*UserFeedback) []string {
	var patterns []string
	websiteIssues := make(map[string]int)

	for range feedback {
		// Stub - UserFeedback doesn't have FeedbackText
		text := "" // strings.ToLower(fb.FeedbackText)
		if strings.Contains(text, "website") || strings.Contains(text, "domain") {
			if strings.Contains(text, "invalid") || strings.Contains(text, "error") {
				websiteIssues["invalid_websites"]++
			}
			if strings.Contains(text, "unreachable") || strings.Contains(text, "down") {
				websiteIssues["unreachable_websites"]++
			}
			if strings.Contains(text, "ssl") || strings.Contains(text, "certificate") {
				websiteIssues["ssl_issues"]++
			}
			if strings.Contains(text, "verification") || strings.Contains(text, "trust") {
				websiteIssues["verification_issues"]++
			}
		}
	}

	// Add patterns that occur frequently
	for pattern, count := range websiteIssues {
		if count > len(feedback)/5 {
			patterns = append(patterns, fmt.Sprintf("website_%s", pattern))
		}
	}

	return patterns
}

// analyzeRootCauses analyzes root causes of misclassifications
func (md *MisclassificationDetector) analyzeRootCauses(feedback []*UserFeedback) []string {
	var rootCauses []string
	causeCounts := make(map[string]int)

	for _, fb := range feedback {
		causes := md.identifyRootCauses(fb)
		for _, cause := range causes {
			causeCounts[cause]++
		}
	}

	// Sort causes by frequency
	var causes []string
	for cause := range causeCounts {
		causes = append(causes, cause)
	}

	sort.Slice(causes, func(i, j int) bool {
		return causeCounts[causes[i]] > causeCounts[causes[j]]
	})

	// Return top causes that occur in at least 20% of feedback
	threshold := len(feedback) / 5
	for _, cause := range causes {
		if causeCounts[cause] >= threshold {
			rootCauses = append(rootCauses, cause)
		}
	}

	return rootCauses
}

// identifyRootCauses identifies root causes from individual feedback
// Stub: UserFeedback doesn't have FeedbackText field - needs refactoring
func (md *MisclassificationDetector) identifyRootCauses(feedback *UserFeedback) []string {
	var causes []string
	// Stub - UserFeedback doesn't have FeedbackText
	text := "" // strings.ToLower(feedback.FeedbackText)

	// Data quality issues
	if strings.Contains(text, "missing") || strings.Contains(text, "incomplete") {
		causes = append(causes, "insufficient_data")
	}
	if strings.Contains(text, "unclear") || strings.Contains(text, "ambiguous") {
		causes = append(causes, "ambiguous_input")
	}
	if strings.Contains(text, "outdated") || strings.Contains(text, "old") {
		causes = append(causes, "outdated_data")
	}

	// Model issues
	if strings.Contains(text, "model") || strings.Contains(text, "algorithm") {
		causes = append(causes, "model_limitation")
	}
	if strings.Contains(text, "training") || strings.Contains(text, "learn") {
		causes = append(causes, "insufficient_training")
	}
	if strings.Contains(text, "bias") || strings.Contains(text, "prejudice") {
		causes = append(causes, "model_bias")
	}

	// Keyword matching issues
	if strings.Contains(text, "keyword") || strings.Contains(text, "match") {
		causes = append(causes, "keyword_matching_issue")
	}
	if strings.Contains(text, "synonym") || strings.Contains(text, "similar") {
		causes = append(causes, "synonym_handling")
	}

	// Security issues
	if strings.Contains(text, "security") || strings.Contains(text, "trust") {
		causes = append(causes, "security_validation_issue")
	}
	if strings.Contains(text, "verification") || strings.Contains(text, "validate") {
		causes = append(causes, "verification_failure")
	}

	// External service issues
	if strings.Contains(text, "api") || strings.Contains(text, "service") {
		causes = append(causes, "external_service_issue")
	}
	if strings.Contains(text, "timeout") || strings.Contains(text, "slow") {
		causes = append(causes, "performance_issue")
	}

	// If no specific causes identified, add general cause
	if len(causes) == 0 {
		causes = append(causes, "general_classification_error")
	}

	return causes
}

// generateMisclassificationRecommendations generates recommendations based on misclassification analysis
func (md *MisclassificationDetector) generateMisclassificationRecommendations(modelType string, feedback []*UserFeedback, rootCauses []string) []string {
	var recommendations []string

	// Generate recommendations based on root causes
	for _, cause := range rootCauses {
		switch cause {
		case "insufficient_data":
			recommendations = append(recommendations, "Improve data collection and validation processes")
		case "ambiguous_input":
			recommendations = append(recommendations, "Enhance input preprocessing and clarification mechanisms")
		case "outdated_data":
			recommendations = append(recommendations, "Implement data freshness monitoring and updates")
		case "model_limitation":
			recommendations = append(recommendations, "Consider model retraining or architecture improvements")
		case "insufficient_training":
			recommendations = append(recommendations, "Increase training data diversity and volume")
		case "model_bias":
			recommendations = append(recommendations, "Implement bias detection and mitigation strategies")
		case "keyword_matching_issue":
			recommendations = append(recommendations, "Enhance keyword matching algorithms and synonym handling")
		case "synonym_handling":
			recommendations = append(recommendations, "Improve synonym and semantic matching capabilities")
		case "security_validation_issue":
			recommendations = append(recommendations, "Strengthen security validation and trust verification")
		case "verification_failure":
			recommendations = append(recommendations, "Improve website and data source verification processes")
		case "external_service_issue":
			recommendations = append(recommendations, "Implement better external service error handling and fallbacks")
		case "performance_issue":
			recommendations = append(recommendations, "Optimize performance and implement timeout handling")
		}
	}

	// Generate model-specific recommendations
	switch modelType {
	case "ml_model", "bert_model":
		recommendations = append(recommendations, "Consider fine-tuning ML models with domain-specific data")
		recommendations = append(recommendations, "Implement model drift detection and retraining")
	case "keyword_model":
		recommendations = append(recommendations, "Expand keyword database with industry-specific terms")
		recommendations = append(recommendations, "Implement fuzzy matching and semantic similarity")
	case "ensemble_model":
		recommendations = append(recommendations, "Optimize ensemble weights based on method performance")
		recommendations = append(recommendations, "Implement dynamic weight adjustment")
	case "similarity_model":
		recommendations = append(recommendations, "Improve semantic similarity algorithms")
		recommendations = append(recommendations, "Enhance vector representations and embeddings")
	}

	// Remove duplicates
	uniqueRecommendations := make(map[string]bool)
	var finalRecommendations []string
	for _, rec := range recommendations {
		if !uniqueRecommendations[rec] {
			uniqueRecommendations[rec] = true
			finalRecommendations = append(finalRecommendations, rec)
		}
	}

	return finalRecommendations
}

// determinePrimaryModelType determines the primary model type from a group of feedback
func (md *MisclassificationDetector) determinePrimaryModelType(feedback []*UserFeedback) string {
	modelCounts := make(map[string]int)

	for _, fb := range feedback {
		modelType := md.determineModelType(fb)
		modelCounts[modelType]++
	}

	// Find the most common model type
	var primaryModel string
	maxCount := 0
	for modelType, count := range modelCounts {
		if count > maxCount {
			maxCount = count
			primaryModel = modelType
		}
	}

	return primaryModel
}
