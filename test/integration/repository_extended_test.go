//go:build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
)

// TestMerchantAnalyticsRepository_Performance tests performance with large datasets
func TestMerchantAnalyticsRepository_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	repo := database.NewMerchantAnalyticsRepository(db, stdLogger)

	merchantID := "merchant-perf-test"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	// Seed analytics with complex data
	complexClassification := map[string]interface{}{
		"primaryIndustry": "Technology",
		"confidenceScore": 0.95,
		"riskLevel":       "low",
		"mccCodes":        []string{"5734", "7372"},
		"naicsCodes":      []string{"541511", "541512"},
		"sicCodes":        []string{"7371", "7372"},
	}
	complexSecurity := map[string]interface{}{
		"trustScore":  0.8,
		"sslValid":    true,
		"sslExpiry":   time.Now().Add(90 * 24 * time.Hour).Format(time.RFC3339),
		"headers":     []string{"X-Frame-Options", "X-Content-Type-Options"},
		"certificate": "valid",
	}
	complexQuality := map[string]interface{}{
		"completenessScore": 0.9,
		"dataPoints":        100,
		"missingFields":     []string{},
		"validationErrors":  []string{},
	}
	complexIntelligence := map[string]interface{}{
		"newsCount":     10,
		"socialMentions": 5,
		"reviews":       20,
	}

	if err := SeedTestAnalytics(db, merchantID, complexClassification, complexSecurity, complexQuality, complexIntelligence); err != nil {
		t.Fatalf("Failed to seed analytics: %v", err)
	}

	// Test all methods with complex data
	start := time.Now()
	_, err = repo.GetClassificationByMerchantID(ctx, merchantID)
	if err != nil {
		t.Fatalf("Failed to get classification: %v", err)
	}
	classificationTime := time.Since(start)

	start = time.Now()
	_, err = repo.GetSecurityDataByMerchantID(ctx, merchantID)
	if err != nil {
		t.Fatalf("Failed to get security: %v", err)
	}
	securityTime := time.Since(start)

	start = time.Now()
	_, err = repo.GetQualityMetricsByMerchantID(ctx, merchantID)
	if err != nil {
		t.Fatalf("Failed to get quality: %v", err)
	}
	qualityTime := time.Since(start)

	start = time.Now()
	_, err = repo.GetIntelligenceDataByMerchantID(ctx, merchantID)
	if err != nil {
		t.Fatalf("Failed to get intelligence: %v", err)
	}
	intelligenceTime := time.Since(start)

	// Verify performance is reasonable (< 1 second per query)
	maxTime := 1 * time.Second
	if classificationTime > maxTime {
		t.Errorf("Classification query took too long: %v", classificationTime)
	}
	if securityTime > maxTime {
		t.Errorf("Security query took too long: %v", securityTime)
	}
	if qualityTime > maxTime {
		t.Errorf("Quality query took too long: %v", qualityTime)
	}
	if intelligenceTime > maxTime {
		t.Errorf("Intelligence query took too long: %v", intelligenceTime)
	}

	t.Logf("Performance: Classification=%v, Security=%v, Quality=%v, Intelligence=%v",
		classificationTime, securityTime, qualityTime, intelligenceTime)
}

// TestRiskAssessmentRepository_LargeHistory tests retrieving large assessment history
func TestRiskAssessmentRepository_LargeHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	repo := database.NewRiskAssessmentRepository(db, stdLogger)

	merchantID := "merchant-large-history"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	// Seed 50 assessments
	for i := 1; i <= 50; i++ {
		assessmentID := fmt.Sprintf("assessment-large-%d", i)
		status := "completed"
		if i%10 == 0 {
			status = "failed"
		}
		if err := SeedTestRiskAssessment(db, merchantID, assessmentID, status, nil); err != nil {
			t.Fatalf("Failed to seed assessment %d: %v", i, err)
		}
	}

	// Retrieve all assessments
	start := time.Now()
	assessments, err := repo.GetAssessmentsByMerchantID(ctx, merchantID)
	if err != nil {
		t.Fatalf("Failed to get assessments: %v", err)
	}
	duration := time.Since(start)

	if len(assessments) != 50 {
		t.Errorf("Expected 50 assessments, got %d", len(assessments))
	}

	// Verify performance is reasonable
	if duration > 5*time.Second {
		t.Errorf("Retrieving 50 assessments took too long: %v", duration)
	}

	t.Logf("Retrieved %d assessments in %v", len(assessments), duration)
}

// TestRiskIndicatorsRepository_LargeDataset tests with large indicator datasets
func TestRiskIndicatorsRepository_LargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	repo := database.NewRiskIndicatorsRepository(db, stdLogger)

	merchantID := "merchant-large-indicators"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	// Seed 100 indicators
	if err := SeedTestRiskIndicators(db, merchantID, 100, ""); err != nil {
		t.Fatalf("Failed to seed indicators: %v", err)
	}

	// Test retrieval without filters
	start := time.Now()
	indicators, err := repo.GetByMerchantID(ctx, merchantID, nil)
	if err != nil {
		t.Fatalf("Failed to get indicators: %v", err)
	}
	duration := time.Since(start)

	if len(indicators) != 100 {
		t.Errorf("Expected 100 indicators, got %d", len(indicators))
	}

	// Test with filters
	start = time.Now()
	filtered, err := repo.GetByMerchantID(ctx, merchantID, &database.RiskIndicatorFilters{Severity: "high"})
	if err != nil {
		t.Fatalf("Failed to get filtered indicators: %v", err)
	}
	filterDuration := time.Since(start)

	// Verify performance
	if duration > 2*time.Second {
		t.Errorf("Retrieving 100 indicators took too long: %v", duration)
	}
	if filterDuration > 2*time.Second {
		t.Errorf("Filtering indicators took too long: %v", filterDuration)
	}

	t.Logf("Retrieved %d indicators in %v, filtered %d in %v", len(indicators), duration, len(filtered), filterDuration)
}

