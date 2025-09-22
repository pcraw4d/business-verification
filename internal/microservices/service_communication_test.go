package microservices

import (
	"context"
	"testing"
	"time"

	"kyb-platform/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestServiceClient_NewServiceClient(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)
	loadBalancer := NewServiceLoadBalancer(discovery, logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)
	metrics := NewServiceMetrics(logger)
	timeout := &MockServiceTimeout{}
	retry := &MockServiceRetry{}
	rateLimiter := &MockServiceRateLimiter{}

	client := NewServiceClient(
		discovery,
		loadBalancer,
		circuitBreaker,
		metrics,
		timeout,
		retry,
		rateLimiter,
		logger,
	)

	assert.NotNil(t, client)
	assert.NotNil(t, client.discovery)
	assert.NotNil(t, client.loadBalancer)
	assert.NotNil(t, client.circuitBreaker)
	assert.NotNil(t, client.metrics)
	assert.NotNil(t, client.timeout)
	assert.NotNil(t, client.retry)
	assert.NotNil(t, client.rateLimiter)
	assert.NotNil(t, client.logger)
}

func TestServiceClient_Call_Success(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)
	loadBalancer := NewServiceLoadBalancer(discovery, logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)
	metrics := NewServiceMetrics(logger)
	timeout := &MockServiceTimeout{}
	retry := &MockServiceRetry{}
	rateLimiter := &MockServiceRateLimiter{AllowFunc: func(serviceName string) bool { return true }}

	client := NewServiceClient(
		discovery,
		loadBalancer,
		circuitBreaker,
		metrics,
		timeout,
		retry,
		rateLimiter,
		logger,
	)

	// Mock circuit breaker to return success
	circuitBreaker.(*ServiceCircuitBreakerImpl).executeServiceCall = func(ctx context.Context, serviceName, method string, request interface{}) (interface{}, error) {
		return "success", nil
	}

	ctx := context.Background()
	result, err := client.Call(ctx, "test-service", "test-method", "test-request")

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
}

func TestServiceClient_Call_RateLimited(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)
	loadBalancer := NewServiceLoadBalancer(discovery, logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)
	metrics := NewServiceMetrics(logger)
	timeout := &MockServiceTimeout{}
	retry := &MockServiceRetry{}
	rateLimiter := &MockServiceRateLimiter{AllowFunc: func(serviceName string) bool { return false }}

	client := NewServiceClient(
		discovery,
		loadBalancer,
		circuitBreaker,
		metrics,
		timeout,
		retry,
		rateLimiter,
		logger,
	)

	ctx := context.Background()
	result, err := client.Call(ctx, "test-service", "test-method", "test-request")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "rate limit exceeded")
}

func TestServiceClient_CallAsync_Success(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)
	loadBalancer := NewServiceLoadBalancer(discovery, logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)
	metrics := NewServiceMetrics(logger)
	timeout := &MockServiceTimeout{}
	retry := &MockServiceRetry{}
	rateLimiter := &MockServiceRateLimiter{AllowFunc: func(serviceName string) bool { return true }}

	client := NewServiceClient(
		discovery,
		loadBalancer,
		circuitBreaker,
		metrics,
		timeout,
		retry,
		rateLimiter,
		logger,
	)

	// Mock circuit breaker to return success
	circuitBreaker.(*ServiceCircuitBreakerImpl).executeServiceCall = func(ctx context.Context, serviceName, method string, request interface{}) (interface{}, error) {
		return "async-success", nil
	}

	ctx := context.Background()
	responseChan, err := client.CallAsync(ctx, "test-service", "test-method", "test-request")

	require.NoError(t, err)
	assert.NotNil(t, responseChan)

	// Wait for response
	select {
	case response := <-responseChan:
		assert.NoError(t, response.Error)
		assert.Equal(t, "async-success", response.Data)
		assert.Greater(t, response.Latency, time.Duration(0))
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for async response")
	}
}

func TestServiceClient_CallAsync_RateLimited(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)
	loadBalancer := NewServiceLoadBalancer(discovery, logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)
	metrics := NewServiceMetrics(logger)
	timeout := &MockServiceTimeout{}
	retry := &MockServiceRetry{}
	rateLimiter := &MockServiceRateLimiter{AllowFunc: func(serviceName string) bool { return false }}

	client := NewServiceClient(
		discovery,
		loadBalancer,
		circuitBreaker,
		metrics,
		timeout,
		retry,
		rateLimiter,
		logger,
	)

	ctx := context.Background()
	responseChan, err := client.CallAsync(ctx, "test-service", "test-method", "test-request")

	require.NoError(t, err)
	assert.NotNil(t, responseChan)

	// Wait for response
	select {
	case response := <-responseChan:
		assert.Error(t, response.Error)
		assert.Nil(t, response.Data)
		assert.Contains(t, response.Error.Error(), "rate limit exceeded")
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for async response")
	}
}

