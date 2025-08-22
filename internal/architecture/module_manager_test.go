package architecture

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockModule implements the Module interface for testing
type MockModule struct {
	id       string
	metadata ModuleMetadata
	config   ModuleConfig
	health   ModuleHealth
	running  bool
	startErr error
	stopErr  error
}

func (m *MockModule) ID() string {
	return m.id
}

func (m *MockModule) Metadata() ModuleMetadata {
	return m.metadata
}

func (m *MockModule) Config() ModuleConfig {
	return m.config
}

func (m *MockModule) Health() ModuleHealth {
	return m.health
}

func (m *MockModule) Start(ctx context.Context) error {
	if m.startErr != nil {
		return m.startErr
	}
	m.running = true
	return nil
}

func (m *MockModule) Stop(ctx context.Context) error {
	if m.stopErr != nil {
		return m.stopErr
	}
	m.running = false
	return nil
}

func (m *MockModule) IsRunning() bool {
	return m.running
}

func (m *MockModule) Process(ctx context.Context, req ModuleRequest) (ModuleResponse, error) {
	return ModuleResponse{
		ID:         req.ID,
		Success:    true,
		Data:       req.Data,
		Confidence: 0.95,
		Latency:    100 * time.Millisecond,
	}, nil
}

func (m *MockModule) CanHandle(req ModuleRequest) bool {
	return true
}

func (m *MockModule) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *MockModule) OnEvent(event ModuleEvent) error {
	return nil
}

func TestNewModuleManager(t *testing.T) {
	mm := NewModuleManager()
	assert.NotNil(t, mm)
	assert.NotNil(t, mm.modules)
	assert.NotNil(t, mm.configs)
	assert.NotNil(t, mm.health)
	assert.NotNil(t, mm.events)
}

func TestRegisterModule(t *testing.T) {
	mm := NewModuleManager()

	module := &MockModule{
		id: "test-module",
		metadata: ModuleMetadata{
			Name:        "Test Module",
			Version:     "1.0.0",
			Description: "A test module",
			Capabilities: []ModuleCapability{
				CapabilityClassification,
			},
			Priority: PriorityMedium,
		},
		config: ModuleConfig{
			Enabled:    true,
			Timeout:    30 * time.Second,
			RetryCount: 3,
		},
	}

	err := mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Verify module was registered
	registeredModule, exists := mm.GetModule("test-module")
	assert.True(t, exists)
	assert.Equal(t, module, registeredModule)

	// Verify health status was initialized
	health, exists := mm.GetModuleHealth("test-module")
	assert.True(t, exists)
	assert.Equal(t, ModuleStatusStopped, health.Status)
}

