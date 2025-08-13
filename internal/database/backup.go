package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// BackupConfig holds configuration for backup operations
type BackupConfig struct {
	BackupDir         string
	Compression       bool
	Encryption        bool
	EncryptionKey     string
	CrossRegion       bool
	CrossRegionBucket string
	MaxBackupAge      time.Duration
	RetentionPolicy   int
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
	Compressed  bool
	Encrypted   bool
	CrossRegion bool
	Error       error
}

// BackupService handles database backup and restore operations
type BackupService struct {
	db     *sql.DB
	config *BackupConfig
	logger *zap.Logger
}

// NewBackupService creates a new backup service instance
func NewBackupService(db *sql.DB, config *BackupConfig, logger *zap.Logger) *BackupService {
	return &BackupService{
		db:     db,
		config: config,
		logger: logger,
	}
}

// validateExecutablePath ensures the executable path is safe and exists
func validateExecutablePath(execPath string) error {
	// Check if path is absolute
	if !filepath.IsAbs(execPath) {
		return fmt.Errorf("executable path must be absolute: %s", execPath)
	}

	// Validate path contains only safe characters
	safePathRegex := regexp.MustCompile(`^[a-zA-Z0-9/._-]+$`)
	if !safePathRegex.MatchString(execPath) {
		return fmt.Errorf("executable path contains unsafe characters: %s", execPath)
	}

	// Check if file exists and is executable
	if _, err := os.Stat(execPath); os.IsNotExist(err) {
		return fmt.Errorf("executable not found: %s", execPath)
	}

	return nil
}

// validateEnvironmentVariable validates environment variable values
func validateEnvironmentVariable(key, value string) error {
	// Check for potentially dangerous characters
	dangerousChars := []string{"`", "$(", "|", "&", ";", ">", "<"}
	for _, char := range dangerousChars {
		if strings.Contains(value, char) {
			return fmt.Errorf("environment variable %s contains dangerous character: %s", key, char)
		}
	}
	return nil
}

// secureExecCommand creates a secure command with proper validation
func (bs *BackupService) secureExecCommand(ctx context.Context, name string, args ...string) (*exec.Cmd, error) {
	// Validate executable path
	if err := validateExecutablePath(name); err != nil {
		return nil, fmt.Errorf("invalid executable path: %w", err)
	}

	// Validate arguments
	for i, arg := range args {
		if strings.Contains(arg, "|") || strings.Contains(arg, "&") || strings.Contains(arg, ";") {
			return nil, fmt.Errorf("argument %d contains dangerous characters: %s", i, arg)
		}
	}

	cmd := exec.CommandContext(ctx, name, args...)
	return cmd, nil
}

