package architecture

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// LifecycleState represents the current state of a module's lifecycle
type LifecycleState string

const (
	LifecycleStateInitialized LifecycleState = "initialized"
	LifecycleStateStarting    LifecycleState = "starting"
	LifecycleStateRunning     LifecycleState = "running"
	LifecycleStateStopping    LifecycleState = "stopping"
	LifecycleStateStopped     LifecycleState = "stopped"
	LifecycleStateFailed      LifecycleState = "failed"
	LifecycleStateDegraded    LifecycleState = "degraded"
)

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	ModuleID  string                 `json:"module_id"`
	Status    ModuleStatus           `json:"status"`
	Healthy   bool                   `json:"healthy"`
	Latency   time.Duration          `json:"latency"`
	Error     string                 `json:"error,omitempty"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics"`
}

// LifecycleConfig contains configuration for lifecycle management
type LifecycleConfig struct {
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	HealthCheckTimeout  time.Duration `json:"health_check_timeout"`
	StartupTimeout      time.Duration `json:"startup_timeout"`
	ShutdownTimeout     time.Duration `json:"shutdown_timeout"`
	MaxRetries          int           `json:"max_retries"`
	RetryDelay          time.Duration `json:"retry_delay"`
	AutoRestart         bool          `json:"auto_restart"`
	AutoRestartDelay    time.Duration `json:"auto_restart_delay"`
}

// LifecycleManager manages the lifecycle of modules
type LifecycleManager struct {
	moduleManager *ModuleManager
	config        LifecycleConfig
	states        map[string]LifecycleState
	healthResults map[string]*HealthCheckResult
	healthTickers map[string]*time.Ticker
	stopChans     map[string]chan struct{}
	mu            sync.RWMutex
	tracer        trace.Tracer
	meter         metric.Meter
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(moduleManager *ModuleManager, config LifecycleConfig) *LifecycleManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &LifecycleManager{
		moduleManager: moduleManager,
		config:        config,
		states:        make(map[string]LifecycleState),
		healthResults: make(map[string]*HealthCheckResult),
		healthTickers: make(map[string]*time.Ticker),
		stopChans:     make(map[string]chan struct{}),
		tracer:        otel.Tracer("lifecycle-manager"),
		meter:         otel.Meter("lifecycle-manager"),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// StartModule starts a module with enhanced lifecycle management
func (lm *LifecycleManager) StartModule(moduleID string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	ctx, span := lm.tracer.Start(lm.ctx, "StartModule")
	defer span.End()

	// Check if module exists
	module, exists := lm.moduleManager.GetModule(moduleID)
	if !exists {
		return fmt.Errorf("module %s not found", moduleID)
	}

	// Check current state
	currentState := lm.getState(moduleID)
	if currentState == LifecycleStateRunning {
		return fmt.Errorf("module %s is already running", moduleID)
	}

	// Set state to starting
	lm.setState(moduleID, LifecycleStateStarting)

	span.SetAttributes(
		attribute.String("module.id", moduleID),
		attribute.String("lifecycle.state", string(LifecycleStateStarting)),
	)

	// Start module with timeout
	startCtx, cancel := context.WithTimeout(ctx, lm.config.StartupTimeout)
	defer cancel()

	if err := module.Start(startCtx); err != nil {
		lm.setState(moduleID, LifecycleStateFailed)
		span.RecordError(err)
		return fmt.Errorf("failed to start module %s: %w", moduleID, err)
	}

	// Set state to running
	lm.setState(moduleID, LifecycleStateRunning)

	// Start health monitoring
	lm.startHealthMonitoring(moduleID)

	// Emit lifecycle event
	lm.emitLifecycleEvent(moduleID, "module_started", map[string]interface{}{
		"state": LifecycleStateRunning,
	})

	span.SetAttributes(attribute.String("lifecycle.state", string(LifecycleStateRunning)))

	return nil
}

// StopModule stops a module with graceful shutdown
func (lm *LifecycleManager) StopModule(moduleID string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	ctx, span := lm.tracer.Start(lm.ctx, "StopModule")
	defer span.End()

	// Check if module exists
	module, exists := lm.moduleManager.GetModule(moduleID)
	if !exists {
		return fmt.Errorf("module %s not found", moduleID)
	}

	// Check current state
	currentState := lm.getState(moduleID)
	if currentState == LifecycleStateStopped || currentState == LifecycleStateFailed {
		return fmt.Errorf("module %s is not running", moduleID)
	}

	// Set state to stopping
	lm.setState(moduleID, LifecycleStateStopping)

	span.SetAttributes(
		attribute.String("module.id", moduleID),
		attribute.String("lifecycle.state", string(LifecycleStateStopping)),
	)

	// Stop health monitoring
	lm.stopHealthMonitoring(moduleID)

	// Stop module with timeout
	stopCtx, cancel := context.WithTimeout(ctx, lm.config.ShutdownTimeout)
	defer cancel()

	if err := module.Stop(stopCtx); err != nil {
		lm.setState(moduleID, LifecycleStateFailed)
		span.RecordError(err)
		return fmt.Errorf("failed to stop module %s: %w", moduleID, err)
	}

	// Set state to stopped
	lm.setState(moduleID, LifecycleStateStopped)

	// Emit lifecycle event
	lm.emitLifecycleEvent(moduleID, "module_stopped", map[string]interface{}{
		"state": LifecycleStateStopped,
	})

	span.SetAttributes(attribute.String("lifecycle.state", string(LifecycleStateStopped)))

	return nil
}

// RestartModule restarts a module
func (lm *LifecycleManager) RestartModule(moduleID string) error {
	_, span := lm.tracer.Start(lm.ctx, "RestartModule")
	defer span.End()

	span.SetAttributes(attribute.String("module.id", moduleID))

	// Stop module first
	if err := lm.StopModule(moduleID); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to stop module %s during restart: %w", moduleID, err)
	}

	// Wait for restart delay
	if lm.config.AutoRestartDelay > 0 {
		time.Sleep(lm.config.AutoRestartDelay)
	}

	// Start module
	if err := lm.StartModule(moduleID); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to start module %s during restart: %w", moduleID, err)
	}

	// Emit lifecycle event
	lm.emitLifecycleEvent(moduleID, "module_restarted", map[string]interface{}{
		"state": LifecycleStateRunning,
	})

	return nil
}

