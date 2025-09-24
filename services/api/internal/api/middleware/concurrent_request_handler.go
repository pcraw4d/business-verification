package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RequestQueue represents a queue for handling concurrent requests
type RequestQueue struct {
	queue   chan *QueuedRequest
	workers int
	limiter *rate.Limiter
	mu      sync.RWMutex
	metrics *QueueMetrics
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// QueuedRequest represents a request in the queue
type QueuedRequest struct {
	ID        string
	Request   *http.Request
	Response  http.ResponseWriter
	Handler   http.HandlerFunc
	Priority  int
	CreatedAt time.Time
	Context   context.Context
}

// QueueMetrics tracks queue performance metrics
type QueueMetrics struct {
	mu                 sync.RWMutex
	TotalRequests      int64
	ProcessedRequests  int64
	FailedRequests     int64
	QueueSize          int
	AverageWaitTime    time.Duration
	AverageProcessTime time.Duration
	ActiveWorkers      int
	LastUpdated        time.Time
}

// QueueConfig holds configuration for the request queue
type QueueConfig struct {
	MaxWorkers     int           // Maximum number of concurrent workers
	QueueSize      int           // Maximum queue size
	RequestTimeout time.Duration // Timeout for individual requests
	RateLimit      float64       // Requests per second
	BurstLimit     int           // Burst limit for rate limiting
	EnableMetrics  bool          // Enable detailed metrics collection
	PriorityLevels int           // Number of priority levels (1-10)
}

// DefaultQueueConfig returns default configuration for the request queue
func DefaultQueueConfig() *QueueConfig {
	return &QueueConfig{
		MaxWorkers:     50,               // Support 100+ concurrent users with 50 workers
		QueueSize:      1000,             // Large queue to handle bursts
		RequestTimeout: 30 * time.Second, // 30 second timeout per request
		RateLimit:      100.0,            // 100 requests per second
		BurstLimit:     200,              // Allow bursts up to 200 requests
		EnableMetrics:  true,             // Enable metrics by default
		PriorityLevels: 5,                // 5 priority levels
	}
}

// NewRequestQueue creates a new request queue with the given configuration
func NewRequestQueue(config *QueueConfig) *RequestQueue {
	if config == nil {
		config = DefaultQueueConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	queue := &RequestQueue{
		queue:   make(chan *QueuedRequest, config.QueueSize),
		workers: config.MaxWorkers,
		limiter: rate.NewLimiter(rate.Limit(config.RateLimit), config.BurstLimit),
		metrics: &QueueMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Start worker goroutines
	for i := 0; i < config.MaxWorkers; i++ {
		queue.wg.Add(1)
		go queue.worker(i, config)
	}

	log.Printf("RequestQueue initialized with %d workers, queue size %d, rate limit %.2f req/s",
		config.MaxWorkers, config.QueueSize, config.RateLimit)

	return queue
}

// worker processes requests from the queue
func (rq *RequestQueue) worker(id int, config *QueueConfig) {
	defer rq.wg.Done()

	log.Printf("Worker %d started", id)

	for {
		select {
		case queuedReq := <-rq.queue:
			if queuedReq == nil {
				return // Queue closed
			}

			rq.processRequest(queuedReq, config)

		case <-rq.ctx.Done():
			log.Printf("Worker %d shutting down", id)
			return
		}
	}
}

// processRequest handles a single request from the queue
func (rq *RequestQueue) processRequest(queuedReq *QueuedRequest, config *QueueConfig) {
	startTime := time.Now()

	// Update metrics
	rq.metrics.mu.Lock()
	rq.metrics.ActiveWorkers++
	rq.metrics.ProcessedRequests++
	rq.metrics.LastUpdated = time.Now()
	rq.metrics.mu.Unlock()

	defer func() {
		rq.metrics.mu.Lock()
		rq.metrics.ActiveWorkers--
		rq.metrics.AverageProcessTime = time.Since(startTime)
		rq.metrics.mu.Unlock()
	}()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(queuedReq.Context, config.RequestTimeout)
	defer cancel()

	// Update request context
	queuedReq.Request = queuedReq.Request.WithContext(ctx)

	// Process the request
	queuedReq.Handler(queuedReq.Response, queuedReq.Request)

	// Log processing time
	processTime := time.Since(startTime)
	waitTime := startTime.Sub(queuedReq.CreatedAt)

	log.Printf("Request %s processed in %v (waited %v)",
		queuedReq.ID, processTime, waitTime)
}

// EnqueueRequest adds a request to the queue
func (rq *RequestQueue) EnqueueRequest(req *http.Request, w http.ResponseWriter, handler http.HandlerFunc, priority int) error {
	// Check if queue is full
	select {
	case rq.queue <- &QueuedRequest{
		ID:        generateQueueRequestID(),
		Request:   req,
		Response:  w,
		Handler:   handler,
		Priority:  priority,
		CreatedAt: time.Now(),
		Context:   req.Context(),
	}:
		rq.updateMetrics(1, 0)
		return nil
	default:
		rq.updateMetrics(0, 1)
		return fmt.Errorf("queue is full")
	}
}

// updateMetrics updates queue metrics
func (rq *RequestQueue) updateMetrics(queued, failed int64) {
	rq.metrics.mu.Lock()
	defer rq.metrics.mu.Unlock()

	rq.metrics.TotalRequests += queued
	rq.metrics.FailedRequests += failed
	rq.metrics.QueueSize = len(rq.queue)
	rq.metrics.LastUpdated = time.Now()
}

// GetMetrics returns current queue metrics
func (rq *RequestQueue) GetMetrics() *QueueMetrics {
	rq.metrics.mu.RLock()
	defer rq.metrics.mu.RUnlock()

	// Return a copy to avoid race conditions
	return &QueueMetrics{
		TotalRequests:      rq.metrics.TotalRequests,
		ProcessedRequests:  rq.metrics.ProcessedRequests,
		FailedRequests:     rq.metrics.FailedRequests,
		QueueSize:          rq.metrics.QueueSize,
		AverageWaitTime:    rq.metrics.AverageWaitTime,
		AverageProcessTime: rq.metrics.AverageProcessTime,
		ActiveWorkers:      rq.metrics.ActiveWorkers,
		LastUpdated:        rq.metrics.LastUpdated,
	}
}

// Shutdown gracefully shuts down the queue
func (rq *RequestQueue) Shutdown() {
	log.Println("Shutting down request queue...")
	rq.cancel()
	close(rq.queue)
	rq.wg.Wait()
	log.Println("Request queue shutdown complete")
}

// generateQueueRequestID generates a unique request ID for queue
func generateQueueRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// ConcurrentRequestMiddleware creates middleware for handling concurrent requests
func ConcurrentRequestMiddleware(config *QueueConfig) func(http.HandlerFunc) http.HandlerFunc {
	queue := NewRequestQueue(config)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Check rate limit
			if !queue.limiter.Allow() {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// Determine priority based on request type
			priority := determinePriority(r)

			// Try to enqueue the request
			err := queue.EnqueueRequest(r, w, next, priority)
			if err != nil {
				http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
				return
			}
		}
	}
}

// determinePriority determines the priority of a request based on its characteristics
func determinePriority(r *http.Request) int {
	// High priority for health checks and status endpoints
	if r.URL.Path == "/health" || r.URL.Path == "/status" {
		return 1
	}

	// Medium priority for classification requests
	if r.URL.Path == "/v1/classify" {
		return 3
	}

	// Lower priority for batch requests
	if r.URL.Path == "/v1/classify/batch" {
		return 4
	}

	// Default priority
	return 5
}

// QueueHealthMiddleware provides health check endpoint for the queue
func QueueHealthMiddleware(queue *RequestQueue) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/queue/health" {
				metrics := queue.GetMetrics()

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				// Use a simple JSON encoder for the response
				jsonResponse := fmt.Sprintf(`{
					"status": "healthy",
					"queue": {
						"total_requests": %d,
						"processed_requests": %d,
						"failed_requests": %d,
						"queue_size": %d,
						"active_workers": %d,
						"average_wait_time": "%s",
						"average_process_time": "%s",
						"last_updated": "%s"
					},
					"timestamp": "%s"
				}`,
					metrics.TotalRequests,
					metrics.ProcessedRequests,
					metrics.FailedRequests,
					metrics.QueueSize,
					metrics.ActiveWorkers,
					metrics.AverageWaitTime.String(),
					metrics.AverageProcessTime.String(),
					metrics.LastUpdated.Format(time.RFC3339),
					time.Now().UTC().Format(time.RFC3339))

				w.Write([]byte(jsonResponse))
				return
			}

			next(w, r)
		}
	}
}
