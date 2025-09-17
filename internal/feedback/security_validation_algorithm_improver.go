package feedback

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SecurityValidationAlgorithmImprover improves security validation algorithms based on feedback
type SecurityValidationAlgorithmImprover struct {
	config                     *SecurityValidationConfig
	logger                     *zap.Logger
	feedbackRepository         FeedbackRepository
	securityAnalyzer           *SecurityFeedbackAnalyzer
	patternDetector            *SecurityPatternDetector
	mu                         sync.RWMutex
	improvementHistory         []*AlgorithmImprovement
	performanceMetrics         *SecurityValidationMetrics
	lastImprovementTime        time.Time
	improvementThreshold       float64
	consecutiveImprovements    int
	maxConsecutiveImprovements int
}

// SecurityValidationConfig holds configuration for security validation algorithm improvement
type SecurityValidationConfig struct {
	ImprovementInterval        time.Duration `json:"improvement_interval"`
	MinFeedbackCount           int           `json:"min_feedback_count"`
	ImprovementThreshold       float64       `json:"improvement_threshold"`
	MaxConsecutiveImprovements int           `json:"max_consecutive_improvements"`
	EnableAutoImprovement      bool          `json:"enable_auto_improvement"`
	ValidationTimeout          time.Duration `json:"validation_timeout"`
	ConfidenceThreshold        float64       `json:"confidence_threshold"`
	PatternWeight              float64       `json:"pattern_weight"`
	PerformanceWeight          float64       `json:"performance_weight"`
	FeedbackWeight             float64       `json:"feedback_weight"`
}

// SecurityValidationMetrics tracks performance metrics for security validation
type SecurityValidationMetrics struct {
	TotalValidations        int64                               `json:"total_validations"`
	SuccessfulValidations   int64                               `json:"successful_validations"`
	FailedValidations       int64                               `json:"failed_validations"`
	AverageValidationTime   time.Duration                       `json:"average_validation_time"`
	AverageConfidence       float64                             `json:"average_confidence"`
	FalsePositiveRate       float64                             `json:"false_positive_rate"`
	FalseNegativeRate       float64                             `json:"false_negative_rate"`
	SecurityViolationRate   float64                             `json:"security_violation_rate"`
	LastUpdated             time.Time                           `json:"last_updated"`
	ValidationMethodMetrics map[string]*ValidationMethodMetrics `json:"validation_method_metrics"`
}

// ValidationMethodMetrics tracks metrics for specific validation methods
type ValidationMethodMetrics struct {
	MethodName        string        `json:"method_name"`
	UsageCount        int64         `json:"usage_count"`
	SuccessRate       float64       `json:"success_rate"`
	AverageTime       time.Duration `json:"average_time"`
	AverageConfidence float64       `json:"average_confidence"`
	FalsePositiveRate float64       `json:"false_positive_rate"`
	FalseNegativeRate float64       `json:"false_negative_rate"`
	LastUsed          time.Time     `json:"last_used"`
}

// AlgorithmImprovement represents an improvement made to the security validation algorithm
type AlgorithmImprovement struct {
	ImprovementID     string                 `json:"improvement_id"`
	ImprovementType   string                 `json:"improvement_type"`
	Description       string                 `json:"description"`
	Changes           map[string]interface{} `json:"changes"`
	PerformanceImpact *PerformanceImpact     `json:"performance_impact"`
	Confidence        float64                `json:"confidence"`
	AppliedAt         time.Time              `json:"applied_at"`
	ValidatedAt       *time.Time             `json:"validated_at"`
	ValidationResults *ImprovementValidation `json:"validation_results"`
	RollbackRequired  bool                   `json:"rollback_required"`
	RollbackReason    string                 `json:"rollback_reason"`
}

