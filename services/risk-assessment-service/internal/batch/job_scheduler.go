package batch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// JobScheduler manages scheduled batch jobs
type JobScheduler interface {
	// ScheduleJob schedules a batch job for future execution
	ScheduleJob(ctx context.Context, job *BatchJob, executeAt time.Time) error

	// CancelScheduledJob cancels a scheduled job
	CancelScheduledJob(ctx context.Context, jobID string) error

	// GetScheduledJobs gets all scheduled jobs
	GetScheduledJobs(ctx context.Context, tenantID string) ([]*ScheduledJob, error)

	// Start starts the scheduler
	Start(ctx context.Context) error

	// Stop stops the scheduler
	Stop() error
}

// ScheduledJob represents a job scheduled for future execution
type ScheduledJob struct {
	ID        string                 `json:"id" db:"id"`
	JobID     string                 `json:"job_id" db:"job_id"`
	TenantID  string                 `json:"tenant_id" db:"tenant_id"`
	ExecuteAt time.Time              `json:"execute_at" db:"execute_at"`
	Status    ScheduleStatus         `json:"status" db:"status"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata" db:"metadata"`
}

// ScheduleStatus represents the status of a scheduled job
type ScheduleStatus string

const (
	ScheduleStatusPending   ScheduleStatus = "pending"
	ScheduleStatusExecuting ScheduleStatus = "executing"
	ScheduleStatusCompleted ScheduleStatus = "completed"
	ScheduleStatusCancelled ScheduleStatus = "cancelled"
	ScheduleStatusFailed    ScheduleStatus = "failed"
)

// DefaultJobScheduler implements JobScheduler with default behavior
type DefaultJobScheduler struct {
	repository    BatchJobRepository
	jobManager    JobManager
	logger        *zap.Logger
	scheduledJobs map[string]*ScheduledJob
	scheduledMu   sync.RWMutex
	ticker        *time.Ticker
	ctx           context.Context
	cancel        context.CancelFunc
	started       bool
	startedMu     sync.RWMutex
}

// NewDefaultJobScheduler creates a new default job scheduler
func NewDefaultJobScheduler(
	repository BatchJobRepository,
	jobManager JobManager,
	logger *zap.Logger,
) *DefaultJobScheduler {
	return &DefaultJobScheduler{
		repository:    repository,
		jobManager:    jobManager,
		logger:        logger,
		scheduledJobs: make(map[string]*ScheduledJob),
	}
}

// ScheduleJob schedules a batch job for future execution
func (js *DefaultJobScheduler) ScheduleJob(ctx context.Context, job *BatchJob, executeAt time.Time) error {
	js.logger.Info("Scheduling batch job",
		zap.String("job_id", job.ID),
		zap.String("tenant_id", job.TenantID),
		zap.Time("execute_at", executeAt))

	// Validate execution time
	if executeAt.Before(time.Now()) {
		return fmt.Errorf("execution time cannot be in the past")
	}

	// Create scheduled job
	scheduledJob := &ScheduledJob{
		ID:        fmt.Sprintf("sched_%s", job.ID),
		JobID:     job.ID,
		TenantID:  job.TenantID,
		ExecuteAt: executeAt,
		Status:    ScheduleStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  job.Metadata,
	}

	// Add to scheduled jobs map
	js.scheduledMu.Lock()
	js.scheduledJobs[scheduledJob.ID] = scheduledJob
	js.scheduledMu.Unlock()

	// Update job status to pending
	job.Status = JobStatusPending
	if err := js.repository.UpdateBatchJob(ctx, job); err != nil {
		// Remove from scheduled jobs if update fails
		js.scheduledMu.Lock()
		delete(js.scheduledJobs, scheduledJob.ID)
		js.scheduledMu.Unlock()
		return fmt.Errorf("failed to update job status: %w", err)
	}

	js.logger.Info("Batch job scheduled successfully",
		zap.String("job_id", job.ID),
		zap.String("scheduled_id", scheduledJob.ID),
		zap.Time("execute_at", executeAt))

	return nil
}

