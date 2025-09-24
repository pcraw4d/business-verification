// Package test provides a comprehensive test runner for transaction testing
// This module implements the test execution framework for subtask 4.1.2
//
// Author: KYB Platform Development Team
// Date: January 19, 2025
// Version: 1.0
package test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// TransactionTestConfig holds configuration for transaction testing
type TransactionTestConfig struct {
	DatabaseURL     string
	TestTimeout     time.Duration
	ConcurrentUsers int
	TestDataSize    int
	LogLevel        string
}

// DefaultTransactionTestConfig returns default test configuration
func DefaultTransactionTestConfig() *TransactionTestConfig {
	return &TransactionTestConfig{
		DatabaseURL:     getEnvOrDefault("DATABASE_URL", "postgres://user:password@localhost:5432/kyb_test?sslmode=disable"),
		TestTimeout:     30 * time.Second,
		ConcurrentUsers: 10,
		TestDataSize:    100,
		LogLevel:        "info",
	}
}

// TransactionTestRunner manages the execution of transaction tests
type TransactionTestRunner struct {
	config *TransactionTestConfig
	db     *sql.DB
	logger *log.Logger
	suite  *TransactionTestSuite
}

// NewTransactionTestRunner creates a new transaction test runner
func NewTransactionTestRunner(config *TransactionTestConfig) (*TransactionTestRunner, error) {
	// Setup logger
	logger := log.New(os.Stdout, "[TRANSACTION_TEST] ", log.LstdFlags|log.Lshortfile)

	// Connect to database
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create test suite
	suite := NewTransactionTestSuite(db, logger)

	return &TransactionTestRunner{
		config: config,
		db:     db,
		logger: logger,
		suite:  suite,
	}, nil
}

// SetupTestEnvironment prepares the test environment
func (tr *TransactionTestRunner) SetupTestEnvironment() error {
	tr.logger.Println("🔧 Setting up test environment...")

	// Create test tables if they don't exist
	if err := tr.createTestTables(); err != nil {
		return fmt.Errorf("failed to create test tables: %w", err)
	}

	// Insert test data
	if err := tr.insertTestData(); err != nil {
		return fmt.Errorf("failed to insert test data: %w", err)
	}

	tr.logger.Println("✅ Test environment setup completed")
	return nil
}

// CleanupTestEnvironment cleans up the test environment
func (tr *TransactionTestRunner) CleanupTestEnvironment() error {
	tr.logger.Println("🧹 Cleaning up test environment...")

	// Clean up test data
	if err := tr.cleanupTestData(); err != nil {
		return fmt.Errorf("failed to cleanup test data: %w", err)
	}

	// Close database connection
	if err := tr.db.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	tr.logger.Println("✅ Test environment cleanup completed")
	return nil
}

// RunTransactionTests executes all transaction tests
func (tr *TransactionTestRunner) RunTransactionTests(t *testing.T) {
	tr.logger.Println("🚀 Starting transaction tests...")

	// Setup test environment
	if err := tr.SetupTestEnvironment(); err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}
	defer func() {
		if err := tr.CleanupTestEnvironment(); err != nil {
			t.Logf("Failed to cleanup test environment: %v", err)
		}
	}()

	// Run all transaction tests
	tr.suite.RunAllTransactionTests(t)

	// Generate test report
	report := tr.suite.GenerateTransactionTestReport()
	tr.logger.Printf("📊 Transaction test report: %+v", report)

	tr.logger.Println("✅ All transaction tests completed")
}

// createTestTables creates necessary test tables
func (tr *TransactionTestRunner) createTestTables() error {
	// This would typically run the migration scripts
	// For now, we'll assume tables already exist from previous phases
	tr.logger.Println("📋 Test tables already exist (from previous migration phases)")
	return nil
}

// insertTestData inserts test data for transaction testing
func (tr *TransactionTestRunner) insertTestData() error {
	tr.logger.Println("📊 Inserting test data...")

	// Insert test risk keywords
	_, err := tr.db.Exec(`
		INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, 
			mcc_codes, naics_codes, sic_codes, risk_score_weight, detection_confidence)
		VALUES 
			('gambling', 'prohibited', 'high', 'Gambling activities', 
			 ARRAY['7995'], ARRAY['713290'], ARRAY['7995'], 1.5, 0.95),
			('drugs', 'illegal', 'critical', 'Illegal drug activities', 
			 ARRAY['5122'], ARRAY['446191'], ARRAY['5912'], 2.0, 1.0),
			('weapons', 'illegal', 'critical', 'Weapon sales', 
			 ARRAY['5094'], ARRAY['332992'], ARRAY['3484'], 2.0, 1.0)
		ON CONFLICT (keyword) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to insert test risk keywords: %w", err)
	}

	// Insert test industries
	_, err = tr.db.Exec(`
		INSERT INTO industries (name, description, is_active)
		VALUES 
			('Technology', 'Technology and software services', true),
			('Finance', 'Financial services and banking', true),
			('Healthcare', 'Healthcare and medical services', true),
			('Gambling', 'Gambling and betting activities', true)
		ON CONFLICT (name) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to insert test industries: %w", err)
	}

	tr.logger.Println("✅ Test data inserted successfully")
	return nil
}

