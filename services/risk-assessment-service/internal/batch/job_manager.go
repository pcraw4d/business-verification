package batch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// JobManager manages batch job processing
type JobManager interface {
	// SubmitBatchJob submits a new batch job for processing
	SubmitBatchJob(ctx context.Context, request *BatchJobRequest) (*BatchJobResponse, error)

	// GetBatchJobStatus gets the current status of a batch job
	GetBatchJobStatus(ctx context.Context, tenantID, jobID string) (*BatchJobStatus, error)

	// GetBatchJobResults gets the results of a completed batch job
	GetBatchJobResults(ctx context.Context, tenantID, jobID string) (*BatchJobResults, error)

	// CancelBatchJob cancels a running batch job
	CancelBatchJob(ctx context.Context, tenantID, jobID string) error

	// ResumeBatchJob resumes a paused or failed batch job
	ResumeBatchJob(ctx context.Context, tenantID, jobID string) error

	// ListBatchJobs lists batch jobs with filters
	ListBatchJobs(ctx context.Context, filter *BatchJobFilter) ([]*BatchJob, error)

	// GetBatchJobMetrics gets metrics for batch jobs
	GetBatchJobMetrics(ctx context.Context, tenantID string, startDate, endDate time.Time) (*BatchJobMetrics, error)

	// Start starts the job manager
	Start(ctx context.Context) error

	// Stop stops the job manager
	Stop() error
}

// DefaultJobManager implements JobManager with default configuration
type DefaultJobManager struct {
	repository   BatchJobRepository
	processor    BatchProcessor
	config       *BatchJobConfig
	logger       *zap.Logger
	workerPool   *WorkerPool
	jobQueue     chan *BatchJob
	activeJobs   map[string]*BatchJob
	activeJobsMu sync.RWMutex
	stopChan     chan struct{}
	started      bool
	startedMu    sync.RWMutex
}

// NewDefaultJobManager creates a new default job manager
func NewDefaultJobManager(
	repository BatchJobRepository,
	processor BatchProcessor,
	config *BatchJobConfig,
	logger *zap.Logger,
) *DefaultJobManager {
	if config == nil {
		config = DefaultBatchJobConfig()
	}

	return &DefaultJobManager{
		repository: repository,
		processor:  processor,
		config:     config,
		logger:     logger,
		jobQueue:   make(chan *BatchJob, config.MaxConcurrentJobs*2),
		activeJobs: make(map[string]*BatchJob),
		stopChan:   make(chan struct{}),
	}
}

// SubmitBatchJob submits a new batch job for processing
func (jm *DefaultJobManager) SubmitBatchJob(ctx context.Context, request *BatchJobRequest) (*BatchJobResponse, error) {
	jm.logger.Info("Submitting batch job",
		zap.String("job_type", request.JobType),
		zap.Int("total_requests", len(request.Requests)),
		zap.String("created_by", request.CreatedBy))

	// Validate request
	if len(request.Requests) > jm.config.MaxRequestsPerJob {
		return nil, fmt.Errorf("too many requests: %d exceeds maximum %d", len(request.Requests), jm.config.MaxRequestsPerJob)
	}

	// Generate job ID
	jobID := generateJobID()

	// Create batch job
	job := &BatchJob{
		ID:            jobID,
		TenantID:      getTenantIDFromContext(ctx), // This would be extracted from context
		Status:        JobStatusPending,
		JobType:       request.JobType,
		TotalRequests: len(request.Requests),
		Completed:     0,
		Failed:        0,
		Progress:      0.0,
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
		CreatedBy:     request.CreatedBy,
		Priority:      request.Priority,
		Metadata:      request.Metadata,
		Results:       []BatchResult{},
		RetryCount:    0,
		MaxRetries:    request.MaxRetries,
	}

	// Store requests in metadata for processing
	if job.Metadata == nil {
		job.Metadata = make(map[string]interface{})
	}
	job.Metadata["requests"] = request.Requests

	// Set default values
	if job.Priority == 0 {
		job.Priority = 5 // Default priority
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = jm.config.MaxRetryAttempts
	}

	// Save job to repository
	if err := jm.repository.SaveBatchJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to save batch job: %w", err)
	}

	// Add to job queue
	select {
	case jm.jobQueue <- job:
		jm.logger.Info("Batch job queued successfully",
			zap.String("job_id", jobID),
			zap.String("job_type", request.JobType))
	default:
		// Queue is full, update job status
		job.Status = JobStatusPending
		job.Error = "Job queue is full, will be processed when capacity is available"
		jm.repository.UpdateBatchJob(ctx, job)
		jm.logger.Warn("Job queue is full, job will be processed later",
			zap.String("job_id", jobID))
	}

	// Calculate estimated time
	estimatedTime := jm.calculateEstimatedTime(len(request.Requests))

	response := &BatchJobResponse{
		ID:            jobID,
		Status:        job.Status,
		TotalRequests: job.TotalRequests,
		CreatedAt:     job.CreatedAt,
		EstimatedTime: estimatedTime,
	}

	jm.logger.Info("Batch job submitted successfully",
		zap.String("job_id", jobID),
		zap.String("status", string(job.Status)),
		zap.Int("total_requests", job.TotalRequests))

	return response, nil
}

