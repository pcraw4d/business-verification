package routing

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
)

// IntelligentRouter provides intelligent routing based on feature flags and performance metrics
type IntelligentRouter struct {
	// Feature flag manager
	featureFlagManager *config.GranularFeatureFlagManager

	// Service endpoints
	endpoints map[string]*ServiceEndpoint

	// Performance metrics
	metrics *RoutingMetrics

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger

	// Configuration
	config IntelligentRouterConfig
}

// IntelligentRouterConfig holds configuration for the intelligent router
type IntelligentRouterConfig struct {
	// Routing strategy
	RoutingStrategy string `json:"routing_strategy"` // performance_based, feature_flag_based, hybrid

	// Performance thresholds
	MaxLatencyThreshold     time.Duration `json:"max_latency_threshold"`
	MaxErrorRateThreshold   float64       `json:"max_error_rate_threshold"`
	MinSuccessRateThreshold float64       `json:"min_success_rate_threshold"`

	// Fallback configuration
	FallbackEnabled     bool   `json:"fallback_enabled"`
	FallbackEndpoint    string `json:"fallback_endpoint"`
	MaxFallbackAttempts int    `json:"max_fallback_attempts"`

	// Circuit breaker configuration
	CircuitBreakerEnabled   bool          `json:"circuit_breaker_enabled"`
	CircuitBreakerTimeout   time.Duration `json:"circuit_breaker_timeout"`
	CircuitBreakerThreshold int           `json:"circuit_breaker_threshold"`

	// Load balancing
	LoadBalancingEnabled  bool   `json:"load_balancing_enabled"`
	LoadBalancingStrategy string `json:"load_balancing_strategy"` // round_robin, weighted, least_connections

	// Monitoring
	MetricsEnabled      bool `json:"metrics_enabled"`
	PerformanceTracking bool `json:"performance_tracking"`
}

// ServiceEndpoint represents a service endpoint
type ServiceEndpoint struct {
	// Endpoint information
	Name      string `json:"name"`
	URL       string `json:"url"`
	Type      string `json:"type"`       // python_ml_service, go_rule_engine, api_gateway
	ModelType string `json:"model_type"` // bert, distilbert, custom, rule_based

	// Performance metrics
	Metrics *EndpointMetrics `json:"metrics"`

	// Health status
	IsHealthy       bool      `json:"is_healthy"`
	LastHealthCheck time.Time `json:"last_health_check"`

	// Circuit breaker
	CircuitBreaker *CircuitBreaker `json:"circuit_breaker"`

	// Load balancing
	Weight            int `json:"weight"`
	ActiveConnections int `json:"active_connections"`
}

// EndpointMetrics holds performance metrics for an endpoint
type EndpointMetrics struct {
	// Request metrics
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency"`
	P95Latency         time.Duration `json:"p95_latency"`
	P99Latency         time.Duration `json:"p99_latency"`

	// Error metrics
	ErrorRate     float64   `json:"error_rate"`
	SuccessRate   float64   `json:"success_rate"`
	LastErrorTime time.Time `json:"last_error_time"`

	// Performance metrics
	Throughput         float64 `json:"throughput"` // requests per second
	ConcurrentRequests int     `json:"concurrent_requests"`

	// Model-specific metrics
	ModelAccuracy   float64       `json:"model_accuracy"`
	ModelConfidence float64       `json:"model_confidence"`
	ModelLatency    time.Duration `json:"model_latency"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`
}

// CircuitBreaker implements circuit breaker pattern for service endpoints
type CircuitBreaker struct {
	// Circuit breaker state
	State string `json:"state"` // closed, open, half_open

	// Configuration
	FailureThreshold int           `json:"failure_threshold"`
	Timeout          time.Duration `json:"timeout"`
	SuccessThreshold int           `json:"success_threshold"`

	// Current state
	FailureCount    int       `json:"failure_count"`
	SuccessCount    int       `json:"success_count"`
	LastFailureTime time.Time `json:"last_failure_time"`
	LastSuccessTime time.Time `json:"last_success_time"`

	// Thread safety
	mu sync.RWMutex
}

// RoutingMetrics holds routing performance metrics
type RoutingMetrics struct {
	// Routing decisions
	TotalRoutingDecisions int64 `json:"total_routing_decisions"`
	SuccessfulRoutings    int64 `json:"successful_routings"`
	FailedRoutings        int64 `json:"failed_routings"`

	// Performance metrics
	AverageRoutingLatency time.Duration `json:"average_routing_latency"`
	RoutingErrorRate      float64       `json:"routing_error_rate"`

	// Feature flag usage
	FeatureFlagDecisions map[string]int64 `json:"feature_flag_decisions"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`
}

