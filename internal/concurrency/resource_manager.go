package concurrency

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"kyb-platform/internal/observability"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ResourceManager manages system resources for optimal performance
type ResourceManager struct {
	// Configuration
	config *ResourceManagerConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Resource monitoring
	cpuMonitor     *CPUMonitor
	memoryMonitor  *MemoryMonitor
	networkMonitor *NetworkMonitor
	diskMonitor    *DiskMonitor

	// Worker pool management
	workerPool    *ManagedWorkerPool
	workerPoolMux sync.RWMutex

	// Load balancing
	loadBalancer    *LoadBalancer
	loadBalancerMux sync.RWMutex

	// Resource allocation
	allocator    *ResourceAllocator
	allocatorMux sync.RWMutex

	// Monitoring and alerts
	monitor    *ResourceMonitor
	monitorMux sync.RWMutex

	// Context for shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// ResourceManagerConfig configuration for resource management
type ResourceManagerConfig struct {
	// CPU settings
	MaxCPUUsage      float64
	CPUThreshold     float64
	CPUCheckInterval time.Duration

	// Memory settings
	MaxMemoryUsage      float64
	MemoryThreshold     float64
	MemoryCheckInterval time.Duration

	// Worker pool settings
	MinWorkers        int
	MaxWorkers        int
	MaxConcurrentOps  int
	WorkerIdleTimeout time.Duration
	WorkerMaxTasks    int

	// Load balancing settings
	LoadBalancingStrategy string
	HealthCheckInterval   time.Duration
	LoadCheckInterval     time.Duration

	// Resource allocation settings
	AllocationStrategy      string
	ResourceTimeout         time.Duration
	AllocationCheckInterval time.Duration

	// Monitoring settings
	EnableMonitoring   bool
	MonitoringInterval time.Duration
	AlertThreshold     float64
	AlertCooldown      time.Duration

	// Network settings
	MaxNetworkUsage      float64
	NetworkThreshold     float64
	NetworkCheckInterval time.Duration

	// Disk settings
	MaxDiskUsage      float64
	DiskThreshold     float64
	DiskCheckInterval time.Duration
}

// CPUMonitor monitors CPU usage
type CPUMonitor struct {
	CurrentUsage   float64
	AverageUsage   float64
	PeakUsage      float64
	UsageHistory   []float64
	MaxHistorySize int
	LastUpdate     time.Time
	Mux            sync.RWMutex
}

// MemoryMonitor monitors memory usage
type MemoryMonitor struct {
	CurrentUsage   uint64
	MaxUsage       uint64
	AverageUsage   uint64
	UsageHistory   []uint64
	MaxHistorySize int
	LastUpdate     time.Time
	Mux            sync.RWMutex
}

// NetworkMonitor monitors network usage
type NetworkMonitor struct {
	BytesIn    uint64
	BytesOut   uint64
	PacketsIn  uint64
	PacketsOut uint64
	LastUpdate time.Time
	Mux        sync.RWMutex
}

// DiskMonitor monitors disk usage
type DiskMonitor struct {
	TotalSpace      uint64
	UsedSpace       uint64
	FreeSpace       uint64
	UsagePercentage float64
	LastUpdate      time.Time
	Mux             sync.RWMutex
}

// ManagedWorkerPool manages worker pool with resource constraints
type ManagedWorkerPool struct {
	Workers        map[string]*ManagedWorker
	IdleWorkers    []string
	ActiveWorkers  []string
	MaxWorkers     int
	MinWorkers     int
	CurrentWorkers int
	IdleTimeout    time.Duration
	MaxTasks       int
	Mux            sync.RWMutex
}

// ManagedWorker represents a managed worker with resource tracking
type ManagedWorker struct {
	ID            string
	Status        WorkerStatus
	CurrentTask   *TaskInfo
	TaskCount     int
	CPUUsage      float64
	MemoryUsage   uint64
	LastActivity  time.Time
	ResourceUsage *ResourceUsage
	HealthScore   float64
	Mux           sync.RWMutex
}

// LoadBalancer manages load distribution across workers
type LoadBalancer struct {
	Strategy      string
	Workers       map[string]*WorkerLoad
	HealthChecks  map[string]*HealthCheck
	LastRebalance time.Time
	Mux           sync.RWMutex
}

// WorkerLoad represents worker load information
type WorkerLoad struct {
	WorkerID     string
	CurrentLoad  float64
	AverageLoad  float64
	TaskQueue    int
	ResponseTime time.Duration
	ErrorRate    float64
	LastUpdate   time.Time
}

// HealthCheck represents worker health information
type HealthCheck struct {
	WorkerID     string
	Status       string
	LastCheck    time.Time
	ResponseTime time.Duration
	ErrorCount   int
	SuccessCount int
}

// ResourceAllocator manages resource allocation
type ResourceAllocator struct {
	Strategy       string
	Allocations    map[string]*ResourceAllocation
	ResourcePool   *ResourcePool
	LastAllocation time.Time
	Mux            sync.RWMutex
}

// ResourceAllocation represents a resource allocation
type ResourceAllocation struct {
	ID                string
	WorkerID          string
	CPUAllocation     float64
	MemoryAllocation  uint64
	NetworkAllocation uint64
	DiskAllocation    uint64
	AllocatedAt       time.Time
	ExpiresAt         time.Time
	Status            string
}

// ResourcePool represents available resources
type ResourcePool struct {
	TotalCPU         float64
	AvailableCPU     float64
	TotalMemory      uint64
	AvailableMemory  uint64
	TotalNetwork     uint64
	AvailableNetwork uint64
	TotalDisk        uint64
	AvailableDisk    uint64
	LastUpdate       time.Time
	Mux              sync.RWMutex
}

// ResourceMonitor monitors overall resource usage
type ResourceMonitor struct {
	Alerts     []*ResourceAlert
	Metrics    *ResourceMetrics
	Thresholds map[string]float64
	LastAlert  time.Time
	Mux        sync.RWMutex
}

// ResourceAlert represents a resource alert
type ResourceAlert struct {
	ID           string
	Type         string
	Severity     string
	Message      string
	Resource     string
	Value        float64
	Threshold    float64
	Timestamp    time.Time
	Acknowledged bool
}

// ResourceMetrics represents resource metrics
type ResourceMetrics struct {
	CPUUtilization     float64
	MemoryUtilization  float64
	NetworkUtilization float64
	DiskUtilization    float64
	WorkerUtilization  float64
	QueueLength        int
	ResponseTime       time.Duration
	Throughput         float64
	ErrorRate          float64
	LastUpdate         time.Time
}

// NewResourceManager creates a new resource manager
func NewResourceManager(config *ResourceManagerConfig, logger *observability.Logger, tracer trace.Tracer) *ResourceManager {
	if config == nil {
		config = &ResourceManagerConfig{
			MaxCPUUsage:             80.0,
			CPUThreshold:            70.0,
			CPUCheckInterval:        5 * time.Second,
			MaxMemoryUsage:          80.0,
			MemoryThreshold:         70.0,
			MemoryCheckInterval:     5 * time.Second,
			MinWorkers:              2,
			MaxWorkers:              10,
			WorkerIdleTimeout:       5 * time.Minute,
			WorkerMaxTasks:          100,
			LoadBalancingStrategy:   "round_robin",
			HealthCheckInterval:     30 * time.Second,
			LoadCheckInterval:       10 * time.Second,
			AllocationStrategy:      "fair",
			ResourceTimeout:         10 * time.Minute,
			AllocationCheckInterval: 1 * time.Minute,
			EnableMonitoring:        true,
			MonitoringInterval:      30 * time.Second,
			AlertThreshold:          80.0,
			AlertCooldown:           5 * time.Minute,
			MaxNetworkUsage:         80.0,
			NetworkThreshold:        70.0,
			NetworkCheckInterval:    10 * time.Second,
			MaxDiskUsage:            80.0,
			DiskThreshold:           70.0,
			DiskCheckInterval:       30 * time.Second,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	rm := &ResourceManager{
		config: config,
		logger: logger,
		tracer: tracer,
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize monitors
	rm.cpuMonitor = &CPUMonitor{
		MaxHistorySize: 100,
		UsageHistory:   make([]float64, 0, 100),
	}
	rm.memoryMonitor = &MemoryMonitor{
		MaxHistorySize: 100,
		UsageHistory:   make([]uint64, 0, 100),
	}
	rm.networkMonitor = &NetworkMonitor{}
	rm.diskMonitor = &DiskMonitor{}

	// Initialize worker pool
	rm.workerPool = &ManagedWorkerPool{
		Workers:       make(map[string]*ManagedWorker),
		IdleWorkers:   make([]string, 0),
		ActiveWorkers: make([]string, 0),
		MaxWorkers:    config.MaxWorkers,
		MinWorkers:    config.MinWorkers,
		IdleTimeout:   config.WorkerIdleTimeout,
		MaxTasks:      config.WorkerMaxTasks,
	}

	// Initialize load balancer
	rm.loadBalancer = &LoadBalancer{
		Strategy:     config.LoadBalancingStrategy,
		Workers:      make(map[string]*WorkerLoad),
		HealthChecks: make(map[string]*HealthCheck),
	}

	// Initialize resource allocator
	rm.allocator = &ResourceAllocator{
		Strategy:    config.AllocationStrategy,
		Allocations: make(map[string]*ResourceAllocation),
		ResourcePool: &ResourcePool{
			TotalCPU:        float64(runtime.NumCPU()),
			AvailableCPU:    float64(runtime.NumCPU()),
			TotalMemory:     getTotalMemory(),
			AvailableMemory: getTotalMemory(),
		},
	}

	// Initialize monitor
	rm.monitor = &ResourceMonitor{
		Alerts:     make([]*ResourceAlert, 0),
		Metrics:    &ResourceMetrics{},
		Thresholds: make(map[string]float64),
	}

	// Set default thresholds
	rm.monitor.Thresholds["cpu"] = config.CPUThreshold
	rm.monitor.Thresholds["memory"] = config.MemoryThreshold
	rm.monitor.Thresholds["network"] = config.NetworkThreshold
	rm.monitor.Thresholds["disk"] = config.DiskThreshold

	// Start background workers
	rm.startBackgroundWorkers()

	return rm
}

// startBackgroundWorkers starts background monitoring workers
func (rm *ResourceManager) startBackgroundWorkers() {
	// CPU monitoring worker
	go rm.cpuMonitoringWorker()

	// Memory monitoring worker
	go rm.memoryMonitoringWorker()

	// Network monitoring worker
	go rm.networkMonitoringWorker()

	// Disk monitoring worker
	go rm.diskMonitoringWorker()

	// Worker pool management worker
	go rm.workerPoolManagementWorker()

	// Load balancing worker
	go rm.loadBalancingWorker()

	// Resource allocation worker
	go rm.resourceAllocationWorker()

	// Monitoring and alerting worker
	go rm.monitoringWorker()
}

// cpuMonitoringWorker monitors CPU usage
func (rm *ResourceManager) cpuMonitoringWorker() {
	ticker := time.NewTicker(rm.config.CPUCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.updateCPUUsage()
		}
	}
}

// memoryMonitoringWorker monitors memory usage
func (rm *ResourceManager) memoryMonitoringWorker() {
	ticker := time.NewTicker(rm.config.MemoryCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.updateMemoryUsage()
		}
	}
}

// networkMonitoringWorker monitors network usage
func (rm *ResourceManager) networkMonitoringWorker() {
	ticker := time.NewTicker(rm.config.NetworkCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.updateNetworkUsage()
		}
	}
}

// diskMonitoringWorker monitors disk usage
func (rm *ResourceManager) diskMonitoringWorker() {
	ticker := time.NewTicker(rm.config.DiskCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.updateDiskUsage()
		}
	}
}

