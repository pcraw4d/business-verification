package risk

import (
	"math"
	"strings"

	"go.uber.org/zap"
)

// PriorityEngine calculates priority scores for recommendations
type PriorityEngine struct {
	logger *zap.Logger
}

// NewPriorityEngine creates a new priority engine
func NewPriorityEngine(logger *zap.Logger) *PriorityEngine {
	return &PriorityEngine{
		logger: logger,
	}
}

// CalculatePriorityScores calculates priority scores for recommendations
func (pe *PriorityEngine) CalculatePriorityScores(recommendations []RiskRecommendation, request RecommendationRequest) []RiskRecommendation {
	// TODO: RiskRecommendation doesn't have PriorityScore field
	// for i := range recommendations {
	// 	recommendations[i].PriorityScore = pe.calculatePriorityScore(recommendations[i], request)
	// }
	_ = recommendations // Suppress unused variable warning
	_ = request
	return recommendations
}

// calculatePriorityScore calculates the priority score for a single recommendation
func (pe *PriorityEngine) calculatePriorityScore(recommendation RiskRecommendation, request RecommendationRequest) float64 {
	// Base priority from risk level
	basePriority := pe.getBasePriorityScore(recommendation.Priority)

	// Risk factor impact
	riskImpact := pe.calculateRiskImpact(recommendation, request.RiskFactors)

	// Business context impact
	businessImpact := pe.calculateBusinessImpact(recommendation, request.BusinessContext)

	// Cost-benefit ratio
	costBenefit := pe.calculateCostBenefitRatio(recommendation)

	// Urgency factor
	urgency := pe.calculateUrgency(recommendation, request.BusinessContext)

	// Compliance factor
	compliance := pe.calculateComplianceFactor(recommendation)

	// Calculate weighted priority score
	priorityScore := (basePriority * 0.25) +
		(riskImpact * 0.20) +
		(businessImpact * 0.15) +
		(costBenefit * 0.15) +
		(urgency * 0.15) +
		(compliance * 0.10)

	// Normalize to 0-100 scale
	return math.Max(0, math.Min(100, priorityScore))
}

// getBasePriorityScore returns the base priority score for a risk level
func (pe *PriorityEngine) getBasePriorityScore(level RiskLevel) float64 {
	switch level {
	case RiskLevelCritical:
		return 100.0
	case RiskLevelHigh:
		return 75.0
	case RiskLevelMedium:
		return 50.0
	case RiskLevelLow:
		return 25.0
	default:
		return 50.0
	}
}

// calculateRiskImpact calculates the impact of the recommendation on risk factors
func (pe *PriorityEngine) calculateRiskImpact(recommendation RiskRecommendation, riskFactors []RiskScore) float64 {
	// Find the risk factor this recommendation addresses
	var targetFactor *RiskScore
	for _, factor := range riskFactors {
		// TODO: RiskRecommendation doesn't have Category field
		// if factor.Category == recommendation.Category {
		if false { // Stub
			targetFactor = &factor
			break
		}
	}

	if targetFactor == nil {
		return 50.0 // Default if no matching factor found
	}

	// Higher risk scores should have higher priority
	riskScore := targetFactor.Score

	// Adjust based on confidence
	confidenceAdjustment := targetFactor.Confidence

	// Calculate impact score
	impactScore := (riskScore * confidenceAdjustment) / 100.0 * 100.0

	return math.Max(0, math.Min(100, impactScore))
}

