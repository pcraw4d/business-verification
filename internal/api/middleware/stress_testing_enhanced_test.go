package middleware

import (
	"context"
	"net/http"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestEnhancedStressTester(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultEnhancedStressTestConfig()

	// Short duration and lower limits for testing
	config.MaxDuration = 5 * time.Second
	config.StartRPS = 5
	config.MaxRPS = 20
	config.StepSize = 5
	config.StepDuration = 1 * time.Second

	tester := NewEnhancedStressTester(config, logger)
	defer tester.Shutdown()

	t.Run("initialization", func(t *testing.T) {
		if tester.config == nil {
			t.Error("expected config to be set")
		}
		if tester.executor == nil {
			t.Error("expected executor to be initialized")
		}
		if tester.monitor == nil {
			t.Error("expected monitor to be initialized")
		}
		if tester.reporter == nil {
			t.Error("expected reporter to be initialized")
		}
		if tester.breakPoint == nil {
			t.Error("expected breaking point detector to be initialized")
		}
	})

	t.Run("configuration validation", func(t *testing.T) {
		// Test valid configuration
		err := tester.validateConfig()
		if err != nil {
			t.Errorf("expected valid config to pass validation, got error: %v", err)
		}

		// Test invalid configuration
		invalidConfig := &EnhancedStressTestConfig{
			MaxDuration: -1 * time.Second, // Invalid
			StartRPS:    0,                // Invalid
			MaxRPS:      5,                // Invalid (less than StartRPS)
			StepSize:    0,                // Invalid
			BaseURL:     "",               // Invalid
		}
		invalidTester := NewEnhancedStressTester(invalidConfig, logger)

		err = invalidTester.validateConfig()
		if err == nil {
			t.Error("expected validation error for invalid config")
		}
	})

	t.Run("default endpoints", func(t *testing.T) {
		endpoints := tester.getDefaultEndpoints()
		if len(endpoints) == 0 {
			t.Error("expected default endpoints to be generated")
		}

		// Check endpoint weights sum to 1.0
		totalWeight := 0.0
		for _, endpoint := range endpoints {
			totalWeight += endpoint.Weight
		}
		if totalWeight != 1.0 {
			t.Errorf("expected total weight to be 1.0, got %f", totalWeight)
		}

		// Check for critical endpoints
		foundCritical := false
		for _, endpoint := range endpoints {
			if endpoint.Critical {
				foundCritical = true
				break
			}
		}
		if !foundCritical {
			t.Error("expected at least one critical endpoint")
		}
	})

	t.Run("resource limits validation", func(t *testing.T) {
		limits := config.ResourceThreshold

		if limits.MaxCPUUsage <= 0 || limits.MaxCPUUsage > 100 {
			t.Error("max CPU usage should be between 0 and 100")
		}
		if limits.MaxMemoryUsage <= 0 {
			t.Error("max memory usage should be positive")
		}
		if limits.MaxDiskUsage <= 0 || limits.MaxDiskUsage > 100 {
			t.Error("max disk usage should be between 0 and 100")
		}
		if limits.MaxNetworkUsage <= 0 {
			t.Error("max network usage should be positive")
		}
	})
}

func TestBreakingPointDetector(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultEnhancedStressTestConfig()
	detector := NewBreakingPointDetector(config, logger)

	t.Run("threshold detection", func(t *testing.T) {
		// Test error rate threshold
		breakingPoint := detector.CheckBreakingPoint(0.15, 1*time.Second, 50.0, 100*1024*1024)
		if !breakingPoint {
			t.Error("expected breaking point detection for high error rate")
		}

		// Test response time threshold
		breakingPoint = detector.CheckBreakingPoint(0.05, 10*time.Second, 50.0, 100*1024*1024)
		if !breakingPoint {
			t.Error("expected breaking point detection for high response time")
		}

		// Test CPU threshold
		breakingPoint = detector.CheckBreakingPoint(0.05, 1*time.Second, 95.0, 100*1024*1024)
		if !breakingPoint {
			t.Error("expected breaking point detection for high CPU usage")
		}

		// Test memory threshold
		breakingPoint = detector.CheckBreakingPoint(0.05, 1*time.Second, 50.0, 3*1024*1024*1024)
		if !breakingPoint {
			t.Error("expected breaking point detection for high memory usage")
		}

		// Test normal conditions
		breakingPoint = detector.CheckBreakingPoint(0.05, 1*time.Second, 50.0, 100*1024*1024)
		if breakingPoint {
			t.Error("expected no breaking point detection for normal conditions")
		}
	})

	t.Run("trigger reason", func(t *testing.T) {
		reason := detector.GetTriggerReason()
		if reason == "" {
			t.Error("expected trigger reason to be set")
		}
	})
}

func TestRecoveryTester(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultEnhancedStressTestConfig()
	config.RecoveryDuration = 2 * time.Second
	config.RecoverySteps = 2

	client := &http.Client{Timeout: 5 * time.Second}
	tester := NewRecoveryTester(config, logger, client)

	t.Run("recovery test execution", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		metrics, err := tester.TestRecovery(ctx)
		if err != nil {
			t.Errorf("recovery test failed: %v", err)
		}

		if metrics == nil {
			t.Error("expected recovery metrics to be returned")
		}

		if len(metrics.RecoverySteps) != config.RecoverySteps {
			t.Errorf("expected %d recovery steps, got %d", config.RecoverySteps, len(metrics.RecoverySteps))
		}

		// Check recovery step progression
		for i, step := range metrics.RecoverySteps {
			if step.Step != i+1 {
				t.Errorf("expected step %d, got %d", i+1, step.Step)
			}
			if step.ErrorRate < 0 || step.ErrorRate > 1 {
				t.Errorf("error rate should be between 0 and 1, got %f", step.ErrorRate)
			}
		}
	})

	t.Run("recovery step execution", func(t *testing.T) {
		step := tester.executeRecoveryStep(1, 1*time.Second)

		if step.Step != 1 {
			t.Errorf("expected step 1, got %d", step.Step)
		}
		if step.ErrorRate < 0 || step.ErrorRate > 1 {
			t.Errorf("error rate should be between 0 and 1, got %f", step.ErrorRate)
		}
		if step.AvgResponseTime <= 0 {
			t.Error("average response time should be positive")
		}
		if step.RPS <= 0 {
			t.Error("RPS should be positive")
		}
	})
}

