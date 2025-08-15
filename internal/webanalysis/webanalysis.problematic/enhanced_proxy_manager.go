package webanalysis

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// EnhancedProxyManager provides enterprise-level proxy management
type EnhancedProxyManager struct {
	proxies         map[string]*EnhancedProxy
	geographicPools map[string][]*EnhancedProxy
	healthChecker   *HealthChecker
	rateLimiter     *EnhancedRateLimiter
	rotationEngine  *RotationEngine
	mu              sync.RWMutex
	config          EnhancedProxyConfig
}

// EnhancedProxy represents an enterprise-level proxy
type EnhancedProxy struct {
	ID              string            `json:"id"`
	IP              string            `json:"ip"`
	Port            int               `json:"port"`
	Protocol        string            `json:"protocol"` // http, https, socks5
	Region          string            `json:"region"`
	Country         string            `json:"country"`
	City            string            `json:"city"`
	Provider        string            `json:"provider"`
	Health          bool              `json:"health"`
	LastUsed        time.Time         `json:"last_used"`
	LastCheck       time.Time         `json:"last_check"`
	FailCount       int               `json:"fail_count"`
	SuccessCount    int               `json:"success_count"`
	Latency         time.Duration     `json:"latency"`
	Uptime          float64           `json:"uptime"`
	Bandwidth       int64             `json:"bandwidth"` // bytes per second
	ConcurrentLimit int               `json:"concurrent_limit"`
	CurrentLoad     int               `json:"current_load"`
	Capabilities    map[string]bool   `json:"capabilities"`
	Headers         map[string]string `json:"headers"`
	Cookies         map[string]string `json:"cookies"`
	UserAgents      []string          `json:"user_agents"`
	SSLSupport      bool              `json:"ssl_support"`
	AnonymityLevel  string            `json:"anonymity_level"` // transparent, anonymous, elite
	LastRotation    time.Time         `json:"last_rotation"`
	RotationCount   int               `json:"rotation_count"`
}

// EnhancedProxyConfig holds configuration for enhanced proxy management
type EnhancedProxyConfig struct {
	MaxProxies             int           `json:"max_proxies"`
	HealthCheckInterval    time.Duration `json:"health_check_interval"`
	MaxFailures            int           `json:"max_failures"`
	MinUptime              float64       `json:"min_uptime"`
	MaxLatency             time.Duration `json:"max_latency"`
	RotationInterval       time.Duration `json:"rotation_interval"`
	GeographicDistribution bool          `json:"geographic_distribution"`
	LoadBalancing          bool          `json:"load_balancing"`
	RateLimiting           bool          `json:"rate_limiting"`
	BotDetectionEvasion    bool          `json:"bot_detection_evasion"`
	SSLVerification        bool          `json:"ssl_verification"`
	ConcurrentLimit        int           `json:"concurrent_limit"`
	RequestTimeout         time.Duration `json:"request_timeout"`
	RetryAttempts          int           `json:"retry_attempts"`
	BackoffMultiplier      float64       `json:"backoff_multiplier"`
}

// HealthChecker manages proxy health monitoring
type HealthChecker struct {
	checkURLs     []string
	checkInterval time.Duration
	timeout       time.Duration
	client        *http.Client
	mu            sync.RWMutex
}

// EnhancedRateLimiter manages request rate limiting
type EnhancedRateLimiter struct {
	limits map[string]*RateLimit
	mu     sync.RWMutex
	config RateLimitConfig
}

// RateLimit represents rate limiting for a proxy
type RateLimit struct {
	ProxyID     string
	Requests    int
	Window      time.Duration
	LastRequest time.Time
	mu          sync.Mutex
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	DefaultRequests int           `json:"default_requests"`
	DefaultWindow   time.Duration `json:"default_window"`
	BurstLimit      int           `json:"burst_limit"`
	StrictMode      bool          `json:"strict_mode"`
}

// RotationEngine manages proxy rotation strategies
type RotationEngine struct {
	strategies map[string]RotationStrategy
	current    string
	mu         sync.RWMutex
}

