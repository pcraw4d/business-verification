package performance

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMerchantPortfolioPerformance tests merchant portfolio operations with large datasets
func TestMerchantPortfolioPerformance(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.MaxMerchants = 5000
	config.ResponseTimeLimit = 3 * time.Second

	suite := NewPerformanceTestSuite(config)

	// Add merchant portfolio performance tests
	suite.AddTest(PerformanceTest{
		Name:        "MerchantPortfolioListPerformance",
		Description: "Test merchant portfolio listing with 5000 merchants",
		TestFunc:    testMerchantPortfolioListPerformance,
	})

	suite.AddTest(PerformanceTest{
		Name:        "MerchantPortfolioSearchPerformance",
		Description: "Test merchant portfolio search with 5000 merchants",
		TestFunc:    testMerchantPortfolioSearchPerformance,
	})

	suite.AddTest(PerformanceTest{
		Name:        "MerchantPortfolioFilteringPerformance",
		Description: "Test merchant portfolio filtering with 5000 merchants",
		TestFunc:    testMerchantPortfolioFilteringPerformance,
	})

	suite.AddTest(PerformanceTest{
		Name:        "MerchantPortfolioPaginationPerformance",
		Description: "Test merchant portfolio pagination with 5000 merchants",
		TestFunc:    testMerchantPortfolioPaginationPerformance,
	})

	suite.AddTest(PerformanceTest{
		Name:        "MerchantDetailViewPerformance",
		Description: "Test individual merchant detail view performance",
		TestFunc:    testMerchantDetailViewPerformance,
	})

	// Run all tests
	suite.RunAllTests(t)
}

// testMerchantPortfolioListPerformance tests listing merchants with large dataset
func testMerchantPortfolioListPerformance(t *testing.T, config *PerformanceTestConfig) error {
	helper := NewBenchmarkHelper(config)
	runner := NewPerformanceTestRunner(config)
	runner.Start()

	// Generate test data
	testData := helper.GenerateTestData(config.MaxMerchants)

	// Simulate merchant portfolio listing
	operation := func() error {
		// Simulate API call to list merchants
		start := time.Now()

		// Simulate processing time based on dataset size
		processingTime := time.Duration(len(testData)/1000) * time.Millisecond
		time.Sleep(processingTime)

		responseTime := time.Since(start)
		runner.RecordRequest(responseTime, true)

		return nil
	}

	// Run multiple iterations
	for i := 0; i < 10; i++ {
		err := operation()
		if err != nil {
			return fmt.Errorf("merchant portfolio listing failed: %w", err)
		}
	}

	runner.Stop()
	metrics := runner.GetMetrics()
	runner.PrintMetrics("Merchant Portfolio List Performance")

	// Assert performance requirements
	AssertPerformanceRequirements(t, metrics, config)

	// Specific assertions for merchant portfolio listing
	assert.LessOrEqual(t, metrics.AverageResponseTime, 2*time.Second,
		"Merchant portfolio listing should complete within 2 seconds")
	assert.GreaterOrEqual(t, metrics.RequestsPerSecond, 5.0,
		"Should handle at least 5 list requests per second")

	return nil
}

// testMerchantPortfolioSearchPerformance tests merchant search with large dataset
func testMerchantPortfolioSearchPerformance(t *testing.T, config *PerformanceTestConfig) error {
	helper := NewBenchmarkHelper(config)
	runner := NewPerformanceTestRunner(config)
	runner.Start()

	// Generate test data
	testData := helper.GenerateTestData(config.MaxMerchants)

	// Test different search scenarios
	searchQueries := []string{
		"Test Merchant",
		"Technology",
		"merchant_100",
		"@test.com",
		"555-",
	}

	for _, query := range searchQueries {
		operation := func() error {
			start := time.Now()

			// Simulate search operation
			results := 0
			for _, merchant := range testData {
				// Simple string matching simulation
				if contains(merchant["name"].(string), query) ||
					contains(merchant["email"].(string), query) ||
					contains(merchant["phone"].(string), query) {
					results++
				}
			}

			// Simulate additional processing time
			processingTime := time.Duration(results/100) * time.Millisecond
			time.Sleep(processingTime)

			responseTime := time.Since(start)
			runner.RecordRequest(responseTime, true)

			return nil
		}

		// Run multiple iterations per query
		for i := 0; i < 5; i++ {
			err := operation()
			if err != nil {
				return fmt.Errorf("merchant search failed for query '%s': %w", query, err)
			}
		}
	}

	runner.Stop()
	metrics := runner.GetMetrics()
	runner.PrintMetrics("Merchant Portfolio Search Performance")

	// Assert performance requirements
	AssertPerformanceRequirements(t, metrics, config)

	// Specific assertions for search performance
	assert.LessOrEqual(t, metrics.AverageResponseTime, 1*time.Second,
		"Merchant search should complete within 1 second")
	assert.GreaterOrEqual(t, metrics.RequestsPerSecond, 10.0,
		"Should handle at least 10 search requests per second")

	return nil
}

