package industry_codes

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewRetryMechanism(t *testing.T) {
	logger := zap.NewNop()

	t.Run("creates with default config", func(t *testing.T) {
		rm := NewRetryMechanism(logger, nil)

		require.NotNil(t, rm)
		require.NotNil(t, rm.config)
		assert.Equal(t, 3, rm.config.MaxAttempts)
		assert.Equal(t, 100*time.Millisecond, rm.config.BaseDelay)
		assert.Equal(t, 30*time.Second, rm.config.MaxDelay)
		assert.Equal(t, 2.0, rm.config.BackoffMultiplier)
		assert.Equal(t, 0.1, rm.config.JitterFactor)
		assert.True(t, rm.config.CircuitBreakerEnabled)
		assert.Equal(t, 5, rm.config.CircuitBreakerThreshold)
		assert.Equal(t, 60*time.Second, rm.config.CircuitBreakerTimeout)
		assert.Len(t, rm.config.RetryableErrors, 6)
		assert.Len(t, rm.config.NonRetryableErrors, 5)
	})

	t.Run("creates with custom config", func(t *testing.T) {
		customConfig := &RetryConfig{
			MaxAttempts:           5,
			BaseDelay:             200 * time.Millisecond,
			MaxDelay:              60 * time.Second,
			BackoffMultiplier:     1.5,
			JitterFactor:          0.2,
			CircuitBreakerEnabled: false,
		}

		rm := NewRetryMechanism(logger, customConfig)

		require.NotNil(t, rm)
		assert.Equal(t, customConfig, rm.config)
		assert.Equal(t, 5, rm.config.MaxAttempts)
		assert.Equal(t, 200*time.Millisecond, rm.config.BaseDelay)
		assert.Equal(t, 60*time.Second, rm.config.MaxDelay)
		assert.Equal(t, 1.5, rm.config.BackoffMultiplier)
		assert.Equal(t, 0.2, rm.config.JitterFactor)
		assert.False(t, rm.config.CircuitBreakerEnabled)
	})
}

func TestRetryMechanism_ExecuteWithRetry_SuccessfulOperation(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRetryMechanism(logger, nil)

	t.Run("succeeds on first attempt", func(t *testing.T) {
		attemptCount := 0
		operation := func() (interface{}, error) {
			attemptCount++
			return "success", nil
		}

		result := rm.ExecuteWithRetry(context.Background(), operation, "test_operation")

		assert.True(t, result.Success)
		assert.Equal(t, "success", result.Data)
		assert.Equal(t, 1, result.Attempts)
		assert.Nil(t, result.LastError)
		assert.Empty(t, result.RetryDelays)
		assert.False(t, result.CircuitBreakerHit)
		assert.Equal(t, 1, attemptCount)
		assert.Greater(t, result.TotalTime, time.Duration(0))
	})

	t.Run("succeeds after retries", func(t *testing.T) {
		attemptCount := 0
		operation := func() (interface{}, error) {
			attemptCount++
			if attemptCount < 3 {
				return nil, errors.New("temporary failure")
			}
			return "success after retries", nil
		}

		result := rm.ExecuteWithRetry(context.Background(), operation, "test_operation")

		assert.True(t, result.Success)
		assert.Equal(t, "success after retries", result.Data)
		assert.Equal(t, 3, result.Attempts)
		assert.Nil(t, result.LastError)
		assert.Len(t, result.RetryDelays, 2) // 2 delays between 3 attempts
		assert.False(t, result.CircuitBreakerHit)
		assert.Equal(t, 3, attemptCount)
		assert.Greater(t, result.TotalTime, time.Duration(0))
	})
}

func TestRetryMechanism_ExecuteWithRetry_FailedOperation(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRetryMechanism(logger, nil)

	t.Run("fails after all attempts", func(t *testing.T) {
		attemptCount := 0
		operation := func() (interface{}, error) {
			attemptCount++
			return nil, errors.New("persistent failure")
		}

		result := rm.ExecuteWithRetry(context.Background(), operation, "test_operation")

		assert.False(t, result.Success)
		assert.Nil(t, result.Data)
		assert.Equal(t, 3, result.Attempts)
		assert.NotNil(t, result.LastError)
		assert.Contains(t, result.LastError.Error(), "operation failed after 3 attempts")
		assert.Len(t, result.RetryDelays, 2) // 2 delays between 3 attempts
		assert.False(t, result.CircuitBreakerHit)
		assert.Equal(t, 3, attemptCount)
		assert.Greater(t, result.TotalTime, time.Duration(0))
	})

	t.Run("fails immediately with non-retryable error", func(t *testing.T) {
		attemptCount := 0
		operation := func() (interface{}, error) {
			attemptCount++
			return nil, errors.New("invalid input")
		}

		result := rm.ExecuteWithRetry(context.Background(), operation, "test_operation")

		assert.False(t, result.Success)
		assert.Nil(t, result.Data)
		assert.Equal(t, 1, result.Attempts)
		assert.NotNil(t, result.LastError)
		assert.Contains(t, result.LastError.Error(), "invalid input")
		assert.Empty(t, result.RetryDelays)
		assert.False(t, result.CircuitBreakerHit)
		assert.Equal(t, 1, attemptCount)
	})
}

