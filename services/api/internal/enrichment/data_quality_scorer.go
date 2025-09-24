package enrichment

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// DataQualityScorer provides comprehensive data quality scoring and validation
type DataQualityScorer struct {
	logger *zap.Logger
	tracer trace.Tracer
	config *DataQualityConfig
}

// DataQualityConfig contains configuration for data quality scoring
type DataQualityConfig struct {
	// Scoring weights
	CompletenessWeight float64 `json:"completeness_weight"` // Weight for data completeness
	AccuracyWeight     float64 `json:"accuracy_weight"`     // Weight for data accuracy
	ConsistencyWeight  float64 `json:"consistency_weight"`  // Weight for data consistency
	FreshnessWeight    float64 `json:"freshness_weight"`    // Weight for data freshness
	ReliabilityWeight  float64 `json:"reliability_weight"`  // Weight for source reliability
	ValidityWeight     float64 `json:"validity_weight"`     // Weight for data validity

	// Quality thresholds
	MinQualityThreshold   float64 `json:"min_quality_threshold"`     // Minimum acceptable quality score
	HighQualityThreshold  float64 `json:"high_quality_threshold"`    // Threshold for high quality data
	CompletenessThreshold float64 `json:"completeness_threshold"`    // Minimum completeness required
	AccuracyThreshold     float64 `json:"accuracy_threshold"`        // Minimum accuracy required
	FreshnessThreshold    int     `json:"freshness_threshold_hours"` // Hours before data is considered stale

	// Validation settings
	EnableValidation       bool `json:"enable_validation"`        // Enable field validation
	EnableConsistencyCheck bool `json:"enable_consistency_check"` // Enable cross-field consistency checks
	EnableFreshnessCheck   bool `json:"enable_freshness_check"`   // Enable freshness validation
	EnableOutlierDetection bool `json:"enable_outlier_detection"` // Enable outlier detection

	// Field weights for completeness scoring
	RequiredFieldWeights map[string]float64 `json:"required_field_weights"`
	OptionalFieldWeights map[string]float64 `json:"optional_field_weights"`
}

// DataQualityResult contains comprehensive data quality assessment
type DataQualityResult struct {
	// Overall quality metrics
	OverallScore   float64       `json:"overall_score"`   // Overall quality score (0-1)
	QualityLevel   string        `json:"quality_level"`   // "low", "medium", "high", "excellent"
	IsAcceptable   bool          `json:"is_acceptable"`   // Whether data meets minimum quality standards
	AssessedAt     time.Time     `json:"assessed_at"`     // When assessment was performed
	ProcessingTime time.Duration `json:"processing_time"` // Time taken for assessment

	// Component scores
	CompletenessScore float64 `json:"completeness_score"` // Data completeness (0-1)
	AccuracyScore     float64 `json:"accuracy_score"`     // Data accuracy (0-1)
	ConsistencyScore  float64 `json:"consistency_score"`  // Data consistency (0-1)
	FreshnessScore    float64 `json:"freshness_score"`    // Data freshness (0-1)
	ReliabilityScore  float64 `json:"reliability_score"`  // Source reliability (0-1)
	ValidityScore     float64 `json:"validity_score"`     // Data validity (0-1)

	// Detailed assessment
	CompletenessBreakdown map[string]float64 `json:"completeness_breakdown"` // Per-field completeness
	AccuracyBreakdown     map[string]float64 `json:"accuracy_breakdown"`     // Per-field accuracy
	ValidationResults     []ValidationResult `json:"validation_results"`     // Field validation results
	ConsistencyIssues     []ConsistencyIssue `json:"consistency_issues"`     // Detected inconsistencies
	QualityIssues         []QualityIssue     `json:"quality_issues"`         // Quality problems found
	Recommendations       []string           `json:"recommendations"`        // Improvement recommendations

	// Metadata
	DataType         string                 `json:"data_type"`         // Type of data assessed
	TotalFields      int                    `json:"total_fields"`      // Total number of fields
	PopulatedFields  int                    `json:"populated_fields"`  // Number of populated fields
	RequiredFields   int                    `json:"required_fields"`   // Number of required fields
	OptionalFields   int                    `json:"optional_fields"`   // Number of optional fields
	ValidationErrors int                    `json:"validation_errors"` // Number of validation errors
	DataSourceInfo   map[string]interface{} `json:"data_source_info"`  // Information about data sources
}

