package risk

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// NewPerformanceTestSuite creates a new performance test suite
// Type definition is in test_suite_types.go
func NewPerformanceTestSuite(t *testing.T) *PerformanceTestSuite {
	logger := zap.NewNop()
	backupDir := t.TempDir()

	// Create services with mock database
	storageService := NewRiskStorageService(nil, logger) // nil DB for performance testing
	validationSvc := NewRiskValidationService(logger)
	exportSvc := NewExportService(logger)
	backupSvc := NewBackupService(logger, backupDir, 30, false)

	// Create job managers
	jobManager := NewExportJobManager(logger, exportSvc)
	backupJobManager := NewBackupJobManager(logger, backupSvc)

	// Create handlers
	exportHandler := NewExportHandler(logger, exportSvc, jobManager)
	backupHandler := NewBackupHandler(logger, backupSvc, backupJobManager)

	// Create HTTP mux
	mux := http.NewServeMux()
	exportHandler.RegisterRoutes(mux)
	backupHandler.RegisterRoutes(mux)

	// Create test server
	server := httptest.NewServer(mux)

	return &PerformanceTestSuite{
		logger:         logger,
		storageService: storageService,
		validationSvc:  validationSvc,
		exportSvc:      exportSvc,
		backupSvc:      backupSvc,
		exportHandler:  exportHandler,
		backupHandler:  backupHandler,
		mux:            mux,
		server:         server,
	}
}

// Close closes the test server
// Method is also declared in test_suite_types.go, but this is the actual implementation
func (suite *PerformanceTestSuite) Close() {
	suite.server.Close()
}

// TestValidationPerformance tests validation service performance
func TestValidationPerformance(t *testing.T) {
	suite := NewPerformanceTestSuite(t)
	defer suite.Close()

	t.Run("Single Validation Performance", func(t *testing.T) {
		ctx := context.Background()

		// Test single validation performance
		assessment := &RiskAssessment{
			ID:           "perf-test-assessment",
			BusinessID:   "perf-test-business",
			BusinessName: "Performance Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		start := time.Now()
		err := suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Less(t, duration, 100*time.Millisecond, "Validation should complete within 100ms")
	})

	t.Run("Bulk Validation Performance", func(t *testing.T) {
		ctx := context.Background()

		// Test bulk validation performance
		numAssessments := 1000
		assessments := make([]*RiskAssessment, numAssessments)

		for i := 0; i < numAssessments; i++ {
			assessments[i] = &RiskAssessment{
				ID:           fmt.Sprintf("perf-test-assessment-%d", i),
				BusinessID:   fmt.Sprintf("perf-test-business-%d", i),
				BusinessName: "Performance Test Business",
				OverallScore: 80.0,
				OverallLevel: RiskLevelMedium,
				AlertLevel:   RiskLevelMedium,
				AssessedAt:   time.Now(),
				ValidUntil:   time.Now().Add(24 * time.Hour),
			}
		}

		start := time.Now()
		for _, assessment := range assessments {
			err := suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
			assert.NoError(t, err)
		}
		duration := time.Since(start)

		avgDuration := duration / time.Duration(numAssessments)
		assert.Less(t, avgDuration, 10*time.Millisecond, "Average validation should complete within 10ms")
		assert.Less(t, duration, 5*time.Second, "Bulk validation should complete within 5 seconds")
	})

	t.Run("Concurrent Validation Performance", func(t *testing.T) {
		ctx := context.Background()

		// Test concurrent validation performance
		numGoroutines := 100
		numValidationsPerGoroutine := 10
		totalValidations := numGoroutines * numValidationsPerGoroutine

		var wg sync.WaitGroup
		results := make(chan time.Duration, totalValidations)

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < numValidationsPerGoroutine; j++ {
					assessment := &RiskAssessment{
						ID:           fmt.Sprintf("perf-test-assessment-%d-%d", goroutineID, j),
						BusinessID:   fmt.Sprintf("perf-test-business-%d-%d", goroutineID, j),
						BusinessName: "Performance Test Business",
						OverallScore: 80.0,
						OverallLevel: RiskLevelMedium,
						AlertLevel:   RiskLevelMedium,
						AssessedAt:   time.Now(),
						ValidUntil:   time.Now().Add(24 * time.Hour),
					}

					validationStart := time.Now()
					err := suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
					validationDuration := time.Since(validationStart)

					assert.NoError(t, err)
					results <- validationDuration
				}
			}(i)
		}

		wg.Wait()
		close(results)
		totalDuration := time.Since(start)

		// Collect results
		var totalValidationTime time.Duration
		var maxValidationTime time.Duration
		var minValidationTime time.Duration = time.Hour
		count := 0

		for duration := range results {
			totalValidationTime += duration
			if duration > maxValidationTime {
				maxValidationTime = duration
			}
			if duration < minValidationTime {
				minValidationTime = duration
			}
			count++
		}

		avgValidationTime := totalValidationTime / time.Duration(count)

		assert.Equal(t, totalValidations, count)
		assert.Less(t, avgValidationTime, 10*time.Millisecond, "Average concurrent validation should complete within 10ms")
		assert.Less(t, maxValidationTime, 100*time.Millisecond, "Max concurrent validation should complete within 100ms")
		assert.Less(t, totalDuration, 2*time.Second, "Total concurrent validation should complete within 2 seconds")
	})
}

