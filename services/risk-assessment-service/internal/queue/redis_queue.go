package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisQueue implements a Redis-based job queue
type RedisQueue struct {
	client    *redis.Client
	logger    *zap.Logger
	keyPrefix string
	stats     *QueueStats
}

// QueueStats represents statistics for the queue
type QueueStats struct {
	TotalJobs     int64 `json:"total_jobs"`
	ProcessedJobs int64 `json:"processed_jobs"`
	FailedJobs    int64 `json:"failed_jobs"`
	RetriedJobs   int64 `json:"retried_jobs"`
	ActiveJobs    int64 `json:"active_jobs"`
	PendingJobs   int64 `json:"pending_jobs"`
	CompletedJobs int64 `json:"completed_jobs"`
}

// Job represents a job in the queue
type Job struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Data        map[string]interface{} `json:"data"`
	Priority    int                    `json:"priority"`
	MaxRetries  int                    `json:"max_retries"`
	RetryCount  int                    `json:"retry_count"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
}

// QueueConfig represents configuration for the Redis queue
type QueueConfig struct {
	KeyPrefix    string        `json:"key_prefix"`
	DefaultTTL   time.Duration `json:"default_ttl"`
	MaxRetries   int           `json:"max_retries"`
	RetryDelay   time.Duration `json:"retry_delay"`
	BatchSize    int           `json:"batch_size"`
	PollInterval time.Duration `json:"poll_interval"`
	EnableStats  bool          `json:"enable_stats"`
}

// NewRedisQueue creates a new Redis-based queue
func NewRedisQueue(client *redis.Client, config *QueueConfig, logger *zap.Logger) *RedisQueue {
	if config == nil {
		config = &QueueConfig{
			KeyPrefix:    "queue:",
			DefaultTTL:   24 * time.Hour,
			MaxRetries:   3,
			RetryDelay:   1 * time.Second,
			BatchSize:    10,
			PollInterval: 1 * time.Second,
			EnableStats:  true,
		}
	}

	return &RedisQueue{
		client:    client,
		logger:    logger,
		keyPrefix: config.KeyPrefix,
		stats:     &QueueStats{},
	}
}

// Enqueue adds a job to the queue
func (rq *RedisQueue) Enqueue(ctx context.Context, job *Job) error {
	if job == nil {
		return fmt.Errorf("job cannot be nil")
	}

	// Set default values
	if job.ID == "" {
		job.ID = generateJobID()
	}
	if job.Status == "" {
		job.Status = "pending"
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	job.UpdatedAt = time.Now()

	// Serialize job
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Add to pending queue
	pendingKey := rq.keyPrefix + "pending"
	if err := rq.client.LPush(ctx, pendingKey, data).Err(); err != nil {
		return fmt.Errorf("failed to add job to pending queue: %w", err)
	}

	// Store job details
	jobKey := rq.keyPrefix + "job:" + job.ID
	if err := rq.client.Set(ctx, jobKey, data, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store job details: %w", err)
	}

	// Update stats
	rq.stats.TotalJobs++
	rq.stats.PendingJobs++

	rq.logger.Info("Job enqueued",
		zap.String("job_id", job.ID),
		zap.String("job_type", job.Type),
		zap.Int("priority", job.Priority))

	return nil
}

// Dequeue retrieves a job from the queue
func (rq *RedisQueue) Dequeue(ctx context.Context) (*Job, error) {
	// Get job from pending queue
	pendingKey := rq.keyPrefix + "pending"
	result, err := rq.client.BRPop(ctx, 1*time.Second, pendingKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No jobs available
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid queue result")
	}

	// Deserialize job
	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Update job status
	job.Status = "processing"
	job.UpdatedAt = time.Now()

	// Update job in storage
	if err := rq.updateJob(ctx, &job); err != nil {
		rq.logger.Warn("Failed to update job status", zap.Error(err))
	}

	// Update stats
	rq.stats.PendingJobs--
	rq.stats.ActiveJobs++

	return &job, nil
}

// CompleteJob marks a job as completed
func (rq *RedisQueue) CompleteJob(ctx context.Context, jobID string, result map[string]interface{}) error {
	job, err := rq.getJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	job.Status = "completed"
	job.UpdatedAt = time.Now()
	now := time.Now()
	job.ProcessedAt = &now
	job.Result = result

	if err := rq.updateJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	// Move to completed queue
	completedKey := rq.keyPrefix + "completed"
	jobData, _ := json.Marshal(job)
	if err := rq.client.LPush(ctx, completedKey, jobData).Err(); err != nil {
		rq.logger.Warn("Failed to move job to completed queue", zap.Error(err))
	}

	// Update stats
	rq.stats.ActiveJobs--
	rq.stats.ProcessedJobs++
	rq.stats.CompletedJobs++

	rq.logger.Info("Job completed",
		zap.String("job_id", jobID),
		zap.String("job_type", job.Type))

	return nil
}

// FailJob marks a job as failed
func (rq *RedisQueue) FailJob(ctx context.Context, jobID string, errorMsg string) error {
	job, err := rq.getJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	job.Status = "failed"
	job.UpdatedAt = time.Now()
	job.Error = errorMsg

	if err := rq.updateJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	// Move to failed queue
	failedKey := rq.keyPrefix + "failed"
	jobData, _ := json.Marshal(job)
	if err := rq.client.LPush(ctx, failedKey, jobData).Err(); err != nil {
		rq.logger.Warn("Failed to move job to failed queue", zap.Error(err))
	}

	// Update stats
	rq.stats.ActiveJobs--
	rq.stats.FailedJobs++

	rq.logger.Info("Job failed",
		zap.String("job_id", jobID),
		zap.String("job_type", job.Type),
		zap.String("error", errorMsg))

	return nil
}

// RetryJob retries a failed job
func (rq *RedisQueue) RetryJob(ctx context.Context, jobID string) error {
	job, err := rq.getJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if job.RetryCount >= job.MaxRetries {
		return fmt.Errorf("job has exceeded maximum retries")
	}

	job.RetryCount++
	job.Status = "pending"
	job.UpdatedAt = time.Now()
	job.Error = ""

	if err := rq.updateJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	// Re-enqueue job
	jobData, _ := json.Marshal(job)
	pendingKey := rq.keyPrefix + "pending"
	if err := rq.client.LPush(ctx, pendingKey, jobData).Err(); err != nil {
		return fmt.Errorf("failed to re-enqueue job: %w", err)
	}

	// Update stats
	rq.stats.RetriedJobs++
	rq.stats.PendingJobs++

	rq.logger.Info("Job retried",
		zap.String("job_id", jobID),
		zap.String("job_type", job.Type),
		zap.Int("retry_count", job.RetryCount))

	return nil
}

// GetJob retrieves a job by ID
func (rq *RedisQueue) GetJob(ctx context.Context, jobID string) (*Job, error) {
	return rq.getJob(ctx, jobID)
}

// GetStats returns queue statistics
func (rq *RedisQueue) GetStats() *QueueStats {
	return rq.stats
}

// GetQueueLength returns the length of the pending queue
func (rq *RedisQueue) GetQueueLength(ctx context.Context) (int64, error) {
	pendingKey := rq.keyPrefix + "pending"
	return rq.client.LLen(ctx, pendingKey).Result()
}

// ClearQueue clears all jobs from the queue
func (rq *RedisQueue) ClearQueue(ctx context.Context) error {
	keys := []string{
		rq.keyPrefix + "pending",
		rq.keyPrefix + "completed",
		rq.keyPrefix + "failed",
	}

	for _, key := range keys {
		if err := rq.client.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("failed to clear queue %s: %w", key, err)
		}
	}

	// Reset stats
	rq.stats = &QueueStats{}

	rq.logger.Info("Queue cleared")
	return nil
}

// Helper methods

func (rq *RedisQueue) getJob(ctx context.Context, jobID string) (*Job, error) {
	jobKey := rq.keyPrefix + "job:" + jobID
	data, err := rq.client.Get(ctx, jobKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("job not found: %s", jobID)
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	var job Job
	if err := json.Unmarshal([]byte(data), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

func (rq *RedisQueue) updateJob(ctx context.Context, job *Job) error {
	jobKey := rq.keyPrefix + "job:" + job.ID
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	return rq.client.Set(ctx, jobKey, data, 24*time.Hour).Err()
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}
