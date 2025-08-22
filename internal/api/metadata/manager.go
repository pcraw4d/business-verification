package metadata

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// metadataManager implements the MetadataManager interface
type metadataManager struct {
	logger *zap.Logger

	// Data storage
	dataSources       map[string]*DataSourceMetadata
	confidenceHistory map[string][]*ConfidenceMetadata
	responseMetadata  map[string]*ResponseMetadata
	traceability      map[string]*TraceabilityMetadata

	// Configuration
	config *MetadataConfig

	// Mutex for thread safety
	mu sync.RWMutex
}

// MetadataConfig contains configuration for the metadata manager
type MetadataConfig struct {
	// Confidence calculation settings
	ConfidenceWeights    map[string]float64 `json:"confidence_weights"`
	ConfidenceThresholds map[string]float64 `json:"confidence_thresholds"`

	// Quality assessment settings
	QualityWeights    map[string]float64 `json:"quality_weights"`
	QualityThresholds map[string]float64 `json:"quality_thresholds"`

	// Validation settings
	ValidationRules []ValidationRule `json:"validation_rules"`

	// Compliance settings
	ComplianceFrameworks []ComplianceFramework `json:"compliance_frameworks"`

	// Performance settings
	MaxHistorySize  int           `json:"max_history_size"`
	CleanupInterval time.Duration `json:"cleanup_interval"`

	// Calibration settings
	EnableCalibration bool    `json:"enable_calibration"`
	CalibrationFactor float64 `json:"calibration_factor"`
}

// NewMetadataManager creates a new metadata manager instance
func NewMetadataManager(logger *zap.Logger, config *MetadataConfig) MetadataManager {
	if config == nil {
		config = getDefaultConfig()
	}

	mm := &metadataManager{
		logger:            logger,
		config:            config,
		dataSources:       make(map[string]*DataSourceMetadata),
		confidenceHistory: make(map[string][]*ConfidenceMetadata),
		responseMetadata:  make(map[string]*ResponseMetadata),
		traceability:      make(map[string]*TraceabilityMetadata),
	}

	// Start cleanup routine
	go mm.cleanupRoutine()

	return mm
}

// AddDataSource adds a new data source to the metadata manager
func (mm *metadataManager) AddDataSource(ctx context.Context, source *DataSourceMetadata) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if source == nil {
		return fmt.Errorf("data source cannot be nil")
	}

	if source.SourceID == "" {
		return fmt.Errorf("data source ID cannot be empty")
	}

	// Check if data source already exists
	if _, exists := mm.dataSources[source.SourceID]; exists {
		return fmt.Errorf("data source already exists: %s", source.SourceID)
	}

	// Set default values if not provided
	if source.LastUpdated.IsZero() {
		source.LastUpdated = time.Now()
	}

	if source.Metadata == nil {
		source.Metadata = make(map[string]interface{})
	}

	mm.dataSources[source.SourceID] = source

	mm.logger.Info("Data source added",
		zap.String("source_id", source.SourceID),
		zap.String("source_name", source.SourceName),
		zap.String("source_type", source.SourceType))

	return nil
}

// GetDataSource retrieves a data source by ID
func (mm *metadataManager) GetDataSource(ctx context.Context, sourceID string) (*DataSourceMetadata, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	source, exists := mm.dataSources[sourceID]
	if !exists {
		return nil, fmt.Errorf("data source not found: %s", sourceID)
	}

	return source, nil
}

// UpdateDataSource updates an existing data source
func (mm *metadataManager) UpdateDataSource(ctx context.Context, source *DataSourceMetadata) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if source == nil {
		return fmt.Errorf("data source cannot be nil")
	}

	if source.SourceID == "" {
		return fmt.Errorf("data source ID cannot be empty")
	}

	if _, exists := mm.dataSources[source.SourceID]; !exists {
		return fmt.Errorf("data source not found: %s", source.SourceID)
	}

	// Update timestamp
	source.LastUpdated = time.Now()

	mm.dataSources[source.SourceID] = source

	mm.logger.Info("Data source updated",
		zap.String("source_id", source.SourceID),
		zap.String("source_name", source.SourceName))

	return nil
}

