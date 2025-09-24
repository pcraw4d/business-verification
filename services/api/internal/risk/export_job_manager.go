package risk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ExportJobManager manages background export jobs
type ExportJobManager struct {
	logger    *zap.Logger
	jobs      map[string]*ExportJob
	jobsMutex sync.RWMutex
	exportSvc *ExportService
}

// NewExportJobManager creates a new export job manager
func NewExportJobManager(logger *zap.Logger, exportSvc *ExportService) *ExportJobManager {
	return &ExportJobManager{
		logger:    logger,
		jobs:      make(map[string]*ExportJob),
		exportSvc: exportSvc,
	}
}

// CreateExportJob creates a new export job
func (ejm *ExportJobManager) CreateExportJob(ctx context.Context, request *ExportRequest) (*ExportJob, error) {
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Validate the export request
	if err := ejm.exportSvc.ValidateExportRequest(request); err != nil {
		return nil, fmt.Errorf("invalid export request: %w", err)
	}

	// Create new export job
	job := &ExportJob{
		ID:         fmt.Sprintf("export_job_%d", time.Now().UnixNano()),
		BusinessID: request.BusinessID,
		ExportType: request.ExportType,
		Format:     request.Format,
		Status:     "pending",
		Progress:   0,
		CreatedAt:  time.Now(),
		Metadata:   request.Metadata,
	}

	// Store the job
	ejm.jobsMutex.Lock()
	ejm.jobs[job.ID] = job
	ejm.jobsMutex.Unlock()

	ejm.logger.Info("Export job created",
		zap.String("request_id", requestID.(string)),
		zap.String("job_id", job.ID),
		zap.String("business_id", job.BusinessID),
		zap.String("export_type", string(job.ExportType)),
		zap.String("format", string(job.Format)))

	// Start the job in background
	go ejm.processExportJob(ctx, job)

	return job, nil
}

// GetExportJob retrieves an export job by ID
func (ejm *ExportJobManager) GetExportJob(jobID string) (*ExportJob, error) {
	ejm.jobsMutex.RLock()
	defer ejm.jobsMutex.RUnlock()

	job, exists := ejm.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("export job not found: %s", jobID)
	}

	return job, nil
}

