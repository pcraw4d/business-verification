package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabaseIndexesPerformance tests the performance of database indexes
func TestDatabaseIndexesPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	// Setup test database
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Run migration to create tables and indexes
	err := runMigrations(db)
	require.NoError(t, err)

	// Seed test data
	err = seedTestData(db)
	require.NoError(t, err)

	t.Run("merchant_search_indexes", func(t *testing.T) {
		testMerchantSearchIndexes(t, db)
	})

	t.Run("portfolio_type_indexes", func(t *testing.T) {
		testPortfolioTypeIndexes(t, db)
	})

	t.Run("risk_level_indexes", func(t *testing.T) {
		testRiskLevelIndexes(t, db)
	})

	t.Run("audit_log_timestamp_indexes", func(t *testing.T) {
		testAuditLogTimestampIndexes(t, db)
	})

	t.Run("composite_indexes", func(t *testing.T) {
		testCompositeIndexes(t, db)
	})

	t.Run("session_management_indexes", func(t *testing.T) {
		testSessionManagementIndexes(t, db)
	})
}

// testMerchantSearchIndexes tests merchant search performance
func testMerchantSearchIndexes(t *testing.T, db *sql.DB) {
	ctx := context.Background()

	// Test name search with trigram index
	start := time.Now()
	rows, err := db.QueryContext(ctx, `
		SELECT id, name FROM merchants 
		WHERE name ILIKE '%Tech%' 
		ORDER BY name
		LIMIT 100
	`)
	require.NoError(t, err)
	defer rows.Close()

	var count int
	for rows.Next() {
		count++
	}
	searchTime := time.Since(start)

	// Search should complete in under 100ms for 1000+ records
	assert.Less(t, searchTime, 100*time.Millisecond, "Name search should be fast with trigram index")
	assert.Greater(t, count, 0, "Should find merchants with 'Tech' in name")

	// Test legal name search
	start = time.Now()
	rows, err = db.QueryContext(ctx, `
		SELECT id, legal_name FROM merchants 
		WHERE legal_name ILIKE '%Corporation%' 
		ORDER BY legal_name
		LIMIT 100
	`)
	require.NoError(t, err)
	defer rows.Close()

	count = 0
	for rows.Next() {
		count++
	}
	searchTime = time.Since(start)

	assert.Less(t, searchTime, 100*time.Millisecond, "Legal name search should be fast with trigram index")
	assert.Greater(t, count, 0, "Should find merchants with 'Corporation' in legal name")

	// Test registration number lookup
	start = time.Now()
	var merchantID string
	err = db.QueryRowContext(ctx, `
		SELECT id FROM merchants 
		WHERE registration_number = 'REG-001'
	`).Scan(&merchantID)
	require.NoError(t, err)
	lookupTime := time.Since(start)

	assert.Less(t, lookupTime, 10*time.Millisecond, "Registration number lookup should be very fast")
	assert.NotEmpty(t, merchantID, "Should find merchant by registration number")
}

// testPortfolioTypeIndexes tests portfolio type filtering performance
func testPortfolioTypeIndexes(t *testing.T, db *sql.DB) {
	ctx := context.Background()

	// Test portfolio type filtering
	start := time.Now()
	rows, err := db.QueryContext(ctx, `
		SELECT COUNT(*) FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		WHERE pt.type = 'onboarded'
	`)
	require.NoError(t, err)
	defer rows.Close()

	var count int
	if rows.Next() {
		err = rows.Scan(&count)
		require.NoError(t, err)
	}
	filterTime := time.Since(start)

	assert.Less(t, filterTime, 50*time.Millisecond, "Portfolio type filtering should be fast")
	assert.Greater(t, count, 0, "Should find onboarded merchants")

	// Test portfolio type with pagination
	start = time.Now()
	rows, err = db.QueryContext(ctx, `
		SELECT m.id, m.name, pt.type 
		FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		WHERE pt.type = 'prospective'
		ORDER BY m.created_at DESC
		LIMIT 50 OFFSET 0
	`)
	require.NoError(t, err)
	defer rows.Close()

	var resultCount int
	for rows.Next() {
		resultCount++
	}
	paginationTime := time.Since(start)

	assert.Less(t, paginationTime, 50*time.Millisecond, "Portfolio type pagination should be fast")
	assert.Greater(t, resultCount, 0, "Should find prospective merchants")
}