// ClassificationRequest represents a classification request
type ClassificationRequest struct {
	// Request information
	RequestID    string `json:"request_id"`
	BusinessName string `json:"business_name"`
	Description  string `json:"description"`
	WebsiteURL   string `json:"website_url"`
	RequestType  string `json:"request_type"` // classification, risk_detection

	// Request metadata
	UserID    string            `json:"user_id"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata"`

	// Performance requirements
	MaxLatency  time.Duration `json:"max_latency"`
	MinAccuracy float64       `json:"min_accuracy"`
	Priority    string        `json:"priority"` // low, medium, high, critical
}

// ClassificationResponse represents a classification response
type ClassificationResponse struct {
	// Response information
	RequestID      string        `json:"request_id"`
	ModelUsed      string        `json:"model_used"`
	ServiceUsed    string        `json:"service_used"`
	ProcessingTime time.Duration `json:"processing_time"`

	// Classification results (defined in handlers package)
	Classification interface{} `json:"classification"`
	RiskAssessment interface{} `json:"risk_assessment"`

	// Performance metrics
	Latency    time.Duration `json:"latency"`
	Accuracy   float64       `json:"accuracy"`
	Confidence float64       `json:"confidence"`

	// Metadata
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// NewIntelligentRouter creates a new intelligent router
func NewIntelligentRouter(
	featureFlagManager *config.GranularFeatureFlagManager,
	config IntelligentRouterConfig,
	logger *log.Logger,
) *IntelligentRouter {
	if logger == nil {
		logger = log.Default()
	}

	router := &IntelligentRouter{
		featureFlagManager: featureFlagManager,
		endpoints:          make(map[string]*ServiceEndpoint),
		metrics: &RoutingMetrics{
			FeatureFlagDecisions: make(map[string]int64),
		},
		logger: logger,
		config: config,
	}

	// Initialize default endpoints
	router.initializeDefaultEndpoints()

	// Start background processes
	go router.startBackgroundProcesses()

	return router
}

// RouteRequest routes a classification request to the optimal endpoint
func (ir *IntelligentRouter) RouteRequest(ctx context.Context, req *ClassificationRequest) (*ServiceEndpoint, error) {
	startTime := time.Now()
	defer func() {
		ir.updateRoutingMetrics(time.Since(startTime), true)
	}()

	// Get feature flags
	flags := ir.featureFlagManager.GetFlags()

	// Determine optimal model based on feature flags and performance
	optimalModel, err := ir.determineOptimalModel(ctx, req, flags)
	if err != nil {
		ir.logger.Printf("Failed to determine optimal model: %v", err)
		return ir.getFallbackEndpoint(), nil
	}

	// Find endpoint for the optimal model
	endpoint, exists := ir.endpoints[optimalModel]
	if !exists {
		ir.logger.Printf("Endpoint not found for model: %s", optimalModel)
		return ir.getFallbackEndpoint(), nil
	}

	// Check if endpoint is healthy
	if !ir.isEndpointHealthy(endpoint) {
		ir.logger.Printf("Endpoint is not healthy: %s", endpoint.Name)
		return ir.getFallbackEndpoint(), nil
	}

	// Check circuit breaker
	if ir.isCircuitBreakerOpen(endpoint) {
		ir.logger.Printf("Circuit breaker is open for endpoint: %s", endpoint.Name)
		return ir.getFallbackEndpoint(), nil
	}

	// Update routing metrics
	ir.updateFeatureFlagMetrics(optimalModel)

	ir.logger.Printf("Routed request %s to endpoint: %s (model: %s)", req.RequestID, endpoint.Name, optimalModel)
	return endpoint, nil
}

// determineOptimalModel determines the optimal model based on feature flags and performance
func (ir *IntelligentRouter) determineOptimalModel(ctx context.Context, req *ClassificationRequest, flags *config.GranularFeatureFlags) (string, error) {
	// Check if A/B testing is enabled
	if flags.ABTesting.Enabled {
		// Use A/B testing to determine model
		testVariant, err := ir.featureFlagManager.GetOptimalModel(ctx, req.RequestType)
		if err == nil {
			return testVariant, nil
		}
	}

	// Check if gradual rollout is enabled
	if flags.Rollout.GradualRolloutEnabled {
		// Use rollout manager to determine if request should use new model
		optimalModel, err := ir.featureFlagManager.GetOptimalModel(ctx, req.RequestType)
		if err == nil {
			return optimalModel, nil
		}
	}

	// Use performance-based routing
	if ir.config.RoutingStrategy == "performance_based" {
		return ir.getBestPerformingModel(req.RequestType), nil
	}

	// Use feature flag-based routing
	if ir.config.RoutingStrategy == "feature_flag_based" {
		return ir.getFeatureFlagBasedModel(req.RequestType, flags), nil
	}

	// Use hybrid routing (default)
	return ir.getHybridModel(ctx, req, flags), nil
}

// getBestPerformingModel returns the best performing model for the request type
func (ir *IntelligentRouter) getBestPerformingModel(requestType string) string {
	ir.mu.RLock()
	defer ir.mu.RUnlock()

	var bestModel string
	var bestScore float64

	for name, endpoint := range ir.endpoints {
		if endpoint.Type == "python_ml_service" && requestType == "classification" {
			score := ir.calculatePerformanceScore(endpoint)
			if score > bestScore {
				bestScore = score
				bestModel = name
			}
		}
	}

	if bestModel == "" {
		return "rule_based"
	}

	return bestModel
}

// getFeatureFlagBasedModel returns the model based on feature flags
func (ir *IntelligentRouter) getFeatureFlagBasedModel(requestType string, flags *config.GranularFeatureFlags) string {
	switch requestType {
	case "classification":
		if flags.Models.BERTClassificationEnabled {
			return "bert_classification"
		}
		if flags.Models.DistilBERTClassificationEnabled {
			return "distilbert_classification"
		}
		if flags.Models.CustomNeuralNetEnabled {
			return "custom_neural_net"
		}
	case "risk_detection":
		if flags.Models.BERTRiskDetectionEnabled {
			return "bert_risk_detection"
		}
		if flags.Models.AnomalyDetectionEnabled {
			return "anomaly_detection"
		}
		if flags.Models.PatternRecognitionEnabled {
			return "pattern_recognition"
		}
	}

	// Fallback to rule-based system
	return "rule_based"
}

// getHybridModel returns the model using hybrid routing strategy
func (ir *IntelligentRouter) getHybridModel(ctx context.Context, req *ClassificationRequest, flags *config.GranularFeatureFlags) string {
	// First check feature flags
	featureFlagModel := ir.getFeatureFlagBasedModel(req.RequestType, flags)

	// Then check performance
	performanceModel := ir.getBestPerformingModel(req.RequestType)

	// If both are the same, use that model
	if featureFlagModel == performanceModel {
		return featureFlagModel
	}

	// If different, use performance-based model if it meets requirements
	if ir.meetsPerformanceRequirements(performanceModel, req) {
		return performanceModel
	}

	// Otherwise use feature flag-based model
	return featureFlagModel
}

// calculatePerformanceScore calculates a performance score for an endpoint
func (ir *IntelligentRouter) calculatePerformanceScore(endpoint *ServiceEndpoint) float64 {
	if endpoint.Metrics == nil {
		return 0.0
	}

	// Calculate score based on multiple factors
	latencyScore := 1.0 - float64(endpoint.Metrics.AverageLatency)/float64(ir.config.MaxLatencyThreshold)
	accuracyScore := endpoint.Metrics.ModelAccuracy
	successRateScore := endpoint.Metrics.SuccessRate

	// Weighted average
	score := (latencyScore*0.3 + accuracyScore*0.4 + successRateScore*0.3)
	return score
}

// meetsPerformanceRequirements checks if a model meets performance requirements
func (ir *IntelligentRouter) meetsPerformanceRequirements(modelName string, req *ClassificationRequest) bool {
	endpoint, exists := ir.endpoints[modelName]
	if !exists {
		return false
	}

	if endpoint.Metrics == nil {
		return false
	}

	// Check latency requirements
	if req.MaxLatency > 0 && endpoint.Metrics.AverageLatency > req.MaxLatency {
		return false
	}

	// Check accuracy requirements
	if req.MinAccuracy > 0 && endpoint.Metrics.ModelAccuracy < req.MinAccuracy {
		return false
	}

	return true
}

// isEndpointHealthy checks if an endpoint is healthy
func (ir *IntelligentRouter) isEndpointHealthy(endpoint *ServiceEndpoint) bool {
	if !endpoint.IsHealthy {
		return false
	}

	// Check if health check is recent
	if time.Since(endpoint.LastHealthCheck) > time.Minute*5 {
		return false
	}

	return true
}

// isCircuitBreakerOpen checks if the circuit breaker is open for an endpoint
func (ir *IntelligentRouter) isCircuitBreakerOpen(endpoint *ServiceEndpoint) bool {
	if endpoint.CircuitBreaker == nil {
		return false
	}

	endpoint.CircuitBreaker.mu.RLock()
	defer endpoint.CircuitBreaker.mu.RUnlock()

	return endpoint.CircuitBreaker.State == "open"
}

// getFallbackEndpoint returns the fallback endpoint
func (ir *IntelligentRouter) getFallbackEndpoint() *ServiceEndpoint {
	// Return rule-based endpoint as fallback
	if endpoint, exists := ir.endpoints["rule_based"]; exists {
		return endpoint
	}

	// If no rule-based endpoint, return the first available endpoint
	for _, endpoint := range ir.endpoints {
		if endpoint.IsHealthy {
			return endpoint
		}
	}

	// Return nil if no endpoints are available
	return nil
}

// initializeDefaultEndpoints initializes default service endpoints
func (ir *IntelligentRouter) initializeDefaultEndpoints() {
	// Python ML Service endpoints
	ir.endpoints["bert_classification"] = &ServiceEndpoint{
		Name:            "BERT Classification",
		URL:             "http://python-ml-service:8000/classify/bert",
		Type:            "python_ml_service",
		ModelType:       "bert",
		Metrics:         &EndpointMetrics{},
		IsHealthy:       true,
		LastHealthCheck: time.Now(),
		CircuitBreaker: &CircuitBreaker{
			State:            "closed",
			FailureThreshold: 5,
			Timeout:          time.Minute * 5,
			SuccessThreshold: 3,
		},
		Weight: 100,
	}

	ir.endpoints["distilbert_classification"] = &ServiceEndpoint{
		Name:            "DistilBERT Classification",
		URL:             "http://python-ml-service:8000/classify/distilbert",
		Type:            "python_ml_service",
		ModelType:       "distilbert",
		Metrics:         &EndpointMetrics{},
		IsHealthy:       true,
		LastHealthCheck: time.Now(),
		CircuitBreaker: &CircuitBreaker{
			State:            "closed",
			FailureThreshold: 5,
			Timeout:          time.Minute * 5,
			SuccessThreshold: 3,
		},
		Weight: 80,
	}

	ir.endpoints["custom_neural_net"] = &ServiceEndpoint{
		Name:            "Custom Neural Network",
		URL:             "http://python-ml-service:8000/classify/custom",
		Type:            "python_ml_service",
		ModelType:       "custom",
		Metrics:         &EndpointMetrics{},
		IsHealthy:       true,
		LastHealthCheck: time.Now(),
		CircuitBreaker: &CircuitBreaker{
			State:            "closed",
			FailureThreshold: 5,
			Timeout:          time.Minute * 5,
			SuccessThreshold: 3,
		},
		Weight: 60,
	}

	// Risk detection endpoints
	ir.endpoints["bert_risk_detection"] = &ServiceEndpoint{
		Name:            "BERT Risk Detection",
		URL:             "http://python-ml-service:8000/risk/bert",
		Type:            "python_ml_service",
		ModelType:       "bert",
		Metrics:         &EndpointMetrics{},
		IsHealthy:       true,
		LastHealthCheck: time.Now(),
		CircuitBreaker: &CircuitBreaker{
			State:            "closed",
			FailureThreshold: 5,
			Timeout:          time.Minute * 5,
			SuccessThreshold: 3,
		},
		Weight: 90,
	}

	ir.endpoints["anomaly_detection"] = &ServiceEndpoint{
		Name:            "Anomaly Detection",
		URL:             "http://python-ml-service:8000/risk/anomaly",
		Type:            "python_ml_service",
		ModelType:       "anomaly",
		Metrics:         &EndpointMetrics{},
		IsHealthy:       true,
		LastHealthCheck: time.Now(),
		CircuitBreaker: &CircuitBreaker{
			State:            "closed",
			FailureThreshold: 5,
			Timeout:          time.Minute * 5,
			SuccessThreshold: 3,
		},
		Weight: 70,
	}

	// Rule-based endpoints
	ir.endpoints["rule_based"] = &ServiceEndpoint{
		Name:            "Rule-based System",
		URL:             "http://go-rule-engine:8080/classify",
		Type:            "go_rule_engine",
		ModelType:       "rule_based",
		Metrics:         &EndpointMetrics{},
		IsHealthy:       true,
		LastHealthCheck: time.Now(),
		CircuitBreaker: &CircuitBreaker{
			State:            "closed",
			FailureThreshold: 10,
			Timeout:          time.Minute * 2,
			SuccessThreshold: 5,
		},
		Weight: 50,
	}
}

// updateRoutingMetrics updates routing performance metrics
func (ir *IntelligentRouter) updateRoutingMetrics(latency time.Duration, success bool) {
	ir.mu.Lock()
	defer ir.mu.Unlock()

	ir.metrics.TotalRoutingDecisions++
	if success {
		ir.metrics.SuccessfulRoutings++
	} else {
		ir.metrics.FailedRoutings++
	}

	// Update average latency
	if ir.metrics.TotalRoutingDecisions == 1 {
		ir.metrics.AverageRoutingLatency = latency
	} else {
		ir.metrics.AverageRoutingLatency = time.Duration(
			(float64(ir.metrics.AverageRoutingLatency)*float64(ir.metrics.TotalRoutingDecisions-1) + float64(latency)) / float64(ir.metrics.TotalRoutingDecisions),
		)
	}

	// Update error rate
	ir.metrics.RoutingErrorRate = float64(ir.metrics.FailedRoutings) / float64(ir.metrics.TotalRoutingDecisions)
	ir.metrics.LastUpdated = time.Now()
}

// updateFeatureFlagMetrics updates feature flag usage metrics
func (ir *IntelligentRouter) updateFeatureFlagMetrics(modelName string) {
	ir.mu.Lock()
	defer ir.mu.Unlock()

	ir.metrics.FeatureFlagDecisions[modelName]++
}

// startBackgroundProcesses starts background processes for the router
func (ir *IntelligentRouter) startBackgroundProcesses() {
	// Start health checking
	go ir.startHealthChecking()

	// Start metrics collection
	if ir.config.MetricsEnabled {
		go ir.startMetricsCollection()
	}

	// Start circuit breaker monitoring
	if ir.config.CircuitBreakerEnabled {
		go ir.startCircuitBreakerMonitoring()
	}
}

// startHealthChecking starts health checking for all endpoints
func (ir *IntelligentRouter) startHealthChecking() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ir.checkEndpointHealth()
	}
}