// testMerchantPortfolioFilteringPerformance tests merchant filtering with large dataset
func testMerchantPortfolioFilteringPerformance(t *testing.T, config *PerformanceTestConfig) error {
	helper := NewBenchmarkHelper(config)
	runner := NewPerformanceTestRunner(config)
	runner.Start()

	// Generate test data
	testData := helper.GenerateTestData(config.MaxMerchants)

	// Test different filter scenarios
	filterScenarios := []struct {
		name   string
		filter func(map[string]interface{}) bool
	}{
		{
			name: "Portfolio Type Filter",
			filter: func(merchant map[string]interface{}) bool {
				return merchant["portfolio_type"] == "onboarded"
			},
		},
		{
			name: "Risk Level Filter",
			filter: func(merchant map[string]interface{}) bool {
				return merchant["risk_level"] == "medium"
			},
		},
		{
			name: "Industry Filter",
			filter: func(merchant map[string]interface{}) bool {
				return merchant["industry"] == "Technology"
			},
		},
		{
			name: "Combined Filter",
			filter: func(merchant map[string]interface{}) bool {
				return merchant["portfolio_type"] == "onboarded" &&
					merchant["risk_level"] == "medium" &&
					merchant["industry"] == "Technology"
			},
		},
	}

	for _, scenario := range filterScenarios {
		operation := func() error {
			start := time.Now()

			// Simulate filtering operation
			results := 0
			for _, merchant := range testData {
				if scenario.filter(merchant) {
					results++
				}
			}

			// Simulate additional processing time
			processingTime := time.Duration(results/100) * time.Millisecond
			time.Sleep(processingTime)

			responseTime := time.Since(start)
			runner.RecordRequest(responseTime, true)

			return nil
		}

		// Run multiple iterations per filter scenario
		for i := 0; i < 5; i++ {
			err := operation()
			if err != nil {
				return fmt.Errorf("merchant filtering failed for scenario '%s': %w", scenario.name, err)
			}
		}
	}

	runner.Stop()
	metrics := runner.GetMetrics()
	runner.PrintMetrics("Merchant Portfolio Filtering Performance")

	// Assert performance requirements
	AssertPerformanceRequirements(t, metrics, config)

	// Specific assertions for filtering performance
	assert.LessOrEqual(t, metrics.AverageResponseTime, 1*time.Second,
		"Merchant filtering should complete within 1 second")
	assert.GreaterOrEqual(t, metrics.RequestsPerSecond, 15.0,
		"Should handle at least 15 filter requests per second")

	return nil
}

