package observability

import (
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
)

// NewPerformanceRuleEngine creates a new performance rule engine
func NewPerformanceRuleEngine(
	rules map[string]*PerformanceAlertRule,
	config PerformanceAlertingConfig,
	logger *zap.Logger,
) *PerformanceRuleEngine {
	return &PerformanceRuleEngine{
		rules:  rules,
		config: config,
		logger: logger,
	}
}

// EvaluateRule evaluates a performance alert rule against current metrics
func (pre *PerformanceRuleEngine) EvaluateRule(
	rule *PerformanceAlertRule,
	metrics *PerformanceMetrics,
) (shouldFire bool, currentValue, baselineValue, trendValue, anomalyScore float64) {
	// Get current value for the metric type
	currentValue = pre.getMetricValue(rule.MetricType, metrics)
	if currentValue < 0 {
		pre.logger.Warn("Unable to get metric value",
			zap.String("metric_type", rule.MetricType),
			zap.String("rule_id", rule.ID))
		return false, 0, 0, 0, 0
	}

	// Get baseline value for comparison
	baselineValue = pre.getBaselineValue(rule.MetricType, metrics)

	// Calculate trend value if needed
	if rule.Condition == "trend" || rule.TrendWindow > 0 {
		trendValue = pre.calculateTrendValue(rule.MetricType, rule.TrendWindow, metrics)
	}

	// Calculate anomaly score if needed
	if rule.Condition == "anomaly" || rule.AnomalyScore > 0 {
		anomalyScore = pre.calculateAnomalyScore(rule.MetricType, metrics)
	}

	// Evaluate the condition
	shouldFire = pre.evaluateCondition(rule, currentValue, baselineValue, trendValue, anomalyScore)

	return shouldFire, currentValue, baselineValue, trendValue, anomalyScore
}

// getMetricValue gets the current value for a metric type
func (pre *PerformanceRuleEngine) getMetricValue(metricType string, metrics *PerformanceMetrics) float64 {
	switch metricType {
	case "response_time":
		return float64(metrics.AverageResponseTime.Milliseconds())
	case "success_rate":
		return metrics.OverallSuccessRate
	case "throughput":
		return metrics.RequestsPerSecond
	case "cpu":
		return metrics.CPUUsage
	case "memory":
		return metrics.MemoryUsage
	case "disk":
		return metrics.DiskUsage
	case "error_rate":
		return metrics.ErrorRate
	case "availability":
		return metrics.Availability
	case "latency_p95":
		return float64(metrics.P95ResponseTime.Milliseconds())
	case "latency_p99":
		return float64(metrics.P99ResponseTime.Milliseconds())
	default:
		pre.logger.Warn("Unknown metric type", zap.String("metric_type", metricType))
		return -1
	}
}

// getBaselineValue gets the baseline value for a metric type
func (pre *PerformanceRuleEngine) getBaselineValue(metricType string, metrics *PerformanceMetrics) float64 {
	// In a real implementation, this would calculate baseline from historical data
	// For now, we'll use a simple approach based on the metric type
	switch metricType {
	case "response_time":
		return 200.0 // 200ms baseline
	case "success_rate":
		return 0.99 // 99% baseline
	case "throughput":
		return 1000.0 // 1000 req/s baseline
	case "cpu":
		return 50.0 // 50% baseline
	case "memory":
		return 60.0 // 60% baseline
	case "disk":
		return 70.0 // 70% baseline
	case "error_rate":
		return 0.01 // 1% baseline
	case "availability":
		return 0.999 // 99.9% baseline
	case "latency_p95":
		return 500.0 // 500ms baseline
	case "latency_p99":
		return 1000.0 // 1000ms baseline
	default:
		return 0.0
	}
}

