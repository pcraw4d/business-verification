package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// MemoryOptimizationAPI provides RESTful API endpoints for memory optimization
type MemoryOptimizationAPI struct {
	memoryManager *MemoryOptimizationManager
}

// NewMemoryOptimizationAPI creates a new memory optimization API
func NewMemoryOptimizationAPI(memoryManager *MemoryOptimizationManager) *MemoryOptimizationAPI {
	return &MemoryOptimizationAPI{
		memoryManager: memoryManager,
	}
}

// RegisterMemoryOptimizationRoutes registers all memory optimization routes
func (moa *MemoryOptimizationAPI) RegisterMemoryOptimizationRoutes(mux *http.ServeMux) {
	// Memory profiling endpoints
	mux.HandleFunc("GET /v1/memory/profile", moa.GetMemoryProfile)
	mux.HandleFunc("GET /v1/memory/profile/history", moa.GetMemoryProfileHistory)

	// GC optimization endpoints
	mux.HandleFunc("GET /v1/memory/gc/stats", moa.GetGCStats)
	mux.HandleFunc("GET /v1/memory/gc/history", moa.GetGCOptimizationHistory)
	mux.HandleFunc("POST /v1/memory/gc/optimize", moa.TriggerGCOptimization)

	// Memory pooling endpoints
	mux.HandleFunc("GET /v1/memory/pools", moa.GetMemoryPools)
	mux.HandleFunc("POST /v1/memory/pools", moa.CreateMemoryPool)
	mux.HandleFunc("GET /v1/memory/pools/{name}", moa.GetMemoryPool)
	mux.HandleFunc("DELETE /v1/memory/pools/{name}", moa.DeleteMemoryPool)

	// Memory leak detection endpoints
	mux.HandleFunc("GET /v1/memory/leaks", moa.GetLeakDetectionHistory)
	mux.HandleFunc("POST /v1/memory/leaks/detect", moa.TriggerLeakDetection)
	mux.HandleFunc("GET /v1/memory/leaks/patterns", moa.GetLeakPatterns)

	// Memory compaction endpoints
	mux.HandleFunc("GET /v1/memory/compaction/stats", moa.GetCompactionStats)
	mux.HandleFunc("GET /v1/memory/compaction/history", moa.GetCompactionHistory)
	mux.HandleFunc("POST /v1/memory/compaction/compact", moa.TriggerMemoryCompaction)

	// Comprehensive memory optimization endpoints
	mux.HandleFunc("POST /v1/memory/optimize", moa.TriggerMemoryOptimization)
	mux.HandleFunc("GET /v1/memory/status", moa.GetMemoryStatus)
	mux.HandleFunc("GET /v1/memory/health", moa.GetMemoryHealth)
}

// GetMemoryProfile returns the current memory profile
func (moa *MemoryOptimizationAPI) GetMemoryProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	profile := moa.memoryManager.GetMemoryProfile()

	response := map[string]interface{}{
		"timestamp": profile.Timestamp.Format(time.RFC3339),
		"heap": map[string]interface{}{
			"alloc":    profile.HeapAlloc,
			"sys":      profile.HeapSys,
			"inuse":    profile.HeapInuse,
			"idle":     profile.HeapIdle,
			"released": profile.HeapReleased,
			"objects":  profile.HeapObjects,
		},
		"stack": map[string]interface{}{
			"inuse": profile.StackInuse,
			"sys":   profile.StackSys,
		},
		"gc": map[string]interface{}{
			"num_gc":         profile.NumGC,
			"num_forced_gc":  profile.NumForcedGC,
			"cpu_fraction":   profile.GCCPUFraction,
			"pause_total_ns": profile.PauseTotalNs,
			"next_gc":        profile.NextGC,
			"last_gc":        profile.LastGC,
		},
		"allocation_rate": profile.AllocationRate,
		"gc_trigger_rate": profile.GCTriggerRate,
	}

	json.NewEncoder(w).Encode(response)
}

// GetMemoryProfileHistory returns memory profile history
func (moa *MemoryOptimizationAPI) GetMemoryProfileHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get limit from query parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// For now, return a simplified response since we don't expose the full history
	// In a real implementation, you would return the actual profile history
	response := map[string]interface{}{
		"profiles": []map[string]interface{}{},
		"limit":    limit,
		"message":  "Profile history endpoint - implementation needed",
	}

	json.NewEncoder(w).Encode(response)
}

