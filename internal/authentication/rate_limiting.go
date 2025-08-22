package authentication

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RateLimitStrategy defines the strategy for handling rate limit exceeded scenarios
type RateLimitStrategy string

const (
	StrategyFailFast       RateLimitStrategy = "fail_fast"
	StrategyRetry          RateLimitStrategy = "retry"
	StrategyFallback       RateLimitStrategy = "fallback"
	StrategyExponential    RateLimitStrategy = "exponential_backoff"
	StrategyJitter         RateLimitStrategy = "jitter"
	StrategyCircuitBreaker RateLimitStrategy = "circuit_breaker"
)

// FallbackProvider represents an alternative data source when primary API is rate limited
type FallbackProvider struct {
	Name        string
	Priority    int
	SuccessRate float64
	LastUsed    time.Time
	IsAvailable bool
	Config      FallbackConfig
}

// FallbackConfig contains configuration for fallback providers
type FallbackConfig struct {
	MaxRetries       int
	RetryDelay       time.Duration
	Timeout          time.Duration
	SuccessThreshold float64
	FailureThreshold float64
}

// RetryStrategy defines retry behavior for rate limited requests
type RetryStrategy struct {
	MaxRetries        int
	BaseDelay         time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
	JitterFactor      float64
	RetryableErrors   []string
}

// CircuitBreakerConfig defines circuit breaker behavior
type CircuitBreakerConfig struct {
	FailureThreshold int
	RecoveryTimeout  time.Duration
	HalfOpenMaxCalls int
	SuccessThreshold float64
}

// EnhancedRateLimiter provides advanced rate limiting with fallback and retry strategies
type EnhancedRateLimiter struct {
	config            *EnhancedRateLimitConfig
	logger            *zap.Logger
	mu                sync.RWMutex
	providers         map[string]*ProviderRateLimit
	fallbackProviders map[string]*FallbackProvider
	circuitBreakers   map[string]*CircuitBreaker
	strategies        map[string]RateLimitStrategy
	cache             map[string]*CacheEntry
	cacheMu           sync.RWMutex
	optimizationStats map[string]*OptimizationStats
	statsMu           sync.RWMutex
}

// OptimizationStats contains optimization statistics
type OptimizationStats struct {
	CacheHits            int64
	CacheMisses          int64
	PredictiveHits       int64
	AdaptiveAdjustments  int64
	LoadBalancedRequests int64
	RateShapedRequests   int64
	LastUpdated          time.Time
}

// ProviderRateLimit contains rate limit information for a provider
type ProviderRateLimit struct {
	ProviderName      string
	RequestsPerMinute int
	RequestsPerHour   int
	CurrentRequests   int
	LastResetTime     time.Time
	QuotaExceeded     bool
	RetryAfter        time.Time
	FailureCount      int
	SuccessCount      int
	LastFailure       time.Time
	LastSuccess       time.Time
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	ProviderName string
	State        CircuitBreakerState
	FailureCount int
	SuccessCount int
	LastFailure  time.Time
	LastSuccess  time.Time
	NextAttempt  time.Time
	Config       CircuitBreakerConfig
	mu           sync.RWMutex
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "closed"
	StateOpen     CircuitBreakerState = "open"
	StateHalfOpen CircuitBreakerState = "half_open"
)

// CacheEntry represents a cached rate limit result
type CacheEntry struct {
	Result      *RateLimitResult
	ExpiresAt   time.Time
	AccessCount int
	LastAccess  time.Time
}

// OptimizationConfig contains optimization settings
type OptimizationConfig struct {
	EnableCaching            bool
	CacheTTL                 time.Duration
	CacheMaxSize             int
	EnablePredictiveLimiting bool
	PredictiveWindow         time.Duration
	EnableAdaptiveLimiting   bool
	AdaptiveThreshold        float64
	EnableLoadBalancing      bool
	LoadBalancingStrategy    string
	EnableRateShaping        bool
	RateShapingWindow        time.Duration
}

// EnhancedRateLimitConfig contains configuration for enhanced rate limiting
type EnhancedRateLimitConfig struct {
	DefaultStrategy      RateLimitStrategy
	GlobalRateLimit      int
	ProviderRateLimit    int
	RetryStrategy        RetryStrategy
	CircuitBreakerConfig CircuitBreakerConfig
	FallbackProviders    []FallbackProvider
	EnableMonitoring     bool
	EnableMetrics        bool
	Optimization         OptimizationConfig
}

