package industry_codes

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// RetryMechanism provides intelligent retry capabilities with exponential backoff
type RetryMechanism struct {
	config          *RetryConfig
	logger          *zap.Logger
	stats           *RetryStats
	circuitBreakers map[string]*CircuitBreaker
}

// RetryConfig defines configuration for retry behavior
type RetryConfig struct {
	MaxAttempts             int           `json:"max_attempts"`
	BaseDelay               time.Duration `json:"base_delay"`
	MaxDelay                time.Duration `json:"max_delay"`
	BackoffMultiplier       float64       `json:"backoff_multiplier"`
	JitterFactor            float64       `json:"jitter_factor"`
	RetryableErrors         []string      `json:"retryable_errors"`
	NonRetryableErrors      []string      `json:"non_retryable_errors"`
	TimeoutPerAttempt       time.Duration `json:"timeout_per_attempt"`
	CircuitBreakerEnabled   bool          `json:"circuit_breaker_enabled"`
	CircuitBreakerThreshold int           `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration `json:"circuit_breaker_timeout"`
}

// RetryStats tracks retry performance and statistics
type RetryStats struct {
	TotalAttempts       int64         `json:"total_attempts"`
	SuccessfulRetries   int64         `json:"successful_retries"`
	FailedRetries       int64         `json:"failed_retries"`
	TotalRetryTime      time.Duration `json:"total_retry_time"`
	AverageRetryTime    time.Duration `json:"average_retry_time"`
	LastFailureTime     time.Time     `json:"last_failure_time"`
	ConsecutiveFailures int64         `json:"consecutive_failures"`
	CircuitBreakerTrips int64         `json:"circuit_breaker_trips"`
}

// RetryResult represents the result of a retry operation
type RetryResult struct {
	Success           bool                   `json:"success"`
	Data              interface{}            `json:"data"`
	Attempts          int                    `json:"attempts"`
	TotalTime         time.Duration          `json:"total_time"`
	LastError         error                  `json:"last_error"`
	RetryDelays       []time.Duration        `json:"retry_delays"`
	CircuitBreakerHit bool                   `json:"circuit_breaker_hit"`
	BackoffStrategy   string                 `json:"backoff_strategy"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// CircuitBreaker tracks circuit breaker state
type CircuitBreaker struct {
	state           CircuitBreakerState
	failureCount    int64
	lastFailureTime time.Time
	threshold       int
	timeout         time.Duration
	logger          *zap.Logger
}

// CircuitBreakerState represents the current state of the circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerClosed   CircuitBreakerState = "closed"
	CircuitBreakerOpen     CircuitBreakerState = "open"
	CircuitBreakerHalfOpen CircuitBreakerState = "half_open"
)

// NewRetryMechanism creates a new retry mechanism with default configuration
func NewRetryMechanism(logger *zap.Logger, config *RetryConfig) *RetryMechanism {
	if config == nil {
		config = &RetryConfig{
			MaxAttempts:             3,
			BaseDelay:               100 * time.Millisecond,
			MaxDelay:                30 * time.Second,
			BackoffMultiplier:       2.0,
			JitterFactor:            0.1,
			TimeoutPerAttempt:       5 * time.Second,
			CircuitBreakerEnabled:   true,
			CircuitBreakerThreshold: 5,
			CircuitBreakerTimeout:   60 * time.Second,
			RetryableErrors: []string{
				"timeout",
				"connection refused",
				"network error",
				"temporary failure",
				"rate limit exceeded",
				"service unavailable",
			},
			NonRetryableErrors: []string{
				"invalid input",
				"authentication failed",
				"authorization denied",
				"not found",
				"bad request",
			},
		}
	}

	return &RetryMechanism{
		config:          config,
		logger:          logger,
		stats:           &RetryStats{},
		circuitBreakers: make(map[string]*CircuitBreaker),
	}
}

