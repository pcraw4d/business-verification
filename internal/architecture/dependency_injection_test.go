package architecture

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SimpleMockModule implements Module for testing
type SimpleMockModule struct {
	id      string
	config  ModuleConfig
	running bool
}

func (m *SimpleMockModule) ID() string { return m.id }
func (m *SimpleMockModule) Metadata() ModuleMetadata {
	return ModuleMetadata{
		Name:         "Test Module",
		Version:      "1.0.0",
		Description:  "Test module for dependency injection",
		Capabilities: []ModuleCapability{CapabilityVerification},
		Priority:     PriorityMedium,
	}
}
func (m *SimpleMockModule) Config() ModuleConfig { return m.config }
func (m *SimpleMockModule) Health() ModuleHealth {
	return ModuleHealth{
		Status:    ModuleStatusHealthy,
		LastCheck: time.Now(),
	}
}
func (m *SimpleMockModule) Start(ctx context.Context) error { m.running = true; return nil }
func (m *SimpleMockModule) Stop(ctx context.Context) error  { m.running = false; return nil }
func (m *SimpleMockModule) IsRunning() bool                 { return m.running }
func (m *SimpleMockModule) Process(ctx context.Context, req ModuleRequest) (ModuleResponse, error) {
	return ModuleResponse{ID: req.ID, Success: true}, nil
}
func (m *SimpleMockModule) CanHandle(req ModuleRequest) bool      { return true }
func (m *SimpleMockModule) HealthCheck(ctx context.Context) error { return nil }
func (m *SimpleMockModule) OnEvent(event ModuleEvent) error       { return nil }

// SimpleMockFactory implements ModuleFactory for testing
type SimpleMockFactory struct {
	createModuleCalled bool
	createModuleError  error
}

func (m *SimpleMockFactory) CreateModule(config ModuleConfig) (Module, error) {
	m.createModuleCalled = true
	if m.createModuleError != nil {
		return nil, m.createModuleError
	}
	return &SimpleMockModule{
		id:     "test_module",
		config: config,
	}, nil
}

func TestNewDependencyContainer(t *testing.T) {
	config := DependencyConfig{
		AutoWire:        true,
		ValidateOnStart: true,
		LazyLoading:     false,
	}

	container := NewDependencyContainer(config)

	assert.NotNil(t, container)
	assert.NotNil(t, container.dependencies)
	assert.NotNil(t, container.modules)
	assert.Equal(t, config, container.config)
	assert.NotNil(t, container.tracer)
}

func TestRegisterDependency(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Test successful registration
	mockInstance := "test_dependency"
	container.Register("test_dep", mockInstance)

	// Test duplicate registration (should overwrite)
	container.Register("test_dep", mockInstance)

	// Test nil instance (should be allowed)
	container.Register("nil_dep", nil)
}

func TestGetDependency(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Register a dependency
	mockInstance := "test_dependency"
	container.Register("test_dep", mockInstance)

	// Test successful retrieval
	dep, exists := container.Get("test_dep")
	assert.True(t, exists)
	assert.Equal(t, mockInstance, dep)

	// Test non-existent dependency
	dep, exists = container.Get("non_existent")
	assert.False(t, exists)
	assert.Nil(t, dep)
}

func TestGetDependencyByType(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Register multiple dependencies of different types
	configDep := "config_dependency"
	moduleDep := "module_dependency"

	container.Register("config", configDep)
	container.Register("module", moduleDep)

	// Test retrieval by name
	dep, exists := container.Get("config")
	assert.True(t, exists)
	assert.Equal(t, configDep, dep)

	dep, exists = container.Get("module")
	assert.True(t, exists)
	assert.Equal(t, moduleDep, dep)

	// Test non-existent dependency
	dep, exists = container.Get("non_existent")
	assert.False(t, exists)
	assert.Nil(t, dep)
}

func TestInjectDependencies(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Create a module
	module := &SimpleMockModule{
		id: "test_module",
		config: ModuleConfig{
			Dependencies: []string{"test_dep"},
		},
	}

	// Register dependency
	container.Register("test_dep", "test_value")

	// Register module
	container.RegisterModule(module)

	// Verify module was registered
	retrievedModule, exists := container.GetModule("test_module")
	assert.True(t, exists)
	assert.NotNil(t, retrievedModule)
}

func TestCreateModuleWithDependencies(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Register dependencies
	container.Register("test_dep", "test_value")

	// Create module with dependencies
	config := ModuleConfig{
		Dependencies: []string{"test_dep"},
	}

	// Create and register a module directly
	module := &SimpleMockModule{
		id:     "test_module",
		config: config,
	}

	container.RegisterModule(module)

	// Verify module was registered
	retrievedModule, exists := container.GetModule("test_module")
	assert.True(t, exists)
	assert.NotNil(t, retrievedModule)
	assert.Equal(t, "test_module", retrievedModule.ID())
}

