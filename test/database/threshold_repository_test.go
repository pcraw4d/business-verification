package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kyb-platform/internal/database"
	"kyb-platform/internal/risk"
)

// getTestDatabaseURL returns the test database URL from environment
func getTestDatabaseURLForThreshold(t *testing.T) string {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://kyb_test:kyb_test_password@localhost:5433/kyb_test?sslmode=disable"
		t.Logf("Using default DATABASE_URL: %s", dbURL)
	}
	return dbURL
}

// TestThresholdRepositoryCRUD tests all CRUD operations for threshold repository
func TestThresholdRepositoryCRUD(t *testing.T) {
	dbURL := getTestDatabaseURLForThreshold(t)

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Ensure migration is applied
	applyThresholdMigration(t, db)

	ctx := context.Background()
	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)
	repo := database.NewThresholdRepository(db, logger)

	// Clean up any existing test data
	cleanupTestThresholds(t, db, ctx)

	t.Run("CreateThreshold", func(t *testing.T) {
		config := &risk.ThresholdConfig{
			ID:          uuid.New().String(),
			Name:        "Test Financial Threshold",
			Description: "Test threshold for financial risk category",
			Category:    risk.RiskCategoryFinancial,
			RiskLevels: map[risk.RiskLevel]float64{
				risk.RiskLevelLow:      25.0,
				risk.RiskLevelMedium:   50.0,
				risk.RiskLevelHigh:     75.0,
				risk.RiskLevelCritical: 90.0,
			},
			IsDefault:      false,
			IsActive:       true,
			Priority:       5,
			CreatedBy:      "test-user",
			LastModifiedBy: "test-user",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// Convert to adapter format
		adapter := risk.NewThresholdRepositoryAdapter(repo)
		err := adapter.CreateThreshold(ctx, config)
		require.NoError(t, err, "Should create threshold successfully")

		// Verify it was created
		retrieved, err := adapter.GetThreshold(ctx, config.ID)
		require.NoError(t, err, "Should retrieve created threshold")
		assert.Equal(t, config.ID, retrieved.ID)
		assert.Equal(t, config.Name, retrieved.Name)
		assert.Equal(t, config.Category, retrieved.Category)
		assert.Equal(t, len(config.RiskLevels), len(retrieved.RiskLevels))
	})

	t.Run("GetThreshold", func(t *testing.T) {
		// Create a test threshold first
		config := &risk.ThresholdConfig{
			ID:          uuid.New().String(),
			Name:        "Test Get Threshold",
			Description: "Test threshold for get operation",
			Category:    risk.RiskCategoryOperational,
			RiskLevels: map[risk.RiskLevel]float64{
				risk.RiskLevelLow:      20.0,
				risk.RiskLevelMedium:   45.0,
				risk.RiskLevelHigh:     70.0,
				risk.RiskLevelCritical: 85.0,
			},
			IsDefault:      false,
			IsActive:       true,
			Priority:       10,
			CreatedBy:      "test-user",
			LastModifiedBy: "test-user",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		adapter := risk.NewThresholdRepositoryAdapter(repo)
		err := adapter.CreateThreshold(ctx, config)
		require.NoError(t, err)

		// Test GetThreshold
		retrieved, err := adapter.GetThreshold(ctx, config.ID)
		require.NoError(t, err, "Should retrieve threshold")
		assert.Equal(t, config.ID, retrieved.ID)
		assert.Equal(t, config.Name, retrieved.Name)
		assert.Equal(t, config.Category, retrieved.Category)
		assert.Equal(t, config.IsActive, retrieved.IsActive)
		assert.Equal(t, config.Priority, retrieved.Priority)

		// Verify risk levels
		for level, value := range config.RiskLevels {
			assert.Equal(t, value, retrieved.RiskLevels[level], "Risk level %s should match", level)
		}

		// Test GetThreshold with non-existent ID
		_, err = adapter.GetThreshold(ctx, "non-existent-id")
		assert.Error(t, err, "Should return error for non-existent threshold")
	})

	t.Run("UpdateThreshold", func(t *testing.T) {
		// Create a test threshold first
		config := &risk.ThresholdConfig{
			ID:          uuid.New().String(),
			Name:        "Test Update Threshold",
			Description: "Original description",
			Category:    risk.RiskCategoryRegulatory,
			RiskLevels: map[risk.RiskLevel]float64{
				risk.RiskLevelLow:      30.0,
				risk.RiskLevelMedium:   55.0,
				risk.RiskLevelHigh:     80.0,
				risk.RiskLevelCritical: 95.0,
			},
			IsDefault:      false,
			IsActive:       true,
			Priority:       15,
			CreatedBy:      "test-user",
			LastModifiedBy: "test-user",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		adapter := risk.NewThresholdRepositoryAdapter(repo)
		manager := risk.NewThresholdManagerWithRepository(adapter)

		// Register the threshold
		err := manager.RegisterConfig(config)
		require.NoError(t, err)

		// Update the threshold using manager
		updates := map[string]interface{}{
			"name":             "Updated Test Threshold",
			"description":      "Updated description",
			"is_active":        false,
			"priority":         20,
			"last_modified_by": "test-user-updated",
		}
		err = manager.UpdateConfig(config.ID, updates)
		require.NoError(t, err, "Should update threshold successfully")

		// Verify the update in memory
		retrieved, exists := manager.GetConfig(config.ID)
		require.True(t, exists, "Should exist in memory")
		assert.Equal(t, "Updated Test Threshold", retrieved.Name)
		assert.Equal(t, "Updated description", retrieved.Description)
		assert.Equal(t, false, retrieved.IsActive)
		assert.Equal(t, 20, retrieved.Priority)
		assert.Equal(t, "test-user-updated", retrieved.LastModifiedBy)

		// Verify the update in database
		dbRetrieved, err := adapter.GetThreshold(ctx, config.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Test Threshold", dbRetrieved.Name)
		assert.Equal(t, "Updated description", dbRetrieved.Description)
		assert.Equal(t, false, dbRetrieved.IsActive)
		assert.Equal(t, 20, dbRetrieved.Priority)
	})

	t.Run("DeleteThreshold", func(t *testing.T) {
		// Create a test threshold first
		config := &risk.ThresholdConfig{
			ID:          uuid.New().String(),
			Name:        "Test Delete Threshold",
			Description: "Test threshold for delete operation",
			Category:    risk.RiskCategoryReputational,
			RiskLevels: map[risk.RiskLevel]float64{
				risk.RiskLevelLow:      20.0,
				risk.RiskLevelMedium:   45.0,
				risk.RiskLevelHigh:     75.0,
				risk.RiskLevelCritical: 90.0,
			},
			IsDefault:      false,
			IsActive:       true,
			Priority:       5,
			CreatedBy:      "test-user",
			LastModifiedBy: "test-user",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		adapter := risk.NewThresholdRepositoryAdapter(repo)
		manager := risk.NewThresholdManagerWithRepository(adapter)

		// Register the threshold
		err := manager.RegisterConfig(config)
		require.NoError(t, err)

		// Verify it exists
		_, exists := manager.GetConfig(config.ID)
		assert.True(t, exists, "Threshold should exist before deletion")

		// Delete the threshold using manager
		err = manager.DeleteConfig(config.ID)
		require.NoError(t, err, "Should delete threshold successfully")

		// Verify it's deleted from memory
		_, exists = manager.GetConfig(config.ID)
		assert.False(t, exists, "Should not exist in memory after deletion")

		// Verify it's deleted from database
		_, err = adapter.GetThreshold(ctx, config.ID)
		assert.Error(t, err, "Should return error for deleted threshold")

		// Test DeleteConfig with non-existent ID
		err = manager.DeleteConfig("non-existent-id")
		assert.Error(t, err, "Should return error for non-existent threshold")
	})

	t.Run("ListThresholds", func(t *testing.T) {
		// Create multiple test thresholds
		categories := []risk.RiskCategory{
			risk.RiskCategoryFinancial,
			risk.RiskCategoryOperational,
			risk.RiskCategoryRegulatory,
		}

		adapter := risk.NewThresholdRepositoryAdapter(repo)
		for i, category := range categories {
			config := &risk.ThresholdConfig{
				ID:          uuid.New().String(),
				Name:        "Test Threshold " + string(category),
				Description: "Test threshold for list operation",
				Category:    category,
				RiskLevels: map[risk.RiskLevel]float64{
					risk.RiskLevelLow:      25.0 + float64(i),
					risk.RiskLevelMedium:   50.0 + float64(i),
					risk.RiskLevelHigh:     75.0 + float64(i),
					risk.RiskLevelCritical: 90.0 + float64(i),
				},
				IsDefault:      false,
				IsActive:       true,
				Priority:       5 + i,
				CreatedBy:      "test-user",
				LastModifiedBy: "test-user",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			err := adapter.CreateThreshold(ctx, config)
			require.NoError(t, err)
		}

		// Test ListThresholds without filters
		allConfigs, err := adapter.ListThresholds(ctx, nil, nil, false)
		require.NoError(t, err, "Should list all thresholds")
		assert.GreaterOrEqual(t, len(allConfigs), 3, "Should have at least 3 test thresholds")

		// Test ListThresholds with category filter
		financialCategory := risk.RiskCategoryFinancial
		financialConfigs, err := adapter.ListThresholds(ctx, &financialCategory, nil, false)
		require.NoError(t, err, "Should list financial thresholds")
		assert.GreaterOrEqual(t, len(financialConfigs), 1, "Should have at least 1 financial threshold")
		for _, cfg := range financialConfigs {
			assert.Equal(t, risk.RiskCategoryFinancial, cfg.Category)
		}

		// Test ListThresholds with activeOnly filter
		activeConfigs, err := adapter.ListThresholds(ctx, nil, nil, true)
		require.NoError(t, err, "Should list active thresholds")
		for _, cfg := range activeConfigs {
			assert.True(t, cfg.IsActive, "All listed thresholds should be active")
		}
	})

	t.Run("LoadAllThresholds", func(t *testing.T) {
		adapter := risk.NewThresholdRepositoryAdapter(repo)
		configs, err := adapter.LoadAllThresholds(ctx)
		require.NoError(t, err, "Should load all thresholds")
		assert.GreaterOrEqual(t, len(configs), 0, "Should return at least 0 thresholds")

		// Verify all returned thresholds are active
		for _, cfg := range configs {
			assert.True(t, cfg.IsActive, "All loaded thresholds should be active")
		}
	})

	// Clean up test data
	cleanupTestThresholds(t, db, ctx)
}

// TestThresholdManagerWithDatabase tests ThresholdManager with database persistence
func TestThresholdManagerWithDatabase(t *testing.T) {
	dbURL := getTestDatabaseURLForThreshold(t)

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Ensure migration is applied
	applyThresholdMigration(t, db)

	ctx := context.Background()
	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)
	repo := database.NewThresholdRepository(db, logger)
	adapter := risk.NewThresholdRepositoryAdapter(repo)

	// Clean up any existing test data
	cleanupTestThresholds(t, db, ctx)

	t.Run("RegisterConfig with persistence", func(t *testing.T) {
		manager := risk.NewThresholdManagerWithRepository(adapter)

		config := &risk.ThresholdConfig{
			ID:          uuid.New().String(),
			Name:        "Test Manager Threshold",
			Description: "Test threshold via manager",
			Category:    risk.RiskCategoryFinancial,
			RiskLevels: map[risk.RiskLevel]float64{
				risk.RiskLevelLow:      25.0,
				risk.RiskLevelMedium:   50.0,
				risk.RiskLevelHigh:     75.0,
				risk.RiskLevelCritical: 90.0,
			},
			IsDefault:      false,
			IsActive:       true,
			Priority:       5,
			CreatedBy:      "test-user",
			LastModifiedBy: "test-user",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := manager.RegisterConfig(config)
		require.NoError(t, err, "Should register threshold via manager")

		// Verify it's in memory
		retrieved, exists := manager.GetConfig(config.ID)
		assert.True(t, exists, "Should exist in memory")
		assert.Equal(t, config.ID, retrieved.ID)

		// Verify it's in database
		dbConfig, err := adapter.GetThreshold(ctx, config.ID)
		require.NoError(t, err, "Should exist in database")
		assert.Equal(t, config.ID, dbConfig.ID)
	})

	t.Run("LoadFromDatabase", func(t *testing.T) {
		// Create threshold directly in database
		testID := uuid.New().String()
		config := &risk.ThresholdConfig{
			ID:          testID,
			Name:        "Test Load From DB",
			Description: "Test loading from database",
			Category:    risk.RiskCategoryOperational,
			RiskLevels: map[risk.RiskLevel]float64{
				risk.RiskLevelLow:      20.0,
				risk.RiskLevelMedium:   45.0,
				risk.RiskLevelHigh:     70.0,
				risk.RiskLevelCritical: 85.0,
			},
			IsDefault:      false,
			IsActive:       true,
			Priority:       10,
			CreatedBy:      "test-user",
			LastModifiedBy: "test-user",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := adapter.CreateThreshold(ctx, config)
		require.NoError(t, err)

		// Create new manager and load from database
		manager := risk.NewThresholdManagerWithRepository(adapter)
		err = manager.LoadFromDatabase(ctx)
		require.NoError(t, err, "Should load thresholds from database")

		// Verify it's loaded into memory
		retrieved, exists := manager.GetConfig(testID)
		assert.True(t, exists, "Should be loaded into memory")
		assert.Equal(t, testID, retrieved.ID)
		assert.Equal(t, config.Name, retrieved.Name)
	})

	t.Run("SyncToDatabase", func(t *testing.T) {
		manager := risk.NewThresholdManagerWithRepository(adapter)

		// Create threshold and register it
		config := &risk.ThresholdConfig{
			ID:          uuid.New().String(),
			Name:        "Test Sync To DB",
			Description: "Test syncing to database",
			Category:    risk.RiskCategoryRegulatory,
			RiskLevels: map[risk.RiskLevel]float64{
				risk.RiskLevelLow:      30.0,
				risk.RiskLevelMedium:   55.0,
				risk.RiskLevelHigh:     80.0,
				risk.RiskLevelCritical: 95.0,
			},
			IsDefault:      false,
			IsActive:       true,
			Priority:       15,
			CreatedBy:      "test-user",
			LastModifiedBy: "test-user",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// Register with manager (this will persist to DB)
		err := manager.RegisterConfig(config)
		require.NoError(t, err)

		// Verify it's in database
		dbConfig, err := adapter.GetThreshold(ctx, config.ID)
		require.NoError(t, err, "Should exist in database after registration")
		assert.Equal(t, config.ID, dbConfig.ID)

		// Test SyncToDatabase (should be idempotent)
		err = manager.SyncToDatabase(ctx)
		require.NoError(t, err, "Should sync to database successfully")
	})

	// Clean up test data
	cleanupTestThresholds(t, db, ctx)
}

// TestThresholdExportImport tests export/import functionality with database
func TestThresholdExportImport(t *testing.T) {
	dbURL := getTestDatabaseURLForThreshold(t)

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Ensure migration is applied
	applyThresholdMigration(t, db)

	ctx := context.Background()
	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)
	repo := database.NewThresholdRepository(db, logger)
	adapter := risk.NewThresholdRepositoryAdapter(repo)
	manager := risk.NewThresholdManagerWithRepository(adapter)
	service := risk.NewThresholdConfigService(manager)

	// Clean up any existing test data
	cleanupTestThresholds(t, db, ctx)

	t.Run("ExportThresholds from database", func(t *testing.T) {
		// Create test thresholds
		exportID1 := uuid.New().String()
		exportID2 := uuid.New().String()
		configs := []*risk.ThresholdConfig{
			{
				ID:          exportID1,
				Name:        "Test Export Threshold 1",
				Description: "First test threshold for export",
				Category:    risk.RiskCategoryFinancial,
				RiskLevels: map[risk.RiskLevel]float64{
					risk.RiskLevelLow:      25.0,
					risk.RiskLevelMedium:   50.0,
					risk.RiskLevelHigh:     75.0,
					risk.RiskLevelCritical: 90.0,
				},
				IsDefault:      false,
				IsActive:       true,
				Priority:       5,
				CreatedBy:      "test-user",
				LastModifiedBy: "test-user",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			{
				ID:          exportID2,
				Name:        "Test Export Threshold 2",
				Description: "Second test threshold for export",
				Category:    risk.RiskCategoryOperational,
				RiskLevels: map[risk.RiskLevel]float64{
					risk.RiskLevelLow:      20.0,
					risk.RiskLevelMedium:   45.0,
					risk.RiskLevelHigh:     70.0,
					risk.RiskLevelCritical: 85.0,
				},
				IsDefault:      false,
				IsActive:       true,
				Priority:       10,
				CreatedBy:      "test-user",
				LastModifiedBy: "test-user",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
		}

		// Register thresholds
		for _, config := range configs {
			err := manager.RegisterConfig(config)
			require.NoError(t, err)
		}

		// Export thresholds
		exportData, err := service.ExportThresholds()
		require.NoError(t, err, "Should export thresholds successfully")
		assert.NotEmpty(t, exportData, "Export data should not be empty")

		// Verify JSON is valid
		var exportedConfigs []*risk.ThresholdConfig
		err = json.Unmarshal(exportData, &exportedConfigs)
		require.NoError(t, err, "Exported JSON should be valid")
		assert.GreaterOrEqual(t, len(exportedConfigs), 2, "Should export at least 2 thresholds")

		// Verify exported configs contain our test thresholds
		exportedIDs := make(map[string]bool)
		for _, cfg := range exportedConfigs {
			exportedIDs[cfg.ID] = true
		}
		assert.True(t, exportedIDs[exportID1], "Should export test-export-1")
		assert.True(t, exportedIDs[exportID2], "Should export test-export-2")
	})

	t.Run("ImportThresholds to database", func(t *testing.T) {
		// Create export data
		importID := uuid.New().String()
		exportConfigs := []*risk.ThresholdConfig{
			{
				ID:          importID,
				Name:        "Test Import Threshold 1",
				Description: "First test threshold for import",
				Category:    risk.RiskCategoryRegulatory,
				RiskLevels: map[risk.RiskLevel]float64{
					risk.RiskLevelLow:      30.0,
					risk.RiskLevelMedium:   55.0,
					risk.RiskLevelHigh:     80.0,
					risk.RiskLevelCritical: 95.0,
				},
				IsDefault:      false,
				IsActive:       true,
				Priority:       15,
				CreatedBy:      "test-user",
				LastModifiedBy: "test-user",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
		}

		exportJSON, err := json.Marshal(exportConfigs)
		require.NoError(t, err)

		// Create a new manager for import test
		newManager := risk.NewThresholdManagerWithRepository(adapter)
		newService := risk.NewThresholdConfigService(newManager)

		// Import thresholds
		err = newService.ImportThresholds(exportJSON)
		require.NoError(t, err, "Should import thresholds successfully")

		// Verify imported threshold exists in memory
		retrieved, exists := newManager.GetConfig(importID)
		assert.True(t, exists, "Should exist in memory after import")
		assert.Equal(t, importID, retrieved.ID)
		assert.Equal(t, "Test Import Threshold 1", retrieved.Name)

		// Verify imported threshold exists in database
		dbRetrieved, err := adapter.GetThreshold(ctx, importID)
		require.NoError(t, err, "Should exist in database after import")
		assert.Equal(t, importID, dbRetrieved.ID)
		assert.Equal(t, "Test Import Threshold 1", dbRetrieved.Name)
	})

	t.Run("Export and Import roundtrip", func(t *testing.T) {
		// Create a threshold
		roundtripID := uuid.New().String()
		originalConfig := &risk.ThresholdConfig{
			ID:          roundtripID,
			Name:        "Test Roundtrip Threshold",
			Description: "Test threshold for export/import roundtrip",
			Category:    risk.RiskCategoryCybersecurity,
			RiskLevels: map[risk.RiskLevel]float64{
				risk.RiskLevelLow:      15.0,
				risk.RiskLevelMedium:   40.0,
				risk.RiskLevelHigh:     75.0,
				risk.RiskLevelCritical: 90.0,
			},
			IsDefault:      false,
			IsActive:       true,
			Priority:       20,
			Metadata:       map[string]interface{}{"test": "metadata"},
			CreatedBy:      "test-user",
			LastModifiedBy: "test-user",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := manager.RegisterConfig(originalConfig)
		require.NoError(t, err)

		// Export
		exportData, err := service.ExportThresholds()
		require.NoError(t, err)

		// Create new manager and import
		newManager := risk.NewThresholdManagerWithRepository(adapter)
		newService := risk.NewThresholdConfigService(newManager)

		err = newService.ImportThresholds(exportData)
		require.NoError(t, err)

		// Verify roundtrip
		retrieved, exists := newManager.GetConfig(roundtripID)
		assert.True(t, exists, "Should exist after roundtrip")
		assert.Equal(t, originalConfig.Name, retrieved.Name)
		assert.Equal(t, originalConfig.Category, retrieved.Category)
		assert.Equal(t, originalConfig.Priority, retrieved.Priority)
		assert.Equal(t, len(originalConfig.RiskLevels), len(retrieved.RiskLevels))
		for level, value := range originalConfig.RiskLevels {
			assert.Equal(t, value, retrieved.RiskLevels[level], "Risk level %s should match", level)
		}
	})

	// Clean up test data
	cleanupTestThresholds(t, db, ctx)
}

// applyThresholdMigration applies the threshold migration if not already applied
func applyThresholdMigration(t *testing.T, db *sql.DB) {
	// Try multiple possible paths for the migration file
	migrationPaths := []string{
		"internal/database/migrations/012_create_risk_thresholds_table.sql",
		"../internal/database/migrations/012_create_risk_thresholds_table.sql",
		"../../internal/database/migrations/012_create_risk_thresholds_table.sql",
	}

	var migrationSQL []byte
	var err error
	var migrationPath string

	for _, path := range migrationPaths {
		migrationSQL, err = os.ReadFile(path)
		if err == nil {
			migrationPath = path
			break
		}
	}

	if err != nil {
		// If migration file not found, create the table directly
		t.Logf("Note: Could not read migration file, creating table directly: %v", err)
		// Drop table if it exists to ensure clean state
		_, _ = db.Exec("DROP TABLE IF EXISTS risk_thresholds CASCADE")
		
		createTableSQL := `
			CREATE TABLE risk_thresholds (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				name VARCHAR(255) NOT NULL,
				description TEXT,
				category VARCHAR(50) NOT NULL,
				industry_code VARCHAR(20),
				business_type VARCHAR(50),
				risk_levels JSONB NOT NULL,
				is_default BOOLEAN NOT NULL DEFAULT FALSE,
				is_active BOOLEAN NOT NULL DEFAULT TRUE,
				priority INTEGER NOT NULL DEFAULT 0,
				metadata JSONB,
				created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
				created_by VARCHAR(255) NOT NULL DEFAULT 'system',
				last_modified_by VARCHAR(255) NOT NULL DEFAULT 'system',
				CONSTRAINT unique_active_threshold UNIQUE (category, industry_code, business_type, is_active) WHERE is_active IS TRUE,
				CONSTRAINT risk_levels_not_empty CHECK (risk_levels::text != '{}' AND risk_levels::text != 'null')
			);
			CREATE INDEX IF NOT EXISTS idx_risk_thresholds_category ON risk_thresholds (category);
			CREATE INDEX IF NOT EXISTS idx_risk_thresholds_industry_code ON risk_thresholds (industry_code);
			CREATE INDEX IF NOT EXISTS idx_risk_thresholds_is_active ON risk_thresholds (is_active);
			CREATE INDEX IF NOT EXISTS idx_risk_thresholds_priority ON risk_thresholds (priority);
		`
		_, err = db.Exec(createTableSQL)
		if err != nil {
			t.Logf("Note: Could not create risk_thresholds table: %v", err)
		}
		return
	}
	
	// Drop and recreate table to ensure constraint is correct
	_, _ = db.Exec("DROP TABLE IF EXISTS risk_thresholds CASCADE")

	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		// Migration might already be applied, which is fine
		t.Logf("Note: Migration may already be applied: %v", err)
	} else {
		t.Logf("Applied migration from: %s", migrationPath)
	}
}

// cleanupTestThresholds removes test thresholds
func cleanupTestThresholds(t *testing.T, db *sql.DB, ctx context.Context) {
	_, err := db.ExecContext(ctx, `
		DELETE FROM risk_thresholds 
		WHERE name LIKE 'Test%' OR description LIKE '%test%'
	`)
	if err != nil {
		t.Logf("Note: Could not clean up test thresholds: %v", err)
	}
}
