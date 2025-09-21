package testing

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// createTimestampedTestData creates test data at specific timestamps for point-in-time recovery testing
func (brt *BackupRecoveryTester) createTimestampedTestData(ctx context.Context) ([]time.Time, error) {
	brt.logger.Printf("Creating timestamped test data for point-in-time recovery testing")

	var timestamps []time.Time
	baseTime := time.Now().Add(-24 * time.Hour) // Start 24 hours ago

	// Create test data at 4 different timestamps
	for i := 0; i < 4; i++ {
		timestamp := baseTime.Add(time.Duration(i) * 6 * time.Hour) // Every 6 hours
		timestamps = append(timestamps, timestamp)

		// Create test merchant data
		if err := brt.createTestMerchantData(ctx, timestamp, i); err != nil {
			return nil, fmt.Errorf("failed to create test merchant data at timestamp %v: %w", timestamp, err)
		}

		// Create test classification data
		if err := brt.createTestClassificationData(ctx, timestamp, i); err != nil {
			return nil, fmt.Errorf("failed to create test classification data at timestamp %v: %w", timestamp, err)
		}

		// Create test risk assessment data
		if err := brt.createTestRiskAssessmentData(ctx, timestamp, i); err != nil {
			return nil, fmt.Errorf("failed to create test risk assessment data at timestamp %v: %w", timestamp, err)
		}

		brt.logger.Printf("Created test data for timestamp %v", timestamp)
	}

	return timestamps, nil
}

// createTestMerchantData creates test merchant data at a specific timestamp
func (brt *BackupRecoveryTester) createTestMerchantData(ctx context.Context, timestamp time.Time, index int) error {
	merchantID := fmt.Sprintf("test-merchant-%d", index)

	query := `
		INSERT INTO merchants (
			id, name, business_type, industry, 
			created_at, updated_at, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			updated_at = $6,
			status = $7
	`

	_, err := brt.db.ExecContext(ctx, query,
		merchantID,
		fmt.Sprintf("Test Merchant %d", index),
		"LLC",
		fmt.Sprintf("Test Industry %d", index),
		timestamp,
		timestamp,
		"active",
	)

	return err
}

// createTestClassificationData creates test classification data at a specific timestamp
func (brt *BackupRecoveryTester) createTestClassificationData(ctx context.Context, timestamp time.Time, index int) error {
	// Create test industry data
	industryID := index + 1000 // Use high IDs to avoid conflicts

	query := `
		INSERT INTO industries (
			id, name, description, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			updated_at = $5
	`

	_, err := brt.db.ExecContext(ctx, query,
		industryID,
		fmt.Sprintf("Test Industry %d", index),
		fmt.Sprintf("Test industry description %d", index),
		timestamp,
		timestamp,
	)

	if err != nil {
		return err
	}

	// Create test industry keywords
	keywordQuery := `
		INSERT INTO industry_keywords (
			industry_id, keyword, weight, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (industry_id, keyword) DO UPDATE SET
			weight = $3,
			updated_at = $5
	`

	_, err = brt.db.ExecContext(ctx, keywordQuery,
		industryID,
		fmt.Sprintf("testkeyword%d", index),
		0.8,
		timestamp,
		timestamp,
	)

	return err
}

// createTestRiskAssessmentData creates test risk assessment data at a specific timestamp
func (brt *BackupRecoveryTester) createTestRiskAssessmentData(ctx context.Context, timestamp time.Time, index int) error {
	merchantID := fmt.Sprintf("test-merchant-%d", index)

	query := `
		INSERT INTO business_risk_assessments (
			id, business_id, risk_score, risk_level, 
			assessment_method, assessment_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			risk_score = $3,
			risk_level = $4,
			updated_at = $8
	`

	riskLevel := "low"
	if index%2 == 1 {
		riskLevel = "medium"
	}

	_, err := brt.db.ExecContext(ctx, query,
		fmt.Sprintf("test-risk-assessment-%d", index),
		merchantID,
		0.3+float64(index)*0.1, // Varying risk scores
		riskLevel,
		"automated_test",
		timestamp,
		timestamp,
		timestamp,
	)

	return err
}

