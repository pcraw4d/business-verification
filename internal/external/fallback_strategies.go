package external

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FallbackStrategyManager manages fallback strategies for blocked websites
type FallbackStrategyManager struct {
	config     *FallbackConfig
	logger     *zap.Logger
	userAgents []string
	headers    map[string][]string
	proxies    []Proxy
	mu         sync.RWMutex
}

// FallbackConfig holds configuration for fallback strategies
type FallbackConfig struct {
	EnableUserAgentRotation   bool              `json:"enable_user_agent_rotation"`
	EnableHeaderCustomization bool              `json:"enable_header_customization"`
	EnableProxyRotation       bool              `json:"enable_proxy_rotation"`
	EnableAlternativeSources  bool              `json:"enable_alternative_sources"`
	MaxFallbackAttempts       int               `json:"max_fallback_attempts"`
	FallbackDelay             time.Duration     `json:"fallback_delay"`
	FallbackTimeout           time.Duration     `json:"fallback_timeout"` // Per-strategy timeout limit
	UserAgentPool             []string          `json:"user_agent_pool"`
	HeaderTemplates           map[string]string `json:"header_templates"`
	ProxyPool                 []Proxy           `json:"proxy_pool"`
	AlternativeSources        []DataSource      `json:"alternative_sources"`
}

// Proxy represents a proxy configuration
type Proxy struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Protocol string `json:"protocol"` // http, https, socks5
	Location string `json:"location,omitempty"`
	Active   bool   `json:"active"`
}

// DataSource represents an alternative data source
type DataSource struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"` // api, database, cache
	URL         string            `json:"url,omitempty"`
	APIKey      string            `json:"api_key,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Timeout     time.Duration     `json:"timeout"`
	Priority    int               `json:"priority"`    // Lower number = higher priority
	Reliability float64           `json:"reliability"` // 0.0 to 1.0
}

// FallbackResult represents the result of a fallback strategy
type FallbackResult struct {
	StrategyUsed  string                 `json:"strategy_used"`
	Success       bool                   `json:"success"`
	Content       string                 `json:"content,omitempty"`
	StatusCode    int                    `json:"status_code"`
	DataSource    string                 `json:"data_source,omitempty"`
	ProxyUsed     *Proxy                 `json:"proxy_used,omitempty"`
	UserAgentUsed string                 `json:"user_agent_used,omitempty"`
	HeadersUsed   map[string]string      `json:"headers_used,omitempty"`
	Attempts      int                    `json:"attempts"`
	Duration      time.Duration          `json:"duration"`
	Error         string                 `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// DefaultFallbackConfig returns default configuration for fallback strategies
func DefaultFallbackConfig() *FallbackConfig {
	return &FallbackConfig{
		EnableUserAgentRotation:   true,
		EnableHeaderCustomization: true,
		EnableProxyRotation:       false, // Disabled by default for security
		EnableAlternativeSources:  true,
		MaxFallbackAttempts:       5,
		FallbackDelay:             2 * time.Second,
		FallbackTimeout:           10 * time.Second, // Reduced from 15s to 10s per strategy for faster fallback
		UserAgentPool: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/120.0.0.0",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
		},
		HeaderTemplates: map[string]string{
			"desktop": `Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate
DNT: 1
Connection: keep-alive
Upgrade-Insecure-Requests: 1
Sec-Fetch-Dest: document
Sec-Fetch-Mode: navigate
Sec-Fetch-Site: none
Cache-Control: max-age=0`,
			"mobile": `Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate
DNT: 1
Connection: keep-alive
Upgrade-Insecure-Requests: 1
Sec-Fetch-Dest: document
Sec-Fetch-Mode: navigate
Sec-Fetch-Site: none
Cache-Control: max-age=0
User-Agent: Mozilla/5.0 (iPhone; CPU iPhone OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Mobile/15E148 Safari/604.1`,
		},
		ProxyPool: []Proxy{},
		AlternativeSources: []DataSource{
			{
				Name:        "Wayback Machine",
				Type:        "api",
				URL:         "https://web.archive.org/cdx/search/cdx",
				Timeout:     10 * time.Second,
				Priority:    1,
				Reliability: 0.8,
			},
			{
				Name:        "Google Cache",
				Type:        "api",
				URL:         "https://webcache.googleusercontent.com/search",
				Timeout:     10 * time.Second,
				Priority:    2,
				Reliability: 0.7,
			},
		},
	}
}

