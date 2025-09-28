package redisoptimization

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisOptimizer provides advanced Redis optimization capabilities
type RedisOptimizer struct {
	client *redis.Client
	config *OptimizationConfig
}

// OptimizationConfig contains Redis optimization settings
type OptimizationConfig struct {
	// Connection Pool Settings
	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolSize        int
	MinIdleConns    int
	MaxIdleConns    int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration

	// Cache Strategy Settings
	DefaultTTL        time.Duration
	ClassificationTTL time.Duration
	AnalyticsTTL      time.Duration
	MetricsTTL        time.Duration
	HealthTTL         time.Duration

	// Performance Settings
	EnableCompression bool
	CompressionLevel  int
	EnablePipelining  bool
	PipelineSize      int
}

// DefaultOptimizationConfig returns optimized Redis configuration
func DefaultOptimizationConfig() *OptimizationConfig {
	return &OptimizationConfig{
		// Connection Pool - Optimized for high throughput
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolSize:        100, // Increased for high concurrency
		MinIdleConns:    10,  // Keep connections ready
		MaxIdleConns:    50,  // Balance memory vs performance
		ConnMaxIdleTime: 5 * time.Minute,
		ConnMaxLifetime: 30 * time.Minute,

		// Cache Strategy - Optimized TTL values
		DefaultTTL:        1 * time.Hour,    // General cache
		ClassificationTTL: 24 * time.Hour,   // Business classifications rarely change
		AnalyticsTTL:      5 * time.Minute,  // Analytics data changes frequently
		MetricsTTL:        1 * time.Minute,  // Metrics need to be fresh
		HealthTTL:         30 * time.Second, // Health checks need to be very fresh

		// Performance Settings
		EnableCompression: true,
		CompressionLevel:  6, // Balanced compression
		EnablePipelining:  true,
		PipelineSize:      100,
	}
}

// NewRedisOptimizer creates a new optimized Redis client
func NewRedisOptimizer(addr, password string, config *OptimizationConfig) *RedisOptimizer {
	if config == nil {
		config = DefaultOptimizationConfig()
	}

	client := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        password,
		DB:              0,
		MaxRetries:      config.MaxRetries,
		MinRetryBackoff: config.MinRetryBackoff,
		MaxRetryBackoff: config.MaxRetryBackoff,
		DialTimeout:     config.DialTimeout,
		ReadTimeout:     config.ReadTimeout,
		WriteTimeout:    config.WriteTimeout,
		PoolSize:        config.PoolSize,
		MinIdleConns:    config.MinIdleConns,
		MaxIdleConns:    config.MaxIdleConns,
		ConnMaxIdleTime: config.ConnMaxIdleTime,
		ConnMaxLifetime: config.ConnMaxLifetime,
	})

	return &RedisOptimizer{
		client: client,
		config: config,
	}
}

// GetClient returns the optimized Redis client
func (ro *RedisOptimizer) GetClient() *redis.Client {
	return ro.client
}

// OptimizeCacheStrategy applies intelligent caching based on data type
func (ro *RedisOptimizer) OptimizeCacheStrategy(ctx context.Context, key string, data interface{}, dataType string) error {
	var ttl time.Duration

	switch dataType {
	case "classification":
		ttl = ro.config.ClassificationTTL
	case "analytics":
		ttl = ro.config.AnalyticsTTL
	case "metrics":
		ttl = ro.config.MetricsTTL
	case "health":
		ttl = ro.config.HealthTTL
	default:
		ttl = ro.config.DefaultTTL
	}

	// Use pipeline for better performance if enabled
	if ro.config.EnablePipelining {
		pipe := ro.client.Pipeline()
		pipe.Set(ctx, key, data, ttl)
		pipe.Expire(ctx, key, ttl)
		_, err := pipe.Exec(ctx)
		return err
	}

	return ro.client.Set(ctx, key, data, ttl).Err()
}

// BatchOperations performs multiple Redis operations efficiently
func (ro *RedisOptimizer) BatchOperations(ctx context.Context, operations []RedisOperation) error {
	if !ro.config.EnablePipelining {
		// Fallback to individual operations
		for _, op := range operations {
			if err := ro.executeOperation(ctx, op); err != nil {
				return err
			}
		}
		return nil
	}

	// Use pipeline for batch operations
	pipe := ro.client.Pipeline()

	for _, op := range operations {
		switch op.Type {
		case "SET":
			pipe.Set(ctx, op.Key, op.Value, op.TTL)
		case "GET":
			pipe.Get(ctx, op.Key)
		case "DEL":
			pipe.Del(ctx, op.Key)
		case "EXPIRE":
			pipe.Expire(ctx, op.Key, op.TTL)
		}
	}

	_, err := pipe.Exec(ctx)
	return err
}

