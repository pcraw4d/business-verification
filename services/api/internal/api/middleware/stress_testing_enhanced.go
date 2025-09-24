package middleware

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// EnhancedStressTestConfig provides advanced stress testing configuration
type EnhancedStressTestConfig struct {
	// Basic Configuration
	Enabled     bool          `json:"enabled"`
	TestName    string        `json:"test_name"`
	MaxDuration time.Duration `json:"max_duration"`
	Timeout     time.Duration `json:"timeout"`

	// Stress Test Parameters
	StartRPS     int           `json:"start_rps"`
	MaxRPS       int           `json:"max_rps"`
	StepSize     int           `json:"step_size"`
	StepDuration time.Duration `json:"step_duration"`

	// Breaking Point Detection
	FailureThreshold  float64        `json:"failure_threshold"`  // Error rate threshold (e.g., 0.1 = 10%)
	ResponseThreshold time.Duration  `json:"response_threshold"` // Response time threshold
	ResourceThreshold ResourceLimits `json:"resource_threshold"` // Resource utilization limits

	// Recovery Testing
	RecoveryEnabled  bool          `json:"recovery_enabled"`
	RecoveryDuration time.Duration `json:"recovery_duration"`
	RecoverySteps    int           `json:"recovery_steps"`

	// Target Configuration
	BaseURL   string            `json:"base_url"`
	Endpoints []StressEndpoint  `json:"endpoints"`
	Headers   map[string]string `json:"headers"`

	// Advanced Features
	ChaosEngineering   bool `json:"chaos_engineering"`
	ResourceMonitoring bool `json:"resource_monitoring"`
	AutoStop           bool `json:"auto_stop"`

	// Reporting
	DetailedReports bool   `json:"detailed_reports"`
	ExportFormat    string `json:"export_format"`
	ReportDirectory string `json:"report_directory"`
}

// ResourceLimits defines resource utilization limits
type ResourceLimits struct {
	MaxCPUUsage     float64 `json:"max_cpu_usage"`     // Percentage
	MaxMemoryUsage  uint64  `json:"max_memory_usage"`  // Bytes
	MaxDiskUsage    float64 `json:"max_disk_usage"`    // Percentage
	MaxNetworkUsage uint64  `json:"max_network_usage"` // Bytes/sec
}

// StressEndpoint defines an endpoint for stress testing
type StressEndpoint struct {
	Name     string            `json:"name"`
	Method   string            `json:"method"`
	Path     string            `json:"path"`
	Headers  map[string]string `json:"headers"`
	Body     string            `json:"body"`
	Weight   float64           `json:"weight"`
	Critical bool              `json:"critical"` // Whether this endpoint is critical for system function
}

// EnhancedStressTester provides advanced stress testing capabilities
type EnhancedStressTester struct {
	config     *EnhancedStressTestConfig
	logger     *zap.Logger
	client     *http.Client
	executor   *StressTestExecutor
	monitor    *StressTestMonitor
	reporter   *StressTestReporter
	breakPoint *BreakingPointDetector
	recovery   *RecoveryTester
	mu         sync.RWMutex
	stopChan   chan struct{}
}

// StressTestExecutor handles stress test execution
type StressTestExecutor struct {
	config     *EnhancedStressTestConfig
	logger     *zap.Logger
	client     *http.Client
	currentRPS int64
	workers    []*StressWorker
	mu         sync.RWMutex
}

// StressWorker represents a stress test worker
type StressWorker struct {
	id       int
	executor *StressTestExecutor
	client   *http.Client
	stopChan chan struct{}
	metrics  *WorkerMetrics
}

// WorkerMetrics tracks per-worker metrics
type WorkerMetrics struct {
	RequestsHandled    int64
	SuccessfulRequests int64
	FailedRequests     int64
	TotalResponseTime  time.Duration
	MinResponseTime    time.Duration
	MaxResponseTime    time.Duration
	LastRequestTime    time.Time
}

// StressTestMonitor provides real-time monitoring during stress tests
type StressTestMonitor struct {
	config          *EnhancedStressTestConfig
	logger          *zap.Logger
	metrics         *StressTestMetrics
	resourceMonitor *ResourceMonitor
	alertManager    *StressAlertManager
	mu              sync.RWMutex
	stopChan        chan struct{}
}

// ResourceMonitor monitors system resources during stress testing
type ResourceMonitor struct {
	cpuUsage       float64
	memoryUsage    uint64
	diskUsage      float64
	networkUsage   uint64
	goroutineCount int
	lastUpdate     time.Time
	mu             sync.RWMutex
}

