package external

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ConnectionPool manages HTTP connection pools for external API calls
type ConnectionPool struct {
	clients map[string]*http.Client
	configs map[string]*PoolConfig
	logger  *zap.Logger
	mu      sync.RWMutex
	stats   *PoolStats
}

// PoolConfig represents configuration for a connection pool
type PoolConfig struct {
	MaxIdleConns        int           `json:"max_idle_conns"`
	MaxIdleConnsPerHost int           `json:"max_idle_conns_per_host"`
	MaxConnsPerHost     int           `json:"max_conns_per_host"`
	IdleConnTimeout     time.Duration `json:"idle_conn_timeout"`
	DisableKeepAlives   bool          `json:"disable_keep_alives"`
	Timeout             time.Duration `json:"timeout"`
	MaxRetries          int           `json:"max_retries"`
	RetryDelay          time.Duration `json:"retry_delay"`
}

// PoolStats represents statistics for connection pools
type PoolStats struct {
	TotalRequests      int64                     `json:"total_requests"`
	SuccessfulRequests int64                     `json:"successful_requests"`
	FailedRequests     int64                     `json:"failed_requests"`
	RetryAttempts      int64                     `json:"retry_attempts"`
	AverageLatency     time.Duration             `json:"average_latency"`
	TotalLatency       time.Duration             `json:"total_latency"`
	LastRequest        time.Time                 `json:"last_request"`
	ProviderStats      map[string]*ProviderStats `json:"provider_stats"`
}

// ProviderStats represents statistics for a specific provider
type ProviderStats struct {
	Requests           int64            `json:"requests"`
	SuccessfulRequests int64            `json:"successful_requests"`
	FailedRequests     int64            `json:"failed_requests"`
	RetryAttempts      int64            `json:"retry_attempts"`
	AverageLatency     time.Duration    `json:"average_latency"`
	TotalLatency       time.Duration    `json:"total_latency"`
	LastRequest        time.Time        `json:"last_request"`
	ErrorCounts        map[string]int64 `json:"error_counts"`
}

// NewConnectionPool creates a new connection pool manager
func NewConnectionPool(logger *zap.Logger) *ConnectionPool {
	return &ConnectionPool{
		clients: make(map[string]*http.Client),
		configs: make(map[string]*PoolConfig),
		logger:  logger,
		stats: &PoolStats{
			ProviderStats: make(map[string]*ProviderStats),
		},
	}
}

// AddProvider adds a new provider with its connection pool configuration
func (cp *ConnectionPool) AddProvider(provider string, config *PoolConfig) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if config == nil {
		return fmt.Errorf("pool config cannot be nil for provider %s", provider)
	}

	// Set defaults
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 100
	}
	if config.MaxIdleConnsPerHost == 0 {
		config.MaxIdleConnsPerHost = 10
	}
	if config.MaxConnsPerHost == 0 {
		config.MaxConnsPerHost = 50
	}
	if config.IdleConnTimeout == 0 {
		config.IdleConnTimeout = 90 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	// Create HTTP client with connection pool
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		MaxConnsPerHost:     config.MaxConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableKeepAlives:   config.DisableKeepAlives,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	cp.clients[provider] = client
	cp.configs[provider] = config
	cp.stats.ProviderStats[provider] = &ProviderStats{
		ErrorCounts: make(map[string]int64),
	}

	cp.logger.Info("Added provider to connection pool",
		zap.String("provider", provider),
		zap.Int("max_idle_conns", config.MaxIdleConns),
		zap.Int("max_idle_conns_per_host", config.MaxIdleConnsPerHost),
		zap.Int("max_conns_per_host", config.MaxConnsPerHost),
		zap.Duration("idle_conn_timeout", config.IdleConnTimeout),
		zap.Duration("timeout", config.Timeout))

	return nil
}

// GetClient returns the HTTP client for a specific provider
func (cp *ConnectionPool) GetClient(provider string) (*http.Client, error) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	client, exists := cp.clients[provider]
	if !exists {
		return nil, fmt.Errorf("provider %s not found in connection pool", provider)
	}

	return client, nil
}