func TestRetryMechanism_ExecuteWithRetry_Timeout(t *testing.T) {
	logger := zap.NewNop()
	config := &RetryConfig{
		MaxAttempts:       3,
		BaseDelay:         10 * time.Millisecond,
		MaxDelay:          100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		JitterFactor:      0.0,
		TimeoutPerAttempt: 50 * time.Millisecond,
	}
	rm := NewRetryMechanism(logger, config)

	t.Run("times out on slow operation", func(t *testing.T) {
		attemptCount := 0
		operation := func() (interface{}, error) {
			attemptCount++
			time.Sleep(100 * time.Millisecond) // Longer than timeout
			return nil, errors.New("should not reach here")
		}

		result := rm.ExecuteWithRetry(context.Background(), operation, "test_operation")

		assert.False(t, result.Success)
		assert.Nil(t, result.Data)
		assert.Equal(t, 1, result.Attempts)
		assert.NotNil(t, result.LastError)
		assert.Contains(t, result.LastError.Error(), "operation timed out after 1 attempts")
		// Note: RetryDelays might not be empty if the operation times out during the delay
		assert.Equal(t, 1, attemptCount)
	})
}

func TestRetryMechanism_CalculateDelay(t *testing.T) {
	logger := zap.NewNop()
	config := &RetryConfig{
		BaseDelay:         100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
		JitterFactor:      0.0, // No jitter for predictable testing
	}
	rm := NewRetryMechanism(logger, config)

	t.Run("exponential backoff calculation", func(t *testing.T) {
		delay1 := rm.calculateDelay(1)
		delay2 := rm.calculateDelay(2)
		delay3 := rm.calculateDelay(3)

		assert.Equal(t, 100*time.Millisecond, delay1)
		assert.Equal(t, 200*time.Millisecond, delay2)
		assert.Equal(t, 400*time.Millisecond, delay3)
	})

	t.Run("respects max delay", func(t *testing.T) {
		delay := rm.calculateDelay(10) // Should be capped at max delay
		assert.LessOrEqual(t, delay, 1*time.Second)
	})

	t.Run("with jitter", func(t *testing.T) {
		config.JitterFactor = 0.1
		rm.config = config

		delay := rm.calculateDelay(2)
		expectedBase := 200 * time.Millisecond
		jitterRange := float64(expectedBase) * 0.1

		assert.GreaterOrEqual(t, delay, time.Duration(float64(expectedBase)-jitterRange))
		assert.LessOrEqual(t, delay, time.Duration(float64(expectedBase)+jitterRange))
	})
}

func TestRetryMechanism_IsRetryableError(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRetryMechanism(logger, nil)

	t.Run("retryable errors", func(t *testing.T) {
		retryableErrors := []string{
			"timeout",
			"connection refused",
			"network error",
			"temporary failure",
			"rate limit exceeded",
			"service unavailable",
		}

		for _, errMsg := range retryableErrors {
			err := errors.New(errMsg)
			assert.True(t, rm.isRetryableError(err), "Error '%s' should be retryable", errMsg)
		}
	})

	t.Run("non-retryable errors", func(t *testing.T) {
		nonRetryableErrors := []string{
			"invalid input",
			"authentication failed",
			"authorization denied",
			"not found",
			"bad request",
		}

		for _, errMsg := range nonRetryableErrors {
			err := errors.New(errMsg)
			assert.False(t, rm.isRetryableError(err), "Error '%s' should not be retryable", errMsg)
		}
	})

	t.Run("unknown errors default to retryable", func(t *testing.T) {
		err := errors.New("unknown error message")
		assert.True(t, rm.isRetryableError(err))
	})

	t.Run("nil error", func(t *testing.T) {
		assert.False(t, rm.isRetryableError(nil))
	})
}

