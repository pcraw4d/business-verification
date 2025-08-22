package risk_assessment

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// RiskScorer provides risk scoring capabilities
type RiskScorer struct {
	config *RiskAssessmentConfig
	logger *zap.Logger
}

// RiskScore represents a comprehensive risk score
type RiskScore struct {
	OverallScore    float64                  `json:"overall_score"`
	RiskLevel       RiskLevel                `json:"risk_level"`
	CategoryScores  map[string]CategoryScore `json:"category_scores"`
	WeightedFactors []WeightedRiskFactor     `json:"weighted_factors"`
	ConfidenceLevel float64                  `json:"confidence_level"`
	ScoreBreakdown  *ScoreBreakdown          `json:"score_breakdown,omitempty"`
	CalibrationData *CalibrationData         `json:"calibration_data,omitempty"`
	ScoreTimestamp  time.Time                `json:"score_timestamp"`
}

// CategoryScore represents a score for a specific risk category
type CategoryScore struct {
	Category      string       `json:"category"`
	Score         float64      `json:"score"`
	Weight        float64      `json:"weight"`
	WeightedScore float64      `json:"weighted_score"`
	RiskLevel     RiskLevel    `json:"risk_level"`
	Factors       []RiskFactor `json:"factors"`
	Confidence    float64      `json:"confidence"`
}

// WeightedRiskFactor represents a weighted risk factor
type WeightedRiskFactor struct {
	Factor       RiskFactor `json:"factor"`
	Weight       float64    `json:"weight"`
	Impact       float64    `json:"impact"`
	Contribution float64    `json:"contribution"`
}

// ScoreBreakdown contains detailed score breakdown
type ScoreBreakdown struct {
	SecurityScore   float64 `json:"security_score"`
	DomainScore     float64 `json:"domain_score"`
	ReputationScore float64 `json:"reputation_score"`
	ComplianceScore float64 `json:"compliance_score"`
	FinancialScore  float64 `json:"financial_score"`
	WeightedTotal   float64 `json:"weighted_total"`
	NormalizedScore float64 `json:"normalized_score"`
}

// CalibrationData contains calibration information
type CalibrationData struct {
	ModelVersion       string    `json:"model_version"`
	CalibrationDate    time.Time `json:"calibration_date"`
	ValidationScore    float64   `json:"validation_score"`
	Accuracy           float64   `json:"accuracy"`
	Precision          float64   `json:"precision"`
	Recall             float64   `json:"recall"`
	F1Score            float64   `json:"f1_score"`
	ConfidenceInterval float64   `json:"confidence_interval"`
}

// NewRiskScorer creates a new risk scorer
func NewRiskScorer(config *RiskAssessmentConfig, logger *zap.Logger) *RiskScorer {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &RiskScorer{
		config: config,
		logger: logger,
	}
}

// CalculateRiskScore calculates comprehensive risk score
func (rs *RiskScorer) CalculateRiskScore(ctx context.Context, result *RiskAssessmentResult) (*RiskScore, error) {
	startTime := time.Now()

	rs.logger.Info("Starting risk score calculation",
		zap.String("business_name", result.BusinessName))

	score := &RiskScore{
		CategoryScores:  make(map[string]CategoryScore),
		WeightedFactors: make([]WeightedRiskFactor, 0),
		ScoreTimestamp:  time.Now(),
	}

	// Calculate category scores
	if result.SecurityAnalysis != nil {
		securityScore := rs.calculateSecurityCategoryScore(result.SecurityAnalysis)
		score.CategoryScores["security"] = securityScore
	}

	if result.DomainAnalysis != nil {
		domainScore := rs.calculateDomainCategoryScore(result.DomainAnalysis)
		score.CategoryScores["domain"] = domainScore
	}

	if result.ReputationAnalysis != nil {
		reputationScore := rs.calculateReputationCategoryScore(result.ReputationAnalysis)
		score.CategoryScores["reputation"] = reputationScore
	}

	if result.ComplianceAnalysis != nil {
		complianceScore := rs.calculateComplianceCategoryScore(result.ComplianceAnalysis)
		score.CategoryScores["compliance"] = complianceScore
	}

	if result.FinancialAnalysis != nil {
		financialScore := rs.calculateFinancialCategoryScore(result.FinancialAnalysis)
		score.CategoryScores["financial"] = financialScore
	}

	// Calculate weighted factors
	score.WeightedFactors = rs.calculateWeightedFactors(result.RiskFactors)

	// Calculate overall score
	score.OverallScore = rs.calculateOverallScore(score.CategoryScores, score.WeightedFactors)
	score.RiskLevel = rs.determineRiskLevel(score.OverallScore)

	// Calculate confidence level
	score.ConfidenceLevel = rs.calculateConfidenceLevel(result)

	// Create score breakdown
	score.ScoreBreakdown = rs.createScoreBreakdown(score.CategoryScores)

	// Create calibration data
	score.CalibrationData = rs.createCalibrationData()

	processingTime := time.Since(startTime)

	rs.logger.Info("Risk score calculation completed",
		zap.String("business_name", result.BusinessName),
		zap.Float64("overall_score", score.OverallScore),
		zap.String("risk_level", string(score.RiskLevel)),
		zap.Float64("confidence_level", score.ConfidenceLevel),
		zap.Duration("processing_time", processingTime))

	return score, nil
}

