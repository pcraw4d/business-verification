package scenario

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// StressTester performs stress testing for risk scenarios
type StressTester struct {
	logger *zap.Logger
}

// NewStressTester creates a new stress tester
func NewStressTester(logger *zap.Logger) *StressTester {
	return &StressTester{
		logger: logger,
	}
}

// StressTestResult represents the result of a stress test
type StressTestResult struct {
	TestName            string                 `json:"test_name"`
	BaseRiskScore       float64                `json:"base_risk_score"`
	StressedRiskScore   float64                `json:"stressed_risk_score"`
	RiskIncrease        float64                `json:"risk_increase"`
	RiskIncreasePercent float64                `json:"risk_increase_percent"`
	StressFactors       []StressFactor         `json:"stress_factors"`
	ImpactAnalysis      ImpactAnalysis         `json:"impact_analysis"`
	MitigationOptions   []MitigationOption     `json:"mitigation_options"`
	TestMetadata        map[string]interface{} `json:"test_metadata"`
}

// StressFactor represents a factor applied during stress testing
type StressFactor struct {
	FactorName    string  `json:"factor_name"`
	FactorType    string  `json:"factor_type"`
	BaseValue     float64 `json:"base_value"`
	StressedValue float64 `json:"stressed_value"`
	Impact        float64 `json:"impact"`
	Description   string  `json:"description"`
	Severity      string  `json:"severity"`
}

// ImpactAnalysis represents the analysis of stress test impact
type ImpactAnalysis struct {
	OverallImpact      string   `json:"overall_impact"`
	RiskLevelChange    string   `json:"risk_level_change"`
	CriticalFactors    []string `json:"critical_factors"`
	ModerateFactors    []string `json:"moderate_factors"`
	LowFactors         []string `json:"low_factors"`
	ImpactSummary      string   `json:"impact_summary"`
	RecoveryTime       string   `json:"recovery_time"`
	BusinessContinuity string   `json:"business_continuity"`
}

// MitigationOption represents a mitigation option for stress scenarios
type MitigationOption struct {
	OptionName         string  `json:"option_name"`
	Description        string  `json:"description"`
	Effectiveness      float64 `json:"effectiveness"`
	ImplementationCost string  `json:"implementation_cost"`
	TimeToImplement    string  `json:"time_to_implement"`
	RiskReduction      float64 `json:"risk_reduction"`
	Priority           string  `json:"priority"`
}

// StressTestScenario defines a stress test scenario
type StressTestScenario struct {
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Severity    string                        `json:"severity"`
	Duration    int                           `json:"duration"` // months
	Factors     map[string]StressFactor       `json:"factors"`
	Business    *models.RiskAssessmentRequest `json:"business"`
}

// RunStressTest runs a stress test for a given scenario
func (st *StressTester) RunStressTest(ctx context.Context, scenario *StressTestScenario, baseRiskScore float64) (*StressTestResult, error) {
	st.logger.Info("Running stress test",
		zap.String("scenario", scenario.Name),
		zap.Float64("base_risk_score", baseRiskScore))

	// Calculate stressed risk score
	stressedScore := st.calculateStressedRiskScore(baseRiskScore, scenario)

	// Calculate risk increase
	riskIncrease := stressedScore - baseRiskScore
	riskIncreasePercent := (riskIncrease / baseRiskScore) * 100

	// Analyze impact
	impactAnalysis := st.analyzeImpact(baseRiskScore, stressedScore, scenario)

	// Generate mitigation options
	mitigationOptions := st.generateMitigationOptions(scenario, riskIncrease)

	// Create test metadata
	metadata := map[string]interface{}{
		"test_timestamp":    time.Now(),
		"scenario_severity": scenario.Severity,
		"test_duration":     scenario.Duration,
		"factor_count":      len(scenario.Factors),
	}

	result := &StressTestResult{
		TestName:            scenario.Name,
		BaseRiskScore:       baseRiskScore,
		StressedRiskScore:   stressedScore,
		RiskIncrease:        riskIncrease,
		RiskIncreasePercent: riskIncreasePercent,
		StressFactors:       st.convertStressFactors(scenario.Factors),
		ImpactAnalysis:      impactAnalysis,
		MitigationOptions:   mitigationOptions,
		TestMetadata:        metadata,
	}

	st.logger.Info("Stress test completed",
		zap.String("scenario", scenario.Name),
		zap.Float64("stressed_risk_score", stressedScore),
		zap.Float64("risk_increase", riskIncrease))

	return result, nil
}