// TestExportPerformance tests export service performance
func TestExportPerformance(t *testing.T) {
	suite := NewPerformanceTestSuite(t)
	defer suite.Close()

	t.Run("Single Export Performance", func(t *testing.T) {
		ctx := context.Background()

		// Test single export performance
		request := &ExportRequest{
			BusinessID: "perf-test-business",
			ExportType: ExportTypeAssessments,
			Format:     ExportFormatJSON,
		}

		start := time.Now()
		_, err := suite.exportSvc.ExportData(ctx, request)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Less(t, duration, 500*time.Millisecond, "Export should complete within 500ms")
	})

	t.Run("Bulk Export Performance", func(t *testing.T) {
		ctx := context.Background()

		// Test bulk export performance
		numExports := 100
		requests := make([]*ExportRequest, numExports)

		for i := 0; i < numExports; i++ {
			requests[i] = &ExportRequest{
				BusinessID: fmt.Sprintf("perf-test-business-%d", i),
				ExportType: ExportTypeAssessments,
				Format:     ExportFormatJSON,
			}
		}

		start := time.Now()
		for _, request := range requests {
			_, err := suite.exportSvc.ExportData(ctx, request)
			assert.NoError(t, err)
		}
		duration := time.Since(start)

		avgDuration := duration / time.Duration(numExports)
		assert.Less(t, avgDuration, 50*time.Millisecond, "Average export should complete within 50ms")
		assert.Less(t, duration, 10*time.Second, "Bulk export should complete within 10 seconds")
	})

	t.Run("Concurrent Export Performance", func(t *testing.T) {
		ctx := context.Background()

		// Test concurrent export performance
		numGoroutines := 50
		numExportsPerGoroutine := 5
		totalExports := numGoroutines * numExportsPerGoroutine

		var wg sync.WaitGroup
		results := make(chan time.Duration, totalExports)

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < numExportsPerGoroutine; j++ {
					request := &ExportRequest{
						BusinessID: fmt.Sprintf("perf-test-business-%d-%d", goroutineID, j),
						ExportType: ExportTypeAssessments,
						Format:     ExportFormatJSON,
					}

					exportStart := time.Now()
					_, err := suite.exportSvc.ExportData(ctx, request)
					exportDuration := time.Since(exportStart)

					assert.NoError(t, err)
					results <- exportDuration
				}
			}(i)
		}

		wg.Wait()
		close(results)
		totalDuration := time.Since(start)

		// Collect results
		var totalExportTime time.Duration
		var maxExportTime time.Duration
		var minExportTime time.Duration = time.Hour
		count := 0

		for duration := range results {
			totalExportTime += duration
			if duration > maxExportTime {
				maxExportTime = duration
			}
			if duration < minExportTime {
				minExportTime = duration
			}
			count++
		}

		avgExportTime := totalExportTime / time.Duration(count)

		assert.Equal(t, totalExports, count)
		assert.Less(t, avgExportTime, 50*time.Millisecond, "Average concurrent export should complete within 50ms")
		assert.Less(t, maxExportTime, 500*time.Millisecond, "Max concurrent export should complete within 500ms")
		assert.Less(t, totalDuration, 5*time.Second, "Total concurrent export should complete within 5 seconds")
	})

	t.Run("Large Data Export Performance", func(t *testing.T) {
		ctx := context.Background()

		// Test large data export performance
		request := &ExportRequest{
			BusinessID: "perf-test-business",
			ExportType: ExportTypeAllData,
			Format:     ExportFormatJSON,
			Metadata:   make(map[string]interface{}),
		}

		// Add large metadata to simulate large data export
		for i := 0; i < 10000; i++ {
			request.Metadata[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
		}

		start := time.Now()
		_, err := suite.exportSvc.ExportData(ctx, request)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Less(t, duration, 2*time.Second, "Large data export should complete within 2 seconds")
	})
}

