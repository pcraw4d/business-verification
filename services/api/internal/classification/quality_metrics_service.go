package classification

import (
	"context"
	"fmt"
	"log"
	"time"

	"kyb-platform/internal/shared"
)

// QualityMetricsService provides comprehensive quality metrics for classification results
type QualityMetricsService struct {
	logger *log.Logger
	config *QualityMetricsConfig
}

// QualityMetricsConfig holds configuration for quality metrics calculation
type QualityMetricsConfig struct {
	// Quality thresholds
	HighQualityThreshold   float64 `json:"high_quality_threshold"`   // Threshold for high quality (0.8)
	MediumQualityThreshold float64 `json:"medium_quality_threshold"` // Threshold for medium quality (0.6)
	LowQualityThreshold    float64 `json:"low_quality_threshold"`    // Threshold for low quality (0.4)

	// Method agreement settings
	StrongAgreementThreshold float64 `json:"strong_agreement_threshold"` // Threshold for strong agreement (0.8)
	WeakAgreementThreshold   float64 `json:"weak_agreement_threshold"`   // Threshold for weak agreement (0.4)

	// Evidence settings
	StrongEvidenceThreshold float64 `json:"strong_evidence_threshold"` // Threshold for strong evidence (0.7)
	WeakEvidenceThreshold   float64 `json:"weak_evidence_threshold"`   // Threshold for weak evidence (0.3)

	// Data completeness settings
	CompleteDataThreshold   float64 `json:"complete_data_threshold"`   // Threshold for complete data (0.8)
	IncompleteDataThreshold float64 `json:"incomplete_data_threshold"` // Threshold for incomplete data (0.4)

	// Confidence settings
	HighConfidenceThreshold float64 `json:"high_confidence_threshold"` // Threshold for high confidence (0.8)
	LowConfidenceThreshold  float64 `json:"low_confidence_threshold"`  // Threshold for low confidence (0.4)
}

// DefaultQualityMetricsConfig returns the default configuration for quality metrics
func DefaultQualityMetricsConfig() *QualityMetricsConfig {
	return &QualityMetricsConfig{
		HighQualityThreshold:     0.8,
		MediumQualityThreshold:   0.6,
		LowQualityThreshold:      0.4,
		StrongAgreementThreshold: 0.8,
		WeakAgreementThreshold:   0.4,
		StrongEvidenceThreshold:  0.7,
		WeakEvidenceThreshold:    0.3,
		CompleteDataThreshold:    0.8,
		IncompleteDataThreshold:  0.4,
		HighConfidenceThreshold:  0.8,
		LowConfidenceThreshold:   0.4,
	}
}

// NewQualityMetricsService creates a new quality metrics service
func NewQualityMetricsService(logger *log.Logger) *QualityMetricsService {
	if logger == nil {
		logger = log.Default()
	}

	return &QualityMetricsService{
		logger: logger,
		config: DefaultQualityMetricsConfig(),
	}
}

// NewQualityMetricsServiceWithConfig creates a new quality metrics service with custom configuration
func NewQualityMetricsServiceWithConfig(logger *log.Logger, config *QualityMetricsConfig) *QualityMetricsService {
	if logger == nil {
		logger = log.Default()
	}

	if config == nil {
		config = DefaultQualityMetricsConfig()
	}

	return &QualityMetricsService{
		logger: logger,
		config: config,
	}
}

// ComprehensiveQualityMetrics represents comprehensive quality metrics for classification
type ComprehensiveQualityMetrics struct {
	OverallQuality float64 `json:"overall_quality"`
	QualityGrade   string  `json:"quality_grade"`
	QualityLevel   string  `json:"quality_level"`

	// Core quality components
	MethodAgreement       float64 `json:"method_agreement"`
	EvidenceStrength      float64 `json:"evidence_strength"`
	DataCompleteness      float64 `json:"data_completeness"`
	ConfidenceConsistency float64 `json:"confidence_consistency"`

	// Detailed metrics
	MethodMetrics     *MethodQualityMetrics     `json:"method_metrics"`
	EvidenceMetrics   *EvidenceQualityMetrics   `json:"evidence_metrics"`
	DataMetrics       *DataQualityMetrics       `json:"data_metrics"`
	ConfidenceMetrics *ConfidenceQualityMetrics `json:"confidence_metrics"`

	// Quality indicators
	QualityIndicators      []QualityIndicator `json:"quality_indicators"`
	QualityIssues          []QualityIssue     `json:"quality_issues"`
	QualityRecommendations []string           `json:"quality_recommendations"`

	// Performance metrics
	ProcessingTime    time.Duration       `json:"processing_time"`
	MethodPerformance []MethodPerformance `json:"method_performance"`

	// Metadata
	GeneratedAt time.Time `json:"generated_at"`
	RequestID   string    `json:"request_id"`
}

