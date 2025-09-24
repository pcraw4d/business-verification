package data_extraction

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"kyb-platform/internal/observability"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DataQualityFramework provides comprehensive data quality assessment
type DataQualityFramework struct {
	// Configuration
	config *DataQualityConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Validation rules
	validationRules map[string][]ValidationRule

	// Quality metrics
	qualityMetrics map[string]QualityMetric

	// Cross-validation rules
	crossValidationRules []CrossValidationRule
}

// DataQualityConfig holds configuration for the data quality framework
type DataQualityConfig struct {
	// Quality scoring settings
	EnableAccuracyScoring     bool
	EnableCompletenessScoring bool
	EnableFreshnessScoring    bool
	EnableConsistencyScoring  bool

	// Validation settings
	StrictValidation    bool
	MinQualityThreshold float64
	MaxValidationErrors int

	// Freshness settings
	DataFreshnessThreshold time.Duration
	StaleDataPenalty       float64

	// Cross-validation settings
	EnableCrossValidation bool
	CrossValidationWeight float64

	// Processing settings
	Timeout time.Duration
}

// DataQualityScore represents a comprehensive quality assessment
type DataQualityScore struct {
	// Overall quality score
	OverallScore float64 `json:"overall_score"`

	// Individual dimension scores
	AccuracyScore     float64 `json:"accuracy_score"`
	CompletenessScore float64 `json:"completeness_score"`
	FreshnessScore    float64 `json:"freshness_score"`
	ConsistencyScore  float64 `json:"consistency_score"`

	// Quality level assessment
	QualityLevel string `json:"quality_level"`
	QualityGrade string `json:"quality_grade"`

	// Detailed metrics
	AccuracyMetrics     AccuracyMetrics     `json:"accuracy_metrics"`
	CompletenessMetrics CompletenessMetrics `json:"completeness_metrics"`
	FreshnessMetrics    FreshnessMetrics    `json:"freshness_metrics"`
	ConsistencyMetrics  ConsistencyMetrics  `json:"consistency_metrics"`

	// Validation results
	ValidationResults []ValidationResult `json:"validation_results"`
	ValidationErrors  []ValidationError  `json:"validation_errors"`

	// Cross-validation results
	CrossValidationResults []CrossValidationResult `json:"cross_validation_results"`

	// Metadata
	AssessedAt  time.Time `json:"assessed_at"`
	DataSources []string  `json:"data_sources"`
}

// AccuracyMetrics represents accuracy assessment metrics
type AccuracyMetrics struct {
	// Pattern matching accuracy
	PatternMatchAccuracy float64 `json:"pattern_match_accuracy"`

	// Format validation accuracy
	FormatValidationAccuracy float64 `json:"format_validation_accuracy"`

	// Semantic validation accuracy
	SemanticValidationAccuracy float64 `json:"semantic_validation_accuracy"`

	// Confidence correlation
	ConfidenceCorrelation float64 `json:"confidence_correlation"`

	// Error rate
	ErrorRate float64 `json:"error_rate"`

	// Supporting evidence
	SupportingEvidence []string `json:"supporting_evidence"`
}

// CompletenessMetrics represents completeness assessment metrics
type CompletenessMetrics struct {
	// Required fields completeness
	RequiredFieldsCompleteness float64 `json:"required_fields_completeness"`

	// Optional fields completeness
	OptionalFieldsCompleteness float64 `json:"optional_fields_completeness"`

	// Data point coverage
	DataPointCoverage float64 `json:"data_point_coverage"`

	// Missing critical data
	MissingCriticalData []string `json:"missing_critical_data"`

	// Partial data indicators
	PartialDataIndicators []string `json:"partial_data_indicators"`
}

// FreshnessMetrics represents freshness assessment metrics
type FreshnessMetrics struct {
	// Data age
	DataAge time.Duration `json:"data_age"`

	// Freshness score
	FreshnessScore float64 `json:"freshness_score"`

	// Last update time
	LastUpdateTime time.Time `json:"last_update_time"`

	// Update frequency
	UpdateFrequency string `json:"update_frequency"`

	// Stale data indicators
	StaleDataIndicators []string `json:"stale_data_indicators"`
}

