package error_resilience

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestErrorResilienceManager_NewErrorResilienceManager(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)

	manager := NewErrorResilienceManager(logger)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.circuitBreakers)
	assert.NotNil(t, manager.retryPolicies)
	assert.NotNil(t, manager.fallbackStrategies)
	assert.NotNil(t, manager.degradationPolicies)
	assert.NotNil(t, manager.logger)
	assert.NotNil(t, manager.metrics)
	assert.Equal(t, 0, len(manager.circuitBreakers))
	assert.Equal(t, 0, len(manager.retryPolicies))
	assert.Equal(t, 0, len(manager.fallbackStrategies))
	assert.Equal(t, 0, len(manager.degradationPolicies))
}

func TestErrorResilienceManager_RegisterCircuitBreaker(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Register circuit breaker
	manager.RegisterCircuitBreaker("test-module", 3, 2, 30*time.Second)

	// Verify circuit breaker is registered
	state := manager.GetCircuitBreakerState("test-module")
	assert.NotNil(t, state)
	assert.Equal(t, "test-module", state["name"])
	assert.Equal(t, CircuitBreakerClosed, state["state"])
	assert.Equal(t, int64(0), state["failure_count"])
	assert.Equal(t, int64(0), state["success_count"])
}

func TestErrorResilienceManager_RegisterRetryPolicy(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Register retry policy
	retryableErrors := []string{"timeout", "connection", "temporary"}
	manager.RegisterRetryPolicy("test-module", 3, 1*time.Second, 10*time.Second, 2.0, retryableErrors)

	// Verify retry policy is registered (indirectly through execution)
	result := manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return "success", nil
	})

	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "success", result.Data)
	assert.Equal(t, DegradationLevelNone, result.DegradationLevel)
	assert.Equal(t, 1.0, result.Confidence)
}

func TestErrorResilienceManager_RegisterFallbackStrategy(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Register fallback strategy
	fallbackData := map[string]interface{}{
		"default_response": "fallback data",
	}
	manager.RegisterFallbackStrategy("test-module", true, "static_data", fallbackData, "", DegradationLevelFallback)

	// Execute with failure to trigger fallback
	result := manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return nil, fmt.Errorf("module failure")
	})

	assert.NotNil(t, result)
	assert.True(t, result.Success) // Should succeed due to fallback
	assert.Equal(t, DegradationLevelFallback, result.DegradationLevel)
	assert.Equal(t, 0.7, result.Confidence) // Lower confidence for fallback
	assert.True(t, result.FallbackUsed)
}

func TestErrorResilienceManager_RegisterDegradationPolicy(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Register degradation policy
	degradationLevels := []DegradationLevel{DegradationLevelPartial, DegradationLevelMinimal}
	manager.RegisterDegradationPolicy("test-module", true, degradationLevels, 0.6, 0.2)

	// Execute with failure to trigger degradation
	result := manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return nil, fmt.Errorf("module failure")
	})

	assert.NotNil(t, result)
	// Should either succeed with degradation or fail completely
	assert.Contains(t, []DegradationLevel{DegradationLevelPartial, DegradationLevelMinimal, DegradationLevelFallback}, result.DegradationLevel)
}

func TestErrorResilienceManager_ExecuteWithResilience_Success(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Execute successful operation
	result := manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return "success", nil
	})

	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "success", result.Data)
	assert.Nil(t, result.Error)
	assert.Equal(t, DegradationLevelNone, result.DegradationLevel)
	assert.Equal(t, 1.0, result.Confidence)
	assert.False(t, result.FallbackUsed)
	assert.Equal(t, 0, result.RetryAttempts)
}

func TestErrorResilienceManager_ExecuteWithResilience_CircuitBreakerOpen(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Register circuit breaker with low threshold
	manager.RegisterCircuitBreaker("test-module", 1, 1, 1*time.Second)

	// Execute operation that fails to open circuit breaker
	result := manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return nil, fmt.Errorf("module failure")
	})

	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, DegradationLevelFallback, result.DegradationLevel)
	assert.Equal(t, 0.0, result.Confidence)

	// Try again - should be blocked by circuit breaker
	result2 := manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return "success", nil
	})

	assert.NotNil(t, result2)
	assert.False(t, result2.Success)
	assert.Contains(t, result2.Error.Error(), "circuit breaker open")
}

func TestErrorResilienceManager_ExecuteWithResilience_RetrySuccess(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Register retry policy
	manager.RegisterRetryPolicy("test-module", 3, 10*time.Millisecond, 100*time.Millisecond, 2.0, []string{"temporary"})

	attemptCount := 0
	result := manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		attemptCount++
		if attemptCount < 3 {
			return nil, fmt.Errorf("temporary failure")
		}
		return "success", nil
	})

	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "success", result.Data)
	assert.Equal(t, DegradationLevelNone, result.DegradationLevel)
	assert.Equal(t, 1.0, result.Confidence)
	assert.Equal(t, 3, result.RetryAttempts)
}

