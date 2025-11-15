package risk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// NewIntegrationTestSuite creates a new integration test suite
// Type definition is in test_suite_types.go
func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	logger := zap.NewNop()
	backupDir := t.TempDir()

	// Create services
	storageService := NewRiskStorageService(logger)
	validationSvc := NewRiskValidationService(logger)
	exportSvc := NewExportService(logger)
	backupSvc := NewBackupService(logger, backupDir, 30, false)
	jobManager := NewExportJobManager(logger, exportSvc)
	backupJobManager := NewBackupJobManager(logger, backupSvc)

	// Create handlers
	exportHandler := NewExportHandler(logger, exportSvc, jobManager)
	backupHandler := NewBackupHandler(logger, backupSvc, backupJobManager)

	// Create HTTP mux
	mux := http.NewServeMux()
	exportHandler.RegisterRoutes(mux)
	backupHandler.RegisterRoutes(mux)

	return &IntegrationTestSuite{
		logger:           logger,
		backupDir:        backupDir,
		storageService:   storageService,
		validationSvc:    validationSvc,
		exportSvc:        exportSvc,
		backupSvc:        backupSvc,
		jobManager:       jobManager,
		backupJobManager: backupJobManager,
		exportHandler:    exportHandler,
		backupHandler:    backupHandler,
		mux:              mux,
	}
}

