package intelligent_routing

import (
	"context"
	"time"
)

// RequestType represents the type of verification request
type RequestType string

const (
	RequestTypeBasic      RequestType = "basic"
	RequestTypeEnhanced   RequestType = "enhanced"
	RequestTypeCompliance RequestType = "compliance"
	RequestTypeRisk       RequestType = "risk"
	RequestTypeCustom     RequestType = "custom"
)

// RequestPriority represents the priority level of a request
type RequestPriority string

const (
	PriorityLow    RequestPriority = "low"
	PriorityNormal RequestPriority = "normal"
	PriorityHigh   RequestPriority = "high"
	PriorityUrgent RequestPriority = "urgent"
)

// RequestComplexity represents the complexity level of a request
type RequestComplexity string

const (
	ComplexitySimple   RequestComplexity = "simple"
	ComplexityModerate RequestComplexity = "moderate"
	ComplexityComplex  RequestComplexity = "complex"
	ComplexityAdvanced RequestComplexity = "advanced"
)

// VerificationRequest represents a business verification request
type VerificationRequest struct {
	ID              string            `json:"id"`
	BusinessName    string            `json:"business_name"`
	BusinessAddress string            `json:"business_address"`
	Industry        string            `json:"industry,omitempty"`
	RequestType     RequestType       `json:"request_type"`
	Priority        RequestPriority   `json:"priority"`
	Complexity      RequestComplexity `json:"complexity"`
	UserID          string            `json:"user_id"`
	ClientID        string            `json:"client_id"`
	CreatedAt       time.Time         `json:"created_at"`
	Deadline        *time.Time        `json:"deadline,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// ModuleCapability represents the capabilities of a verification module
type ModuleCapability struct {
	ModuleID       string             `json:"module_id"`
	ModuleName     string             `json:"module_name"`
	Capabilities   []string           `json:"capabilities"`
	RequestTypes   []RequestType      `json:"request_types"`
	Complexity     RequestComplexity  `json:"complexity"`
	Performance    ModulePerformance  `json:"performance"`
	Availability   ModuleAvailability `json:"availability"`
	Specialization map[string]float64 `json:"specialization"` // Industry/type specialization scores
}

// ModulePerformance represents the performance metrics of a module
type ModulePerformance struct {
	SuccessRate    float64   `json:"success_rate"`
	AverageLatency float64   `json:"average_latency"`
	Throughput     float64   `json:"throughput"`
	ErrorRate      float64   `json:"error_rate"`
	LastUpdated    time.Time `json:"last_updated"`
}

// ModuleAvailability represents the availability status of a module
type ModuleAvailability struct {
	IsAvailable     bool      `json:"is_available"`
	LastHealthCheck time.Time `json:"last_health_check"`
	HealthScore     float64   `json:"health_score"`
	LoadPercentage  float64   `json:"load_percentage"`
	QueueLength     int       `json:"queue_length"`
}

// RoutingDecision represents a routing decision made by the system
type RoutingDecision struct {
	ID              string                 `json:"id"`
	RequestID       string                 `json:"request_id"`
	SelectedModules []string               `json:"selected_modules"`
	DecisionReason  string                 `json:"decision_reason"`
	Confidence      float64                `json:"confidence"`
	RoutingStrategy RoutingStrategy        `json:"routing_strategy"`
	CreatedAt       time.Time              `json:"created_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// RoutingStrategy represents the strategy used for routing
type RoutingStrategy string

const (
	StrategySingleModule    RoutingStrategy = "single_module"
	StrategyParallelModules RoutingStrategy = "parallel_modules"
	StrategyFallback        RoutingStrategy = "fallback"
	StrategyLoadBalanced    RoutingStrategy = "load_balanced"
	StrategyOptimized       RoutingStrategy = "optimized"
)

// ProcessingResult represents the result of processing a request
type ProcessingResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	ModuleID       string                 `json:"module_id"`
	Status         ProcessingStatus       `json:"status"`
	Result         map[string]interface{} `json:"result,omitempty"`
	Error          string                 `json:"error,omitempty"`
	ProcessingTime time.Duration          `json:"processing_time"`
	CompletedAt    time.Time              `json:"completed_at"`
}

// ProcessingStatus represents the status of request processing
type ProcessingStatus string

const (
	StatusPending    ProcessingStatus = "pending"
	StatusProcessing ProcessingStatus = "processing"
	StatusCompleted  ProcessingStatus = "completed"
	StatusFailed     ProcessingStatus = "failed"
	StatusCancelled  ProcessingStatus = "cancelled"
)

// RoutingMetrics represents metrics about routing decisions and performance
type RoutingMetrics struct {
	TotalRequests    int64              `json:"total_requests"`
	SuccessfulRoutes int64              `json:"successful_routes"`
	FailedRoutes     int64              `json:"failed_routes"`
	AverageLatency   float64            `json:"average_latency"`
	SuccessRate      float64            `json:"success_rate"`
	LoadDistribution map[string]float64 `json:"load_distribution"`
	LastUpdated      time.Time          `json:"last_updated"`
}

// FallbackStrategy represents a fallback strategy for failed modules
type FallbackStrategy struct {
	ID              string   `json:"id"`
	PrimaryModules  []string `json:"primary_modules"`
	FallbackModules []string `json:"fallback_modules"`
	Conditions      []string `json:"conditions"`
	Priority        int      `json:"priority"`
}

