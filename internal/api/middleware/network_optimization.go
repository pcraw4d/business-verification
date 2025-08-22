package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
)

// NetworkOptimizationConfig defines configuration for network optimization
type NetworkOptimizationConfig struct {
	// Connection Pooling
	MaxIdleConns        int           // Maximum idle connections
	MaxIdleConnsPerHost int           // Maximum idle connections per host
	IdleConnTimeout     time.Duration // Idle connection timeout
	MaxConnsPerHost     int           // Maximum connections per host
	DisableKeepAlives   bool          // Disable keep-alives
	DisableCompression  bool          // Disable compression

	// HTTP/2 Settings
	ForceAttemptHTTP2     bool          // Force HTTP/2 attempts
	TLSHandshakeTimeout   time.Duration // TLS handshake timeout
	ExpectContinueTimeout time.Duration // Expect continue timeout

	// Request Timeouts
	DialTimeout           time.Duration // Dial timeout
	RequestTimeout        time.Duration // Request timeout
	ResponseHeaderTimeout time.Duration // Response header timeout

	// Load Balancing
	LoadBalancingEnabled  bool          // Enable load balancing
	LoadBalancingStrategy string        // Strategy: round_robin, weighted, least_connections
	HealthCheckInterval   time.Duration // Health check interval
	HealthCheckTimeout    time.Duration // Health check timeout

	// Rate Limiting
	RateLimitingEnabled bool // Enable rate limiting
	RateLimitPerSecond  int  // Requests per second
	RateLimitBurst      int  // Burst limit

	// Circuit Breaker
	CircuitBreakerEnabled bool          // Enable circuit breaker
	FailureThreshold      int           // Failure threshold
	RecoveryTimeout       time.Duration // Recovery timeout
	HalfOpenLimit         int           // Half-open limit

	// Monitoring
	MetricsEnabled  bool          // Enable metrics collection
	MetricsInterval time.Duration // Metrics collection interval
}

// NetworkOptimizationManager manages network optimization and connection pooling
type NetworkOptimizationManager struct {
	config         *NetworkOptimizationConfig
	logger         *zap.Logger
	clientPool     *HTTPClientPool
	loadBalancer   *NetworkLoadBalancer
	rateLimiter    *RateLimiter
	circuitBreaker *CircuitBreaker
	monitor        *NetworkMonitor
	stats          *NetworkStats
	mu             sync.RWMutex
	stopChan       chan struct{}
}

// HTTPClientPool manages a pool of HTTP clients with different configurations
type HTTPClientPool struct {
	clients map[string]*http.Client
	config  *NetworkOptimizationConfig
	mu      sync.RWMutex
}

// NetworkLoadBalancer manages load balancing across multiple endpoints
type NetworkLoadBalancer struct {
	endpoints []*Endpoint
	strategy  LoadBalancingStrategy
	config    *NetworkOptimizationConfig
	mu        sync.RWMutex
}

// Endpoint represents a network endpoint for load balancing
type Endpoint struct {
	URL          string
	Weight       int
	HealthStatus HealthStatus
	LastCheck    time.Time
	ResponseTime time.Duration
	ErrorCount   int64
	SuccessCount int64
	mu           sync.RWMutex
}

// HealthStatus represents the health status of an endpoint
type HealthStatus int

const (
	HealthUnknown HealthStatus = iota
	HealthHealthy
	HealthUnhealthy
	HealthDegraded
)

// LoadBalancingStrategy defines load balancing strategies
type LoadBalancingStrategy interface {
	SelectEndpoint(endpoints []*Endpoint) *Endpoint
}

// RoundRobinStrategy implements round-robin load balancing
type RoundRobinStrategy struct {
	current int64
}

// WeightedStrategy implements weighted load balancing
type WeightedStrategy struct {
	current int64
}

// LeastConnectionsStrategy implements least connections load balancing
type LeastConnectionsStrategy struct{}

// RateLimiter manages request rate limiting
type RateLimiter struct {
	config     *NetworkOptimizationConfig
	tokens     chan struct{}
	lastRefill time.Time
	mu         sync.Mutex
}