// ListDataSources returns all data sources
func (mm *metadataManager) ListDataSources(ctx context.Context) ([]*DataSourceMetadata, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	sources := make([]*DataSourceMetadata, 0, len(mm.dataSources))
	for _, source := range mm.dataSources {
		sources = append(sources, source)
	}

	return sources, nil
}

// CalculateConfidence calculates confidence metadata from factors
func (mm *metadataManager) CalculateConfidence(ctx context.Context, factors []ConfidenceFactor) (*ConfidenceMetadata, error) {
	if len(factors) == 0 {
		return nil, fmt.Errorf("at least one confidence factor is required")
	}

	confidence := &ConfidenceMetadata{
		ComponentScores: make(map[string]float64),
		Factors:         factors,
		CalculatedAt:    time.Now(),
		Metadata:        make(map[string]interface{}),
	}

	// Calculate component scores
	var totalWeightedScore float64
	var totalWeight float64

	for _, factor := range factors {
		weightedScore := factor.FactorValue * factor.FactorWeight
		totalWeightedScore += weightedScore
		totalWeight += factor.FactorWeight

		confidence.ComponentScores[factor.FactorName] = factor.FactorValue
	}

	// Calculate overall confidence
	if totalWeight > 0 {
		confidence.OverallConfidence = totalWeightedScore / totalWeight
	} else {
		confidence.OverallConfidence = 0.0
	}

	// Apply calibration if enabled
	if mm.config.EnableCalibration {
		confidence.OverallConfidence *= mm.config.CalibrationFactor
		confidence.OverallConfidence = math.Min(1.0, confidence.OverallConfidence)
		confidence.Calibration = &CalibrationData{
			CalibrationFactor: mm.config.CalibrationFactor,
			LastCalibrated:    time.Now(),
			CalibrationMethod: "global_factor",
		}
	}

	// Determine confidence level
	confidence.ConfidenceLevel = mm.determineConfidenceLevel(confidence.OverallConfidence)

	// Calculate uncertainty metrics
	confidence.Uncertainty = mm.calculateUncertainty(factors, confidence.OverallConfidence)

	mm.logger.Debug("Confidence calculated",
		zap.Float64("overall_confidence", confidence.OverallConfidence),
		zap.String("confidence_level", string(confidence.ConfidenceLevel)),
		zap.Int("factor_count", len(factors)))

	return confidence, nil
}

// UpdateConfidence updates confidence metadata
func (mm *metadataManager) UpdateConfidence(ctx context.Context, confidence *ConfidenceMetadata) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if confidence == nil {
		return fmt.Errorf("confidence metadata cannot be nil")
	}

	// Store in history (this would typically be stored in a database)
	// For now, we'll keep it in memory
	mm.logger.Debug("Confidence updated",
		zap.Float64("overall_confidence", confidence.OverallConfidence),
		zap.String("confidence_level", string(confidence.ConfidenceLevel)))

	return nil
}

// GetConfidenceHistory retrieves confidence history for a request
func (mm *metadataManager) GetConfidenceHistory(ctx context.Context, requestID string) ([]*ConfidenceMetadata, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	history, exists := mm.confidenceHistory[requestID]
	if !exists {
		return []*ConfidenceMetadata{}, nil
	}

	return history, nil
}

// CreateResponseMetadata creates new response metadata
func (mm *metadataManager) CreateResponseMetadata(ctx context.Context, requestID string) (*ResponseMetadata, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if requestID == "" {
		requestID = mm.GenerateRequestID()
	}

	metadata := &ResponseMetadata{
		RequestID:   requestID,
		Timestamp:   time.Now(),
		APIVersion:  "v3",
		DataSources: []DataSourceMetadata{},
		Metadata:    make(map[string]interface{}),
	}

	mm.responseMetadata[requestID] = metadata

	mm.logger.Debug("Response metadata created",
		zap.String("request_id", requestID))

	return metadata, nil
}

