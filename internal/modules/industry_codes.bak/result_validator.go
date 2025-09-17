package industry_codes

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"time"

	"go.uber.org/zap"
)

// ValidationLevel represents the severity of a validation issue
type ValidationLevel string

const (
	ValidationLevelError   ValidationLevel = "error"
	ValidationLevelWarning ValidationLevel = "warning"
	ValidationLevelInfo    ValidationLevel = "info"
)

// ValidationIssue represents a specific validation issue
type ValidationIssue struct {
	Level       ValidationLevel `json:"level"`
	Field       string          `json:"field"`
	Message     string          `json:"message"`
	Code        string          `json:"code,omitempty"`
	Confidence  float64         `json:"confidence,omitempty"`
	Rule        string          `json:"rule"`
	Suggestions []string        `json:"suggestions,omitempty"`
}

// ResultValidationResult represents the result of validation
type ResultValidationResult struct {
	IsValid         bool                   `json:"is_valid"`
	OverallScore    float64                `json:"overall_score"`
	Issues          []ValidationIssue      `json:"issues"`
	QualityMetrics  ResultQualityMetrics   `json:"quality_metrics"`
	ValidationTime  time.Duration          `json:"validation_time"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ResultQualityMetrics represents quality metrics for validation
type ResultQualityMetrics struct {
	DataCompleteness      float64 `json:"data_completeness"`
	DataConsistency       float64 `json:"data_consistency"`
	ConfidenceReliability float64 `json:"confidence_reliability"`
	CodeAccuracy          float64 `json:"code_accuracy"`
	OverallQuality        float64 `json:"overall_quality"`
}

// ResultValidationRule represents a validation rule
type ResultValidationRule struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Level       ValidationLevel        `json:"level"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"`
}

// ResultValidator provides comprehensive validation for classification results
type ResultValidator struct {
	rules  map[string]ResultValidationRule
	logger *zap.Logger
	config *ValidationConfig
}

// ValidationConfig represents validation configuration
type ValidationConfig struct {
	MinConfidenceThreshold float64 `json:"min_confidence_threshold"`
	MaxConfidenceThreshold float64 `json:"max_confidence_threshold"`
	MinResultsCount        int     `json:"min_results_count"`
	MaxResultsCount        int     `json:"max_results_count"`
	MinQualityScore        float64 `json:"min_quality_score"`
	EnableStrictValidation bool    `json:"enable_strict_validation"`
	EnableQualityMetrics   bool    `json:"enable_quality_metrics"`
	EnableRecommendations  bool    `json:"enable_recommendations"`
}

// NewResultValidator creates a new result validator
func NewResultValidator(logger *zap.Logger) *ResultValidator {
	config := &ValidationConfig{
		MinConfidenceThreshold: 0.1,
		MaxConfidenceThreshold: 1.0,
		MinResultsCount:        1,
		MaxResultsCount:        50,
		MinQualityScore:        0.7,
		EnableStrictValidation: false,
		EnableQualityMetrics:   true,
		EnableRecommendations:  true,
	}

	validator := &ResultValidator{
		rules:  make(map[string]ResultValidationRule),
		logger: logger,
		config: config,
	}

	validator.initializeDefaultRules()
	return validator
}

// NewResultValidatorWithConfig creates a new result validator with custom configuration
func NewResultValidatorWithConfig(config *ValidationConfig, logger *zap.Logger) *ResultValidator {
	validator := &ResultValidator{
		rules:  make(map[string]ResultValidationRule),
		logger: logger,
		config: config,
	}

	validator.initializeDefaultRules()
	return validator
}

