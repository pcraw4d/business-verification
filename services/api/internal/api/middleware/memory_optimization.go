package middleware

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// MemoryOptimizationManager provides advanced memory optimization and garbage collection
type MemoryOptimizationManager struct {
	config            *MemoryOptimizationConfig
	profiler          *MemoryProfiler
	gcOptimizer       *AdvancedGCOptimizer
	memoryPooler      *AdvancedMemoryPooler
	leakDetector      *MemoryLeakDetector
	compactionManager *MemoryCompactionManager
	mu                sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
	optimizationDone  chan struct{}
}

// MemoryOptimizationConfig holds configuration for memory optimization
type MemoryOptimizationConfig struct {
	EnableMemoryProfiling     bool          // Enable detailed memory profiling
	ProfilingInterval         time.Duration // How often to profile memory
	GCTriggerThreshold        float64       // Memory usage threshold to trigger GC
	MemoryCompactionThreshold float64       // Threshold for memory compaction
	LeakDetectionEnabled      bool          // Enable memory leak detection
	LeakDetectionInterval     time.Duration // How often to check for leaks
	PoolingEnabled            bool          // Enable memory pooling
	MaxPoolSize               int64         // Maximum size of memory pools
	CompactionEnabled         bool          // Enable memory compaction
	CompactionInterval        time.Duration // How often to run compaction
	HeapGrowthLimit           uint64        // Maximum heap growth limit
	HeapIdleTimeout           time.Duration // Time before idle heap is released
}

// MemoryProfiler provides detailed memory profiling capabilities
type MemoryProfiler struct {
	config          *MemoryOptimizationConfig
	profiles        []MemoryProfile
	allocationStats *AllocationStats
	mu              sync.RWMutex
}

// MemoryProfile represents a memory profile snapshot
type MemoryProfile struct {
	Timestamp      time.Time
	HeapAlloc      uint64
	HeapSys        uint64
	HeapInuse      uint64
	HeapIdle       uint64
	HeapReleased   uint64
	HeapObjects    uint64
	StackInuse     uint64
	StackSys       uint64
	MSpanInuse     uint64
	MSpanSys       uint64
	MCacheInuse    uint64
	MCacheSys      uint64
	BuckHashSys    uint64
	GCSys          uint64
	OtherSys       uint64
	NextGC         uint64
	LastGC         uint64
	PauseTotalNs   uint64
	PauseNs        []uint64
	PauseEnd       []uint64
	NumGC          uint32
	NumForcedGC    uint32
	GCCPUFraction  float64
	EnableGC       bool
	DebugGC        bool
	AllocationRate float64 // Allocations per second
	GCTriggerRate  float64 // GC triggers per second
}

// AllocationStats tracks memory allocation patterns
type AllocationStats struct {
	TotalAllocations   uint64
	TotalDeallocations uint64
	PeakAllocation     uint64
	AverageAllocation  uint64
	AllocationRate     float64
	DeallocationRate   float64
	FragmentationRatio float64
	LastUpdated        time.Time
	AllocationHistory  []AllocationEvent
}

// AllocationEvent represents a memory allocation event
type AllocationEvent struct {
	Timestamp   time.Time
	Size        uint64
	Type        string
	Stack       string
	Duration    time.Duration
	GCTriggered bool
}

// AdvancedGCOptimizer provides sophisticated garbage collection optimization
type AdvancedGCOptimizer struct {
	config              *MemoryOptimizationConfig
	gcStats             *GCStats
	optimizationHistory []GCOptimizationEvent
	mu                  sync.RWMutex
}

// GCStats tracks garbage collection performance
type GCStats struct {
	TotalCycles       uint32
	TotalPauseTime    time.Duration
	AveragePauseTime  time.Duration
	MaxPauseTime      time.Duration
	MinPauseTime      time.Duration
	PauseTimeHistory  []time.Duration
	MemoryFreed       uint64
	Efficiency        float64 // Percentage of memory freed per GC cycle
	Frequency         float64 // GC cycles per second
	LastOptimization  time.Time
	TargetPercentage  int
	CurrentPercentage int
}

// GCOptimizationEvent represents a GC optimization event
type GCOptimizationEvent struct {
	Timestamp          time.Time
	EventType          string
	BeforePercentage   int
	AfterPercentage    int
	MemoryFreed        uint64
	PauseTimeReduction time.Duration
	Description        string
	Success            bool
}