// RedisOperation represents a single Redis operation
type RedisOperation struct {
	Type  string
	Key   string
	Value interface{}
	TTL   time.Duration
}

func (ro *RedisOptimizer) executeOperation(ctx context.Context, op RedisOperation) error {
	switch op.Type {
	case "SET":
		return ro.client.Set(ctx, op.Key, op.Value, op.TTL).Err()
	case "GET":
		_, err := ro.client.Get(ctx, op.Key).Result()
		return err
	case "DEL":
		return ro.client.Del(ctx, op.Key).Err()
	case "EXPIRE":
		return ro.client.Expire(ctx, op.Key, op.TTL).Err()
	default:
		return fmt.Errorf("unknown operation type: %s", op.Type)
	}
}

// GetCacheStats returns Redis performance statistics
func (ro *RedisOptimizer) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	_, err := ro.client.Info(ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	// Parse Redis info to extract key metrics
	stats := &CacheStats{
		Timestamp: time.Now(),
	}

	// This is a simplified version - in production, you'd parse the full Redis INFO output
	poolStats := ro.client.PoolStats()
	stats.TotalConnections = int(poolStats.TotalConns)
	stats.ActiveConnections = int(poolStats.TotalConns - poolStats.IdleConns)
	stats.IdleConnections = int(poolStats.IdleConns)

	return stats, nil
}

// CacheStats contains Redis performance metrics
type CacheStats struct {
	Timestamp         time.Time
	TotalConnections  int
	ActiveConnections int
	IdleConnections   int
	HitRate           float64
	MissRate          float64
	MemoryUsage       int64
	KeyCount          int64
}

// WarmupCache preloads frequently accessed data
func (ro *RedisOptimizer) WarmupCache(ctx context.Context, warmupData map[string]interface{}) error {
	if !ro.config.EnablePipelining {
		// Sequential warmup
		for key, value := range warmupData {
			if err := ro.client.Set(ctx, key, value, ro.config.DefaultTTL).Err(); err != nil {
				log.Printf("Warning: Failed to warmup key %s: %v", key, err)
			}
		}
		return nil
	}

	// Pipeline warmup for better performance
	pipe := ro.client.Pipeline()
	for key, value := range warmupData {
		pipe.Set(ctx, key, value, ro.config.DefaultTTL)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// HealthCheck performs Redis health check with optimization metrics
func (ro *RedisOptimizer) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	start := time.Now()

	// Test basic connectivity
	if err := ro.client.Ping(ctx).Err(); err != nil {
		return &HealthStatus{
			Status:    "unhealthy",
			Latency:   time.Since(start),
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	// Test write/read performance
	testKey := "health_check_test"
	testValue := fmt.Sprintf("test_%d", time.Now().Unix())

	writeStart := time.Now()
	if err := ro.client.Set(ctx, testKey, testValue, 10*time.Second).Err(); err != nil {
		return &HealthStatus{
			Status:    "unhealthy",
			Latency:   time.Since(start),
			Error:     fmt.Sprintf("write test failed: %v", err),
			Timestamp: time.Now(),
		}, nil
	}
	writeLatency := time.Since(writeStart)

	readStart := time.Now()
	_, readErr := ro.client.Get(ctx, testKey).Result()
	if readErr != nil {
		return &HealthStatus{
			Status:    "unhealthy",
			Latency:   time.Since(start),
			Error:     fmt.Sprintf("read test failed: %v", readErr),
			Timestamp: time.Now(),
		}, nil
	}
	readLatency := time.Since(readStart)

	// Cleanup test key
	ro.client.Del(ctx, testKey)

	// Get connection pool stats
	poolStats := ro.client.PoolStats()

	return &HealthStatus{
		Status:            "healthy",
		Latency:           time.Since(start),
		WriteLatency:      writeLatency,
		ReadLatency:       readLatency,
		TotalConnections:  int(poolStats.TotalConns),
		ActiveConnections: int(poolStats.TotalConns - poolStats.IdleConns),
		IdleConnections:   int(poolStats.IdleConns),
		Timestamp:         time.Now(),
	}, nil
}

// HealthStatus contains Redis health information
type HealthStatus struct {
	Status            string        `json:"status"`
	Latency           time.Duration `json:"latency"`
	WriteLatency      time.Duration `json:"write_latency"`
	ReadLatency       time.Duration `json:"read_latency"`
	TotalConnections  int           `json:"total_connections"`
	ActiveConnections int           `json:"active_connections"`
	IdleConnections   int           `json:"idle_connections"`
	Error             string        `json:"error,omitempty"`
	Timestamp         time.Time     `json:"timestamp"`
}

// Close properly closes the Redis connection
func (ro *RedisOptimizer) Close() error {
	return ro.client.Close()
}
