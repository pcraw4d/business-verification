package classification

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap/zaptest"
)

// BenchmarkPerformanceMonitoringComponents benchmarks individual monitoring components
func BenchmarkPerformanceMonitoringComponents(b *testing.B) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(b)

	b.Run("ResponseTimeTracker", func(b *testing.B) {
		tracker := NewResponseTimeTracker(DefaultResponseTimeConfig(), logger)
		defer tracker.Stop()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				tracker.TrackResponseTime("test_endpoint", "GET", 50*time.Millisecond, 200, nil)
				i++
			}
		})
	})

	b.Run("MemoryMonitor", func(b *testing.B) {
		monitor := NewAdvancedMemoryMonitor(logger, DefaultMemoryMonitorConfig())
		defer monitor.Stop()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			monitor.CollectMemoryMetrics()
		}
	})

	b.Run("DatabaseMonitor", func(b *testing.B) {
		monitor := NewEnhancedDatabaseMonitor(db, logger, DefaultDatabaseMonitorConfig(), nil)
		defer monitor.Stop()

		ctx := context.Background()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				monitor.RecordQueryExecution(ctx, "SELECT * FROM test_table", 25*time.Millisecond, 1, 10, nil)
				i++
			}
		})
	})

	b.Run("SecurityMonitor", func(b *testing.B) {
		monitor := NewAdvancedSecurityValidationMonitor(logger, DefaultSecurityValidationConfig())
		defer monitor.Stop()

		ctx := context.Background()
		result := &AdvancedSecurityValidationResult{
			ValidationID:               "benchmark_validation",
			ValidationType:             "data_source_validation",
			ValidationName:             "benchmark_data_source",
			ExecutionTime:              30 * time.Millisecond,
			Success:                    true,
			Error:                      nil,
			SecurityViolation:          false,
			ComplianceViolation:        false,
			ThreatDetected:             false,
			VulnerabilityFound:         false,
			TrustScore:                 0.95,
			ConfidenceLevel:            0.90,
			RiskLevel:                  "low",
			SecurityRecommendations:    []string{},
			PerformanceRecommendations: []string{},
			Metadata:                   make(map[string]interface{}),
			Timestamp:                  time.Now(),
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				result.ValidationID = fmt.Sprintf("benchmark_validation_%d", i)
				monitor.RecordSecurityValidation(ctx, result)
				i++
			}
		})
	})

	b.Run("ComprehensiveMonitor", func(b *testing.B) {
		monitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
		defer monitor.Stop()

		ctx := context.Background()
		metric := &ComprehensivePerformanceMetric{
			ID:             "benchmark_metric",
			Timestamp:      time.Now(),
			MetricType:     "benchmark",
			ServiceName:    "benchmark_service",
			ResponseTimeMs: 50.0,
			Metadata:       make(map[string]interface{}),
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				metric.ID = fmt.Sprintf("benchmark_metric_%d", i)
				monitor.RecordPerformanceMetric(ctx, metric)
				i++
			}
		})
	})
}

// BenchmarkPerformanceMonitoringRetrieval benchmarks metric retrieval operations
func BenchmarkPerformanceMonitoringRetrieval(b *testing.B) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(b)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	// Pre-populate with metrics
	metricCount := 10000
	for i := 0; i < metricCount; i++ {
		metric := &ComprehensivePerformanceMetric{
			ID:             fmt.Sprintf("retrieval_benchmark_metric_%d", i),
			Timestamp:      time.Now(),
			MetricType:     "benchmark_retrieval",
			ServiceName:    fmt.Sprintf("benchmark_service_%d", i%10),
			ResponseTimeMs: float64(i % 100),
			Metadata:       make(map[string]interface{}),
		}
		comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
	}

	b.Run("GetPerformanceMetrics", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			comprehensiveMonitor.GetPerformanceMetrics(100)
		}
	})

	b.Run("GetPerformanceMetricsByService", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			serviceName := fmt.Sprintf("benchmark_service_%d", i%10)
			comprehensiveMonitor.GetPerformanceMetricsByService(serviceName, 100)
		}
	})

	b.Run("GetPerformanceAlerts", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			comprehensiveMonitor.GetPerformanceAlerts(false, 50)
		}
	})
}

