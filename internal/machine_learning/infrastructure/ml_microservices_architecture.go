package infrastructure

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// MLMicroservicesArchitecture defines the microservices architecture for ML operations
type MLMicroservicesArchitecture struct {
	// Core services
	pythonMLService  *PythonMLService
	goRuleEngine     *GoRuleEngine
	apiGateway       *APIGateway
	modelRegistry    *ModelRegistry
	serviceDiscovery *ServiceDiscovery
	loadBalancer     *LoadBalancer

	// Configuration
	config MLArchitectureConfig

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

// MLArchitectureConfig holds configuration for the ML microservices architecture
type MLArchitectureConfig struct {
	// Service configuration
	PythonMLServiceEnabled bool `json:"python_ml_service_enabled"`
	GoRuleEngineEnabled    bool `json:"go_rule_engine_enabled"`
	APIGatewayEnabled      bool `json:"api_gateway_enabled"`

	// Service endpoints
	PythonMLServiceEndpoint string `json:"python_ml_service_endpoint"`
	GoRuleEngineEndpoint    string `json:"go_rule_engine_endpoint"`
	APIGatewayEndpoint      string `json:"api_gateway_endpoint"`

	// Load balancing
	LoadBalancingEnabled  bool   `json:"load_balancing_enabled"`
	LoadBalancingStrategy string `json:"load_balancing_strategy"` // round_robin, weighted, least_connections

	// Service discovery
	ServiceDiscoveryEnabled bool   `json:"service_discovery_enabled"`
	ServiceRegistryEndpoint string `json:"service_registry_endpoint"`

	// Health checking
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	HealthCheckTimeout  time.Duration `json:"health_check_timeout"`

	// Performance
	MaxConcurrentRequests int           `json:"max_concurrent_requests"`
	RequestTimeout        time.Duration `json:"request_timeout"`
	CircuitBreakerEnabled bool          `json:"circuit_breaker_enabled"`

	// Monitoring
	MetricsEnabled  bool `json:"metrics_enabled"`
	TracingEnabled  bool `json:"tracing_enabled"`
	LoggingEnabled  bool `json:"logging_enabled"`
	AlertingEnabled bool `json:"alerting_enabled"`
}

// PythonMLService and PythonMLServiceConfig are defined in python_ml_service.go
// GoRuleEngine and GoRuleEngineConfig are defined in go_rule_engine.go
// Types removed to avoid redeclaration

// GoRuleEngineConfig is defined in go_rule_engine.go - removed duplicate

// APIGateway represents the API gateway with intelligent routing
type APIGateway struct {
	// Service configuration
	endpoint string
	config   APIGatewayConfig

	// Routing
	router *IntelligentRouter

	// Feature flags
	featureFlags *FeatureFlagManager

	// Load balancing
	loadBalancer *LoadBalancer

	// Circuit breaker
	circuitBreaker *CircuitBreaker

	// Performance tracking
	metrics *ServiceMetrics

	// Health status
	healthStatus *HealthStatus

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger
}

// NewAPIGateway creates a new API gateway instance
func NewAPIGateway(config APIGatewayConfig, logger *log.Logger) *APIGateway {
	if logger == nil {
		logger = log.Default()
	}
	endpoint := fmt.Sprintf("%s:%d", config.Host, config.Port)
	return &APIGateway{
		endpoint: endpoint,
		config:   config,
		router: &IntelligentRouter{
			config: IntelligentRouterConfig{
				DefaultStrategy: config.DefaultRoutingStrategy,
			},
			strategy: config.DefaultRoutingStrategy,
			logger:   logger,
		},
		featureFlags: &FeatureFlagManager{
			flags: make(map[string]bool),
			config: FeatureFlagConfig{
				UpdateInterval: config.FeatureFlagUpdateInterval,
				Enabled:        config.FeatureFlagsEnabled,
			},
			logger: logger,
		},
		circuitBreaker: &CircuitBreaker{
			config: CircuitBreakerConfig{
				FailureThreshold: config.CircuitBreakerThreshold,
				Timeout:          config.CircuitBreakerTimeout,
			},
			state: "closed",
		},
		metrics: &ServiceMetrics{},
		healthStatus: &HealthStatus{
			Status:    "unknown",
			LastCheck: time.Now(),
			Checks:    make(map[string]HealthCheck),
		},
		logger: logger,
	}
}

// Initialize initializes the API gateway
func (ag *APIGateway) Initialize(ctx context.Context) error {
	ag.mu.Lock()
	defer ag.mu.Unlock()
	ag.logger.Printf("🚪 API Gateway initialized at %s", ag.endpoint)
	return nil
}

// Start starts the API gateway
func (ag *APIGateway) Start(ctx context.Context) error {
	ag.mu.Lock()
	defer ag.mu.Unlock()
	ag.logger.Printf("🚀 API Gateway started")
	return nil
}

// Stop stops the API gateway
func (ag *APIGateway) Stop(ctx context.Context) error {
	ag.mu.Lock()
	defer ag.mu.Unlock()
	ag.logger.Printf("🛑 API Gateway stopped")
	return nil
}

// HealthCheck performs a health check on the API gateway
func (ag *APIGateway) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	ag.mu.RLock()
	defer ag.mu.RUnlock()
	return ag.healthStatus, nil
}