// CancelScheduledJob cancels a scheduled job
func (js *DefaultJobScheduler) CancelScheduledJob(ctx context.Context, jobID string) error {
	js.logger.Info("Cancelling scheduled job",
		zap.String("job_id", jobID))

	js.scheduledMu.Lock()
	defer js.scheduledMu.Unlock()

	// Find scheduled job
	var scheduledJob *ScheduledJob
	for _, sj := range js.scheduledJobs {
		if sj.JobID == jobID {
			scheduledJob = sj
			break
		}
	}

	if scheduledJob == nil {
		return fmt.Errorf("scheduled job not found: %s", jobID)
	}

	if scheduledJob.Status != ScheduleStatusPending {
		return fmt.Errorf("cannot cancel job in status: %s", scheduledJob.Status)
	}

	// Update status to cancelled
	scheduledJob.Status = ScheduleStatusCancelled
	scheduledJob.UpdatedAt = time.Now()

	// Update the actual batch job status
	job, err := js.repository.GetBatchJob(ctx, scheduledJob.TenantID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get batch job: %w", err)
	}

	if job != nil {
		job.MarkCancelled()
		if err := js.repository.UpdateBatchJob(ctx, job); err != nil {
			return fmt.Errorf("failed to update batch job: %w", err)
		}
	}

	js.logger.Info("Scheduled job cancelled successfully",
		zap.String("job_id", jobID),
		zap.String("scheduled_id", scheduledJob.ID))

	return nil
}

// GetScheduledJobs gets all scheduled jobs for a tenant
func (js *DefaultJobScheduler) GetScheduledJobs(ctx context.Context, tenantID string) ([]*ScheduledJob, error) {
	js.scheduledMu.RLock()
	defer js.scheduledMu.RUnlock()

	var jobs []*ScheduledJob
	for _, job := range js.scheduledJobs {
		if job.TenantID == tenantID {
			jobs = append(jobs, job)
		}
	}

	js.logger.Debug("Retrieved scheduled jobs",
		zap.String("tenant_id", tenantID),
		zap.Int("count", len(jobs)))

	return jobs, nil
}

// Start starts the scheduler
func (js *DefaultJobScheduler) Start(ctx context.Context) error {
	js.startedMu.Lock()
	defer js.startedMu.Unlock()

	if js.started {
		return fmt.Errorf("scheduler is already started")
	}

	js.logger.Info("Starting job scheduler")

	js.ctx, js.cancel = context.WithCancel(ctx)

	// Start ticker to check for scheduled jobs every minute
	js.ticker = time.NewTicker(1 * time.Minute)

	// Start scheduling loop
	go js.schedulingLoop()

	js.started = true

	js.logger.Info("Job scheduler started successfully")

	return nil
}

// Stop stops the scheduler
func (js *DefaultJobScheduler) Stop() error {
	js.startedMu.Lock()
	defer js.startedMu.Unlock()

	if !js.started {
		return fmt.Errorf("scheduler is not started")
	}

	js.logger.Info("Stopping job scheduler")

	// Cancel context
	js.cancel()

	// Stop ticker
	if js.ticker != nil {
		js.ticker.Stop()
	}

	js.started = false

	js.logger.Info("Job scheduler stopped successfully")

	return nil
}

// schedulingLoop runs the main scheduling loop
func (js *DefaultJobScheduler) schedulingLoop() {
	js.logger.Info("Starting scheduling loop")

	for {
		select {
		case <-js.ctx.Done():
			js.logger.Info("Scheduling loop stopped due to context cancellation")
			return
		case <-js.ticker.C:
			js.processScheduledJobs()
		}
	}
}

// processScheduledJobs processes jobs that are ready to execute
func (js *DefaultJobScheduler) processScheduledJobs() {
	js.logger.Debug("Processing scheduled jobs")

	now := time.Now()
	var jobsToExecute []*ScheduledJob

	js.scheduledMu.RLock()
	for _, job := range js.scheduledJobs {
		if job.Status == ScheduleStatusPending && job.ExecuteAt.Before(now) {
			jobsToExecute = append(jobsToExecute, job)
		}
	}
	js.scheduledMu.RUnlock()

	js.logger.Debug("Found jobs ready for execution",
		zap.Int("count", len(jobsToExecute)))

	// Execute ready jobs
	for _, scheduledJob := range jobsToExecute {
		js.executeScheduledJob(scheduledJob)
	}
}

