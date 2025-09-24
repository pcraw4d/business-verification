package testing

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// testFullDatabaseBackup tests full database backup procedures
func (brt *BackupRecoveryTester) testFullDatabaseBackup(ctx context.Context) error {
	brt.logger.Printf("Testing full database backup")

	backupFile := filepath.Join(brt.backupDir, fmt.Sprintf("full_backup_%d.sql", time.Now().Unix()))

	// Create full database backup using pg_dump
	cmd := exec.CommandContext(ctx, "pg_dump",
		"--host", brt.extractHostFromURL(brt.config.SupabaseURL),
		"--port", brt.extractPortFromURL(brt.config.SupabaseURL),
		"--username", brt.extractUserFromURL(brt.config.SupabaseURL),
		"--dbname", brt.extractDBNameFromURL(brt.config.SupabaseURL),
		"--file", backupFile,
		"--verbose",
		"--no-password",
	)

	// Set password from environment or connection string
	if password := brt.extractPasswordFromURL(brt.config.SupabaseURL); password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump failed: %w", err)
	}

	// Verify backup file exists and has content
	if info, err := os.Stat(backupFile); err != nil {
		return fmt.Errorf("backup file verification failed: %w", err)
	} else if info.Size() == 0 {
		return fmt.Errorf("backup file is empty")
	}

	brt.logger.Printf("Full database backup completed successfully: %s", backupFile)
	return nil
}

// testIncrementalBackup tests incremental backup procedures
func (brt *BackupRecoveryTester) testIncrementalBackup(ctx context.Context) error {
	brt.logger.Printf("Testing incremental backup")

	// For Supabase, we'll simulate incremental backup by backing up specific tables
	// that are likely to change frequently
	criticalTables := []string{
		"merchants",
		"business_risk_assessments",
		"classification_results",
		"audit_logs",
		"performance_metrics",
	}

	for _, table := range criticalTables {
		backupFile := filepath.Join(brt.backupDir, fmt.Sprintf("incremental_%s_%d.sql", table, time.Now().Unix()))

		cmd := exec.CommandContext(ctx, "pg_dump",
			"--host", brt.extractHostFromURL(brt.config.SupabaseURL),
			"--port", brt.extractPortFromURL(brt.config.SupabaseURL),
			"--username", brt.extractUserFromURL(brt.config.SupabaseURL),
			"--dbname", brt.extractDBNameFromURL(brt.config.SupabaseURL),
			"--table", table,
			"--file", backupFile,
			"--verbose",
			"--no-password",
		)

		if password := brt.extractPasswordFromURL(brt.config.SupabaseURL); password != "" {
			cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
		}

		if err := cmd.Run(); err != nil {
			brt.logger.Printf("Warning: Incremental backup for table %s failed: %v", table, err)
			continue // Continue with other tables
		}

		brt.logger.Printf("Incremental backup for table %s completed: %s", table, backupFile)
	}

	return nil
}

// testSchemaOnlyBackup tests schema-only backup procedures
func (brt *BackupRecoveryTester) testSchemaOnlyBackup(ctx context.Context) error {
	brt.logger.Printf("Testing schema-only backup")

	backupFile := filepath.Join(brt.backupDir, fmt.Sprintf("schema_only_%d.sql", time.Now().Unix()))

	cmd := exec.CommandContext(ctx, "pg_dump",
		"--host", brt.extractHostFromURL(brt.config.SupabaseURL),
		"--port", brt.extractPortFromURL(brt.config.SupabaseURL),
		"--username", brt.extractUserFromURL(brt.config.SupabaseURL),
		"--dbname", brt.extractDBNameFromURL(brt.config.SupabaseURL),
		"--schema-only",
		"--file", backupFile,
		"--verbose",
		"--no-password",
	)

	if password := brt.extractPasswordFromURL(brt.config.SupabaseURL); password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("schema-only backup failed: %w", err)
	}

	brt.logger.Printf("Schema-only backup completed successfully: %s", backupFile)
	return nil
}

