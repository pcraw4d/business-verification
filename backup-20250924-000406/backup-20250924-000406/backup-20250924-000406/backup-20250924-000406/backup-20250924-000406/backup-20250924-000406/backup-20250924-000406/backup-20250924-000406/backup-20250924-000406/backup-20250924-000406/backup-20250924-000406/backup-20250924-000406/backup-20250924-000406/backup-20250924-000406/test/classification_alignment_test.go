package test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/classification"
)

// TestClassificationAlignmentTypes tests the classification alignment types and structures
func TestClassificationAlignmentTypes(t *testing.T) {
	t.Run("AlignmentConfig", func(t *testing.T) {
		config := &classification.AlignmentConfig{
			EnableMCCAlignment:       true,
			EnableNAICSAlignment:     true,
			EnableSICAlignment:       true,
			MinAlignmentScore:        0.8,
			MaxAlignmentTime:         60,
			EnableConflictResolution: true,
			EnableGapAnalysis:        true,
		}

		if !config.EnableMCCAlignment {
			t.Error("Expected MCC alignment to be enabled")
		}
		if config.MinAlignmentScore != 0.8 {
			t.Errorf("Expected min alignment score 0.8, got %f", config.MinAlignmentScore)
		}
	})

	t.Run("AlignmentResult", func(t *testing.T) {
		result := &classification.AlignmentResult{
			AnalysisID:           "test_analysis",
			TotalIndustries:      10,
			AlignedIndustries:    8,
			MisalignedIndustries: 2,
			Conflicts:            []classification.ClassificationConflict{},
			Gaps:                 []classification.ClassificationGap{},
			Recommendations:      []classification.AlignmentRecommendation{},
			AlignmentScores:      make(map[string]float64),
		}

		if result.AnalysisID != "test_analysis" {
			t.Errorf("Expected analysis ID 'test_analysis', got %s", result.AnalysisID)
		}
		if result.TotalIndustries != 10 {
			t.Errorf("Expected 10 total industries, got %d", result.TotalIndustries)
		}
		if result.AlignedIndustries != 8 {
			t.Errorf("Expected 8 aligned industries, got %d", result.AlignedIndustries)
		}
	})

	t.Run("Industry", func(t *testing.T) {
		industry := classification.Industry{
			ID:          1,
			Name:        "Technology",
			Description: "Technology and software companies",
		}

		if industry.ID != 1 {
			t.Errorf("Expected industry ID 1, got %d", industry.ID)
		}
		if industry.Name != "Technology" {
			t.Errorf("Expected industry name 'Technology', got %s", industry.Name)
		}
	})

	t.Run("ConflictTypes", func(t *testing.T) {
		conflict := classification.ClassificationConflict{
			IndustryID:   1,
			IndustryName: "Test Industry",
			ConflictType: classification.ConflictTypeConfidenceMismatch,
			Severity:     classification.ConflictSeverityMedium,
			CreatedAt:    time.Now(),
		}

		if conflict.ConflictType != classification.ConflictTypeConfidenceMismatch {
			t.Errorf("Expected confidence mismatch conflict type, got %s", conflict.ConflictType)
		}
		if conflict.Severity != classification.ConflictSeverityMedium {
			t.Errorf("Expected medium severity, got %s", conflict.Severity)
		}
	})

	t.Run("GapTypes", func(t *testing.T) {
		gap := classification.ClassificationGap{
			IndustryID:   1,
			IndustryName: "Test Industry",
			GapType:      classification.GapTypeMissingMCC,
			Severity:     classification.GapSeverityHigh,
			CreatedAt:    time.Now(),
		}

		if gap.GapType != classification.GapTypeMissingMCC {
			t.Errorf("Expected missing MCC gap type, got %s", gap.GapType)
		}
		if gap.Severity != classification.GapSeverityHigh {
			t.Errorf("Expected high severity, got %s", gap.Severity)
		}
	})
}

