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

// E2ETestSuite provides end-to-end testing for rollback procedures
type E2ETestSuite struct {
	projectRoot string
	backupDir   string
	logDir      string
	configDir   string
	db          *sql.DB
}

// SetupE2ETestSuite initializes the end-to-end test suite
func SetupE2ETestSuite(t *testing.T) *E2ETestSuite {
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		projectRoot = "."
	}

	suite := &E2ETestSuite{
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

	// Create comprehensive test environment
	suite.setupE2ETestEnvironment(t)

	return suite
}

// TeardownE2ETestSuite cleans up the end-to-end test suite
func (suite *E2ETestSuite) TeardownE2ETestSuite(t *testing.T) {
	if suite.db != nil {
		suite.db.Close()
	}
}

// setupTestDatabase establishes a test database connection
func (suite *E2ETestSuite) setupTestDatabase(t *testing.T) {
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

// setupE2ETestEnvironment creates a comprehensive test environment
func (suite *E2ETestSuite) setupE2ETestEnvironment(t *testing.T) {
	// Create production-like configuration
	suite.createProductionConfigurations(t)

	// Create comprehensive database schema
	suite.createComprehensiveDatabaseSchema(t)

	// Create realistic backup files
	suite.createRealisticBackupFiles(t)

	// Create application binaries
	suite.createTestApplicationBinaries(t)
}

// createProductionConfigurations creates production-like configuration files
func (suite *E2ETestSuite) createProductionConfigurations(t *testing.T) {
	// Create production database configuration
	dbConfig := `database:
  host: localhost
  port: 5432
  name: kyb_platform
  user: kyb_user
  password: secure_password
  ssl_mode: require
  max_connections: 200
  connection_timeout: 30s
  idle_timeout: 300s
  max_lifetime: 3600s
  pool_size: 25
  retry_attempts: 3
  retry_delay: 5s
`
	err := os.WriteFile(filepath.Join(suite.configDir, "database.yaml"), []byte(dbConfig), 0644)
	require.NoError(t, err)

	// Create production API configuration
	apiConfig := `api:
  host: 0.0.0.0
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  max_header_bytes: 1048576
  tls:
    enabled: true
    cert_file: /etc/ssl/certs/kyb-platform.crt
    key_file: /etc/ssl/private/kyb-platform.key
  cors:
    allowed_origins: ["https://kyb-platform.com", "https://app.kyb-platform.com"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allowed_headers: ["Content-Type", "Authorization", "X-Requested-With"]
    allow_credentials: true
  rate_limiting:
    enabled: true
    requests_per_minute: 1000
    burst: 100
`
	err = os.WriteFile(filepath.Join(suite.configDir, "api.yaml"), []byte(apiConfig), 0644)
	require.NoError(t, err)

	// Create production security configuration
	securityConfig := `security:
  jwt:
    secret: production-jwt-secret-key-very-long-and-secure
    expiration: 24h
    refresh_expiration: 168h
    issuer: kyb-platform
    audience: kyb-platform-users
  rate_limiting:
    enabled: true
    requests_per_minute: 100
    burst: 10
    skip_successful_requests: false
  encryption:
    algorithm: AES-256-GCM
    key: production-encryption-key-very-long-and-secure
  password_policy:
    min_length: 12
    require_uppercase: true
    require_lowercase: true
    require_numbers: true
    require_special_chars: true
    max_age_days: 90
  session:
    timeout: 30m
    max_concurrent: 5
    secure_cookies: true
`
	err = os.WriteFile(filepath.Join(suite.configDir, "security.yaml"), []byte(securityConfig), 0644)
	require.NoError(t, err)

	// Create production features configuration
	featuresConfig := `{
  "features": {
    "merchant_portfolio": true,
    "bulk_operations": true,
    "merchant_comparison": true,
    "audit_logging": true,
    "compliance_tracking": true,
    "real_time_updates": true,
    "advanced_analytics": true,
    "external_integrations": true,
    "api_rate_limiting": true,
    "data_encryption": true,
    "session_management": true,
    "multi_tenant_support": false,
    "advanced_reporting": false,
    "machine_learning": false
  },
  "version": "1.2.3",
  "last_updated": "2025-01-19T10:00:00Z",
  "environment": "production",
  "maintenance_mode": false
}`
	err = os.WriteFile(filepath.Join(suite.configDir, "features.json"), []byte(featuresConfig), 0644)
	require.NoError(t, err)

	// Create production environment file
	envConfig := `# Production Environment Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=kyb_platform
DB_USER=kyb_user
DB_PASSWORD=secure_password
API_HOST=0.0.0.0
API_PORT=8080
LOG_LEVEL=info
ENVIRONMENT=production
TLS_ENABLED=true
RATE_LIMITING_ENABLED=true
ENCRYPTION_ENABLED=true
AUDIT_LOGGING_ENABLED=true
COMPLIANCE_TRACKING_ENABLED=true
`
	err = os.WriteFile(filepath.Join(suite.projectRoot, ".env"), []byte(envConfig), 0644)
	require.NoError(t, err)
}

// createComprehensiveDatabaseSchema creates a comprehensive database schema
func (suite *E2ETestSuite) createComprehensiveDatabaseSchema(t *testing.T) {
	// Create comprehensive schema
	schema := `
-- Create merchants table
CREATE TABLE IF NOT EXISTS merchants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(50),
    address TEXT,
    website VARCHAR(255),
    industry VARCHAR(100),
    portfolio_type VARCHAR(50) DEFAULT 'prospective',
    risk_level VARCHAR(20) DEFAULT 'medium',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create audit_logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    merchant_id INTEGER REFERENCES merchants(id),
    user_id INTEGER,
    action VARCHAR(100) NOT NULL,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create compliance_records table
CREATE TABLE IF NOT EXISTS compliance_records (
    id SERIAL PRIMARY KEY,
    merchant_id INTEGER REFERENCES merchants(id),
    compliance_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    score DECIMAL(5,2),
    details JSONB,
    checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP
);

-- Create user_sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    merchant_id INTEGER REFERENCES merchants(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert test data
INSERT INTO merchants (name, email, phone, address, website, industry, portfolio_type, risk_level) VALUES 
    ('Acme Corporation', 'contact@acme.com', '+1-555-123-4567', '123 Main St, Anytown, ST 12345', 'https://acme.com', 'Technology', 'onboarded', 'low'),
    ('Beta Industries', 'info@beta.com', '+1-555-234-5678', '456 Oak Ave, Somewhere, ST 67890', 'https://beta.com', 'Manufacturing', 'onboarded', 'medium'),
    ('Gamma Services', 'hello@gamma.com', '+1-555-345-6789', '789 Pine Rd, Elsewhere, ST 54321', 'https://gamma.com', 'Services', 'prospective', 'high'),
    ('Delta Trading', 'sales@delta.com', '+1-555-456-7890', '321 Elm St, Nowhere, ST 98765', 'https://delta.com', 'Retail', 'pending', 'medium'),
    ('Epsilon Solutions', 'support@epsilon.com', '+1-555-567-8901', '654 Maple Dr, Anywhere, ST 13579', 'https://epsilon.com', 'Technology', 'deactivated', 'low');

-- Insert audit log data
INSERT INTO audit_logs (merchant_id, user_id, action, details, ip_address) VALUES 
    (1, 1, 'merchant_created', '{"name": "Acme Corporation", "email": "contact@acme.com"}', '192.168.1.100'),
    (2, 1, 'merchant_updated', '{"field": "risk_level", "old_value": "low", "new_value": "medium"}', '192.168.1.101'),
    (3, 2, 'compliance_check', '{"type": "kyc", "status": "passed"}', '192.168.1.102');

-- Insert compliance records
INSERT INTO compliance_records (merchant_id, compliance_type, status, score, details) VALUES 
    (1, 'kyc', 'passed', 95.5, '{"checks": ["identity", "address", "business_registration"]}'),
    (2, 'aml', 'passed', 88.0, '{"checks": ["sanctions", "pep", "adverse_media"]}'),
    (3, 'kyc', 'failed', 45.0, '{"checks": ["identity", "address"], "issues": ["invalid_address"]}');

-- Insert user sessions
INSERT INTO user_sessions (user_id, session_token, merchant_id, expires_at) VALUES 
    (1, 'session_token_123', 1, NOW() + INTERVAL '24 hours'),
    (2, 'session_token_456', 2, NOW() + INTERVAL '24 hours');
`

	_, err := suite.db.Exec(schema)
	require.NoError(t, err)
}

// createRealisticBackupFiles creates realistic backup files
func (suite *E2ETestSuite) createRealisticBackupFiles(t *testing.T) {
	// Create comprehensive database backup
	dbBackup := `-- KYB Platform Database Backup
-- Generated: 2025-01-19 10:00:00
-- Version: 1.2.3
-- Environment: production

-- Create merchants table
CREATE TABLE IF NOT EXISTS merchants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(50),
    address TEXT,
    website VARCHAR(255),
    industry VARCHAR(100),
    portfolio_type VARCHAR(50) DEFAULT 'prospective',
    risk_level VARCHAR(20) DEFAULT 'medium',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert backup data
INSERT INTO merchants (name, email, phone, address, website, industry, portfolio_type, risk_level) VALUES 
    ('Backup Corp 1', 'backup1@example.com', '+1-555-111-1111', '111 Backup St, Backup City, BC 11111', 'https://backup1.com', 'Technology', 'onboarded', 'low'),
    ('Backup Corp 2', 'backup2@example.com', '+1-555-222-2222', '222 Backup Ave, Backup Town, BT 22222', 'https://backup2.com', 'Manufacturing', 'onboarded', 'medium'),
    ('Backup Corp 3', 'backup3@example.com', '+1-555-333-3333', '333 Backup Rd, Backup Village, BV 33333', 'https://backup3.com', 'Services', 'prospective', 'high');
`

	err := os.WriteFile(filepath.Join(suite.backupDir, "database-backup-v1.2.3.sql"), []byte(dbBackup), 0644)
	require.NoError(t, err)

	// Create application backup
	appBackup := `test application backup content for version 1.2.3`
	err = os.WriteFile(filepath.Join(suite.backupDir, "app-backup-v1.2.3.tar.gz"), []byte(appBackup), 0644)
	require.NoError(t, err)

	// Create configuration backup
	configBackup := `test configuration backup content for version 1.2.3`
	err = os.WriteFile(filepath.Join(suite.backupDir, "config-backup-v1.2.3.tar.gz"), []byte(configBackup), 0644)
	require.NoError(t, err)
}

// createTestApplicationBinaries creates test application binaries
func (suite *E2ETestSuite) createTestApplicationBinaries(t *testing.T) {
	// Create test binary directory
	binaryDir := filepath.Join(suite.backupDir, "binaries")
	err := os.MkdirAll(binaryDir, 0755)
	require.NoError(t, err)

	// Create test binary files
	binaryContent := `#!/bin/bash
echo "Test KYB Platform Binary v1.2.3"
echo "Starting application..."
echo "Application started successfully"
`
	err = os.WriteFile(filepath.Join(binaryDir, "kyb-platform-v1.2.3"), []byte(binaryContent), 0755)
	require.NoError(t, err)
}

// TestE2ERollbackWorkflow tests the complete end-to-end rollback workflow
func TestE2ERollbackWorkflow(t *testing.T) {
	suite := SetupE2ETestSuite(t)
	defer suite.TeardownE2ETestSuite(t)

	t.Run("complete_e2e_rollback_workflow", func(t *testing.T) {
		// Phase 1: Pre-rollback verification
		t.Run("pre_rollback_verification", func(t *testing.T) {
			// Verify current system state
			var merchantCount int
			err := suite.db.QueryRow("SELECT COUNT(*) FROM merchants").Scan(&merchantCount)
			require.NoError(t, err)
			assert.Greater(t, merchantCount, 0, "Should have merchants in database")

			// Verify configuration files exist
			configFiles := []string{
				"database.yaml",
				"api.yaml",
				"security.yaml",
				"features.json",
			}

			for _, configFile := range configFiles {
				configPath := filepath.Join(suite.configDir, configFile)
				_, err := os.Stat(configPath)
				assert.NoError(t, err, "Configuration file should exist: %s", configFile)
			}
		})

		// Phase 2: Database rollback
		t.Run("database_rollback_e2e", func(t *testing.T) {
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
			assert.Contains(t, outputStr, "Starting database rollback process", "Should show start message")
			assert.Contains(t, outputStr, "DRY RUN", "Should be in dry run mode")
			assert.Contains(t, outputStr, "completed successfully", "Should show completion message")
		})

		// Phase 3: Application rollback
		t.Run("application_rollback_e2e", func(t *testing.T) {
			scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "application-rollback.sh")

			cmd := exec.Command("bash", scriptPath, "--dry-run", "--backup", "--target", "v1.2.3", "--environment", "production", "full")

			output, err := cmd.CombinedOutput()
			assert.NoError(t, err, "Application rollback should succeed")

			outputStr := string(output)
			assert.Contains(t, outputStr, "Starting application rollback process", "Should show start message")
			assert.Contains(t, outputStr, "DRY RUN", "Should be in dry run mode")
			assert.Contains(t, outputStr, "completed successfully", "Should show completion message")
		})

		// Phase 4: Configuration rollback
		t.Run("configuration_rollback_e2e", func(t *testing.T) {
			scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "configuration-rollback.sh")

			cmd := exec.Command("bash", scriptPath, "--dry-run", "--backup", "--target", "v1.2.3", "--environment", "production", "full")

			output, err := cmd.CombinedOutput()
			assert.NoError(t, err, "Configuration rollback should succeed")

			outputStr := string(output)
			assert.Contains(t, outputStr, "Starting configuration rollback process", "Should show start message")
			assert.Contains(t, outputStr, "DRY RUN", "Should be in dry run mode")
			assert.Contains(t, outputStr, "completed successfully", "Should show completion message")
		})

		// Phase 5: Post-rollback verification
		t.Run("post_rollback_verification", func(t *testing.T) {
			// Verify system state after rollback
			var merchantCount int
			err := suite.db.QueryRow("SELECT COUNT(*) FROM merchants").Scan(&merchantCount)
			require.NoError(t, err)
			assert.Greater(t, merchantCount, 0, "Should still have merchants in database")

			// Verify log files were created
			logFiles, err := filepath.Glob(filepath.Join(suite.logDir, "*-rollback-*.log"))
			require.NoError(t, err)
			assert.Greater(t, len(logFiles), 0, "Should create rollback log files")
		})
	})
}

