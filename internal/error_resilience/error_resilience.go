package error_resilience

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ErrorResilienceManager manages error resilience and graceful degradation
type ErrorResilienceManager struct {
	circuitBreakers     map[string]*CircuitBreaker
	retryPolicies       map[string]*RetryPolicy
	fallbackStrategies  map[string]*FallbackStrategy
	degradationPolicies map[string]*DegradationPolicy
	mu                  sync.RWMutex
	logger              *observability.Logger
	metrics             *ResilienceMetrics
}

// CircuitBreaker represents a circuit breaker for external dependencies
type CircuitBreaker struct {
	Name             string
	FailureThreshold int64
	SuccessThreshold int64
	Timeout          time.Duration
	FailureCount     int64
	SuccessCount     int64
	LastFailureTime  time.Time
	State            CircuitBreakerState
	mu               sync.RWMutex
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerClosed   CircuitBreakerState = "closed"
	CircuitBreakerOpen     CircuitBreakerState = "open"
	CircuitBreakerHalfOpen CircuitBreakerState = "half_open"
)

// RetryPolicy represents a retry policy with exponential backoff
type RetryPolicy struct {
	Name            string
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []string
}

// FallbackStrategy represents a fallback strategy for failed modules
type FallbackStrategy struct {
	Name              string
	Enabled           bool
	Strategy          string
	FallbackData      map[string]interface{}
	AlternativeModule string
	DegradationLevel  DegradationLevel
}

// DegradationPolicy represents a degradation policy for graceful degradation
type DegradationPolicy struct {
	Name                   string
	Enabled                bool
	DegradationLevels      []DegradationLevel
	PartialResultThreshold float64
	MinimalResultThreshold float64
}

// DegradationLevel represents a level of service degradation
type DegradationLevel string

const (
	DegradationLevelNone     DegradationLevel = "none"
	DegradationLevelPartial  DegradationLevel = "partial"
	DegradationLevelMinimal  DegradationLevel = "minimal"
	DegradationLevelFallback DegradationLevel = "fallback"
)

// ResilienceMetrics represents metrics for error resilience
type ResilienceMetrics struct {
	CircuitBreakerTrips  int64
	RetryAttempts        int64
	FallbackExecutions   int64
	DegradationEvents    int64
	SuccessfulRecoveries int64
	FailedRecoveries     int64
	mu                   sync.RWMutex
}

// ModuleResult represents a result from a module with degradation information
type ModuleResult struct {
	ModuleName       string
	Success          bool
	Data             interface{}
	Error            error
	DegradationLevel DegradationLevel
	Confidence       float64
	ProcessingTime   time.Duration
	FallbackUsed     bool
	RetryAttempts    int
}

// NewErrorResilienceManager creates a new error resilience manager
func NewErrorResilienceManager(logger *observability.Logger) *ErrorResilienceManager {
	return &ErrorResilienceManager{
		circuitBreakers:     make(map[string]*CircuitBreaker),
		retryPolicies:       make(map[string]*RetryPolicy),
		fallbackStrategies:  make(map[string]*FallbackStrategy),
		degradationPolicies: make(map[string]*DegradationPolicy),
		logger:              logger,
		metrics:             &ResilienceMetrics{},
	}
}

// RegisterCircuitBreaker registers a circuit breaker for an external dependency
func (erm *ErrorResilienceManager) RegisterCircuitBreaker(
	name string,
	failureThreshold int64,
	successThreshold int64,
	timeout time.Duration,
) {
	erm.mu.Lock()
	defer erm.mu.Unlock()

	erm.circuitBreakers[name] = &CircuitBreaker{
		Name:             name,
		FailureThreshold: failureThreshold,
		SuccessThreshold: successThreshold,
		Timeout:          timeout,
		State:            CircuitBreakerClosed,
	}

	erm.logger.Info("Circuit breaker registered", map[string]interface{}{
		"name":              name,
		"failure_threshold": failureThreshold,
		"success_threshold": successThreshold,
		"timeout":           timeout,
	})
}