// calculateStressedRiskScore calculates the risk score under stress conditions
func (st *StressTester) calculateStressedRiskScore(baseScore float64, scenario *StressTestScenario) float64 {
	stressedScore := baseScore

	// Apply stress factors
	for factorName, factor := range scenario.Factors {
		impact := st.calculateFactorImpact(factor, scenario.Business)
		stressedScore += impact

		// Update factor with calculated impact
		factor.Impact = impact
		scenario.Factors[factorName] = factor
	}

	// Apply scenario severity multiplier
	severityMultiplier := st.getSeverityMultiplier(scenario.Severity)
	stressedScore *= severityMultiplier

	// Apply duration adjustment
	durationAdjustment := st.getDurationAdjustment(scenario.Duration)
	stressedScore += durationAdjustment

	// Apply business-specific adjustments
	stressedScore = st.applyBusinessStressAdjustments(stressedScore, scenario.Business)

	// Ensure score is between 0 and 1
	if stressedScore > 1.0 {
		stressedScore = 1.0
	} else if stressedScore < 0.0 {
		stressedScore = 0.0
	}

	return stressedScore
}

// calculateFactorImpact calculates the impact of a stress factor
func (st *StressTester) calculateFactorImpact(factor StressFactor, business *models.RiskAssessmentRequest) float64 {
	// Base impact from factor value change
	valueChange := factor.StressedValue - factor.BaseValue
	baseImpact := valueChange * 0.1 // Scale factor

	// Factor type specific adjustments
	switch factor.FactorType {
	case "market_volatility":
		baseImpact *= 0.15
	case "regulatory_change":
		baseImpact *= 0.12
	case "economic_shock":
		baseImpact *= 0.18
	case "operational_disruption":
		baseImpact *= 0.14
	case "cybersecurity_incident":
		baseImpact *= 0.16
	case "natural_disaster":
		baseImpact *= 0.20
	case "geopolitical_event":
		baseImpact *= 0.13
	case "technology_failure":
		baseImpact *= 0.11
	default:
		baseImpact *= 0.10
	}

	// Business-specific adjustments
	if business != nil {
		baseImpact = st.adjustImpactForBusiness(baseImpact, factor, business)
	}

	return baseImpact
}

// adjustImpactForBusiness adjusts impact based on business characteristics
func (st *StressTester) adjustImpactForBusiness(impact float64, factor StressFactor, business *models.RiskAssessmentRequest) float64 {
	// Industry-specific adjustments
	switch business.Industry {
	case "technology":
		if factor.FactorType == "technology_failure" || factor.FactorType == "cybersecurity_incident" {
			impact *= 1.5 // Technology companies more sensitive to tech factors
		}
	case "finance":
		if factor.FactorType == "market_volatility" || factor.FactorType == "regulatory_change" {
			impact *= 1.3 // Financial companies more sensitive to market/regulatory factors
		}
	case "healthcare":
		if factor.FactorType == "regulatory_change" {
			impact *= 1.4 // Healthcare companies more sensitive to regulatory changes
		}
	case "retail":
		if factor.FactorType == "economic_shock" || factor.FactorType == "operational_disruption" {
			impact *= 1.2 // Retail companies more sensitive to economic/operational factors
		}
	}

	// Country-specific adjustments
	switch business.Country {
	case "US", "CA", "GB":
		impact *= 0.9 // Lower impact in stable countries
	default:
		impact *= 1.1 // Higher impact in other countries
	}

	// Digital presence adjustments
	if business.Website != "" {
		if factor.FactorType == "cybersecurity_incident" || factor.FactorType == "technology_failure" {
			impact *= 1.2 // Higher impact for digital businesses
		}
	}

	return impact
}

// getSeverityMultiplier returns a multiplier based on scenario severity
func (st *StressTester) getSeverityMultiplier(severity string) float64 {
	switch severity {
	case "mild":
		return 1.1
	case "moderate":
		return 1.3
	case "severe":
		return 1.6
	case "extreme":
		return 2.0
	default:
		return 1.2
	}
}

// getDurationAdjustment returns an adjustment based on scenario duration
func (st *StressTester) getDurationAdjustment(duration int) float64 {
	// Longer duration scenarios have higher cumulative impact
	if duration <= 1 {
		return 0.0
	} else if duration <= 3 {
		return 0.05
	} else if duration <= 6 {
		return 0.10
	} else if duration <= 12 {
		return 0.15
	} else {
		return 0.20
	}
}

