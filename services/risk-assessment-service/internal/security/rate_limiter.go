package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RateLimiter provides rate limiting and DDoS protection
type RateLimiter struct {
	redisClient *redis.Client
	logger      *zap.Logger
	config      RateLimitConfig
	localCache  *sync.Map
}

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	RequestsPerMinute int           `json:"requests_per_minute"`
	BurstAllowance    int           `json:"burst_allowance"`
	WindowSize        time.Duration `json:"window_size"`
	BlockDuration     time.Duration `json:"block_duration"`
	EnableRedis       bool          `json:"enable_redis"`
	EnableLocalCache  bool          `json:"enable_local_cache"`
}

// RateLimitResult represents the result of rate limiting check
type RateLimitResult struct {
	Allowed     bool          `json:"allowed"`
	Remaining   int           `json:"remaining"`
	ResetTime   time.Time     `json:"reset_time"`
	RetryAfter  time.Duration `json:"retry_after,omitempty"`
	Reason      string        `json:"reason,omitempty"`
	IsBlocked   bool          `json:"is_blocked"`
	BlockExpiry time.Time     `json:"block_expiry,omitempty"`
}

// ClientInfo holds information about a client
type ClientInfo struct {
	IP           string    `json:"ip"`
	UserAgent    string    `json:"user_agent"`
	Country      string    `json:"country"`
	FirstSeen    time.Time `json:"first_seen"`
	LastSeen     time.Time `json:"last_seen"`
	RequestCount int       `json:"request_count"`
	BlockCount   int       `json:"block_count"`
	IsBlocked    bool      `json:"is_blocked"`
	BlockExpiry  time.Time `json:"block_expiry"`
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisClient *redis.Client, logger *zap.Logger, config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		logger:      logger,
		config:      config,
		localCache:  &sync.Map{},
	}
}

// Allow checks if a request should be allowed based on rate limiting rules
func (rl *RateLimiter) Allow(ctx context.Context, clientIP string, userAgent string) *RateLimitResult {
	now := time.Now()

	// Check if client is currently blocked
	if rl.isClientBlocked(ctx, clientIP) {
		return &RateLimitResult{
			Allowed:    false,
			IsBlocked:  true,
			Reason:     "client is blocked due to excessive requests",
			RetryAfter: rl.getBlockRemainingTime(ctx, clientIP),
		}
	}

	// Get or create client info
	clientInfo := rl.getClientInfo(ctx, clientIP, userAgent)

	// Reset request count if window has passed
	if now.Sub(clientInfo.LastSeen) > rl.config.WindowSize {
		clientInfo.RequestCount = 0
	}

	// Update request count
	clientInfo.RequestCount++
	clientInfo.LastSeen = now

	// Check rate limits
	allowed, remaining, resetTime := rl.checkRateLimit(ctx, clientIP, clientInfo)

	if !allowed {
		// Increment block count
		clientInfo.BlockCount++

		// Block client if they exceed threshold
		if clientInfo.BlockCount >= 3 {
			rl.blockClient(ctx, clientIP, clientInfo)
			return &RateLimitResult{
				Allowed:    false,
				IsBlocked:  true,
				Reason:     "client blocked due to repeated rate limit violations",
				RetryAfter: rl.config.BlockDuration,
			}
		}

		return &RateLimitResult{
			Allowed:    false,
			Remaining:  remaining,
			ResetTime:  resetTime,
			RetryAfter: time.Until(resetTime),
			Reason:     "rate limit exceeded",
		}
	}

	// Reset block count on successful request
	if clientInfo.BlockCount > 0 {
		clientInfo.BlockCount = 0
	}

	// Update client info
	rl.updateClientInfo(ctx, clientIP, clientInfo)

	return &RateLimitResult{
		Allowed:   true,
		Remaining: remaining,
		ResetTime: resetTime,
	}
}

// isClientBlocked checks if a client is currently blocked
func (rl *RateLimiter) isClientBlocked(ctx context.Context, clientIP string) bool {
	if rl.config.EnableRedis && rl.redisClient != nil {
		blockKey := fmt.Sprintf("block:%s", clientIP)
		exists, err := rl.redisClient.Exists(ctx, blockKey).Result()
		if err == nil && exists > 0 {
			return true
		}
	}

	if rl.config.EnableLocalCache {
		if info, ok := rl.localCache.Load(clientIP); ok {
			clientInfo := info.(*ClientInfo)
			return clientInfo.IsBlocked && clientInfo.BlockExpiry.After(time.Now())
		}
	}

	return false
}