// RateLimitResult contains the result of a rate limit check
type RateLimitResult struct {
	Allowed             bool
	ProviderName        string
	RemainingRequests   int
	ResetTime           time.Time
	RetryAfter          time.Time
	WaitTime            time.Duration
	Strategy            RateLimitStrategy
	FallbackAvailable   bool
	FallbackProvider    *FallbackProvider
	CircuitBreakerState CircuitBreakerState
	RetryCount          int
	Error               error
}

// NewEnhancedRateLimiter creates a new enhanced rate limiter
func NewEnhancedRateLimiter(config *EnhancedRateLimitConfig, logger *zap.Logger) *EnhancedRateLimiter {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config.DefaultStrategy == "" {
		config.DefaultStrategy = StrategyRetry
	}

	limiter := &EnhancedRateLimiter{
		config:            config,
		logger:            logger,
		providers:         make(map[string]*ProviderRateLimit),
		fallbackProviders: make(map[string]*FallbackProvider),
		circuitBreakers:   make(map[string]*CircuitBreaker),
		strategies:        make(map[string]RateLimitStrategy),
		cache:             make(map[string]*CacheEntry),
		optimizationStats: make(map[string]*OptimizationStats),
	}

	// Initialize fallback providers
	for _, provider := range config.FallbackProviders {
		limiter.fallbackProviders[provider.Name] = &provider
	}

	return limiter
}

// RegisterProvider registers a provider with rate limiting
func (erl *EnhancedRateLimiter) RegisterProvider(providerName string, rateLimit int, strategy RateLimitStrategy) {
	erl.mu.Lock()
	defer erl.mu.Unlock()

	erl.providers[providerName] = &ProviderRateLimit{
		ProviderName:      providerName,
		RequestsPerMinute: rateLimit,
		RequestsPerHour:   rateLimit * 60,
		LastResetTime:     time.Now(),
	}

	if strategy != "" {
		erl.strategies[providerName] = strategy
	} else {
		erl.strategies[providerName] = erl.config.DefaultStrategy
	}

	// Create circuit breaker for provider
	erl.circuitBreakers[providerName] = &CircuitBreaker{
		ProviderName: providerName,
		State:        StateClosed,
		Config:       erl.config.CircuitBreakerConfig,
	}
}

// CheckRateLimit checks if a request is allowed with fallback and retry strategies
func (erl *EnhancedRateLimiter) CheckRateLimit(ctx context.Context, providerName string) (*RateLimitResult, error) {
	erl.mu.Lock()
	defer erl.mu.Unlock()

	// Check circuit breaker first
	circuitBreaker := erl.circuitBreakers[providerName]
	if circuitBreaker != nil && !circuitBreaker.IsAllowed() {
		return &RateLimitResult{
			Allowed:             false,
			ProviderName:        providerName,
			Strategy:            StrategyCircuitBreaker,
			CircuitBreakerState: circuitBreaker.State,
			Error:               fmt.Errorf("circuit breaker is open for provider %s", providerName),
		}, nil
	}

	// Get or create provider rate limit
	provider := erl.getOrCreateProvider(providerName)
	strategy := erl.getStrategy(providerName)

	// Reset counters if needed
	erl.resetProviderCounters(provider)

	result := &RateLimitResult{
		ProviderName: providerName,
		Strategy:     strategy,
		ResetTime:    provider.LastResetTime.Add(time.Minute),
	}

	// Check if we're within limits
	if provider.CurrentRequests < provider.RequestsPerMinute {
		provider.CurrentRequests++
		provider.SuccessCount++
		provider.LastSuccess = time.Now()
		result.Allowed = true
		result.RemainingRequests = provider.RequestsPerMinute - provider.CurrentRequests

		// Update circuit breaker on success
		if circuitBreaker != nil {
			circuitBreaker.RecordSuccess()
		}
	} else {
		provider.QuotaExceeded = true
		provider.FailureCount++
		provider.LastFailure = time.Now()
		provider.RetryAfter = provider.LastResetTime.Add(time.Minute)

		result.Allowed = false
		result.RemainingRequests = 0
		result.RetryAfter = provider.RetryAfter
		result.WaitTime = result.RetryAfter.Sub(time.Now())

		// Update circuit breaker on failure
		if circuitBreaker != nil {
			circuitBreaker.RecordFailure()
		}

		// Apply strategy-specific handling
		erl.applyStrategy(result, provider, strategy)
	}

	// Check for fallback availability
	if result.FallbackAvailable = erl.hasFallbackProvider(providerName); result.FallbackAvailable {
		result.FallbackProvider = erl.getBestFallbackProvider(providerName)
	}

	// Log rate limit check
	erl.logger.Debug("Rate limit check",
		zap.String("provider", providerName),
		zap.Bool("allowed", result.Allowed),
		zap.String("strategy", string(strategy)),
		zap.Bool("fallback_available", result.FallbackAvailable),
	)

	return result, nil
}