// MethodQualityMetrics represents quality metrics for classification methods
type MethodQualityMetrics struct {
	TotalMethods       int                   `json:"total_methods"`
	SuccessfulMethods  int                   `json:"successful_methods"`
	FailedMethods      int                   `json:"failed_methods"`
	SuccessRate        float64               `json:"success_rate"`
	AverageConfidence  float64               `json:"average_confidence"`
	ConfidenceVariance float64               `json:"confidence_variance"`
	MethodAgreement    float64               `json:"method_agreement"`
	AgreementLevel     string                `json:"agreement_level"`
	MethodDetails      []MethodQualityDetail `json:"method_details"`
}

// EvidenceQualityMetrics represents quality metrics for evidence
type EvidenceQualityMetrics struct {
	TotalEvidenceItems      int            `json:"total_evidence_items"`
	StrongEvidenceItems     int            `json:"strong_evidence_items"`
	WeakEvidenceItems       int            `json:"weak_evidence_items"`
	AverageEvidenceStrength float64        `json:"average_evidence_strength"`
	EvidenceDiversity       float64        `json:"evidence_diversity"`
	EvidenceConsistency     float64        `json:"evidence_consistency"`
	EvidenceTypes           map[string]int `json:"evidence_types"`
}

// DataQualityMetrics represents quality metrics for input data
type DataQualityMetrics struct {
	DataCompleteness float64  `json:"data_completeness"`
	DataQuality      string   `json:"data_quality"`
	AvailableFields  []string `json:"available_fields"`
	MissingFields    []string `json:"missing_fields"`
	DataConsistency  float64  `json:"data_consistency"`
	DataReliability  float64  `json:"data_reliability"`
}

// ConfidenceQualityMetrics represents quality metrics for confidence scores
type ConfidenceQualityMetrics struct {
	OverallConfidence     float64         `json:"overall_confidence"`
	ConfidenceRange       ConfidenceRange `json:"confidence_range"`
	ConfidenceVariance    float64         `json:"confidence_variance"`
	ConfidenceConsistency float64         `json:"confidence_consistency"`
	ConfidenceLevel       string          `json:"confidence_level"`
	UncertaintyLevel      string          `json:"uncertainty_level"`
	ReliabilityScore      float64         `json:"reliability_score"`
}

// QualityIndicator represents a quality indicator
type QualityIndicator struct {
	Indicator   string  `json:"indicator"`
	Value       float64 `json:"value"`
	Threshold   float64 `json:"threshold"`
	Status      string  `json:"status"` // "good", "warning", "critical"
	Description string  `json:"description"`
	Impact      string  `json:"impact"` // "positive", "negative", "neutral"
}

// QualityIssue represents a quality issue
type QualityIssue struct {
	Issue              string   `json:"issue"`
	Severity           string   `json:"severity"` // "low", "medium", "high", "critical"
	Description        string   `json:"description"`
	Recommendation     string   `json:"recommendation"`
	AffectedComponents []string `json:"affected_components"`
}

// MethodQualityDetail represents quality details for a specific method
type MethodQualityDetail struct {
	MethodName     string        `json:"method_name"`
	MethodType     string        `json:"method_type"`
	Success        bool          `json:"success"`
	Confidence     float64       `json:"confidence"`
	ProcessingTime time.Duration `json:"processing_time"`
	QualityScore   float64       `json:"quality_score"`
	Issues         []string      `json:"issues"`
	Strengths      []string      `json:"strengths"`
}

// MethodPerformance represents performance metrics for a method
type MethodPerformance struct {
	MethodName        string        `json:"method_name"`
	MethodType        string        `json:"method_type"`
	ProcessingTime    time.Duration `json:"processing_time"`
	SuccessRate       float64       `json:"success_rate"`
	AverageConfidence float64       `json:"average_confidence"`
	PerformanceScore  float64       `json:"performance_score"`
}

