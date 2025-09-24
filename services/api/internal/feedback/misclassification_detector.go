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
func (md *MisclassificationDetector) DetectModelMisclassifications(ctx context.Context, feedback []*UserFeedback) ([]*ModelMisclassification, error) {
	if len(feedback) < md.config.MinFeedbackThreshold {
		return []*ModelMisclassification{}, nil
	}

	var misclassifications []*ModelMisclassification

	// Group feedback by model type and misclassification type
	modelGroups := make(map[string][]*UserFeedback)
	misclassificationGroups := make(map[string][]*UserFeedback)

	for _, fb := range feedback {
		// Determine model type from feedback
		modelType := md.determineModelType(fb)
		modelGroups[modelType] = append(modelGroups[modelType], fb)

		// Determine misclassification type
		misclassificationType := md.determineMisclassificationType(fb)
		misclassificationGroups[misclassificationType] = append(misclassificationGroups[misclassificationType], fb)
	}

	// Analyze model-specific misclassifications
	for modelType, modelFeedback := range modelGroups {
		if len(modelFeedback) < md.config.MinFeedbackThreshold {
			continue
		}

		misclassification := md.analyzeModelMisclassification(modelType, modelFeedback)
		if misclassification != nil {
			misclassifications = append(misclassifications, misclassification)
		}
	}

	// Analyze misclassification type patterns
	for misclassificationType, typeFeedback := range misclassificationGroups {
		if len(typeFeedback) < md.config.MinFeedbackThreshold {
			continue
		}

		misclassification := md.analyzeMisclassificationType(misclassificationType, typeFeedback)
		if misclassification != nil {
			misclassifications = append(misclassifications, misclassification)
		}
	}

	// Sort by frequency and confidence
	sort.Slice(misclassifications, func(i, j int) bool {
		if misclassifications[i].Frequency == misclassifications[j].Frequency {
			return misclassifications[i].Confidence > misclassifications[j].Confidence
		}
		return misclassifications[i].Frequency > misclassifications[j].Frequency
	})

	return misclassifications, nil
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
func (md *MisclassificationDetector) determineModelType(feedback *UserFeedback) string {
	// Check classification method
	if feedback.ClassificationMethod == MethodML {
		return "ml_model"
	}

	// Check feedback text for model indicators
	text := strings.ToLower(feedback.FeedbackText)
	if strings.Contains(text, "bert") || strings.Contains(text, "transformer") {
		return "bert_model"
	} else if strings.Contains(text, "keyword") || strings.Contains(text, "matching") {
		return "keyword_model"
	} else if strings.Contains(text, "ensemble") || strings.Contains(text, "combined") {
		return "ensemble_model"
	} else if strings.Contains(text, "similarity") || strings.Contains(text, "semantic") {
		return "similarity_model"
	}

	// Default based on classification method
	switch feedback.ClassificationMethod {
	case MethodKeyword:
		return "keyword_model"
	case MethodSimilarity:
		return "similarity_model"
	case MethodEnsemble:
		return "ensemble_model"
	default:
		return "unknown_model"
	}
}

// determineMisclassificationType determines the type of misclassification
func (md *MisclassificationDetector) determineMisclassificationType(feedback *UserFeedback) string {
	text := strings.ToLower(feedback.FeedbackText)

	// Check for specific misclassification types
	if strings.Contains(text, "industry") || strings.Contains(text, "category") {
		return "industry_misclassification"
	} else if strings.Contains(text, "confidence") || strings.Contains(text, "score") {
		return "confidence_miscalibration"
	} else if strings.Contains(text, "keyword") || strings.Contains(text, "match") {
		return "keyword_mismatch"
	} else if strings.Contains(text, "security") || strings.Contains(text, "trust") {
		return "security_validation_failure"
	} else if strings.Contains(text, "website") || strings.Contains(text, "domain") {
		return "website_analysis_failure"
	} else if strings.Contains(text, "ml") || strings.Contains(text, "model") {
		return "ml_model_failure"
	} else if strings.Contains(text, "ensemble") || strings.Contains(text, "combination") {
		return "ensemble_disagreement"
	}

	// Check feedback type
	switch feedback.FeedbackType {
	case FeedbackTypeCorrection:
		return "classification_correction"
	case FeedbackTypeAccuracy:
		return "accuracy_issue"
	case FeedbackTypeConfidence:
		return "confidence_issue"
	case FeedbackTypeRelevance:
		return "relevance_issue"
	case FeedbackTypeSecurityValidation:
		return "security_validation_issue"
	default:
		return "general_misclassification"
	}
}

// calculateMisclassificationConfidence calculates confidence in the misclassification detection
func (md *MisclassificationDetector) calculateMisclassificationConfidence(feedback []*UserFeedback) float64 {
	if len(feedback) == 0 {
		return 0.0
	}

	var totalConfidence float64
	var confidenceCount int

	for _, fb := range feedback {
		// Use feedback confidence score if available
		if fb.ConfidenceScore > 0 {
			totalConfidence += fb.ConfidenceScore
			confidenceCount++
		} else {
			// Calculate confidence based on feedback type and content
			confidence := md.calculateFeedbackConfidence(fb)
			totalConfidence += confidence
			confidenceCount++
		}
	}

	if confidenceCount == 0 {
		return 0.5 // Default confidence
	}

	return totalConfidence / float64(confidenceCount)
}

// calculateFeedbackConfidence calculates confidence based on feedback content
func (md *MisclassificationDetector) calculateFeedbackConfidence(feedback *UserFeedback) float64 {
	confidence := 0.5 // Base confidence

	// Adjust based on feedback type
	switch feedback.FeedbackType {
	case FeedbackTypeCorrection:
		confidence += 0.3 // High confidence for corrections
	case FeedbackTypeAccuracy:
		confidence += 0.2 // Medium-high confidence for accuracy feedback
	case FeedbackTypeConfidence:
		confidence += 0.1 // Medium confidence for confidence feedback
	case FeedbackTypeRelevance:
		confidence += 0.1 // Medium confidence for relevance feedback
	case FeedbackTypeSecurityValidation:
		confidence += 0.4 // Very high confidence for security issues
	}

	// Adjust based on feedback text quality
	text := strings.ToLower(feedback.FeedbackText)
	if len(text) > 50 {
		confidence += 0.1 // More detailed feedback is more reliable
	}
	if strings.Contains(text, "definitely") || strings.Contains(text, "clearly") {
		confidence += 0.1 // Strong language indicates high confidence
	}
	if strings.Contains(text, "maybe") || strings.Contains(text, "possibly") {
		confidence -= 0.1 // Uncertain language reduces confidence
	}

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	} else if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// analyzeAffectedIndustries analyzes which industries are most affected
func (md *MisclassificationDetector) analyzeAffectedIndustries(feedback []*UserFeedback) []string {
	industryCounts := make(map[string]int)

	for _, fb := range feedback {
		// Extract industry information from feedback
		industries := md.extractIndustriesFromFeedback(fb)
		for _, industry := range industries {
			industryCounts[industry]++
		}
	}

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
func (md *MisclassificationDetector) extractIndustriesFromFeedback(feedback *UserFeedback) []string {
	var industries []string
	text := strings.ToLower(feedback.FeedbackText)

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
	for keyword, industry := range industryKeywords {
		if strings.Contains(text, keyword) {
			// Avoid duplicates
			found := false
			for _, existing := range industries {
				if existing == industry {
					found = true
					break
				}
			}
			if !found {
				industries = append(industries, industry)
			}
		}
	}

	// If no industries found, try to extract from business name or metadata
	if len(industries) == 0 {
		// Check business name
		businessName := strings.ToLower(feedback.BusinessName)
		for keyword, industry := range industryKeywords {
			if strings.Contains(businessName, keyword) {
				industries = append(industries, industry)
				break
			}
		}
	}

	return industries
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
func (md *MisclassificationDetector) analyzeBusinessNamePatterns(feedback []*UserFeedback) []string {
	var patterns []string
	nameLengths := make(map[string]int)
	nameWords := make(map[string]int)

	for _, fb := range feedback {
		name := strings.TrimSpace(fb.BusinessName)

		// Analyze name length
		length := len(name)
		if length < 10 {
			nameLengths["short_names"]++
		} else if length > 50 {
			nameLengths["long_names"]++
		} else {
			nameLengths["medium_names"]++
		}

		// Analyze word count
		words := strings.Fields(name)
		wordCount := len(words)
		if wordCount == 1 {
			nameWords["single_word"]++
		} else if wordCount > 5 {
			nameWords["many_words"]++
		} else {
			nameWords["few_words"]++
		}
	}

	// Add patterns that occur frequently
	for pattern, count := range nameLengths {
		if count > len(feedback)/3 {
			patterns = append(patterns, fmt.Sprintf("business_name_%s", pattern))
		}
	}

	for pattern, count := range nameWords {
		if count > len(feedback)/3 {
			patterns = append(patterns, fmt.Sprintf("business_name_%s", pattern))
		}
	}

	return patterns
}

// analyzeDescriptionPatterns analyzes patterns in descriptions
func (md *MisclassificationDetector) analyzeDescriptionPatterns(feedback []*UserFeedback) []string {
	var patterns []string
	descriptionLengths := make(map[string]int)
	hasWebsite := 0
	hasEmail := 0

	for _, fb := range feedback {
		// Check if feedback mentions description issues
		text := strings.ToLower(fb.FeedbackText)
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

	for _, fb := range feedback {
		text := strings.ToLower(fb.FeedbackText)
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
func (md *MisclassificationDetector) identifyRootCauses(feedback *UserFeedback) []string {
	var causes []string
	text := strings.ToLower(feedback.FeedbackText)

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