// UpdateResponseMetadata updates response metadata
func (mm *metadataManager) UpdateResponseMetadata(ctx context.Context, metadata *ResponseMetadata) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if metadata == nil {
		return fmt.Errorf("response metadata cannot be nil")
	}

	if metadata.RequestID == "" {
		return fmt.Errorf("request ID cannot be empty")
	}

	mm.responseMetadata[metadata.RequestID] = metadata

	mm.logger.Debug("Response metadata updated",
		zap.String("request_id", metadata.RequestID))

	return nil
}

// GetResponseMetadata retrieves response metadata
func (mm *metadataManager) GetResponseMetadata(ctx context.Context, requestID string) (*ResponseMetadata, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	metadata, exists := mm.responseMetadata[requestID]
	if !exists {
		return nil, fmt.Errorf("response metadata not found: %s", requestID)
	}

	return metadata, nil
}

// ValidateMetadata validates response metadata
func (mm *metadataManager) ValidateMetadata(ctx context.Context, metadata *ResponseMetadata) (*ValidationMetadata, error) {
	if metadata == nil {
		return nil, fmt.Errorf("metadata cannot be nil")
	}

	validation := &ValidationMetadata{
		ValidationStatus: "valid",
		ValidationScore:  1.0,
		ValidationRules:  []ValidationRule{},
		ValidatedAt:      time.Now(),
		ValidatorVersion: "1.0.0",
		Metadata:         make(map[string]interface{}),
	}

	// Apply validation rules
	for _, rule := range mm.config.ValidationRules {
		ruleResult := mm.applyValidationRule(metadata, rule)
		validation.ValidationRules = append(validation.ValidationRules, ruleResult)

		if ruleResult.Status == "failed" {
			validation.ValidationStatus = "invalid"
			validation.ValidationScore *= 0.8 // Reduce score for failed rules
		}
	}

	mm.logger.Debug("Metadata validation completed",
		zap.String("request_id", metadata.RequestID),
		zap.String("validation_status", validation.ValidationStatus),
		zap.Float64("validation_score", validation.ValidationScore))

	return validation, nil
}

// AssessQuality assesses the quality of response metadata
func (mm *metadataManager) AssessQuality(ctx context.Context, metadata *ResponseMetadata) (*QualityMetadata, error) {
	if metadata == nil {
		return nil, fmt.Errorf("metadata cannot be nil")
	}

	quality := &QualityMetadata{
		QualityFactors: []QualityFactor{},
		QualityLevel:   "high",
		Metadata:       make(map[string]interface{}),
	}

	// Assess data quality
	quality.DataQuality = mm.assessDataQuality(metadata)

	// Assess process quality
	quality.ProcessQuality = mm.assessProcessQuality(metadata)

	// Assess output quality
	quality.OutputQuality = mm.assessOutputQuality(metadata)

	// Calculate overall quality
	quality.OverallQuality = (quality.DataQuality + quality.ProcessQuality + quality.OutputQuality) / 3.0
	quality.QualityScore = quality.OverallQuality

	// Determine quality level
	quality.QualityLevel = mm.determineQualityLevel(quality.OverallQuality)

	mm.logger.Debug("Quality assessment completed",
		zap.String("request_id", metadata.RequestID),
		zap.Float64("overall_quality", quality.OverallQuality),
		zap.String("quality_level", quality.QualityLevel))

	return quality, nil
}