func TestErrorResilienceManager_ExecuteWithResilience_RetryFailure(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Register retry policy
	manager.RegisterRetryPolicy("test-module", 3, 10*time.Millisecond, 100*time.Millisecond, 2.0, []string{"temporary"})

	result := manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return nil, fmt.Errorf("permanent failure")
	})

	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "permanent failure", result.Error.Error())
	assert.Equal(t, DegradationLevelFallback, result.DegradationLevel)
	assert.Equal(t, 0.0, result.Confidence)
	assert.Equal(t, 1, result.RetryAttempts) // Should not retry non-retryable error
}

func TestErrorResilienceManager_ExecuteWithResilience_FallbackSuccess(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Register fallback strategy
	fallbackData := map[string]interface{}{
		"default_response": "fallback data",
	}
	manager.RegisterFallbackStrategy("test-module", true, "static_data", fallbackData, "", DegradationLevelFallback)

	result := manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return nil, fmt.Errorf("module failure")
	})

	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, fallbackData, result.Data)
	assert.Equal(t, DegradationLevelFallback, result.DegradationLevel)
	assert.Equal(t, 0.7, result.Confidence)
	assert.True(t, result.FallbackUsed)
}

func TestErrorResilienceManager_ExecuteWithResilience_GracefulDegradation(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Register degradation policy
	degradationLevels := []DegradationLevel{DegradationLevelPartial, DegradationLevelMinimal}
	manager.RegisterDegradationPolicy("test-module", true, degradationLevels, 0.6, 0.2)

	result := manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return nil, fmt.Errorf("module failure")
	})

	assert.NotNil(t, result)
	// Should succeed with degradation since thresholds are met
	assert.True(t, result.Success)
	assert.Contains(t, []DegradationLevel{DegradationLevelPartial, DegradationLevelMinimal}, result.DegradationLevel)
	assert.Greater(t, result.Confidence, 0.0)
	assert.False(t, result.FallbackUsed)
}

func TestErrorResilienceManager_GetMetrics(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Register policies
	manager.RegisterCircuitBreaker("test-module", 1, 1, 1*time.Second)
	manager.RegisterRetryPolicy("test-module", 2, 10*time.Millisecond, 100*time.Millisecond, 2.0, []string{"temporary"})
	manager.RegisterFallbackStrategy("test-module", true, "static_data", map[string]interface{}{"data": "fallback"}, "", DegradationLevelFallback)

	// Execute operations to generate metrics
	manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return nil, fmt.Errorf("temporary failure")
	})

	manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return nil, fmt.Errorf("permanent failure")
	})

	// Get metrics
	metrics := manager.GetMetrics()

	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics["retry_attempts"], int64(0))
	assert.GreaterOrEqual(t, metrics["fallback_executions"], int64(0))
	assert.GreaterOrEqual(t, metrics["circuit_breaker_trips"], int64(0))
}

func TestErrorResilienceManager_ResetCircuitBreaker(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Register circuit breaker
	manager.RegisterCircuitBreaker("test-module", 1, 1, 1*time.Second)

	// Open circuit breaker
	manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return nil, fmt.Errorf("module failure")
	})

	// Verify circuit breaker is open
	state := manager.GetCircuitBreakerState("test-module")
	assert.Equal(t, CircuitBreakerOpen, state["state"])

	// Reset circuit breaker
	err := manager.ResetCircuitBreaker("test-module")
	require.NoError(t, err)

	// Verify circuit breaker is closed
	state = manager.GetCircuitBreakerState("test-module")
	assert.Equal(t, CircuitBreakerClosed, state["state"])
	assert.Equal(t, int64(0), state["failure_count"])
	assert.Equal(t, int64(0), state["success_count"])
}

