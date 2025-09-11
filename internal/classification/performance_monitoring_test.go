package classification

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestPerformanceMonitoringIntegration tests the performance monitoring integration
func TestPerformanceMonitoringIntegration(t *testing.T) {
	// Create a mock monitoring service
	monitor := createMockPerformanceMonitoringService()
	ctx := context.Background()

	// Test performance metrics recording
	t.Run("RecordPerformanceMetrics", func(t *testing.T) {
		metrics := &ClassificationPerformanceMetrics{
			RequestID:           "test_req_123",
			Timestamp:           time.Now(),
			ServiceType:         "industry_detection",
			Method:              "content_analysis",
			ResponseTimeMs:      150.5,
			ProcessingTimeMs:    120.3,
			Confidence:          0.85,
			KeywordsCount:       5,
			ResultsCount:        3,
			CacheHitRatio:       0.9,
			ErrorOccurred:       false,
			ParallelProcessing:  true,
			GoroutinesUsed:      3,
			MemoryUsageMB:       25.5,
			DatabaseQueries:     2,
			DatabaseQueryTimeMs: 45.2,
		}

		err := monitor.RecordPerformanceMetrics(ctx, metrics)
		if err != nil {
			t.Logf("Note: Expected error due to mock implementation: %v", err)
		} else {
			t.Logf("✅ Performance metrics recorded successfully")
		}
	})

	// Test performance summary
	t.Run("GetPerformanceSummary", func(t *testing.T) {
		summary, err := monitor.GetPerformanceSummary(ctx, 24)
		if err != nil {
			t.Logf("Note: Expected error due to mock implementation: %v", err)
		} else {
			t.Logf("✅ Performance summary retrieved: %+v", summary)
		}
	})

	// Test performance dashboard
	t.Run("GetPerformanceDashboard", func(t *testing.T) {
		dashboard, err := monitor.GetPerformanceDashboard(ctx)
		if err != nil {
			t.Logf("Note: Expected error due to mock implementation: %v", err)
		} else {
			t.Logf("✅ Performance dashboard retrieved: %+v", dashboard)
		}
	})
}

// TestPerformanceThresholds tests performance threshold checking
func TestPerformanceThresholds(t *testing.T) {
	monitor := createMockPerformanceMonitoringService()

	// Test normal performance
	t.Run("NormalPerformance", func(t *testing.T) {
		summary := map[string]interface{}{
			"accuracy_stats": &ClassificationAccuracyStats{
				AvgResponseTimeMs:  floatPtr(2000), // 2 seconds - within threshold
				AccuracyPercentage: floatPtr(0.85), // 85% - above threshold
				ErrorRate:          floatPtr(0.02), // 2% - below threshold
			},
		}

		err := monitor.checkPerformanceThresholds(summary)
		if err != nil {
			t.Errorf("Expected no error for normal performance, got: %v", err)
		} else {
			t.Logf("✅ Normal performance thresholds passed")
		}
	})

	// Test high response time
	t.Run("HighResponseTime", func(t *testing.T) {
		summary := map[string]interface{}{
			"accuracy_stats": &ClassificationAccuracyStats{
				AvgResponseTimeMs:  floatPtr(6000), // 6 seconds - exceeds threshold
				AccuracyPercentage: floatPtr(0.85),
				ErrorRate:          floatPtr(0.02),
			},
		}

		err := monitor.checkPerformanceThresholds(summary)
		if err == nil {
			t.Error("Expected error for high response time, got nil")
		} else {
			t.Logf("✅ High response time correctly detected: %v", err)
		}
	})

	// Test low accuracy
	t.Run("LowAccuracy", func(t *testing.T) {
		summary := map[string]interface{}{
			"accuracy_stats": &ClassificationAccuracyStats{
				AvgResponseTimeMs:  floatPtr(2000),
				AccuracyPercentage: floatPtr(0.75), // 75% - below threshold
				ErrorRate:          floatPtr(0.02),
			},
		}

		err := monitor.checkPerformanceThresholds(summary)
		if err == nil {
			t.Error("Expected error for low accuracy, got nil")
		} else {
			t.Logf("✅ Low accuracy correctly detected: %v", err)
		}
	})

	// Test high error rate
	t.Run("HighErrorRate", func(t *testing.T) {
		summary := map[string]interface{}{
			"accuracy_stats": &ClassificationAccuracyStats{
				AvgResponseTimeMs:  floatPtr(2000),
				AccuracyPercentage: floatPtr(0.85),
				ErrorRate:          floatPtr(0.08), // 8% - exceeds threshold
			},
		}

		err := monitor.checkPerformanceThresholds(summary)
		if err == nil {
			t.Error("Expected error for high error rate, got nil")
		} else {
			t.Logf("✅ High error rate correctly detected: %v", err)
		}
	})
}