// GetModuleState returns the current lifecycle state of a module
func (lm *LifecycleManager) GetModuleState(moduleID string) (LifecycleState, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	state, exists := lm.states[moduleID]
	return state, exists
}

// GetAllModuleStates returns the lifecycle states of all modules
func (lm *LifecycleManager) GetAllModuleStates() map[string]LifecycleState {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	states := make(map[string]LifecycleState)
	for id, state := range lm.states {
		states[id] = state
	}
	return states
}

// GetHealthResult returns the latest health check result for a module
func (lm *LifecycleManager) GetHealthResult(moduleID string) (*HealthCheckResult, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	result, exists := lm.healthResults[moduleID]
	return result, exists
}

// GetAllHealthResults returns health check results for all modules
func (lm *LifecycleManager) GetAllHealthResults() map[string]*HealthCheckResult {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	results := make(map[string]*HealthCheckResult)
	for id, result := range lm.healthResults {
		results[id] = result
	}
	return results
}

// PerformHealthCheck performs a health check on a specific module
func (lm *LifecycleManager) PerformHealthCheck(moduleID string) (*HealthCheckResult, error) {
	ctx, span := lm.tracer.Start(lm.ctx, "PerformHealthCheck")
	defer span.End()

	span.SetAttributes(attribute.String("module.id", moduleID))

	// Check if module exists
	module, exists := lm.moduleManager.GetModule(moduleID)
	if !exists {
		return nil, fmt.Errorf("module %s not found", moduleID)
	}

	// Create health check context with timeout
	healthCtx, cancel := context.WithTimeout(ctx, lm.config.HealthCheckTimeout)
	defer cancel()

	// Perform health check
	start := time.Now()
	err := module.HealthCheck(healthCtx)
	latency := time.Since(start)

	// Create health check result
	result := &HealthCheckResult{
		ModuleID:  moduleID,
		Latency:   latency,
		Timestamp: time.Now(),
		Metrics:   make(map[string]interface{}),
	}

	if err != nil {
		result.Status = ModuleStatusUnhealthy
		result.Healthy = false
		result.Error = err.Error()
		result.Message = "Health check failed"

		// Update module health status
		lm.updateModuleHealth(moduleID, result)

		span.RecordError(err)
		span.SetAttributes(
			attribute.Bool("health.healthy", false),
			attribute.String("health.error", err.Error()),
		)
	} else {
		result.Status = ModuleStatusHealthy
		result.Healthy = true
		result.Message = "Health check passed"

		// Update module health status
		lm.updateModuleHealth(moduleID, result)

		span.SetAttributes(attribute.Bool("health.healthy", true))
	}

	// Store health result
	lm.mu.Lock()
	lm.healthResults[moduleID] = result
	lm.mu.Unlock()

	// Emit health check event
	lm.emitLifecycleEvent(moduleID, "health_check_completed", map[string]interface{}{
		"healthy": result.Healthy,
		"latency": result.Latency.String(),
		"error":   result.Error,
	})

	return result, nil
}