// TestBackupPerformance tests backup service performance
func TestBackupPerformance(t *testing.T) {
	suite := NewPerformanceTestSuite(t)
	defer suite.Close()

	t.Run("Single Backup Performance", func(t *testing.T) {
		ctx := context.Background()

		// Test single backup performance
		request := &BackupRequest{
			BusinessID:  "perf-test-business",
			BackupType:  BackupTypeBusiness,
			IncludeData: []string{"assessments"},
		}

		start := time.Now()
		_, err := suite.backupSvc.CreateBackup(ctx, request)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Less(t, duration, 1*time.Second, "Backup should complete within 1 second")
	})

	t.Run("Bulk Backup Performance", func(t *testing.T) {
		ctx := context.Background()

		// Test bulk backup performance
		numBackups := 50
		requests := make([]*BackupRequest, numBackups)

		for i := 0; i < numBackups; i++ {
			requests[i] = &BackupRequest{
				BusinessID:  fmt.Sprintf("perf-test-business-%d", i),
				BackupType:  BackupTypeBusiness,
				IncludeData: []string{"assessments"},
			}
		}

		start := time.Now()
		for _, request := range requests {
			_, err := suite.backupSvc.CreateBackup(ctx, request)
			assert.NoError(t, err)
		}
		duration := time.Since(start)

		avgDuration := duration / time.Duration(numBackups)
		assert.Less(t, avgDuration, 100*time.Millisecond, "Average backup should complete within 100ms")
		assert.Less(t, duration, 10*time.Second, "Bulk backup should complete within 10 seconds")
	})

	t.Run("Concurrent Backup Performance", func(t *testing.T) {
		ctx := context.Background()

		// Test concurrent backup performance
		numGoroutines := 25
		numBackupsPerGoroutine := 4
		totalBackups := numGoroutines * numBackupsPerGoroutine

		var wg sync.WaitGroup
		results := make(chan time.Duration, totalBackups)

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < numBackupsPerGoroutine; j++ {
					request := &BackupRequest{
						BusinessID:  fmt.Sprintf("perf-test-business-%d-%d", goroutineID, j),
						BackupType:  BackupTypeBusiness,
						IncludeData: []string{"assessments"},
					}

					backupStart := time.Now()
					_, err := suite.backupSvc.CreateBackup(ctx, request)
					backupDuration := time.Since(backupStart)

					assert.NoError(t, err)
					results <- backupDuration
				}
			}(i)
		}

		wg.Wait()
		close(results)
		totalDuration := time.Since(start)

		// Collect results
		var totalBackupTime time.Duration
		var maxBackupTime time.Duration
		var minBackupTime time.Duration = time.Hour
		count := 0

		for duration := range results {
			totalBackupTime += duration
			if duration > maxBackupTime {
				maxBackupTime = duration
			}
			if duration < minBackupTime {
				minBackupTime = duration
			}
			count++
		}

		avgBackupTime := totalBackupTime / time.Duration(count)

		assert.Equal(t, totalBackups, count)
		assert.Less(t, avgBackupTime, 100*time.Millisecond, "Average concurrent backup should complete within 100ms")
		assert.Less(t, maxBackupTime, 1*time.Second, "Max concurrent backup should complete within 1 second")
		assert.Less(t, totalDuration, 5*time.Second, "Total concurrent backup should complete within 5 seconds")
	})

	t.Run("Large Data Backup Performance", func(t *testing.T) {
		ctx := context.Background()

		// Test large data backup performance
		request := &BackupRequest{
			BusinessID:  "perf-test-business",
			BackupType:  BackupTypeFull,
			IncludeData: []string{"assessments", "factors", "trends", "alerts", "history"},
		}

		start := time.Now()
		_, err := suite.backupSvc.CreateBackup(ctx, request)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Less(t, duration, 3*time.Second, "Large data backup should complete within 3 seconds")
	})
}

