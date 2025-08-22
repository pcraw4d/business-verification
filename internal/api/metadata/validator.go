package metadata

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
)

// MetadataValidator provides validation and consistency checking for metadata
type MetadataValidator struct {
	logger *zap.Logger
	config *ValidationConfig
}

// ValidationConfig contains configuration for metadata validation
type ValidationConfig struct {
	// Validation rules
	RequiredFields []string              `json:"required_fields"`
	FieldTypes     map[string]string     `json:"field_types"`
	FieldRanges    map[string]FieldRange `json:"field_ranges"`

	// Consistency checks
	EnableConsistencyChecks bool `json:"enable_consistency_checks"`
	EnableCrossValidation   bool `json:"enable_cross_validation"`

	// Validation thresholds
	MinConfidenceScore float64       `json:"min_confidence_score"`
	MaxProcessingTime  time.Duration `json:"max_processing_time"`

	// Error handling
	StrictMode bool `json:"strict_mode"`
	MaxErrors  int  `json:"max_errors"`
}

// FieldRange represents valid range for a field
type FieldRange struct {
	Min  interface{} `json:"min"`
	Max  interface{} `json:"max"`
	Type string      `json:"type"`
}

// ValidationResult represents the result of metadata validation
type ValidationResult struct {
	IsValid           bool                   `json:"is_valid"`
	ValidationScore   float64                `json:"validation_score"`
	Errors            []ValidationError      `json:"errors"`
	Warnings          []ValidationWarning    `json:"warnings"`
	ConsistencyChecks []ConsistencyCheck     `json:"consistency_checks"`
	ValidatedAt       time.Time              `json:"validated_at"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field      string `json:"field"`
	Message    string `json:"message"`
	Severity   string `json:"severity"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ConsistencyCheck represents a consistency check result
type ConsistencyCheck struct {
	CheckID     string                 `json:"check_id"`
	CheckName   string                 `json:"check_name"`
	Status      string                 `json:"status"` // "passed", "failed", "warning"
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// NewMetadataValidator creates a new metadata validator
func NewMetadataValidator(logger *zap.Logger, config *ValidationConfig) *MetadataValidator {
	if config == nil {
		config = getDefaultValidationConfig()
	}

	return &MetadataValidator{
		logger: logger,
		config: config,
	}
}

// ValidateMetadata validates response metadata
func (mv *MetadataValidator) ValidateMetadata(ctx context.Context, metadata *ResponseMetadata) (*ValidationResult, error) {
	if metadata == nil {
		return nil, fmt.Errorf("metadata cannot be nil")
	}

	result := &ValidationResult{
		IsValid:           true,
		ValidationScore:   1.0,
		Errors:            []ValidationError{},
		Warnings:          []ValidationWarning{},
		ConsistencyChecks: []ConsistencyCheck{},
		ValidatedAt:       time.Now(),
		Metadata:          make(map[string]interface{}),
	}

	// Validate required fields
	mv.validateRequiredFields(metadata, result)

	// Validate field types
	mv.validateFieldTypes(metadata, result)

	// Validate field ranges
	mv.validateFieldRanges(metadata, result)

	// Perform consistency checks
	if mv.config.EnableConsistencyChecks {
		mv.performConsistencyChecks(metadata, result)
	}

	// Perform cross-validation
	if mv.config.EnableCrossValidation {
		mv.performCrossValidation(metadata, result)
	}

	// Calculate validation score
	mv.calculateValidationScore(result)

	// Determine overall validity
	result.IsValid = len(result.Errors) == 0

	mv.logger.Debug("Metadata validation completed",
		zap.String("request_id", metadata.RequestID),
		zap.Bool("is_valid", result.IsValid),
		zap.Float64("validation_score", result.ValidationScore),
		zap.Int("error_count", len(result.Errors)),
		zap.Int("warning_count", len(result.Warnings)))

	return result, nil
}

// ValidateDataSource validates data source metadata
func (mv *MetadataValidator) ValidateDataSource(ctx context.Context, source *DataSourceMetadata) (*ValidationResult, error) {
	if source == nil {
		return nil, fmt.Errorf("data source cannot be nil")
	}

	result := &ValidationResult{
		IsValid:           true,
		ValidationScore:   1.0,
		Errors:            []ValidationError{},
		Warnings:          []ValidationWarning{},
		ConsistencyChecks: []ConsistencyCheck{},
		ValidatedAt:       time.Now(),
		Metadata:          make(map[string]interface{}),
	}

	// Validate required fields
	if source.SourceID == "" {
		result.Errors = append(result.Errors, ValidationError{
			ErrorCode:    "MISSING_REQUIRED_FIELD",
			ErrorMessage: "Source ID is required",
			Field:        "source_id",
			Severity:     "high",
		})
	}

	if source.SourceName == "" {
		result.Errors = append(result.Errors, ValidationError{
			ErrorCode:    "MISSING_REQUIRED_FIELD",
			ErrorMessage: "Source name is required",
			Field:        "source_name",
			Severity:     "high",
		})
	}

	if source.SourceType == "" {
		result.Errors = append(result.Errors, ValidationError{
			ErrorCode:    "MISSING_REQUIRED_FIELD",
			ErrorMessage: "Source type is required",
			Field:        "source_type",
			Severity:     "high",
		})
	}

	// Validate reliability score
	if source.ReliabilityScore < 0.0 || source.ReliabilityScore > 1.0 {
		result.Errors = append(result.Errors, ValidationError{
			ErrorCode:    "INVALID_RANGE",
			ErrorMessage: "Reliability score must be between 0.0 and 1.0",
			Field:        "reliability_score",
			Severity:     "medium",
		})
	}

	// Validate timestamps
	if source.LastUpdated.IsZero() {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:    "last_updated",
			Message:  "Last updated timestamp is not set",
			Severity: "low",
		})
	}

	// Calculate validation score
	mv.calculateValidationScore(result)

	// Determine overall validity
	result.IsValid = len(result.Errors) == 0

	return result, nil
}

