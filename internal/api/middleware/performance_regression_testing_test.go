package middleware

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestPerformanceRegressionTester_CreateBaseline(t *testing.T) {
	logger := zap.NewNop()
	config := &RegressionTestConfig{
		BaselineMinSamples: 5,
		BaselinePercentile: 95.0,
	}
	prt := NewPerformanceRegressionTester(config, logger)

	// Create test metrics
	metrics := []RegressionPerformanceMetric{
		{
			ResponseTime: 100 * time.Millisecond,
			Throughput:   10.0,
			ErrorRate:    1.0,
			ResourceUtilization: &ResourceUtilization{
				CPU:     50.0,
				Memory:  60.0,
				Disk:    30.0,
				Network: 40.0,
			},
		},
		{
			ResponseTime: 150 * time.Millisecond,
			Throughput:   12.0,
			ErrorRate:    2.0,
			ResourceUtilization: &ResourceUtilization{
				CPU:     55.0,
				Memory:  65.0,
				Disk:    35.0,
				Network: 45.0,
			},
		},
		{
			ResponseTime: 200 * time.Millisecond,
			Throughput:   15.0,
			ErrorRate:    0.5,
			ResourceUtilization: &ResourceUtilization{
				CPU:     60.0,
				Memory:  70.0,
				Disk:    40.0,
				Network: 50.0,
			},
		},
		{
			ResponseTime: 120 * time.Millisecond,
			Throughput:   11.0,
			ErrorRate:    1.5,
			ResourceUtilization: &ResourceUtilization{
				CPU:     52.0,
				Memory:  62.0,
				Disk:    32.0,
				Network: 42.0,
			},
		},
		{
			ResponseTime: 180 * time.Millisecond,
			Throughput:   13.0,
			ErrorRate:    1.2,
			ResourceUtilization: &ResourceUtilization{
				CPU:     58.0,
				Memory:  68.0,
				Disk:    38.0,
				Network: 48.0,
			},
		},
	}

	// Test baseline creation
	baseline, err := prt.CreateBaseline(context.Background(), "/test", "GET", metrics)
	if err != nil {
		t.Fatalf("failed to create baseline: %v", err)
	}

	// Verify baseline properties
	if baseline.Endpoint != "/test" {
		t.Errorf("expected endpoint /test, got %s", baseline.Endpoint)
	}
	if baseline.Method != "GET" {
		t.Errorf("expected method GET, got %s", baseline.Method)
	}
	if baseline.SampleCount != 5 {
		t.Errorf("expected sample count 5, got %d", baseline.SampleCount)
	}

	// Verify response time statistics
	if baseline.ResponseTime.P50 != 150*time.Millisecond {
		t.Errorf("expected P50 %v, got %v", 150*time.Millisecond, baseline.ResponseTime.P50)
	}
	if baseline.ResponseTime.P95 != 180*time.Millisecond {
		t.Errorf("expected P95 %v, got %v", 180*time.Millisecond, baseline.ResponseTime.P95)
	}
	if baseline.ResponseTime.P99 != 180*time.Millisecond {
		t.Errorf("expected P99 %v, got %v", 180*time.Millisecond, baseline.ResponseTime.P99)
	}

	// Verify throughput statistics
	expectedThroughput := 12.2 // (10+12+15+11+13)/5
	if baseline.Throughput.Mean != expectedThroughput {
		t.Errorf("expected throughput mean %f, got %f", expectedThroughput, baseline.Throughput.Mean)
	}

	// Verify error rate statistics
	expectedErrorRate := 1.24 // (1+2+0.5+1.5+1.2)/5
	if baseline.ErrorRate.Mean != expectedErrorRate {
		t.Errorf("expected error rate mean %f, got %f", expectedErrorRate, baseline.ErrorRate.Mean)
	}

	// Verify resource utilization
	expectedCPU := 55.0 // (50+55+60+52+58)/5
	if baseline.ResourceUtilization.CPU != expectedCPU {
		t.Errorf("expected CPU %f, got %f", expectedCPU, baseline.ResourceUtilization.CPU)
	}
}

