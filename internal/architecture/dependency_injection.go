package architecture

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/trace"
)

// DependencyConfig represents the configuration for dependency injection
type DependencyConfig struct {
	AutoWire        bool
	ValidateOnStart bool
	LazyLoading     bool
}

// DependencyTypeConfig represents configuration for a specific dependency type
type DependencyTypeConfig struct {
	Type        string
	Interface   interface{}
	Implementation interface{}
	Singleton   bool
}

// DependencyContainer manages dependency injection
type DependencyContainer struct {
	dependencies map[string]interface{}
	modules      map[string]Module
	config       DependencyConfig
	tracer       trace.Tracer
	mu           sync.RWMutex
}

// NewDependencyContainer creates a new dependency container
func NewDependencyContainer(config DependencyConfig) *DependencyContainer {
	return &DependencyContainer{
		dependencies: make(map[string]interface{}),
		modules:      make(map[string]Module),
		config:       config,
		tracer:       trace.NewNoopTracerProvider().Tracer("dependency-container"),
	}
}

// Register registers a dependency
func (dc *DependencyContainer) Register(name string, dependency interface{}) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.dependencies[name] = dependency
}

// Get retrieves a dependency
func (dc *DependencyContainer) Get(name string) (interface{}, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	dependency, exists := dc.dependencies[name]
	return dependency, exists
}

// RegisterModule registers a module
func (dc *DependencyContainer) RegisterModule(module Module) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.modules[module.ID()] = module
}

// GetModule retrieves a module
func (dc *DependencyContainer) GetModule(id string) (Module, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	module, exists := dc.modules[id]
	return module, exists
}

// Start initializes the dependency container
func (dc *DependencyContainer) Start(ctx context.Context) error {
	if dc.config.ValidateOnStart {
		return dc.validateDependencies()
	}
	return nil
}

// Stop shuts down the dependency container
func (dc *DependencyContainer) Stop(ctx context.Context) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	
	// Stop all modules
	for _, module := range dc.modules {
		if module.IsRunning() {
			module.Stop(ctx)
		}
	}
	
	// Clear dependencies
	dc.dependencies = make(map[string]interface{})
	dc.modules = make(map[string]Module)
	
	return nil
}

// validateDependencies validates all registered dependencies
func (dc *DependencyContainer) validateDependencies() error {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	
	// Basic validation - check that all dependencies are non-nil
	for name, dependency := range dc.dependencies {
		if dependency == nil {
			return &DependencyError{
				Name:    name,
				Message: "dependency is nil",
			}
		}
	}
	
	return nil
}

// DependencyError represents a dependency-related error
type DependencyError struct {
	Name    string
	Message string
}

func (e *DependencyError) Error() string {
	return "dependency error [" + e.Name + "]: " + e.Message
}