// executeScheduledJob executes a scheduled job
func (js *DefaultJobScheduler) executeScheduledJob(scheduledJob *ScheduledJob) {
	js.logger.Info("Executing scheduled job",
		zap.String("job_id", scheduledJob.JobID),
		zap.String("scheduled_id", scheduledJob.ID))

	// Update status to executing
	js.scheduledMu.Lock()
	scheduledJob.Status = ScheduleStatusExecuting
	scheduledJob.UpdatedAt = time.Now()
	js.scheduledMu.Unlock()

	// Get the batch job
	job, err := js.repository.GetBatchJob(js.ctx, scheduledJob.TenantID, scheduledJob.JobID)
	if err != nil {
		js.logger.Error("Failed to get batch job for execution",
			zap.String("job_id", scheduledJob.JobID),
			zap.Error(err))

		js.scheduledMu.Lock()
		scheduledJob.Status = ScheduleStatusFailed
		scheduledJob.UpdatedAt = time.Now()
		js.scheduledMu.Unlock()
		return
	}

	if job == nil {
		js.logger.Error("Batch job not found for execution",
			zap.String("job_id", scheduledJob.JobID))

		js.scheduledMu.Lock()
		scheduledJob.Status = ScheduleStatusFailed
		scheduledJob.UpdatedAt = time.Now()
		js.scheduledMu.Unlock()
		return
	}

	// Submit job to job manager
	_, err = js.jobManager.SubmitBatchJob(js.ctx, &BatchJobRequest{
		JobType:    job.JobType,
		Requests:   []map[string]interface{}{}, // This would be populated from job data
		Priority:   job.Priority,
		MaxRetries: job.MaxRetries,
		Metadata:   job.Metadata,
		CreatedBy:  job.CreatedBy,
	})
	if err != nil {
		js.logger.Error("Failed to submit scheduled job to job manager",
			zap.String("job_id", scheduledJob.JobID),
			zap.Error(err))

		js.scheduledMu.Lock()
		scheduledJob.Status = ScheduleStatusFailed
		scheduledJob.UpdatedAt = time.Now()
		js.scheduledMu.Unlock()
		return
	}

	// Mark as completed
	js.scheduledMu.Lock()
	scheduledJob.Status = ScheduleStatusCompleted
	scheduledJob.UpdatedAt = time.Now()
	js.scheduledMu.Unlock()

	js.logger.Info("Scheduled job executed successfully",
		zap.String("job_id", scheduledJob.JobID),
		zap.String("scheduled_id", scheduledJob.ID))
}

// RecurringJobScheduler manages recurring batch jobs
type RecurringJobScheduler struct {
	repository    BatchJobRepository
	jobManager    JobManager
	logger        *zap.Logger
	recurringJobs map[string]*RecurringJob
	recurringMu   sync.RWMutex
	ticker        *time.Ticker
	ctx           context.Context
	cancel        context.CancelFunc
	started       bool
	startedMu     sync.RWMutex
}