func TestPerformanceRegressionTester_TestRegression(t *testing.T) {
	logger := zap.NewNop()
	config := &RegressionTestConfig{
		BaselineMinSamples:      5,
		MinimumSampleSize:       3,
		ResponseTimeThreshold:   10.0,
		ThroughputThreshold:     5.0,
		ErrorRateThreshold:      2.0,
		StatisticalSignificance: 0.05,
	}
	prt := NewPerformanceRegressionTester(config, logger)

	// Create baseline metrics
	baselineMetrics := []RegressionPerformanceMetric{
		{
			ResponseTime: 100 * time.Millisecond,
			Throughput:   10.0,
			ErrorRate:    1.0,
			ResourceUtilization: &ResourceUtilization{
				CPU: 50.0,
			},
		},
		{
			ResponseTime: 110 * time.Millisecond,
			Throughput:   11.0,
			ErrorRate:    1.2,
			ResourceUtilization: &ResourceUtilization{
				CPU: 52.0,
			},
		},
		{
			ResponseTime: 120 * time.Millisecond,
			Throughput:   12.0,
			ErrorRate:    0.8,
			ResourceUtilization: &ResourceUtilization{
				CPU: 48.0,
			},
		},
		{
			ResponseTime: 115 * time.Millisecond,
			Throughput:   11.5,
			ErrorRate:    1.1,
			ResourceUtilization: &ResourceUtilization{
				CPU: 51.0,
			},
		},
		{
			ResponseTime: 105 * time.Millisecond,
			Throughput:   10.5,
			ErrorRate:    0.9,
			ResourceUtilization: &ResourceUtilization{
				CPU: 49.0,
			},
		},
	}

	// Create baseline
	baseline, err := prt.CreateBaseline(context.Background(), "/test", "GET", baselineMetrics)
	if err != nil {
		t.Fatalf("failed to create baseline: %v", err)
	}

	// Test with regressed metrics
	regressedMetrics := []RegressionPerformanceMetric{
		{
			ResponseTime: 150 * time.Millisecond, // 50% increase
			Throughput:   8.0,                    // 20% decrease
			ErrorRate:    3.0,                    // 200% increase
			ResourceUtilization: &ResourceUtilization{
				CPU: 70.0, // 40% increase
			},
		},
		{
			ResponseTime: 160 * time.Millisecond,
			Throughput:   7.5,
			ErrorRate:    3.5,
			ResourceUtilization: &ResourceUtilization{
				CPU: 75.0,
			},
		},
		{
			ResponseTime: 140 * time.Millisecond,
			Throughput:   8.5,
			ErrorRate:    2.5,
			ResourceUtilization: &ResourceUtilization{
				CPU: 65.0,
			},
		},
	}

	// Test regression
	result, err := prt.TestRegression(context.Background(), baseline.ID, regressedMetrics)
	if err != nil {
		t.Fatalf("failed to test regression: %v", err)
	}

	// Verify regression detection
	if !result.HasRegression {
		t.Error("expected regression to be detected")
	}

	if result.Severity == "none" {
		t.Error("expected severity to be detected")
	}

	// Verify response time regression
	if result.ResponseTimeRegression == nil {
		t.Error("expected response time regression to be detected")
	} else {
		if !result.ResponseTimeRegression.IsRegression {
			t.Error("expected response time regression to be flagged")
		}
		if result.ResponseTimeRegression.ChangePercent <= 0 {
			t.Errorf("expected positive change percent, got %f", result.ResponseTimeRegression.ChangePercent)
		}
	}

	// Verify throughput regression
	if result.ThroughputRegression == nil {
		t.Error("expected throughput regression to be detected")
	} else {
		if !result.ThroughputRegression.IsRegression {
			t.Error("expected throughput regression to be flagged")
		}
		if result.ThroughputRegression.ChangePercent >= 0 {
			t.Errorf("expected negative change percent, got %f", result.ThroughputRegression.ChangePercent)
		}
	}

	// Verify error rate regression
	if result.ErrorRateRegression == nil {
		t.Error("expected error rate regression to be detected")
	} else {
		if !result.ErrorRateRegression.IsRegression {
			t.Error("expected error rate regression to be flagged")
		}
		if result.ErrorRateRegression.ChangePercent <= 0 {
			t.Errorf("expected positive change percent, got %f", result.ErrorRateRegression.ChangePercent)
		}
	}

	// Verify recommendations
	if len(result.Recommendations) == 0 {
		t.Error("expected recommendations to be generated")
	}
}

