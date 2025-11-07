package resilience

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxAttempts  int           // Maximum number of retry attempts
	InitialDelay time.Duration // Initial delay before first retry
	MaxDelay     time.Duration // Maximum delay between retries
	Multiplier   float64       // Exponential backoff multiplier (default: 2.0)
	Jitter       bool          // Enable jitter to prevent thundering herd
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// RetryWithBackoff executes a function with exponential backoff retry logic.
//
// The function will be retried up to MaxAttempts times if it returns an error.
// Delays between retries follow an exponential backoff pattern with optional jitter.
//
// Example:
//
//	result, err := RetryWithBackoff(ctx, DefaultRetryConfig(), func() (interface{}, error) {
//	    return someOperation()
//	})
func RetryWithBackoff[T any](ctx context.Context, config RetryConfig, fn func() (T, error)) (T, error) {
	var zero T
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return zero, fmt.Errorf("context cancelled: %w", ctx.Err())
		}

		// Execute the function
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't wait after the last attempt
		if attempt < config.MaxAttempts-1 {
			// Calculate delay with exponential backoff
			delay := config.InitialDelay * time.Duration(math.Pow(config.Multiplier, float64(attempt)))

			// Cap at max delay
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}

			// Add jitter if enabled (up to 25% of delay)
			if config.Jitter {
				jitterAmount := time.Duration(rand.Float64() * 0.25 * float64(delay))
				delay += jitterAmount
			}

			// Wait before retry
			select {
			case <-ctx.Done():
				return zero, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	// All attempts failed
	return zero, fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// RetryWithBackoffSimple is a convenience function that uses default retry configuration.
//
// Example:
//
//	result, err := RetryWithBackoffSimple(ctx, 3, 100*time.Millisecond, func() (interface{}, error) {
//	    return someOperation()
//	})
func RetryWithBackoffSimple[T any](ctx context.Context, maxAttempts int, initialDelay time.Duration, fn func() (T, error)) (T, error) {
	config := RetryConfig{
		MaxAttempts:  maxAttempts,
		InitialDelay: initialDelay,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
	return RetryWithBackoff(ctx, config, fn)
}