// RecurringJob represents a recurring batch job
type RecurringJob struct {
	ID        string                 `json:"id" db:"id"`
	TenantID  string                 `json:"tenant_id" db:"tenant_id"`
	Name      string                 `json:"name" db:"name"`
	JobType   string                 `json:"job_type" db:"job_type"`
	Schedule  string                 `json:"schedule" db:"schedule"` // Cron expression
	IsActive  bool                   `json:"is_active" db:"is_active"`
	LastRun   *time.Time             `json:"last_run" db:"last_run"`
	NextRun   time.Time              `json:"next_run" db:"next_run"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy string                 `json:"created_by" db:"created_by"`
	Metadata  map[string]interface{} `json:"metadata" db:"metadata"`
}

// NewRecurringJobScheduler creates a new recurring job scheduler
func NewRecurringJobScheduler(
	repository BatchJobRepository,
	jobManager JobManager,
	logger *zap.Logger,
) *RecurringJobScheduler {
	return &RecurringJobScheduler{
		repository:    repository,
		jobManager:    jobManager,
		logger:        logger,
		recurringJobs: make(map[string]*RecurringJob),
	}
}

// ScheduleRecurringJob schedules a recurring batch job
func (rjs *RecurringJobScheduler) ScheduleRecurringJob(ctx context.Context, job *RecurringJob) error {
	rjs.logger.Info("Scheduling recurring job",
		zap.String("job_id", job.ID),
		zap.String("name", job.Name),
		zap.String("schedule", job.Schedule))

	// Validate cron expression (this would use a cron library in real implementation)
	if job.Schedule == "" {
		return fmt.Errorf("schedule cannot be empty")
	}

	// Calculate next run time (simplified - would use cron library)
	job.NextRun = time.Now().Add(24 * time.Hour) // Default to daily

	rjs.recurringMu.Lock()
	rjs.recurringJobs[job.ID] = job
	rjs.recurringMu.Unlock()

	rjs.logger.Info("Recurring job scheduled successfully",
		zap.String("job_id", job.ID),
		zap.String("name", job.Name),
		zap.Time("next_run", job.NextRun))

	return nil
}

// Start starts the recurring job scheduler
func (rjs *RecurringJobScheduler) Start(ctx context.Context) error {
	rjs.startedMu.Lock()
	defer rjs.startedMu.Unlock()

	if rjs.started {
		return fmt.Errorf("recurring scheduler is already started")
	}

	rjs.logger.Info("Starting recurring job scheduler")

	rjs.ctx, rjs.cancel = context.WithCancel(ctx)

	// Start ticker to check for recurring jobs every minute
	rjs.ticker = time.NewTicker(1 * time.Minute)

	// Start scheduling loop
	go rjs.recurringSchedulingLoop()

	rjs.started = true

	rjs.logger.Info("Recurring job scheduler started successfully")

	return nil
}

// Stop stops the recurring job scheduler
func (rjs *RecurringJobScheduler) Stop() error {
	rjs.startedMu.Lock()
	defer rjs.startedMu.Unlock()

	if !rjs.started {
		return fmt.Errorf("recurring scheduler is not started")
	}

	rjs.logger.Info("Stopping recurring job scheduler")

	// Cancel context
	rjs.cancel()

	// Stop ticker
	if rjs.ticker != nil {
		rjs.ticker.Stop()
	}

	rjs.started = false

	rjs.logger.Info("Recurring job scheduler stopped successfully")

	return nil
}

// recurringSchedulingLoop runs the main recurring scheduling loop
func (rjs *RecurringJobScheduler) recurringSchedulingLoop() {
	rjs.logger.Info("Starting recurring scheduling loop")

	for {
		select {
		case <-rjs.ctx.Done():
			rjs.logger.Info("Recurring scheduling loop stopped due to context cancellation")
			return
		case <-rjs.ticker.C:
			rjs.processRecurringJobs()
		}
	}
}

// processRecurringJobs processes recurring jobs that are ready to execute
func (rjs *RecurringJobScheduler) processRecurringJobs() {
	rjs.logger.Debug("Processing recurring jobs")

	now := time.Now()
	var jobsToExecute []*RecurringJob

	rjs.recurringMu.RLock()
	for _, job := range rjs.recurringJobs {
		if job.IsActive && job.NextRun.Before(now) {
			jobsToExecute = append(jobsToExecute, job)
		}
	}
	rjs.recurringMu.RUnlock()

	rjs.logger.Debug("Found recurring jobs ready for execution",
		zap.Int("count", len(jobsToExecute)))

	// Execute ready jobs
	for _, recurringJob := range jobsToExecute {
		rjs.executeRecurringJob(recurringJob)
	}
}

// executeRecurringJob executes a recurring job
func (rjs *RecurringJobScheduler) executeRecurringJob(recurringJob *RecurringJob) {
	rjs.logger.Info("Executing recurring job",
		zap.String("job_id", recurringJob.ID),
		zap.String("name", recurringJob.Name))

	// Update last run time
	now := time.Now()
	recurringJob.LastRun = &now
	recurringJob.UpdatedAt = now

	// Calculate next run time (simplified - would use cron library)
	recurringJob.NextRun = now.Add(24 * time.Hour) // Default to daily

	// Submit job to job manager
	_, err := rjs.jobManager.SubmitBatchJob(rjs.ctx, &BatchJobRequest{
		JobType:    recurringJob.JobType,
		Requests:   []map[string]interface{}{}, // This would be populated from job configuration
		Priority:   5,                          // Default priority
		MaxRetries: 3,                          // Default retries
		Metadata:   recurringJob.Metadata,
		CreatedBy:  recurringJob.CreatedBy,
	})
	if err != nil {
		rjs.logger.Error("Failed to submit recurring job to job manager",
			zap.String("job_id", recurringJob.ID),
			zap.Error(err))
		return
	}

	// Update recurring job
	rjs.recurringMu.Lock()
	rjs.recurringJobs[recurringJob.ID] = recurringJob
	rjs.recurringMu.Unlock()

	rjs.logger.Info("Recurring job executed successfully",
		zap.String("job_id", recurringJob.ID),
		zap.String("name", recurringJob.Name),
		zap.Time("next_run", recurringJob.NextRun))
}