// initializeDefaultRules initializes default validation rules
func (rv *ResultValidator) initializeDefaultRules() {
	defaultRules := []ResultValidationRule{
		{
			Name:        "confidence_range",
			Description: "Validate confidence scores are within acceptable range",
			Level:       ValidationLevelError,
			Enabled:     true,
			Config: map[string]interface{}{
				"min_confidence": rv.config.MinConfidenceThreshold,
				"max_confidence": rv.config.MaxConfidenceThreshold,
			},
		},
		{
			Name:        "results_count",
			Description: "Validate number of results is within acceptable range",
			Level:       ValidationLevelWarning,
			Enabled:     true,
			Config: map[string]interface{}{
				"min_count": rv.config.MinResultsCount,
				"max_count": rv.config.MaxResultsCount,
			},
		},
		{
			Name:        "code_format",
			Description: "Validate industry code formats",
			Level:       ValidationLevelError,
			Enabled:     true,
			Config: map[string]interface{}{
				"sic_pattern":   `^\d{4}$`,
				"naics_pattern": `^\d{6}$`,
				"mcc_pattern":   `^\d{4}$`,
			},
		},
		{
			Name:        "confidence_consistency",
			Description: "Validate confidence score consistency across results",
			Level:       ValidationLevelWarning,
			Enabled:     true,
			Config: map[string]interface{}{
				"max_variance": 0.3,
			},
		},
		{
			Name:        "code_uniqueness",
			Description: "Validate no duplicate codes in results",
			Level:       ValidationLevelWarning,
			Enabled:     true,
			Config:      map[string]interface{}{},
		},
		{
			Name:        "type_distribution",
			Description: "Validate distribution of code types",
			Level:       ValidationLevelInfo,
			Enabled:     true,
			Config: map[string]interface{}{
				"min_types": 1,
				"max_types": 3,
			},
		},
		{
			Name:        "quality_threshold",
			Description: "Validate overall quality meets minimum threshold",
			Level:       ValidationLevelError,
			Enabled:     true,
			Config: map[string]interface{}{
				"min_quality": rv.config.MinQualityScore,
			},
		},
	}

	for _, rule := range defaultRules {
		rv.rules[rule.Name] = rule
	}
}

// ValidateResults validates classification results
func (rv *ResultValidator) ValidateResults(ctx context.Context, response *ClassificationResponse) (*ResultValidationResult, error) {
	startTime := time.Now()

	rv.logger.Info("Starting result validation",
		zap.String("business_name", response.Request.BusinessName),
		zap.Int("results_count", len(response.Results)))

	validationResult := &ResultValidationResult{
		IsValid:         true,
		OverallScore:    0.0,
		Issues:          make([]ValidationIssue, 0),
		QualityMetrics:  ResultQualityMetrics{},
		ValidationTime:  0,
		Recommendations: make([]string, 0),
		Metadata:        make(map[string]interface{}),
	}

	// Apply validation rules
	rv.applyValidationRules(ctx, response, validationResult)

	// Calculate quality metrics
	if rv.config.EnableQualityMetrics {
		rv.calculateQualityMetrics(ctx, response, validationResult)
	}

	// Apply quality threshold validation after metrics are calculated
	rv.applyQualityThresholdValidation(validationResult)

	// Generate recommendations
	if rv.config.EnableRecommendations {
		rv.generateRecommendations(ctx, validationResult)
	}

	// Calculate overall score
	rv.calculateOverallScore(validationResult)

	// Determine if results are valid
	validationResult.IsValid = rv.determineValidity(validationResult)

	validationResult.ValidationTime = time.Since(startTime)

	rv.logger.Info("Result validation completed",
		zap.Bool("is_valid", validationResult.IsValid),
		zap.Float64("overall_score", validationResult.OverallScore),
		zap.Int("issues_count", len(validationResult.Issues)),
		zap.Duration("validation_time", validationResult.ValidationTime))

	return validationResult, nil
}

// applyValidationRules applies all enabled validation rules
func (rv *ResultValidator) applyValidationRules(ctx context.Context, response *ClassificationResponse, result *ResultValidationResult) {
	for ruleName, rule := range rv.rules {
		if !rule.Enabled {
			continue
		}

		switch ruleName {
		case "confidence_range":
			rv.validateConfidenceRange(response, rule, result)
		case "results_count":
			rv.validateResultsCount(response, rule, result)
		case "code_format":
			rv.validateCodeFormat(response, rule, result)
		case "confidence_consistency":
			rv.validateConfidenceConsistency(response, rule, result)
		case "code_uniqueness":
			rv.validateCodeUniqueness(response, rule, result)
		case "type_distribution":
			rv.validateTypeDistribution(response, rule, result)
		}
	}
}

