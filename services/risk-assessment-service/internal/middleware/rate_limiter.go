package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RateLimiter implements distributed rate limiting using Redis
type RateLimiter struct {
	redisClient       *redis.Client
	logger            *zap.Logger
	fallbackLimiter   *InMemoryRateLimiter
	useRedis          bool
	requestsPerMinute int
	burstAllowance    int
	windowSize        time.Duration
}

// InMemoryRateLimiter implements a simple in-memory rate limiter as fallback
type InMemoryRateLimiter struct {
	requestsPerMinute int
	clients           map[string]*clientInfo
	mutex             sync.RWMutex
}

// clientInfo tracks rate limit information for a client
type clientInfo struct {
	requests   []time.Time
	lastAccess time.Time
}

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	RequestsPerMinute int           `json:"requests_per_minute"`
	BurstAllowance    int           `json:"burst_allowance"`
	WindowSize        time.Duration `json:"window_size"`
	UseRedis          bool          `json:"use_redis"`
	RedisKeyPrefix    string        `json:"redis_key_prefix"`
}

// RateLimitTier represents different rate limit tiers
type RateLimitTier string

const (
	TierFree       RateLimitTier = "free"
	TierPro        RateLimitTier = "pro"
	TierEnterprise RateLimitTier = "enterprise"
	TierInternal   RateLimitTier = "internal"
)

// NewRateLimiter creates a new distributed rate limiter
func NewRateLimiter(redisClient *redis.Client, logger *zap.Logger, config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		redisClient:       redisClient,
		logger:            logger,
		useRedis:          config.UseRedis && redisClient != nil,
		requestsPerMinute: config.RequestsPerMinute,
		burstAllowance:    config.BurstAllowance,
		windowSize:        config.WindowSize,
	}

	// Create fallback in-memory rate limiter
	rl.fallbackLimiter = NewInMemoryRateLimiter(config.RequestsPerMinute)

	// Test Redis connection
	if rl.useRedis {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err != nil {
			rl.logger.Warn("Redis connection failed, falling back to in-memory rate limiting", zap.Error(err))
			rl.useRedis = false
		}
	}

	return rl
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
func NewInMemoryRateLimiter(requestsPerMinute int) *InMemoryRateLimiter {
	rl := &InMemoryRateLimiter{
		requestsPerMinute: requestsPerMinute,
		clients:           make(map[string]*clientInfo),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request from the given client is allowed using Redis or fallback
func (rl *RateLimiter) Allow(clientIP string, tier RateLimitTier) bool {
	if rl.useRedis {
		return rl.allowRedis(clientIP, tier)
	}
	return rl.fallbackLimiter.Allow(clientIP)
}

// AllowWithAPIKey checks rate limit using API key instead of IP
func (rl *RateLimiter) AllowWithAPIKey(apiKey string, tier RateLimitTier) bool {
	if rl.useRedis {
		return rl.allowRedis(apiKey, tier)
	}
	return rl.fallbackLimiter.Allow(apiKey)
}

// allowRedis implements distributed rate limiting using Redis
func (rl *RateLimiter) allowRedis(identifier string, tier RateLimitTier) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get rate limit for tier
	limit := rl.getRateLimitForTier(tier)
	key := fmt.Sprintf("rate_limit:%s:%s", tier, identifier)

	// Use Redis sliding window counter
	now := time.Now().Unix()
	windowStart := now - int64(rl.windowSize.Seconds())

	// Use Redis pipeline for atomic operations
	pipe := rl.redisClient.Pipeline()

	// Remove old entries
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

	// Count current requests
	pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now),
		Member: now,
	})

	// Set expiration
	pipe.Expire(ctx, key, rl.windowSize)

	// Execute pipeline
	results, err := pipe.Exec(ctx)
	if err != nil {
		rl.logger.Error("Redis rate limit check failed", zap.Error(err))
		return rl.fallbackLimiter.Allow(identifier)
	}

	// Get current count
	currentCount := results[1].(*redis.IntCmd).Val()

	// Check if under limit
	if currentCount < int64(limit) {
		return true
	}

	// Log rate limit exceeded
	rl.logger.Warn("Rate limit exceeded",
		zap.String("identifier", identifier),
		zap.String("tier", string(tier)),
		zap.Int64("current_count", currentCount),
		zap.Int("limit", limit))

	return false
}

// getRateLimitForTier returns the rate limit for a given tier
func (rl *RateLimiter) getRateLimitForTier(tier RateLimitTier) int {
	switch tier {
	case TierFree:
		return 100 // 100 requests per hour
	case TierPro:
		return 1000 // 1000 requests per hour
	case TierEnterprise:
		return 10000 // 10000 requests per hour
	case TierInternal:
		return 100000 // 100000 requests per hour (internal services)
	default:
		return rl.requestsPerMinute
	}
}

