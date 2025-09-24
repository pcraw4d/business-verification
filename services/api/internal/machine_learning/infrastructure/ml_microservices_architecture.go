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

// PythonMLService represents the Python ML service for all ML models
type PythonMLService struct {
	// Service configuration
	endpoint string
	config   PythonMLServiceConfig

	// Model management
	models map[string]*MLModel

	// Performance tracking
	metrics *ServiceMetrics

	// Health status
	healthStatus *HealthStatus

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger
}

// PythonMLServiceConfig holds configuration for the Python ML service
type PythonMLServiceConfig struct {
	// Service configuration
	Host string `json:"host"`
	Port int    `json:"port"`

	// Model configuration
	DefaultModelType    string        `json:"default_model_type"` // bert, distilbert, custom
	SupportedModelTypes []string      `json:"supported_model_types"`
	ModelCacheEnabled   bool          `json:"model_cache_enabled"`
	ModelCacheSize      int           `json:"model_cache_size"`
	ModelUpdateInterval time.Duration `json:"model_update_interval"`

	// Performance configuration
	MaxBatchSize        int           `json:"max_batch_size"`
	InferenceTimeout    time.Duration `json:"inference_timeout"`
	ModelLoadingTimeout time.Duration `json:"model_loading_timeout"`

	// Resource limits
	MaxMemoryUsage      int64 `json:"max_memory_usage"` // in MB
	MaxCPUUsage         int   `json:"max_cpu_usage"`    // percentage
	MaxConcurrentModels int   `json:"max_concurrent_models"`

	// Monitoring
	MetricsEnabled      bool `json:"metrics_enabled"`
	PerformanceTracking bool `json:"performance_tracking"`
	ModelVersioning     bool `json:"model_versioning"`
}

// GoRuleEngine represents the Go rule engine for rule-based systems
type GoRuleEngine struct {
	// Service configuration
	endpoint string
	config   GoRuleEngineConfig

	// Rule systems
	keywordMatcher   *KeywordMatcher
	mccCodeLookup    *MCCCodeLookup
	blacklistChecker *BlacklistChecker

	// Caching
	cache *RuleEngineCache

	// Performance tracking
	metrics *ServiceMetrics

	// Health status
	healthStatus *HealthStatus

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger
}

// GoRuleEngineConfig holds configuration for the Go rule engine
type GoRuleEngineConfig struct {
	// Service configuration
	Host string `json:"host"`
	Port int    `json:"port"`

	// Rule system configuration
	KeywordMatchingEnabled bool `json:"keyword_matching_enabled"`
	MCCCodeLookupEnabled   bool `json:"mcc_code_lookup_enabled"`
	BlacklistCheckEnabled  bool `json:"blacklist_check_enabled"`

	// Performance configuration
	MaxConcurrentRules int           `json:"max_concurrent_rules"`
	RuleTimeout        time.Duration `json:"rule_timeout"`
	CacheEnabled       bool          `json:"cache_enabled"`
	CacheSize          int           `json:"cache_size"`
	CacheTTL           time.Duration `json:"cache_ttl"`

	// Rule data sources
	KeywordDatabasePath   string `json:"keyword_database_path"`
	MCCCodeDatabasePath   string `json:"mcc_code_database_path"`
	BlacklistDatabasePath string `json:"blacklist_database_path"`

	// Performance targets
	TargetResponseTime time.Duration `json:"target_response_time"` // <10ms
	TargetAccuracy     float64       `json:"target_accuracy"`      // >90%
}

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

// ModelRegistry manages model versions and deployments
type ModelRegistry struct {
	// Registry configuration
	config ModelRegistryConfig

	// Model storage
	models map[string]*ModelVersion

	// Deployment tracking
	deployments map[string]*ModelDeployment

	// Version history
	versionHistory map[string][]*ModelVersion

	// Performance tracking
	metrics *ServiceMetrics

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger
}

// ModelRegistryConfig holds configuration for the model registry
type ModelRegistryConfig struct {
	// Registry configuration
	StorageType    string        `json:"storage_type"` // local, s3, gcs, azure
	StoragePath    string        `json:"storage_path"`
	BackupEnabled  bool          `json:"backup_enabled"`
	BackupInterval time.Duration `json:"backup_interval"`

	// Model management
	MaxModelVersions int           `json:"max_model_versions"`
	ModelRetention   time.Duration `json:"model_retention"`
	AutoCleanup      bool          `json:"auto_cleanup"`

	// Versioning
	VersioningEnabled  bool   `json:"versioning_enabled"`
	VersioningStrategy string `json:"versioning_strategy"` // semantic, timestamp, hash

	// Deployment
	AutoDeploymentEnabled bool `json:"auto_deployment_enabled"`
	DeploymentValidation  bool `json:"deployment_validation"`

	// Monitoring
	MetricsEnabled      bool `json:"metrics_enabled"`
	DeploymentTracking  bool `json:"deployment_tracking"`
	PerformanceTracking bool `json:"performance_tracking"`
}

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