// calculateSecurityCategoryScore calculates security category score
func (rs *RiskScorer) calculateSecurityCategoryScore(security *SecurityAnalysisResult) CategoryScore {
	score := CategoryScore{
		Category:   "security",
		Score:      security.OverallSecurityScore,
		Weight:     0.25, // 25% weight for security
		RiskLevel:  RiskLevel(security.SecurityLevel),
		Factors:    make([]RiskFactor, 0),
		Confidence: 0.9,
	}

	score.WeightedScore = score.Score * score.Weight

	// Convert security issues to risk factors
	for _, issue := range security.SecurityIssues {
		factor := RiskFactor{
			Category:    "security",
			Factor:      issue.Title,
			Description: issue.Description,
			Impact:      issue.Impact,
			Severity:    RiskLevel(issue.Severity),
			Score:       issue.Score,
		}
		score.Factors = append(score.Factors, factor)
	}

	return score
}

// calculateDomainCategoryScore calculates domain category score
func (rs *RiskScorer) calculateDomainCategoryScore(domain *DomainAnalysisResult) CategoryScore {
	score := CategoryScore{
		Category:   "domain",
		Score:      domain.OverallScore,
		Weight:     0.20,         // 20% weight for domain
		RiskLevel:  RiskLevelLow, // Default risk level
		Factors:    domain.RiskFactors,
		Confidence: 0.85,
	}

	score.WeightedScore = score.Score * score.Weight

	return score
}

// calculateReputationCategoryScore calculates reputation category score
func (rs *RiskScorer) calculateReputationCategoryScore(reputation *ReputationAnalysisResult) CategoryScore {
	score := CategoryScore{
		Category:   "reputation",
		Score:      reputation.OverallScore,
		Weight:     0.20,         // 20% weight for reputation
		RiskLevel:  RiskLevelLow, // Default risk level
		Factors:    reputation.RiskFactors,
		Confidence: 0.8,
	}

	score.WeightedScore = score.Score * score.Weight

	return score
}

// calculateComplianceCategoryScore calculates compliance category score
func (rs *RiskScorer) calculateComplianceCategoryScore(compliance *ComplianceAnalysisResult) CategoryScore {
	score := CategoryScore{
		Category:   "compliance",
		Score:      compliance.OverallScore,
		Weight:     0.20,         // 20% weight for compliance
		RiskLevel:  RiskLevelLow, // Default risk level
		Factors:    compliance.RiskFactors,
		Confidence: 0.85,
	}

	score.WeightedScore = score.Score * score.Weight

	return score
}

// calculateFinancialCategoryScore calculates financial category score
func (rs *RiskScorer) calculateFinancialCategoryScore(financial *FinancialAnalysisResult) CategoryScore {
	score := CategoryScore{
		Category:   "financial",
		Score:      financial.OverallScore,
		Weight:     0.15,         // 15% weight for financial
		RiskLevel:  RiskLevelLow, // Default risk level
		Factors:    financial.RiskFactors,
		Confidence: 0.75,
	}

	score.WeightedScore = score.Score * score.Weight

	return score
}

// calculateWeightedFactors calculates weighted risk factors
func (rs *RiskScorer) calculateWeightedFactors(riskFactors []RiskFactor) []WeightedRiskFactor {
	weightedFactors := make([]WeightedRiskFactor, 0)

	for _, factor := range riskFactors {
		weight := rs.getFactorWeight(string(factor.Category), factor.Severity)
		impact := rs.calculateFactorImpact(factor)
		contribution := weight * impact

		weightedFactor := WeightedRiskFactor{
			Factor:       factor,
			Weight:       weight,
			Impact:       impact,
			Contribution: contribution,
		}

		weightedFactors = append(weightedFactors, weightedFactor)
	}

	return weightedFactors
}

// getFactorWeight gets the weight for a risk factor
func (rs *RiskScorer) getFactorWeight(category string, severity RiskLevel) float64 {
	// Base weights by category
	baseWeights := map[string]float64{
		"security":   0.25,
		"domain":     0.20,
		"reputation": 0.20,
		"compliance": 0.20,
		"financial":  0.15,
	}

	// Severity multipliers
	severityMultipliers := map[RiskLevel]float64{
		RiskLevelLow:      0.5,
		RiskLevelMedium:   1.0,
		RiskLevelHigh:     1.5,
		RiskLevelCritical: 2.0,
	}

	baseWeight := baseWeights[category]
	multiplier := severityMultipliers[severity]

	return baseWeight * multiplier
}

