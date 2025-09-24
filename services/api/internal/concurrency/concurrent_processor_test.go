package concurrency

import (
	"context"
	"testing"
	"time"

	"kyb-platform/internal/observability"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func TestConcurrentProcessor(t *testing.T) {
	logger := zap.NewNop()

	config := &ConcurrentProcessorConfig{
		MaxWorkers:              5,
		WorkerTimeout:           5 * time.Second,
		QueueSize:               100,
		MaxConcurrentOps:        10,
		ResourceTimeout:         5 * time.Second,
		EnableDeadlockDetection: true,
		EnableMetrics:           true,
		MetricsInterval:         1 * time.Second,
	}

	processor := NewConcurrentProcessor(config, logger)

	// Test initialization
	if processor == nil {
		t.Fatal("ConcurrentProcessor should not be nil")
	}

	// Test starting the processor
	if err := processor.Start(); err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}

	// Test processing a simple request
	request := &ConcurrentRequest{
		ID:        "test-request-1",
		Type:      "test",
		Data:      "test data",
		Priority:  1,
		Timeout:   5 * time.Second,
		CreatedAt: time.Now(),
	}

	ctx := context.Background()
	response, err := processor.ProcessRequest(ctx, request)
	if err != nil {
		t.Fatalf("Failed to process request: %v", err)
	}

	if response == nil {
		t.Fatal("Response should not be nil")
	}

	if response.RequestID != request.ID {
		t.Errorf("Expected request ID %s, got %s", request.ID, response.RequestID)
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got %s", response.Status)
	}

	// Test statistics
	stats := processor.GetStats()
	if stats == nil {
		t.Fatal("Stats should not be nil")
	}

	if stats.TotalRequests != 1 {
		t.Errorf("Expected 1 total request, got %d", stats.TotalRequests)
	}

	if stats.SuccessfulRequests != 1 {
		t.Errorf("Expected 1 successful request, got %d", stats.SuccessfulRequests)
	}

	// Test stopping the processor
	processor.Stop()

	// Wait a bit for cleanup
	time.Sleep(100 * time.Millisecond)
}

func TestResourceManager(t *testing.T) {
	logger := zap.NewNop()
	config := &ResourceManagerConfig{
		MaxConcurrentOps: 10,
		ResourceTimeout:  30 * time.Second,
	}

	obsLogger := observability.NewLogger(logger)
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	rm := NewResourceManager(config, obsLogger, tracer)

	// Test starting the resource manager
	err := rm.Start()
	if err != nil {
		t.Fatalf("Failed to start resource manager: %v", err)
	}
	defer rm.Stop()

	// Test getting resource metrics
	stats := rm.GetStats()
	if stats == nil {
		t.Error("Expected stats to be non-nil")
	}

	// Basic stats validation
	if stats.CPUUtilization < 0 || stats.CPUUtilization > 100 {
		t.Errorf("CPU utilization should be between 0 and 100, got %f", stats.CPUUtilization)
	}
}

func TestThreadSafeDataStructures(t *testing.T) {
	ds := NewThreadSafeDataStructures()

	// Test thread-safe map
	ds.stringMap.Set("key1", "value1")
	value, exists := ds.stringMap.Get("key1")
	if !exists {
		t.Fatal("Key should exist")
	}
	if value != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}

	// Test thread-safe slice
	ds.stringSlice.Append("item1")
	ds.stringSlice.Append("item2")
	if ds.stringSlice.Length() != 2 {
		t.Errorf("Expected length 2, got %d", ds.stringSlice.Length())
	}

	// Test thread-safe queue
	ds.requestQueue.Enqueue(&ConcurrentRequest{ID: "test"})
	if ds.requestQueue.Size() != 1 {
		t.Errorf("Expected queue size 1, got %d", ds.requestQueue.Size())
	}

	// Test thread-safe counter
	ds.counter.Increment()
	ds.counter.Increment()
	if ds.counter.Get() != 2 {
		t.Errorf("Expected counter value 2, got %d", ds.counter.Get())
	}
}

