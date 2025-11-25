package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// TestDBConfig holds test database configuration
type TestDBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// GetTestDBConfig returns test database configuration from environment or defaults
func GetTestDBConfig() *TestDBConfig {
	config := &TestDBConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     5432,
		User:     getEnv("TEST_DB_USER", "postgres"),
		Password: getEnv("TEST_DB_PASSWORD", "postgres"),
		DBName:   getEnv("TEST_DB_NAME", "kyb_platform_test"),
		SSLMode:  getEnv("TEST_DB_SSLMODE", "disable"),
	}

	if port := os.Getenv("TEST_DB_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &config.Port)
	}

	return config
}

// ConnectTestDB connects to the test database
func ConnectTestDB(t *testing.T) (*sql.DB, error) {
	config := GetTestDBConfig()
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// SetupTestDB sets up a test database with required tables
func SetupTestDB(t *testing.T, db *sql.DB) error {
	ctx := context.Background()

	// Create tables if they don't exist
	schema := `
	CREATE TABLE IF NOT EXISTS industries (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		category VARCHAR(100),
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS classification_codes (
		id SERIAL PRIMARY KEY,
		industry_id INTEGER REFERENCES industries(id),
		code VARCHAR(50) NOT NULL,
		code_type VARCHAR(10) NOT NULL,
		description TEXT,
		is_active BOOLEAN DEFAULT true,
		is_primary BOOLEAN DEFAULT false,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS code_keywords (
		id SERIAL PRIMARY KEY,
		code_id INTEGER REFERENCES classification_codes(id),
		keyword VARCHAR(255) NOT NULL,
		relevance_score FLOAT DEFAULT 0.5,
		match_type VARCHAR(20) DEFAULT 'exact',
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_code_keywords_keyword ON code_keywords(keyword);
	CREATE INDEX IF NOT EXISTS idx_code_keywords_code_id ON code_keywords(code_id);
	CREATE INDEX IF NOT EXISTS idx_classification_codes_code_type ON classification_codes(code_type);
	CREATE INDEX IF NOT EXISTS idx_classification_codes_industry_id ON classification_codes(industry_id);
	`

	_, err := db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// CleanupTestDB cleans up test data from the database
func CleanupTestDB(t *testing.T, db *sql.DB) error {
	ctx := context.Background()

	// Delete test data in reverse order of dependencies
	queries := []string{
		"DELETE FROM code_keywords",
		"DELETE FROM classification_codes",
		"DELETE FROM industries",
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to cleanup: %w", err)
		}
	}

	return nil
}

// SeedTestData seeds the test database with sample data
func SeedTestData(t *testing.T, db *sql.DB) error {
	ctx := context.Background()

	// Insert industries
	industries := []struct {
		id   int
		name string
	}{
		{1, "Technology"},
		{2, "Financial Services"},
		{3, "Healthcare"},
	}

	for _, industry := range industries {
		_, err := db.ExecContext(ctx,
			"INSERT INTO industries (id, name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING",
			industry.id, industry.name)
		if err != nil {
			return fmt.Errorf("failed to insert industry: %w", err)
		}
	}

	// Insert classification codes
	codes := []struct {
		id          int
		industryID  int
		code        string
		codeType    string
		description string
	}{
		{1, 1, "5734", "MCC", "Computer Software Stores"},
		{2, 1, "7372", "SIC", "Prepackaged Software"},
		{3, 1, "541511", "NAICS", "Custom Computer Programming Services"},
		{4, 2, "6011", "MCC", "Automated Teller Machine Services"},
		{5, 2, "6021", "SIC", "National Commercial Banks"},
		{6, 2, "522110", "NAICS", "Commercial Banking"},
	}

	for _, code := range codes {
		_, err := db.ExecContext(ctx,
			`INSERT INTO classification_codes (id, industry_id, code, code_type, description)
			 VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING`,
			code.id, code.industryID, code.code, code.codeType, code.description)
		if err != nil {
			return fmt.Errorf("failed to insert code: %w", err)
		}
	}

	// Insert keyword mappings
	keywordMappings := []struct {
		codeID         int
		keyword        string
		relevanceScore float64
		matchType      string
	}{
		{1, "software", 0.9, "exact"},
		{1, "technology", 0.8, "partial"},
		{3, "software", 0.85, "exact"},
		{3, "development", 0.75, "partial"},
		{4, "bank", 0.9, "exact"},
		{4, "finance", 0.8, "partial"},
		{6, "bank", 0.85, "exact"},
		{6, "banking", 0.8, "exact"},
	}

	for _, mapping := range keywordMappings {
		_, err := db.ExecContext(ctx,
			`INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
			 VALUES ($1, $2, $3, $4)`,
			mapping.codeID, mapping.keyword, mapping.relevanceScore, mapping.matchType)
		if err != nil {
			return fmt.Errorf("failed to insert keyword mapping: %w", err)
		}
	}

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

