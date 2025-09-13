package integration

import (
	"os"
	"testing"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	DatabaseURL string
	TestUserID  string
	TestTimeout int
}

// GetTestConfig returns the test configuration
func GetTestConfig() *TestConfig {
	return &TestConfig{
		DatabaseURL: getEnvOrDefault("TEST_DATABASE_URL", "postgres://postgres:password@localhost:5432/kyb_test?sslmode=disable"),
		TestUserID:  getEnvOrDefault("TEST_USER_ID", "test-user-123"),
		TestTimeout: 30, // seconds
	}
}

// getEnvOrDefault returns the environment variable value or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SkipIfShort skips the test if the short flag is set
func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
}

// SkipIfNoDatabase skips the test if no database is available
func SkipIfNoDatabase(t *testing.T) {
	if os.Getenv("SKIP_DATABASE_TESTS") == "true" {
		t.Skip("skipping database tests")
	}
}