// TestEndToEndRiskAssessmentWorkflow tests the complete risk assessment workflow
func TestEndToEndRiskAssessmentWorkflow(t *testing.T) {
	suite := NewIntegrationTestSuite(t)

	// Test data
	businessID := "test-business-123"
	businessName := "Test Business Inc."

	t.Run("Complete Risk Assessment Workflow", func(t *testing.T) {
		// Step 1: Create initial risk assessment
		assessment := &RiskAssessment{
			ID:           "assessment-1",
			BusinessID:   businessID,
			BusinessName: businessName,
			OverallScore: 75.5,
			OverallLevel: RiskLevelHigh,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
			AlertLevel:   RiskLevelMedium,
			CategoryScores: map[RiskCategory]float64{
				RiskCategoryFinancial:   80.0,
				RiskCategoryOperational: 70.0,
				RiskCategoryCompliance:  75.0,
			},
			FactorScores: []RiskScore{
				{
					FactorID:     "financial-1",
					FactorName:   "Financial Stability",
					Category:     RiskCategoryFinancial,
					Score:        80.0,
					Level:        RiskLevelHigh,
					Confidence:   0.9,
					Explanation:  "Strong financial position",
					CalculatedAt: time.Now(),
				},
			},
			Recommendations: []RiskRecommendation{
				{
					ID:          "rec-1",
					Category:    RiskCategoryFinancial,
					Priority:    RecommendationPriorityHigh,
					Title:       "Improve Financial Monitoring",
					Description: "Implement regular financial monitoring",
					Actions:     []string{"Set up monthly reviews", "Implement alerts"},
				},
			},
			Alerts: []RiskAlert{
				{
					ID:             "alert-1",
					BusinessID:     businessID,
					RiskFactor:     "financial-1",
					Level:          RiskLevelHigh,
					Message:        "High financial risk detected",
					Score:          80.0,
					Threshold:      75.0,
					TriggeredAt:    time.Now(),
					Acknowledged:   false,
					AcknowledgedAt: nil,
				},
			},
		}

		// Step 2: Store the assessment
		ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
		err := suite.storageService.StoreRiskAssessment(ctx, assessment)
		require.NoError(t, err)

		// Step 3: Validate the assessment
		err = suite.validationSvc.ValidateRiskAssessment(assessment)
		require.NoError(t, err)

		// Step 4: Retrieve the assessment
		retrievedAssessment, err := suite.storageService.GetRiskAssessment(ctx, assessment.ID)
		require.NoError(t, err)
		assert.Equal(t, assessment.ID, retrievedAssessment.ID)
		assert.Equal(t, assessment.BusinessID, retrievedAssessment.BusinessID)
		assert.Equal(t, assessment.OverallScore, retrievedAssessment.OverallScore)

		// Step 5: Update the assessment
		assessment.OverallScore = 85.0
		assessment.OverallLevel = RiskLevelMedium
		err = suite.storageService.UpdateRiskAssessment(ctx, assessment)
		require.NoError(t, err)

		// Step 6: Verify the update
		updatedAssessment, err := suite.storageService.GetRiskAssessment(ctx, assessment.ID)
		require.NoError(t, err)
		assert.Equal(t, 85.0, updatedAssessment.OverallScore)
		assert.Equal(t, RiskLevelMedium, updatedAssessment.OverallLevel)

		// Step 7: Export the assessment data
		exportRequest := &ExportRequest{
			BusinessID: businessID,
			ExportType: ExportTypeAssessments,
			Format:     ExportFormatJSON,
			Metadata:   map[string]interface{}{"test": "integration"},
		}

		exportResponse, err := suite.exportSvc.ExportRiskAssessments(ctx, []*RiskAssessment{updatedAssessment}, ExportFormatJSON)
		require.NoError(t, err)
		assert.NotNil(t, exportResponse)
		assert.Equal(t, businessID, exportResponse.BusinessID)
		assert.Equal(t, ExportTypeAssessments, exportResponse.ExportType)

		// Step 8: Create a backup
		backupRequest := &BackupRequest{
			BusinessID:  businessID,
			BackupType:  BackupTypeBusiness,
			IncludeData: []BackupDataType{BackupDataTypeAssessments},
			Metadata:    map[string]interface{}{"test": "integration"},
		}

		backupResponse, err := suite.backupSvc.CreateBackup(ctx, backupRequest)
		require.NoError(t, err)
		assert.NotNil(t, backupResponse)
		assert.Equal(t, businessID, backupResponse.BusinessID)
		assert.Equal(t, BackupTypeBusiness, backupResponse.BackupType)

		// Step 9: Restore from backup
		restoreRequest := &RestoreRequest{
			BackupID:    backupResponse.BackupID,
			BusinessID:  businessID,
			RestoreType: RestoreTypeBusiness,
			Metadata:    map[string]interface{}{"test": "integration"},
		}

		restoreResponse, err := suite.backupSvc.RestoreBackup(ctx, restoreRequest)
		require.NoError(t, err)
		assert.NotNil(t, restoreResponse)
		assert.Equal(t, backupResponse.BackupID, restoreResponse.BackupID)
		assert.Equal(t, RestoreTypeBusiness, restoreResponse.RestoreType)

		// Step 10: Clean up - delete the assessment
		err = suite.storageService.DeleteRiskAssessment(ctx, assessment.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = suite.storageService.GetRiskAssessment(ctx, assessment.ID)
		assert.Error(t, err)
	})
}

// TestAPIIntegration tests API integration
func TestAPIIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite(t)

	t.Run("API Integration Tests", func(t *testing.T) {
		// Test that the mux is properly configured
		assert.NotNil(t, suite.mux)
	})

	t.Run("Export API Integration", func(t *testing.T) {
		// Test POST /api/v1/export/jobs
		exportData := map[string]interface{}{
			"business_id": "test-business-123",
			"export_type": "assessments",
			"format":      "json",
			"metadata":    map[string]interface{}{"test": "integration"},
		}

		reqBody, _ := json.Marshal(exportData)
		req := httptest.NewRequest("POST", "/api/v1/export/jobs", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

		w := httptest.NewRecorder()
		suite.mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response["job_id"])
	})

	t.Run("Backup API Integration", func(t *testing.T) {
		// Test POST /api/v1/backup
		backupData := map[string]interface{}{
			"business_id":  "test-business-123",
			"backup_type":  "business",
			"include_data": []string{"assessments"},
			"metadata":     map[string]interface{}{"test": "integration"},
		}

		reqBody, _ := json.Marshal(backupData)
		req := httptest.NewRequest("POST", "/api/v1/backup", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

		w := httptest.NewRecorder()
		suite.mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response["backup_id"])
	})
}

