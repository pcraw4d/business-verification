package microservices

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ServiceIsolationManager manages service isolation and fault tolerance
type ServiceIsolationManager struct {
	services       map[string]*IsolatedService
	mu             sync.RWMutex
	logger         *observability.Logger
	metrics        ServiceMetrics
	circuitBreaker ServiceCircuitBreaker
}

// IsolatedService represents an isolated service with fault tolerance
type IsolatedService struct {
	Name           string
	Contract       ServiceContract
	Instances      []ServiceInstance
	Health         ServiceHealth
	IsolationLevel IsolationLevel
	FallbackConfig FallbackConfig
	LastUpdated    time.Time
}

// IsolationLevel represents the level of service isolation
type IsolationLevel string

const (
	IsolationLevelNone     IsolationLevel = "none"
	IsolationLevelBasic    IsolationLevel = "basic"
	IsolationLevelEnhanced IsolationLevel = "enhanced"
	IsolationLevelFull     IsolationLevel = "full"
)

// Fallback strategy constants
const (
	FallbackStrategyStatic      = "static_data"
	FallbackStrategyCached      = "cached_data"
	FallbackStrategyAlternative = "alternative_service"
	FallbackStrategyDegraded    = "degraded_response"
)

// FallbackConfig represents fallback configuration for a service
type FallbackConfig struct {
	Enabled        bool                   `json:"enabled"`
	Strategy       string                 `json:"strategy"`
	MaxRetries     int                    `json:"max_retries"`
	RetryDelay     time.Duration          `json:"retry_delay"`
	Timeout        time.Duration          `json:"timeout"`
	CircuitBreaker bool                   `json:"circuit_breaker"`
	FallbackData   map[string]interface{} `json:"fallback_data,omitempty"`
}

// NewServiceIsolationManager creates a new service isolation manager
func NewServiceIsolationManager(
	logger *observability.Logger,
	metrics ServiceMetrics,
	circuitBreaker ServiceCircuitBreaker,
) *ServiceIsolationManager {
	return &ServiceIsolationManager{
		services:       make(map[string]*IsolatedService),
		logger:         logger,
		metrics:        metrics,
		circuitBreaker: circuitBreaker,
	}
}

// RegisterService registers a service with isolation management
func (sim *ServiceIsolationManager) RegisterService(
	service ServiceContract,
	isolationLevel IsolationLevel,
	fallbackConfig FallbackConfig,
) error {
	sim.mu.Lock()
	defer sim.mu.Unlock()

	serviceName := service.ServiceName()
	if _, exists := sim.services[serviceName]; exists {
		return fmt.Errorf("service %s is already registered for isolation", serviceName)
	}

	isolatedService := &IsolatedService{
		Name:           serviceName,
		Contract:       service,
		Instances:      make([]ServiceInstance, 0),
		Health:         service.Health(),
		IsolationLevel: isolationLevel,
		FallbackConfig: fallbackConfig,
		LastUpdated:    time.Now(),
	}

	sim.services[serviceName] = isolatedService

	sim.logger.Info("Service registered for isolation", map[string]interface{}{
		"service_name":     serviceName,
		"isolation_level":  isolationLevel,
		"fallback_enabled": fallbackConfig.Enabled,
	})

	return nil
}

// UnregisterService removes a service from isolation management
func (sim *ServiceIsolationManager) UnregisterService(serviceName string) error {
	sim.mu.Lock()
	defer sim.mu.Unlock()

	if _, exists := sim.services[serviceName]; !exists {
		return fmt.Errorf("service %s is not registered for isolation", serviceName)
	}

	delete(sim.services, serviceName)
	sim.logger.Info("Service unregistered from isolation", map[string]interface{}{
		"service_name": serviceName,
	})

	return nil
}

// ExecuteWithIsolation executes a service call with isolation and fault tolerance
func (sim *ServiceIsolationManager) ExecuteWithIsolation(
	ctx context.Context,
	serviceName string,
	method string,
	request interface{},
) (interface{}, error) {
	sim.mu.RLock()
	isolatedService, exists := sim.services[serviceName]
	sim.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service %s is not registered for isolation", serviceName)
	}

	// Apply isolation level
	switch isolatedService.IsolationLevel {
	case IsolationLevelNone:
		return sim.executeDirect(ctx, serviceName, method, request)
	case IsolationLevelBasic:
		return sim.executeWithBasicIsolation(ctx, isolatedService, method, request)
	case IsolationLevelEnhanced:
		return sim.executeWithEnhancedIsolation(ctx, isolatedService, method, request)
	case IsolationLevelFull:
		return sim.executeWithFullIsolation(ctx, isolatedService, method, request)
	default:
		return sim.executeDirect(ctx, serviceName, method, request)
	}
}

