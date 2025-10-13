package batch

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// BatchJobRepository defines the interface for batch job data access
type BatchJobRepository interface {
	// SaveBatchJob saves a batch job
	SaveBatchJob(ctx context.Context, job *BatchJob) error

	// GetBatchJob retrieves a batch job by ID
	GetBatchJob(ctx context.Context, tenantID, jobID string) (*BatchJob, error)

	// ListBatchJobs lists batch jobs with filters
	ListBatchJobs(ctx context.Context, filter *BatchJobFilter) ([]*BatchJob, error)

	// UpdateBatchJob updates a batch job
	UpdateBatchJob(ctx context.Context, job *BatchJob) error

	// DeleteBatchJob deletes a batch job
	DeleteBatchJob(ctx context.Context, tenantID, jobID string) error

	// GetBatchJobResults gets results for a batch job
	GetBatchJobResults(ctx context.Context, tenantID, jobID string) ([]*BatchResult, error)

	// SaveBatchResult saves a batch result
	SaveBatchResult(ctx context.Context, result *BatchResult) error

	// GetBatchJobMetrics gets metrics for batch jobs
	GetBatchJobMetrics(ctx context.Context, tenantID string, startDate, endDate time.Time) (*BatchJobMetrics, error)

	// GetActiveBatchJobs gets all active batch jobs
	GetActiveBatchJobs(ctx context.Context) ([]*BatchJob, error)

	// CleanupOldJobs removes old completed jobs
	CleanupOldJobs(ctx context.Context, olderThan time.Time) error
}

// SQLBatchJobRepository implements BatchJobRepository using SQL database
type SQLBatchJobRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewSQLBatchJobRepository creates a new SQL batch job repository
func NewSQLBatchJobRepository(db *sql.DB, logger *zap.Logger) *SQLBatchJobRepository {
	return &SQLBatchJobRepository{
		db:     db,
		logger: logger,
	}
}

