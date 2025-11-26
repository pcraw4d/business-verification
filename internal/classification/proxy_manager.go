package classification

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ProxyManager manages proxy rotation for web scraping
type ProxyManager struct {
	enabled     bool
	proxies     []string
	currentIdx  int
	proxyMutex  sync.Mutex
	healthCheck map[string]time.Time
	healthMutex sync.RWMutex
}

// NewProxyManager creates a new proxy manager
func NewProxyManager() *ProxyManager {
	enabled := os.Getenv("SCRAPING_USE_PROXIES")
	enabledBool := false // Default to disabled
	if enabled != "" {
		if val, err := strconv.ParseBool(enabled); err == nil {
			enabledBool = val
		}
	}

	// Get proxy list from environment variable (comma-separated)
	proxyList := os.Getenv("SCRAPING_PROXY_LIST")
	var proxies []string
	if proxyList != "" {
		// Parse comma-separated proxy list
		proxyParts := strings.Split(proxyList, ",")
		for _, proxy := range proxyParts {
			proxy = strings.TrimSpace(proxy)
			if proxy != "" {
				// Validate proxy URL format
				if _, err := url.Parse(proxy); err == nil {
					proxies = append(proxies, proxy)
				}
			}
		}
	}

	return &ProxyManager{
		enabled:     enabledBool && len(proxies) > 0,
		proxies:     proxies,
		currentIdx:  0,
		healthCheck: make(map[string]time.Time),
	}
}

// GetProxyForDomain returns a proxy URL for the given domain
// Returns empty string if proxies are disabled or unavailable
func (pm *ProxyManager) GetProxyForDomain(domain string) (string, error) {
	if !pm.enabled {
		return "", nil // Proxies disabled, use direct connection
	}

	if len(pm.proxies) == 0 {
		return "", fmt.Errorf("no proxies configured")
	}

	pm.proxyMutex.Lock()
	defer pm.proxyMutex.Unlock()

	// Rotate proxies (round-robin)
	proxy := pm.proxies[pm.currentIdx]
	pm.currentIdx = (pm.currentIdx + 1) % len(pm.proxies)

	return proxy, nil
}

// GetProxyTransport returns an HTTP transport configured with the proxy
func (pm *ProxyManager) GetProxyTransport(domain string, baseTransport *http.Transport) (*http.Transport, error) {
	proxyURL, err := pm.GetProxyForDomain(domain)
	if err != nil || proxyURL == "" {
		// No proxy or error, return base transport
		return baseTransport, nil
	}

	// Parse proxy URL
	parsedProxy, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	// Create new transport with proxy
	transport := baseTransport.Clone()
	transport.Proxy = http.ProxyURL(parsedProxy)

	return transport, nil
}

// MarkProxyHealthy marks a proxy as healthy
func (pm *ProxyManager) MarkProxyHealthy(proxyURL string) {
	pm.healthMutex.Lock()
	defer pm.healthMutex.Unlock()
	pm.healthCheck[proxyURL] = time.Now()
}

// MarkProxyUnhealthy marks a proxy as unhealthy
func (pm *ProxyManager) MarkProxyUnhealthy(proxyURL string) {
	pm.healthMutex.Lock()
	defer pm.healthMutex.Unlock()
	delete(pm.healthCheck, proxyURL)
}

// IsProxyHealthy checks if a proxy is healthy
func (pm *ProxyManager) IsProxyHealthy(proxyURL string) bool {
	pm.healthMutex.RLock()
	defer pm.healthMutex.RUnlock()

	lastHealthy, exists := pm.healthCheck[proxyURL]
	if !exists {
		return true // Assume healthy if not checked yet
	}

	// Consider proxy healthy if last check was within 5 minutes
	return time.Since(lastHealthy) < 5*time.Minute
}

// GetProxyForDomain is a convenience function using a default manager
func GetProxyForDomain(domain string) (string, error) {
	pm := NewProxyManager()
	return pm.GetProxyForDomain(domain)
}

// IsEnabled checks if proxy management is enabled
func (pm *ProxyManager) IsEnabled() bool {
	return pm.enabled
}

// SetEnabled enables or disables proxy management
func (pm *ProxyManager) SetEnabled(enabled bool) {
	pm.enabled = enabled
}


