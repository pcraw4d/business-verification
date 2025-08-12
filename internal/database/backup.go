package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// BackupConfig holds configuration for database backups
type BackupConfig struct {
	Enabled           bool
	BackupDir         string
	RetentionDays     int
	Compression       bool
	Encryption        bool
	EncryptionKey     string
	CrossRegion       bool
	CrossRegionBucket string
	Schedule          string
}

// BackupService handles database backup operations
type BackupService struct {
	db     *sql.DB
	config BackupConfig
	logger *observability.Logger
}

// BackupResult represents the result of a backup operation
type BackupResult struct {
	BackupID    string
	Filename    string
	Size        int64
	Checksum    string
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Success     bool
	Error       error
	Compressed  bool
	Encrypted   bool
	CrossRegion bool
}

// NewBackupService creates a new backup service
func NewBackupService(db *sql.DB, config BackupConfig, logger *observability.Logger) *BackupService {
	return &BackupService{
		db:     db,
		config: config,
		logger: logger,
	}
}

// CreateBackup creates a full database backup
func (bs *BackupService) CreateBackup(ctx context.Context) (*BackupResult, error) {
	startTime := time.Now()
	backupID := fmt.Sprintf("backup-%s", startTime.Format("20060102-150405"))

	// Validate database configuration before executing commands
	if err := bs.validateDatabaseConfig(); err != nil {
		bs.logger.Error("Database configuration validation failed", "error", err, "backup_id", backupID)
		return &BackupResult{
			BackupID:  backupID,
			StartTime: startTime,
			EndTime:   time.Now(),
			Success:   false,
			Error:     err,
		}, err
	}

	// Validate backup directory with secure permissions
	if err := os.MkdirAll(bs.config.BackupDir, 0750); err != nil {
		bs.logger.Error("Failed to create backup directory", "error", err, "backup_id", backupID)
		return &BackupResult{
			BackupID:  backupID,
			StartTime: startTime,
			EndTime:   time.Now(),
			Success:   false,
			Error:     err,
		}, err
	}

	// Generate backup filename
	filename := filepath.Join(bs.config.BackupDir, fmt.Sprintf("%s.sql", backupID))

	// Create backup using pg_dump with validated environment variables
	cmd := exec.CommandContext(ctx, "pg_dump",
		"--host="+os.Getenv("DB_HOST"),
		"--port="+os.Getenv("DB_PORT"),
		"--username="+os.Getenv("DB_USER"),
		"--dbname="+os.Getenv("DB_NAME"),
		"--verbose",
		"--no-password",
		"--file="+filename,
	)

	// Set environment variables for authentication
	cmd.Env = append(os.Environ(),
		"PGPASSWORD="+os.Getenv("DB_PASSWORD"),
	)

	// Execute backup command
	if err := cmd.Run(); err != nil {
		bs.logger.Error("Backup command failed", "error", err, "backup_id", backupID)
		return &BackupResult{
			BackupID:  backupID,
			StartTime: startTime,
			EndTime:   time.Now(),
			Success:   false,
			Error:     err,
		}, err
	}

	// Get file info
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup file info: %w", err)
	}

	// Calculate checksum
	checksum, err := bs.calculateChecksum(filename)
	if err != nil {
		bs.logger.Warn("Failed to calculate checksum", "error", err, "backup_id", backupID)
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	result := &BackupResult{
		BackupID:    backupID,
		Filename:    filename,
		Size:        fileInfo.Size(),
		Checksum:    checksum,
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    duration,
		Success:     true,
		Compressed:  bs.config.Compression,
		Encrypted:   bs.config.Encryption,
		CrossRegion: bs.config.CrossRegion,
	}

	bs.logger.Info("Database backup completed successfully",
		"backup_id", backupID,
		"filename", filename,
		"size", fileInfo.Size(),
		"duration", duration,
	)

	// Compress backup if enabled
	if bs.config.Compression {
		if err := bs.compressBackup(filename); err != nil {
			bs.logger.Warn("Failed to compress backup", "error", err, "backup_id", backupID)
		} else {
			result.Filename = filename + ".gz"
			if fileInfo, err := os.Stat(result.Filename); err == nil {
				result.Size = fileInfo.Size()
			}
		}
	}

	// Encrypt backup if enabled
	if bs.config.Encryption {
		if err := bs.encryptBackup(result.Filename); err != nil {
			bs.logger.Warn("Failed to encrypt backup", "error", err, "backup_id", backupID)
		} else {
			result.Filename = result.Filename + ".enc"
			if fileInfo, err := os.Stat(result.Filename); err == nil {
				result.Size = fileInfo.Size()
			}
		}
	}

	// Upload to cross-region if enabled
	if bs.config.CrossRegion && bs.config.CrossRegionBucket != "" {
		if err := bs.uploadToCrossRegion(result.Filename, backupID); err != nil {
			bs.logger.Warn("Failed to upload to cross-region", "error", err, "backup_id", backupID)
		}
	}

	// Store backup metadata in database
	if err := bs.storeBackupMetadata(ctx, result); err != nil {
		bs.logger.Warn("Failed to store backup metadata", "error", err, "backup_id", backupID)
	}

	return result, nil
}