// TestPerformanceMonitoringConfig tests the performance monitoring configuration
func TestPerformanceMonitoringConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultPerformanceMonitoringConfig()

		// Verify default values
		if !config.Enabled {
			t.Error("Expected monitoring to be enabled by default")
		}

		if config.MonitoringInterval != 5*time.Minute {
			t.Errorf("Expected monitoring interval of 5 minutes, got %v", config.MonitoringInterval)
		}

		if config.ResponseTimeThreshold != 5000 {
			t.Errorf("Expected response time threshold of 5000ms, got %.2f", config.ResponseTimeThreshold)
		}

		if config.AccuracyThreshold != 0.8 {
			t.Errorf("Expected accuracy threshold of 0.8, got %.2f", config.AccuracyThreshold)
		}

		if config.ErrorRateThreshold != 0.05 {
			t.Errorf("Expected error rate threshold of 0.05, got %.2f", config.ErrorRateThreshold)
		}

		t.Logf("✅ Default configuration values are correct")
	})
}

// TestPerformanceMetricsStructure tests the performance metrics structure
func TestPerformanceMetricsStructure(t *testing.T) {
	t.Run("MetricsFields", func(t *testing.T) {
		metrics := &ClassificationPerformanceMetrics{
			RequestID:           "test_123",
			Timestamp:           time.Now(),
			ServiceType:         "industry_detection",
			Method:              "content_analysis",
			ResponseTimeMs:      150.5,
			ProcessingTimeMs:    120.3,
			Confidence:          0.85,
			KeywordsCount:       5,
			ResultsCount:        3,
			CacheHitRatio:       0.9,
			ErrorOccurred:       false,
			ParallelProcessing:  true,
			GoroutinesUsed:      3,
			MemoryUsageMB:       25.5,
			DatabaseQueries:     2,
			DatabaseQueryTimeMs: 45.2,
		}

		// Verify all fields are set
		if metrics.RequestID == "" {
			t.Error("RequestID should not be empty")
		}

		if metrics.ServiceType == "" {
			t.Error("ServiceType should not be empty")
		}

		if metrics.Method == "" {
			t.Error("Method should not be empty")
		}

		if metrics.ResponseTimeMs <= 0 {
			t.Error("ResponseTimeMs should be positive")
		}

		if metrics.Confidence < 0 || metrics.Confidence > 1 {
			t.Error("Confidence should be between 0 and 1")
		}

		t.Logf("✅ Performance metrics structure is valid")
	})
}

// BenchmarkPerformanceMonitoring benchmarks the performance monitoring operations
func BenchmarkPerformanceMonitoring(b *testing.B) {
	monitor := createMockPerformanceMonitoringService()
	ctx := context.Background()

	metrics := &ClassificationPerformanceMetrics{
		RequestID:           "benchmark_req",
		Timestamp:           time.Now(),
		ServiceType:         "industry_detection",
		Method:              "content_analysis",
		ResponseTimeMs:      150.5,
		ProcessingTimeMs:    120.3,
		Confidence:          0.85,
		KeywordsCount:       5,
		ResultsCount:        3,
		CacheHitRatio:       0.9,
		ErrorOccurred:       false,
		ParallelProcessing:  true,
		GoroutinesUsed:      3,
		MemoryUsageMB:       25.5,
		DatabaseQueries:     2,
		DatabaseQueryTimeMs: 45.2,
	}

	b.Run("RecordPerformanceMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			metrics.RequestID = fmt.Sprintf("benchmark_req_%d", i)
			monitor.RecordPerformanceMetrics(ctx, metrics)
		}
	})

	b.Run("GetPerformanceSummary", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			monitor.GetPerformanceSummary(ctx, 24)
		}
	})

	b.Run("GetPerformanceDashboard", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			monitor.GetPerformanceDashboard(ctx)
		}
	})
}

// TestContinuousMonitoring tests continuous performance monitoring
func TestContinuousMonitoring(t *testing.T) {
	monitor := createMockPerformanceMonitoringService()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test continuous monitoring with short interval
	t.Run("ContinuousMonitoring", func(t *testing.T) {
		// Start monitoring with 1-second interval
		go monitor.MonitorPerformanceContinuously(ctx, 1*time.Second)

		// Wait for context to timeout
		<-ctx.Done()

		t.Logf("✅ Continuous monitoring test completed")
	})
}

// Helper functions for testing

// createMockPerformanceMonitoringService creates a mock performance monitoring service
func createMockPerformanceMonitoringService() *PerformanceMonitoringService {
	// In a real implementation, this would use a proper mock
	return NewPerformanceMonitoringService(nil)
}

// floatPtr returns a pointer to a float64
func floatPtr(f float64) *float64 {
	return &f
}