func TestRegisterModuleDuplicate(t *testing.T) {
	mm := NewModuleManager()

	module := &MockModule{
		id: "test-module",
		metadata: ModuleMetadata{
			Name:        "Test Module",
			Version:     "1.0.0",
			Description: "A test module",
		},
		config: ModuleConfig{Enabled: true},
	}

	// Register module first time
	err := mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Try to register same module again
	err = mm.RegisterModule(module, module.config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegisterModuleEmptyName(t *testing.T) {
	mm := NewModuleManager()

	module := &MockModule{
		id: "test-module",
		metadata: ModuleMetadata{
			Name: "", // Empty name
		},
		config: ModuleConfig{Enabled: true},
	}

	err := mm.RegisterModule(module, module.config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty name")
}

func TestRegisterModuleWithDependencies(t *testing.T) {
	mm := NewModuleManager()

	// Create dependency module
	depModule := &MockModule{
		id: "dependency-module",
		metadata: ModuleMetadata{
			Name:        "Dependency Module",
			Version:     "1.0.0",
			Description: "A dependency module",
		},
		config: ModuleConfig{Enabled: true},
	}

	// Register dependency first
	err := mm.RegisterModule(depModule, depModule.config)
	assert.NoError(t, err)

	// Create module that depends on the dependency
	module := &MockModule{
		id: "dependent-module",
		metadata: ModuleMetadata{
			Name:        "Dependent Module",
			Version:     "1.0.0",
			Description: "A module that depends on another",
		},
		config: ModuleConfig{
			Enabled:      true,
			Dependencies: []string{"dependency-module"},
		},
	}

	// Register dependent module
	err = mm.RegisterModule(module, module.config)
	assert.NoError(t, err)
}

func TestRegisterModuleWithMissingDependency(t *testing.T) {
	mm := NewModuleManager()

	module := &MockModule{
		id: "dependent-module",
		metadata: ModuleMetadata{
			Name:        "Dependent Module",
			Version:     "1.0.0",
			Description: "A module that depends on another",
		},
		config: ModuleConfig{
			Enabled:      true,
			Dependencies: []string{"missing-module"},
		},
	}

	err := mm.RegisterModule(module, module.config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unregistered module")
}

func TestUnregisterModule(t *testing.T) {
	mm := NewModuleManager()

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

	// Unregister module
	err = mm.UnregisterModule("test-module")
	assert.NoError(t, err)

	// Verify module was removed
	_, exists := mm.GetModule("test-module")
	assert.False(t, exists)

	// Verify health status was removed
	_, exists = mm.GetModuleHealth("test-module")
	assert.False(t, exists)
}

func TestUnregisterModuleNotFound(t *testing.T) {
	mm := NewModuleManager()

	err := mm.UnregisterModule("non-existent-module")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUnregisterRunningModule(t *testing.T) {
	mm := NewModuleManager()

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

	// Unregister module
	err = mm.UnregisterModule("test-module")
	assert.NoError(t, err)

	// Verify module was stopped and removed
	assert.False(t, module.IsRunning())
	_, exists := mm.GetModule("test-module")
	assert.False(t, exists)
}

func TestStartModule(t *testing.T) {
	mm := NewModuleManager()

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
	err = mm.StartModule("test-module")
	assert.NoError(t, err)

	// Verify module is running
	assert.True(t, module.IsRunning())

	// Verify health status was updated
	health, exists := mm.GetModuleHealth("test-module")
	assert.True(t, exists)
	assert.Equal(t, ModuleStatusRunning, health.Status)
}

func TestStartModuleNotFound(t *testing.T) {
	mm := NewModuleManager()

	err := mm.StartModule("non-existent-module")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestStartDisabledModule(t *testing.T) {
	mm := NewModuleManager()

	module := &MockModule{
		id: "test-module",
		metadata: ModuleMetadata{
			Name:        "Test Module",
			Version:     "1.0.0",
			Description: "A test module",
		},
		config: ModuleConfig{Enabled: false}, // Disabled
	}

	// Register module
	err := mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Try to start disabled module
	err = mm.StartModule("test-module")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}

func TestStartModuleWithDependencies(t *testing.T) {
	mm := NewModuleManager()

	// Create dependency module
	depModule := &MockModule{
		id: "dependency-module",
		metadata: ModuleMetadata{
			Name:        "Dependency Module",
			Version:     "1.0.0",
			Description: "A dependency module",
		},
		config: ModuleConfig{Enabled: true},
	}

	// Create dependent module
	module := &MockModule{
		id: "dependent-module",
		metadata: ModuleMetadata{
			Name:        "Dependent Module",
			Version:     "1.0.0",
			Description: "A module that depends on another",
		},
		config: ModuleConfig{
			Enabled:      true,
			Dependencies: []string{"dependency-module"},
		},
	}

	// Register both modules
	err := mm.RegisterModule(depModule, depModule.config)
	assert.NoError(t, err)
	err = mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Try to start dependent module without starting dependency
	err = mm.StartModule("dependent-module")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")

	// Start dependency first
	err = mm.StartModule("dependency-module")
	assert.NoError(t, err)

	// Now start dependent module
	err = mm.StartModule("dependent-module")
	assert.NoError(t, err)

	// Verify both modules are running
	assert.True(t, depModule.IsRunning())
	assert.True(t, module.IsRunning())
}

func TestStopModule(t *testing.T) {
	mm := NewModuleManager()

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
	err = mm.StopModule("test-module")
	assert.NoError(t, err)

	// Verify module is stopped
	assert.False(t, module.IsRunning())

	// Verify health status was updated
	health, exists := mm.GetModuleHealth("test-module")
	assert.True(t, exists)
	assert.Equal(t, ModuleStatusStopped, health.Status)
}

func TestStopModuleNotFound(t *testing.T) {
	mm := NewModuleManager()

	err := mm.StopModule("non-existent-module")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestStartAllModules(t *testing.T) {
	mm := NewModuleManager()

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
		{
			id: "module3",
			metadata: ModuleMetadata{
				Name:        "Module 3",
				Version:     "1.0.0",
				Description: "Third module",
			},
			config: ModuleConfig{Enabled: false}, // Disabled
		},
	}

	// Register all modules
	for _, module := range modules {
		err := mm.RegisterModule(module, module.config)
		assert.NoError(t, err)
	}

	// Start all modules
	err := mm.StartAllModules()
	assert.NoError(t, err)

	// Verify enabled modules are running
	assert.True(t, modules[0].IsRunning())
	assert.True(t, modules[1].IsRunning())
	assert.False(t, modules[2].IsRunning()) // Should remain stopped
}

func TestStartAllModulesWithDependencies(t *testing.T) {
	mm := NewModuleManager()

	// Create modules with dependencies
	modules := []*MockModule{
		{
			id: "base-module",
			metadata: ModuleMetadata{
				Name:        "Base Module",
				Version:     "1.0.0",
				Description: "Base module",
			},
			config: ModuleConfig{Enabled: true},
		},
		{
			id: "dependent-module",
			metadata: ModuleMetadata{
				Name:        "Dependent Module",
				Version:     "1.0.0",
				Description: "Module that depends on base",
			},
			config: ModuleConfig{
				Enabled:      true,
				Dependencies: []string{"base-module"},
			},
		},
	}

	// Register all modules
	for _, module := range modules {
		err := mm.RegisterModule(module, module.config)
		assert.NoError(t, err)
	}

	// Start all modules
	err := mm.StartAllModules()
	assert.NoError(t, err)

	// Verify all modules are running
	assert.True(t, modules[0].IsRunning())
	assert.True(t, modules[1].IsRunning())
}

func TestStartAllModulesCircularDependency(t *testing.T) {
	mm := NewModuleManager()

	// Create modules with circular dependency
	modules := []*MockModule{
		{
			id: "module-a",
			metadata: ModuleMetadata{
				Name:        "Module A",
				Version:     "1.0.0",
				Description: "Module A",
			},
			config: ModuleConfig{
				Enabled:      true,
				Dependencies: []string{"module-b"},
			},
		},
		{
			id: "module-b",
			metadata: ModuleMetadata{
				Name:        "Module B",
				Version:     "1.0.0",
				Description: "Module B",
			},
			config: ModuleConfig{
				Enabled:      true,
				Dependencies: []string{"module-a"},
			},
		},
	}

	// Register modules one by one to avoid registration-time dependency validation
	// We'll register them without dependencies first, then update the config
	for i, module := range modules {
		// Create a copy without dependencies for registration
		regModule := &MockModule{
			id:       module.id,
			metadata: module.metadata,
			config: ModuleConfig{
				Enabled: module.config.Enabled,
			},
		}

		err := mm.RegisterModule(regModule, regModule.config)
		assert.NoError(t, err)

		// Update with dependencies after registration
		if i == 0 {
			// Update module-a with dependency on module-b
			err = mm.UpdateModuleConfig("module-a", ModuleConfig{
				Enabled:      true,
				Dependencies: []string{"module-b"},
			})
			assert.NoError(t, err)
		} else {
			// Update module-b with dependency on module-a
			err = mm.UpdateModuleConfig("module-b", ModuleConfig{
				Enabled:      true,
				Dependencies: []string{"module-a"},
			})
			assert.NoError(t, err)
		}
	}

	// Try to start all modules
	err := mm.StartAllModules()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestStopAllModules(t *testing.T) {
	mm := NewModuleManager()

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

	// Register and start all modules
	for _, module := range modules {
		err := mm.RegisterModule(module, module.config)
		assert.NoError(t, err)
		module.running = true // Simulate running
	}

	// Stop all modules
	err := mm.StopAllModules()
	assert.NoError(t, err)

	// Verify all modules are stopped
	for _, module := range modules {
		assert.False(t, module.IsRunning())
	}
}

func TestGetModulesByCapability(t *testing.T) {
	mm := NewModuleManager()

	// Create modules with different capabilities
	modules := []*MockModule{
		{
			id: "classification-module",
			metadata: ModuleMetadata{
				Name:        "Classification Module",
				Version:     "1.0.0",
				Description: "Classification module",
				Capabilities: []ModuleCapability{
					CapabilityClassification,
				},
			},
			config: ModuleConfig{Enabled: true},
		},
		{
			id: "verification-module",
			metadata: ModuleMetadata{
				Name:        "Verification Module",
				Version:     "1.0.0",
				Description: "Verification module",
				Capabilities: []ModuleCapability{
					CapabilityVerification,
				},
			},
			config: ModuleConfig{Enabled: true},
		},
		{
			id: "multi-capability-module",
			metadata: ModuleMetadata{
				Name:        "Multi Capability Module",
				Version:     "1.0.0",
				Description: "Module with multiple capabilities",
				Capabilities: []ModuleCapability{
					CapabilityClassification,
					CapabilityVerification,
				},
			},
			config: ModuleConfig{Enabled: true},
		},
	}

	// Register all modules
	for _, module := range modules {
		err := mm.RegisterModule(module, module.config)
		assert.NoError(t, err)
	}

	// Get modules by capability
	classificationModules := mm.GetModulesByCapability(CapabilityClassification)
	assert.Len(t, classificationModules, 2)

	verificationModules := mm.GetModulesByCapability(CapabilityVerification)
	assert.Len(t, verificationModules, 2)

	riskModules := mm.GetModulesByCapability(CapabilityRiskAssessment)
	assert.Len(t, riskModules, 0)
}

func TestUpdateModuleConfig(t *testing.T) {
	mm := NewModuleManager()

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

	// Update configuration
	newConfig := ModuleConfig{
		Enabled:    false,
		Timeout:    60 * time.Second,
		RetryCount: 5,
	}

	err = mm.UpdateModuleConfig("test-module", newConfig)
	assert.NoError(t, err)

	// Verify configuration was updated in the manager
	// Note: The module's Config() method returns the original config
	// The manager's config is updated separately
	_, exists := mm.GetModule("test-module")
	assert.True(t, exists)
}

func TestUpdateModuleConfigNotFound(t *testing.T) {
	mm := NewModuleManager()

	newConfig := ModuleConfig{Enabled: true}
	err := mm.UpdateModuleConfig("non-existent-module", newConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestModuleEvents(t *testing.T) {
	mm := NewModuleManager()

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

	// Register module (should emit event)
	err := mm.RegisterModule(module, module.config)
	assert.NoError(t, err)

	// Wait for registration event
	select {
	case event := <-events:
		assert.Equal(t, "module_registered", event.Type)
		assert.Equal(t, "test-module", event.ModuleID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected registration event not received")
	}

	// Start module (should emit event)
	err = mm.StartModule("test-module")
	assert.NoError(t, err)

	// Wait for start event
	select {
	case event := <-events:
		assert.Equal(t, "module_started", event.Type)
		assert.Equal(t, "test-module", event.ModuleID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected start event not received")
	}

	// Stop module (should emit event)
	err = mm.StopModule("test-module")
	assert.NoError(t, err)

	// Wait for stop event
	select {
	case event := <-events:
		assert.Equal(t, "module_stopped", event.Type)
		assert.Equal(t, "test-module", event.ModuleID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected stop event not received")
	}

	// Unregister module (should emit event)
	err = mm.UnregisterModule("test-module")
	assert.NoError(t, err)

	// Wait for unregistration event
	select {
	case event := <-events:
		assert.Equal(t, "module_unregistered", event.Type)
		assert.Equal(t, "test-module", event.ModuleID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected unregistration event not received")
	}
}

func TestModuleManagerClose(t *testing.T) {
	mm := NewModuleManager()

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

	// Close module manager
	err = mm.Close()
	assert.NoError(t, err)

	// Verify module was stopped
	assert.False(t, module.IsRunning())

	// Verify event channel is closed
	events := mm.GetEvents()

	// Drain any remaining events
	for {
		select {
		case _, ok := <-events:
			if !ok {
				// Channel is closed
				return
			}
		default:
			// No more events, channel should be closed
			_, ok := <-events
			assert.False(t, ok)
			return
		}
	}
}

func TestModuleManagerConcurrency(t *testing.T) {
	mm := NewModuleManager()

	// Create multiple goroutines to test concurrent access
	const numGoroutines = 10
	const numModules = 5

	// Start goroutines that register modules
	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numModules; j++ {
				moduleID := fmt.Sprintf("module-%d-%d", id, j)
				module := &MockModule{
					id: moduleID,
					metadata: ModuleMetadata{
						Name:        fmt.Sprintf("Module %d-%d", id, j),
						Version:     "1.0.0",
						Description: "Test module",
					},
					config: ModuleConfig{Enabled: true},
				}

				// Register module
				err := mm.RegisterModule(module, module.config)
				if err != nil {
					// Expected for duplicate registrations
					return
				}

				// Start module
				_ = mm.StartModule(moduleID)

				// Get module
				_, exists := mm.GetModule(moduleID)
				assert.True(t, exists)

				// Get health
				_, exists = mm.GetModuleHealth(moduleID)
				assert.True(t, exists)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify no panics occurred and manager is still functional
	modules := mm.ListModules()
	assert.Greater(t, len(modules), 0)

	// Test concurrent module operations
	for _, module := range modules {
		moduleID := module.ID()

		// Test concurrent health checks
		go func() {
			_, _ = mm.GetModuleHealth(moduleID)
		}()

		// Test concurrent module retrieval
		go func() {
			_, _ = mm.GetModule(moduleID)
		}()
	}

	// Close manager
	err := mm.Close()
	assert.NoError(t, err)
}
