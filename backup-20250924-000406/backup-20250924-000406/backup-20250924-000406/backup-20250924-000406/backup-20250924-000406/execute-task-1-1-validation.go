package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// getDatabaseConfig loads database configuration from environment variables
func getDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "db.qpqhuqqmkjxsltzshfam.supabase.co"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USERNAME", "postgres"),
		Password: getEnv("DB_PASSWORD", "Geaux44tigers!"),
		DBName:   getEnv("DB_DATABASE", "postgres"),
		SSLMode:  getEnv("DB_SSL_MODE", "require"),
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// connectToDatabase establishes a connection to the database
func connectToDatabase(config DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// executeTask11 executes Task 1.1: Fix Database Schema Issues
func executeTask11(db *sql.DB) error {
	fmt.Println("🚀 Executing Task 1.1: Fix Database Schema Issues")
	fmt.Println(strings.Repeat("=", 60))

	// Step 1: Add missing is_active column
	fmt.Println("\n📊 Step 1: Adding missing is_active column...")
	_, err := db.Exec("ALTER TABLE keyword_weights ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;")
	if err != nil {
		return fmt.Errorf("failed to add is_active column: %w", err)
	}
	fmt.Println("   ✅ is_active column added successfully")

	// Step 2: Update existing records
	fmt.Println("\n📊 Step 2: Updating existing records...")
	result, err := db.Exec("UPDATE keyword_weights SET is_active = true WHERE is_active IS NULL;")
	if err != nil {
		return fmt.Errorf("failed to update existing records: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	fmt.Printf("   ✅ Updated %d records to is_active = true\n", rowsAffected)

	// Step 3: Create performance indexes
	fmt.Println("\n📊 Step 3: Creating performance indexes...")

	// Index 1: Basic is_active index
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_keyword_weights_active ON keyword_weights(is_active);")
	if err != nil {
		return fmt.Errorf("failed to create idx_keyword_weights_active: %w", err)
	}
	fmt.Println("   ✅ Created idx_keyword_weights_active index")

	// Index 2: Composite industry_id + is_active index
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_keyword_weights_industry_active ON keyword_weights(industry_id, is_active);")
	if err != nil {
		return fmt.Errorf("failed to create idx_keyword_weights_industry_active: %w", err)
	}
	fmt.Println("   ✅ Created idx_keyword_weights_industry_active index")

	// Enhanced indexes (professional best practices)
	fmt.Println("\n📊 Step 4: Creating enhanced indexes...")

	// Index 3: Keyword + is_active composite index
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_keyword_weights_keyword_active ON keyword_weights (keyword, is_active) WHERE is_active = true;")
	if err != nil {
		return fmt.Errorf("failed to create idx_keyword_weights_keyword_active: %w", err)
	}
	fmt.Println("   ✅ Created idx_keyword_weights_keyword_active index")

	// Index 4: Industry + is_active + weight ordering
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_keyword_weights_industry_weight_active ON keyword_weights (industry_id, is_active, base_weight DESC) WHERE is_active = true;")
	if err != nil {
		return fmt.Errorf("failed to create idx_keyword_weights_industry_weight_active: %w", err)
	}
	fmt.Println("   ✅ Created idx_keyword_weights_industry_weight_active index")

	// Index 5: Search optimization index
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_keyword_weights_search_active ON keyword_weights (is_active, base_weight DESC, keyword) WHERE is_active = true;")
	if err != nil {
		return fmt.Errorf("failed to create idx_keyword_weights_search_active: %w", err)
	}
	fmt.Println("   ✅ Created idx_keyword_weights_search_active index")

	// Update table statistics
	fmt.Println("\n📊 Step 5: Updating table statistics...")
	_, err = db.Exec("ANALYZE keyword_weights;")
	if err != nil {
		return fmt.Errorf("failed to analyze keyword_weights table: %w", err)
	}
	fmt.Println("   ✅ Table statistics updated")

	return nil
}

// verifyTask11 verifies that Task 1.1 has been completed successfully
func verifyTask11(db *sql.DB) error {
	fmt.Println("\n🔍 Verifying Task 1.1 Implementation")
	fmt.Println(strings.Repeat("=", 60))

	// Test 1: Verify column exists
	fmt.Println("\n📊 Test 1: Verifying is_active column exists...")
	var columnName, dataType, isNullable string
	err := db.QueryRow(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'keyword_weights' AND column_name = 'is_active'
	`).Scan(&columnName, &dataType, &isNullable)

	if err != nil {
		return fmt.Errorf("❌ FAIL: is_active column does not exist: %w", err)
	}

	fmt.Printf("   ✅ PASS: Column exists - %s (%s, nullable: %s)\n", columnName, dataType, isNullable)

	// Test 2: Verify all records are active
	fmt.Println("\n📊 Test 2: Verifying all records are active...")
	var totalRecords, activeRecords, inactiveRecords, nullRecords int
	err = db.QueryRow(`
		SELECT 
			COUNT(*) as total_records,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_records,
			COUNT(CASE WHEN is_active = false THEN 1 END) as inactive_records,
			COUNT(CASE WHEN is_active IS NULL THEN 1 END) as null_records
		FROM keyword_weights
	`).Scan(&totalRecords, &activeRecords, &inactiveRecords, &nullRecords)

	if err != nil {
		return fmt.Errorf("failed to check record status: %w", err)
	}

	fmt.Printf("   📊 Total Records: %d\n", totalRecords)
	fmt.Printf("   📊 Active Records: %d\n", activeRecords)
	fmt.Printf("   📊 Inactive Records: %d\n", inactiveRecords)
	fmt.Printf("   📊 NULL Records: %d\n", nullRecords)

	if nullRecords > 0 {
		return fmt.Errorf("❌ FAIL: Found %d records with NULL is_active values", nullRecords)
	}

	if totalRecords != activeRecords {
		return fmt.Errorf("❌ FAIL: Not all records are active (Total: %d, Active: %d)", totalRecords, activeRecords)
	}

	fmt.Println("   ✅ PASS: All records have is_active = true")

	// Test 3: Verify indexes exist
	fmt.Println("\n📊 Test 3: Verifying performance indexes exist...")
	rows, err := db.Query(`
		SELECT indexname, indexdef 
		FROM pg_indexes 
		WHERE tablename = 'keyword_weights' AND indexname LIKE '%active%'
		ORDER BY indexname
	`)
	if err != nil {
		return fmt.Errorf("failed to check indexes: %w", err)
	}
	defer rows.Close()

	indexCount := 0
	requiredIndexes := []string{
		"idx_keyword_weights_active",
		"idx_keyword_weights_industry_active",
	}

	foundIndexes := make(map[string]bool)

	for rows.Next() {
		var indexName, indexDef string
		err := rows.Scan(&indexName, &indexDef)
		if err != nil {
			return fmt.Errorf("failed to scan index: %w", err)
		}

		foundIndexes[indexName] = true
		indexCount++
		fmt.Printf("   ✅ Found index: %s\n", indexName)
	}

	// Check for required indexes
	missingIndexes := []string{}
	for _, required := range requiredIndexes {
		if !foundIndexes[required] {
			missingIndexes = append(missingIndexes, required)
		}
	}

	if len(missingIndexes) > 0 {
		return fmt.Errorf("❌ FAIL: Missing required indexes: %v", missingIndexes)
	}

	fmt.Printf("   ✅ PASS: Found %d indexes (including %d required indexes)\n", indexCount, len(requiredIndexes))

	// Test 4: Test query performance
	fmt.Println("\n📊 Test 4: Testing query performance...")

	// Test basic is_active query
	start := time.Now()
	rows, err = db.Query("SELECT COUNT(*) FROM keyword_weights WHERE is_active = true")
	if err != nil {
		return fmt.Errorf("failed to test basic query: %w", err)
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		rows.Scan(&count)
	}
	duration := time.Since(start)

	fmt.Printf("   📊 Basic query performance: %v (returned %d records)\n", duration, count)

	if duration > 100*time.Millisecond {
		fmt.Printf("   ⚠️  WARNING: Query slower than expected (%v > 100ms)\n", duration)
	} else {
		fmt.Printf("   ✅ PASS: Query performance within acceptable range\n")
	}

	// Test 5: Test industry-based query
	start = time.Now()
	rows, err = db.Query("SELECT COUNT(*) FROM keyword_weights WHERE industry_id = 1 AND is_active = true")
	if err != nil {
		return fmt.Errorf("failed to test industry query: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&count)
	}
	duration = time.Since(start)

	fmt.Printf("   📊 Industry query performance: %v (returned %d records)\n", duration, count)

	if duration > 100*time.Millisecond {
		fmt.Printf("   ⚠️  WARNING: Industry query slower than expected (%v > 100ms)\n", duration)
	} else {
		fmt.Printf("   ✅ PASS: Industry query performance within acceptable range\n")
	}

	return nil
}

func main() {
	fmt.Println("🚀 Task 1.1 Database Schema Fix - Execution and Validation")
	fmt.Println(strings.Repeat("=", 70))

	// Load configuration
	config := getDatabaseConfig()

	// Connect to database
	fmt.Println("🔌 Connecting to Supabase database...")
	db, err := connectToDatabase(config)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("✅ Connected to database successfully")

	// Execute Task 1.1
	err = executeTask11(db)
	if err != nil {
		log.Fatalf("❌ Task 1.1 execution failed: %v", err)
	}

	// Verify Task 1.1
	err = verifyTask11(db)
	if err != nil {
		log.Fatalf("❌ Task 1.1 verification failed: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🎉 Task 1.1 Execution and Validation Complete!")
	fmt.Println("✅ All database schema fixes applied successfully")
	fmt.Println("✅ All verification tests passed")
	fmt.Println("✅ Classification system ready for improved performance")
	fmt.Println("✅ Ready to proceed with Task 1.2: Add Restaurant Industry Data")
	fmt.Println(strings.Repeat("=", 70))
}