// RegisterRetryPolicy registers a retry policy for a module
func (erm *ErrorResilienceManager) RegisterRetryPolicy(
	name string,
	maxAttempts int,
	initialDelay time.Duration,
	maxDelay time.Duration,
	backoffFactor float64,
	retryableErrors []string,
) {
	erm.mu.Lock()
	defer erm.mu.Unlock()

	erm.retryPolicies[name] = &RetryPolicy{
		Name:            name,
		MaxAttempts:     maxAttempts,
		InitialDelay:    initialDelay,
		MaxDelay:        maxDelay,
		BackoffFactor:   backoffFactor,
		RetryableErrors: retryableErrors,
	}

	erm.logger.Info("Retry policy registered", map[string]interface{}{
		"name":           name,
		"max_attempts":   maxAttempts,
		"initial_delay":  initialDelay,
		"max_delay":      maxDelay,
		"backoff_factor": backoffFactor,
	})
}

// RegisterFallbackStrategy registers a fallback strategy for a module
func (erm *ErrorResilienceManager) RegisterFallbackStrategy(
	name string,
	enabled bool,
	strategy string,
	fallbackData map[string]interface{},
	alternativeModule string,
	degradationLevel DegradationLevel,
) {
	erm.mu.Lock()
	defer erm.mu.Unlock()

	erm.fallbackStrategies[name] = &FallbackStrategy{
		Name:              name,
		Enabled:           enabled,
		Strategy:          strategy,
		FallbackData:      fallbackData,
		AlternativeModule: alternativeModule,
		DegradationLevel:  degradationLevel,
	}

	erm.logger.Info("Fallback strategy registered", map[string]interface{}{
		"name":               name,
		"enabled":            enabled,
		"strategy":           strategy,
		"alternative_module": alternativeModule,
		"degradation_level":  degradationLevel,
	})
}

// RegisterDegradationPolicy registers a degradation policy for a module
func (erm *ErrorResilienceManager) RegisterDegradationPolicy(
	name string,
	enabled bool,
	degradationLevels []DegradationLevel,
	partialResultThreshold float64,
	minimalResultThreshold float64,
) {
	erm.mu.Lock()
	defer erm.mu.Unlock()

	erm.degradationPolicies[name] = &DegradationPolicy{
		Name:                   name,
		Enabled:                enabled,
		DegradationLevels:      degradationLevels,
		PartialResultThreshold: partialResultThreshold,
		MinimalResultThreshold: minimalResultThreshold,
	}

	erm.logger.Info("Degradation policy registered", map[string]interface{}{
		"name":              name,
		"enabled":           enabled,
		"partial_threshold": partialResultThreshold,
		"minimal_threshold": minimalResultThreshold,
	})
}

// ExecuteWithResilience executes a module with full error resilience
func (erm *ErrorResilienceManager) ExecuteWithResilience(
	ctx context.Context,
	moduleName string,
	operation func() (interface{}, error),
) *ModuleResult {
	start := time.Now()

	// Check circuit breaker first
	if !erm.checkCircuitBreaker(moduleName) {
		return erm.createDegradedResult(moduleName, nil, fmt.Errorf("circuit breaker open"), DegradationLevelFallback, 0.0)
	}

	// Execute with retry policy
	result, err, retryAttempts := erm.executeWithRetry(ctx, moduleName, operation)

	// Handle success
	if err == nil {
		erm.recordSuccess(moduleName)
		return &ModuleResult{
			ModuleName:       moduleName,
			Success:          true,
			Data:             result,
			Error:            nil,
			DegradationLevel: DegradationLevelNone,
			Confidence:       1.0,
			ProcessingTime:   time.Since(start),
			FallbackUsed:     false,
			RetryAttempts:    retryAttempts,
		}
	}

	// Handle failure with fallback
	erm.recordFailure(moduleName)
	return erm.handleFailure(moduleName, result, err, time.Since(start), retryAttempts)
}

// checkCircuitBreaker checks if the circuit breaker allows the operation
func (erm *ErrorResilienceManager) checkCircuitBreaker(moduleName string) bool {
	erm.mu.RLock()
	cb, exists := erm.circuitBreakers[moduleName]
	erm.mu.RUnlock()

	if !exists {
		return true // No circuit breaker configured
	}

	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.State {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		if time.Since(cb.LastFailureTime) > cb.Timeout {
			cb.mu.Lock()
			cb.State = CircuitBreakerHalfOpen
			cb.mu.Unlock()
			return true
		}
		return false
	case CircuitBreakerHalfOpen:
		return true
	default:
		return true
	}
}

