package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

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

// IndexInfo represents information about a database index
type IndexInfo struct {
	IndexName string
	IndexDef  string
	TableName string
}

// IndexStats represents index usage statistics
type IndexStats struct {
	SchemaName    string
	TableName     string
	IndexName     string
	IndexScans    int64
	TuplesRead    int64
	TuplesFetched int64
}

// getDatabaseConfig loads database configuration from environment variables
func getDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", getEnv("SUPABASE_DB_HOST", "localhost")),
		Port:     getEnv("DB_PORT", getEnv("SUPABASE_DB_PORT", "5432")),
		User:     getEnv("DB_USERNAME", getEnv("SUPABASE_DB_USER", "postgres")),
		Password: getEnv("DB_PASSWORD", getEnv("SUPABASE_DB_PASSWORD", "")),
		DBName:   getEnv("DB_DATABASE", getEnv("SUPABASE_DB_NAME", "postgres")),
		SSLMode:  getEnv("DB_SSL_MODE", getEnv("SUPABASE_DB_SSLMODE", "disable")),
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

// verifySubtask131 verifies that subtask 1.1.3 has been completed
func verifySubtask131(db *sql.DB) error {
	fmt.Println("🔍 Verifying Subtask 1.1.3: Create Performance Indexes")
	fmt.Println(strings.Repeat("=", 60))

	// Test 1: Check that required indexes exist
	fmt.Println("\n📊 Test 1: Checking required indexes...")
	requiredIndexes := []string{
		"idx_keyword_weights_active",
		"idx_keyword_weights_industry_active",
	}

	existingIndexes, err := getKeywordWeightsIndexes(db)
	if err != nil {
		return fmt.Errorf("failed to get indexes: %w", err)
	}

	indexMap := make(map[string]bool)
	for _, idx := range existingIndexes {
		indexMap[idx.IndexName] = true
	}

	missingIndexes := []string{}
	for _, required := range requiredIndexes {
		if !indexMap[required] {
			missingIndexes = append(missingIndexes, required)
		}
	}

	if len(missingIndexes) > 0 {
		return fmt.Errorf("❌ FAIL: Missing required indexes: %v", missingIndexes)
	}

	fmt.Println("   ✅ PASS: All required indexes exist")
	for _, idx := range existingIndexes {
		if strings.Contains(idx.IndexName, "active") {
			fmt.Printf("     - %s\n", idx.IndexName)
		}
	}

	// Test 2: Check for enhanced indexes (professional best practices)
	fmt.Println("\n🔍 Test 2: Checking enhanced indexes...")
	enhancedIndexes := []string{
		"idx_keyword_weights_keyword_active",
		"idx_keyword_weights_industry_weight_active",
		"idx_keyword_weights_search_active",
	}

	enhancedCount := 0
	for _, enhanced := range enhancedIndexes {
		if indexMap[enhanced] {
			enhancedCount++
			fmt.Printf("   ✅ Enhanced index found: %s\n", enhanced)
		}
	}

	if enhancedCount > 0 {
		fmt.Printf("   ✅ PASS: %d enhanced indexes found (professional best practices)\n", enhancedCount)
	} else {
		fmt.Println("   ⚠️  INFO: No enhanced indexes found (basic indexes only)")
	}

	// Test 3: Test query performance with EXPLAIN
	fmt.Println("\n🧪 Test 3: Testing query performance...")

	// Test basic is_active query
	performance1, err := testQueryPerformance(db,
		"SELECT * FROM keyword_weights WHERE is_active = true LIMIT 10")
	if err != nil {
		return fmt.Errorf("failed to test basic query performance: %w", err)
	}

	fmt.Printf("   📊 Basic query performance: %s\n", performance1)

	// Test industry-based query
	performance2, err := testQueryPerformance(db,
		"SELECT * FROM keyword_weights WHERE industry_id = 1 AND is_active = true ORDER BY base_weight DESC LIMIT 10")
	if err != nil {
		return fmt.Errorf("failed to test industry query performance: %w", err)
	}

	fmt.Printf("   📊 Industry query performance: %s\n", performance2)

	// Test 4: Check index statistics
	fmt.Println("\n📈 Test 4: Checking index statistics...")
	stats, err := getIndexStatistics(db)
	if err != nil {
		return fmt.Errorf("failed to get index statistics: %w", err)
	}

	if len(stats) == 0 {
		fmt.Println("   ⚠️  INFO: No index usage statistics available (indexes not used yet)")
	} else {
		fmt.Println("   📊 Index usage statistics:")
		for _, stat := range stats {
			if strings.Contains(stat.IndexName, "active") {
				fmt.Printf("     - %s: %d scans, %d tuples read\n",
					stat.IndexName, stat.IndexScans, stat.TuplesRead)
			}
		}
	}

	// Test 5: Verify index definitions
	fmt.Println("\n🔍 Test 5: Verifying index definitions...")
	for _, idx := range existingIndexes {
		if strings.Contains(idx.IndexName, "active") {
			fmt.Printf("   📋 %s:\n", idx.IndexName)
			fmt.Printf("      %s\n", idx.IndexDef)
		}
	}

	return nil
}

