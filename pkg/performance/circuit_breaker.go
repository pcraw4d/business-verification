package performance

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitStateClosed   CircuitState = iota // Normal operation
	CircuitStateOpen                         // Circuit is open, failing fast
	CircuitStateHalfOpen                     // Testing if service is back
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu sync.RWMutex

	// Configuration
	failureThreshold int           // Number of failures before opening
	timeout          time.Duration // How long to stay open
	successThreshold int           // Successes needed to close from half-open

	// State
	state           CircuitState
	failureCount    int
	successCount    int
	lastFailTime    time.Time
	lastSuccessTime time.Time
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold int, timeout time.Duration, successThreshold int) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		timeout:          timeout,
		successThreshold: successThreshold,
		state:            CircuitStateClosed,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, operation func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if circuit should transition
	cb.checkState()

	switch cb.state {
	case CircuitStateOpen:
		return errors.New("circuit breaker is open")
	case CircuitStateHalfOpen:
		// Allow limited requests to test if service is back
		if cb.successCount >= cb.successThreshold {
			cb.state = CircuitStateClosed
			cb.failureCount = 0
			cb.successCount = 0
		}
	}

	// Execute the operation
	err := operation()

	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// checkState checks if the circuit state should change
func (cb *CircuitBreaker) checkState() {
	switch cb.state {
	case CircuitStateOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.state = CircuitStateHalfOpen
			cb.successCount = 0
		}
	case CircuitStateHalfOpen:
		// Check if we have enough successes
		if cb.successCount >= cb.successThreshold {
			cb.state = CircuitStateClosed
			cb.failureCount = 0
			cb.successCount = 0
		}
	}
}

// recordFailure records a failure
func (cb *CircuitBreaker) recordFailure() {
	cb.failureCount++
	cb.lastFailTime = time.Now()

	if cb.failureCount >= cb.failureThreshold {
		cb.state = CircuitStateOpen
	}
}

// recordSuccess records a success
func (cb *CircuitBreaker) recordSuccess() {
	cb.successCount++
	cb.lastSuccessTime = time.Now()

	if cb.state == CircuitStateHalfOpen {
		// In half-open state, we're testing if service is back
		return
	}

	// In closed state, reset failure count on success
	if cb.failureCount > 0 {
		cb.failureCount = 0
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	cb.checkState()
	return cb.state
}

// GetMetrics returns circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	stateStr := "closed"
	switch cb.state {
	case CircuitStateOpen:
		stateStr = "open"
	case CircuitStateHalfOpen:
		stateStr = "half-open"
	}

	return map[string]interface{}{
		"state":             stateStr,
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"last_fail_time":    cb.lastFailTime,
		"last_success_time": cb.lastSuccessTime,
		"failure_threshold": cb.failureThreshold,
		"timeout":           cb.timeout.String(),
		"success_threshold": cb.successThreshold,
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitStateClosed
	cb.failureCount = 0
	cb.successCount = 0
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// GetOrCreate gets an existing circuit breaker or creates a new one
func (cbm *CircuitBreakerManager) GetOrCreate(name string, failureThreshold int, timeout time.Duration, successThreshold int) *CircuitBreaker {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	if breaker, exists := cbm.breakers[name]; exists {
		return breaker
	}

	breaker := NewCircuitBreaker(failureThreshold, timeout, successThreshold)
	cbm.breakers[name] = breaker
	return breaker
}

// GetMetrics returns metrics for all circuit breakers
func (cbm *CircuitBreakerManager) GetMetrics() map[string]interface{} {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	metrics := make(map[string]interface{})
	for name, breaker := range cbm.breakers {
		metrics[name] = breaker.GetMetrics()
	}

	return metrics
}
