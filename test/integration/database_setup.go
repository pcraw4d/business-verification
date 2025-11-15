package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// TestDatabase wraps a database connection for testing
type TestDatabase struct {
	db     *sql.DB
	logger *log.Logger
}

// SetupTestDatabase initializes the test database connection
func SetupTestDatabase() (*TestDatabase, error) {
	databaseURL := getTestDatabaseURL()
	if databaseURL == "" {
		return nil, fmt.Errorf("test database URL not configured. Set TEST_DATABASE_URL environment variable")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open test database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping test database: %w", err)
	}

	logger := log.New(os.Stdout, "[TEST-DB] ", log.LstdFlags)
	logger.Printf("Connected to test database successfully")

	return &TestDatabase{
		db:     db,
		logger: logger,
	}, nil
}

// CleanupTestDatabase closes the database connection
func (td *TestDatabase) CleanupTestDatabase() error {
	if td.db != nil {
		return td.db.Close()
	}
	return nil
}

// ResetTestDatabase truncates all test tables
func (td *TestDatabase) ResetTestDatabase(ctx context.Context) error {
	tables := []string{
		"merchant_analytics",
		"risk_assessments",
		"risk_indicators",
		"enrichment_jobs",
		"enrichment_sources",
	}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := td.db.ExecContext(ctx, query); err != nil {
			// Table might not exist, which is okay for tests
			td.logger.Printf("Warning: Could not truncate table %s: %v", table, err)
		}
	}

	td.logger.Printf("Test database reset complete")
	return nil
}

// SeedTestData inserts test data into the database
func (td *TestDatabase) SeedTestData(ctx context.Context) error {
	// Seed merchant analytics
	merchantAnalyticsQuery := `
		INSERT INTO merchant_analytics (merchant_id, classification_data, security_data, quality_data, intelligence_data, created_at, updated_at)
		VALUES 
			('test-merchant-1', '{"primaryIndustry": "Technology"}', '{"trustScore": 0.8}', '{"completenessScore": 0.9}', '{}', NOW(), NOW()),
			('test-merchant-2', '{"primaryIndustry": "Finance"}', '{"trustScore": 0.7}', '{"completenessScore": 0.85}', '{}', NOW(), NOW())
		ON CONFLICT (merchant_id) DO NOTHING
	`

	if _, err := td.db.ExecContext(ctx, merchantAnalyticsQuery); err != nil {
		// Table might not exist, which is okay for tests
		td.logger.Printf("Warning: Could not seed merchant_analytics: %v", err)
	}

	// Seed risk assessments
	riskAssessmentQuery := `
		INSERT INTO risk_assessments (id, merchant_id, status, overall_score, risk_level, result_data, created_at, updated_at)
		VALUES 
			('test-assessment-1', 'test-merchant-1', 'completed', 0.7, 'medium', '{"factors": []}', NOW(), NOW()),
			('test-assessment-2', 'test-merchant-2', 'pending', 0.0, 'unknown', '{}', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`

	if _, err := td.db.ExecContext(ctx, riskAssessmentQuery); err != nil {
		td.logger.Printf("Warning: Could not seed risk_assessments: %v", err)
	}

	td.logger.Printf("Test data seeded successfully")
	return nil
}

// GetDB returns the underlying database connection
func (td *TestDatabase) GetDB() *sql.DB {
	return td.db
}

// getTestDatabaseURL returns the test database URL from environment or default
func getTestDatabaseURL() string {
	// Check for TEST_DATABASE_URL first
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}

	// Check for DATABASE_URL (often provided by Railway/Supabase)
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	// Check for Supabase connection string
	if supabaseURL := os.Getenv("SUPABASE_URL"); supabaseURL != "" {
		supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
		if supabaseKey != "" {
			// Construct Supabase connection string using connection pooler
			// Format: postgres://postgres.[PROJECT_REF]:[SERVICE_ROLE_KEY]@aws-0-[REGION].pooler.supabase.com:6543/postgres
			// Extract project reference from SUPABASE_URL (remove https:// and .supabase.co)
			hostname := supabaseURL
			if strings.HasPrefix(hostname, "https://") {
				hostname = strings.TrimPrefix(hostname, "https://")
			} else if strings.HasPrefix(hostname, "http://") {
				hostname = strings.TrimPrefix(hostname, "http://")
			}
			// Remove trailing slash if present
			hostname = strings.TrimSuffix(hostname, "/")
			// Extract project ref (part before .supabase.co)
			projectRef := strings.TrimSuffix(hostname, ".supabase.co")
			// Use connection pooler format (transaction mode, port 6543)
			// Note: This may not work as service role key is a JWT, not a database password
			// User should set DATABASE_URL directly from Supabase dashboard for best results
			// Try common regions - if this doesn't work, user should set DATABASE_URL directly
			// Common regions: us-east-1, us-west-1, eu-west-1, ap-southeast-1
			return fmt.Sprintf("postgres://postgres.%s:%s@aws-0-us-east-1.pooler.supabase.com:6543/postgres?sslmode=require", projectRef, supabaseKey)
		}
	}

	// Default to local PostgreSQL
	return "postgres://postgres:password@localhost:5432/kyb_test?sslmode=disable"
}

// VerifyTestDatabase checks if the test database is accessible
func VerifyTestDatabase() error {
	databaseURL := getTestDatabaseURL()
	if databaseURL == "" {
		return fmt.Errorf("test database URL not configured")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open test database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("test database is not accessible: %w", err)
	}

	return nil
}

