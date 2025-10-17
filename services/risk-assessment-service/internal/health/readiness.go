package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ReadinessChecker checks if the service is ready to serve requests
type ReadinessChecker struct {
	db          *sql.DB
	redisClient *redis.Client
	logger      *zap.Logger
	checks      map[string]HealthCheck
	mu          sync.RWMutex
}

// HealthCheck represents a health check function
type HealthCheck func(ctx context.Context) error

// ReadinessStatus represents the readiness status of the service
type ReadinessStatus struct {
	Ready     bool                   `json:"ready"`
	Checks    map[string]CheckResult `json:"checks"`
	Summary   string                 `json:"summary"`
	Timestamp time.Time              `json:"timestamp"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Status    string        `json:"status"`
	Message   string        `json:"message"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// NewReadinessChecker creates a new readiness checker
func NewReadinessChecker(db *sql.DB, redisClient *redis.Client, logger *zap.Logger) *ReadinessChecker {
	checker := &ReadinessChecker{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
		checks:      make(map[string]HealthCheck),
	}

	// Register default health checks
	checker.registerDefaultChecks()

	return checker
}

// RegisterCheck registers a custom health check
func (rc *ReadinessChecker) RegisterCheck(name string, check HealthCheck) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.checks[name] = check
}

// CheckReadiness performs all registered health checks
func (rc *ReadinessChecker) CheckReadiness(ctx context.Context) *ReadinessStatus {
	rc.mu.RLock()
	checks := make(map[string]HealthCheck)
	for name, check := range rc.checks {
		checks[name] = check
	}
	rc.mu.RUnlock()

	results := make(map[string]CheckResult)
	allReady := true

	// Perform all checks
	for name, check := range checks {
		result := rc.performCheck(ctx, name, check)
		results[name] = result

		if result.Status != "healthy" {
			allReady = false
		}
	}

	// Generate summary
	summary := rc.generateSummary(results)

	return &ReadinessStatus{
		Ready:     allReady,
		Checks:    results,
		Summary:   summary,
		Timestamp: time.Now(),
	}
}

// HTTPHandler returns an HTTP handler for readiness checks
func (rc *ReadinessChecker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		status := rc.CheckReadiness(ctx)

		w.Header().Set("Content-Type", "application/json")

		if status.Ready {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		// Write JSON response
		if err := json.NewEncoder(w).Encode(status); err != nil {
			rc.logger.Error("Failed to encode readiness status", zap.Error(err))
		}
	}
}

// Helper methods

func (rc *ReadinessChecker) registerDefaultChecks() {
	// Database connectivity check
	rc.RegisterCheck("database", rc.checkDatabase)

	// Redis connectivity check
	rc.RegisterCheck("redis", rc.checkRedis)

	// ML model loaded check
	rc.RegisterCheck("ml_models", rc.checkMLModels)

	// External API availability check
	rc.RegisterCheck("external_apis", rc.checkExternalAPIs)

	// Cache health check
	rc.RegisterCheck("cache", rc.checkCache)

	// Configuration check
	rc.RegisterCheck("configuration", rc.checkConfiguration)
}

func (rc *ReadinessChecker) performCheck(ctx context.Context, name string, check HealthCheck) CheckResult {
	start := time.Now()

	err := check(ctx)
	duration := time.Since(start)

	result := CheckResult{
		Duration:  duration,
		Timestamp: time.Now(),
	}

	if err != nil {
		result.Status = "unhealthy"
		result.Message = fmt.Sprintf("Check failed: %v", err)
		result.Error = err.Error()

		rc.logger.Warn("Health check failed",
			zap.String("check", name),
			zap.Error(err),
			zap.Duration("duration", duration))
	} else {
		result.Status = "healthy"
		result.Message = "Check passed"

		rc.logger.Debug("Health check passed",
			zap.String("check", name),
			zap.Duration("duration", duration))
	}

	return result
}

func (rc *ReadinessChecker) checkDatabase(ctx context.Context) error {
	if rc.db == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Test database connectivity
	if err := rc.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Test a simple query
	var result int
	if err := rc.db.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected query result: %d", result)
	}

	return nil
}

func (rc *ReadinessChecker) checkRedis(ctx context.Context) error {
	if rc.redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	// Test Redis connectivity
	if err := rc.redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	// Test Redis operations
	testKey := "health_check_test"
	testValue := "test_value"

	if err := rc.redisClient.Set(ctx, testKey, testValue, time.Second).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}

	val, err := rc.redisClient.Get(ctx, testKey).Result()
	if err != nil {
		return fmt.Errorf("redis get failed: %w", err)
	}

	if val != testValue {
		return fmt.Errorf("unexpected redis value: %s", val)
	}

	// Clean up test key
	rc.redisClient.Del(ctx, testKey)

	return nil
}

func (rc *ReadinessChecker) checkMLModels(ctx context.Context) error {
	// This is a placeholder implementation
	// In a real implementation, you would check if ML models are loaded and ready

	// Simulate model loading check
	time.Sleep(100 * time.Millisecond)

	// Check if models directory exists and contains model files
	// This would be implemented based on your ML model loading logic

	return nil
}

func (rc *ReadinessChecker) checkExternalAPIs(ctx context.Context) error {
	// This is a placeholder implementation
	// In a real implementation, you would check if external APIs are reachable

	// For now, we'll just return success
	// In production, you might want to check:
	// - Thomson Reuters API
	// - OFAC API
	// - News API
	// - OpenCorporates API

	return nil
}

func (rc *ReadinessChecker) checkCache(ctx context.Context) error {
	// This is a placeholder implementation
	// In a real implementation, you would check cache health

	// For now, we'll just return success
	// In production, you might want to check:
	// - Cache hit rates
	// - Cache memory usage
	// - Cache connectivity

	return nil
}

func (rc *ReadinessChecker) checkConfiguration(ctx context.Context) error {
	// This is a placeholder implementation
	// In a real implementation, you would check if all required configuration is present

	// For now, we'll just return success
	// In production, you might want to check:
	// - Required environment variables
	// - Configuration file validity
	// - API keys and secrets

	return nil
}

func (rc *ReadinessChecker) generateSummary(results map[string]CheckResult) string {
	healthy := 0
	total := len(results)

	for _, result := range results {
		if result.Status == "healthy" {
			healthy++
		}
	}

	if healthy == total {
		return fmt.Sprintf("All %d checks passed", total)
	} else if healthy == 0 {
		return fmt.Sprintf("All %d checks failed", total)
	} else {
		return fmt.Sprintf("%d of %d checks passed", healthy, total)
	}
}