// CreateTraceability creates new traceability metadata
func (mm *metadataManager) CreateTraceability(ctx context.Context, requestID string) (*TraceabilityMetadata, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	traceability := &TraceabilityMetadata{
		TraceID:         mm.GenerateTraceID(),
		CorrelationID:   mm.GenerateCorrelationID(),
		RequestTrace:    []TraceEvent{},
		DataLineage:     []DataLineageItem{},
		ProcessingSteps: []ProcessingStep{},
		Metadata:        make(map[string]interface{}),
	}

	mm.traceability[requestID] = traceability

	mm.logger.Debug("Traceability created",
		zap.String("request_id", requestID),
		zap.String("trace_id", traceability.TraceID))

	return traceability, nil
}

// AddTraceEvent adds a trace event
func (mm *metadataManager) AddTraceEvent(ctx context.Context, traceID string, event *TraceEvent) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if event == nil {
		return fmt.Errorf("trace event cannot be nil")
	}

	// Find traceability by trace ID
	var traceability *TraceabilityMetadata
	for _, t := range mm.traceability {
		if t.TraceID == traceID {
			traceability = t
			break
		}
	}

	if traceability == nil {
		return fmt.Errorf("traceability not found for trace ID: %s", traceID)
	}

	// Set event ID if not provided
	if event.EventID == "" {
		event.EventID = mm.GenerateEventID()
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	traceability.RequestTrace = append(traceability.RequestTrace, *event)

	mm.logger.Debug("Trace event added",
		zap.String("trace_id", traceID),
		zap.String("event_id", event.EventID),
		zap.String("event_type", event.EventType))

	return nil
}

// AddDataLineage adds a data lineage item
func (mm *metadataManager) AddDataLineage(ctx context.Context, traceID string, lineage *DataLineageItem) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if lineage == nil {
		return fmt.Errorf("data lineage item cannot be nil")
	}

	// Find traceability by trace ID
	var traceability *TraceabilityMetadata
	for _, t := range mm.traceability {
		if t.TraceID == traceID {
			traceability = t
			break
		}
	}

	if traceability == nil {
		return fmt.Errorf("traceability not found for trace ID: %s", traceID)
	}

	// Set item ID if not provided
	if lineage.ItemID == "" {
		lineage.ItemID = mm.GenerateEventID()
	}

	// Set timestamp if not provided
	if lineage.Timestamp.IsZero() {
		lineage.Timestamp = time.Now()
	}

	traceability.DataLineage = append(traceability.DataLineage, *lineage)

	mm.logger.Debug("Data lineage added",
		zap.String("trace_id", traceID),
		zap.String("item_id", lineage.ItemID),
		zap.String("source_id", lineage.SourceID))

	return nil
}

// AddProcessingStep adds a processing step
func (mm *metadataManager) AddProcessingStep(ctx context.Context, traceID string, step *ProcessingStep) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if step == nil {
		return fmt.Errorf("processing step cannot be nil")
	}

	// Find traceability by trace ID
	var traceability *TraceabilityMetadata
	for _, t := range mm.traceability {
		if t.TraceID == traceID {
			traceability = t
			break
		}
	}

	if traceability == nil {
		return fmt.Errorf("traceability not found for trace ID: %s", traceID)
	}

	// Set step ID if not provided
	if step.StepID == "" {
		step.StepID = mm.GenerateEventID()
	}

	// Set timestamps if not provided
	if step.StartTime.IsZero() {
		step.StartTime = time.Now()
	}
	if step.EndTime.IsZero() {
		step.EndTime = time.Now()
	}

	// Calculate duration
	step.Duration = step.EndTime.Sub(step.StartTime)

	traceability.ProcessingSteps = append(traceability.ProcessingSteps, *step)

	mm.logger.Debug("Processing step added",
		zap.String("trace_id", traceID),
		zap.String("step_id", step.StepID),
		zap.String("step_name", step.StepName))

	return nil
}