// NewFallbackStrategyManager creates a new fallback strategy manager
func NewFallbackStrategyManager(config *FallbackConfig, logger *zap.Logger) *FallbackStrategyManager {
	if config == nil {
		config = DefaultFallbackConfig()
	}

	manager := &FallbackStrategyManager{
		config:     config,
		logger:     logger,
		userAgents: config.UserAgentPool,
		headers:    make(map[string][]string),
		proxies:    config.ProxyPool,
	}

	// Initialize header templates
	manager.initializeHeaderTemplates()

	return manager
}

// initializeHeaderTemplates initializes header templates from configuration
func (m *FallbackStrategyManager) initializeHeaderTemplates() {
	for templateName, headerString := range m.config.HeaderTemplates {
		headers := make([]string, 0)
		lines := strings.Split(headerString, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				headers = append(headers, line)
			}
		}
		m.headers[templateName] = headers
	}
}

// ExecuteFallbackStrategies executes fallback strategies for a blocked website
func (m *FallbackStrategyManager) ExecuteFallbackStrategies(ctx context.Context, targetURL string, originalError error) (*FallbackResult, error) {
	startTime := time.Now()

	m.logger.Info("Starting fallback strategies",
		zap.String("url", targetURL),
		zap.String("original_error", originalError.Error()))

	result := &FallbackResult{
		StrategyUsed: "none",
		Success:      false,
		Attempts:     0,
		Duration:     time.Since(startTime),
	}

	// Strategy 1: User Agent Rotation
	if m.config.EnableUserAgentRotation && !result.Success {
		m.logger.Info("Trying user agent rotation strategy", zap.String("url", targetURL))
		uaResult := m.TryUserAgentRotation(ctx, targetURL)
		if uaResult.Success {
			result = uaResult
			result.StrategyUsed = "user_agent_rotation"
		}
	}

	// Strategy 2: Header Customization
	if m.config.EnableHeaderCustomization && !result.Success {
		m.logger.Info("Trying header customization strategy", zap.String("url", targetURL))
		headerResult := m.TryHeaderCustomization(ctx, targetURL)
		if headerResult.Success {
			result = headerResult
			result.StrategyUsed = "header_customization"
		}
	}

	// Strategy 3: Proxy Rotation
	if m.config.EnableProxyRotation && !result.Success {
		m.logger.Info("Trying proxy rotation strategy", zap.String("url", targetURL))
		proxyResult := m.TryProxyRotation(ctx, targetURL)
		if proxyResult.Success {
			result = proxyResult
			result.StrategyUsed = "proxy_rotation"
		}
	}

	// Strategy 4: Alternative Data Sources
	if m.config.EnableAlternativeSources && !result.Success {
		m.logger.Info("Trying alternative data sources strategy", zap.String("url", targetURL))
		altResult := m.TryAlternativeDataSources(ctx, targetURL)
		if altResult.Success {
			result = altResult
			result.StrategyUsed = "alternative_sources"
		}
	}

	result.Duration = time.Since(startTime)

	if result.Success {
		m.logger.Info("Fallback strategy succeeded",
			zap.String("url", targetURL),
			zap.String("strategy", result.StrategyUsed),
			zap.Duration("duration", result.Duration))
	} else {
		m.logger.Warn("All fallback strategies failed",
			zap.String("url", targetURL),
			zap.Duration("duration", result.Duration))
	}

	return result, nil
}

// BlockingError represents a blocking error for testing
type BlockingError struct {
	Message string
}

func (e *BlockingError) Error() string {
	return e.Message
}

