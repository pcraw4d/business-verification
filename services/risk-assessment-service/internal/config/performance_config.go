package config

import (
	"fmt"
	"time"
)

// PerformanceConfig represents performance optimization configuration
type PerformanceConfig struct {
	// Cache Configuration
	Cache CacheConfig `json:"cache" yaml:"cache"`

	// Connection Pool Configuration
	ConnectionPool ConnectionPoolConfig `json:"connection_pool" yaml:"connection_pool"`

	// Query Optimization Configuration
	QueryOptimization QueryOptimizationConfig `json:"query_optimization" yaml:"query_optimization"`

	// Performance Monitoring Configuration
	Monitoring PerformanceMonitoringConfig `json:"monitoring" yaml:"monitoring"`

	// Rate Limiting Configuration
	RateLimiting RateLimitingConfig `json:"rate_limiting" yaml:"rate_limiting"`

	// Resource Limits Configuration
	ResourceLimits ResourceLimitsConfig `json:"resource_limits" yaml:"resource_limits"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Enabled           bool          `json:"enabled" yaml:"enabled"`
	Type              string        `json:"type" yaml:"type"` // redis, memory
	DefaultTTL        time.Duration `json:"default_ttl" yaml:"default_ttl"`
	MaxSize           int           `json:"max_size" yaml:"max_size"`
	CleanupInterval   time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
	EnableMetrics     bool          `json:"enable_metrics" yaml:"enable_metrics"`
	EnableCompression bool          `json:"enable_compression" yaml:"enable_compression"`

	// Redis specific configuration
	Redis RedisCacheConfig `json:"redis" yaml:"redis"`
}

// RedisCacheConfig represents Redis cache configuration
type RedisCacheConfig struct {
	Addrs         []string      `json:"addrs" yaml:"addrs"`
	Password      string        `json:"password" yaml:"password"`
	DB            int           `json:"db" yaml:"db"`
	PoolSize      int           `json:"pool_size" yaml:"pool_size"`
	MinIdleConns  int           `json:"min_idle_conns" yaml:"min_idle_conns"`
	MaxRetries    int           `json:"max_retries" yaml:"max_retries"`
	DialTimeout   time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout   time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout  time.Duration `json:"write_timeout" yaml:"write_timeout"`
	PoolTimeout   time.Duration `json:"pool_timeout" yaml:"pool_timeout"`
	IdleTimeout   time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	IdleCheckFreq time.Duration `json:"idle_check_freq" yaml:"idle_check_freq"`
	MaxConnAge    time.Duration `json:"max_conn_age" yaml:"max_conn_age"`
	KeyPrefix     string        `json:"key_prefix" yaml:"key_prefix"`
}

// ConnectionPoolConfig represents connection pool configuration
type ConnectionPoolConfig struct {
	MaxConnections     int           `json:"max_connections" yaml:"max_connections"`
	MinConnections     int           `json:"min_connections" yaml:"min_connections"`
	MaxIdleConnections int           `json:"max_idle_connections" yaml:"max_idle_connections"`
	ConnectionTimeout  time.Duration `json:"connection_timeout" yaml:"connection_timeout"`
	IdleTimeout        time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	MaxLifetime        time.Duration `json:"max_lifetime" yaml:"max_lifetime"`
	HealthCheckPeriod  time.Duration `json:"health_check_period" yaml:"health_check_period"`
	RetryAttempts      int           `json:"retry_attempts" yaml:"retry_attempts"`
	RetryDelay         time.Duration `json:"retry_delay" yaml:"retry_delay"`
}

// QueryOptimizationConfig represents query optimization configuration
type QueryOptimizationConfig struct {
	EnableCaching        bool          `json:"enable_caching" yaml:"enable_caching"`
	CacheTTL             time.Duration `json:"cache_ttl" yaml:"cache_ttl"`
	EnablePreparedStmts  bool          `json:"enable_prepared_stmts" yaml:"enable_prepared_stmts"`
	EnableBatchQueries   bool          `json:"enable_batch_queries" yaml:"enable_batch_queries"`
	EnableQueryAnalysis  bool          `json:"enable_query_analysis" yaml:"enable_query_analysis"`
	SlowQueryThreshold   time.Duration `json:"slow_query_threshold" yaml:"slow_query_threshold"`
	MaxRetries           int           `json:"max_retries" yaml:"max_retries"`
	RetryDelay           time.Duration `json:"retry_delay" yaml:"retry_delay"`
	EnableIndexHints     bool          `json:"enable_index_hints" yaml:"enable_index_hints"`
	EnableQueryPlanCache bool          `json:"enable_query_plan_cache" yaml:"enable_query_plan_cache"`
}

// PerformanceMonitoringConfig represents performance monitoring configuration
type PerformanceMonitoringConfig struct {
	Enabled               bool          `json:"enabled" yaml:"enabled"`
	CollectionInterval    time.Duration `json:"collection_interval" yaml:"collection_interval"`
	EnableMetrics         bool          `json:"enable_metrics" yaml:"enable_metrics"`
	EnableHealthChecks    bool          `json:"enable_health_checks" yaml:"enable_health_checks"`
	EnableSlowQueryLog    bool          `json:"enable_slow_query_log" yaml:"enable_slow_query_log"`
	EnableMemoryProfiling bool          `json:"enable_memory_profiling" yaml:"enable_memory_profiling"`
	EnableCPUProfiling    bool          `json:"enable_cpu_profiling" yaml:"enable_cpu_profiling"`
	MetricsEndpoint       string        `json:"metrics_endpoint" yaml:"metrics_endpoint"`
	HealthEndpoint        string        `json:"health_endpoint" yaml:"health_endpoint"`
	RetentionPeriod       time.Duration `json:"retention_period" yaml:"retention_period"`
}

// RateLimitingConfig represents rate limiting configuration
type RateLimitingConfig struct {
	Enabled           bool          `json:"enabled" yaml:"enabled"`
	RequestsPerMinute int           `json:"requests_per_minute" yaml:"requests_per_minute"`
	Burst             int           `json:"burst" yaml:"burst"`
	SkipOnError       bool          `json:"skip_on_error" yaml:"skip_on_error"`
	Window            time.Duration `json:"window" yaml:"window"`
	EnablePerIP       bool          `json:"enable_per_ip" yaml:"enable_per_ip"`
	EnablePerUser     bool          `json:"enable_per_user" yaml:"enable_per_user"`
	EnablePerTenant   bool          `json:"enable_per_tenant" yaml:"enable_per_tenant"`
}

// ResourceLimitsConfig represents resource limits configuration
type ResourceLimitsConfig struct {
	MaxMemoryMB             int           `json:"max_memory_mb" yaml:"max_memory_mb"`
	MaxCPUPercent           float64       `json:"max_cpu_percent" yaml:"max_cpu_percent"`
	MaxConcurrentReqs       int           `json:"max_concurrent_reqs" yaml:"max_concurrent_reqs"`
	MaxRequestSizeMB        int           `json:"max_request_size_mb" yaml:"max_request_size_mb"`
	MaxResponseSizeMB       int           `json:"max_response_size_mb" yaml:"max_response_size_mb"`
	RequestTimeout          time.Duration `json:"request_timeout" yaml:"request_timeout"`
	ResponseTimeout         time.Duration `json:"response_timeout" yaml:"response_timeout"`
	EnableCircuitBreaker    bool          `json:"enable_circuit_breaker" yaml:"enable_circuit_breaker"`
	CircuitBreakerThreshold int           `json:"circuit_breaker_threshold" yaml:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration `json:"circuit_breaker_timeout" yaml:"circuit_breaker_timeout"`
}

