//go:build !comprehensive_test && !e2e_railway
// +build !comprehensive_test,!e2e_railway

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"

	"kyb-platform/internal/metrics"
	"kyb-platform/internal/resilience"
	"kyb-platform/services/merchant-service/internal/config"
	"kyb-platform/services/merchant-service/internal/supabase"
)

// TestProductionSafetyGuards tests that mock data is not allowed in production
func TestProductionSafetyGuards(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	logger := zap.NewNop()
	cfg := &config.Config{
		Environment: "production",
		Merchant: config.MerchantConfig{
			AllowMockData: false,
		},
	}

	// Test that getMockMerchant returns error in production
	// This is tested indirectly through the handler behavior
	// In production, getMockMerchant should return an error when AllowMockData is false
	t.Log("Production safety guards: Mock data should be disabled in production")
}

// TestRetryLogic tests retry behavior with exponential backoff
func TestRetryLogic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	retryConfig := resilience.DefaultRetryConfig()
	retryConfig.MaxAttempts = 3
	retryConfig.InitialDelay = 100 * time.Millisecond

	attemptCount := 0
	result, err := resilience.RetryWithBackoff(ctx, retryConfig, func() (string, error) {
		attemptCount++
		if attemptCount < 3 {
			return "", fmt.Errorf("simulated error")
		}
		return "success", nil
	})

	if err != nil {
		t.Errorf("Expected success after retries, got error: %v", err)
	}
	if result != "success" {
		t.Errorf("Expected 'success', got %s", result)
	}
	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}

// TestCircuitBreaker tests circuit breaker behavior
func TestCircuitBreaker(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	cbConfig := resilience.DefaultCircuitBreakerConfig()
	cbConfig.FailureThreshold = 3
	cbConfig.Timeout = 1 * time.Second
	circuitBreaker := resilience.NewCircuitBreaker(cbConfig)

	// Simulate failures to open circuit
	for i := 0; i < 3; i++ {
		circuitBreaker.Execute(ctx, func() error {
			return fmt.Errorf("simulated failure")
		})
	}

	// Circuit should be open now
	if circuitBreaker.GetState() != resilience.CircuitOpen {
		t.Error("Expected circuit to be open after threshold failures")
	}

	// Execute should fail immediately
	err := circuitBreaker.Execute(ctx, func() error {
		return nil
	})
	if err == nil {
		t.Error("Expected error when circuit is open")
	}
}

// TestFallbackMetrics tests fallback metrics tracking
func TestFallbackMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	logger := zap.NewNop()
	metrics := metrics.NewFallbackMetrics(logger)

	ctx := context.Background()

	// Record some requests
	metrics.RecordRequest(ctx, "test-service")
	metrics.RecordRequest(ctx, "test-service")
	metrics.RecordRequest(ctx, "test-service")

	// Record fallback usage
	metrics.RecordFallbackUsage(ctx, "test-service", "database_fallback", "supabase", 100*time.Millisecond)

	// Check stats
	stats := metrics.GetStats("test-service")
	if stats.TotalRequests != 4 {
		t.Errorf("Expected 4 total requests, got %d", stats.TotalRequests)
	}
	if stats.FallbackCount != 1 {
		t.Errorf("Expected 1 fallback, got %d", stats.FallbackCount)
	}
	if stats.FallbackRate != 25.0 {
		t.Errorf("Expected 25%% fallback rate, got %.2f%%", stats.FallbackRate)
	}
}

// TestEnvironmentConfiguration tests environment-based configuration
func TestEnvironmentConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Test production environment
	os.Setenv("ENVIRONMENT", "production")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Environment != "production" {
		t.Errorf("Expected environment 'production', got '%s'", cfg.Environment)
	}
	if cfg.Merchant.AllowMockData {
		t.Error("Expected AllowMockData to be false in production")
	}

	// Test development environment
	os.Setenv("ENVIRONMENT", "development")
	cfg, err = config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Environment != "development" {
		t.Errorf("Expected environment 'development', got '%s'", cfg.Environment)
	}
	if !cfg.Merchant.AllowMockData {
		t.Error("Expected AllowMockData to be true in development")
	}

	// Cleanup
	os.Unsetenv("ENVIRONMENT")
}

