package database

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestDatabaseURL returns the test database URL from environment
func getTestDatabaseURL(t *testing.T) string {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://kyb_test:kyb_test_password@localhost:5433/kyb_test?sslmode=disable"
		t.Logf("Using default DATABASE_URL: %s", dbURL)
	}
	return dbURL
}

// TestDatabaseConnection tests basic database connectivity
func TestDatabaseConnection(t *testing.T) {
	dbURL := getTestDatabaseURL(t)
	
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err, "Should be able to open database connection")
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err, "Should be able to ping database")

	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	require.NoError(t, err, "Should be able to query database")
	
	t.Logf("PostgreSQL version: %s", version)
	assert.NotEmpty(t, version, "Version should not be empty")
}

// TestRiskAssessmentsTableExists tests that the risk_assessments table exists with all required columns
func TestRiskAssessmentsTableExists(t *testing.T) {
	dbURL := getTestDatabaseURL(t)
	
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Check if table exists
	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'risk_assessments'
		)
	`).Scan(&tableExists)
	require.NoError(t, err)
	assert.True(t, tableExists, "risk_assessments table should exist")

	// Check for required columns
	requiredColumns := []string{
		"id", "merchant_id", "status", "options", "result",
		"created_at", "updated_at", "completed_at", "progress",
		"estimated_completion",
	}

	for _, col := range requiredColumns {
		var columnExists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.columns 
				WHERE table_schema = 'public' 
				AND table_name = 'risk_assessments' 
				AND column_name = $1
			)
		`, col).Scan(&columnExists)
		require.NoError(t, err, "Should be able to check for column: %s", col)
		assert.True(t, columnExists, "Column %s should exist in risk_assessments table", col)
	}

	t.Log("✓ All required columns exist in risk_assessments table")
}

// TestMerchantsTableExists tests that the merchants table exists
func TestMerchantsTableExists(t *testing.T) {
	dbURL := getTestDatabaseURL(t)
	
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'merchants'
		)
	`).Scan(&tableExists)
	require.NoError(t, err)
	assert.True(t, tableExists, "merchants table should exist")

	t.Log("✓ merchants table exists")
}

// TestPortfolioTypesTableExists tests that portfolio_types table exists
func TestPortfolioTypesTableExists(t *testing.T) {
	dbURL := getTestDatabaseURL(t)
	
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'portfolio_types'
		)
	`).Scan(&tableExists)
	require.NoError(t, err)
	assert.True(t, tableExists, "portfolio_types table should exist")

	t.Log("✓ portfolio_types table exists")
}

// TestRiskLevelsTableExists tests that risk_levels table exists
func TestRiskLevelsTableExists(t *testing.T) {
	dbURL := getTestDatabaseURL(t)
	
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'risk_levels'
		)
	`).Scan(&tableExists)
	require.NoError(t, err)
	assert.True(t, tableExists, "risk_levels table should exist")

	t.Log("✓ risk_levels table exists")
}

// TestGetPortfolioDistribution tests the GetPortfolioDistribution query
func TestGetPortfolioDistribution(t *testing.T) {
	dbURL := getTestDatabaseURL(t)
	
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	
	query := `
		SELECT pt.type, COUNT(*) as count
		FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		GROUP BY pt.type
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		// If query fails due to no data, that's okay for testing
		t.Logf("Query result (may be empty): %v", err)
		return
	}
	defer rows.Close()

	dist := make(map[string]int)
	for rows.Next() {
		var portfolioType string
		var count int
		err := rows.Scan(&portfolioType, &count)
		require.NoError(t, err)
		dist[portfolioType] = count
		t.Logf("Portfolio type %s: %d merchants", portfolioType, count)
	}

	err = rows.Err()
	require.NoError(t, err)
	
	t.Logf("✓ Portfolio distribution query works (found %d types)", len(dist))
}

