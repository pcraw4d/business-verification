package concurrency

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ConcurrentRequestHandler manages concurrent request processing with worker pools
type ConcurrentRequestHandler struct {
	config       *ConcurrentRequestHandlerConfig
	logger       *zap.Logger
	resourceMgr  *ResourceManager
	processor    RequestProcessor
	workerPool   map[string]*Worker
	requestQueue chan *ConcurrentRequest
	responseChan chan *ConcurrentResponse
	stopChan     chan struct{}
	mutex        sync.RWMutex
	stats        *ProcessingStats
	wg           sync.WaitGroup
}

// NewConcurrentRequestHandler creates a new concurrent request handler
func NewConcurrentRequestHandler(
	config *ConcurrentRequestHandlerConfig,
	logger *zap.Logger,
	resourceMgr *ResourceManager,
	processor RequestProcessor,
) *ConcurrentRequestHandler {
	if config == nil {
		config = &ConcurrentRequestHandlerConfig{
			MaxWorkers:    10,
			WorkerTimeout: 30 * time.Second,
			QueueSize:     1000,
		}
	}

	handler := &ConcurrentRequestHandler{
		config:       config,
		logger:       logger,
		resourceMgr:  resourceMgr,
		processor:    processor,
		workerPool:   make(map[string]*Worker),
		requestQueue: make(chan *ConcurrentRequest, config.QueueSize),
		responseChan: make(chan *ConcurrentResponse, config.QueueSize),
		stopChan:     make(chan struct{}),
		stats: &ProcessingStats{
			LastProcessed: time.Now(),
		},
	}

	return handler
}

// Start starts the request handler and worker pool
func (h *ConcurrentRequestHandler) Start() error {
	h.logger.Info("starting concurrent request handler",
		zap.Int("max_workers", h.config.MaxWorkers),
		zap.Int("queue_size", h.config.QueueSize))

	// Start worker goroutines
	for i := 0; i < h.config.MaxWorkers; i++ {
		workerID := fmt.Sprintf("worker-%d", i)
		h.startWorker(workerID)
	}

	// Start response processor
	h.wg.Add(1)
	go h.processResponses()

	h.logger.Info("concurrent request handler started")
	return nil
}

// Stop stops the request handler and all workers
func (h *ConcurrentRequestHandler) Stop() error {
	h.logger.Info("stopping concurrent request handler")

	close(h.stopChan)
	close(h.requestQueue)
	close(h.responseChan)

	// Wait for all workers to finish
	h.wg.Wait()

	h.logger.Info("concurrent request handler stopped")
	return nil
}

// ProcessRequest processes a request asynchronously
func (h *ConcurrentRequestHandler) ProcessRequest(ctx context.Context, request *ConcurrentRequest) (*ConcurrentResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Set request metadata
	if request.CreatedAt.IsZero() {
		request.CreatedAt = time.Now()
	}

	if request.Timeout == 0 {
		request.Timeout = h.config.WorkerTimeout
	}

	// Send request to queue
	select {
	case h.requestQueue <- request:
		h.logger.Debug("request queued for processing",
			zap.String("request_id", request.ID),
			zap.String("type", request.Type))

		// Wait for response
		select {
		case response := <-h.responseChan:
			if response.RequestID == request.ID {
				return response, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(request.Timeout):
			return nil, fmt.Errorf("request processing timeout")
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(100 * time.Millisecond): // Queue timeout
		return nil, fmt.Errorf("request queue full")
	}

	return nil, fmt.Errorf("unexpected error processing request")
}

// startWorker starts a worker goroutine
func (h *ConcurrentRequestHandler) startWorker(workerID string) {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()

		worker := &Worker{
			ID:           workerID,
			Status:       "idle",
			StartedAt:    time.Now(),
			LastActivity: time.Now(),
		}

		h.mutex.Lock()
		h.workerPool[workerID] = worker
		h.mutex.Unlock()

		h.logger.Debug("worker started", zap.String("worker_id", workerID))

		for {
			select {
			case request := <-h.requestQueue:
				h.processRequest(worker, request)
			case <-h.stopChan:
				h.logger.Debug("worker stopping", zap.String("worker_id", workerID))
				return
			}
		}
	}()
}

// processRequest processes a single request
func (h *ConcurrentRequestHandler) processRequest(worker *Worker, request *ConcurrentRequest) {
	startTime := time.Now()

	worker.Status = "processing"
	worker.CurrentTask = request.ID
	worker.LastActivity = time.Now()

	h.logger.Debug("worker processing request",
		zap.String("worker_id", worker.ID),
		zap.String("request_id", request.ID))

	// Acquire required resources
	var acquiredResources []*Resource
	if len(request.RequiredResources) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), request.Timeout)
		defer cancel()

		var err error
		acquiredResources, err = h.resourceMgr.Acquire(ctx, request.RequiredResources)
		if err != nil {
			h.sendResponse(&ConcurrentResponse{
				RequestID:      request.ID,
				Status:         "failed",
				Error:          fmt.Errorf("failed to acquire resources: %w", err),
				ProcessingTime: time.Since(startTime),
				CompletedAt:    time.Now(),
			})
			return
		}
		defer h.resourceMgr.Release(acquiredResources)
	}

	// Process the request
	ctx, cancel := context.WithTimeout(context.Background(), request.Timeout)
	defer cancel()

	response, err := h.processor.Process(ctx, request)
	if err != nil {
		response = &ConcurrentResponse{
			RequestID:      request.ID,
			Status:         "failed",
			Error:          err,
			ProcessingTime: time.Since(startTime),
			CompletedAt:    time.Now(),
		}
	} else {
		response.ProcessingTime = time.Since(startTime)
		response.CompletedAt = time.Now()
	}

	// Update worker stats
	worker.Status = "idle"
	worker.CurrentTask = ""
	worker.TaskCount++
	worker.LastActivity = time.Now()

	// Update processing stats
	h.updateStats(response)

	// Send response
	h.sendResponse(response)

	h.logger.Debug("worker completed request",
		zap.String("worker_id", worker.ID),
		zap.String("request_id", request.ID),
		zap.Duration("processing_time", response.ProcessingTime))
}

