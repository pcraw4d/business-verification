package risk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ThresholdMonitor represents a risk threshold monitoring system
type ThresholdMonitor interface {
	MonitorThreshold(ctx context.Context, assessment *RiskAssessment) ([]ThresholdAlert, error)
	GetThresholdConfig(category RiskCategory) (*ThresholdMonitoringConfig, error)
	UpdateThresholdConfig(category RiskCategory, config *ThresholdMonitoringConfig) error
	GetMonitoringStatus() (*MonitoringStatus, error)
	GetProviderName() string
	IsAvailable() bool
}

// ThresholdAlert represents an alert triggered by threshold monitoring
type ThresholdAlert struct {
	ID             string                 `json:"id"`
	BusinessID     string                 `json:"business_id"`
	Category       RiskCategory           `json:"category"`
	FactorID       string                 `json:"factor_id,omitempty"`
	AlertType      ThresholdAlertType     `json:"alert_type"`
	Level          RiskLevel              `json:"level"`
	Message        string                 `json:"message"`
	CurrentValue   float64                `json:"current_value"`
	ThresholdValue float64                `json:"threshold_value"`
	ExceededBy     float64                `json:"exceeded_by"`
	TriggeredAt    time.Time              `json:"triggered_at"`
	Acknowledged   bool                   `json:"acknowledged"`
	AcknowledgedAt *time.Time             `json:"acknowledged_at,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ThresholdAlertType represents the type of threshold alert
type ThresholdAlertType string

const (
	ThresholdAlertTypeExceeded    ThresholdAlertType = "exceeded"
	ThresholdAlertTypeApproaching ThresholdAlertType = "approaching"
	ThresholdAlertTypeTrending    ThresholdAlertType = "trending"
	ThresholdAlertTypeVolatility  ThresholdAlertType = "volatility"
	ThresholdAlertTypeAnomaly     ThresholdAlertType = "anomaly"
	ThresholdAlertTypeImprovement ThresholdAlertType = "improvement"
)

// ThresholdMonitoringConfig represents configuration for threshold monitoring
type ThresholdMonitoringConfig struct {
	Category             RiskCategory       `json:"category"`
	FactorID             string             `json:"factor_id,omitempty"`
	WarningThreshold     float64            `json:"warning_threshold"`
	CriticalThreshold    float64            `json:"critical_threshold"`
	ApproachingThreshold float64            `json:"approaching_threshold"`
	TrendingThreshold    float64            `json:"trending_threshold"`
	VolatilityThreshold  float64            `json:"volatility_threshold"`
	AnomalyThreshold     float64            `json:"anomaly_threshold"`
	ImprovementThreshold float64            `json:"improvement_threshold"`
	Enabled              bool               `json:"enabled"`
	AlertChannels        []string           `json:"alert_channels"`
	NotificationRules    []NotificationRule `json:"notification_rules"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}