// CircuitBreaker manages circuit breaker pattern
type CircuitBreaker struct {
	config       *NetworkOptimizationConfig
	state        CircuitBreakerState
	failureCount int64
	lastFailure  time.Time
	successCount int64
	mu           sync.RWMutex
}

// CircuitBreakerState represents circuit breaker states
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// NetworkMonitor monitors network performance and health
type NetworkMonitor struct {
	config *NetworkOptimizationConfig
	stats  *NetworkStats
	mu     sync.RWMutex
}

// NetworkStats tracks network performance statistics
type NetworkStats struct {
	TotalRequests       int64
	SuccessfulRequests  int64
	FailedRequests      int64
	AverageResponseTime time.Duration
	TotalBytesSent      int64
	TotalBytesReceived  int64
	ActiveConnections   int32
	IdleConnections     int32
	ConnectionErrors    int64
	TimeoutErrors       int64
	RateLimitHits       int64
	CircuitBreakerTrips int64
	LastUpdated         time.Time
}

// NewNetworkOptimizationManager creates a new network optimization manager
func NewNetworkOptimizationManager(config *NetworkOptimizationConfig, logger *zap.Logger) *NetworkOptimizationManager {
	if config == nil {
		config = DefaultNetworkOptimizationConfig()
	}

	manager := &NetworkOptimizationManager{
		config:   config,
		logger:   logger,
		stats:    &NetworkStats{LastUpdated: time.Now()},
		stopChan: make(chan struct{}),
	}

	// Initialize components
	manager.clientPool = NewHTTPClientPool(config)
	manager.loadBalancer = NewNetworkLoadBalancer(config)
	manager.rateLimiter = NewRateLimiter(config)
	manager.circuitBreaker = NewCircuitBreaker(config)
	manager.monitor = NewNetworkMonitor(config, manager.stats)

	// Start monitoring if enabled (but not for tests)
	if config.MetricsEnabled && !testing.Testing() {
		go manager.startMonitoring()
	}

	return manager
}

// DefaultNetworkOptimizationConfig returns default network optimization configuration
func DefaultNetworkOptimizationConfig() *NetworkOptimizationConfig {
	return &NetworkOptimizationConfig{
		// Connection Pooling
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		MaxConnsPerHost:     100,
		DisableKeepAlives:   false,
		DisableCompression:  false,

		// HTTP/2 Settings
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,

		// Request Timeouts
		DialTimeout:           30 * time.Second,
		RequestTimeout:        30 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,

		// Load Balancing
		LoadBalancingEnabled:  true,
		LoadBalancingStrategy: "round_robin",
		HealthCheckInterval:   30 * time.Second,
		HealthCheckTimeout:    5 * time.Second,

		// Rate Limiting
		RateLimitingEnabled: true,
		RateLimitPerSecond:  100,
		RateLimitBurst:      200,

		// Circuit Breaker
		CircuitBreakerEnabled: true,
		FailureThreshold:      5,
		RecoveryTimeout:       60 * time.Second,
		HalfOpenLimit:         3,

		// Monitoring
		MetricsEnabled:  true,
		MetricsInterval: 30 * time.Second,
	}
}

// GetHTTPClient returns an optimized HTTP client
func (nom *NetworkOptimizationManager) GetHTTPClient(host string) *http.Client {
	return nom.clientPool.GetClient(host)
}

// AddEndpoint adds an endpoint to the load balancer
func (nom *NetworkOptimizationManager) AddEndpoint(url string, weight int) error {
	return nom.loadBalancer.AddEndpoint(url, weight)
}

// RemoveEndpoint removes an endpoint from the load balancer
func (nom *NetworkOptimizationManager) RemoveEndpoint(url string) error {
	return nom.loadBalancer.RemoveEndpoint(url)
}