// RestoreBackup restores a database from backup
func (bs *BackupService) RestoreBackup(ctx context.Context, backupID string) error {
	bs.logger.Info("Starting database restore", "backup_id", backupID)

	// Get backup metadata
	metadata, err := bs.getBackupMetadata(ctx, backupID)
	if err != nil {
		return fmt.Errorf("failed to get backup metadata: %w", err)
	}

	// Check if backup file exists
	if _, err := os.Stat(metadata.Filename); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", metadata.Filename)
	}

	// Decrypt backup if it was encrypted
	filename := metadata.Filename
	if metadata.Encrypted {
		if err := bs.decryptBackup(filename); err != nil {
			return fmt.Errorf("failed to decrypt backup: %w", err)
		}
		filename = filename[:len(filename)-4] // Remove .enc extension
	}

	// Decompress backup if it was compressed
	if metadata.Compressed {
		if err := bs.decompressBackup(filename); err != nil {
			return fmt.Errorf("failed to decompress backup: %w", err)
		}
		filename = filename[:len(filename)-3] // Remove .gz extension
	}

	// Restore database using psql
	cmd := exec.CommandContext(ctx, "psql",
		"--host="+os.Getenv("DB_HOST"),
		"--port="+os.Getenv("DB_PORT"),
		"--username="+os.Getenv("DB_USER"),
		"--dbname="+os.Getenv("DB_NAME"),
		"--verbose",
		"--no-password",
		"--file="+filename,
	)

	// Set environment variables for authentication
	cmd.Env = append(os.Environ(),
		"PGPASSWORD="+os.Getenv("DB_PASSWORD"),
	)

	// Execute restore command
	if err := cmd.Run(); err != nil {
		bs.logger.Error("Restore command failed", "error", err, "backup_id", backupID)
		return err
	}

	bs.logger.Info("Database restore completed successfully", "backup_id", backupID)
	return nil
}

// ListBackups returns a list of available backups
func (bs *BackupService) ListBackups(ctx context.Context) ([]*BackupResult, error) {
	query := `
		SELECT backup_id, filename, size, checksum, start_time, end_time, 
		       duration_ms, success, compressed, encrypted, cross_region
		FROM backup_metadata 
		ORDER BY start_time DESC
	`

	rows, err := bs.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query backup metadata: %w", err)
	}
	defer rows.Close()

	var backups []*BackupResult
	for rows.Next() {
		var backup BackupResult
		var durationMs int64
		var success bool

		err := rows.Scan(
			&backup.BackupID,
			&backup.Filename,
			&backup.Size,
			&backup.Checksum,
			&backup.StartTime,
			&backup.EndTime,
			&durationMs,
			&success,
			&backup.Compressed,
			&backup.Encrypted,
			&backup.CrossRegion,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan backup metadata: %w", err)
		}

		backup.Duration = time.Duration(durationMs) * time.Millisecond
		backup.Success = success
		backups = append(backups, &backup)
	}

	return backups, nil
}