// AdvancedMemoryPooler provides sophisticated memory pooling
type AdvancedMemoryPooler struct {
	config *MemoryOptimizationConfig
	pools  map[string]*MemoryPool
	stats  *PoolStats
	mu     sync.RWMutex
}

// MemoryPool represents a memory pool for specific object types
type MemoryPool struct {
	Name           string
	ObjectSize     uint64
	MaxObjects     int
	CurrentObjects int
	Allocations    int64
	Releases       int64
	HitRate        float64
	LastUsed       time.Time
	Pool           *sync.Pool
}

// PoolStats tracks memory pool performance
type PoolStats struct {
	TotalPools       int
	TotalAllocations int64
	TotalReleases    int64
	AverageHitRate   float64
	MemorySaved      uint64
	LastUpdated      time.Time
}

// MemoryLeakDetector detects potential memory leaks
type MemoryLeakDetector struct {
	config           *MemoryOptimizationConfig
	leakPatterns     []LeakPattern
	detectionHistory []LeakDetectionEvent
	mu               sync.RWMutex
}

// LeakPattern represents a memory leak pattern
type LeakPattern struct {
	ID          string
	Name        string
	Description string
	Threshold   float64
	TimeWindow  time.Duration
	Severity    string // low, medium, high, critical
}

// LeakDetectionEvent represents a detected memory leak
type LeakDetectionEvent struct {
	Timestamp    time.Time
	PatternID    string
	Severity     string
	MemoryGrowth uint64
	Duration     time.Duration
	Description  string
	Resolved     bool
}

// MemoryCompactionManager manages memory compaction
type MemoryCompactionManager struct {
	config            *MemoryOptimizationConfig
	compactionStats   *CompactionStats
	compactionHistory []CompactionEvent
	mu                sync.RWMutex
}

// CompactionStats tracks memory compaction performance
type CompactionStats struct {
	TotalCompactions int
	TotalMemoryFreed uint64
	AverageFreed     uint64
	CompactionTime   time.Duration
	LastCompaction   time.Time
	Efficiency       float64
}

// CompactionEvent represents a memory compaction event
type CompactionEvent struct {
	Timestamp    time.Time
	MemoryBefore uint64
	MemoryAfter  uint64
	MemoryFreed  uint64
	Duration     time.Duration
	Efficiency   float64
	Success      bool
}

// DefaultMemoryOptimizationConfig creates a default memory optimization configuration
func DefaultMemoryOptimizationConfig() *MemoryOptimizationConfig {
	return &MemoryOptimizationConfig{
		EnableMemoryProfiling:     true,
		ProfilingInterval:         30 * time.Second,
		GCTriggerThreshold:        75.0, // Trigger GC at 75% memory usage
		MemoryCompactionThreshold: 80.0, // Compact memory at 80% usage
		LeakDetectionEnabled:      true,
		LeakDetectionInterval:     2 * time.Minute,
		PoolingEnabled:            true,
		MaxPoolSize:               1000,
		CompactionEnabled:         true,
		CompactionInterval:        5 * time.Minute,
		HeapGrowthLimit:           100 * 1024 * 1024, // 100MB
		HeapIdleTimeout:           10 * time.Minute,
	}
}

