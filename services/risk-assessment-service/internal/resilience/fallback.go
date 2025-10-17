package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FallbackStrategy implements fallback strategies for service failures
type FallbackStrategy struct {
	logger    *zap.Logger
	mu        sync.RWMutex
	stats     *FallbackStats
	config    *FallbackConfig
	fallbacks map[string]*FallbackHandler
}

// FallbackStats represents statistics for fallback operations
type FallbackStats struct {
	TotalRequests       int64         `json:"total_requests"`
	FallbackRequests    int64         `json:"fallback_requests"`
	SuccessfulFallbacks int64         `json:"successful_fallbacks"`
	FailedFallbacks     int64         `json:"failed_fallbacks"`
	AverageFallbackTime time.Duration `json:"average_fallback_time"`
	LastFallback        time.Time     `json:"last_fallback"`
}

// FallbackConfig represents configuration for fallback strategies
type FallbackConfig struct {
	DefaultTimeout time.Duration `json:"default_timeout"`
	EnableCaching  bool          `json:"enable_caching"`
	CacheTTL       time.Duration `json:"cache_ttl"`
	EnableMetrics  bool          `json:"enable_metrics"`
	EnableLogging  bool          `json:"enable_logging"`
	MaxRetries     int           `json:"max_retries"`
	RetryDelay     time.Duration `json:"retry_delay"`
}

// FallbackHandler represents a fallback handler for a specific service
type FallbackHandler struct {
	Name                string                 `json:"name"`
	Service             string                 `json:"service"`
	FallbackType        string                 `json:"fallback_type"`
	Config              map[string]interface{} `json:"config"`
	Enabled             bool                   `json:"enabled"`
	Timeout             time.Duration          `json:"timeout"`
	CacheEnabled        bool                   `json:"cache_enabled"`
	CacheTTL            time.Duration          `json:"cache_ttl"`
	TotalRequests       int64                  `json:"total_requests"`
	FallbackRequests    int64                  `json:"fallback_requests"`
	SuccessfulFallbacks int64                  `json:"successful_fallbacks"`
	FailedFallbacks     int64                  `json:"failed_fallbacks"`
	LastFallback        time.Time              `json:"last_fallback"`
	mu                  sync.RWMutex
}

// FallbackRequest represents a request that may need fallback
type FallbackRequest struct {
	ID        string                 `json:"id"`
	Service   string                 `json:"service"`
	Operation string                 `json:"operation"`
	Data      map[string]interface{} `json:"data"`
	Timeout   time.Duration          `json:"timeout"`
	Priority  int                    `json:"priority"`
	CreatedAt time.Time              `json:"created_at"`
}

// FallbackResponse represents a response from fallback
type FallbackResponse struct {
	ID           string                 `json:"id"`
	Success      bool                   `json:"success"`
	Result       map[string]interface{} `json:"result"`
	Error        string                 `json:"error,omitempty"`
	FallbackUsed bool                   `json:"fallback_used"`
	FallbackType string                 `json:"fallback_type,omitempty"`
	ProcessTime  time.Duration          `json:"process_time"`
	CreatedAt    time.Time              `json:"created_at"`
}

// NewFallbackStrategy creates a new fallback strategy
func NewFallbackStrategy(config *FallbackConfig, logger *zap.Logger) *FallbackStrategy {
	if config == nil {
		config = &FallbackConfig{
			DefaultTimeout: 30 * time.Second,
			EnableCaching:  true,
			CacheTTL:       5 * time.Minute,
			EnableMetrics:  true,
			EnableLogging:  true,
			MaxRetries:     3,
			RetryDelay:     1 * time.Second,
		}
	}

	return &FallbackStrategy{
		logger:    logger,
		stats:     &FallbackStats{},
		config:    config,
		fallbacks: make(map[string]*FallbackHandler),
	}
}

// RegisterFallback registers a fallback handler for a service
func (fs *FallbackStrategy) RegisterFallback(handler *FallbackHandler) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, exists := fs.fallbacks[handler.Service]; exists {
		return fmt.Errorf("fallback for service %s already exists", handler.Service)
	}

	fs.fallbacks[handler.Service] = handler

	fs.logger.Info("Fallback handler registered",
		zap.String("service", handler.Service),
		zap.String("fallback_type", handler.FallbackType),
		zap.Bool("enabled", handler.Enabled))

	return nil
}