// DefaultPerformanceConfig returns default performance configuration
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		Cache: CacheConfig{
			Enabled:           true,
			Type:              "redis",
			DefaultTTL:        5 * time.Minute,
			MaxSize:           1000,
			CleanupInterval:   10 * time.Minute,
			EnableMetrics:     true,
			EnableCompression: false,
			Redis: RedisCacheConfig{
				Addrs:         []string{"localhost:6379"},
				Password:      "",
				DB:            0,
				PoolSize:      10,
				MinIdleConns:  5,
				MaxRetries:    3,
				DialTimeout:   5 * time.Second,
				ReadTimeout:   3 * time.Second,
				WriteTimeout:  3 * time.Second,
				PoolTimeout:   4 * time.Second,
				IdleTimeout:   5 * time.Minute,
				IdleCheckFreq: 1 * time.Minute,
				MaxConnAge:    30 * time.Minute,
				KeyPrefix:     "risk_assessment:",
			},
		},
		ConnectionPool: ConnectionPoolConfig{
			MaxConnections:     25,
			MinConnections:     5,
			MaxIdleConnections: 5,
			ConnectionTimeout:  30 * time.Second,
			IdleTimeout:        1 * time.Minute,
			MaxLifetime:        5 * time.Minute,
			HealthCheckPeriod:  30 * time.Second,
			RetryAttempts:      3,
			RetryDelay:         1 * time.Second,
		},
		QueryOptimization: QueryOptimizationConfig{
			EnableCaching:        true,
			CacheTTL:             5 * time.Minute,
			EnablePreparedStmts:  true,
			EnableBatchQueries:   true,
			EnableQueryAnalysis:  true,
			SlowQueryThreshold:   1 * time.Second,
			MaxRetries:           3,
			RetryDelay:           1 * time.Second,
			EnableIndexHints:     true,
			EnableQueryPlanCache: true,
		},
		Monitoring: PerformanceMonitoringConfig{
			Enabled:               true,
			CollectionInterval:    30 * time.Second,
			EnableMetrics:         true,
			EnableHealthChecks:    true,
			EnableSlowQueryLog:    true,
			EnableMemoryProfiling: false,
			EnableCPUProfiling:    false,
			MetricsEndpoint:       "/metrics",
			HealthEndpoint:        "/health",
			RetentionPeriod:       24 * time.Hour,
		},
		RateLimiting: RateLimitingConfig{
			Enabled:           true,
			RequestsPerMinute: 1000,
			Burst:             100,
			SkipOnError:       false,
			Window:            1 * time.Minute,
			EnablePerIP:       true,
			EnablePerUser:     true,
			EnablePerTenant:   true,
		},
		ResourceLimits: ResourceLimitsConfig{
			MaxMemoryMB:             2048,
			MaxCPUPercent:           80.0,
			MaxConcurrentReqs:       10000,
			MaxRequestSizeMB:        10,
			MaxResponseSizeMB:       10,
			RequestTimeout:          30 * time.Second,
			ResponseTimeout:         30 * time.Second,
			EnableCircuitBreaker:    true,
			CircuitBreakerThreshold: 50,
			CircuitBreakerTimeout:   5 * time.Minute,
		},
	}
}