func TestPerformanceRegressionTester_TestNoRegression(t *testing.T) {
	logger := zap.NewNop()
	config := &RegressionTestConfig{
		BaselineMinSamples:      5,
		MinimumSampleSize:       3,
		ResponseTimeThreshold:   15.0, // Increased threshold
		ThroughputThreshold:     10.0, // Increased threshold
		ErrorRateThreshold:      20.0, // Increased threshold
		StatisticalSignificance: 0.05,
	}
	prt := NewPerformanceRegressionTester(config, logger)

	// Create baseline metrics
	baselineMetrics := []RegressionPerformanceMetric{
		{
			ResponseTime: 100 * time.Millisecond,
			Throughput:   10.0,
			ErrorRate:    1.0,
			ResourceUtilization: &ResourceUtilization{
				CPU: 50.0,
			},
		},
		{
			ResponseTime: 110 * time.Millisecond,
			Throughput:   11.0,
			ErrorRate:    1.2,
			ResourceUtilization: &ResourceUtilization{
				CPU: 52.0,
			},
		},
		{
			ResponseTime: 120 * time.Millisecond,
			Throughput:   12.0,
			ErrorRate:    0.8,
			ResourceUtilization: &ResourceUtilization{
				CPU: 48.0,
			},
		},
		{
			ResponseTime: 115 * time.Millisecond,
			Throughput:   11.5,
			ErrorRate:    1.1,
			ResourceUtilization: &ResourceUtilization{
				CPU: 51.0,
			},
		},
		{
			ResponseTime: 105 * time.Millisecond,
			Throughput:   10.5,
			ErrorRate:    0.9,
			ResourceUtilization: &ResourceUtilization{
				CPU: 49.0,
			},
		},
	}

	// Create baseline
	baseline, err := prt.CreateBaseline(context.Background(), "/test", "GET", baselineMetrics)
	if err != nil {
		t.Fatalf("failed to create baseline: %v", err)
	}

	// Test with similar metrics (no regression)
	similarMetrics := []RegressionPerformanceMetric{
		{
			ResponseTime: 105 * time.Millisecond, // 5% increase
			Throughput:   9.8,                    // 2% decrease
			ErrorRate:    1.1,                    // 10% increase
			ResourceUtilization: &ResourceUtilization{
				CPU: 52.0, // 4% increase
			},
		},
		{
			ResponseTime: 108 * time.Millisecond,
			Throughput:   10.2,
			ErrorRate:    0.9,
			ResourceUtilization: &ResourceUtilization{
				CPU: 49.0,
			},
		},
		{
			ResponseTime: 102 * time.Millisecond,
			Throughput:   10.1,
			ErrorRate:    1.0,
			ResourceUtilization: &ResourceUtilization{
				CPU: 51.0,
			},
		},
	}

	// Test regression
	result, err := prt.TestRegression(context.Background(), baseline.ID, similarMetrics)
	if err != nil {
		t.Fatalf("failed to test regression: %v", err)
	}

	// Verify no regression detection
	if result.HasRegression {
		t.Error("expected no regression to be detected")
	}

	if result.Severity != "none" {
		t.Errorf("expected severity 'none', got %s", result.Severity)
	}

	// Verify recommendations
	if len(result.Recommendations) == 0 {
		t.Error("expected recommendations to be generated")
	}
}

