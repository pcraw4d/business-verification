package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// CircuitClosed means the circuit is closed and requests are allowed
	CircuitClosed CircuitState = iota
	// CircuitOpen means the circuit is open and requests are immediately rejected
	CircuitOpen
	// CircuitHalfOpen means the circuit is half-open and testing if service has recovered
	CircuitHalfOpen
)

// String returns a string representation of the circuit state
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig holds configuration for circuit breaker behavior
type CircuitBreakerConfig struct {
	FailureThreshold    int           // Number of failures before opening circuit
	SuccessThreshold    int           // Number of successes in half-open state to close circuit
	Timeout             time.Duration // Time to wait before attempting half-open
	MaxRequests         int           // Max requests in half-open state (default: 1)
	ResetTimeout        time.Duration // Time to wait before resetting failure count
}

// DefaultCircuitBreakerConfig returns a default circuit breaker configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          30 * time.Second,
		MaxRequests:      1,
		ResetTimeout:     60 * time.Second,
	}
}

// CircuitBreaker implements the circuit breaker pattern to prevent cascading failures
type CircuitBreaker struct {
	config        CircuitBreakerConfig
	state         CircuitState
	failureCount  int
	successCount  int
	halfOpenCount int
	lastFailure   time.Time
	stateChange   time.Time
	mu            sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config:      config,
		state:       CircuitClosed,
		stateChange: time.Now(),
	}
}

// Execute executes a function through the circuit breaker
//
// If the circuit is open, it immediately returns an error without executing the function.
// If the circuit is half-open, it executes the function and transitions based on the result.
// If the circuit is closed, it executes the function and tracks failures/successes.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check if we should allow the request
	if !cb.allowRequest() {
		return fmt.Errorf("circuit breaker is %s", cb.getState())
	}
	
	// Execute the function
	err := fn()
	
	// Update circuit state based on result
	cb.onResult(err)
	
	return err
}

// ExecuteWithResult executes a function that returns a result through the circuit breaker
func ExecuteWithResult[T any](cb *CircuitBreaker, ctx context.Context, fn func() (T, error)) (T, error) {
	var zero T
	
	// Check if we should allow the request
	if !cb.allowRequest() {
		return zero, fmt.Errorf("circuit breaker is %s", cb.getState())
	}
	
	// Execute the function
	result, err := fn()
	
	// Update circuit state based on result
	cb.onResult(err)
	
	return result, err
}

// allowRequest checks if a request should be allowed based on current circuit state
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if timeout has elapsed to transition to half-open
		if time.Since(cb.stateChange) >= cb.config.Timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			if cb.state == CircuitOpen && time.Since(cb.stateChange) >= cb.config.Timeout {
				cb.state = CircuitHalfOpen
				cb.stateChange = time.Now()
				cb.halfOpenCount = 0
			}
			cb.mu.Unlock()
			cb.mu.RLock()
			return cb.state == CircuitHalfOpen
		}
		return false
	case CircuitHalfOpen:
		// Allow limited requests in half-open state
		return cb.halfOpenCount < cb.config.MaxRequests
	default:
		return false
	}
}

// onResult updates the circuit breaker state based on the function result
func (cb *CircuitBreaker) onResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailure = time.Now()
	
	switch cb.state {
	case CircuitClosed:
		// Open circuit if failure threshold reached
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.state = CircuitOpen
			cb.stateChange = time.Now()
		}
	case CircuitHalfOpen:
		// Immediately open circuit on failure in half-open state
		cb.state = CircuitOpen
		cb.stateChange = time.Now()
		cb.halfOpenCount = 0
		cb.successCount = 0
	}
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess() {
	cb.failureCount = 0
	
	switch cb.state {
	case CircuitHalfOpen:
		cb.successCount++
		cb.halfOpenCount++
		// Close circuit if success threshold reached
		if cb.successCount >= cb.config.SuccessThreshold {
			cb.state = CircuitClosed
			cb.stateChange = time.Now()
			cb.successCount = 0
			cb.halfOpenCount = 0
		}
	case CircuitClosed:
		// Reset failure count after reset timeout
		if time.Since(cb.lastFailure) >= cb.config.ResetTimeout {
			cb.failureCount = 0
		}
	}
}

// getState returns the current circuit breaker state (thread-safe)
func (cb *CircuitBreaker) getState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	return cb.getState()
}

// GetStats returns statistics about the circuit breaker
func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	return CircuitBreakerStats{
		State:         cb.state.String(),
		FailureCount:  cb.failureCount,
		SuccessCount:  cb.successCount,
		HalfOpenCount: cb.halfOpenCount,
		LastFailure:   cb.lastFailure,
		StateChange:   cb.stateChange,
	}
}

// CircuitBreakerStats holds statistics about circuit breaker state
type CircuitBreakerStats struct {
	State         string
	FailureCount  int
	SuccessCount  int
	HalfOpenCount int
	LastFailure   time.Time
	StateChange   time.Time
}

