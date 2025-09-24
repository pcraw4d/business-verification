package classification

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"kyb-platform/internal/shared"
)

// ReasoningEngine provides sophisticated classification reasoning and evidence generation
type ReasoningEngine struct {
	logger *log.Logger
	config *ReasoningConfig
}

// ReasoningConfig holds configuration for the reasoning engine
type ReasoningConfig struct {
	// Evidence generation settings
	MaxEvidenceItems    int     `json:"max_evidence_items"`    // Maximum number of evidence items to include
	MinEvidenceStrength float64 `json:"min_evidence_strength"` // Minimum strength for evidence inclusion
	EvidenceWeight      float64 `json:"evidence_weight"`       // Weight for evidence in reasoning

	// Reasoning settings
	IncludeMethodDetails       bool `json:"include_method_details"`       // Include detailed method information
	IncludeConfidenceBreakdown bool `json:"include_confidence_breakdown"` // Include confidence breakdown
	IncludeQualityMetrics      bool `json:"include_quality_metrics"`      // Include quality metrics in reasoning

	// Language settings
	Language    string `json:"language"`     // Language for reasoning text
	DetailLevel string `json:"detail_level"` // "brief", "standard", "detailed"
}

// DefaultReasoningConfig returns the default configuration for the reasoning engine
func DefaultReasoningConfig() *ReasoningConfig {
	return &ReasoningConfig{
		MaxEvidenceItems:           10,
		MinEvidenceStrength:        0.1,
		EvidenceWeight:             0.3,
		IncludeMethodDetails:       true,
		IncludeConfidenceBreakdown: true,
		IncludeQualityMetrics:      true,
		Language:                   "en",
		DetailLevel:                "standard",
	}
}

// NewReasoningEngine creates a new reasoning engine
func NewReasoningEngine(logger *log.Logger) *ReasoningEngine {
	if logger == nil {
		logger = log.Default()
	}

	return &ReasoningEngine{
		logger: logger,
		config: DefaultReasoningConfig(),
	}
}

// NewReasoningEngineWithConfig creates a new reasoning engine with custom configuration
func NewReasoningEngineWithConfig(logger *log.Logger, config *ReasoningConfig) *ReasoningEngine {
	if logger == nil {
		logger = log.Default()
	}

	if config == nil {
		config = DefaultReasoningConfig()
	}

	return &ReasoningEngine{
		logger: logger,
		config: config,
	}
}

// ClassificationReasoning represents the reasoning and evidence for a classification
type ClassificationReasoning struct {
	Summary            string              `json:"summary"`
	DetailedReasoning  string              `json:"detailed_reasoning"`
	Evidence           []EvidenceItem      `json:"evidence"`
	MethodBreakdown    []MethodReasoning   `json:"method_breakdown"`
	ConfidenceAnalysis *ConfidenceAnalysis `json:"confidence_analysis"`
	QualityAssessment  *QualityAssessment  `json:"quality_assessment"`
	Recommendations    []string            `json:"recommendations"`
	GeneratedAt        time.Time           `json:"generated_at"`
	ProcessingTime     time.Duration       `json:"processing_time"`
}

// EvidenceItem represents a piece of evidence supporting the classification
type EvidenceItem struct {
	Type        string   `json:"type"`        // "keyword", "ml_prediction", "description", "pattern"
	Strength    float64  `json:"strength"`    // Strength of the evidence (0.0-1.0)
	Description string   `json:"description"` // Human-readable description
	Source      string   `json:"source"`      // Source of the evidence
	Keywords    []string `json:"keywords"`    // Keywords associated with the evidence
	Confidence  float64  `json:"confidence"`  // Confidence in this evidence
	Relevance   float64  `json:"relevance"`   // Relevance to the classification
}

// MethodReasoning represents reasoning for a specific classification method
type MethodReasoning struct {
	MethodName     string         `json:"method_name"`
	MethodType     string         `json:"method_type"`
	Success        bool           `json:"success"`
	Confidence     float64        `json:"confidence"`
	Reasoning      string         `json:"reasoning"`
	Evidence       []EvidenceItem `json:"evidence"`
	ProcessingTime time.Duration  `json:"processing_time"`
	Error          string         `json:"error,omitempty"`
}