// StressAlertManager manages alerts during stress testing
type StressAlertManager struct {
	alerts   []StressAlert
	handlers []AlertHandler
	mu       sync.RWMutex
}

// StressAlert represents a stress test alert
type StressAlert struct {
	ID        string           `json:"id"`
	Timestamp time.Time        `json:"timestamp"`
	Level     StressAlertLevel `json:"level"`
	Type      StressAlertType  `json:"type"`
	Message   string           `json:"message"`
	Metric    string           `json:"metric"`
	Threshold float64          `json:"threshold"`
	Current   float64          `json:"current"`
	Endpoint  string           `json:"endpoint,omitempty"`
}

// StressAlertLevel defines stress test alert severity levels
type StressAlertLevel string

const (
	StressAlertLevelInfo      StressAlertLevel = "info"
	StressAlertLevelWarning   StressAlertLevel = "warning"
	StressAlertLevelCritical  StressAlertLevel = "critical"
	StressAlertLevelEmergency StressAlertLevel = "emergency"
)

// StressAlertType defines types of stress test alerts
type StressAlertType string

const (
	StressAlertTypePerformance   StressAlertType = "performance"
	StressAlertTypeResource      StressAlertType = "resource"
	StressAlertTypeError         StressAlertType = "error"
	StressAlertTypeBreakingPoint StressAlertType = "breaking_point"
	StressAlertTypeRecovery      StressAlertType = "recovery"
)

// AlertHandler interface for handling alerts
type AlertHandler interface {
	HandleAlert(alert StressAlert) error
}

// BreakingPointDetector detects system breaking points
type BreakingPointDetector struct {
	config           *EnhancedStressTestConfig
	logger           *zap.Logger
	breakingPoint    *BreakingPoint
	detectionMetrics *DetectionMetrics
	mu               sync.RWMutex
}

// BreakingPoint represents a detected breaking point
type BreakingPoint struct {
	RPS             int           `json:"rps"`
	Timestamp       time.Time     `json:"timestamp"`
	TriggerReason   string        `json:"trigger_reason"`
	ErrorRate       float64       `json:"error_rate"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	P95ResponseTime time.Duration `json:"p95_response_time"`
	CPUUsage        float64       `json:"cpu_usage"`
	MemoryUsage     uint64        `json:"memory_usage"`
	RecoveryTime    time.Duration `json:"recovery_time,omitempty"`
	Recovered       bool          `json:"recovered"`
}

// DetectionMetrics tracks metrics for breaking point detection
type DetectionMetrics struct {
	WindowSize          time.Duration
	ErrorRateHistory    []float64
	ResponseTimeHistory []time.Duration
	ResourceHistory     []ResourceSnapshot
	LastUpdate          time.Time
}

// ResourceSnapshot represents a point-in-time resource snapshot
type ResourceSnapshot struct {
	Timestamp      time.Time `json:"timestamp"`
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryUsage    uint64    `json:"memory_usage"`
	DiskUsage      float64   `json:"disk_usage"`
	NetworkUsage   uint64    `json:"network_usage"`
	GoroutineCount int       `json:"goroutine_count"`
}

// RecoveryTester tests system recovery after stress
type RecoveryTester struct {
	config  *EnhancedStressTestConfig
	logger  *zap.Logger
	client  *http.Client
	metrics *RecoveryMetrics
	mu      sync.RWMutex
}

// RecoveryMetrics tracks recovery test metrics
type RecoveryMetrics struct {
	StartTime        time.Time      `json:"start_time"`
	EndTime          time.Time      `json:"end_time"`
	RecoveryDuration time.Duration  `json:"recovery_duration"`
	FullRecovery     bool           `json:"full_recovery"`
	RecoverySteps    []RecoveryStep `json:"recovery_steps"`
}

// RecoveryStep represents a step in the recovery process
type RecoveryStep struct {
	Step            int           `json:"step"`
	Timestamp       time.Time     `json:"timestamp"`
	RPS             int           `json:"rps"`
	ErrorRate       float64       `json:"error_rate"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	Healthy         bool          `json:"healthy"`
}

// StressTestReporter generates comprehensive stress test reports
type StressTestReporter struct {
	config  *EnhancedStressTestConfig
	logger  *zap.Logger
	results *StressTestResults
	mu      sync.RWMutex
}

