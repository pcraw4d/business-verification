package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ClassificationAlertManager provides advanced alerting for classification system monitoring
type ClassificationAlertManager struct {
	baseAlertManager    *AlertManager
	logger              *zap.Logger
	classificationRules map[string]*ClassificationAlertRule
	mu                  sync.RWMutex
	ctx                 context.Context
	cancel              context.CancelFunc
	started             bool
}

// ClassificationAlertRule represents a classification-specific alert rule
type ClassificationAlertRule struct {
	ID                   string                   `json:"id"`
	Name                 string                   `json:"name"`
	Description          string                   `json:"description"`
	Category             AlertCategory            `json:"category"`
	MetricType           ClassificationMetricType `json:"metric_type"`
	Query                string                   `json:"query"`
	Condition            string                   `json:"condition"`
	Threshold            float64                  `json:"threshold"`
	Severity             AlertSeverity            `json:"severity"`
	Duration             time.Duration            `json:"duration"`
	Labels               map[string]string        `json:"labels"`
	Annotations          map[string]string        `json:"annotations"`
	NotificationChannels []string                 `json:"notification_channels"`
	EscalationPolicy     string                   `json:"escalation_policy"`
	Enabled              bool                     `json:"enabled"`
	CreatedAt            time.Time                `json:"created_at"`
	UpdatedAt            time.Time                `json:"updated_at"`
}

// AlertCategory represents the category of classification alert
type AlertCategory string

const (
	AlertCategoryAccuracy    AlertCategory = "accuracy"
	AlertCategoryMLModel     AlertCategory = "ml_model"
	AlertCategoryEnsemble    AlertCategory = "ensemble"
	AlertCategorySecurity    AlertCategory = "security"
	AlertCategoryPerformance AlertCategory = "performance"
	AlertCategoryDataQuality AlertCategory = "data_quality"
)

// ClassificationMetricType represents the type of classification metric
type ClassificationMetricType string

const (
	MetricTypeOverallAccuracy      ClassificationMetricType = "overall_accuracy"
	MetricTypeIndustryAccuracy     ClassificationMetricType = "industry_accuracy"
	MetricTypeConfidenceScore      ClassificationMetricType = "confidence_score"
	MetricTypeBERTModelDrift       ClassificationMetricType = "bert_model_drift"
	MetricTypeBERTModelAccuracy    ClassificationMetricType = "bert_model_accuracy"
	MetricTypeEnsembleDisagreement ClassificationMetricType = "ensemble_disagreement"
	MetricTypeWeightDistribution   ClassificationMetricType = "weight_distribution"
	MetricTypeSecurityViolation    ClassificationMetricType = "security_violation"
	MetricTypeDataSourceTrust      ClassificationMetricType = "data_source_trust"
	MetricTypeWebsiteVerification  ClassificationMetricType = "website_verification"
	MetricTypeProcessingLatency    ClassificationMetricType = "processing_latency"
	MetricTypeErrorRate            ClassificationMetricType = "error_rate"
	MetricTypeThroughput           ClassificationMetricType = "throughput"
)

// ClassificationAlertConfig holds configuration for classification alert manager
type ClassificationAlertConfig struct {
	Enabled              bool
	EvaluationInterval   time.Duration
	NotificationTimeout  time.Duration
	MaxRetries           int
	RetryInterval        time.Duration
	SuppressionEnabled   bool
	SuppressionDuration  time.Duration
	DeduplicationEnabled bool
	EscalationEnabled    bool
	Environment          string
	ServiceName          string
	Version              string
}

// NewClassificationAlertManager creates a new classification alert manager
func NewClassificationAlertManager(
	baseAlertManager *AlertManager,
	logger *zap.Logger,
	config *ClassificationAlertConfig,
) *ClassificationAlertManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &ClassificationAlertManager{
		baseAlertManager:    baseAlertManager,
		logger:              logger,
		classificationRules: make(map[string]*ClassificationAlertRule),
		ctx:                 ctx,
		cancel:              cancel,
	}
}

