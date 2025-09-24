package data_discovery

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// QualityScorer provides comprehensive quality and relevance scoring for discovered data points
type QualityScorer struct {
	config *DataDiscoveryConfig
	logger *zap.Logger
}

// NewQualityScorer creates a new quality scorer
func NewQualityScorer(config *DataDiscoveryConfig, logger *zap.Logger) *QualityScorer {
	return &QualityScorer{
		config: config,
		logger: logger,
	}
}

// QualityScore represents comprehensive quality metrics for a data point
type QualityScore struct {
	OverallScore      float64                 `json:"overall_score"`      // 0.0-1.0 overall quality
	RelevanceScore    float64                 `json:"relevance_score"`    // 0.0-1.0 business relevance
	AccuracyScore     float64                 `json:"accuracy_score"`     // 0.0-1.0 data accuracy
	CompletenessScore float64                 `json:"completeness_score"` // 0.0-1.0 field completeness
	FreshnessScore    float64                 `json:"freshness_score"`    // 0.0-1.0 data freshness
	CredibilityScore  float64                 `json:"credibility_score"`  // 0.0-1.0 source credibility
	ConsistencyScore  float64                 `json:"consistency_score"`  // 0.0-1.0 cross-field consistency
	QualityIndicators map[string]interface{}  `json:"quality_indicators"` // Detailed quality metrics
	ScoringComponents []ScoringComponent      `json:"scoring_components"` // Individual component scores
	Recommendations   []QualityRecommendation `json:"recommendations"`    // Improvement recommendations
	LastUpdated       time.Time               `json:"last_updated"`
}

// ScoringComponent represents an individual scoring component
type ScoringComponent struct {
	ComponentName string   `json:"component_name"`
	Score         float64  `json:"score"`
	Weight        float64  `json:"weight"`
	Description   string   `json:"description"`
	Evidence      []string `json:"evidence"`
	ImpactLevel   string   `json:"impact_level"` // "high", "medium", "low"
}

// QualityRecommendation represents a recommendation for improving data quality
type QualityRecommendation struct {
	RecommendationID   string   `json:"recommendation_id"`
	Type               string   `json:"type"`     // "accuracy", "completeness", "freshness", etc.
	Priority           string   `json:"priority"` // "high", "medium", "low"
	Description        string   `json:"description"`
	ExpectedImpact     float64  `json:"expected_impact"`     // Expected score improvement
	ImplementationCost string   `json:"implementation_cost"` // "low", "medium", "high"
	Actions            []string `json:"actions"`
}

// BusinessContext represents context for relevance scoring
type BusinessContext struct {
	Industry       string             `json:"industry"`
	BusinessType   string             `json:"business_type"` // B2B, B2C, etc.
	Geography      string             `json:"geography"`
	CompanySize    string             `json:"company_size"`     // startup, small, medium, large
	UseCaseProfile string             `json:"use_case_profile"` // verification, analysis, etc.
	PriorityFields []string           `json:"priority_fields"`
	CustomWeights  map[string]float64 `json:"custom_weights"`
}

// ScoreDiscoveredFields calculates quality scores for all discovered fields
func (qs *QualityScorer) ScoreDiscoveredFields(ctx context.Context, fields []DiscoveredField, patterns []PatternMatch, classification *ClassificationResult, businessContext *BusinessContext) ([]FieldQualityAssessment, error) {
	qs.logger.Debug("Starting quality scoring for discovered fields",
		zap.Int("field_count", len(fields)),
		zap.Int("pattern_count", len(patterns)))

	var assessments []FieldQualityAssessment

	for _, field := range fields {
		select {
		case <-ctx.Done():
			return assessments, ctx.Err()
		default:
			assessment, err := qs.ScoreField(ctx, field, patterns, classification, businessContext)
			if err != nil {
				qs.logger.Warn("Failed to score field",
					zap.String("field_name", field.FieldName),
					zap.Error(err))
				continue
			}
			assessments = append(assessments, *assessment)
		}
	}

	qs.logger.Info("Quality scoring completed",
		zap.Int("assessments_generated", len(assessments)))

	return assessments, nil
}

