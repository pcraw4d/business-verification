package concurrency

import (
	"context"
	"fmt"
	"sync"
	"time"

	"kyb-platform/internal/observability"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ConcurrentProcessor provides thread-safe concurrent processing capabilities
type ConcurrentProcessor struct {
	// Configuration
	config ConcurrentProcessorConfig

	// Resource management
	resourceManager *ResourceManager

	// Request handling
	requestHandler *ConcurrentRequestHandler

	// Thread-safe data structures
	dataStructures *ThreadSafeDataStructures

	// Synchronization mechanisms
	syncManager *SynchronizationManager

	// Deadlock prevention
	deadlockDetector *DeadlockDetector

	// Statistics and monitoring
	stats     *ConcurrentProcessorStats
	statsLock sync.RWMutex

	// Control
	stopChannel chan struct{}

	// Logging
	logger *zap.Logger
}

// ConcurrentProcessorConfig holds configuration for the concurrent processor
type ConcurrentProcessorConfig struct {
	// Worker pool settings
	MaxWorkers    int           `json:"max_workers"`    // Maximum number of worker goroutines
	WorkerTimeout time.Duration `json:"worker_timeout"` // Timeout for worker operations
	QueueSize     int           `json:"queue_size"`     // Size of request queue

	// Resource management
	MaxConcurrentOps        int           `json:"max_concurrent_ops"` // Maximum concurrent operations
	ResourceTimeout         time.Duration `json:"resource_timeout"`   // Timeout for resource acquisition
	EnableDeadlockDetection bool          `json:"enable_deadlock_detection"`

	// Performance settings
	EnableMetrics   bool          `json:"enable_metrics"`
	MetricsInterval time.Duration `json:"metrics_interval"`

	// Safety settings
	EnableCircuitBreaker    bool          `json:"enable_circuit_breaker"`
	CircuitBreakerThreshold int           `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration `json:"circuit_breaker_timeout"`
}

// ConcurrentProcessorStats holds concurrent processor statistics
type ConcurrentProcessorStats struct {
	// Request statistics
	TotalRequests      int64 `json:"total_requests"`
	SuccessfulRequests int64 `json:"successful_requests"`
	FailedRequests     int64 `json:"failed_requests"`
	TimeoutRequests    int64 `json:"timeout_requests"`

	// Performance metrics
	AverageResponseTime time.Duration `json:"average_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`

	// Resource usage
	ActiveWorkers       int64   `json:"active_workers"`
	QueueLength         int64   `json:"queue_length"`
	ResourceUtilization float64 `json:"resource_utilization"`

	// Error metrics
	DeadlockDetections  int64 `json:"deadlock_detections"`
	ResourceConflicts   int64 `json:"resource_conflicts"`
	CircuitBreakerTrips int64 `json:"circuit_breaker_trips"`

	// Timestamps
	LastUpdated time.Time `json:"last_updated"`
	StartTime   time.Time `json:"start_time"`
}

// NewConcurrentProcessor creates a new concurrent processor
func NewConcurrentProcessor(config *ConcurrentProcessorConfig, logger *zap.Logger) *ConcurrentProcessor {
	if config == nil {
		config = &ConcurrentProcessorConfig{
			MaxWorkers:              100,
			WorkerTimeout:           30 * time.Second,
			QueueSize:               1000,
			MaxConcurrentOps:        50,
			ResourceTimeout:         10 * time.Second,
			EnableDeadlockDetection: true,
			EnableMetrics:           true,
			MetricsInterval:         1 * time.Minute,
			EnableCircuitBreaker:    false,
			CircuitBreakerThreshold: 5,
			CircuitBreakerTimeout:   1 * time.Minute,
		}
	}

	cp := &ConcurrentProcessor{
		config:      *config,
		stats:       &ConcurrentProcessorStats{StartTime: time.Now()},
		logger:      logger,
		stopChannel: make(chan struct{}),
	}

	// Initialize components
	if err := cp.initializeComponents(); err != nil {
		cp.logger.Error("failed to initialize concurrent processor components", zap.Error(err))
	}

	return cp
}

// Start starts the concurrent processor
func (cp *ConcurrentProcessor) Start() error {
	cp.logger.Info("Starting concurrent processor",
		zap.Int("max_workers", cp.config.MaxWorkers),
		zap.Int("queue_size", cp.config.QueueSize),
		zap.Int("max_concurrent_ops", cp.config.MaxConcurrentOps))

	// Start resource manager
	if err := cp.resourceManager.Start(); err != nil {
		return fmt.Errorf("failed to start resource manager: %w", err)
	}

	// Start request handler
	if err := cp.requestHandler.Start(); err != nil {
		return fmt.Errorf("failed to start request handler: %w", err)
	}

	// Start synchronization manager
	if err := cp.syncManager.Start(); err != nil {
		return fmt.Errorf("failed to start synchronization manager: %w", err)
	}

	// Start deadlock detector if enabled
	if cp.config.EnableDeadlockDetection {
		if err := cp.deadlockDetector.Start(); err != nil {
			return fmt.Errorf("failed to start deadlock detector: %w", err)
		}
	}

	// Start metrics collection if enabled
	if cp.config.EnableMetrics {
		go cp.startMetricsCollection()
	}

	cp.logger.Info("Concurrent processor started successfully")
	return nil
}

// Stop stops the concurrent processor
func (cp *ConcurrentProcessor) Stop() {
	cp.logger.Info("Stopping concurrent processor")
	close(cp.stopChannel)

	// Stop components
	if cp.resourceManager != nil {
		cp.resourceManager.Stop()
	}
	if cp.requestHandler != nil {
		cp.requestHandler.Stop()
	}
	if cp.syncManager != nil {
		cp.syncManager.Stop()
	}
	if cp.deadlockDetector != nil {
		cp.deadlockDetector.Stop()
	}

	cp.logger.Info("Concurrent processor stopped")
}

// ProcessRequest processes a request concurrently with resource safety
func (cp *ConcurrentProcessor) ProcessRequest(ctx context.Context, request *ConcurrentRequest) (*ConcurrentResponse, error) {
	start := time.Now()

	// Track request
	cp.incrementRequest("total")

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, cp.config.WorkerTimeout)
	defer cancel()

	// Acquire resources
	resources, err := cp.resourceManager.Acquire(ctx, request.RequiredResources)
	if err != nil {
		cp.incrementRequest("failed")
		cp.incrementResourceConflict()
		return nil, fmt.Errorf("failed to acquire resources: %w", err)
	}
	defer cp.resourceManager.Release(resources)

	// Submit request for processing
	response, err := cp.requestHandler.ProcessRequest(ctx, request)
	if err != nil {
		cp.incrementRequest("failed")
		return nil, fmt.Errorf("failed to submit request: %w", err)
	}

	// Update statistics
	duration := time.Since(start)
	cp.recordResponseTime(duration)
	cp.incrementRequest("successful")

	return response, nil
}