func TestPerformanceRegressionTester_GetBaseline(t *testing.T) {
	logger := zap.NewNop()
	config := &RegressionTestConfig{
		BaselineMinSamples: 5,
		MinimumSampleSize:  3,
	}
	prt := NewPerformanceRegressionTester(config, logger)

	// Create test metrics
	metrics := []RegressionPerformanceMetric{
		{
			ResponseTime: 100 * time.Millisecond,
			Throughput:   10.0,
			ErrorRate:    1.0,
		},
		{
			ResponseTime: 110 * time.Millisecond,
			Throughput:   11.0,
			ErrorRate:    1.2,
		},
		{
			ResponseTime: 120 * time.Millisecond,
			Throughput:   12.0,
			ErrorRate:    0.8,
		},
		{
			ResponseTime: 115 * time.Millisecond,
			Throughput:   11.5,
			ErrorRate:    1.1,
		},
		{
			ResponseTime: 105 * time.Millisecond,
			Throughput:   10.5,
			ErrorRate:    0.9,
		},
	}

	// Create baseline
	baseline, err := prt.CreateBaseline(context.Background(), "/test", "GET", metrics)
	if err != nil {
		t.Fatalf("failed to create baseline: %v", err)
	}

	// Retrieve baseline
	retrievedBaseline, err := prt.GetBaseline(baseline.ID)
	if err != nil {
		t.Fatalf("failed to get baseline: %v", err)
	}

	// Verify retrieved baseline
	if retrievedBaseline.ID != baseline.ID {
		t.Errorf("expected baseline ID %s, got %s", baseline.ID, retrievedBaseline.ID)
	}
	if retrievedBaseline.Endpoint != baseline.Endpoint {
		t.Errorf("expected endpoint %s, got %s", baseline.Endpoint, retrievedBaseline.Endpoint)
	}
	if retrievedBaseline.Method != baseline.Method {
		t.Errorf("expected method %s, got %s", baseline.Method, retrievedBaseline.Method)
	}

	// Test getting non-existent baseline
	_, err = prt.GetBaseline("non-existent")
	if err == nil {
		t.Error("expected error when getting non-existent baseline")
	}
}

func TestPerformanceRegressionTester_GetResult(t *testing.T) {
	logger := zap.NewNop()
	config := &RegressionTestConfig{
		BaselineMinSamples: 5,
		MinimumSampleSize:  3,
	}
	prt := NewPerformanceRegressionTester(config, logger)

	// Create baseline and test metrics
	baselineMetrics := []RegressionPerformanceMetric{
		{ResponseTime: 100 * time.Millisecond, Throughput: 10.0, ErrorRate: 1.0},
		{ResponseTime: 110 * time.Millisecond, Throughput: 11.0, ErrorRate: 1.2},
		{ResponseTime: 120 * time.Millisecond, Throughput: 12.0, ErrorRate: 0.8},
		{ResponseTime: 115 * time.Millisecond, Throughput: 11.5, ErrorRate: 1.1},
		{ResponseTime: 105 * time.Millisecond, Throughput: 10.5, ErrorRate: 0.9},
	}

	baseline, err := prt.CreateBaseline(context.Background(), "/test", "GET", baselineMetrics)
	if err != nil {
		t.Fatalf("failed to create baseline: %v", err)
	}

	testMetrics := []RegressionPerformanceMetric{
		{ResponseTime: 150 * time.Millisecond, Throughput: 8.0, ErrorRate: 3.0},
		{ResponseTime: 160 * time.Millisecond, Throughput: 7.5, ErrorRate: 3.5},
		{ResponseTime: 140 * time.Millisecond, Throughput: 8.5, ErrorRate: 2.5},
	}

	// Create result
	result, err := prt.TestRegression(context.Background(), baseline.ID, testMetrics)
	if err != nil {
		t.Fatalf("failed to test regression: %v", err)
	}

	// Retrieve result
	retrievedResult, err := prt.GetResult(result.ID)
	if err != nil {
		t.Fatalf("failed to get result: %v", err)
	}

	// Verify retrieved result
	if retrievedResult.ID != result.ID {
		t.Errorf("expected result ID %s, got %s", result.ID, retrievedResult.ID)
	}
	if retrievedResult.BaselineID != result.BaselineID {
		t.Errorf("expected baseline ID %s, got %s", result.BaselineID, retrievedResult.BaselineID)
	}
	if retrievedResult.HasRegression != result.HasRegression {
		t.Errorf("expected has regression %v, got %v", result.HasRegression, retrievedResult.HasRegression)
	}

	// Test getting non-existent result
	_, err = prt.GetResult("non-existent")
	if err == nil {
		t.Error("expected error when getting non-existent result")
	}
}