// SaveBatchJob saves a batch job to the database
func (r *SQLBatchJobRepository) SaveBatchJob(ctx context.Context, job *BatchJob) error {
	r.logger.Info("Saving batch job",
		zap.String("job_id", job.ID),
		zap.String("tenant_id", job.TenantID),
		zap.String("status", string(job.Status)))

	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(job.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Convert results to JSON
	resultsJSON, err := json.Marshal(job.Results)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	// Check if job exists
	existingJob, err := r.GetBatchJob(ctx, job.TenantID, job.ID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing job: %w", err)
	}

	if existingJob != nil {
		// Update existing job
		query := `
			UPDATE batch_jobs SET
				status = $1, total_requests = $2, completed = $3, failed = $4,
				progress = $5, started_at = $6, completed_at = $7, last_updated_at = $8,
				priority = $9, metadata = $10, results = $11, error = $12,
				retry_count = $13, max_retries = $14
			WHERE id = $15 AND tenant_id = $16
		`
		_, err = r.db.ExecContext(ctx, query,
			job.Status, job.TotalRequests, job.Completed, job.Failed,
			job.Progress, job.StartedAt, job.CompletedAt, job.LastUpdatedAt,
			job.Priority, metadataJSON, resultsJSON, job.Error,
			job.RetryCount, job.MaxRetries, job.ID, job.TenantID)
	} else {
		// Insert new job
		query := `
			INSERT INTO batch_jobs (
				id, tenant_id, status, job_type, total_requests, completed, failed,
				progress, created_at, started_at, completed_at, last_updated_at,
				created_by, priority, metadata, results, error, retry_count, max_retries
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
			)
		`
		_, err = r.db.ExecContext(ctx, query,
			job.ID, job.TenantID, job.Status, job.JobType, job.TotalRequests,
			job.Completed, job.Failed, job.Progress, job.CreatedAt, job.StartedAt,
			job.CompletedAt, job.LastUpdatedAt, job.CreatedBy, job.Priority,
			metadataJSON, resultsJSON, job.Error, job.RetryCount, job.MaxRetries)
	}

	if err != nil {
		return fmt.Errorf("failed to save batch job: %w", err)
	}

	r.logger.Info("Batch job saved successfully",
		zap.String("job_id", job.ID),
		zap.String("tenant_id", job.TenantID))

	return nil
}

// GetBatchJob retrieves a batch job by ID
func (r *SQLBatchJobRepository) GetBatchJob(ctx context.Context, tenantID, jobID string) (*BatchJob, error) {
	r.logger.Debug("Getting batch job",
		zap.String("job_id", jobID),
		zap.String("tenant_id", tenantID))

	query := `
		SELECT id, tenant_id, status, job_type, total_requests, completed, failed,
			   progress, created_at, started_at, completed_at, last_updated_at,
			   created_by, priority, metadata, results, error, retry_count, max_retries
		FROM batch_jobs
		WHERE id = $1 AND tenant_id = $2
	`

	var result struct {
		ID            string     `json:"id"`
		TenantID      string     `json:"tenant_id"`
		Status        string     `json:"status"`
		JobType       string     `json:"job_type"`
		TotalRequests int        `json:"total_requests"`
		Completed     int        `json:"completed"`
		Failed        int        `json:"failed"`
		Progress      float64    `json:"progress"`
		CreatedAt     time.Time  `json:"created_at"`
		StartedAt     *time.Time `json:"started_at"`
		CompletedAt   *time.Time `json:"completed_at"`
		LastUpdatedAt time.Time  `json:"last_updated_at"`
		CreatedBy     string     `json:"created_by"`
		Priority      int        `json:"priority"`
		Metadata      string     `json:"metadata"`
		Results       string     `json:"results"`
		Error         string     `json:"error"`
		RetryCount    int        `json:"retry_count"`
		MaxRetries    int        `json:"max_retries"`
	}

	err := r.db.QueryRowContext(ctx, query, jobID, tenantID).Scan(
		&result.ID, &result.TenantID, &result.Status, &result.JobType,
		&result.TotalRequests, &result.Completed, &result.Failed, &result.Progress,
		&result.CreatedAt, &result.StartedAt, &result.CompletedAt, &result.LastUpdatedAt,
		&result.CreatedBy, &result.Priority, &result.Metadata, &result.Results,
		&result.Error, &result.RetryCount, &result.MaxRetries)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Job not found
		}
		return nil, fmt.Errorf("failed to get batch job: %w", err)
	}

	// Convert the result to BatchJob
	job, err := r.convertToBatchJob(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result: %w", err)
	}

	r.logger.Debug("Batch job retrieved successfully",
		zap.String("job_id", jobID),
		zap.String("tenant_id", tenantID))

	return job, nil
}