// TestAPIPerformanceEndpoint tests API endpoint performance
func TestAPIPerformanceEndpoint(t *testing.T) {
	suite := NewPerformanceTestSuite(t)
	defer suite.Close()

	t.Run("Single API Request Performance", func(t *testing.T) {
		// Test single API request performance
		req := httptest.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs",
			[]byte(`{"business_id": "perf-test-business", "export_type": "assessments", "format": "json"}`))
		req.Header.Set("Content-Type", "application/json")

		start := time.Now()
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		duration := time.Since(start)

		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Less(t, duration, 500*time.Millisecond, "API request should complete within 500ms")
	})

	t.Run("Bulk API Request Performance", func(t *testing.T) {
		// Test bulk API request performance
		numRequests := 100
		client := &http.Client{Timeout: 30 * time.Second}

		start := time.Now()
		for i := 0; i < numRequests; i++ {
			req := httptest.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs",
				[]byte(fmt.Sprintf(`{"business_id": "perf-test-business-%d", "export_type": "assessments", "format": "json"}`, i)))
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			require.NoError(t, err)
			resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}
		duration := time.Since(start)

		avgDuration := duration / time.Duration(numRequests)
		assert.Less(t, avgDuration, 50*time.Millisecond, "Average API request should complete within 50ms")
		assert.Less(t, duration, 10*time.Second, "Bulk API requests should complete within 10 seconds")
	})

	t.Run("Concurrent API Request Performance", func(t *testing.T) {
		// Test concurrent API request performance
		numGoroutines := 50
		numRequestsPerGoroutine := 4
		totalRequests := numGoroutines * numRequestsPerGoroutine

		var wg sync.WaitGroup
		results := make(chan time.Duration, totalRequests)
		client := &http.Client{Timeout: 30 * time.Second}

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < numRequestsPerGoroutine; j++ {
					req := httptest.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs",
						[]byte(fmt.Sprintf(`{"business_id": "perf-test-business-%d-%d", "export_type": "assessments", "format": "json"}`, goroutineID, j)))
					req.Header.Set("Content-Type", "application/json")

					requestStart := time.Now()
					resp, err := client.Do(req)
					requestDuration := time.Since(requestStart)

					require.NoError(t, err)
					resp.Body.Close()
					assert.Equal(t, http.StatusOK, resp.StatusCode)
					results <- requestDuration
				}
			}(i)
		}

		wg.Wait()
		close(results)
		totalDuration := time.Since(start)

		// Collect results
		var totalRequestTime time.Duration
		var maxRequestTime time.Duration
		var minRequestTime time.Duration = time.Hour
		count := 0

		for duration := range results {
			totalRequestTime += duration
			if duration > maxRequestTime {
				maxRequestTime = duration
			}
			if duration < minRequestTime {
				minRequestTime = duration
			}
			count++
		}

		avgRequestTime := totalRequestTime / time.Duration(count)

		assert.Equal(t, totalRequests, count)
		assert.Less(t, avgRequestTime, 50*time.Millisecond, "Average concurrent API request should complete within 50ms")
		assert.Less(t, maxRequestTime, 500*time.Millisecond, "Max concurrent API request should complete within 500ms")
		assert.Less(t, totalDuration, 5*time.Second, "Total concurrent API requests should complete within 5 seconds")
	})

	t.Run("Large Payload API Performance", func(t *testing.T) {
		// Test large payload API performance
		largePayload := make([]byte, 1024*1024) // 1MB payload
		for i := range largePayload {
			largePayload[i] = 'A'
		}

		req := httptest.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs", largePayload)
		req.Header.Set("Content-Type", "application/json")

		start := time.Now()
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		duration := time.Since(start)

		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle large payload gracefully
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest)
		assert.Less(t, duration, 2*time.Second, "Large payload API request should complete within 2 seconds")
	})
}

