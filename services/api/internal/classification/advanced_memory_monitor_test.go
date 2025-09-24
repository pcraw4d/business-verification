package classification

import (
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestAdvancedMemoryMonitor_New(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Test with default config
	monitor := NewAdvancedMemoryMonitor(logger, nil)
	if monitor == nil {
		t.Error("Expected monitor to be created")
	}

	// Test with custom config
	config := &MemoryMonitorConfig{
		Enabled:            true,
		CollectionInterval: 1 * time.Second,
		MaxHistorySize:     100,
		PressureThreshold:  70.0,
		GCThreshold:        5 * time.Millisecond,
		AlertingEnabled:    true,
	}

	monitor2 := NewAdvancedMemoryMonitor(logger, config)
	if monitor2 == nil {
		t.Error("Expected monitor to be created with custom config")
	}

	// Stop monitors
	monitor.Stop()
	monitor2.Stop()
}

func TestAdvancedMemoryMonitor_StartStop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &MemoryMonitorConfig{
		Enabled:            true,
		CollectionInterval: 100 * time.Millisecond,
		MaxHistorySize:     10,
		AlertingEnabled:    true,
	}

	monitor := NewAdvancedMemoryMonitor(logger, config)

	// Start monitoring
	monitor.Start()

	// Wait for some data collection
	time.Sleep(200 * time.Millisecond)

	// Check that stats are being collected
	stats := monitor.GetCurrentStats()
	if stats == nil {
		t.Error("Expected stats to be collected")
	}

	if stats.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	if stats.AllocatedMB < 0 {
		t.Error("Expected allocated memory to be non-negative")
	}

	if stats.GoroutineCount <= 0 {
		t.Error("Expected goroutine count to be positive")
	}

	// Stop monitoring
	monitor.Stop()
}

func TestAdvancedMemoryMonitor_GetCurrentStats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &MemoryMonitorConfig{
		Enabled:            true,
		CollectionInterval: 50 * time.Millisecond,
		MaxHistorySize:     10,
	}

	monitor := NewAdvancedMemoryMonitor(logger, config)
	defer monitor.Stop()

	// Wait for data collection
	time.Sleep(100 * time.Millisecond)

	stats := monitor.GetCurrentStats()
	if stats == nil {
		t.Error("Expected stats to be available")
	}

	// Check basic fields
	if stats.AllocatedMB < 0 {
		t.Error("Expected allocated memory to be non-negative")
	}

	if stats.SystemMB < 0 {
		t.Error("Expected system memory to be non-negative")
	}

	if stats.GoroutineCount <= 0 {
		t.Error("Expected goroutine count to be positive")
	}

	if stats.NumGC < 0 {
		t.Error("Expected GC count to be non-negative")
	}

	// Check calculated fields
	if stats.MemoryPressureLevel == "" {
		t.Error("Expected memory pressure level to be set")
	}

	if stats.MemoryPressureScore < 0 {
		t.Error("Expected memory pressure score to be non-negative")
	}

	if stats.MemoryEfficiencyScore < 0 || stats.MemoryEfficiencyScore > 100 {
		t.Error("Expected memory efficiency score to be between 0 and 100")
	}
}

func TestAdvancedMemoryMonitor_GetMemoryHistory(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &MemoryMonitorConfig{
		Enabled:            true,
		CollectionInterval: 50 * time.Millisecond,
		MaxHistorySize:     5,
	}

	monitor := NewAdvancedMemoryMonitor(logger, config)
	defer monitor.Stop()

	// Wait for multiple data collections
	time.Sleep(300 * time.Millisecond)

	history := monitor.GetMemoryHistory(0) // Get all history
	if len(history) == 0 {
		t.Error("Expected history to contain data")
	}

	// Check that history is ordered by timestamp
	for i := 1; i < len(history); i++ {
		if history[i].Timestamp.Before(history[i-1].Timestamp) {
			t.Error("Expected history to be ordered by timestamp")
		}
	}

	// Test limited history
	limitedHistory := monitor.GetMemoryHistory(2)
	if len(limitedHistory) > 2 {
		t.Error("Expected limited history to respect limit")
	}

	// Test that limited history returns most recent entries
	if len(limitedHistory) == 2 && len(history) >= 2 {
		if limitedHistory[0].Timestamp.Before(history[len(history)-2].Timestamp) {
			t.Error("Expected limited history to return most recent entries")
		}
	}
}

func TestAdvancedMemoryMonitor_GetMemorySummary(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &MemoryMonitorConfig{
		Enabled:            true,
		CollectionInterval: 50 * time.Millisecond,
		MaxHistorySize:     10,
	}

	monitor := NewAdvancedMemoryMonitor(logger, config)
	defer monitor.Stop()

	// Wait for data collection
	time.Sleep(100 * time.Millisecond)

	summary := monitor.GetMemorySummary()
	if summary == nil {
		t.Error("Expected summary to be available")
	}

	// Check summary structure
	if summary["current"] == nil {
		t.Error("Expected summary to contain current stats")
	}

	if summary["gc"] == nil {
		t.Error("Expected summary to contain GC stats")
	}

	if summary["trends"] == nil {
		t.Error("Expected summary to contain trends")
	}

	if summary["history"] == nil {
		t.Error("Expected summary to contain history info")
	}

	// Check current stats
	current := summary["current"].(map[string]interface{})
	if current["allocated_mb"] == nil {
		t.Error("Expected current stats to contain allocated_mb")
	}

	if current["memory_pressure_level"] == nil {
		t.Error("Expected current stats to contain memory_pressure_level")
	}

	if current["memory_efficiency_score"] == nil {
		t.Error("Expected current stats to contain memory_efficiency_score")
	}
}