// startHealthMonitoring starts periodic health monitoring for a module
func (lm *LifecycleManager) startHealthMonitoring(moduleID string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Stop existing monitoring if any
	lm.stopHealthMonitoring(moduleID)

	// Create stop channel
	stopChan := make(chan struct{})
	lm.stopChans[moduleID] = stopChan

	// Create ticker for health checks
	ticker := time.NewTicker(lm.config.HealthCheckInterval)
	lm.healthTickers[moduleID] = ticker

	// Start health monitoring goroutine
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Perform health check
				result, err := lm.PerformHealthCheck(moduleID)
				if err != nil {
					// Log error but continue monitoring
					continue
				}

				// Check if module needs restart
				if !result.Healthy && lm.config.AutoRestart {
					// Check if module is in failed state
					state, _ := lm.GetModuleState(moduleID)
					if state == LifecycleStateFailed {
						// Attempt auto-restart
						go func() {
							if err := lm.RestartModule(moduleID); err != nil {
								// Log restart failure
								lm.emitLifecycleEvent(moduleID, "auto_restart_failed", map[string]interface{}{
									"error": err.Error(),
								})
							}
						}()
					}
				}

			case <-stopChan:
				return
			case <-lm.ctx.Done():
				return
			}
		}
	}()
}

// stopHealthMonitoring stops health monitoring for a module
func (lm *LifecycleManager) stopHealthMonitoring(moduleID string) {
	// Stop ticker
	if ticker, exists := lm.healthTickers[moduleID]; exists {
		ticker.Stop()
		delete(lm.healthTickers, moduleID)
	}

	// Stop goroutine
	if stopChan, exists := lm.stopChans[moduleID]; exists {
		close(stopChan)
		delete(lm.stopChans, moduleID)
	}
}

// setState sets the lifecycle state of a module
func (lm *LifecycleManager) setState(moduleID string, state LifecycleState) {
	lm.states[moduleID] = state

	// Emit state change event
	lm.emitLifecycleEvent(moduleID, "state_changed", map[string]interface{}{
		"previous_state": lm.getState(moduleID),
		"new_state":      state,
	})
}

// getState gets the lifecycle state of a module
func (lm *LifecycleManager) getState(moduleID string) LifecycleState {
	if state, exists := lm.states[moduleID]; exists {
		return state
	}
	return LifecycleStateInitialized
}

// updateModuleHealth updates the module's health status in the module manager
func (lm *LifecycleManager) updateModuleHealth(moduleID string, result *HealthCheckResult) {
	// This would update the module manager's health tracking
	// For now, we'll just emit an event
	lm.emitLifecycleEvent(moduleID, "health_updated", map[string]interface{}{
		"healthy": result.Healthy,
		"latency": result.Latency.String(),
		"status":  result.Status,
	})
}

// emitLifecycleEvent emits a lifecycle event
func (lm *LifecycleManager) emitLifecycleEvent(moduleID, eventType string, data map[string]interface{}) {
	event := ModuleEvent{
		Type:      eventType,
		ModuleID:  moduleID,
		Timestamp: time.Now(),
		Data:      data,
	}

	// Emit event through module manager
	lm.moduleManager.emitEvent(event)
}

// StartAllModules starts all modules with lifecycle management
func (lm *LifecycleManager) StartAllModules() error {
	_, span := lm.tracer.Start(lm.ctx, "StartAllModules")
	defer span.End()

	modules := lm.moduleManager.ListModules()

	for _, module := range modules {
		moduleID := module.ID()

		if err := lm.StartModule(moduleID); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to start module %s: %w", moduleID, err)
		}
	}

	return nil
}

// StopAllModules stops all modules with lifecycle management
func (lm *LifecycleManager) StopAllModules() error {
	_, span := lm.tracer.Start(lm.ctx, "StopAllModules")
	defer span.End()

	modules := lm.moduleManager.ListModules()

	for _, module := range modules {
		moduleID := module.ID()

		if err := lm.StopModule(moduleID); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to stop module %s: %w", moduleID, err)
		}
	}

	return nil
}

// Close shuts down the lifecycle manager
func (lm *LifecycleManager) Close() error {
	lm.cancel()

	// Stop all health monitoring
	lm.mu.Lock()
	defer lm.mu.Unlock()

	for moduleID := range lm.healthTickers {
		lm.stopHealthMonitoring(moduleID)
	}

	// Stop all modules
	if err := lm.StopAllModules(); err != nil {
		return fmt.Errorf("failed to stop all modules: %w", err)
	}

	return nil
}