// ConfidenceAnalysis represents analysis of confidence scores
type ConfidenceAnalysis struct {
	OverallConfidence float64            `json:"overall_confidence"`
	ConfidenceRange   ConfidenceRange    `json:"confidence_range"`
	MethodConfidences map[string]float64 `json:"method_confidences"`
	ConfidenceFactors []ConfidenceFactor `json:"confidence_factors"`
	UncertaintyLevel  string             `json:"uncertainty_level"`
	ReliabilityScore  float64            `json:"reliability_score"`
}

// ConfidenceRange represents the range of confidence scores
type ConfidenceRange struct {
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Median float64 `json:"median"`
	Mean   float64 `json:"mean"`
}

// ConfidenceFactor represents a factor affecting confidence
type ConfidenceFactor struct {
	Factor      string  `json:"factor"`
	Impact      float64 `json:"impact"` // Positive or negative impact on confidence
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
}

// QualityAssessment represents assessment of classification quality
type QualityAssessment struct {
	OverallQuality   float64         `json:"overall_quality"`
	QualityGrade     string          `json:"quality_grade"`
	QualityFactors   []QualityFactor `json:"quality_factors"`
	DataCompleteness float64         `json:"data_completeness"`
	MethodAgreement  float64         `json:"method_agreement"`
	EvidenceStrength float64         `json:"evidence_strength"`
	Recommendations  []string        `json:"recommendations"`
}

// QualityFactor represents a factor affecting quality
type QualityFactor struct {
	Factor      string  `json:"factor"`
	Score       float64 `json:"score"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"` // "positive", "negative", "neutral"
}

// GenerateReasoning generates comprehensive reasoning and evidence for a classification
func (re *ReasoningEngine) GenerateReasoning(
	ctx context.Context,
	businessName string,
	methodResults []shared.ClassificationMethodResult,
	primaryClassification *shared.IndustryClassification,
	qualityMetrics *shared.ClassificationQuality,
) (*ClassificationReasoning, error) {
	startTime := time.Now()
	requestID := re.generateRequestID()

	re.logger.Printf("🧠 Generating classification reasoning for: %s (request: %s)", businessName, requestID)

	// Step 1: Generate evidence
	evidence := re.generateEvidence(methodResults, primaryClassification)

	// Step 2: Generate method breakdown reasoning
	methodBreakdown := re.generateMethodBreakdown(methodResults)

	// Step 3: Analyze confidence
	confidenceAnalysis := re.analyzeConfidence(methodResults, primaryClassification)

	// Step 4: Assess quality
	qualityAssessment := re.assessQuality(qualityMetrics, methodResults)

	// Step 5: Generate summary and detailed reasoning
	summary := re.generateSummary(businessName, primaryClassification, methodResults)
	detailedReasoning := re.generateDetailedReasoning(businessName, primaryClassification, methodResults, evidence, confidenceAnalysis, qualityAssessment)

	// Step 6: Generate recommendations
	recommendations := re.generateRecommendations(methodResults, confidenceAnalysis, qualityAssessment)

	// Create final reasoning result
	reasoning := &ClassificationReasoning{
		Summary:            summary,
		DetailedReasoning:  detailedReasoning,
		Evidence:           evidence,
		MethodBreakdown:    methodBreakdown,
		ConfidenceAnalysis: confidenceAnalysis,
		QualityAssessment:  qualityAssessment,
		Recommendations:    recommendations,
		GeneratedAt:        time.Now(),
		ProcessingTime:     time.Since(startTime),
	}

	re.logger.Printf("✅ Classification reasoning generated (request: %s)", requestID)

	return reasoning, nil
}