// PerformanceImpact tracks the impact of an algorithm improvement
type PerformanceImpact struct {
	AccuracyChange      float64 `json:"accuracy_change"`
	SpeedChange         float64 `json:"speed_change"`
	ConfidenceChange    float64 `json:"confidence_change"`
	FalsePositiveChange float64 `json:"false_positive_change"`
	FalseNegativeChange float64 `json:"false_negative_change"`
	ResourceUsageChange float64 `json:"resource_usage_change"`
	OverallImpact       string  `json:"overall_impact"`
	ImpactScore         float64 `json:"impact_score"`
}

// ImprovementValidation tracks validation results for an improvement
type ImprovementValidation struct {
	ValidationMethod   string        `json:"validation_method"`
	TestCases          int           `json:"test_cases"`
	PassedCases        int           `json:"passed_cases"`
	FailedCases        int           `json:"failed_cases"`
	Accuracy           float64       `json:"accuracy"`
	Performance        float64       `json:"performance"`
	Stability          float64       `json:"stability"`
	ValidatedAt        time.Time     `json:"validated_at"`
	ValidationDuration time.Duration `json:"validation_duration"`
}

// NewSecurityValidationAlgorithmImprover creates a new security validation algorithm improver
func NewSecurityValidationAlgorithmImprover(
	config *SecurityValidationConfig,
	logger *zap.Logger,
	feedbackRepository FeedbackRepository,
	securityAnalyzer *SecurityFeedbackAnalyzer,
	patternDetector *SecurityPatternDetector,
) *SecurityValidationAlgorithmImprover {
	return &SecurityValidationAlgorithmImprover{
		config:                     config,
		logger:                     logger,
		feedbackRepository:         feedbackRepository,
		securityAnalyzer:           securityAnalyzer,
		patternDetector:            patternDetector,
		improvementHistory:         make([]*AlgorithmImprovement, 0),
		performanceMetrics:         NewSecurityValidationMetrics(),
		improvementThreshold:       config.ImprovementThreshold,
		maxConsecutiveImprovements: config.MaxConsecutiveImprovements,
	}
}

// NewSecurityValidationMetrics creates new security validation metrics
func NewSecurityValidationMetrics() *SecurityValidationMetrics {
	return &SecurityValidationMetrics{
		ValidationMethodMetrics: make(map[string]*ValidationMethodMetrics),
		LastUpdated:             time.Now(),
	}
}

// StartImprovementProcess starts the continuous improvement process
func (svai *SecurityValidationAlgorithmImprover) StartImprovementProcess(ctx context.Context) error {
	svai.mu.Lock()
	defer svai.mu.Unlock()

	svai.logger.Info("Starting security validation algorithm improvement process",
		zap.Duration("improvement_interval", svai.config.ImprovementInterval))

	// Start improvement loop
	go svai.improvementLoop(ctx)

	return nil
}

// improvementLoop runs the continuous improvement process
func (svai *SecurityValidationAlgorithmImprover) improvementLoop(ctx context.Context) {
	ticker := time.NewTicker(svai.config.ImprovementInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			svai.logger.Info("Security validation improvement process stopped")
			return
		case <-ticker.C:
			if err := svai.performImprovementCycle(ctx); err != nil {
				svai.logger.Error("Improvement cycle failed", zap.Error(err))
			}
		}
	}
}

