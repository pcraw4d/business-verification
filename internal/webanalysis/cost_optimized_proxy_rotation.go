package webanalysis

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

// CostOptimizedProxyRotationSystem provides enterprise-grade proxy rotation with cost optimization
type CostOptimizedProxyRotationSystem struct {
	proxyPools     map[string]*ProxyPool
	geographicMgr  *GeographicManager
	healthMonitor  *HealthMonitor
	loadBalancer   *LoadBalancer
	costOptimizer  *CostOptimizer
	rotationEngine *CostOptimizedRotationEngine
	config         CostOptimizedProxyConfig
	mu             sync.RWMutex
}

// CostOptimizedProxyConfig holds configuration for cost-optimized proxy rotation
type CostOptimizedProxyConfig struct {
	// Geographic distribution across 10+ regions
	GeographicRegions []string `json:"geographic_regions"`
	MaxProxiesPerRegion int    `json:"max_proxies_per_region"`
	
	// Health monitoring and failover
	HealthCheckInterval    time.Duration `json:"health_check_interval"`
	MaxFailures            int           `json:"max_failures"`
	FailoverThreshold      float64       `json:"failover_threshold"`
	AutoFailover           bool          `json:"auto_failover"`
	
	// Performance optimization
	MinLatency             time.Duration `json:"min_latency"`
	MaxLatency             time.Duration `json:"max_latency"`
	LoadBalancingStrategy  string        `json:"load_balancing_strategy"` // round-robin, least-connections, geographic
	
	// Cost optimization
	CostOptimizationEnabled bool    `json:"cost_optimization_enabled"`
	MaxCostPerRequest       float64 `json:"max_cost_per_request"`
	BudgetLimit             float64 `json:"budget_limit"`
	ResidentialProxyRatio   float64 `json:"residential_proxy_ratio"` // 0.0 to 1.0
	
	// Rotation strategies
	RotationStrategies      []string      `json:"rotation_strategies"`
	RotationInterval        time.Duration `json:"rotation_interval"`
	StickySessionDuration   time.Duration `json:"sticky_session_duration"`
	
	// Security and authentication
	ProxyAuthentication     bool          `json:"proxy_authentication"`
	SSLVerification         bool          `json:"ssl_verification"`
	BotDetectionEvasion     bool          `json:"bot_detection_evasion"`
	
	// Monitoring and analytics
	PerformanceTracking     bool          `json:"performance_tracking"`
	CostTracking            bool          `json:"cost_tracking"`
	AnalyticsEnabled        bool          `json:"analytics_enabled"`
}

// ProxyPool represents a pool of proxies for a specific region or type
type ProxyPool struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Region       string            `json:"region"`
	ProxyType    string            `json:"proxy_type"` // residential, datacenter, mobile
	Proxies      []*CostOptimizedProxy  `json:"proxies"`
	Health       PoolHealth        `json:"health"`
	Performance  PoolPerformance   `json:"performance"`
	Cost         PoolCost          `json:"cost"`
	mu           sync.RWMutex
}

// CostOptimizedProxy represents a cost-optimized proxy with comprehensive features
type CostOptimizedProxy struct {
	ID              string            `json:"id"`
	IP              string            `json:"ip"`
	Port            int               `json:"port"`
	Protocol        string            `json:"protocol"` // http, https, socks5
	Type            string            `json:"type"`     // residential, datacenter, mobile
	
	// Geographic information
	Region          string            `json:"region"`
	Country         string            `json:"country"`
	City            string            `json:"city"`
	ISP             string            `json:"isp"`
	
	// Health and performance
	Health          bool              `json:"health"`
	LastUsed        time.Time         `json:"last_used"`
	LastCheck       time.Time         `json:"last_check"`
	FailCount       int               `json:"fail_count"`
	SuccessCount    int               `json:"success_count"`
	Latency         time.Duration     `json:"latency"`
	Uptime          float64           `json:"uptime"`
	
	// Load and capacity
	ConcurrentLimit int               `json:"concurrent_limit"`
	CurrentLoad     int               `json:"current_load"`
	Bandwidth       int64             `json:"bandwidth"` // bytes per second
	
	// Cost information
	CostPerRequest  float64           `json:"cost_per_request"`
	CostPerGB       float64           `json:"cost_per_gb"`
	MonthlyCost     float64           `json:"monthly_cost"`
	
	// Capabilities and features
	Capabilities    map[string]bool   `json:"capabilities"`
	Headers         map[string]string `json:"headers"`
	Cookies         map[string]string `json:"cookies"`
	UserAgents      []string          `json:"user_agents"`
	SSLSupport      bool              `json:"ssl_support"`
	AnonymityLevel  string            `json:"anonymity_level"` // transparent, anonymous, elite
	
	// Rotation and session management
	LastRotation    time.Time         `json:"last_rotation"`
	RotationCount   int               `json:"rotation_count"`
	SessionID       string            `json:"session_id"`
	
	// Authentication
	Username        string            `json:"username,omitempty"`
	Password        string            `json:"password,omitempty"`
	
	// Provider information
	Provider        string            `json:"provider"`
	ProviderID      string            `json:"provider_id"`
}

