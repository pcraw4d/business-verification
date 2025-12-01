package retry

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// RetryableError indicates if an error should be retried
type RetryableError struct {
	Error      error
	Retryable  bool
	RetryCount int // Suggested retry count based on error type
}

// AdaptiveRetryStrategy provides intelligent retry logic based on error types
type AdaptiveRetryStrategy struct {
	// Error history tracking
	errorHistory map[string]*errorStats
	historyMutex sync.RWMutex
	
	// Default retry configuration
	defaultMaxRetries int
	defaultBackoff    time.Duration
}

type errorStats struct {
	TotalAttempts int
	SuccessCount  int
	FailureCount  int
	LastSuccess   time.Time
	LastFailure   time.Time
}

// NewAdaptiveRetryStrategy creates a new adaptive retry strategy
func NewAdaptiveRetryStrategy(defaultMaxRetries int, defaultBackoff time.Duration) *AdaptiveRetryStrategy {
	return &AdaptiveRetryStrategy{
		errorHistory:      make(map[string]*errorStats),
		defaultMaxRetries: defaultMaxRetries,
		defaultBackoff:    defaultBackoff,
	}
}

// ShouldRetry determines if an error should be retried and how many times
func (ars *AdaptiveRetryStrategy) ShouldRetry(err error, httpStatusCode int) (bool, int) {
	// Permanent errors - never retry
	if httpStatusCode > 0 {
		if httpStatusCode == 400 || httpStatusCode == 403 || httpStatusCode == 404 {
			return false, 0
		}
		// 429 (Rate Limited) - retry with more attempts
		if httpStatusCode == 429 {
			return true, 5 // More retries for rate limiting
		}
		// 500, 502, 503, 504 - retry with standard attempts
		if httpStatusCode >= 500 {
			return true, ars.defaultMaxRetries
		}
	}

	// Check error type
	if err == nil {
		return false, 0
	}

	// DNS errors - retry with more attempts
	if _, ok := err.(*net.DNSError); ok {
		return true, ars.defaultMaxRetries + 1
	}

	// Timeout errors - retry with fewer attempts (likely network issue)
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true, ars.defaultMaxRetries
	}

	// Network errors - retry
	if _, ok := err.(net.Error); ok {
		return true, ars.defaultMaxRetries
	}

	// Unknown errors - check history
	errorKey := ars.getErrorKey(err, httpStatusCode)
	ars.historyMutex.RLock()
	stats, exists := ars.errorHistory[errorKey]
	ars.historyMutex.RUnlock()

	if exists {
		// Calculate success rate
		total := stats.SuccessCount + stats.FailureCount
		if total > 0 {
			successRate := float64(stats.SuccessCount) / float64(total)
			// If success rate is very low (< 20%), reduce retries
			if successRate < 0.2 {
				return true, 1 // Only 1 retry
			}
			// If success rate is high (> 80%), use default retries
			if successRate > 0.8 {
				return true, ars.defaultMaxRetries
			}
		}
	}

	// Default: retry with standard attempts
	return true, ars.defaultMaxRetries
}

// RecordResult records the result of a retry attempt for learning
func (ars *AdaptiveRetryStrategy) RecordResult(err error, httpStatusCode int, success bool) {
	errorKey := ars.getErrorKey(err, httpStatusCode)
	
	ars.historyMutex.Lock()
	defer ars.historyMutex.Unlock()

	stats, exists := ars.errorHistory[errorKey]
	if !exists {
		stats = &errorStats{}
		ars.errorHistory[errorKey] = stats
	}

	stats.TotalAttempts++
	if success {
		stats.SuccessCount++
		stats.LastSuccess = time.Now()
	} else {
		stats.FailureCount++
		stats.LastFailure = time.Now()
	}

	// Clean up old entries (older than 1 hour)
	now := time.Now()
	for key, s := range ars.errorHistory {
		if now.Sub(s.LastFailure) > 1*time.Hour && now.Sub(s.LastSuccess) > 1*time.Hour {
			delete(ars.errorHistory, key)
		}
	}
}

// CalculateBackoff calculates exponential backoff with jitter
func (ars *AdaptiveRetryStrategy) CalculateBackoff(attempt int, baseDelay time.Duration) time.Duration {
	// Exponential backoff: baseDelay * 2^(attempt-1)
	backoff := baseDelay * time.Duration(1<<uint(attempt-1))
	
	// Add jitter (±20%)
	jitter := time.Duration(float64(backoff) * 0.2)
	// For simplicity, we'll use a fixed jitter range
	if jitter > 2*time.Second {
		jitter = 2 * time.Second
	}
	
	// Add random jitter (simplified - in production, use crypto/rand)
	// For now, use attempt-based pseudo-random jitter
	jitterValue := time.Duration((attempt % 3) - 1) * jitter / 3
	backoff = backoff + jitterValue
	
	// Cap at 30 seconds
	if backoff > 30*time.Second {
		backoff = 30 * time.Second
	}
	
	return backoff
}

// getErrorKey generates a key for error tracking
func (ars *AdaptiveRetryStrategy) getErrorKey(err error, httpStatusCode int) string {
	if httpStatusCode > 0 {
		return fmt.Sprintf("http_%d", httpStatusCode)
	}
	if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok {
			return fmt.Sprintf("dns_%s", dnsErr.Server)
		}
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				return "timeout"
			}
			return "network"
		}
		return fmt.Sprintf("error_%T", err)
	}
	return "unknown"
}

// RetryWithStrategy executes a function with adaptive retry strategy
func RetryWithStrategy(
	ctx context.Context,
	strategy *AdaptiveRetryStrategy,
	operation func() error,
	httpStatusCode *int, // Optional: pointer to HTTP status code
) error {
	var lastErr error
	maxRetries := strategy.defaultMaxRetries
	
	// Determine retry strategy based on first attempt
	firstErr := operation()
	if firstErr == nil {
		strategy.RecordResult(firstErr, getStatusCode(httpStatusCode), true)
		return nil
	}

	// Check if we should retry
	retryable, retryCount := strategy.ShouldRetry(firstErr, getStatusCode(httpStatusCode))
	if !retryable {
		strategy.RecordResult(firstErr, getStatusCode(httpStatusCode), false)
		return firstErr
	}

	if retryCount > 0 {
		maxRetries = retryCount
	}

	lastErr = firstErr

	// Retry loop
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Calculate backoff
		backoff := strategy.CalculateBackoff(attempt, strategy.defaultBackoff)
		
		// Wait before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}

		// Retry operation
		err := operation()
		if err == nil {
			strategy.RecordResult(err, getStatusCode(httpStatusCode), true)
			return nil
		}

		lastErr = err

		// Check if we should continue retrying
		retryable, _ := strategy.ShouldRetry(err, getStatusCode(httpStatusCode))
		if !retryable {
			strategy.RecordResult(err, getStatusCode(httpStatusCode), false)
			break
		}
	}

	strategy.RecordResult(lastErr, getStatusCode(httpStatusCode), false)
	return fmt.Errorf("operation failed after %d attempts: %w", maxRetries+1, lastErr)
}

// getStatusCode safely gets status code from pointer
func getStatusCode(statusCode *int) int {
	if statusCode == nil {
		return 0
	}
	return *statusCode
}