// validateConfidenceRange validates confidence scores are within acceptable range
func (rv *ResultValidator) validateConfidenceRange(response *ClassificationResponse, rule ResultValidationRule, result *ResultValidationResult) {
	minConfidence := rule.Config["min_confidence"].(float64)
	maxConfidence := rule.Config["max_confidence"].(float64)

	for i, classificationResult := range response.Results {
		if classificationResult.Confidence < minConfidence || classificationResult.Confidence > maxConfidence {
			issue := ValidationIssue{
				Level: rule.Level,
				Field: fmt.Sprintf("results[%d].confidence", i),
				Message: fmt.Sprintf("Confidence score %.3f is outside acceptable range [%.3f, %.3f]",
					classificationResult.Confidence, minConfidence, maxConfidence),
				Code:       classificationResult.Code.Code,
				Confidence: classificationResult.Confidence,
				Rule:       rule.Name,
				Suggestions: []string{
					"Review confidence scoring algorithm",
					"Check input data quality",
					"Consider adjusting confidence thresholds",
				},
			}
			result.Issues = append(result.Issues, issue)
		}
	}
}

// validateResultsCount validates number of results is within acceptable range
func (rv *ResultValidator) validateResultsCount(response *ClassificationResponse, rule ResultValidationRule, result *ResultValidationResult) {
	minCount := rule.Config["min_count"].(int)
	maxCount := rule.Config["max_count"].(int)
	resultsCount := len(response.Results)

	if resultsCount < minCount {
		issue := ValidationIssue{
			Level:   rule.Level,
			Field:   "results_count",
			Message: fmt.Sprintf("Results count %d is below minimum required %d", resultsCount, minCount),
			Rule:    rule.Name,
			Suggestions: []string{
				"Lower confidence threshold",
				"Expand search criteria",
				"Add more keywords",
			},
		}
		result.Issues = append(result.Issues, issue)
	}

	if resultsCount > maxCount {
		issue := ValidationIssue{
			Level:   rule.Level,
			Field:   "results_count",
			Message: fmt.Sprintf("Results count %d exceeds maximum allowed %d", resultsCount, maxCount),
			Rule:    rule.Name,
			Suggestions: []string{
				"Increase confidence threshold",
				"Limit search to specific code types",
				"Apply stricter filtering",
			},
		}
		result.Issues = append(result.Issues, issue)
	}
}

// validateCodeFormat validates industry code formats
func (rv *ResultValidator) validateCodeFormat(response *ClassificationResponse, rule ResultValidationRule, result *ResultValidationResult) {
	sicPattern := regexp.MustCompile(rule.Config["sic_pattern"].(string))
	naicsPattern := regexp.MustCompile(rule.Config["naics_pattern"].(string))
	mccPattern := regexp.MustCompile(rule.Config["mcc_pattern"].(string))

	for i, classificationResult := range response.Results {
		code := classificationResult.Code.Code
		codeType := classificationResult.Code.Type

		var isValid bool
		switch codeType {
		case CodeTypeSIC:
			isValid = sicPattern.MatchString(code)
		case CodeTypeNAICS:
			isValid = naicsPattern.MatchString(code)
		case CodeTypeMCC:
			isValid = mccPattern.MatchString(code)
		default:
			isValid = true // Unknown type, skip validation
		}

		if !isValid {
			issue := ValidationIssue{
				Level:   rule.Level,
				Field:   fmt.Sprintf("results[%d].code.code", i),
				Message: fmt.Sprintf("Code '%s' does not match expected format for type '%s'", code, codeType),
				Code:    code,
				Rule:    rule.Name,
				Suggestions: []string{
					"Verify code format in database",
					"Check code type assignment",
					"Review code standardization process",
				},
			}
			result.Issues = append(result.Issues, issue)
		}
	}
}