// ListExportJobs lists all export jobs for a business
func (ejm *ExportJobManager) ListExportJobs(businessID string) ([]*ExportJob, error) {
	ejm.jobsMutex.RLock()
	defer ejm.jobsMutex.RUnlock()

	var jobs []*ExportJob
	for _, job := range ejm.jobs {
		if job.BusinessID == businessID {
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}

// CancelExportJob cancels a pending export job
func (ejm *ExportJobManager) CancelExportJob(jobID string) error {
	ejm.jobsMutex.Lock()
	defer ejm.jobsMutex.Unlock()

	job, exists := ejm.jobs[jobID]
	if !exists {
		return fmt.Errorf("export job not found: %s", jobID)
	}

	if job.Status != "pending" {
		return fmt.Errorf("cannot cancel job with status: %s", job.Status)
	}

	job.Status = "cancelled"
	job.CompletedAt = &[]time.Time{time.Now()}[0]

	ejm.logger.Info("Export job cancelled",
		zap.String("job_id", jobID),
		zap.String("business_id", job.BusinessID))

	return nil
}

// CleanupOldJobs removes old completed jobs
func (ejm *ExportJobManager) CleanupOldJobs(olderThan time.Time) error {
	ejm.jobsMutex.Lock()
	defer ejm.jobsMutex.Unlock()

	var jobsToDelete []string
	for jobID, job := range ejm.jobs {
		if job.Status == "completed" && job.CompletedAt != nil && job.CompletedAt.Before(olderThan) {
			jobsToDelete = append(jobsToDelete, jobID)
		}
	}

	for _, jobID := range jobsToDelete {
		delete(ejm.jobs, jobID)
	}

	ejm.logger.Info("Cleaned up old export jobs",
		zap.Int("jobs_deleted", len(jobsToDelete)),
		zap.Time("cutoff_time", olderThan))

	return nil
}

// processExportJob processes an export job in the background
func (ejm *ExportJobManager) processExportJob(ctx context.Context, job *ExportJob) {
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	ejm.logger.Info("Starting export job processing",
		zap.String("request_id", requestID.(string)),
		zap.String("job_id", job.ID),
		zap.String("business_id", job.BusinessID))

	// Update job status to processing
	ejm.updateJobStatus(job, "processing", 10, nil)

	// Simulate data retrieval (in real implementation, this would query the database)
	time.Sleep(100 * time.Millisecond)
	ejm.updateJobStatus(job, "processing", 30, nil)

	// Perform the export based on type
	var response *ExportResponse
	var err error

	switch job.ExportType {
	case ExportTypeAssessments:
		response, err = ejm.exportAssessments(ctx, job)
	case ExportTypeFactors:
		response, err = ejm.exportFactors(ctx, job)
	case ExportTypeTrends:
		response, err = ejm.exportTrends(ctx, job)
	case ExportTypeAlerts:
		response, err = ejm.exportAlerts(ctx, job)
	case ExportTypeReports:
		response, err = ejm.exportReports(ctx, job)
	case ExportTypeAll:
		response, err = ejm.exportAll(ctx, job)
	default:
		err = fmt.Errorf("unsupported export type: %s", job.ExportType)
	}

	ejm.updateJobStatus(job, "processing", 80, nil)

	if err != nil {
		ejm.logger.Error("Export job failed",
			zap.String("request_id", requestID.(string)),
			zap.String("job_id", job.ID),
			zap.String("business_id", job.BusinessID),
			zap.Error(err))

		ejm.updateJobStatus(job, "failed", 100, err)
		return
	}

	// Update job with result
	now := time.Now()
	job.Status = "completed"
	job.Progress = 100
	job.CompletedAt = &now
	job.Result = response

	ejm.logger.Info("Export job completed successfully",
		zap.String("request_id", requestID.(string)),
		zap.String("job_id", job.ID),
		zap.String("business_id", job.BusinessID),
		zap.String("export_id", response.ExportID),
		zap.Int("record_count", response.RecordCount))
}

// updateJobStatus updates the status of an export job
func (ejm *ExportJobManager) updateJobStatus(job *ExportJob, status string, progress int, err error) {
	ejm.jobsMutex.Lock()
	defer ejm.jobsMutex.Unlock()

	job.Status = status
	job.Progress = progress

	if err != nil {
		job.Error = err.Error()
	}

	if status == "processing" {
		job.StartedAt = &[]time.Time{time.Now()}[0]
	}
}

// exportAssessments exports risk assessments
func (ejm *ExportJobManager) exportAssessments(ctx context.Context, job *ExportJob) (*ExportResponse, error) {
	// In a real implementation, this would query the database for assessments
	// For now, we'll create mock data
	assessments := []*RiskAssessment{
		{
			ID:           "mock-assessment-1",
			BusinessID:   job.BusinessID,
			BusinessName: "Mock Business",
			OverallScore: 75.5,
			OverallLevel: RiskLevelHigh,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
			AlertLevel:   RiskLevelMedium,
		},
	}

	return ejm.exportSvc.ExportRiskAssessments(ctx, assessments, job.Format)
}

// exportFactors exports risk factors
func (ejm *ExportJobManager) exportFactors(ctx context.Context, job *ExportJob) (*ExportResponse, error) {
	// In a real implementation, this would query the database for factors
	// For now, we'll create mock data
	factors := []RiskScore{
		{
			FactorID:     "mock-factor-1",
			FactorName:   "Mock Financial Risk",
			Category:     RiskCategoryFinancial,
			Score:        80.0,
			Level:        RiskLevelHigh,
			Confidence:   0.9,
			Explanation:  "Mock financial risk explanation",
			CalculatedAt: time.Now(),
		},
	}

	return ejm.exportSvc.ExportRiskFactors(ctx, factors, job.Format)
}

// exportTrends exports risk trends
func (ejm *ExportJobManager) exportTrends(ctx context.Context, job *ExportJob) (*ExportResponse, error) {
	// In a real implementation, this would query the database for trends
	// For now, we'll create mock data
	trends := []RiskTrend{
		{
			BusinessID:   job.BusinessID,
			Category:     RiskCategoryFinancial,
			Score:        75.0,
			Level:        RiskLevelHigh,
			RecordedAt:   time.Now(),
			ChangeFrom:   5.0,
			ChangePeriod: "1 month",
		},
	}

	return ejm.exportSvc.ExportRiskTrends(ctx, trends, job.Format)
}

// exportAlerts exports risk alerts
func (ejm *ExportJobManager) exportAlerts(ctx context.Context, job *ExportJob) (*ExportResponse, error) {
	// In a real implementation, this would query the database for alerts
	// For now, we'll create mock data
	alerts := []RiskAlert{
		{
			ID:             "mock-alert-1",
			BusinessID:     job.BusinessID,
			RiskFactor:     "mock-risk-factor",
			Level:          RiskLevelHigh,
			Message:        "Mock alert message",
			Score:          85.0,
			Threshold:      80.0,
			TriggeredAt:    time.Now(),
			Acknowledged:   false,
			AcknowledgedAt: nil,
		},
	}

	return ejm.exportSvc.ExportRiskAlerts(ctx, alerts, job.Format)
}

// exportReports exports risk reports
func (ejm *ExportJobManager) exportReports(ctx context.Context, job *ExportJob) (*ExportResponse, error) {
	// In a real implementation, this would generate comprehensive reports
	// For now, we'll create a simple report
	report := map[string]interface{}{
		"business_id":    job.BusinessID,
		"export_type":    job.ExportType,
		"format":         job.Format,
		"generated_at":   time.Now(),
		"report_content": "Mock risk report content",
	}

	// Convert to the requested format
	switch job.Format {
	case ExportFormatJSON:
		return &ExportResponse{
			ExportID:    fmt.Sprintf("report_%s_%d", job.BusinessID, time.Now().Unix()),
			BusinessID:  job.BusinessID,
			ExportType:  ExportTypeReports,
			Format:      job.Format,
			Data:        report,
			RecordCount: 1,
			GeneratedAt: time.Now(),
			ExpiresAt:   time.Now().Add(24 * time.Hour),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported format for reports: %s", job.Format)
	}
}

// exportAll exports all risk data
func (ejm *ExportJobManager) exportAll(ctx context.Context, job *ExportJob) (*ExportResponse, error) {
	// In a real implementation, this would export all types of risk data
	// For now, we'll create a comprehensive export
	allData := map[string]interface{}{
		"business_id":  job.BusinessID,
		"export_type":  job.ExportType,
		"format":       job.Format,
		"generated_at": time.Now(),
		"assessments":  []interface{}{},
		"factors":      []interface{}{},
		"trends":       []interface{}{},
		"alerts":       []interface{}{},
		"summary":      "Comprehensive risk data export",
	}

	// Convert to the requested format
	switch job.Format {
	case ExportFormatJSON:
		return &ExportResponse{
			ExportID:    fmt.Sprintf("all_data_%s_%d", job.BusinessID, time.Now().Unix()),
			BusinessID:  job.BusinessID,
			ExportType:  ExportTypeAll,
			Format:      job.Format,
			Data:        allData,
			RecordCount: 1,
			GeneratedAt: time.Now(),
			ExpiresAt:   time.Now().Add(24 * time.Hour),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported format for all data export: %s", job.Format)
	}
}

// GetJobStatistics returns statistics about export jobs
func (ejm *ExportJobManager) GetJobStatistics() map[string]interface{} {
	ejm.jobsMutex.RLock()
	defer ejm.jobsMutex.RUnlock()

	stats := map[string]interface{}{
		"total_jobs":      len(ejm.jobs),
		"pending_jobs":    0,
		"processing_jobs": 0,
		"completed_jobs":  0,
		"failed_jobs":     0,
		"cancelled_jobs":  0,
	}

	for _, job := range ejm.jobs {
		switch job.Status {
		case "pending":
			stats["pending_jobs"] = stats["pending_jobs"].(int) + 1
		case "processing":
			stats["processing_jobs"] = stats["processing_jobs"].(int) + 1
		case "completed":
			stats["completed_jobs"] = stats["completed_jobs"].(int) + 1
		case "failed":
			stats["failed_jobs"] = stats["failed_jobs"].(int) + 1
		case "cancelled":
			stats["cancelled_jobs"] = stats["cancelled_jobs"].(int) + 1
		}
	}

	return stats
}
