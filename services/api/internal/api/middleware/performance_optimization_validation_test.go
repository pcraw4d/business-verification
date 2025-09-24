package middleware

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestPerformanceOptimizationValidator_RegisterOptimization(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultOptimizationValidationConfig()
	pov := NewPerformanceOptimizationValidator(config, logger)

	tests := []struct {
		name          string
		optimization  *OptimizationResult
		expectedError string
	}{
		{
			name: "valid optimization",
			optimization: &OptimizationResult{
				ID:               "opt-123",
				OptimizationType: "cache_optimization",
				Description:      "Redis cache implementation",
				AppliedAt:        time.Now(),
				BaselineMetrics: ValidationPerformanceMetric{
					ResponseTime: 200 * time.Millisecond,
					Throughput:   100.0,
					ErrorRate:    2.0,
				},
				ExpectedImpact: ExpectedImpact{
					ResponseTimeReduction: 20.0,
					ThroughputIncrease:    15.0,
					Confidence:            0.85,
				},
			},
		},
		{
			name:          "nil optimization",
			optimization:  nil,
			expectedError: "optimization cannot be nil",
		},
		{
			name: "missing ID",
			optimization: &OptimizationResult{
				OptimizationType: "test",
			},
			expectedError: "optimization ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pov.RegisterOptimization(context.Background(), tt.optimization)

			if tt.expectedError != "" {
				if err == nil || err.Error() != tt.expectedError {
					t.Errorf("expected error %q, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Verify optimization was stored
				stored, err := pov.GetOptimization(tt.optimization.ID)
				if err != nil {
					t.Errorf("failed to retrieve stored optimization: %v", err)
				}
				if stored.ID != tt.optimization.ID {
					t.Errorf("stored optimization ID mismatch: got %s, want %s", stored.ID, tt.optimization.ID)
				}
			}
		})
	}
}

func TestPerformanceOptimizationValidator_ValidateOptimization(t *testing.T) {
	logger := zap.NewNop()
	config := &OptimizationValidationConfig{
		MinSampleSize:              5,
		MinResponseTimeImprovement: 10.0,
		MinThroughputImprovement:   5.0,
		MaxErrorRateIncrease:       2.0,
		StatisticalSignificance:    0.05,
		MaxPerformanceVariability:  15.0,
		ValidationRetentionDays:    30,
	}
	pov := NewPerformanceOptimizationValidator(config, logger)

	// Register test optimization
	optimization := &OptimizationResult{
		ID:               "opt-test-123",
		OptimizationType: "response_time_optimization",
		Description:      "Algorithm optimization",
		AppliedAt:        time.Now().Add(-1 * time.Hour),
		BaselineMetrics: ValidationPerformanceMetric{
			ResponseTime: 200 * time.Millisecond,
			Throughput:   100.0,
			ErrorRate:    2.0,
			CPUUsage:     70.0,
			MemoryUsage:  60.0,
		},
		ExpectedImpact: ExpectedImpact{
			ResponseTimeReduction: 15.0,
			ThroughputIncrease:    10.0,
			Confidence:            0.8,
		},
	}

	err := pov.RegisterOptimization(context.Background(), optimization)
	if err != nil {
		t.Fatalf("failed to register optimization: %v", err)
	}

	tests := []struct {
		name            string
		optimizationID  string
		currentMetrics  []ValidationPerformanceMetric
		expectedSuccess bool
		expectedError   string
	}{
		{
			name:           "successful optimization validation",
			optimizationID: "opt-test-123",
			currentMetrics: []ValidationPerformanceMetric{
				{ResponseTime: 160 * time.Millisecond, Throughput: 110.0, ErrorRate: 1.8, CPUUsage: 65.0, MemoryUsage: 55.0, Timestamp: time.Now()},
				{ResponseTime: 155 * time.Millisecond, Throughput: 112.0, ErrorRate: 1.9, CPUUsage: 66.0, MemoryUsage: 54.0, Timestamp: time.Now()},
				{ResponseTime: 165 * time.Millisecond, Throughput: 108.0, ErrorRate: 2.0, CPUUsage: 64.0, MemoryUsage: 56.0, Timestamp: time.Now()},
				{ResponseTime: 158 * time.Millisecond, Throughput: 109.0, ErrorRate: 1.7, CPUUsage: 67.0, MemoryUsage: 55.0, Timestamp: time.Now()},
				{ResponseTime: 162 * time.Millisecond, Throughput: 111.0, ErrorRate: 1.8, CPUUsage: 65.0, MemoryUsage: 54.0, Timestamp: time.Now()},
				{ResponseTime: 159 * time.Millisecond, Throughput: 110.5, ErrorRate: 1.8, CPUUsage: 65.5, MemoryUsage: 54.5, Timestamp: time.Now()},
				{ResponseTime: 161 * time.Millisecond, Throughput: 110.2, ErrorRate: 1.7, CPUUsage: 66.2, MemoryUsage: 55.2, Timestamp: time.Now()},
				{ResponseTime: 157 * time.Millisecond, Throughput: 111.8, ErrorRate: 1.9, CPUUsage: 64.8, MemoryUsage: 54.8, Timestamp: time.Now()},
				{ResponseTime: 163 * time.Millisecond, Throughput: 109.5, ErrorRate: 1.8, CPUUsage: 66.5, MemoryUsage: 55.5, Timestamp: time.Now()},
				{ResponseTime: 160 * time.Millisecond, Throughput: 110.8, ErrorRate: 1.7, CPUUsage: 65.2, MemoryUsage: 54.2, Timestamp: time.Now()},
				{ResponseTime: 158 * time.Millisecond, Throughput: 111.2, ErrorRate: 1.8, CPUUsage: 64.8, MemoryUsage: 55.8, Timestamp: time.Now()},
				{ResponseTime: 164 * time.Millisecond, Throughput: 109.8, ErrorRate: 1.9, CPUUsage: 66.8, MemoryUsage: 54.8, Timestamp: time.Now()},
			},
			expectedSuccess: true,
		},
		{
			name:           "optimization with regression",
			optimizationID: "opt-test-123",
			currentMetrics: []ValidationPerformanceMetric{
				{ResponseTime: 250 * time.Millisecond, Throughput: 90.0, ErrorRate: 3.5, CPUUsage: 80.0, MemoryUsage: 70.0, Timestamp: time.Now()},
				{ResponseTime: 240 * time.Millisecond, Throughput: 92.0, ErrorRate: 3.2, CPUUsage: 82.0, MemoryUsage: 72.0, Timestamp: time.Now()},
				{ResponseTime: 260 * time.Millisecond, Throughput: 88.0, ErrorRate: 3.8, CPUUsage: 78.0, MemoryUsage: 68.0, Timestamp: time.Now()},
				{ResponseTime: 245 * time.Millisecond, Throughput: 91.0, ErrorRate: 3.4, CPUUsage: 81.0, MemoryUsage: 71.0, Timestamp: time.Now()},
				{ResponseTime: 255 * time.Millisecond, Throughput: 89.0, ErrorRate: 3.6, CPUUsage: 79.0, MemoryUsage: 69.0, Timestamp: time.Now()},
			},
			expectedSuccess: false,
		},
		{
			name:           "insufficient metrics",
			optimizationID: "opt-test-123",
			currentMetrics: []ValidationPerformanceMetric{
				{ResponseTime: 160 * time.Millisecond, Throughput: 110.0, ErrorRate: 1.8, Timestamp: time.Now()},
			},
			expectedError: "insufficient metrics samples: got 1, need 5",
		},
		{
			name:           "optimization not found",
			optimizationID: "nonexistent",
			currentMetrics: []ValidationPerformanceMetric{
				{ResponseTime: 160 * time.Millisecond, Throughput: 110.0, ErrorRate: 1.8, Timestamp: time.Now()},
				{ResponseTime: 155 * time.Millisecond, Throughput: 112.0, ErrorRate: 1.9, Timestamp: time.Now()},
				{ResponseTime: 165 * time.Millisecond, Throughput: 108.0, ErrorRate: 2.0, Timestamp: time.Now()},
				{ResponseTime: 158 * time.Millisecond, Throughput: 109.0, ErrorRate: 1.7, Timestamp: time.Now()},
				{ResponseTime: 162 * time.Millisecond, Throughput: 111.0, ErrorRate: 1.8, Timestamp: time.Now()},
			},
			expectedError: "optimization not found: nonexistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pov.ValidateOptimization(context.Background(), tt.optimizationID, tt.currentMetrics)

			if tt.expectedError != "" {
				if err == nil || err.Error() != tt.expectedError {
					t.Errorf("expected error %q, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("validation result should not be nil")
				return
			}

			if result.IsSuccess != tt.expectedSuccess {
				t.Errorf("expected success %v, got %v", tt.expectedSuccess, result.IsSuccess)
				t.Logf("Improvements: %+v", result.Improvements)
				t.Logf("Regressions: %+v", result.Regressions)
				t.Logf("Sustainability: %+v", result.Sustainability)
				t.Logf("ActualImpact: %+v", result.ActualImpact)
			}

			// Verify result structure
			if result.ID == "" {
				t.Error("validation result ID should not be empty")
			}
			if result.OptimizationID != tt.optimizationID {
				t.Errorf("optimization ID mismatch: got %s, want %s", result.OptimizationID, tt.optimizationID)
			}
			if result.ValidationCompleted == nil {
				t.Error("validation completed timestamp should be set")
			}

			// Verify recommendations are generated
			if len(result.Recommendations) == 0 {
				t.Error("recommendations should be generated")
			}
		})
	}
}

func TestPerformanceOptimizationValidator_GetValidationResult(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultOptimizationValidationConfig()
	config.MinSampleSize = 3
	pov := NewPerformanceOptimizationValidator(config, logger)

	// Create and store a validation result
	optimization := &OptimizationResult{
		ID:               "opt-get-test",
		OptimizationType: "test_optimization",
		AppliedAt:        time.Now(),
		BaselineMetrics: ValidationPerformanceMetric{
			ResponseTime: 100 * time.Millisecond,
			Throughput:   50.0,
			ErrorRate:    1.0,
		},
	}

	err := pov.RegisterOptimization(context.Background(), optimization)
	if err != nil {
		t.Fatalf("failed to register optimization: %v", err)
	}

	metrics := []ValidationPerformanceMetric{
		{ResponseTime: 80 * time.Millisecond, Throughput: 55.0, ErrorRate: 0.8, Timestamp: time.Now()},
		{ResponseTime: 85 * time.Millisecond, Throughput: 53.0, ErrorRate: 0.9, Timestamp: time.Now()},
		{ResponseTime: 78 * time.Millisecond, Throughput: 56.0, ErrorRate: 0.7, Timestamp: time.Now()},
	}

	result, err := pov.ValidateOptimization(context.Background(), optimization.ID, metrics)
	if err != nil {
		t.Fatalf("failed to validate optimization: %v", err)
	}

	tests := []struct {
		name          string
		validationID  string
		expectedError string
		shouldBeFound bool
	}{
		{
			name:          "existing validation result",
			validationID:  result.ID,
			shouldBeFound: true,
		},
		{
			name:          "nonexistent validation result",
			validationID:  "nonexistent",
			expectedError: "validation result not found: nonexistent",
		},
		{
			name:          "empty validation ID",
			validationID:  "",
			expectedError: "validation ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrieved, err := pov.GetValidationResult(tt.validationID)

			if tt.expectedError != "" {
				if err == nil || err.Error() != tt.expectedError {
					t.Errorf("expected error %q, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.shouldBeFound && retrieved == nil {
					t.Error("validation result should be found")
				}
				if retrieved != nil && retrieved.ID != tt.validationID {
					t.Errorf("validation ID mismatch: got %s, want %s", retrieved.ID, tt.validationID)
				}
			}
		})
	}
}

func TestPerformanceOptimizationValidator_GetOptimization(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultOptimizationValidationConfig()
	pov := NewPerformanceOptimizationValidator(config, logger)

	optimization := &OptimizationResult{
		ID:               "opt-get-123",
		OptimizationType: "test_optimization",
		Description:      "Test optimization",
		AppliedAt:        time.Now(),
	}

	err := pov.RegisterOptimization(context.Background(), optimization)
	if err != nil {
		t.Fatalf("failed to register optimization: %v", err)
	}

	tests := []struct {
		name           string
		optimizationID string
		expectedError  string
		shouldBeFound  bool
	}{
		{
			name:           "existing optimization",
			optimizationID: "opt-get-123",
			shouldBeFound:  true,
		},
		{
			name:           "nonexistent optimization",
			optimizationID: "nonexistent",
			expectedError:  "optimization not found: nonexistent",
		},
		{
			name:           "empty optimization ID",
			optimizationID: "",
			expectedError:  "optimization ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrieved, err := pov.GetOptimization(tt.optimizationID)

			if tt.expectedError != "" {
				if err == nil || err.Error() != tt.expectedError {
					t.Errorf("expected error %q, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.shouldBeFound && retrieved == nil {
					t.Error("optimization should be found")
				}
				if retrieved != nil && retrieved.ID != tt.optimizationID {
					t.Errorf("optimization ID mismatch: got %s, want %s", retrieved.ID, tt.optimizationID)
				}
			}
		})
	}
}

func TestPerformanceOptimizationValidator_ListValidationResults(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultOptimizationValidationConfig()
	config.MinSampleSize = 3
	pov := NewPerformanceOptimizationValidator(config, logger)

	// Register optimizations and create validation results
	optimization1 := &OptimizationResult{
		ID:               "opt-list-1",
		OptimizationType: "cache_optimization",
		AppliedAt:        time.Now().Add(-2 * time.Hour),
		BaselineMetrics: ValidationPerformanceMetric{
			ResponseTime: 100 * time.Millisecond,
			Throughput:   50.0,
		},
	}

	optimization2 := &OptimizationResult{
		ID:               "opt-list-2",
		OptimizationType: "database_optimization",
		AppliedAt:        time.Now().Add(-1 * time.Hour),
		BaselineMetrics: ValidationPerformanceMetric{
			ResponseTime: 150 * time.Millisecond,
			Throughput:   40.0,
		},
	}

	err := pov.RegisterOptimization(context.Background(), optimization1)
	if err != nil {
		t.Fatalf("failed to register optimization1: %v", err)
	}

	err = pov.RegisterOptimization(context.Background(), optimization2)
	if err != nil {
		t.Fatalf("failed to register optimization2: %v", err)
	}

	metrics := []ValidationPerformanceMetric{
		{ResponseTime: 80 * time.Millisecond, Throughput: 55.0, Timestamp: time.Now()},
		{ResponseTime: 85 * time.Millisecond, Throughput: 53.0, Timestamp: time.Now()},
		{ResponseTime: 78 * time.Millisecond, Throughput: 56.0, Timestamp: time.Now()},
	}

	// Create validation results
	_, err = pov.ValidateOptimization(context.Background(), optimization1.ID, metrics)
	if err != nil {
		t.Fatalf("failed to validate optimization1: %v", err)
	}

	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

	_, err = pov.ValidateOptimization(context.Background(), optimization2.ID, metrics)
	if err != nil {
		t.Fatalf("failed to validate optimization2: %v", err)
	}

	// Test listing
	results := pov.ListValidationResults()

	if len(results) != 2 {
		t.Errorf("expected 2 validation results, got %d", len(results))
	}

	// Verify results are sorted by validation time (newest first)
	if len(results) >= 2 {
		if results[0].ValidationStarted.Before(results[1].ValidationStarted) {
			t.Error("validation results should be sorted by validation time (newest first)")
		}
	}
}

func TestPerformanceOptimizationValidator_ListOptimizations(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultOptimizationValidationConfig()
	pov := NewPerformanceOptimizationValidator(config, logger)

	// Register optimizations with different applied times
	optimization1 := &OptimizationResult{
		ID:               "opt-list-test-1",
		OptimizationType: "cache_optimization",
		AppliedAt:        time.Now().Add(-2 * time.Hour),
	}

	optimization2 := &OptimizationResult{
		ID:               "opt-list-test-2",
		OptimizationType: "database_optimization",
		AppliedAt:        time.Now().Add(-1 * time.Hour),
	}

	err := pov.RegisterOptimization(context.Background(), optimization1)
	if err != nil {
		t.Fatalf("failed to register optimization1: %v", err)
	}

	err = pov.RegisterOptimization(context.Background(), optimization2)
	if err != nil {
		t.Fatalf("failed to register optimization2: %v", err)
	}

	// Test listing
	optimizations := pov.ListOptimizations()

	if len(optimizations) != 2 {
		t.Errorf("expected 2 optimizations, got %d", len(optimizations))
	}

	// Verify optimizations are sorted by applied time (newest first)
	if len(optimizations) >= 2 {
		if optimizations[0].AppliedAt.Before(optimizations[1].AppliedAt) {
			t.Error("optimizations should be sorted by applied time (newest first)")
		}
	}

	// Verify IDs
	foundIDs := make(map[string]bool)
	for _, opt := range optimizations {
		foundIDs[opt.ID] = true
	}

	if !foundIDs["opt-list-test-1"] || !foundIDs["opt-list-test-2"] {
		t.Error("not all registered optimizations were found in the list")
	}
}

func TestPerformanceOptimizationValidator_Cleanup(t *testing.T) {
	logger := zap.NewNop()
	config := &OptimizationValidationConfig{
		MinSampleSize:           3,
		ValidationRetentionDays: 1, // 1 day retention
	}
	pov := NewPerformanceOptimizationValidator(config, logger)

	// Register old optimization (should be cleaned up)
	oldOptimization := &OptimizationResult{
		ID:        "opt-old",
		AppliedAt: time.Now().AddDate(0, 0, -2), // 2 days ago
	}

	// Register recent optimization (should be kept)
	recentOptimization := &OptimizationResult{
		ID:        "opt-recent",
		AppliedAt: time.Now().Add(-1 * time.Hour),
		BaselineMetrics: ValidationPerformanceMetric{
			ResponseTime: 100 * time.Millisecond,
			Throughput:   50.0,
		},
	}

	err := pov.RegisterOptimization(context.Background(), oldOptimization)
	if err != nil {
		t.Fatalf("failed to register old optimization: %v", err)
	}

	err = pov.RegisterOptimization(context.Background(), recentOptimization)
	if err != nil {
		t.Fatalf("failed to register recent optimization: %v", err)
	}

	// Create validation for recent optimization
	metrics := []ValidationPerformanceMetric{
		{ResponseTime: 80 * time.Millisecond, Throughput: 55.0, Timestamp: time.Now()},
		{ResponseTime: 85 * time.Millisecond, Throughput: 53.0, Timestamp: time.Now()},
		{ResponseTime: 78 * time.Millisecond, Throughput: 56.0, Timestamp: time.Now()},
	}

	_, err = pov.ValidateOptimization(context.Background(), recentOptimization.ID, metrics)
	if err != nil {
		t.Fatalf("failed to validate recent optimization: %v", err)
	}

	// Verify both optimizations exist before cleanup
	optimizations := pov.ListOptimizations()
	if len(optimizations) != 2 {
		t.Errorf("expected 2 optimizations before cleanup, got %d", len(optimizations))
	}

	// Perform cleanup
	err = pov.Cleanup()
	if err != nil {
		t.Errorf("cleanup failed: %v", err)
	}

	// Verify old optimization was removed
	optimizations = pov.ListOptimizations()
	if len(optimizations) != 1 {
		t.Errorf("expected 1 optimization after cleanup, got %d", len(optimizations))
	}

	if optimizations[0].ID != "opt-recent" {
		t.Errorf("wrong optimization kept after cleanup: got %s, want opt-recent", optimizations[0].ID)
	}

	// Verify old optimization is no longer accessible
	_, err = pov.GetOptimization("opt-old")
	if err == nil {
		t.Error("old optimization should not be accessible after cleanup")
	}

	// Verify recent optimization is still accessible
	_, err = pov.GetOptimization("opt-recent")
	if err != nil {
		t.Errorf("recent optimization should still be accessible after cleanup: %v", err)
	}
}

func TestPerformanceOptimizationValidator_Shutdown(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultOptimizationValidationConfig()
	pov := NewPerformanceOptimizationValidator(config, logger)

	err := pov.Shutdown()
	if err != nil {
		t.Errorf("shutdown failed: %v", err)
	}

	// Verify shutdown channel is closed
	select {
	case <-pov.stopCh:
		// Channel is closed, expected
	default:
		t.Error("stop channel should be closed after shutdown")
	}
}

func TestCalculatePercentageChange(t *testing.T) {
	tests := []struct {
		name     string
		baseline float64
		current  float64
		expected float64
	}{
		{
			name:     "improvement",
			baseline: 100.0,
			current:  80.0,
			expected: -20.0, // 20% improvement (reduction)
		},
		{
			name:     "regression",
			baseline: 100.0,
			current:  120.0,
			expected: 20.0, // 20% regression (increase)
		},
		{
			name:     "no change",
			baseline: 100.0,
			current:  100.0,
			expected: 0.0,
		},
		{
			name:     "zero baseline",
			baseline: 0.0,
			current:  50.0,
			expected: 100.0, // Arbitrary large change
		},
		{
			name:     "both zero",
			baseline: 0.0,
			current:  0.0,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePercentageChange(tt.baseline, tt.current)
			if result != tt.expected {
				t.Errorf("calculatePercentageChange(%f, %f) = %f, want %f", tt.baseline, tt.current, result, tt.expected)
			}
		})
	}
}

func TestCalculateMean(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "normal values",
			values:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 3.0,
		},
		{
			name:     "single value",
			values:   []float64{42.0},
			expected: 42.0,
		},
		{
			name:     "empty slice",
			values:   []float64{},
			expected: 0.0,
		},
		{
			name:     "negative values",
			values:   []float64{-1.0, -2.0, -3.0},
			expected: -2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMean(tt.values)
			if result != tt.expected {
				t.Errorf("calculateMean(%v) = %f, want %f", tt.values, result, tt.expected)
			}
		})
	}
}

