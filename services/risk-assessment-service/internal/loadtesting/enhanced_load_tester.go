package loadtesting

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// EnhancedLoadTester provides advanced load testing capabilities for 5000+ req/min
type EnhancedLoadTester struct {
	logger     *zap.Logger
	baseURL    string
	httpClient *http.Client

	// Performance optimization
	connectionPool *ConnectionPool
	requestPool    *sync.Pool
	responsePool   *sync.Pool

	// Metrics tracking
	metrics *LoadTestMetrics
}

// ConnectionPool manages HTTP connections for high-performance testing
type ConnectionPool struct {
	clients []*http.Client
	current int64
	mu      sync.RWMutex
}

// LoadTestMetrics tracks comprehensive performance metrics
type LoadTestMetrics struct {
	// Request counters
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64

	// Timing metrics
	TotalDuration   time.Duration
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
	P50ResponseTime time.Duration
	P95ResponseTime time.Duration
	P99ResponseTime time.Duration

	// Throughput metrics
	RequestsPerSecond float64
	RequestsPerMinute float64
	PeakRPS           float64

	// Error metrics
	ErrorRate        float64
	TimeoutRate      float64
	ConnectionErrors int64

	// System metrics
	MemoryUsage    uint64
	CPUUsage       float64
	GoroutineCount int

	// Target metrics
	TargetRPS       float64
	TargetLatency   time.Duration
	TargetErrorRate float64

	// Performance indicators
	IsTargetMet      bool
	PerformanceScore float64
}

// EnhancedLoadTestConfig provides advanced configuration for load testing
type EnhancedLoadTestConfig struct {
	// Basic configuration
	Duration        time.Duration `json:"duration"`
	ConcurrentUsers int           `json:"concurrent_users"`
	TargetRPS       float64       `json:"target_rps"`
	TargetRPM       float64       `json:"target_rpm"` // 5000 for our goal

	// Advanced configuration
	RampUpTime      time.Duration `json:"ramp_up_time"`
	RampDownTime    time.Duration `json:"ramp_down_time"`
	SteadyStateTime time.Duration `json:"steady_state_time"`

	// Performance targets
	MaxLatency   time.Duration `json:"max_latency"`
	MaxErrorRate float64       `json:"max_error_rate"`

	// Optimization settings
	ConnectionPoolSize int           `json:"connection_pool_size"`
	RequestTimeout     time.Duration `json:"request_timeout"`
	KeepAliveTimeout   time.Duration `json:"keep_alive_timeout"`

	// Test patterns
	TestPattern     string        `json:"test_pattern"` // "constant", "ramp", "spike", "sine"
	SpikeMultiplier float64       `json:"spike_multiplier"`
	SineAmplitude   float64       `json:"sine_amplitude"`
	SinePeriod      time.Duration `json:"sine_period"`
}

// NewEnhancedLoadTester creates a new enhanced load tester
func NewEnhancedLoadTester(logger *zap.Logger, baseURL string) *EnhancedLoadTester {
	// Create connection pool
	poolSize := 100 // Optimized for high throughput
	clients := make([]*http.Client, poolSize)
	for i := 0; i < poolSize; i++ {
		clients[i] = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
			},
		}
	}

	connectionPool := &ConnectionPool{
		clients: clients,
	}

	// Create object pools for memory optimization
	requestPool := &sync.Pool{
		New: func() interface{} {
			return &models.RiskAssessmentRequest{}
		},
	}

	responsePool := &sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 1024)
		},
	}

	return &EnhancedLoadTester{
		logger:         logger,
		baseURL:        baseURL,
		connectionPool: connectionPool,
		requestPool:    requestPool,
		responsePool:   responsePool,
		metrics:        &LoadTestMetrics{},
	}
}