// FieldQualityAssessment represents a comprehensive quality assessment for a field
type FieldQualityAssessment struct {
	FieldName       string           `json:"field_name"`
	FieldType       string           `json:"field_type"`
	QualityScore    QualityScore     `json:"quality_score"`
	ConfidenceScore float64          `json:"confidence_score"`
	BusinessImpact  string           `json:"business_impact"`  // "critical", "high", "medium", "low"
	QualityCategory string           `json:"quality_category"` // "excellent", "good", "fair", "poor"
	RiskFactors     []RiskFactor     `json:"risk_factors"`
	ValueMetrics    ValueMetrics     `json:"value_metrics"`
	ImprovementPlan *ImprovementPlan `json:"improvement_plan,omitempty"`
}

// RiskFactor represents a quality risk factor
type RiskFactor struct {
	RiskType    string  `json:"risk_type"`
	Severity    string  `json:"severity"` // "critical", "high", "medium", "low"
	Description string  `json:"description"`
	Probability float64 `json:"probability"` // 0.0-1.0
	Impact      string  `json:"impact"`
	Mitigation  string  `json:"mitigation"`
}

// ValueMetrics represents business value metrics for a field
type ValueMetrics struct {
	BusinessValue       float64 `json:"business_value"`       // 0.0-1.0
	OperationalImpact   float64 `json:"operational_impact"`   // 0.0-1.0
	ComplianceRelevance float64 `json:"compliance_relevance"` // 0.0-1.0
	CustomerImpact      float64 `json:"customer_impact"`      // 0.0-1.0
	RevenueContribution float64 `json:"revenue_contribution"` // 0.0-1.0
	CostReduction       float64 `json:"cost_reduction"`       // 0.0-1.0
}

// ImprovementPlan represents a plan to improve field quality
type ImprovementPlan struct {
	PlanID              string                 `json:"plan_id"`
	ExpectedImprovement float64                `json:"expected_improvement"`
	TimeToImplement     time.Duration          `json:"time_to_implement"`
	ResourceRequirement string                 `json:"resource_requirement"`
	Actions             []ImprovementAction    `json:"actions"`
	Milestones          []ImprovementMilestone `json:"milestones"`
}

// ImprovementAction represents a specific improvement action
type ImprovementAction struct {
	ActionID       string   `json:"action_id"`
	Description    string   `json:"description"`
	Priority       string   `json:"priority"`
	ExpectedImpact float64  `json:"expected_impact"`
	Effort         string   `json:"effort"`
	Dependencies   []string `json:"dependencies"`
}

// ImprovementMilestone represents a milestone in the improvement plan
type ImprovementMilestone struct {
	MilestoneID     string    `json:"milestone_id"`
	Description     string    `json:"description"`
	TargetDate      time.Time `json:"target_date"`
	SuccessCriteria []string  `json:"success_criteria"`
	TargetScore     float64   `json:"target_score"`
}

// ScoreField calculates comprehensive quality score for a single field
func (qs *QualityScorer) ScoreField(ctx context.Context, field DiscoveredField, patterns []PatternMatch, classification *ClassificationResult, businessContext *BusinessContext) (*FieldQualityAssessment, error) {
	qs.logger.Debug("Scoring field",
		zap.String("field_name", field.FieldName),
		zap.String("field_type", field.FieldType))

	// Calculate individual quality components
	relevanceScore := qs.calculateRelevanceScore(field, classification, businessContext)
	accuracyScore := qs.calculateAccuracyScore(field, patterns)
	completenessScore := qs.calculateCompletenessScore(field)
	freshnessScore := qs.calculateFreshnessScore(field)
	credibilityScore := qs.calculateCredibilityScore(field, patterns)
	consistencyScore := qs.calculateConsistencyScore(field, patterns)

	// Calculate overall quality score
	overallScore := qs.calculateOverallQualityScore(
		relevanceScore, accuracyScore, completenessScore,
		freshnessScore, credibilityScore, consistencyScore)

	// Build quality score structure
	qualityScore := QualityScore{
		OverallScore:      overallScore,
		RelevanceScore:    relevanceScore,
		AccuracyScore:     accuracyScore,
		CompletenessScore: completenessScore,
		FreshnessScore:    freshnessScore,
		CredibilityScore:  credibilityScore,
		ConsistencyScore:  consistencyScore,
		QualityIndicators: qs.buildQualityIndicators(field, patterns),
		ScoringComponents: qs.buildScoringComponents(relevanceScore, accuracyScore, completenessScore, freshnessScore, credibilityScore, consistencyScore),
		Recommendations:   qs.generateRecommendations(field, overallScore),
		LastUpdated:       time.Now(),
	}

	// Determine business impact and quality category
	businessImpact := qs.determineBusinessImpact(field, qualityScore, businessContext)
	qualityCategory := qs.determineQualityCategory(overallScore)

	// Identify risk factors
	riskFactors := qs.identifyRiskFactors(field, qualityScore)

	// Calculate value metrics
	valueMetrics := qs.calculateValueMetrics(field, businessContext)

	// Generate improvement plan if needed
	var improvementPlan *ImprovementPlan
	if overallScore < 0.8 {
		improvementPlan = qs.generateImprovementPlan(field, qualityScore)
	}

	assessment := &FieldQualityAssessment{
		FieldName:       field.FieldName,
		FieldType:       field.FieldType,
		QualityScore:    qualityScore,
		ConfidenceScore: field.ConfidenceScore,
		BusinessImpact:  businessImpact,
		QualityCategory: qualityCategory,
		RiskFactors:     riskFactors,
		ValueMetrics:    valueMetrics,
		ImprovementPlan: improvementPlan,
	}

	return assessment, nil
}