// RotationStrategy defines proxy rotation behavior
type RotationStrategy interface {
	SelectProxy(proxies []*EnhancedProxy, request *ProxyRequest) (*EnhancedProxy, error)
	Name() string
}

// ProxyRequest represents a request for proxy selection
type ProxyRequest struct {
	URL                  string            `json:"url"`
	Method               string            `json:"method"`
	Headers              map[string]string `json:"headers"`
	GeographicPreference string            `json:"geographic_preference"`
	AnonymityLevel       string            `json:"anonymity_level"`
	SSLRequired          bool              `json:"ssl_required"`
	Priority             int               `json:"priority"`
	Timeout              time.Duration     `json:"timeout"`
}

// NewEnhancedProxyManager creates a new enhanced proxy manager
func NewEnhancedProxyManager(config EnhancedProxyConfig) *EnhancedProxyManager {
	if config.MaxProxies == 0 {
		config.MaxProxies = 100
	}
	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = time.Minute * 5
	}
	if config.MaxFailures == 0 {
		config.MaxFailures = 3
	}
	if config.MinUptime == 0 {
		config.MinUptime = 0.95
	}
	if config.MaxLatency == 0 {
		config.MaxLatency = time.Second * 10
	}
	if config.RotationInterval == 0 {
		config.RotationInterval = time.Minute * 30
	}
	if config.ConcurrentLimit == 0 {
		config.ConcurrentLimit = 10
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = time.Second * 30
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.BackoffMultiplier == 0 {
		config.BackoffMultiplier = 2.0
	}

	epm := &EnhancedProxyManager{
		proxies:         make(map[string]*EnhancedProxy),
		geographicPools: make(map[string][]*EnhancedProxy),
		config:          config,
	}

	// Initialize components
	epm.healthChecker = NewHealthChecker(config.HealthCheckInterval, config.RequestTimeout)
	epm.rateLimiter = NewEnhancedRateLimiter(RateLimitConfig{
		DefaultRequests: 100,
		DefaultWindow:   time.Minute,
		BurstLimit:      10,
		StrictMode:      false,
	})
	epm.rotationEngine = NewRotationEngine()

	// Start health checking (disabled for testing)
	// go epm.healthChecker.Start(epm)

	return epm
}

// AddProxy adds a new proxy to the manager
func (epm *EnhancedProxyManager) AddProxy(proxy *EnhancedProxy) error {
	epm.mu.Lock()
	defer epm.mu.Unlock()

	if len(epm.proxies) >= epm.config.MaxProxies {
		return fmt.Errorf("maximum number of proxies reached (%d)", epm.config.MaxProxies)
	}

	// Set default values
	if proxy.ID == "" {
		proxy.ID = fmt.Sprintf("proxy_%d", time.Now().UnixNano())
	}
	if proxy.Protocol == "" {
		proxy.Protocol = "http"
	}
	if proxy.AnonymityLevel == "" {
		proxy.AnonymityLevel = "anonymous"
	}
	if proxy.Capabilities == nil {
		proxy.Capabilities = make(map[string]bool)
	}
	if proxy.Headers == nil {
		proxy.Headers = make(map[string]string)
	}
	if proxy.Cookies == nil {
		proxy.Cookies = make(map[string]string)
	}
	if proxy.UserAgents == nil {
		proxy.UserAgents = []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
		}
	}

	// Add to main proxy pool
	epm.proxies[proxy.ID] = proxy

	// Add to geographic pool if enabled
	if epm.config.GeographicDistribution && proxy.Region != "" {
		epm.geographicPools[proxy.Region] = append(epm.geographicPools[proxy.Region], proxy)
	}

	return nil
}

// GetProxy selects the best proxy for a request
func (epm *EnhancedProxyManager) GetProxy(request *ProxyRequest) (*EnhancedProxy, error) {
	epm.mu.RLock()
	defer epm.mu.RUnlock()

	// Get available proxies
	availableProxies := epm.getAvailableProxies(request)

	if len(availableProxies) == 0 {
		return nil, fmt.Errorf("no available proxies for request")
	}

	// Use rotation engine to select proxy
	proxy, err := epm.rotationEngine.SelectProxy(availableProxies, request)
	if err != nil {
		return nil, fmt.Errorf("failed to select proxy: %w", err)
	}

	// Update proxy usage
	epm.updateProxyUsage(proxy)

	return proxy, nil
}