// BenchmarkPerformanceMonitoringMemoryUsage benchmarks memory usage patterns
func BenchmarkPerformanceMonitoringMemoryUsage(b *testing.B) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(b)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	b.Run("MemoryUsageUnderLoad", func(b *testing.B) {
		var memStatsBefore, memStatsAfter runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&memStatsBefore)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			metric := &ComprehensivePerformanceMetric{
				ID:             fmt.Sprintf("memory_benchmark_metric_%d", i),
				Timestamp:      time.Now(),
				MetricType:     "memory_benchmark",
				ServiceName:    "memory_benchmark_service",
				ResponseTimeMs: float64(i % 100),
				MemoryUsageMB:  float64(i % 500),
				Metadata: map[string]interface{}{
					"large_data": make([]byte, 1024), // 1KB of data
				},
			}
			comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
		}

		runtime.GC()
		runtime.ReadMemStats(&memStatsAfter)

		memoryIncrease := memStatsAfter.Alloc - memStatsBefore.Alloc
		memoryIncreaseMB := float64(memoryIncrease) / 1024 / 1024

		b.ReportMetric(memoryIncreaseMB, "MB/memory_increase")
		b.ReportMetric(float64(memoryIncrease)/float64(b.N), "bytes/op")
	})
}

// BenchmarkPerformanceMonitoringConcurrency benchmarks concurrent operations
func BenchmarkPerformanceMonitoringConcurrency(b *testing.B) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(b)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	b.Run("ConcurrentMetricRecording", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				metric := &ComprehensivePerformanceMetric{
					ID:             fmt.Sprintf("concurrent_benchmark_metric_%d", i),
					Timestamp:      time.Now(),
					MetricType:     "concurrent_benchmark",
					ServiceName:    "concurrent_benchmark_service",
					ResponseTimeMs: float64(i % 100),
					Metadata:       make(map[string]interface{}),
				}
				comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
				i++
			}
		})
	})

	b.Run("ConcurrentMetricRetrieval", func(b *testing.B) {
		// Pre-populate with metrics
		for i := 0; i < 1000; i++ {
			metric := &ComprehensivePerformanceMetric{
				ID:             fmt.Sprintf("concurrent_retrieval_metric_%d", i),
				Timestamp:      time.Now(),
				MetricType:     "concurrent_retrieval",
				ServiceName:    "concurrent_retrieval_service",
				ResponseTimeMs: float64(i % 100),
				Metadata:       make(map[string]interface{}),
			}
			comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				comprehensiveMonitor.GetPerformanceMetrics(100)
			}
		})
	})
}

// BenchmarkPerformanceMonitoringDataPersistence benchmarks data persistence operations
func BenchmarkPerformanceMonitoringDataPersistence(b *testing.B) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(b)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	b.Run("DataPersistence", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			metric := &ComprehensivePerformanceMetric{
				ID:             fmt.Sprintf("persistence_benchmark_metric_%d", i),
				Timestamp:      time.Now(),
				MetricType:     "persistence_benchmark",
				ServiceName:    "persistence_benchmark_service",
				ResponseTimeMs: float64(i % 100),
				Metadata: map[string]interface{}{
					"persistent_data": fmt.Sprintf("data_%d", i),
				},
			}
			comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
		}
	})
}

// BenchmarkPerformanceMonitoringAlerting benchmarks alerting operations
func BenchmarkPerformanceMonitoringAlerting(b *testing.B) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(b)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	b.Run("AlertGeneration", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create metrics that should trigger alerts
			alertMetric := &ComprehensivePerformanceMetric{
				ID:             fmt.Sprintf("alert_benchmark_metric_%d", i),
				Timestamp:      time.Now(),
				MetricType:     "alert_benchmark",
				ServiceName:    "alert_benchmark_service",
				ResponseTimeMs: 5000.0,   // High response time to trigger alert
				ErrorOccurred:  i%2 == 0, // Alternate errors
				ErrorMessage:   "Benchmark error",
				Metadata:       make(map[string]interface{}),
			}
			comprehensiveMonitor.RecordPerformanceMetric(ctx, alertMetric)
		}
	})

	b.Run("AlertRetrieval", func(b *testing.B) {
		// Pre-populate with alert-triggering metrics
		for i := 0; i < 1000; i++ {
			alertMetric := &ComprehensivePerformanceMetric{
				ID:             fmt.Sprintf("alert_retrieval_metric_%d", i),
				Timestamp:      time.Now(),
				MetricType:     "alert_retrieval",
				ServiceName:    "alert_retrieval_service",
				ResponseTimeMs: 5000.0,
				ErrorOccurred:  true,
				ErrorMessage:   "Benchmark error",
				Metadata:       make(map[string]interface{}),
			}
			comprehensiveMonitor.RecordPerformanceMetric(ctx, alertMetric)
		}

		// Allow time for alert processing
		time.Sleep(100 * time.Millisecond)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			comprehensiveMonitor.GetPerformanceAlerts(false, 100)
		}
	})
}