// generateEvidence generates evidence items from method results
func (re *ReasoningEngine) generateEvidence(
	methodResults []shared.ClassificationMethodResult,
	primaryClassification *shared.IndustryClassification,
) []EvidenceItem {
	var evidence []EvidenceItem

	for _, method := range methodResults {
		if !method.Success || method.Result == nil {
			continue
		}

		// Generate evidence based on method type
		switch method.MethodType {
		case "keyword":
			evidence = append(evidence, re.generateKeywordEvidence(method)...)
		case "ml":
			evidence = append(evidence, re.generateMLEvidence(method)...)
		case "description":
			evidence = append(evidence, re.generateDescriptionEvidence(method)...)
		}
	}

	// Sort evidence by strength and relevance
	sort.Slice(evidence, func(i, j int) bool {
		scoreI := evidence[i].Strength * evidence[i].Relevance * evidence[i].Confidence
		scoreJ := evidence[j].Strength * evidence[j].Relevance * evidence[j].Confidence
		return scoreI > scoreJ
	})

	// Limit to max evidence items
	if len(evidence) > re.config.MaxEvidenceItems {
		evidence = evidence[:re.config.MaxEvidenceItems]
	}

	return evidence
}

// generateKeywordEvidence generates evidence from keyword-based classification
func (re *ReasoningEngine) generateKeywordEvidence(method shared.ClassificationMethodResult) []EvidenceItem {
	var evidence []EvidenceItem

	if len(method.Keywords) > 0 {
		// Create evidence for keyword matches
		evidence = append(evidence, EvidenceItem{
			Type:        "keyword",
			Strength:    method.Confidence,
			Description: fmt.Sprintf("Keyword matching identified %d relevant keywords: %s", len(method.Keywords), strings.Join(method.Keywords, ", ")),
			Source:      "keyword_classification",
			Keywords:    method.Keywords,
			Confidence:  method.Confidence,
			Relevance:   0.8, // Keywords are highly relevant
		})
	}

	// Add evidence from method result
	if method.Result != nil && method.Result.Evidence != "" {
		evidence = append(evidence, EvidenceItem{
			Type:        "pattern",
			Strength:    method.Confidence * 0.8,
			Description: method.Result.Evidence,
			Source:      "keyword_classification",
			Keywords:    method.Keywords,
			Confidence:  method.Confidence,
			Relevance:   0.7,
		})
	}

	return evidence
}

// generateMLEvidence generates evidence from ML-based classification
func (re *ReasoningEngine) generateMLEvidence(method shared.ClassificationMethodResult) []EvidenceItem {
	var evidence []EvidenceItem

	if method.Result != nil {
		evidence = append(evidence, EvidenceItem{
			Type:        "ml_prediction",
			Strength:    method.Confidence,
			Description: fmt.Sprintf("Machine learning model predicted '%s' with %.1f%% confidence", method.Result.IndustryName, method.Confidence*100),
			Source:      "ml_classification",
			Keywords:    []string{"machine_learning", "prediction"},
			Confidence:  method.Confidence,
			Relevance:   0.9, // ML predictions are highly relevant
		})

		// Add model-specific evidence if available
		if method.Result.Metadata != nil {
			if modelID, exists := method.Result.Metadata["model_id"]; exists {
				evidence = append(evidence, EvidenceItem{
					Type:        "model_info",
					Strength:    method.Confidence * 0.6,
					Description: fmt.Sprintf("Classification based on %s model", modelID),
					Source:      "ml_classification",
					Keywords:    []string{"model", modelID.(string)},
					Confidence:  method.Confidence,
					Relevance:   0.5,
				})
			}
		}
	}

	return evidence
}

// generateDescriptionEvidence generates evidence from description-based classification
func (re *ReasoningEngine) generateDescriptionEvidence(method shared.ClassificationMethodResult) []EvidenceItem {
	var evidence []EvidenceItem

	if method.Result != nil {
		evidence = append(evidence, EvidenceItem{
			Type:        "description",
			Strength:    method.Confidence,
			Description: fmt.Sprintf("Description analysis identified '%s' based on content patterns", method.Result.IndustryName),
			Source:      "description_classification",
			Keywords:    method.Keywords,
			Confidence:  method.Confidence,
			Relevance:   0.6, // Description analysis is moderately relevant
		})
	}

	return evidence
}

