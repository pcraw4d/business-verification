package engine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

func createTestRiskEngine() *RiskEngine {
	logger := zap.NewNop()
	config := &Config{
		MaxConcurrentRequests: 100,
		RequestTimeout:        30 * time.Second,
		CacheTTL:              5 * time.Minute,
		CircuitBreakerConfig: CircuitBreakerConfig{
			FailureThreshold: 5,
			RecoveryTimeout:  30 * time.Second,
			HalfOpenMaxCalls: 3,
		},
		EnableMetrics: true,
		EnableCaching: true,
	}

	// For testing, we'll use nil for services that aren't needed
	engine := NewRiskEngine(nil, logger, config, nil)

	return engine
}

func TestRiskEngine_AssessRisk_NilService(t *testing.T) {
	engine := createTestRiskEngine()

	req := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	// Since we're using nil services, we expect a panic or error
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to nil service
			t.Logf("Expected panic: %v", r)
		}
	}()

	result, err := engine.AssessRisk(context.Background(), req)

	// If we get here, it means no panic occurred, so we should have an error
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestRiskEngine_GetCacheStats(t *testing.T) {
	engine := createTestRiskEngine()

	stats := engine.GetCacheStats()

	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.Hits, int64(0))
	assert.GreaterOrEqual(t, stats.Misses, int64(0))
	assert.GreaterOrEqual(t, stats.Size, 0)
}

func TestRiskEngine_GetCircuitBreakerState(t *testing.T) {
	engine := createTestRiskEngine()

	state := engine.GetCircuitBreakerState()

	assert.NotNil(t, state)
	assert.Contains(t, []CircuitBreakerState{StateClosed, StateOpen, StateHalfOpen}, state)
}

func TestRiskEngine_Shutdown_Success(t *testing.T) {
	engine := createTestRiskEngine()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := engine.Shutdown(ctx)

	assert.NoError(t, err)
}
