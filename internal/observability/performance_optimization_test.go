package observability

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewPerformanceOptimizationSystem(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{
		AnalysisInterval:         1 * time.Hour,
		RecommendationThreshold:  5.0,
		ConfidenceThreshold:      0.7,
		MaxRecommendations:       10,
		RecommendationExpiry:     7 * 24 * time.Hour,
		AutoPrioritization:       true,
		AutoImplementation:       false,
		ImplementationDelay:      24 * time.Hour,
		RollbackThreshold:        -10.0,
		MaxAnalysisDuration:      30 * time.Minute,
		MinDataPoints:            100,
		AnalysisWindow:           24 * time.Hour,
		EnableOptimizationAlerts: true,
		AlertSeverity:            make(map[string]string),
	}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}
	benchmarkingSystem := &PerformanceBenchmarkingSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}

	pos := NewPerformanceOptimizationSystem(performanceMonitor, regressionDetection, benchmarkingSystem, predictiveAnalytics, config, logger)

	assert.NotNil(t, pos)
	assert.Equal(t, performanceMonitor, pos.performanceMonitor)
	assert.Equal(t, regressionDetection, pos.regressionDetection)
	assert.Equal(t, benchmarkingSystem, pos.benchmarkingSystem)
	assert.Equal(t, predictiveAnalytics, pos.predictiveAnalytics)
	assert.Equal(t, config, pos.config)
	assert.NotNil(t, pos.recommendationEngine)
	assert.NotNil(t, pos.optimizationHistory)
	assert.NotNil(t, pos.implementedOptimizations)
}

func TestPerformanceOptimizationSystem_StartStop(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{
		AnalysisInterval: 100 * time.Millisecond, // Short interval for testing
	}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}
	benchmarkingSystem := &PerformanceBenchmarkingSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}

	pos := NewPerformanceOptimizationSystem(performanceMonitor, regressionDetection, benchmarkingSystem, predictiveAnalytics, config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the system
	err := pos.Start(ctx)
	assert.NoError(t, err)

	// Wait a bit for goroutines to start
	time.Sleep(50 * time.Millisecond)

	// Stop the system
	err = pos.Stop()
	assert.NoError(t, err)
}

func TestPerformanceOptimizationSystem_GetRecommendations(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}
	benchmarkingSystem := &PerformanceBenchmarkingSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}

	pos := NewPerformanceOptimizationSystem(performanceMonitor, regressionDetection, benchmarkingSystem, predictiveAnalytics, config, logger)

	// Initially no recommendations
	recommendations := pos.GetRecommendations()
	assert.Empty(t, recommendations)

	// Add a test recommendation
	testRec := &OptimizationRecommendation{
		ID:          "test_rec_1",
		Type:        "performance",
		Category:    "response_time",
		Priority:    "high",
		Confidence:  0.85,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
		IsActive:    true,
		Status:      "pending",
		Title:       "Test Recommendation",
		Description: "Test optimization recommendation",
	}

	pos.mu.Lock()
	pos.optimizationHistory = append(pos.optimizationHistory, testRec)
	pos.mu.Unlock()

	// Check that recommendation is returned
	recommendations = pos.GetRecommendations()
	assert.Len(t, recommendations, 1)
	assert.Equal(t, "test_rec_1", recommendations[0].ID)
}

func TestPerformanceOptimizationSystem_GetRecommendation(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}
	benchmarkingSystem := &PerformanceBenchmarkingSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}

	pos := NewPerformanceOptimizationSystem(performanceMonitor, regressionDetection, benchmarkingSystem, predictiveAnalytics, config, logger)

	// Add a test recommendation
	testRec := &OptimizationRecommendation{
		ID:          "test_rec_1",
		Type:        "performance",
		Category:    "response_time",
		Priority:    "high",
		Confidence:  0.85,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
		IsActive:    true,
		Status:      "pending",
		Title:       "Test Recommendation",
		Description: "Test optimization recommendation",
	}

	pos.mu.Lock()
	pos.optimizationHistory = append(pos.optimizationHistory, testRec)
	pos.mu.Unlock()

	// Get existing recommendation
	rec := pos.GetRecommendation("test_rec_1")
	assert.NotNil(t, rec)
	assert.Equal(t, "Test Recommendation", rec.Title)

	// Get non-existent recommendation
	rec = pos.GetRecommendation("non_existent")
	assert.Nil(t, rec)
}