// Start starts the classification alert manager
func (cam *ClassificationAlertManager) Start() error {
	cam.mu.Lock()
	if cam.started {
		cam.mu.Unlock()
		return fmt.Errorf("classification alert manager already started")
	}
	cam.started = true
	cam.mu.Unlock()

	cam.logger.Info("Starting classification alert manager",
		zap.String("service_name", cam.baseAlertManager.config.ServiceName),
		zap.String("version", cam.baseAlertManager.config.Version),
		zap.String("environment", cam.baseAlertManager.config.Environment),
	)

	// Initialize classification-specific alert rules
	if err := cam.initializeClassificationAlertRules(); err != nil {
		cam.mu.Lock()
		cam.started = false
		cam.mu.Unlock()
		return fmt.Errorf("failed to initialize classification alert rules: %w", err)
	}

	// Start alert evaluation
	go cam.startClassificationAlertEvaluation()
	cam.logger.Info("Classification alert manager started successfully")
	return nil
}

// Stop stops the classification alert manager
func (cam *ClassificationAlertManager) Stop() error {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	if !cam.started {
		return fmt.Errorf("classification alert manager not started")
	}

	cam.logger.Info("Stopping classification alert manager")

	cam.cancel()
	cam.started = false

	cam.logger.Info("Classification alert manager stopped successfully")
	return nil
}