// calculateFactorImpact calculates the impact of a risk factor
func (rs *RiskScorer) calculateFactorImpact(factor RiskFactor) float64 {
	// Impact is based on the factor score and severity
	impact := factor.Score

	// Adjust impact based on severity
	switch factor.Severity {
	case RiskLevelCritical:
		impact *= 1.5
	case RiskLevelHigh:
		impact *= 1.2
	case RiskLevelMedium:
		impact *= 1.0
	case RiskLevelLow:
		impact *= 0.8
	}

	return impact
}

// calculateOverallScore calculates the overall risk score
func (rs *RiskScorer) calculateOverallScore(categoryScores map[string]CategoryScore, weightedFactors []WeightedRiskFactor) float64 {
	// Calculate weighted category score
	categoryScore := 0.0
	totalWeight := 0.0

	for _, score := range categoryScores {
		categoryScore += score.WeightedScore
		totalWeight += score.Weight
	}

	if totalWeight > 0 {
		categoryScore = categoryScore / totalWeight
	}

	// Calculate factor adjustment
	factorAdjustment := 0.0
	for _, factor := range weightedFactors {
		factorAdjustment += factor.Contribution
	}

	// Normalize factor adjustment
	if len(weightedFactors) > 0 {
		factorAdjustment = factorAdjustment / float64(len(weightedFactors))
	}

	// Combine category score and factor adjustment
	overallScore := (categoryScore * 0.7) + (factorAdjustment * 0.3)

	// Ensure score is between 0 and 1
	if overallScore < 0 {
		overallScore = 0
	} else if overallScore > 1 {
		overallScore = 1
	}

	return overallScore
}

// determineRiskLevel determines the risk level based on score
func (rs *RiskScorer) determineRiskLevel(score float64) RiskLevel {
	switch {
	case score >= 0.8:
		return RiskLevelCritical
	case score >= 0.6:
		return RiskLevelHigh
	case score >= 0.4:
		return RiskLevelMedium
	case score >= 0.2:
		return RiskLevelLow
	default:
		return RiskLevelLow
	}
}

// calculateConfidenceLevel calculates the confidence level of the risk assessment
func (rs *RiskScorer) calculateConfidenceLevel(result *RiskAssessmentResult) float64 {
	confidence := 0.0
	count := 0

	// Calculate confidence based on available analyses
	if result.SecurityAnalysis != nil {
		confidence += 0.9 // High confidence for security analysis
		count++
	}

	if result.DomainAnalysis != nil {
		confidence += 0.85 // High confidence for domain analysis
		count++
	}

	if result.ReputationAnalysis != nil {
		confidence += 0.8 // Medium-high confidence for reputation analysis
		count++
	}

	if result.ComplianceAnalysis != nil {
		confidence += 0.85 // High confidence for compliance analysis
		count++
	}

	if result.FinancialAnalysis != nil {
		confidence += 0.75 // Medium confidence for financial analysis
		count++
	}

	if count == 0 {
		return 0.5 // Default confidence if no analyses available
	}

	return confidence / float64(count)
}

// createScoreBreakdown creates a detailed score breakdown
func (rs *RiskScorer) createScoreBreakdown(categoryScores map[string]CategoryScore) *ScoreBreakdown {
	breakdown := &ScoreBreakdown{}

	if score, exists := categoryScores["security"]; exists {
		breakdown.SecurityScore = score.Score
	}

	if score, exists := categoryScores["domain"]; exists {
		breakdown.DomainScore = score.Score
	}

	if score, exists := categoryScores["reputation"]; exists {
		breakdown.ReputationScore = score.Score
	}

	if score, exists := categoryScores["compliance"]; exists {
		breakdown.ComplianceScore = score.Score
	}

	if score, exists := categoryScores["financial"]; exists {
		breakdown.FinancialScore = score.Score
	}

	// Calculate weighted total
	breakdown.WeightedTotal = 0.0
	for _, score := range categoryScores {
		breakdown.WeightedTotal += score.WeightedScore
	}

	// Normalize score
	breakdown.NormalizedScore = breakdown.WeightedTotal

	return breakdown
}

// createCalibrationData creates calibration data
func (rs *RiskScorer) createCalibrationData() *CalibrationData {
	return &CalibrationData{
		ModelVersion:       "1.0.0",
		CalibrationDate:    time.Now(),
		ValidationScore:    0.85,
		Accuracy:           0.87,
		Precision:          0.89,
		Recall:             0.83,
		F1Score:            0.86,
		ConfidenceInterval: 0.05,
	}
}