// ValidationResult represents field-level validation result
type ValidationResult struct {
	FieldName    string   `json:"field_name"`    // Name of the field
	IsValid      bool     `json:"is_valid"`      // Whether field value is valid
	Score        float64  `json:"score"`         // Validation score (0-1)
	ErrorType    string   `json:"error_type"`    // Type of validation error
	ErrorMessage string   `json:"error_message"` // Detailed error message
	Suggestions  []string `json:"suggestions"`   // Suggestions for improvement
}

// ConsistencyIssue represents a data consistency problem
type ConsistencyIssue struct {
	Type        string   `json:"type"`        // Type of consistency issue
	Description string   `json:"description"` // Description of the issue
	Fields      []string `json:"fields"`      // Fields involved in the issue
	Severity    string   `json:"severity"`    // "low", "medium", "high", "critical"
	Impact      float64  `json:"impact"`      // Impact on overall quality (0-1)
}

// QualityIssue represents a general data quality problem
type QualityIssue struct {
	Category    string  `json:"category"`    // Category of quality issue
	Description string  `json:"description"` // Description of the issue
	Field       string  `json:"field"`       // Field affected (if applicable)
	Severity    string  `json:"severity"`    // Severity level
	Impact      float64 `json:"impact"`      // Impact on quality score
	Fixable     bool    `json:"fixable"`     // Whether issue can be automatically fixed
}

// NewDataQualityScorer creates a new data quality scorer
func NewDataQualityScorer(logger *zap.Logger, config *DataQualityConfig) *DataQualityScorer {
	if config == nil {
		config = getDefaultDataQualityConfig()
	}

	return &DataQualityScorer{
		logger: logger,
		tracer: trace.NewNoopTracerProvider().Tracer("data_quality_scorer"),
		config: config,
	}
}

// AssessDataQuality performs comprehensive data quality assessment
func (dqs *DataQualityScorer) AssessDataQuality(ctx context.Context, data interface{}, dataType string) (*DataQualityResult, error) {
	ctx, span := dqs.tracer.Start(ctx, "data_quality_scorer.assess",
		trace.WithAttributes(
			attribute.String("data_type", dataType),
		))
	defer span.End()

	startTime := time.Now()

	dqs.logger.Info("Starting data quality assessment",
		zap.String("data_type", dataType))

	result := &DataQualityResult{
		AssessedAt:            time.Now(),
		DataType:              dataType,
		CompletenessBreakdown: make(map[string]float64),
		AccuracyBreakdown:     make(map[string]float64),
		ValidationResults:     []ValidationResult{},
		ConsistencyIssues:     []ConsistencyIssue{},
		QualityIssues:         []QualityIssue{},
		Recommendations:       []string{},
		DataSourceInfo:        make(map[string]interface{}),
	}

	// Analyze data structure
	dqs.analyzeDataStructure(data, result)

	// Calculate completeness score
	result.CompletenessScore = dqs.calculateCompletenessScore(data, result)

	// Calculate accuracy score
	result.AccuracyScore = dqs.calculateAccuracyScore(data, result)

	// Calculate consistency score
	if dqs.config.EnableConsistencyCheck {
		result.ConsistencyScore = dqs.calculateConsistencyScore(data, result)
	} else {
		result.ConsistencyScore = 1.0 // Assume consistent if not checking
	}

	// Calculate freshness score
	if dqs.config.EnableFreshnessCheck {
		result.FreshnessScore = dqs.calculateFreshnessScore(data, result)
	} else {
		result.FreshnessScore = 1.0 // Assume fresh if not checking
	}

	// Calculate reliability score
	result.ReliabilityScore = dqs.calculateReliabilityScore(data, result)

	// Calculate validity score
	if dqs.config.EnableValidation {
		result.ValidityScore = dqs.calculateValidityScore(data, result)
	} else {
		result.ValidityScore = 1.0 // Assume valid if not validating
	}

	// Calculate overall score
	result.OverallScore = dqs.calculateOverallScore(result)

	// Determine quality level
	result.QualityLevel = dqs.determineQualityLevel(result.OverallScore)

	// Check if acceptable
	result.IsAcceptable = result.OverallScore >= dqs.config.MinQualityThreshold

	// Generate recommendations
	result.Recommendations = dqs.generateRecommendations(result)

	// Detect outliers if enabled
	if dqs.config.EnableOutlierDetection {
		dqs.detectOutliers(data, result)
	}

	result.ProcessingTime = time.Since(startTime)

	dqs.logger.Info("Data quality assessment completed",
		zap.String("data_type", dataType),
		zap.Float64("overall_score", result.OverallScore),
		zap.String("quality_level", result.QualityLevel),
		zap.Bool("is_acceptable", result.IsAcceptable),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// analyzeDataStructure analyzes the structure of the input data
func (dqs *DataQualityScorer) analyzeDataStructure(data interface{}, result *DataQualityResult) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		dqs.logger.Warn("Data is not a struct, limited analysis available")
		return
	}

	t := v.Type()
	totalFields := 0
	populatedFields := 0
	requiredFields := 0
	optionalFields := 0

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		fieldName := fieldType.Name

		totalFields++

		// Check if field is populated
		isPopulated := !field.IsZero()
		if isPopulated {
			populatedFields++
		}

		// Determine if field is required (simplified logic)
		isRequired := dqs.isRequiredField(fieldName, fieldType)
		if isRequired {
			requiredFields++
		} else {
			optionalFields++
		}

		// Store field-level completeness
		if isPopulated {
			result.CompletenessBreakdown[fieldName] = 1.0
		} else {
			result.CompletenessBreakdown[fieldName] = 0.0
		}
	}

	result.TotalFields = totalFields
	result.PopulatedFields = populatedFields
	result.RequiredFields = requiredFields
	result.OptionalFields = optionalFields
}