func TestServiceClient_Health(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)
	loadBalancer := NewServiceLoadBalancer(discovery, logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)
	metrics := NewServiceMetrics(logger)
	timeout := &MockServiceTimeout{}
	retry := &MockServiceRetry{}
	rateLimiter := &MockServiceRateLimiter{}

	client := NewServiceClient(
		discovery,
		loadBalancer,
		circuitBreaker,
		metrics,
		timeout,
		retry,
		rateLimiter,
		logger,
	)

	ctx := context.Background()
	health, err := client.Health(ctx, "test-service")

	// Should not error even if service doesn't exist
	assert.NoError(t, err)
	assert.NotNil(t, health)
}

func TestServiceLoadBalancer_NewServiceLoadBalancer(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	loadBalancer := NewServiceLoadBalancer(discovery, logger)

	assert.NotNil(t, loadBalancer)
	assert.NotNil(t, loadBalancer.discovery)
	assert.NotNil(t, loadBalancer.logger)
}

func TestServiceLoadBalancer_Select_NoInstances(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)
	loadBalancer := NewServiceLoadBalancer(discovery, logger)

	instance, err := loadBalancer.Select("non-existent-service")

	assert.Error(t, err)
	assert.Empty(t, instance)
	assert.Contains(t, err.Error(), "no instances available")
}

func TestServiceLoadBalancer_Select_WithInstances(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)
	loadBalancer := NewServiceLoadBalancer(discovery, logger)

	// Register service instances
	instance1 := ServiceInstance{
		ID:          "instance-1",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Host:        "localhost",
		Port:        8080,
		Protocol:    "http",
		Health: ServiceHealth{
			Status:    "healthy",
			Message:   "Service is running",
			Timestamp: time.Now(),
		},
		LastSeen: time.Now(),
	}

	instance2 := ServiceInstance{
		ID:          "instance-2",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Host:        "localhost",
		Port:        8081,
		Protocol:    "http",
		Health: ServiceHealth{
			Status:    "healthy",
			Message:   "Service is running",
			Timestamp: time.Now(),
		},
		LastSeen: time.Now(),
	}

	discovery.RegisterInstance(instance1)
	discovery.RegisterInstance(instance2)

	// Select instance
	instance, err := loadBalancer.Select("test-service")

	assert.NoError(t, err)
	assert.NotEmpty(t, instance)
	assert.Contains(t, []string{"instance-1", "instance-2"}, instance.ID)
	assert.Equal(t, "test-service", instance.ServiceName)
}

func TestServiceLoadBalancer_Select_OnlyUnhealthyInstances(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)
	loadBalancer := NewServiceLoadBalancer(discovery, logger)

	// Register unhealthy service instance
	instance := ServiceInstance{
		ID:          "instance-1",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Host:        "localhost",
		Port:        8080,
		Protocol:    "http",
		Health: ServiceHealth{
			Status:    "unhealthy",
			Message:   "Service is down",
			Timestamp: time.Now(),
		},
		LastSeen: time.Now(),
	}

	discovery.RegisterInstance(instance)

	// Select instance
	instanceResult, err := loadBalancer.Select("test-service")

	assert.Error(t, err)
	assert.Empty(t, instanceResult)
	assert.Contains(t, err.Error(), "no healthy instances available")
}

func TestServiceLoadBalancer_UpdateHealth(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)
	loadBalancer := NewServiceLoadBalancer(discovery, logger)

	// Update health for non-existent instance
	health := ServiceHealth{
		Status:    "healthy",
		Message:   "Service is running",
		Timestamp: time.Now(),
	}

	err := loadBalancer.UpdateHealth("non-existent", health)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "instance not found")
}

func TestServiceLoadBalancer_GetInstances(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)
	loadBalancer := NewServiceLoadBalancer(discovery, logger)

	// Register service instances
	instance1 := ServiceInstance{
		ID:          "instance-1",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Host:        "localhost",
		Port:        8080,
		Protocol:    "http",
		LastSeen:    time.Now(),
	}

	instance2 := ServiceInstance{
		ID:          "instance-2",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Host:        "localhost",
		Port:        8081,
		Protocol:    "http",
		LastSeen:    time.Now(),
	}

	discovery.RegisterInstance(instance1)
	discovery.RegisterInstance(instance2)

	// Get instances
	instances, err := loadBalancer.GetInstances("test-service")

	assert.NoError(t, err)
	assert.Equal(t, 2, len(instances))

	instanceIDs := make(map[string]bool)
	for _, instance := range instances {
		instanceIDs[instance.ID] = true
	}
	assert.True(t, instanceIDs["instance-1"])
	assert.True(t, instanceIDs["instance-2"])
}

