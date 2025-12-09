package handlers

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestRequestQueue_EnqueueDequeue(t *testing.T) {
	queue := NewRequestQueue(10)
	
	// Test enqueue
	req := &queuedRequest{
		req:       &ClassificationRequest{RequestID: "test-1"},
		ctx:       context.Background(),
		response:  make(chan *ClassificationResponse, 1),
		errChan:   make(chan error, 1),
		startTime: time.Now(),
	}
	
	err := queue.Enqueue(req)
	if err != nil {
		t.Fatalf("Failed to enqueue: %v", err)
	}
	
	if queue.Size() != 1 {
		t.Errorf("Expected queue size 1, got %d", queue.Size())
	}
	
	// Test dequeue
	dequeued, ok := queue.Dequeue()
	if !ok {
		t.Fatal("Failed to dequeue")
	}
	
	if dequeued.req.RequestID != "test-1" {
		t.Errorf("Expected request ID 'test-1', got '%s'", dequeued.req.RequestID)
	}
	
	if queue.Size() != 0 {
		t.Errorf("Expected queue size 0, got %d", queue.Size())
	}
}

func TestRequestQueue_Full(t *testing.T) {
	queue := NewRequestQueue(2)
	
	// Fill queue
	req1 := &queuedRequest{
		req:       &ClassificationRequest{RequestID: "test-1"},
		ctx:       context.Background(),
		response:  make(chan *ClassificationResponse, 1),
		errChan:   make(chan error, 1),
		startTime: time.Now(),
	}
	req2 := &queuedRequest{
		req:       &ClassificationRequest{RequestID: "test-2"},
		ctx:       context.Background(),
		response:  make(chan *ClassificationResponse, 1),
		errChan:   make(chan error, 1),
		startTime: time.Now(),
	}
	req3 := &queuedRequest{
		req:       &ClassificationRequest{RequestID: "test-3"},
		ctx:       context.Background(),
		response:  make(chan *ClassificationResponse, 1),
		errChan:   make(chan error, 1),
		startTime: time.Now(),
	}
	
	if err := queue.Enqueue(req1); err != nil {
		t.Fatalf("Failed to enqueue req1: %v", err)
	}
	if err := queue.Enqueue(req2); err != nil {
		t.Fatalf("Failed to enqueue req2: %v", err)
	}
	
	// Try to enqueue third request - should fail
	err := queue.Enqueue(req3)
	if err == nil {
		t.Error("Expected error when enqueueing to full queue")
	}
	
	if queue.Size() != 2 {
		t.Errorf("Expected queue size 2, got %d", queue.Size())
	}
}

func TestRequestQueue_Concurrent(t *testing.T) {
	queue := NewRequestQueue(100)
	done := make(chan bool, 10)
	
	// Concurrent enqueue
	for i := 0; i < 10; i++ {
		go func(id int) {
			req := &queuedRequest{
				req:       &ClassificationRequest{RequestID: string(rune(id))},
				ctx:       context.Background(),
				response:  make(chan *ClassificationResponse, 1),
				errChan:   make(chan error, 1),
				startTime: time.Now(),
			}
			queue.Enqueue(req)
			done <- true
		}(i)
	}
	
	// Wait for all enqueues
	for i := 0; i < 10; i++ {
		<-done
	}
	
	if queue.Size() != 10 {
		t.Errorf("Expected queue size 10, got %d", queue.Size())
	}
	
	// Concurrent dequeue
	dequeued := make(chan *queuedRequest, 10)
	for i := 0; i < 10; i++ {
		go func() {
			req, ok := queue.Dequeue()
			if ok {
				dequeued <- req
			}
			done <- true
		}()
	}
	
	// Wait for all dequeues
	for i := 0; i < 10; i++ {
		<-done
	}
	
	close(dequeued)
	count := 0
	for range dequeued {
		count++
	}
	
	if count != 10 {
		t.Errorf("Expected 10 dequeued requests, got %d", count)
	}
	
	if queue.Size() != 0 {
		t.Errorf("Expected queue size 0, got %d", queue.Size())
	}
}

func TestRequestQueue_Size(t *testing.T) {
	queue := NewRequestQueue(10)
	
	if queue.Size() != 0 {
		t.Errorf("Expected initial size 0, got %d", queue.Size())
	}
	
	// Add requests
	for i := 0; i < 5; i++ {
		req := &queuedRequest{
			req:       &ClassificationRequest{RequestID: string(rune(i))},
			ctx:       context.Background(),
			response:  make(chan *ClassificationResponse, 1),
			errChan:   make(chan error, 1),
			startTime: time.Now(),
		}
		queue.Enqueue(req)
	}
	
	if queue.Size() != 5 {
		t.Errorf("Expected size 5, got %d", queue.Size())
	}
	
	// Remove requests
	for i := 0; i < 3; i++ {
		queue.Dequeue()
	}
	
	if queue.Size() != 2 {
		t.Errorf("Expected size 2, got %d", queue.Size())
	}
}

func TestWorkerPool_StartStop(t *testing.T) {
	logger := zap.NewNop()
	queue := NewRequestQueue(10)
	
	// Create a minimal handler for testing
	handler := &ClassificationHandler{
		logger: logger,
	}
	
	pool := NewWorkerPool(2, queue, handler, logger)
	
	// Start pool
	pool.Start()
	
	// Give workers time to start
	time.Sleep(100 * time.Millisecond)
	
	// Stop pool
	pool.Stop()
	
	// Pool should be stopped
	// (No easy way to verify this without exposing internals, but Stop() should not block)
}

func TestWorkerPool_ProcessRequest(t *testing.T) {
	logger := zap.NewNop()
	queue := NewRequestQueue(10)
	
	// Create a minimal handler with a mock processClassification
	handler := &ClassificationHandler{
		logger: logger,
	}
	
	pool := NewWorkerPool(1, queue, handler, logger)
	pool.Start()
	defer pool.Stop()
	
	// Create a request
	req := &queuedRequest{
		req: &ClassificationRequest{
			RequestID:   "test-1",
			BusinessName: "Test Business",
		},
		ctx:       context.Background(),
		response:  make(chan *ClassificationResponse, 1),
		errChan:   make(chan error, 1),
		startTime: time.Now(),
	}
	
	// Enqueue request
	if err := queue.Enqueue(req); err != nil {
		t.Fatalf("Failed to enqueue: %v", err)
	}
	
	// Wait for processing (with timeout)
	select {
	case <-req.response:
		// Request processed (even if it fails, we just check it was processed)
		t.Log("Request processed")
	case <-req.errChan:
		// Error occurred (expected since we don't have full handler setup)
		t.Log("Request processing returned error (expected)")
	case <-time.After(5 * time.Second):
		t.Error("Request processing timed out")
	}
}

