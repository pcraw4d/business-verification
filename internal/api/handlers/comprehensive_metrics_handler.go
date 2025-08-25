package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/health"
	"go.uber.org/zap"
)

// ComprehensiveMetricsData represents comprehensive system metrics
type ComprehensiveMetricsData struct {
	Timestamp          time.Time              `json:"timestamp"`
	Version            string                 `json:"version"`
	Environment        string                 `json:"environment"`
	SystemMetrics      SystemMetrics          `json:"system_metrics"`
	APIMetrics         APIMetrics             `json:"api_metrics"`
	BusinessMetrics    BusinessMetrics        `json:"business_metrics"`
	PerformanceMetrics PerformanceMetrics     `json:"performance_metrics"`
	ResourceMetrics    ResourceMetrics        `json:"resource_metrics"`
	ErrorMetrics       ErrorMetrics           `json:"error_metrics"`
	CustomMetrics      map[string]interface{} `json:"custom_metrics,omitempty"`
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	Uptime        time.Duration `json:"uptime"`
	StartTime     time.Time     `json:"start_time"`
	ProcessID     int           `json:"process_id"`
	GoVersion     string        `json:"go_version"`
	NumCPU        int           `json:"num_cpu"`
	NumGoroutines int           `json:"num_goroutines"`
	NumCgoCall    int64         `json:"num_cgo_call"`
	MemoryStats   MemoryStats   `json:"memory_stats"`
	GCStats       GCStats       `json:"gc_stats"`
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Alloc         uint64      `json:"alloc_bytes"`
	TotalAlloc    uint64      `json:"total_alloc_bytes"`
	Sys           uint64      `json:"sys_bytes"`
	HeapAlloc     uint64      `json:"heap_alloc_bytes"`
	HeapSys       uint64      `json:"heap_sys_bytes"`
	HeapIdle      uint64      `json:"heap_idle_bytes"`
	HeapInuse     uint64      `json:"heap_inuse_bytes"`
	HeapReleased  uint64      `json:"heap_released_bytes"`
	HeapObjects   uint64      `json:"heap_objects"`
	StackInuse    uint64      `json:"stack_inuse_bytes"`
	StackSys      uint64      `json:"stack_sys_bytes"`
	MSpanInuse    uint64      `json:"mspan_inuse_bytes"`
	MSpanSys      uint64      `json:"mspan_sys_bytes"`
	MCacheInuse   uint64      `json:"mcache_inuse_bytes"`
	MCacheSys     uint64      `json:"mcache_sys_bytes"`
	BuckHashSys   uint64      `json:"buck_hash_sys_bytes"`
	GCSys         uint64      `json:"gc_sys_bytes"`
	OtherSys      uint64      `json:"other_sys_bytes"`
	NextGC        uint64      `json:"next_gc_bytes"`
	LastGC        uint64      `json:"last_gc_nanoseconds"`
	PauseTotalNs  uint64      `json:"pause_total_nanoseconds"`
	PauseNs       [256]uint64 `json:"pause_nanoseconds"`
	PauseEnd      [256]uint64 `json:"pause_end_nanoseconds"`
	NumGC         uint32      `json:"num_gc"`
	NumForcedGC   uint32      `json:"num_forced_gc"`
	GCCPUFraction float64     `json:"gc_cpu_fraction"`
	EnableGC      bool        `json:"enable_gc"`
	DebugGC       bool        `json:"debug_gc"`
}

// GCStats represents garbage collection statistics
type GCStats struct {
	NumGC         uint32      `json:"num_gc"`
	PauseTotalNs  uint64      `json:"pause_total_nanoseconds"`
	PauseNs       [256]uint64 `json:"pause_nanoseconds"`
	PauseEnd      [256]uint64 `json:"pause_end_nanoseconds"`
	GCCPUFraction float64     `json:"gc_cpu_fraction"`
	LastGC        uint64      `json:"last_gc_nanoseconds"`
	NextGC        uint64      `json:"next_gc_bytes"`
}