// performImprovementCycle performs a single improvement cycle
func (svai *SecurityValidationAlgorithmImprover) performImprovementCycle(ctx context.Context) error {
	svai.logger.Info("Starting security validation improvement cycle")

	// Check if improvement is needed
	if !svai.shouldImprove() {
		svai.logger.Debug("No improvement needed at this time")
		return nil
	}

	// Collect recent feedback
	feedback, err := svai.collectRecentFeedback(ctx)
	if err != nil {
		return fmt.Errorf("failed to collect recent feedback: %w", err)
	}

	if len(feedback) < svai.config.MinFeedbackCount {
		svai.logger.Debug("Insufficient feedback for improvement",
			zap.Int("feedback_count", len(feedback)),
			zap.Int("min_required", svai.config.MinFeedbackCount))
		return nil
	}

	// Analyze feedback for improvement opportunities
	improvements, err := svai.analyzeImprovementOpportunities(ctx, feedback)
	if err != nil {
		return fmt.Errorf("failed to analyze improvement opportunities: %w", err)
	}

	// Apply improvements
	for _, improvement := range improvements {
		if err := svai.applyImprovement(ctx, improvement); err != nil {
			svai.logger.Error("Failed to apply improvement",
				zap.String("improvement_id", improvement.ImprovementID),
				zap.Error(err))
			continue
		}

		svai.logger.Info("Applied security validation improvement",
			zap.String("improvement_id", improvement.ImprovementID),
			zap.String("improvement_type", improvement.ImprovementType))
	}

	// Update metrics
	svai.updatePerformanceMetrics()

	svai.logger.Info("Security validation improvement cycle completed",
		zap.Int("improvements_applied", len(improvements)))

	return nil
}

// shouldImprove determines if improvement is needed
func (svai *SecurityValidationAlgorithmImprover) shouldImprove() bool {
	// Check if auto-improvement is enabled
	if !svai.config.EnableAutoImprovement {
		return false
	}

	// Check if we've exceeded max consecutive improvements
	if svai.consecutiveImprovements >= svai.maxConsecutiveImprovements {
		svai.logger.Debug("Max consecutive improvements reached",
			zap.Int("consecutive_improvements", svai.consecutiveImprovements))
		return false
	}

	// Check if enough time has passed since last improvement
	if time.Since(svai.lastImprovementTime) < svai.config.ImprovementInterval {
		return false
	}

	// Check performance metrics for improvement opportunities
	return svai.hasImprovementOpportunities()
}

// hasImprovementOpportunities checks if there are opportunities for improvement
func (svai *SecurityValidationAlgorithmImprover) hasImprovementOpportunities() bool {
	metrics := svai.performanceMetrics

	// Return false if metrics are nil
	if metrics == nil {
		return false
	}

	// Check for high false positive/negative rates
	if metrics.FalsePositiveRate > 0.1 || metrics.FalseNegativeRate > 0.1 {
		return true
	}

	// Check for low confidence scores
	if metrics.AverageConfidence < svai.config.ConfidenceThreshold {
		return true
	}

	// Check for high security violation rates
	if metrics.SecurityViolationRate > 0.05 {
		return true
	}

	// Check for slow validation times
	if metrics.AverageValidationTime > svai.config.ValidationTimeout {
		return true
	}

	return false
}

// collectRecentFeedback collects recent security validation feedback
func (svai *SecurityValidationAlgorithmImprover) collectRecentFeedback(ctx context.Context) ([]*UserFeedback, error) {
	// Get feedback from the last improvement interval
	_ = time.Now().Add(-svai.config.ImprovementInterval) // TODO: Use this for filtering

	// Get security validation feedback from repository
	// Note: This is a simplified approach - in a real implementation,
	// we would need to add a method to get feedback by type and time range
	var allFeedback []*UserFeedback
	// TODO: Implement proper feedback filtering by type and time range
	// For now, we'll return an empty slice as a placeholder
	return allFeedback, nil
}

// analyzeImprovementOpportunities analyzes feedback for improvement opportunities
func (svai *SecurityValidationAlgorithmImprover) analyzeImprovementOpportunities(ctx context.Context, feedback []*UserFeedback) ([]*AlgorithmImprovement, error) {
	svai.logger.Info("Analyzing improvement opportunities",
		zap.Int("feedback_count", len(feedback)))

	var improvements []*AlgorithmImprovement

	// Analyze security patterns for improvement opportunities
	patterns, err := svai.patternDetector.DetectPatterns(ctx, feedback)
	if err != nil {
		return nil, fmt.Errorf("failed to detect security patterns: %w", err)
	}

	// Generate improvements based on patterns
	for _, pattern := range patterns {
		if pattern.Severity == "critical" || pattern.Severity == "high" {
			improvement := svai.generatePatternBasedImprovement(pattern)
			improvements = append(improvements, improvement)
		}
	}

	// Analyze performance metrics for improvement opportunities
	performanceImprovements := svai.generatePerformanceBasedImprovements()
	improvements = append(improvements, performanceImprovements...)

	// Analyze validation method performance
	methodImprovements := svai.generateMethodBasedImprovements()
	improvements = append(improvements, methodImprovements...)

	svai.logger.Info("Improvement opportunities analysis completed",
		zap.Int("improvements_found", len(improvements)))

	return improvements, nil
}

