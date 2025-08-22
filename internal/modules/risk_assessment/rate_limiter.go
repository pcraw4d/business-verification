package risk_assessment

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RateLimiter provides rate limiting capabilities
type RateLimiter struct {
	config *RiskAssessmentConfig
	logger *zap.Logger
	mu     sync.RWMutex
	limits map[string]*APIRateLimit
}

// APIRateLimit contains rate limit information for an API
type APIRateLimit struct {
	APIEndpoint       string
	RequestsPerMinute int
	RequestsPerHour   int
	RequestsPerDay    int
	CurrentRequests   int
	LastResetTime     time.Time
	QuotaExceeded     bool
	RetryAfter        time.Time
}

// RateLimitResult contains rate limit check result
type RateLimitResult struct {
	Allowed           bool
	RemainingRequests int
	ResetTime         time.Time
	RetryAfter        time.Time
	QuotaExceeded     bool
	WaitTime          time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RiskAssessmentConfig, logger *zap.Logger) *RateLimiter {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &RateLimiter{
		config: config,
		logger: logger,
		limits: make(map[string]*APIRateLimit),
	}
}

// CheckRateLimit checks if a request is allowed based on rate limits
func (rl *RateLimiter) CheckRateLimit(ctx context.Context, apiEndpoint string) (*RateLimitResult, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, exists := rl.limits[apiEndpoint]
	if !exists {
		// Create default rate limit
		limit = &APIRateLimit{
			APIEndpoint:       apiEndpoint,
			RequestsPerMinute: rl.config.RateLimitPerMinute,
			RequestsPerHour:   rl.config.RateLimitPerMinute * 60,
			RequestsPerDay:    rl.config.RateLimitPerMinute * 60 * 24,
			LastResetTime:     time.Now(),
		}
		rl.limits[apiEndpoint] = limit
	}

	// Check if we need to reset the counter
	now := time.Now()
	if now.Sub(limit.LastResetTime) >= time.Minute {
		limit.CurrentRequests = 0
		limit.LastResetTime = now
		limit.QuotaExceeded = false
	}

	result := &RateLimitResult{
		ResetTime: limit.LastResetTime.Add(time.Minute),
	}

	// Check if we're within limits
	if limit.CurrentRequests < limit.RequestsPerMinute {
		limit.CurrentRequests++
		result.Allowed = true
		result.RemainingRequests = limit.RequestsPerMinute - limit.CurrentRequests
	} else {
		result.Allowed = false
		result.RemainingRequests = 0
		result.QuotaExceeded = true
		result.RetryAfter = limit.LastResetTime.Add(time.Minute)
		result.WaitTime = result.RetryAfter.Sub(now)
		limit.QuotaExceeded = true
	}

	return result, nil
}

// WaitForRateLimit waits until rate limit allows the request
func (rl *RateLimiter) WaitForRateLimit(ctx context.Context, apiEndpoint string) error {
	for {
		result, err := rl.CheckRateLimit(ctx, apiEndpoint)
		if err != nil {
			return err
		}

		if result.Allowed {
			return nil
		}

		// Wait for the rate limit to reset
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(result.WaitTime):
			continue
		}
	}
}

// GetRateLimitStatus gets the current rate limit status for an API
func (rl *RateLimiter) GetRateLimitStatus(apiEndpoint string) *APIRateLimit {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if limit, exists := rl.limits[apiEndpoint]; exists {
		return limit
	}
	return nil
}

// ResetRateLimit resets the rate limit for an API
func (rl *RateLimiter) ResetRateLimit(apiEndpoint string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if limit, exists := rl.limits[apiEndpoint]; exists {
		limit.CurrentRequests = 0
		limit.LastResetTime = time.Now()
		limit.QuotaExceeded = false
	}
}