func TestPerformanceOptimizationSystem_ApproveRecommendation(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}
	benchmarkingSystem := &PerformanceBenchmarkingSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}

	pos := NewPerformanceOptimizationSystem(performanceMonitor, regressionDetection, benchmarkingSystem, predictiveAnalytics, config, logger)

	// Add a test recommendation
	testRec := &OptimizationRecommendation{
		ID:          "test_rec_1",
		Type:        "performance",
		Category:    "response_time",
		Priority:    "high",
		Confidence:  0.85,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
		IsActive:    true,
		Status:      "pending",
		Title:       "Test Recommendation",
		Description: "Test optimization recommendation",
	}

	pos.mu.Lock()
	pos.optimizationHistory = append(pos.optimizationHistory, testRec)
	pos.mu.Unlock()

	// Approve the recommendation
	err := pos.ApproveRecommendation("test_rec_1", "test_user")
	assert.NoError(t, err)

	// Check that recommendation was approved
	rec := pos.GetRecommendation("test_rec_1")
	assert.Equal(t, "approved", rec.Status)
	assert.Equal(t, "test_user", rec.ApprovedBy)
	assert.NotZero(t, rec.ApprovedAt)

	// Try to approve already approved recommendation
	err = pos.ApproveRecommendation("test_rec_1", "another_user")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in pending status")

	// Try to approve non-existent recommendation
	err = pos.ApproveRecommendation("non_existent", "test_user")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPerformanceOptimizationSystem_RejectRecommendation(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}
	benchmarkingSystem := &PerformanceBenchmarkingSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}

	pos := NewPerformanceOptimizationSystem(performanceMonitor, regressionDetection, benchmarkingSystem, predictiveAnalytics, config, logger)

	// Add a test recommendation
	testRec := &OptimizationRecommendation{
		ID:          "test_rec_1",
		Type:        "performance",
		Category:    "response_time",
		Priority:    "high",
		Confidence:  0.85,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
		IsActive:    true,
		Status:      "pending",
		Title:       "Test Recommendation",
		Description: "Test optimization recommendation",
	}

	pos.mu.Lock()
	pos.optimizationHistory = append(pos.optimizationHistory, testRec)
	pos.mu.Unlock()

	// Reject the recommendation
	err := pos.RejectRecommendation("test_rec_1", "Not needed")
	assert.NoError(t, err)

	// Check that recommendation was rejected
	rec := pos.GetRecommendation("test_rec_1")
	assert.Equal(t, "rejected", rec.Status)
	assert.False(t, rec.IsActive)
	assert.Equal(t, "Not needed", rec.Notes)

	// Try to reject already rejected recommendation
	err = pos.RejectRecommendation("test_rec_1", "Another reason")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in pending status")

	// Try to reject non-existent recommendation
	err = pos.RejectRecommendation("non_existent", "Not needed")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPerformanceOptimizationSystem_ImplementRecommendation(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}
	benchmarkingSystem := &PerformanceBenchmarkingSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}

	pos := NewPerformanceOptimizationSystem(performanceMonitor, regressionDetection, benchmarkingSystem, predictiveAnalytics, config, logger)

	// Add a test recommendation
	testRec := &OptimizationRecommendation{
		ID:          "test_rec_1",
		Type:        "performance",
		Category:    "response_time",
		Priority:    "high",
		Confidence:  0.85,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
		IsActive:    true,
		Status:      "pending",
		Title:       "Test Recommendation",
		Description: "Test optimization recommendation",
	}

	pos.mu.Lock()
	pos.optimizationHistory = append(pos.optimizationHistory, testRec)
	pos.mu.Unlock()

	// Implement the recommendation
	result, err := pos.ImplementRecommendation("test_rec_1", "Successfully implemented")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check that recommendation was implemented
	rec := pos.GetRecommendation("test_rec_1")
	assert.Equal(t, "implemented", rec.Status)
	assert.False(t, rec.IsActive)
	assert.NotZero(t, rec.ImplementedAt)

	// Check that result was created
	assert.Equal(t, "test_rec_1", result.RecommendationID)
	assert.Equal(t, "success", result.Status)
	assert.Equal(t, "Successfully implemented", result.ImplementationNotes)

	// Try to implement non-existent recommendation
	result, err = pos.ImplementRecommendation("non_existent", "Notes")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestPerformanceOptimizationSystem_GetOptimizationHistory(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}
	benchmarkingSystem := &PerformanceBenchmarkingSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}

	pos := NewPerformanceOptimizationSystem(performanceMonitor, regressionDetection, benchmarkingSystem, predictiveAnalytics, config, logger)

	// Initially no history
	history := pos.GetOptimizationHistory()
	assert.Empty(t, history)

	// Add test recommendations
	testRec1 := &OptimizationRecommendation{
		ID:         "test_rec_1",
		Type:       "performance",
		Category:   "response_time",
		Priority:   "high",
		Confidence: 0.85,
		CreatedAt:  time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(24 * time.Hour),
		IsActive:   true,
		Status:     "pending",
		Title:      "Test Recommendation 1",
	}

	testRec2 := &OptimizationRecommendation{
		ID:         "test_rec_2",
		Type:       "resource",
		Category:   "cpu_optimization",
		Priority:   "medium",
		Confidence: 0.75,
		CreatedAt:  time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(24 * time.Hour),
		IsActive:   true,
		Status:     "implemented",
		Title:      "Test Recommendation 2",
	}

	pos.mu.Lock()
	pos.optimizationHistory = append(pos.optimizationHistory, testRec1, testRec2)
	pos.mu.Unlock()

	// Check that history is returned
	history = pos.GetOptimizationHistory()
	assert.Len(t, history, 2)
}