// TestRiskAssessmentRepository_ComplexResults tests complex result structures
func TestRiskAssessmentRepository_ComplexResults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	repo := database.NewRiskAssessmentRepository(db, stdLogger)

	merchantID := "merchant-complex-results"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	assessmentID := "assessment-complex"
	if err := SeedTestRiskAssessment(db, merchantID, assessmentID, "processing", nil); err != nil {
		t.Fatalf("Failed to seed assessment: %v", err)
	}

	// Create complex result with many factors
	complexResult := &models.RiskAssessmentResult{
		OverallScore: 0.75,
		RiskLevel:    "medium",
		Factors: []models.RiskFactor{
			{Name: "Financial Stability", Score: 0.8, Weight: 0.2, Description: "Strong financial position"},
			{Name: "Operational Risk", Score: 0.7, Weight: 0.2, Description: "Moderate operational concerns"},
			{Name: "Compliance", Score: 0.9, Weight: 0.2, Description: "Good compliance record"},
			{Name: "Market Position", Score: 0.6, Weight: 0.2, Description: "Competitive market"},
			{Name: "Management Quality", Score: 0.7, Weight: 0.2, Description: "Experienced management"},
		},
		Recommendations: []string{
			"Monitor financial metrics quarterly",
			"Enhance operational controls",
			"Maintain compliance standards",
		},
		Confidence: 0.85,
	}

	err = repo.UpdateAssessmentResult(ctx, assessmentID, complexResult)
	if err != nil {
		t.Fatalf("Failed to update complex result: %v", err)
	}

	// Retrieve and verify
	assessment, err := repo.GetAssessmentByID(ctx, assessmentID)
	if err != nil {
		t.Fatalf("Failed to get assessment: %v", err)
	}

	if assessment.Result == nil {
		t.Fatal("Expected result to be set")
	}

	if len(assessment.Result.Factors) != 5 {
		t.Errorf("Expected 5 factors, got %d", len(assessment.Result.Factors))
	}

	if len(assessment.Result.Recommendations) != 3 {
		t.Errorf("Expected 3 recommendations, got %d", len(assessment.Result.Recommendations))
	}

	if assessment.Result.Confidence != 0.85 {
		t.Errorf("Expected confidence 0.85, got %f", assessment.Result.Confidence)
	}
}

// TestMerchantAnalyticsRepository_DataConsistency tests data consistency across methods
func TestMerchantAnalyticsRepository_DataConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	repo := database.NewMerchantAnalyticsRepository(db, stdLogger)

	merchantID := "merchant-consistency"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	// Seed analytics with known values
	classification := map[string]interface{}{
		"primaryIndustry": "Technology",
		"confidenceScore": 0.95,
	}
	security := map[string]interface{}{
		"trustScore": 0.8,
		"sslValid":   true,
	}
	quality := map[string]interface{}{
		"completenessScore": 0.9,
		"dataPoints":        100,
	}
	intelligence := map[string]interface{}{
		"newsCount": 10,
	}

	if err := SeedTestAnalytics(db, merchantID, classification, security, quality, intelligence); err != nil {
		t.Fatalf("Failed to seed analytics: %v", err)
	}

	// Retrieve all data types
	classificationData, err := repo.GetClassificationByMerchantID(ctx, merchantID)
	if err != nil {
		t.Fatalf("Failed to get classification: %v", err)
	}

	securityData, err := repo.GetSecurityDataByMerchantID(ctx, merchantID)
	if err != nil {
		t.Fatalf("Failed to get security: %v", err)
	}

	qualityData, err := repo.GetQualityMetricsByMerchantID(ctx, merchantID)
	if err != nil {
		t.Fatalf("Failed to get quality: %v", err)
	}

	intelligenceData, err := repo.GetIntelligenceDataByMerchantID(ctx, merchantID)
	if err != nil {
		t.Fatalf("Failed to get intelligence: %v", err)
	}

	// Verify consistency - all should reference the same merchant
	if classificationData == nil || securityData == nil || qualityData == nil || intelligenceData == nil {
		t.Fatal("All data types should be retrievable")
	}

	// Verify data integrity
	if classificationData.PrimaryIndustry != "Technology" {
		t.Errorf("Expected primary industry Technology, got %s", classificationData.PrimaryIndustry)
	}

	if securityData.TrustScore != 0.8 {
		t.Errorf("Expected trust score 0.8, got %f", securityData.TrustScore)
	}

	if qualityData.CompletenessScore != 0.9 {
		t.Errorf("Expected completeness score 0.9, got %f", qualityData.CompletenessScore)
	}
}

