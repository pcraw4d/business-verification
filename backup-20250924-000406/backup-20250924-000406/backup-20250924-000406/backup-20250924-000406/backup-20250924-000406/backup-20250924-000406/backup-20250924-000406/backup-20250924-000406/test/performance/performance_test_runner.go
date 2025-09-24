package performance

import (
	"log"
	"os"
	"testing"
	"time"
)

// PerformanceTestRunner manages the execution of all performance tests
type PerformanceTestRunner struct {
	config    *PerformanceTestConfig
	outputDir string
}

// NewPerformanceTestRunner creates a new performance test runner
func NewPerformanceTestRunner(config *PerformanceTestConfig, outputDir string) *PerformanceTestRunner {
	return &PerformanceTestRunner{
		config:    config,
		outputDir: outputDir,
	}
}

// RunAllPerformanceTests runs all performance tests and generates reports
func (ptr *PerformanceTestRunner) RunAllPerformanceTests(t *testing.T) error {
	log.Printf("Starting performance test suite with configuration:")
	log.Printf("- Concurrent Users: %d", ptr.config.ConcurrentUsers)
	log.Printf("- Max Merchants: %d", ptr.config.MaxMerchants)
	log.Printf("- Test Duration: %v", ptr.config.TestDuration)
	log.Printf("- Response Time Limit: %v", ptr.config.ResponseTimeLimit)
	log.Printf("- Bulk Operation Size: %d", ptr.config.BulkOperationSize)

	// Run merchant portfolio performance tests
	if err := ptr.runMerchantPortfolioTests(t); err != nil {
		log.Printf("Merchant portfolio tests failed: %v", err)
	}

	// Run bulk operations performance tests
	if err := ptr.runBulkOperationsTests(t); err != nil {
		log.Printf("Bulk operations tests failed: %v", err)
	}

	// Run concurrent user performance tests
	if err := ptr.runConcurrentUserTests(t); err != nil {
		log.Printf("Concurrent user tests failed: %v", err)
	}

	return nil
}

// runMerchantPortfolioTests runs merchant portfolio performance tests
func (ptr *PerformanceTestRunner) runMerchantPortfolioTests(t *testing.T) error {
	log.Printf("Running merchant portfolio performance tests...")

	testSuite := NewPerformanceTestSuite(ptr.config)

	// Add merchant portfolio tests
	testSuite.AddTest(PerformanceTest{
		Name:        "MerchantPortfolioListPerformance",
		Description: "Test merchant portfolio listing with large datasets",
		TestFunc:    testMerchantPortfolioListPerformance,
	})

	testSuite.AddTest(PerformanceTest{
		Name:        "MerchantPortfolioSearchPerformance",
		Description: "Test merchant portfolio search with large datasets",
		TestFunc:    testMerchantPortfolioSearchPerformance,
	})

	testSuite.AddTest(PerformanceTest{
		Name:        "MerchantPortfolioFilteringPerformance",
		Description: "Test merchant portfolio filtering with large datasets",
		TestFunc:    testMerchantPortfolioFilteringPerformance,
	})

	testSuite.AddTest(PerformanceTest{
		Name:        "MerchantPortfolioPaginationPerformance",
		Description: "Test merchant portfolio pagination with large datasets",
		TestFunc:    testMerchantPortfolioPaginationPerformance,
	})

	testSuite.AddTest(PerformanceTest{
		Name:        "MerchantDetailViewPerformance",
		Description: "Test individual merchant detail view performance",
		TestFunc:    testMerchantDetailViewPerformance,
	})

	// Run tests
	testSuite.RunAllTests(t)

	return nil
}

