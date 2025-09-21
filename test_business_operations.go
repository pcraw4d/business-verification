package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pcraw4d/business-verification/internal/database"
)

func main() {
	fmt.Println("Testing Business Operations with Consolidated Merchants Table")
	fmt.Println("============================================================")

	// Create a test database configuration
	config := &database.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		Username:        "postgres",
		Password:        "password",
		Database:        "kyb_platform",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	// Create database instance
	db := database.NewPostgresDB(config)

	// Connect to database
	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("✓ Database connection established")

	// Test 1: Create a business (should be stored in merchants table)
	fmt.Println("\n1. Testing CreateBusiness...")
	business := &database.Business{
		ID:                 "test-business-001",
		Name:               "Test Business Inc",
		LegalName:          "Test Business Incorporated",
		RegistrationNumber: "REG-001",
		TaxID:              "TAX-001",
		Industry:           "Technology",
		IndustryCode:       "541511",
		BusinessType:       "Corporation",
		FoundedDate:        &[]time.Time{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)}[0],
		EmployeeCount:      50,
		AnnualRevenue:      &[]float64{1000000.0}[0],
		Address: database.Address{
			Street1:     "123 Test Street",
			Street2:     "Suite 100",
			City:        "Test City",
			State:       "TS",
			PostalCode:  "12345",
			Country:     "United States",
			CountryCode: "US",
		},
		ContactInfo: database.ContactInfo{
			Phone:          "+1-555-123-4567",
			Email:          "contact@testbusiness.com",
			Website:        "https://testbusiness.com",
			PrimaryContact: "John Doe",
		},
		Status:           "active",
		RiskLevel:        "medium",
		ComplianceStatus: "pending",
		CreatedBy:        "test-user-001",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := db.CreateBusiness(ctx, business); err != nil {
		log.Printf("✗ CreateBusiness failed: %v", err)
	} else {
		fmt.Println("✓ CreateBusiness successful")
	}

	// Test 2: Get business by ID
	fmt.Println("\n2. Testing GetBusinessByID...")
	retrievedBusiness, err := db.GetBusinessByID(ctx, business.ID)
	if err != nil {
		log.Printf("✗ GetBusinessByID failed: %v", err)
	} else {
		fmt.Printf("✓ GetBusinessByID successful - Retrieved: %s\n", retrievedBusiness.Name)
	}

	// Test 3: Get business by registration number
	fmt.Println("\n3. Testing GetBusinessByRegistrationNumber...")
	retrievedByReg, err := db.GetBusinessByRegistrationNumber(ctx, business.RegistrationNumber)
	if err != nil {
		log.Printf("✗ GetBusinessByRegistrationNumber failed: %v", err)
	} else {
		fmt.Printf("✓ GetBusinessByRegistrationNumber successful - Retrieved: %s\n", retrievedByReg.Name)
	}

	// Test 4: Update business
	fmt.Println("\n4. Testing UpdateBusiness...")
	business.Name = "Updated Test Business Inc"
	business.UpdatedAt = time.Now()
	if err := db.UpdateBusiness(ctx, business); err != nil {
		log.Printf("✗ UpdateBusiness failed: %v", err)
	} else {
		fmt.Println("✓ UpdateBusiness successful")
	}

	// Test 5: List businesses
	fmt.Println("\n5. Testing ListBusinesses...")
	businesses, err := db.ListBusinesses(ctx, 10, 0)
	if err != nil {
		log.Printf("✗ ListBusinesses failed: %v", err)
	} else {
		fmt.Printf("✓ ListBusinesses successful - Retrieved %d businesses\n", len(businesses))
	}

	// Test 6: Search businesses
	fmt.Println("\n6. Testing SearchBusinesses...")
	searchResults, err := db.SearchBusinesses(ctx, "Test", 10, 0)
	if err != nil {
		log.Printf("✗ SearchBusinesses failed: %v", err)
	} else {
		fmt.Printf("✓ SearchBusinesses successful - Found %d businesses matching 'Test'\n", len(searchResults))
	}

	// Test 7: Delete business
	fmt.Println("\n7. Testing DeleteBusiness...")
	if err := db.DeleteBusiness(ctx, business.ID); err != nil {
		log.Printf("✗ DeleteBusiness failed: %v", err)
	} else {
		fmt.Println("✓ DeleteBusiness successful")
	}

	// Test 8: Verify deletion
	fmt.Println("\n8. Verifying deletion...")
	_, err = db.GetBusinessByID(ctx, business.ID)
	if err != nil {
		fmt.Println("✓ Business successfully deleted (not found)")
	} else {
		fmt.Println("✗ Business still exists after deletion")
	}

	fmt.Println("\n============================================================")
	fmt.Println("Business Operations Test Completed")
	fmt.Println("All operations should now work with the consolidated merchants table")
}