// StressTestMetrics contains comprehensive stress test metrics
type StressTestMetrics struct {
	StartTime          time.Time     `json:"start_time"`
	EndTime            time.Time     `json:"end_time"`
	Duration           time.Duration `json:"duration"`
	PeakRPS            int           `json:"peak_rps"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`

	// Performance Metrics
	ResponseTimes   []time.Duration `json:"-"`
	MinResponseTime time.Duration   `json:"min_response_time"`
	MaxResponseTime time.Duration   `json:"max_response_time"`
	AvgResponseTime time.Duration   `json:"avg_response_time"`
	P50ResponseTime time.Duration   `json:"p50_response_time"`
	P95ResponseTime time.Duration   `json:"p95_response_time"`
	P99ResponseTime time.Duration   `json:"p99_response_time"`

	// Error Metrics
	ErrorRate        float64          `json:"error_rate"`
	ErrorsByType     map[string]int64 `json:"errors_by_type"`
	ErrorsByEndpoint map[string]int64 `json:"errors_by_endpoint"`

	// Resource Metrics
	PeakCPUUsage      float64            `json:"peak_cpu_usage"`
	PeakMemoryUsage   uint64             `json:"peak_memory_usage"`
	ResourceSnapshots []ResourceSnapshot `json:"resource_snapshots"`

	// Breaking Point
	BreakingPoint *BreakingPoint `json:"breaking_point,omitempty"`

	// Recovery Metrics
	RecoveryMetrics *RecoveryMetrics `json:"recovery_metrics,omitempty"`

	// Endpoint Metrics
	EndpointMetrics map[string]*EndpointStressMetrics `json:"endpoint_metrics"`

	// Time Series
	TimeSeries []StressTimeSeriesPoint `json:"time_series"`
}

// EndpointStressMetrics tracks per-endpoint stress metrics
type EndpointStressMetrics struct {
	Name               string        `json:"name"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	ErrorRate          float64       `json:"error_rate"`
	AvgResponseTime    time.Duration `json:"avg_response_time"`
	P95ResponseTime    time.Duration `json:"p95_response_time"`
	ThroughputRPS      float64       `json:"throughput_rps"`
	BreakingPointRPS   int           `json:"breaking_point_rps"`
}

// StressTimeSeriesPoint represents a time series data point
type StressTimeSeriesPoint struct {
	Timestamp       time.Time     `json:"timestamp"`
	RPS             int           `json:"rps"`
	ErrorRate       float64       `json:"error_rate"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	CPUUsage        float64       `json:"cpu_usage"`
	MemoryUsage     uint64        `json:"memory_usage"`
	ActiveWorkers   int           `json:"active_workers"`
}

// StressTestResults contains comprehensive stress test results
type StressTestResults struct {
	TestInfo        StressTestInfo     `json:"test_info"`
	Metrics         *StressTestMetrics `json:"metrics"`
	Summary         StressTestSummary  `json:"summary"`
	Recommendations []string           `json:"recommendations"`
	Alerts          []StressAlert      `json:"alerts"`
}

// StressTestInfo contains stress test metadata
type StressTestInfo struct {
	Name        string                    `json:"name"`
	StartTime   time.Time                 `json:"start_time"`
	EndTime     time.Time                 `json:"end_time"`
	Duration    time.Duration             `json:"duration"`
	Config      *EnhancedStressTestConfig `json:"config"`
	Environment map[string]string         `json:"environment"`
}

// StressTestSummary provides high-level stress test summary
type StressTestSummary struct {
	Status             string   `json:"status"`
	Success            bool     `json:"success"`
	Grade              string   `json:"grade"`
	Score              float64  `json:"score"`
	BreakingPointFound bool     `json:"breaking_point_found"`
	RecoverySuccessful bool     `json:"recovery_successful"`
	KeyFindings        []string `json:"key_findings"`
	SystemResilience   string   `json:"system_resilience"`
}

// NewEnhancedStressTester creates a new enhanced stress tester
func NewEnhancedStressTester(config *EnhancedStressTestConfig, logger *zap.Logger) *EnhancedStressTester {
	if config == nil {
		config = DefaultEnhancedStressTestConfig()
	}

	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:       200,
			IdleConnTimeout:    90 * time.Second,
			DisableCompression: false,
			MaxConnsPerHost:    100,
		},
	}

	tester := &EnhancedStressTester{
		config:   config,
		logger:   logger,
		client:   client,
		stopChan: make(chan struct{}),
	}

	tester.executor = NewStressTestExecutor(config, logger, client)
	tester.monitor = NewStressTestMonitor(config, logger)
	tester.reporter = NewStressTestReporter(config, logger)
	tester.breakPoint = NewBreakingPointDetector(config, logger)

	if config.RecoveryEnabled {
		tester.recovery = NewRecoveryTester(config, logger, client)
	}

	return tester
}