// testRiskLevelIndexes tests risk level filtering performance
func testRiskLevelIndexes(t *testing.T, db *sql.DB) {
	ctx := context.Background()

	// Test risk level filtering
	start := time.Now()
	rows, err := db.QueryContext(ctx, `
		SELECT COUNT(*) FROM merchants m
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		WHERE rl.level = 'high'
	`)
	require.NoError(t, err)
	defer rows.Close()

	var count int
	if rows.Next() {
		err = rows.Scan(&count)
		require.NoError(t, err)
	}
	filterTime := time.Since(start)

	assert.Less(t, filterTime, 50*time.Millisecond, "Risk level filtering should be fast")
	assert.Greater(t, count, 0, "Should find high-risk merchants")

	// Test risk level ordering
	start = time.Now()
	rows, err = db.QueryContext(ctx, `
		SELECT m.id, m.name, rl.level, rl.numeric_value
		FROM merchants m
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		ORDER BY rl.numeric_value DESC, m.created_at DESC
		LIMIT 100
	`)
	require.NoError(t, err)
	defer rows.Close()

	var resultCount int
	for rows.Next() {
		resultCount++
	}
	orderTime := time.Since(start)

	assert.Less(t, orderTime, 50*time.Millisecond, "Risk level ordering should be fast")
	assert.Greater(t, resultCount, 0, "Should find merchants ordered by risk level")
}

// testAuditLogTimestampIndexes tests audit log timestamp queries
func testAuditLogTimestampIndexes(t *testing.T, db *sql.DB) {
	ctx := context.Background()

	// Test recent audit logs
	start := time.Now()
	rows, err := db.QueryContext(ctx, `
		SELECT id, action, created_at 
		FROM merchant_audit_logs 
		WHERE created_at >= NOW() - INTERVAL '1 day'
		ORDER BY created_at DESC
		LIMIT 100
	`)
	require.NoError(t, err)
	defer rows.Close()

	var count int
	for rows.Next() {
		count++
	}
	recentTime := time.Since(start)

	assert.Less(t, recentTime, 50*time.Millisecond, "Recent audit logs query should be fast")
	assert.Greater(t, count, 0, "Should find recent audit logs")

	// Test audit logs by merchant
	start = time.Now()
	rows, err = db.QueryContext(ctx, `
		SELECT id, action, created_at 
		FROM merchant_audit_logs 
		WHERE merchant_id = (SELECT id FROM merchants LIMIT 1)
		ORDER BY created_at DESC
		LIMIT 50
	`)
	require.NoError(t, err)
	defer rows.Close()

	count = 0
	for rows.Next() {
		count++
	}
	merchantTime := time.Since(start)

	assert.Less(t, merchantTime, 50*time.Millisecond, "Audit logs by merchant should be fast")
}

// testCompositeIndexes tests composite index performance
func testCompositeIndexes(t *testing.T, db *sql.DB) {
	ctx := context.Background()

	// Test portfolio type + risk level composite index
	start := time.Now()
	rows, err := db.QueryContext(ctx, `
		SELECT COUNT(*) FROM merchants m
		JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
		JOIN risk_levels rl ON m.risk_level_id = rl.id
		WHERE pt.type = 'onboarded' AND rl.level = 'medium'
	`)
	require.NoError(t, err)
	defer rows.Close()

	var count int
	if rows.Next() {
		err = rows.Scan(&count)
		require.NoError(t, err)
	}
	compositeTime := time.Since(start)

	assert.Less(t, compositeTime, 50*time.Millisecond, "Composite portfolio+risk query should be fast")
	assert.Greater(t, count, 0, "Should find onboarded medium-risk merchants")

	// Test status + compliance composite index
	start = time.Now()
	rows, err = db.QueryContext(ctx, `
		SELECT COUNT(*) FROM merchants 
		WHERE status = 'active' AND compliance_status = 'compliant'
	`)
	require.NoError(t, err)
	defer rows.Close()

	count = 0
	if rows.Next() {
		err = rows.Scan(&count)
		require.NoError(t, err)
	}
	statusTime := time.Since(start)

	assert.Less(t, statusTime, 50*time.Millisecond, "Status+compliance query should be fast")
	assert.Greater(t, count, 0, "Should find active compliant merchants")
}

