package microservices

import (
	"context"
	"fmt"
	"testing"
	"time"

	"kyb-platform/internal/config"
	"kyb-platform/internal/observability"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceIsolationManager_NewServiceIsolationManager(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.services)
	assert.Equal(t, 0, len(manager.services))
	assert.NotNil(t, manager.logger)
	assert.NotNil(t, manager.metrics)
	assert.NotNil(t, manager.circuitBreaker)
}

func TestServiceIsolationManager_RegisterService(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Create mock service
	service := &MockService{
		name:    "test-service",
		version: "1.0.0",
		health: ServiceHealth{
			Status:    "healthy",
			Message:   "Service is running",
			Timestamp: time.Now(),
		},
	}

	// Configure fallback
	fallbackConfig := FallbackConfig{
		Enabled:    true,
		Strategy:   FallbackStrategyStatic,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
		Timeout:    30 * time.Second,
		FallbackData: map[string]interface{}{
			"default_response": "fallback data",
		},
	}

	// Register service
	err := manager.RegisterService(service, IsolationLevelEnhanced, fallbackConfig)
	require.NoError(t, err)

	// Verify service is registered
	services := manager.ListIsolatedServices()
	assert.Equal(t, 1, len(services))
	assert.Contains(t, services, "test-service")

	// Get service info
	info, err := manager.GetServiceIsolationInfo("test-service")
	require.NoError(t, err)
	assert.Equal(t, "test-service", info["name"])
	assert.Equal(t, "enhanced", info["isolation_level"])
	assert.Equal(t, true, info["fallback_enabled"])
}

