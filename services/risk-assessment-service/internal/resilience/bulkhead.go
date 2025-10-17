package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Bulkhead implements the bulkhead pattern for resource isolation
type Bulkhead struct {
	logger *zap.Logger
	mu     sync.RWMutex
	stats  *BulkheadStats
	config *BulkheadConfig
	pools  map[string]*ResourcePool
}

// BulkheadStats represents statistics for bulkhead operations
type BulkheadStats struct {
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	RejectedRequests   int64         `json:"rejected_requests"`
	AverageWaitTime    time.Duration `json:"average_wait_time"`
	MaxWaitTime        time.Duration `json:"max_wait_time"`
	ActivePools        int           `json:"active_pools"`
	LastRequest        time.Time     `json:"last_request"`
}

// BulkheadConfig represents configuration for bulkhead
type BulkheadConfig struct {
	DefaultMaxConcurrency int           `json:"default_max_concurrency"`
	DefaultMaxQueueSize   int           `json:"default_max_queue_size"`
	DefaultTimeout        time.Duration `json:"default_timeout"`
	EnableMetrics         bool          `json:"enable_metrics"`
	EnableLogging         bool          `json:"enable_logging"`
}

// ResourcePool represents a resource pool for a specific service
type ResourcePool struct {
	Name             string        `json:"name"`
	MaxConcurrency   int           `json:"max_concurrency"`
	MaxQueueSize     int           `json:"max_queue_size"`
	Timeout          time.Duration `json:"timeout"`
	ActiveRequests   int           `json:"active_requests"`
	QueuedRequests   int           `json:"queued_requests"`
	TotalRequests    int64         `json:"total_requests"`
	FailedRequests   int64         `json:"failed_requests"`
	RejectedRequests int64         `json:"rejected_requests"`
	AverageWaitTime  time.Duration `json:"average_wait_time"`
	LastRequest      time.Time     `json:"last_request"`
	mu               sync.RWMutex
}

// BulkheadRequest represents a request to be processed by bulkhead
type BulkheadRequest struct {
	ID        string                 `json:"id"`
	Service   string                 `json:"service"`
	Operation string                 `json:"operation"`
	Data      map[string]interface{} `json:"data"`
	Timeout   time.Duration          `json:"timeout"`
	Priority  int                    `json:"priority"`
	CreatedAt time.Time              `json:"created_at"`
}

// BulkheadResponse represents a response from bulkhead
type BulkheadResponse struct {
	ID          string                 `json:"id"`
	Success     bool                   `json:"success"`
	Result      map[string]interface{} `json:"result"`
	Error       string                 `json:"error,omitempty"`
	WaitTime    time.Duration          `json:"wait_time"`
	ProcessTime time.Duration          `json:"process_time"`
	CreatedAt   time.Time              `json:"created_at"`
}

// NewBulkhead creates a new bulkhead instance
func NewBulkhead(config *BulkheadConfig, logger *zap.Logger) *Bulkhead {
	if config == nil {
		config = &BulkheadConfig{
			DefaultMaxConcurrency: 10,
			DefaultMaxQueueSize:   100,
			DefaultTimeout:        30 * time.Second,
			EnableMetrics:         true,
			EnableLogging:         true,
		}
	}

	return &Bulkhead{
		logger: logger,
		stats:  &BulkheadStats{},
		config: config,
		pools:  make(map[string]*ResourcePool),
	}
}

// CreatePool creates a new resource pool
func (b *Bulkhead) CreatePool(name string, maxConcurrency int, maxQueueSize int, timeout time.Duration) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.pools[name]; exists {
		return fmt.Errorf("pool %s already exists", name)
	}

	pool := &ResourcePool{
		Name:             name,
		MaxConcurrency:   maxConcurrency,
		MaxQueueSize:     maxQueueSize,
		Timeout:          timeout,
		ActiveRequests:   0,
		QueuedRequests:   0,
		TotalRequests:    0,
		FailedRequests:   0,
		RejectedRequests: 0,
		LastRequest:      time.Now(),
	}

	b.pools[name] = pool
	b.stats.ActivePools = len(b.pools)

	b.logger.Info("Resource pool created",
		zap.String("pool_name", name),
		zap.Int("max_concurrency", maxConcurrency),
		zap.Int("max_queue_size", maxQueueSize),
		zap.Duration("timeout", timeout))

	return nil
}

