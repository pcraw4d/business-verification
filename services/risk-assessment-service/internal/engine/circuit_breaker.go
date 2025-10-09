package engine

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	RecoveryTimeout  time.Duration `json:"recovery_timeout"`
	HalfOpenMaxCalls int           `json:"half_open_max_calls"`
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config        CircuitBreakerConfig
	state         CircuitBreakerState
	failures      int
	lastFailTime  time.Time
	halfOpenCalls int
	mu            sync.RWMutex
	logger        *zap.Logger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig, logger *zap.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
		logger: logger,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if circuit breaker should allow the call
	if !cb.canExecute() {
		return nil, fmt.Errorf("circuit breaker is %s", cb.state.String())
	}

	// Execute the function
	result, err := fn()

	// Record the result
	cb.recordResult(err)

	return result, err
}

// canExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if recovery timeout has passed
		if time.Since(cb.lastFailTime) >= cb.config.RecoveryTimeout {
			cb.state = StateHalfOpen
			cb.halfOpenCalls = 0
			cb.logger.Info("Circuit breaker transitioning to HALF_OPEN")
			return true
		}
		return false
	case StateHalfOpen:
		// Allow limited calls in half-open state
		if cb.halfOpenCalls < cb.config.HalfOpenMaxCalls {
			cb.halfOpenCalls++
			return true
		}
		return false
	default:
		return false
	}
}

// recordResult records the result of an execution
func (cb *CircuitBreaker) recordResult(err error) {
	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()

		// Check if we should open the circuit
		if cb.failures >= cb.config.FailureThreshold {
			if cb.state != StateOpen {
				cb.state = StateOpen
				cb.logger.Warn("Circuit breaker opened due to failures",
					zap.Int("failures", cb.failures),
					zap.Int("threshold", cb.config.FailureThreshold))
			}
		}
	} else {
		// Success - reset failures and close circuit if needed
		if cb.state == StateHalfOpen {
			cb.state = StateClosed
			cb.failures = 0
			cb.halfOpenCalls = 0
			cb.logger.Info("Circuit breaker closed after successful call")
		} else if cb.state == StateClosed {
			// Reset failure count on success
			cb.failures = 0
		}
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitBreakerStats{
		State:         cb.state,
		Failures:      cb.failures,
		LastFailTime:  cb.lastFailTime,
		HalfOpenCalls: cb.halfOpenCalls,
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.halfOpenCalls = 0
	cb.lastFailTime = time.Time{}

	cb.logger.Info("Circuit breaker reset to CLOSED state")
}

// CircuitBreakerStats holds circuit breaker statistics
type CircuitBreakerStats struct {
	State         CircuitBreakerState `json:"state"`
	Failures      int                 `json:"failures"`
	LastFailTime  time.Time           `json:"last_fail_time"`
	HalfOpenCalls int                 `json:"half_open_calls"`
}