// workerPoolManagementWorker manages worker pool
func (rm *ResourceManager) workerPoolManagementWorker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.manageWorkerPool()
		}
	}
}

// loadBalancingWorker manages load balancing
func (rm *ResourceManager) loadBalancingWorker() {
	ticker := time.NewTicker(rm.config.LoadCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.updateLoadBalancing()
		}
	}
}

// resourceAllocationWorker manages resource allocation
func (rm *ResourceManager) resourceAllocationWorker() {
	ticker := time.NewTicker(rm.config.AllocationCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.updateResourceAllocation()
		}
	}
}

// monitoringWorker handles monitoring and alerting
func (rm *ResourceManager) monitoringWorker() {
	ticker := time.NewTicker(rm.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.updateMonitoring()
		}
	}
}

// updateCPUUsage updates CPU usage metrics
func (rm *ResourceManager) updateCPUUsage() {
	_, span := rm.tracer.Start(rm.ctx, "ResourceManager.updateCPUUsage")
	defer span.End()

	// Get CPU usage (simplified implementation)
	cpuUsage := getCPUUsage()

	rm.cpuMonitor.Mux.Lock()
	defer rm.cpuMonitor.Mux.Unlock()

	rm.cpuMonitor.CurrentUsage = cpuUsage
	rm.cpuMonitor.UsageHistory = append(rm.cpuMonitor.UsageHistory, cpuUsage)

	// Keep history size manageable
	if len(rm.cpuMonitor.UsageHistory) > rm.cpuMonitor.MaxHistorySize {
		rm.cpuMonitor.UsageHistory = rm.cpuMonitor.UsageHistory[1:]
	}

	// Calculate average
	total := 0.0
	for _, usage := range rm.cpuMonitor.UsageHistory {
		total += usage
	}
	rm.cpuMonitor.AverageUsage = total / float64(len(rm.cpuMonitor.UsageHistory))

	// Update peak usage
	if cpuUsage > rm.cpuMonitor.PeakUsage {
		rm.cpuMonitor.PeakUsage = cpuUsage
	}

	rm.cpuMonitor.LastUpdate = time.Now()

	// Check for alerts
	if cpuUsage > rm.config.CPUThreshold {
		rm.createAlert("cpu", "high", fmt.Sprintf("CPU usage is %.2f%%", cpuUsage), cpuUsage, rm.config.CPUThreshold)
	}

	span.SetAttributes(
		attribute.Float64("cpu_usage", cpuUsage),
		attribute.Float64("cpu_average", rm.cpuMonitor.AverageUsage),
		attribute.Float64("cpu_peak", rm.cpuMonitor.PeakUsage),
	)
}