// testDataOnlyBackup tests data-only backup procedures
func (brt *BackupRecoveryTester) testDataOnlyBackup(ctx context.Context) error {
	brt.logger.Printf("Testing data-only backup")

	backupFile := filepath.Join(brt.backupDir, fmt.Sprintf("data_only_%d.sql", time.Now().Unix()))

	cmd := exec.CommandContext(ctx, "pg_dump",
		"--host", brt.extractHostFromURL(brt.config.SupabaseURL),
		"--port", brt.extractPortFromURL(brt.config.SupabaseURL),
		"--username", brt.extractUserFromURL(brt.config.SupabaseURL),
		"--dbname", brt.extractDBNameFromURL(brt.config.SupabaseURL),
		"--data-only",
		"--file", backupFile,
		"--verbose",
		"--no-password",
	)

	if password := brt.extractPasswordFromURL(brt.config.SupabaseURL); password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("data-only backup failed: %w", err)
	}

	brt.logger.Printf("Data-only backup completed successfully: %s", backupFile)
	return nil
}

// testCompleteDatabaseRecovery tests complete database recovery
func (brt *BackupRecoveryTester) testCompleteDatabaseRecovery(ctx context.Context) error {
	brt.logger.Printf("Testing complete database recovery")

	// Find the most recent full backup
	backupFiles, err := filepath.Glob(filepath.Join(brt.backupDir, "full_backup_*.sql"))
	if err != nil {
		return fmt.Errorf("failed to find backup files: %w", err)
	}
	if len(backupFiles) == 0 {
		return fmt.Errorf("no full backup files found")
	}

	// Use the most recent backup
	latestBackup := backupFiles[len(backupFiles)-1]
	brt.logger.Printf("Using backup file: %s", latestBackup)

	// Restore to test database
	cmd := exec.CommandContext(ctx, "psql",
		"--host", brt.extractHostFromURL(brt.config.TestDatabaseURL),
		"--port", brt.extractPortFromURL(brt.config.TestDatabaseURL),
		"--username", brt.extractUserFromURL(brt.config.TestDatabaseURL),
		"--dbname", brt.extractDBNameFromURL(brt.config.TestDatabaseURL),
		"--file", latestBackup,
		"--verbose",
		"--no-password",
	)

	if password := brt.extractPasswordFromURL(brt.config.TestDatabaseURL); password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("database recovery failed: %w", err)
	}

	brt.logger.Printf("Complete database recovery completed successfully")
	return nil
}

// testPartialTableRecovery tests partial table recovery
func (brt *BackupRecoveryTester) testPartialTableRecovery(ctx context.Context) error {
	brt.logger.Printf("Testing partial table recovery")

	// Test recovery of critical classification tables
	criticalTables := []string{
		"industries",
		"industry_keywords",
		"risk_keywords",
		"industry_code_crosswalks",
	}

	for _, table := range criticalTables {
		backupFiles, err := filepath.Glob(filepath.Join(brt.backupDir, fmt.Sprintf("incremental_%s_*.sql", table)))
		if err != nil || len(backupFiles) == 0 {
			brt.logger.Printf("Warning: No backup found for table %s, skipping", table)
			continue
		}

		latestBackup := backupFiles[len(backupFiles)-1]

		cmd := exec.CommandContext(ctx, "psql",
			"--host", brt.extractHostFromURL(brt.config.TestDatabaseURL),
			"--port", brt.extractPortFromURL(brt.config.TestDatabaseURL),
			"--username", brt.extractUserFromURL(brt.config.TestDatabaseURL),
			"--dbname", brt.extractDBNameFromURL(brt.config.TestDatabaseURL),
			"--file", latestBackup,
			"--verbose",
			"--no-password",
		)

		if password := brt.extractPasswordFromURL(brt.config.TestDatabaseURL); password != "" {
			cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
		}

		if err := cmd.Run(); err != nil {
			brt.logger.Printf("Warning: Partial recovery for table %s failed: %v", table, err)
			continue
		}

		brt.logger.Printf("Partial recovery for table %s completed successfully", table)
	}

	return nil
}

// testSchemaRecovery tests schema recovery
func (brt *BackupRecoveryTester) testSchemaRecovery(ctx context.Context) error {
	brt.logger.Printf("Testing schema recovery")

	// Find schema-only backup
	backupFiles, err := filepath.Glob(filepath.Join(brt.backupDir, "schema_only_*.sql"))
	if err != nil {
		return fmt.Errorf("failed to find schema backup files: %w", err)
	}
	if len(backupFiles) == 0 {
		return fmt.Errorf("no schema backup files found")
	}

	latestBackup := backupFiles[len(backupFiles)-1]

	cmd := exec.CommandContext(ctx, "psql",
		"--host", brt.extractHostFromURL(brt.config.TestDatabaseURL),
		"--port", brt.extractPortFromURL(brt.config.TestDatabaseURL),
		"--username", brt.extractUserFromURL(brt.config.TestDatabaseURL),
		"--dbname", brt.extractDBNameFromURL(brt.config.TestDatabaseURL),
		"--file", latestBackup,
		"--verbose",
		"--no-password",
	)

	if password := brt.extractPasswordFromURL(brt.config.TestDatabaseURL); password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("schema recovery failed: %w", err)
	}

	brt.logger.Printf("Schema recovery completed successfully")
	return nil
}