// applyBusinessStressAdjustments applies business-specific stress adjustments
func (st *StressTester) applyBusinessStressAdjustments(score float64, business *models.RiskAssessmentRequest) float64 {
	if business == nil {
		return score
	}

	// Business size adjustment (using name length as proxy)
	if len(business.BusinessName) > 20 {
		score += 0.05 // Larger businesses may be more resilient
	} else if len(business.BusinessName) < 10 {
		score += 0.10 // Smaller businesses may be more vulnerable
	}

	// Address completeness adjustment
	if len(business.BusinessAddress) > 30 {
		score -= 0.02 // Complete address suggests established business
	}

	// Contact information adjustment
	contactScore := 0.0
	if business.Phone != "" {
		contactScore += 0.01
	}
	if business.Email != "" {
		contactScore += 0.01
	}
	if business.Website != "" {
		contactScore += 0.02
	}
	score -= contactScore

	return score
}

// analyzeImpact analyzes the impact of the stress test
func (st *StressTester) analyzeImpact(baseScore, stressedScore float64, scenario *StressTestScenario) ImpactAnalysis {
	riskIncrease := stressedScore - baseScore

	// Determine overall impact
	overallImpact := "low"
	if riskIncrease > 0.3 {
		overallImpact = "severe"
	} else if riskIncrease > 0.2 {
		overallImpact = "high"
	} else if riskIncrease > 0.1 {
		overallImpact = "moderate"
	}

	// Determine risk level change
	baseLevel := st.getRiskLevel(baseScore)
	stressedLevel := st.getRiskLevel(stressedScore)
	riskLevelChange := fmt.Sprintf("%s to %s", baseLevel, stressedLevel)

	// Categorize factors by impact
	criticalFactors := make([]string, 0)
	moderateFactors := make([]string, 0)
	lowFactors := make([]string, 0)

	for factorName, factor := range scenario.Factors {
		if factor.Impact > 0.1 {
			criticalFactors = append(criticalFactors, factorName)
		} else if factor.Impact > 0.05 {
			moderateFactors = append(moderateFactors, factorName)
		} else {
			lowFactors = append(lowFactors, factorName)
		}
	}

	// Generate impact summary
	impactSummary := st.generateImpactSummary(overallImpact, riskIncrease, len(criticalFactors))

	// Estimate recovery time
	recoveryTime := st.estimateRecoveryTime(overallImpact, scenario.Duration)

	// Assess business continuity
	businessContinuity := st.assessBusinessContinuity(stressedScore, scenario.Severity)

	return ImpactAnalysis{
		OverallImpact:      overallImpact,
		RiskLevelChange:    riskLevelChange,
		CriticalFactors:    criticalFactors,
		ModerateFactors:    moderateFactors,
		LowFactors:         lowFactors,
		ImpactSummary:      impactSummary,
		RecoveryTime:       recoveryTime,
		BusinessContinuity: businessContinuity,
	}
}

// getRiskLevel returns the risk level for a given score
func (st *StressTester) getRiskLevel(score float64) string {
	if score > 0.8 {
		return "critical"
	} else if score > 0.6 {
		return "high"
	} else if score > 0.4 {
		return "medium"
	} else {
		return "low"
	}
}

// generateImpactSummary generates a summary of the impact
func (st *StressTester) generateImpactSummary(overallImpact string, riskIncrease float64, criticalFactorCount int) string {
	return fmt.Sprintf("The stress test shows a %s impact with a risk increase of %.1f%%. %d critical factors contribute to this impact, requiring immediate attention and mitigation measures.",
		overallImpact, riskIncrease*100, criticalFactorCount)
}

// estimateRecoveryTime estimates the time to recover from the stress scenario
func (st *StressTester) estimateRecoveryTime(overallImpact string, duration int) string {
	switch overallImpact {
	case "severe":
		return fmt.Sprintf("%d-18 months", duration*2)
	case "high":
		return fmt.Sprintf("%d-12 months", duration+3)
	case "moderate":
		return fmt.Sprintf("%d-6 months", duration+1)
	default:
		return fmt.Sprintf("%d-3 months", duration)
	}
}

// assessBusinessContinuity assesses the impact on business continuity
func (st *StressTester) assessBusinessContinuity(stressedScore float64, severity string) string {
	if stressedScore > 0.8 {
		return "Critical - Business operations severely impacted, immediate intervention required"
	} else if stressedScore > 0.6 {
		return "High - Significant operational disruptions expected, contingency plans needed"
	} else if stressedScore > 0.4 {
		return "Moderate - Some operational impact expected, monitoring and adjustments required"
	} else {
		return "Low - Minimal impact on business operations, standard monitoring sufficient"
	}
}