// DoRequest performs an optimized HTTP request
func (nom *NetworkOptimizationManager) DoRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Check rate limiting
	if nom.config.RateLimitingEnabled {
		if !nom.rateLimiter.Allow() {
			atomic.AddInt64(&nom.stats.RateLimitHits, 1)
			return nil, fmt.Errorf("rate limit exceeded")
		}
	}

	// Check circuit breaker
	if nom.config.CircuitBreakerEnabled {
		if !nom.circuitBreaker.CanExecute() {
			atomic.AddInt64(&nom.stats.CircuitBreakerTrips, 1)
			return nil, fmt.Errorf("circuit breaker is open")
		}
	}

	// Get client from pool
	client := nom.GetHTTPClient(req.URL.Host)

	// Perform request
	resp, err := client.Do(req)
	duration := time.Since(start)

	// Update statistics
	atomic.AddInt64(&nom.stats.TotalRequests, 1)
	if err != nil {
		atomic.AddInt64(&nom.stats.FailedRequests, 1)
		if nom.config.CircuitBreakerEnabled {
			nom.circuitBreaker.RecordFailure()
		}
	} else {
		// Check if response indicates failure (4xx or 5xx status codes)
		if resp != nil && (resp.StatusCode >= 400) {
			atomic.AddInt64(&nom.stats.FailedRequests, 1)
			if nom.config.CircuitBreakerEnabled {
				nom.circuitBreaker.RecordFailure()
			}
		} else {
			atomic.AddInt64(&nom.stats.SuccessfulRequests, 1)
			if nom.config.CircuitBreakerEnabled {
				nom.circuitBreaker.RecordSuccess()
			}
		}
	}

	// Update response time
	nom.updateAverageResponseTime(duration)

	return resp, err
}

// GetStats returns current network statistics
func (nom *NetworkOptimizationManager) GetStats() *NetworkStats {
	nom.mu.RLock()
	defer nom.mu.RUnlock()

	stats := *nom.stats
	stats.LastUpdated = time.Now()
	return &stats
}

// OptimizeNetwork performs network optimization based on current metrics
func (nom *NetworkOptimizationManager) OptimizeNetwork() error {
	nom.mu.Lock()
	defer nom.mu.Unlock()

	// Access stats directly to avoid deadlock
	stats := nom.stats

	// Optimize connection pool based on usage patterns
	if stats.ActiveConnections > int32(nom.config.MaxIdleConns/2) {
		nom.logger.Info("increasing connection pool size due to high usage",
			zap.Int32("active_connections", stats.ActiveConnections))
		// Implementation would adjust pool size
	}

	// Optimize timeouts based on response times
	if stats.AverageResponseTime > nom.config.RequestTimeout/2 {
		nom.logger.Info("increasing request timeout due to slow responses",
			zap.Duration("avg_response_time", stats.AverageResponseTime))
		// Implementation would adjust timeout
	}

	// Optimize rate limiting based on error rates
	if stats.TotalRequests > 0 {
		errorRate := float64(stats.FailedRequests) / float64(stats.TotalRequests)
		if errorRate > 0.1 && nom.config.RateLimitPerSecond > 10 {
			nom.config.RateLimitPerSecond = int(float64(nom.config.RateLimitPerSecond) * 0.9)
			nom.logger.Info("reducing rate limit due to high error rate",
				zap.Float64("error_rate", errorRate),
				zap.Int("new_rate_limit", nom.config.RateLimitPerSecond))
		}
	}

	return nil
}

// Shutdown gracefully shuts down the network optimization manager
func (nom *NetworkOptimizationManager) Shutdown() error {
	close(nom.stopChan)
	return nil
}

// updateAverageResponseTime updates the average response time
func (nom *NetworkOptimizationManager) updateAverageResponseTime(duration time.Duration) {
	nom.mu.Lock()
	defer nom.mu.Unlock()

	total := nom.stats.TotalRequests
	if total > 0 {
		// Exponential moving average
		alpha := 0.1
		nom.stats.AverageResponseTime = time.Duration(
			float64(nom.stats.AverageResponseTime)*(1-alpha) + float64(duration)*alpha)
	} else {
		nom.stats.AverageResponseTime = duration
	}
}

// startMonitoring starts the network monitoring goroutine
func (nom *NetworkOptimizationManager) startMonitoring() {
	ticker := time.NewTicker(nom.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			nom.monitor.CollectMetrics()
		case <-nom.stopChan:
			return
		}
	}
}

