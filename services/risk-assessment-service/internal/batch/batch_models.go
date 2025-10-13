package batch

import (
	"time"
)

// JobStatus represents the status of a batch job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCancelled  JobStatus = "cancelled"
	JobStatusPaused     JobStatus = "paused"
)

// BatchJob represents an asynchronous batch processing job
type BatchJob struct {
	ID            string                 `json:"id" db:"id"`
	TenantID      string                 `json:"tenant_id" db:"tenant_id"`
	Status        JobStatus              `json:"status" db:"status"`
	JobType       string                 `json:"job_type" db:"job_type"` // "risk_assessment", "compliance_check", "custom_model_test"
	TotalRequests int                    `json:"total_requests" db:"total_requests"`
	Completed     int                    `json:"completed" db:"completed"`
	Failed        int                    `json:"failed" db:"failed"`
	Progress      float64                `json:"progress" db:"progress"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	StartedAt     *time.Time             `json:"started_at" db:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at" db:"completed_at"`
	LastUpdatedAt time.Time              `json:"last_updated_at" db:"last_updated_at"`
	CreatedBy     string                 `json:"created_by" db:"created_by"`
	Priority      int                    `json:"priority" db:"priority"` // Higher number = higher priority
	Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
	Results       []BatchResult          `json:"results" db:"results"`
	Error         string                 `json:"error,omitempty" db:"error"`
	RetryCount    int                    `json:"retry_count" db:"retry_count"`
	MaxRetries    int                    `json:"max_retries" db:"max_retries"`
}

// BatchResult represents the result of a single request within a batch job
type BatchResult struct {
	ID           string                 `json:"id" db:"id"`
	JobID        string                 `json:"job_id" db:"job_id"`
	RequestIndex int                    `json:"request_index" db:"request_index"`
	Status       string                 `json:"status" db:"status"` // "success", "failed", "skipped"
	Request      map[string]interface{} `json:"request" db:"request"`
	Response     map[string]interface{} `json:"response" db:"response"`
	Error        string                 `json:"error,omitempty" db:"error"`
	ProcessedAt  time.Time              `json:"processed_at" db:"processed_at"`
	Duration     time.Duration          `json:"duration" db:"duration"`
}

// BatchJobRequest represents a request to create a batch job
type BatchJobRequest struct {
	JobType    string                   `json:"job_type" validate:"required,oneof=risk_assessment compliance_check custom_model_test"`
	Requests   []map[string]interface{} `json:"requests" validate:"required,min=1,max=10000"`
	Priority   int                      `json:"priority,omitempty" validate:"min=1,max=10"`
	MaxRetries int                      `json:"max_retries,omitempty" validate:"min=0,max=5"`
	Metadata   map[string]interface{}   `json:"metadata,omitempty"`
	CreatedBy  string                   `json:"created_by" validate:"required"`
}

// BatchJobResponse represents the response from creating a batch job
type BatchJobResponse struct {
	ID            string    `json:"id"`
	Status        JobStatus `json:"status"`
	TotalRequests int       `json:"total_requests"`
	CreatedAt     time.Time `json:"created_at"`
	EstimatedTime string    `json:"estimated_time,omitempty"`
}

