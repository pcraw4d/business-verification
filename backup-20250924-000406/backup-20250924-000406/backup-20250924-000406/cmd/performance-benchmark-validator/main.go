package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pcraw4d/business-verification/test"
)

func main() {
	// Command line flags
	var (
		configFile = flag.String("config", "", "Path to performance benchmark configuration file (JSON)")
		verbose    = flag.Bool("verbose", false, "Enable verbose output")
		help       = flag.Bool("help", false, "Show help information")
	)

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Load configuration
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Printf("Warning: Failed to parse config file %s: %v", *configFile, err)
		config = getDefaultConfig()
	}

	if *verbose {
		printConfig(config)
	}

	// Create logger
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Create test runner
	mockRepo := &test.MockKeywordRepository{}
	testRunner := test.NewClassificationAccuracyTestRunner(mockRepo, logger)

	// Create validator
	validator := test.NewPerformanceBenchmarkingValidator(testRunner, logger, config)

	logger.Printf("⚡ Starting Performance Benchmarking...")

	// Run benchmarking
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	result, err := validator.RunPerformanceBenchmarking(ctx)
	if err != nil {
		logger.Fatalf("❌ Performance benchmarking failed: %v", err)
	}

	// Save report
	if err := validator.SavePerformanceReport(result); err != nil {
		logger.Printf("⚠️  Failed to save report: %v", err)
	}

	// Print summary
	printPerformanceSummary(result, *verbose)

	logger.Printf("✅ Performance Benchmarking Completed!")
}

