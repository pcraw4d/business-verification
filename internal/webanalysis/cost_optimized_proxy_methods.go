package webanalysis

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"
)

// GeographicManager methods

// AddProxyToRegion adds a proxy to a geographic region
func (gm *GeographicManager) AddProxyToRegion(region, proxyID string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if gm.distributionMap[region] == nil {
		gm.distributionMap[region] = make([]string, 0)
	}
	gm.distributionMap[region] = append(gm.distributionMap[region], proxyID)
}

// GetProxiesForRegion returns all proxy IDs for a specific region
func (gm *GeographicManager) GetProxiesForRegion(region string) []string {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	return gm.distributionMap[region]
}

// GetRegionDistribution returns the distribution of proxies across regions
func (gm *GeographicManager) GetRegionDistribution() map[string]int {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	distribution := make(map[string]int)
	for region, proxies := range gm.distributionMap {
		distribution[region] = len(proxies)
	}
	return distribution
}

// HealthMonitor methods

// AddProxy adds a proxy to health monitoring
func (hm *HealthMonitor) AddProxy(proxyID string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.healthChecks[proxyID] = &HealthCheck{
		ProxyID:   proxyID,
		Status:    "unknown",
		LastCheck: time.Now(),
	}
}

// CheckProxyHealth performs a health check on a specific proxy
func (hm *HealthMonitor) CheckProxyHealth(proxy *CostOptimizedProxy) error {
	start := time.Now()
	
	// Create proxy URL
	proxyURL := fmt.Sprintf("%s://%s:%d", proxy.Protocol, proxy.IP, proxy.Port)
	
	// Create HTTP client with proxy
	client := &http.Client{
		Timeout: hm.timeout,
		Transport: &http.Transport{
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return url.Parse(proxyURL)
			},
		},
	}

	// Test with a simple request
	resp, err := client.Get(hm.checkURLs[0])
	if err != nil {
		hm.updateHealthCheck(proxy.ID, "unhealthy", 0, err.Error())
		return err
	}
	defer resp.Body.Close()

	latency := time.Since(start)
	
	if resp.StatusCode == 200 {
		hm.updateHealthCheck(proxy.ID, "healthy", latency, "")
		proxy.Health = true
		proxy.SuccessCount++
		proxy.Latency = latency
	} else {
		hm.updateHealthCheck(proxy.ID, "unhealthy", latency, fmt.Sprintf("status code: %d", resp.StatusCode))
		proxy.Health = false
		proxy.FailCount++
	}

	proxy.LastCheck = time.Now()
	return nil
}

// updateHealthCheck updates the health check status for a proxy
func (hm *HealthMonitor) updateHealthCheck(proxyID, status string, latency time.Duration, error string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if check, exists := hm.healthChecks[proxyID]; exists {
		check.Status = status
		check.LastCheck = time.Now()
		check.Latency = latency
		check.Error = error
	}
}

// GetHealthStatus returns the health status for a proxy
func (hm *HealthMonitor) GetHealthStatus(proxyID string) (*HealthCheck, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	check, exists := hm.healthChecks[proxyID]
	return check, exists
}

// LoadBalancer methods

// SelectPool selects the best pool based on load balancing strategy
func (lb *LoadBalancer) SelectPool(pools []*ProxyPool, requirements map[string]interface{}) *ProxyPool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	switch lb.strategy {
	case "round-robin":
		return lb.roundRobinSelect(pools)
	case "least-connections":
		return lb.leastConnectionsSelect(pools)
	case "geographic":
		return lb.geographicSelect(pools, requirements)
	case "cost-based":
		return lb.costBasedSelect(pools, requirements)
	default:
		return lb.roundRobinSelect(pools)
	}
}

// roundRobinSelect implements round-robin selection
func (lb *LoadBalancer) roundRobinSelect(pools []*ProxyPool) *ProxyPool {
	if len(pools) == 0 {
		return nil
	}
	
	// Simple round-robin selection
	selected := pools[lb.connections["round-robin"]%len(pools)]
	lb.connections["round-robin"]++
	return selected
}

