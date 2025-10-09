package middleware

import (
	"sync"
	"time"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requestsPerMinute int
	clients           map[string]*clientInfo
	mutex             sync.RWMutex
}

// clientInfo tracks rate limit information for a client
type clientInfo struct {
	requests   []time.Time
	lastAccess time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		requestsPerMinute: requestsPerMinute,
		clients:           make(map[string]*clientInfo),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request from the given client is allowed
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Get or create client info
	client, exists := rl.clients[clientIP]
	if !exists {
		client = &clientInfo{
			requests:   make([]time.Time, 0),
			lastAccess: now,
		}
		rl.clients[clientIP] = client
	}

	// Clean old requests (older than 1 minute)
	cutoff := now.Add(-time.Minute)
	validRequests := make([]time.Time, 0)
	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	client.requests = validRequests

	// Check if under limit
	if len(client.requests) < rl.requestsPerMinute {
		// Add current request
		client.requests = append(client.requests, now)
		client.lastAccess = now
		return true
	}

	// Over limit
	client.lastAccess = now
	return false
}

// GetRemaining returns the number of remaining requests for a client
func (rl *RateLimiter) GetRemaining(clientIP string) int {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	client, exists := rl.clients[clientIP]
	if !exists {
		return rl.requestsPerMinute
	}

	now := time.Now()
	cutoff := now.Add(-time.Minute)

	// Count valid requests
	validCount := 0
	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			validCount++
		}
	}

	remaining := rl.requestsPerMinute - validCount
	if remaining < 0 {
		remaining = 0
	}

	return remaining
}

// GetResetTime returns when the rate limit resets for a client
func (rl *RateLimiter) GetResetTime(clientIP string) time.Time {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	client, exists := rl.clients[clientIP]
	if !exists {
		return time.Now().Add(time.Minute)
	}

	if len(client.requests) == 0 {
		return time.Now().Add(time.Minute)
	}

	// Find the oldest request
	oldest := client.requests[0]
	for _, reqTime := range client.requests {
		if reqTime.Before(oldest) {
			oldest = reqTime
		}
	}

	// Reset time is 1 minute after the oldest request
	return oldest.Add(time.Minute)
}

// cleanup removes old client entries to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()

		now := time.Now()
		cutoff := now.Add(-10 * time.Minute) // Remove clients inactive for 10 minutes

		for clientIP, client := range rl.clients {
			if client.lastAccess.Before(cutoff) {
				delete(rl.clients, clientIP)
			}
		}

		rl.mutex.Unlock()
	}
}

// GetStats returns rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	now := time.Now()
	cutoff := now.Add(-time.Minute)

	totalClients := len(rl.clients)
	activeClients := 0
	totalRequests := 0

	for _, client := range rl.clients {
		if client.lastAccess.After(cutoff) {
			activeClients++
		}

		// Count valid requests
		for _, reqTime := range client.requests {
			if reqTime.After(cutoff) {
				totalRequests++
			}
		}
	}

	return map[string]interface{}{
		"requests_per_minute": rl.requestsPerMinute,
		"total_clients":       totalClients,
		"active_clients":      activeClients,
		"total_requests":      totalRequests,
	}
}
