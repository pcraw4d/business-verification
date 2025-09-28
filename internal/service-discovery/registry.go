package servicediscovery

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Service represents a registered service
type Service struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	Version   string            `json:"version"`
	Status    string            `json:"status"`
	HealthURL string            `json:"health_url"`
	LastCheck time.Time         `json:"last_check"`
	Metadata  map[string]string `json:"metadata"`
	Tags      []string          `json:"tags"`
}

// ServiceRegistry manages service discovery
type ServiceRegistry struct {
	services map[string]*Service
	mutex    sync.RWMutex
	client   *http.Client
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]*Service),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// RegisterService registers a new service
func (sr *ServiceRegistry) RegisterService(service *Service) error {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	// Validate service
	if service.ID == "" || service.Name == "" || service.URL == "" {
		return fmt.Errorf("service ID, name, and URL are required")
	}

	// Set default values
	if service.Status == "" {
		service.Status = "unknown"
	}
	if service.HealthURL == "" {
		service.HealthURL = service.URL + "/health"
	}
	if service.Metadata == nil {
		service.Metadata = make(map[string]string)
	}

	service.LastCheck = time.Now()
	sr.services[service.ID] = service

	log.Printf("✅ Registered service: %s (%s) at %s", service.Name, service.ID, service.URL)
	return nil
}

// UnregisterService removes a service from the registry
func (sr *ServiceRegistry) UnregisterService(serviceID string) error {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	if service, exists := sr.services[serviceID]; exists {
		delete(sr.services, serviceID)
		log.Printf("❌ Unregistered service: %s (%s)", service.Name, serviceID)
		return nil
	}

	return fmt.Errorf("service %s not found", serviceID)
}

// GetService retrieves a service by ID
func (sr *ServiceRegistry) GetService(serviceID string) (*Service, error) {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	service, exists := sr.services[serviceID]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceID)
	}

	return service, nil
}

// GetServicesByTag retrieves services by tag
func (sr *ServiceRegistry) GetServicesByTag(tag string) []*Service {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	var services []*Service
	for _, service := range sr.services {
		for _, serviceTag := range service.Tags {
			if serviceTag == tag {
				services = append(services, service)
				break
			}
		}
	}

	return services
}

// GetAllServices returns all registered services
func (sr *ServiceRegistry) GetAllServices() []*Service {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	services := make([]*Service, 0, len(sr.services))
	for _, service := range sr.services {
		services = append(services, service)
	}

	return services
}

// CheckServiceHealth checks the health of a service
func (sr *ServiceRegistry) CheckServiceHealth(serviceID string) error {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	service, exists := sr.services[serviceID]
	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	// Check health endpoint
	resp, err := sr.client.Get(service.HealthURL)
	if err != nil {
		service.Status = "unhealthy"
		service.LastCheck = time.Now()
		return fmt.Errorf("health check failed for %s: %w", serviceID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		service.Status = "unhealthy"
		service.LastCheck = time.Now()
		return fmt.Errorf("health check returned status %d for %s", resp.StatusCode, serviceID)
	}

	service.Status = "healthy"
	service.LastCheck = time.Now()
	return nil
}

// CheckAllServicesHealth checks the health of all services
func (sr *ServiceRegistry) CheckAllServicesHealth() map[string]error {
	sr.mutex.RLock()
	serviceIDs := make([]string, 0, len(sr.services))
	for id := range sr.services {
		serviceIDs = append(serviceIDs, id)
	}
	sr.mutex.RUnlock()

	results := make(map[string]error)
	for _, serviceID := range serviceIDs {
		results[serviceID] = sr.CheckServiceHealth(serviceID)
	}

	return results
}

// GetHealthyServices returns only healthy services
func (sr *ServiceRegistry) GetHealthyServices() []*Service {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	var healthyServices []*Service
	for _, service := range sr.services {
		if service.Status == "healthy" {
			healthyServices = append(healthyServices, service)
		}
	}

	return healthyServices
}

// GetServiceURL retrieves the URL of a service by name
func (sr *ServiceRegistry) GetServiceURL(serviceName string) (string, error) {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	for _, service := range sr.services {
		if service.Name == serviceName && service.Status == "healthy" {
			return service.URL, nil
		}
	}

	return "", fmt.Errorf("healthy service %s not found", serviceName)
}

// StartHealthCheckLoop starts a background health check loop
func (sr *ServiceRegistry) StartHealthCheckLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			log.Printf("🔍 Running health checks for %d services", len(sr.services))
			results := sr.CheckAllServicesHealth()

			healthyCount := 0
			for serviceID, err := range results {
				if err != nil {
					log.Printf("❌ Service %s health check failed: %v", serviceID, err)
				} else {
					healthyCount++
				}
			}

			log.Printf("✅ Health check complete: %d/%d services healthy", healthyCount, len(sr.services))
		}
	}()
}

// GetRegistryStatus returns the current registry status
func (sr *ServiceRegistry) GetRegistryStatus() map[string]interface{} {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	services := make([]*Service, 0, len(sr.services))
	healthyCount := 0

	for _, service := range sr.services {
		services = append(services, service)
		if service.Status == "healthy" {
			healthyCount++
		}
	}

	return map[string]interface{}{
		"total_services":     len(services),
		"healthy_services":   healthyCount,
		"unhealthy_services": len(services) - healthyCount,
		"services":           services,
		"last_updated":       time.Now().Format(time.RFC3339),
	}
}