// Execute executes a request through the bulkhead
func (b *Bulkhead) Execute(ctx context.Context, request *BulkheadRequest, processor func(context.Context, *BulkheadRequest) (*BulkheadResponse, error)) (*BulkheadResponse, error) {
	start := time.Now()

	// Get or create pool
	pool, err := b.getOrCreatePool(request.Service)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool: %w", err)
	}

	// Check if we can accept the request
	if !b.canAcceptRequest(pool) {
		pool.mu.Lock()
		pool.RejectedRequests++
		pool.mu.Unlock()

		b.mu.Lock()
		b.stats.RejectedRequests++
		b.mu.Unlock()

		return nil, fmt.Errorf("request rejected: pool %s is at capacity", request.Service)
	}

	// Wait for available slot
	waitStart := time.Now()
	if err := b.waitForSlot(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to wait for slot: %w", err)
	}
	waitTime := time.Since(waitStart)

	// Update wait time statistics
	pool.mu.Lock()
	pool.AverageWaitTime = (pool.AverageWaitTime + waitTime) / 2
	pool.mu.Unlock()

	b.mu.Lock()
	b.stats.AverageWaitTime = (b.stats.AverageWaitTime + waitTime) / 2
	if waitTime > b.stats.MaxWaitTime {
		b.stats.MaxWaitTime = waitTime
	}
	b.mu.Unlock()

	// Acquire slot
	b.acquireSlot(pool)
	defer b.releaseSlot(pool)

	// Process request
	processStart := time.Now()
	response, err := processor(ctx, request)
	processTime := time.Since(processStart)

	// Update statistics
	b.updateStats(pool, err == nil, waitTime, processTime)

	if err != nil {
		return nil, fmt.Errorf("request processing failed: %w", err)
	}

	// Set response metadata
	response.WaitTime = waitTime
	response.ProcessTime = processTime

	b.logger.Debug("Request processed through bulkhead",
		zap.String("request_id", request.ID),
		zap.String("service", request.Service),
		zap.Duration("wait_time", waitTime),
		zap.Duration("process_time", processTime),
		zap.Duration("total_time", time.Since(start)))

	return response, nil
}

// GetStats returns bulkhead statistics
func (b *Bulkhead) GetStats() *BulkheadStats {
	b.mu.RLock()
	defer b.mu.RUnlock()

	stats := *b.stats
	stats.ActivePools = len(b.pools)
	return &stats
}

// GetPoolStats returns statistics for a specific pool
func (b *Bulkhead) GetPoolStats(poolName string) (*ResourcePool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	pool, exists := b.pools[poolName]
	if !exists {
		return nil, fmt.Errorf("pool %s not found", poolName)
	}

	pool.mu.RLock()
	defer pool.mu.RUnlock()

	// Return a copy to avoid race conditions (excluding mutex)
	poolCopy := ResourcePool{
		Name:             pool.Name,
		MaxConcurrency:   pool.MaxConcurrency,
		MaxQueueSize:     pool.MaxQueueSize,
		Timeout:          pool.Timeout,
		ActiveRequests:   pool.ActiveRequests,
		QueuedRequests:   pool.QueuedRequests,
		TotalRequests:    pool.TotalRequests,
		FailedRequests:   pool.FailedRequests,
		RejectedRequests: pool.RejectedRequests,
		AverageWaitTime:  pool.AverageWaitTime,
		LastRequest:      pool.LastRequest,
	}
	return &poolCopy, nil
}

// GetAllPoolStats returns statistics for all pools
func (b *Bulkhead) GetAllPoolStats() map[string]*ResourcePool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	stats := make(map[string]*ResourcePool)
	for name, pool := range b.pools {
		pool.mu.RLock()
		poolCopy := ResourcePool{
			Name:             pool.Name,
			MaxConcurrency:   pool.MaxConcurrency,
			MaxQueueSize:     pool.MaxQueueSize,
			Timeout:          pool.Timeout,
			ActiveRequests:   pool.ActiveRequests,
			QueuedRequests:   pool.QueuedRequests,
			TotalRequests:    pool.TotalRequests,
			FailedRequests:   pool.FailedRequests,
			RejectedRequests: pool.RejectedRequests,
			AverageWaitTime:  pool.AverageWaitTime,
			LastRequest:      pool.LastRequest,
		}
		pool.mu.RUnlock()
		stats[name] = &poolCopy
	}

	return stats
}

// ResetStats resets all statistics
func (b *Bulkhead) ResetStats() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.stats = &BulkheadStats{}

	for _, pool := range b.pools {
		pool.mu.Lock()
		pool.TotalRequests = 0
		pool.FailedRequests = 0
		pool.RejectedRequests = 0
		pool.AverageWaitTime = 0
		pool.mu.Unlock()
	}

	b.logger.Info("Bulkhead statistics reset")
}

// Helper methods