// TestDatabaseIntegration tests database integration
func TestDatabaseIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite(t)

	t.Run("Database CRUD Operations", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

		// Create multiple assessments
		assessments := []*RiskAssessment{
			{
				ID:           "assessment-1",
				BusinessID:   "business-1",
				BusinessName: "Business 1",
				OverallScore: 75.0,
				OverallLevel: RiskLevelHigh,
				AssessedAt:   time.Now(),
				ValidUntil:   time.Now().Add(24 * time.Hour),
				AlertLevel:   RiskLevelMedium,
			},
			{
				ID:           "assessment-2",
				BusinessID:   "business-2",
				BusinessName: "Business 2",
				OverallScore: 85.0,
				OverallLevel: RiskLevelMedium,
				AssessedAt:   time.Now(),
				ValidUntil:   time.Now().Add(24 * time.Hour),
				AlertLevel:   RiskLevelLow,
			},
			{
				ID:           "assessment-3",
				BusinessID:   "business-1",
				BusinessName: "Business 1",
				OverallScore: 90.0,
				OverallLevel: RiskLevelLow,
				AssessedAt:   time.Now(),
				ValidUntil:   time.Now().Add(24 * time.Hour),
				AlertLevel:   RiskLevelLow,
			},
		}

		// Store all assessments
		for _, assessment := range assessments {
			err := suite.storageService.StoreRiskAssessment(ctx, assessment)
			require.NoError(t, err)
		}

		// Test listing assessments by business
		business1Assessments, err := suite.storageService.ListRiskAssessments(ctx, "business-1", 10, 0)
		require.NoError(t, err)
		assert.Len(t, business1Assessments, 2)

		business2Assessments, err := suite.storageService.ListRiskAssessments(ctx, "business-2", 10, 0)
		require.NoError(t, err)
		assert.Len(t, business2Assessments, 1)

		// Test pagination
		paginatedAssessments, err := suite.storageService.ListRiskAssessments(ctx, "business-1", 1, 0)
		require.NoError(t, err)
		assert.Len(t, paginatedAssessments, 1)

		// Test updating assessment
		assessment := assessments[0]
		assessment.OverallScore = 95.0
		assessment.OverallLevel = RiskLevelLow
		err = suite.storageService.UpdateRiskAssessment(ctx, assessment)
		require.NoError(t, err)

		// Verify update
		updatedAssessment, err := suite.storageService.GetRiskAssessment(ctx, assessment.ID)
		require.NoError(t, err)
		assert.Equal(t, 95.0, updatedAssessment.OverallScore)
		assert.Equal(t, RiskLevelLow, updatedAssessment.OverallLevel)

		// Test deleting assessment
		err = suite.storageService.DeleteRiskAssessment(ctx, assessment.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = suite.storageService.GetRiskAssessment(ctx, assessment.ID)
		assert.Error(t, err)

		// Clean up remaining assessments
		for _, assessment := range assessments[1:] {
			suite.storageService.DeleteRiskAssessment(ctx, assessment.ID)
		}
	})
}

// TestErrorHandling tests error handling scenarios
func TestErrorHandling(t *testing.T) {
	suite := NewIntegrationTestSuite(t)

	t.Run("Invalid Data Handling", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

		// Test invalid assessment data
		invalidAssessment := &RiskAssessment{
			ID:           "", // Invalid: empty ID
			BusinessID:   "test-business-123",
			BusinessName: "Test Business",
			OverallScore: -10.0, // Invalid: negative score
			OverallLevel: RiskLevelHigh,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(-24 * time.Hour), // Invalid: past date
			AlertLevel:   RiskLevelMedium,
		}

		// Validation should fail
		err := suite.validationSvc.ValidateRiskAssessment(invalidAssessment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID is required")

		// Storage should fail
		err = suite.storageService.StoreRiskAssessment(ctx, invalidAssessment)
		assert.Error(t, err)
	})

	t.Run("Non-existent Resource Handling", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

		// Test getting non-existent assessment
		_, err := suite.storageService.GetRiskAssessment(ctx, "non-existent-id")
		assert.Error(t, err)

		// Test updating non-existent assessment
		assessment := &RiskAssessment{
			ID:           "non-existent-id",
			BusinessID:   "test-business-123",
			BusinessName: "Test Business",
			OverallScore: 75.0,
			OverallLevel: RiskLevelHigh,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
			AlertLevel:   RiskLevelMedium,
		}
		err = suite.storageService.UpdateRiskAssessment(ctx, assessment)
		assert.Error(t, err)

		// Test deleting non-existent assessment
		err = suite.storageService.DeleteRiskAssessment(ctx, "non-existent-id")
		assert.Error(t, err)
	})

	t.Run("Invalid Export Request", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

		// Test invalid export request
		invalidRequest := &ExportRequest{
			BusinessID: "test-business-123",
			ExportType: "invalid-type", // Invalid export type
			Format:     ExportFormatJSON,
		}

		_, err := suite.exportSvc.ExportRiskAssessments(ctx, []*RiskAssessment{}, ExportFormatJSON)
		// This should work with empty data, but let's test with invalid format
		_, err = suite.exportSvc.ExportRiskAssessments(ctx, []*RiskAssessment{}, "invalid-format")
		assert.Error(t, err)
	})

	t.Run("Invalid Backup Request", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

		// Test invalid backup request
		invalidRequest := &BackupRequest{
			BusinessID:  "test-business-123",
			BackupType:  "invalid-type", // Invalid backup type
			IncludeData: []BackupDataType{BackupDataTypeAssessments},
		}

		_, err := suite.backupSvc.CreateBackup(ctx, invalidRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid backup type")
	})
}

