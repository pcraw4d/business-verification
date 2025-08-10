package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Migration represents a database migration
type Migration struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SQL         string    `json:"sql"`
	AppliedAt   time.Time `json:"applied_at"`
	Checksum    string    `json:"checksum"`
}

// MigrationSystem handles database migrations
type MigrationSystem struct {
	db     *sql.DB
	config *DatabaseConfig
}

// NewMigrationSystem creates a new migration system
func NewMigrationSystem(db *sql.DB, config *DatabaseConfig) *MigrationSystem {
	return &MigrationSystem{
		db:     db,
		config: config,
	}
}

// InitializeMigrationTable creates the migrations table if it doesn't exist
func (m *MigrationSystem) InitializeMigrationTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			sql_content TEXT NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
			checksum VARCHAR(64) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_migrations_applied_at ON migrations(applied_at);
	`

	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

// GetAppliedMigrations returns all applied migrations
func (m *MigrationSystem) GetAppliedMigrations(ctx context.Context) ([]*Migration, error) {
	query := `
		SELECT id, name, description, sql_content, applied_at, checksum
		FROM migrations
		ORDER BY applied_at ASC
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var migrations []*Migration
	for rows.Next() {
		var migration Migration
		err := rows.Scan(
			&migration.ID,
			&migration.Name,
			&migration.Description,
			&migration.SQL,
			&migration.AppliedAt,
			&migration.Checksum,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migration: %w", err)
		}
		migrations = append(migrations, &migration)
	}

	return migrations, nil
}

// LoadMigrationFiles loads migration files from the migrations directory
func (m *MigrationSystem) LoadMigrationFiles() ([]*Migration, error) {
	migrationsDir := "internal/database/migrations"

	// Check if migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory does not exist: %s", migrationsDir)
	}

	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []*Migration
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migration, err := m.loadMigrationFile(filepath.Join(migrationsDir, file.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to load migration file %s: %w", file.Name(), err)
			}
			migrations = append(migrations, migration)
		}
	}

	// Sort migrations by ID
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	return migrations, nil
}

// loadMigrationFile loads a single migration file
func (m *MigrationSystem) loadMigrationFile(filePath string) (*Migration, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration file: %w", err)
	}

	filename := filepath.Base(filePath)
	// Extract migration ID from filename (e.g., "001_initial_schema.sql" -> "001")
	parts := strings.Split(filename, "_")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	migrationID := parts[0]

	// Extract description from filename
	description := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".sql")
	description = strings.ReplaceAll(description, "_", " ")

	migration := &Migration{
		ID:          migrationID,
		Name:        filename,
		Description: description,
		SQL:         string(content),
		Checksum:    calculateChecksum(string(content)),
	}

	return migration, nil
}

// calculateChecksum calculates a simple checksum for migration content
func calculateChecksum(content string) string {
	// Simple hash function for demonstration
	// In production, you might want to use a proper hash function
	hash := 0
	for _, char := range content {
		hash = (hash*31 + int(char)) % 1000000007
	}
	return fmt.Sprintf("%08x", hash)
}

