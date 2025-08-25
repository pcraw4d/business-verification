package website_verification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// FallbackStrategy implements verification fallback strategies
type FallbackStrategy struct {
	// Configuration
	config *FallbackStrategyConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Verification methods
	verifiers    map[VerificationMethodType]Verifier
	verifiersMux sync.RWMutex

	// Fallback chain management
	fallbackChain *FallbackChain
	chainMux      sync.RWMutex

	// Retry management
	retryManager *RetryManager
	retryMux     sync.RWMutex

	// Cache management
	cacheManager *CacheManager
	cacheMux     sync.RWMutex

	// Timeout management
	timeoutManager *TimeoutManager
	timeoutMux     sync.RWMutex
}

// FallbackStrategyConfig configuration for fallback strategies
type FallbackStrategyConfig struct {
	// Fallback chain settings
	FallbackChainEnabled bool
	MaxFallbackDepth     int
	FallbackTimeout      time.Duration

	// Retry settings
	RetryEnabled       bool
	MaxRetryAttempts   int
	BaseRetryDelay     time.Duration
	MaxRetryDelay      time.Duration
	RetryBackoffFactor float64

	// Cache settings
	CacheEnabled         bool
	CacheTTL             time.Duration
	CacheMaxSize         int
	CacheCleanupInterval time.Duration

	// Timeout settings
	TimeoutEnabled bool
	DefaultTimeout time.Duration
	MethodTimeouts map[VerificationMethodType]time.Duration
	OverallTimeout time.Duration
}

// VerificationMethodType represents the type of verification method
type VerificationMethodType string

const (
	VerificationMethodDNS     VerificationMethodType = "dns"
	VerificationMethodWHOIS   VerificationMethodType = "whois"
	VerificationMethodContent VerificationMethodType = "content"
	VerificationMethodName    VerificationMethodType = "name"
	VerificationMethodAddress VerificationMethodType = "address"
	VerificationMethodPhone   VerificationMethodType = "phone"
	VerificationMethodEmail   VerificationMethodType = "email"
)

// Verifier interface for verification methods
type Verifier interface {
	Verify(ctx context.Context, domain, businessName, address, phone, email string) (*VerificationResult, error)
	GetMethodType() VerificationMethodType
	GetPriority() int
	IsEnabled() bool
}

// FallbackChain manages the fallback chain for verification methods
type FallbackChain struct {
	chains     map[string][]VerificationMethodType
	priorities map[VerificationMethodType]int
	mux        sync.RWMutex
}

// RetryManager manages retry logic with exponential backoff
type RetryManager struct {
	enabled         bool
	maxAttempts     int
	baseDelay       time.Duration
	maxDelay        time.Duration
	backoffFactor   float64
	retryHistory    map[string]*RetryHistory
	retryHistoryMux sync.RWMutex
}

// RetryHistory tracks retry attempts for a specific verification
type RetryHistory struct {
	Attempts     int
	LastAttempt  time.Time
	LastError    error
	TotalDelay   time.Duration
	SuccessCount int
	FailureCount int
}

// CacheManager manages verification result caching
type CacheManager struct {
	enabled         bool
	ttl             time.Duration
	maxSize         int
	cleanupInterval time.Duration
	cache           map[string]*CachedResult
	cacheMux        sync.RWMutex
	cleanupTicker   *time.Ticker
	stopCleanup     chan bool
}

// CachedResult represents a cached verification result
type CachedResult struct {
	Result      *VerificationResult
	Timestamp   time.Time
	TTL         time.Duration
	AccessCount int
}

// TimeoutManager manages timeouts for verification methods
type TimeoutManager struct {
	enabled           bool
	defaultTimeout    time.Duration
	methodTimeouts    map[VerificationMethodType]time.Duration
	overallTimeout    time.Duration
	timeoutHistory    map[string]*TimeoutHistory
	timeoutHistoryMux sync.RWMutex
}

// TimeoutHistory tracks timeout events for monitoring
type TimeoutHistory struct {
	Timeouts     int
	LastTimeout  time.Time
	TotalTimeout time.Duration
	Method       VerificationMethodType
}

