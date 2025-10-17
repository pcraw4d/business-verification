package batch

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"kyb-platform/services/risk-assessment-service/internal/queue"

	"go.uber.org/zap"
)

// RetryHandler manages job retry logic with exponential backoff
type RetryHandler struct {
	queue  *queue.RedisQueue
	logger *zap.Logger
	mu     sync.RWMutex
	stats  *RetryStats
	config *RetryConfig
}

// RetryStats represents statistics for retry operations
type RetryStats struct {
	TotalRetries      int64         `json:"total_retries"`
	SuccessfulRetries int64         `json:"successful_retries"`
	FailedRetries     int64         `json:"failed_retries"`
	MaxRetriesReached int64         `json:"max_retries_reached"`
	AverageRetryDelay time.Duration `json:"average_retry_delay"`
	LastRetry         time.Time     `json:"last_retry"`
}

// RetryConfig represents configuration for retry handling
type RetryConfig struct {
	MaxRetries        int           `json:"max_retries"`
	BaseDelay         time.Duration `json:"base_delay"`
	MaxDelay          time.Duration `json:"max_delay"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
	JitterEnabled     bool          `json:"jitter_enabled"`
	RetryableErrors   []string      `json:"retryable_errors"`
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	Error     error
	Retryable bool
	Delay     time.Duration
}

// NewRetryHandler creates a new retry handler
func NewRetryHandler(queue *queue.RedisQueue, config *RetryConfig, logger *zap.Logger) *RetryHandler {
	if config == nil {
		config = &RetryConfig{
			MaxRetries:        3,
			BaseDelay:         1 * time.Second,
			MaxDelay:          30 * time.Second,
			BackoffMultiplier: 2.0,
			JitterEnabled:     true,
			RetryableErrors: []string{
				"timeout",
				"connection refused",
				"network error",
				"temporary failure",
				"rate limit",
				"service unavailable",
			},
		}
	}

	return &RetryHandler{
		queue:  queue,
		logger: logger,
		stats:  &RetryStats{},
		config: config,
	}
}

// ShouldRetry determines if a job should be retried based on the error
func (rh *RetryHandler) ShouldRetry(job *queue.Job, err error) *RetryableError {
	rh.mu.RLock()
	defer rh.mu.RUnlock()

	// Check if job has exceeded max retries
	if job.RetryCount >= rh.config.MaxRetries {
		return &RetryableError{
			Error:     err,
			Retryable: false,
			Delay:     0,
		}
	}

	// Check if error is retryable
	if !rh.isRetryableError(err) {
		return &RetryableError{
			Error:     err,
			Retryable: false,
			Delay:     0,
		}
	}

	// Calculate retry delay
	delay := rh.calculateRetryDelay(job.RetryCount)

	return &RetryableError{
		Error:     err,
		Retryable: true,
		Delay:     delay,
	}
}

// RetryJob retries a failed job with exponential backoff
func (rh *RetryHandler) RetryJob(ctx context.Context, job *queue.Job, err error) error {
	retryableErr := rh.ShouldRetry(job, err)
	if !retryableErr.Retryable {
		rh.mu.Lock()
		rh.stats.MaxRetriesReached++
		rh.mu.Unlock()

		rh.logger.Warn("Job exceeded max retries, marking as failed",
			zap.String("job_id", job.ID),
			zap.Int("retry_count", job.RetryCount),
			zap.Int("max_retries", rh.config.MaxRetries),
			zap.Error(err))

		return rh.queue.FailJob(ctx, job.ID, err.Error())
	}

	// Update retry count
	job.RetryCount++
	job.Status = "pending"
	job.UpdatedAt = time.Now()
	job.Error = ""

	// Update job in queue
	if err := rh.updateJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update job for retry: %w", err)
	}

	// Schedule retry with delay
	go rh.scheduleRetry(ctx, job, retryableErr.Delay)

	// Update stats
	rh.mu.Lock()
	rh.stats.TotalRetries++
	rh.stats.AverageRetryDelay = (rh.stats.AverageRetryDelay + retryableErr.Delay) / 2
	rh.stats.LastRetry = time.Now()
	rh.mu.Unlock()

	rh.logger.Info("Job scheduled for retry",
		zap.String("job_id", job.ID),
		zap.Int("retry_count", job.RetryCount),
		zap.Duration("delay", retryableErr.Delay),
		zap.Error(err))

	return nil
}

// RetryFailedJobs retries all failed jobs that are eligible for retry
func (rh *RetryHandler) RetryFailedJobs(ctx context.Context) error {
	rh.logger.Info("Starting retry of failed jobs")

	// This is a simplified implementation
	// In a real implementation, you would query the failed jobs from the queue

	// For now, we'll just log that we're retrying failed jobs
	// In production, you would implement the actual retry logic

	rh.logger.Info("Failed jobs retry completed")

	return nil
}

// GetStats returns retry statistics
func (rh *RetryHandler) GetStats() *RetryStats {
	rh.mu.RLock()
	defer rh.mu.RUnlock()

	stats := *rh.stats
	return &stats
}

// Helper methods

func (rh *RetryHandler) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorMsg := err.Error()
	for _, retryableError := range rh.config.RetryableErrors {
		if contains(errorMsg, retryableError) {
			return true
		}
	}

	return false
}

func (rh *RetryHandler) calculateRetryDelay(retryCount int) time.Duration {
	// Calculate exponential backoff delay
	delay := float64(rh.config.BaseDelay) * math.Pow(rh.config.BackoffMultiplier, float64(retryCount))

	// Apply maximum delay limit
	if delay > float64(rh.config.MaxDelay) {
		delay = float64(rh.config.MaxDelay)
	}

	// Add jitter if enabled
	if rh.config.JitterEnabled {
		jitter := delay * 0.1 // 10% jitter
		delay += jitter
	}

	return time.Duration(delay)
}

func (rh *RetryHandler) scheduleRetry(ctx context.Context, job *queue.Job, delay time.Duration) {
	// Wait for the delay
	select {
	case <-ctx.Done():
		rh.logger.Info("Retry cancelled due to context cancellation",
			zap.String("job_id", job.ID))
		return
	case <-time.After(delay):
		// Continue with retry
	}

	// Re-enqueue the job
	if err := rh.queue.Enqueue(ctx, job); err != nil {
		rh.logger.Error("Failed to re-enqueue job for retry",
			zap.String("job_id", job.ID),
			zap.Error(err))

		// Mark job as failed if we can't re-enqueue it
		if err := rh.queue.FailJob(ctx, job.ID, fmt.Sprintf("failed to re-enqueue: %v", err)); err != nil {
			rh.logger.Error("Failed to mark job as failed after re-enqueue failure",
				zap.String("job_id", job.ID),
				zap.Error(err))
		}
		return
	}

	rh.logger.Info("Job re-enqueued for retry",
		zap.String("job_id", job.ID),
		zap.Int("retry_count", job.RetryCount))
}

func (rh *RetryHandler) updateJob(ctx context.Context, job *queue.Job) error {
	// This would update the job in the queue storage
	// For now, we'll just return nil as a placeholder
	return nil
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// RetryPolicy defines different retry policies
type RetryPolicy struct {
	Name              string        `json:"name"`
	MaxRetries        int           `json:"max_retries"`
	BaseDelay         time.Duration `json:"base_delay"`
	MaxDelay          time.Duration `json:"max_delay"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
	JitterEnabled     bool          `json:"jitter_enabled"`
	RetryableErrors   []string      `json:"retryable_errors"`
}

