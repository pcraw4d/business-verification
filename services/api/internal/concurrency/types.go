package concurrency

import (
	"context"
	"time"
)

// ConcurrentRequest represents a request to be processed concurrently
type ConcurrentRequest struct {
	ID                string                 `json:"id"`
	Type              string                 `json:"type"`
	Data              interface{}            `json:"data"`
	RequiredResources []string               `json:"required_resources"`
	Priority          int                    `json:"priority"`
	Timeout           time.Duration          `json:"timeout"`
	Metadata          map[string]interface{} `json:"metadata"`
	CreatedAt         time.Time              `json:"created_at"`
}

// ConcurrentResponse represents the response from a concurrent request
type ConcurrentResponse struct {
	RequestID      string                 `json:"request_id"`
	Status         string                 `json:"status"`
	Data           interface{}            `json:"data"`
	Error          error                  `json:"error,omitempty"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata"`
	CompletedAt    time.Time              `json:"completed_at"`
}

// Resource represents a system resource that can be acquired/released
type Resource struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Capacity   int       `json:"capacity"`
	Used       int       `json:"used"`
	Available  int       `json:"available"`
	AcquiredAt time.Time `json:"acquired_at,omitempty"`
	ReleasedAt time.Time `json:"released_at,omitempty"`
}

// Note: ResourceManagerConfig is defined in resource_manager.go to avoid conflicts

// ConcurrentRequestHandlerConfig holds configuration for the request handler
type ConcurrentRequestHandlerConfig struct {
	MaxWorkers    int           `json:"max_workers"`
	WorkerTimeout time.Duration `json:"worker_timeout"`
	QueueSize     int           `json:"queue_size"`
}

// SynchronizationManagerConfig holds configuration for the synchronization manager
type SynchronizationManagerConfig struct {
	EnableDeadlockDetection bool `json:"enable_deadlock_detection"`
}

// DeadlockDetectorConfig holds configuration for the deadlock detector
type DeadlockDetectorConfig struct {
	DetectionInterval time.Duration `json:"detection_interval"`
}

// LockRequest represents a request to acquire a lock
type LockRequest struct {
	ResourceID string        `json:"resource_id"`
	Timeout    time.Duration `json:"timeout"`
	Priority   int           `json:"priority"`
}

// Lock represents an acquired lock on a resource
type Lock struct {
	ResourceID  string    `json:"resource_id"`
	AcquiredAt  time.Time `json:"acquired_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Owner       string    `json:"owner"`
	IsExclusive bool      `json:"is_exclusive"`
}

// DeadlockInfo represents information about a detected deadlock
type DeadlockInfo struct {
	ID         string    `json:"id"`
	Resources  []string  `json:"resources"`
	Processes  []string  `json:"processes"`
	DetectedAt time.Time `json:"detected_at"`
	ResolvedAt time.Time `json:"resolved_at,omitempty"`
	Resolution string    `json:"resolution,omitempty"`
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerClosed   CircuitBreakerState = "closed"
	CircuitBreakerOpen     CircuitBreakerState = "open"
	CircuitBreakerHalfOpen CircuitBreakerState = "half_open"
)

// CircuitBreaker represents a circuit breaker for fault tolerance
type CircuitBreaker struct {
	ID           string              `json:"id"`
	State        CircuitBreakerState `json:"state"`
	FailureCount int                 `json:"failure_count"`
	Threshold    int                 `json:"threshold"`
	Timeout      time.Duration       `json:"timeout"`
	LastFailure  time.Time           `json:"last_failure"`
	LastSuccess  time.Time           `json:"last_success"`
}

// Worker represents a worker goroutine in the pool
type Worker struct {
	ID           string    `json:"id"`
	Status       string    `json:"status"`
	CurrentTask  string    `json:"current_task"`
	StartedAt    time.Time `json:"started_at"`
	LastActivity time.Time `json:"last_activity"`
	TaskCount    int64     `json:"task_count"`
}

// ProcessingStats holds processing statistics
type ProcessingStats struct {
	TotalProcessed int64         `json:"total_processed"`
	Successful     int64         `json:"successful"`
	Failed         int64         `json:"failed"`
	AverageTime    time.Duration `json:"average_time"`
	MaxTime        time.Duration `json:"max_time"`
	MinTime        time.Duration `json:"min_time"`
	LastProcessed  time.Time     `json:"last_processed"`
}

// ResourceStats holds resource usage statistics
type ResourceStats struct {
	TotalResources     int           `json:"total_resources"`
	AvailableResources int           `json:"available_resources"`
	Utilization        float64       `json:"utilization"`
	WaitTime           time.Duration `json:"wait_time"`
	LastUpdated        time.Time     `json:"last_updated"`
}

// RequestProcessor defines the interface for processing requests
type RequestProcessor interface {
	Process(ctx context.Context, request *ConcurrentRequest) (*ConcurrentResponse, error)
}

// ResourceAcquirer defines the interface for acquiring resources
type ResourceAcquirer interface {
	Acquire(ctx context.Context, resources []string) ([]*Resource, error)
	Release(resources []*Resource) error
}

// LockManager defines the interface for managing locks
type LockManager interface {
	AcquireLock(ctx context.Context, request *LockRequest) (*Lock, error)
	ReleaseLock(lock *Lock) error
	IsLocked(resourceID string) bool
}

// CircuitBreakerManager defines the interface for managing circuit breakers
type CircuitBreakerManager interface {
	GetCircuitBreaker(id string) *CircuitBreaker
	RecordSuccess(id string)
	RecordFailure(id string)
	IsOpen(id string) bool
}
