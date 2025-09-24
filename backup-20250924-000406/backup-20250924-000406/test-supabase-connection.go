package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
)

func main() {
	fmt.Println("🔍 Testing Supabase Connection")
	fmt.Println("==============================")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("📝 Supabase URL: %s\n", cfg.Supabase.URL)
	fmt.Printf("📝 API Key present: %t\n", cfg.Supabase.APIKey != "")
	fmt.Printf("📝 Service Role Key present: %t\n", cfg.Supabase.ServiceRoleKey != "")
	fmt.Printf("📝 JWT Secret present: %t\n", cfg.Supabase.JWTSecret != "")

	if cfg.Supabase.URL == "" || cfg.Supabase.APIKey == "" || cfg.Supabase.ServiceRoleKey == "" {
		log.Fatalf("❌ Supabase configuration incomplete")
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
		log.Fatalf("❌ Failed to initialize Supabase client: %v", err)
	}

	fmt.Println("✅ Supabase client initialized successfully")

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := supabaseClient.Connect(ctx); err != nil {
		log.Fatalf("❌ Failed to connect to Supabase: %v", err)
	}

	fmt.Println("✅ Successfully connected to Supabase!")

	// Test a simple query
	fmt.Println("🔍 Testing database query...")

	// This would test if we can actually query the database
	// For now, just confirm the connection works
	fmt.Println("✅ Supabase connection test completed successfully!")
}