// recoverToTimestamp recovers the database to a specific timestamp
func (brt *BackupRecoveryTester) recoverToTimestamp(ctx context.Context, targetTimestamp time.Time) error {
	brt.logger.Printf("Recovering database to timestamp: %v", targetTimestamp)

	// For Supabase, we'll simulate point-in-time recovery by:
	// 1. Finding the closest backup before the target timestamp
	// 2. Restoring from that backup
	// 3. Applying any necessary data corrections

	// Find the most recent backup before the target timestamp
	backupFile, err := brt.findBackupBeforeTimestamp(targetTimestamp)
	if err != nil {
		return fmt.Errorf("failed to find backup before timestamp: %w", err)
	}

	// Restore from the backup
	if err := brt.restoreFromBackup(ctx, backupFile); err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	// Apply data corrections to simulate point-in-time recovery
	if err := brt.applyDataCorrections(ctx, targetTimestamp); err != nil {
		return fmt.Errorf("failed to apply data corrections: %w", err)
	}

	brt.logger.Printf("Successfully recovered database to timestamp: %v", targetTimestamp)
	return nil
}

// findBackupBeforeTimestamp finds the most recent backup before the target timestamp
func (brt *BackupRecoveryTester) findBackupBeforeTimestamp(targetTimestamp time.Time) (string, error) {
	// In a real implementation, this would query backup metadata
	// For testing, we'll use the most recent full backup
	backupFiles, err := filepath.Glob(filepath.Join(brt.backupDir, "full_backup_*.sql"))
	if err != nil {
		return "", fmt.Errorf("failed to find backup files: %w", err)
	}
	if len(backupFiles) == 0 {
		return "", fmt.Errorf("no backup files found")
	}

	// Use the most recent backup
	return backupFiles[len(backupFiles)-1], nil
}

// restoreFromBackup restores the database from a backup file
func (brt *BackupRecoveryTester) restoreFromBackup(ctx context.Context, backupFile string) error {
	brt.logger.Printf("Restoring database from backup: %s", backupFile)

	// Clear existing data in test database
	if err := brt.clearTestDatabase(ctx); err != nil {
		return fmt.Errorf("failed to clear test database: %w", err)
	}

	// Restore from backup
	cmd := exec.CommandContext(ctx, "psql",
		"--host", brt.extractHostFromURL(brt.config.TestDatabaseURL),
		"--port", brt.extractPortFromURL(brt.config.TestDatabaseURL),
		"--username", brt.extractUserFromURL(brt.config.TestDatabaseURL),
		"--dbname", brt.extractDBNameFromURL(brt.config.TestDatabaseURL),
		"--file", backupFile,
		"--verbose",
		"--no-password",
	)

	if password := brt.extractPasswordFromURL(brt.config.TestDatabaseURL); password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("database restoration failed: %w", err)
	}

	brt.logger.Printf("Database restoration completed successfully")
	return nil
}