func TestDetermineSeverity(t *testing.T) {
	tests := []struct {
		name          string
		changePercent float64
		expected      string
	}{
		{
			name:          "critical positive",
			changePercent: 60.0,
			expected:      "critical",
		},
		{
			name:          "critical negative",
			changePercent: -75.0,
			expected:      "critical",
		},
		{
			name:          "high severity",
			changePercent: 25.0,
			expected:      "high",
		},
		{
			name:          "medium severity",
			changePercent: 15.0,
			expected:      "medium",
		},
		{
			name:          "low severity",
			changePercent: 5.0,
			expected:      "low",
		},
		{
			name:          "very low",
			changePercent: 1.0,
			expected:      "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineSeverity(tt.changePercent)
			if result != tt.expected {
				t.Errorf("determineSeverity(%f) = %s, want %s", tt.changePercent, result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkPerformanceOptimizationValidator_RegisterOptimization(b *testing.B) {
	logger := zap.NewNop()
	config := DefaultOptimizationValidationConfig()
	pov := NewPerformanceOptimizationValidator(config, logger)

	optimization := &OptimizationResult{
		ID:               "benchmark-opt",
		OptimizationType: "benchmark_optimization",
		AppliedAt:        time.Now(),
		BaselineMetrics: ValidationPerformanceMetric{
			ResponseTime: 100 * time.Millisecond,
			Throughput:   50.0,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		optimization.ID = fmt.Sprintf("benchmark-opt-%d", i)
		pov.RegisterOptimization(context.Background(), optimization)
	}
}

func BenchmarkPerformanceOptimizationValidator_ValidateOptimization(b *testing.B) {
	logger := zap.NewNop()
	config := &OptimizationValidationConfig{
		MinSampleSize:              10,
		MinResponseTimeImprovement: 10.0,
		MinThroughputImprovement:   5.0,
		MaxErrorRateIncrease:       2.0,
		StatisticalSignificance:    0.05,
		MaxPerformanceVariability:  15.0,
		ValidationRetentionDays:    30,
	}
	pov := NewPerformanceOptimizationValidator(config, logger)

	// Register optimization
	optimization := &OptimizationResult{
		ID:               "benchmark-validate-opt",
		OptimizationType: "benchmark_optimization",
		AppliedAt:        time.Now(),
		BaselineMetrics: ValidationPerformanceMetric{
			ResponseTime: 200 * time.Millisecond,
			Throughput:   100.0,
			ErrorRate:    2.0,
		},
	}

	pov.RegisterOptimization(context.Background(), optimization)

	// Create metrics
	metrics := make([]ValidationPerformanceMetric, 10)
	for i := range metrics {
		metrics[i] = ValidationPerformanceMetric{
			ResponseTime: time.Duration(160+i) * time.Millisecond,
			Throughput:   110.0 + float64(i),
			ErrorRate:    1.8 - float64(i)*0.1,
			Timestamp:    time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pov.ValidateOptimization(context.Background(), optimization.ID, metrics)
	}
}

func BenchmarkStatisticalCalculations(b *testing.B) {
	values := make([]float64, 100)
	for i := range values {
		values[i] = float64(i) + 50.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateMean(values)
		calculateStdDev(values)
		calculateTTestSingle(values, 75.0)
	}
}
