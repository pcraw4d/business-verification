package risk

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// NewDatabaseIntegrationTestSuite creates a new database integration test suite
// Type definition is in test_suite_types.go
func NewDatabaseIntegrationTestSuite(t *testing.T) *DatabaseIntegrationTestSuite {
	logger := zap.NewNop()
	backupDir := t.TempDir()

	// Create mock database connection
	// In a real implementation, this would connect to a test database
	db := &sql.DB{} // Mock database connection

	// Create services
	storageService := NewRiskStorageService(logger, db)
	validationSvc := NewRiskValidationService(logger)
	exportSvc := NewExportService(logger)
	backupSvc := NewBackupService(logger, backupDir, 30, false)

	// Generate test data
	testData := generateDatabaseTestData()

	return &DatabaseIntegrationTestSuite{
		logger:           logger,
		db:               db,
		storageService:   storageService,
		validationSvc:    validationSvc,
		exportSvc:        exportSvc,
		backupSvc:        backupSvc,
		backupDir:        backupDir,
		testData:         testData,
		cleanupFunctions: make([]func() error, 0),
	}
}

// generateDatabaseTestData generates comprehensive test data for database testing
func generateDatabaseTestData() *DatabaseTestData {
	businessID := "test-business-db-123"
	assessmentID := "test-assessment-db-123"

	// Generate risk factors
	riskFactors := []RiskFactorInput{
		{
			ID:          "factor-1",
			Name:        "Financial Stability",
			Description: "Assessment of financial stability",
			Weight:      0.3,
			Value:       85.0,
		},
		{
			ID:          "factor-2",
			Name:        "Operational Risk",
			Description: "Assessment of operational risk",
			Weight:      0.4,
			Value:       70.0,
		},
		{
			ID:          "factor-3",
			Name:        "Compliance Risk",
			Description: "Assessment of compliance risk",
			Weight:      0.3,
			Value:       75.0,
		},
	}

	// Generate risk scores
	riskScores := []RiskScore{
		{
			ID:          "score-1",
			FactorID:    "factor-1",
			Score:       85.0,
			Level:       RiskLevelLow,
			Confidence:  0.95,
			LastUpdated: time.Now(),
		},
		{
			ID:          "score-2",
			FactorID:    "factor-2",
			Score:       70.0,
			Level:       RiskLevelMedium,
			Confidence:  0.85,
			LastUpdated: time.Now(),
		},
		{
			ID:          "score-3",
			FactorID:    "factor-3",
			Score:       75.0,
			Level:       RiskLevelMedium,
			Confidence:  0.90,
			LastUpdated: time.Now(),
		},
	}

	// Generate risk alerts
	riskAlerts := []RiskAlert{
		{
			ID:         "alert-1",
			BusinessID: businessID,
			RiskFactor: "factor-2",
			AlertType:  "threshold_exceeded",
			AlertLevel: RiskLevelMedium,
			Message:    "Operational risk threshold exceeded",
			IsActive:   true,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "alert-2",
			BusinessID: businessID,
			RiskFactor: "factor-3",
			AlertType:  "trend_warning",
			AlertLevel: RiskLevelLow,
			Message:    "Compliance risk trend warning",
			IsActive:   true,
			CreatedAt:  time.Now(),
		},
	}

	// Generate risk trends
	riskTrends := []RiskTrend{
		{
			ID:         "trend-1",
			BusinessID: businessID,
			FactorID:   "factor-1",
			TrendType:  "improving",
			TrendValue: 5.0,
			Confidence: 0.80,
			Period:     "30d",
			CreatedAt:  time.Now(),
		},
		{
			ID:         "trend-2",
			BusinessID: businessID,
			FactorID:   "factor-2",
			TrendType:  "deteriorating",
			TrendValue: -3.0,
			Confidence: 0.75,
			Period:     "30d",
			CreatedAt:  time.Now(),
		},
	}

	// Generate risk history
	riskHistory := []RiskHistoryEntry{
		{
			BusinessID: businessID,
			Score:      80.0,
			Timestamp:  time.Now().Add(-24 * time.Hour),
			Details: map[string]interface{}{
				"change_type": "score_update",
				"old_value":   "75.0",
				"new_value":   "80.0",
				"changed_by":  "system",
			},
		},
		{
			BusinessID: businessID,
			Score:      75.0,
			Timestamp:  time.Now().Add(-48 * time.Hour),
			Details: map[string]interface{}{
				"change_type": "score_update",
				"old_value":   "70.0",
				"new_value":   "75.0",
				"changed_by":  "system",
			},
		},
	}

	// Generate export jobs
	exportJobs := []ExportJob{
		{
			ID:          "export-job-1",
			BusinessID:  businessID,
			ExportType:  ExportTypeAssessments,
			Format:      ExportFormatJSON,
			Status:      ExportStatusCompleted,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			CompletedAt: time.Now().Add(-30 * time.Minute),
		},
		{
			ID:         "export-job-2",
			BusinessID: businessID,
			ExportType: ExportTypeFactors,
			Format:     ExportFormatCSV,
			Status:     ExportStatusPending,
			CreatedAt:  time.Now().Add(-10 * time.Minute),
		},
	}

	// Generate backup jobs
	backupJobs := []BackupJob{
		{
			ID:          "backup-job-1",
			BusinessID:  businessID,
			BackupType:  BackupTypeBusiness,
			Status:      BackupStatusCompleted,
			CreatedAt:   time.Now().Add(-2 * time.Hour),
			CompletedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:         "backup-job-2",
			BusinessID: businessID,
			BackupType: BackupTypeIncremental,
			Status:     BackupStatusInProgress,
			CreatedAt:  time.Now().Add(-5 * time.Minute),
		},
	}

	// Generate backup schedules
	backupSchedules := []BackupSchedule{
		{
			ID:            "schedule-1",
			BusinessID:    businessID,
			Name:          "Daily Backup",
			Description:   "Daily backup schedule",
			BackupType:    BackupTypeBusiness,
			Schedule:      "0 2 * * *",
			RetentionDays: 30,
			Enabled:       true,
			CreatedAt:     time.Now().Add(-7 * 24 * time.Hour),
		},
		{
			ID:            "schedule-2",
			BusinessID:    businessID,
			Name:          "Weekly Backup",
			Description:   "Weekly backup schedule",
			BackupType:    BackupTypeFull,
			Schedule:      "0 3 * * 0",
			RetentionDays: 90,
			Enabled:       true,
			CreatedAt:     time.Now().Add(-14 * 24 * time.Hour),
		},
	}

	return &DatabaseTestData{
		BusinessID:      businessID,
		AssessmentID:    assessmentID,
		RiskFactors:     riskFactors,
		RiskScores:      riskScores,
		RiskAlerts:      riskAlerts,
		RiskTrends:      riskTrends,
		RiskHistory:     riskHistory,
		ExportJobs:      exportJobs,
		BackupJobs:      backupJobs,
		BackupSchedules: backupSchedules,
	}
}