// ExecuteWithFallback executes a request with fallback support
func (fs *FallbackStrategy) ExecuteWithFallback(ctx context.Context, request *FallbackRequest, primaryProcessor func(context.Context, *FallbackRequest) (*FallbackResponse, error)) (*FallbackResponse, error) {
	start := time.Now()

	fs.mu.Lock()
	fs.stats.TotalRequests++
	fs.mu.Unlock()

	// Try primary processor first
	response, err := primaryProcessor(ctx, request)
	if err == nil {
		// Primary processor succeeded
		response.ProcessTime = time.Since(start)
		return response, nil
	}

	fs.logger.Warn("Primary processor failed, attempting fallback",
		zap.String("request_id", request.ID),
		zap.String("service", request.Service),
		zap.Error(err))

	// Get fallback handler
	handler, exists := fs.getFallbackHandler(request.Service)
	if !exists {
		return nil, fmt.Errorf("no fallback handler available for service %s", request.Service)
	}

	if !handler.Enabled {
		return nil, fmt.Errorf("fallback handler for service %s is disabled", request.Service)
	}

	// Execute fallback
	fallbackResponse, fallbackErr := fs.executeFallback(ctx, request, handler)
	if fallbackErr != nil {
		fs.mu.Lock()
		fs.stats.FailedFallbacks++
		fs.mu.Unlock()

		handler.mu.Lock()
		handler.FailedFallbacks++
		handler.mu.Unlock()

		return nil, fmt.Errorf("fallback execution failed: %w", fallbackErr)
	}

	// Update statistics
	fs.mu.Lock()
	fs.stats.FallbackRequests++
	fs.stats.SuccessfulFallbacks++
	fs.stats.AverageFallbackTime = (fs.stats.AverageFallbackTime + time.Since(start)) / 2
	fs.stats.LastFallback = time.Now()
	fs.mu.Unlock()

	handler.mu.Lock()
	handler.FallbackRequests++
	handler.SuccessfulFallbacks++
	handler.LastFallback = time.Now()
	handler.mu.Unlock()

	// Set fallback metadata
	fallbackResponse.FallbackUsed = true
	fallbackResponse.FallbackType = handler.FallbackType
	fallbackResponse.ProcessTime = time.Since(start)

	fs.logger.Info("Fallback executed successfully",
		zap.String("request_id", request.ID),
		zap.String("service", request.Service),
		zap.String("fallback_type", handler.FallbackType),
		zap.Duration("process_time", fallbackResponse.ProcessTime))

	return fallbackResponse, nil
}

// GetStats returns fallback statistics
func (fs *FallbackStrategy) GetStats() *FallbackStats {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	stats := *fs.stats
	return &stats
}

// GetFallbackStats returns statistics for a specific fallback handler
func (fs *FallbackStrategy) GetFallbackStats(serviceName string) (*FallbackHandler, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	handler, exists := fs.fallbacks[serviceName]
	if !exists {
		return nil, fmt.Errorf("fallback handler for service %s not found", serviceName)
	}

	handler.mu.RLock()
	defer handler.mu.RUnlock()

	// Return a copy to avoid race conditions (excluding mutex)
	handlerCopy := FallbackHandler{
		Name:                handler.Name,
		Service:             handler.Service,
		FallbackType:        handler.FallbackType,
		Config:              handler.Config,
		Enabled:             handler.Enabled,
		Timeout:             handler.Timeout,
		CacheEnabled:        handler.CacheEnabled,
		CacheTTL:            handler.CacheTTL,
		TotalRequests:       handler.TotalRequests,
		FallbackRequests:    handler.FallbackRequests,
		SuccessfulFallbacks: handler.SuccessfulFallbacks,
		FailedFallbacks:     handler.FailedFallbacks,
		LastFallback:        handler.LastFallback,
	}
	return &handlerCopy, nil
}

