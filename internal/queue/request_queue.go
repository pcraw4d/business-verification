package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RequestPriority represents the priority of a queued request
type RequestPriority int

const (
	PriorityLow RequestPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// QueuedRequest represents a request that failed and is queued for retry
type QueuedRequest struct {
	ID          string
	Type        string
	Data        interface{}
	Priority    RequestPriority
	Attempts    int
	MaxAttempts int
	CreatedAt   time.Time
	NextRetry   time.Time
	Error       string
}

// RequestQueue manages a queue of failed requests for retry
type RequestQueue struct {
	mu          sync.RWMutex
	queue       []*QueuedRequest
	priorityMap map[RequestPriority][]*QueuedRequest
	logger      *zap.Logger
	maxSize     int
	processor   *RequestProcessor
}

// RequestProcessor processes queued requests
type RequestProcessor struct {
	queue      *RequestQueue
	logger     *zap.Logger
	processing bool
	mu         sync.Mutex
}

// NewRequestQueue creates a new request queue
func NewRequestQueue(logger *zap.Logger, maxSize int) *RequestQueue {
	rq := &RequestQueue{
		queue:       make([]*QueuedRequest, 0),
		priorityMap: make(map[RequestPriority][]*QueuedRequest),
		logger:      logger,
		maxSize:     maxSize,
	}
	rq.processor = &RequestProcessor{
		queue:  rq,
		logger: logger,
	}
	return rq
}

// Enqueue adds a request to the queue
func (rq *RequestQueue) Enqueue(ctx context.Context, req *QueuedRequest) error {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	// Check queue size
	if len(rq.queue) >= rq.maxSize {
		return fmt.Errorf("request queue is full (max size: %d)", rq.maxSize)
	}

	// Set defaults
	if req.ID == "" {
		req.ID = fmt.Sprintf("req_%d_%d", time.Now().Unix(), len(rq.queue))
	}
	if req.CreatedAt.IsZero() {
		req.CreatedAt = time.Now()
	}
	if req.NextRetry.IsZero() {
		req.NextRetry = time.Now().Add(1 * time.Minute) // Default retry after 1 minute
	}
	if req.MaxAttempts == 0 {
		req.MaxAttempts = 3
	}

	// Add to queue
	rq.queue = append(rq.queue, req)

	// Add to priority map
	if rq.priorityMap[req.Priority] == nil {
		rq.priorityMap[req.Priority] = make([]*QueuedRequest, 0)
	}
	rq.priorityMap[req.Priority] = append(rq.priorityMap[req.Priority], req)

	rq.logger.Info("Request queued for retry",
		zap.String("request_id", req.ID),
		zap.String("type", req.Type),
		zap.Int("priority", int(req.Priority)),
		zap.Int("queue_size", len(rq.queue)))

	return nil
}

// Dequeue removes and returns the next request to process (highest priority first)
func (rq *RequestQueue) Dequeue(ctx context.Context) (*QueuedRequest, error) {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	// Process by priority (Critical -> High -> Normal -> Low)
	priorities := []RequestPriority{PriorityCritical, PriorityHigh, PriorityNormal, PriorityLow}

	for _, priority := range priorities {
		queue := rq.priorityMap[priority]
		if len(queue) == 0 {
			continue
		}

		// Find first request ready for retry
		for i, req := range queue {
			if time.Now().After(req.NextRetry) {
				// Remove from queue
				rq.queue = rq.removeFromQueue(rq.queue, req.ID)
				rq.priorityMap[priority] = append(queue[:i], queue[i+1:]...)

				return req, nil
			}
		}
	}

	return nil, fmt.Errorf("no requests ready for retry")
}

// removeFromQueue removes a request from the queue by ID
func (rq *RequestQueue) removeFromQueue(queue []*QueuedRequest, id string) []*QueuedRequest {
	for i, req := range queue {
		if req.ID == id {
			return append(queue[:i], queue[i+1:]...)
		}
	}
	return queue
}

// ProcessQueue processes queued requests
func (rq *RequestQueue) ProcessQueue(ctx context.Context, handler func(ctx context.Context, req *QueuedRequest) error) {
	rq.processor.mu.Lock()
	if rq.processor.processing {
		rq.processor.mu.Unlock()
		return
	}
	rq.processor.processing = true
	rq.processor.mu.Unlock()

	defer func() {
		rq.processor.mu.Lock()
		rq.processor.processing = false
		rq.processor.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			req, err := rq.Dequeue(ctx)
			if err != nil {
				// No requests ready, wait a bit
				time.Sleep(5 * time.Second)
				continue
			}

			// Process request
			err = handler(ctx, req)
			if err != nil {
				// Retry failed, re-queue if attempts remaining
				req.Attempts++
				if req.Attempts < req.MaxAttempts {
					// Exponential backoff
					delay := time.Duration(req.Attempts) * time.Minute
					req.NextRetry = time.Now().Add(delay)
					req.Error = err.Error()

					rq.mu.Lock()
					rq.queue = append(rq.queue, req)
					if rq.priorityMap[req.Priority] == nil {
						rq.priorityMap[req.Priority] = make([]*QueuedRequest, 0)
					}
					rq.priorityMap[req.Priority] = append(rq.priorityMap[req.Priority], req)
					rq.mu.Unlock()

					rq.logger.Warn("Request retry failed, re-queued",
						zap.String("request_id", req.ID),
						zap.Int("attempts", req.Attempts),
						zap.Error(err))
				} else {
					rq.logger.Error("Request exceeded max attempts, dropping",
						zap.String("request_id", req.ID),
						zap.Int("attempts", req.Attempts),
						zap.Error(err))
				}
			} else {
				rq.logger.Info("Queued request processed successfully",
					zap.String("request_id", req.ID),
					zap.Int("attempts", req.Attempts))
			}
		}
	}
}

// GetQueueSize returns the current queue size
func (rq *RequestQueue) GetQueueSize() int {
	rq.mu.RLock()
	defer rq.mu.RUnlock()
	return len(rq.queue)
}

// GetQueueStats returns statistics about the queue
func (rq *RequestQueue) GetQueueStats() map[string]interface{} {
	rq.mu.RLock()
	defer rq.mu.RUnlock()

	stats := map[string]interface{}{
		"total_size":  len(rq.queue),
		"by_priority": make(map[string]int),
	}

	for priority, queue := range rq.priorityMap {
		stats["by_priority"].(map[string]int)[priority.String()] = len(queue)
	}

	return stats
}

// String returns a string representation of RequestPriority
func (p RequestPriority) String() string {
	switch p {
	case PriorityCritical:
		return "critical"
	case PriorityHigh:
		return "high"
	case PriorityNormal:
		return "normal"
	case PriorityLow:
		return "low"
	default:
		return "unknown"
	}
}

// MarshalJSON implements json.Marshaler for QueuedRequest
func (req *QueuedRequest) MarshalJSON() ([]byte, error) {
	type Alias QueuedRequest
	return json.Marshal(&struct {
		*Alias
		Priority string `json:"priority"`
	}{
		Alias:    (*Alias)(req),
		Priority: req.Priority.String(),
	})
}
