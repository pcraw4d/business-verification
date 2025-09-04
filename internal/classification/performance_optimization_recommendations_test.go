package classification

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// createMockDBForPerformanceOptimizationRecommendations creates a mock database connection for testing
func createMockDBForPerformanceOptimizationRecommendations() *sql.DB {
	// In a real test, you would use a test database or mock
	// For now, we'll return nil and skip tests that require DB
	return nil
}

func TestPerformanceOptimizationRecommendations_GenerateDatabasePerformanceRecommendations(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	recommendations, err := por.GenerateDatabasePerformanceRecommendations(ctx)
	if err != nil {
		t.Fatalf("GenerateDatabasePerformanceRecommendations failed: %v", err)
	}

	t.Logf("Found %d database performance recommendations", len(recommendations))
	for _, rec := range recommendations {
		t.Logf("Database performance recommendation: %+v", rec)
	}
}

func TestPerformanceOptimizationRecommendations_GenerateClassificationPerformanceRecommendations(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	recommendations, err := por.GenerateClassificationPerformanceRecommendations(ctx)
	if err != nil {
		t.Fatalf("GenerateClassificationPerformanceRecommendations failed: %v", err)
	}

	t.Logf("Found %d classification performance recommendations", len(recommendations))
	for _, rec := range recommendations {
		t.Logf("Classification performance recommendation: %+v", rec)
	}
}

func TestPerformanceOptimizationRecommendations_GenerateSystemResourceRecommendations(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	recommendations, err := por.GenerateSystemResourceRecommendations(ctx)
	if err != nil {
		t.Fatalf("GenerateSystemResourceRecommendations failed: %v", err)
	}

	t.Logf("Found %d system resource recommendations", len(recommendations))
	for _, rec := range recommendations {
		t.Logf("System resource recommendation: %+v", rec)
	}
}

func TestPerformanceOptimizationRecommendations_GetAllPerformanceRecommendations(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	recommendations, err := por.GetAllPerformanceRecommendations(ctx)
	if err != nil {
		t.Fatalf("GetAllPerformanceRecommendations failed: %v", err)
	}

	t.Logf("Found %d total performance recommendations", len(recommendations))
	for _, rec := range recommendations {
		t.Logf("Performance recommendation: %+v", rec)
	}
}

func TestPerformanceOptimizationRecommendations_SavePerformanceRecommendations(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	savedCount, err := por.SavePerformanceRecommendations(ctx)
	if err != nil {
		t.Fatalf("SavePerformanceRecommendations failed: %v", err)
	}

	t.Logf("Saved %d performance recommendations", savedCount)
}

func TestPerformanceOptimizationRecommendations_GetRecommendationsByPriority(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test critical recommendations
	criticalRecommendations, err := por.GetRecommendationsByPriority(ctx, "CRITICAL")
	if err != nil {
		t.Fatalf("GetRecommendationsByPriority failed: %v", err)
	}

	t.Logf("Found %d critical recommendations", len(criticalRecommendations))
	for _, rec := range criticalRecommendations {
		t.Logf("Critical recommendation: %+v", rec)
	}

	// Test high priority recommendations
	highRecommendations, err := por.GetRecommendationsByPriority(ctx, "HIGH")
	if err != nil {
		t.Fatalf("GetRecommendationsByPriority failed: %v", err)
	}

	t.Logf("Found %d high priority recommendations", len(highRecommendations))
	for _, rec := range highRecommendations {
		t.Logf("High priority recommendation: %+v", rec)
	}
}

func TestPerformanceOptimizationRecommendations_GetRecommendationsByCategory(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test performance category recommendations
	performanceRecommendations, err := por.GetRecommendationsByCategory(ctx, "PERFORMANCE")
	if err != nil {
		t.Fatalf("GetRecommendationsByCategory failed: %v", err)
	}

	t.Logf("Found %d performance category recommendations", len(performanceRecommendations))
	for _, rec := range performanceRecommendations {
		t.Logf("Performance category recommendation: %+v", rec)
	}

	// Test storage category recommendations
	storageRecommendations, err := por.GetRecommendationsByCategory(ctx, "STORAGE")
	if err != nil {
		t.Fatalf("GetRecommendationsByCategory failed: %v", err)
	}

	t.Logf("Found %d storage category recommendations", len(storageRecommendations))
	for _, rec := range storageRecommendations {
		t.Logf("Storage category recommendation: %+v", rec)
	}
}