// validateConfidenceConsistency validates confidence score consistency
func (rv *ResultValidator) validateConfidenceConsistency(response *ClassificationResponse, rule ResultValidationRule, result *ResultValidationResult) {
	if len(response.Results) < 2 {
		return // Need at least 2 results to check consistency
	}

	maxVariance := rule.Config["max_variance"].(float64)
	confidences := make([]float64, len(response.Results))

	for i, classificationResult := range response.Results {
		confidences[i] = classificationResult.Confidence
	}

	variance := rv.calculateVariance(confidences)
	if variance > maxVariance {
		issue := ValidationIssue{
			Level:   rule.Level,
			Field:   "confidence_consistency",
			Message: fmt.Sprintf("Confidence variance %.3f exceeds maximum allowed %.3f", variance, maxVariance),
			Rule:    rule.Name,
			Suggestions: []string{
				"Review confidence scoring algorithm",
				"Check for data quality issues",
				"Consider normalizing confidence scores",
			},
		}
		result.Issues = append(result.Issues, issue)
	}
}

// validateCodeUniqueness validates no duplicate codes in results
func (rv *ResultValidator) validateCodeUniqueness(response *ClassificationResponse, rule ResultValidationRule, result *ResultValidationResult) {
	seenCodes := make(map[string]int)

	for i, classificationResult := range response.Results {
		codeKey := fmt.Sprintf("%s-%s", classificationResult.Code.Type, classificationResult.Code.Code)

		if existingIndex, exists := seenCodes[codeKey]; exists {
			issue := ValidationIssue{
				Level:   rule.Level,
				Field:   fmt.Sprintf("results[%d].code", i),
				Message: fmt.Sprintf("Duplicate code '%s' found at index %d", codeKey, existingIndex),
				Code:    classificationResult.Code.Code,
				Rule:    rule.Name,
				Suggestions: []string{
					"Implement deduplication logic",
					"Review result aggregation process",
					"Check for duplicate database entries",
				},
			}
			result.Issues = append(result.Issues, issue)
		} else {
			seenCodes[codeKey] = i
		}
	}
}

// validateTypeDistribution validates distribution of code types
func (rv *ResultValidator) validateTypeDistribution(response *ClassificationResponse, rule ResultValidationRule, result *ResultValidationResult) {
	minTypes := rule.Config["min_types"].(int)
	maxTypes := rule.Config["max_types"].(int)

	typeCount := make(map[CodeType]int)
	for _, classificationResult := range response.Results {
		typeCount[classificationResult.Code.Type]++
	}

	uniqueTypes := len(typeCount)
	if uniqueTypes < minTypes {
		issue := ValidationIssue{
			Level:   rule.Level,
			Field:   "type_distribution",
			Message: fmt.Sprintf("Only %d unique code types found, minimum expected %d", uniqueTypes, minTypes),
			Rule:    rule.Name,
			Suggestions: []string{
				"Expand search to include more code types",
				"Lower confidence thresholds for different types",
				"Review code type preferences",
			},
		}
		result.Issues = append(result.Issues, issue)
	}

	if uniqueTypes > maxTypes {
		issue := ValidationIssue{
			Level:   rule.Level,
			Field:   "type_distribution",
			Message: fmt.Sprintf("%d unique code types found, maximum expected %d", uniqueTypes, maxTypes),
			Rule:    rule.Name,
			Suggestions: []string{
				"Limit search to specific code types",
				"Increase confidence thresholds",
				"Apply type-specific filtering",
			},
		}
		result.Issues = append(result.Issues, issue)
	}
}

// validateQualityThreshold validates overall quality meets minimum threshold
func (rv *ResultValidator) validateQualityThreshold(response *ClassificationResponse, rule ResultValidationRule, result *ResultValidationResult) {
	minQuality := rule.Config["min_quality"].(float64)

	if result.QualityMetrics.OverallQuality < minQuality {
		issue := ValidationIssue{
			Level: rule.Level,
			Field: "overall_quality",
			Message: fmt.Sprintf("Overall quality score %.3f is below minimum threshold %.3f",
				result.QualityMetrics.OverallQuality, minQuality),
			Rule: rule.Name,
			Suggestions: []string{
				"Improve input data quality",
				"Review classification algorithms",
				"Adjust quality thresholds",
			},
		}
		result.Issues = append(result.Issues, issue)
	}
}