func TestAdvancedMemoryMonitor_ForceGC(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &MemoryMonitorConfig{
		Enabled:            true,
		CollectionInterval: 100 * time.Millisecond,
		MaxHistorySize:     10,
		DetailedGCStats:    true,
	}

	monitor := NewAdvancedMemoryMonitor(logger, config)
	defer monitor.Stop()

	// Force GC
	monitor.ForceGC()

	// Wait a bit for event processing
	time.Sleep(50 * time.Millisecond)

	// Check that GC stats were updated
	gcHistory := monitor.GetGCHistory(1)
	if len(gcHistory) == 0 {
		t.Error("Expected GC history to contain forced GC event")
	}

	gcStats := gcHistory[0]
	if gcStats.Timestamp.IsZero() {
		t.Error("Expected GC stats to have timestamp")
	}

	if gcStats.PauseTotalMs < 0 {
		t.Error("Expected GC pause time to be non-negative")
	}
}

func TestAdvancedMemoryMonitor_MemoryPressureDetection(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &MemoryMonitorConfig{
		Enabled:            true,
		CollectionInterval: 50 * time.Millisecond,
		MaxHistorySize:     10,
		PressureThreshold:  10.0, // Very low threshold for testing
		AlertingEnabled:    true,
	}

	monitor := NewAdvancedMemoryMonitor(logger, config)
	defer monitor.Stop()

	// Wait for data collection
	time.Sleep(100 * time.Millisecond)

	// Check for pressure events
	pressureEvents := monitor.GetMemoryPressureEvents(0)

	// Note: In a real test, you might need to artificially create memory pressure
	// For now, we just verify the structure works
	if pressureEvents == nil {
		t.Error("Expected pressure events list to be available")
	}
}

func TestAdvancedMemoryMonitor_ConfigDefaults(t *testing.T) {
	config := DefaultMemoryMonitorConfig()

	// Check default values
	if !config.Enabled {
		t.Error("Default config should be enabled")
	}

	if config.CollectionInterval != 5*time.Second {
		t.Error("Default collection interval should be 5 seconds")
	}

	if config.MaxHistorySize != 1000 {
		t.Error("Default max history size should be 1000")
	}

	if config.PressureThreshold != 80.0 {
		t.Error("Default pressure threshold should be 80%")
	}

	if config.GCThreshold != 10*time.Millisecond {
		t.Error("Default GC threshold should be 10ms")
	}

	if config.MemoryLeakThreshold != 10.0 {
		t.Error("Default memory leak threshold should be 10MB/sec")
	}

	if !config.AlertingEnabled {
		t.Error("Default config should have alerting enabled")
	}

	if !config.DetailedGCStats {
		t.Error("Default config should have detailed GC stats enabled")
	}

	if !config.TrackMemoryAllocations {
		t.Error("Default config should have memory allocation tracking enabled")
	}
}

func TestAdvancedMemoryMonitor_DisabledMode(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &MemoryMonitorConfig{
		Enabled: false,
	}

	monitor := NewAdvancedMemoryMonitor(logger, config)

	// Should not be monitoring
	stats := monitor.GetCurrentStats()
	if stats != nil {
		t.Error("Expected no stats when monitoring is disabled")
	}

	// Stop should not cause issues
	monitor.Stop()
}

func TestAdvancedMemoryMonitor_ConcurrentAccess(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &MemoryMonitorConfig{
		Enabled:            true,
		CollectionInterval: 10 * time.Millisecond,
		MaxHistorySize:     100,
	}

	monitor := NewAdvancedMemoryMonitor(logger, config)
	defer monitor.Stop()

	// Wait for some data collection
	time.Sleep(50 * time.Millisecond)

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Concurrent reads
			stats := monitor.GetCurrentStats()
			_ = stats

			history := monitor.GetMemoryHistory(10)
			_ = history

			summary := monitor.GetMemorySummary()
			_ = summary

			pressureEvents := monitor.GetMemoryPressureEvents(10)
			_ = pressureEvents
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not have caused any panics or deadlocks
}

// Benchmark tests
func BenchmarkAdvancedMemoryMonitor_GetCurrentStats(b *testing.B) {
	logger := zap.NewNop()
	config := &MemoryMonitorConfig{
		Enabled:            true,
		CollectionInterval: 1 * time.Second,
		MaxHistorySize:     100,
	}

	monitor := NewAdvancedMemoryMonitor(logger, config)
	defer monitor.Stop()

	// Wait for initial data collection
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats := monitor.GetCurrentStats()
		_ = stats
	}
}

func BenchmarkAdvancedMemoryMonitor_GetMemoryHistory(b *testing.B) {
	logger := zap.NewNop()
	config := &MemoryMonitorConfig{
		Enabled:            true,
		CollectionInterval: 10 * time.Millisecond,
		MaxHistorySize:     1000,
	}

	monitor := NewAdvancedMemoryMonitor(logger, config)
	defer monitor.Stop()

	// Wait for data collection
	time.Sleep(200 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		history := monitor.GetMemoryHistory(100)
		_ = history
	}
}

func BenchmarkAdvancedMemoryMonitor_ForceGC(b *testing.B) {
	logger := zap.NewNop()
	config := &MemoryMonitorConfig{
		Enabled:            true,
		CollectionInterval: 1 * time.Second,
		MaxHistorySize:     100,
		DetailedGCStats:    true,
	}

	monitor := NewAdvancedMemoryMonitor(logger, config)
	defer monitor.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.ForceGC()
	}
}