// ValidateConfidence validates confidence metadata
func (mv *MetadataValidator) ValidateConfidence(ctx context.Context, confidence *ConfidenceMetadata) (*ValidationResult, error) {
	if confidence == nil {
		return nil, fmt.Errorf("confidence metadata cannot be nil")
	}

	result := &ValidationResult{
		IsValid:           true,
		ValidationScore:   1.0,
		Errors:            []ValidationError{},
		Warnings:          []ValidationWarning{},
		ConsistencyChecks: []ConsistencyCheck{},
		ValidatedAt:       time.Now(),
		Metadata:          make(map[string]interface{}),
	}

	// Validate overall confidence
	if confidence.OverallConfidence < 0.0 || confidence.OverallConfidence > 1.0 {
		result.Errors = append(result.Errors, ValidationError{
			ErrorCode:    "INVALID_RANGE",
			ErrorMessage: "Overall confidence must be between 0.0 and 1.0",
			Field:        "overall_confidence",
			Severity:     "high",
		})
	}

	// Check minimum confidence threshold
	if confidence.OverallConfidence < mv.config.MinConfidenceScore {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      "overall_confidence",
			Message:    fmt.Sprintf("Confidence score %.2f is below recommended threshold %.2f", confidence.OverallConfidence, mv.config.MinConfidenceScore),
			Severity:   "medium",
			Suggestion: "Consider improving data quality or using additional data sources",
		})
	}

	// Validate component scores
	for factorName, score := range confidence.ComponentScores {
		if score < 0.0 || score > 1.0 {
			result.Errors = append(result.Errors, ValidationError{
				ErrorCode:    "INVALID_RANGE",
				ErrorMessage: fmt.Sprintf("Component score for %s must be between 0.0 and 1.0", factorName),
				Field:        fmt.Sprintf("component_scores.%s", factorName),
				Severity:     "medium",
			})
		}
	}

	// Validate factors
	for i, factor := range confidence.Factors {
		if factor.FactorValue < 0.0 || factor.FactorValue > 1.0 {
			result.Errors = append(result.Errors, ValidationError{
				ErrorCode:    "INVALID_RANGE",
				ErrorMessage: fmt.Sprintf("Factor value for %s must be between 0.0 and 1.0", factor.FactorName),
				Field:        fmt.Sprintf("factors[%d].factor_value", i),
				Severity:     "medium",
			})
		}

		if factor.FactorWeight < 0.0 || factor.FactorWeight > 1.0 {
			result.Errors = append(result.Errors, ValidationError{
				ErrorCode:    "INVALID_RANGE",
				ErrorMessage: fmt.Sprintf("Factor weight for %s must be between 0.0 and 1.0", factor.FactorName),
				Field:        fmt.Sprintf("factors[%d].factor_weight", i),
				Severity:     "medium",
			})
		}
	}

	// Validate uncertainty metrics
	if confidence.Uncertainty != nil {
		mv.validateUncertaintyMetrics(confidence.Uncertainty, result)
	}

	// Calculate validation score
	mv.calculateValidationScore(result)

	// Determine overall validity
	result.IsValid = len(result.Errors) == 0

	return result, nil
}

