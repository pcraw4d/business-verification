package jobs

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// mockJob is a simple mock job for testing
type mockJob struct {
	id         string
	merchantID string
	jobType    string
	status     JobStatus
	processFn  func(ctx context.Context) error
	mu         sync.RWMutex
}

func newMockJob(id, merchantID, jobType string, processFn func(ctx context.Context) error) *mockJob {
	return &mockJob{
		id:         id,
		merchantID: merchantID,
		jobType:    jobType,
		status:     StatusPending,
		processFn:  processFn,
	}
}

func (m *mockJob) GetID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.id
}

func (m *mockJob) GetMerchantID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.merchantID
}

func (m *mockJob) GetType() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.jobType
}

func (m *mockJob) GetStatus() JobStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *mockJob) SetStatus(status JobStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status = status
}

func (m *mockJob) Process(ctx context.Context) error {
	if m.processFn != nil {
		return m.processFn(ctx)
	}
	m.SetStatus(StatusCompleted)
	return nil
}

func TestJobProcessor_NewJobProcessor(t *testing.T) {
	processor := NewJobProcessor(5, 100, zaptest.NewLogger(t))
	
	assert.NotNil(t, processor)
	assert.Equal(t, 5, processor.workers)
	assert.Equal(t, 100, cap(processor.jobQueue))
	assert.False(t, processor.IsRunning())
}

func TestJobProcessor_Start(t *testing.T) {
	processor := NewJobProcessor(2, 10, zaptest.NewLogger(t))
	
	processor.Start()
	
	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)
	
	assert.True(t, processor.IsRunning())
	
	// Stop should complete quickly
	stopDone := make(chan bool, 1)
	go func() {
		processor.Stop()
		stopDone <- true
	}()
	
	// Wait for stop with timeout
	select {
	case <-stopDone:
		// Successfully stopped
		assert.False(t, processor.IsRunning())
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() timed out after 2 seconds")
	}
}

func TestJobProcessor_Enqueue(t *testing.T) {
	processor := NewJobProcessor(2, 10, zaptest.NewLogger(t))
	processor.Start()
	defer processor.Stop()

	job := newMockJob("job_1", "merchant_123", "test", nil)
	
	err := processor.Enqueue(job)
	
	assert.NoError(t, err)
	assert.Equal(t, 1, processor.GetQueueSize())
}

func TestJobProcessor_Enqueue_NotRunning(t *testing.T) {
	processor := NewJobProcessor(2, 10, zaptest.NewLogger(t))
	// Don't start the processor
	
	job := newMockJob("job_1", "merchant_123", "test", nil)
	
	err := processor.Enqueue(job)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestJobProcessor_Enqueue_QueueFull(t *testing.T) {
	processor := NewJobProcessor(1, 1, zaptest.NewLogger(t))
	processor.Start()
	defer processor.Stop()

	// Fill the queue
	job1 := newMockJob("job_1", "merchant_123", "test", func(ctx context.Context) error {
		time.Sleep(2 * time.Second) // Long-running job
		return nil
	})
	processor.Enqueue(job1)
	
	// Try to enqueue another job (queue is full)
	job2 := newMockJob("job_2", "merchant_456", "test", nil)
	err := processor.Enqueue(job2)
	
	// Should timeout after 5 seconds
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue is full")
}

func TestJobProcessor_ProcessJob(t *testing.T) {
	processor := NewJobProcessor(2, 10, zaptest.NewLogger(t))
	processor.Start()
	defer processor.Stop()

	processed := make(chan bool, 1)
	job := newMockJob("job_1", "merchant_123", "test", func(ctx context.Context) error {
		processed <- true
		return nil
	})
	
	err := processor.Enqueue(job)
	require.NoError(t, err)
	
	// Wait for job to be processed
	select {
	case <-processed:
		// Job was processed successfully
		assert.Equal(t, StatusCompleted, job.GetStatus())
	case <-time.After(5 * time.Second):
		t.Fatal("Job was not processed within timeout")
	}
}

func TestJobProcessor_ProcessJob_Error(t *testing.T) {
	processor := NewJobProcessor(2, 10, zaptest.NewLogger(t))
	processor.Start()
	defer processor.Stop()

	job := newMockJob("job_1", "merchant_123", "test", func(ctx context.Context) error {
		return assert.AnError
	})
	
	err := processor.Enqueue(job)
	require.NoError(t, err)
	
	// Give it time to process
	time.Sleep(500 * time.Millisecond)
	
	// Job should have failed
	assert.Equal(t, StatusFailed, job.GetStatus())
}

func TestJobProcessor_Stop(t *testing.T) {
	processor := NewJobProcessor(2, 10, zaptest.NewLogger(t))
	processor.Start()
	
	assert.True(t, processor.IsRunning())
	
	processor.Stop()
	
	// Give it a moment to stop
	time.Sleep(100 * time.Millisecond)
	
	assert.False(t, processor.IsRunning())
}

func TestJobProcessor_ConcurrentJobs(t *testing.T) {
	processor := NewJobProcessor(3, 20, zaptest.NewLogger(t))
	processor.Start()
	defer processor.Stop()

	numJobs := 5
	processed := make(chan string, numJobs)
	
	// Enqueue multiple jobs
	for i := 0; i < numJobs; i++ {
		jobID := fmt.Sprintf("job_%d", i)
		job := newMockJob(jobID, fmt.Sprintf("merchant_%d", i), "test", func(ctx context.Context) error {
			processed <- jobID
			return nil
		})
		
		err := processor.Enqueue(job)
		require.NoError(t, err)
	}
	
	// Wait for all jobs to complete
	completed := make(map[string]bool)
	timeout := time.After(10 * time.Second)
	
	for len(completed) < numJobs {
		select {
		case jobID := <-processed:
			completed[jobID] = true
		case <-timeout:
			t.Fatalf("Not all jobs completed. Completed: %d/%d", len(completed), numJobs)
		}
	}
	
	assert.Equal(t, numJobs, len(completed))
}

func TestJobProcessor_GetQueueSize(t *testing.T) {
	processor := NewJobProcessor(2, 10, zaptest.NewLogger(t))
	processor.Start()
	defer processor.Stop()

	assert.Equal(t, 0, processor.GetQueueSize())
	
	// Enqueue a job that takes time to process
	job := newMockJob("job_1", "merchant_123", "test", func(ctx context.Context) error {
		time.Sleep(1 * time.Second)
		return nil
	})
	
	processor.Enqueue(job)
	
	// Queue size should be 1 (or 0 if already picked up by worker)
	// Give it a moment
	time.Sleep(50 * time.Millisecond)
	
	size := processor.GetQueueSize()
	assert.GreaterOrEqual(t, size, 0)
	assert.LessOrEqual(t, size, 1)
}

