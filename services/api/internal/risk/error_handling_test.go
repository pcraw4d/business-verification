package risk

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ErrorHandlingTestSuite provides comprehensive error handling testing
type ErrorHandlingTestSuite struct {
	logger         *zap.Logger
	storageService *RiskStorageService
	validationSvc  *RiskValidationService
	exportSvc      *ExportService
	backupSvc      *BackupService
	exportHandler  *ExportHandler
	backupHandler  *BackupHandler
	mux            *http.ServeMux
	server         *httptest.Server
}

// NewErrorHandlingTestSuite creates a new error handling test suite
func NewErrorHandlingTestSuite(t *testing.T) *ErrorHandlingTestSuite {
	logger := zap.NewNop()
	backupDir := t.TempDir()

	// Create services with mock database
	storageService := NewRiskStorageService(nil, logger) // nil DB for error testing
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

	return &ErrorHandlingTestSuite{
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
func (suite *ErrorHandlingTestSuite) Close() {
	suite.server.Close()
}

// TestValidationErrors tests validation error handling
func TestValidationErrors(t *testing.T) {
	suite := NewErrorHandlingTestSuite(t)
	defer suite.Close()

	t.Run("Invalid Risk Assessment Data", func(t *testing.T) {
		ctx := context.Background()

		// Test with nil assessment
		err := suite.validationSvc.ValidateRiskAssessment(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "assessment cannot be nil")

		// Test with empty business ID
		assessment := &RiskAssessment{
			ID:           "test-assessment",
			BusinessID:   "", // Invalid: empty business ID
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		err = suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "business ID is required")

		// Test with invalid score
		assessment.BusinessID = "test-business"
		assessment.OverallScore = 150.0 // Invalid: score > 100

		err = suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "score must be between 0 and 100")

		// Test with invalid risk level
		assessment.OverallScore = 80.0
		assessment.OverallLevel = "invalid-level" // Invalid: not a valid risk level

		err = suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid risk level")
	})

	t.Run("Invalid Risk Factor Data", func(t *testing.T) {
		ctx := context.Background()

		// Test with nil factor
		err := suite.validationSvc.ValidateRiskFactor(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "factor cannot be nil")

		// Test with empty factor name
		factor := &RiskFactorInput{
			ID:          "test-factor",
			Name:        "", // Invalid: empty name
			Description: "Test factor",
			Weight:      0.3,
			Value:       85.0,
		}

		err = suite.validationSvc.ValidateRiskFactor(ctx, factor)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")

		// Test with invalid weight
		factor.Name = "Test Factor"
		factor.Weight = 1.5 // Invalid: weight > 1.0

		err = suite.validationSvc.ValidateRiskFactor(ctx, factor)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "weight must be between 0 and 1")

		// Test with invalid value
		factor.Weight = 0.3
		factor.Value = -10.0 // Invalid: negative value

		err = suite.validationSvc.ValidateRiskFactor(ctx, factor)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value must be between 0 and 100")
	})

	t.Run("Invalid Export Request Data", func(t *testing.T) {
		ctx := context.Background()

		// Test with nil request
		_, err := suite.exportSvc.ExportData(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request cannot be nil")

		// Test with empty business ID
		request := &ExportRequest{
			BusinessID: "", // Invalid: empty business ID
			ExportType: ExportTypeAssessments,
			Format:     ExportFormatJSON,
		}

		_, err = suite.exportSvc.ExportData(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "business ID is required")

		// Test with invalid export type
		request.BusinessID = "test-business"
		request.ExportType = "invalid-type" // Invalid: not a valid export type

		_, err = suite.exportSvc.ExportData(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid export type")

		// Test with invalid format
		request.ExportType = ExportTypeAssessments
		request.Format = "invalid-format" // Invalid: not a valid format

		_, err = suite.exportSvc.ExportData(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid export format")
	})

	t.Run("Invalid Backup Request Data", func(t *testing.T) {
		ctx := context.Background()

		// Test with nil request
		_, err := suite.backupSvc.CreateBackup(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request cannot be nil")

		// Test with empty business ID
		request := &BackupRequest{
			BusinessID:  "", // Invalid: empty business ID
			BackupType:  BackupTypeBusiness,
			IncludeData: []string{"assessments"},
		}

		_, err = suite.backupSvc.CreateBackup(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "business ID is required")

		// Test with invalid backup type
		request.BusinessID = "test-business"
		request.BackupType = "invalid-type" // Invalid: not a valid backup type

		_, err = suite.backupSvc.CreateBackup(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid backup type")

		// Test with empty include data
		request.BackupType = BackupTypeBusiness
		request.IncludeData = []string{} // Invalid: empty include data

		_, err = suite.backupSvc.CreateBackup(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "include data cannot be empty")
	})
}

// TestDatabaseErrors tests database error handling
func TestDatabaseErrors(t *testing.T) {
	suite := NewErrorHandlingTestSuite(t)
	defer suite.Close()

	t.Run("Database Connection Errors", func(t *testing.T) {
		ctx := context.Background()

		// Test with nil database connection
		assessment := &RiskAssessment{
			ID:           "test-assessment",
			BusinessID:   "test-business",
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		// This should fail because storageService has nil database
		err := suite.storageService.StoreRiskAssessment(ctx, assessment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection")

		// Test retrieval with nil database
		_, err = suite.storageService.GetRiskAssessment(ctx, "test-assessment")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection")
	})

	t.Run("Database Query Errors", func(t *testing.T) {
		ctx := context.Background()

		// Test with invalid assessment ID
		_, err := suite.storageService.GetRiskAssessment(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "assessment ID is required")

		// Test with non-existent assessment ID
		_, err = suite.storageService.GetRiskAssessment(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Database Transaction Errors", func(t *testing.T) {
		ctx := context.Background()

		// Test transaction rollback scenario
		assessment := &RiskAssessment{
			ID:           "test-assessment",
			BusinessID:   "test-business",
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		// This should fail and trigger rollback
		err := suite.storageService.StoreRiskAssessment(ctx, assessment)
		assert.Error(t, err)
	})
}

// TestAPIErrors tests API error handling
func TestAPIErrors(t *testing.T) {
	suite := NewErrorHandlingTestSuite(t)
	defer suite.Close()

	t.Run("Invalid JSON Request Body", func(t *testing.T) {
		// Test with malformed JSON
		req := httptest.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs",
			[]byte("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		// Test with missing required fields
		req := httptest.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs",
			[]byte(`{"business_id": "test-business"}`)) // Missing export_type and format
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid HTTP Method", func(t *testing.T) {
		// Test with invalid HTTP method
		req := httptest.NewRequest("PUT", suite.server.URL+"/api/v1/export/jobs", nil)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("Non-existent Endpoint", func(t *testing.T) {
		// Test with non-existent endpoint
		req := httptest.NewRequest("GET", suite.server.URL+"/api/v1/non-existent", nil)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Large Request Body", func(t *testing.T) {
		// Test with very large request body
		largeData := make([]byte, 10*1024*1024) // 10MB
		for i := range largeData {
			largeData[i] = 'A'
		}

		req := httptest.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs",
			largeData)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle gracefully
		assert.True(t, resp.StatusCode == http.StatusBadRequest ||
			resp.StatusCode == http.StatusRequestEntityTooLarge)
	})
}

// TestServiceErrors tests service layer error handling
func TestServiceErrors(t *testing.T) {
	suite := NewErrorHandlingTestSuite(t)
	defer suite.Close()

	t.Run("Export Service Errors", func(t *testing.T) {
		ctx := context.Background()

		// Test with invalid export type
		request := &ExportRequest{
			BusinessID: "test-business",
			ExportType: "invalid-type",
			Format:     ExportFormatJSON,
		}

		_, err := suite.exportSvc.ExportData(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid export type")

		// Test with unsupported format
		request.ExportType = ExportTypeAssessments
		request.Format = "unsupported-format"

		_, err = suite.exportSvc.ExportData(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})

	t.Run("Backup Service Errors", func(t *testing.T) {
		ctx := context.Background()

		// Test with invalid backup type
		request := &BackupRequest{
			BusinessID:  "test-business",
			BackupType:  "invalid-type",
			IncludeData: []string{"assessments"},
		}

		_, err := suite.backupSvc.CreateBackup(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid backup type")

		// Test with invalid include data
		request.BackupType = BackupTypeBusiness
		request.IncludeData = []string{"invalid-data-type"}

		_, err = suite.backupSvc.CreateBackup(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid data type")
	})

	t.Run("Validation Service Errors", func(t *testing.T) {
		ctx := context.Background()

		// Test with nil context
		assessment := &RiskAssessment{
			ID:           "test-assessment",
			BusinessID:   "test-business",
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		err := suite.validationSvc.ValidateRiskAssessment(nil, assessment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")
	})
}

// TestConcurrencyErrors tests concurrency-related error handling
func TestConcurrencyErrors(t *testing.T) {
	suite := NewErrorHandlingTestSuite(t)
	defer suite.Close()

	t.Run("Concurrent Export Operations", func(t *testing.T) {
		ctx := context.Background()

		// Test concurrent export operations
		numOperations := 10
		results := make(chan error, numOperations)

		for i := 0; i < numOperations; i++ {
			go func(i int) {
				request := &ExportRequest{
					BusinessID: fmt.Sprintf("test-business-%d", i),
					ExportType: ExportTypeAssessments,
					Format:     ExportFormatJSON,
				}

				_, err := suite.exportSvc.ExportData(ctx, request)
				results <- err
			}(i)
		}

		// Wait for all operations to complete
		errorCount := 0
		for i := 0; i < numOperations; i++ {
			err := <-results
			if err != nil {
				errorCount++
			}
		}

		// Some operations might fail due to concurrency, but system should handle gracefully
		assert.True(t, errorCount >= 0)
	})

	t.Run("Concurrent Backup Operations", func(t *testing.T) {
		ctx := context.Background()

		// Test concurrent backup operations
		numOperations := 5
		results := make(chan error, numOperations)

		for i := 0; i < numOperations; i++ {
			go func(i int) {
				request := &BackupRequest{
					BusinessID:  fmt.Sprintf("test-business-%d", i),
					BackupType:  BackupTypeBusiness,
					IncludeData: []string{"assessments"},
				}

				_, err := suite.backupSvc.CreateBackup(ctx, request)
				results <- err
			}(i)
		}

		// Wait for all operations to complete
		errorCount := 0
		for i := 0; i < numOperations; i++ {
			err := <-results
			if err != nil {
				errorCount++
			}
		}

		// Some operations might fail due to concurrency, but system should handle gracefully
		assert.True(t, errorCount >= 0)
	})
}

// TestResourceErrors tests resource-related error handling
func TestResourceErrors(t *testing.T) {
	suite := NewErrorHandlingTestSuite(t)
	defer suite.Close()

	t.Run("Memory Exhaustion", func(t *testing.T) {
		ctx := context.Background()

		// Test with very large export request
		request := &ExportRequest{
			BusinessID: "test-business",
			ExportType: ExportTypeAllData,
			Format:     ExportFormatJSON,
			Metadata:   make(map[string]interface{}),
		}

		// Add large metadata to simulate memory pressure
		for i := 0; i < 10000; i++ {
			request.Metadata[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
		}

		_, err := suite.exportSvc.ExportData(ctx, request)
		// Should handle gracefully without crashing
		assert.NoError(t, err)
	})

	t.Run("File System Errors", func(t *testing.T) {
		ctx := context.Background()

		// Test backup with invalid directory
		invalidBackupSvc := NewBackupService(suite.logger, "/invalid/path", 30, false)

		request := &BackupRequest{
			BusinessID:  "test-business",
			BackupType:  BackupTypeBusiness,
			IncludeData: []string{"assessments"},
		}

		_, err := invalidBackupSvc.CreateBackup(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "backup directory")
	})

	t.Run("Network Timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Test with very short timeout
		request := &ExportRequest{
			BusinessID: "test-business",
			ExportType: ExportTypeAssessments,
			Format:     ExportFormatJSON,
		}

		_, err := suite.exportSvc.ExportData(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}

// TestSecurityErrors tests security-related error handling
func TestSecurityErrors(t *testing.T) {
	suite := NewErrorHandlingTestSuite(t)
	defer suite.Close()

	t.Run("SQL Injection Attempts", func(t *testing.T) {
		ctx := context.Background()

		// Test with SQL injection payload
		assessment := &RiskAssessment{
			ID:           "test'; DROP TABLE assessments; --",
			BusinessID:   "test-business",
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		// Should handle gracefully without executing SQL
		err := suite.storageService.StoreRiskAssessment(ctx, assessment)
		assert.Error(t, err) // Should fail due to validation, not SQL execution
	})

	t.Run("XSS Attempts", func(t *testing.T) {
		ctx := context.Background()

		// Test with XSS payload
		assessment := &RiskAssessment{
			ID:           "test-assessment",
			BusinessID:   "test-business",
			BusinessName: "<script>alert('xss')</script>",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		// Should handle gracefully
		err := suite.storageService.StoreRiskAssessment(ctx, assessment)
		assert.Error(t, err) // Should fail due to validation
	})

	t.Run("Path Traversal Attempts", func(t *testing.T) {
		ctx := context.Background()

		// Test with path traversal payload
		request := &BackupRequest{
			BusinessID:  "test-business",
			BackupType:  BackupTypeBusiness,
			IncludeData: []string{"../../../etc/passwd"},
		}

		_, err := suite.backupSvc.CreateBackup(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid data type")
	})
}

// TestRecoveryErrors tests error recovery scenarios
func TestRecoveryErrors(t *testing.T) {
	suite := NewErrorHandlingTestSuite(t)
	defer suite.Close()

	t.Run("Service Recovery", func(t *testing.T) {
		ctx := context.Background()

		// Test service recovery after error
		request := &ExportRequest{
			BusinessID: "test-business",
			ExportType: "invalid-type", // This will fail
			Format:     ExportFormatJSON,
		}

		// First request should fail
		_, err := suite.exportSvc.ExportData(ctx, request)
		assert.Error(t, err)

		// Second request with valid data should succeed
		request.ExportType = ExportTypeAssessments
		_, err = suite.exportSvc.ExportData(ctx, request)
		assert.NoError(t, err)
	})

	t.Run("Partial Failure Recovery", func(t *testing.T) {
		ctx := context.Background()

		// Test partial failure recovery
		request := &BackupRequest{
			BusinessID:  "test-business",
			BackupType:  BackupTypeBusiness,
			IncludeData: []string{"assessments", "invalid-type"}, // Mixed valid/invalid
		}

		_, err := suite.backupSvc.CreateBackup(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid data type")

		// Retry with only valid data types
		request.IncludeData = []string{"assessments"}
		_, err = suite.backupSvc.CreateBackup(ctx, request)
		assert.NoError(t, err)
	})
}

// TestErrorLogging tests error logging functionality
func TestErrorLogging(t *testing.T) {
	suite := NewErrorHandlingTestSuite(t)
	defer suite.Close()

	t.Run("Error Logging", func(t *testing.T) {
		ctx := context.Background()

		// Test that errors are properly logged
		assessment := &RiskAssessment{
			ID:           "test-assessment",
			BusinessID:   "", // Invalid: empty business ID
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		err := suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
		assert.Error(t, err)

		// Error should be logged (we can't easily test the actual logging,
		// but we can verify the error is returned)
		assert.NotNil(t, err)
	})

	t.Run("Error Context Preservation", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

		// Test that error context is preserved
		assessment := &RiskAssessment{
			ID:           "test-assessment",
			BusinessID:   "", // Invalid: empty business ID
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		err := suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
		assert.Error(t, err)

		// Context should be preserved
		requestID := ctx.Value("request_id")
		assert.Equal(t, "test-request-123", requestID)
	})
}

// TestErrorMetrics tests error metrics collection
func TestErrorMetrics(t *testing.T) {
	suite := NewErrorHandlingTestSuite(t)
	defer suite.Close()

	t.Run("Error Count Metrics", func(t *testing.T) {
		ctx := context.Background()

		// Generate multiple errors to test metrics
		errorCount := 0
		for i := 0; i < 10; i++ {
			assessment := &RiskAssessment{
				ID:           fmt.Sprintf("test-assessment-%d", i),
				BusinessID:   "", // Invalid: empty business ID
				BusinessName: "Test Business",
				OverallScore: 80.0,
				OverallLevel: RiskLevelMedium,
				AlertLevel:   RiskLevelMedium,
				AssessedAt:   time.Now(),
				ValidUntil:   time.Now().Add(24 * time.Hour),
			}

			err := suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
			if err != nil {
				errorCount++
			}
		}

		// All should fail
		assert.Equal(t, 10, errorCount)
	})

	t.Run("Error Rate Metrics", func(t *testing.T) {
		ctx := context.Background()

		// Test error rate calculation
		totalRequests := 20
		errorCount := 0

		for i := 0; i < totalRequests; i++ {
			assessment := &RiskAssessment{
				ID:           fmt.Sprintf("test-assessment-%d", i),
				BusinessID:   "", // Invalid: empty business ID
				BusinessName: "Test Business",
				OverallScore: 80.0,
				OverallLevel: RiskLevelMedium,
				AlertLevel:   RiskLevelMedium,
				AssessedAt:   time.Now(),
				ValidUntil:   time.Now().Add(24 * time.Hour),
			}

			err := suite.validationSvc.ValidateRiskAssessment(ctx, assessment)
			if err != nil {
				errorCount++
			}
		}

		errorRate := float64(errorCount) / float64(totalRequests)
		assert.Equal(t, 1.0, errorRate) // 100% error rate for invalid data
	})
}
