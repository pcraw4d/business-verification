package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/shared"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ModuleRegistry provides centralized management of all intelligent routing modules
type ModuleRegistry struct {
	// Thread-safe module storage
	modules   map[string]shared.ClassificationModule
	modulesMu sync.RWMutex

	// Module metadata and capabilities
	capabilities   map[string]*ModuleCapability
	capabilitiesMu sync.RWMutex

	// Performance tracking
	performance   map[string]*ModulePerformance
	performanceMu sync.RWMutex

	// Health status tracking
	healthStatus   map[string]*ModuleHealth
	healthStatusMu sync.RWMutex

	// Registry configuration
	config *RegistryConfig

	// Observability
	logger  *observability.Logger
	metrics *observability.Metrics
	tracer  trace.Tracer

	// Control channels
	stopChan chan struct{}
}

// RegistryConfig holds configuration for the module registry
type RegistryConfig struct {
	// Health check settings
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
	MaxHealthFailures   int

	// Performance tracking settings
	PerformanceWindowSize time.Duration
	MaxPerformanceHistory int

	// Module discovery settings
	AutoDiscoveryEnabled bool
	DiscoveryInterval    time.Duration

	// Registry limits
	MaxModules                int
	MaxConcurrentHealthChecks int
}

// ModuleCapability represents a module's capabilities and requirements
type ModuleCapability struct {
	ModuleID          string                 `json:"module_id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	Version           string                 `json:"version"`
	SupportedFeatures []string               `json:"supported_features"`
	RequiredInputs    []string               `json:"required_inputs"`
	OutputFormats     []string               `json:"output_formats"`
	PerformanceClass  string                 `json:"performance_class"` // fast, medium, slow
	ResourceUsage     map[string]interface{} `json:"resource_usage"`
	Metadata          map[string]interface{} `json:"metadata"`
	RegisteredAt      time.Time              `json:"registered_at"`
	LastUpdated       time.Time              `json:"last_updated"`
}

// ModulePerformance tracks performance metrics for a module
type ModulePerformance struct {
	ModuleID           string        `json:"module_id"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency"`
	MinLatency         time.Duration `json:"min_latency"`
	MaxLatency         time.Duration `json:"max_latency"`
	LastRequestTime    time.Time     `json:"last_request_time"`
	LastSuccessTime    time.Time     `json:"last_success_time"`
	LastFailureTime    time.Time     `json:"last_failure_time"`
	ErrorRate          float64       `json:"error_rate"`
	SuccessRate        float64       `json:"success_rate"`
	Throughput         float64       `json:"throughput"` // requests per second
	LastCalculated     time.Time     `json:"last_calculated"`
}

// ModuleHealth represents the health status of a module
type ModuleHealth struct {
	ModuleID        string        `json:"module_id"`
	Status          string        `json:"status"` // healthy, degraded, unhealthy, unknown
	LastCheckTime   time.Time     `json:"last_check_time"`
	LastSuccessTime time.Time     `json:"last_success_time"`
	FailureCount    int           `json:"failure_count"`
	ErrorMessage    string        `json:"error_message,omitempty"`
	ResponseTime    time.Duration `json:"response_time"`
	IsAvailable     bool          `json:"is_available"`
}

// NewModuleRegistry creates a new module registry
func NewModuleRegistry(
	config *RegistryConfig,
	logger *observability.Logger,
	metrics *observability.Metrics,
	tracer trace.Tracer,
) *ModuleRegistry {
	// Set default configuration
	if config == nil {
		config = &RegistryConfig{
			HealthCheckInterval:       30 * time.Second,
			HealthCheckTimeout:        10 * time.Second,
			MaxHealthFailures:         3,
			PerformanceWindowSize:     5 * time.Minute,
			MaxPerformanceHistory:     100,
			AutoDiscoveryEnabled:      false,
			DiscoveryInterval:         1 * time.Minute,
			MaxModules:                100,
			MaxConcurrentHealthChecks: 10,
		}
	}

	registry := &ModuleRegistry{
		modules:      make(map[string]shared.ClassificationModule),
		capabilities: make(map[string]*ModuleCapability),
		performance:  make(map[string]*ModulePerformance),
		healthStatus: make(map[string]*ModuleHealth),
		config:       config,
		logger:       logger,
		metrics:      metrics,
		tracer:       tracer,
		stopChan:     make(chan struct{}),
	}

	// Start background health checking
	go registry.healthCheckWorker()

	// Start performance calculation worker
	go registry.performanceCalculationWorker()

	return registry
}