// BatchJobStatus represents the current status of a batch job
type BatchJobStatus struct {
	ID            string                 `json:"id"`
	Status        JobStatus              `json:"status"`
	Progress      float64                `json:"progress"`
	TotalRequests int                    `json:"total_requests"`
	Completed     int                    `json:"completed"`
	Failed        int                    `json:"failed"`
	StartedAt     *time.Time             `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at"`
	EstimatedTime *time.Time             `json:"estimated_time,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// BatchJobResults represents the results of a completed batch job
type BatchJobResults struct {
	ID           string        `json:"id"`
	Status       JobStatus     `json:"status"`
	TotalResults int           `json:"total_results"`
	Results      []BatchResult `json:"results"`
	Summary      BatchSummary  `json:"summary"`
	DownloadURL  string        `json:"download_url,omitempty"`
}

// BatchSummary provides a summary of batch job results
type BatchSummary struct {
	TotalRequests   int     `json:"total_requests"`
	Successful      int     `json:"successful"`
	Failed          int     `json:"failed"`
	SuccessRate     float64 `json:"success_rate"`
	AverageDuration float64 `json:"average_duration_ms"`
	MinDuration     float64 `json:"min_duration_ms"`
	MaxDuration     float64 `json:"max_duration_ms"`
	TotalDuration   float64 `json:"total_duration_ms"`
}

// BatchJobFilter represents filters for querying batch jobs
type BatchJobFilter struct {
	TenantID  string     `json:"tenant_id,omitempty"`
	Status    JobStatus  `json:"status,omitempty"`
	JobType   string     `json:"job_type,omitempty"`
	CreatedBy string     `json:"created_by,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// BatchJobMetrics represents metrics for batch job performance
type BatchJobMetrics struct {
	TotalJobs        int     `json:"total_jobs"`
	CompletedJobs    int     `json:"completed_jobs"`
	FailedJobs       int     `json:"failed_jobs"`
	PendingJobs      int     `json:"pending_jobs"`
	ProcessingJobs   int     `json:"processing_jobs"`
	AverageDuration  float64 `json:"average_duration_minutes"`
	SuccessRate      float64 `json:"success_rate"`
	ThroughputPerMin float64 `json:"throughput_per_minute"`
}

// BatchJobConfig represents configuration for batch job processing
type BatchJobConfig struct {
	MaxConcurrentJobs      int           `json:"max_concurrent_jobs"`
	MaxRequestsPerJob      int           `json:"max_requests_per_job"`
	DefaultTimeout         time.Duration `json:"default_timeout"`
	RetryInterval          time.Duration `json:"retry_interval"`
	MaxRetryAttempts       int           `json:"max_retry_attempts"`
	ProgressUpdateInterval time.Duration `json:"progress_update_interval"`
	CleanupAfterDays       int           `json:"cleanup_after_days"`
}

// DefaultBatchJobConfig returns the default configuration for batch jobs
func DefaultBatchJobConfig() *BatchJobConfig {
	return &BatchJobConfig{
		MaxConcurrentJobs:      20,               // Increased for better throughput
		MaxRequestsPerJob:      10000,            // Support up to 10,000 requests per job
		DefaultTimeout:         60 * time.Minute, // Increased timeout for large batches
		RetryInterval:          5 * time.Minute,
		MaxRetryAttempts:       3,
		ProgressUpdateInterval: 5 * time.Second, // More frequent progress updates
		CleanupAfterDays:       30,
	}
}

// IsCompleted returns true if the batch job is in a completed state
func (bj *BatchJob) IsCompleted() bool {
	return bj.Status == JobStatusCompleted || bj.Status == JobStatusFailed || bj.Status == JobStatusCancelled
}

// IsActive returns true if the batch job is currently processing
func (bj *BatchJob) IsActive() bool {
	return bj.Status == JobStatusProcessing || bj.Status == JobStatusPending
}

// UpdateProgress updates the progress of the batch job
func (bj *BatchJob) UpdateProgress(completed, failed int) {
	bj.Completed = completed
	bj.Failed = failed
	bj.TotalRequests = completed + failed + (bj.TotalRequests - bj.Completed - bj.Failed)

	if bj.TotalRequests > 0 {
		bj.Progress = float64(completed+failed) / float64(bj.TotalRequests) * 100
	}

	bj.LastUpdatedAt = time.Now()
}

// MarkCompleted marks the batch job as completed
func (bj *BatchJob) MarkCompleted() {
	bj.Status = JobStatusCompleted
	now := time.Now()
	bj.CompletedAt = &now
	bj.Progress = 100.0
	bj.LastUpdatedAt = now
}

// MarkFailed marks the batch job as failed
func (bj *BatchJob) MarkFailed(error string) {
	bj.Status = JobStatusFailed
	now := time.Now()
	bj.CompletedAt = &now
	bj.Error = error
	bj.LastUpdatedAt = now
}

// MarkCancelled marks the batch job as cancelled
func (bj *BatchJob) MarkCancelled() {
	bj.Status = JobStatusCancelled
	now := time.Now()
	bj.CompletedAt = &now
	bj.LastUpdatedAt = now
}

// CanRetry returns true if the batch job can be retried
func (bj *BatchJob) CanRetry() bool {
	return bj.Status == JobStatusFailed && bj.RetryCount < bj.MaxRetries
}

// IncrementRetry increments the retry count
func (bj *BatchJob) IncrementRetry() {
	bj.RetryCount++
	bj.Status = JobStatusPending
	bj.LastUpdatedAt = time.Now()
}

// GetEstimatedCompletionTime returns the estimated completion time
func (bj *BatchJob) GetEstimatedCompletionTime() *time.Time {
	if bj.Status != JobStatusProcessing || bj.Progress == 0 {
		return nil
	}

	if bj.StartedAt == nil {
		return nil
	}

	elapsed := time.Since(*bj.StartedAt)
	remaining := float64(100-bj.Progress) / bj.Progress * float64(elapsed)
	estimated := bj.StartedAt.Add(time.Duration(remaining))

	return &estimated
}