// calculateBusinessImpact calculates the business impact of the recommendation
func (pe *PriorityEngine) calculateBusinessImpact(recommendation RiskRecommendation, context map[string]interface{}) float64 {
	baseImpact := 50.0 // Default impact

	// Adjust based on business size
	if businessSize, exists := context["business_size"]; exists {
		if size, ok := businessSize.(string); ok {
			switch strings.ToLower(size) {
			case "small":
				baseImpact *= 0.7
			case "medium":
				baseImpact *= 1.0
			case "large":
				baseImpact *= 1.3
			case "enterprise":
				baseImpact *= 1.5
			}
		}
	}

	// Adjust based on industry
	if industry, exists := context["industry"]; exists {
		if industryStr, ok := industry.(string); ok {
			// TODO: RiskRecommendation doesn't have Category field
			// baseImpact *= pe.getIndustryMultiplier(industryStr, recommendation.Category)
			_ = industryStr // Suppress unused variable warning
			baseImpact *= 1.0 // Stub
		}
	}

	// Adjust based on business stage
	if stage, exists := context["business_stage"]; exists {
		if stageStr, ok := stage.(string); ok {
			// TODO: RiskRecommendation doesn't have Category field
			// baseImpact *= pe.getStageMultiplier(stageStr, recommendation.Category)
			_ = stageStr // Suppress unused variable warning
			baseImpact *= 1.0 // Stub
		}
	}

	return math.Max(0, math.Min(100, baseImpact))
}

// calculateCostBenefitRatio calculates the cost-benefit ratio
func (pe *PriorityEngine) calculateCostBenefitRatio(recommendation RiskRecommendation) float64 {
	// TODO: RiskRecommendation doesn't have CostScore or RiskReduction fields
	// if recommendation.CostScore == 0 {
	// 	return 50.0 // Default if no cost information
	// }
	// benefit := recommendation.RiskReduction * 10000
	// ratio := benefit / recommendation.CostScore
	// TODO: RiskRecommendation doesn't have CostScore or RiskReduction fields
	// The code below is commented out because ratio is not defined
	// Convert to 0-100 scale
	// Higher ratio = higher priority
	// score := math.Min(100, ratio*10) // Scale factor
	// return math.Max(0, score)
	return 50.0 // Stub - return default
}

// calculateUrgency calculates the urgency factor
func (pe *PriorityEngine) calculateUrgency(recommendation RiskRecommendation, context map[string]interface{}) float64 {
	baseUrgency := 50.0

	// Adjust based on timeline
	// TODO: RiskRecommendation doesn't have TimelineDays field
	// if recommendation.TimelineDays > 0 {
	// 	if recommendation.TimelineDays <= 30 {
	// 		baseUrgency = 90.0 // High urgency for short timelines
	// 	} else if recommendation.TimelineDays <= 60 {
	// 		baseUrgency = 70.0 // Medium-high urgency
	// 	} else if recommendation.TimelineDays <= 90 {
	// 		baseUrgency = 50.0 // Medium urgency
	// 	} else {
	// 		baseUrgency = 30.0 // Low urgency for long timelines
	// 	}
	// }
	// Using default urgency for now

	// Adjust based on business context
	if deadline, exists := context["deadline"]; exists {
		// If there's a specific deadline, increase urgency
		_ = deadline // Suppress unused variable warning
		baseUrgency *= 1.2
	}

	// Adjust based on regulatory requirements
	// TODO: RiskRecommendation doesn't have Category field
	// if recommendation.Category == RiskCategoryRegulatory {
	// 	baseUrgency *= 1.3 // Higher urgency for regulatory issues
	// }

	return math.Max(0, math.Min(100, baseUrgency))
}

// calculateComplianceFactor calculates the compliance factor
func (pe *PriorityEngine) calculateComplianceFactor(recommendation RiskRecommendation) float64 {
	baseCompliance := 50.0

	// Higher priority for regulatory and cybersecurity recommendations
	// TODO: RiskRecommendation doesn't have Category field
	// switch recommendation.Category {
	// case RiskCategoryRegulatory:
	// 	baseCompliance = 90.0
	// case RiskCategoryCybersecurity:
	// 	baseCompliance = 85.0
	// case RiskCategoryFinancial:
	// 	baseCompliance = 80.0
	// default:
	// 	baseCompliance = 50.0
	// }
	// Using default compliance for now
	
	// Adjust based on compliance notes
	// TODO: RiskRecommendation doesn't have ComplianceNotes field
	// if len(recommendation.ComplianceNotes) > 0 {
	// 	baseCompliance *= 1.1 // Slight increase for compliance-related recommendations
	// }
	
	return math.Max(0, math.Min(100, baseCompliance))
}

