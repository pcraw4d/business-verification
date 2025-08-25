package middleware

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// RateLimitStore defines the interface for rate limit storage
type RateLimitStore interface {
	Allow(key string, limit int, window time.Duration) (bool, int, error)
	GetRemaining(key string) (int, error)
	Reset(key string) error
	Cleanup() error
}

// MemoryRateLimitStore implements in-memory rate limiting storage
type MemoryRateLimitStore struct {
	strategy  string
	maxKeys   int
	logger    *zap.Logger
	mu        sync.RWMutex
	limiters  map[string]*TokenBucket
	lastClean time.Time
}

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewMemoryRateLimitStore creates a new in-memory rate limit store
func NewMemoryRateLimitStore(strategy string, maxKeys int, logger *zap.Logger) *MemoryRateLimitStore {
	return &MemoryRateLimitStore{
		strategy:  strategy,
		maxKeys:   maxKeys,
		logger:    logger,
		limiters:  make(map[string]*TokenBucket),
		lastClean: time.Now(),
	}
}

// Allow checks if a request should be allowed for the given key
func (store *MemoryRateLimitStore) Allow(key string, limit int, window time.Duration) (bool, int, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	bucket, exists := store.limiters[key]
	if !exists {
		// Create new bucket
		if len(store.limiters) >= store.maxKeys {
			// Clean up old entries if we're at capacity
			store.Cleanup()
		}

		bucket = &TokenBucket{
			tokens:     limit,
			maxTokens:  limit,
			refillRate: window / time.Duration(limit),
			lastRefill: time.Now(),
		}
		store.limiters[key] = bucket
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)

	// Refill tokens based on elapsed time
	tokensToAdd := int(elapsed / bucket.refillRate)
	if tokensToAdd > 0 {
		bucket.tokens = minInt(bucket.maxTokens, bucket.tokens+tokensToAdd)
		bucket.lastRefill = now
	}

	// Check if we have tokens available
	if bucket.tokens > 0 {
		bucket.tokens--
		return true, bucket.tokens, nil
	}

	return false, 0, nil
}

// GetRemaining returns the number of remaining requests for a key
func (store *MemoryRateLimitStore) GetRemaining(key string) (int, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	bucket, exists := store.limiters[key]
	if !exists {
		return 0, nil
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	return bucket.tokens, nil
}

// Reset resets the rate limit for a specific key
func (store *MemoryRateLimitStore) Reset(key string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	delete(store.limiters, key)
	return nil
}

// Cleanup removes expired entries
func (store *MemoryRateLimitStore) Cleanup() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	now := time.Now()
	if now.Sub(store.lastClean) < 5*time.Minute {
		return nil // Don't clean too frequently
	}

	store.lastClean = now
	expiredKeys := []string{}

	for key, bucket := range store.limiters {
		bucket.mu.Lock()
		if now.Sub(bucket.lastRefill) > 10*time.Minute {
			expiredKeys = append(expiredKeys, key)
		}
		bucket.mu.Unlock()
	}

	for _, key := range expiredKeys {
		delete(store.limiters, key)
	}

	if len(expiredKeys) > 0 {
		store.logger.Debug("Cleaned up expired rate limit entries", zap.Int("count", len(expiredKeys)))
	}

	return nil
}

// RedisRateLimitStore implements Redis-based rate limiting storage
type RedisRateLimitStore struct {
	redisURL  string
	keyPrefix string
	logger    *zap.Logger
	client    interface{} // Will be Redis client when implemented
	mu        sync.RWMutex
}

// NewRedisRateLimitStore creates a new Redis-based rate limit store
func NewRedisRateLimitStore(redisURL, keyPrefix string, logger *zap.Logger) *RedisRateLimitStore {
	store := &RedisRateLimitStore{
		redisURL:  redisURL,
		keyPrefix: keyPrefix,
		logger:    logger,
	}

	// TODO: Initialize Redis client
	// For now, fall back to memory store
	logger.Warn("Redis rate limiting not implemented, falling back to memory store")

	return store
}