func TestServiceCircuitBreaker_NewServiceCircuitBreaker(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)

	circuitBreaker := NewServiceCircuitBreaker(logger)

	assert.NotNil(t, circuitBreaker)
	assert.NotNil(t, circuitBreaker.states)
	assert.NotNil(t, circuitBreaker.logger)
}

func TestServiceCircuitBreaker_Execute_Success(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	// Mock successful service call
	circuitBreaker.executeServiceCall = func(ctx context.Context, serviceName, method string, request interface{}) (interface{}, error) {
		return "success", nil
	}

	ctx := context.Background()
	result, err := circuitBreaker.Execute(ctx, "test-service", "test-method", "test-request")

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
}

func TestServiceCircuitBreaker_Execute_Failure(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	// Mock failed service call
	circuitBreaker.executeServiceCall = func(ctx context.Context, serviceName, method string, request interface{}) (interface{}, error) {
		return nil, assert.AnError
	}

	ctx := context.Background()
	result, err := circuitBreaker.Execute(ctx, "test-service", "test-method", "test-request")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, assert.AnError, err)
}

func TestServiceCircuitBreaker_GetState(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	// Get state for non-existent service
	state := circuitBreaker.GetState("non-existent-service")

	assert.Equal(t, "non-existent-service", state.ServiceName)
	assert.Equal(t, "closed", state.State)
	assert.Equal(t, int64(0), state.FailureCount)
	assert.Equal(t, int64(0), state.SuccessCount)
}

func TestServiceCircuitBreaker_Reset(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	// Reset non-existent service
	err := circuitBreaker.Reset("non-existent-service")
	assert.NoError(t, err)

	// Verify state is reset
	state := circuitBreaker.GetState("non-existent-service")
	assert.Equal(t, "closed", state.State)
	assert.Equal(t, int64(0), state.FailureCount)
	assert.Equal(t, int64(0), state.SuccessCount)
}

func TestServiceMetrics_NewServiceMetrics(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)

	metrics := NewServiceMetrics(logger)

	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.metrics)
	assert.NotNil(t, metrics.logger)
}

func TestServiceMetrics_RecordRequest(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	metrics := NewServiceMetrics(logger)

	// Record successful request
	metrics.RecordRequest("test-service", "test-method", 100*time.Millisecond, true)

	// Record failed request
	metrics.RecordRequest("test-service", "test-method", 200*time.Millisecond, false)

	// Get metrics
	metricsData := metrics.GetMetrics("test-service")

	assert.Equal(t, "test-service", metricsData.ServiceName)
	assert.Equal(t, int64(2), metricsData.RequestCount)
	assert.Equal(t, int64(1), metricsData.SuccessCount)
	assert.Equal(t, int64(1), metricsData.ErrorCount)
	assert.Equal(t, 0.5, metricsData.SuccessRate)
	assert.Equal(t, 0.5, metricsData.ErrorRate)
}

func TestServiceMetrics_RecordLatency(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	metrics := NewServiceMetrics(logger)

	// Record latency
	metrics.RecordLatency("test-service", "test-method", 100*time.Millisecond)
	metrics.RecordLatency("test-service", "test-method", 200*time.Millisecond)

	// Get metrics
	metricsData := metrics.GetMetrics("test-service")

	assert.Equal(t, "test-service", metricsData.ServiceName)
	assert.Equal(t, int64(2), metricsData.RequestCount)
	assert.Equal(t, 150*time.Millisecond, metricsData.AverageLatency)
}

func TestServiceMetrics_RecordError(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	metrics := NewServiceMetrics(logger)

	// Record errors
	metrics.RecordError("test-service", "test-method", "timeout")
	metrics.RecordError("test-service", "test-method", "connection_failed")

	// Get metrics
	metricsData := metrics.GetMetrics("test-service")

	assert.Equal(t, "test-service", metricsData.ServiceName)
	assert.Equal(t, int64(2), metricsData.ErrorCount)
}