func TestStressTestMetrics(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultEnhancedStressTestConfig()
	tester := NewEnhancedStressTester(config, logger)

	t.Run("metrics calculation", func(t *testing.T) {
		metrics := &StressTestMetrics{
			ResponseTimes: []time.Duration{
				100 * time.Millisecond,
				200 * time.Millisecond,
				300 * time.Millisecond,
				400 * time.Millisecond,
				500 * time.Millisecond,
			},
			TotalRequests:      100,
			SuccessfulRequests: 90,
			FailedRequests:     10,
			ResourceSnapshots: []ResourceSnapshot{
				{CPUUsage: 50.0, MemoryUsage: 100 * 1024 * 1024},
				{CPUUsage: 70.0, MemoryUsage: 200 * 1024 * 1024},
				{CPUUsage: 90.0, MemoryUsage: 300 * 1024 * 1024},
			},
		}

		tester.calculateFinalStressMetrics(metrics)

		// Check error rate calculation
		expectedErrorRate := 0.1 // 10% error rate
		if metrics.ErrorRate != expectedErrorRate {
			t.Errorf("expected error rate %f, got %f", expectedErrorRate, metrics.ErrorRate)
		}

		// Check percentiles (using the same calculation logic as the implementation)
		if metrics.P50ResponseTime != 300*time.Millisecond {
			t.Errorf("expected P50 %v, got %v", 300*time.Millisecond, metrics.P50ResponseTime)
		}
		if metrics.P95ResponseTime != 400*time.Millisecond {
			t.Errorf("expected P95 %v, got %v", 400*time.Millisecond, metrics.P95ResponseTime)
		}

		// Check peak resource usage
		expectedPeakCPU := 90.0
		if metrics.PeakCPUUsage != expectedPeakCPU {
			t.Errorf("expected peak CPU %f, got %f", expectedPeakCPU, metrics.PeakCPUUsage)
		}

		expectedPeakMemory := uint64(300 * 1024 * 1024)
		if metrics.PeakMemoryUsage != expectedPeakMemory {
			t.Errorf("expected peak memory %d, got %d", expectedPeakMemory, metrics.PeakMemoryUsage)
		}
	})
}

