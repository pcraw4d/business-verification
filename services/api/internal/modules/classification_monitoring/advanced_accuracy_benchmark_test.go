package classification_monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

// BenchmarkAdvancedAccuracyTracker_TrackClassification benchmarks the main tracking function
func BenchmarkAdvancedAccuracyTracker_TrackClassification(b *testing.B) {
	logger := zap.NewNop()
	tracker := NewAdvancedAccuracyTracker(DefaultAdvancedAccuracyConfig(), logger)

	result := &ClassificationResult{
		ID:                     "benchmark_test",
		BusinessName:           "Benchmark Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry":       "restaurant",
			"model_version":  "v1.2.3",
			"trusted_source": "government_database",
		},
		IsCorrect: boolPtr(true),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result.ID = fmt.Sprintf("benchmark_test_%d", i)
		tracker.TrackClassification(context.Background(), result)
	}
}

// BenchmarkIndustryAccuracyMonitor_TrackIndustryClassification benchmarks industry tracking
func BenchmarkIndustryAccuracyMonitor_TrackIndustryClassification(b *testing.B) {
	logger := zap.NewNop()
	industryConfig := DefaultIndustryAccuracyConfig()
	monitor := NewIndustryAccuracyMonitor(industryConfig, logger)

	result := &ClassificationResult{
		ID:                     "benchmark_test",
		BusinessName:           "Benchmark Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry": "restaurant",
		},
		IsCorrect: boolPtr(true),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result.ID = fmt.Sprintf("benchmark_test_%d", i)
		monitor.TrackClassification(context.Background(), result)
	}
}

// BenchmarkRealTimeEnsembleMethodTracker_TrackMethodResult benchmarks ensemble method tracking
func BenchmarkRealTimeEnsembleMethodTracker_TrackMethodResult(b *testing.B) {
	logger := zap.NewNop()
	ensembleConfig := DefaultEnsembleMethodConfig()
	tracker := NewRealTimeEnsembleMethodTracker(ensembleConfig, logger)

	result := &ClassificationResult{
		ID:                     "benchmark_test",
		BusinessName:           "Benchmark Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ensemble",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"method": "ensemble",
		},
		IsCorrect: boolPtr(true),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result.ID = fmt.Sprintf("benchmark_test_%d", i)
		tracker.TrackMethodResult(context.Background(), result)
	}
}

// BenchmarkMLModelAccuracyMonitor_TrackModelPrediction benchmarks ML model tracking
func BenchmarkMLModelAccuracyMonitor_TrackModelPrediction(b *testing.B) {
	logger := zap.NewNop()
	mlConfig := DefaultMLModelMonitorConfig()
	monitor := NewMLModelAccuracyMonitor(mlConfig, logger)

	result := &ClassificationResult{
		ID:                     "benchmark_test",
		BusinessName:           "Benchmark Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ml_model",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"model_version": "v1.2.3",
		},
		IsCorrect: boolPtr(true),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result.ID = fmt.Sprintf("benchmark_test_%d", i)
		monitor.TrackModelPrediction(context.Background(), result)
	}
}

// BenchmarkSecurityMetricsAccuracyTracker_TrackTrustedDataSourceResult benchmarks security tracking
func BenchmarkSecurityMetricsAccuracyTracker_TrackTrustedDataSourceResult(b *testing.B) {
	config := DefaultSecurityMetricsConfig()
	logger := zap.NewNop()
	tracker := NewSecurityMetricsAccuracyTracker(config, logger)

	result := &ClassificationResult{
		ID:                     "benchmark_test",
		BusinessName:           "Benchmark Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "trusted_source",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"trusted_source": "government_database",
			"source_type":    "government",
		},
		IsCorrect: boolPtr(true),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result.ID = fmt.Sprintf("benchmark_test_%d", i)
		tracker.TrackTrustedDataSourceResult(context.Background(), result)
	}
}

// BenchmarkAdvancedAccuracyTracker_GetOverallMetrics benchmarks metrics retrieval
func BenchmarkAdvancedAccuracyTracker_GetOverallMetrics(b *testing.B) {
	logger := zap.NewNop()
	tracker := NewAdvancedAccuracyTracker(DefaultAdvancedAccuracyConfig(), logger)

	// Pre-populate with data
	for i := 0; i < 1000; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("benchmark_data_%d", i),
			BusinessName:           fmt.Sprintf("Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ensemble",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry": "restaurant",
			},
			IsCorrect: boolPtr(true),
		}
		tracker.TrackClassification(context.Background(), result)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.GetOverallAccuracy()
	}
}

// BenchmarkIndustryAccuracyMonitor_GetIndustryMetrics benchmarks industry metrics retrieval
func BenchmarkIndustryAccuracyMonitor_GetIndustryMetrics(b *testing.B) {
	logger := zap.NewNop()
	industryConfig := DefaultIndustryAccuracyConfig()
	monitor := NewIndustryAccuracyMonitor(industryConfig, logger)

	// Pre-populate with data
	for i := 0; i < 1000; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("benchmark_data_%d", i),
			BusinessName:           fmt.Sprintf("Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ensemble",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry": "restaurant",
			},
			IsCorrect: boolPtr(true),
		}
		monitor.TrackClassification(context.Background(), result)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.GetIndustryMetrics("restaurant")
	}
}