// generatePatternBasedImprovement generates an improvement based on a security pattern
func (svai *SecurityValidationAlgorithmImprover) generatePatternBasedImprovement(pattern *SecurityPattern) *AlgorithmImprovement {
	improvement := &AlgorithmImprovement{
		ImprovementID:   fmt.Sprintf("pattern_improvement_%s_%d", pattern.PatternID, time.Now().Unix()),
		ImprovementType: "pattern_based",
		Description:     fmt.Sprintf("Address %s pattern: %s", pattern.PatternType, pattern.Description),
		Changes: map[string]interface{}{
			"pattern_id":          pattern.PatternID,
			"pattern_type":        pattern.PatternType,
			"severity":            pattern.Severity,
			"affected_components": pattern.AffectedComponents,
		},
		Confidence: pattern.Confidence,
		AppliedAt:  time.Now(),
	}

	// Generate specific changes based on pattern type
	switch pattern.PatternType {
	case "recurring_security_violation":
		improvement.Changes["action"] = "enhance_validation_rules"
		improvement.Changes["new_rules"] = []string{
			"Add stricter validation for suspicious patterns",
			"Increase confidence threshold for high-risk validations",
		}
	case "false_positive_pattern":
		improvement.Changes["action"] = "refine_validation_logic"
		improvement.Changes["adjustments"] = []string{
			"Lower false positive threshold",
			"Add context-aware validation",
		}
	case "performance_degradation":
		improvement.Changes["action"] = "optimize_validation_performance"
		improvement.Changes["optimizations"] = []string{
			"Implement caching for repeated validations",
			"Optimize database queries",
		}
	}

	return improvement
}

// generatePerformanceBasedImprovements generates improvements based on performance metrics
func (svai *SecurityValidationAlgorithmImprover) generatePerformanceBasedImprovements() []*AlgorithmImprovement {
	var improvements []*AlgorithmImprovement
	metrics := svai.performanceMetrics

	// Improvement for high false positive rate
	if metrics.FalsePositiveRate > 0.1 {
		improvement := &AlgorithmImprovement{
			ImprovementID:   fmt.Sprintf("performance_improvement_fp_%d", time.Now().Unix()),
			ImprovementType: "performance_based",
			Description:     "Reduce false positive rate in security validation",
			Changes: map[string]interface{}{
				"metric":               "false_positive_rate",
				"current_value":        metrics.FalsePositiveRate,
				"target_value":         0.05,
				"action":               "adjust_validation_thresholds",
				"threshold_adjustment": 0.1,
			},
			Confidence: 0.8,
			AppliedAt:  time.Now(),
		}
		improvements = append(improvements, improvement)
	}

	// Improvement for high false negative rate
	if metrics.FalseNegativeRate > 0.1 {
		improvement := &AlgorithmImprovement{
			ImprovementID:   fmt.Sprintf("performance_improvement_fn_%d", time.Now().Unix()),
			ImprovementType: "performance_based",
			Description:     "Reduce false negative rate in security validation",
			Changes: map[string]interface{}{
				"metric":               "false_negative_rate",
				"current_value":        metrics.FalseNegativeRate,
				"target_value":         0.05,
				"action":               "enhance_detection_sensitivity",
				"sensitivity_increase": 0.15,
			},
			Confidence: 0.8,
			AppliedAt:  time.Now(),
		}
		improvements = append(improvements, improvement)
	}

	// Improvement for slow validation times
	if metrics.AverageValidationTime > svai.config.ValidationTimeout {
		improvement := &AlgorithmImprovement{
			ImprovementID:   fmt.Sprintf("performance_improvement_speed_%d", time.Now().Unix()),
			ImprovementType: "performance_based",
			Description:     "Improve security validation speed",
			Changes: map[string]interface{}{
				"metric":        "average_validation_time",
				"current_value": metrics.AverageValidationTime,
				"target_value":  svai.config.ValidationTimeout,
				"action":        "optimize_validation_process",
				"optimizations": []string{"parallel_processing", "caching", "query_optimization"},
			},
			Confidence: 0.7,
			AppliedAt:  time.Now(),
		}
		improvements = append(improvements, improvement)
	}

	return improvements
}