// Helper methods

func (mv *MetadataValidator) validateRequiredFields(metadata *ResponseMetadata, result *ValidationResult) {
	// Check required fields
	if metadata.RequestID == "" {
		result.Errors = append(result.Errors, ValidationError{
			ErrorCode:    "MISSING_REQUIRED_FIELD",
			ErrorMessage: "Request ID is required",
			Field:        "request_id",
			Severity:     "high",
		})
	}

	if metadata.Timestamp.IsZero() {
		result.Errors = append(result.Errors, ValidationError{
			ErrorCode:    "MISSING_REQUIRED_FIELD",
			ErrorMessage: "Timestamp is required",
			Field:        "timestamp",
			Severity:     "high",
		})
	}

	if metadata.APIVersion == "" {
		result.Errors = append(result.Errors, ValidationError{
			ErrorCode:    "MISSING_REQUIRED_FIELD",
			ErrorMessage: "API version is required",
			Field:        "api_version",
			Severity:     "high",
		})
	}
}

func (mv *MetadataValidator) validateFieldTypes(metadata *ResponseMetadata, result *ValidationResult) {
	// Validate processing time
	if metadata.ProcessingTime < 0 {
		result.Errors = append(result.Errors, ValidationError{
			ErrorCode:    "INVALID_TYPE",
			ErrorMessage: "Processing time cannot be negative",
			Field:        "processing_time",
			Severity:     "medium",
		})
	}

	// Check maximum processing time
	if metadata.ProcessingTime > mv.config.MaxProcessingTime {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      "processing_time",
			Message:    fmt.Sprintf("Processing time %v exceeds recommended maximum %v", metadata.ProcessingTime, mv.config.MaxProcessingTime),
			Severity:   "medium",
			Suggestion: "Consider optimizing processing performance",
		})
	}
}