// RegisterModule registers a new module with the registry
func (r *ModuleRegistry) RegisterModule(
	ctx context.Context,
	module shared.ClassificationModule,
	capability *ModuleCapability,
) error {
	ctx, span := r.tracer.Start(ctx, "ModuleRegistry.RegisterModule")
	defer span.End()

	moduleID := module.ID()
	span.SetAttributes(attribute.String("module_id", moduleID))

	// Validate module
	if module == nil {
		return fmt.Errorf("module cannot be nil")
	}

	if moduleID == "" {
		return fmt.Errorf("module ID cannot be empty")
	}

	// Check registry limits
	r.modulesMu.RLock()
	if len(r.modules) >= r.config.MaxModules {
		r.modulesMu.RUnlock()
		return fmt.Errorf("registry is at maximum capacity (%d modules)", r.config.MaxModules)
	}
	r.modulesMu.RUnlock()

	// Register module
	r.modulesMu.Lock()
	if _, exists := r.modules[moduleID]; exists {
		r.modulesMu.Unlock()
		return fmt.Errorf("module %s is already registered", moduleID)
	}
	r.modules[moduleID] = module
	r.modulesMu.Unlock()

	// Register capability
	if capability != nil {
		capability.ModuleID = moduleID
		capability.RegisteredAt = time.Now()
		capability.LastUpdated = time.Now()

		r.capabilitiesMu.Lock()
		r.capabilities[moduleID] = capability
		r.capabilitiesMu.Unlock()
	}

	// Initialize performance tracking
	r.performanceMu.Lock()
	r.performance[moduleID] = &ModulePerformance{
		ModuleID:       moduleID,
		LastCalculated: time.Now(),
	}
	r.performanceMu.Unlock()

	// Initialize health status
	r.healthStatusMu.Lock()
	r.healthStatus[moduleID] = &ModuleHealth{
		ModuleID:      moduleID,
		Status:        "unknown",
		LastCheckTime: time.Now(),
		IsAvailable:   false,
	}
	r.healthStatusMu.Unlock()

	r.logger.Info("module registered successfully", map[string]interface{}{
		"module_id": moduleID,
		"name":      capability.Name,
		"version":   capability.Version,
	})

	return nil
}

// UnregisterModule removes a module from the registry
func (r *ModuleRegistry) UnregisterModule(ctx context.Context, moduleID string) error {
	ctx, span := r.tracer.Start(ctx, "ModuleRegistry.UnregisterModule")
	defer span.End()

	span.SetAttributes(attribute.String("module_id", moduleID))

	// Remove module
	r.modulesMu.Lock()
	if _, exists := r.modules[moduleID]; !exists {
		r.modulesMu.Unlock()
		return fmt.Errorf("module %s is not registered", moduleID)
	}
	delete(r.modules, moduleID)
	r.modulesMu.Unlock()

	// Remove capability
	r.capabilitiesMu.Lock()
	delete(r.capabilities, moduleID)
	r.capabilitiesMu.Unlock()

	// Remove performance tracking
	r.performanceMu.Lock()
	delete(r.performance, moduleID)
	r.performanceMu.Unlock()

	// Remove health status
	r.healthStatusMu.Lock()
	delete(r.healthStatus, moduleID)
	r.healthStatusMu.Unlock()

	r.logger.Info("module unregistered successfully", map[string]interface{}{
		"module_id": moduleID,
	})

	return nil
}

