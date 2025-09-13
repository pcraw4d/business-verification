package risk

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestBackupJobManager_CreateBackupJob(t *testing.T) {
	logger := zap.NewNop()
	mockBackupSvc := &MockBackupService{}
	jobManager := NewBackupJobManager(logger, mockBackupSvc)

	// Mock the validation to pass
	mockBackupSvc.On("validateBackupRequest", mock.Anything).Return(nil)

	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
		Metadata:    map[string]interface{}{"test": "value"},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

	job, err := jobManager.CreateBackupJob(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "test-business-123", job.BusinessID)
	assert.Equal(t, BackupTypeFull, job.BackupType)
	assert.Equal(t, BackupJobStatusPending, job.Status)
	assert.Equal(t, 0, job.Progress)
	assert.NotEmpty(t, job.ID)
	assert.NotNil(t, job.CreatedAt)

	mockBackupSvc.AssertExpectations(t)
}

func TestBackupJobManager_CreateBackupJob_ValidationError(t *testing.T) {
	logger := zap.NewNop()
	mockBackupSvc := &MockBackupService{}
	jobManager := NewBackupJobManager(logger, mockBackupSvc)

	// Mock the validation to fail
	mockBackupSvc.On("validateBackupRequest", mock.Anything).Return(assert.AnError)

	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

	job, err := jobManager.CreateBackupJob(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Contains(t, err.Error(), "invalid backup request")

	mockBackupSvc.AssertExpectations(t)
}

func TestBackupJobManager_GetBackupJob(t *testing.T) {
	logger := zap.NewNop()
	mockBackupSvc := &MockBackupService{}
	jobManager := NewBackupJobManager(logger, mockBackupSvc)

	// Create a job first
	mockBackupSvc.On("validateBackupRequest", mock.Anything).Return(nil)

	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateBackupJob(ctx, request)
	assert.NoError(t, err)

	// Retrieve the job
	retrievedJob, err := jobManager.GetBackupJob(job.ID)

	assert.NoError(t, err)
	assert.NotNil(t, retrievedJob)
	assert.Equal(t, job.ID, retrievedJob.ID)
	assert.Equal(t, job.BusinessID, retrievedJob.BusinessID)

	mockBackupSvc.AssertExpectations(t)
}

func TestBackupJobManager_GetBackupJob_NotFound(t *testing.T) {
	logger := zap.NewNop()
	mockBackupSvc := &MockBackupService{}
	jobManager := NewBackupJobManager(logger, mockBackupSvc)

	job, err := jobManager.GetBackupJob("non-existent-job-id")

	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Contains(t, err.Error(), "backup job not found")
}

func TestBackupJobManager_ListBackupJobs(t *testing.T) {
	logger := zap.NewNop()
	mockBackupSvc := &MockBackupService{}
	jobManager := NewBackupJobManager(logger, mockBackupSvc)

	// Create multiple jobs for the same business
	mockBackupSvc.On("validateBackupRequest", mock.Anything).Return(nil)

	businessID := "test-business-123"
	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

	// Create first job
	request1 := &BackupRequest{
		BusinessID:  businessID,
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}
	job1, err := jobManager.CreateBackupJob(ctx, request1)
	assert.NoError(t, err)

	// Create second job
	request2 := &BackupRequest{
		BusinessID:  businessID,
		BackupType:  BackupTypeIncremental,
		IncludeData: []BackupDataType{BackupDataTypeFactors},
	}
	job2, err := jobManager.CreateBackupJob(ctx, request2)
	assert.NoError(t, err)

	// Create job for different business
	request3 := &BackupRequest{
		BusinessID:  "other-business-456",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}
	_, err = jobManager.CreateBackupJob(ctx, request3)
	assert.NoError(t, err)

	// List jobs for the first business
	jobs, err := jobManager.ListBackupJobs(businessID)

	assert.NoError(t, err)
	assert.Len(t, jobs, 2)

	// Verify the jobs are for the correct business
	for _, job := range jobs {
		assert.Equal(t, businessID, job.BusinessID)
		assert.True(t, job.ID == job1.ID || job.ID == job2.ID)
	}

	// List all jobs
	allJobs, err := jobManager.ListBackupJobs("")
	assert.NoError(t, err)
	assert.Len(t, allJobs, 3)

	mockBackupSvc.AssertExpectations(t)
}

func TestBackupJobManager_CancelBackupJob(t *testing.T) {
	logger := zap.NewNop()
	mockBackupSvc := &MockBackupService{}
	jobManager := NewBackupJobManager(logger, mockBackupSvc)

	// Create a job first
	mockBackupSvc.On("validateBackupRequest", mock.Anything).Return(nil)

	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateBackupJob(ctx, request)
	assert.NoError(t, err)

	// Cancel the job
	err = jobManager.CancelBackupJob(job.ID)

	assert.NoError(t, err)

	// Verify the job was cancelled
	retrievedJob, err := jobManager.GetBackupJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, BackupJobStatusCancelled, retrievedJob.Status)
	assert.NotNil(t, retrievedJob.CompletedAt)

	mockBackupSvc.AssertExpectations(t)
}