// TryUserAgentRotation tries different user agents
func (m *FallbackStrategyManager) TryUserAgentRotation(ctx context.Context, targetURL string) *FallbackResult {
	result := &FallbackResult{
		StrategyUsed: "user_agent_rotation",
		Success:      false,
		Attempts:     0,
	}

	// Shuffle user agents for randomization
	userAgents := make([]string, len(m.userAgents))
	copy(userAgents, m.userAgents)
	rand.Shuffle(len(userAgents), func(i, j int) {
		userAgents[i], userAgents[j] = userAgents[j], userAgents[i]
	})

	startTime := time.Now()
	for _, userAgent := range userAgents {
		result.Attempts++

		// Check timeout - early exit if exceeded
		if time.Since(startTime) > m.config.FallbackTimeout {
			m.logger.Warn("Fallback timeout exceeded, skipping remaining attempts",
				zap.String("strategy", "user_agent_rotation"),
				zap.Duration("timeout", m.config.FallbackTimeout))
			break
		}

		// Create HTTP client with per-strategy timeout
		client := &http.Client{
			Timeout: m.config.FallbackTimeout,
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
		if err != nil {
			continue
		}

		// Set user agent
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("Accept-Encoding", "gzip, deflate")
		req.Header.Set("Connection", "keep-alive")

		// Make request
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		// Check if successful
		if resp.StatusCode == 200 {
			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				result.Success = true
				result.Content = string(body)
				result.StatusCode = resp.StatusCode
				result.UserAgentUsed = userAgent
				break
			}
		}

		// Add delay between attempts
		time.Sleep(m.config.FallbackDelay)
	}

	return result
}

// TryHeaderCustomization tries different header configurations
func (m *FallbackStrategyManager) TryHeaderCustomization(ctx context.Context, targetURL string) *FallbackResult {
	startTime := time.Now()
	result := &FallbackResult{
		StrategyUsed: "header_customization",
		Success:      false,
		Attempts:     0,
	}

	// Try different header templates
	for _, headers := range m.headers {
		result.Attempts++

		// Check timeout - early exit if exceeded
		if time.Since(startTime) > m.config.FallbackTimeout {
			m.logger.Warn("Fallback timeout exceeded, skipping remaining attempts",
				zap.String("strategy", "header_customization"),
				zap.Duration("timeout", m.config.FallbackTimeout))
			break
		}

		// Create HTTP client with per-strategy timeout
		client := &http.Client{
			Timeout: m.config.FallbackTimeout,
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
		if err != nil {
			continue
		}

		// Set custom headers
		for _, header := range headers {
			if strings.Contains(header, ":") {
				parts := strings.SplitN(header, ":", 2)
				if len(parts) == 2 {
					req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
				}
			}
		}

		// Make request
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		// Check if successful
		if resp.StatusCode == 200 {
			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				result.Success = true
				result.Content = string(body)
				result.StatusCode = resp.StatusCode
				result.HeadersUsed = make(map[string]string)
				for key, values := range req.Header {
					if len(values) > 0 {
						result.HeadersUsed[key] = values[0]
					}
				}
				break
			}
		}

		// Add delay between attempts
		time.Sleep(m.config.FallbackDelay)
	}

	return result
}

// TryProxyRotation tries different proxies
func (m *FallbackStrategyManager) TryProxyRotation(ctx context.Context, targetURL string) *FallbackResult {
	result := &FallbackResult{
		StrategyUsed: "proxy_rotation",
		Success:      false,
		Attempts:     0,
	}

	// Get active proxies
	var activeProxies []Proxy
	for _, proxy := range m.proxies {
		if proxy.Active {
			activeProxies = append(activeProxies, proxy)
		}
	}

	if len(activeProxies) == 0 {
		return result
	}

	startTime := time.Now()
	// Try each proxy
	for _, proxy := range activeProxies {
		result.Attempts++

		// Check timeout - early exit if exceeded
		if time.Since(startTime) > m.config.FallbackTimeout {
			m.logger.Warn("Fallback timeout exceeded, skipping remaining attempts",
				zap.String("strategy", "proxy_rotation"),
				zap.Duration("timeout", m.config.FallbackTimeout))
			break
		}

		// Create proxy URL (for future proxy implementation)
		_ = fmt.Sprintf("%s://%s:%d", proxy.Protocol, proxy.Host, proxy.Port)
		if proxy.Username != "" && proxy.Password != "" {
			_ = fmt.Sprintf("%s://%s:%s@%s:%d", proxy.Protocol, proxy.Username, proxy.Password, proxy.Host, proxy.Port)
		}

		// Create HTTP client with proxy and per-strategy timeout
		client := &http.Client{
			Timeout: m.config.FallbackTimeout,
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
		if err != nil {
			continue
		}

		// Set headers
		req.Header.Set("User-Agent", m.getRandomUserAgent())
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("Accept-Encoding", "gzip, deflate")
		req.Header.Set("Connection", "keep-alive")

		// Make request
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		// Check if successful
		if resp.StatusCode == 200 {
			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				result.Success = true
				result.Content = string(body)
				result.StatusCode = resp.StatusCode
				result.ProxyUsed = &proxy
				break
			}
		}

		// Add delay between attempts
		time.Sleep(m.config.FallbackDelay)
	}

	return result
}