// executeWithRetry executes the operation with retry policy
func (erm *ErrorResilienceManager) executeWithRetry(
	ctx context.Context,
	moduleName string,
	operation func() (interface{}, error),
) (interface{}, error, int) {
	erm.mu.RLock()
	policy, exists := erm.retryPolicies[moduleName]
	erm.mu.RUnlock()

	if !exists {
		// No retry policy, execute once
		result, err := operation()
		return result, err, 0
	}

	var lastErr error
	delay := policy.InitialDelay

	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		erm.metrics.mu.Lock()
		erm.metrics.RetryAttempts++
		erm.metrics.mu.Unlock()

		result, err := operation()
		if err == nil {
			return result, nil, attempt + 1
		}

		lastErr = err

		// Check if error is retryable
		if !erm.isRetryableError(err, policy.RetryableErrors) {
			break
		}

		// Don't sleep on the last attempt
		if attempt < policy.MaxAttempts-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err(), attempt + 1
			case <-time.After(delay):
				// Calculate next delay with exponential backoff
				delay = time.Duration(float64(delay) * policy.BackoffFactor)
				if delay > policy.MaxDelay {
					delay = policy.MaxDelay
				}
			}
		}
	}

	return nil, lastErr, policy.MaxAttempts
}

// isRetryableError checks if an error is retryable
func (erm *ErrorResilienceManager) isRetryableError(err error, retryableErrors []string) bool {
	if len(retryableErrors) == 0 {
		// Default retryable errors
		retryableErrors = []string{"timeout", "connection", "temporary", "rate_limit"}
	}

	errStr := err.Error()
	for _, retryableError := range retryableErrors {
		if contains(errStr, retryableError) {
			return true
		}
	}
	return false
}

// handleFailure handles module failure with fallback and degradation
func (erm *ErrorResilienceManager) handleFailure(
	moduleName string,
	result interface{},
	err error,
	processingTime time.Duration,
	retryAttempts int,
) *ModuleResult {
	// Try fallback strategy
	fallbackResult, fallbackErr := erm.executeFallback(moduleName, result, err)
	if fallbackErr == nil {
		erm.metrics.mu.Lock()
		erm.metrics.FallbackExecutions++
		erm.metrics.mu.Unlock()

		return &ModuleResult{
			ModuleName:       moduleName,
			Success:          true,
			Data:             fallbackResult,
			Error:            nil,
			DegradationLevel: erm.getFallbackDegradationLevel(moduleName),
			Confidence:       0.7, // Lower confidence for fallback
			ProcessingTime:   processingTime,
			FallbackUsed:     true,
			RetryAttempts:    retryAttempts,
		}
	}

	// Try graceful degradation
	degradedResult, degradationLevel, confidence := erm.executeGracefulDegradation(moduleName, result, err)
	if degradedResult != nil {
		erm.metrics.mu.Lock()
		erm.metrics.DegradationEvents++
		erm.metrics.mu.Unlock()

		return &ModuleResult{
			ModuleName:       moduleName,
			Success:          true,
			Data:             degradedResult,
			Error:            nil,
			DegradationLevel: degradationLevel,
			Confidence:       confidence,
			ProcessingTime:   processingTime,
			FallbackUsed:     false,
			RetryAttempts:    retryAttempts,
		}
	}

	// Complete failure
	erm.metrics.mu.Lock()
	erm.metrics.FailedRecoveries++
	erm.metrics.mu.Unlock()

	return &ModuleResult{
		ModuleName:       moduleName,
		Success:          false,
		Data:             nil,
		Error:            err,
		DegradationLevel: DegradationLevelFallback,
		Confidence:       0.0,
		ProcessingTime:   processingTime,
		FallbackUsed:     false,
		RetryAttempts:    retryAttempts,
	}
}