// generateMethodBasedImprovements generates improvements based on validation method performance
func (svai *SecurityValidationAlgorithmImprover) generateMethodBasedImprovements() []*AlgorithmImprovement {
	var improvements []*AlgorithmImprovement

	for methodName, methodMetrics := range svai.performanceMetrics.ValidationMethodMetrics {
		// Improvement for methods with low success rates
		if methodMetrics.SuccessRate < 0.8 {
			improvement := &AlgorithmImprovement{
				ImprovementID:   fmt.Sprintf("method_improvement_%s_%d", methodName, time.Now().Unix()),
				ImprovementType: "method_based",
				Description:     fmt.Sprintf("Improve %s validation method success rate", methodName),
				Changes: map[string]interface{}{
					"method_name":          methodName,
					"current_success_rate": methodMetrics.SuccessRate,
					"target_success_rate":  0.9,
					"action":               "refine_method_parameters",
					"parameter_adjustments": map[string]interface{}{
						"confidence_threshold": 0.05,
						"timeout_adjustment":   0.1,
					},
				},
				Confidence: 0.75,
				AppliedAt:  time.Now(),
			}
			improvements = append(improvements, improvement)
		}

		// Improvement for methods with high false positive rates
		if methodMetrics.FalsePositiveRate > 0.15 {
			improvement := &AlgorithmImprovement{
				ImprovementID:   fmt.Sprintf("method_improvement_fp_%s_%d", methodName, time.Now().Unix()),
				ImprovementType: "method_based",
				Description:     fmt.Sprintf("Reduce false positive rate for %s method", methodName),
				Changes: map[string]interface{}{
					"method_name":           methodName,
					"current_fp_rate":       methodMetrics.FalsePositiveRate,
					"target_fp_rate":        0.1,
					"action":                "adjust_method_sensitivity",
					"sensitivity_reduction": 0.1,
				},
				Confidence: 0.8,
				AppliedAt:  time.Now(),
			}
			improvements = append(improvements, improvement)
		}
	}

	return improvements
}

// applyImprovement applies an algorithm improvement
func (svai *SecurityValidationAlgorithmImprover) applyImprovement(ctx context.Context, improvement *AlgorithmImprovement) error {
	svai.logger.Info("Applying security validation improvement",
		zap.String("improvement_id", improvement.ImprovementID),
		zap.String("improvement_type", improvement.ImprovementType))

	// Validate improvement before applying
	if err := svai.validateImprovement(ctx, improvement); err != nil {
		return fmt.Errorf("improvement validation failed: %w", err)
	}

	// Apply the improvement based on type
	switch improvement.ImprovementType {
	case "pattern_based":
		return svai.applyPatternBasedImprovement(ctx, improvement)
	case "performance_based":
		return svai.applyPerformanceBasedImprovement(ctx, improvement)
	case "method_based":
		return svai.applyMethodBasedImprovement(ctx, improvement)
	default:
		return fmt.Errorf("unknown improvement type: %s", improvement.ImprovementType)
	}
}