func TestServiceMetrics_GetMetrics_NonExistent(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	metrics := NewServiceMetrics(logger)

	// Get metrics for non-existent service
	metricsData := metrics.GetMetrics("non-existent-service")

	assert.Equal(t, "non-existent-service", metricsData.ServiceName)
	assert.Equal(t, int64(0), metricsData.RequestCount)
	assert.Equal(t, int64(0), metricsData.SuccessCount)
	assert.Equal(t, int64(0), metricsData.ErrorCount)
	assert.Equal(t, 0.0, metricsData.SuccessRate)
	assert.Equal(t, 0.0, metricsData.ErrorRate)
}

// Mock implementations for testing

type MockServiceTimeout struct{}

func (m *MockServiceTimeout) GetTimeout(serviceName, method string) time.Duration {
	return 30 * time.Second
}

func (m *MockServiceTimeout) SetTimeout(serviceName, method string, timeout time.Duration) error {
	return nil
}

type MockServiceRetry struct{}

func (m *MockServiceRetry) ShouldRetry(serviceName, method string, attempt int, err error) bool {
	return attempt < 3
}

func (m *MockServiceRetry) GetBackoff(serviceName, method string, attempt int) time.Duration {
	return time.Duration(attempt) * time.Second
}

type MockServiceRateLimiter struct {
	AllowFunc func(serviceName string) bool
}

func (m *MockServiceRateLimiter) Allow(serviceName string) bool {
	if m.AllowFunc != nil {
		return m.AllowFunc(serviceName)
	}
	return true
}

func (m *MockServiceRateLimiter) SetRateLimit(serviceName string, requestsPerSecond int) error {
	return nil
}

// Test concurrent operations
func TestServiceCommunication_ConcurrentOperations(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)
	loadBalancer := NewServiceLoadBalancer(discovery, logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)
	metrics := NewServiceMetrics(logger)
	timeout := &MockServiceTimeout{}
	retry := &MockServiceRetry{}
	rateLimiter := &MockServiceRateLimiter{AllowFunc: func(serviceName string) bool { return true }}

	client := NewServiceClient(
		discovery,
		loadBalancer,
		circuitBreaker,
		metrics,
		timeout,
		retry,
		rateLimiter,
		logger,
	)

	// Mock successful service call
	circuitBreaker.(*ServiceCircuitBreakerImpl).executeServiceCall = func(ctx context.Context, serviceName, method string, request interface{}) (interface{}, error) {
		return "success", nil
	}

	// Start multiple goroutines to test concurrent operations
	const numGoroutines = 10
	const numOperations = 100

	// Channel to signal completion
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				ctx := context.Background()
				result, err := client.Call(ctx, "test-service", "test-method", "test-request")

				if err != nil {
					t.Logf("Failed to call service: %v", err)
					continue
				}

				if result != "success" {
					t.Logf("Unexpected result: %v", result)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify metrics were recorded
	metricsData := metrics.GetMetrics("test-service")
	assert.Greater(t, metricsData.RequestCount, int64(0))
	assert.Greater(t, metricsData.SuccessCount, int64(0))
}

// Test circuit breaker state transitions
func TestServiceCircuitBreaker_StateTransitions(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	logger := observability.NewLogger(zapLogger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	// Configure circuit breaker for faster testing
	circuitBreaker.states["test-service"] = &CircuitBreakerState{
		ServiceName:     "test-service",
		State:           "closed",
		FailureCount:    0,
		SuccessCount:    0,
		Threshold:       3,               // Low threshold for testing
		Timeout:         1 * time.Second, // Short timeout for testing
		LastStateChange: time.Now(),
	}

	// Mock service call that fails
	circuitBreaker.executeServiceCall = func(ctx context.Context, serviceName, method string, request interface{}) (interface{}, error) {
		return nil, assert.AnError
	}

	ctx := context.Background()

	// Make calls until circuit breaker opens
	for i := 0; i < 5; i++ {
		_, err := circuitBreaker.Execute(ctx, "test-service", "test-method", "test-request")
		assert.Error(t, err)
	}

	// Verify circuit breaker is open
	state := circuitBreaker.GetState("test-service")
	assert.Equal(t, "open", state.State)
	assert.Equal(t, int64(3), state.FailureCount)

	// Wait for timeout
	time.Sleep(2 * time.Second)

	// Mock successful service call for half-open state
	circuitBreaker.executeServiceCall = func(ctx context.Context, serviceName, method string, request interface{}) (interface{}, error) {
		return "success", nil
	}

	// Make a call to test half-open state
	result, err := circuitBreaker.Execute(ctx, "test-service", "test-method", "test-request")

	assert.NoError(t, err)
	assert.Equal(t, "success", result)

	// Verify circuit breaker is closed again
	state = circuitBreaker.GetState("test-service")
	assert.Equal(t, "closed", state.State)
	assert.Equal(t, int64(1), state.SuccessCount)
}
