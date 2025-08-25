package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"

	"go.opentelemetry.io/otel/trace"
)

// UnifiedResponse represents the standardized response format across all modules
type UnifiedResponse struct {
	// Core response data
	Data interface{} `json:"data"`

	// Response metadata
	Metadata *ResponseMetadata `json:"metadata"`

	// Confidence and quality information
	Confidence *ConfidenceInfo `json:"confidence"`

	// Processing information
	Processing *ProcessingInfo `json:"processing"`

	// Error information (if applicable)
	Error *ErrorInfo `json:"error,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ResponseMetadata contains metadata about the response
type ResponseMetadata struct {
	// Request information
	RequestID     string `json:"request_id"`
	CorrelationID string `json:"correlation_id"`
	UserID        string `json:"user_id,omitempty"`

	// Module information
	ModuleID      string `json:"module_id"`
	ModuleVersion string `json:"module_version"`
	ModuleName    string `json:"module_name"`

	// Data source information
	DataSources []string  `json:"data_sources"`
	LastUpdated time.Time `json:"last_updated"`

	// Response characteristics
	ResponseType string        `json:"response_type"`
	ResponseSize int           `json:"response_size"`
	IsCached     bool          `json:"is_cached"`
	CacheTTL     time.Duration `json:"cache_ttl,omitempty"`

	// Custom metadata
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// ConfidenceInfo contains confidence scoring and quality metrics
type ConfidenceInfo struct {
	// Overall confidence score (0.0 to 1.0)
	OverallScore float64 `json:"overall_score"`

	// Individual confidence scores
	AccuracyScore     float64 `json:"accuracy_score"`
	CompletenessScore float64 `json:"completeness_score"`
	FreshnessScore    float64 `json:"freshness_score"`
	ConsistencyScore  float64 `json:"consistency_score"`

	// Quality indicators
	QualityLevel     string  `json:"quality_level"` // high, medium, low
	ReliabilityScore float64 `json:"reliability_score"`

	// Confidence breakdown by data point
	DataPointConfidence map[string]float64 `json:"data_point_confidence,omitempty"`

	// Confidence factors
	ConfidenceFactors []ConfidenceFactor `json:"confidence_factors,omitempty"`
}

// ConfidenceFactor represents a factor that influences confidence scoring
type ConfidenceFactor struct {
	Factor string  `json:"factor"`
	Score  float64 `json:"score"`
	Weight float64 `json:"weight"`
	Reason string  `json:"reason"`
	Impact string  `json:"impact"` // positive, negative, neutral
}

// ProcessingInfo contains information about response processing
type ProcessingInfo struct {
	// Processing metrics
	ProcessingTime time.Duration `json:"processing_time"`
	QueueTime      time.Duration `json:"queue_time"`
	TotalTime      time.Duration `json:"total_time"`

	// Processing details
	ProcessingSteps    []ProcessingStep `json:"processing_steps"`
	ParallelProcessing bool             `json:"parallel_processing"`
	RetryCount         int              `json:"retry_count"`

	// Resource usage
	MemoryUsage int64   `json:"memory_usage_bytes"`
	CPUUsage    float64 `json:"cpu_usage_percent"`

	// Performance indicators
	PerformanceClass string  `json:"performance_class"` // fast, medium, slow
	EfficiencyScore  float64 `json:"efficiency_score"`
}

// ProcessingStep represents a step in the processing pipeline
type ProcessingStep struct {
	StepName   string        `json:"step_name"`
	StepOrder  int           `json:"step_order"`
	Duration   time.Duration `json:"duration"`
	Status     string        `json:"status"` // success, failed, skipped
	Error      string        `json:"error,omitempty"`
	InputSize  int           `json:"input_size"`
	OutputSize int           `json:"output_size"`
}

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	// Error details
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	ErrorType    string `json:"error_type"` // validation, processing, external, system

	// Error context
	ErrorContext map[string]interface{} `json:"error_context,omitempty"`
	ErrorStack   []string               `json:"error_stack,omitempty"`

	// Recovery information
	IsRecoverable bool          `json:"is_recoverable"`
	RetryAfter    time.Duration `json:"retry_after,omitempty"`
	Suggestions   []string      `json:"suggestions,omitempty"`
}

// ResponseBuilder provides a fluent interface for building unified responses
type ResponseBuilder struct {
	response *UnifiedResponse
	logger   *observability.Logger
	tracer   trace.Tracer
}

// NewResponseBuilder creates a new response builder
func NewResponseBuilder(
	logger *observability.Logger,
	tracer trace.Tracer,
) *ResponseBuilder {
	return &ResponseBuilder{
		response: &UnifiedResponse{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		logger: logger,
		tracer: tracer,
	}
}

// WithData sets the response data
func (rb *ResponseBuilder) WithData(data interface{}) *ResponseBuilder {
	rb.response.Data = data
	rb.response.UpdatedAt = time.Now()
	return rb
}

// WithMetadata sets the response metadata
func (rb *ResponseBuilder) WithMetadata(metadata *ResponseMetadata) *ResponseBuilder {
	rb.response.Metadata = metadata
	rb.response.UpdatedAt = time.Now()
	return rb
}

// WithConfidence sets the confidence information
func (rb *ResponseBuilder) WithConfidence(confidence *ConfidenceInfo) *ResponseBuilder {
	rb.response.Confidence = confidence
	rb.response.UpdatedAt = time.Now()
	return rb
}

// WithProcessing sets the processing information
func (rb *ResponseBuilder) WithProcessing(processing *ProcessingInfo) *ResponseBuilder {
	rb.response.Processing = processing
	rb.response.UpdatedAt = time.Now()
	return rb
}

// WithError sets the error information
func (rb *ResponseBuilder) WithError(error *ErrorInfo) *ResponseBuilder {
	rb.response.Error = error
	rb.response.UpdatedAt = time.Now()
	return rb
}

// WithRequestID sets the request ID
func (rb *ResponseBuilder) WithRequestID(requestID string) *ResponseBuilder {
	if rb.response.Metadata == nil {
		rb.response.Metadata = &ResponseMetadata{}
	}
	rb.response.Metadata.RequestID = requestID
	rb.response.UpdatedAt = time.Now()
	return rb
}

// WithCorrelationID sets the correlation ID
func (rb *ResponseBuilder) WithCorrelationID(correlationID string) *ResponseBuilder {
	if rb.response.Metadata == nil {
		rb.response.Metadata = &ResponseMetadata{}
	}
	rb.response.Metadata.CorrelationID = correlationID
	rb.response.UpdatedAt = time.Now()
	return rb
}

// WithModuleInfo sets the module information
func (rb *ResponseBuilder) WithModuleInfo(moduleID, moduleVersion, moduleName string) *ResponseBuilder {
	if rb.response.Metadata == nil {
		rb.response.Metadata = &ResponseMetadata{}
	}
	rb.response.Metadata.ModuleID = moduleID
	rb.response.Metadata.ModuleVersion = moduleVersion
	rb.response.Metadata.ModuleName = moduleName
	rb.response.UpdatedAt = time.Now()
	return rb
}

// WithDataSources sets the data sources
func (rb *ResponseBuilder) WithDataSources(dataSources []string) *ResponseBuilder {
	if rb.response.Metadata == nil {
		rb.response.Metadata = &ResponseMetadata{}
	}
	rb.response.Metadata.DataSources = dataSources
	rb.response.UpdatedAt = time.Now()
	return rb
}

// WithProcessingTime sets the processing time
func (rb *ResponseBuilder) WithProcessingTime(processingTime time.Duration) *ResponseBuilder {
	if rb.response.Processing == nil {
		rb.response.Processing = &ProcessingInfo{}
	}
	rb.response.Processing.ProcessingTime = processingTime
	rb.response.Processing.TotalTime = processingTime
	rb.response.UpdatedAt = time.Now()
	return rb
}

// WithConfidenceScore sets the overall confidence score
func (rb *ResponseBuilder) WithConfidenceScore(score float64) *ResponseBuilder {
	if rb.response.Confidence == nil {
		rb.response.Confidence = &ConfidenceInfo{}
	}
	rb.response.Confidence.OverallScore = score
	rb.response.UpdatedAt = time.Now()
	return rb
}

// WithQualityScores sets the quality scores
func (rb *ResponseBuilder) WithQualityScores(accuracy, completeness, freshness, consistency float64) *ResponseBuilder {
	if rb.response.Confidence == nil {
		rb.response.Confidence = &ConfidenceInfo{}
	}
	rb.response.Confidence.AccuracyScore = accuracy
	rb.response.Confidence.CompletenessScore = completeness
	rb.response.Confidence.FreshnessScore = freshness
	rb.response.Confidence.ConsistencyScore = consistency
	rb.response.UpdatedAt = time.Now()
	return rb
}

// WithCacheInfo sets cache information
func (rb *ResponseBuilder) WithCacheInfo(isCached bool, ttl time.Duration) *ResponseBuilder {
	if rb.response.Metadata == nil {
		rb.response.Metadata = &ResponseMetadata{}
	}
	rb.response.Metadata.IsCached = isCached
	rb.response.Metadata.CacheTTL = ttl
	rb.response.UpdatedAt = time.Now()
	return rb
}

// WithCustomField adds a custom field to metadata
func (rb *ResponseBuilder) WithCustomField(key string, value interface{}) *ResponseBuilder {
	if rb.response.Metadata == nil {
		rb.response.Metadata = &ResponseMetadata{}
	}
	if rb.response.Metadata.CustomFields == nil {
		rb.response.Metadata.CustomFields = make(map[string]interface{})
	}
	rb.response.Metadata.CustomFields[key] = value
	rb.response.UpdatedAt = time.Now()
	return rb
}

// Build creates the final unified response
func (rb *ResponseBuilder) Build() *UnifiedResponse {
	// Set response size
	if rb.response.Metadata != nil {
		if data, err := json.Marshal(rb.response.Data); err == nil {
			rb.response.Metadata.ResponseSize = len(data)
		}
	}

	// Set response type
	if rb.response.Metadata != nil {
		rb.response.Metadata.ResponseType = rb.determineResponseType()
	}

	// Validate response
	if err := rb.validateResponse(); err != nil {
		rb.logger.Error("response validation failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return rb.response
}

// determineResponseType determines the response type based on the data
func (rb *ResponseBuilder) determineResponseType() string {
	if rb.response.Data == nil {
		return "empty"
	}

	switch rb.response.Data.(type) {
	case *BusinessClassificationResponse:
		return "business_classification"
	case *BatchClassificationResponse:
		return "batch_classification"
	case []*BusinessClassificationResponse:
		return "business_classification_list"
	case map[string]interface{}:
		return "generic_object"
	case []interface{}:
		return "generic_array"
	default:
		return "unknown"
	}
}

// validateResponse validates the unified response
func (rb *ResponseBuilder) validateResponse() error {
	// Validate required fields
	if rb.response.Data == nil {
		return fmt.Errorf("response data cannot be nil")
	}

	if rb.response.Metadata == nil {
		return fmt.Errorf("response metadata cannot be nil")
	}

	if rb.response.Metadata.RequestID == "" {
		return fmt.Errorf("request ID is required")
	}

	// Validate confidence scores
	if rb.response.Confidence != nil {
		if rb.response.Confidence.OverallScore < 0 || rb.response.Confidence.OverallScore > 1 {
			return fmt.Errorf("confidence score must be between 0 and 1")
		}
	}

	// Validate processing time
	if rb.response.Processing != nil {
		if rb.response.Processing.ProcessingTime < 0 {
			return fmt.Errorf("processing time cannot be negative")
		}
	}

	return nil
}

// ResponseValidator provides validation for unified responses
type ResponseValidator struct {
	logger *observability.Logger
}

// NewResponseValidator creates a new response validator
func NewResponseValidator(logger *observability.Logger) *ResponseValidator {
	return &ResponseValidator{
		logger: logger,
	}
}

// ValidateResponse validates a unified response
func (rv *ResponseValidator) ValidateResponse(ctx context.Context, response *UnifiedResponse) error {
	ctx, span := rv.logger.Tracer().Start(ctx, "ResponseValidator.ValidateResponse")
	defer span.End()

	if response == nil {
		return fmt.Errorf("response cannot be nil")
	}

	// Validate basic structure
	if err := rv.validateBasicStructure(response); err != nil {
		return fmt.Errorf("basic structure validation failed: %w", err)
	}

	// Validate metadata
	if err := rv.validateMetadata(response.Metadata); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	// Validate confidence information
	if err := rv.validateConfidence(response.Confidence); err != nil {
		return fmt.Errorf("confidence validation failed: %w", err)
	}

	// Validate processing information
	if err := rv.validateProcessing(response.Processing); err != nil {
		return fmt.Errorf("processing validation failed: %w", err)
	}

	// Validate error information
	if err := rv.validateError(response.Error); err != nil {
		return fmt.Errorf("error validation failed: %w", err)
	}

	return nil
}

// validateBasicStructure validates the basic response structure
func (rv *ResponseValidator) validateBasicStructure(response *UnifiedResponse) error {
	if response.Data == nil {
		return fmt.Errorf("data cannot be nil")
	}

	if response.Metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}

	if response.CreatedAt.IsZero() {
		return fmt.Errorf("created_at cannot be zero")
	}

	if response.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at cannot be zero")
	}

	if response.UpdatedAt.Before(response.CreatedAt) {
		return fmt.Errorf("updated_at cannot be before created_at")
	}

	return nil
}

// validateMetadata validates response metadata
func (rv *ResponseValidator) validateMetadata(metadata *ResponseMetadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}

	if metadata.RequestID == "" {
		return fmt.Errorf("request_id is required")
	}

	if metadata.ModuleID == "" {
		return fmt.Errorf("module_id is required")
	}

	if metadata.ResponseSize < 0 {
		return fmt.Errorf("response_size cannot be negative")
	}

	return nil
}

// validateConfidence validates confidence information
func (rv *ResponseValidator) validateConfidence(confidence *ConfidenceInfo) error {
	if confidence == nil {
		return nil // Confidence is optional
	}

	if confidence.OverallScore < 0 || confidence.OverallScore > 1 {
		return fmt.Errorf("overall_score must be between 0 and 1")
	}

	if confidence.AccuracyScore < 0 || confidence.AccuracyScore > 1 {
		return fmt.Errorf("accuracy_score must be between 0 and 1")
	}

	if confidence.CompletenessScore < 0 || confidence.CompletenessScore > 1 {
		return fmt.Errorf("completeness_score must be between 0 and 1")
	}

	if confidence.FreshnessScore < 0 || confidence.FreshnessScore > 1 {
		return fmt.Errorf("freshness_score must be between 0 and 1")
	}

	if confidence.ConsistencyScore < 0 || confidence.ConsistencyScore > 1 {
		return fmt.Errorf("consistency_score must be between 0 and 1")
	}

	if confidence.ReliabilityScore < 0 || confidence.ReliabilityScore > 1 {
		return fmt.Errorf("reliability_score must be between 0 and 1")
	}

	return nil
}

// validateProcessing validates processing information
func (rv *ResponseValidator) validateProcessing(processing *ProcessingInfo) error {
	if processing == nil {
		return nil // Processing is optional
	}

	if processing.ProcessingTime < 0 {
		return fmt.Errorf("processing_time cannot be negative")
	}

	if processing.QueueTime < 0 {
		return fmt.Errorf("queue_time cannot be negative")
	}

	if processing.TotalTime < 0 {
		return fmt.Errorf("total_time cannot be negative")
	}

	if processing.RetryCount < 0 {
		return fmt.Errorf("retry_count cannot be negative")
	}

	if processing.MemoryUsage < 0 {
		return fmt.Errorf("memory_usage cannot be negative")
	}

	if processing.CPUUsage < 0 || processing.CPUUsage > 100 {
		return fmt.Errorf("cpu_usage must be between 0 and 100")
	}

	if processing.EfficiencyScore < 0 || processing.EfficiencyScore > 1 {
		return fmt.Errorf("efficiency_score must be between 0 and 1")
	}

	return nil
}

// validateError validates error information
func (rv *ResponseValidator) validateError(error *ErrorInfo) error {
	if error == nil {
		return nil // Error is optional
	}

	if error.ErrorCode == "" {
		return fmt.Errorf("error_code is required when error is present")
	}

	if error.ErrorMessage == "" {
		return fmt.Errorf("error_message is required when error is present")
	}

	if error.ErrorType == "" {
		return fmt.Errorf("error_type is required when error is present")
	}

	return nil
}

// ConfidenceAggregator provides confidence scoring aggregation
type ConfidenceAggregator struct {
	logger *observability.Logger
}

// NewConfidenceAggregator creates a new confidence aggregator
func NewConfidenceAggregator(logger *observability.Logger) *ConfidenceAggregator {
	return &ConfidenceAggregator{
		logger: logger,
	}
}

// AggregateConfidence aggregates confidence scores from multiple sources
func (ca *ConfidenceAggregator) AggregateConfidence(
	ctx context.Context,
	scores []float64,
	weights []float64,
) *ConfidenceInfo {
	ctx, span := ca.logger.Tracer().Start(ctx, "ConfidenceAggregator.AggregateConfidence")
	defer span.End()

	if len(scores) == 0 {
		return &ConfidenceInfo{
			OverallScore: 0.0,
			QualityLevel: "unknown",
		}
	}

	// Use equal weights if not provided
	if len(weights) == 0 {
		weights = make([]float64, len(scores))
		for i := range weights {
			weights[i] = 1.0 / float64(len(scores))
		}
	}

	// Normalize weights
	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}
	for i := range weights {
		weights[i] /= totalWeight
	}

	// Calculate weighted average
	overallScore := 0.0
	for i, score := range scores {
		if i < len(weights) {
			overallScore += score * weights[i]
		}
	}

	// Calculate individual scores (assuming equal distribution for now)
	accuracyScore := overallScore * 0.25
	completenessScore := overallScore * 0.25
	freshnessScore := overallScore * 0.25
	consistencyScore := overallScore * 0.25

	// Determine quality level
	qualityLevel := ca.determineQualityLevel(overallScore)

	// Calculate reliability score
	reliabilityScore := ca.calculateReliabilityScore(scores, weights)

	return &ConfidenceInfo{
		OverallScore:      overallScore,
		AccuracyScore:     accuracyScore,
		CompletenessScore: completenessScore,
		FreshnessScore:    freshnessScore,
		ConsistencyScore:  consistencyScore,
		QualityLevel:      qualityLevel,
		ReliabilityScore:  reliabilityScore,
	}
}

// determineQualityLevel determines the quality level based on confidence score
func (ca *ConfidenceAggregator) determineQualityLevel(score float64) string {
	switch {
	case score >= 0.8:
		return "high"
	case score >= 0.6:
		return "medium"
	case score >= 0.4:
		return "low"
	default:
		return "very_low"
	}
}

// calculateReliabilityScore calculates reliability based on score consistency
func (ca *ConfidenceAggregator) calculateReliabilityScore(scores []float64, weights []float64) float64 {
	if len(scores) == 0 {
		return 0.0
	}

	// Calculate weighted mean
	weightedMean := 0.0
	totalWeight := 0.0
	for i, score := range scores {
		weight := 1.0
		if i < len(weights) {
			weight = weights[i]
		}
		weightedMean += score * weight
		totalWeight += weight
	}
	weightedMean /= totalWeight

	// Calculate weighted variance
	variance := 0.0
	for i, score := range scores {
		weight := 1.0
		if i < len(weights) {
			weight = weights[i]
		}
		variance += weight * (score - weightedMean) * (score - weightedMean)
	}
	variance /= totalWeight

	// Reliability is inversely proportional to variance
	reliability := 1.0 / (1.0 + variance)
	return reliability
}

// ResponseFormatter provides formatting utilities for unified responses
type ResponseFormatter struct {
	logger *observability.Logger
}

// NewResponseFormatter creates a new response formatter
func NewResponseFormatter(logger *observability.Logger) *ResponseFormatter {
	return &ResponseFormatter{
		logger: logger,
	}
}

// FormatForAPI formats a unified response for API consumption
func (rf *ResponseFormatter) FormatForAPI(ctx context.Context, response *UnifiedResponse) map[string]interface{} {
	ctx, span := rf.logger.Tracer().Start(ctx, "ResponseFormatter.FormatForAPI")
	defer span.End()

	formatted := map[string]interface{}{
		"data":       response.Data,
		"created_at": response.CreatedAt.Format(time.RFC3339),
		"updated_at": response.UpdatedAt.Format(time.RFC3339),
	}

	if response.Metadata != nil {
		formatted["metadata"] = map[string]interface{}{
			"request_id":     response.Metadata.RequestID,
			"correlation_id": response.Metadata.CorrelationID,
			"module_id":      response.Metadata.ModuleID,
			"module_name":    response.Metadata.ModuleName,
			"response_type":  response.Metadata.ResponseType,
			"is_cached":      response.Metadata.IsCached,
		}
	}

	if response.Confidence != nil {
		formatted["confidence"] = map[string]interface{}{
			"overall_score":     response.Confidence.OverallScore,
			"quality_level":     response.Confidence.QualityLevel,
			"reliability_score": response.Confidence.ReliabilityScore,
		}
	}

	if response.Processing != nil {
		formatted["processing"] = map[string]interface{}{
			"processing_time_ms": response.Processing.ProcessingTime.Milliseconds(),
			"performance_class":  response.Processing.PerformanceClass,
		}
	}

	if response.Error != nil {
		formatted["error"] = map[string]interface{}{
			"code":    response.Error.ErrorCode,
			"message": response.Error.ErrorMessage,
			"type":    response.Error.ErrorType,
		}
	}

	return formatted
}

// FormatForLogging formats a unified response for logging
func (rf *ResponseFormatter) FormatForLogging(ctx context.Context, response *UnifiedResponse) map[string]interface{} {
	ctx, span := rf.logger.Tracer().Start(ctx, "ResponseFormatter.FormatForLogging")
	defer span.End()

	logData := map[string]interface{}{
		"request_id":    response.Metadata.RequestID,
		"module_id":     response.Metadata.ModuleID,
		"response_type": response.Metadata.ResponseType,
		"created_at":    response.CreatedAt.Format(time.RFC3339),
	}

	if response.Confidence != nil {
		logData["confidence_score"] = response.Confidence.OverallScore
		logData["quality_level"] = response.Confidence.QualityLevel
	}

	if response.Processing != nil {
		logData["processing_time_ms"] = response.Processing.ProcessingTime.Milliseconds()
		logData["performance_class"] = response.Processing.PerformanceClass
	}

	if response.Error != nil {
		logData["error_code"] = response.Error.ErrorCode
		logData["error_type"] = response.Error.ErrorType
	}

	return logData
}