func TestPerformanceOptimizationRecommendations_ImplementRecommendation(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	recommendationID := "TEST_REC_001"
	implementedBy := "test-user"
	implementationNotes := "Test implementation"

	implemented, err := por.ImplementRecommendation(ctx, recommendationID, implementedBy, &implementationNotes)
	if err != nil {
		t.Fatalf("ImplementRecommendation failed: %v", err)
	}

	t.Logf("Recommendation implemented: %t", implemented)
}

func TestPerformanceOptimizationRecommendations_GetRecommendationStatistics(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	stats, err := por.GetRecommendationStatistics(ctx)
	if err != nil {
		t.Fatalf("GetRecommendationStatistics failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}

	t.Logf("Recommendation statistics: %+v", stats)
}

func TestPerformanceOptimizationRecommendations_ValidateRecommendationsSetup(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	validation, err := por.ValidateRecommendationsSetup(ctx)
	if err != nil {
		t.Fatalf("ValidateRecommendationsSetup failed: %v", err)
	}

	t.Logf("Found %d recommendation validation results", len(validation))
	for _, result := range validation {
		t.Logf("Recommendation validation: %+v", result)
	}
}

func TestPerformanceOptimizationRecommendations_GetCurrentRecommendationsStatus(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	status, err := por.GetCurrentRecommendationsStatus(ctx)
	if err != nil {
		t.Fatalf("GetCurrentRecommendationsStatus failed: %v", err)
	}

	t.Logf("Current recommendations status: %+v", status)
}

func TestPerformanceOptimizationRecommendations_GenerateAndSaveRecommendations(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	savedCount, err := por.GenerateAndSaveRecommendations(ctx)
	if err != nil {
		t.Fatalf("GenerateAndSaveRecommendations failed: %v", err)
	}

	t.Logf("Generated and saved %d recommendations", savedCount)
}

func TestPerformanceOptimizationRecommendations_GetPerformanceOptimizationSummary(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	summary, err := por.GetPerformanceOptimizationSummary(ctx)
	if err != nil {
		t.Fatalf("GetPerformanceOptimizationSummary failed: %v", err)
	}

	t.Logf("Performance optimization summary: %+v", summary)
}

// Benchmark tests for performance optimization recommendations
func BenchmarkPerformanceOptimizationRecommendations_GenerateDatabasePerformanceRecommendations(b *testing.B) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := por.GenerateDatabasePerformanceRecommendations(ctx)
		if err != nil {
			b.Fatalf("GenerateDatabasePerformanceRecommendations failed: %v", err)
		}
	}
}

func BenchmarkPerformanceOptimizationRecommendations_GetAllPerformanceRecommendations(b *testing.B) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := por.GetAllPerformanceRecommendations(ctx)
		if err != nil {
			b.Fatalf("GetAllPerformanceRecommendations failed: %v", err)
		}
	}
}

func BenchmarkPerformanceOptimizationRecommendations_SavePerformanceRecommendations(b *testing.B) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := por.SavePerformanceRecommendations(ctx)
		if err != nil {
			b.Fatalf("SavePerformanceRecommendations failed: %v", err)
		}
	}
}

// Integration test that would require a real database
func TestPerformanceOptimizationRecommendations_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This would require a real database connection
	// db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
	// if err != nil {
	//     t.Fatalf("Failed to connect to database: %v", err)
	// }
	// defer db.Close()

	// por := NewPerformanceOptimizationRecommendations(db)
	// ctx := context.Background()

	// Test all functions with real data
	t.Log("Integration test would run here with real database")
}