// TestPerformance tests performance scenarios
func TestPerformance(t *testing.T) {
	suite := NewIntegrationTestSuite(t)

	t.Run("Concurrent Operations", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

		// Test concurrent assessment creation
		numGoroutines := 10
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(i int) {
				assessment := &RiskAssessment{
					ID:           fmt.Sprintf("assessment-%d", i),
					BusinessID:   fmt.Sprintf("business-%d", i),
					BusinessName: fmt.Sprintf("Business %d", i),
					OverallScore: 75.0 + float64(i),
					OverallLevel: RiskLevelHigh,
					AssessedAt:   time.Now(),
					ValidUntil:   time.Now().Add(24 * time.Hour),
					AlertLevel:   RiskLevelMedium,
				}

				err := suite.storageService.StoreRiskAssessment(ctx, assessment)
				results <- err
			}(i)
		}

		// Wait for all operations to complete
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(t, err)
		}

		// Verify all assessments were created
		for i := 0; i < numGoroutines; i++ {
			assessment, err := suite.storageService.GetRiskAssessment(ctx, fmt.Sprintf("assessment-%d", i))
			assert.NoError(t, err)
			assert.NotNil(t, assessment)
		}

		// Clean up
		for i := 0; i < numGoroutines; i++ {
			suite.storageService.DeleteRiskAssessment(ctx, fmt.Sprintf("assessment-%d", i))
		}
	})

	t.Run("Large Dataset Operations", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

		// Create a large number of assessments
		numAssessments := 100
		assessments := make([]*RiskAssessment, numAssessments)

		for i := 0; i < numAssessments; i++ {
			assessments[i] = &RiskAssessment{
				ID:           fmt.Sprintf("large-assessment-%d", i),
				BusinessID:   "large-business",
				BusinessName: "Large Business",
				OverallScore: 75.0 + float64(i%25),
				OverallLevel: RiskLevelHigh,
				AssessedAt:   time.Now(),
				ValidUntil:   time.Now().Add(24 * time.Hour),
				AlertLevel:   RiskLevelMedium,
			}
		}

		// Store all assessments
		start := time.Now()
		for _, assessment := range assessments {
			err := suite.storageService.StoreRiskAssessment(ctx, assessment)
			require.NoError(t, err)
		}
		storageTime := time.Since(start)

		// Test listing with large dataset
		start = time.Now()
		listAssessments, err := suite.storageService.ListRiskAssessments(ctx, "large-business", 100, 0)
		require.NoError(t, err)
		listTime := time.Since(start)

		assert.Len(t, listAssessments, numAssessments)

		// Performance assertions (adjust thresholds as needed)
		assert.Less(t, storageTime, 5*time.Second, "Storage should complete within 5 seconds")
		assert.Less(t, listTime, 2*time.Second, "Listing should complete within 2 seconds")

		// Clean up
		for _, assessment := range assessments {
			suite.storageService.DeleteRiskAssessment(ctx, assessment.ID)
		}
	})

	t.Run("Export Performance", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

		// Create test data
		assessments := make([]*RiskAssessment, 50)
		for i := 0; i < 50; i++ {
			assessments[i] = &RiskAssessment{
				ID:           fmt.Sprintf("export-assessment-%d", i),
				BusinessID:   "export-business",
				BusinessName: "Export Business",
				OverallScore: 75.0 + float64(i%25),
				OverallLevel: RiskLevelHigh,
				AssessedAt:   time.Now(),
				ValidUntil:   time.Now().Add(24 * time.Hour),
				AlertLevel:   RiskLevelMedium,
			}
		}

		// Test export performance
		start := time.Now()
		exportResponse, err := suite.exportSvc.ExportRiskAssessments(ctx, assessments, ExportFormatJSON)
		exportTime := time.Since(start)

		require.NoError(t, err)
		assert.NotNil(t, exportResponse)
		assert.Less(t, exportTime, 3*time.Second, "Export should complete within 3 seconds")
	})

	t.Run("Backup Performance", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

		// Test backup performance
		backupRequest := &BackupRequest{
			BusinessID:  "performance-business",
			BackupType:  BackupTypeFull,
			IncludeData: []BackupDataType{BackupDataTypeAssessments, BackupDataTypeFactors, BackupDataTypeTrends},
		}

		start := time.Now()
		backupResponse, err := suite.backupSvc.CreateBackup(ctx, backupRequest)
		backupTime := time.Since(start)

		require.NoError(t, err)
		assert.NotNil(t, backupResponse)
		assert.Less(t, backupTime, 5*time.Second, "Backup should complete within 5 seconds")
	})
}

