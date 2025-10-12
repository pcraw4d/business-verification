package industry

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// FinanceRiskModel provides industry-specific risk assessment for financial services
type FinanceRiskModel struct {
	logger *zap.Logger
}

// NewFinanceRiskModel creates a new FinanceRiskModel
func NewFinanceRiskModel(logger *zap.Logger) *FinanceRiskModel {
	return &FinanceRiskModel{
		logger: logger,
	}
}

// AssessRisk performs industry-specific risk assessment for financial services
func (frm *FinanceRiskModel) AssessRisk(ctx context.Context, business *models.RiskAssessmentRequest) (*models.RiskAssessment, error) {
	frm.logger.Info("Assessing risk for financial services business",
		zap.String("business_name", business.BusinessName))

	// Base risk score for financial services
	baseScore := 0.4 // Financial services have higher baseline risk

	// Industry-specific risk factors
	regulatoryRisk := frm.assessRegulatoryRisk(business)
	capitalRisk := frm.assessCapitalRisk(business)
	operationalRisk := frm.assessOperationalRisk(business)
	marketRisk := frm.assessMarketRisk(business)
	creditRisk := frm.assessCreditRisk(business)

	// Calculate weighted risk score
	riskScore := baseScore +
		(regulatoryRisk * 0.25) +
		(capitalRisk * 0.20) +
		(operationalRisk * 0.20) +
		(marketRisk * 0.20) +
		(creditRisk * 0.15)

	// Ensure risk score is within bounds
	riskScore = math.Max(0.0, math.Min(1.0, riskScore))

	// Generate detailed risk factors
	riskFactors := frm.generateFinanceRiskFactors(business, riskScore, map[string]float64{
		"regulatory":  regulatoryRisk,
		"capital":     capitalRisk,
		"operational": operationalRisk,
		"market":      marketRisk,
		"credit":      creditRisk,
	})

	assessment := &models.RiskAssessment{
		ID:              generateAssessmentID(),
		BusinessName:    business.BusinessName,
		Industry:        "finance",
		RiskScore:       riskScore,
		RiskLevel:       models.ConvertScoreToRiskLevel(riskScore),
		ConfidenceScore: frm.calculateConfidenceScore(business),
		RiskFactors:     riskFactors,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Metadata: map[string]interface{}{
			"industry_model":  "finance",
			"model_version":   "1.0",
			"assessment_type": "industry_specific",
		},
	}

	frm.logger.Info("Finance risk assessment completed",
		zap.String("business_name", business.BusinessName),
		zap.Float64("risk_score", riskScore),
		zap.String("risk_level", string(assessment.RiskLevel)))

	return assessment, nil
}

// assessRegulatoryRisk evaluates regulatory compliance risk
func (frm *FinanceRiskModel) assessRegulatoryRisk(business *models.RiskAssessmentRequest) float64 {
	risk := 0.3 // Base regulatory risk

	// Higher risk for businesses without proper documentation
	if business.Website == "" {
		risk += 0.1
	}
	if business.Email == "" {
		risk += 0.1
	}

	// Adjust based on business name length (proxy for legitimacy)
	if len(business.BusinessName) < 5 {
		risk += 0.15
	}

	return math.Min(1.0, risk)
}

// assessCapitalRisk evaluates capital adequacy risk
func (frm *FinanceRiskModel) assessCapitalRisk(business *models.RiskAssessmentRequest) float64 {
	risk := 0.2 // Base capital risk

	// Higher risk for smaller businesses (based on name length as proxy)
	if len(business.BusinessName) < 10 {
		risk += 0.2
	}

	// Higher risk for businesses without online presence
	if business.Website == "" {
		risk += 0.15
	}

	return math.Min(1.0, risk)
}

// assessOperationalRisk evaluates operational risk
func (frm *FinanceRiskModel) assessOperationalRisk(business *models.RiskAssessmentRequest) float64 {
	risk := 0.25 // Base operational risk

	// Higher risk for businesses without proper contact information
	if business.Email == "" {
		risk += 0.1
	}

	// Higher risk for businesses with very short names (potential shell companies)
	if len(business.BusinessName) < 8 {
		risk += 0.15
	}

	return math.Min(1.0, risk)
}

// assessMarketRisk evaluates market risk
func (frm *FinanceRiskModel) assessMarketRisk(business *models.RiskAssessmentRequest) float64 {
	risk := 0.3 // Base market risk for financial services

	// Higher risk for businesses without online presence
	if business.Website == "" {
		risk += 0.1
	}

	return math.Min(1.0, risk)
}