// AddClassificationAlertRule adds a new classification alert rule
func (cam *ClassificationAlertManager) AddClassificationAlertRule(rule *ClassificationAlertRule) error {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	if rule.ID == "" {
		return fmt.Errorf("classification alert rule ID cannot be empty")
	}

	if _, exists := cam.classificationRules[rule.ID]; exists {
		return fmt.Errorf("classification alert rule with ID %s already exists", rule.ID)
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	cam.classificationRules[rule.ID] = rule

	cam.logger.Info("Classification alert rule added",
		zap.String("rule_id", rule.ID),
		zap.String("name", rule.Name),
		zap.String("category", string(rule.Category)),
		zap.String("severity", string(rule.Severity)),
	)

	return nil
}

// RemoveClassificationAlertRule removes a classification alert rule
func (cam *ClassificationAlertManager) RemoveClassificationAlertRule(ruleID string) error {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	if _, exists := cam.classificationRules[ruleID]; !exists {
		return fmt.Errorf("classification alert rule with ID %s not found", ruleID)
	}

	delete(cam.classificationRules, ruleID)

	cam.logger.Info("Classification alert rule removed",
		zap.String("rule_id", ruleID),
	)

	return nil
}

// GetClassificationAlertRule returns a classification alert rule
func (cam *ClassificationAlertManager) GetClassificationAlertRule(ruleID string) (*ClassificationAlertRule, error) {
	cam.mu.RLock()
	defer cam.mu.RUnlock()

	rule, exists := cam.classificationRules[ruleID]
	if !exists {
		return nil, fmt.Errorf("classification alert rule with ID %s not found", ruleID)
	}

	// Return a copy
	return &ClassificationAlertRule{
		ID:                   rule.ID,
		Name:                 rule.Name,
		Description:          rule.Description,
		Category:             rule.Category,
		MetricType:           rule.MetricType,
		Query:                rule.Query,
		Condition:            rule.Condition,
		Threshold:            rule.Threshold,
		Severity:             rule.Severity,
		Duration:             rule.Duration,
		Labels:               rule.Labels,
		Annotations:          rule.Annotations,
		NotificationChannels: rule.NotificationChannels,
		EscalationPolicy:     rule.EscalationPolicy,
		Enabled:              rule.Enabled,
		CreatedAt:            rule.CreatedAt,
		UpdatedAt:            rule.UpdatedAt,
	}, nil
}

// ListClassificationAlertRules returns all classification alert rules
func (cam *ClassificationAlertManager) ListClassificationAlertRules() []*ClassificationAlertRule {
	cam.mu.RLock()
	defer cam.mu.RUnlock()

	rules := make([]*ClassificationAlertRule, 0, len(cam.classificationRules))
	for _, rule := range cam.classificationRules {
		rules = append(rules, &ClassificationAlertRule{
			ID:                   rule.ID,
			Name:                 rule.Name,
			Description:          rule.Description,
			Category:             rule.Category,
			MetricType:           rule.MetricType,
			Query:                rule.Query,
			Condition:            rule.Condition,
			Threshold:            rule.Threshold,
			Severity:             rule.Severity,
			Duration:             rule.Duration,
			Labels:               rule.Labels,
			Annotations:          rule.Annotations,
			NotificationChannels: rule.NotificationChannels,
			EscalationPolicy:     rule.EscalationPolicy,
			Enabled:              rule.Enabled,
			CreatedAt:            rule.CreatedAt,
			UpdatedAt:            rule.UpdatedAt,
		})
	}

	return rules
}

// initializeClassificationAlertRules initializes classification-specific alert rules
func (cam *ClassificationAlertManager) initializeClassificationAlertRules() error {
	// Accuracy alerts (95%+ target)
	accuracyRules := cam.createAccuracyAlertRules()
	for _, rule := range accuracyRules {
		if err := cam.AddClassificationAlertRule(rule); err != nil {
			return fmt.Errorf("failed to add accuracy rule %s: %w", rule.ID, err)
		}
	}

	// ML model performance alerts
	mlModelRules := cam.createMLModelAlertRules()
	for _, rule := range mlModelRules {
		if err := cam.AddClassificationAlertRule(rule); err != nil {
			return fmt.Errorf("failed to add ML model rule %s: %w", rule.ID, err)
		}
	}

	// Ensemble disagreement alerts
	ensembleRules := cam.createEnsembleAlertRules()
	for _, rule := range ensembleRules {
		if err := cam.AddClassificationAlertRule(rule); err != nil {
			return fmt.Errorf("failed to add ensemble rule %s: %w", rule.ID, err)
		}
	}

	// Security violation alerts
	securityRules := cam.createSecurityAlertRules()
	for _, rule := range securityRules {
		if err := cam.AddClassificationAlertRule(rule); err != nil {
			return fmt.Errorf("failed to add security rule %s: %w", rule.ID, err)
		}
	}

	// Performance alerts
	performanceRules := cam.createPerformanceAlertRules()
	for _, rule := range performanceRules {
		if err := cam.AddClassificationAlertRule(rule); err != nil {
			return fmt.Errorf("failed to add performance rule %s: %w", rule.ID, err)
		}
	}

	return nil
}

// createAccuracyAlertRules creates accuracy-related alert rules
func (cam *ClassificationAlertManager) createAccuracyAlertRules() []*ClassificationAlertRule {
	return []*ClassificationAlertRule{
		{
			ID:          "overall_accuracy_low",
			Name:        "Overall Accuracy Below 95%",
			Description: "Alert when overall classification accuracy drops below 95%",
			Category:    AlertCategoryAccuracy,
			MetricType:  MetricTypeOverallAccuracy,
			Query:       "kyb_classification_accuracy_overall",
			Condition:   "lt",
			Threshold:   0.95,
			Severity:    AlertSeverityCritical,
			Duration:    2 * time.Minute,
			Labels: map[string]string{
				"service":     cam.baseAlertManager.config.ServiceName,
				"environment": cam.baseAlertManager.config.Environment,
				"category":    "accuracy",
			},
			Annotations: map[string]string{
				"summary":     "Classification accuracy below 95%",
				"description": "Overall classification accuracy is {{ $value }}% for the last 2 minutes",
			},
			NotificationChannels: []string{"email", "slack"},
			EscalationPolicy:     "critical",
			Enabled:              true,
		},
		{
			ID:          "industry_accuracy_low",
			Name:        "Industry Accuracy Below 90%",
			Description: "Alert when any industry classification accuracy drops below 90%",
			Category:    AlertCategoryAccuracy,
			MetricType:  MetricTypeIndustryAccuracy,
			Query:       "min(kyb_classification_accuracy_by_industry)",
			Condition:   "lt",
			Threshold:   0.90,
			Severity:    AlertSeverityWarning,
			Duration:    5 * time.Minute,
			Labels: map[string]string{
				"service":     cam.baseAlertManager.config.ServiceName,
				"environment": cam.baseAlertManager.config.Environment,
				"category":    "accuracy",
			},
			Annotations: map[string]string{
				"summary":     "Industry classification accuracy below 90%",
				"description": "Minimum industry classification accuracy is {{ $value }}% for the last 5 minutes",
			},
			NotificationChannels: []string{"slack"},
			EscalationPolicy:     "default",
			Enabled:              true,
		},
		{
			ID:          "confidence_score_low",
			Name:        "Average Confidence Score Below 0.8",
			Description: "Alert when average confidence score drops below 0.8",
			Category:    AlertCategoryAccuracy,
			MetricType:  MetricTypeConfidenceScore,
			Query:       "avg(kyb_classification_confidence_score)",
			Condition:   "lt",
			Threshold:   0.8,
			Severity:    AlertSeverityWarning,
			Duration:    3 * time.Minute,
			Labels: map[string]string{
				"service":     cam.baseAlertManager.config.ServiceName,
				"environment": cam.baseAlertManager.config.Environment,
				"category":    "accuracy",
			},
			Annotations: map[string]string{
				"summary":     "Average confidence score below 0.8",
				"description": "Average confidence score is {{ $value }} for the last 3 minutes",
			},
			NotificationChannels: []string{"slack"},
			EscalationPolicy:     "default",
			Enabled:              true,
		},
	}
}

// createMLModelAlertRules creates ML model performance alert rules
func (cam *ClassificationAlertManager) createMLModelAlertRules() []*ClassificationAlertRule {
	return []*ClassificationAlertRule{
		{
			ID:          "bert_model_drift_critical",
			Name:        "BERT Model Drift Critical",
			Description: "Alert when BERT model drift exceeds critical threshold",
			Category:    AlertCategoryMLModel,
			MetricType:  MetricTypeBERTModelDrift,
			Query:       "kyb_bert_model_drift_score",
			Condition:   "gt",
			Threshold:   0.8,
			Severity:    AlertSeverityCritical,
			Duration:    1 * time.Minute,
			Labels: map[string]string{
				"service":     cam.baseAlertManager.config.ServiceName,
				"environment": cam.baseAlertManager.config.Environment,
				"category":    "ml_model",
				"model":       "bert",
			},
			Annotations: map[string]string{
				"summary":     "BERT model drift critical",
				"description": "BERT model drift score is {{ $value }} (threshold: 0.8)",
			},
			NotificationChannels: []string{"email", "slack", "webhook"},
			EscalationPolicy:     "critical",
			Enabled:              true,
		},
		{
			ID:          "bert_model_accuracy_low",
			Name:        "BERT Model Accuracy Low",
			Description: "Alert when BERT model accuracy drops below 85%",
			Category:    AlertCategoryMLModel,
			MetricType:  MetricTypeBERTModelAccuracy,
			Query:       "kyb_bert_model_accuracy",
			Condition:   "lt",
			Threshold:   0.85,
			Severity:    AlertSeverityWarning,
			Duration:    3 * time.Minute,
			Labels: map[string]string{
				"service":     cam.baseAlertManager.config.ServiceName,
				"environment": cam.baseAlertManager.config.Environment,
				"category":    "ml_model",
				"model":       "bert",
			},
			Annotations: map[string]string{
				"summary":     "BERT model accuracy low",
				"description": "BERT model accuracy is {{ $value }}% (threshold: 85%)",
			},
			NotificationChannels: []string{"slack"},
			EscalationPolicy:     "default",
			Enabled:              true,
		},
	}
}

// createEnsembleAlertRules creates ensemble disagreement alert rules
func (cam *ClassificationAlertManager) createEnsembleAlertRules() []*ClassificationAlertRule {
	return []*ClassificationAlertRule{
		{
			ID:          "ensemble_disagreement_high",
			Name:        "High Ensemble Disagreement",
			Description: "Alert when ensemble methods disagree significantly",
			Category:    AlertCategoryEnsemble,
			MetricType:  MetricTypeEnsembleDisagreement,
			Query:       "kyb_ensemble_disagreement_score",
			Condition:   "gt",
			Threshold:   0.3,
			Severity:    AlertSeverityWarning,
			Duration:    5 * time.Minute,
			Labels: map[string]string{
				"service":     cam.baseAlertManager.config.ServiceName,
				"environment": cam.baseAlertManager.config.Environment,
				"category":    "ensemble",
			},
			Annotations: map[string]string{
				"summary":     "High ensemble disagreement",
				"description": "Ensemble disagreement score is {{ $value }} (threshold: 0.3)",
			},
			NotificationChannels: []string{"slack"},
			EscalationPolicy:     "default",
			Enabled:              true,
		},
		{
			ID:          "weight_distribution_unbalanced",
			Name:        "Unbalanced Weight Distribution",
			Description: "Alert when ensemble weight distribution becomes unbalanced",
			Category:    AlertCategoryEnsemble,
			MetricType:  MetricTypeWeightDistribution,
			Query:       "kyb_ensemble_weight_entropy",
			Condition:   "lt",
			Threshold:   0.5,
			Severity:    AlertSeverityWarning,
			Duration:    10 * time.Minute,
			Labels: map[string]string{
				"service":     cam.baseAlertManager.config.ServiceName,
				"environment": cam.baseAlertManager.config.Environment,
				"category":    "ensemble",
			},
			Annotations: map[string]string{
				"summary":     "Unbalanced weight distribution",
				"description": "Ensemble weight entropy is {{ $value }} (threshold: 0.5)",
			},
			NotificationChannels: []string{"slack"},
			EscalationPolicy:     "default",
			Enabled:              true,
		},
	}
}

// createSecurityAlertRules creates security violation alert rules
func (cam *ClassificationAlertManager) createSecurityAlertRules() []*ClassificationAlertRule {
	return []*ClassificationAlertRule{
		{
			ID:          "security_violation_detected",
			Name:        "Security Violation Detected",
			Description: "Alert when security validation fails",
			Category:    AlertCategorySecurity,
			MetricType:  MetricTypeSecurityViolation,
			Query:       "kyb_security_violations_total",
			Condition:   "gt",
			Threshold:   0,
			Severity:    AlertSeverityCritical,
			Duration:    0,
			Labels: map[string]string{
				"service":     cam.baseAlertManager.config.ServiceName,
				"environment": cam.baseAlertManager.config.Environment,
				"category":    "security",
			},
			Annotations: map[string]string{
				"summary":     "Security violation detected",
				"description": "Security validation failed - {{ $value }} violations detected",
			},
			NotificationChannels: []string{"email", "slack", "webhook"},
			EscalationPolicy:     "critical",
			Enabled:              true,
		},
		{
			ID:          "data_source_trust_low",
			Name:        "Data Source Trust Rate Low",
			Description: "Alert when data source trust rate drops below 95%",
			Category:    AlertCategorySecurity,
			MetricType:  MetricTypeDataSourceTrust,
			Query:       "kyb_data_source_trust_rate",
			Condition:   "lt",
			Threshold:   0.95,
			Severity:    AlertSeverityWarning,
			Duration:    2 * time.Minute,
			Labels: map[string]string{
				"service":     cam.baseAlertManager.config.ServiceName,
				"environment": cam.baseAlertManager.config.Environment,
				"category":    "security",
			},
			Annotations: map[string]string{
				"summary":     "Data source trust rate low",
				"description": "Data source trust rate is {{ $value }}% (threshold: 95%)",
			},
			NotificationChannels: []string{"slack"},
			EscalationPolicy:     "default",
			Enabled:              true,
		},
		{
			ID:          "website_verification_failed",
			Name:        "Website Verification Failed",
			Description: "Alert when website verification fails",
			Category:    AlertCategorySecurity,
			MetricType:  MetricTypeWebsiteVerification,
			Query:       "kyb_website_verification_failures_total",
			Condition:   "gt",
			Threshold:   0,
			Severity:    AlertSeverityWarning,
			Duration:    0,
			Labels: map[string]string{
				"service":     cam.baseAlertManager.config.ServiceName,
				"environment": cam.baseAlertManager.config.Environment,
				"category":    "security",
			},
			Annotations: map[string]string{
				"summary":     "Website verification failed",
				"description": "Website verification failed - {{ $value }} failures detected",
			},
			NotificationChannels: []string{"slack"},
			EscalationPolicy:     "default",
			Enabled:              true,
		},
	}
}

// createPerformanceAlertRules creates performance-related alert rules
func (cam *ClassificationAlertManager) createPerformanceAlertRules() []*ClassificationAlertRule {
	return []*ClassificationAlertRule{
		{
			ID:          "processing_latency_high",
			Name:        "High Processing Latency",
			Description: "Alert when processing latency exceeds 1 second",
			Category:    AlertCategoryPerformance,
			MetricType:  MetricTypeProcessingLatency,
			Query:       "histogram_quantile(0.95, kyb_classification_processing_duration_seconds_bucket)",
			Condition:   "gt",
			Threshold:   1.0,
			Severity:    AlertSeverityWarning,
			Duration:    3 * time.Minute,
			Labels: map[string]string{
				"service":     cam.baseAlertManager.config.ServiceName,
				"environment": cam.baseAlertManager.config.Environment,
				"category":    "performance",
			},
			Annotations: map[string]string{
				"summary":     "High processing latency",
				"description": "95th percentile processing latency is {{ $value }} seconds",
			},
			NotificationChannels: []string{"slack"},
			EscalationPolicy:     "default",
			Enabled:              true,
		},
		{
			ID:          "classification_error_rate_high",
			Name:        "High Classification Error Rate",
			Description: "Alert when classification error rate exceeds 5%",
			Category:    AlertCategoryPerformance,
			MetricType:  MetricTypeErrorRate,
			Query:       "rate(kyb_classification_errors_total[5m]) / rate(kyb_classification_requests_total[5m]) * 100",
			Condition:   "gt",
			Threshold:   5.0,
			Severity:    AlertSeverityCritical,
			Duration:    2 * time.Minute,
			Labels: map[string]string{
				"service":     cam.baseAlertManager.config.ServiceName,
				"environment": cam.baseAlertManager.config.Environment,
				"category":    "performance",
			},
			Annotations: map[string]string{
				"summary":     "High classification error rate",
				"description": "Classification error rate is {{ $value }}% for the last 5 minutes",
			},
			NotificationChannels: []string{"email", "slack"},
			EscalationPolicy:     "critical",
			Enabled:              true,
		},
		{
			ID:          "classification_throughput_low",
			Name:        "Low Classification Throughput",
			Description: "Alert when classification throughput drops below 10 requests/second",
			Category:    AlertCategoryPerformance,
			MetricType:  MetricTypeThroughput,
			Query:       "rate(kyb_classification_requests_total[1m])",
			Condition:   "lt",
			Threshold:   10.0,
			Severity:    AlertSeverityWarning,
			Duration:    5 * time.Minute,
			Labels: map[string]string{
				"service":     cam.baseAlertManager.config.ServiceName,
				"environment": cam.baseAlertManager.config.Environment,
				"category":    "performance",
			},
			Annotations: map[string]string{
				"summary":     "Low classification throughput",
				"description": "Classification throughput is {{ $value }} requests/second",
			},
			NotificationChannels: []string{"slack"},
			EscalationPolicy:     "default",
			Enabled:              true,
		},
	}
}

// startClassificationAlertEvaluation starts the classification alert evaluation process
func (cam *ClassificationAlertManager) startClassificationAlertEvaluation() {
	ticker := time.NewTicker(30 * time.Second) // Evaluate every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-cam.ctx.Done():
			cam.logger.Info("Classification alert evaluation stopped")
			return
		case <-ticker.C:
			cam.evaluateClassificationAlertRules()
		}
	}
}