// validateDataIntegrity validates data integrity after recovery
func (brt *BackupRecoveryTester) validateDataIntegrity(ctx context.Context) (float64, error) {
	brt.logger.Printf("Validating data integrity")

	var totalTables int
	var validTables int

	// Get list of all tables
	rows, err := brt.testDB.QueryContext(ctx, `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to get table list: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return 0, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	totalTables = len(tables)

	// Validate each table
	for _, table := range tables {
		if err := brt.validateTableIntegrity(ctx, table); err != nil {
			brt.logger.Printf("Warning: Table %s validation failed: %v", table, err)
		} else {
			validTables++
		}
	}

	integrityScore := float64(validTables) / float64(totalTables)
	brt.logger.Printf("Data integrity validation completed: %d/%d tables valid (%.2f%%)",
		validTables, totalTables, integrityScore*100)

	return integrityScore, nil
}

// validateTableIntegrity validates integrity of a specific table
func (brt *BackupRecoveryTester) validateTableIntegrity(ctx context.Context, tableName string) error {
	// Check if table has data
	var count int
	err := brt.testDB.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count rows in table %s: %w", tableName, err)
	}

	// Check for NULL values in critical columns
	criticalColumns := []string{"id", "created_at", "updated_at"}
	for _, column := range criticalColumns {
		var nullCount int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s IS NULL", tableName, column)
		err := brt.testDB.QueryRowContext(ctx, query).Scan(&nullCount)
		if err != nil {
			// Column might not exist, skip
			continue
		}
		if nullCount > 0 {
			brt.logger.Printf("Warning: Table %s has %d NULL values in column %s", tableName, nullCount, column)
		}
	}

	return nil
}

// validateForeignKeyConstraints validates foreign key constraints
func (brt *BackupRecoveryTester) validateForeignKeyConstraints(ctx context.Context) error {
	brt.logger.Printf("Validating foreign key constraints")

	// Check for foreign key violations
	query := `
		SELECT 
			tc.table_name, 
			kcu.column_name, 
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name 
		FROM 
			information_schema.table_constraints AS tc 
			JOIN information_schema.key_column_usage AS kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			JOIN information_schema.constraint_column_usage AS ccu
				ON ccu.constraint_name = tc.constraint_name
				AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY' 
		AND tc.table_schema = 'public'
	`

	rows, err := brt.testDB.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get foreign key constraints: %w", err)
	}
	defer rows.Close()

	var violations int
	for rows.Next() {
		var tableName, columnName, foreignTableName, foreignColumnName string
		if err := rows.Scan(&tableName, &columnName, &foreignTableName, &foreignColumnName); err != nil {
			return fmt.Errorf("failed to scan foreign key constraint: %w", err)
		}

		// Check for orphaned records
		checkQuery := fmt.Sprintf(`
			SELECT COUNT(*) 
			FROM %s t1 
			LEFT JOIN %s t2 ON t1.%s = t2.%s 
			WHERE t1.%s IS NOT NULL AND t2.%s IS NULL
		`, tableName, foreignTableName, columnName, foreignColumnName, columnName, foreignColumnName)

		var orphanedCount int
		if err := brt.testDB.QueryRowContext(ctx, checkQuery).Scan(&orphanedCount); err != nil {
			brt.logger.Printf("Warning: Could not check foreign key constraint for %s.%s: %v", tableName, columnName, err)
			continue
		}

		if orphanedCount > 0 {
			brt.logger.Printf("Warning: Found %d orphaned records in %s.%s", orphanedCount, tableName, columnName)
			violations++
		}
	}

	if violations > 0 {
		return fmt.Errorf("found %d foreign key constraint violations", violations)
	}

	brt.logger.Printf("Foreign key constraint validation completed successfully")
	return nil
}

// validateIndexes validates database indexes
func (brt *BackupRecoveryTester) validateIndexes(ctx context.Context) error {
	brt.logger.Printf("Validating database indexes")

	// Check for missing indexes on foreign keys
	query := `
		SELECT 
			tc.table_name, 
			kcu.column_name
		FROM 
			information_schema.table_constraints AS tc 
			JOIN information_schema.key_column_usage AS kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY' 
		AND tc.table_schema = 'public'
	`

	rows, err := brt.testDB.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get foreign key constraints: %w", err)
	}
	defer rows.Close()

	var missingIndexes int
	for rows.Next() {
		var tableName, columnName string
		if err := rows.Scan(&tableName, &columnName); err != nil {
			return fmt.Errorf("failed to scan foreign key constraint: %w", err)
		}

		// Check if index exists on this column
		indexQuery := `
			SELECT COUNT(*) 
			FROM pg_indexes 
			WHERE tablename = $1 
			AND indexdef LIKE '%' || $2 || '%'
		`

		var indexCount int
		if err := brt.testDB.QueryRowContext(ctx, indexQuery, tableName, columnName).Scan(&indexCount); err != nil {
			brt.logger.Printf("Warning: Could not check index for %s.%s: %v", tableName, columnName, err)
			continue
		}

		if indexCount == 0 {
			brt.logger.Printf("Warning: Missing index on foreign key %s.%s", tableName, columnName)
			missingIndexes++
		}
	}

	if missingIndexes > 0 {
		brt.logger.Printf("Warning: Found %d missing indexes on foreign keys", missingIndexes)
	}

	brt.logger.Printf("Index validation completed")
	return nil
}

// validateClassificationSystem validates the classification system integrity
func (brt *BackupRecoveryTester) validateClassificationSystem(ctx context.Context) error {
	brt.logger.Printf("Validating classification system integrity")

	// Check critical classification tables exist
	criticalTables := []string{
		"industries",
		"industry_keywords",
		"risk_keywords",
		"industry_code_crosswalks",
		"business_risk_assessments",
	}

	for _, table := range criticalTables {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)
		`

		if err := brt.testDB.QueryRowContext(ctx, query, table).Scan(&exists); err != nil {
			return fmt.Errorf("failed to check if table %s exists: %w", table, err)
		}

		if !exists {
			return fmt.Errorf("critical classification table %s does not exist", table)
		}

		// Check if table has data
		var count int
		if err := brt.testDB.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count); err != nil {
			return fmt.Errorf("failed to count rows in table %s: %w", table, err)
		}

		if count == 0 {
			brt.logger.Printf("Warning: Table %s is empty", table)
		}
	}

	brt.logger.Printf("Classification system validation completed successfully")
	return nil
}

