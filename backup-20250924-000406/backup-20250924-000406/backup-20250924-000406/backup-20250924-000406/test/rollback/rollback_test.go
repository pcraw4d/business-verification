package rollback_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRollbackSuite provides comprehensive testing for rollback procedures
type TestRollbackSuite struct {
	projectRoot string
	backupDir   string
	logDir      string
	db          *sql.DB
}

// SetupTestSuite initializes the test suite
func SetupTestSuite(t *testing.T) *TestRollbackSuite {
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		projectRoot = "."
	}

	suite := &TestRollbackSuite{
		projectRoot: projectRoot,
		backupDir:   filepath.Join(projectRoot, "backups"),
		logDir:      filepath.Join(projectRoot, "logs"),
	}

	// Create necessary directories
	err := os.MkdirAll(suite.backupDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(suite.logDir, 0755)
	require.NoError(t, err)

	// Setup test database connection
	suite.setupTestDatabase(t)

	return suite
}

// TeardownTestSuite cleans up the test suite
func (suite *TestRollbackSuite) TeardownTestSuite(t *testing.T) {
	if suite.db != nil {
		suite.db.Close()
	}
}

// setupTestDatabase establishes a test database connection
func (suite *TestRollbackSuite) setupTestDatabase(t *testing.T) {
	// Use test database configuration
	dbHost := os.Getenv("TEST_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("TEST_DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbName := os.Getenv("TEST_DB_NAME")
	if dbName == "" {
		dbName = "kyb_platform_test"
	}

	dbUser := os.Getenv("TEST_DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPassword := os.Getenv("TEST_DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "password"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	require.NoError(t, err)

	suite.db = db
}

// TestDatabaseRollbackScript tests the database rollback script
func TestDatabaseRollbackScript(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	tests := []struct {
		name          string
		rollbackType  string
		targetVersion string
		dryRun        bool
		expectedError bool
		description   string
	}{
		{
			name:          "database_rollback_dry_run",
			rollbackType:  "schema",
			targetVersion: "005",
			dryRun:        true,
			expectedError: false,
			description:   "Test database schema rollback dry run",
		},
		{
			name:          "database_rollback_list_targets",
			rollbackType:  "list",
			targetVersion: "",
			dryRun:        false,
			expectedError: false,
			description:   "Test listing available rollback targets",
		},
		{
			name:          "database_rollback_invalid_type",
			rollbackType:  "invalid",
			targetVersion: "",
			dryRun:        false,
			expectedError: true,
			description:   "Test database rollback with invalid type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "database-rollback.sh")

			// Build command arguments
			args := []string{scriptPath}
			if tt.dryRun {
				args = append(args, "--dry-run")
			}
			if tt.targetVersion != "" {
				args = append(args, "--target", tt.targetVersion)
			}
			args = append(args, tt.rollbackType)

			// Execute script
			cmd := exec.Command("bash", args...)
			cmd.Env = append(os.Environ(),
				"DB_HOST=localhost",
				"DB_PORT=5432",
				"DB_NAME=kyb_platform_test",
				"DB_USER=postgres",
				"DB_PASSWORD=password",
			)

			output, err := cmd.CombinedOutput()

			if tt.expectedError {
				assert.Error(t, err, "Expected error for test: %s", tt.description)
			} else {
				assert.NoError(t, err, "Unexpected error for test: %s\nOutput: %s", tt.description, string(output))
			}

			// Verify output contains expected content
			outputStr := string(output)
			if tt.rollbackType == "list" {
				assert.Contains(t, outputStr, "Available rollback targets", "Should list available targets")
			} else if tt.dryRun {
				assert.Contains(t, outputStr, "DRY RUN", "Should indicate dry run mode")
			}
		})
	}
}

// TestApplicationRollbackScript tests the application rollback script
func TestApplicationRollbackScript(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	tests := []struct {
		name          string
		rollbackType  string
		targetVersion string
		environment   string
		dryRun        bool
		expectedError bool
		description   string
	}{
		{
			name:          "application_rollback_dry_run",
			rollbackType:  "binary",
			targetVersion: "v1.2.3",
			environment:   "staging",
			dryRun:        true,
			expectedError: false,
			description:   "Test application binary rollback dry run",
		},
		{
			name:          "application_rollback_list_targets",
			rollbackType:  "list",
			targetVersion: "",
			environment:   "",
			dryRun:        false,
			expectedError: false,
			description:   "Test listing available application rollback targets",
		},
		{
			name:          "application_rollback_config",
			rollbackType:  "config",
			targetVersion: "v1.2.3",
			environment:   "production",
			dryRun:        true,
			expectedError: false,
			description:   "Test application configuration rollback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "application-rollback.sh")

			// Build command arguments
			args := []string{scriptPath}
			if tt.dryRun {
				args = append(args, "--dry-run")
			}
			if tt.targetVersion != "" {
				args = append(args, "--target", tt.targetVersion)
			}
			if tt.environment != "" {
				args = append(args, "--environment", tt.environment)
			}
			args = append(args, tt.rollbackType)

			// Execute script
			cmd := exec.Command("bash", args...)
			output, err := cmd.CombinedOutput()

			if tt.expectedError {
				assert.Error(t, err, "Expected error for test: %s", tt.description)
			} else {
				assert.NoError(t, err, "Unexpected error for test: %s\nOutput: %s", tt.description, string(output))
			}

			// Verify output contains expected content
			outputStr := string(output)
			if tt.rollbackType == "list" {
				assert.Contains(t, outputStr, "Available rollback targets", "Should list available targets")
			} else if tt.dryRun {
				assert.Contains(t, outputStr, "DRY RUN", "Should indicate dry run mode")
			}
		})
	}
}

