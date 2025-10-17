package test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/classification"
	"go.uber.org/zap"
)

// TestCrosswalkAnalyzer tests the crosswalk analyzer functionality
func TestCrosswalkAnalyzer(t *testing.T) {
	// Setup test database connection
	db, err := setupCrosswalkTestDatabase()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer db.Close()

	// Setup logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Create crosswalk analyzer
	config := &classification.CrosswalkConfig{
		MinConfidenceScore:    0.80,
		MaxMappingDistance:    2,
		EnableValidation:      true,
		EnableAutoMapping:     true,
		ValidationTimeout:     30 * time.Second,
		BatchSize:             100,
		EnableLogging:         true,
		EnablePerformanceMode: false,
	}

	analyzer := classification.NewCrosswalkAnalyzer(db, logger, config)

	// Test MCC to Industry mapping
	t.Run("MapMCCCodesToIndustries", func(t *testing.T) {
		ctx := context.Background()
		result, err := analyzer.MapMCCCodesToIndustries(ctx)
		if err != nil {
			t.Fatalf("Failed to map MCC codes to industries: %v", err)
		}

		// Validate result
		if result == nil {
			t.Fatal("Result is nil")
		}

		if result.AnalysisID == "" {
			t.Error("Analysis ID is empty")
		}

		if result.TotalMappings == 0 {
			t.Error("No mappings found")
		}

		// Log results
		logger.Info("MCC to Industry mapping test completed",
			zap.String("analysis_id", result.AnalysisID),
			zap.Int("total_mappings", result.TotalMappings),
			zap.Int("valid_mappings", result.ValidMappings),
			zap.Duration("duration", result.Duration))

		// Validate mappings
		for mccCode, industries := range result.MCCToIndustryMappings {
			if len(industries) == 0 {
				t.Errorf("No industries mapped for MCC code: %s", mccCode)
			}
		}

		// Save test results
		if err := saveTestResults("mcc_industry_mapping_test", result); err != nil {
			t.Errorf("Failed to save test results: %v", err)
		}
	})

	// Test crosswalk mapping validation
	t.Run("ValidateCrosswalkMappings", func(t *testing.T) {
		ctx := context.Background()
		result, err := analyzer.MapMCCCodesToIndustries(ctx)
		if err != nil {
			t.Fatalf("Failed to map MCC codes to industries: %v", err)
		}

		// Validate that validation results are present
		if result.ValidationResults == nil {
			t.Error("Validation results are nil")
		} else {
			if result.ValidationResults.TotalValidations == 0 {
				t.Error("No validations performed")
			}

			if result.ValidationResults.ValidationAccuracy < 0.0 || result.ValidationResults.ValidationAccuracy > 1.0 {
				t.Error("Invalid validation accuracy")
			}
		}
	})

	// Test crosswalk mapping saving
	t.Run("SaveCrosswalkMappings", func(t *testing.T) {
		ctx := context.Background()
		result, err := analyzer.MapMCCCodesToIndustries(ctx)
		if err != nil {
			t.Fatalf("Failed to map MCC codes to industries: %v", err)
		}

		// Save mappings to database
		if err := analyzer.SaveCrosswalkMappings(ctx, result.CrosswalkMappings); err != nil {
			t.Errorf("Failed to save crosswalk mappings: %v", err)
		}

		// Verify mappings were saved
		query := "SELECT COUNT(*) FROM crosswalk_mappings WHERE source_system = 'MCC'"
		var count int
		if err := db.QueryRowContext(ctx, query).Scan(&count); err != nil {
			t.Errorf("Failed to verify saved mappings: %v", err)
		}

		if count == 0 {
			t.Error("No crosswalk mappings found in database")
		}
	})

	// Test NAICS to Industry mapping
	t.Run("MapNAICSCodesToIndustries", func(t *testing.T) {
		ctx := context.Background()
		result, err := analyzer.MapNAICSCodesToIndustries(ctx)
		if err != nil {
			t.Fatalf("Failed to map NAICS codes to industries: %v", err)
		}

		// Validate result
		if result == nil {
			t.Fatal("Result is nil")
		}

		if result.AnalysisID == "" {
			t.Error("Analysis ID is empty")
		}

		if result.TotalMappings == 0 {
			t.Error("No mappings found")
		}

		// Log results
		logger.Info("NAICS to Industry mapping test completed",
			zap.String("analysis_id", result.AnalysisID),
			zap.Int("total_mappings", result.TotalMappings),
			zap.Int("valid_mappings", result.ValidMappings),
			zap.Duration("duration", result.Duration))

		// Validate mappings
		for naicsCode, industries := range result.NAICSToIndustryMappings {
			if len(industries) == 0 {
				t.Errorf("No industries mapped for NAICS code: %s", naicsCode)
			}
		}

		// Save test results
		if err := saveTestResults("naics_industry_mapping_test", result); err != nil {
			t.Errorf("Failed to save test results: %v", err)
		}
	})

	// Test SIC to Industry mapping
	t.Run("MapSICCodesToIndustries", func(t *testing.T) {
		ctx := context.Background()
		result, err := analyzer.MapSICCodesToIndustries(ctx)
		if err != nil {
			t.Fatalf("Failed to map SIC codes to industries: %v", err)
		}

		// Validate result
		if result == nil {
			t.Fatal("Result is nil")
		}

		if result.AnalysisID == "" {
			t.Error("Analysis ID is empty")
		}

		if result.TotalMappings == 0 {
			t.Error("No mappings found")
		}

		// Log results
		logger.Info("SIC to Industry mapping test completed",
			zap.String("analysis_id", result.AnalysisID),
			zap.Int("total_mappings", result.TotalMappings),
			zap.Int("valid_mappings", result.ValidMappings),
			zap.Duration("duration", result.Duration))

		// Validate mappings
		for sicCode, industries := range result.SICToIndustryMappings {
			if len(industries) == 0 {
				t.Errorf("No industries mapped for SIC code: %s", sicCode)
			}
		}

		// Save test results
		if err := saveTestResults("sic_industry_mapping_test", result); err != nil {
			t.Errorf("Failed to save test results: %v", err)
		}
	})
}