// updateMemoryUsage updates memory usage metrics
func (rm *ResourceManager) updateMemoryUsage() {
	_, span := rm.tracer.Start(rm.ctx, "ResourceManager.updateMemoryUsage")
	defer span.End()

	// Get memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	rm.memoryMonitor.Mux.Lock()
	defer rm.memoryMonitor.Mux.Unlock()

	rm.memoryMonitor.CurrentUsage = m.Alloc
	rm.memoryMonitor.UsageHistory = append(rm.memoryMonitor.UsageHistory, m.Alloc)

	// Keep history size manageable
	if len(rm.memoryMonitor.UsageHistory) > rm.memoryMonitor.MaxHistorySize {
		rm.memoryMonitor.UsageHistory = rm.memoryMonitor.UsageHistory[1:]
	}

	// Calculate average
	total := uint64(0)
	for _, usage := range rm.memoryMonitor.UsageHistory {
		total += usage
	}
	rm.memoryMonitor.AverageUsage = total / uint64(len(rm.memoryMonitor.UsageHistory))

	// Update max usage
	if m.Alloc > rm.memoryMonitor.MaxUsage {
		rm.memoryMonitor.MaxUsage = m.Alloc
	}

	rm.memoryMonitor.LastUpdate = time.Now()

	// Calculate usage percentage
	usagePercentage := float64(m.Alloc) / float64(rm.memoryMonitor.MaxUsage) * 100

	// Check for alerts
	if usagePercentage > rm.config.MemoryThreshold {
		rm.createAlert("memory", "high", fmt.Sprintf("Memory usage is %.2f%%", usagePercentage), usagePercentage, rm.config.MemoryThreshold)
	}

	span.SetAttributes(
		attribute.Int64("memory_usage", int64(m.Alloc)),
		attribute.Int64("memory_average", int64(rm.memoryMonitor.AverageUsage)),
		attribute.Int64("memory_max", int64(rm.memoryMonitor.MaxUsage)),
		attribute.Float64("memory_percentage", usagePercentage),
	)
}