// generateMethodBreakdown generates reasoning for each classification method
func (re *ReasoningEngine) generateMethodBreakdown(methodResults []shared.ClassificationMethodResult) []MethodReasoning {
	var breakdown []MethodReasoning

	for _, method := range methodResults {
		reasoning := MethodReasoning{
			MethodName:     method.MethodName,
			MethodType:     method.MethodType,
			Success:        method.Success,
			Confidence:     method.Confidence,
			ProcessingTime: method.ProcessingTime,
		}

		if method.Success {
			reasoning.Reasoning = re.generateMethodReasoning(method)
			reasoning.Evidence = re.generateEvidence([]shared.ClassificationMethodResult{method}, method.Result)
		} else {
			reasoning.Reasoning = fmt.Sprintf("Method failed: %s", method.Error)
			reasoning.Error = method.Error
		}

		breakdown = append(breakdown, reasoning)
	}

	return breakdown
}

// generateMethodReasoning generates reasoning for a specific method
func (re *ReasoningEngine) generateMethodReasoning(method shared.ClassificationMethodResult) string {
	switch method.MethodType {
	case "keyword":
		return fmt.Sprintf("Keyword-based classification identified '%s' by matching %d keywords with %.1f%% confidence. Processing time: %v",
			method.Result.IndustryName, len(method.Keywords), method.Confidence*100, method.ProcessingTime)
	case "ml":
		return fmt.Sprintf("Machine learning classification predicted '%s' with %.1f%% confidence using advanced pattern recognition. Processing time: %v",
			method.Result.IndustryName, method.Confidence*100, method.ProcessingTime)
	case "description":
		return fmt.Sprintf("Description-based classification identified '%s' through content analysis with %.1f%% confidence. Processing time: %v",
			method.Result.IndustryName, method.Confidence*100, method.ProcessingTime)
	default:
		return fmt.Sprintf("Classification method identified '%s' with %.1f%% confidence. Processing time: %v",
			method.Result.IndustryName, method.Confidence*100, method.ProcessingTime)
	}
}

// analyzeConfidence analyzes confidence scores across methods
func (re *ReasoningEngine) analyzeConfidence(
	methodResults []shared.ClassificationMethodResult,
	primaryClassification *shared.IndustryClassification,
) *ConfidenceAnalysis {
	var confidences []float64
	methodConfidences := make(map[string]float64)

	// Collect confidence scores
	for _, method := range methodResults {
		if method.Success {
			confidences = append(confidences, method.Confidence)
			methodConfidences[method.MethodType] = method.Confidence
		}
	}

	// Calculate confidence range
	var min, max, sum float64
	if len(confidences) > 0 {
		min = confidences[0]
		max = confidences[0]
		for _, conf := range confidences {
			if conf < min {
				min = conf
			}
			if conf > max {
				max = conf
			}
			sum += conf
		}
	}

	median := re.calculateMedian(confidences)
	mean := sum / float64(len(confidences))

	// Generate confidence factors
	confidenceFactors := re.generateConfidenceFactors(methodResults, confidences)

	// Determine uncertainty level
	uncertaintyLevel := re.determineUncertaintyLevel(confidences, primaryClassification.ConfidenceScore)

	// Calculate reliability score
	reliabilityScore := re.calculateReliabilityScore(confidences, methodResults)

	return &ConfidenceAnalysis{
		OverallConfidence: primaryClassification.ConfidenceScore,
		ConfidenceRange: ConfidenceRange{
			Min:    min,
			Max:    max,
			Median: median,
			Mean:   mean,
		},
		MethodConfidences: methodConfidences,
		ConfidenceFactors: confidenceFactors,
		UncertaintyLevel:  uncertaintyLevel,
		ReliabilityScore:  reliabilityScore,
	}
}