// GetMetrics returns the API gateway metrics
func (ag *APIGateway) GetMetrics(ctx context.Context) (*ServiceMetrics, error) {
	ag.mu.RLock()
	defer ag.mu.RUnlock()
	return ag.metrics, nil
}

// APIGatewayConfig holds configuration for the API gateway
type APIGatewayConfig struct {
	// Service configuration
	Host string `json:"host"`
	Port int    `json:"port"`

	// Routing configuration
	IntelligentRoutingEnabled bool   `json:"intelligent_routing_enabled"`
	DefaultRoutingStrategy    string `json:"default_routing_strategy"` // ml_first, rules_first, hybrid

	// Feature flag configuration
	FeatureFlagsEnabled       bool          `json:"feature_flags_enabled"`
	FeatureFlagUpdateInterval time.Duration `json:"feature_flag_update_interval"`

	// Load balancing configuration
	LoadBalancingEnabled  bool   `json:"load_balancing_enabled"`
	LoadBalancingStrategy string `json:"load_balancing_strategy"`

	// Circuit breaker configuration
	CircuitBreakerEnabled   bool          `json:"circuit_breaker_enabled"`
	CircuitBreakerThreshold int           `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration `json:"circuit_breaker_timeout"`

	// Performance configuration
	MaxConcurrentRequests      int           `json:"max_concurrent_requests"`
	RequestTimeout             time.Duration `json:"request_timeout"`
	RateLimitingEnabled        bool          `json:"rate_limiting_enabled"`
	RateLimitRequestsPerMinute int           `json:"rate_limit_requests_per_minute"`

	// Monitoring
	MetricsEnabled        bool `json:"metrics_enabled"`
	TracingEnabled        bool `json:"tracing_enabled"`
	RequestLoggingEnabled bool `json:"request_logging_enabled"`
}

// ModelRegistry and ModelRegistryConfig are defined in model_registry.go
// Types removed to avoid redeclaration

// ServiceDiscovery manages service discovery and registration
type ServiceDiscovery struct {
	// Discovery configuration
	config ServiceDiscoveryConfig

	// Service registry
	registry map[string]*ServiceInstance

	// Health monitoring
	healthMonitor *HealthMonitor

	// Performance tracking
	metrics *ServiceMetrics

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger
}

// NewServiceDiscovery creates a new service discovery instance
func NewServiceDiscovery(config ServiceDiscoveryConfig, logger *log.Logger) *ServiceDiscovery {
	if logger == nil {
		logger = log.Default()
	}
	return &ServiceDiscovery{
		config:   config,
		registry: make(map[string]*ServiceInstance),
		healthMonitor: &HealthMonitor{
			config: HealthMonitorConfig{
				CheckInterval: config.HealthCheckInterval,
				CheckTimeout:  config.HealthCheckTimeout,
				Enabled:       config.HealthCheckEnabled,
			},
			checks: make(map[string]*HealthCheck),
			logger: logger,
		},
		metrics: &ServiceMetrics{},
		logger:  logger,
	}
}

// Initialize initializes the service discovery
func (sd *ServiceDiscovery) Initialize(ctx context.Context) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	sd.logger.Printf("🔍 Service Discovery initialized")
	return nil
}

// Start starts the service discovery
func (sd *ServiceDiscovery) Start(ctx context.Context) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	sd.logger.Printf("🚀 Service Discovery started")
	return nil
}

// Stop stops the service discovery
func (sd *ServiceDiscovery) Stop(ctx context.Context) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	sd.logger.Printf("🛑 Service Discovery stopped")
	return nil
}

// ServiceDiscoveryConfig holds configuration for service discovery
type ServiceDiscoveryConfig struct {
	// Discovery configuration
	RegistryType     string        `json:"registry_type"` // consul, etcd, zookeeper, custom
	RegistryEndpoint string        `json:"registry_endpoint"`
	ServiceTTL       time.Duration `json:"service_ttl"`

	// Health checking
	HealthCheckEnabled  bool          `json:"health_check_enabled"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	HealthCheckTimeout  time.Duration `json:"health_check_timeout"`

	// Service registration
	AutoRegistrationEnabled bool     `json:"auto_registration_enabled"`
	ServiceName             string   `json:"service_name"`
	ServiceTags             []string `json:"service_tags"`

	// Load balancing
	LoadBalancingEnabled  bool   `json:"load_balancing_enabled"`
	LoadBalancingStrategy string `json:"load_balancing_strategy"`

	// Monitoring
	MetricsEnabled bool `json:"metrics_enabled"`
	EventLogging   bool `json:"event_logging"`
}