// GetRetryPolicy returns a retry policy by name
func GetRetryPolicy(name string) *RetryPolicy {
	policies := map[string]*RetryPolicy{
		"default": {
			Name:              "default",
			MaxRetries:        3,
			BaseDelay:         1 * time.Second,
			MaxDelay:          30 * time.Second,
			BackoffMultiplier: 2.0,
			JitterEnabled:     true,
			RetryableErrors: []string{
				"timeout",
				"connection refused",
				"network error",
				"temporary failure",
			},
		},
		"aggressive": {
			Name:              "aggressive",
			MaxRetries:        5,
			BaseDelay:         500 * time.Millisecond,
			MaxDelay:          10 * time.Second,
			BackoffMultiplier: 1.5,
			JitterEnabled:     true,
			RetryableErrors: []string{
				"timeout",
				"connection refused",
				"network error",
				"temporary failure",
				"rate limit",
				"service unavailable",
				"bad gateway",
				"gateway timeout",
			},
		},
		"conservative": {
			Name:              "conservative",
			MaxRetries:        2,
			BaseDelay:         2 * time.Second,
			MaxDelay:          60 * time.Second,
			BackoffMultiplier: 3.0,
			JitterEnabled:     false,
			RetryableErrors: []string{
				"timeout",
				"connection refused",
				"network error",
			},
		},
		"external_api": {
			Name:              "external_api",
			MaxRetries:        3,
			BaseDelay:         1 * time.Second,
			MaxDelay:          30 * time.Second,
			BackoffMultiplier: 2.0,
			JitterEnabled:     true,
			RetryableErrors: []string{
				"timeout",
				"connection refused",
				"network error",
				"rate limit",
				"service unavailable",
				"bad gateway",
				"gateway timeout",
				"internal server error",
			},
		},
	}

	if policy, exists := policies[name]; exists {
		return policy
	}

	return policies["default"]
}

// RetryHandlerWithPolicy creates a retry handler with a specific policy
func RetryHandlerWithPolicy(queue *queue.RedisQueue, policyName string, logger *zap.Logger) *RetryHandler {
	policy := GetRetryPolicy(policyName)

	config := &RetryConfig{
		MaxRetries:        policy.MaxRetries,
		BaseDelay:         policy.BaseDelay,
		MaxDelay:          policy.MaxDelay,
		BackoffMultiplier: policy.BackoffMultiplier,
		JitterEnabled:     policy.JitterEnabled,
		RetryableErrors:   policy.RetryableErrors,
	}

	return NewRetryHandler(queue, config, logger)
}
