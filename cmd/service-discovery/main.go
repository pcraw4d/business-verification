package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
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

// ServiceDiscoveryServer represents the service discovery server
type ServiceDiscoveryServer struct {
	registry *ServiceRegistry
	port     string
}

// NewServiceDiscoveryServer creates a new service discovery server
func NewServiceDiscoveryServer() *ServiceDiscoveryServer {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	registry := NewServiceRegistry()

	// Start health check loop
	registry.StartHealthCheckLoop(30 * time.Second)

	return &ServiceDiscoveryServer{
		registry: registry,
		port:     port,
	}
}

// handleHealth returns the health status of the service discovery server
func (s *ServiceDiscoveryServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := s.registry.GetRegistryStatus()
	status["service"] = "kyb-service-discovery"
	status["version"] = "4.0.0-SERVICE-DISCOVERY"
	status["timestamp"] = time.Now().Format(time.RFC3339)

	json.NewEncoder(w).Encode(status)
}

// handleRegister handles service registration
func (s *ServiceDiscoveryServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var service Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := s.registry.RegisterService(&service); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Service registered successfully",
		"service": service,
	})
}

// handleUnregister handles service unregistration
func (s *ServiceDiscoveryServer) handleUnregister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	serviceID := vars["id"]

	if err := s.registry.UnregisterService(serviceID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Service unregistered successfully",
		"service_id": serviceID,
	})
}

// handleGetService retrieves a specific service
func (s *ServiceDiscoveryServer) handleGetService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceID := vars["id"]

	service, err := s.registry.GetService(serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service)
}

// handleGetServices retrieves all services
func (s *ServiceDiscoveryServer) handleGetServices(w http.ResponseWriter, r *http.Request) {
	services := s.registry.GetAllServices()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"services":  services,
		"count":     len(services),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// handleGetServicesByTag retrieves services by tag
func (s *ServiceDiscoveryServer) handleGetServicesByTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tag := vars["tag"]

	services := s.registry.GetServicesByTag(tag)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tag":       tag,
		"services":  services,
		"count":     len(services),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// handleGetHealthyServices retrieves only healthy services
func (s *ServiceDiscoveryServer) handleGetHealthyServices(w http.ResponseWriter, r *http.Request) {
	services := s.registry.GetHealthyServices()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"services":  services,
		"count":     len(services),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// handleGetServiceURL retrieves the URL of a service by name
func (s *ServiceDiscoveryServer) handleGetServiceURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	url, err := s.registry.GetServiceURL(serviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service_name": serviceName,
		"url":          url,
		"timestamp":    time.Now().Format(time.RFC3339),
	})
}

// handleCheckHealth checks the health of a specific service
func (s *ServiceDiscoveryServer) handleCheckHealth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceID := vars["id"]

	err := s.registry.CheckServiceHealth(serviceID)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service_id": serviceID,
			"status":     "unhealthy",
			"error":      err.Error(),
			"timestamp":  time.Now().Format(time.RFC3339),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"service_id": serviceID,
		"status":     "healthy",
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

// handleCheckAllHealth checks the health of all services
func (s *ServiceDiscoveryServer) handleCheckAllHealth(w http.ResponseWriter, r *http.Request) {
	results := s.registry.CheckAllServicesHealth()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"health_checks": results,
		"timestamp":     time.Now().Format(time.RFC3339),
	})
}