// BenchmarkPerformanceMonitoringCleanup benchmarks cleanup operations
func BenchmarkPerformanceMonitoringCleanup(b *testing.B) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(b)
	config := DefaultPerformanceMonitorConfig()
	config.MaxMetrics = 1000 // Set a limit for cleanup testing
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, config)
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	b.Run("CleanupOperations", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Record metrics that will trigger cleanup
			metric := &ComprehensivePerformanceMetric{
				ID:             fmt.Sprintf("cleanup_benchmark_metric_%d", i),
				Timestamp:      time.Now(),
				MetricType:     "cleanup_benchmark",
				ServiceName:    "cleanup_benchmark_service",
				ResponseTimeMs: float64(i % 100),
				Metadata:       make(map[string]interface{}),
			}
			comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
		}
	})
}

// BenchmarkPerformanceMonitoringSystemLoad benchmarks the system under various loads
func BenchmarkPerformanceMonitoringSystemLoad(b *testing.B) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(b)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	loadScenarios := []struct {
		name             string
		workers          int
		metricsPerWorker int
	}{
		{"LightLoad", 1, 100},
		{"MediumLoad", 5, 200},
		{"HeavyLoad", 10, 500},
		{"ExtremeLoad", 20, 1000},
	}

	for _, scenario := range loadScenarios {
		b.Run(scenario.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup

				for w := 0; w < scenario.workers; w++ {
					wg.Add(1)
					go func(workerID int) {
						defer wg.Done()
						for m := 0; m < scenario.metricsPerWorker; m++ {
							metric := &ComprehensivePerformanceMetric{
								ID:             fmt.Sprintf("load_benchmark_%d_%d_%d", i, workerID, m),
								Timestamp:      time.Now(),
								MetricType:     "load_benchmark",
								ServiceName:    fmt.Sprintf("load_service_%d", workerID),
								ResponseTimeMs: float64(m % 100),
								Metadata:       make(map[string]interface{}),
							}
							comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
						}
					}(w)
				}

				wg.Wait()
			}
		})
	}
}

// BenchmarkPerformanceMonitoringMemoryEfficiency benchmarks memory efficiency
func BenchmarkPerformanceMonitoringMemoryEfficiency(b *testing.B) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(b)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	b.Run("MemoryEfficiency", func(b *testing.B) {
		var memStatsBefore, memStatsAfter runtime.MemStats

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if i == 0 {
				runtime.GC()
				runtime.ReadMemStats(&memStatsBefore)
			}

			metric := &ComprehensivePerformanceMetric{
				ID:             fmt.Sprintf("memory_efficiency_metric_%d", i),
				Timestamp:      time.Now(),
				MetricType:     "memory_efficiency",
				ServiceName:    "memory_efficiency_service",
				ResponseTimeMs: float64(i % 100),
				Metadata:       make(map[string]interface{}),
			}
			comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)

			if i == b.N-1 {
				runtime.GC()
				runtime.ReadMemStats(&memStatsAfter)
			}
		}

		if b.N > 0 {
			memoryIncrease := memStatsAfter.Alloc - memStatsBefore.Alloc
			memoryPerOp := float64(memoryIncrease) / float64(b.N)
			b.ReportMetric(memoryPerOp, "bytes/op")
		}
	})
}

// BenchmarkPerformanceMonitoringLatency benchmarks operation latency
func BenchmarkPerformanceMonitoringLatency(b *testing.B) {
	db := createComprehensiveTestDB()
	defer db.Close()

	logger := zaptest.NewLogger(b)
	comprehensiveMonitor := NewComprehensivePerformanceMonitor(db, logger, DefaultPerformanceMonitorConfig())
	defer comprehensiveMonitor.Stop()

	ctx := context.Background()

	b.Run("MetricRecordingLatency", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()
			metric := &ComprehensivePerformanceMetric{
				ID:             fmt.Sprintf("latency_benchmark_metric_%d", i),
				Timestamp:      time.Now(),
				MetricType:     "latency_benchmark",
				ServiceName:    "latency_benchmark_service",
				ResponseTimeMs: float64(i % 100),
				Metadata:       make(map[string]interface{}),
			}
			comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
			latency := time.Since(start)
			b.ReportMetric(float64(latency.Nanoseconds()), "ns/op")
		}
	})

	b.Run("MetricRetrievalLatency", func(b *testing.B) {
		// Pre-populate with metrics
		for i := 0; i < 1000; i++ {
			metric := &ComprehensivePerformanceMetric{
				ID:             fmt.Sprintf("retrieval_latency_metric_%d", i),
				Timestamp:      time.Now(),
				MetricType:     "retrieval_latency",
				ServiceName:    "retrieval_latency_service",
				ResponseTimeMs: float64(i % 100),
				Metadata:       make(map[string]interface{}),
			}
			comprehensiveMonitor.RecordPerformanceMetric(ctx, metric)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()
			comprehensiveMonitor.GetPerformanceMetrics(100)
			latency := time.Since(start)
			b.ReportMetric(float64(latency.Nanoseconds()), "ns/op")
		}
	})
}