// CleanupOldBackups removes backups older than the retention period
func (bs *BackupService) CleanupOldBackups(ctx context.Context) error {
	bs.logger.Info("Starting cleanup of old backups", "retention_days", bs.config.RetentionDays)

	cutoffDate := time.Now().AddDate(0, 0, -bs.config.RetentionDays)

	// Get old backups
	query := `
		SELECT backup_id, filename 
		FROM backup_metadata 
		WHERE start_time < $1 AND success = true
	`

	rows, err := bs.db.QueryContext(ctx, query, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to query old backups: %w", err)
	}
	defer rows.Close()

	var deletedCount int
	for rows.Next() {
		var backupID, filename string
		if err := rows.Scan(&backupID, &filename); err != nil {
			bs.logger.Warn("Failed to scan backup record", "error", err)
			continue
		}

		// Delete backup file
		if err := os.Remove(filename); err != nil {
			bs.logger.Warn("Failed to delete backup file", "error", err, "filename", filename)
			continue
		}

		// Delete metadata
		if _, err := bs.db.ExecContext(ctx, "DELETE FROM backup_metadata WHERE backup_id = $1", backupID); err != nil {
			bs.logger.Warn("Failed to delete backup metadata", "error", err, "backup_id", backupID)
			continue
		}

		deletedCount++
		bs.logger.Info("Deleted old backup", "backup_id", backupID, "filename", filename)
	}

	bs.logger.Info("Backup cleanup completed", "deleted_count", deletedCount)
	return nil
}

// ValidateBackup validates the integrity of a backup
func (bs *BackupService) ValidateBackup(ctx context.Context, backupID string) error {
	bs.logger.Info("Validating backup", "backup_id", backupID)

	// Get backup metadata
	metadata, err := bs.getBackupMetadata(ctx, backupID)
	if err != nil {
		return fmt.Errorf("failed to get backup metadata: %w", err)
	}

	// Check if backup file exists
	if _, err := os.Stat(metadata.Filename); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", metadata.Filename)
	}

	// Calculate current checksum
	currentChecksum, err := bs.calculateChecksum(metadata.Filename)
	if err != nil {
		return fmt.Errorf("failed to calculate current checksum: %w", err)
	}

	// Compare checksums
	if currentChecksum != metadata.Checksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", metadata.Checksum, currentChecksum)
	}

	bs.logger.Info("Backup validation completed successfully", "backup_id", backupID)
	return nil
}

// Helper methods

func (bs *BackupService) calculateChecksum(filename string) (string, error) {
	cmd := exec.Command("sha256sum", filename)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Extract checksum from output (format: "checksum filename")
	checksum := string(output[:64])
	return checksum, nil
}

func (bs *BackupService) compressBackup(filename string) error {
	cmd := exec.Command("gzip", filename)
	return cmd.Run()
}

func (bs *BackupService) decompressBackup(filename string) error {
	cmd := exec.Command("gunzip", filename)
	return cmd.Run()
}

func (bs *BackupService) encryptBackup(filename string) error {
	// Simple encryption using openssl
	cmd := exec.Command("openssl", "enc", "-aes-256-cbc", "-salt", "-in", filename, "-out", filename+".enc", "-k", bs.config.EncryptionKey)
	return cmd.Run()
}

func (bs *BackupService) decryptBackup(filename string) error {
	// Simple decryption using openssl
	cmd := exec.Command("openssl", "enc", "-d", "-aes-256-cbc", "-in", filename, "-out", filename[:len(filename)-4], "-k", bs.config.EncryptionKey)
	return cmd.Run()
}

