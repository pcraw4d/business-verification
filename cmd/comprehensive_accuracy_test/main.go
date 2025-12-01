package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"kyb-platform/internal/classification"
	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/database"
	"kyb-platform/internal/machine_learning"
	"kyb-platform/internal/machine_learning/infrastructure"
	testingpkg "kyb-platform/internal/testing"

	_ "github.com/lib/pq"
)

func main() {
	var (
		supabaseURL      = flag.String("supabase-url", "", "Supabase URL")
		supabaseKey      = flag.String("supabase-key", "", "Supabase anon key")
		supabaseServiceKey = flag.String("supabase-service-key", "", "Supabase service role key")
		databaseURL      = flag.String("database-url", "", "Direct PostgreSQL database URL (for test dataset)")
		category         = flag.String("category", "", "Run tests for specific category only (optional)")
		outputFile       = flag.String("output", "", "Output file for JSON report (optional)")
		verbose          = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Setup logging
	logger := log.New(os.Stdout, "[ACCURACY-TEST] ", log.LstdFlags)
	if *verbose {
		logger.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Get environment variables if not provided via flags
	if *supabaseURL == "" {
		*supabaseURL = os.Getenv("SUPABASE_URL")
	}
	if *supabaseKey == "" {
		*supabaseKey = os.Getenv("SUPABASE_ANON_KEY")
	}
	if *supabaseServiceKey == "" {
		*supabaseServiceKey = os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	}
	if *databaseURL == "" {
		*databaseURL = os.Getenv("TEST_DATABASE_URL")
		if *databaseURL == "" {
			*databaseURL = os.Getenv("DATABASE_URL")
		}
	}

	// Validate required parameters
	if *supabaseURL == "" || *supabaseKey == "" {
		logger.Fatal("Missing required parameters: SUPABASE_URL and SUPABASE_ANON_KEY must be set")
	}
	if *databaseURL == "" {
		logger.Fatal("Missing required parameter: TEST_DATABASE_URL or DATABASE_URL must be set for test dataset access")
	}

	logger.Println("🚀 Starting Comprehensive Accuracy Test Suite")
	logger.Printf("   Supabase URL: %s", *supabaseURL)
	logger.Printf("   Database URL: %s", maskURL(*databaseURL))
	if *category != "" {
		logger.Printf("   Category Filter: %s", *category)
	}

	// Create database client for Supabase
	config := &database.SupabaseConfig{
		URL:            *supabaseURL,
		APIKey:         *supabaseKey,
		ServiceRoleKey: *supabaseServiceKey,
	}

	client, err := database.NewSupabaseClient(config, logger)
	if err != nil {
		logger.Fatalf("Failed to create Supabase client: %v", err)
	}
	defer client.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		logger.Fatalf("Failed to ping Supabase: %v", err)
	}

	// Get database connection for test dataset
	db, err := sql.Open("postgres", *databaseURL)
	if err != nil {
		logger.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Test database connection
	if err := db.PingContext(ctx); err != nil {
		logger.Fatalf("Failed to ping database: %v", err)
	}

	logger.Println("✅ Database connections established")

	// Create repository
	repo := repository.NewSupabaseKeywordRepository(client, logger)

	// Initialize Python ML service if available
	var pythonMLService interface{}
	pythonMLServiceURL := os.Getenv("PYTHON_ML_SERVICE_URL")
	if pythonMLServiceURL != "" {
		logger.Printf("🐍 Initializing Python ML Service: %s", pythonMLServiceURL)
		// Import infrastructure package for Python ML service
		// Note: We'll use interface{} to avoid import cycle issues
		pythonMLService = initPythonMLService(pythonMLServiceURL, logger)
		if pythonMLService != nil {
			logger.Println("✅ Python ML Service initialized successfully")
		} else {
			logger.Println("⚠️  Python ML Service initialization failed, continuing without ML")
		}
	} else {
		logger.Println("ℹ️  Python ML Service URL not configured (PYTHON_ML_SERVICE_URL), ML classification will not be available")
	}

	// Create ML classifier (even if Python service is not available, for fallback)
	mlConfig := machine_learning.ContentClassifierConfig{
		ModelType:             "distilbart",
		MaxSequenceLength:     512,
		BatchSize:             32,
		ConfidenceThreshold:   0.7,
		ExplainabilityEnabled: true,
	}
	mlClassifier := machine_learning.NewContentClassifier(mlConfig)
	if mlClassifier == nil {
		logger.Println("⚠️  Failed to create ML classifier, continuing without ML")
		mlClassifier = nil
	}

	// Create services with optional ML support
	var industryService *classification.IndustryDetectionService
	if pythonMLService != nil || mlClassifier != nil {
		// Use ML-enabled service (will use Python ML service if available, Go ML classifier as fallback)
		logger.Println("🤖 Creating IndustryDetectionService with ML support")
		industryService = createIndustryServiceWithML(repo, mlClassifier, pythonMLService, logger)
	} else {
		// Use standard IndustryDetectionService (keyword-based only)
		logger.Println("📝 Creating IndustryDetectionService without ML (keyword-based only)")
		industryService = classification.NewIndustryDetectionService(repo, logger)
	}
	codeGenerator := classification.NewClassificationCodeGenerator(repo, logger)

	// Create test dataset manager
	datasetManager := testingpkg.NewAccuracyTestDataset(db, logger)

	// Create comprehensive accuracy tester
	accuracyTester := testingpkg.NewComprehensiveAccuracyTester(
		datasetManager,
		industryService,
		codeGenerator,
		logger,
	)

	// Create accuracy report generator
	reportGenerator := testingpkg.NewAccuracyReportGenerator(logger)

	// Run accuracy tests
	testCtx, testCancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer testCancel()

	var metrics *testingpkg.ComprehensiveAccuracyMetrics
	if *category != "" {
		logger.Printf("Running accuracy tests for category: %s", *category)
		metrics, err = accuracyTester.RunAccuracyTestsByCategory(testCtx, *category)
	} else {
		logger.Println("Running comprehensive accuracy tests on all test cases...")
		metrics, err = accuracyTester.RunAccuracyTests(testCtx)
	}

	if err != nil {
		logger.Fatalf("Failed to run accuracy tests: %v", err)
	}

	// Display results
	displayResults(logger, metrics)

	// Generate and save reports
	if *outputFile != "" {
		jsonReport, err := reportGenerator.GenerateJSONReport(metrics)
		if err != nil {
			logger.Printf("⚠️  Failed to generate JSON report: %v", err)
		} else {
			if err := os.WriteFile(*outputFile, jsonReport, 0644); err != nil {
				logger.Printf("⚠️  Failed to write JSON report: %v", err)
			} else {
				logger.Printf("✅ JSON report saved to: %s", *outputFile)
			}
		}
	}

	// Generate text report
	textReport, err := reportGenerator.GenerateTextReport(metrics)
	if err != nil {
		logger.Printf("⚠️  Failed to generate text report: %v", err)
	} else {
		fmt.Println("\n" + textReport)
	}

	// Exit with appropriate code
	if metrics.OverallAccuracy >= 0.85 && metrics.IndustryAccuracy >= 0.95 && metrics.CodeAccuracy >= 0.90 {
		logger.Println("✅ All accuracy targets met!")
		os.Exit(0)
	} else {
		logger.Println("⚠️  Some accuracy targets not met")
		if metrics.IndustryAccuracy < 0.95 {
			logger.Printf("   Industry accuracy: %.2f%% (target: 95%%)", metrics.IndustryAccuracy*100)
		}
		if metrics.CodeAccuracy < 0.90 {
			logger.Printf("   Code accuracy: %.2f%% (target: 90%%)", metrics.CodeAccuracy*100)
		}
		os.Exit(1)
	}
}

func displayResults(logger *log.Logger, metrics *testingpkg.ComprehensiveAccuracyMetrics) {
	separator := strings.Repeat("=", 80)
	logger.Println("\n" + separator)
	logger.Println("COMPREHENSIVE ACCURACY TEST RESULTS")
	logger.Println(separator)
	logger.Printf("Total Test Cases: %d", metrics.TotalTestCases)
	logger.Printf("Passed: %d", metrics.PassedTestCases)
	logger.Printf("Failed: %d", metrics.FailedTestCases)
	logger.Printf("Overall Accuracy: %.2f%%", metrics.OverallAccuracy*100)
	logger.Printf("Industry Accuracy: %.2f%% (target: 95%%)", metrics.IndustryAccuracy*100)
	logger.Printf("Code Accuracy: %.2f%% (target: 90%%)", metrics.CodeAccuracy*100)
	logger.Printf("  - MCC Accuracy: %.2f%%", metrics.MCCAccuracy*100)
	logger.Printf("  - NAICS Accuracy: %.2f%%", metrics.NAICSAccuracy*100)
	logger.Printf("  - SIC Accuracy: %.2f%%", metrics.SICAccuracy*100)
	logger.Printf("Average Processing Time: %v", metrics.AverageProcessingTime.Round(time.Millisecond))
	logger.Printf("Total Processing Time: %v", metrics.TotalProcessingTime.Round(time.Second))

	if len(metrics.AccuracyByCategory) > 0 {
		logger.Println("\nAccuracy by Category:")
		for cat, acc := range metrics.AccuracyByCategory {
			logger.Printf("  %s: %.2f%%", cat, acc*100)
		}
	}

	if len(metrics.AccuracyByIndustry) > 0 {
		logger.Println("\nAccuracy by Industry:")
		for ind, acc := range metrics.AccuracyByIndustry {
			logger.Printf("  %s: %.2f%%", ind, acc*100)
		}
	}

	logger.Println(separator)
}

func maskURL(url string) string {
	// Mask password in database URL
	if len(url) > 20 {
		return url[:10] + "***" + url[len(url)-10:]
	}
	return "***"
}

// initPythonMLService initializes the Python ML service if available
// Uses InitializeWithRetry for resilient initialization with exponential backoff (3 retries)
func initPythonMLService(endpoint string, logger *log.Logger) interface{} {
	service := infrastructure.NewPythonMLService(endpoint, logger)
	
	// Initialize with retry logic (3 retries) to handle transient startup issues
	// Increased timeout to accommodate retries: 10s base + (2s + 4s + 6s) retries = ~22s max
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := service.InitializeWithRetry(ctx, 3); err != nil {
		logger.Printf("⚠️  Failed to initialize Python ML Service after retries: %v", err)
		return nil
	}
	
	return service
}

// createIndustryServiceWithML creates an IndustryDetectionService that uses ML when available
func createIndustryServiceWithML(
	repo repository.KeywordRepository,
	mlClassifier *machine_learning.ContentClassifier,
	pythonMLService interface{},
	logger *log.Logger,
) *classification.IndustryDetectionService {
	// Use the new ML-enabled service constructor
	return classification.NewIndustryDetectionServiceWithML(repo, mlClassifier, pythonMLService, logger)
}

