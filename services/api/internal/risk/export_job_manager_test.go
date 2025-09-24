package risk

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestExportJobManager_CreateExportJob(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Mock the validation to pass
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(nil)

	request := &ExportRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
		Metadata:   map[string]interface{}{"test": "value"},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

	job, err := jobManager.CreateExportJob(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "test-business-123", job.BusinessID)
	assert.Equal(t, ExportTypeAssessments, job.ExportType)
	assert.Equal(t, ExportFormatJSON, job.Format)
	assert.Equal(t, "pending", job.Status)
	assert.Equal(t, 0, job.Progress)
	assert.NotEmpty(t, job.ID)
	assert.NotNil(t, job.CreatedAt)

	mockExportSvc.AssertExpectations(t)
}

func TestExportJobManager_CreateExportJob_ValidationError(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Mock the validation to fail
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(assert.AnError)

	request := &ExportRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

	job, err := jobManager.CreateExportJob(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Contains(t, err.Error(), "invalid export request")

	mockExportSvc.AssertExpectations(t)
}

func TestExportJobManager_GetExportJob(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Create a job first
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(nil)

	request := &ExportRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateExportJob(ctx, request)
	assert.NoError(t, err)

	// Retrieve the job
	retrievedJob, err := jobManager.GetExportJob(job.ID)

	assert.NoError(t, err)
	assert.NotNil(t, retrievedJob)
	assert.Equal(t, job.ID, retrievedJob.ID)
	assert.Equal(t, job.BusinessID, retrievedJob.BusinessID)

	mockExportSvc.AssertExpectations(t)
}

func TestExportJobManager_GetExportJob_NotFound(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	job, err := jobManager.GetExportJob("non-existent-job-id")

	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Contains(t, err.Error(), "export job not found")
}

func TestExportJobManager_ListExportJobs(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Create multiple jobs for the same business
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(nil)

	businessID := "test-business-123"
	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

	// Create first job
	request1 := &ExportRequest{
		BusinessID: businessID,
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}
	job1, err := jobManager.CreateExportJob(ctx, request1)
	assert.NoError(t, err)

	// Create second job
	request2 := &ExportRequest{
		BusinessID: businessID,
		ExportType: ExportTypeFactors,
		Format:     ExportFormatCSV,
	}
	job2, err := jobManager.CreateExportJob(ctx, request2)
	assert.NoError(t, err)

	// Create job for different business
	request3 := &ExportRequest{
		BusinessID: "other-business-456",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}
	_, err = jobManager.CreateExportJob(ctx, request3)
	assert.NoError(t, err)

	// List jobs for the first business
	jobs, err := jobManager.ListExportJobs(businessID)

	assert.NoError(t, err)
	assert.Len(t, jobs, 2)

	// Verify the jobs are for the correct business
	for _, job := range jobs {
		assert.Equal(t, businessID, job.BusinessID)
		assert.True(t, job.ID == job1.ID || job.ID == job2.ID)
	}

	mockExportSvc.AssertExpectations(t)
}

func TestExportJobManager_CancelExportJob(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Create a job first
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(nil)

	request := &ExportRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateExportJob(ctx, request)
	assert.NoError(t, err)

	// Cancel the job
	err = jobManager.CancelExportJob(job.ID)

	assert.NoError(t, err)

	// Verify the job was cancelled
	retrievedJob, err := jobManager.GetExportJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, "cancelled", retrievedJob.Status)
	assert.NotNil(t, retrievedJob.CompletedAt)

	mockExportSvc.AssertExpectations(t)
}