// updateNetworkUsage updates network usage metrics
func (rm *ResourceManager) updateNetworkUsage() {
	_, span := rm.tracer.Start(rm.ctx, "ResourceManager.updateNetworkUsage")
	defer span.End()

	// Get network usage (simplified implementation)
	bytesIn, bytesOut := getNetworkUsage()

	rm.networkMonitor.Mux.Lock()
	defer rm.networkMonitor.Mux.Unlock()

	rm.networkMonitor.BytesIn = bytesIn
	rm.networkMonitor.BytesOut = bytesOut
	rm.networkMonitor.LastUpdate = time.Now()

	span.SetAttributes(
		attribute.Int64("network_bytes_in", int64(bytesIn)),
		attribute.Int64("network_bytes_out", int64(bytesOut)),
	)
}

// updateDiskUsage updates disk usage metrics
func (rm *ResourceManager) updateDiskUsage() {
	_, span := rm.tracer.Start(rm.ctx, "ResourceManager.updateDiskUsage")
	defer span.End()

	// Get disk usage (simplified implementation)
	totalSpace, usedSpace, freeSpace := getDiskUsage()

	rm.diskMonitor.Mux.Lock()
	defer rm.diskMonitor.Mux.Unlock()

	rm.diskMonitor.TotalSpace = totalSpace
	rm.diskMonitor.UsedSpace = usedSpace
	rm.diskMonitor.FreeSpace = freeSpace
	rm.diskMonitor.UsagePercentage = float64(usedSpace) / float64(totalSpace) * 100
	rm.diskMonitor.LastUpdate = time.Now()

	// Check for alerts
	if rm.diskMonitor.UsagePercentage > rm.config.DiskThreshold {
		rm.createAlert("disk", "high", fmt.Sprintf("Disk usage is %.2f%%", rm.diskMonitor.UsagePercentage), rm.diskMonitor.UsagePercentage, rm.config.DiskThreshold)
	}

	span.SetAttributes(
		attribute.Int64("disk_total", int64(totalSpace)),
		attribute.Int64("disk_used", int64(usedSpace)),
		attribute.Int64("disk_free", int64(freeSpace)),
		attribute.Float64("disk_percentage", rm.diskMonitor.UsagePercentage),
	)
}