// CalculateComprehensiveQualityMetrics calculates comprehensive quality metrics for classification results
func (qms *QualityMetricsService) CalculateComprehensiveQualityMetrics(
	ctx context.Context,
	methodResults []shared.ClassificationMethodResult,
	primaryClassification *shared.IndustryClassification,
	request *shared.BusinessClassificationRequest,
) (*ComprehensiveQualityMetrics, error) {
	startTime := time.Now()
	requestID := qms.generateRequestID()

	qms.logger.Printf("📊 Calculating comprehensive quality metrics (request: %s)", requestID)

	// Step 1: Calculate method quality metrics
	methodMetrics := qms.calculateMethodQualityMetrics(methodResults)

	// Step 2: Calculate evidence quality metrics
	evidenceMetrics := qms.calculateEvidenceQualityMetrics(methodResults)

	// Step 3: Calculate data quality metrics
	dataMetrics := qms.calculateDataQualityMetrics(request, methodResults)

	// Step 4: Calculate confidence quality metrics
	confidenceMetrics := qms.calculateConfidenceQualityMetrics(methodResults, primaryClassification)

	// Step 5: Calculate overall quality
	overallQuality := qms.calculateOverallQuality(methodMetrics, evidenceMetrics, dataMetrics, confidenceMetrics)

	// Step 6: Generate quality indicators
	qualityIndicators := qms.generateQualityIndicators(methodMetrics, evidenceMetrics, dataMetrics, confidenceMetrics)

	// Step 7: Identify quality issues
	qualityIssues := qms.identifyQualityIssues(methodMetrics, evidenceMetrics, dataMetrics, confidenceMetrics)

	// Step 8: Generate quality recommendations
	qualityRecommendations := qms.generateQualityRecommendations(qualityIssues, methodMetrics, evidenceMetrics, dataMetrics, confidenceMetrics)

	// Step 9: Calculate method performance
	methodPerformance := qms.calculateMethodPerformance(methodResults)

	// Create comprehensive quality metrics
	metrics := &ComprehensiveQualityMetrics{
		OverallQuality:         overallQuality,
		QualityGrade:           qms.determineQualityGrade(overallQuality),
		QualityLevel:           qms.determineQualityLevel(overallQuality),
		MethodAgreement:        methodMetrics.MethodAgreement,
		EvidenceStrength:       evidenceMetrics.AverageEvidenceStrength,
		DataCompleteness:       dataMetrics.DataCompleteness,
		ConfidenceConsistency:  confidenceMetrics.ConfidenceConsistency,
		MethodMetrics:          methodMetrics,
		EvidenceMetrics:        evidenceMetrics,
		DataMetrics:            dataMetrics,
		ConfidenceMetrics:      confidenceMetrics,
		QualityIndicators:      qualityIndicators,
		QualityIssues:          qualityIssues,
		QualityRecommendations: qualityRecommendations,
		ProcessingTime:         time.Since(startTime),
		MethodPerformance:      methodPerformance,
		GeneratedAt:            time.Now(),
		RequestID:              requestID,
	}

	qms.logger.Printf("✅ Comprehensive quality metrics calculated: %.3f (Grade: %s) (request: %s)",
		overallQuality, metrics.QualityGrade, requestID)

	return metrics, nil
}

// calculateMethodQualityMetrics calculates quality metrics for classification methods
func (qms *QualityMetricsService) calculateMethodQualityMetrics(
	methodResults []shared.ClassificationMethodResult,
) *MethodQualityMetrics {
	totalMethods := len(methodResults)
	successfulMethods := 0
	var confidences []float64
	var methodDetails []MethodQualityDetail

	// Count successful methods and collect confidences
	for _, method := range methodResults {
		if method.Success {
			successfulMethods++
			confidences = append(confidences, method.Confidence)
		}

		// Create method quality detail
		detail := MethodQualityDetail{
			MethodName:     method.MethodName,
			MethodType:     method.MethodType,
			Success:        method.Success,
			Confidence:     method.Confidence,
			ProcessingTime: method.ProcessingTime,
			QualityScore:   qms.calculateMethodQualityScore(method),
			Issues:         qms.identifyMethodIssues(method),
			Strengths:      qms.identifyMethodStrengths(method),
		}
		methodDetails = append(methodDetails, detail)
	}

	// Calculate metrics
	successRate := float64(successfulMethods) / float64(totalMethods)
	averageConfidence := qms.calculateAverage(confidences)
	confidenceVariance := qms.calculateVariance(confidences)
	methodAgreement := qms.calculateMethodAgreement(methodResults)
	agreementLevel := qms.determineAgreementLevel(methodAgreement)

	return &MethodQualityMetrics{
		TotalMethods:       totalMethods,
		SuccessfulMethods:  successfulMethods,
		FailedMethods:      totalMethods - successfulMethods,
		SuccessRate:        successRate,
		AverageConfidence:  averageConfidence,
		ConfidenceVariance: confidenceVariance,
		MethodAgreement:    methodAgreement,
		AgreementLevel:     agreementLevel,
		MethodDetails:      methodDetails,
	}
}