// applyQualityThresholdValidation applies quality threshold validation after metrics are calculated
func (rv *ResultValidator) applyQualityThresholdValidation(result *ResultValidationResult) {
	rule, exists := rv.rules["quality_threshold"]
	if !exists || !rule.Enabled {
		return
	}

	minQuality := rule.Config["min_quality"].(float64)

	if result.QualityMetrics.OverallQuality < minQuality {
		issue := ValidationIssue{
			Level: rule.Level,
			Field: "overall_quality",
			Message: fmt.Sprintf("Overall quality score %.3f is below minimum threshold %.3f",
				result.QualityMetrics.OverallQuality, minQuality),
			Rule: rule.Name,
			Suggestions: []string{
				"Improve input data quality",
				"Review classification algorithms",
				"Adjust quality thresholds",
			},
		}
		result.Issues = append(result.Issues, issue)
	}
}

// calculateQualityMetrics calculates quality metrics for the results
func (rv *ResultValidator) calculateQualityMetrics(ctx context.Context, response *ClassificationResponse, result *ResultValidationResult) {
	metrics := &result.QualityMetrics

	// Data completeness
	metrics.DataCompleteness = rv.calculateDataCompleteness(response)

	// Data consistency
	metrics.DataConsistency = rv.calculateDataConsistency(response)

	// Confidence reliability
	metrics.ConfidenceReliability = rv.calculateConfidenceReliability(response)

	// Code accuracy
	metrics.CodeAccuracy = rv.calculateCodeAccuracy(response)

	// Overall quality (weighted average)
	metrics.OverallQuality = (metrics.DataCompleteness*0.25 +
		metrics.DataConsistency*0.25 +
		metrics.ConfidenceReliability*0.25 +
		metrics.CodeAccuracy*0.25)
}

// calculateDataCompleteness calculates data completeness score
func (rv *ResultValidator) calculateDataCompleteness(response *ClassificationResponse) float64 {
	if len(response.Results) == 0 {
		return 0.0
	}

	completeResults := 0
	for _, result := range response.Results {
		if result.Code != nil &&
			result.Code.Code != "" &&
			result.Code.Description != "" &&
			result.Code.Type != "" {
			completeResults++
		}
	}

	return float64(completeResults) / float64(len(response.Results))
}

// calculateDataConsistency calculates data consistency score
func (rv *ResultValidator) calculateDataConsistency(response *ClassificationResponse) float64 {
	if len(response.Results) < 2 {
		return 1.0 // Single result is always consistent
	}

	// Check for consistency in confidence scores
	confidences := make([]float64, len(response.Results))
	for i, result := range response.Results {
		confidences[i] = result.Confidence
	}

	variance := rv.calculateVariance(confidences)
	// Convert variance to consistency score (lower variance = higher consistency)
	consistency := math.Max(0, 1-variance)

	return consistency
}

// calculateConfidenceReliability calculates confidence reliability score
func (rv *ResultValidator) calculateConfidenceReliability(response *ClassificationResponse) float64 {
	if len(response.Results) == 0 {
		return 0.0
	}

	// Calculate average confidence
	totalConfidence := 0.0
	for _, result := range response.Results {
		totalConfidence += result.Confidence
	}
	averageConfidence := totalConfidence / float64(len(response.Results))

	// Calculate confidence stability (how close individual scores are to average)
	stability := 0.0
	for _, result := range response.Results {
		stability += 1 - math.Abs(result.Confidence-averageConfidence)
	}
	stability /= float64(len(response.Results))

	// Combine average confidence with stability
	reliability := (averageConfidence + stability) / 2

	return reliability
}

// calculateCodeAccuracy calculates code accuracy score
func (rv *ResultValidator) calculateCodeAccuracy(response *ClassificationResponse) float64 {
	if len(response.Results) == 0 {
		return 0.0
	}

	validCodes := 0
	for _, result := range response.Results {
		if rv.isValidCodeFormat(result.Code) {
			validCodes++
		}
	}

	return float64(validCodes) / float64(len(response.Results))
}