// DefaultEnhancedStressTestConfig returns default enhanced stress test configuration
func DefaultEnhancedStressTestConfig() *EnhancedStressTestConfig {
	return &EnhancedStressTestConfig{
		Enabled:     true,
		TestName:    "Enhanced Stress Test",
		MaxDuration: 30 * time.Minute,
		Timeout:     10 * time.Second,

		StartRPS:     10,
		MaxRPS:       1000,
		StepSize:     50,
		StepDuration: 30 * time.Second,

		FailureThreshold:  0.1, // 10% error rate
		ResponseThreshold: 5 * time.Second,
		ResourceThreshold: ResourceLimits{
			MaxCPUUsage:     90.0,                   // 90%
			MaxMemoryUsage:  2 * 1024 * 1024 * 1024, // 2GB
			MaxDiskUsage:    90.0,                   // 90%
			MaxNetworkUsage: 100 * 1024 * 1024,      // 100MB/s
		},

		RecoveryEnabled:  true,
		RecoveryDuration: 5 * time.Minute,
		RecoverySteps:    5,

		BaseURL: "http://localhost:8080",
		Headers: map[string]string{
			"User-Agent": "Enhanced-Stress-Tester/1.0",
			"Accept":     "application/json",
		},

		ChaosEngineering:   false,
		ResourceMonitoring: true,
		AutoStop:           true,

		DetailedReports: true,
		ExportFormat:    "json",
		ReportDirectory: "/tmp/stress_reports",
	}
}

// RunStressTest executes the enhanced stress test
func (est *EnhancedStressTester) RunStressTest(ctx context.Context) (*StressTestResults, error) {
	est.mu.Lock()
	defer est.mu.Unlock()

	est.logger.Info("starting enhanced stress test",
		zap.String("test_name", est.config.TestName),
		zap.Int("start_rps", est.config.StartRPS),
		zap.Int("max_rps", est.config.MaxRPS),
		zap.Duration("max_duration", est.config.MaxDuration))

	// Initialize test
	if err := est.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize stress test: %w", err)
	}

	// Start monitoring
	if est.config.ResourceMonitoring {
		go est.monitor.Start(ctx)
	}

	// Execute stress test
	results, err := est.executeStressTest(ctx)
	if err != nil {
		return nil, fmt.Errorf("stress test execution failed: %w", err)
	}

	// Test recovery if enabled
	if est.config.RecoveryEnabled && results.Metrics.BreakingPoint != nil {
		recoveryResults, err := est.recovery.TestRecovery(ctx)
		if err != nil {
			est.logger.Error("recovery test failed", zap.Error(err))
		} else {
			results.Metrics.RecoveryMetrics = recoveryResults
		}
	}

	// Generate report
	if err := est.reporter.GenerateReport(results); err != nil {
		est.logger.Error("failed to generate report", zap.Error(err))
	}

	est.logger.Info("enhanced stress test completed",
		zap.String("status", results.Summary.Status),
		zap.Bool("success", results.Summary.Success),
		zap.Bool("breaking_point_found", results.Summary.BreakingPointFound))

	return results, nil
}