// leastConnectionsSelect selects the pool with the least connections
func (lb *LoadBalancer) leastConnectionsSelect(pools []*ProxyPool) *ProxyPool {
	if len(pools) == 0 {
		return nil
	}

	var selected *ProxyPool
	minConnections := int(^uint(0) >> 1) // Max int

	for _, pool := range pools {
		totalConnections := 0
		for _, proxy := range pool.Proxies {
			totalConnections += proxy.CurrentLoad
		}
		
		if totalConnections < minConnections {
			minConnections = totalConnections
			selected = pool
		}
	}

	return selected
}

// geographicSelect selects pool based on geographic requirements
func (lb *LoadBalancer) geographicSelect(pools []*ProxyPool, requirements map[string]interface{}) *ProxyPool {
	if len(pools) == 0 {
		return nil
	}

	// If specific region is required, find matching pool
	if region, ok := requirements["region"].(string); ok {
		for _, pool := range pools {
			if pool.Region == region {
				return pool
			}
		}
	}

	// Default to first available pool
	return pools[0]
}

// costBasedSelect selects pool based on cost optimization
func (lb *LoadBalancer) costBasedSelect(pools []*ProxyPool, requirements map[string]interface{}) *ProxyPool {
	if len(pools) == 0 {
		return nil
	}

	var selected *ProxyPool
	minCost := float64(^uint(0) >> 1) // Max float64

	for _, pool := range pools {
		if pool.Cost.CostPerRequest < minCost {
			minCost = pool.Cost.CostPerRequest
			selected = pool
		}
	}

	return selected
}

// CostOptimizer methods

// TrackUsage tracks proxy usage for cost optimization
func (co *CostOptimizer) TrackUsage(proxyID string, costPerRequest float64) {
	co.mu.Lock()
	defer co.mu.Unlock()

	co.costPerRequest[proxyID] = costPerRequest
	co.usageTracking[proxyID]++
}

// GetTotalCost returns the total cost for all proxy usage
func (co *CostOptimizer) GetTotalCost() float64 {
	co.mu.RLock()
	defer co.mu.RUnlock()

	totalCost := 0.0
	for proxyID, requests := range co.usageTracking {
		if costPerRequest, exists := co.costPerRequest[proxyID]; exists {
			totalCost += costPerRequest * float64(requests)
		}
	}
	return totalCost
}

// GetBudgetRemaining returns the remaining budget
func (co *CostOptimizer) GetBudgetRemaining() float64 {
	return co.budgetLimit - co.GetTotalCost()
}

// IsBudgetExceeded checks if the budget has been exceeded
func (co *CostOptimizer) IsBudgetExceeded() bool {
	return co.GetTotalCost() > co.budgetLimit
}

// GetCostOptimizationRecommendations returns cost optimization recommendations
func (co *CostOptimizer) GetCostOptimizationRecommendations() []string {
	co.mu.RLock()
	defer co.mu.RUnlock()

	var recommendations []string

	// Find most expensive proxies
	var expensiveProxies []struct {
		ID   string
		Cost float64
	}

	for proxyID, cost := range co.costPerRequest {
		expensiveProxies = append(expensiveProxies, struct {
			ID   string
			Cost float64
		}{proxyID, cost})
	}

	// Sort by cost (descending)
	sort.Slice(expensiveProxies, func(i, j int) bool {
		return expensiveProxies[i].Cost > expensiveProxies[j].Cost
	})

	// Generate recommendations
	if len(expensiveProxies) > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Consider replacing expensive proxy %s (cost: $%.4f per request)", 
				expensiveProxies[0].ID, expensiveProxies[0].Cost))
	}

	if co.IsBudgetExceeded() {
		recommendations = append(recommendations, "Budget exceeded - consider reducing proxy usage or switching to cheaper proxies")
	}

	return recommendations
}

// CostOptimizedRotationEngine methods

// GetCurrentStrategy returns the current rotation strategy
func (cre *CostOptimizedRotationEngine) GetCurrentStrategy() string {
	cre.mu.RLock()
	defer cre.mu.RUnlock()

	return cre.currentStrategy
}

// SetStrategy sets the current rotation strategy
func (cre *CostOptimizedRotationEngine) SetStrategy(strategyID string) error {
	cre.mu.Lock()
	defer cre.mu.Unlock()

	if _, exists := cre.strategies[strategyID]; !exists {
		return fmt.Errorf("strategy %s not found", strategyID)
	}

	cre.currentStrategy = strategyID
	return nil
}