// APIMetrics represents API-level metrics
type APIMetrics struct {
	TotalRequests          int64            `json:"total_requests"`
	RequestsPerSecond      float64          `json:"requests_per_second"`
	AverageResponseTime    time.Duration    `json:"average_response_time"`
	ResponseTimeP95        time.Duration    `json:"response_time_p95"`
	ResponseTimeP99        time.Duration    `json:"response_time_p99"`
	ActiveRequests         int              `json:"active_requests"`
	RequestsByMethod       map[string]int64 `json:"requests_by_method"`
	RequestsByEndpoint     map[string]int64 `json:"requests_by_endpoint"`
	RequestsByStatus       map[int]int64    `json:"requests_by_status"`
	ErrorsByType           map[string]int64 `json:"errors_by_type"`
	RateLimitHits          int64            `json:"rate_limit_hits"`
	AuthenticationFailures int64            `json:"authentication_failures"`
	AuthorizationFailures  int64            `json:"authorization_failures"`
}

// BusinessMetrics represents business-specific metrics
type BusinessMetrics struct {
	TotalVerifications      int64            `json:"total_verifications"`
	VerificationsPerSecond  float64          `json:"verifications_per_second"`
	AverageVerificationTime time.Duration    `json:"average_verification_time"`
	VerificationsByStatus   map[string]int64 `json:"verifications_by_status"`
	VerificationsByType     map[string]int64 `json:"verifications_by_type"`
	VerificationsByIndustry map[string]int64 `json:"verifications_by_industry"`
	SuccessRate             float64          `json:"success_rate"`
	ErrorRate               float64          `json:"error_rate"`
	AverageConfidenceScore  float64          `json:"average_confidence_score"`
	CacheHitRate            float64          `json:"cache_hit_rate"`
	CacheMissRate           float64          `json:"cache_miss_rate"`
	ExternalAPICalls        int64            `json:"external_api_calls"`
	ExternalAPILatency      time.Duration    `json:"external_api_latency"`
	MLModelPredictions      int64            `json:"ml_model_predictions"`
	MLModelAccuracy         float64          `json:"ml_model_accuracy"`
}

// PerformanceMetrics represents performance-related metrics
type PerformanceMetrics struct {
	CPUUsagePercent     float64       `json:"cpu_usage_percent"`
	MemoryUsagePercent  float64       `json:"memory_usage_percent"`
	DiskUsagePercent    float64       `json:"disk_usage_percent"`
	NetworkLatency      time.Duration `json:"network_latency"`
	NetworkThroughput   float64       `json:"network_throughput_mbps"`
	DatabaseConnections int           `json:"database_connections"`
	DatabaseLatency     time.Duration `json:"database_latency"`
	CacheConnections    int           `json:"cache_connections"`
	CacheLatency        time.Duration `json:"cache_latency"`
	QueueDepth          int           `json:"queue_depth"`
	WorkerUtilization   float64       `json:"worker_utilization_percent"`
}

// ResourceMetrics represents resource utilization metrics
type ResourceMetrics struct {
	OpenFiles              int        `json:"open_files"`
	MaxFiles               int        `json:"max_files"`
	FileDescriptors        int        `json:"file_descriptors"`
	Threads                int        `json:"threads"`
	MaxThreads             int        `json:"max_threads"`
	LoadAverage            [3]float64 `json:"load_average"`
	DiskIOReadBytes        int64      `json:"disk_io_read_bytes"`
	DiskIOWriteBytes       int64      `json:"disk_io_write_bytes"`
	DiskIOReadOps          int64      `json:"disk_io_read_ops"`
	DiskIOWriteOps         int64      `json:"disk_io_write_ops"`
	NetworkBytesReceived   int64      `json:"network_bytes_received"`
	NetworkBytesSent       int64      `json:"network_bytes_sent"`
	NetworkPacketsReceived int64      `json:"network_packets_received"`
	NetworkPacketsSent     int64      `json:"network_packets_sent"`
}