// calculateEvidenceQualityMetrics calculates quality metrics for evidence
func (qms *QualityMetricsService) calculateEvidenceQualityMetrics(
	methodResults []shared.ClassificationMethodResult,
) *EvidenceQualityMetrics {
	var allEvidence []string
	evidenceTypes := make(map[string]int)
	strongEvidenceItems := 0
	weakEvidenceItems := 0

	// Collect evidence from all methods
	for _, method := range methodResults {
		if !method.Success {
			continue
		}

		// Count evidence items
		evidenceCount := len(method.Evidence)
		allEvidence = append(allEvidence, method.Evidence...)

		// Categorize evidence strength
		if method.Confidence >= qms.config.StrongEvidenceThreshold {
			strongEvidenceItems += evidenceCount
		} else if method.Confidence < qms.config.WeakEvidenceThreshold {
			weakEvidenceItems += evidenceCount
		}

		// Count evidence types
		evidenceTypes[method.MethodType] += evidenceCount
	}

	totalEvidenceItems := len(allEvidence)
	averageEvidenceStrength := qms.calculateAverageEvidenceStrength(methodResults)
	evidenceDiversity := qms.calculateEvidenceDiversity(evidenceTypes)
	evidenceConsistency := qms.calculateEvidenceConsistency(methodResults)

	return &EvidenceQualityMetrics{
		TotalEvidenceItems:      totalEvidenceItems,
		StrongEvidenceItems:     strongEvidenceItems,
		WeakEvidenceItems:       weakEvidenceItems,
		AverageEvidenceStrength: averageEvidenceStrength,
		EvidenceDiversity:       evidenceDiversity,
		EvidenceConsistency:     evidenceConsistency,
		EvidenceTypes:           evidenceTypes,
	}
}

// calculateDataQualityMetrics calculates quality metrics for input data
func (qms *QualityMetricsService) calculateDataQualityMetrics(
	request *shared.BusinessClassificationRequest,
	methodResults []shared.ClassificationMethodResult,
) *DataQualityMetrics {
	// Calculate data completeness
	availableFields := qms.getAvailableFields(request)
	missingFields := qms.getMissingFields(request)
	dataCompleteness := float64(len(availableFields)) / float64(len(availableFields)+len(missingFields))

	// Determine data quality level
	dataQuality := qms.determineDataQuality(dataCompleteness)

	// Calculate data consistency and reliability
	dataConsistency := qms.calculateDataConsistency(request)
	dataReliability := qms.calculateDataReliability(methodResults)

	return &DataQualityMetrics{
		DataCompleteness: dataCompleteness,
		DataQuality:      dataQuality,
		AvailableFields:  availableFields,
		MissingFields:    missingFields,
		DataConsistency:  dataConsistency,
		DataReliability:  dataReliability,
	}
}

// calculateConfidenceQualityMetrics calculates quality metrics for confidence scores
func (qms *QualityMetricsService) calculateConfidenceQualityMetrics(
	methodResults []shared.ClassificationMethodResult,
	primaryClassification *shared.IndustryClassification,
) *ConfidenceQualityMetrics {
	var confidences []float64
	for _, method := range methodResults {
		if method.Success {
			confidences = append(confidences, method.Confidence)
		}
	}

	// Calculate confidence metrics
	overallConfidence := primaryClassification.ConfidenceScore
	confidenceRange := qms.calculateConfidenceRange(confidences)
	confidenceVariance := qms.calculateVariance(confidences)
	confidenceConsistency := 1.0 - confidenceVariance
	confidenceLevel := qms.determineConfidenceLevel(overallConfidence)
	uncertaintyLevel := qms.determineUncertaintyLevel(confidences, overallConfidence)
	reliabilityScore := qms.calculateReliabilityScore(confidences, methodResults)

	return &ConfidenceQualityMetrics{
		OverallConfidence:     overallConfidence,
		ConfidenceRange:       confidenceRange,
		ConfidenceVariance:    confidenceVariance,
		ConfidenceConsistency: confidenceConsistency,
		ConfidenceLevel:       confidenceLevel,
		UncertaintyLevel:      uncertaintyLevel,
		ReliabilityScore:      reliabilityScore,
	}
}

// calculateOverallQuality calculates the overall quality score
func (qms *QualityMetricsService) calculateOverallQuality(
	methodMetrics *MethodQualityMetrics,
	evidenceMetrics *EvidenceQualityMetrics,
	dataMetrics *DataQualityMetrics,
	confidenceMetrics *ConfidenceQualityMetrics,
) float64 {
	// Weighted combination of quality components
	methodQuality := (methodMetrics.SuccessRate * 0.3) + (methodMetrics.MethodAgreement * 0.2) + (1.0 - methodMetrics.ConfidenceVariance*0.1)
	evidenceQuality := evidenceMetrics.AverageEvidenceStrength * 0.2
	dataQuality := dataMetrics.DataCompleteness * 0.15
	confidenceQuality := confidenceMetrics.ConfidenceConsistency * 0.15

	overallQuality := methodQuality + evidenceQuality + dataQuality + confidenceQuality

	// Ensure quality is within bounds
	if overallQuality > 1.0 {
		overallQuality = 1.0
	}
	if overallQuality < 0.0 {
		overallQuality = 0.0
	}

	return overallQuality
}

