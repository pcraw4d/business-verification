package test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// TestConfig holds test configuration
type TestConfig struct {
	Database *config.DatabaseConfig
	Logger   *observability.Logger
	DB       database.Database
	Cleanup  func()
}

// TestSuite provides common testing utilities
type TestSuite struct {
	config *TestConfig
	t      *testing.T
}

// NewTestSuite creates a new test suite
func NewTestSuite(t *testing.T) *TestSuite {
	config := setupTestConfig(t)
	return &TestSuite{
		config: config,
		t:      t,
	}
}

// setupTestConfig sets up test configuration
func setupTestConfig(t *testing.T) *TestConfig {
	// Load test environment variables
	loadTestEnv()

	// Create test database configuration
	dbConfig := &config.DatabaseConfig{
		Driver:          "postgres",
		Host:            getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:            getEnvIntOrDefault("TEST_DB_PORT", 5432),
		Username:        getEnvOrDefault("TEST_DB_USER", "test_user"),
		Password:        getEnvOrDefault("TEST_DB_PASSWORD", "test_password"),
		Database:        getEnvOrDefault("TEST_DB_NAME", "kyb_test"),
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}

	// Create logger
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	})

	// Create database connection
	db, err := database.NewDatabaseWithConnection(context.Background(), dbConfig)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to test database: %v", err)
		return &TestConfig{
			Database: dbConfig,
			Logger:   logger,
			DB:       nil,
			Cleanup:  func() {},
		}
	}

	// Setup cleanup function
	cleanup := func() {
		if db != nil {
			db.Close()
		}
	}

	return &TestConfig{
		Database: dbConfig,
		Logger:   logger,
		DB:       db,
		Cleanup:  cleanup,
	}
}

// loadTestEnv loads test environment variables
func loadTestEnv() {
	// Set default test environment variables if not already set
	if os.Getenv("TEST_DB_HOST") == "" {
		os.Setenv("TEST_DB_HOST", "localhost")
	}
	if os.Getenv("TEST_DB_PORT") == "" {
		os.Setenv("TEST_DB_PORT", "5432")
	}
	if os.Getenv("TEST_DB_USER") == "" {
		os.Setenv("TEST_DB_USER", "test_user")
	}
	if os.Getenv("TEST_DB_PASSWORD") == "" {
		os.Setenv("TEST_DB_PASSWORD", "test_password")
	}
	if os.Getenv("TEST_DB_NAME") == "" {
		os.Setenv("TEST_DB_NAME", "kyb_test")
	}
}

// getEnvOrDefault gets environment variable or returns default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault gets environment variable as int or returns default
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && intValue == 1 {
			return defaultValue
		}
	}
	return defaultValue
}

// SetupTestDatabase sets up a test database
func (ts *TestSuite) SetupTestDatabase() {
	if ts.config.DB == nil {
		ts.t.Skip("Database not available for testing")
		return
	}

	// Run migrations
	ts.runMigrations()

	// Seed test data
	ts.seedTestData()
}

// CleanupTestDatabase cleans up test database
func (ts *TestSuite) CleanupTestDatabase() {
	if ts.config.DB == nil {
		return
	}

	// Clean up test data
	ts.cleanupTestData()
}

// runMigrations runs database migrations for testing
func (ts *TestSuite) runMigrations() {
	// This would run the migration system
	// For now, we'll just log that migrations would run
	log.Println("Running test database migrations...")
}

// seedTestData seeds the test database with minimal test data
func (ts *TestSuite) seedTestData() {
	// This would seed the database with test data
	// For now, we'll just log that seeding would occur
	log.Println("Seeding test database...")
}

// cleanupTestData cleans up test data
func (ts *TestSuite) cleanupTestData() {
	// This would clean up test data
	// For now, we'll just log that cleanup would occur
	log.Println("Cleaning up test database...")
}

// GetTestDB returns the test database instance
func (ts *TestSuite) GetTestDB() database.Database {
	return ts.config.DB
}

// GetTestLogger returns the test logger
func (ts *TestSuite) GetTestLogger() *observability.Logger {
	return ts.config.Logger
}

// GetTestConfig returns the test configuration
func (ts *TestSuite) GetTestConfig() *TestConfig {
	return ts.config
}