func TestBackupJobManager_CancelBackupJob_NotFound(t *testing.T) {
	logger := zap.NewNop()
	mockBackupSvc := &MockBackupService{}
	jobManager := NewBackupJobManager(logger, mockBackupSvc)

	err := jobManager.CancelBackupJob("non-existent-job-id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backup job not found")
}

func TestBackupJobManager_CancelBackupJob_NotPending(t *testing.T) {
	logger := zap.NewNop()
	mockBackupSvc := &MockBackupService{}
	jobManager := NewBackupJobManager(logger, mockBackupSvc)

	// Create a job first
	mockBackupSvc.On("validateBackupRequest", mock.Anything).Return(nil)

	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateBackupJob(ctx, request)
	assert.NoError(t, err)

	// Manually set the job status to completed
	job.Status = BackupJobStatusCompleted

	// Try to cancel the job
	err = jobManager.CancelBackupJob(job.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel job with status: completed")

	mockBackupSvc.AssertExpectations(t)
}

func TestBackupJobManager_CleanupOldJobs(t *testing.T) {
	logger := zap.NewNop()
	mockBackupSvc := &MockBackupService{}
	jobManager := NewBackupJobManager(logger, mockBackupSvc)

	// Create a job first
	mockBackupSvc.On("validateBackupRequest", mock.Anything).Return(nil)

	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateBackupJob(ctx, request)
	assert.NoError(t, err)

	// Manually set the job as completed with old timestamp
	job.Status = BackupJobStatusCompleted
	oldTime := time.Now().Add(-25 * time.Hour)
	job.CompletedAt = &oldTime

	// Cleanup jobs older than 24 hours
	cutoffTime := time.Now().Add(-24 * time.Hour)
	err = jobManager.CleanupOldJobs(cutoffTime)

	assert.NoError(t, err)

	// Verify the job was removed
	_, err = jobManager.GetBackupJob(job.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backup job not found")

	mockBackupSvc.AssertExpectations(t)
}

func TestBackupJobManager_GetJobStatistics(t *testing.T) {
	logger := zap.NewNop()
	mockBackupSvc := &MockBackupService{}
	jobManager := NewBackupJobManager(logger, mockBackupSvc)

	// Create multiple jobs with different statuses
	mockBackupSvc.On("validateBackupRequest", mock.Anything).Return(nil)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

	// Create a job
	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}
	job, err := jobManager.CreateBackupJob(ctx, request)
	assert.NoError(t, err)

	// Manually set different statuses for testing
	job.Status = BackupJobStatusCompleted

	// Get statistics
	stats := jobManager.GetJobStatistics()

	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats["total_jobs"])
	assert.Equal(t, 0, stats["pending_jobs"])
	assert.Equal(t, 0, stats["running_jobs"])
	assert.Equal(t, 1, stats["completed_jobs"])
	assert.Equal(t, 0, stats["failed_jobs"])
	assert.Equal(t, 0, stats["cancelled_jobs"])

	mockBackupSvc.AssertExpectations(t)
}

func TestBackupJobManager_ProcessBackupJob_Success(t *testing.T) {
	logger := zap.NewNop()
	mockBackupSvc := &MockBackupService{}
	jobManager := NewBackupJobManager(logger, mockBackupSvc)

	// Mock the validation and backup methods
	mockBackupSvc.On("validateBackupRequest", mock.Anything).Return(nil)
	mockBackupSvc.On("CreateBackup", mock.Anything, mock.Anything).Return(&BackupResponse{
		BackupID:    "test-backup-123",
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		Status:      BackupStatusCompleted,
		FilePath:    "/tmp/backup.json",
		FileSize:    1024,
		RecordCount: 10,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}, nil)

	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateBackupJob(ctx, request)
	assert.NoError(t, err)

	// Wait a bit for the background job to complete
	time.Sleep(200 * time.Millisecond)

	// Verify the job was processed
	retrievedJob, err := jobManager.GetBackupJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, BackupJobStatusCompleted, retrievedJob.Status)
	assert.Equal(t, 100, retrievedJob.Progress)
	assert.NotNil(t, retrievedJob.Result)

	mockBackupSvc.AssertExpectations(t)
}