func loadConfig(configFile string) (*test.PerformanceBenchmarkingConfig, error) {
	if configFile == "" {
		return getDefaultConfig(), nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config test.PerformanceBenchmarkingConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func getDefaultConfig() *test.PerformanceBenchmarkingConfig {
	return &test.PerformanceBenchmarkingConfig{
		SessionName:               "Performance Benchmarking Session",
		BenchmarkDirectory:        "./performance-benchmark",
		SampleSize:                50,
		Timeout:                   30 * time.Minute,
		ConcurrencyLevels:         []int{1, 2, 4, 8, 16, 32},
		LoadTestDuration:          30 * time.Second,
		StressTestDuration:        60 * time.Second,
		IncludeMemoryProfiling:    true,
		IncludeCPUProfiling:       true,
		IncludeThroughputTesting:  true,
		IncludeLatencyTesting:     true,
		IncludeScalabilityTesting: true,
		IncludeResourceMonitoring: true,
		GenerateDetailedReport:    true,
	}
}

func printConfig(config *test.PerformanceBenchmarkingConfig) {
	fmt.Printf("📋 Performance Benchmarking Configuration:\n")
	fmt.Printf("   Session Name: %s\n", config.SessionName)
	fmt.Printf("   Benchmark Directory: %s\n", config.BenchmarkDirectory)
	fmt.Printf("   Sample Size: %d\n", config.SampleSize)
	fmt.Printf("   Timeout: %v\n", config.Timeout)
	fmt.Printf("   Concurrency Levels: %v\n", config.ConcurrencyLevels)
	fmt.Printf("   Load Test Duration: %v\n", config.LoadTestDuration)
	fmt.Printf("   Stress Test Duration: %v\n", config.StressTestDuration)
	fmt.Printf("   Memory Profiling: %t\n", config.IncludeMemoryProfiling)
	fmt.Printf("   CPU Profiling: %t\n", config.IncludeCPUProfiling)
	fmt.Printf("   Throughput Testing: %t\n", config.IncludeThroughputTesting)
	fmt.Printf("   Latency Testing: %t\n", config.IncludeLatencyTesting)
	fmt.Printf("   Scalability Testing: %t\n", config.IncludeScalabilityTesting)
	fmt.Printf("   Resource Monitoring: %t\n", config.IncludeResourceMonitoring)
	fmt.Printf("   Detailed Report: %t\n", config.GenerateDetailedReport)
	fmt.Println()
}

func printPerformanceSummary(result *test.PerformanceBenchmarkingResult, verbose bool) {
	fmt.Printf("🏁 Performance Benchmarking Completed!\n")
	fmt.Printf("⏱️  Duration: %v\n", result.Duration)
	fmt.Printf("📊 Session ID: %s\n", result.SessionID)
	fmt.Printf("📁 Benchmark Directory: %s\n", result.SessionID)
	fmt.Printf("📋 Performance Summary:\n")

	if result.PerformanceSummary != nil {
		fmt.Printf("   Overall Performance: %.3f\n", result.PerformanceSummary.OverallPerformance)
		fmt.Printf("   Performance Grade: %s\n", result.PerformanceSummary.PerformanceGrade)
		fmt.Printf("   Is Performance Acceptable: %t\n", result.PerformanceSummary.IsPerformanceAcceptable)
		fmt.Printf("   Average Response Time: %.2f ms\n", result.PerformanceSummary.AverageResponseTime)
		fmt.Printf("   Peak Throughput: %.2f req/sec\n", result.PerformanceSummary.PeakThroughput)
		fmt.Printf("   Average Throughput: %.2f req/sec\n", result.PerformanceSummary.AverageThroughput)
		fmt.Printf("   Max Concurrency: %d\n", result.PerformanceSummary.MaxConcurrency)
		fmt.Printf("   Memory Usage: %.2f MB\n", result.PerformanceSummary.MemoryUsage)
		fmt.Printf("   CPU Usage: %.2f%%\n", result.PerformanceSummary.CPUUsage)
		fmt.Printf("   Error Rate: %.3f%%\n", result.PerformanceSummary.ErrorRate*100)
	}

	if result.ThroughputResults != nil {
		fmt.Printf("📈 Throughput Results:\n")
		fmt.Printf("   Max Throughput: %.2f req/sec\n", result.ThroughputResults.MaxThroughput)
		fmt.Printf("   Average Throughput: %.2f req/sec\n", result.ThroughputResults.AverageThroughput)
		fmt.Printf("   Throughput Stability: %.3f\n", result.ThroughputResults.ThroughputStability)
	}

	if result.LatencyResults != nil {
		fmt.Printf("⏱️  Latency Results:\n")
		fmt.Printf("   Min Latency: %.2f ms\n", result.LatencyResults.MinLatency)
		fmt.Printf("   Max Latency: %.2f ms\n", result.LatencyResults.MaxLatency)
		fmt.Printf("   Average Latency: %.2f ms\n", result.LatencyResults.AverageLatency)
		fmt.Printf("   P95 Latency: %.2f ms\n", result.LatencyResults.P95Latency)
		fmt.Printf("   P99 Latency: %.2f ms\n", result.LatencyResults.P99Latency)
		fmt.Printf("   P999 Latency: %.2f ms\n", result.LatencyResults.P999Latency)
	}

	if result.ScalabilityResults != nil {
		fmt.Printf("📊 Scalability Results:\n")
		fmt.Printf("   Linear Scalability: %t\n", result.ScalabilityResults.LinearScalability)
		fmt.Printf("   Scalability Factor: %.3f\n", result.ScalabilityResults.ScalabilityFactor)
		fmt.Printf("   Max Scalable Concurrency: %d\n", result.ScalabilityResults.MaxScalableConcurrency)
		fmt.Printf("   Performance Degradation Point: %d\n", result.ScalabilityResults.PerformanceDegradationPoint)
		fmt.Printf("   Scalability Efficiency: %.3f\n", result.ScalabilityResults.ScalabilityEfficiency)
		fmt.Printf("   Bottleneck Identification: %s\n", result.ScalabilityResults.BottleneckIdentification)
	}

	if result.ResourceUsageResults != nil {
		fmt.Printf("💾 Resource Usage Results:\n")
		fmt.Printf("   Peak Memory Usage: %.2f MB\n", result.ResourceUsageResults.PeakMemoryUsage)
		fmt.Printf("   Average Memory Usage: %.2f MB\n", result.ResourceUsageResults.AverageMemoryUsage)
		fmt.Printf("   Memory Leak Detected: %t\n", result.ResourceUsageResults.MemoryLeakDetected)
		fmt.Printf("   Peak CPU Usage: %.2f%%\n", result.ResourceUsageResults.PeakCPUUsage)
		fmt.Printf("   Average CPU Usage: %.2f%%\n", result.ResourceUsageResults.AverageCPUUsage)
		fmt.Printf("   CPU Utilization Efficiency: %.3f\n", result.ResourceUsageResults.CPUUtilizationEfficiency)
		fmt.Printf("   GC Pause Time: %.3f s\n", result.ResourceUsageResults.GCPauseTime)
		fmt.Printf("   GC Pause Count: %d\n", result.ResourceUsageResults.GCPauseCount)
		fmt.Printf("   Resource Utilization Grade: %s\n", result.ResourceUsageResults.ResourceUtilizationGrade)
	}

	if result.LoadTestResults != nil {
		fmt.Printf("🔥 Load Test Results:\n")
		fmt.Printf("   Load Test Duration: %v\n", result.LoadTestResults.LoadTestDuration)
		fmt.Printf("   Total Requests: %d\n", result.LoadTestResults.TotalRequests)
		fmt.Printf("   Successful Requests: %d\n", result.LoadTestResults.SuccessfulRequests)
		fmt.Printf("   Failed Requests: %d\n", result.LoadTestResults.FailedRequests)
		fmt.Printf("   Requests Per Second: %.2f\n", result.LoadTestResults.RequestsPerSecond)
		fmt.Printf("   Average Response Time: %.2f ms\n", result.LoadTestResults.AverageResponseTime)
		fmt.Printf("   Error Rate: %.3f%%\n", result.LoadTestResults.ErrorRate*100)
		fmt.Printf("   Load Test Stability: %.3f\n", result.LoadTestResults.LoadTestStability)
	}

	if result.StressTestResults != nil {
		fmt.Printf("💥 Stress Test Results:\n")
		fmt.Printf("   Stress Test Duration: %v\n", result.StressTestResults.StressTestDuration)
		fmt.Printf("   Breaking Point: %d\n", result.StressTestResults.BreakingPoint)
		fmt.Printf("   Recovery Time: %v\n", result.StressTestResults.RecoveryTime)
		fmt.Printf("   Stress Test Stability: %.3f\n", result.StressTestResults.StressTestStability)
		fmt.Printf("   Failure Mode: %s\n", result.StressTestResults.FailureMode)
		fmt.Printf("   Recovery Capability: %s\n", result.StressTestResults.RecoveryCapability)
	}

	if verbose && len(result.ConcurrencyResults) > 0 {
		fmt.Printf("🔄 Concurrency Results:\n")
		for _, result := range result.ConcurrencyResults {
			fmt.Printf("   Concurrency %d: %.2f req/sec, %.2f ms latency, %.3f%% error rate\n",
				result.ConcurrencyLevel, result.Throughput, result.AverageLatency, result.ErrorRate*100)
		}
	}

	if len(result.Recommendations) > 0 {
		fmt.Printf("💡 Recommendations: %d suggestions generated\n", len(result.Recommendations))
		fmt.Printf("📋 Recommendations:\n")
		for i, rec := range result.Recommendations {
			fmt.Printf("   %d. %s\n", i+1, rec)
		}
	}

	fmt.Printf("📁 Benchmark files generated in: %s\n", result.SessionID)
	fmt.Printf("📄 Reports available:\n")
	fmt.Printf("   - performance_benchmark_report.json (comprehensive JSON report)\n")
	fmt.Printf("   - performance_benchmark_report.html (human-readable HTML report)\n")
	fmt.Printf("   - performance_benchmark_summary.json (session summary)\n")
}

func showHelp() {
	fmt.Printf("Performance Benchmarking Validator\n")
	fmt.Printf("==================================\n\n")
	fmt.Printf("Usage: performance-benchmark-validator [options]\n\n")
	fmt.Printf("Options:\n")
	fmt.Printf("  -config string\n")
	fmt.Printf("        Path to performance benchmark configuration file (JSON)\n")
	fmt.Printf("  -verbose\n")
	fmt.Printf("        Enable verbose output\n")
	fmt.Printf("  -help\n")
	fmt.Printf("        Show this help information\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  performance-benchmark-validator\n")
	fmt.Printf("  performance-benchmark-validator -config configs/performance-config.json\n")
	fmt.Printf("  performance-benchmark-validator -verbose\n\n")
	fmt.Printf("Configuration File Format:\n")
	fmt.Printf("  {\n")
	fmt.Printf("    \"session_name\": \"Performance Benchmarking Session\",\n")
	fmt.Printf("    \"benchmark_directory\": \"./performance-benchmark\",\n")
	fmt.Printf("    \"sample_size\": 50,\n")
	fmt.Printf("    \"timeout\": \"30m\",\n")
	fmt.Printf("    \"concurrency_levels\": [1, 2, 4, 8, 16, 32],\n")
	fmt.Printf("    \"load_test_duration\": \"30s\",\n")
	fmt.Printf("    \"stress_test_duration\": \"60s\",\n")
	fmt.Printf("    \"include_memory_profiling\": true,\n")
	fmt.Printf("    \"include_cpu_profiling\": true,\n")
	fmt.Printf("    \"include_throughput_testing\": true,\n")
	fmt.Printf("    \"include_latency_testing\": true,\n")
	fmt.Printf("    \"include_scalability_testing\": true,\n")
	fmt.Printf("    \"include_resource_monitoring\": true,\n")
	fmt.Printf("    \"generate_detailed_report\": true\n")
	fmt.Printf("  }\n")
}
