package security

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements per-user rate limiting
type RateLimiter struct {
	mu sync.RWMutex

	// Configuration
	requestsPerMinute int
	burstSize         int

	// User-specific rate limiting
	userLimits map[string]*UserLimit
}

// UserLimit tracks rate limiting for a specific user
type UserLimit struct {
	requests    []time.Time
	lastCleanup time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute, burstSize int) *RateLimiter {
	return &RateLimiter{
		requestsPerMinute: requestsPerMinute,
		burstSize:         burstSize,
		userLimits:        make(map[string]*UserLimit),
	}
}

// Allow checks if a request is allowed for the given user
func (rl *RateLimiter) Allow(ctx context.Context, userID string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Get or create user limit
	userLimit, exists := rl.userLimits[userID]
	if !exists {
		userLimit = &UserLimit{
			requests:    make([]time.Time, 0),
			lastCleanup: time.Now(),
		}
		rl.userLimits[userID] = userLimit
	}

	now := time.Now()

	// Clean up old requests (older than 1 minute)
	cutoff := now.Add(-1 * time.Minute)
	validRequests := make([]time.Time, 0)
	for _, reqTime := range userLimit.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	userLimit.requests = validRequests

	// Check if user has exceeded rate limit
	if len(userLimit.requests) >= rl.requestsPerMinute {
		return false, fmt.Errorf("rate limit exceeded for user %s", userID)
	}

	// Check burst limit
	if len(userLimit.requests) >= rl.burstSize {
		return false, fmt.Errorf("burst limit exceeded for user %s", userID)
	}

	// Add current request
	userLimit.requests = append(userLimit.requests, now)
	userLimit.lastCleanup = now

	return true, nil
}

// GetUserStats returns rate limiting stats for a user
func (rl *RateLimiter) GetUserStats(userID string) map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	userLimit, exists := rl.userLimits[userID]
	if !exists {
		return map[string]interface{}{
			"user_id":             userID,
			"requests_count":      0,
			"requests_per_minute": rl.requestsPerMinute,
			"burst_size":          rl.burstSize,
			"status":              "no_requests",
		}
	}

	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)
	validRequests := 0

	for _, reqTime := range userLimit.requests {
		if reqTime.After(cutoff) {
			validRequests++
		}
	}

	return map[string]interface{}{
		"user_id":             userID,
		"requests_count":      validRequests,
		"requests_per_minute": rl.requestsPerMinute,
		"burst_size":          rl.burstSize,
		"status":              "active",
		"last_cleanup":        userLimit.lastCleanup,
	}
}

// GetAllStats returns rate limiting stats for all users
func (rl *RateLimiter) GetAllStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := make(map[string]interface{})
	for userID := range rl.userLimits {
		stats[userID] = rl.GetUserStats(userID)
	}

	return map[string]interface{}{
		"total_users":         len(rl.userLimits),
		"requests_per_minute": rl.requestsPerMinute,
		"burst_size":          rl.burstSize,
		"user_stats":          stats,
	}
}

// Cleanup removes old user limits to prevent memory leaks
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-5 * time.Minute) // Remove users inactive for 5 minutes

	for userID, userLimit := range rl.userLimits {
		if userLimit.lastCleanup.Before(cutoff) {
			delete(rl.userLimits, userID)
		}
	}
}

// StartCleanup starts a background cleanup routine
func (rl *RateLimiter) StartCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rl.Cleanup()
		}
	}
}
