package performance

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// ResponseOptimizer handles API response optimization
type ResponseOptimizer struct {
	compressor *gzip.Writer
	config     *CompressionConfig
	pool       *sync.Pool
}

// CompressionConfig contains compression optimization settings
type CompressionConfig struct {
	Level      int      `yaml:"level"`    // 1-9, 6 is default
	MinSize    int      `yaml:"min_size"` // Minimum size to compress
	Types      []string `yaml:"types"`    // Content types to compress
	EnableGzip bool     `yaml:"enable_gzip"`
}

// NewResponseOptimizer creates a new response optimizer
func NewResponseOptimizer(config *CompressionConfig) *ResponseOptimizer {
	pool := &sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}

	return &ResponseOptimizer{
		config: config,
		pool:   pool,
	}
}

// OptimizeResponse optimizes API response for size and performance
func (ro *ResponseOptimizer) OptimizeResponse(data interface{}) ([]byte, error) {
	// Serialize with optimized JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	// Apply compression if beneficial
	if ro.config.EnableGzip && len(jsonData) > ro.config.MinSize {
		return ro.compress(jsonData)
	}

	return jsonData, nil
}

// compress applies gzip compression to data
func (ro *ResponseOptimizer) compress(data []byte) ([]byte, error) {
	buf := ro.pool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		ro.pool.Put(buf)
	}()

	writer, err := gzip.NewWriterLevel(buf, ro.config.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip writer: %w", err)
	}
	defer writer.Close()

	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write compressed data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// ShouldCompress determines if response should be compressed
func (ro *ResponseOptimizer) ShouldCompress(contentType string, size int) bool {
	if !ro.config.EnableGzip || size < ro.config.MinSize {
		return false
	}

	for _, allowedType := range ro.config.Types {
		if contentType == allowedType {
			return true
		}
	}

	return false
}

// ConnectionPool manages optimized HTTP connections
type ConnectionPool struct {
	httpClient *http.Client
	transport  *http.Transport
	config     *PoolConfig
	stats      *PoolStats
	mu         sync.RWMutex
}

// PoolConfig contains connection pool optimization settings
type PoolConfig struct {
	MaxIdleConns          int           `yaml:"max_idle_conns"`
	MaxIdleConnsPerHost   int           `yaml:"max_idle_conns_per_host"`
	IdleConnTimeout       time.Duration `yaml:"idle_conn_timeout"`
	DisableKeepAlives     bool          `yaml:"disable_keep_alives"`
	MaxConnsPerHost       int           `yaml:"max_conns_per_host"`
	ResponseHeaderTimeout time.Duration `yaml:"response_header_timeout"`
	ExpectContinueTimeout time.Duration `yaml:"expect_continue_timeout"`
}

// PoolStats tracks connection pool performance
type PoolStats struct {
	TotalRequests     int64         `json:"total_requests"`
	ActiveConnections int           `json:"active_connections"`
	IdleConnections   int           `json:"idle_connections"`
	AverageLatency    time.Duration `json:"average_latency"`
	ErrorCount        int64         `json:"error_count"`
	mu                sync.RWMutex
}

// NewOptimizedConnectionPool creates a new optimized connection pool
func NewOptimizedConnectionPool(config *PoolConfig) *ConnectionPool {
	transport := &http.Transport{
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		DisableKeepAlives:     config.DisableKeepAlives,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
	}

	return &ConnectionPool{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		transport: transport,
		config:    config,
		stats:     &PoolStats{},
	}
}

// Do performs an HTTP request with performance tracking
func (cp *ConnectionPool) Do(req *http.Request) (*http.Response, error) {
	start := time.Now()

	cp.mu.Lock()
	cp.stats.TotalRequests++
	cp.mu.Unlock()

	resp, err := cp.httpClient.Do(req)

	latency := time.Since(start)

	cp.mu.Lock()
	if err != nil {
		cp.stats.ErrorCount++
	} else {
		// Update average latency
		if cp.stats.AverageLatency == 0 {
			cp.stats.AverageLatency = latency
		} else {
			cp.stats.AverageLatency = (cp.stats.AverageLatency + latency) / 2
		}
	}
	cp.mu.Unlock()

	return resp, err
}

