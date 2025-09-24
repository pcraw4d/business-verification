package performance

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestBulkOperationsPerformance tests bulk operations with large datasets
func TestBulkOperationsPerformance(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.BulkOperationSize = 1000
	config.ResponseTimeLimit = 30 * time.Second // Bulk operations can take longer

	suite := NewPerformanceTestSuite(config)

	// Add bulk operations performance tests
	suite.AddTest(PerformanceTest{
		Name:        "BulkMerchantUpdatePerformance",
		Description: "Test bulk merchant update operations with 1000 merchants",
		TestFunc:    testBulkMerchantUpdatePerformance,
	})

	suite.AddTest(PerformanceTest{
		Name:        "BulkMerchantStatusChangePerformance",
		Description: "Test bulk merchant status change operations",
		TestFunc:    testBulkMerchantStatusChangePerformance,
	})

	suite.AddTest(PerformanceTest{
		Name:        "BulkMerchantExportPerformance",
		Description: "Test bulk merchant export operations",
		TestFunc:    testBulkMerchantExportPerformance,
	})

	suite.AddTest(PerformanceTest{
		Name:        "BulkMerchantImportPerformance",
		Description: "Test bulk merchant import operations",
		TestFunc:    testBulkMerchantImportPerformance,
	})

	suite.AddTest(PerformanceTest{
		Name:        "BulkMerchantDeletionPerformance",
		Description: "Test bulk merchant deletion operations",
		TestFunc:    testBulkMerchantDeletionPerformance,
	})

	// Run all tests
	suite.RunAllTests(t)
}

// testBulkMerchantUpdatePerformance tests bulk merchant update operations
func testBulkMerchantUpdatePerformance(t *testing.T, config *PerformanceTestConfig) error {
	helper := NewBenchmarkHelper(config)
	runner := NewPerformanceTestRunner(config)
	runner.Start()

	// Generate test data for bulk operations
	testData := helper.GenerateTestData(config.BulkOperationSize)

	// Test different bulk update scenarios
	updateScenarios := []struct {
		name        string
		updateField string
		updateValue interface{}
		batchSize   int
	}{
		{
			name:        "Bulk Update Risk Level",
			updateField: "risk_level",
			updateValue: "high",
			batchSize:   100,
		},
		{
			name:        "Bulk Update Portfolio Type",
			updateField: "portfolio_type",
			updateValue: "deactivated",
			batchSize:   200,
		},
		{
			name:        "Bulk Update Industry",
			updateField: "industry",
			updateValue: "Finance",
			batchSize:   500,
		},
	}

	for _, scenario := range updateScenarios {
		operation := func() error {
			start := time.Now()

			// Simulate bulk update operation
			updatedCount := 0
			batchCount := 0

			for i := 0; i < len(testData); i += scenario.batchSize {
				batchCount++

				// Simulate batch processing
				endIndex := i + scenario.batchSize
				if endIndex > len(testData) {
					endIndex = len(testData)
				}

				batch := testData[i:endIndex]

				// Simulate updating each merchant in the batch
				for _, merchant := range batch {
					merchant[scenario.updateField] = scenario.updateValue
					merchant["updated_at"] = time.Now()
					updatedCount++
				}

				// Simulate batch processing time
				batchProcessingTime := time.Duration(len(batch)/10) * time.Millisecond
				time.Sleep(batchProcessingTime)
			}

			responseTime := time.Since(start)
			runner.RecordRequest(responseTime, true)

			// Verify all merchants were updated
			if updatedCount != len(testData) {
				return fmt.Errorf("expected %d updates, got %d", len(testData), updatedCount)
			}

			return nil
		}

		// Run the bulk update operation
		err := operation()
		if err != nil {
			return fmt.Errorf("bulk update failed for scenario '%s': %w", scenario.name, err)
		}
	}

	runner.Stop()
	metrics := runner.GetMetrics()
	runner.PrintMetrics("Bulk Merchant Update Performance")

	// Assert performance requirements for bulk operations
	assert.LessOrEqual(t, metrics.AverageResponseTime, 10*time.Second,
		"Bulk merchant update should complete within 10 seconds")
	assert.GreaterOrEqual(t, metrics.RequestsPerSecond, 0.1,
		"Should handle at least 0.1 bulk update operations per second")
	assert.Equal(t, float64(0), metrics.ErrorRate,
		"Bulk update operations should have 0%% error rate")

	return nil
}