func (mv *MetadataValidator) validateFieldRanges(metadata *ResponseMetadata, result *ValidationResult) {
	// Validate data sources
	for i, source := range metadata.DataSources {
		if source.ReliabilityScore < 0.0 || source.ReliabilityScore > 1.0 {
			result.Errors = append(result.Errors, ValidationError{
				ErrorCode:    "INVALID_RANGE",
				ErrorMessage: fmt.Sprintf("Reliability score for data source %s must be between 0.0 and 1.0", source.SourceID),
				Field:        fmt.Sprintf("data_sources[%d].reliability_score", i),
				Severity:     "medium",
			})
		}
	}

	// Validate confidence
	if metadata.Confidence != nil {
		if metadata.Confidence.OverallConfidence < 0.0 || metadata.Confidence.OverallConfidence > 1.0 {
			result.Errors = append(result.Errors, ValidationError{
				ErrorCode:    "INVALID_RANGE",
				ErrorMessage: "Overall confidence must be between 0.0 and 1.0",
				Field:        "confidence.overall_confidence",
				Severity:     "high",
			})
		}
	}

	// Validate quality
	if metadata.Quality != nil {
		if metadata.Quality.OverallQuality < 0.0 || metadata.Quality.OverallQuality > 1.0 {
			result.Errors = append(result.Errors, ValidationError{
				ErrorCode:    "INVALID_RANGE",
				ErrorMessage: "Overall quality must be between 0.0 and 1.0",
				Field:        "quality.overall_quality",
				Severity:     "medium",
			})
		}
	}
}

func (mv *MetadataValidator) performConsistencyChecks(metadata *ResponseMetadata, result *ValidationResult) {
	// Check consistency between confidence and quality
	if metadata.Confidence != nil && metadata.Quality != nil {
		confidenceDiff := math.Abs(metadata.Confidence.OverallConfidence - metadata.Quality.OverallQuality)
		if confidenceDiff > 0.3 {
			result.ConsistencyChecks = append(result.ConsistencyChecks, ConsistencyCheck{
				CheckID:     "confidence_quality_consistency",
				CheckName:   "Confidence-Quality Consistency",
				Status:      "warning",
				Description: fmt.Sprintf("Large difference between confidence (%.2f) and quality (%.2f) scores", metadata.Confidence.OverallConfidence, metadata.Quality.OverallQuality),
				Confidence:  0.7,
				Details: map[string]interface{}{
					"confidence_score": metadata.Confidence.OverallConfidence,
					"quality_score":    metadata.Quality.OverallQuality,
					"difference":       confidenceDiff,
				},
			})
		}
	}

	// Check data source consistency
	if len(metadata.DataSources) > 1 {
		var totalReliability float64
		for _, source := range metadata.DataSources {
			totalReliability += source.ReliabilityScore
		}
		avgReliability := totalReliability / float64(len(metadata.DataSources))

		// Check for significant variations in reliability scores
		var variance float64
		for _, source := range metadata.DataSources {
			variance += math.Pow(source.ReliabilityScore-avgReliability, 2)
		}
		variance /= float64(len(metadata.DataSources))
		stdDev := math.Sqrt(variance)

		if stdDev > 0.2 {
			result.ConsistencyChecks = append(result.ConsistencyChecks, ConsistencyCheck{
				CheckID:     "data_source_reliability_consistency",
				CheckName:   "Data Source Reliability Consistency",
				Status:      "warning",
				Description: "High variation in data source reliability scores",
				Confidence:  0.8,
				Details: map[string]interface{}{
					"average_reliability": avgReliability,
					"standard_deviation":  stdDev,
					"source_count":        len(metadata.DataSources),
				},
			})
		}
	}
}

func (mv *MetadataValidator) performCrossValidation(metadata *ResponseMetadata, result *ValidationResult) {
	// Cross-validate data sources with confidence
	if metadata.Confidence != nil && len(metadata.DataSources) > 0 {
		var totalSourceReliability float64
		for _, source := range metadata.DataSources {
			totalSourceReliability += source.ReliabilityScore
		}
		avgSourceReliability := totalSourceReliability / float64(len(metadata.DataSources))

		// Check if confidence aligns with source reliability
		confidenceDiff := math.Abs(metadata.Confidence.OverallConfidence - avgSourceReliability)
		if confidenceDiff > 0.25 {
			result.ConsistencyChecks = append(result.ConsistencyChecks, ConsistencyCheck{
				CheckID:     "confidence_source_cross_validation",
				CheckName:   "Confidence-Source Cross Validation",
				Status:      "warning",
				Description: "Confidence score may not align with data source reliability",
				Confidence:  0.6,
				Details: map[string]interface{}{
					"confidence_score":           metadata.Confidence.OverallConfidence,
					"average_source_reliability": avgSourceReliability,
					"difference":                 confidenceDiff,
				},
			})
		}
	}
}