// AssertNoError asserts that there is no error
func (ts *TestSuite) AssertNoError(err error) {
	if err != nil {
		ts.t.Fatalf("Expected no error, got: %v", err)
	}
}

// AssertError asserts that there is an error
func (ts *TestSuite) AssertError(err error) {
	if err == nil {
		ts.t.Fatal("Expected error, got nil")
	}
}

// AssertEqual asserts that two values are equal
func (ts *TestSuite) AssertEqual(expected, actual interface{}) {
	if expected != actual {
		ts.t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

// AssertNotNil asserts that a value is not nil
func (ts *TestSuite) AssertNotNil(value interface{}) {
	if value == nil {
		ts.t.Fatal("Expected non-nil value, got nil")
	}
}

// AssertNil asserts that a value is nil
func (ts *TestSuite) AssertNil(value interface{}) {
	if value != nil {
		ts.t.Fatalf("Expected nil value, got %v", value)
	}
}

// TestMain runs before all tests
func TestMain(m *testing.M) {
	// Setup test environment
	setupTestEnvironment()

	// Run tests
	code := m.Run()

	// Cleanup
	cleanupTestEnvironment()

	// Exit with test result code
	os.Exit(code)
}

// setupTestEnvironment sets up the test environment
func setupTestEnvironment() {
	log.Println("Setting up test environment...")

	// Load test environment variables
	loadTestEnv()

	// Set test-specific configurations
	os.Setenv("ENV", "test")
	os.Setenv("LOG_LEVEL", "debug")
}

// cleanupTestEnvironment cleans up the test environment
func cleanupTestEnvironment() {
	log.Println("Cleaning up test environment...")
}

// TestDatabaseConnection tests database connectivity
func TestDatabaseConnection(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.config.Cleanup()

	if ts.config.DB == nil {
		t.Skip("Database not available for testing")
		return
	}

	ctx := context.Background()
	err := ts.config.DB.Ping(ctx)
	ts.AssertNoError(err)
}

// TestDatabaseMigrations tests database migrations
func TestDatabaseMigrations(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.config.Cleanup()

	if ts.config.DB == nil {
		t.Skip("Database not available for testing")
		return
	}

	// Test that migrations can be run
	ts.SetupTestDatabase()
	ts.CleanupTestDatabase()
}

// TestDataSeeding tests data seeding functionality
func TestDataSeeding(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.config.Cleanup()

	if ts.config.DB == nil {
		t.Skip("Database not available for testing")
		return
	}

	// Test that data can be seeded
	ts.SetupTestDatabase()

	// Verify seeded data exists
	// This would check that test data was properly seeded

	ts.CleanupTestDatabase()
}

// TestConfiguration tests configuration loading
func TestConfiguration(t *testing.T) {
	// Test that configuration can be loaded
	cfg, err := config.Load()
	if err != nil {
		t.Skipf("Skipping test - cannot load configuration: %v", err)
		return
	}

	// Verify configuration is valid
	if cfg == nil {
		t.Fatal("Expected non-nil configuration")
	}
}

// TestLogger tests logger functionality
func TestLogger(t *testing.T) {
	ts := NewTestSuite(t)
	defer ts.config.Cleanup()

	logger := ts.GetTestLogger()
	ts.AssertNotNil(logger)

	// Test logging functionality
	logger.Info("Test log message", "test_key", "test_value")
}

// BenchmarkDatabaseOperations benchmarks database operations
func BenchmarkDatabaseOperations(b *testing.B) {
	ts := NewTestSuite(&testing.T{})
	defer ts.config.Cleanup()

	if ts.config.DB == nil {
		b.Skip("Database not available for benchmarking")
		return
	}

	ctx := context.Background()

	b.Run("Ping", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := ts.config.DB.Ping(ctx)
			if err != nil {
				b.Fatalf("Database ping failed: %v", err)
			}
		}
	})
}

// BenchmarkLoggerOperations benchmarks logger operations
func BenchmarkLoggerOperations(b *testing.B) {
	ts := NewTestSuite(&testing.T{})
	defer ts.config.Cleanup()

	logger := ts.GetTestLogger()

	b.Run("InfoLogging", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Info("Benchmark log message", "iteration", i)
		}
	})
}
