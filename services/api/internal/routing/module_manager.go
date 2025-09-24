package routing

import (
	"context"
	"sync"
	"time"

	"kyb-platform/internal/architecture"
	"kyb-platform/internal/modules/database_classification"
	"kyb-platform/internal/observability"
)

// DefaultModuleManager implements the ModuleManager interface
type DefaultModuleManager struct {
	modules map[string]architecture.Module
	mutex   sync.RWMutex
	logger  *observability.Logger
}

// NewDefaultModuleManager creates a new default module manager
func NewDefaultModuleManager(logger *observability.Logger) *DefaultModuleManager {
	return &DefaultModuleManager{
		modules: make(map[string]architecture.Module),
		logger:  logger,
	}
}

// RegisterModule registers a module with the manager
func (m *DefaultModuleManager) RegisterModule(moduleID string, module architecture.Module) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.modules[moduleID] = module
	m.logger.WithComponent("module_manager").Info("module_registered", map[string]interface{}{
		"module_id":   moduleID,
		"module_type": module.Metadata().Name,
	})
}

// GetAvailableModules returns all available modules
func (m *DefaultModuleManager) GetAvailableModules() map[string]architecture.Module {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a copy to prevent external modification
	modules := make(map[string]architecture.Module)
	for id, module := range m.modules {
		modules[id] = module
	}

	return modules
}

// GetModuleByID returns a module by its ID
func (m *DefaultModuleManager) GetModuleByID(moduleID string) (architecture.Module, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	module, exists := m.modules[moduleID]
	return module, exists
}

// GetModulesByType returns all modules of a specific type
func (m *DefaultModuleManager) GetModulesByType(moduleType string) []architecture.Module {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var modules []architecture.Module
	for _, module := range m.modules {
		metadata := module.Metadata()
		// Check if the module has the requested capability
		for _, capability := range metadata.Capabilities {
			if string(capability) == moduleType {
				modules = append(modules, module)
				break
			}
		}
	}

	return modules
}

// GetModuleHealth returns the health status of a module
func (m *DefaultModuleManager) GetModuleHealth(moduleID string) (architecture.ModuleStatus, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	module, exists := m.modules[moduleID]
	if !exists {
		return architecture.ModuleStatusUnhealthy, nil
	}

	// Check if module is running
	if !module.IsRunning() {
		return architecture.ModuleStatusUnhealthy, nil
	}

	// For database classification module, perform a health check
	if dbModule, ok := module.(*database_classification.DatabaseClassificationModule); ok {
		health := dbModule.Health()
		return health.Status, nil
	}

	// Default to healthy if no specific health check is available
	return architecture.ModuleStatusHealthy, nil
}

// UnregisterModule removes a module from the manager
func (m *DefaultModuleManager) UnregisterModule(moduleID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if module, exists := m.modules[moduleID]; exists {
		// Stop the module if it's running
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := module.Stop(ctx); err != nil {
			m.logger.WithComponent("module_manager").Warn("failed_to_stop_module", map[string]interface{}{
				"module_id": moduleID,
				"error":     err.Error(),
			})
		}

		delete(m.modules, moduleID)
		m.logger.WithComponent("module_manager").Info("module_unregistered", map[string]interface{}{
			"module_id": moduleID,
		})
	}
}

// GetModuleCount returns the number of registered modules
func (m *DefaultModuleManager) GetModuleCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.modules)
}

// GetModuleIDs returns all registered module IDs
func (m *DefaultModuleManager) GetModuleIDs() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	ids := make([]string, 0, len(m.modules))
	for id := range m.modules {
		ids = append(ids, id)
	}

	return ids
}

// Shutdown gracefully shuts down all modules
func (m *DefaultModuleManager) Shutdown(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var errors []error

	for moduleID, module := range m.modules {
		if err := module.Stop(ctx); err != nil {
			m.logger.WithComponent("module_manager").Warn("failed_to_stop_module", map[string]interface{}{
				"module_id": moduleID,
				"error":     err.Error(),
			})
			errors = append(errors, err)
		}
	}

	// Clear all modules
	m.modules = make(map[string]architecture.Module)

	if len(errors) > 0 {
		return errors[0] // Return first error
	}

	return nil
}