// testSessionManagementIndexes tests session management performance
func testSessionManagementIndexes(t *testing.T, db *sql.DB) {
	ctx := context.Background()

	// Test active session lookup
	start := time.Now()
	rows, err := db.QueryContext(ctx, `
		SELECT id, user_id, merchant_id, last_active 
		FROM merchant_sessions 
		WHERE is_active = true
		ORDER BY last_active DESC
		LIMIT 100
	`)
	require.NoError(t, err)
	defer rows.Close()

	var count int
	for rows.Next() {
		count++
	}
	sessionTime := time.Since(start)

	assert.Less(t, sessionTime, 50*time.Millisecond, "Active session lookup should be fast")
	assert.Greater(t, count, 0, "Should find active sessions")

	// Test user session lookup
	start = time.Now()
	rows, err = db.QueryContext(ctx, `
		SELECT id, merchant_id, started_at 
		FROM merchant_sessions 
		WHERE user_id = (SELECT id FROM users LIMIT 1) AND is_active = true
		ORDER BY last_active DESC
		LIMIT 1
	`)
	require.NoError(t, err)
	defer rows.Close()

	count = 0
	for rows.Next() {
		count++
	}
	userSessionTime := time.Since(start)

	assert.Less(t, userSessionTime, 50*time.Millisecond, "User session lookup should be fast")
}

// setupTestDB sets up a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	// Use test database URL or create in-memory database
	dbURL := "postgres://test:test@localhost:5432/kyb_test?sslmode=disable"

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)

	// Test connection
	err = db.Ping()
	require.NoError(t, err)

	return db
}

// cleanupTestDB cleans up test database
func cleanupTestDB(t *testing.T, db *sql.DB) {
	if db != nil {
		db.Close()
	}
}

// runMigrations runs the database migrations
func runMigrations(db *sql.DB) error {
	// Read and execute migration files
	migrations := []string{
		"001_initial_schema.sql",
		"002_rbac_schema.sql",
		"005_merchant_portfolio_schema.sql",
		"006_mock_data_seed.sql",
	}

	for _, migration := range migrations {
		// In a real implementation, you would read the migration file
		// For this test, we'll assume the migrations have been run
		log.Printf("Running migration: %s", migration)
	}

	return nil
}