func TestExportJobManager_CancelExportJob_NotFound(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	err := jobManager.CancelExportJob("non-existent-job-id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "export job not found")
}

func TestExportJobManager_CancelExportJob_NotPending(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Create a job first
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(nil)

	request := &ExportRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateExportJob(ctx, request)
	assert.NoError(t, err)

	// Manually set the job status to completed
	job.Status = "completed"

	// Try to cancel the job
	err = jobManager.CancelExportJob(job.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel job with status: completed")

	mockExportSvc.AssertExpectations(t)
}

func TestExportJobManager_CleanupOldJobs(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Create a job first
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(nil)

	request := &ExportRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateExportJob(ctx, request)
	assert.NoError(t, err)

	// Manually set the job as completed with old timestamp
	job.Status = "completed"
	oldTime := time.Now().Add(-25 * time.Hour)
	job.CompletedAt = &oldTime

	// Cleanup jobs older than 24 hours
	cutoffTime := time.Now().Add(-24 * time.Hour)
	err = jobManager.CleanupOldJobs(cutoffTime)

	assert.NoError(t, err)

	// Verify the job was removed
	_, err = jobManager.GetExportJob(job.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "export job not found")

	mockExportSvc.AssertExpectations(t)
}

func TestExportJobManager_GetJobStatistics(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Create multiple jobs with different statuses
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(nil)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

	// Create a job
	request := &ExportRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}
	job, err := jobManager.CreateExportJob(ctx, request)
	assert.NoError(t, err)

	// Manually set different statuses for testing
	job.Status = "completed"

	// Get statistics
	stats := jobManager.GetJobStatistics()

	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats["total_jobs"])
	assert.Equal(t, 0, stats["pending_jobs"])
	assert.Equal(t, 0, stats["processing_jobs"])
	assert.Equal(t, 1, stats["completed_jobs"])
	assert.Equal(t, 0, stats["failed_jobs"])
	assert.Equal(t, 0, stats["cancelled_jobs"])

	mockExportSvc.AssertExpectations(t)
}

func TestExportJobManager_ProcessExportJob_Assessments(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Mock the validation and export methods
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(nil)
	mockExportSvc.On("ExportRiskAssessments", mock.Anything, mock.Anything, mock.Anything).Return(&ExportResponse{
		ExportID:    "test-export-123",
		BusinessID:  "test-business-123",
		ExportType:  ExportTypeAssessments,
		Format:      ExportFormatJSON,
		Data:        map[string]interface{}{"test": "data"},
		RecordCount: 1,
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}, nil)

	request := &ExportRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateExportJob(ctx, request)
	assert.NoError(t, err)

	// Wait a bit for the background job to complete
	time.Sleep(200 * time.Millisecond)

	// Verify the job was processed
	retrievedJob, err := jobManager.GetExportJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, "completed", retrievedJob.Status)
	assert.Equal(t, 100, retrievedJob.Progress)
	assert.NotNil(t, retrievedJob.Result)

	mockExportSvc.AssertExpectations(t)
}

func TestExportJobManager_ProcessExportJob_Factors(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Mock the validation and export methods
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(nil)
	mockExportSvc.On("ExportRiskFactors", mock.Anything, mock.Anything, mock.Anything).Return(&ExportResponse{
		ExportID:    "test-export-123",
		BusinessID:  "test-business-123",
		ExportType:  ExportTypeFactors,
		Format:      ExportFormatJSON,
		Data:        map[string]interface{}{"test": "data"},
		RecordCount: 1,
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}, nil)

	request := &ExportRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeFactors,
		Format:     ExportFormatJSON,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateExportJob(ctx, request)
	assert.NoError(t, err)

	// Wait a bit for the background job to complete
	time.Sleep(200 * time.Millisecond)

	// Verify the job was processed
	retrievedJob, err := jobManager.GetExportJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, "completed", retrievedJob.Status)
	assert.Equal(t, 100, retrievedJob.Progress)
	assert.NotNil(t, retrievedJob.Result)

	mockExportSvc.AssertExpectations(t)
}

func TestExportJobManager_ProcessExportJob_Trends(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Mock the validation and export methods
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(nil)
	mockExportSvc.On("ExportRiskTrends", mock.Anything, mock.Anything, mock.Anything).Return(&ExportResponse{
		ExportID:    "test-export-123",
		BusinessID:  "test-business-123",
		ExportType:  ExportTypeTrends,
		Format:      ExportFormatJSON,
		Data:        map[string]interface{}{"test": "data"},
		RecordCount: 1,
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}, nil)

	request := &ExportRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeTrends,
		Format:     ExportFormatJSON,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateExportJob(ctx, request)
	assert.NoError(t, err)

	// Wait a bit for the background job to complete
	time.Sleep(200 * time.Millisecond)

	// Verify the job was processed
	retrievedJob, err := jobManager.GetExportJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, "completed", retrievedJob.Status)
	assert.Equal(t, 100, retrievedJob.Progress)
	assert.NotNil(t, retrievedJob.Result)

	mockExportSvc.AssertExpectations(t)
}

