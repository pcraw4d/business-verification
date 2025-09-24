package classification

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

// MethodRegistry manages the registration and lifecycle of classification methods
type MethodRegistry struct {
	methods     map[string]*MethodRegistration
	methodOrder []string // Maintains order of method registration
	mutex       sync.RWMutex
	logger      *log.Logger
}

// NewMethodRegistry creates a new method registry
func NewMethodRegistry(logger *log.Logger) *MethodRegistry {
	if logger == nil {
		logger = log.Default()
	}

	return &MethodRegistry{
		methods:     make(map[string]*MethodRegistration),
		methodOrder: make([]string, 0),
		logger:      logger,
	}
}

// RegisterMethod registers a new classification method
func (mr *MethodRegistry) RegisterMethod(method ClassificationMethod, config MethodConfig) error {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	name := method.GetName()
	if name == "" {
		return fmt.Errorf("method name cannot be empty")
	}

	// Check if method is already registered
	if _, exists := mr.methods[name]; exists {
		return fmt.Errorf("method with name '%s' is already registered", name)
	}

	// Validate method dependencies
	dependencies := method.GetRequiredDependencies()
	for _, dep := range dependencies {
		if !mr.isDependencyAvailable(dep) {
			mr.logger.Printf("⚠️ Warning: Method '%s' requires dependency '%s' which may not be available", name, dep)
		}
	}

	// Initialize the method
	if err := method.Initialize(context.Background()); err != nil {
		return fmt.Errorf("failed to initialize method '%s': %w", name, err)
	}

	// Create registration
	registration := &MethodRegistration{
		Method:  method,
		Config:  config,
		Metrics: NewMethodPerformanceMetrics(),
	}

	// Register the method
	mr.methods[name] = registration
	mr.methodOrder = append(mr.methodOrder, name)

	mr.logger.Printf("✅ Registered classification method: %s (type: %s, weight: %.2f)",
		name, method.GetType(), config.Weight)

	return nil
}

// UnregisterMethod removes a classification method from the registry
func (mr *MethodRegistry) UnregisterMethod(name string) error {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	registration, exists := mr.methods[name]
	if !exists {
		return fmt.Errorf("method '%s' is not registered", name)
	}

	// Cleanup the method
	if err := registration.Method.Cleanup(); err != nil {
		mr.logger.Printf("⚠️ Warning: Failed to cleanup method '%s': %v", name, err)
	}

	// Remove from registry
	delete(mr.methods, name)

	// Remove from order
	for i, methodName := range mr.methodOrder {
		if methodName == name {
			mr.methodOrder = append(mr.methodOrder[:i], mr.methodOrder[i+1:]...)
			break
		}
	}

	mr.logger.Printf("✅ Unregistered classification method: %s", name)
	return nil
}

// GetMethod retrieves a registered method by name
func (mr *MethodRegistry) GetMethod(name string) (ClassificationMethod, error) {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	registration, exists := mr.methods[name]
	if !exists {
		return nil, fmt.Errorf("method '%s' is not registered", name)
	}

	return registration.Method, nil
}

// GetEnabledMethods returns all enabled methods in registration order
func (mr *MethodRegistry) GetEnabledMethods() []ClassificationMethod {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	var enabledMethods []ClassificationMethod
	for _, name := range mr.methodOrder {
		if registration, exists := mr.methods[name]; exists && registration.Method.IsEnabled() {
			enabledMethods = append(enabledMethods, registration.Method)
		}
	}

	return enabledMethods
}

// GetAllMethods returns all registered methods in registration order
func (mr *MethodRegistry) GetAllMethods() []ClassificationMethod {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	var allMethods []ClassificationMethod
	for _, name := range mr.methodOrder {
		if registration, exists := mr.methods[name]; exists {
			allMethods = append(allMethods, registration.Method)
		}
	}

	return allMethods
}

// GetMethodsByType returns all methods of a specific type
func (mr *MethodRegistry) GetMethodsByType(methodType string) []ClassificationMethod {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	var methods []ClassificationMethod
	for _, name := range mr.methodOrder {
		if registration, exists := mr.methods[name]; exists && registration.Method.GetType() == methodType {
			methods = append(methods, registration.Method)
		}
	}

	return methods
}

// GetMethodConfig returns the configuration for a specific method
func (mr *MethodRegistry) GetMethodConfig(name string) (*MethodConfig, error) {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	registration, exists := mr.methods[name]
	if !exists {
		return nil, fmt.Errorf("method '%s' is not registered", name)
	}

	return &registration.Config, nil
}