// TestConfigurationRollbackScript tests the configuration rollback script
func TestConfigurationRollbackScript(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	tests := []struct {
		name          string
		rollbackType  string
		targetVersion string
		environment   string
		dryRun        bool
		expectedError bool
		description   string
	}{
		{
			name:          "configuration_rollback_dry_run",
			rollbackType:  "env",
			targetVersion: "v1.2.3",
			environment:   "production",
			dryRun:        true,
			expectedError: false,
			description:   "Test configuration environment rollback dry run",
		},
		{
			name:          "configuration_rollback_features",
			rollbackType:  "features",
			targetVersion: "v1.2.3",
			environment:   "staging",
			dryRun:        true,
			expectedError: false,
			description:   "Test configuration features rollback",
		},
		{
			name:          "configuration_rollback_list_targets",
			rollbackType:  "list",
			targetVersion: "",
			environment:   "",
			dryRun:        false,
			expectedError: false,
			description:   "Test listing available configuration rollback targets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "configuration-rollback.sh")

			// Build command arguments
			args := []string{scriptPath}
			if tt.dryRun {
				args = append(args, "--dry-run")
			}
			if tt.targetVersion != "" {
				args = append(args, "--target", tt.targetVersion)
			}
			if tt.environment != "" {
				args = append(args, "--environment", tt.environment)
			}
			args = append(args, tt.rollbackType)

			// Execute script
			cmd := exec.Command("bash", args...)
			output, err := cmd.CombinedOutput()

			if tt.expectedError {
				assert.Error(t, err, "Expected error for test: %s", tt.description)
			} else {
				assert.NoError(t, err, "Unexpected error for test: %s\nOutput: %s", tt.description, string(output))
			}

			// Verify output contains expected content
			outputStr := string(output)
			if tt.rollbackType == "list" {
				assert.Contains(t, outputStr, "Available rollback targets", "Should list available targets")
			} else if tt.dryRun {
				assert.Contains(t, outputStr, "DRY RUN", "Should indicate dry run mode")
			}
		})
	}
}

// TestRollbackIntegration tests integrated rollback scenarios
func TestRollbackIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	t.Run("full_rollback_scenario", func(t *testing.T) {
		// Test a complete rollback scenario involving multiple components

		// 1. Create test backup files
		suite.createTestBackups(t)

		// 2. Test database rollback
		scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "database-rollback.sh")
		cmd := exec.Command("bash", scriptPath, "--dry-run", "--backup", "schema")
		cmd.Env = append(os.Environ(),
			"DB_HOST=localhost",
			"DB_PORT=5432",
			"DB_NAME=kyb_platform_test",
			"DB_USER=postgres",
			"DB_PASSWORD=password",
		)

		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "Database rollback should succeed")
		assert.Contains(t, string(output), "DRY RUN", "Should be in dry run mode")

		// 3. Test application rollback
		scriptPath = filepath.Join(suite.projectRoot, "scripts", "rollback", "application-rollback.sh")
		cmd = exec.Command("bash", scriptPath, "--dry-run", "--environment", "staging", "binary")

		output, err = cmd.CombinedOutput()
		assert.NoError(t, err, "Application rollback should succeed")
		assert.Contains(t, string(output), "DRY RUN", "Should be in dry run mode")

		// 4. Test configuration rollback
		scriptPath = filepath.Join(suite.projectRoot, "scripts", "rollback", "configuration-rollback.sh")
		cmd = exec.Command("bash", scriptPath, "--dry-run", "--environment", "production", "full")

		output, err = cmd.CombinedOutput()
		assert.NoError(t, err, "Configuration rollback should succeed")
		assert.Contains(t, string(output), "DRY RUN", "Should be in dry run mode")
	})
}

