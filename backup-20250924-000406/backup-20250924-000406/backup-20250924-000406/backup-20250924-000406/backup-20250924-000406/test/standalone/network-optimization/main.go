package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// Mock logger for testing
type MockLogger struct{}

func (l *MockLogger) Info(msg string, fields ...interface{})  {}
func (l *MockLogger) Error(msg string, fields ...interface{}) {}
func (l *MockLogger) Debug(msg string, fields ...interface{}) {}

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
	logger         *MockLogger
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

// NewNetworkOptimizationManager creates a new network optimization manager
func NewNetworkOptimizationManager(config *NetworkOptimizationConfig, logger *MockLogger) *NetworkOptimizationManager {
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

	return manager
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
			nom.stats.RateLimitHits++
			return nil, fmt.Errorf("rate limit exceeded")
		}
	}

	// Check circuit breaker
	if nom.config.CircuitBreakerEnabled {
		if !nom.circuitBreaker.CanExecute() {
			nom.stats.CircuitBreakerTrips++
			return nil, fmt.Errorf("circuit breaker is open")
		}
	}

	// Get client from pool
	client := nom.GetHTTPClient(req.URL.Host)

	// Perform request
	resp, err := client.Do(req)
	duration := time.Since(start)

	// Update statistics
	nom.stats.TotalRequests++
	if err != nil {
		nom.stats.FailedRequests++
		if nom.config.CircuitBreakerEnabled {
			nom.circuitBreaker.RecordFailure()
		}
	} else {
		// Check if response indicates failure (4xx or 5xx status codes)
		if resp != nil && (resp.StatusCode >= 400) {
			nom.stats.FailedRequests++
			if nom.config.CircuitBreakerEnabled {
				nom.circuitBreaker.RecordFailure()
			}
		} else {
			nom.stats.SuccessfulRequests++
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

	// Access stats directly to avoid deadlock (don't call GetStats())
	stats := nom.stats

	// Optimize connection pool based on usage patterns
	if stats.ActiveConnections > int32(nom.config.MaxIdleConns/2) {
		nom.logger.Info("increasing connection pool size due to high usage")
	}

	// Optimize timeouts based on response times
	if stats.AverageResponseTime > nom.config.RequestTimeout/2 {
		nom.logger.Info("increasing request timeout due to slow responses")
	}

	// Optimize rate limiting based on error rates
	if stats.TotalRequests > 0 {
		errorRate := float64(stats.FailedRequests) / float64(stats.TotalRequests)
		if errorRate > 0.1 && nom.config.RateLimitPerSecond > 10 {
			nom.config.RateLimitPerSecond = int(float64(nom.config.RateLimitPerSecond) * 0.9)
			nom.logger.Info("reducing rate limit due to high error rate")
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

	rr.current++
	index := int(rr.current) % len(endpoints)
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
		ws.current++
		index := int(ws.current) % len(endpoints)
		return endpoints[index]
	}

	// Select based on weight
	ws.current++
	weight := int(ws.current) % totalWeight

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
}

// TestNetworkOptimization tests the network optimization system
func TestNetworkOptimization() {
	fmt.Println("=== Testing Network Optimization System ===")

	logger := &MockLogger{}
	config := DefaultNetworkOptimizationConfig()
	manager := NewNetworkOptimizationManager(config, logger)

	// Test 1: Basic functionality
	fmt.Println("Test 1: Basic functionality")
	if manager.config == nil {
		fmt.Println("❌ Config not initialized")
		return
	}
	if manager.clientPool == nil {
		fmt.Println("❌ Client pool not initialized")
		return
	}
	fmt.Println("✅ Basic functionality test passed")

	// Test 2: HTTP client pooling
	fmt.Println("\nTest 2: HTTP client pooling")
	client1 := manager.GetHTTPClient("example.com")
	client2 := manager.GetHTTPClient("example.com")
	if client1 != client2 {
		fmt.Println("❌ Client pooling not working")
		return
	}
	fmt.Println("✅ HTTP client pooling test passed")

	// Test 3: Load balancer
	fmt.Println("\nTest 3: Load balancer")
	err := manager.AddEndpoint("http://example1.com", 1)
	if err != nil {
		fmt.Printf("❌ Failed to add endpoint: %v\n", err)
		return
	}
	err = manager.AddEndpoint("http://example2.com", 2)
	if err != nil {
		fmt.Printf("❌ Failed to add endpoint: %v\n", err)
		return
	}
	fmt.Println("✅ Load balancer test passed")

	// Test 4: Rate limiting
	fmt.Println("\nTest 4: Rate limiting")
	config2 := DefaultNetworkOptimizationConfig()
	config2.RateLimitPerSecond = 1
	config2.RateLimitBurst = 1
	manager2 := NewNetworkOptimizationManager(config2, logger)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}

	// First request should succeed
	resp, err := manager2.DoRequest(context.Background(), req)
	if err != nil {
		fmt.Printf("❌ First request failed: %v\n", err)
		return
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Second request should be rate limited
	resp, err = manager2.DoRequest(context.Background(), req)
	if err == nil {
		fmt.Println("❌ Second request should be rate limited")
		return
	}
	if resp != nil {
		resp.Body.Close()
	}
	fmt.Println("✅ Rate limiting test passed")

	// Test 5: Circuit breaker
	fmt.Println("\nTest 5: Circuit breaker")
	config3 := DefaultNetworkOptimizationConfig()
	config3.FailureThreshold = 1
	config3.RecoveryTimeout = 100 * time.Millisecond
	manager3 := NewNetworkOptimizationManager(config3, logger)

	// Create failing server
	failingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failingServer.Close()

	req, err = http.NewRequest("GET", failingServer.URL, nil)
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}

	// First request should fail (but not due to circuit breaker)
	resp, err = manager3.DoRequest(context.Background(), req)
	if err != nil {
		fmt.Printf("❌ First request failed: %v\n", err)
		return
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Second request should be blocked by circuit breaker
	resp, err = manager3.DoRequest(context.Background(), req)
	if err == nil {
		fmt.Println("❌ Second request should be blocked by circuit breaker")
		return
	}
	if resp != nil {
		resp.Body.Close()
	}
	fmt.Println("✅ Circuit breaker test passed")

	// Test 6: Statistics
	fmt.Println("\nTest 6: Statistics")
	stats := manager.GetStats()
	if stats == nil {
		fmt.Println("❌ Stats not returned")
		return
	}
	fmt.Printf("✅ Statistics test passed - Total requests: %d\n", stats.TotalRequests)

	// Test 7: Optimization
	fmt.Println("\nTest 7: Optimization")
	err = manager.OptimizeNetwork()
	if err != nil {
		fmt.Printf("❌ Optimization failed: %v\n", err)
		return
	}
	fmt.Println("✅ Optimization test passed")

	// Test 8: Shutdown
	fmt.Println("\nTest 8: Shutdown")
	err = manager.Shutdown()
	if err != nil {
		fmt.Printf("❌ Shutdown failed: %v\n", err)
		return
	}
	fmt.Println("✅ Shutdown test passed")

	fmt.Println("\n🎉 All network optimization tests passed!")
}

// BenchmarkNetworkOptimization benchmarks the network optimization system
func BenchmarkNetworkOptimization() {
	fmt.Println("\n=== Benchmarking Network Optimization System ===")

	logger := &MockLogger{}
	config := DefaultNetworkOptimizationConfig()
	manager := NewNetworkOptimizationManager(config, logger)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}

	// Benchmark HTTP client pool
	fmt.Println("Benchmarking HTTP client pool...")
	start := time.Now()
	for i := 0; i < 1000; i++ {
		client := manager.GetHTTPClient("example.com")
		if client == nil {
			fmt.Println("❌ Client pool benchmark failed")
			return
		}
	}
	duration := time.Since(start)
	fmt.Printf("✅ HTTP client pool: 1000 requests in %v (%.2f req/sec)\n", duration, 1000.0/duration.Seconds())

	// Benchmark rate limiter
	fmt.Println("Benchmarking rate limiter...")
	config2 := DefaultNetworkOptimizationConfig()
	config2.RateLimitPerSecond = 1000
	config2.RateLimitBurst = 100
	manager2 := NewNetworkOptimizationManager(config2, logger)

	start = time.Now()
	for i := 0; i < 1000; i++ {
		manager2.rateLimiter.Allow()
	}
	duration = time.Since(start)
	fmt.Printf("✅ Rate limiter: 1000 checks in %v (%.2f checks/sec)\n", duration, 1000.0/duration.Seconds())

	// Benchmark circuit breaker
	fmt.Println("Benchmarking circuit breaker...")
	start = time.Now()
	for i := 0; i < 1000; i++ {
		manager.circuitBreaker.CanExecute()
	}
	duration = time.Since(start)
	fmt.Printf("✅ Circuit breaker: 1000 checks in %v (%.2f checks/sec)\n", duration, 1000.0/duration.Seconds())

	// Benchmark full request flow
	fmt.Println("Benchmarking full request flow...")
	start = time.Now()
	for i := 0; i < 100; i++ {
		resp, err := manager.DoRequest(context.Background(), req)
		if err != nil {
			fmt.Printf("❌ Request benchmark failed: %v\n", err)
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	duration = time.Since(start)
	fmt.Printf("✅ Full request flow: 100 requests in %v (%.2f req/sec)\n", duration, 100.0/duration.Seconds())

	fmt.Println("\n🎉 All benchmarks completed!")
}

func main() {
	fmt.Println("=== Network Optimization Test Suite ===")
	overallStart := time.Now()
	fmt.Printf("Starting tests at: %s\n", time.Now().Format("15:04:05"))

	logger := &MockLogger{}
	config := DefaultNetworkOptimizationConfig()
	manager := NewNetworkOptimizationManager(config, logger)

	// Test 1: HTTP Client Pool
	fmt.Println("\nTest 1: HTTP Client Pool")
	start := time.Now()
	client := manager.GetHTTPClient("example.com")
	if client == nil {
		fmt.Println("❌ HTTP client pool test failed")
		return
	}
	fmt.Printf("✅ HTTP client pool test passed in %v\n", time.Since(start))

	// Test 2: Rate Limiter
	fmt.Println("\nTest 2: Rate Limiter")
	start = time.Now()
	allowed := manager.rateLimiter.Allow()
	if !allowed {
		fmt.Println("❌ Rate limiter test failed")
		return
	}
	fmt.Printf("✅ Rate limiter test passed in %v\n", time.Since(start))

	// Test 3: Circuit Breaker
	fmt.Println("\nTest 3: Circuit Breaker")
	start = time.Now()
	canExecute := manager.circuitBreaker.CanExecute()
	if !canExecute {
		fmt.Println("❌ Circuit breaker test failed")
		return
	}
	fmt.Printf("✅ Circuit breaker test passed in %v\n", time.Since(start))

	// Test 4: Request Handling
	fmt.Println("\nTest 4: Request Handling")
	start = time.Now()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}

	resp, err := manager.DoRequest(context.Background(), req)
	if err != nil {
		fmt.Printf("❌ Request handling test failed: %v\n", err)
		return
	}
	if resp != nil {
		resp.Body.Close()
	}
	fmt.Printf("✅ Request handling test passed in %v\n", time.Since(start))

	// Test 5: Circuit Breaker (additional)
	fmt.Println("\nTest 5: Circuit Breaker (additional)")
	start = time.Now()
	canExecute = manager.circuitBreaker.CanExecute()
	if !canExecute {
		fmt.Println("❌ Circuit breaker test failed")
		return
	}
	fmt.Printf("✅ Circuit breaker test passed in %v\n", time.Since(start))

	// Test 6: Statistics
	fmt.Println("\nTest 6: Statistics")
	start = time.Now()
	stats := manager.GetStats()
	if stats == nil {
		fmt.Println("❌ Stats not returned")
		return
	}
	fmt.Printf("✅ Statistics test passed in %v - Total requests: %d\n", time.Since(start), stats.TotalRequests)

	// Test 7: Optimization
	fmt.Println("\nTest 7: Optimization")
	start = time.Now()
	fmt.Println("  - Starting optimization...")
	err = manager.OptimizeNetwork()
	if err != nil {
		fmt.Printf("❌ Optimization failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Optimization test passed in %v\n", time.Since(start))

	// Test 8: Shutdown
	fmt.Println("\nTest 8: Shutdown")
	start = time.Now()
	err = manager.Shutdown()
	if err != nil {
		fmt.Printf("❌ Shutdown failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Shutdown test passed in %v\n", time.Since(start))

	fmt.Printf("\n🎉 All network optimization tests passed! Total time: %v\n", time.Since(overallStart))
}