// GetAvailableStrategies returns all available rotation strategies
func (cre *CostOptimizedRotationEngine) GetAvailableStrategies() map[string]CostOptimizedRotationStrategy {
	cre.mu.RLock()
	defer cre.mu.RUnlock()

	strategies := make(map[string]CostOptimizedRotationStrategy)
	for id, strategy := range cre.strategies {
		strategies[id] = strategy
	}
	return strategies
}

// RecordRotationEvent records a proxy rotation event
func (cre *CostOptimizedRotationEngine) RecordRotationEvent(proxyID, strategy, reason string) {
	cre.mu.Lock()
	defer cre.mu.Unlock()

	event := RotationEvent{
		ProxyID:   proxyID,
		Strategy:  strategy,
		Timestamp: time.Now(),
		Reason:    reason,
	}

	cre.rotationHistory[proxyID] = append(cre.rotationHistory[proxyID], event)
}

// GetRotationHistory returns rotation history for a proxy
func (cre *CostOptimizedRotationEngine) GetRotationHistory(proxyID string) []RotationEvent {
	cre.mu.RLock()
	defer cre.mu.RUnlock()

	return cre.rotationHistory[proxyID]
}

// Self-hosted proxy discovery and management

// SelfHostedProxyDiscovery discovers and manages self-hosted proxies
type SelfHostedProxyDiscovery struct {
	discoveryMethods []DiscoveryMethod
	proxySources     map[string]ProxySource
	mu               sync.RWMutex
}

// DiscoveryMethod represents a method for discovering proxies
type DiscoveryMethod struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// ProxySource represents a source of proxy information
type ProxySource struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Type     string `json:"type"` // free, paid, self-hosted
	Reliability float64 `json:"reliability"`
	Enabled  bool   `json:"enabled"`
}

// NewSelfHostedProxyDiscovery creates a new self-hosted proxy discovery system
func NewSelfHostedProxyDiscovery() *SelfHostedProxyDiscovery {
	return &SelfHostedProxyDiscovery{
		discoveryMethods: []DiscoveryMethod{
			{
				ID:          "public_lists",
				Name:        "Public Proxy Lists",
				Description: "Discover proxies from public proxy lists",
				Enabled:     true,
			},
			{
				ID:          "web_scraping",
				Name:        "Web Scraping",
				Description: "Scrape proxy lists from websites",
				Enabled:     true,
			},
			{
				ID:          "p2p_networks",
				Name:        "P2P Networks",
				Description: "Discover proxies from P2P networks",
				Enabled:     false, // Disabled for security
			},
		},
		proxySources: map[string]ProxySource{
			"free-proxy-list": {
				ID:          "free-proxy-list",
				Name:        "Free Proxy List",
				URL:         "https://free-proxy-list.net/",
				Type:        "free",
				Reliability: 0.3,
				Enabled:     true,
			},
			"proxy-list": {
				ID:          "proxy-list",
				Name:        "Proxy List",
				URL:         "https://www.proxy-list.download/",
				Type:        "free",
				Reliability: 0.4,
				Enabled:     true,
			},
		},
	}
}

// DiscoverProxies discovers proxies using enabled methods
func (spd *SelfHostedProxyDiscovery) DiscoverProxies() ([]*CostOptimizedProxy, error) {
	spd.mu.RLock()
	defer spd.mu.RUnlock()

	var discoveredProxies []*CostOptimizedProxy

	for _, method := range spd.discoveryMethods {
		if !method.Enabled {
			continue
		}

		proxies, err := spd.discoverWithMethod(method)
		if err != nil {
			continue // Continue with other methods even if one fails
		}

		discoveredProxies = append(discoveredProxies, proxies...)
	}

	return discoveredProxies, nil
}

// discoverWithMethod discovers proxies using a specific method
func (spd *SelfHostedProxyDiscovery) discoverWithMethod(method DiscoveryMethod) ([]*CostOptimizedProxy, error) {
	switch method.ID {
	case "public_lists":
		return spd.discoverFromPublicLists()
	case "web_scraping":
		return spd.discoverFromWebScraping()
	default:
		return nil, fmt.Errorf("unknown discovery method: %s", method.ID)
	}
}