// generateMitigationOptions generates mitigation options for the stress scenario
func (st *StressTester) generateMitigationOptions(scenario *StressTestScenario, riskIncrease float64) []MitigationOption {
	options := make([]MitigationOption, 0)

	// Generate options based on stress factors
	for factorName, factor := range scenario.Factors {
		if factor.Impact > 0.05 { // Only for significant factors
			option := st.createMitigationOption(factorName, factor, riskIncrease)
			options = append(options, option)
		}
	}

	// Add general mitigation options
	options = append(options, MitigationOption{
		OptionName:         "Enhanced Monitoring",
		Description:        "Implement enhanced monitoring and early warning systems",
		Effectiveness:      0.3,
		ImplementationCost: "Low",
		TimeToImplement:    "1-2 weeks",
		RiskReduction:      0.1,
		Priority:           "High",
	})

	options = append(options, MitigationOption{
		OptionName:         "Contingency Planning",
		Description:        "Develop and implement comprehensive contingency plans",
		Effectiveness:      0.5,
		ImplementationCost: "Medium",
		TimeToImplement:    "1-3 months",
		RiskReduction:      0.2,
		Priority:           "High",
	})

	options = append(options, MitigationOption{
		OptionName:         "Risk Transfer",
		Description:        "Transfer risk through insurance or hedging strategies",
		Effectiveness:      0.4,
		ImplementationCost: "Medium",
		TimeToImplement:    "2-4 weeks",
		RiskReduction:      0.15,
		Priority:           "Medium",
	})

	return options
}

// createMitigationOption creates a mitigation option for a specific factor
func (st *StressTester) createMitigationOption(factorName string, factor StressFactor, riskIncrease float64) MitigationOption {
	optionName := fmt.Sprintf("Mitigate %s", factorName)
	description := fmt.Sprintf("Implement specific measures to address %s risk factor", factorName)

	effectiveness := 0.4
	if factor.Impact > 0.15 {
		effectiveness = 0.6
	} else if factor.Impact > 0.1 {
		effectiveness = 0.5
	}

	priority := "Medium"
	if factor.Impact > 0.15 {
		priority = "High"
	} else if factor.Impact < 0.05 {
		priority = "Low"
	}

	implementationCost := "Medium"
	if factor.FactorType == "technology_failure" || factor.FactorType == "cybersecurity_incident" {
		implementationCost = "High"
	} else if factor.FactorType == "operational_disruption" {
		implementationCost = "Low"
	}

	timeToImplement := "1-2 months"
	if factor.FactorType == "regulatory_change" {
		timeToImplement = "3-6 months"
	} else if factor.FactorType == "technology_failure" {
		timeToImplement = "2-4 weeks"
	}

	riskReduction := factor.Impact * effectiveness

	return MitigationOption{
		OptionName:         optionName,
		Description:        description,
		Effectiveness:      effectiveness,
		ImplementationCost: implementationCost,
		TimeToImplement:    timeToImplement,
		RiskReduction:      riskReduction,
		Priority:           priority,
	}
}

// convertStressFactors converts stress factors to the result format
func (st *StressTester) convertStressFactors(factors map[string]StressFactor) []StressFactor {
	result := make([]StressFactor, 0, len(factors))
	for _, factor := range factors {
		result = append(result, factor)
	}
	return result
}

