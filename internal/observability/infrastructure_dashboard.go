package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// InfrastructureDashboard provides comprehensive infrastructure monitoring dashboard functionality
type InfrastructureDashboard struct {
	logger    *Logger
	config    *InfrastructureDashboardConfig
	infraData map[string]*InfrastructureData
	exporters []InfrastructureDashboardExporter
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	started   bool
}

// InfrastructureDashboardConfig holds configuration for infrastructure dashboard
type InfrastructureDashboardConfig struct {
	Enabled             bool
	RefreshInterval     time.Duration
	DataRetentionPeriod time.Duration
	MaxDataPoints       int
	ExportEnabled       bool
	ExportInterval      time.Duration
	Environment         string
	ServiceName         string
	Version             string
}

// InfrastructureData represents infrastructure dashboard data
type InfrastructureData struct {
	Timestamp            time.Time              `json:"timestamp"`
	SystemMetrics        *SystemMetrics         `json:"system_metrics"`
	DatabaseMetrics      *DatabaseMetrics       `json:"database_metrics"`
	NetworkMetrics       *NetworkMetrics        `json:"network_metrics"`
	StorageMetrics       *StorageMetrics        `json:"storage_metrics"`
	ExternalDependencies *ExternalDependencies  `json:"external_dependencies"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// SystemMetrics represents system resource metrics
type SystemMetrics struct {
	CPUUsage        float64       `json:"cpu_usage"`
	MemoryUsage     float64       `json:"memory_usage"`
	DiskUsage       float64       `json:"disk_usage"`
	LoadAverage     []float64     `json:"load_average"`
	Goroutines      int64         `json:"goroutines"`
	HeapAlloc       int64         `json:"heap_alloc"`
	HeapSys         int64         `json:"heap_sys"`
	GC              int64         `json:"gc"`
	Uptime          time.Duration `json:"uptime"`
	ProcessCount    int64         `json:"process_count"`
	FileDescriptors int64         `json:"file_descriptors"`
}

// DatabaseMetrics represents database performance metrics
type DatabaseMetrics struct {
	Connections      int64            `json:"connections"`
	MaxConnections   int64            `json:"max_connections"`
	QueryCount       int64            `json:"query_count"`
	SlowQueries      int64            `json:"slow_queries"`
	AverageQueryTime time.Duration    `json:"average_query_time"`
	LockWaits        int64            `json:"lock_waits"`
	Deadlocks        int64            `json:"deadlocks"`
	CacheHitRatio    float64          `json:"cache_hit_ratio"`
	DatabaseSize     int64            `json:"database_size"`
	IndexUsage       map[string]int64 `json:"index_usage"`
}

// NetworkMetrics represents network performance metrics
type NetworkMetrics struct {
	BytesIn         int64         `json:"bytes_in"`
	BytesOut        int64         `json:"bytes_out"`
	PacketsIn       int64         `json:"packets_in"`
	PacketsOut      int64         `json:"packets_out"`
	Connections     int64         `json:"connections"`
	Latency         time.Duration `json:"latency"`
	PacketLoss      float64       `json:"packet_loss"`
	Bandwidth       int64         `json:"bandwidth"`
	ErrorRate       float64       `json:"error_rate"`
	Retransmissions int64         `json:"retransmissions"`
}

// StorageMetrics represents storage performance metrics
type StorageMetrics struct {
	TotalSpace      int64         `json:"total_space"`
	UsedSpace       int64         `json:"used_space"`
	FreeSpace       int64         `json:"free_space"`
	IOPS            int64         `json:"iops"`
	ReadLatency     time.Duration `json:"read_latency"`
	WriteLatency    time.Duration `json:"write_latency"`
	Throughput      int64         `json:"throughput"`
	ErrorRate       float64       `json:"error_rate"`
	DiskUtilization float64       `json:"disk_utilization"`
}

// ExternalDependencies represents external service dependencies
type ExternalDependencies struct {
	TotalDependencies     int64                        `json:"total_dependencies"`
	HealthyDependencies   int64                        `json:"healthy_dependencies"`
	UnhealthyDependencies int64                        `json:"unhealthy_dependencies"`
	AverageResponseTime   time.Duration                `json:"average_response_time"`
	ErrorRate             float64                      `json:"error_rate"`
	Dependencies          map[string]*DependencyStatus `json:"dependencies"`
}

// DependencyStatus represents status of an external dependency
type DependencyStatus struct {
	Name         string        `json:"name"`
	URL          string        `json:"url"`
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	LastCheck    time.Time     `json:"last_check"`
	ErrorCount   int64         `json:"error_count"`
	SuccessCount int64         `json:"success_count"`
	Availability float64       `json:"availability"`
}

// InfrastructureDashboardExporter interface for exporting infrastructure dashboard data
type InfrastructureDashboardExporter interface {
	Export(data *InfrastructureData) error
	Name() string
	Type() string
}

// JSONInfrastructureDashboardExporter exports infrastructure dashboard data as JSON
type JSONInfrastructureDashboardExporter struct {
	logger *Logger
}

// NewJSONInfrastructureDashboardExporter creates a new JSON infrastructure dashboard exporter
func NewJSONInfrastructureDashboardExporter(logger *Logger) *JSONInfrastructureDashboardExporter {
	return &JSONInfrastructureDashboardExporter{
		logger: logger,
	}
}

// Export exports infrastructure dashboard data as JSON
func (jide *JSONInfrastructureDashboardExporter) Export(data *InfrastructureData) error {
	jide.logger.Debug("Infrastructure dashboard data exported as JSON", map[string]interface{}{
		"timestamp":    data.Timestamp,
		"cpu_usage":    data.SystemMetrics.CPUUsage,
		"memory_usage": data.SystemMetrics.MemoryUsage,
		"disk_usage":   data.SystemMetrics.DiskUsage,
		"connections":  data.DatabaseMetrics.Connections,
	})

	return nil
}

// Name returns the exporter name
func (jide *JSONInfrastructureDashboardExporter) Name() string {
	return "json"
}

// Type returns the exporter type
func (jide *JSONInfrastructureDashboardExporter) Type() string {
	return "json"
}

// NewInfrastructureDashboard creates a new infrastructure dashboard
func NewInfrastructureDashboard(
	logger *Logger,
	config *InfrastructureDashboardConfig,
) *InfrastructureDashboard {
	ctx, cancel := context.WithCancel(context.Background())

	return &InfrastructureDashboard{
		logger:    logger,
		config:    config,
		infraData: make(map[string]*InfrastructureData),
		exporters: make([]InfrastructureDashboardExporter, 0),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start starts the infrastructure dashboard
func (id *InfrastructureDashboard) Start() error {
	id.mu.Lock()
	defer id.mu.Unlock()

	if id.started {
		return fmt.Errorf("infrastructure dashboard already started")
	}

	id.logger.Info("Starting infrastructure dashboard", map[string]interface{}{
		"service_name": id.config.ServiceName,
		"version":      id.config.Version,
		"environment":  id.config.Environment,
	})

	// Start data collection
	if id.config.Enabled {
		go id.startDataCollection()
	}

	// Start data export
	if id.config.ExportEnabled {
		go id.startDataExport()
	}

	id.started = true
	id.logger.Info("Infrastructure dashboard started successfully", map[string]interface{}{})
	return nil
}

// Stop stops the infrastructure dashboard
func (id *InfrastructureDashboard) Stop() error {
	id.mu.Lock()
	defer id.mu.Unlock()

	if !id.started {
		return fmt.Errorf("infrastructure dashboard not started")
	}

	id.logger.Info("Stopping infrastructure dashboard", map[string]interface{}{})

	id.cancel()
	id.started = false

	id.logger.Info("Infrastructure dashboard stopped successfully", map[string]interface{}{})
	return nil
}

// GetInfrastructureData returns current infrastructure data
func (id *InfrastructureDashboard) GetInfrastructureData() (*InfrastructureData, error) {
	infraData := &InfrastructureData{
		Timestamp:            time.Now(),
		SystemMetrics:        id.collectSystemMetrics(),
		DatabaseMetrics:      id.collectDatabaseMetrics(),
		NetworkMetrics:       id.collectNetworkMetrics(),
		StorageMetrics:       id.collectStorageMetrics(),
		ExternalDependencies: id.collectExternalDependencies(),
		Metadata: map[string]interface{}{
			"service_name": id.config.ServiceName,
			"version":      id.config.Version,
			"environment":  id.config.Environment,
		},
	}

	return infraData, nil
}

// GetInfrastructureSummary returns an infrastructure summary
func (id *InfrastructureDashboard) GetInfrastructureSummary() (map[string]interface{}, error) {
	infraData, err := id.GetInfrastructureData()
	if err != nil {
		return nil, fmt.Errorf("failed to get infrastructure data: %w", err)
	}

	summary := map[string]interface{}{
		"system": map[string]interface{}{
			"cpu_usage":    infraData.SystemMetrics.CPUUsage,
			"memory_usage": infraData.SystemMetrics.MemoryUsage,
			"disk_usage":   infraData.SystemMetrics.DiskUsage,
			"goroutines":   infraData.SystemMetrics.Goroutines,
			"uptime":       infraData.SystemMetrics.Uptime,
		},
		"database": map[string]interface{}{
			"connections":     infraData.DatabaseMetrics.Connections,
			"query_count":     infraData.DatabaseMetrics.QueryCount,
			"slow_queries":    infraData.DatabaseMetrics.SlowQueries,
			"cache_hit_ratio": infraData.DatabaseMetrics.CacheHitRatio,
		},
		"network": map[string]interface{}{
			"bytes_in":    infraData.NetworkMetrics.BytesIn,
			"bytes_out":   infraData.NetworkMetrics.BytesOut,
			"connections": infraData.NetworkMetrics.Connections,
			"latency":     infraData.NetworkMetrics.Latency,
			"error_rate":  infraData.NetworkMetrics.ErrorRate,
		},
		"storage": map[string]interface{}{
			"total_space":      infraData.StorageMetrics.TotalSpace,
			"used_space":       infraData.StorageMetrics.UsedSpace,
			"free_space":       infraData.StorageMetrics.FreeSpace,
			"disk_utilization": infraData.StorageMetrics.DiskUtilization,
		},
		"dependencies": map[string]interface{}{
			"total_dependencies":     infraData.ExternalDependencies.TotalDependencies,
			"healthy_dependencies":   infraData.ExternalDependencies.HealthyDependencies,
			"unhealthy_dependencies": infraData.ExternalDependencies.UnhealthyDependencies,
			"average_response_time":  infraData.ExternalDependencies.AverageResponseTime,
		},
		"last_updated": infraData.Timestamp,
		"metadata":     infraData.Metadata,
	}

	return summary, nil
}

// AddExporter adds an infrastructure dashboard exporter
func (id *InfrastructureDashboard) AddExporter(exporter InfrastructureDashboardExporter) {
	id.mu.Lock()
	defer id.mu.Unlock()

	id.exporters = append(id.exporters, exporter)

	id.logger.Info("Infrastructure dashboard exporter added", map[string]interface{}{
		"exporter": exporter.Name(),
		"type":     exporter.Type(),
	})
}

// collectSystemMetrics collects system metrics
func (id *InfrastructureDashboard) collectSystemMetrics() *SystemMetrics {
	// In a real implementation, this would collect actual metrics
	// For now, return mock data
	return &SystemMetrics{
		CPUUsage:        45.0,
		MemoryUsage:     60.0,
		DiskUsage:       30.0,
		LoadAverage:     []float64{1.2, 1.5, 1.8},
		Goroutines:      150,
		HeapAlloc:       50 * 1024 * 1024,  // 50MB
		HeapSys:         100 * 1024 * 1024, // 100MB
		GC:              25,
		Uptime:          24 * time.Hour,
		ProcessCount:    1,
		FileDescriptors: 50,
	}
}

// collectDatabaseMetrics collects database metrics
func (id *InfrastructureDashboard) collectDatabaseMetrics() *DatabaseMetrics {
	// In a real implementation, this would collect actual metrics
	// For now, return mock data
	return &DatabaseMetrics{
		Connections:      25,
		MaxConnections:   100,
		QueryCount:       10000,
		SlowQueries:      50,
		AverageQueryTime: 10 * time.Millisecond,
		LockWaits:        5,
		Deadlocks:        0,
		CacheHitRatio:    0.95,
		DatabaseSize:     1024 * 1024 * 1024, // 1GB
		IndexUsage: map[string]int64{
			"primary": 5000,
			"index1":  3000,
			"index2":  2000,
		},
	}
}

// collectNetworkMetrics collects network metrics
func (id *InfrastructureDashboard) collectNetworkMetrics() *NetworkMetrics {
	// In a real implementation, this would collect actual metrics
	// For now, return mock data
	return &NetworkMetrics{
		BytesIn:         1024 * 1024 * 100, // 100MB
		BytesOut:        1024 * 1024 * 50,  // 50MB
		PacketsIn:       100000,
		PacketsOut:      80000,
		Connections:     150,
		Latency:         5 * time.Millisecond,
		PacketLoss:      0.01,
		Bandwidth:       1000 * 1024 * 1024, // 1Gbps
		ErrorRate:       0.001,
		Retransmissions: 10,
	}
}

// collectStorageMetrics collects storage metrics
func (id *InfrastructureDashboard) collectStorageMetrics() *StorageMetrics {
	// In a real implementation, this would collect actual metrics
	// For now, return mock data
	return &StorageMetrics{
		TotalSpace:      500 * 1024 * 1024 * 1024, // 500GB
		UsedSpace:       150 * 1024 * 1024 * 1024, // 150GB
		FreeSpace:       350 * 1024 * 1024 * 1024, // 350GB
		IOPS:            1000,
		ReadLatency:     2 * time.Millisecond,
		WriteLatency:    3 * time.Millisecond,
		Throughput:      100 * 1024 * 1024, // 100MB/s
		ErrorRate:       0.0001,
		DiskUtilization: 30.0,
	}
}

// collectExternalDependencies collects external dependencies
func (id *InfrastructureDashboard) collectExternalDependencies() *ExternalDependencies {
	// In a real implementation, this would collect actual metrics
	// For now, return mock data
	dependencies := map[string]*DependencyStatus{
		"database": {
			Name:         "PostgreSQL",
			URL:          "postgres://localhost:5432/kyb",
			Status:       "healthy",
			ResponseTime: 5 * time.Millisecond,
			LastCheck:    time.Now(),
			ErrorCount:   0,
			SuccessCount: 1000,
			Availability: 99.9,
		},
		"redis": {
			Name:         "Redis",
			URL:          "redis://localhost:6379",
			Status:       "healthy",
			ResponseTime: 1 * time.Millisecond,
			LastCheck:    time.Now(),
			ErrorCount:   0,
			SuccessCount: 2000,
			Availability: 99.95,
		},
		"external_api": {
			Name:         "External API",
			URL:          "https://api.external.com",
			Status:       "healthy",
			ResponseTime: 100 * time.Millisecond,
			LastCheck:    time.Now(),
			ErrorCount:   5,
			SuccessCount: 995,
			Availability: 99.5,
		},
	}

	healthyCount := 0
	unhealthyCount := 0
	totalResponseTime := time.Duration(0)
	totalErrors := int64(0)
	totalSuccesses := int64(0)

	for _, dep := range dependencies {
		if dep.Status == "healthy" {
			healthyCount++
		} else {
			unhealthyCount++
		}
		totalResponseTime += dep.ResponseTime
		totalErrors += dep.ErrorCount
		totalSuccesses += dep.SuccessCount
	}

	avgResponseTime := totalResponseTime / time.Duration(len(dependencies))
	errorRate := float64(totalErrors) / float64(totalErrors+totalSuccesses) * 100

	return &ExternalDependencies{
		TotalDependencies:     int64(len(dependencies)),
		HealthyDependencies:   int64(healthyCount),
		UnhealthyDependencies: int64(unhealthyCount),
		AverageResponseTime:   avgResponseTime,
		ErrorRate:             errorRate,
		Dependencies:          dependencies,
	}
}

// startDataCollection starts the data collection process
func (id *InfrastructureDashboard) startDataCollection() {
	ticker := time.NewTicker(id.config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-id.ctx.Done():
			id.logger.Info("Infrastructure data collection stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			id.collectInfrastructureData()
		}
	}
}

// collectInfrastructureData collects current infrastructure data
func (id *InfrastructureDashboard) collectInfrastructureData() {
	infraData, err := id.GetInfrastructureData()
	if err != nil {
		id.logger.Error("Failed to collect infrastructure data", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Store the data
	id.mu.Lock()
	key := infraData.Timestamp.Format("2006-01-02T15:04:05")
	id.infraData[key] = infraData

	// Clean up old data
	id.cleanupOldData()

	id.mu.Unlock()

	id.logger.Debug("Infrastructure data collected", map[string]interface{}{
		"cpu_usage":    infraData.SystemMetrics.CPUUsage,
		"memory_usage": infraData.SystemMetrics.MemoryUsage,
		"disk_usage":   infraData.SystemMetrics.DiskUsage,
		"connections":  infraData.DatabaseMetrics.Connections,
	})
}

// cleanupOldData removes old infrastructure data
func (id *InfrastructureDashboard) cleanupOldData() {
	cutoff := time.Now().Add(-id.config.DataRetentionPeriod)

	for key, data := range id.infraData {
		if data.Timestamp.Before(cutoff) {
			delete(id.infraData, key)
		}
	}

	// Limit the number of data points
	if len(id.infraData) > id.config.MaxDataPoints {
		// Remove oldest entries
		count := 0
		for key := range id.infraData {
			if count >= len(id.infraData)-id.config.MaxDataPoints {
				break
			}
			delete(id.infraData, key)
			count++
		}
	}
}

// startDataExport starts the data export process
func (id *InfrastructureDashboard) startDataExport() {
	ticker := time.NewTicker(id.config.ExportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-id.ctx.Done():
			id.logger.Info("Infrastructure data export stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			id.exportInfrastructureData()
		}
	}
}

// exportInfrastructureData exports current infrastructure data
func (id *InfrastructureDashboard) exportInfrastructureData() {
	infraData, err := id.GetInfrastructureData()
	if err != nil {
		id.logger.Error("Failed to get infrastructure data for export", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	for _, exporter := range id.exporters {
		if err := exporter.Export(infraData); err != nil {
			id.logger.Error("Failed to export infrastructure data", map[string]interface{}{
				"exporter": exporter.Name(),
				"error":    err.Error(),
			})
		}
	}
}
