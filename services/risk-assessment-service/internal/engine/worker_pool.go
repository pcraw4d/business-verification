package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// WorkerPool manages a pool of workers for concurrent processing
type WorkerPool struct {
	workers    int
	jobQueue   chan Job
	workerPool chan chan Job
	quit       chan bool
	wg         sync.WaitGroup
	logger     *zap.Logger
}

// Job represents a job to be processed
type Job struct {
	ID       string
	Function func() (interface{}, error)
	Result   chan JobResult
	Context  context.Context
}

// JobResult represents the result of a job
type JobResult struct {
	ID    string
	Data  interface{}
	Error error
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int, logger *zap.Logger) *WorkerPool {
	return &WorkerPool{
		workers:    workers,
		jobQueue:   make(chan Job, workers*2), // Buffer for job queue
		workerPool: make(chan chan Job, workers),
		quit:       make(chan bool),
		logger:     logger,
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	wp.logger.Info("Starting worker pool", zap.Int("workers", wp.workers))

	for i := 0; i < wp.workers; i++ {
		worker := NewWorker(wp.workerPool, wp.logger)
		worker.Start()
		wp.wg.Add(1)
	}

	go wp.dispatch()
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	wp.logger.Info("Stopping worker pool")

	// Signal all workers to quit
	for i := 0; i < wp.workers; i++ {
		wp.quit <- true
	}

	// Close job queue
	close(wp.jobQueue)

	// Wait for all workers to finish
	wp.wg.Wait()

	wp.logger.Info("Worker pool stopped")
}

// Shutdown gracefully shuts down the worker pool
func (wp *WorkerPool) Shutdown(ctx context.Context) error {
	wp.logger.Info("Shutting down worker pool")

	// Create a channel to signal completion
	done := make(chan struct{})

	go func() {
		wp.Stop()
		close(done)
	}()

	// Wait for shutdown or context cancellation
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Submit submits a job to the worker pool
func (wp *WorkerPool) Submit(job Job) {
	select {
	case wp.jobQueue <- job:
		// Job submitted successfully
	default:
		// Job queue is full, reject the job
		job.Result <- JobResult{
			ID:    job.ID,
			Error: ErrJobQueueFull,
		}
	}
}

// SubmitWithTimeout submits a job with a timeout
func (wp *WorkerPool) SubmitWithTimeout(job Job, timeout time.Duration) error {
	select {
	case wp.jobQueue <- job:
		return nil
	case <-time.After(timeout):
		return ErrJobSubmissionTimeout
	}
}

// dispatch dispatches jobs to available workers
func (wp *WorkerPool) dispatch() {
	for {
		select {
		case job := <-wp.jobQueue:
			// Get an available worker
			workerJobQueue := <-wp.workerPool
			// Submit the job to the worker
			workerJobQueue <- job
		case <-wp.quit:
			return
		}
	}
}

// Worker represents a single worker in the pool
type Worker struct {
	workerPool chan chan Job
	jobChannel chan Job
	quit       chan bool
	logger     *zap.Logger
}

// NewWorker creates a new worker
func NewWorker(workerPool chan chan Job, logger *zap.Logger) *Worker {
	return &Worker{
		workerPool: workerPool,
		jobChannel: make(chan Job),
		quit:       make(chan bool),
		logger:     logger,
	}
}

// Start starts the worker
func (w *Worker) Start() {
	go func() {
		for {
			// Register this worker to the worker pool
			w.workerPool <- w.jobChannel

			select {
			case job := <-w.jobChannel:
				// Process the job
				w.processJob(job)
			case <-w.quit:
				return
			}
		}
	}()
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.quit <- true
}

// processJob processes a job
func (w *Worker) processJob(job Job) {
	start := time.Now()

	// Check if context is cancelled
	select {
	case <-job.Context.Done():
		job.Result <- JobResult{
			ID:    job.ID,
			Error: job.Context.Err(),
		}
		return
	default:
	}

	// Execute the job function
	result, err := job.Function()

	duration := time.Since(start)
	w.logger.Debug("Job processed",
		zap.String("job_id", job.ID),
		zap.Duration("duration", duration),
		zap.Error(err))

	// Send result
	job.Result <- JobResult{
		ID:    job.ID,
		Data:  result,
		Error: err,
	}
}

// BatchProcessor processes multiple jobs concurrently
type BatchProcessor struct {
	pool   *WorkerPool
	logger *zap.Logger
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(pool *WorkerPool, logger *zap.Logger) *BatchProcessor {
	return &BatchProcessor{
		pool:   pool,
		logger: logger,
	}
}

// ProcessBatch processes a batch of jobs concurrently
func (bp *BatchProcessor) ProcessBatch(ctx context.Context, jobs []Job) ([]JobResult, error) {
	if len(jobs) == 0 {
		return []JobResult{}, nil
	}

	start := time.Now()
	results := make([]JobResult, len(jobs))

	// Submit all jobs
	for i, job := range jobs {
		job.ID = fmt.Sprintf("job_%d", i)
		job.Result = make(chan JobResult, 1)
		job.Context = ctx

		if err := bp.pool.SubmitWithTimeout(job, 100*time.Millisecond); err != nil {
			results[i] = JobResult{
				ID:    job.ID,
				Error: err,
			}
			continue
		}
	}

	// Collect results
	for i, job := range jobs {
		select {
		case result := <-job.Result:
			results[i] = result
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	duration := time.Since(start)
	bp.logger.Info("Batch processing completed",
		zap.Int("job_count", len(jobs)),
		zap.Duration("duration", duration))

	return results, nil
}

// Error definitions
var (
	ErrJobQueueFull         = fmt.Errorf("job queue is full")
	ErrJobSubmissionTimeout = fmt.Errorf("job submission timeout")
	ErrWorkerPoolStopped    = fmt.Errorf("worker pool is stopped")
)