// GetAllFallbackStats returns statistics for all fallback handlers
func (fs *FallbackStrategy) GetAllFallbackStats() map[string]*FallbackHandler {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	stats := make(map[string]*FallbackHandler)
	for name, handler := range fs.fallbacks {
		handler.mu.RLock()
		handlerCopy := FallbackHandler{
			Name:                handler.Name,
			Service:             handler.Service,
			FallbackType:        handler.FallbackType,
			Config:              handler.Config,
			Enabled:             handler.Enabled,
			Timeout:             handler.Timeout,
			CacheEnabled:        handler.CacheEnabled,
			CacheTTL:            handler.CacheTTL,
			TotalRequests:       handler.TotalRequests,
			FallbackRequests:    handler.FallbackRequests,
			SuccessfulFallbacks: handler.SuccessfulFallbacks,
			FailedFallbacks:     handler.FailedFallbacks,
			LastFallback:        handler.LastFallback,
		}
		handler.mu.RUnlock()
		stats[name] = &handlerCopy
	}

	return stats
}

// EnableFallback enables a fallback handler
func (fs *FallbackStrategy) EnableFallback(serviceName string) error {
	fs.mu.RLock()
	handler, exists := fs.fallbacks[serviceName]
	fs.mu.RUnlock()

	if !exists {
		return fmt.Errorf("fallback handler for service %s not found", serviceName)
	}

	handler.mu.Lock()
	handler.Enabled = true
	handler.mu.Unlock()

	fs.logger.Info("Fallback handler enabled",
		zap.String("service", serviceName))

	return nil
}

// DisableFallback disables a fallback handler
func (fs *FallbackStrategy) DisableFallback(serviceName string) error {
	fs.mu.RLock()
	handler, exists := fs.fallbacks[serviceName]
	fs.mu.RUnlock()

	if !exists {
		return fmt.Errorf("fallback handler for service %s not found", serviceName)
	}

	handler.mu.Lock()
	handler.Enabled = false
	handler.mu.Unlock()

	fs.logger.Info("Fallback handler disabled",
		zap.String("service", serviceName))

	return nil
}

// ResetStats resets all fallback statistics
func (fs *FallbackStrategy) ResetStats() {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.stats = &FallbackStats{}

	for _, handler := range fs.fallbacks {
		handler.mu.Lock()
		handler.TotalRequests = 0
		handler.FallbackRequests = 0
		handler.SuccessfulFallbacks = 0
		handler.FailedFallbacks = 0
		handler.mu.Unlock()
	}

	fs.logger.Info("Fallback statistics reset")
}

// Helper methods

func (fs *FallbackStrategy) getFallbackHandler(serviceName string) (*FallbackHandler, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	handler, exists := fs.fallbacks[serviceName]
	return handler, exists
}