// ErrorMetrics represents error-related metrics
type ErrorMetrics struct {
	TotalErrors      int64            `json:"total_errors"`
	ErrorsPerSecond  float64          `json:"errors_per_second"`
	ErrorsByType     map[string]int64 `json:"errors_by_type"`
	ErrorsByEndpoint map[string]int64 `json:"errors_by_endpoint"`
	ErrorsByStatus   map[int]int64    `json:"errors_by_status"`
	LastErrorTime    time.Time        `json:"last_error_time"`
	LastErrorMessage string           `json:"last_error_message"`
	ErrorRate        float64          `json:"error_rate"`
	CriticalErrors   int64            `json:"critical_errors"`
	WarningErrors    int64            `json:"warning_errors"`
	InfoErrors       int64            `json:"info_errors"`
}

// ComprehensiveMetricsHandler provides comprehensive metrics collection endpoints
type ComprehensiveMetricsHandler struct {
	logger         *zap.Logger
	healthChecker  *health.RailwayHealthChecker
	startTime      time.Time
	version        string
	environment    string
	mu             sync.RWMutex
	metrics        *ComprehensiveMetricsData
	lastUpdate     time.Time
	updateInterval time.Duration
	collectors     map[string]MetricsCollector
}

// MetricsCollector interface for custom metrics collection
type MetricsCollector interface {
	Collect(ctx context.Context) (map[string]interface{}, error)
	Name() string
}

// NewComprehensiveMetricsHandler creates a new comprehensive metrics handler
func NewComprehensiveMetricsHandler(logger *zap.Logger, healthChecker *health.RailwayHealthChecker, version, environment string) *ComprehensiveMetricsHandler {
	handler := &ComprehensiveMetricsHandler{
		logger:         logger,
		healthChecker:  healthChecker,
		startTime:      time.Now(),
		version:        version,
		environment:    environment,
		updateInterval: 30 * time.Second,
		collectors:     make(map[string]MetricsCollector),
	}

	// Initialize metrics
	handler.metrics = handler.collectMetrics(context.Background())

	// Start background metrics collection
	go handler.startMetricsCollection()

	return handler
}

// RegisterCollector registers a custom metrics collector
func (h *ComprehensiveMetricsHandler) RegisterCollector(collector MetricsCollector) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.collectors[collector.Name()] = collector
}

// HandleComprehensiveMetrics handles the main comprehensive metrics endpoint
func (h *ComprehensiveMetricsHandler) HandleComprehensiveMetrics(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Check if we can serve from cache
	h.mu.RLock()
	if time.Since(h.lastUpdate) < h.updateInterval {
		cachedMetrics := h.metrics
		h.mu.RUnlock()

		h.serveMetricsResponse(w, r, cachedMetrics, time.Since(start))
		return
	}
	h.mu.RUnlock()

	// Collect fresh metrics
	ctx := r.Context()
	metrics := h.collectMetrics(ctx)

	// Update cache
	h.mu.Lock()
	h.metrics = metrics
	h.lastUpdate = time.Now()
	h.mu.Unlock()

	h.serveMetricsResponse(w, r, metrics, time.Since(start))
}

// HandlePrometheusMetrics handles Prometheus-formatted metrics
func (h *ComprehensiveMetricsHandler) HandlePrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Collect metrics
	ctx := r.Context()
	metrics := h.collectMetrics(ctx)

	// Convert to Prometheus format
	prometheusMetrics := h.convertToPrometheusFormat(metrics)

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(prometheusMetrics))

	h.logger.Debug("Prometheus metrics served",
		zap.String("client_ip", r.RemoteAddr),
		zap.Duration("response_time", time.Since(start)),
	)
}

// HandleSystemMetrics handles system-specific metrics
func (h *ComprehensiveMetricsHandler) HandleSystemMetrics(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Collect system metrics only
	systemMetrics := h.collectSystemMetrics()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"timestamp":      time.Now(),
		"version":        h.version,
		"environment":    h.environment,
		"system_metrics": systemMetrics,
	}

	json.NewEncoder(w).Encode(response)

	h.logger.Debug("System metrics served",
		zap.String("client_ip", r.RemoteAddr),
		zap.Duration("response_time", time.Since(start)),
	)
}