// WaitForRateLimit waits until rate limit allows the request with retry strategies
func (erl *EnhancedRateLimiter) WaitForRateLimit(ctx context.Context, providerName string) error {
	maxRetries := erl.config.RetryStrategy.MaxRetries
	retryCount := 0

	for retryCount <= maxRetries {
		result, err := erl.CheckRateLimit(ctx, providerName)
		if err != nil {
			return err
		}

		if result.Allowed {
			return nil
		}

		// Check if we can use fallback
		if result.FallbackAvailable && result.FallbackProvider != nil {
			erl.logger.Info("Using fallback provider",
				zap.String("primary_provider", providerName),
				zap.String("fallback_provider", result.FallbackProvider.Name),
			)
			return nil
		}

		// Apply retry strategy
		delay := erl.calculateRetryDelay(retryCount, result.WaitTime)

		erl.logger.Info("Rate limit exceeded, retrying",
			zap.String("provider", providerName),
			zap.Int("retry_count", retryCount),
			zap.Duration("delay", delay),
			zap.String("strategy", string(result.Strategy)),
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			retryCount++
			continue
		}
	}

	return fmt.Errorf("rate limit exceeded after %d retries for provider %s", maxRetries, providerName)
}

// ExecuteWithFallback executes a function with fallback and retry strategies
func (erl *EnhancedRateLimiter) ExecuteWithFallback(ctx context.Context, providerName string,
	primaryFunc func() (interface{}, error),
	fallbackFuncs map[string]func() (interface{}, error)) (interface{}, error) {

	// Try primary provider first
	result, err := erl.CheckRateLimit(ctx, providerName)
	if err != nil {
		return nil, err
	}

	if result.Allowed {
		// Execute primary function
		if data, err := primaryFunc(); err == nil {
			return data, nil
		} else {
			erl.logger.Warn("Primary provider failed, trying fallback",
				zap.String("provider", providerName),
				zap.Error(err),
			)
		}
	}

	// Try fallback providers
	if result.FallbackAvailable && result.FallbackProvider != nil {
		if fallbackFunc, exists := fallbackFuncs[result.FallbackProvider.Name]; exists {
			erl.logger.Info("Executing fallback provider",
				zap.String("fallback_provider", result.FallbackProvider.Name),
			)

			if data, err := fallbackFunc(); err == nil {
				// Update fallback provider metrics
				result.FallbackProvider.LastUsed = time.Now()
				return data, nil
			} else {
				erl.logger.Warn("Fallback provider failed",
					zap.String("fallback_provider", result.FallbackProvider.Name),
					zap.Error(err),
				)
			}
		}
	}

	// If all providers failed, wait and retry with exponential backoff
	return erl.retryWithBackoff(ctx, providerName, primaryFunc, fallbackFuncs)
}