// getBlockRemainingTime returns the remaining block time for a client
func (rl *RateLimiter) getBlockRemainingTime(ctx context.Context, clientIP string) time.Duration {
	if rl.config.EnableRedis && rl.redisClient != nil {
		blockKey := fmt.Sprintf("block:%s", clientIP)
		ttl, err := rl.redisClient.TTL(ctx, blockKey).Result()
		if err == nil && ttl > 0 {
			return ttl
		}
	}

	if rl.config.EnableLocalCache {
		if info, ok := rl.localCache.Load(clientIP); ok {
			clientInfo := info.(*ClientInfo)
			if clientInfo.IsBlocked && clientInfo.BlockExpiry.After(time.Now()) {
				return time.Until(clientInfo.BlockExpiry)
			}
		}
	}

	return 0
}

// getClientInfo retrieves or creates client information
func (rl *RateLimiter) getClientInfo(ctx context.Context, clientIP, userAgent string) *ClientInfo {
	now := time.Now()

	// Try Redis first
	if rl.config.EnableRedis && rl.redisClient != nil {
		clientKey := fmt.Sprintf("client:%s", clientIP)
		info, err := rl.redisClient.HGetAll(ctx, clientKey).Result()
		if err == nil && len(info) > 0 {
			clientInfo := &ClientInfo{
				IP:           clientIP,
				UserAgent:    info["user_agent"],
				Country:      info["country"],
				FirstSeen:    parseTime(info["first_seen"]),
				LastSeen:     parseTime(info["last_seen"]),
				RequestCount: parseInt(info["request_count"]),
				BlockCount:   parseInt(info["block_count"]),
				IsBlocked:    parseBool(info["is_blocked"]),
				BlockExpiry:  parseTime(info["block_expiry"]),
			}
			return clientInfo
		}
	}

	// Try local cache
	if rl.config.EnableLocalCache {
		if info, ok := rl.localCache.Load(clientIP); ok {
			return info.(*ClientInfo)
		}
	}

	// Create new client info
	clientInfo := &ClientInfo{
		IP:           clientIP,
		UserAgent:    userAgent,
		FirstSeen:    now,
		LastSeen:     now,
		RequestCount: 0,
		BlockCount:   0,
		IsBlocked:    false,
	}

	return clientInfo
}

// checkRateLimit checks if the client is within rate limits
func (rl *RateLimiter) checkRateLimit(ctx context.Context, clientIP string, clientInfo *ClientInfo) (bool, int, time.Time) {
	now := time.Now()
	windowStart := now.Add(-rl.config.WindowSize)

	// Try Redis first
	if rl.config.EnableRedis && rl.redisClient != nil {
		return rl.checkRedisRateLimit(ctx, clientIP, windowStart)
	}

	// Fall back to local cache
	if rl.config.EnableLocalCache {
		return rl.checkLocalRateLimit(clientIP, clientInfo, windowStart)
	}

	// No rate limiting configured
	return true, rl.config.RequestsPerMinute, now.Add(rl.config.WindowSize)
}

// checkRedisRateLimit checks rate limits using Redis
func (rl *RateLimiter) checkRedisRateLimit(ctx context.Context, clientIP string, windowStart time.Time) (bool, int, time.Time) {
	key := fmt.Sprintf("rate_limit:%s", clientIP)
	now := time.Now()

	// Use Redis sliding window counter
	pipe := rl.redisClient.Pipeline()

	// Remove old entries
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.Unix()))

	// Add current request
	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now.Unix()),
		Member: now.UnixNano(),
	})

	// Count requests in window
	pipe.ZCard(ctx, key)

	// Set expiration
	pipe.Expire(ctx, key, rl.config.WindowSize)

	results, err := pipe.Exec(ctx)
	if err != nil {
		rl.logger.Error("Redis rate limit check failed", zap.Error(err))
		return true, rl.config.RequestsPerMinute, now.Add(rl.config.WindowSize)
	}

	count := results[2].(*redis.IntCmd).Val()
	remaining := rl.config.RequestsPerMinute - int(count)
	resetTime := now.Add(rl.config.WindowSize)

	allowed := count <= int64(rl.config.RequestsPerMinute)
	return allowed, remaining, resetTime
}

