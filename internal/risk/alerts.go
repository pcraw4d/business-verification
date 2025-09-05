package risk

import (
	"context"
	"fmt"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// AlertService provides comprehensive risk alert functionality
type AlertService struct {
	logger           *observability.Logger
	thresholdManager *ThresholdManager
}

// NewAlertService creates a new alert service
func NewAlertService(logger *observability.Logger, thresholdManager *ThresholdManager) *AlertService {
	return &AlertService{
		logger:           logger,
		thresholdManager: thresholdManager,
	}
}

// AlertRule represents a rule for generating alerts
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    RiskCategory           `json:"category"`
	FactorID    string                 `json:"factor_id,omitempty"`
	Condition   AlertCondition         `json:"condition"`
	Threshold   float64                `json:"threshold"`
	Level       RiskLevel              `json:"level"`
	Message     string                 `json:"message"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AlertCondition represents the condition for triggering an alert
type AlertCondition string

const (
	AlertConditionGreaterThan  AlertCondition = "greater_than"
	AlertConditionLessThan     AlertCondition = "less_than"
	AlertConditionEquals       AlertCondition = "equals"
	AlertConditionNotEquals    AlertCondition = "not_equals"
	AlertConditionIncreasesBy  AlertCondition = "increases_by"
	AlertConditionDecreasesBy  AlertCondition = "decreases_by"
	AlertConditionCrossesAbove AlertCondition = "crosses_above"
	AlertConditionCrossesBelow AlertCondition = "crosses_below"
)

// AlertNotification represents a notification for an alert
type AlertNotification struct {
	ID         string     `json:"id"`
	AlertID    string     `json:"alert_id"`
	Type       string     `json:"type"` // "email", "webhook", "sms", "dashboard"
	Recipient  string     `json:"recipient"`
	Message    string     `json:"message"`
	Status     string     `json:"status"` // "pending", "sent", "failed"
	SentAt     *time.Time `json:"sent_at,omitempty"`
	RetryCount int        `json:"retry_count"`
	CreatedAt  time.Time  `json:"created_at"`
}

// GenerateAlerts generates comprehensive alerts based on assessment results
func (s *AlertService) GenerateAlerts(ctx context.Context, assessment *RiskAssessment) ([]RiskAlert, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Generating alerts for risk assessment", map[string]interface{}{
		"request_id":    requestID,
		"business_id":   assessment.BusinessID,
		"overall_score": assessment.OverallScore,
		"overall_level": assessment.OverallLevel,
	})

	var alerts []RiskAlert

	// Generate factor-specific alerts
	factorAlerts, err := s.generateFactorAlerts(ctx, assessment)
	if err != nil {
		s.logger.Error("Failed to generate factor alerts", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to generate factor alerts: %w", err)
	}
	alerts = append(alerts, factorAlerts...)

	// Generate category-specific alerts
	categoryAlerts, err := s.generateCategoryAlerts(ctx, assessment)
	if err != nil {
		s.logger.Error("Failed to generate category alerts", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to generate category alerts: %w", err)
	}
	alerts = append(alerts, categoryAlerts...)

	// Generate overall risk alerts
	overallAlerts, err := s.generateOverallAlerts(ctx, assessment)
	if err != nil {
		s.logger.Error("Failed to generate overall alerts", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to generate overall alerts: %w", err)
	}
	alerts = append(alerts, overallAlerts...)

	// Generate trend-based alerts
	trendAlerts, err := s.generateTrendAlerts(ctx, assessment)
	if err != nil {
		s.logger.Error("Failed to generate trend alerts", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to generate trend alerts: %w", err)
	}
	alerts = append(alerts, trendAlerts...)

	s.logger.Info("Alert generation completed", map[string]interface{}{
		"request_id":   requestID,
		"total_alerts": len(alerts),
		"business_id":  assessment.BusinessID,
	})

	return alerts, nil
}

// generateFactorAlerts generates alerts for individual risk factors
func (s *AlertService) generateFactorAlerts(ctx context.Context, assessment *RiskAssessment) ([]RiskAlert, error) {
	var alerts []RiskAlert

	for _, factorScore := range assessment.FactorScores {
		// Check if factor score exceeds threshold
		if factorScore.Level == RiskLevelHigh || factorScore.Level == RiskLevelCritical {
			alert := RiskAlert{
				ID:           fmt.Sprintf("alert_%s_%s", assessment.ID, factorScore.FactorID),
				BusinessID:   assessment.BusinessID,
				RiskFactor:   factorScore.FactorID,
				Level:        factorScore.Level,
				Message:      fmt.Sprintf("High risk detected in %s: %.1f (Level: %s)", factorScore.FactorName, factorScore.Score, factorScore.Level),
				Score:        factorScore.Score,
				Threshold:    s.getThresholdForFactor(factorScore.FactorID, factorScore.Category),
				TriggeredAt:  time.Now(),
				Acknowledged: false,
			}
			alerts = append(alerts, alert)
		}

		// Check for significant score changes
		if factorScore.Confidence > 0.8 && factorScore.Score > 60 {
			alert := RiskAlert{
				ID:           fmt.Sprintf("alert_%s_%s_change", assessment.ID, factorScore.FactorID),
				BusinessID:   assessment.BusinessID,
				RiskFactor:   factorScore.FactorID,
				Level:        RiskLevelMedium,
				Message:      fmt.Sprintf("Significant risk change in %s: %.1f", factorScore.FactorName, factorScore.Score),
				Score:        factorScore.Score,
				Threshold:    60.0,
				TriggeredAt:  time.Now(),
				Acknowledged: false,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// generateCategoryAlerts generates alerts for risk categories
func (s *AlertService) generateCategoryAlerts(ctx context.Context, assessment *RiskAssessment) ([]RiskAlert, error) {
	var alerts []RiskAlert

	for category, score := range assessment.CategoryScores {
		// Check for high-risk categories
		if score.Level == RiskLevelHigh || score.Level == RiskLevelCritical {
			alert := RiskAlert{
				ID:           fmt.Sprintf("alert_%s_%s", assessment.ID, category),
				BusinessID:   assessment.BusinessID,
				RiskFactor:   string(category),
				Level:        score.Level,
				Message:      fmt.Sprintf("High risk in %s category: %.1f (Level: %s)", category, score.Score, score.Level),
				Score:        score.Score,
				Threshold:    s.getThresholdForCategory(category),
				TriggeredAt:  time.Now(),
				Acknowledged: false,
			}
			alerts = append(alerts, alert)
		}

		// Check for category-specific thresholds
		if score.Score > s.getCategoryThreshold(category) {
			alert := RiskAlert{
				ID:           fmt.Sprintf("alert_%s_%s_threshold", assessment.ID, category),
				BusinessID:   assessment.BusinessID,
				RiskFactor:   string(category),
				Level:        RiskLevelMedium,
				Message:      fmt.Sprintf("%s category exceeds threshold: %.1f", category, score.Score),
				Score:        score.Score,
				Threshold:    s.getCategoryThreshold(category),
				TriggeredAt:  time.Now(),
				Acknowledged: false,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// generateOverallAlerts generates alerts for overall risk assessment
func (s *AlertService) generateOverallAlerts(ctx context.Context, assessment *RiskAssessment) ([]RiskAlert, error) {
	var alerts []RiskAlert

	// Check for overall high risk
	if assessment.OverallLevel == RiskLevelHigh || assessment.OverallLevel == RiskLevelCritical {
		alert := RiskAlert{
			ID:           fmt.Sprintf("alert_%s_overall", assessment.ID),
			BusinessID:   assessment.BusinessID,
			RiskFactor:   "overall_risk",
			Level:        assessment.OverallLevel,
			Message:      fmt.Sprintf("Overall risk level is %s: %.1f", assessment.OverallLevel, assessment.OverallScore),
			Score:        assessment.OverallScore,
			Threshold:    75.0,
			TriggeredAt:  time.Now(),
			Acknowledged: false,
		}
		alerts = append(alerts, alert)
	}

	// Check for overall score thresholds
	if assessment.OverallScore > 80 {
		alert := RiskAlert{
			ID:           fmt.Sprintf("alert_%s_overall_critical", assessment.ID),
			BusinessID:   assessment.BusinessID,
			RiskFactor:   "overall_risk_critical",
			Level:        RiskLevelCritical,
			Message:      fmt.Sprintf("Critical overall risk score: %.1f", assessment.OverallScore),
			Score:        assessment.OverallScore,
			Threshold:    80.0,
			TriggeredAt:  time.Now(),
			Acknowledged: false,
		}
		alerts = append(alerts, alert)
	}

	// Check for rapid risk increase
	if assessment.OverallScore > 60 && len(assessment.FactorScores) > 0 {
		highRiskFactors := 0
		for _, factor := range assessment.FactorScores {
			if factor.Score > 70 {
				highRiskFactors++
			}
		}

		if highRiskFactors >= 3 {
			alert := RiskAlert{
				ID:           fmt.Sprintf("alert_%s_multiple_high_risk", assessment.ID),
				BusinessID:   assessment.BusinessID,
				RiskFactor:   "multiple_high_risk_factors",
				Level:        RiskLevelHigh,
				Message:      fmt.Sprintf("Multiple high-risk factors detected: %d factors above 70", highRiskFactors),
				Score:        assessment.OverallScore,
				Threshold:    60.0,
				TriggeredAt:  time.Now(),
				Acknowledged: false,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// generateTrendAlerts generates alerts based on risk trends
func (s *AlertService) generateTrendAlerts(ctx context.Context, assessment *RiskAssessment) ([]RiskAlert, error) {
	var alerts []RiskAlert

	// This would typically analyze historical data
	// For now, we'll create alerts based on current assessment patterns

	// Check for increasing risk patterns
	if assessment.OverallScore > 50 {
		// Simulate trend analysis
		alert := RiskAlert{
			ID:           fmt.Sprintf("alert_%s_trend", assessment.ID),
			BusinessID:   assessment.BusinessID,
			RiskFactor:   "risk_trend",
			Level:        RiskLevelMedium,
			Message:      fmt.Sprintf("Risk trend analysis: Score %.1f indicates potential risk escalation", assessment.OverallScore),
			Score:        assessment.OverallScore,
			Threshold:    50.0,
			TriggeredAt:  time.Now(),
			Acknowledged: false,
		}
		alerts = append(alerts, alert)
	}

	// Check for category imbalances
	categoryCounts := make(map[RiskCategory]int)
	for _, factor := range assessment.FactorScores {
		categoryCounts[factor.Category]++
	}

	for category, count := range categoryCounts {
		if count >= 3 && assessment.CategoryScores[category].Score > 60 {
			alert := RiskAlert{
				ID:           fmt.Sprintf("alert_%s_%s_concentration", assessment.ID, category),
				BusinessID:   assessment.BusinessID,
				RiskFactor:   string(category),
				Level:        RiskLevelMedium,
				Message:      fmt.Sprintf("Risk concentration in %s category: %d factors with high scores", category, count),
				Score:        assessment.CategoryScores[category].Score,
				Threshold:    60.0,
				TriggeredAt:  time.Now(),
				Acknowledged: false,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// getThresholdForFactor returns the threshold for a specific risk factor
func (s *AlertService) getThresholdForFactor(factorID string, category RiskCategory) float64 {
	// Default thresholds based on category
	switch category {
	case RiskCategoryFinancial:
		return 70.0
	case RiskCategoryOperational:
		return 65.0
	case RiskCategoryRegulatory:
		return 80.0
	case RiskCategoryReputational:
		return 75.0
	case RiskCategoryCybersecurity:
		return 85.0
	default:
		return 75.0
	}
}

// getThresholdForCategory returns the threshold for a risk category
func (s *AlertService) getThresholdForCategory(category RiskCategory) float64 {
	switch category {
	case RiskCategoryFinancial:
		return 70.0
	case RiskCategoryOperational:
		return 65.0
	case RiskCategoryRegulatory:
		return 80.0
	case RiskCategoryReputational:
		return 75.0
	case RiskCategoryCybersecurity:
		return 85.0
	default:
		return 75.0
	}
}

// getCategoryThreshold returns the category-specific threshold
func (s *AlertService) getCategoryThreshold(category RiskCategory) float64 {
	return s.getThresholdForCategory(category)
}

// CreateAlertRule creates a new alert rule
func (s *AlertService) CreateAlertRule(rule *AlertRule) error {
	rule.ID = fmt.Sprintf("rule_%d", time.Now().UnixNano())
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	// In a real implementation, this would be stored in the database
	return nil
}

// GetAlertRules returns all alert rules
func (s *AlertService) GetAlertRules() ([]*AlertRule, error) {
	// In a real implementation, this would retrieve from database
	// For now, return default rules
	return s.getDefaultAlertRules(), nil
}

// GetAlerts retrieves alerts for a business
func (s *AlertService) GetAlerts(ctx context.Context, businessID string) ([]RiskAlert, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving alerts for business", map[string]interface{}{
		"request_id":  requestID,
		"business_id": businessID,
	})

	// In a real implementation, this would query the database
	// For now, return mock alerts
	alerts := []RiskAlert{
		{
			ID:           fmt.Sprintf("alert_%s_1", businessID),
			BusinessID:   businessID,
			RiskFactor:   "financial_stability",
			Level:        RiskLevelCritical,
			Message:      "Critical financial stability risk detected",
			Score:        85.0,
			Threshold:    80.0,
			TriggeredAt:  time.Now().Add(-2 * time.Hour),
			Acknowledged: false,
		},
		{
			ID:           fmt.Sprintf("alert_%s_2", businessID),
			BusinessID:   businessID,
			RiskFactor:   "operational_efficiency",
			Level:        RiskLevelHigh,
			Message:      "High operational efficiency risk detected",
			Score:        75.0,
			Threshold:    70.0,
			TriggeredAt:  time.Now().Add(-1 * time.Hour),
			Acknowledged: false,
		},
	}

	s.logger.Info("Retrieved alerts for business", map[string]interface{}{
		"request_id":  requestID,
		"business_id": businessID,
		"alert_count": len(alerts),
	})

	return alerts, nil
}

// getDefaultAlertRules returns default alert rules
func (s *AlertService) getDefaultAlertRules() []*AlertRule {
	return []*AlertRule{
		{
			ID:          "rule_overall_critical",
			Name:        "Overall Critical Risk",
			Description: "Alert when overall risk score exceeds 80",
			Category:    RiskCategoryOperational,
			Condition:   AlertConditionGreaterThan,
			Threshold:   80.0,
			Level:       RiskLevelCritical,
			Message:     "Critical overall risk detected",
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rule_financial_high",
			Name:        "High Financial Risk",
			Description: "Alert when financial risk score exceeds 70",
			Category:    RiskCategoryFinancial,
			Condition:   AlertConditionGreaterThan,
			Threshold:   70.0,
			Level:       RiskLevelHigh,
			Message:     "High financial risk detected",
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rule_regulatory_critical",
			Name:        "Critical Regulatory Risk",
			Description: "Alert when regulatory risk score exceeds 85",
			Category:    RiskCategoryRegulatory,
			Condition:   AlertConditionGreaterThan,
			Threshold:   85.0,
			Level:       RiskLevelCritical,
			Message:     "Critical regulatory risk detected",
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
}