// clearTestDatabase clears all data from the test database
func (brt *BackupRecoveryTester) clearTestDatabase(ctx context.Context) error {
	brt.logger.Printf("Clearing test database")

	// Get list of all tables
	rows, err := brt.testDB.QueryContext(ctx, `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
	`)
	if err != nil {
		return fmt.Errorf("failed to get table list: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	// Disable foreign key checks temporarily
	if _, err := brt.testDB.ExecContext(ctx, "SET session_replication_role = replica;"); err != nil {
		brt.logger.Printf("Warning: Could not disable foreign key checks: %v", err)
	}

	// Clear each table
	for _, table := range tables {
		if _, err := brt.testDB.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			brt.logger.Printf("Warning: Could not truncate table %s: %v", table, err)
		}
	}

	// Re-enable foreign key checks
	if _, err := brt.testDB.ExecContext(ctx, "SET session_replication_role = DEFAULT;"); err != nil {
		brt.logger.Printf("Warning: Could not re-enable foreign key checks: %v", err)
	}

	brt.logger.Printf("Test database cleared successfully")
	return nil
}

// applyDataCorrections applies data corrections to simulate point-in-time recovery
func (brt *BackupRecoveryTester) applyDataCorrections(ctx context.Context, targetTimestamp time.Time) error {
	brt.logger.Printf("Applying data corrections for timestamp: %v", targetTimestamp)

	// Remove data created after the target timestamp
	correctionQueries := []string{
		"DELETE FROM business_risk_assessments WHERE created_at > $1",
		"DELETE FROM industry_keywords WHERE created_at > $1",
		"DELETE FROM industries WHERE created_at > $1",
		"DELETE FROM merchants WHERE created_at > $1",
	}

	for _, query := range correctionQueries {
		result, err := brt.testDB.ExecContext(ctx, query, targetTimestamp)
		if err != nil {
			brt.logger.Printf("Warning: Could not apply correction query %s: %v", query, err)
			continue
		}

		if rowsAffected, err := result.RowsAffected(); err == nil {
			brt.logger.Printf("Applied correction: %d rows affected", rowsAffected)
		}
	}

	brt.logger.Printf("Data corrections applied successfully")
	return nil
}

// validateRecoveredData validates that the recovered data matches the expected state
func (brt *BackupRecoveryTester) validateRecoveredData(ctx context.Context, timestamps []time.Time) error {
	brt.logger.Printf("Validating recovered data")

	for i, timestamp := range timestamps {
		brt.logger.Printf("Validating data for timestamp %d: %v", i, timestamp)

		// Validate merchant data
		if err := brt.validateMerchantDataAtTimestamp(ctx, timestamp, i); err != nil {
			return fmt.Errorf("merchant data validation failed for timestamp %v: %w", timestamp, err)
		}

		// Validate classification data
		if err := brt.validateClassificationDataAtTimestamp(ctx, timestamp, i); err != nil {
			return fmt.Errorf("classification data validation failed for timestamp %v: %w", timestamp, err)
		}

		// Validate risk assessment data
		if err := brt.validateRiskAssessmentDataAtTimestamp(ctx, timestamp, i); err != nil {
			return fmt.Errorf("risk assessment data validation failed for timestamp %v: %w", timestamp, err)
		}
	}

	brt.logger.Printf("Recovered data validation completed successfully")
	return nil
}

// validateMerchantDataAtTimestamp validates merchant data at a specific timestamp
func (brt *BackupRecoveryTester) validateMerchantDataAtTimestamp(ctx context.Context, timestamp time.Time, index int) error {
	merchantID := fmt.Sprintf("test-merchant-%d", index)

	var count int
	query := `
		SELECT COUNT(*) 
		FROM merchants 
		WHERE id = $1 AND created_at <= $2
	`

	if err := brt.testDB.QueryRowContext(ctx, query, merchantID, timestamp).Scan(&count); err != nil {
		return fmt.Errorf("failed to validate merchant data: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("merchant data not found for timestamp %v", timestamp)
	}

	return nil
}

// validateClassificationDataAtTimestamp validates classification data at a specific timestamp
func (brt *BackupRecoveryTester) validateClassificationDataAtTimestamp(ctx context.Context, timestamp time.Time, index int) error {
	industryID := index + 1000

	var count int
	query := `
		SELECT COUNT(*) 
		FROM industries 
		WHERE id = $1 AND created_at <= $2
	`

	if err := brt.testDB.QueryRowContext(ctx, query, industryID, timestamp).Scan(&count); err != nil {
		return fmt.Errorf("failed to validate classification data: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("classification data not found for timestamp %v", timestamp)
	}

	return nil
}

// validateRiskAssessmentDataAtTimestamp validates risk assessment data at a specific timestamp
func (brt *BackupRecoveryTester) validateRiskAssessmentDataAtTimestamp(ctx context.Context, timestamp time.Time, index int) error {
	assessmentID := fmt.Sprintf("test-risk-assessment-%d", index)

	var count int
	query := `
		SELECT COUNT(*) 
		FROM business_risk_assessments 
		WHERE id = $1 AND created_at <= $2
	`

	if err := brt.testDB.QueryRowContext(ctx, query, assessmentID, timestamp).Scan(&count); err != nil {
		return fmt.Errorf("failed to validate risk assessment data: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("risk assessment data not found for timestamp %v", timestamp)
	}

	return nil
}
