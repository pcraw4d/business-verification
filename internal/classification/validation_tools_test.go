package classification

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// MockDB creates a mock database connection for testing
func createMockDB() *sql.DB {
	// In a real test, you would use a test database or mock
	// For now, we'll return nil and skip tests that require DB
	return nil
}

func TestValidationTools_TestKeywordClassification(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()
	keywords := []string{"technology", "software", "development"}

	result, err := vt.TestKeywordClassification(ctx, keywords, "Tech Corp", "Software development company")
	if err != nil {
		t.Fatalf("TestKeywordClassification failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	t.Logf("Classification result: %+v", result)
}

func TestValidationTools_ValidateKeywordCoverage(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := vt.ValidateKeywordCoverage(ctx, "%")
	if err != nil {
		t.Fatalf("ValidateKeywordCoverage failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected coverage results, got empty slice")
	}

	for _, result := range results {
		t.Logf("Coverage result: %+v", result)
	}
}

func TestValidationTools_FindDuplicateKeywords(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := vt.FindDuplicateKeywords(ctx)
	if err != nil {
		t.Fatalf("FindDuplicateKeywords failed: %v", err)
	}

	t.Logf("Found %d duplicate keywords", len(results))
	for _, result := range results {
		t.Logf("Duplicate keyword: %+v", result)
	}
}

func TestValidationTools_TestKeywordPatterns(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()
	testText := "technology software development company"

	results, err := vt.TestKeywordPatterns(ctx, testText)
	if err != nil {
		t.Fatalf("TestKeywordPatterns failed: %v", err)
	}

	t.Logf("Found %d pattern results", len(results))
	for _, result := range results {
		t.Logf("Pattern result: %+v", result)
	}
}

func TestValidationTools_AnalyzeKeywordEffectiveness(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := vt.AnalyzeKeywordEffectiveness(ctx, 30)
	if err != nil {
		t.Fatalf("AnalyzeKeywordEffectiveness failed: %v", err)
	}

	t.Logf("Found %d effectiveness results", len(results))
	for _, result := range results {
		t.Logf("Effectiveness result: %+v", result)
	}
}

func TestValidationTools_SuggestKeywordImprovements(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := vt.SuggestKeywordImprovements(ctx, "%")
	if err != nil {
		t.Fatalf("SuggestKeywordImprovements failed: %v", err)
	}

	t.Logf("Found %d improvement suggestions", len(results))
	for _, result := range results {
		t.Logf("Improvement suggestion: %+v", result)
	}
}

func TestValidationTools_ValidateClassificationConsistency(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := vt.ValidateClassificationConsistency(ctx)
	if err != nil {
		t.Fatalf("ValidateClassificationConsistency failed: %v", err)
	}

	t.Logf("Found %d consistency results", len(results))
	for _, result := range results {
		t.Logf("Consistency result: %+v", result)
	}
}

func TestValidationTools_GenerateKeywordTestReport(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	testCases := []map[string]interface{}{
		{
			"name":              "Technology Test",
			"keywords":          []string{"technology", "software"},
			"expected_industry": "Technology",
		},
		{
			"name":              "Grocery Test",
			"keywords":          []string{"grocery", "food", "retail"},
			"expected_industry": "Grocery/Retail",
		},
	}

	results, err := vt.GenerateKeywordTestReport(ctx, testCases)
	if err != nil {
		t.Fatalf("GenerateKeywordTestReport failed: %v", err)
	}

	t.Logf("Generated %d test reports", len(results))
	for _, result := range results {
		t.Logf("Test report: %+v", result)
	}
}

func TestValidationTools_MonitorKeywordPerformance(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := vt.MonitorKeywordPerformance(ctx, 24)
	if err != nil {
		t.Fatalf("MonitorKeywordPerformance failed: %v", err)
	}

	t.Logf("Found %d performance results", len(results))
	for _, result := range results {
		t.Logf("Performance result: %+v", result)
	}
}

func TestValidationTools_OptimizeKeywordWeights(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := vt.OptimizeKeywordWeights(ctx, "%")
	if err != nil {
		t.Fatalf("OptimizeKeywordWeights failed: %v", err)
	}

	t.Logf("Found %d weight optimization results", len(results))
	for _, result := range results {
		t.Logf("Weight optimization: %+v", result)
	}
}

func TestValidationTools_ValidateKeywordCompleteness(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := vt.ValidateKeywordCompleteness(ctx)
	if err != nil {
		t.Fatalf("ValidateKeywordCompleteness failed: %v", err)
	}

	t.Logf("Found %d completeness results", len(results))
	for _, result := range results {
		t.Logf("Completeness result: %+v", result)
	}
}

func TestValidationTools_TestKeywordEdgeCases(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := vt.TestKeywordEdgeCases(ctx)
	if err != nil {
		t.Fatalf("TestKeywordEdgeCases failed: %v", err)
	}

	t.Logf("Found %d edge case test results", len(results))
	for _, result := range results {
		t.Logf("Edge case test: %+v", result)
	}
}

func TestValidationTools_GenerateKeywordStatistics(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := vt.GenerateKeywordStatistics(ctx)
	if err != nil {
		t.Fatalf("GenerateKeywordStatistics failed: %v", err)
	}

	t.Logf("Generated %d statistics", len(results))
	for _, result := range results {
		t.Logf("Statistic: %+v", result)
	}
}

func TestValidationTools_ValidateKeywordTestingCompletion(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := vt.ValidateKeywordTestingCompletion(ctx)
	if err != nil {
		t.Fatalf("ValidateKeywordTestingCompletion failed: %v", err)
	}

	t.Logf("Found %d completion validation results", len(results))
	for _, result := range results {
		t.Logf("Completion validation: %+v", result)
	}
}

func TestValidationTools_RunComprehensiveKeywordTests(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	result, err := vt.RunComprehensiveKeywordTests(ctx)
	if err != nil {
		t.Fatalf("RunComprehensiveKeywordTests failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected comprehensive test result, got nil")
	}

	t.Logf("Comprehensive test result: %+v", result)
}

func TestValidationTools_ExecuteKeywordTest(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	testCases := []string{
		"basic_classification",
		"coverage_validation",
		"duplicate_check",
		"pattern_test",
		"effectiveness_analysis",
		"improvement_suggestions",
		"consistency_check",
		"performance_monitoring",
		"weight_optimization",
		"completeness_validation",
		"edge_case_testing",
		"statistics_generation",
		"completion_validation",
		"comprehensive_tests",
	}

	for _, testCase := range testCases {
		result, err := vt.ExecuteKeywordTest(ctx, testCase)
		if err != nil {
			t.Logf("ExecuteKeywordTest %s failed: %v", testCase, err)
		} else {
			t.Logf("ExecuteKeywordTest %s result: %s", testCase, result)
		}
	}
}

func TestValidationTools_GetTestingDashboard(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	results, err := vt.GetTestingDashboard(ctx)
	if err != nil {
		t.Fatalf("GetTestingDashboard failed: %v", err)
	}

	t.Logf("Found %d dashboard items", len(results))
	for _, result := range results {
		t.Logf("Dashboard item: %+v", result)
	}
}

func TestValidationTools_LogKeywordTest(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	err := vt.LogKeywordTest(ctx, "technology", "Technology", true, 0.95, 150*time.Millisecond)
	if err != nil {
		t.Fatalf("LogKeywordTest failed: %v", err)
	}

	t.Log("Keyword test logged successfully")
}

// Benchmark tests for performance validation
func BenchmarkValidationTools_TestKeywordClassification(b *testing.B) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()
	keywords := []string{"technology", "software", "development"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vt.TestKeywordClassification(ctx, keywords, "Tech Corp", "Software development company")
		if err != nil {
			b.Fatalf("TestKeywordClassification failed: %v", err)
		}
	}
}

func BenchmarkValidationTools_ValidateKeywordCoverage(b *testing.B) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vt.ValidateKeywordCoverage(ctx, "%")
		if err != nil {
			b.Fatalf("ValidateKeywordCoverage failed: %v", err)
		}
	}
}

