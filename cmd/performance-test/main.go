package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/company/kyb-platform/internal/testing/performance"
)

func main() {
	// Parse command line flags
	var (
		baseURL    = flag.String("base-url", "http://localhost:8080", "Base URL for the KYB platform API")
		reportPath = flag.String("report-path", "./performance-reports", "Path to save performance test reports")
		testType   = flag.String("test-type", "comprehensive", "Type of test to run: load, stress, memory, response-time, end-to-end, comprehensive")
		help       = flag.Bool("help", false, "Show help information")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received interrupt signal, cancelling tests...")
		cancel()
	}()

	// Create performance test orchestrator
	orchestrator := performance.NewPerformanceTestOrchestrator(*baseURL, *reportPath)

	log.Printf("Starting performance tests for KYB Platform")
	log.Printf("Base URL: %s", *baseURL)
	log.Printf("Report Path: %s", *reportPath)
	log.Printf("Test Type: %s", *testType)

	// Run the specified test type
	var err error
	switch *testType {
	case "load":
		err = runLoadTest(ctx, orchestrator)
	case "stress":
		err = runStressTest(ctx, orchestrator)
	case "memory":
		err = runMemoryTest(ctx, orchestrator)
	case "response-time":
		err = runResponseTimeTest(ctx, orchestrator)
	case "end-to-end":
		err = runEndToEndTest(ctx, orchestrator)
	case "comprehensive":
		err = orchestrator.RunComprehensivePerformanceTests(ctx)
	default:
		log.Fatalf("Unknown test type: %s", *testType)
	}

	if err != nil {
		log.Fatalf("Performance test failed: %v", err)
	}

	log.Println("Performance tests completed successfully!")
}

// runLoadTest runs only the load test
func runLoadTest(ctx context.Context, orchestrator *performance.PerformanceTestOrchestrator) error {
	log.Println("Running load test...")

	// Create a simple orchestrator for single test
	config := performance.PerformanceTestConfig{
		BaseURL:          orchestrator.GetBaseURL(),
		ConcurrentUsers:  50,
		TestDuration:     10 * time.Minute,
		RampUpDuration:   1 * time.Minute,
		RequestTimeout:   30 * time.Second,
		MaxMemoryMB:      1000,
		TargetResponseMs: 200,
		ErrorThreshold:   1.0,
		ThroughputTarget: 200,
	}

	framework := performance.NewPerformanceTestFramework(config)
	scenarios := performance.NewKYBTestScenarios(config.BaseURL)
	framework.AddScenarios(scenarios.GetLoadTestScenarios())

	_, err := framework.RunLoadTest(ctx)
	if err != nil {
		return fmt.Errorf("load test failed: %w", err)
	}

	// Generate and save report
	report := framework.GenerateReport()
	return saveSingleTestReport(report, "load_test_report.json", orchestrator.GetReportPath())
}

// runStressTest runs only the stress test
func runStressTest(ctx context.Context, orchestrator *performance.PerformanceTestOrchestrator) error {
	log.Println("Running stress test...")

	config := performance.PerformanceTestConfig{
		BaseURL:          orchestrator.GetBaseURL(),
		ConcurrentUsers:  200,
		TestDuration:     15 * time.Minute,
		RampUpDuration:   2 * time.Minute,
		RequestTimeout:   30 * time.Second,
		MaxMemoryMB:      2000,
		TargetResponseMs: 500,
		ErrorThreshold:   5.0,
		ThroughputTarget: 500,
	}

	framework := performance.NewPerformanceTestFramework(config)
	scenarios := performance.NewKYBTestScenarios(config.BaseURL)
	framework.AddScenarios(scenarios.GetStressTestScenarios())

	_, err := framework.RunLoadTest(ctx)
	if err != nil {
		return fmt.Errorf("stress test failed: %w", err)
	}

	report := framework.GenerateReport()
	return saveSingleTestReport(report, "stress_test_report.json", orchestrator.GetReportPath())
}

// runMemoryTest runs only the memory test
func runMemoryTest(ctx context.Context, orchestrator *performance.PerformanceTestOrchestrator) error {
	log.Println("Running memory test...")

	config := performance.PerformanceTestConfig{
		BaseURL:          orchestrator.GetBaseURL(),
		ConcurrentUsers:  100,
		TestDuration:     20 * time.Minute,
		RampUpDuration:   2 * time.Minute,
		RequestTimeout:   30 * time.Second,
		MaxMemoryMB:      2000,
		TargetResponseMs: 300,
		ErrorThreshold:   2.0,
		ThroughputTarget: 150,
	}

	framework := performance.NewPerformanceTestFramework(config)
	scenarios := performance.NewKYBTestScenarios(config.BaseURL)
	framework.AddScenarios(scenarios.GetComprehensiveScenarios())

	_, err := framework.RunLoadTest(ctx)
	if err != nil {
		return fmt.Errorf("memory test failed: %w", err)
	}

	report := framework.GenerateReport()
	return saveSingleTestReport(report, "memory_test_report.json", orchestrator.GetReportPath())
}

