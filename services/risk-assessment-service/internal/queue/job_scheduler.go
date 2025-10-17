package queue

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// JobScheduler manages job scheduling and prioritization
type JobScheduler struct {
	queue  *RedisQueue
	logger *zap.Logger
	mu     sync.RWMutex
	stats  *SchedulerStats
}

// SchedulerStats represents statistics for the job scheduler
type SchedulerStats struct {
	ScheduledJobs      int64         `json:"scheduled_jobs"`
	ProcessedJobs      int64         `json:"processed_jobs"`
	FailedJobs         int64         `json:"failed_jobs"`
	RetriedJobs        int64         `json:"retried_jobs"`
	AverageWaitTime    time.Duration `json:"average_wait_time"`
	AverageProcessTime time.Duration `json:"average_process_time"`
	LastScheduled      time.Time     `json:"last_scheduled"`
}

// ScheduledJob represents a scheduled job
type ScheduledJob struct {
	Job      *Job      `json:"job"`
	Schedule time.Time `json:"schedule"`
	Priority int       `json:"priority"`
}

// SchedulerConfig represents configuration for the job scheduler
type SchedulerConfig struct {
	MaxConcurrentJobs int           `json:"max_concurrent_jobs"`
	DefaultPriority   int           `json:"default_priority"`
	RetryDelay        time.Duration `json:"retry_delay"`
	MaxRetries        int           `json:"max_retries"`
	EnableStats       bool          `json:"enable_stats"`
}

// NewJobScheduler creates a new job scheduler
func NewJobScheduler(queue *RedisQueue, config *SchedulerConfig, logger *zap.Logger) *JobScheduler {
	if config == nil {
		config = &SchedulerConfig{
			MaxConcurrentJobs: 10,
			DefaultPriority:   5,
			RetryDelay:        1 * time.Second,
			MaxRetries:        3,
			EnableStats:       true,
		}
	}

	return &JobScheduler{
		queue:  queue,
		logger: logger,
		stats:  &SchedulerStats{},
	}
}

// ScheduleJob schedules a job for execution
func (js *JobScheduler) ScheduleJob(ctx context.Context, job *Job, schedule time.Time) error {
	if job == nil {
		return fmt.Errorf("job cannot be nil")
	}

	// Set default values
	if job.Priority == 0 {
		job.Priority = 5
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}

	// If schedule is in the past, enqueue immediately
	if schedule.Before(time.Now()) {
		return js.queue.Enqueue(ctx, job)
	}

	// Store scheduled job (in a real implementation, you might use a different storage)
	// For now, we'll just enqueue it
	return js.queue.Enqueue(ctx, job)
}

// ScheduleJobWithPriority schedules a job with a specific priority
func (js *JobScheduler) ScheduleJobWithPriority(ctx context.Context, job *Job, priority int) error {
	if job == nil {
		return fmt.Errorf("job cannot be nil")
	}

	job.Priority = priority
	return js.queue.Enqueue(ctx, job)
}

// ScheduleRecurringJob schedules a recurring job
func (js *JobScheduler) ScheduleRecurringJob(ctx context.Context, job *Job, interval time.Duration) error {
	if job == nil {
		return fmt.Errorf("job cannot be nil")
	}

	// Create a recurring job
	recurringJob := &Job{
		ID:         job.ID + "_recurring",
		Type:       "recurring",
		Data:       job.Data,
		Priority:   job.Priority,
		MaxRetries: job.MaxRetries,
		Status:     "pending",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Schedule the first execution
	if err := js.queue.Enqueue(ctx, recurringJob); err != nil {
		return fmt.Errorf("failed to schedule recurring job: %w", err)
	}

	// In a real implementation, you would set up a timer to reschedule the job
	// For now, we'll just log it
	js.logger.Info("Recurring job scheduled",
		zap.String("job_id", recurringJob.ID),
		zap.Duration("interval", interval))

	return nil
}

// ProcessJobs processes jobs from the queue
func (js *JobScheduler) ProcessJobs(ctx context.Context, processor JobProcessor) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Get next job
			job, err := js.queue.Dequeue(ctx)
			if err != nil {
				js.logger.Error("Failed to dequeue job", zap.Error(err))
				continue
			}

			if job == nil {
				// No jobs available, wait a bit
				time.Sleep(1 * time.Second)
				continue
			}

			// Process job
			if err := js.processJob(ctx, job, processor); err != nil {
				js.logger.Error("Failed to process job",
					zap.String("job_id", job.ID),
					zap.Error(err))

				// Mark job as failed
				if err := js.queue.FailJob(ctx, job.ID, err.Error()); err != nil {
					js.logger.Error("Failed to mark job as failed", zap.Error(err))
				}
			}
		}
	}
}