// GetGCStats returns garbage collection statistics
func (moa *MemoryOptimizationAPI) GetGCStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current memory profile to calculate GC stats
	profile := moa.memoryManager.GetMemoryProfile()

	// Calculate basic GC stats
	gcStats := map[string]interface{}{
		"num_gc":         profile.NumGC,
		"num_forced_gc":  profile.NumForcedGC,
		"cpu_fraction":   profile.GCCPUFraction,
		"pause_total_ns": profile.PauseTotalNs,
		"next_gc":        profile.NextGC,
		"last_gc":        profile.LastGC,
		"heap_alloc":     profile.HeapAlloc,
		"heap_sys":       profile.HeapSys,
		"heap_inuse":     profile.HeapInuse,
	}

	json.NewEncoder(w).Encode(gcStats)
}

// GetGCOptimizationHistory returns GC optimization history
func (moa *MemoryOptimizationAPI) GetGCOptimizationHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	history := moa.memoryManager.GetGCOptimizationHistory()

	// Convert to response format
	events := make([]map[string]interface{}, len(history))
	for i, event := range history {
		events[i] = map[string]interface{}{
			"timestamp":         event.Timestamp.Format(time.RFC3339),
			"event_type":        event.EventType,
			"before_percentage": event.BeforePercentage,
			"after_percentage":  event.AfterPercentage,
			"memory_freed":      event.MemoryFreed,
			"description":       event.Description,
			"success":           event.Success,
		}
	}

	response := map[string]interface{}{
		"events": events,
		"count":  len(events),
	}

	json.NewEncoder(w).Encode(response)
}

// TriggerGCOptimization triggers manual GC optimization
func (moa *MemoryOptimizationAPI) TriggerGCOptimization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current profile and trigger optimization
	profile := moa.memoryManager.GetMemoryProfile()
	err := moa.memoryManager.gcOptimizer.OptimizeGC(profile)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "GC optimization failed",
			"message": err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"message":   "GC optimization triggered successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// GetMemoryPools returns all memory pools
