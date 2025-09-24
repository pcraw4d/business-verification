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

func TestServiceRegistry_NewServiceRegistry(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)

	registry := NewServiceRegistry(logger)

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.services)
	assert.Equal(t, 0, len(registry.services))
}

func TestServiceRegistry_Register(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)

	// Create a mock service
	mockService := &MockService{
		name:    "test-service",
		version: "1.0.0",
	}

	// Register the service
	err := registry.Register(mockService)
	require.NoError(t, err)

	// Verify service is registered
	service, err := registry.GetService("test-service")
	require.NoError(t, err)
	assert.Equal(t, "test-service", service.ServiceName())
	assert.Equal(t, "1.0.0", service.Version())
}

func TestServiceRegistry_GetService_NotFound(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)

	// Try to get non-existent service
	service, err := registry.GetService("non-existent")
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "service not found")
}

func TestServiceRegistry_ListServices(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)

	// Register multiple services
	service1 := &MockService{name: "service-1", version: "1.0.0"}
	service2 := &MockService{name: "service-2", version: "2.0.0"}

	registry.Register(service1)
	registry.Register(service2)

	// List all services
	services := registry.ListServices()
	assert.Equal(t, 2, len(services))
	assert.Contains(t, services, "service-1")
	assert.Contains(t, services, "service-2")
}

func TestServiceRegistry_GetHealthyServices(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)

	// Register services
	healthyService := &MockService{name: "healthy", version: "1.0.0", health: ServiceHealth{Status: "healthy"}}
	unhealthyService := &MockService{name: "unhealthy", version: "1.0.0", health: ServiceHealth{Status: "unhealthy"}}

	registry.Register(healthyService)
	registry.Register(unhealthyService)

	// Get healthy services
	healthyServices := registry.GetHealthyServices()
	assert.Equal(t, 1, len(healthyServices))
	assert.Equal(t, "healthy", healthyServices[0].ServiceName())
}

func TestServiceDiscovery_NewServiceDiscovery(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)

	discovery := NewServiceDiscovery(logger, registry)

	assert.NotNil(t, discovery)
	assert.NotNil(t, discovery.instances)
	assert.Equal(t, 0, len(discovery.instances))
	assert.NotNil(t, discovery.watchers)
	assert.Equal(t, 0, len(discovery.watchers))
}

