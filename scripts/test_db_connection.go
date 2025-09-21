package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Get database connection string from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	fmt.Printf("🔗 Connecting to database...\n")
	fmt.Printf("Database URL: %s\n", maskURL(dbURL))

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	fmt.Printf("🏓 Testing database connection...\n")
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Printf("✅ Database connection successful!\n")

	// Test a simple query
	fmt.Printf("🔍 Testing simple query...\n")
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM industries").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to query industries table: %v", err)
	}
	fmt.Printf("✅ Found %d industries in database\n", count)

	// Test classification tables
	tables := []string{"industries", "industry_keywords", "classification_codes", "industry_patterns"}
	for _, table := range tables {
		var tableCount int
		err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&tableCount)
		if err != nil {
			fmt.Printf("❌ Table %s: %v\n", table, err)
		} else {
			fmt.Printf("✅ Table %s: %d records\n", table, tableCount)
		}
	}
}

func maskURL(url string) string {
	if len(url) > 50 {
		return url[:20] + "..." + url[len(url)-20:]
	}
	return "***"
}