// evaluateClassificationAlertRules evaluates all enabled classification alert rules
func (cam *ClassificationAlertManager) evaluateClassificationAlertRules() {
	cam.mu.RLock()
	rules := make([]*ClassificationAlertRule, 0, len(cam.classificationRules))
	for _, rule := range cam.classificationRules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	cam.mu.RUnlock()

	for _, rule := range rules {
		cam.evaluateClassificationAlertRule(rule)
	}
}

// evaluateClassificationAlertRule evaluates a specific classification alert rule
func (cam *ClassificationAlertManager) evaluateClassificationAlertRule(rule *ClassificationAlertRule) {
	// In a real implementation, this would evaluate the query against metrics
	// For now, we'll simulate evaluation based on metric type
	value := cam.simulateMetricValue(rule.MetricType)

	// Check if the condition is met
	conditionMet := false
	switch rule.Condition {
	case "gt":
		conditionMet = value > rule.Threshold
	case "gte":
		conditionMet = value >= rule.Threshold
	case "lt":
		conditionMet = value < rule.Threshold
	case "lte":
		conditionMet = value <= rule.Threshold
	case "eq":
		conditionMet = value == rule.Threshold
	case "ne":
		conditionMet = value != rule.Threshold
	}

	if conditionMet {
		cam.triggerClassificationAlert(rule, value)
	} else {
		cam.resolveClassificationAlert(rule.ID)
	}
}