// TestE2ERollbackWithRealApplication tests rollback with a real application scenario
func TestE2ERollbackWithRealApplication(t *testing.T) {
	suite := SetupE2ETestSuite(t)
	defer suite.TeardownE2ETestSuite(t)

	t.Run("rollback_with_real_application", func(t *testing.T) {
		// Simulate application running
		t.Run("simulate_application_running", func(t *testing.T) {
			// Create a mock application process
			appPath := filepath.Join(suite.backupDir, "binaries", "kyb-platform-v1.2.3")

			// Test that the application binary exists and is executable
			_, err := os.Stat(appPath)
			assert.NoError(t, err, "Application binary should exist")

			// Test execution
			cmd := exec.Command("bash", appPath)
			output, err := cmd.CombinedOutput()
			assert.NoError(t, err, "Application binary should be executable")

			outputStr := string(output)
			assert.Contains(t, outputStr, "Test KYB Platform Binary v1.2.3", "Should show version")
			assert.Contains(t, outputStr, "Application started successfully", "Should show success message")
		})

		// Test rollback with running application
		t.Run("rollback_with_running_app", func(t *testing.T) {
			scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "application-rollback.sh")

			cmd := exec.Command("bash", scriptPath, "--dry-run", "--target", "v1.2.3", "--environment", "production", "binary")

			output, err := cmd.CombinedOutput()
			assert.NoError(t, err, "Application rollback should succeed")

			outputStr := string(output)
			assert.Contains(t, outputStr, "Starting application rollback process", "Should show start message")
			assert.Contains(t, outputStr, "DRY RUN", "Should be in dry run mode")
		})
	})
}

