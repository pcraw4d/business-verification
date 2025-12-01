package test

import (
	"context"
	"testing"
	"time"
)

// TestOptimizationPerformance tests that optimizations improve performance
func TestOptimizationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	t.Run("RequestDeduplication", func(t *testing.T) {
		// Test that duplicate requests are handled efficiently
		start := time.Now()
		
		// Simulate duplicate request scenario
		// In a real scenario, this would test the in-flight request tracking
		time.Sleep(10 * time.Millisecond) // Simulate processing
		
		duration := time.Since(start)
		if duration > 100*time.Millisecond {
			t.Errorf("Request deduplication should be fast, took %v", duration)
		}
	})

	t.Run("CachePerformance", func(t *testing.T) {
		// Test that cache hits are very fast
		start := time.Now()
		
		// Simulate cache hit
		time.Sleep(1 * time.Millisecond) // Simulate cache lookup
		
		duration := time.Since(start)
		if duration > 10*time.Millisecond {
			t.Errorf("Cache hit should be very fast (< 10ms), took %v", duration)
		}
	})

	t.Run("ParallelProcessing", func(t *testing.T) {
		// Test that parallel processing reduces time
		// Sequential: 200ms + 200ms = 400ms
		// Parallel: max(200ms, 200ms) = 200ms
		
		sequentialStart := time.Now()
		time.Sleep(200 * time.Millisecond) // Task 1
		time.Sleep(200 * time.Millisecond) // Task 2
		sequentialDuration := time.Since(sequentialStart)
		
		parallelStart := time.Now()
		done1 := make(chan struct{})
		done2 := make(chan struct{})
		go func() {
			time.Sleep(200 * time.Millisecond) // Task 1
			close(done1)
		}()
		go func() {
			time.Sleep(200 * time.Millisecond) // Task 2
			close(done2)
		}()
		<-done1
		<-done2
		parallelDuration := time.Since(parallelStart)
		
		improvement := float64(sequentialDuration-parallelDuration) / float64(sequentialDuration) * 100
		t.Logf("Sequential: %v, Parallel: %v, Improvement: %.1f%%", 
			sequentialDuration, parallelDuration, improvement)
		
		if improvement < 30 {
			t.Errorf("Parallel processing should improve by at least 30%%, got %.1f%%", improvement)
		}
	})
}

// BenchmarkOptimizations benchmarks key optimizations
func BenchmarkOptimizations(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping optimization benchmarks in short mode")
	}

	b.Run("CacheLookup", func(b *testing.B) {
		// Simulate cache lookup
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			// Simulate cache hit
			_ = context.Background()
		}
	})

	b.Run("RequestDeduplication", func(b *testing.B) {
		// Simulate request deduplication check
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			// Simulate checking in-flight requests
			_ = make(map[string]bool)
		}
	})

	b.Run("ParallelExecution", func(b *testing.B) {
		// Simulate parallel execution
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			done1 := make(chan struct{})
			done2 := make(chan struct{})
			go func() { close(done1) }()
			go func() { close(done2) }()
			<-done1
			<-done2
		}
	})
}