func TestPerformanceOptimizationSystem_GetOptimizationResults(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}
	benchmarkingSystem := &PerformanceBenchmarkingSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}

	pos := NewPerformanceOptimizationSystem(performanceMonitor, regressionDetection, benchmarkingSystem, predictiveAnalytics, config, logger)

	// Initially no results
	results := pos.GetOptimizationResults()
	assert.Empty(t, results)

	// Add a test result
	testResult := &OptimizationResult{
		ID:                  "test_result_1",
		RecommendationID:    "test_rec_1",
		ImplementedAt:       time.Now().UTC(),
		Status:              "success",
		ImplementationNotes: "Test implementation",
		Tags:                make(map[string]string),
	}

	pos.mu.Lock()
	pos.implementedOptimizations["test_result_1"] = testResult
	pos.mu.Unlock()

	// Check that results are returned
	results = pos.GetOptimizationResults()
	assert.Len(t, results, 1)
	assert.Equal(t, "test_result_1", results["test_result_1"].ID)
}

func TestOptimizationRecommendationEngine_GenerateRecommendations(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{}

	engine := NewOptimizationRecommendationEngine(config, logger)

	// Create test current metrics
	currentMetrics := &PerformanceMetrics{
		ResponseTime: struct {
			Current     time.Duration `json:"current"`
			Expected    time.Duration `json:"expected"`
			Improvement float64       `json:"improvement"`
		}{
			Current:     600 * time.Millisecond,
			Expected:    500 * time.Millisecond,
			Improvement: 0.0,
		},
		Throughput: struct {
			Current     float64 `json:"current"`
			Expected    float64 `json:"expected"`
			Improvement float64 `json:"improvement"`
		}{
			Current:     700.0,
			Expected:    800.0,
			Improvement: 0.0,
		},
		SuccessRate: struct {
			Current     float64 `json:"current"`
			Expected    float64 `json:"expected"`
			Improvement float64 `json:"improvement"`
		}{
			Current:     0.93,
			Expected:    0.95,
			Improvement: 0.0,
		},
		ResourceUsage: struct {
			CPU struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"cpu"`
			Memory struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"memory"`
			Disk struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"disk"`
		}{
			CPU: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current:     85.0,
				Expected:    80.0,
				Improvement: 0.0,
			},
			Memory: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current:     90.0,
				Expected:    85.0,
				Improvement: 0.0,
			},
			Disk: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current:     70.0,
				Expected:    75.0,
				Improvement: 0.0,
			},
		},
	}

	// Create test historical data
	historicalData := make([]*PerformanceDataPoint, 0)
	now := time.Now().UTC()

	for i := 0; i < 200; i++ {
		dataPoint := &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(-i) * time.Minute),
			ResponseTime: time.Duration(600+i%50) * time.Millisecond, // Above threshold
			SuccessRate:  0.93 + float64(i%5)*0.002,                  // Below threshold
			Throughput:   700.0 + float64(i%100),                     // Below threshold
			ErrorRate:    0.07 - float64(i%5)*0.002,                  // Above threshold
			CPUUsage:     85.0 + float64(i%20),                       // Above threshold
			MemoryUsage:  90.0 + float64(i%15),                       // Above threshold
			DiskUsage:    70.0 + float64(i%10),
			NetworkIO:    100.0 + float64(i%30),
			ActiveUsers:  int64(100 + i%50),
			DataVolume:   int64(1000000 + i*1000),
		}
		historicalData = append(historicalData, dataPoint)
	}

	// Generate recommendations
	recommendations := engine.GenerateRecommendations(currentMetrics, historicalData)

	// Should generate multiple recommendations based on the test data
	assert.NotEmpty(t, recommendations)

	// Check for specific recommendation types
	foundResponseTime := false
	foundThroughput := false
	foundSuccessRate := false
	foundCPU := false
	foundMemory := false
	foundErrorHandling := false

	for _, rec := range recommendations {
		switch rec.Category {
		case "response_time":
			foundResponseTime = true
			assert.Equal(t, "high", rec.Priority)
			assert.Greater(t, rec.Confidence, 0.8)
		case "throughput":
			foundThroughput = true
			assert.Equal(t, "medium", rec.Priority)
		case "success_rate":
			foundSuccessRate = true
			assert.Equal(t, "critical", rec.Priority)
		case "cpu_optimization":
			foundCPU = true
			assert.Equal(t, "high", rec.Priority)
		case "memory_optimization":
			foundMemory = true
			assert.Equal(t, "medium", rec.Priority)
		case "error_handling":
			foundErrorHandling = true
			assert.Equal(t, "high", rec.Priority)
		}
	}

	// Verify that expected recommendations were generated
	assert.True(t, foundResponseTime, "Response time optimization recommendation should be generated")
	assert.True(t, foundThroughput, "Throughput optimization recommendation should be generated")
	assert.True(t, foundSuccessRate, "Success rate optimization recommendation should be generated")
	assert.True(t, foundCPU, "CPU optimization recommendation should be generated")
	assert.True(t, foundMemory, "Memory optimization recommendation should be generated")
	assert.True(t, foundErrorHandling, "Error handling optimization recommendation should be generated")
}

func TestOptimizationRecommendationEngine_AnalyzeResponseTime(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{}

	engine := NewOptimizationRecommendationEngine(config, logger)

	// Create test current metrics
	currentMetrics := &PerformanceMetrics{}

	// Create historical data with high response time
	historicalData := make([]*PerformanceDataPoint, 0)
	now := time.Now().UTC()

	for i := 0; i < 200; i++ {
		dataPoint := &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(-i) * time.Minute),
			ResponseTime: time.Duration(600+i%50) * time.Millisecond, // Above 500ms threshold
			SuccessRate:  0.98,
			Throughput:   1000.0,
			ErrorRate:    0.02,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
			DiskUsage:    70.0,
			NetworkIO:    100.0,
			ActiveUsers:  int64(100 + i%50),
			DataVolume:   int64(1000000 + i*1000),
		}
		historicalData = append(historicalData, dataPoint)
	}

	// Analyze response time
	rec := engine.analyzeResponseTime(currentMetrics, historicalData)

	// Should generate recommendation for high response time
	assert.NotNil(t, rec)
	assert.Equal(t, "performance", rec.Type)
	assert.Equal(t, "response_time", rec.Category)
	assert.Equal(t, "high", rec.Priority)
	assert.Greater(t, rec.Confidence, 0.8)
	assert.Contains(t, rec.Title, "Response Time")
	assert.Contains(t, rec.Solution, "caching")
	assert.NotNil(t, rec.ImprovementEstimate)
	assert.Greater(t, rec.ImprovementEstimate.ResponseTimeImprovement, 0.0)
}

func TestOptimizationRecommendationEngine_AnalyzeThroughput(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{}

	engine := NewOptimizationRecommendationEngine(config, logger)

	// Create test current metrics
	currentMetrics := &PerformanceMetrics{}

	// Create historical data with low throughput
	historicalData := make([]*PerformanceDataPoint, 0)
	now := time.Now().UTC()

	for i := 0; i < 200; i++ {
		dataPoint := &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(-i) * time.Minute),
			ResponseTime: time.Duration(250+i%50) * time.Millisecond,
			SuccessRate:  0.98,
			Throughput:   700.0 + float64(i%100), // Below 800 threshold
			ErrorRate:    0.02,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
			DiskUsage:    70.0,
			NetworkIO:    100.0,
			ActiveUsers:  int64(100 + i%50),
			DataVolume:   int64(1000000 + i*1000),
		}
		historicalData = append(historicalData, dataPoint)
	}

	// Analyze throughput
	rec := engine.analyzeThroughput(currentMetrics, historicalData)

	// Should generate recommendation for low throughput
	assert.NotNil(t, rec)
	assert.Equal(t, "performance", rec.Type)
	assert.Equal(t, "throughput", rec.Category)
	assert.Equal(t, "medium", rec.Priority)
	assert.Greater(t, rec.Confidence, 0.7)
	assert.Contains(t, rec.Title, "Throughput")
	assert.Contains(t, rec.Solution, "connection pooling")
	assert.NotNil(t, rec.ImprovementEstimate)
	assert.Greater(t, rec.ImprovementEstimate.ThroughputImprovement, 0.0)
}

func TestOptimizationRecommendationEngine_AnalyzeSuccessRate(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{}

	engine := NewOptimizationRecommendationEngine(config, logger)

	// Create test current metrics
	currentMetrics := &PerformanceMetrics{}

	// Create historical data with low success rate
	historicalData := make([]*PerformanceDataPoint, 0)
	now := time.Now().UTC()

	for i := 0; i < 200; i++ {
		dataPoint := &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(-i) * time.Minute),
			ResponseTime: time.Duration(250+i%50) * time.Millisecond,
			SuccessRate:  0.93 + float64(i%5)*0.002, // Below 0.95 threshold
			Throughput:   1000.0,
			ErrorRate:    0.07 - float64(i%5)*0.002,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
			DiskUsage:    70.0,
			NetworkIO:    100.0,
			ActiveUsers:  int64(100 + i%50),
			DataVolume:   int64(1000000 + i*1000),
		}
		historicalData = append(historicalData, dataPoint)
	}

	// Analyze success rate
	rec := engine.analyzeSuccessRate(currentMetrics, historicalData)

	// Should generate recommendation for low success rate
	assert.NotNil(t, rec)
	assert.Equal(t, "reliability", rec.Type)
	assert.Equal(t, "success_rate", rec.Category)
	assert.Equal(t, "critical", rec.Priority)
	assert.Greater(t, rec.Confidence, 0.8)
	assert.Contains(t, rec.Title, "Reliability")
	assert.Contains(t, rec.Solution, "error handling")
	assert.NotNil(t, rec.ImprovementEstimate)
	assert.Greater(t, rec.ImprovementEstimate.SuccessRateImprovement, 0.0)
}

func TestOptimizationRecommendationEngine_AnalyzeResourceUsage(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{}

	engine := NewOptimizationRecommendationEngine(config, logger)

	// Create test current metrics
	currentMetrics := &PerformanceMetrics{}

	// Create historical data with high resource usage
	historicalData := make([]*PerformanceDataPoint, 0)
	now := time.Now().UTC()

	for i := 0; i < 200; i++ {
		dataPoint := &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(-i) * time.Minute),
			ResponseTime: time.Duration(250+i%50) * time.Millisecond,
			SuccessRate:  0.98,
			Throughput:   1000.0,
			ErrorRate:    0.02,
			CPUUsage:     85.0 + float64(i%20), // Above 80% threshold
			MemoryUsage:  90.0 + float64(i%15), // Above 85% threshold
			DiskUsage:    70.0 + float64(i%10),
			NetworkIO:    100.0 + float64(i%30),
			ActiveUsers:  int64(100 + i%50),
			DataVolume:   int64(1000000 + i*1000),
		}
		historicalData = append(historicalData, dataPoint)
	}

	// Analyze resource usage
	recs := engine.analyzeResourceUsage(currentMetrics, historicalData)

	// Should generate recommendations for high resource usage
	assert.NotEmpty(t, recs)

	// Check for CPU optimization recommendation
	foundCPU := false
	foundMemory := false

	for _, rec := range recs {
		if rec.Category == "cpu_optimization" {
			foundCPU = true
			assert.Equal(t, "resource", rec.Type)
			assert.Equal(t, "high", rec.Priority)
			assert.Greater(t, rec.Confidence, 0.7)
			assert.Contains(t, rec.Title, "CPU")
			assert.Contains(t, rec.Solution, "caching")
		} else if rec.Category == "memory_optimization" {
			foundMemory = true
			assert.Equal(t, "resource", rec.Type)
			assert.Equal(t, "medium", rec.Priority)
			assert.Greater(t, rec.Confidence, 0.7)
			assert.Contains(t, rec.Title, "Memory")
			assert.Contains(t, rec.Solution, "pooling")
		}
	}

	assert.True(t, foundCPU, "CPU optimization recommendation should be generated")
	assert.True(t, foundMemory, "Memory optimization recommendation should be generated")
}

func TestOptimizationRecommendationEngine_AnalyzeErrorPatterns(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{}

	engine := NewOptimizationRecommendationEngine(config, logger)

	// Create test current metrics
	currentMetrics := &PerformanceMetrics{}

	// Create historical data with high error rate
	historicalData := make([]*PerformanceDataPoint, 0)
	now := time.Now().UTC()

	for i := 0; i < 200; i++ {
		dataPoint := &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(-i) * time.Minute),
			ResponseTime: time.Duration(250+i%50) * time.Millisecond,
			SuccessRate:  0.93,
			Throughput:   1000.0,
			ErrorRate:    0.07 + float64(i%5)*0.002, // Above 5% threshold
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
			DiskUsage:    70.0,
			NetworkIO:    100.0,
			ActiveUsers:  int64(100 + i%50),
			DataVolume:   int64(1000000 + i*1000),
		}
		historicalData = append(historicalData, dataPoint)
	}

	// Analyze error patterns
	rec := engine.analyzeErrorPatterns(currentMetrics, historicalData)

	// Should generate recommendation for high error rate
	assert.NotNil(t, rec)
	assert.Equal(t, "reliability", rec.Type)
	assert.Equal(t, "error_handling", rec.Category)
	assert.Equal(t, "high", rec.Priority)
	assert.Greater(t, rec.Confidence, 0.8)
	assert.Contains(t, rec.Title, "Error Handling")
	assert.Contains(t, rec.Solution, "error handling")
	assert.NotNil(t, rec.ImprovementEstimate)
	assert.Greater(t, rec.ImprovementEstimate.SuccessRateImprovement, 0.0)
}

func TestPerformanceOptimizationSystem_FilterRecommendations(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{
		ConfidenceThreshold: 0.7,
		MaxRecommendations:  5,
	}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}
	benchmarkingSystem := &PerformanceBenchmarkingSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}

	pos := NewPerformanceOptimizationSystem(performanceMonitor, regressionDetection, benchmarkingSystem, predictiveAnalytics, config, logger)

	// Create test recommendations
	recommendations := []*OptimizationRecommendation{
		{
			Type:       "performance",
			Category:   "response_time",
			Priority:   "high",
			Confidence: 0.85,
			ExpiresAt:  time.Now().UTC().Add(24 * time.Hour),
		},
		{
			Type:       "performance",
			Category:   "throughput",
			Priority:   "medium",
			Confidence: 0.75,
			ExpiresAt:  time.Now().UTC().Add(24 * time.Hour),
		},
		{
			Type:       "resource",
			Category:   "cpu_optimization",
			Priority:   "low",
			Confidence: 0.6, // Below threshold
			ExpiresAt:  time.Now().UTC().Add(24 * time.Hour),
		},
		{
			Type:       "reliability",
			Category:   "error_handling",
			Priority:   "critical",
			Confidence: 0.90,
			ExpiresAt:  time.Now().UTC().Add(-1 * time.Hour), // Expired
		},
	}

	// Filter recommendations
	filtered := pos.filterRecommendations(recommendations)

	// Should filter out low confidence and expired recommendations
	assert.Len(t, filtered, 2)
	assert.Equal(t, "critical", filtered[0].Priority) // Should be sorted by priority
	assert.Equal(t, "high", filtered[1].Priority)
}

func TestPerformanceOptimizationSystem_SimilarRecommendationExists(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}
	benchmarkingSystem := &PerformanceBenchmarkingSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}

	pos := NewPerformanceOptimizationSystem(performanceMonitor, regressionDetection, benchmarkingSystem, predictiveAnalytics, config, logger)

	// Add existing recommendation
	existingRec := &OptimizationRecommendation{
		ID:         "existing_rec",
		Type:       "performance",
		Category:   "response_time",
		Priority:   "high",
		Confidence: 0.85,
		CreatedAt:  time.Now().UTC().Add(-12 * time.Hour), // Within 24 hours
		Status:     "pending",
	}

	pos.mu.Lock()
	pos.optimizationHistory = append(pos.optimizationHistory, existingRec)
	pos.mu.Unlock()

	// Test similar recommendation
	newRec := &OptimizationRecommendation{
		Type:     "performance",
		Category: "response_time",
	}

	exists := pos.similarRecommendationExists(newRec)
	assert.True(t, exists, "Similar recommendation should be detected")

	// Test different recommendation
	differentRec := &OptimizationRecommendation{
		Type:     "resource",
		Category: "cpu_optimization",
	}

	exists = pos.similarRecommendationExists(differentRec)
	assert.False(t, exists, "Different recommendation should not be detected")
}

func TestPerformanceOptimizationSystem_GetPriorityWeight(t *testing.T) {
	logger := zap.NewNop()
	config := OptimizationConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}
	benchmarkingSystem := &PerformanceBenchmarkingSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}

	pos := NewPerformanceOptimizationSystem(performanceMonitor, regressionDetection, benchmarkingSystem, predictiveAnalytics, config, logger)

	// Test priority weights
	assert.Equal(t, 4, pos.getPriorityWeight("critical"))
	assert.Equal(t, 3, pos.getPriorityWeight("high"))
	assert.Equal(t, 2, pos.getPriorityWeight("medium"))
	assert.Equal(t, 1, pos.getPriorityWeight("low"))
	assert.Equal(t, 0, pos.getPriorityWeight("unknown"))
}