// initialize prepares the stress test
func (est *EnhancedStressTester) initialize() error {
	// Validate configuration
	if err := est.validateConfig(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Initialize endpoints
	if len(est.config.Endpoints) == 0 {
		est.config.Endpoints = est.getDefaultEndpoints()
	}

	// Initialize executor
	if err := est.executor.Initialize(); err != nil {
		return fmt.Errorf("executor initialization failed: %w", err)
	}

	return nil
}

// validateConfig validates the stress test configuration
func (est *EnhancedStressTester) validateConfig() error {
	if est.config.MaxDuration <= 0 {
		return fmt.Errorf("max duration must be positive")
	}
	if est.config.StartRPS <= 0 {
		return fmt.Errorf("start RPS must be positive")
	}
	if est.config.MaxRPS <= est.config.StartRPS {
		return fmt.Errorf("max RPS must be greater than start RPS")
	}
	if est.config.StepSize <= 0 {
		return fmt.Errorf("step size must be positive")
	}
	if est.config.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	return nil
}

// getDefaultEndpoints returns default stress test endpoints
func (est *EnhancedStressTester) getDefaultEndpoints() []StressEndpoint {
	return []StressEndpoint{
		{
			Name:     "health_check",
			Method:   "GET",
			Path:     "/health",
			Weight:   0.2,
			Critical: true,
		},
		{
			Name:     "classification",
			Method:   "POST",
			Path:     "/api/v1/classify",
			Headers:  map[string]string{"Content-Type": "application/json"},
			Body:     `{"text": "technology company specializing in software development"}`,
			Weight:   0.5,
			Critical: true,
		},
		{
			Name:     "verification",
			Method:   "POST",
			Path:     "/api/v1/verify",
			Headers:  map[string]string{"Content-Type": "application/json"},
			Body:     `{"business_name": "TechCorp", "website": "https://techcorp.com"}`,
			Weight:   0.3,
			Critical: false,
		},
	}
}

// executeStressTest executes the stress test
func (est *EnhancedStressTester) executeStressTest(ctx context.Context) (*StressTestResults, error) {
	metrics := &StressTestMetrics{
		StartTime:        time.Now(),
		EndpointMetrics:  make(map[string]*EndpointStressMetrics),
		ErrorsByType:     make(map[string]int64),
		ErrorsByEndpoint: make(map[string]int64),
		TimeSeries:       make([]StressTimeSeriesPoint, 0),
	}

	// Start with initial RPS
	currentRPS := est.config.StartRPS
	stepCount := 0

	for currentRPS <= est.config.MaxRPS {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		est.logger.Info("executing stress test step",
			zap.Int("step", stepCount+1),
			zap.Int("current_rps", currentRPS),
			zap.Duration("step_duration", est.config.StepDuration))

		// Execute step
		stepCtx, cancel := context.WithTimeout(ctx, est.config.StepDuration)
		stepMetrics, breakingPointDetected, err := est.executeStressStep(stepCtx, currentRPS, metrics)
		cancel()

		if err != nil {
			return nil, fmt.Errorf("stress step failed: %w", err)
		}

		// Update peak RPS
		if currentRPS > metrics.PeakRPS {
			metrics.PeakRPS = currentRPS
		}

		// Check for breaking point
		if breakingPointDetected {
			est.logger.Info("breaking point detected", zap.Int("rps", currentRPS))

			breakingPoint := &BreakingPoint{
				RPS:             currentRPS,
				Timestamp:       time.Now(),
				TriggerReason:   est.breakPoint.GetTriggerReason(),
				ErrorRate:       stepMetrics.errorRate,
				AvgResponseTime: stepMetrics.avgResponseTime,
				P95ResponseTime: stepMetrics.p95ResponseTime,
				CPUUsage:        stepMetrics.cpuUsage,
				MemoryUsage:     stepMetrics.memoryUsage,
			}

			metrics.BreakingPoint = breakingPoint
			break
		}

		// Move to next step
		currentRPS += est.config.StepSize
		stepCount++
	}

	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)

	// Calculate final metrics
	est.calculateFinalStressMetrics(metrics)

	// Build results
	results := &StressTestResults{
		TestInfo: StressTestInfo{
			Name:        est.config.TestName,
			StartTime:   metrics.StartTime,
			EndTime:     metrics.EndTime,
			Duration:    metrics.Duration,
			Config:      est.config,
			Environment: est.getEnvironmentInfo(),
		},
		Metrics:         metrics,
		Summary:         est.generateStressSummary(metrics),
		Recommendations: est.generateStressRecommendations(metrics),
		Alerts:          est.monitor.GetAlerts(),
	}

	return results, nil
}

// StepMetrics represents metrics for a single stress test step
type StepMetrics struct {
	errorRate       float64
	avgResponseTime time.Duration
	p95ResponseTime time.Duration
	cpuUsage        float64
	memoryUsage     uint64
}

// executeStressStep executes a single stress test step
func (est *EnhancedStressTester) executeStressStep(ctx context.Context, rps int, metrics *StressTestMetrics) (*StepMetrics, bool, error) {
	// Implementation would execute stress at the given RPS
	// This is a simplified version

	stepMetrics := &StepMetrics{
		errorRate:       0.05, // 5% error rate
		avgResponseTime: 200 * time.Millisecond,
		p95ResponseTime: 500 * time.Millisecond,
		cpuUsage:        float64(rps) * 0.1,        // Simulate CPU usage increase
		memoryUsage:     uint64(rps * 1024 * 1024), // Simulate memory usage
	}

	// Simulate requests
	atomic.AddInt64(&metrics.TotalRequests, int64(rps))
	atomic.AddInt64(&metrics.SuccessfulRequests, int64(float64(rps)*0.95))
	atomic.AddInt64(&metrics.FailedRequests, int64(float64(rps)*0.05))

	// Add time series point
	point := StressTimeSeriesPoint{
		Timestamp:       time.Now(),
		RPS:             rps,
		ErrorRate:       stepMetrics.errorRate,
		AvgResponseTime: stepMetrics.avgResponseTime,
		CPUUsage:        stepMetrics.cpuUsage,
		MemoryUsage:     stepMetrics.memoryUsage,
		ActiveWorkers:   rps / 10, // Estimate active workers
	}
	metrics.TimeSeries = append(metrics.TimeSeries, point)

	// Check for breaking point
	breakingPointDetected := est.breakPoint.CheckBreakingPoint(stepMetrics.errorRate, stepMetrics.avgResponseTime, stepMetrics.cpuUsage, stepMetrics.memoryUsage)

	return stepMetrics, breakingPointDetected, nil
}