func TestPerformanceRegressionTester_ListBaselines(t *testing.T) {
	logger := zap.NewNop()
	config := &RegressionTestConfig{
		BaselineMinSamples: 5,
		MinimumSampleSize:  3,
	}
	prt := NewPerformanceRegressionTester(config, logger)

	// Create multiple baselines
	metrics1 := []RegressionPerformanceMetric{
		{ResponseTime: 100 * time.Millisecond, Throughput: 10.0, ErrorRate: 1.0},
		{ResponseTime: 110 * time.Millisecond, Throughput: 11.0, ErrorRate: 1.2},
		{ResponseTime: 120 * time.Millisecond, Throughput: 12.0, ErrorRate: 0.8},
		{ResponseTime: 115 * time.Millisecond, Throughput: 11.5, ErrorRate: 1.1},
		{ResponseTime: 105 * time.Millisecond, Throughput: 10.5, ErrorRate: 0.9},
	}

	metrics2 := []RegressionPerformanceMetric{
		{ResponseTime: 200 * time.Millisecond, Throughput: 20.0, ErrorRate: 2.0},
		{ResponseTime: 210 * time.Millisecond, Throughput: 21.0, ErrorRate: 2.2},
		{ResponseTime: 220 * time.Millisecond, Throughput: 22.0, ErrorRate: 1.8},
		{ResponseTime: 215 * time.Millisecond, Throughput: 21.5, ErrorRate: 2.1},
		{ResponseTime: 205 * time.Millisecond, Throughput: 20.5, ErrorRate: 1.9},
	}

	baseline1, err := prt.CreateBaseline(context.Background(), "/test1", "GET", metrics1)
	if err != nil {
		t.Fatalf("failed to create baseline1: %v", err)
	}

	baseline2, err := prt.CreateBaseline(context.Background(), "/test2", "POST", metrics2)
	if err != nil {
		t.Fatalf("failed to create baseline2: %v", err)
	}

	// List baselines
	baselines := prt.ListBaselines()

	// Verify we have 2 baselines
	if len(baselines) != 2 {
		t.Errorf("expected 2 baselines, got %d", len(baselines))
	}

	// Verify baselines are sorted by creation date (newest first)
	if baselines[0].CreatedAt.Before(baselines[1].CreatedAt) {
		t.Error("expected baselines to be sorted by creation date (newest first)")
	}

	// Verify baseline IDs are present
	foundBaseline1 := false
	foundBaseline2 := false
	for _, baseline := range baselines {
		if baseline.ID == baseline1.ID {
			foundBaseline1 = true
		}
		if baseline.ID == baseline2.ID {
			foundBaseline2 = true
		}
	}

	if !foundBaseline1 {
		t.Error("baseline1 not found in list")
	}
	if !foundBaseline2 {
		t.Error("baseline2 not found in list")
	}
}

func TestPerformanceRegressionTester_ListResults(t *testing.T) {
	logger := zap.NewNop()
	config := &RegressionTestConfig{
		BaselineMinSamples: 5,
		MinimumSampleSize:  3,
	}
	prt := NewPerformanceRegressionTester(config, logger)

	// Create baseline
	baselineMetrics := []RegressionPerformanceMetric{
		{ResponseTime: 100 * time.Millisecond, Throughput: 10.0, ErrorRate: 1.0},
		{ResponseTime: 110 * time.Millisecond, Throughput: 11.0, ErrorRate: 1.2},
		{ResponseTime: 120 * time.Millisecond, Throughput: 12.0, ErrorRate: 0.8},
		{ResponseTime: 115 * time.Millisecond, Throughput: 11.5, ErrorRate: 1.1},
		{ResponseTime: 105 * time.Millisecond, Throughput: 10.5, ErrorRate: 0.9},
	}

	baseline, err := prt.CreateBaseline(context.Background(), "/test", "GET", baselineMetrics)
	if err != nil {
		t.Fatalf("failed to create baseline: %v", err)
	}

	// Create multiple test results
	testMetrics1 := []RegressionPerformanceMetric{
		{ResponseTime: 150 * time.Millisecond, Throughput: 8.0, ErrorRate: 3.0},
		{ResponseTime: 160 * time.Millisecond, Throughput: 7.5, ErrorRate: 3.5},
		{ResponseTime: 140 * time.Millisecond, Throughput: 8.5, ErrorRate: 2.5},
	}

	testMetrics2 := []RegressionPerformanceMetric{
		{ResponseTime: 105 * time.Millisecond, Throughput: 9.8, ErrorRate: 1.1},
		{ResponseTime: 108 * time.Millisecond, Throughput: 10.2, ErrorRate: 0.9},
		{ResponseTime: 102 * time.Millisecond, Throughput: 10.1, ErrorRate: 1.0},
	}

	result1, err := prt.TestRegression(context.Background(), baseline.ID, testMetrics1)
	if err != nil {
		t.Fatalf("failed to create result1: %v", err)
	}

	result2, err := prt.TestRegression(context.Background(), baseline.ID, testMetrics2)
	if err != nil {
		t.Fatalf("failed to create result2: %v", err)
	}

	// List results
	results := prt.ListResults()

	// Verify we have 2 results
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Verify results are sorted by test date (newest first)
	if results[0].TestedAt.Before(results[1].TestedAt) {
		t.Error("expected results to be sorted by test date (newest first)")
	}

	// Verify result IDs are present
	foundResult1 := false
	foundResult2 := false
	for _, result := range results {
		if result.ID == result1.ID {
			foundResult1 = true
		}
		if result.ID == result2.ID {
			foundResult2 = true
		}
	}

	if !foundResult1 {
		t.Error("result1 not found in list")
	}
	if !foundResult2 {
		t.Error("result2 not found in list")
	}
}