// GetStats returns connection pool statistics
func (cp *ConnectionPool) GetStats() *PoolStats {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	// Get current connection stats from transport
	cp.stats.ActiveConnections = cp.transport.MaxIdleConns
	cp.stats.IdleConnections = cp.transport.MaxIdleConnsPerHost

	return &PoolStats{
		TotalRequests:     cp.stats.TotalRequests,
		ActiveConnections: cp.stats.ActiveConnections,
		IdleConnections:   cp.stats.IdleConnections,
		AverageLatency:    cp.stats.AverageLatency,
		ErrorCount:        cp.stats.ErrorCount,
	}
}

// Close closes the connection pool
func (cp *ConnectionPool) Close() {
	cp.transport.CloseIdleConnections()
}

// AsyncProcessor handles asynchronous processing for heavy operations
type AsyncProcessor struct {
	workers    int
	jobQueue   chan Job
	resultChan chan Result
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	stats      *AsyncStats
	mu         sync.RWMutex
}

// Job represents an asynchronous job
type Job struct {
	ID       string
	Type     string
	Data     interface{}
	Priority int
	Timeout  time.Duration
}

// Result represents the result of an asynchronous job
type Result struct {
	JobID    string
	Data     interface{}
	Error    error
	Duration time.Duration
}

// AsyncStats tracks asynchronous processing statistics
type AsyncStats struct {
	TotalJobs     int64         `json:"total_jobs"`
	CompletedJobs int64         `json:"completed_jobs"`
	FailedJobs    int64         `json:"failed_jobs"`
	AverageTime   time.Duration `json:"average_time"`
	QueueSize     int           `json:"queue_size"`
	ActiveWorkers int           `json:"active_workers"`
}

// NewAsyncProcessor creates a new asynchronous processor
func NewAsyncProcessor(workers int, queueSize int) *AsyncProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	ap := &AsyncProcessor{
		workers:    workers,
		jobQueue:   make(chan Job, queueSize),
		resultChan: make(chan Result, queueSize),
		ctx:        ctx,
		cancel:     cancel,
		stats:      &AsyncStats{},
	}

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		ap.wg.Add(1)
		go ap.worker(i)
	}

	return ap
}

// ProcessBusinessVerification processes business verification asynchronously
func (ap *AsyncProcessor) ProcessBusinessVerification(business BusinessData) (*VerificationResult, error) {
	job := Job{
		ID:       generateID(),
		Type:     "business_verification",
		Data:     business,
		Priority: 1,
		Timeout:  30 * time.Second,
	}

	ap.mu.Lock()
	ap.stats.TotalJobs++
	ap.stats.QueueSize = len(ap.jobQueue)
	ap.mu.Unlock()

	// Send to async queue
	select {
	case ap.jobQueue <- job:
		// Wait for result
		timeout := time.NewTimer(job.Timeout)
		defer timeout.Stop()

		for {
			select {
			case result := <-ap.resultChan:
				if result.JobID == job.ID {
					ap.mu.Lock()
					ap.stats.CompletedJobs++
					if result.Error != nil {
						ap.stats.FailedJobs++
					}
					ap.mu.Unlock()

					if result.Error != nil {
						return nil, result.Error
					}
					return result.Data.(*VerificationResult), nil
				}
			case <-timeout.C:
				return nil, fmt.Errorf("processing timeout for job %s", job.ID)
			case <-ap.ctx.Done():
				return nil, fmt.Errorf("processor shutdown")
			}
		}
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("queue full, unable to process job")
	}
}

// worker processes jobs from the queue
func (ap *AsyncProcessor) worker(workerID int) {
	defer ap.wg.Done()

	for {
		select {
		case job := <-ap.jobQueue:
			ap.mu.Lock()
			ap.stats.ActiveWorkers++
			ap.stats.QueueSize = len(ap.jobQueue)
			ap.mu.Unlock()

			start := time.Now()
			result := ap.processJob(job)
			duration := time.Since(start)

			result.Duration = duration

			// Update average processing time
			ap.mu.Lock()
			if ap.stats.AverageTime == 0 {
				ap.stats.AverageTime = duration
			} else {
				ap.stats.AverageTime = (ap.stats.AverageTime + duration) / 2
			}
			ap.stats.ActiveWorkers--
			ap.mu.Unlock()

			// Send result
			select {
			case ap.resultChan <- result:
			case <-ap.ctx.Done():
				return
			}

		case <-ap.ctx.Done():
			return
		}
	}
}

