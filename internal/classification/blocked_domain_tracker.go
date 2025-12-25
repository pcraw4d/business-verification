package classification

import (
	"sync"
	"time"
)

// BlockedDomainTracker tracks domains that have returned 403 errors
// Implements circuit breaker pattern: opens after 3 consecutive 403s, auto-resets after 1 hour
type BlockedDomainTracker struct {
	blockedDomains map[string]*blockedDomainInfo
	mutex          sync.RWMutex
}

// blockedDomainInfo stores information about a blocked domain
type blockedDomainInfo struct {
	blockedAt    time.Time
	failureCount int
	lastFailure  time.Time
}

// NewBlockedDomainTracker creates a new BlockedDomainTracker
func NewBlockedDomainTracker() *BlockedDomainTracker {
	return &BlockedDomainTracker{
		blockedDomains: make(map[string]*blockedDomainInfo),
	}
}

// IsBlocked checks if a domain is blocked (circuit breaker is open)
// Returns true if domain has 3+ consecutive 403s and last failure was < 1 hour ago
func (bdt *BlockedDomainTracker) IsBlocked(domain string) bool {
	if domain == "" {
		return false
	}

	bdt.mutex.RLock()
	defer bdt.mutex.RUnlock()

	info, exists := bdt.blockedDomains[domain]
	if !exists {
		return false
	}

	// Check if entry is recent (< 1 hour old)
	age := time.Since(info.lastFailure)
	if age >= 1*time.Hour {
		// Entry expired, circuit breaker auto-resets
		return false
	}

	// Circuit breaker is open if we have 3+ consecutive failures
	return info.failureCount >= 3
}

// RecordFailure records a 403 failure for a domain
// Increments failure count and updates last failure time
func (bdt *BlockedDomainTracker) RecordFailure(domain string) {
	if domain == "" {
		return
	}

	bdt.mutex.Lock()
	defer bdt.mutex.Unlock()

	now := time.Now()

	// Clean up old entries (older than 1 hour) to prevent memory leak
	cutoff := now.Add(-1 * time.Hour)
	cleanedCount := 0
	for key, info := range bdt.blockedDomains {
		if info.lastFailure.Before(cutoff) {
			delete(bdt.blockedDomains, key)
			cleanedCount++
		}
	}
	if cleanedCount > 0 {
		// Log cleanup if we have a logger (for now, just track it)
		// In production, inject a logger
	}

	// Get or create domain info
	info, exists := bdt.blockedDomains[domain]
	if !exists {
		info = &blockedDomainInfo{
			blockedAt:    now,
			failureCount: 0,
			lastFailure:  now,
		}
		bdt.blockedDomains[domain] = info
	}

	// Increment failure count
	info.failureCount++
	info.lastFailure = now

	// Log circuit breaker state transitions
	if info.failureCount == 3 {
		// Circuit breaker just opened
		// In production, log this event
	}
}

// Reset resets the failure count for a domain (circuit breaker closes)
// Useful when a domain starts working again
func (bdt *BlockedDomainTracker) Reset(domain string) {
	if domain == "" {
		return
	}

	bdt.mutex.Lock()
	defer bdt.mutex.Unlock()

	delete(bdt.blockedDomains, domain)
}

// GetFailureCount returns the current failure count for a domain
func (bdt *BlockedDomainTracker) GetFailureCount(domain string) int {
	if domain == "" {
		return 0
	}

	bdt.mutex.RLock()
	defer bdt.mutex.RUnlock()

	info, exists := bdt.blockedDomains[domain]
	if !exists {
		return 0
	}

	return info.failureCount
}