// CreateBackup creates a database backup
func (bs *BackupService) CreateBackup(ctx context.Context) (*BackupResult, error) {
	startTime := time.Now()
	backupID := fmt.Sprintf("backup_%s", startTime.Format("20060102_150405"))

	bs.logger.Info("Starting database backup", zap.String("backup_id", backupID))

	// Validate backup directory with secure permissions
	if err := os.MkdirAll(bs.config.BackupDir, 0750); err != nil {
		bs.logger.Error("Failed to create backup directory", zap.Error(err), zap.String("backup_id", backupID))
		return &BackupResult{
			BackupID:  backupID,
			StartTime: startTime,
			EndTime:   time.Now(),
			Success:   false,
			Error:     err,
		}, err
	}

	// Generate backup filename with validation
	filename := filepath.Join(bs.config.BackupDir, fmt.Sprintf("%s.sql", backupID))

	// Validate filename doesn't contain path traversal
	if !strings.HasPrefix(filepath.Clean(filename), filepath.Clean(bs.config.BackupDir)) {
		return nil, fmt.Errorf("invalid backup filename: path traversal detected")
	}

	// Validate environment variables
	envVars := map[string]string{
		"DB_HOST": os.Getenv("DB_HOST"),
		"DB_PORT": os.Getenv("DB_PORT"),
		"DB_USER": os.Getenv("DB_USER"),
		"DB_NAME": os.Getenv("DB_NAME"),
	}

	for key, value := range envVars {
		if err := validateEnvironmentVariable(key, value); err != nil {
			return nil, fmt.Errorf("invalid environment variable: %w", err)
		}
	}

	// Create backup using pg_dump with secure command execution
	cmd, err := bs.secureExecCommand(ctx, "/usr/bin/pg_dump",
		"--host="+envVars["DB_HOST"],
		"--port="+envVars["DB_PORT"],
		"--username="+envVars["DB_USER"],
		"--dbname="+envVars["DB_NAME"],
		"--verbose",
		"--no-password",
		"--file="+filename,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure command: %w", err)
	}

	// Set environment variables for authentication with validation
	dbPassword := os.Getenv("DB_PASSWORD")
	if err := validateEnvironmentVariable("DB_PASSWORD", dbPassword); err != nil {
		return nil, fmt.Errorf("invalid password environment variable: %w", err)
	}

	cmd.Env = append(os.Environ(), "PGPASSWORD="+dbPassword)

	// Execute backup command
	if err := cmd.Run(); err != nil {
		bs.logger.Error("Backup command failed", zap.Error(err), zap.String("backup_id", backupID))
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
		bs.logger.Warn("Failed to calculate checksum", zap.Error(err), zap.String("backup_id", backupID))
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

	// Log backup completion
	bs.logger.Info("Database backup completed successfully", 
		zap.String("backup_id", backupID),
		zap.String("filename", filename),
		zap.Int64("size", fileInfo.Size()),
		zap.Duration("duration", duration),
	)

	// Compress backup if enabled
	if bs.config.Compression {
		if err := bs.compressBackup(filename); err != nil {
			bs.logger.Warn("Failed to compress backup", zap.Error(err), zap.String("backup_id", backupID))
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
			bs.logger.Warn("Failed to encrypt backup", zap.Error(err), zap.String("backup_id", backupID))
		} else {
			result.Filename = result.Filename + ".enc"
			if fileInfo, err := os.Stat(result.Filename); err == nil {
				result.Size = fileInfo.Size()
			}
		}
	}

	// Upload to cross-region if enabled
	if bs.config.CrossRegion {
		if err := bs.uploadToCrossRegion(result.Filename, backupID); err != nil {
			bs.logger.Warn("Failed to upload backup to cross-region", zap.Error(err), zap.String("backup_id", backupID))
		}
	}

	// Store backup metadata
	if err := bs.storeBackupMetadata(ctx, result); err != nil {
		bs.logger.Error("Failed to store backup metadata", zap.Error(err), zap.String("backup_id", backupID))
	}

	return result, nil
}

// RestoreBackup restores a database from a backup
func (bs *BackupService) RestoreBackup(ctx context.Context, backupID string) error {
	bs.logger.Info("Starting database restore", zap.String("backup_id", backupID))

	// Get backup metadata
	metadata, err := bs.getBackupMetadata(ctx, backupID)
	if err != nil {
		return fmt.Errorf("failed to get backup metadata: %w", err)
	}

	// Check if backup file exists
	if _, err := os.Stat(metadata.Filename); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", metadata.Filename)
	}

	// Validate filename doesn't contain path traversal
	if !strings.HasPrefix(filepath.Clean(metadata.Filename), filepath.Clean(bs.config.BackupDir)) {
		return fmt.Errorf("invalid backup filename: path traversal detected")
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

	// Validate environment variables
	envVars := map[string]string{
		"DB_HOST": os.Getenv("DB_HOST"),
		"DB_PORT": os.Getenv("DB_PORT"),
		"DB_USER": os.Getenv("DB_USER"),
		"DB_NAME": os.Getenv("DB_NAME"),
	}

	for key, value := range envVars {
		if err := validateEnvironmentVariable(key, value); err != nil {
			return fmt.Errorf("invalid environment variable: %w", err)
		}
	}

	// Restore database using psql with secure command execution
	cmd, err := bs.secureExecCommand(ctx, "/usr/bin/psql",
		"--host="+envVars["DB_HOST"],
		"--port="+envVars["DB_PORT"],
		"--username="+envVars["DB_USER"],
		"--dbname="+envVars["DB_NAME"],
		"--verbose",
		"--no-password",
		"--file="+filename,
	)
	if err != nil {
		return fmt.Errorf("failed to create secure restore command: %w", err)
	}

	// Set environment variables for authentication with validation
	dbPassword := os.Getenv("DB_PASSWORD")
	if err := validateEnvironmentVariable("DB_PASSWORD", dbPassword); err != nil {
		return fmt.Errorf("invalid password environment variable: %w", err)
	}

	cmd.Env = append(os.Environ(), "PGPASSWORD="+dbPassword)

	// Execute restore command
	if err := cmd.Run(); err != nil {
		bs.logger.Error("Restore command failed", zap.Error(err), zap.String("backup_id", backupID))
		return err
	}

	bs.logger.Info("Database restore completed successfully", zap.String("backup_id", backupID))
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
	bs.logger.Info("Starting cleanup of old backups", zap.Int("retention_days", bs.config.RetentionPolicy))

	cutoffDate := time.Now().AddDate(0, 0, -bs.config.RetentionPolicy)

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
			bs.logger.Warn("Failed to scan backup record", zap.Error(err))
			continue
		}

		// Delete backup file
		if err := os.Remove(filename); err != nil {
			bs.logger.Warn("Failed to delete backup file", zap.Error(err), zap.String("filename", filename))
			continue
		}

		// Delete metadata
		if _, err := bs.db.ExecContext(ctx, "DELETE FROM backup_metadata WHERE backup_id = $1", backupID); err != nil {
			bs.logger.Warn("Failed to delete backup metadata", zap.Error(err), zap.String("backup_id", backupID))
			continue
		}

		deletedCount++
		bs.logger.Info("Deleted old backup", zap.String("backup_id", backupID), zap.String("filename", filename))
	}

	bs.logger.Info("Backup cleanup completed", zap.Int("deleted_count", deletedCount))
	return nil
}