// LoadBalancer manages load balancing across service instances
type LoadBalancer struct {
	// Load balancer configuration
	config LoadBalancerConfig

	// Service instances
	instances map[string][]*ServiceInstance

	// Load balancing strategy
	strategy LoadBalancingStrategy

	// Health monitoring
	healthMonitor *HealthMonitor

	// Performance tracking
	metrics *ServiceMetrics

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger
}

// NewLoadBalancer creates a new load balancer instance
func NewLoadBalancer(config LoadBalancerConfig, logger *log.Logger) *LoadBalancer {
	if logger == nil {
		logger = log.Default()
	}
	return &LoadBalancer{
		config:    config,
		instances: make(map[string][]*ServiceInstance),
		healthMonitor: &HealthMonitor{
			config: HealthMonitorConfig{
				CheckInterval: config.HealthCheckInterval,
				CheckTimeout:  config.HealthCheckTimeout,
				Enabled:       config.HealthCheckEnabled,
			},
			checks: make(map[string]*HealthCheck),
			logger: logger,
		},
		metrics: &ServiceMetrics{},
		logger:  logger,
	}
}

// Initialize initializes the load balancer
func (lb *LoadBalancer) Initialize(ctx context.Context) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.logger.Printf("⚖️ Load Balancer initialized")
	return nil
}

// Start starts the load balancer
func (lb *LoadBalancer) Start(ctx context.Context) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.logger.Printf("🚀 Load Balancer started")
	return nil
}

// Stop stops the load balancer
func (lb *LoadBalancer) Stop(ctx context.Context) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.logger.Printf("🛑 Load Balancer stopped")
	return nil
}