// TestRollbackErrorHandling tests error handling in rollback procedures
func TestRollbackErrorHandling(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	tests := []struct {
		name          string
		script        string
		args          []string
		expectedError bool
		description   string
	}{
		{
			name:          "missing_rollback_type",
			script:        "database-rollback.sh",
			args:          []string{},
			expectedError: true,
			description:   "Should fail when rollback type is missing",
		},
		{
			name:          "invalid_rollback_type",
			script:        "database-rollback.sh",
			args:          []string{"invalid_type"},
			expectedError: true,
			description:   "Should fail with invalid rollback type",
		},
		{
			name:          "missing_target_version",
			script:        "database-rollback.sh",
			args:          []string{"data"},
			expectedError: true,
			description:   "Should fail when target version is missing for data rollback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", tt.script)

			args := []string{scriptPath}
			args = append(args, tt.args...)

			cmd := exec.Command("bash", args...)
			output, err := cmd.CombinedOutput()

			if tt.expectedError {
				assert.Error(t, err, "Expected error for test: %s", tt.description)
			} else {
				assert.NoError(t, err, "Unexpected error for test: %s\nOutput: %s", tt.description, string(output))
			}
		})
	}
}

// createTestBackups creates test backup files for rollback testing
func (suite *TestRollbackSuite) createTestBackups(t *testing.T) {
	// Create test backup directory structure
	testBackupDir := filepath.Join(suite.backupDir, "test")
	err := os.MkdirAll(testBackupDir, 0755)
	require.NoError(t, err)

	// Create test database backup
	dbBackup := filepath.Join(testBackupDir, "database-backup-v1.2.3.sql")
	err = os.WriteFile(dbBackup, []byte("-- Test database backup\nSELECT 1;"), 0644)
	require.NoError(t, err)

	// Create test application backup
	appBackup := filepath.Join(testBackupDir, "app-backup-v1.2.3.tar.gz")
	err = os.WriteFile(appBackup, []byte("test backup content"), 0644)
	require.NoError(t, err)

	// Create test configuration backup
	configBackup := filepath.Join(testBackupDir, "config-backup-v1.2.3.tar.gz")
	err = os.WriteFile(configBackup, []byte("test config backup content"), 0644)
	require.NoError(t, err)
}

// TestRollbackPerformance tests rollback script performance
func TestRollbackPerformance(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	t.Run("rollback_script_performance", func(t *testing.T) {
		scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "database-rollback.sh")

		// Measure execution time
		start := time.Now()
		cmd := exec.Command("bash", scriptPath, "--dry-run", "schema")
		output, err := cmd.CombinedOutput()
		duration := time.Since(start)

		assert.NoError(t, err, "Rollback script should execute successfully")
		assert.Less(t, duration, 5*time.Second, "Rollback script should complete within 5 seconds")

		// Verify output is reasonable
		outputStr := string(output)
		assert.Contains(t, outputStr, "Starting database rollback process", "Should show start message")
		assert.Contains(t, outputStr, "completed successfully", "Should show completion message")
	})
}

// TestRollbackLogging tests rollback script logging functionality
func TestRollbackLogging(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	t.Run("rollback_logging", func(t *testing.T) {
		scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "database-rollback.sh")

		cmd := exec.Command("bash", scriptPath, "--dry-run", "schema")
		_, err := cmd.CombinedOutput()

		assert.NoError(t, err, "Rollback script should execute successfully")

		// Check that log file was created
		logFiles, err := filepath.Glob(filepath.Join(suite.logDir, "rollback-*.log"))
		require.NoError(t, err)
		assert.Greater(t, len(logFiles), 0, "Should create log files")

		// Check log file content
		if len(logFiles) > 0 {
			logContent, err := os.ReadFile(logFiles[0])
			require.NoError(t, err)

			logStr := string(logContent)
			assert.Contains(t, logStr, "[INFO]", "Should contain info logs")
			assert.Contains(t, logStr, "Starting database rollback process", "Should log start message")
		}
	})
}
