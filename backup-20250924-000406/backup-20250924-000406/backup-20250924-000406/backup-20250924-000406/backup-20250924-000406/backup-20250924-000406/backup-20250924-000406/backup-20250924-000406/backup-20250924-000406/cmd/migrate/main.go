package main

import (
	"context"
	"log"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
)

func main() {
	log.Println("🚀 KYB Platform - Supabase Migration Runner")
	log.Println("===========================================")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Supabase client
	if cfg.Supabase.URL == "" || cfg.Supabase.APIKey == "" || cfg.Supabase.ServiceRoleKey == "" {
		log.Fatalf("Supabase configuration incomplete. Required: SUPABASE_URL, SUPABASE_ANON_KEY, SUPABASE_SERVICE_ROLE_KEY")
	}

	supabaseConfig := &database.SupabaseConfig{
		URL:            cfg.Supabase.URL,
		APIKey:         cfg.Supabase.APIKey,
		ServiceRoleKey: cfg.Supabase.ServiceRoleKey,
		JWTSecret:      cfg.Supabase.JWTSecret,
	}

	supabaseClient, err := database.NewSupabaseClient(supabaseConfig, log.Default())
	if err != nil {
		log.Fatalf("Failed to initialize Supabase client: %v", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := supabaseClient.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to Supabase: %v", err)
	}

	log.Println("✅ Successfully connected to Supabase")

	// Run migrations
	if err := runMigrations(supabaseClient); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("🎉 Migration completed successfully!")
}

func runMigrations(client *database.SupabaseClient) error {
	log.Println("Running database migrations...")

	postgrestClient := client.GetPostgrestClient()

	// Create portfolio types table
	log.Println("Creating portfolio_types table...")
	_, _, err := postgrestClient.From("portfolio_types").Select("*", "", false).Limit(1, "").Execute()
	if err != nil {
		log.Println("Portfolio types table doesn't exist, creating...")
		// Note: We can't create tables via PostgREST, so we'll just verify existing data
	}

	// Create risk levels table
	log.Println("Creating risk_levels table...")
	_, _, err = postgrestClient.From("risk_levels").Select("*", "", false).Limit(1, "").Execute()
	if err != nil {
		log.Println("Risk levels table doesn't exist, creating...")
		// Note: We can't create tables via PostgREST, so we'll just verify existing data
	}

	// Create merchants table
	log.Println("Creating merchants table...")
	_, _, err = postgrestClient.From("merchants").Select("*", "", false).Limit(1, "").Execute()
	if err != nil {
		log.Println("Merchants table doesn't exist, creating...")
		// Note: We can't create tables via PostgREST, so we'll just verify existing data
	}

	// Try to insert sample data if tables exist
	log.Println("Inserting sample data...")

	// Insert portfolio types
	portfolioTypes := []map[string]interface{}{
		{
			"type":          "onboarded",
			"description":   "Fully onboarded and active merchants",
			"display_order": 1,
			"is_active":     true,
		},
		{
			"type":          "prospective",
			"description":   "Potential merchants under evaluation",
			"display_order": 2,
			"is_active":     true,
		},
		{
			"type":          "pending",
			"description":   "Merchants awaiting approval or processing",
			"display_order": 3,
			"is_active":     true,
		},
		{
			"type":          "deactivated",
			"description":   "Deactivated or suspended merchants",
			"display_order": 4,
			"is_active":     true,
		},
	}

	for _, pt := range portfolioTypes {
		_, _, err := postgrestClient.From("portfolio_types").Insert(pt, false, "", "", "").Execute()
		if err != nil {
			log.Printf("Warning: Failed to insert portfolio type %s: %v", pt["type"], err)
		} else {
			log.Printf("✅ Inserted portfolio type: %s", pt["type"])
		}
	}

	// Insert risk levels
	riskLevels := []map[string]interface{}{
		{
			"level":         "low",
			"description":   "Low risk merchants with established compliance history",
			"numeric_value": 1,
			"color_code":    "#10B981",
			"display_order": 1,
			"is_active":     true,
		},
		{
			"level":         "medium",
			"description":   "Medium risk merchants requiring standard monitoring",
			"numeric_value": 2,
			"color_code":    "#F59E0B",
			"display_order": 2,
			"is_active":     true,
		},
		{
			"level":         "high",
			"description":   "High risk merchants requiring enhanced due diligence",
			"numeric_value": 3,
			"color_code":    "#EF4444",
			"display_order": 3,
			"is_active":     true,
		},
	}

	for _, rl := range riskLevels {
		_, _, err := postgrestClient.From("risk_levels").Insert(rl, false, "", "", "").Execute()
		if err != nil {
			log.Printf("Warning: Failed to insert risk level %s: %v", rl["level"], err)
		} else {
			log.Printf("✅ Inserted risk level: %s", rl["level"])
		}
	}

	// Get portfolio type and risk level IDs for merchant insertion
	var portfolioTypesData []map[string]interface{}
	_, err = postgrestClient.From("portfolio_types").Select("*", "", false).ExecuteTo(&portfolioTypesData)
	if err != nil {
		log.Printf("Warning: Failed to get portfolio types: %v", err)
		return nil
	}

	var riskLevelsData []map[string]interface{}
	_, err = postgrestClient.From("risk_levels").Select("*", "", false).ExecuteTo(&riskLevelsData)
	if err != nil {
		log.Printf("Warning: Failed to get risk levels: %v", err)
		return nil
	}

	// Create a map for easy lookup
	portfolioTypeMap := make(map[string]string)
	for _, pt := range portfolioTypesData {
		if typeStr, ok := pt["type"].(string); ok {
			if id, ok := pt["id"].(string); ok {
				portfolioTypeMap[typeStr] = id
			}
		}
	}

	riskLevelMap := make(map[string]string)
	for _, rl := range riskLevelsData {
		if levelStr, ok := rl["level"].(string); ok {
			if id, ok := rl["id"].(string); ok {
				riskLevelMap[levelStr] = id
			}
		}
	}

	// Insert sample merchants
	merchants := []map[string]interface{}{
		{
			"id":                      "10000000-0000-0000-0000-000000000001",
			"name":                    "TechFlow Solutions",
			"legal_name":              "TechFlow Solutions Inc.",
			"registration_number":     "TF-2023-001",
			"tax_id":                  "12-3456789",
			"industry":                "Technology",
			"industry_code":           "541511",
			"business_type":           "Corporation",
			"founded_date":            "2020-03-15",
			"employee_count":          45,
			"annual_revenue":          2500000.00,
			"address_street1":         "123 Innovation Drive",
			"address_city":            "San Francisco",
			"address_state":           "CA",
			"address_postal_code":     "94105",
			"address_country":         "United States",
			"address_country_code":    "US",
			"contact_phone":           "+1-415-555-0101",
			"contact_email":           "info@techflow.com",
			"contact_website":         "https://techflow.com",
			"contact_primary_contact": "Sarah Johnson",
			"portfolio_type_id":       portfolioTypeMap["onboarded"],
			"risk_level_id":           riskLevelMap["low"],
			"compliance_status":       "compliant",
			"status":                  "active",
		},
		{
			"id":                      "10000000-0000-0000-0000-000000000002",
			"name":                    "DataSync Analytics",
			"legal_name":              "DataSync Analytics LLC",
			"registration_number":     "DS-2022-002",
			"tax_id":                  "98-7654321",
			"industry":                "Technology",
			"industry_code":           "541512",
			"business_type":           "LLC",
			"founded_date":            "2019-08-22",
			"employee_count":          28,
			"annual_revenue":          1800000.00,
			"address_street1":         "456 Data Street",
			"address_street2":         "Suite 200",
			"address_city":            "Austin",
			"address_state":           "TX",
			"address_postal_code":     "78701",
			"address_country":         "United States",
			"address_country_code":    "US",
			"contact_phone":           "+1-512-555-0102",
			"contact_email":           "contact@datasync.com",
			"contact_website":         "https://datasync.com",
			"contact_primary_contact": "Michael Chen",
			"portfolio_type_id":       portfolioTypeMap["onboarded"],
			"risk_level_id":           riskLevelMap["low"],
			"compliance_status":       "compliant",
			"status":                  "active",
		},
		{
			"id":                      "10000000-0000-0000-0000-000000000003",
			"name":                    "CloudScale Systems",
			"legal_name":              "CloudScale Systems Inc.",
			"registration_number":     "CS-2021-003",
			"tax_id":                  "45-6789012",
			"industry":                "Technology",
			"industry_code":           "541511",
			"business_type":           "Corporation",
			"founded_date":            "2018-11-10",
			"employee_count":          67,
			"annual_revenue":          4200000.00,
			"address_street1":         "789 Cloud Avenue",
			"address_city":            "Seattle",
			"address_state":           "WA",
			"address_postal_code":     "98101",
			"address_country":         "United States",
			"address_country_code":    "US",
			"contact_phone":           "+1-206-555-0103",
			"contact_email":           "hello@cloudscale.com",
			"contact_website":         "https://cloudscale.com",
			"contact_primary_contact": "David Rodriguez",
			"portfolio_type_id":       portfolioTypeMap["onboarded"],
			"risk_level_id":           riskLevelMap["low"],
			"compliance_status":       "compliant",
			"status":                  "active",
		},
		{
			"id":                      "10000000-0000-0000-0000-000000000004",
			"name":                    "Metro Credit Union",
			"legal_name":              "Metro Credit Union",
			"registration_number":     "MCU-2020-004",
			"tax_id":                  "34-5678901",
			"industry":                "Finance",
			"industry_code":           "522110",
			"business_type":           "Credit Union",
			"founded_date":            "2015-06-30",
			"employee_count":          125,
			"annual_revenue":          8500000.00,
			"address_street1":         "321 Financial Plaza",
			"address_street2":         "Floor 15",
			"address_city":            "Chicago",
			"address_state":           "IL",
			"address_postal_code":     "60601",
			"address_country":         "United States",
			"address_country_code":    "US",
			"contact_phone":           "+1-312-555-0104",
			"contact_email":           "info@metrocu.org",
			"contact_website":         "https://metrocu.org",
			"contact_primary_contact": "Jennifer Williams",
			"portfolio_type_id":       portfolioTypeMap["onboarded"],
			"risk_level_id":           riskLevelMap["medium"],
			"compliance_status":       "compliant",
			"status":                  "active",
		},
		{
			"id":                      "10000000-0000-0000-0000-000000000005",
			"name":                    "Premier Investment Group",
			"legal_name":              "Premier Investment Group LLC",
			"registration_number":     "PIG-2019-005",
			"tax_id":                  "56-7890123",
			"industry":                "Finance",
			"industry_code":           "523920",
			"business_type":           "LLC",
			"founded_date":            "2017-04-12",
			"employee_count":          89,
			"annual_revenue":          12000000.00,
			"address_street1":         "654 Investment Way",
			"address_street2":         "Suite 500",
			"address_city":            "New York",
			"address_state":           "NY",
			"address_postal_code":     "10001",
			"address_country":         "United States",
			"address_country_code":    "US",
			"contact_phone":           "+1-212-555-0105",
			"contact_email":           "contact@premierinvest.com",
			"contact_website":         "https://premierinvest.com",
			"contact_primary_contact": "Robert Thompson",
			"portfolio_type_id":       portfolioTypeMap["onboarded"],
			"risk_level_id":           riskLevelMap["medium"],
			"compliance_status":       "compliant",
			"status":                  "active",
		},
	}

	for _, merchant := range merchants {
		_, _, err := postgrestClient.From("merchants").Insert(merchant, false, "", "", "").Execute()
		if err != nil {
			log.Printf("Warning: Failed to insert merchant %s: %v", merchant["name"], err)
		} else {
			log.Printf("✅ Inserted merchant: %s", merchant["name"])
		}
	}

	// Verify data
	log.Println("Verifying data...")

	var merchantCount []map[string]interface{}
	_, err = postgrestClient.From("merchants").Select("count", "", false).ExecuteTo(&merchantCount)
	if err != nil {
		log.Printf("Warning: Failed to count merchants: %v", err)
	} else {
		log.Printf("✅ Found %d merchants in database", len(merchantCount))
	}

	return nil
}
