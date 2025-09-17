package classification

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// ParallelProcessingConfig holds configuration for parallel processing
type ParallelProcessingConfig struct {
	// MaxConcurrency limits the number of concurrent goroutines
	MaxConcurrency int

	// Timeout for individual operations
	OperationTimeout time.Duration

	// Enable parallel processing
	Enabled bool

	// Batch size for processing multiple items
	BatchSize int

	// Retry configuration
	MaxRetries int
	RetryDelay time.Duration
}

// DefaultParallelProcessingConfig returns the default configuration
func DefaultParallelProcessingConfig() *ParallelProcessingConfig {
	return &ParallelProcessingConfig{
		MaxConcurrency:   runtime.NumCPU() * 2, // Use 2x CPU cores
		OperationTimeout: 30 * time.Second,
		Enabled:          true,
		BatchSize:        10,
		MaxRetries:       3,
		RetryDelay:       1 * time.Second,
	}
}

// ParallelProcessor provides parallel processing capabilities
type ParallelProcessor struct {
	config    *ParallelProcessingConfig
	semaphore chan struct{} // Semaphore to limit concurrency
}

// NewParallelProcessor creates a new parallel processor
func NewParallelProcessor(config *ParallelProcessingConfig) *ParallelProcessor {
	if config == nil {
		config = DefaultParallelProcessingConfig()
	}

	return &ParallelProcessor{
		config:    config,
		semaphore: make(chan struct{}, config.MaxConcurrency),
	}
}

// ProcessInParallel processes items in parallel with concurrency control
// Note: This is a simplified version without generics for Go 1.18 compatibility
func (p *ParallelProcessor) ProcessInParallel(
	ctx context.Context,
	items []interface{},
	processor func(context.Context, interface{}) (interface{}, error),
) ([]interface{}, []error) {
	if !p.config.Enabled || len(items) == 0 {
		// Process sequentially if parallel processing is disabled
		results := make([]interface{}, len(items))
		errors := make([]error, len(items))

		for i, item := range items {
			result, err := processor(ctx, item)
			results[i] = result
			errors[i] = err
		}

		return results, errors
	}

	// Process in parallel with concurrency control
	results := make([]interface{}, len(items))
	errors := make([]error, len(items))
	var wg sync.WaitGroup

	for i, item := range items {
		wg.Add(1)
		go func(index int, item interface{}) {
			defer wg.Done()

			// Acquire semaphore
			p.semaphore <- struct{}{}
			defer func() { <-p.semaphore }()

			// Create timeout context
			itemCtx, cancel := context.WithTimeout(ctx, p.config.OperationTimeout)
			defer cancel()

			// Process with retry logic
			result, err := p.processWithRetry(itemCtx, item, processor)
			results[index] = result
			errors[index] = err
		}(i, item)
	}

	wg.Wait()
	return results, errors
}

// processWithRetry processes an item with retry logic
func (p *ParallelProcessor) processWithRetry(
	ctx context.Context,
	item interface{},
	processor func(context.Context, interface{}) (interface{}, error),
) (interface{}, error) {
	var result interface{}
	var err error

	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(p.config.RetryDelay):
			}
		}

		result, err = processor(ctx, item)
		if err == nil {
			return result, nil
		}
	}

	return result, err
}

// ProcessBatchesInParallel processes items in batches in parallel
func (p *ParallelProcessor) ProcessBatchesInParallel(
	ctx context.Context,
	items []interface{},
	processor func(context.Context, []interface{}) ([]interface{}, error),
) ([]interface{}, error) {
	if !p.config.Enabled || len(items) == 0 {
		// Process as single batch if parallel processing is disabled
		return processor(ctx, items)
	}

	// Split items into batches
	batches := p.createBatches(items)

	// Process batches in parallel
	results := make([][]interface{}, len(batches))
	errors := make([]error, len(batches))
	var wg sync.WaitGroup

	for i, batch := range batches {
		wg.Add(1)
		go func(index int, batch []interface{}) {
			defer wg.Done()

			// Acquire semaphore
			p.semaphore <- struct{}{}
			defer func() { <-p.semaphore }()

			// Create timeout context
			batchCtx, cancel := context.WithTimeout(ctx, p.config.OperationTimeout)
			defer cancel()

			// Process batch
			batchResults, err := processor(batchCtx, batch)
			results[index] = batchResults
			errors[index] = err
		}(i, batch)
	}

	wg.Wait()

	// Combine results
	var allResults []interface{}
	var allErrors []error

	for i, batchResults := range results {
		allResults = append(allResults, batchResults...)
		if errors[i] != nil {
			allErrors = append(allErrors, errors[i])
		}
	}

	// Return combined error if any batch failed
	if len(allErrors) > 0 {
		return allResults, allErrors[0] // Return first error for simplicity
	}

	return allResults, nil
}

// createBatches splits items into batches of the configured size
func (p *ParallelProcessor) createBatches(items []interface{}) [][]interface{} {
	if len(items) <= p.config.BatchSize {
		return [][]interface{}{items}
	}

	batches := make([][]interface{}, 0, (len(items)+p.config.BatchSize-1)/p.config.BatchSize)

	for i := 0; i < len(items); i += p.config.BatchSize {
		end := i + p.config.BatchSize
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}

	return batches
}

// PerformanceMetrics tracks parallel processing performance
type PerformanceMetrics struct {
	TotalItems     int
	ProcessedItems int
	FailedItems    int
	TotalTime      time.Duration
	AverageTime    time.Duration
	Throughput     float64 // items per second
}

// CalculateMetrics calculates performance metrics
func (p *ParallelProcessor) CalculateMetrics(
	totalItems int,
	processedItems int,
	failedItems int,
	totalTime time.Duration,
) *PerformanceMetrics {
	avgTime := time.Duration(0)
	throughput := 0.0

	if processedItems > 0 {
		avgTime = totalTime / time.Duration(processedItems)
		throughput = float64(processedItems) / totalTime.Seconds()
	}

	return &PerformanceMetrics{
		TotalItems:     totalItems,
		ProcessedItems: processedItems,
		FailedItems:    failedItems,
		TotalTime:      totalTime,
		AverageTime:    avgTime,
		Throughput:     throughput,
	}
}

// ParallelProcessingStats provides statistics about parallel processing
type ParallelProcessingStats struct {
	Config           *ParallelProcessingConfig
	ActiveGoroutines int
	QueuedItems      int
	ProcessedItems   int64
	FailedItems      int64
	TotalTime        time.Duration
}

// GetStats returns current parallel processing statistics
func (p *ParallelProcessor) GetStats() *ParallelProcessingStats {
	return &ParallelProcessingStats{
		Config:           p.config,
		ActiveGoroutines: len(p.semaphore),
		QueuedItems:      cap(p.semaphore) - len(p.semaphore),
		ProcessedItems:   0, // Would be tracked in real implementation
		FailedItems:      0, // Would be tracked in real implementation
		TotalTime:        0, // Would be tracked in real implementation
	}
}