// ExecuteWithRetry executes an operation with retry logic and exponential backoff
func (rm *RetryMechanism) ExecuteWithRetry(ctx context.Context, operation func() (interface{}, error), operationName string) *RetryResult {
	startTime := time.Now()
	result := &RetryResult{
		RetryDelays: []time.Duration{},
		Metadata:    make(map[string]interface{}),
	}

	// Check circuit breaker first
	if rm.config.CircuitBreakerEnabled {
		circuitBreaker := rm.getCircuitBreaker(operationName)
		if !circuitBreaker.CanExecute() {
			result.CircuitBreakerHit = true
			result.LastError = fmt.Errorf("circuit breaker is open for operation: %s", operationName)
			rm.logger.Warn("circuit breaker blocked execution",
				zap.String("operation", operationName),
				zap.String("state", string(circuitBreaker.state)))
			return result
		}
	}

	// Execute operation with retries
	for attempt := 1; attempt <= rm.config.MaxAttempts; attempt++ {
		rm.stats.TotalAttempts++

		// Create timeout context for this attempt
		attemptCtx, cancel := context.WithTimeout(ctx, rm.config.TimeoutPerAttempt)

		// Execute operation
		data, err := operation()

		// Check if operation succeeded
		if err == nil {
			result.Success = true
			result.Data = data
			result.Attempts = attempt
			result.TotalTime = time.Since(startTime)

			// Update stats
			rm.stats.SuccessfulRetries++
			rm.updateAverageRetryTime(result.TotalTime)

			// Reset circuit breaker on success
			if rm.config.CircuitBreakerEnabled {
				circuitBreaker := rm.getCircuitBreaker(operationName)
				circuitBreaker.OnSuccess()
			}

			rm.logger.Info("operation succeeded",
				zap.String("operation", operationName),
				zap.Int("attempts", attempt),
				zap.Duration("total_time", result.TotalTime))

			cancel()
			return result
		}

		// Check if error is retryable
		if !rm.isRetryableError(err) {
			result.LastError = err
			result.Attempts = attempt
			result.TotalTime = time.Since(startTime)

			rm.logger.Warn("non-retryable error encountered",
				zap.String("operation", operationName),
				zap.Error(err),
				zap.Int("attempt", attempt))

			cancel()
			return result
		}

		// Update circuit breaker on failure
		if rm.config.CircuitBreakerEnabled {
			circuitBreaker := rm.getCircuitBreaker(operationName)
			circuitBreaker.OnFailure()
		}

		// Log retry attempt
		rm.logger.Info("retry attempt",
			zap.String("operation", operationName),
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", rm.config.MaxAttempts),
			zap.Error(err))

		// If this is not the last attempt, calculate delay and wait
		if attempt < rm.config.MaxAttempts {
			delay := rm.calculateDelay(attempt)
			result.RetryDelays = append(result.RetryDelays, delay)

			rm.logger.Info("waiting before retry",
				zap.String("operation", operationName),
				zap.Duration("delay", delay))

			// Wait before next attempt
			select {
			case <-attemptCtx.Done():
				cancel()
				result.LastError = fmt.Errorf("operation timed out after %d attempts", attempt)
				result.Attempts = attempt
				result.TotalTime = time.Since(startTime)
				return result
			case <-time.After(delay):
				// Continue to next attempt
			}
		}

		cancel()
	}

	// All attempts failed
	result.LastError = fmt.Errorf("operation failed after %d attempts", rm.config.MaxAttempts)
	result.Attempts = rm.config.MaxAttempts
	result.TotalTime = time.Since(startTime)

	// Update stats
	rm.stats.FailedRetries++
	rm.stats.LastFailureTime = time.Now()
	rm.stats.ConsecutiveFailures++

	rm.logger.Error("operation failed after all retry attempts",
		zap.String("operation", operationName),
		zap.Int("attempts", rm.config.MaxAttempts),
		zap.Duration("total_time", result.TotalTime),
		zap.Error(result.LastError))

	return result
}

// calculateDelay calculates the delay for a specific attempt using exponential backoff with jitter
func (rm *RetryMechanism) calculateDelay(attempt int) time.Duration {
	// Calculate exponential backoff
	delay := float64(rm.config.BaseDelay) * math.Pow(rm.config.BackoffMultiplier, float64(attempt-1))

	// Apply maximum delay cap
	if delay > float64(rm.config.MaxDelay) {
		delay = float64(rm.config.MaxDelay)
	}

	// Add jitter to prevent thundering herd
	if rm.config.JitterFactor > 0 {
		jitter := delay * rm.config.JitterFactor * (rand.Float64()*2 - 1)
		delay += jitter

		// Ensure delay doesn't go below base delay
		if delay < float64(rm.config.BaseDelay) {
			delay = float64(rm.config.BaseDelay)
		}
	}

	return time.Duration(delay)
}