// assessQuality assesses the quality of the classification
func (re *ReasoningEngine) assessQuality(
	qualityMetrics *shared.ClassificationQuality,
	methodResults []shared.ClassificationMethodResult,
) *QualityAssessment {
	if qualityMetrics == nil {
		return &QualityAssessment{
			OverallQuality:   0.0,
			QualityGrade:     "F",
			QualityFactors:   []QualityFactor{},
			DataCompleteness: 0.0,
			MethodAgreement:  0.0,
			EvidenceStrength: 0.0,
			Recommendations:  []string{"Quality metrics not available"},
		}
	}

	// Generate quality factors
	qualityFactors := re.generateQualityFactors(qualityMetrics)

	// Determine quality grade
	qualityGrade := re.determineQualityGrade(qualityMetrics.OverallQuality)

	// Generate quality recommendations
	recommendations := re.generateQualityRecommendations(qualityMetrics, methodResults)

	return &QualityAssessment{
		OverallQuality:   qualityMetrics.OverallQuality,
		QualityGrade:     qualityGrade,
		QualityFactors:   qualityFactors,
		DataCompleteness: qualityMetrics.DataCompleteness,
		MethodAgreement:  qualityMetrics.MethodAgreement,
		EvidenceStrength: qualityMetrics.EvidenceStrength,
		Recommendations:  recommendations,
	}
}

// generateSummary generates a brief summary of the classification
func (re *ReasoningEngine) generateSummary(
	businessName string,
	primaryClassification *shared.IndustryClassification,
	methodResults []shared.ClassificationMethodResult,
) string {
	successfulMethods := 0
	for _, method := range methodResults {
		if method.Success {
			successfulMethods++
		}
	}

	return fmt.Sprintf("Business '%s' was classified as '%s' with %.1f%% confidence using %d successful classification methods.",
		businessName, primaryClassification.IndustryName, primaryClassification.ConfidenceScore*100, successfulMethods)
}

// generateDetailedReasoning generates detailed reasoning for the classification
func (re *ReasoningEngine) generateDetailedReasoning(
	businessName string,
	primaryClassification *shared.IndustryClassification,
	methodResults []shared.ClassificationMethodResult,
	evidence []EvidenceItem,
	confidenceAnalysis *ConfidenceAnalysis,
	qualityAssessment *QualityAssessment,
) string {
	var reasoning strings.Builder

	// Introduction
	reasoning.WriteString(fmt.Sprintf("Detailed Classification Analysis for '%s'\n\n", businessName))

	// Primary classification
	reasoning.WriteString(fmt.Sprintf("Primary Classification: %s (%.1f%% confidence)\n",
		primaryClassification.IndustryName, primaryClassification.ConfidenceScore*100))

	// Method summary
	successfulMethods := 0
	for _, method := range methodResults {
		if method.Success {
			successfulMethods++
		}
	}
	reasoning.WriteString(fmt.Sprintf("Classification Methods: %d successful out of %d total methods\n\n",
		successfulMethods, len(methodResults)))

	// Evidence summary
	reasoning.WriteString(fmt.Sprintf("Evidence: %d pieces of supporting evidence identified\n", len(evidence)))
	if len(evidence) > 0 {
		reasoning.WriteString("Key evidence includes:\n")
		for i, ev := range evidence {
			if i >= 3 { // Show only top 3 evidence items
				break
			}
			reasoning.WriteString(fmt.Sprintf("- %s (strength: %.1f%%)\n", ev.Description, ev.Strength*100))
		}
	}
	reasoning.WriteString("\n")

	// Confidence analysis
	reasoning.WriteString(fmt.Sprintf("Confidence Analysis:\n"))
	reasoning.WriteString(fmt.Sprintf("- Overall confidence: %.1f%%\n", confidenceAnalysis.OverallConfidence*100))
	reasoning.WriteString(fmt.Sprintf("- Confidence range: %.1f%% - %.1f%%\n",
		confidenceAnalysis.ConfidenceRange.Min*100, confidenceAnalysis.ConfidenceRange.Max*100))
	reasoning.WriteString(fmt.Sprintf("- Uncertainty level: %s\n", confidenceAnalysis.UncertaintyLevel))
	reasoning.WriteString(fmt.Sprintf("- Reliability score: %.1f%%\n\n", confidenceAnalysis.ReliabilityScore*100))

	// Quality assessment
	reasoning.WriteString(fmt.Sprintf("Quality Assessment:\n"))
	reasoning.WriteString(fmt.Sprintf("- Overall quality: %.1f%% (Grade: %s)\n",
		qualityAssessment.OverallQuality*100, qualityAssessment.QualityGrade))
	reasoning.WriteString(fmt.Sprintf("- Method agreement: %.1f%%\n", qualityAssessment.MethodAgreement*100))
	reasoning.WriteString(fmt.Sprintf("- Evidence strength: %.1f%%\n", qualityAssessment.EvidenceStrength*100))
	reasoning.WriteString(fmt.Sprintf("- Data completeness: %.1f%%\n", qualityAssessment.DataCompleteness*100))

	return reasoning.String()
}