// GetBatchJobStatus gets the current status of a batch job
func (jm *DefaultJobManager) GetBatchJobStatus(ctx context.Context, tenantID, jobID string) (*BatchJobStatus, error) {
	jm.logger.Debug("Getting batch job status",
		zap.String("job_id", jobID),
		zap.String("tenant_id", tenantID))

	job, err := jm.repository.GetBatchJob(ctx, tenantID, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch job: %w", err)
	}

	if job == nil {
		return nil, fmt.Errorf("batch job not found: %s", jobID)
	}

	status := &BatchJobStatus{
		ID:            job.ID,
		Status:        job.Status,
		Progress:      job.Progress,
		TotalRequests: job.TotalRequests,
		Completed:     job.Completed,
		Failed:        job.Failed,
		StartedAt:     job.StartedAt,
		CompletedAt:   job.CompletedAt,
		Error:         job.Error,
		Metadata:      job.Metadata,
	}

	// Calculate estimated completion time
	if job.Status == JobStatusProcessing {
		status.EstimatedTime = job.GetEstimatedCompletionTime()
	}

	jm.logger.Debug("Batch job status retrieved",
		zap.String("job_id", jobID),
		zap.String("status", string(job.Status)),
		zap.Float64("progress", job.Progress))

	return status, nil
}

// GetBatchJobResults gets the results of a completed batch job
func (jm *DefaultJobManager) GetBatchJobResults(ctx context.Context, tenantID, jobID string) (*BatchJobResults, error) {
	jm.logger.Debug("Getting batch job results",
		zap.String("job_id", jobID),
		zap.String("tenant_id", tenantID))

	job, err := jm.repository.GetBatchJob(ctx, tenantID, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch job: %w", err)
	}

	if job == nil {
		return nil, fmt.Errorf("batch job not found: %s", jobID)
	}

	// Get detailed results
	results, err := jm.repository.GetBatchJobResults(ctx, tenantID, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch results: %w", err)
	}

	// Calculate summary
	summary := jm.calculateBatchSummary(results)

	// Generate download URL (this would be implemented based on your storage solution)
	downloadURL := jm.generateDownloadURL(jobID)

	batchResults := &BatchJobResults{
		ID:           job.ID,
		Status:       job.Status,
		TotalResults: len(results),
		Results:      convertToBatchResults(results),
		Summary:      summary,
		DownloadURL:  downloadURL,
	}

	jm.logger.Debug("Batch job results retrieved",
		zap.String("job_id", jobID),
		zap.Int("total_results", len(results)),
		zap.Float64("success_rate", summary.SuccessRate))

	return batchResults, nil
}

// CancelBatchJob cancels a running batch job
func (jm *DefaultJobManager) CancelBatchJob(ctx context.Context, tenantID, jobID string) error {
	jm.logger.Info("Cancelling batch job",
		zap.String("job_id", jobID),
		zap.String("tenant_id", tenantID))

	job, err := jm.repository.GetBatchJob(ctx, tenantID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get batch job: %w", err)
	}

	if job == nil {
		return fmt.Errorf("batch job not found: %s", jobID)
	}

	if !job.IsActive() {
		return fmt.Errorf("cannot cancel job in status: %s", job.Status)
	}

	// Mark job as cancelled
	job.MarkCancelled()
	if err := jm.repository.UpdateBatchJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update batch job: %w", err)
	}

	// Remove from active jobs
	jm.activeJobsMu.Lock()
	delete(jm.activeJobs, jobID)
	jm.activeJobsMu.Unlock()

	jm.logger.Info("Batch job cancelled successfully",
		zap.String("job_id", jobID),
		zap.String("tenant_id", tenantID))

	return nil
}