// DoRequest performs an HTTP request with retry logic and statistics tracking
func (cp *ConnectionPool) DoRequest(ctx context.Context, provider string, req *http.Request) (*http.Response, error) {
	client, err := cp.GetClient(provider)
	if err != nil {
		return nil, err
	}

	config := cp.configs[provider]
	start := time.Now()

	// Track request
	cp.mu.Lock()
	cp.stats.TotalRequests++
	cp.stats.LastRequest = time.Now()
	providerStats := cp.stats.ProviderStats[provider]
	providerStats.Requests++
	providerStats.LastRequest = time.Now()
	cp.mu.Unlock()

	var resp *http.Response
	var lastErr error

	// Retry logic
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(config.RetryDelay * time.Duration(attempt)):
			}

			cp.mu.Lock()
			cp.stats.RetryAttempts++
			providerStats.RetryAttempts++
			cp.mu.Unlock()
		}

		// Create request with context
		reqWithCtx := req.WithContext(ctx)

		// Perform request
		resp, lastErr = client.Do(reqWithCtx)

		// Check if request was successful
		if lastErr == nil && resp.StatusCode < 500 {
			// Success
			duration := time.Since(start)

			cp.mu.Lock()
			cp.stats.SuccessfulRequests++
			cp.stats.TotalLatency += duration
			cp.stats.AverageLatency = cp.stats.TotalLatency / time.Duration(cp.stats.TotalRequests)

			providerStats.SuccessfulRequests++
			providerStats.TotalLatency += duration
			providerStats.AverageLatency = providerStats.TotalLatency / time.Duration(providerStats.Requests)
			cp.mu.Unlock()

			cp.logger.Debug("Request successful",
				zap.String("provider", provider),
				zap.String("method", req.Method),
				zap.String("url", req.URL.String()),
				zap.Int("status_code", resp.StatusCode),
				zap.Duration("duration", duration),
				zap.Int("attempt", attempt+1))

			return resp, nil
		}

		// Log error
		if lastErr != nil {
			cp.logger.Warn("Request failed",
				zap.String("provider", provider),
				zap.String("method", req.Method),
				zap.String("url", req.URL.String()),
				zap.Error(lastErr),
				zap.Int("attempt", attempt+1))
		} else if resp != nil {
			cp.logger.Warn("Request failed with status code",
				zap.String("provider", provider),
				zap.String("method", req.Method),
				zap.String("url", req.URL.String()),
				zap.Int("status_code", resp.StatusCode),
				zap.Int("attempt", attempt+1))
		}

		// Close response body if present
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}

	// All retries failed
	duration := time.Since(start)

	cp.mu.Lock()
	cp.stats.FailedRequests++
	cp.stats.TotalLatency += duration
	cp.stats.AverageLatency = cp.stats.TotalLatency / time.Duration(cp.stats.TotalRequests)

	providerStats.FailedRequests++
	providerStats.TotalLatency += duration
	providerStats.AverageLatency = providerStats.TotalLatency / time.Duration(providerStats.Requests)

	// Track error type
	errorType := "unknown"
	if lastErr != nil {
		errorType = lastErr.Error()
	} else if resp != nil {
		errorType = fmt.Sprintf("status_%d", resp.StatusCode)
	}
	providerStats.ErrorCounts[errorType]++
	cp.mu.Unlock()

	cp.logger.Error("Request failed after all retries",
		zap.String("provider", provider),
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
		zap.Error(lastErr),
		zap.Int("max_retries", config.MaxRetries),
		zap.Duration("total_duration", duration))

	return nil, fmt.Errorf("request failed after %d retries: %w", config.MaxRetries, lastErr)
}