// GetModule retrieves a module by ID
func (r *ModuleRegistry) GetModule(ctx context.Context, moduleID string) (shared.ClassificationModule, error) {
	ctx, span := r.tracer.Start(ctx, "ModuleRegistry.GetModule")
	defer span.End()

	span.SetAttributes(attribute.String("module_id", moduleID))

	r.modulesMu.RLock()
	module, exists := r.modules[moduleID]
	r.modulesMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("module %s not found", moduleID)
	}

	return module, nil
}

// ListModules returns all registered modules
func (r *ModuleRegistry) ListModules(ctx context.Context) ([]shared.ClassificationModule, error) {
	ctx, span := r.tracer.Start(ctx, "ModuleRegistry.ListModules")
	defer span.End()

	r.modulesMu.RLock()
	defer r.modulesMu.RUnlock()

	modules := make([]shared.ClassificationModule, 0, len(r.modules))
	for _, module := range r.modules {
		modules = append(modules, module)
	}

	span.SetAttributes(attribute.Int("module_count", len(modules)))
	return modules, nil
}

// GetModuleCapability returns the capability information for a module
func (r *ModuleRegistry) GetModuleCapability(ctx context.Context, moduleID string) (*ModuleCapability, error) {
	ctx, span := r.tracer.Start(ctx, "ModuleRegistry.GetModuleCapability")
	defer span.End()

	span.SetAttributes(attribute.String("module_id", moduleID))

	r.capabilitiesMu.RLock()
	capability, exists := r.capabilities[moduleID]
	r.capabilitiesMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("capability for module %s not found", moduleID)
	}

	return capability, nil
}

// GetModulePerformance returns performance metrics for a module
func (r *ModuleRegistry) GetModulePerformance(ctx context.Context, moduleID string) (*ModulePerformance, error) {
	ctx, span := r.tracer.Start(ctx, "ModuleRegistry.GetModulePerformance")
	defer span.End()

	span.SetAttributes(attribute.String("module_id", moduleID))

	r.performanceMu.RLock()
	performance, exists := r.performance[moduleID]
	r.performanceMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("performance data for module %s not found", moduleID)
	}

	return performance, nil
}

// GetModuleHealth returns health status for a module
func (r *ModuleRegistry) GetModuleHealth(ctx context.Context, moduleID string) (*ModuleHealth, error) {
	ctx, span := r.tracer.Start(ctx, "ModuleRegistry.GetModuleHealth")
	defer span.End()

	span.SetAttributes(attribute.String("module_id", moduleID))

	r.healthStatusMu.RLock()
	health, exists := r.healthStatus[moduleID]
	r.healthStatusMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("health status for module %s not found", moduleID)
	}

	return health, nil
}

// FindModulesByCapability finds modules that support specific features
func (r *ModuleRegistry) FindModulesByCapability(ctx context.Context, features []string) ([]shared.ClassificationModule, error) {
	ctx, span := r.tracer.Start(ctx, "ModuleRegistry.FindModulesByCapability")
	defer span.End()

	span.SetAttributes(attribute.StringSlice("features", features))

	r.capabilitiesMu.RLock()
	defer r.capabilitiesMu.RUnlock()

	var matchingModules []shared.ClassificationModule

	for moduleID, capability := range r.capabilities {
		// Check if module supports all requested features
		supportsAll := true
		for _, feature := range features {
			found := false
			for _, supportedFeature := range capability.SupportedFeatures {
				if supportedFeature == feature {
					found = true
					break
				}
			}
			if !found {
				supportsAll = false
				break
			}
		}

		if supportsAll {
			r.modulesMu.RLock()
			if module, exists := r.modules[moduleID]; exists {
				matchingModules = append(matchingModules, module)
			}
			r.modulesMu.RUnlock()
		}
	}

	span.SetAttributes(attribute.Int("matching_modules", len(matchingModules)))
	return matchingModules, nil
}