// Allow checks if a request should be allowed for the given key
func (store *RedisRateLimitStore) Allow(key string, limit int, window time.Duration) (bool, int, error) {
	// TODO: Implement Redis-based rate limiting
	// For now, always allow requests
	store.logger.Debug("Redis rate limiting not implemented, allowing request", zap.String("key", key))
	return true, limit - 1, nil
}

// GetRemaining returns the number of remaining requests for a key
func (store *RedisRateLimitStore) GetRemaining(key string) (int, error) {
	// TODO: Implement Redis-based remaining count
	return 0, nil
}

// Reset resets the rate limit for a specific key
func (store *RedisRateLimitStore) Reset(key string) error {
	// TODO: Implement Redis-based reset
	return nil
}

// Cleanup removes expired entries
func (store *RedisRateLimitStore) Cleanup() error {
	// TODO: Implement Redis-based cleanup
	return nil
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AuthAttemptRecord represents a failed authentication attempt
type AuthAttemptRecord struct {
	AttemptType string    `json:"attempt_type"`
	Timestamp   time.Time `json:"timestamp"`
	Count       int       `json:"count"`
}

// AuthLockoutRecord represents a lockout record
type AuthLockoutRecord struct {
	LockoutUntil time.Time `json:"lockout_until"`
	Reason       string    `json:"reason"`
	Count        int       `json:"count"`
}

// MemoryAuthRateLimitStore implements in-memory authentication rate limiting storage
type MemoryAuthRateLimitStore struct {
	logger    *zap.Logger
	mu        sync.RWMutex
	attempts  map[string]*AuthAttemptRecord
	lockouts  map[string]*AuthLockoutRecord
	lastClean time.Time
}

// NewMemoryAuthRateLimitStore creates a new in-memory auth rate limit store
func NewMemoryAuthRateLimitStore(logger *zap.Logger) *MemoryAuthRateLimitStore {
	return &MemoryAuthRateLimitStore{
		logger:    logger,
		attempts:  make(map[string]*AuthAttemptRecord),
		lockouts:  make(map[string]*AuthLockoutRecord),
		lastClean: time.Now(),
	}
}

// CheckAuthLimit checks if an authentication attempt should be allowed
func (store *MemoryAuthRateLimitStore) CheckAuthLimit(key string, limit int, window time.Duration) (bool, int, time.Time, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	now := time.Now()
	record, exists := store.attempts[key]

	if !exists {
		// First attempt
		store.attempts[key] = &AuthAttemptRecord{
			AttemptType: "auth",
			Timestamp:   now,
			Count:       1,
		}
		return true, limit - 1, now.Add(window), nil
	}

	// Check if the window has expired
	if now.Sub(record.Timestamp) > window {
		// Reset the record
		record.Timestamp = now
		record.Count = 1
		return true, limit - 1, now.Add(window), nil
	}

	// Check if we're within the limit
	if record.Count < limit {
		record.Count++
		return true, limit - record.Count, record.Timestamp.Add(window), nil
	}

	// Rate limit exceeded
	return false, 0, record.Timestamp.Add(window), nil
}

// RecordFailedAttempt records a failed authentication attempt
func (store *MemoryAuthRateLimitStore) RecordFailedAttempt(key string, attemptType string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	now := time.Now()
	record, exists := store.attempts[key]

	if !exists {
		store.attempts[key] = &AuthAttemptRecord{
			AttemptType: attemptType,
			Timestamp:   now,
			Count:       1,
		}
		return nil
	}

	// Update the record
	record.Count++
	record.AttemptType = attemptType

	// Check if we should create a lockout
	if record.Count >= 5 { // Configurable threshold
		lockoutDuration := 15 * time.Minute // Configurable
		store.lockouts[key] = &AuthLockoutRecord{
			LockoutUntil: now.Add(lockoutDuration),
			Reason:       "Too many failed attempts",
			Count:        record.Count,
		}
	}

	return nil
}

// IsLocked checks if a key is currently locked out
func (store *MemoryAuthRateLimitStore) IsLocked(key string) (bool, time.Time, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	lockout, exists := store.lockouts[key]
	if !exists {
		return false, time.Time{}, nil
	}

	now := time.Now()
	if now.Before(lockout.LockoutUntil) {
		return true, lockout.LockoutUntil, nil
	}

	// Lockout has expired, remove it
	store.mu.RUnlock()
	store.mu.Lock()
	delete(store.lockouts, key)
	store.mu.Unlock()
	store.mu.RLock()

	return false, time.Time{}, nil
}

// Reset resets the rate limit for a specific key
func (store *MemoryAuthRateLimitStore) Reset(key string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	delete(store.attempts, key)
	delete(store.lockouts, key)
	return nil
}

// Cleanup removes expired entries
func (store *MemoryAuthRateLimitStore) Cleanup() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	now := time.Now()
	if now.Sub(store.lastClean) < 5*time.Minute {
		return nil // Don't clean too frequently
	}

	store.lastClean = now
	expiredAttempts := []string{}
	expiredLockouts := []string{}

	// Clean up expired attempts
	for key, record := range store.attempts {
		if now.Sub(record.Timestamp) > 10*time.Minute {
			expiredAttempts = append(expiredAttempts, key)
		}
	}

	// Clean up expired lockouts
	for key, lockout := range store.lockouts {
		if now.After(lockout.LockoutUntil) {
			expiredLockouts = append(expiredLockouts, key)
		}
	}

	// Remove expired entries
	for _, key := range expiredAttempts {
		delete(store.attempts, key)
	}
	for _, key := range expiredLockouts {
		delete(store.lockouts, key)
	}

	if len(expiredAttempts) > 0 || len(expiredLockouts) > 0 {
		store.logger.Debug("Cleaned up expired auth rate limit entries",
			zap.Int("attempts", len(expiredAttempts)),
			zap.Int("lockouts", len(expiredLockouts)))
	}

	return nil
}