// manageWorkerPool manages the worker pool
func (rm *ResourceManager) manageWorkerPool() {
	_, span := rm.tracer.Start(rm.ctx, "ResourceManager.manageWorkerPool")
	defer span.End()

	rm.workerPool.Mux.Lock()
	defer rm.workerPool.Mux.Unlock()

	// Scale up if needed
	if rm.shouldScaleUp() {
		rm.scaleUp()
	}

	// Scale down if needed
	if rm.shouldScaleDown() {
		rm.scaleDown()
	}

	// Clean up idle workers
	rm.cleanupIdleWorkers()

	// Update worker health
	rm.updateWorkerHealth()

	span.SetAttributes(
		attribute.Int("current_workers", rm.workerPool.CurrentWorkers),
		attribute.Int("active_workers", len(rm.workerPool.ActiveWorkers)),
		attribute.Int("idle_workers", len(rm.workerPool.IdleWorkers)),
	)
}

// shouldScaleUp determines if we should scale up
func (rm *ResourceManager) shouldScaleUp() bool {
	// Scale up if CPU usage is high and we have capacity
	if rm.cpuMonitor.CurrentUsage > rm.config.CPUThreshold &&
		rm.workerPool.CurrentWorkers < rm.workerPool.MaxWorkers {
		return true
	}

	// Scale up if queue length is high
	if len(rm.workerPool.ActiveWorkers) == rm.workerPool.CurrentWorkers &&
		rm.workerPool.CurrentWorkers < rm.workerPool.MaxWorkers {
		return true
	}

	return false
}

// shouldScaleDown determines if we should scale down
func (rm *ResourceManager) shouldScaleDown() bool {
	// Scale down if CPU usage is low and we have excess workers
	if rm.cpuMonitor.CurrentUsage < rm.config.CPUThreshold/2 &&
		rm.workerPool.CurrentWorkers > rm.workerPool.MinWorkers {
		return true
	}

	// Scale down if we have idle workers for too long
	if len(rm.workerPool.IdleWorkers) > 0 &&
		rm.workerPool.CurrentWorkers > rm.workerPool.MinWorkers {
		return true
	}

	return false
}

// scaleUp scales up the worker pool
func (rm *ResourceManager) scaleUp() {
	workerID := fmt.Sprintf("worker-%d", rm.workerPool.CurrentWorkers+1)

	worker := &ManagedWorker{
		ID:           workerID,
		Status:       WorkerStatusIdle,
		LastActivity: time.Now(),
		HealthScore:  1.0,
	}

	rm.workerPool.Workers[workerID] = worker
	rm.workerPool.IdleWorkers = append(rm.workerPool.IdleWorkers, workerID)
	rm.workerPool.CurrentWorkers++

	rm.logger.Info("worker pool scaled up", map[string]interface{}{
		"worker_id":     workerID,
		"total_workers": rm.workerPool.CurrentWorkers,
	})
}

// scaleDown scales down the worker pool
func (rm *ResourceManager) scaleDown() {
	if len(rm.workerPool.IdleWorkers) == 0 {
		return
	}

	// Remove the oldest idle worker
	workerID := rm.workerPool.IdleWorkers[0]
	rm.workerPool.IdleWorkers = rm.workerPool.IdleWorkers[1:]

	delete(rm.workerPool.Workers, workerID)
	rm.workerPool.CurrentWorkers--

	rm.logger.Info("worker pool scaled down", map[string]interface{}{
		"worker_id":     workerID,
		"total_workers": rm.workerPool.CurrentWorkers,
	})
}

// cleanupIdleWorkers cleans up idle workers
func (rm *ResourceManager) cleanupIdleWorkers() {
	now := time.Now()
	newIdleWorkers := make([]string, 0)

	for _, workerID := range rm.workerPool.IdleWorkers {
		worker := rm.workerPool.Workers[workerID]
		if worker == nil {
			continue
		}

		// Check if worker has been idle too long
		if now.Sub(worker.LastActivity) > rm.workerPool.IdleTimeout {
			delete(rm.workerPool.Workers, workerID)
			rm.workerPool.CurrentWorkers--

			rm.logger.Info("idle worker cleaned up", map[string]interface{}{
				"worker_id":     workerID,
				"idle_duration": now.Sub(worker.LastActivity),
			})
		} else {
			newIdleWorkers = append(newIdleWorkers, workerID)
		}
	}

	rm.workerPool.IdleWorkers = newIdleWorkers
}