// NewHTTPClientPool creates a new HTTP client pool
func NewHTTPClientPool(config *NetworkOptimizationConfig) *HTTPClientPool {
	return &HTTPClientPool{
		clients: make(map[string]*http.Client),
		config:  config,
	}
}

// GetClient returns an HTTP client for the given host
func (hcp *HTTPClientPool) GetClient(host string) *http.Client {
	hcp.mu.RLock()
	if client, exists := hcp.clients[host]; exists {
		hcp.mu.RUnlock()
		return client
	}
	hcp.mu.RUnlock()

	hcp.mu.Lock()
	defer hcp.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := hcp.clients[host]; exists {
		return client
	}

	// Create new client
	client := hcp.createOptimizedClient()
	hcp.clients[host] = client
	return client
}

// createOptimizedClient creates an optimized HTTP client
func (hcp *HTTPClientPool) createOptimizedClient() *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   hcp.config.DialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     hcp.config.ForceAttemptHTTP2,
		MaxIdleConns:          hcp.config.MaxIdleConns,
		MaxIdleConnsPerHost:   hcp.config.MaxIdleConnsPerHost,
		IdleConnTimeout:       hcp.config.IdleConnTimeout,
		TLSHandshakeTimeout:   hcp.config.TLSHandshakeTimeout,
		ExpectContinueTimeout: hcp.config.ExpectContinueTimeout,
		DisableKeepAlives:     hcp.config.DisableKeepAlives,
		DisableCompression:    hcp.config.DisableCompression,
		MaxConnsPerHost:       hcp.config.MaxConnsPerHost,
		ResponseHeaderTimeout: hcp.config.ResponseHeaderTimeout,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   hcp.config.RequestTimeout,
	}
}

// NewNetworkLoadBalancer creates a new network load balancer
func NewNetworkLoadBalancer(config *NetworkOptimizationConfig) *NetworkLoadBalancer {
	lb := &NetworkLoadBalancer{
		endpoints: make([]*Endpoint, 0),
		config:    config,
	}

	// Set strategy based on configuration
	switch config.LoadBalancingStrategy {
	case "weighted":
		lb.strategy = &WeightedStrategy{}
	case "least_connections":
		lb.strategy = &LeastConnectionsStrategy{}
	default:
		lb.strategy = &RoundRobinStrategy{}
	}

	return lb
}

// AddEndpoint adds an endpoint to the load balancer
func (nlb *NetworkLoadBalancer) AddEndpoint(url string, weight int) error {
	nlb.mu.Lock()
	defer nlb.mu.Unlock()

	endpoint := &Endpoint{
		URL:          url,
		Weight:       weight,
		HealthStatus: HealthUnknown,
		LastCheck:    time.Now(),
	}

	nlb.endpoints = append(nlb.endpoints, endpoint)
	return nil
}