// ConsistencyMetrics represents consistency assessment metrics
type ConsistencyMetrics struct {
	// Internal consistency
	InternalConsistency float64 `json:"internal_consistency"`

	// Cross-field consistency
	CrossFieldConsistency float64 `json:"cross_field_consistency"`

	// Format consistency
	FormatConsistency float64 `json:"format_consistency"`

	// Value consistency
	ValueConsistency float64 `json:"value_consistency"`

	// Inconsistency indicators
	InconsistencyIndicators []string `json:"inconsistency_indicators"`
}

// ValidationRule represents a data validation rule
type ValidationRule struct {
	ID              string                  `json:"id"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	Field           string                  `json:"field"`
	Type            string                  `json:"type"`
	Pattern         string                  `json:"pattern,omitempty"`
	Required        bool                    `json:"required"`
	MinLength       int                     `json:"min_length,omitempty"`
	MaxLength       int                     `json:"max_length,omitempty"`
	MinValue        float64                 `json:"min_value,omitempty"`
	MaxValue        float64                 `json:"max_value,omitempty"`
	AllowedValues   []string                `json:"allowed_values,omitempty"`
	CustomValidator func(interface{}) error `json:"-"`
}

// ValidationResult represents the result of a validation rule
type ValidationResult struct {
	RuleID     string      `json:"rule_id"`
	RuleName   string      `json:"rule_name"`
	Field      string      `json:"field"`
	Value      interface{} `json:"value"`
	IsValid    bool        `json:"is_valid"`
	Error      string      `json:"error,omitempty"`
	Severity   string      `json:"severity"`
	Confidence float64     `json:"confidence"`
}

// ValidationError represents a validation error
type ValidationError struct {
	RuleID      string      `json:"rule_id"`
	Field       string      `json:"field"`
	Value       interface{} `json:"value"`
	Error       string      `json:"error"`
	Severity    string      `json:"severity"`
	Suggestions []string    `json:"suggestions"`
}

// CrossValidationRule represents a cross-validation rule
type CrossValidationRule struct {
	ID          string                             `json:"id"`
	Name        string                             `json:"name"`
	Description string                             `json:"description"`
	Fields      []string                           `json:"fields"`
	Validator   func(map[string]interface{}) error `json:"-"`
	Weight      float64                            `json:"weight"`
}

// CrossValidationResult represents the result of cross-validation
type CrossValidationResult struct {
	RuleID     string   `json:"rule_id"`
	RuleName   string   `json:"rule_name"`
	Fields     []string `json:"fields"`
	IsValid    bool     `json:"is_valid"`
	Error      string   `json:"error,omitempty"`
	Confidence float64  `json:"confidence"`
	Weight     float64  `json:"weight"`
}

// QualityMetric represents a quality metric
type QualityMetric struct {
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Weight      float64 `json:"weight"`
	Threshold   float64 `json:"threshold"`
	Description string  `json:"description"`
}

// NewDataQualityFramework creates a new data quality framework
func NewDataQualityFramework(
	config *DataQualityConfig,
	logger *observability.Logger,
	tracer trace.Tracer,
) *DataQualityFramework {
	// Set default configuration
	if config == nil {
		config = &DataQualityConfig{
			EnableAccuracyScoring:     true,
			EnableCompletenessScoring: true,
			EnableFreshnessScoring:    true,
			EnableConsistencyScoring:  true,
			StrictValidation:          false,
			MinQualityThreshold:       0.6,
			MaxValidationErrors:       10,
			DataFreshnessThreshold:    24 * time.Hour,
			StaleDataPenalty:          0.2,
			EnableCrossValidation:     true,
			CrossValidationWeight:     0.3,
			Timeout:                   30 * time.Second,
		}
	}

	framework := &DataQualityFramework{
		config:          config,
		logger:          logger,
		tracer:          tracer,
		validationRules: make(map[string][]ValidationRule),
		qualityMetrics:  make(map[string]QualityMetric),
	}

	// Initialize validation rules
	framework.initializeValidationRules()

	// Initialize cross-validation rules
	framework.initializeCrossValidationRules()

	return framework
}

// initializeValidationRules initializes all validation rules
func (dqf *DataQualityFramework) initializeValidationRules() {
	// Email validation rules
	dqf.validationRules["email"] = []ValidationRule{
		{
			ID:          "email_format",
			Name:        "Email Format Validation",
			Description: "Validates email format using regex pattern",
			Field:       "email",
			Type:        "regex",
			Pattern:     `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			Required:    false,
		},
		{
			ID:          "email_length",
			Name:        "Email Length Validation",
			Description: "Validates email length",
			Field:       "email",
			Type:        "length",
			MinLength:   5,
			MaxLength:   254,
			Required:    false,
		},
	}

	// Phone validation rules
	dqf.validationRules["phone"] = []ValidationRule{
		{
			ID:          "phone_format",
			Name:        "Phone Format Validation",
			Description: "Validates phone number format",
			Field:       "phone",
			Type:        "regex",
			Pattern:     `^(\+\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}$`,
			Required:    false,
		},
		{
			ID:          "phone_length",
			Name:        "Phone Length Validation",
			Description: "Validates phone number length",
			Field:       "phone",
			Type:        "length",
			MinLength:   10,
			MaxLength:   15,
			Required:    false,
		},
	}

	// Address validation rules
	dqf.validationRules["address"] = []ValidationRule{
		{
			ID:          "address_format",
			Name:        "Address Format Validation",
			Description: "Validates address format",
			Field:       "address",
			Type:        "regex",
			Pattern:     `^\d+\s+[a-zA-Z\s]+(?:street|st|avenue|ave|road|rd|boulevard|blvd|lane|ln|drive|dr|way|place|pl|court|ct)\.?`,
			Required:    false,
		},
		{
			ID:          "address_length",
			Name:        "Address Length Validation",
			Description: "Validates address length",
			Field:       "address",
			Type:        "length",
			MinLength:   10,
			MaxLength:   200,
			Required:    false,
		},
	}

	// Business name validation rules
	dqf.validationRules["business_name"] = []ValidationRule{
		{
			ID:          "business_name_required",
			Name:        "Business Name Required",
			Description: "Validates that business name is provided",
			Field:       "business_name",
			Type:        "required",
			Required:    true,
		},
		{
			ID:          "business_name_length",
			Name:        "Business Name Length",
			Description: "Validates business name length",
			Field:       "business_name",
			Type:        "length",
			MinLength:   2,
			MaxLength:   255,
			Required:    true,
		},
	}

	// Website validation rules
	dqf.validationRules["website"] = []ValidationRule{
		{
			ID:          "website_format",
			Name:        "Website Format Validation",
			Description: "Validates website URL format",
			Field:       "website",
			Type:        "regex",
			Pattern:     `^https?://[^\s/$.?#].[^\s]*$`,
			Required:    false,
		},
	}

	// Confidence validation rules
	dqf.validationRules["confidence"] = []ValidationRule{
		{
			ID:          "confidence_range",
			Name:        "Confidence Range Validation",
			Description: "Validates confidence score range",
			Field:       "confidence",
			Type:        "range",
			MinValue:    0.0,
			MaxValue:    1.0,
			Required:    false,
		},
	}
}

