package infrastructure

import (
	"sync"
	"time"
)

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
	ModelID       string        `json:"model_id"`
	ModelVersion  string        `json:"model_version"`
	Accuracy      float64       `json:"accuracy"`
	Precision     float64       `json:"precision"`
	Recall        float64       `json:"recall"`
	F1Score       float64       `json:"f1_score"`
	InferenceTime time.Duration `json:"inference_time"`
	Throughput    int           `json:"throughput"` // requests per second
	RequestCount  int64         `json:"request_count"`
	SuccessCount  int64         `json:"success_count"`
	ErrorCount    int64         `json:"error_count"`
	LastUpdated   time.Time     `json:"last_updated"`
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

// ClassificationPrediction represents a single classification prediction
type ClassificationPrediction struct {
	Label       string  `json:"label"`
	Confidence  float64 `json:"confidence"`
	Probability float64 `json:"probability"`
	Rank        int     `json:"rank"`
}

// DetectedRisk represents a detected risk
type DetectedRisk struct {
	Category    string   `json:"category"` // illegal, prohibited, high_risk, tbml
	Severity    string   `json:"severity"` // low, medium, high, critical
	Confidence  float64  `json:"confidence"`
	Keywords    []string `json:"keywords"`
	Description string   `json:"description"`
}

// ClassificationRequest represents a request for classification
type ClassificationRequest struct {
	BusinessName        string  `json:"business_name"`
	Description         string  `json:"description"`
	WebsiteURL          string  `json:"website_url"`
	ModelType           string  `json:"model_type,omitempty"` // bert, distilbert, custom
	ModelVersion        string  `json:"model_version,omitempty"`
	MaxResults          int     `json:"max_results,omitempty"`
	ConfidenceThreshold float64 `json:"confidence_threshold,omitempty"`
}

// RiskDetectionRequest represents a request for risk detection
type RiskDetectionRequest struct {
	BusinessName   string   `json:"business_name"`
	Description    string   `json:"description"`
	WebsiteURL     string   `json:"website_url"`
	WebsiteContent string   `json:"website_content,omitempty"`
	ModelType      string   `json:"model_type,omitempty"`
	ModelVersion   string   `json:"model_version,omitempty"`
	RiskCategories []string `json:"risk_categories,omitempty"` // illegal, prohibited, high_risk, tbml
}

// ClassificationResponse represents a response from classification
type ClassificationResponse struct {
	RequestID       string                     `json:"request_id"`
	ModelID         string                     `json:"model_id"`
	ModelVersion    string                     `json:"model_version"`
	Classifications []ClassificationPrediction `json:"classifications"`
	Confidence      float64                    `json:"confidence"`
	ProcessingTime  time.Duration              `json:"processing_time"`
	Timestamp       time.Time                  `json:"timestamp"`
	Success         bool                       `json:"success"`
	Error           string                     `json:"error,omitempty"`
}

// RiskDetectionResponse represents a response from risk detection
type RiskDetectionResponse struct {
	RequestID      string         `json:"request_id"`
	ModelID        string         `json:"model_id"`
	ModelVersion   string         `json:"model_version"`
	RiskScore      float64        `json:"risk_score"`
	RiskLevel      string         `json:"risk_level"` // low, medium, high, critical
	DetectedRisks  []DetectedRisk `json:"detected_risks"`
	ProcessingTime time.Duration  `json:"processing_time"`
	Timestamp      time.Time      `json:"timestamp"`
	Success        bool           `json:"success"`
	Error          string         `json:"error,omitempty"`
}

// EnhancedClassificationRequest represents a request for enhanced classification
type EnhancedClassificationRequest struct {
	BusinessName     string `json:"business_name"`
	Description      string `json:"description,omitempty"`
	WebsiteURL       string `json:"website_url,omitempty"`
	WebsiteContent   string `json:"website_content,omitempty"`
	MaxResults       int    `json:"max_results,omitempty"`
	MaxContentLength int    `json:"max_content_length,omitempty"`
}