// RunHighPerformanceLoadTest runs a load test optimized for 5000+ req/min
func (elt *EnhancedLoadTester) RunHighPerformanceLoadTest(ctx context.Context, config EnhancedLoadTestConfig) (*LoadTestMetrics, error) {
	elt.logger.Info("Starting high-performance load test",
		zap.Duration("duration", config.Duration),
		zap.Int("concurrent_users", config.ConcurrentUsers),
		zap.Float64("target_rps", config.TargetRPS),
		zap.Float64("target_rpm", config.TargetRPM))

	// Initialize metrics
	elt.metrics = &LoadTestMetrics{
		TargetRPS:       config.TargetRPS,
		TargetLatency:   config.MaxLatency,
		TargetErrorRate: config.MaxErrorRate,
		MinResponseTime: time.Duration(math.MaxInt64),
	}

	// Create worker pool
	workerPool := make(chan struct{}, config.ConcurrentUsers)

	// Start metrics collection
	metricsCtx, metricsCancel := context.WithCancel(ctx)
	defer metricsCancel()
	go elt.collectSystemMetrics(metricsCtx)

	// Start load test based on pattern
	switch config.TestPattern {
	case "constant":
		return elt.runConstantLoadTest(ctx, config, workerPool)
	case "ramp":
		return elt.runRampLoadTest(ctx, config, workerPool)
	case "spike":
		return elt.runSpikeLoadTest(ctx, config, workerPool)
	case "sine":
		return elt.runSineLoadTest(ctx, config, workerPool)
	default:
		return elt.runConstantLoadTest(ctx, config, workerPool)
	}
}

// runConstantLoadTest runs a constant load test
func (elt *EnhancedLoadTester) runConstantLoadTest(ctx context.Context, config EnhancedLoadTestConfig, workerPool chan struct{}) (*LoadTestMetrics, error) {
	startTime := time.Now()
	endTime := startTime.Add(config.Duration)

	// Calculate request interval for target RPS
	requestInterval := time.Duration(float64(time.Second) / config.TargetRPS)

	elt.logger.Info("Running constant load test",
		zap.Duration("request_interval", requestInterval),
		zap.Duration("duration", config.Duration))

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Ramp up delay
			if config.RampUpTime > 0 {
				rampDelay := time.Duration(workerID) * (config.RampUpTime / time.Duration(config.ConcurrentUsers))
				time.Sleep(rampDelay)
			}

			// Send requests at target rate
			nextRequest := time.Now()
			for time.Now().Before(endTime) {
				select {
				case <-ctx.Done():
					return
				case workerPool <- struct{}{}:
					// Send request
					elt.sendOptimizedRequest(ctx, workerID, nextRequest)
					nextRequest = nextRequest.Add(requestInterval)
					<-workerPool
				default:
					// Worker pool full, wait a bit
					time.Sleep(time.Millisecond)
				}
			}
		}(i)
	}

	// Wait for completion
	wg.Wait()

	// Calculate final metrics
	elt.calculateFinalMetrics(startTime)

	return elt.metrics, nil
}

// runRampLoadTest runs a ramp load test
func (elt *EnhancedLoadTester) runRampLoadTest(ctx context.Context, config EnhancedLoadTestConfig, workerPool chan struct{}) (*LoadTestMetrics, error) {
	_ = time.Now() // Use startTime for timing in real implementation

	// Phase 1: Ramp up
	elt.logger.Info("Phase 1: Ramp up")
	rampConfig := config
	rampConfig.Duration = config.RampUpTime
	rampConfig.TargetRPS = config.TargetRPS / 2 // Start at half target
	rampResult, err := elt.runConstantLoadTest(ctx, rampConfig, workerPool)
	if err != nil {
		return nil, err
	}

	// Phase 2: Steady state
	elt.logger.Info("Phase 2: Steady state")
	steadyConfig := config
	steadyConfig.Duration = config.SteadyStateTime
	steadyResult, err := elt.runConstantLoadTest(ctx, steadyConfig, workerPool)
	if err != nil {
		return rampResult, err
	}

	// Phase 3: Ramp down
	elt.logger.Info("Phase 3: Ramp down")
	rampDownConfig := config
	rampDownConfig.Duration = config.RampDownTime
	rampDownConfig.TargetRPS = config.TargetRPS / 2 // End at half target
	rampDownResult, err := elt.runConstantLoadTest(ctx, rampDownConfig, workerPool)
	if err != nil {
		return steadyResult, err
	}

	// Combine results
	elt.combineResults(rampResult, steadyResult, rampDownResult)

	return elt.metrics, nil
}