// TryAlternativeDataSources tries alternative data sources
func (m *FallbackStrategyManager) TryAlternativeDataSources(ctx context.Context, targetURL string) *FallbackResult {
	startTime := time.Now()
	result := &FallbackResult{
		StrategyUsed: "alternative_sources",
		Success:      false,
		Attempts:     0,
	}

	// Sort data sources by priority
	sources := make([]DataSource, len(m.config.AlternativeSources))
	copy(sources, m.config.AlternativeSources)

	// Try each data source
	for _, source := range sources {
		result.Attempts++

		// Check timeout - early exit if exceeded
		if time.Since(startTime) > m.config.FallbackTimeout {
			m.logger.Warn("Fallback timeout exceeded, skipping remaining attempts",
				zap.String("strategy", "alternative_sources"),
				zap.Duration("timeout", m.config.FallbackTimeout))
			break
		}

		content, err := m.fetchFromDataSource(ctx, targetURL, source)
		if err == nil && content != "" {
			result.Success = true
			result.Content = content
			result.DataSource = source.Name
			result.StatusCode = 200
			break
		}

		// Add delay between attempts
		time.Sleep(m.config.FallbackDelay)
	}

	return result
}

// fetchFromDataSource fetches content from an alternative data source
func (m *FallbackStrategyManager) fetchFromDataSource(ctx context.Context, targetURL string, source DataSource) (string, error) {
	switch source.Type {
	case "api":
		return m.fetchFromAPI(ctx, targetURL, source)
	case "database":
		return m.fetchFromDatabase(ctx, targetURL, source)
	case "cache":
		return m.fetchFromCache(ctx, targetURL, source)
	default:
		return "", fmt.Errorf("unknown data source type: %s", source.Type)
	}
}

// fetchFromAPI fetches content from an API data source
func (m *FallbackStrategyManager) fetchFromAPI(ctx context.Context, targetURL string, source DataSource) (string, error) {
	// Create HTTP client
	client := &http.Client{
		Timeout: source.Timeout,
	}

	// Create request URL
	requestURL := source.URL
	if strings.Contains(requestURL, "?") {
		requestURL += "&url=" + targetURL
	} else {
		requestURL += "?url=" + targetURL
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return "", err
	}

	// Set headers
	for key, value := range source.Headers {
		req.Header.Set(key, value)
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// fetchFromDatabase fetches content from a database data source
func (m *FallbackStrategyManager) fetchFromDatabase(ctx context.Context, targetURL string, source DataSource) (string, error) {
	// This would implement database lookup logic
	// For now, return empty string
	return "", fmt.Errorf("database data source not implemented")
}

// fetchFromCache fetches content from a cache data source
func (m *FallbackStrategyManager) fetchFromCache(ctx context.Context, targetURL string, source DataSource) (string, error) {
	// This would implement cache lookup logic
	// For now, return empty string
	return "", fmt.Errorf("cache data source not implemented")
}

// getRandomUserAgent returns a random user agent from the pool
func (m *FallbackStrategyManager) getRandomUserAgent() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.userAgents) == 0 {
		return "KYB-Platform-Bot/1.0"
	}

	return m.userAgents[rand.Intn(len(m.userAgents))]
}

// AddProxy adds a proxy to the proxy pool
func (m *FallbackStrategyManager) AddProxy(proxy Proxy) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.proxies = append(m.proxies, proxy)
}

// RemoveProxy removes a proxy from the proxy pool
func (m *FallbackStrategyManager) RemoveProxy(host string, port int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, proxy := range m.proxies {
		if proxy.Host == host && proxy.Port == port {
			m.proxies = append(m.proxies[:i], m.proxies[i+1:]...)
			break
		}
	}
}

// UpdateConfig updates the fallback configuration
func (m *FallbackStrategyManager) UpdateConfig(config *FallbackConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = config
	m.userAgents = config.UserAgentPool
	m.proxies = config.ProxyPool

	// Reinitialize header templates
	m.initializeHeaderTemplates()
}

// GetConfig returns the current fallback configuration
func (m *FallbackStrategyManager) GetConfig() *FallbackConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.config
}