// Test error handling
func TestPerformanceOptimizationRecommendations_ErrorHandling(t *testing.T) {
	// Test with nil database
	por := NewPerformanceOptimizationRecommendations(nil)
	ctx := context.Background()

	// These should return errors due to nil database
	_, err := por.GetAllPerformanceRecommendations(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = por.GetCurrentRecommendationsStatus(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = por.SavePerformanceRecommendations(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}
}

// Test continuous monitoring
func TestPerformanceOptimizationRecommendations_ContinuousMonitoring(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start continuous monitoring with short interval
	go por.MonitorRecommendationsContinuously(ctx, 1*time.Second)

	// Wait for context to timeout
	<-ctx.Done()

	t.Log("Continuous performance recommendations monitoring test completed")
}

// Test recommendation management
func TestPerformanceOptimizationRecommendations_RecommendationManagement(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test recommendation generation
	allRecommendations, err := por.GetAllPerformanceRecommendations(ctx)
	if err != nil {
		t.Fatalf("GetAllPerformanceRecommendations failed: %v", err)
	}

	// Test saving recommendations
	savedCount, err := por.SavePerformanceRecommendations(ctx)
	if err != nil {
		t.Fatalf("SavePerformanceRecommendations failed: %v", err)
	}

	// Test getting recommendations by priority
	criticalRecommendations, err := por.GetRecommendationsByPriority(ctx, "CRITICAL")
	if err != nil {
		t.Fatalf("GetRecommendationsByPriority failed: %v", err)
	}

	// Test getting recommendations by category
	performanceRecommendations, err := por.GetRecommendationsByCategory(ctx, "PERFORMANCE")
	if err != nil {
		t.Fatalf("GetRecommendationsByCategory failed: %v", err)
	}

	// Test implementing a recommendation
	if len(allRecommendations) > 0 {
		recommendationID := allRecommendations[0].RecommendationID
		implementedBy := "test-user"
		implementationNotes := "Test implementation"

		implemented, err := por.ImplementRecommendation(ctx, recommendationID, implementedBy, &implementationNotes)
		if err != nil {
			t.Fatalf("ImplementRecommendation failed: %v", err)
		}

		t.Logf("Recommendation management test: total=%d, saved=%d, critical=%d, performance=%d, implemented=%t",
			len(allRecommendations), savedCount, len(criticalRecommendations), len(performanceRecommendations), implemented)
	}
}

// Test recommendation statistics
func TestPerformanceOptimizationRecommendations_RecommendationStatistics(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test recommendation statistics
	stats, err := por.GetRecommendationStatistics(ctx)
	if err != nil {
		t.Fatalf("GetRecommendationStatistics failed: %v", err)
	}

	// Test current recommendations status
	status, err := por.GetCurrentRecommendationsStatus(ctx)
	if err != nil {
		t.Fatalf("GetCurrentRecommendationsStatus failed: %v", err)
	}

	// Test performance optimization summary
	summary, err := por.GetPerformanceOptimizationSummary(ctx)
	if err != nil {
		t.Fatalf("GetPerformanceOptimizationSummary failed: %v", err)
	}

	t.Logf("Recommendation statistics: %+v", stats)
	t.Logf("Recommendations status: %+v", status)
	t.Logf("Performance optimization summary keys: %d", len(summary))
}

// Test recommendation validation
func TestPerformanceOptimizationRecommendations_RecommendationValidation(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test recommendations setup validation
	validation, err := por.ValidateRecommendationsSetup(ctx)
	if err != nil {
		t.Fatalf("ValidateRecommendationsSetup failed: %v", err)
	}

	t.Logf("Recommendation validation results: %d", len(validation))
	for _, result := range validation {
		t.Logf("Validation: %+v", result)
	}
}

// Test recommendation scenarios
func TestPerformanceOptimizationRecommendations_RecommendationScenarios(t *testing.T) {
	por := NewPerformanceOptimizationRecommendations(createMockDBForPerformanceOptimizationRecommendations())
	if por.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test database performance recommendations
	dbRecommendations, err := por.GenerateDatabasePerformanceRecommendations(ctx)
	if err != nil {
		t.Fatalf("GenerateDatabasePerformanceRecommendations failed: %v", err)
	}

	// Test classification performance recommendations
	classificationRecommendations, err := por.GenerateClassificationPerformanceRecommendations(ctx)
	if err != nil {
		t.Fatalf("GenerateClassificationPerformanceRecommendations failed: %v", err)
	}

	// Test system resource recommendations
	systemRecommendations, err := por.GenerateSystemResourceRecommendations(ctx)
	if err != nil {
		t.Fatalf("GenerateSystemResourceRecommendations failed: %v", err)
	}

	// Test all recommendations
	allRecommendations, err := por.GetAllPerformanceRecommendations(ctx)
	if err != nil {
		t.Fatalf("GetAllPerformanceRecommendations failed: %v", err)
	}

	t.Logf("Recommendation scenarios: DB=%d, Classification=%d, System=%d, Total=%d",
		len(dbRecommendations), len(classificationRecommendations), len(systemRecommendations), len(allRecommendations))
}