// testMerchantPortfolioPaginationPerformance tests merchant pagination with large dataset
func testMerchantPortfolioPaginationPerformance(t *testing.T, config *PerformanceTestConfig) error {
	helper := NewBenchmarkHelper(config)
	runner := NewPerformanceTestRunner(config)
	runner.Start()

	// Generate test data
	testData := helper.GenerateTestData(config.MaxMerchants)

	// Test pagination scenarios
	pageSizes := []int{10, 25, 50, 100}
	totalPages := len(testData) / 50 // Assume 50 items per page for testing

	for _, pageSize := range pageSizes {
		// Test first page
		operation := func() error {
			start := time.Now()

			// Simulate pagination operation
			startIndex := 0
			endIndex := startIndex + pageSize
			if endIndex > len(testData) {
				endIndex = len(testData)
			}

			// Simulate processing paginated results
			pageData := testData[startIndex:endIndex]
			processingTime := time.Duration(len(pageData)/10) * time.Millisecond
			time.Sleep(processingTime)

			responseTime := time.Since(start)
			runner.RecordRequest(responseTime, true)

			return nil
		}

		// Run multiple iterations per page size
		for i := 0; i < 5; i++ {
			err := operation()
			if err != nil {
				return fmt.Errorf("merchant pagination failed for page size %d: %w", pageSize, err)
			}
		}
	}

	// Test middle and last pages
	for page := 1; page < totalPages && page < 10; page++ {
		operation := func() error {
			start := time.Now()

			// Simulate pagination operation
			startIndex := page * 50
			endIndex := startIndex + 50
			if endIndex > len(testData) {
				endIndex = len(testData)
			}

			// Simulate processing paginated results
			pageData := testData[startIndex:endIndex]
			processingTime := time.Duration(len(pageData)/10) * time.Millisecond
			time.Sleep(processingTime)

			responseTime := time.Since(start)
			runner.RecordRequest(responseTime, true)

			return nil
		}

		err := operation()
		if err != nil {
			return fmt.Errorf("merchant pagination failed for page %d: %w", page, err)
		}
	}

	runner.Stop()
	metrics := runner.GetMetrics()
	runner.PrintMetrics("Merchant Portfolio Pagination Performance")

	// Assert performance requirements
	AssertPerformanceRequirements(t, metrics, config)

	// Specific assertions for pagination performance
	assert.LessOrEqual(t, metrics.AverageResponseTime, 500*time.Millisecond,
		"Merchant pagination should complete within 500ms")
	assert.GreaterOrEqual(t, metrics.RequestsPerSecond, 20.0,
		"Should handle at least 20 pagination requests per second")

	return nil
}

// testMerchantDetailViewPerformance tests individual merchant detail view
func testMerchantDetailViewPerformance(t *testing.T, config *PerformanceTestConfig) error {
	helper := NewBenchmarkHelper(config)
	runner := NewPerformanceTestRunner(config)
	runner.Start()

	// Generate test data
	testData := helper.GenerateTestData(100) // Smaller dataset for detail views

	// Test merchant detail view performance
	for _, merchant := range testData {
		operation := func() error {
			start := time.Now()

			// Simulate merchant detail view operation
			// This would typically involve:
			// - Fetching merchant data
			// - Loading related data (transactions, compliance, etc.)
			// - Rendering the view

			// Simulate processing time
			processingTime := 50 * time.Millisecond
			time.Sleep(processingTime)

			responseTime := time.Since(start)
			runner.RecordRequest(responseTime, true)

			return nil
		}

		err := operation()
		if err != nil {
			return fmt.Errorf("merchant detail view failed for merchant %s: %w", merchant["id"], err)
		}
	}

	runner.Stop()
	metrics := runner.GetMetrics()
	runner.PrintMetrics("Merchant Detail View Performance")

	// Assert performance requirements
	AssertPerformanceRequirements(t, metrics, config)

	// Specific assertions for detail view performance
	assert.LessOrEqual(t, metrics.AverageResponseTime, 200*time.Millisecond,
		"Merchant detail view should complete within 200ms")
	assert.GreaterOrEqual(t, metrics.RequestsPerSecond, 50.0,
		"Should handle at least 50 detail view requests per second")

	return nil
}

// BenchmarkMerchantPortfolioOperations benchmarks merchant portfolio operations
func BenchmarkMerchantPortfolioOperations(b *testing.B) {
	config := DefaultPerformanceConfig()
	config.MaxMerchants = 1000

	helper := NewBenchmarkHelper(config)
	testData := helper.GenerateTestData(config.MaxMerchants)

	b.ResetTimer()

	b.Run("ListMerchants", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate listing merchants
			_ = len(testData)
		}
	})

	b.Run("SearchMerchants", func(b *testing.B) {
		query := "Test Merchant"
		for i := 0; i < b.N; i++ {
			// Simulate searching merchants
			results := 0
			for _, merchant := range testData {
				if contains(merchant["name"].(string), query) {
					results++
				}
			}
			_ = results
		}
	})

	b.Run("FilterMerchants", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate filtering merchants
			results := 0
			for _, merchant := range testData {
				if merchant["portfolio_type"] == "onboarded" {
					results++
				}
			}
			_ = results
		}
	})
}