// discoverFromPublicLists discovers proxies from public proxy lists
func (spd *SelfHostedProxyDiscovery) discoverFromPublicLists() ([]*CostOptimizedProxy, error) {
	var proxies []*CostOptimizedProxy

	// This would implement actual proxy list parsing
	// For now, return a sample proxy
	sampleProxy := &CostOptimizedProxy{
		ID:              fmt.Sprintf("discovered-%d", time.Now().UnixNano()),
		IP:              "192.168.1.100",
		Port:            8080,
		Protocol:        "http",
		Type:            "datacenter",
		Region:          "us-east",
		Country:         "US",
		City:            "New York",
		ISP:             "Sample ISP",
		Health:          true,
		ConcurrentLimit: 10,
		CostPerRequest:  0.001, // Very low cost for self-hosted
		CostPerGB:       0.01,
		SSLSupport:      true,
		AnonymityLevel:  "anonymous",
		Capabilities:    make(map[string]bool),
		Headers:         make(map[string]string),
		Cookies:         make(map[string]string),
		UserAgents:      []string{},
		Provider:        "self-hosted",
	}

	proxies = append(proxies, sampleProxy)
	return proxies, nil
}

// discoverFromWebScraping discovers proxies by scraping websites
func (spd *SelfHostedProxyDiscovery) discoverFromWebScraping() ([]*CostOptimizedProxy, error) {
	// This would implement web scraping to find proxy lists
	// For now, return empty list
	return []*CostOptimizedProxy{}, nil
}

// ProxyHealthChecker provides comprehensive health checking for proxies
type ProxyHealthChecker struct {
	checkURLs       []string
	timeout         time.Duration
	concurrentLimit int
	results         map[string]*HealthCheckResult
	mu              sync.RWMutex
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	ProxyID     string        `json:"proxy_id"`
	Status      string        `json:"status"`
	Latency     time.Duration `json:"latency"`
	Error       string        `json:"error,omitempty"`
	LastCheck   time.Time     `json:"last_check"`
	SuccessRate float64       `json:"success_rate"`
}

// NewProxyHealthChecker creates a new proxy health checker
func NewProxyHealthChecker() *ProxyHealthChecker {
	return &ProxyHealthChecker{
		checkURLs: []string{
			"https://httpbin.org/ip",
			"https://api.ipify.org",
			"https://ipinfo.io/json",
		},
		timeout:         10 * time.Second,
		concurrentLimit: 10,
		results:         make(map[string]*HealthCheckResult),
	}
}

// CheckProxyHealth performs a comprehensive health check on a proxy
func (phc *ProxyHealthChecker) CheckProxyHealth(proxy *CostOptimizedProxy) (*HealthCheckResult, error) {
	start := time.Now()
	
	// Create proxy URL
	proxyURL := fmt.Sprintf("%s://%s:%d", proxy.Protocol, proxy.IP, proxy.Port)
	
	// Create HTTP client with proxy
	client := &http.Client{
		Timeout: phc.timeout,
		Transport: &http.Transport{
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return url.Parse(proxyURL)
			},
		},
	}

	// Test with multiple URLs
	successCount := 0

	for _, checkURL := range phc.checkURLs {
		resp, err := client.Get(checkURL)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			successCount++
		}
	}

	latency := time.Since(start)
	successRate := float64(successCount) / float64(len(phc.checkURLs))

	result := &HealthCheckResult{
		ProxyID:     proxy.ID,
		Status:      phc.determineStatus(successRate),
		Latency:     latency,
		LastCheck:   time.Now(),
		SuccessRate: successRate,
	}

	// Update proxy with results
	proxy.Health = result.Status == "healthy"
	proxy.Latency = latency
	if proxy.Health {
		proxy.SuccessCount++
	} else {
		proxy.FailCount++
	}
	proxy.LastCheck = time.Now()

	// Store result
	phc.mu.Lock()
	phc.results[proxy.ID] = result
	phc.mu.Unlock()

	return result, nil
}

// determineStatus determines the health status based on success rate
func (phc *ProxyHealthChecker) determineStatus(successRate float64) string {
	if successRate >= 0.8 {
		return "healthy"
	} else if successRate >= 0.5 {
		return "degraded"
	} else {
		return "unhealthy"
	}
}

// GetHealthResults returns all health check results
func (phc *ProxyHealthChecker) GetHealthResults() map[string]*HealthCheckResult {
	phc.mu.RLock()
	defer phc.mu.RUnlock()

	results := make(map[string]*HealthCheckResult)
	for id, result := range phc.results {
		results[id] = result
	}
	return results
}