func TestStressTestSummary(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultEnhancedStressTestConfig()
	tester := NewEnhancedStressTester(config, logger)

	t.Run("summary generation with breaking point", func(t *testing.T) {
		metrics := &StressTestMetrics{
			BreakingPoint: &BreakingPoint{
				RPS:             100,
				TriggerReason:   "error_rate_exceeded",
				ErrorRate:       0.15,
				AvgResponseTime: 2 * time.Second,
			},
			ErrorRate:    0.15,
			PeakCPUUsage: 85.0,
			RecoveryMetrics: &RecoveryMetrics{
				FullRecovery: true,
			},
		}

		summary := tester.generateStressSummary(metrics)

		if !summary.BreakingPointFound {
			t.Error("expected breaking point to be found")
		}
		if !summary.Success {
			t.Error("expected stress test to be successful when breaking point is found")
		}
		if summary.Status != "COMPLETED" {
			t.Errorf("expected status COMPLETED, got %s", summary.Status)
		}
		if !summary.RecoverySuccessful {
			t.Error("expected recovery to be successful")
		}
		if len(summary.KeyFindings) == 0 {
			t.Error("expected key findings to be generated")
		}
	})

	t.Run("summary generation without breaking point", func(t *testing.T) {
		metrics := &StressTestMetrics{
			BreakingPoint: nil,
			ErrorRate:     0.05,
			PeakCPUUsage:  50.0,
		}

		summary := tester.generateStressSummary(metrics)

		if summary.BreakingPointFound {
			t.Error("expected no breaking point to be found")
		}
		if summary.Status != "NO_BREAKING_POINT" {
			t.Errorf("expected status NO_BREAKING_POINT, got %s", summary.Status)
		}
	})
}

func TestResilienceScore(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultEnhancedStressTestConfig()
	tester := NewEnhancedStressTester(config, logger)

	t.Run("resilience score calculation", func(t *testing.T) {
		testCases := []struct {
			name     string
			metrics  *StressTestMetrics
			minScore float64
			maxScore float64
		}{
			{
				name: "excellent resilience",
				metrics: &StressTestMetrics{
					BreakingPoint:   &BreakingPoint{RPS: 950}, // Close to max RPS
					ErrorRate:       0.02,                     // Low error rate
					PeakCPUUsage:    70.0,                     // Moderate CPU
					RecoveryMetrics: &RecoveryMetrics{FullRecovery: true},
				},
				minScore: 90.0,
				maxScore: 110.0, // Allow for bonus points from recovery
			},
			{
				name: "poor resilience",
				metrics: &StressTestMetrics{
					BreakingPoint:   &BreakingPoint{RPS: 100}, // Low breaking point
					ErrorRate:       0.15,                     // High error rate
					PeakCPUUsage:    95.0,                     // High CPU
					RecoveryMetrics: nil,                      // No recovery
				},
				minScore: 0.0,
				maxScore: 50.0,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				score := tester.calculateResilienceScore(tc.metrics)
				if score < tc.minScore || score > tc.maxScore {
					t.Errorf("expected score between %f and %f, got %f", tc.minScore, tc.maxScore, score)
				}
			})
		}
	})

	t.Run("resilience levels", func(t *testing.T) {
		testCases := []struct {
			score    float64
			expected string
		}{
			{95.0, "Excellent"},
			{85.0, "Good"},
			{75.0, "Fair"},
			{65.0, "Poor"},
			{45.0, "Critical"},
		}

		for _, tc := range testCases {
			level := tester.getResilienceLevel(tc.score)
			if level != tc.expected {
				t.Errorf("expected resilience level %s for score %f, got %s", tc.expected, tc.score, level)
			}
		}
	})
}

