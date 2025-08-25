package concurrency

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// SmartParallelProcessor provides intelligent parallel processing with deduplication
type SmartParallelProcessor struct {
	// Configuration
	config *SmartProcessorConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Task management
	taskRegistry    map[string]*TaskInfo
	taskRegistryMux sync.RWMutex

	// Result sharing
	resultCache    map[string]*TaskResult
	resultCacheMux sync.RWMutex

	// Dependency tracking
	dependencyGraph map[string][]string
	dependencyMux   sync.RWMutex

	// Worker pool
	workerPool    *WorkerPool
	workerPoolMux sync.RWMutex

	// Performance monitoring
	performanceMetrics *PerformanceMetrics
	metricsMux         sync.RWMutex

	// Task scheduling
	scheduler    *TaskScheduler
	schedulerMux sync.RWMutex

	// Context for shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// SmartProcessorConfig holds configuration for the smart parallel processor
type SmartProcessorConfig struct {
	// Worker pool settings
	MaxWorkers        int
	MinWorkers        int
	WorkerIdleTimeout time.Duration
	WorkerMaxTasks    int

	// Task deduplication settings
	EnableDeduplication bool
	DeduplicationWindow time.Duration
	MaxDuplicateTasks   int

	// Result sharing settings
	EnableResultSharing bool
	ResultCacheSize     int
	ResultCacheTTL      time.Duration

	// Dependency tracking settings
	EnableDependencyTracking bool
	MaxDependencyDepth       int
	DependencyTimeout        time.Duration

	// Performance settings
	EnablePerformanceMonitoring bool
	PerformanceUpdateInterval   time.Duration
	MaxConcurrentTasks          int

	// Scheduling settings
	EnableSmartScheduling bool
	SchedulingAlgorithm   string
	LoadBalancingStrategy string

	// Processing settings
	TaskTimeout            time.Duration
	RetryAttempts          int
	RetryBackoffMultiplier float64
}

// TaskInfo represents information about a task
type TaskInfo struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"`
	Priority       int                    `json:"priority"`
	Status         TaskStatus             `json:"status"`
	CreatedAt      time.Time              `json:"created_at"`
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	Input          map[string]interface{} `json:"input"`
	Output         interface{}            `json:"output,omitempty"`
	Error          error                  `json:"error,omitempty"`
	Dependencies   []string               `json:"dependencies"`
	Dependents     []string               `json:"dependents"`
	WorkerID       string                 `json:"worker_id,omitempty"`
	RetryCount     int                    `json:"retry_count"`
	ProcessingTime time.Duration          `json:"processing_time"`
	ResourceUsage  *ResourceUsage         `json:"resource_usage,omitempty"`
	DuplicateOf    string                 `json:"duplicate_of,omitempty"`
	CacheKey       string                 `json:"cache_key,omitempty"`
}

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending      TaskStatus = "pending"
	TaskStatusRunning      TaskStatus = "running"
	TaskStatusCompleted    TaskStatus = "completed"
	TaskStatusFailed       TaskStatus = "failed"
	TaskStatusCancelled    TaskStatus = "cancelled"
	TaskStatusDeduplicated TaskStatus = "deduplicated"
)