// AssessCompliance assesses compliance of response metadata
func (mm *metadataManager) AssessCompliance(ctx context.Context, metadata *ResponseMetadata) (*ComplianceMetadata, error) {
	if metadata == nil {
		return nil, fmt.Errorf("metadata cannot be nil")
	}

	compliance := &ComplianceMetadata{
		ComplianceStatus: "compliant",
		ComplianceScore:  1.0,
		Frameworks:       []ComplianceFramework{},
		Requirements:     []ComplianceRequirement{},
		AuditTrail:       []AuditEvent{},
		Metadata:         make(map[string]interface{}),
	}

	// Assess compliance frameworks
	for _, framework := range mm.config.ComplianceFrameworks {
		frameworkResult := mm.assessComplianceFramework(metadata, framework)
		compliance.Frameworks = append(compliance.Frameworks, frameworkResult)

		if frameworkResult.Status != "compliant" {
			compliance.ComplianceStatus = "non_compliant"
			compliance.ComplianceScore *= 0.9 // Reduce score for non-compliant frameworks
		}
	}

	mm.logger.Debug("Compliance assessment completed",
		zap.String("request_id", metadata.RequestID),
		zap.String("compliance_status", compliance.ComplianceStatus),
		zap.Float64("compliance_score", compliance.ComplianceScore))

	return compliance, nil
}

// GenerateRequestID generates a unique request ID
func (mm *metadataManager) GenerateRequestID() string {
	return generateID("req")
}

// GenerateTraceID generates a unique trace ID
func (mm *metadataManager) GenerateTraceID() string {
	return generateID("trace")
}

// GenerateCorrelationID generates a unique correlation ID
func (mm *metadataManager) GenerateCorrelationID() string {
	return generateID("corr")
}

// Helper methods

func (mm *metadataManager) determineConfidenceLevel(confidence float64) ConfidenceLevel {
	thresholds := mm.config.ConfidenceThresholds

	if confidence >= thresholds["very_high"] {
		return ConfidenceLevelVeryHigh
	} else if confidence >= thresholds["high"] {
		return ConfidenceLevelHigh
	} else if confidence >= thresholds["medium"] {
		return ConfidenceLevelMedium
	} else if confidence >= thresholds["low"] {
		return ConfidenceLevelLow
	} else {
		return ConfidenceLevelVeryLow
	}
}

func (mm *metadataManager) calculateUncertainty(factors []ConfidenceFactor, overallConfidence float64) *UncertaintyMetrics {
	// Calculate variance from factors
	var variance float64
	var sum float64
	var count int

	for _, factor := range factors {
		sum += factor.FactorValue
		count++
	}

	if count > 0 {
		mean := sum / float64(count)
		for _, factor := range factors {
			variance += math.Pow(factor.FactorValue-mean, 2)
		}
		variance /= float64(count)
	}

	uncertaintyLevel := 1.0 - overallConfidence
	standardError := math.Sqrt(variance / float64(count))

	return &UncertaintyMetrics{
		UncertaintyLevel: uncertaintyLevel,
		ConfidenceInterval: ConfidenceInterval{
			LowerBound: math.Max(0.0, overallConfidence-1.96*standardError),
			UpperBound: math.Min(1.0, overallConfidence+1.96*standardError),
			Level:      0.95,
		},
		StandardError:    standardError,
		Variance:         variance,
		ReliabilityScore: overallConfidence,
	}
}

func (mm *metadataManager) applyValidationRule(metadata *ResponseMetadata, rule ValidationRule) ValidationRule {
	// Simple validation logic - in a real implementation, this would be more sophisticated
	rule.Status = "passed"
	rule.Confidence = 1.0

	// Example validation: check if required fields are present
	if rule.RuleID == "required_fields" {
		if metadata.RequestID == "" {
			rule.Status = "failed"
			rule.Confidence = 0.0
		}
	}

	return rule
}

