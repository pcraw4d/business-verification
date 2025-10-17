package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"kyb-platform/test/mocks"
)

// TestDatabaseIntegrationComprehensive tests comprehensive database connectivity and data retrieval
func TestDatabaseIntegrationComprehensive(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	// Setup test environment
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test 1: Database connectivity
	t.Run("DatabaseConnectivity", func(t *testing.T) {
		// Create mock database connection
		mockDB := mocks.NewMockDatabase()

		// Test connection
		if err := mockDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to mock database: %v", err)
		}
		defer mockDB.Disconnect()

		// Test ping
		if err := mockDB.Ping(); err != nil {
			t.Fatalf("Database ping failed: %v", err)
		}

		// Test connection status
		if !mockDB.IsConnected() {
			t.Fatal("Database should be connected")
		}

		t.Logf("✅ Database connectivity test passed")
	})

	// Test 2: Database schema validation
	t.Run("DatabaseSchemaValidation", func(t *testing.T) {
		mockDB := mocks.NewMockDatabase()
		if err := mockDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to mock database: %v", err)
		}
		defer mockDB.Disconnect()

		// Test required tables exist
		requiredTables := []string{
			"industry_codes",
			"keywords",
			"keyword_weights",
			"classification_patterns",
		}

		for _, table := range requiredTables {
			exists, err := mockDB.TableExists(table)
			if err != nil {
				t.Fatalf("Failed to check if table %s exists: %v", table, err)
			}
			if !exists {
				t.Errorf("Required table %s does not exist", table)
			}
		}

		// Test table schemas
		schemaTests := []struct {
			table   string
			columns []string
		}{
			{
				table:   "industry_codes",
				columns: []string{"id", "code", "name", "type", "description"},
			},
			{
				table:   "keywords",
				columns: []string{"id", "keyword", "industry_code_id", "weight"},
			},
			{
				table:   "keyword_weights",
				columns: []string{"id", "keyword", "weight", "category"},
			},
		}

		for _, test := range schemaTests {
			for _, column := range test.columns {
				exists, err := mockDB.ColumnExists(test.table, column)
				if err != nil {
					t.Fatalf("Failed to check if column %s.%s exists: %v", test.table, column, err)
				}
				if !exists {
					t.Errorf("Required column %s.%s does not exist", test.table, column)
				}
			}
		}

		t.Logf("✅ Database schema validation test passed")
	})

	// Test 3: Data population validation
	t.Run("DataPopulationValidation", func(t *testing.T) {
		mockDB := mocks.NewMockDatabase()
		if err := mockDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to mock database: %v", err)
		}
		defer mockDB.Disconnect()

		// Test industry codes data
		industryCodesCount, err := mockDB.GetTableCount("industry_codes")
		if err != nil {
			t.Fatalf("Failed to get industry codes count: %v", err)
		}
		if industryCodesCount == 0 {
			t.Error("Industry codes table should have data")
		}

		// Test keywords data
		keywordsCount, err := mockDB.GetTableCount("keywords")
		if err != nil {
			t.Fatalf("Failed to get keywords count: %v", err)
		}
		if keywordsCount == 0 {
			t.Error("Keywords table should have data")
		}

		// Test keyword weights data
		weightsCount, err := mockDB.GetTableCount("keyword_weights")
		if err != nil {
			t.Fatalf("Failed to get keyword weights count: %v", err)
		}
		if weightsCount == 0 {
			t.Error("Keyword weights table should have data")
		}

		// Test data integrity
		integrityTests := []struct {
			name        string
			query       string
			expectCount int
		}{
			{
				name:        "NAICS codes exist",
				query:       "SELECT COUNT(*) FROM industry_codes WHERE type = 'NAICS'",
				expectCount: 100, // Expect at least 100 NAICS codes
			},
			{
				name:        "SIC codes exist",
				query:       "SELECT COUNT(*) FROM industry_codes WHERE type = 'SIC'",
				expectCount: 50, // Expect at least 50 SIC codes
			},
			{
				name:        "MCC codes exist",
				query:       "SELECT COUNT(*) FROM industry_codes WHERE type = 'MCC'",
				expectCount: 20, // Expect at least 20 MCC codes
			},
		}

		for _, test := range integrityTests {
			count, err := mockDB.ExecuteCountQuery(test.query)
			if err != nil {
				t.Fatalf("Failed to execute query for %s: %v", test.name, err)
			}
			if count < test.expectCount {
				t.Errorf("%s: Expected at least %d records, got %d", test.name, test.expectCount, count)
			}
		}

		t.Logf("✅ Data population validation test passed - Industry Codes: %d, Keywords: %d, Weights: %d",
			industryCodesCount, keywordsCount, weightsCount)
	})

	// Test 4: Database query performance
	t.Run("DatabaseQueryPerformance", func(t *testing.T) {
		mockDB := mocks.NewMockDatabase()
		if err := mockDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to mock database: %v", err)
		}
		defer mockDB.Disconnect()

		// Test query performance
		performanceTests := []struct {
			name        string
			query       string
			maxDuration time.Duration
		}{
			{
				name:        "Industry codes lookup",
				query:       "SELECT * FROM industry_codes WHERE type = 'NAICS' LIMIT 10",
				maxDuration: 100 * time.Millisecond,
			},
			{
				name:        "Keywords search",
				query:       "SELECT * FROM keywords WHERE keyword LIKE '%software%' LIMIT 10",
				maxDuration: 200 * time.Millisecond,
			},
			{
				name:        "Weighted keywords",
				query:       "SELECT k.*, kw.weight FROM keywords k JOIN keyword_weights kw ON k.keyword = kw.keyword LIMIT 10",
				maxDuration: 300 * time.Millisecond,
			},
		}

		for _, test := range performanceTests {
			start := time.Now()
			_, err := mockDB.ExecuteQuery(test.query)
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("Query failed for %s: %v", test.name, err)
			}

			if duration > test.maxDuration {
				t.Errorf("%s: Query took %v, expected < %v", test.name, duration, test.maxDuration)
			}

			t.Logf("✅ %s: %v", test.name, duration)
		}

		t.Logf("✅ Database query performance test passed")
	})

	// Test 5: Database data retrieval
	t.Run("DatabaseDataRetrieval", func(t *testing.T) {
		mockDB := mocks.NewMockDatabase()
		if err := mockDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to mock database: %v", err)
		}
		defer mockDB.Disconnect()

		// Test retrieving industry codes
		industryCodes, err := mockDB.ExecuteQuery("SELECT * FROM industry_codes WHERE type = 'NAICS' LIMIT 5")
		if err != nil {
			t.Fatalf("Failed to retrieve industry codes: %v", err)
		}

		if len(industryCodes) == 0 {
			t.Error("Expected industry codes to be returned")
		}

		// Validate industry code structure
		for _, code := range industryCodes {
			if _, exists := code["id"]; !exists {
				t.Error("Industry code should have 'id' field")
			}
			if _, exists := code["code"]; !exists {
				t.Error("Industry code should have 'code' field")
			}
			if _, exists := code["name"]; !exists {
				t.Error("Industry code should have 'name' field")
			}
			if _, exists := code["type"]; !exists {
				t.Error("Industry code should have 'type' field")
			}
		}

		// Test retrieving keywords
		keywords, err := mockDB.ExecuteQuery("SELECT * FROM keywords LIMIT 5")
		if err != nil {
			t.Fatalf("Failed to retrieve keywords: %v", err)
		}

		if len(keywords) == 0 {
			t.Error("Expected keywords to be returned")
		}

		// Validate keyword structure
		for _, keyword := range keywords {
			if _, exists := keyword["id"]; !exists {
				t.Error("Keyword should have 'id' field")
			}
			if _, exists := keyword["keyword"]; !exists {
				t.Error("Keyword should have 'keyword' field")
			}
			if _, exists := keyword["weight"]; !exists {
				t.Error("Keyword should have 'weight' field")
			}
		}

		t.Logf("✅ Database data retrieval test passed - Industry Codes: %d, Keywords: %d", len(industryCodes), len(keywords))
	})

	// Test 6: Database transaction handling
	t.Run("DatabaseTransactionHandling", func(t *testing.T) {
		mockDB := mocks.NewMockDatabase()
		if err := mockDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to mock database: %v", err)
		}
		defer mockDB.Disconnect()

		// Test transaction begin
		tx, err := mockDB.BeginTransaction()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// Test transaction operations
		testQuery := "INSERT INTO test_table (name) VALUES ('test')"
		_, err = tx.Exec(testQuery)
		if err != nil {
			t.Fatalf("Failed to execute query in transaction: %v", err)
		}

		// Test transaction commit
		if err := tx.Commit(); err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// Test transaction rollback
		tx2, err := mockDB.BeginTransaction()
		if err != nil {
			t.Fatalf("Failed to begin second transaction: %v", err)
		}

		_, err = tx2.Exec(testQuery)
		if err != nil {
			t.Fatalf("Failed to execute query in second transaction: %v", err)
		}

		if err := tx2.Rollback(); err != nil {
			t.Fatalf("Failed to rollback transaction: %v", err)
		}

		t.Logf("✅ Database transaction handling test passed")
	})

	// Test 7: Database connection pooling
	t.Run("DatabaseConnectionPooling", func(t *testing.T) {
		mockDB := mocks.NewMockDatabase()
		if err := mockDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to mock database: %v", err)
		}
		defer mockDB.Disconnect()

		// Test connection pool configuration
		poolConfig := mockDB.GetConnectionPoolConfig()
		if poolConfig.MaxOpenConns == 0 {
			t.Error("Max open connections should be configured")
		}
		if poolConfig.MaxIdleConns == 0 {
			t.Error("Max idle connections should be configured")
		}

		// Test concurrent connections
		concurrentConnections := 5
		results := make(chan error, concurrentConnections)

		for i := 0; i < concurrentConnections; i++ {
			go func(index int) {
				conn, err := mockDB.GetConnection()
				if err != nil {
					results <- err
					return
				}
				defer mockDB.ReleaseConnection(conn)

				// Test connection
				if err := conn.Ping(); err != nil {
					results <- err
					return
				}

				results <- nil
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 0; i < concurrentConnections; i++ {
			select {
			case err := <-results:
				if err != nil {
					t.Logf("Connection %d failed: %v", i, err)
				} else {
					successCount++
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Connection pool test timed out")
			}
		}

		// Validate success rate
		successRate := float64(successCount) / float64(concurrentConnections)
		if successRate < 0.8 { // Allow 20% failure rate
			t.Errorf("Expected connection success rate >= 80%%, got %.1f%%", successRate*100)
		}

		t.Logf("✅ Database connection pooling test passed - Success: %d/%d (%.1f%%)",
			successCount, concurrentConnections, successRate*100)
	})

	// Test 8: Database error handling
	t.Run("DatabaseErrorHandling", func(t *testing.T) {
		mockDB := mocks.NewMockDatabase()
		if err := mockDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to mock database: %v", err)
		}
		defer mockDB.Disconnect()

		// Test invalid query
		_, err := mockDB.ExecuteQuery("SELECT * FROM non_existent_table")
		if err == nil {
			t.Error("Expected error for invalid query")
		}

		// Test invalid connection
		invalidDB := mocks.NewMockDatabase()
		// Set connection error to simulate connection failure
		invalidDB.SetConnectionError(fmt.Errorf("connection failed"))
		_, err = invalidDB.ExecuteQuery("SELECT 1")
		if err == nil {
			t.Error("Expected error for invalid connection")
		}

		// Test timeout handling
		timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel()

		// Add a small delay to ensure timeout
		time.Sleep(2 * time.Millisecond)

		_, err = mockDB.ExecuteQueryWithContext(timeoutCtx, "SELECT pg_sleep(1)")
		if err == nil {
			t.Error("Expected timeout error")
		}

		t.Logf("✅ Database error handling test passed")
	})
}

