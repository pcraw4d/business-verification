package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"

	"github.com/company/kyb-platform/internal/config"
	"github.com/company/kyb-platform/internal/testing"
	_ "github.com/lib/pq"
)

func main() {
	// Command line flags
	var (
		configFile   = flag.String("config", "configs/dev/config.yaml", "Configuration file path")
		verbose      = flag.Bool("verbose", false, "Enable verbose logging")
		generateData = flag.Bool("generate-data", false, "Generate test data")
		randomCount  = flag.Int("random-count", 100, "Number of random test samples to generate")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set log level
	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Initialize database connection
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test database connection
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize test data generator
	logger := log.New(os.Stdout, "[TEST_DATA_GENERATOR] ", log.LstdFlags|log.Lshortfile)
	generator := testing.NewTestDataGenerator(db, logger)

	if *generateData {
		log.Println("Generating comprehensive test data...")

		// Generate comprehensive test data
		if err := generator.GenerateComprehensiveTestData(ctx); err != nil {
			log.Fatalf("Failed to generate comprehensive test data: %v", err)
		}

		// Generate random test samples for stress testing
		if *randomCount > 0 {
			log.Printf("Generating %d random test samples...", *randomCount)
			if err := generator.GenerateRandomTestSamples(ctx, *randomCount); err != nil {
				log.Fatalf("Failed to generate random test samples: %v", err)
			}
		}

		log.Println("Test data generation completed successfully!")
	} else {
		log.Println("Use -generate-data flag to generate test data")
	}
}