// generateRecommendations generates recommendations based on the classification results
func (re *ReasoningEngine) generateRecommendations(
	methodResults []shared.ClassificationMethodResult,
	confidenceAnalysis *ConfidenceAnalysis,
	qualityAssessment *QualityAssessment,
) []string {
	var recommendations []string

	// Confidence-based recommendations
	if confidenceAnalysis.OverallConfidence < 0.7 {
		recommendations = append(recommendations, "Consider gathering additional business information to improve classification confidence")
	}

	if confidenceAnalysis.UncertaintyLevel == "high" {
		recommendations = append(recommendations, "High uncertainty detected - manual review may be beneficial")
	}

	// Quality-based recommendations
	if qualityAssessment.MethodAgreement < 0.6 {
		recommendations = append(recommendations, "Low method agreement - consider using additional classification methods")
	}

	if qualityAssessment.EvidenceStrength < 0.5 {
		recommendations = append(recommendations, "Weak evidence strength - more detailed business information would improve classification")
	}

	// Method-specific recommendations
	failedMethods := 0
	for _, method := range methodResults {
		if !method.Success {
			failedMethods++
		}
	}

	if failedMethods > 0 {
		recommendations = append(recommendations, fmt.Sprintf("%d classification methods failed - investigate and improve method reliability", failedMethods))
	}

	// Default recommendation if no specific issues
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Classification quality is good - no specific recommendations")
	}

	return recommendations
}

// Helper methods

func (re *ReasoningEngine) calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2.0
	}
	return sorted[n/2]
}

func (re *ReasoningEngine) generateConfidenceFactors(
	methodResults []shared.ClassificationMethodResult,
	confidences []float64,
) []ConfidenceFactor {
	var factors []ConfidenceFactor

	// Method agreement factor
	agreementFactor := re.calculateMethodAgreement(methodResults)
	factors = append(factors, ConfidenceFactor{
		Factor:      "method_agreement",
		Impact:      agreementFactor - 0.5, // Center around 0
		Description: fmt.Sprintf("Agreement between classification methods: %.1f%%", agreementFactor*100),
		Weight:      0.3,
	})

	// Evidence strength factor
	evidenceStrength := re.calculateEvidenceStrength(methodResults)
	factors = append(factors, ConfidenceFactor{
		Factor:      "evidence_strength",
		Impact:      evidenceStrength - 0.5,
		Description: fmt.Sprintf("Strength of supporting evidence: %.1f%%", evidenceStrength*100),
		Weight:      0.2,
	})

	// Data completeness factor
	dataCompleteness := re.calculateDataCompleteness(methodResults)
	factors = append(factors, ConfidenceFactor{
		Factor:      "data_completeness",
		Impact:      dataCompleteness - 0.5,
		Description: fmt.Sprintf("Completeness of input data: %.1f%%", dataCompleteness*100),
		Weight:      0.2,
	})

	return factors
}

func (re *ReasoningEngine) determineUncertaintyLevel(confidences []float64, overallConfidence float64) string {
	if overallConfidence >= 0.8 {
		return "low"
	} else if overallConfidence >= 0.6 {
		return "medium"
	} else {
		return "high"
	}
}