// BenchmarkRealTimeEnsembleMethodTracker_GetMethodMetrics benchmarks method metrics retrieval
func BenchmarkRealTimeEnsembleMethodTracker_GetMethodMetrics(b *testing.B) {
	logger := zap.NewNop()
	ensembleConfig := DefaultEnsembleMethodConfig()
	tracker := NewRealTimeEnsembleMethodTracker(ensembleConfig, logger)

	// Pre-populate with data
	for i := 0; i < 1000; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("benchmark_data_%d", i),
			BusinessName:           fmt.Sprintf("Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ensemble",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"method": "ensemble",
			},
			IsCorrect: boolPtr(true),
		}
		tracker.TrackMethodResult(context.Background(), result)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.GetMethodMetrics("ensemble")
	}
}

// BenchmarkMLModelAccuracyMonitor_GetModelMetrics benchmarks ML model metrics retrieval
func BenchmarkMLModelAccuracyMonitor_GetModelMetrics(b *testing.B) {
	logger := zap.NewNop()
	mlConfig := DefaultMLModelMonitorConfig()
	monitor := NewMLModelAccuracyMonitor(mlConfig, logger)

	// Pre-populate with data
	for i := 0; i < 1000; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("benchmark_data_%d", i),
			BusinessName:           fmt.Sprintf("Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ml_model",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"model_version": "v1.2.3",
			},
			IsCorrect: boolPtr(true),
		}
		monitor.TrackModelPrediction(context.Background(), result)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.GetModelMetrics("ml_model", "v1.2.3")
	}
}

// BenchmarkSecurityMetricsAccuracyTracker_GetTrustedDataSourceMetrics benchmarks security metrics retrieval
func BenchmarkSecurityMetricsAccuracyTracker_GetTrustedDataSourceMetrics(b *testing.B) {
	config := DefaultSecurityMetricsConfig()
	logger := zap.NewNop()
	tracker := NewSecurityMetricsAccuracyTracker(config, logger)

	// Pre-populate with data
	for i := 0; i < 1000; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("benchmark_data_%d", i),
			BusinessName:           fmt.Sprintf("Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "trusted_source",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"trusted_source": "government_database",
				"source_type":    "government",
			},
			IsCorrect: boolPtr(true),
		}
		tracker.TrackTrustedDataSourceResult(context.Background(), result)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.GetTrustedDataSourceMetrics("government_database")
	}
}

// BenchmarkAdvancedAccuracyTracker_ConcurrentTracking benchmarks concurrent tracking
func BenchmarkAdvancedAccuracyTracker_ConcurrentTracking(b *testing.B) {
	logger := zap.NewNop()
	tracker := NewAdvancedAccuracyTracker(DefaultAdvancedAccuracyConfig(), logger)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			result := &ClassificationResult{
				ID:                     fmt.Sprintf("concurrent_benchmark_%d", i),
				BusinessName:           fmt.Sprintf("Business %d", i),
				ActualClassification:   "restaurant",
				ExpectedClassification: stringPtr("restaurant"),
				ConfidenceScore:        0.95,
				ClassificationMethod:   "ensemble",
				Timestamp:              time.Now(),
				Metadata: map[string]interface{}{
					"industry": "restaurant",
				},
				IsCorrect: boolPtr(true),
			}
			tracker.TrackClassification(context.Background(), result)
			i++
		}
	})
}

// BenchmarkAdvancedAccuracyTracker_MemoryUsage benchmarks memory usage
func BenchmarkAdvancedAccuracyTracker_MemoryUsage(b *testing.B) {
	logger := zap.NewNop()
	tracker := NewAdvancedAccuracyTracker(DefaultAdvancedAccuracyConfig(), logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("memory_benchmark_%d", i),
			BusinessName:           fmt.Sprintf("Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ensemble",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry":       "restaurant",
				"model_version":  "v1.2.3",
				"trusted_source": "government_database",
			},
			IsCorrect: boolPtr(true),
		}
		tracker.TrackClassification(context.Background(), result)
	}
}

// BenchmarkAdvancedAccuracyTracker_LargeDataset benchmarks with large dataset
func BenchmarkAdvancedAccuracyTracker_LargeDataset(b *testing.B) {
	logger := zap.NewNop()
	tracker := NewAdvancedAccuracyTracker(DefaultAdvancedAccuracyConfig(), logger)

	// Pre-populate with large dataset
	for i := 0; i < 10000; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("large_dataset_%d", i),
			BusinessName:           fmt.Sprintf("Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ensemble",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry": "restaurant",
			},
			IsCorrect: boolPtr(true),
		}
		tracker.TrackClassification(context.Background(), result)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("large_dataset_new_%d", i),
			BusinessName:           fmt.Sprintf("New Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "ensemble",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry": "restaurant",
			},
			IsCorrect: boolPtr(true),
		}
		tracker.TrackClassification(context.Background(), result)
	}
}