// initializeCrossValidationRules initializes cross-validation rules
func (dqf *DataQualityFramework) initializeCrossValidationRules() {
	dqf.crossValidationRules = []CrossValidationRule{
		{
			ID:          "contact_consistency",
			Name:        "Contact Information Consistency",
			Description: "Validates consistency between email and phone contact information",
			Fields:      []string{"email", "phone"},
			Validator:   dqf.validateContactConsistency,
			Weight:      0.3,
		},
		{
			ID:          "address_consistency",
			Name:        "Address Information Consistency",
			Description: "Validates consistency between address components",
			Fields:      []string{"address", "city", "state", "postal_code"},
			Validator:   dqf.validateAddressConsistency,
			Weight:      0.2,
		},
		{
			ID:          "business_consistency",
			Name:        "Business Information Consistency",
			Description: "Validates consistency between business name and website",
			Fields:      []string{"business_name", "website"},
			Validator:   dqf.validateBusinessConsistency,
			Weight:      0.2,
		},
		{
			ID:          "confidence_consistency",
			Name:        "Confidence Consistency",
			Description: "Validates consistency between confidence scores",
			Fields:      []string{"overall_confidence", "accuracy_score", "completeness_score"},
			Validator:   dqf.validateConfidenceConsistency,
			Weight:      0.3,
		},
	}
}