// calculateRelevanceScore calculates business relevance score
func (qs *QualityScorer) calculateRelevanceScore(field DiscoveredField, classification *ClassificationResult, businessContext *BusinessContext) float64 {
	score := 0.5 // Base relevance score

	// Industry-specific relevance
	if businessContext != nil {
		score += qs.getIndustryRelevance(field.FieldType, businessContext.Industry) * 0.3

		// Business type relevance
		score += qs.getBusinessTypeRelevance(field.FieldType, businessContext.BusinessType) * 0.2

		// Use case profile relevance
		score += qs.getUseCaseRelevance(field.FieldType, businessContext.UseCaseProfile) * 0.2

		// Priority field bonus
		for _, priorityField := range businessContext.PriorityFields {
			if strings.EqualFold(field.FieldType, priorityField) {
				score += 0.2
				break
			}
		}

		// Custom weight application
		if weight, exists := businessContext.CustomWeights[field.FieldType]; exists {
			score = score * weight
		}
	}

	// Classification-based relevance
	if classification != nil {
		score += qs.getClassificationRelevance(field.FieldType, classification) * 0.1
	}

	// Field priority impact
	priorityBonus := (10.0 - float64(field.Priority)) / 10.0 * 0.2
	score += priorityBonus

	// Ensure score is within bounds
	if score > 1.0 {
		score = 1.0
	} else if score < 0.0 {
		score = 0.0
	}

	return score
}

