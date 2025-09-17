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

// RecordStatus represents the status of records in keyword_weights table
type RecordStatus struct {
	TotalRecords    int
	ActiveRecords   int
	InactiveRecords int
	NullRecords     int
}

// verifySubtask121 verifies that subtask 1.1.2 has been completed
func verifySubtask121(db *sql.DB) error {
	fmt.Println("🔍 Verifying Subtask 1.1.2: Update existing records")
	fmt.Println(strings.Repeat("=", 60))

	// Test 1: Check record status
	fmt.Println("\n📊 Test 1: Checking record status...")
	status, err := getRecordStatus(db)
	if err != nil {
		return fmt.Errorf("failed to get record status: %w", err)
	}

	fmt.Printf("   Total Records: %d\n", status.TotalRecords)
	fmt.Printf("   Active Records: %d\n", status.ActiveRecords)
	fmt.Printf("   Inactive Records: %d\n", status.InactiveRecords)
	fmt.Printf("   NULL Records: %d\n", status.NullRecords)

	// Verify success criteria
	if status.NullRecords > 0 {
		return fmt.Errorf("❌ FAIL: Found %d records with NULL is_active values", status.NullRecords)
	}

	if status.TotalRecords != status.ActiveRecords {
		return fmt.Errorf("❌ FAIL: Not all records are active (Total: %d, Active: %d)",
			status.TotalRecords, status.ActiveRecords)
	}

	fmt.Println("   ✅ PASS: All records have is_active = true")

	// Test 2: Check for NULL values specifically
	fmt.Println("\n🔍 Test 2: Checking for NULL values...")
	nullCount, err := getNullCount(db)
	if err != nil {
		return fmt.Errorf("failed to get NULL count: %w", err)
	}

	fmt.Printf("   NULL Records: %d\n", nullCount)

	if nullCount > 0 {
		return fmt.Errorf("❌ FAIL: Found %d records with NULL is_active values", nullCount)
	}

	fmt.Println("   ✅ PASS: No NULL values found")

	// Test 3: Show sample records
	fmt.Println("\n📋 Test 3: Showing sample records...")
	samples, err := getSampleRecords(db, 5)
	if err != nil {
		return fmt.Errorf("failed to get sample records: %w", err)
	}

	fmt.Println("   Sample Records:")
	for _, sample := range samples {
		fmt.Printf("     ID: %d, Industry: %d, Keyword: %s, is_active: %t\n",
			sample.ID, sample.IndustryID, sample.Keyword, sample.IsActive)
	}

	// Test 4: Check if UPDATE statement would affect any records
	fmt.Println("\n🧪 Test 4: Testing UPDATE statement impact...")
	updateCount, err := getUpdateImpact(db)
	if err != nil {
		return fmt.Errorf("failed to get UPDATE impact: %w", err)
	}

	fmt.Printf("   Records that would be updated: %d\n", updateCount)

	if updateCount > 0 {
		return fmt.Errorf("❌ FAIL: UPDATE statement would still affect %d records", updateCount)
	}

	fmt.Println("   ✅ PASS: No records need updating")

	return nil
}

// getRecordStatus gets the status of all records in keyword_weights table
func getRecordStatus(db *sql.DB) (RecordStatus, error) {
	query := `
		SELECT 
			COUNT(*) as total_records,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_records,
			COUNT(CASE WHEN is_active = false THEN 1 END) as inactive_records,
			COUNT(CASE WHEN is_active IS NULL THEN 1 END) as null_records
		FROM keyword_weights
	`

	var status RecordStatus
	err := db.QueryRow(query).Scan(
		&status.TotalRecords,
		&status.ActiveRecords,
		&status.InactiveRecords,
		&status.NullRecords,
	)

	return status, err
}

// getNullCount gets the count of records with NULL is_active values
func getNullCount(db *sql.DB) (int, error) {
	query := `SELECT COUNT(*) FROM keyword_weights WHERE is_active IS NULL`

	var count int
	err := db.QueryRow(query).Scan(&count)
	return count, err
}

// SampleRecord represents a sample record from keyword_weights table
type SampleRecord struct {
	ID         int
	IndustryID int
	Keyword    string
	IsActive   bool
	BaseWeight float64
	UpdatedAt  string
}

// getSampleRecords gets a sample of records from keyword_weights table
func getSampleRecords(db *sql.DB, limit int) ([]SampleRecord, error) {
	query := `
		SELECT id, industry_id, keyword, is_active, base_weight, updated_at
		FROM keyword_weights 
		ORDER BY id 
		LIMIT $1
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []SampleRecord
	for rows.Next() {
		var record SampleRecord
		err := rows.Scan(
			&record.ID,
			&record.IndustryID,
			&record.Keyword,
			&record.IsActive,
			&record.BaseWeight,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

// getUpdateImpact gets the count of records that would be affected by the UPDATE statement
func getUpdateImpact(db *sql.DB) (int, error) {
	query := `SELECT COUNT(*) FROM keyword_weights WHERE is_active IS NULL`

	var count int
	err := db.QueryRow(query).Scan(&count)
	return count, err
}

func main() {
	fmt.Println("🚀 Subtask 1.1.2 Verification Tool")
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

	// Verify subtask 1.1.2
	err = verifySubtask121(db)
	if err != nil {
		log.Fatalf("❌ Verification failed: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 Subtask 1.1.2 Verification Complete!")
	fmt.Println("✅ All tests passed - Subtask 1.1.2 is successfully completed")
	fmt.Println(strings.Repeat("=", 60))
}
