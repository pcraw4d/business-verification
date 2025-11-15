package main

import (
	"context"
	"fmt"
	"log"
	"os"
	
	_ "github.com/lib/pq"
	"kyb-platform/test/integration"
)

func main() {
	testDB, err := integration.SetupTestDatabase()
	if err != nil {
		log.Fatalf("Failed to setup test database: %v", err)
	}
	defer testDB.CleanupTestDatabase()
	
	db := testDB.GetDB()
	ctx := context.Background()
	
	requiredTables := []string{
		"merchant_analytics",
		"merchants",
		"risk_assessments",
		"risk_indicators",
		"enrichment_jobs",
		"enrichment_sources",
	}
	
	fmt.Println("Checking for required test tables...")
	fmt.Println("=====================================")
	
	allExist := true
	for _, table := range requiredTables {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)
		`
		err := db.QueryRowContext(ctx, query, table).Scan(&exists)
		if err != nil {
			log.Printf("Error checking table %s: %v", table, err)
			allExist = false
			continue
		}
		
		if exists {
			fmt.Printf("✅ Table '%s' exists\n", table)
		} else {
			fmt.Printf("❌ Table '%s' does NOT exist\n", table)
			allExist = false
		}
	}
	
	fmt.Println("=====================================")
	if allExist {
		fmt.Println("✅ All required tables exist!")
		os.Exit(0)
	} else {
		fmt.Println("❌ Some tables are missing. Please run migrations.")
		os.Exit(1)
	}
}