// getAvailableProxies returns available proxies for a request
func (epm *EnhancedProxyManager) getAvailableProxies(request *ProxyRequest) []*EnhancedProxy {
	var availableProxies []*EnhancedProxy

	for _, proxy := range epm.proxies {
		if !epm.isProxyAvailable(proxy, request) {
			continue
		}

		// Check rate limiting
		if epm.config.RateLimiting {
			if !epm.rateLimiter.AllowRequest(proxy.ID) {
				continue
			}
		}

		// Check geographic preference
		if request.GeographicPreference != "" && proxy.Region != request.GeographicPreference {
			continue
		}

		// Check anonymity level
		if request.AnonymityLevel != "" && proxy.AnonymityLevel != request.AnonymityLevel {
			continue
		}

		// Check SSL requirement
		if request.SSLRequired && !proxy.SSLSupport {
			continue
		}

		availableProxies = append(availableProxies, proxy)
	}

	return availableProxies
}

// isProxyAvailable checks if a proxy is available for use
func (epm *EnhancedProxyManager) isProxyAvailable(proxy *EnhancedProxy, request *ProxyRequest) bool {
	// Check health
	if !proxy.Health {
		return false
	}

	// Check uptime
	if proxy.Uptime < epm.config.MinUptime {
		return false
	}

	// Check latency
	if proxy.Latency > epm.config.MaxLatency {
		return false
	}

	// Check concurrent limit
	if proxy.CurrentLoad >= proxy.ConcurrentLimit {
		return false
	}

	// Check last usage (rotation)
	if time.Since(proxy.LastUsed) < epm.config.RotationInterval {
		return false
	}

	return true
}

// updateProxyUsage updates proxy usage statistics
func (epm *EnhancedProxyManager) updateProxyUsage(proxy *EnhancedProxy) {
	// Update proxy directly since we already have a reference
	proxy.LastUsed = time.Now()
	proxy.CurrentLoad++
	proxy.RotationCount++
}

// ReleaseProxy releases a proxy after use
func (epm *EnhancedProxyManager) ReleaseProxy(proxyID string) {
	epm.mu.Lock()
	defer epm.mu.Unlock()

	if proxy, exists := epm.proxies[proxyID]; exists {
		if proxy.CurrentLoad > 0 {
			proxy.CurrentLoad--
		}
	}
}

// MarkProxySuccess marks a proxy as successful
func (epm *EnhancedProxyManager) MarkProxySuccess(proxyID string, latency time.Duration) {
	epm.mu.Lock()
	defer epm.mu.Unlock()

	if proxy, exists := epm.proxies[proxyID]; exists {
		proxy.SuccessCount++
		proxy.FailCount = 0
		proxy.Latency = latency
		proxy.Uptime = float64(proxy.SuccessCount) / float64(proxy.SuccessCount+proxy.FailCount)
	}
}

// MarkProxyFailure marks a proxy as failed
func (epm *EnhancedProxyManager) MarkProxyFailure(proxyID string) {
	epm.mu.Lock()
	defer epm.mu.Unlock()

	if proxy, exists := epm.proxies[proxyID]; exists {
		proxy.FailCount++
		proxy.Uptime = float64(proxy.SuccessCount) / float64(proxy.SuccessCount+proxy.FailCount)

		// Mark as unhealthy if too many failures
		if proxy.FailCount >= epm.config.MaxFailures {
			proxy.Health = false
		}
	}
}