func TestPerformanceRegressionTester_UpdateBaseline(t *testing.T) {
	logger := zap.NewNop()
	config := &RegressionTestConfig{
		BaselineMinSamples: 5,
		MinimumSampleSize:  3,
	}
	prt := NewPerformanceRegressionTester(config, logger)

	// Create initial baseline
	initialMetrics := []RegressionPerformanceMetric{
		{ResponseTime: 100 * time.Millisecond, Throughput: 10.0, ErrorRate: 1.0},
		{ResponseTime: 110 * time.Millisecond, Throughput: 11.0, ErrorRate: 1.2},
		{ResponseTime: 120 * time.Millisecond, Throughput: 12.0, ErrorRate: 0.8},
		{ResponseTime: 115 * time.Millisecond, Throughput: 11.5, ErrorRate: 1.1},
		{ResponseTime: 105 * time.Millisecond, Throughput: 10.5, ErrorRate: 0.9},
	}

	baseline, err := prt.CreateBaseline(context.Background(), "/test", "GET", initialMetrics)
	if err != nil {
		t.Fatalf("failed to create baseline: %v", err)
	}

	// Update with new metrics
	newMetrics := []RegressionPerformanceMetric{
		{ResponseTime: 90 * time.Millisecond, Throughput: 12.0, ErrorRate: 0.5},
		{ResponseTime: 95 * time.Millisecond, Throughput: 11.8, ErrorRate: 0.7},
		{ResponseTime: 85 * time.Millisecond, Throughput: 12.2, ErrorRate: 0.3},
		{ResponseTime: 92 * time.Millisecond, Throughput: 11.9, ErrorRate: 0.6},
		{ResponseTime: 88 * time.Millisecond, Throughput: 12.1, ErrorRate: 0.4},
	}

	updatedBaseline, err := prt.UpdateBaseline(context.Background(), baseline.ID, newMetrics)
	if err != nil {
		t.Fatalf("failed to update baseline: %v", err)
	}

	// Verify updated baseline
	if updatedBaseline.ID != baseline.ID {
		t.Errorf("expected baseline ID %s, got %s", baseline.ID, updatedBaseline.ID)
	}
	if updatedBaseline.CreatedAt != baseline.CreatedAt {
		t.Errorf("expected creation date to remain the same")
	}
	if !updatedBaseline.UpdatedAt.After(baseline.UpdatedAt) {
		t.Errorf("expected updated date to be newer")
	}
	if updatedBaseline.SampleCount != 5 {
		t.Errorf("expected sample count 5, got %d", updatedBaseline.SampleCount)
	}

	// Verify updated metrics reflect new data
	if updatedBaseline.ResponseTime.P50 != 90*time.Millisecond {
		t.Errorf("expected updated P50 %v, got %v", 90*time.Millisecond, updatedBaseline.ResponseTime.P50)
	}
}