// NewFallbackStrategy creates a new fallback strategy
func NewFallbackStrategy(config *FallbackStrategyConfig, logger *observability.Logger, tracer trace.Tracer) *FallbackStrategy {
	if config == nil {
		config = &FallbackStrategyConfig{
			FallbackChainEnabled: true,
			MaxFallbackDepth:     5,
			FallbackTimeout:      60 * time.Second,
			RetryEnabled:         true,
			MaxRetryAttempts:     3,
			BaseRetryDelay:       1 * time.Second,
			MaxRetryDelay:        30 * time.Second,
			RetryBackoffFactor:   2.0,
			CacheEnabled:         true,
			CacheTTL:             1 * time.Hour,
			CacheMaxSize:         1000,
			CacheCleanupInterval: 10 * time.Minute,
			TimeoutEnabled:       true,
			DefaultTimeout:       30 * time.Second,
			MethodTimeouts: map[VerificationMethodType]time.Duration{
				VerificationMethodDNS:     10 * time.Second,
				VerificationMethodWHOIS:   15 * time.Second,
				VerificationMethodContent: 30 * time.Second,
				VerificationMethodName:    5 * time.Second,
				VerificationMethodAddress: 10 * time.Second,
				VerificationMethodPhone:   5 * time.Second,
				VerificationMethodEmail:   10 * time.Second,
			},
			OverallTimeout: 120 * time.Second,
		}
	}

	fs := &FallbackStrategy{
		config:    config,
		logger:    logger,
		tracer:    tracer,
		verifiers: make(map[VerificationMethodType]Verifier),
	}

	// Initialize components
	fs.fallbackChain = &FallbackChain{
		chains: make(map[string][]VerificationMethodType),
		priorities: map[VerificationMethodType]int{
			VerificationMethodDNS:     1,
			VerificationMethodWHOIS:   2,
			VerificationMethodContent: 3,
			VerificationMethodName:    4,
			VerificationMethodAddress: 5,
			VerificationMethodPhone:   6,
			VerificationMethodEmail:   7,
		},
	}

	fs.retryManager = &RetryManager{
		enabled:       config.RetryEnabled,
		maxAttempts:   config.MaxRetryAttempts,
		baseDelay:     config.BaseRetryDelay,
		maxDelay:      config.MaxRetryDelay,
		backoffFactor: config.RetryBackoffFactor,
		retryHistory:  make(map[string]*RetryHistory),
	}

	fs.cacheManager = &CacheManager{
		enabled:         config.CacheEnabled,
		ttl:             config.CacheTTL,
		maxSize:         config.CacheMaxSize,
		cleanupInterval: config.CacheCleanupInterval,
		cache:           make(map[string]*CachedResult),
		stopCleanup:     make(chan bool),
	}

	fs.timeoutManager = &TimeoutManager{
		enabled:        config.TimeoutEnabled,
		defaultTimeout: config.DefaultTimeout,
		methodTimeouts: config.MethodTimeouts,
		overallTimeout: config.OverallTimeout,
		timeoutHistory: make(map[string]*TimeoutHistory),
	}

	// Start cache cleanup if enabled
	if config.CacheEnabled {
		fs.cacheManager.startCacheCleanup()
	}

	return fs
}

// RegisterVerifier registers a verification method
func (fs *FallbackStrategy) RegisterVerifier(verifier Verifier) {
	fs.verifiersMux.Lock()
	defer fs.verifiersMux.Unlock()

	methodType := verifier.GetMethodType()
	fs.verifiers[methodType] = verifier

	fs.logger.Info("verifier registered", map[string]interface{}{
		"method_type": methodType,
		"priority":    verifier.GetPriority(),
		"enabled":     verifier.IsEnabled(),
	})
}

// VerifyWithFallback performs verification with fallback strategies
func (fs *FallbackStrategy) VerifyWithFallback(ctx context.Context, domain, businessName, address, phone, email string) (*VerificationResult, error) {
	ctx, span := fs.tracer.Start(ctx, "FallbackStrategy.VerifyWithFallback")
	defer span.End()

	span.SetAttributes(
		attribute.String("domain", domain),
		attribute.String("business_name", businessName),
	)

	// Check cache first
	if fs.config.CacheEnabled {
		if cached := fs.cacheManager.Get(domain); cached != nil {
			fs.logger.Info("verification result found in cache", map[string]interface{}{
				"domain":    domain,
				"cached_at": cached.Timestamp,
			})
			return cached, nil
		}
	}

	// Create overall timeout context
	if fs.config.TimeoutEnabled {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, fs.config.OverallTimeout)
		defer cancel()
	}

	// Get fallback chain for this verification
	chain := fs.getFallbackChain(domain, businessName, address, phone, email)

	// Execute verification with fallback
	result, err := fs.executeWithFallback(ctx, chain, domain, businessName, address, phone, email)

	// Cache result if successful
	if err == nil && fs.config.CacheEnabled {
		fs.cacheManager.Set(domain, result)
	}

	return result, err
}