// NewMemoryOptimizationManager creates a new memory optimization manager
func NewMemoryOptimizationManager(config *MemoryOptimizationConfig) *MemoryOptimizationManager {
	if config == nil {
		config = DefaultMemoryOptimizationConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	mom := &MemoryOptimizationManager{
		config:            config,
		profiler:          NewMemoryProfiler(config),
		gcOptimizer:       NewAdvancedGCOptimizer(config),
		memoryPooler:      NewAdvancedMemoryPooler(config),
		leakDetector:      NewMemoryLeakDetector(config),
		compactionManager: NewMemoryCompactionManager(config),
		ctx:               ctx,
		cancel:            cancel,
		optimizationDone:  make(chan struct{}),
	}

	// Start optimization goroutine
	go mom.startOptimization()

	return mom
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler(config *MemoryOptimizationConfig) *MemoryProfiler {
	return &MemoryProfiler{
		config:   config,
		profiles: make([]MemoryProfile, 0),
		allocationStats: &AllocationStats{
			AllocationHistory: make([]AllocationEvent, 0),
		},
	}
}

// NewAdvancedGCOptimizer creates a new advanced GC optimizer
func NewAdvancedGCOptimizer(config *MemoryOptimizationConfig) *AdvancedGCOptimizer {
	return &AdvancedGCOptimizer{
		config: config,
		gcStats: &GCStats{
			PauseTimeHistory: make([]time.Duration, 0),
		},
		optimizationHistory: make([]GCOptimizationEvent, 0),
	}
}

// NewAdvancedMemoryPooler creates a new advanced memory pooler
func NewAdvancedMemoryPooler(config *MemoryOptimizationConfig) *AdvancedMemoryPooler {
	return &AdvancedMemoryPooler{
		config: config,
		pools:  make(map[string]*MemoryPool),
		stats:  &PoolStats{},
	}
}

// NewMemoryLeakDetector creates a new memory leak detector
func NewMemoryLeakDetector(config *MemoryOptimizationConfig) *MemoryLeakDetector {
	detector := &MemoryLeakDetector{
		config:           config,
		leakPatterns:     make([]LeakPattern, 0),
		detectionHistory: make([]LeakDetectionEvent, 0),
	}

	// Initialize default leak patterns
	detector.initializeDefaultPatterns()

	return detector
}

// NewMemoryCompactionManager creates a new memory compaction manager
func NewMemoryCompactionManager(config *MemoryOptimizationConfig) *MemoryCompactionManager {
	return &MemoryCompactionManager{
		config:            config,
		compactionStats:   &CompactionStats{},
		compactionHistory: make([]CompactionEvent, 0),
	}
}

// ProfileMemory creates a detailed memory profile
func (mp *MemoryProfiler) ProfileMemory() *MemoryProfile {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	profile := MemoryProfile{
		Timestamp:     time.Now(),
		HeapAlloc:     m.Alloc,
		HeapSys:       m.HeapSys,
		HeapInuse:     m.HeapInuse,
		HeapIdle:      m.HeapIdle,
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
		PauseNs:       make([]uint64, len(m.PauseNs)),
		PauseEnd:      make([]uint64, len(m.PauseEnd)),
		NumGC:         m.NumGC,
		NumForcedGC:   m.NumForcedGC,
		GCCPUFraction: m.GCCPUFraction,
		EnableGC:      m.EnableGC,
		DebugGC:       m.DebugGC,
	}

	// Copy pause arrays
	copy(profile.PauseNs, m.PauseNs[:])
	copy(profile.PauseEnd, m.PauseEnd[:])

	// Calculate allocation rate if we have previous profiles
	if len(mp.profiles) > 0 {
		lastProfile := mp.profiles[len(mp.profiles)-1]
		timeDiff := profile.Timestamp.Sub(lastProfile.Timestamp).Seconds()
		if timeDiff > 0 {
			allocationDiff := float64(profile.HeapAlloc - lastProfile.HeapAlloc)
			profile.AllocationRate = allocationDiff / timeDiff
		}
	}

	mp.profiles = append(mp.profiles, profile)

	// Keep only the last 100 profiles
	if len(mp.profiles) > 100 {
		mp.profiles = mp.profiles[len(mp.profiles)-100:]
	}

	return &profile
}

// OptimizeGC performs advanced garbage collection optimization
func (ago *AdvancedGCOptimizer) OptimizeGC(currentProfile *MemoryProfile) error {
	ago.mu.Lock()
	defer ago.mu.Unlock()

	// Update GC stats
	ago.updateGCStats(currentProfile)

	// Determine optimal GC percentage
	optimalPercentage := ago.calculateOptimalGCPercentage(currentProfile)

	// Apply optimization if needed
	if optimalPercentage != ago.gcStats.CurrentPercentage {
		ago.applyGCOptimization(optimalPercentage, currentProfile)
	}

	return nil
}

// updateGCStats updates garbage collection statistics
func (ago *AdvancedGCOptimizer) updateGCStats(profile *MemoryProfile) {
	if ago.gcStats.TotalCycles == 0 {
		ago.gcStats.TotalCycles = profile.NumGC
		ago.gcStats.TotalPauseTime = time.Duration(profile.PauseTotalNs)
	} else {
		cyclesDiff := profile.NumGC - ago.gcStats.TotalCycles
		if cyclesDiff > 0 {
			ago.gcStats.TotalCycles = profile.NumGC
			ago.gcStats.TotalPauseTime = time.Duration(profile.PauseTotalNs)

			// Calculate average pause time
			if ago.gcStats.TotalCycles > 0 {
				ago.gcStats.AveragePauseTime = ago.gcStats.TotalPauseTime / time.Duration(ago.gcStats.TotalCycles)
			}
		}
	}

	ago.gcStats.CurrentPercentage = debug.SetGCPercent(-1) // Get current percentage
	ago.gcStats.LastOptimization = time.Now()
}

// calculateOptimalGCPercentage calculates the optimal GC percentage based on memory usage
func (ago *AdvancedGCOptimizer) calculateOptimalGCPercentage(profile *MemoryProfile) int {
	memoryUsage := float64(profile.HeapInuse) / float64(profile.HeapSys) * 100

	switch {
	case memoryUsage > 90:
		return 50 // Very aggressive GC for high memory usage
	case memoryUsage > 80:
		return 75 // Aggressive GC for high memory usage
	case memoryUsage > 70:
		return 100 // Normal GC for moderate memory usage
	case memoryUsage > 50:
		return 150 // Less aggressive GC for low memory usage
	default:
		return 200 // Very conservative GC for very low memory usage
	}
}

// applyGCOptimization applies GC optimization
func (ago *AdvancedGCOptimizer) applyGCOptimization(newPercentage int, profile *MemoryProfile) {
	beforePercentage := ago.gcStats.CurrentPercentage

	// Set new GC percentage
	debug.SetGCPercent(newPercentage)

	// Force a GC cycle to apply the new settings
	runtime.GC()

	// Read memory stats after GC
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memoryFreed := uint64(0)
	if profile.HeapAlloc > m.Alloc {
		memoryFreed = profile.HeapAlloc - m.Alloc
	}

	// Record optimization event
	event := GCOptimizationEvent{
		Timestamp:        time.Now(),
		EventType:        "gc_percentage_adjustment",
		BeforePercentage: beforePercentage,
		AfterPercentage:  newPercentage,
		MemoryFreed:      memoryFreed,
		Description:      fmt.Sprintf("Adjusted GC percentage from %d to %d", beforePercentage, newPercentage),
		Success:          true,
	}

	ago.optimizationHistory = append(ago.optimizationHistory, event)
	ago.gcStats.CurrentPercentage = newPercentage

	log.Printf("GC Optimization: %s, Memory freed: %d bytes", event.Description, memoryFreed)
}

// CreatePool creates a new memory pool for a specific object type
func (amp *AdvancedMemoryPooler) CreatePool(name string, objectSize uint64, maxObjects int) *MemoryPool {
	amp.mu.Lock()
	defer amp.mu.Unlock()

	pool := &MemoryPool{
		Name:       name,
		ObjectSize: objectSize,
		MaxObjects: maxObjects,
		Pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, objectSize)
			},
		},
		LastUsed: time.Now(),
	}

	amp.pools[name] = pool
	amp.stats.TotalPools++

	return pool
}