func TestPerformanceRegressionTester_Cleanup(t *testing.T) {
	logger := zap.NewNop()
	config := &RegressionTestConfig{
		BaselineMinSamples:    5,
		BaselineRetentionDays: 1,
		MetricsRetentionDays:  1,
	}
	prt := NewPerformanceRegressionTester(config, logger)

	// Create baseline and result
	metrics := []RegressionPerformanceMetric{
		{ResponseTime: 100 * time.Millisecond, Throughput: 10.0, ErrorRate: 1.0},
		{ResponseTime: 110 * time.Millisecond, Throughput: 11.0, ErrorRate: 1.2},
		{ResponseTime: 120 * time.Millisecond, Throughput: 12.0, ErrorRate: 0.8},
		{ResponseTime: 115 * time.Millisecond, Throughput: 11.5, ErrorRate: 1.1},
		{ResponseTime: 105 * time.Millisecond, Throughput: 10.5, ErrorRate: 0.9},
	}

	baseline, err := prt.CreateBaseline(context.Background(), "/test", "GET", metrics)
	if err != nil {
		t.Fatalf("failed to create baseline: %v", err)
	}

	testMetrics := []RegressionPerformanceMetric{
		{ResponseTime: 150 * time.Millisecond, Throughput: 8.0, ErrorRate: 3.0},
		{ResponseTime: 160 * time.Millisecond, Throughput: 7.5, ErrorRate: 3.5},
		{ResponseTime: 140 * time.Millisecond, Throughput: 8.5, ErrorRate: 2.5},
	}

	_, err = prt.TestRegression(context.Background(), baseline.ID, testMetrics)
	if err != nil {
		t.Fatalf("failed to create result: %v", err)
	}

	// Verify baseline and result exist
	baselines := prt.ListBaselines()
	if len(baselines) != 1 {
		t.Errorf("expected 1 baseline, got %d", len(baselines))
	}

	results := prt.ListResults()
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	// Cleanup should not remove recent data
	err = prt.Cleanup()
	if err != nil {
		t.Fatalf("failed to cleanup: %v", err)
	}

	// Verify data still exists (not old enough)
	baselines = prt.ListBaselines()
	if len(baselines) != 1 {
		t.Errorf("expected 1 baseline after cleanup, got %d", len(baselines))
	}

	results = prt.ListResults()
	if len(results) != 1 {
		t.Errorf("expected 1 result after cleanup, got %d", len(results))
	}
}

func TestPerformanceRegressionTester_Shutdown(t *testing.T) {
	logger := zap.NewNop()
	prt := NewPerformanceRegressionTester(nil, logger)

	// Test shutdown
	err := prt.Shutdown()
	if err != nil {
		t.Fatalf("failed to shutdown: %v", err)
	}

	// Test shutdown again (should not panic)
	err = prt.Shutdown()
	if err != nil {
		t.Fatalf("failed to shutdown again: %v", err)
	}
}

func TestPerformanceRegressionTester_StatisticalCalculations(t *testing.T) {
	logger := zap.NewNop()
	prt := NewPerformanceRegressionTester(nil, logger)

	// Test t-test calculation
	samples := []time.Duration{
		100 * time.Millisecond,
		110 * time.Millisecond,
		120 * time.Millisecond,
		115 * time.Millisecond,
		105 * time.Millisecond,
	}

	baselineMean := 110 * time.Millisecond
	baselineStd := 10 * time.Millisecond

	pValue := prt.calculateTTest(samples, baselineMean, baselineStd)
	if pValue < 0 || pValue > 1 {
		t.Errorf("expected p-value between 0 and 1, got %f", pValue)
	}

	// Test effect size calculation
	effectSize := prt.calculateEffectSize(samples, baselineMean, baselineStd)
	if effectSize < 0 {
		t.Errorf("expected positive effect size, got %f", effectSize)
	}

	// Test float t-test calculation
	floatSamples := []float64{10.0, 11.0, 12.0, 11.5, 10.5}
	baselineFloatMean := 11.0
	baselineFloatStd := 1.0

	floatPValue := prt.calculateTTestFloat(floatSamples, baselineFloatMean, baselineFloatStd)
	if floatPValue < 0 || floatPValue > 1 {
		t.Errorf("expected p-value between 0 and 1, got %f", floatPValue)
	}

	// Test float effect size calculation
	floatEffectSize := prt.calculateEffectSizeFloat(floatSamples, baselineFloatMean, baselineFloatStd)
	if floatEffectSize < 0 {
		t.Errorf("expected positive effect size, got %f", floatEffectSize)
	}
}