func TestServiceDiscovery_RegisterInstance(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Create service instance
	instance := ServiceInstance{
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

	// Register instance
	err := discovery.RegisterInstance(instance)
	require.NoError(t, err)

	// Verify instance is registered
	instances, err := discovery.Discover("test-service")
	require.NoError(t, err)
	assert.Equal(t, 1, len(instances))
	assert.Equal(t, "instance-1", instances[0].ID)
	assert.Equal(t, "test-service", instances[0].ServiceName)
}

func TestServiceDiscovery_RegisterInstance_Duplicate(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Create service instance
	instance := ServiceInstance{
		ID:          "instance-1",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Host:        "localhost",
		Port:        8080,
		Protocol:    "http",
		LastSeen:    time.Now(),
	}

	// Register instance twice
	err := discovery.RegisterInstance(instance)
	require.NoError(t, err)

	err = discovery.RegisterInstance(instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "instance already exists")
}

func TestServiceDiscovery_UnregisterInstance(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Register instance
	instance := ServiceInstance{
		ID:          "instance-1",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Host:        "localhost",
		Port:        8080,
		Protocol:    "http",
		LastSeen:    time.Now(),
	}

	err := discovery.RegisterInstance(instance)
	require.NoError(t, err)

	// Unregister instance
	err = discovery.UnregisterInstance("test-service", "instance-1")
	require.NoError(t, err)

	// Verify instance is removed
	instances, err := discovery.Discover("test-service")
	require.NoError(t, err)
	assert.Equal(t, 0, len(instances))
}

func TestServiceDiscovery_UnregisterInstance_NotFound(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Try to unregister non-existent instance
	err := discovery.UnregisterInstance("test-service", "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "instance not found")
}

func TestServiceDiscovery_Discover(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Register multiple instances
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

	// Discover instances
	instances, err := discovery.Discover("test-service")
	require.NoError(t, err)
	assert.Equal(t, 2, len(instances))

	// Verify instance IDs
	instanceIDs := make(map[string]bool)
	for _, instance := range instances {
		instanceIDs[instance.ID] = true
	}
	assert.True(t, instanceIDs["instance-1"])
	assert.True(t, instanceIDs["instance-2"])
}

func TestServiceDiscovery_Discover_ServiceNotFound(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Try to discover non-existent service
	instances, err := discovery.Discover("non-existent")
	require.NoError(t, err)
	assert.Equal(t, 0, len(instances))
}

func TestServiceDiscovery_DiscoverAll(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Register instances for multiple services
	instance1 := ServiceInstance{
		ID:          "instance-1",
		ServiceName: "service-1",
		Version:     "1.0.0",
		Host:        "localhost",
		Port:        8080,
		Protocol:    "http",
		LastSeen:    time.Now(),
	}

	instance2 := ServiceInstance{
		ID:          "instance-2",
		ServiceName: "service-2",
		Version:     "2.0.0",
		Host:        "localhost",
		Port:        8081,
		Protocol:    "http",
		LastSeen:    time.Now(),
	}

	discovery.RegisterInstance(instance1)
	discovery.RegisterInstance(instance2)

	// Discover all services
	allServices := discovery.DiscoverAll()
	assert.Equal(t, 2, len(allServices))
	assert.Equal(t, 1, len(allServices["service-1"]))
	assert.Equal(t, 1, len(allServices["service-2"]))
}

func TestServiceDiscovery_Watch(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Watch a service
	eventChan, err := discovery.Watch("test-service")
	require.NoError(t, err)
	assert.NotNil(t, eventChan)

	// Verify watcher is registered
	discovery.mu.RLock()
	watchers, exists := discovery.watchers["test-service"]
	discovery.mu.RUnlock()
	assert.True(t, exists)
	assert.Equal(t, 1, len(watchers))
}

func TestServiceDiscovery_Unwatch(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Watch a service
	eventChan, err := discovery.Watch("test-service")
	require.NoError(t, err)

	// Unwatch the service
	err = discovery.Unwatch("test-service")
	require.NoError(t, err)

	// Verify watcher is removed
	discovery.mu.RLock()
	watchers, exists := discovery.watchers["test-service"]
	discovery.mu.RUnlock()
	assert.False(t, exists || len(watchers) == 0)

	// Verify channel is closed
	select {
	case _, ok := <-eventChan:
		assert.False(t, ok) // Channel should be closed
	default:
		// Channel might not be closed immediately, which is okay
	}
}

func TestServiceDiscovery_UpdateInstanceHealth(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Register instance
	instance := ServiceInstance{
		ID:          "instance-1",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Host:        "localhost",
		Port:        8080,
		Protocol:    "http",
		LastSeen:    time.Now(),
	}

	err := discovery.RegisterInstance(instance)
	require.NoError(t, err)

	// Update health
	newHealth := ServiceHealth{
		Status:    "degraded",
		Message:   "High latency detected",
		Timestamp: time.Now(),
	}

	err = discovery.UpdateInstanceHealth("test-service", "instance-1", newHealth)
	require.NoError(t, err)

	// Verify health is updated
	instances, err := discovery.Discover("test-service")
	require.NoError(t, err)
	assert.Equal(t, 1, len(instances))
	assert.Equal(t, "degraded", instances[0].Health.Status)
	assert.Equal(t, "High latency detected", instances[0].Health.Message)
}

func TestServiceDiscovery_UpdateInstanceHealth_NotFound(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Try to update health for non-existent instance
	health := ServiceHealth{
		Status:    "healthy",
		Message:   "Service is running",
		Timestamp: time.Now(),
	}

	err := discovery.UpdateInstanceHealth("test-service", "non-existent", health)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "instance not found")
}

func TestServiceDiscovery_CleanupStaleInstances(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Register instances with different timestamps
	recentInstance := ServiceInstance{
		ID:          "recent",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Host:        "localhost",
		Port:        8080,
		Protocol:    "http",
		LastSeen:    time.Now(),
	}

	staleInstance := ServiceInstance{
		ID:          "stale",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Host:        "localhost",
		Port:        8081,
		Protocol:    "http",
		LastSeen:    time.Now().Add(-2 * time.Hour), // 2 hours ago
	}

	discovery.RegisterInstance(recentInstance)
	discovery.RegisterInstance(staleInstance)

	// Cleanup stale instances (older than 1 hour)
	discovery.CleanupStaleInstances(1 * time.Hour)

	// Verify only recent instance remains
	instances, err := discovery.Discover("test-service")
	require.NoError(t, err)
	assert.Equal(t, 1, len(instances))
	assert.Equal(t, "recent", instances[0].ID)
}

func TestServiceDiscovery_StartHealthCheck(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Register instance
	instance := ServiceInstance{
		ID:          "instance-1",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Host:        "localhost",
		Port:        8080,
		Protocol:    "http",
		LastSeen:    time.Now(),
	}

	err := discovery.RegisterInstance(instance)
	require.NoError(t, err)

	// Start health check
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	discovery.StartHealthCheck(ctx, 100*time.Millisecond) // Short interval for testing

	// Wait a bit for health check to run
	time.Sleep(200 * time.Millisecond)

	// Verify health check is running (simplified check)
	// Note: The actual implementation doesn't track running state
	// We just verify the method doesn't panic and runs successfully
	assert.NotNil(t, discovery)
}

func TestServiceDiscovery_GetServiceInfo(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Register service and instances
	service := &MockService{
		name:    "test-service",
		version: "1.0.0",
		health: ServiceHealth{
			Status:    "healthy",
			Message:   "Service is running",
			Timestamp: time.Now(),
		},
	}

	registry.Register(service)

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

	// Get service info
	info, err := discovery.GetServiceInfo("test-service")
	require.NoError(t, err)

	// Verify service info
	assert.Equal(t, "test-service", info["name"])
	assert.Equal(t, "1.0.0", info["version"])
	assert.Equal(t, "healthy", info["health_status"])
	assert.Equal(t, 2, info["instance_count"])
	assert.NotNil(t, info["instances"])
	assert.NotNil(t, info["capabilities"])
}

func TestServiceDiscovery_GetServiceInfo_NotFound(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Try to get info for non-existent service
	info, err := discovery.GetServiceInfo("non-existent")
	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "service not found")
}