// setupCrosswalkTestDatabase sets up a test database connection for crosswalk tests
func setupCrosswalkTestDatabase() (*sql.DB, error) {
	// Get database URL from environment or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/kyb_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// saveTestResults saves test results to a file
func saveTestResults(testName string, result *classification.CrosswalkAnalysisResult) error {
	// Create test results directory
	resultsDir := "test_results"
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}

	// Create filename with timestamp
	filename := fmt.Sprintf("%s/%s_%d.json", resultsDir, testName, time.Now().Unix())

	// Write results to file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create results file: %w", err)
	}
	defer file.Close()

	// Marshal and write JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	if _, err := file.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}

	log.Printf("✅ Test results saved to: %s", filename)
	return nil
}

// BenchmarkCrosswalkAnalyzer benchmarks the crosswalk analyzer performance
func BenchmarkCrosswalkAnalyzer(b *testing.B) {
	// Setup test database connection
	db, err := setupCrosswalkTestDatabase()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer db.Close()

	// Setup logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Create crosswalk analyzer
	config := &classification.CrosswalkConfig{
		MinConfidenceScore:    0.80,
		MaxMappingDistance:    2,
		EnableValidation:      true,
		EnableAutoMapping:     true,
		ValidationTimeout:     30 * time.Second,
		BatchSize:             100,
		EnableLogging:         false, // Disable logging for benchmarks
		EnablePerformanceMode: true,  // Enable performance mode
	}

	analyzer := classification.NewCrosswalkAnalyzer(db, logger, config)

	// Benchmark MCC to Industry mapping
	b.Run("MapMCCCodesToIndustries", func(b *testing.B) {
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result, err := analyzer.MapMCCCodesToIndustries(ctx)
			if err != nil {
				b.Fatalf("Failed to map MCC codes to industries: %v", err)
			}

			// Prevent optimization
			_ = result
		}
	})
}

// TestCrosswalkAnalyzerIntegration tests the crosswalk analyzer with real data
func TestCrosswalkAnalyzerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test database connection
	db, err := setupCrosswalkTestDatabase()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer db.Close()

	// Setup logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Create crosswalk analyzer
	config := &classification.CrosswalkConfig{
		MinConfidenceScore:    0.80,
		MaxMappingDistance:    2,
		EnableValidation:      true,
		EnableAutoMapping:     true,
		ValidationTimeout:     30 * time.Second,
		BatchSize:             100,
		EnableLogging:         true,
		EnablePerformanceMode: false,
	}

	analyzer := classification.NewCrosswalkAnalyzer(db, logger, config)

	// Test with real data
	ctx := context.Background()
	result, err := analyzer.MapMCCCodesToIndustries(ctx)
	if err != nil {
		t.Fatalf("Failed to map MCC codes to industries: %v", err)
	}

	// Validate integration results
	if result.TotalMappings == 0 {
		t.Error("No mappings found in integration test")
	}

	// Check for specific MCC codes that should have mappings
	expectedMCCCodes := []string{"5734", "7372", "7373", "8062", "5047", "5122"}
	for _, mccCode := range expectedMCCCodes {
		if industries, exists := result.MCCToIndustryMappings[mccCode]; !exists || len(industries) == 0 {
			t.Errorf("Expected MCC code %s to have industry mappings", mccCode)
		}
	}

	// Validate confidence scores
	for _, mapping := range result.CrosswalkMappings {
		if mapping.ConfidenceScore < 0.0 || mapping.ConfidenceScore > 1.0 {
			t.Errorf("Invalid confidence score: %f", mapping.ConfidenceScore)
		}
	}

	// Save integration test results
	if err := saveTestResults("mcc_industry_mapping_integration", result); err != nil {
		t.Errorf("Failed to save integration test results: %v", err)
	}
}