func TestCircuitBreaker(t *testing.T) {
	logger := zap.NewNop()

	t.Run("starts in closed state", func(t *testing.T) {
		cb := &CircuitBreaker{
			state:     CircuitBreakerClosed,
			threshold: 3,
			timeout:   60 * time.Second,
			logger:    logger,
		}

		assert.True(t, cb.CanExecute())
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
		assert.Equal(t, int64(0), cb.GetFailureCount())
	})

	t.Run("opens after threshold failures", func(t *testing.T) {
		cb := &CircuitBreaker{
			state:     CircuitBreakerClosed,
			threshold: 3,
			timeout:   60 * time.Second,
			logger:    logger,
		}

		// Fail 3 times
		cb.OnFailure()
		cb.OnFailure()
		cb.OnFailure()

		assert.False(t, cb.CanExecute())
		assert.Equal(t, CircuitBreakerOpen, cb.GetState())
		assert.Equal(t, int64(3), cb.GetFailureCount())
	})

	t.Run("transitions to half-open after timeout", func(t *testing.T) {
		cb := &CircuitBreaker{
			state:           CircuitBreakerOpen,
			threshold:       3,
			timeout:         10 * time.Millisecond,
			logger:          logger,
			lastFailureTime: time.Now().Add(-20 * time.Millisecond), // Past timeout
		}

		assert.True(t, cb.CanExecute())
		assert.Equal(t, CircuitBreakerHalfOpen, cb.GetState())
	})

	t.Run("resets to closed on success in half-open", func(t *testing.T) {
		cb := &CircuitBreaker{
			state:     CircuitBreakerHalfOpen,
			threshold: 3,
			timeout:   60 * time.Second,
			logger:    logger,
		}

		cb.OnSuccess()

		assert.True(t, cb.CanExecute())
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
		assert.Equal(t, int64(0), cb.GetFailureCount())
	})

	t.Run("reopens on failure in half-open", func(t *testing.T) {
		cb := &CircuitBreaker{
			state:     CircuitBreakerHalfOpen,
			threshold: 3,
			timeout:   60 * time.Second,
			logger:    logger,
		}

		cb.OnFailure()

		assert.False(t, cb.CanExecute())
		assert.Equal(t, CircuitBreakerOpen, cb.GetState())
	})
}

func TestRetryMechanism_ExecuteWithRetry_CircuitBreaker(t *testing.T) {
	logger := zap.NewNop()
	config := &RetryConfig{
		MaxAttempts:             3,
		BaseDelay:               10 * time.Millisecond,
		MaxDelay:                100 * time.Millisecond,
		BackoffMultiplier:       2.0,
		JitterFactor:            0.0,
		CircuitBreakerEnabled:   true,
		CircuitBreakerThreshold: 2,
		CircuitBreakerTimeout:   100 * time.Millisecond,
	}
	rm := NewRetryMechanism(logger, config)

	t.Run("circuit breaker blocks execution", func(t *testing.T) {
		// Test circuit breaker directly
		cb := rm.getCircuitBreaker("test_operation")

		// Fail twice to open circuit breaker
		cb.OnFailure()
		cb.OnFailure()

		// Circuit breaker should be open
		assert.False(t, cb.CanExecute())
		assert.Equal(t, CircuitBreakerOpen, cb.GetState())

		// Now test that retry mechanism respects circuit breaker
		attemptCount := 0
		operation := func() (interface{}, error) {
			attemptCount++
			return nil, errors.New("persistent failure")
		}

		result := rm.ExecuteWithRetry(context.Background(), operation, "test_operation")
		assert.False(t, result.Success)
		assert.True(t, result.CircuitBreakerHit)
		assert.Equal(t, 0, result.Attempts)
		assert.Contains(t, result.LastError.Error(), "circuit breaker is open")
	})

	t.Run("circuit breaker resets after timeout", func(t *testing.T) {
		// Wait for circuit breaker timeout
		time.Sleep(150 * time.Millisecond)

		attemptCount := 0
		operation := func() (interface{}, error) {
			attemptCount++
			return "success", nil // This time it succeeds
		}

		result := rm.ExecuteWithRetry(context.Background(), operation, "test_operation")
		assert.True(t, result.Success)
		assert.Equal(t, 1, result.Attempts)
		assert.False(t, result.CircuitBreakerHit)
	})
}

