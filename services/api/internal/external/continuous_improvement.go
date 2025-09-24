package external

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ContinuousImprovementManager manages automatic improvements based on failure analysis
type ContinuousImprovementManager struct {
	config     *ContinuousImprovementConfig
	logger     *zap.Logger
	monitor    *VerificationSuccessMonitor
	strategies map[string]*ImprovementStrategy
	mu         sync.RWMutex
	startTime  time.Time
}

// ContinuousImprovementConfig holds configuration for continuous improvement
type ContinuousImprovementConfig struct {
	EnableAutoImprovement      bool          `json:"enable_auto_improvement"`
	EnableStrategyOptimization bool          `json:"enable_strategy_optimization"`
	EnableThresholdAdjustment  bool          `json:"enable_threshold_adjustment"`
	EnableRetryOptimization    bool          `json:"enable_retry_optimization"`
	ImprovementInterval        time.Duration `json:"improvement_interval"`         // How often to run improvement analysis
	MinDataPointsForAnalysis   int           `json:"min_data_points_for_analysis"` // Minimum data points needed
	MaxImprovementHistory      int           `json:"max_improvement_history"`      // Maximum improvement history to keep
	ConfidenceThreshold        float64       `json:"confidence_threshold"`         // Confidence threshold for auto-improvements
	RollbackThreshold          float64       `json:"rollback_threshold"`           // Threshold for rolling back improvements
}

// ImprovementStrategy represents a specific improvement strategy
type ImprovementStrategy struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Type         string                 `json:"type"` // "strategy", "threshold", "retry", "custom"
	Parameters   map[string]interface{} `json:"parameters"`
	Confidence   float64                `json:"confidence"`
	Impact       float64                `json:"impact"` // Expected improvement in success rate
	Status       string                 `json:"status"` // "pending", "active", "paused", "rolled_back"
	CreatedAt    time.Time              `json:"created_at"`
	ActivatedAt  *time.Time             `json:"activated_at,omitempty"`
	RolledBackAt *time.Time             `json:"rolled_back_at,omitempty"`
	Metrics      *StrategyMetrics       `json:"metrics"`
}