// handleDashboard returns a service discovery dashboard
func (s *ServiceDiscoveryServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	status := s.registry.GetRegistryStatus()
	services := status["services"].([]*Service)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>KYB Service Discovery Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: #3498db; color: white; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 20px; }
        .metric-card { background: white; padding: 20px; border-radius: 5px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
        .metric-title { font-size: 18px; font-weight: bold; margin-bottom: 10px; color: #3498db; }
        .metric-value { font-size: 24px; font-weight: bold; color: #27ae60; }
        .service-card { background: white; padding: 20px; margin: 10px 0; border-radius: 5px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
        .service-name { font-size: 18px; font-weight: bold; color: #2c3e50; }
        .service-url { color: #7f8c8d; font-size: 14px; }
        .status-healthy { color: #27ae60; font-weight: bold; }
        .status-unhealthy { color: #e74c3c; font-weight: bold; }
        .status-unknown { color: #f39c12; font-weight: bold; }
        .endpoint { background: #ecf0f1; padding: 15px; margin: 10px 0; border-radius: 3px; }
        .method { background: #3498db; color: white; padding: 3px 8px; border-radius: 3px; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>KYB Service Discovery Dashboard</h1>
            <p>Service registry and health monitoring for KYB Platform</p>
        </div>
        
        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-title">Total Services</div>
                <div class="metric-value">%d</div>
            </div>
            
            <div class="metric-card">
                <div class="metric-title">Healthy Services</div>
                <div class="metric-value">%d</div>
            </div>
            
            <div class="metric-card">
                <div class="metric-title">Unhealthy Services</div>
                <div class="metric-value">%d</div>
            </div>
        </div>
        
        <div class="metric-card">
            <h3>Registered Services</h3>`,
		status["total_services"],
		status["healthy_services"],
		status["unhealthy_services"])

	for _, service := range services {
		statusClass := "status-unknown"
		if service.Status == "healthy" {
			statusClass = "status-healthy"
		} else if service.Status == "unhealthy" {
			statusClass = "status-unhealthy"
		}

		html += fmt.Sprintf(`
            <div class="service-card">
                <div class="service-name">%s (%s)</div>
                <div class="service-url">%s</div>
                <div class="%s">Status: %s</div>
                <div>Version: %s</div>
                <div>Last Check: %s</div>
            </div>`,
			service.Name, service.ID, service.URL, statusClass, service.Status,
			service.Version, service.LastCheck.Format("2006-01-02 15:04:05"))
	}

	html += `
        </div>
        
        <div class="metric-card">
            <h3>API Endpoints</h3>
            <div class="endpoint">
                <span class="method">GET</span> /health - Service discovery health
            </div>
            <div class="endpoint">
                <span class="method">GET</span> /services - List all services
            </div>
            <div class="endpoint">
                <span class="method">GET</span> /services/healthy - List healthy services
            </div>
            <div class="endpoint">
                <span class="method">GET</span> /services/{id} - Get specific service
            </div>
            <div class="endpoint">
                <span class="method">GET</span> /services/tag/{tag} - Get services by tag
            </div>
            <div class="endpoint">
                <span class="method">GET</span> /services/name/{name}/url - Get service URL by name
            </div>
            <div class="endpoint">
                <span class="method">POST</span> /register - Register new service
            </div>
            <div class="endpoint">
                <span class="method">DELETE</span> /services/{id} - Unregister service
            </div>
            <div class="endpoint">
                <span class="method">GET</span> /health/{id} - Check service health
            </div>
            <div class="endpoint">
                <span class="method">GET</span> /health/all - Check all services health
            </div>
        </div>
    </div>
</body>
</html>`

	fmt.Fprint(w, html)
}

// setupRoutes configures the HTTP routes
func (s *ServiceDiscoveryServer) setupRoutes() {
	router := mux.NewRouter()

	// Health and status endpoints
	router.HandleFunc("/health", s.handleHealth).Methods("GET")
	router.HandleFunc("/status", s.handleHealth).Methods("GET")

	// Service management endpoints
	router.HandleFunc("/register", s.handleRegister).Methods("POST")
	router.HandleFunc("/services", s.handleGetServices).Methods("GET")
	router.HandleFunc("/services/healthy", s.handleGetHealthyServices).Methods("GET")
	router.HandleFunc("/services/{id}", s.handleGetService).Methods("GET")
	router.HandleFunc("/services/{id}", s.handleUnregister).Methods("DELETE")
	router.HandleFunc("/services/tag/{tag}", s.handleGetServicesByTag).Methods("GET")
	router.HandleFunc("/services/name/{name}/url", s.handleGetServiceURL).Methods("GET")

	// Health check endpoints
	router.HandleFunc("/health/{id}", s.handleCheckHealth).Methods("GET")
	router.HandleFunc("/health/all", s.handleCheckAllHealth).Methods("GET")

	// Dashboard endpoint
	router.HandleFunc("/dashboard", s.handleDashboard).Methods("GET")
	router.HandleFunc("/", s.handleDashboard).Methods("GET")

	http.Handle("/", router)
}

// registerDefaultServices registers the default KYB services
func (s *ServiceDiscoveryServer) registerDefaultServices() {
	defaultServices := []Service{
		{
			ID:        "api-gateway",
			Name:      "API Gateway",
			URL:       "https://api-gateway-service-production-21fd.up.railway.app",
			Version:   "4.0.0-CACHE-BUST-REBUILD",
			HealthURL: "https://api-gateway-service-production-21fd.up.railway.app/health",
			Tags:      []string{"gateway", "api", "core"},
			Metadata: map[string]string{
				"description": "Main API Gateway for KYB Platform",
				"environment": "production",
			},
		},
		{
			ID:        "classification-service",
			Name:      "Classification Service",
			URL:       "https://classification-service-production.up.railway.app",
			Version:   "3.2.0",
			HealthURL: "https://classification-service-production.up.railway.app/health",
			Tags:      []string{"classification", "core", "business"},
			Metadata: map[string]string{
				"description": "Business classification and risk assessment",
				"environment": "production",
			},
		},
		{
			ID:        "merchant-service",
			Name:      "Merchant Service",
			URL:       "https://merchant-service-production.up.railway.app",
			Version:   "3.2.0",
			HealthURL: "https://merchant-service-production.up.railway.app/health",
			Tags:      []string{"merchant", "core", "business"},
			Metadata: map[string]string{
				"description": "Merchant management and operations",
				"environment": "production",
			},
		},
		{
			ID:        "monitoring-service",
			Name:      "Monitoring Service",
			URL:       "https://monitoring-service-production.up.railway.app",
			Version:   "4.0.0-CACHE-BUST-REBUILD",
			HealthURL: "https://monitoring-service-production.up.railway.app/health",
			Tags:      []string{"monitoring", "observability", "core"},
			Metadata: map[string]string{
				"description": "System monitoring and alerting",
				"environment": "production",
			},
		},
		{
			ID:        "pipeline-service",
			Name:      "Pipeline Service",
			URL:       "https://pipeline-service-production.up.railway.app",
			Version:   "4.0.0-CACHE-BUST-REBUILD",
			HealthURL: "https://pipeline-service-production.up.railway.app/health",
			Tags:      []string{"pipeline", "processing", "core"},
			Metadata: map[string]string{
				"description": "Event processing pipeline",
				"environment": "production",
			},
		},
		{
			ID:        "frontend-service",
			Name:      "Frontend Service",
			URL:       "https://frontend-service-production-b225.up.railway.app",
			Version:   "4.0.0-CACHE-BUST-REBUILD",
			HealthURL: "https://frontend-service-production-b225.up.railway.app/health",
			Tags:      []string{"frontend", "ui", "core"},
			Metadata: map[string]string{
				"description": "Web frontend interface",
				"environment": "production",
			},
		},
		{
			ID:        "business-intelligence-gateway",
			Name:      "Business Intelligence Gateway",
			URL:       "https://bi-service-production.up.railway.app",
			Version:   "4.0.0-BI",
			HealthURL: "https://bi-service-production.up.railway.app/health",
			Tags:      []string{"bi", "analytics", "reports"},
			Metadata: map[string]string{
				"description": "Business intelligence and analytics",
				"environment": "production",
			},
		},
		{
			ID:        "risk-assessment-service",
			Name:      "Risk Assessment Service",
			URL:       "https://risk-assessment-service-production.up.railway.app",
			Version:   "1.0.0",
			HealthURL: "https://risk-assessment-service-production.up.railway.app/health",
			Tags:      []string{"risk", "assessment", "core"},
			Metadata: map[string]string{
				"description": "Risk assessment and analysis",
				"environment": "production",
			},
		},
		{
			ID:        "legacy-api-service",
			Name:      "Legacy API Service",
			URL:       "https://shimmering-comfort-production.up.railway.app",
			Version:   "4.0.0-CACHE-BUST-REBUILD",
			HealthURL: "https://shimmering-comfort-production.up.railway.app/health",
			Tags:      []string{"legacy", "api", "monolithic"},
			Metadata: map[string]string{
				"description": "Legacy monolithic API service",
				"environment": "production",
			},
		},
		{
			ID:        "legacy-frontend-service",
			Name:      "Legacy Frontend Service",
			URL:       "https://frontend-ui-production-e727.up.railway.app",
			Version:   "3.2.0",
			HealthURL: "https://frontend-ui-production-e727.up.railway.app/health",
			Tags:      []string{"legacy", "frontend", "ui", "monolithic"},
			Metadata: map[string]string{
				"description": "Legacy monolithic frontend service",
				"environment": "production",
			},
		},
	}

	// Register all default services
	for _, service := range defaultServices {
		if err := s.registry.RegisterService(&service); err != nil {
			log.Printf("❌ Failed to register service %s: %v", service.Name, err)
		}
	}

	log.Printf("✅ Registered %d default services", len(defaultServices))
}

func main() {
	server := NewServiceDiscoveryServer()

	// Register default services
	server.registerDefaultServices()

	// Setup routes
	server.setupRoutes()

	log.Printf("🚀 Starting KYB Service Discovery Server on :%s", server.port)
	log.Printf("✅ Service Discovery Server is ready and listening on :%s", server.port)
	log.Printf("🔗 Health: http://localhost:%s/health", server.port)
	log.Printf("📊 Dashboard: http://localhost:%s/dashboard", server.port)
	log.Printf("📋 Services: http://localhost:%s/services", server.port)

	if err := http.ListenAndServe(":"+server.port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