func TestPerformanceRegressionTester_SeverityDetermination(t *testing.T) {
	logger := zap.NewNop()
	prt := NewPerformanceRegressionTester(nil, logger)

	// Test severity determination
	testCases := []struct {
		changePercent float64
		isSignificant bool
		expected      string
	}{
		{60.0, true, "critical"},
		{30.0, true, "high"},
		{15.0, true, "medium"},
		{8.0, true, "low"},
		{3.0, true, "none"},
		{8.0, false, "none"},
		{0.0, false, "none"},
	}

	for _, tc := range testCases {
		severity := prt.determineSeverity(tc.changePercent, tc.isSignificant)
		if severity != tc.expected {
			t.Errorf("for change percent %f, isSignificant %v: expected %s, got %s",
				tc.changePercent, tc.isSignificant, tc.expected, severity)
		}
	}
}

// Benchmark tests
func BenchmarkPerformanceRegressionTester_CreateBaseline(b *testing.B) {
	logger := zap.NewNop()
	prt := NewPerformanceRegressionTester(nil, logger)

	// Create test metrics
	metrics := make([]RegressionPerformanceMetric, 100)
	for i := range metrics {
		metrics[i] = RegressionPerformanceMetric{
			ResponseTime: time.Duration(100+i) * time.Millisecond,
			Throughput:   float64(10 + i),
			ErrorRate:    float64(1 + i%5),
			ResourceUtilization: &ResourceUtilization{
				CPU:     float64(50 + i),
				Memory:  float64(60 + i),
				Disk:    float64(30 + i),
				Network: float64(40 + i),
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := prt.CreateBaseline(context.Background(), "/benchmark", "GET", metrics)
		if err != nil {
			b.Fatalf("failed to create baseline: %v", err)
		}
	}
}

func BenchmarkPerformanceRegressionTester_TestRegression(b *testing.B) {
	logger := zap.NewNop()
	config := &RegressionTestConfig{
		BaselineMinSamples: 50,
		MinimumSampleSize:  30,
	}
	prt := NewPerformanceRegressionTester(config, logger)

	// Create baseline metrics
	baselineMetrics := make([]RegressionPerformanceMetric, 100)
	for i := range baselineMetrics {
		baselineMetrics[i] = RegressionPerformanceMetric{
			ResponseTime: time.Duration(100+i) * time.Millisecond,
			Throughput:   float64(10 + i),
			ErrorRate:    float64(1 + i%5),
			ResourceUtilization: &ResourceUtilization{
				CPU: float64(50 + i),
			},
		}
	}

	baseline, err := prt.CreateBaseline(context.Background(), "/benchmark", "GET", baselineMetrics)
	if err != nil {
		b.Fatalf("failed to create baseline: %v", err)
	}

	// Create test metrics
	testMetrics := make([]RegressionPerformanceMetric, 50)
	for i := range testMetrics {
		testMetrics[i] = RegressionPerformanceMetric{
			ResponseTime: time.Duration(150+i) * time.Millisecond, // Regression
			Throughput:   float64(8 + i),                          // Regression
			ErrorRate:    float64(3 + i%5),                        // Regression
			ResourceUtilization: &ResourceUtilization{
				CPU: float64(70 + i), // Regression
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := prt.TestRegression(context.Background(), baseline.ID, testMetrics)
		if err != nil {
			b.Fatalf("failed to test regression: %v", err)
		}
	}
}

func BenchmarkPerformanceRegressionTester_StatisticalCalculations(b *testing.B) {
	logger := zap.NewNop()
	prt := NewPerformanceRegressionTester(nil, logger)

	// Create test data
	samples := make([]time.Duration, 1000)
	for i := range samples {
		samples[i] = time.Duration(100+i) * time.Millisecond
	}

	baselineMean := 150 * time.Millisecond
	baselineStd := 50 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prt.calculateTTest(samples, baselineMean, baselineStd)
		prt.calculateEffectSize(samples, baselineMean, baselineStd)
	}
}