// TaskResult represents the result of a task
type TaskResult struct {
	TaskID         string                 `json:"task_id"`
	Status         TaskStatus             `json:"status"`
	Result         interface{}            `json:"result,omitempty"`
	Error          error                  `json:"error,omitempty"`
	ProcessingTime time.Duration          `json:"processing_time"`
	ResourceUsage  *ResourceUsage         `json:"resource_usage,omitempty"`
	CacheKey       string                 `json:"cache_key"`
	CreatedAt      time.Time              `json:"created_at"`
	ExpiresAt      time.Time              `json:"expires_at"`
	AccessCount    int                    `json:"access_count"`
	LastAccessed   time.Time              `json:"last_accessed"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ResourceUsage represents resource usage information
type ResourceUsage struct {
	CPUTime     time.Duration `json:"cpu_time"`
	MemoryUsage int64         `json:"memory_usage_bytes"`
	NetworkIO   int64         `json:"network_io_bytes"`
	DiskIO      int64         `json:"disk_io_bytes"`
	Concurrency int           `json:"concurrency"`
	WaitTime    time.Duration `json:"wait_time"`
}

// WorkerPool manages a pool of workers
type WorkerPool struct {
	Workers          map[string]*SmartWorker
	AvailableWorkers chan *SmartWorker
	MaxWorkers       int
	MinWorkers       int
	IdleTimeout      time.Duration
	MaxTasks         int
	Mux              sync.RWMutex
}

// SmartWorker represents a worker in the smart parallel processor pool
type SmartWorker struct {
	ID                  string
	Status              WorkerStatus
	CurrentTask         *TaskInfo
	TaskCount           int
	TotalProcessingTime time.Duration
	LastActivity        time.Time
	ResourceUsage       *ResourceUsage
	Mux                 sync.RWMutex
}

// WorkerStatus represents the status of a worker
type WorkerStatus string

const (
	WorkerStatusIdle     WorkerStatus = "idle"
	WorkerStatusBusy     WorkerStatus = "busy"
	WorkerStatusStopping WorkerStatus = "stopping"
	WorkerStatusStopped  WorkerStatus = "stopped"
)

// TaskScheduler manages task scheduling
type TaskScheduler struct {
	Algorithm       string
	LoadBalancer    string
	TaskQueue       []*TaskInfo
	PriorityQueue   []*TaskInfo
	DependencyQueue []*TaskInfo
	Mux             sync.RWMutex
}

// PerformanceMetrics tracks performance metrics
type PerformanceMetrics struct {
	TotalTasksProcessed   int64
	SuccessfulTasks       int64
	FailedTasks           int64
	DeduplicatedTasks     int64
	CacheHits             int64
	CacheMisses           int64
	AverageProcessingTime time.Duration
	TotalProcessingTime   time.Duration
	ActiveWorkers         int
	IdleWorkers           int
	QueueLength           int
	ResourceUtilization   float64
	Throughput            float64
	ErrorRate             float64
	LastUpdated           time.Time
	Mux                   sync.RWMutex
}

// NewSmartParallelProcessor creates a new smart parallel processor
func NewSmartParallelProcessor(
	config *SmartProcessorConfig,
	logger *observability.Logger,
	tracer trace.Tracer,
) *SmartParallelProcessor {
	// Set default configuration
	if config == nil {
		config = &SmartProcessorConfig{
			MaxWorkers:                  10,
			MinWorkers:                  2,
			WorkerIdleTimeout:           5 * time.Minute,
			WorkerMaxTasks:              100,
			EnableDeduplication:         true,
			DeduplicationWindow:         1 * time.Minute,
			MaxDuplicateTasks:           5,
			EnableResultSharing:         true,
			ResultCacheSize:             1000,
			ResultCacheTTL:              30 * time.Minute,
			EnableDependencyTracking:    true,
			MaxDependencyDepth:          5,
			DependencyTimeout:           10 * time.Minute,
			EnablePerformanceMonitoring: true,
			PerformanceUpdateInterval:   30 * time.Second,
			MaxConcurrentTasks:          50,
			EnableSmartScheduling:       true,
			SchedulingAlgorithm:         "priority",
			LoadBalancingStrategy:       "round_robin",
			TaskTimeout:                 5 * time.Minute,
			RetryAttempts:               3,
			RetryBackoffMultiplier:      2.0,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	processor := &SmartParallelProcessor{
		config:          config,
		logger:          logger,
		tracer:          tracer,
		taskRegistry:    make(map[string]*TaskInfo),
		resultCache:     make(map[string]*TaskResult),
		dependencyGraph: make(map[string][]string),
		performanceMetrics: &PerformanceMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize worker pool
	processor.initializeWorkerPool()

	// Initialize task scheduler
	processor.initializeTaskScheduler()

	// Start background workers
	processor.startBackgroundWorkers()

	return processor
}

// initializeWorkerPool initializes the worker pool
func (spp *SmartParallelProcessor) initializeWorkerPool() {
	spp.workerPool = &WorkerPool{
		Workers:          make(map[string]*SmartWorker),
		AvailableWorkers: make(chan *SmartWorker, spp.config.MaxWorkers),
		MaxWorkers:       spp.config.MaxWorkers,
		MinWorkers:       spp.config.MinWorkers,
		IdleTimeout:      spp.config.WorkerIdleTimeout,
		MaxTasks:         spp.config.WorkerMaxTasks,
	}

	// Create initial workers
	for i := 0; i < spp.config.MinWorkers; i++ {
		worker := spp.createWorker(fmt.Sprintf("worker-%d", i))
		spp.workerPool.Workers[worker.ID] = worker
		spp.workerPool.AvailableWorkers <- worker
	}
}

// initializeTaskScheduler initializes the task scheduler
func (spp *SmartParallelProcessor) initializeTaskScheduler() {
	spp.scheduler = &TaskScheduler{
		Algorithm:       spp.config.SchedulingAlgorithm,
		LoadBalancer:    spp.config.LoadBalancingStrategy,
		TaskQueue:       []*TaskInfo{},
		PriorityQueue:   []*TaskInfo{},
		DependencyQueue: []*TaskInfo{},
	}
}

// createWorker creates a new worker
func (spp *SmartParallelProcessor) createWorker(id string) *SmartWorker {
	return &SmartWorker{
		ID:                  id,
		Status:              WorkerStatusIdle,
		CurrentTask:         nil,
		TaskCount:           0,
		TotalProcessingTime: 0,
		LastActivity:        time.Now(),
		ResourceUsage:       &ResourceUsage{},
	}
}

// startBackgroundWorkers starts background workers for maintenance
func (spp *SmartParallelProcessor) startBackgroundWorkers() {
	// Start performance monitoring worker
	if spp.config.EnablePerformanceMonitoring {
		go spp.performanceMonitoringWorker()
	}

	// Start cache cleanup worker
	if spp.config.EnableResultSharing {
		go spp.cacheCleanupWorker()
	}

	// Start worker pool maintenance worker
	go spp.workerPoolMaintenanceWorker()

	// Start task deduplication worker
	if spp.config.EnableDeduplication {
		go spp.taskDeduplicationWorker()
	}
}

// SubmitTask submits a task for processing
func (spp *SmartParallelProcessor) SubmitTask(
	ctx context.Context,
	taskType string,
	input map[string]interface{},
	priority int,
	handler TaskHandler,
) (string, error) {
	ctx, span := spp.tracer.Start(ctx, "SmartParallelProcessor.SubmitTask")
	defer span.End()

	span.SetAttributes(
		attribute.String("task_type", taskType),
		attribute.Int("priority", priority),
	)

	// Generate task ID
	taskID := spp.generateTaskID(taskType, input)

	// Check for deduplication
	if spp.config.EnableDeduplication {
		if duplicateTaskID := spp.checkForDuplicate(taskID, taskType, input); duplicateTaskID != "" {
			spp.logger.Info("task deduplicated", map[string]interface{}{
				"task_id":      taskID,
				"duplicate_of": duplicateTaskID,
			})
			return duplicateTaskID, nil
		}
	}

	// Check result cache
	if spp.config.EnableResultSharing {
		if cachedResult := spp.getCachedResult(taskID); cachedResult != nil {
			spp.logger.Info("using cached result", map[string]interface{}{
				"task_id": taskID,
			})
			return taskID, nil
		}
	}

	// Create task info
	taskInfo := &TaskInfo{
		ID:           taskID,
		Type:         taskType,
		Priority:     priority,
		Status:       TaskStatusPending,
		CreatedAt:    time.Now(),
		Input:        input,
		Dependencies: []string{},
		Dependents:   []string{},
		RetryCount:   0,
		CacheKey:     taskID,
	}

	// Register task
	spp.registerTask(taskInfo)

	// Add to scheduler
	spp.addTaskToScheduler(taskInfo)

	// Process task
	go spp.processTask(ctx, taskInfo, handler)

	spp.logger.Info("task submitted", map[string]interface{}{
		"task_id":   taskID,
		"task_type": taskType,
		"priority":  priority,
	})

	return taskID, nil
}

// processTask processes a task
func (spp *SmartParallelProcessor) processTask(
	ctx context.Context,
	taskInfo *TaskInfo,
	handler TaskHandler,
) {
	ctx, span := spp.tracer.Start(ctx, "SmartParallelProcessor.processTask")
	defer span.End()

	span.SetAttributes(
		attribute.String("task_id", taskInfo.ID),
		attribute.String("task_type", taskInfo.Type),
	)

	// Wait for dependencies
	if err := spp.waitForDependencies(ctx, taskInfo); err != nil {
		spp.handleTaskFailure(taskInfo, err)
		return
	}

	// Get worker
	worker := spp.getAvailableWorker()
	if worker == nil {
		spp.handleTaskFailure(taskInfo, fmt.Errorf("no available workers"))
		return
	}

	// Assign task to worker
	spp.assignTaskToWorker(taskInfo, worker)

	// Execute task
	result, err := spp.executeTask(ctx, taskInfo, handler)

	// Handle result
	if err != nil {
		spp.handleTaskFailure(taskInfo, err)
	} else {
		spp.handleTaskSuccess(taskInfo, result)
	}

	// Release worker
	spp.releaseWorker(worker)
}

// executeTask executes a task
func (spp *SmartParallelProcessor) executeTask(
	ctx context.Context,
	taskInfo *TaskInfo,
	handler TaskHandler,
) (interface{}, error) {
	ctx, span := spp.tracer.Start(ctx, "SmartParallelProcessor.executeTask")
	defer span.End()

	startTime := time.Now()

	// Update task status
	spp.updateTaskStatus(taskInfo.ID, TaskStatusRunning)
	now := time.Now()
	taskInfo.StartedAt = &now

	// Execute handler
	result, err := handler(ctx, taskInfo.Input)

	// Calculate processing time
	processingTime := time.Since(startTime)
	taskInfo.ProcessingTime = processingTime

	// Update resource usage
	taskInfo.ResourceUsage = &ResourceUsage{
		CPUTime:     processingTime,
		Concurrency: 1,
	}

	span.SetAttributes(
		attribute.String("processing_time", processingTime.String()),
		attribute.Bool("success", err == nil),
	)

	return result, err
}

// handleTaskSuccess handles successful task completion
func (spp *SmartParallelProcessor) handleTaskSuccess(taskInfo *TaskInfo, result interface{}) {
	// Update task info
	taskInfo.Status = TaskStatusCompleted
	taskInfo.Output = result
	now := time.Now()
	taskInfo.CompletedAt = &now

	// Cache result
	if spp.config.EnableResultSharing {
		spp.cacheResult(taskInfo.ID, result, taskInfo.ProcessingTime)
	}

	// Update metrics
	spp.updatePerformanceMetrics(true, taskInfo.ProcessingTime)

	// Notify dependents
	spp.notifyDependents(taskInfo.ID, result)

	spp.logger.Info("task completed successfully", map[string]interface{}{
		"task_id":         taskInfo.ID,
		"processing_time": taskInfo.ProcessingTime,
	})
}

// handleTaskFailure handles task failure
func (spp *SmartParallelProcessor) handleTaskFailure(taskInfo *TaskInfo, err error) {
	// Update task info
	taskInfo.Status = TaskStatusFailed
	taskInfo.Error = err
	now := time.Now()
	taskInfo.CompletedAt = &now

	// Retry logic
	if taskInfo.RetryCount < spp.config.RetryAttempts {
		taskInfo.RetryCount++
		taskInfo.Status = TaskStatusPending
		taskInfo.Error = nil
		taskInfo.CompletedAt = nil

		// Exponential backoff
		backoffTime := time.Duration(float64(spp.config.TaskTimeout) *
			(spp.config.RetryBackoffMultiplier * float64(taskInfo.RetryCount)))

		spp.logger.Info("retrying task", map[string]interface{}{
			"task_id":      taskInfo.ID,
			"retry_count":  taskInfo.RetryCount,
			"backoff_time": backoffTime,
		})

		time.AfterFunc(backoffTime, func() {
			spp.addTaskToScheduler(taskInfo)
		})

		return
	}

	// Update metrics
	spp.updatePerformanceMetrics(false, taskInfo.ProcessingTime)

	// Notify dependents of failure
	spp.notifyDependents(taskInfo.ID, err)

	spp.logger.Error("task failed permanently", map[string]interface{}{
		"task_id":     taskInfo.ID,
		"error":       err.Error(),
		"retry_count": taskInfo.RetryCount,
	})
}

// checkForDuplicate checks for duplicate tasks
func (spp *SmartParallelProcessor) checkForDuplicate(taskID, taskType string, input map[string]interface{}) string {
	spp.taskRegistryMux.RLock()
	defer spp.taskRegistryMux.RUnlock()

	// Check recent tasks of the same type
	windowStart := time.Now().Add(-spp.config.DeduplicationWindow)

	for id, task := range spp.taskRegistry {
		if task.Type == taskType &&
			task.CreatedAt.After(windowStart) &&
			spp.isInputEquivalent(task.Input, input) {
			return id
		}
	}

	return ""
}

// isInputEquivalent checks if two inputs are equivalent
func (spp *SmartParallelProcessor) isInputEquivalent(input1, input2 map[string]interface{}) bool {
	// Simple equivalence check - in a real implementation, this would be more sophisticated
	if len(input1) != len(input2) {
		return false
	}

	for key, value1 := range input1 {
		if value2, exists := input2[key]; !exists || value1 != value2 {
			return false
		}
	}

	return true
}

// getCachedResult gets a cached result
func (spp *SmartParallelProcessor) getCachedResult(taskID string) *TaskResult {
	spp.resultCacheMux.RLock()
	defer spp.resultCacheMux.RUnlock()

	if result, exists := spp.resultCache[taskID]; exists && time.Now().Before(result.ExpiresAt) {
		result.AccessCount++
		result.LastAccessed = time.Now()
		return result
	}

	return nil
}

// cacheResult caches a task result
func (spp *SmartParallelProcessor) cacheResult(taskID string, result interface{}, processingTime time.Duration) {
	spp.resultCacheMux.Lock()
	defer spp.resultCacheMux.Unlock()

	// Check cache size limit
	if len(spp.resultCache) >= spp.config.ResultCacheSize {
		spp.evictOldestCacheEntry()
	}

	cachedResult := &TaskResult{
		TaskID:         taskID,
		Status:         TaskStatusCompleted,
		Result:         result,
		ProcessingTime: processingTime,
		CacheKey:       taskID,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(spp.config.ResultCacheTTL),
		AccessCount:    1,
		LastAccessed:   time.Now(),
		Metadata:       make(map[string]interface{}),
	}

	spp.resultCache[taskID] = cachedResult
}

// evictOldestCacheEntry evicts the oldest cache entry
func (spp *SmartParallelProcessor) evictOldestCacheEntry() {
	var oldestKey string
	var oldestTime time.Time

	for key, result := range spp.resultCache {
		if oldestKey == "" || result.LastAccessed.Before(oldestTime) {
			oldestKey = key
			oldestTime = result.LastAccessed
		}
	}

	if oldestKey != "" {
		delete(spp.resultCache, oldestKey)
	}
}

// getAvailableWorker gets an available worker
func (spp *SmartParallelProcessor) getAvailableWorker() *SmartWorker {
	select {
	case worker := <-spp.workerPool.AvailableWorkers:
		return worker
	case <-time.After(5 * time.Second):
		return nil
	}
}

// assignTaskToWorker assigns a task to a worker
func (spp *SmartParallelProcessor) assignTaskToWorker(taskInfo *TaskInfo, worker *SmartWorker) {
	worker.Mux.Lock()
	defer worker.Mux.Unlock()

	worker.Status = WorkerStatusBusy
	worker.CurrentTask = taskInfo
	worker.TaskCount++
	worker.LastActivity = time.Now()

	taskInfo.WorkerID = worker.ID
}

// releaseWorker releases a worker back to the pool
func (spp *SmartParallelProcessor) releaseWorker(worker *SmartWorker) {
	worker.Mux.Lock()
	defer worker.Mux.Unlock()

	worker.Status = WorkerStatusIdle
	worker.CurrentTask = nil
	worker.LastActivity = time.Now()

	// Return worker to pool
	select {
	case spp.workerPool.AvailableWorkers <- worker:
	default:
		// Pool is full, worker will be cleaned up by maintenance worker
	}
}

// waitForDependencies waits for task dependencies to complete
func (spp *SmartParallelProcessor) waitForDependencies(ctx context.Context, taskInfo *TaskInfo) error {
	if len(taskInfo.Dependencies) == 0 {
		return nil
	}

	timeout := time.After(spp.config.DependencyTimeout)

	for _, depID := range taskInfo.Dependencies {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timeout:
				return fmt.Errorf("dependency timeout for task %s", depID)
			default:
				if spp.isTaskCompleted(depID) {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	return nil
}

// isTaskCompleted checks if a task is completed
func (spp *SmartParallelProcessor) isTaskCompleted(taskID string) bool {
	spp.taskRegistryMux.RLock()
	defer spp.taskRegistryMux.RUnlock()

	if task, exists := spp.taskRegistry[taskID]; exists {
		return task.Status == TaskStatusCompleted
	}

	return false
}

// notifyDependents notifies dependent tasks
func (spp *SmartParallelProcessor) notifyDependents(taskID string, result interface{}) {
	spp.dependencyMux.RLock()
	dependents := spp.dependencyGraph[taskID]
	spp.dependencyMux.RUnlock()

	for _, depID := range dependents {
		// In a real implementation, this would trigger dependent task processing
		spp.logger.Debug("notified dependent task", map[string]interface{}{
			"task_id":      taskID,
			"dependent_id": depID,
		})
	}
}

// generateTaskID generates a unique task ID
func (spp *SmartParallelProcessor) generateTaskID(taskType string, input map[string]interface{}) string {
	// Simple ID generation - in a real implementation, this would be more sophisticated
	return fmt.Sprintf("%s-%d", taskType, time.Now().UnixNano())
}

// registerTask registers a task
func (spp *SmartParallelProcessor) registerTask(taskInfo *TaskInfo) {
	spp.taskRegistryMux.Lock()
	defer spp.taskRegistryMux.Unlock()

	spp.taskRegistry[taskInfo.ID] = taskInfo
}

// updateTaskStatus updates task status
func (spp *SmartParallelProcessor) updateTaskStatus(taskID string, status TaskStatus) {
	spp.taskRegistryMux.Lock()
	defer spp.taskRegistryMux.Unlock()

	if task, exists := spp.taskRegistry[taskID]; exists {
		task.Status = status
	}
}

// addTaskToScheduler adds a task to the scheduler
func (spp *SmartParallelProcessor) addTaskToScheduler(taskInfo *TaskInfo) {
	spp.schedulerMux.Lock()
	defer spp.schedulerMux.Unlock()

	spp.scheduler.TaskQueue = append(spp.scheduler.TaskQueue, taskInfo)
}

// updatePerformanceMetrics updates performance metrics
func (spp *SmartParallelProcessor) updatePerformanceMetrics(success bool, processingTime time.Duration) {
	spp.metricsMux.Lock()
	defer spp.metricsMux.Unlock()

	spp.performanceMetrics.TotalTasksProcessed++
	spp.performanceMetrics.TotalProcessingTime += processingTime

	if success {
		spp.performanceMetrics.SuccessfulTasks++
	} else {
		spp.performanceMetrics.FailedTasks++
	}

	// Update average processing time
	totalTasks := spp.performanceMetrics.SuccessfulTasks + spp.performanceMetrics.FailedTasks
	if totalTasks > 0 {
		spp.performanceMetrics.AverageProcessingTime =
			spp.performanceMetrics.TotalProcessingTime / time.Duration(totalTasks)
	}

	// Update error rate
	if spp.performanceMetrics.TotalTasksProcessed > 0 {
		spp.performanceMetrics.ErrorRate =
			float64(spp.performanceMetrics.FailedTasks) / float64(spp.performanceMetrics.TotalTasksProcessed)
	}

	spp.performanceMetrics.LastUpdated = time.Now()
}

// background workers
func (spp *SmartParallelProcessor) performanceMonitoringWorker() {
	ticker := time.NewTicker(spp.config.PerformanceUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-spp.ctx.Done():
			return
		case <-ticker.C:
			spp.updatePerformanceMetricsPeriodic()
		}
	}
}

func (spp *SmartParallelProcessor) cacheCleanupWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-spp.ctx.Done():
			return
		case <-ticker.C:
			spp.cleanupExpiredCacheEntries()
		}
	}
}

func (spp *SmartParallelProcessor) workerPoolMaintenanceWorker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-spp.ctx.Done():
			return
		case <-ticker.C:
			spp.maintainWorkerPool()
		}
	}
}

func (spp *SmartParallelProcessor) taskDeduplicationWorker() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-spp.ctx.Done():
			return
		case <-ticker.C:
			spp.cleanupDuplicateTasks()
		}
	}
}

// maintenance functions
func (spp *SmartParallelProcessor) updatePerformanceMetricsPeriodic() {
	spp.metricsMux.Lock()
	defer spp.metricsMux.Unlock()

	// Update active/idle worker counts
	activeWorkers := 0
	idleWorkers := 0

	spp.workerPoolMux.RLock()
	for _, worker := range spp.workerPool.Workers {
		if worker.Status == WorkerStatusBusy {
			activeWorkers++
		} else {
			idleWorkers++
		}
	}
	spp.workerPoolMux.RUnlock()

	spp.performanceMetrics.ActiveWorkers = activeWorkers
	spp.performanceMetrics.IdleWorkers = idleWorkers

	// Update queue length
	spp.schedulerMux.RLock()
	queueLength := len(spp.scheduler.TaskQueue)
	spp.schedulerMux.RUnlock()

	spp.performanceMetrics.QueueLength = queueLength

	// Update throughput
	if spp.performanceMetrics.AverageProcessingTime > 0 {
		spp.performanceMetrics.Throughput =
			float64(time.Second) / float64(spp.performanceMetrics.AverageProcessingTime)
	}

	spp.performanceMetrics.LastUpdated = time.Now()
}

func (spp *SmartParallelProcessor) cleanupExpiredCacheEntries() {
	spp.resultCacheMux.Lock()
	defer spp.resultCacheMux.Unlock()

	now := time.Now()
	for key, result := range spp.resultCache {
		if now.After(result.ExpiresAt) {
			delete(spp.resultCache, key)
		}
	}
}

func (spp *SmartParallelProcessor) maintainWorkerPool() {
	spp.workerPoolMux.Lock()
	defer spp.workerPoolMux.Unlock()

	now := time.Now()

	// Clean up idle workers
	for id, worker := range spp.workerPool.Workers {
		if worker.Status == WorkerStatusIdle &&
			now.Sub(worker.LastActivity) > spp.config.WorkerIdleTimeout &&
			len(spp.workerPool.Workers) > spp.config.MinWorkers {
			delete(spp.workerPool.Workers, id)
		}
	}

	// Create new workers if needed
	if len(spp.workerPool.Workers) < spp.config.MinWorkers {
		for i := len(spp.workerPool.Workers); i < spp.config.MinWorkers; i++ {
			worker := spp.createWorker(fmt.Sprintf("worker-%d", i))
			spp.workerPool.Workers[worker.ID] = worker
			spp.workerPool.AvailableWorkers <- worker
		}
	}
}

func (spp *SmartParallelProcessor) cleanupDuplicateTasks() {
	spp.taskRegistryMux.Lock()
	defer spp.taskRegistryMux.Unlock()

	windowStart := time.Now().Add(-spp.config.DeduplicationWindow)

	for id, task := range spp.taskRegistry {
		if task.Status == TaskStatusDeduplicated &&
			task.CreatedAt.Before(windowStart) {
			delete(spp.taskRegistry, id)
		}
	}
}

// GetTaskResult gets the result of a task
func (spp *SmartParallelProcessor) GetTaskResult(taskID string) (*TaskResult, error) {
	// Check cache first
	if result := spp.getCachedResult(taskID); result != nil {
		return result, nil
	}

	// Check task registry
	spp.taskRegistryMux.RLock()
	task, exists := spp.taskRegistry[taskID]
	spp.taskRegistryMux.RUnlock()

	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	if task.Status != TaskStatusCompleted {
		return nil, fmt.Errorf("task not completed: %s (status: %s)", taskID, task.Status)
	}

	// Create result from task
	result := &TaskResult{
		TaskID:         task.ID,
		Status:         task.Status,
		Result:         task.Output,
		Error:          task.Error,
		ProcessingTime: task.ProcessingTime,
		ResourceUsage:  task.ResourceUsage,
		CacheKey:       task.CacheKey,
		CreatedAt:      task.CreatedAt,
		ExpiresAt:      time.Now().Add(spp.config.ResultCacheTTL),
		AccessCount:    1,
		LastAccessed:   time.Now(),
		Metadata:       make(map[string]interface{}),
	}

	return result, nil
}

// GetPerformanceMetrics gets performance metrics
func (spp *SmartParallelProcessor) GetPerformanceMetrics() *PerformanceMetrics {
	spp.metricsMux.RLock()
	defer spp.metricsMux.RUnlock()

	// Return a copy to avoid race conditions
	metrics := *spp.performanceMetrics
	return &metrics
}

// Shutdown shuts down the processor
func (spp *SmartParallelProcessor) Shutdown() {
	spp.cancel()

	// Wait for background workers to finish
	time.Sleep(1 * time.Second)

	spp.logger.Info("smart parallel processor shutdown complete", map[string]interface{}{
		"total_tasks_processed": spp.performanceMetrics.TotalTasksProcessed,
		"successful_tasks":      spp.performanceMetrics.SuccessfulTasks,
		"failed_tasks":          spp.performanceMetrics.FailedTasks,
		"deduplicated_tasks":    spp.performanceMetrics.DeduplicatedTasks,
	})
}

// TaskHandler represents a task handler function
type TaskHandler func(ctx context.Context, input map[string]interface{}) (interface{}, error)