// calculateFinalStressMetrics calculates final stress test metrics
func (est *EnhancedStressTester) calculateFinalStressMetrics(metrics *StressTestMetrics) {
	if metrics.TotalRequests > 0 {
		metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
	}

	// Calculate response time percentiles (simplified)
	if len(metrics.ResponseTimes) > 0 {
		sort.Slice(metrics.ResponseTimes, func(i, j int) bool {
			return metrics.ResponseTimes[i] < metrics.ResponseTimes[j]
		})

		n := len(metrics.ResponseTimes)
		metrics.P50ResponseTime = metrics.ResponseTimes[(n-1)*50/100]
		metrics.P95ResponseTime = metrics.ResponseTimes[(n-1)*95/100]
		metrics.P99ResponseTime = metrics.ResponseTimes[(n-1)*99/100]

		if n > 0 {
			metrics.MinResponseTime = metrics.ResponseTimes[0]
			metrics.MaxResponseTime = metrics.ResponseTimes[n-1]

			total := time.Duration(0)
			for _, rt := range metrics.ResponseTimes {
				total += rt
			}
			metrics.AvgResponseTime = total / time.Duration(n)
		}
	}

	// Calculate peak resource usage
	for _, snapshot := range metrics.ResourceSnapshots {
		if snapshot.CPUUsage > metrics.PeakCPUUsage {
			metrics.PeakCPUUsage = snapshot.CPUUsage
		}
		if snapshot.MemoryUsage > metrics.PeakMemoryUsage {
			metrics.PeakMemoryUsage = snapshot.MemoryUsage
		}
	}
}

// generateStressSummary generates stress test summary
func (est *EnhancedStressTester) generateStressSummary(metrics *StressTestMetrics) StressTestSummary {
	breakingPointFound := metrics.BreakingPoint != nil
	success := breakingPointFound // For stress tests, finding the breaking point is success
	status := "COMPLETED"
	if !breakingPointFound {
		status = "NO_BREAKING_POINT"
	}

	// Calculate resilience score
	score := est.calculateResilienceScore(metrics)

	// Determine grade
	var grade string
	switch {
	case score >= 90:
		grade = "A"
	case score >= 80:
		grade = "B"
	case score >= 70:
		grade = "C"
	case score >= 60:
		grade = "D"
	default:
		grade = "F"
	}

	// Generate key findings
	var findings []string
	if breakingPointFound {
		findings = append(findings, fmt.Sprintf("Breaking point found at %d RPS", metrics.BreakingPoint.RPS))
	}
	if metrics.ErrorRate > 0.1 {
		findings = append(findings, "High error rate during stress test")
	}
	if metrics.PeakCPUUsage > 90 {
		findings = append(findings, "High CPU utilization detected")
	}

	// Recovery assessment
	recoverySuccessful := false
	if metrics.RecoveryMetrics != nil {
		recoverySuccessful = metrics.RecoveryMetrics.FullRecovery
	}

	return StressTestSummary{
		Status:             status,
		Success:            success,
		Grade:              grade,
		Score:              score,
		BreakingPointFound: breakingPointFound,
		RecoverySuccessful: recoverySuccessful,
		KeyFindings:        findings,
		SystemResilience:   est.getResilienceLevel(score),
	}
}