// simulateMetricValue simulates metric values for testing
func (cam *ClassificationAlertManager) simulateMetricValue(metricType ClassificationMetricType) float64 {
	// In production, this would query actual metrics from the monitoring system
	switch metricType {
	case MetricTypeOverallAccuracy:
		return 0.92 + float64(time.Now().Unix()%10)/100 // Simulate 92-101% accuracy
	case MetricTypeIndustryAccuracy:
		return 0.88 + float64(time.Now().Unix()%15)/100 // Simulate 88-102% accuracy
	case MetricTypeConfidenceScore:
		return 0.75 + float64(time.Now().Unix()%25)/100 // Simulate 75-99% confidence
	case MetricTypeBERTModelDrift:
		return float64(time.Now().Unix()%100) / 100 // Simulate 0-99% drift
	case MetricTypeBERTModelAccuracy:
		return 0.80 + float64(time.Now().Unix()%20)/100 // Simulate 80-99% accuracy
	case MetricTypeEnsembleDisagreement:
		return float64(time.Now().Unix()%50) / 100 // Simulate 0-49% disagreement
	case MetricTypeWeightDistribution:
		return 0.3 + float64(time.Now().Unix()%70)/100 // Simulate 30-99% entropy
	case MetricTypeSecurityViolation:
		return float64(time.Now().Unix() % 5) // Simulate 0-4 violations
	case MetricTypeDataSourceTrust:
		return 0.90 + float64(time.Now().Unix()%10)/100 // Simulate 90-99% trust
	case MetricTypeWebsiteVerification:
		return float64(time.Now().Unix() % 3) // Simulate 0-2 failures
	case MetricTypeProcessingLatency:
		return 0.5 + float64(time.Now().Unix()%150)/100 // Simulate 0.5-2.0 seconds
	case MetricTypeErrorRate:
		return float64(time.Now().Unix() % 10) // Simulate 0-9% error rate
	case MetricTypeThroughput:
		return 5.0 + float64(time.Now().Unix()%20) // Simulate 5-24 requests/second
	default:
		return 0.0
	}
}