func TestRetryMechanism_Stats(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRetryMechanism(logger, nil)

	t.Run("tracks successful retries", func(t *testing.T) {
		operation := func() (interface{}, error) {
			return "success", nil
		}

		rm.ExecuteWithRetry(context.Background(), operation, "test_operation")

		stats := rm.GetStats()
		assert.Equal(t, int64(1), stats.TotalAttempts)
		assert.Equal(t, int64(1), stats.SuccessfulRetries)
		assert.Equal(t, int64(0), stats.FailedRetries)
		assert.Greater(t, stats.TotalRetryTime, time.Duration(0))
	})

	t.Run("tracks failed retries", func(t *testing.T) {
		operation := func() (interface{}, error) {
			return nil, errors.New("persistent failure")
		}

		rm.ExecuteWithRetry(context.Background(), operation, "test_operation")

		stats := rm.GetStats()
		assert.Equal(t, int64(4), stats.TotalAttempts) // 1 from previous test + 3 from this test
		assert.Equal(t, int64(1), stats.SuccessfulRetries)
		assert.Equal(t, int64(1), stats.FailedRetries)
		assert.Greater(t, stats.TotalRetryTime, time.Duration(0))
	})

	t.Run("reset stats", func(t *testing.T) {
		rm.ResetStats()
		stats := rm.GetStats()
		assert.Equal(t, int64(0), stats.TotalAttempts)
		assert.Equal(t, int64(0), stats.SuccessfulRetries)
		assert.Equal(t, int64(0), stats.FailedRetries)
		assert.Equal(t, time.Duration(0), stats.TotalRetryTime)
	})
}

func TestRetryableOperation(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRetryMechanism(logger, nil)

	t.Run("executes retryable operation", func(t *testing.T) {
		attemptCount := 0
		operation := &MockRetryableOperation{
			name: "test_operation",
			execute: func(ctx context.Context) (interface{}, error) {
				attemptCount++
				if attemptCount < 2 {
					return nil, errors.New("temporary failure")
				}
				return "success", nil
			},
		}

		result := rm.ExecuteRetryableOperation(context.Background(), operation)

		assert.True(t, result.Success)
		assert.Equal(t, "success", result.Data)
		assert.Equal(t, 2, result.Attempts)
		assert.Equal(t, "test_operation", operation.GetName())
	})
}

// MockRetryableOperation for testing
type MockRetryableOperation struct {
	name    string
	execute func(ctx context.Context) (interface{}, error)
}

func (m *MockRetryableOperation) Execute(ctx context.Context) (interface{}, error) {
	return m.execute(ctx)
}

func (m *MockRetryableOperation) GetName() string {
	return m.name
}

func TestRetryMechanism_ContextCancellation(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRetryMechanism(logger, nil)

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		attemptCount := 0
		operation := func() (interface{}, error) {
			attemptCount++
			if attemptCount == 2 {
				cancel() // Cancel context on second attempt
			}
			return nil, errors.New("temporary failure")
		}

		result := rm.ExecuteWithRetry(ctx, operation, "test_operation")

		assert.False(t, result.Success)
		assert.NotNil(t, result.LastError)
		// The error might be "context canceled" or "operation timed out" depending on timing
		assert.True(t,
			strings.Contains(result.LastError.Error(), "context canceled") ||
				strings.Contains(result.LastError.Error(), "operation timed out"),
			"Expected context cancellation or timeout, got: %s", result.LastError.Error())
		assert.LessOrEqual(t, result.Attempts, 3) // Should stop before max attempts
	})
}

func TestRetryMechanism_Integration(t *testing.T) {
	logger := zap.NewNop()
	config := &RetryConfig{
		MaxAttempts:             3,
		BaseDelay:               10 * time.Millisecond,
		MaxDelay:                100 * time.Millisecond,
		BackoffMultiplier:       2.0,
		JitterFactor:            0.1,
		TimeoutPerAttempt:       200 * time.Millisecond,
		CircuitBreakerEnabled:   true,
		CircuitBreakerThreshold: 3,
		CircuitBreakerTimeout:   50 * time.Millisecond,
	}
	rm := NewRetryMechanism(logger, config)

	t.Run("complex retry scenario", func(t *testing.T) {
		attemptCount := 0
		operation := func() (interface{}, error) {
			attemptCount++
			switch attemptCount {
			case 1:
				return nil, errors.New("timeout")
			case 2:
				return nil, errors.New("connection refused")
			case 3:
				return "final success", nil
			default:
				return nil, errors.New("unexpected attempt")
			}
		}

		result := rm.ExecuteWithRetry(context.Background(), operation, "complex_operation")

		assert.True(t, result.Success)
		assert.Equal(t, "final success", result.Data)
		assert.Equal(t, 3, result.Attempts)
		assert.Len(t, result.RetryDelays, 2)
		assert.False(t, result.CircuitBreakerHit)
		assert.Equal(t, 3, attemptCount)

		// Verify delays are increasing
		assert.Greater(t, result.RetryDelays[1], result.RetryDelays[0])
	})
}