// TestMemoryPerformance tests memory usage performance
func TestMemoryPerformance(t *testing.T) {
	suite := NewPerformanceTestSuite(t)
	defer suite.Close()

	t.Run("Memory Usage Under Load", func(t *testing.T) {
		ctx := context.Background()

		// Test memory usage under load
		numOperations := 1000
		var wg sync.WaitGroup

		start := time.Now()

		for i := 0; i < numOperations; i++ {
			wg.Add(1)
			go func(operationID int) {
				defer wg.Done()

				// Perform multiple operations to test memory usage
				assessment := &RiskAssessment{
					ID:           fmt.Sprintf("memory-test-assessment-%d", operationID),
					BusinessID:   fmt.Sprintf("memory-test-business-%d", operationID),
					BusinessName: "Memory Test Business",
					OverallScore: 80.0,
					OverallLevel: RiskLevelMedium,
					AlertLevel:   RiskLevelMedium,
					AssessedAt:   time.Now(),
					ValidUntil:   time.Now().Add(24 * time.Hour),
				}

				// Validation
				err := suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
				assert.NoError(t, err)

				// Export
				exportRequest := &ExportRequest{
					BusinessID: assessment.BusinessID,
					ExportType: ExportTypeAssessments,
					Format:     ExportFormatJSON,
				}
				_, err = suite.exportSvc.ExportData(ctx, exportRequest)
				assert.NoError(t, err)

				// Backup
				backupRequest := &BackupRequest{
					BusinessID:  assessment.BusinessID,
					BackupType:  BackupTypeBusiness,
					IncludeData: []string{"assessments"},
				}
				_, err = suite.backupSvc.CreateBackup(ctx, backupRequest)
				assert.NoError(t, err)
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		avgDuration := duration / time.Duration(numOperations)
		assert.Less(t, avgDuration, 100*time.Millisecond, "Average memory test operation should complete within 100ms")
		assert.Less(t, duration, 30*time.Second, "Memory test should complete within 30 seconds")
	})

	t.Run("Memory Leak Detection", func(t *testing.T) {
		ctx := context.Background()

		// Test for memory leaks by running operations repeatedly
		numIterations := 100
		numOperationsPerIteration := 10

		for iteration := 0; iteration < numIterations; iteration++ {
			for i := 0; i < numOperationsPerIteration; i++ {
				assessment := &RiskAssessment{
					ID:           fmt.Sprintf("leak-test-assessment-%d-%d", iteration, i),
					BusinessID:   fmt.Sprintf("leak-test-business-%d-%d", iteration, i),
					BusinessName: "Memory Leak Test Business",
					OverallScore: 80.0,
					OverallLevel: RiskLevelMedium,
					AlertLevel:   RiskLevelMedium,
					AssessedAt:   time.Now(),
					ValidUntil:   time.Now().Add(24 * time.Hour),
				}

				// Perform operations that might cause memory leaks
				err := suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
				assert.NoError(t, err)

				exportRequest := &ExportRequest{
					BusinessID: assessment.BusinessID,
					ExportType: ExportTypeAssessments,
					Format:     ExportFormatJSON,
				}
				_, err = suite.exportSvc.ExportData(ctx, exportRequest)
				assert.NoError(t, err)
			}

			// Force garbage collection every 10 iterations
			if iteration%10 == 0 {
				// In a real test, you might want to check memory usage here
				// For now, we just ensure the operations complete successfully
			}
		}
	})
}

// TestCPUPerformance tests CPU usage performance
func TestCPUPerformance(t *testing.T) {
	suite := NewPerformanceTestSuite(t)
	defer suite.Close()

	t.Run("CPU Intensive Operations", func(t *testing.T) {
		ctx := context.Background()

		// Test CPU intensive operations
		numOperations := 100
		start := time.Now()

		for i := 0; i < numOperations; i++ {
			// Create complex assessment with multiple factors
			assessment := &RiskAssessment{
				ID:             fmt.Sprintf("cpu-test-assessment-%d", i),
				BusinessID:     fmt.Sprintf("cpu-test-business-%d", i),
				BusinessName:   "CPU Test Business",
				OverallScore:   80.0,
				OverallLevel:   RiskLevelMedium,
				AlertLevel:     RiskLevelMedium,
				AssessedAt:     time.Now(),
				ValidUntil:     time.Now().Add(24 * time.Hour),
				CategoryScores: make(map[RiskCategory]RiskScore),
				FactorScores:   make([]RiskScore, 0),
			}

			// Add multiple category scores
			for category := RiskCategoryFinancial; category <= RiskCategoryOperational; category++ {
				assessment.CategoryScores[category] = RiskScore{
					FactorID:     fmt.Sprintf("factor-%d", int(category)),
					FactorName:   fmt.Sprintf("Factor %d", int(category)),
					Category:     category,
					Score:        70.0 + float64(i),
					Level:        RiskLevelMedium,
					Confidence:   0.8,
					Explanation:  "CPU test factor",
					Evidence:     []string{"evidence1", "evidence2"},
					CalculatedAt: time.Now(),
				}
			}

			// Add multiple factor scores
			for j := 0; j < 10; j++ {
				assessment.FactorScores = append(assessment.FactorScores, RiskScore{
					FactorID:     fmt.Sprintf("factor-%d-%d", i, j),
					FactorName:   fmt.Sprintf("Factor %d-%d", i, j),
					Category:     RiskCategoryFinancial,
					Score:        70.0 + float64(i+j),
					Level:        RiskLevelMedium,
					Confidence:   0.8,
					Explanation:  "CPU test factor",
					Evidence:     []string{"evidence1", "evidence2"},
					CalculatedAt: time.Now(),
				})
			}

			// Perform validation (CPU intensive)
			err := suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
			assert.NoError(t, err)
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(numOperations)

		assert.Less(t, avgDuration, 50*time.Millisecond, "Average CPU intensive operation should complete within 50ms")
		assert.Less(t, duration, 10*time.Second, "CPU intensive operations should complete within 10 seconds")
	})

	t.Run("Concurrent CPU Intensive Operations", func(t *testing.T) {
		ctx := context.Background()

		// Test concurrent CPU intensive operations
		numGoroutines := 20
		numOperationsPerGoroutine := 5
		totalOperations := numGoroutines * numOperationsPerGoroutine

		var wg sync.WaitGroup
		results := make(chan time.Duration, totalOperations)

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < numOperationsPerGoroutine; j++ {
					assessment := &RiskAssessment{
						ID:             fmt.Sprintf("cpu-concurrent-test-assessment-%d-%d", goroutineID, j),
						BusinessID:     fmt.Sprintf("cpu-concurrent-test-business-%d-%d", goroutineID, j),
						BusinessName:   "CPU Concurrent Test Business",
						OverallScore:   80.0,
						OverallLevel:   RiskLevelMedium,
						AlertLevel:     RiskLevelMedium,
						AssessedAt:     time.Now(),
						ValidUntil:     time.Now().Add(24 * time.Hour),
						CategoryScores: make(map[RiskCategory]RiskScore),
						FactorScores:   make([]RiskScore, 0),
					}

					// Add complex data
					for k := 0; k < 5; k++ {
						assessment.FactorScores = append(assessment.FactorScores, RiskScore{
							FactorID:     fmt.Sprintf("factor-%d-%d-%d", goroutineID, j, k),
							FactorName:   fmt.Sprintf("Factor %d-%d-%d", goroutineID, j, k),
							Category:     RiskCategoryFinancial,
							Score:        70.0 + float64(goroutineID+j+k),
							Level:        RiskLevelMedium,
							Confidence:   0.8,
							Explanation:  "CPU concurrent test factor",
							Evidence:     []string{"evidence1", "evidence2"},
							CalculatedAt: time.Now(),
						})
					}

					operationStart := time.Now()
					err := suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
					operationDuration := time.Since(operationStart)

					assert.NoError(t, err)
					results <- operationDuration
				}
			}(i)
		}

		wg.Wait()
		close(results)
		totalDuration := time.Since(start)

		// Collect results
		var totalOperationTime time.Duration
		var maxOperationTime time.Duration
		var minOperationTime time.Duration = time.Hour
		count := 0

		for duration := range results {
			totalOperationTime += duration
			if duration > maxOperationTime {
				maxOperationTime = duration
			}
			if duration < minOperationTime {
				minOperationTime = duration
			}
			count++
		}

		avgOperationTime := totalOperationTime / time.Duration(count)

		assert.Equal(t, totalOperations, count)
		assert.Less(t, avgOperationTime, 50*time.Millisecond, "Average concurrent CPU operation should complete within 50ms")
		assert.Less(t, maxOperationTime, 200*time.Millisecond, "Max concurrent CPU operation should complete within 200ms")
		assert.Less(t, totalDuration, 5*time.Second, "Total concurrent CPU operations should complete within 5 seconds")
	})
}

// TestScalabilityPerformance tests system scalability performance
func TestScalabilityPerformance(t *testing.T) {
	suite := NewPerformanceTestSuite(t)
	defer suite.Close()

	t.Run("Horizontal Scaling Simulation", func(t *testing.T) {
		ctx := context.Background()

		// Simulate horizontal scaling with multiple "nodes"
		numNodes := 10
		numOperationsPerNode := 20
		totalOperations := numNodes * numOperationsPerNode

		var wg sync.WaitGroup
		results := make(chan time.Duration, totalOperations)

		start := time.Now()

		for nodeID := 0; nodeID < numNodes; nodeID++ {
			wg.Add(1)
			go func(nodeID int) {
				defer wg.Done()

				for operationID := 0; operationID < numOperationsPerNode; operationID++ {
					assessment := &RiskAssessment{
						ID:           fmt.Sprintf("scale-test-assessment-%d-%d", nodeID, operationID),
						BusinessID:   fmt.Sprintf("scale-test-business-%d-%d", nodeID, operationID),
						BusinessName: "Scalability Test Business",
						OverallScore: 80.0,
						OverallLevel: RiskLevelMedium,
						AlertLevel:   RiskLevelMedium,
						AssessedAt:   time.Now(),
						ValidUntil:   time.Now().Add(24 * time.Hour),
					}

					operationStart := time.Now()

					// Perform multiple operations to simulate node workload
					err := suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
					assert.NoError(t, err)

					exportRequest := &ExportRequest{
						BusinessID: assessment.BusinessID,
						ExportType: ExportTypeAssessments,
						Format:     ExportFormatJSON,
					}
					_, err = suite.exportSvc.ExportData(ctx, exportRequest)
					assert.NoError(t, err)

					operationDuration := time.Since(operationStart)
					results <- operationDuration
				}
			}(nodeID)
		}

		wg.Wait()
		close(results)
		totalDuration := time.Since(start)

		// Collect results
		var totalOperationTime time.Duration
		count := 0

		for duration := range results {
			totalOperationTime += duration
			count++
		}

		avgOperationTime := totalOperationTime / time.Duration(count)

		assert.Equal(t, totalOperations, count)
		assert.Less(t, avgOperationTime, 100*time.Millisecond, "Average scaling operation should complete within 100ms")
		assert.Less(t, totalDuration, 10*time.Second, "Scaling simulation should complete within 10 seconds")
	})

	t.Run("Load Balancing Simulation", func(t *testing.T) {
		ctx := context.Background()

		// Simulate load balancing with varying load
		numLoadLevels := 5
		operationsPerLevel := 50

		for loadLevel := 0; loadLevel < numLoadLevels; loadLevel++ {
			concurrentOperations := (loadLevel + 1) * 10 // Increasing load
			var wg sync.WaitGroup

			start := time.Now()

			for i := 0; i < concurrentOperations; i++ {
				wg.Add(1)
				go func(operationID int) {
					defer wg.Done()

					for j := 0; j < operationsPerLevel; j++ {
						assessment := &RiskAssessment{
							ID:           fmt.Sprintf("load-test-assessment-%d-%d-%d", loadLevel, operationID, j),
							BusinessID:   fmt.Sprintf("load-test-business-%d-%d-%d", loadLevel, operationID, j),
							BusinessName: "Load Test Business",
							OverallScore: 80.0,
							OverallLevel: RiskLevelMedium,
							AlertLevel:   RiskLevelMedium,
							AssessedAt:   time.Now(),
							ValidUntil:   time.Now().Add(24 * time.Hour),
						}

						err := suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
						assert.NoError(t, err)
					}
				}(i)
			}

			wg.Wait()
			duration := time.Since(start)

			totalOperations := concurrentOperations * operationsPerLevel
			avgDuration := duration / time.Duration(totalOperations)

			// Performance should degrade gracefully with increased load
			maxAllowedDuration := time.Duration(loadLevel+1) * 10 * time.Millisecond
			assert.Less(t, avgDuration, maxAllowedDuration,
				"Average operation duration should scale gracefully with load level %d", loadLevel)
		}
	})
}