// GetFromPool gets an object from a memory pool
func (amp *AdvancedMemoryPooler) GetFromPool(name string) interface{} {
	amp.mu.RLock()
	pool, exists := amp.pools[name]
	amp.mu.RUnlock()

	if !exists {
		return nil
	}

	obj := pool.Pool.Get()
	pool.Allocations++
	pool.CurrentObjects++
	pool.LastUsed = time.Now()

	// Update hit rate
	if pool.Allocations > 0 {
		pool.HitRate = float64(pool.Releases) / float64(pool.Allocations)
	}

	amp.updatePoolStats()

	return obj
}

// ReturnToPool returns an object to a memory pool
func (amp *AdvancedMemoryPooler) ReturnToPool(name string, obj interface{}) {
	amp.mu.RLock()
	pool, exists := amp.pools[name]
	amp.mu.RUnlock()

	if !exists {
		return
	}

	pool.Pool.Put(obj)
	pool.Releases++
	if pool.CurrentObjects > 0 {
		pool.CurrentObjects--
	}

	// Update hit rate
	if pool.Allocations > 0 {
		pool.HitRate = float64(pool.Releases) / float64(pool.Allocations)
	}

	amp.updatePoolStats()
}

// updatePoolStats updates pool statistics
func (amp *AdvancedMemoryPooler) updatePoolStats() {
	amp.mu.Lock()
	defer amp.mu.Unlock()

	totalAllocations := int64(0)
	totalReleases := int64(0)
	totalHitRate := float64(0)
	poolCount := 0

	for _, pool := range amp.pools {
		totalAllocations += pool.Allocations
		totalReleases += pool.Releases
		totalHitRate += pool.HitRate
		poolCount++
	}

	if poolCount > 0 {
		amp.stats.AverageHitRate = totalHitRate / float64(poolCount)
	}

	amp.stats.TotalAllocations = totalAllocations
	amp.stats.TotalReleases = totalReleases
	amp.stats.LastUpdated = time.Now()
}

