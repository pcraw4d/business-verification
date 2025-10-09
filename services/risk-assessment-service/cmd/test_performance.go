package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/testing"
)

func main() {
	var verbose = flag.Bool("v", false, "verbose output")
	var single = flag.Bool("single", false, "run single request test only")
	var concurrent = flag.Bool("concurrent", false, "run concurrent request test only")
	var batch = flag.Bool("batch", false, "run batch request test only")
	var cache = flag.Bool("cache", false, "run cache test only")
	var metrics = flag.Bool("metrics", false, "run metrics test only")
	flag.Parse()

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("🚀 Starting Risk Assessment Service Performance Tests")

	// Create performance test suite
	testSuite := testing.NewPerformanceTestSuite()
	defer testSuite.Close()

	logger.Info("✅ Performance test suite initialized")

	// Create test instance
	t := &testing.T{}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("RISK ASSESSMENT SERVICE PERFORMANCE TESTS")
	fmt.Println(strings.Repeat("=", 60))

	if *single {
		fmt.Println("\n🔍 Running Single Request Performance Test...")
		testSuite.TestSingleRequestPerformance(t)
	} else if *concurrent {
		fmt.Println("\n🔍 Running Concurrent Request Performance Test...")
		testSuite.TestConcurrentRequestPerformance(t)
	} else if *batch {
		fmt.Println("\n🔍 Running Batch Request Performance Test...")
		testSuite.TestBatchRequestPerformance(t)
	} else if *cache {
		fmt.Println("\n🔍 Running Cache Performance Test...")
		testSuite.TestCachePerformance(t)
	} else if *metrics {
		fmt.Println("\n🔍 Running Metrics Endpoint Performance Test...")
		testSuite.TestMetricsEndpoint(t)
	} else {
		fmt.Println("\n🔍 Running All Performance Tests...")
		testSuite.RunAllPerformanceTests(t)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("PERFORMANCE TESTS COMPLETED")
	fmt.Println(strings.Repeat("=", 60))

	if t.Failed() {
		fmt.Println("❌ Some tests failed - check output above")
		os.Exit(1)
	} else {
		fmt.Println("✅ All performance tests passed!")
		fmt.Println("🎯 Sub-1-second response time target achieved!")
	}
}