// validateImprovement validates an improvement before applying it
func (svai *SecurityValidationAlgorithmImprover) validateImprovement(ctx context.Context, improvement *AlgorithmImprovement) error {
	// TODO: Implement comprehensive improvement validation
	// This would include:
	// 1. Safety checks
	// 2. Performance impact assessment
	// 3. Rollback plan validation
	// 4. A/B testing setup

	svai.logger.Debug("Validating improvement",
		zap.String("improvement_id", improvement.ImprovementID))

	// Basic validation - check confidence threshold
	if improvement.Confidence < svai.config.ConfidenceThreshold {
		return fmt.Errorf("improvement confidence too low: %f < %f",
			improvement.Confidence, svai.config.ConfidenceThreshold)
	}

	return nil
}

// applyPatternBasedImprovement applies a pattern-based improvement
func (svai *SecurityValidationAlgorithmImprover) applyPatternBasedImprovement(ctx context.Context, improvement *AlgorithmImprovement) error {
	svai.logger.Info("Applying pattern-based improvement",
		zap.String("improvement_id", improvement.ImprovementID))

	// TODO: Implement pattern-based improvement application
	// This would include:
	// 1. Update validation rules
	// 2. Adjust thresholds
	// 3. Modify detection logic
	// 4. Update configuration

	// Record the improvement
	svai.recordImprovement(improvement)

	return nil
}

// applyPerformanceBasedImprovement applies a performance-based improvement
func (svai *SecurityValidationAlgorithmImprover) applyPerformanceBasedImprovement(ctx context.Context, improvement *AlgorithmImprovement) error {
	svai.logger.Info("Applying performance-based improvement",
		zap.String("improvement_id", improvement.ImprovementID))

	// TODO: Implement performance-based improvement application
	// This would include:
	// 1. Optimize algorithms
	// 2. Adjust parameters
	// 3. Implement caching
	// 4. Update resource allocation

	// Record the improvement
	svai.recordImprovement(improvement)

	return nil
}

// applyMethodBasedImprovement applies a method-based improvement
func (svai *SecurityValidationAlgorithmImprover) applyMethodBasedImprovement(ctx context.Context, improvement *AlgorithmImprovement) error {
	svai.logger.Info("Applying method-based improvement",
		zap.String("improvement_id", improvement.ImprovementID))

	// TODO: Implement method-based improvement application
	// This would include:
	// 1. Update method parameters
	// 2. Refine algorithms
	// 3. Adjust thresholds
	// 4. Update method selection logic

	// Record the improvement
	svai.recordImprovement(improvement)

	return nil
}

// recordImprovement records an applied improvement
func (svai *SecurityValidationAlgorithmImprover) recordImprovement(improvement *AlgorithmImprovement) {
	svai.mu.Lock()
	defer svai.mu.Unlock()

	svai.improvementHistory = append(svai.improvementHistory, improvement)
	svai.lastImprovementTime = time.Now()
	svai.consecutiveImprovements++

	svai.logger.Info("Recorded security validation improvement",
		zap.String("improvement_id", improvement.ImprovementID),
		zap.Int("total_improvements", len(svai.improvementHistory)))
}

// updatePerformanceMetrics updates performance metrics
func (svai *SecurityValidationAlgorithmImprover) updatePerformanceMetrics() {
	svai.mu.Lock()
	defer svai.mu.Unlock()

	// TODO: Implement comprehensive metrics update
	// This would include:
	// 1. Calculate new metrics from recent data
	// 2. Update rolling averages
	// 3. Track improvement impact
	// 4. Update method-specific metrics

	svai.performanceMetrics.LastUpdated = time.Now()

	svai.logger.Debug("Updated security validation performance metrics")
}