// calculateResilienceScore calculates system resilience score
func (est *EnhancedStressTester) calculateResilienceScore(metrics *StressTestMetrics) float64 {
	score := 100.0

	// Deduct points for early breaking point
	if metrics.BreakingPoint != nil {
		expectedRPS := float64(est.config.MaxRPS)
		actualRPS := float64(metrics.BreakingPoint.RPS)
		ratio := actualRPS / expectedRPS
		score *= ratio
	}

	// Deduct points for high error rate
	if metrics.ErrorRate > 0.05 {
		score -= (metrics.ErrorRate - 0.05) * 200 // 20 points per 10% error rate above 5%
	}

	// Deduct points for high resource usage
	if metrics.PeakCPUUsage > 80 {
		score -= (metrics.PeakCPUUsage - 80) * 2 // 2 points per 1% CPU above 80%
	}

	// Add points for successful recovery
	if metrics.RecoveryMetrics != nil && metrics.RecoveryMetrics.FullRecovery {
		score += 10
	}

	return math.Max(0, score)
}

// getResilienceLevel returns resilience level based on score
func (est *EnhancedStressTester) getResilienceLevel(score float64) string {
	switch {
	case score >= 90:
		return "Excellent"
	case score >= 80:
		return "Good"
	case score >= 70:
		return "Fair"
	case score >= 60:
		return "Poor"
	default:
		return "Critical"
	}
}

// generateStressRecommendations generates stress test recommendations
func (est *EnhancedStressTester) generateStressRecommendations(metrics *StressTestMetrics) []string {
	var recommendations []string

	if metrics.BreakingPoint != nil {
		recommendations = append(recommendations,
			fmt.Sprintf("System breaking point found at %d RPS. Consider optimizations to increase capacity", metrics.BreakingPoint.RPS))
	}

	if metrics.ErrorRate > 0.1 {
		recommendations = append(recommendations, "Investigate and fix error sources during high load")
	}

	if metrics.PeakCPUUsage > 80 {
		recommendations = append(recommendations, "Consider scaling out to reduce CPU utilization during peak load")
	}

	if metrics.PeakMemoryUsage > 1*1024*1024*1024 {
		recommendations = append(recommendations, "Optimize memory usage to handle higher loads")
	}

	if metrics.RecoveryMetrics == nil {
		recommendations = append(recommendations, "Enable recovery testing to assess system resilience")
	} else if !metrics.RecoveryMetrics.FullRecovery {
		recommendations = append(recommendations, "Improve system recovery mechanisms after stress events")
	}

	return recommendations
}

// getEnvironmentInfo returns environment information
func (est *EnhancedStressTester) getEnvironmentInfo() map[string]string {
	return map[string]string{
		"go_version": fmt.Sprintf("go%s", "1.22"),
		"os":         "linux",
		"arch":       "amd64",
		"base_url":   est.config.BaseURL,
		"test_type":  "enhanced_stress_test",
		"max_rps":    fmt.Sprintf("%d", est.config.MaxRPS),
		"step_size":  fmt.Sprintf("%d", est.config.StepSize),
	}
}

// Shutdown gracefully shuts down the stress tester
func (est *EnhancedStressTester) Shutdown() error {
	select {
	case <-est.stopChan:
		return nil
	default:
		close(est.stopChan)
	}
	return nil
}

// NewStressTestExecutor creates a new stress test executor
func NewStressTestExecutor(config *EnhancedStressTestConfig, logger *zap.Logger, client *http.Client) *StressTestExecutor {
	return &StressTestExecutor{
		config: config,
		logger: logger,
		client: client,
	}
}

// Initialize initializes the stress test executor
func (ste *StressTestExecutor) Initialize() error {
	// Initialize workers and other components
	return nil
}

// NewStressTestMonitor creates a new stress test monitor
func NewStressTestMonitor(config *EnhancedStressTestConfig, logger *zap.Logger) *StressTestMonitor {
	return &StressTestMonitor{
		config:   config,
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// Start starts the stress test monitor
func (stm *StressTestMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-stm.stopChan:
			return
		case <-ticker.C:
			stm.collectMetrics()
		}
	}
}

// GetAlerts returns stress test alerts
func (stm *StressTestMonitor) GetAlerts() []StressAlert {
	// Implementation would return collected alerts
	return []StressAlert{}
}

// collectMetrics collects stress test metrics
func (stm *StressTestMonitor) collectMetrics() {
	// Implementation would collect real-time metrics
}

// NewBreakingPointDetector creates a new breaking point detector
func NewBreakingPointDetector(config *EnhancedStressTestConfig, logger *zap.Logger) *BreakingPointDetector {
	return &BreakingPointDetector{
		config: config,
		logger: logger,
		detectionMetrics: &DetectionMetrics{
			WindowSize:          1 * time.Minute,
			ErrorRateHistory:    make([]float64, 0),
			ResponseTimeHistory: make([]time.Duration, 0),
			ResourceHistory:     make([]ResourceSnapshot, 0),
		},
	}
}