// MLModel represents a machine learning model
type MLModel struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"` // bert, distilbert, custom
	Version    string    `json:"version"`
	ModelPath  string    `json:"model_path"`
	ConfigPath string    `json:"config_path"`
	IsActive   bool      `json:"is_active"`
	IsDeployed bool      `json:"is_deployed"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastUsed   time.Time `json:"last_used"`
}

// ModelVersion represents a version of a model
type ModelVersion struct {
	ID         string       `json:"id"`
	ModelID    string       `json:"model_id"`
	Version    string       `json:"version"`
	ModelPath  string       `json:"model_path"`
	ConfigPath string       `json:"config_path"`
	Metrics    ModelMetrics `json:"metrics"`
	IsActive   bool         `json:"is_active"`
	CreatedAt  time.Time    `json:"created_at"`
	DeployedAt time.Time    `json:"deployed_at"`
}

// ModelDeployment represents a model deployment
type ModelDeployment struct {
	ID          string    `json:"id"`
	ModelID     string    `json:"model_id"`
	Version     string    `json:"version"`
	ServiceName string    `json:"service_name"`
	Endpoint    string    `json:"endpoint"`
	Status      string    `json:"status"` // deploying, active, failed, stopped
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ModelMetrics represents model performance metrics
type ModelMetrics struct {
	Accuracy      float64       `json:"accuracy"`
	Precision     float64       `json:"precision"`
	Recall        float64       `json:"recall"`
	F1Score       float64       `json:"f1_score"`
	InferenceTime time.Duration `json:"inference_time"`
	Throughput    int           `json:"throughput"` // requests per second
}

// ServiceInstance represents a service instance
type ServiceInstance struct {
	ID          string                 `json:"id"`
	ServiceName string                 `json:"service_name"`
	Endpoint    string                 `json:"endpoint"`
	Health      string                 `json:"health"` // healthy, unhealthy, unknown
	LastCheck   time.Time              `json:"last_check"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ServiceMetrics represents service performance metrics
type ServiceMetrics struct {
	RequestCount   int64         `json:"request_count"`
	SuccessCount   int64         `json:"success_count"`
	ErrorCount     int64         `json:"error_count"`
	AverageLatency time.Duration `json:"average_latency"`
	P95Latency     time.Duration `json:"p95_latency"`
	P99Latency     time.Duration `json:"p99_latency"`
	Throughput     float64       `json:"throughput"` // requests per second
	ErrorRate      float64       `json:"error_rate"`
	LastUpdated    time.Time     `json:"last_updated"`
}

// HealthStatus represents the health status of a service
type HealthStatus struct {
	Status    string                 `json:"status"` // healthy, unhealthy, degraded
	LastCheck time.Time              `json:"last_check"`
	Checks    map[string]HealthCheck `json:"checks"`
	Message   string                 `json:"message"`
}

// HealthCheck represents a health check result
type HealthCheck struct {
	Name      string        `json:"name"`
	Status    string        `json:"status"` // pass, fail, warn
	Message   string        `json:"message"`
	LastCheck time.Time     `json:"last_check"`
	Duration  time.Duration `json:"duration"`
}

// LoadBalancingStrategy interface for different load balancing strategies
type LoadBalancingStrategy interface {
	SelectInstance(instances []*ServiceInstance) (*ServiceInstance, error)
	UpdateInstanceHealth(instance *ServiceInstance, healthy bool)
}

// CircuitBreaker represents a circuit breaker for service protection
type CircuitBreaker struct {
	config CircuitBreakerConfig
	state  string // closed, open, half_open
	mu     sync.RWMutex
}

// CircuitBreakerConfig holds configuration for the circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	Timeout          time.Duration `json:"timeout"`
	MaxRequests      int           `json:"max_requests"`
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

	// Initialize Python ML Service
	if arch.config.PythonMLServiceEnabled {
		arch.logger.Printf("🐍 Initializing Python ML Service")
		arch.pythonMLService = NewPythonMLService(arch.config.PythonMLServiceEndpoint, arch.logger)
		if err := arch.pythonMLService.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize Python ML Service: %w", err)
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
		arch.serviceDiscovery = NewServiceDiscovery(arch.logger)
		if err := arch.serviceDiscovery.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize Service Discovery: %w", err)
		}
	}

	// Initialize Load Balancer
	if arch.config.LoadBalancingEnabled {
		arch.logger.Printf("⚖️ Initializing Load Balancer")
		arch.loadBalancer = NewLoadBalancer(arch.logger)
		if err := arch.loadBalancer.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize Load Balancer: %w", err)
		}
	}

	// Initialize API Gateway
	if arch.config.APIGatewayEnabled {
		arch.logger.Printf("🌐 Initializing API Gateway")
		arch.apiGateway = NewAPIGateway(arch.config.APIGatewayEndpoint, arch.logger)
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
		arch.apiGateway.Stop()
	}

	// Stop Load Balancer
	if arch.loadBalancer != nil {
		arch.loadBalancer.Stop()
	}

	// Stop Service Discovery
	if arch.serviceDiscovery != nil {
		arch.serviceDiscovery.Stop()
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
		check, err := arch.apiGateway.HealthCheck(ctx)
		if err != nil {
			healthStatus.Checks["api_gateway"] = HealthCheck{
				Name:      "api_gateway",
				Status:    "fail",
				Message:   err.Error(),
				LastCheck: time.Now(),
			}
			healthStatus.Status = "unhealthy"
		} else {
			healthStatus.Checks["api_gateway"] = *check
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