// EnhancedClassificationResponse represents an enhanced classification response
type EnhancedClassificationResponse struct {
	RequestID          string                  `json:"request_id"`
	ModelID            string                  `json:"model_id"`
	ModelVersion       string                  `json:"model_version"`
	Classifications    []ClassificationPrediction `json:"classifications"`
	Confidence         float64                 `json:"confidence"`
	Summary            string                  `json:"summary"`
	Explanation        string                  `json:"explanation"`
	ProcessingTime     float64                 `json:"processing_time"`
	QuantizationEnabled bool                   `json:"quantization_enabled"`
	Timestamp          time.Time               `json:"timestamp"`
	Success            bool                    `json:"success"`
	Error              string                  `json:"error,omitempty"`
}

// RuleEngineClassificationRequest represents a classification request for the rule engine
type RuleEngineClassificationRequest struct {
	BusinessName string `json:"business_name"`
	Description  string `json:"description"`
	WebsiteURL   string `json:"website_url"`
	MaxResults   int    `json:"max_results,omitempty"`
}

// RuleEngineRiskRequest represents a risk detection request for the rule engine
type RuleEngineRiskRequest struct {
	BusinessName   string   `json:"business_name"`
	Description    string   `json:"description"`
	WebsiteURL     string   `json:"website_url"`
	WebsiteContent string   `json:"website_content,omitempty"`
	RiskCategories []string `json:"risk_categories,omitempty"`
}

// RuleEngineClassificationResponse represents a classification response from the rule engine
type RuleEngineClassificationResponse struct {
	RequestID       string                     `json:"request_id"`
	Classifications []ClassificationPrediction `json:"classifications"`
	Confidence      float64                    `json:"confidence"`
	ProcessingTime  time.Duration              `json:"processing_time"`
	Timestamp       time.Time                  `json:"timestamp"`
	Success         bool                       `json:"success"`
	Error           string                     `json:"error,omitempty"`
	Method          string                     `json:"method"` // keyword_matching, mcc_lookup
}

// RuleEngineRiskResponse represents a risk detection response from the rule engine
type RuleEngineRiskResponse struct {
	RequestID      string         `json:"request_id"`
	RiskScore      float64        `json:"risk_score"`
	RiskLevel      string         `json:"risk_level"`
	DetectedRisks  []DetectedRisk `json:"detected_risks"`
	ProcessingTime time.Duration  `json:"processing_time"`
	Timestamp      time.Time      `json:"timestamp"`
	Success        bool           `json:"success"`
	Error          string         `json:"error,omitempty"`
	Method         string         `json:"method"` // keyword_matching, mcc_lookup, blacklist_check
}

// MCCCodeInfo represents MCC code information
type MCCCodeInfo struct {
	Code         string `json:"code"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	IsProhibited bool   `json:"is_prohibited"`
	RiskLevel    string `json:"risk_level"` // low, medium, high, critical
}

// BlacklistEntry represents a blacklist entry
type BlacklistEntry struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // business, domain, ip
	Value     string    `json:"value"`
	Reason    string    `json:"reason"`
	RiskLevel string    `json:"risk_level"` // low, medium, high, critical
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CachedClassificationResult represents a cached classification result
type CachedClassificationResult struct {
	Result      *RuleEngineClassificationResponse `json:"result"`
	CachedAt    time.Time                         `json:"cached_at"`
	ExpiresAt   time.Time                         `json:"expires_at"`
	AccessCount int                               `json:"access_count"`
}

// CachedRiskResult represents a cached risk detection result
type CachedRiskResult struct {
	Result      *RuleEngineRiskResponse `json:"result"`
	CachedAt    time.Time               `json:"cached_at"`
	ExpiresAt   time.Time               `json:"expires_at"`
	AccessCount int                     `json:"access_count"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled         bool          `json:"enabled"`
	MaxSize         int           `json:"max_size"`
	DefaultTTL      time.Duration `json:"default_ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	MaxAccessCount  int           `json:"max_access_count"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	ClassificationEntries int     `json:"classification_entries"`
	RiskEntries           int     `json:"risk_entries"`
	TotalEntries          int     `json:"total_entries"`
	MaxSize               int     `json:"max_size"`
	HitRate               float64 `json:"hit_rate"`
}
