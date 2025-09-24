package architecture

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ModuleStatus represents the current status of a module
type ModuleStatus string

const (
	ModuleStatusStarting  ModuleStatus = "starting"
	ModuleStatusRunning   ModuleStatus = "running"
	ModuleStatusStopped   ModuleStatus = "stopped"
	ModuleStatusFailed    ModuleStatus = "failed"
	ModuleStatusHealthy   ModuleStatus = "healthy"
	ModuleStatusUnhealthy ModuleStatus = "unhealthy"
)

// ModuleCapability represents what a module can do
type ModuleCapability string

const (
	CapabilityClassification ModuleCapability = "classification"
	CapabilityVerification   ModuleCapability = "verification"
	CapabilityDataExtraction ModuleCapability = "data_extraction"
	CapabilityRiskAssessment ModuleCapability = "risk_assessment"
	CapabilityWebAnalysis    ModuleCapability = "web_analysis"
	CapabilityMLPrediction   ModuleCapability = "ml_prediction"
)

// ModulePriority represents the priority level of a module
type ModulePriority int

const (
	PriorityLow      ModulePriority = 1
	PriorityMedium   ModulePriority = 2
	PriorityHigh     ModulePriority = 3
	PriorityCritical ModulePriority = 4
)

// ModuleMetadata contains information about a module
type ModuleMetadata struct {
	Name         string             `json:"name"`
	Version      string             `json:"version"`
	Description  string             `json:"description"`
	Capabilities []ModuleCapability `json:"capabilities"`
	Priority     ModulePriority     `json:"priority"`
	Author       string             `json:"author"`
	Tags         []string           `json:"tags"`
}

// ModuleHealth contains health information about a module
type ModuleHealth struct {
	Status      ModuleStatus  `json:"status"`
	LastCheck   time.Time     `json:"last_check"`
	ErrorCount  int64         `json:"error_count"`
	SuccessRate float64       `json:"success_rate"`
	Latency     time.Duration `json:"latency"`
	Message     string        `json:"message"`
}

// ModuleConfig contains configuration for a module
type ModuleConfig struct {
	Enabled      bool                   `json:"enabled"`
	Timeout      time.Duration          `json:"timeout"`
	RetryCount   int                    `json:"retry_count"`
	Parameters   map[string]interface{} `json:"parameters"`
	Dependencies []string               `json:"dependencies"`
}

// ModuleRequest represents a request to a module
type ModuleRequest struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Data     map[string]interface{} `json:"data"`
	Priority ModulePriority         `json:"priority"`
	Timeout  time.Duration          `json:"timeout"`
	Context  context.Context        `json:"-"`
}