// GetHealthyModules returns only modules that are currently healthy
func (r *ModuleRegistry) GetHealthyModules(ctx context.Context) ([]shared.ClassificationModule, error) {
	ctx, span := r.tracer.Start(ctx, "ModuleRegistry.GetHealthyModules")
	defer span.End()

	r.healthStatusMu.RLock()
	defer r.healthStatusMu.RUnlock()

	var healthyModules []shared.ClassificationModule

	for moduleID, health := range r.healthStatus {
		if health.Status == "healthy" && health.IsAvailable {
			r.modulesMu.RLock()
			if module, exists := r.modules[moduleID]; exists {
				healthyModules = append(healthyModules, module)
			}
			r.modulesMu.RUnlock()
		}
	}

	span.SetAttributes(attribute.Int("healthy_modules", len(healthyModules)))
	return healthyModules, nil
}

// RecordModuleRequest records a request to a module for performance tracking
func (r *ModuleRegistry) RecordModuleRequest(
	ctx context.Context,
	moduleID string,
	success bool,
	latency time.Duration,
) error {
	ctx, span := r.tracer.Start(ctx, "ModuleRegistry.RecordModuleRequest")
	defer span.End()

	span.SetAttributes(
		attribute.String("module_id", moduleID),
		attribute.Bool("success", success),
		attribute.Int64("latency_ms", latency.Milliseconds()),
	)

	r.performanceMu.Lock()
	defer r.performanceMu.Unlock()

	performance, exists := r.performance[moduleID]
	if !exists {
		return fmt.Errorf("performance tracking not found for module %s", moduleID)
	}

	// Update performance metrics
	performance.TotalRequests++
	if success {
		performance.SuccessfulRequests++
		performance.LastSuccessTime = time.Now()
	} else {
		performance.FailedRequests++
		performance.LastFailureTime = time.Now()
	}

	performance.LastRequestTime = time.Now()

	// Update latency metrics
	if performance.MinLatency == 0 || latency < performance.MinLatency {
		performance.MinLatency = latency
	}
	if latency > performance.MaxLatency {
		performance.MaxLatency = latency
	}

	// Calculate average latency (simple moving average)
	if performance.TotalRequests == 1 {
		performance.AverageLatency = latency
	} else {
		totalLatency := performance.AverageLatency * time.Duration(performance.TotalRequests-1)
		performance.AverageLatency = (totalLatency + latency) / time.Duration(performance.TotalRequests)
	}

	return nil
}

// healthCheckWorker performs periodic health checks on all modules
func (r *ModuleRegistry) healthCheckWorker() {
	ticker := time.NewTicker(r.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.performHealthChecks()
		case <-r.stopChan:
			return
		}
	}
}

// performHealthChecks performs health checks on all registered modules
func (r *ModuleRegistry) performHealthChecks() {
	ctx, span := r.tracer.Start(context.Background(), "ModuleRegistry.performHealthChecks")
	defer span.End()

	r.modulesMu.RLock()
	modules := make(map[string]shared.ClassificationModule)
	for id, module := range r.modules {
		modules[id] = module
	}
	r.modulesMu.RUnlock()

	// Use semaphore to limit concurrent health checks
	semaphore := make(chan struct{}, r.config.MaxConcurrentHealthChecks)
	var wg sync.WaitGroup

	for moduleID, module := range modules {
		wg.Add(1)
		go func(id string, mod shared.ClassificationModule) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			r.checkModuleHealth(ctx, id, mod)
		}(moduleID, module)
	}

	wg.Wait()
}