// TestDataIntegrity tests data integrity scenarios
func TestDataIntegrity(t *testing.T) {
	suite := NewIntegrationTestSuite(t)

	t.Run("Data Consistency", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "test-request-123")

		// Create assessment with complex data
		assessment := &RiskAssessment{
			ID:           "integrity-assessment-1",
			BusinessID:   "integrity-business",
			BusinessName: "Integrity Business",
			OverallScore: 75.5,
			OverallLevel: RiskLevelHigh,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
			AlertLevel:   RiskLevelMedium,
			CategoryScores: map[RiskCategory]float64{
				RiskCategoryFinancial:    80.0,
				RiskCategoryOperational:  70.0,
				RiskCategoryCompliance:   75.0,
				RiskCategoryReputational: 65.0,
			},
			FactorScores: []RiskScore{
				{
					FactorID:     "financial-1",
					FactorName:   "Financial Stability",
					Category:     RiskCategoryFinancial,
					Score:        80.0,
					Level:        RiskLevelHigh,
					Confidence:   0.9,
					Explanation:  "Strong financial position",
					Evidence:     []string{"High cash reserves", "Low debt ratio"},
					CalculatedAt: time.Now(),
				},
				{
					FactorID:     "operational-1",
					FactorName:   "Operational Efficiency",
					Category:     RiskCategoryOperational,
					Score:        70.0,
					Level:        RiskLevelMedium,
					Confidence:   0.8,
					Explanation:  "Moderate operational efficiency",
					Evidence:     []string{"Standard processes", "Adequate staffing"},
					CalculatedAt: time.Now(),
				},
			},
			Recommendations: []RiskRecommendation{
				{
					ID:          "rec-1",
					Category:    RiskCategoryFinancial,
					Priority:    RecommendationPriorityHigh,
					Title:       "Improve Financial Monitoring",
					Description: "Implement regular financial monitoring",
					Actions:     []string{"Set up monthly reviews", "Implement alerts"},
					Impact:      "High",
					Effort:      "Medium",
					Timeline:    "3 months",
				},
			},
			Alerts: []RiskAlert{
				{
					ID:             "alert-1",
					BusinessID:     "integrity-business",
					RiskFactor:     "financial-1",
					Level:          RiskLevelHigh,
					Message:        "High financial risk detected",
					Score:          80.0,
					Threshold:      75.0,
					TriggeredAt:    time.Now(),
					Acknowledged:   false,
					AcknowledgedAt: nil,
				},
			},
		}

		// Store assessment
		err := suite.storageService.StoreRiskAssessment(ctx, assessment)
		require.NoError(t, err)

		// Retrieve and verify data integrity
		retrievedAssessment, err := suite.storageService.GetRiskAssessment(ctx, assessment.ID)
		require.NoError(t, err)

		// Verify all fields are preserved
		assert.Equal(t, assessment.ID, retrievedAssessment.ID)
		assert.Equal(t, assessment.BusinessID, retrievedAssessment.BusinessID)
		assert.Equal(t, assessment.BusinessName, retrievedAssessment.BusinessName)
		assert.Equal(t, assessment.OverallScore, retrievedAssessment.OverallScore)
		assert.Equal(t, assessment.OverallLevel, retrievedAssessment.OverallLevel)
		assert.Equal(t, assessment.AlertLevel, retrievedAssessment.AlertLevel)

		// Verify category scores
		assert.Equal(t, len(assessment.CategoryScores), len(retrievedAssessment.CategoryScores))
		for category, score := range assessment.CategoryScores {
			assert.Equal(t, score, retrievedAssessment.CategoryScores[category])
		}

		// Verify factor scores
		assert.Equal(t, len(assessment.FactorScores), len(retrievedAssessment.FactorScores))
		for i, factor := range assessment.FactorScores {
			retrievedFactor := retrievedAssessment.FactorScores[i]
			assert.Equal(t, factor.FactorID, retrievedFactor.FactorID)
			assert.Equal(t, factor.FactorName, retrievedFactor.FactorName)
			assert.Equal(t, factor.Category, retrievedFactor.Category)
			assert.Equal(t, factor.Score, retrievedFactor.Score)
			assert.Equal(t, factor.Level, retrievedFactor.Level)
			assert.Equal(t, factor.Confidence, retrievedFactor.Confidence)
			assert.Equal(t, factor.Explanation, retrievedFactor.Explanation)
			assert.Equal(t, factor.Evidence, retrievedFactor.Evidence)
		}

		// Verify recommendations
		assert.Equal(t, len(assessment.Recommendations), len(retrievedAssessment.Recommendations))
		for i, rec := range assessment.Recommendations {
			retrievedRec := retrievedAssessment.Recommendations[i]
			assert.Equal(t, rec.ID, retrievedRec.ID)
			assert.Equal(t, rec.Category, retrievedRec.Category)
			assert.Equal(t, rec.Priority, retrievedRec.Priority)
			assert.Equal(t, rec.Title, retrievedRec.Title)
			assert.Equal(t, rec.Description, retrievedRec.Description)
			assert.Equal(t, rec.Actions, retrievedRec.Actions)
			assert.Equal(t, rec.Impact, retrievedRec.Impact)
			assert.Equal(t, rec.Effort, retrievedRec.Effort)
			assert.Equal(t, rec.Timeline, retrievedRec.Timeline)
		}

		// Verify alerts
		assert.Equal(t, len(assessment.Alerts), len(retrievedAssessment.Alerts))
		for i, alert := range assessment.Alerts {
			retrievedAlert := retrievedAssessment.Alerts[i]
			assert.Equal(t, alert.ID, retrievedAlert.ID)
			assert.Equal(t, alert.BusinessID, retrievedAlert.BusinessID)
			assert.Equal(t, alert.RiskFactor, retrievedAlert.RiskFactor)
			assert.Equal(t, alert.Level, retrievedAlert.Level)
			assert.Equal(t, alert.Message, retrievedAlert.Message)
			assert.Equal(t, alert.Score, retrievedAlert.Score)
			assert.Equal(t, alert.Threshold, retrievedAlert.Threshold)
			assert.Equal(t, alert.Acknowledged, retrievedAlert.Acknowledged)
		}

		// Clean up
		err = suite.storageService.DeleteRiskAssessment(ctx, assessment.ID)
		require.NoError(t, err)
	})
}