// RedisAuthRateLimitStore implements Redis-based authentication rate limiting storage
type RedisAuthRateLimitStore struct {
	redisURL  string
	keyPrefix string
	logger    *zap.Logger
	client    interface{} // Will be Redis client when implemented
	mu        sync.RWMutex
}

// NewRedisAuthRateLimitStore creates a new Redis-based auth rate limit store
func NewRedisAuthRateLimitStore(redisURL, keyPrefix string, logger *zap.Logger) *RedisAuthRateLimitStore {
	store := &RedisAuthRateLimitStore{
		redisURL:  redisURL,
		keyPrefix: keyPrefix,
		logger:    logger,
	}

	// TODO: Initialize Redis client
	// For now, fall back to memory store
	logger.Warn("Redis auth rate limiting not implemented, falling back to memory store")

	return store
}

// CheckAuthLimit checks if an authentication attempt should be allowed
func (store *RedisAuthRateLimitStore) CheckAuthLimit(key string, limit int, window time.Duration) (bool, int, time.Time, error) {
	// TODO: Implement Redis-based auth rate limiting
	// For now, always allow requests
	store.logger.Debug("Redis auth rate limiting not implemented, allowing request", zap.String("key", key))
	return true, limit - 1, time.Now().Add(window), nil
}

// RecordFailedAttempt records a failed authentication attempt
func (store *RedisAuthRateLimitStore) RecordFailedAttempt(key string, attemptType string) error {
	// TODO: Implement Redis-based failed attempt recording
	return nil
}

// IsLocked checks if a key is currently locked out
func (store *RedisAuthRateLimitStore) IsLocked(key string) (bool, time.Time, error) {
	// TODO: Implement Redis-based lockout checking
	return false, time.Time{}, nil
}

// Reset resets the rate limit for a specific key
func (store *RedisAuthRateLimitStore) Reset(key string) error {
	// TODO: Implement Redis-based reset
	return nil
}

// Cleanup removes expired entries
func (store *RedisAuthRateLimitStore) Cleanup() error {
	// TODO: Implement Redis-based cleanup
	return nil
}