// getIndustryMultiplier returns a multiplier based on industry and category
func (pe *PriorityEngine) getIndustryMultiplier(industry string, category RiskCategory) float64 {
	industryLower := strings.ToLower(industry)

	// Financial services
	if strings.Contains(industryLower, "financial") || strings.Contains(industryLower, "banking") {
		switch category {
		case RiskCategoryFinancial:
			return 1.3
		case RiskCategoryRegulatory:
			return 1.4
		case RiskCategoryCybersecurity:
			return 1.2
		default:
			return 1.1
		}
	}

	// Healthcare
	if strings.Contains(industryLower, "healthcare") || strings.Contains(industryLower, "medical") {
		switch category {
		case RiskCategoryRegulatory:
			return 1.4
		case RiskCategoryCybersecurity:
			return 1.3
		case RiskCategoryReputational:
			return 1.2
		default:
			return 1.1
		}
	}

	// Technology
	if strings.Contains(industryLower, "technology") || strings.Contains(industryLower, "software") {
		switch category {
		case RiskCategoryCybersecurity:
			return 1.3
		case RiskCategoryOperational:
			return 1.2
		case RiskCategoryReputational:
			return 1.1
		default:
			return 1.0
		}
	}

	// Manufacturing
	if strings.Contains(industryLower, "manufacturing") || strings.Contains(industryLower, "production") {
		switch category {
		case RiskCategoryOperational:
			return 1.3
		case RiskCategoryRegulatory:
			return 1.2
		case RiskCategoryFinancial:
			return 1.1
		default:
			return 1.0
		}
	}

	// Retail
	if strings.Contains(industryLower, "retail") || strings.Contains(industryLower, "commerce") {
		switch category {
		case RiskCategoryReputational:
			return 1.2
		case RiskCategoryOperational:
			return 1.1
		case RiskCategoryFinancial:
			return 1.1
		default:
			return 1.0
		}
	}

	// Default multiplier
	return 1.0
}

// getStageMultiplier returns a multiplier based on business stage and category
func (pe *PriorityEngine) getStageMultiplier(stage string, category RiskCategory) float64 {
	stageLower := strings.ToLower(stage)

	switch stageLower {
	case "startup", "early":
		switch category {
		case RiskCategoryFinancial:
			return 1.3
		case RiskCategoryOperational:
			return 1.2
		case RiskCategoryRegulatory:
			return 1.1
		default:
			return 1.0
		}

	case "growth", "expanding":
		switch category {
		case RiskCategoryOperational:
			return 1.3
		case RiskCategoryFinancial:
			return 1.2
		case RiskCategoryReputational:
			return 1.1
		default:
			return 1.0
		}

	case "mature", "established":
		switch category {
		case RiskCategoryRegulatory:
			return 1.2
		case RiskCategoryCybersecurity:
			return 1.2
		case RiskCategoryReputational:
			return 1.1
		default:
			return 1.0
		}

	case "declining", "transition":
		switch category {
		case RiskCategoryFinancial:
			return 1.4
		case RiskCategoryOperational:
			return 1.3
		case RiskCategoryReputational:
			return 1.2
		default:
			return 1.1
		}

	default:
		return 1.0
	}
}

