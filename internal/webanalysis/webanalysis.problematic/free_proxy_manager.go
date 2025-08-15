package webanalysis

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// FreeProxyManager manages free proxies from public sources
type FreeProxyManager struct {
	proxies        []string
	mu             sync.RWMutex
	lastUpdate     time.Time
	updateInterval time.Duration
}

// NewFreeProxyManager creates a new free proxy manager
func NewFreeProxyManager() *FreeProxyManager {
	return &FreeProxyManager{
		proxies:        make([]string, 0),
		updateInterval: 30 * time.Minute, // Update every 30 minutes
	}
}

// GetProxy returns a random proxy from the pool
func (fpm *FreeProxyManager) GetProxy() (string, error) {
	fpm.mu.RLock()
	defer fpm.mu.RUnlock()

	if len(fpm.proxies) == 0 {
		return "", fmt.Errorf("no proxies available")
	}

	// Simple round-robin for now
	proxy := fpm.proxies[0]
	fpm.proxies = append(fpm.proxies[1:], proxy)

	return proxy, nil
}

// UpdateProxies fetches fresh proxy list from free sources
func (fpm *FreeProxyManager) UpdateProxies() error {
	fpm.mu.Lock()
	defer fpm.mu.Unlock()

	// Fetch from multiple free proxy sources
	var allProxies []string

	// Source 1: FreeProxyList
	if proxies, err := fpm.fetchFromFreeProxyList(); err == nil {
		allProxies = append(allProxies, proxies...)
	}

	// Source 2: ProxyNova
	if proxies, err := fpm.fetchFromProxyNova(); err == nil {
		allProxies = append(allProxies, proxies...)
	}

	// Source 3: Geonode
	if proxies, err := fpm.fetchFromGeonode(); err == nil {
		allProxies = append(allProxies, proxies...)
	}

	// Remove duplicates
	uniqueProxies := fpm.removeDuplicates(allProxies)

	fpm.proxies = uniqueProxies
	fpm.lastUpdate = time.Now()

	return nil
}

// fetchFromFreeProxyList fetches proxies from free-proxy-list.net
func (fpm *FreeProxyManager) fetchFromFreeProxyList() ([]string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://free-proxy-list.net/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the response to extract proxy IPs
	// This is a simplified parser - in production you'd want more robust parsing
	var proxies []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "data-ip=") {
			// Extract IP from data-ip attribute
			parts := strings.Split(line, "data-ip=\"")
			if len(parts) > 1 {
				ip := strings.Split(parts[1], "\"")[0]
				if fpm.isValidIP(ip) {
					proxies = append(proxies, ip+":8080") // Default port
				}
			}
		}
	}

	return proxies, nil
}

// fetchFromProxyNova fetches proxies from proxynova.com
func (fpm *FreeProxyManager) fetchFromProxyNova() ([]string, error) {
	// Similar implementation to above
	// This would parse proxynova.com's proxy list
	return []string{}, nil
}

// fetchFromGeonode fetches proxies from geonode.com
func (fpm *FreeProxyManager) fetchFromGeonode() ([]string, error) {
	// Similar implementation to above
	// This would parse geonode.com's proxy list
	return []string{}, nil
}

// removeDuplicates removes duplicate proxies
func (fpm *FreeProxyManager) removeDuplicates(proxies []string) []string {
	seen := make(map[string]bool)
	var unique []string
	for _, proxy := range proxies {
		if !seen[proxy] {
			seen[proxy] = true
			unique = append(unique, proxy)
		}
	}
	return unique
}

// isValidIP validates if a string looks like an IP address
func (fpm *FreeProxyManager) isValidIP(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
	}
	return true
}

// StartBackgroundUpdate starts background proxy updates
func (fpm *FreeProxyManager) StartBackgroundUpdate() {
	go func() {
		for {
			time.Sleep(fpm.updateInterval)
			fpm.UpdateProxies()
		}
	}()
}