// sendResponse sends a response to the response channel
func (h *ConcurrentRequestHandler) sendResponse(response *ConcurrentResponse) {
	select {
	case h.responseChan <- response:
		// Response sent successfully
	default:
		h.logger.Warn("response channel full, dropping response",
			zap.String("request_id", response.RequestID))
	}
}

// processResponses processes responses from workers
func (h *ConcurrentRequestHandler) processResponses() {
	defer h.wg.Done()

	for {
		select {
		case response := <-h.responseChan:
			// Process response (e.g., logging, metrics)
			h.logger.Debug("response processed",
				zap.String("request_id", response.RequestID),
				zap.String("status", response.Status),
				zap.Duration("processing_time", response.ProcessingTime))
		case <-h.stopChan:
			return
		}
	}
}

// GetStats returns processing statistics
func (h *ConcurrentRequestHandler) GetStats() *ProcessingStats {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.stats
}

// GetWorkers returns information about all workers
func (h *ConcurrentRequestHandler) GetWorkers() map[string]*Worker {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	result := make(map[string]*Worker)
	for id, worker := range h.workerPool {
		result[id] = worker
	}
	return result
}

// updateStats updates processing statistics
func (h *ConcurrentRequestHandler) updateStats(response *ConcurrentResponse) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.stats.TotalProcessed++
	h.stats.LastProcessed = time.Now()

	if response.Error == nil {
		h.stats.Successful++
	} else {
		h.stats.Failed++
	}

	// Update timing statistics
	if h.stats.TotalProcessed == 1 {
		h.stats.MinTime = response.ProcessingTime
		h.stats.MaxTime = response.ProcessingTime
		h.stats.AverageTime = response.ProcessingTime
	} else {
		if response.ProcessingTime < h.stats.MinTime {
			h.stats.MinTime = response.ProcessingTime
		}
		if response.ProcessingTime > h.stats.MaxTime {
			h.stats.MaxTime = response.ProcessingTime
		}

		// Calculate running average
		totalTime := h.stats.AverageTime * time.Duration(h.stats.TotalProcessed-1)
		totalTime += response.ProcessingTime
		h.stats.AverageTime = totalTime / time.Duration(h.stats.TotalProcessed)
	}
}

// GetQueueSize returns the current queue size
func (h *ConcurrentRequestHandler) GetQueueSize() int {
	return len(h.requestQueue)
}

// GetActiveWorkers returns the number of active workers
func (h *ConcurrentRequestHandler) GetActiveWorkers() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	active := 0
	for _, worker := range h.workerPool {
		if worker.Status == "processing" {
			active++
		}
	}
	return active
}