// ListBatchJobs lists batch jobs with filters
func (r *SQLBatchJobRepository) ListBatchJobs(ctx context.Context, filter *BatchJobFilter) ([]*BatchJob, error) {
	r.logger.Debug("Listing batch jobs",
		zap.String("tenant_id", filter.TenantID),
		zap.String("status", string(filter.Status)))

	// Build query with filters
	query := `
		SELECT id, tenant_id, status, job_type, total_requests, completed, failed,
			   progress, created_at, started_at, completed_at, last_updated_at,
			   created_by, priority, metadata, results, error, retry_count, max_retries
		FROM batch_jobs
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if filter.TenantID != "" {
		query += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, filter.TenantID)
		argIndex++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, string(filter.Status))
		argIndex++
	}

	if filter.JobType != "" {
		query += fmt.Sprintf(" AND job_type = $%d", argIndex)
		args = append(args, filter.JobType)
		argIndex++
	}

	if filter.CreatedBy != "" {
		query += fmt.Sprintf(" AND created_by = $%d", argIndex)
		args = append(args, filter.CreatedBy)
		argIndex++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list batch jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*BatchJob
	for rows.Next() {
		var result struct {
			ID            string     `json:"id"`
			TenantID      string     `json:"tenant_id"`
			Status        string     `json:"status"`
			JobType       string     `json:"job_type"`
			TotalRequests int        `json:"total_requests"`
			Completed     int        `json:"completed"`
			Failed        int        `json:"failed"`
			Progress      float64    `json:"progress"`
			CreatedAt     time.Time  `json:"created_at"`
			StartedAt     *time.Time `json:"started_at"`
			CompletedAt   *time.Time `json:"completed_at"`
			LastUpdatedAt time.Time  `json:"last_updated_at"`
			CreatedBy     string     `json:"created_by"`
			Priority      int        `json:"priority"`
			Metadata      string     `json:"metadata"`
			Results       string     `json:"results"`
			Error         string     `json:"error"`
			RetryCount    int        `json:"retry_count"`
			MaxRetries    int        `json:"max_retries"`
		}

		err := rows.Scan(
			&result.ID, &result.TenantID, &result.Status, &result.JobType,
			&result.TotalRequests, &result.Completed, &result.Failed, &result.Progress,
			&result.CreatedAt, &result.StartedAt, &result.CompletedAt, &result.LastUpdatedAt,
			&result.CreatedBy, &result.Priority, &result.Metadata, &result.Results,
			&result.Error, &result.RetryCount, &result.MaxRetries)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		job, err := r.convertToBatchJob(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert result: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	r.logger.Debug("Batch jobs listed successfully",
		zap.String("tenant_id", filter.TenantID),
		zap.Int("count", len(jobs)))

	return jobs, nil
}

// UpdateBatchJob updates a batch job
func (r *SQLBatchJobRepository) UpdateBatchJob(ctx context.Context, job *BatchJob) error {
	return r.SaveBatchJob(ctx, job) // SaveBatchJob handles both insert and update
}

// DeleteBatchJob deletes a batch job
func (r *SQLBatchJobRepository) DeleteBatchJob(ctx context.Context, tenantID, jobID string) error {
	r.logger.Info("Deleting batch job",
		zap.String("job_id", jobID),
		zap.String("tenant_id", tenantID))

	// Delete batch results first
	_, err := r.db.ExecContext(ctx, "DELETE FROM batch_results WHERE job_id = $1", jobID)
	if err != nil {
		return fmt.Errorf("failed to delete batch results: %w", err)
	}

	// Delete batch job
	_, err = r.db.ExecContext(ctx, "DELETE FROM batch_jobs WHERE id = $1 AND tenant_id = $2", jobID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete batch job: %w", err)
	}

	r.logger.Info("Batch job deleted successfully",
		zap.String("job_id", jobID),
		zap.String("tenant_id", tenantID))

	return nil
}

// GetBatchJobResults gets results for a batch job
func (r *SQLBatchJobRepository) GetBatchJobResults(ctx context.Context, tenantID, jobID string) ([]*BatchResult, error) {
	r.logger.Debug("Getting batch job results",
		zap.String("job_id", jobID),
		zap.String("tenant_id", tenantID))

	query := `
		SELECT id, job_id, request_index, status, request, response, error, processed_at, duration
		FROM batch_results
		WHERE job_id = $1
		ORDER BY request_index
	`

	rows, err := r.db.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch results: %w", err)
	}
	defer rows.Close()

	var results []*BatchResult
	for rows.Next() {
		var result struct {
			ID           string    `json:"id"`
			JobID        string    `json:"job_id"`
			RequestIndex int       `json:"request_index"`
			Status       string    `json:"status"`
			Request      string    `json:"request"`
			Response     string    `json:"response"`
			Error        string    `json:"error"`
			ProcessedAt  time.Time `json:"processed_at"`
			Duration     int64     `json:"duration"` // Duration in nanoseconds
		}

		err := rows.Scan(
			&result.ID, &result.JobID, &result.RequestIndex, &result.Status,
			&result.Request, &result.Response, &result.Error, &result.ProcessedAt, &result.Duration)
		if err != nil {
			return nil, fmt.Errorf("failed to scan result row: %w", err)
		}

		// Parse JSON fields
		var request, response map[string]interface{}
		if result.Request != "" {
			if err := json.Unmarshal([]byte(result.Request), &request); err != nil {
				return nil, fmt.Errorf("failed to unmarshal request: %w", err)
			}
		}
		if result.Response != "" {
			if err := json.Unmarshal([]byte(result.Response), &response); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}
		}

		batchResult := &BatchResult{
			ID:           result.ID,
			JobID:        result.JobID,
			RequestIndex: result.RequestIndex,
			Status:       result.Status,
			Request:      request,
			Response:     response,
			Error:        result.Error,
			ProcessedAt:  result.ProcessedAt,
			Duration:     time.Duration(result.Duration),
		}

		results = append(results, batchResult)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	r.logger.Debug("Batch job results retrieved successfully",
		zap.String("job_id", jobID),
		zap.Int("count", len(results)))

	return results, nil
}

// SaveBatchResult saves a batch result
func (r *SQLBatchJobRepository) SaveBatchResult(ctx context.Context, result *BatchResult) error {
	// Convert request and response to JSON
	requestJSON, err := json.Marshal(result.Request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	responseJSON, err := json.Marshal(result.Response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	query := `
		INSERT INTO batch_results (
			id, job_id, request_index, status, request, response, error, processed_at, duration
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		ON CONFLICT (job_id, request_index) DO UPDATE SET
			status = EXCLUDED.status,
			response = EXCLUDED.response,
			error = EXCLUDED.error,
			processed_at = EXCLUDED.processed_at,
			duration = EXCLUDED.duration
	`

	_, err = r.db.ExecContext(ctx, query,
		result.ID, result.JobID, result.RequestIndex, result.Status,
		requestJSON, responseJSON, result.Error, result.ProcessedAt, result.Duration)

	if err != nil {
		return fmt.Errorf("failed to save batch result: %w", err)
	}

	return nil
}

// GetBatchJobMetrics gets metrics for batch jobs
func (r *SQLBatchJobRepository) GetBatchJobMetrics(ctx context.Context, tenantID string, startDate, endDate time.Time) (*BatchJobMetrics, error) {
	query := `
		SELECT 
			COUNT(*) as total_jobs,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_jobs,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_jobs,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_jobs,
			COUNT(CASE WHEN status = 'processing' THEN 1 END) as processing_jobs,
			AVG(EXTRACT(EPOCH FROM (completed_at - started_at))/60) as avg_duration_minutes,
			(COUNT(CASE WHEN status = 'completed' THEN 1 END)::float / NULLIF(COUNT(*), 0)) * 100 as success_rate
		FROM batch_jobs
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
	`

	var metrics BatchJobMetrics
	err := r.db.QueryRowContext(ctx, query, tenantID, startDate, endDate).Scan(
		&metrics.TotalJobs, &metrics.CompletedJobs, &metrics.FailedJobs,
		&metrics.PendingJobs, &metrics.ProcessingJobs, &metrics.AverageDuration,
		&metrics.SuccessRate)

	if err != nil {
		return nil, fmt.Errorf("failed to get batch job metrics: %w", err)
	}

	// Calculate throughput (requests per minute)
	throughputQuery := `
		SELECT COALESCE(SUM(total_requests), 0) as total_requests
		FROM batch_jobs
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3 AND status = 'completed'
	`

	var totalRequests int
	err = r.db.QueryRowContext(ctx, throughputQuery, tenantID, startDate, endDate).Scan(&totalRequests)
	if err != nil {
		return nil, fmt.Errorf("failed to get throughput: %w", err)
	}

	durationMinutes := endDate.Sub(startDate).Minutes()
	if durationMinutes > 0 {
		metrics.ThroughputPerMin = float64(totalRequests) / durationMinutes
	}

	return &metrics, nil
}

// GetActiveBatchJobs gets all active batch jobs
func (r *SQLBatchJobRepository) GetActiveBatchJobs(ctx context.Context) ([]*BatchJob, error) {
	filter := &BatchJobFilter{
		Status: JobStatusProcessing,
		Limit:  1000, // Reasonable limit for active jobs
	}
	return r.ListBatchJobs(ctx, filter)
}

// CleanupOldJobs removes old completed jobs
func (r *SQLBatchJobRepository) CleanupOldJobs(ctx context.Context, olderThan time.Time) error {
	r.logger.Info("Cleaning up old batch jobs", zap.Time("older_than", olderThan))

	// Delete batch results first
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM batch_results 
		WHERE job_id IN (
			SELECT id FROM batch_jobs 
			WHERE completed_at < $1 AND status IN ('completed', 'failed', 'cancelled')
		)
	`, olderThan)
	if err != nil {
		return fmt.Errorf("failed to delete old batch results: %w", err)
	}

	// Delete batch jobs
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM batch_jobs 
		WHERE completed_at < $1 AND status IN ('completed', 'failed', 'cancelled')
	`, olderThan)
	if err != nil {
		return fmt.Errorf("failed to delete old batch jobs: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	r.logger.Info("Old batch jobs cleaned up successfully", zap.Int64("rows_affected", rowsAffected))

	return nil
}

// convertToBatchJob converts a database result to BatchJob
func (r *SQLBatchJobRepository) convertToBatchJob(result interface{}) (*BatchJob, error) {
	// Type assertion to get the fields
	var id, tenantID, status, jobType, metadata, results, error string
	var totalRequests, completed, failed, priority, retryCount, maxRetries int
	var progress float64
	var createdAt, lastUpdatedAt time.Time
	var startedAt, completedAt *time.Time
	var createdBy string

	// Use reflection or type assertion based on the actual structure
	switch v := result.(type) {
	case *struct {
		ID            string     `json:"id"`
		TenantID      string     `json:"tenant_id"`
		Status        string     `json:"status"`
		JobType       string     `json:"job_type"`
		TotalRequests int        `json:"total_requests"`
		Completed     int        `json:"completed"`
		Failed        int        `json:"failed"`
		Progress      float64    `json:"progress"`
		CreatedAt     time.Time  `json:"created_at"`
		StartedAt     *time.Time `json:"started_at"`
		CompletedAt   *time.Time `json:"completed_at"`
		LastUpdatedAt time.Time  `json:"last_updated_at"`
		CreatedBy     string     `json:"created_by"`
		Priority      int        `json:"priority"`
		Metadata      string     `json:"metadata"`
		Results       string     `json:"results"`
		Error         string     `json:"error"`
		RetryCount    int        `json:"retry_count"`
		MaxRetries    int        `json:"max_retries"`
	}:
		id = v.ID
		tenantID = v.TenantID
		status = v.Status
		jobType = v.JobType
		totalRequests = v.TotalRequests
		completed = v.Completed
		failed = v.Failed
		progress = v.Progress
		createdAt = v.CreatedAt
		startedAt = v.StartedAt
		completedAt = v.CompletedAt
		lastUpdatedAt = v.LastUpdatedAt
		createdBy = v.CreatedBy
		priority = v.Priority
		metadata = v.Metadata
		results = v.Results
		error = v.Error
		retryCount = v.RetryCount
		maxRetries = v.MaxRetries
	default:
		return nil, fmt.Errorf("unsupported result type: %T", result)
	}

	// Parse metadata
	var metadataMap map[string]interface{}
	if metadata != "" {
		if err := json.Unmarshal([]byte(metadata), &metadataMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Parse results
	var resultsList []BatchResult
	if results != "" {
		if err := json.Unmarshal([]byte(results), &resultsList); err != nil {
			return nil, fmt.Errorf("failed to unmarshal results: %w", err)
		}
	}

	job := &BatchJob{
		ID:            id,
		TenantID:      tenantID,
		Status:        JobStatus(status),
		JobType:       jobType,
		TotalRequests: totalRequests,
		Completed:     completed,
		Failed:        failed,
		Progress:      progress,
		CreatedAt:     createdAt,
		StartedAt:     startedAt,
		CompletedAt:   completedAt,
		LastUpdatedAt: lastUpdatedAt,
		CreatedBy:     createdBy,
		Priority:      priority,
		Metadata:      metadataMap,
		Results:       resultsList,
		Error:         error,
		RetryCount:    retryCount,
		MaxRetries:    maxRetries,
	}

	return job, nil
}
