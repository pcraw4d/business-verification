package architecture

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
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
	err := container.RegisterDependency("test_dep", DependencyTypeConfig, mockInstance, nil)

	assert.NoError(t, err)

	// Test duplicate registration
	err = container.RegisterDependency("test_dep", DependencyTypeConfig, mockInstance, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Test nil instance
	err = container.RegisterDependency("nil_dep", DependencyTypeConfig, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestGetDependency(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Register a dependency
	mockInstance := "test_dependency"
	err := container.RegisterDependency("test_dep", DependencyTypeConfig, mockInstance, nil)
	require.NoError(t, err)

	// Test successful retrieval
	dep, err := container.GetDependency("test_dep")
	assert.NoError(t, err)
	assert.Equal(t, mockInstance, dep)

	// Test non-existent dependency
	dep, err = container.GetDependency("non_existent")
	assert.Error(t, err)
	assert.Nil(t, dep)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetDependencyByType(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Register multiple dependencies of different types
	configDep := "config_dependency"
	moduleDep := "module_dependency"

	err := container.RegisterDependency("config", DependencyTypeConfig, configDep, nil)
	require.NoError(t, err)

	err = container.RegisterDependency("module", DependencyTypeModule, moduleDep, nil)
	require.NoError(t, err)

	// Test retrieval by type
	configDeps := container.GetDependencyByType(DependencyTypeConfig)
	assert.Len(t, configDeps, 1)
	assert.Equal(t, configDep, configDeps[0])

	moduleDeps := container.GetDependencyByType(DependencyTypeModule)
	assert.Len(t, moduleDeps, 1)
	assert.Equal(t, moduleDep, moduleDeps[0])

	// Test non-existent type
	nonExistentDeps := container.GetDependencyByType("non_existent")
	assert.Len(t, nonExistentDeps, 0)
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
	err := container.RegisterDependency("test_dep", DependencyTypeConfig, "test_value", nil)
	require.NoError(t, err)

	// Inject dependencies
	err = container.InjectDependencies(module)
	assert.NoError(t, err)

	// Verify module was registered
	dependencies := container.GetModuleDependencies("test_module")
	assert.NotNil(t, dependencies)
}

func TestCreateModuleWithDependencies(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Register dependencies
	err := container.RegisterDependency("test_dep", DependencyTypeConfig, "test_value", nil)
	require.NoError(t, err)

	// Create module factory
	factory := &SimpleMockFactory{}

	// Create module with dependencies
	config := ModuleConfig{
		Dependencies: []string{"test_dep"},
	}

	module, err := container.CreateModuleWithDependencies(factory, config)
	assert.NoError(t, err)
	assert.NotNil(t, module)
	assert.True(t, factory.createModuleCalled)
	assert.Equal(t, "test_module", module.ID())
}

func TestGetModuleDependencies(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Register dependencies
	err := container.RegisterDependency("test_dep", DependencyTypeConfig, "test_value", nil)
	require.NoError(t, err)

	// Create and register a module
	module := &SimpleMockModule{
		id: "test_module",
		config: ModuleConfig{
			Dependencies: []string{"test_dep"},
		},
	}

	err = container.InjectDependencies(module)
	require.NoError(t, err)

	// Get module dependencies
	dependencies := container.GetModuleDependencies("test_module")
	assert.NotNil(t, dependencies)
	assert.Len(t, dependencies, 1)
	assert.Contains(t, dependencies, "test_dep")

	// Test non-existent module
	dependencies = container.GetModuleDependencies("non_existent")
	assert.Nil(t, dependencies)
}

func TestDependencyResolver(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Register dependencies
	err := container.RegisterDependency("test_dep", DependencyTypeConfig, "test_value", nil)
	require.NoError(t, err)

	// Create dependency resolver
	resolverConfig := DependencyInjectionConfig{
		AutoWire: true,
	}

	resolver := NewDependencyResolver(container, resolverConfig)

	// Create a module
	module := &SimpleMockModule{
		id: "test_module",
		config: ModuleConfig{
			Dependencies: []string{"test_dep"},
		},
	}

	// Resolve dependencies
	err = resolver.ResolveDependencies(module)
	assert.NoError(t, err)
}

func TestDependencyResolverWithMissingDependency(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	resolverConfig := DependencyInjectionConfig{
		AutoWire: true,
	}

	resolver := NewDependencyResolver(container, resolverConfig)

	// Create a module with missing dependency
	module := &SimpleMockModule{
		id: "test_module",
		config: ModuleConfig{
			Dependencies: []string{"missing_dependency"},
		},
	}

	// Resolve dependencies should fail
	err := resolver.ResolveDependencies(module)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDependencyContainerClose(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Create and register a module
	module := &SimpleMockModule{id: "test_module"}
	err := container.InjectDependencies(module)
	require.NoError(t, err)

	// Start the module
	err = module.Start(context.Background())
	require.NoError(t, err)
	assert.True(t, module.IsRunning())

	// Close the container
	err = container.Close()
	assert.NoError(t, err)

	// Verify module was stopped
	assert.False(t, module.IsRunning())
}

func TestDependencyPriority(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Test priority ordering
	assert.Equal(t, 1, container.getPriorityForType(DependencyTypeDatabase))
	assert.Equal(t, 2, container.getPriorityForType(DependencyTypeLogger))
	assert.Equal(t, 3, container.getPriorityForType(DependencyTypeMetrics))
	assert.Equal(t, 4, container.getPriorityForType(DependencyTypeTracer))
	assert.Equal(t, 5, container.getPriorityForType(DependencyTypeConfig))
	assert.Equal(t, 6, container.getPriorityForType(DependencyTypeModule))
	assert.Equal(t, 10, container.getPriorityForType("unknown"))
}

func TestRequiredDependencyTypes(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Test required types
	assert.True(t, container.isRequiredType(DependencyTypeDatabase))
	assert.True(t, container.isRequiredType(DependencyTypeLogger))

	// Test non-required types
	assert.False(t, container.isRequiredType(DependencyTypeMetrics))
	assert.False(t, container.isRequiredType(DependencyTypeTracer))
	assert.False(t, container.isRequiredType(DependencyTypeConfig))
	assert.False(t, container.isRequiredType(DependencyTypeModule))
}

func TestValidateDependencies(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Register a required dependency
	err := container.RegisterDependency("database", DependencyTypeDatabase, "test_db", nil)
	require.NoError(t, err)

	// Validation should pass
	err = container.validateDependencies()
	assert.NoError(t, err)

	// Register a nil required dependency
	err = container.RegisterDependency("logger", DependencyTypeLogger, nil, nil)
	require.NoError(t, err)

	// Validation should fail
	err = container.validateDependencies()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is nil")
}

func TestProviderSpecificDependencies(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{
		ProviderConfig: config.ProviderConfig{
			Database: "supabase",
		},
	})

	// Test Supabase dependency registration
	err := container.registerSupabaseDependencies()
	assert.NoError(t, err)

	// Test Railway dependency registration
	err = container.registerRailwayDependencies()
	assert.NoError(t, err)
}

func TestDependencyContainerConcurrency(t *testing.T) {
	container := NewDependencyContainer(DependencyConfig{})

	// Test concurrent dependency registration
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			depValue := fmt.Sprintf("dependency_%d", id)
			err := container.RegisterDependency(fmt.Sprintf("dep_%d", id), DependencyTypeConfig, depValue, nil)
			assert.NoError(t, err)

			dep, err := container.GetDependency(fmt.Sprintf("dep_%d", id))
			assert.NoError(t, err)
			assert.Equal(t, depValue, dep)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all dependencies were registered
	configDeps := container.GetDependencyByType(DependencyTypeConfig)
	assert.Len(t, configDeps, 10)
}