func TestSynchronizationManager(t *testing.T) {
	logger := zap.NewNop()
	config := &SynchronizationManagerConfig{
		EnableDeadlockDetection: true,
	}

	sm := NewSynchronizationManager(config, logger)

	// Test acquiring a lock
	ctx := context.Background()
	lockRequest := &LockRequest{
		ResourceID: "test-resource",
		Timeout:    5 * time.Second,
		Priority:   1,
	}

	lock, err := sm.AcquireLock(ctx, lockRequest)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	if lock == nil {
		t.Fatal("Lock should not be nil")
	}

	if lock.ResourceID != "test-resource" {
		t.Errorf("Expected resource ID 'test-resource', got %s", lock.ResourceID)
	}

	// Test checking if resource is locked
	if !sm.IsLocked("test-resource") {
		t.Error("Resource should be locked")
	}

	// Test releasing the lock
	err = sm.ReleaseLock(lock)
	if err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}

	if sm.IsLocked("test-resource") {
		t.Error("Resource should not be locked after release")
	}
}

func TestDeadlockDetector(t *testing.T) {
	logger := zap.NewNop()
	config := &DeadlockDetectorConfig{
		DetectionInterval: 1 * time.Second,
	}

	dd := NewDeadlockDetector(config, logger)

	// Test starting the detector
	err := dd.Start()
	if err != nil {
		t.Fatalf("Failed to start deadlock detector: %v", err)
	}

	// Test adding a deadlock
	deadlock := &DeadlockInfo{
		ID:         "test-deadlock",
		Resources:  []string{"resource1", "resource2"},
		Processes:  []string{"process1", "process2"},
		DetectedAt: time.Now(),
	}

	dd.AddDeadlock(deadlock)

	// Test detecting deadlocks
	deadlocks, err := dd.DetectDeadlocks()
	if err != nil {
		t.Fatalf("Failed to detect deadlocks: %v", err)
	}

	if len(deadlocks) != 1 {
		t.Errorf("Expected 1 deadlock, got %d", len(deadlocks))
	}

	// Test resolving a deadlock
	err = dd.ResolveDeadlock(deadlock)
	if err != nil {
		t.Fatalf("Failed to resolve deadlock: %v", err)
	}

	// Test stopping the detector
	err = dd.Stop()
	if err != nil {
		t.Fatalf("Failed to stop deadlock detector: %v", err)
	}
}

func TestConcurrentRequestHandler(t *testing.T) {
	logger := zap.NewNop()

	// Create resource manager
	resourceConfig := &ResourceManagerConfig{
		MaxConcurrentOps: 10,
		ResourceTimeout:  30 * time.Second,
	}
	obsLogger := observability.NewLogger(logger)
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	resourceMgr := NewResourceManager(resourceConfig, obsLogger, tracer)

	// Create request handler config
	requestConfig := &ConcurrentRequestHandlerConfig{
		MaxWorkers:    2,
		WorkerTimeout: 5 * time.Second,
		QueueSize:     10,
	}

	// Create a simple processor
	processor := &testProcessor{}

	handler := NewConcurrentRequestHandler(requestConfig, logger, resourceMgr, processor)

	// Test starting the handler
	err := handler.Start()
	if err != nil {
		t.Fatalf("Failed to start request handler: %v", err)
	}

	// Test processing a request
	request := &ConcurrentRequest{
		ID:        "test-request",
		Type:      "test",
		Data:      "test data",
		Priority:  1,
		Timeout:   5 * time.Second,
		CreatedAt: time.Now(),
	}

	ctx := context.Background()
	response, err := handler.ProcessRequest(ctx, request)
	if err != nil {
		t.Fatalf("Failed to process request: %v", err)
	}

	if response == nil {
		t.Fatal("Response should not be nil")
	}

	if response.RequestID != request.ID {
		t.Errorf("Expected request ID %s, got %s", request.ID, response.RequestID)
	}

	// Test statistics
	stats := handler.GetStats()
	if stats == nil {
		t.Fatal("Stats should not be nil")
	}

	if stats.TotalProcessed != 1 {
		t.Errorf("Expected 1 processed request, got %d", stats.TotalProcessed)
	}

	// Test stopping the handler
	err = handler.Stop()
	if err != nil {
		t.Fatalf("Failed to stop request handler: %v", err)
	}
}

// testProcessor is a simple processor for testing
type testProcessor struct{}

func (p *testProcessor) Process(ctx context.Context, request *ConcurrentRequest) (*ConcurrentResponse, error) {
	return &ConcurrentResponse{
		RequestID:   request.ID,
		Status:      "success",
		Data:        request.Data,
		CompletedAt: time.Now(),
	}, nil
}
