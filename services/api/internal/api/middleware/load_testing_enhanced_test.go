package middleware

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestEnhancedLoadTester(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultEnhancedLoadTestConfig()

	// Short duration for testing
	config.Duration = 5 * time.Second
	config.RampUp.Duration = 1 * time.Second
	config.SteadyState.Duration = 2 * time.Second
	config.RampDown.Duration = 1 * time.Second
	config.MaxUsers = 10

	tester := NewEnhancedLoadTester(config, logger)
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
	})

	t.Run("configuration validation", func(t *testing.T) {
		// Test invalid configuration
		invalidConfig := &EnhancedLoadTestConfig{
			Duration: -1 * time.Second, // Invalid
			MaxUsers: 0,                // Invalid
		}
		invalidTester := NewEnhancedLoadTester(invalidConfig, logger)

		err := invalidTester.validateConfig()
		if err == nil {
			t.Error("expected validation error for invalid config")
		}
	})

	t.Run("default scenarios", func(t *testing.T) {
		scenarios := tester.getDefaultScenarios()
		if len(scenarios) == 0 {
			t.Error("expected default scenarios to be generated")
		}

		// Check scenario weights sum to 1.0
		totalWeight := 0.0
		for _, scenario := range scenarios {
			totalWeight += scenario.Weight
		}
		if totalWeight != 1.0 {
			t.Errorf("expected total weight to be 1.0, got %f", totalWeight)
		}
	})

	t.Run("test phases", func(t *testing.T) {
		// Test individual phases
		metrics := &LoadTestMetrics{
			StartTime:        time.Now(),
			ScenarioMetrics:  make(map[string]*ScenarioMetrics),
			ErrorsByType:     make(map[string]int64),
			ErrorsByEndpoint: make(map[string]int64),
			TimeSeries:       make([]TimeSeriesPoint, 0),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Test ramp up phase
		err := tester.executeRampUp(ctx, metrics)
		if err != nil {
			t.Errorf("ramp up phase failed: %v", err)
		}

		// Test steady state phase
		err = tester.executeSteadyState(ctx, metrics)
		if err != nil {
			t.Errorf("steady state phase failed: %v", err)
		}

		// Test ramp down phase
		err = tester.executeRampDown(ctx, metrics)
		if err != nil {
			t.Errorf("ramp down phase failed: %v", err)
		}
	})

	t.Run("SLA compliance calculation", func(t *testing.T) {
		metrics := &LoadTestMetrics{
			TotalRequests:      1000,
			SuccessfulRequests: 950,
			FailedRequests:     50,
			ErrorRate:          0.05,
			AvgResponseTime:    300 * time.Millisecond,
			RequestsPerSecond:  100,
			Duration:           10 * time.Second,
		}

		compliance := tester.calculateSLACompliance(metrics)
		if compliance == nil {
			t.Error("expected SLA compliance to be calculated")
		}

		// Check availability calculation
		expectedAvailability := 95.0 // 95% success rate
		if !compliance.AvailabilityMet && expectedAvailability >= config.SLA.AvailabilityTarget {
			t.Error("expected availability to meet SLA target")
		}
	})

	t.Run("performance level calculation", func(t *testing.T) {
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
			level := tester.getPerformanceLevel(tc.score)
			if level != tc.expected {
				t.Errorf("expected performance level %s for score %f, got %s", tc.expected, tc.score, level)
			}
		}
	})
}

