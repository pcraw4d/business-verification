package microservices

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ServiceRegistryImpl implements the ServiceRegistry interface
type ServiceRegistryImpl struct {
	services map[string]ServiceContract
	mu       sync.RWMutex
	logger   *observability.Logger
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(logger *observability.Logger) *ServiceRegistryImpl {
	return &ServiceRegistryImpl{
		services: make(map[string]ServiceContract),
		logger:   logger,
	}
}

// Register registers a service with the registry
func (r *ServiceRegistryImpl) Register(service ServiceContract) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	serviceName := service.ServiceName()
	if _, exists := r.services[serviceName]; exists {
		return fmt.Errorf("service %s is already registered", serviceName)
	}

	r.services[serviceName] = service
	r.logger.Info("Service registered",
		"service_name", serviceName,
		"version", service.Version(),
		"capabilities_count", len(service.Capabilities()),
	)

	return nil
}

// Unregister removes a service from the registry
func (r *ServiceRegistryImpl) Unregister(serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[serviceName]; !exists {
		return fmt.Errorf("service %s is not registered", serviceName)
	}

	delete(r.services, serviceName)
	r.logger.Info("Service unregistered", "service_name", serviceName)

	return nil
}

// GetService retrieves a service by name
func (r *ServiceRegistryImpl) GetService(serviceName string) (ServiceContract, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	return service, nil
}

// ListServices returns all registered services
func (r *ServiceRegistryImpl) ListServices() []ServiceContract {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]ServiceContract, 0, len(r.services))
	for _, service := range r.services {
		services = append(services, service)
	}

	return services
}

// GetHealthyServices returns only healthy services
func (r *ServiceRegistryImpl) GetHealthyServices() []ServiceContract {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var healthyServices []ServiceContract
	for _, service := range r.services {
		health := service.Health()
		if health.Status == "healthy" {
			healthyServices = append(healthyServices, service)
		}
	}

	return healthyServices
}

// ServiceDiscoveryImpl implements the ServiceDiscovery interface
type ServiceDiscoveryImpl struct {
	instances map[string][]ServiceInstance
	watchers  map[string][]chan ServiceEvent
	mu        sync.RWMutex
	logger    *observability.Logger
	registry  *ServiceRegistryImpl
}

// NewServiceDiscovery creates a new service discovery instance
func NewServiceDiscovery(logger *observability.Logger, registry *ServiceRegistryImpl) *ServiceDiscoveryImpl {
	return &ServiceDiscoveryImpl{
		instances: make(map[string][]ServiceInstance),
		watchers:  make(map[string][]chan ServiceEvent),
		logger:    logger,
		registry:  registry,
	}
}

// RegisterInstance registers a service instance for discovery
func (d *ServiceDiscoveryImpl) RegisterInstance(instance ServiceInstance) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	serviceName := instance.ServiceName
	if d.instances[serviceName] == nil {
		d.instances[serviceName] = make([]ServiceInstance, 0)
	}

	// Check if instance already exists
	for i, existingInstance := range d.instances[serviceName] {
		if existingInstance.ID == instance.ID {
			// Update existing instance
			d.instances[serviceName][i] = instance
			d.notifyWatchers(serviceName, ServiceEvent{
				Type:      ServiceEventUpdated,
				Instance:  instance,
				Timestamp: time.Now(),
			})
			return nil
		}
	}

	// Add new instance
	d.instances[serviceName] = append(d.instances[serviceName], instance)
	d.notifyWatchers(serviceName, ServiceEvent{
		Type:      ServiceEventAdded,
		Instance:  instance,
		Timestamp: time.Now(),
	})

	d.logger.Info("Service instance registered",
		"service_name", serviceName,
		"instance_id", instance.ID,
		"host", instance.Host,
		"port", instance.Port,
	)

	return nil
}

// UnregisterInstance removes a service instance from discovery
func (d *ServiceDiscoveryImpl) UnregisterInstance(serviceName, instanceID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	instances, exists := d.instances[serviceName]
	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	for i, instance := range instances {
		if instance.ID == instanceID {
			// Remove instance
			d.instances[serviceName] = append(instances[:i], instances[i+1:]...)

			d.notifyWatchers(serviceName, ServiceEvent{
				Type:      ServiceEventRemoved,
				Instance:  instance,
				Timestamp: time.Now(),
			})

			d.logger.Info("Service instance unregistered",
				"service_name", serviceName,
				"instance_id", instanceID,
			)

			return nil
		}
	}

	return fmt.Errorf("instance %s not found for service %s", instanceID, serviceName)
}

