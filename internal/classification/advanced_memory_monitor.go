package classification

import (
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AdvancedMemoryMonitor provides comprehensive memory monitoring with garbage collection tracking
type AdvancedMemoryMonitor struct {
	logger         *zap.Logger
	config         *MemoryMonitorConfig
	stats          *AdvancedMemoryStats
	history        []*AdvancedMemoryStats
	maxHistorySize int
	mu             sync.RWMutex

	// GC tracking
	gcStats        *GCStats
	gcHistory      []*GCStats
	gcEventChannel chan *GCEvent

	// Memory pressure detection
	pressureThreshold float64
	pressureEvents    []*MemoryPressureEvent

	// Monitoring control
	stopCh           chan struct{}
	monitoringActive bool
}

// MemoryMonitorConfig holds configuration for advanced memory monitoring
type MemoryMonitorConfig struct {
	Enabled                bool          `json:"enabled"`
	CollectionInterval     time.Duration `json:"collection_interval"`
	MaxHistorySize         int           `json:"max_history_size"`
	PressureThreshold      float64       `json:"pressure_threshold"`    // Percentage of memory usage to trigger pressure
	GCThreshold            time.Duration `json:"gc_threshold"`          // GC pause time threshold
	MemoryLeakThreshold    float64       `json:"memory_leak_threshold"` // Memory growth rate threshold
	AlertingEnabled        bool          `json:"alerting_enabled"`
	DetailedGCStats        bool          `json:"detailed_gc_stats"`
	TrackMemoryAllocations bool          `json:"track_memory_allocations"`
}

// AdvancedMemoryStats represents comprehensive memory statistics
type AdvancedMemoryStats struct {
	Timestamp time.Time `json:"timestamp"`

	// Basic memory stats
	AllocatedMB      float64 `json:"allocated_mb"`
	TotalAllocatedMB float64 `json:"total_allocated_mb"`
	SystemMB         float64 `json:"system_mb"`
	HeapObjects      uint64  `json:"heap_objects"`
	StackInUseMB     float64 `json:"stack_in_use_mb"`
	GoroutineCount   int     `json:"goroutine_count"`

	// Advanced memory stats
	HeapInUseMB    float64 `json:"heap_in_use_mb"`
	HeapReleasedMB float64 `json:"heap_released_mb"`
	HeapIdleMB     float64 `json:"heap_idle_mb"`
	HeapSysMB      float64 `json:"heap_sys_mb"`
	HeapAllocMB    float64 `json:"heap_alloc_mb"`
	HeapNextGC     uint64  `json:"heap_next_gc"`
	HeapGoalMB     float64 `json:"heap_goal_mb"`

	// Memory pressure indicators
	MemoryPressureLevel string  `json:"memory_pressure_level"` // "low", "medium", "high", "critical"
	MemoryPressureScore float64 `json:"memory_pressure_score"`
	MemoryGrowthRate    float64 `json:"memory_growth_rate_mb_per_sec"`

	// GC statistics
	NumGC          uint32    `json:"num_gc"`
	PauseTotalMs   float64   `json:"pause_total_ms"`
	PauseAverageMs float64   `json:"pause_average_ms"`
	PauseMaxMs     float64   `json:"pause_max_ms"`
	LastGC         time.Time `json:"last_gc"`
	NextGC         time.Time `json:"next_gc"`
	GCForced       uint32    `json:"gc_forced"`

	// Memory allocation patterns
	AllocationRateMBPerSec    float64 `json:"allocation_rate_mb_per_sec"`
	DeallocationRateMBPerSec  float64 `json:"deallocation_rate_mb_per_sec"`
	NetAllocationRateMBPerSec float64 `json:"net_allocation_rate_mb_per_sec"`

	// Memory efficiency metrics
	MemoryEfficiencyScore float64 `json:"memory_efficiency_score"`
	FragmentationLevel    float64 `json:"fragmentation_level"`
	WasteRatio            float64 `json:"waste_ratio"`

	// System context
	CPUUsagePercent float64 `json:"cpu_usage_percent"`
	LoadAverage     float64 `json:"load_average"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GCStats represents garbage collection statistics
type GCStats struct {
	Timestamp        time.Time          `json:"timestamp"`
	NumGC            uint32             `json:"num_gc"`
	PauseTotalMs     float64            `json:"pause_total_ms"`
	PauseAverageMs   float64            `json:"pause_average_ms"`
	PauseMaxMs       float64            `json:"pause_max_ms"`
	PauseMinMs       float64            `json:"pause_min_ms"`
	PausePercentiles map[string]float64 `json:"pause_percentiles"`
	GCForced         uint32             `json:"gc_forced"`
	GCTriggered      uint32             `json:"gc_triggered"`
	LastGC           time.Time          `json:"last_gc"`
	NextGC           time.Time          `json:"next_gc"`
	HeapGoalMB       float64            `json:"heap_goal_mb"`
	HeapLiveMB       float64            `json:"heap_live_mb"`
	HeapSysMB        float64            `json:"heap_sys_mb"`
	HeapIdleMB       float64            `json:"heap_idle_mb"`
	HeapInUseMB      float64            `json:"heap_in_use_mb"`
	HeapReleasedMB   float64            `json:"heap_released_mb"`
	HeapAllocMB      float64            `json:"heap_alloc_mb"`
	HeapObjects      uint64             `json:"heap_objects"`
	StackInUseMB     float64            `json:"stack_in_use_mb"`
	GoroutineCount   int                `json:"goroutine_count"`
}

// GCEvent represents a garbage collection event
type GCEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"` // "start", "end", "pause"
	PauseTimeMs float64                `json:"pause_time_ms"`
	HeapSizeMB  float64                `json:"heap_size_mb"`
	HeapGoalMB  float64                `json:"heap_goal_mb"`
	NumGC       uint32                 `json:"num_gc"`
	Forced      bool                   `json:"forced"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MemoryPressureEvent represents a memory pressure event
type MemoryPressureEvent struct {
	Timestamp          time.Time              `json:"timestamp"`
	PressureLevel      string                 `json:"pressure_level"`
	PressureScore      float64                `json:"pressure_score"`
	MemoryUsageMB      float64                `json:"memory_usage_mb"`
	MemoryThresholdMB  float64                `json:"memory_threshold_mb"`
	GrowthRateMBPerSec float64                `json:"growth_rate_mb_per_sec"`
	Recommendations    []string               `json:"recommendations"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// NewAdvancedMemoryMonitor creates a new advanced memory monitor
func NewAdvancedMemoryMonitor(logger *zap.Logger, config *MemoryMonitorConfig) *AdvancedMemoryMonitor {
	if config == nil {
		config = DefaultMemoryMonitorConfig()
	}

	monitor := &AdvancedMemoryMonitor{
		logger:            logger,
		config:            config,
		stats:             &AdvancedMemoryStats{},
		history:           make([]*AdvancedMemoryStats, 0),
		maxHistorySize:    config.MaxHistorySize,
		gcStats:           &GCStats{},
		gcHistory:         make([]*GCStats, 0),
		gcEventChannel:    make(chan *GCEvent, 100),
		pressureThreshold: config.PressureThreshold,
		pressureEvents:    make([]*MemoryPressureEvent, 0),
		stopCh:            make(chan struct{}),
		monitoringActive:  false,
	}

	// Start monitoring if enabled
	if config.Enabled {
		monitor.Start()
	}

	return monitor
}

// DefaultMemoryMonitorConfig returns default configuration
func DefaultMemoryMonitorConfig() *MemoryMonitorConfig {
	return &MemoryMonitorConfig{
		Enabled:                true,
		CollectionInterval:     5 * time.Second,
		MaxHistorySize:         1000,
		PressureThreshold:      80.0,                  // 80% memory usage
		GCThreshold:            10 * time.Millisecond, // 10ms GC pause
		MemoryLeakThreshold:    10.0,                  // 10MB/sec growth rate
		AlertingEnabled:        true,
		DetailedGCStats:        true,
		TrackMemoryAllocations: true,
	}
}

// Start starts the memory monitoring
func (amm *AdvancedMemoryMonitor) Start() {
	amm.mu.Lock()
	defer amm.mu.Unlock()

	if amm.monitoringActive {
		return
	}

	amm.monitoringActive = true

	// Start background monitoring
	go amm.monitoringLoop()

	// Start GC event monitoring if detailed stats are enabled
	if amm.config.DetailedGCStats {
		go amm.gcEventLoop()
	}

	amm.logger.Info("Advanced memory monitoring started",
		zap.Duration("collection_interval", amm.config.CollectionInterval),
		zap.Float64("pressure_threshold", amm.config.PressureThreshold),
		zap.Bool("detailed_gc_stats", amm.config.DetailedGCStats))
}

// Stop stops the memory monitoring
func (amm *AdvancedMemoryMonitor) Stop() {
	amm.mu.Lock()
	defer amm.mu.Unlock()

	if !amm.monitoringActive {
		return
	}

	amm.monitoringActive = false
	close(amm.stopCh)

	amm.logger.Info("Advanced memory monitoring stopped")
}

// monitoringLoop runs the main monitoring loop
func (amm *AdvancedMemoryMonitor) monitoringLoop() {
	ticker := time.NewTicker(amm.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			amm.collectMemoryStats()
		case <-amm.stopCh:
			return
		}
	}
}

// gcEventLoop monitors GC events
func (amm *AdvancedMemoryMonitor) gcEventLoop() {
	for {
		select {
		case event := <-amm.gcEventChannel:
			amm.processGCEvent(event)
		case <-amm.stopCh:
			return
		}
	}
}

// collectMemoryStats collects comprehensive memory statistics
func (amm *AdvancedMemoryMonitor) collectMemoryStats() {
	amm.mu.Lock()
	defer amm.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	now := time.Now()

	// Create new stats
	stats := &AdvancedMemoryStats{
		Timestamp:        now,
		AllocatedMB:      float64(m.Alloc) / (1024 * 1024),
		TotalAllocatedMB: float64(m.TotalAlloc) / (1024 * 1024),
		SystemMB:         float64(m.Sys) / (1024 * 1024),
		HeapObjects:      m.HeapObjects,
		StackInUseMB:     float64(m.StackInuse) / (1024 * 1024),
		GoroutineCount:   runtime.NumGoroutine(),

		// Advanced heap stats
		HeapInUseMB:    float64(m.HeapInuse) / (1024 * 1024),
		HeapReleasedMB: float64(m.HeapReleased) / (1024 * 1024),
		HeapIdleMB:     float64(m.HeapIdle) / (1024 * 1024),
		HeapSysMB:      float64(m.HeapSys) / (1024 * 1024),
		HeapAllocMB:    float64(m.HeapAlloc) / (1024 * 1024),
		HeapNextGC:     m.NextGC,
		HeapGoalMB:     float64(m.NextGC) / (1024 * 1024),

		// GC stats
		NumGC:          m.NumGC,
		PauseTotalMs:   float64(m.PauseTotalNs) / 1000000,
		PauseAverageMs: float64(m.PauseTotalNs) / float64(m.NumGC) / 1000000,
		PauseMaxMs:     float64(m.PauseNs[(m.NumGC+255)%256]) / 1000000,
		LastGC:         time.Unix(0, int64(m.LastGC)),
		NextGC:         time.Unix(0, int64(m.NextGC)),
		GCForced:       m.NumForcedGC,

		Metadata: make(map[string]interface{}),
	}

	// Calculate memory pressure
	amm.calculateMemoryPressure(stats)

	// Calculate memory growth rate
	amm.calculateMemoryGrowthRate(stats)

	// Calculate memory efficiency metrics
	amm.calculateMemoryEfficiency(stats)

	// Update current stats
	amm.stats = stats

	// Add to history
	amm.history = append(amm.history, stats)
	if len(amm.history) > amm.maxHistorySize {
		amm.history = amm.history[1:]
	}

	// Check for memory pressure events
	amm.checkMemoryPressure(stats)

	// Log significant events
	amm.logSignificantEvents(stats)
}

// calculateMemoryPressure calculates memory pressure level and score
func (amm *AdvancedMemoryMonitor) calculateMemoryPressure(stats *AdvancedMemoryStats) {
	// Calculate pressure based on multiple factors
	pressureScore := 0.0

	// Factor 1: Memory usage percentage
	memoryUsagePercent := (stats.AllocatedMB / stats.SystemMB) * 100
	if memoryUsagePercent > 80 {
		pressureScore += 40
	} else if memoryUsagePercent > 60 {
		pressureScore += 20
	} else if memoryUsagePercent > 40 {
		pressureScore += 10
	}

	// Factor 2: GC frequency and pause time
	if stats.PauseAverageMs > float64(amm.config.GCThreshold.Milliseconds()) {
		pressureScore += 30
	} else if stats.PauseAverageMs > float64(amm.config.GCThreshold.Milliseconds())/2 {
		pressureScore += 15
	}

	// Factor 3: Memory growth rate
	if stats.MemoryGrowthRate > amm.config.MemoryLeakThreshold {
		pressureScore += 20
	} else if stats.MemoryGrowthRate > amm.config.MemoryLeakThreshold/2 {
		pressureScore += 10
	}

	// Factor 4: Heap fragmentation
	if stats.FragmentationLevel > 0.5 {
		pressureScore += 10
	}

	stats.MemoryPressureScore = pressureScore

	// Determine pressure level
	if pressureScore >= 80 {
		stats.MemoryPressureLevel = "critical"
	} else if pressureScore >= 60 {
		stats.MemoryPressureLevel = "high"
	} else if pressureScore >= 30 {
		stats.MemoryPressureLevel = "medium"
	} else {
		stats.MemoryPressureLevel = "low"
	}
}

// calculateMemoryGrowthRate calculates memory growth rate
func (amm *AdvancedMemoryMonitor) calculateMemoryGrowthRate(stats *AdvancedMemoryStats) {
	if len(amm.history) < 2 {
		stats.MemoryGrowthRate = 0.0
		return
	}

	// Calculate growth rate over the last few samples
	recentHistory := amm.history
	if len(recentHistory) > 10 {
		recentHistory = recentHistory[len(recentHistory)-10:]
	}

	if len(recentHistory) < 2 {
		stats.MemoryGrowthRate = 0.0
		return
	}

	first := recentHistory[0]
	last := recentHistory[len(recentHistory)-1]

	timeDiff := last.Timestamp.Sub(first.Timestamp).Seconds()
	memoryDiff := last.AllocatedMB - first.AllocatedMB

	if timeDiff > 0 {
		stats.MemoryGrowthRate = memoryDiff / timeDiff
	} else {
		stats.MemoryGrowthRate = 0.0
	}
}

// calculateMemoryEfficiency calculates memory efficiency metrics
func (amm *AdvancedMemoryMonitor) calculateMemoryEfficiency(stats *AdvancedMemoryStats) {
	// Calculate fragmentation level
	if stats.HeapSysMB > 0 {
		stats.FragmentationLevel = (stats.HeapSysMB - stats.HeapInUseMB) / stats.HeapSysMB
	} else {
		stats.FragmentationLevel = 0.0
	}

	// Calculate waste ratio
	if stats.HeapSysMB > 0 {
		stats.WasteRatio = (stats.HeapIdleMB + stats.HeapReleasedMB) / stats.HeapSysMB
	} else {
		stats.WasteRatio = 0.0
	}

	// Calculate memory efficiency score (0-100)
	efficiencyScore := 100.0

	// Deduct points for fragmentation
	efficiencyScore -= stats.FragmentationLevel * 30

	// Deduct points for waste
	efficiencyScore -= stats.WasteRatio * 20

	// Deduct points for high GC overhead
	if stats.PauseAverageMs > 5 {
		efficiencyScore -= (stats.PauseAverageMs - 5) * 2
	}

	// Ensure score is between 0 and 100
	if efficiencyScore < 0 {
		efficiencyScore = 0
	}
	if efficiencyScore > 100 {
		efficiencyScore = 100
	}

	stats.MemoryEfficiencyScore = efficiencyScore
}

// checkMemoryPressure checks for memory pressure events
func (amm *AdvancedMemoryMonitor) checkMemoryPressure(stats *AdvancedMemoryStats) {
	if !amm.config.AlertingEnabled {
		return
	}

	// Check if we should trigger a pressure event
	shouldTrigger := false
	var recommendations []string

	if stats.MemoryPressureLevel == "critical" {
		shouldTrigger = true
		recommendations = append(recommendations, "Immediate memory optimization required")
		recommendations = append(recommendations, "Consider reducing memory allocation")
		recommendations = append(recommendations, "Check for memory leaks")
	} else if stats.MemoryPressureLevel == "high" {
		shouldTrigger = true
		recommendations = append(recommendations, "Monitor memory usage closely")
		recommendations = append(recommendations, "Consider optimizing memory allocation patterns")
	}

	if shouldTrigger {
		event := &MemoryPressureEvent{
			Timestamp:          stats.Timestamp,
			PressureLevel:      stats.MemoryPressureLevel,
			PressureScore:      stats.MemoryPressureScore,
			MemoryUsageMB:      stats.AllocatedMB,
			MemoryThresholdMB:  stats.SystemMB * (amm.config.PressureThreshold / 100),
			GrowthRateMBPerSec: stats.MemoryGrowthRate,
			Recommendations:    recommendations,
			Metadata:           make(map[string]interface{}),
		}

		amm.pressureEvents = append(amm.pressureEvents, event)

		// Keep only recent events
		if len(amm.pressureEvents) > 100 {
			amm.pressureEvents = amm.pressureEvents[1:]
		}

		amm.logger.Warn("Memory pressure event detected",
			zap.String("pressure_level", event.PressureLevel),
			zap.Float64("pressure_score", event.PressureScore),
			zap.Float64("memory_usage_mb", event.MemoryUsageMB),
			zap.Float64("growth_rate_mb_per_sec", event.GrowthRateMBPerSec))
	}
}

// logSignificantEvents logs significant memory events
func (amm *AdvancedMemoryMonitor) logSignificantEvents(stats *AdvancedMemoryStats) {
	// Log high memory usage
	if stats.AllocatedMB > 1000 { // 1GB
		amm.logger.Warn("High memory usage detected",
			zap.Float64("allocated_mb", stats.AllocatedMB),
			zap.Float64("system_mb", stats.SystemMB),
			zap.String("pressure_level", stats.MemoryPressureLevel))
	}

	// Log high GC pause times
	if stats.PauseAverageMs > float64(amm.config.GCThreshold.Milliseconds()) {
		amm.logger.Warn("High GC pause time detected",
			zap.Float64("pause_average_ms", stats.PauseAverageMs),
			zap.Float64("pause_max_ms", stats.PauseMaxMs),
			zap.Uint32("num_gc", stats.NumGC))
	}

	// Log memory growth
	if stats.MemoryGrowthRate > amm.config.MemoryLeakThreshold {
		amm.logger.Warn("High memory growth rate detected",
			zap.Float64("growth_rate_mb_per_sec", stats.MemoryGrowthRate),
			zap.Float64("allocated_mb", stats.AllocatedMB))
	}
}

// processGCEvent processes a garbage collection event
func (amm *AdvancedMemoryMonitor) processGCEvent(event *GCEvent) {
	amm.mu.Lock()
	defer amm.mu.Unlock()

	// Update GC stats
	amm.gcStats = &GCStats{
		Timestamp:      event.Timestamp,
		NumGC:          event.NumGC,
		PauseTotalMs:   event.PauseTimeMs,
		PauseAverageMs: event.PauseTimeMs,
		PauseMaxMs:     event.PauseTimeMs,
		PauseMinMs:     event.PauseTimeMs,
		LastGC:         event.Timestamp,
		NextGC:         time.Unix(0, int64(event.HeapGoalMB*1024*1024)),
		HeapGoalMB:     event.HeapGoalMB,
		HeapLiveMB:     event.HeapSizeMB,
		HeapAllocMB:    event.HeapSizeMB,
		HeapObjects:    0, // Would need additional tracking
		GoroutineCount: runtime.NumGoroutine(),
	}

	// Add to history
	amm.gcHistory = append(amm.gcHistory, amm.gcStats)
	if len(amm.gcHistory) > amm.maxHistorySize {
		amm.gcHistory = amm.gcHistory[1:]
	}

	// Log significant GC events
	if event.PauseTimeMs > float64(amm.config.GCThreshold.Milliseconds()) {
		amm.logger.Warn("Long GC pause detected",
			zap.Float64("pause_time_ms", event.PauseTimeMs),
			zap.Float64("heap_size_mb", event.HeapSizeMB),
			zap.Bool("forced", event.Forced))
	}
}

// GetCurrentStats returns current memory statistics
func (amm *AdvancedMemoryMonitor) GetCurrentStats() *AdvancedMemoryStats {
	amm.mu.RLock()
	defer amm.mu.RUnlock()

	return amm.stats
}

// GetMemoryHistory returns memory statistics history
func (amm *AdvancedMemoryMonitor) GetMemoryHistory(limit int) []*AdvancedMemoryStats {
	amm.mu.RLock()
	defer amm.mu.RUnlock()

	if limit <= 0 || limit > len(amm.history) {
		limit = len(amm.history)
	}

	start := len(amm.history) - limit
	if start < 0 {
		start = 0
	}

	result := make([]*AdvancedMemoryStats, limit)
	copy(result, amm.history[start:])
	return result
}

// GetGCHistory returns GC statistics history
func (amm *AdvancedMemoryMonitor) GetGCHistory(limit int) []*GCStats {
	amm.mu.RLock()
	defer amm.mu.RUnlock()

	if limit <= 0 || limit > len(amm.gcHistory) {
		limit = len(amm.gcHistory)
	}

	start := len(amm.gcHistory) - limit
	if start < 0 {
		start = 0
	}

	result := make([]*GCStats, limit)
	copy(result, amm.gcHistory[start:])
	return result
}

// GetMemoryPressureEvents returns memory pressure events
func (amm *AdvancedMemoryMonitor) GetMemoryPressureEvents(limit int) []*MemoryPressureEvent {
	amm.mu.RLock()
	defer amm.mu.RUnlock()

	if limit <= 0 || limit > len(amm.pressureEvents) {
		limit = len(amm.pressureEvents)
	}

	start := len(amm.pressureEvents) - limit
	if start < 0 {
		start = 0
	}

	result := make([]*MemoryPressureEvent, limit)
	copy(result, amm.pressureEvents[start:])
	return result
}

// GetMemorySummary returns a summary of memory statistics
func (amm *AdvancedMemoryMonitor) GetMemorySummary() map[string]interface{} {
	amm.mu.RLock()
	defer amm.mu.RUnlock()

	if amm.stats == nil {
		return map[string]interface{}{
			"error": "no memory statistics available",
		}
	}

	summary := map[string]interface{}{
		"current": map[string]interface{}{
			"allocated_mb":            amm.stats.AllocatedMB,
			"system_mb":               amm.stats.SystemMB,
			"heap_objects":            amm.stats.HeapObjects,
			"goroutine_count":         amm.stats.GoroutineCount,
			"memory_pressure_level":   amm.stats.MemoryPressureLevel,
			"memory_pressure_score":   amm.stats.MemoryPressureScore,
			"memory_efficiency_score": amm.stats.MemoryEfficiencyScore,
			"fragmentation_level":     amm.stats.FragmentationLevel,
			"waste_ratio":             amm.stats.WasteRatio,
		},
		"gc": map[string]interface{}{
			"num_gc":           amm.stats.NumGC,
			"pause_average_ms": amm.stats.PauseAverageMs,
			"pause_max_ms":     amm.stats.PauseMaxMs,
			"pause_total_ms":   amm.stats.PauseTotalMs,
			"gc_forced":        amm.stats.GCForced,
		},
		"trends": map[string]interface{}{
			"memory_growth_rate_mb_per_sec": amm.stats.MemoryGrowthRate,
			"allocation_rate_mb_per_sec":    amm.stats.AllocationRateMBPerSec,
		},
		"history": map[string]interface{}{
			"memory_stats_count": len(amm.history),
			"gc_stats_count":     len(amm.gcHistory),
			"pressure_events":    len(amm.pressureEvents),
		},
	}

	return summary
}

// ForceGC forces a garbage collection and records the event
func (amm *AdvancedMemoryMonitor) ForceGC() {
	before := time.Now()
	var beforeStats runtime.MemStats
	runtime.ReadMemStats(&beforeStats)

	runtime.GC()

	after := time.Now()
	var afterStats runtime.MemStats
	runtime.ReadMemStats(&afterStats)

	pauseTime := after.Sub(before).Seconds() * 1000 // Convert to milliseconds

	event := &GCEvent{
		Timestamp:   after,
		EventType:   "forced",
		PauseTimeMs: pauseTime,
		HeapSizeMB:  float64(afterStats.HeapAlloc) / (1024 * 1024),
		HeapGoalMB:  float64(afterStats.NextGC) / (1024 * 1024),
		NumGC:       afterStats.NumGC,
		Forced:      true,
		Metadata: map[string]interface{}{
			"before_heap_mb": float64(beforeStats.HeapAlloc) / (1024 * 1024),
			"after_heap_mb":  float64(afterStats.HeapAlloc) / (1024 * 1024),
			"freed_mb":       (float64(beforeStats.HeapAlloc) - float64(afterStats.HeapAlloc)) / (1024 * 1024),
		},
	}

	// Send event to channel
	select {
	case amm.gcEventChannel <- event:
	default:
		// Channel full, process directly
		amm.processGCEvent(event)
	}

	amm.logger.Info("Forced garbage collection completed",
		zap.Float64("pause_time_ms", pauseTime),
		zap.Float64("heap_size_mb", event.HeapSizeMB),
		zap.Float64("freed_mb", event.Metadata["freed_mb"].(float64)))
}