func TestLoadTestMetrics(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultEnhancedLoadTestConfig()
	tester := NewEnhancedLoadTester(config, logger)

	t.Run("metrics calculation", func(t *testing.T) {
		metrics := &LoadTestMetrics{
			ResponseTimes: []time.Duration{
				100 * time.Millisecond,
				200 * time.Millisecond,
				300 * time.Millisecond,
				400 * time.Millisecond,
				500 * time.Millisecond,
			},
			TotalRequests:      100,
			SuccessfulRequests: 95,
			FailedRequests:     5,
			Duration:           10 * time.Second,
		}

		tester.calculateFinalMetrics(metrics)

		// Check error rate calculation
		expectedErrorRate := 0.05
		if metrics.ErrorRate != expectedErrorRate {
			t.Errorf("expected error rate %f, got %f", expectedErrorRate, metrics.ErrorRate)
		}

		// Check RPS calculation
		expectedRPS := 10.0 // 100 requests in 10 seconds
		if metrics.RequestsPerSecond != expectedRPS {
			t.Errorf("expected RPS %f, got %f", expectedRPS, metrics.RequestsPerSecond)
		}

		// Check percentiles (using the same calculation logic as the implementation)
		if metrics.P50ResponseTime != 300*time.Millisecond {
			t.Errorf("expected P50 %v, got %v", 300*time.Millisecond, metrics.P50ResponseTime)
		}
		if metrics.P95ResponseTime != 400*time.Millisecond {
			t.Errorf("expected P95 %v, got %v", 400*time.Millisecond, metrics.P95ResponseTime)
		}
	})
}

func TestLoadTestPatterns(t *testing.T) {
	t.Run("load patterns", func(t *testing.T) {
		patterns := []LoadPattern{
			PatternConstant,
			PatternLinear,
			PatternExponential,
			PatternStep,
			PatternSpike,
			PatternWave,
		}

		for _, pattern := range patterns {
			if string(pattern) == "" {
				t.Errorf("load pattern should not be empty: %v", pattern)
			}
		}
	})
}

func TestTestScenarios(t *testing.T) {
	t.Run("scenario validation", func(t *testing.T) {
		scenario := TestScenario{
			Name:      "test_scenario",
			Weight:    0.5,
			Method:    "GET",
			Endpoint:  "/test",
			Headers:   map[string]string{"Accept": "application/json"},
			ThinkTime: 1 * time.Second,
		}

		if scenario.Name == "" {
			t.Error("scenario name should not be empty")
		}
		if scenario.Weight <= 0 {
			t.Error("scenario weight should be positive")
		}
		if scenario.Method == "" {
			t.Error("scenario method should not be empty")
		}
	})

	t.Run("assertion types", func(t *testing.T) {
		assertionTypes := []AssertionType{
			AssertionResponse,
			AssertionStatus,
			AssertionHeader,
			AssertionBody,
			AssertionTime,
		}

		for _, assertionType := range assertionTypes {
			if string(assertionType) == "" {
				t.Errorf("assertion type should not be empty: %v", assertionType)
			}
		}
	})
}

func TestLoadThresholds(t *testing.T) {
	t.Run("threshold validation", func(t *testing.T) {
		thresholds := LoadThresholds{
			MaxResponseTime:    2 * time.Second,
			MaxP95ResponseTime: 5 * time.Second,
			MaxP99ResponseTime: 10 * time.Second,
			MaxErrorRate:       0.05,
			MinThroughput:      50,
			MaxMemoryUsage:     1 * 1024 * 1024 * 1024, // 1GB
			MaxCPUUsage:        80.0,
		}

		if thresholds.MaxResponseTime <= 0 {
			t.Error("max response time should be positive")
		}
		if thresholds.MaxErrorRate < 0 || thresholds.MaxErrorRate > 1 {
			t.Error("max error rate should be between 0 and 1")
		}
		if thresholds.MinThroughput <= 0 {
			t.Error("min throughput should be positive")
		}
		if thresholds.MaxCPUUsage <= 0 || thresholds.MaxCPUUsage > 100 {
			t.Error("max CPU usage should be between 0 and 100")
		}
	})
}

func TestSLAConfig(t *testing.T) {
	t.Run("SLA validation", func(t *testing.T) {
		sla := SLAConfig{
			AvailabilityTarget: 99.9,
			ResponseTimeTarget: 500 * time.Millisecond,
			ThroughputTarget:   100,
			ErrorRateTarget:    0.01,
			UptimeTarget:       9*time.Minute + 54*time.Second,
		}

		if sla.AvailabilityTarget <= 0 || sla.AvailabilityTarget > 100 {
			t.Error("availability target should be between 0 and 100")
		}
		if sla.ResponseTimeTarget <= 0 {
			t.Error("response time target should be positive")
		}
		if sla.ThroughputTarget <= 0 {
			t.Error("throughput target should be positive")
		}
		if sla.ErrorRateTarget < 0 || sla.ErrorRateTarget > 1 {
			t.Error("error rate target should be between 0 and 1")
		}
	})
}