func (re *ReasoningEngine) calculateReliabilityScore(confidences []float64, methodResults []shared.ClassificationMethodResult) float64 {
	if len(confidences) == 0 {
		return 0.0
	}

	// Base reliability on confidence consistency and method success rate
	successRate := float64(len(confidences)) / float64(len(methodResults))

	// Calculate confidence variance
	var sum, mean float64
	for _, conf := range confidences {
		sum += conf
	}
	mean = sum / float64(len(confidences))

	var variance float64
	for _, conf := range confidences {
		variance += (conf - mean) * (conf - mean)
	}
	variance /= float64(len(confidences))

	// Lower variance = higher reliability
	consistency := 1.0 - variance

	// Combine success rate and consistency
	reliability := (successRate * 0.6) + (consistency * 0.4)

	if reliability > 1.0 {
		reliability = 1.0
	}

	return reliability
}

func (re *ReasoningEngine) generateQualityFactors(qualityMetrics *shared.ClassificationQuality) []QualityFactor {
	var factors []QualityFactor

	factors = append(factors, QualityFactor{
		Factor:      "method_agreement",
		Score:       qualityMetrics.MethodAgreement,
		Weight:      0.4,
		Description: "Agreement between different classification methods",
		Impact:      re.getImpactString(qualityMetrics.MethodAgreement),
	})

	factors = append(factors, QualityFactor{
		Factor:      "evidence_strength",
		Score:       qualityMetrics.EvidenceStrength,
		Weight:      0.3,
		Description: "Strength of evidence supporting the classification",
		Impact:      re.getImpactString(qualityMetrics.EvidenceStrength),
	})

	factors = append(factors, QualityFactor{
		Factor:      "data_completeness",
		Score:       qualityMetrics.DataCompleteness,
		Weight:      0.2,
		Description: "Completeness of input data used for classification",
		Impact:      re.getImpactString(qualityMetrics.DataCompleteness),
	})

	factors = append(factors, QualityFactor{
		Factor:      "confidence_consistency",
		Score:       1.0 - qualityMetrics.ConfidenceVariance,
		Weight:      0.1,
		Description: "Consistency of confidence scores across methods",
		Impact:      re.getImpactString(1.0 - qualityMetrics.ConfidenceVariance),
	})

	return factors
}

func (re *ReasoningEngine) determineQualityGrade(overallQuality float64) string {
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

func (re *ReasoningEngine) generateQualityRecommendations(
	qualityMetrics *shared.ClassificationQuality,
	methodResults []shared.ClassificationMethodResult,
) []string {
	var recommendations []string

	if qualityMetrics.MethodAgreement < 0.6 {
		recommendations = append(recommendations, "Improve method agreement by enhancing classification algorithms")
	}

	if qualityMetrics.EvidenceStrength < 0.5 {
		recommendations = append(recommendations, "Strengthen evidence collection and analysis")
	}

	if qualityMetrics.DataCompleteness < 0.7 {
		recommendations = append(recommendations, "Collect more comprehensive business data")
	}

	return recommendations
}

func (re *ReasoningEngine) calculateMethodAgreement(methodResults []shared.ClassificationMethodResult) float64 {
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

func (re *ReasoningEngine) calculateEvidenceStrength(methodResults []shared.ClassificationMethodResult) float64 {
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

func (re *ReasoningEngine) calculateDataCompleteness(methodResults []shared.ClassificationMethodResult) float64 {
	successfulMethods := 0
	for _, method := range methodResults {
		if method.Success {
			successfulMethods++
		}
	}

	return float64(successfulMethods) / float64(len(methodResults))
}

func (re *ReasoningEngine) getImpactString(score float64) string {
	if score >= 0.8 {
		return "positive"
	} else if score >= 0.6 {
		return "neutral"
	} else {
		return "negative"
	}
}

func (re *ReasoningEngine) generateRequestID() string {
	return fmt.Sprintf("reasoning_%d", time.Now().UnixNano())
}