// getFallbackChain determines the fallback chain for verification
func (fs *FallbackStrategy) getFallbackChain(domain, businessName, address, phone, email string) []VerificationMethodType {
	fs.chainMux.RLock()
	defer fs.chainMux.RUnlock()

	// Create a key for this verification request
	key := fmt.Sprintf("%s:%s:%s:%s:%s", domain, businessName, address, phone, email)

	// Check if we have a cached chain
	if chain, exists := fs.fallbackChain.chains[key]; exists {
		return chain
	}

	// Build fallback chain based on available data and priorities
	var chain []VerificationMethodType

	// Always start with DNS and WHOIS
	chain = append(chain, VerificationMethodDNS, VerificationMethodWHOIS)

	// Add content verification if we have business name
	if businessName != "" {
		chain = append(chain, VerificationMethodContent)
	}

	// Add name matching if we have business name
	if businessName != "" {
		chain = append(chain, VerificationMethodName)
	}

	// Add address matching if we have address
	if address != "" {
		chain = append(chain, VerificationMethodAddress)
	}

	// Add phone matching if we have phone
	if phone != "" {
		chain = append(chain, VerificationMethodPhone)
	}

	// Add email verification if we have email
	if email != "" {
		chain = append(chain, VerificationMethodEmail)
	}

	// Limit chain depth
	if len(chain) > fs.config.MaxFallbackDepth {
		chain = chain[:fs.config.MaxFallbackDepth]
	}

	// Cache the chain
	fs.fallbackChain.chains[key] = chain

	return chain
}

// executeWithFallback executes verification with fallback logic
func (fs *FallbackStrategy) executeWithFallback(ctx context.Context, chain []VerificationMethodType, domain, businessName, address, phone, email string) (*VerificationResult, error) {
	var lastError error
	var bestResult *VerificationResult
	bestConfidence := 0.0

	// Try each method in the chain
	for i, methodType := range chain {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("verification cancelled: %w", ctx.Err())
		default:
		}

		// Get verifier for this method
		verifier := fs.getVerifier(methodType)
		if verifier == nil || !verifier.IsEnabled() {
			fs.logger.Debug("verifier not available or disabled", map[string]interface{}{
				"method_type": methodType,
			})
			continue
		}

		// Execute verification with retry logic
		result, err := fs.executeWithRetry(ctx, verifier, domain, businessName, address, phone, email)

		if err != nil {
			lastError = err
			fs.logger.Warn("verification method failed", map[string]interface{}{
				"method_type": methodType,
				"attempt":     i + 1,
				"error":       err.Error(),
			})
			continue
		}

		// Update best result if this one has higher confidence
		if result.Confidence > bestConfidence {
			bestResult = result
			bestConfidence = result.Confidence
		}

		// If we have a high confidence result, we can stop
		if result.Confidence >= 0.9 {
			fs.logger.Info("high confidence result achieved, stopping fallback chain", map[string]interface{}{
				"method_type": methodType,
				"confidence":  result.Confidence,
			})
			break
		}
	}

	// Return best result or error
	if bestResult != nil {
		return bestResult, nil
	}

	return nil, fmt.Errorf("all verification methods failed: %w", lastError)
}