// ApplyMigration applies a single migration
func (m *MigrationSystem) ApplyMigration(ctx context.Context, migration *Migration) error {
	// Check if migration is already applied
	applied, err := m.IsMigrationApplied(ctx, migration.ID)
	if err != nil {
		return fmt.Errorf("failed to check if migration is applied: %w", err)
	}

	if applied {
		log.Printf("Migration %s is already applied, skipping", migration.ID)
		return nil
	}

	// Begin transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	_, err = tx.ExecContext(ctx, migration.SQL)
	if err != nil {
		return fmt.Errorf("failed to execute migration %s: %w", migration.ID, err)
	}

	// Record migration as applied
	insertQuery := `
		INSERT INTO migrations (id, name, description, sql_content, checksum)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = tx.ExecContext(ctx, insertQuery,
		migration.ID,
		migration.Name,
		migration.Description,
		migration.SQL,
		migration.Checksum,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration %s: %w", migration.ID, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration %s: %w", migration.ID, err)
	}

	log.Printf("Successfully applied migration %s: %s", migration.ID, migration.Description)
	return nil
}

// IsMigrationApplied checks if a migration is already applied
func (m *MigrationSystem) IsMigrationApplied(ctx context.Context, migrationID string) (bool, error) {
	query := `SELECT COUNT(*) FROM migrations WHERE id = $1`

	var count int
	err := m.db.QueryRowContext(ctx, query, migrationID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check migration status: %w", err)
	}

	return count > 0, nil
}

// RunMigrations runs all pending migrations
func (m *MigrationSystem) RunMigrations(ctx context.Context) error {
	// Initialize migration table
	if err := m.InitializeMigrationTable(ctx); err != nil {
		return fmt.Errorf("failed to initialize migration table: %w", err)
	}

	// Load migration files
	migrations, err := m.LoadMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to load migration files: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Create a map of applied migration IDs for quick lookup
	appliedMap := make(map[string]bool)
	for _, applied := range appliedMigrations {
		appliedMap[applied.ID] = true
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if !appliedMap[migration.ID] {
			if err := m.ApplyMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", migration.ID, err)
			}
		}
	}

	log.Printf("All migrations completed successfully")
	return nil
}

// RollbackMigration rolls back a specific migration
func (m *MigrationSystem) RollbackMigration(ctx context.Context, migrationID string) error {
	// Get migration details
	query := `SELECT id, name, sql_content FROM migrations WHERE id = $1`

	var migration Migration
	err := m.db.QueryRowContext(ctx, query, migrationID).Scan(
		&migration.ID,
		&migration.Name,
		&migration.SQL,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("migration %s not found", migrationID)
		}
		return fmt.Errorf("failed to get migration details: %w", err)
	}

	// Note: In a real implementation, you would need to implement rollback logic
	// This is a simplified version that just removes the migration record
	// For production, you'd want to implement proper rollback SQL

	log.Printf("Rolling back migration %s: %s", migration.ID, migration.Name)

	// Begin transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Remove migration record
	deleteQuery := `DELETE FROM migrations WHERE id = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, migrationID)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	log.Printf("Successfully rolled back migration %s", migrationID)
	return nil
}

// GetMigrationStatus returns the status of all migrations
func (m *MigrationSystem) GetMigrationStatus(ctx context.Context) (map[string]interface{}, error) {
	// Load all migration files
	allMigrations, err := m.LoadMigrationFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to load migration files: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Create a map of applied migration IDs
	appliedMap := make(map[string]*Migration)
	for _, applied := range appliedMigrations {
		appliedMap[applied.ID] = applied
	}

	// Build status report
	status := map[string]interface{}{
		"total_migrations":   len(allMigrations),
		"applied_migrations": len(appliedMigrations),
		"pending_migrations": len(allMigrations) - len(appliedMigrations),
		"migrations":         []map[string]interface{}{},
		"last_applied":       nil,
		"last_applied_at":    nil,
	}

	var lastApplied *Migration
	for _, migration := range allMigrations {
		migrationStatus := map[string]interface{}{
			"id":          migration.ID,
			"name":        migration.Name,
			"description": migration.Description,
			"applied":     false,
			"applied_at":  nil,
		}

		if applied, exists := appliedMap[migration.ID]; exists {
			migrationStatus["applied"] = true
			migrationStatus["applied_at"] = applied.AppliedAt
			lastApplied = applied
		}

		status["migrations"] = append(status["migrations"].([]map[string]interface{}), migrationStatus)
	}

	if lastApplied != nil {
		status["last_applied"] = lastApplied.ID
		status["last_applied_at"] = lastApplied.AppliedAt
	}

	return status, nil
}

// CreateMigrationFile creates a new migration file
func (m *MigrationSystem) CreateMigrationFile(name, description string) error {
	migrationsDir := "internal/database/migrations"

	// Get the next migration ID
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	nextID := 1
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			parts := strings.Split(file.Name(), "_")
			if len(parts) >= 2 {
				var id int
				if _, err := fmt.Sscanf(parts[0], "%d", &id); err == nil {
					if id >= nextID {
						nextID = id + 1
					}
				}
			}
		}
	}

	// Create migration filename
	filename := fmt.Sprintf("%03d_%s.sql", nextID, strings.ToLower(strings.ReplaceAll(name, " ", "_")))
	filepath := filepath.Join(migrationsDir, filename)

	// Create migration content
	content := fmt.Sprintf(`-- Migration: %s
-- Description: %s
-- Created: %s

-- Add your migration SQL here
-- Example:
-- CREATE TABLE example_table (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
-- );

`, filename, description, time.Now().Format("2006-01-02"))

	// Write migration file
	err = ioutil.WriteFile(filepath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	log.Printf("Created migration file: %s", filepath)
	return nil
}