func TestEnvironmentInfo(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultEnhancedLoadTestConfig()
	tester := NewEnhancedLoadTester(config, logger)

	t.Run("environment information", func(t *testing.T) {
		env := tester.getEnvironmentInfo()

		if env["base_url"] != config.BaseURL {
			t.Errorf("expected base_url %s, got %s", config.BaseURL, env["base_url"])
		}
		if env["test_type"] != "enhanced_load_test" {
			t.Errorf("expected test_type to be enhanced_load_test, got %s", env["test_type"])
		}
	})
}

func TestRecommendationGeneration(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultEnhancedLoadTestConfig()
	tester := NewEnhancedLoadTester(config, logger)

	t.Run("recommendation generation", func(t *testing.T) {
		// High error rate scenario
		metrics := &LoadTestMetrics{
			ErrorRate:         0.15,              // 15% error rate (high)
			P95ResponseTime:   2 * time.Second,   // High response time
			RequestsPerSecond: 30,                // Low throughput
			CPUUsage:          85,                // High CPU
			MemoryUsage:       600 * 1024 * 1024, // High memory
		}

		recommendations := tester.generateRecommendations(metrics)

		if len(recommendations) == 0 {
			t.Error("expected recommendations to be generated for problematic metrics")
		}

		// Check for specific recommendations
		foundErrorRateRec := false
		foundResponseTimeRec := false
		foundThroughputRec := false
		foundCPURec := false
		foundMemoryRec := false

		for _, rec := range recommendations {
			if rec == "Investigate and fix error sources to reduce error rate" {
				foundErrorRateRec = true
			}
			if rec == "Optimize response time by improving database queries and caching" {
				foundResponseTimeRec = true
			}
			if rec == "Increase throughput by optimizing resource utilization" {
				foundThroughputRec = true
			}
			if rec == "Consider scaling out to reduce CPU utilization" {
				foundCPURec = true
			}
			if rec == "Optimize memory usage to prevent memory pressure" {
				foundMemoryRec = true
			}
		}

		if !foundErrorRateRec {
			t.Error("expected error rate recommendation")
		}
		if !foundResponseTimeRec {
			t.Error("expected response time recommendation")
		}
		if !foundThroughputRec {
			t.Error("expected throughput recommendation")
		}
		if !foundCPURec {
			t.Error("expected CPU recommendation")
		}
		if !foundMemoryRec {
			t.Error("expected memory recommendation")
		}
	})
}

func BenchmarkEnhancedLoadTester_CalculatePercentiles(b *testing.B) {
	logger := zap.NewNop()
	config := DefaultEnhancedLoadTestConfig()
	tester := NewEnhancedLoadTester(config, logger)

	// Create test response times
	responseTimes := make([]time.Duration, 1000)
	for i := 0; i < 1000; i++ {
		responseTimes[i] = time.Duration(i) * time.Millisecond
	}

	metrics := &LoadTestMetrics{
		ResponseTimes: responseTimes,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tester.calculateFinalMetrics(metrics)
	}
}

func BenchmarkEnhancedLoadTester_GenerateRecommendations(b *testing.B) {
	logger := zap.NewNop()
	config := DefaultEnhancedLoadTestConfig()
	tester := NewEnhancedLoadTester(config, logger)

	metrics := &LoadTestMetrics{
		ErrorRate:         0.15,
		P95ResponseTime:   2 * time.Second,
		RequestsPerSecond: 30,
		CPUUsage:          85,
		MemoryUsage:       600 * 1024 * 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tester.generateRecommendations(metrics)
	}
}