// CalculateDynamicPriority calculates dynamic priority based on changing conditions
func (pe *PriorityEngine) CalculateDynamicPriority(recommendation RiskRecommendation, currentConditions map[string]interface{}) float64 {
	// TODO: RiskRecommendation doesn't have PriorityScore field
	// basePriority := recommendation.PriorityScore
	basePriority := 50.0 // Stub - default priority

	// Adjust based on current market conditions
	if marketCondition, exists := currentConditions["market_condition"]; exists {
		if condition, ok := marketCondition.(string); ok {
			// TODO: RiskRecommendation doesn't have Category field
			// basePriority *= pe.getMarketConditionMultiplier(condition, recommendation.Category)
			_ = condition // Suppress unused variable warning
			basePriority *= 1.0 // Stub
		}
	}

	// Adjust based on recent events
	if recentEvents, exists := currentConditions["recent_events"]; exists {
		if events, ok := recentEvents.([]string); ok {
			// TODO: RiskRecommendation doesn't have Category field
			// basePriority *= pe.getEventMultiplier(events, recommendation.Category)
			_ = events // Suppress unused variable warning
			basePriority *= 1.0 // Stub
		}
	}

	// Adjust based on seasonal factors
	if season, exists := currentConditions["season"]; exists {
		if seasonStr, ok := season.(string); ok {
			// TODO: RiskRecommendation doesn't have Category field
			// basePriority *= pe.getSeasonalMultiplier(seasonStr, recommendation.Category)
			_ = seasonStr // Suppress unused variable warning
			basePriority *= 1.0 // Stub
		}
	}

	return math.Max(0, math.Min(100, basePriority))
}

// getMarketConditionMultiplier returns multiplier based on market conditions
func (pe *PriorityEngine) getMarketConditionMultiplier(condition string, category RiskCategory) float64 {
	conditionLower := strings.ToLower(condition)

	switch conditionLower {
	case "recession", "downturn":
		switch category {
		case RiskCategoryFinancial:
			return 1.4
		case RiskCategoryOperational:
			return 1.3
		case RiskCategoryReputational:
			return 1.2
		default:
			return 1.1
		}

	case "growth", "expansion":
		switch category {
		case RiskCategoryOperational:
			return 1.2
		case RiskCategoryFinancial:
			return 1.1
		default:
			return 1.0
		}

	case "volatile", "uncertain":
		switch category {
		case RiskCategoryFinancial:
			return 1.3
		case RiskCategoryRegulatory:
			return 1.2
		case RiskCategoryCybersecurity:
			return 1.1
		default:
			return 1.0
		}

	default:
		return 1.0
	}
}

// getEventMultiplier returns multiplier based on recent events
func (pe *PriorityEngine) getEventMultiplier(events []string, category RiskCategory) float64 {
	multiplier := 1.0

	for _, event := range events {
		eventLower := strings.ToLower(event)

		switch {
		case strings.Contains(eventLower, "breach") || strings.Contains(eventLower, "security"):
			if category == RiskCategoryCybersecurity {
				multiplier *= 1.5
			}

		case strings.Contains(eventLower, "regulation") || strings.Contains(eventLower, "compliance"):
			if category == RiskCategoryRegulatory {
				multiplier *= 1.4
			}

		case strings.Contains(eventLower, "financial") || strings.Contains(eventLower, "debt"):
			if category == RiskCategoryFinancial {
				multiplier *= 1.3
			}

		case strings.Contains(eventLower, "reputation") || strings.Contains(eventLower, "scandal"):
			if category == RiskCategoryReputational {
				multiplier *= 1.4
			}

		case strings.Contains(eventLower, "operational") || strings.Contains(eventLower, "process"):
			if category == RiskCategoryOperational {
				multiplier *= 1.2
			}
		}
	}

	return multiplier
}

// getSeasonalMultiplier returns multiplier based on seasonal factors
func (pe *PriorityEngine) getSeasonalMultiplier(season string, category RiskCategory) float64 {
	seasonLower := strings.ToLower(season)

	switch seasonLower {
	case "q4", "quarter4", "year-end":
		switch category {
		case RiskCategoryFinancial:
			return 1.2
		case RiskCategoryRegulatory:
			return 1.1
		default:
			return 1.0
		}

	case "q1", "quarter1", "new-year":
		switch category {
		case RiskCategoryRegulatory:
			return 1.1
		case RiskCategoryOperational:
			return 1.1
		default:
			return 1.0
		}

	default:
		return 1.0
	}
}