// HandleAPIMetrics handles API-specific metrics
func (h *ComprehensiveMetricsHandler) HandleAPIMetrics(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Collect API metrics only
	apiMetrics := h.collectAPIMetrics()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"timestamp":   time.Now(),
		"version":     h.version,
		"environment": h.environment,
		"api_metrics": apiMetrics,
	}

	json.NewEncoder(w).Encode(response)

	h.logger.Debug("API metrics served",
		zap.String("client_ip", r.RemoteAddr),
		zap.Duration("response_time", time.Since(start)),
	)
}

// HandleBusinessMetrics handles business-specific metrics
func (h *ComprehensiveMetricsHandler) HandleBusinessMetrics(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Collect business metrics only
	businessMetrics := h.collectBusinessMetrics()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"timestamp":        time.Now(),
		"version":          h.version,
		"environment":      h.environment,
		"business_metrics": businessMetrics,
	}

	json.NewEncoder(w).Encode(response)

	h.logger.Debug("Business metrics served",
		zap.String("client_ip", r.RemoteAddr),
		zap.Duration("response_time", time.Since(start)),
	)
}

// collectMetrics collects comprehensive metrics
func (h *ComprehensiveMetricsHandler) collectMetrics(ctx context.Context) *ComprehensiveMetricsData {
	metrics := &ComprehensiveMetricsData{
		Timestamp:          time.Now(),
		Version:            h.version,
		Environment:        h.environment,
		SystemMetrics:      h.collectSystemMetrics(),
		APIMetrics:         h.collectAPIMetrics(),
		BusinessMetrics:    h.collectBusinessMetrics(),
		PerformanceMetrics: h.collectPerformanceMetrics(),
		ResourceMetrics:    h.collectResourceMetrics(),
		ErrorMetrics:       h.collectErrorMetrics(),
		CustomMetrics:      make(map[string]interface{}),
	}

	// Collect custom metrics from registered collectors
	h.mu.RLock()
	for _, collector := range h.collectors {
		if customMetrics, err := collector.Collect(ctx); err == nil {
			metrics.CustomMetrics[collector.Name()] = customMetrics
		} else {
			h.logger.Warn("Failed to collect custom metrics",
				zap.String("collector", collector.Name()),
				zap.Error(err),
			)
		}
	}
	h.mu.RUnlock()

	return metrics
}

// collectSystemMetrics collects system-level metrics
func (h *ComprehensiveMetricsHandler) collectSystemMetrics() SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemMetrics{
		Uptime:        time.Since(h.startTime),
		StartTime:     h.startTime,
		ProcessID:     runtime.NumCPU(), // Placeholder for actual PID
		GoVersion:     runtime.Version(),
		NumCPU:        runtime.NumCPU(),
		NumGoroutines: runtime.NumGoroutine(),
		NumCgoCall:    runtime.NumCgoCall(),
		MemoryStats: MemoryStats{
			Alloc:         m.Alloc,
			TotalAlloc:    m.TotalAlloc,
			Sys:           m.Sys,
			HeapAlloc:     m.HeapAlloc,
			HeapSys:       m.HeapSys,
			HeapIdle:      m.HeapIdle,
			HeapInuse:     m.HeapInuse,
			HeapReleased:  m.HeapReleased,
			HeapObjects:   m.HeapObjects,
			StackInuse:    m.StackInuse,
			StackSys:      m.StackSys,
			MSpanInuse:    m.MSpanInuse,
			MSpanSys:      m.MSpanSys,
			MCacheInuse:   m.MCacheInuse,
			MCacheSys:     m.MCacheSys,
			BuckHashSys:   m.BuckHashSys,
			GCSys:         m.GCSys,
			OtherSys:      m.OtherSys,
			NextGC:        m.NextGC,
			LastGC:        m.LastGC,
			PauseTotalNs:  m.PauseTotalNs,
			PauseNs:       m.PauseNs,
			PauseEnd:      m.PauseEnd,
			NumGC:         m.NumGC,
			NumForcedGC:   m.NumForcedGC,
			GCCPUFraction: m.GCCPUFraction,
			EnableGC:      m.EnableGC,
			DebugGC:       m.DebugGC,
		},
		GCStats: GCStats{
			NumGC:         m.NumGC,
			PauseTotalNs:  m.PauseTotalNs,
			PauseNs:       m.PauseNs,
			PauseEnd:      m.PauseEnd,
			GCCPUFraction: m.GCCPUFraction,
			LastGC:        m.LastGC,
			NextGC:        m.NextGC,
		},
	}
}