func (mm *metadataManager) assessDataQuality(metadata *ResponseMetadata) float64 {
	// Simple data quality assessment
	quality := 1.0

	// Check data source quality
	for _, source := range metadata.DataSources {
		quality *= source.ReliabilityScore
	}

	// Check confidence
	if metadata.Confidence != nil {
		quality *= metadata.Confidence.OverallConfidence
	}

	return quality
}

func (mm *metadataManager) assessProcessQuality(metadata *ResponseMetadata) float64 {
	// Simple process quality assessment
	quality := 1.0

	// Check processing time (shorter is better, up to a point)
	if metadata.ProcessingTime > 0 {
		if metadata.ProcessingTime < 100*time.Millisecond {
			quality *= 1.0
		} else if metadata.ProcessingTime < 1*time.Second {
			quality *= 0.9
		} else {
			quality *= 0.8
		}
	}

	return quality
}

func (mm *metadataManager) assessOutputQuality(metadata *ResponseMetadata) float64 {
	// Simple output quality assessment
	quality := 1.0

	// Check if validation passed
	if metadata.Validation != nil {
		quality *= metadata.Validation.ValidationScore
	}

	return quality
}

func (mm *metadataManager) determineQualityLevel(quality float64) string {
	if quality >= 0.9 {
		return "excellent"
	} else if quality >= 0.8 {
		return "high"
	} else if quality >= 0.7 {
		return "medium"
	} else if quality >= 0.6 {
		return "low"
	} else {
		return "poor"
	}
}

func (mm *metadataManager) assessComplianceFramework(metadata *ResponseMetadata, framework ComplianceFramework) ComplianceFramework {
	// Simple compliance assessment
	framework.Status = "compliant"
	framework.Score = 1.0
	framework.Compliant = framework.Requirements
	framework.NonCompliant = 0

	return framework
}

func (mm *metadataManager) cleanupRoutine() {
	ticker := time.NewTicker(mm.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		mm.cleanup()
	}
}

func (mm *metadataManager) cleanup() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Clean up old confidence history
	for requestID, history := range mm.confidenceHistory {
		if len(history) > mm.config.MaxHistorySize {
			mm.confidenceHistory[requestID] = history[len(history)-mm.config.MaxHistorySize:]
		}
	}

	mm.logger.Debug("Metadata cleanup completed")
}

func generateID(prefix string) string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(bytes))
}

func (mm *metadataManager) GenerateEventID() string {
	return generateID("event")
}

func getDefaultConfig() *MetadataConfig {
	return &MetadataConfig{
		ConfidenceWeights: map[string]float64{
			"data_quality":       0.3,
			"source_reliability": 0.2,
			"validation":         0.2,
			"freshness":          0.15,
			"consistency":        0.15,
		},
		ConfidenceThresholds: map[string]float64{
			"very_high": 0.9,
			"high":      0.8,
			"medium":    0.7,
			"low":       0.6,
		},
		QualityWeights: map[string]float64{
			"data_quality":    0.4,
			"process_quality": 0.3,
			"output_quality":  0.3,
		},
		QualityThresholds: map[string]float64{
			"excellent": 0.9,
			"high":      0.8,
			"medium":    0.7,
			"low":       0.6,
		},
		ValidationRules: []ValidationRule{
			{
				RuleID:          "required_fields",
				RuleName:        "Required Fields Present",
				RuleDescription: "Check that all required fields are present",
				Status:          "passed",
				Severity:        "high",
				Confidence:      1.0,
			},
		},
		ComplianceFrameworks: []ComplianceFramework{
			{
				FrameworkID:   "gdpr",
				FrameworkName: "GDPR",
				Version:       "1.0",
				Status:        "compliant",
				Score:         1.0,
				Requirements:  10,
				Compliant:     10,
				NonCompliant:  0,
			},
		},
		MaxHistorySize:    1000,
		CleanupInterval:   1 * time.Hour,
		EnableCalibration: true,
		CalibrationFactor: 1.0,
	}
}
