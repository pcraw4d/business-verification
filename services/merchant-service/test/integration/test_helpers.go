package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"kyb-platform/services/merchant-service/internal/config"
	"kyb-platform/services/merchant-service/internal/supabase"
)

// TestDatabase wraps a Supabase client for testing
type TestDatabase struct {
	client *supabase.Client
	logger *zap.Logger
}

// SetupTestDatabase creates a test database connection
func SetupTestDatabase(t *testing.T) *TestDatabase {
	logger := zaptest.NewLogger(t)

	// Load test configuration from environment
	cfg := &config.Config{
		Environment: "test",
		Supabase: config.SupabaseConfig{
			URL:    getEnvOrDefault("SUPABASE_URL", ""),
			APIKey: getEnvOrDefault("SUPABASE_SERVICE_ROLE_KEY", ""),
		},
	}

	// If no Supabase config, skip test
	if cfg.Supabase.URL == "" || cfg.Supabase.APIKey == "" {
		t.Skip("Skipping integration test: SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY must be set")
	}

	// Create Supabase client
	client, err := supabase.NewClient(&cfg.Supabase, logger)
	if err != nil {
		t.Skipf("Skipping integration test: Failed to create Supabase client: %v", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.HealthCheck(ctx); err != nil {
		t.Skipf("Skipping integration test: Database health check failed: %v", err)
	}

	return &TestDatabase{
		client: client,
		logger: logger,
	}
}

// TeardownTestDatabase cleans up test database connection
func (td *TestDatabase) TeardownTestDatabase() {
	if td.client != nil {
		// Close any connections if needed
		td.logger.Info("Test database connection closed")
	}
}

// CleanupTestData removes test data from database
func (td *TestDatabase) CleanupTestData(t *testing.T, merchantIDs []string) {
	if len(merchantIDs) == 0 {
		return
	}

	// Delete merchant analytics
	for _, merchantID := range merchantIDs {
		_, _, err := td.client.GetClient().From("merchant_analytics").
			Delete("", "").
			Eq("merchant_id", merchantID).
			Execute()
		if err != nil {
			td.logger.Warn("Failed to cleanup merchant_analytics",
				zap.String("merchant_id", merchantID),
				zap.Error(err))
		}
	}

	// Delete merchants
	for _, merchantID := range merchantIDs {
		_, _, err := td.client.GetClient().From("merchants").
			Delete("", "").
			Eq("id", merchantID).
			Execute()
		if err != nil {
			td.logger.Warn("Failed to cleanup merchant",
				zap.String("merchant_id", merchantID),
				zap.Error(err))
		}
	}

	td.logger.Info("Test data cleaned up", zap.Int("merchants_cleaned", len(merchantIDs)))
}

// GetClient returns the Supabase client
func (td *TestDatabase) GetClient() *supabase.Client {
	return td.client
}

// getEnvOrDefault returns environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CreateTestMerchant creates a test merchant in the database
func (td *TestDatabase) CreateTestMerchant(t *testing.T, merchantData map[string]interface{}) (string, error) {
	// Generate merchant ID if not provided
	if _, exists := merchantData["id"]; !exists {
		// Use timestamp-based ID similar to handler
		merchantData["id"] = fmt.Sprintf("merchant_%d", time.Now().UnixNano())
	}
	
	// Ensure required fields have defaults
	if _, exists := merchantData["created_at"]; !exists {
		merchantData["created_at"] = time.Now().Format(time.RFC3339)
	}
	if _, exists := merchantData["updated_at"]; !exists {
		merchantData["updated_at"] = time.Now().Format(time.RFC3339)
	}

	var result []map[string]interface{}
	_, err := td.client.GetClient().From("merchants").
		Insert(merchantData, false, "", "", "").
		ExecuteTo(&result)

	if err != nil {
		return "", fmt.Errorf("failed to create test merchant: %w", err)
	}

	if len(result) == 0 {
		return "", fmt.Errorf("no merchant ID returned from insert")
	}

	merchantID, ok := result[0]["id"].(string)
	if !ok {
		return "", fmt.Errorf("merchant ID is not a string")
	}

	return merchantID, nil
}

// GetTestMerchant retrieves a test merchant from the database
func (td *TestDatabase) GetTestMerchant(t *testing.T, merchantID string) (map[string]interface{}, error) {
	var result []map[string]interface{}
	_, err := td.client.GetClient().From("merchants").
		Select("*", "", false).
		Eq("id", merchantID).
		Limit(1, "").
		ExecuteTo(&result)

	if err != nil {
		return nil, fmt.Errorf("failed to get test merchant: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("merchant not found")
	}

	return result[0], nil
}