func TestErrorResilienceManager_ResetCircuitBreaker_NotFound(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Try to reset non-existent circuit breaker
	err := manager.ResetCircuitBreaker("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestErrorResilienceManager_GetCircuitBreakerState_NotFound(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Get state of non-existent circuit breaker
	state := manager.GetCircuitBreakerState("non-existent")
	assert.Nil(t, state)
}

// Test circuit breaker state transitions
func TestCircuitBreaker_StateTransitions(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Register circuit breaker with low thresholds for testing
	manager.RegisterCircuitBreaker("test-module", 2, 1, 1*time.Second)

	// First failure - should still be closed
	manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return nil, fmt.Errorf("failure 1")
	})

	state := manager.GetCircuitBreakerState("test-module")
	assert.Equal(t, CircuitBreakerClosed, state["state"])
	assert.Equal(t, int64(1), state["failure_count"])

	// Second failure - should open circuit breaker
	manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return nil, fmt.Errorf("failure 2")
	})

	state = manager.GetCircuitBreakerState("test-module")
	assert.Equal(t, CircuitBreakerOpen, state["state"])
	assert.Equal(t, int64(0), state["failure_count"]) // Reset after opening

	// Wait for timeout and try again - should go to half-open
	time.Sleep(2 * time.Second)

	result := manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return "success", nil
	})

	assert.True(t, result.Success)
	state = manager.GetCircuitBreakerState("test-module")
	assert.Equal(t, CircuitBreakerClosed, state["state"]) // Should close after success
}

// Test retry with exponential backoff
func TestRetryPolicy_ExponentialBackoff(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Register retry policy with exponential backoff
	manager.RegisterRetryPolicy("test-module", 3, 10*time.Millisecond, 100*time.Millisecond, 2.0, []string{"temporary"})

	start := time.Now()
	result := manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
		return nil, fmt.Errorf("temporary failure")
	})

	duration := time.Since(start)

	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, 3, result.RetryAttempts)

	// Should have taken at least the sum of delays: 10ms + 20ms = 30ms
	assert.GreaterOrEqual(t, duration, 30*time.Millisecond)
}

// Test context cancellation
func TestErrorResilienceManager_ContextCancellation(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Register retry policy
	manager.RegisterRetryPolicy("test-module", 5, 100*time.Millisecond, 1*time.Second, 2.0, []string{"temporary"})

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result := manager.ExecuteWithResilience(ctx, "test-module", func() (interface{}, error) {
		return nil, fmt.Errorf("temporary failure")
	})

	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, context.DeadlineExceeded, result.Error)
}

// Test concurrent operations
func TestErrorResilienceManager_ConcurrentOperations(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	manager := NewErrorResilienceManager(logger)

	// Register policies
	manager.RegisterCircuitBreaker("test-module", 5, 2, 1*time.Second)
	manager.RegisterRetryPolicy("test-module", 2, 10*time.Millisecond, 100*time.Millisecond, 2.0, []string{"temporary"})

	// Start multiple goroutines
	const numGoroutines = 10
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				result := manager.ExecuteWithResilience(context.Background(), "test-module", func() (interface{}, error) {
					if j%2 == 0 {
						return "success", nil
					}
					return nil, fmt.Errorf("temporary failure")
				})

				assert.NotNil(t, result)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify no panics occurred
	metrics := manager.GetMetrics()
	assert.NotNil(t, metrics)
}

// Test degradation level constants
func TestDegradationLevel_Constants(t *testing.T) {
	assert.Equal(t, DegradationLevel("none"), DegradationLevelNone)
	assert.Equal(t, DegradationLevel("partial"), DegradationLevelPartial)
	assert.Equal(t, DegradationLevel("minimal"), DegradationLevelMinimal)
	assert.Equal(t, DegradationLevel("fallback"), DegradationLevelFallback)
}

// Test circuit breaker state constants
func TestCircuitBreakerState_Constants(t *testing.T) {
	assert.Equal(t, CircuitBreakerState("closed"), CircuitBreakerClosed)
	assert.Equal(t, CircuitBreakerState("open"), CircuitBreakerOpen)
	assert.Equal(t, CircuitBreakerState("half_open"), CircuitBreakerHalfOpen)
}

// Test ModuleResult struct
func TestModuleResult_Struct(t *testing.T) {
	result := &ModuleResult{
		ModuleName:       "test-module",
		Success:          true,
		Data:             "test data",
		Error:            nil,
		DegradationLevel: DegradationLevelNone,
		Confidence:       1.0,
		ProcessingTime:   100 * time.Millisecond,
		FallbackUsed:     false,
		RetryAttempts:    0,
	}

	assert.Equal(t, "test-module", result.ModuleName)
	assert.True(t, result.Success)
	assert.Equal(t, "test data", result.Data)
	assert.Nil(t, result.Error)
	assert.Equal(t, DegradationLevelNone, result.DegradationLevel)
	assert.Equal(t, 1.0, result.Confidence)
	assert.False(t, result.FallbackUsed)
	assert.Equal(t, 0, result.RetryAttempts)
}

// Test helper functions
func TestContains_HelperFunction(t *testing.T) {
	assert.True(t, contains("timeout error", "timeout"))
	assert.True(t, contains("connection failed", "connection"))
	assert.True(t, contains("temporary failure", "temporary"))
	assert.False(t, contains("permanent failure", "temporary"))
	assert.False(t, contains("success", "failure"))
}