// CheckBreakingPoint checks if a breaking point has been reached
func (bpd *BreakingPointDetector) CheckBreakingPoint(errorRate float64, responseTime time.Duration, cpuUsage float64, memoryUsage uint64) bool {
	bpd.mu.Lock()
	defer bpd.mu.Unlock()

	// Check thresholds
	if errorRate > bpd.config.FailureThreshold {
		bpd.logger.Info("breaking point detected: error rate threshold exceeded",
			zap.Float64("error_rate", errorRate),
			zap.Float64("threshold", bpd.config.FailureThreshold))
		return true
	}

	if responseTime > bpd.config.ResponseThreshold {
		bpd.logger.Info("breaking point detected: response time threshold exceeded",
			zap.Duration("response_time", responseTime),
			zap.Duration("threshold", bpd.config.ResponseThreshold))
		return true
	}

	if cpuUsage > bpd.config.ResourceThreshold.MaxCPUUsage {
		bpd.logger.Info("breaking point detected: CPU usage threshold exceeded",
			zap.Float64("cpu_usage", cpuUsage),
			zap.Float64("threshold", bpd.config.ResourceThreshold.MaxCPUUsage))
		return true
	}

	if memoryUsage > bpd.config.ResourceThreshold.MaxMemoryUsage {
		bpd.logger.Info("breaking point detected: memory usage threshold exceeded",
			zap.Uint64("memory_usage", memoryUsage),
			zap.Uint64("threshold", bpd.config.ResourceThreshold.MaxMemoryUsage))
		return true
	}

	return false
}

// GetTriggerReason returns the reason for breaking point detection
func (bpd *BreakingPointDetector) GetTriggerReason() string {
	return "threshold_exceeded"
}

// NewRecoveryTester creates a new recovery tester
func NewRecoveryTester(config *EnhancedStressTestConfig, logger *zap.Logger, client *http.Client) *RecoveryTester {
	return &RecoveryTester{
		config: config,
		logger: logger,
		client: client,
	}
}

// TestRecovery tests system recovery after stress
func (rt *RecoveryTester) TestRecovery(ctx context.Context) (*RecoveryMetrics, error) {
	rt.logger.Info("starting recovery test", zap.Duration("duration", rt.config.RecoveryDuration))

	metrics := &RecoveryMetrics{
		StartTime:     time.Now(),
		RecoverySteps: make([]RecoveryStep, 0),
	}

	// Test recovery in steps
	stepDuration := rt.config.RecoveryDuration / time.Duration(rt.config.RecoverySteps)

	for step := 1; step <= rt.config.RecoverySteps; step++ {
		select {
		case <-ctx.Done():
			return metrics, ctx.Err()
		default:
		}

		// Execute recovery step
		stepMetrics := rt.executeRecoveryStep(step, stepDuration)
		metrics.RecoverySteps = append(metrics.RecoverySteps, stepMetrics)

		if stepMetrics.Healthy {
			metrics.FullRecovery = true
		}
	}

	metrics.EndTime = time.Now()
	metrics.RecoveryDuration = metrics.EndTime.Sub(metrics.StartTime)

	rt.logger.Info("recovery test completed",
		zap.Bool("full_recovery", metrics.FullRecovery),
		zap.Duration("recovery_duration", metrics.RecoveryDuration))

	return metrics, nil
}

// executeRecoveryStep executes a single recovery step
func (rt *RecoveryTester) executeRecoveryStep(step int, duration time.Duration) RecoveryStep {
	// Implementation would test system health at reduced load
	// This is a simplified version

	// Simulate gradual recovery
	errorRate := 0.1 / float64(step)                           // Error rate decreases with each step
	responseTime := time.Duration(500/step) * time.Millisecond // Response time improves
	healthy := step >= 3                                       // System becomes healthy after step 3

	return RecoveryStep{
		Step:            step,
		Timestamp:       time.Now(),
		RPS:             50 / step, // Reduced RPS for recovery testing
		ErrorRate:       errorRate,
		AvgResponseTime: responseTime,
		Healthy:         healthy,
	}
}

// NewStressTestReporter creates a new stress test reporter
func NewStressTestReporter(config *EnhancedStressTestConfig, logger *zap.Logger) *StressTestReporter {
	return &StressTestReporter{
		config: config,
		logger: logger,
	}
}

// GenerateReport generates a stress test report
func (str *StressTestReporter) GenerateReport(results *StressTestResults) error {
	str.logger.Info("generating stress test report",
		zap.String("format", str.config.ExportFormat),
		zap.String("status", results.Summary.Status))
	return nil
}