// GetRateLimitInfo returns current rate limit information for a client
func (rl *RateLimiter) GetRateLimitInfo(identifier string, tier RateLimitTier) (current int, limit int, resetTime time.Time) {
	limit = rl.getRateLimitForTier(tier)

	if rl.useRedis {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		key := fmt.Sprintf("rate_limit:%s:%s", tier, identifier)
		count, err := rl.redisClient.ZCard(ctx, key).Result()
		if err != nil {
			rl.logger.Error("Failed to get rate limit info", zap.Error(err))
			return 0, limit, time.Now().Add(rl.windowSize)
		}

		// Get oldest entry to calculate reset time
		oldest, err := rl.redisClient.ZRangeWithScores(ctx, key, 0, 0).Result()
		if err != nil || len(oldest) == 0 {
			return int(count), limit, time.Now().Add(rl.windowSize)
		}

		resetTime = time.Unix(int64(oldest[0].Score), 0).Add(rl.windowSize)
		return int(count), limit, resetTime
	}

	// Fallback to in-memory
	return rl.fallbackLimiter.GetCurrentCount(identifier), limit, time.Now().Add(time.Minute)
}

// RateLimitMiddleware creates a middleware function for rate limiting
func (rl *RateLimiter) RateLimitMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client identifier
			clientIP := getClientIP(r)
			apiKey := r.Header.Get("X-API-Key")

			// Determine tier based on API key or default to free
			tier := rl.determineTier(apiKey)

			// Check rate limit
			var allowed bool
			if apiKey != "" {
				allowed = rl.AllowWithAPIKey(apiKey, tier)
			} else {
				allowed = rl.Allow(clientIP, tier)
			}

			if !allowed {
				// Get rate limit info for headers
				identifier := clientIP
				if apiKey != "" {
					identifier = apiKey
				}

				current, limit, resetTime := rl.GetRateLimitInfo(identifier, tier)

				// Set rate limit headers
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
				w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(max(0, limit-current)))
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
				w.Header().Set("X-RateLimit-Tier", string(tier))

				// Return 429 Too Many Requests
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// Add rate limit headers to successful responses
			identifier := clientIP
			if apiKey != "" {
				identifier = apiKey
			}

			current, limit, resetTime := rl.GetRateLimitInfo(identifier, tier)
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(max(0, limit-current)))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
			w.Header().Set("X-RateLimit-Tier", string(tier))

			next.ServeHTTP(w, r)
		})
	}
}

// determineTier determines the rate limit tier based on API key
func (rl *RateLimiter) determineTier(apiKey string) RateLimitTier {
	if apiKey == "" {
		return TierFree
	}

	// In a real implementation, you would look up the API key in the database
	// to determine the tier. For now, we'll use a simple prefix-based approach
	if len(apiKey) > 10 {
		switch apiKey[:3] {
		case "pro":
			return TierPro
		case "ent":
			return TierEnterprise
		case "int":
			return TierInternal
		default:
			return TierFree
		}
	}

	return TierFree
}

// In-memory rate limiter methods

// Allow checks if a request from the given client is allowed (in-memory)
func (imrl *InMemoryRateLimiter) Allow(clientIP string) bool {
	imrl.mutex.Lock()
	defer imrl.mutex.Unlock()

	now := time.Now()

	// Get or create client info
	client, exists := imrl.clients[clientIP]
	if !exists {
		client = &clientInfo{
			requests:   make([]time.Time, 0),
			lastAccess: now,
		}
		imrl.clients[clientIP] = client
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
	if len(client.requests) < imrl.requestsPerMinute {
		// Add current request
		client.requests = append(client.requests, now)
		client.lastAccess = now
		return true
	}

	// Over limit
	client.lastAccess = now
	return false
}

// GetCurrentCount returns the current request count for a client
func (imrl *InMemoryRateLimiter) GetCurrentCount(clientIP string) int {
	imrl.mutex.RLock()
	defer imrl.mutex.RUnlock()

	client, exists := imrl.clients[clientIP]
	if !exists {
		return 0
	}

	// Count valid requests
	now := time.Now()
	cutoff := now.Add(-time.Minute)
	count := 0
	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			count++
		}
	}

	return count
}

// cleanup removes old client entries to prevent memory leaks
func (imrl *InMemoryRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		imrl.mutex.Lock()
		now := time.Now()
		cutoff := now.Add(-10 * time.Minute) // Remove clients inactive for 10 minutes

		for ip, client := range imrl.clients {
			if client.lastAccess.Before(cutoff) {
				delete(imrl.clients, ip)
			}
		}
		imrl.mutex.Unlock()
	}
}

// Helper functions

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := len(xff); idx > 0 {
			for i, c := range xff {
				if c == ',' {
					idx = i
					break
				}
			}
			return xff[:idx]
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := len(ip); idx > 0 {
		for i, c := range ip {
			if c == ':' {
				idx = i
				break
			}
		}
		return ip[:idx]
	}

	return "unknown"
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