func (fs *FallbackStrategy) executeFallback(ctx context.Context, request *FallbackRequest, handler *FallbackHandler) (*FallbackResponse, error) {
	handler.mu.Lock()
	handler.TotalRequests++
	handler.mu.Unlock()

	// Create timeout context
	timeout := handler.Timeout
	if timeout == 0 {
		timeout = fs.config.DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute based on fallback type
	switch handler.FallbackType {
	case "cached_response":
		return fs.executeCachedResponseFallback(ctx, request, handler)
	case "alternative_service":
		return fs.executeAlternativeServiceFallback(ctx, request, handler)
	case "degraded_mode":
		return fs.executeDegradedModeFallback(ctx, request, handler)
	case "default_response":
		return fs.executeDefaultResponseFallback(ctx, request, handler)
	default:
		return nil, fmt.Errorf("unknown fallback type: %s", handler.FallbackType)
	}
}

func (fs *FallbackStrategy) executeCachedResponseFallback(ctx context.Context, request *FallbackRequest, handler *FallbackHandler) (*FallbackResponse, error) {
	fs.logger.Debug("Executing cached response fallback",
		zap.String("request_id", request.ID),
		zap.String("service", request.Service))

	// Simulate cached response retrieval
	// In a real implementation, you would retrieve from cache
	time.Sleep(10 * time.Millisecond)

	result := map[string]interface{}{
		"risk_score": 0.75,
		"risk_level": "medium",
		"factors":    []string{"industry_risk", "country_risk"},
		"cached":     true,
		"fallback":   true,
	}

	return &FallbackResponse{
		ID:        request.ID,
		Success:   true,
		Result:    result,
		CreatedAt: time.Now(),
	}, nil
}

func (fs *FallbackStrategy) executeAlternativeServiceFallback(ctx context.Context, request *FallbackRequest, handler *FallbackHandler) (*FallbackResponse, error) {
	fs.logger.Debug("Executing alternative service fallback",
		zap.String("request_id", request.ID),
		zap.String("service", request.Service))

	// Simulate alternative service call
	// In a real implementation, you would call an alternative service
	time.Sleep(50 * time.Millisecond)

	result := map[string]interface{}{
		"risk_score": 0.70,
		"risk_level": "medium",
		"factors":    []string{"alternative_risk", "backup_assessment"},
		"fallback":   true,
		"source":     "alternative_service",
	}

	return &FallbackResponse{
		ID:        request.ID,
		Success:   true,
		Result:    result,
		CreatedAt: time.Now(),
	}, nil
}

func (fs *FallbackStrategy) executeDegradedModeFallback(ctx context.Context, request *FallbackRequest, handler *FallbackHandler) (*FallbackResponse, error) {
	fs.logger.Debug("Executing degraded mode fallback",
		zap.String("request_id", request.ID),
		zap.String("service", request.Service))

	// Simulate degraded mode processing
	// In a real implementation, you would use simplified logic
	time.Sleep(20 * time.Millisecond)

	result := map[string]interface{}{
		"risk_score": 0.80,
		"risk_level": "high",
		"factors":    []string{"degraded_assessment"},
		"fallback":   true,
		"degraded":   true,
		"warning":    "Service operating in degraded mode",
	}

	return &FallbackResponse{
		ID:        request.ID,
		Success:   true,
		Result:    result,
		CreatedAt: time.Now(),
	}, nil
}

func (fs *FallbackStrategy) executeDefaultResponseFallback(ctx context.Context, request *FallbackRequest, handler *FallbackHandler) (*FallbackResponse, error) {
	fs.logger.Debug("Executing default response fallback",
		zap.String("request_id", request.ID),
		zap.String("service", request.Service))

	// Simulate default response
	time.Sleep(5 * time.Millisecond)

	result := map[string]interface{}{
		"risk_score": 0.50,
		"risk_level": "unknown",
		"factors":    []string{"default_assessment"},
		"fallback":   true,
		"default":    true,
		"warning":    "Using default risk assessment",
	}

	return &FallbackResponse{
		ID:        request.ID,
		Success:   true,
		Result:    result,
		CreatedAt: time.Now(),
	}, nil
}

// FallbackManager manages multiple fallback strategies
type FallbackManager struct {
	strategies map[string]*FallbackStrategy
	logger     *zap.Logger
	mu         sync.RWMutex
}

// NewFallbackManager creates a new fallback manager
func NewFallbackManager(logger *zap.Logger) *FallbackManager {
	return &FallbackManager{
		strategies: make(map[string]*FallbackStrategy),
		logger:     logger,
	}
}

// GetStrategy gets or creates a fallback strategy
func (fm *FallbackManager) GetStrategy(strategyName string, config *FallbackConfig) *FallbackStrategy {
	fm.mu.RLock()
	strategy, exists := fm.strategies[strategyName]
	fm.mu.RUnlock()

	if exists {
		return strategy
	}

	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Check again in case another goroutine created it
	if strategy, exists := fm.strategies[strategyName]; exists {
		return strategy
	}

	strategy = NewFallbackStrategy(config, fm.logger)
	fm.strategies[strategyName] = strategy

	fm.logger.Info("Fallback strategy created",
		zap.String("strategy_name", strategyName))

	return strategy
}

// GetAllStats returns statistics for all fallback strategies
func (fm *FallbackManager) GetAllStats() map[string]*FallbackStats {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	stats := make(map[string]*FallbackStats)
	for name, strategy := range fm.strategies {
		stats[name] = strategy.GetStats()
	}

	return stats
}

// ResetAllStats resets statistics for all fallback strategies
func (fm *FallbackManager) ResetAllStats() {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	for _, strategy := range fm.strategies {
		strategy.ResetStats()
	}

	fm.logger.Info("All fallback statistics reset")
}