// startMetricsCollection starts metrics collection
func (ir *IntelligentRouter) startMetricsCollection() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ir.collectMetrics()
	}
}

// startCircuitBreakerMonitoring starts circuit breaker monitoring
func (ir *IntelligentRouter) startCircuitBreakerMonitoring() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for range ticker.C {
		ir.monitorCircuitBreakers()
	}
}

// checkEndpointHealth checks the health of all endpoints
func (ir *IntelligentRouter) checkEndpointHealth() {
	ir.mu.Lock()
	defer ir.mu.Unlock()

	for _, endpoint := range ir.endpoints {
		// Perform health check
		healthy := ir.performHealthCheck(endpoint)
		endpoint.IsHealthy = healthy
		endpoint.LastHealthCheck = time.Now()
	}
}

// performHealthCheck performs a health check for an endpoint
func (ir *IntelligentRouter) performHealthCheck(endpoint *ServiceEndpoint) bool {
	// Simple health check implementation
	// In a real implementation, this would make an HTTP request to the endpoint
	return true
}

// collectMetrics collects metrics from all endpoints
func (ir *IntelligentRouter) collectMetrics() {
	// Implementation for collecting metrics
}

// monitorCircuitBreakers monitors circuit breaker states
func (ir *IntelligentRouter) monitorCircuitBreakers() {
	ir.mu.Lock()
	defer ir.mu.Unlock()

	for _, endpoint := range ir.endpoints {
		if endpoint.CircuitBreaker != nil {
			ir.updateCircuitBreakerState(endpoint.CircuitBreaker)
		}
	}
}