// Helper methods

func (qms *QualityMetricsService) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	var sum float64
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func (qms *QualityMetricsService) calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	mean := qms.calculateAverage(values)
	var variance float64
	for _, value := range values {
		variance += (value - mean) * (value - mean)
	}
	return variance / float64(len(values))
}

func (qms *QualityMetricsService) calculateMethodAgreement(methodResults []shared.ClassificationMethodResult) float64 {
	// Count successful methods
	successfulMethods := 0
	industryCounts := make(map[string]int)

	for _, method := range methodResults {
		if method.Success && method.Result != nil {
			successfulMethods++
			industryCounts[method.Result.IndustryName]++
		}
	}

	if successfulMethods == 0 {
		return 0.0
	}

	// Find the most common industry
	maxCount := 0
	for _, count := range industryCounts {
		if count > maxCount {
			maxCount = count
		}
	}

	return float64(maxCount) / float64(successfulMethods)
}

func (qms *QualityMetricsService) determineAgreementLevel(agreement float64) string {
	if agreement >= qms.config.StrongAgreementThreshold {
		return "strong"
	} else if agreement >= qms.config.WeakAgreementThreshold {
		return "moderate"
	} else {
		return "weak"
	}
}

func (qms *QualityMetricsService) calculateMethodQualityScore(method shared.ClassificationMethodResult) float64 {
	score := 0.0

	if method.Success {
		score += 0.5                     // Base score for success
		score += method.Confidence * 0.3 // Confidence contribution

		// Processing time contribution (faster is better)
		if method.ProcessingTime < 2*time.Second {
			score += 0.1
		} else if method.ProcessingTime > 10*time.Second {
			score -= 0.1
		}

		// Evidence contribution
		evidenceScore := float64(len(method.Evidence)) * 0.1
		if evidenceScore > 0.1 {
			evidenceScore = 0.1
		}
		score += evidenceScore
	}

	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

func (qms *QualityMetricsService) identifyMethodIssues(method shared.ClassificationMethodResult) []string {
	var issues []string

	if !method.Success {
		issues = append(issues, "Method failed")
	}

	if method.Confidence < qms.config.LowConfidenceThreshold {
		issues = append(issues, "Low confidence score")
	}

	if method.ProcessingTime > 10*time.Second {
		issues = append(issues, "Slow processing time")
	}

	if len(method.Evidence) == 0 {
		issues = append(issues, "No evidence provided")
	}

	return issues
}

func (qms *QualityMetricsService) identifyMethodStrengths(method shared.ClassificationMethodResult) []string {
	var strengths []string

	if method.Success {
		strengths = append(strengths, "Successful execution")
	}

	if method.Confidence >= qms.config.HighConfidenceThreshold {
		strengths = append(strengths, "High confidence score")
	}

	if method.ProcessingTime < 2*time.Second {
		strengths = append(strengths, "Fast processing time")
	}

	if len(method.Evidence) > 3 {
		strengths = append(strengths, "Strong evidence")
	}

	return strengths
}

func (qms *QualityMetricsService) calculateAverageEvidenceStrength(methodResults []shared.ClassificationMethodResult) float64 {
	var totalStrength float64
	var methodCount int

	for _, method := range methodResults {
		if method.Success {
			evidenceStrength := float64(len(method.Evidence)) * method.Confidence
			totalStrength += evidenceStrength
			methodCount++
		}
	}

	if methodCount == 0 {
		return 0.0
	}

	return totalStrength / float64(methodCount)
}

func (qms *QualityMetricsService) calculateEvidenceDiversity(evidenceTypes map[string]int) float64 {
	if len(evidenceTypes) == 0 {
		return 0.0
	}

	// Calculate diversity based on number of different evidence types
	maxTypes := 3 // keyword, ml, description
	diversity := float64(len(evidenceTypes)) / float64(maxTypes)

	if diversity > 1.0 {
		diversity = 1.0
	}

	return diversity
}

func (qms *QualityMetricsService) calculateEvidenceConsistency(methodResults []shared.ClassificationMethodResult) float64 {
	var confidences []float64
	for _, method := range methodResults {
		if method.Success {
			confidences = append(confidences, method.Confidence)
		}
	}

	if len(confidences) == 0 {
		return 0.0
	}

	// Consistency is inverse of variance
	variance := qms.calculateVariance(confidences)
	consistency := 1.0 - variance

	if consistency < 0.0 {
		consistency = 0.0
	}

	return consistency
}

func (qms *QualityMetricsService) getAvailableFields(request *shared.BusinessClassificationRequest) []string {
	var fields []string

	if request.BusinessName != "" {
		fields = append(fields, "business_name")
	}
	if request.Description != "" {
		fields = append(fields, "description")
	}
	if request.WebsiteURL != "" {
		fields = append(fields, "website_url")
	}
	if request.Address != "" {
		fields = append(fields, "address")
	}
	if len(request.Keywords) > 0 {
		fields = append(fields, "keywords")
	}

	return fields
}

func (qms *QualityMetricsService) getMissingFields(request *shared.BusinessClassificationRequest) []string {
	var missing []string

	if request.BusinessName == "" {
		missing = append(missing, "business_name")
	}
	if request.Description == "" {
		missing = append(missing, "description")
	}
	if request.WebsiteURL == "" {
		missing = append(missing, "website_url")
	}
	if request.Address == "" {
		missing = append(missing, "address")
	}
	if len(request.Keywords) == 0 {
		missing = append(missing, "keywords")
	}

	return missing
}

func (qms *QualityMetricsService) determineDataQuality(completeness float64) string {
	if completeness >= qms.config.CompleteDataThreshold {
		return "high"
	} else if completeness >= qms.config.IncompleteDataThreshold {
		return "medium"
	} else {
		return "low"
	}
}

func (qms *QualityMetricsService) calculateDataConsistency(request *shared.BusinessClassificationRequest) float64 {
	// Simple consistency check based on field presence and format
	consistency := 0.0

	if request.BusinessName != "" && len(request.BusinessName) > 2 {
		consistency += 0.3
	}
	if request.Description != "" && len(request.Description) > 10 {
		consistency += 0.3
	}
	if request.WebsiteURL != "" && qms.isValidURL(request.WebsiteURL) {
		consistency += 0.2
	}
	if len(request.Keywords) > 0 {
		consistency += 0.2
	}

	return consistency
}

func (qms *QualityMetricsService) calculateDataReliability(methodResults []shared.ClassificationMethodResult) float64 {
	// Reliability based on method success rate and consistency
	successfulMethods := 0
	for _, method := range methodResults {
		if method.Success {
			successfulMethods++
		}
	}

	successRate := float64(successfulMethods) / float64(len(methodResults))

	// Add consistency factor
	var confidences []float64
	for _, method := range methodResults {
		if method.Success {
			confidences = append(confidences, method.Confidence)
		}
	}

	consistency := 1.0 - qms.calculateVariance(confidences)

	reliability := (successRate * 0.7) + (consistency * 0.3)

	if reliability > 1.0 {
		reliability = 1.0
	}

	return reliability
}

func (qms *QualityMetricsService) calculateConfidenceRange(confidences []float64) ConfidenceRange {
	if len(confidences) == 0 {
		return ConfidenceRange{Min: 0.0, Max: 0.0, Median: 0.0, Mean: 0.0}
	}

	min := confidences[0]
	max := confidences[0]
	var sum float64

	for _, conf := range confidences {
		if conf < min {
			min = conf
		}
		if conf > max {
			max = conf
		}
		sum += conf
	}

	mean := sum / float64(len(confidences))
	median := qms.calculateMedian(confidences)

	return ConfidenceRange{
		Min:    min,
		Max:    max,
		Median: median,
		Mean:   mean,
	}
}

func (qms *QualityMetricsService) calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	// Create a copy and sort
	sorted := make([]float64, len(values))
	copy(sorted, values)

	// Simple bubble sort for small arrays
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2.0
	}
	return sorted[n/2]
}