// NotificationRule represents a rule for threshold notifications
type NotificationRule struct {
	ID         string                 `json:"id"`
	AlertType  ThresholdAlertType     `json:"alert_type"`
	Level      RiskLevel              `json:"level"`
	Channels   []string               `json:"channels"` // "email", "webhook", "sms", "dashboard"
	Recipients []string               `json:"recipients"`
	Template   string                 `json:"template"`
	Enabled    bool                   `json:"enabled"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// MonitoringStatus represents the status of threshold monitoring
type MonitoringStatus struct {
	ActiveMonitors   int                    `json:"active_monitors"`
	TotalAlerts      int                    `json:"total_alerts"`
	CriticalAlerts   int                    `json:"critical_alerts"`
	WarningAlerts    int                    `json:"warning_alerts"`
	LastAlertTime    *time.Time             `json:"last_alert_time,omitempty"`
	MonitoringHealth string                 `json:"monitoring_health"` // "healthy", "degraded", "unhealthy"
	Uptime           time.Duration          `json:"uptime"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ThresholdMonitoringManager manages multiple threshold monitors
type ThresholdMonitoringManager struct {
	logger           *observability.Logger
	monitors         map[string]ThresholdMonitor
	primaryMonitor   string
	fallbackMonitors []string
	configs          map[string]*ThresholdMonitoringConfig
	alertHistory     map[string][]ThresholdAlert
	status           *MonitoringStatus
	mutex            sync.RWMutex
}

// NewThresholdMonitoringManager creates a new threshold monitoring manager
func NewThresholdMonitoringManager(logger *observability.Logger) *ThresholdMonitoringManager {
	return &ThresholdMonitoringManager{
		logger:           logger,
		monitors:         make(map[string]ThresholdMonitor),
		primaryMonitor:   "default_monitor",
		fallbackMonitors: []string{"backup_monitor"},
		configs:          make(map[string]*ThresholdMonitoringConfig),
		alertHistory:     make(map[string][]ThresholdAlert),
		status: &MonitoringStatus{
			ActiveMonitors:   0,
			TotalAlerts:      0,
			CriticalAlerts:   0,
			WarningAlerts:    0,
			MonitoringHealth: "healthy",
			Uptime:           time.Since(time.Now()),
		},
	}
}

// RegisterMonitor registers a threshold monitor
func (m *ThresholdMonitoringManager) RegisterMonitor(name string, monitor ThresholdMonitor) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.monitors[name] = monitor
	m.status.ActiveMonitors = len(m.monitors)

	m.logger.Info("Threshold monitor registered",
		"monitor_name", name,
		"available", monitor.IsAvailable(),
	)
}