// ResumeBatchJob resumes a paused or failed batch job
func (jm *DefaultJobManager) ResumeBatchJob(ctx context.Context, tenantID, jobID string) error {
	jm.logger.Info("Resuming batch job",
		zap.String("job_id", jobID),
		zap.String("tenant_id", tenantID))

	job, err := jm.repository.GetBatchJob(ctx, tenantID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get batch job: %w", err)
	}

	if job == nil {
		return fmt.Errorf("batch job not found: %s", jobID)
	}

	if job.Status != JobStatusFailed && job.Status != JobStatusPaused {
		return fmt.Errorf("cannot resume job in status: %s", job.Status)
	}

	// Check if job can be retried
	if job.Status == JobStatusFailed && !job.CanRetry() {
		return fmt.Errorf("job has exceeded maximum retry attempts")
	}

	// Increment retry count if needed
	if job.Status == JobStatusFailed {
		job.IncrementRetry()
	} else {
		job.Status = JobStatusPending
		job.LastUpdatedAt = time.Now()
	}

	// Update job in repository
	if err := jm.repository.UpdateBatchJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update batch job: %w", err)
	}

	// Add to job queue
	select {
	case jm.jobQueue <- job:
		jm.logger.Info("Batch job queued for resumption",
			zap.String("job_id", jobID))
	default:
		jm.logger.Warn("Job queue is full, job will be processed later",
			zap.String("job_id", jobID))
	}

	jm.logger.Info("Batch job resumed successfully",
		zap.String("job_id", jobID),
		zap.String("tenant_id", tenantID))

	return nil
}

// ListBatchJobs lists batch jobs with filters
func (jm *DefaultJobManager) ListBatchJobs(ctx context.Context, filter *BatchJobFilter) ([]*BatchJob, error) {
	return jm.repository.ListBatchJobs(ctx, filter)
}

// GetBatchJobMetrics gets metrics for batch jobs
func (jm *DefaultJobManager) GetBatchJobMetrics(ctx context.Context, tenantID string, startDate, endDate time.Time) (*BatchJobMetrics, error) {
	return jm.repository.GetBatchJobMetrics(ctx, tenantID, startDate, endDate)
}

// Start starts the job manager
func (jm *DefaultJobManager) Start(ctx context.Context) error {
	jm.startedMu.Lock()
	defer jm.startedMu.Unlock()

	if jm.started {
		return fmt.Errorf("job manager is already started")
	}

	jm.logger.Info("Starting job manager",
		zap.Int("max_concurrent_jobs", jm.config.MaxConcurrentJobs),
		zap.Int("max_requests_per_job", jm.config.MaxRequestsPerJob))

	// Create worker pool
	jm.workerPool = NewWorkerPool(jm.config.MaxConcurrentJobs, jm.processor, jm.logger)

	// Start worker pool
	if err := jm.workerPool.Start(ctx); err != nil {
		return fmt.Errorf("failed to start worker pool: %w", err)
	}

	// Start job processing loop
	go jm.jobProcessingLoop(ctx)

	// Start cleanup routine
	go jm.cleanupRoutine(ctx)

	jm.started = true

	jm.logger.Info("Job manager started successfully")

	return nil
}

// Stop stops the job manager
func (jm *DefaultJobManager) Stop() error {
	jm.startedMu.Lock()
	defer jm.startedMu.Unlock()

	if !jm.started {
		return fmt.Errorf("job manager is not started")
	}

	jm.logger.Info("Stopping job manager")

	// Signal stop
	close(jm.stopChan)

	// Stop worker pool
	if jm.workerPool != nil {
		if err := jm.workerPool.Stop(); err != nil {
			jm.logger.Error("Failed to stop worker pool", zap.Error(err))
		}
	}

	jm.started = false

	jm.logger.Info("Job manager stopped successfully")

	return nil
}

// jobProcessingLoop processes jobs from the queue
func (jm *DefaultJobManager) jobProcessingLoop(ctx context.Context) {
	jm.logger.Info("Starting job processing loop")

	for {
		select {
		case <-ctx.Done():
			jm.logger.Info("Job processing loop stopped due to context cancellation")
			return
		case <-jm.stopChan:
			jm.logger.Info("Job processing loop stopped")
			return
		case job := <-jm.jobQueue:
			jm.processJob(ctx, job)
		}
	}
}