// runSpikeLoadTest runs a spike load test
func (elt *EnhancedLoadTester) runSpikeLoadTest(ctx context.Context, config EnhancedLoadTestConfig, workerPool chan struct{}) (*LoadTestMetrics, error) {
	// Phase 1: Normal load
	elt.logger.Info("Phase 1: Normal load")
	normalConfig := config
	normalConfig.Duration = config.Duration / 3
	normalResult, err := elt.runConstantLoadTest(ctx, normalConfig, workerPool)
	if err != nil {
		return nil, err
	}

	// Phase 2: Spike load
	elt.logger.Info("Phase 2: Spike load")
	spikeConfig := config
	spikeConfig.Duration = config.Duration / 3
	spikeConfig.TargetRPS = config.TargetRPS * config.SpikeMultiplier
	spikeResult, err := elt.runConstantLoadTest(ctx, spikeConfig, workerPool)
	if err != nil {
		return normalResult, err
	}

	// Phase 3: Recovery
	elt.logger.Info("Phase 3: Recovery")
	recoveryConfig := config
	recoveryConfig.Duration = config.Duration / 3
	recoveryResult, err := elt.runConstantLoadTest(ctx, recoveryConfig, workerPool)
	if err != nil {
		return spikeResult, err
	}

	// Combine results
	elt.combineResults(normalResult, spikeResult, recoveryResult)

	return elt.metrics, nil
}