// TestWorkflowIntegration tests workflow integration scenarios
func TestWorkflowIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite(t)

	t.Run("Complete Risk Management Workflow", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
		businessID := "workflow-business-123"

		// Step 1: Create initial assessment
		assessment := &RiskAssessment{
			ID:           "workflow-assessment-1",
			BusinessID:   businessID,
			BusinessName: "Workflow Business",
			OverallScore: 60.0,
			OverallLevel: RiskLevelHigh,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
			AlertLevel:   RiskLevelHigh,
		}

		err := suite.storageService.StoreRiskAssessment(ctx, assessment)
		require.NoError(t, err)

		// Step 2: Export assessment data
		exportResponse, err := suite.exportSvc.ExportRiskAssessments(ctx, []*RiskAssessment{assessment}, ExportFormatJSON)
		require.NoError(t, err)
		assert.NotNil(t, exportResponse)

		// Step 3: Create backup
		backupResponse, err := suite.backupSvc.CreateBackup(ctx, &BackupRequest{
			BusinessID:  businessID,
			BackupType:  BackupTypeBusiness,
			IncludeData: []BackupDataType{BackupDataTypeAssessments},
		})
		require.NoError(t, err)
		assert.NotNil(t, backupResponse)

		// Step 4: Update assessment (risk improvement)
		assessment.OverallScore = 80.0
		assessment.OverallLevel = RiskLevelMedium
		assessment.AlertLevel = RiskLevelMedium
		err = suite.storageService.UpdateRiskAssessment(ctx, assessment)
		require.NoError(t, err)

		// Step 5: Export updated data
		updatedExportResponse, err := suite.exportSvc.ExportRiskAssessments(ctx, []*RiskAssessment{assessment}, ExportFormatJSON)
		require.NoError(t, err)
		assert.NotNil(t, updatedExportResponse)

		// Step 6: Create another backup
		backupResponse2, err := suite.backupSvc.CreateBackup(ctx, &BackupRequest{
			BusinessID:  businessID,
			BackupType:  BackupTypeIncremental,
			IncludeData: []BackupDataType{BackupDataTypeAssessments},
		})
		require.NoError(t, err)
		assert.NotNil(t, backupResponse2)

		// Step 7: Restore from first backup
		restoreResponse, err := suite.backupSvc.RestoreBackup(ctx, &RestoreRequest{
			BackupID:    backupResponse.BackupID,
			BusinessID:  businessID,
			RestoreType: RestoreTypeBusiness,
		})
		require.NoError(t, err)
		assert.NotNil(t, restoreResponse)

		// Step 8: Verify data consistency
		retrievedAssessment, err := suite.storageService.GetRiskAssessment(ctx, assessment.ID)
		require.NoError(t, err)
		assert.Equal(t, 60.0, retrievedAssessment.OverallScore) // Should be restored to original value
		assert.Equal(t, RiskLevelHigh, retrievedAssessment.OverallLevel)

		// Clean up
		suite.storageService.DeleteRiskAssessment(ctx, assessment.ID)
	})
}