// ModuleResponse represents a response from a module
type ModuleResponse struct {
	ID         string                 `json:"id"`
	Success    bool                   `json:"success"`
	Data       map[string]interface{} `json:"data"`
	Error      string                 `json:"error,omitempty"`
	Confidence float64                `json:"confidence"`
	Latency    time.Duration          `json:"latency"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ModuleEvent represents an event from a module
type ModuleEvent struct {
	Type      string                 `json:"type"`
	ModuleID  string                 `json:"module_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Module interface that all modules must implement
type Module interface {
	// Core module methods
	ID() string
	Metadata() ModuleMetadata
	Config() ModuleConfig
	Health() ModuleHealth

	// Lifecycle methods
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// Processing methods
	Process(ctx context.Context, req ModuleRequest) (ModuleResponse, error)
	CanHandle(req ModuleRequest) bool

	// Health check
	HealthCheck(ctx context.Context) error

	// Event handling
	OnEvent(event ModuleEvent) error
}

// ModuleManager manages all modules in the system
type ModuleManager struct {
	modules map[string]Module
	configs map[string]ModuleConfig
	health  map[string]ModuleHealth
	events  chan ModuleEvent
	mu      sync.RWMutex
	tracer  trace.Tracer
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewModuleManager creates a new module manager
func NewModuleManager() *ModuleManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &ModuleManager{
		modules: make(map[string]Module),
		configs: make(map[string]ModuleConfig),
		health:  make(map[string]ModuleHealth),
		events:  make(chan ModuleEvent, 100),
		tracer:  otel.Tracer("module-manager"),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// RegisterModule registers a new module with the manager
func (mm *ModuleManager) RegisterModule(module Module, config ModuleConfig) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	_, span := mm.tracer.Start(mm.ctx, "RegisterModule")
	defer span.End()

	moduleID := module.ID()

	// Check if module already exists
	if _, exists := mm.modules[moduleID]; exists {
		return fmt.Errorf("module %s already registered", moduleID)
	}

	// Validate module metadata
	metadata := module.Metadata()
	if metadata.Name == "" {
		return fmt.Errorf("module %s has empty name", moduleID)
	}

	// Validate dependencies
	for _, dep := range config.Dependencies {
		if _, exists := mm.modules[dep]; !exists {
			return fmt.Errorf("module %s depends on unregistered module %s", moduleID, dep)
		}
	}

	// Register module
	mm.modules[moduleID] = module
	mm.configs[moduleID] = config
	mm.health[moduleID] = ModuleHealth{
		Status:    ModuleStatusStopped,
		LastCheck: time.Now(),
	}

	span.SetAttributes(
		attribute.String("module.id", moduleID),
		attribute.String("module.name", metadata.Name),
		attribute.String("module.version", metadata.Version),
	)

	// Emit registration event
	mm.emitEvent(ModuleEvent{
		Type:      "module_registered",
		ModuleID:  moduleID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"name":         metadata.Name,
			"version":      metadata.Version,
			"capabilities": metadata.Capabilities,
		},
	})

	return nil
}

// UnregisterModule removes a module from the manager
func (mm *ModuleManager) UnregisterModule(moduleID string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	_, span := mm.tracer.Start(mm.ctx, "UnregisterModule")
	defer span.End()

	// Check if module exists
	module, exists := mm.modules[moduleID]
	if !exists {
		return fmt.Errorf("module %s not found", moduleID)
	}

	// Stop module if running
	if module.IsRunning() {
		if err := module.Stop(mm.ctx); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to stop module %s: %w", moduleID, err)
		}
	}

	// Remove module
	delete(mm.modules, moduleID)
	delete(mm.configs, moduleID)
	delete(mm.health, moduleID)

	span.SetAttributes(attribute.String("module.id", moduleID))

	// Emit unregistration event
	mm.emitEvent(ModuleEvent{
		Type:      "module_unregistered",
		ModuleID:  moduleID,
		Timestamp: time.Now(),
	})

	return nil
}

// GetModule retrieves a module by ID
func (mm *ModuleManager) GetModule(moduleID string) (Module, bool) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	module, exists := mm.modules[moduleID]
	return module, exists
}

// ListModules returns all registered modules
func (mm *ModuleManager) ListModules() []Module {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	modules := make([]Module, 0, len(mm.modules))
	for _, module := range mm.modules {
		modules = append(modules, module)
	}
	return modules
}

// GetModulesByCapability returns modules that have a specific capability
func (mm *ModuleManager) GetModulesByCapability(capability ModuleCapability) []Module {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var modules []Module
	for _, module := range mm.modules {
		metadata := module.Metadata()
		for _, cap := range metadata.Capabilities {
			if cap == capability {
				modules = append(modules, module)
				break
			}
		}
	}
	return modules
}

// StartModule starts a specific module
func (mm *ModuleManager) StartModule(moduleID string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	ctx, span := mm.tracer.Start(mm.ctx, "StartModule")
	defer span.End()

	module, exists := mm.modules[moduleID]
	if !exists {
		return fmt.Errorf("module %s not found", moduleID)
	}

	config := mm.configs[moduleID]
	if !config.Enabled {
		return fmt.Errorf("module %s is disabled", moduleID)
	}

	// Check dependencies
	for _, dep := range config.Dependencies {
		depModule, exists := mm.modules[dep]
		if !exists {
			return fmt.Errorf("dependency module %s not found", dep)
		}
		if !depModule.IsRunning() {
			return fmt.Errorf("dependency module %s is not running", dep)
		}
	}

	// Start module
	if err := module.Start(ctx); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to start module %s: %w", moduleID, err)
	}

	// Update health status
	mm.health[moduleID] = ModuleHealth{
		Status:    ModuleStatusRunning,
		LastCheck: time.Now(),
	}

	span.SetAttributes(attribute.String("module.id", moduleID))

	// Emit start event
	mm.emitEvent(ModuleEvent{
		Type:      "module_started",
		ModuleID:  moduleID,
		Timestamp: time.Now(),
	})

	return nil
}