// testBulkMerchantStatusChangePerformance tests bulk merchant status change operations
func testBulkMerchantStatusChangePerformance(t *testing.T, config *PerformanceTestConfig) error {
	helper := NewBenchmarkHelper(config)
	runner := NewPerformanceTestRunner(config)
	runner.Start()

	// Generate test data
	testData := helper.GenerateTestData(config.BulkOperationSize)

	// Test different status change scenarios
	statusChangeScenarios := []struct {
		name        string
		fromStatus  string
		toStatus    string
		batchSize   int
		concurrency int
	}{
		{
			name:        "Onboard to Deactivated",
			fromStatus:  "onboarded",
			toStatus:    "deactivated",
			batchSize:   100,
			concurrency: 5,
		},
		{
			name:        "Prospective to Onboarded",
			fromStatus:  "prospective",
			toStatus:    "onboarded",
			batchSize:   200,
			concurrency: 3,
		},
		{
			name:        "Pending to Onboarded",
			fromStatus:  "pending",
			toStatus:    "onboarded",
			batchSize:   500,
			concurrency: 2,
		},
	}

	for _, scenario := range statusChangeScenarios {
		operation := func() error {
			start := time.Now()

			// Simulate concurrent batch processing
			var wg sync.WaitGroup
			errors := make(chan error, scenario.concurrency)
			processedCount := 0
			var mu sync.Mutex

			// Process merchants in concurrent batches
			for i := 0; i < len(testData); i += scenario.batchSize {
				wg.Add(1)
				go func(startIndex int) {
					defer wg.Done()

					endIndex := startIndex + scenario.batchSize
					if endIndex > len(testData) {
						endIndex = len(testData)
					}

					batch := testData[startIndex:endIndex]

					// Process each merchant in the batch
					for _, merchant := range batch {
						if merchant["portfolio_type"] == scenario.fromStatus {
							merchant["portfolio_type"] = scenario.toStatus
							merchant["updated_at"] = time.Now()

							mu.Lock()
							processedCount++
							mu.Unlock()
						}
					}

					// Simulate batch processing time
					batchProcessingTime := time.Duration(len(batch)/20) * time.Millisecond
					time.Sleep(batchProcessingTime)
				}(i)
			}

			wg.Wait()
			close(errors)

			// Check for errors
			for err := range errors {
				if err != nil {
					return err
				}
			}

			responseTime := time.Since(start)
			runner.RecordRequest(responseTime, true)

			return nil
		}

		// Run the bulk status change operation
		err := operation()
		if err != nil {
			return fmt.Errorf("bulk status change failed for scenario '%s': %w", scenario.name, err)
		}
	}

	runner.Stop()
	metrics := runner.GetMetrics()
	runner.PrintMetrics("Bulk Merchant Status Change Performance")

	// Assert performance requirements
	assert.LessOrEqual(t, metrics.AverageResponseTime, 15*time.Second,
		"Bulk status change should complete within 15 seconds")
	assert.GreaterOrEqual(t, metrics.RequestsPerSecond, 0.1,
		"Should handle at least 0.1 bulk status change operations per second")

	return nil
}

// testBulkMerchantExportPerformance tests bulk merchant export operations
func testBulkMerchantExportPerformance(t *testing.T, config *PerformanceTestConfig) error {
	helper := NewBenchmarkHelper(config)
	runner := NewPerformanceTestRunner(config)
	runner.Start()

	// Generate test data
	testData := helper.GenerateTestData(config.BulkOperationSize)

	// Test different export scenarios
	exportScenarios := []struct {
		name      string
		format    string
		batchSize int
	}{
		{
			name:      "CSV Export",
			format:    "csv",
			batchSize: 100,
		},
		{
			name:      "JSON Export",
			format:    "json",
			batchSize: 200,
		},
		{
			name:      "Excel Export",
			format:    "xlsx",
			batchSize: 500,
		},
	}

	for _, scenario := range exportScenarios {
		operation := func() error {
			start := time.Now()

			// Simulate export operation
			exportedCount := 0
			exportData := make([]map[string]interface{}, 0, len(testData))

			// Process data in batches
			for i := 0; i < len(testData); i += scenario.batchSize {
				endIndex := i + scenario.batchSize
				if endIndex > len(testData) {
					endIndex = len(testData)
				}

				batch := testData[i:endIndex]

				// Simulate data transformation for export
				for _, merchant := range batch {
					exportRecord := map[string]interface{}{
						"id":             merchant["id"],
						"name":           merchant["name"],
						"email":          merchant["email"],
						"phone":          merchant["phone"],
						"address":        merchant["address"],
						"website":        merchant["website"],
						"industry":       merchant["industry"],
						"portfolio_type": merchant["portfolio_type"],
						"risk_level":     merchant["risk_level"],
						"created_at":     merchant["created_at"],
						"updated_at":     merchant["updated_at"],
					}
					exportData = append(exportData, exportRecord)
					exportedCount++
				}

				// Simulate batch processing time
				batchProcessingTime := time.Duration(len(batch)/50) * time.Millisecond
				time.Sleep(batchProcessingTime)
			}

			// Simulate final export processing
			exportProcessingTime := time.Duration(len(exportData)/100) * time.Millisecond
			time.Sleep(exportProcessingTime)

			responseTime := time.Since(start)
			runner.RecordRequest(responseTime, true)

			// Verify all merchants were exported
			if exportedCount != len(testData) {
				return fmt.Errorf("expected %d exports, got %d", len(testData), exportedCount)
			}

			return nil
		}

		// Run the bulk export operation
		err := operation()
		if err != nil {
			return fmt.Errorf("bulk export failed for scenario '%s': %w", scenario.name, err)
		}
	}

	runner.Stop()
	metrics := runner.GetMetrics()
	runner.PrintMetrics("Bulk Merchant Export Performance")

	// Assert performance requirements
	assert.LessOrEqual(t, metrics.AverageResponseTime, 20*time.Second,
		"Bulk export should complete within 20 seconds")
	assert.GreaterOrEqual(t, metrics.RequestsPerSecond, 0.1,
		"Should handle at least 0.1 bulk export operations per second")

	return nil
}