// MockService implements ServiceContract for testing
type MockService struct {
	name    string
	version string
	health  ServiceHealth
}

func (s *MockService) ServiceName() string {
	return s.name
}

func (s *MockService) Version() string {
	return s.version
}

func (s *MockService) Health() ServiceHealth {
	return s.health
}

func (s *MockService) Capabilities() []ServiceCapability {
	return []ServiceCapability{
		{
			Name:        "test-capability",
			Description: "Test capability for unit testing",
			Version:     "1.0.0",
		},
	}
}

// Test concurrent operations
func TestServiceDiscovery_ConcurrentOperations(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Start multiple goroutines to test concurrent operations
	const numGoroutines = 10
	const numOperations = 100

	// Channel to signal completion
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				instance := ServiceInstance{
					ID:          fmt.Sprintf("instance-%d-%d", id, j),
					ServiceName: fmt.Sprintf("service-%d", id),
					Version:     "1.0.0",
					Host:        "localhost",
					Port:        8080 + j,
					Protocol:    "http",
					LastSeen:    time.Now(),
				}

				// Register instance
				err := discovery.RegisterInstance(instance)
				if err != nil {
					t.Logf("Failed to register instance: %v", err)
					continue
				}

				// Discover instances
				instances, err := discovery.Discover(fmt.Sprintf("service-%d", id))
				if err != nil {
					t.Logf("Failed to discover instances: %v", err)
					continue
				}

				// Update health
				if len(instances) > 0 {
					health := ServiceHealth{
						Status:    "healthy",
						Message:   "Service is running",
						Timestamp: time.Now(),
					}
					discovery.UpdateInstanceHealth(fmt.Sprintf("service-%d", id), instances[0].ID, health)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify no panics occurred and some instances were registered
	allServices := discovery.DiscoverAll()
	assert.Greater(t, len(allServices), 0)
}

// Test event notification
func TestServiceDiscovery_EventNotification(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}
	logger := observability.NewLogger(cfg)
	registry := NewServiceRegistry(logger)
	discovery := NewServiceDiscovery(logger, registry)

	// Watch a service
	eventChan, err := discovery.Watch("test-service")
	require.NoError(t, err)

	// Register instance in a goroutine
	go func() {
		instance := ServiceInstance{
			ID:          "instance-1",
			ServiceName: "test-service",
			Version:     "1.0.0",
			Host:        "localhost",
			Port:        8080,
			Protocol:    "http",
			LastSeen:    time.Now(),
		}
		discovery.RegisterInstance(instance)
	}()

	// Wait for event
	select {
	case event := <-eventChan:
		assert.Equal(t, "test-service", event.Instance.ServiceName)
		assert.Equal(t, ServiceEventTypeAdded, event.Type)
		assert.Equal(t, "instance-1", event.Instance.ID)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for event")
	}

	// Cleanup
	discovery.Unwatch("test-service")
}