func (bs *BackupService) uploadToCrossRegion(filename, backupID string) error {
	// Upload to S3 using AWS CLI
	cmd := exec.Command("aws", "s3", "cp", filename, fmt.Sprintf("s3://%s/backups/%s", bs.config.CrossRegionBucket, backupID))
	return cmd.Run()
}

func (bs *BackupService) storeBackupMetadata(ctx context.Context, result *BackupResult) error {
	query := `
		INSERT INTO backup_metadata (
			backup_id, filename, size, checksum, start_time, end_time, 
			duration_ms, success, compressed, encrypted, cross_region
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := bs.db.ExecContext(ctx, query,
		result.BackupID,
		result.Filename,
		result.Size,
		result.Checksum,
		result.StartTime,
		result.EndTime,
		result.Duration.Milliseconds(),
		result.Success,
		result.Compressed,
		result.Encrypted,
		result.CrossRegion,
	)

	return err
}

func (bs *BackupService) getBackupMetadata(ctx context.Context, backupID string) (*BackupResult, error) {
	query := `
		SELECT backup_id, filename, size, checksum, start_time, end_time, 
		       duration_ms, success, compressed, encrypted, cross_region
		FROM backup_metadata 
		WHERE backup_id = $1
	`

	var backup BackupResult
	var durationMs int64
	var success bool

	err := bs.db.QueryRowContext(ctx, query, backupID).Scan(
		&backup.BackupID,
		&backup.Filename,
		&backup.Size,
		&backup.Checksum,
		&backup.StartTime,
		&backup.EndTime,
		&durationMs,
		&success,
		&backup.Compressed,
		&backup.Encrypted,
		&backup.CrossRegion,
	)

	if err != nil {
		return nil, err
	}

	backup.Duration = time.Duration(durationMs) * time.Millisecond
	backup.Success = success

	return &backup, nil
}

// CreateBackupTable creates the backup metadata table
func (bs *BackupService) CreateBackupTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS backup_metadata (
			id SERIAL PRIMARY KEY,
			backup_id VARCHAR(50) UNIQUE NOT NULL,
			filename VARCHAR(500) NOT NULL,
			size BIGINT NOT NULL,
			checksum VARCHAR(64),
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			duration_ms BIGINT NOT NULL,
			success BOOLEAN NOT NULL,
			compressed BOOLEAN DEFAULT FALSE,
			encrypted BOOLEAN DEFAULT FALSE,
			cross_region BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := bs.db.ExecContext(ctx, query)
	return err
}

// validateEnvironmentVariable validates that an environment variable is safe for use in commands
func validateEnvironmentVariable(name, value string) error {
	if value == "" {
		return fmt.Errorf("environment variable %s is empty", name)
	}

	// Check for potentially dangerous characters
	dangerousChars := []string{";", "&", "|", "`", "$(", ")", ">", "<", "\"", "'"}
	for _, char := range dangerousChars {
		if strings.Contains(value, char) {
			return fmt.Errorf("environment variable %s contains dangerous character: %s", name, char)
		}
	}

	// Check for command injection patterns
	if strings.Contains(strings.ToLower(value), "exec") ||
		strings.Contains(strings.ToLower(value), "system") ||
		strings.Contains(strings.ToLower(value), "eval") {
		return fmt.Errorf("environment variable %s contains potentially dangerous content", name)
	}

	return nil
}

// validateDatabaseConfig validates database configuration for backup operations
func (bs *BackupService) validateDatabaseConfig() error {
	requiredVars := map[string]string{
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_PORT":     os.Getenv("DB_PORT"),
		"DB_USER":     os.Getenv("DB_USER"),
		"DB_NAME":     os.Getenv("DB_NAME"),
		"DB_PASSWORD": os.Getenv("DB_PASSWORD"),
	}

	for name, value := range requiredVars {
		if err := validateEnvironmentVariable(name, value); err != nil {
			return fmt.Errorf("database configuration validation failed: %w", err)
		}
	}

	return nil
}