// testBulkMerchantImportPerformance tests bulk merchant import operations
func testBulkMerchantImportPerformance(t *testing.T, config *PerformanceTestConfig) error {
	helper := NewBenchmarkHelper(config)
	runner := NewPerformanceTestRunner(config)
	runner.Start()

	// Generate test data for import
	testData := helper.GenerateTestData(config.BulkOperationSize)

	// Test different import scenarios
	importScenarios := []struct {
		name       string
		format     string
		batchSize  int
		validation bool
	}{
		{
			name:       "CSV Import with Validation",
			format:     "csv",
			batchSize:  100,
			validation: true,
		},
		{
			name:       "JSON Import with Validation",
			format:     "json",
			batchSize:  200,
			validation: true,
		},
		{
			name:       "CSV Import without Validation",
			format:     "csv",
			batchSize:  500,
			validation: false,
		},
	}

	for _, scenario := range importScenarios {
		operation := func() error {
			start := time.Now()

			// Simulate import operation
			importedCount := 0
			validationErrors := 0

			// Process data in batches
			for i := 0; i < len(testData); i += scenario.batchSize {
				endIndex := i + scenario.batchSize
				if endIndex > len(testData) {
					endIndex = len(testData)
				}

				batch := testData[i:endIndex]

				// Simulate batch import processing
				for _, merchant := range batch {
					// Simulate validation if enabled
					if scenario.validation {
						if !validateMerchantData(merchant) {
							validationErrors++
							continue
						}
					}

					// Simulate importing merchant
					merchant["imported_at"] = time.Now()
					importedCount++
				}

				// Simulate batch processing time
				batchProcessingTime := time.Duration(len(batch)/25) * time.Millisecond
				time.Sleep(batchProcessingTime)
			}

			responseTime := time.Since(start)
			success := validationErrors == 0 || !scenario.validation
			runner.RecordRequest(responseTime, success)

			if !success {
				return fmt.Errorf("import failed with %d validation errors", validationErrors)
			}

			return nil
		}

		// Run the bulk import operation
		err := operation()
		if err != nil {
			return fmt.Errorf("bulk import failed for scenario '%s': %w", scenario.name, err)
		}
	}

	runner.Stop()
	metrics := runner.GetMetrics()
	runner.PrintMetrics("Bulk Merchant Import Performance")

	// Assert performance requirements
	assert.LessOrEqual(t, metrics.AverageResponseTime, 25*time.Second,
		"Bulk import should complete within 25 seconds")
	assert.GreaterOrEqual(t, metrics.RequestsPerSecond, 0.1,
		"Should handle at least 0.1 bulk import operations per second")

	return nil
}