// RetryFailedJobs retries failed jobs
func (js *JobScheduler) RetryFailedJobs(ctx context.Context) error {
	// This is a simplified implementation
	// In a real implementation, you would get failed jobs from the queue

	js.logger.Info("Retrying failed jobs")

	// For now, we'll just log that we're retrying
	// In production, you would implement the actual retry logic

	return nil
}

// GetStats returns scheduler statistics
func (js *JobScheduler) GetStats() *SchedulerStats {
	js.mu.RLock()
	defer js.mu.RUnlock()

	stats := *js.stats
	return &stats
}

// Helper methods

func (js *JobScheduler) processJob(ctx context.Context, job *Job, processor JobProcessor) error {
	start := time.Now()

	// Update stats
	js.mu.Lock()
	js.stats.ProcessedJobs++
	js.mu.Unlock()

	// Process the job
	result, err := processor.Process(ctx, job)

	duration := time.Since(start)

	// Update stats
	js.mu.Lock()
	js.stats.AverageProcessTime = (js.stats.AverageProcessTime + duration) / 2
	js.mu.Unlock()

	if err != nil {
		// Check if we should retry
		if job.RetryCount < job.MaxRetries {
			js.logger.Info("Job failed, will retry",
				zap.String("job_id", job.ID),
				zap.Int("retry_count", job.RetryCount),
				zap.Int("max_retries", job.MaxRetries),
				zap.Error(err))

			// Retry the job
			if retryErr := js.queue.RetryJob(ctx, job.ID); retryErr != nil {
				js.logger.Error("Failed to retry job", zap.Error(retryErr))
			}

			// Update stats
			js.mu.Lock()
			js.stats.RetriedJobs++
			js.mu.Unlock()

			return nil // Don't treat retry as an error
		}

		// Update stats
		js.mu.Lock()
		js.stats.FailedJobs++
		js.mu.Unlock()

		return fmt.Errorf("job processing failed: %w", err)
	}

	// Mark job as completed
	if err := js.queue.CompleteJob(ctx, job.ID, result); err != nil {
		return fmt.Errorf("failed to mark job as completed: %w", err)
	}

	js.logger.Info("Job processed successfully",
		zap.String("job_id", job.ID),
		zap.String("job_type", job.Type),
		zap.Duration("duration", duration))

	return nil
}

// JobProcessor interface for processing jobs
type JobProcessor interface {
	Process(ctx context.Context, job *Job) (map[string]interface{}, error)
}

// DefaultJobProcessor is a default implementation of JobProcessor
type DefaultJobProcessor struct {
	logger *zap.Logger
}

// NewDefaultJobProcessor creates a new default job processor
func NewDefaultJobProcessor(logger *zap.Logger) *DefaultJobProcessor {
	return &DefaultJobProcessor{
		logger: logger,
	}
}

// Process processes a job
func (djp *DefaultJobProcessor) Process(ctx context.Context, job *Job) (map[string]interface{}, error) {
	djp.logger.Info("Processing job",
		zap.String("job_id", job.ID),
		zap.String("job_type", job.Type))

	// Simulate job processing
	time.Sleep(100 * time.Millisecond)

	// Return a simple result
	result := map[string]interface{}{
		"status":       "completed",
		"processed_at": time.Now(),
		"job_id":       job.ID,
	}

	return result, nil
}

// PriorityQueue implements a priority queue for jobs
type PriorityQueue struct {
	jobs []*ScheduledJob
	mu   sync.RWMutex
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		jobs: make([]*ScheduledJob, 0),
	}
}

// Push adds a job to the priority queue
func (pq *PriorityQueue) Push(job *ScheduledJob) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	pq.jobs = append(pq.jobs, job)

	// Sort by priority (higher priority first)
	sort.Slice(pq.jobs, func(i, j int) bool {
		return pq.jobs[i].Priority > pq.jobs[j].Priority
	})
}

// Pop removes and returns the highest priority job
func (pq *PriorityQueue) Pop() *ScheduledJob {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.jobs) == 0 {
		return nil
	}

	job := pq.jobs[0]
	pq.jobs = pq.jobs[1:]

	return job
}

// Len returns the number of jobs in the queue
func (pq *PriorityQueue) Len() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	return len(pq.jobs)
}

// IsEmpty returns true if the queue is empty
func (pq *PriorityQueue) IsEmpty() bool {
	return pq.Len() == 0
}