// GetStats returns current concurrent processor statistics
func (cp *ConcurrentProcessor) GetStats() *ConcurrentProcessorStats {
	cp.statsLock.RLock()
	defer cp.statsLock.RUnlock()

	stats := *cp.stats
	return &stats
}

// ResetStats resets all statistics
func (cp *ConcurrentProcessor) ResetStats() {
	cp.statsLock.Lock()
	defer cp.statsLock.Unlock()

	cp.stats = &ConcurrentProcessorStats{
		StartTime:   time.Now(),
		LastUpdated: time.Now(),
	}
}

// GetResourceManager returns the resource manager
func (cp *ConcurrentProcessor) GetResourceManager() *ResourceManager {
	return cp.resourceManager
}

// GetRequestHandler returns the request handler
func (cp *ConcurrentProcessor) GetRequestHandler() *ConcurrentRequestHandler {
	return cp.requestHandler
}

// GetDataStructures returns the thread-safe data structures
func (cp *ConcurrentProcessor) GetDataStructures() *ThreadSafeDataStructures {
	return cp.dataStructures
}

// GetSynchronizationManager returns the synchronization manager
func (cp *ConcurrentProcessor) GetSynchronizationManager() *SynchronizationManager {
	return cp.syncManager
}

// Helper methods

func (cp *ConcurrentProcessor) initializeComponents() error {
	// Initialize resource manager
	resourceConfig := &ResourceManagerConfig{
		MaxConcurrentOps: cp.config.MaxConcurrentOps,
		ResourceTimeout:  cp.config.ResourceTimeout,
	}
	// Create observability logger and tracer
	obsLogger := observability.NewLogger(cp.logger)
	tracer := trace.NewNoopTracerProvider().Tracer("concurrent-processor")
	cp.resourceManager = NewResourceManager(resourceConfig, obsLogger, tracer)

	// Initialize thread-safe data structures
	cp.dataStructures = NewThreadSafeDataStructures()

	// Initialize synchronization manager
	syncConfig := &SynchronizationManagerConfig{
		EnableDeadlockDetection: cp.config.EnableDeadlockDetection,
	}
	cp.syncManager = NewSynchronizationManager(syncConfig, cp.logger)

	// Initialize deadlock detector if enabled
	if cp.config.EnableDeadlockDetection {
		deadlockConfig := &DeadlockDetectorConfig{
			DetectionInterval: 5 * time.Second,
		}
		cp.deadlockDetector = NewDeadlockDetector(deadlockConfig, cp.logger)
	}

	// Initialize request handler with a default processor
	requestConfig := &ConcurrentRequestHandlerConfig{
		MaxWorkers:    cp.config.MaxWorkers,
		WorkerTimeout: cp.config.WorkerTimeout,
		QueueSize:     cp.config.QueueSize,
	}

	// Create a default processor that just returns success
	defaultProcessor := &defaultRequestProcessor{}
	cp.requestHandler = NewConcurrentRequestHandler(requestConfig, cp.logger, cp.resourceManager, defaultProcessor)

	return nil
}

