package industry_codes

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RecoveryStrategy defines the type of recovery action to take
type RecoveryStrategy string

const (
	StrategyRetry               RecoveryStrategy = "retry"
	StrategyFallback            RecoveryStrategy = "fallback"
	StrategyCircuitBreaker      RecoveryStrategy = "circuit_breaker"
	StrategyManualIntervention  RecoveryStrategy = "manual_intervention"
	StrategyGracefulDegradation RecoveryStrategy = "graceful_degradation"
	StrategyRollback            RecoveryStrategy = "rollback"
	StrategyCompensation        RecoveryStrategy = "compensation"
	StrategyTimeout             RecoveryStrategy = "timeout"
	StrategyResourceCleanup     RecoveryStrategy = "resource_cleanup"
	StrategyDataRecovery        RecoveryStrategy = "data_recovery"
)

// RecoveryAction represents a specific recovery action to be executed
type RecoveryAction struct {
	ID          string                 `json:"id"`
	Strategy    RecoveryStrategy       `json:"strategy"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Priority    int                    `json:"priority"`
	Timeout     time.Duration          `json:"timeout"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	Backoff     time.Duration          `json:"backoff"`
}

// RecoveryResult represents the outcome of a recovery action
type RecoveryResult struct {
	ActionID     string                 `json:"action_id"`
	Strategy     RecoveryStrategy       `json:"strategy"`
	Success      bool                   `json:"success"`
	Error        error                  `json:"error,omitempty"`
	Duration     time.Duration          `json:"duration"`
	Attempts     int                    `json:"attempts"`
	Metadata     map[string]interface{} `json:"metadata"`
	RecoveryData interface{}            `json:"recovery_data,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// RecoveryPlan represents a complete recovery plan for an error
type RecoveryPlan struct {
	ID                 string           `json:"id"`
	ErrorID            string           `json:"error_id"`
	Category           ErrorCategory    `json:"category"`
	Severity           ErrorSeverity    `json:"severity"`
	Priority           int              `json:"priority"`
	Actions            []RecoveryAction `json:"actions"`
	EstimatedTime      time.Duration    `json:"estimated_time"`
	SuccessProbability float64          `json:"success_probability"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

// RecoveryExecution represents the execution of a recovery plan
type RecoveryExecution struct {
	PlanID    string                 `json:"plan_id"`
	Status    ExecutionStatus        `json:"status"`
	Results   []RecoveryResult       `json:"results"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time,omitempty"`
	Duration  time.Duration          `json:"duration,omitempty"`
	Error     error                  `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ExecutionStatus represents the status of a recovery execution
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusCancelled ExecutionStatus = "cancelled"
	StatusTimeout   ExecutionStatus = "timeout"
)

// RecoveryConfig defines configuration for the error recovery service
type RecoveryConfig struct {
	MaxConcurrentRecoveries     int           `json:"max_concurrent_recoveries"`
	DefaultTimeout              time.Duration `json:"default_timeout"`
	MaxRetryAttempts            int           `json:"max_retry_attempts"`
	RetryBackoffMultiplier      float64       `json:"retry_backoff_multiplier"`
	CircuitBreakerThreshold     int           `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout       time.Duration `json:"circuit_breaker_timeout"`
	EnableAutoRecovery          bool          `json:"enable_auto_recovery"`
	ManualInterventionThreshold ErrorSeverity `json:"manual_intervention_threshold"`
	RecoveryHistorySize         int           `json:"recovery_history_size"`
}

// RecoveryStats tracks recovery performance metrics
type RecoveryStats struct {
	TotalRecoveries      int64                               `json:"total_recoveries"`
	SuccessfulRecoveries int64                               `json:"successful_recoveries"`
	FailedRecoveries     int64                               `json:"failed_recoveries"`
	AverageRecoveryTime  time.Duration                       `json:"average_recovery_time"`
	TotalRecoveryTime    time.Duration                       `json:"total_recovery_time"`
	StrategyStats        map[RecoveryStrategy]*StrategyStats `json:"strategy_stats"`
	CategoryStats        map[ErrorCategory]*CategoryStats    `json:"category_stats"`
	LastUpdated          time.Time                           `json:"last_updated"`
	mu                   sync.RWMutex                        `json:"-"`
}

// StrategyStats tracks performance metrics for each recovery strategy
type StrategyStats struct {
	UsageCount      int64         `json:"usage_count"`
	SuccessCount    int64         `json:"success_count"`
	FailureCount    int64         `json:"failure_count"`
	AverageDuration time.Duration `json:"average_duration"`
	TotalDuration   time.Duration `json:"total_duration"`
	SuccessRate     float64       `json:"success_rate"`
}

// CategoryStats tracks performance metrics for each error category
type CategoryStats struct {
	RecoveryCount         int64            `json:"recovery_count"`
	SuccessCount          int64            `json:"success_count"`
	FailureCount          int64            `json:"failure_count"`
	AverageRecoveryTime   time.Duration    `json:"average_recovery_time"`
	TotalRecoveryTime     time.Duration    `json:"total_recovery_time"`
	SuccessRate           float64          `json:"success_rate"`
	MostEffectiveStrategy RecoveryStrategy `json:"most_effective_strategy"`
}

// ErrorRecoveryService provides comprehensive error recovery capabilities
type ErrorRecoveryService struct {
	config          *RecoveryConfig
	logger          *zap.Logger
	categorizer     *ErrorCategorizer
	stats           *RecoveryStats
	executions      map[string]*RecoveryExecution
	recoveryHistory []*RecoveryExecution
	mu              sync.RWMutex
}

// NewErrorRecoveryService creates a new error recovery service
func NewErrorRecoveryService(config *RecoveryConfig, categorizer *ErrorCategorizer, logger *zap.Logger) *ErrorRecoveryService {
	if config == nil {
		config = &RecoveryConfig{
			MaxConcurrentRecoveries:     10,
			DefaultTimeout:              30 * time.Second,
			MaxRetryAttempts:            3,
			RetryBackoffMultiplier:      2.0,
			CircuitBreakerThreshold:     5,
			CircuitBreakerTimeout:       60 * time.Second,
			EnableAutoRecovery:          true,
			ManualInterventionThreshold: SeverityHigh,
			RecoveryHistorySize:         1000,
		}
	}

	return &ErrorRecoveryService{
		config:      config,
		logger:      logger,
		categorizer: categorizer,
		stats: &RecoveryStats{
			StrategyStats: make(map[RecoveryStrategy]*StrategyStats),
			CategoryStats: make(map[ErrorCategory]*CategoryStats),
		},
		executions:      make(map[string]*RecoveryExecution),
		recoveryHistory: make([]*RecoveryExecution, 0),
	}
}

// CreateRecoveryPlan creates a recovery plan based on error categorization
func (r *ErrorRecoveryService) CreateRecoveryPlan(ctx context.Context, err error, context map[string]interface{}) (*RecoveryPlan, error) {
	r.logger.Info("Creating recovery plan",
		zap.String("error_message", err.Error()))

	// Categorize the error to get recommendations
	categorization := r.categorizer.CategorizeError(ctx, err, context)
	if categorization == nil {
		r.logger.Error("Categorization returned nil",
			zap.String("error_message", err.Error()))
		return nil, fmt.Errorf("failed to categorize error")
	}

	// Generate recovery actions based on categorization
	actions := r.generateRecoveryActions(categorization, err)

	// Calculate estimated time and success probability
	estimatedTime := r.calculateEstimatedTime(actions)
	successProbability := r.calculateSuccessProbability(categorization, actions)

	plan := &RecoveryPlan{
		ID:                 generateRecoveryPlanID(),
		ErrorID:            categorization.ID,
		Category:           categorization.Category,
		Severity:           categorization.Severity,
		Priority:           r.convertPriorityToInt(categorization.Priority),
		Actions:            actions,
		EstimatedTime:      estimatedTime,
		SuccessProbability: successProbability,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	r.logger.Info("Recovery plan created",
		zap.String("plan_id", plan.ID),
		zap.Int("action_count", len(actions)),
		zap.Duration("estimated_time", estimatedTime),
		zap.Float64("success_probability", successProbability))

	return plan, nil
}

// ExecuteRecoveryPlan executes a recovery plan
func (r *ErrorRecoveryService) ExecuteRecoveryPlan(ctx context.Context, plan *RecoveryPlan) (*RecoveryExecution, error) {
	r.logger.Info("Executing recovery plan",
		zap.String("plan_id", plan.ID),
		zap.Int("action_count", len(plan.Actions)))

	execution := &RecoveryExecution{
		PlanID:    plan.ID,
		Status:    StatusRunning,
		Results:   make([]RecoveryResult, 0),
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	r.mu.Lock()
	r.executions[plan.ID] = execution
	r.mu.Unlock()

	defer func() {
		execution.EndTime = time.Now()
		execution.Duration = execution.EndTime.Sub(execution.StartTime)
		r.updateStats(execution)
		r.addToHistory(execution)
	}()

	// Execute actions in priority order
	for _, action := range plan.Actions {
		select {
		case <-ctx.Done():
			execution.Status = StatusCancelled
			execution.Error = ctx.Err()
			return execution, ctx.Err()
		default:
		}

		r.logger.Debug("Executing recovery action",
			zap.String("plan_id", plan.ID),
			zap.String("action_id", action.ID),
			zap.String("strategy", string(action.Strategy)))

		result := r.executeAction(ctx, action)
		execution.Results = append(execution.Results, result)

		// If action succeeds and it's a terminal action, stop execution
		if result.Success && r.isTerminalAction(action.Strategy) {
			execution.Status = StatusCompleted
			r.logger.Info("Recovery plan completed successfully",
				zap.String("plan_id", plan.ID),
				zap.String("final_action", action.ID))
			return execution, nil
		}

		// If action fails and it's critical, mark execution as failed
		if !result.Success && r.isCriticalAction(action.Strategy) {
			execution.Status = StatusFailed
			execution.Error = result.Error
			r.logger.Error("Recovery plan failed on critical action",
				zap.String("plan_id", plan.ID),
				zap.String("action_id", action.ID),
				zap.Error(result.Error))
			return execution, result.Error
		}
	}

	// Check if any action succeeded
	anySuccess := false
	for _, result := range execution.Results {
		if result.Success {
			anySuccess = true
			break
		}
	}

	if anySuccess {
		execution.Status = StatusCompleted
		r.logger.Info("Recovery plan completed with partial success",
			zap.String("plan_id", plan.ID))
	} else {
		execution.Status = StatusFailed
		execution.Error = fmt.Errorf("all recovery actions failed")
		r.logger.Error("Recovery plan failed - all actions failed",
			zap.String("plan_id", plan.ID))
	}

	return execution, execution.Error
}

// AutoRecover automatically creates and executes a recovery plan
func (r *ErrorRecoveryService) AutoRecover(ctx context.Context, err error, context map[string]interface{}) (*RecoveryExecution, error) {
	if !r.config.EnableAutoRecovery {
		return nil, fmt.Errorf("auto recovery is disabled")
	}

	// First categorize the error to check severity
	categorization := r.categorizer.CategorizeError(ctx, err, context)
	if categorization == nil {
		return nil, fmt.Errorf("failed to categorize error")
	}

	// Check if manual intervention is required
	if r.isSeverityGreaterOrEqual(categorization.Severity, r.config.ManualInterventionThreshold) {
		r.logger.Warn("Manual intervention required for high severity error",
			zap.String("error_id", categorization.ID),
			zap.String("severity", string(categorization.Severity)))
		return nil, fmt.Errorf("manual intervention required for severity %s", categorization.Severity)
	}

	plan, err := r.CreateRecoveryPlan(ctx, err, context)
	if err != nil {
		return nil, fmt.Errorf("failed to create recovery plan: %w", err)
	}

	return r.ExecuteRecoveryPlan(ctx, plan)
}

// GetRecoveryStats returns current recovery statistics
func (r *ErrorRecoveryService) GetRecoveryStats() *RecoveryStats {
	r.stats.mu.RLock()
	defer r.stats.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := &RecoveryStats{
		TotalRecoveries:      r.stats.TotalRecoveries,
		SuccessfulRecoveries: r.stats.SuccessfulRecoveries,
		FailedRecoveries:     r.stats.FailedRecoveries,
		AverageRecoveryTime:  r.stats.AverageRecoveryTime,
		TotalRecoveryTime:    r.stats.TotalRecoveryTime,
		LastUpdated:          r.stats.LastUpdated,
		StrategyStats:        make(map[RecoveryStrategy]*StrategyStats),
		CategoryStats:        make(map[ErrorCategory]*CategoryStats),
	}

	// Copy strategy stats
	for strategy, stat := range r.stats.StrategyStats {
		stats.StrategyStats[strategy] = &StrategyStats{
			UsageCount:      stat.UsageCount,
			SuccessCount:    stat.SuccessCount,
			FailureCount:    stat.FailureCount,
			AverageDuration: stat.AverageDuration,
			TotalDuration:   stat.TotalDuration,
			SuccessRate:     stat.SuccessRate,
		}
	}

	// Copy category stats
	for category, stat := range r.stats.CategoryStats {
		stats.CategoryStats[category] = &CategoryStats{
			RecoveryCount:         stat.RecoveryCount,
			SuccessCount:          stat.SuccessCount,
			FailureCount:          stat.FailureCount,
			AverageRecoveryTime:   stat.AverageRecoveryTime,
			TotalRecoveryTime:     stat.TotalRecoveryTime,
			SuccessRate:           stat.SuccessRate,
			MostEffectiveStrategy: stat.MostEffectiveStrategy,
		}
	}

	return stats
}

// GetRecoveryHistory returns recent recovery executions
func (r *ErrorRecoveryService) GetRecoveryHistory(limit int) []*RecoveryExecution {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 || limit > len(r.recoveryHistory) {
		limit = len(r.recoveryHistory)
	}

	history := make([]*RecoveryExecution, limit)
	copy(history, r.recoveryHistory[len(r.recoveryHistory)-limit:])
	return history
}

// generateRecoveryActions generates recovery actions based on error categorization
func (r *ErrorRecoveryService) generateRecoveryActions(categorization *CategorizedError, err error) []RecoveryAction {
	var actions []RecoveryAction

	// Add actions based on error category
	switch categorization.Category {
	case CategoryNetwork:
		actions = append(actions, RecoveryAction{
			ID:          generateActionID(),
			Strategy:    StrategyRetry,
			Description: "Retry network operation with exponential backoff",
			Parameters: map[string]interface{}{
				"max_retries":        r.config.MaxRetryAttempts,
				"backoff_multiplier": r.config.RetryBackoffMultiplier,
			},
			Priority:   1,
			Timeout:    r.config.DefaultTimeout,
			MaxRetries: r.config.MaxRetryAttempts,
			Backoff:    1 * time.Second,
		})
		actions = append(actions, RecoveryAction{
			ID:          generateActionID(),
			Strategy:    StrategyFallback,
			Description: "Use cached data as fallback",
			Parameters:  map[string]interface{}{},
			Priority:    2,
			Timeout:     r.config.DefaultTimeout,
		})

	case CategoryDatabase:
		actions = append(actions, RecoveryAction{
			ID:          generateActionID(),
			Strategy:    StrategyRetry,
			Description: "Retry database operation",
			Parameters: map[string]interface{}{
				"max_retries": r.config.MaxRetryAttempts,
			},
			Priority:   1,
			Timeout:    r.config.DefaultTimeout,
			MaxRetries: r.config.MaxRetryAttempts,
		})
		actions = append(actions, RecoveryAction{
			ID:          generateActionID(),
			Strategy:    StrategyCircuitBreaker,
			Description: "Activate circuit breaker for database operations",
			Parameters: map[string]interface{}{
				"threshold": r.config.CircuitBreakerThreshold,
				"timeout":   r.config.CircuitBreakerTimeout,
			},
			Priority: 2,
			Timeout:  r.config.DefaultTimeout,
		})

	case CategoryValidation:
		actions = append(actions, RecoveryAction{
			ID:          generateActionID(),
			Strategy:    StrategyDataRecovery,
			Description: "Attempt to recover and validate data",
			Parameters:  map[string]interface{}{},
			Priority:    1,
			Timeout:     r.config.DefaultTimeout,
		})

	case CategorySecurity:
		actions = append(actions, RecoveryAction{
			ID:          generateActionID(),
			Strategy:    StrategyManualIntervention,
			Description: "Security error requires manual intervention",
			Parameters:  map[string]interface{}{},
			Priority:    1,
			Timeout:     r.config.DefaultTimeout,
		})

	case CategoryPerformance:
		actions = append(actions, RecoveryAction{
			ID:          generateActionID(),
			Strategy:    StrategyTimeout,
			Description: "Increase timeout for performance issues",
			Parameters: map[string]interface{}{
				"timeout_multiplier": 2.0,
			},
			Priority: 1,
			Timeout:  r.config.DefaultTimeout * 2,
		})
		actions = append(actions, RecoveryAction{
			ID:          generateActionID(),
			Strategy:    StrategyGracefulDegradation,
			Description: "Enable graceful degradation mode",
			Parameters:  map[string]interface{}{},
			Priority:    2,
			Timeout:     r.config.DefaultTimeout,
		})

	default:
		// Default actions for unknown categories
		actions = append(actions, RecoveryAction{
			ID:          generateActionID(),
			Strategy:    StrategyRetry,
			Description: "Generic retry mechanism",
			Parameters: map[string]interface{}{
				"max_retries": r.config.MaxRetryAttempts,
			},
			Priority:   1,
			Timeout:    r.config.DefaultTimeout,
			MaxRetries: r.config.MaxRetryAttempts,
		})
	}

	// Add resource cleanup action if needed
	if r.requiresResourceCleanup(err) {
		actions = append(actions, RecoveryAction{
			ID:          generateActionID(),
			Strategy:    StrategyResourceCleanup,
			Description: "Clean up resources to prevent leaks",
			Parameters:  map[string]interface{}{},
			Priority:    10, // High priority for cleanup
			Timeout:     r.config.DefaultTimeout,
		})
	}

	return actions
}

// executeAction executes a single recovery action
func (r *ErrorRecoveryService) executeAction(ctx context.Context, action RecoveryAction) RecoveryResult {
	startTime := time.Now()
	result := RecoveryResult{
		ActionID:  action.ID,
		Strategy:  action.Strategy,
		Metadata:  make(map[string]interface{}),
		Timestamp: startTime,
	}

	r.logger.Debug("Executing recovery action",
		zap.String("action_id", action.ID),
		zap.String("strategy", string(action.Strategy)))

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, action.Timeout)
	defer cancel()

	// Execute based on strategy
	switch action.Strategy {
	case StrategyRetry:
		result = r.executeRetryAction(timeoutCtx, action)
	case StrategyFallback:
		result = r.executeFallbackAction(timeoutCtx, action)
	case StrategyCircuitBreaker:
		result = r.executeCircuitBreakerAction(timeoutCtx, action)
	case StrategyManualIntervention:
		result = r.executeManualInterventionAction(timeoutCtx, action)
	case StrategyGracefulDegradation:
		result = r.executeGracefulDegradationAction(timeoutCtx, action)
	case StrategyRollback:
		result = r.executeRollbackAction(timeoutCtx, action)
	case StrategyCompensation:
		result = r.executeCompensationAction(timeoutCtx, action)
	case StrategyTimeout:
		result = r.executeTimeoutAction(timeoutCtx, action)
	case StrategyResourceCleanup:
		result = r.executeResourceCleanupAction(timeoutCtx, action)
	case StrategyDataRecovery:
		result = r.executeDataRecoveryAction(timeoutCtx, action)
	default:
		result.Success = false
		result.Error = fmt.Errorf("unknown recovery strategy: %s", action.Strategy)
	}

	result.Duration = time.Since(startTime)

	r.logger.Debug("Recovery action completed",
		zap.String("action_id", action.ID),
		zap.Bool("success", result.Success),
		zap.Duration("duration", result.Duration),
		zap.Error(result.Error))

	return result
}

// executeRetryAction implements retry strategy
func (r *ErrorRecoveryService) executeRetryAction(ctx context.Context, action RecoveryAction) RecoveryResult {
	result := RecoveryResult{
		ActionID:  action.ID,
		Strategy:  action.Strategy,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	maxRetries := action.MaxRetries
	if maxRetries == 0 {
		maxRetries = r.config.MaxRetryAttempts
	}

	backoff := action.Backoff
	if backoff == 0 {
		backoff = 1 * time.Second
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			result.Success = false
			result.Error = ctx.Err()
			result.Attempts = attempt
			return result
		default:
		}

		// Simulate retry logic (in real implementation, this would call the actual operation)
		if r.simulateRetrySuccess(attempt, maxRetries) {
			result.Success = true
			result.Attempts = attempt + 1
			result.Metadata["attempts"] = attempt + 1
			result.Metadata["backoff_used"] = backoff
			return result
		}

		if attempt < maxRetries {
			// Wait before next attempt
			select {
			case <-ctx.Done():
				result.Success = false
				result.Error = ctx.Err()
				result.Attempts = attempt + 1
				return result
			case <-time.After(backoff):
				backoff = time.Duration(float64(backoff) * r.config.RetryBackoffMultiplier)
			}
		}
	}

	result.Success = false
	result.Error = fmt.Errorf("retry failed after %d attempts", maxRetries+1)
	result.Attempts = maxRetries + 1
	return result
}

// executeFallbackAction implements fallback strategy
func (r *ErrorRecoveryService) executeFallbackAction(ctx context.Context, action RecoveryAction) RecoveryResult {
	result := RecoveryResult{
		ActionID:  action.ID,
		Strategy:  action.Strategy,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Simulate fallback logic (in real implementation, this would use cached data)
	if r.simulateFallbackSuccess() {
		result.Success = true
		result.Metadata["fallback_source"] = "cache"
		result.RecoveryData = map[string]interface{}{
			"data_source": "cached",
			"timestamp":   time.Now().Add(-5 * time.Minute),
		}
	} else {
		result.Success = false
		result.Error = fmt.Errorf("fallback data not available")
	}

	return result
}

// executeCircuitBreakerAction implements circuit breaker strategy
func (r *ErrorRecoveryService) executeCircuitBreakerAction(ctx context.Context, action RecoveryAction) RecoveryResult {
	result := RecoveryResult{
		ActionID:  action.ID,
		Strategy:  action.Strategy,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Simulate circuit breaker logic
	threshold := r.config.CircuitBreakerThreshold
	if val, ok := action.Parameters["threshold"].(int); ok {
		threshold = val
	}

	if r.simulateCircuitBreakerSuccess(threshold) {
		result.Success = true
		result.Metadata["circuit_state"] = "closed"
	} else {
		result.Success = false
		result.Error = fmt.Errorf("circuit breaker open")
		result.Metadata["circuit_state"] = "open"
	}

	return result
}

// executeManualInterventionAction implements manual intervention strategy
func (r *ErrorRecoveryService) executeManualInterventionAction(ctx context.Context, action RecoveryAction) RecoveryResult {
	result := RecoveryResult{
		ActionID:  action.ID,
		Strategy:  action.Strategy,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Manual intervention always requires human action
	result.Success = false
	result.Error = fmt.Errorf("manual intervention required")
	result.Metadata["intervention_type"] = "human_action_required"
	result.Metadata["escalation_level"] = "high"

	return result
}

// executeGracefulDegradationAction implements graceful degradation strategy
func (r *ErrorRecoveryService) executeGracefulDegradationAction(ctx context.Context, action RecoveryAction) RecoveryResult {
	result := RecoveryResult{
		ActionID:  action.ID,
		Strategy:  action.Strategy,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Simulate graceful degradation
	if r.simulateGracefulDegradationSuccess() {
		result.Success = true
		result.Metadata["degradation_level"] = "reduced_functionality"
		result.RecoveryData = map[string]interface{}{
			"service_level":      "degraded",
			"features_available": []string{"basic", "essential"},
		}
	} else {
		result.Success = false
		result.Error = fmt.Errorf("graceful degradation failed")
	}

	return result
}

// executeRollbackAction implements rollback strategy
func (r *ErrorRecoveryService) executeRollbackAction(ctx context.Context, action RecoveryAction) RecoveryResult {
	result := RecoveryResult{
		ActionID:  action.ID,
		Strategy:  action.Strategy,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Simulate rollback logic
	if r.simulateRollbackSuccess() {
		result.Success = true
		result.Metadata["rollback_point"] = "previous_state"
		result.RecoveryData = map[string]interface{}{
			"state_restored":   true,
			"data_consistency": "verified",
		}
	} else {
		result.Success = false
		result.Error = fmt.Errorf("rollback failed")
	}

	return result
}

// executeCompensationAction implements compensation strategy
func (r *ErrorRecoveryService) executeCompensationAction(ctx context.Context, action RecoveryAction) RecoveryResult {
	result := RecoveryResult{
		ActionID:  action.ID,
		Strategy:  action.Strategy,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Simulate compensation logic
	if r.simulateCompensationSuccess() {
		result.Success = true
		result.Metadata["compensation_type"] = "data_reconciliation"
		result.RecoveryData = map[string]interface{}{
			"compensated_operations": 3,
			"data_integrity":         "maintained",
		}
	} else {
		result.Success = false
		result.Error = fmt.Errorf("compensation failed")
	}

	return result
}

// executeTimeoutAction implements timeout strategy
func (r *ErrorRecoveryService) executeTimeoutAction(ctx context.Context, action RecoveryAction) RecoveryResult {
	result := RecoveryResult{
		ActionID:  action.ID,
		Strategy:  action.Strategy,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Simulate timeout adjustment
	timeoutMultiplier := 2.0
	if val, ok := action.Parameters["timeout_multiplier"].(float64); ok {
		timeoutMultiplier = val
	}

	adjustedTimeout := time.Duration(float64(r.config.DefaultTimeout) * timeoutMultiplier)
	result.Metadata["original_timeout"] = r.config.DefaultTimeout
	result.Metadata["adjusted_timeout"] = adjustedTimeout
	result.Metadata["multiplier"] = timeoutMultiplier

	// Simulate success with adjusted timeout
	if r.simulateTimeoutAdjustmentSuccess() {
		result.Success = true
		result.RecoveryData = map[string]interface{}{
			"timeout_adjusted":    true,
			"operation_completed": true,
		}
	} else {
		result.Success = false
		result.Error = fmt.Errorf("operation still timed out with adjusted timeout")
	}

	return result
}

// executeResourceCleanupAction implements resource cleanup strategy
func (r *ErrorRecoveryService) executeResourceCleanupAction(ctx context.Context, action RecoveryAction) RecoveryResult {
	result := RecoveryResult{
		ActionID:  action.ID,
		Strategy:  action.Strategy,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Simulate resource cleanup
	if r.simulateResourceCleanupSuccess() {
		result.Success = true
		result.Metadata["resources_cleaned"] = 5
		result.Metadata["memory_freed"] = "50MB"
		result.RecoveryData = map[string]interface{}{
			"cleanup_complete": true,
			"leaks_prevented":  true,
		}
	} else {
		result.Success = false
		result.Error = fmt.Errorf("resource cleanup failed")
	}

	return result
}

// executeDataRecoveryAction implements data recovery strategy
func (r *ErrorRecoveryService) executeDataRecoveryAction(ctx context.Context, action RecoveryAction) RecoveryResult {
	result := RecoveryResult{
		ActionID:  action.ID,
		Strategy:  action.Strategy,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Simulate data recovery
	if r.simulateDataRecoverySuccess() {
		result.Success = true
		result.Metadata["data_recovered"] = true
		result.Metadata["validation_passed"] = true
		result.RecoveryData = map[string]interface{}{
			"recovery_method": "data_repair",
			"integrity_check": "passed",
		}
	} else {
		result.Success = false
		result.Error = fmt.Errorf("data recovery failed")
	}

	return result
}

// Helper methods for simulation (in real implementation, these would be actual operations)

func (r *ErrorRecoveryService) simulateRetrySuccess(attempt, maxRetries int) bool {
	// Simulate success on later attempts
	return attempt >= maxRetries/2
}

func (r *ErrorRecoveryService) simulateFallbackSuccess() bool {
	// Simulate 80% success rate for fallback
	return time.Now().UnixNano()%5 != 0
}

func (r *ErrorRecoveryService) simulateCircuitBreakerSuccess(threshold int) bool {
	// Simulate circuit breaker state
	return time.Now().UnixNano()%3 != 0
}

func (r *ErrorRecoveryService) simulateGracefulDegradationSuccess() bool {
	// Simulate 90% success rate for graceful degradation
	return time.Now().UnixNano()%10 != 0
}

func (r *ErrorRecoveryService) simulateRollbackSuccess() bool {
	// Simulate 85% success rate for rollback
	return time.Now().UnixNano()%7 != 0
}

func (r *ErrorRecoveryService) simulateCompensationSuccess() bool {
	// Simulate 75% success rate for compensation
	return time.Now().UnixNano()%4 != 0
}

func (r *ErrorRecoveryService) simulateTimeoutAdjustmentSuccess() bool {
	// Simulate 70% success rate for timeout adjustment
	return time.Now().UnixNano()%3 != 0
}

func (r *ErrorRecoveryService) simulateResourceCleanupSuccess() bool {
	// Simulate 95% success rate for resource cleanup
	return time.Now().UnixNano()%20 != 0
}

func (r *ErrorRecoveryService) simulateDataRecoverySuccess() bool {
	// Simulate 60% success rate for data recovery
	return time.Now().UnixNano()%5 != 0
}

// isSeverityGreaterOrEqual compares two severity levels
func (r *ErrorRecoveryService) isSeverityGreaterOrEqual(severity1, severity2 ErrorSeverity) bool {
	// Define severity levels in order from lowest to highest
	severityLevels := map[ErrorSeverity]int{
		SeverityInfo:     0,
		SeverityLow:      1,
		SeverityMedium:   2,
		SeverityHigh:     3,
		SeverityCritical: 4,
	}

	level1, exists1 := severityLevels[severity1]
	level2, exists2 := severityLevels[severity2]

	// If either severity is unknown, default to medium
	if !exists1 {
		level1 = 2
	}
	if !exists2 {
		level2 = 2
	}

	return level1 >= level2
}

// Helper methods

func (r *ErrorRecoveryService) calculateEstimatedTime(actions []RecoveryAction) time.Duration {
	var totalTime time.Duration
	for _, action := range actions {
		totalTime += action.Timeout
	}
	return totalTime
}

func (r *ErrorRecoveryService) calculateSuccessProbability(categorization *CategorizedError, actions []RecoveryAction) float64 {
	// Base probability based on category
	baseProbability := map[ErrorCategory]float64{
		CategoryNetwork:        0.8,
		CategoryDatabase:       0.7,
		CategoryValidation:     0.9,
		CategorySecurity:       0.3, // Manual intervention required
		CategoryPerformance:    0.6,
		CategoryAuthentication: 0.5,
		CategoryAuthorization:  0.4,
		CategoryUnknown:        0.5,
	}

	probability, exists := baseProbability[categorization.Category]
	if !exists {
		r.logger.Warn("Category not found in base probability map, using default",
			zap.String("category", string(categorization.Category)))
		probability = 0.5 // Default probability
	}

	// Adjust based on severity
	switch categorization.Severity {
	case SeverityLow:
		probability *= 1.1
	case SeverityMedium:
		probability *= 1.0
	case SeverityHigh:
		probability *= 0.8
	case SeverityCritical:
		probability *= 0.5
	}

	// Adjust based on number of actions
	if len(actions) > 1 {
		probability *= 0.9 // Multiple actions reduce probability
	}

	return probability
}

func (r *ErrorRecoveryService) isTerminalAction(strategy RecoveryStrategy) bool {
	terminalStrategies := []RecoveryStrategy{
		StrategyManualIntervention,
		StrategyGracefulDegradation,
		StrategyRollback,
		StrategyCompensation,
	}

	for _, s := range terminalStrategies {
		if s == strategy {
			return true
		}
	}
	return false
}

func (r *ErrorRecoveryService) isCriticalAction(strategy RecoveryStrategy) bool {
	criticalStrategies := []RecoveryStrategy{
		StrategyManualIntervention,
		StrategyResourceCleanup,
	}

	for _, s := range criticalStrategies {
		if s == strategy {
			return true
		}
	}
	return false
}

func (r *ErrorRecoveryService) requiresResourceCleanup(err error) bool {
	// Check if error indicates resource leaks
	resourceLeakKeywords := []string{"memory", "connection", "file", "socket", "goroutine"}
	errorMessage := err.Error()
	for _, keyword := range resourceLeakKeywords {
		if containsString(errorMessage, keyword) {
			return true
		}
	}
	return false
}

// convertPriorityToInt converts ErrorPriority to integer for sorting
func (r *ErrorRecoveryService) convertPriorityToInt(priority ErrorPriority) int {
	switch priority {
	case PriorityUrgent:
		return 1
	case PriorityHigh:
		return 2
	case PriorityMedium:
		return 3
	case PriorityLow:
		return 4
	case PriorityDeferred:
		return 5
	default:
		return 3 // Default to medium
	}
}

func (r *ErrorRecoveryService) updateStats(execution *RecoveryExecution) {
	r.stats.mu.Lock()
	defer r.stats.mu.Unlock()

	r.stats.TotalRecoveries++
	r.stats.TotalRecoveryTime += execution.Duration

	if execution.Status == StatusCompleted {
		r.stats.SuccessfulRecoveries++
	} else {
		r.stats.FailedRecoveries++
	}

	// Update average recovery time
	if r.stats.TotalRecoveries > 0 {
		r.stats.AverageRecoveryTime = r.stats.TotalRecoveryTime / time.Duration(r.stats.TotalRecoveries)
	}

	// Update strategy stats
	for _, result := range execution.Results {
		strategy := result.Strategy
		if r.stats.StrategyStats[strategy] == nil {
			r.stats.StrategyStats[strategy] = &StrategyStats{}
		}

		stats := r.stats.StrategyStats[strategy]
		stats.UsageCount++
		stats.TotalDuration += result.Duration

		if result.Success {
			stats.SuccessCount++
		} else {
			stats.FailureCount++
		}

		if stats.UsageCount > 0 {
			stats.AverageDuration = stats.TotalDuration / time.Duration(stats.UsageCount)
			stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.UsageCount)
		}
	}

	r.stats.LastUpdated = time.Now()
}

func (r *ErrorRecoveryService) addToHistory(execution *RecoveryExecution) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.recoveryHistory = append(r.recoveryHistory, execution)

	// Maintain history size limit
	if len(r.recoveryHistory) > r.config.RecoveryHistorySize {
		r.recoveryHistory = r.recoveryHistory[1:]
	}
}

// Utility functions

func generateRecoveryPlanID() string {
	return fmt.Sprintf("rp_%d", time.Now().UnixNano())
}

func generateActionID() string {
	return fmt.Sprintf("ra_%d", time.Now().UnixNano())
}
