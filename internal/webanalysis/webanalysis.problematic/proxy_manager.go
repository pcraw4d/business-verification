package webanalysis

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Proxy represents a single proxy server
type Proxy struct {
	ID        string        `json:"id"`
	IP        string        `json:"ip"`
	Port      int           `json:"port"`
	Region    string        `json:"region"`
	Provider  string        `json:"provider"`
	Health    bool          `json:"health"`
	LastUsed  time.Time     `json:"last_used"`
	LastCheck time.Time     `json:"last_check"`
	FailCount int           `json:"fail_count"`
	Latency   time.Duration `json:"latency"`
}

// ProxyManager manages a pool of proxy servers
type ProxyManager struct {
	proxies []*Proxy
	current int
	mu      sync.RWMutex
	config  ProxyConfig
}

// ProxyConfig holds configuration for proxy management
type ProxyConfig struct {
	MaxFailures    int           `json:"max_failures"`
	HealthCheckURL string        `json:"health_check_url"`
	CheckInterval  time.Duration `json:"check_interval"`
	Timeout        time.Duration `json:"timeout"`
	MinLatency     time.Duration `json:"min_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
}

// NewProxyManager creates a new proxy manager with default configuration
func NewProxyManager() *ProxyManager {
	return &ProxyManager{
		proxies: make([]*Proxy, 0),
		config: ProxyConfig{
			MaxFailures:    3,
			HealthCheckURL: "https://httpbin.org/ip",
			CheckInterval:  time.Minute * 5,
			Timeout:        time.Second * 10,
			MinLatency:     time.Millisecond * 50,
			MaxLatency:     time.Second * 2,
		},
	}
}

// AddProxy adds a new proxy to the pool
func (pm *ProxyManager) AddProxy(proxy *Proxy) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Set default values
	if proxy.ID == "" {
		proxy.ID = fmt.Sprintf("proxy-%d", len(pm.proxies))
	}
	proxy.Health = true
	proxy.LastCheck = time.Now()

	pm.proxies = append(pm.proxies, proxy)
}

// GetNextProxy returns the next available proxy using round-robin with health checking
func (pm *ProxyManager) GetNextProxy() (*Proxy, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.proxies) == 0 {
		return nil, fmt.Errorf("no proxies available")
	}

	// Try to find a healthy proxy
	attempts := 0
	maxAttempts := len(pm.proxies) * 2 // Allow multiple passes through the list

	for attempts < maxAttempts {
		proxy := pm.proxies[pm.current]
		pm.current = (pm.current + 1) % len(pm.proxies)

		// Check if proxy is healthy and hasn't been used recently
		if proxy.Health &&
			time.Since(proxy.LastUsed) > time.Second*30 &&
			time.Since(proxy.LastCheck) < pm.config.CheckInterval {

			proxy.LastUsed = time.Now()
			return proxy, nil
		}

		attempts++
	}

	return nil, fmt.Errorf("no healthy proxies available")
}

// CheckHealth performs health check on a proxy
func (pm *ProxyManager) CheckHealth(proxy *Proxy) error {
	start := time.Now()

	// Create HTTP client with proxy
	proxyURL, err := url.Parse(fmt.Sprintf("http://%s:%d", proxy.IP, proxy.Port))
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: pm.config.Timeout,
	}

	// Make health check request
	resp, err := client.Get(pm.config.HealthCheckURL)
	if err != nil {
		pm.markProxyUnhealthy(proxy)
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		pm.markProxyUnhealthy(proxy)
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	// Update proxy health and latency
	latency := time.Since(start)
	pm.mu.Lock()
	proxy.Health = true
	proxy.LastCheck = time.Now()
	proxy.Latency = latency
	proxy.FailCount = 0
	pm.mu.Unlock()

	return nil
}

// markProxyUnhealthy marks a proxy as unhealthy
func (pm *ProxyManager) markProxyUnhealthy(proxy *Proxy) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	proxy.FailCount++
	if proxy.FailCount >= pm.config.MaxFailures {
		proxy.Health = false
	}
	proxy.LastCheck = time.Now()
}

// MarkProxyUnhealthy is a public method for marking proxies unhealthy (for testing)
func (pm *ProxyManager) MarkProxyUnhealthy(proxy *Proxy) {
	pm.markProxyUnhealthy(proxy)
}

// StartHealthChecker starts the background health checking routine
func (pm *ProxyManager) StartHealthChecker(ctx context.Context) {
	ticker := time.NewTicker(pm.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pm.checkAllProxies()
		}
	}
}

// checkAllProxies checks health of all proxies
func (pm *ProxyManager) checkAllProxies() {
	pm.mu.RLock()
	proxies := make([]*Proxy, len(pm.proxies))
	copy(proxies, pm.proxies)
	pm.mu.RUnlock()

	for _, proxy := range proxies {
		go func(p *Proxy) {
			if err := pm.CheckHealth(p); err != nil {
				// Log error but don't fail the entire check
				fmt.Printf("Health check failed for proxy %s: %v\n", p.ID, err)
			}
		}(proxy)
	}
}

// GetStats returns statistics about the proxy pool
func (pm *ProxyManager) GetStats() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	healthy := 0
	totalLatency := time.Duration(0)
	healthyCount := 0

	for _, proxy := range pm.proxies {
		if proxy.Health {
			healthy++
			if proxy.Latency > 0 {
				totalLatency += proxy.Latency
				healthyCount++
			}
		}
	}

	avgLatency := time.Duration(0)
	if healthyCount > 0 {
		avgLatency = totalLatency / time.Duration(healthyCount)
	}

	return map[string]interface{}{
		"total_proxies":     len(pm.proxies),
		"healthy_proxies":   healthy,
		"unhealthy_proxies": len(pm.proxies) - healthy,
		"health_percentage": float64(healthy) / float64(len(pm.proxies)) * 100,
		"average_latency":   avgLatency.String(),
	}
}

// GetProxyByRegion returns a proxy from a specific region
func (pm *ProxyManager) GetProxyByRegion(region string) (*Proxy, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, proxy := range pm.proxies {
		if proxy.Region == region && proxy.Health {
			return proxy, nil
		}
	}

	return nil, fmt.Errorf("no healthy proxy available in region %s", region)
}

// GetRandomProxy returns a random healthy proxy
func (pm *ProxyManager) GetRandomProxy() (*Proxy, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var healthyProxies []*Proxy
	for _, proxy := range pm.proxies {
		if proxy.Health {
			healthyProxies = append(healthyProxies, proxy)
		}
	}

	if len(healthyProxies) == 0 {
		return nil, fmt.Errorf("no healthy proxies available")
	}

	// Return random proxy
	randomIndex := rand.Intn(len(healthyProxies))
	proxy := healthyProxies[randomIndex]
	proxy.LastUsed = time.Now()

	return proxy, nil
}