func TestBackupJobManager_ProcessBackupJob_Failure(t *testing.T) {
	logger := zap.NewNop()
	mockBackupSvc := &MockBackupService{}
	jobManager := NewBackupJobManager(logger, mockBackupSvc)

	// Mock the validation and backup methods
	mockBackupSvc.On("validateBackupRequest", mock.Anything).Return(nil)
	mockBackupSvc.On("CreateBackup", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	request := &BackupRequest{
		BusinessID:  "test-business-123",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateBackupJob(ctx, request)
	assert.NoError(t, err)

	// Wait a bit for the background job to complete
	time.Sleep(200 * time.Millisecond)

	// Verify the job failed
	retrievedJob, err := jobManager.GetBackupJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, BackupJobStatusFailed, retrievedJob.Status)
	assert.Equal(t, 100, retrievedJob.Progress)
	assert.NotEmpty(t, retrievedJob.Error)

	mockBackupSvc.AssertExpectations(t)
}

func TestBackupJobManager_CreateScheduledBackup(t *testing.T) {
	logger := zap.NewNop()
	mockBackupSvc := &MockBackupService{}
	jobManager := NewBackupJobManager(logger, mockBackupSvc)

	schedule := &BackupSchedule{
		ID:            "test-schedule-123",
		BusinessID:    "test-business-123",
		Name:          "Daily Backup",
		Description:   "Daily full backup",
		BackupType:    BackupTypeFull,
		IncludeData:   []BackupDataType{BackupDataTypeAssessments},
		Schedule:      "0 2 * * *", // Daily at 2 AM
		RetentionDays: 30,
		Enabled:       true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := jobManager.CreateScheduledBackup(schedule)

	assert.NoError(t, err)

	// Verify schedule was added
	schedules := jobManager.scheduler.ListSchedules()
	assert.Len(t, schedules, 1)
	assert.Equal(t, schedule.ID, schedules[0].ID)
}

func TestBackupJobManager_CreateScheduledBackup_InvalidSchedule(t *testing.T) {
	logger := zap.NewNop()
	mockBackupSvc := &MockBackupService{}
	jobManager := NewBackupJobManager(logger, mockBackupSvc)

	tests := []struct {
		name     string
		schedule *BackupSchedule
		wantErr  string
	}{
		{
			name:     "nil schedule",
			schedule: nil,
			wantErr:  "backup schedule cannot be nil",
		},
		{
			name: "empty schedule ID",
			schedule: &BackupSchedule{
				ID:          "",
				Name:        "Test Schedule",
				BackupType:  BackupTypeFull,
				IncludeData: []BackupDataType{BackupDataTypeAssessments},
				Schedule:    "0 2 * * *",
			},
			wantErr: "schedule ID is required",
		},
		{
			name: "empty schedule name",
			schedule: &BackupSchedule{
				ID:          "test-schedule-123",
				Name:        "",
				BackupType:  BackupTypeFull,
				IncludeData: []BackupDataType{BackupDataTypeAssessments},
				Schedule:    "0 2 * * *",
			},
			wantErr: "schedule name is required",
		},
		{
			name: "empty schedule expression",
			schedule: &BackupSchedule{
				ID:          "test-schedule-123",
				Name:        "Test Schedule",
				BackupType:  BackupTypeFull,
				IncludeData: []BackupDataType{BackupDataTypeAssessments},
				Schedule:    "",
			},
			wantErr: "schedule expression is required",
		},
		{
			name: "empty backup type",
			schedule: &BackupSchedule{
				ID:          "test-schedule-123",
				Name:        "Test Schedule",
				BackupType:  "",
				IncludeData: []BackupDataType{BackupDataTypeAssessments},
				Schedule:    "0 2 * * *",
			},
			wantErr: "backup type is required",
		},
		{
			name: "empty include data",
			schedule: &BackupSchedule{
				ID:          "test-schedule-123",
				Name:        "Test Schedule",
				BackupType:  BackupTypeFull,
				IncludeData: []BackupDataType{},
				Schedule:    "0 2 * * *",
			},
			wantErr: "include data is required",
		},
		{
			name: "invalid backup type",
			schedule: &BackupSchedule{
				ID:          "test-schedule-123",
				Name:        "Test Schedule",
				BackupType:  "invalid_type",
				IncludeData: []BackupDataType{BackupDataTypeAssessments},
				Schedule:    "0 2 * * *",
			},
			wantErr: "invalid backup type",
		},
		{
			name: "invalid backup data type",
			schedule: &BackupSchedule{
				ID:          "test-schedule-123",
				Name:        "Test Schedule",
				BackupType:  BackupTypeFull,
				IncludeData: []BackupDataType{"invalid_data_type"},
				Schedule:    "0 2 * * *",
			},
			wantErr: "invalid backup data type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := jobManager.CreateScheduledBackup(tt.schedule)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestBackupScheduler_AddSchedule(t *testing.T) {
	logger := zap.NewNop()
	scheduler := NewBackupScheduler(logger)

	schedule := &BackupSchedule{
		ID:          "test-schedule-123",
		Name:        "Test Schedule",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
		Schedule:    "0 2 * * *",
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := scheduler.AddSchedule(schedule)

	assert.NoError(t, err)

	// Verify schedule was added
	schedules := scheduler.ListSchedules()
	assert.Len(t, schedules, 1)
	assert.Equal(t, schedule.ID, schedules[0].ID)
}

func TestBackupScheduler_RemoveSchedule(t *testing.T) {
	logger := zap.NewNop()
	scheduler := NewBackupScheduler(logger)

	schedule := &BackupSchedule{
		ID:          "test-schedule-123",
		Name:        "Test Schedule",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
		Schedule:    "0 2 * * *",
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Add schedule
	err := scheduler.AddSchedule(schedule)
	assert.NoError(t, err)

	// Remove schedule
	err = scheduler.RemoveSchedule(schedule.ID)
	assert.NoError(t, err)

	// Verify schedule was removed
	schedules := scheduler.ListSchedules()
	assert.Len(t, schedules, 0)
}

func TestBackupScheduler_RemoveSchedule_NotFound(t *testing.T) {
	logger := zap.NewNop()
	scheduler := NewBackupScheduler(logger)

	err := scheduler.RemoveSchedule("non-existent-schedule")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "schedule not found")
}

func TestBackupScheduler_StartStop(t *testing.T) {
	logger := zap.NewNop()
	scheduler := NewBackupScheduler(logger)

	// Start scheduler
	ctx := context.Background()
	err := scheduler.Start(ctx, nil)
	assert.NoError(t, err)
	assert.True(t, scheduler.running)

	// Stop scheduler
	err = scheduler.Stop()
	assert.NoError(t, err)
	assert.False(t, scheduler.running)
}

func TestBackupScheduler_Start_AlreadyRunning(t *testing.T) {
	logger := zap.NewNop()
	scheduler := NewBackupScheduler(logger)

	// Start scheduler
	ctx := context.Background()
	err := scheduler.Start(ctx, nil)
	assert.NoError(t, err)

	// Try to start again
	err = scheduler.Start(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scheduler is already running")

	// Clean up
	scheduler.Stop()
}

func TestBackupScheduler_Stop_NotRunning(t *testing.T) {
	logger := zap.NewNop()
	scheduler := NewBackupScheduler(logger)

	err := scheduler.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scheduler is not running")
}

func TestBackupScheduler_ListSchedules(t *testing.T) {
	logger := zap.NewNop()
	scheduler := NewBackupScheduler(logger)

	// Add multiple schedules
	schedule1 := &BackupSchedule{
		ID:          "schedule-1",
		Name:        "Schedule 1",
		BackupType:  BackupTypeFull,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
		Schedule:    "0 2 * * *",
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	schedule2 := &BackupSchedule{
		ID:          "schedule-2",
		Name:        "Schedule 2",
		BackupType:  BackupTypeIncremental,
		IncludeData: []BackupDataType{BackupDataTypeFactors},
		Schedule:    "0 3 * * *",
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := scheduler.AddSchedule(schedule1)
	assert.NoError(t, err)

	err = scheduler.AddSchedule(schedule2)
	assert.NoError(t, err)

	// List schedules
	schedules := scheduler.ListSchedules()
	assert.Len(t, schedules, 2)

	// Verify schedules
	scheduleIDs := make(map[string]bool)
	for _, schedule := range schedules {
		scheduleIDs[schedule.ID] = true
	}
	assert.True(t, scheduleIDs["schedule-1"])
	assert.True(t, scheduleIDs["schedule-2"])
}