// RunMultipleStressTests runs multiple stress test scenarios
func (st *StressTester) RunMultipleStressTests(ctx context.Context, scenarios []*StressTestScenario, baseRiskScore float64) ([]*StressTestResult, error) {
	st.logger.Info("Running multiple stress tests",
		zap.Int("scenario_count", len(scenarios)))

	results := make([]*StressTestResult, 0, len(scenarios))

	for _, scenario := range scenarios {
		result, err := st.RunStressTest(ctx, scenario, baseRiskScore)
		if err != nil {
			return nil, fmt.Errorf("failed to run stress test for scenario %s: %w", scenario.Name, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// CreateStandardScenarios creates standard stress test scenarios
func (st *StressTester) CreateStandardScenarios(business *models.RiskAssessmentRequest) []*StressTestScenario {
	scenarios := []*StressTestScenario{
		st.createMarketCrisisScenario(business),
		st.createRegulatoryChangeScenario(business),
		st.createOperationalDisruptionScenario(business),
		st.createCybersecurityIncidentScenario(business),
		st.createEconomicRecessionScenario(business),
		st.createNaturalDisasterScenario(business),
	}

	return scenarios
}

// createMarketCrisisScenario creates a market crisis stress scenario
func (st *StressTester) createMarketCrisisScenario(business *models.RiskAssessmentRequest) *StressTestScenario {
	return &StressTestScenario{
		Name:        "Market Crisis",
		Description: "Simulation of a severe market downturn affecting business operations",
		Severity:    "severe",
		Duration:    12,
		Business:    business,
		Factors: map[string]StressFactor{
			"market_volatility": {
				FactorName:    "Market Volatility",
				FactorType:    "market_volatility",
				BaseValue:     0.3,
				StressedValue: 0.8,
				Description:   "Significant increase in market volatility",
				Severity:      "high",
			},
			"economic_indicators": {
				FactorName:    "Economic Indicators",
				FactorType:    "economic_shock",
				BaseValue:     0.4,
				StressedValue: 0.9,
				Description:   "Deterioration of key economic indicators",
				Severity:      "high",
			},
		},
	}
}

// createRegulatoryChangeScenario creates a regulatory change stress scenario
func (st *StressTester) createRegulatoryChangeScenario(business *models.RiskAssessmentRequest) *StressTestScenario {
	return &StressTestScenario{
		Name:        "Regulatory Change",
		Description: "Simulation of significant regulatory changes affecting the business",
		Severity:    "moderate",
		Duration:    6,
		Business:    business,
		Factors: map[string]StressFactor{
			"regulatory_environment": {
				FactorName:    "Regulatory Environment",
				FactorType:    "regulatory_change",
				BaseValue:     0.2,
				StressedValue: 0.7,
				Description:   "Major regulatory changes requiring compliance",
				Severity:      "moderate",
			},
		},
	}
}

// createOperationalDisruptionScenario creates an operational disruption stress scenario
func (st *StressTester) createOperationalDisruptionScenario(business *models.RiskAssessmentRequest) *StressTestScenario {
	return &StressTestScenario{
		Name:        "Operational Disruption",
		Description: "Simulation of major operational disruptions",
		Severity:    "severe",
		Duration:    3,
		Business:    business,
		Factors: map[string]StressFactor{
			"operational_efficiency": {
				FactorName:    "Operational Efficiency",
				FactorType:    "operational_disruption",
				BaseValue:     0.3,
				StressedValue: 0.8,
				Description:   "Significant reduction in operational efficiency",
				Severity:      "high",
			},
		},
	}
}

// createCybersecurityIncidentScenario creates a cybersecurity incident stress scenario
func (st *StressTester) createCybersecurityIncidentScenario(business *models.RiskAssessmentRequest) *StressTestScenario {
	return &StressTestScenario{
		Name:        "Cybersecurity Incident",
		Description: "Simulation of a major cybersecurity breach",
		Severity:    "severe",
		Duration:    6,
		Business:    business,
		Factors: map[string]StressFactor{
			"cybersecurity_risk": {
				FactorName:    "Cybersecurity Risk",
				FactorType:    "cybersecurity_incident",
				BaseValue:     0.2,
				StressedValue: 0.9,
				Description:   "Major cybersecurity incident affecting operations",
				Severity:      "critical",
			},
		},
	}
}

// createEconomicRecessionScenario creates an economic recession stress scenario
func (st *StressTester) createEconomicRecessionScenario(business *models.RiskAssessmentRequest) *StressTestScenario {
	return &StressTestScenario{
		Name:        "Economic Recession",
		Description: "Simulation of an economic recession",
		Severity:    "severe",
		Duration:    18,
		Business:    business,
		Factors: map[string]StressFactor{
			"economic_conditions": {
				FactorName:    "Economic Conditions",
				FactorType:    "economic_shock",
				BaseValue:     0.4,
				StressedValue: 0.9,
				Description:   "Severe economic recession",
				Severity:      "high",
			},
			"market_conditions": {
				FactorName:    "Market Conditions",
				FactorType:    "market_volatility",
				BaseValue:     0.3,
				StressedValue: 0.8,
				Description:   "Adverse market conditions during recession",
				Severity:      "high",
			},
		},
	}
}

// createNaturalDisasterScenario creates a natural disaster stress scenario
func (st *StressTester) createNaturalDisasterScenario(business *models.RiskAssessmentRequest) *StressTestScenario {
	return &StressTestScenario{
		Name:        "Natural Disaster",
		Description: "Simulation of a natural disaster affecting operations",
		Severity:    "extreme",
		Duration:    6,
		Business:    business,
		Factors: map[string]StressFactor{
			"environmental_risk": {
				FactorName:    "Environmental Risk",
				FactorType:    "natural_disaster",
				BaseValue:     0.1,
				StressedValue: 0.9,
				Description:   "Major natural disaster affecting business operations",
				Severity:      "extreme",
			},
		},
	}
}