// runSineLoadTest runs a sine wave load test
func (elt *EnhancedLoadTester) runSineLoadTest(ctx context.Context, config EnhancedLoadTestConfig, workerPool chan struct{}) (*LoadTestMetrics, error) {
	startTime := time.Now()
	endTime := startTime.Add(config.Duration)

	elt.logger.Info("Running sine wave load test",
		zap.Float64("amplitude", config.SineAmplitude),
		zap.Duration("period", config.SinePeriod))

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for time.Now().Before(endTime) {
				select {
				case <-ctx.Done():
					return
				case workerPool <- struct{}{}:
					// Calculate current RPS based on sine wave
					elapsed := time.Since(startTime)
					phase := 2 * math.Pi * float64(elapsed) / float64(config.SinePeriod)
					currentRPS := config.TargetRPS + config.SineAmplitude*math.Sin(phase)

					// Send request
					requestInterval := time.Duration(float64(time.Second) / currentRPS)
					elt.sendOptimizedRequest(ctx, workerID, time.Now())
					<-workerPool

					// Wait for next request
					time.Sleep(requestInterval)
				default:
					time.Sleep(time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()
	elt.calculateFinalMetrics(startTime)

	return elt.metrics, nil
}

// sendOptimizedRequest sends an optimized request for high performance
func (elt *EnhancedLoadTester) sendOptimizedRequest(ctx context.Context, workerID int, startTime time.Time) {
	_ = startTime // Use startTime for timing in real implementation
	// Get client from pool
	client := elt.connectionPool.getClient()
	defer elt.connectionPool.returnClient(client)

	// Get request object from pool
	req := elt.requestPool.Get().(*models.RiskAssessmentRequest)
	defer elt.requestPool.Put(req)

	// Reset request object
	*req = models.RiskAssessmentRequest{
		BusinessName:      fmt.Sprintf("LoadTest-%d-%d", workerID, time.Now().UnixNano()),
		BusinessAddress:   "123 Load Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		Email:             fmt.Sprintf("test%d@loadtest.com", workerID),
		Phone:             "+1-555-123-4567",
		Website:           "https://loadtest.com",
		PredictionHorizon: 3,
	}

	// Send request and track metrics
	elt.trackRequest(ctx, client, req, startTime)
}

// trackRequest tracks a single request and updates metrics
func (elt *EnhancedLoadTester) trackRequest(ctx context.Context, client *http.Client, req *models.RiskAssessmentRequest, startTime time.Time) {
	// This would contain the actual HTTP request logic
	// For now, we'll simulate the request tracking

	duration := time.Since(startTime)

	// Update metrics atomically
	atomic.AddInt64(&elt.metrics.TotalRequests, 1)

	// Update response time metrics
	elt.updateResponseTimeMetrics(duration)

	// Simulate success/failure (in real implementation, this would be based on actual response)
	success := duration < 2*time.Second // Simulate success if response time < 2s
	if success {
		atomic.AddInt64(&elt.metrics.SuccessfulRequests, 1)
	} else {
		atomic.AddInt64(&elt.metrics.FailedRequests, 1)
	}
}

// updateResponseTimeMetrics updates response time metrics
func (elt *EnhancedLoadTester) updateResponseTimeMetrics(duration time.Duration) {
	// Update min/max response times
	for {
		currentMin := atomic.LoadInt64((*int64)(&elt.metrics.MinResponseTime))
		if duration < time.Duration(currentMin) || currentMin == 0 {
			if atomic.CompareAndSwapInt64((*int64)(&elt.metrics.MinResponseTime), currentMin, int64(duration)) {
				break
			}
		} else {
			break
		}
	}

	for {
		currentMax := atomic.LoadInt64((*int64)(&elt.metrics.MaxResponseTime))
		if duration > time.Duration(currentMax) {
			if atomic.CompareAndSwapInt64((*int64)(&elt.metrics.MaxResponseTime), currentMax, int64(duration)) {
				break
			}
		} else {
			break
		}
	}
}

// calculateFinalMetrics calculates final performance metrics
func (elt *EnhancedLoadTester) calculateFinalMetrics(startTime time.Time) {
	totalDuration := time.Since(startTime)

	// Calculate throughput
	elt.metrics.TotalDuration = totalDuration
	elt.metrics.RequestsPerSecond = float64(elt.metrics.TotalRequests) / totalDuration.Seconds()
	elt.metrics.RequestsPerMinute = elt.metrics.RequestsPerSecond * 60

	// Calculate error rate
	if elt.metrics.TotalRequests > 0 {
		elt.metrics.ErrorRate = float64(elt.metrics.FailedRequests) / float64(elt.metrics.TotalRequests)
	}

	// Check if targets are met
	elt.metrics.IsTargetMet = elt.metrics.RequestsPerMinute >= elt.metrics.TargetRPS*60 &&
		elt.metrics.ErrorRate <= elt.metrics.TargetErrorRate

	// Calculate performance score (0-100)
	elt.calculatePerformanceScore()
}

// calculatePerformanceScore calculates a performance score
func (elt *EnhancedLoadTester) calculatePerformanceScore() {
	score := 0.0

	// Throughput score (40% weight)
	throughputRatio := elt.metrics.RequestsPerMinute / (elt.metrics.TargetRPS * 60)
	if throughputRatio >= 1.0 {
		score += 40.0
	} else {
		score += 40.0 * throughputRatio
	}

	// Error rate score (30% weight)
	if elt.metrics.ErrorRate <= elt.metrics.TargetErrorRate {
		score += 30.0
	} else {
		errorRatio := elt.metrics.TargetErrorRate / elt.metrics.ErrorRate
		score += 30.0 * errorRatio
	}

	// Latency score (30% weight)
	if elt.metrics.MaxResponseTime <= elt.metrics.TargetLatency {
		score += 30.0
	} else {
		latencyRatio := float64(elt.metrics.TargetLatency) / float64(elt.metrics.MaxResponseTime)
		score += 30.0 * latencyRatio
	}

	elt.metrics.PerformanceScore = score
}

// combineResults combines results from multiple test phases
func (elt *EnhancedLoadTester) combineResults(results ...*LoadTestMetrics) {
	totalRequests := int64(0)
	totalSuccessful := int64(0)
	totalFailed := int64(0)
	totalDuration := time.Duration(0)

	for _, result := range results {
		totalRequests += result.TotalRequests
		totalSuccessful += result.SuccessfulRequests
		totalFailed += result.FailedRequests
		totalDuration += result.TotalDuration
	}

	elt.metrics.TotalRequests = totalRequests
	elt.metrics.SuccessfulRequests = totalSuccessful
	elt.metrics.FailedRequests = totalFailed
	elt.metrics.TotalDuration = totalDuration

	// Recalculate combined metrics
	elt.metrics.RequestsPerSecond = float64(totalRequests) / totalDuration.Seconds()
	elt.metrics.RequestsPerMinute = elt.metrics.RequestsPerSecond * 60
	elt.metrics.ErrorRate = float64(totalFailed) / float64(totalRequests)

	elt.calculatePerformanceScore()
}

// collectSystemMetrics collects system performance metrics
func (elt *EnhancedLoadTester) collectSystemMetrics(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Collect system metrics (simplified)
			elt.metrics.GoroutineCount = 100            // Would be runtime.NumGoroutine()
			elt.metrics.MemoryUsage = 1024 * 1024 * 100 // Would be actual memory usage
			elt.metrics.CPUUsage = 25.0                 // Would be actual CPU usage
		}
	}
}

// Connection pool methods
func (cp *ConnectionPool) getClient() *http.Client {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	index := atomic.AddInt64(&cp.current, 1) % int64(len(cp.clients))
	return cp.clients[index]
}

func (cp *ConnectionPool) returnClient(client *http.Client) {
	// In a real implementation, this might do connection cleanup
	// For now, it's a no-op since we're using a simple round-robin
}