// AssessDataQuality performs comprehensive data quality assessment
func (dqf *DataQualityFramework) AssessDataQuality(
	ctx context.Context,
	data map[string]interface{},
	metadata *DataQualityMetadata,
) (*DataQualityScore, error) {
	ctx, span := dqf.tracer.Start(ctx, "DataQualityFramework.AssessDataQuality")
	defer span.End()

	span.SetAttributes(
		attribute.Int("data_fields", len(data)),
		attribute.String("business_name", dqf.getStringValue(data, "business_name")),
	)

	// Create quality score structure
	qualityScore := &DataQualityScore{
		AssessedAt:             time.Now(),
		DataSources:            []string{"validation", "quality_assessment"},
		ValidationResults:      []ValidationResult{},
		ValidationErrors:       []ValidationError{},
		CrossValidationResults: []CrossValidationResult{},
	}

	// Perform validation
	if err := dqf.performValidation(ctx, data, qualityScore); err != nil {
		dqf.logger.Warn("validation failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Assess accuracy
	if dqf.config.EnableAccuracyScoring {
		dqf.assessAccuracy(ctx, data, qualityScore)
	}

	// Assess completeness
	if dqf.config.EnableCompletenessScoring {
		dqf.assessCompleteness(ctx, data, qualityScore)
	}

	// Assess freshness
	if dqf.config.EnableFreshnessScoring {
		dqf.assessFreshness(ctx, data, metadata, qualityScore)
	}

	// Assess consistency
	if dqf.config.EnableConsistencyScoring {
		dqf.assessConsistency(ctx, data, qualityScore)
	}

	// Perform cross-validation
	if dqf.config.EnableCrossValidation {
		dqf.performCrossValidation(ctx, data, qualityScore)
	}

	// Calculate overall score
	dqf.calculateOverallScore(qualityScore)

	// Determine quality level and grade
	dqf.determineQualityLevel(qualityScore)

	dqf.logger.Info("data quality assessment completed", map[string]interface{}{
		"business_name":     dqf.getStringValue(data, "business_name"),
		"overall_score":     qualityScore.OverallScore,
		"quality_level":     qualityScore.QualityLevel,
		"quality_grade":     qualityScore.QualityGrade,
		"validation_errors": len(qualityScore.ValidationErrors),
	})

	return qualityScore, nil
}

// performValidation performs data validation
func (dqf *DataQualityFramework) performValidation(
	ctx context.Context,
	data map[string]interface{},
	qualityScore *DataQualityScore,
) error {
	ctx, span := dqf.tracer.Start(ctx, "DataQualityFramework.performValidation")
	defer span.End()

	errorCount := 0

	// Validate each field
	for field, rules := range dqf.validationRules {
		value, exists := data[field]

		for _, rule := range rules {
			result := ValidationResult{
				RuleID:   rule.ID,
				RuleName: rule.Name,
				Field:    field,
				Value:    value,
				IsValid:  true,
				Severity: "info",
			}

			// Check if field is required
			if rule.Required && (!exists || value == nil || value == "") {
				result.IsValid = false
				result.Error = "Field is required"
				result.Severity = "error"
				errorCount++
			} else if exists && value != nil && value != "" {
				// Perform type-specific validation
				if err := dqf.validateField(rule, value); err != nil {
					result.IsValid = false
					result.Error = err.Error()
					result.Severity = "error"
					errorCount++
				}
			}

			qualityScore.ValidationResults = append(qualityScore.ValidationResults, result)

			// Add validation error if validation failed
			if !result.IsValid {
				validationError := ValidationError{
					RuleID:   rule.ID,
					Field:    field,
					Value:    value,
					Error:    result.Error,
					Severity: result.Severity,
				}
				qualityScore.ValidationErrors = append(qualityScore.ValidationErrors, validationError)
			}

			// Stop if too many errors
			if errorCount >= dqf.config.MaxValidationErrors {
				return fmt.Errorf("maximum validation errors reached (%d)", dqf.config.MaxValidationErrors)
			}
		}
	}

	return nil
}

// validateField validates a single field based on rule
func (dqf *DataQualityFramework) validateField(rule ValidationRule, value interface{}) error {
	switch rule.Type {
	case "regex":
		if rule.Pattern != "" {
			strValue := fmt.Sprintf("%v", value)
			matched, err := regexp.MatchString(rule.Pattern, strValue)
			if err != nil {
				return fmt.Errorf("regex validation error: %w", err)
			}
			if !matched {
				return fmt.Errorf("value does not match pattern: %s", rule.Pattern)
			}
		}
	case "length":
		strValue := fmt.Sprintf("%v", value)
		length := len(strValue)
		if rule.MinLength > 0 && length < rule.MinLength {
			return fmt.Errorf("length %d is less than minimum %d", length, rule.MinLength)
		}
		if rule.MaxLength > 0 && length > rule.MaxLength {
			return fmt.Errorf("length %d is greater than maximum %d", length, rule.MaxLength)
		}
	case "range":
		if numValue, ok := value.(float64); ok {
			if rule.MinValue > 0 && numValue < rule.MinValue {
				return fmt.Errorf("value %f is less than minimum %f", numValue, rule.MinValue)
			}
			if rule.MaxValue > 0 && numValue > rule.MaxValue {
				return fmt.Errorf("value %f is greater than maximum %f", numValue, rule.MaxValue)
			}
		}
	case "required":
		if value == nil || value == "" {
			return fmt.Errorf("field is required")
		}
	case "custom":
		if rule.CustomValidator != nil {
			return rule.CustomValidator(value)
		}
	}

	return nil
}

// assessAccuracy assesses data accuracy
func (dqf *DataQualityFramework) assessAccuracy(
	ctx context.Context,
	data map[string]interface{},
	qualityScore *DataQualityScore,
) {
	ctx, span := dqf.tracer.Start(ctx, "DataQualityFramework.assessAccuracy")
	defer span.End()

	accuracyMetrics := AccuracyMetrics{
		SupportingEvidence: []string{},
	}

	// Pattern matching accuracy
	patternMatchCount := 0
	totalPatternChecks := 0

	for _, result := range qualityScore.ValidationResults {
		if strings.Contains(result.RuleName, "Format") || strings.Contains(result.RuleName, "Pattern") {
			totalPatternChecks++
			if result.IsValid {
				patternMatchCount++
			}
		}
	}

	if totalPatternChecks > 0 {
		accuracyMetrics.PatternMatchAccuracy = float64(patternMatchCount) / float64(totalPatternChecks)
	}

	// Format validation accuracy
	formatValidationCount := 0
	totalFormatChecks := 0

	for _, result := range qualityScore.ValidationResults {
		if strings.Contains(result.RuleName, "Format") || strings.Contains(result.RuleName, "Length") {
			totalFormatChecks++
			if result.IsValid {
				formatValidationCount++
			}
		}
	}

	if totalFormatChecks > 0 {
		accuracyMetrics.FormatValidationAccuracy = float64(formatValidationCount) / float64(totalFormatChecks)
	}

	// Semantic validation accuracy
	semanticValidationCount := 0
	totalSemanticChecks := 0

	for _, result := range qualityScore.ValidationResults {
		if strings.Contains(result.RuleName, "Required") || strings.Contains(result.RuleName, "Custom") {
			totalSemanticChecks++
			if result.IsValid {
				semanticValidationCount++
			}
		}
	}

	if totalSemanticChecks > 0 {
		accuracyMetrics.SemanticValidationAccuracy = float64(semanticValidationCount) / float64(totalSemanticChecks)
	}

	// Confidence correlation
	if confidence, exists := data["confidence"]; exists {
		if confValue, ok := confidence.(float64); ok {
			accuracyMetrics.ConfidenceCorrelation = confValue
		}
	}

	// Error rate
	totalValidations := len(qualityScore.ValidationResults)
	errorCount := len(qualityScore.ValidationErrors)
	if totalValidations > 0 {
		accuracyMetrics.ErrorRate = float64(errorCount) / float64(totalValidations)
	}

	// Calculate overall accuracy score
	accuracyScore := (accuracyMetrics.PatternMatchAccuracy * 0.4) +
		(accuracyMetrics.FormatValidationAccuracy * 0.3) +
		(accuracyMetrics.SemanticValidationAccuracy * 0.3)

	qualityScore.AccuracyScore = accuracyScore
	qualityScore.AccuracyMetrics = accuracyMetrics

	span.SetAttributes(
		attribute.Float64("accuracy_score", accuracyScore),
		attribute.Float64("pattern_match_accuracy", accuracyMetrics.PatternMatchAccuracy),
		attribute.Float64("format_validation_accuracy", accuracyMetrics.FormatValidationAccuracy),
		attribute.Float64("error_rate", accuracyMetrics.ErrorRate),
	)
}

// assessCompleteness assesses data completeness
func (dqf *DataQualityFramework) assessCompleteness(
	ctx context.Context,
	data map[string]interface{},
	qualityScore *DataQualityScore,
) {
	ctx, span := dqf.tracer.Start(ctx, "DataQualityFramework.assessCompleteness")
	defer span.End()

	completenessMetrics := CompletenessMetrics{
		MissingCriticalData:   []string{},
		PartialDataIndicators: []string{},
	}

	// Required fields completeness
	requiredFields := []string{"business_name"}
	optionalFields := []string{"email", "phone", "address", "website", "description"}

	requiredFieldsPresent := 0
	for _, field := range requiredFields {
		if value, exists := data[field]; exists && value != nil && value != "" {
			requiredFieldsPresent++
		} else {
			completenessMetrics.MissingCriticalData = append(completenessMetrics.MissingCriticalData, field)
		}
	}

	if len(requiredFields) > 0 {
		completenessMetrics.RequiredFieldsCompleteness = float64(requiredFieldsPresent) / float64(len(requiredFields))
	}

	// Optional fields completeness
	optionalFieldsPresent := 0
	for _, field := range optionalFields {
		if value, exists := data[field]; exists && value != nil && value != "" {
			optionalFieldsPresent++
		} else {
			completenessMetrics.PartialDataIndicators = append(completenessMetrics.PartialDataIndicators, field)
		}
	}

	if len(optionalFields) > 0 {
		completenessMetrics.OptionalFieldsCompleteness = float64(optionalFieldsPresent) / float64(len(optionalFields))
	}

	// Data point coverage
	totalFields := len(requiredFields) + len(optionalFields)
	presentFields := requiredFieldsPresent + optionalFieldsPresent
	if totalFields > 0 {
		completenessMetrics.DataPointCoverage = float64(presentFields) / float64(totalFields)
	}

	// Calculate overall completeness score
	completenessScore := (completenessMetrics.RequiredFieldsCompleteness * 0.6) +
		(completenessMetrics.OptionalFieldsCompleteness * 0.4)

	qualityScore.CompletenessScore = completenessScore
	qualityScore.CompletenessMetrics = completenessMetrics

	span.SetAttributes(
		attribute.Float64("completeness_score", completenessScore),
		attribute.Float64("required_fields_completeness", completenessMetrics.RequiredFieldsCompleteness),
		attribute.Float64("optional_fields_completeness", completenessMetrics.OptionalFieldsCompleteness),
		attribute.Float64("data_point_coverage", completenessMetrics.DataPointCoverage),
	)
}

// assessFreshness assesses data freshness
func (dqf *DataQualityFramework) assessFreshness(
	ctx context.Context,
	data map[string]interface{},
	metadata *DataQualityMetadata,
	qualityScore *DataQualityScore,
) {
	ctx, span := dqf.tracer.Start(ctx, "DataQualityFramework.assessFreshness")
	defer span.End()

	freshnessMetrics := FreshnessMetrics{
		StaleDataIndicators: []string{},
	}

	// Data age calculation
	if metadata != nil && !metadata.LastUpdateTime.IsZero() {
		freshnessMetrics.LastUpdateTime = metadata.LastUpdateTime
		freshnessMetrics.DataAge = time.Since(metadata.LastUpdateTime)
	} else {
		freshnessMetrics.DataAge = 24 * time.Hour // Default to 24 hours if no metadata
		freshnessMetrics.StaleDataIndicators = append(freshnessMetrics.StaleDataIndicators, "no_update_time")
	}

	// Freshness score calculation
	if freshnessMetrics.DataAge <= dqf.config.DataFreshnessThreshold {
		freshnessMetrics.FreshnessScore = 1.0
	} else {
		// Apply penalty for stale data
		ageRatio := float64(freshnessMetrics.DataAge) / float64(dqf.config.DataFreshnessThreshold)
		freshnessMetrics.FreshnessScore = 1.0 - (ageRatio * dqf.config.StaleDataPenalty)
		if freshnessMetrics.FreshnessScore < 0 {
			freshnessMetrics.FreshnessScore = 0
		}
	}

	// Update frequency assessment
	if freshnessMetrics.DataAge <= time.Hour {
		freshnessMetrics.UpdateFrequency = "hourly"
	} else if freshnessMetrics.DataAge <= 24*time.Hour {
		freshnessMetrics.UpdateFrequency = "daily"
	} else if freshnessMetrics.DataAge <= 7*24*time.Hour {
		freshnessMetrics.UpdateFrequency = "weekly"
	} else {
		freshnessMetrics.UpdateFrequency = "monthly"
		freshnessMetrics.StaleDataIndicators = append(freshnessMetrics.StaleDataIndicators, "infrequent_updates")
	}

	qualityScore.FreshnessScore = freshnessMetrics.FreshnessScore
	qualityScore.FreshnessMetrics = freshnessMetrics

	span.SetAttributes(
		attribute.Float64("freshness_score", freshnessMetrics.FreshnessScore),
		attribute.String("update_frequency", freshnessMetrics.UpdateFrequency),
		attribute.Int64("data_age_hours", int64(freshnessMetrics.DataAge.Hours())),
	)
}

// assessConsistency assesses data consistency
func (dqf *DataQualityFramework) assessConsistency(
	ctx context.Context,
	data map[string]interface{},
	qualityScore *DataQualityScore,
) {
	ctx, span := dqf.tracer.Start(ctx, "DataQualityFramework.assessConsistency")
	defer span.End()

	consistencyMetrics := ConsistencyMetrics{
		InconsistencyIndicators: []string{},
	}

	// Internal consistency (validation results)
	validValidations := 0
	totalValidations := len(qualityScore.ValidationResults)
	for _, result := range qualityScore.ValidationResults {
		if result.IsValid {
			validValidations++
		}
	}

	if totalValidations > 0 {
		consistencyMetrics.InternalConsistency = float64(validValidations) / float64(totalValidations)
	}

	// Cross-field consistency
	crossFieldConsistency := 0.0
	crossFieldChecks := 0

	// Check email and phone consistency
	if email, emailExists := data["email"]; emailExists && email != "" {
		if phone, phoneExists := data["phone"]; phoneExists && phone != "" {
			crossFieldChecks++
			// Simple consistency check: both contact methods present
			crossFieldConsistency += 1.0
		}
	}

	// Check address consistency
	if address, addressExists := data["address"]; addressExists && address != "" {
		if city, cityExists := data["city"]; cityExists && city != "" {
			crossFieldChecks++
			crossFieldConsistency += 1.0
		}
	}

	if crossFieldChecks > 0 {
		consistencyMetrics.CrossFieldConsistency = crossFieldConsistency / float64(crossFieldChecks)
	}

	// Format consistency
	formatConsistency := 0.0
	formatChecks := 0

	// Check email format consistency
	if email, exists := data["email"]; exists && email != "" {
		formatChecks++
		if dqf.isValidEmail(fmt.Sprintf("%v", email)) {
			formatConsistency += 1.0
		}
	}

	// Check phone format consistency
	if phone, exists := data["phone"]; exists && phone != "" {
		formatChecks++
		if dqf.isValidPhone(fmt.Sprintf("%v", phone)) {
			formatConsistency += 1.0
		}
	}

	if formatChecks > 0 {
		consistencyMetrics.FormatConsistency = formatConsistency / float64(formatChecks)
	}

	// Value consistency
	consistencyMetrics.ValueConsistency = 1.0 // Default to high consistency

	// Calculate overall consistency score
	consistencyScore := (consistencyMetrics.InternalConsistency * 0.4) +
		(consistencyMetrics.CrossFieldConsistency * 0.3) +
		(consistencyMetrics.FormatConsistency * 0.2) +
		(consistencyMetrics.ValueConsistency * 0.1)

	qualityScore.ConsistencyScore = consistencyScore
	qualityScore.ConsistencyMetrics = consistencyMetrics

	span.SetAttributes(
		attribute.Float64("consistency_score", consistencyScore),
		attribute.Float64("internal_consistency", consistencyMetrics.InternalConsistency),
		attribute.Float64("cross_field_consistency", consistencyMetrics.CrossFieldConsistency),
		attribute.Float64("format_consistency", consistencyMetrics.FormatConsistency),
	)
}

// performCrossValidation performs cross-validation
func (dqf *DataQualityFramework) performCrossValidation(
	ctx context.Context,
	data map[string]interface{},
	qualityScore *DataQualityScore,
) {
	ctx, span := dqf.tracer.Start(ctx, "DataQualityFramework.performCrossValidation")
	defer span.End()

	for _, rule := range dqf.crossValidationRules {
		result := CrossValidationResult{
			RuleID:     rule.ID,
			RuleName:   rule.Name,
			Fields:     rule.Fields,
			IsValid:    true,
			Confidence: 1.0,
			Weight:     rule.Weight,
		}

		// Check if all required fields are present
		allFieldsPresent := true
		for _, field := range rule.Fields {
			if value, exists := data[field]; !exists || value == nil || value == "" {
				allFieldsPresent = false
				break
			}
		}

		if allFieldsPresent && rule.Validator != nil {
			if err := rule.Validator(data); err != nil {
				result.IsValid = false
				result.Error = err.Error()
				result.Confidence = 0.5
			}
		}

		qualityScore.CrossValidationResults = append(qualityScore.CrossValidationResults, result)
	}
}

// calculateOverallScore calculates the overall quality score
func (dqf *DataQualityFramework) calculateOverallScore(qualityScore *DataQualityScore) {
	// Weighted average of all scores
	weights := map[string]float64{
		"accuracy":     0.3,
		"completeness": 0.3,
		"freshness":    0.2,
		"consistency":  0.2,
	}

	overallScore := (qualityScore.AccuracyScore * weights["accuracy"]) +
		(qualityScore.CompletenessScore * weights["completeness"]) +
		(qualityScore.FreshnessScore * weights["freshness"]) +
		(qualityScore.ConsistencyScore * weights["consistency"])

	qualityScore.OverallScore = overallScore
}

// determineQualityLevel determines the quality level and grade
func (dqf *DataQualityFramework) determineQualityLevel(qualityScore *DataQualityScore) {
	score := qualityScore.OverallScore

	switch {
	case score >= 0.9:
		qualityScore.QualityLevel = "excellent"
		qualityScore.QualityGrade = "A"
	case score >= 0.8:
		qualityScore.QualityLevel = "good"
		qualityScore.QualityGrade = "B"
	case score >= 0.7:
		qualityScore.QualityLevel = "fair"
		qualityScore.QualityGrade = "C"
	case score >= 0.6:
		qualityScore.QualityLevel = "poor"
		qualityScore.QualityGrade = "D"
	default:
		qualityScore.QualityLevel = "very_poor"
		qualityScore.QualityGrade = "F"
	}
}

// Cross-validation helper functions
func (dqf *DataQualityFramework) validateContactConsistency(data map[string]interface{}) error {
	email, emailExists := data["email"]
	phone, phoneExists := data["phone"]

	if emailExists && phoneExists && email != "" && phone != "" {
		// Both contact methods present - good consistency
		return nil
	}

	if !emailExists && !phoneExists {
		return fmt.Errorf("no contact information provided")
	}

	return fmt.Errorf("incomplete contact information")
}

func (dqf *DataQualityFramework) validateAddressConsistency(data map[string]interface{}) error {
	address, addressExists := data["address"]
	city, cityExists := data["city"]

	if addressExists && cityExists && address != "" && city != "" {
		// Address components present - good consistency
		return nil
	}

	return fmt.Errorf("incomplete address information")
}

func (dqf *DataQualityFramework) validateBusinessConsistency(data map[string]interface{}) error {
	businessName, nameExists := data["business_name"]
	website, websiteExists := data["website"]

	if nameExists && businessName != "" {
		// Business name is required and present
		if websiteExists && website != "" {
			// Both name and website present - good consistency
			return nil
		}
		// Name present but no website - acceptable
		return nil
	}

	return fmt.Errorf("missing business name")
}

func (dqf *DataQualityFramework) validateConfidenceConsistency(data map[string]interface{}) error {
	overallConfidence, overallExists := data["overall_confidence"]
	accuracyScore, accuracyExists := data["accuracy_score"]

	if overallExists && accuracyExists {
		overall := dqf.getFloatValue(overallConfidence)
		accuracy := dqf.getFloatValue(accuracyScore)

		// Check if overall confidence is reasonable given accuracy
		if overall > accuracy+0.3 {
			return fmt.Errorf("overall confidence significantly higher than accuracy")
		}
	}

	return nil
}

// Utility functions
func (dqf *DataQualityFramework) getStringValue(data map[string]interface{}, key string) string {
	if value, exists := data[key]; exists && value != nil {
		return fmt.Sprintf("%v", value)
	}
	return ""
}

func (dqf *DataQualityFramework) getFloatValue(value interface{}) float64 {
	if value == nil {
		return 0.0
	}

	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		// Try to parse as float
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0.0
}

func (dqf *DataQualityFramework) isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

func (dqf *DataQualityFramework) isValidPhone(phone string) bool {
	pattern := `^(\+\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

// DataQualityMetadata represents metadata for quality assessment
type DataQualityMetadata struct {
	LastUpdateTime time.Time     `json:"last_update_time"`
	SourceSystem   string        `json:"source_system"`
	DataVersion    string        `json:"data_version"`
	ProcessingTime time.Duration `json:"processing_time"`
}