// updateWorkerHealth updates worker health scores
func (rm *ResourceManager) updateWorkerHealth() {
	for _, worker := range rm.workerPool.Workers {
		worker.Mux.Lock()

		// Calculate health score based on various factors
		healthScore := 1.0

		// Reduce score for high CPU usage
		if worker.CPUUsage > 80 {
			healthScore -= 0.2
		}

		// Reduce score for high memory usage
		if worker.MemoryUsage > rm.memoryMonitor.MaxUsage*80/100 {
			healthScore -= 0.2
		}

		// Reduce score for inactivity
		if time.Since(worker.LastActivity) > 5*time.Minute {
			healthScore -= 0.1
		}

		// Ensure health score is between 0 and 1
		if healthScore < 0 {
			healthScore = 0
		}
		if healthScore > 1 {
			healthScore = 1
		}

		worker.HealthScore = healthScore
		worker.Mux.Unlock()
	}
}

// updateLoadBalancing updates load balancing
func (rm *ResourceManager) updateLoadBalancing() {
	_, span := rm.tracer.Start(rm.ctx, "ResourceManager.updateLoadBalancing")
	defer span.End()

	rm.loadBalancer.Mux.Lock()
	defer rm.loadBalancer.Mux.Unlock()

	// Update worker loads
	for workerID, worker := range rm.workerPool.Workers {
		worker.Mux.RLock()
		load := &WorkerLoad{
			WorkerID:    workerID,
			CurrentLoad: worker.CPUUsage,
			TaskQueue:   worker.TaskCount,
			LastUpdate:  time.Now(),
		}
		worker.Mux.RUnlock()

		rm.loadBalancer.Workers[workerID] = load
	}

	// Perform health checks
	rm.performHealthChecks()

	span.SetAttributes(
		attribute.Int("total_workers", len(rm.loadBalancer.Workers)),
		attribute.String("strategy", rm.loadBalancer.Strategy),
	)
}

// performHealthChecks performs health checks on workers
func (rm *ResourceManager) performHealthChecks() {
	for workerID, worker := range rm.workerPool.Workers {
		healthCheck := &HealthCheck{
			WorkerID:  workerID,
			Status:    "healthy",
			LastCheck: time.Now(),
		}

		worker.Mux.RLock()
		if worker.HealthScore < 0.5 {
			healthCheck.Status = "unhealthy"
			healthCheck.ErrorCount++
		} else {
			healthCheck.SuccessCount++
		}
		worker.Mux.RUnlock()

		rm.loadBalancer.HealthChecks[workerID] = healthCheck
	}
}

// updateResourceAllocation updates resource allocation
func (rm *ResourceManager) updateResourceAllocation() {
	_, span := rm.tracer.Start(rm.ctx, "ResourceManager.updateResourceAllocation")
	defer span.End()

	rm.allocator.Mux.Lock()
	defer rm.allocator.Mux.Unlock()

	// Update resource pool
	rm.allocator.ResourcePool.TotalCPU = float64(runtime.NumCPU())
	rm.allocator.ResourcePool.AvailableCPU = rm.allocator.ResourcePool.TotalCPU
	rm.allocator.ResourcePool.TotalMemory = getTotalMemory()
	rm.allocator.ResourcePool.AvailableMemory = rm.allocator.ResourcePool.TotalMemory

	// Calculate allocated resources
	for _, allocation := range rm.allocator.Allocations {
		if allocation.Status == "active" {
			rm.allocator.ResourcePool.AvailableCPU -= allocation.CPUAllocation
			rm.allocator.ResourcePool.AvailableMemory -= allocation.MemoryAllocation
		}
	}

	// Clean up expired allocations
	rm.cleanupExpiredAllocations()

	rm.allocator.ResourcePool.LastUpdate = time.Now()

	span.SetAttributes(
		attribute.Float64("available_cpu", rm.allocator.ResourcePool.AvailableCPU),
		attribute.Int64("available_memory", int64(rm.allocator.ResourcePool.AvailableMemory)),
		attribute.Int("active_allocations", len(rm.allocator.Allocations)),
	)
}

// cleanupExpiredAllocations cleans up expired resource allocations
func (rm *ResourceManager) cleanupExpiredAllocations() {
	now := time.Now()
	for allocationID, allocation := range rm.allocator.Allocations {
		if allocation.ExpiresAt.Before(now) {
			allocation.Status = "expired"
			rm.logger.Info("resource allocation expired", map[string]interface{}{
				"allocation_id": allocationID,
				"worker_id":     allocation.WorkerID,
			})
		}
	}
}