// GetStats returns proxy manager statistics
func (epm *EnhancedProxyManager) GetStats() map[string]interface{} {
	epm.mu.RLock()
	defer epm.mu.RUnlock()

	totalProxies := len(epm.proxies)
	healthyProxies := 0
	totalLatency := time.Duration(0)
	activeLoad := 0

	for _, proxy := range epm.proxies {
		if proxy.Health {
			healthyProxies++
		}
		totalLatency += proxy.Latency
		activeLoad += proxy.CurrentLoad
	}

	avgLatency := time.Duration(0)
	if healthyProxies > 0 {
		avgLatency = totalLatency / time.Duration(healthyProxies)
	}

	return map[string]interface{}{
		"total_proxies":    totalProxies,
		"healthy_proxies":  healthyProxies,
		"health_ratio":     float64(healthyProxies) / float64(totalProxies),
		"average_latency":  avgLatency.String(),
		"active_load":      activeLoad,
		"geographic_pools": len(epm.geographicPools),
		"rotation_count":   epm.getTotalRotationCount(),
	}
}

// getTotalRotationCount returns total rotation count across all proxies
func (epm *EnhancedProxyManager) getTotalRotationCount() int {
	total := 0
	for _, proxy := range epm.proxies {
		total += proxy.RotationCount
	}
	return total
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(interval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		checkURLs: []string{
			"http://httpbin.org/ip",
			"https://httpbin.org/ip",
			"http://ip-api.com/json",
		},
		checkInterval: interval,
		timeout:       timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Start begins health checking
func (hc *HealthChecker) Start(epm *EnhancedProxyManager) {
	ticker := time.NewTicker(hc.checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		hc.checkAllProxies(epm)
	}
}

// checkAllProxies checks health of all proxies
func (hc *HealthChecker) checkAllProxies(epm *EnhancedProxyManager) {
	epm.mu.RLock()
	proxies := make([]*EnhancedProxy, 0, len(epm.proxies))
	for _, proxy := range epm.proxies {
		proxies = append(proxies, proxy)
	}
	epm.mu.RUnlock()

	for _, proxy := range proxies {
		go hc.checkProxy(proxy, epm)
	}
}

// checkProxy checks health of a single proxy
func (hc *HealthChecker) checkProxy(proxy *EnhancedProxy, epm *EnhancedProxyManager) {
	start := time.Now()

	// Create proxy URL
	proxyURL := fmt.Sprintf("%s://%s:%d", proxy.Protocol, proxy.IP, proxy.Port)

	// Create transport with proxy
	transport := &http.Transport{
		Proxy: func(_ *http.Request) (*url.URL, error) {
			return url.Parse(proxyURL)
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   hc.timeout,
	}

	// Test with a simple request
	resp, err := client.Get(hc.checkURLs[0])
	if err != nil {
		epm.MarkProxyFailure(proxy.ID)
		return
	}
	defer resp.Body.Close()

	latency := time.Since(start)

	if resp.StatusCode == 200 {
		epm.MarkProxySuccess(proxy.ID, latency)
		proxy.Health = true
		proxy.LastCheck = time.Now()
	} else {
		epm.MarkProxyFailure(proxy.ID)
	}
}

// NewEnhancedRateLimiter creates a new enhanced rate limiter
func NewEnhancedRateLimiter(config RateLimitConfig) *EnhancedRateLimiter {
	return &EnhancedRateLimiter{
		limits: make(map[string]*RateLimit),
		config: config,
	}
}

// AllowRequest checks if a request is allowed for a proxy
func (rl *EnhancedRateLimiter) AllowRequest(proxyID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, exists := rl.limits[proxyID]
	if !exists {
		limit = &RateLimit{
			ProxyID:  proxyID,
			Requests: 0,
			Window:   rl.config.DefaultWindow,
		}
		rl.limits[proxyID] = limit
	}

	limit.mu.Lock()
	defer limit.mu.Unlock()

	now := time.Now()

	// Reset window if expired
	if now.Sub(limit.LastRequest) > limit.Window {
		limit.Requests = 0
	}

	// Check if request is allowed
	if limit.Requests >= rl.config.DefaultRequests {
		return false
	}

	limit.Requests++
	limit.LastRequest = now
	return true
}

// NewRotationEngine creates a new rotation engine
func NewRotationEngine() *RotationEngine {
	re := &RotationEngine{
		strategies: make(map[string]RotationStrategy),
		current:    "round_robin",
	}

	// Register default strategies
	re.RegisterStrategy(&RoundRobinStrategy{})
	re.RegisterStrategy(&LoadBalancedStrategy{})
	re.RegisterStrategy(&GeographicStrategy{})
	re.RegisterStrategy(&LatencyBasedStrategy{})

	return re
}

// RegisterStrategy registers a rotation strategy
func (re *RotationEngine) RegisterStrategy(strategy RotationStrategy) {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.strategies[strategy.Name()] = strategy
}

// SelectProxy selects a proxy using the current strategy
func (re *RotationEngine) SelectProxy(proxies []*EnhancedProxy, request *ProxyRequest) (*EnhancedProxy, error) {
	re.mu.RLock()
	strategy, exists := re.strategies[re.current]
	re.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("strategy %s not found", re.current)
	}

	return strategy.SelectProxy(proxies, request)
}

// RoundRobinStrategy implements round-robin proxy selection
type RoundRobinStrategy struct {
	lastIndex int
	mu        sync.Mutex
}

func (rrs *RoundRobinStrategy) Name() string {
	return "round_robin"
}

func (rrs *RoundRobinStrategy) SelectProxy(proxies []*EnhancedProxy, request *ProxyRequest) (*EnhancedProxy, error) {
	rrs.mu.Lock()
	defer rrs.mu.Unlock()

	if len(proxies) == 0 {
		return nil, fmt.Errorf("no proxies available")
	}

	rrs.lastIndex = (rrs.lastIndex + 1) % len(proxies)
	return proxies[rrs.lastIndex], nil
}

// LoadBalancedStrategy implements load-balanced proxy selection
type LoadBalancedStrategy struct{}

func (lbs *LoadBalancedStrategy) Name() string {
	return "load_balanced"
}

func (lbs *LoadBalancedStrategy) SelectProxy(proxies []*EnhancedProxy, request *ProxyRequest) (*EnhancedProxy, error) {
	if len(proxies) == 0 {
		return nil, fmt.Errorf("no proxies available")
	}

	// Select proxy with lowest load
	var selectedProxy *EnhancedProxy
	minLoad := int(^uint(0) >> 1) // Max int

	for _, proxy := range proxies {
		load := proxy.CurrentLoad
		if load < minLoad {
			minLoad = load
			selectedProxy = proxy
		}
	}

	return selectedProxy, nil
}

// GeographicStrategy implements geographic-based proxy selection
type GeographicStrategy struct{}

func (gs *GeographicStrategy) Name() string {
	return "geographic"
}

func (gs *GeographicStrategy) SelectProxy(proxies []*EnhancedProxy, request *ProxyRequest) (*EnhancedProxy, error) {
	if len(proxies) == 0 {
		return nil, fmt.Errorf("no proxies available")
	}

	// If geographic preference is specified, prioritize it
	if request.GeographicPreference != "" {
		for _, proxy := range proxies {
			if proxy.Region == request.GeographicPreference {
				return proxy, nil
			}
		}
	}

	// Otherwise, select randomly
	return proxies[rand.Intn(len(proxies))], nil
}

// LatencyBasedStrategy implements latency-based proxy selection
type LatencyBasedStrategy struct{}

func (lbs *LatencyBasedStrategy) Name() string {
	return "latency_based"
}

func (lbs *LatencyBasedStrategy) SelectProxy(proxies []*EnhancedProxy, request *ProxyRequest) (*EnhancedProxy, error) {
	if len(proxies) == 0 {
		return nil, fmt.Errorf("no proxies available")
	}

	// Select proxy with lowest latency
	var selectedProxy *EnhancedProxy
	minLatency := time.Duration(^uint64(0) >> 1) // Max duration

	for _, proxy := range proxies {
		if proxy.Latency < minLatency {
			minLatency = proxy.Latency
			selectedProxy = proxy
		}
	}

	return selectedProxy, nil
}