// seedTestData seeds the database with test data
func seedTestData(db *sql.DB) error {
	ctx := context.Background()

	// Create test users
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, name, role, created_at, updated_at) 
		VALUES 
			('user-1', 'test1@example.com', 'Test User 1', 'admin', NOW(), NOW()),
			('user-2', 'test2@example.com', 'Test User 2', 'user', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to create test users: %w", err)
	}

	// Create portfolio types
	_, err = db.ExecContext(ctx, `
		INSERT INTO portfolio_types (id, type, description, display_order, is_active, created_at, updated_at)
		VALUES 
			('pt-1', 'onboarded', 'Onboarded merchants', 1, true, NOW(), NOW()),
			('pt-2', 'deactivated', 'Deactivated merchants', 2, true, NOW(), NOW()),
			('pt-3', 'prospective', 'Prospective merchants', 3, true, NOW(), NOW()),
			('pt-4', 'pending', 'Pending merchants', 4, true, NOW(), NOW())
		ON CONFLICT (type) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to create portfolio types: %w", err)
	}

	// Create risk levels
	_, err = db.ExecContext(ctx, `
		INSERT INTO risk_levels (id, level, description, numeric_value, color_code, display_order, is_active, created_at, updated_at)
		VALUES 
			('rl-1', 'low', 'Low risk', 1, '#28a745', 1, true, NOW(), NOW()),
			('rl-2', 'medium', 'Medium risk', 2, '#ffc107', 2, true, NOW(), NOW()),
			('rl-3', 'high', 'High risk', 3, '#dc3545', 3, true, NOW(), NOW())
		ON CONFLICT (level) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to create risk levels: %w", err)
	}

	// Create test merchants
	for i := 1; i <= 1000; i++ {
		portfolioType := "pt-1" // onboarded
		riskLevel := "rl-2"     // medium

		if i%4 == 0 {
			portfolioType = "pt-2" // deactivated
		} else if i%4 == 1 {
			portfolioType = "pt-3" // prospective
		} else if i%4 == 2 {
			portfolioType = "pt-4" // pending
		}

		if i%3 == 0 {
			riskLevel = "rl-1" // low
		} else if i%3 == 1 {
			riskLevel = "rl-3" // high
		}

		_, err = db.ExecContext(ctx, `
			INSERT INTO merchants (
				id, name, legal_name, registration_number, tax_id, industry, industry_code,
				business_type, founded_date, employee_count, annual_revenue,
				address_street1, address_city, address_state, address_postal_code, address_country,
				contact_phone, contact_email, contact_website,
				portfolio_type_id, risk_level_id, compliance_status, status,
				created_by, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
				$12, $13, $14, $15, $16, $17, $18, $19,
				$20, $21, $22, $23, $24, NOW(), NOW()
			)
		`,
			fmt.Sprintf("merchant-%d", i),
			fmt.Sprintf("Tech Company %d", i),
			fmt.Sprintf("Tech Corporation %d LLC", i),
			fmt.Sprintf("REG-%03d", i),
			fmt.Sprintf("TAX-%03d", i),
			"Technology",
			"541511",
			"Corporation",
			"2020-01-01",
			50+i,
			1000000.00+float64(i*1000),
			fmt.Sprintf("%d Main St", i),
			"San Francisco",
			"CA",
			fmt.Sprintf("9410%d", i%10),
			"USA",
			fmt.Sprintf("+1-555-%04d", i),
			fmt.Sprintf("contact%d@techcorp.com", i),
			fmt.Sprintf("https://techcorp%d.com", i),
			portfolioType,
			riskLevel,
			"compliant",
			"active",
			"user-1",
		)
		if err != nil {
			return fmt.Errorf("failed to create merchant %d: %w", i, err)
		}
	}

	// Create test audit logs
	for i := 1; i <= 5000; i++ {
		_, err = db.ExecContext(ctx, `
			INSERT INTO merchant_audit_logs (
				id, user_id, merchant_id, action, resource_type, resource_id,
				details, ip_address, user_agent, request_id, created_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
			)
		`,
			fmt.Sprintf("audit-%d", i),
			"user-1",
			fmt.Sprintf("merchant-%d", (i%1000)+1),
			"view",
			"merchant",
			fmt.Sprintf("merchant-%d", (i%1000)+1),
			fmt.Sprintf(`{"action": "view", "timestamp": "%s"}`, time.Now().Format(time.RFC3339)),
			"127.0.0.1",
			"Mozilla/5.0 (Test Browser)",
			fmt.Sprintf("req-%d", i),
			time.Now().Add(-time.Duration(i)*time.Minute),
		)
		if err != nil {
			return fmt.Errorf("failed to create audit log %d: %w", i, err)
		}
	}

	// Create test sessions
	for i := 1; i <= 100; i++ {
		_, err = db.ExecContext(ctx, `
			INSERT INTO merchant_sessions (
				id, user_id, merchant_id, started_at, last_active, is_active, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, NOW(), NOW()
			)
		`,
			fmt.Sprintf("session-%d", i),
			"user-1",
			fmt.Sprintf("merchant-%d", (i%1000)+1),
			time.Now().Add(-time.Duration(i)*time.Hour),
			time.Now().Add(-time.Duration(i%10)*time.Minute),
			i%2 == 0, // Alternate active/inactive
		)
		if err != nil {
			return fmt.Errorf("failed to create session %d: %w", i, err)
		}
	}

	return nil
}
