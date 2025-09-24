package testing

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// DefaultBackupTestConfig creates a default configuration for backup testing
func DefaultBackupTestConfig() *BackupTestConfig {
	return &BackupTestConfig{
		SupabaseURL:       getEnvOrDefault("SUPABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform"),
		TestDatabaseURL:   getEnvOrDefault("TEST_DATABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform_test"),
		BackupDirectory:   getEnvOrDefault("BACKUP_DIRECTORY", "/tmp/backup_recovery_test"),
		TestDataSize:      getEnvIntOrDefault("TEST_DATA_SIZE", 1000),
		RecoveryTimeout:   getEnvDurationOrDefault("RECOVERY_TIMEOUT", 10*time.Minute),
		ValidationRetries: getEnvIntOrDefault("VALIDATION_RETRIES", 3),
	}
}

// ProductionBackupTestConfig creates a configuration suitable for production testing
func ProductionBackupTestConfig() *BackupTestConfig {
	return &BackupTestConfig{
		SupabaseURL:       getEnvOrDefault("SUPABASE_URL", ""),
		TestDatabaseURL:   getEnvOrDefault("TEST_DATABASE_URL", ""),
		BackupDirectory:   getEnvOrDefault("BACKUP_DIRECTORY", "/var/backups/kyb_platform"),
		TestDataSize:      getEnvIntOrDefault("TEST_DATA_SIZE", 10000),
		RecoveryTimeout:   getEnvDurationOrDefault("RECOVERY_TIMEOUT", 30*time.Minute),
		ValidationRetries: getEnvIntOrDefault("VALIDATION_RETRIES", 5),
	}
}

// DevelopmentBackupTestConfig creates a configuration suitable for development testing
func DevelopmentBackupTestConfig() *BackupTestConfig {
	return &BackupTestConfig{
		SupabaseURL:       getEnvOrDefault("SUPABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform_dev"),
		TestDatabaseURL:   getEnvOrDefault("TEST_DATABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform_test"),
		BackupDirectory:   getEnvOrDefault("BACKUP_DIRECTORY", "/tmp/backup_recovery_dev"),
		TestDataSize:      getEnvIntOrDefault("TEST_DATA_SIZE", 100),
		RecoveryTimeout:   getEnvDurationOrDefault("RECOVERY_TIMEOUT", 5*time.Minute),
		ValidationRetries: getEnvIntOrDefault("VALIDATION_RETRIES", 2),
	}
}

// ValidateConfig validates the backup test configuration
func (config *BackupTestConfig) Validate() error {
	if config.SupabaseURL == "" {
		return fmt.Errorf("SupabaseURL is required")
	}

	if config.TestDatabaseURL == "" {
		return fmt.Errorf("TestDatabaseURL is required")
	}

	if config.BackupDirectory == "" {
		return fmt.Errorf("BackupDirectory is required")
	}

	if config.TestDataSize <= 0 {
		return fmt.Errorf("TestDataSize must be greater than 0")
	}

	if config.RecoveryTimeout <= 0 {
		return fmt.Errorf("RecoveryTimeout must be greater than 0")
	}

	if config.ValidationRetries <= 0 {
		return fmt.Errorf("ValidationRetries must be greater than 0")
	}

	return nil
}

// GetEnvironmentInfo returns information about the current environment
func GetEnvironmentInfo() map[string]interface{} {
	return map[string]interface{}{
		"supabase_url":        getEnvOrDefault("SUPABASE_URL", "not set"),
		"test_database_url":   getEnvOrDefault("TEST_DATABASE_URL", "not set"),
		"backup_directory":    getEnvOrDefault("BACKUP_DIRECTORY", "not set"),
		"test_data_size":      getEnvIntOrDefault("TEST_DATA_SIZE", 0),
		"recovery_timeout":    getEnvDurationOrDefault("RECOVERY_TIMEOUT", 0).String(),
		"validation_retries":  getEnvIntOrDefault("VALIDATION_RETRIES", 0),
		"pg_dump_available":   isCommandAvailable("pg_dump"),
		"psql_available":      isCommandAvailable("psql"),
		"backup_dir_writable": isDirectoryWritable(getEnvOrDefault("BACKUP_DIRECTORY", "/tmp")),
	}
}

// Helper functions for environment variable handling
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func isDirectoryWritable(path string) bool {
	// Try to create a temporary file in the directory
	tempFile := filepath.Join(path, "test_write_permission.tmp")
	file, err := os.Create(tempFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(tempFile)
	return true
}

// String returns a string representation of the configuration
func (config *BackupTestConfig) String() string {
	return fmt.Sprintf(`
Backup Test Configuration:
  Supabase URL: %s
  Test Database URL: %s
  Backup Directory: %s
  Test Data Size: %d
  Recovery Timeout: %v
  Validation Retries: %d
`,
		config.SupabaseURL,
		config.TestDatabaseURL,
		config.BackupDirectory,
		config.TestDataSize,
		config.RecoveryTimeout,
		config.ValidationRetries,
	)
}