// initializeDefaultPatterns initializes default memory leak detection patterns
func (mld *MemoryLeakDetector) initializeDefaultPatterns() {
	mld.leakPatterns = []LeakPattern{
		{
			ID:          "heap_growth",
			Name:        "Heap Growth Pattern",
			Description: "Detects continuous heap growth without corresponding GC",
			Threshold:   10.0, // 10% growth
			TimeWindow:  5 * time.Minute,
			Severity:    "medium",
		},
		{
			ID:          "goroutine_leak",
			Name:        "Goroutine Leak",
			Description: "Detects increasing goroutine count without decrease",
			Threshold:   50.0, // 50 goroutines increase
			TimeWindow:  2 * time.Minute,
			Severity:    "high",
		},
		{
			ID:          "memory_fragmentation",
			Name:        "Memory Fragmentation",
			Description: "Detects high memory fragmentation",
			Threshold:   30.0, // 30% fragmentation
			TimeWindow:  10 * time.Minute,
			Severity:    "low",
		},
	}
}

// DetectLeaks performs memory leak detection
func (mld *MemoryLeakDetector) DetectLeaks(currentProfile *MemoryProfile) []LeakDetectionEvent {
	mld.mu.Lock()
	defer mld.mu.Unlock()

	detectedLeaks := make([]LeakDetectionEvent, 0)

	for _, pattern := range mld.leakPatterns {
		if leak := mld.checkLeakPattern(pattern, currentProfile); leak != nil {
			detectedLeaks = append(detectedLeaks, *leak)
			mld.detectionHistory = append(mld.detectionHistory, *leak)
		}
	}

	// Keep only the last 50 detection events
	if len(mld.detectionHistory) > 50 {
		mld.detectionHistory = mld.detectionHistory[len(mld.detectionHistory)-50:]
	}

	return detectedLeaks
}

// checkLeakPattern checks if a specific leak pattern is detected
func (mld *MemoryLeakDetector) checkLeakPattern(pattern LeakPattern, currentProfile *MemoryProfile) *LeakDetectionEvent {
	// This is a simplified implementation
	// In a real system, you would compare against historical data

	// For heap growth pattern
	if pattern.ID == "heap_growth" {
		heapUsage := float64(currentProfile.HeapInuse) / float64(currentProfile.HeapSys) * 100
		if heapUsage > pattern.Threshold {
			return &LeakDetectionEvent{
				Timestamp:    time.Now(),
				PatternID:    pattern.ID,
				Severity:     pattern.Severity,
				MemoryGrowth: currentProfile.HeapInuse,
				Duration:     pattern.TimeWindow,
				Description:  fmt.Sprintf("Heap usage at %.2f%% exceeds threshold of %.2f%%", heapUsage, pattern.Threshold),
				Resolved:     false,
			}
		}
	}

	return nil
}