// collectAPIMetrics collects API-level metrics
func (h *ComprehensiveMetricsHandler) collectAPIMetrics() APIMetrics {
	// TODO: Implement actual API metrics collection
	// This is a placeholder implementation
	return APIMetrics{
		TotalRequests:       1000,
		RequestsPerSecond:   50.5,
		AverageResponseTime: 150 * time.Millisecond,
		ResponseTimeP95:     300 * time.Millisecond,
		ResponseTimeP99:     500 * time.Millisecond,
		ActiveRequests:      5,
		RequestsByMethod: map[string]int64{
			"GET":    600,
			"POST":   300,
			"PUT":    50,
			"DELETE": 50,
		},
		RequestsByEndpoint: map[string]int64{
			"/health":   200,
			"/metrics":  100,
			"/verify":   400,
			"/classify": 200,
			"/risk":     100,
		},
		RequestsByStatus: map[int]int64{
			200: 950,
			400: 30,
			401: 10,
			500: 10,
		},
		ErrorsByType: map[string]int64{
			"validation":     25,
			"authentication": 10,
			"authorization":  5,
			"internal":       10,
		},
		RateLimitHits:          5,
		AuthenticationFailures: 10,
		AuthorizationFailures:  5,
	}
}

// collectBusinessMetrics collects business-specific metrics
func (h *ComprehensiveMetricsHandler) collectBusinessMetrics() BusinessMetrics {
	// TODO: Implement actual business metrics collection
	// This is a placeholder implementation
	return BusinessMetrics{
		TotalVerifications:      5000,
		VerificationsPerSecond:  25.3,
		AverageVerificationTime: 2 * time.Second,
		VerificationsByStatus: map[string]int64{
			"pending":   100,
			"completed": 4800,
			"failed":    100,
		},
		VerificationsByType: map[string]int64{
			"basic":    3000,
			"advanced": 1500,
			"premium":  500,
		},
		VerificationsByIndustry: map[string]int64{
			"technology":    1500,
			"finance":       1200,
			"healthcare":    800,
			"retail":        1000,
			"manufacturing": 500,
		},
		SuccessRate:            96.0,
		ErrorRate:              4.0,
		AverageConfidenceScore: 0.85,
		CacheHitRate:           75.0,
		CacheMissRate:          25.0,
		ExternalAPICalls:       15000,
		ExternalAPILatency:     500 * time.Millisecond,
		MLModelPredictions:     8000,
		MLModelAccuracy:        92.5,
	}
}

// collectPerformanceMetrics collects performance-related metrics
func (h *ComprehensiveMetricsHandler) collectPerformanceMetrics() PerformanceMetrics {
	// TODO: Implement actual performance metrics collection
	// This is a placeholder implementation
	return PerformanceMetrics{
		CPUUsagePercent:     45.2,
		MemoryUsagePercent:  62.8,
		DiskUsagePercent:    35.5,
		NetworkLatency:      25 * time.Millisecond,
		NetworkThroughput:   850.5,
		DatabaseConnections: 15,
		DatabaseLatency:     50 * time.Millisecond,
		CacheConnections:    8,
		CacheLatency:        5 * time.Millisecond,
		QueueDepth:          12,
		WorkerUtilization:   78.5,
	}
}