// GetStats returns connection pool statistics
func (cp *ConnectionPool) GetStats() *PoolStats {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	// Create a deep copy of stats
	stats := &PoolStats{
		TotalRequests:      cp.stats.TotalRequests,
		SuccessfulRequests: cp.stats.SuccessfulRequests,
		FailedRequests:     cp.stats.FailedRequests,
		RetryAttempts:      cp.stats.RetryAttempts,
		AverageLatency:     cp.stats.AverageLatency,
		TotalLatency:       cp.stats.TotalLatency,
		LastRequest:        cp.stats.LastRequest,
		ProviderStats:      make(map[string]*ProviderStats),
	}

	// Copy provider stats
	for provider, providerStats := range cp.stats.ProviderStats {
		stats.ProviderStats[provider] = &ProviderStats{
			Requests:           providerStats.Requests,
			SuccessfulRequests: providerStats.SuccessfulRequests,
			FailedRequests:     providerStats.FailedRequests,
			RetryAttempts:      providerStats.RetryAttempts,
			AverageLatency:     providerStats.AverageLatency,
			TotalLatency:       providerStats.TotalLatency,
			LastRequest:        providerStats.LastRequest,
			ErrorCounts:        make(map[string]int64),
		}

		// Copy error counts
		for errorType, count := range providerStats.ErrorCounts {
			stats.ProviderStats[provider].ErrorCounts[errorType] = count
		}
	}

	return stats
}

// GetProviderStats returns statistics for a specific provider
func (cp *ConnectionPool) GetProviderStats(provider string) (*ProviderStats, error) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	providerStats, exists := cp.stats.ProviderStats[provider]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", provider)
	}

	// Create a deep copy
	stats := &ProviderStats{
		Requests:           providerStats.Requests,
		SuccessfulRequests: providerStats.SuccessfulRequests,
		FailedRequests:     providerStats.FailedRequests,
		RetryAttempts:      providerStats.RetryAttempts,
		AverageLatency:     providerStats.AverageLatency,
		TotalLatency:       providerStats.TotalLatency,
		LastRequest:        providerStats.LastRequest,
		ErrorCounts:        make(map[string]int64),
	}

	// Copy error counts
	for errorType, count := range providerStats.ErrorCounts {
		stats.ErrorCounts[errorType] = count
	}

	return stats, nil
}

// Health checks the health of all connection pools
func (cp *ConnectionPool) Health() map[string]error {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	health := make(map[string]error)

	for provider, client := range cp.clients {
		// Simple health check - try to create a request
		req, err := http.NewRequest("GET", "http://example.com", nil)
		if err != nil {
			health[provider] = fmt.Errorf("failed to create test request: %w", err)
			continue
		}

		// Set a very short timeout for health check
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		req = req.WithContext(ctx)

		// This will likely fail, but we're just checking if the client is properly configured
		_, err = client.Do(req)
		if err != nil && err != context.DeadlineExceeded {
			health[provider] = fmt.Errorf("client health check failed: %w", err)
		} else {
			health[provider] = nil // Client is healthy
		}
	}

	return health
}

// Close closes all HTTP clients in the connection pool
func (cp *ConnectionPool) Close() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for provider, client := range cp.clients {
		// Close idle connections
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}

		cp.logger.Info("Closed connection pool for provider",
			zap.String("provider", provider))
	}

	cp.clients = make(map[string]*http.Client)
	cp.configs = make(map[string]*PoolConfig)

	return nil
}

// ListProviders returns a list of all configured providers
func (cp *ConnectionPool) ListProviders() []string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	providers := make([]string, 0, len(cp.clients))
	for provider := range cp.clients {
		providers = append(providers, provider)
	}

	return providers
}

// RemoveProvider removes a provider from the connection pool
func (cp *ConnectionPool) RemoveProvider(provider string) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	client, exists := cp.clients[provider]
	if !exists {
		return fmt.Errorf("provider %s not found", provider)
	}

	// Close idle connections
	if transport, ok := client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}

	// Remove from maps
	delete(cp.clients, provider)
	delete(cp.configs, provider)
	delete(cp.stats.ProviderStats, provider)

	cp.logger.Info("Removed provider from connection pool",
		zap.String("provider", provider))

	return nil
}