// isRetryableError determines if an error should trigger a retry
func (rm *RetryMechanism) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()

	// Check non-retryable errors first
	for _, nonRetryable := range rm.config.NonRetryableErrors {
		if containsString(errorStr, nonRetryable) {
			return false
		}
	}

	// Check retryable errors
	for _, retryable := range rm.config.RetryableErrors {
		if containsString(errorStr, retryable) {
			return true
		}
	}

	// Default to retryable for unknown errors
	return true
}

// containsString checks if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(len(s) == len(substr) && s == substr ||
			len(s) > len(substr) && (containsStringHelper(s, substr)))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// getCircuitBreaker returns or creates a circuit breaker for the operation
func (rm *RetryMechanism) getCircuitBreaker(operationName string) *CircuitBreaker {
	if cb, exists := rm.circuitBreakers[operationName]; exists {
		return cb
	}

	cb := &CircuitBreaker{
		state:           CircuitBreakerClosed,
		failureCount:    0,
		lastFailureTime: time.Time{},
		threshold:       rm.config.CircuitBreakerThreshold,
		timeout:         rm.config.CircuitBreakerTimeout,
		logger:          rm.logger,
	}

	rm.circuitBreakers[operationName] = cb
	return cb
}

// updateAverageRetryTime updates the average retry time statistic
func (rm *RetryMechanism) updateAverageRetryTime(retryTime time.Duration) {
	rm.stats.TotalRetryTime += retryTime
	if rm.stats.SuccessfulRetries > 0 {
		rm.stats.AverageRetryTime = rm.stats.TotalRetryTime / time.Duration(rm.stats.SuccessfulRetries)
	}
}

// GetStats returns the current retry statistics
func (rm *RetryMechanism) GetStats() *RetryStats {
	return rm.stats
}

// ResetStats resets all retry statistics
func (rm *RetryMechanism) ResetStats() {
	rm.stats = &RetryStats{}
}

// Circuit Breaker Methods

// CanExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() bool {
	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		// Check if timeout has passed to transition to half-open
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = CircuitBreakerHalfOpen
			cb.logger.Info("circuit breaker transitioning to half-open")
			return true
		}
		return false
	case CircuitBreakerHalfOpen:
		return true
	default:
		return true
	}
}

// OnSuccess handles successful execution
func (cb *CircuitBreaker) OnSuccess() {
	switch cb.state {
	case CircuitBreakerHalfOpen:
		// Reset to closed state
		cb.state = CircuitBreakerClosed
		cb.failureCount = 0
		cb.logger.Info("circuit breaker reset to closed state")
	case CircuitBreakerClosed:
		// Reset failure count on success
		cb.failureCount = 0
	}
}

// OnFailure handles failed execution
func (cb *CircuitBreaker) OnFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case CircuitBreakerClosed:
		if cb.failureCount >= int64(cb.threshold) {
			cb.state = CircuitBreakerOpen
			cb.logger.Warn("circuit breaker opened",
				zap.Int64("failure_count", cb.failureCount),
				zap.Int("threshold", cb.threshold))
		}
	case CircuitBreakerHalfOpen:
		// Immediately open on failure in half-open state
		cb.state = CircuitBreakerOpen
		cb.logger.Warn("circuit breaker reopened after failure in half-open state")
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	return cb.state
}

// GetFailureCount returns the current failure count
func (cb *CircuitBreaker) GetFailureCount() int64 {
	return cb.failureCount
}

// RetryableOperation defines an interface for operations that can be retried
type RetryableOperation interface {
	Execute(ctx context.Context) (interface{}, error)
	GetName() string
}

// RetryableOperationFunc is a function that implements RetryableOperation
type RetryableOperationFunc func(ctx context.Context) (interface{}, error)

func (f RetryableOperationFunc) Execute(ctx context.Context) (interface{}, error) {
	return f(ctx)
}

func (f RetryableOperationFunc) GetName() string {
	return "anonymous_operation"
}

// ExecuteRetryableOperation executes a retryable operation with retry logic
func (rm *RetryMechanism) ExecuteRetryableOperation(ctx context.Context, operation RetryableOperation) *RetryResult {
	return rm.ExecuteWithRetry(ctx, func() (interface{}, error) {
		return operation.Execute(ctx)
	}, operation.GetName())
}