// collectResourceMetrics collects resource utilization metrics
func (h *ComprehensiveMetricsHandler) collectResourceMetrics() ResourceMetrics {
	// TODO: Implement actual resource metrics collection
	// This is a placeholder implementation
	return ResourceMetrics{
		OpenFiles:              125,
		MaxFiles:               1024,
		FileDescriptors:        125,
		Threads:                45,
		MaxThreads:             1000,
		LoadAverage:            [3]float64{1.2, 1.5, 1.8},
		DiskIOReadBytes:        1024 * 1024 * 100, // 100MB
		DiskIOWriteBytes:       1024 * 1024 * 50,  // 50MB
		DiskIOReadOps:          1000,
		DiskIOWriteOps:         500,
		NetworkBytesReceived:   1024 * 1024 * 200, // 200MB
		NetworkBytesSent:       1024 * 1024 * 150, // 150MB
		NetworkPacketsReceived: 50000,
		NetworkPacketsSent:     40000,
	}
}

// collectErrorMetrics collects error-related metrics
func (h *ComprehensiveMetricsHandler) collectErrorMetrics() ErrorMetrics {
	// TODO: Implement actual error metrics collection
	// This is a placeholder implementation
	return ErrorMetrics{
		TotalErrors:     50,
		ErrorsPerSecond: 2.5,
		ErrorsByType: map[string]int64{
			"validation":     20,
			"authentication": 10,
			"authorization":  5,
			"internal":       10,
			"external":       5,
		},
		ErrorsByEndpoint: map[string]int64{
			"/verify":   25,
			"/classify": 15,
			"/risk":     10,
		},
		ErrorsByStatus: map[int]int64{
			400: 30,
			401: 10,
			403: 5,
			500: 5,
		},
		LastErrorTime:    time.Now().Add(-5 * time.Minute),
		LastErrorMessage: "Database connection timeout",
		ErrorRate:        5.0,
		CriticalErrors:   5,
		WarningErrors:    20,
		InfoErrors:       25,
	}
}

