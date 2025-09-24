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

// IntegrationTestSuite provides comprehensive integration testing for rollback procedures
type IntegrationTestSuite struct {
	projectRoot string
	backupDir   string
	logDir      string
	configDir   string
	db          *sql.DB
}

// SetupIntegrationTestSuite initializes the integration test suite
func SetupIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		projectRoot = "."
	}

	suite := &IntegrationTestSuite{
		projectRoot: projectRoot,
		backupDir:   filepath.Join(projectRoot, "backups"),
		logDir:      filepath.Join(projectRoot, "logs"),
		configDir:   filepath.Join(projectRoot, "configs"),
	}

	// Create necessary directories
	err := os.MkdirAll(suite.backupDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(suite.logDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(suite.configDir, 0755)
	require.NoError(t, err)

	// Setup test database connection
	suite.setupTestDatabase(t)

	// Create test data and configurations
	suite.setupTestData(t)

	return suite
}

// TeardownIntegrationTestSuite cleans up the integration test suite
func (suite *IntegrationTestSuite) TeardownIntegrationTestSuite(t *testing.T) {
	if suite.db != nil {
		suite.db.Close()
	}
}

// setupTestDatabase establishes a test database connection
func (suite *IntegrationTestSuite) setupTestDatabase(t *testing.T) {
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

// setupTestData creates test data and configurations for rollback testing
func (suite *IntegrationTestSuite) setupTestData(t *testing.T) {
	// Create test configuration files
	suite.createTestConfigurations(t)

	// Create test database schema
	suite.createTestDatabaseSchema(t)

	// Create test backup files
	suite.createTestBackupFiles(t)
}

// createTestConfigurations creates test configuration files
func (suite *IntegrationTestSuite) createTestConfigurations(t *testing.T) {
	// Create database configuration
	dbConfig := `database:
  host: localhost
  port: 5432
  name: kyb_platform_test
  user: postgres
  password: password
  ssl_mode: disable
  max_connections: 100
  connection_timeout: 30s
`
	err := os.WriteFile(filepath.Join(suite.configDir, "database.yaml"), []byte(dbConfig), 0644)
	require.NoError(t, err)

	// Create API configuration
	apiConfig := `api:
  host: localhost
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  max_header_bytes: 1048576
  cors:
    allowed_origins: ["*"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE"]
    allowed_headers: ["*"]
`
	err = os.WriteFile(filepath.Join(suite.configDir, "api.yaml"), []byte(apiConfig), 0644)
	require.NoError(t, err)

	// Create security configuration
	securityConfig := `security:
  jwt:
    secret: test-secret-key
    expiration: 24h
  rate_limiting:
    requests_per_minute: 100
    burst: 10
  encryption:
    algorithm: AES-256-GCM
    key: test-encryption-key
`
	err = os.WriteFile(filepath.Join(suite.configDir, "security.yaml"), []byte(securityConfig), 0644)
	require.NoError(t, err)

	// Create features configuration
	featuresConfig := `{
  "features": {
    "merchant_portfolio": true,
    "bulk_operations": true,
    "merchant_comparison": true,
    "audit_logging": true,
    "compliance_tracking": true,
    "real_time_updates": false,
    "advanced_analytics": false,
    "external_integrations": false
  },
  "version": "1.2.3",
  "last_updated": "2025-01-19T10:00:00Z"
}`
	err = os.WriteFile(filepath.Join(suite.configDir, "features.json"), []byte(featuresConfig), 0644)
	require.NoError(t, err)

	// Create environment file
	envConfig := `# Test Environment Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=kyb_platform_test
DB_USER=postgres
DB_PASSWORD=password
API_HOST=localhost
API_PORT=8080
LOG_LEVEL=info
ENVIRONMENT=test
`
	err = os.WriteFile(filepath.Join(suite.projectRoot, ".env"), []byte(envConfig), 0644)
	require.NoError(t, err)
}

// createTestDatabaseSchema creates test database schema
func (suite *IntegrationTestSuite) createTestDatabaseSchema(t *testing.T) {
	// Create test tables
	schema := `
CREATE TABLE IF NOT EXISTS test_merchants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS test_audit_logs (
    id SERIAL PRIMARY KEY,
    merchant_id INTEGER REFERENCES test_merchants(id),
    action VARCHAR(100) NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO test_merchants (name, email) VALUES 
    ('Test Merchant 1', 'test1@example.com'),
    ('Test Merchant 2', 'test2@example.com'),
    ('Test Merchant 3', 'test3@example.com');
`

	_, err := suite.db.Exec(schema)
	require.NoError(t, err)
}

// createTestBackupFiles creates test backup files
func (suite *IntegrationTestSuite) createTestBackupFiles(t *testing.T) {
	// Create database backup
	dbBackup := `-- Test Database Backup
-- Generated: 2025-01-19 10:00:00

CREATE TABLE IF NOT EXISTS test_merchants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO test_merchants (name, email) VALUES 
    ('Backup Merchant 1', 'backup1@example.com'),
    ('Backup Merchant 2', 'backup2@example.com');
`

	err := os.WriteFile(filepath.Join(suite.backupDir, "database-backup-v1.2.3.sql"), []byte(dbBackup), 0644)
	require.NoError(t, err)

	// Create application backup (simplified)
	appBackup := `test application backup content`
	err = os.WriteFile(filepath.Join(suite.backupDir, "app-backup-v1.2.3.tar.gz"), []byte(appBackup), 0644)
	require.NoError(t, err)

	// Create configuration backup
	configBackup := `test configuration backup content`
	err = os.WriteFile(filepath.Join(suite.backupDir, "config-backup-v1.2.3.tar.gz"), []byte(configBackup), 0644)
	require.NoError(t, err)
}

// TestFullRollbackScenario tests a complete rollback scenario
func TestFullRollbackScenario(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)
	defer suite.TeardownIntegrationTestSuite(t)

	t.Run("complete_rollback_workflow", func(t *testing.T) {
		// Step 1: Test database rollback
		t.Run("database_rollback", func(t *testing.T) {
			scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "database-rollback.sh")

			cmd := exec.Command("bash", scriptPath, "--dry-run", "--backup", "--target", "v1.2.3", "full")
			cmd.Env = append(os.Environ(),
				"DB_HOST=localhost",
				"DB_PORT=5432",
				"DB_NAME=kyb_platform_test",
				"DB_USER=postgres",
				"DB_PASSWORD=password",
			)

			output, err := cmd.CombinedOutput()
			assert.NoError(t, err, "Database rollback should succeed")

			outputStr := string(output)
			assert.Contains(t, outputStr, "DRY RUN", "Should be in dry run mode")
			assert.Contains(t, outputStr, "Starting database rollback process", "Should show start message")
			assert.Contains(t, outputStr, "completed successfully", "Should show completion message")
		})

		// Step 2: Test application rollback
		t.Run("application_rollback", func(t *testing.T) {
			scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "application-rollback.sh")

			cmd := exec.Command("bash", scriptPath, "--dry-run", "--backup", "--target", "v1.2.3", "--environment", "staging", "full")

			output, err := cmd.CombinedOutput()
			assert.NoError(t, err, "Application rollback should succeed")

			outputStr := string(output)
			assert.Contains(t, outputStr, "DRY RUN", "Should be in dry run mode")
			assert.Contains(t, outputStr, "Starting application rollback process", "Should show start message")
			assert.Contains(t, outputStr, "completed successfully", "Should show completion message")
		})

		// Step 3: Test configuration rollback
		t.Run("configuration_rollback", func(t *testing.T) {
			scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "configuration-rollback.sh")

			cmd := exec.Command("bash", scriptPath, "--dry-run", "--backup", "--target", "v1.2.3", "--environment", "production", "full")

			output, err := cmd.CombinedOutput()
			assert.NoError(t, err, "Configuration rollback should succeed")

			outputStr := string(output)
			assert.Contains(t, outputStr, "DRY RUN", "Should be in dry run mode")
			assert.Contains(t, outputStr, "Starting configuration rollback process", "Should show start message")
			assert.Contains(t, outputStr, "completed successfully", "Should show completion message")
		})
	})
}

