package risk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// BackupJobManager manages background backup jobs
type BackupJobManager struct {
	logger    *zap.Logger
	jobs      map[string]*BackupJob
	jobsMutex sync.RWMutex
	backupSvc *BackupService
	scheduler *BackupScheduler
}

// NewBackupJobManager creates a new backup job manager
func NewBackupJobManager(logger *zap.Logger, backupSvc *BackupService) *BackupJobManager {
	return &BackupJobManager{
		logger:    logger,
		jobs:      make(map[string]*BackupJob),
		backupSvc: backupSvc,
		scheduler: NewBackupScheduler(logger),
	}
}

// BackupJob represents a backup job
type BackupJob struct {
	ID          string                 `json:"id"`
	BusinessID  string                 `json:"business_id,omitempty"`
	BackupType  BackupType             `json:"backup_type"`
	Status      BackupJobStatus        `json:"status"`
	Progress    int                    `json:"progress"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Result      *BackupResponse        `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Schedule    *BackupSchedule        `json:"schedule,omitempty"`
}

// BackupJobStatus represents the status of a backup job
type BackupJobStatus string

const (
	BackupJobStatusPending   BackupJobStatus = "pending"
	BackupJobStatusRunning   BackupJobStatus = "running"
	BackupJobStatusCompleted BackupJobStatus = "completed"
	BackupJobStatusFailed    BackupJobStatus = "failed"
	BackupJobStatusCancelled BackupJobStatus = "cancelled"
)

// BackupSchedule represents a backup schedule
type BackupSchedule struct {
	ID            string                 `json:"id"`
	BusinessID    string                 `json:"business_id,omitempty"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	BackupType    BackupType             `json:"backup_type"`
	IncludeData   []BackupDataType       `json:"include_data"`
	Schedule      string                 `json:"schedule"` // Cron expression
	RetentionDays int                    `json:"retention_days"`
	Enabled       bool                   `json:"enabled"`
	LastRun       *time.Time             `json:"last_run,omitempty"`
	NextRun       *time.Time             `json:"next_run,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CreateBackupJob creates a new backup job
func (bjm *BackupJobManager) CreateBackupJob(ctx context.Context, request *BackupRequest) (*BackupJob, error) {
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Validate the backup request
	if err := bjm.backupSvc.validateBackupRequest(request); err != nil {
		return nil, fmt.Errorf("invalid backup request: %w", err)
	}

	// Create new backup job
	job := &BackupJob{
		ID:         fmt.Sprintf("backup_job_%d", time.Now().UnixNano()),
		BusinessID: request.BusinessID,
		BackupType: request.BackupType,
		Status:     BackupJobStatusPending,
		Progress:   0,
		CreatedAt:  time.Now(),
		Metadata:   request.Metadata,
	}

	// Store the job
	bjm.jobsMutex.Lock()
	bjm.jobs[job.ID] = job
	bjm.jobsMutex.Unlock()

	bjm.logger.Info("Backup job created",
		zap.String("request_id", requestID.(string)),
		zap.String("job_id", job.ID),
		zap.String("business_id", job.BusinessID),
		zap.String("backup_type", string(job.BackupType)))

	// Start the job in background
	go bjm.processBackupJob(ctx, job, request)

	return job, nil
}

// CreateScheduledBackup creates a scheduled backup
func (bjm *BackupJobManager) CreateScheduledBackup(schedule *BackupSchedule) error {
	if err := bjm.validateBackupSchedule(schedule); err != nil {
		return fmt.Errorf("invalid backup schedule: %w", err)
	}

	// Add to scheduler
	if err := bjm.scheduler.AddSchedule(schedule); err != nil {
		return fmt.Errorf("failed to add schedule: %w", err)
	}

	bjm.logger.Info("Scheduled backup created",
		zap.String("schedule_id", schedule.ID),
		zap.String("business_id", schedule.BusinessID),
		zap.String("name", schedule.Name),
		zap.String("schedule", schedule.Schedule))

	return nil
}

// GetBackupJob retrieves a backup job by ID
func (bjm *BackupJobManager) GetBackupJob(jobID string) (*BackupJob, error) {
	bjm.jobsMutex.RLock()
	defer bjm.jobsMutex.RUnlock()

	job, exists := bjm.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("backup job not found: %s", jobID)
	}

	return job, nil
}