func TestExportJobManager_ProcessExportJob_Alerts(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Mock the validation and export methods
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(nil)
	mockExportSvc.On("ExportRiskAlerts", mock.Anything, mock.Anything, mock.Anything).Return(&ExportResponse{
		ExportID:    "test-export-123",
		BusinessID:  "test-business-123",
		ExportType:  ExportTypeAlerts,
		Format:      ExportFormatJSON,
		Data:        map[string]interface{}{"test": "data"},
		RecordCount: 1,
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}, nil)

	request := &ExportRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAlerts,
		Format:     ExportFormatJSON,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateExportJob(ctx, request)
	assert.NoError(t, err)

	// Wait a bit for the background job to complete
	time.Sleep(200 * time.Millisecond)

	// Verify the job was processed
	retrievedJob, err := jobManager.GetExportJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, "completed", retrievedJob.Status)
	assert.Equal(t, 100, retrievedJob.Progress)
	assert.NotNil(t, retrievedJob.Result)

	mockExportSvc.AssertExpectations(t)
}

func TestExportJobManager_ProcessExportJob_Reports(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Mock the validation method
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(nil)

	request := &ExportRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeReports,
		Format:     ExportFormatJSON,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateExportJob(ctx, request)
	assert.NoError(t, err)

	// Wait a bit for the background job to complete
	time.Sleep(200 * time.Millisecond)

	// Verify the job was processed
	retrievedJob, err := jobManager.GetExportJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, "completed", retrievedJob.Status)
	assert.Equal(t, 100, retrievedJob.Progress)
	assert.NotNil(t, retrievedJob.Result)

	mockExportSvc.AssertExpectations(t)
}

func TestExportJobManager_ProcessExportJob_All(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Mock the validation method
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(nil)

	request := &ExportRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAll,
		Format:     ExportFormatJSON,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateExportJob(ctx, request)
	assert.NoError(t, err)

	// Wait a bit for the background job to complete
	time.Sleep(200 * time.Millisecond)

	// Verify the job was processed
	retrievedJob, err := jobManager.GetExportJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, "completed", retrievedJob.Status)
	assert.Equal(t, 100, retrievedJob.Progress)
	assert.NotNil(t, retrievedJob.Result)

	mockExportSvc.AssertExpectations(t)
}

func TestExportJobManager_ProcessExportJob_UnsupportedType(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Mock the validation method
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(nil)

	request := &ExportRequest{
		BusinessID: "test-business-123",
		ExportType: "unsupported_type",
		Format:     ExportFormatJSON,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateExportJob(ctx, request)
	assert.NoError(t, err)

	// Wait a bit for the background job to complete
	time.Sleep(200 * time.Millisecond)

	// Verify the job failed
	retrievedJob, err := jobManager.GetExportJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, "failed", retrievedJob.Status)
	assert.Equal(t, 100, retrievedJob.Progress)
	assert.NotEmpty(t, retrievedJob.Error)
	assert.Contains(t, retrievedJob.Error, "unsupported export type")

	mockExportSvc.AssertExpectations(t)
}

func TestExportJobManager_ProcessExportJob_ExportError(t *testing.T) {
	logger := zap.NewNop()
	mockExportSvc := &MockExportService{}
	jobManager := NewExportJobManager(logger, mockExportSvc)

	// Mock the validation and export methods
	mockExportSvc.On("ValidateExportRequest", mock.Anything).Return(nil)
	mockExportSvc.On("ExportRiskAssessments", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	request := &ExportRequest{
		BusinessID: "test-business-123",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	job, err := jobManager.CreateExportJob(ctx, request)
	assert.NoError(t, err)

	// Wait a bit for the background job to complete
	time.Sleep(200 * time.Millisecond)

	// Verify the job failed
	retrievedJob, err := jobManager.GetExportJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, "failed", retrievedJob.Status)
	assert.Equal(t, 100, retrievedJob.Progress)
	assert.NotEmpty(t, retrievedJob.Error)

	mockExportSvc.AssertExpectations(t)
}