// DataSource represents an external data source for verification
type DataSource struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	QualityScore float64 `json:"quality_score"`
	Reliability  float64 `json:"reliability"`
	Latency      float64 `json:"latency"`
	Cost         float64 `json:"cost"`
	IsAvailable  bool    `json:"is_available"`
}

// RoutingConfig represents the configuration for the routing system
type RoutingConfig struct {
	DefaultStrategy      RoutingStrategy `json:"default_strategy"`
	LoadBalancingEnabled bool            `json:"load_balancing_enabled"`
	ParallelProcessing   bool            `json:"parallel_processing"`
	FallbackEnabled      bool            `json:"fallback_enabled"`
	HealthCheckInterval  time.Duration   `json:"health_check_interval"`
	DecisionTimeout      time.Duration   `json:"decision_timeout"`
	MaxRetries           int             `json:"max_retries"`
	CacheEnabled         bool            `json:"cache_enabled"`
	CacheTTL             time.Duration   `json:"cache_ttl"`
}

// RoutingService defines the interface for the intelligent routing service
type RoutingService interface {
	// Core routing functionality
	RouteRequest(ctx context.Context, request *VerificationRequest) (*RoutingDecision, error)
	ProcessRequest(ctx context.Context, request *VerificationRequest) ([]*ProcessingResult, error)

	// Module management
	RegisterModule(ctx context.Context, capability *ModuleCapability) error
	UnregisterModule(ctx context.Context, moduleID string) error
	GetModuleCapabilities(ctx context.Context) ([]*ModuleCapability, error)

	// Health and monitoring
	CheckModuleHealth(ctx context.Context, moduleID string) (*ModuleAvailability, error)
	GetRoutingMetrics(ctx context.Context) (*RoutingMetrics, error)

	// Configuration
	UpdateConfig(ctx context.Context, config *RoutingConfig) error
	GetConfig(ctx context.Context) (*RoutingConfig, error)
}

// RequestAnalyzer defines the interface for request analysis
type RequestAnalyzer interface {
	AnalyzeRequest(ctx context.Context, request *VerificationRequest) (*RequestAnalysis, error)
	ClassifyRequest(ctx context.Context, request *VerificationRequest) (*RequestClassification, error)
	AssessComplexity(ctx context.Context, request *VerificationRequest) (RequestComplexity, error)
	DeterminePriority(ctx context.Context, request *VerificationRequest) (RequestPriority, error)
}

// RequestAnalysis represents the analysis of a verification request
type RequestAnalysis struct {
	RequestID      string                 `json:"request_id"`
	AnalysisID     string                 `json:"analysis_id"`
	Classification *RequestClassification `json:"classification"`
	Complexity     RequestComplexity      `json:"complexity"`
	Priority       RequestPriority        `json:"priority"`
	ResourceNeeds  ResourceNeeds          `json:"resource_needs"`
	RiskFactors    []string               `json:"risk_factors"`
	Confidence     float64                `json:"confidence"`
	CreatedAt      time.Time              `json:"created_at"`
}

// RequestClassification represents the classification of a request
type RequestClassification struct {
	RequestType      RequestType `json:"request_type"`
	Industry         string      `json:"industry"`
	GeographicRegion string      `json:"geographic_region"`
	BusinessSize     string      `json:"business_size"`
	ComplianceLevel  string      `json:"compliance_level"`
	RiskLevel        string      `json:"risk_level"`
	Confidence       float64     `json:"confidence"`
}

// ResourceNeeds represents the resource requirements for processing a request
type ResourceNeeds struct {
	CPUUsage       float64       `json:"cpu_usage"`
	MemoryUsage    float64       `json:"memory_usage"`
	NetworkUsage   float64       `json:"network_usage"`
	ProcessingTime time.Duration `json:"processing_time"`
	Concurrency    int           `json:"concurrency"`
}

// ModuleSelector defines the interface for module selection
type ModuleSelector interface {
	SelectModules(ctx context.Context, request *VerificationRequest, analysis *RequestAnalysis) ([]*ModuleCapability, error)
	OptimizeSelection(ctx context.Context, candidates []*ModuleCapability, request *VerificationRequest) ([]*ModuleCapability, error)
	LoadBalance(ctx context.Context, modules []*ModuleCapability) ([]*ModuleCapability, error)
	RegisterModule(module *ModuleCapability) error
	UnregisterModule(moduleID string) error
	GetModuleCapabilities() []*ModuleCapability
}

// LoadBalancer defines the interface for load balancing
type LoadBalancer interface {
	DistributeLoad(ctx context.Context, modules []*ModuleCapability, request *VerificationRequest) (map[string]float64, error)
	GetModuleLoad(ctx context.Context, moduleID string) (float64, error)
	UpdateModuleLoad(ctx context.Context, moduleID string, load float64) error
}

// HealthChecker defines the interface for health checking
type HealthChecker interface {
	CheckHealth(ctx context.Context, moduleID string) (*ModuleAvailability, error)
	CheckAllModules(ctx context.Context) (map[string]*ModuleAvailability, error)
	RegisterHealthCallback(ctx context.Context, moduleID string, callback func(*ModuleAvailability)) error
}

// MetricsCollector defines the interface for metrics collection
type MetricsCollector interface {
	RecordRoutingDecision(ctx context.Context, decision *RoutingDecision) error
	RecordProcessingResult(ctx context.Context, result *ProcessingResult) error
	GetMetrics(ctx context.Context) (*RoutingMetrics, error)
	ResetMetrics(ctx context.Context) error
}