// runBulkOperationsTests runs bulk operations performance tests
func (ptr *PerformanceTestRunner) runBulkOperationsTests(t *testing.T) error {
	log.Printf("Running bulk operations performance tests...")

	testSuite := NewPerformanceTestSuite(ptr.config)

	// Add bulk operations tests
	testSuite.AddTest(PerformanceTest{
		Name:        "BulkMerchantUpdatePerformance",
		Description: "Test bulk merchant update operations",
		TestFunc:    testBulkMerchantUpdatePerformance,
	})

	testSuite.AddTest(PerformanceTest{
		Name:        "BulkMerchantStatusChangePerformance",
		Description: "Test bulk merchant status change operations",
		TestFunc:    testBulkMerchantStatusChangePerformance,
	})

	testSuite.AddTest(PerformanceTest{
		Name:        "BulkMerchantExportPerformance",
		Description: "Test bulk merchant export operations",
		TestFunc:    testBulkMerchantExportPerformance,
	})

	testSuite.AddTest(PerformanceTest{
		Name:        "BulkMerchantImportPerformance",
		Description: "Test bulk merchant import operations",
		TestFunc:    testBulkMerchantImportPerformance,
	})

	testSuite.AddTest(PerformanceTest{
		Name:        "BulkMerchantDeletionPerformance",
		Description: "Test bulk merchant deletion operations",
		TestFunc:    testBulkMerchantDeletionPerformance,
	})

	// Run tests
	testSuite.RunAllTests(t)

	return nil
}

// runConcurrentUserTests runs concurrent user performance tests
func (ptr *PerformanceTestRunner) runConcurrentUserTests(t *testing.T) error {
	log.Printf("Running concurrent user performance tests...")

	testSuite := NewPerformanceTestSuite(ptr.config)

	// Add concurrent user tests
	testSuite.AddTest(PerformanceTest{
		Name:        "ConcurrentUserPortfolioAccess",
		Description: "Test concurrent user portfolio access",
		TestFunc:    testConcurrentUserPortfolioAccess,
	})

	testSuite.AddTest(PerformanceTest{
		Name:        "ConcurrentUserMerchantDetailView",
		Description: "Test concurrent user merchant detail views",
		TestFunc:    testConcurrentUserMerchantDetailView,
	})

	testSuite.AddTest(PerformanceTest{
		Name:        "ConcurrentUserSearchOperations",
		Description: "Test concurrent user search operations",
		TestFunc:    testConcurrentUserSearchOperations,
	})

	testSuite.AddTest(PerformanceTest{
		Name:        "ConcurrentUserBulkOperations",
		Description: "Test concurrent user bulk operations",
		TestFunc:    testConcurrentUserBulkOperations,
	})

	testSuite.AddTest(PerformanceTest{
		Name:        "ConcurrentUserSessionManagement",
		Description: "Test concurrent user session management",
		TestFunc:    testConcurrentUserSessionManagement,
	})

	// Run tests
	testSuite.RunAllTests(t)

	return nil
}

// Main performance test function
func TestPerformanceSuite(t *testing.T) {
	// Create performance test configuration
	config := DefaultPerformanceConfig()
	config.MaxMerchants = 5000
	config.ConcurrentUsers = 20
	config.TestDuration = 2 * time.Minute
	config.ResponseTimeLimit = 3 * time.Second
	config.BulkOperationSize = 1000

	// Create output directory
	outputDir := "test-results/performance"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create and run performance test runner
	runner := NewPerformanceTestRunner(config, outputDir)

	// Run all performance tests
	if err := runner.RunAllPerformanceTests(t); err != nil {
		t.Errorf("Performance tests failed: %v", err)
	}
}

// TestPerformanceBenchmarks runs performance benchmarks
func TestPerformanceBenchmarks(t *testing.T) {
	config := DefaultPerformanceConfig()
	outputDir := "test-results/benchmarks"

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	_ = NewPerformanceTestRunner(config, outputDir)

	// Run basic benchmarks
	helper := NewBenchmarkHelper(config)
	testData := helper.GenerateTestData(1000)

	// Benchmark list operations
	t.Run("BenchmarkListMerchants", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 100; i++ {
			simulateListMerchants(testData)
		}
		duration := time.Since(start)
		t.Logf("List merchants benchmark: %v for 100 iterations", duration)
	})

	// Benchmark search operations
	t.Run("BenchmarkSearchMerchants", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 100; i++ {
			simulateSearchOperation(testData, "Test")
		}
		duration := time.Since(start)
		t.Logf("Search merchants benchmark: %v for 100 iterations", duration)
	})

	// Benchmark filter operations
	t.Run("BenchmarkFilterMerchants", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 100; i++ {
			simulateFilterMerchants(testData)
		}
		duration := time.Since(start)
		t.Logf("Filter merchants benchmark: %v for 100 iterations", duration)
	})
}