// executeWithRetry executes a verification method with retry logic
func (fs *FallbackStrategy) executeWithRetry(ctx context.Context, verifier Verifier, domain, businessName, address, phone, email string) (*VerificationResult, error) {
	if !fs.config.RetryEnabled {
		return verifier.Verify(ctx, domain, businessName, address, phone, email)
	}

	methodType := verifier.GetMethodType()
	key := fmt.Sprintf("%s:%s", methodType, domain)

	// Get retry history
	fs.retryManager.retryHistoryMux.Lock()
	history, exists := fs.retryManager.retryHistory[key]
	if !exists {
		history = &RetryHistory{}
		fs.retryManager.retryHistory[key] = history
	}
	fs.retryManager.retryHistoryMux.Unlock()

	// Check if we've exceeded max attempts
	if history.Attempts >= fs.config.MaxRetryAttempts {
		return nil, fmt.Errorf("max retry attempts exceeded for %s", methodType)
	}

	var lastError error
	delay := fs.config.BaseRetryDelay

	for attempt := 0; attempt < fs.config.MaxRetryAttempts; attempt++ {
		// Create timeout context for this attempt
		timeoutCtx := ctx
		if fs.config.TimeoutEnabled {
			methodTimeout := fs.getMethodTimeout(methodType)
			var cancel context.CancelFunc
			timeoutCtx, cancel = context.WithTimeout(ctx, methodTimeout)
			defer cancel()
		}

		// Execute verification
		result, err := verifier.Verify(timeoutCtx, domain, businessName, address, phone, email)

		if err == nil {
			// Success - update history
			fs.updateRetryHistory(key, true, nil)
			return result, nil
		}

		lastError = err
		fs.updateRetryHistory(key, false, err)

		// If this is the last attempt, don't wait
		if attempt == fs.config.MaxRetryAttempts-1 {
			break
		}

		// Wait before retry with exponential backoff
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("verification cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}

		// Calculate next delay
		delay = time.Duration(float64(delay) * fs.config.RetryBackoffFactor)
		if delay > fs.config.MaxRetryDelay {
			delay = fs.config.MaxRetryDelay
		}
	}

	return nil, fmt.Errorf("verification failed after %d attempts: %w", fs.config.MaxRetryAttempts, lastError)
}

// getVerifier gets a verifier for a specific method type
func (fs *FallbackStrategy) getVerifier(methodType VerificationMethodType) Verifier {
	fs.verifiersMux.RLock()
	defer fs.verifiersMux.RUnlock()

	return fs.verifiers[methodType]
}

// getMethodTimeout gets the timeout for a specific method
func (fs *FallbackStrategy) getMethodTimeout(methodType VerificationMethodType) time.Duration {
	if timeout, exists := fs.config.MethodTimeouts[methodType]; exists {
		return timeout
	}
	return fs.config.DefaultTimeout
}

// updateRetryHistory updates retry history for a verification
func (fs *FallbackStrategy) updateRetryHistory(key string, success bool, err error) {
	fs.retryManager.retryHistoryMux.Lock()
	defer fs.retryManager.retryHistoryMux.Unlock()

	history := fs.retryManager.retryHistory[key]
	history.Attempts++
	history.LastAttempt = time.Now()

	if success {
		history.SuccessCount++
		history.LastError = nil
	} else {
		history.FailureCount++
		history.LastError = err
	}
}

// CacheManager methods

func (cm *CacheManager) Get(domain string) *VerificationResult {
	cm.cacheMux.RLock()
	defer cm.cacheMux.RUnlock()

	if cached, exists := cm.cache[domain]; exists {
		// Check if cache entry is still valid
		if time.Since(cached.Timestamp) < cached.TTL {
			cached.AccessCount++
			return cached.Result
		}
		// Remove expired entry
		delete(cm.cache, domain)
	}

	return nil
}

func (cm *CacheManager) Set(domain string, result *VerificationResult) {
	cm.cacheMux.Lock()
	defer cm.cacheMux.Unlock()

	// Check cache size limit
	if len(cm.cache) >= cm.maxSize {
		cm.evictOldest()
	}

	cm.cache[domain] = &CachedResult{
		Result:      result,
		Timestamp:   time.Now(),
		TTL:         cm.ttl,
		AccessCount: 1,
	}
}

func (cm *CacheManager) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, cached := range cm.cache {
		if oldestKey == "" || cached.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.Timestamp
		}
	}

	if oldestKey != "" {
		delete(cm.cache, oldestKey)
	}
}

func (cm *CacheManager) startCacheCleanup() {
	cm.cleanupTicker = time.NewTicker(cm.cleanupInterval)

	go func() {
		for {
			select {
			case <-cm.cleanupTicker.C:
				cm.cleanup()
			case <-cm.stopCleanup:
				cm.cleanupTicker.Stop()
				return
			}
		}
	}()
}

func (cm *CacheManager) cleanup() {
	cm.cacheMux.Lock()
	defer cm.cacheMux.Unlock()

	now := time.Now()
	for key, cached := range cm.cache {
		if now.Sub(cached.Timestamp) > cached.TTL {
			delete(cm.cache, key)
		}
	}
}

// GetRetryStatistics returns retry statistics for monitoring
func (fs *FallbackStrategy) GetRetryStatistics() map[string]*RetryHistory {
	fs.retryManager.retryHistoryMux.RLock()
	defer fs.retryManager.retryHistoryMux.RUnlock()

	stats := make(map[string]*RetryHistory)
	for key, history := range fs.retryManager.retryHistory {
		stats[key] = history
	}

	return stats
}

// GetCacheStatistics returns cache statistics for monitoring
func (fs *FallbackStrategy) GetCacheStatistics() map[string]interface{} {
	fs.cacheManager.cacheMux.RLock()
	defer fs.cacheManager.cacheMux.RUnlock()

	totalAccess := 0
	for _, cached := range fs.cacheManager.cache {
		totalAccess += cached.AccessCount
	}

	return map[string]interface{}{
		"cache_size":   len(fs.cacheManager.cache),
		"max_size":     fs.cacheManager.maxSize,
		"total_access": totalAccess,
		"enabled":      fs.cacheManager.enabled,
	}
}

// GetTimeoutStatistics returns timeout statistics for monitoring
func (fs *FallbackStrategy) GetTimeoutStatistics() map[string]*TimeoutHistory {
	fs.timeoutManager.timeoutHistoryMux.RLock()
	defer fs.timeoutManager.timeoutHistoryMux.RUnlock()

	stats := make(map[string]*TimeoutHistory)
	for key, history := range fs.timeoutManager.timeoutHistory {
		stats[key] = history
	}

	return stats
}

// Shutdown shuts down the fallback strategy
func (fs *FallbackStrategy) Shutdown() {
	if fs.cacheManager.cleanupTicker != nil {
		fs.cacheManager.stopCleanup <- true
	}

	fs.logger.Info("fallback strategy shutting down", map[string]interface{}{})
}