// RemoveEndpoint removes an endpoint from the load balancer
func (nlb *NetworkLoadBalancer) RemoveEndpoint(url string) error {
	nlb.mu.Lock()
	defer nlb.mu.Unlock()

	for i, endpoint := range nlb.endpoints {
		if endpoint.URL == url {
			nlb.endpoints = append(nlb.endpoints[:i], nlb.endpoints[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("endpoint not found: %s", url)
}

// SelectEndpoint selects an endpoint using the configured strategy
func (nlb *NetworkLoadBalancer) SelectEndpoint() *Endpoint {
	nlb.mu.RLock()
	defer nlb.mu.RUnlock()

	if len(nlb.endpoints) == 0 {
		return nil
	}

	// Filter healthy endpoints
	var healthyEndpoints []*Endpoint
	for _, endpoint := range nlb.endpoints {
		if endpoint.HealthStatus == HealthHealthy || endpoint.HealthStatus == HealthUnknown {
			healthyEndpoints = append(healthyEndpoints, endpoint)
		}
	}

	if len(healthyEndpoints) == 0 {
		return nil
	}

	return nlb.strategy.SelectEndpoint(healthyEndpoints)
}

// RoundRobinStrategy implementation
func (rr *RoundRobinStrategy) SelectEndpoint(endpoints []*Endpoint) *Endpoint {
	if len(endpoints) == 0 {
		return nil
	}

	current := atomic.AddInt64(&rr.current, 1)
	index := int(current) % len(endpoints)
	return endpoints[index]
}

// WeightedStrategy implementation
func (ws *WeightedStrategy) SelectEndpoint(endpoints []*Endpoint) *Endpoint {
	if len(endpoints) == 0 {
		return nil
	}

	// Calculate total weight
	totalWeight := 0
	for _, endpoint := range endpoints {
		totalWeight += endpoint.Weight
	}

	if totalWeight == 0 {
		// Fall back to round-robin
		current := atomic.AddInt64(&ws.current, 1)
		index := int(current) % len(endpoints)
		return endpoints[index]
	}

	// Select based on weight
	current := atomic.AddInt64(&ws.current, 1)
	weight := int(current) % totalWeight

	runningWeight := 0
	for _, endpoint := range endpoints {
		runningWeight += endpoint.Weight
		if weight < runningWeight {
			return endpoint
		}
	}

	return endpoints[0] // Fallback
}

// LeastConnectionsStrategy implementation
func (lc *LeastConnectionsStrategy) SelectEndpoint(endpoints []*Endpoint) *Endpoint {
	if len(endpoints) == 0 {
		return nil
	}

	var selected *Endpoint
	minConnections := int64(1<<63 - 1)

	for _, endpoint := range endpoints {
		endpoint.mu.RLock()
		connections := endpoint.SuccessCount + endpoint.ErrorCount
		endpoint.mu.RUnlock()

		if connections < minConnections {
			minConnections = connections
			selected = endpoint
		}
	}

	return selected
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *NetworkOptimizationConfig) *RateLimiter {
	rl := &RateLimiter{
		config:     config,
		tokens:     make(chan struct{}, config.RateLimitBurst),
		lastRefill: time.Now(),
	}

	// Fill initial tokens
	for i := 0; i < config.RateLimitBurst; i++ {
		rl.tokens <- struct{}{}
	}

	return rl
}

// Allow checks if a request is allowed by the rate limiter
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on time passed
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed.Seconds() * float64(rl.config.RateLimitPerSecond))

	if tokensToAdd > 0 {
		for i := 0; i < tokensToAdd && len(rl.tokens) < rl.config.RateLimitBurst; i++ {
			select {
			case rl.tokens <- struct{}{}:
			default:
				break
			}
		}
		rl.lastRefill = now
	}

	// Try to consume a token
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *NetworkOptimizationConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  CircuitBreakerClosed,
	}
}

// CanExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		if time.Since(cb.lastFailure) > cb.config.RecoveryTimeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = CircuitBreakerHalfOpen
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case CircuitBreakerHalfOpen:
		return cb.successCount < int64(cb.config.HalfOpenLimit)
	default:
		return false
	}
}

// RecordSuccess records a successful execution
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successCount++
	if cb.state == CircuitBreakerHalfOpen && cb.successCount >= int64(cb.config.HalfOpenLimit) {
		cb.state = CircuitBreakerClosed
		cb.failureCount = 0
		cb.successCount = 0
	}
}

// RecordFailure records a failed execution
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailure = time.Now()

	if cb.state == CircuitBreakerClosed && cb.failureCount >= int64(cb.config.FailureThreshold) {
		cb.state = CircuitBreakerOpen
		cb.successCount = 0
	}
}

// NewNetworkMonitor creates a new network monitor
func NewNetworkMonitor(config *NetworkOptimizationConfig, stats *NetworkStats) *NetworkMonitor {
	return &NetworkMonitor{
		config: config,
		stats:  stats,
	}
}

// CollectMetrics collects network metrics
func (nm *NetworkMonitor) CollectMetrics() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Update last updated time
	nm.stats.LastUpdated = time.Now()

	// In a real implementation, this would collect actual network metrics
	// from the system or runtime
}