func (moa *MemoryOptimizationAPI) GetMemoryPools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get pool information from the memory pooler
	pools := make([]map[string]interface{}, 0)

	moa.memoryManager.memoryPooler.mu.RLock()
	for name, pool := range moa.memoryManager.memoryPooler.pools {
		pools = append(pools, map[string]interface{}{
			"name":            name,
			"object_size":     pool.ObjectSize,
			"max_objects":     pool.MaxObjects,
			"current_objects": pool.CurrentObjects,
			"allocations":     pool.Allocations,
			"releases":        pool.Releases,
			"hit_rate":        pool.HitRate,
			"last_used":       pool.LastUsed.Format(time.RFC3339),
		})
	}
	moa.memoryManager.memoryPooler.mu.RUnlock()

	response := map[string]interface{}{
		"pools": pools,
		"count": len(pools),
		"stats": map[string]interface{}{
			"total_pools":       moa.memoryManager.memoryPooler.stats.TotalPools,
			"total_allocations": moa.memoryManager.memoryPooler.stats.TotalAllocations,
			"total_releases":    moa.memoryManager.memoryPooler.stats.TotalReleases,
			"average_hit_rate":  moa.memoryManager.memoryPooler.stats.AverageHitRate,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// CreateMemoryPool creates a new memory pool
func (moa *MemoryOptimizationAPI) CreateMemoryPool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request struct {
		Name       string `json:"name"`
		ObjectSize uint64 `json:"object_size"`
		MaxObjects int    `json:"max_objects"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid request body",
		})
		return
	}

	if request.Name == "" || request.ObjectSize == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Name and object_size are required",
		})
		return
	}

	// Create the pool
	pool := moa.memoryManager.memoryPooler.CreatePool(request.Name, request.ObjectSize, request.MaxObjects)

	response := map[string]interface{}{
		"message": "Memory pool created successfully",
		"pool": map[string]interface{}{
			"name":        pool.Name,
			"object_size": pool.ObjectSize,
			"max_objects": pool.MaxObjects,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// GetMemoryPool returns a specific memory pool
func (moa *MemoryOptimizationAPI) GetMemoryPool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract pool name from URL path
	// This is a simplified implementation - in a real system you'd use a proper router
	poolName := r.URL.Query().Get("name")
	if poolName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Pool name is required",
		})
		return
	}

	moa.memoryManager.memoryPooler.mu.RLock()
	pool, exists := moa.memoryManager.memoryPooler.pools[poolName]
	moa.memoryManager.memoryPooler.mu.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Pool not found",
		})
		return
	}

	response := map[string]interface{}{
		"name":            pool.Name,
		"object_size":     pool.ObjectSize,
		"max_objects":     pool.MaxObjects,
		"current_objects": pool.CurrentObjects,
		"allocations":     pool.Allocations,
		"releases":        pool.Releases,
		"hit_rate":        pool.HitRate,
		"last_used":       pool.LastUsed.Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// DeleteMemoryPool deletes a memory pool
func (moa *MemoryOptimizationAPI) DeleteMemoryPool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract pool name from URL path
	poolName := r.URL.Query().Get("name")
	if poolName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Pool name is required",
		})
		return
	}

	moa.memoryManager.memoryPooler.mu.Lock()
	delete(moa.memoryManager.memoryPooler.pools, poolName)
	moa.memoryManager.memoryPooler.mu.Unlock()

	response := map[string]interface{}{
		"message":   "Memory pool deleted successfully",
		"pool_name": poolName,
	}

	json.NewEncoder(w).Encode(response)
}

// GetLeakDetectionHistory returns memory leak detection history
func (moa *MemoryOptimizationAPI) GetLeakDetectionHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	history := moa.memoryManager.GetLeakDetectionHistory()

	// Convert to response format
	events := make([]map[string]interface{}, len(history))
	for i, event := range history {
		events[i] = map[string]interface{}{
			"timestamp":     event.Timestamp.Format(time.RFC3339),
			"pattern_id":    event.PatternID,
			"severity":      event.Severity,
			"memory_growth": event.MemoryGrowth,
			"duration":      event.Duration.String(),
			"description":   event.Description,
			"resolved":      event.Resolved,
		}
	}

	response := map[string]interface{}{
		"events": events,
		"count":  len(events),
	}

	json.NewEncoder(w).Encode(response)
}

// TriggerLeakDetection triggers manual memory leak detection
func (moa *MemoryOptimizationAPI) TriggerLeakDetection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current profile and trigger leak detection
	profile := moa.memoryManager.GetMemoryProfile()
	leaks := moa.memoryManager.leakDetector.DetectLeaks(profile)

	response := map[string]interface{}{
		"message":     "Leak detection completed",
		"leaks_found": len(leaks),
		"timestamp":   time.Now().Format(time.RFC3339),
		"leaks":       make([]map[string]interface{}, len(leaks)),
	}

	for i, leak := range leaks {
		response["leaks"].([]map[string]interface{})[i] = map[string]interface{}{
			"pattern_id":    leak.PatternID,
			"severity":      leak.Severity,
			"description":   leak.Description,
			"memory_growth": leak.MemoryGrowth,
		}
	}

	json.NewEncoder(w).Encode(response)
}

// GetLeakPatterns returns configured leak detection patterns
func (moa *MemoryOptimizationAPI) GetLeakPatterns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	patterns := make([]map[string]interface{}, len(moa.memoryManager.leakDetector.leakPatterns))

	moa.memoryManager.leakDetector.mu.RLock()
	for i, pattern := range moa.memoryManager.leakDetector.leakPatterns {
		patterns[i] = map[string]interface{}{
			"id":          pattern.ID,
			"name":        pattern.Name,
			"description": pattern.Description,
			"threshold":   pattern.Threshold,
			"time_window": pattern.TimeWindow.String(),
			"severity":    pattern.Severity,
		}
	}
	moa.memoryManager.leakDetector.mu.RUnlock()

	response := map[string]interface{}{
		"patterns": patterns,
		"count":    len(patterns),
	}

	json.NewEncoder(w).Encode(response)
}

// GetCompactionStats returns memory compaction statistics
func (moa *MemoryOptimizationAPI) GetCompactionStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats := moa.memoryManager.compactionManager.compactionStats

	response := map[string]interface{}{
		"total_compactions":  stats.TotalCompactions,
		"total_memory_freed": stats.TotalMemoryFreed,
		"average_freed":      stats.AverageFreed,
		"compaction_time":    stats.CompactionTime.String(),
		"last_compaction":    stats.LastCompaction.Format(time.RFC3339),
		"efficiency":         stats.Efficiency,
	}

	json.NewEncoder(w).Encode(response)
}

// GetCompactionHistory returns memory compaction history
func (moa *MemoryOptimizationAPI) GetCompactionHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	history := moa.memoryManager.GetCompactionHistory()

	// Convert to response format
	events := make([]map[string]interface{}, len(history))
	for i, event := range history {
		events[i] = map[string]interface{}{
			"timestamp":     event.Timestamp.Format(time.RFC3339),
			"memory_before": event.MemoryBefore,
			"memory_after":  event.MemoryAfter,
			"memory_freed":  event.MemoryFreed,
			"duration":      event.Duration.String(),
			"efficiency":    event.Efficiency,
			"success":       event.Success,
		}
	}

	response := map[string]interface{}{
		"events": events,
		"count":  len(events),
	}

	json.NewEncoder(w).Encode(response)
}

// TriggerMemoryCompaction triggers manual memory compaction
func (moa *MemoryOptimizationAPI) TriggerMemoryCompaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := moa.memoryManager.compactionManager.CompactMemory()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Memory compaction failed",
			"message": err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"message":   "Memory compaction completed successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// TriggerMemoryOptimization triggers comprehensive memory optimization
func (moa *MemoryOptimizationAPI) TriggerMemoryOptimization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := moa.memoryManager.OptimizeMemory()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Memory optimization failed",
			"message": err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"message":   "Memory optimization completed successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// GetMemoryStatus returns comprehensive memory status
func (moa *MemoryOptimizationAPI) GetMemoryStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	profile := moa.memoryManager.GetMemoryProfile()

	// Calculate memory usage percentages
	heapUsage := float64(profile.HeapInuse) / float64(profile.HeapSys) * 100
	systemUsage := float64(profile.HeapSys) / float64(profile.HeapSys) * 100

	response := map[string]interface{}{
		"timestamp": profile.Timestamp.Format(time.RFC3339),
		"memory_usage": map[string]interface{}{
			"heap_percentage":   heapUsage,
			"system_percentage": systemUsage,
			"heap_alloc":        profile.HeapAlloc,
			"heap_sys":          profile.HeapSys,
			"heap_inuse":        profile.HeapInuse,
			"heap_idle":         profile.HeapIdle,
		},
		"gc_status": map[string]interface{}{
			"num_gc":       profile.NumGC,
			"cpu_fraction": profile.GCCPUFraction,
			"next_gc":      profile.NextGC,
			"last_gc":      profile.LastGC,
		},
		"optimization": map[string]interface{}{
			"profiling_enabled":      moa.memoryManager.config.EnableMemoryProfiling,
			"leak_detection_enabled": moa.memoryManager.config.LeakDetectionEnabled,
			"pooling_enabled":        moa.memoryManager.config.PoolingEnabled,
			"compaction_enabled":     moa.memoryManager.config.CompactionEnabled,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// GetMemoryHealth returns memory health status
func (moa *MemoryOptimizationAPI) GetMemoryHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	profile := moa.memoryManager.GetMemoryProfile()

	// Calculate health score based on various factors
	healthScore := 100

	// Reduce score based on memory usage
	heapUsage := float64(profile.HeapInuse) / float64(profile.HeapSys) * 100
	if heapUsage > 90 {
		healthScore -= 30
	} else if heapUsage > 80 {
		healthScore -= 20
	} else if heapUsage > 70 {
		healthScore -= 10
	}

	// Reduce score based on GC frequency
	if profile.GCCPUFraction > 0.5 {
		healthScore -= 20
	} else if profile.GCCPUFraction > 0.3 {
		healthScore -= 10
	}

	// Ensure score doesn't go below 0
	if healthScore < 0 {
		healthScore = 0
	}

	status := "healthy"
	if healthScore < 50 {
		status = "critical"
	} else if healthScore < 70 {
		status = "warning"
	}

	response := map[string]interface{}{
		"status":          status,
		"health_score":    healthScore,
		"timestamp":       profile.Timestamp.Format(time.RFC3339),
		"heap_usage":      heapUsage,
		"gc_cpu_fraction": profile.GCCPUFraction,
		"recommendations": []string{},
	}

	// Add recommendations based on health
	if heapUsage > 80 {
		response["recommendations"] = append(response["recommendations"].([]string), "Consider memory compaction")
	}
	if profile.GCCPUFraction > 0.3 {
		response["recommendations"] = append(response["recommendations"].([]string), "High GC overhead - consider optimization")
	}

	json.NewEncoder(w).Encode(response)
}