// retryWithBackoff implements exponential backoff retry strategy
func (erl *EnhancedRateLimiter) retryWithBackoff(ctx context.Context, providerName string,
	primaryFunc func() (interface{}, error),
	fallbackFuncs map[string]func() (interface{}, error)) (interface{}, error) {

	baseDelay := erl.config.RetryStrategy.BaseDelay
	maxDelay := erl.config.RetryStrategy.MaxDelay
	multiplier := erl.config.RetryStrategy.BackoffMultiplier

	for retryCount := 0; retryCount <= erl.config.RetryStrategy.MaxRetries; retryCount++ {
		// Calculate delay with exponential backoff and jitter
		delay := erl.calculateExponentialDelay(baseDelay, multiplier, retryCount, maxDelay)
		delay = erl.addJitter(delay, erl.config.RetryStrategy.JitterFactor)

		erl.logger.Info("Retrying with exponential backoff",
			zap.String("provider", providerName),
			zap.Int("retry_count", retryCount),
			zap.Duration("delay", delay),
		)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			// Try primary provider
			if result, err := erl.CheckRateLimit(ctx, providerName); err == nil && result.Allowed {
				if data, err := primaryFunc(); err == nil {
					return data, nil
				}
			}

			// Try fallback providers
			for fallbackName, fallbackFunc := range fallbackFuncs {
				if data, err := fallbackFunc(); err == nil {
					erl.logger.Info("Fallback succeeded after retry",
						zap.String("fallback_provider", fallbackName),
						zap.Int("retry_count", retryCount),
					)
					return data, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("all providers failed after %d retries", erl.config.RetryStrategy.MaxRetries)
}

// Helper methods

func (erl *EnhancedRateLimiter) getOrCreateProvider(providerName string) *ProviderRateLimit {
	if provider, exists := erl.providers[providerName]; exists {
		return provider
	}

	provider := &ProviderRateLimit{
		ProviderName:      providerName,
		RequestsPerMinute: erl.config.ProviderRateLimit,
		RequestsPerHour:   erl.config.ProviderRateLimit * 60,
		LastResetTime:     time.Now(),
	}
	erl.providers[providerName] = provider
	return provider
}

func (erl *EnhancedRateLimiter) getStrategy(providerName string) RateLimitStrategy {
	if strategy, exists := erl.strategies[providerName]; exists {
		return strategy
	}
	return erl.config.DefaultStrategy
}

func (erl *EnhancedRateLimiter) resetProviderCounters(provider *ProviderRateLimit) {
	now := time.Now()
	if now.Sub(provider.LastResetTime) >= time.Minute {
		provider.CurrentRequests = 0
		provider.LastResetTime = now
		provider.QuotaExceeded = false
	}
}

func (erl *EnhancedRateLimiter) applyStrategy(result *RateLimitResult, provider *ProviderRateLimit, strategy RateLimitStrategy) {
	switch strategy {
	case StrategyFailFast:
		// Do nothing, just return the result as is
	case StrategyRetry:
		// Set retry count for exponential backoff
		result.RetryCount = 1
	case StrategyExponential:
		// Calculate exponential backoff delay
		result.WaitTime = erl.calculateExponentialDelay(
			erl.config.RetryStrategy.BaseDelay,
			erl.config.RetryStrategy.BackoffMultiplier,
			provider.FailureCount,
			erl.config.RetryStrategy.MaxDelay,
		)
	case StrategyJitter:
		// Add jitter to the wait time
		result.WaitTime = erl.addJitter(result.WaitTime, erl.config.RetryStrategy.JitterFactor)
	}
}

func (erl *EnhancedRateLimiter) hasFallbackProvider(providerName string) bool {
	return len(erl.fallbackProviders) > 0
}

func (erl *EnhancedRateLimiter) getBestFallbackProvider(providerName string) *FallbackProvider {
	var bestProvider *FallbackProvider
	highestPriority := -1

	for _, provider := range erl.fallbackProviders {
		if provider.IsAvailable && provider.Priority > highestPriority {
			bestProvider = provider
			highestPriority = provider.Priority
		}
	}

	return bestProvider
}

func (erl *EnhancedRateLimiter) calculateRetryDelay(retryCount int, baseWaitTime time.Duration) time.Duration {
	if retryCount == 0 {
		return baseWaitTime
	}

	delay := erl.config.RetryStrategy.BaseDelay
	for i := 0; i < retryCount; i++ {
		delay = time.Duration(float64(delay) * erl.config.RetryStrategy.BackoffMultiplier)
		if delay > erl.config.RetryStrategy.MaxDelay {
			delay = erl.config.RetryStrategy.MaxDelay
			break
		}
	}

	return delay
}

func (erl *EnhancedRateLimiter) calculateExponentialDelay(baseDelay time.Duration, multiplier float64, retryCount int, maxDelay time.Duration) time.Duration {
	delay := baseDelay
	for i := 0; i < retryCount; i++ {
		delay = time.Duration(float64(delay) * multiplier)
		if delay > maxDelay {
			delay = maxDelay
			break
		}
	}
	return delay
}

func (erl *EnhancedRateLimiter) addJitter(delay time.Duration, jitterFactor float64) time.Duration {
	if jitterFactor <= 0 {
		return delay
	}

	jitter := time.Duration(float64(delay) * jitterFactor * (0.5 + 0.5*float64(time.Now().UnixNano()%1000)/1000.0))
	return delay + jitter
}

// Circuit Breaker methods

func (cb *CircuitBreaker) IsAllowed() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.State {
	case StateClosed:
		return true
	case StateOpen:
		if time.Now().After(cb.NextAttempt) {
			// Transition to half-open state
			cb.State = StateHalfOpen
			cb.SuccessCount = 0
			cb.FailureCount = 0
			return true
		}
		return false
	case StateHalfOpen:
		return cb.SuccessCount < cb.Config.HalfOpenMaxCalls
	default:
		return true
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.SuccessCount++
	cb.LastSuccess = time.Now()

	switch cb.State {
	case StateHalfOpen:
		if float64(cb.SuccessCount)/float64(cb.SuccessCount+cb.FailureCount) >= cb.Config.SuccessThreshold {
			cb.State = StateClosed
			cb.FailureCount = 0
			cb.SuccessCount = 0
		}
	case StateClosed:
		// Reset failure count on success
		cb.FailureCount = 0
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.FailureCount++
	cb.LastFailure = time.Now()

	switch cb.State {
	case StateClosed:
		if cb.FailureCount >= cb.Config.FailureThreshold {
			cb.State = StateOpen
			cb.NextAttempt = time.Now().Add(cb.Config.RecoveryTimeout)
		}
	case StateHalfOpen:
		cb.State = StateOpen
		cb.NextAttempt = time.Now().Add(cb.Config.RecoveryTimeout)
		cb.SuccessCount = 0
		cb.FailureCount = 0
	}
}

// GetRateLimitStats returns comprehensive rate limiting statistics
func (erl *EnhancedRateLimiter) GetRateLimitStats() map[string]interface{} {
	erl.mu.RLock()
	defer erl.mu.RUnlock()

	stats := map[string]interface{}{
		"total_providers":           len(erl.providers),
		"fallback_providers":        len(erl.fallbackProviders),
		"circuit_breakers":          len(erl.circuitBreakers),
		"default_strategy":          string(erl.config.DefaultStrategy),
		"providers":                 make(map[string]interface{}),
		"fallback_provider_details": make(map[string]interface{}),
		"circuit_breaker_details":   make(map[string]interface{}),
	}

	// Provider statistics
	for name, provider := range erl.providers {
		stats["providers"].(map[string]interface{})[name] = map[string]interface{}{
			"requests_per_minute": provider.RequestsPerMinute,
			"current_requests":    provider.CurrentRequests,
			"success_count":       provider.SuccessCount,
			"failure_count":       provider.FailureCount,
			"quota_exceeded":      provider.QuotaExceeded,
			"last_success":        provider.LastSuccess,
			"last_failure":        provider.LastFailure,
			"strategy":            string(erl.strategies[name]),
		}
	}

	// Fallback provider statistics
	for name, provider := range erl.fallbackProviders {
		stats["fallback_provider_details"].(map[string]interface{})[name] = map[string]interface{}{
			"priority":     provider.Priority,
			"success_rate": provider.SuccessRate,
			"is_available": provider.IsAvailable,
			"last_used":    provider.LastUsed,
		}
	}

	// Circuit breaker statistics
	for name, cb := range erl.circuitBreakers {
		cb.mu.RLock()
		stats["circuit_breaker_details"].(map[string]interface{})[name] = map[string]interface{}{
			"state":         string(cb.State),
			"failure_count": cb.FailureCount,
			"success_count": cb.SuccessCount,
			"last_failure":  cb.LastFailure,
			"last_success":  cb.LastSuccess,
			"next_attempt":  cb.NextAttempt,
		}
		cb.mu.RUnlock()
	}

	return stats
}

// Cache Management Methods

// getCacheKey generates a cache key for rate limit checks
func (erl *EnhancedRateLimiter) getCacheKey(providerName string) string {
	return fmt.Sprintf("rate_limit:%s", providerName)
}

// getCachedResult retrieves a cached rate limit result
func (erl *EnhancedRateLimiter) getCachedResult(providerName string) (*RateLimitResult, bool) {
	if !erl.config.Optimization.EnableCaching {
		return nil, false
	}

	erl.cacheMu.RLock()
	defer erl.cacheMu.RUnlock()

	cacheKey := erl.getCacheKey(providerName)
	entry, exists := erl.cache[cacheKey]
	if !exists {
		return nil, false
	}

	// Check if cache entry has expired
	if time.Now().After(entry.ExpiresAt) {
		erl.cacheMu.RUnlock()
		erl.cacheMu.Lock()
		delete(erl.cache, cacheKey)
		erl.cacheMu.Unlock()
		erl.cacheMu.RLock()
		return nil, false
	}

	// Update access statistics
	entry.AccessCount++
	entry.LastAccess = time.Now()

	// Record cache hit
	erl.recordCacheHit(providerName)

	return entry.Result, true
}

// setCachedResult stores a rate limit result in cache
func (erl *EnhancedRateLimiter) setCachedResult(providerName string, result *RateLimitResult) {
	if !erl.config.Optimization.EnableCaching {
		return
	}

	erl.cacheMu.Lock()
	defer erl.cacheMu.Unlock()

	// Check cache size limit
	if len(erl.cache) >= erl.config.Optimization.CacheMaxSize {
		erl.evictLeastUsed()
	}

	cacheKey := erl.getCacheKey(providerName)
	entry := &CacheEntry{
		Result:      result,
		ExpiresAt:   time.Now().Add(erl.config.Optimization.CacheTTL),
		AccessCount: 1,
		LastAccess:  time.Now(),
	}

	erl.cache[cacheKey] = entry
}

// evictLeastUsed removes the least used cache entry
func (erl *EnhancedRateLimiter) evictLeastUsed() {
	var leastUsedKey string
	var leastUsedCount int
	var oldestAccess time.Time

	for key, entry := range erl.cache {
		if leastUsedKey == "" || entry.AccessCount < leastUsedCount ||
			(entry.AccessCount == leastUsedCount && entry.LastAccess.Before(oldestAccess)) {
			leastUsedKey = key
			leastUsedCount = entry.AccessCount
			oldestAccess = entry.LastAccess
		}
	}

	if leastUsedKey != "" {
		delete(erl.cache, leastUsedKey)
	}
}

// recordCacheHit records a cache hit for statistics
func (erl *EnhancedRateLimiter) recordCacheHit(providerName string) {
	erl.statsMu.Lock()
	defer erl.statsMu.Unlock()

	stats := erl.getOrCreateOptimizationStats(providerName)
	stats.CacheHits++
	stats.LastUpdated = time.Now()
}

// recordCacheMiss records a cache miss for statistics
func (erl *EnhancedRateLimiter) recordCacheMiss(providerName string) {
	erl.statsMu.Lock()
	defer erl.statsMu.Unlock()

	stats := erl.getOrCreateOptimizationStats(providerName)
	stats.CacheMisses++
	stats.LastUpdated = time.Now()
}

// getOrCreateOptimizationStats gets or creates optimization statistics
func (erl *EnhancedRateLimiter) getOrCreateOptimizationStats(providerName string) *OptimizationStats {
	stats, exists := erl.optimizationStats[providerName]
	if !exists {
		stats = &OptimizationStats{
			LastUpdated: time.Now(),
		}
		erl.optimizationStats[providerName] = stats
	}
	return stats
}

// Optimization Methods

// CheckRateLimitOptimized checks rate limits with optimization features
func (erl *EnhancedRateLimiter) CheckRateLimitOptimized(ctx context.Context, providerName string) (*RateLimitResult, error) {
	// Try cache first
	if cached, found := erl.getCachedResult(providerName); found {
		erl.logger.Debug("Rate limit cache hit", zap.String("provider", providerName))
		return cached, nil
	}

	// Record cache miss
	erl.recordCacheMiss(providerName)

	// Perform predictive limiting if enabled
	if erl.config.Optimization.EnablePredictiveLimiting {
		if result := erl.predictiveLimitCheck(providerName); result != nil {
			erl.setCachedResult(providerName, result)
			return result, nil
		}
	}

	// Perform adaptive limiting if enabled
	if erl.config.Optimization.EnableAdaptiveLimiting {
		erl.adaptiveLimitAdjustment(providerName)
	}

	// Perform load balancing if enabled
	if erl.config.Optimization.EnableLoadBalancing {
		providerName = erl.loadBalanceProvider(providerName)
	}

	// Perform rate shaping if enabled
	if erl.config.Optimization.EnableRateShaping {
		erl.rateShapeRequest(providerName)
	}

	// Check rate limit normally
	result, err := erl.CheckRateLimit(ctx, providerName)
	if err != nil {
		return nil, err
	}

	// Cache the result
	erl.setCachedResult(providerName, result)

	return result, nil
}

// predictiveLimitCheck performs predictive rate limiting
func (erl *EnhancedRateLimiter) predictiveLimitCheck(providerName string) *RateLimitResult {
	erl.mu.RLock()
	provider, exists := erl.providers[providerName]
	erl.mu.RUnlock()

	if !exists {
		return nil
	}

	// Analyze recent request patterns
	now := time.Now()
	windowStart := now.Add(-erl.config.Optimization.PredictiveWindow)

	// If we're approaching the rate limit based on recent patterns, predict rejection
	if erl.isApproachingLimit(provider, windowStart) {
		erl.recordPredictiveHit(providerName)
		return &RateLimitResult{
			Allowed:           false,
			ProviderName:      providerName,
			Strategy:          StrategyFailFast,
			RemainingRequests: 0,
			RetryAfter:        provider.LastResetTime.Add(time.Minute),
			WaitTime:          provider.LastResetTime.Add(time.Minute).Sub(now),
		}
	}

	return nil
}

// isApproachingLimit checks if we're approaching the rate limit
func (erl *EnhancedRateLimiter) isApproachingLimit(provider *ProviderRateLimit, windowStart time.Time) bool {
	// Simple heuristic: if we've used more than 80% of our limit in the last window
	usageRatio := float64(provider.CurrentRequests) / float64(provider.RequestsPerMinute)
	return usageRatio >= 0.8
}

// adaptiveLimitAdjustment adjusts rate limits based on performance
func (erl *EnhancedRateLimiter) adaptiveLimitAdjustment(providerName string) {
	erl.mu.Lock()
	defer erl.mu.Unlock()

	provider, exists := erl.providers[providerName]
	if !exists {
		return
	}

	// Calculate success rate
	totalRequests := provider.SuccessCount + provider.FailureCount
	if totalRequests == 0 {
		return
	}

	successRate := float64(provider.SuccessCount) / float64(totalRequests)

	// Adjust rate limit based on success rate
	if successRate > erl.config.Optimization.AdaptiveThreshold {
		// Increase rate limit for high success rates
		provider.RequestsPerMinute = int(float64(provider.RequestsPerMinute) * 1.1)
	} else if successRate < erl.config.Optimization.AdaptiveThreshold*0.5 {
		// Decrease rate limit for low success rates
		provider.RequestsPerMinute = int(float64(provider.RequestsPerMinute) * 0.9)
	}

	erl.recordAdaptiveAdjustment(providerName)
}

// loadBalanceProvider selects the best provider for load balancing
func (erl *EnhancedRateLimiter) loadBalanceProvider(originalProvider string) string {
	if erl.config.Optimization.LoadBalancingStrategy == "round_robin" {
		return erl.roundRobinProvider(originalProvider)
	} else if erl.config.Optimization.LoadBalancingStrategy == "least_loaded" {
		return erl.leastLoadedProvider(originalProvider)
	}
	return originalProvider
}

// roundRobinProvider implements round-robin load balancing
func (erl *EnhancedRateLimiter) roundRobinProvider(originalProvider string) string {
	erl.mu.RLock()
	defer erl.mu.RUnlock()

	// Simple round-robin: find next available provider
	providers := make([]string, 0, len(erl.providers))
	for name := range erl.providers {
		providers = append(providers, name)
	}

	if len(providers) <= 1 {
		return originalProvider
	}

	// Find current provider index
	currentIndex := -1
	for i, name := range providers {
		if name == originalProvider {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return originalProvider
	}

	// Return next provider in round-robin
	nextIndex := (currentIndex + 1) % len(providers)
	erl.recordLoadBalancedRequest(originalProvider)
	return providers[nextIndex]
}

// leastLoadedProvider implements least-loaded load balancing
func (erl *EnhancedRateLimiter) leastLoadedProvider(originalProvider string) string {
	erl.mu.RLock()
	defer erl.mu.RUnlock()

	var leastLoaded string
	var minUsage float64 = 1.0

	for name, provider := range erl.providers {
		usageRatio := float64(provider.CurrentRequests) / float64(provider.RequestsPerMinute)
		if usageRatio < minUsage {
			minUsage = usageRatio
			leastLoaded = name
		}
	}

	if leastLoaded == "" {
		return originalProvider
	}

	erl.recordLoadBalancedRequest(originalProvider)
	return leastLoaded
}

// rateShapeRequest applies rate shaping to requests
func (erl *EnhancedRateLimiter) rateShapeRequest(providerName string) {
	// Simple rate shaping: add small delay to smooth out request bursts
	shapingDelay := time.Duration(float64(erl.config.Optimization.RateShapingWindow) / float64(erl.config.ProviderRateLimit))
	if shapingDelay > 0 {
		time.Sleep(shapingDelay)
	}
	erl.recordRateShapedRequest(providerName)
}

// Statistics Recording Methods

func (erl *EnhancedRateLimiter) recordPredictiveHit(providerName string) {
	erl.statsMu.Lock()
	defer erl.statsMu.Unlock()

	stats := erl.getOrCreateOptimizationStats(providerName)
	stats.PredictiveHits++
	stats.LastUpdated = time.Now()
}

func (erl *EnhancedRateLimiter) recordAdaptiveAdjustment(providerName string) {
	erl.statsMu.Lock()
	defer erl.statsMu.Unlock()

	stats := erl.getOrCreateOptimizationStats(providerName)
	stats.AdaptiveAdjustments++
	stats.LastUpdated = time.Now()
}

func (erl *EnhancedRateLimiter) recordLoadBalancedRequest(providerName string) {
	erl.statsMu.Lock()
	defer erl.statsMu.Unlock()

	stats := erl.getOrCreateOptimizationStats(providerName)
	stats.LoadBalancedRequests++
	stats.LastUpdated = time.Now()
}

func (erl *EnhancedRateLimiter) recordRateShapedRequest(providerName string) {
	erl.statsMu.Lock()
	defer erl.statsMu.Unlock()

	stats := erl.getOrCreateOptimizationStats(providerName)
	stats.RateShapedRequests++
	stats.LastUpdated = time.Now()
}

// GetOptimizationStats returns optimization statistics
func (erl *EnhancedRateLimiter) GetOptimizationStats() map[string]interface{} {
	erl.statsMu.RLock()
	defer erl.statsMu.RUnlock()

	stats := make(map[string]interface{})
	for providerName, providerStats := range erl.optimizationStats {
		stats[providerName] = map[string]interface{}{
			"cache_hits":             providerStats.CacheHits,
			"cache_misses":           providerStats.CacheMisses,
			"cache_hit_rate":         float64(providerStats.CacheHits) / float64(providerStats.CacheHits+providerStats.CacheMisses),
			"predictive_hits":        providerStats.PredictiveHits,
			"adaptive_adjustments":   providerStats.AdaptiveAdjustments,
			"load_balanced_requests": providerStats.LoadBalancedRequests,
			"rate_shaped_requests":   providerStats.RateShapedRequests,
			"last_updated":           providerStats.LastUpdated,
		}
	}

	return stats
}

// ClearCache clears the rate limit cache
func (erl *EnhancedRateLimiter) ClearCache() {
	erl.cacheMu.Lock()
	defer erl.cacheMu.Unlock()

	erl.cache = make(map[string]*CacheEntry)
	erl.logger.Info("Rate limit cache cleared")
}

// GetCacheStats returns cache statistics
func (erl *EnhancedRateLimiter) GetCacheStats() map[string]interface{} {
	erl.cacheMu.RLock()
	defer erl.cacheMu.RUnlock()

	totalEntries := len(erl.cache)
	var totalAccessCount int
	var oldestEntry time.Time
	var newestEntry time.Time

	for _, entry := range erl.cache {
		totalAccessCount += entry.AccessCount
		if oldestEntry.IsZero() || entry.LastAccess.Before(oldestEntry) {
			oldestEntry = entry.LastAccess
		}
		if newestEntry.IsZero() || entry.LastAccess.After(newestEntry) {
			newestEntry = entry.LastAccess
		}
	}

	return map[string]interface{}{
		"total_entries":      totalEntries,
		"max_size":           erl.config.Optimization.CacheMaxSize,
		"total_access_count": totalAccessCount,
		"oldest_entry":       oldestEntry,
		"newest_entry":       newestEntry,
		"cache_ttl":          erl.config.Optimization.CacheTTL,
	}
}