func (qms *QualityMetricsService) determineConfidenceLevel(confidence float64) string {
	if confidence >= qms.config.HighConfidenceThreshold {
		return "high"
	} else if confidence >= qms.config.LowConfidenceThreshold {
		return "medium"
	} else {
		return "low"
	}
}

func (qms *QualityMetricsService) determineUncertaintyLevel(confidences []float64, overallConfidence float64) string {
	if overallConfidence >= 0.8 {
		return "low"
	} else if overallConfidence >= 0.6 {
		return "medium"
	} else {
		return "high"
	}
}

func (qms *QualityMetricsService) calculateReliabilityScore(confidences []float64, methodResults []shared.ClassificationMethodResult) float64 {
	if len(confidences) == 0 {
		return 0.0
	}

	// Base reliability on confidence consistency and method success rate
	successRate := float64(len(confidences)) / float64(len(methodResults))

	// Calculate confidence variance
	variance := qms.calculateVariance(confidences)
	consistency := 1.0 - variance

	// Combine success rate and consistency
	reliability := (successRate * 0.6) + (consistency * 0.4)

	if reliability > 1.0 {
		reliability = 1.0
	}

	return reliability
}

func (qms *QualityMetricsService) determineQualityGrade(overallQuality float64) string {
	switch {
	case overallQuality >= 0.9:
		return "A"
	case overallQuality >= 0.8:
		return "B"
	case overallQuality >= 0.7:
		return "C"
	case overallQuality >= 0.6:
		return "D"
	default:
		return "F"
	}
}