// StrategyMetrics tracks performance metrics for a strategy
type StrategyMetrics struct {
	TotalAttempts       int64         `json:"total_attempts"`
	SuccessfulAttempts  int64         `json:"successful_attempts"`
	FailedAttempts      int64         `json:"failed_attempts"`
	SuccessRate         float64       `json:"success_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastUpdated         time.Time     `json:"last_updated"`
}

// ImprovementRecommendation represents a recommended improvement
type ImprovementRecommendation struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Priority    string                 `json:"priority"` // "high", "medium", "low"
	Description string                 `json:"description"`
	Impact      float64                `json:"impact"`
	Confidence  float64                `json:"confidence"`
	Parameters  map[string]interface{} `json:"parameters"`
	Reasoning   string                 `json:"reasoning"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ImprovementHistory tracks all improvements made
type ImprovementHistory struct {
	ID                string                 `json:"id"`
	StrategyID        string                 `json:"strategy_id"`
	Action            string                 `json:"action"` // "activate", "pause", "rollback", "modify"
	Parameters        map[string]interface{} `json:"parameters"`
	Reason            string                 `json:"reason"`
	SuccessRateBefore float64                `json:"success_rate_before"`
	SuccessRateAfter  float64                `json:"success_rate_after"`
	Timestamp         time.Time              `json:"timestamp"`
}

// NewContinuousImprovementManager creates a new continuous improvement manager
func NewContinuousImprovementManager(config *ContinuousImprovementConfig, monitor *VerificationSuccessMonitor, logger *zap.Logger) *ContinuousImprovementManager {
	if config == nil {
		config = DefaultContinuousImprovementConfig()
	}

	manager := &ContinuousImprovementManager{
		config:     config,
		logger:     logger,
		monitor:    monitor,
		strategies: make(map[string]*ImprovementStrategy),
		startTime:  time.Now(),
	}

	// Start background improvement analysis if enabled
	if config.EnableAutoImprovement {
		go manager.startBackgroundImprovement()
	}

	return manager
}

// DefaultContinuousImprovementConfig returns default configuration
func DefaultContinuousImprovementConfig() *ContinuousImprovementConfig {
	return &ContinuousImprovementConfig{
		EnableAutoImprovement:      true,
		EnableStrategyOptimization: true,
		EnableThresholdAdjustment:  true,
		EnableRetryOptimization:    true,
		ImprovementInterval:        1 * time.Hour,
		MinDataPointsForAnalysis:   100,
		MaxImprovementHistory:      1000,
		ConfidenceThreshold:        0.7,
		RollbackThreshold:          -0.05, // 5% decrease in success rate
	}
}

// AnalyzeAndRecommend analyzes failure patterns and generates improvement recommendations
func (m *ContinuousImprovementManager) AnalyzeAndRecommend(ctx context.Context) ([]*ImprovementRecommendation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var recommendations []*ImprovementRecommendation

	// Get current failure analysis
	failureAnalysis, err := m.monitor.AnalyzeFailures(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze failures: %w", err)
	}

	// Get current metrics
	metrics := m.monitor.GetMetrics()
	if metrics.TotalAttempts < int64(m.config.MinDataPointsForAnalysis) {
		return nil, fmt.Errorf("insufficient data points for analysis: %d < %d", metrics.TotalAttempts, m.config.MinDataPointsForAnalysis)
	}

	// Generate strategy optimization recommendations
	if m.config.EnableStrategyOptimization {
		strategyRecs := m.generateStrategyRecommendations(failureAnalysis, metrics)
		recommendations = append(recommendations, strategyRecs...)
	}

	// Generate threshold adjustment recommendations
	if m.config.EnableThresholdAdjustment {
		thresholdRecs := m.generateThresholdRecommendations(failureAnalysis, metrics)
		recommendations = append(recommendations, thresholdRecs...)
	}

	// Generate retry optimization recommendations
	if m.config.EnableRetryOptimization {
		retryRecs := m.generateRetryRecommendations(failureAnalysis, metrics)
		recommendations = append(recommendations, retryRecs...)
	}

	// Sort recommendations by impact and confidence
	sort.Slice(recommendations, func(i, j int) bool {
		scoreI := recommendations[i].Impact * recommendations[i].Confidence
		scoreJ := recommendations[j].Impact * recommendations[j].Confidence
		return scoreI > scoreJ
	})

	return recommendations, nil
}

// ApplyImprovement applies a specific improvement strategy
func (m *ContinuousImprovementManager) ApplyImprovement(ctx context.Context, recommendation *ImprovementRecommendation) (*ImprovementStrategy, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create improvement strategy
	strategy := &ImprovementStrategy{
		ID:          generateStrategyID(),
		Name:        recommendation.Description,
		Description: recommendation.Description,
		Type:        recommendation.Type,
		Parameters:  recommendation.Parameters,
		Confidence:  recommendation.Confidence,
		Impact:      recommendation.Impact,
		Status:      "pending",
		CreatedAt:   time.Now(),
		Metrics:     &StrategyMetrics{},
	}

	// Apply the improvement based on type
	switch recommendation.Type {
	case "strategy":
		err := m.applyStrategyImprovement(strategy)
		if err != nil {
			return nil, fmt.Errorf("failed to apply strategy improvement: %w", err)
		}
	case "threshold":
		err := m.applyThresholdImprovement(strategy)
		if err != nil {
			return nil, fmt.Errorf("failed to apply threshold improvement: %w", err)
		}
	case "retry":
		err := m.applyRetryImprovement(strategy)
		if err != nil {
			return nil, fmt.Errorf("failed to apply retry improvement: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown improvement type: %s", recommendation.Type)
	}

	// Store the strategy
	m.strategies[strategy.ID] = strategy

	// Log the improvement
	m.logger.Info("Applied improvement strategy",
		zap.String("strategy_id", strategy.ID),
		zap.String("type", strategy.Type),
		zap.Float64("confidence", strategy.Confidence),
		zap.Float64("expected_impact", strategy.Impact))

	return strategy, nil
}

// EvaluateStrategy evaluates the performance of an active strategy
func (m *ContinuousImprovementManager) EvaluateStrategy(ctx context.Context, strategyID string) (*StrategyEvaluation, error) {
	m.mu.RLock()
	strategy, exists := m.strategies[strategyID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("strategy not found: %s", strategyID)
	}

	if strategy.Status != "active" {
		return nil, fmt.Errorf("strategy is not active: %s", strategy.Status)
	}

	// Get current metrics
	currentMetrics := m.monitor.GetMetrics()

	// Calculate improvement
	improvement := currentMetrics.SuccessRate - strategy.Metrics.SuccessRate

	evaluation := &StrategyEvaluation{
		StrategyID:        strategyID,
		SuccessRateBefore: strategy.Metrics.SuccessRate,
		SuccessRateAfter:  currentMetrics.SuccessRate,
		Improvement:       improvement,
		IsBeneficial:      improvement > 0,
		ShouldRollback:    improvement < m.config.RollbackThreshold,
		EvaluatedAt:       time.Now(),
	}

	// Update strategy metrics
	strategy.Metrics = &StrategyMetrics{
		TotalAttempts:       currentMetrics.TotalAttempts,
		SuccessfulAttempts:  currentMetrics.SuccessfulAttempts,
		FailedAttempts:      currentMetrics.FailedAttempts,
		SuccessRate:         currentMetrics.SuccessRate,
		AverageResponseTime: currentMetrics.AverageResponseTime,
		LastUpdated:         time.Now(),
	}

	return evaluation, nil
}

// RollbackStrategy rolls back a strategy if it's not performing well
func (m *ContinuousImprovementManager) RollbackStrategy(ctx context.Context, strategyID string, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	strategy, exists := m.strategies[strategyID]
	if !exists {
		return fmt.Errorf("strategy not found: %s", strategyID)
	}

	if strategy.Status != "active" {
		return fmt.Errorf("strategy is not active: %s", strategy.Status)
	}

	// Rollback based on strategy type
	switch strategy.Type {
	case "strategy":
		err := m.rollbackStrategyImprovement(strategy)
		if err != nil {
			return fmt.Errorf("failed to rollback strategy improvement: %w", err)
		}
	case "threshold":
		err := m.rollbackThresholdImprovement(strategy)
		if err != nil {
			return fmt.Errorf("failed to rollback threshold improvement: %w", err)
		}
	case "retry":
		err := m.rollbackRetryImprovement(strategy)
		if err != nil {
			return fmt.Errorf("failed to rollback retry improvement: %w", err)
		}
	}

	// Update strategy status
	strategy.Status = "rolled_back"
	now := time.Now()
	strategy.RolledBackAt = &now

	// Log the rollback
	m.logger.Warn("Rolled back improvement strategy",
		zap.String("strategy_id", strategyID),
		zap.String("reason", reason))

	return nil
}

// GetActiveStrategies returns all active improvement strategies
func (m *ContinuousImprovementManager) GetActiveStrategies() []*ImprovementStrategy {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var activeStrategies []*ImprovementStrategy
	for _, strategy := range m.strategies {
		if strategy.Status == "active" {
			activeStrategies = append(activeStrategies, strategy)
		}
	}

	return activeStrategies
}

// GetImprovementHistory returns the improvement history
func (m *ContinuousImprovementManager) GetImprovementHistory() []*ImprovementHistory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// This would typically come from a database
	// For now, return empty slice
	return []*ImprovementHistory{}
}

// GetConfig returns the current configuration
func (m *ContinuousImprovementManager) GetConfig() *ContinuousImprovementConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// UpdateConfig updates the configuration
func (m *ContinuousImprovementManager) UpdateConfig(config *ContinuousImprovementConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate configuration
	if config.ConfidenceThreshold < 0 || config.ConfidenceThreshold > 1 {
		return fmt.Errorf("confidence threshold must be between 0 and 1")
	}

	if config.ImprovementInterval < time.Minute {
		return fmt.Errorf("improvement interval must be at least 1 minute")
	}

	m.config = config
	return nil
}

// generateStrategyRecommendations generates recommendations for strategy optimization
func (m *ContinuousImprovementManager) generateStrategyRecommendations(analysis *FailureAnalysis, metrics *SuccessMetrics) []*ImprovementRecommendation {
	var recommendations []*ImprovementRecommendation

	// Analyze strategy failures
	for strategy, failures := range analysis.StrategyFailures {
		if failures > 10 { // High failure threshold
			recommendations = append(recommendations, &ImprovementRecommendation{
				ID:          generateRecommendationID(),
				Type:        "strategy",
				Priority:    "high",
				Description: fmt.Sprintf("Optimize strategy '%s'", strategy),
				Impact:      0.05, // Estimated 5% improvement
				Confidence:  0.8,
				Parameters: map[string]interface{}{
					"strategy_name": strategy,
					"failure_count": failures,
					"action":        "optimize",
				},
				Reasoning: fmt.Sprintf("Strategy '%s' has %d failures, indicating optimization needed", strategy, failures),
				CreatedAt: time.Now(),
			})
		}
	}

	// Analyze error types for strategy improvements
	for errorType, count := range analysis.CommonErrorTypes {
		if count > 20 { // High error threshold
			recommendations = append(recommendations, &ImprovementRecommendation{
				ID:          generateRecommendationID(),
				Type:        "strategy",
				Priority:    "medium",
				Description: fmt.Sprintf("Add fallback strategy for '%s' errors", errorType),
				Impact:      0.03, // Estimated 3% improvement
				Confidence:  0.7,
				Parameters: map[string]interface{}{
					"error_type":  errorType,
					"error_count": count,
					"action":      "add_fallback",
				},
				Reasoning: fmt.Sprintf("Error type '%s' occurs %d times, suggesting need for fallback strategy", errorType, count),
				CreatedAt: time.Now(),
			})
		}
	}

	return recommendations
}

// generateThresholdRecommendations generates recommendations for threshold adjustments
func (m *ContinuousImprovementManager) generateThresholdRecommendations(analysis *FailureAnalysis, metrics *SuccessMetrics) []*ImprovementRecommendation {
	var recommendations []*ImprovementRecommendation

	// If success rate is below target, suggest threshold adjustments
	if metrics.SuccessRate < 0.90 { // Target success rate
		recommendations = append(recommendations, &ImprovementRecommendation{
			ID:          generateRecommendationID(),
			Type:        "threshold",
			Priority:    "high",
			Description: "Adjust verification thresholds for better success rate",
			Impact:      0.02, // Estimated 2% improvement
			Confidence:  0.6,
			Parameters: map[string]interface{}{
				"current_rate": metrics.SuccessRate,
				"target_rate":  0.90,
				"action":       "lower_thresholds",
			},
			Reasoning: fmt.Sprintf("Current success rate %.2f%% is below target 90%%, suggesting threshold adjustment needed", metrics.SuccessRate*100),
			CreatedAt: time.Now(),
		})
	}

	return recommendations
}

// generateRetryRecommendations generates recommendations for retry optimization
func (m *ContinuousImprovementManager) generateRetryRecommendations(analysis *FailureAnalysis, metrics *SuccessMetrics) []*ImprovementRecommendation {
	var recommendations []*ImprovementRecommendation

	// Analyze timeout errors for retry optimization
	if timeoutCount, exists := analysis.CommonErrorTypes["timeout"]; exists && timeoutCount > 15 {
		recommendations = append(recommendations, &ImprovementRecommendation{
			ID:          generateRecommendationID(),
			Type:        "retry",
			Priority:    "medium",
			Description: "Optimize retry strategy for timeout errors",
			Impact:      0.04, // Estimated 4% improvement
			Confidence:  0.75,
			Parameters: map[string]interface{}{
				"error_type":  "timeout",
				"error_count": timeoutCount,
				"action":      "increase_retries",
			},
			Reasoning: fmt.Sprintf("Timeout errors occur %d times, suggesting retry strategy optimization needed", timeoutCount),
			CreatedAt: time.Now(),
		})
	}

	return recommendations
}

// applyStrategyImprovement applies a strategy-based improvement
func (m *ContinuousImprovementManager) applyStrategyImprovement(strategy *ImprovementStrategy) error {
	// This would typically modify the verification strategy configuration
	// For now, just mark as active
	strategy.Status = "active"
	now := time.Now()
	strategy.ActivatedAt = &now

	m.logger.Info("Applied strategy improvement",
		zap.String("strategy_id", strategy.ID),
		zap.String("strategy_name", strategy.Name))

	return nil
}

// applyThresholdImprovement applies a threshold-based improvement
func (m *ContinuousImprovementManager) applyThresholdImprovement(strategy *ImprovementStrategy) error {
	// This would typically modify verification thresholds
	// For now, just mark as active
	strategy.Status = "active"
	now := time.Now()
	strategy.ActivatedAt = &now

	m.logger.Info("Applied threshold improvement",
		zap.String("strategy_id", strategy.ID),
		zap.String("strategy_name", strategy.Name))

	return nil
}

// applyRetryImprovement applies a retry-based improvement
func (m *ContinuousImprovementManager) applyRetryImprovement(strategy *ImprovementStrategy) error {
	// This would typically modify retry configuration
	// For now, just mark as active
	strategy.Status = "active"
	now := time.Now()
	strategy.ActivatedAt = &now

	m.logger.Info("Applied retry improvement",
		zap.String("strategy_id", strategy.ID),
		zap.String("strategy_name", strategy.Name))

	return nil
}

// rollbackStrategyImprovement rolls back a strategy-based improvement
func (m *ContinuousImprovementManager) rollbackStrategyImprovement(strategy *ImprovementStrategy) error {
	// This would typically restore the original strategy configuration
	m.logger.Info("Rolled back strategy improvement",
		zap.String("strategy_id", strategy.ID),
		zap.String("strategy_name", strategy.Name))

	return nil
}

// rollbackThresholdImprovement rolls back a threshold-based improvement
func (m *ContinuousImprovementManager) rollbackThresholdImprovement(strategy *ImprovementStrategy) error {
	// This would typically restore the original thresholds
	m.logger.Info("Rolled back threshold improvement",
		zap.String("strategy_id", strategy.ID),
		zap.String("strategy_name", strategy.Name))

	return nil
}

// rollbackRetryImprovement rolls back a retry-based improvement
func (m *ContinuousImprovementManager) rollbackRetryImprovement(strategy *ImprovementStrategy) error {
	// This would typically restore the original retry configuration
	m.logger.Info("Rolled back retry improvement",
		zap.String("strategy_id", strategy.ID),
		zap.String("strategy_name", strategy.Name))

	return nil
}

// startBackgroundImprovement starts the background improvement analysis
func (m *ContinuousImprovementManager) startBackgroundImprovement() {
	ticker := time.NewTicker(m.config.ImprovementInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()

			// Generate recommendations
			recommendations, err := m.AnalyzeAndRecommend(ctx)
			if err != nil {
				m.logger.Error("Failed to generate improvement recommendations", zap.Error(err))
				continue
			}

			// Apply high-confidence recommendations automatically
			for _, rec := range recommendations {
				if rec.Confidence >= m.config.ConfidenceThreshold && rec.Priority == "high" {
					strategy, err := m.ApplyImprovement(ctx, rec)
					if err != nil {
						m.logger.Error("Failed to apply improvement",
							zap.String("recommendation_id", rec.ID),
							zap.Error(err))
					} else {
						m.logger.Info("Auto-applied improvement",
							zap.String("strategy_id", strategy.ID),
							zap.String("recommendation_id", rec.ID),
							zap.Float64("confidence", rec.Confidence))
					}
				}
			}

			// Evaluate active strategies
			activeStrategies := m.GetActiveStrategies()
			for _, strategy := range activeStrategies {
				evaluation, err := m.EvaluateStrategy(ctx, strategy.ID)
				if err != nil {
					m.logger.Error("Failed to evaluate strategy",
						zap.String("strategy_id", strategy.ID),
						zap.Error(err))
					continue
				}

				// Rollback if not performing well
				if evaluation.ShouldRollback {
					err := m.RollbackStrategy(ctx, strategy.ID, "Poor performance")
					if err != nil {
						m.logger.Error("Failed to rollback strategy",
							zap.String("strategy_id", strategy.ID),
							zap.Error(err))
					}
				}
			}
		}
	}
}

// StrategyEvaluation represents the evaluation of a strategy's performance
type StrategyEvaluation struct {
	StrategyID        string    `json:"strategy_id"`
	SuccessRateBefore float64   `json:"success_rate_before"`
	SuccessRateAfter  float64   `json:"success_rate_after"`
	Improvement       float64   `json:"improvement"`
	IsBeneficial      bool      `json:"is_beneficial"`
	ShouldRollback    bool      `json:"should_rollback"`
	EvaluatedAt       time.Time `json:"evaluated_at"`
}

// Helper functions
func generateStrategyID() string {
	return fmt.Sprintf("strategy_%d", time.Now().UnixNano())
}

func generateRecommendationID() string {
	return fmt.Sprintf("rec_%d", time.Now().UnixNano())
}