// runResponseTimeTest runs only the response time test
func runResponseTimeTest(ctx context.Context, orchestrator *performance.PerformanceTestOrchestrator) error {
	log.Println("Running response time test...")

	config := performance.PerformanceTestConfig{
		BaseURL:          orchestrator.GetBaseURL(),
		ConcurrentUsers:  25,
		TestDuration:     5 * time.Minute,
		RampUpDuration:   30 * time.Second,
		RequestTimeout:   30 * time.Second,
		MaxMemoryMB:      1000,
		TargetResponseMs: 100,
		ErrorThreshold:   0.5,
		ThroughputTarget: 150,
	}

	framework := performance.NewPerformanceTestFramework(config)
	scenarios := performance.NewKYBTestScenarios(config.BaseURL)
	framework.AddScenarios(scenarios.GetComprehensiveScenarios())

	_, err := framework.RunLoadTest(ctx)
	if err != nil {
		return fmt.Errorf("response time test failed: %w", err)
	}

	report := framework.GenerateReport()
	return saveSingleTestReport(report, "response_time_test_report.json", orchestrator.GetReportPath())
}

// runEndToEndTest runs only the end-to-end test
func runEndToEndTest(ctx context.Context, orchestrator *performance.PerformanceTestOrchestrator) error {
	log.Println("Running end-to-end test...")

	config := performance.PerformanceTestConfig{
		BaseURL:          orchestrator.GetBaseURL(),
		ConcurrentUsers:  75,
		TestDuration:     30 * time.Minute,
		RampUpDuration:   3 * time.Minute,
		RequestTimeout:   30 * time.Second,
		MaxMemoryMB:      1500,
		TargetResponseMs: 300,
		ErrorThreshold:   1.5,
		ThroughputTarget: 100,
	}

	framework := performance.NewPerformanceTestFramework(config)
	scenarios := performance.NewKYBTestScenarios(config.BaseURL)
	framework.AddScenarios(scenarios.GetComprehensiveScenarios())

	_, err := framework.RunLoadTest(ctx)
	if err != nil {
		return fmt.Errorf("end-to-end test failed: %w", err)
	}

	report := framework.GenerateReport()
	return saveSingleTestReport(report, "end_to_end_test_report.json", orchestrator.GetReportPath())
}

// saveSingleTestReport saves a single test report
func saveSingleTestReport(report *performance.PerformanceTestReport, filename, reportPath string) error {
	// Create report directory if it doesn't exist
	if err := os.MkdirAll(reportPath, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	// Save report
	filePath := fmt.Sprintf("%s/%s", reportPath, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode report: %w", err)
	}

	log.Printf("Test report saved to: %s", filePath)
	return nil
}

// showHelp displays help information
func showHelp() {
	fmt.Println(`
KYB Platform Performance Testing Tool

Usage:
  performance-test [flags]

Flags:
  -base-url string
        Base URL for the KYB platform API (default "http://localhost:8080")
  -report-path string
        Path to save performance test reports (default "./performance-reports")
  -test-type string
        Type of test to run: load, stress, memory, response-time, end-to-end, comprehensive (default "comprehensive")
  -help
        Show this help information

Test Types:
  load          - Load testing with realistic data (50 concurrent users, 10 minutes)
  stress        - Stress testing under high load (200 concurrent users, 15 minutes)
  memory        - Memory usage testing (100 concurrent users, 20 minutes)
  response-time - Response time validation (25 concurrent users, 5 minutes)
  end-to-end    - End-to-end performance test (75 concurrent users, 30 minutes)
  comprehensive - Run all test types sequentially

Examples:
  # Run comprehensive performance tests
  performance-test -base-url https://api.kyb-platform.com -report-path ./reports

  # Run only load test
  performance-test -test-type load -base-url http://localhost:8080

  # Run stress test with custom report path
  performance-test -test-type stress -report-path /tmp/performance-reports

Performance Targets:
  - API response times: <200ms average
  - ML model inference: <100ms for classification, <50ms for risk detection
  - Database query performance: 50% improvement
  - System uptime: 99.9%
  - Error rate: <1%
  - Throughput: >100 req/s

The tool will generate detailed JSON reports and markdown summaries for each test.
`)
}
