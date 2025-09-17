package main

import (
	"context"
	"fmt"
	"log"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
)

func main() {
	fmt.Println("🔍 Testing keyword_weights table schema")
	fmt.Println("=======================================")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Supabase client
	supabaseConfig := &database.SupabaseConfig{
		URL:            cfg.Supabase.URL,
		APIKey:         cfg.Supabase.APIKey,
		ServiceRoleKey: cfg.Supabase.ServiceRoleKey,
		JWTSecret:      cfg.Supabase.JWTSecret,
	}

	supabaseClient, err := database.NewSupabaseClient(supabaseConfig, log.Default())
	if err != nil {
		log.Fatalf("Failed to create Supabase client: %v", err)
	}

	// Connect to Supabase
	ctx := context.Background()
	if err := supabaseClient.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to Supabase: %v", err)
	}
	fmt.Println("✅ Connected to Supabase")

	// Test the keyword_weights table structure
	if err := testKeywordWeightsTable(ctx, supabaseClient); err != nil {
		log.Fatalf("Failed to test keyword_weights table: %v", err)
	}

	fmt.Println("✅ keyword_weights table schema test completed!")
}

func testKeywordWeightsTable(ctx context.Context, client *database.SupabaseClient) error {
	fmt.Println("🔍 Testing keyword_weights table structure...")

	// Get the PostgREST client
	postgrestClient := client.GetPostgrestClient()
	if postgrestClient == nil {
		return fmt.Errorf("PostgREST client not available")
	}

	// Test 1: Try to query the table with basic columns
	fmt.Println("📝 Test 1: Querying basic columns...")
	_, _, err := postgrestClient.From("keyword_weights").Select("id, industry_id, keyword", "", false).Limit(1, "").Execute()
	if err != nil {
		fmt.Printf("❌ Failed to query basic columns: %v\n", err)
		return fmt.Errorf("keyword_weights table may not exist or be accessible: %w", err)
	}
	fmt.Println("✅ Basic columns are accessible")

	// Test 2: Try to query with is_active column
	fmt.Println("📝 Test 2: Testing is_active column...")
	_, _, err = postgrestClient.From("keyword_weights").Select("id, is_active", "", false).Limit(1, "").Execute()
	if err != nil {
		fmt.Printf("❌ is_active column does not exist: %v\n", err)
		fmt.Println("🔧 Task 1.1.1 needs to be executed to add the is_active column")
		return fmt.Errorf("is_active column missing: %w", err)
	}
	fmt.Println("✅ is_active column exists and is accessible")

	// Test 3: Try to query all columns
	fmt.Println("📝 Test 3: Testing all columns...")
	_, _, err = postgrestClient.From("keyword_weights").Select("*", "", false).Limit(1, "").Execute()
	if err != nil {
		fmt.Printf("❌ Failed to query all columns: %v\n", err)
		return fmt.Errorf("failed to query all columns: %w", err)
	}
	fmt.Println("✅ All columns are accessible")

	// Test 4: Count records
	fmt.Println("📝 Test 4: Counting records...")
	_, _, err = postgrestClient.From("keyword_weights").Select("id", "", false).Execute()
	if err != nil {
		fmt.Printf("❌ Failed to count records: %v\n", err)
		return fmt.Errorf("failed to count records: %w", err)
	}
	fmt.Println("✅ Records are accessible")

	// Test 5: Try to filter by is_active
	fmt.Println("📝 Test 5: Testing is_active filter...")
	_, _, err = postgrestClient.From("keyword_weights").Select("id, is_active", "", false).Eq("is_active", "true").Limit(1, "").Execute()
	if err != nil {
		fmt.Printf("❌ Failed to filter by is_active: %v\n", err)
		return fmt.Errorf("failed to filter by is_active: %w", err)
	}
	fmt.Println("✅ is_active filter works correctly")

	fmt.Println("🎉 All tests passed! The keyword_weights table has the is_active column.")
	return nil
}