// TestRollbackWithRealData tests rollback procedures with real data
func TestRollbackWithRealData(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)
	defer suite.TeardownIntegrationTestSuite(t)

	t.Run("rollback_with_real_data", func(t *testing.T) {
		// Insert test data
		_, err := suite.db.Exec("INSERT INTO test_merchants (name, email) VALUES ($1, $2)", "Real Test Merchant", "real@example.com")
		require.NoError(t, err)

		// Verify data exists
		var count int
		err = suite.db.QueryRow("SELECT COUNT(*) FROM test_merchants").Scan(&count)
		require.NoError(t, err)
		assert.Greater(t, count, 0, "Should have test data")

		// Test database rollback (dry run)
		scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "database-rollback.sh")

		cmd := exec.Command("bash", scriptPath, "--dry-run", "--target", "v1.2.3", "data")
		cmd.Env = append(os.Environ(),
			"DB_HOST=localhost",
			"DB_PORT=5432",
			"DB_NAME=kyb_platform_test",
			"DB_USER=postgres",
			"DB_PASSWORD=password",
		)

		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "Database rollback with real data should succeed")

		outputStr := string(output)
		assert.Contains(t, outputStr, "DRY RUN", "Should be in dry run mode")
	})
}

// TestRollbackErrorRecovery tests error recovery scenarios
func TestRollbackErrorRecovery(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)
	defer suite.TeardownIntegrationTestSuite(t)

	t.Run("rollback_error_recovery", func(t *testing.T) {
		// Test with invalid database connection
		scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "database-rollback.sh")

		cmd := exec.Command("bash", scriptPath, "--dry-run", "--target", "v1.2.3", "data")
		cmd.Env = append(os.Environ(),
			"DB_HOST=invalid-host",
			"DB_PORT=5432",
			"DB_NAME=kyb_platform_test",
			"DB_USER=postgres",
			"DB_PASSWORD=password",
		)

		output, err := cmd.CombinedOutput()
		// Should still succeed in dry run mode even with invalid connection
		assert.NoError(t, err, "Dry run should succeed even with invalid connection")

		outputStr := string(output)
		assert.Contains(t, outputStr, "DRY RUN", "Should be in dry run mode")
	})
}