// Discover returns all instances of a service
func (d *ServiceDiscoveryImpl) Discover(serviceName string) ([]ServiceInstance, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	instances, exists := d.instances[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Filter out unhealthy instances
	var healthyInstances []ServiceInstance
	for _, instance := range instances {
		if instance.Health.Status == "healthy" {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	return healthyInstances, nil
}

// DiscoverAll returns all service instances
func (d *ServiceDiscoveryImpl) DiscoverAll() map[string][]ServiceInstance {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make(map[string][]ServiceInstance)
	for serviceName, instances := range d.instances {
		result[serviceName] = make([]ServiceInstance, len(instances))
		copy(result[serviceName], instances)
	}

	return result
}

// Watch watches for service events
func (d *ServiceDiscoveryImpl) Watch(serviceName string) (<-chan ServiceEvent, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	eventChan := make(chan ServiceEvent, 100)

	if d.watchers[serviceName] == nil {
		d.watchers[serviceName] = make([]chan ServiceEvent, 0)
	}

	d.watchers[serviceName] = append(d.watchers[serviceName], eventChan)

	d.logger.Info("Service watcher added", "service_name", serviceName)

	return eventChan, nil
}

// Unwatch stops watching for service events
func (d *ServiceDiscoveryImpl) Unwatch(serviceName string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	watchers, exists := d.watchers[serviceName]
	if !exists {
		return fmt.Errorf("no watchers for service %s", serviceName)
	}

	// Close all watcher channels
	for _, watcher := range watchers {
		close(watcher)
	}

	delete(d.watchers, serviceName)
	d.logger.Info("Service watchers removed", "service_name", serviceName)

	return nil
}

// notifyWatchers notifies all watchers of a service event
func (d *ServiceDiscoveryImpl) notifyWatchers(serviceName string, event ServiceEvent) {
	watchers, exists := d.watchers[serviceName]
	if !exists {
		return
	}

	for _, watcher := range watchers {
		select {
		case watcher <- event:
			// Event sent successfully
		default:
			// Channel is full, skip this event
			d.logger.Warn("Service event channel full, skipping event",
				"service_name", serviceName,
				"event_type", event.Type,
			)
		}
	}
}

// UpdateInstanceHealth updates the health status of a service instance
func (d *ServiceDiscoveryImpl) UpdateInstanceHealth(serviceName, instanceID string, health ServiceHealth) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	instances, exists := d.instances[serviceName]
	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	for i, instance := range instances {
		if instance.ID == instanceID {
			// Update health status
			instance.Health = health
			instance.LastSeen = time.Now()
			d.instances[serviceName][i] = instance

			d.notifyWatchers(serviceName, ServiceEvent{
				Type:      ServiceEventUpdated,
				Instance:  instance,
				Timestamp: time.Now(),
			})

			d.logger.Info("Service instance health updated",
				"service_name", serviceName,
				"instance_id", instanceID,
				"health_status", health.Status,
			)

			return nil
		}
	}

	return fmt.Errorf("instance %s not found for service %s", instanceID, serviceName)
}

// CleanupStaleInstances removes instances that haven't been seen recently
func (d *ServiceDiscoveryImpl) CleanupStaleInstances(timeout time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	for serviceName, instances := range d.instances {
		var activeInstances []ServiceInstance
		for _, instance := range instances {
			if now.Sub(instance.LastSeen) < timeout {
				activeInstances = append(activeInstances, instance)
			} else {
				d.logger.Info("Removing stale service instance",
					"service_name", serviceName,
					"instance_id", instance.ID,
					"last_seen", instance.LastSeen,
				)

				d.notifyWatchers(serviceName, ServiceEvent{
					Type:      ServiceEventRemoved,
					Instance:  instance,
					Timestamp: now,
				})
			}
		}
		d.instances[serviceName] = activeInstances
	}
}

// StartHealthCheck starts periodic health checks for all instances
func (d *ServiceDiscoveryImpl) StartHealthCheck(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.performHealthChecks()
		}
	}
}

// performHealthChecks performs health checks on all service instances
func (d *ServiceDiscoveryImpl) performHealthChecks() {
	d.mu.RLock()
	instances := make(map[string][]ServiceInstance)
	for serviceName, serviceInstances := range d.instances {
		instances[serviceName] = make([]ServiceInstance, len(serviceInstances))
		copy(instances[serviceName], serviceInstances)
	}
	d.mu.RUnlock()

	for serviceName, serviceInstances := range instances {
		for _, instance := range serviceInstances {
			// Perform health check (simplified for now)
			health := ServiceHealth{
				Status:    "healthy", // TODO: Implement actual health check
				Message:   "Health check passed",
				Timestamp: time.Now(),
			}

			if err := d.UpdateInstanceHealth(serviceName, instance.ID, health); err != nil {
				d.logger.Error("Failed to update instance health",
					"service_name", serviceName,
					"instance_id", instance.ID,
					"error", err,
				)
			}
		}
	}
}

// GetServiceInfo returns detailed information about a service
func (d *ServiceDiscoveryImpl) GetServiceInfo(serviceName string) (map[string]interface{}, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	instances, exists := d.instances[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Get service contract from registry
	service, err := d.registry.GetService(serviceName)
	if err != nil {
		return nil, err
	}

	// Count instances by health status
	healthyCount := 0
	unhealthyCount := 0
	for _, instance := range instances {
		if instance.Health.Status == "healthy" {
			healthyCount++
		} else {
			unhealthyCount++
		}
	}

	info := map[string]interface{}{
		"service_name":    serviceName,
		"version":         service.Version(),
		"capabilities":    service.Capabilities(),
		"health":          service.Health(),
		"total_instances": len(instances),
		"healthy_count":   healthyCount,
		"unhealthy_count": unhealthyCount,
		"instances":       instances,
	}

	return info, nil
}
