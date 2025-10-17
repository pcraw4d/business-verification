package batch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"kyb-platform/services/risk-assessment-service/internal/queue"

	"go.uber.org/zap"
)

// AsyncWorker processes batch jobs asynchronously
type AsyncWorker struct {
	queue     *queue.RedisQueue
	scheduler *queue.JobScheduler
	logger    *zap.Logger
	workers   int
	mu        sync.RWMutex
	stats     *WorkerStats
	running   bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// WorkerStats represents statistics for the async worker
type WorkerStats struct {
	TotalJobsProcessed int64         `json:"total_jobs_processed"`
	SuccessfulJobs     int64         `json:"successful_jobs"`
	FailedJobs         int64         `json:"failed_jobs"`
	RetriedJobs        int64         `json:"retried_jobs"`
	AverageProcessTime time.Duration `json:"average_process_time"`
	ActiveWorkers      int           `json:"active_workers"`
	LastProcessed      time.Time     `json:"last_processed"`
}

// WorkerConfig represents configuration for the async worker
type WorkerConfig struct {
	Workers             int           `json:"workers"`
	BatchSize           int           `json:"batch_size"`
	ProcessTimeout      time.Duration `json:"process_timeout"`
	RetryDelay          time.Duration `json:"retry_delay"`
	MaxRetries          int           `json:"max_retries"`
	EnableStats         bool          `json:"enable_stats"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
}

// NewAsyncWorker creates a new async worker
func NewAsyncWorker(queue *queue.RedisQueue, scheduler *queue.JobScheduler, config *WorkerConfig, logger *zap.Logger) *AsyncWorker {
	if config == nil {
		config = &WorkerConfig{
			Workers:             5,
			BatchSize:           10,
			ProcessTimeout:      30 * time.Minute,
			RetryDelay:          1 * time.Second,
			MaxRetries:          3,
			EnableStats:         true,
			HealthCheckInterval: 1 * time.Minute,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &AsyncWorker{
		queue:     queue,
		scheduler: scheduler,
		logger:    logger,
		workers:   config.Workers,
		stats:     &WorkerStats{},
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start starts the async worker
func (aw *AsyncWorker) Start() error {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	if aw.running {
		return fmt.Errorf("worker is already running")
	}

	aw.running = true
	aw.stats.ActiveWorkers = aw.workers

	// Start worker goroutines
	for i := 0; i < aw.workers; i++ {
		go aw.workerLoop(i)
	}

	// Start health check goroutine
	go aw.healthCheckLoop()

	aw.logger.Info("Async worker started",
		zap.Int("workers", aw.workers))

	return nil
}

// Stop stops the async worker
func (aw *AsyncWorker) Stop() error {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	if !aw.running {
		return fmt.Errorf("worker is not running")
	}

	aw.running = false
	aw.cancel()

	aw.logger.Info("Async worker stopped")

	return nil
}

// SubmitBatchJob submits a batch job for processing
func (aw *AsyncWorker) SubmitBatchJob(ctx context.Context, jobType string, data map[string]interface{}) error {
	job := &queue.Job{
		ID:         generateAsyncJobID(),
		Type:       jobType,
		Data:       data,
		Priority:   5,
		MaxRetries: 3,
		Status:     "pending",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := aw.queue.Enqueue(ctx, job); err != nil {
		return fmt.Errorf("failed to enqueue batch job: %w", err)
	}

	aw.logger.Info("Batch job submitted",
		zap.String("job_id", job.ID),
		zap.String("job_type", jobType))

	return nil
}

// GetStats returns worker statistics
func (aw *AsyncWorker) GetStats() *WorkerStats {
	aw.mu.RLock()
	defer aw.mu.RUnlock()

	stats := *aw.stats
	return &stats
}

// IsRunning returns true if the worker is running
func (aw *AsyncWorker) IsRunning() bool {
	aw.mu.RLock()
	defer aw.mu.RUnlock()

	return aw.running
}

// Helper methods

func (aw *AsyncWorker) workerLoop(workerID int) {
	aw.logger.Info("Worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-aw.ctx.Done():
			aw.logger.Info("Worker stopped", zap.Int("worker_id", workerID))
			return
		default:
			// Process jobs
			if err := aw.processJobs(workerID); err != nil {
				aw.logger.Error("Worker error",
					zap.Int("worker_id", workerID),
					zap.Error(err))
			}
		}
	}
}

func (aw *AsyncWorker) processJobs(workerID int) error {
	// Get next job
	job, err := aw.queue.Dequeue(aw.ctx)
	if err != nil {
		return fmt.Errorf("failed to dequeue job: %w", err)
	}

	if job == nil {
		// No jobs available, wait a bit
		time.Sleep(1 * time.Second)
		return nil
	}

	// Process the job
	start := time.Now()

	aw.logger.Info("Processing job",
		zap.Int("worker_id", workerID),
		zap.String("job_id", job.ID),
		zap.String("job_type", job.Type))

	result, err := aw.processJob(aw.ctx, job)
	duration := time.Since(start)

	// Update stats
	aw.mu.Lock()
	aw.stats.TotalJobsProcessed++
	aw.stats.AverageProcessTime = (aw.stats.AverageProcessTime + duration) / 2
	aw.stats.LastProcessed = time.Now()
	aw.mu.Unlock()

	if err != nil {
		// Check if we should retry
		if job.RetryCount < job.MaxRetries {
			aw.logger.Info("Job failed, will retry",
				zap.String("job_id", job.ID),
				zap.Int("retry_count", job.RetryCount),
				zap.Int("max_retries", job.MaxRetries),
				zap.Error(err))

			// Retry the job
			if retryErr := aw.queue.RetryJob(aw.ctx, job.ID); retryErr != nil {
				aw.logger.Error("Failed to retry job", zap.Error(retryErr))
			}

			// Update stats
			aw.mu.Lock()
			aw.stats.RetriedJobs++
			aw.mu.Unlock()

			return nil // Don't treat retry as an error
		}

		// Mark job as failed
		if err := aw.queue.FailJob(aw.ctx, job.ID, err.Error()); err != nil {
			aw.logger.Error("Failed to mark job as failed", zap.Error(err))
		}

		// Update stats
		aw.mu.Lock()
		aw.stats.FailedJobs++
		aw.mu.Unlock()

		return fmt.Errorf("job processing failed: %w", err)
	}

	// Mark job as completed
	if err := aw.queue.CompleteJob(aw.ctx, job.ID, result); err != nil {
		return fmt.Errorf("failed to mark job as completed: %w", err)
	}

	// Update stats
	aw.mu.Lock()
	aw.stats.SuccessfulJobs++
	aw.mu.Unlock()

	aw.logger.Info("Job completed successfully",
		zap.Int("worker_id", workerID),
		zap.String("job_id", job.ID),
		zap.String("job_type", job.Type),
		zap.Duration("duration", duration))

	return nil
}

func (aw *AsyncWorker) processJob(ctx context.Context, job *queue.Job) (map[string]interface{}, error) {
	// Create a timeout context
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// Process based on job type
	switch job.Type {
	case "risk_assessment":
		return aw.processRiskAssessmentJob(ctx, job)
	case "compliance_check":
		return aw.processComplianceCheckJob(ctx, job)
	case "custom_model_test":
		return aw.processCustomModelTestJob(ctx, job)
	case "batch_verification":
		return aw.processBatchVerificationJob(ctx, job)
	default:
		return nil, fmt.Errorf("unknown job type: %s", job.Type)
	}
}

func (aw *AsyncWorker) processRiskAssessmentJob(ctx context.Context, job *queue.Job) (map[string]interface{}, error) {
	aw.logger.Info("Processing risk assessment job",
		zap.String("job_id", job.ID))

	// Simulate risk assessment processing
	time.Sleep(2 * time.Second)

	result := map[string]interface{}{
		"status":       "completed",
		"processed_at": time.Now(),
		"job_id":       job.ID,
		"job_type":     "risk_assessment",
		"risk_score":   0.75,
		"risk_level":   "medium",
		"factors":      []string{"industry_risk", "country_risk"},
	}

	return result, nil
}

func (aw *AsyncWorker) processComplianceCheckJob(ctx context.Context, job *queue.Job) (map[string]interface{}, error) {
	aw.logger.Info("Processing compliance check job",
		zap.String("job_id", job.ID))

	// Simulate compliance check processing
	time.Sleep(1 * time.Second)

	result := map[string]interface{}{
		"status":            "completed",
		"processed_at":      time.Now(),
		"job_id":            job.ID,
		"job_type":          "compliance_check",
		"compliance_status": "passed",
		"checks_performed":  []string{"kyc", "aml", "sanctions"},
	}

	return result, nil
}

func (aw *AsyncWorker) processCustomModelTestJob(ctx context.Context, job *queue.Job) (map[string]interface{}, error) {
	aw.logger.Info("Processing custom model test job",
		zap.String("job_id", job.ID))

	// Simulate custom model test processing
	time.Sleep(3 * time.Second)

	result := map[string]interface{}{
		"status":         "completed",
		"processed_at":   time.Now(),
		"job_id":         job.ID,
		"job_type":       "custom_model_test",
		"model_accuracy": 0.92,
		"test_results": map[string]interface{}{
			"precision": 0.89,
			"recall":    0.94,
			"f1_score":  0.91,
		},
	}

	return result, nil
}

func (aw *AsyncWorker) processBatchVerificationJob(ctx context.Context, job *queue.Job) (map[string]interface{}, error) {
	aw.logger.Info("Processing batch verification job",
		zap.String("job_id", job.ID))

	// Simulate batch verification processing
	time.Sleep(5 * time.Second)

	result := map[string]interface{}{
		"status":          "completed",
		"processed_at":    time.Now(),
		"job_id":          job.ID,
		"job_type":        "batch_verification",
		"total_processed": 1000,
		"successful":      950,
		"failed":          50,
		"success_rate":    0.95,
	}

	return result, nil
}

func (aw *AsyncWorker) healthCheckLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-aw.ctx.Done():
			return
		case <-ticker.C:
			aw.performHealthCheck()
		}
	}
}

func (aw *AsyncWorker) performHealthCheck() {
	// Check queue length
	queueLength, err := aw.queue.GetQueueLength(aw.ctx)
	if err != nil {
		aw.logger.Error("Failed to get queue length", zap.Error(err))
		return
	}

	// Check worker stats
	stats := aw.GetStats()

	aw.logger.Info("Health check",
		zap.Int64("queue_length", queueLength),
		zap.Int64("total_processed", stats.TotalJobsProcessed),
		zap.Int64("successful", stats.SuccessfulJobs),
		zap.Int64("failed", stats.FailedJobs),
		zap.Int64("retried", stats.RetriedJobs),
		zap.Duration("avg_process_time", stats.AverageProcessTime))

	// Alert if queue is getting too long
	if queueLength > 1000 {
		aw.logger.Warn("Queue length is high",
			zap.Int64("queue_length", queueLength))
	}

	// Alert if failure rate is too high
	if stats.TotalJobsProcessed > 0 {
		failureRate := float64(stats.FailedJobs) / float64(stats.TotalJobsProcessed)
		if failureRate > 0.1 { // 10% failure rate
			aw.logger.Warn("High failure rate detected",
				zap.Float64("failure_rate", failureRate))
		}
	}
}

func generateAsyncJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}
