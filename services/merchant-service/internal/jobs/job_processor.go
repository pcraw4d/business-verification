package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// JobProcessor manages a pool of workers to process analysis jobs
type JobProcessor struct {
	workers    int
	jobQueue   chan Job
	workerPool chan chan Job
	quit       chan bool
	wg         sync.WaitGroup
	logger     *zap.Logger
	running    bool
	mu         sync.RWMutex
}

// NewJobProcessor creates a new job processor
func NewJobProcessor(workers int, queueSize int, logger *zap.Logger) *JobProcessor {
	return &JobProcessor{
		workers:    workers,
		jobQueue:   make(chan Job, queueSize),
		workerPool: make(chan chan Job, workers),
		quit:       make(chan bool),
		logger:     logger,
		running:    false,
	}
}

// Start starts the job processor and worker pool
func (jp *JobProcessor) Start() {
	jp.mu.Lock()
	if jp.running {
		jp.mu.Unlock()
		return
	}
	jp.running = true
	jp.mu.Unlock()

	jp.logger.Info("Starting job processor",
		zap.Int("workers", jp.workers),
		zap.Int("queue_size", cap(jp.jobQueue)))

	// Start worker goroutines
	for i := 0; i < jp.workers; i++ {
		jp.wg.Add(1)
		go jp.worker(i)
	}

	// Start dispatcher
	go jp.dispatch()
}

// Stop stops the job processor gracefully
func (jp *JobProcessor) Stop() {
	jp.mu.Lock()
	if !jp.running {
		jp.mu.Unlock()
		return
	}
	jp.running = false
	jp.mu.Unlock()

	jp.logger.Info("Stopping job processor")

	// Signal dispatcher to quit
	select {
	case jp.quit <- true:
	default:
		// Quit channel might be full, but that's OK
	}

	// Signal all workers to quit
	for i := 0; i < jp.workers; i++ {
		select {
		case jp.quit <- true:
		default:
			// Quit channel might be full, but that's OK
		}
	}

	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		jp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		jp.logger.Info("Job processor stopped")
	case <-time.After(5 * time.Second):
		jp.logger.Warn("Job processor stop timed out, some workers may still be running")
	}
}

// Enqueue adds a job to the processing queue
func (jp *JobProcessor) Enqueue(job Job) error {
	jp.mu.RLock()
	running := jp.running
	jp.mu.RUnlock()

	if !running {
		return fmt.Errorf("job processor is not running")
	}

	select {
	case jp.jobQueue <- job:
		jp.logger.Info("Job enqueued",
			zap.String("job_id", job.GetID()),
			zap.String("job_type", job.GetType()),
			zap.String("merchant_id", job.GetMerchantID()))
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("job queue is full, unable to enqueue job")
	}
}

// dispatch dispatches jobs from the queue to available workers
func (jp *JobProcessor) dispatch() {
	for {
		select {
		case job, ok := <-jp.jobQueue:
			if !ok {
				// Queue is closed, exit
				return
			}
			// Get an available worker channel
			select {
			case workerChannel := <-jp.workerPool:
				// Send job to worker
				select {
				case workerChannel <- job:
					// Successfully sent
				case <-jp.quit:
					// Put job back in queue if we're quitting
					select {
					case jp.jobQueue <- job:
					default:
						// Queue is full, job will be lost (acceptable during shutdown)
					}
					return
				}
			case <-jp.quit:
				// Put job back in queue if we're quitting
				select {
				case jp.jobQueue <- job:
				default:
					// Queue is full, job will be lost (acceptable during shutdown)
				}
				return
			}
		case <-jp.quit:
			return
		}
	}
}

// worker processes jobs from the queue
func (jp *JobProcessor) worker(workerID int) {
	defer jp.wg.Done()

	// Create worker's own job channel
	workerChannel := make(chan Job)
	
	// Register worker channel in pool
	select {
	case jp.workerPool <- workerChannel:
		// Successfully registered
	case <-jp.quit:
		jp.logger.Info("Worker stopping before starting", zap.Int("worker_id", workerID))
		return
	}

	jp.logger.Info("Worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case job := <-workerChannel:
			// Process the job
			jp.processJob(context.Background(), job, workerID)
			
			// Re-register worker channel for next job
			select {
			case jp.workerPool <- workerChannel:
				// Successfully re-registered
			case <-jp.quit:
				jp.logger.Info("Worker stopping", zap.Int("worker_id", workerID))
				return
			}

		case <-jp.quit:
			jp.logger.Info("Worker stopping", zap.Int("worker_id", workerID))
			return
		}
	}
}

// processJob processes a single job
func (jp *JobProcessor) processJob(ctx context.Context, job Job, workerID int) {
	if job == nil {
		jp.logger.Warn("Received nil job, skipping",
			zap.Int("worker_id", workerID))
		return
	}

	startTime := time.Now()
	
	jp.logger.Info("Processing job",
		zap.Int("worker_id", workerID),
		zap.String("job_id", job.GetID()),
		zap.String("job_type", job.GetType()),
		zap.String("merchant_id", job.GetMerchantID()))

	// Create context with timeout
	jobCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Process the job
	err := job.Process(jobCtx)
	duration := time.Since(startTime)

	if err != nil {
		jp.logger.Error("Job processing failed",
			zap.Int("worker_id", workerID),
			zap.String("job_id", job.GetID()),
			zap.String("job_type", job.GetType()),
			zap.String("merchant_id", job.GetMerchantID()),
			zap.Duration("duration", duration),
			zap.Error(err))
	} else {
		jp.logger.Info("Job processed successfully",
			zap.Int("worker_id", workerID),
			zap.String("job_id", job.GetID()),
			zap.String("job_type", job.GetType()),
			zap.String("merchant_id", job.GetMerchantID()),
			zap.Duration("duration", duration))
	}
}

// GetQueueSize returns the current queue size
func (jp *JobProcessor) GetQueueSize() int {
	return len(jp.jobQueue)
}

// IsRunning returns whether the processor is running
func (jp *JobProcessor) IsRunning() bool {
	jp.mu.RLock()
	defer jp.mu.RUnlock()
	return jp.running
}