// StopModule stops a specific module
func (mm *ModuleManager) StopModule(moduleID string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	ctx, span := mm.tracer.Start(mm.ctx, "StopModule")
	defer span.End()

	module, exists := mm.modules[moduleID]
	if !exists {
		return fmt.Errorf("module %s not found", moduleID)
	}

	// Stop module
	if err := module.Stop(ctx); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to stop module %s: %w", moduleID, err)
	}

	// Update health status
	mm.health[moduleID] = ModuleHealth{
		Status:    ModuleStatusStopped,
		LastCheck: time.Now(),
	}

	span.SetAttributes(attribute.String("module.id", moduleID))

	// Emit stop event
	mm.emitEvent(ModuleEvent{
		Type:      "module_stopped",
		ModuleID:  moduleID,
		Timestamp: time.Now(),
	})

	return nil
}

// StartAllModules starts all enabled modules
func (mm *ModuleManager) StartAllModules() error {
	mm.mu.RLock()
	modules := make(map[string]Module)
	configs := make(map[string]ModuleConfig)
	for id, module := range mm.modules {
		modules[id] = module
		configs[id] = mm.configs[id]
	}
	mm.mu.RUnlock()

	_, span := mm.tracer.Start(mm.ctx, "StartAllModules")
	defer span.End()

	// Start modules in dependency order
	started := make(map[string]bool)

	for len(started) < len(modules) {
		progress := false

		for moduleID := range modules {
			if started[moduleID] {
				continue
			}

			config := configs[moduleID]
			if !config.Enabled {
				started[moduleID] = true
				progress = true
				continue
			}

			// Check if all dependencies are started
			depsReady := true
			for _, dep := range config.Dependencies {
				if !started[dep] {
					depsReady = false
					break
				}
			}

			if depsReady {
				if err := mm.StartModule(moduleID); err != nil {
					span.RecordError(err)
					return fmt.Errorf("failed to start module %s: %w", moduleID, err)
				}
				started[moduleID] = true
				progress = true
			}
		}

		if !progress {
			return fmt.Errorf("circular dependency detected in modules")
		}
	}

	return nil
}

// StopAllModules stops all modules
func (mm *ModuleManager) StopAllModules() error {
	mm.mu.RLock()
	modules := make(map[string]Module)
	for id, module := range mm.modules {
		modules[id] = module
	}
	mm.mu.RUnlock()

	_, span := mm.tracer.Start(mm.ctx, "StopAllModules")
	defer span.End()

	for moduleID := range modules {
		if err := mm.StopModule(moduleID); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to stop module %s: %w", moduleID, err)
		}
	}

	return nil
}

// GetModuleHealth returns the health status of a module
func (mm *ModuleManager) GetModuleHealth(moduleID string) (ModuleHealth, bool) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	health, exists := mm.health[moduleID]
	return health, exists
}

// GetAllModuleHealth returns health status of all modules
func (mm *ModuleManager) GetAllModuleHealth() map[string]ModuleHealth {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	health := make(map[string]ModuleHealth)
	for id, h := range mm.health {
		health[id] = h
	}
	return health
}

// UpdateModuleConfig updates the configuration of a module
func (mm *ModuleManager) UpdateModuleConfig(moduleID string, config ModuleConfig) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if _, exists := mm.modules[moduleID]; !exists {
		return fmt.Errorf("module %s not found", moduleID)
	}

	mm.configs[moduleID] = config

	// Emit config update event
	mm.emitEvent(ModuleEvent{
		Type:      "module_config_updated",
		ModuleID:  moduleID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"enabled": config.Enabled,
		},
	})

	return nil
}

// emitEvent emits an event to the event channel
func (mm *ModuleManager) emitEvent(event ModuleEvent) {
	select {
	case mm.events <- event:
	default:
		// Channel is full, drop event
	}
}

// GetEvents returns the event channel
func (mm *ModuleManager) GetEvents() <-chan ModuleEvent {
	return mm.events
}

// Close shuts down the module manager
func (mm *ModuleManager) Close() error {
	mm.cancel()

	// Stop all modules
	if err := mm.StopAllModules(); err != nil {
		return fmt.Errorf("failed to stop all modules: %w", err)
	}

	// Close event channel
	mm.mu.Lock()
	defer mm.mu.Unlock()
	close(mm.events)

	return nil
}