// triggerClassificationAlert triggers a classification alert
func (cam *ClassificationAlertManager) triggerClassificationAlert(rule *ClassificationAlertRule, value float64) {
	// Create a base alert rule for the base alert manager
	baseRule := &AlertRule{
		ID:                   rule.ID,
		Name:                 rule.Name,
		Description:          rule.Description,
		Query:                rule.Query,
		Condition:            rule.Condition,
		Threshold:            rule.Threshold,
		Severity:             rule.Severity,
		Duration:             rule.Duration,
		Labels:               rule.Labels,
		Annotations:          rule.Annotations,
		NotificationChannels: rule.NotificationChannels,
		EscalationPolicy:     rule.EscalationPolicy,
		Enabled:              rule.Enabled,
		CreatedAt:            rule.CreatedAt,
		UpdatedAt:            rule.UpdatedAt,
	}

	// Add the rule to the base alert manager if it doesn't exist
	cam.baseAlertManager.AddAlertRule(baseRule)

	// Trigger the alert
	cam.baseAlertManager.triggerAlert(baseRule, value)

	cam.logger.Info("Classification alert triggered",
		zap.String("rule_id", rule.ID),
		zap.String("name", rule.Name),
		zap.String("category", string(rule.Category)),
		zap.String("severity", string(rule.Severity)),
		zap.Float64("value", value),
		zap.Float64("threshold", rule.Threshold),
	)
}