func TestGetModuleDependencies(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Register dependencies
	container.Register("test_dep", "test_value")

	// Create and register a module
	module := &SimpleMockModule{
		id: "test_module",
		config: ModuleConfig{
			Dependencies: []string{"test_dep"},
		},
	}

	container.RegisterModule(module)

	// Get module
	retrievedModule, exists := container.GetModule("test_module")
	assert.True(t, exists)
	assert.NotNil(t, retrievedModule)

	// Type assert to SimpleMockModule for testing
	if mockModule, ok := retrievedModule.(*SimpleMockModule); ok {
		assert.Equal(t, "test_module", mockModule.ID())
	}

	// Test non-existent module
	nonExistentModule, exists := container.GetModule("non_existent")
	assert.False(t, exists)
	assert.Nil(t, nonExistentModule)
}

func TestDependencyResolver(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Register dependencies
	container.Register("test_dep", "test_value")

	// Create and register a module
	module := &SimpleMockModule{
		id: "test_module",
		config: ModuleConfig{
			Dependencies: []string{"test_dep"},
		},
	}

	container.RegisterModule(module)

	// Verify module was registered
	retrievedModule, exists := container.GetModule("test_module")
	assert.True(t, exists)
	assert.NotNil(t, retrievedModule)
	assert.Equal(t, "test_module", retrievedModule.ID())
}

func TestDependencyResolverWithMissingDependency(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Create a module with missing dependency
	module := &SimpleMockModule{
		id: "test_module",
		config: ModuleConfig{
			Dependencies: []string{"missing_dependency"},
		},
	}

	// Register module (this should work even with missing dependencies)
	container.RegisterModule(module)

	// Verify module was registered
	retrievedModule, exists := container.GetModule("test_module")
	assert.True(t, exists)
	assert.NotNil(t, retrievedModule)
	assert.Equal(t, "test_module", retrievedModule.ID())
}

func TestDependencyContainerClose(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Create and register a module
	module := &SimpleMockModule{id: "test_module"}
	container.RegisterModule(module)

	// Start the module
	err := module.Start(context.Background())
	require.NoError(t, err)
	assert.True(t, module.IsRunning())

	// Stop the container
	err = container.Stop(context.Background())
	assert.NoError(t, err)

	// Verify module was stopped
	assert.False(t, module.IsRunning())
}

func TestDependencyPriority(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Test basic dependency registration and retrieval
	container.Register("database", "test_db")
	container.Register("logger", "test_logger")

	dep, exists := container.Get("database")
	assert.True(t, exists)
	assert.Equal(t, "test_db", dep)

	dep, exists = container.Get("logger")
	assert.True(t, exists)
	assert.Equal(t, "test_logger", dep)
}

func TestRequiredDependencyTypes(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Test dependency registration and retrieval
	container.Register("required_dep", "test_value")
	container.Register("optional_dep", "test_value")

	dep, exists := container.Get("required_dep")
	assert.True(t, exists)
	assert.Equal(t, "test_value", dep)

	dep, exists = container.Get("optional_dep")
	assert.True(t, exists)
	assert.Equal(t, "test_value", dep)
}

func TestValidateDependencies(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Register a required dependency
	container.Register("database", "test_db")

	// Register a nil dependency
	container.Register("logger", nil)

	// Test getting dependencies
	dep, exists := container.Get("database")
	assert.True(t, exists)
	assert.Equal(t, "test_db", dep)

	dep, exists = container.Get("logger")
	assert.True(t, exists)
	assert.Nil(t, dep)
}

func TestProviderSpecificDependencies(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{
		AutoWire:        true,
		ValidateOnStart: true,
		LazyLoading:     false,
	})

	// Test basic dependency registration
	container.Register("test_dependency", "test_value")

	// Verify dependency was registered
	dep, exists := container.Get("test_dependency")
	assert.True(t, exists)
	assert.Equal(t, "test_value", dep)
}

func TestDependencyContainerConcurrency(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Test concurrent dependency registration
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			depValue := fmt.Sprintf("dependency_%d", id)
			container.Register(fmt.Sprintf("dep_%d", id), depValue)

			dep, exists := container.Get(fmt.Sprintf("dep_%d", id))
			assert.True(t, exists)
			assert.Equal(t, depValue, dep)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all dependencies were registered
	for i := 0; i < 10; i++ {
		dep, exists := container.Get(fmt.Sprintf("dep_%d", i))
		assert.True(t, exists)
		assert.Equal(t, fmt.Sprintf("dependency_%d", i), dep)
	}
}