// executeDirect executes a service call without isolation
func (sim *ServiceIsolationManager) executeDirect(
	ctx context.Context,
	serviceName string,
	method string,
	request interface{},
) (interface{}, error) {
	start := time.Now()

	// Use circuit breaker if enabled
	if sim.circuitBreaker != nil {
		result, err := sim.circuitBreaker.Execute(ctx, serviceName, method, request)

		// Record metrics
		duration := time.Since(start)
		success := err == nil
		sim.metrics.RecordRequest(serviceName, method, duration, success)

		return result, err
	}

	// Fallback to direct execution
	return map[string]interface{}{
		"service": serviceName,
		"method":  method,
		"result":  "success",
		"data":    request,
	}, nil
}

// executeWithBasicIsolation executes with basic isolation (timeout and retry)
func (sim *ServiceIsolationManager) executeWithBasicIsolation(
	ctx context.Context,
	isolatedService *IsolatedService,
	method string,
	request interface{},
) (interface{}, error) {
	config := isolatedService.FallbackConfig

	// Apply timeout
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// Execute with retry
	var lastErr error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(config.RetryDelay)
		}

		result, err := sim.executeDirect(ctx, isolatedService.Name, method, request)
		if err == nil {
			return result, nil
		}

		lastErr = err
		sim.logger.Warn("Service call failed, retrying", map[string]interface{}{
			"service_name": isolatedService.Name,
			"method":       method,
			"attempt":      attempt + 1,
			"max_attempts": config.MaxRetries + 1,
			"error":        err.Error(),
		})
	}

	// All retries failed, use fallback if enabled
	if config.Enabled {
		return sim.executeFallback(isolatedService, method, request, lastErr)
	}

	return nil, lastErr
}

// executeWithEnhancedIsolation executes with enhanced isolation (circuit breaker + fallback)
func (sim *ServiceIsolationManager) executeWithEnhancedIsolation(
	ctx context.Context,
	isolatedService *IsolatedService,
	method string,
	request interface{},
) (interface{}, error) {
	config := isolatedService.FallbackConfig

	// Apply timeout
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// Execute with circuit breaker
	result, err := sim.executeDirect(ctx, isolatedService.Name, method, request)
	if err == nil {
		return result, nil
	}

	// Check if circuit breaker is open
	state := sim.circuitBreaker.GetState(isolatedService.Name)
	if state.State == "open" && config.Enabled {
		sim.logger.Warn("Circuit breaker open, using fallback", map[string]interface{}{
			"service_name": isolatedService.Name,
			"method":       method,
		})
		return sim.executeFallback(isolatedService, method, request, err)
	}

	return nil, err
}

// executeWithFullIsolation executes with full isolation (all protections)
func (sim *ServiceIsolationManager) executeWithFullIsolation(
	ctx context.Context,
	isolatedService *IsolatedService,
	method string,
	request interface{},
) (interface{}, error) {
	config := isolatedService.FallbackConfig

	// Apply timeout
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// Execute with all protections
	result, err := sim.executeWithEnhancedIsolation(ctx, isolatedService, method, request)
	if err == nil {
		return result, nil
	}

	// Always use fallback for full isolation
	if config.Enabled {
		return sim.executeFallback(isolatedService, method, request, err)
	}

	return nil, err
}

// executeFallback executes the fallback strategy
func (sim *ServiceIsolationManager) executeFallback(
	isolatedService *IsolatedService,
	method string,
	request interface{},
	originalError error,
) (interface{}, error) {
	config := isolatedService.FallbackConfig

	sim.logger.Info("Executing fallback strategy", map[string]interface{}{
		"service_name":   isolatedService.Name,
		"method":         method,
		"fallback_type":  config.Strategy,
		"original_error": originalError.Error(),
	})

	switch config.Strategy {
	case FallbackStrategyStatic:
		return config.FallbackData, nil
	case FallbackStrategyCached:
		return sim.getCachedData(isolatedService.Name, method)
	case FallbackStrategyAlternative:
		// For alternative service, we would need additional config
		return sim.callAlternativeService("alternative-service", method, request)
	case FallbackStrategyDegraded:
		return sim.generateDegradedResponse(isolatedService.Name, method, request)
	default:
		return nil, fmt.Errorf("unknown fallback strategy: %s", config.Strategy)
	}
}

// getCachedData retrieves cached data for fallback
func (sim *ServiceIsolationManager) getCachedData(serviceName, method string) (interface{}, error) {
	// TODO: Implement actual cache retrieval
	return map[string]interface{}{
		"service": serviceName,
		"method":  method,
		"source":  "cache",
		"data":    "cached_response",
	}, nil
}