// updateMonitoring updates monitoring and alerting
func (rm *ResourceManager) updateMonitoring() {
	_, span := rm.tracer.Start(rm.ctx, "ResourceManager.updateMonitoring")
	defer span.End()

	rm.monitor.Mux.Lock()
	defer rm.monitor.Mux.Unlock()

	// Update metrics
	rm.monitor.Metrics.CPUUtilization = rm.cpuMonitor.CurrentUsage
	rm.monitor.Metrics.MemoryUtilization = float64(rm.memoryMonitor.CurrentUsage) / float64(rm.memoryMonitor.MaxUsage) * 100
	rm.monitor.Metrics.NetworkUtilization = 0 // Simplified
	rm.monitor.Metrics.DiskUtilization = rm.diskMonitor.UsagePercentage
	rm.monitor.Metrics.WorkerUtilization = float64(len(rm.workerPool.ActiveWorkers)) / float64(rm.workerPool.CurrentWorkers) * 100
	rm.monitor.Metrics.QueueLength = len(rm.workerPool.IdleWorkers)
	rm.monitor.Metrics.LastUpdate = time.Now()

	// Clean up old alerts
	rm.cleanupOldAlerts()

	span.SetAttributes(
		attribute.Float64("cpu_utilization", rm.monitor.Metrics.CPUUtilization),
		attribute.Float64("memory_utilization", rm.monitor.Metrics.MemoryUtilization),
		attribute.Float64("worker_utilization", rm.monitor.Metrics.WorkerUtilization),
	)
}

// createAlert creates a resource alert
func (rm *ResourceManager) createAlert(resourceType, severity, message string, value, threshold float64) {
	// Check cooldown
	if time.Since(rm.monitor.LastAlert) < rm.config.AlertCooldown {
		return
	}

	alert := &ResourceAlert{
		ID:        fmt.Sprintf("alert-%d", time.Now().Unix()),
		Type:      resourceType,
		Severity:  severity,
		Message:   message,
		Resource:  resourceType,
		Value:     value,
		Threshold: threshold,
		Timestamp: time.Now(),
	}

	rm.monitor.Alerts = append(rm.monitor.Alerts, alert)
	rm.monitor.LastAlert = time.Now()

	rm.logger.Warn("resource alert created", map[string]interface{}{
		"alert_id":  alert.ID,
		"type":      alert.Type,
		"severity":  alert.Severity,
		"message":   alert.Message,
		"value":     alert.Value,
		"threshold": alert.Threshold,
	})
}

// cleanupOldAlerts cleans up old alerts
func (rm *ResourceManager) cleanupOldAlerts() {
	cutoff := time.Now().Add(-24 * time.Hour) // Keep alerts for 24 hours
	newAlerts := make([]*ResourceAlert, 0)

	for _, alert := range rm.monitor.Alerts {
		if alert.Timestamp.After(cutoff) {
			newAlerts = append(newAlerts, alert)
		}
	}

	rm.monitor.Alerts = newAlerts
}

// GetWorker returns the best available worker
func (rm *ResourceManager) GetWorker() *ManagedWorker {
	rm.workerPool.Mux.RLock()
	defer rm.workerPool.Mux.RUnlock()

	// Use load balancing strategy
	switch rm.loadBalancer.Strategy {
	case "round_robin":
		return rm.getWorkerRoundRobin()
	case "least_loaded":
		return rm.getWorkerLeastLoaded()
	case "health_based":
		return rm.getWorkerHealthBased()
	default:
		return rm.getWorkerRoundRobin()
	}
}

// getWorkerRoundRobin returns worker using round-robin strategy
func (rm *ResourceManager) getWorkerRoundRobin() *ManagedWorker {
	if len(rm.workerPool.IdleWorkers) == 0 {
		return nil
	}

	// Get the first idle worker
	workerID := rm.workerPool.IdleWorkers[0]
	worker := rm.workerPool.Workers[workerID]

	// Move to active workers
	rm.workerPool.IdleWorkers = rm.workerPool.IdleWorkers[1:]
	rm.workerPool.ActiveWorkers = append(rm.workerPool.ActiveWorkers, workerID)

	return worker
}

// getWorkerLeastLoaded returns the least loaded worker
func (rm *ResourceManager) getWorkerLeastLoaded() *ManagedWorker {
	var bestWorker *ManagedWorker
	lowestLoad := float64(100)

	for _, workerID := range rm.workerPool.IdleWorkers {
		worker := rm.workerPool.Workers[workerID]
		if worker.CPUUsage < lowestLoad {
			lowestLoad = worker.CPUUsage
			bestWorker = worker
		}
	}

	if bestWorker != nil {
		// Move to active workers
		rm.removeFromIdleWorkers(bestWorker.ID)
		rm.workerPool.ActiveWorkers = append(rm.workerPool.ActiveWorkers, bestWorker.ID)
	}

	return bestWorker
}

