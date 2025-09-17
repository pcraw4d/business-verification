package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pcraw4d/business-verification/internal/architecture"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/modules/database_classification"
)

func main() {
	fmt.Println("🔍 Testing Database Module")
	fmt.Println("==========================")

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

	// Create database module
	config := &database_classification.Config{
		ModuleID:          "test-classification",
		ModuleName:        "Test Classification Module",
		ModuleVersion:     "1.0.0",
		ModuleDescription: "Test database-driven classification",
		RequestTimeout:    30 * time.Second,
		MaxConcurrency:    10,
		EnableCaching:     true,
		CacheTTL:          5 * time.Minute,
	}

	databaseModule, err := database_classification.NewDatabaseClassificationModule(supabaseClient, log.Default(), config)
	if err != nil {
		log.Fatalf("Failed to create database module: %v", err)
	}
	fmt.Println("✅ Created database module")

	// Start the database module
	if err := databaseModule.Start(ctx); err != nil {
		log.Fatalf("Failed to start database module: %v", err)
	}
	fmt.Println("✅ Started database module")

	// Test the module
	moduleReq := architecture.ModuleRequest{
		ID: fmt.Sprintf("test_%d", time.Now().Unix()),
		Data: map[string]interface{}{
			"business_name": "Green Grape Company",
			"description":   "Sustainable wine production and distribution",
			"website_url":   "https://greenegrape.com/",
		},
	}

	fmt.Println("🔍 Testing module processing...")
	moduleResp, err := databaseModule.Process(ctx, moduleReq)
	if err != nil {
		log.Fatalf("❌ Module processing failed: %v", err)
	}

	if moduleResp.Success {
		fmt.Println("✅ Module processing successful!")
		fmt.Printf("📊 Response data keys: %v\n", getMapKeys(moduleResp.Data))

		if metadata, ok := moduleResp.Data["metadata"].(map[string]interface{}); ok {
			fmt.Printf("📝 Metadata: %v\n", metadata)
		} else {
			fmt.Println("⚠️ No metadata found in response")
		}
	} else {
		fmt.Printf("❌ Module processing failed: %s\n", moduleResp.Error)
	}
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