// PoolHealth represents the health status of a proxy pool
type PoolHealth struct {
	TotalProxies    int     `json:"total_proxies"`
	HealthyProxies  int     `json:"healthy_proxies"`
	UnhealthyProxies int    `json:"unhealthy_proxies"`
	HealthPercentage float64 `json:"health_percentage"`
	LastCheck       time.Time `json:"last_check"`
}

// PoolPerformance represents performance metrics for a proxy pool
type PoolPerformance struct {
	AverageLatency  time.Duration `json:"average_latency"`
	SuccessRate     float64       `json:"success_rate"`
	Throughput      int64         `json:"throughput"` // requests per second
	ErrorRate       float64       `json:"error_rate"`
	LastUpdated     time.Time     `json:"last_updated"`
}

// PoolCost represents cost metrics for a proxy pool
type PoolCost struct {
	TotalCost       float64 `json:"total_cost"`
	CostPerRequest  float64 `json:"cost_per_request"`
	MonthlyBudget   float64 `json:"monthly_budget"`
	BudgetUsed      float64 `json:"budget_used"`
	BudgetRemaining float64 `json:"budget_remaining"`
}

// GeographicManager manages geographic distribution of proxies
type GeographicManager struct {
	regions         map[string]*GeographicRegion
	distributionMap map[string][]string // region -> proxy IDs
	mu              sync.RWMutex
}

// GeographicRegion represents a geographic region with proxy distribution
type GeographicRegion struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Countries       []string `json:"countries"`
	ProxyPools      []string `json:"proxy_pools"`
	LoadBalancing   bool     `json:"load_balancing"`
	FailoverEnabled bool     `json:"failover_enabled"`
}

// HealthMonitor manages proxy health monitoring
type HealthMonitor struct {
	checkURLs       []string
	checkInterval   time.Duration
	timeout         time.Duration
	client          *http.Client
	healthChecks    map[string]*HealthCheck
	mu              sync.RWMutex
}

// HealthCheck represents a health check for a proxy
type HealthCheck struct {
	ProxyID     string    `json:"proxy_id"`
	Status      string    `json:"status"` // healthy, unhealthy, unknown
	LastCheck   time.Time `json:"last_check"`
	Latency     time.Duration `json:"latency"`
	Error       string    `json:"error,omitempty"`
}

// LoadBalancer manages load balancing across proxy pools
type LoadBalancer struct {
	strategy        string // round-robin, least-connections, geographic, cost-based
	weights         map[string]float64
	connections     map[string]int
	mu              sync.RWMutex
}

// CostOptimizer manages cost optimization for proxy usage
type CostOptimizer struct {
	budgetLimit     float64
	costPerRequest  map[string]float64
	usageTracking   map[string]int64
	optimizationRules []OptimizationRule
	mu              sync.RWMutex
}

// OptimizationRule represents a cost optimization rule
type OptimizationRule struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Condition   string  `json:"condition"`
	Action      string  `json:"action"`
	Priority    int     `json:"priority"`
	Enabled     bool    `json:"enabled"`
}