// calculateAccuracyScore calculates data accuracy score
func (qs *QualityScorer) calculateAccuracyScore(field DiscoveredField, patterns []PatternMatch) float64 {
	score := field.ConfidenceScore // Start with field confidence

	// Pattern match validation
	for _, pattern := range patterns {
		if pattern.FieldType == field.FieldType {
			score += pattern.ConfidenceScore * 0.1
		}
	}

	// Sample value validation
	if len(field.SampleValues) > 0 {
		validSamples := 0
		for _, sample := range field.SampleValues {
			if qs.validateSampleValue(sample, field.FieldType) {
				validSamples++
			}
		}
		sampleAccuracy := float64(validSamples) / float64(len(field.SampleValues))
		score = (score + sampleAccuracy) / 2.0
	}

	// Validation rules compliance
	if len(field.ValidationRules) > 0 {
		score += 0.1 // Bonus for having validation rules
	}

	// Ensure score is within bounds
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// calculateCompletenessScore calculates field completeness score
func (qs *QualityScorer) calculateCompletenessScore(field DiscoveredField) float64 {
	score := 0.5 // Base completeness

	// Sample values availability
	if len(field.SampleValues) > 0 {
		score += 0.3

		// More samples = better completeness
		sampleBonus := math.Min(float64(len(field.SampleValues))/10.0, 0.2)
		score += sampleBonus
	}

	// Validation rules presence
	if len(field.ValidationRules) > 0 {
		score += 0.2
	}

	// Metadata richness
	if field.Metadata != nil && len(field.Metadata) > 0 {
		score += 0.1
	}

	// Data type specification
	if field.DataType != "" {
		score += 0.1
	}

	// Extraction method specification
	if field.ExtractionMethod != "" {
		score += 0.1
	}

	return math.Min(score, 1.0)
}

// calculateFreshnessScore calculates data freshness score
func (qs *QualityScorer) calculateFreshnessScore(field DiscoveredField) float64 {
	// Check for timestamp metadata
	if field.Metadata != nil {
		if extractedAt, exists := field.Metadata["extracted_at"]; exists {
			if timestamp, ok := extractedAt.(time.Time); ok {
				age := time.Since(timestamp)

				// Freshness decreases over time
				if age < time.Hour {
					return 1.0
				} else if age < 24*time.Hour {
					return 0.8
				} else if age < 7*24*time.Hour {
					return 0.6
				} else if age < 30*24*time.Hour {
					return 0.4
				} else {
					return 0.2
				}
			}
		}
	}

	// Default freshness for newly discovered fields
	return 0.9
}

// calculateCredibilityScore calculates source credibility score
func (qs *QualityScorer) calculateCredibilityScore(field DiscoveredField, patterns []PatternMatch) float64 {
	score := 0.5 // Base credibility

	// Pattern source credibility
	for _, pattern := range patterns {
		if pattern.FieldType == field.FieldType {
			// Well-structured patterns are more credible
			if pattern.ConfidenceScore > 0.8 {
				score += 0.2
			}

			// Context validation
			if pattern.Context != "" && len(pattern.Context) > 10 {
				score += 0.1
			}
		}
	}

	// Field metadata indicators
	if len(field.Metadata) > 0 {
		// Source URL credibility
		if sourceURL, exists := field.Metadata["source_url"]; exists {
			if url, ok := sourceURL.(string); ok {
				score += qs.evaluateURLCredibility(url) * 0.3
			}
		}
	}

	// Extraction method credibility
	switch field.ExtractionMethod {
	case "structured_data":
		score += 0.2
	case "xpath", "css_selector":
		score += 0.15
	case "regex":
		score += 0.1
	case "ml_classification":
		score += 0.05
	}

	return math.Min(score, 1.0)
}

// calculateConsistencyScore calculates cross-field consistency score
func (qs *QualityScorer) calculateConsistencyScore(field DiscoveredField, patterns []PatternMatch) float64 {
	score := 0.7 // Base consistency

	// Pattern consistency
	fieldPatterns := 0
	consistentPatterns := 0

	for _, pattern := range patterns {
		if pattern.FieldType == field.FieldType {
			fieldPatterns++

			// Check pattern consistency
			if pattern.ConfidenceScore > 0.7 {
				consistentPatterns++
			}
		}
	}

	if fieldPatterns > 0 {
		patternConsistency := float64(consistentPatterns) / float64(fieldPatterns)
		score = (score + patternConsistency) / 2.0
	}

	// Sample value consistency
	if len(field.SampleValues) > 1 {
		score += qs.evaluateSampleConsistency(field.SampleValues, field.FieldType) * 0.3
	}

	return math.Min(score, 1.0)
}

// calculateOverallQualityScore calculates the weighted overall quality score
func (qs *QualityScorer) calculateOverallQualityScore(relevance, accuracy, completeness, freshness, credibility, consistency float64) float64 {
	// Define weights for each component
	weights := map[string]float64{
		"relevance":    0.25,
		"accuracy":     0.25,
		"completeness": 0.20,
		"freshness":    0.10,
		"credibility":  0.15,
		"consistency":  0.05,
	}

	overallScore := relevance*weights["relevance"] +
		accuracy*weights["accuracy"] +
		completeness*weights["completeness"] +
		freshness*weights["freshness"] +
		credibility*weights["credibility"] +
		consistency*weights["consistency"]

	return math.Min(overallScore, 1.0)
}

// Helper methods for relevance scoring
func (qs *QualityScorer) getIndustryRelevance(fieldType, industry string) float64 {
	industryRelevance := map[string]map[string]float64{
		"technology": {
			"email":        0.9,
			"phone":        0.8,
			"address":      0.7,
			"url":          0.9,
			"social_media": 0.8,
			"tax_id":       0.6,
		},
		"finance": {
			"email":   0.9,
			"phone":   0.9,
			"address": 0.8,
			"tax_id":  0.9,
			"url":     0.7,
		},
		"retail": {
			"email":        0.8,
			"phone":        0.9,
			"address":      0.9,
			"social_media": 0.8,
			"url":          0.8,
		},
	}

	if industryFields, exists := industryRelevance[strings.ToLower(industry)]; exists {
		if relevance, fieldExists := industryFields[fieldType]; fieldExists {
			return relevance
		}
	}

	return 0.5 // Default relevance
}

func (qs *QualityScorer) getBusinessTypeRelevance(fieldType, businessType string) float64 {
	businessTypeRelevance := map[string]map[string]float64{
		"b2b": {
			"email":   0.9,
			"phone":   0.9,
			"address": 0.8,
			"tax_id":  0.8,
		},
		"b2c": {
			"email":        0.8,
			"phone":        0.7,
			"address":      0.9,
			"social_media": 0.9,
		},
	}

	if typeFields, exists := businessTypeRelevance[strings.ToLower(businessType)]; exists {
		if relevance, fieldExists := typeFields[fieldType]; fieldExists {
			return relevance
		}
	}

	return 0.5
}

func (qs *QualityScorer) getUseCaseRelevance(fieldType, useCase string) float64 {
	useCaseRelevance := map[string]map[string]float64{
		"verification": {
			"email":   0.9,
			"phone":   0.9,
			"address": 0.9,
			"tax_id":  0.8,
		},
		"analysis": {
			"url":          0.8,
			"social_media": 0.8,
			"email":        0.7,
		},
	}

	if caseFields, exists := useCaseRelevance[strings.ToLower(useCase)]; exists {
		if relevance, fieldExists := caseFields[fieldType]; fieldExists {
			return relevance
		}
	}

	return 0.5
}

func (qs *QualityScorer) getClassificationRelevance(fieldType string, classification *ClassificationResult) float64 {
	// Classification confidence affects relevance
	return classification.ConfidenceScore * 0.5
}

// Helper methods for validation
func (qs *QualityScorer) validateSampleValue(value, fieldType string) bool {
	switch fieldType {
	case "email":
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		return emailRegex.MatchString(value)
	case "phone":
		// Remove all non-digit characters except +
		normalized := regexp.MustCompile(`[^\d+]`).ReplaceAllString(value, "")
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{6,14}$`)
		return phoneRegex.MatchString(normalized)
	case "url":
		return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
	case "address":
		return len(value) > 10 && (strings.Contains(value, ",") || strings.Contains(value, " "))
	default:
		return len(value) > 0
	}
}

func (qs *QualityScorer) evaluateURLCredibility(url string) float64 {
	score := 0.5

	// HTTPS bonus
	if strings.HasPrefix(url, "https://") {
		score += 0.3
	}

	// Domain credibility indicators
	if strings.Contains(url, ".gov") || strings.Contains(url, ".edu") {
		score += 0.4
	} else if strings.Contains(url, ".org") {
		score += 0.2
	} else if strings.Contains(url, ".com") {
		score += 0.1
	}

	return math.Min(score, 1.0)
}

func (qs *QualityScorer) evaluateSampleConsistency(samples []string, fieldType string) float64 {
	if len(samples) < 2 {
		return 1.0
	}

	consistentSamples := 0
	totalComparisons := 0

	for i := 0; i < len(samples); i++ {
		for j := i + 1; j < len(samples); j++ {
			totalComparisons++
			if qs.samplesAreConsistent(samples[i], samples[j], fieldType) {
				consistentSamples++
			}
		}
	}

	if totalComparisons == 0 {
		return 1.0
	}

	return float64(consistentSamples) / float64(totalComparisons)
}

func (qs *QualityScorer) samplesAreConsistent(sample1, sample2, fieldType string) bool {
	switch fieldType {
	case "email":
		domain1 := qs.extractEmailDomain(sample1)
		domain2 := qs.extractEmailDomain(sample2)
		return domain1 == domain2
	case "phone":
		return qs.normalizePhone(sample1) == qs.normalizePhone(sample2)
	case "address":
		return qs.addressesAreConsistent(sample1, sample2)
	default:
		return sample1 == sample2
	}
}

func (qs *QualityScorer) extractEmailDomain(email string) string {
	if idx := strings.LastIndex(email, "@"); idx != -1 {
		return email[idx+1:]
	}
	return ""
}

func (qs *QualityScorer) normalizePhone(phone string) string {
	// Remove all non-numeric characters except +
	normalized := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")
	return normalized
}

func (qs *QualityScorer) addressesAreConsistent(addr1, addr2 string) bool {
	// Basic consistency check - same city/state
	parts1 := strings.Split(addr1, ",")
	parts2 := strings.Split(addr2, ",")

	if len(parts1) >= 2 && len(parts2) >= 2 {
		// Compare last two parts (typically city, state)
		return strings.TrimSpace(parts1[len(parts1)-2]) == strings.TrimSpace(parts2[len(parts2)-2])
	}

	return false
}

// Builder methods for quality score components
func (qs *QualityScorer) buildQualityIndicators(field DiscoveredField, patterns []PatternMatch) map[string]interface{} {
	indicators := make(map[string]interface{})

	indicators["sample_count"] = len(field.SampleValues)
	indicators["validation_rules_count"] = len(field.ValidationRules)
	indicators["pattern_matches"] = len(patterns)
	indicators["field_priority"] = field.Priority
	indicators["business_value"] = field.BusinessValue
	indicators["extraction_method"] = field.ExtractionMethod

	return indicators
}

func (qs *QualityScorer) buildScoringComponents(relevance, accuracy, completeness, freshness, credibility, consistency float64) []ScoringComponent {
	return []ScoringComponent{
		{
			ComponentName: "Relevance",
			Score:         relevance,
			Weight:        0.25,
			Description:   "Business relevance and importance",
			ImpactLevel:   qs.getImpactLevel(relevance),
		},
		{
			ComponentName: "Accuracy",
			Score:         accuracy,
			Weight:        0.25,
			Description:   "Data accuracy and validation",
			ImpactLevel:   qs.getImpactLevel(accuracy),
		},
		{
			ComponentName: "Completeness",
			Score:         completeness,
			Weight:        0.20,
			Description:   "Field completeness and richness",
			ImpactLevel:   qs.getImpactLevel(completeness),
		},
		{
			ComponentName: "Freshness",
			Score:         freshness,
			Weight:        0.10,
			Description:   "Data recency and timeliness",
			ImpactLevel:   qs.getImpactLevel(freshness),
		},
		{
			ComponentName: "Credibility",
			Score:         credibility,
			Weight:        0.15,
			Description:   "Source credibility and trustworthiness",
			ImpactLevel:   qs.getImpactLevel(credibility),
		},
		{
			ComponentName: "Consistency",
			Score:         consistency,
			Weight:        0.05,
			Description:   "Cross-field consistency",
			ImpactLevel:   qs.getImpactLevel(consistency),
		},
	}
}

func (qs *QualityScorer) getImpactLevel(score float64) string {
	if score >= 0.8 {
		return "high"
	} else if score >= 0.6 {
		return "medium"
	} else {
		return "low"
	}
}

func (qs *QualityScorer) generateRecommendations(field DiscoveredField, overallScore float64) []QualityRecommendation {
	var recommendations []QualityRecommendation

	if overallScore < 0.8 {
		if len(field.SampleValues) == 0 {
			recommendations = append(recommendations, QualityRecommendation{
				RecommendationID:   "add_samples",
				Type:               "completeness",
				Priority:           "high",
				Description:        "Add sample values to improve field understanding",
				ExpectedImpact:     0.2,
				ImplementationCost: "low",
				Actions:            []string{"Extract sample values", "Validate samples"},
			})
		}

		if len(field.ValidationRules) == 0 {
			recommendations = append(recommendations, QualityRecommendation{
				RecommendationID:   "add_validation",
				Type:               "accuracy",
				Priority:           "medium",
				Description:        "Add validation rules to improve accuracy",
				ExpectedImpact:     0.15,
				ImplementationCost: "medium",
				Actions:            []string{"Define validation rules", "Implement validators"},
			})
		}
	}

	return recommendations
}

func (qs *QualityScorer) determineBusinessImpact(field DiscoveredField, score QualityScore, businessContext *BusinessContext) string {
	impact := field.BusinessValue * score.OverallScore * score.RelevanceScore

	if impact >= 0.8 {
		return "critical"
	} else if impact >= 0.6 {
		return "high"
	} else if impact >= 0.4 {
		return "medium"
	} else {
		return "low"
	}
}

func (qs *QualityScorer) determineQualityCategory(overallScore float64) string {
	if overallScore >= 0.9 {
		return "excellent"
	} else if overallScore >= 0.7 {
		return "good"
	} else if overallScore >= 0.5 {
		return "fair"
	} else {
		return "poor"
	}
}

func (qs *QualityScorer) identifyRiskFactors(field DiscoveredField, score QualityScore) []RiskFactor {
	var risks []RiskFactor

	if score.AccuracyScore < 0.6 {
		risks = append(risks, RiskFactor{
			RiskType:    "accuracy",
			Severity:    "high",
			Description: "Low accuracy score may lead to incorrect business decisions",
			Probability: 1.0 - score.AccuracyScore,
			Impact:      "business_decisions",
			Mitigation:  "Improve validation rules and sample verification",
		})
	}

	if score.FreshnessScore < 0.5 {
		risks = append(risks, RiskFactor{
			RiskType:    "staleness",
			Severity:    "medium",
			Description: "Data may be outdated and unreliable",
			Probability: 1.0 - score.FreshnessScore,
			Impact:      "data_reliability",
			Mitigation:  "Implement regular data refresh processes",
		})
	}

	return risks
}

func (qs *QualityScorer) calculateValueMetrics(field DiscoveredField, businessContext *BusinessContext) ValueMetrics {
	baseValue := field.BusinessValue

	return ValueMetrics{
		BusinessValue:       baseValue,
		OperationalImpact:   baseValue * 0.8,
		ComplianceRelevance: qs.getComplianceRelevance(field.FieldType),
		CustomerImpact:      qs.getCustomerImpact(field.FieldType),
		RevenueContribution: baseValue * 0.6,
		CostReduction:       baseValue * 0.4,
	}
}

func (qs *QualityScorer) getComplianceRelevance(fieldType string) float64 {
	complianceFields := map[string]float64{
		"tax_id":  0.9,
		"address": 0.8,
		"email":   0.7,
		"phone":   0.7,
	}

	if relevance, exists := complianceFields[fieldType]; exists {
		return relevance
	}

	return 0.3
}

func (qs *QualityScorer) getCustomerImpact(fieldType string) float64 {
	customerFields := map[string]float64{
		"email":        0.9,
		"phone":        0.9,
		"address":      0.8,
		"social_media": 0.7,
		"url":          0.6,
	}

	if impact, exists := customerFields[fieldType]; exists {
		return impact
	}

	return 0.3
}

func (qs *QualityScorer) generateImprovementPlan(field DiscoveredField, score QualityScore) *ImprovementPlan {
	planID := fmt.Sprintf("improvement_%s_%d", field.FieldName, time.Now().Unix())

	var actions []ImprovementAction
	expectedImprovement := 0.0

	if score.CompletenessScore < 0.7 {
		actions = append(actions, ImprovementAction{
			ActionID:       "improve_completeness",
			Description:    "Add more sample values and metadata",
			Priority:       "high",
			ExpectedImpact: 0.2,
			Effort:         "low",
		})
		expectedImprovement += 0.2
	}

	if score.AccuracyScore < 0.8 {
		actions = append(actions, ImprovementAction{
			ActionID:       "improve_accuracy",
			Description:    "Enhance validation rules and pattern matching",
			Priority:       "high",
			ExpectedImpact: 0.25,
			Effort:         "medium",
		})
		expectedImprovement += 0.25
	}

	milestones := []ImprovementMilestone{
		{
			MilestoneID:     "milestone_1",
			Description:     "Complete accuracy improvements",
			TargetDate:      time.Now().Add(7 * 24 * time.Hour),
			SuccessCriteria: []string{"Accuracy score > 0.8"},
			TargetScore:     score.OverallScore + 0.15,
		},
		{
			MilestoneID:     "milestone_2",
			Description:     "Complete all improvements",
			TargetDate:      time.Now().Add(14 * 24 * time.Hour),
			SuccessCriteria: []string{"Overall score > 0.8"},
			TargetScore:     score.OverallScore + expectedImprovement,
		},
	}

	return &ImprovementPlan{
		PlanID:              planID,
		ExpectedImprovement: expectedImprovement,
		TimeToImplement:     14 * 24 * time.Hour,
		ResourceRequirement: "medium",
		Actions:             actions,
		Milestones:          milestones,
	}
}