func (b *Bulkhead) getOrCreatePool(serviceName string) (*ResourcePool, error) {
	b.mu.RLock()
	pool, exists := b.pools[serviceName]
	b.mu.RUnlock()

	if exists {
		return pool, nil
	}

	// Create default pool
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check again in case another goroutine created it
	if pool, exists := b.pools[serviceName]; exists {
		return pool, nil
	}

	pool = &ResourcePool{
		Name:             serviceName,
		MaxConcurrency:   b.config.DefaultMaxConcurrency,
		MaxQueueSize:     b.config.DefaultMaxQueueSize,
		Timeout:          b.config.DefaultTimeout,
		ActiveRequests:   0,
		QueuedRequests:   0,
		TotalRequests:    0,
		FailedRequests:   0,
		RejectedRequests: 0,
		LastRequest:      time.Now(),
	}

	b.pools[serviceName] = pool
	b.stats.ActivePools = len(b.pools)

	b.logger.Info("Default resource pool created",
		zap.String("pool_name", serviceName),
		zap.Int("max_concurrency", pool.MaxConcurrency),
		zap.Int("max_queue_size", pool.MaxQueueSize),
		zap.Duration("timeout", pool.Timeout))

	return pool, nil
}

func (b *Bulkhead) canAcceptRequest(pool *ResourcePool) bool {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	// Check if we can accept more requests
	return pool.ActiveRequests < pool.MaxConcurrency || pool.QueuedRequests < pool.MaxQueueSize
}

func (b *Bulkhead) waitForSlot(ctx context.Context, pool *ResourcePool) error {
	timeout := pool.Timeout
	if timeout == 0 {
		timeout = b.config.DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			pool.mu.RLock()
			canAccept := pool.ActiveRequests < pool.MaxConcurrency
			pool.mu.RUnlock()

			if canAccept {
				return nil
			}
		}
	}
}

func (b *Bulkhead) acquireSlot(pool *ResourcePool) {
	pool.mu.Lock()
	pool.ActiveRequests++
	pool.TotalRequests++
	pool.LastRequest = time.Now()
	pool.mu.Unlock()

	b.mu.Lock()
	b.stats.TotalRequests++
	b.stats.LastRequest = time.Now()
	b.mu.Unlock()
}

func (b *Bulkhead) releaseSlot(pool *ResourcePool) {
	pool.mu.Lock()
	pool.ActiveRequests--
	pool.mu.Unlock()
}

func (b *Bulkhead) updateStats(pool *ResourcePool, success bool, waitTime, processTime time.Duration) {
	pool.mu.Lock()
	if success {
		pool.AverageWaitTime = (pool.AverageWaitTime + waitTime) / 2
	} else {
		pool.FailedRequests++
	}
	pool.mu.Unlock()

	b.mu.Lock()
	if success {
		b.stats.SuccessfulRequests++
	} else {
		b.stats.FailedRequests++
	}
	b.mu.Unlock()
}

// BulkheadManager manages multiple bulkheads
type BulkheadManager struct {
	bulkheads map[string]*Bulkhead
	logger    *zap.Logger
	mu        sync.RWMutex
}

// NewBulkheadManager creates a new bulkhead manager
func NewBulkheadManager(logger *zap.Logger) *BulkheadManager {
	return &BulkheadManager{
		bulkheads: make(map[string]*Bulkhead),
		logger:    logger,
	}
}

// GetBulkhead gets or creates a bulkhead for a service
func (bm *BulkheadManager) GetBulkhead(serviceName string, config *BulkheadConfig) *Bulkhead {
	bm.mu.RLock()
	bulkhead, exists := bm.bulkheads[serviceName]
	bm.mu.RUnlock()

	if exists {
		return bulkhead
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Check again in case another goroutine created it
	if bulkhead, exists := bm.bulkheads[serviceName]; exists {
		return bulkhead
	}

	bulkhead = NewBulkhead(config, bm.logger)
	bm.bulkheads[serviceName] = bulkhead

	bm.logger.Info("Bulkhead created for service",
		zap.String("service_name", serviceName))

	return bulkhead
}

// GetAllStats returns statistics for all bulkheads
func (bm *BulkheadManager) GetAllStats() map[string]*BulkheadStats {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	stats := make(map[string]*BulkheadStats)
	for name, bulkhead := range bm.bulkheads {
		stats[name] = bulkhead.GetStats()
	}

	return stats
}

// ResetAllStats resets statistics for all bulkheads
func (bm *BulkheadManager) ResetAllStats() {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	for _, bulkhead := range bm.bulkheads {
		bulkhead.ResetStats()
	}

	bm.logger.Info("All bulkhead statistics reset")
}
