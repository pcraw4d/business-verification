package microservices

import (
	"context"
	"time"
)

// ServiceContract defines the interface for service contracts
type ServiceContract interface {
	ServiceName() string
	Version() string
	Health() ServiceHealth
	Capabilities() []ServiceCapability
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ServiceCapability represents a capability provided by a service
type ServiceCapability struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Enabled     bool   `json:"enabled"`
}

// ServiceEndpoint represents a service endpoint
type ServiceEndpoint struct {
	Path        string            `json:"path"`
	Method      string            `json:"method"`
	Description string            `json:"description"`
	Parameters  []EndpointParam   `json:"parameters,omitempty"`
	Responses   []EndpointResponse `json:"responses,omitempty"`
}

// EndpointParam represents a parameter for a service endpoint
type EndpointParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     interface{} `json:"default,omitempty"`
}

// EndpointResponse represents a response from a service endpoint
type EndpointResponse struct {
	StatusCode  int    `json:"status_code"`
	Description string `json:"description"`
	Schema      string `json:"schema,omitempty"`
}

// ServiceRegistry manages service registration and discovery
type ServiceRegistry interface {
	Register(service ServiceContract) error
	Unregister(serviceName string) error
	GetService(serviceName string) (ServiceContract, error)
	ListServices() []ServiceContract
	GetHealthyServices() []ServiceContract
}

// ServiceDiscovery provides service discovery capabilities
type ServiceDiscovery interface {
	Discover(serviceName string) ([]ServiceInstance, error)
	DiscoverAll() map[string][]ServiceInstance
	Watch(serviceName string) (<-chan ServiceEvent, error)
	Unwatch(serviceName string) error
}

// ServiceInstance represents an instance of a service
type ServiceInstance struct {
	ID          string            `json:"id"`
	ServiceName string            `json:"service_name"`
	Version     string            `json:"version"`
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	Protocol    string            `json:"protocol"`
	Health      ServiceHealth     `json:"health"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	LastSeen    time.Time         `json:"last_seen"`
}

// ServiceEvent represents a service discovery event
type ServiceEvent struct {
	Type     ServiceEventType `json:"type"`
	Instance ServiceInstance  `json:"instance"`
	Timestamp time.Time       `json:"timestamp"`
}

// ServiceEventType represents the type of service event
type ServiceEventType string

const (
	ServiceEventAdded   ServiceEventType = "added"
	ServiceEventRemoved ServiceEventType = "removed"
	ServiceEventUpdated ServiceEventType = "updated"
)

// ServiceClient provides a client interface for service communication
type ServiceClient interface {
	Call(ctx context.Context, serviceName, method string, request interface{}) (interface{}, error)
	CallAsync(ctx context.Context, serviceName, method string, request interface{}) (<-chan ServiceResponse, error)
	Health(ctx context.Context, serviceName string) (ServiceHealth, error)
}

// ServiceResponse represents a response from a service call
type ServiceResponse struct {
	Data    interface{} `json:"data"`
	Error   error       `json:"error,omitempty"`
	Latency time.Duration `json:"latency"`
}

// ServiceLoadBalancer provides load balancing capabilities
type ServiceLoadBalancer interface {
	Select(serviceName string) (ServiceInstance, error)
	UpdateHealth(instanceID string, health ServiceHealth) error
	GetInstances(serviceName string) ([]ServiceInstance, error)
}

// ServiceCircuitBreaker provides circuit breaker functionality
type ServiceCircuitBreaker interface {
	Execute(ctx context.Context, serviceName, method string, request interface{}) (interface{}, error)
	GetState(serviceName string) CircuitBreakerState
	Reset(serviceName string) error
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState struct {
	ServiceName    string        `json:"service_name"`
	State          string        `json:"state"`
	FailureCount   int64         `json:"failure_count"`
	SuccessCount   int64         `json:"success_count"`
	LastFailure    *time.Time    `json:"last_failure,omitempty"`
	LastSuccess    *time.Time    `json:"last_success,omitempty"`
	Threshold      int64         `json:"threshold"`
	Timeout        time.Duration `json:"timeout"`
	LastStateChange time.Time    `json:"last_state_change"`
}

// ServiceMetrics provides metrics for service monitoring
type ServiceMetrics interface {
	RecordRequest(serviceName, method string, duration time.Duration, success bool)
	RecordLatency(serviceName, method string, latency time.Duration)
	RecordError(serviceName, method string, errorType string)
	GetMetrics(serviceName string) ServiceMetricsData
}

// ServiceMetricsData represents metrics data for a service
type ServiceMetricsData struct {
	ServiceName     string                 `json:"service_name"`
	RequestCount    int64                  `json:"request_count"`
	SuccessCount    int64                  `json:"success_count"`
	ErrorCount      int64                  `json:"error_count"`
	AverageLatency  time.Duration          `json:"average_latency"`
	P95Latency      time.Duration          `json:"p95_latency"`
	P99Latency      time.Duration          `json:"p99_latency"`
	ErrorRate       float64                `json:"error_rate"`
	SuccessRate     float64                `json:"success_rate"`
	MethodMetrics   map[string]MethodMetrics `json:"method_metrics,omitempty"`
	LastUpdated     time.Time              `json:"last_updated"`
}

// MethodMetrics represents metrics for a specific method
type MethodMetrics struct {
	Method         string        `json:"method"`
	RequestCount   int64         `json:"request_count"`
	SuccessCount   int64         `json:"success_count"`
	ErrorCount     int64         `json:"error_count"`
	AverageLatency time.Duration `json:"average_latency"`
	P95Latency     time.Duration `json:"p95_latency"`
	P99Latency     time.Duration `json:"p99_latency"`
	ErrorRate      float64       `json:"error_rate"`
	SuccessRate    float64       `json:"success_rate"`
}

// ServiceConfiguration provides configuration management for services
type ServiceConfiguration interface {
	Get(serviceName, key string) (interface{}, error)
	Set(serviceName, key string, value interface{}) error
	GetAll(serviceName string) (map[string]interface{}, error)
	Watch(serviceName, key string) (<-chan ConfigChange, error)
	Unwatch(serviceName, key string) error
}

// ConfigChange represents a configuration change
type ConfigChange struct {
	ServiceName string      `json:"service_name"`
	Key         string      `json:"key"`
	OldValue    interface{} `json:"old_value,omitempty"`
	NewValue    interface{} `json:"new_value"`
	Timestamp   time.Time   `json:"timestamp"`
}

// ServiceSecurity provides security capabilities for services
type ServiceSecurity interface {
	Authenticate(ctx context.Context, credentials interface{}) (ServiceToken, error)
	Authorize(ctx context.Context, token ServiceToken, resource, action string) (bool, error)
	ValidateToken(ctx context.Context, token ServiceToken) (bool, error)
	RefreshToken(ctx context.Context, token ServiceToken) (ServiceToken, error)
}

// ServiceToken represents a service authentication token
type ServiceToken struct {
	Token     string            `json:"token"`
	Type      string            `json:"type"`
	ExpiresAt time.Time         `json:"expires_at"`
	Claims    map[string]interface{} `json:"claims,omitempty"`
}

// ServiceRetry provides retry capabilities for service calls
type ServiceRetry interface {
	ExecuteWithRetry(ctx context.Context, serviceName, method string, request interface{}, retryConfig RetryConfig) (interface{}, error)
	GetRetryStats(serviceName string) RetryStats
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	MaxAttempts     int           `json:"max_attempts"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors,omitempty"`
}