// isValidCodeFormat checks if a code has valid format
func (rv *ResultValidator) isValidCodeFormat(code *IndustryCode) bool {
	if code == nil || code.Code == "" {
		return false
	}

	switch code.Type {
	case CodeTypeSIC:
		return regexp.MustCompile(`^\d{4}$`).MatchString(code.Code)
	case CodeTypeNAICS:
		return regexp.MustCompile(`^\d{6}$`).MatchString(code.Code)
	case CodeTypeMCC:
		return regexp.MustCompile(`^\d{4}$`).MatchString(code.Code)
	default:
		return true // Unknown type, assume valid
	}
}

// calculateVariance calculates variance of a slice of float64 values
func (rv *ResultValidator) calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	// Calculate mean
	mean := 0.0
	for _, value := range values {
		mean += value
	}
	mean /= float64(len(values))

	// Calculate variance
	variance := 0.0
	for _, value := range values {
		variance += math.Pow(value-mean, 2)
	}
	variance /= float64(len(values))

	return variance
}

// generateRecommendations generates recommendations based on validation results
func (rv *ResultValidator) generateRecommendations(ctx context.Context, result *ResultValidationResult) {
	recommendations := make([]string, 0)

	// Analyze issues and generate recommendations
	errorCount := 0
	warningCount := 0

	for _, issue := range result.Issues {
		switch issue.Level {
		case ValidationLevelError:
			errorCount++
		case ValidationLevelWarning:
			warningCount++
		}
	}

	if errorCount > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Address %d critical validation errors before using results", errorCount))
	}

	if warningCount > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Review %d validation warnings to improve result quality", warningCount))
	}

	if result.QualityMetrics.OverallQuality < 0.8 {
		recommendations = append(recommendations,
			"Consider improving input data quality for better classification results")
	}

	if len(result.Issues) == 0 {
		recommendations = append(recommendations,
			"Results passed all validation checks - ready for use")
	}

	result.Recommendations = recommendations
}

// calculateOverallScore calculates overall validation score
func (rv *ResultValidator) calculateOverallScore(result *ResultValidationResult) {
	// Start with quality score
	score := result.QualityMetrics.OverallQuality

	// Penalize for issues
	errorPenalty := 0.0
	warningPenalty := 0.0

	for _, issue := range result.Issues {
		switch issue.Level {
		case ValidationLevelError:
			errorPenalty += 0.2
		case ValidationLevelWarning:
			warningPenalty += 0.05
		}
	}

	// Apply penalties
	score -= errorPenalty
	score -= warningPenalty

	// Ensure score is within [0, 1] range
	result.OverallScore = math.Max(0, math.Min(1, score))
}

// determineValidity determines if results are valid based on validation criteria
func (rv *ResultValidator) determineValidity(result *ResultValidationResult) bool {
	// Check for critical errors
	for _, issue := range result.Issues {
		if issue.Level == ValidationLevelError {
			return false
		}
	}

	// Check quality threshold
	if result.QualityMetrics.OverallQuality < rv.config.MinQualityScore {
		return false
	}

	// Check overall score
	if result.OverallScore < rv.config.MinQualityScore {
		return false
	}

	return true
}

// AddRule adds a custom validation rule
func (rv *ResultValidator) AddRule(rule ResultValidationRule) {
	rv.rules[rule.Name] = rule
}

// RemoveRule removes a validation rule
func (rv *ResultValidator) RemoveRule(ruleName string) {
	delete(rv.rules, ruleName)
}

// GetRule gets a validation rule by name
func (rv *ResultValidator) GetRule(ruleName string) (ResultValidationRule, bool) {
	rule, exists := rv.rules[ruleName]
	return rule, exists
}

// ListRules lists all validation rules
func (rv *ResultValidator) ListRules() []ResultValidationRule {
	rules := make([]ResultValidationRule, 0, len(rv.rules))
	for _, rule := range rv.rules {
		rules = append(rules, rule)
	}
	return rules
}

// UpdateConfig updates validation configuration
func (rv *ResultValidator) UpdateConfig(config *ValidationConfig) {
	rv.config = config
	// Reinitialize rules with new config
	rv.initializeDefaultRules()
}

// GetConfig returns current validation configuration
func (rv *ResultValidator) GetConfig() *ValidationConfig {
	return rv.config
}