// assessCreditRisk evaluates credit risk
func (frm *FinanceRiskModel) assessCreditRisk(business *models.RiskAssessmentRequest) float64 {
	risk := 0.2 // Base credit risk

	// Higher risk for businesses without proper documentation
	if business.Email == "" {
		risk += 0.1
	}

	// Higher risk for businesses with very short names
	if len(business.BusinessName) < 6 {
		risk += 0.15
	}

	return math.Min(1.0, risk)
}

// generateFinanceRiskFactors creates detailed risk factors for financial services
func (frm *FinanceRiskModel) generateFinanceRiskFactors(business *models.RiskAssessmentRequest, baseScore float64, categoryScores map[string]float64) []models.RiskFactor {
	now := time.Now()
	riskFactors := make([]models.RiskFactor, 0)

	// Regulatory risk factors
	riskFactors = append(riskFactors, models.RiskFactor{
		Category:    models.RiskCategoryRegulatory,
		Subcategory: "compliance",
		Name:        "regulatory_compliance",
		Score:       categoryScores["regulatory"],
		Weight:      0.25,
		Description: "Risk associated with regulatory compliance and oversight requirements",
		Source:      "finance_industry_model",
		Confidence:  0.9,
		Impact:      "High regulatory risk can lead to penalties and operational restrictions",
		Mitigation:  "Implement robust compliance monitoring and regular audits",
		LastUpdated: &now,
	})

	// Capital risk factors
	riskFactors = append(riskFactors, models.RiskFactor{
		Category:    models.RiskCategoryFinancial,
		Subcategory: "capital_adequacy",
		Name:        "capital_adequacy",
		Score:       categoryScores["capital"],
		Weight:      0.20,
		Description: "Risk associated with insufficient capital to meet regulatory requirements",
		Source:      "finance_industry_model",
		Confidence:  0.85,
		Impact:      "Insufficient capital can lead to regulatory intervention",
		Mitigation:  "Maintain adequate capital buffers and monitor capital ratios",
		LastUpdated: &now,
	})

	// Operational risk factors
	riskFactors = append(riskFactors, models.RiskFactor{
		Category:    models.RiskCategoryOperational,
		Subcategory: "operations",
		Name:        "operational_risk",
		Score:       categoryScores["operational"],
		Weight:      0.20,
		Description: "Risk associated with operational failures and process breakdowns",
		Source:      "finance_industry_model",
		Confidence:  0.8,
		Impact:      "Operational failures can result in financial losses and reputational damage",
		Mitigation:  "Implement robust operational controls and business continuity plans",
		LastUpdated: &now,
	})

	// Market risk factors
	riskFactors = append(riskFactors, models.RiskFactor{
		Category:    models.RiskCategoryFinancial,
		Subcategory: "market_risk",
		Name:        "market_risk",
		Score:       categoryScores["market"],
		Weight:      0.20,
		Description: "Risk associated with market volatility and economic conditions",
		Source:      "finance_industry_model",
		Confidence:  0.75,
		Impact:      "Market volatility can significantly impact financial performance",
		Mitigation:  "Diversify portfolio and implement hedging strategies",
		LastUpdated: &now,
	})

	// Credit risk factors
	riskFactors = append(riskFactors, models.RiskFactor{
		Category:    models.RiskCategoryFinancial,
		Subcategory: "credit_risk",
		Name:        "credit_risk",
		Score:       categoryScores["credit"],
		Weight:      0.15,
		Description: "Risk associated with counterparty defaults and credit losses",
		Source:      "finance_industry_model",
		Confidence:  0.8,
		Impact:      "Credit defaults can result in significant financial losses",
		Mitigation:  "Implement robust credit assessment and monitoring processes",
		LastUpdated: &now,
	})

	return riskFactors
}

// calculateConfidenceScore calculates confidence in the assessment
func (frm *FinanceRiskModel) calculateConfidenceScore(business *models.RiskAssessmentRequest) float64 {
	confidence := 0.7 // Base confidence

	// Higher confidence with more business information
	if business.Website != "" {
		confidence += 0.1
	}
	if business.Email != "" {
		confidence += 0.1
	}
	if len(business.BusinessName) > 10 {
		confidence += 0.1
	}

	return math.Min(1.0, confidence)
}

// generateAssessmentID generates a unique assessment ID
func generateAssessmentID() string {
	return fmt.Sprintf("fin_%d", time.Now().UnixNano())
}