// calculateTrendValue calculates the trend value for a metric
func (pre *PerformanceRuleEngine) calculateTrendValue(metricType string, trendWindow time.Duration, metrics *PerformanceMetrics) float64 {
	// In a real implementation, this would analyze historical data
	// For now, we'll return a simple trend calculation
	currentValue := pre.getMetricValue(metricType, metrics)
	baselineValue := pre.getBaselineValue(metricType, metrics)

	// Calculate percentage change from baseline
	if baselineValue > 0 {
		return ((currentValue - baselineValue) / baselineValue) * 100
	}
	return 0.0
}

// calculateAnomalyScore calculates the anomaly score for a metric
func (pre *PerformanceRuleEngine) calculateAnomalyScore(metricType string, metrics *PerformanceMetrics) float64 {
	// In a real implementation, this would use statistical methods
	// For now, we'll use a simple z-score approach
	currentValue := pre.getMetricValue(metricType, metrics)
	baselineValue := pre.getBaselineValue(metricType, metrics)

	// Simple standard deviation estimation
	stdDev := baselineValue * 0.1 // Assume 10% standard deviation

	if stdDev > 0 {
		zScore := math.Abs((currentValue - baselineValue) / stdDev)
		// Convert z-score to anomaly score (0-1)
		return math.Min(zScore/3.0, 1.0) // Cap at 1.0
	}
	return 0.0
}

// evaluateCondition evaluates whether an alert condition is met
func (pre *PerformanceRuleEngine) evaluateCondition(
	rule *PerformanceAlertRule,
	currentValue, baselineValue, trendValue, anomalyScore float64,
) bool {
	switch rule.Condition {
	case "threshold":
		return pre.evaluateThresholdCondition(rule, currentValue)
	case "trend":
		return pre.evaluateTrendCondition(rule, trendValue)
	case "anomaly":
		return pre.evaluateAnomalyCondition(rule, anomalyScore)
	case "prediction":
		return pre.evaluatePredictionCondition(rule, currentValue, baselineValue)
	case "baseline":
		return pre.evaluateBaselineCondition(rule, currentValue, baselineValue)
	default:
		pre.logger.Warn("Unknown condition type", zap.String("condition", rule.Condition))
		return false
	}
}

// evaluateThresholdCondition evaluates a threshold-based condition
func (pre *PerformanceRuleEngine) evaluateThresholdCondition(rule *PerformanceAlertRule, currentValue float64) bool {
	switch rule.Operator {
	case "gt":
		return currentValue > rule.Threshold
	case "gte":
		return currentValue >= rule.Threshold
	case "lt":
		return currentValue < rule.Threshold
	case "lte":
		return currentValue <= rule.Threshold
	case "eq":
		return math.Abs(currentValue-rule.Threshold) < 0.001 // Small epsilon for float comparison
	case "ne":
		return math.Abs(currentValue-rule.Threshold) >= 0.001
	default:
		pre.logger.Warn("Unknown operator", zap.String("operator", rule.Operator))
		return false
	}
}

// evaluateTrendCondition evaluates a trend-based condition
func (pre *PerformanceRuleEngine) evaluateTrendCondition(rule *PerformanceAlertRule, trendValue float64) bool {
	switch rule.TrendDirection {
	case "increasing":
		return trendValue > rule.Threshold
	case "decreasing":
		return trendValue < -rule.Threshold
	case "changing":
		return math.Abs(trendValue) > rule.Threshold
	default:
		// Default to absolute change
		return math.Abs(trendValue) > rule.Threshold
	}
}

// evaluateAnomalyCondition evaluates an anomaly-based condition
func (pre *PerformanceRuleEngine) evaluateAnomalyCondition(rule *PerformanceAlertRule, anomalyScore float64) bool {
	return anomalyScore > rule.AnomalyScore
}

// evaluatePredictionCondition evaluates a prediction-based condition
func (pre *PerformanceRuleEngine) evaluatePredictionCondition(rule *PerformanceAlertRule, currentValue, baselineValue float64) bool {
	// In a real implementation, this would use ML models for prediction
	// For now, we'll use a simple extrapolation
	predictedValue := currentValue * 1.1 // Assume 10% increase

	switch rule.Operator {
	case "gt":
		return predictedValue > rule.Threshold
	case "lt":
		return predictedValue < rule.Threshold
	default:
		return false
	}
}