func TestStressTestRecommendations(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultEnhancedStressTestConfig()
	tester := NewEnhancedStressTester(config, logger)

	t.Run("recommendation generation", func(t *testing.T) {
		metrics := &StressTestMetrics{
			BreakingPoint: &BreakingPoint{
				RPS: 50,
			},
			ErrorRate:       0.15,                   // High error rate
			PeakCPUUsage:    85.0,                   // High CPU
			PeakMemoryUsage: 2 * 1024 * 1024 * 1024, // High memory
			RecoveryMetrics: &RecoveryMetrics{
				FullRecovery: false, // Poor recovery
			},
		}

		recommendations := tester.generateStressRecommendations(metrics)

		if len(recommendations) == 0 {
			t.Error("expected recommendations to be generated")
		}

		// Check for specific recommendations
		foundBreakingPointRec := false
		foundErrorRateRec := false
		foundCPURec := false
		foundMemoryRec := false
		foundRecoveryRec := false

		for _, rec := range recommendations {
			if rec == "System breaking point found at 50 RPS. Consider optimizations to increase capacity" {
				foundBreakingPointRec = true
			}
			if rec == "Investigate and fix error sources during high load" {
				foundErrorRateRec = true
			}
			if rec == "Consider scaling out to reduce CPU utilization during peak load" {
				foundCPURec = true
			}
			if rec == "Optimize memory usage to handle higher loads" {
				foundMemoryRec = true
			}
			if rec == "Improve system recovery mechanisms after stress events" {
				foundRecoveryRec = true
			}
		}

		if !foundBreakingPointRec {
			t.Error("expected breaking point recommendation")
		}
		if !foundErrorRateRec {
			t.Error("expected error rate recommendation")
		}
		if !foundCPURec {
			t.Error("expected CPU recommendation")
		}
		if !foundMemoryRec {
			t.Error("expected memory recommendation")
		}
		if !foundRecoveryRec {
			t.Error("expected recovery recommendation")
		}
	})
}

func TestAlertTypes(t *testing.T) {
	t.Run("alert levels", func(t *testing.T) {
		levels := []StressAlertLevel{
			StressAlertLevelInfo,
			StressAlertLevelWarning,
			StressAlertLevelCritical,
			StressAlertLevelEmergency,
		}

		for _, level := range levels {
			if string(level) == "" {
				t.Errorf("alert level should not be empty: %v", level)
			}
		}
	})

	t.Run("alert types", func(t *testing.T) {
		types := []StressAlertType{
			StressAlertTypePerformance,
			StressAlertTypeResource,
			StressAlertTypeError,
			StressAlertTypeBreakingPoint,
			StressAlertTypeRecovery,
		}

		for _, alertType := range types {
			if string(alertType) == "" {
				t.Errorf("alert type should not be empty: %v", alertType)
			}
		}
	})
}

func TestStressEndpoint(t *testing.T) {
	t.Run("endpoint validation", func(t *testing.T) {
		endpoint := StressEndpoint{
			Name:     "test_endpoint",
			Method:   "POST",
			Path:     "/api/test",
			Headers:  map[string]string{"Content-Type": "application/json"},
			Body:     `{"test": "data"}`,
			Weight:   0.5,
			Critical: true,
		}

		if endpoint.Name == "" {
			t.Error("endpoint name should not be empty")
		}
		if endpoint.Method == "" {
			t.Error("endpoint method should not be empty")
		}
		if endpoint.Path == "" {
			t.Error("endpoint path should not be empty")
		}
		if endpoint.Weight <= 0 {
			t.Error("endpoint weight should be positive")
		}
	})
}

func TestEnvironmentInfoStress(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultEnhancedStressTestConfig()
	tester := NewEnhancedStressTester(config, logger)

	t.Run("environment information", func(t *testing.T) {
		env := tester.getEnvironmentInfo()

		if env["base_url"] != config.BaseURL {
			t.Errorf("expected base_url %s, got %s", config.BaseURL, env["base_url"])
		}
		if env["test_type"] != "enhanced_stress_test" {
			t.Errorf("expected test_type to be enhanced_stress_test, got %s", env["test_type"])
		}
		if env["max_rps"] != "1000" {
			t.Errorf("expected max_rps to be 1000, got %s", env["max_rps"])
		}
		if env["step_size"] != "50" {
			t.Errorf("expected step_size to be 50, got %s", env["step_size"])
		}
	})
}

func BenchmarkEnhancedStressTester_CalculateResilienceScore(b *testing.B) {
	logger := zap.NewNop()
	config := DefaultEnhancedStressTestConfig()
	tester := NewEnhancedStressTester(config, logger)

	metrics := &StressTestMetrics{
		BreakingPoint:   &BreakingPoint{RPS: 500},
		ErrorRate:       0.10,
		PeakCPUUsage:    85.0,
		RecoveryMetrics: &RecoveryMetrics{FullRecovery: true},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tester.calculateResilienceScore(metrics)
	}
}

func BenchmarkBreakingPointDetector_CheckBreakingPoint(b *testing.B) {
	logger := zap.NewNop()
	config := DefaultEnhancedStressTestConfig()
	detector := NewBreakingPointDetector(config, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.CheckBreakingPoint(0.05, 1*time.Second, 50.0, 100*1024*1024)
	}
}