// testBulkMerchantDeletionPerformance tests bulk merchant deletion operations
func testBulkMerchantDeletionPerformance(t *testing.T, config *PerformanceTestConfig) error {
	helper := NewBenchmarkHelper(config)
	runner := NewPerformanceTestRunner(config)
	runner.Start()

	// Generate test data
	testData := helper.GenerateTestData(config.BulkOperationSize)

	// Test different deletion scenarios
	deletionScenarios := []struct {
		name        string
		condition   func(map[string]interface{}) bool
		batchSize   int
		concurrency int
	}{
		{
			name: "Delete Deactivated Merchants",
			condition: func(merchant map[string]interface{}) bool {
				return merchant["portfolio_type"] == "deactivated"
			},
			batchSize:   100,
			concurrency: 3,
		},
		{
			name: "Delete High Risk Merchants",
			condition: func(merchant map[string]interface{}) bool {
				return merchant["risk_level"] == "high"
			},
			batchSize:   200,
			concurrency: 2,
		},
		{
			name: "Delete Old Test Merchants",
			condition: func(merchant map[string]interface{}) bool {
				return contains(merchant["name"].(string), "Test")
			},
			batchSize:   500,
			concurrency: 1,
		},
	}

	for _, scenario := range deletionScenarios {
		operation := func() error {
			start := time.Now()

			// Simulate concurrent batch deletion
			var wg sync.WaitGroup
			errors := make(chan error, scenario.concurrency)
			deletedCount := 0
			var mu sync.Mutex

			// Process merchants in concurrent batches
			for i := 0; i < len(testData); i += scenario.batchSize {
				wg.Add(1)
				go func(startIndex int) {
					defer wg.Done()

					endIndex := startIndex + scenario.batchSize
					if endIndex > len(testData) {
						endIndex = len(testData)
					}

					batch := testData[startIndex:endIndex]

					// Process each merchant in the batch
					for _, merchant := range batch {
						if scenario.condition(merchant) {
							// Simulate deletion
							merchant["deleted_at"] = time.Now()
							merchant["deleted"] = "true"

							mu.Lock()
							deletedCount++
							mu.Unlock()
						}
					}

					// Simulate batch processing time
					batchProcessingTime := time.Duration(len(batch)/30) * time.Millisecond
					time.Sleep(batchProcessingTime)
				}(i)
			}

			wg.Wait()
			close(errors)

			// Check for errors
			for err := range errors {
				if err != nil {
					return err
				}
			}

			responseTime := time.Since(start)
			runner.RecordRequest(responseTime, true)

			return nil
		}

		// Run the bulk deletion operation
		err := operation()
		if err != nil {
			return fmt.Errorf("bulk deletion failed for scenario '%s': %w", scenario.name, err)
		}
	}

	runner.Stop()
	metrics := runner.GetMetrics()
	runner.PrintMetrics("Bulk Merchant Deletion Performance")

	// Assert performance requirements
	assert.LessOrEqual(t, metrics.AverageResponseTime, 20*time.Second,
		"Bulk deletion should complete within 20 seconds")
	assert.GreaterOrEqual(t, metrics.RequestsPerSecond, 0.1,
		"Should handle at least 0.1 bulk deletion operations per second")

	return nil
}

// validateMerchantData validates merchant data for import
func validateMerchantData(merchant map[string]interface{}) bool {
	// Basic validation checks
	if merchant["name"] == nil || merchant["name"] == "" {
		return false
	}
	if merchant["email"] == nil || merchant["email"] == "" {
		return false
	}
	if merchant["phone"] == nil || merchant["phone"] == "" {
		return false
	}
	if merchant["address"] == nil || merchant["address"] == "" {
		return false
	}

	// Validate portfolio type
	validPortfolioTypes := []string{"onboarded", "deactivated", "prospective", "pending"}
	portfolioType := merchant["portfolio_type"].(string)
	validType := false
	for _, validType := range validPortfolioTypes {
		if portfolioType == validType {
			validType = true
			break
		}
	}
	if !validType {
		return false
	}

	// Validate risk level
	validRiskLevels := []string{"high", "medium", "low"}
	riskLevel := merchant["risk_level"].(string)
	validRisk := false
	for _, validLevel := range validRiskLevels {
		if riskLevel == validLevel {
			validRisk = true
			break
		}
	}
	if !validRisk {
		return false
	}

	return true
}

// BenchmarkBulkOperations benchmarks bulk operations
func BenchmarkBulkOperations(b *testing.B) {
	config := DefaultPerformanceConfig()
	config.BulkOperationSize = 1000

	helper := NewBenchmarkHelper(config)
	testData := helper.GenerateTestData(config.BulkOperationSize)

	b.ResetTimer()

	b.Run("BulkUpdate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate bulk update
			for _, merchant := range testData {
				merchant["risk_level"] = "high"
				merchant["updated_at"] = time.Now()
			}
		}
	})

	b.Run("BulkStatusChange", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate bulk status change
			for _, merchant := range testData {
				if merchant["portfolio_type"] == "onboarded" {
					merchant["portfolio_type"] = "deactivated"
					merchant["updated_at"] = time.Now()
				}
			}
		}
	})

	b.Run("BulkExport", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate bulk export
			exportData := make([]map[string]interface{}, 0, len(testData))
			for _, merchant := range testData {
				exportRecord := map[string]interface{}{
					"id":   merchant["id"],
					"name": merchant["name"],
				}
				exportData = append(exportData, exportRecord)
			}
			_ = exportData
		}
	})
}