// evaluateBaselineCondition evaluates a baseline-based condition
func (pre *PerformanceRuleEngine) evaluateBaselineCondition(rule *PerformanceAlertRule, currentValue, baselineValue float64) bool {
	if baselineValue <= 0 {
		return false
	}

	percentageChange := ((currentValue - baselineValue) / baselineValue) * 100

	switch rule.Operator {
	case "gt":
		return percentageChange > rule.Threshold
	case "lt":
		return percentageChange < -rule.Threshold
	case "gte":
		return percentageChange >= rule.Threshold
	case "lte":
		return percentageChange <= -rule.Threshold
	default:
		return false
	}
}

// ValidateRule validates a performance alert rule
func (pre *PerformanceRuleEngine) ValidateRule(rule *PerformanceAlertRule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}

	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if rule.MetricType == "" {
		return fmt.Errorf("metric type is required")
	}

	if rule.Condition == "" {
		return fmt.Errorf("condition is required")
	}

	if rule.Threshold <= 0 {
		return fmt.Errorf("threshold must be positive")
	}

	if rule.Duration <= 0 {
		return fmt.Errorf("duration must be positive")
	}

	// Validate metric type
	validMetricTypes := []string{
		"response_time", "success_rate", "throughput", "cpu", "memory", "disk",
		"error_rate", "availability", "latency_p95", "latency_p99",
	}

	valid := false
	for _, validType := range validMetricTypes {
		if rule.MetricType == validType {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid metric type: %s", rule.MetricType)
	}

	// Validate condition
	validConditions := []string{"threshold", "trend", "anomaly", "prediction", "baseline"}
	valid = false
	for _, validCondition := range validConditions {
		if rule.Condition == validCondition {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid condition: %s", rule.Condition)
	}

	// Validate operator
	validOperators := []string{"gt", "gte", "lt", "lte", "eq", "ne"}
	valid = false
	for _, validOp := range validOperators {
		if rule.Operator == validOp {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid operator: %s", rule.Operator)
	}

	// Validate severity
	validSeverities := []string{"info", "warning", "critical", "emergency"}
	valid = false
	for _, validSev := range validSeverities {
		if rule.Severity == validSev {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid severity: %s", rule.Severity)
	}

	return nil
}

// GetRuleStatistics returns statistics for a rule
func (pre *PerformanceRuleEngine) GetRuleStatistics(ruleID string) (*RuleStatistics, error) {
	rule, exists := pre.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule not found: %s", ruleID)
	}

	stats := &RuleStatistics{
		RuleID:           rule.ID,
		RuleName:         rule.Name,
		TotalEvaluations: rule.EvaluationCount,
		FiringCount:      rule.FiringCount,
		ResolvedCount:    rule.ResolvedCount,
		LastEvaluation:   rule.LastEvaluation,
	}

	if rule.EvaluationCount > 0 {
		stats.FiringRate = float64(rule.FiringCount) / float64(rule.EvaluationCount)
		stats.ResolutionRate = float64(rule.ResolvedCount) / float64(rule.EvaluationCount)
	}

	return stats, nil
}

// RuleStatistics represents statistics for a rule
type RuleStatistics struct {
	RuleID           string    `json:"rule_id"`
	RuleName         string    `json:"rule_name"`
	TotalEvaluations int64     `json:"total_evaluations"`
	FiringCount      int64     `json:"firing_count"`
	ResolvedCount    int64     `json:"resolved_count"`
	FiringRate       float64   `json:"firing_rate"`
	ResolutionRate   float64   `json:"resolution_rate"`
	LastEvaluation   time.Time `json:"last_evaluation"`
}