// CostOptimizedRotationEngine manages proxy rotation strategies
type CostOptimizedRotationEngine struct {
	strategies      map[string]CostOptimizedRotationStrategy
	currentStrategy string
	rotationHistory map[string][]RotationEvent
	mu              sync.RWMutex
}

// CostOptimizedRotationStrategy represents a proxy rotation strategy
type CostOptimizedRotationStrategy struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Interval    time.Duration `json:"interval"`
	Enabled     bool          `json:"enabled"`
}

// RotationEvent represents a proxy rotation event
type RotationEvent struct {
	ProxyID     string    `json:"proxy_id"`
	Strategy    string    `json:"strategy"`
	Timestamp   time.Time `json:"timestamp"`
	Reason      string    `json:"reason"`
}

// NewCostOptimizedProxyRotationSystem creates a new cost-optimized proxy rotation system
func NewCostOptimizedProxyRotationSystem(config CostOptimizedProxyConfig) *CostOptimizedProxyRotationSystem {
	if config.GeographicRegions == nil {
		config.GeographicRegions = []string{
			"us-east", "us-west", "us-central",
			"eu-west", "eu-central", "eu-north",
			"asia-east", "asia-southeast", "asia-south",
			"australia", "south-america", "africa",
		}
	}
	
	if config.MaxProxiesPerRegion == 0 {
		config.MaxProxiesPerRegion = 50
	}
	
	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = 5 * time.Minute
	}
	
	if config.MaxFailures == 0 {
		config.MaxFailures = 3
	}
	
	if config.RotationInterval == 0 {
		config.RotationInterval = 30 * time.Second
	}

	return &CostOptimizedProxyRotationSystem{
		proxyPools:     make(map[string]*ProxyPool),
		geographicMgr:  NewGeographicManager(),
		healthMonitor:  NewHealthMonitor(config.HealthCheckInterval),
		loadBalancer:   NewLoadBalancer(config.LoadBalancingStrategy),
		costOptimizer:  NewCostOptimizer(config.BudgetLimit),
		rotationEngine: NewCostOptimizedRotationEngine(),
		config:         config,
	}
}

// AddProxy adds a proxy to the appropriate pool
func (cprs *CostOptimizedProxyRotationSystem) AddProxy(proxy *CostOptimizedProxy) error {
	cprs.mu.Lock()
	defer cprs.mu.Unlock()

	// Determine pool based on region and type
	poolID := fmt.Sprintf("%s-%s", proxy.Region, proxy.Type)
	
	pool, exists := cprs.proxyPools[poolID]
	if !exists {
		pool = &ProxyPool{
			ID:        poolID,
			Name:      fmt.Sprintf("%s %s Proxies", proxy.Region, proxy.Type),
			Region:    proxy.Region,
			ProxyType: proxy.Type,
			Proxies:   make([]*CostOptimizedProxy, 0),
			Health:    PoolHealth{},
			Performance: PoolPerformance{},
			Cost:      PoolCost{},
		}
		cprs.proxyPools[poolID] = pool
	}

	// Add proxy to pool
	pool.mu.Lock()
	pool.Proxies = append(pool.Proxies, proxy)
	pool.mu.Unlock()

	// Update geographic distribution
	cprs.geographicMgr.AddProxyToRegion(proxy.Region, proxy.ID)

	// Start health monitoring
	cprs.healthMonitor.AddProxy(proxy.ID)

	return nil
}

// GetProxy returns the best available proxy based on strategy
func (cprs *CostOptimizedProxyRotationSystem) GetProxy(region string, requirements map[string]interface{}) (*CostOptimizedProxy, error) {
	cprs.mu.RLock()
	defer cprs.mu.RUnlock()

	// Get available pools for the region
	var availablePools []*ProxyPool
	for _, pool := range cprs.proxyPools {
		if pool.Region == region && pool.Health.HealthPercentage > cprs.config.FailoverThreshold {
			availablePools = append(availablePools, pool)
		}
	}

	if len(availablePools) == 0 {
		return nil, fmt.Errorf("no available proxy pools for region %s", region)
	}

	// Apply load balancing strategy
	selectedPool := cprs.loadBalancer.SelectPool(availablePools, requirements)

	// Get best proxy from selected pool
	proxy := cprs.selectBestProxy(selectedPool, requirements)
	if proxy == nil {
		return nil, fmt.Errorf("no available proxies in pool %s", selectedPool.ID)
	}

	// Update proxy usage
	proxy.LastUsed = time.Now()
	proxy.CurrentLoad++

	// Track cost if enabled
	if cprs.config.CostTracking {
		cprs.costOptimizer.TrackUsage(proxy.ID, proxy.CostPerRequest)
	}

	return proxy, nil
}