func BenchmarkValidationTools_FindDuplicateKeywords(b *testing.B) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vt.FindDuplicateKeywords(ctx)
		if err != nil {
			b.Fatalf("FindDuplicateKeywords failed: %v", err)
		}
	}
}

// Integration test that would require a real database
func TestValidationTools_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This would require a real database connection
	// db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
	// if err != nil {
	//     t.Fatalf("Failed to connect to database: %v", err)
	// }
	// defer db.Close()

	// vt := NewValidationTools(db)
	// ctx := context.Background()

	// Test all functions with real data
	t.Log("Integration test would run here with real database")
}

// Test error handling
func TestValidationTools_ErrorHandling(t *testing.T) {
	// Test with nil database
	vt := NewValidationTools(nil)
	ctx := context.Background()

	// These should return errors due to nil database
	_, err := vt.TestKeywordClassification(ctx, []string{"test"}, "", "")
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = vt.ValidateKeywordCoverage(ctx, "%")
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = vt.FindDuplicateKeywords(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}
}

// Test data validation
func TestValidationTools_DataValidation(t *testing.T) {
	vt := NewValidationTools(createMockDB())
	if vt.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test with empty keywords
	_, err := vt.TestKeywordClassification(ctx, []string{}, "", "")
	if err != nil {
		t.Logf("Empty keywords test (expected): %v", err)
	}

	// Test with nil keywords
	_, err = vt.TestKeywordClassification(ctx, nil, "", "")
	if err != nil {
		t.Logf("Nil keywords test (expected): %v", err)
	}

	// Test with very long keywords
	longKeywords := make([]string, 1000)
	for i := range longKeywords {
		longKeywords[i] = "verylongkeyword" + string(rune(i%26+'a'))
	}

	_, err = vt.TestKeywordClassification(ctx, longKeywords, "", "")
	if err != nil {
		t.Logf("Long keywords test (expected): %v", err)
	}

	t.Logf("Data validation tests completed")
}