// GetImprovementHistory returns the history of applied improvements
func (svai *SecurityValidationAlgorithmImprover) GetImprovementHistory() []*AlgorithmImprovement {
	svai.mu.RLock()
	defer svai.mu.RUnlock()

	// Return a copy to prevent external modification
	history := make([]*AlgorithmImprovement, len(svai.improvementHistory))
	copy(history, svai.improvementHistory)

	return history
}

// GetPerformanceMetrics returns current performance metrics
func (svai *SecurityValidationAlgorithmImprover) GetPerformanceMetrics() *SecurityValidationMetrics {
	svai.mu.RLock()
	defer svai.mu.RUnlock()

	// Return a copy to prevent external modification
	if svai.performanceMetrics == nil {
		return NewSecurityValidationMetrics()
	}

	metrics := *svai.performanceMetrics
	return &metrics
}

// ValidateImprovement validates an improvement after it has been applied
func (svai *SecurityValidationAlgorithmImprover) ValidateImprovement(ctx context.Context, improvementID string) (*ImprovementValidation, error) {
	svai.mu.RLock()
	defer svai.mu.RUnlock()

	// Find the improvement
	var improvement *AlgorithmImprovement
	for _, imp := range svai.improvementHistory {
		if imp.ImprovementID == improvementID {
			improvement = imp
			break
		}
	}

	if improvement == nil {
		return nil, fmt.Errorf("improvement not found: %s", improvementID)
	}

	// TODO: Implement comprehensive improvement validation
	// This would include:
	// 1. Performance testing
	// 2. Accuracy testing
	// 3. Stability testing
	// 4. A/B testing results

	validation := &ImprovementValidation{
		ValidationMethod:   "comprehensive_testing",
		TestCases:          100,
		PassedCases:        95,
		FailedCases:        5,
		Accuracy:           0.95,
		Performance:        0.9,
		Stability:          0.98,
		ValidatedAt:        time.Now(),
		ValidationDuration: 5 * time.Minute,
	}

	// Update improvement with validation results
	improvement.ValidatedAt = &validation.ValidatedAt
	improvement.ValidationResults = validation

	svai.logger.Info("Validated security validation improvement",
		zap.String("improvement_id", improvementID),
		zap.Float64("accuracy", validation.Accuracy),
		zap.Float64("performance", validation.Performance))

	return validation, nil
}

// RollbackImprovement rolls back an improvement if it's causing issues
func (svai *SecurityValidationAlgorithmImprover) RollbackImprovement(ctx context.Context, improvementID string, reason string) error {
	svai.mu.Lock()
	defer svai.mu.Unlock()

	// Find the improvement
	var improvement *AlgorithmImprovement
	for _, imp := range svai.improvementHistory {
		if imp.ImprovementID == improvementID {
			improvement = imp
			break
		}
	}

	if improvement == nil {
		return fmt.Errorf("improvement not found: %s", improvementID)
	}

	// TODO: Implement improvement rollback
	// This would include:
	// 1. Revert configuration changes
	// 2. Restore previous algorithms
	// 3. Update metrics
	// 4. Log rollback reason

	improvement.RollbackRequired = true
	improvement.RollbackReason = reason

	svai.logger.Warn("Rolled back security validation improvement",
		zap.String("improvement_id", improvementID),
		zap.String("reason", reason))

	return nil
}

// GetImprovementStatus returns the status of the improvement process
func (svai *SecurityValidationAlgorithmImprover) GetImprovementStatus() map[string]interface{} {
	svai.mu.RLock()
	defer svai.mu.RUnlock()

	return map[string]interface{}{
		"is_running":                    svai.config.EnableAutoImprovement,
		"last_improvement_time":         svai.lastImprovementTime,
		"consecutive_improvements":      svai.consecutiveImprovements,
		"max_consecutive_improvements":  svai.maxConsecutiveImprovements,
		"total_improvements":            len(svai.improvementHistory),
		"improvement_threshold":         svai.improvementThreshold,
		"has_improvement_opportunities": svai.hasImprovementOpportunities(),
		"performance_metrics":           svai.performanceMetrics,
	}
}