// resolveClassificationAlert resolves a classification alert
func (cam *ClassificationAlertManager) resolveClassificationAlert(ruleID string) {
	cam.baseAlertManager.resolveAlert(ruleID)

	cam.logger.Debug("Classification alert resolved",
		zap.String("rule_id", ruleID),
	)
}

// GetActiveClassificationAlerts returns all active classification alerts
func (cam *ClassificationAlertManager) GetActiveClassificationAlerts() []*Alert {
	return cam.baseAlertManager.GetActiveAlerts()
}

// GetClassificationAlertsByCategory returns alerts filtered by category
func (cam *ClassificationAlertManager) GetClassificationAlertsByCategory(category AlertCategory) []*Alert {
	activeAlerts := cam.baseAlertManager.GetActiveAlerts()
	filteredAlerts := make([]*Alert, 0)

	for _, alert := range activeAlerts {
		if alert.Labels["category"] == string(category) {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}

	return filteredAlerts
}

// GetClassificationAlertsBySeverity returns alerts filtered by severity
func (cam *ClassificationAlertManager) GetClassificationAlertsBySeverity(severity AlertSeverity) []*Alert {
	activeAlerts := cam.baseAlertManager.GetActiveAlerts()
	filteredAlerts := make([]*Alert, 0)

	for _, alert := range activeAlerts {
		if alert.Severity == severity {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}

	return filteredAlerts
}

// AcknowledgeClassificationAlert acknowledges a classification alert
func (cam *ClassificationAlertManager) AcknowledgeClassificationAlert(alertID string) error {
	// This would be implemented to acknowledge alerts in the base alert manager
	cam.logger.Info("Classification alert acknowledged", zap.String("alert_id", alertID))
	return nil
}

// GetClassificationAlertSummary returns a summary of classification alerts
func (cam *ClassificationAlertManager) GetClassificationAlertSummary() *ClassificationAlertSummary {
	activeAlerts := cam.baseAlertManager.GetActiveAlerts()

	summary := &ClassificationAlertSummary{
		Timestamp:          time.Now(),
		TotalAlerts:        len(activeAlerts),
		CriticalAlerts:     0,
		WarningAlerts:      0,
		InfoAlerts:         0,
		AlertsByCategory:   make(map[AlertCategory]int),
		AlertsBySeverity:   make(map[AlertSeverity]int),
		AlertsByMetricType: make(map[ClassificationMetricType]int),
	}

	for _, alert := range activeAlerts {
		// Count by severity
		summary.AlertsBySeverity[alert.Severity]++
		switch alert.Severity {
		case AlertSeverityCritical:
			summary.CriticalAlerts++
		case AlertSeverityWarning:
			summary.WarningAlerts++
		case AlertSeverityInfo:
			summary.InfoAlerts++
		}

		// Count by category
		if category, exists := alert.Labels["category"]; exists {
			summary.AlertsByCategory[AlertCategory(category)]++
		}

		// Count by metric type
		if metricType, exists := alert.Labels["metric_type"]; exists {
			summary.AlertsByMetricType[ClassificationMetricType(metricType)]++
		}
	}

	return summary
}

// ClassificationAlertSummary represents a summary of classification alerts
type ClassificationAlertSummary struct {
	Timestamp          time.Time                        `json:"timestamp"`
	TotalAlerts        int                              `json:"total_alerts"`
	CriticalAlerts     int                              `json:"critical_alerts"`
	WarningAlerts      int                              `json:"warning_alerts"`
	InfoAlerts         int                              `json:"info_alerts"`
	AlertsByCategory   map[AlertCategory]int            `json:"alerts_by_category"`
	AlertsBySeverity   map[AlertSeverity]int            `json:"alerts_by_severity"`
	AlertsByMetricType map[ClassificationMetricType]int `json:"alerts_by_metric_type"`
}