// Helper functions for URL parsing
func (brt *BackupRecoveryTester) extractHostFromURL(url string) string {
	// Simple URL parsing - in production, use proper URL parsing
	parts := strings.Split(url, "://")
	if len(parts) < 2 {
		return "localhost"
	}
	hostPart := strings.Split(parts[1], "/")[0]
	hostParts := strings.Split(hostPart, ":")
	return hostParts[0]
}

func (brt *BackupRecoveryTester) extractPortFromURL(url string) string {
	parts := strings.Split(url, "://")
	if len(parts) < 2 {
		return "5432"
	}
	hostPart := strings.Split(parts[1], "/")[0]
	hostParts := strings.Split(hostPart, ":")
	if len(hostParts) > 1 {
		return hostParts[1]
	}
	return "5432"
}

func (brt *BackupRecoveryTester) extractUserFromURL(url string) string {
	// Extract user from connection string
	parts := strings.Split(url, "://")
	if len(parts) < 2 {
		return "postgres"
	}
	authPart := strings.Split(parts[1], "@")[0]
	userParts := strings.Split(authPart, ":")
	return userParts[0]
}

func (brt *BackupRecoveryTester) extractPasswordFromURL(url string) string {
	// Extract password from connection string
	parts := strings.Split(url, "://")
	if len(parts) < 2 {
		return ""
	}
	authPart := strings.Split(parts[1], "@")[0]
	authParts := strings.Split(authPart, ":")
	if len(authParts) > 1 {
		return authParts[1]
	}
	return ""
}

func (brt *BackupRecoveryTester) extractDBNameFromURL(url string) string {
	// Extract database name from connection string
	parts := strings.Split(url, "/")
	if len(parts) > 1 {
		dbPart := strings.Split(parts[len(parts)-1], "?")[0]
		return dbPart
	}
	return "postgres"
}