// selectBestProxy selects the best proxy from a pool based on requirements
func (cprs *CostOptimizedProxyRotationSystem) selectBestProxy(pool *ProxyPool, requirements map[string]interface{}) *CostOptimizedProxy {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	var candidates []*CostOptimizedProxy
	
	// Filter proxies based on requirements
	for _, proxy := range pool.Proxies {
		if !proxy.Health || proxy.CurrentLoad >= proxy.ConcurrentLimit {
			continue
		}

		// Check latency requirements
		if maxLatency, ok := requirements["max_latency"].(time.Duration); ok {
			if proxy.Latency > maxLatency {
				continue
			}
		}

		// Check cost requirements
		if maxCost, ok := requirements["max_cost"].(float64); ok {
			if proxy.CostPerRequest > maxCost {
				continue
			}
		}

		// Check anonymity requirements
		if anonymity, ok := requirements["anonymity"].(string); ok {
			if proxy.AnonymityLevel != anonymity {
				continue
			}
		}

		candidates = append(candidates, proxy)
	}

	if len(candidates) == 0 {
		return nil
	}

	// Sort candidates by performance score
	sort.Slice(candidates, func(i, j int) bool {
		scoreI := cprs.calculateProxyScore(candidates[i])
		scoreJ := cprs.calculateProxyScore(candidates[j])
		return scoreI > scoreJ
	})

	return candidates[0]
}

// calculateProxyScore calculates a performance score for a proxy
func (cprs *CostOptimizedProxyRotationSystem) calculateProxyScore(proxy *CostOptimizedProxy) float64 {
	// Base score from uptime and success rate
	baseScore := proxy.Uptime * (float64(proxy.SuccessCount) / float64(proxy.SuccessCount+proxy.FailCount))
	
	// Latency penalty (lower latency = higher score)
	latencyScore := 1.0 - (float64(proxy.Latency) / float64(cprs.config.MaxLatency))
	
	// Load penalty (lower load = higher score)
	loadScore := 1.0 - (float64(proxy.CurrentLoad) / float64(proxy.ConcurrentLimit))
	
	// Cost penalty (lower cost = higher score)
	costScore := 1.0 - (proxy.CostPerRequest / cprs.config.MaxCostPerRequest)
	
	// Weighted combination
	return baseScore*0.4 + latencyScore*0.3 + loadScore*0.2 + costScore*0.1
}

// RotateProxies performs proxy rotation based on configured strategy
func (cprs *CostOptimizedProxyRotationSystem) RotateProxies() error {
	cprs.mu.Lock()
	defer cprs.mu.Unlock()

	// Get current rotation strategy
	strategy := cprs.rotationEngine.GetCurrentStrategy()
	
	// Apply rotation to all pools
	for _, pool := range cprs.proxyPools {
		cprs.rotatePoolProxies(pool, strategy)
	}

	return nil
}

// rotatePoolProxies rotates proxies within a specific pool
func (cprs *CostOptimizedProxyRotationSystem) rotatePoolProxies(pool *ProxyPool, strategy string) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	switch strategy {
	case "round-robin":
		// Simple round-robin rotation
		for i := range pool.Proxies {
			pool.Proxies[i].LastRotation = time.Now()
			pool.Proxies[i].RotationCount++
		}
		
	case "load-based":
		// Rotate based on load
		for _, proxy := range pool.Proxies {
			if proxy.CurrentLoad > proxy.ConcurrentLimit/2 {
				proxy.LastRotation = time.Now()
				proxy.RotationCount++
			}
		}
		
	case "time-based":
		// Rotate based on time since last rotation
		now := time.Now()
		for _, proxy := range pool.Proxies {
			if now.Sub(proxy.LastRotation) > cprs.config.RotationInterval {
				proxy.LastRotation = now
				proxy.RotationCount++
			}
		}
	}
}