// TestClassificationAlignment tests the classification alignment engine
func TestClassificationAlignment(t *testing.T) {
	// Skip if database not available
	t.Skip("Database connection required for integration test")

	// Setup test database
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

	// Create alignment engine
	config := &classification.AlignmentConfig{
		EnableMCCAlignment:       true,
		EnableNAICSAlignment:     true,
		EnableSICAlignment:       true,
		MinAlignmentScore:        0.8,
		MaxAlignmentTime:         60,
		EnableConflictResolution: true,
		EnableGapAnalysis:        true,
	}

	engine := classification.NewClassificationAlignmentEngine(db, logger, config)

	// Test alignment analysis
	t.Run("AnalyzeClassificationAlignment", func(t *testing.T) {
		ctx := context.Background()
		result, err := engine.AnalyzeClassificationAlignment(ctx)
		if err != nil {
			t.Fatalf("Failed to analyze classification alignment: %v", err)
		}

		// Validate result
		if result == nil {
			t.Fatal("Result is nil")
		}

		if result.AnalysisID == "" {
			t.Error("Analysis ID is empty")
		}

		if result.TotalIndustries == 0 {
			t.Error("No industries found")
		}

		// Log results
		logger.Info("Classification alignment analysis completed",
			zap.String("analysis_id", result.AnalysisID),
			zap.Int("total_industries", result.TotalIndustries),
			zap.Int("aligned_industries", result.AlignedIndustries),
			zap.Int("misaligned_industries", result.MisalignedIndustries),
			zap.Int("conflicts", len(result.Conflicts)),
			zap.Int("gaps", len(result.Gaps)),
			zap.Int("recommendations", len(result.Recommendations)),
			zap.Duration("duration", result.Duration))

		// Validate alignment scores
		if result.AlignmentScores == nil {
			t.Error("Alignment scores are nil")
		}

		// Check for expected alignment scores
		expectedScores := []string{"overall", "mcc", "naics", "sic"}
		for _, scoreType := range expectedScores {
			if _, exists := result.AlignmentScores[scoreType]; !exists {
				t.Errorf("Missing alignment score for type: %s", scoreType)
			}
		}

		// Validate summary
		if result.Summary.OverallAlignmentScore < 0 || result.Summary.OverallAlignmentScore > 1 {
			t.Errorf("Invalid overall alignment score: %f", result.Summary.OverallAlignmentScore)
		}

		// Save test results
		if err := saveAlignmentTestResults("classification_alignment_test", result); err != nil {
			t.Errorf("Failed to save test results: %v", err)
		}
	})

	// Test individual alignment components
	t.Run("MCCAlignment", func(t *testing.T) {
		ctx := context.Background()

		// Create test industry
		industry := classification.Industry{
			ID:          1,
			Name:        "Test Technology Industry",
			Description: "Test technology industry for alignment testing",
		}

		// Create test mappings
		mappings := []classification.CrosswalkMapping{
			{
				ID:              1,
				IndustryID:      1,
				MCCCode:         "5734",
				NAICSCode:       "",
				SICCode:         "",
				Description:     "Computer Software Stores",
				ConfidenceScore: 0.9,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
		}

		// Test MCC alignment analysis
		conflicts, gaps, err := engine.AnalyzeMCCAlignment(ctx, industry, mappings)
		if err != nil {
			t.Fatalf("Failed to analyze MCC alignment: %v", err)
		}

		// Validate results
		if len(conflicts) > 0 {
			t.Logf("Found %d MCC conflicts", len(conflicts))
			for _, conflict := range conflicts {
				if conflict.ConflictType != classification.ConflictTypeConfidenceMismatch {
					t.Errorf("Unexpected conflict type: %s", conflict.ConflictType)
				}
			}
		}

		if len(gaps) > 0 {
			t.Logf("Found %d MCC gaps", len(gaps))
			for _, gap := range gaps {
				if gap.GapType != classification.GapTypeMissingMCC {
					t.Errorf("Unexpected gap type: %s", gap.GapType)
				}
			}
		}
	})

	// Test NAICS alignment
	t.Run("NAICSAlignment", func(t *testing.T) {
		ctx := context.Background()

		// Create test industry
		industry := classification.Industry{
			ID:          2,
			Name:        "Test Healthcare Industry",
			Description: "Test healthcare industry for alignment testing",
		}

		// Create test mappings
		mappings := []classification.CrosswalkMapping{
			{
				ID:              2,
				IndustryID:      2,
				MCCCode:         "",
				NAICSCode:       "621111",
				SICCode:         "",
				Description:     "Offices of Physicians",
				ConfidenceScore: 0.95,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
		}

		// Test NAICS alignment analysis
		conflicts, gaps, err := engine.AnalyzeNAICSAlignment(ctx, industry, mappings)
		if err != nil {
			t.Fatalf("Failed to analyze NAICS alignment: %v", err)
		}

		// Validate results
		if len(conflicts) > 0 {
			t.Logf("Found %d NAICS conflicts", len(conflicts))
			for _, conflict := range conflicts {
				if conflict.ConflictType != classification.ConflictTypeHierarchyMismatch {
					t.Errorf("Unexpected conflict type: %s", conflict.ConflictType)
				}
			}
		}

		if len(gaps) > 0 {
			t.Logf("Found %d NAICS gaps", len(gaps))
			for _, gap := range gaps {
				if gap.GapType != classification.GapTypeMissingNAICS {
					t.Errorf("Unexpected gap type: %s", gap.GapType)
				}
			}
		}
	})

	// Test SIC alignment
	t.Run("SICAlignment", func(t *testing.T) {
		ctx := context.Background()

		// Create test industry
		industry := classification.Industry{
			ID:          3,
			Name:        "Test Manufacturing Industry",
			Description: "Test manufacturing industry for alignment testing",
		}

		// Create test mappings
		mappings := []classification.CrosswalkMapping{
			{
				ID:              3,
				IndustryID:      3,
				MCCCode:         "",
				NAICSCode:       "",
				SICCode:         "3571",
				Description:     "Electronic Computers",
				ConfidenceScore: 0.85,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
		}

		// Test SIC alignment analysis
		conflicts, gaps, err := engine.AnalyzeSICAlignment(ctx, industry, mappings)
		if err != nil {
			t.Fatalf("Failed to analyze SIC alignment: %v", err)
		}

		// Validate results
		if len(conflicts) > 0 {
			t.Logf("Found %d SIC conflicts", len(conflicts))
			for _, conflict := range conflicts {
				if conflict.ConflictType != classification.ConflictTypeHierarchyMismatch {
					t.Errorf("Unexpected conflict type: %s", conflict.ConflictType)
				}
			}
		}

		if len(gaps) > 0 {
			t.Logf("Found %d SIC gaps", len(gaps))
			for _, gap := range gaps {
				if gap.GapType != classification.GapTypeMissingSIC {
					t.Errorf("Unexpected gap type: %s", gap.GapType)
				}
			}
		}
	})

	// Test alignment score calculation
	t.Run("AlignmentScoreCalculation", func(t *testing.T) {
		ctx := context.Background()

		// Create test result
		result := &classification.AlignmentResult{
			TotalIndustries:      10,
			AlignedIndustries:    8,
			MisalignedIndustries: 2,
			AlignmentScores:      make(map[string]float64),
		}

		// Test score calculation
		err := engine.CalculateAlignmentScores(ctx, result)
		if err != nil {
			t.Fatalf("Failed to calculate alignment scores: %v", err)
		}

		// Validate overall score
		expectedOverall := 0.8 // 8/10
		if result.AlignmentScores["overall"] != expectedOverall {
			t.Errorf("Expected overall score %f, got %f", expectedOverall, result.AlignmentScores["overall"])
		}

		// Validate score range
		for scoreType, score := range result.AlignmentScores {
			if score < 0 || score > 1 {
				t.Errorf("Invalid score for %s: %f", scoreType, score)
			}
		}
	})

	// Test recommendation generation
	t.Run("RecommendationGeneration", func(t *testing.T) {
		// Create test result with conflicts and gaps
		result := &classification.AlignmentResult{
			Conflicts: []classification.ClassificationConflict{
				{
					IndustryID:   1,
					IndustryName: "Test Industry",
					ConflictType: classification.ConflictTypeConfidenceMismatch,
					Severity:     classification.ConflictSeverityMedium,
					CreatedAt:    time.Now(),
				},
				{
					IndustryID:   2,
					IndustryName: "Test Industry 2",
					ConflictType: classification.ConflictTypeHierarchyMismatch,
					Severity:     classification.ConflictSeverityHigh,
					CreatedAt:    time.Now(),
				},
			},
			Gaps: []classification.ClassificationGap{
				{
					IndustryID:   3,
					IndustryName: "Test Industry 3",
					GapType:      classification.GapTypeMissingMCC,
					Severity:     classification.GapSeverityHigh,
					CreatedAt:    time.Now(),
				},
			},
		}

		// Test recommendation generation
		recommendations := engine.GenerateAlignmentRecommendations(result)
		if len(recommendations) == 0 {
			t.Error("No recommendations generated")
		}

		// Validate recommendations
		for _, rec := range recommendations {
			if rec.RecommendationID == "" {
				t.Error("Recommendation ID is empty")
			}
			if rec.Title == "" {
				t.Error("Recommendation title is empty")
			}
			if rec.Description == "" {
				t.Error("Recommendation description is empty")
			}
			if len(rec.ActionItems) == 0 {
				t.Error("No action items in recommendation")
			}
		}

		logger.Info("Generated recommendations",
			zap.Int("count", len(recommendations)))
	})

	// Test alignment summary creation
	t.Run("AlignmentSummaryCreation", func(t *testing.T) {
		// Create test result
		result := &classification.AlignmentResult{
			AlignmentScores: map[string]float64{
				"overall": 0.8,
				"mcc":     0.9,
				"naics":   0.7,
				"sic":     0.85,
			},
			Conflicts: []classification.ClassificationConflict{
				{Severity: classification.ConflictSeverityCritical},
				{Severity: classification.ConflictSeverityHigh},
				{Severity: classification.ConflictSeverityMedium},
			},
			Gaps: []classification.ClassificationGap{
				{Severity: classification.GapSeverityHigh},
				{Severity: classification.GapSeverityLow},
			},
		}

		// Test summary creation
		summary := engine.CreateAlignmentSummary(result)

		// Validate summary
		if summary.OverallAlignmentScore != 0.8 {
			t.Errorf("Expected overall score 0.8, got %f", summary.OverallAlignmentScore)
		}
		if summary.MCCAlignmentScore != 0.9 {
			t.Errorf("Expected MCC score 0.9, got %f", summary.MCCAlignmentScore)
		}
		if summary.TotalConflicts != 3 {
			t.Errorf("Expected 3 conflicts, got %d", summary.TotalConflicts)
		}
		if summary.TotalGaps != 2 {
			t.Errorf("Expected 2 gaps, got %d", summary.TotalGaps)
		}
		if summary.CriticalIssues != 1 {
			t.Errorf("Expected 1 critical issue, got %d", summary.CriticalIssues)
		}
		if summary.HighPriorityIssues != 2 {
			t.Errorf("Expected 2 high priority issues, got %d", summary.HighPriorityIssues)
		}
		if summary.MediumPriorityIssues != 1 {
			t.Errorf("Expected 1 medium priority issue, got %d", summary.MediumPriorityIssues)
		}
		if summary.LowPriorityIssues != 1 {
			t.Errorf("Expected 1 low priority issue, got %d", summary.LowPriorityIssues)
		}
	})

	// Test performance
	t.Run("AlignmentPerformance", func(t *testing.T) {
		ctx := context.Background()

		startTime := time.Now()
		result, err := engine.AnalyzeClassificationAlignment(ctx)
		duration := time.Since(startTime)

		if err != nil {
			t.Fatalf("Failed to analyze classification alignment: %v", err)
		}

		// Check that analysis completed within reasonable time
		if duration > 30*time.Second {
			t.Errorf("Alignment analysis took too long: %v", duration)
		}

		// Validate result completeness
		if result.TotalIndustries == 0 {
			t.Error("No industries analyzed")
		}

		logger.Info("Alignment performance test completed",
			zap.Duration("duration", duration),
			zap.Int("total_industries", result.TotalIndustries),
			zap.Duration("avg_industry_time", duration/time.Duration(result.TotalIndustries)))
	})
}

// Helper functions for testing
func saveAlignmentTestResults(testName string, result *classification.AlignmentResult) error {
	// Create test results directory if it doesn't exist
	if err := os.MkdirAll("test_results", 0755); err != nil {
		return err
	}

	// Create filename with timestamp
	filename := fmt.Sprintf("test_results/%s_%d.json", testName, time.Now().Unix())

	// Marshal result to JSON
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(filename, data, 0644)
}

// Benchmark tests
func BenchmarkClassificationAlignment(b *testing.B) {
	// Setup test database
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

	// Create alignment engine
	config := &classification.AlignmentConfig{
		EnableMCCAlignment:       true,
		EnableNAICSAlignment:     true,
		EnableSICAlignment:       true,
		MinAlignmentScore:        0.8,
		MaxAlignmentTime:         60,
		EnableConflictResolution: true,
		EnableGapAnalysis:        true,
	}

	engine := classification.NewClassificationAlignmentEngine(db, logger, config)

	// Benchmark alignment analysis
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		result, err := engine.AnalyzeClassificationAlignment(ctx)
		if err != nil {
			b.Fatalf("Failed to analyze classification alignment: %v", err)
		}

		if result == nil {
			b.Fatal("Alignment result is nil")
		}
	}
}