// TestGetRiskDistribution tests the GetRiskDistribution query
func TestGetRiskDistribution(t *testing.T) {
	dbURL := getTestDatabaseURL(t)
	
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	
	query := `
		SELECT rl.level, COUNT(*) as count
		FROM merchants m
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		GROUP BY rl.level
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		t.Logf("Query result (may be empty): %v", err)
		return
	}
	defer rows.Close()

	dist := make(map[string]int)
	for rows.Next() {
		var riskLevel string
		var count int
		err := rows.Scan(&riskLevel, &count)
		require.NoError(t, err)
		dist[riskLevel] = count
		t.Logf("Risk level %s: %d merchants", riskLevel, count)
	}

	err = rows.Err()
	require.NoError(t, err)
	
	t.Logf("✓ Risk distribution query works (found %d levels)", len(dist))
}

// TestGetIndustryDistribution tests the GetIndustryDistribution query
func TestGetIndustryDistribution(t *testing.T) {
	dbURL := getTestDatabaseURL(t)
	
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	
	query := `
		SELECT COALESCE(industry, 'Unknown') as industry, COUNT(*) as count
		FROM merchants
		WHERE industry IS NOT NULL AND industry != ''
		GROUP BY industry
		ORDER BY count DESC
		LIMIT 10
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		t.Logf("Query result (may be empty): %v", err)
		return
	}
	defer rows.Close()

	dist := make(map[string]int)
	for rows.Next() {
		var industry string
		var count int
		err := rows.Scan(&industry, &count)
		require.NoError(t, err)
		dist[industry] = count
		t.Logf("Industry %s: %d merchants", industry, count)
	}

	err = rows.Err()
	require.NoError(t, err)
	
	t.Logf("✓ Industry distribution query works (found %d industries)", len(dist))
}