// RetryStats represents retry statistics
type RetryStats struct {
	ServiceName    string `json:"service_name"`
	TotalAttempts  int64  `json:"total_attempts"`
	SuccessfulRetries int64 `json:"successful_retries"`
	FailedRetries  int64  `json:"failed_retries"`
	AverageAttempts float64 `json:"average_attempts"`
}

// ServiceTimeout provides timeout management for service calls
type ServiceTimeout interface {
	WithTimeout(ctx context.Context, timeout time.Duration) context.Context
	GetDefaultTimeout(serviceName string) time.Duration
	SetDefaultTimeout(serviceName string, timeout time.Duration) error
}

// ServiceRateLimiter provides rate limiting capabilities
type ServiceRateLimiter interface {
	Allow(serviceName string) bool
	AllowN(serviceName string, n int) bool
	GetRateLimit(serviceName string) RateLimit
	SetRateLimit(serviceName string, limit RateLimit) error
}

// RateLimit represents rate limiting configuration
type RateLimit struct {
	RequestsPerSecond float64 `json:"requests_per_second"`
	BurstSize         int     `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
}

// ServiceFaultTolerance provides fault tolerance capabilities
type ServiceFaultTolerance interface {
	ExecuteWithFallback(ctx context.Context, serviceName, method string, request interface{}, fallback FallbackStrategy) (interface{}, error)
	GetFaultToleranceStats(serviceName string) FaultToleranceStats
}

// FallbackStrategy represents a fallback strategy
type FallbackStrategy struct {
	Type        string      `json:"type"`
	FallbackData interface{} `json:"fallback_data,omitempty"`
	FallbackService string   `json:"fallback_service,omitempty"`
	FallbackMethod string    `json:"fallback_method,omitempty"`
}

// FaultToleranceStats represents fault tolerance statistics
type FaultToleranceStats struct {
	ServiceName      string `json:"service_name"`
	TotalCalls       int64  `json:"total_calls"`
	FallbackCalls    int64  `json:"fallback_calls"`
	FallbackRate     float64 `json:"fallback_rate"`
	SuccessfulFallbacks int64 `json:"successful_fallbacks"`
	FailedFallbacks  int64  `json:"failed_fallbacks"`
}