// callAlternativeService calls an alternative service
func (sim *ServiceIsolationManager) callAlternativeService(serviceName, method string, request interface{}) (interface{}, error) {
	// TODO: Implement actual alternative service call
	return map[string]interface{}{
		"service": serviceName,
		"method":  method,
		"source":  "alternative_service",
		"data":    request,
	}, nil
}

// generateDegradedResponse generates a degraded response
func (sim *ServiceIsolationManager) generateDegradedResponse(serviceName, method string, request interface{}) (interface{}, error) {
	return map[string]interface{}{
		"service":     serviceName,
		"method":      method,
		"source":      "degraded_response",
		"status":      "degraded",
		"data":        request,
		"degraded_at": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// UpdateServiceHealth updates the health status of a service
func (sim *ServiceIsolationManager) UpdateServiceHealth(serviceName string, health ServiceHealth) error {
	sim.mu.Lock()
	defer sim.mu.Unlock()

	isolatedService, exists := sim.services[serviceName]
	if !exists {
		return fmt.Errorf("service %s is not registered for isolation", serviceName)
	}

	isolatedService.Health = health
	isolatedService.LastUpdated = time.Now()

	sim.logger.Info("Service health updated", map[string]interface{}{
		"service_name":   serviceName,
		"health_status":  health.Status,
		"health_message": health.Message,
	})

	return nil
}

// GetServiceIsolationInfo returns isolation information for a service
func (sim *ServiceIsolationManager) GetServiceIsolationInfo(serviceName string) (map[string]interface{}, error) {
	sim.mu.RLock()
	defer sim.mu.RUnlock()

	isolatedService, exists := sim.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s is not registered for isolation", serviceName)
	}

	info := map[string]interface{}{
		"service_name":    isolatedService.Name,
		"isolation_level": isolatedService.IsolationLevel,
		"health":          isolatedService.Health,
		"fallback_config": isolatedService.FallbackConfig,
		"last_updated":    isolatedService.LastUpdated,
		"capabilities":    isolatedService.Contract.Capabilities(),
		"version":         isolatedService.Contract.Version(),
	}

	return info, nil
}

// ListIsolatedServices returns all services with isolation management
func (sim *ServiceIsolationManager) ListIsolatedServices() []string {
	sim.mu.RLock()
	defer sim.mu.RUnlock()

	services := make([]string, 0, len(sim.services))
	for serviceName := range sim.services {
		services = append(services, serviceName)
	}

	return services
}

// GetIsolationStats returns statistics about service isolation
func (sim *ServiceIsolationManager) GetIsolationStats() map[string]interface{} {
	sim.mu.RLock()
	defer sim.mu.RUnlock()

	stats := map[string]interface{}{
		"total_services": len(sim.services),
		"isolation_levels": map[string]int{
			"none":     0,
			"basic":    0,
			"enhanced": 0,
			"full":     0,
		},
		"fallback_enabled":   0,
		"healthy_services":   0,
		"unhealthy_services": 0,
	}

	for _, service := range sim.services {
		// Count isolation levels
		level := string(service.IsolationLevel)
		if count, exists := stats["isolation_levels"].(map[string]int); exists {
			count[level]++
		}

		// Count fallback enabled
		if service.FallbackConfig.Enabled {
			stats["fallback_enabled"] = stats["fallback_enabled"].(int) + 1
		}

		// Count health status
		if service.Health.Status == "healthy" {
			stats["healthy_services"] = stats["healthy_services"].(int) + 1
		} else {
			stats["unhealthy_services"] = stats["unhealthy_services"].(int) + 1
		}
	}

	return stats
}

// SetIsolationLevel sets the isolation level for a service
func (sim *ServiceIsolationManager) SetIsolationLevel(serviceName string, level IsolationLevel) error {
	sim.mu.Lock()
	defer sim.mu.Unlock()

	isolatedService, exists := sim.services[serviceName]
	if !exists {
		return fmt.Errorf("service %s is not registered for isolation", serviceName)
	}

	isolatedService.IsolationLevel = level
	isolatedService.LastUpdated = time.Now()

	sim.logger.Info("Service isolation level updated", map[string]interface{}{
		"service_name":    serviceName,
		"isolation_level": level,
	})

	return nil
}

// UpdateFallbackConfig updates the fallback configuration for a service
func (sim *ServiceIsolationManager) UpdateFallbackConfig(serviceName string, config FallbackConfig) error {
	sim.mu.Lock()
	defer sim.mu.Unlock()

	isolatedService, exists := sim.services[serviceName]
	if !exists {
		return fmt.Errorf("service %s is not registered for isolation", serviceName)
	}

	isolatedService.FallbackConfig = config
	isolatedService.LastUpdated = time.Now()

	sim.logger.Info("Service fallback config updated", map[string]interface{}{
		"service_name":      serviceName,
		"fallback_enabled":  config.Enabled,
		"fallback_strategy": config.Strategy,
	})

	return nil
}