// Validate validates the performance configuration
func (pc *PerformanceConfig) Validate() error {
	// Validate cache configuration
	if pc.Cache.Enabled {
		if pc.Cache.DefaultTTL <= 0 {
			return fmt.Errorf("cache default TTL must be positive")
		}
		if pc.Cache.MaxSize <= 0 {
			return fmt.Errorf("cache max size must be positive")
		}
	}

	// Validate connection pool configuration
	if pc.ConnectionPool.MaxConnections <= 0 {
		return fmt.Errorf("max connections must be positive")
	}
	if pc.ConnectionPool.MinConnections < 0 {
		return fmt.Errorf("min connections cannot be negative")
	}
	if pc.ConnectionPool.MaxIdleConnections < 0 {
		return fmt.Errorf("max idle connections cannot be negative")
	}

	// Validate query optimization configuration
	if pc.QueryOptimization.SlowQueryThreshold <= 0 {
		return fmt.Errorf("slow query threshold must be positive")
	}
	if pc.QueryOptimization.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	// Validate monitoring configuration
	if pc.Monitoring.Enabled {
		if pc.Monitoring.CollectionInterval <= 0 {
			return fmt.Errorf("collection interval must be positive")
		}
	}

	// Validate rate limiting configuration
	if pc.RateLimiting.Enabled {
		if pc.RateLimiting.RequestsPerMinute <= 0 {
			return fmt.Errorf("requests per minute must be positive")
		}
		if pc.RateLimiting.Burst <= 0 {
			return fmt.Errorf("burst must be positive")
		}
	}

	// Validate resource limits configuration
	if pc.ResourceLimits.MaxMemoryMB <= 0 {
		return fmt.Errorf("max memory must be positive")
	}
	if pc.ResourceLimits.MaxCPUPercent <= 0 || pc.ResourceLimits.MaxCPUPercent > 100 {
		return fmt.Errorf("max CPU percent must be between 0 and 100")
	}
	if pc.ResourceLimits.MaxConcurrentReqs <= 0 {
		return fmt.Errorf("max concurrent requests must be positive")
	}

	return nil
}