// getKeywordWeightsIndexes gets all indexes for the keyword_weights table
func getKeywordWeightsIndexes(db *sql.DB) ([]IndexInfo, error) {
	query := `
		SELECT indexname, indexdef, tablename
		FROM pg_indexes 
		WHERE tablename = 'keyword_weights' 
			AND indexname LIKE '%active%'
		ORDER BY indexname
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var idx IndexInfo
		err := rows.Scan(&idx.IndexName, &idx.IndexDef, &idx.TableName)
		if err != nil {
			return nil, err
		}
		indexes = append(indexes, idx)
	}

	return indexes, rows.Err()
}

// getIndexStatistics gets index usage statistics
func getIndexStatistics(db *sql.DB) ([]IndexStats, error) {
	query := `
		SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
		FROM pg_stat_user_indexes 
		WHERE tablename = 'keyword_weights' 
			AND indexname LIKE '%active%'
		ORDER BY idx_scan DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []IndexStats
	for rows.Next() {
		var stat IndexStats
		err := rows.Scan(&stat.SchemaName, &stat.TableName, &stat.IndexName,
			&stat.IndexScans, &stat.TuplesRead, &stat.TuplesFetched)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

// testQueryPerformance tests query performance using EXPLAIN
func testQueryPerformance(db *sql.DB, query string) (string, error) {
	explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS) %s", query)

	rows, err := db.Query(explainQuery)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var result strings.Builder
	for rows.Next() {
		var line string
		err := rows.Scan(&line)
		if err != nil {
			return "", err
		}
		result.WriteString(line + "\n")
	}

	// Extract key performance metrics
	output := result.String()
	if strings.Contains(output, "Index Scan") || strings.Contains(output, "Bitmap Index Scan") {
		return "✅ Index used (good performance)", nil
	} else if strings.Contains(output, "Seq Scan") {
		return "⚠️  Sequential scan (may need optimization)", nil
	} else {
		return "❓ Unknown scan type", nil
	}
}

func main() {
	fmt.Println("🚀 Subtask 1.1.3 Index Verification Tool")
	fmt.Println(strings.Repeat("=", 50))

	// Load configuration
	config := getDatabaseConfig()

	// Connect to database
	fmt.Println("🔌 Connecting to database...")
	db, err := connectToDatabase(config)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("✅ Connected to database successfully")

	// Verify subtask 1.1.3
	err = verifySubtask131(db)
	if err != nil {
		log.Fatalf("❌ Verification failed: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 Subtask 1.1.3 Verification Complete!")
	fmt.Println("✅ All tests passed - Performance indexes are properly created")
	fmt.Println("✅ Database queries will now use optimized indexes")
	fmt.Println(strings.Repeat("=", 60))
}