// TestRollbackLoggingIntegration tests logging integration
func TestRollbackLoggingIntegration(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)
	defer suite.TeardownIntegrationTestSuite(t)

	t.Run("rollback_logging_integration", func(t *testing.T) {
		scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "database-rollback.sh")

		cmd := exec.Command("bash", scriptPath, "--dry-run", "schema")
		_, err := cmd.CombinedOutput()

		assert.NoError(t, err, "Rollback script should execute successfully")

		// Check that log files were created
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
			assert.Contains(t, logStr, "completed successfully", "Should log completion message")
		}
	})
}

// TestRollbackPerformanceIntegration tests performance of rollback procedures
func TestRollbackPerformanceIntegration(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)
	defer suite.TeardownIntegrationTestSuite(t)

	t.Run("rollback_performance", func(t *testing.T) {
		scripts := []string{
			"database-rollback.sh",
			"application-rollback.sh",
			"configuration-rollback.sh",
		}

		for _, script := range scripts {
			t.Run(script, func(t *testing.T) {
				scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", script)

				// Measure execution time
				start := time.Now()
				cmd := exec.Command("bash", scriptPath, "--dry-run", "schema")
				output, err := cmd.CombinedOutput()
				duration := time.Since(start)

				assert.NoError(t, err, "Rollback script should execute successfully")
				assert.Less(t, duration, 10*time.Second, "Rollback script should complete within 10 seconds")

				// Verify output is reasonable
				outputStr := string(output)
				assert.Contains(t, outputStr, "Starting", "Should show start message")
				assert.Contains(t, outputStr, "completed successfully", "Should show completion message")
			})
		}
	})
}

// TestRollbackConcurrency tests rollback procedures under concurrent access
func TestRollbackConcurrency(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)
	defer suite.TeardownIntegrationTestSuite(t)

	t.Run("rollback_concurrency", func(t *testing.T) {
		// Run multiple rollback scripts concurrently
		scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "database-rollback.sh")

		// Start multiple concurrent rollback operations
		done := make(chan error, 3)

		for i := 0; i < 3; i++ {
			go func(id int) {
				cmd := exec.Command("bash", scriptPath, "--dry-run", "schema")
				cmd.Env = append(os.Environ(),
					"DB_HOST=localhost",
					"DB_PORT=5432",
					"DB_NAME=kyb_platform_test",
					"DB_USER=postgres",
					"DB_PASSWORD=password",
				)

				_, err := cmd.CombinedOutput()
				done <- err
			}(i)
		}

		// Wait for all operations to complete
		for i := 0; i < 3; i++ {
			err := <-done
			assert.NoError(t, err, "Concurrent rollback operation should succeed")
		}
	})
}