func (mv *MetadataValidator) validateUncertaintyMetrics(uncertainty *UncertaintyMetrics, result *ValidationResult) {
	// Validate uncertainty level
	if uncertainty.UncertaintyLevel < 0.0 || uncertainty.UncertaintyLevel > 1.0 {
		result.Errors = append(result.Errors, ValidationError{
			ErrorCode:    "INVALID_RANGE",
			ErrorMessage: "Uncertainty level must be between 0.0 and 1.0",
			Field:        "uncertainty.uncertainty_level",
			Severity:     "medium",
		})
	}

	// Validate confidence interval
	if uncertainty.ConfidenceInterval.LowerBound < 0.0 || uncertainty.ConfidenceInterval.UpperBound > 1.0 {
		result.Errors = append(result.Errors, ValidationError{
			ErrorCode:    "INVALID_RANGE",
			ErrorMessage: "Confidence interval bounds must be between 0.0 and 1.0",
			Field:        "uncertainty.confidence_interval",
			Severity:     "medium",
		})
	}

	if uncertainty.ConfidenceInterval.LowerBound >= uncertainty.ConfidenceInterval.UpperBound {
		result.Errors = append(result.Errors, ValidationError{
			ErrorCode:    "INVALID_RANGE",
			ErrorMessage: "Lower bound must be less than upper bound",
			Field:        "uncertainty.confidence_interval",
			Severity:     "medium",
		})
	}

	// Validate standard error
	if uncertainty.StandardError < 0.0 {
		result.Errors = append(result.Errors, ValidationError{
			ErrorCode:    "INVALID_RANGE",
			ErrorMessage: "Standard error cannot be negative",
			Field:        "uncertainty.standard_error",
			Severity:     "medium",
		})
	}

	// Validate variance
	if uncertainty.Variance < 0.0 {
		result.Errors = append(result.Errors, ValidationError{
			ErrorCode:    "INVALID_RANGE",
			ErrorMessage: "Variance cannot be negative",
			Field:        "uncertainty.variance",
			Severity:     "medium",
		})
	}
}

func (mv *MetadataValidator) calculateValidationScore(result *ValidationResult) {
	// Start with perfect score
	score := 1.0

	// Reduce score for errors
	errorPenalty := float64(len(result.Errors)) * 0.1
	score -= errorPenalty

	// Reduce score for warnings
	warningPenalty := float64(len(result.Warnings)) * 0.05
	score -= warningPenalty

	// Reduce score for failed consistency checks
	for _, check := range result.ConsistencyChecks {
		if check.Status == "failed" {
			score -= 0.1
		} else if check.Status == "warning" {
			score -= 0.05
		}
	}

	// Ensure score is within valid range
	score = math.Max(0.0, math.Min(1.0, score))

	result.ValidationScore = score
}

func getDefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		RequiredFields: []string{
			"request_id",
			"timestamp",
			"api_version",
		},
		FieldTypes: map[string]string{
			"request_id":      "string",
			"timestamp":       "time.Time",
			"processing_time": "time.Duration",
			"api_version":     "string",
		},
		FieldRanges: map[string]FieldRange{
			"overall_confidence": {
				Min:  0.0,
				Max:  1.0,
				Type: "float64",
			},
			"reliability_score": {
				Min:  0.0,
				Max:  1.0,
				Type: "float64",
			},
			"overall_quality": {
				Min:  0.0,
				Max:  1.0,
				Type: "float64",
			},
		},
		EnableConsistencyChecks: true,
		EnableCrossValidation:   true,
		MinConfidenceScore:      0.5,
		MaxProcessingTime:       30 * time.Second,
		StrictMode:              false,
		MaxErrors:               100,
	}
}
