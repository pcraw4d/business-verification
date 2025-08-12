package main

import (
	"context"
	"log"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger := observability.NewLogger(&cfg.Observability)

	logger.Info("Starting database migration process")

	// Initialize database
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Connect to database
	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	logger.Info("Database connection established")

	// Get migration system
	postgresDB, ok := db.(*database.PostgresDB)
	if !ok {
		log.Fatalf("Expected PostgresDB, got %T", db)
	}
	dbConfig := database.NewDatabaseConfig(&cfg.Database)
	migrationSystem := database.NewMigrationSystem(postgresDB.GetDB(), dbConfig)

	// Initialize migration table
	err = migrationSystem.InitializeMigrationTable(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize migration table: %v", err)
	}

	logger.Info("Migration table initialized")

	// Run migrations
	err = migrationSystem.RunMigrations(ctx)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Get migration status
	status, err := migrationSystem.GetMigrationStatus(ctx)
	if err != nil {
		log.Fatalf("Failed to get migration status: %v", err)
	}

	logger.Info("Migrations completed successfully", "status", status)
	println("✅ Migrations completed successfully!")
	println("📊 Migration Status:")
	for key, value := range status {
		println("   ", key, ":", value)
	}
}
