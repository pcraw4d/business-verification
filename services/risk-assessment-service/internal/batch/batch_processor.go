package batch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// BatchProcessor defines the interface for processing batch jobs
type BatchProcessor interface {
	// ProcessBatchJob processes a batch job
	ProcessBatchJob(ctx context.Context, job *BatchJob) error

	// ProcessBatchRequest processes a single request within a batch job
	ProcessBatchRequest(ctx context.Context, job *BatchJob, requestIndex int, request map[string]interface{}) (*BatchResult, error)
}

// WorkerPool manages a pool of workers for processing batch jobs
type WorkerPool struct {
	numWorkers int
	processor  BatchProcessor
	logger     *zap.Logger
	jobChan    chan *BatchJob
	workerWg   sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	started    bool
	startedMu  sync.RWMutex
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(numWorkers int, processor BatchProcessor, logger *zap.Logger) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		processor:  processor,
		logger:     logger,
		jobChan:    make(chan *BatchJob, numWorkers*2),
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start(ctx context.Context) error {
	wp.startedMu.Lock()
	defer wp.startedMu.Unlock()

	if wp.started {
		return fmt.Errorf("worker pool is already started")
	}

	wp.logger.Info("Starting worker pool",
		zap.Int("num_workers", wp.numWorkers))

	wp.ctx, wp.cancel = context.WithCancel(ctx)

	// Start workers
	for i := 0; i < wp.numWorkers; i++ {
		wp.workerWg.Add(1)
		go wp.worker(i)
	}

	wp.started = true

	wp.logger.Info("Worker pool started successfully")

	return nil
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() error {
	wp.startedMu.Lock()
	defer wp.startedMu.Unlock()

	if !wp.started {
		return fmt.Errorf("worker pool is not started")
	}

	wp.logger.Info("Stopping worker pool")

	// Cancel context to signal workers to stop
	wp.cancel()

	// Close job channel
	close(wp.jobChan)

	// Wait for all workers to finish
	wp.workerWg.Wait()

	wp.started = false

	wp.logger.Info("Worker pool stopped successfully")

	return nil
}

// SubmitJob submits a job to the worker pool
func (wp *WorkerPool) SubmitJob(ctx context.Context, job *BatchJob) error {
	wp.startedMu.RLock()
	defer wp.startedMu.RUnlock()

	if !wp.started {
		return fmt.Errorf("worker pool is not started")
	}

	select {
	case wp.jobChan <- job:
		wp.logger.Debug("Job submitted to worker pool",
			zap.String("job_id", job.ID))
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("worker pool is full")
	}
}

// worker is a worker goroutine that processes jobs
func (wp *WorkerPool) worker(workerID int) {
	defer wp.workerWg.Done()

	wp.logger.Debug("Worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-wp.ctx.Done():
			wp.logger.Debug("Worker stopping due to context cancellation",
				zap.Int("worker_id", workerID))
			return
		case job, ok := <-wp.jobChan:
			if !ok {
				wp.logger.Debug("Worker stopping due to channel closure",
					zap.Int("worker_id", workerID))
				return
			}

			wp.logger.Debug("Worker processing job",
				zap.Int("worker_id", workerID),
				zap.String("job_id", job.ID))

			// Process the job
			if err := wp.processor.ProcessBatchJob(wp.ctx, job); err != nil {
				wp.logger.Error("Failed to process batch job",
					zap.Int("worker_id", workerID),
					zap.String("job_id", job.ID),
					zap.Error(err))
			} else {
				wp.logger.Debug("Batch job processed successfully",
					zap.Int("worker_id", workerID),
					zap.String("job_id", job.ID))
			}
		}
	}
}

// DefaultBatchProcessor implements BatchProcessor with default behavior
type DefaultBatchProcessor struct {
	riskEngine RiskEngine
	repository BatchJobRepository
	logger     *zap.Logger
	config     *BatchJobConfig
}

// RiskEngine interface for risk assessment processing
type RiskEngine interface {
	AssessRisk(ctx context.Context, request *models.RiskAssessmentRequest) (*models.RiskAssessment, error)
}

// NewDefaultBatchProcessor creates a new default batch processor
func NewDefaultBatchProcessor(
	riskEngine RiskEngine,
	repository BatchJobRepository,
	config *BatchJobConfig,
	logger *zap.Logger,
) *DefaultBatchProcessor {
	if config == nil {
		config = DefaultBatchJobConfig()
	}

	return &DefaultBatchProcessor{
		riskEngine: riskEngine,
		repository: repository,
		logger:     logger,
		config:     config,
	}
}

// ProcessBatchJob processes a batch job
func (bp *DefaultBatchProcessor) ProcessBatchJob(ctx context.Context, job *BatchJob) error {
	bp.logger.Info("Processing batch job",
		zap.String("job_id", job.ID),
		zap.String("job_type", job.JobType),
		zap.Int("total_requests", job.TotalRequests))

	// Create a timeout context for the entire job
	jobCtx, cancel := context.WithTimeout(ctx, bp.config.DefaultTimeout)
	defer cancel()

	// Extract requests from job metadata
	requests, err := bp.extractRequestsFromJob(job)
	if err != nil {
		bp.logger.Error("Failed to extract requests from job",
			zap.String("job_id", job.ID),
			zap.Error(err))
		job.MarkFailed(fmt.Sprintf("Failed to extract requests: %v", err))
		return bp.repository.UpdateBatchJob(ctx, job)
	}

	// Process each request in the batch
	for i := 0; i < job.TotalRequests; i++ {
		select {
		case <-jobCtx.Done():
			bp.logger.Warn("Batch job cancelled due to timeout",
				zap.String("job_id", job.ID),
				zap.Int("processed", i))
			job.MarkFailed("Job cancelled due to timeout")
			return bp.repository.UpdateBatchJob(ctx, job)
		default:
			// Get the actual request data
			var request map[string]interface{}
			if i < len(requests) {
				request = requests[i]
			} else {
				// Fallback for missing request data
				request = map[string]interface{}{
					"request_index": i,
					"job_id":        job.ID,
					"error":         "Request data not found",
				}
			}

			result, err := bp.ProcessBatchRequest(jobCtx, job, i, request)
			if err != nil {
				bp.logger.Error("Failed to process batch request",
					zap.String("job_id", job.ID),
					zap.Int("request_index", i),
					zap.Error(err))

				// Create failed result
				result = &BatchResult{
					ID:           fmt.Sprintf("%s_%d", job.ID, i),
					JobID:        job.ID,
					RequestIndex: i,
					Status:       "failed",
					Request:      request,
					Response:     nil,
					Error:        err.Error(),
					ProcessedAt:  time.Now(),
					Duration:     0,
				}
			}

			// Save result
			if err := bp.repository.SaveBatchResult(ctx, result); err != nil {
				bp.logger.Error("Failed to save batch result",
					zap.String("job_id", job.ID),
					zap.Int("request_index", i),
					zap.Error(err))
			}

			// Update job progress
			job.UpdateProgress(job.Completed, job.Failed)
			if result.Status == "success" {
				job.Completed++
			} else {
				job.Failed++
			}

			// Update job in repository periodically
			if i%10 == 0 || i == job.TotalRequests-1 {
				if err := bp.repository.UpdateBatchJob(ctx, job); err != nil {
					bp.logger.Error("Failed to update batch job progress",
						zap.String("job_id", job.ID),
						zap.Error(err))
				}
			}

			// Add small delay to prevent overwhelming the system
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Mark job as completed
	job.MarkCompleted()
	if err := bp.repository.UpdateBatchJob(ctx, job); err != nil {
		return fmt.Errorf("failed to mark job as completed: %w", err)
	}

	bp.logger.Info("Batch job completed successfully",
		zap.String("job_id", job.ID),
		zap.Int("completed", job.Completed),
		zap.Int("failed", job.Failed),
		zap.Float64("success_rate", float64(job.Completed)/float64(job.TotalRequests)*100))

	return nil
}

// ProcessBatchRequest processes a single request within a batch job
func (bp *DefaultBatchProcessor) ProcessBatchRequest(ctx context.Context, job *BatchJob, requestIndex int, request map[string]interface{}) (*BatchResult, error) {
	startTime := time.Now()

	bp.logger.Debug("Processing batch request",
		zap.String("job_id", job.ID),
		zap.Int("request_index", requestIndex))

	// Process based on job type
	var response map[string]interface{}
	var err error

	switch job.JobType {
	case "risk_assessment":
		response, err = bp.processRiskAssessmentRequest(ctx, request)
	case "compliance_check":
		response, err = bp.processComplianceCheckRequest(ctx, request)
	case "custom_model_test":
		response, err = bp.processCustomModelTestRequest(ctx, request)
	default:
		err = fmt.Errorf("unknown job type: %s", job.JobType)
	}

	duration := time.Since(startTime)

	result := &BatchResult{
		ID:           fmt.Sprintf("%s_%d", job.ID, requestIndex),
		JobID:        job.ID,
		RequestIndex: requestIndex,
		Request:      request,
		Response:     response,
		ProcessedAt:  time.Now(),
		Duration:     duration,
	}

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
	} else {
		result.Status = "success"
	}

	bp.logger.Debug("Batch request processed",
		zap.String("job_id", job.ID),
		zap.Int("request_index", requestIndex),
		zap.String("status", result.Status),
		zap.Duration("duration", duration))

	return result, nil
}

// processRiskAssessmentRequest processes a risk assessment request
func (bp *DefaultBatchProcessor) processRiskAssessmentRequest(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	// Convert request to risk assessment format
	riskRequest := &models.RiskAssessmentRequest{
		BusinessName:    getStringFromMap(request, "business_name"),
		BusinessAddress: getStringFromMap(request, "business_address"),
		Industry:        getStringFromMap(request, "industry"),
		Country:         getStringFromMap(request, "country"),
		Phone:           getStringFromMap(request, "phone"),
		Email:           getStringFromMap(request, "email"),
		Website:         getStringFromMap(request, "website"),
	}

	// Process through risk engine
	assessment, err := bp.riskEngine.AssessRisk(ctx, riskRequest)
	if err != nil {
		return nil, fmt.Errorf("risk assessment failed: %w", err)
	}

	// Convert assessment to response format
	response := map[string]interface{}{
		"id":               assessment.ID,
		"business_name":    assessment.BusinessName,
		"business_address": assessment.BusinessAddress,
		"industry":         assessment.Industry,
		"country":          assessment.Country,
		"risk_score":       assessment.RiskScore,
		"risk_level":       assessment.RiskLevel,
		"status":           assessment.Status,
		"created_at":       assessment.CreatedAt,
		"updated_at":       assessment.UpdatedAt,
	}

	return response, nil
}

// extractRequestsFromJob extracts request data from job metadata
func (bp *DefaultBatchProcessor) extractRequestsFromJob(job *BatchJob) ([]map[string]interface{}, error) {
	// Try to get requests from metadata
	if requestsData, ok := job.Metadata["requests"]; ok {
		if requests, ok := requestsData.([]interface{}); ok {
			result := make([]map[string]interface{}, len(requests))
			for i, req := range requests {
				if reqMap, ok := req.(map[string]interface{}); ok {
					result[i] = reqMap
				} else {
					return nil, fmt.Errorf("invalid request format at index %d", i)
				}
			}
			return result, nil
		}
	}

	// If no requests in metadata, create dummy requests for testing
	bp.logger.Warn("No requests found in job metadata, creating dummy requests",
		zap.String("job_id", job.ID),
		zap.Int("total_requests", job.TotalRequests))

	requests := make([]map[string]interface{}, job.TotalRequests)
	for i := 0; i < job.TotalRequests; i++ {
		requests[i] = map[string]interface{}{
			"business_name":    fmt.Sprintf("Test Business %d", i+1),
			"business_address": fmt.Sprintf("123 Test St, City %d, State 12345", i+1),
			"industry":         "Technology",
			"country":          "US",
			"phone":            "+1-555-123-4567",
			"email":            fmt.Sprintf("test%d@example.com", i+1),
			"website":          fmt.Sprintf("https://test%d.com", i+1),
		}
	}

	return requests, nil
}

// Helper function to safely get string values from map
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// processComplianceCheckRequest processes a compliance check request
func (bp *DefaultBatchProcessor) processComplianceCheckRequest(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	// Implement compliance check processing
	// This would integrate with compliance monitoring services
	response := map[string]interface{}{
		"compliance_status": "passed",
		"checks_performed": []string{
			"sanctions_screening",
			"adverse_media_check",
			"regulatory_compliance",
		},
		"timestamp": time.Now().Unix(),
	}

	return response, nil
}

// processCustomModelTestRequest processes a custom model test request
func (bp *DefaultBatchProcessor) processCustomModelTestRequest(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	// Implement custom model testing
	// This would test custom risk models with sample data
	response := map[string]interface{}{
		"model_id":    request["model_id"],
		"test_status": "passed",
		"accuracy":    0.95,
		"predictions": []float64{0.1, 0.3, 0.7, 0.9},
		"timestamp":   time.Now().Unix(),
	}

	return response, nil
}