// MonitorThreshold monitors risk thresholds for an assessment
func (m *ThresholdMonitoringManager) MonitorThreshold(ctx context.Context, assessment *RiskAssessment) ([]ThresholdAlert, error) {
	requestID := ctx.Value("request_id").(string)

	m.logger.Info("Monitoring thresholds for risk assessment",
		"request_id", requestID,
		"business_id", assessment.BusinessID,
		"overall_score", assessment.OverallScore,
	)

	var allAlerts []ThresholdAlert

	// Monitor overall risk threshold
	overallAlerts, err := m.monitorOverallThreshold(ctx, assessment)
	if err != nil {
		m.logger.Error("Failed to monitor overall threshold",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to monitor overall threshold: %w", err)
	}
	allAlerts = append(allAlerts, overallAlerts...)

	// Monitor category thresholds
	categoryAlerts, err := m.monitorCategoryThresholds(ctx, assessment)
	if err != nil {
		m.logger.Error("Failed to monitor category thresholds",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to monitor category thresholds: %w", err)
	}
	allAlerts = append(allAlerts, categoryAlerts...)

	// Monitor factor thresholds
	factorAlerts, err := m.monitorFactorThresholds(ctx, assessment)
	if err != nil {
		m.logger.Error("Failed to monitor factor thresholds",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to monitor factor thresholds: %w", err)
	}
	allAlerts = append(allAlerts, factorAlerts...)

	// Monitor trending thresholds
	trendingAlerts, err := m.monitorTrendingThresholds(ctx, assessment)
	if err != nil {
		m.logger.Error("Failed to monitor trending thresholds",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to monitor trending thresholds: %w", err)
	}
	allAlerts = append(allAlerts, trendingAlerts...)

	// Update monitoring status
	m.updateMonitoringStatus(allAlerts)

	// Store alerts in history
	m.storeAlertHistory(assessment.BusinessID, allAlerts)

	m.logger.Info("Threshold monitoring completed",
		"request_id", requestID,
		"business_id", assessment.BusinessID,
		"total_alerts", len(allAlerts),
	)

	return allAlerts, nil
}

// monitorOverallThreshold monitors the overall risk threshold
func (m *ThresholdMonitoringManager) monitorOverallThreshold(ctx context.Context, assessment *RiskAssessment) ([]ThresholdAlert, error) {
	var alerts []ThresholdAlert

	// Get overall threshold config
	config, err := m.getThresholdConfig(RiskCategoryOperational)
	if err != nil {
		config = m.getDefaultThresholdConfig(RiskCategoryOperational)
	}

	// Check critical threshold
	if assessment.OverallScore >= config.CriticalThreshold {
		alert := ThresholdAlert{
			ID:             fmt.Sprintf("threshold_%s_overall_critical", assessment.ID),
			BusinessID:     assessment.BusinessID,
			Category:       RiskCategoryOperational,
			AlertType:      ThresholdAlertTypeExceeded,
			Level:          RiskLevelCritical,
			Message:        fmt.Sprintf("Critical overall risk threshold exceeded: %.1f >= %.1f", assessment.OverallScore, config.CriticalThreshold),
			CurrentValue:   assessment.OverallScore,
			ThresholdValue: config.CriticalThreshold,
			ExceededBy:     assessment.OverallScore - config.CriticalThreshold,
			TriggeredAt:    time.Now(),
			Acknowledged:   false,
		}
		alerts = append(alerts, alert)
	}

	// Check warning threshold
	if assessment.OverallScore >= config.WarningThreshold && assessment.OverallScore < config.CriticalThreshold {
		alert := ThresholdAlert{
			ID:             fmt.Sprintf("threshold_%s_overall_warning", assessment.ID),
			BusinessID:     assessment.BusinessID,
			Category:       RiskCategoryOperational,
			AlertType:      ThresholdAlertTypeExceeded,
			Level:          RiskLevelHigh,
			Message:        fmt.Sprintf("Warning overall risk threshold exceeded: %.1f >= %.1f", assessment.OverallScore, config.WarningThreshold),
			CurrentValue:   assessment.OverallScore,
			ThresholdValue: config.WarningThreshold,
			ExceededBy:     assessment.OverallScore - config.WarningThreshold,
			TriggeredAt:    time.Now(),
			Acknowledged:   false,
		}
		alerts = append(alerts, alert)
	}

	// Check approaching threshold
	if assessment.OverallScore >= config.ApproachingThreshold && assessment.OverallScore < config.WarningThreshold {
		alert := ThresholdAlert{
			ID:             fmt.Sprintf("threshold_%s_overall_approaching", assessment.ID),
			BusinessID:     assessment.BusinessID,
			Category:       RiskCategoryOperational,
			AlertType:      ThresholdAlertTypeApproaching,
			Level:          RiskLevelMedium,
			Message:        fmt.Sprintf("Overall risk approaching threshold: %.1f >= %.1f", assessment.OverallScore, config.ApproachingThreshold),
			CurrentValue:   assessment.OverallScore,
			ThresholdValue: config.ApproachingThreshold,
			ExceededBy:     assessment.OverallScore - config.ApproachingThreshold,
			TriggeredAt:    time.Now(),
			Acknowledged:   false,
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// monitorCategoryThresholds monitors category-specific thresholds
func (m *ThresholdMonitoringManager) monitorCategoryThresholds(ctx context.Context, assessment *RiskAssessment) ([]ThresholdAlert, error) {
	var alerts []ThresholdAlert

	for category, score := range assessment.CategoryScores {
		config, err := m.getThresholdConfig(category)
		if err != nil {
			config = m.getDefaultThresholdConfig(category)
		}

		// Check critical threshold
		if score.Score >= config.CriticalThreshold {
			alert := ThresholdAlert{
				ID:             fmt.Sprintf("threshold_%s_%s_critical", assessment.ID, category),
				BusinessID:     assessment.BusinessID,
				Category:       category,
				AlertType:      ThresholdAlertTypeExceeded,
				Level:          RiskLevelCritical,
				Message:        fmt.Sprintf("Critical %s risk threshold exceeded: %.1f >= %.1f", category, score.Score, config.CriticalThreshold),
				CurrentValue:   score.Score,
				ThresholdValue: config.CriticalThreshold,
				ExceededBy:     score.Score - config.CriticalThreshold,
				TriggeredAt:    time.Now(),
				Acknowledged:   false,
			}
			alerts = append(alerts, alert)
		}

		// Check warning threshold
		if score.Score >= config.WarningThreshold && score.Score < config.CriticalThreshold {
			alert := ThresholdAlert{
				ID:             fmt.Sprintf("threshold_%s_%s_warning", assessment.ID, category),
				BusinessID:     assessment.BusinessID,
				Category:       category,
				AlertType:      ThresholdAlertTypeExceeded,
				Level:          RiskLevelHigh,
				Message:        fmt.Sprintf("Warning %s risk threshold exceeded: %.1f >= %.1f", category, score.Score, config.WarningThreshold),
				CurrentValue:   score.Score,
				ThresholdValue: config.WarningThreshold,
				ExceededBy:     score.Score - config.WarningThreshold,
				TriggeredAt:    time.Now(),
				Acknowledged:   false,
			}
			alerts = append(alerts, alert)
		}

		// Check approaching threshold
		if score.Score >= config.ApproachingThreshold && score.Score < config.WarningThreshold {
			alert := ThresholdAlert{
				ID:             fmt.Sprintf("threshold_%s_%s_approaching", assessment.ID, category),
				BusinessID:     assessment.BusinessID,
				Category:       category,
				AlertType:      ThresholdAlertTypeApproaching,
				Level:          RiskLevelMedium,
				Message:        fmt.Sprintf("%s risk approaching threshold: %.1f >= %.1f", category, score.Score, config.ApproachingThreshold),
				CurrentValue:   score.Score,
				ThresholdValue: config.ApproachingThreshold,
				ExceededBy:     score.Score - config.ApproachingThreshold,
				TriggeredAt:    time.Now(),
				Acknowledged:   false,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// monitorFactorThresholds monitors factor-specific thresholds
func (m *ThresholdMonitoringManager) monitorFactorThresholds(ctx context.Context, assessment *RiskAssessment) ([]ThresholdAlert, error) {
	var alerts []ThresholdAlert

	for _, factorScore := range assessment.FactorScores {
		config, err := m.getThresholdConfig(factorScore.Category)
		if err != nil {
			config = m.getDefaultThresholdConfig(factorScore.Category)
		}

		// Check critical threshold
		if factorScore.Score >= config.CriticalThreshold {
			alert := ThresholdAlert{
				ID:             fmt.Sprintf("threshold_%s_%s_critical", assessment.ID, factorScore.FactorID),
				BusinessID:     assessment.BusinessID,
				Category:       factorScore.Category,
				FactorID:       factorScore.FactorID,
				AlertType:      ThresholdAlertTypeExceeded,
				Level:          RiskLevelCritical,
				Message:        fmt.Sprintf("Critical %s risk threshold exceeded: %.1f >= %.1f", factorScore.FactorName, factorScore.Score, config.CriticalThreshold),
				CurrentValue:   factorScore.Score,
				ThresholdValue: config.CriticalThreshold,
				ExceededBy:     factorScore.Score - config.CriticalThreshold,
				TriggeredAt:    time.Now(),
				Acknowledged:   false,
			}
			alerts = append(alerts, alert)
		}

		// Check warning threshold
		if factorScore.Score >= config.WarningThreshold && factorScore.Score < config.CriticalThreshold {
			alert := ThresholdAlert{
				ID:             fmt.Sprintf("threshold_%s_%s_warning", assessment.ID, factorScore.FactorID),
				BusinessID:     assessment.BusinessID,
				Category:       factorScore.Category,
				FactorID:       factorScore.FactorID,
				AlertType:      ThresholdAlertTypeExceeded,
				Level:          RiskLevelHigh,
				Message:        fmt.Sprintf("Warning %s risk threshold exceeded: %.1f >= %.1f", factorScore.FactorName, factorScore.Score, config.WarningThreshold),
				CurrentValue:   factorScore.Score,
				ThresholdValue: config.WarningThreshold,
				ExceededBy:     factorScore.Score - config.WarningThreshold,
				TriggeredAt:    time.Now(),
				Acknowledged:   false,
			}
			alerts = append(alerts, alert)
		}

		// Check approaching threshold
		if factorScore.Score >= config.ApproachingThreshold && factorScore.Score < config.WarningThreshold {
			alert := ThresholdAlert{
				ID:             fmt.Sprintf("threshold_%s_%s_approaching", assessment.ID, factorScore.FactorID),
				BusinessID:     assessment.BusinessID,
				Category:       factorScore.Category,
				FactorID:       factorScore.FactorID,
				AlertType:      ThresholdAlertTypeApproaching,
				Level:          RiskLevelMedium,
				Message:        fmt.Sprintf("%s risk approaching threshold: %.1f >= %.1f", factorScore.FactorName, factorScore.Score, config.ApproachingThreshold),
				CurrentValue:   factorScore.Score,
				ThresholdValue: config.ApproachingThreshold,
				ExceededBy:     factorScore.Score - config.ApproachingThreshold,
				TriggeredAt:    time.Now(),
				Acknowledged:   false,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// monitorTrendingThresholds monitors trending thresholds
func (m *ThresholdMonitoringManager) monitorTrendingThresholds(ctx context.Context, assessment *RiskAssessment) ([]ThresholdAlert, error) {
	var alerts []ThresholdAlert

	// This would typically analyze historical data for trends
	// For now, we'll create alerts based on current assessment patterns

	// Check for rapid risk increase
	if assessment.OverallScore > 60 {
		highRiskFactors := 0
		for _, factor := range assessment.FactorScores {
			if factor.Score > 70 {
				highRiskFactors++
			}
		}

		if highRiskFactors >= 3 {
			alert := ThresholdAlert{
				ID:             fmt.Sprintf("threshold_%s_trending", assessment.ID),
				BusinessID:     assessment.BusinessID,
				Category:       RiskCategoryOperational,
				AlertType:      ThresholdAlertTypeTrending,
				Level:          RiskLevelHigh,
				Message:        fmt.Sprintf("Risk trending upward: %d factors above 70", highRiskFactors),
				CurrentValue:   assessment.OverallScore,
				ThresholdValue: 60.0,
				ExceededBy:     assessment.OverallScore - 60.0,
				TriggeredAt:    time.Now(),
				Acknowledged:   false,
			}
			alerts = append(alerts, alert)
		}
	}

	// Check for volatility (multiple factors with high variance)
	volatilityCount := 0
	for _, factor := range assessment.FactorScores {
		if factor.Score > 50 && factor.Score < 80 {
			volatilityCount++
		}
	}

	if volatilityCount >= 4 {
		alert := ThresholdAlert{
			ID:             fmt.Sprintf("threshold_%s_volatility", assessment.ID),
			BusinessID:     assessment.BusinessID,
			Category:       RiskCategoryOperational,
			AlertType:      ThresholdAlertTypeVolatility,
			Level:          RiskLevelMedium,
			Message:        fmt.Sprintf("High risk volatility detected: %d factors in mid-range", volatilityCount),
			CurrentValue:   assessment.OverallScore,
			ThresholdValue: 50.0,
			ExceededBy:     assessment.OverallScore - 50.0,
			TriggeredAt:    time.Now(),
			Acknowledged:   false,
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// GetThresholdConfig retrieves threshold configuration for a category
func (m *ThresholdMonitoringManager) GetThresholdConfig(category RiskCategory) (*ThresholdMonitoringConfig, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	config, exists := m.configs[string(category)]
	if !exists {
		return m.getDefaultThresholdConfig(category), nil
	}

	return config, nil
}

// UpdateThresholdConfig updates threshold configuration for a category
func (m *ThresholdMonitoringManager) UpdateThresholdConfig(category RiskCategory, config *ThresholdMonitoringConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	config.Category = category
	config.UpdatedAt = time.Now()

	m.configs[string(category)] = config

	m.logger.Info("Threshold config updated",
		"category", category,
		"warning_threshold", config.WarningThreshold,
		"critical_threshold", config.CriticalThreshold,
	)

	return nil
}

// GetMonitoringStatus retrieves the current monitoring status
func (m *ThresholdMonitoringManager) GetMonitoringStatus() (*MonitoringStatus, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.status, nil
}

// Helper methods
func (m *ThresholdMonitoringManager) getThresholdConfig(category RiskCategory) (*ThresholdMonitoringConfig, error) {
	return m.GetThresholdConfig(category)
}

func (m *ThresholdMonitoringManager) getDefaultThresholdConfig(category RiskCategory) *ThresholdMonitoringConfig {
	now := time.Now()

	switch category {
	case RiskCategoryFinancial:
		return &ThresholdMonitoringConfig{
			Category:             RiskCategoryFinancial,
			WarningThreshold:     70.0,
			CriticalThreshold:    85.0,
			ApproachingThreshold: 60.0,
			TrendingThreshold:    65.0,
			VolatilityThreshold:  50.0,
			AnomalyThreshold:     90.0,
			ImprovementThreshold: 30.0,
			Enabled:              true,
			AlertChannels:        []string{"email", "dashboard"},
			CreatedAt:            now,
			UpdatedAt:            now,
		}
	case RiskCategoryOperational:
		return &ThresholdMonitoringConfig{
			Category:             RiskCategoryOperational,
			WarningThreshold:     65.0,
			CriticalThreshold:    80.0,
			ApproachingThreshold: 55.0,
			TrendingThreshold:    60.0,
			VolatilityThreshold:  45.0,
			AnomalyThreshold:     85.0,
			ImprovementThreshold: 25.0,
			Enabled:              true,
			AlertChannels:        []string{"email", "dashboard"},
			CreatedAt:            now,
			UpdatedAt:            now,
		}
	case RiskCategoryRegulatory:
		return &ThresholdMonitoringConfig{
			Category:             RiskCategoryRegulatory,
			WarningThreshold:     80.0,
			CriticalThreshold:    90.0,
			ApproachingThreshold: 70.0,
			TrendingThreshold:    75.0,
			VolatilityThreshold:  60.0,
			AnomalyThreshold:     95.0,
			ImprovementThreshold: 40.0,
			Enabled:              true,
			AlertChannels:        []string{"email", "dashboard", "webhook"},
			CreatedAt:            now,
			UpdatedAt:            now,
		}
	case RiskCategoryReputational:
		return &ThresholdMonitoringConfig{
			Category:             RiskCategoryReputational,
			WarningThreshold:     75.0,
			CriticalThreshold:    85.0,
			ApproachingThreshold: 65.0,
			TrendingThreshold:    70.0,
			VolatilityThreshold:  55.0,
			AnomalyThreshold:     90.0,
			ImprovementThreshold: 35.0,
			Enabled:              true,
			AlertChannels:        []string{"email", "dashboard"},
			CreatedAt:            now,
			UpdatedAt:            now,
		}
	case RiskCategoryCybersecurity:
		return &ThresholdMonitoringConfig{
			Category:             RiskCategoryCybersecurity,
			WarningThreshold:     85.0,
			CriticalThreshold:    95.0,
			ApproachingThreshold: 75.0,
			TrendingThreshold:    80.0,
			VolatilityThreshold:  70.0,
			AnomalyThreshold:     98.0,
			ImprovementThreshold: 50.0,
			Enabled:              true,
			AlertChannels:        []string{"email", "dashboard", "webhook", "sms"},
			CreatedAt:            now,
			UpdatedAt:            now,
		}
	default:
		return &ThresholdMonitoringConfig{
			Category:             category,
			WarningThreshold:     75.0,
			CriticalThreshold:    85.0,
			ApproachingThreshold: 65.0,
			TrendingThreshold:    70.0,
			VolatilityThreshold:  55.0,
			AnomalyThreshold:     90.0,
			ImprovementThreshold: 35.0,
			Enabled:              true,
			AlertChannels:        []string{"email", "dashboard"},
			CreatedAt:            now,
			UpdatedAt:            now,
		}
	}
}

func (m *ThresholdMonitoringManager) updateMonitoringStatus(alerts []ThresholdAlert) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.status.TotalAlerts += len(alerts)

	criticalCount := 0
	warningCount := 0

	for _, alert := range alerts {
		switch alert.Level {
		case RiskLevelCritical:
			criticalCount++
		case RiskLevelHigh:
			warningCount++
		}
	}

	m.status.CriticalAlerts += criticalCount
	m.status.WarningAlerts += warningCount

	if len(alerts) > 0 {
		now := time.Now()
		m.status.LastAlertTime = &now
	}

	// Update health status based on alert patterns
	if m.status.CriticalAlerts > 10 {
		m.status.MonitoringHealth = "unhealthy"
	} else if m.status.CriticalAlerts > 5 {
		m.status.MonitoringHealth = "degraded"
	} else {
		m.status.MonitoringHealth = "healthy"
	}
}

func (m *ThresholdMonitoringManager) storeAlertHistory(businessID string, alerts []ThresholdAlert) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.alertHistory[businessID] == nil {
		m.alertHistory[businessID] = []ThresholdAlert{}
	}

	m.alertHistory[businessID] = append(m.alertHistory[businessID], alerts...)

	// Keep only last 100 alerts per business
	if len(m.alertHistory[businessID]) > 100 {
		m.alertHistory[businessID] = m.alertHistory[businessID][len(m.alertHistory[businessID])-100:]
	}
}

// RealThresholdMonitor represents a real threshold monitor with API integration
type RealThresholdMonitor struct {
	name          string
	apiKey        string
	baseURL       string
	timeout       time.Duration
	retryAttempts int
	available     bool
	logger        *observability.Logger
	httpClient    *http.Client
}

// NewRealThresholdMonitor creates a new real threshold monitor
func NewRealThresholdMonitor(name, apiKey, baseURL string, logger *observability.Logger) *RealThresholdMonitor {
	return &RealThresholdMonitor{
		name:          name,
		apiKey:        apiKey,
		baseURL:       baseURL,
		timeout:       30 * time.Second,
		retryAttempts: 3,
		available:     true,
		logger:        logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// MonitorThreshold implements ThresholdMonitor interface for real monitors
func (m *RealThresholdMonitor) MonitorThreshold(ctx context.Context, assessment *RiskAssessment) ([]ThresholdAlert, error) {
	requestID := ctx.Value("request_id").(string)

	m.logger.Info("Monitoring thresholds with real monitor",
		"request_id", requestID,
		"business_id", assessment.BusinessID,
		"provider", m.name,
	)

	url := fmt.Sprintf("%s/monitor/thresholds", m.baseURL)

	// Create request body
	requestBody := map[string]interface{}{
		"business_id": assessment.BusinessID,
		"assessment":  assessment,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	for attempt := 0; attempt < m.retryAttempts; attempt++ {
		resp, err = m.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if attempt < m.retryAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if err != nil {
		m.logger.Error("Failed to monitor thresholds with real monitor",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
			"provider", m.name,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.logger.Error("Real monitor returned error status for threshold monitoring",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
			"provider", m.name,
			"status_code", resp.StatusCode,
		)
		return nil, fmt.Errorf("monitor returned status %d", resp.StatusCode)
	}

	var thresholdAlerts []ThresholdAlert
	if err := json.NewDecoder(resp.Body).Decode(&thresholdAlerts); err != nil {
		m.logger.Error("Failed to decode threshold alerts",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
			"provider", m.name,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	m.logger.Info("Successfully monitored thresholds with real monitor",
		"request_id", requestID,
		"business_id", assessment.BusinessID,
		"provider", m.name,
		"alert_count", len(thresholdAlerts),
	)

	return thresholdAlerts, nil
}

// Implement other ThresholdMonitor methods for real monitor
func (m *RealThresholdMonitor) GetThresholdConfig(category RiskCategory) (*ThresholdMonitoringConfig, error) {
	// Similar implementation for getting threshold config
	return &ThresholdMonitoringConfig{
		Category:             category,
		WarningThreshold:     75.0,
		CriticalThreshold:    85.0,
		ApproachingThreshold: 65.0,
		Enabled:              true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}, nil
}

func (m *RealThresholdMonitor) UpdateThresholdConfig(category RiskCategory, config *ThresholdMonitoringConfig) error {
	// Similar implementation for updating threshold config
	return nil
}

func (m *RealThresholdMonitor) GetMonitoringStatus() (*MonitoringStatus, error) {
	// Similar implementation for getting monitoring status
	return &MonitoringStatus{
		ActiveMonitors:   1,
		TotalAlerts:      0,
		CriticalAlerts:   0,
		WarningAlerts:    0,
		MonitoringHealth: "healthy",
		Uptime:           time.Since(time.Now()),
	}, nil
}

func (m *RealThresholdMonitor) GetProviderName() string {
	return m.name
}

func (m *RealThresholdMonitor) IsAvailable() bool {
	return m.available
}

func (m *RealThresholdMonitor) SetAvailable(available bool) {
	m.available = available
}

// Specialized monitor types
type FinancialThresholdMonitor struct {
	*RealThresholdMonitor
}

func NewFinancialThresholdMonitor(apiKey, baseURL string, logger *observability.Logger) *FinancialThresholdMonitor {
	return &FinancialThresholdMonitor{
		RealThresholdMonitor: NewRealThresholdMonitor("financial_threshold_monitor", apiKey, baseURL, logger),
	}
}

type OperationalThresholdMonitor struct {
	*RealThresholdMonitor
}

func NewOperationalThresholdMonitor(apiKey, baseURL string, logger *observability.Logger) *OperationalThresholdMonitor {
	return &OperationalThresholdMonitor{
		RealThresholdMonitor: NewRealThresholdMonitor("operational_threshold_monitor", apiKey, baseURL, logger),
	}
}

type RegulatoryThresholdMonitor struct {
	*RealThresholdMonitor
}

func NewRegulatoryThresholdMonitor(apiKey, baseURL string, logger *observability.Logger) *RegulatoryThresholdMonitor {
	return &RegulatoryThresholdMonitor{
		RealThresholdMonitor: NewRealThresholdMonitor("regulatory_threshold_monitor", apiKey, baseURL, logger),
	}
}

type ReputationalThresholdMonitor struct {
	*RealThresholdMonitor
}

func NewReputationalThresholdMonitor(apiKey, baseURL string, logger *observability.Logger) *ReputationalThresholdMonitor {
	return &ReputationalThresholdMonitor{
		RealThresholdMonitor: NewRealThresholdMonitor("reputational_threshold_monitor", apiKey, baseURL, logger),
	}
}

type CybersecurityThresholdMonitor struct {
	*RealThresholdMonitor
}

func NewCybersecurityThresholdMonitor(apiKey, baseURL string, logger *observability.Logger) *CybersecurityThresholdMonitor {
	return &CybersecurityThresholdMonitor{
		RealThresholdMonitor: NewRealThresholdMonitor("cybersecurity_threshold_monitor", apiKey, baseURL, logger),
	}
}
