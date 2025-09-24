package architecture

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// FailingMockModule is a mock module that fails health checks
type FailingMockModule struct {
	MockModule
}

func (m *FailingMockModule) HealthCheck(ctx context.Context) error {
	return fmt.Errorf("health check failed")
}

func TestNewLifecycleManager(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
		MaxRetries:          3,
		RetryDelay:          1 * time.Second,
		AutoRestart:         true,
		AutoRestartDelay:    5 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)
	assert.NotNil(t, lm)
	assert.Equal(t, mm, lm.moduleManager)
	assert.Equal(t, config, lm.config)
	assert.NotNil(t, lm.states)
	assert.NotNil(t, lm.healthResults)
	assert.NotNil(t, lm.healthTickers)
	assert.NotNil(t, lm.stopChans)
}

func TestLifecycleStartModule(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	module := &MockModule{
		id: "test-module",
		metadata: ModuleMetadata{
			Name:        "Test Module",
			Version:     "1.0.0",
			Description: "A test module",
		},
		config: ModuleConfig{Enabled: true},
	}

	// Register module
	err := mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Start module
	err = lm.StartModule("test-module")
	assert.NoError(t, err)

	// Verify module is running
	assert.True(t, module.IsRunning())

	// Verify state
	state, exists := lm.GetModuleState("test-module")
	assert.True(t, exists)
	assert.Equal(t, LifecycleStateRunning, state)
}