// executeFallback executes the fallback strategy for a module
func (erm *ErrorResilienceManager) executeFallback(
	moduleName string,
	originalResult interface{},
	originalError error,
) (interface{}, error) {
	erm.mu.RLock()
	strategy, exists := erm.fallbackStrategies[moduleName]
	erm.mu.RUnlock()

	if !exists || !strategy.Enabled {
		return nil, fmt.Errorf("no fallback strategy available")
	}

	switch strategy.Strategy {
	case "static_data":
		return strategy.FallbackData, nil
	case "cached_data":
		return erm.getCachedData(moduleName)
	case "alternative_module":
		return erm.callAlternativeModule(strategy.AlternativeModule, originalResult)
	case "degraded_response":
		return erm.generateDegradedResponse(moduleName, originalResult, originalError)
	default:
		return nil, fmt.Errorf("unknown fallback strategy: %s", strategy.Strategy)
	}
}

// executeGracefulDegradation executes graceful degradation for a module
func (erm *ErrorResilienceManager) executeGracefulDegradation(
	moduleName string,
	originalResult interface{},
	originalError error,
) (interface{}, DegradationLevel, float64) {
	erm.mu.RLock()
	policy, exists := erm.degradationPolicies[moduleName]
	erm.mu.RUnlock()

	if !exists || !policy.Enabled {
		return nil, DegradationLevelNone, 0.0
	}

	// Try partial degradation
	if erm.canProvidePartialResult(originalResult, policy.PartialResultThreshold) {
		partialResult := erm.generatePartialResult(moduleName, originalResult)
		return partialResult, DegradationLevelPartial, policy.PartialResultThreshold
	}

	// Try minimal degradation
	if erm.canProvideMinimalResult(originalResult, policy.MinimalResultThreshold) {
		minimalResult := erm.generateMinimalResult(moduleName, originalResult)
		return minimalResult, DegradationLevelMinimal, policy.MinimalResultThreshold
	}

	return nil, DegradationLevelNone, 0.0
}

// Helper methods for fallback and degradation
func (erm *ErrorResilienceManager) getCachedData(moduleName string) (interface{}, error) {
	// TODO: Implement actual cache retrieval
	return map[string]interface{}{
		"module": moduleName,
		"source": "cache",
		"data":   "cached_response",
	}, nil
}

func (erm *ErrorResilienceManager) callAlternativeModule(moduleName string, originalResult interface{}) (interface{}, error) {
	// TODO: Implement actual alternative module call
	return map[string]interface{}{
		"module": moduleName,
		"source": "alternative_module",
		"data":   originalResult,
	}, nil
}