// calculateCompletenessScore calculates data completeness score
func (dqs *DataQualityScorer) calculateCompletenessScore(data interface{}, result *DataQualityResult) float64 {
	if result.TotalFields == 0 {
		return 0.0
	}

	// Weighted completeness calculation
	totalWeight := 0.0
	achievedWeight := 0.0

	for fieldName, completeness := range result.CompletenessBreakdown {
		weight := dqs.getFieldWeight(fieldName)
		totalWeight += weight
		achievedWeight += completeness * weight
	}

	if totalWeight == 0 {
		// Fallback to simple ratio
		return float64(result.PopulatedFields) / float64(result.TotalFields)
	}

	return achievedWeight / totalWeight
}

// calculateAccuracyScore calculates data accuracy score
func (dqs *DataQualityScorer) calculateAccuracyScore(data interface{}, result *DataQualityResult) float64 {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return 0.5 // Default score for non-struct data
	}

	totalScore := 0.0
	fieldCount := 0

	// Look for confidence scores in the data
	confidenceScore := dqs.extractConfidenceScore(v)
	if confidenceScore > 0 {
		totalScore += confidenceScore
		fieldCount++
	}

	// Validate field formats and values
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		fieldName := fieldType.Name

		if field.IsZero() {
			continue // Skip empty fields
		}

		accuracy := dqs.validateFieldAccuracy(field, fieldName, fieldType)
		result.AccuracyBreakdown[fieldName] = accuracy
		totalScore += accuracy
		fieldCount++
	}

	if fieldCount == 0 {
		return 0.5 // Default score when no fields to evaluate
	}

	return totalScore / float64(fieldCount)
}

// calculateConsistencyScore calculates data consistency score
func (dqs *DataQualityScorer) calculateConsistencyScore(data interface{}, result *DataQualityResult) float64 {
	// Check for internal consistency issues
	consistencyScore := 1.0
	issueCount := 0

	// Example consistency checks (would be expanded based on data type)
	issues := dqs.detectConsistencyIssues(data)
	for _, issue := range issues {
		result.ConsistencyIssues = append(result.ConsistencyIssues, issue)
		consistencyScore -= issue.Impact
		issueCount++
	}

	// Ensure score doesn't go below 0
	return math.Max(0.0, consistencyScore)
}

// calculateFreshnessScore calculates data freshness score
func (dqs *DataQualityScorer) calculateFreshnessScore(data interface{}, result *DataQualityResult) float64 {
	// Extract timestamp information from data
	timestamp := dqs.extractTimestamp(data)
	if timestamp.IsZero() {
		return 0.5 // Default score when no timestamp available
	}

	// Calculate age in hours
	ageHours := time.Since(timestamp).Hours()
	thresholdHours := float64(dqs.config.FreshnessThreshold)

	if ageHours <= thresholdHours {
		return 1.0 // Fresh data
	}

	// Exponential decay for older data
	decay := math.Exp(-ageHours / thresholdHours)
	return math.Max(0.1, decay) // Minimum score of 0.1
}