// GetGeographicDistribution returns the geographic distribution of proxies
func (cprs *CostOptimizedProxyRotationSystem) GetGeographicDistribution() map[string]int {
	cprs.mu.RLock()
	defer cprs.mu.RUnlock()

	distribution := make(map[string]int)
	for _, pool := range cprs.proxyPools {
		distribution[pool.Region] += len(pool.Proxies)
	}
	
	return distribution
}

// GetPerformanceMetrics returns performance metrics for all proxy pools
func (cprs *CostOptimizedProxyRotationSystem) GetPerformanceMetrics() map[string]PoolPerformance {
	cprs.mu.RLock()
	defer cprs.mu.RUnlock()

	metrics := make(map[string]PoolPerformance)
	for poolID, pool := range cprs.proxyPools {
		pool.mu.RLock()
		metrics[poolID] = pool.Performance
		pool.mu.RUnlock()
	}
	
	return metrics
}

// GetCostMetrics returns cost metrics for all proxy pools
func (cprs *CostOptimizedProxyRotationSystem) GetCostMetrics() map[string]PoolCost {
	cprs.mu.RLock()
	defer cprs.mu.RUnlock()

	costs := make(map[string]PoolCost)
	for poolID, pool := range cprs.proxyPools {
		pool.mu.RLock()
		costs[poolID] = pool.Cost
		pool.mu.RUnlock()
	}
	
	return costs
}

// NewGeographicManager creates a new geographic manager
func NewGeographicManager() *GeographicManager {
	return &GeographicManager{
		regions:         make(map[string]*GeographicRegion),
		distributionMap: make(map[string][]string),
	}
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(checkInterval time.Duration) *HealthMonitor {
	return &HealthMonitor{
		checkURLs:     []string{"https://httpbin.org/ip", "https://api.ipify.org"},
		checkInterval: checkInterval,
		timeout:       10 * time.Second,
		client:        &http.Client{Timeout: 10 * time.Second},
		healthChecks:  make(map[string]*HealthCheck),
	}
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(strategy string) *LoadBalancer {
	return &LoadBalancer{
		strategy:    strategy,
		weights:     make(map[string]float64),
		connections: make(map[string]int),
	}
}

// NewCostOptimizer creates a new cost optimizer
func NewCostOptimizer(budgetLimit float64) *CostOptimizer {
	return &CostOptimizer{
		budgetLimit:    budgetLimit,
		costPerRequest: make(map[string]float64),
		usageTracking:  make(map[string]int64),
		optimizationRules: []OptimizationRule{
			{
				ID:       "cost_threshold",
				Name:     "Cost Threshold Rule",
				Condition: "cost > threshold",
				Action:   "switch_to_cheaper_proxy",
				Priority: 1,
				Enabled:  true,
			},
		},
	}
}

// NewCostOptimizedRotationEngine creates a new cost-optimized rotation engine
func NewCostOptimizedRotationEngine() *CostOptimizedRotationEngine {
	return &CostOptimizedRotationEngine{
		strategies: map[string]CostOptimizedRotationStrategy{
			"round-robin": {
				ID:          "round-robin",
				Name:        "Round Robin",
				Description: "Rotate proxies in round-robin fashion",
				Interval:    30 * time.Second,
				Enabled:     true,
			},
			"load-based": {
				ID:          "load-based",
				Name:        "Load Based",
				Description: "Rotate based on proxy load",
				Interval:    60 * time.Second,
				Enabled:     true,
			},
			"time-based": {
				ID:          "time-based",
				Name:        "Time Based",
				Description: "Rotate based on time intervals",
				Interval:    5 * time.Minute,
				Enabled:     true,
			},
		},
		currentStrategy: "round-robin",
		rotationHistory: make(map[string][]RotationEvent),
	}
}