// updateCircuitBreakerState updates the state of a circuit breaker
func (ir *IntelligentRouter) updateCircuitBreakerState(cb *CircuitBreaker) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.State {
	case "closed":
		// Check if failure threshold is reached
		if cb.FailureCount >= cb.FailureThreshold {
			cb.State = "open"
			cb.LastFailureTime = now
		}
	case "open":
		// Check if timeout has passed
		if now.Sub(cb.LastFailureTime) >= cb.Timeout {
			cb.State = "half_open"
			cb.SuccessCount = 0
		}
	case "half_open":
		// Check if success threshold is reached
		if cb.SuccessCount >= cb.SuccessThreshold {
			cb.State = "closed"
			cb.FailureCount = 0
		} else if cb.FailureCount >= cb.FailureThreshold {
			cb.State = "open"
			cb.LastFailureTime = now
		}
	}
}

// GetMetrics returns the current routing metrics
func (ir *IntelligentRouter) GetMetrics() *RoutingMetrics {
	ir.mu.RLock()
	defer ir.mu.RUnlock()
	return ir.metrics
}

// GetEndpoints returns all service endpoints
func (ir *IntelligentRouter) GetEndpoints() map[string]*ServiceEndpoint {
	ir.mu.RLock()
	defer ir.mu.RUnlock()
	return ir.endpoints
}