// checkLocalRateLimit checks rate limits using local cache
func (rl *RateLimiter) checkLocalRateLimit(clientIP string, clientInfo *ClientInfo, windowStart time.Time) (bool, int, time.Time) {
	now := time.Now()

	// Simple counter-based rate limiting
	// In a real implementation, you'd want to use a sliding window
	allowed := clientInfo.RequestCount < rl.config.RequestsPerMinute
	remaining := rl.config.RequestsPerMinute - clientInfo.RequestCount
	if remaining < 0 {
		remaining = 0
	}
	resetTime := now.Add(rl.config.WindowSize)

	return allowed, remaining, resetTime
}

// blockClient blocks a client for the configured duration
func (rl *RateLimiter) blockClient(ctx context.Context, clientIP string, clientInfo *ClientInfo) {
	now := time.Now()
	blockExpiry := now.Add(rl.config.BlockDuration)
	clientInfo.IsBlocked = true
	clientInfo.BlockExpiry = blockExpiry

	// Block in Redis
	if rl.config.EnableRedis && rl.redisClient != nil {
		blockKey := fmt.Sprintf("block:%s", clientIP)
		rl.redisClient.Set(ctx, blockKey, "1", rl.config.BlockDuration)
	}

	// Update local cache
	if rl.config.EnableLocalCache {
		rl.localCache.Store(clientIP, clientInfo)
	}

	rl.logger.Warn("Client blocked due to rate limit violations",
		zap.String("client_ip", clientIP),
		zap.Duration("block_duration", rl.config.BlockDuration),
		zap.Time("block_expiry", blockExpiry))
}

// updateClientInfo updates client information in storage
func (rl *RateLimiter) updateClientInfo(ctx context.Context, clientIP string, clientInfo *ClientInfo) {
	// Update Redis
	if rl.config.EnableRedis && rl.redisClient != nil {
		clientKey := fmt.Sprintf("client:%s", clientIP)
		rl.redisClient.HMSet(ctx, clientKey, map[string]interface{}{
			"user_agent":    clientInfo.UserAgent,
			"country":       clientInfo.Country,
			"first_seen":    clientInfo.FirstSeen.Format(time.RFC3339),
			"last_seen":     clientInfo.LastSeen.Format(time.RFC3339),
			"request_count": clientInfo.RequestCount,
			"block_count":   clientInfo.BlockCount,
			"is_blocked":    clientInfo.IsBlocked,
			"block_expiry":  clientInfo.BlockExpiry.Format(time.RFC3339),
		})
		rl.redisClient.Expire(ctx, clientKey, 24*time.Hour)
	}

	// Update local cache
	if rl.config.EnableLocalCache {
		rl.localCache.Store(clientIP, clientInfo)
	}
}

// GetClientStats returns statistics for a client
func (rl *RateLimiter) GetClientStats(ctx context.Context, clientIP string) *ClientInfo {
	return rl.getClientInfo(ctx, clientIP, "")
}

// UnblockClient manually unblocks a client
func (rl *RateLimiter) UnblockClient(ctx context.Context, clientIP string) error {
	// Remove from Redis
	if rl.config.EnableRedis && rl.redisClient != nil {
		blockKey := fmt.Sprintf("block:%s", clientIP)
		rl.redisClient.Del(ctx, blockKey)
	}

	// Update local cache
	if rl.config.EnableLocalCache {
		if info, ok := rl.localCache.Load(clientIP); ok {
			clientInfo := info.(*ClientInfo)
			clientInfo.IsBlocked = false
			clientInfo.BlockExpiry = time.Time{}
			rl.localCache.Store(clientIP, clientInfo)
		}
	}

	rl.logger.Info("Client manually unblocked", zap.String("client_ip", clientIP))
	return nil
}

// GetRateLimitConfig returns the current rate limit configuration
func (rl *RateLimiter) GetRateLimitConfig() RateLimitConfig {
	return rl.config
}

// UpdateRateLimitConfig updates the rate limit configuration
func (rl *RateLimiter) UpdateRateLimitConfig(config RateLimitConfig) {
	rl.config = config
	rl.logger.Info("Rate limit configuration updated", zap.Any("config", config))
}

// Helper functions for parsing stored values
func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func parseInt(s string) int {
	if s == "" {
		return 0
	}
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}

func parseBool(s string) bool {
	return s == "true"
}