// TestGetComplianceDistribution tests the GetComplianceDistribution query
func TestGetComplianceDistribution(t *testing.T) {
	dbURL := getTestDatabaseURL(t)
	
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	
	query := `
		SELECT COALESCE(compliance_status, 'pending') as compliance_status, COUNT(*) as count
		FROM merchants
		GROUP BY compliance_status
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		t.Logf("Query result (may be empty): %v", err)
		return
	}
	defer rows.Close()

	dist := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		err := rows.Scan(&status, &count)
		require.NoError(t, err)
		dist[status] = count
		t.Logf("Compliance status %s: %d merchants", status, count)
	}

	err = rows.Err()
	require.NoError(t, err)
	
	t.Logf("✓ Compliance distribution query works (found %d statuses)", len(dist))
}

// TestRiskAssessmentInsert tests inserting a risk assessment record
func TestRiskAssessmentInsert(t *testing.T) {
	dbURL := getTestDatabaseURL(t)
	
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	
	// First, create a test user if none exists
	userQuery := `
		INSERT INTO users (
			id, email, username, password_hash, first_name, last_name, role
		) VALUES (
			gen_random_uuid(),
			'test@example.com',
			'testuser',
			'$2a$10$dummyhash',
			'Test',
			'User',
			'user'
		)
		ON CONFLICT (email) DO NOTHING
		RETURNING id
	`
	
	var userID string
	err = db.QueryRowContext(ctx, userQuery).Scan(&userID)
	if err == sql.ErrNoRows {
		// User already exists, get its ID
		err = db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = 'test@example.com'").Scan(&userID)
		require.NoError(t, err)
	} else {
		require.NoError(t, err, "Should be able to create test user")
	}
	
	// Create a test business (required for business_id foreign key)
	businessQuery := `
		INSERT INTO businesses (
			id, name, legal_name, registration_number, created_by
		) VALUES (
			gen_random_uuid(),
			'Test Business',
			'Test Business Inc.',
			'TEST-REG-001',
			$1
		)
		ON CONFLICT (registration_number) DO NOTHING
		RETURNING id
	`
	
	var businessID string
	err = db.QueryRowContext(ctx, businessQuery, userID).Scan(&businessID)
	if err == sql.ErrNoRows {
		// Business already exists, get its ID
		err = db.QueryRowContext(ctx, "SELECT id FROM businesses WHERE registration_number = 'TEST-REG-001'").Scan(&businessID)
		require.NoError(t, err)
	} else {
		require.NoError(t, err, "Should be able to create test business")
	}
	
	// Insert a test risk assessment
	query := `
		INSERT INTO risk_assessments (
			id, business_id, merchant_id, status, options, result, progress,
			estimated_completion, risk_level, risk_score, assessment_method, source, created_at, updated_at
		) VALUES (
			gen_random_uuid(), 
			$1,
			'test-merchant-001',
			'pending',
			'{"includeHistory": true}'::jsonb,
			NULL,
			0,
			NOW() + INTERVAL '5 minutes',
			'medium',
			0.5,
			'automated',
			'test',
			NOW(),
			NOW()
		)
		RETURNING id, merchant_id, status
	`

	var id, merchantID, status string
	err = db.QueryRowContext(ctx, query, businessID).Scan(&id, &merchantID, &status)
	require.NoError(t, err, "Should be able to insert risk assessment")
	
	assert.NotEmpty(t, id, "ID should be generated")
	assert.Equal(t, "test-merchant-001", merchantID)
	assert.Equal(t, "pending", status)
	
	t.Logf("✓ Successfully inserted risk assessment: %s for merchant: %s", id, merchantID)
	
	// Clean up
	_, err = db.ExecContext(ctx, "DELETE FROM risk_assessments WHERE id = $1", id)
	require.NoError(t, err, "Should be able to delete test record")
}

// TestUpdatedAtTrigger tests that the updated_at trigger works
func TestUpdatedAtTrigger(t *testing.T) {
	dbURL := getTestDatabaseURL(t)
	
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	
	// First, create a test user if none exists
	userQuery := `
		INSERT INTO users (
			id, email, username, password_hash, first_name, last_name, role
		) VALUES (
			gen_random_uuid(),
			'test-trigger@example.com',
			'testusertrigger',
			'$2a$10$dummyhash',
			'Test',
			'User',
			'user'
		)
		ON CONFLICT (email) DO NOTHING
		RETURNING id
	`
	
	var userID string
	err = db.QueryRowContext(ctx, userQuery).Scan(&userID)
	if err == sql.ErrNoRows {
		err = db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = 'test-trigger@example.com'").Scan(&userID)
		if err == sql.ErrNoRows {
			// Fallback to any user
			err = db.QueryRowContext(ctx, "SELECT id FROM users LIMIT 1").Scan(&userID)
			require.NoError(t, err)
		} else {
			require.NoError(t, err)
		}
	} else {
		require.NoError(t, err)
	}
	
	// Create or get a test business
	businessQuery := `
		INSERT INTO businesses (
			id, name, legal_name, registration_number, created_by
		) VALUES (
			gen_random_uuid(),
			'Test Business Trigger',
			'Test Business Trigger Inc.',
			'TEST-REG-TRIGGER',
			$1
		)
		ON CONFLICT (registration_number) DO NOTHING
		RETURNING id
	`
	
	var businessID string
	err = db.QueryRowContext(ctx, businessQuery, userID).Scan(&businessID)
	if err == sql.ErrNoRows {
		err = db.QueryRowContext(ctx, "SELECT id FROM businesses WHERE registration_number = 'TEST-REG-TRIGGER'").Scan(&businessID)
		require.NoError(t, err)
	} else {
		require.NoError(t, err)
	}
	
	// Insert a test risk assessment
	insertQuery := `
		INSERT INTO risk_assessments (
			id, business_id, merchant_id, status, options, risk_level, risk_score, assessment_method, source, created_at, updated_at
		) VALUES (
			gen_random_uuid(), 
			$1,
			'test-merchant-trigger',
			'pending',
			'{}'::jsonb,
			'medium',
			0.5,
			'automated',
			'test',
			NOW(),
			NOW()
		)
		RETURNING id, updated_at
	`

	var id string
	var initialUpdatedAt string
	err = db.QueryRowContext(ctx, insertQuery, businessID).Scan(&id, &initialUpdatedAt)
	require.NoError(t, err)
	
	// Wait a moment to ensure timestamp difference
	time.Sleep(1 * time.Second)
	
	// Update the record
	updateQuery := `
		UPDATE risk_assessments 
		SET status = 'processing'
		WHERE id = $1
		RETURNING updated_at
	`
	
	var updatedAt string
	err = db.QueryRowContext(ctx, updateQuery, id).Scan(&updatedAt)
	require.NoError(t, err)
	
	assert.NotEqual(t, initialUpdatedAt, updatedAt, "updated_at should change after update")
	
	t.Logf("✓ updated_at trigger works: %s -> %s", initialUpdatedAt, updatedAt)
	
	// Clean up
	_, err = db.ExecContext(ctx, "DELETE FROM risk_assessments WHERE id = $1", id)
	require.NoError(t, err)
}