// getWorkerHealthBased returns the healthiest worker
func (rm *ResourceManager) getWorkerHealthBased() *ManagedWorker {
	var bestWorker *ManagedWorker
	highestHealth := float64(0)

	for _, workerID := range rm.workerPool.IdleWorkers {
		worker := rm.workerPool.Workers[workerID]
		if worker.HealthScore > highestHealth {
			highestHealth = worker.HealthScore
			bestWorker = worker
		}
	}

	if bestWorker != nil {
		// Move to active workers
		rm.removeFromIdleWorkers(bestWorker.ID)
		rm.workerPool.ActiveWorkers = append(rm.workerPool.ActiveWorkers, bestWorker.ID)
	}

	return bestWorker
}

// removeFromIdleWorkers removes a worker from idle workers
func (rm *ResourceManager) removeFromIdleWorkers(workerID string) {
	for i, id := range rm.workerPool.IdleWorkers {
		if id == workerID {
			rm.workerPool.IdleWorkers = append(rm.workerPool.IdleWorkers[:i], rm.workerPool.IdleWorkers[i+1:]...)
			break
		}
	}
}

// ReleaseWorker releases a worker back to the pool
func (rm *ResourceManager) ReleaseWorker(workerID string) {
	rm.workerPool.Mux.Lock()
	defer rm.workerPool.Mux.Unlock()

	worker := rm.workerPool.Workers[workerID]
	if worker == nil {
		return
	}

	// Update worker status
	worker.Status = WorkerStatusIdle
	worker.CurrentTask = nil
	worker.LastActivity = time.Now()

	// Move to idle workers
	rm.removeFromActiveWorkers(workerID)
	rm.workerPool.IdleWorkers = append(rm.workerPool.IdleWorkers, workerID)
}

// removeFromActiveWorkers removes a worker from active workers
func (rm *ResourceManager) removeFromActiveWorkers(workerID string) {
	for i, id := range rm.workerPool.ActiveWorkers {
		if id == workerID {
			rm.workerPool.ActiveWorkers = append(rm.workerPool.ActiveWorkers[:i], rm.workerPool.ActiveWorkers[i+1:]...)
			break
		}
	}
}

// GetResourceMetrics returns current resource metrics
func (rm *ResourceManager) GetResourceMetrics() *ResourceMetrics {
	rm.monitor.Mux.RLock()
	defer rm.monitor.Mux.RUnlock()

	return rm.monitor.Metrics
}

// GetAlerts returns current alerts
func (rm *ResourceManager) GetAlerts() []*ResourceAlert {
	rm.monitor.Mux.RLock()
	defer rm.monitor.Mux.RUnlock()

	return rm.monitor.Alerts
}

// Shutdown shuts down the resource manager
func (rm *ResourceManager) Shutdown() {
	rm.cancel()
	rm.logger.Info("resource manager shutting down", map[string]interface{}{})
}

// Start starts the resource manager
func (rm *ResourceManager) Start() error {
	rm.startBackgroundWorkers()
	rm.logger.Info("resource manager started", map[string]interface{}{})
	return nil
}

// Stop stops the resource manager
func (rm *ResourceManager) Stop() {
	rm.Shutdown()
}

// Acquire acquires a resource
func (rm *ResourceManager) Acquire(ctx context.Context, resourceTypes []string) ([]*Resource, error) {
	// Simplified implementation - just return success for now
	resources := make([]*Resource, len(resourceTypes))
	for i, resourceType := range resourceTypes {
		resources[i] = &Resource{
			ID:        resourceType,
			Type:      resourceType,
			Capacity:  1,
			Used:      1,
			Available: 0,
		}
	}
	return resources, nil
}

// Release releases a resource
func (rm *ResourceManager) Release(resources []*Resource) error {
	// Simplified implementation - just return success for now
	return nil
}

// GetStats returns resource manager statistics
func (rm *ResourceManager) GetStats() *ResourceMetrics {
	return rm.GetResourceMetrics()
}

// Helper functions (simplified implementations)

// getCPUUsage returns current CPU usage percentage
func getCPUUsage() float64 {
	// Simplified implementation - in production, use proper CPU monitoring
	return 50.0 + float64(time.Now().Unix()%30) // Simulate varying CPU usage
}

// getTotalMemory returns total system memory
func getTotalMemory() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Sys
}

// getNetworkUsage returns network usage statistics
func getNetworkUsage() (uint64, uint64) {
	// Simplified implementation - in production, use proper network monitoring
	return 1024 * 1024, 512 * 1024 // 1MB in, 512KB out
}

// getDiskUsage returns disk usage statistics
func getDiskUsage() (uint64, uint64, uint64) {
	// Simplified implementation - in production, use proper disk monitoring
	total := uint64(100 * 1024 * 1024 * 1024) // 100GB
	used := uint64(60 * 1024 * 1024 * 1024)   // 60GB
	free := total - used
	return total, used, free
}