// convertToPrometheusFormat converts metrics to Prometheus format
func (h *ComprehensiveMetricsHandler) convertToPrometheusFormat(metrics *ComprehensiveMetricsData) string {
	var prometheusMetrics string

	// System metrics
	prometheusMetrics += fmt.Sprintf("# HELP go_goroutines Number of goroutines that currently exist.\n")
	prometheusMetrics += fmt.Sprintf("# TYPE go_goroutines gauge\n")
	prometheusMetrics += fmt.Sprintf("go_goroutines{version=\"%s\",environment=\"%s\"} %d\n",
		metrics.Version, metrics.Environment, metrics.SystemMetrics.NumGoroutines)

	prometheusMetrics += fmt.Sprintf("# HELP go_threads Number of OS threads created.\n")
	prometheusMetrics += fmt.Sprintf("# TYPE go_threads gauge\n")
	prometheusMetrics += fmt.Sprintf("go_threads{version=\"%s\",environment=\"%s\"} %d\n",
		metrics.Version, metrics.Environment, metrics.SystemMetrics.NumCPU)

	prometheusMetrics += fmt.Sprintf("# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.\n")
	prometheusMetrics += fmt.Sprintf("# TYPE go_memstats_alloc_bytes gauge\n")
	prometheusMetrics += fmt.Sprintf("go_memstats_alloc_bytes{version=\"%s\",environment=\"%s\"} %d\n",
		metrics.Version, metrics.Environment, metrics.SystemMetrics.MemoryStats.Alloc)

	prometheusMetrics += fmt.Sprintf("# HELP go_memstats_sys_bytes Number of bytes obtained from system.\n")
	prometheusMetrics += fmt.Sprintf("# TYPE go_memstats_sys_bytes gauge\n")
	prometheusMetrics += fmt.Sprintf("go_memstats_sys_bytes{version=\"%s\",environment=\"%s\"} %d\n",
		metrics.Version, metrics.Environment, metrics.SystemMetrics.MemoryStats.Sys)

	// API metrics
	prometheusMetrics += fmt.Sprintf("# HELP api_requests_total Total number of API requests.\n")
	prometheusMetrics += fmt.Sprintf("# TYPE api_requests_total counter\n")
	prometheusMetrics += fmt.Sprintf("api_requests_total{version=\"%s\",environment=\"%s\"} %d\n",
		metrics.Version, metrics.Environment, metrics.APIMetrics.TotalRequests)

	prometheusMetrics += fmt.Sprintf("# HELP api_requests_per_second Number of API requests per second.\n")
	prometheusMetrics += fmt.Sprintf("# TYPE api_requests_per_second gauge\n")
	prometheusMetrics += fmt.Sprintf("api_requests_per_second{version=\"%s\",environment=\"%s\"} %f\n",
		metrics.Version, metrics.Environment, metrics.APIMetrics.RequestsPerSecond)

	// Business metrics
	prometheusMetrics += fmt.Sprintf("# HELP business_verifications_total Total number of business verifications.\n")
	prometheusMetrics += fmt.Sprintf("# TYPE business_verifications_total counter\n")
	prometheusMetrics += fmt.Sprintf("business_verifications_total{version=\"%s\",environment=\"%s\"} %d\n",
		metrics.Version, metrics.Environment, metrics.BusinessMetrics.TotalVerifications)

	prometheusMetrics += fmt.Sprintf("# HELP business_success_rate Success rate of business verifications.\n")
	prometheusMetrics += fmt.Sprintf("# TYPE business_success_rate gauge\n")
	prometheusMetrics += fmt.Sprintf("business_success_rate{version=\"%s\",environment=\"%s\"} %f\n",
		metrics.Version, metrics.Environment, metrics.BusinessMetrics.SuccessRate)

	// Performance metrics
	prometheusMetrics += fmt.Sprintf("# HELP performance_cpu_usage_percent CPU usage percentage.\n")
	prometheusMetrics += fmt.Sprintf("# TYPE performance_cpu_usage_percent gauge\n")
	prometheusMetrics += fmt.Sprintf("performance_cpu_usage_percent{version=\"%s\",environment=\"%s\"} %f\n",
		metrics.Version, metrics.Environment, metrics.PerformanceMetrics.CPUUsagePercent)

	prometheusMetrics += fmt.Sprintf("# HELP performance_memory_usage_percent Memory usage percentage.\n")
	prometheusMetrics += fmt.Sprintf("# TYPE performance_memory_usage_percent gauge\n")
	prometheusMetrics += fmt.Sprintf("performance_memory_usage_percent{version=\"%s\",environment=\"%s\"} %f\n",
		metrics.Version, metrics.Environment, metrics.PerformanceMetrics.MemoryUsagePercent)

	// Error metrics
	prometheusMetrics += fmt.Sprintf("# HELP error_total Total number of errors.\n")
	prometheusMetrics += fmt.Sprintf("# TYPE error_total counter\n")
	prometheusMetrics += fmt.Sprintf("error_total{version=\"%s\",environment=\"%s\"} %d\n",
		metrics.Version, metrics.Environment, metrics.ErrorMetrics.TotalErrors)

	prometheusMetrics += fmt.Sprintf("# HELP error_rate Error rate percentage.\n")
	prometheusMetrics += fmt.Sprintf("# TYPE error_rate gauge\n")
	prometheusMetrics += fmt.Sprintf("error_rate{version=\"%s\",environment=\"%s\"} %f\n",
		metrics.Version, metrics.Environment, metrics.ErrorMetrics.ErrorRate)

	return prometheusMetrics
}

// serveMetricsResponse serves metrics response
func (h *ComprehensiveMetricsHandler) serveMetricsResponse(w http.ResponseWriter, r *http.Request, metrics *ComprehensiveMetricsData, responseTime time.Duration) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(metrics)

	h.logger.Debug("Comprehensive metrics request served",
		zap.String("status", "success"),
		zap.String("client_ip", r.RemoteAddr),
		zap.Duration("response_time", responseTime),
		zap.String("user_agent", r.UserAgent()),
	)
}

// startMetricsCollection starts background metrics collection
func (h *ComprehensiveMetricsHandler) startMetricsCollection() {
	ticker := time.NewTicker(h.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			metrics := h.collectMetrics(ctx)

			h.mu.Lock()
			h.metrics = metrics
			h.lastUpdate = time.Now()
			h.mu.Unlock()

			h.logger.Debug("Background comprehensive metrics collection completed",
				zap.Time("timestamp", metrics.Timestamp),
				zap.Duration("uptime", metrics.SystemMetrics.Uptime),
			)
		}
	}
}