// processJob processes a single batch job
func (jm *DefaultJobManager) processJob(ctx context.Context, job *BatchJob) {
	jm.logger.Info("Processing batch job",
		zap.String("job_id", job.ID),
		zap.String("job_type", job.JobType),
		zap.Int("total_requests", job.TotalRequests))

	// Add to active jobs
	jm.activeJobsMu.Lock()
	jm.activeJobs[job.ID] = job
	jm.activeJobsMu.Unlock()

	// Update job status to processing
	job.Status = JobStatusProcessing
	now := time.Now()
	job.StartedAt = &now
	job.LastUpdatedAt = now

	if err := jm.repository.UpdateBatchJob(ctx, job); err != nil {
		jm.logger.Error("Failed to update job status", zap.Error(err))
		job.MarkFailed(fmt.Sprintf("Failed to update job status: %v", err))
		jm.repository.UpdateBatchJob(ctx, job)
		return
	}

	// Submit job to worker pool
	if err := jm.workerPool.SubmitJob(ctx, job); err != nil {
		jm.logger.Error("Failed to submit job to worker pool", zap.Error(err))
		job.MarkFailed(fmt.Sprintf("Failed to submit job: %v", err))
		jm.repository.UpdateBatchJob(ctx, job)
		return
	}

	jm.logger.Info("Batch job submitted to worker pool",
		zap.String("job_id", job.ID))
}

// cleanupRoutine periodically cleans up old jobs
func (jm *DefaultJobManager) cleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour) // Run cleanup daily
	defer ticker.Stop()

	jm.logger.Info("Starting cleanup routine")

	for {
		select {
		case <-ctx.Done():
			jm.logger.Info("Cleanup routine stopped due to context cancellation")
			return
		case <-jm.stopChan:
			jm.logger.Info("Cleanup routine stopped")
			return
		case <-ticker.C:
			jm.performCleanup(ctx)
		}
	}
}

// performCleanup performs cleanup of old jobs
func (jm *DefaultJobManager) performCleanup(ctx context.Context) {
	jm.logger.Info("Performing cleanup of old jobs")

	olderThan := time.Now().AddDate(0, 0, -jm.config.CleanupAfterDays)
	if err := jm.repository.CleanupOldJobs(ctx, olderThan); err != nil {
		jm.logger.Error("Failed to cleanup old jobs", zap.Error(err))
	} else {
		jm.logger.Info("Cleanup completed successfully")
	}
}

// Helper functions

func (jm *DefaultJobManager) calculateEstimatedTime(totalRequests int) string {
	// Simple estimation based on average processing time
	avgTimePerRequest := 2 * time.Second // This would be calculated from historical data
	totalTime := time.Duration(totalRequests) * avgTimePerRequest

	if totalTime < time.Minute {
		return fmt.Sprintf("%.0f seconds", totalTime.Seconds())
	} else if totalTime < time.Hour {
		return fmt.Sprintf("%.1f minutes", totalTime.Minutes())
	} else {
		return fmt.Sprintf("%.1f hours", totalTime.Hours())
	}
}

func (jm *DefaultJobManager) calculateBatchSummary(results []*BatchResult) BatchSummary {
	if len(results) == 0 {
		return BatchSummary{}
	}

	var successful, failed int
	var totalDuration, minDuration, maxDuration time.Duration

	for _, result := range results {
		if result.Status == "success" {
			successful++
		} else {
			failed++
		}

		totalDuration += result.Duration
		if minDuration == 0 || result.Duration < minDuration {
			minDuration = result.Duration
		}
		if result.Duration > maxDuration {
			maxDuration = result.Duration
		}
	}

	successRate := float64(successful) / float64(len(results)) * 100
	avgDuration := float64(totalDuration.Nanoseconds()) / float64(len(results)) / 1e6 // Convert to milliseconds

	return BatchSummary{
		TotalRequests:   len(results),
		Successful:      successful,
		Failed:          failed,
		SuccessRate:     successRate,
		AverageDuration: avgDuration,
		MinDuration:     float64(minDuration.Nanoseconds()) / 1e6,
		MaxDuration:     float64(maxDuration.Nanoseconds()) / 1e6,
		TotalDuration:   float64(totalDuration.Nanoseconds()) / 1e6,
	}
}

func (jm *DefaultJobManager) generateDownloadURL(jobID string) string {
	// This would generate a signed URL for downloading results
	// Implementation depends on your storage solution (S3, GCS, etc.)
	return fmt.Sprintf("/api/v1/batch/%s/download", jobID)
}

func convertToBatchResults(results []*BatchResult) []BatchResult {
	converted := make([]BatchResult, len(results))
	for i, result := range results {
		converted[i] = *result
	}
	return converted
}

func generateJobID() string {
	// Generate a unique job ID
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}

func getTenantIDFromContext(ctx context.Context) string {
	// This would extract tenant ID from context
	// Implementation depends on your authentication/authorization system
	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		if id, ok := tenantID.(string); ok {
			return id
		}
	}
	return "default" // Fallback for development
}
