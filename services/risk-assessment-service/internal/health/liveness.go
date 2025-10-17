package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// LivenessChecker checks if the service is alive and responsive
type LivenessChecker struct {
	logger *zap.Logger
	checks map[string]LivenessCheck
	mu     sync.RWMutex
}

// LivenessCheck represents a liveness check function
type LivenessCheck func(ctx context.Context) error

// LivenessStatus represents the liveness status of the service
type LivenessStatus struct {
	Alive     bool                   `json:"alive"`
	Checks    map[string]CheckResult `json:"checks"`
	Summary   string                 `json:"summary"`
	Timestamp time.Time              `json:"timestamp"`
	Uptime    time.Duration          `json:"uptime"`
}

// NewLivenessChecker creates a new liveness checker
func NewLivenessChecker(logger *zap.Logger) *LivenessChecker {
	checker := &LivenessChecker{
		logger: logger,
		checks: make(map[string]LivenessCheck),
	}

	// Register default liveness checks
	checker.registerDefaultChecks()

	return checker
}

// RegisterCheck registers a custom liveness check
func (lc *LivenessChecker) RegisterCheck(name string, check LivenessCheck) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.checks[name] = check
}

// CheckLiveness performs all registered liveness checks
func (lc *LivenessChecker) CheckLiveness(ctx context.Context) *LivenessStatus {
	lc.mu.RLock()
	checks := make(map[string]LivenessCheck)
	for name, check := range lc.checks {
		checks[name] = check
	}
	lc.mu.RUnlock()

	results := make(map[string]CheckResult)
	allAlive := true

	// Perform all checks
	for name, check := range checks {
		result := lc.performCheck(ctx, name, check)
		results[name] = result

		if result.Status != "healthy" {
			allAlive = false
		}
	}

	// Generate summary
	summary := lc.generateSummary(results)

	return &LivenessStatus{
		Alive:     allAlive,
		Checks:    results,
		Summary:   summary,
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime), // This would be set when the service starts
	}
}

// HTTPHandler returns an HTTP handler for liveness checks
func (lc *LivenessChecker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		status := lc.CheckLiveness(ctx)

		w.Header().Set("Content-Type", "application/json")

		if status.Alive {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		// Write JSON response
		if err := json.NewEncoder(w).Encode(status); err != nil {
			lc.logger.Error("Failed to encode liveness status", zap.Error(err))
		}
	}
}

// Helper methods

func (lc *LivenessChecker) registerDefaultChecks() {
	// Basic service health check
	lc.RegisterCheck("service", lc.checkService)

	// Memory usage check
	lc.RegisterCheck("memory", lc.checkMemory)

	// Goroutine count check
	lc.RegisterCheck("goroutines", lc.checkGoroutines)

	// File system check
	lc.RegisterCheck("filesystem", lc.checkFilesystem)
}

func (lc *LivenessChecker) performCheck(ctx context.Context, name string, check LivenessCheck) CheckResult {
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

		lc.logger.Warn("Liveness check failed",
			zap.String("check", name),
			zap.Error(err),
			zap.Duration("duration", duration))
	} else {
		result.Status = "healthy"
		result.Message = "Check passed"

		lc.logger.Debug("Liveness check passed",
			zap.String("check", name),
			zap.Duration("duration", duration))
	}

	return result
}

func (lc *LivenessChecker) checkService(ctx context.Context) error {
	// Basic service health check
	// This is a simple check that the service is responsive

	// For now, we'll just return success
	// In production, you might want to check:
	// - Service responsiveness
	// - Basic functionality
	// - Critical components

	return nil
}

func (lc *LivenessChecker) checkMemory(ctx context.Context) error {
	// Memory usage check
	// This is a placeholder implementation

	// In production, you might want to check:
	// - Memory usage is within limits
	// - No memory leaks
	// - Available memory is sufficient

	return nil
}

func (lc *LivenessChecker) checkGoroutines(ctx context.Context) error {
	// Goroutine count check
	// This is a placeholder implementation

	// In production, you might want to check:
	// - Goroutine count is within limits
	// - No goroutine leaks
	// - Goroutine count is not excessive

	return nil
}

func (lc *LivenessChecker) checkFilesystem(ctx context.Context) error {
	// File system check
	// This is a placeholder implementation

	// In production, you might want to check:
	// - File system is writable
	// - Disk space is available
	// - Required files exist

	return nil
}

func (lc *LivenessChecker) generateSummary(results map[string]CheckResult) string {
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

// Global variable to track service start time
var startTime = time.Now()