// checkModuleHealth performs a health check on a single module
func (r *ModuleRegistry) checkModuleHealth(ctx context.Context, moduleID string, module shared.ClassificationModule) {
	ctx, span := r.tracer.Start(ctx, "ModuleRegistry.checkModuleHealth")
	defer span.End()

	span.SetAttributes(attribute.String("module_id", moduleID))

	start := time.Now()
	err := module.HealthCheck(ctx)
	duration := time.Since(start)

	r.healthStatusMu.Lock()
	health, exists := r.healthStatus[moduleID]
	if !exists {
		health = &ModuleHealth{
			ModuleID: moduleID,
		}
		r.healthStatus[moduleID] = health
	}

	health.LastCheckTime = time.Now()
	health.ResponseTime = duration

	if err != nil {
		health.FailureCount++
		health.ErrorMessage = err.Error()
		health.IsAvailable = false

		if health.FailureCount >= r.config.MaxHealthFailures {
			health.Status = "unhealthy"
		} else {
			health.Status = "degraded"
		}

		r.logger.Warn("module health check failed", map[string]interface{}{
			"module_id":     moduleID,
			"error":         err.Error(),
			"failure_count": health.FailureCount,
			"response_time": duration.String(),
		})
	} else {
		health.FailureCount = 0
		health.ErrorMessage = ""
		health.Status = "healthy"
		health.IsAvailable = true
		health.LastSuccessTime = time.Now()

		r.logger.Debug("module health check passed", map[string]interface{}{
			"module_id":     moduleID,
			"response_time": duration.String(),
		})
	}
	r.healthStatusMu.Unlock()
}

// performanceCalculationWorker calculates performance metrics periodically
func (r *ModuleRegistry) performanceCalculationWorker() {
	ticker := time.NewTicker(1 * time.Minute) // Calculate every minute
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.calculatePerformanceMetrics()
		case <-r.stopChan:
			return
		}
	}
}

// calculatePerformanceMetrics calculates performance metrics for all modules
func (r *ModuleRegistry) calculatePerformanceMetrics() {
	r.performanceMu.Lock()
	defer r.performanceMu.Unlock()

	now := time.Now()

	for _, performance := range r.performance {
		// Calculate rates
		if performance.TotalRequests > 0 {
			performance.SuccessRate = float64(performance.SuccessfulRequests) / float64(performance.TotalRequests)
			performance.ErrorRate = float64(performance.FailedRequests) / float64(performance.TotalRequests)
		}

		// Calculate throughput (requests per second over the last window)
		windowStart := now.Add(-r.config.PerformanceWindowSize)
		if performance.LastRequestTime.After(windowStart) {
			// Simple throughput calculation - in a real implementation, you'd track requests per time window
			performance.Throughput = float64(performance.TotalRequests) / r.config.PerformanceWindowSize.Seconds()
		}

		performance.LastCalculated = now
	}
}

// GetRegistryStats returns overall registry statistics
func (r *ModuleRegistry) GetRegistryStats(ctx context.Context) (*RegistryStats, error) {
	ctx, span := r.tracer.Start(ctx, "ModuleRegistry.GetRegistryStats")
	defer span.End()

	r.modulesMu.RLock()
	totalModules := len(r.modules)
	r.modulesMu.RUnlock()

	r.healthStatusMu.RLock()
	healthyCount := 0
	unhealthyCount := 0
	degradedCount := 0
	for _, health := range r.healthStatus {
		switch health.Status {
		case "healthy":
			healthyCount++
		case "unhealthy":
			unhealthyCount++
		case "degraded":
			degradedCount++
		}
	}
	r.healthStatusMu.RUnlock()

	stats := &RegistryStats{
		TotalModules:     totalModules,
		HealthyModules:   healthyCount,
		UnhealthyModules: unhealthyCount,
		DegradedModules:  degradedCount,
		LastUpdated:      time.Now(),
	}

	return stats, nil
}

// Shutdown gracefully shuts down the registry
func (r *ModuleRegistry) Shutdown(ctx context.Context) error {
	ctx, span := r.tracer.Start(ctx, "ModuleRegistry.Shutdown")
	defer span.End()

	close(r.stopChan)

	r.logger.Info("module registry shutdown complete", map[string]interface{}{
		"total_modules": len(r.modules),
	})

	return nil
}

// RegistryStats represents overall registry statistics
type RegistryStats struct {
	TotalModules     int       `json:"total_modules"`
	HealthyModules   int       `json:"healthy_modules"`
	UnhealthyModules int       `json:"unhealthy_modules"`
	DegradedModules  int       `json:"degraded_modules"`
	LastUpdated      time.Time `json:"last_updated"`
}