// BenchmarkIntegration provides performance benchmarks
func BenchmarkIntegration(b *testing.B) {
	suite := NewIntegrationTestSuite(&testing.T{})
	ctx := context.WithValue(context.Background(), "request_id", "benchmark-request")

	b.Run("StoreRiskAssessment", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			assessment := &RiskAssessment{
				ID:           fmt.Sprintf("benchmark-assessment-%d", i),
				BusinessID:   "benchmark-business",
				BusinessName: "Benchmark Business",
				OverallScore: 75.0,
				OverallLevel: RiskLevelHigh,
				AssessedAt:   time.Now(),
				ValidUntil:   time.Now().Add(24 * time.Hour),
				AlertLevel:   RiskLevelMedium,
			}

			suite.storageService.StoreRiskAssessment(ctx, assessment)
		}
	})

	b.Run("GetRiskAssessment", func(b *testing.B) {
		// Create a test assessment first
		assessment := &RiskAssessment{
			ID:           "benchmark-get-assessment",
			BusinessID:   "benchmark-business",
			BusinessName: "Benchmark Business",
			OverallScore: 75.0,
			OverallLevel: RiskLevelHigh,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
			AlertLevel:   RiskLevelMedium,
		}
		suite.storageService.StoreRiskAssessment(ctx, assessment)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			suite.storageService.GetRiskAssessment(ctx, assessment.ID)
		}

		// Clean up
		suite.storageService.DeleteRiskAssessment(ctx, assessment.ID)
	})

	b.Run("ExportRiskAssessments", func(b *testing.B) {
		// Create test data
		assessments := make([]*RiskAssessment, 10)
		for i := 0; i < 10; i++ {
			assessments[i] = &RiskAssessment{
				ID:           fmt.Sprintf("benchmark-export-%d", i),
				BusinessID:   "benchmark-business",
				BusinessName: "Benchmark Business",
				OverallScore: 75.0,
				OverallLevel: RiskLevelHigh,
				AssessedAt:   time.Now(),
				ValidUntil:   time.Now().Add(24 * time.Hour),
				AlertLevel:   RiskLevelMedium,
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			suite.exportSvc.ExportRiskAssessments(ctx, assessments, ExportFormatJSON)
		}
	})
}