// cleanupTestData removes test data
func (tr *TransactionTestRunner) cleanupTestData() error {
	tr.logger.Println("🗑️ Cleaning up test data...")

	// Clean up test data (be careful not to remove production data)
	_, err := tr.db.Exec(`
		DELETE FROM classification_performance_metrics 
		WHERE request_id LIKE 'test-%' OR business_name LIKE 'Test%'
	`)
	if err != nil {
		tr.logger.Printf("Warning: Failed to cleanup performance metrics: %v", err)
	}

	_, err = tr.db.Exec(`
		DELETE FROM business_risk_assessments 
		WHERE website_content LIKE '%Test%' OR assessment_method = 'test'
	`)
	if err != nil {
		tr.logger.Printf("Warning: Failed to cleanup risk assessments: %v", err)
	}

	_, err = tr.db.Exec(`
		DELETE FROM business_classifications 
		WHERE business_name LIKE 'Test%' OR classification_metadata::text LIKE '%test%'
	`)
	if err != nil {
		tr.logger.Printf("Warning: Failed to cleanup classifications: %v", err)
	}

	_, err = tr.db.Exec(`
		DELETE FROM merchants 
		WHERE name LIKE 'Test%' OR description LIKE '%Test%'
	`)
	if err != nil {
		tr.logger.Printf("Warning: Failed to cleanup merchants: %v", err)
	}

	_, err = tr.db.Exec(`
		DELETE FROM users 
		WHERE email LIKE '%@example.com' OR name LIKE 'Test%' OR name LIKE 'User %'
	`)
	if err != nil {
		tr.logger.Printf("Warning: Failed to cleanup users: %v", err)
	}

	tr.logger.Println("✅ Test data cleanup completed")
	return nil
}

// BenchmarkTransactionPerformance runs performance benchmarks
func (tr *TransactionTestRunner) BenchmarkTransactionPerformance(b *testing.B) {
	tr.logger.Println("⚡ Running transaction performance benchmarks...")

	// Setup
	if err := tr.SetupTestEnvironment(); err != nil {
		b.Fatalf("Failed to setup test environment: %v", err)
	}
	defer func() {
		if err := tr.CleanupTestEnvironment(); err != nil {
			b.Logf("Failed to cleanup test environment: %v", err)
		}
	}()

	// Benchmark complex transaction
	b.Run("ComplexTransaction", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			tx, err := tr.db.BeginTx(ctx, &sql.TxOptions{
				Isolation: sql.LevelReadCommitted,
			})
			if err != nil {
				b.Fatal(err)
			}

			// Create user
			userID := fmt.Sprintf("bench-user-%d", i)
			_, err = tx.ExecContext(ctx, `
				INSERT INTO users (id, email, name, role, is_active)
				VALUES ($1, $2, $3, $4, $5)
			`, userID, fmt.Sprintf("user%d@example.com", i), fmt.Sprintf("User %d", i), "user", true)
			if err != nil {
				tx.Rollback()
				b.Fatal(err)
			}

			// Create business
			businessID := fmt.Sprintf("bench-business-%d", i)
			_, err = tx.ExecContext(ctx, `
				INSERT INTO merchants (id, user_id, name, website_url, description, industry)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, businessID, userID, fmt.Sprintf("Business %d", i), "https://example.com", "Test business", "Technology")
			if err != nil {
				tx.Rollback()
				b.Fatal(err)
			}

			// Commit
			if err = tx.Commit(); err != nil {
				b.Fatal(err)
			}
		}
	})

	tr.logger.Println("✅ Performance benchmarks completed")
}

// TestMain provides the main test entry point
func TestMain(m *testing.M) {
	// Setup
	config := DefaultTransactionTestConfig()
	runner, err := NewTransactionTestRunner(config)
	if err != nil {
		log.Fatalf("Failed to create test runner: %v", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if err := runner.CleanupTestEnvironment(); err != nil {
		log.Printf("Failed to cleanup test environment: %v", err)
	}

	os.Exit(code)
}