// processJob processes a single job
func (ap *AsyncProcessor) processJob(job Job) Result {
	// This would typically call your business logic
	// For now, we'll simulate processing
	time.Sleep(100 * time.Millisecond) // Simulate processing time

	switch job.Type {
	case "business_verification":
		business := job.Data.(BusinessData)
		result := &VerificationResult{
			ID:     job.ID,
			Status: "verified",
			Score:  0.95,
		}
		return Result{
			JobID: job.ID,
			Data:  result,
			Error: nil,
		}
	default:
		return Result{
			JobID: job.ID,
			Data:  nil,
			Error: fmt.Errorf("unknown job type: %s", job.Type),
		}
	}
}

// GetStats returns asynchronous processing statistics
func (ap *AsyncProcessor) GetStats() *AsyncStats {
	ap.mu.RLock()
	defer ap.mu.RUnlock()

	return &AsyncStats{
		TotalJobs:     ap.stats.TotalJobs,
		CompletedJobs: ap.stats.CompletedJobs,
		FailedJobs:    ap.stats.FailedJobs,
		AverageTime:   ap.stats.AverageTime,
		QueueSize:     ap.stats.QueueSize,
		ActiveWorkers: ap.stats.ActiveWorkers,
	}
}

// Shutdown gracefully shuts down the async processor
func (ap *AsyncProcessor) Shutdown() {
	ap.cancel()
	ap.wg.Wait()
	close(ap.jobQueue)
	close(ap.resultChan)
}

// PerformanceMonitor tracks API performance metrics
type PerformanceMonitor struct {
	metrics   map[string]*Metric
	mu        sync.RWMutex
	startTime time.Time
}

// Metric tracks a specific performance metric
type Metric struct {
	Name        string
	Count       int64
	TotalTime   time.Duration
	MinTime     time.Duration
	MaxTime     time.Duration
	AverageTime time.Duration
	ErrorCount  int64
	mu          sync.RWMutex
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics:   make(map[string]*Metric),
		startTime: time.Now(),
	}
}

// TrackRequest tracks a request's performance
func (pm *PerformanceMonitor) TrackRequest(endpoint string, method string, duration time.Duration, statusCode int) {
	key := fmt.Sprintf("%s:%s", method, endpoint)

	pm.mu.Lock()
	metric, exists := pm.metrics[key]
	if !exists {
		metric = &Metric{
			Name:    key,
			MinTime: duration,
			MaxTime: duration,
		}
		pm.metrics[key] = metric
	}
	pm.mu.Unlock()

	metric.mu.Lock()
	metric.Count++
	metric.TotalTime += duration

	if duration < metric.MinTime {
		metric.MinTime = duration
	}
	if duration > metric.MaxTime {
		metric.MaxTime = duration
	}

	metric.AverageTime = metric.TotalTime / time.Duration(metric.Count)

	if statusCode >= 400 {
		metric.ErrorCount++
	}
	metric.mu.Unlock()
}

// GetMetrics returns all performance metrics
func (pm *PerformanceMonitor) GetMetrics() map[string]*Metric {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := make(map[string]*Metric)
	for key, metric := range pm.metrics {
		metric.mu.RLock()
		metrics[key] = &Metric{
			Name:        metric.Name,
			Count:       metric.Count,
			TotalTime:   metric.TotalTime,
			MinTime:     metric.MinTime,
			MaxTime:     metric.MaxTime,
			AverageTime: metric.AverageTime,
			ErrorCount:  metric.ErrorCount,
		}
		metric.mu.RUnlock()
	}

	return metrics
}

// GetUptime returns the monitor uptime
func (pm *PerformanceMonitor) GetUptime() time.Duration {
	return time.Since(pm.startTime)
}

// Data structures for business operations
type BusinessData struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Industry string `json:"industry"`
	Address  string `json:"address"`
}

type VerificationResult struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Score  float64 `json:"score"`
}

// Utility functions
func generateID() string {
	return fmt.Sprintf("job-%d", time.Now().UnixNano())
}

// Middleware for performance optimization
func PerformanceMiddleware(monitor *PerformanceMonitor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapper := &responseWriter{ResponseWriter: w, statusCode: 200}

			next.ServeHTTP(wrapper, r)

			duration := time.Since(start)
			monitor.TrackRequest(r.URL.Path, r.Method, duration, wrapper.statusCode)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