func TestLifecycleStartModuleNotFound(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	err := lm.StartModule("non-existent-module")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLifecycleStartModuleAlreadyRunning(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	module := &MockModule{
		id: "test-module",
		metadata: ModuleMetadata{
			Name:        "Test Module",
			Version:     "1.0.0",
			Description: "A test module",
		},
		config: ModuleConfig{Enabled: true},
	}

	// Register module
	err := mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Start module first time
	err = lm.StartModule("test-module")
	assert.NoError(t, err)

	// Try to start again
	err = lm.StartModule("test-module")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

func TestStopModule(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	module := &MockModule{
		id: "test-module",
		metadata: ModuleMetadata{
			Name:        "Test Module",
			Version:     "1.0.0",
			Description: "A test module",
		},
		config: ModuleConfig{Enabled: true},
	}
	module.running = true // Simulate running module

	// Register module
	err := mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Stop module
	err = lm.StopModule("test-module")
	assert.NoError(t, err)

	// Verify module is stopped
	assert.False(t, module.IsRunning())

	// Verify state
	state, exists := lm.GetModuleState("test-module")
	assert.True(t, exists)
	assert.Equal(t, LifecycleStateStopped, state)
}

func TestStopModuleNotFound(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	err := lm.StopModule("non-existent-module")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestStopModuleNotRunning(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	module := &MockModule{
		id: "test-module",
		metadata: ModuleMetadata{
			Name:        "Test Module",
			Version:     "1.0.0",
			Description: "A test module",
		},
		config: ModuleConfig{Enabled: true},
	}
	module.running = false // Simulate stopped module

	// Register module
	err := mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Try to stop already stopped module
	err = lm.StopModule("test-module")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestRestartModule(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
		AutoRestartDelay:    100 * time.Millisecond, // Short delay for testing
	}

	lm := NewLifecycleManager(mm, config)

	module := &MockModule{
		id: "test-module",
		metadata: ModuleMetadata{
			Name:        "Test Module",
			Version:     "1.0.0",
			Description: "A test module",
		},
		config: ModuleConfig{Enabled: true},
	}
	module.running = true // Simulate running module

	// Register module
	err := mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Restart module
	err = lm.RestartModule("test-module")
	assert.NoError(t, err)

	// Verify module is running after restart
	assert.True(t, module.IsRunning())

	// Verify state
	state, exists := lm.GetModuleState("test-module")
	assert.True(t, exists)
	assert.Equal(t, LifecycleStateRunning, state)
}

func TestGetModuleState(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	module := &MockModule{
		id: "test-module",
		metadata: ModuleMetadata{
			Name:        "Test Module",
			Version:     "1.0.0",
			Description: "A test module",
		},
		config: ModuleConfig{Enabled: true},
	}

	// Register module
	err := mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Get state before starting
	state, exists := lm.GetModuleState("test-module")
	assert.True(t, exists)
	assert.Equal(t, LifecycleStateInitialized, state)

	// Start module
	err = lm.StartModule("test-module")
	assert.NoError(t, err)

	// Get state after starting
	state, exists = lm.GetModuleState("test-module")
	assert.True(t, exists)
	assert.Equal(t, LifecycleStateRunning, state)

	// Get state for non-existent module
	state, exists = lm.GetModuleState("non-existent-module")
	assert.False(t, exists)
	assert.Equal(t, LifecycleStateInitialized, state)
}

func TestGetAllModuleStates(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	// Create multiple modules
	modules := []*MockModule{
		{
			id: "module1",
			metadata: ModuleMetadata{
				Name:        "Module 1",
				Version:     "1.0.0",
				Description: "First module",
			},
			config: ModuleConfig{Enabled: true},
		},
		{
			id: "module2",
			metadata: ModuleMetadata{
				Name:        "Module 2",
				Version:     "1.0.0",
				Description: "Second module",
			},
			config: ModuleConfig{Enabled: true},
		},
	}

	// Register modules
	for _, module := range modules {
		err := mm.RegisterModule(module, module.config)
		assert.NoError(t, err)
	}

	// Start first module
	err := lm.StartModule("module1")
	assert.NoError(t, err)

	// Get all states
	states := lm.GetAllModuleStates()
	assert.Len(t, states, 2)
	assert.Equal(t, LifecycleStateRunning, states["module1"])
	assert.Equal(t, LifecycleStateInitialized, states["module2"])
}

func TestPerformHealthCheck(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	module := &MockModule{
		id: "test-module",
		metadata: ModuleMetadata{
			Name:        "Test Module",
			Version:     "1.0.0",
			Description: "A test module",
		},
		config: ModuleConfig{Enabled: true},
	}

	// Register module
	err := mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Perform health check
	result, err := lm.PerformHealthCheck("test-module")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-module", result.ModuleID)
	assert.True(t, result.Healthy)
	assert.Equal(t, ModuleStatusHealthy, result.Status)
	assert.Equal(t, "Health check passed", result.Message)
	assert.NotZero(t, result.Latency)
	assert.NotZero(t, result.Timestamp)
}

func TestPerformHealthCheckNotFound(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	// Perform health check on non-existent module
	result, err := lm.PerformHealthCheck("non-existent-module")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestPerformHealthCheckWithError(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	// Create a failing mock module
	failingMockModule := &FailingMockModule{
		MockModule: MockModule{
			id: "failing_module",
			metadata: ModuleMetadata{
				Name:        "Failing Module",
				Version:     "1.0.0",
				Description: "A module that fails health checks",
			},
			config: ModuleConfig{Enabled: true},
		},
	}

	// Register failing module
	err := mm.RegisterModule(failingMockModule, failingMockModule.config)
	assert.NoError(t, err)

	// Perform health check
	result, err := lm.PerformHealthCheck("failing-module")
	assert.NoError(t, err) // Health check error is captured in result, not returned
	assert.NotNil(t, result)
	assert.Equal(t, "failing-module", result.ModuleID)
	assert.False(t, result.Healthy)
	assert.Equal(t, ModuleStatusUnhealthy, result.Status)
	assert.Equal(t, "Health check failed", result.Message)
	assert.Contains(t, result.Error, "health check failed")
	assert.NotZero(t, result.Latency)
	assert.NotZero(t, result.Timestamp)
}

func TestGetHealthResult(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	module := &MockModule{
		id: "test-module",
		metadata: ModuleMetadata{
			Name:        "Test Module",
			Version:     "1.0.0",
			Description: "A test module",
		},
		config: ModuleConfig{Enabled: true},
	}

	// Register module
	err := mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Get health result before health check
	result, exists := lm.GetHealthResult("test-module")
	assert.False(t, exists)
	assert.Nil(t, result)

	// Perform health check
	_, err = lm.PerformHealthCheck("test-module")
	assert.NoError(t, err)

	// Get health result after health check
	result, exists = lm.GetHealthResult("test-module")
	assert.True(t, exists)
	assert.NotNil(t, result)
	assert.Equal(t, "test-module", result.ModuleID)
	assert.True(t, result.Healthy)
}

func TestGetAllHealthResults(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	// Create multiple modules
	modules := []*MockModule{
		{
			id: "module1",
			metadata: ModuleMetadata{
				Name:        "Module 1",
				Version:     "1.0.0",
				Description: "First module",
			},
			config: ModuleConfig{Enabled: true},
		},
		{
			id: "module2",
			metadata: ModuleMetadata{
				Name:        "Module 2",
				Version:     "1.0.0",
				Description: "Second module",
			},
			config: ModuleConfig{Enabled: true},
		},
	}

	// Register modules
	for _, module := range modules {
		err := mm.RegisterModule(module, module.config)
		assert.NoError(t, err)
	}

	// Perform health checks
	for _, module := range modules {
		_, err := lm.PerformHealthCheck(module.ID())
		assert.NoError(t, err)
	}

	// Get all health results
	results := lm.GetAllHealthResults()
	assert.Len(t, results, 2)
	assert.NotNil(t, results["module1"])
	assert.NotNil(t, results["module2"])
	assert.True(t, results["module1"].Healthy)
	assert.True(t, results["module2"].Healthy)
}

func TestStartAllModules(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	// Create multiple modules
	modules := []*MockModule{
		{
			id: "module1",
			metadata: ModuleMetadata{
				Name:        "Module 1",
				Version:     "1.0.0",
				Description: "First module",
			},
			config: ModuleConfig{Enabled: true},
		},
		{
			id: "module2",
			metadata: ModuleMetadata{
				Name:        "Module 2",
				Version:     "1.0.0",
				Description: "Second module",
			},
			config: ModuleConfig{Enabled: true},
		},
	}

	// Register modules
	for _, module := range modules {
		err := mm.RegisterModule(module, module.config)
		assert.NoError(t, err)
	}

	// Start all modules
	err := lm.StartAllModules()
	assert.NoError(t, err)

	// Verify all modules are running
	for _, module := range modules {
		assert.True(t, module.IsRunning())

		state, exists := lm.GetModuleState(module.ID())
		assert.True(t, exists)
		assert.Equal(t, LifecycleStateRunning, state)
	}
}

func TestStopAllModules(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	// Create multiple modules
	modules := []*MockModule{
		{
			id: "module1",
			metadata: ModuleMetadata{
				Name:        "Module 1",
				Version:     "1.0.0",
				Description: "First module",
			},
			config: ModuleConfig{Enabled: true},
		},
		{
			id: "module2",
			metadata: ModuleMetadata{
				Name:        "Module 2",
				Version:     "1.0.0",
				Description: "Second module",
			},
			config: ModuleConfig{Enabled: true},
		},
	}

	// Register and start modules
	for _, module := range modules {
		err := mm.RegisterModule(module, module.config)
		assert.NoError(t, err)
		module.running = true // Simulate running
	}

	// Stop all modules
	err := lm.StopAllModules()
	assert.NoError(t, err)

	// Verify all modules are stopped
	for _, module := range modules {
		assert.False(t, module.IsRunning())

		state, exists := lm.GetModuleState(module.ID())
		assert.True(t, exists)
		assert.Equal(t, LifecycleStateStopped, state)
	}
}

func TestLifecycleManagerClose(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	module := &MockModule{
		id: "test-module",
		metadata: ModuleMetadata{
			Name:        "Test Module",
			Version:     "1.0.0",
			Description: "A test module",
		},
		config: ModuleConfig{Enabled: true},
	}
	module.running = true // Simulate running module

	// Register module
	err := mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Close lifecycle manager
	err = lm.Close()
	assert.NoError(t, err)

	// Verify module was stopped
	assert.False(t, module.IsRunning())
}

func TestHealthMonitoring(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 100 * time.Millisecond, // Short interval for testing
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
		AutoRestart:         false, // Disable auto-restart for testing
	}

	lm := NewLifecycleManager(mm, config)

	module := &MockModule{
		id: "test-module",
		metadata: ModuleMetadata{
			Name:        "Test Module",
			Version:     "1.0.0",
			Description: "A test module",
		},
		config: ModuleConfig{Enabled: true},
	}

	// Register module
	err := mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Start module (this will start health monitoring)
	err = lm.StartModule("test-module")
	assert.NoError(t, err)

	// Wait for health check to be performed
	time.Sleep(200 * time.Millisecond)

	// Verify health result exists
	result, exists := lm.GetHealthResult("test-module")
	assert.True(t, exists)
	assert.NotNil(t, result)
	assert.True(t, result.Healthy)

	// Stop module (this will stop health monitoring)
	err = lm.StopModule("test-module")
	assert.NoError(t, err)
}

func TestLifecycleEvents(t *testing.T) {
	mm := NewModuleManager()
	config := LifecycleConfig{
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     10 * time.Second,
	}

	lm := NewLifecycleManager(mm, config)

	// Start listening for events
	events := mm.GetEvents()

	module := &MockModule{
		id: "test-module",
		metadata: ModuleMetadata{
			Name:        "Test Module",
			Version:     "1.0.0",
			Description: "A test module",
		},
		config: ModuleConfig{Enabled: true},
	}

	// Register module
	err := mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Start module (should emit lifecycle events)
	err = lm.StartModule("test-module")
	assert.NoError(t, err)

	// Wait for lifecycle events
	time.Sleep(100 * time.Millisecond)

	// Drain events to verify they were emitted
	eventCount := 0
	for {
		select {
		case event := <-events:
			if event.ModuleID == "test-module" {
				eventCount++
			}
		default:
			goto done
		}
	}
done:

	// Should have received lifecycle events
	assert.Greater(t, eventCount, 0)

	// Stop module
	err = lm.StopModule("test-module")
	assert.NoError(t, err)
}