func TestServiceIsolationManager_RegisterService_Duplicate(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Create mock service
	service := &MockService{
		name:    "test-service",
		version: "1.0.0",
	}

	fallbackConfig := FallbackConfig{
		Enabled:  true,
		Strategy: FallbackStrategyStatic,
	}

	// Register service twice
	err := manager.RegisterService(service, IsolationLevelBasic, fallbackConfig)
	require.NoError(t, err)

	err = manager.RegisterService(service, IsolationLevelBasic, fallbackConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestServiceIsolationManager_UnregisterService(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Create and register service
	service := &MockService{
		name:    "test-service",
		version: "1.0.0",
	}

	fallbackConfig := FallbackConfig{
		Enabled:  true,
		Strategy: FallbackStrategyStatic,
	}

	err := manager.RegisterService(service, IsolationLevelBasic, fallbackConfig)
	require.NoError(t, err)

	// Verify service is registered
	services := manager.ListIsolatedServices()
	assert.Equal(t, 1, len(services))

	// Unregister service
	err = manager.UnregisterService("test-service")
	require.NoError(t, err)

	// Verify service is unregistered
	services = manager.ListIsolatedServices()
	assert.Equal(t, 0, len(services))
}

func TestServiceIsolationManager_UnregisterService_NotFound(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Try to unregister non-existent service
	err := manager.UnregisterService("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestServiceIsolationManager_ExecuteWithIsolation_None(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Create and register service
	service := &MockService{
		name:    "test-service",
		version: "1.0.0",
	}

	fallbackConfig := FallbackConfig{
		Enabled:  true,
		Strategy: FallbackStrategyStatic,
	}

	err := manager.RegisterService(service, IsolationLevelNone, fallbackConfig)
	require.NoError(t, err)

	// Execute with no isolation
	ctx := context.Background()
	result, err := manager.ExecuteWithIsolation(ctx, "test-service", "test-method", "test-request")

	// Should fail because no actual service execution is implemented
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestServiceIsolationManager_ExecuteWithIsolation_Basic(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Create and register service
	service := &MockService{
		name:    "test-service",
		version: "1.0.0",
	}

	fallbackConfig := FallbackConfig{
		Enabled:  true,
		Strategy: FallbackStrategyStatic,
		FallbackData: map[string]interface{}{
			"default_response": "fallback data",
		},
	}

	err := manager.RegisterService(service, IsolationLevelBasic, fallbackConfig)
	require.NoError(t, err)

	// Execute with basic isolation
	ctx := context.Background()
	result, err := manager.ExecuteWithIsolation(ctx, "test-service", "test-method", "test-request")

	// Should fail because no actual service execution is implemented
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestServiceIsolationManager_ExecuteWithIsolation_Enhanced(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Create and register service
	service := &MockService{
		name:    "test-service",
		version: "1.0.0",
	}

	fallbackConfig := FallbackConfig{
		Enabled:  true,
		Strategy: FallbackStrategyStatic,
		FallbackData: map[string]interface{}{
			"default_response": "fallback data",
		},
	}

	err := manager.RegisterService(service, IsolationLevelEnhanced, fallbackConfig)
	require.NoError(t, err)

	// Execute with enhanced isolation
	ctx := context.Background()
	result, err := manager.ExecuteWithIsolation(ctx, "test-service", "test-method", "test-request")

	// Should fail because no actual service execution is implemented
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestServiceIsolationManager_ExecuteWithIsolation_Full(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Create and register service
	service := &MockService{
		name:    "test-service",
		version: "1.0.0",
	}

	fallbackConfig := FallbackConfig{
		Enabled:  true,
		Strategy: FallbackStrategyStatic,
		FallbackData: map[string]interface{}{
			"default_response": "fallback data",
		},
	}

	err := manager.RegisterService(service, IsolationLevelFull, fallbackConfig)
	require.NoError(t, err)

	// Execute with full isolation
	ctx := context.Background()
	result, err := manager.ExecuteWithIsolation(ctx, "test-service", "test-method", "test-request")

	// Should fail because no actual service execution is implemented
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestServiceIsolationManager_ExecuteWithIsolation_ServiceNotFound(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Try to execute with non-existent service
	ctx := context.Background()
	result, err := manager.ExecuteWithIsolation(ctx, "non-existent", "test-method", "test-request")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestServiceIsolationManager_UpdateServiceHealth(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Create and register service
	service := &MockService{
		name:    "test-service",
		version: "1.0.0",
	}

	fallbackConfig := FallbackConfig{
		Enabled:  true,
		Strategy: FallbackStrategyStatic,
	}

	err := manager.RegisterService(service, IsolationLevelBasic, fallbackConfig)
	require.NoError(t, err)

	// Update service health
	newHealth := ServiceHealth{
		Status:    "degraded",
		Message:   "High latency detected",
		Timestamp: time.Now(),
	}

	err = manager.UpdateServiceHealth("test-service", newHealth)
	require.NoError(t, err)

	// Verify health is updated
	info, err := manager.GetServiceIsolationInfo("test-service")
	require.NoError(t, err)
	assert.Equal(t, "degraded", info["health_status"])
}

func TestServiceIsolationManager_UpdateServiceHealth_NotFound(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Try to update health for non-existent service
	health := ServiceHealth{
		Status:    "healthy",
		Message:   "Service is running",
		Timestamp: time.Now(),
	}

	err := manager.UpdateServiceHealth("non-existent", health)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestServiceIsolationManager_GetServiceIsolationInfo(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Create and register service
	service := &MockService{
		name:    "test-service",
		version: "1.0.0",
		health: ServiceHealth{
			Status:    "healthy",
			Message:   "Service is running",
			Timestamp: time.Now(),
		},
	}

	fallbackConfig := FallbackConfig{
		Enabled:    true,
		Strategy:   FallbackStrategyStatic,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
		Timeout:    30 * time.Second,
		FallbackData: map[string]interface{}{
			"default_response": "fallback data",
		},
	}

	err := manager.RegisterService(service, IsolationLevelEnhanced, fallbackConfig)
	require.NoError(t, err)

	// Get service isolation info
	info, err := manager.GetServiceIsolationInfo("test-service")
	require.NoError(t, err)

	// Verify info
	assert.Equal(t, "test-service", info["name"])
	assert.Equal(t, "enhanced", info["isolation_level"])
	assert.Equal(t, true, info["fallback_enabled"])
	assert.Equal(t, "static", info["fallback_strategy"])
	assert.Equal(t, 3, info["max_retries"])
	assert.Equal(t, "healthy", info["health_status"])
	assert.NotNil(t, info["last_updated"])
}

func TestServiceIsolationManager_GetServiceIsolationInfo_NotFound(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Try to get info for non-existent service
	info, err := manager.GetServiceIsolationInfo("non-existent")

	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "not found")
}

func TestServiceIsolationManager_ListIsolatedServices(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Register multiple services
	service1 := &MockService{name: "service-1", version: "1.0.0"}
	service2 := &MockService{name: "service-2", version: "2.0.0"}

	fallbackConfig := FallbackConfig{
		Enabled:  true,
		Strategy: FallbackStrategyStatic,
	}

	manager.RegisterService(service1, IsolationLevelBasic, fallbackConfig)
	manager.RegisterService(service2, IsolationLevelEnhanced, fallbackConfig)

	// List isolated services
	services := manager.ListIsolatedServices()

	assert.Equal(t, 2, len(services))
	assert.Contains(t, services, "service-1")
	assert.Contains(t, services, "service-2")
}

func TestServiceIsolationManager_GetIsolationStats(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Register services with different isolation levels
	service1 := &MockService{name: "service-1", version: "1.0.0"}
	service2 := &MockService{name: "service-2", version: "2.0.0"}
	service3 := &MockService{name: "service-3", version: "3.0.0"}

	fallbackConfig := FallbackConfig{
		Enabled:  true,
		Strategy: FallbackStrategyStatic,
	}

	manager.RegisterService(service1, IsolationLevelNone, fallbackConfig)
	manager.RegisterService(service2, IsolationLevelBasic, fallbackConfig)
	manager.RegisterService(service3, IsolationLevelEnhanced, fallbackConfig)

	// Get isolation stats
	stats := manager.GetIsolationStats()

	// Verify stats
	assert.Equal(t, 3, stats["total_services"])
	assert.Equal(t, 1, stats["none_isolation"])
	assert.Equal(t, 1, stats["basic_isolation"])
	assert.Equal(t, 1, stats["enhanced_isolation"])
	assert.Equal(t, 0, stats["full_isolation"])
	assert.Equal(t, 3, stats["fallback_enabled"])
}

func TestServiceIsolationManager_SetIsolationLevel(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Create and register service
	service := &MockService{
		name:    "test-service",
		version: "1.0.0",
	}

	fallbackConfig := FallbackConfig{
		Enabled:  true,
		Strategy: FallbackStrategyStatic,
	}

	err := manager.RegisterService(service, IsolationLevelBasic, fallbackConfig)
	require.NoError(t, err)

	// Set isolation level
	err = manager.SetIsolationLevel("test-service", IsolationLevelFull)
	require.NoError(t, err)

	// Verify isolation level is updated
	info, err := manager.GetServiceIsolationInfo("test-service")
	require.NoError(t, err)
	assert.Equal(t, "full", info["isolation_level"])
}

func TestServiceIsolationManager_SetIsolationLevel_NotFound(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Try to set isolation level for non-existent service
	err := manager.SetIsolationLevel("non-existent", IsolationLevelFull)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestServiceIsolationManager_UpdateFallbackConfig(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Create and register service
	service := &MockService{
		name:    "test-service",
		version: "1.0.0",
	}

	fallbackConfig := FallbackConfig{
		Enabled:  true,
		Strategy: FallbackStrategyStatic,
	}

	err := manager.RegisterService(service, IsolationLevelBasic, fallbackConfig)
	require.NoError(t, err)

	// Update fallback config
	newFallbackConfig := FallbackConfig{
		Enabled:    true,
		Strategy:   FallbackStrategyCached,
		MaxRetries: 5,
		RetryDelay: 2 * time.Second,
		Timeout:    60 * time.Second,
		FallbackData: map[string]interface{}{
			"new_fallback": "updated data",
		},
	}

	err = manager.UpdateFallbackConfig("test-service", newFallbackConfig)
	require.NoError(t, err)

	// Verify fallback config is updated
	info, err := manager.GetServiceIsolationInfo("test-service")
	require.NoError(t, err)
	assert.Equal(t, "cached", info["fallback_strategy"])
	assert.Equal(t, 5, info["max_retries"])
}

func TestServiceIsolationManager_UpdateFallbackConfig_NotFound(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Try to update fallback config for non-existent service
	fallbackConfig := FallbackConfig{
		Enabled:  true,
		Strategy: FallbackStrategyStatic,
	}

	err := manager.UpdateFallbackConfig("non-existent", fallbackConfig)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// Test fallback strategies
func TestServiceIsolationManager_ExecuteFallback_Static(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Execute fallback with static strategy
	result, err := manager.executeFallback("test-service", "test-method", "test-request", FallbackStrategyStatic)

	// Should return fallback data
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "fallback")
}

func TestServiceIsolationManager_ExecuteFallback_Cached(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Execute fallback with cached strategy
	result, err := manager.executeFallback("test-service", "test-method", "test-request", FallbackStrategyCached)

	// Should return cached data (simulated)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "cached")
}

func TestServiceIsolationManager_ExecuteFallback_Alternative(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Execute fallback with alternative service strategy
	result, err := manager.executeFallback("test-service", "test-method", "test-request", FallbackStrategyAlternative)

	// Should return alternative service data (simulated)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "alternative")
}

func TestServiceIsolationManager_ExecuteFallback_Degraded(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Execute fallback with degraded strategy
	result, err := manager.executeFallback("test-service", "test-method", "test-request", FallbackStrategyDegraded)

	// Should return degraded response
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "degraded")
}

// Test concurrent operations
func TestServiceIsolationManager_ConcurrentOperations(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	metrics := NewServiceMetrics(logger)
	circuitBreaker := NewServiceCircuitBreaker(logger)

	manager := NewServiceIsolationManager(logger, metrics, circuitBreaker)

	// Start multiple goroutines to test concurrent operations
	const numGoroutines = 10
	const numOperations = 100

	// Channel to signal completion
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				serviceName := fmt.Sprintf("service-%d", id)

				// Register service
				service := &MockService{
					name:    serviceName,
					version: "1.0.0",
				}

				fallbackConfig := FallbackConfig{
					Enabled:  true,
					Strategy: FallbackStrategyStatic,
				}

				err := manager.RegisterService(service, IsolationLevelBasic, fallbackConfig)
				if err != nil {
					t.Logf("Failed to register service: %v", err)
					continue
				}

				// Update health
				health := ServiceHealth{
					Status:    "healthy",
					Message:   "Service is running",
					Timestamp: time.Now(),
				}
				manager.UpdateServiceHealth(serviceName, health)

				// Get service info
				manager.GetServiceIsolationInfo(serviceName)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify no panics occurred and some services were registered
	services := manager.ListIsolatedServices()
	assert.Greater(t, len(services), 0)
}

// Test isolation level constants
func TestIsolationLevel_Constants(t *testing.T) {
	assert.Equal(t, IsolationLevel("none"), IsolationLevelNone)
	assert.Equal(t, IsolationLevel("basic"), IsolationLevelBasic)
	assert.Equal(t, IsolationLevel("enhanced"), IsolationLevelEnhanced)
	assert.Equal(t, IsolationLevel("full"), IsolationLevelFull)
}

// Test fallback strategy constants
func TestFallbackStrategy_Constants(t *testing.T) {
	assert.Equal(t, FallbackStrategy("static"), FallbackStrategyStatic)
	assert.Equal(t, FallbackStrategy("cached"), FallbackStrategyCached)
	assert.Equal(t, FallbackStrategy("alternative"), FallbackStrategyAlternative)
	assert.Equal(t, FallbackStrategy("degraded"), FallbackStrategyDegraded)
}

// Test IsolatedService struct
func TestIsolatedService_Struct(t *testing.T) {
	service := &IsolatedService{
		Name:      "test-service",
		Instances: make([]ServiceInstance, 0),
		Health: ServiceHealth{
			Status:    "healthy",
			Message:   "Service is running",
			Timestamp: time.Now(),
		},
		IsolationLevel: IsolationLevelBasic,
		FallbackConfig: FallbackConfig{
			Enabled:  true,
			Strategy: FallbackStrategyStatic,
		},
		LastUpdated: time.Now(),
	}

	assert.Equal(t, "test-service", service.Name)
	assert.Equal(t, IsolationLevelBasic, service.IsolationLevel)
	assert.True(t, service.FallbackConfig.Enabled)
	assert.Equal(t, FallbackStrategyStatic, service.FallbackConfig.Strategy)
}

// Test FallbackConfig struct
func TestFallbackConfig_Struct(t *testing.T) {
	config := FallbackConfig{
		Enabled:        true,
		Strategy:       FallbackStrategyCached,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		Timeout:        30 * time.Second,
		CircuitBreaker: true,
		FallbackData: map[string]interface{}{
			"key": "value",
		},
	}

	assert.True(t, config.Enabled)
	assert.Equal(t, FallbackStrategyCached, config.Strategy)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.RetryDelay)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.True(t, config.CircuitBreaker)
	assert.Equal(t, "value", config.FallbackData["key"])
}