// calculateReliabilityScore calculates source reliability score
func (dqs *DataQualityScorer) calculateReliabilityScore(data interface{}, result *DataQualityResult) float64 {
	// Extract source information
	source := dqs.extractSourceInfo(data)

	// Score based on source type
	switch source {
	case "website_content":
		return 0.8
	case "api_response":
		return 0.9
	case "user_input":
		return 0.6
	case "external_database":
		return 0.95
	default:
		return 0.7 // Default reliability
	}
}

// calculateValidityScore calculates data validity score
func (dqs *DataQualityScorer) calculateValidityScore(data interface{}, result *DataQualityResult) float64 {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return 0.5
	}

	totalScore := 0.0
	fieldCount := 0
	errorCount := 0

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		fieldName := fieldType.Name

		if field.IsZero() {
			continue
		}

		validationResult := dqs.validateField(field, fieldName, fieldType)
		result.ValidationResults = append(result.ValidationResults, validationResult)

		totalScore += validationResult.Score
		fieldCount++

		if !validationResult.IsValid {
			errorCount++
		}
	}

	result.ValidationErrors = errorCount

	if fieldCount == 0 {
		return 1.0
	}

	return totalScore / float64(fieldCount)
}

// calculateOverallScore calculates the weighted overall quality score
func (dqs *DataQualityScorer) calculateOverallScore(result *DataQualityResult) float64 {
	totalWeight := dqs.config.CompletenessWeight + dqs.config.AccuracyWeight +
		dqs.config.ConsistencyWeight + dqs.config.FreshnessWeight +
		dqs.config.ReliabilityWeight + dqs.config.ValidityWeight

	if totalWeight == 0 {
		totalWeight = 6.0 // Equal weights
	}

	weightedScore := (result.CompletenessScore * dqs.config.CompletenessWeight) +
		(result.AccuracyScore * dqs.config.AccuracyWeight) +
		(result.ConsistencyScore * dqs.config.ConsistencyWeight) +
		(result.FreshnessScore * dqs.config.FreshnessWeight) +
		(result.ReliabilityScore * dqs.config.ReliabilityWeight) +
		(result.ValidityScore * dqs.config.ValidityWeight)

	return weightedScore / totalWeight
}

// Helper methods

func (dqs *DataQualityScorer) isRequiredField(fieldName string, fieldType reflect.StructField) bool {
	// Check if field has required tag or is in required fields list
	if _, exists := dqs.config.RequiredFieldWeights[fieldName]; exists {
		return true
	}

	// Check struct tags
	tag := fieldType.Tag.Get("json")
	return !strings.Contains(tag, "omitempty")
}

func (dqs *DataQualityScorer) getFieldWeight(fieldName string) float64 {
	if weight, exists := dqs.config.RequiredFieldWeights[fieldName]; exists {
		return weight
	}
	if weight, exists := dqs.config.OptionalFieldWeights[fieldName]; exists {
		return weight
	}
	return 1.0 // Default weight
}

func (dqs *DataQualityScorer) extractConfidenceScore(v reflect.Value) float64 {
	// Look for common confidence score field names
	confidenceFields := []string{"ConfidenceScore", "Confidence", "Score"}

	for _, fieldName := range confidenceFields {
		field := v.FieldByName(fieldName)
		if field.IsValid() && field.Kind() == reflect.Float64 {
			return field.Float()
		}
	}

	return 0.0
}

func (dqs *DataQualityScorer) validateFieldAccuracy(field reflect.Value, fieldName string, fieldType reflect.StructField) float64 {
	// Basic field validation based on type and content
	switch field.Kind() {
	case reflect.String:
		str := field.String()
		if str == "" {
			return 0.0
		}
		// Check for reasonable string content
		if len(str) < 2 {
			return 0.5
		}
		return 1.0
	case reflect.Float64:
		f := field.Float()
		if f < 0 || f > 1 {
			return 0.8 // Might be valid but unusual
		}
		return 1.0
	case reflect.Int, reflect.Int64:
		i := field.Int()
		if i < 0 {
			return 0.8 // Might be valid but unusual
		}
		return 1.0
	default:
		return 0.9 // Default score for other types
	}
}

func (dqs *DataQualityScorer) detectConsistencyIssues(data interface{}) []ConsistencyIssue {
	// This would be expanded with domain-specific consistency checks
	issues := []ConsistencyIssue{}

	// Example: Check for logical inconsistencies
	// This would be implemented based on specific data types and business rules

	return issues
}