// LoadBalancerConfig holds configuration for the load balancer
type LoadBalancerConfig struct {
	// Load balancing configuration
	Strategy string `json:"strategy"` // round_robin, weighted, least_connections, ip_hash

	// Health checking
	HealthCheckEnabled  bool          `json:"health_check_enabled"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	HealthCheckTimeout  time.Duration `json:"health_check_timeout"`

	// Instance management
	MaxInstancesPerService int           `json:"max_instances_per_service"`
	InstanceTimeout        time.Duration `json:"instance_timeout"`

	// Performance
	MaxConcurrentRequests int           `json:"max_concurrent_requests"`
	RequestTimeout        time.Duration `json:"request_timeout"`

	// Monitoring
	MetricsEnabled bool `json:"metrics_enabled"`
	EventLogging   bool `json:"event_logging"`
}

// Supporting types and interfaces

// Types MLModel, ModelVersion, ModelDeployment, ModelMetrics, ServiceInstance, ServiceMetrics,
// HealthStatus, HealthCheck, LoadBalancingStrategy, CircuitBreaker, and CircuitBreakerConfig
// are defined in types.go to avoid redeclaration

// IntelligentRouter handles intelligent routing decisions
type IntelligentRouter struct {
	config   IntelligentRouterConfig
	strategy string
	logger   *log.Logger
	mu       sync.RWMutex
}

// IntelligentRouterConfig holds configuration for intelligent routing
type IntelligentRouterConfig struct {
	DefaultStrategy string  `json:"default_strategy"` // ml_first, rules_first, hybrid
	MLThreshold     float64 `json:"ml_threshold"`
	RulesThreshold  float64 `json:"rules_threshold"`
}

// FeatureFlagManager manages feature flags for the API gateway
type FeatureFlagManager struct {
	flags  map[string]bool
	config FeatureFlagConfig
	logger *log.Logger
	mu     sync.RWMutex
}

// FeatureFlagConfig holds configuration for feature flags
type FeatureFlagConfig struct {
	UpdateInterval time.Duration `json:"update_interval"`
	Enabled        bool          `json:"enabled"`
}

// HealthMonitor monitors health of services
type HealthMonitor struct {
	config HealthMonitorConfig
	checks map[string]*HealthCheck
	logger *log.Logger
	mu     sync.RWMutex
}

// HealthMonitorConfig holds configuration for health monitoring
type HealthMonitorConfig struct {
	CheckInterval time.Duration `json:"check_interval"`
	CheckTimeout  time.Duration `json:"check_timeout"`
	Enabled       bool          `json:"enabled"`
}

// NewMLMicroservicesArchitecture creates a new ML microservices architecture
func NewMLMicroservicesArchitecture(config MLArchitectureConfig, logger *log.Logger) *MLMicroservicesArchitecture {
	if logger == nil {
		logger = log.Default()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &MLMicroservicesArchitecture{
		config: config,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Initialize initializes the ML microservices architecture
func (arch *MLMicroservicesArchitecture) Initialize(ctx context.Context) error {
	arch.mu.Lock()
	defer arch.mu.Unlock()

	arch.logger.Printf("🚀 Initializing ML Microservices Architecture")

	// Initialize Python ML Service with retry logic for resilience (3 retries with exponential backoff)
	// This handles transient ML service startup issues during system initialization
	if arch.config.PythonMLServiceEnabled {
		arch.logger.Printf("🐍 Initializing Python ML Service")
		arch.pythonMLService = NewPythonMLService(arch.config.PythonMLServiceEndpoint, arch.logger)
		if err := arch.pythonMLService.InitializeWithRetry(ctx, 3); err != nil {
			return fmt.Errorf("failed to initialize Python ML Service after retries: %w", err)
		}
	}

	// Initialize Go Rule Engine
	if arch.config.GoRuleEngineEnabled {
		arch.logger.Printf("🔧 Initializing Go Rule Engine")
		arch.goRuleEngine = NewGoRuleEngine(arch.config.GoRuleEngineEndpoint, arch.logger)
		if err := arch.goRuleEngine.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize Go Rule Engine: %w", err)
		}
	}

	// Initialize Model Registry
	arch.logger.Printf("📚 Initializing Model Registry")
	arch.modelRegistry = NewModelRegistry(arch.logger)
	if err := arch.modelRegistry.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize Model Registry: %w", err)
	}

	// Initialize Service Discovery
	if arch.config.ServiceDiscoveryEnabled {
		arch.logger.Printf("🔍 Initializing Service Discovery")
		arch.serviceDiscovery = NewServiceDiscovery(ServiceDiscoveryConfig{
			RegistryType:        "custom",
			RegistryEndpoint:    arch.config.ServiceRegistryEndpoint,
			ServiceTTL:          30 * time.Second,
			HealthCheckEnabled:  true,
			HealthCheckInterval: arch.config.HealthCheckInterval,
			HealthCheckTimeout:  arch.config.HealthCheckTimeout,
			MetricsEnabled:      arch.config.MetricsEnabled,
		}, arch.logger)
		if err := arch.serviceDiscovery.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize Service Discovery: %w", err)
		}
	}

	// Initialize Load Balancer
	if arch.config.LoadBalancingEnabled {
		arch.logger.Printf("⚖️ Initializing Load Balancer")
		arch.loadBalancer = NewLoadBalancer(LoadBalancerConfig{
			Strategy:              arch.config.LoadBalancingStrategy,
			HealthCheckEnabled:    true,
			HealthCheckInterval:   arch.config.HealthCheckInterval,
			HealthCheckTimeout:    arch.config.HealthCheckTimeout,
			MaxConcurrentRequests: arch.config.MaxConcurrentRequests,
			RequestTimeout:        arch.config.RequestTimeout,
			MetricsEnabled:        arch.config.MetricsEnabled,
		}, arch.logger)
		if err := arch.loadBalancer.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize Load Balancer: %w", err)
		}
	}

	// Initialize API Gateway
	if arch.config.APIGatewayEnabled {
		arch.logger.Printf("🌐 Initializing API Gateway")
		arch.apiGateway = NewAPIGateway(APIGatewayConfig{
			Host:                      "localhost",
			Port:                      8080,
			IntelligentRoutingEnabled: true,
			DefaultRoutingStrategy:    "hybrid",
			FeatureFlagsEnabled:       true,
			LoadBalancingEnabled:      arch.config.LoadBalancingEnabled,
			LoadBalancingStrategy:     arch.config.LoadBalancingStrategy,
			CircuitBreakerEnabled:     arch.config.CircuitBreakerEnabled,
			MaxConcurrentRequests:     arch.config.MaxConcurrentRequests,
			RequestTimeout:            arch.config.RequestTimeout,
			MetricsEnabled:            arch.config.MetricsEnabled,
			TracingEnabled:            arch.config.TracingEnabled,
		}, arch.logger)
		if err := arch.apiGateway.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize API Gateway: %w", err)
		}
	}

	arch.logger.Printf("✅ ML Microservices Architecture initialized successfully")
	return nil
}

// Start starts all services in the architecture
func (arch *MLMicroservicesArchitecture) Start(ctx context.Context) error {
	arch.mu.Lock()
	defer arch.mu.Unlock()

	arch.logger.Printf("🚀 Starting ML Microservices Architecture")

	// Start Python ML Service
	if arch.pythonMLService != nil {
		if err := arch.pythonMLService.Start(ctx); err != nil {
			return fmt.Errorf("failed to start Python ML Service: %w", err)
		}
	}

	// Start Go Rule Engine
	if arch.goRuleEngine != nil {
		if err := arch.goRuleEngine.Start(ctx); err != nil {
			return fmt.Errorf("failed to start Go Rule Engine: %w", err)
		}
	}

	// Start Service Discovery
	if arch.serviceDiscovery != nil {
		if err := arch.serviceDiscovery.Start(ctx); err != nil {
			return fmt.Errorf("failed to start Service Discovery: %w", err)
		}
	}

	// Start Load Balancer
	if arch.loadBalancer != nil {
		if err := arch.loadBalancer.Start(ctx); err != nil {
			return fmt.Errorf("failed to start Load Balancer: %w", err)
		}
	}

	// Start API Gateway
	if arch.apiGateway != nil {
		if err := arch.apiGateway.Start(ctx); err != nil {
			return fmt.Errorf("failed to start API Gateway: %w", err)
		}
	}

	arch.logger.Printf("✅ ML Microservices Architecture started successfully")
	return nil
}

// Stop stops all services in the architecture
func (arch *MLMicroservicesArchitecture) Stop() {
	arch.mu.Lock()
	defer arch.mu.Unlock()

	arch.logger.Printf("🛑 Stopping ML Microservices Architecture")

	// Stop API Gateway
	if arch.apiGateway != nil {
		arch.apiGateway.Stop(arch.ctx)
	}

	// Stop Load Balancer
	if arch.loadBalancer != nil {
		arch.loadBalancer.Stop(arch.ctx)
	}

	// Stop Service Discovery
	if arch.serviceDiscovery != nil {
		arch.serviceDiscovery.Stop(arch.ctx)
	}

	// Stop Go Rule Engine
	if arch.goRuleEngine != nil {
		arch.goRuleEngine.Stop()
	}

	// Stop Python ML Service
	if arch.pythonMLService != nil {
		arch.pythonMLService.Stop()
	}

	// Cancel context
	arch.cancel()

	arch.logger.Printf("✅ ML Microservices Architecture stopped successfully")
}

// HealthCheck performs a health check on the entire architecture
func (arch *MLMicroservicesArchitecture) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	arch.mu.RLock()
	defer arch.mu.RUnlock()

	healthStatus := &HealthStatus{
		Status:    "healthy",
		Checks:    make(map[string]HealthCheck),
		LastCheck: time.Now(),
	}

	// Check Python ML Service
	if arch.pythonMLService != nil {
		check, err := arch.pythonMLService.HealthCheck(ctx)
		if err != nil {
			healthStatus.Checks["python_ml_service"] = HealthCheck{
				Name:      "python_ml_service",
				Status:    "fail",
				Message:   err.Error(),
				LastCheck: time.Now(),
			}
			healthStatus.Status = "unhealthy"
		} else {
			healthStatus.Checks["python_ml_service"] = *check
		}
	}

	// Check Go Rule Engine
	if arch.goRuleEngine != nil {
		check, err := arch.goRuleEngine.HealthCheck(ctx)
		if err != nil {
			healthStatus.Checks["go_rule_engine"] = HealthCheck{
				Name:      "go_rule_engine",
				Status:    "fail",
				Message:   err.Error(),
				LastCheck: time.Now(),
			}
			healthStatus.Status = "unhealthy"
		} else {
			healthStatus.Checks["go_rule_engine"] = *check
		}
	}

	// Check API Gateway
	if arch.apiGateway != nil {
		status, err := arch.apiGateway.HealthCheck(ctx)
		if err != nil {
			healthStatus.Checks["api_gateway"] = HealthCheck{
				Name:      "api_gateway",
				Status:    "fail",
				Message:   err.Error(),
				LastCheck: time.Now(),
			}
			healthStatus.Status = "unhealthy"
		} else if status != nil {
			// Extract individual health checks from HealthStatus
			for name, check := range status.Checks {
				healthStatus.Checks["api_gateway_"+name] = check
			}
			if status.Status != "healthy" {
				healthStatus.Status = "unhealthy"
			}
		}
	}

	// Check Model Registry
	if arch.modelRegistry != nil {
		check, err := arch.modelRegistry.HealthCheck(ctx)
		if err != nil {
			healthStatus.Checks["model_registry"] = HealthCheck{
				Name:      "model_registry",
				Status:    "fail",
				Message:   err.Error(),
				LastCheck: time.Now(),
			}
			healthStatus.Status = "unhealthy"
		} else {
			healthStatus.Checks["model_registry"] = *check
		}
	}

	return healthStatus, nil
}

// GetServiceMetrics returns metrics for all services
func (arch *MLMicroservicesArchitecture) GetServiceMetrics(ctx context.Context) (map[string]*ServiceMetrics, error) {
	arch.mu.RLock()
	defer arch.mu.RUnlock()

	metrics := make(map[string]*ServiceMetrics)

	// Get Python ML Service metrics
	if arch.pythonMLService != nil {
		serviceMetrics, err := arch.pythonMLService.GetMetrics(ctx)
		if err == nil {
			metrics["python_ml_service"] = serviceMetrics
		}
	}

	// Get Go Rule Engine metrics
	if arch.goRuleEngine != nil {
		serviceMetrics, err := arch.goRuleEngine.GetMetrics(ctx)
		if err == nil {
			metrics["go_rule_engine"] = serviceMetrics
		}
	}

	// Get API Gateway metrics
	if arch.apiGateway != nil {
		serviceMetrics, err := arch.apiGateway.GetMetrics(ctx)
		if err == nil {
			metrics["api_gateway"] = serviceMetrics
		}
	}

	return metrics, nil
}