func (qms *QualityMetricsService) determineQualityLevel(overallQuality float64) string {
	if overallQuality >= qms.config.HighQualityThreshold {
		return "high"
	} else if overallQuality >= qms.config.MediumQualityThreshold {
		return "medium"
	} else if overallQuality >= qms.config.LowQualityThreshold {
		return "low"
	} else {
		return "very_low"
	}
}

func (qms *QualityMetricsService) generateQualityIndicators(
	methodMetrics *MethodQualityMetrics,
	evidenceMetrics *EvidenceQualityMetrics,
	dataMetrics *DataQualityMetrics,
	confidenceMetrics *ConfidenceQualityMetrics,
) []QualityIndicator {
	var indicators []QualityIndicator

	// Method success rate indicator
	indicators = append(indicators, QualityIndicator{
		Indicator:   "method_success_rate",
		Value:       methodMetrics.SuccessRate,
		Threshold:   0.8,
		Status:      qms.getIndicatorStatus(methodMetrics.SuccessRate, 0.8),
		Description: "Percentage of classification methods that succeeded",
		Impact:      "positive",
	})

	// Method agreement indicator
	indicators = append(indicators, QualityIndicator{
		Indicator:   "method_agreement",
		Value:       methodMetrics.MethodAgreement,
		Threshold:   0.7,
		Status:      qms.getIndicatorStatus(methodMetrics.MethodAgreement, 0.7),
		Description: "Agreement between different classification methods",
		Impact:      "positive",
	})

	// Evidence strength indicator
	indicators = append(indicators, QualityIndicator{
		Indicator:   "evidence_strength",
		Value:       evidenceMetrics.AverageEvidenceStrength,
		Threshold:   0.6,
		Status:      qms.getIndicatorStatus(evidenceMetrics.AverageEvidenceStrength, 0.6),
		Description: "Average strength of supporting evidence",
		Impact:      "positive",
	})

	// Data completeness indicator
	indicators = append(indicators, QualityIndicator{
		Indicator:   "data_completeness",
		Value:       dataMetrics.DataCompleteness,
		Threshold:   0.7,
		Status:      qms.getIndicatorStatus(dataMetrics.DataCompleteness, 0.7),
		Description: "Completeness of input data",
		Impact:      "positive",
	})

	// Confidence consistency indicator
	indicators = append(indicators, QualityIndicator{
		Indicator:   "confidence_consistency",
		Value:       confidenceMetrics.ConfidenceConsistency,
		Threshold:   0.8,
		Status:      qms.getIndicatorStatus(confidenceMetrics.ConfidenceConsistency, 0.8),
		Description: "Consistency of confidence scores across methods",
		Impact:      "positive",
	})

	return indicators
}

func (qms *QualityMetricsService) identifyQualityIssues(
	methodMetrics *MethodQualityMetrics,
	evidenceMetrics *EvidenceQualityMetrics,
	dataMetrics *DataQualityMetrics,
	confidenceMetrics *ConfidenceQualityMetrics,
) []QualityIssue {
	var issues []QualityIssue

	// Method success rate issues
	if methodMetrics.SuccessRate < 0.8 {
		severity := "medium"
		if methodMetrics.SuccessRate < 0.5 {
			severity = "high"
		}
		issues = append(issues, QualityIssue{
			Issue:              "low_method_success_rate",
			Severity:           severity,
			Description:        fmt.Sprintf("Only %.1f%% of classification methods succeeded", methodMetrics.SuccessRate*100),
			Recommendation:     "Investigate and improve failed classification methods",
			AffectedComponents: []string{"classification_methods"},
		})
	}

	// Method agreement issues
	if methodMetrics.MethodAgreement < 0.6 {
		issues = append(issues, QualityIssue{
			Issue:              "low_method_agreement",
			Severity:           "medium",
			Description:        fmt.Sprintf("Low agreement between methods: %.1f%%", methodMetrics.MethodAgreement*100),
			Recommendation:     "Improve classification algorithms or gather more data",
			AffectedComponents: []string{"classification_methods", "data_quality"},
		})
	}

	// Evidence strength issues
	if evidenceMetrics.AverageEvidenceStrength < 0.5 {
		issues = append(issues, QualityIssue{
			Issue:              "weak_evidence",
			Severity:           "medium",
			Description:        fmt.Sprintf("Weak evidence strength: %.1f%%", evidenceMetrics.AverageEvidenceStrength*100),
			Recommendation:     "Gather more detailed business information",
			AffectedComponents: []string{"evidence_collection"},
		})
	}

	// Data completeness issues
	if dataMetrics.DataCompleteness < 0.6 {
		issues = append(issues, QualityIssue{
			Issue:              "incomplete_data",
			Severity:           "medium",
			Description:        fmt.Sprintf("Incomplete input data: %.1f%%", dataMetrics.DataCompleteness*100),
			Recommendation:     "Collect more comprehensive business data",
			AffectedComponents: []string{"input_data"},
		})
	}

	// Confidence consistency issues
	if confidenceMetrics.ConfidenceConsistency < 0.7 {
		issues = append(issues, QualityIssue{
			Issue:              "inconsistent_confidence",
			Severity:           "low",
			Description:        fmt.Sprintf("Inconsistent confidence scores: %.1f%%", confidenceMetrics.ConfidenceConsistency*100),
			Recommendation:     "Review and calibrate confidence scoring algorithms",
			AffectedComponents: []string{"confidence_scoring"},
		})
	}

	return issues
}