// defaultRequestProcessor is a simple processor for testing
type defaultRequestProcessor struct{}

func (p *defaultRequestProcessor) Process(ctx context.Context, request *ConcurrentRequest) (*ConcurrentResponse, error) {
	return &ConcurrentResponse{
		RequestID:   request.ID,
		Status:      "success",
		Data:        request.Data,
		CompletedAt: time.Now(),
	}, nil
}

func (cp *ConcurrentProcessor) startMetricsCollection() {
	ticker := time.NewTicker(cp.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cp.stopChannel:
			return
		case <-ticker.C:
			cp.collectMetrics()
		}
	}
}

func (cp *ConcurrentProcessor) collectMetrics() {
	cp.statsLock.Lock()
	defer cp.statsLock.Unlock()

	// Update active workers
	if cp.requestHandler != nil {
		cp.stats.ActiveWorkers = int64(cp.requestHandler.GetActiveWorkers())
		cp.stats.QueueLength = int64(cp.requestHandler.GetQueueSize())
	}

	// Update resource utilization
	if cp.resourceManager != nil {
		stats := cp.resourceManager.GetStats()
		if stats != nil {
			cp.stats.ResourceUtilization = stats.CPUUtilization
		}
	}

	// Update deadlock detections
	if cp.deadlockDetector != nil {
		cp.stats.DeadlockDetections = cp.deadlockDetector.GetDetectionCount()
	}

	cp.stats.LastUpdated = time.Now()

	cp.logger.Debug("Collected concurrent processor metrics",
		zap.Int64("active_workers", cp.stats.ActiveWorkers),
		zap.Int64("queue_length", cp.stats.QueueLength),
		zap.Float64("resource_utilization", cp.stats.ResourceUtilization),
		zap.Int64("deadlock_detections", cp.stats.DeadlockDetections))
}

func (cp *ConcurrentProcessor) incrementRequest(requestType string) {
	cp.statsLock.Lock()
	defer cp.statsLock.Unlock()

	switch requestType {
	case "total":
		cp.stats.TotalRequests++
	case "successful":
		cp.stats.SuccessfulRequests++
	case "failed":
		cp.stats.FailedRequests++
	case "timeout":
		cp.stats.TimeoutRequests++
	}

	cp.stats.LastUpdated = time.Now()
}

func (cp *ConcurrentProcessor) recordResponseTime(duration time.Duration) {
	cp.statsLock.Lock()
	defer cp.statsLock.Unlock()

	// Update response time statistics
	if cp.stats.MaxResponseTime < duration {
		cp.stats.MaxResponseTime = duration
	}
	if cp.stats.MinResponseTime == 0 || cp.stats.MinResponseTime > duration {
		cp.stats.MinResponseTime = duration
	}

	// Update average response time
	totalRequests := cp.stats.SuccessfulRequests
	if totalRequests > 0 {
		totalTime := cp.stats.AverageResponseTime * time.Duration(totalRequests-1)
		cp.stats.AverageResponseTime = (totalTime + duration) / time.Duration(totalRequests)
	} else {
		cp.stats.AverageResponseTime = duration
	}

	cp.stats.LastUpdated = time.Now()
}

func (cp *ConcurrentProcessor) incrementResourceConflict() {
	cp.statsLock.Lock()
	defer cp.statsLock.Unlock()
	cp.stats.ResourceConflicts++
	cp.stats.LastUpdated = time.Now()
}
