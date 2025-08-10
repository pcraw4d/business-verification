package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// BackupConfig represents backup configuration
type BackupConfig struct {
	BackupDir      string        `json:"backup_dir"`
	RetentionDays  int           `json:"retention_days"`
	CompressBackup bool          `json:"compress_backup"`
	BackupInterval time.Duration `json:"backup_interval"`
}

// BackupInfo represents backup information
type BackupInfo struct {
	ID        string    `json:"id"`
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	Checksum  string    `json:"checksum,omitempty"`
}

// BackupSystem handles database backups
type BackupSystem struct {
	db           *sql.DB
	config       *DatabaseConfig
	backupConfig *BackupConfig
}

// NewBackupSystem creates a new backup system
func NewBackupSystem(db *sql.DB, config *DatabaseConfig, backupConfig *BackupConfig) *BackupSystem {
	return &BackupSystem{
		db:           db,
		config:       config,
		backupConfig: backupConfig,
	}
}

// CreateBackup creates a database backup
func (b *BackupSystem) CreateBackup(ctx context.Context) (*BackupInfo, error) {
	// Generate backup filename
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupID := fmt.Sprintf("backup_%s", timestamp)
	filename := fmt.Sprintf("%s.sql", backupID)
	filepath := filepath.Join(b.backupConfig.BackupDir, filename)

	// Ensure backup directory exists
	if err := os.MkdirAll(b.backupConfig.BackupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create backup info
	backupInfo := &BackupInfo{
		ID:        backupID,
		Filename:  filename,
		CreatedAt: time.Now(),
		Status:    "in_progress",
	}

	// Execute pg_dump command
	cmd := exec.CommandContext(ctx, "pg_dump",
		"-h", b.config.Host,
		"-p", fmt.Sprintf("%d", b.config.Port),
		"-U", b.config.Username,
		"-d", b.config.Database,
		"-f", filepath,
		"--no-password", // Use environment variable for password
	)

	// Set environment variables
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", b.config.Password))

	// Execute backup
	if err := cmd.Run(); err != nil {
		backupInfo.Status = "failed"
		backupInfo.Error = err.Error()
		return backupInfo, fmt.Errorf("failed to create backup: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		backupInfo.Status = "failed"
		backupInfo.Error = err.Error()
		return backupInfo, fmt.Errorf("failed to get backup file info: %w", err)
	}

	backupInfo.Size = fileInfo.Size()
	backupInfo.Status = "completed"

	// Compress backup if configured
	if b.backupConfig.CompressBackup {
		if err := b.compressBackup(filepath); err != nil {
			log.Printf("Warning: failed to compress backup %s: %v", backupID, err)
		}
	}

	log.Printf("Backup created successfully: %s (%d bytes)", backupID, backupInfo.Size)
	return backupInfo, nil
}

// RestoreBackup restores a database from backup
func (b *BackupSystem) RestoreBackup(ctx context.Context, backupID string) error {
	// Find backup file
	backupFile := b.findBackupFile(backupID)
	if backupFile == "" {
		return fmt.Errorf("backup file not found for ID: %s", backupID)
	}

	// Check if file exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupFile)
	}

	// Execute psql command to restore
	cmd := exec.CommandContext(ctx, "psql",
		"-h", b.config.Host,
		"-p", fmt.Sprintf("%d", b.config.Port),
		"-U", b.config.Username,
		"-d", b.config.Database,
		"-f", backupFile,
		"--no-password", // Use environment variable for password
	)

	// Set environment variables
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", b.config.Password))

	// Execute restore
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	log.Printf("Backup restored successfully: %s", backupID)
	return nil
}

// ListBackups lists all available backups
func (b *BackupSystem) ListBackups() ([]*BackupInfo, error) {
	var backups []*BackupInfo

	// Read backup directory
	files, err := os.ReadDir(b.backupConfig.BackupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return backups, nil // No backups directory yet
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Parse backup info from filename
		backupInfo, err := b.parseBackupInfo(file.Name())
		if err != nil {
			log.Printf("Warning: failed to parse backup info for %s: %v", file.Name(), err)
			continue
		}

		// Get file info
		fileInfo, err := file.Info()
		if err != nil {
			log.Printf("Warning: failed to get file info for %s: %v", file.Name(), err)
			continue
		}

		backupInfo.Size = fileInfo.Size()
		backupInfo.CreatedAt = fileInfo.ModTime()
		backupInfo.Status = "completed"

		backups = append(backups, backupInfo)
	}

	return backups, nil
}

// DeleteBackup deletes a specific backup
func (b *BackupSystem) DeleteBackup(backupID string) error {
	backupFile := b.findBackupFile(backupID)
	if backupFile == "" {
		return fmt.Errorf("backup file not found for ID: %s", backupID)
	}

	if err := os.Remove(backupFile); err != nil {
		return fmt.Errorf("failed to delete backup file: %w", err)
	}

	log.Printf("Backup deleted successfully: %s", backupID)
	return nil
}

// CleanupOldBackups removes backups older than retention period
func (b *BackupSystem) CleanupOldBackups() error {
	backups, err := b.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	cutoffTime := time.Now().AddDate(0, 0, -b.backupConfig.RetentionDays)
	deletedCount := 0

	for _, backup := range backups {
		if backup.CreatedAt.Before(cutoffTime) {
			if err := b.DeleteBackup(backup.ID); err != nil {
				log.Printf("Warning: failed to delete old backup %s: %v", backup.ID, err)
			} else {
				deletedCount++
			}
		}
	}

	if deletedCount > 0 {
		log.Printf("Cleaned up %d old backups", deletedCount)
	}

	return nil
}

// GetBackupStatus returns the status of a specific backup
func (b *BackupSystem) GetBackupStatus(backupID string) (*BackupInfo, error) {
	backups, err := b.ListBackups()
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	for _, backup := range backups {
		if backup.ID == backupID {
			return backup, nil
		}
	}

	return nil, fmt.Errorf("backup not found: %s", backupID)
}

// findBackupFile finds the backup file for a given backup ID
func (b *BackupSystem) findBackupFile(backupID string) string {
	// Try different file extensions
	extensions := []string{".sql", ".sql.gz", ".sql.bz2"}

	for _, ext := range extensions {
		filename := backupID + ext
		filepath := filepath.Join(b.backupConfig.BackupDir, filename)

		if _, err := os.Stat(filepath); err == nil {
			return filepath
		}
	}

	return ""
}

// parseBackupInfo parses backup information from filename
func (b *BackupSystem) parseBackupInfo(filename string) (*BackupInfo, error) {
	// Expected format: backup_YYYY-MM-DD_HH-MM-SS.sql
	if len(filename) < 20 {
		return nil, fmt.Errorf("invalid backup filename format: %s", filename)
	}

	// Extract backup ID (remove extension)
	backupID := filename
	if ext := filepath.Ext(filename); ext != "" {
		backupID = filename[:len(filename)-len(ext)]
	}

	// Validate backup ID format
	if len(backupID) < 20 {
		return nil, fmt.Errorf("invalid backup ID format: %s", backupID)
	}

	return &BackupInfo{
		ID:       backupID,
		Filename: filename,
		Status:   "unknown",
	}, nil
}

// compressBackup compresses a backup file
func (b *BackupSystem) compressBackup(filepath string) error {
	// Use gzip to compress the backup file
	cmd := exec.Command("gzip", filepath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to compress backup: %w", err)
	}

	return nil
}

// decompressBackup decompresses a backup file
func (b *BackupSystem) decompressBackup(filepath string) error {
	// Use gunzip to decompress the backup file
	cmd := exec.Command("gunzip", filepath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to decompress backup: %w", err)
	}

	return nil
}

// ValidateBackup validates a backup file
func (b *BackupSystem) ValidateBackup(backupID string) error {
	backupFile := b.findBackupFile(backupID)
	if backupFile == "" {
		return fmt.Errorf("backup file not found for ID: %s", backupID)
	}

	// Check if file is compressed
	isCompressed := filepath.Ext(backupFile) == ".gz"

	// For compressed files, we can only check if they can be decompressed
	if isCompressed {
		cmd := exec.Command("gunzip", "-t", backupFile)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("backup file is corrupted or invalid: %w", err)
		}
	} else {
		// For uncompressed files, we can try to parse the SQL
		// This is a basic validation - in production you might want more sophisticated validation
		file, err := os.Open(backupFile)
		if err != nil {
			return fmt.Errorf("failed to open backup file: %w", err)
		}
		defer file.Close()

		// Read first few bytes to check if it looks like SQL
		buffer := make([]byte, 100)
		n, err := file.Read(buffer)
		if err != nil {
			return fmt.Errorf("failed to read backup file: %w", err)
		}

		content := string(buffer[:n])
		if len(content) < 10 {
			return fmt.Errorf("backup file appears to be empty or too small")
		}
	}

	return nil
}

// GetBackupStats returns backup statistics
func (b *BackupSystem) GetBackupStats() (map[string]interface{}, error) {
	backups, err := b.ListBackups()
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	totalSize := int64(0)
	oldestBackup := time.Now()
	newestBackup := time.Time{}

	for _, backup := range backups {
		totalSize += backup.Size

		if backup.CreatedAt.Before(oldestBackup) {
			oldestBackup = backup.CreatedAt
		}

		if backup.CreatedAt.After(newestBackup) {
			newestBackup = backup.CreatedAt
		}
	}

	stats := map[string]interface{}{
		"total_backups":    len(backups),
		"total_size_bytes": totalSize,
		"total_size_mb":    float64(totalSize) / (1024 * 1024),
		"oldest_backup":    oldestBackup,
		"newest_backup":    newestBackup,
		"retention_days":   b.backupConfig.RetentionDays,
		"backup_dir":       b.backupConfig.BackupDir,
	}

	return stats, nil
}
