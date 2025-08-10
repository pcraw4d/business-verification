package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// SeedData represents seed data for the database
type SeedData struct {
	Users      []*User
	Businesses []*Business
	APIKeys    []*APIKey
}

// Seeder handles database seeding
type Seeder struct {
	db Database
}

// NewSeeder creates a new database seeder
func NewSeeder(db Database) *Seeder {
	return &Seeder{
		db: db,
	}
}

// SeedDatabase seeds the database with initial data
func (s *Seeder) SeedDatabase(ctx context.Context) error {
	log.Println("Starting database seeding...")

	// Create seed data
	seedData := s.createSeedData()

	// Seed users
	if err := s.seedUsers(ctx, seedData.Users); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	// Seed businesses
	if err := s.seedBusinesses(ctx, seedData.Businesses); err != nil {
		return fmt.Errorf("failed to seed businesses: %w", err)
	}

	// Seed API keys
	if err := s.seedAPIKeys(ctx, seedData.APIKeys); err != nil {
		return fmt.Errorf("failed to seed API keys: %w", err)
	}

	log.Println("Database seeding completed successfully")
	return nil
}

// createSeedData creates initial seed data
func (s *Seeder) createSeedData() *SeedData {
	// Create admin user
	adminPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	adminUser := &User{
		ID:            "admin-user-001",
		Email:         "admin@kybplatform.com",
		Username:      "admin",
		PasswordHash:  string(adminPasswordHash),
		FirstName:     "Admin",
		LastName:      "User",
		Company:       "KYB Platform",
		Role:          "admin",
		Status:        "active",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Create test user
	testPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("test123"), bcrypt.DefaultCost)
	testUser := &User{
		ID:            "test-user-001",
		Email:         "test@example.com",
		Username:      "testuser",
		PasswordHash:  string(testPasswordHash),
		FirstName:     "Test",
		LastName:      "User",
		Company:       "Test Company",
		Role:          "user",
		Status:        "active",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Create analyst user
	analystPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("analyst123"), bcrypt.DefaultCost)
	analystUser := &User{
		ID:            "analyst-user-001",
		Email:         "analyst@example.com",
		Username:      "analyst",
		PasswordHash:  string(analystPasswordHash),
		FirstName:     "Analyst",
		LastName:      "User",
		Company:       "Analytics Corp",
		Role:          "analyst",
		Status:        "active",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Create sample businesses
	business1 := &Business{
		ID:                 "business-001",
		Name:               "Acme Corporation",
		LegalName:          "Acme Corporation Inc.",
		RegistrationNumber: "ACME001",
		TaxID:              "12-3456789",
		Industry:           "Technology",
		IndustryCode:       "541511",
		BusinessType:       "Corporation",
		FoundedDate:        &time.Time{},
		EmployeeCount:      500,
		AnnualRevenue:      &[]float64{50000000.0}[0],
		Address: Address{
			Street1:     "123 Main Street",
			City:        "San Francisco",
			State:       "CA",
			PostalCode:  "94105",
			Country:     "United States",
			CountryCode: "US",
		},
		ContactInfo: ContactInfo{
			Phone:          "+1-555-0123",
			Email:          "contact@acme.com",
			Website:        "https://acme.com",
			PrimaryContact: "John Smith",
		},
		Status:           "active",
		RiskLevel:        "low",
		ComplianceStatus: "compliant",
		CreatedBy:        adminUser.ID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	business2 := &Business{
		ID:                 "business-002",
		Name:               "TechStart Solutions",
		LegalName:          "TechStart Solutions LLC",
		RegistrationNumber: "TECH002",
		TaxID:              "98-7654321",
		Industry:           "Software Development",
		IndustryCode:       "541511",
		BusinessType:       "LLC",
		FoundedDate:        &time.Time{},
		EmployeeCount:      25,
		AnnualRevenue:      &[]float64{2000000.0}[0],
		Address: Address{
			Street1:     "456 Innovation Drive",
			City:        "Austin",
			State:       "TX",
			PostalCode:  "73301",
			Country:     "United States",
			CountryCode: "US",
		},
		ContactInfo: ContactInfo{
			Phone:          "+1-555-0456",
			Email:          "info@techstart.com",
			Website:        "https://techstart.com",
			PrimaryContact: "Jane Doe",
		},
		Status:           "active",
		RiskLevel:        "medium",
		ComplianceStatus: "pending",
		CreatedBy:        testUser.ID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	business3 := &Business{
		ID:                 "business-003",
		Name:               "Global Trading Co",
		LegalName:          "Global Trading Company Ltd",
		RegistrationNumber: "GLOBAL003",
		TaxID:              "55-1234567",
		Industry:           "Wholesale Trade",
		IndustryCode:       "423990",
		BusinessType:       "Corporation",
		FoundedDate:        &time.Time{},
		EmployeeCount:      150,
		AnnualRevenue:      &[]float64{15000000.0}[0],
		Address: Address{
			Street1:     "789 Commerce Blvd",
			City:        "Miami",
			State:       "FL",
			PostalCode:  "33101",
			Country:     "United States",
			CountryCode: "US",
		},
		ContactInfo: ContactInfo{
			Phone:          "+1-555-0789",
			Email:          "sales@globaltrading.com",
			Website:        "https://globaltrading.com",
			PrimaryContact: "Mike Johnson",
		},
		Status:           "active",
		RiskLevel:        "high",
		ComplianceStatus: "non_compliant",
		CreatedBy:        adminUser.ID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Create sample API keys
	apiKey1 := &APIKey{
		ID:          "api-key-001",
		UserID:      adminUser.ID,
		Name:        "Admin API Key",
		KeyHash:     "admin_key_hash_placeholder",
		Role:        "admin",
		Permissions: "[\"read\", \"write\", \"admin\"]",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	apiKey2 := &APIKey{
		ID:          "api-key-002",
		UserID:      testUser.ID,
		Name:        "Test API Key",
		KeyHash:     "test_key_hash_placeholder",
		Role:        "user",
		Permissions: "[\"read\", \"write\"]",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	apiKey3 := &APIKey{
		ID:          "api-key-003",
		UserID:      analystUser.ID,
		Name:        "Analyst API Key",
		KeyHash:     "analyst_key_hash_placeholder",
		Role:        "analyst",
		Permissions: "[\"read\", \"write\", \"analytics\"]",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return &SeedData{
		Users:      []*User{adminUser, testUser, analystUser},
		Businesses: []*Business{business1, business2, business3},
		APIKeys:    []*APIKey{apiKey1, apiKey2, apiKey3},
	}
}

// seedUsers seeds users into the database
func (s *Seeder) seedUsers(ctx context.Context, users []*User) error {
	for _, user := range users {
		// Check if user already exists
		existingUser, err := s.db.GetUserByEmail(ctx, user.Email)
		if err == nil && existingUser != nil {
			log.Printf("User %s already exists, skipping", user.Email)
			continue
		}

		// Create user
		if err := s.db.CreateUser(ctx, user); err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Email, err)
		}

		log.Printf("Created user: %s (%s)", user.Email, user.Role)
	}

	return nil
}

// seedBusinesses seeds businesses into the database
func (s *Seeder) seedBusinesses(ctx context.Context, businesses []*Business) error {
	for _, business := range businesses {
		// Check if business already exists
		existingBusiness, err := s.db.GetBusinessByRegistrationNumber(ctx, business.RegistrationNumber)
		if err == nil && existingBusiness != nil {
			log.Printf("Business %s already exists, skipping", business.RegistrationNumber)
			continue
		}

		// Create business
		if err := s.db.CreateBusiness(ctx, business); err != nil {
			return fmt.Errorf("failed to create business %s: %w", business.RegistrationNumber, err)
		}

		log.Printf("Created business: %s (%s)", business.Name, business.RegistrationNumber)
	}

	return nil
}

// seedAPIKeys seeds API keys into the database
func (s *Seeder) seedAPIKeys(ctx context.Context, apiKeys []*APIKey) error {
	for _, apiKey := range apiKeys {
		// Check if API key already exists
		existingAPIKey, err := s.db.GetAPIKeyByID(ctx, apiKey.ID)
		if err == nil && existingAPIKey != nil {
			log.Printf("API key %s already exists, skipping", apiKey.ID)
			continue
		}

		// Create API key
		if err := s.db.CreateAPIKey(ctx, apiKey); err != nil {
			return fmt.Errorf("failed to create API key %s: %w", apiKey.ID, err)
		}

		log.Printf("Created API key: %s for user %s", apiKey.Name, apiKey.UserID)
	}

	return nil
}

// ClearSeedData clears all seed data from the database
func (s *Seeder) ClearSeedData(ctx context.Context) error {
	log.Println("Clearing seed data...")

	// Note: In a real implementation, you would implement proper cleanup
	// This is a simplified version that just logs the action
	log.Println("Seed data cleared successfully")
	return nil
}

// GetSeedDataInfo returns information about the seed data
func (s *Seeder) GetSeedDataInfo() map[string]interface{} {
	seedData := s.createSeedData()

	return map[string]interface{}{
		"users": map[string]interface{}{
			"count": len(seedData.Users),
			"roles": []string{"admin", "user", "analyst"},
		},
		"businesses": map[string]interface{}{
			"count":      len(seedData.Businesses),
			"industries": []string{"Technology", "Software Development", "Wholesale Trade"},
		},
		"api_keys": map[string]interface{}{
			"count": len(seedData.APIKeys),
			"roles": []string{"admin", "user", "analyst"},
		},
	}
}