func (dqs *DataQualityScorer) extractTimestamp(data interface{}) time.Time {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return time.Time{}
	}

	// Look for common timestamp field names
	timestampFields := []string{"CreatedAt", "UpdatedAt", "ProcessedAt", "ExtractedAt", "Timestamp"}

	for _, fieldName := range timestampFields {
		field := v.FieldByName(fieldName)
		if field.IsValid() && field.Type() == reflect.TypeOf(time.Time{}) {
			return field.Interface().(time.Time)
		}
	}

	return time.Time{}
}

func (dqs *DataQualityScorer) extractSourceInfo(data interface{}) string {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return "unknown"
	}

	// Look for source information fields
	sourceFields := []string{"Source", "DataSource", "Provider", "Origin"}

	for _, fieldName := range sourceFields {
		field := v.FieldByName(fieldName)
		if field.IsValid() && field.Kind() == reflect.String {
			return field.String()
		}
	}

	return "unknown"
}

func (dqs *DataQualityScorer) validateField(field reflect.Value, fieldName string, fieldType reflect.StructField) ValidationResult {
	result := ValidationResult{
		FieldName: fieldName,
		IsValid:   true,
		Score:     1.0,
	}

	// Perform basic validation based on field type and tags
	switch field.Kind() {
	case reflect.String:
		str := field.String()
		if strings.TrimSpace(str) == "" {
			result.IsValid = false
			result.Score = 0.0
			result.ErrorType = "empty_string"
			result.ErrorMessage = "Field contains empty or whitespace-only string"
		}
	case reflect.Float64:
		f := field.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			result.IsValid = false
			result.Score = 0.0
			result.ErrorType = "invalid_float"
			result.ErrorMessage = "Field contains NaN or Inf value"
		}
	}

	return result
}

func (dqs *DataQualityScorer) determineQualityLevel(score float64) string {
	if score >= 0.9 {
		return "excellent"
	} else if score >= dqs.config.HighQualityThreshold {
		return "high"
	} else if score >= dqs.config.MinQualityThreshold {
		return "medium"
	} else {
		return "low"
	}
}

func (dqs *DataQualityScorer) generateRecommendations(result *DataQualityResult) []string {
	recommendations := []string{}

	if result.CompletenessScore < dqs.config.CompletenessThreshold {
		recommendations = append(recommendations, "Improve data completeness by populating more required fields")
	}

	if result.AccuracyScore < dqs.config.AccuracyThreshold {
		recommendations = append(recommendations, "Enhance data accuracy through better validation and verification")
	}

	if result.ValidationErrors > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Fix %d validation errors to improve data quality", result.ValidationErrors))
	}

	if len(result.ConsistencyIssues) > 0 {
		recommendations = append(recommendations, "Resolve data consistency issues")
	}

	if result.FreshnessScore < 0.8 {
		recommendations = append(recommendations, "Update data more frequently to maintain freshness")
	}

	return recommendations
}

func (dqs *DataQualityScorer) detectOutliers(data interface{}, result *DataQualityResult) {
	// This would implement outlier detection algorithms
	// For now, just add a placeholder
	if result.OverallScore > 0.98 {
		result.QualityIssues = append(result.QualityIssues, QualityIssue{
			Category:    "outlier",
			Description: "Suspiciously high quality score may indicate data anomaly",
			Severity:    "low",
			Impact:      0.0,
			Fixable:     false,
		})
	}
}

// getDefaultDataQualityConfig returns default configuration for data quality scoring
func getDefaultDataQualityConfig() *DataQualityConfig {
	return &DataQualityConfig{
		// Equal weights for all components
		CompletenessWeight: 1.0,
		AccuracyWeight:     1.0,
		ConsistencyWeight:  1.0,
		FreshnessWeight:    1.0,
		ReliabilityWeight:  1.0,
		ValidityWeight:     1.0,

		// Quality thresholds
		MinQualityThreshold:   0.6,
		HighQualityThreshold:  0.8,
		CompletenessThreshold: 0.7,
		AccuracyThreshold:     0.8,
		FreshnessThreshold:    24, // 24 hours

		// Enable all features
		EnableValidation:       true,
		EnableConsistencyCheck: true,
		EnableFreshnessCheck:   true,
		EnableOutlierDetection: true,

		// Default field weights
		RequiredFieldWeights: map[string]float64{
			"ID":              2.0,
			"Name":            2.0,
			"BusinessName":    2.0,
			"CompanyName":     2.0,
			"ConfidenceScore": 1.5,
		},
		OptionalFieldWeights: map[string]float64{
			"Description": 0.5,
			"Notes":       0.3,
			"Metadata":    0.3,
		},
	}
}