// UpdateMethodConfig updates the configuration for a specific method
func (mr *MethodRegistry) UpdateMethodConfig(name string, config MethodConfig) error {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	registration, exists := mr.methods[name]
	if !exists {
		return fmt.Errorf("method '%s' is not registered", name)
	}

	// Update the method's weight if it has changed
	if config.Weight != registration.Config.Weight {
		registration.Method.SetWeight(config.Weight)
	}

	// Update the method's enabled status if it has changed
	if config.Enabled != registration.Config.Enabled {
		registration.Method.SetEnabled(config.Enabled)
	}

	// Update the config
	registration.Config = config

	mr.logger.Printf("✅ Updated configuration for method '%s': weight=%.2f, enabled=%t",
		name, config.Weight, config.Enabled)

	return nil
}

// GetMethodMetrics returns performance metrics for a specific method
func (mr *MethodRegistry) GetMethodMetrics(name string) (*MethodPerformanceMetrics, error) {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	registration, exists := mr.methods[name]
	if !exists {
		return nil, fmt.Errorf("method '%s' is not registered", name)
	}

	return registration.Metrics, nil
}

// UpdateMethodMetrics updates performance metrics for a method
func (mr *MethodRegistry) UpdateMethodMetrics(name string, success bool, responseTime time.Duration, err error) error {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	registration, exists := mr.methods[name]
	if !exists {
		return fmt.Errorf("method '%s' is not registered", name)
	}

	registration.Metrics.UpdateMetrics(success, responseTime, err)
	return nil
}

// UpdateMethodAccuracy updates the accuracy score for a method
func (mr *MethodRegistry) UpdateMethodAccuracy(name string, accuracy float64) error {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	registration, exists := mr.methods[name]
	if !exists {
		return fmt.Errorf("method '%s' is not registered", name)
	}

	registration.Metrics.UpdateAccuracy(accuracy)
	return nil
}

// GetRegistryStats returns statistics about the registry
func (mr *MethodRegistry) GetRegistryStats() *RegistryStats {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	stats := &RegistryStats{
		TotalMethods:    len(mr.methods),
		EnabledMethods:  0,
		DisabledMethods: 0,
		MethodTypes:     make(map[string]int),
		TotalRequests:   0,
		AverageAccuracy: 0.0,
		LastUpdated:     time.Now(),
	}

	var totalAccuracy float64
	var methodsWithAccuracy int

	for _, registration := range mr.methods {
		if registration.Method.IsEnabled() {
			stats.EnabledMethods++
		} else {
			stats.DisabledMethods++
		}

		// Count method types
		methodType := registration.Method.GetType()
		stats.MethodTypes[methodType]++

		// Aggregate metrics
		stats.TotalRequests += registration.Metrics.TotalRequests
		if registration.Metrics.AccuracyScore > 0 {
			totalAccuracy += registration.Metrics.AccuracyScore
			methodsWithAccuracy++
		}
	}

	// Calculate average accuracy
	if methodsWithAccuracy > 0 {
		stats.AverageAccuracy = totalAccuracy / float64(methodsWithAccuracy)
	}

	return stats
}

// GetMethodsSortedByWeight returns methods sorted by weight (highest first)
func (mr *MethodRegistry) GetMethodsSortedByWeight() []ClassificationMethod {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	var methods []ClassificationMethod
	for _, name := range mr.methodOrder {
		if registration, exists := mr.methods[name]; exists {
			methods = append(methods, registration.Method)
		}
	}

	// Sort by weight (highest first)
	sort.Slice(methods, func(i, j int) bool {
		return methods[i].GetWeight() > methods[j].GetWeight()
	})

	return methods
}

// isDependencyAvailable checks if a dependency is available
func (mr *MethodRegistry) isDependencyAvailable(dependency string) bool {
	// This is a simplified check - in a real implementation, you might check
	// for specific services, databases, APIs, etc.
	switch dependency {
	case "supabase", "database":
		// Check if database connection is available
		return true // Simplified for now
	case "ml_model", "bert":
		// Check if ML models are loaded
		return true // Simplified for now
	case "external_api":
		// Check if external APIs are accessible
		return true // Simplified for now
	default:
		return true // Default to available
	}
}

// RegistryStats represents statistics about the method registry
type RegistryStats struct {
	TotalMethods    int            `json:"total_methods"`
	EnabledMethods  int            `json:"enabled_methods"`
	DisabledMethods int            `json:"disabled_methods"`
	MethodTypes     map[string]int `json:"method_types"`
	TotalRequests   int64          `json:"total_requests"`
	AverageAccuracy float64        `json:"average_accuracy"`
	LastUpdated     time.Time      `json:"last_updated"`
}