func (qms *QualityMetricsService) generateQualityRecommendations(
	qualityIssues []QualityIssue,
	methodMetrics *MethodQualityMetrics,
	evidenceMetrics *EvidenceQualityMetrics,
	dataMetrics *DataQualityMetrics,
	confidenceMetrics *ConfidenceQualityMetrics,
) []string {
	var recommendations []string

	// Generate recommendations based on issues
	for _, issue := range qualityIssues {
		recommendations = append(recommendations, issue.Recommendation)
	}

	// Generate general recommendations
	if methodMetrics.SuccessRate < 0.9 {
		recommendations = append(recommendations, "Improve classification method reliability")
	}

	if evidenceMetrics.AverageEvidenceStrength < 0.7 {
		recommendations = append(recommendations, "Enhance evidence collection and analysis")
	}

	if dataMetrics.DataCompleteness < 0.8 {
		recommendations = append(recommendations, "Collect more comprehensive business data")
	}

	if confidenceMetrics.ConfidenceConsistency < 0.8 {
		recommendations = append(recommendations, "Improve confidence scoring consistency")
	}

	// Remove duplicates
	uniqueRecommendations := qms.removeDuplicateRecommendations(recommendations)

	// If no specific issues, provide general recommendations
	if len(uniqueRecommendations) == 0 {
		uniqueRecommendations = append(uniqueRecommendations, "Classification quality is good - continue monitoring")
	}

	return uniqueRecommendations
}

func (qms *QualityMetricsService) calculateMethodPerformance(methodResults []shared.ClassificationMethodResult) []MethodPerformance {
	var performance []MethodPerformance

	for _, method := range methodResults {
		perf := MethodPerformance{
			MethodName:        method.MethodName,
			MethodType:        method.MethodType,
			ProcessingTime:    method.ProcessingTime,
			SuccessRate:       qms.boolToFloat(method.Success),
			AverageConfidence: method.Confidence,
			PerformanceScore:  qms.calculateMethodPerformanceScore(method),
		}
		performance = append(performance, perf)
	}

	return performance
}

func (qms *QualityMetricsService) getIndicatorStatus(value, threshold float64) string {
	if value >= threshold {
		return "good"
	} else if value >= threshold*0.8 {
		return "warning"
	} else {
		return "critical"
	}
}

func (qms *QualityMetricsService) isValidURL(url string) bool {
	// Simple URL validation
	return len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://")
}

func (qms *QualityMetricsService) removeDuplicateRecommendations(recommendations []string) []string {
	seen := make(map[string]bool)
	var unique []string

	for _, rec := range recommendations {
		if !seen[rec] {
			seen[rec] = true
			unique = append(unique, rec)
		}
	}

	return unique
}

func (qms *QualityMetricsService) boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func (qms *QualityMetricsService) calculateMethodPerformanceScore(method shared.ClassificationMethodResult) float64 {
	score := 0.0

	if method.Success {
		score += 0.4
	}

	score += method.Confidence * 0.3

	// Processing time score (faster is better)
	if method.ProcessingTime < 1*time.Second {
		score += 0.2
	} else if method.ProcessingTime < 3*time.Second {
		score += 0.1
	}

	// Evidence score
	evidenceScore := float64(len(method.Evidence)) * 0.1
	if evidenceScore > 0.1 {
		evidenceScore = 0.1
	}
	score += evidenceScore

	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (qms *QualityMetricsService) generateRequestID() string {
	return fmt.Sprintf("quality_metrics_%d", time.Now().UnixNano())
}