// ValidateBackup validates the integrity of a backup
func (bs *BackupService) ValidateBackup(ctx context.Context, backupID string) error {
	bs.logger.Info("Validating backup", zap.String("backup_id", backupID))

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

	bs.logger.Info("Backup validation completed successfully", zap.String("backup_id", backupID))
	return nil
}

// Helper methods

func (bs *BackupService) calculateChecksum(filename string) (string, error) {
	// Use Go's crypto/sha256 instead of external command
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

func (bs *BackupService) compressBackup(filename string) error {
	cmd, err := bs.secureExecCommand(context.Background(), "/bin/gzip", filename)
	if err != nil {
		return fmt.Errorf("failed to create secure compression command: %w", err)
	}
	return cmd.Run()
}

func (bs *BackupService) decompressBackup(filename string) error {
	cmd, err := bs.secureExecCommand(context.Background(), "/bin/gunzip", filename)
	if err != nil {
		return fmt.Errorf("failed to create secure decompression command: %w", err)
	}
	return cmd.Run()
}

func (bs *BackupService) encryptBackup(filename string) error {
	// Validate encryption key
	if err := validateEnvironmentVariable("ENCRYPTION_KEY", bs.config.EncryptionKey); err != nil {
		return fmt.Errorf("invalid encryption key: %w", err)
	}

	cmd, err := bs.secureExecCommand(context.Background(), "/usr/bin/openssl",
		"enc", "-aes-256-cbc", "-salt", "-in", filename, "-out", filename+".enc", "-k", bs.config.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to create secure encryption command: %w", err)
	}
	return cmd.Run()
}

func (bs *BackupService) decryptBackup(filename string) error {
	// Validate encryption key
	if err := validateEnvironmentVariable("ENCRYPTION_KEY", bs.config.EncryptionKey); err != nil {
		return fmt.Errorf("invalid encryption key: %w", err)
	}

	cmd, err := bs.secureExecCommand(context.Background(), "/usr/bin/openssl",
		"enc", "-d", "-aes-256-cbc", "-in", filename, "-out", filename[:len(filename)-4], "-k", bs.config.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to create secure decryption command: %w", err)
	}
	return cmd.Run()
}

func (bs *BackupService) uploadToCrossRegion(filename, backupID string) error {
	// Validate bucket name
	if err := validateEnvironmentVariable("BUCKET_NAME", bs.config.CrossRegionBucket); err != nil {
		return fmt.Errorf("invalid bucket name: %w", err)
	}

	cmd, err := bs.secureExecCommand(context.Background(), "/usr/local/bin/aws",
		"s3", "cp", filename, fmt.Sprintf("s3://%s/backups/%s", bs.config.CrossRegionBucket, backupID))
	if err != nil {
		return fmt.Errorf("failed to create secure upload command: %w", err)
	}
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