func (erm *ErrorResilienceManager) generateDegradedResponse(moduleName string, originalResult interface{}, originalError error) (interface{}, error) {
	return map[string]interface{}{
		"module":      moduleName,
		"source":      "degraded_response",
		"status":      "degraded",
		"data":        originalResult,
		"error":       originalError.Error(),
		"degraded_at": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (erm *ErrorResilienceManager) canProvidePartialResult(result interface{}, threshold float64) bool {
	// TODO: Implement logic to determine if partial result is possible
	return threshold > 0.5
}

func (erm *ErrorResilienceManager) canProvideMinimalResult(result interface{}, threshold float64) bool {
	// TODO: Implement logic to determine if minimal result is possible
	return threshold > 0.1
}

func (erm *ErrorResilienceManager) generatePartialResult(moduleName string, originalResult interface{}) interface{} {
	// TODO: Implement partial result generation
	return map[string]interface{}{
		"module": moduleName,
		"source": "partial_result",
		"data":   originalResult,
		"status": "partial",
	}
}

func (erm *ErrorResilienceManager) generateMinimalResult(moduleName string, originalResult interface{}) interface{} {
	// TODO: Implement minimal result generation
	return map[string]interface{}{
		"module": moduleName,
		"source": "minimal_result",
		"data":   originalResult,
		"status": "minimal",
	}
}

func (erm *ErrorResilienceManager) getFallbackDegradationLevel(moduleName string) DegradationLevel {
	erm.mu.RLock()
	strategy, exists := erm.fallbackStrategies[moduleName]
	erm.mu.RUnlock()

	if exists {
		return strategy.DegradationLevel
	}
	return DegradationLevelFallback
}

// recordSuccess records a successful operation
func (erm *ErrorResilienceManager) recordSuccess(moduleName string) {
	erm.mu.RLock()
	cb, exists := erm.circuitBreakers[moduleName]
	erm.mu.RUnlock()

	if !exists {
		return
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.SuccessCount++
	if cb.State == CircuitBreakerHalfOpen && cb.SuccessCount >= cb.SuccessThreshold {
		cb.State = CircuitBreakerClosed
		cb.FailureCount = 0
		cb.SuccessCount = 0

		erm.metrics.mu.Lock()
		erm.metrics.SuccessfulRecoveries++
		erm.metrics.mu.Unlock()

		erm.logger.Info("Circuit breaker closed", map[string]interface{}{
			"name": moduleName,
		})
	}
}

// recordFailure records a failed operation
func (erm *ErrorResilienceManager) recordFailure(moduleName string) {
	erm.mu.RLock()
	cb, exists := erm.circuitBreakers[moduleName]
	erm.mu.RUnlock()

	if !exists {
		return
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.FailureCount++
	cb.LastFailureTime = time.Now()

	if cb.State == CircuitBreakerClosed && cb.FailureCount >= cb.FailureThreshold {
		cb.State = CircuitBreakerOpen
		cb.FailureCount = 0

		erm.metrics.mu.Lock()
		erm.metrics.CircuitBreakerTrips++
		erm.metrics.mu.Unlock()

		erm.logger.Warn("Circuit breaker opened", map[string]interface{}{
			"name": moduleName,
		})
	} else if cb.State == CircuitBreakerHalfOpen {
		cb.State = CircuitBreakerOpen
		cb.FailureCount = 0

		erm.metrics.mu.Lock()
		erm.metrics.CircuitBreakerTrips++
		erm.metrics.mu.Unlock()

		erm.logger.Warn("Circuit breaker reopened", map[string]interface{}{
			"name": moduleName,
		})
	}
}

// createDegradedResult creates a degraded result
func (erm *ErrorResilienceManager) createDegradedResult(
	moduleName string,
	data interface{},
	err error,
	degradationLevel DegradationLevel,
	confidence float64,
) *ModuleResult {
	return &ModuleResult{
		ModuleName:       moduleName,
		Success:          err == nil,
		Data:             data,
		Error:            err,
		DegradationLevel: degradationLevel,
		Confidence:       confidence,
		ProcessingTime:   0,
		FallbackUsed:     true,
		RetryAttempts:    0,
	}
}

// GetMetrics returns the current resilience metrics
func (erm *ErrorResilienceManager) GetMetrics() map[string]interface{} {
	erm.metrics.mu.RLock()
	defer erm.metrics.mu.RUnlock()

	return map[string]interface{}{
		"circuit_breaker_trips": erm.metrics.CircuitBreakerTrips,
		"retry_attempts":        erm.metrics.RetryAttempts,
		"fallback_executions":   erm.metrics.FallbackExecutions,
		"degradation_events":    erm.metrics.DegradationEvents,
		"successful_recoveries": erm.metrics.SuccessfulRecoveries,
		"failed_recoveries":     erm.metrics.FailedRecoveries,
	}
}

// GetCircuitBreakerState returns the state of a circuit breaker
func (erm *ErrorResilienceManager) GetCircuitBreakerState(moduleName string) map[string]interface{} {
	erm.mu.RLock()
	cb, exists := erm.circuitBreakers[moduleName]
	erm.mu.RUnlock()

	if !exists {
		return nil
	}

	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"name":          cb.Name,
		"state":         cb.State,
		"failure_count": cb.FailureCount,
		"success_count": cb.SuccessCount,
		"last_failure":  cb.LastFailureTime,
	}
}

// ResetCircuitBreaker resets a circuit breaker
func (erm *ErrorResilienceManager) ResetCircuitBreaker(moduleName string) error {
	erm.mu.RLock()
	cb, exists := erm.circuitBreakers[moduleName]
	erm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("circuit breaker not found: %s", moduleName)
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.State = CircuitBreakerClosed
	cb.FailureCount = 0
	cb.SuccessCount = 0

	erm.logger.Info("Circuit breaker reset", map[string]interface{}{
		"name": moduleName,
	})
	return nil
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