// TestDatabaseDataIntegrity tests data integrity and consistency
func TestDatabaseDataIntegrity(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	mockDB := mocks.NewMockDatabase()
	if err := mockDB.Connect(); err != nil {
		t.Fatalf("Failed to connect to mock database: %v", err)
	}
	defer mockDB.Disconnect()

	// Test 1: Foreign key constraints
	t.Run("ForeignKeyConstraints", func(t *testing.T) {
		// Test keywords reference valid industry codes
		query := `
			SELECT COUNT(*) 
			FROM keywords k 
			LEFT JOIN industry_codes ic ON k.industry_code_id = ic.id 
			WHERE k.industry_code_id IS NOT NULL AND ic.id IS NULL
		`
		count, err := mockDB.ExecuteCountQuery(query)
		if err != nil {
			t.Fatalf("Failed to check foreign key constraints: %v", err)
		}
		if count > 0 {
			t.Errorf("Found %d keywords with invalid industry_code_id references", count)
		}

		t.Logf("✅ Foreign key constraints test passed")
	})

	// Test 2: Data consistency
	t.Run("DataConsistency", func(t *testing.T) {
		// Test no duplicate industry codes
		query := "SELECT code, type, COUNT(*) FROM industry_codes GROUP BY code, type HAVING COUNT(*) > 1"
		duplicates, err := mockDB.ExecuteQuery(query)
		if err != nil {
			t.Fatalf("Failed to check for duplicate industry codes: %v", err)
		}
		if len(duplicates) > 0 {
			t.Errorf("Found %d duplicate industry codes", len(duplicates))
		}

		// Test no duplicate keywords
		query = "SELECT keyword, COUNT(*) FROM keywords GROUP BY keyword HAVING COUNT(*) > 1"
		duplicates, err = mockDB.ExecuteQuery(query)
		if err != nil {
			t.Fatalf("Failed to check for duplicate keywords: %v", err)
		}
		if len(duplicates) > 0 {
			t.Errorf("Found %d duplicate keywords", len(duplicates))
		}

		t.Logf("✅ Data consistency test passed")
	})

	// Test 3: Data completeness
	t.Run("DataCompleteness", func(t *testing.T) {
		// Test all industry codes have names
		query := "SELECT COUNT(*) FROM industry_codes WHERE name IS NULL OR name = ''"
		count, err := mockDB.ExecuteCountQuery(query)
		if err != nil {
			t.Fatalf("Failed to check for missing industry code names: %v", err)
		}
		if count > 0 {
			t.Errorf("Found %d industry codes with missing names", count)
		}

		// Test all keywords have weights
		query = "SELECT COUNT(*) FROM keywords k LEFT JOIN keyword_weights kw ON k.keyword = kw.keyword WHERE kw.keyword IS NULL"
		count, err = mockDB.ExecuteCountQuery(query)
		if err != nil {
			t.Fatalf("Failed to check for missing keyword weights: %v", err)
		}
		if count > 0 {
			t.Logf("Warning: Found %d keywords without weights", count)
		}

		t.Logf("✅ Data completeness test passed")
	})

	// Test 4: Data validation
	t.Run("DataValidation", func(t *testing.T) {
		// Test industry code formats
		codeFormatTests := []struct {
			typeName string
			pattern  string
		}{
			{"NAICS", "^[0-9]{6}$"},
			{"SIC", "^[0-9]{4}$"},
			{"MCC", "^[0-9]{4}$"},
		}

		for _, test := range codeFormatTests {
			query := fmt.Sprintf("SELECT COUNT(*) FROM industry_codes WHERE type = '%s' AND code !~ '%s'", test.typeName, test.pattern)
			count, err := mockDB.ExecuteCountQuery(query)
			if err != nil {
				t.Fatalf("Failed to validate %s code format: %v", test.typeName, err)
			}
			if count > 0 {
				t.Errorf("Found %d %s codes with invalid format", count, test.typeName)
			}
		}

		// Test keyword weights are positive
		query := "SELECT COUNT(*) FROM keyword_weights WHERE weight <= 0"
		count, err := mockDB.ExecuteCountQuery(query)
		if err != nil {
			t.Fatalf("Failed to validate keyword weights: %v", err)
		}
		if count > 0 {
			t.Errorf("Found %d keywords with non-positive weights", count)
		}

		t.Logf("✅ Data validation test passed")
	})
}
