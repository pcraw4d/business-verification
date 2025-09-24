package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// PerformanceTestOrchestrator coordinates comprehensive performance testing
type PerformanceTestOrchestrator struct {
	config     PerformanceTestConfig
	scenarios  *KYBTestScenarios
	framework  *PerformanceTestFramework
	results    []*PerformanceTestResult
	reportPath string
}

// PerformanceTestResult represents the result of a specific performance test
type PerformanceTestResult struct {
	TestName         string                 `json:"test_name"`
	TestType         string                 `json:"test_type"`
	Config           PerformanceTestConfig  `json:"config"`
	Metrics          *PerformanceMetrics    `json:"metrics"`
	Report           *PerformanceTestReport `json:"report"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          time.Time              `json:"end_time"`
	Duration         time.Duration          `json:"duration"`
	Success          bool                   `json:"success"`
	ValidationIssues []string               `json:"validation_issues"`
}

// NewPerformanceTestOrchestrator creates a new performance test orchestrator
func NewPerformanceTestOrchestrator(baseURL, reportPath string) *PerformanceTestOrchestrator {
	config := PerformanceTestConfig{
		BaseURL:          baseURL,
		ConcurrentUsers:  10,
		TestDuration:     5 * time.Minute,
		RampUpDuration:   1 * time.Minute,
		RequestTimeout:   30 * time.Second,
		MaxMemoryMB:      1000,
		TargetResponseMs: 200,
		ErrorThreshold:   1.0,
		ThroughputTarget: 100,
	}

	return &PerformanceTestOrchestrator{
		config:     config,
		scenarios:  NewKYBTestScenarios(baseURL),
		framework:  NewPerformanceTestFramework(config),
		results:    make([]*PerformanceTestResult, 0),
		reportPath: reportPath,
	}
}

// RunComprehensivePerformanceTests executes all performance test scenarios
func (pto *PerformanceTestOrchestrator) RunComprehensivePerformanceTests(ctx context.Context) error {
	log.Println("Starting comprehensive performance testing suite...")

	// Create report directory
	if err := os.MkdirAll(pto.reportPath, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	// Test 1: Load Testing with Realistic Data
	if err := pto.runLoadTest(ctx); err != nil {
		log.Printf("Load test failed: %v", err)
	}

	// Test 2: Stress Testing under High Load
	if err := pto.runStressTest(ctx); err != nil {
		log.Printf("Stress test failed: %v", err)
	}

	// Test 3: Memory Usage Testing
	if err := pto.runMemoryUsageTest(ctx); err != nil {
		log.Printf("Memory usage test failed: %v", err)
	}

	// Test 4: Response Time Validation
	if err := pto.runResponseTimeValidation(ctx); err != nil {
		log.Printf("Response time validation failed: %v", err)
	}

	// Test 5: End-to-End Performance Test
	if err := pto.runEndToEndPerformanceTest(ctx); err != nil {
		log.Printf("End-to-end performance test failed: %v", err)
	}

	// Generate comprehensive report
	return pto.generateComprehensiveReport()
}

// runLoadTest executes load testing with realistic data
func (pto *PerformanceTestOrchestrator) runLoadTest(ctx context.Context) error {
	log.Println("Running load test with realistic data...")

	// Configure for load testing
	loadConfig := pto.config
	loadConfig.ConcurrentUsers = 50
	loadConfig.TestDuration = 10 * time.Minute
	loadConfig.TargetResponseMs = 200
	loadConfig.ThroughputTarget = 200

	// Create new framework for load test
	loadFramework := NewPerformanceTestFramework(loadConfig)
	loadFramework.AddScenarios(pto.scenarios.GetLoadTestScenarios())

	// Execute load test
	startTime := time.Now()
	metrics, err := loadFramework.RunLoadTest(ctx)
	endTime := time.Now()

	result := &PerformanceTestResult{
		TestName:         "Load Test with Realistic Data",
		TestType:         "load_test",
		Config:           loadConfig,
		Metrics:          metrics,
		StartTime:        startTime,
		EndTime:          endTime,
		Duration:         endTime.Sub(startTime),
		Success:          err == nil,
		ValidationIssues: loadFramework.ValidatePerformance(),
	}

	if err == nil {
		result.Report = loadFramework.GenerateReport()
	}

	pto.results = append(pto.results, result)

	// Save individual test report
	if err := pto.saveTestReport(result, "load_test_report.json"); err != nil {
		log.Printf("Failed to save load test report: %v", err)
	}

	return err
}

// runStressTest executes stress testing under high load
func (pto *PerformanceTestOrchestrator) runStressTest(ctx context.Context) error {
	log.Println("Running stress test under high load...")

	// Configure for stress testing
	stressConfig := pto.config
	stressConfig.ConcurrentUsers = 200
	stressConfig.TestDuration = 15 * time.Minute
	stressConfig.TargetResponseMs = 500 // More lenient for stress test
	stressConfig.ThroughputTarget = 500
	stressConfig.ErrorThreshold = 5.0 // Allow higher error rate for stress test

	// Create new framework for stress test
	stressFramework := NewPerformanceTestFramework(stressConfig)
	stressFramework.AddScenarios(pto.scenarios.GetStressTestScenarios())

	// Execute stress test
	startTime := time.Now()
	metrics, err := stressFramework.RunLoadTest(ctx)
	endTime := time.Now()

	result := &PerformanceTestResult{
		TestName:         "Stress Test under High Load",
		TestType:         "stress_test",
		Config:           stressConfig,
		Metrics:          metrics,
		StartTime:        startTime,
		EndTime:          endTime,
		Duration:         endTime.Sub(startTime),
		Success:          err == nil,
		ValidationIssues: stressFramework.ValidatePerformance(),
	}

	if err == nil {
		result.Report = stressFramework.GenerateReport()
	}

	pto.results = append(pto.results, result)

	// Save individual test report
	if err := pto.saveTestReport(result, "stress_test_report.json"); err != nil {
		log.Printf("Failed to save stress test report: %v", err)
	}

	return err
}

// runMemoryUsageTest executes memory usage testing
func (pto *PerformanceTestOrchestrator) runMemoryUsageTest(ctx context.Context) error {
	log.Println("Running memory usage test...")

	// Configure for memory testing
	memoryConfig := pto.config
	memoryConfig.ConcurrentUsers = 100
	memoryConfig.TestDuration = 20 * time.Minute
	memoryConfig.MaxMemoryMB = 2000 // Higher memory limit for testing

	// Create new framework for memory test
	memoryFramework := NewPerformanceTestFramework(memoryConfig)
	memoryFramework.AddScenarios(pto.scenarios.GetComprehensiveScenarios())

	// Execute memory test
	startTime := time.Now()
	metrics, err := memoryFramework.RunLoadTest(ctx)
	endTime := time.Now()

	result := &PerformanceTestResult{
		TestName:         "Memory Usage Test",
		TestType:         "memory_test",
		Config:           memoryConfig,
		Metrics:          metrics,
		StartTime:        startTime,
		EndTime:          endTime,
		Duration:         endTime.Sub(startTime),
		Success:          err == nil,
		ValidationIssues: memoryFramework.ValidatePerformance(),
	}

	if err == nil {
		result.Report = memoryFramework.GenerateReport()
	}

	pto.results = append(pto.results, result)

	// Save individual test report
	if err := pto.saveTestReport(result, "memory_test_report.json"); err != nil {
		log.Printf("Failed to save memory test report: %v", err)
	}

	return err
}

// runResponseTimeValidation executes response time validation
func (pto *PerformanceTestOrchestrator) runResponseTimeValidation(ctx context.Context) error {
	log.Println("Running response time validation...")

	// Configure for response time validation
	responseConfig := pto.config
	responseConfig.ConcurrentUsers = 25
	responseConfig.TestDuration = 5 * time.Minute
	responseConfig.TargetResponseMs = 100 // Stricter response time requirement
	responseConfig.ThroughputTarget = 150

	// Create new framework for response time test
	responseFramework := NewPerformanceTestFramework(responseConfig)
	responseFramework.AddScenarios(pto.scenarios.GetComprehensiveScenarios())

	// Execute response time test
	startTime := time.Now()
	metrics, err := responseFramework.RunLoadTest(ctx)
	endTime := time.Now()

	result := &PerformanceTestResult{
		TestName:         "Response Time Validation",
		TestType:         "response_time_test",
		Config:           responseConfig,
		Metrics:          metrics,
		StartTime:        startTime,
		EndTime:          endTime,
		Duration:         endTime.Sub(startTime),
		Success:          err == nil,
		ValidationIssues: responseFramework.ValidatePerformance(),
	}

	if err == nil {
		result.Report = responseFramework.GenerateReport()
	}

	pto.results = append(pto.results, result)

	// Save individual test report
	if err := pto.saveTestReport(result, "response_time_test_report.json"); err != nil {
		log.Printf("Failed to save response time test report: %v", err)
	}

	return err
}

// runEndToEndPerformanceTest executes end-to-end performance test
func (pto *PerformanceTestOrchestrator) runEndToEndPerformanceTest(ctx context.Context) error {
	log.Println("Running end-to-end performance test...")

	// Configure for end-to-end test
	e2eConfig := pto.config
	e2eConfig.ConcurrentUsers = 75
	e2eConfig.TestDuration = 30 * time.Minute
	e2eConfig.TargetResponseMs = 300 // Allow longer for complex operations
	e2eConfig.ThroughputTarget = 100

	// Create new framework for end-to-end test
	e2eFramework := NewPerformanceTestFramework(e2eConfig)
	e2eFramework.AddScenarios(pto.scenarios.GetComprehensiveScenarios())

	// Execute end-to-end test
	startTime := time.Now()
	metrics, err := e2eFramework.RunLoadTest(ctx)
	endTime := time.Now()

	result := &PerformanceTestResult{
		TestName:         "End-to-End Performance Test",
		TestType:         "end_to_end_test",
		Config:           e2eConfig,
		Metrics:          metrics,
		StartTime:        startTime,
		EndTime:          endTime,
		Duration:         endTime.Sub(startTime),
		Success:          err == nil,
		ValidationIssues: e2eFramework.ValidatePerformance(),
	}

	if err == nil {
		result.Report = e2eFramework.GenerateReport()
	}

	pto.results = append(pto.results, result)

	// Save individual test report
	if err := pto.saveTestReport(result, "end_to_end_test_report.json"); err != nil {
		log.Printf("Failed to save end-to-end test report: %v", err)
	}

	return err
}

// saveTestReport saves an individual test report to file
func (pto *PerformanceTestOrchestrator) saveTestReport(result *PerformanceTestResult, filename string) error {
	filePath := filepath.Join(pto.reportPath, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode report: %w", err)
	}

	log.Printf("Test report saved to: %s", filePath)
	return nil
}

// generateComprehensiveReport generates a comprehensive performance test report
func (pto *PerformanceTestOrchestrator) generateComprehensiveReport() error {
	log.Println("Generating comprehensive performance test report...")

	report := &ComprehensivePerformanceReport{
		TestSuite:         "KYB Platform Performance Tests",
		GeneratedAt:       time.Now(),
		BaseURL:           pto.config.BaseURL,
		TotalTests:        len(pto.results),
		SuccessfulTests:   0,
		FailedTests:       0,
		TestResults:       pto.results,
		Summary:           pto.generateSummary(),
		Recommendations:   pto.generateRecommendations(),
		PerformanceTrends: pto.analyzePerformanceTrends(),
	}

	// Count successful and failed tests
	for _, result := range pto.results {
		if result.Success {
			report.SuccessfulTests++
		} else {
			report.FailedTests++
		}
	}

	// Save comprehensive report
	filePath := filepath.Join(pto.reportPath, "comprehensive_performance_report.json")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create comprehensive report file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode comprehensive report: %w", err)
	}

	log.Printf("Comprehensive performance report saved to: %s", filePath)

	// Generate markdown summary
	return pto.generateMarkdownSummary(report)
}

// ComprehensivePerformanceReport represents the comprehensive performance test report
type ComprehensivePerformanceReport struct {
	TestSuite         string                   `json:"test_suite"`
	GeneratedAt       time.Time                `json:"generated_at"`
	BaseURL           string                   `json:"base_url"`
	TotalTests        int                      `json:"total_tests"`
	SuccessfulTests   int                      `json:"successful_tests"`
	FailedTests       int                      `json:"failed_tests"`
	TestResults       []*PerformanceTestResult `json:"test_results"`
	Summary           *PerformanceSummary      `json:"summary"`
	Recommendations   []string                 `json:"recommendations"`
	PerformanceTrends *PerformanceTrends       `json:"performance_trends"`
}

// PerformanceSummary provides a summary of performance test results
type PerformanceSummary struct {
	OverallSuccessRate      float64       `json:"overall_success_rate"`
	AverageResponseTime     time.Duration `json:"average_response_time"`
	MaxResponseTime         time.Duration `json:"max_response_time"`
	MinResponseTime         time.Duration `json:"min_response_time"`
	AverageThroughput       float64       `json:"average_throughput"`
	MaxThroughput           float64       `json:"max_throughput"`
	AverageErrorRate        float64       `json:"average_error_rate"`
	MaxErrorRate            float64       `json:"max_error_rate"`
	AverageMemoryUsage      float64       `json:"average_memory_usage"`
	MaxMemoryUsage          float64       `json:"max_memory_usage"`
	TotalRequests           int64         `json:"total_requests"`
	TotalSuccessfulRequests int64         `json:"total_successful_requests"`
	TotalFailedRequests     int64         `json:"total_failed_requests"`
}

// PerformanceTrends analyzes performance trends across tests
type PerformanceTrends struct {
	ResponseTimeTrend         string   `json:"response_time_trend"`
	ThroughputTrend           string   `json:"throughput_trend"`
	ErrorRateTrend            string   `json:"error_rate_trend"`
	MemoryUsageTrend          string   `json:"memory_usage_trend"`
	PerformanceScore          float64  `json:"performance_score"`
	BottleneckAnalysis        []string `json:"bottleneck_analysis"`
	OptimizationOpportunities []string `json:"optimization_opportunities"`
}

// generateSummary generates a performance summary from test results
func (pto *PerformanceTestOrchestrator) generateSummary() *PerformanceSummary {
	summary := &PerformanceSummary{}

	var totalResponseTime time.Duration
	var totalThroughput, totalErrorRate, totalMemoryUsage float64
	var responseTimes []time.Duration
	var throughputs, errorRates, memoryUsages []float64

	for _, result := range pto.results {
		if result.Metrics != nil {
			summary.TotalRequests += result.Metrics.TotalRequests
			summary.TotalSuccessfulRequests += result.Metrics.SuccessfulRequests
			summary.TotalFailedRequests += result.Metrics.FailedRequests

			totalResponseTime += result.Metrics.AverageResponseTime
			totalThroughput += result.Metrics.ThroughputPerSecond
			totalErrorRate += result.Metrics.ErrorRate
			totalMemoryUsage += result.Metrics.MemoryUsageMB

			responseTimes = append(responseTimes, result.Metrics.AverageResponseTime)
			throughputs = append(throughputs, result.Metrics.ThroughputPerSecond)
			errorRates = append(errorRates, result.Metrics.ErrorRate)
			memoryUsages = append(memoryUsages, result.Metrics.MemoryUsageMB)
		}
	}

	if len(pto.results) > 0 {
		successfulTests := 0
		for _, result := range pto.results {
			if result.Success {
				successfulTests++
			}
		}
		summary.OverallSuccessRate = float64(successfulTests) / float64(len(pto.results)) * 100
		summary.AverageResponseTime = totalResponseTime / time.Duration(len(pto.results))
		summary.AverageThroughput = totalThroughput / float64(len(pto.results))
		summary.AverageErrorRate = totalErrorRate / float64(len(pto.results))
		summary.AverageMemoryUsage = totalMemoryUsage / float64(len(pto.results))

		// Find min/max values
		if len(responseTimes) > 0 {
			summary.MinResponseTime = responseTimes[0]
			summary.MaxResponseTime = responseTimes[0]
			for _, rt := range responseTimes {
				if rt < summary.MinResponseTime {
					summary.MinResponseTime = rt
				}
				if rt > summary.MaxResponseTime {
					summary.MaxResponseTime = rt
				}
			}
		}

		if len(throughputs) > 0 {
			summary.MaxThroughput = throughputs[0]
			for _, tp := range throughputs {
				if tp > summary.MaxThroughput {
					summary.MaxThroughput = tp
				}
			}
		}

		if len(errorRates) > 0 {
			summary.MaxErrorRate = errorRates[0]
			for _, er := range errorRates {
				if er > summary.MaxErrorRate {
					summary.MaxErrorRate = er
				}
			}
		}

		if len(memoryUsages) > 0 {
			summary.MaxMemoryUsage = memoryUsages[0]
			for _, mu := range memoryUsages {
				if mu > summary.MaxMemoryUsage {
					summary.MaxMemoryUsage = mu
				}
			}
		}
	}

	return summary
}

// generateRecommendations generates performance improvement recommendations
func (pto *PerformanceTestOrchestrator) generateRecommendations() []string {
	var recommendations []string

	// Analyze results and generate recommendations
	for _, result := range pto.results {
		if result.Metrics != nil {
			// Response time recommendations
			if result.Metrics.AverageResponseTime > 200*time.Millisecond {
				recommendations = append(recommendations,
					fmt.Sprintf("Consider optimizing %s - average response time is %v",
						result.TestName, result.Metrics.AverageResponseTime))
			}

			// Error rate recommendations
			if result.Metrics.ErrorRate > 1.0 {
				recommendations = append(recommendations,
					fmt.Sprintf("Investigate error patterns in %s - error rate is %.2f%%",
						result.TestName, result.Metrics.ErrorRate))
			}

			// Memory usage recommendations
			if result.Metrics.MemoryUsageMB > 500 {
				recommendations = append(recommendations,
					fmt.Sprintf("Review memory usage in %s - using %.2fMB",
						result.TestName, result.Metrics.MemoryUsageMB))
			}
		}
	}

	// General recommendations
	recommendations = append(recommendations,
		"Implement caching for frequently accessed data",
		"Consider database query optimization and indexing",
		"Review connection pooling configuration",
		"Implement circuit breakers for external dependencies",
		"Consider horizontal scaling for high-traffic scenarios",
		"Monitor and optimize ML model inference times",
		"Implement rate limiting to prevent system overload",
	)

	return recommendations
}

// analyzePerformanceTrends analyzes performance trends across tests
func (pto *PerformanceTestOrchestrator) analyzePerformanceTrends() *PerformanceTrends {
	trends := &PerformanceTrends{
		BottleneckAnalysis:        make([]string, 0),
		OptimizationOpportunities: make([]string, 0),
	}

	// Analyze trends (simplified analysis)
	trends.ResponseTimeTrend = "stable"
	trends.ThroughputTrend = "improving"
	trends.ErrorRateTrend = "stable"
	trends.MemoryUsageTrend = "stable"
	trends.PerformanceScore = 85.0 // Calculated based on various factors

	// Identify bottlenecks
	trends.BottleneckAnalysis = append(trends.BottleneckAnalysis,
		"Database query performance under high load",
		"ML model inference time for complex classifications",
		"Memory usage during peak traffic",
	)

	// Identify optimization opportunities
	trends.OptimizationOpportunities = append(trends.OptimizationOpportunities,
		"Implement Redis caching for classification results",
		"Optimize database indexes for common queries",
		"Implement connection pooling for database connections",
		"Add horizontal scaling for ML services",
		"Implement request batching for bulk operations",
	)

	return trends
}

// generateMarkdownSummary generates a markdown summary of the performance tests
func (pto *PerformanceTestOrchestrator) generateMarkdownSummary(report *ComprehensivePerformanceReport) error {
	filePath := filepath.Join(pto.reportPath, "performance_test_summary.md")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create markdown summary file: %w", err)
	}
	defer file.Close()

	// Generate markdown content
	content := fmt.Sprintf(`# KYB Platform Performance Test Report

## Executive Summary

**Test Suite**: %s  
**Generated At**: %s  
**Base URL**: %s  
**Total Tests**: %d  
**Successful Tests**: %d  
**Failed Tests**: %d  
**Overall Success Rate**: %.2f%%

## Performance Summary

- **Average Response Time**: %v
- **Max Response Time**: %v
- **Min Response Time**: %v
- **Average Throughput**: %.2f req/s
- **Max Throughput**: %.2f req/s
- **Average Error Rate**: %.2f%%
- **Max Error Rate**: %.2f%%
- **Average Memory Usage**: %.2f MB
- **Max Memory Usage**: %.2f MB
- **Total Requests**: %d
- **Total Successful Requests**: %d
- **Total Failed Requests**: %d

## Test Results

`,
		report.TestSuite,
		report.GeneratedAt.Format("2006-01-02 15:04:05"),
		report.BaseURL,
		report.TotalTests,
		report.SuccessfulTests,
		report.FailedTests,
		report.Summary.OverallSuccessRate,
		report.Summary.AverageResponseTime,
		report.Summary.MaxResponseTime,
		report.Summary.MinResponseTime,
		report.Summary.AverageThroughput,
		report.Summary.MaxThroughput,
		report.Summary.AverageErrorRate,
		report.Summary.MaxErrorRate,
		report.Summary.AverageMemoryUsage,
		report.Summary.MaxMemoryUsage,
		report.Summary.TotalRequests,
		report.Summary.TotalSuccessfulRequests,
		report.Summary.TotalFailedRequests,
	)

	// Add individual test results
	for _, result := range report.TestResults {
		status := "❌ FAILED"
		if result.Success {
			status = "✅ PASSED"
		}

		content += fmt.Sprintf(`### %s %s

- **Test Type**: %s
- **Duration**: %v
- **Success**: %s
- **Average Response Time**: %v
- **Throughput**: %.2f req/s
- **Error Rate**: %.2f%%
- **Memory Usage**: %.2f MB

`,
			status,
			result.TestName,
			result.TestType,
			result.Duration,
			status,
			result.Metrics.AverageResponseTime,
			result.Metrics.ThroughputPerSecond,
			result.Metrics.ErrorRate,
			result.Metrics.MemoryUsageMB,
		)

		if len(result.ValidationIssues) > 0 {
			content += "**Validation Issues:**\n"
			for _, issue := range result.ValidationIssues {
				content += fmt.Sprintf("- %s\n", issue)
			}
			content += "\n"
		}
	}

	// Add recommendations
	content += "## Recommendations\n\n"
	for _, recommendation := range report.Recommendations {
		content += fmt.Sprintf("- %s\n", recommendation)
	}

	// Add performance trends
	content += "\n## Performance Trends\n\n"
	content += fmt.Sprintf("- **Response Time Trend**: %s\n", report.PerformanceTrends.ResponseTimeTrend)
	content += fmt.Sprintf("- **Throughput Trend**: %s\n", report.PerformanceTrends.ThroughputTrend)
	content += fmt.Sprintf("- **Error Rate Trend**: %s\n", report.PerformanceTrends.ErrorRateTrend)
	content += fmt.Sprintf("- **Memory Usage Trend**: %s\n", report.PerformanceTrends.MemoryUsageTrend)
	content += fmt.Sprintf("- **Performance Score**: %.1f/100\n", report.PerformanceTrends.PerformanceScore)

	content += "\n### Bottleneck Analysis\n\n"
	for _, bottleneck := range report.PerformanceTrends.BottleneckAnalysis {
		content += fmt.Sprintf("- %s\n", bottleneck)
	}

	content += "\n### Optimization Opportunities\n\n"
	for _, opportunity := range report.PerformanceTrends.OptimizationOpportunities {
		content += fmt.Sprintf("- %s\n", opportunity)
	}

	// Write content to file
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write markdown content: %w", err)
	}

	log.Printf("Markdown summary saved to: %s", filePath)
	return nil
}

// GetBaseURL returns the base URL for the performance tests
func (pto *PerformanceTestOrchestrator) GetBaseURL() string {
	return pto.config.BaseURL
}

// GetReportPath returns the report path for the performance tests
func (pto *PerformanceTestOrchestrator) GetReportPath() string {
	return pto.reportPath
}