// TestE2ERollbackErrorScenarios tests error scenarios in end-to-end rollback
func TestE2ERollbackErrorScenarios(t *testing.T) {
	suite := SetupE2ETestSuite(t)
	defer suite.TeardownE2ETestSuite(t)

	t.Run("rollback_error_scenarios", func(t *testing.T) {
		// Test with missing backup files
		t.Run("missing_backup_files", func(t *testing.T) {
			scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "database-rollback.sh")

			cmd := exec.Command("bash", scriptPath, "--dry-run", "--target", "nonexistent", "data")
			cmd.Env = append(os.Environ(),
				"DB_HOST=localhost",
				"DB_PORT=5432",
				"DB_NAME=kyb_platform_test",
				"DB_USER=postgres",
				"DB_PASSWORD=password",
			)

			output, err := cmd.CombinedOutput()
			// Should still succeed in dry run mode
			assert.NoError(t, err, "Dry run should succeed even with missing backup")

			outputStr := string(output)
			assert.Contains(t, outputStr, "DRY RUN", "Should be in dry run mode")
		})

		// Test with invalid configuration
		t.Run("invalid_configuration", func(t *testing.T) {
			// Create invalid configuration file
			invalidConfig := `invalid: yaml: content: [`
			err := os.WriteFile(filepath.Join(suite.configDir, "invalid.yaml"), []byte(invalidConfig), 0644)
			require.NoError(t, err)

			scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", "configuration-rollback.sh")

			cmd := exec.Command("bash", scriptPath, "--dry-run", "--target", "v1.2.3", "--environment", "production", "full")

			output, err := cmd.CombinedOutput()
			assert.NoError(t, err, "Configuration rollback should handle invalid config gracefully")

			outputStr := string(output)
			assert.Contains(t, outputStr, "DRY RUN", "Should be in dry run mode")
		})
	})
}