// TestDatabaseCRUDOperations tests all CRUD operations on database entities
func TestDatabaseCRUDOperations(t *testing.T) {
	suite := NewDatabaseIntegrationTestSuite(t)
	defer suite.cleanup()

	t.Run("Risk Assessment CRUD Operations", func(t *testing.T) {
		ctx := context.Background()

		// Test Create
		assessment := &RiskAssessment{
			ID:           suite.testData.AssessmentID,
			BusinessID:   suite.testData.BusinessID,
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		err := suite.storageService.SaveRiskAssessment(ctx, assessment)
		require.NoError(t, err)

		// Test Read
		retrievedAssessment, err := suite.storageService.GetRiskAssessment(ctx, suite.testData.AssessmentID)
		require.NoError(t, err)
		assert.Equal(t, assessment.ID, retrievedAssessment.ID)
		assert.Equal(t, assessment.BusinessID, retrievedAssessment.BusinessID)
		assert.Equal(t, assessment.OverallScore, retrievedAssessment.OverallScore)
		assert.Equal(t, assessment.OverallLevel, retrievedAssessment.OverallLevel)

		// Test Update
		assessment.OverallScore = 85.0
		assessment.OverallLevel = RiskLevelLow

		err = suite.storageService.UpdateRiskAssessment(ctx, assessment)
		require.NoError(t, err)

		// Verify update
		updatedAssessment, err := suite.storageService.GetRiskAssessment(ctx, suite.testData.AssessmentID)
		require.NoError(t, err)
		assert.Equal(t, 85.0, updatedAssessment.OverallScore)
		assert.Equal(t, RiskLevelLow, updatedAssessment.OverallLevel)

		// Test Delete
		err = suite.storageService.DeleteRiskAssessment(ctx, suite.testData.AssessmentID)
		require.NoError(t, err)

		// Verify deletion
		_, err = suite.storageService.GetRiskAssessment(ctx, suite.testData.AssessmentID)
		assert.Error(t, err)
	})

	t.Run("Risk Factor CRUD Operations", func(t *testing.T) {
		ctx := context.Background()

		for _, factor := range suite.testData.RiskFactors {
			// Test Create
			err := suite.storageService.SaveRiskFactor(ctx, &factor)
			require.NoError(t, err)

			// Test Read
			retrievedFactor, err := suite.storageService.GetRiskFactor(ctx, factor.ID)
			require.NoError(t, err)
			assert.Equal(t, factor.ID, retrievedFactor.ID)
			assert.Equal(t, factor.Name, retrievedFactor.Name)
			assert.Equal(t, factor.Weight, retrievedFactor.Weight)
			assert.Equal(t, factor.Value, retrievedFactor.Value)

			// Test Update
			factor.Value = factor.Value + 5.0
			err = suite.storageService.UpdateRiskFactor(ctx, &factor)
			require.NoError(t, err)

			// Verify update
			updatedFactor, err := suite.storageService.GetRiskFactor(ctx, factor.ID)
			require.NoError(t, err)
			assert.Equal(t, factor.Value, updatedFactor.Value)

			// Test Delete
			err = suite.storageService.DeleteRiskFactor(ctx, factor.ID)
			require.NoError(t, err)

			// Verify deletion
			_, err = suite.storageService.GetRiskFactor(ctx, factor.ID)
			assert.Error(t, err)
		}
	})

	t.Run("Risk Score CRUD Operations", func(t *testing.T) {
		ctx := context.Background()

		for _, score := range suite.testData.RiskScores {
			// Test Create
			err := suite.storageService.SaveRiskScore(ctx, &score)
			require.NoError(t, err)

			// Test Read
			retrievedScore, err := suite.storageService.GetRiskScore(ctx, score.ID)
			require.NoError(t, err)
			assert.Equal(t, score.ID, retrievedScore.ID)
			assert.Equal(t, score.FactorID, retrievedScore.FactorID)
			assert.Equal(t, score.Score, retrievedScore.Score)
			assert.Equal(t, score.Level, retrievedScore.Level)

			// Test Update
			score.Score = score.Score + 2.0
			score.LastUpdated = time.Now()
			err = suite.storageService.UpdateRiskScore(ctx, &score)
			require.NoError(t, err)

			// Verify update
			updatedScore, err := suite.storageService.GetRiskScore(ctx, score.ID)
			require.NoError(t, err)
			assert.Equal(t, score.Score, updatedScore.Score)

			// Test Delete
			err = suite.storageService.DeleteRiskScore(ctx, score.ID)
			require.NoError(t, err)

			// Verify deletion
			_, err = suite.storageService.GetRiskScore(ctx, score.ID)
			assert.Error(t, err)
		}
	})

	t.Run("Risk Alert CRUD Operations", func(t *testing.T) {
		ctx := context.Background()

		for _, alert := range suite.testData.RiskAlerts {
			// Test Create
			err := suite.storageService.SaveRiskAlert(ctx, &alert)
			require.NoError(t, err)

			// Test Read
			retrievedAlert, err := suite.storageService.GetRiskAlert(ctx, alert.ID)
			require.NoError(t, err)
			assert.Equal(t, alert.ID, retrievedAlert.ID)
			assert.Equal(t, alert.BusinessID, retrievedAlert.BusinessID)
			assert.Equal(t, alert.RiskFactor, retrievedAlert.RiskFactor)
			assert.Equal(t, alert.AlertType, retrievedAlert.AlertType)

			// Test Update
			alert.IsActive = false
			alert.Message = "Alert resolved"
			err = suite.storageService.UpdateRiskAlert(ctx, &alert)
			require.NoError(t, err)

			// Verify update
			updatedAlert, err := suite.storageService.GetRiskAlert(ctx, alert.ID)
			require.NoError(t, err)
			assert.Equal(t, false, updatedAlert.IsActive)
			assert.Equal(t, "Alert resolved", updatedAlert.Message)

			// Test Delete
			err = suite.storageService.DeleteRiskAlert(ctx, alert.ID)
			require.NoError(t, err)

			// Verify deletion
			_, err = suite.storageService.GetRiskAlert(ctx, alert.ID)
			assert.Error(t, err)
		}
	})

	t.Run("Risk Trend CRUD Operations", func(t *testing.T) {
		ctx := context.Background()

		for _, trend := range suite.testData.RiskTrends {
			// Test Create
			err := suite.storageService.SaveRiskTrend(ctx, &trend)
			require.NoError(t, err)

			// Test Read
			retrievedTrend, err := suite.storageService.GetRiskTrend(ctx, trend.ID)
			require.NoError(t, err)
			assert.Equal(t, trend.ID, retrievedTrend.ID)
			assert.Equal(t, trend.BusinessID, retrievedTrend.BusinessID)
			assert.Equal(t, trend.FactorID, retrievedTrend.FactorID)
			assert.Equal(t, trend.TrendType, retrievedTrend.TrendType)

			// Test Update
			trend.TrendValue = trend.TrendValue + 1.0
			trend.Confidence = trend.Confidence + 0.05
			err = suite.storageService.UpdateRiskTrend(ctx, &trend)
			require.NoError(t, err)

			// Verify update
			updatedTrend, err := suite.storageService.GetRiskTrend(ctx, trend.ID)
			require.NoError(t, err)
			assert.Equal(t, trend.TrendValue, updatedTrend.TrendValue)
			assert.Equal(t, trend.Confidence, updatedTrend.Confidence)

			// Test Delete
			err = suite.storageService.DeleteRiskTrend(ctx, trend.ID)
			require.NoError(t, err)

			// Verify deletion
			_, err = suite.storageService.GetRiskTrend(ctx, trend.ID)
			assert.Error(t, err)
		}
	})

	t.Run("Risk History CRUD Operations", func(t *testing.T) {
		ctx := context.Background()

		for _, history := range suite.testData.RiskHistory {
			// Test Create
			err := suite.storageService.SaveRiskHistory(ctx, &history)
			require.NoError(t, err)

			// Test Read
			retrievedHistory, err := suite.storageService.GetRiskHistory(ctx, history.BusinessID)
			require.NoError(t, err)
			assert.NotEmpty(t, retrievedHistory)
			assert.Equal(t, history.BusinessID, retrievedHistory[0].BusinessID)
			assert.Equal(t, history.Score, retrievedHistory[0].Score)

			// Test Update (history entries are typically immutable, so we test by adding new entries)
			newHistory := RiskHistoryEntry{
				BusinessID: history.BusinessID,
				Score:      history.Score + 5.0,
				Timestamp:  time.Now(),
				Details: map[string]interface{}{
					"change_type": "score_update",
					"old_value":   fmt.Sprintf("%.1f", history.Score),
					"new_value":   fmt.Sprintf("%.1f", history.Score+5.0),
					"changed_by":  "system",
				},
			}

			err = suite.storageService.SaveRiskHistory(ctx, &newHistory)
			require.NoError(t, err)

			// Verify new entry
			updatedHistory, err := suite.storageService.GetRiskHistory(ctx, history.BusinessID)
			require.NoError(t, err)
			assert.Len(t, updatedHistory, 2) // Should have 2 entries now

			// Test Delete (clean up)
			err = suite.storageService.DeleteRiskHistory(ctx, history.BusinessID)
			require.NoError(t, err)

			// Verify deletion
			deletedHistory, err := suite.storageService.GetRiskHistory(ctx, history.BusinessID)
			require.NoError(t, err)
			assert.Empty(t, deletedHistory)
		}
	})
}

// TestDatabaseQueries tests complex database queries and operations
func TestDatabaseQueries(t *testing.T) {
	suite := NewDatabaseIntegrationTestSuite(t)
	defer suite.cleanup()

	t.Run("Business Risk Assessment Queries", func(t *testing.T) {
		ctx := context.Background()

		// Create test data
		assessment := &RiskAssessment{
			ID:           suite.testData.AssessmentID,
			BusinessID:   suite.testData.BusinessID,
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		err := suite.storageService.SaveRiskAssessment(ctx, assessment)
		require.NoError(t, err)

		// Test query by business ID
		assessments, err := suite.storageService.GetRiskAssessmentsByBusiness(ctx, suite.testData.BusinessID)
		require.NoError(t, err)
		assert.NotEmpty(t, assessments)
		assert.Equal(t, suite.testData.BusinessID, assessments[0].BusinessID)

		// Test query by status
		completedAssessments, err := suite.storageService.GetRiskAssessmentsByStatus(ctx, AssessmentStatusCompleted)
		require.NoError(t, err)
		assert.NotEmpty(t, completedAssessments)
		assert.Equal(t, AssessmentStatusCompleted, completedAssessments[0].Status)

		// Test query by risk level
		mediumRiskAssessments, err := suite.storageService.GetRiskAssessmentsByLevel(ctx, RiskLevelMedium)
		require.NoError(t, err)
		assert.NotEmpty(t, mediumRiskAssessments)
		assert.Equal(t, RiskLevelMedium, mediumRiskAssessments[0].Level)
	})

	t.Run("Risk Factor Queries", func(t *testing.T) {
		ctx := context.Background()

		// Create test factors
		for _, factor := range suite.testData.RiskFactors {
			err := suite.storageService.SaveRiskFactor(ctx, &factor)
			require.NoError(t, err)
		}

		// Test query by business ID
		factors, err := suite.storageService.GetRiskFactorsByBusiness(ctx, suite.testData.BusinessID)
		require.NoError(t, err)
		assert.Len(t, factors, len(suite.testData.RiskFactors))

		// Test query by weight range
		highWeightFactors, err := suite.storageService.GetRiskFactorsByWeightRange(ctx, 0.3, 1.0)
		require.NoError(t, err)
		assert.NotEmpty(t, highWeightFactors)

		// Test query by value range
		highValueFactors, err := suite.storageService.GetRiskFactorsByValueRange(ctx, 70.0, 100.0)
		require.NoError(t, err)
		assert.NotEmpty(t, highValueFactors)
	})

	t.Run("Risk Score Queries", func(t *testing.T) {
		ctx := context.Background()

		// Create test scores
		for _, score := range suite.testData.RiskScores {
			err := suite.storageService.SaveRiskScore(ctx, &score)
			require.NoError(t, err)
		}

		// Test query by factor ID
		scores, err := suite.storageService.GetRiskScoresByFactor(ctx, "factor-1")
		require.NoError(t, err)
		assert.NotEmpty(t, scores)
		assert.Equal(t, "factor-1", scores[0].FactorID)

		// Test query by risk level
		mediumRiskScores, err := suite.storageService.GetRiskScoresByLevel(ctx, RiskLevelMedium)
		require.NoError(t, err)
		assert.NotEmpty(t, mediumRiskScores)

		// Test query by score range
		highScores, err := suite.storageService.GetRiskScoresByRange(ctx, 80.0, 100.0)
		require.NoError(t, err)
		assert.NotEmpty(t, highScores)
	})

	t.Run("Risk Alert Queries", func(t *testing.T) {
		ctx := context.Background()

		// Create test alerts
		for _, alert := range suite.testData.RiskAlerts {
			err := suite.storageService.SaveRiskAlert(ctx, &alert)
			require.NoError(t, err)
		}

		// Test query by business ID
		alerts, err := suite.storageService.GetRiskAlertsByBusiness(ctx, suite.testData.BusinessID)
		require.NoError(t, err)
		assert.Len(t, alerts, len(suite.testData.RiskAlerts))

		// Test query by alert level
		mediumAlerts, err := suite.storageService.GetRiskAlertsByLevel(ctx, RiskLevelMedium)
		require.NoError(t, err)
		assert.NotEmpty(t, mediumAlerts)

		// Test query by active status
		activeAlerts, err := suite.storageService.GetActiveRiskAlerts(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, activeAlerts)

		// Test query by alert type
		thresholdAlerts, err := suite.storageService.GetRiskAlertsByType(ctx, "threshold_exceeded")
		require.NoError(t, err)
		assert.NotEmpty(t, thresholdAlerts)
	})

	t.Run("Risk Trend Queries", func(t *testing.T) {
		ctx := context.Background()

		// Create test trends
		for _, trend := range suite.testData.RiskTrends {
			err := suite.storageService.SaveRiskTrend(ctx, &trend)
			require.NoError(t, err)
		}

		// Test query by business ID
		trends, err := suite.storageService.GetRiskTrendsByBusiness(ctx, suite.testData.BusinessID)
		require.NoError(t, err)
		assert.Len(t, trends, len(suite.testData.RiskTrends))

		// Test query by trend type
		improvingTrends, err := suite.storageService.GetRiskTrendsByType(ctx, "improving")
		require.NoError(t, err)
		assert.NotEmpty(t, improvingTrends)

		// Test query by factor ID
		factorTrends, err := suite.storageService.GetRiskTrendsByFactor(ctx, "factor-1")
		require.NoError(t, err)
		assert.NotEmpty(t, factorTrends)
	})

	t.Run("Risk History Queries", func(t *testing.T) {
		ctx := context.Background()

		// Create test history
		for _, history := range suite.testData.RiskHistory {
			err := suite.storageService.SaveRiskHistory(ctx, &history)
			require.NoError(t, err)
		}

		// Test query by business ID
		history, err := suite.storageService.GetRiskHistory(ctx, suite.testData.BusinessID)
		require.NoError(t, err)
		assert.Len(t, history, len(suite.testData.RiskHistory))

		// Test query by date range
		startDate := time.Now().Add(-48 * time.Hour)
		endDate := time.Now()
		recentHistory, err := suite.storageService.GetRiskHistoryByDateRange(ctx, suite.testData.BusinessID, startDate, endDate)
		require.NoError(t, err)
		assert.NotEmpty(t, recentHistory)

		// Test query by score range
		highScoreHistory, err := suite.storageService.GetRiskHistoryByScoreRange(ctx, suite.testData.BusinessID, 75.0, 100.0)
		require.NoError(t, err)
		assert.NotEmpty(t, highScoreHistory)
	})
}

// TestDatabaseTransactions tests database transaction operations
func TestDatabaseTransactions(t *testing.T) {
	suite := NewDatabaseIntegrationTestSuite(t)
	defer suite.cleanup()

	t.Run("Atomic Risk Assessment Creation", func(t *testing.T) {
		ctx := context.Background()

		// Create assessment with factors and scores in a transaction
		assessment := &RiskAssessment{
			ID:           suite.testData.AssessmentID,
			BusinessID:   suite.testData.BusinessID,
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		// Test successful transaction
		err := suite.storageService.CreateRiskAssessmentWithFactors(ctx, assessment, suite.testData.RiskFactors, suite.testData.RiskScores)
		require.NoError(t, err)

		// Verify all data was created
		retrievedAssessment, err := suite.storageService.GetRiskAssessment(ctx, suite.testData.AssessmentID)
		require.NoError(t, err)
		assert.Equal(t, assessment.ID, retrievedAssessment.ID)

		for _, factor := range suite.testData.RiskFactors {
			retrievedFactor, err := suite.storageService.GetRiskFactor(ctx, factor.ID)
			require.NoError(t, err)
			assert.Equal(t, factor.ID, retrievedFactor.ID)
		}

		for _, score := range suite.testData.RiskScores {
			retrievedScore, err := suite.storageService.GetRiskScore(ctx, score.ID)
			require.NoError(t, err)
			assert.Equal(t, score.ID, retrievedScore.ID)
		}
	})

	t.Run("Transaction Rollback on Error", func(t *testing.T) {
		ctx := context.Background()

		// Create assessment with invalid data to trigger rollback
		assessment := &RiskAssessment{
			ID:           "invalid-assessment",
			BusinessID:   suite.testData.BusinessID,
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		// Create invalid factors that should cause transaction to fail
		invalidFactors := []RiskFactorInput{
			{
				ID:          "invalid-factor",
				Name:        "", // Invalid: empty name
				Description: "Invalid factor",
				Weight:      0.3,
				Value:       85.0,
			},
		}

		// This should fail and rollback
		err := suite.storageService.CreateRiskAssessmentWithFactors(ctx, assessment, invalidFactors, suite.testData.RiskScores)
		assert.Error(t, err)

		// Verify no data was created (rollback worked)
		_, err = suite.storageService.GetRiskAssessment(ctx, "invalid-assessment")
		assert.Error(t, err)

		_, err = suite.storageService.GetRiskFactor(ctx, "invalid-factor")
		assert.Error(t, err)
	})

	t.Run("Concurrent Transaction Handling", func(t *testing.T) {
		ctx := context.Background()

		// Test concurrent updates to the same assessment
		assessment := &RiskAssessment{
			ID:           suite.testData.AssessmentID,
			BusinessID:   suite.testData.BusinessID,
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		err := suite.storageService.SaveRiskAssessment(ctx, assessment)
		require.NoError(t, err)

		// Simulate concurrent updates
		results := make(chan error, 3)

		go func() {
			assessment1, err := suite.storageService.GetRiskAssessment(ctx, suite.testData.AssessmentID)
			if err != nil {
				results <- err
				return
			}
			assessment1.OverallScore = 85.0
			assessment1.AssessedAt = time.Now()
			results <- suite.storageService.UpdateRiskAssessment(ctx, assessment1)
		}()

		go func() {
			assessment2, err := suite.storageService.GetRiskAssessment(ctx, suite.testData.AssessmentID)
			if err != nil {
				results <- err
				return
			}
			assessment2.OverallLevel = RiskLevelLow
			assessment2.AssessedAt = time.Now()
			results <- suite.storageService.UpdateRiskAssessment(ctx, assessment2)
		}()

		go func() {
			assessment3, err := suite.storageService.GetRiskAssessment(ctx, suite.testData.AssessmentID)
			if err != nil {
				results <- err
				return
			}
			assessment3.OverallLevel = RiskLevelHigh
			assessment3.AssessedAt = time.Now()
			results <- suite.storageService.UpdateRiskAssessment(ctx, assessment3)
		}()

		// Wait for all updates to complete
		for i := 0; i < 3; i++ {
			err := <-results
			// At least one should succeed, others might fail due to concurrency
			if i == 0 {
				assert.NoError(t, err)
			}
		}

		// Verify final state
		finalAssessment, err := suite.storageService.GetRiskAssessment(ctx, suite.testData.AssessmentID)
		require.NoError(t, err)
		assert.NotNil(t, finalAssessment)
	})
}

// TestDatabasePerformance tests database performance and optimization
func TestDatabasePerformance(t *testing.T) {
	suite := NewDatabaseIntegrationTestSuite(t)
	defer suite.cleanup()

	t.Run("Bulk Insert Performance", func(t *testing.T) {
		ctx := context.Background()

		// Create large number of risk factors
		numFactors := 1000
		factors := make([]RiskFactorInput, numFactors)

		for i := 0; i < numFactors; i++ {
			factors[i] = RiskFactorInput{
				ID:          fmt.Sprintf("bulk-factor-%d", i),
				Name:        fmt.Sprintf("Bulk Factor %d", i),
				Description: fmt.Sprintf("Bulk factor description %d", i),
				Weight:      0.1 + float64(i%10)*0.01,
				Value:       50.0 + float64(i%50),
			}
		}

		start := time.Now()
		err := suite.storageService.BulkInsertRiskFactors(ctx, factors)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.True(t, duration < 5*time.Second, "Bulk insert should complete within 5 seconds")

		// Verify all factors were inserted
		for i := 0; i < numFactors; i++ {
			factor, err := suite.storageService.GetRiskFactor(ctx, fmt.Sprintf("bulk-factor-%d", i))
			require.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("bulk-factor-%d", i), factor.ID)
		}
	})

	t.Run("Query Performance", func(t *testing.T) {
		ctx := context.Background()

		// Create test data
		for _, factor := range suite.testData.RiskFactors {
			err := suite.storageService.SaveRiskFactor(ctx, &factor)
			require.NoError(t, err)
		}

		// Test query performance
		start := time.Now()
		factors, err := suite.storageService.GetRiskFactorsByBusiness(ctx, suite.testData.BusinessID)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.NotEmpty(t, factors)
		assert.True(t, duration < 100*time.Millisecond, "Query should complete within 100ms")
	})

	t.Run("Index Performance", func(t *testing.T) {
		ctx := context.Background()

		// Create test assessments with different business IDs
		businessIDs := []string{"business-1", "business-2", "business-3", "business-4", "business-5"}

		for i, businessID := range businessIDs {
			assessment := &RiskAssessment{
				ID:           fmt.Sprintf("assessment-%d", i),
				BusinessID:   businessID,
				BusinessName: "Test Business",
				OverallScore: 70.0 + float64(i*5),
				Level:        RiskLevelMedium,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			err := suite.storageService.SaveRiskAssessment(ctx, assessment)
			require.NoError(t, err)
		}

		// Test indexed query performance
		start := time.Now()
		assessments, err := suite.storageService.GetRiskAssessmentsByBusiness(ctx, "business-3")
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Len(t, assessments, 1)
		assert.Equal(t, "business-3", assessments[0].BusinessID)
		assert.True(t, duration < 50*time.Millisecond, "Indexed query should complete within 50ms")
	})
}

// TestDatabaseConstraints tests database constraints and validation
func TestDatabaseConstraints(t *testing.T) {
	suite := NewDatabaseIntegrationTestSuite(t)
	defer suite.cleanup()

	t.Run("Unique Constraint Violations", func(t *testing.T) {
		ctx := context.Background()

		// Create assessment with duplicate ID
		assessment1 := &RiskAssessment{
			ID:           suite.testData.AssessmentID,
			BusinessID:   suite.testData.BusinessID,
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		err := suite.storageService.SaveRiskAssessment(ctx, assessment1)
		require.NoError(t, err)

		// Try to create another assessment with the same ID
		assessment2 := &RiskAssessment{
			ID:           suite.testData.AssessmentID, // Same ID
			BusinessID:   "different-business",
			BusinessName: "Test Business",
			OverallScore: 75.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		err = suite.storageService.SaveRiskAssessment(ctx, assessment2)
		assert.Error(t, err) // Should fail due to unique constraint
	})

	t.Run("Foreign Key Constraint Violations", func(t *testing.T) {
		ctx := context.Background()

		// Try to create a risk score with non-existent factor ID
		score := RiskScore{
			ID:          "invalid-score",
			FactorID:    "non-existent-factor", // Invalid foreign key
			Score:       85.0,
			Level:       RiskLevelLow,
			Confidence:  0.95,
			LastUpdated: time.Now(),
		}

		err := suite.storageService.SaveRiskScore(ctx, &score)
		assert.Error(t, err) // Should fail due to foreign key constraint
	})

	t.Run("Check Constraint Violations", func(t *testing.T) {
		ctx := context.Background()

		// Try to create assessment with invalid score
		assessment := &RiskAssessment{
			ID:           "invalid-assessment",
			BusinessID:   suite.testData.BusinessID,
			BusinessName: "Test Business",
			OverallScore: 150.0, // Invalid: score should be 0-100
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		err := suite.storageService.SaveRiskAssessment(ctx, assessment)
		assert.Error(t, err) // Should fail due to check constraint
	})

	t.Run("Not Null Constraint Violations", func(t *testing.T) {
		ctx := context.Background()

		// Try to create factor with null name
		factor := RiskFactorInput{
			ID:          "invalid-factor",
			Name:        "", // Invalid: name cannot be null
			Description: "Invalid factor",
			Weight:      0.3,
			Value:       85.0,
		}

		err := suite.storageService.SaveRiskFactor(ctx, &factor)
		assert.Error(t, err) // Should fail due to not null constraint
	})
}

// TestDatabaseBackupRestore tests database backup and restore operations
func TestDatabaseBackupRestore(t *testing.T) {
	suite := NewDatabaseIntegrationTestSuite(t)
	defer suite.cleanup()

	t.Run("Database Backup", func(t *testing.T) {
		ctx := context.Background()

		// Create test data
		assessment := &RiskAssessment{
			ID:           suite.testData.AssessmentID,
			BusinessID:   suite.testData.BusinessID,
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		err := suite.storageService.SaveRiskAssessment(ctx, assessment)
		require.NoError(t, err)

		// Create backup
		backupRequest := &BackupRequest{
			BusinessID:  suite.testData.BusinessID,
			BackupType:  BackupTypeBusiness,
			IncludeData: []string{"assessments"},
			Metadata:    map[string]interface{}{"test": "database"},
		}

		backupResult, err := suite.backupSvc.CreateBackup(ctx, backupRequest)
		require.NoError(t, err)
		assert.NotEmpty(t, backupResult.BackupID)
		assert.Equal(t, BackupStatusCompleted, backupResult.Status)
	})

	t.Run("Database Restore", func(t *testing.T) {
		ctx := context.Background()

		// First create a backup
		backupRequest := &BackupRequest{
			BusinessID:  suite.testData.BusinessID,
			BackupType:  BackupTypeBusiness,
			IncludeData: []string{"assessments"},
			Metadata:    map[string]interface{}{"test": "database"},
		}

		backupResult, err := suite.backupSvc.CreateBackup(ctx, backupRequest)
		require.NoError(t, err)

		// Now restore from backup
		restoreRequest := &RestoreRequest{
			BackupID:    backupResult.BackupID,
			BusinessID:  suite.testData.BusinessID,
			RestoreType: RestoreTypeBusiness,
		}

		restoreResult, err := suite.backupSvc.RestoreBackup(ctx, restoreRequest)
		require.NoError(t, err)
		assert.NotEmpty(t, restoreResult.RestoreID)
		assert.Equal(t, RestoreStatusCompleted, restoreResult.Status)
	})
}

// cleanup performs cleanup operations for the test suite
func (suite *DatabaseIntegrationTestSuite) cleanup() {
	for _, cleanupFunc := range suite.cleanupFunctions {
		if err := cleanupFunc(); err != nil {
			suite.logger.Error("Cleanup function failed", zap.Error(err))
		}
	}
}