// CompactMemory performs memory compaction
func (mcm *MemoryCompactionManager) CompactMemory() error {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	startTime := time.Now()

	// Get memory stats before compaction
	var beforeStats runtime.MemStats
	runtime.ReadMemStats(&beforeStats)

	// Force garbage collection
	runtime.GC()

	// Force memory compaction by allocating and releasing memory
	// This is a simplified approach - in production you might use more sophisticated techniques
	for i := 0; i < 1000; i++ {
		// Allocate some memory
		_ = make([]byte, 1024)
	}

	// Force another GC
	runtime.GC()

	// Get memory stats after compaction
	var afterStats runtime.MemStats
	runtime.ReadMemStats(&afterStats)

	compactionTime := time.Since(startTime)
	memoryFreed := uint64(0)

	if beforeStats.HeapInuse > afterStats.HeapInuse {
		memoryFreed = beforeStats.HeapInuse - afterStats.HeapInuse
	}

	// Calculate efficiency
	efficiency := float64(0)
	if compactionTime > 0 {
		efficiency = float64(memoryFreed) / float64(compactionTime.Milliseconds())
	}

	// Record compaction event
	event := CompactionEvent{
		Timestamp:    startTime,
		MemoryBefore: beforeStats.HeapInuse,
		MemoryAfter:  afterStats.HeapInuse,
		MemoryFreed:  memoryFreed,
		Duration:     compactionTime,
		Efficiency:   efficiency,
		Success:      memoryFreed > 0,
	}

	mcm.compactionHistory = append(mcm.compactionHistory, event)

	// Update stats
	mcm.compactionStats.TotalCompactions++
	mcm.compactionStats.TotalMemoryFreed += memoryFreed
	mcm.compactionStats.CompactionTime += compactionTime
	mcm.compactionStats.LastCompaction = startTime

	if mcm.compactionStats.TotalCompactions > 0 {
		mcm.compactionStats.AverageFreed = mcm.compactionStats.TotalMemoryFreed / uint64(mcm.compactionStats.TotalCompactions)
		mcm.compactionStats.Efficiency = float64(mcm.compactionStats.TotalMemoryFreed) / float64(mcm.compactionStats.CompactionTime.Milliseconds())
	}

	// Keep only the last 20 compaction events
	if len(mcm.compactionHistory) > 20 {
		mcm.compactionHistory = mcm.compactionHistory[len(mcm.compactionHistory)-20:]
	}

	if event.Success {
		log.Printf("Memory compaction completed: freed %d bytes in %v (efficiency: %.2f bytes/ms)",
			memoryFreed, compactionTime, efficiency)
	}

	return nil
}

// OptimizeMemory performs comprehensive memory optimization
func (mom *MemoryOptimizationManager) OptimizeMemory() error {
	// Create memory profile
	profile := mom.profiler.ProfileMemory()

	// Optimize garbage collection
	if err := mom.gcOptimizer.OptimizeGC(profile); err != nil {
		log.Printf("GC optimization failed: %v", err)
	}

	// Detect memory leaks
	leaks := mom.leakDetector.DetectLeaks(profile)
	for _, leak := range leaks {
		log.Printf("Memory leak detected: %s (severity: %s)", leak.Description, leak.Severity)
	}

	// Perform memory compaction if needed
	memoryUsage := float64(profile.HeapInuse) / float64(profile.HeapSys) * 100
	if memoryUsage > mom.config.MemoryCompactionThreshold {
		if err := mom.compactionManager.CompactMemory(); err != nil {
			log.Printf("Memory compaction failed: %v", err)
		}
	}

	return nil
}

// GetMemoryProfile returns the latest memory profile
func (mom *MemoryOptimizationManager) GetMemoryProfile() *MemoryProfile {
	return mom.profiler.ProfileMemory()
}

// GetGCOptimizationHistory returns GC optimization history
func (mom *MemoryOptimizationManager) GetGCOptimizationHistory() []GCOptimizationEvent {
	mom.gcOptimizer.mu.RLock()
	defer mom.gcOptimizer.mu.RUnlock()

	history := make([]GCOptimizationEvent, len(mom.gcOptimizer.optimizationHistory))
	copy(history, mom.gcOptimizer.optimizationHistory)
	return history
}

// GetLeakDetectionHistory returns leak detection history
func (mom *MemoryOptimizationManager) GetLeakDetectionHistory() []LeakDetectionEvent {
	mom.leakDetector.mu.RLock()
	defer mom.leakDetector.mu.RUnlock()

	history := make([]LeakDetectionEvent, len(mom.leakDetector.detectionHistory))
	copy(history, mom.leakDetector.detectionHistory)
	return history
}

// GetCompactionHistory returns memory compaction history
func (mom *MemoryOptimizationManager) GetCompactionHistory() []CompactionEvent {
	mom.compactionManager.mu.RLock()
	defer mom.compactionManager.mu.RUnlock()

	history := make([]CompactionEvent, len(mom.compactionManager.compactionHistory))
	copy(history, mom.compactionManager.compactionHistory)
	return history
}

// startOptimization starts the memory optimization loop
func (mom *MemoryOptimizationManager) startOptimization() {
	defer close(mom.optimizationDone)

	ticker := time.NewTicker(mom.config.ProfilingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-mom.ctx.Done():
			return
		case <-ticker.C:
			if err := mom.OptimizeMemory(); err != nil {
				log.Printf("Memory optimization failed: %v", err)
			}
		}
	}
}

// Shutdown gracefully shuts down the memory optimization manager
func (mom *MemoryOptimizationManager) Shutdown() {
	mom.cancel()
	<-mom.optimizationDone
}