// ListBackupJobs lists all backup jobs for a business
func (bjm *BackupJobManager) ListBackupJobs(businessID string) ([]*BackupJob, error) {
	bjm.jobsMutex.RLock()
	defer bjm.jobsMutex.RUnlock()

	var jobs []*BackupJob
	for _, job := range bjm.jobs {
		if businessID == "" || job.BusinessID == businessID {
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}

// CancelBackupJob cancels a pending backup job
func (bjm *BackupJobManager) CancelBackupJob(jobID string) error {
	bjm.jobsMutex.Lock()
	defer bjm.jobsMutex.Unlock()

	job, exists := bjm.jobs[jobID]
	if !exists {
		return fmt.Errorf("backup job not found: %s", jobID)
	}

	if job.Status != BackupJobStatusPending {
		return fmt.Errorf("cannot cancel job with status: %s", job.Status)
	}

	job.Status = BackupJobStatusCancelled
	now := time.Now()
	job.CompletedAt = &now

	bjm.logger.Info("Backup job cancelled",
		zap.String("job_id", jobID),
		zap.String("business_id", job.BusinessID))

	return nil
}

// CleanupOldJobs removes old completed jobs
func (bjm *BackupJobManager) CleanupOldJobs(olderThan time.Time) error {
	bjm.jobsMutex.Lock()
	defer bjm.jobsMutex.Unlock()

	var jobsToDelete []string
	for jobID, job := range bjm.jobs {
		if (job.Status == BackupJobStatusCompleted || job.Status == BackupJobStatusFailed || job.Status == BackupJobStatusCancelled) &&
			job.CompletedAt != nil && job.CompletedAt.Before(olderThan) {
			jobsToDelete = append(jobsToDelete, jobID)
		}
	}

	for _, jobID := range jobsToDelete {
		delete(bjm.jobs, jobID)
	}

	bjm.logger.Info("Cleaned up old backup jobs",
		zap.Int("jobs_deleted", len(jobsToDelete)),
		zap.Time("cutoff_time", olderThan))

	return nil
}

// GetJobStatistics returns statistics about backup jobs
func (bjm *BackupJobManager) GetJobStatistics() map[string]interface{} {
	bjm.jobsMutex.RLock()
	defer bjm.jobsMutex.RUnlock()

	stats := map[string]interface{}{
		"total_jobs":     len(bjm.jobs),
		"pending_jobs":   0,
		"running_jobs":   0,
		"completed_jobs": 0,
		"failed_jobs":    0,
		"cancelled_jobs": 0,
	}

	for _, job := range bjm.jobs {
		switch job.Status {
		case BackupJobStatusPending:
			stats["pending_jobs"] = stats["pending_jobs"].(int) + 1
		case BackupJobStatusRunning:
			stats["running_jobs"] = stats["running_jobs"].(int) + 1
		case BackupJobStatusCompleted:
			stats["completed_jobs"] = stats["completed_jobs"].(int) + 1
		case BackupJobStatusFailed:
			stats["failed_jobs"] = stats["failed_jobs"].(int) + 1
		case BackupJobStatusCancelled:
			stats["cancelled_jobs"] = stats["cancelled_jobs"].(int) + 1
		}
	}

	return stats
}

// StartScheduler starts the backup scheduler
func (bjm *BackupJobManager) StartScheduler(ctx context.Context) error {
	return bjm.scheduler.Start(ctx, bjm)
}

// StopScheduler stops the backup scheduler
func (bjm *BackupJobManager) StopScheduler() error {
	return bjm.scheduler.Stop()
}

// processBackupJob processes a backup job in the background
func (bjm *BackupJobManager) processBackupJob(ctx context.Context, job *BackupJob, request *BackupRequest) {
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	bjm.logger.Info("Starting backup job processing",
		zap.String("request_id", requestID.(string)),
		zap.String("job_id", job.ID),
		zap.String("business_id", job.BusinessID))

	// Update job status to running
	bjm.updateJobStatus(job, BackupJobStatusRunning, 10, nil)

	// Perform the backup
	response, err := bjm.backupSvc.CreateBackup(ctx, request)

	bjm.updateJobStatus(job, BackupJobStatusRunning, 80, nil)

	if err != nil {
		bjm.logger.Error("Backup job failed",
			zap.String("request_id", requestID.(string)),
			zap.String("job_id", job.ID),
			zap.String("business_id", job.BusinessID),
			zap.Error(err))

		bjm.updateJobStatus(job, BackupJobStatusFailed, 100, err)
		return
	}

	// Update job with result
	now := time.Now()
	job.Status = BackupJobStatusCompleted
	job.Progress = 100
	job.CompletedAt = &now
	job.Result = response

	bjm.logger.Info("Backup job completed successfully",
		zap.String("request_id", requestID.(string)),
		zap.String("job_id", job.ID),
		zap.String("business_id", job.BusinessID),
		zap.String("backup_id", response.BackupID),
		zap.Int("record_count", response.RecordCount))
}

// updateJobStatus updates the status of a backup job
func (bjm *BackupJobManager) updateJobStatus(job *BackupJob, status BackupJobStatus, progress int, err error) {
	bjm.jobsMutex.Lock()
	defer bjm.jobsMutex.Unlock()

	job.Status = status
	job.Progress = progress

	if err != nil {
		job.Error = err.Error()
	}

	if status == BackupJobStatusRunning && job.StartedAt == nil {
		now := time.Now()
		job.StartedAt = &now
	}
}

// validateBackupSchedule validates a backup schedule
func (bjm *BackupJobManager) validateBackupSchedule(schedule *BackupSchedule) error {
	if schedule == nil {
		return fmt.Errorf("backup schedule cannot be nil")
	}

	if schedule.ID == "" {
		return fmt.Errorf("schedule ID is required")
	}

	if schedule.Name == "" {
		return fmt.Errorf("schedule name is required")
	}

	if schedule.Schedule == "" {
		return fmt.Errorf("schedule expression is required")
	}

	if schedule.BackupType == "" {
		return fmt.Errorf("backup type is required")
	}

	if len(schedule.IncludeData) == 0 {
		return fmt.Errorf("include data is required")
	}

	// Validate backup type
	switch schedule.BackupType {
	case BackupTypeFull, BackupTypeIncremental, BackupTypeDifferential, BackupTypeBusiness, BackupTypeSystem:
		// Valid types
	default:
		return fmt.Errorf("invalid backup type: %s", schedule.BackupType)
	}

	// Validate include data types
	for _, dataType := range schedule.IncludeData {
		switch dataType {
		case BackupDataTypeAssessments, BackupDataTypeFactors, BackupDataTypeTrends, BackupDataTypeAlerts, BackupDataTypeHistory, BackupDataTypeConfig, BackupDataTypeAll:
			// Valid types
		default:
			return fmt.Errorf("invalid backup data type: %s", dataType)
		}
	}

	return nil
}

// BackupScheduler manages scheduled backups
type BackupScheduler struct {
	logger    *zap.Logger
	schedules map[string]*BackupSchedule
	running   bool
	stopChan  chan struct{}
}

// NewBackupScheduler creates a new backup scheduler
func NewBackupScheduler(logger *zap.Logger) *BackupScheduler {
	return &BackupScheduler{
		logger:    logger,
		schedules: make(map[string]*BackupSchedule),
		stopChan:  make(chan struct{}),
	}
}

// AddSchedule adds a backup schedule
func (bs *BackupScheduler) AddSchedule(schedule *BackupSchedule) error {
	bs.schedules[schedule.ID] = schedule
	bs.logger.Info("Backup schedule added",
		zap.String("schedule_id", schedule.ID),
		zap.String("name", schedule.Name))
	return nil
}

// RemoveSchedule removes a backup schedule
func (bs *BackupScheduler) RemoveSchedule(scheduleID string) error {
	if _, exists := bs.schedules[scheduleID]; !exists {
		return fmt.Errorf("schedule not found: %s", scheduleID)
	}

	delete(bs.schedules, scheduleID)
	bs.logger.Info("Backup schedule removed",
		zap.String("schedule_id", scheduleID))
	return nil
}

// ListSchedules lists all backup schedules
func (bs *BackupScheduler) ListSchedules() []*BackupSchedule {
	schedules := make([]*BackupSchedule, 0, len(bs.schedules))
	for _, schedule := range bs.schedules {
		schedules = append(schedules, schedule)
	}
	return schedules
}

// Start starts the backup scheduler
func (bs *BackupScheduler) Start(ctx context.Context, jobManager *BackupJobManager) error {
	if bs.running {
		return fmt.Errorf("scheduler is already running")
	}

	bs.running = true
	bs.stopChan = make(chan struct{})

	go bs.runScheduler(ctx, jobManager)

	bs.logger.Info("Backup scheduler started")
	return nil
}

// Stop stops the backup scheduler
func (bs *BackupScheduler) Stop() error {
	if !bs.running {
		return fmt.Errorf("scheduler is not running")
	}

	close(bs.stopChan)
	bs.running = false

	bs.logger.Info("Backup scheduler stopped")
	return nil
}

// runScheduler runs the backup scheduler loop
func (bs *BackupScheduler) runScheduler(ctx context.Context, jobManager *BackupJobManager) {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for {
		select {
		case <-bs.stopChan:
			return
		case <-ticker.C:
			bs.checkSchedules(ctx, jobManager)
		}
	}
}

// checkSchedules checks for schedules that need to run
func (bs *BackupScheduler) checkSchedules(ctx context.Context, jobManager *BackupJobManager) {
	now := time.Now()

	for _, schedule := range bs.schedules {
		if !schedule.Enabled {
			continue
		}

		// Simple schedule checking (in production, use a proper cron library)
		if bs.shouldRunSchedule(schedule, now) {
			bs.runScheduledBackup(ctx, jobManager, schedule)
		}
	}
}

// shouldRunSchedule determines if a schedule should run
func (bs *BackupScheduler) shouldRunSchedule(schedule *BackupSchedule, now time.Time) bool {
	// Simple implementation - check if it's time to run
	// In production, use a proper cron expression parser
	if schedule.LastRun == nil {
		return true
	}

	// Run daily at the same time
	return now.Sub(*schedule.LastRun) >= 24*time.Hour
}

// runScheduledBackup runs a scheduled backup
func (bs *BackupScheduler) runScheduledBackup(ctx context.Context, jobManager *BackupJobManager, schedule *BackupSchedule) {
	request := &BackupRequest{
		BusinessID:    schedule.BusinessID,
		BackupType:    schedule.BackupType,
		IncludeData:   schedule.IncludeData,
		RetentionDays: schedule.RetentionDays,
		Metadata:      schedule.Metadata,
	}

	// Create backup job
	job, err := jobManager.CreateBackupJob(ctx, request)
	if err != nil {
		bs.logger.Error("Failed to create scheduled backup job",
			zap.String("schedule_id", schedule.ID),
			zap.String("business_id", schedule.BusinessID),
			zap.Error(err))
		return
	}

	// Update schedule
	now := time.Now()
	schedule.LastRun = &now
	schedule.NextRun = &[]time.Time{now.Add(24 * time.Hour)}[0]

	bs.logger.Info("Scheduled backup job created",
		zap.String("schedule_id", schedule.ID),
		zap.String("job_id", job.ID),
		zap.String("business_id", schedule.BusinessID))
}