// TestE2ERollbackPerformance tests performance of end-to-end rollback procedures
func TestE2ERollbackPerformance(t *testing.T) {
	suite := SetupE2ETestSuite(t)
	defer suite.TeardownE2ETestSuite(t)

	t.Run("e2e_rollback_performance", func(t *testing.T) {
		scripts := []struct {
			name   string
			script string
			args   []string
		}{
			{
				name:   "database_rollback",
				script: "database-rollback.sh",
				args:   []string{"--dry-run", "--backup", "--target", "v1.2.3", "full"},
			},
			{
				name:   "application_rollback",
				script: "application-rollback.sh",
				args:   []string{"--dry-run", "--backup", "--target", "v1.2.3", "--environment", "production", "full"},
			},
			{
				name:   "configuration_rollback",
				script: "configuration-rollback.sh",
				args:   []string{"--dry-run", "--backup", "--target", "v1.2.3", "--environment", "production", "full"},
			},
		}

		for _, test := range scripts {
			t.Run(test.name, func(t *testing.T) {
				scriptPath := filepath.Join(suite.projectRoot, "scripts", "rollback", test.script)

				// Measure execution time
				start := time.Now()
				cmd := exec.Command("bash", append([]string{scriptPath}, test.args...)...)
				output, err := cmd.CombinedOutput()
				duration := time.Since(start)

				assert.NoError(t, err, "Rollback script should execute successfully")
				assert.Less(t, duration, 15*time.Second, "Rollback script should complete within 15 seconds")

				// Verify output is reasonable
				outputStr := string(output)
				assert.Contains(t, outputStr, "Starting", "Should show start message")
				assert.Contains(t, outputStr, "completed successfully", "Should show completion message")
			})
		}
	})
}
